package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a2aproject/a2a-go/a2a"

	"github.com/lunarr-ai/lunarr/agent-broker/internal/registry"
	"github.com/lunarr-ai/lunarr/agent-broker/internal/store"
)

func validAgentCard() a2a.AgentCard {
	return a2a.AgentCard{
		Name:        "Test Agent",
		Description: "A test agent",
		URL:         "http://localhost:9000",
		Version:     "1.0.0",
		Skills: []a2a.AgentSkill{
			{ID: "skill-1", Name: "Skill One"},
		},
	}
}

func validRegisterRequest() RegisterAgentRequest {
	return RegisterAgentRequest{
		AgentID:   "test-agent",
		AgentCard: validAgentCard(),
		Tags:      []string{"test"},
	}
}

func setupHandler() (*AdminHandler, *http.ServeMux) {
	s := store.NewMemoryStore()
	svc := registry.NewRegistryService(s)
	h := NewAdminHandler(svc)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return h, mux
}

func makeJSONRequest(method, path string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestAdminHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		req := makeJSONRequest(http.MethodPost, "/v1/admin/agents", validRegisterRequest())
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
		}
		var resp AgentRecordResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.AgentID != "test-agent" {
			t.Errorf("AgentID = %v, want test-agent", resp.AgentID)
		}
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		req := httptest.NewRequest(http.MethodPost, "/v1/admin/agents", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Code != "INVALID_JSON" {
			t.Errorf("error code = %v, want INVALID_JSON", resp.Code)
		}
	})

	t.Run("duplicate ID returns 409", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		body := validRegisterRequest()
		req1 := makeJSONRequest(http.MethodPost, "/v1/admin/agents", body)
		rec1 := httptest.NewRecorder()
		mux.ServeHTTP(rec1, req1)

		req2 := makeJSONRequest(http.MethodPost, "/v1/admin/agents", body)
		rec2 := httptest.NewRecorder()
		mux.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusConflict {
			t.Errorf("status = %d, want %d", rec2.Code, http.StatusConflict)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(rec2.Body).Decode(&resp)
		if resp.Code != "AGENT_EXISTS" {
			t.Errorf("error code = %v, want AGENT_EXISTS", resp.Code)
		}
	})

}

func TestAdminHandler_Get(t *testing.T) {
	t.Parallel()

	t.Run("existing agent returns 200", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		createReq := makeJSONRequest(http.MethodPost, "/v1/admin/agents", validRegisterRequest())
		createRec := httptest.NewRecorder()
		mux.ServeHTTP(createRec, createReq)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/agents/test-agent", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var resp AgentRecordResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if resp.AgentID != "test-agent" {
			t.Errorf("AgentID = %v, want test-agent", resp.AgentID)
		}
	})

	t.Run("non-existent returns 404", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/agents/not-exists", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
		var resp ErrorResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Code != "AGENT_NOT_FOUND" {
			t.Errorf("error code = %v, want AGENT_NOT_FOUND", resp.Code)
		}
	})
}

func TestAdminHandler_List(t *testing.T) {
	t.Parallel()

	t.Run("empty list returns 200", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/agents", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var resp AgentListResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if len(resp.Agents) != 0 {
			t.Errorf("agents count = %d, want 0", len(resp.Agents))
		}
		if resp.Pagination.Total != 0 {
			t.Errorf("total = %d, want 0", resp.Pagination.Total)
		}
	})

	t.Run("with agents returns pagination", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		body := validRegisterRequest()
		createReq := makeJSONRequest(http.MethodPost, "/v1/admin/agents", body)
		createRec := httptest.NewRecorder()
		mux.ServeHTTP(createRec, createReq)

		req := httptest.NewRequest(http.MethodGet, "/v1/admin/agents", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var resp AgentListResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if len(resp.Agents) != 1 {
			t.Errorf("agents count = %d, want 1", len(resp.Agents))
		}
		if resp.Pagination.Total != 1 {
			t.Errorf("total = %d, want 1", resp.Pagination.Total)
		}
		if resp.Pagination.HasMore {
			t.Error("HasMore should be false")
		}
	})

	t.Run("query params parsed", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		req := httptest.NewRequest(http.MethodGet, "/v1/admin/agents?offset=5&limit=10&tags=a,b&skills=s1&q=search", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var resp AgentListResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.Pagination.Offset != 5 {
			t.Errorf("offset = %d, want 5", resp.Pagination.Offset)
		}
		if resp.Pagination.Limit != 10 {
			t.Errorf("limit = %d, want 10", resp.Pagination.Limit)
		}
	})
}

