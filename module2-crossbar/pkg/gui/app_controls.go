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
)

// createControlWidgets creates all control panel widgets (buttons, sliders, dropdowns).
func (ca *CrossbarApp) createControlWidgets() {
	ca.runMVMButton = widget.NewButton("Run Enhanced MVM", ca.runEnhancedMVM)
	ca.runMVMButton.Importance = widget.HighImportance

	ca.resetButton = widget.NewButton("Reset Array", ca.resetArray)

	ca.arraySizeLabel = widget.NewLabel("Array Size: 64x64")
	ca.arraySizeLabel.Wrapping = fyne.TextWrapOff
	ca.arraySizeSlider = widget.NewSlider(8, 128)
	ca.arraySizeSlider.Step = 8
	ca.arraySizeSlider.Value = 64
	ca.arraySizeSlider.OnChanged = func(v float64) {
		size := int(v)
		ca.arraySizeLabel.SetText(fmt.Sprintf("Array Size: %dx%d", size, size))
		ca.recreateArray(size, ca.config.NoiseLevel, ca.config.ADCBits)
	}

	ca.noiseLabel = widget.NewLabel("Noise: 2.0%")
	ca.noiseLabel.Wrapping = fyne.TextWrapOff
	ca.noiseSlider = widget.NewSlider(0, 20)
	ca.noiseSlider.Step = 0.5
	ca.noiseSlider.Value = 2
	ca.noiseSlider.OnChanged = func(v float64) {
		ca.noiseLabel.SetText(fmt.Sprintf("Noise: %.1f%%", v))
		ca.config.NoiseLevel = v / 100.0
	}

	ca.adcBitsLabel = widget.NewLabel("ADC Bits: 6")
	ca.adcBitsLabel.Wrapping = fyne.TextWrapOff
	ca.adcBitsSlider = widget.NewSlider(4, 10)
	ca.adcBitsSlider.Step = 1
	ca.adcBitsSlider.Value = 6
	ca.adcBitsSlider.OnChanged = func(v float64) {
		bits := int(v)
		ca.adcBitsLabel.SetText(fmt.Sprintf("ADC Bits: %d", bits))
		ca.config.ADCBits = bits
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
}

// createRightPanel creates the right panel with controls and metrics.
func (ca *CrossbarApp) createRightPanel(metricsScroll *container.Scroll) *container.Split {
	exportButton := widget.NewButton("Export Data", ca.exportData)

	actionLabel := widget.NewLabelWithStyle("Actions", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	actionsGroup := container.NewVBox(
		actionLabel,
		ca.runMVMButton,
		ca.resetButton,
		exportButton,
	)

	settingsLabel := widget.NewLabelWithStyle("Array Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	settingsGroup := container.NewVBox(
		widget.NewSeparator(),
		settingsLabel,
		ca.arraySizeLabel,
		ca.arraySizeSlider,
	)

	signalLabel := widget.NewLabelWithStyle("Signal Quality", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	signalGroup := container.NewVBox(
		widget.NewSeparator(),
		signalLabel,
		ca.noiseLabel,
		ca.noiseSlider,
		ca.adcBitsLabel,
		ca.adcBitsSlider,
	)

	displayLabel := widget.NewLabelWithStyle("Display", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	displayGroup := container.NewVBox(
		widget.NewSeparator(),
		displayLabel,
		ca.colormapSelect,
	)

	controlsBox := container.NewVBox(
		actionsGroup,
		settingsGroup,
		signalGroup,
		displayGroup,
	)
	controlsScroll := container.NewVScroll(controlsBox)
	controlsScroll.SetMinSize(fyne.NewSize(240, 300))

	rightPanel := container.NewVSplit(controlsScroll, metricsScroll)
	rightPanel.SetOffset(0.5)

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
