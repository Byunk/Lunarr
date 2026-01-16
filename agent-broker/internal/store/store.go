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
	Ping(ctx context.Context) error
	Close() error
}

// HealthChecker provides health check capability for storage backends.
type HealthChecker interface {
	Ping(ctx context.Context) error
}
