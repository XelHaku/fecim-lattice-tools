//go:build legacy_fyne

package widgets

import (
	"sync"
	"testing"

	"fyne.io/fyne/v2/widget"
)

// Note: Some Fyne widgets (ProgressBar, Entry, Select, etc.) require a running
// Fyne app to function properly due to internal Refresh() calls.
// Tests for those widgets are skipped or test only nil-safety.

func TestSafeUpdateLabel(t *testing.T) {
	label := widget.NewLabel("Initial")

	SafeUpdateLabel(label, "Updated")

	if label.Text != "Updated" {
		t.Errorf("Expected 'Updated', got '%s'", label.Text)
	}
}

func TestSafeUpdateLabelNil(t *testing.T) {
	// Should not panic
	SafeUpdateLabel(nil, "Test")
}

func TestSafeUpdateProgressNil(t *testing.T) {
	// Should not panic
	SafeUpdateProgress(nil, 0.5)
}

func TestSafeUpdateProgressInfiniteNil(t *testing.T) {
	// Should not panic
	SafeUpdateProgressInfinite(nil, true)
	SafeUpdateProgressInfinite(nil, false)
}

func TestSafeRefreshNil(t *testing.T) {
	// Should not panic
	SafeRefresh(nil)
}

func TestSafeShowHide(t *testing.T) {
	label := widget.NewLabel("Test")

	SafeHide(label)
	if label.Visible() {
		t.Error("Label should be hidden")
	}

	SafeShow(label)
	if !label.Visible() {
		t.Error("Label should be visible")
	}

	SafeShowHide(label, false)
	if label.Visible() {
		t.Error("Label should be hidden")
	}

	SafeShowHide(label, true)
	if !label.Visible() {
		t.Error("Label should be visible")
	}
}

func TestSafeShowHideNil(t *testing.T) {
	// Should not panic
	SafeShow(nil)
	SafeHide(nil)
	SafeShowHide(nil, true)
}

func TestSafeEnableDisableNil(t *testing.T) {
	// Should not panic
	SafeEnable(nil)
	SafeDisable(nil)
	SafeEnableDisable(nil, true)
}

func TestSafeSetEntryNil(t *testing.T) {
	// Should not panic
	SafeSetEntry(nil, "Test")
}

func TestSafeSetCheckNil(t *testing.T) {
	// Should not panic
	SafeSetCheck(nil, true)
}

func TestSafeSetSliderNil(t *testing.T) {
	// Should not panic
	SafeSetSlider(nil, 50)
}

func TestSafeSetSelectNil(t *testing.T) {
	// Should not panic
	SafeSetSelect(nil, "Test")
}

func TestSafeSetSelectIndexNil(t *testing.T) {
	// Should not panic
	SafeSetSelectIndex(nil, 0)
}

func TestConcurrentSafeUpdateLabels(t *testing.T) {
	label := widget.NewLabel("Start")

	var wg sync.WaitGroup

	// Simulate concurrent updates from multiple goroutines
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			SafeUpdateLabel(label, "Update")
		}(i)
	}

	wg.Wait()

	// Verify no panics occurred and label has text
	if label.Text == "" {
		t.Error("Label should have text")
	}
}
