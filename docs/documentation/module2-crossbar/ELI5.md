# Module 2: Crossbar - ELI5

## Learning Objectives

- Build intuition for array physics, mvm, and non-idealities.
- Understand what the simulator is modeling versus simplifying.
- Know which page to read next.

## Intuition

Imagine a grid of tiny adjustable resistors.
You apply voltages on one side and currents add up on the other.
That is matrix-vector multiplication happening in parallel, in hardware.

## Key Analogies

- A city grid of water pipes: each intersection sets how much water flows.
- A mixing board: each slider scales a signal, and outputs are sums.

## What the Simulator Simplifies

- Ohm's law is treated as linear in the core model.
- Conductance is quantized to 30 levels (demo baseline; conference claim).
- IR drop and sneak paths are modeled with simplified wire parameters.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).
