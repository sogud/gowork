package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/state"
	"github.com/sogud/gowork/tui/styles"
)

// App is the main TUI application implementing tea.Model.
type App struct {
	store    *state.StateStore
	registry *ScreenRegistry
	theme    styles.Theme
	layout   styles.Layout
	sub      state.Subscription
	width    int
	height   int
}

// NewApp creates a new TUI App.
func NewApp(store *state.StateStore) (*App, error) {
	if store == nil {
		return nil, nil
	}

	theme := styles.DefaultTheme()
	layout := styles.NewLayout(theme)
	registry := NewScreenRegistry()

	app := &App{
		store:    store,
		registry: registry,
		theme:    theme,
		layout:   layout,
	}

	// Subscribe to state changes
	app.sub = store.Subscribe(func(event state.StateChangeEvent) {
		// State changes trigger a redraw via tea.Batch
		// The subscription callback is handled in the Update method
	})

	return app, nil
}

// Init initializes the application (tea.Model interface).
func (a *App) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (tea.Model interface).
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// Quit on Ctrl+C or Escape
			return a, tea.Quit

		case tea.KeyRight:
			// Navigate to next screen
			a.nextScreen()

		case tea.KeyLeft:
			// Navigate to previous screen
			a.prevScreen()
		}

		// Pass key events to current screen
		currentState := a.store.GetState()
		currentScreen := a.registry.Get(currentState.CurrentScreen)
		if currentScreen != nil {
			currentScreen.Update(m, &currentState)
		}

	case tea.WindowSizeMsg:
		a.width = m.Width
		a.height = m.Height
	}

	return a, nil
}

// View renders the application (tea.Model interface).
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Get current screen
	currentState := a.store.GetState()
	screen := a.registry.Get(currentState.CurrentScreen)

	// Render screen content
	var content string
	if screen != nil {
		content = screen.Render(&currentState)
	} else {
		content = a.renderPlaceholder(currentState.CurrentScreen)
	}

	// Build full view with footer
	return a.buildView(content)
}

// GetState returns the current application state.
func (a *App) GetState() state.AppState {
	return a.store.GetState()
}

// SetScreen changes the current screen.
func (a *App) SetScreen(screen state.Screen) {
	a.store.SetScreen(screen)
}

// GetRegistry returns the screen registry.
func (a *App) GetRegistry() *ScreenRegistry {
	return a.registry
}

// GetTheme returns the current theme.
func (a *App) GetTheme() styles.Theme {
	return a.theme
}

// GetLayout returns the current layout.
func (a *App) GetLayout() styles.Layout {
	return a.layout
}

// Close cleans up resources.
func (a *App) Close() {
	if a.sub.ID != "" && a.store != nil {
		a.store.Unsubscribe(a.sub)
	}
}

// nextScreen navigates to the next screen in order.
func (a *App) nextScreen() {
	screens := a.registry.ScreenNames()
	if len(screens) == 0 {
		return
	}

	current := a.store.GetState().CurrentScreen
	for i, s := range screens {
		if s == current && i < len(screens)-1 {
			a.store.SetScreen(screens[i+1])
			return
		}
	}
}

// prevScreen navigates to the previous screen in order.
func (a *App) prevScreen() {
	screens := a.registry.ScreenNames()
	if len(screens) == 0 {
		return
	}

	current := a.store.GetState().CurrentScreen
	for i, s := range screens {
		if s == current && i > 0 {
			a.store.SetScreen(screens[i-1])
			return
		}
	}
}

// renderPlaceholder renders a placeholder for unimplemented screens.
func (a *App) renderPlaceholder(screen state.Screen) string {
	title := a.theme.TitleStyle(screen.String())
	content := a.layout.CenterText("Screen not yet implemented", a.width-4)
	help := a.layout.HelpText.Render("Press Ctrl+C to quit")

	return a.layout.JoinVertical(title, content, help)
}

// buildView builds the full view with header, content, and footer.
func (a *App) buildView(content string) string {
	// Header
	header := a.layout.HeaderLine("gowork", "multi-agent collaboration framework")

	// Footer with navigation info
	footer := a.buildFooter()

	// Build main content area
	mainContent := a.layout.BoxWithBorder("", a.width-2, a.height-6).Render(content)

	return a.layout.JoinVertical(header, mainContent, footer)
}

// buildFooter builds the footer with help text and navigation.
func (a *App) buildFooter() string {
	currentScreen := a.store.GetState().CurrentScreen

	// Get help text from current screen if available
	screen := a.registry.Get(currentScreen)
	var helpText string
	if screen != nil {
		helpText = screen.HelpText()
	} else {
		helpText = "Use arrow keys to navigate, Ctrl+C to quit"
	}

	// Screen indicator
	screenIndicator := a.theme.MutedStyle(currentScreen.String())

	// Build footer
	help := a.layout.HelpText.Render(helpText)
	return a.layout.JoinHorizontal(screenIndicator, "  ", help)
}