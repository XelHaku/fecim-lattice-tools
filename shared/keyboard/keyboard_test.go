package keyboard

import (
	"strings"
	"testing"

	"fyne.io/fyne/v2"
)

func TestDefaultShortcuts(t *testing.T) {
	shortcuts := DefaultShortcuts()
	
	if len(shortcuts) == 0 {
		t.Error("Expected default shortcuts, got none")
	}
	
	// Check that essential shortcuts are present
	essential := map[Action]bool{
		ActionSave:        false,
		ActionExport:      false,
		ActionReset:       false,
		ActionPauseResume: false,
		ActionHelp:        false,
	}
	
	for _, s := range shortcuts {
		if _, ok := essential[s.Action]; ok {
			essential[s.Action] = true
		}
	}
	
	for action, found := range essential {
		if !found {
			t.Errorf("Missing essential shortcut: %s", action)
		}
	}
}

func TestFormatHelpText(t *testing.T) {
	shortcuts := DefaultShortcuts()
	text := FormatHelpText(shortcuts)
	
	if !strings.Contains(text, "Keyboard Shortcuts") {
		t.Error("Help text should contain header")
	}
	
	if !strings.Contains(text, "Space") {
		t.Error("Help text should mention Space key")
	}
	
	if !strings.Contains(text, "Ctrl+S") {
		t.Error("Help text should mention Ctrl+S")
	}
}

func TestFormatKey(t *testing.T) {
	tests := []struct {
		key      fyne.KeyName
		modifier fyne.KeyModifier
		expected string
	}{
		{fyne.KeyS, fyne.KeyModifierControl, "Ctrl+S"},
		{fyne.KeySpace, 0, "Space"},
		{fyne.KeyUp, 0, "↑"},
		{fyne.KeyDown, 0, "↓"},
		{fyne.KeyLeft, 0, "←"},
		{fyne.KeyRight, 0, "→"},
		{fyne.KeyTab, fyne.KeyModifierShift, "Shift+Tab"},
		{fyne.KeyE, fyne.KeyModifierControl | fyne.KeyModifierShift, "Ctrl+Shift+E"},
	}
	
	for _, tt := range tests {
		result := formatKey(tt.key, tt.modifier)
		if result != tt.expected {
			t.Errorf("formatKey(%v, %v) = %s, want %s", tt.key, tt.modifier, result, tt.expected)
		}
	}
}

func TestManagerHandlers(t *testing.T) {
	// Create a mock manager (we can't create a real window in tests)
	m := &Manager{
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
	}
	
	called := false
	m.SetHandler(ActionSave, func() {
		called = true
	})
	
	// Verify handler was set
	if handler, ok := m.handlers[ActionSave]; ok {
		handler()
	} else {
		t.Error("Handler not set")
	}
	
	if !called {
		t.Error("Handler was not called")
	}
}

func TestManagerSetHandlers(t *testing.T) {
	m := &Manager{
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
	}
	
	saveCalled := false
	exportCalled := false
	
	m.SetHandlers(map[Action]func(){
		ActionSave:   func() { saveCalled = true },
		ActionExport: func() { exportCalled = true },
	})
	
	m.handlers[ActionSave]()
	m.handlers[ActionExport]()
	
	if !saveCalled {
		t.Error("Save handler not called")
	}
	if !exportCalled {
		t.Error("Export handler not called")
	}
}

func TestManagerPauseState(t *testing.T) {
	m := &Manager{
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
		paused:    false,
	}
	
	if m.IsPaused() {
		t.Error("Should not be paused initially")
	}
	
	m.SetPaused(true)
	if !m.IsPaused() {
		t.Error("Should be paused after SetPaused(true)")
	}
	
	result := m.TogglePaused()
	if result != false {
		t.Error("TogglePaused should return new state (false)")
	}
	if m.IsPaused() {
		t.Error("Should not be paused after toggle")
	}
}

func TestAddCustomShortcut(t *testing.T) {
	m := &Manager{
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
	}
	
	initialLen := len(m.shortcuts)
	
	m.AddCustomShortcut("custom_action", fyne.KeyF1, 0, "Custom Action")
	
	if len(m.shortcuts) != initialLen+1 {
		t.Error("Custom shortcut was not added")
	}
	
	lastShortcut := m.shortcuts[len(m.shortcuts)-1]
	if lastShortcut.Action != "custom_action" {
		t.Error("Custom shortcut has wrong action")
	}
	if lastShortcut.Key != fyne.KeyF1 {
		t.Error("Custom shortcut has wrong key")
	}
}

func TestQuickHelpText(t *testing.T) {
	text := QuickHelpText()
	
	if !strings.Contains(text, "Space") {
		t.Error("Quick help should mention Space")
	}
	if !strings.Contains(text, "Ctrl+S") {
		t.Error("Quick help should mention Ctrl+S")
	}
	if !strings.Contains(text, "Help") {
		t.Error("Quick help should mention Help")
	}
}
