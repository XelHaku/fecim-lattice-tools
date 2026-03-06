// Package gui - Tab creation and management for enhanced crossbar app
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module2-crossbar/pkg/gui/tabs"
	"fecim-lattice-tools/shared/crossbar"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
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
	// Fixed scale: GetIRDropMap returns values in [0,1] where value = percentage/100
	// e.g., 0.05 = 5% drop, matching the 0-100% legend
	ca.irDropHeatmap.SetFixedScale(0, 1.0)
	ca.irDropHeatmap.OnCellTapped = ca.onIRDropCellTapped
	ca.irDropHeatmap.OnCellHover = ca.onIRDropCellHover

	ca.sneakPathHeatmap = NewCrossbarHeatmap(ca.config.Rows, ca.config.Cols)
	ca.sneakPathHeatmap.SetColormap("plasma")
	// Fixed scale: GetSneakMap returns values normalized to [0,1] where 1.0 = 200% sneak ratio
	// This ensures consistent visual comparison across architectures (1T1R vs 0T1R)
	ca.sneakPathHeatmap.SetFixedScale(0, 1.0)
	ca.sneakPathHeatmap.OnCellTapped = ca.onSneakCellTapped
	ca.sneakPathHeatmap.OnCellHover = ca.onSneakCellHover

	// Create color legends using shared widget and store in app
	ca.condLegend = sharedwidgets.NewColorLegendWithColormap(0, 29, "Level", true, "fecim")
	ca.irLegend = sharedwidgets.NewColorLegendWithColormap(0, 10, "%", true, "viridis")    // Typical IR drop range ~1-10%
	ca.sneakLegend = sharedwidgets.NewColorLegendWithColormap(0, 100, "%", true, "plasma") // Sneak ratio: 0-100% of signal

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
	accWaterfall.SetTarget(96.6) // Literature benchmark: multi-level FeFET crossbar (Nature Comms 2023, DOI: 10.1038/s41467-023-42110-y)

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

	// FeCAP tab (LIT-P2-05): charge-domain MVM visualization with its own internal array.
	// The closures in FeCAPTab button handlers keep the struct alive.
	fecapRows := ca.config.Rows
	if fecapRows > 8 {
		fecapRows = 8 // Cap at 8×8 for readable bar charts
	}
	fecapTab := tabs.NewFeCAPTab(fecapRows)
	fecapContent := fecapTab.Content()

	// L10: 3D multi-layer stack visualization
	stack3dTab := tabs.NewStack3DTab()
	stack3dContent := stack3dTab.Content()

	// Create tabbed view with new tabs
	ca.tabs = container.NewAppTabs(
		container.NewTabItem("Conductance", condContent),
		container.NewTabItem("IR Drop", irContent),
		container.NewTabItem("Sneak Paths", sneakContent),
		container.NewTabItem("Input/Output", container.NewMax(ca.mvmVis)),
		container.NewTabItem("Ideal vs Actual", beforeAfterTab),
		container.NewTabItem("Accuracy Analysis", waterfallTab),
		container.NewTabItem("FeCAP Mode", fecapContent),
		container.NewTabItem("3D Stack", stack3dContent),
	)

	// Update educational panel based on selected tab and preserve selection
	ca.tabs.OnSelected = func(tab *container.TabItem) {
		// Clear badge when tab is viewed (C2 accessibility fix)
		baseName := ca.getBaseTabName(tab.Text)
		ca.clearTabBadge(baseName)

		// Apply persisted selection to the newly selected tab's heatmap
		ca.stateMu.RLock()
		row, col := ca.selectedRow, ca.selectedCol
		ca.stateMu.RUnlock()
		if row >= 0 && col >= 0 && row < ca.config.Rows && col < ca.config.Cols {
			ca.syncSelection(row, col)
			// Update tooltip for the selected cell based on current tab
			ca.updateTooltipForTab(baseName, row, col)
		}

		// Sync colormap dropdown to show current tab's colormap
		switch baseName {
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

		switch baseName {
		case "Conductance":
			ca.setEducationalContent("Conductance Matrix",
				"Each cell = one FeFET\n\n"+
					"Color = conductance level\n"+
					"(0-29 discrete states; demo baseline)\n\n"+
					"This is your weight matrix W\n"+
					"for neural network inference.\n\n"+
					"30 levels (conference claim) = 4.9 bits/cell\n"+
					"vs 1 bit for binary memory\n\n"+
					"💡 Tip: Click ⓘ for physics details")
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
					"Peer-reviewed: 96-98%\n\n"+
					"Shows where accuracy\n"+
					"is lost and why.")
		case "FeCAP Mode":
			ca.setEducationalContent("FeCAP (Charge Domain)",
				"FeCAP = ferroelectric\ncapacitor crossbar.\n\n"+
					"State → capacitance C\n"+
					"(not conductance G)\n\n"+
					"MVM: Q = C × V\n"+
					"(charge-domain)\n\n"+
					"Key advantages:\n"+
					"• No DC path → no sneak\n"+
					"• No IR drop\n"+
					"• E = ½CV² per cell\n"+
					"• 14-57× lower energy\n\n"+
					"Sensing:\n"+
					"Charge amplifier detects\n"+
					"I_disp = ΔQ/Δt during\n"+
					"10 ns write pulse.\n\n"+
					"Reference:\n"+
					"Adv. Intell. Syst. 2022\n"+
					"128×128 demo, 3.8 pJ/MVM")
		case "3D Stack":
			ca.setEducationalContent("3D Layer Stack",
				"Multi-layer FeCIM array\nvisualization (like 3D NAND).\n\n"+
					"Stacked crossbar layers\n"+
					"share vertical interconnects.\n\n"+
					"Controls:\n"+
					"• Drag to rotate camera\n"+
					"• Scroll to zoom\n"+
					"• Click to select layer\n"+
					"• Adjust sliders for\n"+
					"  layer count and grid\n\n"+
					"Performance:\n"+
					"• Up to 512 layers\n"+
					"• Auto-subsampling when\n"+
					"  layers exceed visible cap\n\n"+
					"Pure software renderer\n"+
					"(no GPU required)")
		}
	}

	return ca.createMainLayoutStructure(metricsPanel, compBadge)
}

