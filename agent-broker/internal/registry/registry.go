package registry

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/a2aproject/a2a-go/a2a"

	"github.com/lunarr-ai/lunarr/agent-broker/internal/store"
)

var agentIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// RegistryService manages agent registrations.
type RegistryService struct {
	// store is the agent storage backend.
	store store.Store
}

// NewRegistryService creates a new registry service.
func NewRegistryService(s store.Store) *RegistryService {
	return &RegistryService{
		store: s,
	}
}

// CreateInput contains input for creating an agent.
type CreateInput struct {
	// ID is the unique agent identifier.
	ID string
	// Card is the A2A agent card.
	Card a2a.AgentCard
	// Tags are classification tags.
	Tags []string
}

// Create registers a new agent.
func (s *RegistryService) Create(ctx context.Context, input CreateInput) (*store.RegisteredAgent, error) {
	if err := validateAgentID(input.ID); err != nil {
		return nil, err
	}
	if err := ValidateAgentCard(input.Card); err != nil {
		return nil, err
	}

	now := time.Now()
	agent := &store.RegisteredAgent{
		ID:        input.ID,
		Card:      input.Card,
		Tags:      input.Tags,
		Embedding: nil,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.store.CreateAgent(ctx, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

// Get retrieves an agent by ID.
func (s *RegistryService) Get(ctx context.Context, id string) (*store.RegisteredAgent, error) {
	return s.store.GetAgent(ctx, id)
}

// ListInput contains input for listing agents.
type ListInput struct {
	// Offset is the number of items to skip.
	Offset int
	// Limit is the maximum items to return.
	Limit int
	// Tags filters by any matching tag.
	Tags []string
	// Skills filters by any matching skill ID.
	Skills []string
	// Query searches name/description.
	Query string
}

// List returns agents matching the criteria.
func (s *RegistryService) List(ctx context.Context, input ListInput) (*store.AgentListResult, error) {
	if input.Limit <= 0 {
		input.Limit = 20
	}
	if input.Limit > 100 {
		input.Limit = 100
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	return s.store.ListAgents(ctx, store.AgentFilter{
		Offset: input.Offset,
		Limit:  input.Limit,
		Tags:   input.Tags,
		Skills: input.Skills,
		Query:  input.Query,
	})
}

// UpdateInput contains input for updating an agent.
type UpdateInput struct {
	// ID is the agent identifier.
	ID string
	// Card is the updated A2A agent card.
	Card a2a.AgentCard
	// Tags are the updated classification tags.
	Tags []string
}

// Update modifies an existing agent.
func (s *RegistryService) Update(ctx context.Context, input UpdateInput) (*store.RegisteredAgent, error) {
	if err := ValidateAgentCard(input.Card); err != nil {
		return nil, err
	}

	existing, err := s.store.GetAgent(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	existing.Card = input.Card
	existing.Tags = input.Tags
	existing.Embedding = nil
	existing.UpdatedAt = time.Now()

	if err := s.store.UpdateAgent(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// Delete removes an agent.
func (s *RegistryService) Delete(ctx context.Context, id string) error {
	return s.store.DeleteAgent(ctx, id)
}

// ValidateAgentCard validates required fields in an AgentCard.
func ValidateAgentCard(card a2a.AgentCard) error {
	var errs []string

	if card.Name == "" {
		errs = append(errs, "name is required")
	}
	if card.URL == "" {
		errs = append(errs, "url is required")
	}
	if card.Version == "" {
		errs = append(errs, "version is required")
	}
	if len(card.Skills) == 0 {
		errs = append(errs, "at least one skill is required")
	}

	for i, skill := range card.Skills {
		if skill.ID == "" {
			errs = append(errs, fmt.Sprintf("skill[%d].id is required", i))
		}
		if skill.Name == "" {
			errs = append(errs, fmt.Sprintf("skill[%d].name is required", i))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid agent card: %s", strings.Join(errs, ", "))
	}
	return nil
}

func validateAgentID(id string) error {
	if id == "" {
		return fmt.Errorf("agent_id is required")
	}
	if len(id) > 64 {
		return fmt.Errorf("agent_id must be at most 64 characters")
	}
	if !agentIDPattern.MatchString(id) {
		return fmt.Errorf("agent_id must match pattern ^[a-zA-Z0-9_-]+$")
	}
	return nil
}
