---
Module: module2-crossbar
Name: "Crossbar Array MVM Visualization"
Entry: cmd/crossbar-gui/main.go
Package: module2-crossbar/pkg/gui
Description: "Interactive visualization of FeCIM crossbar array matrix-vector multiplication with non-ideality analysis"
---

# Bugs Summary
- [ ] BUG-1: Potential race condition on lastInput/lastOutput without mutex protection in runMVMAnimated
- [ ] BUG-2: Missing fyne.Do wrapper in app.go:204 updateConductanceDisplay call
- [ ] BUG-3: Heatmap refresh during startup can cause layout oscillation (partially mitigated)
- [ ] BUG-4: Educational content wrapping disabled but can still trigger MinSize changes
- [ ] BUG-5: Auto-demo context cancellation may leak if Stop() not called

# Screens

## Main Screen (Standard Mode)
Layout:
  - Type: container.Border
  - File: app.go:535
  - Top: Header (title + separator)
  - Bottom: Status footer (mode indicator + status + hover info)
  - Center: HSplit (leftCenterSplit → mainSplit)

  Components:
    - Header:
        Type: container.VBox
        File: app.go:419-423
        Purpose: "Title bar with FeCIM branding"
        Children:
          - titleLabel:
              Type: widget.Label
              File: app.go:416
              Purpose: "App title: FeCIM Crossbar Array Visualization"
              State: text="FeCIM Crossbar Array Visualization", bold=true, centered
          - Separator:
              Type: widget.Separator
              Purpose: "Visual boundary"

    - LeftPanel:
        Type: container.VBox (wrapped in VScroll)
        File: app.go:500-507, 393-403
        Purpose: "Educational content and statistics"
        State: width=15% (desktop), 5-15% (responsive)
        Components:
          - eduTitleLabel:
              Type: widget.Label
              File: app_enhanced.go:225
              Purpose: "Educational section title"
              State: text="What You're Seeing", bold=true, centered, wrapping=off
              Bindings: Updated by setEducationalContent(), tab selection
              Bug: None

          - eduContentLabel:
              Type: widget.Label
              File: app_enhanced.go:227
              Purpose: "Context-sensitive educational text"
              State: wrapping=off (prevents MinSize changes)
              Bindings: Updated by tabs.OnSelected, setEducationalContent()
              Bug: BUG-4 (wrapping disabled but content changes can still trigger layout)

          - keyStatLabel:
              Type: widget.Label
              File: app_enhanced.go:229
              Purpose: "Key statistic label (e.g., 'N² Operations')"
              State: centered, wrapping=off

          - keyStatValue:
              Type: widget.Label
              File: app_enhanced.go:232
              Purpose: "Key statistic value (e.g., '4096 MACs')"
              State: bold=true, centered, wrapping=off
              Bindings: Updated by setKeyStatValue()

    - CenterPanel:
        Type: container.AppTabs
        File: app.go:354-360
        Purpose: "Multi-view heatmap and visualization"
        State: 4 tabs (Conductance, IR Drop, Sneak Paths, Input/Output)
        Bindings: tabs.OnSelected updates educational panel
        Components:
          - ConductanceTab:
              Type: container.TabItem
              File: app.go:356
              Content: container.Max(conductanceHeatmap)
              Components:
                - conductanceHeatmap:
                    Type: CrossbarHeatmap
                    File: heatmap.go:64-84, app.go:216
                    Purpose: "Visualize conductance matrix (30 discrete levels)"
                    State: rows=64, cols=64, colormap="fecim", minVal=0, maxVal=1
                    Bindings:
                      - OnCellTapped → onCellTapped (callbacks.go:34)
                      - OnCellHover → onCellHover (callbacks.go:58)
                      - Data updated by updateConductanceDisplay() (app.go:618)
                    Bug: BUG-2 (updateConductanceDisplay at line 204 missing fyne.Do)

          - IRDropTab:
              Type: container.TabItem
              File: app.go:357
              Content: container.Max(irDropHeatmap)
              Components:
                - irDropHeatmap:
                    Type: CrossbarHeatmap
                    File: heatmap.go:64-84, app.go:221
                    Purpose: "Visualize voltage drops due to wire resistance"
                    State: rows=64, cols=64, colormap="viridis"
                    Bindings:
                      - OnCellTapped → onIRDropCellTapped (callbacks.go:74)
                      - OnCellHover → onIRDropCellHover (callbacks.go:98)
                      - Data updated by runIRDropAnalysis() (analysis.go:14)
                    Bug: None

          - SneakPathTab:
              Type: container.TabItem
              File: app.go:358
              Content: container.Max(sneakPathHeatmap)
              Components:
                - sneakPathHeatmap:
                    Type: CrossbarHeatmap
                    File: heatmap.go:64-84, app.go:226
                    Purpose: "Visualize parasitic sneak currents"
                    State: rows=64, cols=64, colormap="plasma"
                    Bindings:
                      - OnCellTapped → onSneakCellTapped (callbacks.go:136)
                      - OnCellHover → onSneakCellHover (callbacks.go:167)
                      - Data updated by runSneakPathAnalysis() (analysis.go:40)
                    Bug: None

          - InputOutputTab:
              Type: container.TabItem
              File: app.go:359
              Content: container.Max(mvmVis)
              Components:
                - mvmVis:
                    Type: MVMVisualization
                    File: vectors.go:388-417
                    Purpose: "Visualize input/output vectors for MVM"
                    State: inputChart, outputChart, miniMatrix
                    Bindings:
                      - SetInput() called from runMVM (animation.go:34)
                      - SetOutput() called from runMVMAnimated (animation.go:93)
                    Bug: None

    - RightPanel:
        Type: container.VSplit
        File: app.go:492-497
        Purpose: "Controls and statistics"
        State: offset=0.6 (60% controls, 40% stats)
        Components:
          - ControlsSection:
              Type: container.VScroll
              File: app.go:472-480
              Purpose: "User controls for array configuration"
              Components:
                - ActionsGroup:
                    Type: container.VBox
                    File: app.go:428-433
                    Components:
                      - runMVMButton:
                          Type: widget.Button
                          File: app.go:255
                          Purpose: "Execute matrix-vector multiplication"
                          State: importance=HighImportance, text="Run MVM"
                          Bindings: OnTapped → runMVM() (animation.go:16)
                          Bug: None

                      - resetButton:
                          Type: widget.Button
                          File: app.go:257
                          Purpose: "Reset array with new random weights"
                          State: text="Reset Array"
                          Bindings: OnTapped → resetArray() (analysis.go:193)
                          Bug: None

                - SettingsGroup:
                    Type: container.VBox
                    File: app.go:436-442
                    Components:
                      - arraySizeLabel:
                          Type: widget.Label
                          File: app.go:259
                          State: text="Array Size: 64x64", wrapping=off

                      - arraySizeSlider:
                          Type: widget.Slider
                          File: app.go:260-267
                          Purpose: "Adjust array dimensions (8-128)"
                          State: min=8, max=128, step=8, value=64
                          Bindings: OnChanged → recreateArray()
                          Bug: None

                - SignalGroup:
                    Type: container.VBox
                    File: app.go:444-453
                    Components:
                      - noiseLabel:
                          Type: widget.Label
                          File: app.go:269
                          State: text="Noise: 2.0%", wrapping=off

                      - noiseSlider:
                          Type: widget.Slider
                          File: app.go:270-276
                          Purpose: "Adjust read noise level (0-20%)"
                          State: min=0, max=20, step=0.5, value=2
                          Bindings: OnChanged updates config.NoiseLevel
                          Bug: None

                      - adcBitsLabel:
                          Type: widget.Label
                          File: app.go:278
                          State: text="ADC Bits: 6", wrapping=off

                      - adcBitsSlider:
                          Type: widget.Slider
                          File: app.go:279-286
                          Purpose: "Adjust ADC resolution (4-10 bits)"
                          State: min=4, max=10, step=1, value=6
                          Bindings: OnChanged updates config.ADCBits
                          Bug: None

                - DisplayGroup:
                    Type: container.VBox
                    File: app.go:463-469
                    Components:
                      - colormapSelect:
                          Type: widget.Select
                          File: app.go:288-292
                          Purpose: "Choose heatmap colormap"
                          State: options=["fecim", "viridis", "plasma", "coolwarm"], selected="fecim"
                          Bindings: OnChanged → SetColormap() on active heatmap
                          Bug: None

          - StatsSection:
              Type: container.VScroll
              File: app.go:489-490
              Purpose: "Display analysis results and cell details"
              Components:
                - statsLabel:
                    Type: widget.Label
                    File: app.go:335-337
                    Purpose: "Show detailed cell/analysis info"
                    State: wrapping=off, monospace=true
                    Bindings:
                      - Updated by onCellTapped (tooltips.go:12-63)
                      - Updated by analyzeIRDrop (analysis.go:98)
                      - Updated by analyzeSneakPaths (analysis.go:166)
                    Bug: None

    - Footer:
        Type: container.HBox
        File: app.go:512-520
        Purpose: "Status bar with mode, status, hover info"
        Components:
          - modeIndicator:
              Type: ModeIndicatorBox
              File: liveslide.go:55-88
              Purpose: "Show current operation mode (IDLE/COMPUTE/WRITE/READ/IR DROP/SNEAK)"
              State: mode=DemoModeIdle, colored background based on mode
              Bindings: SetMode() called by operation functions
              Bug: None

          - statusLabel:
              Type: widget.Label
              File: app.go:340
              Purpose: "Show current operation status"
              State: bold=true, wrapping=off
              Bindings: Updated by updateStatus()
              Bug: None

          - hoverInfoLabel:
              Type: widget.Label
              File: app.go:349-352
              Purpose: "Show cell details on mouse hover"
              State: monospace=true, wrapping=off, truncation=ellipsis
              Bindings: Updated by OnCellHover callbacks
              Bug: None

          - infoLabel:
              Type: widget.Label
              File: app.go:343-346
              Purpose: "Show array configuration info"
              State: wrapping=off
              Bindings: Updated by updateInfoLabel()
              Bug: None

