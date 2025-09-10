package history

import (
	"path/filepath"
	"testing"
	"time"
)

// createTestHistoryManager creates a history manager with a temporary file for testing
func createTestHistoryManager(t *testing.T) *HistoryManager {
	// Create temporary directory
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "test_sshm_history.json")

	hm := &HistoryManager{
		historyPath: historyPath,
		history:     &ConnectionHistory{Connections: make(map[string]ConnectionInfo)},
	}

	return hm
}

func TestNewHistoryManager(t *testing.T) {
	hm, err := NewHistoryManager()
	if err != nil {
		t.Fatalf("NewHistoryManager() error = %v", err)
	}
	if hm == nil {
		t.Fatal("NewHistoryManager() returned nil")
	}
	if hm.historyPath == "" {
		t.Error("Expected historyPath to be set")
	}
}

func TestHistoryManager_RecordConnection(t *testing.T) {
	hm := createTestHistoryManager(t)

	// Add a connection
	err := hm.RecordConnection("testhost")
	if err != nil {
		t.Errorf("RecordConnection() error = %v", err)
	}

	// Check that the connection was added
	lastUsed, exists := hm.GetLastConnectionTime("testhost")
	if !exists || lastUsed.IsZero() {
		t.Error("Expected connection to be recorded")
	}
}

func TestHistoryManager_GetLastConnectionTime(t *testing.T) {
	hm := createTestHistoryManager(t)

	// Test with no connections
	lastUsed, exists := hm.GetLastConnectionTime("nonexistent-testhost")
	if exists || !lastUsed.IsZero() {
		t.Error("Expected no connection for non-existent host")
	}

	// Add a connection
	err := hm.RecordConnection("testhost")
	if err != nil {
		t.Errorf("RecordConnection() error = %v", err)
	}

	// Test with existing connection
	lastUsed, exists = hm.GetLastConnectionTime("testhost")
	if !exists || lastUsed.IsZero() {
		t.Error("Expected non-zero time for existing host")
	}

	// Check that the time is recent (within last minute)
	if time.Since(lastUsed) > time.Minute {
		t.Error("Last used time seems too old")
	}
}

func TestHistoryManager_GetConnectionCount(t *testing.T) {
	hm := createTestHistoryManager(t)

	// Add same host multiple times
	for i := 0; i < 3; i++ {
		err := hm.RecordConnection("testhost-count")
		if err != nil {
			t.Errorf("RecordConnection() error = %v", err)
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Should have correct count
	count := hm.GetConnectionCount("testhost-count")
	if count != 3 {
		t.Errorf("Expected connection count 3, got %d", count)
	}
}
