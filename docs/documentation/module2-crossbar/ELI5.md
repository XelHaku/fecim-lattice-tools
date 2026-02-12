# Module 2: Crossbar - ELI5

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Learning Objectives

- Build intuition for crossbar arrays and hardware MVM.
- Understand what the simulator models versus what it simplifies.
- Know which page to read next for formal detail.

## Intuition

Imagine a grid of tiny adjustable resistors.
You apply voltages on the rows and measure currents on the columns.
Because each column sums currents, the grid performs matrix-vector multiply in parallel.

## Key Analogies

- A city grid of water pipes: each intersection controls flow.
- A mixing board: each slider scales a signal and outputs a sum.

## What The Simulator Simplifies

- Conductance is treated as linear in the core model.
- Device states are quantized to a fixed number of levels.
- Wire resistance and sneak paths use simplified parameters.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).
