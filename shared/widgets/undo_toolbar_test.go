//go:build legacy_fyne

package widgets

import (
	"testing"

	"fecim-lattice-tools/shared/undo"
)

func TestUndoToolbarCreation(t *testing.T) {
	manager := undo.NewManager(10)
	toolbar := NewUndoToolbar(manager)

	if toolbar == nil {
		t.Fatal("Expected toolbar to be created")
	}

	// Initially both buttons should be disabled
	if toolbar.UndoButton().Disabled() != true {
		t.Error("Expected undo button to be disabled initially")
	}
	if toolbar.RedoButton().Disabled() != true {
		t.Error("Expected redo button to be disabled initially")
	}
}

func TestUndoToolbarButtonStatesManual(t *testing.T) {
	// Note: We can't test the automatic button state updates because
	// they rely on fyne.Do which requires a running event loop.
	// This test verifies the updateButtonStates logic directly.
	manager := undo.NewManager(10)
	toolbar := NewUndoToolbar(manager)

	// Clear the onChange callback to prevent fyne.Do issues
	manager.SetOnChange(nil)

	// Execute a command
	cmd := undo.NewIntCommand("test", 0, 42, func(v int) { _ = v })
	manager.Execute(cmd)

	// Manually trigger button state update
	toolbar.updateButtonStates()

	// Undo should be enabled, redo should be disabled
	if toolbar.UndoButton().Disabled() {
		t.Error("Expected undo button to be enabled after execute")
	}
	if !toolbar.RedoButton().Disabled() {
		t.Error("Expected redo button to be disabled after execute")
	}

	// After undo, undo should be disabled, redo should be enabled
	manager.Undo()
	toolbar.updateButtonStates()

	if !toolbar.UndoButton().Disabled() {
		t.Error("Expected undo button to be disabled after undo")
	}
	if toolbar.RedoButton().Disabled() {
		t.Error("Expected redo button to be enabled after undo")
	}
}

func TestUndoRedoKeyHandler(t *testing.T) {
	manager := undo.NewManager(10)
	handler := NewUndoRedoKeyHandler(manager)

	// Execute a command
	value := 0
	cmd := undo.NewIntCommand("test", 0, 42, func(v int) { value = v })
	manager.Execute(cmd)

	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}

	// Note: We can't easily test the keyboard handler without a full Fyne context,
	// but we can verify the handler is created correctly
	if handler == nil {
		t.Fatal("Expected handler to be created")
	}
	if handler.manager != manager {
		t.Error("Handler should reference the correct manager")
	}
}

func TestUndoStatusWidgetCreation(t *testing.T) {
	manager := undo.NewManager(10)

	// Clear the onChange callback before creating widget to prevent fyne.Do issues
	manager.SetOnChange(nil)

	status := NewUndoStatusWidget(manager)

	if status == nil {
		t.Fatal("Expected status widget to be created")
	}

	// Initial state should show 0 counts
	// Note: We can't easily check the label text without rendering,
	// but we can verify the widget is created correctly
	if status.manager != manager {
		t.Error("Status widget should reference the correct manager")
	}
}
