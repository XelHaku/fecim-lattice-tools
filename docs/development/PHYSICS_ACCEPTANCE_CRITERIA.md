# Physics Accuracy Acceptance Criteria (Cross-Module)

Status: Active baseline for CM-P1.
Source-of-truth policy for external claims: `docs/comparison/HONESTY_AUDIT.md`.

## Rule format

Each module must define pass/fail in numeric terms:
- **Metric** (what is measured)
- **Reference** (code formula, golden data, or cited literature)
- **Tolerance** (absolute/relative bound)

## Module criteria

### Module 1 - Hysteresis
- **Loop regression**: RMS(E), RMS(P) vs golden baseline <= **2%** of full-scale.
- **Material parameter checks** (Ec/Pr ranges used in tests): within documented bounds with **10% engineering tolerance** unless test defines tighter bounds.
- **ISPP level-hit**: level error <= **±1 level** (default calibration tolerance).

### Module 2 - Crossbar
- **Formula consistency** (Ohm/KCL helpers): unit tests within **0.1% to 1%** depending on test case.
- **MVM with non-idealities**: output error within test-defined bounds; default stress tolerance up to **10%** for quantization/non-ideality cases.
- **IR-drop/solver convergence**: solver tolerance in volts (default `1e-6 V`) and scenario checks must pass.

### Module 3 - MNIST
- **Pipeline order correctness**: CIM path order fixed to DAC -> MVM -> ADC -> noise -> softmax (must match tests).
- **Quantization/noise bounds**: UI and core limits must match code clamps (no hidden larger range).
- **Mode agreement metrics**: FP vs CIM metrics must be computed by shared core paths and tested.

### Module 4 - Circuits
- **Read chain correctness**: DAC -> array current -> TIA -> ADC end-to-end tests must pass exact/sign checks for known fixtures.
- **Analog block equations** (DAC/ADC/TIA/charge pump): implementation must match documented equations; tolerance per unit tests (typically small numeric epsilon to LSB-level bounds).
- **Coupling behavior**: default read path must use array-level coupling (Tier A+) and remain covered by regression tests.

### Module 5 - Comparison
- **Honesty-first gate**: outputs are model comparisons; no external performance numbers shown as facts unless verified in `HONESTY_AUDIT.md`.
- **Labeling requirement**: all comparative metrics must be labeled modeled/reported/verified explicitly.

### Module 6 - EDA
- **Export correctness**: JSON/CSV/SPICE/Verilog/DEF outputs must be structurally valid and round-trip tested where available.
- **Mapping correctness**: sign handling, level quantization, and indexing must match documented mapping rules.
- **CLI/GUI parity**: same config input should produce equivalent compile outputs for shared features.

### Module 7 - Docs
- **Physics truthfulness**: documentation cannot upgrade unverified claims to facts.
- **Traceability**: claims must either link to testable code behavior or `HONESTY_AUDIT.md`.

## Reporting template

When reporting results, always include measured value, expected value, and tolerance, e.g.:

`P_r = 24.7 uC/cm^2 (expected 25.0, delta 1.2%, within 5% tolerance)`
