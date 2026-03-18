#!/usr/bin/env python3
"""Cross-validation oracle for FeCIM MVM using numpy and optionally badcrossbar.

Usage:
    python3 crossval_badcrossbar.py input.json

Input JSON schema:
    {
        "weights": [[w00, w01, ...], ...],  // 2D weight matrix (normalized 0-1)
        "input_vector": [x0, x1, ...],      // Input vector (normalized 0-1)
        "wire_resistance": {                 // Optional: wire resistance in ohms
            "wordline": 0.0,
            "bitline": 0.0
        },
        "array_size": [rows, cols]           // Array dimensions
    }

Output JSON:
    {
        "ideal_output": [y0, y1, ...],      // W @ x (pure numpy matmul)
        "ir_drop_output": [y0, y1, ...],    // IR-drop-aware MVM (if badcrossbar available)
        "badcrossbar_available": true/false,
        "metadata": {
            "numpy_version": "...",
            "badcrossbar_version": "..." or null,
            "array_rows": N,
            "array_cols": M
        }
    }
"""

import json
import logging
import sys

import numpy as np

# Suppress badcrossbar's verbose logging (it prints to root logger at INFO level)
logging.disable(logging.CRITICAL)


def compute_ideal_mvm(weights, input_vector):
    """Compute ideal MVM: y = W @ x using numpy."""
    W = np.array(weights, dtype=np.float64)
    x = np.array(input_vector, dtype=np.float64)
    return (W @ x).tolist()


def compute_ir_drop_mvm(weights, input_vector, r_wordline, r_bitline):
    """Compute IR-drop-aware MVM using badcrossbar.

    badcrossbar models parasitic wire resistance in passive crossbar arrays.
    It solves the full resistor network including wordline/bitline resistance
    to produce physically accurate output currents.

    badcrossbar API convention:
      - resistances: shape (m, n) where m = wordlines, n = bitlines
      - applied_voltages: shape (m, p) where p = number of examples
      - output currents: shape (p, n) — currents at bitlines

    badcrossbar computes VMM: voltages on wordlines, currents out of bitlines.
    For MVM (y = W @ x, input on columns, output on rows), we transpose:
      - Treat our columns as badcrossbar's wordlines (m = cols)
      - Treat our rows as badcrossbar's bitlines (n = rows)
      - Feed W^T as the resistance matrix
      - Apply input x as voltages on the "wordlines" (cols)
      - Read output currents from "bitlines" (rows)

    Returns (output_currents, version) or (None, version) on failure.
    """
    try:
        import badcrossbar
    except ImportError:
        return None, None

    version = getattr(badcrossbar, "__version__", "unknown")

    W = np.array(weights, dtype=np.float64)
    x = np.array(input_vector, dtype=np.float64)
    rows, cols = W.shape

    # Convert conductance to resistance, avoiding division by zero
    conductances = W.copy()
    conductances[conductances == 0] = 1e-12  # avoid inf resistance
    resistances = 1.0 / conductances

    # Transpose: resistances_T has shape (cols, rows)
    # badcrossbar sees cols as wordlines, rows as bitlines
    resistances_T = resistances.T

    # applied_voltages: shape (m, p) = (cols, 1) — one example
    applied_voltages = x.reshape(-1, 1)

    try:
        solution = badcrossbar.compute(
            applied_voltages,
            resistances_T,
            r_i_word_line=r_wordline if r_wordline > 0 else 1e-12,
            r_i_bit_line=r_bitline if r_bitline > 0 else 1e-12,
        )
        # Output currents: shape (p, n) = (1, rows)
        output_currents = solution.currents.output[0].tolist()
        return output_currents, version
    except TypeError:
        # Older API with single r_i parameter
        try:
            solution = badcrossbar.compute(
                applied_voltages,
                resistances_T,
                r_i=max(r_wordline, r_bitline, 1e-12),
            )
            output_currents = solution.currents.output[0].tolist()
            return output_currents, version
        except Exception:
            return None, version
    except Exception:
        return None, version


def main():
    if len(sys.argv) < 2:
        print("Usage: crossval_badcrossbar.py <input.json>", file=sys.stderr)
        sys.exit(1)

    input_path = sys.argv[1]
    with open(input_path, "r") as f:
        data = json.load(f)

    weights = data["weights"]
    input_vector = data["input_vector"]
    array_size = data.get("array_size", [len(weights), len(weights[0])])
    wire_res = data.get("wire_resistance", {"wordline": 0.0, "bitline": 0.0})
    r_wl = wire_res.get("wordline", 0.0)
    r_bl = wire_res.get("bitline", 0.0)

    # Ideal MVM (numpy only)
    ideal_output = compute_ideal_mvm(weights, input_vector)

    # IR-drop-aware MVM (badcrossbar, optional)
    ir_drop_output = None
    bc_version = None
    bc_available = False

    if r_wl > 0 or r_bl > 0:
        ir_drop_output, bc_version = compute_ir_drop_mvm(
            weights, input_vector, r_wl, r_bl
        )
        bc_available = ir_drop_output is not None

    result = {
        "ideal_output": ideal_output,
        "ir_drop_output": ir_drop_output,
        "badcrossbar_available": bc_available,
        "metadata": {
            "numpy_version": np.__version__,
            "badcrossbar_version": bc_version,
            "array_rows": array_size[0],
            "array_cols": array_size[1],
        },
    }

    print(json.dumps(result))


if __name__ == "__main__":
    main()
