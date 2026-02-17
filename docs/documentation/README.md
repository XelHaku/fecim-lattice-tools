# FeCIM Lattice Tools Curriculum

This curriculum is a concise, research-grade path through the physics, math, and software
that power FeCIM Lattice Tools. It teaches intuition first, then formal models, then
implementation details, with clear separation between demonstrated results and modeled
assumptions.

## Learning Path (Recommended Order)

1. Module 1: Hysteresis - build intuition for ferroelectric memory physics
2. Module 2: Crossbar - learn how arrays compute MVM in hardware
3. Module 3: MNIST - compare full-precision vs CIM inference
4. Module 4: Circuits - understand DAC/ADC/TIA support blocks
5. Module 5: Comparison - interpret system-level tradeoffs honestly
6. Module 6: EDA - compile networks into crossbar mappings
7. Module 7: Docs - navigate, search, and curate the knowledge base

## Why This Sequence

- Device physics sets the limits, so we start with hysteresis.
- Crossbar arrays translate device behavior into computation.
- MNIST shows how hardware constraints shape algorithm behavior.
- Circuits provide the peripheral infrastructure the array needs.
- Comparison and EDA move from device-level understanding to system design.
- The Docs module ties everything together and accelerates learning.

## Prerequisites

- Basic algebra and unit awareness
- Comfort reading plots and simple functions
- Optional: linear algebra (vectors, matrices)

## Optional Deep Dives

- Device physics: `docs/research-papers/by-topic/01-ferroelectric-materials/`
- Crossbar modeling: `docs/research-papers/by-topic/04-cim-architectures/`
- Compilers and mapping: `docs/research-papers/by-topic/10-cim-compilers-mapping/`

## Fast Path (Already Comfortable With Physics)

- Skip ELI5 pages and go straight to PHYSICS and FEATURES.
- Use `MODULES.md` as the index to jump by topic.

## Lab vs Literature (Honesty First)

- Demonstrated: use the Sources sections to trace claims to internal docs.
- Modeled: simulator parameters are modeling choices unless explicitly cited.
- Aspirational: architectural comparisons are directional, not guarantees.
- Source of truth for external claim verification: `docs/4-research/honesty-audit.md`.
- Cross-module physics acceptance criteria: `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`.

## How To Use This Curriculum

- Start with the module index: `MODULES.md`.
- Each module has four pages: ELI5, PHYSICS, FEATURES, OPENSOURCE-TOOLS.
- If a term is unfamiliar, check `docs/GLOSSARY.md`.
