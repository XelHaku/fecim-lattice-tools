---
Module: module3-mnist
Name: MNIST Neural Network Visualization
Entry: cmd/mnist-gui/main.go
Package: module3-mnist/pkg/gui
Last Updated: 2026-01-26
---

Bugs:
  - [x] BUG-M3-001: Canvas mouse events may not use fyne.Do() consistently
  - [x] BUG-M3-002: Auto-demo loop cleanup - ticker and context not fully synchronized
  - [x] BUG-M3-003: Network weight loading failure handling - VERIFIED (2026-01-26): Already implemented with dialogs in dualmode_inference.go:384-453

Variants:
  - MNISTApp: Original single-mode MNIST demo (app.go)
  - DualModeApp: FP vs CIM comparison mode (dualmode.go)

Screens:
  # === MNISTApp (Original) ===
  - name: MNISTApp_MainLayout
    description: 3-column responsive layout with draw/predict/metrics
    layout:
      type: Border
      top: Header
      bottom: Footer
      center:
        type: Stack
        views:
          - DrawView (default)
          - MetricsView (toggle)
    components:
      - type: Header
        purpose: Title, view selector buttons, network info
        file: app.go:336-352
        state:
          - currentView: int (0=draw, 1=metrics)
        children:
          - titleLabel (Label): "FeCIM MNIST Neural Network"
          - viewSelectorRow (HBox):
              - drawBtn (Button): Switch to draw view
              - metricsBtn (Button): Switch to metrics view
              - networkInfo (Label): "784 -> 128 -> 10 | 87% accuracy | 30 Levels"

      - type: LeftPanel
        purpose: Drawing canvas + mode selector + controls
        file: app.go:367-382
        layout: Border (top: mode/title, bottom: actions, center: canvas)
        children:
          - demoModeSelect (Select): "Manual" | "Auto Demo" | "Step-by-Step"
            file: app.go:257-261
            bindings:
              - onChanged: onDemoModeChanged() [app.go:833]
            bug: BUG-M3-002 (auto-demo cleanup)
          - digitCanvas (DigitCanvas): 28x28 interactive drawing
            file: canvas.go:24-47
            size: 350x350
            state:
              - pixels: [28][28]float64
              - activePixels: int (count > 0.1)
              - brushRadius: float64 (1.0/1.5/2.5)
              - brushSize: BrushSize enum
            bindings:
              - OnDigitChanged: onDigitChanged() [app.go:502]
              - OnPixelCountChanged: update label [dualmode.go:405]
            events:
              - Tapped: draw() [canvas.go:268]
              - Dragged: draw() [canvas.go:278]
              - TappedSecondary: Clear() [canvas.go:273]
              - MouseDown/MouseUp: desktop.Mouseable [canvas.go:287-296]
            bug: BUG-M3-001 (mouse events)
          - buttonGrid (Grid 2x2):
              - clearBtn: Clear canvas
              - randomBtn: Load random test digit
              - loadTestBtn: Load MNIST test data (async)
              - evalBtn: Evaluate network on test set (async)

      - type: CenterPanel
        purpose: Network activation visualization
        file: app.go:384-400
        layout: VSplit (60% layer view, 40% output chart)
        children:
          - layerView (LayerActivationView):
              file: activations.go:17-32
              purpose: Show input/hidden/output layer activations
              state:
                - inputLayer: []float64 (784)
                - hiddenLayer: []float64 (128)
                - outputLayer: []float64 (10)
              layout: HSplit (input | hidden | output)
              children:
                - inputRaster: 28x28 pixel grid visualization
                - hiddenRaster: Grid layout for 128 neurons
                - outputRaster: Bar chart for 10 classes
                - predictionLabel: "Prediction: X"
                - confidenceLabel: "Confidence: XX.X%"
          - outputChart (OutputBarChart):
              file: activations.go:386-403
              purpose: Standalone output probability bars
              state:
                - values: []float64 (10 classes)
                - predicted: int
              rendering: Bar chart with predicted class highlighted

      - type: RightPanel
        purpose: Educational content + operation log + key stat
        file: app.go:403-412
        layout: Border (top: educational, bottom: key stat, center: log)
        children:
          - educationalPanel (MNISTEducationalPanel):
              file: liveslide.go:207-230
              purpose: Context-sensitive explanations
              state:
                - title: string
                - content: string
              methods:
                - SetInferenceExplanation(phase): Explain inference phase 1-3
                - SetIdleExplanation(): Show welcome message
                - SetEvaluationExplanation(): Explain evaluation
              rendering: Title + wrapped text
          - operationLog (MNISTOperationLog):
              file: liveslide.go:340-365
              purpose: Timestamped operation history
              state:
                - entries: []string (max 10)
                - startTime: time.Time
              methods:
                - Add(entry): Append with timestamp "t=X.Xs >> ..."
                - AddPrediction(pred, conf): Log prediction result
          - keyStat (MNISTKeyStat):
              file: liveslide.go:614-633
              purpose: Prominent accuracy metric
              state:
                - label: "Target Accuracy"
                - value: "87%"
              bindings:
                - SetValue(): Update after evaluation

      - type: Footer
        purpose: Mode indicator + status + hover info + network info
        file: app.go:440-452
        layout: HBox with separators
        children:
          - modeIndicator (MNISTModeIndicator):
              file: liveslide.go:48-92
              purpose: Show current mode with colored background
              state:
                - mode: MNISTMode enum (IDLE/DRAWING/INFERENCE/EVALUATING/LOADING)
              rendering: Colored rectangle with mode text
              colors:
                - IDLE: gray
                - INFERENCE: blue
                - EVALUATING: orange
                - LOADING: green
          - statusLabel: Current operation status
          - hoverInfoLabel: Activation info on mouse hover
          - infoLabel: "Network: 784→128→10 | Levels: 30 | Target: 87%"

      - type: MetricsView
        purpose: Evaluation results (toggle view)
        file: app.go:414-424
        layout: VSplit (50% confusion matrix, 50% metrics)
        children:
          - confusionMatrix (ConfusionMatrix):
              file: metrics.go:17-41
              purpose: Interactive 10x10 confusion matrix
              state:
                - matrix: [10][10]int (actual x predicted)
                - total: int
                - selectedRow, selectedCol: int
              bindings:
                - OnCellTapped: onConfusionCellTapped() [app.go:694]
              rendering:
                - Green diagonal: correct predictions
                - Red off-diagonal: misclassifications
                - Color intensity: count magnitude
              events:
                - Tapped: Select cell and show class stats
          - metricsPanel (MetricsPanel):
              file: metrics.go:268-289
              purpose: Per-class precision/recall/F1 bars
              state:
                - precision, recall, f1: [10]float64
                - accuracy, avgF1: float64
              rendering:
                - 3 horizontal bars per class (color-coded)
                - Legend: Blue=precision, Orange=recall, Green=F1
          - classStatsPanel (ClassStatsPanel):
              file: metrics.go:434-453
              purpose: Detailed stats for selected class
              state:
                - selectedClass: int
                - precision, recall, f1: float64
                - tp, fp, fn, support: int
              rendering: Monospace text with class metrics

  # === DualModeApp (FP vs CIM Comparison) ===
  - name: DualModeApp_MainLayout
    description: 4-zone responsive layout (draw, results, controls, weights)
    layout:
      type: AdaptiveLayout
      breakpoints:
        - Desktop: HSplit (left + center | right)
        - Mobile: Tabs (Draw, Results, Config, Weights)
    file: dualmode.go:258-333
    components:
      - type: Header
        purpose: Title + Quick Demo + Guided Tour + Info buttons
        file: dualmode.go:336-391
        children:
          - titleLabel: "MNIST FeCIM | 784→128→10 | 30 Levels | 87% Target"
          - quickDemoBtn (WarningImportance):
              purpose: 30-second automated demo
              file: dualmode.go:344-352
              bindings:
                - OnTapped: StartQuickDemo() [dualmode.go:966]
              state:
                - quickDemoRunning: bool
                - quickDemoStopChan: chan struct{}
                - animationEnabled: bool
          - tourBtn (HighImportance):
              purpose: 5-step guided tour
              file: dualmode.go:355-361
              bindings:
                - OnTapped: tour.Start() [tour.go:96]
          - infoButtons (HBox):
              - why30Btn: ShowWhy30LevelsDialog() [dialogs.go:13]
              - realityBtn: ShowHardwareRealityDialog() [dialogs.go:64]
              - failuresBtn: ShowFailureModesDialog() [dialogs.go:118]
              - aboutBtn: ShowAboutDialog() [dialogs.go:178]

      - type: Zone1_DrawingCanvas
        purpose: Interactive digit drawing
        file: dualmode.go:394-445
        children:
          - digitCanvas (DigitCanvas): Same as MNISTApp [canvas.go:24]
          - pixelCountLabel: "Pixels: X/784"
          - brushSelect: "Thin" | "Medium (Recommended)" | "Thick"
            bindings:
              - SetBrushSize(BrushThin/Medium/Thick)
          - clearBtn, randomBtn: Actions

      - type: Zone2_Results
        purpose: FP vs CIM comparison + probability distribution
        file: dualmode.go:450-478
        layout: Border (top: comparison card, center: probability chart)
        children:
          - comparisonCard (ComparisonCard):
              file: comparison_card.go:39-58
              purpose: CIM-focused prediction display with FP as validation reference
              state:
                - result: *ComparisonResult
              rendering:
                - Single card layout with CIM as hero element
                - CIM prediction: Large 5x scale digit in cyan/green
                - FP reference: Small text below CIM ("✓ Matches FP" or "⚠ FP predicts: X")
                - Color-coded background tint: Green for match, Amber for mismatch
                - Thick border: Green (match) or Amber (mismatch)
                - Large confidence bar: Horizontal gradient fill for CIM only
                - Energy display: Compact "Ex more efficient than GPU" format
                - Second-best prediction: CIM 2nd choice only
              layout: Fixed height ~320px
          - dualProbabilityChart (DualProbabilityChart):
              file: comparison_card.go:547-571
              purpose: Probability divergence visualization
              state:
                - fpProbs, cimProbs: []float64 (10 classes)
                - divergences: []float64
                - fpPred, cimPred: int
              rendering:
                - Paired bars per digit (FP blue, CIM green)
                - Yellow marker: Divergence > 2%
                - Legend: "FP | CIM"

      - type: Zone3_Controls
        purpose: Hardware configuration panel
        file: dualmode.go:482-624
        state:
          - network: *core.DualModeNetwork
            - numLevels: int (2-31)
            - noiseLevel: float64 (0.0-0.20)
            - adcBits, dacBits: int
            - singleLayer: bool (Tour Mode)
        children:
          - levelsSlider (2-31):
              file: dualmode.go:487-506
              bindings:
                - OnChanged: SetNumLevels() + tryLoadQATWeights()
              state:
                - currentQATLevel: int (10/20/29/30/31)
              bug: BUG-M3-003 (weight loading)
          - noiseSlider (0.0-0.20):
              file: dualmode.go:514-526
              bindings:
                - OnChanged: SetNoiseLevel()
          - adcSelect, dacSelect (3-16 bits):
              file: dualmode.go:535-555
              bindings:
                - OnChanged: SetADCBits() / SetDACBits()
          - hiddenSelect (64/128/256):
              file: dualmode.go:557-563
              bindings:
                - OnChanged: changeHiddenSize() (async)
          - presetButtons (Grid 3+2):
              - Row 1: Ideal | QuantCliff | Noisy
              - Row 2: BrokenADC | Tour (HighImportance)
              file: dualmode.go:576-603
              bindings:
                - applyPresetWithMode(levels, noise, adc, dac, singleLayer)
          - testButton: "Test (200)"
              file: dualmode.go:610-613
              bindings:
                - OnTapped: runQuickTest() [dualmode.go:1187]
              children:
                - testProgressBar: Hidden until test starts
                - testResultLabel: "FP: X% | CIM: Y% | Agreement: Z%"

      - type: Zone4_Weights
        purpose: Weight visualization tabs
        file: dualmode.go:628-724
        layout: AppTabs (5 tabs)
        children:
          - Tab1: Quantized Heatmap
              - weightHeatmap (Raster): Blue-white-red colormap
                file: dualmode.go:1253-1317
                bindings:
                  - drawWeightHeatmap(): Generate heatmap image
              - weightLegend (ColorLegend): -1.0 to 1.0
              - layerSelect (Radio): "Layer1 (784x128)" | "Layer2 (128x10)"
              - zoomBtn: Open new window with larger heatmap
              - labels:
                  - weightDimLabel: "Dimensions: RxC"
                  - weightRangeLabel: "Range: [min, max]"
                  - weightLevelsLabel: "Levels: X/30"
          - Tab2: FP vs Quant Comparison
              - weightComparisonWidget (WeightComparisonWidget):
                  file: weight_comparison.go:19-55
                  purpose: Side-by-side FP and quantized weights
                  state:
                    - fpWeights, quantWeights: [][]float64
                    - showMode: int (0=FP, 1=Quantized, 2=Difference)
                    - meanError, maxError, errorStdDev: float64
                  rendering:
                    - Mode 0/1: Blue-white-red colormap
                    - Mode 2: Black-yellow-red error gradient
                  controls:
                    - modeSelect: "FP (Float32)" | "Quantized" | "Difference"
          - Tab3: Side-by-Side
              - dualWeightHeatmap (DualWeightHeatmap):
                  file: weight_comparison.go:330-378
                  purpose: Split view (FP left, Quantized right)
                  rendering: Divider line between halves
          - Tab4: Quantization (P1.1 Enhancement)
              - quantizationWidget (QuantizationWidget):
                  file: quantization_widget.go:30-56
                  purpose: Real-time weight quantization visualization
                  state:
                    - samples: []QuantizationSample (max 5)
                    - numLevels: int
                    - totalError: float64
                  rendering:
                    - Per-sample row: FP → Arrow → Quantized (Level X)
                    - Level indicator bar: Position in 0-29 range
                    - Error statistics: "Precision loss: X% | Impact: Y"
          - Tab5: Energy (P1.3 Enhancement)
              - energyWidget (EnergyWidget):
                  file: energy_widget.go:37-79
                  purpose: Energy efficiency comparison
                  state:
                    - config: EnergyConfig (50 fJ/MAC FeCIM, 2000 fJ/MAC GPU)
                    - totalInferences: int
                    - totalEnergyFeCIM, totalEnergyGPU: float64 (Joules)
                    - lastEnergyFeCIM, lastEnergyGPU: float64 (nJ)
                    - efficiencyRatio: float64 (GPU/FeCIM)
                  rendering:
                    - Per-inference bars: GPU (full width) vs FeCIM (proportional)
                    - Efficiency highlight box: "XXx MORE EFFICIENT"
                    - Session totals: "X inferences | FeCIM: Y | GPU: Z"

      - type: Footer
        purpose: Status label
        file: dualmode.go:286
        children:
          - statusLabel: Operation status and feedback

