#!/usr/bin/env python3
# FeCIM PySpice Crossbar Simulation
# Generated: 2026-02-20
# Design: fecim_crossbar_2x2
# Array: 2x2, architecture=passive, technology=sky130
#
# Simulates the FeCIM crossbar as a resistive MVM (matrix-vector multiply).
# Each cell is modeled as Ron (LRS) or Roff (HRS) based on its programmed level.
# Column-sum currents represent the MVM output (weighted sum of row inputs).
#
# Usage:
#   pip install PySpice pyyaml numpy matplotlib
#   python3 run_pyspice.py
#
# Ngspice must be installed:
#   Ubuntu: sudo apt install ngspice
#   macOS:  brew install ngspice
#   PyPI:   pip install PySpice[ngspice]  (bundled ngspice)

import sys
import os
import numpy as np

try:
    from PySpice.Spice.Netlist import Circuit
    from PySpice.Unit import *
    from PySpice.Spice.NgSpice.Shared import NgSpiceShared
except ImportError:
    print("PySpice not installed. Install with:")
    print("  pip install PySpice")
    sys.exit(1)

# ── Array Configuration ─────────────────────────────────────────────────────
ROWS = 2
COLS = 2
ARCHITECTURE = "passive"

# Conductance range (Siemens)
G_MAX = 10.000000e-6          # µS → S (low-resistance state, LRS)
G_MIN = 0.001000000e-6          # µS → S (high-resistance state, HRS)
G_RANGE = G_MAX - G_MIN

# Wire resistance (parasitic line resistance per row)
R_WIRE_OHM = 2.00         # Ω total word-line resistance

# Supply voltage
V_READ = 0.1              # V read voltage (small to avoid disturb)
V_SUPPLY = 1.8            # V supply (for TIA bias)

# ── Build Weight Matrix ─────────────────────────────────────────────────────
# Random conductance levels (0 to 30 discrete states → linear mapping to G)
np.random.seed(42)
N_LEVELS = 30
level_matrix = np.random.randint(0, N_LEVELS + 1, size=(ROWS, COLS))
G_matrix = G_MIN + level_matrix / N_LEVELS * G_RANGE  # Conductance per cell (S)
R_matrix = 1.0 / G_matrix                              # Resistance per cell (Ω)

print(f"FeCIM PySpice Simulation: fecim_crossbar_2x2")
print(f"Array: {{ROWS}}x{{COLS}} = {{ROWS*COLS}} cells")
print(f"G range: [{{G_MIN*1e6:.4f}}, {{G_MAX*1e6:.4f}}] µS")
print(f"R range: [{{(1/G_MAX):.0f}}, {{(1/G_MIN):.0f}}] Ω")
print("")

# ── Build SPICE Netlist ─────────────────────────────────────────────────────
circuit = Circuit('FeCIM_Crossbar_2x2')

# Ground node
circuit.raw_spice += '.global GND\\n'

# Row voltage sources (DAC outputs — uniform input for test)
input_vector = np.ones(ROWS)  # Uniform input (test vector)

for row in range(ROWS):
    v_in = V_READ * input_vector[row]
    circuit.V(f'row{{row}}', f'WL{{row}}', circuit.gnd, v_in)

# FeCIM cells as resistors (R = 1/G per programmed state)
# Row wire resistance modeled as series resistors
for row in range(ROWS):
    for col in range(COLS):
        r_cell = R_matrix[row, col]
        r_wire_seg = R_WIRE_OHM / COLS  # Distribute wire R across WL (COLS segments per word line)

        # Word-line wire segment (parasitic)
        circuit.R(f'Rwire_{{row}}_{{col}}', f'WL{{row}}_{{col}}',
                  f'WL{{row}}_{{col+1}}' if col < COLS-1 else f'WL{{row}}',
                  r_wire_seg)

        # Ferroelectric cell (programmed conductance)
        circuit.R(f'Rcell_{{row}}_{{col}}', f'WL{{row}}_{{col+1}}' if col < COLS-1 else f'WL{{row}}',
                  f'BL{{col}}', r_cell)

# Column virtual ground (TIA sense amplifier holds BL at virtual GND)
for col in range(COLS):
    circuit.V(f'Vtia_{{col}}', f'BL{{col}}', circuit.gnd, 0)  # V=0 (TIA virtual ground)

# ── Run DC Analysis ─────────────────────────────────────────────────────────
print("Running Ngspice DC analysis...")
print(f"Netlist: {{len(list(circuit.elements))}} elements")

try:
    simulator = circuit.simulator(temperature=25, nominal_temperature=25)
    analysis = simulator.operating_point()

    # Column currents (MVM output)
    col_currents = np.array([
        float(analysis[f'vtia{{col}}'])
        for col in range(COLS)
    ])

    # Ideal MVM (no parasitics)
    ideal_output = G_matrix.T @ input_vector  # cols × rows × rows = cols

    print("")
    print("=== FeCIM MVM Result ===")
    for col in range(COLS):
        i_sim = col_currents[col] * 1e6  # A → µA
        i_ideal = ideal_output[col] * V_READ * 1e6  # µA
        print(f"  BL{{col}}: I_sim={{i_sim:+.3f}} µA  I_ideal={{i_ideal:+.3f}} µA")

    # Error analysis
    rmse = np.sqrt(np.mean((col_currents - ideal_output * V_READ) ** 2))
    print(f"")
    print(f"RMSE vs ideal: {{rmse*1e9:.2f}} nA")
    print(f"Relative RMSE: {{rmse / np.mean(np.abs(ideal_output * V_READ)) * 100:.2f}}%")

except Exception as e:
    print(f"Ngspice simulation error: {{e}}")
    print("Make sure ngspice is installed: sudo apt install ngspice")
    sys.exit(1)

print("")
print("Simulation complete.")
print(f"Crossbar: 2x2 cells, {{ARCHITECTURE}} architecture")
