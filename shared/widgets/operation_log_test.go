//go:build legacy_fyne

package widgets

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

func TestNewOperationLog(t *testing.T) {
	config := OperationLogConfig{
		Title:      "Test Log",
		MaxEntries: 5,
		MinSize:    fyne.NewSize(200, 150),
	}
	ol := NewOperationLog(config)

	if ol == nil {
		t.Fatal("NewOperationLog returned nil")
	}

	entries := ol.GetEntries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}

	if ol.MinSize() != config.MinSize {
		t.Errorf("Expected MinSize %v, got %v", config.MinSize, ol.MinSize())
	}
}

func TestOperationLog_Add(t *testing.T) {
	ol := NewOperationLog(OperationLogConfig{MaxEntries: 5})

	ol.Add("First entry")
	ol.Add("Second entry")

	entries := ol.GetEntries()
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Check that entries have timestamps
	for _, entry := range entries {
		if !strings.HasPrefix(entry, "[") {
			t.Errorf("Entry should start with timestamp bracket: %s", entry)
		}
	}
}

func TestOperationLog_MaxEntries(t *testing.T) {
	ol := NewOperationLog(OperationLogConfig{MaxEntries: 3})

	// Add more entries than max
	for i := 1; i <= 5; i++ {
		ol.Add("Entry")
	}

	entries := ol.GetEntries()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries (max), got %d", len(entries))
	}
}

func TestOperationLog_Clear(t *testing.T) {
	ol := NewOperationLog(OperationLogConfig{})

	ol.Add("Entry 1")
	ol.Add("Entry 2")
	ol.Clear()

	entries := ol.GetEntries()
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}
}

func TestOperationLog_AddWithPrefix(t *testing.T) {
	ol := NewOperationLog(OperationLogConfig{})

	ol.AddWithPrefix("[CUSTOM]", "Custom message")

	entries := ol.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if !strings.HasPrefix(entries[0], "[CUSTOM]") {
		t.Errorf("Entry should start with custom prefix: %s", entries[0])
	}
}

func TestOperationLog_FormattedEntries(t *testing.T) {
	ol := NewOperationLog(OperationLogConfig{})

	ol.AddSuccess("Operation succeeded")
	ol.AddError("Operation failed")
	ol.AddInfo("Information message")

	entries := ol.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Check that formatted entries contain the expected symbols
	if !strings.Contains(entries[0], "✓") {
		t.Errorf("Success entry should contain ✓: %s", entries[0])
	}
	if !strings.Contains(entries[1], "✗") {
		t.Errorf("Error entry should contain ✗: %s", entries[1])
	}
	if !strings.Contains(entries[2], "ℹ") {
		t.Errorf("Info entry should contain ℹ: %s", entries[2])
	}
}

func TestOperationLog_DefaultConfig(t *testing.T) {
	ol := NewOperationLog(OperationLogConfig{})

	if ol.MinSize().Width <= 0 || ol.MinSize().Height <= 0 {
		t.Error("Default MinSize should be positive")
	}
}

func TestOperationLog_GetEntriesCopies(t *testing.T) {
	ol := NewOperationLog(OperationLogConfig{})
	ol.Add("Entry")

	entries1 := ol.GetEntries()
	entries2 := ol.GetEntries()

	// Modify the first copy
	if len(entries1) > 0 {
		entries1[0] = "Modified"
	}

	// Second copy should be unchanged
	if len(entries2) > 0 && entries2[0] == "Modified" {
		t.Error("GetEntries should return a copy, not the original slice")
	}
}
