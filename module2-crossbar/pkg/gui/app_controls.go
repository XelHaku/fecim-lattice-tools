// Package gui - Control panel and widget creation for crossbar app
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// createControlWidgets creates all control panel widgets (buttons, sliders, dropdowns).
func (ca *CrossbarApp) createControlWidgets() {
	// Reset button
	ca.resetButton = widget.NewButton("Reset", ca.resetArray)
	ca.resetButton.Importance = widget.MediumImportance

	// Array size slider (8 to 128 in steps of 8)
	// Create without callback first, set value, then add callback to avoid
	// triggering recreateArray before UI is initialized
	ca.arraySizeLabel = widget.NewLabel("Array Size: 64×64")
	ca.arraySizeSlider = widget.NewSlider(8, 128)
	ca.arraySizeSlider.Step = 8
	ca.arraySizeSlider.Value = 64
	ca.arraySizeSlider.OnChanged = func(v float64) {
		size := int(v)
		ca.arraySizeLabel.SetText(fmt.Sprintf("Array Size: %d×%d", size, size))
		ca.recreateArray(size, ca.config.NoiseLevel, ca.config.ADCBits)
	}

	// Keep sliders for continuous values but with compact labels
	ca.noiseLabel = widget.NewLabel("2.0%")
	ca.noiseLabel.Wrapping = fyne.TextWrapOff
	ca.noiseSlider = widget.NewSlider(0, 20)
	ca.noiseSlider.Step = 0.5
	ca.noiseSlider.Value = 2
	ca.noiseSlider.OnChanged = func(v float64) {
		ca.noiseLabel.SetText(fmt.Sprintf("%.1f%%", v))
		ca.config.NoiseLevel = v / 100.0
		ca.runEnhancedMVMInstant()
	}

	ca.adcBitsLabel = widget.NewLabel("6")
	ca.adcBitsLabel.Wrapping = fyne.TextWrapOff
	ca.adcBitsSlider = widget.NewSlider(4, 10)
	ca.adcBitsSlider.Step = 1
	ca.adcBitsSlider.Value = 6
	ca.adcBitsSlider.OnChanged = func(v float64) {
		bits := int(v)
		ca.adcBitsLabel.SetText(fmt.Sprintf("%d", bits))
		ca.config.ADCBits = bits
		ca.runEnhancedMVMInstant()
	}

	ca.colormapSelect = widget.NewSelect([]string{"fecim", "viridis", "plasma", "coolwarm"}, func(s string) {
		// Change colormap for the currently active tab and store the selection
		if ca.tabs != nil {
			switch ca.tabs.Selected().Text {
			case "Conductance":
				ca.conductanceHeatmap.SetColormap(s)
				ca.condLegend.SetColormap(s)
				ca.condColormap = s
			case "IR Drop":
				ca.irDropHeatmap.SetColormap(s)
				ca.irLegend.SetColormap(s)
				ca.irColormap = s
			case "Sneak Paths":
				ca.sneakPathHeatmap.SetColormap(s)
				ca.sneakLegend.SetColormap(s)
				ca.sneakColormap = s
			default:
				// For other tabs, default to conductance
				ca.conductanceHeatmap.SetColormap(s)
				ca.condLegend.SetColormap(s)
				ca.condColormap = s
			}
		} else {
			ca.conductanceHeatmap.SetColormap(s)
			ca.condLegend.SetColormap(s)
			ca.condColormap = s
		}
	})
	ca.colormapSelect.SetSelected("fecim")

	// Architecture toggle: 0T1R (passive) vs 1T1R (with access transistor)
	// This affects sneak path and IR drop physics calculations
	ca.architecture = sharedwidgets.Architecture0T1R // Default to passive

	// Create toggle buttons
	ca.archPassiveBtn = widget.NewButton("PASSIVE", nil)
	ca.arch1T1RBtn = widget.NewButton("1T1R GATE", nil)

	// Helper to update button styles based on selection
	updateArchButtons := func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			ca.archPassiveBtn.Importance = widget.HighImportance
			ca.arch1T1RBtn.Importance = widget.LowImportance
		} else {
			ca.archPassiveBtn.Importance = widget.LowImportance
			ca.arch1T1RBtn.Importance = widget.HighImportance
		}
		ca.archPassiveBtn.Refresh()
		ca.arch1T1RBtn.Refresh()
	}

	// Set initial state
	updateArchButtons()

	// Wire up callbacks
	ca.archPassiveBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			return // Already selected
		}
		debug.Printf("[ARCH TOGGLE] Switched to: PASSIVE (0T1R)")

		ca.stateMu.Lock()
		ca.architecture = sharedwidgets.Architecture0T1R
		ca.stateMu.Unlock()

		updateArchButtons()

		// Update educational content
		title, content := sharedwidgets.ArchitectureInfo(sharedwidgets.Architecture0T1R)
		ca.setEducationalContent(title, content)

		// Re-run MVM
		ca.runEnhancedMVMWithCurrentInput()
	}

	ca.arch1T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture1T1R {
			return // Already selected
		}
		debug.Printf("[ARCH TOGGLE] Switched to: 1T1R GATE")

		ca.stateMu.Lock()
		ca.architecture = sharedwidgets.Architecture1T1R
		ca.stateMu.Unlock()

		updateArchButtons()

		// Update educational content
		title, content := sharedwidgets.ArchitectureInfo(sharedwidgets.Architecture1T1R)
		ca.setEducationalContent(title, content)

		// Re-run MVM
		ca.runEnhancedMVMWithCurrentInput()
	}

	// Create horizontal container for toggle
	ca.archToggle = container.NewGridWithColumns(2, ca.archPassiveBtn, ca.arch1T1RBtn)
}

