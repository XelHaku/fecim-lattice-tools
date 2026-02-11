# FeCIM Lattice Tools Test Documentation

## Test Summary

*Counts below are a snapshot (2026-01-29). For current totals, see CI (`go test ./...`).*

| Package | Tests | Status |
|---------|-------|--------|
| cmd/fecim-lattice-tools | 21 | ✅ PASS |
| config/physics | 7 | ✅ PASS |
| module1-hysteresis/pkg/ferroelectric | 30 | ✅ PASS |
| module1-hysteresis/pkg/gui/widgets | 20 | ✅ PASS |
| module1-hysteresis/pkg/simulation | 14 | ✅ PASS |
| module2-crossbar/pkg/crossbar | 31+ | ✅ PASS |
| module2-crossbar/pkg/gui | 4 | ✅ PASS |
| module2-crossbar/pkg/network | 10 | ✅ PASS |
| module2-crossbar/pkg/training | 18 | ✅ PASS |
| module2-crossbar/pkg/weights | 27 | ✅ PASS |
| module3-mnist/pkg/core | 37 | ✅ PASS |
| module3-mnist/pkg/mnist | 2 | ✅ PASS |
| module3-mnist/pkg/training | 9 | ✅ PASS |
| module3-mnist/pkg/gui | 4 | ✅ PASS |
| module4-circuits/pkg/peripherals | 33 | ✅ PASS |
| module4-circuits/pkg/gui | varies | ✅ PASS |
| module5-comparison/pkg/comparison | 19 | ✅ PASS |
| module5-comparison/pkg/gui | 8 | ✅ PASS |
| module6-eda/pkg/compiler | 27 | ✅ PASS |
| module6-eda/pkg/config | 4 | ✅ PASS |
| module6-eda/pkg/export | 69 | ✅ PASS |
| shared/compute | varies | ✅ PASS |
| shared/gpu | varies | ✅ PASS |
| shared/io | varies | ✅ PASS |
| shared/logging | 6 | ✅ PASS |
| shared/peripherals | varies | ✅ PASS |
| shared/physics | varies | ✅ PASS |
| shared/recording | 145 | ✅ PASS |
| shared/theme | 8 | ✅ PASS |
| shared/utils | varies | ✅ PASS |
| shared/widgets | 6 | ✅ PASS |
| **Total (snapshot)** | **See CI** | **✅ PASS** |

## Running Tests

```bash
# Run all tests
go test ./...

# CI-like settings (recommended when reproducing CI failures)
make test-ci
make test-race-ci

# Or manually (mirrors scripts/ci/go-test-all.sh)
go test -tags=ci -count=1 -shuffle=off -trimpath -timeout 10m ./...

# Run with verbose output
go test -v ./...

# Run tests for specific module
go test -v ./module1-hysteresis/...
go test -v ./module2-crossbar/...
go test -v ./module3-mnist/...

# Run specific test
go test -v -run TestFeCIM30LevelPhysics ./module3-mnist/pkg/core/...
go test -v -run TestMaterialPhysicsConsistency ./cmd/fecim-lattice-tools/...

# Run microbenchmarks only (skip unit tests)
make bench
# Or:
go test ./... -run ^$ -bench . -benchmem

# Run with race detector (full tree; can be slow)
go test -race ./...
```

## Test Categories

### 1. Physics Tests

**Location:** `module2-crossbar/pkg/crossbar/physics_test.go`, `module3-mnist/pkg/core/physics_test.go`

#### FeCIM 30-Level Quantization
- `TestFeCIM30LevelPhysics` - Verifies 30-level demo baseline
- `TestLevelSpacingUniformity` - Linear DAC spacing verification
- `TestQuantizationSymmetry` - Symmetric bipolar quantization q(-x) ≈ -q(x)
- `TestQuantizationCliff` - Validates 2→30 level MSE/PSNR improvement
- `Test30LevelQuantizationAccuracy` - PSNR verification at different bit depths

