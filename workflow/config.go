package workflow

// Type represents the workflow execution type.
type Type string

const (
	// TypeSequential executes agents one after another in order.
	TypeSequential Type = "sequential"

	// TypeParallel executes agents concurrently.
	TypeParallel Type = "parallel"

	// TypeLoop executes agents iteratively for refinement.
	TypeLoop Type = "loop"

	// TypeDynamic lets the coordinator decide execution pattern.
	TypeDynamic Type = "dynamic"
)

// ParseType converts a string to a workflow Type.
// Returns TypeSequential for invalid or unrecognized types.
func ParseType(s string) Type {
	switch s {
	case "sequential":
		return TypeSequential
	case "parallel":
		return TypeParallel
	case "loop":
		return TypeLoop
	case "dynamic":
		return TypeDynamic
	default:
		return TypeSequential
	}
}

// Config defines the workflow configuration.
type Config struct {
	// Type specifies the workflow execution pattern.
	Type Type

	// Agents is the list of agent names to execute.
	// For TypeDynamic, the first agent should be the coordinator.
	Agents []string

	// Task is the initial task description for the workflow.
	Task string

	// MaxIter is the maximum number of iterations for loop/dynamic workflows.
	MaxIter int

	// ContextData holds additional context data for agents (optional).
	ContextData map[string]interface{}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if len(c.Agents) == 0 {
		return &ConfigError{Field: "Agents", Message: "at least one agent is required"}
	}

	if c.Task == "" {
		return &ConfigError{Field: "Task", Message: "task cannot be empty"}
	}

	if c.Type == TypeLoop && c.MaxIter <= 0 {
		return &ConfigError{Field: "MaxIter", Message: "loop workflow requires MaxIter > 0"}
	}

	if c.Type == TypeDynamic {
		if len(c.Agents) < 2 {
			return &ConfigError{Field: "Agents", Message: "dynamic workflow requires at least a coordinator and one agent"}
		}
		// First agent must be the coordinator
		if c.Agents[0] != "coordinator" {
			return &ConfigError{Field: "Agents", Message: "dynamic workflow requires first agent to be coordinator"}
		}
	}

	return nil
}

// Result represents the workflow execution result.
type Result struct {
	// Success indicates whether the workflow completed successfully.
	Success bool

	// Type is the workflow type that was executed.
	Type Type

	// Output is the final output from the workflow.
	Output string

	// AgentResults contains individual agent execution results.
	AgentResults []AgentResult

	// Iterations is the number of iterations executed (for loop/dynamic workflows).
	Iterations int

	// Error contains any error that occurred during execution.
	Error error
}

// AgentResult represents the execution result of a single agent.
type AgentResult struct {
	// AgentName is the name of the agent.
	AgentName string

	// Output is the output from this agent.
	Output string

	// Error contains any error from agent execution.
	Error error
}

// ConfigError represents a configuration validation error.
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error: " + e.Field + " - " + e.Message
}
