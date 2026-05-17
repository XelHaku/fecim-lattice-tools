//go:build legacy_fyne

package gui

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/keyboard"
)

// setupKeyboard initializes keyboard shortcuts using the shared keyboard manager.
func (ca *CrossbarApp) setupKeyboard() {
	km := keyboard.NewManager(ca.window)

	// Register common handlers
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionSave: func() {
			getDebug().Println("Save shortcut triggered (Ctrl+S)")
			go ca.exportData()
		},
		keyboard.ActionExport: func() {
			getDebug().Println("Export shortcut triggered (Ctrl+E)")
			go ca.exportData()
		},
		keyboard.ActionReset: func() {
			getDebug().Println("Reset shortcut triggered (Ctrl+R)")
			ca.resetArray()
		},
		keyboard.ActionPauseResume: func() {
			ca.toggleAutoDemo()
		},
		keyboard.ActionHelp: func() {
			ca.showKeyboardHelp()
		},
		keyboard.ActionNavigateUp: func() {
			ca.navigateCell(0, -1)
		},
		keyboard.ActionNavigateDown: func() {
			ca.navigateCell(0, 1)
		},
		keyboard.ActionNavigateLeft: func() {
			ca.navigateCell(-1, 0)
		},
		keyboard.ActionNavigateRight: func() {
			ca.navigateCell(1, 0)
		},
		keyboard.ActionIncrease: func() {
			ca.adjustNoise(1)
		},
		keyboard.ActionDecrease: func() {
			ca.adjustNoise(-1)
		},
		keyboard.ActionNextTab: func() {
			ca.nextTab()
		},
		keyboard.ActionPrevTab: func() {
			ca.prevTab()
		},
		keyboard.ActionStepForward: func() {
			ca.runEnhancedMVMInstant()
		},
	})

	// Add module-specific shortcuts
	km.AddCustomShortcut("run_mvm", fyne.KeyM, 0, "Run MVM operation")
	km.AddCustomShortcut("randomize", fyne.KeyN, 0, "Randomize weights")
	km.AddCustomShortcut("toggle_arch", fyne.KeyA, 0, "Toggle architecture")
	km.AddCustomShortcut("inc_temp", fyne.KeyT, 0, "Increase temperature")
	km.AddCustomShortcut("dec_temp", fyne.KeyG, 0, "Decrease temperature")

	// Register the manager
	km.Register()

	// Handle additional module-specific keys
	ca.window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		ca.handleKeyPress(ke)
	})
}

// handleKeyPress handles module-specific key events
func (ca *CrossbarApp) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyM:
		// Run MVM
		ca.runEnhancedMVMInstant()

	case fyne.KeyN:
		// Randomize weights
		ca.resetArray()

	case fyne.KeyA:
		// Toggle architecture
		ca.toggleArchitecture()

	case fyne.KeyT:
		// Increase temperature
		ca.adjustTemperature(25)

	case fyne.KeyG:
		// Decrease temperature
		ca.adjustTemperature(-25)

	case fyne.KeySpace:
		// Toggle auto demo
		ca.toggleAutoDemo()

	case fyne.KeySlash:
		// Show help
		ca.showKeyboardHelp()
	}
}

// toggleAutoDemo toggles the automatic demonstration mode
func (ca *CrossbarApp) toggleAutoDemo() {
	if ca.autoDemo {
		ca.stopAutoDemoLoop()
		ca.updateStatus("Auto demo paused")
	} else {
		ca.startAutoDemoLoop()
		ca.updateStatus("Auto demo started - Press Space to pause")
	}
}

// navigateCell moves cell selection in the heatmap
func (ca *CrossbarApp) navigateCell(dx, dy int) {
	ca.stateMu.Lock()
	defer ca.stateMu.Unlock()

	newRow := ca.selectedRow + dy
	newCol := ca.selectedCol + dx

	// Clamp to valid range
	if newRow < 0 {
		newRow = 0
	}
	if newRow >= ca.config.Rows {
		newRow = ca.config.Rows - 1
	}
	if newCol < 0 {
		newCol = 0
	}
	if newCol >= ca.config.Cols {
		newCol = ca.config.Cols - 1
	}

	ca.selectedRow = newRow
	ca.selectedCol = newCol

	// Trigger cell tap handler to update UI
	fyne.Do(func() {
		ca.onCellTapped(newRow, newCol)
	})
}

