package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sogud/gowork/agents"
	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// Engine orchestrates workflow execution across multiple agents.
// It manages agent execution patterns, state sharing, and result aggregation.
type Engine struct {
	config         *Config
	registry       *agents.Registry
	sessionService session.Service
}

// EngineOption is a function that configures the Engine.
type EngineOption func(*Engine) error

// WithRegistry sets the agent registry for the engine.
func WithRegistry(registry *agents.Registry) EngineOption {
	return func(e *Engine) error {
		if registry == nil {
			return errors.New("registry cannot be nil")
		}
		e.registry = registry
		return nil
	}
}

// WithSessionService sets the session service for the engine.
func WithSessionService(service session.Service) EngineOption {
	return func(e *Engine) error {
		if service == nil {
			return errors.New("session service cannot be nil")
		}
		e.sessionService = service
		return nil
	}
}

// NewEngine creates a new workflow engine with the given configuration.
//
// Parameters:
//   - config: Workflow configuration. Must not be nil and must be valid.
//   - opts: Optional engine configuration options.
//
// Returns:
//   - *Engine: The created engine
//   - error: An error if config is nil or invalid
func NewEngine(config *Config, opts ...EngineOption) (*Engine, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	engine := &Engine{
		config:         config,
		sessionService: session.InMemoryService(), // Default to in-memory
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(engine); err != nil {
			return nil, fmt.Errorf("engine option error: %w", err)
		}
	}

	return engine, nil
}

// Execute runs the workflow according to its type.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//
// Returns:
//   - *Result: Workflow execution result
//   - error: An error if execution fails
func (e *Engine) Execute(ctx context.Context) (*Result, error) {
	switch e.config.Type {
	case TypeSequential:
		return e.executeSequential(ctx)
	case TypeParallel:
		return e.executeParallel(ctx)
	case TypeLoop:
		return e.executeLoop(ctx)
	case TypeDynamic:
		return e.executeDynamic(ctx)
	default:
		return nil, fmt.Errorf("unknown workflow type: %s", e.config.Type)
	}
}

// executeSequential runs agents in sequential order with state passing.
// Each agent receives the output from the previous agent as context.
func (e *Engine) executeSequential(ctx context.Context) (*Result, error) {
	// If no registry, use placeholder execution
	if e.registry == nil {
		return e.executeSequentialPlaceholder(ctx)
	}

	agentResults := make([]AgentResult, 0, len(e.config.Agents))
	currentInput := e.config.Task

	for i, agentName := range e.config.Agents {
		// Check context cancellation
		if ctx.Err() != nil {
			return &Result{
				Success:      false,
				Type:         TypeSequential,
				Output:       "",
				AgentResults: agentResults,
				Iterations:   i,
				Error:        ctx.Err(),
			}, ctx.Err()
		}

		// Get agent from registry
		agent, err := e.registry.Get(agentName)
		if err != nil {
			return &Result{
				Success:      false,
				Type:         TypeSequential,
				Output:       "",
				AgentResults: agentResults,
				Iterations:   i,
				Error:        fmt.Errorf("agent '%s' not found: %w", agentName, err),
			}, err
		}

		// Check if agent is executable
		execAgent, ok := agent.(agents.ExecutableAgent)
		if !ok {
			return &Result{
				Success:      false,
				Type:         TypeSequential,
				Output:       "",
				AgentResults: agentResults,
				Iterations:   i,
				Error:        fmt.Errorf("agent '%s' is not executable", agentName),
			}, fmt.Errorf("agent '%s' is not executable", agentName)
		}

		// Execute agent
		output, err := e.runAgent(ctx, execAgent.ADKAgent(), currentInput, agentName)
		if err != nil {
			agentResults = append(agentResults, AgentResult{
				AgentName: agentName,
				Output:    "",
				Error:     err,
			})
			return &Result{
				Success:      false,
				Type:         TypeSequential,
				Output:       "",
				AgentResults: agentResults,
				Iterations:   i + 1,
				Error:        fmt.Errorf("agent '%s' execution failed: %w", agentName, err),
			}, err
		}

		// Record result
		agentResults = append(agentResults, AgentResult{
			AgentName: agentName,
			Output:    output,
			Error:     nil,
		})

		// Pass output to next agent
		currentInput = output
	}

	// Aggregate final output
	finalOutput := e.aggregateSequentialResults(agentResults)

	return &Result{
		Success:      true,
		Type:         TypeSequential,
		Output:       finalOutput,
		AgentResults: agentResults,
		Iterations:   len(e.config.Agents),
	}, nil
}

