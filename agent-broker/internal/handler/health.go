package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lunarr-ai/lunarr/agent-broker/internal/store"
)

// HealthResponse is the JSON response for health check endpoints.
type HealthResponse struct {
	// Status is the overall health status ("healthy" or "unhealthy").
	Status string `json:"status"`
	// Checks contains individual component health statuses.
	Checks HealthChecks `json:"checks"`
}

// HealthChecks contains status of individual health check components.
type HealthChecks struct {
	// Registry is the registry status ("up" or "down").
	Registry string `json:"registry"`
}

// HealthHandler handles HTTP health check requests.
type HealthHandler struct {
	// store is the health checker for storage backend.
	store store.HealthChecker
}

// NewHealthHandler creates a HealthHandler. If checker is nil, always reports healthy.
func NewHealthHandler(checker store.HealthChecker) *HealthHandler {
	return &HealthHandler{store: checker}
}

// ServeHTTP handles GET /health requests.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status: "healthy",
		Checks: HealthChecks{
			Registry: "up",
		},
	}
	statusCode := http.StatusOK

	if h.store != nil {
		if err := h.store.Ping(ctx); err != nil {
			response.Status = "unhealthy"
			response.Checks.Registry = "down"
			statusCode = http.StatusServiceUnavailable
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers health check routes on the given ServeMux.
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("GET /health", h)
}