## Enhanced Screen (Enhanced Mode)
Layout:
  - Type: container.Border
  - File: app_enhanced.go:430-439
  - Top: Header
  - Bottom: Status footer
  - Center: HSplit (leftCenterSplit → mainSplit with offset=0.75)

  Components:
    - CenterPanel (Enhanced):
        Type: container.AppTabs
        File: app_enhanced.go:120-127
        Purpose: "Extended tabs with color legends and comparison views"
        State: 6 tabs
        Components:
          - ConductanceTab:
              Type: container.TabItem
              Content: container.Border(right=condLegend, center=conductanceHeatmap)
              File: app_enhanced.go:82-87
              Components:
                - condLegend:
                    Type: ColorLegend
                    File: widgets.go:18-72, app_enhanced.go:41
                    Purpose: "Show color-to-level mapping for conductance"
                    State: minLabel="0", maxLabel="29", unit="Level", levels=30, colormap="fecim"
                    Bindings: SetColormap() synced with heatmap
                    Bug: None

          - IRDropTab:
              Type: container.TabItem
              Content: container.Border(right=irLegend, center=irDropHeatmap)
              File: app_enhanced.go:90-95
              Components:
                - irLegend:
                    Type: ColorLegend
                    File: app_enhanced.go:44
                    Purpose: "Show color-to-drop mapping"
                    State: minLabel="0%", maxLabel="100%", unit="Drop", colormap="viridis"
                    Bug: None

          - SneakPathTab:
              Type: container.TabItem
              Content: container.Border(right=sneakLegend, center=sneakPathHeatmap)
              File: app_enhanced.go:98-103
              Components:
                - sneakLegend:
                    Type: ColorLegend
                    File: app_enhanced.go:47
                    Purpose: "Show color-to-sneak mapping"
                    State: minLabel="Low", maxLabel="High", unit="Sneak", colormap="plasma"
                    Bug: None

          - IdealVsActualTab:
              Type: container.TabItem
              File: app_enhanced.go:106-110
              Purpose: "Side-by-side comparison of ideal vs actual conductance"
              Content: container.Border(top=titleLabel, center=beforeAfterToggle)
              Components:
                - beforeAfterToggle:
                    Type: BeforeAfterToggle
                    File: widgets.go:836-894
                    Purpose: "Split/before/after/diff comparison view"
                    State: mode="split", leftHeatmap, rightHeatmap, toggleGroup
                    Bindings:
                      - OnCellTapped → onBeforeAfterCellTapped (app_enhanced.go:721)
                      - OnCellHover → onBeforeAfterCellHover (app_enhanced.go:783)
                      - SetData() called by updateEnhancedWidgets (app_enhanced.go:629)
                    Bug: None

          - AccuracyAnalysisTab:
              Type: container.TabItem
              File: app_enhanced.go:113-117
              Purpose: "Step-by-step accuracy degradation waterfall"
              Content: container.Border(top=titleLabel, center=accuracyWaterfall)
              Components:
                - accuracyWaterfall:
                    Type: AccuracyWaterfall
                    File: widgets.go:483-569
                    Purpose: "Visualize cumulative accuracy loss from non-idealities"
                    State: steps=[], targetAccuracy=87.0 (Dr. Tour)
                    Bindings: SetSteps() called by updateEnhancedWidgets (app_enhanced.go:606)
                    Bug: None

    - RightPanel (Enhanced):
        Type: container.VSplit
        File: app_enhanced.go:390-391
        Purpose: "Controls + metrics/comparison"
        State: offset=0.5
        Components:
          - ControlsSection:
              Components:
                - runMVMButton:
                    Type: widget.Button
                    File: app_enhanced.go:236
                    State: text="Run Enhanced MVM", importance=HighImportance
                    Bindings: OnTapped → runEnhancedMVM() (app_enhanced.go:443)
                    Bug: None

                - exportButton:
                    Type: widget.Button
                    File: app_enhanced.go:241
                    Purpose: "Export weights and analysis to CSV/JSON"
                    Bindings: OnTapped → exportData() (app_enhanced.go:681)
                    Bug: None

          - MetricsSection:
              Type: container.VScroll
              File: app_enhanced.go:388-389
              Components:
                - statsLabel:
                    Type: widget.Label
                    File: app_enhanced.go:375-377
                    Purpose: "Detailed cell analysis"
                    State: wrapping=off, monospace=true

                - metricsPanel:
                    Type: MetricsPanel
                    File: widgets.go:273-320
                    Purpose: "Live accuracy, energy, performance metrics"
                    State: idealAccuracy, actualAccuracy, fecimEnergy, gpuEnergy
                    Bindings: UpdateMetrics() called by updateEnhancedWidgets (app_enhanced.go:584)
                    Bug: None

                - comparisonBadge:
                    Type: ComparisonBadge
                    File: widgets.go:405-440
                    Purpose: "FeCIM vs GPU energy comparison badge"
                    State: metric="Energy per 4096 MACs", fecimValue, gpuValue, improvement
                    Bindings: UpdateValues() called by updateEnhancedWidgets (app_enhanced.go:596)
                    Bug: None

