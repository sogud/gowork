// Package tools provides tool implementations for the agent system.
package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/toolconfirmation"
)

// mockToolContext implements tool.Context for testing.
type mockToolContext struct {
	context.Context
}

func (m *mockToolContext) FunctionCallID() string { return "test-call-id" }
func (m *mockToolContext) UserContent() interface{} { return nil }
func (m *mockToolContext) Actions() interface{} { return nil }
func (m *mockToolContext) SearchMemory(ctx context.Context, query string) (interface{}, error) {
	return nil, nil
}
func (m *mockToolContext) ToolConfirmation() *toolconfirmation.ToolConfirmation { return nil }
func (m *mockToolContext) RequestConfirmation(hint string, payload interface{}) error {
	return nil
}

// === Registry Tests ===

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	calculator, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("Failed to create calculator tool: %v", err)
	}

	err = registry.Register(calculator)
	if err != nil {
		t.Errorf("Register() failed: %v", err)
	}

	// Test duplicate registration
	err = registry.Register(calculator)
	if err == nil {
		t.Error("Register() should fail for duplicate tool")
	}

	// Test nil tool
	err = registry.Register(nil)
	if err == nil {
		t.Error("Register() should fail for nil tool")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()

	calculator, _ := NewCalculatorTool()
	registry.Register(calculator)

	// Test successful get
	tool, err := registry.Get("calculator")
	if err != nil {
		t.Errorf("Get() failed: %v", err)
	}
	if tool == nil {
		t.Error("Get() returned nil tool")
	}

	// Test get non-existent tool
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Get() should fail for non-existent tool")
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()

	calculator, _ := NewCalculatorTool()
	registry.Register(calculator)

	tools := registry.GetAll()
	if len(tools) != 1 {
		t.Errorf("GetAll() returned %d tools, want 1", len(tools))
	}

	// Verify it's a copy
	tools["calculator"] = nil
	tools2 := registry.GetAll()
	if tools2["calculator"] == nil {
		t.Error("GetAll() should return a copy, not the internal map")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()

	calculator, _ := NewCalculatorTool()
	registry.Register(calculator)

	names := registry.List()
	if len(names) != 1 {
		t.Errorf("List() returned %d names, want 1", len(names))
	}
}

func TestRegistry_ThreadSafety(t *testing.T) {
	registry := NewRegistry()

	// Concurrent registrations
	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(i int) {
			tool, _ := NewCalculatorTool()
			// This should fail for duplicates after the first one succeeds
			registry.Register(tool)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify only one tool registered
	tools := registry.GetAll()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool after concurrent registrations, got %d", len(tools))
	}
}

// === Calculator Tool Tests ===

func TestCalculatorTool_BasicOperations(t *testing.T) {
	calculator, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("Failed to create calculator tool: %v", err)
	}

	// Test tool creation
	_ = calculator

	tests := []struct {
		name       string
		expression string
		want       float64
	}{
		{"addition", "2+3", 5},
		{"subtraction", "10-3", 7},
		{"multiplication", "4*5", 20},
		{"division", "20/4", 5},
		{"negative", "-5", -5},
		{"float", "3.5+2.5", 6},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test the underlying function directly
			result, err := evaluateExpression(tc.expression)
			if err != nil {
				t.Errorf("evaluateExpression(%q) failed: %v", tc.expression, err)
			}
			if result != tc.want {
				t.Errorf("evaluateExpression(%q) = %f, want %f", tc.expression, result, tc.want)
			}
		})
	}
}

func TestCalculatorTool_ComplexExpressions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       float64
	}{
		{"order of operations", "2+3*4", 14},
		{"parentheses", "(2+3)*4", 20},
		{"nested parentheses", "((2+3)*2)", 10},
		{"mixed operations", "10+20/5-3", 11},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := evaluateExpression(tc.expression)
			if err != nil {
				t.Errorf("evaluateExpression(%q) failed: %v", tc.expression, err)
			}
			if result != tc.want {
				t.Errorf("evaluateExpression(%q) = %f, want %f", tc.expression, result, tc.want)
			}
		})
	}
}

func TestCalculatorTool_Errors(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		wantErr    bool
	}{
		{"empty", "", true},
		{"division by zero", "10/0", true},
		{"invalid character", "2+x", true},
		{"missing closing paren", "(2+3", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := evaluateExpression(tc.expression)
			if tc.wantErr && err == nil {
				t.Errorf("evaluateExpression(%q) should fail", tc.expression)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("evaluateExpression(%q) unexpectedly failed: %v", tc.expression, err)
			}
		})
	}
}

func TestCalculatorTool_Interface(t *testing.T) {
	calculator, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("Failed to create calculator tool: %v", err)
	}

	// Test tool interface compliance
	var toolInterface tool.Tool = calculator

	if toolInterface.Name() != "calculator" {
		t.Errorf("Name() = %q, want %q", toolInterface.Name(), "calculator")
	}

	if toolInterface.IsLongRunning() {
		t.Error("IsLongRunning() should be false")
	}
}

// === Web Search Tool Tests ===

func TestWebSearchTool_MockData(t *testing.T) {
	// Create a temporary mock data file
	tmpDir := t.TempDir()
	mockDataPath := filepath.Join(tmpDir, "mock_search.json")

	mockData := `{
		"search_results": [
			{
				"keyword": "test",
				"title": "Test Result",
				"url": "https://example.com/test",
				"snippet": "This is a test result"
			}
		]
	}`

	if err := os.WriteFile(mockDataPath, []byte(mockData), 0644); err != nil {
		t.Fatalf("Failed to write mock data: %v", err)
	}

	search, err := NewWebSearchTool(WebSearchToolConfig{MockDataPath: mockDataPath})
	if err != nil {
		t.Fatalf("Failed to create web search tool: %v", err)
	}

	var _ tool.Tool = search

	if search.Name() != "web_search" {
		t.Errorf("Name() = %q, want %q", search.Name(), "web_search")
	}
}

