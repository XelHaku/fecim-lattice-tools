# Error Propagation in FeCIM Simulations

## Overview

Uncertainty enters the simulation at three levels:

1. **Parameter uncertainty** -- confidence tiers from the ConfidenceLedger
2. **Device variation** -- process variation applied during MVM
3. **Numerical error** -- solver discretization and clamping

## Uncertainty Sources

### Parameter Confidence (confidence_ledger.go)

| Tier | Confidence | Typical sigma/value | Example Parameters |
|------|-----------|---------------------|-------------------|
| Measured | 0.90--0.98 | 2--5% | Pr, Ec, thickness, epsilon (HZO literature) |
| Calibrated | 0.75--0.86 | 5--15% | beta\_landau, gamma\_landau, tau\_inf (Landau fitted) |
| Estimated | 0.60--0.72 | 15--25% | rho\_viscosity, Q12, series\_resistance |
| Placeholder | 0.20--0.40 | 25--50% | imprint\_field, temp\_coeff\_ec, temp\_coeff\_pr |

Confidence scores are assigned in `NewConfidenceLedger()`. Unknown parameters
default to `Placeholder` with confidence 0.1 via `TagOutput()`.

### Device Variation (crossbar/array.go)

Process variation is applied during MVM via `GetProcessVariationFactor()`:

```
G_effective = G_nominal * random * gradX * gradY * edgeFactor
```

- **random**: per-cell `NoiseFactor = 1 + sigma * N(0,1)`, drawn at array creation
- **gradX / gradY**: systematic spatial gradient from array center (default 0.1%/cell)
- **edgeFactor**: boundary cell degradation (default 5%)
- **DeviceSigma**: default 2% (configurable via `ProcessVariationConfig`)

Variation is applied per-cell, per-read in both `mvmCPU` and `mvmGPU` paths.

### Numerical Error (physics/landau.go)

- **RK4 integration**: O(dt^4) local error, O(dt^3) global error
- **Rate clamping**: `maxAbsRate = 1e12 C/(m^2*s)` reduces effective order to ~1.0 near switching transitions
- **Polydomain ensemble**: converged at N=100 domains (delta < 2% vs N=200); tested over sweep N in {5, 10, 20, 50, 100, 200}
- **DAC/ADC quantization**: fixed-point truncation adds +/-0.5 LSB noise per conversion

## Research Trace Pipeline

The `ResearchTrace` (`research_trace.go`) tracks uncertainty through each
analog inference stage. Each stage produces a `TraceValue` with `Value`,
`Uncertainty` (1-sigma absolute), and `Unit` (SI string).

```
DAC (quantization noise +/-0.5 LSB, Vref uncertainty 1 mV)
  -> Array MVM (conductance variation +/-5%, IR drop +/-25%)
    -> TIA (input-referred noise 20 nA/sqrt(Hz), gain uncertainty +/-1%)
      -> ADC (quantization noise +/-0.5 LSB_rms)
        -> Classifier (logit uncertainty +/-0.15, probability +/-3%)
```

`BuildResearchTrace()` constructs a consistent, unit-tagged sample path using
these hardcoded uncertainty budgets. The uncertainties in `ArrayTrace` reflect
the `CellConductance` 5% sigma and a 3% bitline current uncertainty.

## Known Limitations

1. **MVM does not return uncertainty**: `MVM()` applies device variation
   internally but returns only `[]float64` -- no uncertainty bands on output.
2. **No covariance tracking**: parameters are assumed independent; no
   correlation matrix is maintained across confidence tiers.
3. **Monte Carlo is test-only**: the 200-trial P-E uncertainty quantification
   (`validation/m1_montecarlo_uncertainty_test.go`) runs in CI but is not
   exposed as a production API.
4. **Variation is frozen at init**: `NoiseFactor` is sampled once per cell at
   `NewArray()` time. Cycle-to-cycle (C2C) read noise is not applied during
   MVM unless explicitly enabled via `StateDepC2CConfig`.

## Recommended Workflow

For publication-quality uncertainty analysis:

1. Run Monte Carlo test:
   `go test ./validation/ -run MonteCarlo -v`
2. Inspect parameter tiers:
   `ledger.ExportForReproPack()` returns JSON with provenance and confidence.
3. Per-stage breakdown:
   `BuildResearchTrace(inputCode, gCell, tiaFeedback, adcBits)` returns a
   `ResearchTrace` with uncertainty at each stage.
4. Report confidence intervals from Monte Carlo distributions, not single-run
   nominal values.

## Future Work

- Extend `MVM()` signature to return `(values, uncertainties, error)`.
- Implement covariance-aware propagation via Jacobian matrices.
- Promote Monte Carlo runner from test-only to a production CLI command.
- Add uncertainty bands to CSV/JSON export artifacts.
