package graph

import (
	"fmt"
	"sync"
	"testing"
)

func newTestDB() *GraphDatabase {
	return NewGraphDatabase()
}

func addTestNode(db *GraphDatabase, id string, labels []string) *Node {
	node := &Node{
		ID:         id,
		Labels:     labels,
		Properties: map[string]interface{}{"name": id},
	}
	db.AddNode(node)
	return node
}

func addTestRel(t *testing.T, db *GraphDatabase, id, relType, from, to string) *Relationship {
	t.Helper()
	rel := &Relationship{
		ID:         id,
		Type:       relType,
		StartNode:  from,
		EndNode:    to,
		Properties: map[string]interface{}{},
	}
	if err := db.AddRelationship(rel); err != nil {
		t.Fatalf("AddRelationship(%s) failed: %v", id, err)
	}
	return rel
}

// --- AddNode / GetNode ---

func TestAddAndGetNode(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "vpc-1", []string{"Resource", "VPC"})

	got := db.GetNode("vpc-1")
	if got == nil {
		t.Fatal("expected node, got nil")
	}
	if got.ID != "vpc-1" {
		t.Errorf("ID = %q, want %q", got.ID, "vpc-1")
	}
}

func TestGetNode_NotFound(t *testing.T) {
	db := newTestDB()
	if got := db.GetNode("nonexistent"); got != nil {
		t.Errorf("expected nil for missing node, got %v", got)
	}
}

// --- Label index ---

func TestGetNodesByLabel(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "vpc-1", []string{"Resource", "VPC"})
	addTestNode(db, "vpc-2", []string{"Resource", "VPC"})
	addTestNode(db, "ec2-1", []string{"Resource", "EC2"})

	vpcs := db.GetNodesByLabel("VPC")
	if len(vpcs) != 2 {
		t.Errorf("VPC count = %d, want 2", len(vpcs))
	}

	ec2s := db.GetNodesByLabel("EC2")
	if len(ec2s) != 1 {
		t.Errorf("EC2 count = %d, want 1", len(ec2s))
	}
}

func TestHasLabel(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "vpc-1", []string{"Resource", "VPC"})

	if !db.HasLabel("vpc-1", "VPC") {
		t.Error("expected HasLabel(vpc-1, VPC) = true")
	}
	if db.HasLabel("vpc-1", "EC2") {
		t.Error("expected HasLabel(vpc-1, EC2) = false")
	}
	if db.HasLabel("nonexistent", "VPC") {
		t.Error("expected HasLabel(nonexistent, VPC) = false")
	}
}

// --- Relationships ---

func TestAddAndGetRelationship(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "ec2-1", []string{"EC2"})
	addTestNode(db, "subnet-1", []string{"Subnet"})
	addTestRel(t, db, "rel-1", DEPENDS_ON, "ec2-1", "subnet-1")

	got := db.GetRelationship("rel-1")
	if got == nil {
		t.Fatal("expected relationship, got nil")
	}
	if got.Type != DEPENDS_ON {
		t.Errorf("Type = %q, want %q", got.Type, DEPENDS_ON)
	}
}

func TestAddRelationship_MissingNode(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "ec2-1", []string{"EC2"})

	rel := &Relationship{ID: "rel-1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "missing"}
	if err := db.AddRelationship(rel); err != ErrNodeNotFound {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestGetOutgoingAndIncomingRelationships(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "a", []string{"A"})
	addTestNode(db, "b", []string{"B"})
	addTestNode(db, "c", []string{"C"})
	addTestRel(t, db, "r1", DEPENDS_ON, "a", "b")
	addTestRel(t, db, "r2", DEPENDS_ON, "a", "c")

	out := db.GetOutgoingRelationships("a")
	if len(out) != 2 {
		t.Errorf("outgoing count = %d, want 2", len(out))
	}

	in := db.GetIncomingRelationships("b")
	if len(in) != 1 {
		t.Errorf("incoming count = %d, want 1", len(in))
	}
}

func TestGetRelationshipsByType(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "a", []string{"A"})
	addTestNode(db, "b", []string{"B"})
	addTestNode(db, "c", []string{"C"})
	addTestRel(t, db, "r1", DEPENDS_ON, "a", "b")
	addTestRel(t, db, "r2", CONTAINS, "a", "c")

	deps := db.GetRelationshipsByType(DEPENDS_ON)
	if len(deps) != 1 {
		t.Errorf("DEPENDS_ON count = %d, want 1", len(deps))
	}
}

