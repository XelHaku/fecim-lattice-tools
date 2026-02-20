<!-- Category: ELI5 | Module: module4-circuits | Reading time: ~6 min -->
# Module 4 ELI5: DAC, ADC, and the Circuit Interface

> CIM doesn't work in isolation -- you need circuits to convert between
> the digital world and the analog crossbar.

## The Big Picture

A compute-in-memory chip has a crossbar array at its heart, but the
rest of the system is digital. Peripheral circuits are the translators
that bridge the gap:

```
Digital      Analog          Analog        Digital
Input   -->  Voltage  -->  Current  -->   Output
        DAC          Crossbar        TIA + ADC

 [0110] --> [0.4 V] --> [12.3 uA] --> [0.82 V] --> [1010]
   ^                                                   ^
   bits                                              bits
```

Every CIM operation -- write, read, or compute -- passes through this
signal chain.

---

## The Three Circuit Heroes

### DAC: Digital-to-Analog Converter

The DAC takes a digital code (a number) and produces a proportional
voltage.

```
  Code 0  -->  0.0 V
  Code 7  -->  0.47 V
  Code 15 -->  1.0 V     (4-bit DAC, 16 levels)
```

**Why 4-bit?** Literature shows 4-bit DAC/ADC is the sweet spot for
CIM systems. Going higher adds precision but also adds area, power,
and latency. At 4 bits you get 16 input levels, which is enough for
most neural network inference tasks.

The formula is straightforward:

```
V_out = (code / (2^N - 1)) * V_ref
```

### TIA: Transimpedance Amplifier

After the crossbar multiplies voltage by conductance, the result is a
tiny current (microamps). The TIA converts current to voltage so the
ADC can measure it:

```
V_out = I_in * R_feedback
```

The TIA also acts as a **virtual ground** -- it holds its input node
at approximately 0 V. This is a critical detail for passive (0T1R)
arrays, because it means all wordline (row) voltages sit at 0 V
during operation.

### ADC: Analog-to-Digital Converter

The ADC takes the TIA's output voltage and quantizes it back to a
digital number:

```
  0.00 V  -->  Code 0
  0.33 V  -->  Code 5
  1.00 V  -->  Code 15    (4-bit ADC, 16 levels)
```

**The energy problem:** ADC typically consumes 40-60% of total system
energy in CIM architectures. This is why ADC design is one of the most
important optimization targets -- not the crossbar itself.

---

## The Signal Chain End to End

### WRITE (program a cell)

```
  Target level (e.g., 17 out of 30)
       |
       v
  [DAC] --> voltage pulse --> [Crossbar cell] --> polarization changes
       ^                            |
       |                            v
  ISPP controller <--- [TIA+ADC] reads back level
       |
       +--- "too low? increase voltage. too high? decrease."
```

### READ (sense a cell)

```
  Small read voltage (won't disturb)
       |
       v
  [DAC] --> V_read --> [Crossbar cell] --> current out
                                              |
                                              v
                                         [TIA] --> V_out
                                                     |
                                                     v
                                                [ADC] --> digital code
```

### COMPUTE (matrix-vector multiply)

```
  Input vector [v0, v1, v2, v3]
       |
       v
  [DAC array] --> voltages on all wordlines simultaneously
                       |
                       v
                  [Crossbar NxN]
                  Each column sums: I_col = sum(V_i * G_i,j)
                       |
                       v
                  [TIA per column] --> voltages
                       |
                       v
                  [ADC per column] --> output vector
```

This is the magic of CIM: one matrix-vector multiply happens in a
single analog pass through the crossbar.

---

## 0T1R Passive Mode: The Column-Write Constraint

In a passive (0T1R) crossbar, there are no transistor selectors --
just resistive elements at every cross-point. The DAC drives the
bitline (column), and the TIA holds all wordlines (rows) at 0 V.

```
  WL0 ----[cell]----[cell]----[cell]----  0 V (TIA virtual ground)
  WL1 ----[cell]----[cell]----[cell]----  0 V
  WL2 ----[cell]----[cell]----[cell]----  0 V
              |         |         |
            BL0       BL1       BL2
                        |
                    DAC drives this column
                    with -V_write
```

**Key consequence:** You cannot independently select a single cell in
a column. When you write, the entire column sees the write voltage.
The V/2 half-select scheme is impossible in 0T1R because wordline
voltages cannot be independently controlled.

| Architecture | Target Cell | Same-Row | Same-Col | Other | Sneak Current |
|-------------|-------------|----------|----------|-------|---------------|
| 0T1R | Full V | 0 V (safe) | Full V (disturbed) | 0 V | Yes |
| 1T1R | Full V | Suppressed | Suppressed | Isolated | Minimal |
| 2T1R | Full V | Isolated | Isolated | Isolated | None |

---

## ISPP: Finding the Right Voltage

Programming a ferroelectric cell to an exact conductance level is like
a binary search:

```
  "I want level 17 out of 30."

  Step 1: Try mid-range voltage  --> got level 22 (too high!)
  Step 2: Try lower voltage      --> got level 14 (too low!)
  Step 3: Try between them       --> got level 18 (close!)
  Step 4: Nudge down slightly    --> got level 17 (done!)
```

ISPP (Incremental Step Pulse Programming) applies voltage pulses,
reads back the result, and adjusts. It maintains a search bracket
[V_min, V_max] that narrows with each iteration until the cell
reaches the target level.

Materials with sharp switching thresholds may overshoot repeatedly --
this is a physics limitation, not a bug. The controller recognizes
this and accepts the closest achievable level.

---

## What the Simulator Simplifies

- Circuit behaviors use idealized transfer functions, not SPICE models.
- Noise and nonlinearity are simplified when enabled.
- Timing and power are analytic estimates, not transistor-level results.
- The charge pump model is Dickson-style with simplified losses.

---

## Next Steps

- [PHYSICS.md](PHYSICS.md) -- the equations behind each component.
- [FEATURES.md](FEATURES.md) -- what is implemented vs planned.
- [OPENSOURCE-TOOLS.md](OPENSOURCE-TOOLS.md) -- external circuit tools.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
