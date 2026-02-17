# EDA Trust Boundary (Module 6)

## Explicit trust boundary

**Module 6 EDA exports are educational/research prototypes.**

They are useful for architecture exploration, teaching, and early design-space studies. They are **NOT suitable for signoff, tapeout, or production characterization**.

## Why this is not production signoff

For production-quality EDA collateral, the following are still missing:

- **Measured NLDM timing tables** from silicon-calibrated characterization flows
- **Multi-corner characterization** (PVT corners, variation distributions) validated against silicon data
- **DRC/LVS closure against a real foundry PDK**, decks, and signoff rule interpretations
- **CCS timing/noise models** and corresponding signoff-quality library views

## What Module 6 does provide today

- **Analytical models** with literature-calibrated parameters for FeCIM-oriented exploration
- **Structural correctness** of generated artifacts (valid syntax/formats for supported exporters)
- **Research-grade approximations** for early feasibility, relative comparisons, and workflow prototyping

## Recommended path for users who need production quality

If your target is tapeout or product qualification:

1. Treat Module 6 output as a **starting scaffold only**.
2. Re-characterize timing/power with foundry-qualified extraction and simulation flows.
3. Replace placeholder/analytical timing with **measured NLDM/CCS libraries**.
4. Run complete **signoff DRC/LVS/STA/IR/EM** using real PDK decks and corners.
5. Validate against silicon (or trusted shuttle data) before any production decision.

## Intended use

Module 6 is intentionally positioned for:

- education and training,
- algorithm/architecture co-exploration,
- rapid prototyping of EDA integration paths.

Do not represent Module 6 artifacts as production signoff data.
