package graph

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDatabase tests the Database constructor
func TestNewDatabase(t *testing.T) {
	db := NewDatabase()

	assert.NotNil(t, db)
	assert.Equal(t, 0, db.NodeCount())
	assert.Equal(t, 0, db.RelationshipCount())
	assert.NotNil(t, db.nodes)
	assert.NotNil(t, db.relationships)
	assert.NotNil(t, db.nodesByLabel)
	assert.NotNil(t, db.outgoing)
	assert.NotNil(t, db.incoming)
	assert.NotNil(t, db.relationshipsByType)
}

// TestAddNode_DuplicateID tests adding a node with duplicate ID (overwrites)
func TestAddNode_DuplicateID(t *testing.T) {
	db := NewDatabase()

	node1 := &Node{
		ID:         "node-1",
		Labels:     []string{"Label1"},
		Properties: map[string]interface{}{"key": "value1"},
	}
	node2 := &Node{
		ID:         "node-1",
		Labels:     []string{"Label2"},
		Properties: map[string]interface{}{"key": "value2"},
	}

	db.AddNode(node1)
	assert.Equal(t, 1, db.NodeCount())

	db.AddNode(node2)
	assert.Equal(t, 1, db.NodeCount())

	retrieved := db.GetNode("node-1")
	assert.Equal(t, "Label2", retrieved.Labels[0])
	assert.Equal(t, "value2", retrieved.Properties["key"])
}

// TestRemoveNode_WithRelationships tests deleting a node with relationships
func TestRemoveNode_WithRelationships(t *testing.T) {
	db := NewDatabase()

	node1 := &Node{ID: "node-1", Labels: []string{"A"}, Properties: map[string]interface{}{}}
	node2 := &Node{ID: "node-2", Labels: []string{"B"}, Properties: map[string]interface{}{}}
	node3 := &Node{ID: "node-3", Labels: []string{"C"}, Properties: map[string]interface{}{}}

	db.AddNode(node1)
	db.AddNode(node2)
	db.AddNode(node3)

	// Create relationships: node1 -> node2 and node3 -> node1
	rel1 := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "node-1",
		EndNode:   "node-2",
		Properties: map[string]interface{}{},
	}

	rel2 := &Relationship{
		ID:        "rel-2",
		Type:      DEPENDS_ON,
		StartNode: "node-3",
		EndNode:   "node-1",
		Properties: map[string]interface{}{},
	}

	db.AddRelationship(rel1)
	db.AddRelationship(rel2)

	assert.Equal(t, 3, db.NodeCount())
	assert.Equal(t, 2, db.RelationshipCount())

	// Delete node-1 and verify relationships are cleaned up
	db.DeleteNode("node-1")
	assert.Equal(t, 2, db.NodeCount())
	assert.Equal(t, 0, db.RelationshipCount())

	// Verify node1 is gone
	retrieved := db.GetNode("node-1")
	assert.Nil(t, retrieved)
}

// TestAddRelationship_NodeNotFound tests error handling for missing nodes
func TestAddRelationship_NodeNotFound(t *testing.T) {
	db := NewDatabase()

	rel := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "nonexistent-1",
		EndNode:   "nonexistent-2",
		Properties: map[string]interface{}{},
	}

	err := db.AddRelationship(rel)
	assert.Error(t, err)
	assert.Equal(t, ErrNodeNotFound, err)
}

// TestConcurrentAccessStress tests concurrent read/write safety with stress
func TestConcurrentAccessStress(t *testing.T) {
	db := NewDatabase()

	// Add initial nodes
	for i := 0; i < 100; i++ {
		db.AddNode(&Node{
			ID:         string(rune(i + 1000)),
			Labels:     []string{"Resource"},
			Properties: map[string]interface{}{},
		})
	}

	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Concurrent reads
			for j := 0; j < 20; j++ {
				_ = db.GetNode(string(rune(j + 1000)))
				_ = db.GetAllNodes()
				_ = db.GetAllRelationships()
			}

			// Concurrent writes
			for k := 0; k < 10; k++ {
				node := &Node{
					ID:         string(rune(2000 + id*10 + k)),
					Labels:     []string{"Resource"},
					Properties: map[string]interface{}{"id": id},
				}
				db.AddNode(node)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, 100+numGoroutines*10, db.NodeCount())
}

// TestEdgeCases tests edge cases and nil handling
func TestEdgeCases(t *testing.T) {
	db := NewDatabase()

	// Test getting from empty database
	assert.Nil(t, db.GetNode("nonexistent"))
	assert.Len(t, db.GetAllNodes(), 0)
	assert.Len(t, db.GetAllRelationships(), 0)

	// Test with empty labels
	node := &Node{
		ID:         "node-1",
		Labels:     []string{},
		Properties: map[string]interface{}{},
	}
	db.AddNode(node)

	byLabel := db.GetNodesByLabel("SomeLabel")
	assert.Len(t, byLabel, 0)

	// Test with nil properties
	nodeNilProps := &Node{
		ID:         "node-2",
		Labels:     []string{"A"},
		Properties: nil,
	}
	db.AddNode(nodeNilProps)
	assert.Nil(t, db.GetNode("node-2").Properties)

	// Test deleting non-existent node (should not panic)
	db.DeleteNode("nonexistent")
	assert.Equal(t, 2, db.NodeCount())
}

