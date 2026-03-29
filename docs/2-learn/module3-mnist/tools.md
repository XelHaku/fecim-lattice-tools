# Module 3: MNIST - Open-Source Tools

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Used In This Module (Open-Source Dependencies)

- Go toolchain (training/inference utilities in Go)
- Fyne (GUI)

## Integration Notes

- Weight loading lives in `module3-mnist/pkg/training/network.go`.
- Quantization and inference live in `shared/neural/network_inference.go`.
- External ML frameworks are not integrated in this repo.
