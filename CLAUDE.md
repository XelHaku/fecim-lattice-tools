# CLAUDE.md - FeCIM Lattice Tools

## For AI Agents

**Full reference:** See `docs/3-develop/api-reference.md` for detailed lookups.

| I need to... | Look in |
|--------------|---------|
| Find a function | `docs/3-develop/api-reference.md` |
| Fix an error | `docs/3-develop/testing/TESTING.md` |
| Add a feature | `docs/3-develop/api-reference.md` |
| Check thread safety | `docs/3-develop/gui/FYNE_NOTES.md#threading-critical` |
| Fix Fyne GUI issues | `docs/3-develop/gui/FYNE_NOTES.md` |
| Run/understand tests | `docs/3-develop/testing/TESTING.md` |
| EDA documentation | `docs/2-learn/module6-eda/README.md` |

## Overview

Go-based lattice tool suite for Ferroelectric Compute-in-Memory (FeCIM) visualization and simulation. It includes configurable material presets, crossbar models, and an educational EDA pipeline.

**Status**: Education phase (simulation-only). See `status.md`.

**Core concept**: The simulator quantizes conductance to a default of 30 discrete levels (configurable). This is a **simulation baseline**, not a validated hardware claim.

> Historical note: some conference material informed early exploration, but the simulator is documented against literature-backed sources and clearly labeled educational where evidence is limited.

## Build & Run

```bash
CGO_ENABLED=0 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
# Or: ./launch.sh
```

## Key Rules

### TDD Hard Rule
- No production-code behavior change without a failing automated test first.
- Applies to features, bug fixes, refactors, public API changes, physics model changes, GUI workflow changes, and validation logic changes.
- Required cycle: RED (write/update focused failing test and confirm the expected failure), GREEN (minimal code to pass), REFACTOR (cleanup while tests stay green).
- Every PR, commit summary, and agent handoff must include the RED command/output summary, GREEN command/output summary, and final verification command(s).
- If a change cannot reasonably be test-first, stop and explain the blocker before coding. Documentation-only, comments-only, formatting-only, generated files, and release metadata may use `TDD: N/A` with a short reason.

### Do
- Keep the default app on `gogpu/ui`; Fyne belongs only in the legacy `cmd/fecim-lattice-tools-fyne` path
- Use `fyne.Do(func() { ... })` for legacy Fyne UI updates from goroutines
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Follow embedded app interface: `BuildContent()`, `Start()`, `Stop()`
- Run `go test ./...` before committing

### Don't
- Add demos without implementing embedded interface
- Use blocking operations on main UI thread
- Commit binaries

## Project Structure

```
cmd/fecim-lattice-tools/     # Default gogpu/ui app entry point
cmd/fecim-lattice-tools-fyne/ # Legacy Fyne app entry point
module1-hysteresis/       # P-E curve, Preisach model
module2-crossbar/         # Crossbar GUI (pkg/gui/)
module3-mnist/            # Neural network digit recognition
module4-circuits/         # DAC/ADC/TIA peripherals
module5-comparison/       # Technology comparison
module6-eda/              # EDA tools
shared/                   # Theme, widgets, logging, physics, crossbar core
  crossbar/               # MVM, non-idealities (IR drop, sneak paths, drift)
  physics/                # L-K solver, Preisach engine, ISPP write controller
```

## Model Defaults (Simulation Parameters)

The project includes **preset parameters** for education and visualization. Treat these as **simulation defaults**, not validated device measurements.

- Material presets: `module1-hysteresis/pkg/ferroelectric/material.go`
- Crossbar defaults: `shared/crossbar/array.go`
- EDA defaults: `module6-eda/pkg/config/types.go`

## Accuracy & Honesty Policy

Scientific accuracy over marketing claims. Full audit: `docs/4-research/honesty-audit.md`.

### Verified External Claim (Current Audit)

- **98.24% MNIST accuracy** in HZO FTJ reservoir computing (Journal of Alloys and Compounds 2025). This is **not** a FeCIM device claim and should not be attributed to this simulator.

### Unverified/Removed Claims (Do Not Present as Facts)

- 30 analog states for the educational simulator baseline (conference-only reference)
- 87% MNIST accuracy (conference-only reference)
- Energy multipliers vs NAND or GPUs without reported in literature measurement evidence

## Testing

```bash
go test ./...                            # See CI for latest status
go test ./shared/crossbar/...            # Crossbar only
```

Full test documentation: `docs/3-develop/testing/TESTING.md`

## Cognee (Knowledge Engine)

This repo has a local Cognee instance for persistent AI memory — one DB per repo, no Docker.

**Setup:** `./scripts/cognee-setup.sh` then edit `.env` with your `LLM_API_KEY`.

**Config:** Uses OpenRouter (`openai/gpt-4o-mini`) for LLM and `fastembed` for local embeddings. Env vars must be set BEFORE importing cognee (lru_cache). See `.env.example`.

**Usage (Python API — preferred for scripts):**
```python
import os
os.environ["COGNEE_SKIP_CONNECTION_TEST"] = "true"
os.environ["ENABLE_BACKEND_ACCESS_CONTROL"] = "false"
# ... other env vars from .env.example
import cognee, asyncio

async def main():
    await cognee.add("your text or /absolute/path/to/file.md")
    await cognee.cognify()
    results = await cognee.search("your query")
asyncio.run(main())
```

**Bulk ingest:** `python3 scripts/cognee-ingest.py` (loads key docs into the KG).

**When to use:**
- Store research findings, audit results, physics validation notes
- Build searchable memory across sessions for FeCIM domain knowledge
- Index docs or papers for retrieval during development

**Known gotchas:**
- Gemini Flash does NOT work with cognee (bad structured output). Use gpt-4o-mini via OpenRouter.
- Env vars must be set before `import cognee` due to lru_cache.
- `cognee-cli config set` persists values that override env vars — avoid using it.

**Data location:** `.cognee_system/` (local, gitignored).

## Git Conventions

- Commit: `type: description` (feat, fix, docs, refactor, test, chore)
- Run tests before pushing

## Ignore

- `logs/`, `output/`, generated artifacts
