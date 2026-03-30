// Package handlers provides HTTP handlers for the API.
package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
	params := ParsePagination(r, 50)

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
		}

		// Populate optional fields from Metadata map
		if event.Metadata != nil {
			if v, ok := event.Metadata["region"]; ok {
				eventData["region"] = v
			}
			if v, ok := event.Metadata["project_id"]; ok {
				eventData["project_id"] = v
			}
			if v, ok := event.Metadata["service_name"]; ok {
				eventData["service_name"] = v
			}
		}

		filteredEvents = append(filteredEvents, eventData)
	}

	// Apply pagination
	total := len(filteredEvents)
	paginatedEvents := Paginate(filteredEvents, params)

	response := PaginatedResponseData(paginatedEvents, params, total)
	respondJSON(w, http.StatusOK, response)
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
				"raw_event":     event.RawEvent,
			}

			// Populate optional fields from Metadata map
			if event.Metadata != nil {
				if v, ok := event.Metadata["region"]; ok {
					eventData["region"] = v
				}
				if v, ok := event.Metadata["project_id"]; ok {
					eventData["project_id"] = v
				}
				if v, ok := event.Metadata["service_name"]; ok {
					eventData["service_name"] = v
				}
			}

			respondJSON(w, http.StatusOK, eventData)
			return
		}
	}

	respondError(w, http.StatusNotFound, "Event not found")
}
