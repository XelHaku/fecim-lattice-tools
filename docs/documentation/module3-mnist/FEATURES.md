# Module 3: MNIST - Features

## What This Module Does

- Runs dual-path inference: full precision vs CIM.
- Visualizes activations, confusion matrix, and confidence differences.
- Provides energy-model visualization alongside accuracy.

## Primary Components

- `module3-mnist/pkg/core/network.go`
- `module3-mnist/pkg/core/network_quantization.go`
- `module3-mnist/pkg/gui/dualmode.go`
- `module3-mnist/pkg/gui/energy_widget.go`

## Key Workflows

- Load pretrained weights and run inference.
- Adjust levels, noise, ADC/DAC bits and compare outputs.
- Use drawing canvas for live digit input.

## Extension Points

- Add new weight files (QAT levels) and quantization schemes.
- Extend visualization for more layers or metrics.
- Integrate other datasets for comparison.

## Known Limitations

- Training is offline; GUI focuses on inference.
- Small network is chosen for interactivity, not SOTA accuracy.
- Hardware non-idealities are simplified.