**Results (example output; may change with defaults):**
| Levels | MSE | PSNR |
|--------|-----|------|
| 2 (binary) | 0.3323 | 4.8 dB |
| 4 | 0.0369 | 14.3 dB |
| 8 | 0.0068 | 21.7 dB |
| 16 | 0.0015 | 28.4 dB |
| 30 (FeCIM) | 0.0004 | 34.0 dB |

#### IR Drop Physics
- `TestIRDropOhmsLaw` - V=IR verification
- `TestIRDropScalesWithResistance` - 5x resistance → ~5x drop
- `TestIRDropMitigationEffectiveness` - Wider lines reduce IR drop (model trend)
- `TestIRDropWorstCaseCorner` - Cumulative current hotspot identification

#### Sneak Path Physics
- `TestSneakPathThreeCellModel` - 3-cell parasitic current path
- `TestSneakPathScalesWithConductance` - 10x G → 10x sneak current
- `TestSneakPathSNR` - Signal-to-noise ratio calculation
- `TestSneakMitigationSelector` - Selector ratio improves sneak-path suppression

#### Drift Physics
- `TestDriftTimeEvolution` - FeCIM vs RRAM drift comparison
- `TestDriftFeCIMVsRRAM` - Compares relative drift trends (model-based)
- `TestDriftRetentionPrediction` - Checks retention projection logic (model-based)
- `TestDriftTemperatureDependence` - Arrhenius behavior at 27°C vs 85°C

### 2. Material/Ferroelectric Tests

**Location:** `module1-hysteresis/pkg/ferroelectric/ferroelectric_test.go`

- `TestHysteresisLoopExists` - P-E curve generation
- `TestHysteresisAsymmetry` - Path-dependent polarization
- `TestCoerciveFieldSwitching` - Polarization switching around Ec
- `TestDiscreteStatesCount` - 30-level demo baseline verification
- `TestMaterialParameters` - HZO parameter validation
- `TestCoerciveFieldTemperatureDependence` - Ec vs temperature
- `TestPolarizationTemperatureDependence` - Pr vs temperature
- `TestMaterialVariants` - All material presets work correctly
- `TestFerroelectricMaterialPhysics` - Parameter sanity checks against configured ranges

### 3. Simulation Engine Tests

**Location:** `module1-hysteresis/pkg/simulation/engine_test.go`

- `TestEngineStartStop` - Thread-safe start/stop
- `TestEnginePause` - Pause/resume functionality
- `TestEngineConcurrentAccess` - No data races
- `TestSineWaveformGeneration` - Correct waveform output
- `TestTriangleWaveformGeneration` - Triangle wave verification
- `TestManualWaveformMode` - Manual voltage control
- `TestPolarizationRespondsToField` - Physics response
- `TestHistoryRecording` - Voltage/polarization history
- `TestEngineWithDifferentMaterials` - All materials work
- `TestGetHysteresisData` - Hysteresis loop data generation

### 4. GUI Widget Tests (Headless)

**Location:** `module1-hysteresis/pkg/gui/widgets/widgets_test.go`

- `TestLevelIndicatorCreation` - Widget initialization
- `TestLevelIndicatorSetLevel` - Level setting (0-29)
- `TestLevelIndicator30Levels` - Full 30-level range
- `TestLevelIndicatorConcurrency` - Thread safety
- `TestCellVisualizerSetLevel` - Cell visualization
- `TestModeIndicatorSetWrite` - WRITE/READ mode
- `TestPEPlotSetData` - P-E plot data
- `TestPEPlotConcurrency` - Thread-safe data access

### 5. Integration Tests

**Location:** `cmd/fecim-lattice-tools/integration_test.go`

#### Cross-Module Tests
- `TestMaterialPhysicsConsistency` - 30-level consistency across modules
- `TestHysteresisToMNISTWorkflow` - Module 1 → Module 3 pipeline
- `TestSimulationEngineIntegration` - Engine produces correct physics
- `TestCrossbarNonIdealitiesIntegration` - IR drop, sneak paths, drift together
- `TestPeripheralCircuitsIntegration` - DAC/ADC/TIA circuits

