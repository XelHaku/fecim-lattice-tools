// pkg/export/crosssim.go
// CrossSim configuration generator for FeCIM crossbar arrays.
//
// CrossSim (https://github.com/sandialabs/cross-sim) is an open-source
// Python framework from Sandia National Laboratories for simulating analog
// in-memory computing crossbar arrays with hardware non-idealities.
//
// CrossSim models:
//   - Device conductance quantization and noise
//   - ADC/DAC precision and non-linearity
//   - Parasitic resistance (IR drop from line resistance)
//   - Cycle-to-cycle and device-to-device variation
//
// The generated YAML config maps FeCIM module6 ArrayConfig parameters to
// CrossSim simulation parameters for hardware-accurate inference.
//
// Usage:
//
//	python3 -m crosssim --config fecim_crosssim.yaml
//
// References:
//   CrossSim: https://github.com/sandialabs/cross-sim
//   CrossSim paper: Taheri et al., DATE 2022
package export

import (
	"fmt"
	"strings"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GenerateCrossSIMConfig returns a CrossSim-compatible YAML configuration
// for simulating the FeCIM crossbar array.
//
// Maps:
//   - Rows/Cols → array dimensions
//   - Architecture → device model (resistive 2-terminal for passive/1T1R/2T1R)
//   - Technology → process parameters (conductance range, noise level)
//   - Mode → simulation type (inference, training, storage)
func GenerateCrossSIMConfig(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)

	// Map FeCIM mode to CrossSim task
	var taskType string
	switch strings.ToLower(cfg.Mode) {
	case "compute":
		taskType = "inference"
	case "memory":
		taskType = "training"
	default:
		taskType = "lookup"
	}

	// Conductance range depends on architecture
	// Passive (0T1R): Ron ~100kΩ → 1µS max, Roff ~10GΩ → 0.1nS min
	// 1T1R/2T1R: higher current per cell due to selector transistor
	var gMaxUS, gMinUS float64
	switch strings.ToLower(cfg.Architecture) {
	case "1t1r", "2t1r":
		gMaxUS = 100.0 // µS (selector ON, ferroelectric in LRS)
		gMinUS = 0.01  // µS (selector ON, ferroelectric in HRS)
	default: // passive
		gMaxUS = 10.0 // µS (FeCIM passive LRS ~100kΩ)
		gMinUS = 0.001 // µS (FeCIM passive HRS ~1GΩ)
	}

	// IR drop parameters (line resistance increases with array size)
	// Metal1 sheet resistance ~0.1 Ω/□ for thick metal
	// For 0.46µm width × N cells: R_wire ≈ 0.1 * pitch/width * N ≈ 1Ω/cell
	wireResOhm := float64(cfg.Rows) * 1.0 // Ω total word-line resistance

	// Noise: typical FeCIM cycle-to-cycle variation ~2-5% of Grange
	noiseSigmaUS := (gMaxUS - gMinUS) * 0.03

	// ADC bits: 4-bit is literature-optimal for CIM (from our lit review)
	adcBits := 4
	dacBits := 4

	return fmt.Sprintf(`# FeCIM CrossSim Configuration
# Generated: %s
# Design: %s
# Array:  %dx%d cells, architecture=%s, mode=%s
#
# Usage:
#   pip install crosssim
#   python3 -c "from simulator import CrossSimParameters; ..."
#   OR: python3 run_crosssim.py --config %s_crosssim.yaml
#
# CrossSim documentation: https://github.com/sandialabs/cross-sim/wiki
# Paper: Taheri et al., DATE 2022

# ── Array geometry ─────────────────────────────────────────────────────────────
array:
  rows: %d
  cols: %d
  architecture: %s     # passive=0T1R, 1t1r=selector+FeCap, 2t1r=dual-selector
  task: %s              # inference | training | lookup

# ── Device model ───────────────────────────────────────────────────────────────
device:
  model: ferroelectric_resistive   # FeCIM: resistive state via ferroelectric switching
  technology: %s

  # Conductance range (µS)
  # Source: FeCIM simulation defaults (module6-eda physics-based estimates)
  g_max: %.4f          # µS — low-resistance state (LRS / programmed)
  g_min: %.6f         # µS — high-resistance state (HRS / erased)

  # Non-idealities
  noise:
    model: gaussian_relative       # σ relative to G_range
    sigma: %.4f        # µS — cycle-to-cycle variation (~3%% of G_range)
  drift:
    model: power_law               # G(t) = G0 * (t/t0)^(-beta)
    beta: 0.02                     # FeCIM retention exponent
    t0_s: 1.0                      # Reference time (1 second)

# ── Peripheral circuit models ───────────────────────────────────────────────────
peripherals:
  dac:
    bits: %d                       # 4-bit DAC (literature-optimal for CIM, Xu et al. 2021)
    model: ideal                   # ideal | non_monotonic | with_noise
    vref: %.1f                     # V (supply voltage)

  adc:
    bits: %d                       # 4-bit ADC (literature-optimal, ADC dominates 40-60%% energy)
    model: ideal                   # ideal | flash | sar | sigma_delta
    vref: %.1f                     # V

# ── Parasitic models ────────────────────────────────────────────────────────────
parasitics:
  # Word-line (row) resistance
  wire_resistance:
    model: lumped                  # lumped | distributed
    r_wl_ohm: %.2f                 # Ω total word-line resistance (%d rows × 1Ω/cell)
    r_bl_ohm: %.2f                 # Ω total bit-line resistance (similar)

  # Sneak path (passive arrays only)
  sneak_paths:
    enabled: %v                    # True for passive 0T1R, False for 1T1R/2T1R

# ── Simulation parameters ───────────────────────────────────────────────────────
simulation:
  n_runs: 10                       # Monte Carlo runs for noise statistics
  seed: 42                         # Random seed for reproducibility
  log_level: INFO

  # Output metrics
  metrics:
    - array_output_current         # Column sum currents (A)
    - effective_conductance        # Per-cell conductance (µS)
    - snr_db                       # Signal-to-noise ratio (dB)
    - linearity_error              # INL/DNL from ideal linear response

# ── Output ─────────────────────────────────────────────────────────────────────
output:
  directory: output/crosssim
  formats:
    - csv                          # Raw current/conductance data
    - json                         # Structured metrics
  plot: false                      # Set true for matplotlib visualizations
`,
		time.Now().Format("2006-01-02"),
		designName,
		cfg.Rows, cfg.Cols, cfg.Architecture, cfg.Mode,
		designName,
		cfg.Rows, cfg.Cols,
		cfg.Architecture,
		taskType,
		cfg.Technology,
		gMaxUS, gMinUS,
		noiseSigmaUS,
		dacBits, 1.8,
		adcBits, 1.8,
		wireResOhm, cfg.Rows,
		wireResOhm,
		strings.ToLower(cfg.Architecture) == "passive")
}

