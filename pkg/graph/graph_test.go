package graph

import (
	"fmt"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// TestNodeCreation tests basic node creation and retrieval
func TestNodeCreation(t *testing.T) {
	db := NewGraphDatabase()

	node := &Node{
		ID:     "node-1",
		Labels: []string{"Resource", "EC2"},
		Properties: map[string]interface{}{
			"name": "web-server",
			"type": "aws_instance",
		},
	}

	db.AddNode(node)

	// Test retrieval
	retrieved := db.GetNode("node-1")
	if retrieved == nil {
		t.Fatal("Expected to retrieve node, got nil")
	}

	if retrieved.ID != "node-1" {
		t.Errorf("Expected ID 'node-1', got '%s'", retrieved.ID)
	}

	if len(retrieved.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(retrieved.Labels))
	}

	if retrieved.Properties["name"] != "web-server" {
		t.Errorf("Expected name property 'web-server', got '%v'", retrieved.Properties["name"])
	}
}

// TestGetNodesByLabel tests label-based node retrieval
func TestGetNodesByLabel(t *testing.T) {
	db := NewGraphDatabase()

	node1 := &Node{
		ID:     "ec2-1",
		Labels: []string{"Resource", "EC2"},
		Properties: map[string]interface{}{
			"name": "instance-1",
		},
	}

	node2 := &Node{
		ID:     "ec2-2",
		Labels: []string{"Resource", "EC2", "Drifted"},
		Properties: map[string]interface{}{
			"name": "instance-2",
		},
	}

	node3 := &Node{
		ID:     "vpc-1",
		Labels: []string{"Resource", "VPC"},
		Properties: map[string]interface{}{
			"name": "vpc-1",
		},
	}

	db.AddNode(node1)
	db.AddNode(node2)
	db.AddNode(node3)

	// Get all EC2 nodes
	ec2Nodes := db.GetNodesByLabel("EC2")
	if len(ec2Nodes) != 2 {
		t.Errorf("Expected 2 EC2 nodes, got %d", len(ec2Nodes))
	}

	// Get all VPC nodes
	vpcNodes := db.GetNodesByLabel("VPC")
	if len(vpcNodes) != 1 {
		t.Errorf("Expected 1 VPC node, got %d", len(vpcNodes))
	}

	// Get all Drifted nodes
	driftedNodes := db.GetNodesByLabel("Drifted")
	if len(driftedNodes) != 1 {
		t.Errorf("Expected 1 Drifted node, got %d", len(driftedNodes))
	}
}

// TestHasLabel tests label checking
func TestHasLabel(t *testing.T) {
	db := NewGraphDatabase()

	node := &Node{
		ID:     "test-node",
		Labels: []string{"Resource", "EC2", "Drifted"},
		Properties: map[string]interface{}{},
	}

	db.AddNode(node)

	// Test existing labels
	if !db.HasLabel("test-node", "EC2") {
		t.Error("Expected node to have EC2 label")
	}

	if !db.HasLabel("test-node", "Drifted") {
		t.Error("Expected node to have Drifted label")
	}

	// Test non-existing label
	if db.HasLabel("test-node", "NonExistent") {
		t.Error("Expected node to NOT have NonExistent label")
	}

	// Test non-existing node
	if db.HasLabel("non-existent-node", "EC2") {
		t.Error("Expected non-existent node to return false")
	}
}

// TestRelationshipCreation tests relationship creation and retrieval
func TestRelationshipCreation(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes first
	node1 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	node2 := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}

	db.AddNode(node1)
	db.AddNode(node2)

	// Create relationship
	rel := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "ec2-1",
		EndNode:   "subnet-1",
		Properties: map[string]interface{}{
			"type": "network_placement",
		},
	}

	err := db.AddRelationship(rel)
	if err != nil {
		t.Fatalf("Failed to add relationship: %v", err)
	}

	// Test retrieval
	retrieved := db.GetRelationship("rel-1")
	if retrieved == nil {
		t.Fatal("Expected to retrieve relationship, got nil")
	}

	if retrieved.Type != DEPENDS_ON {
		t.Errorf("Expected type DEPENDS_ON, got %s", retrieved.Type)
	}

	if retrieved.StartNode != "ec2-1" {
		t.Errorf("Expected StartNode 'ec2-1', got '%s'", retrieved.StartNode)
	}

	if retrieved.EndNode != "subnet-1" {
		t.Errorf("Expected EndNode 'subnet-1', got '%s'", retrieved.EndNode)
	}
}

// TestRelationshipValidation tests that relationships require both nodes to exist
func TestRelationshipValidation(t *testing.T) {
	db := NewGraphDatabase()

	node1 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	db.AddNode(node1)

	// Try to create relationship where end node doesn't exist
	rel := &Relationship{
		ID:        "rel-invalid",
		Type:      DEPENDS_ON,
		StartNode: "ec2-1",
		EndNode:   "non-existent",
		Properties: map[string]interface{}{},
	}

	err := db.AddRelationship(rel)
	if err == nil {
		t.Fatal("Expected error when adding relationship with non-existent node")
	}

	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}

// TestGetOutgoingRelationships tests outgoing relationship retrieval
func TestGetOutgoingRelationships(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	ec2 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	subnet := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	sg := &Node{ID: "sg-1", Labels: []string{"SecurityGroup"}, Properties: map[string]interface{}{}}

	db.AddNode(ec2)
	db.AddNode(subnet)
	db.AddNode(sg)

	// Create outgoing relationships from ec2
	rel1 := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "ec2-1",
		EndNode:   "subnet-1",
		Properties: map[string]interface{}{},
	}

	rel2 := &Relationship{
		ID:        "rel-2",
		Type:      SECURES,
		StartNode: "sg-1",
		EndNode:   "ec2-1",
		Properties: map[string]interface{}{},
	}

	db.AddRelationship(rel1)
	db.AddRelationship(rel2)

	// Test outgoing from ec2-1
	outgoing := db.GetOutgoingRelationships("ec2-1")
	if len(outgoing) != 1 {
		t.Errorf("Expected 1 outgoing relationship, got %d", len(outgoing))
	}

	// Test outgoing from sg-1
	sgOutgoing := db.GetOutgoingRelationships("sg-1")
	if len(sgOutgoing) != 1 {
		t.Errorf("Expected 1 outgoing relationship from sg, got %d", len(sgOutgoing))
	}
}

// TestGetIncomingRelationships tests incoming relationship retrieval
func TestGetIncomingRelationships(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	ec2 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	subnet := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	sg := &Node{ID: "sg-1", Labels: []string{"SecurityGroup"}, Properties: map[string]interface{}{}}

	db.AddNode(ec2)
	db.AddNode(subnet)
	db.AddNode(sg)

	// Create relationships
	rel1 := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "ec2-1",
		EndNode:   "subnet-1",
		Properties: map[string]interface{}{},
	}

	rel2 := &Relationship{
		ID:        "rel-2",
		Type:      SECURES,
		StartNode: "sg-1",
		EndNode:   "ec2-1",
		Properties: map[string]interface{}{},
	}

	db.AddRelationship(rel1)
	db.AddRelationship(rel2)

	// Test incoming to ec2-1
	incoming := db.GetIncomingRelationships("ec2-1")
	if len(incoming) != 1 {
		t.Errorf("Expected 1 incoming relationship to ec2-1, got %d", len(incoming))
	}

	// Test incoming to subnet-1
	subnetIncoming := db.GetIncomingRelationships("subnet-1")
	if len(subnetIncoming) != 1 {
		t.Errorf("Expected 1 incoming relationship to subnet-1, got %d", len(subnetIncoming))
	}
}

// TestGetRelationshipsByType tests type-based relationship retrieval
func TestGetRelationshipsByType(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	nodes := []*Node{
		{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}},
		{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}},
		{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}},
		{ID: "sg-1", Labels: []string{"SecurityGroup"}, Properties: map[string]interface{}{}},
	}

	for _, node := range nodes {
		db.AddNode(node)
	}

	// Create relationships
	relationships := []*Relationship{
		{
			ID:        "rel-1",
			Type:      DEPENDS_ON,
			StartNode: "ec2-1",
			EndNode:   "subnet-1",
			Properties: map[string]interface{}{},
		},
		{
			ID:        "rel-2",
			Type:      PART_OF,
			StartNode: "subnet-1",
			EndNode:   "vpc-1",
			Properties: map[string]interface{}{},
		},
		{
			ID:        "rel-3",
			Type:      DEPENDS_ON,
			StartNode: "subnet-1",
			EndNode:   "vpc-1",
			Properties: map[string]interface{}{},
		},
	}

	for _, rel := range relationships {
		db.AddRelationship(rel)
	}

	// Test DEPENDS_ON relationships
	dependsOnRels := db.GetRelationshipsByType(DEPENDS_ON)
	if len(dependsOnRels) != 2 {
		t.Errorf("Expected 2 DEPENDS_ON relationships, got %d", len(dependsOnRels))
	}

	// Test PART_OF relationships
	partOfRels := db.GetRelationshipsByType(PART_OF)
	if len(partOfRels) != 1 {
		t.Errorf("Expected 1 PART_OF relationship, got %d", len(partOfRels))
	}
}

// TestGetNeighbors tests neighbor retrieval
func TestGetNeighbors(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	ec2 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	subnet := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	sg := &Node{ID: "sg-1", Labels: []string{"SecurityGroup"}, Properties: map[string]interface{}{}}
	vpc := &Node{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}}

	db.AddNode(ec2)
	db.AddNode(subnet)
	db.AddNode(sg)
	db.AddNode(vpc)

	// Create relationships
	rels := []*Relationship{
		{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}},
		{ID: "rel-2", Type: SECURES, StartNode: "sg-1", EndNode: "ec2-1", Properties: map[string]interface{}{}},
		{ID: "rel-3", Type: PART_OF, StartNode: "subnet-1", EndNode: "vpc-1", Properties: map[string]interface{}{}},
	}

	for _, rel := range rels {
		db.AddRelationship(rel)
	}

	// Get neighbors of ec2-1 (should include subnet-1 and sg-1)
	neighbors := db.GetNeighbors("ec2-1")
	if len(neighbors) != 2 {
		t.Errorf("Expected 2 neighbors for ec2-1, got %d", len(neighbors))
	}

	neighborIDs := make(map[string]bool)
	for _, n := range neighbors {
		neighborIDs[n.ID] = true
	}

	if !neighborIDs["subnet-1"] || !neighborIDs["sg-1"] {
		t.Error("Expected neighbors to include subnet-1 and sg-1")
	}
}

// TestGetAllNodes tests retrieving all nodes
func TestGetAllNodes(t *testing.T) {
	db := NewGraphDatabase()

	nodes := []*Node{
		{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}},
		{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}},
		{ID: "node-3", Labels: []string{"C"}, Properties: map[string]interface{}{}},
	}

	for _, node := range nodes {
		db.AddNode(node)
	}

	allNodes := db.GetAllNodes()
	if len(allNodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(allNodes))
	}
}

// TestGetAllRelationships tests retrieving all relationships
func TestGetAllRelationships(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	for i := 1; i <= 3; i++ {
		db.AddNode(&Node{ID: "node-" + string(rune('0'+i)), Labels: []string{"A"}, Properties: map[string]interface{}{}})
	}

	// Create relationships
	rels := []*Relationship{
		{ID: "rel-1", Type: DEPENDS_ON, StartNode: "node-1", EndNode: "node-2", Properties: map[string]interface{}{}},
		{ID: "rel-2", Type: PART_OF, StartNode: "node-2", EndNode: "node-3", Properties: map[string]interface{}{}},
	}

	for _, rel := range rels {
		db.AddRelationship(rel)
	}

	allRels := db.GetAllRelationships()
	if len(allRels) != 2 {
		t.Errorf("Expected 2 relationships, got %d", len(allRels))
	}
}

