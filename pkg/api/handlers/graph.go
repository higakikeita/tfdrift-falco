package handlers

import (
	"net/http"

	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// GraphHandler handles graph-related requests
type GraphHandler struct {
	store *graph.Store
}

// NewGraphHandler creates a new graph handler
func NewGraphHandler(store *graph.Store) *GraphHandler {
	return &GraphHandler{
		store: store,
	}
}

// GetGraph handles GET /api/v1/graph
func (h *GraphHandler) GetGraph(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/graph")

	// Build the graph from stored data
	graphData := h.store.BuildGraph()

	respondJSON(w, http.StatusOK, graphData)
}

// GetNodes handles GET /api/v1/graph/nodes with pagination
func (h *GraphHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/graph/nodes")

	// Parse pagination parameters
	params := ParsePagination(r, 50)

	// Build full graph
	graphData := h.store.BuildGraph()

	// Apply pagination
	total := len(graphData.Nodes)
	paginatedNodes := Paginate(graphData.Nodes, params)

	response := PaginatedResponseData(paginatedNodes, params, total)
	respondJSON(w, http.StatusOK, response)
}

// GetEdges handles GET /api/v1/graph/edges with pagination
func (h *GraphHandler) GetEdges(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/graph/edges")

	// Parse pagination parameters
	params := ParsePagination(r, 50)

	// Build full graph
	graphData := h.store.BuildGraph()

	// Apply pagination
	total := len(graphData.Edges)
	paginatedEdges := Paginate(graphData.Edges, params)

	response := PaginatedResponseData(paginatedEdges, params, total)
	respondJSON(w, http.StatusOK, response)
}