// createMainLayoutStructure creates the overall layout structure with panels and splits.
func (ca *CrossbarApp) createMainLayoutStructure(metricsPanel *MetricsPanel, compBadge *ComparisonBadge) fyne.CanvasObject {
	// Create simple LEFT panel labels - disable wrapping to prevent MinSize changes
	ca.eduTitleLabel = widget.NewLabelWithStyle("What You're Seeing", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ca.eduTitleLabel.Wrapping = fyne.TextWrapOff
	ca.eduContentLabel = widget.NewLabel("CROSSBAR MVM\n\nClick a button to start\na demonstration.")
	ca.eduContentLabel.Wrapping = fyne.TextWrapWord
	ca.keyStatLabel = widget.NewLabel("N² Operations")
	ca.keyStatLabel.Alignment = fyne.TextAlignCenter
	ca.keyStatLabel.Wrapping = fyne.TextWrapOff
	ca.keyStatValue = widget.NewLabelWithStyle(fmt.Sprintf("%d MACs", ca.config.Rows*ca.config.Cols), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ca.keyStatValue.Wrapping = fyne.TextWrapOff

	// Create simple RIGHT panel widgets - delegated to app_controls.go
	ca.createControlWidgets()

	// Stats label for cell analysis
	ca.statsLabel = widget.NewLabel("Analysis Results\n\nNo data yet.\nClick a cell or Run MVM.")
	ca.statsLabel.Wrapping = fyne.TextWrapOff
	ca.statsLabel.TextStyle = fyne.TextStyle{Monospace: false}

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

	// Right panel - delegated to app_controls.go
	rightPanel := ca.createRightPanel(metricsScroll)

	// Left panel - keep educational content in a bounded viewport to avoid parent layout shifts.
	// Scroll + fixed viewport avoids clipping for longer descriptions on small windows.
	eduContentScroll := container.NewVScroll(ca.eduContentLabel)
	eduContentScroll.SetMinSize(fyne.NewSize(200, 300))
	eduContentWrapper := container.NewGridWrap(fyne.NewSize(200, 300), eduContentScroll)

	leftPanelContent := container.NewVBox(
		ca.eduTitleLabel,
		widget.NewSeparator(),
		eduContentWrapper,
		widget.NewSeparator(),
		ca.keyStatLabel,
		ca.keyStatValue,
	)
	leftPanel := container.NewVScroll(leftPanelContent)

	// Status footer - delegated to app_controls.go
	simpleFooter := ca.createStatusFooter()

	// Header: lightweight access to shared science primer
	aboutScienceBtn := sharedwidgets.CreateAboutScienceButton(ca.window)
	aboutScienceBtn.Importance = widget.LowImportance
	header := container.NewVBox(
		container.NewHBox(layout.NewSpacer(), aboutScienceBtn),
		widget.NewSeparator(),
	)

	// Layout with HSplit
	ca.leftCenterSplit = container.NewHSplit(leftPanel, ca.tabs)
	ca.leftCenterSplit.SetOffset(0.15)

	ca.mainSplit = container.NewHSplit(ca.leftCenterSplit, rightPanel)
	ca.mainSplit.SetOffset(0.75)

	// Create responsive detector for breakpoint-based layout adjustments
	ca.responsiveDetector = sharedwidgets.NewResponsiveDetector(ca.onBreakpointChangeEnhanced)
	ca.currentBreakpoint = sharedwidgets.BreakpointXL // Default to desktop

	mainContent := container.NewBorder(
		header,
		simpleFooter,
		nil,
		nil,
		ca.mainSplit,
	)

	// Stack with responsive detector overlay
	return container.NewStack(mainContent, ca.responsiveDetector)
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
			tooltip := sharedwidgets.ConductanceTooltip(row, col, value, ca.array)
			ca.statsLabel.SetText(tooltip)
			level := crossbar.GetLevel(value)
			ca.updateStatus(fmt.Sprintf("READ | Cell [%d,%d] = Level %d/30 (%.2f µS)",
				row, col, level, value*99+1))
		}

	case "IR Drop":
		tooltip := sharedwidgets.IRDropTooltip(row, col, ca.lastIRDropAnalysis, ca.array)
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
		tooltip := sharedwidgets.SneakPathTooltip(row, col, ca.lastSneakAnalysis, sneakTargetRow, sneakTargetCol, ca.array)
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

	case "FeCAP Mode":
		ca.statsLabel.SetText(
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n" +
				"FeCAP CHARGE-DOMAIN MVM\n" +
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n" +
				"Physics:\n" +
				"  Q[j] = Σᵢ C[i,j] × V_DAC[i]\n" +
				"  (charge domain, not current)\n\n" +
				"Cell capacitance:\n" +
				"  C_phys ∈ [0.5 fF, 2.0 fF]\n" +
				"  ×4 modulation ratio (HZO)\n\n" +
				"Read chain:\n" +
				"  1. WL pulse (10 ns)\n" +
				"  2. Charge accumulates on BL\n" +
				"  3. Charge amp: V = Q/C_fb\n" +
				"  4. ADC quantizes V_out\n\n" +
				"No sneak paths — DC-blocked.\n" +
				"No IR drop — no resistive WL.\n\n" +
				"Use FeCAP controls to program\n" +
				"capacitances and run Charge MVM.",
		)
		ca.updateStatus("FeCAP MODE | Charge-domain MVM (Q = C×V) — no sneak paths or IR drop")

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
					"Peer-reviewed: 96-98%%\n"+
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

// onBreakpointChangeEnhanced handles responsive layout adjustments for enhanced mode.
func (ca *CrossbarApp) onBreakpointChangeEnhanced(bp sharedwidgets.Breakpoint, size fyne.Size) {
	ca.currentBreakpoint = bp

	// Adjust split offsets based on breakpoint
	// Enhanced mode has more content in the right panel, so adjust accordingly
	switch bp {
	case sharedwidgets.BreakpointSM, sharedwidgets.BreakpointMD:
		// Small/Medium: Minimize side panels, maximize heatmap area
		// Left panel: collapse to 5%
		// Right panel: 15% (need a bit more for metrics)
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.05) // 5% left
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.85) // 85% left+center, 15% right
		}

	case sharedwidgets.BreakpointLG:
		// Large: Balanced layout for laptops
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.12) // 12% left
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.8) // 80% left+center, 20% right
		}

	case sharedwidgets.BreakpointXL:
		// Extra Large: Desktop - original comfortable layout
		if ca.leftCenterSplit != nil {
			ca.leftCenterSplit.SetOffset(0.15) // 15% left
		}
		if ca.mainSplit != nil {
			ca.mainSplit.SetOffset(0.75) // 75% left+center, 25% right
		}
	}
}
