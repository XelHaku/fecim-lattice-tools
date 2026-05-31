# M4-INV-01 — Read margin vs selector Ron

Method: `TestM4INV01_ReadMarginVsSelectorRon` (`module4-circuits/pkg/arraysim/m4_investigations_test.go`).

Margin (LSB) from sense-chain current LSB with series selector Ron:

| array_size | Ron (Ohm) | margin_LSB | loss_vs_Ron0_LSB |
|---|---:|---:|---:|
| 64x64 | 0 | 1875.50 | 0.00 |
| 64x64 | 100 | 1865.24 | 10.26 |
| 64x64 | 500 | 1825.30 | 50.20 |
| 64x64 | 1000 | 1777.73 | 97.77 |
| 64x64 | 5000 | 1470.98 | 404.52 |
| 64x64 | 10000 | 1210.00 | 665.50 |
| 128x128 | 0 | 1875.50 | 0.00 |
| 128x128 | 100 | 1865.24 | 10.26 |
| 128x128 | 500 | 1825.30 | 50.20 |
| 128x128 | 1000 | 1777.73 | 97.77 |
| 128x128 | 5000 | 1470.98 | 404.52 |
| 128x128 | 10000 | 1210.00 | 665.50 |

Result: selector Ron monotonically reduces read margin; kΩ-range Ron causes 100s of LSB loss.
