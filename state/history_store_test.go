package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewHistoryStore(t *testing.T) {
	// Create a temp directory for testing
	tempDir := t.TempDir()

	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	if store.GetHistoryDir() != tempDir {
		t.Errorf("Expected history dir to be %v, got %v", tempDir, store.GetHistoryDir())
	}
}

func TestNewHistoryStoreDefault(t *testing.T) {
	store, err := NewHistoryStore()
	if err != nil {
		t.Fatalf("Failed to create HistoryStore with defaults: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(store.GetHistoryDir()); os.IsNotExist(err) {
		t.Errorf("History directory was not created")
	}

	// Clean up
	os.RemoveAll(filepath.Dir(store.GetHistoryDir()))
}

func TestNewHistoryStoreWithMaxRecords(t *testing.T) {
	tempDir := t.TempDir()

	store, err := NewHistoryStore(WithHistoryDir(tempDir), WithMaxRecords(50))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	if store.maxRecords != 50 {
		t.Errorf("Expected maxRecords to be 50, got %v", store.maxRecords)
	}
}

func TestNewHistoryStoreInvalidMaxRecords(t *testing.T) {
	tempDir := t.TempDir()

	_, err := NewHistoryStore(WithHistoryDir(tempDir), WithMaxRecords(0))
	if err == nil {
		t.Errorf("Expected error for invalid maxRecords")
	}
}

func TestHistoryStoreSaveRecord(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	record := WorkflowRecord{
		ID:          "test-workflow-123",
		Task:        "test task",
		Type:        WorkflowSequential,
		Status:      WorkflowCompleted,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		Duration:    time.Hour,
		Agents:      []AgentRecord{{Name: "agent1"}},
		FinalOutput: "final output",
	}

	err = store.SaveRecord(record)
	if err != nil {
		t.Errorf("Failed to save record: %v", err)
	}

	// Verify file exists
	expectedFile := filepath.Join(tempDir, "test-workflow-123.json")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Record file was not created")
	}
}

func TestHistoryStoreGetRecord(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	record := WorkflowRecord{
		ID:          "test-workflow-456",
		Task:        "test task",
		Type:        WorkflowParallel,
		Status:      WorkflowCompleted,
		StartTime:   time.Now(),
		EndTime:     time.Now().Add(time.Hour),
		Duration:    time.Hour,
		Agents:      []AgentRecord{{Name: "agent1"}},
		FinalOutput: "final output",
	}

	// Save record
	err = store.SaveRecord(record)
	if err != nil {
		t.Fatalf("Failed to save record: %v", err)
	}

	// Retrieve record
	retrieved, err := store.GetRecord("test-workflow-456")
	if err != nil {
		t.Errorf("Failed to get record: %v", err)
	}

	if retrieved.ID != record.ID {
		t.Errorf("Expected ID %v, got %v", record.ID, retrieved.ID)
	}
	if retrieved.Task != record.Task {
		t.Errorf("Expected Task %v, got %v", record.Task, retrieved.Task)
	}
	if retrieved.Type != record.Type {
		t.Errorf("Expected Type %v, got %v", record.Type, retrieved.Type)
	}
	if len(retrieved.Agents) != len(record.Agents) {
		t.Errorf("Expected %d agents, got %d", len(record.Agents), len(retrieved.Agents))
	}
}

func TestHistoryStoreGetRecordNotFound(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	_, err = store.GetRecord("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent record")
	}
}

func TestHistoryStoreListRecords(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	// Save multiple records
	now := time.Now()
	for i := 0; i < 5; i++ {
		record := WorkflowRecord{
			ID:        "workflow-" + string(rune('0'+i)),
			Task:      "test task " + string(rune('0'+i)),
			Type:      WorkflowSequential,
			Status:    WorkflowCompleted,
			StartTime: now.Add(time.Duration(i) * time.Hour),
		}
		err = store.SaveRecord(record)
		if err != nil {
			t.Fatalf("Failed to save record: %v", err)
		}
	}

	// List records
	records, err := store.ListRecords()
	if err != nil {
		t.Errorf("Failed to list records: %v", err)
	}

	if len(records) != 5 {
		t.Errorf("Expected 5 records, got %d", len(records))
	}

	// Verify sorted by start time (newest first)
	for i := 0; i < len(records)-1; i++ {
		if records[i].StartTime.Before(records[i+1].StartTime) {
			t.Errorf("Records not sorted correctly: record %d should be after record %d", i, i+1)
		}
	}
}

