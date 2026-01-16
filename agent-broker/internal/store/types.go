package store

import (
	"time"

	"github.com/a2aproject/a2a-go/a2a"
)

// RegisteredAgent holds an agent registration with broker-internal metadata.
type RegisteredAgent struct {
	// ID is the unique identifier for the agent in the registry.
	ID string
	// Card is the A2A-compliant agent card.
	Card a2a.AgentCard
	// Tags are classification tags for filtering.
	Tags []string
	// Embedding is the vector representation for semantic search.
	Embedding []float32
	// CreatedAt is when the agent was registered.
	CreatedAt time.Time
	// UpdatedAt is when the agent was last updated.
	UpdatedAt time.Time
}
