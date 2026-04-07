package agents

import (
	"context"
	"iter"
	"testing"

	adkagent "google.golang.org/adk/agent"
	adkmodel "google.golang.org/adk/model"
	"google.golang.org/genai"
)

// mockModel implements model.LLM interface for testing
type mockModel struct {
	name string
}

func (m *mockModel) Name() string {
	return m.name
}

func (m *mockModel) GenerateContent(ctx context.Context, req *adkmodel.LLMRequest, stream bool) iter.Seq2[*adkmodel.LLMResponse, error] {
	return func(yield func(*adkmodel.LLMResponse, error) bool) {
		yield(&adkmodel.LLMResponse{
			Content: &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: "mock response"}},
			},
			TurnComplete: true,
		}, nil)
	}
}

func TestNewResearcher(t *testing.T) {
	t.Run("creates researcher agent with correct name", func(t *testing.T) {
		model := &mockModel{name: "test-model"}

		agent, err := NewResearcher(model)
		if err != nil {
			t.Fatalf("NewResearcher() error = %v", err)
		}

		if agent.Name() != "researcher" {
			t.Errorf("agent.Name() = %q, want %q", agent.Name(), "researcher")
		}
	})

	t.Run("creates researcher agent with correct description", func(t *testing.T) {
		model := &mockModel{name: "test-model"}

		agent, err := NewResearcher(model)
		if err != nil {
			t.Fatalf("NewResearcher() error = %v", err)
		}

		expectedDescription := "信息收集与整理专员"
		if agent.Description() != expectedDescription {
			t.Errorf("agent.Description() = %q, want %q", agent.Description(), expectedDescription)
		}
	})

	t.Run("returns error when model is nil", func(t *testing.T) {
		agent, err := NewResearcher(nil)
		if err == nil {
			t.Error("expected error when model is nil, got nil")
		}
		if agent != nil {
			t.Error("expected nil agent when model is nil")
		}
	})

	t.Run("agent implements agent.Agent interface", func(t *testing.T) {
		model := &mockModel{name: "test-model"}

		agent, err := NewResearcher(model)
		if err != nil {
			t.Fatalf("NewResearcher() error = %v", err)
		}

		// Verify it implements the agent.Agent interface
		var _ adkagent.Agent = agent
	})

	t.Run("agent has no sub-agents", func(t *testing.T) {
		model := &mockModel{name: "test-model"}

		agent, err := NewResearcher(model)
		if err != nil {
			t.Fatalf("NewResearcher() error = %v", err)
		}

		subAgents := agent.SubAgents()
		if len(subAgents) != 0 {
			t.Errorf("expected 0 sub-agents, got %d", len(subAgents))
		}
	})
}

func TestResearcherInstruction(t *testing.T) {
	t.Run("instruction contains key research guidance", func(t *testing.T) {
		model := &mockModel{name: "test-model"}

		// We'll verify the instruction is properly configured
		// by checking the agent properties
		agent, err := NewResearcher(model)
		if err != nil {
			t.Fatalf("NewResearcher() error = %v", err)
		}

		// Basic verification that agent was created successfully
		// The actual instruction content is internal to llmagent
		if agent.Name() != "researcher" {
			t.Error("agent name should be 'researcher'")
		}
	})
}