// TestNodeCount and RelationshipCount
func TestCounts(t *testing.T) {
	db := NewGraphDatabase()

	if db.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", db.NodeCount())
	}

	// Add nodes
	db.AddNode(&Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}})

	if db.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes, got %d", db.NodeCount())
	}

	// Add relationship
	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "node-1", EndNode: "node-2", Properties: map[string]interface{}{}})

	if db.RelationshipCount() != 1 {
		t.Errorf("Expected 1 relationship, got %d", db.RelationshipCount())
	}
}

// TestDeleteNode tests node deletion and cascade deletion of relationships
func TestDeleteNode(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	db.AddNode(&Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "sg-1", Labels: []string{"SecurityGroup"}, Properties: map[string]interface{}{}})

	// Create relationships
	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: SECURES, StartNode: "sg-1", EndNode: "ec2-1", Properties: map[string]interface{}{}})

	if db.NodeCount() != 3 {
		t.Errorf("Expected 3 nodes before deletion, got %d", db.NodeCount())
	}

	if db.RelationshipCount() != 2 {
		t.Errorf("Expected 2 relationships before deletion, got %d", db.RelationshipCount())
	}

	// Delete ec2-1
	db.DeleteNode("ec2-1")

	if db.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes after deletion, got %d", db.NodeCount())
	}

	// Both relationships should be deleted (one starts from ec2-1, one ends at ec2-1)
	if db.RelationshipCount() != 0 {
		t.Errorf("Expected 0 relationships after deletion, got %d", db.RelationshipCount())
	}

	// Verify node is really gone
	if db.GetNode("ec2-1") != nil {
		t.Error("Expected deleted node to be nil")
	}
}

// TestClear tests clearing the entire database
func TestClear(t *testing.T) {
	db := NewGraphDatabase()

	// Add some data
	db.AddNode(&Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "node-1", EndNode: "node-2", Properties: map[string]interface{}{}})

	if db.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes before clear, got %d", db.NodeCount())
	}

	// Clear
	db.Clear()

	if db.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes after clear, got %d", db.NodeCount())
	}

	if db.RelationshipCount() != 0 {
		t.Errorf("Expected 0 relationships after clear, got %d", db.RelationshipCount())
	}
}

// TestFindPath tests BFS path finding
func TestFindPath(t *testing.T) {
	db := NewGraphDatabase()

	// Create a chain: EC2 -> Subnet -> VPC
	ec2 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	subnet := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	vpc := &Node{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}}

	db.AddNode(ec2)
	db.AddNode(subnet)
	db.AddNode(vpc)

	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: PART_OF, StartNode: "subnet-1", EndNode: "vpc-1", Properties: map[string]interface{}{}})

	// Find path from EC2 to VPC
	path, err := db.FindPath("ec2-1", "vpc-1")
	if err != nil {
		t.Fatalf("Failed to find path: %v", err)
	}

	if path.Length != 2 {
		t.Errorf("Expected path length 2, got %d", path.Length)
	}

	if len(path.Nodes) != 3 {
		t.Errorf("Expected 3 nodes in path, got %d", len(path.Nodes))
	}

	if path.Nodes[0].ID != "ec2-1" || path.Nodes[2].ID != "vpc-1" {
		t.Error("Path should start with ec2-1 and end with vpc-1")
	}
}

// TestFindPathNotFound tests path finding when no path exists
func TestFindPathNotFound(t *testing.T) {
	db := NewGraphDatabase()

	// Create isolated nodes
	db.AddNode(&Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}})

	// Try to find path between disconnected nodes
	path, err := db.FindPath("node-1", "node-2")
	if err == nil {
		t.Fatal("Expected error when finding path between disconnected nodes")
	}

	if err != ErrInvalidPath {
		t.Errorf("Expected ErrInvalidPath, got %v", err)
	}

	if path != nil {
		t.Error("Expected nil path when path not found")
	}
}

// TestFindPathInvalidNode tests path finding with non-existent nodes
func TestFindPathInvalidNode(t *testing.T) {
	db := NewGraphDatabase()

	db.AddNode(&Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}})

	// Try to find path with non-existent start node
	_, err := db.FindPath("non-existent", "node-1")
	if err == nil {
		t.Fatal("Expected error when finding path with non-existent start node")
	}

	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}

	// Try to find path with non-existent end node
	_, err = db.FindPath("node-1", "non-existent")
	if err == nil {
		t.Fatal("Expected error when finding path with non-existent end node")
	}

	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}

// TestFindImpactRadius tests finding nodes within a depth
func TestFindImpactRadius(t *testing.T) {
	db := NewGraphDatabase()

	// Create a tree structure
	// VPC
	//  ├─ Subnet-1
	//  │   └─ EC2-1
	//  └─ Subnet-2
	//      └─ EC2-2

	vpc := &Node{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}}
	subnet1 := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	subnet2 := &Node{ID: "subnet-2", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	ec2_1 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	ec2_2 := &Node{ID: "ec2-2", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}

	db.AddNode(vpc)
	db.AddNode(subnet1)
	db.AddNode(subnet2)
	db.AddNode(ec2_1)
	db.AddNode(ec2_2)

	db.AddRelationship(&Relationship{ID: "rel-1", Type: CONTAINS, StartNode: "vpc-1", EndNode: "subnet-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: CONTAINS, StartNode: "vpc-1", EndNode: "subnet-2", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-3", Type: CONTAINS, StartNode: "subnet-1", EndNode: "ec2-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-4", Type: CONTAINS, StartNode: "subnet-2", EndNode: "ec2-2", Properties: map[string]interface{}{}})

	// Find all nodes within 1 hop of VPC
	result := db.FindImpactRadius("vpc-1", 1)
	if len(result.Nodes) != 3 { // VPC + 2 subnets
		t.Errorf("Expected 3 nodes within 1 hop, got %d", len(result.Nodes))
	}

	// Find all nodes within 2 hops of VPC
	result = db.FindImpactRadius("vpc-1", 2)
	if len(result.Nodes) != 5 { // VPC + 2 subnets + 2 EC2s
		t.Errorf("Expected 5 nodes within 2 hops, got %d", len(result.Nodes))
	}

	// Verify distances
	if result.Distances["vpc-1"] != 0 {
		t.Error("Expected VPC distance to be 0")
	}

	if result.Distances["subnet-1"] != 1 || result.Distances["subnet-2"] != 1 {
		t.Error("Expected subnet distances to be 1")
	}

	if result.Distances["ec2-1"] != 2 || result.Distances["ec2-2"] != 2 {
		t.Error("Expected EC2 distances to be 2")
	}
}

// TestFindImpactRadiusNonExistent tests FindImpactRadius with non-existent node
func TestFindImpactRadiusNonExistent(t *testing.T) {
	db := NewGraphDatabase()

	result := db.FindImpactRadius("non-existent", 1)
	if len(result.Nodes) != 0 {
		t.Errorf("Expected 0 nodes for non-existent node, got %d", len(result.Nodes))
	}
}

// TestFindDependencies tests finding dependencies
func TestFindDependencies(t *testing.T) {
	db := NewGraphDatabase()

	// Create structure: EC2 -> Subnet -> VPC
	ec2 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	subnet := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	vpc := &Node{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}}

	db.AddNode(ec2)
	db.AddNode(subnet)
	db.AddNode(vpc)

	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "subnet-1", EndNode: "vpc-1", Properties: map[string]interface{}{}})

	// Find dependencies of EC2
	deps := db.FindDependencies("ec2-1", 2)
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies for EC2, got %d", len(deps))
	}
}

// TestFindDependents tests finding nodes that depend on a given node
func TestFindDependents(t *testing.T) {
	db := NewGraphDatabase()

	// Create structure: EC2 -> Subnet -> VPC
	// Means: EC2 depends on Subnet, Subnet depends on VPC
	// So: Subnet is dependent of EC2, VPC is dependent of Subnet

	ec2 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	subnet := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	vpc := &Node{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}}

	db.AddNode(ec2)
	db.AddNode(subnet)
	db.AddNode(vpc)

	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "subnet-1", EndNode: "vpc-1", Properties: map[string]interface{}{}})

	// Find dependents of VPC (nodes that would be affected if VPC changed)
	dependents := db.FindDependents("vpc-1", 2)
	if len(dependents) != 2 {
		t.Errorf("Expected 2 dependents of VPC, got %d", len(dependents))
	}
}

// TestFindCriticalPaths tests identifying critical nodes with many dependents
func TestFindCriticalPaths(t *testing.T) {
	db := NewGraphDatabase()

	// Create structure where VPC is critical
	// EC2-1 -> VPC
	// EC2-2 -> VPC
	// Subnet -> VPC

	ec2_1 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	ec2_2 := &Node{ID: "ec2-2", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	subnet := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	vpc := &Node{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}}

	db.AddNode(ec2_1)
	db.AddNode(ec2_2)
	db.AddNode(subnet)
	db.AddNode(vpc)

	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "vpc-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "ec2-2", EndNode: "vpc-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-3", Type: DEPENDS_ON, StartNode: "subnet-1", EndNode: "vpc-1", Properties: map[string]interface{}{}})

	// Find nodes with at least 3 incoming relationships (critical)
	critical := db.FindCriticalPaths(3)
	if len(critical) != 1 {
		t.Errorf("Expected 1 critical node, got %d", len(critical))
	}

	if critical[0].ID != "vpc-1" {
		t.Errorf("Expected VPC to be critical node, got %s", critical[0].ID)
	}

	// Find nodes with at least 1 incoming relationship
	lessCritical := db.FindCriticalPaths(1)
	if len(lessCritical) != 1 {
		t.Errorf("Expected 1 node with at least 1 dependent, got %d", len(lessCritical))
	}
}

// TestMatch tests pattern matching query
func TestMatch(t *testing.T) {
	db := NewGraphDatabase()

	// Create EC2 instances and subnets
	ec2_1 := &Node{
		ID:     "ec2-1",
		Labels: []string{"EC2", "Resource"},
		Properties: map[string]interface{}{
			"instance_type": "t3.micro",
		},
	}

	ec2_2 := &Node{
		ID:     "ec2-2",
		Labels: []string{"EC2", "Resource"},
		Properties: map[string]interface{}{
			"instance_type": "t3.large",
		},
	}

	subnet := &Node{
		ID:     "subnet-1",
		Labels: []string{"Subnet", "Resource"},
		Properties: map[string]interface{}{
			"id": "subnet-123",
		},
	}

	db.AddNode(ec2_1)
	db.AddNode(ec2_2)
	db.AddNode(subnet)

	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "ec2-2", EndNode: "subnet-1", Properties: map[string]interface{}{}})

	// Match pattern: EC2 -DEPENDS_ON-> Subnet
	pattern := &MatchPattern{
		StartLabels: []string{"EC2"},
		RelType:     DEPENDS_ON,
		EndLabels:   []string{"Subnet"},
		EndFilter:   map[string]interface{}{},
	}

	results := db.Match(pattern)
	if len(results) != 2 {
		t.Errorf("Expected 2 pattern matches, got %d", len(results))
	}

	// Match pattern with property filter
	pattern2 := &MatchPattern{
		StartLabels: []string{"EC2"},
		RelType:     DEPENDS_ON,
		EndLabels:   []string{"Subnet"},
		EndFilter: map[string]interface{}{
			"id": "subnet-123",
		},
	}

	results2 := db.Match(pattern2)
	if len(results2) != 2 {
		t.Errorf("Expected 2 pattern matches with filter, got %d", len(results2))
	}
}

