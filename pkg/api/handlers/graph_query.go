package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// GraphQueryHandler handles graph database queries
type GraphQueryHandler struct {
	graphStore *graph.Store
}

// NewGraphQueryHandler creates a new graph query handler
func NewGraphQueryHandler(graphStore *graph.Store) *GraphQueryHandler {
	return &GraphQueryHandler{
		graphStore: graphStore,
	}
}

// GetNode returns a specific node by ID
// GET /api/v1/graph/nodes/:id
func (h *GraphQueryHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")

	db := h.graphStore.GetGraphDB()
	node := db.GetNode(nodeID)

	if node == nil {
		respondError(w, http.StatusNotFound, "Node not found")
		return
	}

	respondJSON(w, http.StatusOK, node)
}

// GetNodesByLabel returns all nodes with a specific label
// GET /api/v1/graph/nodes?label=EC2
func (h *GraphQueryHandler) GetNodesByLabel(w http.ResponseWriter, r *http.Request) {
	label := r.URL.Query().Get("label")

	if label == "" {
		respondError(w, http.StatusBadRequest, "label parameter is required")
		return
	}

	db := h.graphStore.GetGraphDB()
	nodes := db.GetNodesByLabel(label)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"label": label,
		"count": len(nodes),
		"nodes": nodes,
	})
}

// GetPath finds the shortest path between two nodes
// GET /api/v1/graph/path?from=node1&to=node2
func (h *GraphQueryHandler) GetPath(w http.ResponseWriter, r *http.Request) {
	fromID := r.URL.Query().Get("from")
	toID := r.URL.Query().Get("to")

	if fromID == "" || toID == "" {
		respondError(w, http.StatusBadRequest, "from and to parameters are required")
		return
	}

	db := h.graphStore.GetGraphDB()
	path, err := db.FindPath(fromID, toID)

	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"from":   fromID,
		"to":     toID,
		"path":   path,
		"length": path.Length,
	})
}

// GetImpactRadius finds all nodes within N hops of a node
// GET /api/v1/graph/impact/:id?depth=3
func (h *GraphQueryHandler) GetImpactRadius(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	depthStr := r.URL.Query().Get("depth")
	if depthStr == "" {
		depthStr = "3"
	}

	depth, err := strconv.Atoi(depthStr)
	if err != nil || depth < 1 || depth > 10 {
		respondError(w, http.StatusBadRequest, "depth must be between 1 and 10")
		return
	}

	db := h.graphStore.GetGraphDB()
	result := db.FindImpactRadius(nodeID, depth)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"node_id":        nodeID,
		"depth":          depth,
		"affected_count": len(result.Nodes),
		"nodes":          result.Nodes,
		"distances":      result.Distances,
	})
}

// GetDependencies finds all dependencies of a node
// GET /api/v1/graph/dependencies/:id?depth=5
func (h *GraphQueryHandler) GetDependencies(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	depthStr := r.URL.Query().Get("depth")
	if depthStr == "" {
		depthStr = "5"
	}

	depth, err := strconv.Atoi(depthStr)
	if err != nil || depth < 1 || depth > 10 {
		respondError(w, http.StatusBadRequest, "depth must be between 1 and 10")
		return
	}

	db := h.graphStore.GetGraphDB()
	log.Infof("[GetDependencies Handler] Received GraphDB instance: %p, calling FindDependencies(%s, %d)", db, nodeID, depth)
	dependencies := db.FindDependencies(nodeID, depth)
	log.Infof("[GetDependencies Handler] FindDependencies returned %d dependencies", len(dependencies))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"node_id":      nodeID,
		"depth":        depth,
		"count":        len(dependencies),
		"dependencies": dependencies,
	})
}

// GetDependents finds all nodes that depend on this node
// GET /api/v1/graph/dependents/:id?depth=5
func (h *GraphQueryHandler) GetDependents(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	depthStr := r.URL.Query().Get("depth")
	if depthStr == "" {
		depthStr = "5"
	}

	depth, err := strconv.Atoi(depthStr)
	if err != nil || depth < 1 || depth > 10 {
		respondError(w, http.StatusBadRequest, "depth must be between 1 and 10")
		return
	}

	db := h.graphStore.GetGraphDB()
	log.Infof("[GetDependents Handler] Received GraphDB instance: %p, calling FindDependents(%s, %d)", db, nodeID, depth)
	dependents := db.FindDependents(nodeID, depth)
	log.Infof("[GetDependents Handler] FindDependents returned %d dependents", len(dependents))

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"node_id":    nodeID,
		"depth":      depth,
		"count":      len(dependents),
		"dependents": dependents,
	})
}

