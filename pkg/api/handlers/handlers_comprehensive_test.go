package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/provider"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// ===== GraphQueryHandler Tests (12 methods at 0%) =====

func TestGraphQueryHandler_GetNode_Success(t *testing.T) {
	// Setup
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Add test node
	node := &graph.Node{
		ID:         "test-node-1",
		Labels:     []string{"Resource", "EC2"},
		Properties: map[string]interface{}{"name": "web-server", "type": "t2.micro"},
	}
	db.AddNode(node)

	handler := NewGraphQueryHandler(store)

	// Create request with URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "test-node-1")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/nodes/test-node-1", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	// Execute
	handler.GetNode(w, req)

	// Verify
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify node data is in response
	if resp.Data == nil {
		t.Error("expected data in response")
	}
}

func TestGraphQueryHandler_GetNode_NotFound(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/nodes/nonexistent", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetNode(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Success {
		t.Error("expected success to be false")
	}
}

func TestGraphQueryHandler_GetNodesByLabel_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Add test nodes with label
	for i := 0; i < 3; i++ {
		node := &graph.Node{
			ID:         fmt.Sprintf("ec2-%d", i),
			Labels:     []string{"Resource", "EC2"},
			Properties: map[string]interface{}{"name": fmt.Sprintf("server-%d", i)},
		}
		db.AddNode(node)
	}

	handler := NewGraphQueryHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/graph/nodes?label=EC2", nil)
	w := httptest.NewRecorder()

	handler.GetNodesByLabel(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify response structure
	data := resp.Data.(map[string]interface{})
	if count, ok := data["count"].(float64); !ok || count != 3 {
		t.Errorf("expected count 3, got %v", count)
	}
}