// TestMatchNoLabels tests pattern matching with no label constraints
func TestMatchNoLabels(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	db.AddNode(&Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "node-3", Labels: []string{"C"}, Properties: map[string]interface{}{}})

	// Create relationships
	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "node-1", EndNode: "node-2", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "node-2", EndNode: "node-3", Properties: map[string]interface{}{}})

	// Match pattern with no label constraints
	pattern := &MatchPattern{
		StartLabels: []string{},
		RelType:     DEPENDS_ON,
		EndLabels:   []string{},
		EndFilter:   map[string]interface{}{},
	}

	results := db.Match(pattern)
	if len(results) != 2 {
		t.Errorf("Expected 2 pattern matches with no label constraint, got %d", len(results))
	}
}

// TestConcurrentAccess tests thread-safety with concurrent operations
func TestConcurrentAccess(t *testing.T) {
	db := NewGraphDatabase()

	// Pre-create nodes for relationships
	for i := 0; i < 100; i++ {
		db.AddNode(&Node{
			ID:     "node-" + string(rune('0'+(i%10))),
			Labels: []string{"Test"},
			Properties: map[string]interface{}{
				"index": i,
			},
		})
	}

	done := make(chan bool, 100)

	// Concurrent writes
	for i := 0; i < 100; i++ {
		go func(index int) {
			// Add nodes
			db.AddNode(&Node{
				ID:     "concurrent-node-" + string(rune('0'+(index%10))),
				Labels: []string{"ConcurrentTest"},
				Properties: map[string]interface{}{
					"index": index,
				},
			})

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify counts (some nodes may be overwritten since they share IDs)
	if db.NodeCount() == 0 {
		t.Error("Expected some nodes to be added")
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		go func(index int) {
			_ = db.GetNode("node-1")
			_ = db.GetAllNodes()
			done <- true
		}(i)
	}

	for i := 0; i < 50; i++ {
		<-done
	}
}

// TestMultipleLabels tests nodes with multiple labels
func TestMultipleLabels(t *testing.T) {
	db := NewGraphDatabase()

	node := &Node{
		ID:     "complex-node",
		Labels: []string{"Resource", "EC2", "Compute", "Drifted"},
		Properties: map[string]interface{}{
			"name": "web-server",
		},
	}

	db.AddNode(node)

	// Test retrieving by different labels
	if !db.HasLabel("complex-node", "Resource") {
		t.Error("Expected node to have Resource label")
	}

	if !db.HasLabel("complex-node", "Compute") {
		t.Error("Expected node to have Compute label")
	}

	if !db.HasLabel("complex-node", "Drifted") {
		t.Error("Expected node to have Drifted label")
	}

	// Test GetNodesByLabel for each label
	resourceNodes := db.GetNodesByLabel("Resource")
	if len(resourceNodes) != 1 {
		t.Errorf("Expected 1 Resource node, got %d", len(resourceNodes))
	}

	computeNodes := db.GetNodesByLabel("Compute")
	if len(computeNodes) != 1 {
		t.Errorf("Expected 1 Compute node, got %d", len(computeNodes))
	}
}

// TestRelationshipProperties tests that relationship properties are preserved
func TestRelationshipProperties(t *testing.T) {
	db := NewGraphDatabase()

	db.AddNode(&Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}})

	rel := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "node-1",
		EndNode:   "node-2",
		Properties: map[string]interface{}{
			"weight":      5,
			"description": "network dependency",
			"since":       "2024-01-01",
		},
	}

	db.AddRelationship(rel)

	retrieved := db.GetRelationship("rel-1")
	if retrieved.Properties["weight"] != 5 {
		t.Errorf("Expected weight 5, got %v", retrieved.Properties["weight"])
	}

	if retrieved.Properties["description"] != "network dependency" {
		t.Errorf("Expected description 'network dependency', got %v", retrieved.Properties["description"])
	}

	if retrieved.Properties["since"] != "2024-01-01" {
		t.Errorf("Expected since '2024-01-01', got %v", retrieved.Properties["since"])
	}
}

// TestNodeProperties tests that node properties are preserved
func TestNodeProperties(t *testing.T) {
	db := NewGraphDatabase()

	props := map[string]interface{}{
		"name":          "web-server",
		"instance_type": "t3.micro",
		"tags": map[string]interface{}{
			"Environment": "production",
			"Team":        "backend",
		},
		"availability_zone": "us-east-1a",
		"public_ip":         "203.0.113.42",
	}

	node := &Node{
		ID:         "ec2-123",
		Labels:     []string{"EC2", "Resource"},
		Properties: props,
	}

	db.AddNode(node)
	retrieved := db.GetNode("ec2-123")

	// Verify all properties
	if retrieved.Properties["name"] != "web-server" {
		t.Error("Name property mismatch")
	}

	if retrieved.Properties["instance_type"] != "t3.micro" {
		t.Error("Instance type property mismatch")
	}

	tags := retrieved.Properties["tags"].(map[string]interface{})
	if tags["Environment"] != "production" {
		t.Error("Tags property mismatch")
	}

	if retrieved.Properties["availability_zone"] != "us-east-1a" {
		t.Error("Availability zone property mismatch")
	}
}

// TestEmptyGraph tests operations on empty graph
func TestEmptyGraph(t *testing.T) {
	db := NewGraphDatabase()

	if db.NodeCount() != 0 {
		t.Error("Expected empty graph to have 0 nodes")
	}

	if db.RelationshipCount() != 0 {
		t.Error("Expected empty graph to have 0 relationships")
	}

	if len(db.GetAllNodes()) != 0 {
		t.Error("Expected GetAllNodes to return empty slice")
	}

	if len(db.GetAllRelationships()) != 0 {
		t.Error("Expected GetAllRelationships to return empty slice")
	}

	// Try to retrieve non-existent node
	if db.GetNode("non-existent") != nil {
		t.Error("Expected GetNode to return nil for non-existent node")
	}

	// Try to retrieve nodes by label
	if len(db.GetNodesByLabel("NonExistent")) != 0 {
		t.Error("Expected GetNodesByLabel to return empty slice for non-existent label")
	}
}

// TestSpecialCharactersInIDs tests handling of special characters in IDs
func TestSpecialCharactersInIDs(t *testing.T) {
	db := NewGraphDatabase()

	specialIDs := []string{
		"arn:aws:ec2:us-east-1:123456789012:instance/i-0123456789abcdef0",
		"node-with-dashes",
		"node.with.dots",
		"node_with_underscores",
		"node:with:colons",
	}

	for _, id := range specialIDs {
		node := &Node{
			ID:     id,
			Labels: []string{"Test"},
			Properties: map[string]interface{}{
				"id": id,
			},
		}
		db.AddNode(node)
	}

	// Verify all nodes are retrievable
	for _, id := range specialIDs {
		node := db.GetNode(id)
		if node == nil {
			t.Errorf("Failed to retrieve node with ID: %s", id)
		}

		if node.Properties["id"] != id {
			t.Errorf("Node ID property mismatch for: %s", id)
		}
	}

	if db.NodeCount() != len(specialIDs) {
		t.Errorf("Expected %d nodes, got %d", len(specialIDs), db.NodeCount())
	}
}

// TestFindPathBidirectional tests path finding in both directions
func TestFindPathBidirectional(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes
	node1 := &Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}}
	node2 := &Node{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}}
	node3 := &Node{ID: "node-3", Labels: []string{"C"}, Properties: map[string]interface{}{}}

	db.AddNode(node1)
	db.AddNode(node2)
	db.AddNode(node3)

	// Create relationships (bidirectional path)
	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "node-1", EndNode: "node-2", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: PART_OF, StartNode: "node-2", EndNode: "node-3", Properties: map[string]interface{}{}})

	// Test path from node-1 to node-3
	path, err := db.FindPath("node-1", "node-3")
	if err != nil {
		t.Fatalf("Failed to find path from node-1 to node-3: %v", err)
	}

	if path.Length != 2 {
		t.Errorf("Expected path length 2, got %d", path.Length)
	}

	// Test reverse path (should work due to bidirectional traversal)
	path2, err := db.FindPath("node-3", "node-1")
	if err != nil {
		t.Fatalf("Failed to find reverse path from node-3 to node-1: %v", err)
	}

	if path2.Length != 2 {
		t.Errorf("Expected reverse path length 2, got %d", path2.Length)
	}
}

// TestMatchWithMultipleFilters tests pattern matching with multiple property filters
func TestMatchWithMultipleFilters(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes with different properties
	nodes := []*Node{
		{
			ID:     "node-1",
			Labels: []string{"Resource", "EC2"},
			Properties: map[string]interface{}{
				"type": "compute",
				"size": "large",
			},
		},
		{
			ID:     "node-2",
			Labels: []string{"Resource", "EC2"},
			Properties: map[string]interface{}{
				"type": "compute",
				"size": "small",
			},
		},
		{
			ID:     "node-3",
			Labels: []string{"Resource", "Storage"},
			Properties: map[string]interface{}{
				"type": "storage",
				"size": "large",
			},
		},
	}

	for _, n := range nodes {
		db.AddNode(n)
	}

	target := &Node{
		ID:     "target-1",
		Labels: []string{"Target", "Compute"},
		Properties: map[string]interface{}{
			"type": "compute",
		},
	}
	db.AddNode(target)

	// Create relationships
	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "node-1", EndNode: "target-1", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "node-2", EndNode: "target-1", Properties: map[string]interface{}{}})

	// Match pattern with type filter
	pattern := &MatchPattern{
		StartLabels: []string{"EC2"},
		RelType:     DEPENDS_ON,
		EndLabels:   []string{"Compute"},
		EndFilter: map[string]interface{}{
			"type": "compute",
		},
	}

	results := db.Match(pattern)
	if len(results) != 2 {
		t.Errorf("Expected 2 matches with type filter, got %d", len(results))
	}
}

// TestFindDependenciesWithoutOutgoing tests finding dependencies when node has no outgoing relationships
func TestFindDependenciesWithoutOutgoing(t *testing.T) {
	db := NewGraphDatabase()

	// Create isolated node
	db.AddNode(&Node{ID: "isolated", Labels: []string{"Test"}, Properties: map[string]interface{}{}})

	// Find dependencies (should be empty)
	deps := db.FindDependencies("isolated", 2)
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies for isolated node, got %d", len(deps))
	}
}

// TestFindDependentsWithoutIncoming tests finding dependents when node has no incoming relationships
func TestFindDependentsWithoutIncoming(t *testing.T) {
	db := NewGraphDatabase()

	// Create isolated node
	db.AddNode(&Node{ID: "isolated", Labels: []string{"Test"}, Properties: map[string]interface{}{}})

	// Find dependents (should be empty)
	dependents := db.FindDependents("isolated", 2)
	if len(dependents) != 0 {
		t.Errorf("Expected 0 dependents for isolated node, got %d", len(dependents))
	}
}

