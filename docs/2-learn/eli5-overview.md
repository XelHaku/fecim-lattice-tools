# Ferroelectric CIM: Complete ELI5 Guide - Overview

> **Explain Like I'm 5:** Everything you need to know about Ferroelectric Compute-in-Memory!

---

## The 60-Second Summary

**The Problem:** AI is eating the world, but it's also eating all the electricity. Computers waste 90% of energy moving data around.

**The Root Cause:** Memory (where data lives) and processors (where math happens) are separated. Every calculation requires billions of trips back and forth.

**The Solution:** What if memory could also do math? **Ferroelectric CIM does math where the data already lives.**

Using a special material called HZO (Hafnium-Zirconium-Oxide), we build memory cells that can also compute. When you apply a voltage, the current that flows out IS the multiplication result. Physics does the math for free!

---

## The Story in 8 Chapters

This guide is organized into 6 modules plus an overview. Each module builds on the last:

### [Module 1: Hysteresis](./module1-hysteresis/eli5.md) - The Memory Cell

**Focus:** How ferroelectric materials remember

- What is polarization (charge separation)?
- Hysteresis loops (going up ≠ going down)
- The P-E curve and 30-level baseline
- Preisach model (simulating hysteresis)

**Demo:** Interactive P-E curve visualization

**Key Insight:** The material physically moves to remember!

---

### [Module 2: Crossbar](./module2-crossbar/eli5.md) - Compute-in-Memory

**Focus:** How a grid of cells does matrix multiplication

- Ohm's Law is multiplication (I = V × G)
- Crossbar array architecture
- Matrix-vector multiplication happens in parallel
- Real-world problems: IR drop, sneak paths, variation

**Demo:** Interactive crossbar with IR drop and sneak path analysis

**Key Insight:** Physics does multiplication for free!

---

### [Module 3: MNIST](./module3-mnist/eli5.md) - AI Recognition

**Focus:** Neural networks and handwritten digit recognition

- What is a neural network?
- Complete inference flow (5 steps)
- ReLU activation (adds non-linearity)
- Softmax (probabilities)

**Demo:** Draw a digit, watch the network recognize it

**Key Insight:** Two crossbar layers can recognize handwritten digits!

---

### [Module 4: Circuits](./module4-circuits/eli5.md) - Supporting Cast

**Focus:** The peripheral circuits that make it work

- DAC (Digital-to-Analog Converter)
- Charge Pump (voltage booster)
- TIA (Transimpedance Amplifier)
- ADC (Analog-to-Digital Converter)

**Demo:** Circuit simulations showing DAC linearity, ADC resolution, timing

**Key Insight:** Tiny currents need friends to become readable signals!

---

### [Module 5: Comparison](./module5-comparison/eli5.md) - Why It Matters

**Focus:** Comparing Ferroelectric CIM to other technologies

- Traditional CPU+DRAM bottleneck
- GPU advantages and limitations
- Other compute-in-memory: ReRAM, PCM, MRAM
- Energy/speed/density comparisons

**Demo:** Technology comparison across multiple workloads

**Key Insight:** Ferroelectric CIM could be 1000× more efficient!

---

### [Module 6: EDA](./module6-eda/eli5.md) - Design Tools

**Focus:** How chips are designed and manufactured

- EDA (Electronic Design Automation)
- Layout, routing, timing, power analysis
- Design rule checking
- Yield simulation

**Demo:** Layout viewer, DRC checker, timing analyzer

**Key Insight:** Making chips requires millions of design decisions!

---

## Quick Navigation

### By Learning Level

**Beginner (Never heard of this):**
1. Start with [Module 1: Hysteresis](./module1-hysteresis/eli5.md)
2. Watch the interactive demo
3. Read the key takeaways

**Intermediate (Understand the basics):**
1. Read [Module 2: Crossbar](./module2-crossbar/eli5.md)
2. Understand Ohm's Law and matrix multiplication
3. Try the crossbar demo

**Advanced (Ready for the full picture):**
1. Read all modules in order
2. Run all demos
3. Study the comparison metrics
4. Explore the full ELI5.md for mathematical details

---

### By Question

**"What is Ferroelectric CIM?"**
→ [Module 1: Hysteresis](./module1-hysteresis/eli5.md) + [Module 2: Crossbar](./module2-crossbar/eli5.md)

