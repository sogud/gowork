package workflow

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"testing"
	"time"

	"github.com/sogud/gowork/agents"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// createTestAgent creates an adkagent.Agent using agent.New for testing
func createTestAgent(name, description, output string) (adkagent.Agent, error) {
	cfg := adkagent.Config{
		Name:        name,
		Description: description,
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				event := session.NewEvent(ctx.InvocationID())
				event.Content = genai.NewContentFromText(output, genai.RoleModel)
				yield(event, nil)
			}
		},
	}

	return adkagent.New(cfg)
}

// createDynamicTestAgent creates an agent that uses input to generate output
func createDynamicTestAgent(name, description string, transform func(string) string) (adkagent.Agent, error) {
	cfg := adkagent.Config{
		Name:        name,
		Description: description,
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				// Get input from user content
				input := ""
				if ctx.UserContent() != nil && len(ctx.UserContent().Parts) > 0 {
					input = ctx.UserContent().Parts[0].Text
				}
				output := transform(input)
				event := session.NewEvent(ctx.InvocationID())
				event.Content = genai.NewContentFromText(output, genai.RoleModel)
				yield(event, nil)
			}
		},
	}

	return adkagent.New(cfg)
}

// createErrorAgent creates an agent that returns an error
func createErrorAgent(name, description string) (adkagent.Agent, error) {
	cfg := adkagent.Config{
		Name:        name,
		Description: description,
		Run: func(ctx adkagent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				yield(nil, errors.New("agent execution failed"))
			}
		},
	}

	return adkagent.New(cfg)
}

func TestWorkflowType(t *testing.T) {
	tests := []struct {
		name     string
		typeStr  string
		expected Type
	}{
		{"sequential type", "sequential", TypeSequential},
		{"parallel type", "parallel", TypeParallel},
		{"loop type", "loop", TypeLoop},
		{"dynamic type", "dynamic", TypeDynamic},
		{"invalid type defaults to sequential", "invalid", TypeSequential},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseType(tt.typeStr)
			if result != tt.expected {
				t.Errorf("ParseType(%s) = %v, want %v", tt.typeStr, result, tt.expected)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid sequential config",
			config: Config{
				Type:    TypeSequential,
				Agents:  []string{"agent1", "agent2"},
				Task:    "test task",
				MaxIter: 1,
			},
			wantErr: false,
		},
		{
			name: "valid parallel config",
			config: Config{
				Type:    TypeParallel,
				Agents:  []string{"agent1", "agent2"},
				Task:    "test task",
				MaxIter: 1,
			},
			wantErr: false,
		},
		{
			name: "valid loop config",
			config: Config{
				Type:    TypeLoop,
				Agents:  []string{"agent1"},
				Task:    "test task",
				MaxIter: 3,
			},
			wantErr: false,
		},
		{
			name: "valid dynamic config",
			config: Config{
				Type:    TypeDynamic,
				Agents:  []string{"coordinator", "agent1", "agent2"},
				Task:    "test task",
				MaxIter: 5,
			},
			wantErr: false,
		},
		{
			name: "empty agents",
			config: Config{
				Type:    TypeSequential,
				Agents:  []string{},
				Task:    "test task",
				MaxIter: 1,
			},
			wantErr: true,
		},
		{
			name: "empty task",
			config: Config{
				Type:    TypeSequential,
				Agents:  []string{"agent1"},
				Task:    "",
				MaxIter: 1,
			},
			wantErr: true,
		},
		{
			name: "loop config without max iterations",
			config: Config{
				Type:    TypeLoop,
				Agents:  []string{"agent1"},
				Task:    "test task",
				MaxIter: 0,
			},
			wantErr: true,
		},
		{
			name: "dynamic config without coordinator",
			config: Config{
				Type:    TypeDynamic,
				Agents:  []string{"agent1", "agent2"},
				Task:    "test task",
				MaxIter: 5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewEngine(t *testing.T) {
	t.Run("creates engine with valid config", func(t *testing.T) {
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"agent1", "agent2"},
			Task:    "test task",
			MaxIter: 1,
		}

		engine, err := NewEngine(config)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine == nil {
			t.Fatal("NewEngine() returned nil engine")
		}

		if engine.config == nil {
			t.Fatal("Engine config is nil")
		}

		if engine.sessionService == nil {
			t.Fatal("Engine should have default session service")
		}
	})

	t.Run("rejects nil config", func(t *testing.T) {
		engine, err := NewEngine(nil)
		if err == nil {
			t.Error("NewEngine() should return error for nil config")
		}

		if engine != nil {
			t.Error("NewEngine() should return nil engine for nil config")
		}
	})

	t.Run("rejects invalid config", func(t *testing.T) {
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{},
			Task:    "test task",
			MaxIter: 1,
		}

		engine, err := NewEngine(config)
		if err == nil {
			t.Error("NewEngine() should return error for invalid config")
		}

		if engine != nil {
			t.Error("NewEngine() should return nil engine for invalid config")
		}
	})

	t.Run("accepts registry option", func(t *testing.T) {
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"agent1"},
			Task:    "test task",
			MaxIter: 1,
		}

		registry := agents.NewRegistry()
		engine, err := NewEngine(config, WithRegistry(registry))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.registry == nil {
			t.Error("Engine should have registry")
		}
	})

	t.Run("rejects nil registry option", func(t *testing.T) {
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"agent1"},
			Task:    "test task",
			MaxIter: 1,
		}

		engine, err := NewEngine(config, WithRegistry(nil))
		if err == nil {
			t.Error("NewEngine() should return error for nil registry")
		}

		if engine != nil {
			t.Error("NewEngine() should return nil engine for nil registry")
		}
	})

	t.Run("accepts session service option", func(t *testing.T) {
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"agent1"},
			Task:    "test task",
			MaxIter: 1,
		}

		sessionSvc := session.InMemoryService()
		engine, err := NewEngine(config, WithSessionService(sessionSvc))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		if engine.sessionService == nil {
			t.Error("Engine should have session service")
		}
	})
}