// TestFindDependenciesWithCircularReferences tests handling of circular dependencies
func TestFindDependenciesWithCircularReferences(t *testing.T) {
	db := NewGraphDatabase()

	// Create circular chain: A -> B -> C -> A
	db.AddNode(&Node{ID: "a", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "b", Labels: []string{"B"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "c", Labels: []string{"C"}, Properties: map[string]interface{}{}})

	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "a", EndNode: "b", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "b", EndNode: "c", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-3", Type: DEPENDS_ON, StartNode: "c", EndNode: "a", Properties: map[string]interface{}{}})

	// Find dependencies from A (with depth limit to prevent infinite loop)
	deps := db.FindDependencies("a", 3)
	// Should find B and C (visited map prevents revisiting A)
	if len(deps) < 2 {
		t.Errorf("Expected at least 2 dependencies with circular chain, got %d", len(deps))
	}
}

// TestImpactRadiusDistances tests that distances are correctly calculated in impact radius
func TestImpactRadiusDistances(t *testing.T) {
	db := NewGraphDatabase()

	// Create a star topology with center node
	db.AddNode(&Node{ID: "center", Labels: []string{"Center"}, Properties: map[string]interface{}{}})
	for i := 1; i <= 5; i++ {
		id := "leaf-" + string(rune('0'+i))
		db.AddNode(&Node{ID: id, Labels: []string{"Leaf"}, Properties: map[string]interface{}{}})
		db.AddRelationship(&Relationship{
			ID:        "rel-" + string(rune('0'+i)),
			Type:      CONNECTS_TO,
			StartNode: "center",
			EndNode:   id,
			Properties: map[string]interface{}{},
		})
	}

	result := db.FindImpactRadius("center", 1)

	// Verify center is at distance 0
	if result.Distances["center"] != 0 {
		t.Error("Center node should have distance 0")
	}

	// Verify all leaves are at distance 1
	for i := 1; i <= 5; i++ {
		leafID := "leaf-" + string(rune('0'+i))
		if result.Distances[leafID] != 1 {
			t.Errorf("Leaf node %s should have distance 1, got %d", leafID, result.Distances[leafID])
		}
	}
}

// TestFindPathSelfLoop tests finding path from node to itself
func TestFindPathSelfLoop(t *testing.T) {
	db := NewGraphDatabase()

	node := &Node{ID: "self", Labels: []string{"Test"}, Properties: map[string]interface{}{}}
	db.AddNode(node)

	// Path from a node to itself should have length 0 and contain only that node
	path, err := db.FindPath("self", "self")
	if err != nil {
		t.Fatalf("Should be able to find path from node to itself: %v", err)
	}

	if path.Length != 0 {
		t.Errorf("Expected path length 0 for self-loop, got %d", path.Length)
	}

	if len(path.Nodes) != 1 {
		t.Errorf("Expected 1 node in self-loop path, got %d", len(path.Nodes))
	}
}

// TestAddNodeOverwrite tests that adding a node with same ID overwrites the previous one
func TestAddNodeOverwrite(t *testing.T) {
	db := NewGraphDatabase()

	node1 := &Node{
		ID:     "node-1",
		Labels: []string{"OldLabel"},
		Properties: map[string]interface{}{
			"version": 1,
		},
	}

	node2 := &Node{
		ID:     "node-1",
		Labels: []string{"NewLabel"},
		Properties: map[string]interface{}{
			"version": 2,
		},
	}

	db.AddNode(node1)
	db.AddNode(node2)

	// Should have only 1 node
	if db.NodeCount() != 1 {
		t.Errorf("Expected 1 node after overwrite, got %d", db.NodeCount())
	}

	// The new node should have replaced the old one
	retrieved := db.GetNode("node-1")
	if retrieved.Labels[0] != "NewLabel" {
		t.Error("Expected node to have NewLabel after overwrite")
	}

	if retrieved.Properties["version"] != 2 {
		t.Error("Expected node properties to be updated after overwrite")
	}
}

// TestMatchNoStartLabels tests pattern matching without start label constraints
func TestMatchNoStartLabels(t *testing.T) {
	db := NewGraphDatabase()

	// Create nodes with different labels
	db.AddNode(&Node{ID: "a", Labels: []string{"TypeA"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "b", Labels: []string{"TypeB"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "target", Labels: []string{"TypeC"}, Properties: map[string]interface{}{}})

	db.AddRelationship(&Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "a", EndNode: "target", Properties: map[string]interface{}{}})
	db.AddRelationship(&Relationship{ID: "rel-2", Type: DEPENDS_ON, StartNode: "b", EndNode: "target", Properties: map[string]interface{}{}})

	// Match pattern with no start labels but specific end labels
	pattern := &MatchPattern{
		StartLabels: []string{},
		RelType:     DEPENDS_ON,
		EndLabels:   []string{"TypeC"},
		EndFilter:   map[string]interface{}{},
	}

	results := db.Match(pattern)
	if len(results) != 2 {
		t.Errorf("Expected 2 matches without start labels, got %d", len(results))
	}
}

// TestGetOutgoingRelationshipsEmpty tests getting outgoing relationships for node with none
func TestGetOutgoingRelationshipsEmpty(t *testing.T) {
	db := NewGraphDatabase()

	db.AddNode(&Node{ID: "node-1", Labels: []string{"Test"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "node-2", Labels: []string{"Test"}, Properties: map[string]interface{}{}})

	// Get outgoing relationships from node-1 (should be empty)
	outgoing := db.GetOutgoingRelationships("node-1")
	if len(outgoing) != 0 {
		t.Errorf("Expected 0 outgoing relationships, got %d", len(outgoing))
	}

	// Also test for non-existent node (should return empty, not nil)
	outgoing = db.GetOutgoingRelationships("non-existent")
	if outgoing == nil {
		t.Error("Expected empty slice, got nil")
	}
}

// TestGetIncomingRelationshipsEmpty tests getting incoming relationships for node with none
func TestGetIncomingRelationshipsEmpty(t *testing.T) {
	db := NewGraphDatabase()

	db.AddNode(&Node{ID: "node-1", Labels: []string{"Test"}, Properties: map[string]interface{}{}})

	// Get incoming relationships from node-1 (should be empty)
	incoming := db.GetIncomingRelationships("node-1")
	if len(incoming) != 0 {
		t.Errorf("Expected 0 incoming relationships, got %d", len(incoming))
	}

	// Also test for non-existent node (should return empty, not nil)
	incoming = db.GetIncomingRelationships("non-existent")
	if incoming == nil {
		t.Error("Expected empty slice, got nil")
	}
}

// TestDeleteNodeNonExistent tests deleting a non-existent node
func TestDeleteNodeNonExistent(t *testing.T) {
	db := NewGraphDatabase()

	// Add a node first
	db.AddNode(&Node{ID: "node-1", Labels: []string{"Test"}, Properties: map[string]interface{}{}})

	// Delete non-existent node (should be safe)
	db.DeleteNode("non-existent")

	// Original node should still exist
	if db.NodeCount() != 1 {
		t.Error("Original node should still exist after deleting non-existent node")
	}
}

// TestNodeDeletionWithMultipleRelationships tests deleting node with many relationships
func TestNodeDeletionWithMultipleRelationships(t *testing.T) {
	db := NewGraphDatabase()

	// Create a hub node connected to many others
	hub := &Node{ID: "hub", Labels: []string{"Hub"}, Properties: map[string]interface{}{}}
	db.AddNode(hub)

	// Create and connect 5 other nodes
	for i := 1; i <= 5; i++ {
		id := "node-" + string(rune('0'+i))
		db.AddNode(&Node{ID: id, Labels: []string{"Node"}, Properties: map[string]interface{}{}})

		// Outgoing from hub
		db.AddRelationship(&Relationship{
			ID:        "out-" + id,
			Type:      DEPENDS_ON,
			StartNode: "hub",
			EndNode:   id,
			Properties: map[string]interface{}{},
		})

		// Incoming to hub
		db.AddRelationship(&Relationship{
			ID:        "in-" + id,
			Type:      DEPENDS_ON,
			StartNode: id,
			EndNode:   "hub",
			Properties: map[string]interface{}{},
		})
	}

	if db.RelationshipCount() != 10 {
		t.Errorf("Expected 10 relationships before deletion, got %d", db.RelationshipCount())
	}

	// Delete the hub
	db.DeleteNode("hub")

	// All relationships should be deleted
	if db.RelationshipCount() != 0 {
		t.Errorf("Expected 0 relationships after deletion, got %d", db.RelationshipCount())
	}

	// 5 other nodes should still exist
	if db.NodeCount() != 5 {
		t.Errorf("Expected 5 nodes after deletion, got %d", db.NodeCount())
	}
}

// ============================================================================
// BUILDER TESTS - Store and graph building
// ============================================================================

// TestNewStore tests Store creation
func TestNewStore(t *testing.T) {
	store := NewStore()

	if store == nil {
		t.Fatal("Expected non-nil store")
	}

	if len(store.GetDrifts()) != 0 {
		t.Errorf("Expected 0 drifts initially, got %d", len(store.GetDrifts()))
	}

	if len(store.GetEvents()) != 0 {
		t.Errorf("Expected 0 events initially, got %d", len(store.GetEvents()))
	}

	if len(store.GetUnmanaged()) != 0 {
		t.Errorf("Expected 0 unmanaged initially, got %d", len(store.GetUnmanaged()))
	}
}

// TestStoreDriftOperations tests drift management in Store
func TestStoreDriftOperations(t *testing.T) {
	store := NewStore()

	drift := types.DriftAlert{
		ResourceID:   "res-1",
		ResourceType: "aws_instance",
		ResourceName: "test-instance",
		Severity:     "high",
		Attribute:    "state",
		OldValue:     "running",
		NewValue:     "stopped",
		UserIdentity: types.UserIdentity{
			UserName: "test-user",
			ARN:      "arn:aws:iam::123456789012:user/test",
		},
		Timestamp: "2024-01-01T00:00:00Z",
		AlertType: "drift",
	}

	store.AddDrift(drift)

	drifts := store.GetDrifts()
	if len(drifts) != 1 {
		t.Errorf("Expected 1 drift, got %d", len(drifts))
	}

	if drifts[0].ResourceID != "res-1" {
		t.Errorf("Expected resource ID 'res-1', got '%s'", drifts[0].ResourceID)
	}
}

// TestStoreEventOperations tests event management in Store
func TestStoreEventOperations(t *testing.T) {
	store := NewStore()

	event := types.Event{
		ResourceID:   "res-1",
		ResourceType: "aws_s3_bucket",
		EventName:    "PutBucketEncryption",
		Provider:     "aws",
		UserIdentity: types.UserIdentity{
			UserName: "test-user",
			ARN:      "arn:aws:iam::123456789012:user/test",
		},
		Region: "us-east-1",
	}

	store.AddEvent(event)

	events := store.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].EventName != "PutBucketEncryption" {
		t.Errorf("Expected event name 'PutBucketEncryption', got '%s'", events[0].EventName)
	}
}

// TestStoreUnmanagedOperations tests unmanaged resource management
func TestStoreUnmanagedOperations(t *testing.T) {
	store := NewStore()

	unmanaged := types.UnmanagedResourceAlert{
		ResourceID:   "res-1",
		ResourceType: "aws_instance",
		EventName:    "RunInstances",
		Severity:     "medium",
		UserIdentity: types.UserIdentity{
			UserName: "test-user",
			ARN:      "arn:aws:iam::123456789012:user/test",
		},
		Reason:    "Not in Terraform state",
		Timestamp: "2024-01-01T00:00:00Z",
	}

	store.AddUnmanaged(unmanaged)

	unmanageds := store.GetUnmanaged()
	if len(unmanageds) != 1 {
		t.Errorf("Expected 1 unmanaged resource, got %d", len(unmanageds))
	}

	if unmanageds[0].Reason != "Not in Terraform state" {
		t.Errorf("Expected reason 'Not in Terraform state', got '%s'", unmanageds[0].Reason)
	}
}

// TestStoreClear tests clearing store data
func TestStoreClear(t *testing.T) {
	store := NewStore()

	store.AddDrift(types.DriftAlert{ResourceID: "res-1", ResourceType: "aws_instance", Severity: "high"})
	store.AddEvent(types.Event{ResourceID: "res-2", ResourceType: "aws_s3_bucket", EventName: "Put"})
	store.AddUnmanaged(types.UnmanagedResourceAlert{ResourceID: "res-3", ResourceType: "aws_instance", Severity: "medium"})

	if len(store.GetDrifts()) != 1 || len(store.GetEvents()) != 1 || len(store.GetUnmanaged()) != 1 {
		t.Fatal("Failed to add data to store")
	}

	store.Clear()

	if len(store.GetDrifts()) != 0 {
		t.Errorf("Expected 0 drifts after clear, got %d", len(store.GetDrifts()))
	}

	if len(store.GetEvents()) != 0 {
		t.Errorf("Expected 0 events after clear, got %d", len(store.GetEvents()))
	}

	if len(store.GetUnmanaged()) != 0 {
		t.Errorf("Expected 0 unmanaged after clear, got %d", len(store.GetUnmanaged()))
	}
}

// TestStoreGetGraphDB tests graph database retrieval
func TestStoreGetGraphDB(t *testing.T) {
	store := NewStore()
	db := store.GetGraphDB()

	if db == nil {
		t.Fatal("Expected non-nil GraphDatabase")
	}

	if db.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", db.NodeCount())
	}
}

// TestStoreGetStats tests statistics retrieval
func TestStoreGetStats(t *testing.T) {
	store := NewStore()

	store.AddDrift(types.DriftAlert{
		ResourceID:   "res-1",
		ResourceType: "aws_instance",
		Severity:     "high",
	})
	store.AddDrift(types.DriftAlert{
		ResourceID:   "res-2",
		ResourceType: "aws_s3_bucket",
		Severity:     "critical",
	})
	store.AddEvent(types.Event{ResourceID: "res-3", ResourceType: "aws_lambda", EventName: "UpdateFunction"})

	stats := store.GetStats()

	if stats["total_drifts"] != 2 {
		t.Errorf("Expected 2 total drifts, got %v", stats["total_drifts"])
	}

	if stats["total_events"] != 1 {
		t.Errorf("Expected 1 total event, got %v", stats["total_events"])
	}

	severityCounts := stats["severity_counts"].(map[string]int)
	if severityCounts["high"] != 1 {
		t.Errorf("Expected 1 high severity drift, got %d", severityCounts["high"])
	}
}

// TestPopulateSampleData tests sample data population
func TestPopulateSampleData(t *testing.T) {
	store := NewStore()
	store.PopulateSampleData()

	if len(store.GetDrifts()) < 3 {
		t.Errorf("Expected at least 3 sample drifts, got %d", len(store.GetDrifts()))
	}

	if len(store.GetEvents()) < 2 {
		t.Errorf("Expected at least 2 sample events, got %d", len(store.GetEvents()))
	}

	if len(store.GetUnmanaged()) < 1 {
		t.Errorf("Expected at least 1 sample unmanaged resource, got %d", len(store.GetUnmanaged()))
	}
}

// ============================================================================
// CYTOSCAPE CONVERSION TESTS
// ============================================================================

// TestConvertDriftToCytoscape tests drift to cytoscape conversion
func TestConvertDriftToCytoscape(t *testing.T) {
	drift := types.DriftAlert{
		ResourceID:   "sg-123",
		ResourceName: "web-sg",
		ResourceType: "aws_security_group",
		Severity:     "critical",
		Attribute:    "ingress_rules",
		OldValue:     "restricted",
		NewValue:     "open",
		UserIdentity: types.UserIdentity{
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
		MatchedRules: []string{"rule1", "rule2"},
		Timestamp:    "2024-01-01T00:00:00Z",
		AlertType:    "drift",
	}

	node := ConvertDriftToCytoscape(drift)

	if node.Data.ID != "sg-123" {
		t.Errorf("Expected ID 'sg-123', got '%s'", node.Data.ID)
	}

	if node.Data.Type != "drift" {
		t.Errorf("Expected type 'drift', got '%s'", node.Data.Type)
	}

	if node.Data.Severity != "critical" {
		t.Errorf("Expected severity 'critical', got '%s'", node.Data.Severity)
	}

	if node.Data.Metadata["attribute"] != "ingress_rules" {
		t.Errorf("Expected attribute 'ingress_rules', got '%v'", node.Data.Metadata["attribute"])
	}
}

// TestConvertEventToCytoscape tests event to cytoscape conversion
func TestConvertEventToCytoscape(t *testing.T) {
	event := types.Event{
		ResourceID:   "i-456",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		Provider:     "aws",
		UserIdentity: types.UserIdentity{
			UserName: "developer",
			ARN:      "arn:aws:iam::123456789012:user/developer",
		},
		Region:      "us-east-1",
		ServiceName: "ec2",
		Changes: map[string]interface{}{
			"imageId":      "ami-123",
			"instanceType": "t3.medium",
		},
	}

	node := ConvertEventToCytoscape(event)

	if node.Data.ID != "i-456" {
		t.Errorf("Expected ID 'i-456', got '%s'", node.Data.ID)
	}

	if node.Data.Type != "falco_event" {
		t.Errorf("Expected type 'falco_event', got '%s'", node.Data.Type)
	}

	if node.Data.Metadata["event_name"] != "RunInstances" {
		t.Errorf("Expected event_name 'RunInstances', got '%v'", node.Data.Metadata["event_name"])
	}
}

// TestConvertUnmanagedToCytoscape tests unmanaged resource to cytoscape conversion
func TestConvertUnmanagedToCytoscape(t *testing.T) {
	unmanaged := types.UnmanagedResourceAlert{
		ResourceID:   "i-789",
		ResourceType: "aws_instance",
		EventName:    "RunInstances",
		Severity:     "medium",
		UserIdentity: types.UserIdentity{
			UserName: "ops",
			ARN:      "arn:aws:iam::123456789012:user/ops",
		},
		Timestamp: "2024-01-01T00:00:00Z",
		Reason:    "Created manually outside Terraform",
		Changes: map[string]interface{}{
			"state": "running",
		},
	}

	node := ConvertUnmanagedToCytoscape(unmanaged)

	if node.Data.ID != "i-789" {
		t.Errorf("Expected ID 'i-789', got '%s'", node.Data.ID)
	}

	if node.Data.Type != "unmanaged" {
		t.Errorf("Expected type 'unmanaged', got '%s'", node.Data.Type)
	}

	if node.Data.Metadata["reason"] != "Created manually outside Terraform" {
		t.Errorf("Expected reason 'Created manually outside Terraform', got '%v'", node.Data.Metadata["reason"])
	}
}

// TestExtractResourceIDFromAttributes tests ID extraction
func TestExtractResourceIDFromAttributes(t *testing.T) {
	tests := []struct {
		name       string
		attributes map[string]interface{}
		expected   string
	}{
		{
			name: "with id attribute",
			attributes: map[string]interface{}{
				"id": "i-123456",
			},
			expected: "i-123456",
		},
		{
			name: "with arn attribute",
			attributes: map[string]interface{}{
				"arn": "arn:aws:s3:::my-bucket",
			},
			expected: "arn:aws:s3:::my-bucket",
		},
		{
			name: "with name attribute",
			attributes: map[string]interface{}{
				"name": "my-resource",
			},
			expected: "my-resource",
		},
		{
			name: "with self_link attribute",
			attributes: map[string]interface{}{
				"self_link": "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/default",
			},
			expected: "https://www.googleapis.com/compute/v1/projects/my-project/global/networks/default",
		},
		{
			name:       "no identifying attributes",
			attributes: map[string]interface{}{},
			expected:   "",
		},
		{
			name: "id takes precedence",
			attributes: map[string]interface{}{
				"id":   "i-123456",
				"arn":  "arn:aws:ec2:us-east-1:123456789012:instance/i-123456",
				"name": "instance-name",
			},
			expected: "i-123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractResourceIDFromAttributes(tt.attributes)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestExtractResourceName tests resource name extraction
func TestExtractResourceName(t *testing.T) {
	tests := []struct {
		name     string
		resource *terraform.Resource
		expected string
	}{
		{
			name: "with name attribute",
			resource: &terraform.Resource{
				Name: "tf-resource",
				Type: "aws_instance",
				Attributes: map[string]interface{}{
					"name": "my-instance",
				},
			},
			expected: "my-instance",
		},
		{
			name: "with tags Name",
			resource: &terraform.Resource{
				Name: "tf-resource",
				Type: "aws_instance",
				Attributes: map[string]interface{}{
					"tags": map[string]interface{}{
						"Name": "tagged-instance",
					},
				},
			},
			expected: "tagged-instance",
		},
		{
			name: "fallback to tf name",
			resource: &terraform.Resource{
				Name: "tf-resource",
				Type: "aws_instance",
				Attributes: map[string]interface{}{},
			},
			expected: "tf-resource",
		},
		{
			name: "fallback to resource type",
			resource: &terraform.Resource{
				Name: "",
				Type: "aws_instance",
				Attributes: map[string]interface{}{},
			},
			expected: "aws_instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractResourceName(tt.resource)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestCreateEdge tests edge creation
func TestCreateEdge(t *testing.T) {
	edge := CreateEdge("node-1", "node-2", "depends_on", "dependency", "DEPENDS_ON")

	if edge.Data.Source != "node-1" {
		t.Errorf("Expected source 'node-1', got '%s'", edge.Data.Source)
	}

	if edge.Data.Target != "node-2" {
		t.Errorf("Expected target 'node-2', got '%s'", edge.Data.Target)
	}

	if edge.Data.Label != "depends_on" {
		t.Errorf("Expected label 'depends_on', got '%s'", edge.Data.Label)
	}

	if edge.Data.Type != "dependency" {
		t.Errorf("Expected type 'dependency', got '%s'", edge.Data.Type)
	}

	if edge.Data.Relationship != "DEPENDS_ON" {
		t.Errorf("Expected relationship 'DEPENDS_ON', got '%s'", edge.Data.Relationship)
	}
}

// ============================================================================
// HIERARCHY TESTS
// ============================================================================

// TestNewHierarchyBuilder tests hierarchy builder creation
func TestNewHierarchyBuilder(t *testing.T) {
	builder := NewHierarchyBuilder()

	if builder == nil {
		t.Fatal("Expected non-nil HierarchyBuilder")
	}

	if builder.hierarchy == nil {
		t.Fatal("Expected non-nil hierarchy")
	}

	if len(builder.hierarchy.Regions) != 0 {
		t.Errorf("Expected 0 regions initially, got %d", len(builder.hierarchy.Regions))
	}
}

// TestBuildHierarchyWithVPC tests hierarchy building with VPC
func TestBuildHierarchyWithVPC(t *testing.T) {
	builder := NewHierarchyBuilder()

	resources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-12345678",
				"cidr_block": "10.0.0.0/16",
			},
		},
	}

	hierarchy := builder.BuildHierarchy(resources)

	if hierarchy == nil {
		t.Fatal("Expected non-nil hierarchy")
	}

	if len(hierarchy.Regions) == 0 {
		t.Fatal("Expected at least one region in hierarchy")
	}

	// Check if VPC was added
	found := false
	for _, region := range hierarchy.Regions {
		if _, exists := region.VPCs["vpc-12345678"]; exists {
			found = true
			vpc := region.VPCs["vpc-12345678"]
			if vpc.Name != "main-vpc" {
				t.Errorf("Expected VPC name 'main-vpc', got '%s'", vpc.Name)
			}
			if vpc.CIDR != "10.0.0.0/16" {
				t.Errorf("Expected VPC CIDR '10.0.0.0/16', got '%s'", vpc.CIDR)
			}
			break
		}
	}

	if !found {
		t.Fatal("VPC not found in hierarchy")
	}
}

// TestBuildHierarchyWithSubnet tests hierarchy building with subnet
func TestBuildHierarchyWithSubnet(t *testing.T) {
	builder := NewHierarchyBuilder()

	resources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-12345678",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_subnet",
			Name: "public-subnet",
			Attributes: map[string]interface{}{
				"id":                       "subnet-87654321",
				"vpc_id":                   "vpc-12345678",
				"cidr_block":               "10.0.1.0/24",
				"availability_zone":        "us-east-1a",
				"map_public_ip_on_launch":  true,
			},
		},
	}

	hierarchy := builder.BuildHierarchy(resources)

	found := false
	for _, region := range hierarchy.Regions {
		if vpc, exists := region.VPCs["vpc-12345678"]; exists {
			if az, azExists := vpc.AvailabilityZones["us-east-1a"]; azExists {
				if subnet, subnetExists := az.Subnets["subnet-87654321"]; subnetExists {
					found = true
					if subnet.Name != "public-subnet" {
						t.Errorf("Expected subnet name 'public-subnet', got '%s'", subnet.Name)
					}
					if subnet.Type != "public" {
						t.Errorf("Expected subnet type 'public', got '%s'", subnet.Type)
					}
					break
				}
			}
		}
	}

	if !found {
		t.Fatal("Subnet not found in hierarchy")
	}
}

// TestConvertHierarchyToNodes tests converting hierarchy to nodes
func TestConvertHierarchyToNodes(t *testing.T) {
	builder := NewHierarchyBuilder()
	resources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-12345678",
				"cidr_block": "10.0.0.0/16",
			},
		},
	}

	hierarchy := builder.BuildHierarchy(resources)
	nodes := ConvertHierarchyToNodes(hierarchy)

	if len(nodes) == 0 {
		t.Fatal("Expected at least one node from hierarchy conversion")
	}

	// Should have region node and VPC node
	if len(nodes) < 2 {
		t.Errorf("Expected at least 2 nodes (region + VPC), got %d", len(nodes))
	}

	// Check for region node
	hasRegionNode := false
	hasVPCNode := false
	for _, node := range nodes {
		if node.Data.Type == "region-group" {
			hasRegionNode = true
		}
		if node.Data.Type == "vpc-group" {
			hasVPCNode = true
		}
	}

	if !hasRegionNode {
		t.Error("Expected region node in converted hierarchy")
	}
	if !hasVPCNode {
		t.Error("Expected VPC node in converted hierarchy")
	}
}

