# Module 4 Investigation Results

## Lane 2: Wordline RC Practical Ceiling (M4-INV-02)

Assumptions/method used:
- Wire geometry from `module4-circuits/pkg/arraysim/types.go` `DefaultCellGeometry()`:
  - `PitchX = 0.46 um`, `WireWidth = 80 nm`, `WireThickness = 160 nm`, `MetalResistivity = 2.2e-8 ohm*m`
- Per-segment resistance: `R_segment = rho * PitchX / (WireWidth * WireThickness)`
- Per-cell capacitance for 1T1R write path: `C_segment = C_wire + C_gate`
  - `C_wire ~= 0.1 fF/um * PitchX(um) = 0.046 fF`
  - `C_gate` tech assumptions: SKY130=3.0 fF, 65nm=1.8 fF, 28nm=1.0 fF, 14nm=0.6 fF
- Elmore WL delay: `tau = 0.5 * (N * R_segment) * (N * C_segment)`
- Write pulse width reference: `10 ns` (HZO FeFET typical)

| Tech Node | Array N | R_seg (ohm) | C_seg (fF) | tau_WL (ns) | Pulse Width (ns) | Viable? |
|---|---:|---:|---:|---:|---:|---|
| SKY130 | 8 | 0.790625 | 3.046 | 0.000077 | 10.0 | Yes |
| SKY130 | 16 | 0.790625 | 3.046 | 0.000308 | 10.0 | Yes |
| SKY130 | 32 | 0.790625 | 3.046 | 0.001233 | 10.0 | Yes |
| SKY130 | 64 | 0.790625 | 3.046 | 0.004932 | 10.0 | Yes |
| SKY130 | 128 | 0.790625 | 3.046 | 0.019728 | 10.0 | Yes |
| SKY130 | 256 | 0.790625 | 3.046 | 0.078913 | 10.0 | Yes |
| 65nm | 8 | 1.027812 | 1.846 | 0.000061 | 10.0 | Yes |
| 65nm | 16 | 1.027812 | 1.846 | 0.000243 | 10.0 | Yes |
| 65nm | 32 | 1.027812 | 1.846 | 0.000971 | 10.0 | Yes |
| 65nm | 64 | 1.027812 | 1.846 | 0.003886 | 10.0 | Yes |
| 65nm | 128 | 1.027812 | 1.846 | 0.015543 | 10.0 | Yes |
| 65nm | 256 | 1.027812 | 1.846 | 0.062172 | 10.0 | Yes |
| 28nm | 8 | 1.423125 | 1.046 | 0.000048 | 10.0 | Yes |
| 28nm | 16 | 1.423125 | 1.046 | 0.000191 | 10.0 | Yes |
| 28nm | 32 | 1.423125 | 1.046 | 0.000762 | 10.0 | Yes |
| 28nm | 64 | 1.423125 | 1.046 | 0.003049 | 10.0 | Yes |
| 28nm | 128 | 1.423125 | 1.046 | 0.012195 | 10.0 | Yes |
| 28nm | 256 | 1.423125 | 1.046 | 0.048778 | 10.0 | Yes |
| 14nm | 8 | 1.897500 | 0.646 | 0.000039 | 10.0 | Yes |
| 14nm | 16 | 1.897500 | 0.646 | 0.000157 | 10.0 | Yes |
| 14nm | 32 | 1.897500 | 0.646 | 0.000628 | 10.0 | Yes |
| 14nm | 64 | 1.897500 | 0.646 | 0.002510 | 10.0 | Yes |
| 14nm | 128 | 1.897500 | 0.646 | 0.010042 | 10.0 | Yes |
| 14nm | 256 | 1.897500 | 0.646 | 0.040167 | 10.0 | Yes |

Conclusion: **Maximum practical array size at SKY130 with 10ns pulse: N = 256** (upper bound of evaluated set; all tested sizes are viable).