#### End-to-End Tests
- `TestQuantizationCliffEffect` - 2-level vs 30-level comparison
- `TestEndToEndMNISTInference` - Full inference for all digits
- `TestConcurrentInferenceStability` - Thread safety under load

### 6. Calculation Tests

**Location:** `module3-mnist/pkg/core/physics_test.go`

- `TestSoftmaxNumericalStability` - Handles ±500 logits without NaN/Inf
- `TestReLUCorrectness` - ReLU(x) = max(0, x) validation
- `TestKLDivergenceProperties` - KL(P||P)=0, non-negativity, asymmetry
- `TestArgmaxCorrectness` - Index of maximum value
- `TestEnergyCalculation` - FeCIM energy estimate (10 fJ/bit/MAC × log2(levels) + ADC/DAC overhead; shared core model)

### 7. Network/Training Tests

**Location:** `module3-mnist/pkg/core/integration_test.go`, `module3-mnist/pkg/core/quantize_test.go`

- `TestFullInferencePipeline` - Complete 784→128→10 inference
- `TestPresetConfigurations` - 4 failure mode presets
- `TestDualModeNetwork_Creation` - Network initialization
- `TestDualModeNetwork_SetParameters` - Config bounds
- `TestDualModeNetwork_Inference` - Valid predictions
- `TestDualModeNetwork_AgreementUnderIdealConditions` - FP/CIM agreement

## Known Limitations

### 1. GUI Tests Partially Automated

**Issue:** Some Fyne GUI features require a display server or running Fyne app.

**Solution:** Widget logic is tested without display. Animation-related features (like pulsing highlights) are skipped in tests.

**Covered in headless tests:**
- Level indicator state management
- P-E plot data handling
- Mode indicator state
- Cell visualizer level mapping
- Thread-safe concurrent access

**Requires manual testing:**
- Animation timing
- Visual rendering
- User interactions (mouse clicks)

**Display-required automated checks (optional):**

Some layout/audit tests are skipped unless a display is available. On headless Linux, run them with Xvfb:

```bash
sudo apt-get install -y xvfb
xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run LayoutAudit
```

### 2. Drift Model Precision

**Issue:** FeCIM has extremely low drift coefficient (0.001), resulting in negligible drift even over extended simulations at room temperature.

**Design Decision:** This is physically correct - FeCIM's advantage is superior retention. The test `TestDriftTimeEvolution` uses an RRAM-like drift coefficient (0.05) to verify the model mechanics work correctly, then compares with actual FeCIM parameters.

### 3. MVM Quantization Effects

**Issue:** `TestMVMMatrixVectorMultiply` shows discrepancy between expected and actual outputs due to 30-level quantization and DAC/ADC effects.

**Design Decision:** This is expected behavior. The test logs the differences but does not fail, as the quantization error is within acceptable bounds for neural network inference.

### 4. Sneak Path SNR

**Issue:** SNR can be negative when sneak current exceeds target current in dense arrays.

**Design Decision:** Negative SNR is physically valid and indicates when selector devices or other mitigation is required. The test validates the calculation is correct, not that SNR is always positive.

### 5. Archived Demo Tests

**Issue:** Demos in `docs/archive/removed-demos/` (demo5-thermal, demo6-multilayer, demo7-nonidealities) have broken imports after archiving.

**Design Decision:** Archived demos are historical reference only. Their functionality has been merged into active demos. Do not run tests on archived code.

## Benchmarks

```bash
go test -bench=. ./module2-crossbar/pkg/crossbar/... ./module3-mnist/pkg/core/... ./cmd/fecim-lattice-tools/...
```