// TestLargeGraphOperations tests operations on large graphs
func TestLargeGraphOperations(t *testing.T) {
	db := NewDatabase()

	numNodes := 500
	// Create a large graph
	for i := 0; i < numNodes; i++ {
		db.AddNode(&Node{
			ID:         string(rune(i)),
			Labels:     []string{"Resource", "Type"},
			Properties: map[string]interface{}{"index": i},
		})
	}

	// Add chain relationships
	for i := 0; i < numNodes-1; i++ {
		_ = db.AddRelationship(&Relationship{
			ID:        "rel-" + string(rune(i)),
			Type:      DEPENDS_ON,
			StartNode: string(rune(i)),
			EndNode:   string(rune(i + 1)),
			Properties: map[string]interface{}{},
		})
	}

	// Test retrieval is still fast
	for i := 0; i < 100; i++ {
		node := db.GetNode(string(rune(i % numNodes)))
		assert.NotNil(t, node)
	}

	// Test counts
	assert.Equal(t, numNodes, db.NodeCount())
	assert.Equal(t, numNodes-1, db.RelationshipCount())

	// Test GetAllNodes and GetAllRelationships work correctly
	allNodes := db.GetAllNodes()
	assert.Len(t, allNodes, numNodes)

	allRels := db.GetAllRelationships()
	assert.Len(t, allRels, numNodes-1)
}

// TestMultipleLabelScenario tests nodes with multiple labels
func TestMultipleLabelScenario(t *testing.T) {
	db := NewDatabase()

	node1 := &Node{
		ID:         "node-1",
		Labels:     []string{"Resource", "EC2", "Drifted"},
		Properties: map[string]interface{}{},
	}

	node2 := &Node{
		ID:         "node-2",
		Labels:     []string{"Resource", "EC2"},
		Properties: map[string]interface{}{},
	}

	node3 := &Node{
		ID:         "node-3",
		Labels:     []string{"Resource", "VPC"},
		Properties: map[string]interface{}{},
	}

	db.AddNode(node1)
	db.AddNode(node2)
	db.AddNode(node3)

	// Query by each label
	resources := db.GetNodesByLabel("Resource")
	assert.Len(t, resources, 3)

	ec2s := db.GetNodesByLabel("EC2")
	assert.Len(t, ec2s, 2)

	drifted := db.GetNodesByLabel("Drifted")
	assert.Len(t, drifted, 1)

	vpcs := db.GetNodesByLabel("VPC")
	assert.Len(t, vpcs, 1)

	// Test HasLabel
	assert.True(t, db.HasLabel("node-1", "Drifted"))
	assert.True(t, db.HasLabel("node-1", "EC2"))
	assert.False(t, db.HasLabel("node-3", "Drifted"))
}

// TestComplexGraphScenario tests a complex real-world scenario
func TestComplexGraphScenario(t *testing.T) {
	db := NewDatabase()

	// Simulate: VPC -> Subnets -> EC2 -> Security Groups
	vpc := &Node{ID: "vpc-1", Labels: []string{"VPC"}, Properties: map[string]interface{}{}}
	subnet1 := &Node{ID: "subnet-1", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	subnet2 := &Node{ID: "subnet-2", Labels: []string{"Subnet"}, Properties: map[string]interface{}{}}
	ec2 := &Node{ID: "ec2-1", Labels: []string{"EC2"}, Properties: map[string]interface{}{}}
	sg := &Node{ID: "sg-1", Labels: []string{"SecurityGroup"}, Properties: map[string]interface{}{}}

	for _, node := range []*Node{vpc, subnet1, subnet2, ec2, sg} {
		db.AddNode(node)
	}

	// Relationships
	rels := []*Relationship{
		{ID: "rel-1", Type: PART_OF, StartNode: "subnet-1", EndNode: "vpc-1", Properties: map[string]interface{}{}},
		{ID: "rel-2", Type: PART_OF, StartNode: "subnet-2", EndNode: "vpc-1", Properties: map[string]interface{}{}},
		{ID: "rel-3", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}},
		{ID: "rel-4", Type: SECURES, StartNode: "sg-1", EndNode: "ec2-1", Properties: map[string]interface{}{}},
	}

	for _, rel := range rels {
		_ = db.AddRelationship(rel)
	}

	assert.Equal(t, 5, db.NodeCount())
	assert.Equal(t, 4, db.RelationshipCount())

	// Test various queries
	vpcNodes := db.GetNodesByLabel("VPC")
	assert.Len(t, vpcNodes, 1)

	subnetNodes := db.GetNodesByLabel("Subnet")
	assert.Len(t, subnetNodes, 2)

	partOfRels := db.GetRelationshipsByType(PART_OF)
	assert.Len(t, partOfRels, 2)

	// Test path finding
	path, err := db.FindPath("ec2-1", "vpc-1")
	require.NoError(t, err)
	assert.NotNil(t, path)
	assert.True(t, len(path.Nodes) <= 3)

	// Test neighbors
	ec2Neighbors := db.GetNeighbors("ec2-1")
	assert.Greater(t, len(ec2Neighbors), 0)
}

