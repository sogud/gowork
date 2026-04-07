// Package agents provides agent management functionality.
package agents

import (
	"errors"
	"sync"

	adkagent "google.golang.org/adk/agent"
)

// Agent represents a specialized agent in the system.
type Agent interface {
	// Name returns the unique identifier of the agent.
	Name() string
	// Description returns a human-readable description of the agent.
	Description() string
}

// ExecutableAgent represents an agent that can be executed.
// This interface wraps adkagent.Agent for workflow execution.
type ExecutableAgent interface {
	Agent
	// ADKAgent returns the underlying adk-go agent for execution.
	ADKAgent() adkagent.Agent
}

// adkAgentWrapper wraps an adkagent.Agent to implement ExecutableAgent.
type adkAgentWrapper struct {
	agent adkagent.Agent
}

// NewExecutableAgent wraps an adkagent.Agent to create an ExecutableAgent.
func NewExecutableAgent(agent adkagent.Agent) ExecutableAgent {
	return &adkAgentWrapper{agent: agent}
}

func (w *adkAgentWrapper) Name() string        { return w.agent.Name() }
func (w *adkAgentWrapper) Description() string { return w.agent.Description() }
func (w *adkAgentWrapper) ADKAgent() adkagent.Agent { return w.agent }

// Registry is a thread-safe registry for managing agent instances.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]Agent
}

// NewRegistry creates a new empty agent registry.
func NewRegistry() *Registry {
	return &Registry{
		agents: make(map[string]Agent),
	}
}

// Register adds a new agent to the registry.
// Returns an error if an agent with the same name already exists
// or if the agent is nil.
func (r *Registry) Register(agent Agent) error {
	if agent == nil {
		return errors.New("cannot register nil agent")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := agent.Name()
	if _, exists := r.agents[name]; exists {
		return errors.New("agent with name '" + name + "' already registered")
	}

	r.agents[name] = agent
	return nil
}

// Get retrieves an agent by name.
// Returns an error if the agent is not found.
func (r *Registry) Get(name string) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agent, exists := r.agents[name]
	if !exists {
		return nil, errors.New("agent '" + name + "' not found")
	}

	return agent, nil
}

// GetAll returns a copy of all registered agents.
// Returns a new map to prevent external modification of the internal state.
func (r *Registry) GetAll() map[string]Agent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to prevent external modification
	result := make(map[string]Agent, len(r.agents))
	for name, agent := range r.agents {
		result[name] = agent
	}

	return result
}

// List returns the names of all registered agents.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}

	return names
}