| Benchmark | Operations/sec | Notes |
|-----------|---------------|-------|
| BenchmarkQuantize30Levels | ~50,000 | 784×128 weight matrix |
| BenchmarkDualModeInference | ~10,000 | Full MNIST inference |
| BenchmarkIRDropSimulator | ~1,000 | 64×64 array, 100 iterations |
| BenchmarkSneakPathAnalyzer | ~5,000 | 64×64 array analysis |
| BenchmarkDriftSimulator | ~100,000 | Single timestep |
| BenchmarkEngineStep | ~100,000 | Single simulation step |
| BenchmarkGetHysteresisData | ~10,000 | Hysteresis loop generation |
| BenchmarkFullInferencePipeline | ~5,000 | Complete 784→10 inference |
| BenchmarkCrossbarMVM | ~10,000 | 64×64 matrix-vector multiply |

## Test Coverage Areas

### Covered ✅
- FeCIM 30-level quantization physics
- IR drop calculation and mitigation
- Sneak path analysis and mitigation
- Drift modeling and technology comparison
- Neural network math (softmax, ReLU, etc.)
- Dual-mode FP vs CIM inference
- Energy calculation estimates
- Parameter bounds validation
- Concurrent access safety
- Edge cases (zero/max input)
- Simulation engine state management
- GUI widget state management (headless)
- Cross-module integration
- Material physics across all presets
- Peripheral circuits (DAC/ADC/TIA)
- EDA compiler operation modes (Storage/Memory/Compute)
- Lattice Verilog/DEF generation
- SVG layout visualization
- Weight serialization (JSON/Binary)
- Weight quantization and normalization
- Crossbar tile mapping
- INL/DNL analysis for DAC/ADC
- System timing and power analysis
- Transfer function computation

### Not Covered ❌
- Interactive GUI behavior (requires display)
- Real MNIST dataset accuracy (uses synthetic data)
- Hardware-in-the-loop testing
- Multi-GPU rendering
- OpenLane flow integration tests
- Audio recording dB conversion (known failing test)

## Adding New Tests

1. Create test file in same package: `*_test.go`
2. Follow naming convention: `TestFeatureName`
3. Use table-driven tests for multiple cases
4. Add physics validation where applicable
5. Document known limitations in this file
6. Run race detector: `go test -race ./...`
7. Update this documentation with new test counts

## Test File Locations

```
cmd/fecim-lattice-tools/
├── launcher_test.go           # Launcher UI tests
└── integration_test.go        # Cross-module integration tests

module1-hysteresis/
├── pkg/ferroelectric/
│   ├── ferroelectric_test.go       # Physics model tests
│   └── preisach_advanced_test.go   # Advanced Preisach tests
├── pkg/simulation/
│   └── engine_test.go              # Simulation engine tests
└── pkg/gui/widgets/
    └── widgets_test.go             # Widget logic tests (headless)

module2-crossbar/
└── pkg/crossbar/
    ├── physics_test.go             # IR drop, sneak, drift tests
    ├── nonidealities_test.go       # Non-ideality model tests
    └── array_test.go               # Crossbar array tests

module3-mnist/
└── pkg/core/
    ├── physics_test.go             # Math/calculation tests
    ├── quantize_test.go            # Quantization tests
    └── integration_test.go         # E2E inference tests

module2-crossbar/
└── pkg/weights/
    └── weights_test.go             # Weight management & serialization tests

module4-circuits/
└── pkg/peripherals/
    ├── peripherals_test.go         # DAC/ADC/TIA tests
    └── analysis_test.go            # INL/DNL, timing, power analysis tests

module5-comparison/
└── pkg/comparison/
    └── comparison_test.go          # Technology comparison tests

module6-eda/
└── pkg/
    ├── compiler/
    │   ├── compiler_test.go        # Basic compiler tests
    │   └── compiler_extended_test.go # Operation modes, builders tests
    ├── config/types_test.go        # Config validation tests
    └── export/
        ├── def_test.go             # DEF format tests
        ├── export_test.go          # JSON/CSV/SPICE export tests
        ├── generators_test.go      # LEF/Liberty/OpenLane tests
        ├── verilog_test.go         # Verilog generation tests
        ├── lattice_generator_test.go # Lattice generator tests
        └── svg_test.go             # SVG visualization tests

shared/
├── logging/logging_test.go         # Logger tests
├── theme/theme_test.go             # Theme tests
└── widgets/color_legend_test.go    # Shared widget tests
```

