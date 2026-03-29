package gui

import (
	"fmt"

	"fecim-lattice-tools/shared/keyboard"

	"fyne.io/fyne/v2"
)

// exportShortcut implements fyne.Shortcut for Ctrl+E (export).
type exportShortcut struct{}

func (s *exportShortcut) ShortcutName() string {
	return "Export P-E Data"
}

// setupShortcuts registers custom keyboard shortcuts.
func (a *App) setupShortcuts() {
	km := keyboard.NewManager(a.mainWindow)
	km.SetHandlers(map[keyboard.Action]func(){
		keyboard.ActionExport: func() {
			log.Info("Export shortcut triggered (Ctrl+E)")
			go a.exportPEData()
		},
		keyboard.Action("export_clipboard"): func() {
			go func() {
				if err := a.exportPEDataToClipboard(); err != nil {
					log.Printf("Clipboard export failed: %v", err)
					fyne.Do(func() {
						a.setStatus(fmt.Sprintf("Clipboard export failed: %v", err))
					})
				}
			}()
		},
	})
	km.AddCustomShortcut("export_clipboard", fyne.KeyE, fyne.KeyModifierControl|fyne.KeyModifierShift, "Copy CSV-formatted P-E data to clipboard")
	km.Register()

	// Module 1 has extensive single-key behavior (E/D/T/G/...), keep local dispatch.
	a.mainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		a.handleKeyPress(ke)
	})
}

// handleKeyPress processes keyboard shortcuts
func (a *App) handleKeyPress(ke *fyne.KeyEvent) {
	switch ke.Name {
	case fyne.KeyE:
		// Increase E-field by 0.1*Ec (Manual mode only)
		// Note: Ctrl+E is handled by setupShortcuts() via Canvas.AddShortcut
		if a.waveform == WaveformManual {
			// Read slider value (UI access safe from main thread)
			currentValue := a.eFieldSlider.Value
			newValue := currentValue + 0.1
			if newValue > 2.0 {
				newValue = 2.0
			}
			a.eFieldSlider.SetValue(newValue)
			log.Info("E-field increased to %.2f×Ec", newValue)
		}

	case fyne.KeyD:
		// Decrease E-field by 0.1*Ec (Manual mode only)
		if a.waveform == WaveformManual {
			// Read slider value (UI access safe from main thread)
			currentValue := a.eFieldSlider.Value
			newValue := currentValue - 0.1
			if newValue < -2.0 {
				newValue = -2.0
			}
			a.eFieldSlider.SetValue(newValue)
			log.Info("E-field decreased to %.2f×Ec", newValue)
		}

	case fyne.KeyT:
		// Increase temperature by 25K
		if a.physicsEngine == PhysicsPreisach {
			break
		}
		a.mu.Lock()
		currentTemp := a.currentTemperature()
		newTemp := currentTemp + 25
		if newTemp > 700 {
			newTemp = 700
		}
		a.mu.Unlock()

		// Handle temperature change with calibration (runs in background).
		// Guard: skip if app is shutting down to avoid post-Stop() writes.
		go func() {
			if !a.running.Load() {
				return
			}
			a.mu.Lock()
			a.onTemperatureChanged(newTemp)
			a.mu.Unlock()
		}()
		log.Info("Temperature increased to %.0f K", newTemp)

	case fyne.KeyG:
		// Decrease temperature by 25K
		if a.physicsEngine == PhysicsPreisach {
			break
		}
		a.mu.Lock()
		currentTemp := a.currentTemperature()
		newTemp := currentTemp - 25
		if newTemp < 200 {
			newTemp = 200
		}
		a.mu.Unlock()

		// Handle temperature change with calibration (runs in background).
		// Guard: skip if app is shutting down to avoid post-Stop() writes.
		go func() {
			if !a.running.Load() {
				return
			}
			a.mu.Lock()
			a.onTemperatureChanged(newTemp)
			a.mu.Unlock()
		}()
		log.Info("Temperature decreased to %.0f K", newTemp)

	case fyne.KeyF:
		// Double frequency
		a.mu.Lock()
		newFreq := a.frequency * 2.0
		const minFreq = 1e-9
		if newFreq < minFreq {
			newFreq = minFreq
		}
		a.frequency = newFreq
		// Reset trail when frequency changes
		a.resetHistoryLocked()
		a.simTime = 0
		a.mu.Unlock()
		log.Info("Frequency increased to %.3g Hz", newFreq)

	case fyne.KeyV:
		// Halve frequency
		a.mu.Lock()
		newFreq := a.frequency / 2.0
		const minFreq = 1e-9
		if newFreq < minFreq {
			newFreq = minFreq
		}
		a.frequency = newFreq
		// Reset trail when frequency changes
		a.resetHistoryLocked()
		a.simTime = 0
		a.mu.Unlock()
		log.Info("Frequency decreased to %.3g Hz", newFreq)

	case fyne.KeyW:
		// Cycle to next waveform.
		if keyboard.SelectNextOption(a.waveformSelect) {
			log.Info("Waveform changed to %s", a.waveformSelect.Selected)
		}

	case fyne.KeySpace:
		// Toggle pause
		paused := !a.paused.Load()
		a.paused.Store(paused)
		if paused {
			log.Info("Paused")
			a.pauseBtn.SetText("Resume")
		} else {
			log.Info("Resumed")
			a.pauseBtn.SetText("Pause")
		}

	case fyne.KeyR:
		// Reset simulation
		log.Info("Reset simulation")
		a.mu.Lock()
		a.electricField = 0
		if a.useLKSolver() {
			resetP := a.lkDefaultPolarization()
			if a.lkSolver != nil {
				a.lkSolver.SetState(resetP)
				a.lkSolver.Time = 0
				a.polarization = a.lkSolver.GetState()
			} else {
				a.polarization = resetP
			}
		} else if a.preisach != nil {
			a.preisach.Reset()
			a.polarization = 0
			a.normalizedP = 0
		} else {
			a.polarization = 0
			a.normalizedP = 0
		}
		a.normalizedP = 0
		a.syncDiscreteLevelLocked()
		a.resetHistoryLocked()
		a.simTime = 0
		// Reset Time-Resolved animation state
		a.timeResAnimating = false
		a.timeResIndex = 0
		a.mu.Unlock()
		a.eFieldSlider.SetValue(0)

	case fyne.KeySlash:
		// Show keyboard help dialog (both / and ? key)
		a.showKeyboardHelp()
	}
}