# DataFlow

## MVM Execution Flow
Trigger: User clicks "Run MVM" button
Source: app.go:255, runMVMButton.OnTapped
Updates:
  - runMVM() → app.go:16 (animation.go)
  - Creates random input vector
  - Stores to ca.lastInput (protected by stateMu)
  - Updates mvmVis.SetInput() → vectors.go:420
  - Spawns goroutine runMVMAnimated() → animation.go:41
  - Phase 1: Highlight input columns (300ms)
  - Phase 2: Animate current flow (500ms)
  - Perform ca.array.MVM(input) → crossbar package
  - Phase 3: Highlight output rows (300ms)
  - Updates mvmVis.SetOutput() → vectors.go:426
  - Auto-runs runIRDropAnalysis() and runSneakPathAnalysis()
  - Updates statsLabel with results
  - Re-enables runMVMButton
File: animation.go:16-149
Bug: BUG-1 (lastInput write at line 31-32 should use stateMu.Lock, currently unprotected)

## Enhanced MVM Execution Flow
Trigger: User clicks "Run Enhanced MVM" button
Source: app_enhanced.go:236, runMVMButton.OnTapped
Updates:
  - runEnhancedMVM() → app_enhanced.go:443
  - Creates random input vector
  - Stores to ca.lastInput
  - Spawns goroutine runEnhancedMVMAnimated() → app_enhanced.go:500
  - Animation phases (same as standard MVM)
  - Performs ca.array.MVMWithNonIdealities(input, opts) → crossbar package
  - Stores result to ca.lastMVMResult (protected by stateMu)
  - Updates all enhanced widgets:
    - metricsPanel.UpdateMetrics()
    - comparisonBadge.UpdateValues()
    - accuracyWaterfall.SetSteps()
    - beforeAfterToggle.SetData()
    - irDropHeatmap.SetData()
    - sneakPathHeatmap.SetData()
