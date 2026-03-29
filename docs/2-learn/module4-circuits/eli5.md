# Module 4: Circuits - ELI5

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Learning Objectives

- Understand the role of DACs, ADCs, and TIAs in a CIM system.
- See how peripheral circuits connect to the array model.
- Know which page to read next for formal detail.

## Intuition

The crossbar array speaks analog, but the rest of the system is mostly digital.
DACs translate digital inputs into voltages, TIAs convert output currents to voltages,
and ADCs turn those voltages back into digital numbers.

These blocks are the translators between math and hardware signals.

## Key Analogies

- A language interpreter translating between two speakers.
- A camera sensor chain: light to voltage to digital pixels.
- A sound system: digital audio to analog waves and back.

## What The Simulator Simplifies

- Circuit behaviors are modeled with idealized formulas.
- Noise and nonlinearity are simplified.
- Timing and power are estimates, not SPICE-calibrated.

## Next Steps

- Read the formal model in [physics.md](physics.md).
- Connect to implementation details in [features.md](features.md).
