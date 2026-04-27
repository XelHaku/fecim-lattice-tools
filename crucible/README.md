# The Crucible

The Crucible is the planned continuous scientific validation system for FeCIM Lattice Tools.

Its purpose is simple: every public claim should be tested, attacked, fixed, and documented before it becomes part of the project story.

## Status

Current status: methodology saved, partial execution started.

Already implemented:

- Module 2 KCL conservation gate in `validation/module2/`
- public validation overview in `validation/README.md`
- paper scaffold in `paper/`

Not yet implemented:

- automated Prover/Disprover/Builder orchestration
- per-module Prover reports
- per-module Disprover reports
- metrics dashboard
- full paper figures from validation artifacts

## Success Tiers

| Tier | Definition |
|---|---|
| Educational success | A student can use the tool and understand FeCIM concepts that are hard to grasp from papers alone. |
| Research success | A researcher can trust scoped outputs enough to inform an experiment, hypothesis, or design discussion. |
| Scientific success | The tool's claims are reproducible, validated, versioned, and able to withstand peer scrutiny. |

## The Three Agents

| Agent | Stance | Output |
|---|---|---|
| Prover | Show the data. | Quantitative pass/fail reports with artifacts. |
| Disprover | How is this wrong? | Counterexamples, edge cases, overclaims, and failure reports. |
| Builder | Fix the highest-risk issue. | Code, tests, docs, and changelog updates. |

## Loop

```text
Prover runs validation
  -> Disprover attacks assumptions
  -> Builder fixes highest-priority issues
  -> Prover reruns
  -> repeat
```

## Public Rule

Do not describe a module as validated unless the claim links to:

- a command
- an artifact
- a threshold
- a source or citation
- a limitation statement

## Directory Layout

```text
crucible/
  README.md
  METHODOLOGY.md
  prover/
    README.md
    reports/
  disprover/
    README.md
    reports/
  builder/
    README.md
    changelog.md
```

Generated validation artifacts belong under `output/validation/` or `validation/output/` unless promoted to curated fixtures under `validation/testdata/`.

