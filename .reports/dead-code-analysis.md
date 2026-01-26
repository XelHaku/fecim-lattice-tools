# Dead Code Analysis Report

Generated: 2026-01-26

## Executive Summary

| Category | Count | Impact | Status |
|----------|-------|--------|--------|
| Orphaned Directory | 143 files | HIGH | **DELETED** |
| Unreachable Functions | 411 | MEDIUM | Remaining |
| Unused Standalone CLIs | 6 | LOW | Remaining |

## Cleanup Completed

### 1. DELETED: `module2-crossbar/pkg/_layers_experimental/`

- **143 Go files** (4.9 MB) implementing advanced neural network layers
- **Zero imports** found anywhere in codebase
- Contents: attention mechanisms, transformers, GNN, optical computing, cryo-RERAM
- **Verified:** Build passes, all tests pass after deletion

---

## CAUTION: May Be Intentionally Preserved

### 2. Standalone CLI Entry Points

These may be useful for headless/scripting use cases:

| File | Purpose | Status |
|------|---------|--------|
| `cmd/launcher/main.go` | Legacy multi-module launcher | Likely deprecated by unified app |
| `module1-hysteresis/cmd/hysteresis/main.go` | Standalone hysteresis CLI | May be useful |
| `module2-crossbar/cmd/inference/main.go` | CLI inference tool | May be useful |
| `module3-mnist/cmd/mnist/main.go` | MNIST CLI | May be useful |
| `module3-mnist/cmd/train-network/main.go` | Training CLI | May be useful |
| `module3-mnist/cmd/train-ptq/main.go` | Post-training quantization | May be useful |
| `module3-mnist/cmd/train-single-layer/main.go` | Single layer training | May be useful |
| `module3-mnist/cmd/mnist-gui/main.go` | Standalone MNIST GUI | Likely deprecated |

**Recommendation:** Keep for now unless confirmed unused.

---

## SAFE: Functions Never Called

The following 411 functions are detected as unreachable by Go's deadcode tool. They can be safely removed.

### Module 1: Hysteresis (79 functions)

**pkg/gui/gui.go:**
- `WaveformType.String` (line 146)
- `App.createSectionDivider` (line 478)
- `App.createHeader` (line 484)

**pkg/gui/labench.go:** (entire file appears unused)
- `NewLabBench`, `LabBench.HandleKeyPress`, `LabBench.increaseEField`
- `LabBench.decreaseEField`, `LabBench.increaseTemperature`, `LabBench.decreaseTemperature`
- `LabBench.increaseFrequency`, `LabBench.decreaseFrequency`, `LabBench.cycleWaveform`
- `LabBench.togglePause`, `LabBench.reset`, `LabBench.quit`
- `LabBench.StatusString`, `LabBench.ControlsHelp`, `formatFrequency`
- `ComputeTemperatureEffects`, `SliderValue.Normalized`, `SliderValue.SetFromNormalized`
- `SliderValue.DisplayString`, `LabBench.GetSliders`

**pkg/gui/overlay.go:** (entire file appears unused)
- `NewOverlay`, `Overlay.ToggleHelp`, `Overlay.UpdatePhysics`, `Overlay.RenderText`
- `Overlay.renderStatusBar`, `Overlay.renderParametersPanel`, `Overlay.renderPhysicsPanel`
- `renderSliderBar`, `renderPolarizationBar`, `NewConsoleOverlay`
- `ConsoleOverlay.Render`, `ConsoleOverlay.UpdatePhysics`

**pkg/gui/simulation.go:**
- `App.getCalibratedWriteField` (line 855)

**pkg/gui/theme.go:**
- `fixedWidthLayout.MinSize` (line 83)
- `fixedWidthLayout.Layout` (line 93)

**pkg/render/plot.go:** (entire file appears unused)
- All functions: `DefaultPlotConfig`, `GeneratePlotVertices`, `generateGridVertices`
- `generateAxisVertices`, `generateCurveVertices`, `generateMarkerVertices`
- `GenerateCellVertices`, `FormatAxisLabel`, `CalculateTickPositions`, `FormatTickLabel`

**pkg/render/render.go:** (largely unused)
- 25+ functions including entire `Renderer` type and helpers

