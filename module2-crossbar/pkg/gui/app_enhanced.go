// Package gui - Enhanced app with all new widgets integrated
package gui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
)

// createEnhancedMainLayout builds the main application layout with all new features.
func (ca *CrossbarApp) createEnhancedMainLayout() fyne.CanvasObject {
	// Create heatmaps
	ca.conductanceHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.conductanceHeatmap.SetColormap("fecim")
	ca.conductanceHeatmap.OnCellTapped = ca.onCellTapped
	ca.conductanceHeatmap.OnCellHover = ca.onCellHover

	ca.irDropHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.irDropHeatmap.SetColormap("viridis")
	ca.irDropHeatmap.OnCellTapped = ca.onIRDropCellTapped
	ca.irDropHeatmap.OnCellHover = ca.onIRDropCellHover

	ca.sneakPathHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.sneakPathHeatmap.SetColormap("plasma")
	ca.sneakPathHeatmap.OnCellTapped = ca.onSneakCellTapped
	ca.sneakPathHeatmap.OnCellHover = ca.onSneakCellHover

	// Create color legends and store in app
	ca.condLegend = NewColorLegend("0", "29", "Level", 30)
	ca.condLegend.SetColormap("fecim")

	ca.irLegend = NewColorLegend("0%", "100%", "Drop", 0)
	ca.irLegend.SetColormap("viridis")

	ca.sneakLegend = NewColorLegend("Low", "High", "Sneak", 0)
	ca.sneakLegend.SetColormap("plasma")

	// Initialize per-tab colormap tracking with defaults
	ca.condColormap = "fecim"
	ca.irColormap = "viridis"
	ca.sneakColormap = "plasma"

	// Create MVM visualization
	ca.mvmVis = NewMVMVisualization()

	// Create metrics panel
	metricsPanel := NewMetricsPanel()

	// Create comparison badge
	compBadge := NewComparisonBadge("Energy per 4096 MACs")

	// Create accuracy waterfall
	accWaterfall := NewAccuracyWaterfall()
	accWaterfall.SetTarget(87.0) // Dr. Tour's target

	// Create before/after toggle
	beforeAfter := NewBeforeAfterToggle(ca.config.Rows, ca.config.Cols)

	// Wire up before/after cell interaction handlers
	beforeAfter.OnCellTapped = ca.onBeforeAfterCellTapped
	beforeAfter.OnCellHover = ca.onBeforeAfterCellHover

	// Store references for updates
	ca.metricsPanel = metricsPanel
	ca.comparisonBadge = compBadge
	ca.accuracyWaterfall = accWaterfall
	ca.beforeAfterToggle = beforeAfter

	// Conductance tab with legend
	condContent := container.NewBorder(
		nil, nil,
		nil,
		ca.condLegend,
		ca.conductanceHeatmap,
	)

	// IR Drop tab with legend
	irContent := container.NewBorder(
		nil, nil,
		nil,
		ca.irLegend,
		ca.irDropHeatmap,
	)

	// Sneak Path tab with legend
	sneakContent := container.NewBorder(
		nil, nil,
		nil,
		ca.sneakLegend,
		ca.sneakPathHeatmap,
	)

	// Before/After comparison tab
	beforeAfterTab := container.NewBorder(
		widget.NewLabelWithStyle("Ideal vs Actual Comparison", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		beforeAfter,
	)

	// Accuracy waterfall tab
	waterfallTab := container.NewBorder(
		widget.NewLabelWithStyle("Accuracy Degradation Analysis", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		accWaterfall,
	)

	// Create tabbed view with new tabs
	ca.tabs = container.NewAppTabs(
		container.NewTabItem("Conductance", condContent),
		container.NewTabItem("IR Drop", irContent),
		container.NewTabItem("Sneak Paths", sneakContent),
		container.NewTabItem("Input/Output", container.NewMax(ca.mvmVis)),
		container.NewTabItem("Ideal vs Actual", beforeAfterTab),
		container.NewTabItem("Accuracy Analysis", waterfallTab),
	)

	// Update educational panel based on selected tab and preserve selection
	ca.tabs.OnSelected = func(tab *container.TabItem) {
		// Apply persisted selection to the newly selected tab's heatmap
		if ca.selectedRow >= 0 && ca.selectedCol >= 0 {
			ca.syncSelection(ca.selectedRow, ca.selectedCol)
			// Update tooltip for the selected cell based on current tab
			ca.updateTooltipForTab(tab.Text, ca.selectedRow, ca.selectedCol)
		}

		// Sync colormap dropdown to show current tab's colormap
		switch tab.Text {
		case "Conductance":
			if ca.colormapSelect != nil && ca.condColormap != "" {
				ca.colormapSelect.SetSelected(ca.condColormap)
			}
		case "IR Drop":
			if ca.colormapSelect != nil && ca.irColormap != "" {
				ca.colormapSelect.SetSelected(ca.irColormap)
			}
		case "Sneak Paths":
			if ca.colormapSelect != nil && ca.sneakColormap != "" {
				ca.colormapSelect.SetSelected(ca.sneakColormap)
			}
		}

		switch tab.Text {
		case "Conductance":
			ca.setEducationalContent("Conductance Matrix",
				"Each cell = one FeFET\n\n"+
					"Color = conductance level\n"+
					"(0-29 discrete states)\n\n"+
					"This is your weight matrix W\n"+
					"for neural network inference.\n\n"+
					"30 levels = 4.9 bits/cell\n"+
					"vs 1 bit for binary memory")
		case "IR Drop":
			ca.setEducationalContent("IR Drop Analysis",
				"Voltage drops along wires.\n\n"+
					"Viridis colormap:\n"+
					"Purple = low drop (good)\n"+
					"Yellow = high drop (bad)\n\n"+
					"Cells far from drivers\n"+
					"see reduced voltage.\n\n"+
					"Mitigation: wider wires,\n"+
					"hierarchical drivers,\n"+
					"tiled architecture")
		case "Sneak Paths":
			ca.setEducationalContent("Sneak Path Analysis",
				"Parasitic currents through\n"+
					"unselected cells.\n\n"+
					"Plasma colormap:\n"+
					"Purple = low sneak (good)\n"+
					"Yellow = high sneak (bad)\n\n"+
					"Bigger arrays = worse.\n\n"+
					"Mitigation:\n"+
					"• Selector devices (1T1R)\n"+
					"• Half-select scheme\n"+
					"• Threshold switching")
		case "Input/Output":
			ca.setEducationalContent("MVM Vectors",
				"Top: Input voltages (V)\n"+
					"Bottom: Output currents (I)\n\n"+
					"I = W × V\n"+
					"(matrix-vector multiply)\n\n"+
					"ALL "+fmt.Sprintf("%d", ca.config.Rows*ca.config.Cols)+" MACs\n"+
					"happen in ONE clock cycle!\n\n"+
					"Physics does the math.\n"+
					"~10 ns latency.\n\n"+
					"GPU needs "+fmt.Sprintf("%d", ca.config.Rows*ca.config.Cols)+" cycles.")
		case "Ideal vs Actual":
			ca.setEducationalContent("Comparison View",
				"Side-by-side view of:\n"+
					"• Ideal (perfect physics)\n"+
					"• Actual (with non-idealities)\n\n"+
					"Toggle modes:\n"+
					"• Split View (default)\n"+
					"• Ideal Only\n"+
					"• Actual Only\n"+
					"• Difference Map\n\n"+
					"Shows impact of IR drop,\n"+
					"sneak paths, and variation.")
		case "Accuracy Analysis":
			ca.setEducationalContent("Accuracy Waterfall",
				"Step-by-step accuracy loss:\n\n"+
					"1. Baseline (ideal)\n"+
					"2. + ADC/DAC quantization\n"+
					"3. + IR drop\n"+
					"4. + Device variation\n"+
					"5. + Sneak paths\n\n"+
					"Target: 87% (Dr. Tour)\n\n"+
					"Shows where accuracy\n"+
					"is lost and why.")
		}
	}

	// Create simple LEFT panel labels - disable wrapping to prevent MinSize changes
	ca.eduTitleLabel = widget.NewLabelWithStyle("What You're Seeing", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ca.eduTitleLabel.Wrapping = fyne.TextWrapOff
	ca.eduContentLabel = widget.NewLabel("CROSSBAR MVM\n\nClick a button to start\na demonstration.")
	ca.eduContentLabel.Wrapping = fyne.TextWrapOff // Was TextWrapWord - causes MinSize changes
	ca.keyStatLabel = widget.NewLabel("N² Operations")
	ca.keyStatLabel.Alignment = fyne.TextAlignCenter
	ca.keyStatLabel.Wrapping = fyne.TextWrapOff
	ca.keyStatValue = widget.NewLabelWithStyle(fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ca.keyStatValue.Wrapping = fyne.TextWrapOff

	// Create simple RIGHT panel widgets
	ca.runMVMButton = widget.NewButton("Run Enhanced MVM", ca.runEnhancedMVM)
	ca.runMVMButton.Importance = widget.HighImportance

	ca.resetButton = widget.NewButton("Reset Array", ca.resetArray)

	exportButton := widget.NewButton("Export Data", ca.exportData)

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

	// Create status labels - disable wrapping to prevent MinSize changes on SetText
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	ca.statusLabel.Wrapping = fyne.TextWrapOff

	ca.infoLabel = widget.NewLabel(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))
	ca.infoLabel.Wrapping = fyne.TextWrapOff

	ca.hoverInfoLabel = widget.NewLabel("Hover over cells to see values")
	ca.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}
	ca.hoverInfoLabel.Wrapping = fyne.TextWrapOff

	// Title and header
	titleLabel := widget.NewLabel("FeCIM Crossbar Array Visualization")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
	)

	// Right panel layout
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
	controlsScroll.SetMinSize(fyne.NewSize(200, 300))

	// Stats label for cell analysis
	ca.statsLabel = widget.NewLabel("Analysis Results\n\nNo data yet.\nClick a cell or Run MVM.")
	ca.statsLabel.Wrapping = fyne.TextWrapOff
	ca.statsLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// Metrics and comparison section
	metricsSection := container.NewVBox(
		widget.NewSeparator(),
		ca.statsLabel,
		widget.NewSeparator(),
		metricsPanel,
		widget.NewSeparator(),
		compBadge,
	)
	metricsScroll := container.NewVScroll(metricsSection)

	rightPanel := container.NewVSplit(controlsScroll, metricsScroll)
	rightPanel.SetOffset(0.5)

	// Left panel - wrap in scroll to prevent layout resize on content change
	leftPanelContent := container.NewVBox(
		ca.eduTitleLabel,
		widget.NewSeparator(),
		ca.eduContentLabel,
		widget.NewSeparator(),
		ca.keyStatLabel,
		ca.keyStatValue,
	)
	leftPanel := container.NewVScroll(leftPanelContent)

	// Status footer
	ca.modeIndicator = NewModeIndicatorBox()
	ca.levelIndicator = NewLevelIndicator()
	simpleFooter := container.NewHBox(
		ca.modeIndicator,
		widget.NewSeparator(),
		ca.statusLabel,
		layout.NewSpacer(),
		ca.hoverInfoLabel,
		widget.NewSeparator(),
		ca.infoLabel,
	)

	// Layout with HSplit
	leftCenterSplit := container.NewHSplit(leftPanel, ca.tabs)
	leftCenterSplit.SetOffset(0.15)

	mainSplit := container.NewHSplit(leftCenterSplit, rightPanel)
	mainSplit.SetOffset(0.75)

	mainContent := container.NewBorder(
		header,
		simpleFooter,
		nil,
		nil,
		mainSplit,
	)

	return mainContent
}

// runEnhancedMVM performs MVM with full non-ideality analysis and updates all widgets.
func (ca *CrossbarApp) runEnhancedMVM() {
	debug.Println("runEnhancedMVM: Starting")

	ca.runMVMButton.Disable()

	// Create random input
	input := make([]float64, ca.config.Cols)
	for i := range input {
		input[i] = rand.Float64()
	}
	ca.lastInput = input
	ca.mvmVis.SetInput(input)

	// Run animated MVM in goroutine
	go ca.runEnhancedMVMAnimated(input)
}

// runEnhancedMVMAnimated performs the enhanced MVM with all analysis.
func (ca *CrossbarApp) runEnhancedMVMAnimated(input []float64) {
	// Phase 1: Input voltages applied (300ms)
	fyne.Do(func() {
		ca.modeIndicator.SetMode(DemoModeCompute)
		ca.updateStatus("COMPUTE | Phase 1: Applying input voltages...")

		cols := make([]int, ca.config.Cols)
		for i := range cols {
			cols[i] = i
		}
		ca.conductanceHeatmap.SetInputHighlight(cols)
		ca.conductanceHeatmap.SetAnimPhase(1, 0)
	})
	time.Sleep(300 * time.Millisecond)

	// Phase 2: Current flowing through cells (500ms)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 2: Current flowing through cells...")
	})

	steps := 10
	for i := 0; i <= steps; i++ {
		progress := float64(i) / float64(steps)
		fyne.Do(func() {
			ca.conductanceHeatmap.SetAnimPhase(2, progress)
		})
		time.Sleep(50 * time.Millisecond)
	}

	// Perform enhanced MVM with all non-idealities
	opts := crossbar.DefaultMVMOptions()
	mvmResult, err := ca.array.MVMWithNonIdealities(input, opts)
	if err != nil {
		fyne.Do(func() {
			ca.updateStatus(fmt.Sprintf("COMPUTE | Error: %v", err))
			ca.modeIndicator.SetMode(DemoModeIdle)
			ca.conductanceHeatmap.ClearAnimation()
			ca.runMVMButton.Enable()
		})
		return
	}

	ca.lastOutput = mvmResult.ActualOutput
	ca.lastMVMResult = mvmResult

	// Phase 3: Output currents collected (300ms)
	fyne.Do(func() {
		ca.updateStatus("COMPUTE | Phase 3: Collecting output currents...")
		ca.mvmVis.SetOutput(mvmResult.ActualOutput)

		rows := make([]int, ca.config.Rows)
		for i := range rows {
			rows[i] = i
		}
		ca.conductanceHeatmap.SetOutputHighlight(rows)
		ca.conductanceHeatmap.SetAnimPhase(3, 1)
	})
	time.Sleep(300 * time.Millisecond)

	// Update all visualizations
	fyne.Do(func() {
		ca.conductanceHeatmap.ClearAnimation()

		// Update metrics panel
		// Estimate baseline accuracy (for demo purposes, use 90%)
		baselineAcc := 90.0
		actualAcc := baselineAcc - mvmResult.AccuracyLoss
		ca.metricsPanel.UpdateMetrics(
			baselineAcc,
			actualAcc,
			mvmResult.TotalEnergy,
			mvmResult.GPUEquivalentEnergy,
			mvmResult.MACOperations,
			mvmResult.Latency,
		)

		// Update comparison badge
		ca.comparisonBadge.UpdateValues(
			fmt.Sprintf("%.2f pJ", mvmResult.TotalEnergy),
			fmt.Sprintf("%.0f pJ", mvmResult.GPUEquivalentEnergy),
			fmt.Sprintf("%.0f× better", mvmResult.EnergyEfficiency),
		)

		// Update accuracy waterfall
		degradation, _ := ca.array.ComputeAccuracyDegradation(input, baselineAcc)
		if degradation != nil {
			steps := make([]WaterfallStep, len(degradation.Degradations))
			colors := []color.RGBA{
				{100, 200, 100, 255}, // Green - baseline
				{150, 200, 150, 255}, // Light green
				{200, 200, 100, 255}, // Yellow
				{255, 180, 100, 255}, // Orange
				{255, 100, 100, 255}, // Red
			}
			for i, deg := range degradation.Degradations {
				steps[i] = WaterfallStep{
					Label:    deg.Source,
					Accuracy: deg.AccuracyNow,
					Loss:     deg.Loss,
					Color:    colors[i%len(colors)],
				}
			}
			ca.accuracyWaterfall.SetSteps(steps)
		}

		// Update before/after toggle
		idealMatrix := ca.array.GetConductanceMatrix()
		// For actual, apply a slight variation (simulated)
		actualMatrix := make([][]float64, len(idealMatrix))
		for i := range idealMatrix {
			actualMatrix[i] = make([]float64, len(idealMatrix[i]))
			for j := range idealMatrix[i] {
				// Apply simulated degradation
				factor := 1.0 - mvmResult.RMSE*rand.Float64()
				actualMatrix[i][j] = idealMatrix[i][j] * factor
			}
		}
		ca.beforeAfterToggle.SetData(idealMatrix, actualMatrix)

		// Update IR drop heatmap
		if mvmResult.IRDropAnalysis != nil {
			ca.lastIRDropAnalysis = mvmResult.IRDropAnalysis // Store for hover/tap info
			irMap := mvmResult.IRDropAnalysis.GetIRDropMap()
			ca.irDropHeatmap.SetData(irMap)
			ca.irDropHeatmap.SetSelection(
				mvmResult.IRDropAnalysis.WorstCaseCell[0],
				mvmResult.IRDropAnalysis.WorstCaseCell[1],
			)
		}

		// Update sneak path heatmap
		if mvmResult.SneakPathAnalysis != nil {
			ca.lastSneakAnalysis = mvmResult.SneakPathAnalysis // Store for hover/tap info
			sneakMap := mvmResult.SneakPathAnalysis.GetSneakMap()
			// Apply sqrt for better visibility
			for i := range sneakMap {
				for j := range sneakMap[i] {
					sneakMap[i][j] = math.Sqrt(sneakMap[i][j])
				}
			}
			ca.sneakPathHeatmap.SetData(sneakMap)
		}

		ca.updateStatus(fmt.Sprintf("COMPUTE | Complete: %d MACs, %.2f pJ, %.0f× better than GPU",
			mvmResult.MACOperations, mvmResult.TotalEnergy, mvmResult.EnergyEfficiency))
		ca.modeIndicator.SetMode(DemoModeIdle)
	})

	// Cycle through key tabs
	tabIndices := []int{0, 1, 2, 4, 5} // Conductance, IR Drop, Sneak, Comparison, Waterfall
	for _, idx := range tabIndices {
		fyne.Do(func() {
			ca.tabs.SelectIndex(idx)
		})
		time.Sleep(3 * time.Second)
	}

	// Return to Conductance tab
	fyne.Do(func() {
		ca.tabs.SelectIndex(0)
		ca.runMVMButton.Enable()
	})

	debug.Println("runEnhancedMVM: Complete")
}

