package widgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/undo"
)

// UndoToolbar provides undo/redo buttons with keyboard shortcut support.
type UndoToolbar struct {
	widget.BaseWidget
	manager    *undo.Manager
	undoBtn    *widget.Button
	redoBtn    *widget.Button
	container  *fyne.Container
	
	// Optional: show tooltips with action descriptions
	showTooltips bool
}

// NewUndoToolbar creates a new undo/redo toolbar connected to the given manager.
func NewUndoToolbar(manager *undo.Manager) *UndoToolbar {
	t := &UndoToolbar{
		manager:      manager,
		showTooltips: true,
	}
	t.ExtendBaseWidget(t)

	// Create undo button
	t.undoBtn = widget.NewButtonWithIcon("", theme.ContentUndoIcon(), func() {
		if manager.Undo() {
			DebugInteraction("Undo")
		}
	})
	t.undoBtn.Importance = widget.LowImportance

	// Create redo button
	t.redoBtn = widget.NewButtonWithIcon("", theme.ContentRedoIcon(), func() {
		if manager.Redo() {
			DebugInteraction("Redo")
		}
	})
	t.redoBtn.Importance = widget.LowImportance

	// Initial state
	t.updateButtonStates()

	// Subscribe to manager changes
	manager.SetOnChange(func() {
		// Schedule UI update on main thread
		fyne.Do(func() {
			t.updateButtonStates()
		})
	})

	t.container = container.NewHBox(t.undoBtn, t.redoBtn)
	return t
}

// updateButtonStates enables/disables buttons based on stack state.
func (t *UndoToolbar) updateButtonStates() {
	if t.manager.CanUndo() {
		t.undoBtn.Enable()
		if t.showTooltips {
			desc := t.manager.UndoDescription()
			if desc != "" {
				// Note: Fyne doesn't have native tooltip support on buttons,
				// but we can show in a status bar or as button text
			}
		}
	} else {
		t.undoBtn.Disable()
	}

	if t.manager.CanRedo() {
		t.redoBtn.Enable()
	} else {
		t.redoBtn.Disable()
	}
}

// CreateRenderer implements fyne.Widget.
func (t *UndoToolbar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.container)
}

// UndoButton returns the undo button for custom layout.
func (t *UndoToolbar) UndoButton() *widget.Button {
	return t.undoBtn
}

// RedoButton returns the redo button for custom layout.
func (t *UndoToolbar) RedoButton() *widget.Button {
	return t.redoBtn
}

// SetShowTooltips enables or disables tooltip descriptions.
func (t *UndoToolbar) SetShowTooltips(show bool) {
	t.showTooltips = show
	t.updateButtonStates()
}

// UndoRedoKeyHandler handles Ctrl+Z (undo) and Ctrl+Y/Ctrl+Shift+Z (redo) shortcuts.
// Attach this to a window or canvas with SetOnTypedKey.
type UndoRedoKeyHandler struct {
	manager *undo.Manager
}

// NewUndoRedoKeyHandler creates a new keyboard shortcut handler.
func NewUndoRedoKeyHandler(manager *undo.Manager) *UndoRedoKeyHandler {
	return &UndoRedoKeyHandler{manager: manager}
}

// HandleKey processes keyboard events for undo/redo.
// Returns true if the key was handled.
func (h *UndoRedoKeyHandler) HandleKey(key *fyne.KeyEvent, mods fyne.KeyModifier) bool {
	// Check for Ctrl modifier (Command on macOS)
	hasCtrl := mods&fyne.KeyModifierControl != 0 || mods&fyne.KeyModifierSuper != 0
	hasShift := mods&fyne.KeyModifierShift != 0

	if !hasCtrl {
		return false
	}

	switch key.Name {
	case fyne.KeyZ:
		if hasShift {
			// Ctrl+Shift+Z = Redo
			if h.manager.Redo() {
				DebugInteraction("Redo (Ctrl+Shift+Z)")
				return true
			}
		} else {
			// Ctrl+Z = Undo
			if h.manager.Undo() {
				DebugInteraction("Undo (Ctrl+Z)")
				return true
			}
		}
	case fyne.KeyY:
		// Ctrl+Y = Redo (Windows style)
		if h.manager.Redo() {
			DebugInteraction("Redo (Ctrl+Y)")
			return true
		}
	}

	return false
}

// UndoStatusWidget displays the current undo/redo state (for debugging or status bar).
type UndoStatusWidget struct {
	widget.BaseWidget
	manager *undo.Manager
	label   *widget.Label
}

// NewUndoStatusWidget creates a widget that displays undo/redo stack sizes.
func NewUndoStatusWidget(manager *undo.Manager) *UndoStatusWidget {
	w := &UndoStatusWidget{
		manager: manager,
	}
	w.ExtendBaseWidget(w)
	w.label = widget.NewLabel("")
	w.updateLabel()

	manager.SetOnChange(func() {
		fyne.Do(func() {
			w.updateLabel()
		})
	})

	return w
}

func (w *UndoStatusWidget) updateLabel() {
	undoCount := w.manager.UndoCount()
	redoCount := w.manager.RedoCount()
	
	text := fmt.Sprintf("Undo: %d | Redo: %d", undoCount, redoCount)
	
	if desc := w.manager.UndoDescription(); desc != "" && undoCount > 0 {
		text = fmt.Sprintf("Undo (%d): %s | Redo: %d", undoCount, desc, redoCount)
	}
	
	w.label.SetText(text)
}

// CreateRenderer implements fyne.Widget.
func (w *UndoStatusWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.label)
}

// WithUndoKeyboard wraps a canvas object to handle undo/redo keyboard shortcuts.
// This is useful for adding keyboard support to existing containers.
type WithUndoKeyboard struct {
	widget.BaseWidget
	content  fyne.CanvasObject
	handler  *UndoRedoKeyHandler
	window   fyne.Window
}

// NewWithUndoKeyboard wraps content with undo/redo keyboard handling.
func NewWithUndoKeyboard(content fyne.CanvasObject, manager *undo.Manager, window fyne.Window) *WithUndoKeyboard {
	w := &WithUndoKeyboard{
		content: content,
		handler: NewUndoRedoKeyHandler(manager),
		window:  window,
	}
	w.ExtendBaseWidget(w)
	
	// Install keyboard handler on the window canvas
	if window != nil {
		w.installKeyboardHandler(window)
	}
	
	return w
}

// installKeyboardHandler sets up the window to handle undo/redo shortcuts.
func (w *WithUndoKeyboard) installKeyboardHandler(window fyne.Window) {
	canvas := window.Canvas()
	
	// Add keyboard shortcut for Ctrl+Z
	canvas.AddShortcut(&fyne.ShortcutUndo{}, func(_ fyne.Shortcut) {
		w.handler.manager.Undo()
		DebugInteraction("Undo (shortcut)")
	})
	
	// Add keyboard shortcut for Ctrl+Y (and Ctrl+Shift+Z is handled by Fyne as ShortcutRedo)
	canvas.AddShortcut(&fyne.ShortcutRedo{}, func(_ fyne.Shortcut) {
		w.handler.manager.Redo()
		DebugInteraction("Redo (shortcut)")
	})
}

// CreateRenderer implements fyne.Widget.
func (w *WithUndoKeyboard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.content)
}
