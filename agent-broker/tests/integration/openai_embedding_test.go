package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lunarr-ai/lunarr/agent-broker/pkg/embedding"
)

type embeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model,omitempty"`
}

type embeddingResponse struct {
	Data []embeddingData `json:"data"`
}

type embeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

func TestClient_Embed(t *testing.T) {
	t.Parallel()

	t.Run("single text returns embedding", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/embeddings" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("unexpected method: %s", r.Method)
			}

			var req embeddingRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("decode request: %v", err)
			}
			if len(req.Input) != 1 || req.Input[0] != "hello" {
				t.Errorf("unexpected input: %v", req.Input)
			}

			resp := embeddingResponse{
				Data: []embeddingData{
					{Embedding: []float32{0.1, 0.2, 0.3}, Index: 0},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := embedding.NewClient(server.URL, 3)
		embeddings, err := client.Embed(context.Background(), []string{"hello"})

		if err != nil {
			t.Fatalf("Embed() error = %v", err)
		}
		if len(embeddings) != 1 {
			t.Fatalf("Embed() returned %d embeddings, want 1", len(embeddings))
		}
		if len(embeddings[0]) != 3 {
			t.Errorf("embedding dimension = %d, want 3", len(embeddings[0]))
		}
	})

	t.Run("multiple texts returns embeddings in order", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			resp := embeddingResponse{
				Data: []embeddingData{
					{Embedding: []float32{0.3, 0.4}, Index: 1},
					{Embedding: []float32{0.1, 0.2}, Index: 0},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := embedding.NewClient(server.URL, 2)
		embeddings, err := client.Embed(context.Background(), []string{"first", "second"})

		if err != nil {
			t.Fatalf("Embed() error = %v", err)
		}
		if len(embeddings) != 2 {
			t.Fatalf("Embed() returned %d embeddings, want 2", len(embeddings))
		}
		if embeddings[0][0] != 0.1 {
			t.Errorf("first embedding = %v, want [0.1, 0.2]", embeddings[0])
		}
		if embeddings[1][0] != 0.3 {
			t.Errorf("second embedding = %v, want [0.3, 0.4]", embeddings[1])
		}
	})

	t.Run("empty input returns empty slice", func(t *testing.T) {
		t.Parallel()
		client := embedding.NewClient("http://unused", 384)
		embeddings, err := client.Embed(context.Background(), []string{})

		if err != nil {
			t.Fatalf("Embed() error = %v", err)
		}
		if len(embeddings) != 0 {
			t.Errorf("Embed() returned %d embeddings, want 0", len(embeddings))
		}
	})

	t.Run("server error returns error", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := embedding.NewClient(server.URL, 384)
		_, err := client.Embed(context.Background(), []string{"test"})

		if err == nil {
			t.Error("Embed() expected error for 500 response")
		}
	})

	t.Run("model is included in request", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req embeddingRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Model != "text-embedding-ada-002" {
				t.Errorf("model = %s, want text-embedding-ada-002", req.Model)
			}

			resp := embeddingResponse{
				Data: []embeddingData{{Embedding: []float32{0.1}, Index: 0}},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := embedding.NewClient(server.URL, 1, embedding.WithModel("text-embedding-ada-002"))
		_, _ = client.Embed(context.Background(), []string{"test"})
	})
}

func TestClient_Dimensions(t *testing.T) {
	t.Parallel()

	client := embedding.NewClient("http://unused", 384)
	if client.Dimensions() != 384 {
		t.Errorf("Dimensions() = %d, want 384", client.Dimensions())
	}
}