File: app_enhanced.go:443-678
Bug: None

## Cell Selection Sync Flow
Trigger: User clicks a cell on any heatmap
Source: heatmap.go:228 (Tapped event)
Updates:
  - Heatmap calculates row/col from click position
  - Calls OnCellTapped callback
  - Callback invokes syncSelection(row, col) → callbacks.go:13
  - Updates ca.selectedRow, ca.selectedCol
  - Syncs selection to ALL heatmaps:
    - conductanceHeatmap.SetSelection(row, col)
    - irDropHeatmap.SetSelection(row, col)
    - sneakPathHeatmap.SetSelection(row, col)
    - beforeAfterToggle heatmaps (if exists)
  - Generates tooltip via ConductanceTooltip/IRDropTooltip/SneakPathTooltip
  - Updates statsLabel with detailed cell info
  - Updates statusLabel with summary
File: callbacks.go:13-31, heatmap.go:228-244
Bug: None

## Cell Hover Flow
Trigger: Mouse moves over heatmap
Source: heatmap.go:252 (MouseMoved event)
Updates:
  - Calculates row/col from mouse position
  - Calls OnCellHover callback
  - Callback updates hoverInfoLabel with compact cell info
  - Format: "[row,col] │ L<level> │ G=<conductance> µS │ R=<resistance> kΩ"
  - On mouse exit: Sets generic "Hover over cells..." message
