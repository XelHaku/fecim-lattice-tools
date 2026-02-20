<!-- Category: ELI5 | Module: module2-crossbar | Reading time: ~5 min -->
# Module 2 ELI5: Compute-in-Memory with Crossbar Arrays

> A crossbar array does matrix multiplication using Ohm's Law --
> physics computes for free in parallel.

## The Big Idea: I = V x G

Ohm's Law says: current equals voltage times conductance. That is just
multiplication. If you store a number as a conductance (how easily current
flows through a device) and send in a voltage, the current that comes out
IS the product. No transistors flipping. No clock cycles. Physics does it.

Now put thousands of these devices in a grid:

```
        Columns (input voltages)
           V0    V1    V2    V3
           |     |     |     |
Row 0  ----*-----*-----*-----*----> I0  (output current)
           |     |     |     |
Row 1  ----*-----*-----*-----*----> I1
           |     |     |     |
Row 2  ----*-----*-----*-----*----> I2
           |     |     |     |
Row 3  ----*-----*-----*-----*----> I3

  * = one memory cell with conductance G
  Each row current = sum of (V_col x G_cell) for all columns
```

Each row's output current is the dot product of the input voltage vector with
that row's conductance values. All rows compute simultaneously. That is a
full matrix-vector multiplication in one analog step.

## Why Parallel?

A traditional CPU computes one multiplication at a time (or a small batch).
For a 1000x1000 matrix, that is a million multiplications done sequentially.

A crossbar does all million multiplications at the same time. The voltages
propagate through all cells simultaneously. Kirchhoff's current law sums the
currents automatically at each row output. No loop needed.

## The Sneak Path Problem

In a passive crossbar (no transistor per cell), current can leak through
unintended paths:

```
  Goal: measure current through cell (1,1) only

  V --> [Cell 1,1] --> measure here
        |
        v  (current leaks down)
       [Cell 2,1]
        |
        v  (leaks across)
       [Cell 2,2] <-- V
```

Current "sneaks" through neighboring cells, corrupting the measurement.
Solutions: add a transistor per cell (1T1R architecture) or use selector
devices. The simulator models both 0T1R (passive, with sneak paths) and
1T1R (with transistor isolation).

## IR Drop

Wires have resistance. As current flows along a wire, the voltage drops:

```
  Applied: 1.0V ----> 0.95V ----> 0.90V ----> 0.85V
                         (voltage weakens with distance)
```

Think of water pressure dropping as it flows through a long pipe. Cells far
from the driver see less voltage, so their computations are slightly wrong.

## Non-Idealities Summary

| Problem | Cause | Effect | Model |
|---------|-------|--------|-------|
| IR drop | Wire resistance | Voltage weakens along rows/columns | Cumulative R * I drop |
| Sneak paths | Passive crossbar leakage | Extra current corrupts outputs | 3-cell path enumeration |
| Device variation | Manufacturing differences | Each cell slightly off-target | Gaussian noise per cell |
| Conductance drift | Ionic migration over time | Stored values shift slowly | Power-law G(t) model |
| Temperature | Thermal effects on R, G | Wire resistance up, polarization down | Arrhenius + TCR models |

## The Full Inference Path

From digital input to digital output:

```
  Digital bits
      |
      v
  +-------+
  |  DAC  |  (Digital-to-Analog: bits --> voltages)
  +-------+
      |
      v
  +------------------+
  | CROSSBAR ARRAY   |  (Physics does V x G --> I)
  | (matrix multiply)|
  +------------------+
      |
      v
  +-------+
  |  TIA  |  (Transimpedance Amplifier: current --> voltage)
  +-------+
      |
      v
  +-------+
  |  ADC  |  (Analog-to-Digital: voltage --> bits)
  +-------+
      |
      v
  Digital bits
```

The crossbar is the compute engine. The DAC, TIA, and ADC are the peripheral
circuits that interface between digital and analog worlds.

## Key Takeaways

| Concept | Remember This |
|---------|---------------|
| Ohm's Law | I = V x G -- physics does multiplication |
| Crossbar | Grid where all multiplications happen at once |
| MVM | Matrix-Vector Multiplication -- the core of neural network inference |
| IR Drop | Voltage weakens traveling along wires |
| Sneak Paths | Current leaks through unselected cells (0T1R only) |
| Variation | Each cell is slightly different from manufacturing |
| 30 Levels | Simulation baseline for analog weight precision |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
