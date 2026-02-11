# Project Status

**Last updated**: 2026-02-07

## Phase

**Education** (roadmap: Education → Research → Chip Design)

## Validation

- **Simulation-only**. Experimental validation is pending.
- The “Simulation vs Experiment” tab is a placeholder until real datasets are added.
- EDA outputs are educational artifacts and not signoff/tapeout ready.

## Claims Policy

- **Simulation baselines** (e.g., 30-level baseline) are labeled as unverified.
- **Peer-reviewed claims** are cited explicitly in docs and UI.

## Testing

- CI workflow `ci` runs `go test ./...` on every push/PR.
- Run locally with `go test ./...` (race tests are optional).
- Headless LK ISPP sanity/regression: `go test ./cmd/fecim-lattice-tools -run TestHeadlessLKRun_CompletesISPP`

## Coverage

- Estimated at ~85% (update once CI coverage reporting is added).

## Critique Tracking

- Consolidated critique list (snapshot): `CRITIQUE_MASTER_LIST.md`
- Current roadmap and task updates: `TODO.md`
