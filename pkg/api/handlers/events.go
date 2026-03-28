package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// EventsHandler handles Falco events-related requests
type EventsHandler struct {
	store *graph.Store
}

// NewEventsHandler creates a new events handler
func NewEventsHandler(store *graph.Store) *EventsHandler {
	return &EventsHandler{
		store: store,
	}
}

// GetEvents handles GET /api/v1/events
func (h *EventsHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/events")

	// Parse pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	params := models.NewPaginationParams(page, limit)

	// Parse filter parameters
	severity := r.URL.Query().Get("severity")
	provider := r.URL.Query().Get("provider")

	// Get all events
	allEvents := h.store.GetEvents()

	// Apply filters
	filteredEvents := make([]map[string]interface{}, 0)
	for _, event := range allEvents {
		// Apply severity filter (if provided)
		if severity != "" && event.ResourceType != severity {
			continue
		}

		// Apply provider filter
		if provider != "" && event.Provider != provider {
			continue
		}

		eventData := map[string]interface{}{
			"id":            event.ResourceID,
			"provider":      event.Provider,
			"event_name":    event.EventName,
			"resource_type": event.ResourceType,
			"resource_id":   event.ResourceID,
			"user_identity": event.UserIdentity,
			"changes":       event.Changes,
			"region":        event.Region,
			"project_id":    event.ProjectID,
			"service_name":  event.ServiceName,
		}
		filteredEvents = append(filteredEvents, eventData)
	}

	// Apply pagination
	total := len(filteredEvents)
	start := params.Offset()
	end := start + params.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedEvents := filteredEvents[start:end]

	response := models.PaginatedResponse{
		Data:       paginatedEvents,
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

// GetEvent handles GET /api/v1/events/:id
func (h *EventsHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	log.Debugf("GET /api/v1/events/%s", eventID)

	// Get all events
	allEvents := h.store.GetEvents()

	// Find event by ID
	for _, event := range allEvents {
		if event.ResourceID == eventID {
			eventData := map[string]interface{}{
				"id":            event.ResourceID,
				"provider":      event.Provider,
				"event_name":    event.EventName,
				"resource_type": event.ResourceType,
				"resource_id":   event.ResourceID,
				"user_identity": event.UserIdentity,
				"changes":       event.Changes,
				"region":        event.Region,
				"project_id":    event.ProjectID,
				"service_name":  event.ServiceName,
				"raw_event":     event.RawEvent,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(models.APIResponse{
				Success: true,
				Data:    eventData,
			})
			return
		}
	}

	respondError(w, http.StatusNotFound, "Event not found")
}
