package graph

import (
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// Store maintains the graph data in memory
type Store struct {
	drifts    []types.DriftAlert
	events    []types.Event
	unmanaged []types.UnmanagedResourceAlert
	mu        sync.RWMutex
}

// NewStore creates a new graph store
func NewStore() *Store {
	return &Store{
		drifts:    make([]types.DriftAlert, 0),
		events:    make([]types.Event, 0),
		unmanaged: make([]types.UnmanagedResourceAlert, 0),
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

// BuildGraph builds a Cytoscape graph from stored data
func (s *Store) BuildGraph() models.CytoscapeElements {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]models.CytoscapeNode, 0)
	edges := make([]models.CytoscapeEdge, 0)
	nodeIDs := make(map[string]bool)

	// Add drift nodes
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
