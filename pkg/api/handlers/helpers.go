package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
)

// respondJSON writes a JSON response with the given HTTP status code.
// This is the canonical way to write all successful JSON responses.
func respondJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: true,
		Data:    data,
	})
}

// respondError writes a JSON error response.
// This is the canonical way to write all error responses.
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}

// ParsePagination extracts and validates pagination parameters from request.
// If page or limit are not provided or invalid, defaults are applied:
// - page defaults to 1
// - limit defaults to 50, max 1000
func ParsePagination(r *http.Request, defaultLimit int) models.PaginationParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
		if limit < 1 {
			limit = 50
		}
	}
	if limit > 1000 {
		limit = 1000
	}

	return models.PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// Paginate slices a list of items based on pagination parameters.
// It handles boundary cases where start/end exceed the list length.
func Paginate[T any](items []T, params models.PaginationParams) []T {
	total := len(items)
	start := params.Offset()
	end := start + params.Limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start >= end {
		return make([]T, 0)
	}

	return items[start:end]
}

// PaginatedResponseData wraps a slice of items with pagination metadata.
func PaginatedResponseData(items interface{}, params models.PaginationParams, total int) models.PaginatedResponse {
	return models.PaginatedResponse{
		Data:       items,
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: models.CalculateTotalPages(total, params.Limit),
	}
}
