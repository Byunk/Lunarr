package embedding

import "context"

// Embedder generates vector embeddings from text.
type Embedder interface {
	// Embed generates embeddings for the given texts.
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	// Dimensions returns the embedding vector dimension.
	Dimensions() int
}