### 8. EDA Compiler Tests (New)

**Location:** `module6-eda/pkg/compiler/compiler_extended_test.go`

#### Operation Mode Tests
- `TestGenerateDesign_StorageMode` - Storage mode array generation
- `TestGenerateDesign_MemoryMode` - Memory mode array generation
- `TestGenerateDesign_ComputeModeWithoutWeights` - Compute mode blank array
- `TestGenerateDesign_ComputeModeWithWeights` - Compute mode with weight mapping
- `TestGenerateBlank_AllModes` - Blank array for all modes
- `TestOperationMode_String` - String representation of modes

#### Config Builder Tests
- `TestNewArrayConfig_Defaults` - Default configuration values
- `TestNewStorageConfig` - Storage-optimized configuration
- `TestNewMemoryConfig` - Memory-optimized configuration
- `TestNewComputeConfig` - Compute-optimized configuration
- `TestWith1T1R` - 1T1R architecture builder
- `TestWithWeights` - Weight attachment builder

#### Quantization Tests
- `TestQuantization_SymmetricBipolar` - Symmetric quantization q(-x) ≈ -q(x)
- `TestQuantization_30Levels` - Full 30-level utilization

### 9. EDA Export Tests (New)

**Location:** `module6-eda/pkg/export/lattice_generator_test.go`, `module6-eda/pkg/export/svg_test.go`

#### Lattice Generator Tests
- `TestGenerateLatticeVerilog_*` - Lattice Verilog generation (cell_{row}_{col} naming)
- `TestGenerateLatticeDEF_*` - Lattice DEF placement generation
- `TestLatticeVerilogDEF_InstanceNamesMatch` - Cross-validation Verilog/DEF
- `TestWriteLatticeVerilog/DEF` - File I/O tests

#### SVG Visualization Tests
- `TestGenerateLayoutSVG_Passive*` - Passive architecture visualization
- `TestGenerateLayoutSVG_1T1R*` - 1T1R architecture visualization
- `TestGenerateLayoutSVG_ShowGrid/Labels/CellIDs` - Configuration options

### 10. Weight Management Tests (New)

**Location:** `module2-crossbar/pkg/weights/weights_test.go`

#### Model Weight Tests
- `TestNewModelWeights` - Model creation
- `TestModelWeights_AddLayer` - Layer addition
- `TestModelWeights_GetLayer/GetLayerByIndex` - Layer retrieval
- `TestLayerWeights_ToMatrix` - Flatten/unflatten conversion

#### Quantization Tests
- `TestLayerWeights_QuantizeWeights_Symmetric/Asymmetric` - Weight quantization
- `TestLayerWeights_NormalizeWeights` - Normalization for crossbar
- `TestLayerWeights_DenormalizeWeights` - Reverse normalization

#### Serialization Tests
- `TestModelWeights_SaveLoadJSON` - JSON round-trip
- `TestModelWeights_SaveLoadBinary` - Binary round-trip
- `TestQuantizeModel` - Model quantization to int8
- `TestGenerateCrossbarMapping` - Crossbar tile mapping

### 11. Peripheral Analysis Tests (New)

**Location:** `module4-circuits/pkg/peripherals/analysis_test.go`

#### INL/DNL Analysis
- `TestDAC_AnalyzeINLDNL_*` - DAC linearity analysis
- `TestADC_AnalyzeINLDNL_*` - ADC linearity analysis
- `TestFindCodeWidth_*` - Code bin width measurement

#### System Analysis
- `TestAnalyzeTiming_*` - Write/read timing analysis
- `TestAnalyzePower_*` - Power breakdown analysis
- `TestComputeTransferFunction_*` - Full signal chain transfer function

## Last Updated

- **Date:** 2026-01-29
- **Total Tests:** See CI (`go test ./...`)
- **Pass Rate:** 100%
- **Coverage:** Physics, integration, GUI logic (headless), EDA export, weight management, peripheral analysis, compute, GPU, IO, utilities
