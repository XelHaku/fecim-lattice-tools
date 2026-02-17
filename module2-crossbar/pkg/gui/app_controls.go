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

// setControlsEnabled enables or disables interactive controls during MVM animation (M5 UX fix).
func (ca *CrossbarApp) setControlsEnabled(enabled bool) {
	// Disable/enable sliders during animation
	if ca.arraySizeSlider != nil {
		if enabled {
			ca.arraySizeSlider.Enable()
		} else {
			ca.arraySizeSlider.Disable()
		}
	}
	if ca.noiseSlider != nil {
		if enabled {
			ca.noiseSlider.Enable()
		} else {
			ca.noiseSlider.Disable()
		}
	}
	if ca.adcBitsSlider != nil {
		if enabled {
			ca.adcBitsSlider.Enable()
		} else {
			ca.adcBitsSlider.Disable()
		}
	}
	if ca.temperatureSlider != nil {
		if enabled {
			ca.temperatureSlider.Enable()
		} else {
			ca.temperatureSlider.Disable()
		}
	}
	// Disable/enable architecture toggle buttons
	if ca.archPassiveBtn != nil {
		if enabled {
			ca.archPassiveBtn.Enable()
		} else {
			ca.archPassiveBtn.Disable()
		}
	}
	if ca.arch1T1RBtn != nil {
		if enabled {
			ca.arch1T1RBtn.Enable()
		} else {
			ca.arch1T1RBtn.Disable()
		}
	}
	if ca.arch2T1RBtn != nil {
		if enabled {
			ca.arch2T1RBtn.Enable()
		} else {
			ca.arch2T1RBtn.Disable()
		}
	}
	// Disable/enable action buttons
	if ca.resetButton != nil {
		if enabled {
			ca.resetButton.Enable()
		} else {
			ca.resetButton.Disable()
		}
	}
}

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

	// m1 UX fix: Changed noise step from 0.5% to 1.0% for simpler interaction
	ca.noiseLabel = widget.NewLabel("2.0%")
	ca.noiseLabel.Wrapping = fyne.TextWrapOff
	ca.noiseSlider = widget.NewSlider(0, 50)
	ca.noiseSlider.Step = 0.5
	ca.noiseSlider.Value = 2
	ca.noiseSlider.OnChanged = func(v float64) {
		ca.noiseLabel.SetText(fmt.Sprintf("%.1f%%", v))
		ca.config.NoiseLevel = v / 100.0
		ca.runEnhancedMVMInstant()
	}

	ca.adcBitsLabel = widget.NewLabel("6 bits")
	ca.adcBitsLabel.Wrapping = fyne.TextWrapOff
	ca.adcBitsSlider = widget.NewSlider(4, 10)
	ca.adcBitsSlider.Step = 1
	ca.adcBitsSlider.Value = 6
	ca.adcBitsSlider.OnChanged = func(v float64) {
		bits := int(v)
		ca.adcBitsLabel.SetText(fmt.Sprintf("%d bits", bits))
		ca.config.ADCBits = bits
		ca.runEnhancedMVMInstant()
	}

	// Temperature slider (Kelvin)
	ca.temperatureLabel = widget.NewLabel(ca.formatTemperatureLabel(ca.currentTemperatureK()))
	ca.temperatureLabel.Wrapping = fyne.TextWrapOff
	ca.temperatureSlider = widget.NewSlider(77, 450)
	ca.temperatureSlider.Step = 5
	ca.temperatureSlider.Value = ca.currentTemperatureK()
	ca.temperatureSlider.OnChanged = func(v float64) {
		ca.temperatureLabel.SetText(ca.formatTemperatureLabel(v))
		ca.setTemperatureK(v)
		ca.stateMu.Lock()
		ca.baselineMaxIRDrop = 0
		ca.stateMu.Unlock()
		ca.updateInfoLabel()
		ca.runEnhancedMVMInstant()
	}

	ca.colormapSelect = widget.NewSelect([]string{"fecim", "viridis", "plasma", "coolwarm"}, func(s string) {
		// Change colormap for the currently active tab and store the selection
		if ca.tabs != nil {
			tabName := ca.getBaseTabName(ca.tabs.Selected().Text)
			switch tabName {
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
			case "Ideal vs Actual":
				// No SetColormap method on BeforeAfterToggle - just return
				return
			case "Accuracy Analysis":
				// No heatmap on this tab - ignore colormap changes
				return
			case "Input/Output":
				// No heatmap on this tab - ignore colormap changes
				return
			default:
				// Unknown tab - ignore
				return
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

	toggle := sharedwidgets.NewArchitectureToggle(sharedwidgets.ArchitectureToggleOptions{
		Initial:      ca.architecture,
		Style:        sharedwidgets.ArchitectureToggleStyleBullet,
		LabelPassive: "PASSIVE",
		Label1T1R:    "1T1R GATE",
		Label2T1R:    "2T1R",
		OnChanged: func(arch string) {
			switch arch {
			case sharedwidgets.Architecture1T1R:
				getDebug().Printf("[ARCH TOGGLE] Switched to: 1T1R GATE")
			case sharedwidgets.Architecture2T1R:
				getDebug().Printf("[ARCH TOGGLE] Switched to: 2T1R")
			default:
				getDebug().Printf("[ARCH TOGGLE] Switched to: PASSIVE (0T1R)")
			}

			ca.stateMu.Lock()
			ca.architecture = arch
			ca.stateMu.Unlock()

			title, content := sharedwidgets.ArchitectureInfo(arch)
			ca.setEducationalContent(title, content)
			ca.runEnhancedMVMWithCurrentInput()
		},
	})
	ca.archPassiveBtn = toggle.PassiveButton
	ca.arch1T1RBtn = toggle.OneT1RButton
	ca.arch2T1RBtn = toggle.TwoT1RButton
	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)
}

// createRightPanel creates the right panel with controls and metrics.
// M4 UX fix: Returns a Border container instead of VSplit to avoid nested scroll issues.
func (ca *CrossbarApp) createRightPanel(metricsScroll *container.Scroll) fyne.CanvasObject {
	exportButton := widget.NewButton("Export", ca.exportData)
	exportButton.Importance = widget.MediumImportance

	// === ARRAY CONFIG - Slider with label ===
	// Add min/max labels for array size slider
	minLabel := widget.NewLabel("8")
	minLabel.TextStyle = fyne.TextStyle{Monospace: true}
	maxLabel := widget.NewLabel("128")
	maxLabel.TextStyle = fyne.TextStyle{Monospace: true}

	sliderWithLabels := container.NewBorder(
		nil, nil,
		minLabel,
		maxLabel,
		ca.arraySizeSlider,
	)

	arraySizeRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Array:"),
		ca.arraySizeLabel,
		sliderWithLabels,
	)

	// === ARCHITECTURE TOGGLE ===
	archLabel := widget.NewLabelWithStyle("Architecture", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// === SIGNAL QUALITY - Inline labels with sliders ===
	noiseTooltip := sharedwidgets.TooltipContent{
		Title:       "Read/Write Noise",
		Description: "Simulates random variations in conductance programming and read operations.",
		Physics:     "Real analog hardware suffers from thermal noise and device variability. This parameter adds Gaussian noise (σ) to write and read operations to test network robustness.",
		Range:       "0% - 50% (Typical: 1-5%)",
	}
	noiseRow := sharedwidgets.SliderWithTooltip("Noise:", ca.noiseSlider, ca.noiseLabel, noiseTooltip, ca.window)
	adcRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("ADC:"),
		ca.adcBitsLabel,
		ca.adcBitsSlider,
	)
	tempRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Temp:"),
		ca.temperatureLabel,
		ca.temperatureSlider,
	)

	// === DISPLAY & ACTIONS - Combined compact row ===
	colormapRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Color:"),
		nil,
		ca.colormapSelect,
	)
	actionButtons := container.NewGridWithColumns(2, ca.resetButton, exportButton)

	// === EXTERNAL TOOLS VALIDATION ===
	toolWidgets := sharedwidgets.NewToolValidationWidgets(sharedwidgets.ToolValidationOptions{
		Window:              ca.window,
		ButtonLabel:         "Validate Tools",
		DialogTitle:         "External Crossbar Tools",
		StatusLabelMode:     sharedwidgets.ToolStatusSymbolOnly,
		MessageStyle:        sharedwidgets.ToolMessageUnicode,
		IncludeInstall:      true,
		IncludeInstallNotes: true,
	})
	toolWidgets.Button.Importance = widget.LowImportance

	toolsRow := container.NewHBox(
		widget.NewLabel("CrossSim:"), toolWidgets.CrossSimStatus,
		widget.NewLabel("BadCrossbar:"), toolWidgets.BadCrossbarStatus,
		layout.NewSpacer(),
		toolWidgets.Button,
	)

	// === ASSEMBLE CONTROLS ===
	// M4 UX fix: Remove scroll from controls section to avoid nested scroll issues
	// Controls are fixed-height, only metrics need scrolling
	controlsBox := container.NewVBox(
		arraySizeRow,
		widget.NewSeparator(),
		archLabel,
		ca.archToggle,
		widget.NewSeparator(),
		noiseRow,
		adcRow,
		tempRow,
		widget.NewSeparator(),
		colormapRow,
		widget.NewSeparator(),
		actionButtons,
		widget.NewSeparator(),
		toolsRow,
	)

	// Use Border layout: controls at top (fixed), metrics fills rest (scrollable)
	// This eliminates nested scroll containers that compete for scroll events
	rightPanel := container.NewBorder(
		controlsBox,   // top - fixed controls
		nil,           // bottom
		nil,           // left
		nil,           // right
		metricsScroll, // center - scrollable metrics
	)

	return rightPanel
}

