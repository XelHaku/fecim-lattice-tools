# M4-INV-03 Results — Half-Select Disturb Budget (0T1R vs 1T1R)

Source test: `module4-circuits/pkg/gui/m4_investigations_test.go::TestM4INV03_HalfSelectDisturbBudget`

Method: use current residue model constants from GUI write path (`delta_per_pulse = 0.01 * halfSelectDisturbRate`, with `halfSelectDisturbRate=0.25`).

## Results

- **0T1R**
  - Disturb increment per pulse: `0.002500 level/pulse`
  - Cycles to drift 1 level: `ceil(1 / 0.0025) = 400 cycles`

- **1T1R** (conservative 20× attenuation assumption from test)
  - Disturb increment per pulse: `0.000125 level/pulse`
  - Cycles to drift 1 level: `ceil(1 / 0.000125) = 8000 cycles`

Conclusion: current model predicts ~20× better half-select disturb immunity for 1T1R than 0T1R.