// exportData exports array weights and analysis to files.
func (ca *CrossbarApp) exportData() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Export weights CSV
	weightsPath := fmt.Sprintf("crossbar_weights_%s.csv", timestamp)
	if err := ca.array.ExportWeightsCSV(weightsPath); err != nil {
		ca.updateStatus(fmt.Sprintf("Export failed: %v", err))
		return
	}

	// Export analysis JSON
	if ca.lastMVMResult != nil {
		analysisPath := fmt.Sprintf("crossbar_analysis_%s.json", timestamp)
		if err := ca.array.ExportAnalysisJSON(analysisPath, ca.lastMVMResult); err != nil {
			ca.updateStatus(fmt.Sprintf("Export failed: %v", err))
			return
		}
		ca.updateStatus(fmt.Sprintf("Exported: %s, %s", weightsPath, analysisPath))
	} else {
		ca.updateStatus(fmt.Sprintf("Exported: %s (run MVM for analysis)", weightsPath))
	}
}

// onBeforeAfterCellTapped handles clicks on the Ideal vs Actual comparison heatmaps.
func (ca *CrossbarApp) onBeforeAfterCellTapped(row, col int, isIdeal bool) {
	if ca.beforeAfterToggle == nil {
		return
	}

	// Sync selection across all heatmaps
	ca.syncSelection(row, col)

	var idealVal, actualVal float64
	if ca.beforeAfterToggle.idealData != nil && row < len(ca.beforeAfterToggle.idealData) &&
		col < len(ca.beforeAfterToggle.idealData[0]) {
		idealVal = ca.beforeAfterToggle.idealData[row][col]
	}
	if ca.beforeAfterToggle.actualData != nil && row < len(ca.beforeAfterToggle.actualData) &&
		col < len(ca.beforeAfterToggle.actualData[0]) {
		actualVal = ca.beforeAfterToggle.actualData[row][col]
	}

	idealLevel := crossbar.GetLevel(idealVal)
	actualLevel := crossbar.GetLevel(actualVal)
	diff := idealVal - actualVal
	diffPercent := 0.0
	if idealVal > 0 {
		diffPercent = (diff / idealVal) * 100
	}

	source := "Actual"
	if isIdeal {
		source = "Ideal"
	}

	tooltip := fmt.Sprintf(
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
			"CELL [%d, %d] - IDEAL vs ACTUAL\n"+
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
			"Clicked: %s heatmap\n\n"+
			"Ideal Value:\n"+
			"  Conductance:  %.4f (L%d/29)\n"+
			"  Current:      %.2f µA\n\n"+
			"Actual Value:\n"+
			"  Conductance:  %.4f (L%d/29)\n"+
			"  Current:      %.2f µA\n\n"+
			"Degradation:\n"+
			"  Difference:   %.4f (%.1f%%)\n"+
			"  Level shift:  %d levels\n\n"+
			"Impact:\n"+
			"  %s\n",
		row, col,
		source,
		idealVal, idealLevel, idealVal*99+1,
		actualVal, actualLevel, actualVal*99+1,
		math.Abs(diff), math.Abs(diffPercent),
		int(math.Abs(float64(idealLevel-actualLevel))),
		ca.assessDegradationImpact(diffPercent),
	)

	ca.statsLabel.SetText(tooltip)
	ca.updateStatus(fmt.Sprintf("COMPARISON | Cell [%d,%d]: Ideal L%d → Actual L%d (%.1f%% change)",
		row, col, idealLevel, actualLevel, diffPercent))
}

