# Module 3: MNIST - Features

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## What This Module Does

- Runs dual-path inference: full precision vs CIM.
- Visualizes activations and confidence differences in the dual-mode UI.
- Provides model-based energy visualization alongside modeled accuracy.

## Code Layout (Synced)

### Core inference path (runtime)
- `module3-mnist/pkg/core/` (dual-mode network, quantization, inference, energy model)
- `module3-mnist/pkg/gui/` (Fyne apps/widgets)
- `module3-mnist/pkg/mnist/` (MNIST loader)
- `module3-mnist/cmd/mnist-gui/main.go` (GUI entry)
- `module3-mnist/cmd/mnist/main.go` (CLI entry)

### Training/offline path
- `module3-mnist/pkg/training/` (training network utilities)
- `module3-mnist/cmd/train-network/main.go`
- `module3-mnist/cmd/train-ptq/main.go`
- `module3-mnist/cmd/train-single-layer/main.go`
- Legacy top-level scripts/utilities in `module3-mnist/*.go` are training helpers, not GUI runtime dependencies.

## Key Workflows

- Load pretrained weights and run inference.
- Adjust levels and noise from the UI (noise slider: `0.00-0.20`, clamped in core to match UI/docs).
- Use drawing canvas for live digit input.

## Extension Points

- Add new weight files (QAT levels) and quantization schemes.
- Extend visualization for more layers or metrics.
- Integrate other datasets for comparison.

## Known Limitations

- Training is offline; GUI focuses on inference.
- Small network is chosen for interactivity, not SOTA accuracy.
- Hardware non-idealities are simplified and modeled.
