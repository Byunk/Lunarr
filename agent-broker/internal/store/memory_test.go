package store

import (
	"context"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/a2a"
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

func validAgent(id string) *RegisteredAgent {
	now := time.Now()
	return &RegisteredAgent{
		ID:        id,
		Card:      validAgentCard(),
		Tags:      []string{"test"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestMemoryStore_CreateAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*MemoryStore)
		agent   *RegisteredAgent
		wantErr error
	}{
		{
			name:    "creates new agent",
			setup:   func(_ *MemoryStore) {},
			agent:   validAgent("agent-1"),
			wantErr: nil,
		},
		{
			name: "duplicate ID returns ErrAlreadyExists",
			setup: func(s *MemoryStore) {
				_ = s.CreateAgent(context.Background(), validAgent("agent-1"))
			},
			agent:   validAgent("agent-1"),
			wantErr: ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewMemoryStore()
			tt.setup(s)

			err := s.CreateAgent(context.Background(), tt.agent)

			if err != tt.wantErr {
				t.Errorf("CreateAgent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStore_GetAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*MemoryStore)
		id      string
		wantErr error
	}{
		{
			name: "returns existing agent",
			setup: func(s *MemoryStore) {
				_ = s.CreateAgent(context.Background(), validAgent("agent-1"))
			},
			id:      "agent-1",
			wantErr: nil,
		},
		{
			name:    "non-existent returns ErrNotFound",
			setup:   func(_ *MemoryStore) {},
			id:      "not-exists",
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewMemoryStore()
			tt.setup(s)

			agent, err := s.GetAgent(context.Background(), tt.id)

			if err != tt.wantErr {
				t.Errorf("GetAgent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && agent.ID != tt.id {
				t.Errorf("GetAgent() got ID = %v, want %v", agent.ID, tt.id)
			}
		})
	}
}

func TestMemoryStore_ListAgents(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("empty store returns zero agents", func(t *testing.T) {
		t.Parallel()
		s := NewMemoryStore()

		result, err := s.ListAgents(ctx, AgentFilter{Limit: 10})

		if err != nil {
			t.Fatalf("ListAgents() error = %v", err)
		}
		if len(result.Agents) != 0 {
			t.Errorf("ListAgents() got %d agents, want 0", len(result.Agents))
		}
		if result.Total != 0 {
			t.Errorf("ListAgents() total = %d, want 0", result.Total)
		}
	})

	t.Run("pagination offset and limit", func(t *testing.T) {
		t.Parallel()
		s := NewMemoryStore()
		for i := 0; i < 5; i++ {
			agent := validAgent("agent-" + string(rune('a'+i)))
			agent.CreatedAt = time.Now().Add(time.Duration(i) * time.Second)
			_ = s.CreateAgent(ctx, agent)
		}

		result, err := s.ListAgents(ctx, AgentFilter{Offset: 1, Limit: 2})

		if err != nil {
			t.Fatalf("ListAgents() error = %v", err)
		}
		if len(result.Agents) != 2 {
			t.Errorf("ListAgents() got %d agents, want 2", len(result.Agents))
		}
		if result.Total != 5 {
			t.Errorf("ListAgents() total = %d, want 5", result.Total)
		}
	})

	t.Run("filter by tags", func(t *testing.T) {
		t.Parallel()
		s := NewMemoryStore()
		agent1 := validAgent("agent-1")
		agent1.Tags = []string{"prod", "ml"}
		agent2 := validAgent("agent-2")
		agent2.Tags = []string{"dev"}
		_ = s.CreateAgent(ctx, agent1)
		_ = s.CreateAgent(ctx, agent2)

		result, err := s.ListAgents(ctx, AgentFilter{Tags: []string{"prod"}, Limit: 10})

		if err != nil {
			t.Fatalf("ListAgents() error = %v", err)
		}
		if len(result.Agents) != 1 {
			t.Errorf("ListAgents() got %d agents, want 1", len(result.Agents))
		}
		if result.Agents[0].ID != "agent-1" {
			t.Errorf("ListAgents() got ID = %v, want agent-1", result.Agents[0].ID)
		}
	})

	t.Run("filter by skills", func(t *testing.T) {
		t.Parallel()
		s := NewMemoryStore()
		agent1 := validAgent("agent-1")
		agent1.Card.Skills = []a2a.AgentSkill{{ID: "translate", Name: "Translate"}}
		agent2 := validAgent("agent-2")
		agent2.Card.Skills = []a2a.AgentSkill{{ID: "summarize", Name: "Summarize"}}
		_ = s.CreateAgent(ctx, agent1)
		_ = s.CreateAgent(ctx, agent2)

		result, err := s.ListAgents(ctx, AgentFilter{Skills: []string{"translate"}, Limit: 10})

		if err != nil {
			t.Fatalf("ListAgents() error = %v", err)
		}
		if len(result.Agents) != 1 {
			t.Errorf("ListAgents() got %d agents, want 1", len(result.Agents))
		}
		if result.Agents[0].ID != "agent-1" {
			t.Errorf("ListAgents() got ID = %v, want agent-1", result.Agents[0].ID)
		}
	})

	t.Run("text search case-insensitive", func(t *testing.T) {
		t.Parallel()
		s := NewMemoryStore()
		agent1 := validAgent("agent-1")
		agent1.Card.Name = "Translation Agent"
		agent2 := validAgent("agent-2")
		agent2.Card.Name = "Summarizer"
		_ = s.CreateAgent(ctx, agent1)
		_ = s.CreateAgent(ctx, agent2)

		result, err := s.ListAgents(ctx, AgentFilter{Query: "translation", Limit: 10})

		if err != nil {
			t.Fatalf("ListAgents() error = %v", err)
		}
		if len(result.Agents) != 1 {
			t.Errorf("ListAgents() got %d agents, want 1", len(result.Agents))
		}
	})

	t.Run("sorted by CreatedAt descending", func(t *testing.T) {
		t.Parallel()
		s := NewMemoryStore()
		agent1 := validAgent("agent-old")
		agent1.CreatedAt = time.Now().Add(-time.Hour)
		agent2 := validAgent("agent-new")
		agent2.CreatedAt = time.Now()
		_ = s.CreateAgent(ctx, agent1)
		_ = s.CreateAgent(ctx, agent2)

		result, err := s.ListAgents(ctx, AgentFilter{Limit: 10})

		if err != nil {
			t.Fatalf("ListAgents() error = %v", err)
		}
		if len(result.Agents) != 2 {
			t.Fatalf("ListAgents() got %d agents, want 2", len(result.Agents))
		}
		if result.Agents[0].ID != "agent-new" {
			t.Errorf("ListAgents() first agent = %v, want agent-new", result.Agents[0].ID)
		}
	})
}

func TestMemoryStore_UpdateAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*MemoryStore)
		agent   *RegisteredAgent
		wantErr error
	}{
		{
			name: "updates existing agent",
			setup: func(s *MemoryStore) {
				_ = s.CreateAgent(context.Background(), validAgent("agent-1"))
			},
			agent: func() *RegisteredAgent {
				a := validAgent("agent-1")
				a.Card.Name = "Updated Name"
				return a
			}(),
			wantErr: nil,
		},
		{
			name:    "non-existent returns ErrNotFound",
			setup:   func(_ *MemoryStore) {},
			agent:   validAgent("not-exists"),
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewMemoryStore()
			tt.setup(s)

			err := s.UpdateAgent(context.Background(), tt.agent)

			if err != tt.wantErr {
				t.Errorf("UpdateAgent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemoryStore_DeleteAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*MemoryStore)
		id      string
		wantErr error
	}{
		{
			name: "deletes existing agent",
			setup: func(s *MemoryStore) {
				_ = s.CreateAgent(context.Background(), validAgent("agent-1"))
			},
			id:      "agent-1",
			wantErr: nil,
		},
		{
			name:    "non-existent returns ErrNotFound",
			setup:   func(_ *MemoryStore) {},
			id:      "not-exists",
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewMemoryStore()
			tt.setup(s)

			err := s.DeleteAgent(context.Background(), tt.id)

			if err != tt.wantErr {
				t.Errorf("DeleteAgent() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				_, err := s.GetAgent(context.Background(), tt.id)
				if err != ErrNotFound {
					t.Errorf("GetAgent() after delete should return ErrNotFound, got %v", err)
				}
			}
		})
	}
}
