# What You Can Trust

This page describes trust boundaries for FeCIM Lattice Tools. It should be updated whenever validation changes.

## Trust Levels

| Level | Meaning |
|---|---|
| Highly validated | Reproducible command, artifact, threshold, and passing validation report exist. |
| Literature-backed | Compared to a published source with documented provenance and limitations. |
| Educational | Useful for learning, but not a measurement or production claim. |
| Planned | Intended validation exists in roadmap form only. |
| Not validated | Use only for exploration; do not cite as evidence. |

## Current High-Confidence Claims

| Claim | Status | Evidence |
|---|---|---|
| Module 2 KCL conservation can be checked on deterministic random arrays. | Highly validated | `go test -v ./validation/module2/...` emits `output/validation/module2/kcl_conservation.json`. |
| The repository distinguishes simulation claims from device-measurement claims. | Documentation-backed | `README.md`, `validation/README.md`, `docs/4-research/honesty-audit.md`. |
| The project has an MIT license. | Repository-backed | `LICENSE`. |

## Literature-Backed Areas

| Area | Status | Boundary |
|---|---|---|
| HZO/HfO2 material context | Literature-backed | Parameters and loop comparisons depend on source-specific assumptions and digitized data quality. [claim: hzo-remanent-polarization-range] |
| Park 2015 P-E loop comparison | In progress | Existing validation infrastructure exists; paper-ready plot and final metric still need promotion. |
| Materlik 2015 coefficient context | Literature-backed | Use exact source-specific presets for quantitative comparisons. [claim: materlik-lgd-coefficients] |

## Educational-Only Areas

| Area | Boundary |
|---|---|
| 30-level conductance assumption | Educational/configurable unless tied to a specific peer-reviewed device result. |
| MNIST pipeline demos | Educational unless training, inference, seed, and confusion-matrix artifacts are committed. |
| EDA export flow | Workflow exploration, not tape-out signoff. |
| Peripheral models | Educational abstractions unless compared to SPICE, datasheets, or measured circuits. |

## Not Validated for Production Use

- Long-term retention prediction.
- Yield projection.
- Production timing closure.
- Tape-out readiness.
- Automotive or harsh-environment qualification.
- New silicon performance claims.

## How to Upgrade Trust

To move a claim up a level, add:

1. Command.
2. Artifact.
3. Threshold.
4. Citation or source.
5. Limitation statement.
6. Regression test.
