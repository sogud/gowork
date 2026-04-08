package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
)

func TestNewApp(t *testing.T) {
	initialState := state.NewAppState()
	store := state.NewStateStore(initialState, nil)

	app, err := NewApp(store)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	if app == nil {
		t.Fatal("Expected app to be non-nil")
	}

	if app.store == nil {
		t.Error("Expected store to be set")
	}

	if app.registry == nil {
		t.Error("Expected registry to be set")
	}
}

func TestAppInit(t *testing.T) {
	initialState := state.NewAppState()
	store := state.NewStateStore(initialState, nil)

	app, err := NewApp(store)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	cmd := app.Init()
	// Init should return a command (or nil for no initial command)
	// We just verify it doesn't panic
	_ = cmd
}

func TestAppUpdateQuit(t *testing.T) {
	initialState := state.NewAppState()
	store := state.NewStateStore(initialState, nil)

	app, err := NewApp(store)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	// Test ctrl+c quit
	updatedModel, cmd := app.Update(tea.KeyMsg{
		Type: tea.KeyCtrlC,
	})

	// Should return a quit command
	if cmd == nil {
		t.Error("Expected quit command for Ctrl+C")
	}

	// Check that the model is still an App
	updatedApp, ok := updatedModel.(*App)
	if !ok {
		t.Error("Updated model should be an App")
	}
	_ = updatedApp
}

func TestAppUpdateScreenNavigation(t *testing.T) {
	initialState := state.NewAppState()
	store := state.NewStateStore(initialState, nil)

	app, err := NewApp(store)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	// Register a mock screen for testing
	mockScreen := &mockScreen{name: state.ScreenHome, renderStr: "home"}
	app.registry.Register(mockScreen)

	// Test that navigation keys are handled (implementation specific)
	// For now, we just verify Update doesn't panic
	_, _ = app.Update(tea.KeyMsg{
		Type: tea.KeyRight,
	})
}

func TestAppView(t *testing.T) {
	initialState := state.NewAppState()
	store := state.NewStateStore(initialState, nil)

	app, err := NewApp(store)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	// Register mock screens
	app.registry.Register(&mockScreen{name: state.ScreenHome, renderStr: "Home Screen"})

	view := app.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestAppGetState(t *testing.T) {
	initialState := state.NewAppState()
	store := state.NewStateStore(initialState, nil)

	app, err := NewApp(store)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	retrievedState := app.GetState()
	if retrievedState.CurrentScreen != initialState.CurrentScreen {
		t.Errorf("Expected CurrentScreen %v, got %v", initialState.CurrentScreen, retrievedState.CurrentScreen)
	}
}

func TestAppSetScreen(t *testing.T) {
	initialState := state.NewAppState()
	store := state.NewStateStore(initialState, nil)

	app, err := NewApp(store)
	if err != nil {
		t.Fatalf("NewApp failed: %v", err)
	}

	// Register mock screens
	app.registry.Register(&mockScreen{name: state.ScreenHome, renderStr: "Home Screen"})
	app.registry.Register(&mockScreen{name: state.ScreenTaskInput, renderStr: "Task Input Screen"})

	// Change screen
	app.SetScreen(state.ScreenTaskInput)

	// Verify screen was changed
	if app.GetState().CurrentScreen != state.ScreenTaskInput {
		t.Errorf("Expected CurrentScreen to be ScreenTaskInput, got %v", app.GetState().CurrentScreen)
	}
}