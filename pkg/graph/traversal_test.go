package graph

import (
	"testing"
)

// buildLinearGraph: a -> b -> c -> d
func buildLinearGraph(t *testing.T) *GraphDatabase {
	t.Helper()
	db := NewGraphDatabase()
	for _, id := range []string{"a", "b", "c", "d"} {
		db.AddNode(&Node{ID: id, Labels: []string{"Resource"}, Properties: map[string]interface{}{}})
	}
	for _, r := range []struct{ id, from, to string }{
		{"r1", "a", "b"}, {"r2", "b", "c"}, {"r3", "c", "d"},
	} {
		rel := &Relationship{ID: r.id, Type: DEPENDS_ON, StartNode: r.from, EndNode: r.to, Properties: map[string]interface{}{}}
		if err := db.AddRelationship(rel); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

// buildDiamondGraph: a -> b, a -> c, b -> d, c -> d
func buildDiamondGraph(t *testing.T) *GraphDatabase {
	t.Helper()
	db := NewGraphDatabase()
	for _, id := range []string{"a", "b", "c", "d"} {
		db.AddNode(&Node{ID: id, Labels: []string{"Resource"}, Properties: map[string]interface{}{}})
	}
	rels := []struct{ id, from, to string }{
		{"r1", "a", "b"}, {"r2", "a", "c"}, {"r3", "b", "d"}, {"r4", "c", "d"},
	}
	for _, r := range rels {
		rel := &Relationship{ID: r.id, Type: DEPENDS_ON, StartNode: r.from, EndNode: r.to, Properties: map[string]interface{}{}}
		if err := db.AddRelationship(rel); err != nil {
			t.Fatal(err)
		}
	}
	return db
}

// --- FindPath ---

func TestFindPath_Linear(t *testing.T) {
	db := buildLinearGraph(t)
	path, err := db.FindPath("a", "d")
	if err != nil {
		t.Fatalf("FindPath: %v", err)
	}
	if path.Length != 3 {
		t.Errorf("path length = %d, want 3", path.Length)
	}
	if len(path.Nodes) != 4 {
		t.Errorf("path nodes = %d, want 4", len(path.Nodes))
	}
}

func TestFindPath_Diamond_ShortestPath(t *testing.T) {
	db := buildDiamondGraph(t)
	path, err := db.FindPath("a", "d")
	if err != nil {
		t.Fatalf("FindPath: %v", err)
	}
	// BFS finds shortest: a -> b -> d or a -> c -> d (length 2)
	if path.Length != 2 {
		t.Errorf("path length = %d, want 2", path.Length)
	}
}

func TestFindPath_SameNode(t *testing.T) {
	db := buildLinearGraph(t)
	path, err := db.FindPath("a", "a")
	if err != nil {
		t.Fatalf("FindPath: %v", err)
	}
	if path.Length != 0 {
		t.Errorf("path length = %d, want 0", path.Length)
	}
}

func TestFindPath_NodeNotFound(t *testing.T) {
	db := buildLinearGraph(t)
	_, err := db.FindPath("a", "nonexistent")
	if err != ErrNodeNotFound {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestFindPath_NoPath(t *testing.T) {
	db := NewGraphDatabase()
	db.AddNode(&Node{ID: "x", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "y", Labels: []string{"B"}, Properties: map[string]interface{}{}})
	// No relationships
	_, err := db.FindPath("x", "y")
	if err != ErrInvalidPath {
		t.Errorf("expected ErrInvalidPath, got %v", err)
	}
}

// --- FindImpactRadius ---

func TestFindImpactRadius(t *testing.T) {
	db := buildLinearGraph(t)
	result := db.FindImpactRadius("a", 2)

	// a (0), b (1), c (2) — d should not be included (depth 3)
	if len(result.Nodes) != 3 {
		t.Errorf("impact radius nodes = %d, want 3", len(result.Nodes))
	}
	if result.Distances["a"] != 0 {
		t.Error("distance to self should be 0")
	}
	if result.Distances["b"] != 1 {
		t.Error("distance to b should be 1")
	}
}

func TestFindImpactRadius_NodeNotFound(t *testing.T) {
	db := buildLinearGraph(t)
	result := db.FindImpactRadius("nonexistent", 5)
	if len(result.Nodes) != 0 {
		t.Errorf("expected 0 nodes for missing node, got %d", len(result.Nodes))
	}
}

// --- FindDependencies ---

func TestFindDependencies(t *testing.T) {
	db := buildLinearGraph(t)
	deps := db.FindDependencies("a", 10)
	// a -> b -> c -> d, so deps should include b, c, d
	if len(deps) != 3 {
		t.Errorf("dependencies = %d, want 3", len(deps))
	}
}

func TestFindDependencies_MaxDepth(t *testing.T) {
	db := buildLinearGraph(t)
	deps := db.FindDependencies("a", 1)
	// Only b at depth 1
	if len(deps) != 1 {
		t.Errorf("dependencies with maxDepth=1 = %d, want 1", len(deps))
	}
}

func TestFindDependencies_NotFound(t *testing.T) {
	db := buildLinearGraph(t)
	deps := db.FindDependencies("nonexistent", 5)
	if len(deps) != 0 {
		t.Errorf("expected 0 deps for missing node, got %d", len(deps))
	}
}

// --- FindDependents ---

func TestFindDependents(t *testing.T) {
	db := buildLinearGraph(t)
	// d has no outgoing, c -> d, b -> c, a -> b
	// Dependents of d: c (incoming), b (c's incoming), a (b's incoming)
	dependents := db.FindDependents("d", 10)
	if len(dependents) != 3 {
		t.Errorf("dependents = %d, want 3", len(dependents))
	}
}

func TestFindDependents_MaxDepth(t *testing.T) {
	db := buildLinearGraph(t)
	dependents := db.FindDependents("d", 1)
	if len(dependents) != 1 {
		t.Errorf("dependents with maxDepth=1 = %d, want 1", len(dependents))
	}
}

// --- FindCriticalPaths ---

func TestFindCriticalPaths(t *testing.T) {
	db := buildDiamondGraph(t)
	// d has 2 incoming (from b and c)
	critical := db.FindCriticalPaths(2)
	if len(critical) != 1 {
		t.Errorf("critical nodes with minDependents=2 = %d, want 1", len(critical))
	}
	if len(critical) > 0 && critical[0].ID != "d" {
		t.Errorf("critical node = %q, want d", critical[0].ID)
	}
}

// --- Match ---

func TestMatch(t *testing.T) {
	db := NewGraphDatabase()
	db.AddNode(&Node{ID: "ec2-1", Labels: []string{"Resource", "EC2"}, Properties: map[string]interface{}{"name": "web"}})
	db.AddNode(&Node{ID: "subnet-1", Labels: []string{"Resource", "Subnet"}, Properties: map[string]interface{}{"name": "sub-a"}})
	db.AddNode(&Node{ID: "subnet-2", Labels: []string{"Resource", "Subnet"}, Properties: map[string]interface{}{"name": "sub-b"}})
	rel := &Relationship{ID: "r1", Type: DEPENDS_ON, StartNode: "ec2-1", EndNode: "subnet-1", Properties: map[string]interface{}{}}
	db.AddRelationship(rel)

	// Match EC2 -> Subnet
	pattern := &MatchPattern{
		StartLabels: []string{"EC2"},
		RelType:     DEPENDS_ON,
		EndLabels:   []string{"Subnet"},
	}
	results := db.Match(pattern)
	if len(results) != 1 {
		t.Errorf("match results = %d, want 1", len(results))
	}

	// Match with property filter
	pattern2 := &MatchPattern{
		StartLabels: []string{"EC2"},
		RelType:     DEPENDS_ON,
		EndLabels:   []string{"Subnet"},
		EndFilter:   map[string]interface{}{"name": "sub-b"},
	}
	results2 := db.Match(pattern2)
	if len(results2) != 0 {
		t.Errorf("match with filter = %d, want 0", len(results2))
	}
}

func TestMatch_EmptyLabels(t *testing.T) {
	db := NewGraphDatabase()
	db.AddNode(&Node{ID: "a", Labels: []string{"A"}, Properties: map[string]interface{}{}})
	db.AddNode(&Node{ID: "b", Labels: []string{"B"}, Properties: map[string]interface{}{}})
	rel := &Relationship{ID: "r1", Type: DEPENDS_ON, StartNode: "a", EndNode: "b", Properties: map[string]interface{}{}}
	db.AddRelationship(rel)

	// Empty labels = match all
	results := db.Match(&MatchPattern{})
	// Should return all nodes with outgoing rels matching empty type filter
	if len(results) < 1 {
		t.Errorf("match with empty pattern = %d, want >= 1", len(results))
	}
}
