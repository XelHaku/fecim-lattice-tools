// Package gui provides info panel creation and management for the hysteresis demo.
package gui

import (
	"fmt"
	"math"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// createInfoPanel creates the device status panel with state, material, metrics and cycle health.
func (a *App) createInfoPanel() fyne.CanvasObject {
	a.pLabel = widget.NewLabel("0.00 µC/cm²")
	a.levelLabel = widget.NewLabel(fmt.Sprintf("%d/%d", a.numLevels/2, a.numLevels))
	// Bold state label — most prominent element in the panel
	a.stateLabel = widget.NewLabelWithStyle("Intermediate", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Phase indicator: PROGRAM | VERIFY | RESULT
	a.phaseIndicator = widgets.NewPhaseIndicator()
	a.phaseIndicator.SetMinSize(fyne.NewSize(140, 50))

	// State stability indicator (M12): green=stable, yellow=moderate, red=edge
	a.stabilityIndicator = widgets.NewStabilityIndicator()
	a.stabilityIndicator.SetLevel(15, a.numLevels)

	// Level + polarization row
	levelRow := container.NewHBox(
		widget.NewLabelWithStyle("L:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.levelLabel,
		widget.NewLabel("  "),
		widget.NewLabelWithStyle("P:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.pLabel,
	)

	// ── MATERIAL section ─────────────────────────────────────────────────────

	// Icon-only info button — opens material details dialog
	matInfoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		resume := a.pauseSimulationForModal()
		d := dialog.NewInformation("Material Properties",
			fmt.Sprintf("Material: %s\n\n"+
				"Pr (Remanent): %.0f µC/cm² [literature: %.0f–%.0f]\n"+
				"Ps (Saturation): %.0f µC/cm² [±10%%]\n"+
				"Ec (Coercive): %.2f MV/cm [literature: %.2f–%.2f]\n"+
				"Endurance: %.0e cycles [demonstrated: 10⁹-10¹²]\n\n"+
				"Pr = polarization at E=0 (memory!)\n"+
				"Ec = field needed to switch\n\n"+
				"Note: Ranges from peer-reviewed literature.\n"+
				"Actual values depend on process conditions.",
				a.material.Name, a.material.Pr*100, a.material.Pr*80, a.material.Pr*120, a.material.Ps*100,
				a.material.Ec/1e8, a.material.Ec*0.7/1e8, a.material.Ec*1.3/1e8, a.material.EnduranceCycles),
			a.mainWindow,
		)
		d.SetOnClosed(resume)
		d.Show()
	})
	matInfoBtn.Importance = widget.LowImportance

	// "MATERIAL" label left, info button right
	materialHeader := container.NewBorder(
		nil, nil,
		widget.NewLabelWithStyle("MATERIAL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		matInfoBtn,
		nil,
	)

	// Material name — prominent, updated on material change
	a.materialNameLabel = widget.NewLabelWithStyle(a.material.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	a.materialNameLabel.Truncation = fyne.TextTruncateEllipsis

	// Confidence badge in a swappable container (replaced on material change)
	confidence := sharedwidgets.Estimated
	if strings.Contains(strings.ToLower(a.material.Name), "materlik") || strings.Contains(strings.ToLower(a.material.Name), "fecim hzo") {
		confidence = sharedwidgets.Calibrated
	}
	a.materialBadgeBox = container.NewHBox(sharedwidgets.NewConfidenceBadge(confidence))

	// Pr / Ec summary — updated on material change
	a.materialPropsLabel = widget.NewLabel(fmt.Sprintf("Pr %.1f  Ec %.2f MV/cm",
		a.material.Pr*100, a.material.Ec/1e8))
	a.materialPropsLabel.Truncation = fyne.TextTruncateEllipsis

	// ── METRICS section ──────────────────────────────────────────────────────

	ecVal := a.material.Ec / 1e8
	prVal := a.material.Pr * 100
	sqVal := 0.0
	if a.material.Ps > 0 {
		sqVal = a.material.Pr / a.material.Ps
	}
	a.effEcLabel = widget.NewLabel(fmt.Sprintf("Ec(T): %.2f±%.2f MV/cm", ecVal, ecVal*0.15))
	a.effEcLabel.Truncation = fyne.TextTruncateEllipsis
	a.effPrLabel = widget.NewLabel(fmt.Sprintf("Pr(T): %.1f±%.1f µC/cm²", prVal, prVal*0.20))
	a.effPrLabel.Truncation = fyne.TextTruncateEllipsis
	a.squarenessLabel = widget.NewLabel(fmt.Sprintf("Sq: %.2f", sqVal))
	a.squarenessLabel.Truncation = fyne.TextTruncateEllipsis
	a.switchedLabel = widget.NewLabel("Sw: 0%")
	a.switchedLabel.Truncation = fyne.TextTruncateEllipsis

	metricsGrid := container.NewGridWithColumns(2,
		a.effEcLabel,
		a.effPrLabel,
		a.squarenessLabel,
		a.switchedLabel,
	)

	// ── CYCLE HEALTH section ─────────────────────────────────────────────────

	a.cyclesLabel = widget.NewLabel("0")
	a.wakeupLabel = widget.NewLabel("100%")
	a.fatigueLabel = widget.NewLabel("0%")
	// L01: Cycle phase indicator (WAKE-UP, STABLE, FATIGUE)
	a.cyclePhaseLabel = widget.NewLabel("WAKE-UP")
	a.cyclePhaseLabel.TextStyle = fyne.TextStyle{Bold: true}

	cycleRow1 := container.NewHBox(
		widget.NewLabel("Cyc:"), a.cyclesLabel,
		widget.NewLabel("  "),
		a.cyclePhaseLabel,
	)
	cycleRow2 := container.NewHBox(
		widget.NewLabel("W:"), a.wakeupLabel,
		widget.NewLabel("  F:"), a.fatigueLabel,
	)

	return container.NewVBox(
		// — STATE —
		a.stateLabel,
		levelRow,
		a.stabilityIndicator,
		a.phaseIndicator,
		widget.NewSeparator(),
		// — MATERIAL —
		materialHeader,
		a.materialNameLabel,
		a.materialBadgeBox,
		a.materialPropsLabel,
		widget.NewSeparator(),
		// — METRICS —
		metricsGrid,
		widget.NewSeparator(),
		// — CYCLE HEALTH —
		cycleRow1,
		cycleRow2,
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

// createLogPanel creates the memory operations log panel with toggle (M07)
func (a *App) createLogPanel() fyne.CanvasObject {
	logTitle := widget.NewLabelWithStyle("Memory Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	a.logText = widget.NewLabel("Waiting...")
	a.logText.Wrapping = fyne.TextWrapWord

	// M07: Toggle button for verbose/compact log view
	a.logVerbose = true // Default to verbose
	a.logToggleBtn = widget.NewButton("Compact", func() {
		a.mu.Lock()
		a.logVerbose = !a.logVerbose
		if a.logVerbose {
			a.logToggleBtn.SetText("Compact")
		} else {
			a.logToggleBtn.SetText("Verbose")
		}
		a.mu.Unlock()
	})
	a.logToggleBtn.Importance = widget.LowImportance

	headerRow := container.NewBorder(nil, nil, logTitle, a.logToggleBtn, nil)
	logScroll := container.NewVScroll(a.logText)
	logScroll.SetMinSize(fyne.NewSize(0, 140))

	return container.NewVBox(
		headerRow,
		logScroll,
	)
}

// getSlideText returns the contextual explanation text based on current waveform mode
func (a *App) getSlideText() string {
	a.mu.RLock()
	level := a.discreteLevel + 1
	numLevels := a.numLevels
	wrdTarget := a.wrdTargetLevel
	ctrlState := controller.StateIdle
	if a.writeController != nil {
		ctrlState = a.writeController.State
	}
	lastPhase := a.lastLogPhase
	const wrdPhaseVerify = 1
	eField := a.electricField
	matEc := 0.0
	if a.material != nil {
		matEc = a.material.Ec
	}
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

		// Get ISPP stats for phase 3 display
		var avgPulses, failRate float64
		if a.isppWidget != nil {
			stats := a.isppWidget.GetStats()
			avgPulses = stats.GetAveragePulses()
			failRate = stats.GetFailureRate() * 100 // Convert to percentage
		}

		eFieldSettled := matEc > 0 && math.Abs(eField) < 0.01*matEc
		atTarget := level == wrdTarget
		settled := atTarget && eFieldSettled
		phase := "PROGRAM"
		if settled {
			phase = "RESULT"
		} else if ctrlState == controller.StateVerify || ctrlState == controller.StateSuccess || lastPhase == wrdPhaseVerify {
			phase = "VERIFY"
		}

		switch phase {
		case "PROGRAM":
			return fmt.Sprintf("PROGRAM L%d\nISPP pulses toward target", wrdTarget)
		case "VERIFY":
			return fmt.Sprintf("VERIFY L%d\nE=0 check", wrdTarget)
		case "RESULT":
			if level == wrdTarget {
				return fmt.Sprintf("RESULT L%d ✓\nWrites: %d (%.0f%%)\nEnergy: %.1f pJ\nAvg: %.1f pulses | Fail: %.2f%%",
					level, wrdTotalWrites, successRate, wrdTotalEnergyfJ/1000, avgPulses, failRate)
			}
			return fmt.Sprintf("RESULT L%d (want %d)\nWrites: %d (%.0f%%)\nAvg: %.1f pulses | Fail: %.2f%%",
				level, wrdTarget, wrdTotalWrites, successRate, avgPulses, failRate)
		}
		return ""

	case WaveformTimeResolved:
		a.mu.RLock()
		animating := a.timeResAnimating
		idx := a.timeResIndex
		dataLen := len(a.timeResDataTimes)
		a.mu.RUnlock()

		if animating && dataLen > 0 && idx < dataLen {
			a.mu.RLock()
			currentTime := a.timeResDataTimes[idx]
			switchedCount := a.timeResDataSwitch[idx]
			totalHysterons := len(a.timeResDataSwitch)
			a.mu.RUnlock()

			switchedFrac := float64(switchedCount) / float64(totalHysterons) * 100
			tau := a.material.Tau
			return fmt.Sprintf("KAI SWITCHING L%d\nt = %.1f ns\n%.0f%% switched\nτ = %.1f ns\nP(t)=Ps(1-e^(-(t/τ)²))",
				level, currentTime*1e9, switchedFrac, tau*1e9)
		}
		return "TIME-RESOLVED SWITCHING\nKAI dynamics\nNanosecond switching"

	default:
		return "Select mode"
	}
}

// addLogEntry adds a timestamped entry to the memory log
// NOTE: Caller must hold a.mu lock to prevent race conditions or deadlocks
func (a *App) addLogEntry(entry string) {
	// Add timestamp prefix
	timestamp := fmt.Sprintf("t=%.1fs", a.simTime)
	fullEntry := fmt.Sprintf("%s %s", timestamp, entry)
	a.logEntries = append(a.logEntries, fullEntry)
	if len(a.logEntries) > a.maxLogLines {
		a.logEntries = a.logEntries[1:]
	}
}

func formatLogEntries(logEntries []string, logVerbose bool) string {
	if len(logEntries) == 0 {
		return "Waiting for operations..."
	}

	// M07: Compact mode shows only last 3 entries without decorative lines
	if !logVerbose {
		compactEntries := logEntries
		// Skip header decorations (lines starting with ═ or ─ or spaces)
		filtered := make([]string, 0, len(compactEntries))
		for _, e := range compactEntries {
			if len(e) == 0 {
				continue
			}
			// Get first rune for unicode-safe comparison
			firstRune := []rune(e)[0]
			if firstRune != '═' && firstRune != '─' && firstRune != ' ' {
				// Shorten timestamp format for compact view
				if len(e) > 7 && e[:2] == "t=" {
					// Extract just the operation part after timestamp
					for i := 0; i < len(e); i++ {
						if e[i] == ' ' && i < len(e)-1 {
							filtered = append(filtered, e[i+1:])
							break
						}
					}
				} else {
					filtered = append(filtered, e)
				}
			}
		}
		// Show only last 3
		if len(filtered) > 3 {
			filtered = filtered[len(filtered)-3:]
		}
		result := ""
		for _, e := range filtered {
			result += e + "\n"
		}
		if result == "" {
			return "Ready"
		}
		return result
	}

	// Verbose mode: show everything
	result := ""
	for _, e := range logEntries {
		result += e + "\n"
	}
	return result
}

// getLogText returns the formatted log text (M07: supports verbose/compact modes)
func (a *App) getLogText() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return formatLogEntries(a.logEntries, a.logVerbose)
}

// showELI5Dialog displays the "Explain Like I'm 5" hysteresis guide
func (a *App) showELI5Dialog() {
	resume := a.pauseSimulationForModal()
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
			"🎚️ Why 30 Levels (Demo Baseline)?\n" +
			"• Binary: Like a light switch (ON/OFF) = 1 bit\n" +
			"• FeCIM: Like a dimmer with 30 positions = ~5 bits (conference claim)\n" +
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
		resume()
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
