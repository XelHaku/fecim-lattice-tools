# Module 1: Hysteresis - The Memory Cell's P-E Curve

> **Explain Like I'm 5:** The ferroelectric material remembers which way you pushed it!

---

## The Core Problem: How Does a Memory Cell Remember?

Regular materials behave like a spring:
```
Push the spring → it compresses
Stop pushing → it springs back to normal
No memory!
```

Ferroelectric materials (like HZO) behave like a stubborn kid:
```
Push left → stays pushed left!
Push right → stays pushed right!
(doesn't naturally bounce back)

That's how it remembers! 🧠
```

---

## What is Polarization?

Inside materials, positive (+) and negative (-) charges can separate:

```
Before:                After pushing:
  ⊕⊖                     ⊕────⊖
(together)            (separated = polarized)
```

**Polarization (P)** = how much the charges are separated.

---

## The Hysteresis Loop (The Main Idea)

When you push and release, the polarization traces a loop:

```
        Polarization (P)
              ↑
         Ps ──┼────────╮      Ps = saturation
              │        │           (maximum)
         Pr ──┼─╮      │      Pr = remanent
              │ │      │           (stays when V=0)
    ──────────┼─●──────●────→ Voltage (V)
              │       │ │
        -Pr ──┼───────╯ │
              │         │
        -Ps ──┼─────────╯
              │
            -Ec  0  +Ec

        Ec = coercive field
            (voltage needed to flip)
```

**The key insight:** Going UP is NOT the same as going DOWN! The material remembers where it came from.

---

## Why This Matters for Memory

```
Normal memory: ON or OFF (1 bit)
              ░░░░░░░░░░
              bit = 0 or 1

Ferroelectric: Stop at different points on the loop!
              ░░░░░░░░░░
              State 0 ●
              State 1 ●
              State 2 ●
              ... (up to 30 states in demo baseline)

More states per cell = store more information!
```

---

## The Three Key Numbers for HZO

| Parameter | Symbol | Value | Why It Matters |
|-----------|--------|-------|----------------|
| **Saturation Polarization** | Ps | 25 μC/cm² | Maximum charge separation (memory capacity) |
| **Coercive Field** | Ec | 1.0 MV/cm | Voltage to flip the polarization (write effort) |
| **Remanent** | Pr | ~15 μC/cm² | What stays after you stop pushing (what you read) |

---

## The Preisach Model (How We Simulate It)

Real materials have trillions of atoms. We can't simulate each one!

Instead, imagine the material is made of millions of tiny switches called **hysterons**:

```
One Hysteron (a tiny sticky switch):

Output (+1 or -1)
    (+1) ├───────────────────╮
         │         α         │
     (0) ├───────────────────┼──→ Input Voltage
         │         β         │
    (-1) ├───────────────────╯

α = voltage to turn ON
β = voltage to turn OFF
(they're different!)
```

Each hysteron:
- Turns ON at voltage α (and stays ON)
- Turns OFF at voltage β (and stays OFF)
- Ignores the input between β and α

The whole hysteresis loop comes from adding up millions of hysterons!

---

## Our Simplified Version: Tanh Function

Instead of millions of hysterons, we use a smooth formula:

```go
P = Ps × tanh((V - Ec) / δ)
```

This gives us:
- A smooth S-shaped curve
- Similar shape to real data
- Fast to compute
- Easy to understand

---

## What the Demo Shows You

When you run Module 1, you see:

1. **Interactive P-E Curve**: Drag voltage up/down, watch polarization follow
2. **Material Selector**: Default HZO, Optimized, Ferroelectric CIM
3. **Waveform Modes**:
   - Sine: Smooth oscillation
   - Triangle: Linear up/down
   - Square: Snap to extremes
   - Manual: You control it
4. **30 Discrete Levels**: Visualization of multi-level storage (demo baseline)

---

## How to Run Module 1

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools hysteresis
```

---

## Key Takeaways

| Concept | Remember This |
|---------|---------------|
| **Polarization** | How separated the charges are (storage amount) |
| **Hysteresis** | The loop shape shows memory (up ≠ down) |
| **Coercive Field** | Voltage needed to flip (write effort) |
| **30 Levels** | Demo baseline for discrete storage states |
| **Preisach Model** | Simulates hysteresis using tiny hysterons |

---

## Next Steps

- **Want to see it compute?** Go to [Module 2: Crossbar](../module2-crossbar/eli5.md)
- **Want to recognize digits?** Go to [Module 3: MNIST](../module3-mnist/eli5.md)
- **Back to overview?** See [ELI5 Overview](../eli5-overview.md)
