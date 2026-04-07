package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
)

// Run starts the TUI application.
func Run() error {
	// Create initial state
	initialState := state.NewAppState()

	// Create state store
	store := state.NewStateStore(initialState, nil)

	// Create app
	app, err := NewApp(store)
	if err != nil {
		return fmt.Errorf("failed to create TUI app: %w", err)
	}

	// Ensure cleanup on exit
	defer app.Close()

	// Create bubbletea program
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Handle final model if needed
	if finalModel != nil {
		// Can check for any final state updates here
	}

	return nil
}

// RunWithState starts the TUI application with a custom initial state.
func RunWithState(initialState state.AppState) error {
	// Create state store with custom initial state
	store := state.NewStateStore(initialState, nil)

	// Create app
	app, err := NewApp(store)
	if err != nil {
		return fmt.Errorf("failed to create TUI app: %w", err)
	}

	// Ensure cleanup on exit
	defer app.Close()

	// Create bubbletea program
	p := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	_, err = p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

// Exit gracefully exits the TUI.
func Exit(code int) {
	os.Exit(code)
}

// IsTerminal checks if stdout is a terminal.
func IsTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}