func TestAdminHandler_Update(t *testing.T) {
	t.Parallel()

	t.Run("valid update returns 200", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		createReq := makeJSONRequest(http.MethodPost, "/v1/admin/agents", validRegisterRequest())
		createRec := httptest.NewRecorder()
		mux.ServeHTTP(createRec, createReq)

		updateBody := UpdateAgentRequest{
			AgentCard: validAgentCard(),
			Tags:      []string{"updated"},
		}
		updateBody.AgentCard.Name = "Updated Name"
		req := makeJSONRequest(http.MethodPut, "/v1/admin/agents/test-agent", updateBody)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		var resp AgentRecordResponse
		_ = json.NewDecoder(rec.Body).Decode(&resp)
		if resp.AgentCard.Name != "Updated Name" {
			t.Errorf("Name = %v, want Updated Name", resp.AgentCard.Name)
		}
	})

	t.Run("non-existent returns 404", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		body := UpdateAgentRequest{AgentCard: validAgentCard()}
		req := makeJSONRequest(http.MethodPut, "/v1/admin/agents/not-exists", body)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestAdminHandler_Delete(t *testing.T) {
	t.Parallel()

	t.Run("existing returns 204", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		createReq := makeJSONRequest(http.MethodPost, "/v1/admin/agents", validRegisterRequest())
		createRec := httptest.NewRecorder()
		mux.ServeHTTP(createRec, createReq)

		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/agents/test-agent", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if rec.Body.Len() != 0 {
			t.Error("body should be empty")
		}
	})

	t.Run("non-existent returns 404", func(t *testing.T) {
		t.Parallel()
		_, mux := setupHandler()
		req := httptest.NewRequest(http.MethodDelete, "/v1/admin/agents/not-exists", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestToAgentResponse(t *testing.T) {
	t.Parallel()

	t.Run("maps fields correctly", func(t *testing.T) {
		t.Parallel()
		agent := &store.RegisteredAgent{
			ID:   "test-agent",
			Card: validAgentCard(),
			Tags: []string{"tag1", "tag2"},
		}

		resp := toAgentResponse(agent)

		if resp.AgentID != "test-agent" {
			t.Errorf("AgentID = %v, want test-agent", resp.AgentID)
		}
		if resp.Endpoint != "http://localhost:9000" {
			t.Errorf("Endpoint = %v, want http://localhost:9000", resp.Endpoint)
		}
		if len(resp.Skills) != 1 || resp.Skills[0] != "skill-1" {
			t.Errorf("Skills = %v, want [skill-1]", resp.Skills)
		}
		if len(resp.Tags) != 2 {
			t.Errorf("Tags = %v, want [tag1 tag2]", resp.Tags)
		}
	})

	t.Run("nil tags returns empty array", func(t *testing.T) {
		t.Parallel()
		agent := &store.RegisteredAgent{
			ID:   "test-agent",
			Card: validAgentCard(),
			Tags: nil,
		}

		resp := toAgentResponse(agent)

		if resp.Tags == nil {
			t.Error("Tags should be empty array, not nil")
		}
		if len(resp.Tags) != 0 {
			t.Errorf("Tags = %v, want []", resp.Tags)
		}
	})
}
