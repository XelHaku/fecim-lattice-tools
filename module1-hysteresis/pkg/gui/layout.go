package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

func (a *App) createUI() fyne.CanvasObject {
	// SAFETY: No mutex needed - createUI() called from run() before simulation goroutine starts (line 319).
	// No concurrent access to a.material exists during UI initialization.

	// Create cell visualizer (THE memory cell!) - larger for better visibility
	a.cellViz = widgets.NewCellVisualizer()
	a.cellViz.SetMinSize(fyne.NewSize(160, 180)) // Slightly smaller for responsive layouts

	// Create P-E plot - will expand to fill space
	// Use engine-specific Ec and nominal Pr for initial plot setup
	effEc := a.effectiveEc()
	// Use material's nominal Pr (not GetEffectivePr which recalculates from current state)
	effPr := a.material.Pr
	a.plot = widgets.NewPEPlot(effEc*2.5, effPr*1.2, ColorBackground, ColorGrid, ColorAxis, ColorPositive, ColorNegative, ColorWarning)
	a.plot.SetMinSize(fyne.NewSize(360, 300))
	a.plot.SetMaterialParams(effEc, effPr)

	// Create level indicator (wider for better labels, clickable in Manual mode)
	a.levelIndicator = widgets.NewLevelIndicator()
	a.levelIndicator.SetMinSize(fyne.NewSize(70, 300))
	// Wire up click callback for Manual mode
	a.levelIndicator.OnLevelClicked = func(targetLevel int) {
		a.mu.Lock()
		defer a.mu.Unlock()

		// Log click with detailed state (always log for debugging)
		log.Printf("LEVEL CLICK: target=%d, currentDiscrete=%d, normalizedP=%.4f, waveform=%v, animating=%v",
			targetLevel, a.discreteLevel, a.normalizedP, a.waveform, a.manualAnimating)

		if a.waveform == WaveformManual && !a.manualAnimating {
			currentLevel := a.discreteLevel + 1
			if targetLevel == currentLevel {
				a.addLogEntry(fmt.Sprintf("ALREADY → Level %d", targetLevel))
				return
			}
			// Reset WriteController for clean ISPP sequence
			if a.writeController != nil {
				a.writeController.ResetState()
			}
			// Start ISPP animation to target level
			a.manualTargetLevel = targetLevel
			a.manualStartLevel = currentLevel // Capture starting level (1-indexed)
			a.manualAnimating = true
			a.manualPhase = 0 // Phase 0 = PREP (or skip to WRITE for Preisach in sim_loop)
			a.manualPhaseTime = 0

			log.Printf("MANUAL ISPP START: target=%d, start=%d, numLevels=%d",
				targetLevel, a.manualStartLevel, a.numLevels)
			a.addLogEntry(fmt.Sprintf("WRITE → Level %d", targetLevel))
			return
		}

		if a.waveform == WaveformWriteReadDemo {
			// Queue a manual target for the next ISPP cycle.
			a.wrdNextTargetLevel = targetLevel
			a.addLogEntry(fmt.Sprintf("NEXT TARGET → Level %d", targetLevel))
			log.Printf("WRD MANUAL TARGET: queued=%d currentLevel=%d phase=%d", targetLevel, a.discreteLevel+1, a.wrdPhase)
		}
	}
	a.updateLevelIndicatorRange()

	// Create controls panel
	controls := a.createControlsPanel()
	controlsContent := container.New(&fixedMinWidthLayout{minWidth: 260}, controls)
	controlsScroll := container.NewVScroll(controlsContent)
	controlsScroll.SetMinSize(fyne.NewSize(260, 0))

	// Create info panel
	info := a.createInfoPanel()

	// Create log panel
	logPanel := a.createLogPanel()

	// Create simulation vs experiment comparison widget (H16)
	a.simVsExpWidget = widgets.NewSimVsExpComparison()
	a.updateSimVsExpWidget()

	// Create ISPP visualization widget (H14)
	a.isppWidget = widgets.NewISPPVisualization()

	// Cell title with underline effect - more prominent
	cellTitle := canvas.NewText("MEMORY CELL", color.RGBA{0, 212, 255, 255})
	cellTitle.TextSize = 18 // Increased from 16
	cellTitle.TextStyle = fyne.TextStyle{Bold: true}
	cellTitle.Alignment = fyne.TextAlignCenter

	cellUnderline := canvas.NewRectangle(color.RGBA{0, 212, 255, 200}) // More opaque
	cellUnderline.SetMinSize(fyne.NewSize(140, 3))                     // Wider and thicker

	cellHeader := container.NewVBox(
		container.NewCenter(cellTitle),
		container.NewCenter(cellUnderline),
	)

	// Fixed header: Cell visualization and basic level display
	cellDisplay := container.NewVBox(
		cellHeader,
		container.NewCenter(a.cellViz),
	)

	infoCard := widget.NewCard("Device Status", "", info)
	literatureCard := a.createLiteratureOverlayPanel()
	infoStack := container.NewVBox(
		infoCard,
		a.isppWidget,
		a.simVsExpWidget,
		literatureCard,
	)
	infoScroll := container.NewVScroll(infoStack)
	infoScroll.SetMinSize(fyne.NewSize(220, 0))

	leftSplit := container.NewVSplit(infoScroll, logPanel)
	leftSplit.SetOffset(0.66) // 66% info/status panel, 34% log panel

	// Left column: Fixed cell at top, scrollable info below
	leftColumn := container.NewBorder(
		cellDisplay,
		nil, nil, nil,
		container.NewPadded(leftSplit),
	)

	// Right column: Controls only (compact)
	rightColumn := container.NewBorder(nil, nil, nil, nil, controlsScroll)

	// Plot + level in same row using Border layout
	// Border allows level indicator to be tappable (HSplit may intercept events)
	plotAndLevel := container.NewBorder(
		nil, nil, nil,
		a.levelIndicator,
		a.plot,
	)
	plotPanel := container.NewPadded(plotAndLevel)

	// Status bar at bottom
	a.statusLabel = widget.NewLabel("Running...")
	statusBar := container.NewHBox(
		layout.NewSpacer(),
		a.statusLabel,
		layout.NewSpacer(),
	)

	// Adaptive layout: Left | Center | Right
	// Zones: [0]=Left (Cell+Info), [1]=Plot+Level, [2]=Controls
	zones := []fyne.CanvasObject{
		leftColumn,
		plotPanel,
		rightColumn,
	}
	tabLabels := []string{"Info", "P-E Plot", "Controls"}

	adaptive := sharedwidgets.NewAdaptiveLayout(zones, tabLabels)
	adaptive.SetDesktopLayout(func(zones []fyne.CanvasObject) fyne.CanvasObject {
		// Desktop: Left (Cell+Info) | Plot | Controls
		// At 1024px: left=256px, inner=(768px * 0.65)=499px plot, 269px controls
		innerSplit := container.NewHSplit(zones[1], zones[2])
		innerSplit.SetOffset(0.65) // 65% P-E plot, 35% controls panel

		outerSplit := container.NewHSplit(zones[0], innerSplit)
		outerSplit.SetOffset(0.25) // 25% cell+info column, 75% plot+controls

		return outerSplit
	})

	return container.NewBorder(
		nil,
		statusBar,
		nil, nil,
		adaptive.Content(),
	)
}

// createSectionDivider creates a subtle divider between sections
func (a *App) createSectionDivider() fyne.CanvasObject {
	line := canvas.NewRectangle(color.RGBA{0, 100, 180, 100})
	line.SetMinSize(fyne.NewSize(0, 1))
	return container.NewPadded(line)
}

func (a *App) createHeader() fyne.CanvasObject {
	title := widget.NewLabelWithStyle(
		"FeCIM Ferroelectric Hysteresis Visualization",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	return container.NewVBox(title, widget.NewSeparator())
}
