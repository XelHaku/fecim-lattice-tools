# FeCIM External-Tool Compatibility Matrix

| FeCIM Module | External Tool | Comparison Type | Apples-to-Apples? | Caveats |
|---|---|---|---|---|
| Module 1 (Hysteresis / P-E) | Heracles | Quantitative (RMSE/MAE/max-error on P-E) | Partial | Uses digitized baseline CSV when Heracles binary is unavailable; model-parameter mismatch can dominate absolute error. |
| Module 2 (Crossbar arrays) | CrossSim | Trend | No (conceptual-trend) | Different device compact-model internals and peripheral assumptions; compare monotonicity/sensitivity, not absolute currents. |
| Module 2 (Crossbar arrays) | ngspice | Quantitative (small resistive-array current fixtures) | Yes for resistive netlist, Partial for FeFET behavior | Optional external-tool gate; validates circuit consistency of generated resistive fixtures, not calibrated FeFET compact-model behavior. |
| Module 3 (MNIST inference) | CrossSim + NumPy/SciPy helpers | Trend / Conceptual | No (conceptual-trend) | Dataset pre/post-processing, quantization policy, and ADC/DAC noise models differ. |
| Module 4 (Circuit-level integration) | ngspice | Quantitative (netlist-level sanity currents/voltages) | Partial | FeFET model abstraction in generated netlist is simplified and not foundry-calibrated. |
| Module 6 (EDA export) | ngspice | Quantitative (SPICE round-trip) | Yes (syntax/parse), Partial (behavior) | Syntax and parser checks are apples-to-apples; transistor-level realism remains approximate. |
| Module 6 (EDA export) | iverilog / Verilator | Quantitative (RTL compile/sim) | Yes | Validates digital syntax/structure only; does not validate ferroelectric analog behavior. |
| Module 6 (EDA flow) | OpenROAD / OpenLane2 | Conceptual / Trend | No (conceptual) | Flow compatibility and artifact generation validated; timing/power numbers not silicon-calibrated. |

## Caveat Summary

- **Apples-to-apples strongest at syntax/interface level** (SPICE/Verilog parsing, tool acceptance).
- **Physics-level cross-tool comparisons are mostly partial/trend-based** due to differing compact models and calibration sets.
- **Heracles comparator fallback mode** is explicitly supported via stored baseline CSV when no Heracles executable is installed.
