// Package keyboard provides unified keyboard shortcut handling for all FeCIM modules.
// It defines common shortcuts (Ctrl+S save, Ctrl+E export, Ctrl+R reset, Space pause/resume,
// arrow keys for navigation) and provides helpers for registering them on Fyne windows.
package keyboard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// Action represents a keyboard-triggered action.
type Action string

// Common keyboard actions across all modules.
const (
	ActionSave        Action = "save"         // Ctrl+S: Save current state/data
	ActionExport      Action = "export"       // Ctrl+E: Export data
	ActionReset       Action = "reset"        // Ctrl+R: Reset simulation/view
	ActionPauseResume Action = "pause_resume" // Space: Toggle pause/resume
	ActionNavigateUp  Action = "nav_up"       // Up arrow: Navigate up
	ActionNavigateDown Action = "nav_down"    // Down arrow: Navigate down
	ActionNavigateLeft Action = "nav_left"    // Left arrow: Navigate left/previous
	ActionNavigateRight Action = "nav_right"  // Right arrow: Navigate right/next
	ActionHelp        Action = "help"         // ? or /: Show keyboard help
	ActionStepForward Action = "step_forward" // ]: Step forward (animations)
	ActionStepBack    Action = "step_back"    // [: Step backward (animations)
	ActionIncrease    Action = "increase"     // +/=: Increase value
	ActionDecrease    Action = "decrease"     // -: Decrease value
	ActionNextTab     Action = "next_tab"     // Tab: Next tab
	ActionPrevTab     Action = "prev_tab"     // Shift+Tab: Previous tab
)

// Handler is a function called when a keyboard action is triggered.
type Handler func(action Action)

// Shortcut defines a keyboard shortcut configuration.
type Shortcut struct {
	Action   Action
	Key      fyne.KeyName
	Modifier fyne.KeyModifier
	Label    string // Human-readable description
}

// DefaultShortcuts returns the standard shortcuts for all FeCIM modules.
func DefaultShortcuts() []Shortcut {
	return []Shortcut{
		{ActionSave, fyne.KeyS, fyne.KeyModifierControl, "Save state/data"},
		{ActionExport, fyne.KeyE, fyne.KeyModifierControl, "Export data"},
		{ActionReset, fyne.KeyR, fyne.KeyModifierControl, "Reset simulation"},
		{ActionPauseResume, fyne.KeySpace, 0, "Pause/Resume"},
		{ActionNavigateUp, fyne.KeyUp, 0, "Navigate up"},
		{ActionNavigateDown, fyne.KeyDown, 0, "Navigate down"},
		{ActionNavigateLeft, fyne.KeyLeft, 0, "Navigate left/previous"},
		{ActionNavigateRight, fyne.KeyRight, 0, "Navigate right/next"},
		{ActionHelp, fyne.KeySlash, 0, "Show keyboard help"},
		{ActionStepForward, fyne.KeyRightBracket, 0, "Step forward"},
		{ActionStepBack, fyne.KeyLeftBracket, 0, "Step backward"},
		{ActionIncrease, fyne.KeyEqual, 0, "Increase value"},
		{ActionDecrease, fyne.KeyMinus, 0, "Decrease value"},
		{ActionNextTab, fyne.KeyTab, 0, "Next tab"},
		{ActionPrevTab, fyne.KeyTab, fyne.KeyModifierShift, "Previous tab"},
	}
}

// Manager handles keyboard shortcuts for a window.
type Manager struct {
	window    fyne.Window
	handlers  map[Action]func()
	shortcuts []Shortcut
	paused    bool
}

// NewManager creates a new keyboard shortcut manager for the given window.
func NewManager(window fyne.Window) *Manager {
	return &Manager{
		window:    window,
		handlers:  make(map[Action]func()),
		shortcuts: DefaultShortcuts(),
	}
}

// SetHandler registers a handler for a specific action.
// The handler will be called when the corresponding shortcut is triggered.
func (m *Manager) SetHandler(action Action, handler func()) {
	m.handlers[action] = handler
}

// SetHandlers registers multiple handlers at once.
func (m *Manager) SetHandlers(handlers map[Action]func()) {
	for action, handler := range handlers {
		m.handlers[action] = handler
	}
}

// IsPaused returns the current pause state.
func (m *Manager) IsPaused() bool {
	return m.paused
}

// SetPaused sets the pause state.
func (m *Manager) SetPaused(paused bool) {
	m.paused = paused
}

// TogglePaused toggles the pause state and returns the new state.
func (m *Manager) TogglePaused() bool {
	m.paused = !m.paused
	return m.paused
}

