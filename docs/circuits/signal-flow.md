# Circuits Signal Flow (Text Diagram)

This diagram describes how Module 4 peripheral circuits drive and sense FeFET/ferroelectric crossbar cells in **WRITE**, **READ**, and **MVM** modes.

## Reference Parameters (HZO-like operating point)

- Ferroelectric thickness: `d_FE = 10 nm`
- Coercive field: `E_c = 1.16 MV/cm = 1.16e8 V/m`
- Coercive voltage estimate: `V_c = E_c * d_FE ≈ 1.16 V`
- Practical write DAC range (with margin for ISPP/program-verify): `V_DAC ≈ 0 to 3.0 V`
- Half-select bias: `V/2` scheme for non-target lines (disturb reduction)
- Read voltage: `V_read ≈ 0.1 to 0.3 V` (non-destructive read region)
- TIA feedback resistor: `R_f = 10 kΩ`
- ADC: `5-bit SAR` (32 codes)
- Example ADC full-scale: `V_min = -1.0 V`, `V_max = +1.0 V`, `V_LSB = (V_max - V_min)/32 = 62.5 mV`

---

## 1) WRITE MODE (Programming a Cell)

```text
[User target selection]
  User picks cell [row, col] and target conductance G_target
            |
            v
[Digital program code / ISPP step index]
            |
            v
[DAC]
  Converts digital code -> analog V_DAC (0..~3.0 V)
            |
            v
[Wordline driver]
  WORDLINE[row] = V_DAC
  Other wordlines = V_DAC/2  (half-select)
            |
            v
[Bitline biasing]
  BITLINE[col] = 0 V (or V/2 scheme for unselected columns)
            |
            v
[Target cell voltage]
  V_cell = V_DAC - V_BL
            |
            v
[E-field in ferroelectric]
  E = V_cell / d_FE
  (d_FE = 10 nm)
            |
            v
[Physics update]
  Preisach/LK solver: E -> P(E, history) -> updated polarization state
            |
            v
[Conductance mapping]
  G = Gmin + (Gmax - Gmin) * (P/Ps + 1)/2
            |
            v
[ISPP verify loop]
  Measure/verify G
  if G not in target window:
      adjust V_DAC (or pulse width/count)
      re-apply pulse
  repeat until G ~= G_target
```

---

## 2) READ MODE (Sensing a Cell)

```text
[Read command for cell row,col]
            |
            v
[DAC read bias]
  Apply V_read to WORDLINE[row] (typically 0.1..0.3 V)
            |
            v
[Bitline front-end]
  BITLINE[col] connected to TIA input (virtual ground ~0 V)
            |
            v
[Cell current]
  I_cell = G_cell * V_read
            |
            v
[TIA conversion]
  V_TIA = -I_cell * R_f
  with R_f = 10 kΩ
            |
            v
[ADC input conditioning]
  V_ADC = clamp(V_TIA, V_min, V_max)
            |
            v
[5-bit SAR ADC]
  Digital code = floor((V_ADC - V_min)/V_LSB)
  where V_LSB = (V_max - V_min)/32
            |
            v
[Output]
  Quantized readout -> estimated conductance/state
```

---

## 3) MVM MODE (Matrix-Vector Multiply)

```text
[Input vector x = {V1, V2, ..., VN}]
            |
            v
[Row DAC bank]
  Apply all Vi simultaneously to WORDLINE[i]
            |
            v
[Crossbar current summation per column j]
  I_col_j = Sum_i ( G[i,j] * V_i )
            |
            v
[Column TIAs]
  Convert each I_col_j -> V_col_j = -I_col_j * R_f
  (R_f = 10 kΩ)
            |
            v
[Column ADCs]
  5-bit SAR digitization of each V_col_j
            |
            v
[Digital output vector y]
  y_j corresponds to quantized column response
  Ideal math form: y = G * x
```

---

## Practical Notes

- `V_DAC` range should be set from material physics (`E_c`, `d_FE`) plus reliability margin, not arbitrarily.
- `V/2` half-select bias mitigates disturb on non-target cells during programming.
- `R_f = 10 kΩ` trades gain vs bandwidth/noise; adjust per array size and expected current.
- 5-bit ADC is often sufficient for compact CIM demonstrations; higher ENOB may be needed for high-accuracy inference.
