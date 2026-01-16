package store

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a requested agent does not exist.
var ErrNotFound = errors.New("agent not found")

// ErrAlreadyExists is returned when creating a duplicate agent.
var ErrAlreadyExists = errors.New("agent already exists")

// Store defines the interface for agent storage operations.
type Store interface {
	// Ping checks if the storage backend is reachable.
	Ping(ctx context.Context) error
	// Close releases resources.
	Close() error
	// CreateAgent stores a new agent. Returns ErrAlreadyExists if ID exists.
	CreateAgent(ctx context.Context, agent *RegisteredAgent) error
	// GetAgent retrieves an agent by ID. Returns ErrNotFound if not exists.
	GetAgent(ctx context.Context, id string) (*RegisteredAgent, error)
	// ListAgents returns agents matching the filter criteria.
	ListAgents(ctx context.Context, filter AgentFilter) (*AgentListResult, error)
	// UpdateAgent updates an existing agent. Returns ErrNotFound if not exists.
	UpdateAgent(ctx context.Context, agent *RegisteredAgent) error
	// DeleteAgent removes an agent. Returns ErrNotFound if not exists.
	DeleteAgent(ctx context.Context, id string) error
}

// HealthChecker provides health check capability for storage backends.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// AgentFilter specifies criteria for listing agents.
type AgentFilter struct {
	// Offset is the number of items to skip.
	Offset int
	// Limit is the maximum number of items to return.
	Limit int
	// Tags filters by any matching tag.
	Tags []string
	// Skills filters by any matching skill ID.
	Skills []string
	// Query is a text search in name/description.
	Query string
}

// AgentListResult contains the list result with pagination info.
type AgentListResult struct {
	// Agents is the list of matching agents.
	Agents []*RegisteredAgent
	// Total is the total count before pagination.
	Total int
}