// showKeyboardHelp displays a dialog with all keyboard shortcuts
func (a *App) showKeyboardHelp() {
	resume := a.pauseSimulationForModal()
	helpText := keyboard.FormatHelpMetadata(keyboard.HelpMetadata{
		Sections: []keyboard.ShortcutSection{
			{Title: "E-Field Control (Manual Mode Only)", Shortcuts: []keyboard.ShortcutMetadata{{Key: "E", Description: "Increase E-field by 0.1×Ec"}, {Key: "D", Description: "Decrease E-field by 0.1×Ec"}}},
			{Title: "Temperature Control", Shortcuts: []keyboard.ShortcutMetadata{{Key: "T", Description: "Increase temperature by 25K"}, {Key: "G", Description: "Decrease temperature by 25K"}}},
			{Title: "Frequency Control", Shortcuts: []keyboard.ShortcutMetadata{{Key: "F", Description: "Double frequency"}, {Key: "V", Description: "Halve frequency (min 1e-9 Hz)"}}},
			{Title: "Waveform & Simulation", Shortcuts: []keyboard.ShortcutMetadata{{Key: "W", Description: "Cycle to next waveform"}, {Key: "Space", Description: "Toggle pause/resume"}, {Key: "R", Description: "Reset simulation"}}},
			{Title: "Data Export", Shortcuts: []keyboard.ShortcutMetadata{{Key: "Ctrl+E", Description: "Export P-E data to JSON and CSV"}, {Key: "Ctrl+Shift+E", Description: "Copy CSV-formatted P-E data to clipboard"}}},
			{Title: "Help", Shortcuts: []keyboard.ShortcutMetadata{{Key: "/ or ?", Description: "Show this help dialog"}}},
		},
		Tips: []string{
			"E and D keys only work in Manual mode.",
			"Switch to Manual mode using the waveform selector.",
			"Exported files are saved to the data/ directory with timestamps.",
		},
	})

	keyboard.ShowHelpTextDialogWithCallback(a.mainWindow, "Keyboard Shortcuts", helpText, 500, 400, resume)
}
