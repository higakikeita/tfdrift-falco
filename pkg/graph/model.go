package graph

import (
	"sync"
)

// Node represents a graph node (similar to Neo4j nodes)
// Example: (:Resource:EC2 {id: "i-123", name: "web-server"})
type Node struct {
	ID         string                 `json:"id"`
	Labels     []string               `json:"labels"` // ["Resource", "EC2", "Drifted"]
	Properties map[string]interface{} `json:"properties"`
}

// Relationship represents a directed edge between two nodes
// Example: (ec2)-[:DEPENDS_ON {since: "2024-01-01"}]->(subnet)
type Relationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // "DEPENDS_ON", "CONTAINS", etc.
	StartNode  string                 `json:"start_node"` // Source node ID
	EndNode    string                 `json:"end_node"`   // Target node ID
	Properties map[string]interface{} `json:"properties"`
}

// Relationship types (similar to Neo4j relationship types)
const (
	// Hierarchical relationships
	CONTAINS = "CONTAINS" // VPC contains Subnet
	PART_OF  = "PART_OF"  // Subnet part of VPC

	// Dependency relationships
	DEPENDS_ON  = "DEPENDS_ON"  // EC2 depends on Subnet
	REQUIRED_BY = "REQUIRED_BY" // Inverse of DEPENDS_ON

	// Network relationships
	CONNECTS_TO = "CONNECTS_TO" // EC2 connects to RDS
	ROUTES_TO   = "ROUTES_TO"   // Route table routes to Gateway

	// Security relationships
	ALLOWS  = "ALLOWS"  // SecurityGroup allows traffic
	BLOCKS  = "BLOCKS"  // NACL blocks traffic
	SECURES = "SECURES" // SecurityGroup secures resource

	// Application relationships
	RUNS_IN      = "RUNS_IN"      // ECS Service runs in ECS Cluster
	REGISTERS_TO = "REGISTERS_TO" // ECS Service registers to Target Group
	APPLIES_TO   = "APPLIES_TO"   // IAM Policy applies to Role
	ASSOCIATES   = "ASSOCIATES"   // Route Table associates with Subnet

	// Change relationships
	DRIFTED_FROM    = "DRIFTED_FROM"    // Current state drifted from desired
	CAUSED_DRIFT_IN = "CAUSED_DRIFT_IN" // Change A caused drift in B
)

// GraphDatabase is an in-memory graph database inspired by Neo4j
type GraphDatabase struct {
	mu sync.RWMutex

	// Core storage
	nodes         map[string]*Node
	relationships map[string]*Relationship

	// Indexes for fast lookups
	nodesByLabel map[string]map[string]*Node         // label -> {node_id -> node}
	outgoing     map[string]map[string]*Relationship // node_id -> {rel_id -> relationship}
	incoming     map[string]map[string]*Relationship // node_id -> {rel_id -> relationship}

	// Type-based relationship indexes
	relationshipsByType map[string]map[string]*Relationship // type -> {rel_id -> relationship}
}

// NewGraphDatabase creates a new in-memory graph database
func NewGraphDatabase() *GraphDatabase {
	return &GraphDatabase{
		nodes:               make(map[string]*Node),
		relationships:       make(map[string]*Relationship),
		nodesByLabel:        make(map[string]map[string]*Node),
		outgoing:            make(map[string]map[string]*Relationship),
		incoming:            make(map[string]map[string]*Relationship),
		relationshipsByType: make(map[string]map[string]*Relationship),
	}
}

// AddNode adds a node to the graph
func (db *GraphDatabase) AddNode(node *Node) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.nodes[node.ID] = node

	// Index by labels
	for _, label := range node.Labels {
		if db.nodesByLabel[label] == nil {
			db.nodesByLabel[label] = make(map[string]*Node)
		}
		db.nodesByLabel[label][node.ID] = node
	}

	// Initialize relationship maps for this node
	if db.outgoing[node.ID] == nil {
		db.outgoing[node.ID] = make(map[string]*Relationship)
	}
	if db.incoming[node.ID] == nil {
		db.incoming[node.ID] = make(map[string]*Relationship)
	}
}

// GetNode retrieves a node by ID
func (db *GraphDatabase) GetNode(id string) *Node {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.nodes[id]
}

// GetNodesByLabel retrieves all nodes with a specific label
func (db *GraphDatabase) GetNodesByLabel(label string) []*Node {
	db.mu.RLock()
	defer db.mu.RUnlock()

	nodeMap := db.nodesByLabel[label]
	nodes := make([]*Node, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}
	return nodes
}

// HasLabel checks if a node has a specific label
func (db *GraphDatabase) HasLabel(nodeID string, label string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	node := db.nodes[nodeID]
	if node == nil {
		return false
	}

	for _, l := range node.Labels {
		if l == label {
			return true
		}
	}
	return false
}

// AddRelationship adds a relationship to the graph
func (db *GraphDatabase) AddRelationship(rel *Relationship) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Verify nodes exist
	if db.nodes[rel.StartNode] == nil || db.nodes[rel.EndNode] == nil {
		return ErrNodeNotFound
	}

	db.relationships[rel.ID] = rel

	// Index outgoing relationships
	if db.outgoing[rel.StartNode] == nil {
		db.outgoing[rel.StartNode] = make(map[string]*Relationship)
	}
	db.outgoing[rel.StartNode][rel.ID] = rel

	// Index incoming relationships
	if db.incoming[rel.EndNode] == nil {
		db.incoming[rel.EndNode] = make(map[string]*Relationship)
	}
	db.incoming[rel.EndNode][rel.ID] = rel

	// Index by type
	if db.relationshipsByType[rel.Type] == nil {
		db.relationshipsByType[rel.Type] = make(map[string]*Relationship)
	}
	db.relationshipsByType[rel.Type][rel.ID] = rel

	return nil
}

