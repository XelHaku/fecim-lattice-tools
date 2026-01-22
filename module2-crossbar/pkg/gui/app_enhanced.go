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

	// Create color legends
	condLegend := NewColorLegend("0", "29", "Level", 30)
	condLegend.SetColormap("fecim")

	irLegend := NewColorLegend("0%", "100%", "Drop", 0)
	irLegend.SetColormap("viridis")

	sneakLegend := NewColorLegend("Low", "High", "Sneak", 0)
	sneakLegend.SetColormap("plasma")

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

	// Store references for updates
	ca.metricsPanel = metricsPanel
	ca.comparisonBadge = compBadge
	ca.accuracyWaterfall = accWaterfall
	ca.beforeAfterToggle = beforeAfter

	// Conductance tab with legend
	condContent := container.NewBorder(
		nil, nil,
		nil,
		condLegend,
		ca.conductanceHeatmap,
	)

	// IR Drop tab with legend
	irContent := container.NewBorder(
		nil, nil,
		nil,
		irLegend,
		ca.irDropHeatmap,
	)

	// Sneak Path tab with legend
	sneakContent := container.NewBorder(
		nil, nil,
		nil,
		sneakLegend,
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

	// Update educational panel based on selected tab
	ca.tabs.OnSelected = func(tab *container.TabItem) {
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

	// Create simple LEFT panel labels
	ca.eduTitleLabel = widget.NewLabelWithStyle("What You're Seeing", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ca.eduContentLabel = widget.NewLabel("CROSSBAR MVM\n\nClick a button to start\na demonstration.")
	ca.eduContentLabel.Wrapping = fyne.TextWrapWord
	ca.keyStatLabel = widget.NewLabel("N² Operations")
	ca.keyStatLabel.Alignment = fyne.TextAlignCenter
	ca.keyStatValue = widget.NewLabelWithStyle(fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Create simple RIGHT panel widgets
	ca.runMVMButton = widget.NewButton("Run Enhanced MVM", ca.runEnhancedMVM)
	ca.runMVMButton.Importance = widget.HighImportance

	ca.resetButton = widget.NewButton("Reset Array", ca.resetArray)

	exportButton := widget.NewButton("Export Data", ca.exportData)

	ca.arraySizeLabel = widget.NewLabel("Array Size: 64x64")
	ca.arraySizeSlider = widget.NewSlider(8, 128)
	ca.arraySizeSlider.Step = 8
	ca.arraySizeSlider.Value = 64
	ca.arraySizeSlider.OnChanged = func(v float64) {
		size := int(v)
		ca.arraySizeLabel.SetText(fmt.Sprintf("Array Size: %dx%d", size, size))
		ca.recreateArray(size, ca.config.NoiseLevel, ca.config.ADCBits)
	}

	ca.noiseLabel = widget.NewLabel("Noise: 2.0%")
	ca.noiseSlider = widget.NewSlider(0, 20)
	ca.noiseSlider.Step = 0.5
	ca.noiseSlider.Value = 2
	ca.noiseSlider.OnChanged = func(v float64) {
		ca.noiseLabel.SetText(fmt.Sprintf("Noise: %.1f%%", v))
		ca.config.NoiseLevel = v / 100.0
	}

	ca.adcBitsLabel = widget.NewLabel("ADC Bits: 6")
	ca.adcBitsSlider = widget.NewSlider(4, 10)
	ca.adcBitsSlider.Step = 1
	ca.adcBitsSlider.Value = 6
	ca.adcBitsSlider.OnChanged = func(v float64) {
		bits := int(v)
		ca.adcBitsLabel.SetText(fmt.Sprintf("ADC Bits: %d", bits))
		ca.config.ADCBits = bits
	}

	ca.colormapSelect = widget.NewSelect([]string{"fecim", "viridis", "plasma", "coolwarm"}, func(s string) {
		ca.conductanceHeatmap.SetColormap(s)
		condLegend.SetColormap(s)
	})
	ca.colormapSelect.SetSelected("fecim")

	// Create status labels
	ca.statusLabel = widget.NewLabel("● IDLE | Ready for operations")
	ca.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	ca.infoLabel = widget.NewLabel(fmt.Sprintf(
		"Crossbar: %dx%d | Levels: 30 | Noise: %.1f%% | ADC: %d bits",
		ca.config.Rows, ca.config.Cols, ca.config.NoiseLevel*100, ca.config.ADCBits,
	))

	ca.hoverInfoLabel = widget.NewLabel("Hover over cells to see values")
	ca.hoverInfoLabel.TextStyle = fyne.TextStyle{Monospace: true}

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

	// Metrics and comparison section
	metricsSection := container.NewVBox(
		widget.NewSeparator(),
		metricsPanel,
		widget.NewSeparator(),
		compBadge,
	)
	metricsScroll := container.NewVScroll(metricsSection)

	rightPanel := container.NewVSplit(controlsScroll, metricsScroll)
	rightPanel.SetOffset(0.5)

	// Left panel
	leftPanel := container.NewVBox(
		ca.eduTitleLabel,
		widget.NewSeparator(),
		ca.eduContentLabel,
		widget.NewSeparator(),
		ca.keyStatLabel,
		ca.keyStatValue,
	)

	// Status footer
	ca.modeIndicator = NewModeIndicatorBox()
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
			irMap := mvmResult.IRDropAnalysis.GetIRDropMap()
			ca.irDropHeatmap.SetData(irMap)
			ca.irDropHeatmap.SetSelection(
				mvmResult.IRDropAnalysis.WorstCaseCell[0],
				mvmResult.IRDropAnalysis.WorstCaseCell[1],
			)
		}

		// Update sneak path heatmap
		if mvmResult.SneakPathAnalysis != nil {
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
