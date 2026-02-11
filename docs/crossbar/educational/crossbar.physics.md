# Crossbar Array Physics: Deep Technical Reference

**FeCIM Lattice Tools - Module 2: Compute-in-Memory Crossbar Arrays**

> Start here for comprehensive physics understanding of crossbar arrays for ferroelectric compute-in-memory.

---

## Overview

This document provides the deep technical foundation for understanding crossbar array physics, from basic principles to real-world non-idealities. It covers:
- Matrix-Vector Multiplication (MVM) using Ohm's and Kirchhoff's laws
- 30-level conductance quantization in ferroelectric devices (demo baseline; simulation baseline)
- Non-idealities: IR drop, sneak paths, device variation, ADC quantization
- Physics-accurate simulation methodology

**Note:** References to “30 levels” refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

---

## Part 1: The Problem We're Solving

### Traditional Computing is Wasteful

In normal computers, data lives in MEMORY and processing happens in the CPU:

```
Traditional Computer:
┌─────────┐     data      ┌─────────┐
│         │ ←──────────→  │         │
│  MEMORY │   (slow bus)  │   CPU   │
│  (RAM)  │ ←──────────→  │         │
└─────────┘               └─────────┘
         ↑
     "von Neumann bottleneck"
     Moving data wastes 90% of energy!
```

**For AI workloads:** Most operations are matrix-vector multiplications (MVMs). The CPU fetches billions of weights from memory, multiplies, and stores back—over and over.

### Compute-in-Memory: The Solution

**What if we computed WHERE the data is stored?**

```
Compute-in-Memory:
┌─────────────────────────────────────┐
│                                     │
│  WEIGHTS LIVE HERE                  │
│  AND                                │
│  MULTIPLY HAPPENS HERE              │
│  (no data movement!)                │
│                                     │
└─────────────────────────────────────┘
           ↑
    10-1000× more energy efficient!
```

This is what Ferroelectric CIM does with **crossbar arrays**.

---

## Part 2: What is a Crossbar Array?

### The Grid Structure

A crossbar is a grid of horizontal and vertical wires with a **memory cell at each intersection**:

```
           Columns (input voltages)
           V₀    V₁    V₂    V₃
           │     │     │     │
           ↓     ↓     ↓     ↓
         ──●─────●─────●─────●──→ I₀ (Row 0 output)
           │     │     │     │
         ──●─────●─────●─────●──→ I₁ (Row 1 output)
           │     │     │     │
         ──●─────●─────●─────●──→ I₂ (Row 2 output)
           │     │     │     │
         ──●─────●─────●─────●──→ I₃ (Row 3 output)

           ● = one memory cell (stores a weight)
```

- **Vertical wires:** Apply input voltages (the data)
- **Horizontal wires:** Collect output currents (the result)
- **Cells at intersections:** Store weights (neural network parameters)

### What's at Each Intersection?

Each cell is a programmable resistor/conductor. For Ferroelectric CIM, it's a ferroelectric device:

```
One cell:
     │ column wire
     │
    ─┼─ ← Ferroelectric capacitor/transistor
     │     (conductance G = weight value)
─────┴───── row wire

Ohm's Law: I = G × V
- V = voltage from column (input)
- G = conductance of cell (stored weight)
- I = current contributed to row (partial result)
```

---

## Part 3: Matrix-Vector Multiplication (MVM)

### What is MVM?

Given a **matrix** W and a **vector** x, compute **y = W × x**:

```
Matrix W:                 Vector x:        Result y:
┌───────────────┐         ┌───┐           ┌───┐
│ w₀₀ w₀₁ w₀₂ │         │ x₀│           │ y₀│
│ w₁₀ w₁₁ w₁₂ │    ×    │ x₁│     =     │ y₁│
│ w₂₀ w₂₁ w₂₂ │         │ x₂│           │ y₂│
└───────────────┘         └───┘           └───┘

y₀ = w₀₀×x₀ + w₀₁×x₁ + w₀₂×x₂
y₁ = w₁₀×x₀ + w₁₁×x₁ + w₁₂×x₂
y₂ = w₂₀×x₀ + w₂₁×x₁ + w₂₂×x₂
```

