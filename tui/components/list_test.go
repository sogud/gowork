package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sogud/gowork/tui/styles"
)

func TestNewList(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme)

	if list.theme != theme {
		t.Error("Expected theme to be set")
	}
	if list.multiSelect {
		t.Error("List should be single-select by default")
	}
}

func TestListWithItems(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1", Description: "First item"},
		{ID: "2", Title: "Item 2", Description: "Second item"},
	}
	list := NewList(theme).WithItems(items)

	if len(list.items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(list.items))
	}
	if list.items[0].Title != "Item 1" {
		t.Errorf("Expected first item title 'Item 1', got '%s'", list.items[0].Title)
	}
}

func TestListWithMultiSelect(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme).WithMultiSelect(true)

	if !list.multiSelect {
		t.Error("List should be multi-select")
	}
}

func TestListWithWidth(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme).WithWidth(50)

	if list.width != 50 {
		t.Errorf("Expected width 50, got %d", list.width)
	}
}

func TestListWithHeight(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme).WithHeight(10)

	if list.height != 10 {
		t.Errorf("Expected height 10, got %d", list.height)
	}
}

func TestListWithTitle(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme).WithTitle("Select Items")

	if list.title != "Select Items" {
		t.Errorf("Expected title 'Select Items', got '%s'", list.title)
	}
}

func TestListInit(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme)

	cmd := list.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestListUpdateNavigation(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
		{ID: "3", Title: "Item 3"},
	}
	list := NewList(theme).WithItems(items)

	// Test down navigation
	updatedModel, _ := list.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedList, ok := updatedModel.(List)
	if !ok {
		t.Fatal("Updated model should be a List")
	}
	if updatedList.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 after down, got %d", updatedList.selectedIndex)
	}

	// Test up navigation
	updatedModel, _ = updatedList.Update(tea.KeyMsg{Type: tea.KeyUp})
	updatedList = updatedModel.(List)
	if updatedList.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 after up, got %d", updatedList.selectedIndex)
	}
}

func TestListUpdateToggleSelect(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
	}
	list := NewList(theme).WithItems(items).WithMultiSelect(true)

	// Toggle selection with space
	updatedModel, _ := list.Update(tea.KeyMsg{Type: tea.KeySpace})
	updatedList, ok := updatedModel.(List)
	if !ok {
		t.Fatal("Updated model should be a List")
	}
	if !updatedList.items[0].Selected {
		t.Error("First item should be selected after space")
	}
}

func TestListUpdateEnterSingleSelect(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
	}
	list := NewList(theme).WithItems(items)

	// Navigate to second item and press enter
	updatedModel, _ := list.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedList := updatedModel.(List)
	updatedModel, _ = updatedList.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updatedList = updatedModel.(List)

	// In single-select mode, enter should select the current item
	if updatedList.selectedID != "2" {
		t.Errorf("Expected selectedID '2' after enter, got '%s'", updatedList.selectedID)
	}
}

func TestListView(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1", Description: "First"},
		{ID: "2", Title: "Item 2", Description: "Second"},
	}
	list := NewList(theme).WithItems(items).WithWidth(40)

	view := list.View()
	if view == "" {
		t.Error("View should not be empty")
	}
}

func TestListGetSelectedIndex(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
	}
	list := NewList(theme).WithItems(items)
	list.selectedIndex = 1

	if list.GetSelectedIndex() != 1 {
		t.Errorf("Expected selected index 1, got %d", list.GetSelectedIndex())
	}
}

func TestListGetSelectedItem(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
	}
	list := NewList(theme).WithItems(items)
	list.selectedIndex = 1

	item := list.GetSelectedItem()
	if item.ID != "2" {
		t.Errorf("Expected selected item ID '2', got '%s'", item.ID)
	}
}

func TestListGetSelectedItems(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1", Selected: true},
		{ID: "2", Title: "Item 2", Selected: false},
		{ID: "3", Title: "Item 3", Selected: true},
	}
	list := NewList(theme).WithItems(items).WithMultiSelect(true)

	selected := list.GetSelectedItems()
	if len(selected) != 2 {
		t.Errorf("Expected 2 selected items, got %d", len(selected))
	}
}

func TestListGetSelectedIDs(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1", Selected: true},
		{ID: "2", Title: "Item 2", Selected: false},
		{ID: "3", Title: "Item 3", Selected: true},
	}
	list := NewList(theme).WithItems(items).WithMultiSelect(true)

	ids := list.GetSelectedIDs()
	if len(ids) != 2 {
		t.Errorf("Expected 2 selected IDs, got %d", len(ids))
	}
	if ids[0] != "1" || ids[1] != "3" {
		t.Errorf("Expected selected IDs ['1', '3'], got %v", ids)
	}
}

func TestListSetSelectedIndex(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
	}
	list := NewList(theme).WithItems(items)

	list.SetSelectedIndex(1)
	if list.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1, got %d", list.selectedIndex)
	}
}

func TestListClearSelection(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1", Selected: true},
		{ID: "2", Title: "Item 2", Selected: true},
	}
	list := NewList(theme).WithItems(items).WithMultiSelect(true)

	list.ClearSelection()
	for _, item := range list.items {
		if item.Selected {
			t.Error("All items should be unselected after ClearSelection")
		}
	}
}

func TestListSelectAll(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
	}
	list := NewList(theme).WithItems(items).WithMultiSelect(true)

	list.SelectAll()
	for _, item := range list.items {
		if !item.Selected {
			t.Error("All items should be selected after SelectAll")
		}
	}
}

func TestListIsEmpty(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme)

	if !list.IsEmpty() {
		t.Error("Empty list should return true for IsEmpty")
	}

	list = NewList(theme).WithItems([]ListItem{{ID: "1", Title: "Item 1"}})
	if list.IsEmpty() {
		t.Error("Non-empty list should return false for IsEmpty")
	}
}

func TestListCount(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
		{ID: "3", Title: "Item 3"},
	}
	list := NewList(theme).WithItems(items)

	if list.Count() != 3 {
		t.Errorf("Expected count 3, got %d", list.Count())
	}
}

func TestListBoundaryNavigation(t *testing.T) {
	theme := styles.DefaultTheme()
	items := []ListItem{
		{ID: "1", Title: "Item 1"},
		{ID: "2", Title: "Item 2"},
	}
	list := NewList(theme).WithItems(items)

	// Try to go up from first item - should stay at 0
	updatedModel, _ := list.Update(tea.KeyMsg{Type: tea.KeyUp})
	updatedList := updatedModel.(List)
	if updatedList.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0 (boundary), got %d", updatedList.selectedIndex)
	}

	// Go to last item and try to go down - should stay at last
	updatedList.selectedIndex = 1
	updatedModel, _ = updatedList.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedList = updatedModel.(List)
	if updatedList.selectedIndex != 1 {
		t.Errorf("Expected selectedIndex 1 (boundary), got %d", updatedList.selectedIndex)
	}
}

func TestListFocus(t *testing.T) {
	theme := styles.DefaultTheme()
	list := NewList(theme)

	list.Focus()
	if !list.focused {
		t.Error("List should be focused after Focus()")
	}

	list.Blur()
	if list.focused {
		t.Error("List should not be focused after Blur()")
	}
}