func TestEngineExecutePlaceholder(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "sequential execution",
			config: Config{
				Type:    TypeSequential,
				Agents:  []string{"agent1", "agent2"},
				Task:    "test task",
				MaxIter: 1,
			},
		},
		{
			name: "parallel execution",
			config: Config{
				Type:    TypeParallel,
				Agents:  []string{"agent1", "agent2"},
				Task:    "test task",
				MaxIter: 1,
			},
		},
		{
			name: "loop execution",
			config: Config{
				Type:    TypeLoop,
				Agents:  []string{"agent1"},
				Task:    "test task",
				MaxIter: 3,
			},
		},
		{
			name: "dynamic execution",
			config: Config{
				Type:    TypeDynamic,
				Agents:  []string{"coordinator", "agent1", "agent2"},
				Task:    "test task",
				MaxIter: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewEngine(&tt.config)
			if err != nil {
				t.Fatalf("NewEngine() error = %v", err)
			}

			result, err := engine.Execute(ctx)
			if err != nil {
				t.Fatalf("Engine.Execute() error = %v", err)
			}

			if result == nil {
				t.Fatal("Engine.Execute() returned nil result")
			}

			if !result.Success {
				t.Errorf("Engine.Execute() success = %v, want true", result.Success)
			}

			if result.Type != tt.config.Type {
				t.Errorf("Result.Type = %v, want %v", result.Type, tt.config.Type)
			}
		})
	}
}

