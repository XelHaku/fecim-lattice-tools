# MNIST Module Developer Guide

**Version:** 1.0
**Date:** 2026-01-27

This guide covers common development tasks for the MNIST module.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Adding New Features](#adding-new-features)
3. [Testing](#testing)
4. [Common Patterns](#common-patterns)
5. [Debugging](#debugging)
6. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Build and Run

```bash
# Build the unified launcher
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools

# Or use the launch script
./launch.sh

# Run MNIST module standalone
go run ./cmd/fecim-lattice-tools mnist
```

### Run Tests

```bash
# All tests
go test ./module3-mnist/...

# With race detector
go test ./module3-mnist/... -race

# Specific package
go test ./module3-mnist/pkg/core -v
```

### Download MNIST Data

```bash
cd module3-mnist/data
./download.sh  # Or download manually from http://yann.lecun.com/exdb/mnist/
```

---

## Adding New Features

### Adding a New Control Widget

1. **Create the widget in `pkg/gui/`:**

```go
// my_widget.go
type MyWidget struct {
    widget.BaseWidget
    value float64
    onChange func(float64)
}

func NewMyWidget() *MyWidget {
    w := &MyWidget{}
    w.ExtendBaseWidget(w)
    return w
}

func (w *MyWidget) CreateRenderer() fyne.WidgetRenderer {
    // Create UI elements
    return widget.NewSimpleRenderer(container)
}
```

2. **Add to DualModeApp:**

```go
// In dualmode.go
type DualModeApp struct {
    // ... existing fields
    myWidget *MyWidget
}
```

3. **Initialize in createXxxZone():**

```go
func (app *DualModeApp) createControlsZone() fyne.CanvasObject {
    app.myWidget = NewMyWidget()
    app.myWidget.onChange = func(v float64) {
        // Handle changes
    }
    // Add to layout
}
```

### Adding a New Inference Feature

1. **Add to InferenceResult:**

```go
// In pkg/core/network.go
type InferenceResult struct {
    // ... existing fields
    MyNewMetric float64
}
```

2. **Calculate in Infer():**

```go
// In pkg/core/network_inference.go
func (net *DualModeNetwork) Infer(input []float64) *InferenceResult {
    result := &InferenceResult{}
    // ... existing code
    result.MyNewMetric = calculateMetric()
    return result
}
```

3. **Display in GUI:**

```go
// In pkg/gui/dualmode_inference.go
func (app *DualModeApp) updateResultDisplays(result *core.InferenceResult, ...) {
    // Add UI update
    app.myLabel.SetText(fmt.Sprintf("Metric: %.2f", result.MyNewMetric))
}
```

### Adding QAT Weight Levels

1. **Train weights at new level:**

```bash
go run ./cmd/fecim-lattice-tools mnist train-ptq -levels 15 -epochs 10
```

2. **Add to available levels:**

```go
// In pkg/core/network.go
var AvailableQATLevels = []int{10, 15, 20, 29, 30, 31}
```

3. **Place weight file in `weights/`:**

```
module3-mnist/weights/mnist_weights_15lvl.json
```

---

## Testing

### Writing Tests

```go
// In pkg/core/my_feature_test.go
func TestMyFeature(t *testing.T) {
    // Use test constants for reproducibility
    rng := rand.New(rand.NewSource(42))

    // Create network
    net := NewDualModeNetwork(784, 128, 10)

    // Test behavior
    result := net.Infer(testInput)
    if result.MyMetric < 0 {
        t.Errorf("Expected positive metric, got %f", result.MyMetric)
    }
}
```

### Testing with Synthetic Data

```go
// Generate synthetic MNIST-like data for testing
images, labels := generateSyntheticMNIST(100, 42)
```

### Concurrency Testing

```go
func TestConcurrentInference(t *testing.T) {
    net := NewDualModeNetwork(16, 8, 4)

    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for j := 0; j < 100; j++ {
                net.Infer(randomInput())
            }
        }()
    }
    wg.Wait()
}
```

Run with race detector:
```bash
go test -race ./module3-mnist/pkg/core -run TestConcurrent
```

---

## Common Patterns

### Thread-Safe UI Updates

Always use `fyne.Do()` for UI updates from goroutines:

```go
go func() {
    result := slowOperation()
    fyne.Do(func() {
        app.label.SetText(result) // Safe
    })
}()
```

### Using the Debug Logger

```go
import "fecim-lattice-tools/module3-mnist/pkg/gui"

// In gui package, use mnistLog
mnistLog.Printf("Loading weights from %s", filename)

// For production, log is disabled unless FECIM_DEBUG=1
```

### Reproducible Training

Set environment variable for fixed RNG seed:

```bash
FECIM_DEBUG_SEED=42 ./fecim-lattice-tools
```

### Weight Quantization

```go
// Quantize a single value to 30 levels
quantized := core.QuantizeTo30Levels(0.333)  // Returns 0.344828

// Quantize entire weight matrix
net.RequantizeWeights()  // Uses current NumLevels setting
```

---

## Debugging

### Enable Debug Logging

```bash
FECIM_DEBUG=1 ./fecim-lattice-tools
```

### Inspect Weights

```go
// Get current weights
fpW1, fpW2, fpB1, fpB2 := net.GetFPWeights()
quantW1, quantW2, quantB1, quantB2 := net.GetQuantWeights()

// Check quantization levels used
distinctLevels := make(map[float64]bool)
for _, row := range quantW1 {
    for _, w := range row {
        distinctLevels[w] = true
    }
}
fmt.Printf("Using %d distinct levels\n", len(distinctLevels))
```

### Profile Inference

```go
import "time"

start := time.Now()
result := net.Infer(input)
elapsed := time.Since(start)
fmt.Printf("Inference took %v\n", elapsed)
```

### Check for Race Conditions

```bash
go test -race ./module3-mnist/...
```

---

## Troubleshooting

### "MNIST data not found"

Download MNIST data:
```bash
cd module3-mnist/data
wget http://yann.lecun.com/exdb/mnist/train-images-idx3-ubyte.gz
wget http://yann.lecun.com/exdb/mnist/train-labels-idx1-ubyte.gz
wget http://yann.lecun.com/exdb/mnist/t10k-images-idx3-ubyte.gz
wget http://yann.lecun.com/exdb/mnist/t10k-labels-idx1-ubyte.gz
```

### "Weight file not found"

Check weight file path:
```bash
ls module3-mnist/weights/
```

Expected files: `mnist_weights_30lvl.json`, etc.

### GUI Not Responding

1. Check for blocking operations on main thread
2. Ensure `fyne.Do()` is used for UI updates
3. Look for deadlocks in mutex usage

### Tests Fail with Race

1. Ensure RNG uses `rand.New(rand.NewSource(...))` not global `rand`
2. Check all shared state has mutex protection
3. Verify `fyne.Do()` wraps UI operations

### Quantization Produces Wrong Values

1. Verify NumLevels setting (should be 2-30)
2. Check weight range is [0, 1]
3. Call `RequantizeWeights()` after changing levels

---

## References

- [Architecture Overview](mnist.architecture.md)
- [Fixes TODO](mnist.fixes.todo.md)
- [FeCIM Honesty Audit](../4-research/honesty-audit.md)
- [Fyne Documentation](https://developer.fyne.io/)

---

*Last updated: 2026-01-27*