File: callbacks.go:58-71, heatmap.go:252-276
Bug: None

## Tab Selection Flow
Trigger: User clicks a tab
Source: app.go:363, tabs.OnSelected
Updates:
  - Detects tab.Text (Conductance/IR Drop/Sneak Paths/Input/Output)
  - Syncs colormapSelect to current tab's colormap
  - Calls setEducationalContent() with tab-specific explanation
  - If cell selected: calls updateTooltipForTab() to refresh statsLabel
File: app.go:363-413, app_enhanced.go:130-222
Bug: None

## Array Resize Flow
Trigger: User moves arraySizeSlider
Source: app.go:263, arraySizeSlider.OnChanged
Updates:
  - Updates arraySizeLabel text
  - Calls recreateArray(size, noise, adcBits) → app.go:575
  - Creates new crossbar.Config
  - Creates new crossbar.Array
  - Calls heatmap.SetDimensions(size, size) for all heatmaps
  - Preserves heatmap widget references (no new widgets)
  - Calls programRandomWeights()
  - Updates all displays and labels
File: app.go:575-605
Bug: None

## Auto Demo Loop Flow
Trigger: User selects "Auto Demo" mode (deprecated in current version)
Source: animation.go:172, startAutoDemoLoop()
Updates:
  - Creates context.Context and cancel function
  - Stores to ca.autoCtx, ca.autoCancel
  - Creates 3-second ticker
  - Spawns goroutine autoDemoLoop()
  - Loop executes steps: MVM → IR Drop → Sneak Paths → Reset
  - Continues until context cancelled
  - Stopped by stopAutoDemoLoop() calling ca.autoCancel()
