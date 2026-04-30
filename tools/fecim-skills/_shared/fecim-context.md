# FeCIM Domain Context (Skill Reference)

Shared context referenced by FeCIM skills. Do not duplicate this content in individual SKILL.md files.

## Published-physics citation list (canonical short form)

| Short form | Full citation | Used for |
|---|---|---|
| Materlik 2015 | Materlik et al., J. Appl. Phys. 117, 134109 (2015) | HZO Landau parameters |
| Park 2015 | Park et al., Adv. Mater. 27, 1811 (2015) | HZO ferroelectricity |
| Alessandri 2018 | Alessandri et al., IEEE TED 65, 4503 (2018) | Polarization switching dynamics |
| Guo 2018 | Guo et al., Nano Lett. 18, 1727 (2018) | Crossbar device characterization |
| HZO FTJ 2025 | J. Alloys Compd. (2025) | 98.24% MNIST reservoir computing — NOT a FeCIM device claim |

## Honesty-audit policy (`docs/4-research/honesty-audit.md`)

**Verified claims (peer-reviewed):** 98.24% MNIST in HZO FTJ reservoir computing — must be attributed to HZO FTJ 2025 paper, NOT this simulator.

**Simulation defaults (educational, must be labeled as such):**
- 30 analog conductance levels (configurable simulation default)
- Material parameters (HZO, BTO, PZT) from published physics

**Removed/unverified — flag as policy violation:**
- "30 analog states" presented as device fact (conference-only baseline)
- "87% MNIST accuracy" attributed to FeCIM simulator (conference-only)
- Energy multipliers vs NAND or GPUs without published measurement evidence

## 5 known bug patterns (`MEMORY.md`)

1. **Guard-band sign direction flip** — limit guard pulses to 2 max, clamp `calcLevel` to prevent overshoot.
2. **Bounds collapse `[VMin, VMax]`** — widen minimally using direction info, don't reset to full range.
3. **ACCEPT ±1 guard interaction** — skip ACCEPT ±1 when `guardActive=true`, raise overshoot threshold from 3 to 8.
4. **Zero-field bounds reset** — when `absField < 0.01*Ec`, reset to full `[0, MaxField]`.
5. **Preisach Everett zero-clamp** — use product-form (always non-negative) instead of factorized (goes negative).

## UI thread-safety rule (`docs/3-develop/gui/FYNE_NOTES.md`)

All UI updates from goroutines MUST use `fyne.Do(func() { ... })`. Direct widget mutation from a non-main goroutine causes hangs and freezes. Common patterns to audit:

```go
// WRONG
go func() {
    result := compute()
    label.SetText(result)   // RACE
}()

// RIGHT
go func() {
    result := compute()
    fyne.Do(func() { label.SetText(result) })
}()
```

## UI boundary rule (per `AGENTS.md`)

New UI-neutral, physics, simulation, validation, and export work must NOT add `fyne.io/...` or `github.com/gogpu/ui` imports. Use `shared/viewmodel/` as the UI-neutral bridge. Fyne and gogpu/ui imports belong only in shell/UI packages.

## Build target matrix

| Target | CGO | Entry | Use |
|---|---|---|---|
| Legacy | `CGO_ENABLED=1` (default) | `./cmd/fecim-lattice-tools` | Current Fyne GUI |
| Next | `CGO_ENABLED=0` | `./cmd/fecim-lattice-tools-next` | Future zero-CGO gogpu/ui shell |

## Host preflight and search fallback

Before using any FeCIM skill, verify the active repo and tools:
```bash
pwd
test -d /home/xel/git/sages-openclaw/workspace-riju/fecim-lattice-tools && echo "primary repo present" || echo "primary repo MISSING"
command -v go && command -v git
command -v rg >/dev/null 2>&1 || echo "rg missing; use grep or file-search fallback"
git status --short --branch
```
Do not install host packages from a skill. Missing compilers, headers, qmd, agent-browser, or rg are blockers/fallbacks to report with exact evidence.

## Test invocations

| Command | Use | Evidence to report |
|---|---|---|
| `go test ./...` | Full suite | exit code plus package/test pass/fail/skip counts |
| `go test -json ./...` | Full-suite count extraction | summarized pass/fail/skip counts |
| `go test -race ./...` | Race detection (mandatory when changing concurrency) | exit code plus package/test counts |
| `make test-next-ui` | Future zero-CGO UI shell tests | exit code and target output |
| `FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...` | Regenerate physics regression goldens (only when divergence is intentional) | written justification and golden diff |
