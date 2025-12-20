package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    graphData,
	})
}

// GetNodes handles GET /api/v1/graph/nodes with pagination
func (h *GraphHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/graph/nodes")

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	params := models.NewPaginationParams(page, limit)

	// Build full graph
	graphData := h.store.BuildGraph()

	// Apply pagination
	total := len(graphData.Nodes)
	start := params.Offset()
	end := start + params.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedNodes := graphData.Nodes[start:end]

	response := models.PaginatedResponse{
		Data:       paginatedNodes,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: models.CalculateTotalPages(total, params.Limit),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetEdges handles GET /api/v1/graph/edges with pagination
func (h *GraphHandler) GetEdges(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/graph/edges")

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	params := models.NewPaginationParams(page, limit)

	// Build full graph
	graphData := h.store.BuildGraph()

	// Apply pagination
	total := len(graphData.Edges)
	start := params.Offset()
	end := start + params.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedEdges := graphData.Edges[start:end]

	response := models.PaginatedResponse{
		Data:       paginatedEdges,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: models.CalculateTotalPages(total, params.Limit),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    response,
	})
}
