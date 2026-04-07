// Package tools provides tool implementations for the agent system.
package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// FileInput represents the input for file operations.
type FileInput struct {
	Path string `json:"path" description:"The file path to read or write"`
}

// FileWriteInput represents the input for file write operations.
type FileWriteInput struct {
	Path    string `json:"path" description:"The file path to write"`
	Content string `json:"content" description:"The content to write to the file"`
}

// FileOutput represents the output from file operations.
type FileOutput struct {
	Content string `json:"content" description:"The content read from the file"`
	Path    string `json:"path" description:"The file path that was accessed"`
}

// FileToolConfig holds configuration for file tools.
type FileToolConfig struct {
	AllowedReadDirs  []string // Directories allowed for reading
	AllowedWriteDirs []string // Directories allowed for writing
}

// NewFileReaderTool creates a file reader tool with security checks.
// Only allows reads from configured directories.
func NewFileReaderTool(cfg FileToolConfig) (tool.Tool, error) {
	if len(cfg.AllowedReadDirs) == 0 {
		cfg.AllowedReadDirs = []string{"."} // Default to current directory
	}

	handler := func(ctx tool.Context, input FileInput) (FileOutput, error) {
		// Security check: verify path is within allowed directories
		if !isPathAllowed(input.Path, cfg.AllowedReadDirs) {
			return FileOutput{}, fmt.Errorf("path '%s' is not in allowed read directories", input.Path)
		}

		// Read file content
		content, err := os.ReadFile(input.Path)
		if err != nil {
			return FileOutput{}, fmt.Errorf("failed to read file '%s': %w", input.Path, err)
		}

		return FileOutput{
			Content: string(content),
			Path:    input.Path,
		}, nil
	}

	return functiontool.New(functiontool.Config{
		Name:        "file_reader",
		Description: "Read file content from allowed directories",
	}, handler)
}

// NewFileWriterTool creates a file writer tool with security checks.
// Only allows writes to configured directories.
func NewFileWriterTool(cfg FileToolConfig) (tool.Tool, error) {
	if len(cfg.AllowedWriteDirs) == 0 {
		cfg.AllowedWriteDirs = []string{"."} // Default to current directory
	}

	handler := func(ctx tool.Context, input FileWriteInput) (FileOutput, error) {
		// Security check: verify path is within allowed directories
		if !isPathAllowed(input.Path, cfg.AllowedWriteDirs) {
			return FileOutput{}, fmt.Errorf("path '%s' is not in allowed write directories", input.Path)
		}

		// Ensure directory exists
		dir := filepath.Dir(input.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return FileOutput{}, fmt.Errorf("failed to create directory '%s': %w", dir, err)
		}

		// Write file content
		if err := os.WriteFile(input.Path, []byte(input.Content), 0644); err != nil {
			return FileOutput{}, fmt.Errorf("failed to write file '%s': %w", input.Path, err)
		}

		return FileOutput{
			Content: input.Content,
			Path:    input.Path,
		}, nil
	}

	return functiontool.New(functiontool.Config{
		Name:        "file_writer",
		Description: "Write content to files in allowed directories",
	}, handler)
}

// isPathAllowed checks if a path is within the allowed directories.
func isPathAllowed(path string, allowedDirs []string) bool {
	// Get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check against each allowed directory
	for _, dir := range allowedDirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			continue
		}

		// Check if path is within the allowed directory
		relPath, err := filepath.Rel(absDir, absPath)
		if err != nil {
			continue
		}

		// If relative path doesn't start with "..", it's within the directory
		if !startsWithDotDot(relPath) {
			return true
		}
	}

	return false
}

// startsWithDotDot checks if a relative path starts with ".." (escaping parent directory).
func startsWithDotDot(path string) bool {
	// Clean the path first
	path = filepath.Clean(path)

	// Check if path starts with ".."
	return path == ".." || strings.HasPrefix(path, "../")
}