// onBeforeAfterCellHover handles hover on the Ideal vs Actual comparison heatmaps.
func (ca *CrossbarApp) onBeforeAfterCellHover(row, col int, value float64, isIdeal bool) {
	if row < 0 || col < 0 {
		ca.hoverInfoLabel.SetText("Hover over cells to compare ideal vs actual values")
		return
	}

	if ca.beforeAfterToggle == nil {
		return
	}

	var idealVal, actualVal float64
	if ca.beforeAfterToggle.idealData != nil && row < len(ca.beforeAfterToggle.idealData) &&
		col < len(ca.beforeAfterToggle.idealData[0]) {
		idealVal = ca.beforeAfterToggle.idealData[row][col]
	}
	if ca.beforeAfterToggle.actualData != nil && row < len(ca.beforeAfterToggle.actualData) &&
		col < len(ca.beforeAfterToggle.actualData[0]) {
		actualVal = ca.beforeAfterToggle.actualData[row][col]
	}

	idealLevel := crossbar.GetLevel(idealVal)
	actualLevel := crossbar.GetLevel(actualVal)
	diff := math.Abs(idealVal - actualVal)

	source := "Actual"
	if isIdeal {
		source = "Ideal"
	}

	ca.hoverInfoLabel.SetText(fmt.Sprintf(
		"[%d,%d] %s │ Ideal: L%d (%.3f) │ Actual: L%d (%.3f) │ Δ=%.4f (%d levels)",
		row, col, source, idealLevel, idealVal, actualLevel, actualVal, diff, int(math.Abs(float64(idealLevel-actualLevel)))))
}

