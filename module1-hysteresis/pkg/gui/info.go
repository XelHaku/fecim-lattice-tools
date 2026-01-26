// Package gui provides info panel creation and management for the hysteresis demo.
package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/gui/widgets"
)

// createInfoPanel creates the state and material information panel
func (a *App) createInfoPanel() fyne.CanvasObject {
	a.pLabel = widget.NewLabel("0.00 µC/cm²")
	a.levelLabel = widget.NewLabel("15/30")
	a.stateLabel = widget.NewLabel("Intermediate")
	a.modeIndicator = widgets.NewModeIndicator()
	a.modeIndicator.SetMinSize(fyne.NewSize(180, 50))

	// State display - horizontal layout for compactness
	pRow := container.NewHBox(
		widget.NewLabelWithStyle("P:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.pLabel,
	)
	levelRow := container.NewHBox(
		widget.NewLabelWithStyle("Level:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.levelLabel,
	)
	a.stateLabel.Alignment = fyne.TextAlignCenter

	// Material params - compact grid
	matParamsLabel := widget.NewLabelWithStyle("Material", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	matParams := widget.NewLabel(fmt.Sprintf(
		"Pr=%.0f Ps=%.0f Ec=%.2f",
		a.material.Pr*100, a.material.Ps*100,
		a.material.Ec/1e8,
	))
	matParams.Wrapping = fyne.TextWrapWord

	enduranceLabel := widget.NewLabel(fmt.Sprintf("Endurance: %.0e", a.material.EnduranceCycles))
	enduranceLabel.Wrapping = fyne.TextWrapWord

	// Wake-up/Fatigue display - compact
	a.cyclesLabel = widget.NewLabel("0")
	a.wakeupLabel = widget.NewLabel("80%")
	a.fatigueLabel = widget.NewLabel("0.0%")

	cyclingLabel := widget.NewLabelWithStyle("Cycling Stats", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fatigueRow := container.NewHBox(
		widget.NewLabel("Cyc:"), a.cyclesLabel,
		widget.NewLabel("Wake:"), a.wakeupLabel,
		widget.NewLabel("Fat:"), a.fatigueLabel,
	)

	// Divider
	divider := widget.NewSeparator()

	return container.NewVBox(
		levelRow,
		container.NewCenter(a.stateLabel),
		divider,
		pRow,
		divider,
		a.modeIndicator,
		divider,
		matParamsLabel,
		matParams,
		enduranceLabel,
		divider,
		cyclingLabel,
		fatigueRow,
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
	level := a.discreteLevel + 1 // 1-indexed for display
	wrdPhase := a.wrdPhase
	wrdTarget := a.wrdTargetLevel
	isWrite := math.Abs(a.electricField) > a.material.Ec
	waveform := a.waveform
	wrdTotalWrites := a.wrdTotalWrites
	wrdSuccessWrites := a.wrdSuccessWrites
	wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
	a.mu.RUnlock()

	switch waveform {
	case WaveformManual:
		a.mu.RLock()
		animating := a.manualAnimating
		a.mu.RUnlock()

		if animating {
			return fmt.Sprintf("WRITING → L%d\n\n"+
				"Applying E-field:\n"+
				"• Higher target → +E\n"+
				"• Lower target → -E\n\n"+
				"Watch the P-E plot!\n"+
				"Click level bar for new target.", level)
		}
		if isWrite {
			return fmt.Sprintf("██ WRITING LEVEL %d ██\n\n"+
				"Electric field E > Ec.\n"+
				"Domains are switching.\n"+
				"Polarization is changing.\n\n"+
				"Use slider OR click\n"+
				"level bar to program!", level)
		}
		return fmt.Sprintf("░░ HOLDING LEVEL %d ░░\n\n"+
			"E-field is low or zero.\n"+
			"Polarization PERSISTS.\n"+
			"No power needed.\n\n"+
			"MANUAL MODE:\n"+
			"• Drag slider to apply E-field\n"+
			"• Click level bar to auto-program", level)

	case WaveformSine, WaveformTriangle:
		phaseText := "░░ READING ░░"
		if isWrite {
			phaseText = "██ WRITING ██"
		}
		return fmt.Sprintf("%s\n\n"+
			"Level: %d/30\n\n"+
			"The P-E loop shows hysteresis:\n"+
			"• Upper branch: E increasing\n"+
			"• Lower branch: E decreasing\n"+
			"• Area inside = energy loss\n\n"+
			"The SQUARE shape means:\n"+
			"sharp switching at ±Ec.", phaseText, level)

	case WaveformWriteReadDemo:
		// Calculate stats (using local copies from RLock above)
		successRate := 0.0
		if wrdTotalWrites > 0 {
			successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
		}
		energyPerOp := 10.0 // ~10 fJ per operation (FeFET switching energy)

		var phaseExplanation string
		switch wrdPhase {
		case 0: // WRITE
			direction := "+E (positive)"
			if wrdTarget < level {
				direction = "-E (negative)"
			}
			phaseExplanation = fmt.Sprintf("▓▓ WRITE → L%d ▓▓\n\n"+
				"Applying %s\n"+
				"|E| > Ec to switch domains.\n\n"+
				"Higher level → +E field\n"+
				"Lower level → -E field\n\n"+
				"Energy: ~%.0f fJ\n"+
				"(10M× less than NAND!)", wrdTarget, direction, energyPerOp)
		case 1: // HOLD
			phaseExplanation = fmt.Sprintf("░░ HOLD L%d ░░\n\n"+
				"E = 0, P persists!\n\n"+
				"ZERO POWER NEEDED.\n"+
				"Data retention: 10+ years\n\n"+
				"This is TRUE non-volatile:\n"+
				"No refresh like DRAM.\n"+
				"No charge leakage.\n\n"+
				"30 levels = 4.9 bits/cell", level)
		case 2: // READ
			phaseExplanation = fmt.Sprintf("▒▒ READ L%d ▒▒\n\n"+
				"Sense pulse: |E| < Ec\n"+
				"State UNCHANGED!\n\n"+
				"Non-destructive read:\n"+
				"Unlike NAND, data stays.\n"+
				"No rewrite needed.\n\n"+
				"Read energy: ~%.0f fJ", level, energyPerOp*0.1)
		case 3: // DISPLAY
			status := "✓ SUCCESS"
			accuracy := ""
			if level != wrdTarget {
				status = fmt.Sprintf("△ L%d (want %d)", level, wrdTarget)
				accuracy = "\n(±1 is normal)"
			} else {
				accuracy = "\nPerfect!"
			}
			phaseExplanation = fmt.Sprintf("%s%s\n\n"+
				"Writes: %d | %.0f%%\n"+
				"Energy: %.1f pJ total\n\n"+
				"Next target coming...", status, accuracy,
				wrdTotalWrites, successRate, wrdTotalEnergyfJ/1000)
		}
		return phaseExplanation

	default:
		return "Select a waveform mode\nto see explanation."
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
