#!/usr/bin/env python3
# FeCIM CrossSim Runner
# Generated: 2026-02-20
# Design: fecim_crossbar_2x2
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

CONFIG_FILE = "fecim_crossbar_2x2_crosssim.yaml"

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