**"How does it recognize images?"**
→ [Module 3: MNIST](./module3-mnist/eli5.md)

**"What circuits does it need?"**
→ [Module 4: Circuits](./module4-circuits/eli5.md)

**"Is it better than GPUs?"**
→ [Module 5: Comparison](./module5-comparison/eli5.md)

**"How would engineers design it?"**
→ [Module 6: EDA](./module6-eda/eli5.md)

**"I want all the technical details"**
→ [Complete ELI5.md](../../ELI5.md) (2100+ lines)

---

## Running the Demos

All demos are interactive and runnable. You need:
- Go 1.21+
- Fyne GUI library (for modules 1-3)
- Standard libraries (for modules 4-6)

### Quick Start

```bash
# Build the main executable
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools

# Run Module 1 (Hysteresis)
./fecim-lattice-tools hysteresis

# Run Module 2 (Crossbar)
./fecim-lattice-tools crossbar

# Run Module 3 (MNIST)
./fecim-lattice-tools mnist

# Run Module 4 (Circuits CLI)
go run ./cmd/fecim-lattice-tools circuits cli --all

# Run Module 5 (Comparison CLI)
go run ./cmd/fecim-lattice-tools comparison cli --all

# Run Module 6 (EDA CLI)
go run ./cmd/fecim-lattice-tools eda layout --view
```

---

## The Story Arc

```
Module 1
The memory cell remembers
     │
     ▼
Module 2
The array computes
     │
     ▼
Module 3
The network learns
     │
     ▼
Module 4
The circuits support it
     │
     ▼
Module 5
It beats everything else
     │
     ▼
Module 6
Engineers can design it
     │
     ▼
The future: Ferroelectric CIM everywhere!
```

---

## Key Concepts (All Modules)

| Concept | Appears In | Simple Definition |
|---------|-----------|-------------------|
| **Polarization (P)** | Module 1 | How separated charges are (storage amount) |
| **Hysteresis** | Module 1 | Memory (up ≠ down) |
| **Coercive Field (Ec)** | Module 1 | Voltage to flip polarization |
| **Crossbar** | Module 2 | Grid where math happens |
| **Matrix Multiplication** | Module 2, 3 | Core AI math (all at once!) |
| **Ohm's Law** | Module 2 | I = V × G (physics does multiplication) |
| **IR Drop** | Module 2 | Voltage weakens along wires |
| **Sneak Paths** | Module 2 | Current leaks through wrong cells |
| **Neural Network** | Module 3 | Layers of neurons doing math |
| **ReLU** | Module 3 | Activation function (non-linearity) |
| **Softmax** | Module 3 | Convert scores to probabilities |
| **DAC** | Module 4 | Digital → Analog converter |
| **ADC** | Module 4 | Analog → Digital converter |
| **TIA** | Module 4 | Current → Voltage amplifier |
| **Charge Pump** | Module 4 | Voltage booster |
| **Energy Efficiency** | Module 5 | 1000× better than CPU (projected) |
| **Latency** | Module 5 | 1000× faster than CPU (projected) |
| **EDA** | Module 6 | Software for chip design |

---

## What You'll Learn

By the end of all 6 modules, you'll understand:

1. ✓ How ferroelectric materials store information
2. ✓ How crossbars do matrix multiplication
3. ✓ How neural networks recognize patterns
4. ✓ How support circuits enable analog computing
5. ✓ Why Ferroelectric CIM could revolutionize AI
6. ✓ How engineers design chips with EDA tools

---

## Important Disclaimers

This simulator is **educational only**:

- **No hardware claims:** The simulator models concepts but does not validate hardware performance
- **Simulation baseline:** 30-level quantization is configurable, not a hardware measurement
- **Energy projections:** Illustrative values based on physics principles, not measured devices
- **Performance claims:** Comparisons are theoretical, pending real silicon validation

Real Ferroelectric CIM devices will require:
- Prototype fabrication
- Characterization testing
- Integration validation
- System-level benchmarking

This project helps engineers and students **explore the design space**. Scientists will validate (or refute) the claims!

---

## Citation Note

Numeric/performance claims without DOI citations are illustrative. Before external use, add proper citations from peer-reviewed papers.

See `docs/4-research/honesty-audit.md` for detailed audit of claims.

---

## Next Steps