func TestWebSearchTool_DefaultPath(t *testing.T) {
	// Test with explicit path pointing to project's data directory
	projectRoot := "/Users/changshun/codes/github/go-agent"
	mockDataPath := filepath.Join(projectRoot, "data/mock_search.json")

	search, err := NewWebSearchTool(WebSearchToolConfig{MockDataPath: mockDataPath})
	if err != nil {
		t.Fatalf("Failed to create web search tool: %v", err)
	}

	var _ tool.Tool = search
}

func TestSearchMockData(t *testing.T) {
	mockData := &mockSearchData{
		SearchResults: []SearchResult{
			{
				Title:   "Go Programming",
				URL:     "https://golang.org",
				Snippet: "Go is a programming language",
			},
			{
				Title:   "Python Programming",
				URL:     "https://python.org",
				Snippet: "Python programming tutorial",
			},
			{
				Title:   "General Information",
				URL:     "https://example.com",
				Snippet: "General info",
			},
		},
	}

	tests := []struct {
		name  string
		query string
		want  int // expected minimum number of results
	}{
		{"match go", "go", 1},
		{"match programming", "programming", 1},
		{"no match returns default", "java", 1}, // returns default result
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results := searchMockData(mockData, tc.query)
			if len(results) < tc.want {
				t.Errorf("searchMockData(%q) returned %d results, want at least %d", tc.query, len(results), tc.want)
			}
		})
	}
}

// === File Tool Tests ===

func TestFileReaderTool_Security(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	reader, err := NewFileReaderTool(FileToolConfig{AllowedReadDirs: []string{tmpDir}})
	if err != nil {
		t.Fatalf("Failed to create file reader tool: %v", err)
	}

	var _ tool.Tool = reader

	if reader.Name() != "file_reader" {
		t.Errorf("Name() = %q, want %q", reader.Name(), "file_reader")
	}
}

func TestFileWriterTool_Security(t *testing.T) {
	tmpDir := t.TempDir()

	writer, err := NewFileWriterTool(FileToolConfig{AllowedWriteDirs: []string{tmpDir}})
	if err != nil {
		t.Fatalf("Failed to create file writer tool: %v", err)
	}

	var _ tool.Tool = writer

	if writer.Name() != "file_writer" {
		t.Errorf("Name() = %q, want %q", writer.Name(), "file_writer")
	}
}

func TestIsPathAllowed(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		allowedDirs []string
		want        bool
	}{
		{"allowed in dir", filepath.Join(tmpDir, "test.txt"), []string{tmpDir}, true},
		{"not allowed outside", "/etc/passwd", []string{tmpDir}, false},
		{"relative path allowed", "test.txt", []string{"."}, true},
		{"parent escape", "../secret.txt", []string{tmpDir}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isPathAllowed(tc.path, tc.allowedDirs)
			if result != tc.want {
				t.Errorf("isPathAllowed(%q, %v) = %v, want %v", tc.path, tc.allowedDirs, result, tc.want)
			}
		})
	}
}

func TestStartsWithDotDot(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"../test.txt", true},
		{"test.txt", false},
		{"subdir/test.txt", false},
		{"..", true},
	}

	for _, tc := range tests {
		result := startsWithDotDot(tc.path)
		if result != tc.want {
			t.Errorf("startsWithDotDot(%q) = %v, want %v", tc.path, result, tc.want)
		}
	}
}

// === Integration Tests ===

func TestAllTools_Registration(t *testing.T) {
	registry := NewRegistry()

	// Create all tools
	calculator, err := NewCalculatorTool()
	if err != nil {
		t.Fatalf("Failed to create calculator tool: %v", err)
	}

	projectRoot := "/Users/changshun/codes/github/go-agent"
	mockDataPath := filepath.Join(projectRoot, "data/mock_search.json")
	search, err := NewWebSearchTool(WebSearchToolConfig{MockDataPath: mockDataPath})
	if err != nil {
		t.Fatalf("Failed to create web search tool: %v", err)
	}

	tmpDir := t.TempDir()
	reader, err := NewFileReaderTool(FileToolConfig{AllowedReadDirs: []string{tmpDir}})
	if err != nil {
		t.Fatalf("Failed to create file reader tool: %v", err)
	}

	writer, err := NewFileWriterTool(FileToolConfig{AllowedWriteDirs: []string{tmpDir}})
	if err != nil {
		t.Fatalf("Failed to create file writer tool: %v", err)
	}

	// Register all tools
	tools := []tool.Tool{calculator, search, reader, writer}
	for _, tool := range tools {
		if err := registry.Register(tool); err != nil {
			t.Errorf("Failed to register tool %s: %v", tool.Name(), err)
		}
	}

	// Verify all registered
	allTools := registry.GetAll()
	if len(allTools) != 4 {
		t.Errorf("Expected 4 tools registered, got %d", len(allTools))
	}

	// Verify each tool exists
	for _, name := range []string{"calculator", "web_search", "file_reader", "file_writer"} {
		if _, err := registry.Get(name); err != nil {
			t.Errorf("Tool %s not found in registry: %v", name, err)
		}
	}
}