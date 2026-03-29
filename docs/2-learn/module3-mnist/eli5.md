# Module 3: MNIST - ELI5

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Learning Objectives

- Understand why MNIST is a useful demo for hardware-aware inference.
- See how full-precision and CIM paths differ.
- Know which page to read next for formal detail.

## Intuition

The MNIST demo asks a simple question: can the hardware-inspired path
predict digits as reliably as the ideal digital path? We run the same
network twice, once with perfect math and once with quantization and noise.

The gap between the two shows how hardware constraints shape accuracy.

## Key Analogies

- Two chefs following the same recipe, one with perfect measurements and
  one with coarse measuring cups.
- Two calculators: one exact, one with rounded results.

## What The Simulator Simplifies

- The network is small and fixed to be fast and visual.
- Noise and quantization are modeled, not measured from devices.
- Training is done offline; the demo focuses on inference.

## Next Steps

- Read the formal model in [physics.md](physics.md).
- Connect to implementation details in [features.md](features.md).