### Ready to Learn?

1. **Start with Module 1:** [Hysteresis](./module1-hysteresis/eli5.md)
2. **Follow the story arc** through all 6 modules
3. **Run the demos** (interactive learning!)
4. **Explore the full guide** for technical depth

### Want Specific Topics?

- **Material Science:** Module 1 + full ELI5.md Part 7
- **Circuit Design:** Module 4 + full ELI5.md Part 9
- **AI/ML Basics:** Module 3 + full ELI5.md Part 8
- **Energy Analysis:** Module 5 + full ELI5.md Part 21
- **Chip Design:** Module 6 + full ELI5.md Part 23

### Want to Contribute?

Check `CONTRIBUTING.md` and `docs/development/SCRIPT_REFERENCE.md` for:
- Code contribution guidelines
- Module improvement ideas
- Demo enhancement opportunities
- Documentation needs

---

## The Big Picture

```
Today (2026):
  AI is transforming everything, but:
  - Too much energy (data centers use 8% of global power)
  - Too slow (moving data dominates processing time)
  - Too expensive (electricity costs everything)

The Problem:
  Memory and processors are separate (von Neumann architecture)
  Result: 90% of energy wasted moving data

The Vision:
  Compute where memory lives (Ferroelectric CIM)
  Result: Potential energy savings, faster inference

The Path Forward:
  - Research prototypes (2020-2025)
  - Integration challenges (2025-2030)
  - Commercial products (2030+)

This simulator helps imagine the destination and plan the journey!
```

---

## Glossary (Quick Reference)

| Term | Simple Meaning |
|------|----------------|
| **Ferroelectric** | Material that remembers how you pushed it |
| **Polarization** | How separated the charges are |
| **Hysteresis** | Going up ≠ going down (history matters) |
| **Crossbar** | Grid of wires with memory at intersections |
| **CIM** | Compute-in-Memory (math where data lives) |
| **MVM** | Matrix-Vector Multiplication |
| **HZO** | Hafnium-Zirconium-Oxide (the magic material) |
| **DAC/ADC** | Digital/Analog converters |
| **TIA** | Transimpedance Amplifier |
| **MNIST** | Handwritten digit recognition test |
| **ReLU** | Activation function (non-linearity) |
| **Softmax** | Probability converter |
| **EDA** | Electronic Design Automation (chip design tools) |
| **DRC** | Design Rule Check (verification) |
| **Yield** | % of chips that work |

---

## Resources

### In This Repository

- **Full Technical Guide:** [docs/ELI5.md](../../ELI5.md) (2100+ lines)
- **Project Status:** [docs/project/STATUS.md](../../project/STATUS.md)
- **Accuracy Audit:** [docs/4-research/honesty-audit.md](../../4-research/honesty-audit.md)
- **Development Guide:** [docs/development/SCRIPT_REFERENCE.md](../../development/SCRIPT_REFERENCE.md)

### External Resources

- **3Blue1Brown:** Neural Networks Playlist (visual, beautiful)
- **IBM:** In-Memory Computing Overview
- **Dr. external research group:** external research institution research

---

## Questions?

Each module has:
- Simple explanations
- Diagrams and examples
- Interactive demos
- Key takeaways
- Links to related modules

If you get stuck:
1. Re-read the "Key Insight" box
2. Try the demo (hands-on learning!)
3. Check the glossary
4. Read the full ELI5.md for more detail

---

## The Dream

```
After reading this guide, anyone can:
  ✓ Understand how ferroelectric compute works
  ✓ See why it matters for AI
  ✓ Imagine the energy savings
  ✓ Appreciate the engineering challenges
  ✓ Contribute to the research

No PhD required!
```

---

## Let's Go!

Pick a module and dive in:

- **[Module 1: Hysteresis](./module1-hysteresis/eli5.md)** ← Start here!
- [Module 2: Crossbar](./module2-crossbar/eli5.md)
- [Module 3: MNIST](./module3-mnist/eli5.md)
- [Module 4: Circuits](./module4-circuits/eli5.md)
- [Module 5: Comparison](./module5-comparison/eli5.md)
- [Module 6: EDA](./module6-eda/eli5.md)

Happy learning! 🎉

---

**Last updated:** 2026-02-16
**Status:** All 6 modules complete with demos
**Next:** Real silicon will tell the true story!