// ============================================================================
// CONVERTER TESTS
// ============================================================================

// TestTerraformToGraph tests basic Terraform to graph conversion
func TestTerraformToGraph(t *testing.T) {
	resources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main",
			Attributes: map[string]interface{}{
				"id":         "vpc-123",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_subnet",
			Name: "public",
			Attributes: map[string]interface{}{
				"id":         "subnet-123",
				"vpc_id":     "vpc-123",
				"cidr_block": "10.0.1.0/24",
			},
		},
	}

	driftedIDs := map[string]bool{}
	graph := TerraformToGraph(resources, driftedIDs)

	if graph == nil {
		t.Fatal("Expected non-nil graph")
	}

	if graph.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes, got %d", graph.NodeCount())
	}

	// Should have a PART_OF relationship between subnet and VPC
	if graph.RelationshipCount() == 0 {
		t.Error("Expected at least one relationship in graph")
	}
}

// TestTerraformToGraphWithDrift tests Terraform to graph conversion with drifted resources
func TestTerraformToGraphWithDrift(t *testing.T) {
	resources := []*terraform.Resource{
		{
			Type: "aws_instance",
			Name: "web",
			Attributes: map[string]interface{}{
				"id":            "i-123",
				"instance_type": "t2.micro",
			},
		},
	}

	driftedIDs := map[string]bool{"i-123": true}
	graph := TerraformToGraph(resources, driftedIDs)

	node := graph.GetNode("i-123")
	if node == nil {
		t.Fatal("Expected to find node i-123")
	}

	// Should have Drifted label
	hasDriftedLabel := false
	for _, label := range node.Labels {
		if label == "Drifted" {
			hasDriftedLabel = true
			break
		}
	}

	if !hasDriftedLabel {
		t.Error("Expected node to have 'Drifted' label")
	}

	if node.Properties["has_drift"] != true {
		t.Error("Expected has_drift property to be true")
	}
}

