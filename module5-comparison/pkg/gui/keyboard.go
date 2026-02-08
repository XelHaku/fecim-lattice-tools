package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/keyboard"
)

// setupKeyboard initializes keyboard shortcuts using the shared keyboard manager.
func (ca *ComparisonApp) setupKeyboard() {
	km := keyboard.NewManager(ca.window)

	// Register common handlers
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionSave: func() {
			debug.Println("Save shortcut triggered (Ctrl+S)")
			ca.exportComparison()
		},
		keyboard.ActionExport: func() {
			debug.Println("Export shortcut triggered (Ctrl+E)")
			ca.exportComparison()
		},
		keyboard.ActionReset: func() {
			debug.Println("Reset shortcut triggered (Ctrl+R)")
			ca.resetToDefaults()
		},
		keyboard.ActionPauseResume: func() {
			ca.togglePause()
		},
		keyboard.ActionHelp: func() {
			ca.showKeyboardHelp()
		},
		keyboard.ActionNavigateUp: func() {
			ca.adjustInferences(1000)
		},
		keyboard.ActionNavigateDown: func() {
			ca.adjustInferences(-1000)
		},
		keyboard.ActionNavigateLeft: func() {
			ca.prevWorkload()
		},
		keyboard.ActionNavigateRight: func() {
			ca.nextWorkload()
		},
		keyboard.ActionIncrease: func() {
			ca.adjustInferences(1000)
		},
		keyboard.ActionDecrease: func() {
			ca.adjustInferences(-1000)
		},
	})

	// Add module-specific shortcuts
	km.AddCustomShortcut("next_phase", fyne.KeyN, 0, "Next demo phase")
	km.AddCustomShortcut("prev_phase", fyne.KeyP, 0, "Previous demo phase")

	// Register the manager
	km.Register()

	// Handle additional module-specific keys
	ca.window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		ca.handleKeyPress(ke)
	})
}

// handleKeyPress handles module-specific key events
func (ca *ComparisonApp) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyN:
		// Next demo phase
		ca.nextPhase()

	case fyne.KeyP:
		// Previous demo phase
		ca.prevPhase()

	case fyne.KeySpace:
		// Toggle pause
		ca.togglePause()

	case fyne.KeySlash:
		// Show help
		ca.showKeyboardHelp()
	}
}

// togglePause toggles animation pause state
func (ca *ComparisonApp) togglePause() {
	ca.animMu.Lock()
	ca.paused = !ca.paused
	paused := ca.paused
	ca.animMu.Unlock()

	if paused {
		ca.updateStatus("Paused - Press Space to resume")
	} else {
		ca.updateStatus("Resumed - Press Space to pause")
	}
}

// adjustInferences adjusts the inferences slider
func (ca *ComparisonApp) adjustInferences(delta float64) {
	if ca.inferencesSlider == nil {
		return
	}
	newValue := ca.inferencesSlider.Value + delta
	if newValue < ca.inferencesSlider.Min {
		newValue = ca.inferencesSlider.Min
	}
	if newValue > ca.inferencesSlider.Max {
		newValue = ca.inferencesSlider.Max
	}
	ca.inferencesSlider.SetValue(newValue)
}

// nextWorkload cycles to the next workload
func (ca *ComparisonApp) nextWorkload() {
	if ca.workloadSelect == nil {
		return
	}
	options := ca.workloadSelect.Options
	if len(options) == 0 {
		return
	}
	currentIdx := 0
	for i, opt := range options {
		if opt == ca.workloadSelect.Selected {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(options)
	ca.workloadSelect.SetSelected(options[nextIdx])
}

// prevWorkload cycles to the previous workload
func (ca *ComparisonApp) prevWorkload() {
	if ca.workloadSelect == nil {
		return
	}
	options := ca.workloadSelect.Options
	if len(options) == 0 {
		return
	}
	currentIdx := 0
	for i, opt := range options {
		if opt == ca.workloadSelect.Selected {
			currentIdx = i
			break
		}
	}
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(options) - 1
	}
	ca.workloadSelect.SetSelected(options[prevIdx])
}

// nextPhase advances to the next demo phase
func (ca *ComparisonApp) nextPhase() {
	ca.animMu.Lock()
	ca.currentPhase = (ca.currentPhase + 1) % 4 // Assuming 4 phases
	ca.phaseTimer = 0
	ca.animMu.Unlock()
}

// prevPhase goes back to the previous demo phase
func (ca *ComparisonApp) prevPhase() {
	ca.animMu.Lock()
	if ca.currentPhase == 0 {
		ca.currentPhase = 3 // Wrap around
	} else {
		ca.currentPhase--
	}
	ca.phaseTimer = 0
	ca.animMu.Unlock()
}

// resetToDefaults resets settings to defaults
func (ca *ComparisonApp) resetToDefaults() {
	if ca.workloadSelect != nil {
		ca.workloadSelect.SetSelected("GPT-2")
	}
	if ca.inferencesSlider != nil {
		ca.inferencesSlider.SetValue(10000)
	}
	ca.updateStatus("Reset to defaults")
}

// exportComparison exports the comparison data
func (ca *ComparisonApp) exportComparison() {
	ca.updateStatus("Export: Feature coming soon")
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (ca *ComparisonApp) showKeyboardHelp() {
	helpText := `Keyboard Shortcuts:

Navigation:
  ←/→       Cycle workloads
  ↑/↓       Adjust inference count
  +/=       Increase inferences
  -         Decrease inferences

Simulation:
  Space     Toggle pause/resume
  Ctrl+R    Reset to defaults
  N         Next demo phase
  P         Previous demo phase

Data:
  Ctrl+S    Save/Export comparison
  Ctrl+E    Export comparison

Help:
  / or ?    Show this help dialog

Tips:
• Use arrow keys to explore different workloads
• Watch the energy comparison animations
• Adjust inference count to see scaling effects`

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(380, 340))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, ca.window)
	helpDialog.Show()
}
