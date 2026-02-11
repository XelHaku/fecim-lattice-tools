package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// exportShortcut implements fyne.Shortcut for Ctrl+E (export).
type exportShortcut struct{}

func (s *exportShortcut) ShortcutName() string {
	return "Export P-E Data"
}

// setupShortcuts registers custom keyboard shortcuts
func (a *App) setupShortcuts() {
	// Register Ctrl+E for export
	ctrlE := &desktop.CustomShortcut{
		KeyName:  fyne.KeyE,
		Modifier: fyne.KeyModifierControl,
	}

	a.mainWindow.Canvas().AddShortcut(ctrlE, func(shortcut fyne.Shortcut) {
		log.Info("Export shortcut triggered (Ctrl+E)")
		// Run export in background to avoid blocking UI
		go a.exportPEData()
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

		// Handle temperature change with calibration (runs in background)
		go func() {
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

		// Handle temperature change with calibration (runs in background)
		go func() {
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
		// Cycle to next waveform (5 modes: Manual, Sine, Triangle, ISPP (Write/Read), Time-Resolved)
		a.mu.Lock()
		nextWaveform := (a.waveform + 1) % 5
		a.mu.Unlock()

		// Trigger the same logic as waveformSelect.OnChanged
		waveformNames := []string{"Manual", "Sine Wave", "Triangle Wave", "ISPP (Write/Read)", "Time-Resolved Switching"}
		selectedName := waveformNames[nextWaveform]

		// Call the select widget to trigger the OnChanged callback
		a.waveformSelect.SetSelected(selectedName)
		log.Info("Waveform changed to %s", selectedName)

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
	helpText := `Keyboard Shortcuts:

E-Field Control (Manual Mode Only):
  E         Increase E-field by 0.1×Ec
  D         Decrease E-field by 0.1×Ec

Temperature Control:
  T         Increase temperature by 25K
  G         Decrease temperature by 25K

Frequency Control:
  F         Double frequency
  V         Halve frequency (min 1e-9 Hz)

Waveform & Simulation:
  W         Cycle to next waveform
  Space     Toggle pause/resume
  R         Reset simulation

Data Export:
  Ctrl+E    Export P-E data to JSON and CSV

Help:
  / or ?    Show this help dialog

Note: E and D keys only work in Manual mode.
Switch to Manual mode using the waveform selector.
Exported files are saved to the data/ directory with timestamps.`

	// Create a scrollable label for the help text
	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(500, 400))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, a.mainWindow)
	helpDialog.Show()
}
