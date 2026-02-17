# Learn the Technology

**Educational content for Ferroelectric Compute-in-Memory - from basics to advanced.**

---

## 🎯 Where to Start

### Never Heard of FeCIM?
→ Read [eli5-overview.md](eli5-overview.md) - Everything explained simply in 20 minutes.

### Want to Follow a Curriculum?
→ See [Sequential Learning Path](#sequential-learning-path) below - Module 1 through 7 in order.

### Looking for a Specific Topic?
→ See [Learn by Topic](#learn-by-topic) or use the module index below.

---

## 📚 Module Index

| Module | Topic | ELI5 | Physics | Features | Tools |
|--------|-------|-------|---------|----------|-------|
| **[1](module1-hysteresis/)** | Hysteresis & Materials | [eli5](module1-hysteresis/eli5.md) | [physics](module1-hysteresis/physics.md) | [features](module1-hysteresis/features.md) | [tools](module1-hysteresis/tools.md) |
| **[2](module2-crossbar/)** | Crossbar Arrays | [eli5](module2-crossbar/eli5.md) | [physics](module2-crossbar/physics.md) | [features](module2-crossbar/features.md) | [tools](module2-crossbar/tools.md) |
| **[3](module3-mnist/)** | Neural Networks | [eli5](module3-mnist/eli5.md) | [physics](module3-mnist/physics.md) | [features](module3-mnist/features.md) | [tools](module3-mnist/tools.md) |
| **[4](module4-circuits/)** | Peripheral Circuits | [eli5](module4-circuits/eli5.md) | [physics](module4-circuits/physics.md) | [features](module4-circuits/features.md) | [tools](module4-circuits/tools.md) |
| **[5](module5-comparison/)** | Technology Comparison | [eli5](module5-comparison/eli5.md) | [physics](module5-comparison/physics.md) | [features](module5-comparison/features.md) | [tools](module5-comparison/tools.md) |
| **[6](module6-eda/)** | EDA & Chip Design | [eli5](module6-eda/eli5.md) | [physics](module6-eda/physics.md) | [features](module6-eda/features.md) | [tools](module6-eda/tools.md) |
| **[7](module7-docs/)** | Documentation Tools | [eli5](module7-docs/eli5.md) | [physics](module7-docs/physics.md) | [features](module7-docs/features.md) | [tools](module7-docs/tools.md) |

---

## 🗺️ Sequential Learning Path

Follow this order to build a complete understanding of FeCIM. Each module builds on the previous.

### Module 1: Hysteresis & Materials
**The Memory Cell**

- What is polarization and why it matters
- P-E curves (hysteresis loops)
- How materials remember states
- 8 ferroelectric materials compared
- The 30-level baseline and quantization

**Key Insight:** The material physically moves to remember!

**Start Here:** [module1-hysteresis/README.md](module1-hysteresis/README.md)

**Time:** 30-45 minutes

---

### Module 2: Crossbar Arrays
**Compute-in-Memory**

- Ohm's Law as multiplication (I = V × G)
- Crossbar array architecture (1T1R, 0T1R, 2T1R)
- Matrix-vector multiplication in hardware
- Non-ideal effects: IR drop, sneak paths, drift
- Parasitic solving and compensation

**Key Insight:** Physics does matrix multiplication for free!

**Start Here:** [module2-crossbar/README.md](module2-crossbar/README.md)

**Time:** 45-60 minutes

---

### Module 3: Neural Networks
**AI Recognition**

- What is a neural network?
- MNIST handwriting dataset
- Complete 5-step inference flow
- FP32 mode vs CIM analog mode comparison
- ReLU activation, softmax probabilities

**Key Insight:** Two crossbar layers can recognize handwritten digits!

**Start Here:** [module3-mnist/README.md](module3-mnist/README.md)

**Time:** 30-45 minutes

---

### Module 4: Peripheral Circuits
**Supporting Cast**

- DAC: digital-to-analog conversion for write
- TIA: transimpedance amplifier for readout
- ADC: analog-to-digital conversion for output
- Charge pump: voltage boosting for write
- 4-phase write/read timing cycle

**Key Insight:** Tiny analog currents need support circuits to be useful!

**Start Here:** [module4-circuits/README.md](module4-circuits/README.md)

**Time:** 30-45 minutes

---

### Module 5: Technology Comparison
**Why FeCIM Matters**

- Von Neumann bottleneck explained
- CPU vs GPU vs FeCIM trade-offs
- Other CIM technologies: ReRAM, PCM, MRAM
- Energy, speed, and density comparisons
- Limitations and open challenges

**Key Insight:** FeCIM could be 1000× more efficient (projected, unverified).

**Start Here:** [module5-comparison/README.md](module5-comparison/README.md)

**Time:** 20-30 minutes

---

### Module 6: EDA Tools
**Chip Design**

- Electronic Design Automation (EDA) overview
- RTL-to-GDSII flow
- Layout, routing, timing, power analysis
- Design rule checking (DRC)
- SKY130 and GF180 process design kits
- Integration with OpenLane

**Key Insight:** Real chip design requires millions of automated decisions!

**Start Here:** [module6-eda/README.md](module6-eda/README.md)

**Time:** 30-45 minutes

---

### Module 7: Documentation Tools
**Knowledge Sharing**

- Glossary viewer (100+ terms)
- Full-text search across docs
- Markdown rendering
- Cross-reference navigation
- How to contribute documentation

**Start Here:** [module7-docs/README.md](module7-docs/README.md)

**Time:** 10-15 minutes

---

## 📖 Learn by Topic

### Material Science

| Resource | Level | Description |
|----------|-------|-------------|
| [eli5-overview.md](eli5-overview.md) | Beginner | Full technology overview |
| [module1-hysteresis/eli5.md](module1-hysteresis/eli5.md) | Beginner | Hysteresis explained simply |
| [module1-hysteresis/physics.md](module1-hysteresis/physics.md) | Advanced | Landau-Khalatnikov equations |
| [module1-hysteresis/materials.md](module1-hysteresis/materials.md) | Intermediate | All 8 materials compared |

### Compute-in-Memory

| Resource | Level | Description |
|----------|-------|-------------|
| [module2-crossbar/eli5.md](module2-crossbar/eli5.md) | Beginner | Crossbar explained simply |
| [module2-crossbar/physics.md](module2-crossbar/physics.md) | Advanced | IR drop, sneak path math |
| [module2-crossbar/architecture.md](module2-crossbar/architecture.md) | Intermediate | 1T1R vs 0T1R vs 2T1R |

### Neural Networks

| Resource | Level | Description |
|----------|-------|-------------|
| [module3-mnist/eli5.md](module3-mnist/eli5.md) | Beginner | Neural networks explained simply |
| [module3-mnist/physics.md](module3-mnist/physics.md) | Advanced | CIM inference physics |
| [module3-mnist/features.md](module3-mnist/features.md) | Intermediate | FP32 vs CIM mode comparison |

### Circuit Design

| Resource | Level | Description |
|----------|-------|-------------|
| [module4-circuits/eli5.md](module4-circuits/eli5.md) | Beginner | DAC/ADC/TIA explained simply |
| [module4-circuits/physics.md](module4-circuits/physics.md) | Advanced | Circuit non-idealities |
| [module4-circuits/features.md](module4-circuits/features.md) | Intermediate | All peripheral models |

### Technology Comparison

| Resource | Level | Description |
|----------|-------|-------------|
| [module5-comparison/eli5.md](module5-comparison/eli5.md) | Beginner | Why FeCIM matters |
| [module5-comparison/physics.md](module5-comparison/physics.md) | Advanced | Energy/speed analysis |

### EDA & Manufacturing

| Resource | Level | Description |
|----------|-------|-------------|
| [module6-eda/eli5.md](module6-eda/eli5.md) | Beginner | Chip design explained simply |
| [module6-eda/physics.md](module6-eda/physics.md) | Advanced | VLSI physical design |

---

## 🎓 Learn by Level

### Beginner

Start here if you have no background in electronics or materials science:

1. [eli5-overview.md](eli5-overview.md) - 60-second pitch and story arc
2. [module1-hysteresis/eli5.md](module1-hysteresis/eli5.md) - Memory cell in simple terms
3. [module2-crossbar/eli5.md](module2-crossbar/eli5.md) - How math happens in hardware
4. [module3-mnist/eli5.md](module3-mnist/eli5.md) - Neural networks in simple terms

**Goal:** Understand what FeCIM is and why it matters.

### Intermediate

Recommended if you have undergraduate-level physics or engineering background:

1. [module1-hysteresis/features.md](module1-hysteresis/features.md) - Preisach model and calibration
2. [module2-crossbar/architecture.md](module2-crossbar/architecture.md) - Array architectures
3. [module4-circuits/features.md](module4-circuits/features.md) - DAC/ADC non-idealities
4. [module5-comparison/features.md](module5-comparison/features.md) - Quantitative comparisons

**Goal:** Build and modify the simulation tools.

### Advanced

Recommended if you have graduate-level materials or circuit background:

1. [module1-hysteresis/physics.md](module1-hysteresis/physics.md) - Landau-Khalatnikov dynamics
2. [module2-crossbar/physics.md](module2-crossbar/physics.md) - Parasitic solver mathematics
3. [module4-circuits/physics.md](module4-circuits/physics.md) - Circuit physics models
4. [../4-research/physics-validation.md](../4-research/physics-validation.md) - Validation methodology

**Goal:** Evaluate scientific accuracy and contribute physics improvements.

---

## 🔑 Key Concepts Summary

### The Core Idea

```
Traditional computing:
  Memory ←(data bus)→ Processor
  Energy wasted = 90% on data movement

FeCIM:
  Memory = Processor
  Energy wasted = ~0% on movement (math happens in-place)
```

### Essential Terms

| Term | Simple Meaning |
|------|----------------|
| **FeCIM** | Ferroelectric Compute-in-Memory |
| **HZO** | The special memory material (Hafnium-Zirconium-Oxide) |
| **Hysteresis** | Memory property: going up ≠ going down |
| **Polarization** | How much charge is stored |
| **Crossbar** | Grid of wires where math happens |
| **MVM** | Matrix-Vector Multiplication (AI's core operation) |
| **CIM** | Compute-in-Memory |
| **DAC/ADC** | Digital↔Analog converters |
| **TIA** | Current-to-voltage amplifier |

Full glossary: [../GLOSSARY.md](../GLOSSARY.md)

---

## 🎬 Recommended Learning Sequence

### Option A: 2-Hour Session (Comprehensive)

```
Hour 1:
  0:00 - Read eli5-overview.md (20 min)
  0:20 - Run Module 1 demo + read eli5 (25 min)
  0:45 - Run Module 2 demo + read eli5 (15 min)

Hour 2:
  1:00 - Run Module 3 demo + read eli5 (20 min)
  1:20 - Read Module 4 eli5 (15 min)
  1:35 - Read Module 5 eli5 (15 min)
  1:50 - Explore on your own (10 min)
```

### Option B: 30-Minute Session (Quick Overview)

```
  0:00 - Read eli5-overview.md (10 min)
  0:10 - Run Module 1 demo (5 min)
  0:15 - Run Module 2 demo (5 min)
  0:20 - Run Module 3 demo (5 min)
  0:25 - Skim Module 5 comparison (5 min)
```

### Option C: Self-Paced (Deep Dive)

- Week 1: Modules 1-2 (material science + computation)
- Week 2: Modules 3-4 (neural networks + circuits)
- Week 3: Modules 5-6 (comparison + design)
- Week 4: Research papers in [../4-research/](../4-research/)

---

## ⚠️ Important: What Is and Isn't Simulated

This is an **educational simulator**. Before drawing conclusions:

### Simulation Baselines (Not Verified Hardware)

- **30 analog states** per cell - configurable default, not measured hardware
- **Energy projections** - physics-based estimates, pending silicon validation
- **Performance numbers** - illustrative, not benchmarked devices

### Verified Claims (From Literature)

- **98.24% MNIST accuracy** - HZO FTJ reservoir computing (J. Alloys & Compounds 2025)
- **885 TOPS/W** - Multi-level FeFET crossbar (Nature Comms 2023)

Full details: [../4-research/honesty-audit.md](../4-research/honesty-audit.md)

---

## 🔗 Quick Links

**Start Learning:**
- [ELI5 Overview](eli5-overview.md) - Start here!
- [Module 1](module1-hysteresis/README.md) - First module

**Go Deeper:**
- [Research Papers](../4-research/papers/) - 230+ papers
- [Physics Validation](../4-research/physics-validation.md) - Accuracy report
- [API Reference](../3-develop/api-reference.md) - Code docs

**Reference:**
- [Glossary](../GLOSSARY.md) - All terms
- [Honesty Audit](../4-research/honesty-audit.md) - Claims status

---

**Last Updated:** 2026-02-16
**Status:** All 7 modules complete with interactive demos
