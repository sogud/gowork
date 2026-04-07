package tui

import (
	"testing"

	"github.com/sogud/gowork/state"
)

// mockScreen is a test implementation of the Screen interface.
type mockScreen struct {
	name       state.Screen
	renderStr  string
	helpText   string
}

func (m *mockScreen) Name() state.Screen {
	return m.name
}

func (m *mockScreen) Render(model *state.AppState) string {
	return m.renderStr
}

func (m *mockScreen) Update(msg interface{}, model *state.AppState) interface{} {
	return nil
}

func (m *mockScreen) HelpText() string {
	return m.helpText
}

func TestScreenRegistry_Register(t *testing.T) {
	registry := NewScreenRegistry()

	screen1 := &mockScreen{name: state.ScreenHome, renderStr: "home", helpText: "home help"}
	screen2 := &mockScreen{name: state.ScreenTaskInput, renderStr: "task", helpText: "task help"}

	// Test registering screens
	registry.Register(screen1)
	registry.Register(screen2)

	// Verify screens are registered
	if len(registry.screens) != 2 {
		t.Errorf("Expected 2 screens, got %d", len(registry.screens))
	}

	// Verify screen order
	if len(registry.screenOrder) != 2 {
		t.Errorf("Expected 2 screens in order, got %d", len(registry.screenOrder))
	}

	if registry.screenOrder[0] != state.ScreenHome {
		t.Errorf("Expected first screen to be ScreenHome, got %v", registry.screenOrder[0])
	}

	if registry.screenOrder[1] != state.ScreenTaskInput {
		t.Errorf("Expected second screen to be ScreenTaskInput, got %v", registry.screenOrder[1])
	}
}

func TestScreenRegistry_Get(t *testing.T) {
	registry := NewScreenRegistry()

	screen := &mockScreen{name: state.ScreenHome, renderStr: "home", helpText: "home help"}
	registry.Register(screen)

	// Test getting existing screen
	got := registry.Get(state.ScreenHome)
	if got == nil {
		t.Error("Expected to get screen, got nil")
	}

	if got.Name() != state.ScreenHome {
		t.Errorf("Expected screen name ScreenHome, got %v", got.Name())
	}

	// Test getting non-existent screen
	notFound := registry.Get(state.ScreenConfig)
	if notFound != nil {
		t.Error("Expected nil for non-existent screen")
	}
}

func TestScreenRegistry_GetAll(t *testing.T) {
	registry := NewScreenRegistry()

	screen1 := &mockScreen{name: state.ScreenHome, renderStr: "home", helpText: "home help"}
	screen2 := &mockScreen{name: state.ScreenTaskInput, renderStr: "task", helpText: "task help"}
	screen3 := &mockScreen{name: state.ScreenMonitor, renderStr: "monitor", helpText: "monitor help"}

	registry.Register(screen1)
	registry.Register(screen2)
	registry.Register(screen3)

	screens := registry.GetAll()

	if len(screens) != 3 {
		t.Errorf("Expected 3 screens, got %d", len(screens))
	}

	// Verify order is maintained
	if screens[0].Name() != state.ScreenHome {
		t.Errorf("Expected first screen to be ScreenHome, got %v", screens[0].Name())
	}

	if screens[1].Name() != state.ScreenTaskInput {
		t.Errorf("Expected second screen to be ScreenTaskInput, got %v", screens[1].Name())
	}

	if screens[2].Name() != state.ScreenMonitor {
		t.Errorf("Expected third screen to be ScreenMonitor, got %v", screens[2].Name())
	}
}

func TestScreenRegistry_HasScreen(t *testing.T) {
	registry := NewScreenRegistry()

	screen := &mockScreen{name: state.ScreenHome, renderStr: "home", helpText: "home help"}
	registry.Register(screen)

	// Test existing screen
	if !registry.HasScreen(state.ScreenHome) {
		t.Error("Expected HasScreen to return true for ScreenHome")
	}

	// Test non-existent screen
	if registry.HasScreen(state.ScreenConfig) {
		t.Error("Expected HasScreen to return false for non-existent screen")
	}
}