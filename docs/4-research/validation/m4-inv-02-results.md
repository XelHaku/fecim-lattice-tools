# M4-INV-02 Results — Wordline RC Delay Ceiling

Source test: `module4-circuits/pkg/arraysim/m4_investigations_test.go::TestM4INV02_WordlineRCDelayBudget`

Method: per-node `Rseg` and `Ccell` sweep, find maximum `N` where WL delay remains `<= 10ns`.

## Results

| Tech node | Rseg (Ω/segment) | Ccell (fF) | Max N before delay > 10ns |
|---|---:|---:|---:|
| 130nm | 0.215 | 2.200 | >4096 (not exceeded in sweep) |
| 65nm | 0.249 | 1.100 | >4096 (not exceeded in sweep) |
| 28nm | 0.187 | 0.550 | >4096 (not exceeded in sweep) |
| 14nm | 0.220 | 0.320 | >4096 (not exceeded in sweep) |

Conclusion: with current compact RC assumptions, 10ns WL-delay limit is not the bottleneck up to at least 4096 cells/line.