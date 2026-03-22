package handlers

import (
	_ "embed"
	"net/http"
)

//go:embed swagger_ui.html
var swaggerUIHTML []byte

//go:embed openapi.yaml
var openapiSpec []byte

// SwaggerHandler serves the Swagger UI and OpenAPI spec.
type SwaggerHandler struct{}

// NewSwaggerHandler creates a new Swagger UI handler.
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// ServeUI serves the Swagger UI HTML page.
func (h *SwaggerHandler) ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(swaggerUIHTML)
}

// ServeSpec serves the OpenAPI YAML specification.
func (h *SwaggerHandler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(openapiSpec)
}
