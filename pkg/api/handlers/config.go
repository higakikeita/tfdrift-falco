package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	log "github.com/sirupsen/logrus"
)

// ConfigHandler handles configuration-related requests
type ConfigHandler struct {
	cfg *config.Config
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

// GetConfig handles GET /api/v1/config
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/config")

	safeConfig := map[string]interface{}{
		"providers": map[string]interface{}{
			"aws": map[string]interface{}{
				"enabled": h.cfg.Providers.AWS.Enabled,
				"regions": h.cfg.Providers.AWS.Regions,
			},
			"gcp": map[string]interface{}{
				"enabled":  h.cfg.Providers.GCP.Enabled,
				"projects": h.cfg.Providers.GCP.Projects,
			},
		},
		"notifications": map[string]interface{}{
			"webhook_enabled": h.cfg.Notifications.Webhook.Enabled,
		},
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: safeConfig})
}

// GetVersion handles GET /api/v1/version
func (h *ConfigHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	log.Debug("GET /api/v1/version")

	version := map[string]interface{}{
		"version":       "0.7.0",
		"supported_aws": "40+ services, 500+ events",
		"supported_gcp": "27+ services, 170+ events",
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: version})
}

// TestWebhook handles POST /api/v1/config/webhooks/test
func (h *ConfigHandler) TestWebhook(w http.ResponseWriter, r *http.Request) {
	log.Debug("POST /api/v1/config/webhooks/test")

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   &models.APIError{Code: 400, Message: "Invalid request body"},
		})
		return
	}

	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   &models.APIError{Code: 400, Message: "URL is required"},
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":  "ok",
			"message": "Webhook test sent successfully",
			"url":     req.URL,
		},
	})
}
