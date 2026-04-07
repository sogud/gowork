// Package tools provides tool implementations for the agent system.
package tools

import (
	"errors"
	"sync"

	"google.golang.org/adk/tool"
)

// registry is a thread-safe implementation of ToolRegistry.
type registry struct {
	mu    sync.RWMutex
	tools map[string]tool.Tool
}

// NewRegistry creates a new empty tool registry.
func NewRegistry() ToolRegistry {
	return &registry{
		tools: make(map[string]tool.Tool),
	}
}

// Register adds a new tool to the registry.
// Returns an error if a tool with the same name already exists
// or if the tool is nil.
func (r *registry) Register(tool tool.Tool) error {
	if tool == nil {
		return errors.New("cannot register nil tool")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return errors.New("tool with name '" + name + "' already registered")
	}

	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name.
// Returns an error if the tool is not found.
func (r *registry) Get(name string) (tool.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, errors.New("tool '" + name + "' not found")
	}

	return tool, nil
}

// GetAll returns a copy of all registered tools.
// Returns a new map to prevent external modification of the internal state.
func (r *registry) GetAll() map[string]tool.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to prevent external modification
	result := make(map[string]tool.Tool, len(r.tools))
	for name, tool := range r.tools {
		result[name] = tool
	}

	return result
}

// List returns the names of all registered tools.
func (r *registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}

	return names
}