// createRightPanel creates the right panel with controls and metrics.
func (ca *CrossbarApp) createRightPanel(metricsScroll *container.Scroll) *container.Split {
	exportButton := widget.NewButton("Export", ca.exportData)
	exportButton.Importance = widget.MediumImportance

	// === ARRAY CONFIG - Slider with label ===
	arraySizeRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Array:"),
		ca.arraySizeLabel,
		ca.arraySizeSlider,
	)

	// === ARCHITECTURE TOGGLE ===
	archLabel := widget.NewLabelWithStyle("Architecture", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// === SIGNAL QUALITY - Inline labels with sliders ===
	noiseRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Noise:"),
		ca.noiseLabel,
		ca.noiseSlider,
	)
	adcRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("ADC:"),
		ca.adcBitsLabel,
		ca.adcBitsSlider,
	)

	// === DISPLAY & ACTIONS - Combined compact row ===
	colormapRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Color:"),
		nil,
		ca.colormapSelect,
	)
	actionButtons := container.NewGridWithColumns(2, ca.resetButton, exportButton)

	// === ASSEMBLE CONTROLS ===
	controlsBox := container.NewVBox(
		arraySizeRow,
		widget.NewSeparator(),
		archLabel,
		ca.archToggle,
		widget.NewSeparator(),
		noiseRow,
		adcRow,
		widget.NewSeparator(),
		colormapRow,
		layout.NewSpacer(),
		actionButtons,
	)
	controlsScroll := container.NewVScroll(controlsBox)
	controlsScroll.SetMinSize(fyne.NewSize(220, 280))

	rightPanel := container.NewVSplit(controlsScroll, metricsScroll)
	rightPanel.SetOffset(0.45) // Give more space to metrics

	return rightPanel
}

// createStatusFooter creates the status footer with indicators and info labels.
func (ca *CrossbarApp) createStatusFooter() *fyne.Container {
	// Create status labels - disable wrapping to prevent MinSize changes on SetText
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	ca.statusLabel.Wrapping = fyne.TextWrapOff

	// 30 Levels info tooltip
	levelsInfoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("30 Analog Levels",
			"FeCIM uses 30 discrete analog conductance states per cell.\n\n"+
				"30 levels = ~4.9 bits/cell vs 1 bit for binary memory.\n\n"+
				"Each level represents a stable polarization state in the FeFET.", ca.window)
	})
	levelsInfoBtn.Importance = widget.LowImportance

	ca.infoLabel = widget.NewLabel(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))
	ca.infoLabel.Wrapping = fyne.TextWrapOff

	infoRow := container.NewHBox(ca.infoLabel, levelsInfoBtn)

	ca.hoverInfoLabel = widget.NewLabel("Hover over cells to see values")
	ca.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}
	ca.hoverInfoLabel.Wrapping = fyne.TextWrapOff
	ca.hoverInfoLabel.Truncation = fyne.TextTruncateEllipsis

	ca.modeIndicator = NewModeIndicatorBox()
	ca.levelIndicator = NewLevelIndicator()

	// Wrap hoverInfoLabel in fixed-size container to prevent layout recalc on text change
	hoverInfoContainer := container.NewGridWrap(fyne.NewSize(450, 20), ca.hoverInfoLabel)

	return container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		hoverInfoContainer,
		widget.NewSeparator(),
		infoRow,
	)
}