// --- Neighbors ---

func TestGetNeighbors(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "a", []string{"A"})
	addTestNode(db, "b", []string{"B"})
	addTestNode(db, "c", []string{"C"})
	addTestNode(db, "d", []string{"D"})
	addTestRel(t, db, "r1", DEPENDS_ON, "a", "b")
	addTestRel(t, db, "r2", DEPENDS_ON, "c", "a")

	neighbors := db.GetNeighbors("a")
	if len(neighbors) != 2 {
		t.Errorf("neighbor count = %d, want 2", len(neighbors))
	}
}

// --- Counts ---

func TestNodeAndRelationshipCount(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "a", []string{"A"})
	addTestNode(db, "b", []string{"B"})
	addTestRel(t, db, "r1", DEPENDS_ON, "a", "b")

	if db.NodeCount() != 2 {
		t.Errorf("NodeCount = %d, want 2", db.NodeCount())
	}
	if db.RelationshipCount() != 1 {
		t.Errorf("RelationshipCount = %d, want 1", db.RelationshipCount())
	}
}

// --- GetAll ---

func TestGetAllNodesAndRelationships(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "a", []string{"A"})
	addTestNode(db, "b", []string{"B"})
	addTestRel(t, db, "r1", DEPENDS_ON, "a", "b")

	nodes := db.GetAllNodes()
	if len(nodes) != 2 {
		t.Errorf("GetAllNodes len = %d, want 2", len(nodes))
	}

	rels := db.GetAllRelationships()
	if len(rels) != 1 {
		t.Errorf("GetAllRelationships len = %d, want 1", len(rels))
	}
}

// --- DeleteNode ---

func TestDeleteNode(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "a", []string{"Resource", "VPC"})
	addTestNode(db, "b", []string{"Resource", "Subnet"})
	addTestNode(db, "c", []string{"Resource", "EC2"})
	addTestRel(t, db, "r1", CONTAINS, "a", "b")
	addTestRel(t, db, "r2", DEPENDS_ON, "c", "b")

	db.DeleteNode("b")

	if db.GetNode("b") != nil {
		t.Error("node b should be deleted")
	}
	if db.NodeCount() != 2 {
		t.Errorf("NodeCount = %d, want 2", db.NodeCount())
	}
	if db.RelationshipCount() != 0 {
		t.Errorf("RelationshipCount = %d, want 0 after deleting b", db.RelationshipCount())
	}
	// Label index should be cleaned
	subnets := db.GetNodesByLabel("Subnet")
	if len(subnets) != 0 {
		t.Errorf("Subnet label count = %d, want 0", len(subnets))
	}
}

func TestDeleteNode_Nonexistent(t *testing.T) {
	db := newTestDB()
	db.DeleteNode("nonexistent") // should not panic
}

// --- Clear ---

func TestClear(t *testing.T) {
	db := newTestDB()
	addTestNode(db, "a", []string{"A"})
	addTestNode(db, "b", []string{"B"})
	addTestRel(t, db, "r1", DEPENDS_ON, "a", "b")

	db.Clear()

	if db.NodeCount() != 0 {
		t.Errorf("NodeCount after Clear = %d, want 0", db.NodeCount())
	}
	if db.RelationshipCount() != 0 {
		t.Errorf("RelationshipCount after Clear = %d, want 0", db.RelationshipCount())
	}
}

// --- Concurrency ---

func TestConcurrentAccess(t *testing.T) {
	db := newTestDB()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("node-%d", i)
			addTestNode(db, id, []string{"Resource"})
		}(i)
	}
	wg.Wait()

	if db.NodeCount() != 100 {
		t.Errorf("NodeCount = %d, want 100", db.NodeCount())
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("node-%d", i)
			_ = db.GetNode(id)
			_ = db.GetNodesByLabel("Resource")
			_ = db.NodeCount()
		}(i)
	}
	wg.Wait()
}
