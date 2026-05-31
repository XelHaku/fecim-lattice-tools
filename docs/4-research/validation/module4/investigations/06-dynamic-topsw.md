# M4-INV-06 Results — Dynamic TOPS/W from Array Config

Implementation:
- Dynamic comparison metrics are computed in `module4-circuits/pkg/gui/comparison_metrics.go` (`computeComparisonMetrics`).
- Validation test: `module4-circuits/pkg/gui/m4_investigations_test.go::TestM4INV06_DynamicTOPSWMetrics`.

## FeFET row outputs from test

| Array | Latency (ns) | Energy (pJ) | TOPS/W | Energy/op (pJ) |
|---|---:|---:|---:|---:|
| 8×8 | 76.0 | 2.90 | 0.0017 | 0.0453 |
| 16×16 | 304.0 | 11.60 | 0.0017 | 0.0453 |
| 32×32 | 1216.0 | 46.40 | 0.0017 | 0.0453 |
| 64×64 | 4864.0 | 185.60 | 0.0017 | 0.0453 |

Conclusion: comparison view now uses architecture-aware computed metrics instead of static strings; energy/op and TOPS/W remain consistent under the current linear scaling model.