func TestHistoryStoreListRecordsWithFilter(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	// Save records with different tasks
	records := []WorkflowRecord{
		{ID: "workflow-1", Task: "research machine learning", StartTime: time.Now()},
		{ID: "workflow-2", Task: "write documentation", StartTime: time.Now()},
		{ID: "workflow-3", Task: "analyze machine learning results", StartTime: time.Now()},
		{ID: "workflow-4", Task: "create test cases", StartTime: time.Now()},
	}

	for _, r := range records {
		err = store.SaveRecord(r)
		if err != nil {
			t.Fatalf("Failed to save record: %v", err)
		}
	}

	// Filter by "machine"
	filtered, err := store.ListRecordsWithFilter("machine")
	if err != nil {
		t.Errorf("Failed to list records with filter: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("Expected 2 records matching 'machine', got %d", len(filtered))
	}

	// Verify they are the correct records
	for _, r := range filtered {
		if r.ID != "workflow-1" && r.ID != "workflow-3" {
			t.Errorf("Unexpected record in filtered results: %v", r.ID)
		}
	}
}

func TestHistoryStoreDeleteRecord(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	record := WorkflowRecord{
		ID:        "workflow-to-delete",
		Task:      "test task",
		StartTime: time.Now(),
	}

	// Save record
	err = store.SaveRecord(record)
	if err != nil {
		t.Fatalf("Failed to save record: %v", err)
	}

	// Delete record
	err = store.DeleteRecord("workflow-to-delete")
	if err != nil {
		t.Errorf("Failed to delete record: %v", err)
	}

	// Verify file is deleted
	_, err = store.GetRecord("workflow-to-delete")
	if err == nil {
		t.Errorf("Expected error after deletion")
	}
}

func TestHistoryStoreDeleteRecordNotFound(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	err = store.DeleteRecord("non-existent")
	if err == nil {
		t.Errorf("Expected error for non-existent record")
	}
}

func TestHistoryStoreClearHistory(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	// Save multiple records
	for i := 0; i < 3; i++ {
		record := WorkflowRecord{
			ID:        "workflow-" + string(rune('0'+i)),
			Task:      "test task",
			StartTime: time.Now(),
		}
		err = store.SaveRecord(record)
		if err != nil {
			t.Fatalf("Failed to save record: %v", err)
		}
	}

	// Clear history
	err = store.ClearHistory()
	if err != nil {
		t.Errorf("Failed to clear history: %v", err)
	}

	// Verify all records are deleted
	records, err := store.ListRecords()
	if err != nil {
		t.Errorf("Failed to list records after clear: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("Expected 0 records after clear, got %d", len(records))
	}
}

func TestHistoryStoreCleanupOldRecords(t *testing.T) {
	tempDir := t.TempDir()
	store, err := NewHistoryStore(WithHistoryDir(tempDir), WithMaxRecords(3))
	if err != nil {
		t.Fatalf("Failed to create HistoryStore: %v", err)
	}

	// Save more records than max
	now := time.Now()
	for i := 0; i < 5; i++ {
		record := WorkflowRecord{
			ID:        "workflow-" + string(rune('0'+i)),
			Task:      "test task",
			StartTime: now.Add(time.Duration(i) * time.Hour),
		}
		err = store.SaveRecord(record)
		if err != nil {
			t.Fatalf("Failed to save record: %v", err)
		}
	}

	// Verify only maxRecords remain
	records, err := store.ListRecords()
	if err != nil {
		t.Errorf("Failed to list records: %v", err)
	}

	if len(records) > 3 {
		t.Errorf("Expected at most 3 records after cleanup, got %d", len(records))
	}

	// Verify newest records are kept (sorted by start time, newest first)
	// So records should be workflow-4, workflow-3, workflow-2
	expectedIDs := []string{"workflow-4", "workflow-3", "workflow-2"}
	for i, r := range records {
		if i < len(expectedIDs) && r.ID != expectedIDs[i] {
			t.Errorf("Expected record %d to have ID %v, got %v", i, expectedIDs[i], r.ID)
		}
	}
}

func TestWorkflowRecordToState(t *testing.T) {
	records := []WorkflowRecord{
		{ID: "record-1", Task: "task 1"},
		{ID: "record-2", Task: "task 2"},
	}

	state := WorkflowRecordToState(records)

	if len(state.Records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(state.Records))
	}
	if state.SearchQuery != "" {
		t.Errorf("Expected empty SearchQuery")
	}
	if state.SelectedIndex != 0 {
		t.Errorf("Expected SelectedIndex 0")
	}
	if state.ViewDetail {
		t.Errorf("Expected ViewDetail to be false")
	}
	if state.DetailRecord != nil {
		t.Errorf("Expected DetailRecord to be nil")
	}
}

func TestWorkflowStateToRecord(t *testing.T) {
	ts := time.Now()
	state := &WorkflowState{
		ID:              "workflow-123",
		Task:            "test task",
		Type:            WorkflowSequential,
		Status:          WorkflowCompleted,
		StartTime:       ts,
		ElapsedTime:     time.Hour,
		AgentExecutions: []AgentExecution{
			{
				Name:       "agent1",
				Output:     "output",
				TokensUsed: 100,
				StartTime:  ts,
				ToolCalls:  []ToolCallInfo{{ToolName: "tool1"}},
			},
		},
		FinalOutput: "final",
	}

	record := WorkflowStateToRecord(state)

	if record.ID != state.ID {
		t.Errorf("Expected ID %v, got %v", state.ID, record.ID)
	}
	if record.Task != state.Task {
		t.Errorf("Expected Task %v, got %v", state.Task, record.Task)
	}
	if len(record.Agents) != len(state.AgentExecutions) {
		t.Errorf("Expected %d agents, got %d", len(state.AgentExecutions), len(record.Agents))
	}
	if record.Agents[0].Name != "agent1" {
		t.Errorf("Expected agent name to be agent1")
	}
	if record.Agents[0].Tokens != 100 {
		t.Errorf("Expected tokens to be 100")
	}
	if len(record.Agents[0].ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call")
	}
}