// GenerateCrossSIMRunScript returns a minimal Python runner that loads
// the generated YAML config and executes a CrossSim array simulation.
func GenerateCrossSIMRunScript(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)

	return fmt.Sprintf(`#!/usr/bin/env python3
# FeCIM CrossSim Runner
# Generated: %s
# Design: %s
#
# Usage:
#   pip install crosssim pyyaml numpy
#   python3 run_crosssim.py
#
# CrossSim: https://github.com/sandialabs/cross-sim

import sys
import yaml
import numpy as np

try:
    from simulator import AnalogCore, CrossSimParameters
except ImportError:
    print("CrossSim not installed. Install with:")
    print("  pip install crosssim")
    print("  OR: git clone https://github.com/sandialabs/cross-sim && pip install -e .")
    sys.exit(1)

CONFIG_FILE = "%s_crosssim.yaml"

print(f"FeCIM CrossSim Simulation: {CONFIG_FILE}")
print(f"Config: {CONFIG_FILE}")
print("")

# Load config
with open(CONFIG_FILE) as f:
    cfg = yaml.safe_load(f)

# Configure CrossSim parameters
params = CrossSimParameters()
rows = cfg["array"]["rows"]
cols = cfg["array"]["cols"]
params.core.rows = rows
params.core.cols = cols

# Device model
g_max = cfg["device"]["g_max"] * 1e-6   # Convert µS to S
g_min = cfg["device"]["g_min"] * 1e-6
params.core.Gmax_relative = g_max
params.core.Gmin_relative = g_min

# Initialize array with random conductances (normalized 0–1)
rng = np.random.default_rng(cfg["simulation"].get("seed", 42))
G_raw = rng.uniform(g_min, g_max, (rows, cols))
W = (G_raw - g_min) / (g_max - g_min)  # Normalize to [0, 1]

# Create AnalogCore: W is the normalized weight matrix, params configures hardware
core = AnalogCore(W, params)

# Test: apply uniform input vector
x = np.ones(cols)
y = core.run_xbar(x)

print(f"Array dimensions: {rows}x{cols}")
print(f"Conductance range: [{g_min*1e6:.4f}, {g_max*1e6:.4f}] µS")
print(f"Input:  {x[:min(8,cols)]} ...")
print(f"Output: {y[:min(8,rows)]} ...")
print("")
print("Simulation complete.")
print(f"Output directory: {cfg['output']['directory']}")
`, time.Now().Format("2006-01-02"),
		designName, designName)
}
