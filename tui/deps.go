// Package tui provides the terminal user interface for gowork.
package tui

// This file ensures required TUI dependencies are tracked in go.mod.
// The bubbles library provides ready-made UI components like:
// - textinput: Text input fields
// - progress: Progress bars
// - spinner: Loading spinners
// - viewport: Scrollable content areas
// These will be used in subsequent phases of TUI implementation.

import (
	_ "github.com/charmbracelet/bubbles" // Required TUI component library
)