DataFlow:
  - name: DigitDrawing_to_Inference
    trigger: User draws on canvas
    source: DigitCanvas.draw() [canvas.go:299]
    updates:
      - pixels: [28][28]float64
      - activePixels: int
    calls:
      - notifyChange() → OnDigitChanged() [canvas.go:336]
    file: canvas.go:299-333
    threading: Main thread (Fyne UI)

  - name: MNISTApp_Inference
    trigger: OnDigitChanged(pixels []float64)
    source: app.go:502
    steps:
      - runInferenceAnimated() in goroutine [app.go:510]
      - Phase 1 (200ms): "Processing input pixels..."
        - fyne.Do: Update mode indicator, status, educational panel
      - Phase 2 (300ms): "Hidden layer MVM (784→128)..."
        - GetLayerActivations(pixels) [training package]
        - fyne.Do: Update layer view
      - Phase 3 (200ms): "Output layer MVM (128→10)..."
        - fyne.Do: Update output chart
      - Phase 4: Display results
        - Predict(pixels) → pred, conf
        - fyne.Do: Update prediction labels, status, operation log
    file: app.go:502-563
    threading: Goroutine with fyne.Do() for UI updates

  - name: DualModeApp_Inference
    trigger: OnDigitChanged(pixels []float64)
    source: dualmode.go:748
    steps:
      - Check animationEnabled flag
      - If enabled: runInferenceAnimated() [dualmode.go:846]
      - If disabled: runInference() [dualmode.go:759]
      - network.Infer(pixels) → InferenceResult
      - fyne.Do: Update comparison card, probability chart, energy widget
    file: dualmode.go:748-842
    threading: Mixed (main for sync, goroutine for animated)

  - name: AutoDemo_Loop
    trigger: demoModeSelect = "Auto Demo"
    source: app.go:833 → startAutoDemoLoop() [app.go:859]
    state:
      - autoDemo: bool (protected by autoDemoMu)
      - autoDemoTimer: *time.Ticker (2s interval)
      - autoDemoCtx, autoDemoCancel: context.Context
    flow:
      - Start ticker (2s)
      - Launch autoDemoLoop() goroutine [app.go:902]
      - Loop: Wait for ticker OR context.Done()
        - fyne.Do(loadRandomTestDigit)
    cleanup:
      - stopAutoDemoLoop() [app.go:881]
        - Cancel context
        - Stop ticker
    file: app.go:859-926
    bug: BUG-M3-002 (ticker not stopped before cancel)

  - name: QuickTest_Evaluation
    trigger: testButton.OnTapped
    source: dualmode.go:1187
    steps:
      - Disable button, show progress bar
      - Launch goroutine
      - Load test data if not loaded
      - Loop 200 samples:
        - network.Infer(testImages[i])
        - Count correct predictions
        - Every 10 samples: fyne.Do(update progress bar)
      - fyne.Do: Hide progress bar, show results
    file: dualmode.go:1187-1250
    threading: Goroutine with fyne.Do() for UI updates

  - name: QAT_Weight_Loading
    trigger: levelsSlider.OnChanged
    source: dualmode.go:492
    steps:
      - GetBestMatchingWeightsLevel(levels) → 10/20/29/30/31
      - tryLoadQATWeights(bestLevel) [dualmode.go:1470]
      - Check if already loaded (currentQATLevel == targetLevel)
      - GetWeightsFilename(dataDir, targetLevel)
      - os.Stat(weightsPath) - check if file exists
      - network.LoadWeights(weightsPath)
      - Update currentQATLevel
    file: dualmode.go:492-506, dualmode.go:1470-1500
    bug: BUG-M3-003 (no error dialog on failure)