File: animation.go:172-249
Bug: BUG-5 (if Stop() not called, context may leak)

## Refresh Rate Limiting Flow
Trigger: Heatmap data updated
Source: heatmap.go:139, SetData()
Updates:
  - Calls rateLimitedRefresh() → heatmap.go:90
  - Checks time since last refresh
  - If <33ms elapsed: schedules delayed refresh (max 30 FPS)
  - Uses mutex to prevent concurrent refreshes
  - Spawns goroutine for delayed refresh if needed
  - Calls BaseWidget.Refresh() wrapped in fyne.Do()
File: heatmap.go:87-136
Bug: None

# BugDetails

BUG-1:
  component: runMVMAnimated
  severity: medium
  description: "lastInput write without mutex protection"
  expected: "ca.stateMu.Lock() before writing ca.lastInput"
  actual: "Direct write at animation.go:31-32 without lock"
  file: module2-crossbar/pkg/gui/animation.go:31-32
  suggested_fix: |
    // Before line 31:
    ca.stateMu.Lock()
    ca.lastInput = input
    ca.stateMu.Unlock()

BUG-2:
  component: updateConductanceDisplay
  severity: low
  description: "Missing fyne.Do wrapper for UI update from potentially background context"
  expected: "fyne.Do(func() { ca.conductanceHeatmap.SetData(matrix) })"
  actual: "Direct call ca.conductanceHeatmap.SetData(matrix) at app.go:621"
  file: module2-crossbar/pkg/gui/app.go:621
  suggested_fix: |
    func (ca *CrossbarApp) updateConductanceDisplay() {
        matrix := ca.array.GetConductanceMatrix()
        fyne.Do(func() {
            ca.conductanceHeatmap.SetData(matrix)
        })
    }

BUG-3:
  component: CrossbarHeatmap.rateLimitedRefresh
  severity: low
  description: "Heatmap refresh during startup can trigger layout oscillation"
  expected: "Skip refreshes during startup stabilization period"
  actual: "Refresh calls check IsStartupStabilizing() but some callers bypass it"
  file: module2-crossbar/pkg/gui/heatmap.go:92
  suggested_fix: "Already partially mitigated - ensure all SetData callers respect startup flag"

BUG-4:
  component: eduContentLabel
  severity: low
  description: "Educational content label wrapping disabled but content changes can still trigger layout"
  expected: "Fixed-size container or maximum content length enforcement"
  actual: "Wrapping=off at app_enhanced.go:228 but long content can still resize"
  file: module2-crossbar/pkg/gui/app_enhanced.go:228
  suggested_fix: |
    // Wrap in fixed-size container:
    leftPanelContent := container.NewVBox(
        ca.eduTitleLabel,
        widget.NewSeparator(),
        container.NewGridWrap(fyne.NewSize(200, 400), ca.eduContentLabel),
        ...
    )

BUG-5:
  component: Auto Demo Loop
  severity: medium
  description: "Auto demo context cancellation may leak if Stop() not called"
  expected: "Cleanup in window close handler or destructor"
  actual: "Relies on explicit Stop() call from embedded interface"
  file: module2-crossbar/pkg/gui/animation.go:172-207
  suggested_fix: |
    // Add cleanup to window lifecycle:
    ca.window.SetOnClosed(func() {
        ca.stopAutoDemoLoop()
    })