**pkg/render/vulkan.go:**
- `VulkanRenderer.SetKeyCallback` (line 130)
- `VulkanRenderer.Stop` (line 1403)

**pkg/simulation/engine.go:**
- `Engine.SetVoltage`, `Engine.SetWaveform`, `Engine.SetAmplitude`
- `Engine.GetHysteresisData`, `Engine.RunRealtime`

---

### Module 2: Crossbar (185 functions)

**pkg/crossbar/array.go:**
- `Array.VMM` (line 165) - unused MVM variant

**pkg/crossbar/drift.go:**
- `DriftSimulator.SetWeightMatrix`, `DriftSimulator.SimulateRead`
- `DriftSimulator.RecordSnapshot`, `DriftSimulator.GetInitialLevel`, `DriftSimulator.Reset`

**pkg/crossbar/enhanced.go:** (DifferentialArray entirely unused)
- `NewDifferentialArray`, `DifferentialArray.ProgramSignedWeight`
- `DifferentialArray.ProgramSignedMatrix`, `DifferentialArray.MVM`
- `DifferentialArray.MVMWithNonIdealities`, `DifferentialArray.GetSignedWeight`
- `DifferentialArray.GetSignedWeightMatrix`, `DefaultWriteVerifyConfig`
- `Array.ProgramWeightVerified`

**pkg/crossbar/irdrop.go:**
- `IRDropSimulator.SetAllInputs`, `IRDropSimulator.CompareWithIdeal`

**pkg/crossbar/reference.go:** (entire file unused)
- `NewCPUReference`, `CPUReference.MVM`, `CPUReference.MVMWithQuantization`
- `CPUReference.MVMWithNoise`, `clamp01`, `CompareOutputs`, `VerifyMVM`, `BenchmarkMVM`

**pkg/crossbar/sneakpath.go:**
- `SneakPathAnalyzer.SetConductancePattern`, `SneakPathAnalyzer.GetSneakCurrentMap`
- `SneakPathAnalyzer.GetTopSneakPaths`, `SneakPathAnalyzer.AnalyzeArray`

**pkg/data/mnist.go:** (entire file unused)
- All 8 functions including `NewMNISTDataset`, `Download`, `Load`, `GetBatch`

**pkg/evaluation/accuracy.go:** (entire file unused)
- All 16 functions including `NewMetrics`, `Update`, `GetPrecision`, `GetRecall`, `GetF1Score`