// createStatusFooter creates the status footer with indicators and info labels.
func (ca *CrossbarApp) createStatusFooter() *fyne.Container {
	// Create status labels - disable wrapping to prevent MinSize changes on SetText
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	ca.statusLabel.Wrapping = fyne.TextWrapOff
	ca.statusBar = sharedwidgets.NewStatusBarWithLabel(ca.statusLabel, "Status: ")

	// 30 Levels info tooltip
	levelsInfoBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("30 Analog Levels (Conference Claim)",
			"This demo assumes 30 discrete conductance states per cell (conference claim; pending peer review).\n"+
				"Peer-reviewed devices report 32–140 states in related materials.\n\n"+
				"30-level baseline (claim) ≈ 4.9 bits/cell vs 1 bit for binary memory.\n\n"+
				"Each level represents a stable polarization state in the FeFET.", ca.window)
	})
	levelsInfoBtn.Importance = widget.LowImportance

	ca.infoLabel = widget.NewLabel("")
	ca.infoLabel.Wrapping = fyne.TextWrapOff
	ca.updateInfoLabel()

	infoRow := container.NewHBox(ca.infoLabel, levelsInfoBtn)

	ca.hoverInfoLabel = widget.NewLabel("Hover over cells to see values")
	ca.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}
	ca.hoverInfoLabel.Wrapping = fyne.TextWrapOff
	ca.hoverInfoLabel.Truncation = fyne.TextTruncateEllipsis

	ca.modeIndicator = newModeIndicator()
	ca.levelIndicator = NewLevelIndicator()

	// Use direct label for flexibility at narrow widths
	return container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		ca.hoverInfoLabel,
		widget.NewSeparator(),
		infoRow,
	)
}
