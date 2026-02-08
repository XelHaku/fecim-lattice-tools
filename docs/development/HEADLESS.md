# Headless / Non-GUI Usage

This repo is primarily a GUI application (Go + Fyne), but several parts can run **without a display server**.

## 1) Headless diagnostics via the main binary

The top-level app supports a small set of “headless modes” intended for debugging and CI.

### Hysteresis headless mode

```bash
# Build once
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Run headless hysteresis diagnostics and exit
./fecim-lattice-tools --mode hysteresis

# Select engine (Preisach is default)
./fecim-lattice-tools --mode hysteresis --engine lk
```

#### Environment variables

These are read by `cmd/fecim-lattice-tools/mode.go`.

- `FECIM_MATERIAL` (optional): selects a material preset.
  - Examples: `fecim_hzo`, `literature_superlattice`, `default_hzo`, `cryogenic_hzo`, `hzo32`, `ftj140`, `alscn`
- `FECIM_RANGE_FRAC` (optional): scales the effective polarization range (0 < frac ≤ 1).

Example:

```bash
FECIM_MATERIAL=literature_superlattice FECIM_RANGE_FRAC=0.7 \
  ./fecim-lattice-tools --mode hysteresis --engine preisach
```

### Output

When CSV logging is available, the headless runner will emit a path like:

- `Headless CSV logging enabled: <path>`

(The logger is created by `module1-hysteresis/pkg/gui/hysteresis_logger.go`.)

## 2) Hysteresis subcommand (GUI/TUI/headless)

The `hysteresis` subcommand has explicit non-GUI modes:

```bash
# Headless ASCII output
./fecim-lattice-tools hysteresis --headless --material superlattice

# Terminal UI (useful over SSH)
./fecim-lattice-tools hysteresis --tui --material superlattice
```

Run `./fecim-lattice-tools hysteresis --help` to see all flags.

## 3) Running GUI tests on headless Linux (Xvfb)

Most unit tests are written to avoid needing a display.

A small number of GUI/layout audit tests intentionally require a display server and will be skipped when `DISPLAY`/`WAYLAND_DISPLAY` are not set.

On Linux CI or servers, you can provide a virtual display:

```bash
sudo apt-get install -y xvfb
xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run LayoutAudit
```

See also: `docs/development/TESTING.md`.