// Register sets up all keyboard shortcuts on the window.
// Call this after setting up all handlers.
func (m *Manager) Register() {
	// Register modifier shortcuts (Ctrl+X, Shift+Tab, etc.) via Canvas.AddShortcut
	for _, s := range m.shortcuts {
		if s.Modifier != 0 {
			shortcut := &desktop.CustomShortcut{
				KeyName:  s.Key,
				Modifier: s.Modifier,
			}
			action := s.Action // Capture for closure
			m.window.Canvas().AddShortcut(shortcut, func(_ fyne.Shortcut) {
				if handler, ok := m.handlers[action]; ok {
					handler()
				}
			})
		}
	}

	// Register non-modifier keys via SetOnTypedKey
	m.window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		m.handleKeyPress(ke)
	})
}

// handleKeyPress processes non-modifier key events.
func (m *Manager) handleKeyPress(ke *fyne.KeyEvent) {
	for _, s := range m.shortcuts {
		if s.Modifier == 0 && s.Key == ke.Name {
			if handler, ok := m.handlers[s.Action]; ok {
				handler()
			}
			return
		}
	}
}

// AddCustomShortcut adds a custom shortcut to the manager.
func (m *Manager) AddCustomShortcut(action Action, key fyne.KeyName, modifier fyne.KeyModifier, label string) {
	m.shortcuts = append(m.shortcuts, Shortcut{
		Action:   action,
		Key:      key,
		Modifier: modifier,
		Label:    label,
	})
}

// GetShortcuts returns all registered shortcuts.
func (m *Manager) GetShortcuts() []Shortcut {
	return m.shortcuts
}

// GetHelpText returns a formatted help text listing all shortcuts.
func (m *Manager) GetHelpText() string {
	return FormatHelpText(m.shortcuts)
}

// FormatHelpText formats shortcuts into a human-readable help string.
func FormatHelpText(shortcuts []Shortcut) string {
	text := "Keyboard Shortcuts:\n\n"
	
	// Group shortcuts by category
	categories := map[string][]Shortcut{
		"Navigation":     {},
		"Simulation":     {},
		"Data":           {},
		"Animation":      {},
		"Misc":           {},
	}
	
	for _, s := range shortcuts {
		switch s.Action {
		case ActionNavigateUp, ActionNavigateDown, ActionNavigateLeft, ActionNavigateRight, ActionNextTab, ActionPrevTab:
			categories["Navigation"] = append(categories["Navigation"], s)
		case ActionPauseResume, ActionReset:
			categories["Simulation"] = append(categories["Simulation"], s)
		case ActionSave, ActionExport:
			categories["Data"] = append(categories["Data"], s)
		case ActionStepForward, ActionStepBack, ActionIncrease, ActionDecrease:
			categories["Animation"] = append(categories["Animation"], s)
		default:
			categories["Misc"] = append(categories["Misc"], s)
		}
	}
	
	order := []string{"Navigation", "Simulation", "Data", "Animation", "Misc"}
	for _, cat := range order {
		shortcuts := categories[cat]
		if len(shortcuts) == 0 {
			continue
		}
		text += cat + ":\n"
		for _, s := range shortcuts {
			text += "  " + formatKey(s.Key, s.Modifier) + " - " + s.Label + "\n"
		}
		text += "\n"
	}
	
	return text
}

// formatKey formats a key and modifier combination for display.
func formatKey(key fyne.KeyName, modifier fyne.KeyModifier) string {
	result := ""
	if modifier&fyne.KeyModifierControl != 0 {
		result += "Ctrl+"
	}
	if modifier&fyne.KeyModifierShift != 0 {
		result += "Shift+"
	}
	if modifier&fyne.KeyModifierAlt != 0 {
		result += "Alt+"
	}
	if modifier&fyne.KeyModifierSuper != 0 {
		result += "Super+"
	}
	
	// Map key names to human-readable strings
	keyStr := string(key)
	switch key {
	case fyne.KeySpace:
		keyStr = "Space"
	case fyne.KeyUp:
		keyStr = "↑"
	case fyne.KeyDown:
		keyStr = "↓"
	case fyne.KeyLeft:
		keyStr = "←"
	case fyne.KeyRight:
		keyStr = "→"
	case fyne.KeyTab:
		keyStr = "Tab"
	case fyne.KeyEqual:
		keyStr = "+/="
	case fyne.KeyMinus:
		keyStr = "-"
	case fyne.KeyLeftBracket:
		keyStr = "["
	case fyne.KeyRightBracket:
		keyStr = "]"
	case fyne.KeySlash:
		keyStr = "?//"
	}
	
	return result + keyStr
}