func TestGraphQueryHandler_GetNodesByLabel_MissingLabel(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/graph/nodes?label=", nil)
	w := httptest.NewRecorder()

	handler.GetNodesByLabel(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGraphQueryHandler_GetPath_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create a simple path: node1 -> node2 -> node3
	node1 := &graph.Node{ID: "node1", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	node2 := &graph.Node{ID: "node2", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	node3 := &graph.Node{ID: "node3", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}

	db.AddNode(node1)
	db.AddNode(node2)
	db.AddNode(node3)

	db.AddRelationship(&graph.Relationship{
		ID:        "rel1",
		Type:      "DEPENDS_ON",
		StartNode: "node1",
		EndNode:   "node2",
	})
	db.AddRelationship(&graph.Relationship{
		ID:        "rel2",
		Type:      "DEPENDS_ON",
		StartNode: "node2",
		EndNode:   "node3",
	})

	handler := NewGraphQueryHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/graph/path?from=node1&to=node3", nil)
	w := httptest.NewRecorder()

	handler.GetPath(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGraphQueryHandler_GetPath_MissingParams(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	tests := []string{
		"/api/v1/graph/path?from=node1", // missing 'to'
		"/api/v1/graph/path?to=node2",   // missing 'from'
		"/api/v1/graph/path",            // missing both
	}

	for _, url := range tests {
		req := httptest.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		handler.GetPath(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("for %s: expected status %d, got %d", url, http.StatusBadRequest, w.Code)
		}
	}
}

func TestGraphQueryHandler_GetPath_NoPath(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create two disconnected nodes
	node1 := &graph.Node{ID: "node1", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	node2 := &graph.Node{ID: "node2", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(node1)
	db.AddNode(node2)
	// Don't add any relationship between them

	handler := NewGraphQueryHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/graph/path?from=node1&to=node2", nil)
	w := httptest.NewRecorder()

	handler.GetPath(w, req)

	// Should return NotFound when no path exists
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Success {
		t.Error("expected success to be false for non-existent path")
	}
}

func TestGraphQueryHandler_GetImpactRadius_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create a node with neighbors
	node := &graph.Node{ID: "central", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(node)

	for i := 0; i < 3; i++ {
		neighbor := &graph.Node{
			ID:     fmt.Sprintf("neighbor-%d", i),
			Labels: []string{"Resource"},
		}
		db.AddNode(neighbor)
		db.AddRelationship(&graph.Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      "DEPENDS_ON",
			StartNode: "central",
			EndNode:   fmt.Sprintf("neighbor-%d", i),
		})
	}

	handler := NewGraphQueryHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/graph/impact/central?depth=2", nil)
	w := httptest.NewRecorder()

	handler.GetImpactRadius(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGraphQueryHandler_GetImpactRadius_DefaultDepth(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create a node
	node := &graph.Node{ID: "central", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(node)

	handler := NewGraphQueryHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/graph/impact/central", nil)
	w := httptest.NewRecorder()

	handler.GetImpactRadius(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGraphQueryHandler_GetImpactRadius_InvalidDepth(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	tests := []struct {
		url  string
		desc string
	}{
		{"/api/v1/graph/impact/node?depth=0", "depth too low"},
		{"/api/v1/graph/impact/node?depth=11", "depth too high"},
		{"/api/v1/graph/impact/node?depth=invalid", "invalid depth"},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", test.url, nil)
		w := httptest.NewRecorder()
		handler.GetImpactRadius(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: expected status %d, got %d", test.desc, http.StatusBadRequest, w.Code)
		}
	}
}

func TestGraphQueryHandler_GetDependencies_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create a dependency chain: A -> B -> C
	nodeA := &graph.Node{ID: "A", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	nodeB := &graph.Node{ID: "B", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	nodeC := &graph.Node{ID: "C", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}

	db.AddNode(nodeA)
	db.AddNode(nodeB)
	db.AddNode(nodeC)

	db.AddRelationship(&graph.Relationship{
		ID: "rel1", Type: "DEPENDS_ON", StartNode: "A", EndNode: "B",
	})
	db.AddRelationship(&graph.Relationship{
		ID: "rel2", Type: "DEPENDS_ON", StartNode: "B", EndNode: "C",
	})

	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "A")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/dependencies/A?depth=5", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDependencies(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGraphQueryHandler_GetDependencies_InvalidDepth(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "node1")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/dependencies/node1?depth=invalid", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDependencies(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGraphQueryHandler_GetDependents_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create: A -> B -> C (so A has dependents B and C)
	nodeA := &graph.Node{ID: "A", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	nodeB := &graph.Node{ID: "B", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	nodeC := &graph.Node{ID: "C", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}

	db.AddNode(nodeA)
	db.AddNode(nodeB)
	db.AddNode(nodeC)

	db.AddRelationship(&graph.Relationship{
		ID: "rel1", Type: "DEPENDS_ON", StartNode: "B", EndNode: "A",
	})
	db.AddRelationship(&graph.Relationship{
		ID: "rel2", Type: "DEPENDS_ON", StartNode: "C", EndNode: "A",
	})

	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "A")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/dependents/A?depth=5", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDependents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGraphQueryHandler_GetDependents_InvalidDepth(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "node1")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/dependents/node1?depth=15", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDependents(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGraphQueryHandler_GetCriticalNodes_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create a critical node with many dependents
	centralNode := &graph.Node{ID: "vpc-1", Labels: []string{"Resource", "VPC"}, Properties: map[string]interface{}{}}
	db.AddNode(centralNode)

	// Add 5 dependent nodes
	for i := 0; i < 5; i++ {
		subnet := &graph.Node{
			ID:     fmt.Sprintf("subnet-%d", i),
			Labels: []string{"Resource", "Subnet"},
		}
		db.AddNode(subnet)
		db.AddRelationship(&graph.Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      "DEPENDS_ON",
			StartNode: fmt.Sprintf("subnet-%d", i),
			EndNode:   "vpc-1",
		})
	}

	handler := NewGraphQueryHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/graph/critical?min=3", nil)
	w := httptest.NewRecorder()

	handler.GetCriticalNodes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}

	data := resp.Data.(map[string]interface{})
	if count, ok := data["count"].(float64); !ok || count < 1 {
		t.Errorf("expected at least 1 critical node, got %v", count)
	}
}

func TestGraphQueryHandler_GetCriticalNodes_InvalidMin(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/graph/critical?min=0", nil)
	w := httptest.NewRecorder()

	handler.GetCriticalNodes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGraphQueryHandler_GetNeighbors_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create a node with neighbors
	center := &graph.Node{ID: "center", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(center)

	neighbor1 := &graph.Node{ID: "neighbor1", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	neighbor2 := &graph.Node{ID: "neighbor2", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(neighbor1)
	db.AddNode(neighbor2)

	db.AddRelationship(&graph.Relationship{
		ID: "rel1", Type: "DEPENDS_ON", StartNode: "center", EndNode: "neighbor1",
	})
	db.AddRelationship(&graph.Relationship{
		ID: "rel2", Type: "DEPENDS_ON", StartNode: "neighbor2", EndNode: "center",
	})

	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "center")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/neighbors/center", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetNeighbors(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}

	data := resp.Data.(map[string]interface{})
	if count, ok := data["count"].(float64); !ok || count != 2 {
		t.Errorf("expected 2 neighbors, got %v", count)
	}
}

func TestGraphQueryHandler_GetRelationships_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	node1 := &graph.Node{ID: "node1", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	node2 := &graph.Node{ID: "node2", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(node1)
	db.AddNode(node2)

	db.AddRelationship(&graph.Relationship{
		ID: "rel1", Type: "DEPENDS_ON", StartNode: "node1", EndNode: "node2",
	})

	handler := NewGraphQueryHandler(store)

	tests := []struct {
		direction string
		url       string
	}{
		{"outgoing", "/api/v1/graph/relationships/node1?direction=outgoing"},
		{"incoming", "/api/v1/graph/relationships/node2?direction=incoming"},
		{"both", "/api/v1/graph/relationships/node1?direction=both"},
	}

	for _, test := range tests {
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "node1")
		ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
		req := httptest.NewRequest("GET", test.url, nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetRelationships(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("%s: expected status %d, got %d", test.direction, http.StatusOK, w.Code)
		}
	}
}

func TestGraphQueryHandler_GetRelationships_InvalidDirection(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "node1")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/graph/relationships/node1?direction=invalid", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetRelationships(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGraphQueryHandler_GetGraphStats_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Add nodes and relationships
	for i := 0; i < 5; i++ {
		node := &graph.Node{
			ID:     fmt.Sprintf("node-%d", i),
			Labels: []string{"Resource", "EC2"},
		}
		db.AddNode(node)
	}

	for i := 0; i < 4; i++ {
		db.AddRelationship(&graph.Relationship{
			ID:        fmt.Sprintf("rel-%d", i),
			Type:      "DEPENDS_ON",
			StartNode: fmt.Sprintf("node-%d", i),
			EndNode:   fmt.Sprintf("node-%d", i+1),
		})
	}

	handler := NewGraphQueryHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/graph/stats", nil)
	w := httptest.NewRecorder()

	handler.GetGraphStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	data := resp.Data.(map[string]interface{})
	if nodeCount, ok := data["node_count"].(float64); !ok || nodeCount != 5 {
		t.Errorf("expected 5 nodes, got %v", nodeCount)
	}
}

func TestGraphQueryHandler_MatchPattern_Success(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	ec2 := &graph.Node{
		ID:         "i-123",
		Labels:     []string{"Resource", "EC2"},
		Properties: map[string]interface{}{"id": "i-123"},
	}
	subnet := &graph.Node{
		ID:         "subnet-456",
		Labels:     []string{"Resource", "Subnet"},
		Properties: map[string]interface{}{"id": "subnet-456"},
	}

	db.AddNode(ec2)
	db.AddNode(subnet)
	db.AddRelationship(&graph.Relationship{
		ID: "rel1", Type: "DEPENDS_ON", StartNode: "i-123", EndNode: "subnet-456",
	})

	handler := NewGraphQueryHandler(store)

	pattern := `{
		"start_labels": ["EC2"],
		"rel_type": "DEPENDS_ON",
		"end_labels": ["Subnet"],
		"end_filter": {}
	}`

	req := httptest.NewRequest("POST", "/api/v1/graph/match", strings.NewReader(pattern))
	w := httptest.NewRecorder()

	handler.MatchPattern(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGraphQueryHandler_MatchPattern_InvalidJSON(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	req := httptest.NewRequest("POST", "/api/v1/graph/match", strings.NewReader("invalid json"))
	w := httptest.NewRecorder()

	handler.MatchPattern(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Note: StateHandler tests require real terraform.StateManager with backend config
// These are more integration-style tests. For unit testing we focus on graph handlers
// which have zero coverage (12 methods in GraphQueryHandler alone)

// ===== Additional Coverage Tests =====

func TestGraphQueryHandler_GetRelationships_DefaultDirection(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	node1 := &graph.Node{ID: "node1", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	node2 := &graph.Node{ID: "node2", Labels: []string{"Resource"}, Properties: map[string]interface{}{}}
	db.AddNode(node1)
	db.AddNode(node2)

	db.AddRelationship(&graph.Relationship{
		ID: "rel1", Type: "DEPENDS_ON", StartNode: "node1", EndNode: "node2",
	})

	handler := NewGraphQueryHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "node1")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	// Don't specify direction - should default to "both"
	req := httptest.NewRequest("GET", "/api/v1/graph/relationships/node1", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetRelationships(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	if direction, ok := data["direction"].(string); !ok || direction != "both" {
		t.Errorf("expected direction 'both', got %v", direction)
	}
}

func TestStatsHandler_GetStats_EmptyBreakdown(t *testing.T) {
	store := graph.NewStore()

	// Add data without specific severities to test empty breakdown
	store.AddDrift(types.DriftAlert{
		ResourceID:   "i-1",
		ResourceType: "aws_instance",
		Severity:     "critical",
		Timestamp:    time.Now().Format(time.RFC3339),
	})

	handler := NewStatsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify severity breakdown includes all severities
	stats := resp.Data.(map[string]interface{})
	if breakdown, ok := stats["severity_breakdown"].(map[string]interface{}); ok {
		expected := []string{"critical", "high", "medium", "low"}
		for _, sev := range expected {
			if _, exists := breakdown[sev]; !exists {
				t.Errorf("expected severity %s in breakdown", sev)
			}
		}
	} else {
		t.Error("expected severity_breakdown in stats")
	}
}

func TestProviderStatusHandler_MultipleRecordCalls(t *testing.T) {
	mockRegistry := createMockRegistry()
	handler := NewProviderStatusHandler(mockRegistry)

	// Record multiple events and errors
	for i := 0; i < 5; i++ {
		handler.RecordEvent("aws", i%2 == 0)
	}
	handler.RecordError("aws")

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Verify AWS provider shows proper stats
	data := resp.Data.(map[string]interface{})
	if providers, ok := data["providers"].([]interface{}); ok && len(providers) > 0 {
		// Find AWS provider
		for _, p := range providers {
			if provider, ok := p.(map[string]interface{}); ok {
				if name, ok := provider["name"].(string); ok && name == "aws" {
					// Check event counts
					if events, ok := provider["events_received"].(float64); ok {
						if events != 5 {
							t.Errorf("expected 5 events, got %v", events)
						}
					}
				}
			}
		}
	}
}

func TestGraphQueryHandler_MatchPattern_EmptyLabels(t *testing.T) {
	store := graph.NewStore()
	db := store.GetGraphDB()

	// Create nodes with various labels
	for i := 0; i < 3; i++ {
		node := &graph.Node{
			ID:         fmt.Sprintf("node-%d", i),
			Labels:     []string{"Resource", "EC2"},
			Properties: map[string]interface{}{},
		}
		db.AddNode(node)
	}

	handler := NewGraphQueryHandler(store)

	// Pattern with empty start labels should match all nodes
	pattern := `{"start_labels": [], "rel_type": "", "end_labels": [], "end_filter": {}}`

	req := httptest.NewRequest("POST", "/api/v1/graph/match", strings.NewReader(pattern))
	w := httptest.NewRecorder()

	handler.MatchPattern(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestDriftsHandler_GetDrifts_EdgeCaseOffset(t *testing.T) {
	store := graph.NewStore()

	// Add 5 drifts
	for i := 0; i < 5; i++ {
		store.AddDrift(types.DriftAlert{
			ResourceID:   fmt.Sprintf("i-%d", i),
			ResourceType: "aws_instance",
			Timestamp:    time.Now().Format(time.RFC3339),
		})
	}

	handler := NewDriftsHandler(store)

	// Request page beyond total data
	req := httptest.NewRequest("GET", "/api/v1/drifts?page=100&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetDrifts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	pagResp := resp.Data.(map[string]interface{})
	data := pagResp["data"].([]interface{})

	// Should have no data but still be valid response
	if len(data) != 0 {
		t.Errorf("expected empty data for page beyond range, got %d items", len(data))
	}
}

// ===== DriftsHandler Tests =====

func TestDriftsHandler_GetDrifts_WithData(t *testing.T) {
	store := graph.NewStore()

	// Add multiple drifts
	for i := 0; i < 5; i++ {
		store.AddDrift(types.DriftAlert{
			ResourceID:   fmt.Sprintf("i-%d", i),
			ResourceType: "aws_instance",
			Severity:     map[int]string{0: "high", 1: "high", 2: "medium", 3: "low", 4: "critical"}[i],
			Timestamp:    time.Now().Format(time.RFC3339),
		})
	}

	handler := NewDriftsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/drifts?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetDrifts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify pagination
	pagResp := resp.Data.(map[string]interface{})
	if total, ok := pagResp["total"].(float64); !ok || total != 5 {
		t.Errorf("expected total 5, got %v", total)
	}
}

func TestDriftsHandler_GetDrifts_WithSeverityFilter(t *testing.T) {
	store := graph.NewStore()

	store.AddDrift(types.DriftAlert{
		ResourceID:   "i-1",
		ResourceType: "aws_instance",
		Severity:     "high",
		Timestamp:    time.Now().Format(time.RFC3339),
	})
	store.AddDrift(types.DriftAlert{
		ResourceID:   "i-2",
		ResourceType: "aws_instance",
		Severity:     "low",
		Timestamp:    time.Now().Format(time.RFC3339),
	})

	handler := NewDriftsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/drifts?severity=high", nil)
	w := httptest.NewRecorder()

	handler.GetDrifts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	pagResp := resp.Data.(map[string]interface{})
	if total, ok := pagResp["total"].(float64); !ok || total != 1 {
		t.Errorf("expected total 1 (filtered), got %v", total)
	}
}

func TestDriftsHandler_GetDrifts_WithResourceTypeFilter(t *testing.T) {
	store := graph.NewStore()

	store.AddDrift(types.DriftAlert{
		ResourceID:   "i-1",
		ResourceType: "aws_instance",
		Timestamp:    time.Now().Format(time.RFC3339),
	})
	store.AddDrift(types.DriftAlert{
		ResourceID:   "bucket-1",
		ResourceType: "aws_s3_bucket",
		Timestamp:    time.Now().Format(time.RFC3339),
	})

	handler := NewDriftsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/drifts?resource_type=aws_s3_bucket", nil)
	w := httptest.NewRecorder()

	handler.GetDrifts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	pagResp := resp.Data.(map[string]interface{})
	if total, ok := pagResp["total"].(float64); !ok || total != 1 {
		t.Errorf("expected total 1 (filtered), got %v", total)
	}
}

func TestDriftsHandler_GetDrifts_Pagination(t *testing.T) {
	store := graph.NewStore()

	// Add 15 drifts
	for i := 0; i < 15; i++ {
		store.AddDrift(types.DriftAlert{
			ResourceID:   fmt.Sprintf("i-%d", i),
			ResourceType: "aws_instance",
			Timestamp:    time.Now().Format(time.RFC3339),
		})
	}

	handler := NewDriftsHandler(store)

	// Test page 1 with limit 10
	req := httptest.NewRequest("GET", "/api/v1/drifts?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	handler.GetDrifts(w, req)

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	pagResp := resp.Data.(map[string]interface{})

	if limit, ok := pagResp["limit"].(float64); !ok || limit != 10 {
		t.Errorf("expected limit 10, got %v", limit)
	}
	if page, ok := pagResp["page"].(float64); !ok || page != 1 {
		t.Errorf("expected page 1, got %v", page)
	}

	// Test page 2
	req = httptest.NewRequest("GET", "/api/v1/drifts?page=2&limit=10", nil)
	w = httptest.NewRecorder()
	handler.GetDrifts(w, req)
	json.Unmarshal(w.Body.Bytes(), &resp)
	pagResp = resp.Data.(map[string]interface{})

	if page, ok := pagResp["page"].(float64); !ok || page != 2 {
		t.Errorf("expected page 2, got %v", page)
	}
}

func TestDriftsHandler_GetDrift_Success(t *testing.T) {
	store := graph.NewStore()

	store.AddDrift(types.DriftAlert{
		ResourceID:   "i-123",
		ResourceType: "aws_instance",
		ResourceName: "web-server",
		Severity:     "high",
		Attribute:    "instance_type",
		OldValue:     "t2.micro",
		NewValue:     "t2.small",
		Timestamp:    time.Now().Format(time.RFC3339),
	})

	handler := NewDriftsHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "i-123")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/drifts/i-123", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDrift(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestDriftsHandler_GetDrift_NotFound(t *testing.T) {
	store := graph.NewStore()
	handler := NewDriftsHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/drifts/nonexistent", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetDrift(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ===== EventsHandler Tests =====

func TestEventsHandler_GetEvents_WithData(t *testing.T) {
	store := graph.NewStore()

	// Add multiple events
	for i := 0; i < 3; i++ {
		store.AddEvent(types.Event{
			Provider:     "aws",
			EventName:    "RunInstances",
			ResourceType: "aws_instance",
			ResourceID:   fmt.Sprintf("i-%d", i),
			UserIdentity: types.UserIdentity{UserName: "alice"},
		})
	}

	handler := NewEventsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/events?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	pagResp := resp.Data.(map[string]interface{})
	if total, ok := pagResp["total"].(float64); !ok || total != 3 {
		t.Errorf("expected total 3, got %v", total)
	}
}

func TestEventsHandler_GetEvents_WithProviderFilter(t *testing.T) {
	store := graph.NewStore()

	store.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		ResourceID:   "i-1",
	})
	store.AddEvent(types.Event{
		Provider:     "gcp",
		EventName:    "compute.instances.insert",
		ResourceType: "google_compute_instance",
		ResourceID:   "inst-1",
	})

	handler := NewEventsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/events?provider=aws", nil)
	w := httptest.NewRecorder()

	handler.GetEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	pagResp := resp.Data.(map[string]interface{})
	if total, ok := pagResp["total"].(float64); !ok || total != 1 {
		t.Errorf("expected total 1 (filtered), got %v", total)
	}
}

func TestEventsHandler_GetEvent_Success(t *testing.T) {
	store := graph.NewStore()

	store.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "PutObject",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "bucket-123",
		UserIdentity: types.UserIdentity{UserName: "alice"},
		Changes:      map[string]interface{}{"acl": "private"},
	})

	handler := NewEventsHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "bucket-123")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/events/bucket-123", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestEventsHandler_GetEvents_Pagination(t *testing.T) {
	store := graph.NewStore()

	// Add 20 events
	for i := 0; i < 20; i++ {
		store.AddEvent(types.Event{
			Provider:     "aws",
			EventName:    "RunInstances",
			ResourceType: "aws_instance",
			ResourceID:   fmt.Sprintf("i-%d", i),
		})
	}

	handler := NewEventsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/events?page=1&limit=5", nil)
	w := httptest.NewRecorder()

	handler.GetEvents(w, req)

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	pagResp := resp.Data.(map[string]interface{})

	if totalPages, ok := pagResp["total_pages"].(float64); !ok || totalPages != 4 {
		t.Errorf("expected 4 total pages, got %v", totalPages)
	}
}

func TestEventsHandler_GetEvent_NotFound(t *testing.T) {
	store := graph.NewStore()
	handler := NewEventsHandler(store)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	ctx := context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
	req := httptest.NewRequest("GET", "/api/v1/events/nonexistent", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetEvent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ===== HealthHandler Tests =====

func TestHealthHandler_GetHealth(t *testing.T) {
	handler := NewHealthHandler("1.0.0")
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.GetHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	healthResp := resp.Data.(map[string]interface{})
	if version, ok := healthResp["version"].(string); !ok || version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %v", version)
	}
}

// ===== StatsHandler Tests =====

func TestStatsHandler_GetStats_Success(t *testing.T) {
	store := graph.NewStore()

	// Add some data
	store.AddDrift(types.DriftAlert{
		ResourceID:   "i-1",
		ResourceType: "aws_instance",
		Severity:     "high",
		Timestamp:    time.Now().Format(time.RFC3339),
	})

	store.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		ResourceID:   "i-2",
	})

	handler := NewStatsHandler(store)
	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

// ===== ProviderStatusHandler Tests =====

func TestProviderStatusHandler_RecordEvent(t *testing.T) {
	mockRegistry := createMockRegistry()
	handler := NewProviderStatusHandler(mockRegistry)

	// Record an event
	handler.RecordEvent("aws", true)

	// Verify through GetProviderStatus
	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestProviderStatusHandler_RecordError(t *testing.T) {
	mockRegistry := createMockRegistry()
	handler := NewProviderStatusHandler(mockRegistry)

	handler.RecordError("aws")

	// Verify through GetProviderStatus
	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestProviderStatusHandler_GetProviderStatus_WithStats(t *testing.T) {
	mockRegistry := createMockRegistry()
	handler := NewProviderStatusHandler(mockRegistry)

	// Record various events to populate stats
	handler.RecordEvent("aws", true)
	handler.RecordEvent("aws", true)
	handler.RecordEvent("aws", false)
	handler.RecordError("gcp")

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify structure has providers and stats
	data := resp.Data.(map[string]interface{})
	if _, ok := data["providers"]; !ok {
		t.Error("expected 'providers' key in response")
	}
	if _, ok := data["count"]; !ok {
		t.Error("expected 'count' key in response")
	}
	if _, ok := data["timestamp"]; !ok {
		t.Error("expected 'timestamp' key in response")
	}
}

func TestProviderStatusHandler_GetProviderSummary(t *testing.T) {
	mockRegistry := createMockRegistry()
	handler := NewProviderStatusHandler(mockRegistry)

	handler.RecordEvent("aws", true)
	handler.RecordEvent("gcp", false)

	req := httptest.NewRequest("GET", "/api/v1/providers/summary", nil)
	w := httptest.NewRecorder()

	handler.GetProviderSummary(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify summary contains expected fields
	data := resp.Data.(map[string]interface{})
	if _, ok := data["total_providers"]; !ok {
		t.Error("expected 'total_providers' key")
	}
	if _, ok := data["total_events"]; !ok {
		t.Error("expected 'total_events' key")
	}
	if _, ok := data["match_rate_percent"]; !ok {
		t.Error("expected 'match_rate_percent' key")
	}
}

// ===== Helper Functions =====

type MockProviderForTests struct {
	name string
}

func (m *MockProviderForTests) Name() string {
	return m.name
}

func (m *MockProviderForTests) ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event {
	return nil
}

func (m *MockProviderForTests) IsRelevantEvent(eventName string) bool {
	return true
}

func (m *MockProviderForTests) MapEventToResource(eventName string, eventSource string) string {
	return "test_resource"
}

func (m *MockProviderForTests) ExtractChanges(eventName string, fields map[string]string) map[string]interface{} {
	return make(map[string]interface{})
}

func (m *MockProviderForTests) SupportedEventCount() int {
	return 50
}

func (m *MockProviderForTests) SupportedResourceTypes() []string {
	return []string{"test_resource"}
}

func createMockRegistry() *provider.Registry {
	registry := provider.NewRegistry()
	registry.Register(&MockProviderForTests{name: "aws"})
	registry.Register(&MockProviderForTests{name: "gcp"})
	return registry
}
