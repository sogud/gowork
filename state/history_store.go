package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// HistoryStore manages workflow history records.
// Records are stored as JSON files in ~/.gowork/history/
type HistoryStore struct {
	mu          sync.RWMutex
	historyDir  string
	maxRecords  int
}

// HistoryStoreOption is a functional option for configuring HistoryStore.
type HistoryStoreOption func(*HistoryStore) error

// WithHistoryDir sets a custom history directory.
func WithHistoryDir(dir string) HistoryStoreOption {
	return func(h *HistoryStore) error {
		h.historyDir = dir
		return nil
	}
}

// WithMaxRecords sets the maximum number of records to keep.
func WithMaxRecords(max int) HistoryStoreOption {
	return func(h *HistoryStore) error {
		if max <= 0 {
			return errors.New("max records must be positive")
		}
		h.maxRecords = max
		return nil
	}
}

// NewHistoryStore creates a new HistoryStore.
// By default, it uses ~/.gowork/history and keeps up to 1000 records.
func NewHistoryStore(opts ...HistoryStoreOption) (*HistoryStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	h := &HistoryStore{
		historyDir: filepath.Join(homeDir, ".gowork", "history"),
		maxRecords: 1000,
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Ensure history directory exists
	if err := os.MkdirAll(h.historyDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	return h, nil
}

// SaveRecord saves a workflow record to history.
// The record is saved as a JSON file with the workflow ID as the filename.
func (h *HistoryStore) SaveRecord(record WorkflowRecord) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Create the record file path
	filename := fmt.Sprintf("%s.json", record.ID)
	recordPath := filepath.Join(h.historyDir, filename)

	// Marshal record to JSON
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	// Write to file
	if err := os.WriteFile(recordPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write record file: %w", err)
	}

	// Cleanup old records if we exceed max
	if err := h.cleanupOldRecords(); err != nil {
		// Log the error but don't fail the save
		fmt.Fprintf(os.Stderr, "warning: failed to cleanup old records: %v\n", err)
	}

	return nil
}

// GetRecord retrieves a workflow record by ID.
func (h *HistoryStore) GetRecord(id string) (*WorkflowRecord, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	filename := fmt.Sprintf("%s.json", id)
	filepath := filepath.Join(h.historyDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("record not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read record file: %w", err)
	}

	var record WorkflowRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal record: %w", err)
	}

	return &record, nil
}

// ListRecords lists all workflow records, sorted by start time (newest first).
func (h *HistoryStore) ListRecords() ([]WorkflowRecord, error) {
	return h.ListRecordsWithFilter("")
}

// ListRecordsWithFilter lists workflow records filtered by a search query.
// The query is matched against the task description (case-insensitive).
func (h *HistoryStore) ListRecordsWithFilter(query string) ([]WorkflowRecord, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Read all record files
	files, err := os.ReadDir(h.historyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var records []WorkflowRecord
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filepath := filepath.Join(h.historyDir, file.Name())
		data, err := os.ReadFile(filepath)
		if err != nil {
			// Skip files we can't read
			continue
		}

		var record WorkflowRecord
		if err := json.Unmarshal(data, &record); err != nil {
			// Skip invalid files
			continue
		}

		// Apply filter
		if query == "" || strings.Contains(strings.ToLower(record.Task), strings.ToLower(query)) {
			records = append(records, record)
		}
	}

	// Sort by start time (newest first)
	sort.Slice(records, func(i, j int) bool {
		return records[i].StartTime.After(records[j].StartTime)
	})

	return records, nil
}

// DeleteRecord deletes a workflow record by ID.
func (h *HistoryStore) DeleteRecord(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	filename := fmt.Sprintf("%s.json", id)
	filepath := filepath.Join(h.historyDir, filename)

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("record not found: %s", id)
		}
		return fmt.Errorf("failed to delete record file: %w", err)
	}

	return nil
}

// ClearHistory deletes all workflow records.
func (h *HistoryStore) ClearHistory() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	files, err := os.ReadDir(h.historyDir)
	if err != nil {
		return fmt.Errorf("failed to read history directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filepath := filepath.Join(h.historyDir, file.Name())
		if err := os.Remove(filepath); err != nil {
			// Log but continue
			fmt.Fprintf(os.Stderr, "warning: failed to delete %s: %v\n", file.Name(), err)
		}
	}

	return nil
}

// GetHistoryDir returns the history directory path.
func (h *HistoryStore) GetHistoryDir() string {
	return h.historyDir
}

// cleanupOldRecords removes old records when we exceed maxRecords.
// This must be called with the lock already held.
func (h *HistoryStore) cleanupOldRecords() error {
	files, err := os.ReadDir(h.historyDir)
	if err != nil {
		return err
	}

	// Filter to only JSON files
	var jsonFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, file)
		}
	}

	// If we're under the limit, nothing to do
	if len(jsonFiles) <= h.maxRecords {
		return nil
	}

	// Read file info to get modification times
	type fileWithTime struct {
		name    string
		modTime time.Time
	}

	var filesWithTimes []fileWithTime
	for _, file := range jsonFiles {
		info, err := file.Info()
		if err != nil {
			continue
		}
		filesWithTimes = append(filesWithTimes, fileWithTime{
			name:    file.Name(),
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time (oldest first)
	sort.Slice(filesWithTimes, func(i, j int) bool {
		return filesWithTimes[i].modTime.Before(filesWithTimes[j].modTime)
	})

	// Delete oldest files
	toDelete := len(filesWithTimes) - h.maxRecords
	for i := 0; i < toDelete; i++ {
		filepath := filepath.Join(h.historyDir, filesWithTimes[i].name)
		if err := os.Remove(filepath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to delete old record %s: %v\n", filesWithTimes[i].name, err)
		}
	}

	return nil
}

// WorkflowRecordToState creates a HistoryState from a list of workflow records.
func WorkflowRecordToState(records []WorkflowRecord) HistoryState {
	return HistoryState{
		Records:       records,
		SearchQuery:    "",
		SelectedIndex: 0,
		ViewDetail:    false,
		DetailRecord:  nil,
	}
}

// WorkflowStateToRecord converts a WorkflowState to a WorkflowRecord for storage.
func WorkflowStateToRecord(state *WorkflowState) WorkflowRecord {
	agents := make([]AgentRecord, len(state.AgentExecutions))
	for i, exec := range state.AgentExecutions {
		agents[i] = AgentRecord{
			Name:      exec.Name,
			Input:     "", // Input not tracked in execution
			Output:    exec.Output,
			Tokens:    exec.TokensUsed,
			Duration:  now().Sub(exec.StartTime), // Approximate duration
			ToolCalls: exec.ToolCalls,
		}
	}

	return WorkflowRecord{
		ID:          state.ID,
		Task:        state.Task,
		Type:        state.Type,
		Status:      state.Status,
		StartTime:   state.StartTime,
		EndTime:     now(),
		Duration:    state.ElapsedTime,
		Agents:      agents,
		FinalOutput: state.FinalOutput,
	}
}