// assessDegradationImpact returns a qualitative assessment of degradation.
func (ca *CrossbarApp) assessDegradationImpact(diffPercent float64) string {
	absDiff := math.Abs(diffPercent)
	if absDiff < 1 {
		return "Negligible - within noise margin"
	} else if absDiff < 5 {
		return "Minor - acceptable for most applications"
	} else if absDiff < 10 {
		return "Moderate - may affect precision tasks"
	} else if absDiff < 20 {
		return "Significant - requires compensation"
	}
	return "Critical - exceeds tolerance limits"
}

// updateTooltipForTab updates the stats panel tooltip based on which tab is selected.
func (ca *CrossbarApp) updateTooltipForTab(tabName string, row, col int) {
	if row < 0 || col < 0 {
		return
	}

	switch tabName {
	case "Conductance":
		matrix := ca.array.GetConductanceMatrix()
		if row < len(matrix) && col < len(matrix[0]) {
			value := matrix[row][col]
			tooltip := ConductanceTooltip(row, col, value, ca.array)
			ca.statsLabel.SetText(tooltip)
			level := crossbar.GetLevel(value)
			ca.updateStatus(fmt.Sprintf("READ | Cell [%d,%d] = Level %d/30 (%.2f µS)",
				row, col, level, value*99+1))
		}

	case "IR Drop":
		tooltip := IRDropTooltip(row, col, ca.lastIRDropAnalysis, ca.array)
		ca.statsLabel.SetText(tooltip)
		if ca.lastIRDropAnalysis != nil && row < len(ca.lastIRDropAnalysis.EffectiveVoltage) &&
			col < len(ca.lastIRDropAnalysis.EffectiveVoltage[0]) {
			effectiveV := ca.lastIRDropAnalysis.EffectiveVoltage[row][col]
			dropPercent := (1.0 - effectiveV) * 100
			ca.updateStatus(fmt.Sprintf("IR DROP | Cell [%d,%d]: %.3f V (%.1f%% drop)",
				row, col, effectiveV, dropPercent))
		}

	case "Sneak Paths":
		sneakTargetRow := ca.config.Rows / 2
		sneakTargetCol := ca.config.Cols / 2
		tooltip := SneakPathTooltip(row, col, ca.lastSneakAnalysis, sneakTargetRow, sneakTargetCol, ca.array)
		ca.statsLabel.SetText(tooltip)
		if ca.lastSneakAnalysis != nil && row < len(ca.lastSneakAnalysis.SneakCurrents) &&
			col < len(ca.lastSneakAnalysis.SneakCurrents[0]) {
			sneakCurrent := ca.lastSneakAnalysis.SneakCurrents[row][col]
			sneakRatio := 0.0
			if ca.lastSneakAnalysis.TotalSignal > 0 {
				sneakRatio = sneakCurrent / ca.lastSneakAnalysis.TotalSignal * 100
			}
			ca.updateStatus(fmt.Sprintf("SNEAK | Cell [%d,%d]: %.6f (%.2f%% of signal)",
				row, col, sneakCurrent, sneakRatio))
		}

	case "Ideal vs Actual":
		ca.onBeforeAfterCellTapped(row, col, true) // Reuse existing handler

	case "Accuracy Analysis":
		// Show summary of accuracy degradation, not cell-specific data
		if ca.lastMVMResult != nil {
			tooltip := fmt.Sprintf(
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"+
					"ACCURACY DEGRADATION ANALYSIS\n"+
					"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"+
					"Baseline (Ideal):     90.0%%\n"+
					"After Non-Idealities: %.1f%%\n"+
					"Total Loss:           %.2f%%\n\n"+
					"Breakdown:\n"+
					"  ADC/DAC Quantization: ~%.1f%%\n"+
					"  IR Drop:              ~%.1f%%\n"+
					"  Device Variation:     ~%.1f%%\n"+
					"  Sneak Paths:          ~%.1f%%\n\n"+
					"Target: 87%% (Dr. Tour)\n"+
					"Status: %s\n",
				90.0-ca.lastMVMResult.AccuracyLoss,
				ca.lastMVMResult.AccuracyLoss,
				ca.lastMVMResult.AccuracyLoss*0.2, // Estimated breakdown
				ca.lastMVMResult.AccuracyLoss*0.3,
				ca.lastMVMResult.AccuracyLoss*0.3,
				ca.lastMVMResult.AccuracyLoss*0.2,
				ca.getAccuracyStatus(90.0-ca.lastMVMResult.AccuracyLoss),
			)
			ca.statsLabel.SetText(tooltip)
			ca.updateStatus(fmt.Sprintf("ACCURACY | Final: %.1f%% (%.2f%% loss from ideal)",
				90.0-ca.lastMVMResult.AccuracyLoss, ca.lastMVMResult.AccuracyLoss))
		} else {
			ca.statsLabel.SetText("Run MVM to see accuracy degradation analysis")
			ca.updateStatus("ACCURACY | Run Enhanced MVM to analyze degradation")
		}
	}
}

// getAccuracyStatus returns a status message based on accuracy.
func (ca *CrossbarApp) getAccuracyStatus(accuracy float64) string {
	if accuracy >= 87.0 {
		return "✓ Meets Dr. Tour's 87% target"
	} else if accuracy >= 85.0 {
		return "⚠ Close to target (within 2%)"
	} else if accuracy >= 80.0 {
		return "⚠ Below target - optimization needed"
	}
	return "✗ Significant optimization required"
}
