# Example 02: MNIST First Layer

Compile the first layer of a trained MNIST classifier to a FeCIM crossbar.

## Overview

This example compiles an example neural-network weight matrix: the first fully connected layer of an MNIST digit classifier (784x128).

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
go run ../cmd/fecim-lattice-tools eda cli \
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

Weight distributions are dataset- and training-dependent. With symmetric quantization, many weights often cluster near mid-levels, with fewer values at the extremes.

### PSNR

PSNR is reported as a **rough indicator** of quantization error. It is not a direct measure of task accuracy.

### Conductance Mapping (Default)

```
Weight: -0.9 -> Level: 0  -> G: 10.0 uS
Weight:  0.0 -> Level: 15 -> G: 50.5 uS
Weight: +0.9 -> Level: 29 -> G: 100.0 uS
```

Values above assume the CLI defaults for `gmin`/`gmax`. Change those flags to alter the mapping.

## Extending to Full Network

For a complete MNIST classifier on FeCIM:

1. **Layer 1 (784x128):** 6 crossbars of 128x128 + partial
2. **Layer 2 (128x10):** 1 crossbar of 128x16 (padded)

Total: ~101,000 FeCIM cells

## Next Steps

1. Measure inference accuracy with quantized weights in your own pipeline.
2. Compare with a floating-point baseline.
3. Integrate with Module 3 (MNIST visualization) for an end-to-end demo.
