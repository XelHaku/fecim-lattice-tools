// Package gui provides info panel creation and management for the hysteresis demo.
package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
)

// createInfoPanel creates the state and material information panel
func (a *App) createInfoPanel() fyne.CanvasObject {
	a.pLabel = widget.NewLabel("0.00")
	a.levelLabel = widget.NewLabel("15/30")
	a.stateLabel = widget.NewLabel("Intermediate")
	a.modeIndicator = widgets.NewModeIndicator()
	a.modeIndicator.SetMinSize(fyne.NewSize(140, 36))

	// Phase indicator for state machine visualization
	a.phaseIndicator = widgets.NewPhaseIndicator()
	a.phaseIndicator.SetMinSize(fyne.NewSize(140, 50))

	// Combined level + polarization row
	levelRow := container.NewHBox(
		widget.NewLabelWithStyle("L:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.levelLabel,
		widget.NewLabel(" "),
		widget.NewLabelWithStyle("P:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.pLabel,
	)

	// Material info button (shows details in dialog)
	matInfoBtn := widget.NewButtonWithIcon("Material Info", theme.InfoIcon(), func() {
		dialog.ShowInformation("Material Properties",
			fmt.Sprintf("Material: %s\n\n"+
				"Pr (Remanent): %.0f µC/cm²\n"+
				"Ps (Saturation): %.0f µC/cm²\n"+
				"Ec (Coercive): %.2f MV/cm\n"+
				"Endurance: %.0e cycles\n\n"+
				"Pr = polarization at E=0 (memory!)\n"+
				"Ec = field needed to switch",
				a.material.Name, a.material.Pr*100, a.material.Ps*100,
				a.material.Ec/1e8, a.material.EnduranceCycles), a.mainWindow)
	})
	matInfoBtn.Importance = widget.LowImportance

	// Compact material line
	matLine := widget.NewLabel(fmt.Sprintf("Pr=%.0f Ec=%.1f End=%.0e",
		a.material.Pr*100, a.material.Ec/1e8, a.material.EnduranceCycles))

	// Wake-up/Fatigue - single compact line
	a.cyclesLabel = widget.NewLabel("0")
	a.wakeupLabel = widget.NewLabel("80%")
	a.fatigueLabel = widget.NewLabel("0%")

	statsRow := container.NewHBox(
		widget.NewLabel("Cyc:"), a.cyclesLabel,
		widget.NewLabel("W:"), a.wakeupLabel,
		widget.NewLabel("F:"), a.fatigueLabel,
	)

	return container.NewVBox(
		levelRow,
		container.NewCenter(a.stateLabel),
		widget.NewSeparator(),
		a.modeIndicator,
		a.phaseIndicator,
		widget.NewSeparator(),
		container.NewHBox(matLine, matInfoBtn),
		statsRow,
	)
}

// createSlidePanel creates the explanation panel
func (a *App) createSlidePanel() fyne.CanvasObject {
	a.slideTitle = widget.NewLabelWithStyle("What's Happening", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	a.slideText = widget.NewLabel(a.getSlideText())
	a.slideText.Wrapping = fyne.TextWrapWord
	return container.NewVBox(
		a.slideTitle,
		a.slideText,
	)
}

// createLogPanel creates the memory operations log panel
func (a *App) createLogPanel() fyne.CanvasObject {
	logTitle := widget.NewLabelWithStyle("Memory Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	a.logText = widget.NewLabel("Waiting...")
	a.logText.Wrapping = fyne.TextWrapWord
	return container.NewVBox(
		logTitle,
		a.logText,
	)
}

// getSlideText returns the contextual explanation text based on current waveform mode
func (a *App) getSlideText() string {
	a.mu.RLock()
	level := a.discreteLevel + 1
	numLevels := a.numLevels
	wrdPhase := a.wrdPhase
	wrdTarget := a.wrdTargetLevel
	isWrite := math.Abs(a.electricField) > a.material.Ec
	waveform := a.waveform
	wrdTotalWrites := a.wrdTotalWrites
	wrdSuccessWrites := a.wrdSuccessWrites
	wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
	animating := a.manualAnimating
	a.mu.RUnlock()

	switch waveform {
	case WaveformManual:
		if animating {
			return fmt.Sprintf("WRITING → L%d\nClick level bar for target", level)
		}
		if isWrite {
			return fmt.Sprintf("WRITING L%d\n|E| > Ec, switching...", level)
		}
		return fmt.Sprintf("HOLD L%d\nE=0, data persists\nClick level bar to write", level)

	case WaveformSine, WaveformTriangle:
		mode := "READ"
		if isWrite {
			mode = "WRITE"
		}
		return fmt.Sprintf("%s L%d/%d\nP-E loop = hysteresis\nSquare shape = sharp switch", mode, level, numLevels)

	case WaveformWriteReadDemo:
		successRate := 0.0
		if wrdTotalWrites > 0 {
			successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
		}

		switch wrdPhase {
		case 0:
			return fmt.Sprintf("WRITE → L%d\n|E| > Ec", wrdTarget)
		case 1:
			return fmt.Sprintf("HOLD L%d\nE=0, P persists", level)
		case 2:
			return fmt.Sprintf("READ L%d\n|E| < Ec, non-destructive", level)
		case 3:
			if level == wrdTarget {
				return fmt.Sprintf("OK L%d\nWrites: %d (%.0f%%)\nEnergy: %.1f pJ", level, wrdTotalWrites, successRate, wrdTotalEnergyfJ/1000)
			}
			return fmt.Sprintf("L%d (want %d)\nWrites: %d (%.0f%%)", level, wrdTarget, wrdTotalWrites, successRate)
		}
		return ""

	default:
		return "Select mode"
	}
}

// addLogEntry adds a timestamped entry to the memory log
func (a *App) addLogEntry(entry string) {
	// Add timestamp prefix
	timestamp := fmt.Sprintf("t=%.1fs", a.simTime)
	fullEntry := fmt.Sprintf("%s %s", timestamp, entry)
	a.logEntries = append(a.logEntries, fullEntry)
	if len(a.logEntries) > a.maxLogLines {
		a.logEntries = a.logEntries[1:]
	}
}

// getLogText returns the formatted log text
func (a *App) getLogText() string {
	if len(a.logEntries) == 0 {
		return "Waiting for operations..."
	}
	result := ""
	for _, e := range a.logEntries {
		result += e + "\n"
	}
	return result
}

// showELI5Dialog displays the "Explain Like I'm 5" hysteresis guide
func (a *App) showELI5Dialog() {
	// Create content with key concepts from the ELI5 guide
	content := widget.NewLabel(
		"HYSTERESIS EXPLAINED LIKE YOU'RE 5\n\n" +
			"🔁 What is Hysteresis?\n" +
			"Like a rubber band that \"remembers\" being stretched.\n" +
			"The path going UP is different from the path coming DOWN.\n\n" +
			"💾 Why It Matters for Memory?\n" +
			"• Regular memory (DRAM): Like a whiteboard - erase & gone\n" +
			"• FeCIM memory: Like carving in clay - stays after power off!\n\n" +
			"📊 The P-E Loop:\n" +
			"• E = Electric Field (the \"push\" you apply)\n" +
			"• P = Polarization (material's response)\n" +
			"• When E = 0, P stays at ±Pr → MEMORY!\n\n" +
			"🎚️ Why 30 Levels?\n" +
			"• Binary: Like a light switch (ON/OFF) = 1 bit\n" +
			"• FeCIM: Like a dimmer with 30 positions = ~5 bits\n" +
			"• Same chip, 5× more storage!\n\n" +
			"📝 Write vs Read:\n" +
			"• WRITE: |E| > Ec → Data changes\n" +
			"• READ: |E| < Ec → Data unchanged, just sense\n" +
			"• HOLD: E = 0 → Data persists (no power!)\n\n" +
			"🎯 The Key Insight:\n" +
			"Hysteresis isn't a bug - it's the FEATURE that\n" +
			"enables memory! The loop REMEMBERS which\n" +
			"way you pushed it.\n\n" +
			"📚 Full Documentation:\n" +
			"See docs/hysteresis/hysteresis.ELI5.md for\n" +
			"detailed explanations with diagrams.")
	content.Wrapping = fyne.TextWrapWord

	// Create scrollable container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 500))

	// Create dialog (declare as var first so button callback can reference it)
	var dialog *widget.PopUp
	closeBtn := widget.NewButton("Got it!", func() {
		if dialog != nil {
			dialog.Hide()
		}
	})

	dialog = widget.NewModalPopUp(
		container.NewVBox(
			container.NewPadded(scroll),
			closeBtn,
		),
		a.mainWindow.Canvas(),
	)

	dialog.Show()
}