// GetRelationship retrieves a relationship by ID
func (db *GraphDatabase) GetRelationship(id string) *Relationship {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.relationships[id]
}

// GetOutgoingRelationships returns all outgoing relationships from a node
func (db *GraphDatabase) GetOutgoingRelationships(nodeID string) []*Relationship {
	db.mu.RLock()
	defer db.mu.RUnlock()

	relMap := db.outgoing[nodeID]
	rels := make([]*Relationship, 0, len(relMap))
	for _, rel := range relMap {
		rels = append(rels, rel)
	}
	return rels
}

// GetIncomingRelationships returns all incoming relationships to a node
func (db *GraphDatabase) GetIncomingRelationships(nodeID string) []*Relationship {
	db.mu.RLock()
	defer db.mu.RUnlock()

	relMap := db.incoming[nodeID]
	rels := make([]*Relationship, 0, len(relMap))
	for _, rel := range relMap {
		rels = append(rels, rel)
	}
	return rels
}

// GetRelationshipsByType returns all relationships of a specific type
func (db *GraphDatabase) GetRelationshipsByType(relType string) []*Relationship {
	db.mu.RLock()
	defer db.mu.RUnlock()

	relMap := db.relationshipsByType[relType]
	rels := make([]*Relationship, 0, len(relMap))
	for _, rel := range relMap {
		rels = append(rels, rel)
	}
	return rels
}

// GetNeighbors returns all directly connected nodes (both incoming and outgoing)
func (db *GraphDatabase) GetNeighbors(nodeID string) []*Node {
	db.mu.RLock()
	defer db.mu.RUnlock()

	neighborIDs := make(map[string]bool)

	// Add outgoing neighbors
	for _, rel := range db.outgoing[nodeID] {
		neighborIDs[rel.EndNode] = true
	}

	// Add incoming neighbors
	for _, rel := range db.incoming[nodeID] {
		neighborIDs[rel.StartNode] = true
	}

	// Convert to node list
	neighbors := make([]*Node, 0, len(neighborIDs))
	for nID := range neighborIDs {
		if node := db.nodes[nID]; node != nil {
			neighbors = append(neighbors, node)
		}
	}

	return neighbors
}

// GetAllNodes returns all nodes in the graph
func (db *GraphDatabase) GetAllNodes() []*Node {
	db.mu.RLock()
	defer db.mu.RUnlock()

	nodes := make([]*Node, 0, len(db.nodes))
	for _, node := range db.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetAllRelationships returns all relationships in the graph
func (db *GraphDatabase) GetAllRelationships() []*Relationship {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rels := make([]*Relationship, 0, len(db.relationships))
	for _, rel := range db.relationships {
		rels = append(rels, rel)
	}
	return rels
}

// NodeCount returns the total number of nodes
func (db *GraphDatabase) NodeCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.nodes)
}

// RelationshipCount returns the total number of relationships
func (db *GraphDatabase) RelationshipCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.relationships)
}

// DeleteNode removes a node and all its relationships
func (db *GraphDatabase) DeleteNode(nodeID string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	node := db.nodes[nodeID]
	if node == nil {
		return
	}

	// Remove from label indexes
	for _, label := range node.Labels {
		if labelMap := db.nodesByLabel[label]; labelMap != nil {
			delete(labelMap, nodeID)
		}
	}

	// Remove all outgoing relationships
	for relID := range db.outgoing[nodeID] {
		rel := db.relationships[relID]
		if rel != nil {
			// Remove from incoming index of target node
			if incMap := db.incoming[rel.EndNode]; incMap != nil {
				delete(incMap, relID)
			}
			// Remove from type index
			if typeMap := db.relationshipsByType[rel.Type]; typeMap != nil {
				delete(typeMap, relID)
			}
			delete(db.relationships, relID)
		}
	}
	delete(db.outgoing, nodeID)

	// Remove all incoming relationships
	for relID := range db.incoming[nodeID] {
		rel := db.relationships[relID]
		if rel != nil {
			// Remove from outgoing index of source node
			if outMap := db.outgoing[rel.StartNode]; outMap != nil {
				delete(outMap, relID)
			}
			// Remove from type index
			if typeMap := db.relationshipsByType[rel.Type]; typeMap != nil {
				delete(typeMap, relID)
			}
			delete(db.relationships, relID)
		}
	}
	delete(db.incoming, nodeID)

	// Finally, remove the node itself
	delete(db.nodes, nodeID)
}

// Clear removes all nodes and relationships
func (db *GraphDatabase) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.nodes = make(map[string]*Node)
	db.relationships = make(map[string]*Relationship)
	db.nodesByLabel = make(map[string]map[string]*Node)
	db.outgoing = make(map[string]map[string]*Relationship)
	db.incoming = make(map[string]map[string]*Relationship)
	db.relationshipsByType = make(map[string]map[string]*Relationship)
}