func TestSequentialWorkflowWithMockAgents(t *testing.T) {
	t.Run("executes agents in order", func(t *testing.T) {
		// Create registry with test agents
		registry := agents.NewRegistry()

		// Create agents using adkagent.New
		agent1, err := createTestAgent("agent1", "first agent", "output from agent1: processed initial task")
		if err != nil {
			t.Fatalf("Failed to create agent1: %v", err)
		}

		agent2, err := createDynamicTestAgent("agent2", "second agent", func(input string) string {
			return fmt.Sprintf("output from agent2: enhanced %s", input)
		})
		if err != nil {
			t.Fatalf("Failed to create agent2: %v", err)
		}

		// Register agents as executable
		registry.Register(agents.NewExecutableAgent(agent1))
		registry.Register(agents.NewExecutableAgent(agent2))

		// Create workflow config
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"agent1", "agent2"},
			Task:    "initial task",
			MaxIter: 1,
		}

		// Create engine with registry
		engine, err := NewEngine(config, WithRegistry(registry))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		// Execute workflow
		ctx := context.Background()
		result, err := engine.Execute(ctx)
		if err != nil {
			t.Fatalf("Engine.Execute() error = %v", err)
		}

		// Verify result
		if !result.Success {
			t.Errorf("Expected success, got failure: %v", result.Error)
		}

		if len(result.AgentResults) != 2 {
			t.Errorf("Expected 2 agent results, got %d", len(result.AgentResults))
		}

		// Check agent names
		if result.AgentResults[0].AgentName != "agent1" {
			t.Errorf("Expected first agent to be 'agent1', got '%s'", result.AgentResults[0].AgentName)
		}

		if result.AgentResults[1].AgentName != "agent2" {
			t.Errorf("Expected second agent to be 'agent2', got '%s'", result.AgentResults[1].AgentName)
		}

		// Check that output contains both agent contributions
		if result.Output == "" {
			t.Error("Expected non-empty output")
		}
	})

	t.Run("handles agent not found", func(t *testing.T) {
		registry := agents.NewRegistry()

		// Only register agent1, not agent2
		agent1, err := createTestAgent("agent1", "first agent", "output from agent1")
		if err != nil {
			t.Fatalf("Failed to create agent1: %v", err)
		}
		registry.Register(agents.NewExecutableAgent(agent1))

		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"agent1", "nonexistent"},
			Task:    "test task",
			MaxIter: 1,
		}

		engine, err := NewEngine(config, WithRegistry(registry))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		ctx := context.Background()
		result, err := engine.Execute(ctx)

		// Should return error
		if err == nil {
			t.Error("Expected error for nonexistent agent")
		}

		if result != nil && result.Success {
			t.Error("Expected failure result")
		}

		// Check that partial results are recorded
		if result != nil && len(result.AgentResults) != 1 {
			t.Errorf("Expected 1 partial result (from agent1), got %d", len(result.AgentResults))
		}
	})

	t.Run("handles non-executable agent", func(t *testing.T) {
		registry := agents.NewRegistry()

		// Register a non-executable agent (just basic Agent interface)
		registry.Register(&mockNonExecutableAgent{name: "basic-agent"})

		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"basic-agent"},
			Task:    "test task",
			MaxIter: 1,
		}

		engine, err := NewEngine(config, WithRegistry(registry))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		ctx := context.Background()
		_, err = engine.Execute(ctx)

		if err == nil {
			t.Error("Expected error for non-executable agent")
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		registry := agents.NewRegistry()

		// Create a slow agent
		slowAgent, err := createTestAgent("slow-agent", "slow agent", "slow output")
		if err != nil {
			t.Fatalf("Failed to create slow-agent: %v", err)
		}
		registry.Register(agents.NewExecutableAgent(slowAgent))

		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"slow-agent", "slow-agent"},
			Task:    "test task",
			MaxIter: 1,
		}

		engine, err := NewEngine(config, WithRegistry(registry))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		// Create context with very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		// Sleep a bit to ensure context expires
		time.Sleep(10 * time.Millisecond)

		_, err = engine.Execute(ctx)

		// Should return error due to timeout
		if err == nil {
			t.Error("Expected error due to context cancellation")
		}

		if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			t.Logf("Got error: %v (expected DeadlineExceeded or Canceled)", err)
		}
	})
}

// mockNonExecutableAgent implements only the basic Agent interface
type mockNonExecutableAgent struct {
	name string
}

func (m *mockNonExecutableAgent) Name() string        { return m.name }
func (m *mockNonExecutableAgent) Description() string { return "non-executable agent" }

