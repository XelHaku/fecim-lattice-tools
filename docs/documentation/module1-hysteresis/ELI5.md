# Module 1: Hysteresis - ELI5

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Learning Objectives

- Build intuition for ferroelectric memory and P-E hysteresis loops.
- Understand what the simulator models versus what it simplifies.
- Know which page to read next for formal detail.

## Intuition

A ferroelectric cell is like a tiny switch that remembers its last strong push.
When you push the electric field one way, the polarization settles into that state.
If you reverse the field, it resists until you push hard enough, creating a loop.

This loop is the signature of memory: the output depends on history, not just the
current input.

## Key Analogies

- A ball in a double-well: it stays in one valley until you push over the hill.
- A sticky light switch: small nudges do not flip it.
- A rubber band with lag: it responds, but with memory.

## What The Simulator Simplifies

- The material is treated as uniform, without grain boundaries or defects.
- Switching is modeled as many idealized hysterons (Preisach elements).
- Dynamic domain motion is simplified into quasi-static loops.

## Next Steps

- Read the formal model in [PHYSICS.md](PHYSICS.md).
- Connect to implementation details in [FEATURES.md](FEATURES.md).