// TestConvertTerraformResourceToCytoscape tests Terraform resource to Cytoscape conversion
func TestConvertTerraformResourceToCytoscape(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Name: "web-server",
		Attributes: map[string]interface{}{
			"id":            "i-123",
			"instance_type": "t3.medium",
			"tags": map[string]interface{}{
				"Name": "MyWebServer",
			},
		},
	}

	node := ConvertTerraformResourceToCytoscape(resource, false)

	if node.Data.ID != "i-123" {
		t.Errorf("Expected ID 'i-123', got '%s'", node.Data.ID)
	}

	if node.Data.Type != "terraform_resource" {
		t.Errorf("Expected type 'terraform_resource', got '%s'", node.Data.Type)
	}

	if node.Data.Severity != "low" {
		t.Errorf("Expected severity 'low', got '%s'", node.Data.Severity)
	}
}

// TestConvertTerraformResourceToCytoscapeDrifted tests Cytoscape conversion of drifted resource
func TestConvertTerraformResourceToCytoscapeDrifted(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Name: "web-server",
		Attributes: map[string]interface{}{
			"id": "i-123",
		},
	}

	node := ConvertTerraformResourceToCytoscape(resource, true)

	if node.Data.Type != "terraform_resource_drifted" {
		t.Errorf("Expected type 'terraform_resource_drifted', got '%s'", node.Data.Type)
	}

	if node.Data.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", node.Data.Severity)
	}

	if node.Data.Metadata["has_drift"] != true {
		t.Error("Expected has_drift metadata to be true")
	}
}

// ============================================================================
// ADDITIONAL RELATIONSHIP AND MODEL TESTS
// ============================================================================

// TestAddRelationshipMissingStartNode tests adding relationship with missing start node
func TestAddRelationshipMissingStartNode(t *testing.T) {
	db := NewGraphDatabase()

	endNode := &Node{ID: "node-2", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(endNode)

	rel := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "missing-node",
		EndNode:   "node-2",
		Properties: map[string]interface{}{},
	}

	err := db.AddRelationship(rel)
	if err == nil {
		t.Error("Expected error when adding relationship with missing start node")
	}

	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}

// TestAddRelationshipMissingEndNode tests adding relationship with missing end node
func TestAddRelationshipMissingEndNode(t *testing.T) {
	db := NewGraphDatabase()

	startNode := &Node{ID: "node-1", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(startNode)

	rel := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "node-1",
		EndNode:   "missing-node",
		Properties: map[string]interface{}{},
	}

	err := db.AddRelationship(rel)
	if err == nil {
		t.Error("Expected error when adding relationship with missing end node")
	}

	if err != ErrNodeNotFound {
		t.Errorf("Expected ErrNodeNotFound, got %v", err)
	}
}


// TestConcurrentNodeAddition tests concurrent node additions
func TestConcurrentNodeAddition(t *testing.T) {
	db := NewGraphDatabase()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			for j := 0; j < 10; j++ {
				node := &Node{
					ID:     fmt.Sprintf("node-%d-%d", index, j),
					Labels: []string{"Resource"},
					Properties: map[string]interface{}{
						"index": index,
					},
				}
				db.AddNode(node)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if db.NodeCount() != 100 {
		t.Errorf("Expected 100 nodes from concurrent additions, got %d", db.NodeCount())
	}
}

// TestTraversalEdgeCases tests traversal edge cases
func TestTraversalEdgeCases(t *testing.T) {
	db := NewGraphDatabase()

	// Add isolated node
	db.AddNode(&Node{ID: "isolated", Labels: []string{"Resource"}, Properties: map[string]interface{}{}})

	// Test FindPath with non-existent nodes
	_, err := db.FindPath("non-existent-1", "non-existent-2")
	if err == nil {
		t.Error("Expected error for non-existent nodes")
	}

	// Test FindImpactRadius with non-existent node
	result := db.FindImpactRadius("non-existent", 2)
	if result.Nodes == nil || len(result.Nodes) > 0 {
		t.Error("Expected empty result for non-existent node")
	}
}

// ============================================================================
// EXTENDED BUILDER AND CONVERTER TESTS
// ============================================================================

// TestSetStateManager tests setting state manager
func TestSetStateManager(t *testing.T) {
	store := NewStore()

	// Initially no state manager
	graphDB := store.GetGraphDB()
	if graphDB.NodeCount() != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", graphDB.NodeCount())
	}

	// Mock state manager would go here
	// For now, test that SetStateManager doesn't panic
	store.SetStateManager(nil)

	// Graph should still exist
	graphDB = store.GetGraphDB()
	if graphDB == nil {
		t.Error("Expected graph DB to exist after SetStateManager")
	}
}

// TestRebuildGraphDB tests rebuilding the graph database
func TestRebuildGraphDB(t *testing.T) {
	store := NewStore()

	// Add some data
	store.AddDrift(types.DriftAlert{
		ResourceID:   "res-1",
		ResourceType: "aws_instance",
		Severity:     "high",
	})

	// Calling RebuildGraphDB should not panic
	store.RebuildGraphDB()

	// Graph should still exist
	graphDB := store.GetGraphDB()
	if graphDB == nil {
		t.Error("Expected graph DB to exist after RebuildGraphDB")
	}
}

// TestBuildGraphEmpty tests building graph with empty store
func TestBuildGraphEmpty(t *testing.T) {
	store := NewStore()

	elements := store.BuildGraph()

	if len(elements.Nodes) != 0 {
		t.Errorf("Expected 0 nodes in empty graph, got %d", len(elements.Nodes))
	}

	if len(elements.Edges) != 0 {
		t.Errorf("Expected 0 edges in empty graph, got %d", len(elements.Edges))
	}
}

// TestBuildGraphWithDrifts tests building graph with drifts
func TestBuildGraphWithDrifts(t *testing.T) {
	store := NewStore()

	store.AddDrift(types.DriftAlert{
		ResourceID:   "res-1",
		ResourceType: "aws_instance",
		ResourceName: "instance-1",
		Severity:     "high",
	})

	store.AddDrift(types.DriftAlert{
		ResourceID:   "res-2",
		ResourceType: "aws_s3_bucket",
		ResourceName: "bucket-1",
		Severity:     "critical",
	})

	elements := store.BuildGraph()

	if len(elements.Nodes) < 2 {
		t.Errorf("Expected at least 2 nodes, got %d", len(elements.Nodes))
	}
}

// TestBuildGraphWithEventsToDrifts tests building graph with events causing drifts
func TestBuildGraphWithEventsToDrifts(t *testing.T) {
	store := NewStore()

	store.AddEvent(types.Event{
		ResourceID:   "res-1",
		ResourceType: "aws_instance",
		EventName:    "StopInstances",
	})

	store.AddDrift(types.DriftAlert{
		ResourceID:   "res-1",
		ResourceType: "aws_instance",
		Severity:     "high",
	})

	elements := store.BuildGraph()

	if len(elements.Edges) < 1 {
		t.Errorf("Expected at least 1 edge connecting event to drift, got %d", len(elements.Edges))
	}
}

