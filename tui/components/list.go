package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sogud/gowork/tui/styles"
)

// ListItem represents a single item in the list.
type ListItem struct {
	ID          string
	Title       string
	Description string
	Selected    bool
	Disabled    bool
}

// List is a list selection component that implements tea.Model.
// It supports both single-select and multi-select modes.
type List struct {
	theme        styles.Theme
	items        []ListItem
	title        string
	selectedIndex int
	selectedID   string
	width        int
	height       int
	multiSelect  bool
	focused      bool
	showHelp     bool
}

// NewList creates a new List component with the given theme.
func NewList(theme styles.Theme) List {
	return List{
		theme:        theme,
		items:        []ListItem{},
		selectedIndex: 0,
		multiSelect:   false,
		focused:       true,
		showHelp:      true,
	}
}

// WithItems sets the list items.
func (l List) WithItems(items []ListItem) List {
	l.items = items
	if len(items) > 0 && l.selectedIndex >= len(items) {
		l.selectedIndex = len(items) - 1
	}
	return l
}

// WithTitle sets the list title.
func (l List) WithTitle(title string) List {
	l.title = title
	return l
}

// WithWidth sets the width of the list.
func (l List) WithWidth(width int) List {
	l.width = width
	return l
}

// WithHeight sets the height of the list.
func (l List) WithHeight(height int) List {
	l.height = height
	return l
}

// WithMultiSelect enables or disables multi-select mode.
func (l List) WithMultiSelect(multiSelect bool) List {
	l.multiSelect = multiSelect
	return l
}

// WithShowHelp sets whether to show help text.
func (l List) WithShowHelp(show bool) List {
	l.showHelp = show
	return l
}

// Focus focuses the list.
func (l *List) Focus() {
	l.focused = true
}

// Blur unfocuses the list.
func (l *List) Blur() {
	l.focused = false
}

// Focused returns whether the list is focused.
func (l List) Focused() bool {
	return l.focused
}

// Init initializes the list (tea.Model interface).
func (l List) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model (tea.Model interface).
func (l List) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !l.focused || len(l.items) == 0 {
		return l, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			// Move selection up
			if l.selectedIndex > 0 {
				l.selectedIndex--
			}

		case tea.KeyDown:
			// Move selection down
			if l.selectedIndex < len(l.items) - 1 {
				l.selectedIndex++
			}

		case tea.KeySpace:
			// Toggle selection in multi-select mode
			if l.multiSelect {
				l.items[l.selectedIndex].Selected = !l.items[l.selectedIndex].Selected
			}

		case tea.KeyEnter:
			// Confirm selection
			if l.multiSelect {
				// In multi-select, enter just confirms current selections
			} else {
				// In single-select, select the current item
				l.selectedID = l.items[l.selectedIndex].ID
				l.items[l.selectedIndex].Selected = true
			}

		case tea.KeyHome:
			l.selectedIndex = 0

		case tea.KeyEnd:
			l.selectedIndex = len(l.items) - 1
		}

		// Handle rune-based shortcuts
		switch msg.String() {
		case "k": // vim-style up
			if l.selectedIndex > 0 {
				l.selectedIndex--
			}
		case "j": // vim-style down
			if l.selectedIndex < len(l.items) - 1 {
				l.selectedIndex++
			}
		case "a": // Select all in multi-select
			if l.multiSelect {
				for i := range l.items {
					l.items[i].Selected = true
				}
			}
		case "d": // Deselect all in multi-select
			if l.multiSelect {
				for i := range l.items {
					l.items[i].Selected = false
				}
			}
		}
	}

	return l, nil
}

// View renders the list (tea.Model interface).
func (l List) View() string {
	if len(l.items) == 0 {
		return l.renderEmpty()
	}

	var sections []string

	// Add title if set
	if l.title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(l.theme.Primary.ToLipgloss()).
			Bold(true).
			Padding(0, 1)
		sections = append(sections, titleStyle.Render(l.title))
	}

	// Render items
	itemsView := l.renderItems()
	sections = append(sections, itemsView)

	// Add help text if enabled
	if l.showHelp {
		helpText := l.renderHelp()
		sections = append(sections, helpText)
	}

	// Apply border and sizing
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(l.theme.Border.ToLipgloss())

	if l.focused {
		borderStyle = borderStyle.BorderForeground(l.theme.Primary.ToLipgloss())
	}

	if l.width > 0 {
		borderStyle = borderStyle.Width(l.width)
	}

	if l.height > 0 {
		borderStyle = borderStyle.Height(l.height)
	}

	return borderStyle.Render(content)
}

