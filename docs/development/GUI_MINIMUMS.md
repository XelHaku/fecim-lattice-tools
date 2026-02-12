# GUI Minimum Supported Size

Status: Active baseline for G13.

- **Minimum supported window size:** **1024 x 768**
- Default startup size remains larger for comfort, but persisted/saved sizes are clamped to this minimum.

## Scope

This baseline applies to the main Fyne desktop app (`cmd/fecim-lattice-tools`).

## Evidence

- `cmd/fecim-lattice-tools/main.go` clamps loaded window dimensions to:
  - `minWindowWidth = 1024`
  - `minWindowHeight = 768`
- Visual and layout tests already include 1024x768 scenarios (`cmd/fecim-lattice-tools/e2e_visual_test.go`).
