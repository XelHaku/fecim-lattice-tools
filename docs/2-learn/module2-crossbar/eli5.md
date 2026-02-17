# Module 2: Crossbar - Compute-in-Memory

> **Explain Like I'm 5:** Physics does multiplication for free when you use a crossbar array!

---

## The Magic: Ohm's Law is Multiplication

Discovered in 1827, Ohm's Law is simple:

```
Current = Voltage × Conductance
   I    =    V    ×      G

This is just... multiplication! Physics does it for free!
```

**The insight:** If we:
1. Store weights as conductance (G)
2. Send inputs as voltage (V)
3. The current that flows out (I) IS the result!

---

## What is a Crossbar?

It's a grid of wires with a memory cell at each intersection:

```
        Columns (send voltages in)
           V₀    V₁    V₂    V₃
           │     │     │     │
Row 0  ────●─────●─────●─────●────→ I₀ (current out)
           │     │     │     │
Row 1  ────●─────●─────●─────●────→ I₁
           │     │     │     │
Row 2  ────●─────●─────●─────●────→ I₂
           │     │     │     │
Row 3  ────●─────●─────●─────●────→ I₃

● = one memory cell (conductance = weight)
```

---

## How It Computes (Step by Step)

```
Step 1: Store weights in cells
        ┌─────────────┐
        │ G₀₀ G₀₁ G₀₂ │ Row 0
        │ G₁₀ G₁₁ G₁₂ │ Row 1
        │ G₂₀ G₂₁ G₂₂ │ Row 2
        └─────────────┘

Step 2: Apply input voltages to columns
        V₀ = 0.3V
        V₁ = 0.5V
        V₂ = 0.2V

Step 3: Current flows (multiplication happens!)
        Row 0 output = G₀₀×V₀ + G₀₁×V₁ + G₀₂×V₂
                     = 0.2×0.3 + 0.8×0.5 + 0.4×0.2
                     = 0.06 + 0.40 + 0.08 = 0.54A

Step 4: All rows compute in PARALLEL!
        Row 0: 0.54A
        Row 1: 0.42A
        Row 2: 0.61A
```

**This is matrix-vector multiplication!**

---

## Why This is Powerful

### Traditional CPU
```
To multiply a 1000×1000 matrix by a vector:
- 1 million multiplications
- Each one: fetch → multiply → store
- Sequential: one at a time
- Time: millions of clock cycles
- Energy: billions of transistor switches
```

### Crossbar
```
To multiply a 1000×1000 matrix by a vector:
- 1 million multiplications
- All happen AT THE SAME TIME
- Parallel: all in one analog operation
- Time: ~nanoseconds
- Energy: physics doing the work
```

**Same result. 1000× faster. 1000× less energy.**

---

## The Three Real-World Problems

Real crossbars aren't perfect. Three things go wrong:

### 1. IR Drop (Voltage Gets Weak)

```
Sent voltage: 1.0V ───→ 0.95V ───→ 0.90V ───→ 0.85V
                          ↓ gets weaker as it travels!

Why? Wires have resistance (like friction)
```

**Impact:** Inputs at the end of a long wire are weaker, so outputs are wrong.

**Solution:** Interleaving decoders, distributed drivers, or shorter wires.

---

### 2. Sneak Paths (Current Takes Shortcuts)

```
Goal: Current through only the target cell (●)
      →→→●→→→

Reality: Current is lazy and takes ALL paths!
        →→→●→→→
        ↓   ↑
       →→→●→→→ (snuck through unselected cells!)

Current leaks through unwanted cells!
```

**Impact:** The current you measure isn't just from the target cell.

**Solution:** Selector devices (like diodes) at each cell to block backflow.

---

### 3. Variation (Each Cell is Slightly Different)

```
Set two cells to "0.5 conductance":
Cell A: 0.48 (slightly less conductive)
Cell B: 0.52 (slightly more conductive)

Why? Fabrication isn't perfect (atoms arrange differently)
```

**Impact:** You get ±2-5% errors per cell, which add up.

**Solution:** Error correction, redundancy, or careful characterization.

---

