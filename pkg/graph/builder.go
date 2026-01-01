package graph

import (
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Store maintains the graph data in memory
type Store struct {
	drifts       []types.DriftAlert
	events       []types.Event
	unmanaged    []types.UnmanagedResourceAlert
	stateManager *terraform.StateManager
	graphDB      *GraphDatabase // Neo4j-style graph database
	mu           sync.RWMutex
}

// NewStore creates a new graph store
func NewStore() *Store {
	return &Store{
		drifts:    make([]types.DriftAlert, 0),
		events:    make([]types.Event, 0),
		unmanaged: make([]types.UnmanagedResourceAlert, 0),
		graphDB:   NewGraphDatabase(),
	}
}

// AddDrift adds a drift alert to the store
func (s *Store) AddDrift(drift types.DriftAlert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drifts = append(s.drifts, drift)
}

// AddEvent adds a Falco event to the store
func (s *Store) AddEvent(event types.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

// AddUnmanaged adds an unmanaged resource to the store
func (s *Store) AddUnmanaged(unmanaged types.UnmanagedResourceAlert) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unmanaged = append(s.unmanaged, unmanaged)
}

// GetDrifts returns all drift alerts
func (s *Store) GetDrifts() []types.DriftAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]types.DriftAlert, len(s.drifts))
	copy(result, s.drifts)
	return result
}

// GetEvents returns all events
func (s *Store) GetEvents() []types.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]types.Event, len(s.events))
	copy(result, s.events)
	return result
}

// GetUnmanaged returns all unmanaged resources
func (s *Store) GetUnmanaged() []types.UnmanagedResourceAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]types.UnmanagedResourceAlert, len(s.unmanaged))
	copy(result, s.unmanaged)
	return result
}

// Clear clears all data from the store
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.drifts = make([]types.DriftAlert, 0)
	s.events = make([]types.Event, 0)
	s.unmanaged = make([]types.UnmanagedResourceAlert, 0)
}

// SetStateManager sets the Terraform state manager for graph building
func (s *Store) SetStateManager(sm *terraform.StateManager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stateManager = sm

	// Rebuild graph database when state manager is set
	s.rebuildGraphDB()
}

// GetGraphDB returns the graph database
func (s *Store) GetGraphDB() *GraphDatabase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	log.Infof("[GetGraphDB] Returning GraphDatabase instance: %p (node count: %d)", s.graphDB, s.graphDB.NodeCount())
	return s.graphDB
}

// RebuildGraphDB rebuilds the graph database from current resources
// Can be called externally after state manager loads resources
func (s *Store) RebuildGraphDB() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rebuildGraphDB()
}

// rebuildGraphDB rebuilds the graph database from current resources
// Should be called with lock held
func (s *Store) rebuildGraphDB() {
	if s.stateManager == nil {
		return
	}

	resources := s.stateManager.GetAllResources()

	// Track which resources have drifts
	driftedIDs := make(map[string]bool)
	for _, drift := range s.drifts {
		driftedIDs[drift.ResourceID] = true
	}

	// Convert to graph database
	s.graphDB = TerraformToGraph(resources, driftedIDs)
}

// buildDependencyEdges extracts dependency relationships from resources
func buildDependencyEdges(resources []*terraform.Resource) []models.CytoscapeEdge {
	edges := []models.CytoscapeEdge{}
	resourceMap := make(map[string]*terraform.Resource)

	// Index resources by ID
	for _, resource := range resources {
		resourceID := extractResourceIDFromAttributes(resource.Attributes)
		if resourceID != "" {
			resourceMap[resourceID] = resource
		}
	}

	// Extract dependencies
	for _, resource := range resources {
		resourceID := extractResourceIDFromAttributes(resource.Attributes)
		if resourceID == "" {
			continue
		}

		switch resource.Type {
		case "aws_instance":
			// EC2 depends on Subnet
			if subnetID, ok := resource.Attributes["subnet_id"].(string); ok && subnetID != "" {
				if _, exists := resourceMap[subnetID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + subnetID,
							Source: resourceID,
							Target: subnetID,
							Label:  "in subnet",
						},
					})
				}
			}
			// EC2 depends on Security Groups
			if sgIDs, ok := resource.Attributes["vpc_security_group_ids"].([]interface{}); ok {
				for _, sgID := range sgIDs {
					if sgIDStr, ok := sgID.(string); ok && sgIDStr != "" {
						if _, exists := resourceMap[sgIDStr]; exists {
							edges = append(edges, models.CytoscapeEdge{
								Data: models.EdgeData{
									ID:     resourceID + "->" + sgIDStr,
									Source: resourceID,
									Target: sgIDStr,
									Label:  "uses",
								},
							})
						}
					}
				}
			}

		case "aws_subnet":
			// Subnet depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "in vpc",
						},
					})
				}
			}

		case "aws_nat_gateway":
			// NAT Gateway depends on Subnet
			if subnetID, ok := resource.Attributes["subnet_id"].(string); ok && subnetID != "" {
				if _, exists := resourceMap[subnetID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + subnetID,
							Source: resourceID,
							Target: subnetID,
							Label:  "in subnet",
						},
					})
				}
			}

		case "aws_route_table":
			// Route table depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "routes for",
						},
					})
				}
			}

		case "aws_security_group":
			// Security Group depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "secures",
						},
					})
				}
			}

		case "aws_internet_gateway":
			// IGW depends on VPC
			if vpcID, ok := resource.Attributes["vpc_id"].(string); ok && vpcID != "" {
				if _, exists := resourceMap[vpcID]; exists {
					edges = append(edges, models.CytoscapeEdge{
						Data: models.EdgeData{
							ID:     resourceID + "->" + vpcID,
							Source: resourceID,
							Target: vpcID,
							Label:  "gateway for",
						},
					})
				}
			}
		}
	}

	return edges
}