// adjustNoise adjusts the noise level
func (ca *CrossbarApp) adjustNoise(delta float64) {
	if ca.noiseSlider == nil {
		return
	}
	newValue := ca.noiseSlider.Value + delta
	if newValue < 0 {
		newValue = 0
	}
	if newValue > 20 {
		newValue = 20
	}
	ca.noiseSlider.SetValue(newValue)
}

// adjustTemperature adjusts the temperature
func (ca *CrossbarApp) adjustTemperature(delta float64) {
	if ca.temperatureSlider == nil {
		return
	}
	newValue := ca.temperatureSlider.Value + delta
	if newValue < 77 {
		newValue = 77
	}
	if newValue > 450 {
		newValue = 450
	}
	ca.temperatureSlider.SetValue(newValue)
}

// toggleArchitecture cycles through architecture options
func (ca *CrossbarApp) toggleArchitecture() {
	ca.stateMu.Lock()
	currentArch := ca.architecture
	ca.stateMu.Unlock()

	// Cycle: PASSIVE -> 1T1R -> 2T1R -> PASSIVE
	switch currentArch {
	case "0T1R (Passive)":
		if ca.arch1T1RBtn != nil {
			ca.arch1T1RBtn.OnTapped()
		}
	case "1T1R (Transistor)":
		if ca.arch2T1RBtn != nil {
			ca.arch2T1RBtn.OnTapped()
		}
	default:
		if ca.archPassiveBtn != nil {
			ca.archPassiveBtn.OnTapped()
		}
	}
}

// nextTab switches to the next tab
func (ca *CrossbarApp) nextTab() {
	keyboard.SelectNextTab(ca.tabs)
}

// prevTab switches to the previous tab
func (ca *CrossbarApp) prevTab() {
	keyboard.SelectPrevTab(ca.tabs)
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (ca *CrossbarApp) showKeyboardHelp() {
	helpText := keyboard.FormatHelpMetadata(keyboard.HelpMetadata{
		Sections: []keyboard.ShortcutSection{
			{Title: "Navigation", Shortcuts: []keyboard.ShortcutMetadata{{Key: "↑/↓/←/→", Description: "Navigate selected cell"}, {Key: "Tab", Description: "Next tab"}, {Key: "Shift+Tab", Description: "Previous tab"}}},
			{Title: "Simulation", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Space", Description: "Toggle auto demo pause/resume"}, {Key: "Ctrl+R", Description: "Reset array with random weights"}, {Key: "M", Description: "Run MVM operation"}, {Key: "N", Description: "Randomize weights"}, {Key: "]", Description: "Step forward (run MVM)"}}},
			{Title: "Data", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+S", Description: "Save/Export data"}, {Key: "Ctrl+E", Description: "Export data"}}},
			{Title: "Configuration", Shortcuts: []keyboard.ShortcutMetadata{{Key: "+/=", Description: "Increase noise level"}, {Key: "-", Description: "Decrease noise level"}, {Key: "T", Description: "Increase temperature (+25K)"}, {Key: "G", Description: "Decrease temperature (-25K)"}, {Key: "A", Description: "Toggle architecture (0T1R/1T1R/2T1R)"}}},
			{Title: "Help", Shortcuts: []keyboard.ShortcutMetadata{{Key: "/ or ?", Description: "Show this help dialog"}}},
		},
		Tips: []string{"Use arrow keys to explore cells in the heatmap", "Watch IR Drop and Sneak Path tabs after running MVM", "Toggle architecture to see sneak path differences"},
	})
	keyboard.ShowHelpTextDialog(ca.window, "Keyboard Shortcuts", helpText, 450, 420)
}