// runAgent executes a single agent using adk-go's runner.
func (e *Engine) runAgent(ctx context.Context, agent adkagent.Agent, input string, agentName string) (string, error) {
	// Create runner for this agent
	runnerCfg := runner.Config{
		AppName:           "workflow-engine",
		Agent:             agent,
		SessionService:    e.sessionService,
		AutoCreateSession: true,
	}

	adkRunner, err := runner.New(runnerCfg)
	if err != nil {
		return "", fmt.Errorf("failed to create runner: %w", err)
	}

	// Create user message
	userContent := genai.NewContentFromText(input, genai.RoleUser)

	// Run the agent and collect output
	var outputBuilder strings.Builder
	userID := "workflow-user"
	sessionID := fmt.Sprintf("session-%s-%s", e.config.Type, agentName)
	runConfig := adkagent.RunConfig{} // Default config

	for event, err := range adkRunner.Run(ctx, userID, sessionID, userContent, runConfig) {
		if err != nil {
			return "", fmt.Errorf("agent run error: %w", err)
		}

		// Extract text from event
		if event.IsFinalResponse() {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					outputBuilder.WriteString(part.Text)
				}
			}
		}
	}

	return outputBuilder.String(), nil
}

// aggregateSequentialResults combines all agent outputs into a final result.
func (e *Engine) aggregateSequentialResults(results []AgentResult) string {
	if len(results) == 0 {
		return ""
	}

	// For sequential workflow, the last agent's output is the final output
	// But we also include a summary of all intermediate results
	var builder strings.Builder

	builder.WriteString("=== Sequential Workflow Results ===\n\n")

	for i, result := range results {
		builder.WriteString(fmt.Sprintf("Step %d: %s\n", i+1, result.AgentName))
		builder.WriteString(fmt.Sprintf("Output: %s\n\n", result.Output))
	}

	builder.WriteString("=== Final Output ===\n")
	builder.WriteString(results[len(results)-1].Output)

	return builder.String()
}

// executeSequentialPlaceholder runs placeholder sequential execution when no registry is available.
// This is used for testing and basic validation.
func (e *Engine) executeSequentialPlaceholder(ctx context.Context) (*Result, error) {
	agentResults := make([]AgentResult, len(e.config.Agents))
	for i, agentName := range e.config.Agents {
		// Check context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		// Placeholder: create dummy result
		agentResults[i] = AgentResult{
			AgentName: agentName,
			Output:    fmt.Sprintf("placeholder output from %s", agentName),
		}
	}

	return &Result{
		Success:      true,
		Type:         TypeSequential,
		Output:       "placeholder: sequential workflow completed",
		AgentResults: agentResults,
		Iterations:   1,
	}, nil
}

// executeParallel runs agents concurrently.
// Placeholder implementation for Task 7.
func (e *Engine) executeParallel(ctx context.Context) (*Result, error) {
	// TODO: Task 8 - Implement real parallel execution
	// 1. Launch all agents concurrently
	// 2. Wait for all to complete
	// 3. Aggregate results

	agentResults := make([]AgentResult, len(e.config.Agents))
	for i, agentName := range e.config.Agents {
		// Placeholder: create dummy result
		agentResults[i] = AgentResult{
			AgentName: agentName,
			Output:    fmt.Sprintf("placeholder output from %s", agentName),
		}
	}

	return &Result{
		Success:      true,
		Type:         TypeParallel,
		Output:       "placeholder: parallel workflow completed",
		AgentResults: agentResults,
		Iterations:   1,
	}, nil
}

// executeLoop runs agents iteratively for refinement.
// Placeholder implementation for Task 7.
func (e *Engine) executeLoop(ctx context.Context) (*Result, error) {
	// TODO: Task 8 - Implement real loop execution
	// 1. Execute agent(s)
	// 2. Check if result meets criteria
	// 3. If not and iterations < MaxIter, refine and repeat
	// 4. Return final result

	agentResults := make([]AgentResult, len(e.config.Agents))
	for i, agentName := range e.config.Agents {
		// Placeholder: create dummy result
		agentResults[i] = AgentResult{
			AgentName: agentName,
			Output:    fmt.Sprintf("placeholder output from %s (iteration %d)", agentName, e.config.MaxIter),
		}
	}

	return &Result{
		Success:      true,
		Type:         TypeLoop,
		Output:       "placeholder: loop workflow completed",
		AgentResults: agentResults,
		Iterations:   e.config.MaxIter,
	}, nil
}

// executeDynamic lets coordinator decide execution pattern.
// Placeholder implementation for Task 7.
func (e *Engine) executeDynamic(ctx context.Context) (*Result, error) {
	// TODO: Task 8 - Implement real dynamic execution
	// 1. Coordinator analyzes task
	// 2. Coordinator decides which agents to use and in what order
	// 3. Execute according to coordinator's plan
	// 4. Return aggregated results

	agentResults := make([]AgentResult, len(e.config.Agents))
	for i, agentName := range e.config.Agents {
		// Placeholder: create dummy result
		agentResults[i] = AgentResult{
			AgentName: agentName,
			Output:    fmt.Sprintf("placeholder output from %s", agentName),
		}
	}

	return &Result{
		Success:      true,
		Type:         TypeDynamic,
		Output:       "placeholder: dynamic workflow completed",
		AgentResults: agentResults,
		Iterations:   1,
	}, nil
}

// GetConfig returns the workflow configuration.
func (e *Engine) GetConfig() *Config {
	return e.config
}
