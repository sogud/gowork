// Package tui provides the terminal user interface for gowork.
package tui

import (
	"github.com/sogud/gowork/state"
)

// Screen defines the interface for TUI screens.
type Screen interface {
	// Name returns the screen's identifier.
	Name() state.Screen
	// Render renders the screen content.
	Render(model *state.AppState) string
	// Update handles messages and updates the model.
	Update(msg interface{}, model *state.AppState) interface{}
	// HelpText returns the help text for this screen.
	HelpText() string
}

// ScreenRegistry manages registered screens.
type ScreenRegistry struct {
	screens     map[state.Screen]Screen
	screenOrder []state.Screen
}

// NewScreenRegistry creates a new ScreenRegistry.
func NewScreenRegistry() *ScreenRegistry {
	return &ScreenRegistry{
		screens:     make(map[state.Screen]Screen),
		screenOrder: make([]state.Screen, 0),
	}
}

// Register adds a screen to the registry.
func (r *ScreenRegistry) Register(screen Screen) {
	if r.screens == nil {
		r.screens = make(map[state.Screen]Screen)
	}

	// Only add to order if not already registered
	if _, exists := r.screens[screen.Name()]; !exists {
		r.screenOrder = append(r.screenOrder, screen.Name())
	}

	r.screens[screen.Name()] = screen
}

// Get retrieves a screen by name.
func (r *ScreenRegistry) Get(name state.Screen) Screen {
	if r.screens == nil {
		return nil
	}
	return r.screens[name]
}

// GetAll returns all registered screens in registration order.
func (r *ScreenRegistry) GetAll() []Screen {
	if r.screens == nil {
		return nil
	}

	screens := make([]Screen, 0, len(r.screenOrder))
	for _, name := range r.screenOrder {
		if screen, ok := r.screens[name]; ok {
			screens = append(screens, screen)
		}
	}
	return screens
}

// HasScreen checks if a screen is registered.
func (r *ScreenRegistry) HasScreen(name state.Screen) bool {
	if r.screens == nil {
		return false
	}
	_, ok := r.screens[name]
	return ok
}

// ScreenNames returns all registered screen names in order.
func (r *ScreenRegistry) ScreenNames() []state.Screen {
	if r.screenOrder == nil {
		return nil
	}
	names := make([]state.Screen, len(r.screenOrder))
	copy(names, r.screenOrder)
	return names
}

// Count returns the number of registered screens.
func (r *ScreenRegistry) Count() int {
	return len(r.screens)
}