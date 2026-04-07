// Package tools provides tool implementations for the agent system.
package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// SearchInput represents the input for the web search tool.
type SearchInput struct {
	Query string `json:"query" description:"The search query string"`
}

// SearchOutput represents the output from the web search tool.
type SearchOutput struct {
	Results []SearchResult `json:"results" description:"List of search results"`
}

// SearchResult represents a single search result.
type SearchResult struct {
	Title   string `json:"title" description:"Title of the result"`
	URL     string `json:"url" description:"URL of the result"`
	Snippet string `json:"snippet" description:"Brief snippet/description"`
}

// mockSearchData represents the structure of mock_search.json.
type mockSearchData struct {
	SearchResults []SearchResult `json:"search_results"`
}

// WebSearchToolConfig holds configuration for the web search tool.
type WebSearchToolConfig struct {
	MockDataPath string // Path to mock data JSON file (optional)
}

// NewWebSearchTool creates a web search tool with mock data for demo.
// This tool returns predefined results based on keywords in the query.
func NewWebSearchTool(cfg WebSearchToolConfig) (tool.Tool, error) {
	mockDataPath := cfg.MockDataPath
	if mockDataPath == "" {
		mockDataPath = "data/mock_search.json"
	}

	// Load mock data
	mockData, err := loadMockSearchData(mockDataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load mock search data: %w", err)
	}

	handler := func(ctx tool.Context, input SearchInput) (SearchOutput, error) {
		results := searchMockData(mockData, input.Query)
		return SearchOutput{Results: results}, nil
	}

	return functiontool.New(functiontool.Config{
		Name:        "web_search",
		Description: "Search the web for information (demo mode with mock data)",
	}, handler)
}

// loadMockSearchData loads mock search data from a JSON file.
func loadMockSearchData(path string) (*mockSearchData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mockData mockSearchData
	if err := json.Unmarshal(data, &mockData); err != nil {
		return nil, err
	}

	return &mockData, nil
}

// searchMockData searches the mock data for results matching the query.
func searchMockData(mockData *mockSearchData, query string) []SearchResult {
	queryLower := strings.ToLower(query)

	// Search for results that match keywords in the query
	var results []SearchResult
	for _, result := range mockData.SearchResults {
		// Check if any part of the query matches the result
		titleLower := strings.ToLower(result.Title)
		snippetLower := strings.ToLower(result.Snippet)

		if strings.Contains(titleLower, queryLower) ||
			strings.Contains(snippetLower, queryLower) ||
			strings.Contains(queryLower, strings.ToLower(result.Title)) {
			results = append(results, result)
		}
	}

	// If no results found, return default result
	if len(results) == 0 {
		// Find the "default" result
		for _, result := range mockData.SearchResults {
			if strings.ToLower(result.Title) == "general information" {
				results = append(results, result)
				break
			}
		}
	}

	return results
}