// renderItems renders all list items.
func (l List) renderItems() string {
	var lines []string

	for i, item := range l.items {
		line := l.renderItem(item, i == l.selectedIndex)
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// renderItem renders a single list item.
func (l List) renderItem(item ListItem, isSelected bool) string {
	// Checkbox indicator
	var checkbox string
	if l.multiSelect {
		if item.Selected {
			checkbox = "[x] "
		} else {
			checkbox = "[ ] "
		}
	} else {
		if isSelected {
			checkbox = "> "
		} else {
			checkbox = "  "
		}
	}

	// Title styling
	var titleStyle lipgloss.Style
	if isSelected && l.focused {
		titleStyle = lipgloss.NewStyle().
			Foreground(l.theme.Text.ToLipgloss()).
			Bold(true)
	} else if item.Disabled {
		titleStyle = lipgloss.NewStyle().
			Foreground(l.theme.TextMuted.ToLipgloss())
	} else {
		titleStyle = lipgloss.NewStyle().
			Foreground(l.theme.Text.ToLipgloss())
	}

	// Description styling
	descStyle := lipgloss.NewStyle().
		Foreground(l.theme.TextMuted.ToLipgloss())

	// Build the line
	title := titleStyle.Render(item.Title)
	var line string
	if item.Description != "" {
		desc := descStyle.Render(" - " + item.Description)
		line = checkbox + title + desc
	} else {
		line = checkbox + title
	}

	// Highlight the selected line
	if isSelected && l.focused {
		highlightStyle := lipgloss.NewStyle().
			Background(l.theme.Surface.ToLipgloss())
		return highlightStyle.Render(line)
	}

	return line
}

// renderHelp renders help text.
func (l List) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(l.theme.TextMuted.ToLipgloss()).
		Italic(true).
		Padding(0, 1)

	if l.multiSelect {
		return helpStyle.Render("Space: toggle | Enter: confirm | ↑↓: navigate")
	}
	return helpStyle.Render("Enter: select | ↑↓: navigate")
}

// renderEmpty renders empty list message.
func (l List) renderEmpty() string {
	emptyStyle := lipgloss.NewStyle().
		Foreground(l.theme.TextMuted.ToLipgloss()).
		Italic(true).
		Padding(1, 1)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(l.theme.Border.ToLipgloss())

	if l.width > 0 {
		borderStyle = borderStyle.Width(l.width)
	}

	return borderStyle.Render(emptyStyle.Render("No items available"))
}

// GetSelectedIndex returns the currently selected index.
func (l List) GetSelectedIndex() int {
	return l.selectedIndex
}

// GetSelectedItem returns the currently selected item.
func (l List) GetSelectedItem() ListItem {
	if len(l.items) == 0 || l.selectedIndex >= len(l.items) {
		return ListItem{}
	}
	return l.items[l.selectedIndex]
}

// GetSelectedItems returns all selected items (for multi-select mode).
func (l List) GetSelectedItems() []ListItem {
	selected := []ListItem{}
	for _, item := range l.items {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// GetSelectedIDs returns IDs of all selected items.
func (l List) GetSelectedIDs() []string {
	ids := []string{}
	for _, item := range l.items {
		if item.Selected {
			ids = append(ids, item.ID)
		}
	}
	return ids
}

// GetSelectedID returns the ID of the selected item (single-select mode).
func (l List) GetSelectedID() string {
	return l.selectedID
}

// SetSelectedIndex sets the selected index.
func (l *List) SetSelectedIndex(index int) {
	if index < 0 {
		index = 0
	}
	if index >= len(l.items) {
		index = len(l.items) - 1
	}
	if index < 0 {
		index = 0
	}
	l.selectedIndex = index
}

// ClearSelection clears all selections.
func (l *List) ClearSelection() {
	for i := range l.items {
		l.items[i].Selected = false
	}
	l.selectedID = ""
}

// SelectAll selects all items (multi-select mode).
func (l *List) SelectAll() {
	for i := range l.items {
		l.items[i].Selected = true
	}
}

// IsEmpty returns true if the list has no items.
func (l List) IsEmpty() bool {
	return len(l.items) == 0
}

// Count returns the number of items in the list.
func (l List) Count() int {
	return len(l.items)
}

// SetItems updates the list items.
func (l *List) SetItems(items []ListItem) {
	l.items = items
	if len(items) > 0 && l.selectedIndex >= len(items) {
		l.selectedIndex = len(items) - 1
	}
}

// SetWidth sets the width dynamically.
func (l *List) SetWidth(width int) {
	l.width = width
}

// SetHeight sets the height dynamically.
func (l *List) SetHeight(height int) {
	l.height = height
}

// Items returns the list items.
func (l List) Items() []ListItem {
	return l.items
}

// SetTitle sets the title.
func (l *List) SetTitle(title string) {
	l.title = title
}

// MultiSelect returns whether multi-select mode is enabled.
func (l List) MultiSelect() bool {
	return l.multiSelect
}

// SetMultiSelect enables or disables multi-select mode.
func (l *List) SetMultiSelect(enabled bool) {
	l.multiSelect = enabled
}

// String returns a summary of the list state.
func (l List) String() string {
	if l.multiSelect {
		selected := l.GetSelectedIDs()
		return fmt.Sprintf("List: %d items, %d selected (%v)", len(l.items), len(selected), selected)
	}
	return fmt.Sprintf("List: %d items, selected index %d", len(l.items), l.selectedIndex)
}

// FilterItems returns items filtered by a predicate function.
func (l List) FilterItems(predicate func(ListItem) bool) []ListItem {
	filtered := []ListItem{}
	for _, item := range l.items {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// EnableItem enables or disables a specific item by ID.
func (l *List) EnableItem(id string, enabled bool) {
	for i := range l.items {
		if l.items[i].ID == id {
			l.items[i].Disabled = !enabled
			break
		}
	}
}

// SelectItemByID selects an item by its ID.
func (l *List) SelectItemByID(id string) {
	for i, item := range l.items {
		if item.ID == id {
			l.selectedIndex = i
			if !l.multiSelect {
				l.selectedID = id
			}
			break
		}
	}
}

// ToggleItemByID toggles selection for an item by ID (multi-select mode).
func (l *List) ToggleItemByID(id string) {
	for i := range l.items {
		if l.items[i].ID == id {
			l.items[i].Selected = !l.items[i].Selected
			break
		}
	}
}

// FindItemByID finds an item by its ID.
func (l List) FindItemByID(id string) (ListItem, bool) {
	for _, item := range l.items {
		if item.ID == id {
			return item, true
		}
	}
	return ListItem{}, false
}

// AddItem adds a new item to the list.
func (l *List) AddItem(item ListItem) {
	l.items = append(l.items, item)
}

// RemoveItemByID removes an item by its ID.
func (l *List) RemoveItemByID(id string) {
	for i, item := range l.items {
		if item.ID == id {
			l.items = append(l.items[:i], l.items[i+1:]...)
			if l.selectedIndex >= len(l.items) {
				l.selectedIndex = len(l.items) - 1
			}
			break
		}
	}
}

// ViewCompact renders the list without border (for embedding).
func (l List) ViewCompact() string {
	if len(l.items) == 0 {
		return ""
	}
	return l.renderItems()
}

// Header returns the title as a header string.
func (l List) Header() string {
	if l.title == "" {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(l.theme.Primary.ToLipgloss()).
		Bold(true).
		Render(l.title)
}

// Footer returns the help text.
func (l List) Footer() string {
	if !l.showHelp {
		return ""
	}
	return l.renderHelp()
}

// Summary returns a summary string for display.
func (l List) Summary() string {
	if l.multiSelect {
		selected := len(l.GetSelectedItems())
		return strings.Join([]string{
			fmt.Sprintf("%d/%d selected", selected, len(l.items)),
		}, " | ")
	}
	return fmt.Sprintf("Item %d of %d", l.selectedIndex+1, len(l.items))
}