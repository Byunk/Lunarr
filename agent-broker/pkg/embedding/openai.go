package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is an OpenAI-compatible embeddings client.
// Works with OpenAI, TEI, Ollama, vLLM, and other compatible providers.
type Client struct {
	// url is the base URL of the embeddings API.
	url string
	// model is the model name to use for embeddings.
	model string
	// dim is the embedding vector dimension.
	dim int
	// httpClient is the HTTP client for making requests.
	httpClient *http.Client
}

// Options configures the Client.
type Options struct {
	// Model is the model name to use.
	Model string
	// HTTPClient is the HTTP client to use.
	HTTPClient *http.Client
}

// DefaultOptions returns sensible defaults.
func DefaultOptions() Options {
	return Options{
		Model: "",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Option is a functional option for Client.
type Option func(*Options)

// WithModel sets the model name.
func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = model
	}
}

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(o *Options) {
		o.HTTPClient = client
	}
}

// embeddingRequest is the request body for POST /v1/embeddings.
type embeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model,omitempty"`
}

// embeddingResponse is the response from POST /v1/embeddings.
type embeddingResponse struct {
	Data []embeddingData `json:"data"`
}

// embeddingData represents a single embedding in the response.
type embeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// NewClient creates a new OpenAI-compatible embeddings client.
func NewClient(url string, dim int, opts ...Option) *Client {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Client{
		url:        url,
		model:      options.Model,
		dim:        dim,
		httpClient: options.HTTPClient,
	}
}

// Embed generates embeddings for the given texts.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	reqBody := embeddingRequest{
		Input: texts,
		Model: c.model,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	// Sort by index and extract embeddings
	embeddings := make([][]float32, len(embResp.Data))
	for _, d := range embResp.Data {
		if d.Index < len(embeddings) {
			embeddings[d.Index] = d.Embedding
		}
	}

	return embeddings, nil
}

// Dimensions returns the embedding vector dimension.
func (c *Client) Dimensions() int {
	return c.dim
}
