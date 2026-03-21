package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
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

// eventToMap converts an Event struct to a map for JSON serialization
func eventToMap(event types.Event) map[string]interface{} {
	status := string(event.Status)
	if status == "" {
		status = string(types.EventStatusOpen)
	}
	return map[string]interface{}{
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
		"timestamp":     event.Timestamp,
		"severity":      event.Severity,
		"status":        status,
		"status_reason": event.StatusReason,
	}
}

// eventToDetailMap converts an Event struct to a detailed map (includes raw_event)
func eventToDetailMap(event types.Event) map[string]interface{} {
	m := eventToMap(event)
	m["raw_event"] = event.RawEvent
	return m
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
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	sortField := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order") // "asc" or "desc" (default: desc)

	// Parse time filters
	var fromTime, toTime time.Time
	var hasFrom, hasTo bool
	if fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			fromTime = t
			hasFrom = true
		}
	}
	if toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			toTime = t
			hasTo = true
		}
	}

	// Get all events
	allEvents := h.store.GetEvents()

	// Apply filters
	filteredEvents := make([]map[string]interface{}, 0)
	for _, event := range allEvents {
		// Apply severity filter
		if severity != "" && !strings.EqualFold(event.Severity, severity) {
			continue
		}

		// Apply provider filter
		if provider != "" && !strings.EqualFold(event.Provider, provider) {
			continue
		}

		// Apply status filter
		if status != "" {
			eventStatus := string(event.Status)
			if eventStatus == "" {
				eventStatus = string(types.EventStatusOpen)
			}
			if !strings.EqualFold(eventStatus, status) {
				continue
			}
		}

		// Apply time range filters
		if hasFrom || hasTo {
			if event.Timestamp != "" {
				if eventTime, err := time.Parse(time.RFC3339, event.Timestamp); err == nil {
					if hasFrom && eventTime.Before(fromTime) {
						continue
					}
					if hasTo && eventTime.After(toTime) {
						continue
					}
				}
			}
		}

		// Apply search filter (searches across resource_id, resource_type, event_name, user)
		if search != "" {
			searchLower := strings.ToLower(search)
			match := strings.Contains(strings.ToLower(event.ResourceID), searchLower) ||
				strings.Contains(strings.ToLower(event.ResourceType), searchLower) ||
				strings.Contains(strings.ToLower(event.EventName), searchLower) ||
				strings.Contains(strings.ToLower(event.UserIdentity.UserName), searchLower) ||
				strings.Contains(strings.ToLower(event.UserIdentity.ARN), searchLower) ||
				strings.Contains(strings.ToLower(event.Region), searchLower)
			if !match {
				continue
			}
		}

		filteredEvents = append(filteredEvents, eventToMap(event))
	}

	// Apply sorting
	if sortField != "" {
		ascending := strings.ToLower(sortOrder) == "asc"
		sort.Slice(filteredEvents, func(i, j int) bool {
			vi := fmt.Sprintf("%v", filteredEvents[i][sortField])
			vj := fmt.Sprintf("%v", filteredEvents[j][sortField])
			if ascending {
				return vi < vj
			}
			return vi > vj
		})
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

	event, found := h.store.GetEventByID(eventID)
	if !found {
		respondError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Get related drifts for this event
	relatedDrifts := make([]map[string]interface{}, 0)
	for _, drift := range h.store.GetDrifts() {
		if drift.ResourceID == eventID {
			relatedDrifts = append(relatedDrifts, map[string]interface{}{
				"severity":      drift.Severity,
				"attribute":     drift.Attribute,
				"old_value":     drift.OldValue,
				"new_value":     drift.NewValue,
				"matched_rules": drift.MatchedRules,
				"timestamp":     drift.Timestamp,
				"alert_type":    drift.AlertType,
			})
		}
	}

	eventData := eventToDetailMap(*event)
	eventData["related_drifts"] = relatedDrifts

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    eventData,
	})
}

// UpdateEventStatusRequest is the request body for PATCH /api/v1/events/:id
type UpdateEventStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// UpdateEventStatus handles PATCH /api/v1/events/:id
func (h *EventsHandler) UpdateEventStatus(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	log.Debugf("PATCH /api/v1/events/%s", eventID)

	var req UpdateEventStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate status
	validStatuses := map[string]types.EventStatus{
		"open":         types.EventStatusOpen,
		"acknowledged": types.EventStatusAcknowledged,
		"ignored":      types.EventStatusIgnored,
		"resolved":     types.EventStatusResolved,
	}

	newStatus, valid := validStatuses[strings.ToLower(req.Status)]
	if !valid {
		respondError(w, http.StatusBadRequest, "Invalid status. Must be one of: open, acknowledged, ignored, resolved")
		return
	}

	// Require reason for ignored status
	if newStatus == types.EventStatusIgnored && req.Reason == "" {
		respondError(w, http.StatusBadRequest, "Reason is required when setting status to 'ignored'")
		return
	}

	// Update in store
	if !h.store.UpdateEventStatus(eventID, newStatus, req.Reason) {
		respondError(w, http.StatusNotFound, "Event not found")
		return
	}

	// Return updated event
	event, _ := h.store.GetEventByID(eventID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    eventToMap(*event),
	})
}