BugDetails:
  - id: BUG-M3-001
    component: DigitCanvas
    severity: Medium
    description: Mouse event handlers may not consistently use fyne.Do() for UI updates
    expected: All UI updates from mouse events wrapped in fyne.Do()
    actual: draw() calls fyne.Do() but may be called from desktop.Mouseable handlers
    file: canvas.go:287-333
    suggested_fix: |
      Audit all paths to draw() and ensure fyne.Do() wrapping:
      - Tapped → draw() [canvas.go:268]
      - Dragged → draw() [canvas.go:278]
      - MouseDown → draw() [canvas.go:290]

      Current: draw() has fyne.Do() internally (line 328)
      Risk: If draw() is async, nested fyne.Do() may cause issues

      Fix: Ensure draw() is only called from main thread OR remove internal fyne.Do()

  - id: BUG-M3-002
    component: MNISTApp auto-demo loop
    severity: Low
    description: Auto-demo cleanup not fully synchronized (ticker stopped after cancel)
    expected: Context cancelled first, then ticker stopped
    actual: Ticker may fire after context cancel but before Stop()
    file: app.go:881-900
    suggested_fix: |
      In stopAutoDemoLoop():
      1. Call autoDemoCancel() first (signal goroutine)
      2. Wait briefly for goroutine to exit
      3. THEN call autoDemoTimer.Stop()

      Current order:
      - Line 890: autoDemoCancel()
      - Line 896: autoDemoTimer.Stop()

      Risk: Timer fires between cancel and Stop, triggers fyne.Do(loadRandomTestDigit)

      Fix:
      ```go
      if ma.autoDemoCancel != nil {
        ma.autoDemoCancel()
        time.Sleep(10 * time.Millisecond) // Let goroutine exit
      }
      if ma.autoDemoTimer != nil {
        ma.autoDemoTimer.Stop()
      }
      ```

  - id: BUG-M3-003
    component: DualModeApp QAT weight loading
    severity: Low
    description: Weight loading failure not reported to user (silent failure)
    expected: Dialog shown if weight file not found or load fails
    actual: Errors silently ignored (early return)
    file: dualmode.go:1470-1500
    suggested_fix: |
      In tryLoadQATWeights():
      - If os.Stat fails: Show warning dialog "QAT weights for {level} levels not found"
      - If LoadWeights fails: Show error dialog "Failed to load weights: {err}"

      Current behavior:
      - Line 1483: File not found → return (silent)
      - Line 1489: Load failed → return (silent)

      Fix:
      ```go
      if _, err := os.Stat(weightsPath); os.IsNotExist(err) {
        fyne.Do(func() {
          dialog.ShowInformation("Weights Not Found",
            fmt.Sprintf("QAT weights for %d levels not available. Using default.", targetLevel),
            app.window)
        })
        return
      }
      ```