**pkg/gui/**: 80+ unreachable functions in widgets, controls, tabs

**pkg/training/training.go:**
- `Trainer.GetBiases`

**pkg/visualization/heatmap.go:** (entire file unused)
- All functions including `NewCrossbarHeatmap`, `SetData`, `GenerateVertices`

**pkg/weights/**: (entire directory largely unused)
- `serialization.go`: 20+ functions
- `weights.go`: 20+ functions

---

### Module 3: MNIST (15 functions)

**pkg/gui/comparison_card.go:**
- `ComparisonCard.drawGlowCircle`, `ComparisonCard.drawLargeCheckmark`
- `ComparisonCard.drawLargeX`, `ComparisonCard.drawLargeDigit`

**pkg/gui/embedded.go:**
- `NewEmbeddedMNISTApp`, `EmbeddedMNISTApp.BuildContent`
- `EmbeddedMNISTApp.Start`, `EmbeddedMNISTApp.Stop`

**pkg/training/network.go:**
- `NewMNISTNetworkWithWeights`, `MNISTNetwork.GetBiases1`, `MNISTNetwork.GetBiases2`
- `MNISTNetwork.SetBias1`, `MNISTNetwork.SetBias2`
- `MNISTNetwork.GetHiddenActivations`, `MNISTNetwork.QuantizeWeightsTo30Levels`

**pkg/training/single_layer.go:**
- `SingleLayerNetwork.LoadWeights`

---

### Module 4: Circuits (2 functions)

**pkg/gui/font.go:**
- `drawSimpleChar`

**pkg/peripherals/analysis.go:**
- `ComputeTransferFunction`

---

### Module 5: Comparison (52 functions)

**pkg/comparison/architecture.go:**
- `Architecture.CalculateEfficiency`

**pkg/comparison/render.go:**
- `Renderer.RenderSummary`

**pkg/gui/**: 50+ functions in widgets, liveslide, market

---

### Module 6: EDA (20 functions)

**pkg/compiler/types.go:**
- `NewStorageConfig`, `NewMemoryConfig`

**pkg/export/def.go:**
- `ExportDEFWithConfig`

**pkg/export/lattice_generator.go:**
- `GenerateLattice`

**pkg/export/verilog.go:**
- `ExportVerilogWithConfig`, `findArrayDimensions`

**pkg/layout/**: (entire directory unused)
- `def_generator.go`: `GenerateDEF`
- `verilog_generator.go`: `GenerateVerilog`

**pkg/openlane/config.go:**
- 8 functions including `GetConfigPath`, `LoadConfig`, `SaveConfig`

**pkg/openlane/manager.go:**
- `Manager.GetDockerImageVersion`, `Manager.GetPDKSetupInstructions`

**pkg/validate/yosys.go:**
- `RunYosysCheck`

**pkg/validation/openlane.go:**
- `RunCellUsageReport`, `parseCellUsageOutput`, `IsOpenROADAvailable`

**pkg/validation/yosys.go:**
- `ValidateVerilog`

---

### Shared Package (32 functions)

**logging/logging.go:**
- `Logger.Trace`, `Logger.ValueChange`, `Logger.CheckboxChange`, `Logger.EntryChange`

**utils/font.go:**
- `DrawSimpleChar`

**widgets/adaptive_layout.go:**
- `AdaptiveLayout.onBreakpointChange`, `AdaptiveLayout.switchToMobile`
- `AdaptiveLayout.switchToDesktop`, `NewResponsiveSplit`, `ResponsiveSplit.ApplyBreakpoint`

**widgets/base_renderer.go:** (largely unused)
- 15+ helper functions

**widgets/debug.go:**
- 10+ debug helper functions

**widgets/resize_detector.go:** (entire file unused)
- All functions

---

## Recommended Cleanup Order

### Phase 1: SAFE (High Impact, No Risk)

1. **Delete `module2-crossbar/pkg/_layers_experimental/`** (143 files)
   ```bash
   rm -rf module2-crossbar/pkg/_layers_experimental
   ```

2. **Delete entire unused files:**
   - `module2-crossbar/pkg/data/mnist.go`
   - `module2-crossbar/pkg/evaluation/accuracy.go`
   - `module2-crossbar/pkg/visualization/heatmap.go`
   - `module1-hysteresis/pkg/gui/labench.go`
   - `module1-hysteresis/pkg/gui/overlay.go`
   - `module1-hysteresis/pkg/render/plot.go`
   - `module6-eda/pkg/layout/def_generator.go`
   - `module6-eda/pkg/layout/verilog_generator.go`
   - `shared/widgets/resize_detector.go`

### Phase 2: CAUTION (Review Before Delete)

3. **Standalone CLIs** - Confirm with user before removing:
   - `cmd/launcher/`
   - `module3-mnist/cmd/mnist-gui/`

4. **Widget helper functions** - May be useful for future development

### Phase 3: Individual Function Removal

5. Remove individual unused functions from:
   - `module2-crossbar/pkg/crossbar/enhanced.go` (DifferentialArray)
   - `module2-crossbar/pkg/crossbar/reference.go` (CPUReference)
   - `module2-crossbar/pkg/weights/` (entire directory needs review)

---

## Test Plan

Before each deletion:

```bash
# 1. Run tests to establish baseline
go test ./... 2>&1 | tee baseline-tests.txt

# 2. Make deletion
rm <file>

# 3. Verify build
go build ./cmd/fecim-visualizer

# 4. Re-run tests
go test ./... 2>&1 | tee post-deletion-tests.txt

# 5. Compare results
diff baseline-tests.txt post-deletion-tests.txt
```

---

## Metrics

| Metric | Before | After (Est.) |
|--------|--------|--------------|
| Total Go Files | ~250+ | ~100 |
| Dead Functions | 411 | 0 |
| Orphaned Packages | 1 (143 files) | 0 |
| Build Time | Baseline | -30% (est.) |
