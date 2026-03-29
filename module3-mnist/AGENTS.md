<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module3-mnist

## Purpose

MNIST digit recognition module using ferroelectric crossbar arrays for inference. Demonstrates neural network inference with configurable quantization levels, noise injection, and hardware-in-the-loop simulation with peripheral circuits (DAC, ADC, TIA). Provides both training tools and interactive GUI with quick demo mode.

## Key Files

| File | Purpose |
|------|---------|
| `pkg/core/network.go` | Network configuration and inference path |
| `pkg/gui/dualmode_demo.go` | Quick demo (~30s) showing 30→2 level degradation |
| `pkg/gui/gui.go` (DualModeApp) | Main Fyne app with training and inference tabs |
| `pkg/training/network.go` | Full 2-layer network training (784→256→10) |
| `pkg/training/single_layer.go` | Single-layer network (784→10) for hardware calibration |
| `pkg/mnist/loader.go` | MNIST dataset loader |

## Subdirectories

| Directory | Purpose | Key Files |
|-----------|---------|-----------|
| `cmd/mnist/` | Standalone CLI entry point | `main.go` |
| `pkg/core/` | Network configuration and inference interface | `network.go`, `inference.go` |
| `pkg/gui/` | Fyne GUI for training and inference visualization | `dualmode_demo.go`, `dualmode_controls.go`, `keyboard.go` |
| `pkg/mnist/` | MNIST dataset loading | `loader.go` (with bounds validation) |
| `pkg/training/` | Network training (full and single-layer) | `network.go`, `single_layer.go`, `trainer_foundation.go` |
| `data/` | MNIST dataset files (train/test CSV) | `mnist_train.csv`, `mnist_test.csv` |
| `scripts/` | Training helper scripts | `*.sh` |

## For AI Agents

### Working In This Directory

- **Network config**: `NetworkConfig` struct in `pkg/core/network.go` controls quantization, noise, peripherals.
- **Quantization levels**: Default 30 (conference claim). Configurable per-layer via `PerLayerQuant` flag.
- **Inference path**: Two engines:
  1. **Full-precision**: No quantization (baseline)
  2. **CIM (Compute-in-Memory)**: Quantized with noise and peripheral effects
- **Peripheral effects**: DAC bits, ADC bits, TIA feedback/capacitance, ADC conversion time
- **Noise decomposition**: ADC noise, thermal, flicker, cell variation (σ/μ per component)
- **Training**: Two modes:
  1. **Full network**: 784→256→10 (2 hidden layers)
  2. **Single-layer**: 784→10 (hardware calibration mode)

### Testing Requirements

```bash
go test ./module3-mnist/...
go test ./module3-mnist/pkg/core/...
go test ./module3-mnist/pkg/training/...
```

Key test files:
- `pkg/training/network_test.go` - Full network training and inference
- `pkg/training/training_test.go` - Quantization and noise injection
- `pkg/training/extended_test.go` - Edge cases and stress tests
- `pkg/mnist/loader_test.go` - Dataset loading and bounds validation
- `pkg/core/inference_test.go` - Inference path with various configs

### Common Patterns

- Default config: `core.DefaultNetworkConfig()` returns optimal FeCIM settings
- Inference: `network.Infer(inputVector)` returns predicted digit (0-9)
- Quantization: `QuantizeValue(value, numLevels)` for per-layer quantization
- Noise injection: Applied via `NoiseLevel` (σ/μ coefficient, typical 0.01-0.20)
- Single-layer training: `training.NewSingleLayerNetwork()` for hardware demo
- Full network training: `training.NewNetwork()` then `network.Train(epochs)`

### Critical Bug Patterns (Do Not Re-Introduce)

1. **Quantization Mismatch**: Per-layer quantization requires separate models. Fixed: Validate Layer1Levels ≠ Layer2Levels in config.
2. **Noise Composition**: Noise σ/μ should be decomposed into components. Fixed: Use NoiseADC, NoiseThermal, NoiseFlicker, NoiseCellVariation.
3. **TIA Feedback Effects**: TIA can introduce non-linear clipping. Fixed: Validate TIAFeedbackOhms and TIACapF against known hardware.
4. **ADC Saturation**: ADC quantization can saturate at rail. Fixed: Clamp output to ADC range before rounding.
5. **Single-Layer Mode**: Must use 784→10 topology. Fixed: Validate SingleLayer flag enforces correct layer sizes.

## Dependencies

### Internal

- `module2-crossbar/` - Crossbar array simulation and MVM
- `module1-hysteresis/` - Ferroelectric material models (for understanding quantization)
- `shared/physics/` - Material presets, quantization utilities
- `shared/widgets/` - GUI components (sliders, progress bars, charts)
- `shared/theme/` - Styling
- `shared/logging/` - Structured logging
- `shared/peripherals/` - DAC, ADC, TIA models (if available)

### External

- Fyne v2 (`fyne.io/fyne/v2`) - Cross-platform native GUI
- gonum (`gonum.org/v1/gonum`) - Linear algebra (matrix-vector multiplication, activation functions)
- Standard Go `math` package - Sigmoid and other nonlinearities

## Configuration

Network settings in `pkg/core/NetworkConfig`:
- `NumLevels` - Quantization levels (1-30, default 30)
- `NoiseLevel` - Overall σ/μ noise coefficient (0.0-0.20, default 0.01)
- `ADCBits`, `DACBits` - Peripheral resolution (typical 8)
- `EnableSneak` - Enable sneak path simulation
- `IRDrop` - Enable IR drop simulation
- `SingleLayer` - Use 784→10 topology for hardware calibration
- `PerLayerQuant` - Enable per-layer quantization (Layer1Levels, Layer2Levels)
- Peripheral parameters: `TIAFeedbackOhms`, `TIACapF`, `TIAGBWHZ`, `ADCConversionTimeS`

## Training & Data

- **MNIST dataset**: `data/mnist_train.csv` and `data/mnist_test.csv`
- **Dataset format**: CSV with pixel values (0-255) and label (0-9)
- **Training epochs**: Typical 10-50 for good convergence
- **Batch processing**: Optional (framework supports mini-batches)
- **Weights persistence**: Saved to JSON via `training.SaveWeights(filename)`

## Performance Notes

- Inference: ~1-10ms per digit (CPU, quantized)
- Training: ~100-500ms per epoch (784→256→10 network, 60K samples)
- Single-layer training: ~10-50ms per epoch (784→10, faster for calibration)
- Memory: ~10-20MB for full network weights (784×256 + 256×10)
- GPU acceleration: Optional via module2-crossbar GPU path

## Quick Demo

The GUI includes a `StartQuickDemo()` method that:
1. Loads MNIST sample
2. Shows 30-level ideal prediction (success)
3. Breaks to 2 levels
4. Shows failure
5. Restores and explains FeCIM insight
- Duration: ~30 seconds
- Automated with status updates every 1-4 seconds

<!-- MANUAL: -->