// BuildGraph builds a Cytoscape graph from stored data
func (s *Store) BuildGraph() models.CytoscapeElements {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]models.CytoscapeNode, 0)
	edges := make([]models.CytoscapeEdge, 0)
	nodeIDs := make(map[string]bool)
	driftedIDs := make(map[string]bool)

	// Track which resources have drifts
	for _, drift := range s.drifts {
		driftedIDs[drift.ResourceID] = true
	}

	// FIRST: Add all Terraform State resources as flat graph
	if s.stateManager != nil {
		resources := s.stateManager.GetAllResources()

		// Add all resources as nodes
		for _, resource := range resources {
			resourceID := extractResourceIDFromAttributes(resource.Attributes)
			if resourceID != "" && !nodeIDs[resourceID] {
				// Determine if this resource has drifted
				hasDrift := driftedIDs[resourceID]
				resourceNode := ConvertTerraformResourceToCytoscape(resource, hasDrift)
				nodes = append(nodes, resourceNode)
				nodeIDs[resourceID] = true
			}
		}

		// Build dependency edges between resources
		dependencyEdges := buildDependencyEdges(resources)
		edges = append(edges, dependencyEdges...)
	}

	// SECOND: Add drift nodes (for resources not in Terraform State)
	for _, drift := range s.drifts {
		if !nodeIDs[drift.ResourceID] {
			nodes = append(nodes, ConvertDriftToCytoscape(drift))
			nodeIDs[drift.ResourceID] = true
		}
	}

	// Add event nodes
	for _, event := range s.events {
		if !nodeIDs[event.ResourceID] {
			nodes = append(nodes, ConvertEventToCytoscape(event))
			nodeIDs[event.ResourceID] = true
		}
	}

	// Add unmanaged nodes
	for _, unmanaged := range s.unmanaged {
		if !nodeIDs[unmanaged.ResourceID] {
			nodes = append(nodes, ConvertUnmanagedToCytoscape(unmanaged))
			nodeIDs[unmanaged.ResourceID] = true
		}
	}

	// Create edges (causal relationships)
	// For now, create simple sequential edges
	// This can be enhanced with more sophisticated causal analysis
	for i := 0; i < len(s.events)-1; i++ {
		edge := CreateEdge(
			s.events[i].ResourceID,
			s.events[i+1].ResourceID,
			"triggered",
			"caused_by",
			"event_sequence",
		)
		edges = append(edges, edge)
	}

	// Connect events to drifts
	for _, drift := range s.drifts {
		for _, event := range s.events {
			if event.ResourceID == drift.ResourceID || event.ResourceType == drift.ResourceType {
				edge := CreateEdge(
					event.ResourceID,
					drift.ResourceID,
					"caused",
					"caused_by",
					"drift_detection",
				)
				edges = append(edges, edge)
				break
			}
		}
	}

	return models.CytoscapeElements{
		Nodes: nodes,
		Edges: edges,
	}
}

// GetStats returns graph statistics
func (s *Store) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	severityCounts := make(map[string]int)
	for _, drift := range s.drifts {
		severityCounts[drift.Severity]++
	}

	resourceTypeCounts := make(map[string]int)
	for _, drift := range s.drifts {
		resourceTypeCounts[drift.ResourceType]++
	}

	return map[string]interface{}{
		"total_drifts":         len(s.drifts),
		"total_events":         len(s.events),
		"total_unmanaged":      len(s.unmanaged),
		"severity_counts":      severityCounts,
		"resource_type_counts": resourceTypeCounts,
	}
}
