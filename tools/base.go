// Package tools provides tool implementations for the agent system.
package tools

import "google.golang.org/adk/tool"

// ToolRegistry provides a thread-safe registry for managing tool instances.
// This interface is similar to agents.Registry but specifically for tools.
type ToolRegistry interface {
	// Register adds a new tool to the registry.
	Register(tool tool.Tool) error

	// Get retrieves a tool by name.
	Get(name string) (tool.Tool, error)

	// GetAll returns a copy of all registered tools.
	GetAll() map[string]tool.Tool

	// List returns the names of all registered tools.
	List() []string
}