**In a digital CPU:** Must do each multiplication separately, one after another.

### How Crossbar Does MVM in ONE STEP

The crossbar computes MVM using physics (Ohm's Law + Kirchhoff's Current Law):

```
Step 1: Apply inputs as VOLTAGES on columns
                V₀=x₀  V₁=x₁  V₂=x₂
                  │      │      │
                  ↓      ↓      ↓

Step 2: Current through each cell = G × V (Ohm's Law)

              ──●──────●──────●──→  Sum on row = y₀
                │      │      │
                G₀₀    G₀₁    G₀₂
                ×      ×      ×
                V₀     V₁     V₂

Step 3: Currents on each row sum automatically (Kirchhoff's Law)

         I_row0 = G₀₀×V₀ + G₀₁×V₁ + G₀₂×V₂ = y₀
```

**The physics does the math!** All multiplications and additions happen simultaneously in ~nanoseconds.

### Why This is Amazing

| Approach | Operations | Time |
|----------|------------|------|
| Digital CPU | Multiply each pair, add all | O(n²) sequential |
| Crossbar | Physics does all at once | O(1) parallel! |

For a 1000×1000 matrix: CPU needs 1,000,000 multiplies. Crossbar: 1 analog operation.

---

## Part 4: Conductance = Weight

### Programming the Weights

Each cell's conductance G represents a neural network weight. More conductance = higher weight:

```
Low conductance (small weight):      High conductance (large weight):
         │                                    │
        ─┼─ thin/resistive                   ─┼─ thick/conductive
         │                                    │
    Less current flows                   More current flows
```

For ferroelectric cells, we control conductance by:
- Polarization state (from Demo 1!) controls how conductive the channel is
- Demo baseline (30 levels) → 30 possible weight values per cell

### Weight Range

| Conductance | Weight Value | Meaning |
|-------------|--------------|---------|
| G_min | 0.0 | Minimum weight |
| G_mid | 0.5 | Medium weight |
| G_max | 1.0 | Maximum weight |

**Resolution:** With the 30-level baseline, we get ~5 bits of precision per cell.

---

## Part 5: Non-Idealities (The Real World)

Real crossbars aren't perfect. Here are the problems:

### 1. IR Drop (Voltage Loss Along Wires)

Wires have resistance. Voltage drops as current flows through them:

```
Ideal:                          Real (with IR drop):
   1.0V  1.0V  1.0V  1.0V          1.0V  0.95V  0.9V  0.85V
    │     │     │     │              │     │     │     │
    ↓     ↓     ↓     ↓              ↓     ↓     ↓     ↓
   ─●─────●─────●─────●─→           ─●─────●─────●─────●─→
    │     │     │     │              │     │     │     │
                                  Cells far from source
                                  see lower voltage!
```

**Effect:** Cells at array edges get wrong inputs → computation errors.

### 2. Sneak Paths (Current Takes Wrong Route)

In passive arrays, current can flow through unintended paths:

```
Want: Current through target cell only
                │
              ──●── target
                │

Got: Current "sneaks" through neighbors!
         │          │
      ───●──────────●───
         │    ←     │
         ↓ sneak ↑  │
      ───●──────────●───
         │          │
```

**Effect:** Output includes contributions from wrong cells → incorrect result.

### 3. Device-to-Device Variation

Manufacturing isn't perfect. Each cell has slightly different properties:

```
Programmed same weight (0.5):

Cell A: G = 0.52       Cell B: G = 0.48       Cell C: G = 0.51
        ↑                      ↑                      ↑
     slightly              slightly               close to
      high                   low                   ideal
```

**Effect:** Random errors in every computation.

### 4. ADC Quantization (Limited Output Precision)

The analog output current must be converted to digital. ADCs have limited bits:

```
Actual analog output: 0.7328451...

3-bit ADC levels: 0, 0.14, 0.29, 0.43, 0.57, 0.71, 0.86, 1.0

Quantized output: 0.71 (rounded down)
                     ↑
              Lost precision!
```