## The Demo: Three Views

Module 2 shows you three tabbed views:

### View 1: Conductance Map
```
┌─────────────────────────────────┐
│ ░░░ ░░░ ███ ░░░                │
│ ░░░ ██░ ░░░ ░░░                │
│ ███ ░░░ ░░░ ░░░                │
│ ░░░ ░░░ ░░░ ██░                │
└─────────────────────────────────┘

Dark = high conductance (strong weight)
Light = low conductance (weak weight)

Click a cell to select it and see details.
```

### View 2: IR Drop Analysis
```
Shows voltage loss along each row/column.
Highlights problem areas that need fixing.
```

### View 3: Sneak Path Visualization
```
Shows unwanted current leakage.
Demonstrates why selector devices matter.
```

---

## Multi-Level Baseline (30 States)

Instead of binary (0 or 1), we store multiple levels:

```
Conductance Levels (30-level demo baseline):

0.00 ├─ Level 0  (lowest)
0.03 ├─ Level 1
0.07 ├─ Level 2
⋮    ├─ ⋮
0.50 ├─ Level 15  (middle)
⋮    ├─ ⋮
0.93 ├─ Level 29
1.00 ├─ Level 30  (maximum)

More levels = more precision per cell!
```

---

## Key Parameters for Design

| Parameter | Symbol | Typical | Trade-off |
|-----------|--------|---------|-----------|
| **Array Size** | M×N | 128×128 | Larger = more compute, but more IR drop |
| **Max Voltage** | V_max | 1.0V | Higher = stronger computation, but more power |
| **Min Conductance** | G_min | 0.001 S | Lower = smaller leakage, but harder to control |
| **Wire Resistance** | R_wire | 0.1 Ω | Lower = less IR drop, but need better wires |

---

## How to Run Module 2

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools crossbar
```

**Try this:**
1. Click a cell in the conductance map
2. Watch the IR drop and sneak paths change
3. Adjust the array size slider
4. See how problems get worse with larger arrays
5. Think about solutions!

---

## The Full Inference Path

From input to output:

```
Digital Input (bits)
    │
    ▼
┌─────────┐
│   DAC   │ ← Digital-to-Analog Converter
│         │   (turns bits into voltages)
└────┬────┘
     │ Analog voltages
     ▼
┌──────────────────────────────┐
│   CROSSBAR ARRAY             │  ← Physics does
│   (matrix-vector multiply)   │     multiplication!
└────┬───────────────────────────┘
     │ Analog currents
     ▼
┌─────────┐
│   TIA   │ ← Transimpedance Amplifier
│         │   (current to voltage)
└────┬────┘
     │ Small voltage
     ▼
┌─────────┐
│   ADC   │ ← Analog-to-Digital Converter
│         │   (voltages back to bits)
└────┬────┘
     │
     ▼
Digital Output (bits)
```

---

## Key Takeaways

| Concept | Remember This |
|---------|---------------|
| **Ohm's Law** | I = V × G (physics does multiplication) |
| **Crossbar** | Grid where all multiplications happen at once |
| **MVM** | Matrix-Vector Multiplication (core of AI) |
| **IR Drop** | Voltage weakens traveling along wires |
| **Sneak Paths** | Current leaks through unselected cells |
| **Variation** | Each cell is slightly different |
| **30 Levels** | Demo baseline for analog weight precision |

---

## Real Crossbars in Practice

Modern ReRAM, MRAM, and PCM crossbars solve these problems with:
- Selector devices (1T1R, 1D1R) to block sneak paths
- Careful circuit design to minimize IR drop
- Statistical testing to characterize variation
- Error correction codes to handle noise

Ferroelectric CIM uses the same techniques!

---

## Next Steps

- **Want to see it recognize digits?** Go to [Module 3: MNIST](../module3-mnist/eli5.md)
- **Want to learn about the circuits?** Go to [Module 4: Circuits](../module4-circuits/eli5.md)
- **Back to Module 1?** See [Module 1: Hysteresis](../module1-hysteresis/eli5.md)
- **Back to overview?** See [ELI5 Overview](../eli5-overview.md)
