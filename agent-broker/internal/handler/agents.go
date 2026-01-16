package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lunarr-ai/lunarr/agent-broker/internal/registry"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/store"
)

// AgentsHandler handles public agent endpoints.
type AgentsHandler struct {
	// registry is the service for agent lookups.
	registry *registry.RegistryService
}

// NewAgentsHandler creates an AgentsHandler.
func NewAgentsHandler(reg *registry.RegistryService) *AgentsHandler {
	return &AgentsHandler{registry: reg}
}

// RegisterRoutes registers agent routes on the given ServeMux.
func (h *AgentsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/agents/{id}/card", h.handleGetCard)
}

func (h *AgentsHandler) handleGetCard(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("id")

	agent, err := h.registry.Get(r.Context(), agentID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "AGENT_NOT_FOUND",
				"agent with ID '"+agentID+"' not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agent.Card)
}