// GetCriticalNodes finds critical infrastructure nodes
// GET /api/v1/graph/critical?min=3
func (h *GraphQueryHandler) GetCriticalNodes(w http.ResponseWriter, r *http.Request) {
	minStr := r.URL.Query().Get("min")
	if minStr == "" {
		minStr = "3"
	}

	min, err := strconv.Atoi(minStr)
	if err != nil || min < 1 {
		respondError(w, http.StatusBadRequest, "min must be a positive integer")
		return
	}

	db := h.graphStore.GetGraphDB()
	criticalNodes := db.FindCriticalPaths(min)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"min_dependents": min,
		"count":          len(criticalNodes),
		"critical_nodes": criticalNodes,
	})
}

// GetNeighbors returns all directly connected nodes
// GET /api/v1/graph/neighbors/:id
func (h *GraphQueryHandler) GetNeighbors(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")

	db := h.graphStore.GetGraphDB()
	neighbors := db.GetNeighbors(nodeID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"node_id":   nodeID,
		"count":     len(neighbors),
		"neighbors": neighbors,
	})
}

// GetRelationships returns relationships for a node
// GET /api/v1/graph/relationships/:id?direction=outgoing
func (h *GraphQueryHandler) GetRelationships(w http.ResponseWriter, r *http.Request) {
	nodeID := chi.URLParam(r, "id")
	direction := r.URL.Query().Get("direction")
	if direction == "" {
		direction = "both"
	}

	db := h.graphStore.GetGraphDB()

	var relationships []*graph.Relationship

	switch direction {
	case "outgoing":
		relationships = db.GetOutgoingRelationships(nodeID)
	case "incoming":
		relationships = db.GetIncomingRelationships(nodeID)
	case "both":
		outgoing := db.GetOutgoingRelationships(nodeID)
		incoming := db.GetIncomingRelationships(nodeID)
		relationships = append(outgoing, incoming...)
	default:
		respondError(w, http.StatusBadRequest, "direction must be 'outgoing', 'incoming', or 'both'")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"node_id":       nodeID,
		"direction":     direction,
		"count":         len(relationships),
		"relationships": relationships,
	})
}

// GetGraphStats returns overall graph statistics
// GET /api/v1/graph/stats
func (h *GraphQueryHandler) GetGraphStats(w http.ResponseWriter, r *http.Request) {
	db := h.graphStore.GetGraphDB()

	// Count nodes by label
	labelCounts := make(map[string]int)
	allNodes := db.GetAllNodes()
	for _, node := range allNodes {
		for _, label := range node.Labels {
			labelCounts[label]++
		}
	}

	// Count relationships by type
	typeCounts := make(map[string]int)
	allRels := db.GetAllRelationships()
	for _, rel := range allRels {
		typeCounts[rel.Type]++
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"node_count":            db.NodeCount(),
		"relationship_count":    db.RelationshipCount(),
		"nodes_by_label":        labelCounts,
		"relationships_by_type": typeCounts,
	})
}

// MatchPattern performs pattern matching query
// POST /api/v1/graph/match
// Body: {"start_labels": ["EC2"], "rel_type": "DEPENDS_ON", "end_labels": ["Subnet"], "end_filter": {"id": "subnet-123"}}
func (h *GraphQueryHandler) MatchPattern(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartLabels []string               `json:"start_labels"`
		RelType     string                 `json:"rel_type"`
		EndLabels   []string               `json:"end_labels"`
		EndFilter   map[string]interface{} `json:"end_filter"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pattern := &graph.MatchPattern{
		StartLabels: req.StartLabels,
		RelType:     req.RelType,
		EndLabels:   req.EndLabels,
		EndFilter:   req.EndFilter,
	}

	db := h.graphStore.GetGraphDB()
	matches := db.Match(pattern)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pattern": pattern,
		"count":   len(matches),
		"matches": matches,
	})
}