// TestExtractRelationshipsEC2 tests relationship extraction for EC2
func TestExtractRelationshipsEC2(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Name: "web-server",
		Attributes: map[string]interface{}{
			"id":                     "i-123",
			"subnet_id":              "subnet-456",
			"vpc_security_group_ids": []interface{}{"sg-789"},
		},
	}

	rels := extractRelationships(resource)

	// Should have relationships for subnet and security group
	if len(rels) < 2 {
		t.Errorf("Expected at least 2 relationships for EC2, got %d", len(rels))
	}

	// Check for DEPENDS_ON relationship to subnet
	foundSubnetDep := false
	for _, rel := range rels {
		if rel.Type == DEPENDS_ON && rel.EndNode == "subnet-456" {
			foundSubnetDep = true
		}
	}

	if !foundSubnetDep {
		t.Error("Expected DEPENDS_ON relationship to subnet")
	}
}

// TestExtractRelationshipsSubnet tests relationship extraction for Subnet
func TestExtractRelationshipsSubnet(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_subnet",
		Name: "private-subnet",
		Attributes: map[string]interface{}{
			"id":     "subnet-123",
			"vpc_id": "vpc-456",
		},
	}

	rels := extractRelationships(resource)

	// Should have PART_OF relationship to VPC
	if len(rels) == 0 {
		t.Error("Expected at least 1 relationship for subnet")
	}

	foundVPCRel := false
	for _, rel := range rels {
		if rel.Type == PART_OF && rel.EndNode == "vpc-456" {
			foundVPCRel = true
		}
	}

	if !foundVPCRel {
		t.Error("Expected PART_OF relationship to VPC")
	}
}

// TestExtractRelationshipsRDS tests relationship extraction for RDS
func TestExtractRelationshipsRDS(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_db_instance",
		Name: "prod-db",
		Attributes: map[string]interface{}{
			"id":                   "prod-db-instance",
			"db_subnet_group_name": "subnet-group-1",
			"vpc_security_group_ids": []interface{}{"sg-123", "sg-456"},
		},
	}

	rels := extractRelationships(resource)

	if len(rels) < 3 {
		t.Errorf("Expected at least 3 relationships for RDS, got %d", len(rels))
	}

	// Check for subnet group dependency
	foundSubnetGroup := false
	for _, rel := range rels {
		if rel.Type == DEPENDS_ON && rel.EndNode == "subnet-group-1" {
			foundSubnetGroup = true
		}
	}

	if !foundSubnetGroup {
		t.Error("Expected DEPENDS_ON relationship to subnet group")
	}
}

// TestExtractRelationshipsEKS tests relationship extraction for EKS
func TestExtractRelationshipsEKS(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_eks_cluster",
		Name: "prod-cluster",
		Attributes: map[string]interface{}{
			"id": "prod-cluster",
			"vpc_config": []interface{}{
				map[string]interface{}{
					"subnet_ids":         []interface{}{"subnet-1", "subnet-2"},
					"security_group_ids": []interface{}{"sg-123"},
				},
			},
		},
	}

	rels := extractRelationships(resource)

	if len(rels) < 3 {
		t.Errorf("Expected at least 3 relationships for EKS, got %d", len(rels))
	}

	// Check for subnet dependencies
	subnetDeps := 0
	for _, rel := range rels {
		if rel.Type == DEPENDS_ON && (rel.EndNode == "subnet-1" || rel.EndNode == "subnet-2") {
			subnetDeps++
		}
	}

	if subnetDeps < 2 {
		t.Errorf("Expected at least 2 subnet dependencies for EKS, got %d", subnetDeps)
	}
}

// TestAddResourceSpecificPropertiesVPC tests adding VPC-specific properties
func TestAddResourceSpecificPropertiesVPC(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			"id":         "vpc-123",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	if properties["cidr"] != "10.0.0.0/16" {
		t.Errorf("Expected CIDR '10.0.0.0/16', got '%v'", properties["cidr"])
	}
}

// TestAddResourceSpecificPropertiesSubnet tests adding Subnet-specific properties
func TestAddResourceSpecificPropertiesSubnet(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_subnet",
		Attributes: map[string]interface{}{
			"cidr_block":               "10.0.1.0/24",
			"availability_zone":        "us-east-1a",
			"map_public_ip_on_launch":  true,
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	if properties["cidr"] != "10.0.1.0/24" {
		t.Errorf("Expected CIDR '10.0.1.0/24', got '%v'", properties["cidr"])
	}

	if properties["availability_zone"] != "us-east-1a" {
		t.Errorf("Expected AZ 'us-east-1a', got '%v'", properties["availability_zone"])
	}

	if properties["public"] != true {
		t.Errorf("Expected public 'true', got '%v'", properties["public"])
	}
}

// TestAddResourceSpecificPropertiesEC2 tests adding EC2-specific properties
func TestAddResourceSpecificPropertiesEC2(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"instance_type": "t3.medium",
			"instance_state": "running",
			"private_ip":    "10.0.1.100",
			"public_ip":     "203.0.113.1",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	if properties["instance_type"] != "t3.medium" {
		t.Errorf("Expected instance_type 't3.medium', got '%v'", properties["instance_type"])
	}

	if properties["state"] != "running" {
		t.Errorf("Expected state 'running', got '%v'", properties["state"])
	}

	if properties["private_ip"] != "10.0.1.100" {
		t.Errorf("Expected private_ip '10.0.1.100', got '%v'", properties["private_ip"])
	}
}

// TestHierarchyAssignResourceToSubnet tests assigning resources to subnets
func TestHierarchyAssignResourceToSubnet(t *testing.T) {
	builder := NewHierarchyBuilder()

	resources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-12345678",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_subnet",
			Name: "public-subnet",
			Attributes: map[string]interface{}{
				"id":         "subnet-87654321",
				"vpc_id":     "vpc-12345678",
				"cidr_block": "10.0.1.0/24",
				"availability_zone": "us-east-1a",
			},
		},
		{
			Type: "aws_instance",
			Name: "web-server",
			Attributes: map[string]interface{}{
				"id":        "i-instance123",
				"subnet_id": "subnet-87654321",
			},
		},
	}

	hierarchy := builder.BuildHierarchy(resources)

	// Find the instance in the hierarchy
	found := false
	for _, region := range hierarchy.Regions {
		for _, vpc := range region.VPCs {
			for _, az := range vpc.AvailabilityZones {
				for _, subnet := range az.Subnets {
					for _, resID := range subnet.Resources {
						if resID == "i-instance123" {
							found = true
							break
						}
					}
				}
			}
		}
	}

	if !found {
		t.Error("Expected EC2 instance to be assigned to subnet in hierarchy")
	}
}

// TestHierarchyAssignResourceToVPC tests assigning resources to VPCs
func TestHierarchyAssignResourceToVPC(t *testing.T) {
	builder := NewHierarchyBuilder()

	resources := []*terraform.Resource{
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id":         "vpc-12345678",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type: "aws_security_group",
			Name: "web-sg",
			Attributes: map[string]interface{}{
				"id":     "sg-12345678",
				"vpc_id": "vpc-12345678",
			},
		},
	}

	hierarchy := builder.BuildHierarchy(resources)

	// Find the security group in the VPC
	found := false
	for _, region := range hierarchy.Regions {
		for _, vpc := range region.VPCs {
			for _, resID := range vpc.Resources {
				if resID == "sg-12345678" {
					found = true
					break
				}
			}
		}
	}

	if !found {
		t.Error("Expected security group to be assigned to VPC in hierarchy")
	}
}

// TestMatchPropertyFiltering tests Match pattern with property filtering
func TestMatchPropertyFiltering(t *testing.T) {
	db := NewGraphDatabase()

	node1 := &Node{
		ID:     "vpc-1",
		Labels: []string{"Resource", "VPC"},
		Properties: map[string]interface{}{
			"id":   "vpc-1",
			"cidr": "10.0.0.0/16",
		},
	}

	node2 := &Node{
		ID:     "subnet-1",
		Labels: []string{"Resource", "Subnet"},
		Properties: map[string]interface{}{
			"id":   "subnet-1",
			"cidr": "10.0.1.0/24",
		},
	}

	db.AddNode(node1)
	db.AddNode(node2)

	db.AddRelationship(&Relationship{
		ID:        "rel-1",
		Type:      CONTAINS,
		StartNode: "vpc-1",
		EndNode:   "subnet-1",
		Properties: map[string]interface{}{},
	})

	// Test pattern matching with property filter
	pattern := &MatchPattern{
		StartLabels: []string{"VPC"},
		RelType:     CONTAINS,
		EndLabels:   []string{"Subnet"},
		EndFilter: map[string]interface{}{
			"cidr": "10.0.1.0/24",
		},
	}

	results := db.Match(pattern)

	if len(results) != 1 {
		t.Errorf("Expected 1 match with property filter, got %d", len(results))
	}
}

// TestFindNodesWithLabelsAll tests finding all nodes when no labels specified
func TestFindNodesWithLabelsAll(t *testing.T) {
	db := NewGraphDatabase()

	nodes := []*Node{
		{ID: "node-1", Labels: []string{"Resource"}, Properties: map[string]interface{}{}},
		{ID: "node-2", Labels: []string{"EC2"}, Properties: map[string]interface{}{}},
		{ID: "node-3", Labels: []string{"VPC"}, Properties: map[string]interface{}{}},
	}

	for _, node := range nodes {
		db.AddNode(node)
	}

	// Should return all nodes when no labels specified
	foundNodes := db.findNodesWithLabels([]string{})

	if len(foundNodes) != 3 {
		t.Errorf("Expected 3 nodes when no labels specified, got %d", len(foundNodes))
	}
}

// TestHasAllLabelsMultiple tests checking multiple labels on a node
func TestHasAllLabelsMultiple(t *testing.T) {
	db := NewGraphDatabase()

	node := &Node{
		ID:     "node-1",
		Labels: []string{"Resource", "EC2", "Drifted"},
		Properties: map[string]interface{}{},
	}

	db.AddNode(node)

	// Should have all three labels
	if !db.hasAllLabels(node, []string{"Resource", "EC2", "Drifted"}) {
		t.Error("Expected node to have all three labels")
	}

	// Should not have a label it doesn't have
	if db.hasAllLabels(node, []string{"Resource", "NonExistent"}) {
		t.Error("Expected hasAllLabels to return false for non-existent label")
	}
}

// ============================================================================
// EXTENDED BUILDER, CONVERTER, AND HIERARCHY EDGE CASE TESTS
// ============================================================================

// TestBuildDependencyEdgesEC2 tests building dependency edges for EC2
func TestBuildDependencyEdgesEC2(t *testing.T) {
	resources := []*terraform.Resource{
		{
			Type: "aws_instance",
			Name: "web-server",
			Attributes: map[string]interface{}{
				"id":        "i-123",
				"subnet_id": "subnet-456",
				"vpc_security_group_ids": []interface{}{"sg-789"},
			},
		},
		{
			Type: "aws_subnet",
			Name: "public-subnet",
			Attributes: map[string]interface{}{
				"id": "subnet-456",
			},
		},
		{
			Type: "aws_security_group",
			Name: "web-sg",
			Attributes: map[string]interface{}{
				"id": "sg-789",
			},
		},
	}

	edges := buildDependencyEdges(resources)

	// Should have edges from EC2 to subnet and security group
	if len(edges) < 2 {
		t.Errorf("Expected at least 2 edges, got %d", len(edges))
	}

	// Check for subnet edge
	foundSubnetEdge := false
	for _, edge := range edges {
		if edge.Data.Source == "i-123" && edge.Data.Target == "subnet-456" {
			foundSubnetEdge = true
		}
	}

	if !foundSubnetEdge {
		t.Error("Expected edge from EC2 to subnet")
	}
}

