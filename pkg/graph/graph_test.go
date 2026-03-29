package graph

import (
	"testing"
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
