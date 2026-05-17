//go:build legacy_fyne

package widgets

import (
	"sync"
	"testing"

	"fyne.io/fyne/v2/widget"
)

func TestNewStatusBar(t *testing.T) {
	sb := NewStatusBar("Status: ")
	if sb == nil {
		t.Fatal("NewStatusBar should not return nil")
	}
	if sb.GetLabel() == nil {
		t.Error("StatusBar should have a label")
	}
	if sb.GetLabel().Text != "Status: Ready" {
		t.Errorf("Initial text should be 'Status: Ready', got '%s'", sb.GetLabel().Text)
	}
}

func TestNewStatusBarWithLabel(t *testing.T) {
	existingLabel := widget.NewLabel("Existing")
	sb := NewStatusBarWithLabel(existingLabel, "Status: ")

	if sb.GetLabel() != existingLabel {
		t.Error("StatusBar should use the provided label")
	}
}

func TestNewStatusBarWithNilLabel(t *testing.T) {
	sb := NewStatusBarWithLabel(nil, "Status: ")
	if sb == nil {
		t.Fatal("NewStatusBarWithLabel should not return nil for nil label")
	}
	if sb.GetLabel() == nil {
		t.Error("StatusBar should create a new label when given nil")
	}
}

func TestStatusBarUpdate(t *testing.T) {
	sb := NewStatusBar("")

	sb.Update("Processing")
	if sb.GetText() != "Processing" {
		t.Errorf("Expected 'Processing', got '%s'", sb.GetText())
	}

	sb.Update("Done")
	if sb.GetText() != "Done" {
		t.Errorf("Expected 'Done', got '%s'", sb.GetText())
	}
}

func TestStatusBarUpdateWithPrefix(t *testing.T) {
	sb := NewStatusBar("Status: ")

	sb.Update("Running")
	// The internal lastText should have the prefix
	if sb.lastText != "Status: Running" {
		t.Errorf("Expected lastText 'Status: Running', got '%s'", sb.lastText)
	}
	// GetText should return without prefix
	if sb.GetText() != "Running" {
		t.Errorf("Expected GetText() 'Running', got '%s'", sb.GetText())
	}
}

func TestStatusBarCachePrevention(t *testing.T) {
	sb := NewStatusBar("")

	// First update
	sb.Update("Test")
	initialText := sb.lastText

	// Same update should not change lastText (cache prevention)
	sb.Update("Test")
	if sb.lastText != initialText {
		t.Error("Duplicate update should not modify cache")
	}
}

func TestStatusBarClear(t *testing.T) {
	sb := NewStatusBar("Status: ")
	sb.Update("Processing")
	sb.Clear()

	if sb.GetText() != "Ready" {
		t.Errorf("Clear should reset to 'Ready', got '%s'", sb.GetText())
	}
}

func TestStatusBarUpdatef(t *testing.T) {
	sb := NewStatusBar("")

	sb.Updatef("Progress: %d%%", 50)
	if sb.GetText() != "Progress: 50%" {
		t.Errorf("Expected 'Progress: 50%%', got '%s'", sb.GetText())
	}
}

func TestStatusBarConcurrentUpdates(t *testing.T) {
	sb := NewStatusBar("")
	var wg sync.WaitGroup

	// Simulate concurrent updates
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			sb.Updatef("Update %d", n)
		}(i)
	}

	wg.Wait()

	// Just verify no panics occurred and we have some text
	if sb.GetText() == "" {
		t.Error("StatusBar should have some text after updates")
	}
}

func TestStatusBarNilLabel(t *testing.T) {
	sb := &StatusBar{label: nil, prefix: ""}

	// These should not panic
	sb.Update("Test")
	sb.Updatef("Test %d", 1)
	sb.Clear()
}