// TestBuildDependencyEdgesNAT tests building dependency edges for NAT Gateway
func TestBuildDependencyEdgesNAT(t *testing.T) {
	resources := []*terraform.Resource{
		{
			Type: "aws_nat_gateway",
			Name: "nat-gw",
			Attributes: map[string]interface{}{
				"id":        "nat-123",
				"subnet_id": "subnet-456",
			},
		},
		{
			Type: "aws_subnet",
			Name: "public-subnet",
			Attributes: map[string]interface{}{
				"id": "subnet-456",
			},
		},
	}

	edges := buildDependencyEdges(resources)

	if len(edges) < 1 {
		t.Errorf("Expected at least 1 edge for NAT, got %d", len(edges))
	}

	foundNATEdge := false
	for _, edge := range edges {
		if edge.Data.Source == "nat-123" && edge.Data.Target == "subnet-456" {
			foundNATEdge = true
		}
	}

	if !foundNATEdge {
		t.Error("Expected edge from NAT to subnet")
	}
}

// TestBuildDependencyEdgesRouteTable tests building dependency edges for Route Table
func TestBuildDependencyEdgesRouteTable(t *testing.T) {
	resources := []*terraform.Resource{
		{
			Type: "aws_route_table",
			Name: "main-rt",
			Attributes: map[string]interface{}{
				"id":     "rt-123",
				"vpc_id": "vpc-456",
			},
		},
		{
			Type: "aws_vpc",
			Name: "main-vpc",
			Attributes: map[string]interface{}{
				"id": "vpc-456",
			},
		},
	}

	edges := buildDependencyEdges(resources)

	if len(edges) < 1 {
		t.Errorf("Expected at least 1 edge for route table, got %d", len(edges))
	}

	foundRTEdge := false
	for _, edge := range edges {
		if edge.Data.Source == "rt-123" && edge.Data.Target == "vpc-456" {
			foundRTEdge = true
		}
	}

	if !foundRTEdge {
		t.Error("Expected edge from route table to VPC")
	}
}

// TestExtractRelationshipsIGW tests relationship extraction for Internet Gateway
func TestExtractRelationshipsIGW(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_internet_gateway",
		Name: "main-igw",
		Attributes: map[string]interface{}{
			"id":     "igw-123",
			"vpc_id": "vpc-456",
		},
	}

	rels := extractRelationships(resource)

	if len(rels) == 0 {
		t.Error("Expected at least 1 relationship for IGW")
	}

	foundIGWRel := false
	for _, rel := range rels {
		if rel.Type == CONNECTS_TO && rel.EndNode == "vpc-456" {
			foundIGWRel = true
		}
	}

	if !foundIGWRel {
		t.Error("Expected CONNECTS_TO relationship from IGW to VPC")
	}
}

// TestExtractRelationshipsLoadBalancer tests relationship extraction for ALB
func TestExtractRelationshipsLoadBalancer(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_lb",
		Name: "web-alb",
		Attributes: map[string]interface{}{
			"id":              "alb-123",
			"subnets":         []interface{}{"subnet-1", "subnet-2"},
			"security_groups": []interface{}{"sg-123"},
		},
	}

	rels := extractRelationships(resource)

	if len(rels) < 3 {
		t.Errorf("Expected at least 3 relationships for ALB, got %d", len(rels))
	}

	// Check for subnet dependencies
	subnetDeps := 0
	for _, rel := range rels {
		if rel.Type == DEPENDS_ON && (rel.EndNode == "subnet-1" || rel.EndNode == "subnet-2") {
			subnetDeps++
		}
	}

	if subnetDeps < 2 {
		t.Errorf("Expected at least 2 subnet dependencies, got %d", subnetDeps)
	}
}

// TestExtractRelationshipsECSService tests relationship extraction for ECS Service
func TestExtractRelationshipsECSService(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_ecs_service",
		Name: "web-service",
		Attributes: map[string]interface{}{
			"id":      "web-service",
			"cluster": "web-cluster-arn",
			"load_balancer": []interface{}{
				map[string]interface{}{
					"target_group_arn": "tg-arn-123",
				},
			},
		},
	}

	rels := extractRelationships(resource)

	if len(rels) < 2 {
		t.Errorf("Expected at least 2 relationships for ECS service, got %d", len(rels))
	}

	// Check for cluster relationship
	foundClusterRel := false
	for _, rel := range rels {
		if rel.Type == RUNS_IN && rel.EndNode == "web-cluster-arn" {
			foundClusterRel = true
		}
	}

	if !foundClusterRel {
		t.Error("Expected RUNS_IN relationship to cluster")
	}
}

// TestExtractRelationshipsIAMPolicy tests relationship extraction for IAM Policy
func TestExtractRelationshipsIAMPolicy(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_iam_policy",
		Name: "admin-policy",
		Attributes: map[string]interface{}{
			"id":   "admin-policy-arn",
			"role": "admin-role-arn",
		},
	}

	rels := extractRelationships(resource)

	if len(rels) == 0 {
		t.Error("Expected at least 1 relationship for IAM policy")
	}

	foundPolicyRel := false
	for _, rel := range rels {
		if rel.Type == APPLIES_TO && rel.EndNode == "admin-role-arn" {
			foundPolicyRel = true
		}
	}

	if !foundPolicyRel {
		t.Error("Expected APPLIES_TO relationship to role")
	}
}

// TestHierarchyExtractRegionFromProvider tests extracting region from provider
func TestHierarchyExtractRegionFromProvider(t *testing.T) {
	builder := NewHierarchyBuilder()

	resource := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"provider": "provider.aws.us-west-2",
		},
	}

	region := builder.extractRegion(resource)

	if region == "" {
		t.Error("Expected region to be extracted from provider")
	}
}

// TestHierarchyExtractRegionFromARN tests extracting region from ARN
func TestHierarchyExtractRegionFromARN(t *testing.T) {
	builder := NewHierarchyBuilder()

	resource := &terraform.Resource{
		Type: "aws_instance",
		Attributes: map[string]interface{}{
			"arn": "arn:aws:ec2:eu-west-1:123456789012:instance/i-123",
		},
	}

	region := builder.extractRegion(resource)

	if region != "eu-west-1" {
		t.Errorf("Expected region 'eu-west-1', got '%s'", region)
	}
}

// TestHierarchyExtractRegionAttribute tests extracting region from region attribute
func TestHierarchyExtractRegionAttribute(t *testing.T) {
	builder := NewHierarchyBuilder()

	resource := &terraform.Resource{
		Type: "aws_s3_bucket",
		Attributes: map[string]interface{}{
			"region": "ap-southeast-1",
		},
	}

	region := builder.extractRegion(resource)

	if region != "ap-southeast-1" {
		t.Errorf("Expected region 'ap-southeast-1', got '%s'", region)
	}
}

// TestConvertTerraformResourceToCytoscapeWithTagsName tests resource name from tags
func TestConvertTerraformResourceToCytoscapeWithTagsName(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Name: "tf-instance",
		Attributes: map[string]interface{}{
			"id": "i-123",
			"tags": map[string]interface{}{
				"Name": "production-web-server",
			},
		},
	}

	node := ConvertTerraformResourceToCytoscape(resource, false)

	if node.Data.Label != "production-web-server" {
		t.Errorf("Expected label from tags 'production-web-server', got '%s'", node.Data.Label)
	}
}

// TestResourceToNodeWithDrift tests resourceToNode with drift label
func TestResourceToNodeWithDrift(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_instance",
		Name: "web-server",
		Attributes: map[string]interface{}{
			"id": "i-123",
		},
	}

	driftedIDs := map[string]bool{"i-123": true}
	node := resourceToNode(resource, driftedIDs)

	if node == nil {
		t.Fatal("Expected non-nil node")
	}

	// Should have Drifted label
	hasDriftLabel := false
	for _, label := range node.Labels {
		if label == "Drifted" {
			hasDriftLabel = true
		}
	}

	if !hasDriftLabel {
		t.Error("Expected node to have 'Drifted' label")
	}

	if node.Properties["has_drift"] != true {
		t.Error("Expected has_drift property to be true")
	}
}

// TestAddResourceSpecificPropertiesRDS tests adding RDS-specific properties
func TestAddResourceSpecificPropertiesRDS(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_db_instance",
		Attributes: map[string]interface{}{
			"engine":         "postgres",
			"instance_class": "db.t3.micro",
			"status":         "available",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	if properties["engine"] != "postgres" {
		t.Errorf("Expected engine 'postgres', got '%v'", properties["engine"])
	}

	if properties["instance_class"] != "db.t3.micro" {
		t.Errorf("Expected instance_class 'db.t3.micro', got '%v'", properties["instance_class"])
	}

	if properties["status"] != "available" {
		t.Errorf("Expected status 'available', got '%v'", properties["status"])
	}
}

// TestAddResourceSpecificPropertiesSG tests adding SecurityGroup-specific properties
func TestAddResourceSpecificPropertiesSG(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_security_group",
		Attributes: map[string]interface{}{
			"name":        "web-sg",
			"description": "Security group for web servers",
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	if properties["sg_name"] != "web-sg" {
		t.Errorf("Expected sg_name 'web-sg', got '%v'", properties["sg_name"])
	}

	if properties["description"] != "Security group for web servers" {
		t.Errorf("Expected description 'Security group for web servers', got '%v'", properties["description"])
	}
}

// TestAddResourceSpecificPropertiesWithTags tests adding properties with tags
func TestAddResourceSpecificPropertiesWithTags(t *testing.T) {
	resource := &terraform.Resource{
		Type: "aws_vpc",
		Attributes: map[string]interface{}{
			"cidr_block": "10.0.0.0/16",
			"tags": map[string]interface{}{
				"Environment": "production",
				"Team":        "platform",
			},
		},
	}

	properties := make(map[string]interface{})
	addResourceSpecificProperties(resource, properties)

	if properties["cidr"] != "10.0.0.0/16" {
		t.Errorf("Expected CIDR '10.0.0.0/16', got '%v'", properties["cidr"])
	}

	tags := properties["tags"].(map[string]interface{})
	if tags["Environment"] != "production" {
		t.Errorf("Expected tag Environment 'production', got '%v'", tags["Environment"])
	}
}

// TestExtractRelationshipsGenericVPC tests generic VPC dependency extraction
func TestExtractRelationshipsGenericVPC(t *testing.T) {
	// Test a resource type not explicitly handled but with vpc_id
	resource := &terraform.Resource{
		Type: "aws_custom_resource",
		Name: "custom",
		Attributes: map[string]interface{}{
			"id":     "custom-123",
			"vpc_id": "vpc-456",
		},
	}

	rels := extractRelationships(resource)

	// Should have generic VPC dependency
	foundVPCDep := false
	for _, rel := range rels {
		if rel.Type == DEPENDS_ON && rel.EndNode == "vpc-456" {
			foundVPCDep = true
		}
	}

	if !foundVPCDep {
		t.Error("Expected generic VPC dependency relationship")
	}
}

// TestBuildGraphWithAllResourceTypes tests BuildGraph with multiple resource types
func TestBuildGraphWithAllResourceTypes(t *testing.T) {
	store := NewStore()

	// Add various types of data
	store.AddDrift(types.DriftAlert{
		ResourceID:   "vpc-1",
		ResourceType: "aws_vpc",
		Severity:     "high",
	})

	store.AddEvent(types.Event{
		ResourceID:   "subnet-1",
		ResourceType: "aws_subnet",
		EventName:    "CreateSubnet",
	})

	store.AddUnmanaged(types.UnmanagedResourceAlert{
		ResourceID:   "sg-1",
		ResourceType: "aws_security_group",
		Severity:     "medium",
	})

	elements := store.BuildGraph()

	// Should have nodes for all resource types
	if len(elements.Nodes) < 3 {
		t.Errorf("Expected at least 3 nodes, got %d", len(elements.Nodes))
	}
}
