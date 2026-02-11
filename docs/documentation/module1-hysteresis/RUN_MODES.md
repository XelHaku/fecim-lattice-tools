# Module 1 Run Modes and Physics Defaults

This document is the canonical reference for how `module1-hysteresis` runs in:

- GUI
- TUI
- Headless
- Vulkan

It also clarifies **Preisach vs Landau-Khalatnikov (L-K)** defaults by entry path.

---

## 1) Entry Points

Module 1 can be started from two CLI paths:

1. `fecim-lattice-tools hysteresis ...` (subcommand)
   - Source: `module1-hysteresis/cmd/hysteresis/main.go`
2. `fecim-lattice-tools --mode hysteresis [--engine preisach|lk]` (headless diagnostics path)
   - Source: `cmd/fecim-lattice-tools/mode.go`

These paths are related but not identical.

---

## 2) `hysteresis` Subcommand Run Modes (GUI/TUI/headless/Vulkan)

Command:

```bash
fecim-lattice-tools hysteresis [--headless] [--tui] [--vulkan] [--material ...] [--freq ...]
```

### Default mode (no mode flags)
- Runs **Fyne GUI** (`gui.RunWithMaterial(...)`).
- If GUI fails, fallback chain is:
  1. TUI (`tui.RunWithMaterial(...)`)
  2. Headless (`runHeadless(...)`)

### `--tui`
- Runs terminal UI (`pkg/tui`).
- Physics model in TUI is **Preisach** (`ferroelectric.NewPreisachModel(...)`).
- If TUI fails, falls back to headless.

### `--headless`
- Runs static ASCII renderer.
- Physics model is **Preisach** (`ferroelectric.NewPreisachModel(...)`).

### `--vulkan`
- Runs Vulkan renderer (`pkg/render` + simulation engine loop).
- If Vulkan init fails, falls back to headless.
- This path does **not** expose a CLI engine selector for L-K vs Preisach.

### Mode-flag precedence (important)
If multiple mode flags are passed, `main.go` checks in this order:

1. `--headless`
2. `--tui`
3. `--vulkan`
4. default GUI

So `--headless` wins over `--tui`/`--vulkan` when combined.

---

## 3) Physics Defaults (Preisach vs L-K)

## A. `fecim-lattice-tools hysteresis ...` (subcommand)

- **Default in GUI app state:** `PhysicsPreisach`
  - Set in `pkg/gui/gui.go` during app initialization.
- GUI can be changed at runtime via the **Physics Engine** dropdown:
  - `L-K (dynamic)`
  - `Preisach (quasi-static)`
- TUI/headless subcommand paths currently run Preisach directly.

Summary: for the subcommand path, **Preisach is the effective default**.

## B. `fecim-lattice-tools --mode hysteresis ...` (headless diagnostics)

- Engine normalization accepts aliases (`lk`, `l-k`, `landau`, `preisach`, etc.).
- If engine is omitted, default is **Preisach**.
- L-K is opt-in via `--engine lk`.

Summary: for diagnostics mode, **Preisach is also the default**.

---

## 4) Material Defaults (to avoid confusion)

Defaults differ by entry path:

- `hysteresis` subcommand default material flag: `superlattice`
  - `--material` default in `module1-hysteresis/cmd/hysteresis/main.go`
- `--mode hysteresis` headless diagnostics default material (when `FECIM_MATERIAL` unset): `FeCIMMaterial`
  - in `cmd/fecim-lattice-tools/mode.go`

This is intentional in current code, but easy to misread in docs.

---

## 5) Quick Command Reference

```bash
# GUI (default)
fecim-lattice-tools hysteresis

# TUI
fecim-lattice-tools hysteresis --tui

# Headless ASCII (Preisach)
fecim-lattice-tools hysteresis --headless

# Vulkan render path
fecim-lattice-tools hysteresis --vulkan

# Headless diagnostics (default Preisach)
fecim-lattice-tools --mode hysteresis

# Headless diagnostics with L-K
fecim-lattice-tools --mode hysteresis --engine lk
```

---

## 6) Source of Truth

- Run-mode selection and fallback: `module1-hysteresis/cmd/hysteresis/main.go`
- GUI default physics engine: `module1-hysteresis/pkg/gui/gui.go`
- TUI physics model: `module1-hysteresis/pkg/tui/tui.go`
- Headless diagnostics engine default/normalization: `cmd/fecim-lattice-tools/mode.go`
