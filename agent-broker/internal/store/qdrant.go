package store

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

// Options configures the QdrantStore.
type Options struct {
	// Host is the Qdrant server hostname.
	Host string
	// Port is the Qdrant gRPC port.
	Port int
	// APIKey is the optional API key for authentication.
	APIKey string
	// UseTLS enables TLS for the connection.
	UseTLS bool
}

// DefaultOptions returns Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Host:   "localhost",
		Port:   6334,
		APIKey: "",
		UseTLS: false,
	}
}

// Option is a functional option for configuring QdrantStore.
type Option func(*Options)

// WithHost sets the Qdrant server hostname.
func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

// WithPort sets the Qdrant gRPC port.
func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// WithAPIKey sets the API key for authentication.
func WithAPIKey(apiKey string) Option {
	return func(o *Options) {
		o.APIKey = apiKey
	}
}

// WithTLS enables TLS for the connection.
func WithTLS(useTLS bool) Option {
	return func(o *Options) {
		o.UseTLS = useTLS
	}
}

// QdrantStore implements Store using Qdrant as the vector database.
type QdrantStore struct {
	// client is the Qdrant gRPC client.
	client *qdrant.Client
}

// NewQdrantStore creates a QdrantStore with the given options.
func NewQdrantStore(ctx context.Context, opts ...Option) (*QdrantStore, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   options.Host,
		Port:   options.Port,
		APIKey: options.APIKey,
		UseTLS: options.UseTLS,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	store := &QdrantStore{
		client: client,
	}

	if err := store.Ping(ctx); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("failed to connect to qdrant: %w", err)
	}

	return store, nil
}

// Ping checks if Qdrant is reachable and healthy.
func (s *QdrantStore) Ping(ctx context.Context) error {
	_, err := s.client.HealthCheck(ctx)
	if err != nil {
		return fmt.Errorf("qdrant health check failed: %w", err)
	}
	return nil
}

// Close closes the Qdrant client connection.
func (s *QdrantStore) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