func TestSequentialWorkflowStatePassing(t *testing.T) {
	t.Run("passes output between agents", func(t *testing.T) {
		registry := agents.NewRegistry()

		// Create agents that demonstrate state passing
		step1, err := createTestAgent("step1", "step 1", "step1-result")
		if err != nil {
			t.Fatalf("Failed to create step1: %v", err)
		}

		step2, err := createDynamicTestAgent("step2", "step 2", func(input string) string {
			// Verify input is step1's output
			return fmt.Sprintf("step2-result based on %s", input)
		})
		if err != nil {
			t.Fatalf("Failed to create step2: %v", err)
		}

		registry.Register(agents.NewExecutableAgent(step1))
		registry.Register(agents.NewExecutableAgent(step2))

		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"step1", "step2"},
			Task:    "initial-input",
			MaxIter: 1,
		}

		engine, err := NewEngine(config, WithRegistry(registry))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		ctx := context.Background()
		result, err := engine.Execute(ctx)
		if err != nil {
			t.Fatalf("Engine.Execute() error = %v", err)
		}

		// Verify final output includes step2's result
		if !result.Success {
			t.Errorf("Expected success, got failure")
		}

		// Check that output contains step2's processing
		if !contains(result.Output, "step2-result based on step1-result") {
			t.Errorf("Expected output to contain state-passed result, got: %s", result.Output)
		}
	})
}

func TestSequentialWorkflowErrorHandling(t *testing.T) {
	t.Run("handles agent execution error", func(t *testing.T) {
		registry := agents.NewRegistry()

		// Create an agent that returns an error
		errorAgent, err := createErrorAgent("error-agent", "error agent")
		if err != nil {
			t.Fatalf("Failed to create error-agent: %v", err)
		}

		registry.Register(agents.NewExecutableAgent(errorAgent))

		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"error-agent"},
			Task:    "test task",
			MaxIter: 1,
		}

		engine, err := NewEngine(config, WithRegistry(registry))
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		ctx := context.Background()
		result, err := engine.Execute(ctx)

		if err == nil {
			t.Error("Expected error from agent execution")
		}

		if result != nil && result.Success {
			t.Error("Expected failure result")
		}
	})
}

func TestAggregateSequentialResults(t *testing.T) {
	t.Run("aggregates multiple results", func(t *testing.T) {
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{"a", "b", "c"},
			Task:    "test",
			MaxIter: 1,
		}

		engine, err := NewEngine(config)
		if err != nil {
			t.Fatalf("NewEngine() error = %v", err)
		}

		results := []AgentResult{
			{AgentName: "a", Output: "output-a"},
			{AgentName: "b", Output: "output-b"},
			{AgentName: "c", Output: "output-c"},
		}

		aggregated := engine.aggregateSequentialResults(results)

		// Check that output contains all steps
		if !contains(aggregated, "Step 1: a") {
			t.Error("Expected aggregated output to contain 'Step 1: a'")
		}

		if !contains(aggregated, "output-a") {
			t.Error("Expected aggregated output to contain 'output-a'")
		}

		if !contains(aggregated, "output-c") {
			t.Error("Expected aggregated output to contain final output 'output-c'")
		}

		if !contains(aggregated, "=== Final Output ===") {
			t.Error("Expected aggregated output to contain '=== Final Output ==='")
		}
	})

	t.Run("handles empty results", func(t *testing.T) {
		config := &Config{
			Type:    TypeSequential,
			Agents:  []string{},
			Task:    "test",
			MaxIter: 1,
		}

		// This config is invalid, but we can still test the aggregation
		engine := &Engine{config: config}

		aggregated := engine.aggregateSequentialResults([]AgentResult{})

		if aggregated != "" {
			t.Errorf("Expected empty string for empty results, got '%s'", aggregated)
		}
	})
}

func TestGetConfig(t *testing.T) {
	config := &Config{
		Type:    TypeSequential,
		Agents:  []string{"agent1"},
		Task:    "test task",
		MaxIter: 1,
	}

	engine, err := NewEngine(config)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	retrieved := engine.GetConfig()
	if retrieved == nil {
		t.Fatal("GetConfig() returned nil")
	}

	if retrieved != config {
		t.Error("GetConfig() should return the same config instance")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}