// TestRelationshipTypeIndexing tests the relationship type index
func TestRelationshipTypeIndexing(t *testing.T) {
	db := NewDatabase()

	// Create nodes
	nodes := []*Node{
		{ID: "a", Labels: []string{"Node"}, Properties: map[string]interface{}{}},
		{ID: "b", Labels: []string{"Node"}, Properties: map[string]interface{}{}},
		{ID: "c", Labels: []string{"Node"}, Properties: map[string]interface{}{}},
		{ID: "d", Labels: []string{"Node"}, Properties: map[string]interface{}{}},
	}

	for _, n := range nodes {
		db.AddNode(n)
	}

	// Add different relationship types
	rels := []*Relationship{
		{ID: "r1", Type: DEPENDS_ON, StartNode: "a", EndNode: "b", Properties: map[string]interface{}{}},
		{ID: "r2", Type: DEPENDS_ON, StartNode: "b", EndNode: "c", Properties: map[string]interface{}{}},
		{ID: "r3", Type: CONTAINS, StartNode: "a", EndNode: "d", Properties: map[string]interface{}{}},
		{ID: "r4", Type: CONNECTS_TO, StartNode: "b", EndNode: "d", Properties: map[string]interface{}{}},
	}

	for _, r := range rels {
		_ = db.AddRelationship(r)
	}

	// Test type queries
	dependsOn := db.GetRelationshipsByType(DEPENDS_ON)
	assert.Len(t, dependsOn, 2)

	contains := db.GetRelationshipsByType(CONTAINS)
	assert.Len(t, contains, 1)

	connectsTo := db.GetRelationshipsByType(CONNECTS_TO)
	assert.Len(t, connectsTo, 1)

	notExists := db.GetRelationshipsByType("NONEXISTENT")
	assert.Len(t, notExists, 0)
}

// TestGetNodesByLabel_Empty tests getting nodes by label when none exist
func TestGetNodesByLabel_Empty(t *testing.T) {
	db := NewDatabase()

	node := &Node{
		ID:         "node-1",
		Labels:     []string{"Label1"},
		Properties: map[string]interface{}{},
	}

	db.AddNode(node)

	byLabel := db.GetNodesByLabel("Label2")
	assert.Len(t, byLabel, 0)
}

// TestRelationshipWithProperties tests relationships with complex properties
func TestRelationshipWithProperties(t *testing.T) {
	db := NewDatabase()

	node1 := &Node{ID: "n1", Labels: []string{"A"}, Properties: map[string]interface{}{}}
	node2 := &Node{ID: "n2", Labels: []string{"B"}, Properties: map[string]interface{}{}}

	db.AddNode(node1)
	db.AddNode(node2)

	rel := &Relationship{
		ID:        "rel-1",
		Type:      DEPENDS_ON,
		StartNode: "n1",
		EndNode:   "n2",
		Properties: map[string]interface{}{
			"weight":        10,
			"direction":     "forward",
			"metadata":      map[string]interface{}{"key": "value"},
			"tags":          []string{"important", "critical"},
		},
	}

	err := db.AddRelationship(rel)
	require.NoError(t, err)

	retrieved := db.GetRelationship("rel-1")
	assert.NotNil(t, retrieved)
	assert.Equal(t, 10, retrieved.Properties["weight"])
	assert.Equal(t, "forward", retrieved.Properties["direction"])
}

// TestNodePropertiesUpdate tests updating node properties through add
func TestNodePropertiesUpdate(t *testing.T) {
	db := NewDatabase()

	node1 := &Node{
		ID:         "node-1",
		Labels:     []string{"Resource"},
		Properties: map[string]interface{}{"version": 1, "name": "test"},
	}

	db.AddNode(node1)

	// Update the same node
	node2 := &Node{
		ID:         "node-1",
		Labels:     []string{"Resource"},
		Properties: map[string]interface{}{"version": 2, "name": "updated"},
	}

	db.AddNode(node2)

	retrieved := db.GetNode("node-1")
	assert.Equal(t, 2, retrieved.Properties["version"])
	assert.Equal(t, "updated", retrieved.Properties["name"])
}
