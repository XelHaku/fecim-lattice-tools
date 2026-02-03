# Example 02: MNIST First Layer

Compile the first layer of a trained MNIST classifier to a FeCIM crossbar.

## Overview

This example compiles a realistic neural-network weight matrix: the first fully connected layer of an MNIST digit classifier (784x128).

**Note:** For practicality, we use a 32x32 subset (first 32 inputs for 32 output neurons).

## Files

| File | Description |
|---|---|
| `weights.json` | 32x32 subset of MNIST layer 1 weights |
| `run.sh` | Compilation script |

## Running the Example

```bash
cd module6-eda

# Compile weights
./examples/02-mnist-layer/run.sh

# Or manually:
go run ./cmd/eda-cli \
  -input examples/02-mnist-layer/weights.json \
  -output examples/02-mnist-layer/output \
  -rows 32 -cols 32 -levels 30
```

## Output Files

Default output names (if `-name` is not provided):

- `fecim_array_design.json`
- `fecim_array_cells.csv`
- `fecim_array.sp`
- `fecim_array.v`
- `fecim_array.def`

## Weight Matrix Origin (Reference)

```python
# Training code (PyTorch)
class MNISTClassifier(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = nn.Linear(784, 128)  # 784x128 weights
        self.fc2 = nn.Linear(128, 10)

model = MNISTClassifier()
# ... train on MNIST ...
weights = model.fc1.weight.detach().numpy()  # Shape: (128, 784)

# Extract 32x32 subset for this example
subset = weights[:32, :32]
```

## Interpreting Results

### Level Distribution

For trained neural-network weights, expect:
- Most weights near zero -> levels 14-15 (middle range)
- Some positive outliers -> levels 20-29
- Some negative outliers -> levels 0-9

### PSNR

Quantization error is measured as PSNR:
- **> 40 dB:** Excellent
- **30-40 dB:** Good
- **< 30 dB:** May need more levels or retraining

### Conductance Mapping (Default)

```
Weight: -0.9 -> Level: 0  -> G: 10.0 uS
Weight:  0.0 -> Level: 15 -> G: 50.5 uS
Weight: +0.9 -> Level: 29 -> G: 100.0 uS
```

## Extending to Full Network

For a complete MNIST classifier on FeCIM:

1. **Layer 1 (784x128):** 6 crossbars of 128x128 + partial
2. **Layer 2 (128x10):** 1 crossbar of 128x16 (padded)

Total: ~101,000 FeCIM cells

## Next Steps

1. Measure inference accuracy with quantized weights.
2. Compare with floating-point baseline.
3. Integrate with Module 3 (MNIST visualization) for end-to-end demo.
