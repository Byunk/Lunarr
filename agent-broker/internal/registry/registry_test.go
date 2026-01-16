package registry

import (
	"context"
	"strings"
	"testing"

	"github.com/a2aproject/a2a-go/a2a"

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

func validCreateInput() CreateInput {
	return CreateInput{
		ID:   "test-agent",
		Card: validAgentCard(),
		Tags: []string{"test"},
	}
}

func TestValidateAgentCard(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		card    a2a.AgentCard
		wantErr string
	}{
		{
			name:    "valid card",
			card:    validAgentCard(),
			wantErr: "",
		},
		{
			name: "missing name",
			card: func() a2a.AgentCard {
				c := validAgentCard()
				c.Name = ""
				return c
			}(),
			wantErr: "name is required",
		},
		{
			name: "missing URL",
			card: func() a2a.AgentCard {
				c := validAgentCard()
				c.URL = ""
				return c
			}(),
			wantErr: "url is required",
		},
		{
			name: "missing version",
			card: func() a2a.AgentCard {
				c := validAgentCard()
				c.Version = ""
				return c
			}(),
			wantErr: "version is required",
		},
		{
			name: "no skills",
			card: func() a2a.AgentCard {
				c := validAgentCard()
				c.Skills = nil
				return c
			}(),
			wantErr: "at least one skill is required",
		},
		{
			name: "skill missing ID",
			card: func() a2a.AgentCard {
				c := validAgentCard()
				c.Skills = []a2a.AgentSkill{{ID: "", Name: "Test"}}
				return c
			}(),
			wantErr: "skill[0].id is required",
		},
		{
			name: "skill missing name",
			card: func() a2a.AgentCard {
				c := validAgentCard()
				c.Skills = []a2a.AgentSkill{{ID: "test", Name: ""}}
				return c
			}(),
			wantErr: "skill[0].name is required",
		},
		{
			name:    "multiple errors",
			card:    a2a.AgentCard{},
			wantErr: "name is required, url is required, version is required, at least one skill is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateAgentCard(tt.card)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateAgentCard() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Errorf("ValidateAgentCard() error = nil, want error containing %q", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("ValidateAgentCard() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateInput
		wantErr string
	}{
		{
			name:    "valid input creates agent",
			input:   validCreateInput(),
			wantErr: "",
		},
		{
			name: "empty ID rejected",
			input: func() CreateInput {
				i := validCreateInput()
				i.ID = ""
				return i
			}(),
			wantErr: "agent_id is required",
		},
		{
			name: "too long ID rejected",
			input: func() CreateInput {
				i := validCreateInput()
				i.ID = strings.Repeat("a", 65)
				return i
			}(),
			wantErr: "agent_id must be at most 64 characters",
		},
		{
			name: "invalid chars in ID rejected",
			input: func() CreateInput {
				i := validCreateInput()
				i.ID = "agent@invalid"
				return i
			}(),
			wantErr: "agent_id must match pattern",
		},
		{
			name: "invalid card rejected",
			input: func() CreateInput {
				i := validCreateInput()
				i.Card.Name = ""
				return i
			}(),
			wantErr: "name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := store.NewMemoryStore()
			svc := NewRegistryService(s)

			agent, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Create() error = %v, want nil", err)
					return
				}
				if agent.ID != tt.input.ID {
					t.Errorf("Create() ID = %v, want %v", agent.ID, tt.input.ID)
				}
				if agent.CreatedAt.IsZero() {
					t.Error("Create() CreatedAt should be set")
				}
				if agent.UpdatedAt.IsZero() {
					t.Error("Create() UpdatedAt should be set")
				}
				return
			}
			if err == nil {
				t.Errorf("Create() error = nil, want error containing %q", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Create() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRegistryService_Create_Duplicate(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)
	input := validCreateInput()

	_, err := svc.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = svc.Create(context.Background(), input)
	if err != store.ErrAlreadyExists {
		t.Errorf("second Create() error = %v, want ErrAlreadyExists", err)
	}
}

func TestRegistryService_Get(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)
	input := validCreateInput()

	created, _ := svc.Create(context.Background(), input)

	agent, err := svc.Get(context.Background(), input.ID)
	if err != nil {
		t.Errorf("Get() error = %v", err)
		return
	}
	if agent.ID != created.ID {
		t.Errorf("Get() ID = %v, want %v", agent.ID, created.ID)
	}
}

func TestRegistryService_Get_NotFound(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)

	_, err := svc.Get(context.Background(), "not-exists")
	if err != store.ErrNotFound {
		t.Errorf("Get() error = %v, want ErrNotFound", err)
	}
}

func TestRegistryService_Update(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)
	input := validCreateInput()

	created, _ := svc.Create(context.Background(), input)
	originalCreatedAt := created.CreatedAt

	updateInput := UpdateInput{
		ID:   input.ID,
		Card: validAgentCard(),
		Tags: []string{"updated"},
	}
	updateInput.Card.Name = "Updated Name"

	updated, err := svc.Update(context.Background(), updateInput)
	if err != nil {
		t.Errorf("Update() error = %v", err)
		return
	}
	if updated.Card.Name != "Updated Name" {
		t.Errorf("Update() Name = %v, want %v", updated.Card.Name, "Updated Name")
	}
	if updated.Tags[0] != "updated" {
		t.Errorf("Update() Tags = %v, want [updated]", updated.Tags)
	}
	if !updated.CreatedAt.Equal(originalCreatedAt) {
		t.Error("Update() should preserve CreatedAt")
	}
	if !updated.UpdatedAt.After(originalCreatedAt) {
		t.Error("Update() should update UpdatedAt")
	}
}

func TestRegistryService_Update_NotFound(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)

	_, err := svc.Update(context.Background(), UpdateInput{
		ID:   "not-exists",
		Card: validAgentCard(),
	})
	if err != store.ErrNotFound {
		t.Errorf("Update() error = %v, want ErrNotFound", err)
	}
}

func TestRegistryService_Update_InvalidCard(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)
	input := validCreateInput()

	_, _ = svc.Create(context.Background(), input)

	invalidCard := validAgentCard()
	invalidCard.Name = ""

	_, err := svc.Update(context.Background(), UpdateInput{
		ID:   input.ID,
		Card: invalidCard,
	})
	if err == nil {
		t.Error("Update() with invalid card should return error")
	}
}

func TestRegistryService_Delete(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)
	input := validCreateInput()

	_, _ = svc.Create(context.Background(), input)

	err := svc.Delete(context.Background(), input.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	_, err = svc.Get(context.Background(), input.ID)
	if err != store.ErrNotFound {
		t.Errorf("Get() after Delete() should return ErrNotFound, got %v", err)
	}
}

func TestRegistryService_Delete_NotFound(t *testing.T) {
	t.Parallel()
	s := store.NewMemoryStore()
	svc := NewRegistryService(s)

	err := svc.Delete(context.Background(), "not-exists")
	if err != store.ErrNotFound {
		t.Errorf("Delete() error = %v, want ErrNotFound", err)
	}
}