**Effect:** Output is approximate, not exact.

---

## Part 6: Why Ferroelectrics Win

| Memory Type | Pros | Cons |
|-------------|------|------|
| RRAM (resistive) | Simple, dense | High Joule heating, sneak paths |
| PCM (phase change) | High density | Very hot during switching |
| MRAM (magnetic) | Fast, endurance | Low density, complex |
| **FeFET/FTJ** | Low power, CMOS-compatible, no heat | Newer technology |

**Ferroelectric CIM uses ferroelectric** because:
- ✅ No Joule heating (displacement current, not filament)
- ✅ Self-rectifying possible (reduces sneak paths)
- ✅ CMOS-compatible (same fab as regular chips)
- ⚠️ 30-level baseline (simulation baseline; high precision)
- ✅ 10¹² cycle endurance

---

## Part 7: Neural Network Connection

### How AI Uses MVM

Neural networks are mostly matrix multiplications:

```
Input layer → Hidden layer → Output layer
    x           W₁ × x          W₂ × (W₁ × x)
              (MVM #1)           (MVM #2)
```

Each layer = one MVM. A typical AI model does billions of MVMs.

### Crossbar = One Layer

Map one layer's weights to one crossbar:

```
Neural network layer:              Crossbar array:

     x₀  x₁  x₂  x₃               V₀  V₁  V₂  V₃
      │   │   │   │                │   │   │   │
      ↓   ↓   ↓   ↓                ↓   ↓   ↓   ↓
  ┌───┴───┴───┴───┴───┐         ──●───●───●───●──→ y₀
  │   FULLY CONNECTED │         ──●───●───●───●──→ y₁
  │      LAYER        │         ──●───●───●───●──→ y₂
  └───┬───┬───┬───┬───┘
      ↓   ↓   ↓   ↓              Same structure!
     y₀  y₁  y₂  y₃
```

### MNIST Example

Handwritten digit recognition:
- Input: 28×28 = 784 pixels
- Output: 10 classes (digits 0-9)
- Crossbar: 784 columns × 10 rows (or multiple smaller arrays)

---

## Summary Table

| Term | Plain English |
|------|---------------|
| **Crossbar Array** | Grid of wires with memory cells at intersections |
| **MVM** | Matrix-vector multiplication (the core AI operation) |
| **Conductance (G)** | How easily current flows through a cell = the weight |
| **Ohm's Law** | Current = Conductance × Voltage (I = G×V) |
| **Kirchhoff's Law** | Currents on a wire add up automatically |
| **IR Drop** | Voltage loss along resistive wires |
| **Sneak Path** | Current flowing through unintended cells |
| **Device Variation** | Each cell is slightly different due to manufacturing |
| **ADC** | Analog-to-Digital Converter (reads output currents) |

---

## Implementation in FeCIM Tools

### Code Location

All crossbar physics is implemented in:
- `module2-crossbar/pkg/crossbar/array.go` - Core MVM
- `module2-crossbar/pkg/crossbar/nonidealities.go` - IR drop, sneak paths
- `module2-crossbar/pkg/crossbar/cell.go` - FeFET/FTJ cell model
- `module2-crossbar/pkg/crossbar/enhanced.go` - Integrated simulation

### Key Functions

```go
// Basic MVM
array.MVM(input []float64) ([]float64, error)

// MVM with all non-idealities
array.MVMWithNonIdealities(input, options) (*MVMResult, error)

// Program weights
array.ProgramWeight(row, col int, weight float64) error
array.ProgramMatrix(weights [][]float64) error
```

---

## Related Documentation

- **[Crossbar Demo Guide](../educational/crossbar.demo.md)** - How to run the visualization
- **[Crossbar ELI5](crossbar.ELI5.md)** - Simple analogies and explanations
- **[Research Papers](../educational/crossbar.research.md)** - Academic references
- **[Open Source Context](../crossbar/educational/crossbar.opensource.md)** - Relationship to other tools

---

## References

See [../educational/crossbar.research.md](../educational/crossbar.research.md) for complete academic citations.

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
**Source:** Archival conference reference (not validated)
**License:** See project root for details
