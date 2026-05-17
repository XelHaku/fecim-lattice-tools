//go:build legacy_fyne

package gui

import (
	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/keyboard"
)

// setupKeyboard initializes keyboard shortcuts using the shared keyboard manager.
func (ma *MNISTApp) setupKeyboard() {
	km := keyboard.NewManager(ma.window)

	// Register common handlers
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionSave: func() {
			debug.Println("Save shortcut triggered (Ctrl+S)")
			// Future: Save network weights
			ma.updateStatus("Save: Feature not yet implemented")
		},
		keyboard.ActionExport: func() {
			debug.Println("Export shortcut triggered (Ctrl+E)")
			// Future: Export evaluation results
			ma.updateStatus("Export: Feature not yet implemented")
		},
		keyboard.ActionReset: func() {
			debug.Println("Reset shortcut triggered (Ctrl+R)")
			ma.clearCanvas()
		},
		keyboard.ActionPauseResume: func() {
			ma.toggleAutoDemo()
		},
		keyboard.ActionHelp: func() {
			ma.showKeyboardHelp()
		},
		keyboard.ActionNavigateLeft: func() {
			ma.loadRandomTestDigit()
		},
		keyboard.ActionNavigateRight: func() {
			ma.loadRandomTestDigit()
		},
		keyboard.ActionStepForward: func() {
			ma.loadRandomTestDigit()
		},
	})

	// Add module-specific shortcuts
	km.AddCustomShortcut("clear", fyne.KeyC, 0, "Clear canvas")
	km.AddCustomShortcut("random", fyne.KeyR, 0, "Load random test digit")
	km.AddCustomShortcut("evaluate", fyne.KeyE, 0, "Evaluate network")
	km.AddCustomShortcut("load", fyne.KeyL, 0, "Load test data")

	// Register the manager
	km.Register()

	// Handle additional module-specific keys
	ma.window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		ma.handleKeyPress(ke)
	})
}

// handleKeyPress handles module-specific key events
func (ma *MNISTApp) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyC:
		// Clear canvas
		ma.clearCanvas()

	case fyne.KeyR:
		// Load random test digit
		ma.loadRandomTestDigit()

	case fyne.KeyE:
		// Evaluate network (without Ctrl modifier)
		go ma.evaluateNetwork()

	case fyne.KeyL:
		// Load test data
		go ma.loadTestData()

	case fyne.KeySpace:
		// Toggle auto demo
		ma.toggleAutoDemo()

	case fyne.KeySlash:
		// Show help
		ma.showKeyboardHelp()
	}
}

// clearCanvas clears the digit canvas and resets predictions
func (ma *MNISTApp) clearCanvas() {
	if ma.digitCanvas != nil {
		ma.digitCanvas.Clear()
	}
	if ma.predictionDisplay != nil {
		ma.predictionDisplay.SetPrediction(-1, 0)
	}
	if ma.operationLog != nil {
		ma.operationLog.Add("Canvas cleared")
	}
	ma.updateStatus("● IDLE | Canvas cleared")
}

// toggleAutoDemo toggles the automatic demonstration mode
func (ma *MNISTApp) toggleAutoDemo() {
	ma.autoDemoMu.Lock()
	isRunning := ma.autoDemo
	ma.autoDemoMu.Unlock()

	if isRunning {
		ma.stopAutoDemoLoop()
		if ma.demoModeSelect != nil {
			ma.demoModeSelect.SetSelected("Manual")
		}
		ma.updateStatus("Auto demo paused - Press Space to resume")
	} else {
		if ma.demoModeSelect != nil {
			ma.demoModeSelect.SetSelected("Auto Demo")
		}
		ma.updateStatus("Auto demo started - Press Space to pause")
	}
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (ma *MNISTApp) showKeyboardHelp() {
	helpText := keyboard.FormatHelpMetadata(keyboard.HelpMetadata{
		Sections: []keyboard.ShortcutSection{
			{Title: "Navigation", Shortcuts: []keyboard.ShortcutMetadata{{Key: "←/→", Description: "Load random test digit"}, {Key: "]", Description: "Step forward (load random digit)"}}},
			{Title: "Simulation", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Space", Description: "Toggle auto demo pause/resume"}, {Key: "Ctrl+R", Description: "Clear canvas"}}},
			{Title: "Data", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+S", Description: "Save (future)"}, {Key: "Ctrl+E", Description: "Export (future)"}, {Key: "C", Description: "Clear canvas"}, {Key: "R", Description: "Load random test digit"}, {Key: "E", Description: "Evaluate network on test data"}, {Key: "L", Description: "Load test data"}}},
			{Title: "Demo Mode", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Space", Description: "Toggle auto demo"}}},
			{Title: "Help", Shortcuts: []keyboard.ShortcutMetadata{{Key: "/ or ?", Description: "Show this help dialog"}}},
		},
		Tips: []string{"Draw a digit on the canvas to see predictions", "Use R to load random MNIST test samples", "Press E to run full evaluation on test set", "Watch the network activations in the center panel"},
	})
	keyboard.ShowHelpTextDialog(ma.window, "Keyboard Shortcuts", helpText, 400, 380)
}
