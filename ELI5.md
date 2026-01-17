# IronLattice: The Complete ELI5 Guide

**Goal:** After reading this, a 5-year-old could become the lead engineer.

---

# Part 1: The Very Basics

## What is Electricity?

Everything is made of tiny balls called **atoms**. Inside atoms are:
- **Protons** (+) - live in the middle, don't move much
- **Electrons** (-) - zoom around the outside, love to travel

When electrons flow from one place to another, that's **electricity**! Like water flowing through a pipe.

```
Battery ──────────────────→ Light Bulb
        electrons flowing
```

## What is Voltage?

**Voltage** is like water pressure. Higher voltage = more push for the electrons.

```
Low Voltage:        High Voltage:
   💧                  💧💧💧
   drip drip           WHOOOOSH!
```

**Units:** Volts (V). Your phone uses ~5V. A power outlet uses ~120V.

## What is a Computer?

A computer is a machine that:
1. **Stores** information (memory)
2. **Does math** (processor)
3. **Shows you the answer** (screen)

All information is stored as **1s and 0s** (binary). Like lots of tiny light switches:
- 0 = OFF
- 1 = ON

---

# Part 2: Why Current Computers Are Bad at AI

## The Problem: The Commute

Regular computers have two parts that don't live together:

```
┌─────────────┐                    ┌─────────────┐
│   MEMORY    │                    │  PROCESSOR  │
│  (storage)  │ ←─── long road ──→ │   (brain)   │
│             │                    │             │
│ "I keep     │                    │ "I do the   │
│  the data"  │                    │  thinking"  │
└─────────────┘                    └─────────────┘
```

Every time the computer thinks, it has to:
1. Walk to memory
2. Grab some data
3. Walk back to processor
4. Do math
5. Walk back to memory
6. Store the answer
7. Repeat millions of times!

**This "commute" wastes 90% of the energy and makes everything slow.**

This is called the **von Neumann bottleneck** (named after a smart person who designed computers this way a long time ago).

## Why AI Makes It Worse

AI does A LOT of math. Specifically, it multiplies big tables of numbers together. Like this:

```
Input (picture):       Weights (learned):       Output:
   [1, 0, 1]        ×    [0.5, 0.2, 0.8]     =  [answer!]
   [0, 1, 0]        ×    [0.1, 0.9, 0.3]     
   [1, 1, 0]        ×    [0.7, 0.4, 0.6]     
```

For one AI to recognize a cat in a picture, it might do **billions** of these multiplications. Each one requires walking to memory and back!

---

# Part 3: The IronLattice Solution

## Compute-in-Memory: Think Where You Store

What if memory could also do math? No walking needed!

```
┌─────────────────────────────────┐
│                                 │
│     MEMORY + PROCESSOR          │
│           TOGETHER!             │
│                                 │
│  "I store AND think!"           │
│                                 │
└─────────────────────────────────┘

Walking distance: ZERO! 🎉
```

**Result:**
- 10,000,000× less energy
- 1,000,000× faster
- Smaller chips

## How Can Memory Do Math?

Remember how AI multiplies tables? Here's the magic:

**Ohm's Law** (discovered in 1827): 
```
Current = Voltage × Conductance
   I    =    V    ×      G

This is just... multiplication! Physics does it for free!
```

If we:
1. Store the "weights" as how conductive each memory cell is (G)
2. Send in the "input" as voltage (V)
3. The current that comes out (I) IS the multiplication result!

**Physics does the math at the speed of light. No instructions needed!**

---

# Part 4: The Crossbar Array (Demo 2)

## What is a Crossbar?

It's a grid of wires with a memory cell at each crossing:

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

## How Does It Work?

1. **Each memory cell stores a weight (G)** - how much it conducts
2. **You apply input voltages to columns (V)**
3. **Current flows through each cell: I = G × V** (multiplication!)
4. **All currents on a row add up** (addition!)

```
Row output = G₀₀×V₀ + G₀₁×V₁ + G₀₂×V₂ + G₀₃×V₃

This is exactly matrix-vector multiplication!
All 16 multiplications happen AT THE SAME TIME!
```

## Why This is Amazing

| Method | Operations | Time |
|--------|-----------|------|
| Regular CPU | One multiply at a time | 1 million steps |
| Crossbar | ALL multiplies at once | ~1 step! |

For a 1000×1000 matrix: Regular CPU needs 1,000,000 operations. Crossbar does it in ONE analog operation!

## What Can Go Wrong (Non-Idealities)

Real crossbars aren't perfect:

### 1. IR Drop (Voltage gets weak)
The wires have some resistance. Voltage gets lower as it travels:
```
Sent: 1.0V → 0.95V → 0.90V → 0.85V
                ↓ gets weaker!
```

### 2. Sneak Paths (Current takes shortcuts)
Current is lazy and takes all possible paths:
```
Want:  →→→●→→→ (through target cell only)
Got:   →→→●→→→
        ↓   ↑
       →→→●→→→ (snuck through other cells!)
```

### 3. Variation (Each cell is a little different)
Factories aren't perfect. Two cells set to "0.5" might actually be:
- Cell A: 0.48
- Cell B: 0.52

---

# Part 5: Ferroelectric Materials (The Magic Crystal)

## What is Polarization?

Inside materials, positive (+) and negative (-) charges can separate:

```
Before:                After pushing:
  ⊕⊖                     ⊕────⊖
(together)            (separated = polarized)
```

**Polarization (P)** = how much the charges are separated.

## Normal Materials vs. Ferroelectric

**Normal material (like glass):**
```
Push charges → they separate
Stop pushing → they go back together
No memory!
```

**Ferroelectric material (like HZO):**
```
Push charges → they separate
Stop pushing → they STAY separated!
MEMORY! 🧠
```

## Why Do They Stay?

The crystal structure actually **shifts** to a new stable position:

```
Before:                After (new stable position):
┌───┬───┬───┐          ┌───┬───┬───┐
│   │ ● │   │          │   │   │   │
│ ● │   │ ● │  →push→  │ ● │ ● │ ● │ ← center atom
│   │ ● │   │          │   │   │   │   moved UP
└───┴───┴───┘          └───┴───┴───┘

The atom PHYSICALLY moved to a new home!
```

## The Hysteresis Loop (Demo 1)

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

**Key insight:** Going up is NOT the same as going down! The material remembers where it came from.

## 30 Analog States

By stopping at different points, HZO can store 30 different levels:

```
Polarization
     ↑
  +Ps├─ State 30 ●
     ├─ State 29 ●
     ├─ State 28 ●
     ⋮
     ├─ State 16 ●
   0 ├─ State 15 ●
     ├─ State 14 ●
     ⋮
     ├─ State 2  ●
     ├─ State 1  ●
  -Ps├─ State 0  ●
```

Regular memory: 1 bit (ON/OFF)
IronLattice: 5 bits (30 states ≈ 2⁵)

---

# Part 6: The Preisach Model (How We Simulate It)

## The Problem

Simulating trillions of atoms is impossible. We need a simpler model!

## The Idea: Hysterons

Imagine the material is made of millions of tiny switches called **hysterons**:

```
One Hysteron (a tiny switch):

Output (+1 or -1)
    (+1) ├───────────────────╮
         │         α         │
     (0) ├───────────────────┼──→ Input
         │         β         │
    (-1) ├───────────────────╯

α = voltage to turn ON
β = voltage to turn OFF
(they're different!)
```

Each hysteron is like a sticky light switch that turns ON at one voltage and OFF at a different one.

## Many Hysterons = Complete Model

Real material = millions of hysterons with different (α, β) values:

```
The whole hysteresis loop comes from adding up
millions of tiny hysterons!

Big loop = [h₁] + [h₂] + [h₃] + ... millions
           α₁β₁   α₂β₂   α₃β₃
```

## Our Simplified Version

Instead of simulating millions of hysterons, we use a **hyperbolic tangent** function:

```python
P = Ps × tanh((V - Ec) / δ)
```

This gives us a smooth S-shaped curve that looks like real data!

---

# Part 7: Phase-Field Simulation (Demo 3)

## What are Domains?

The crystal doesn't all point the same way. It breaks into **domains**:

```
┌────────────────────────────────────┐
│▓▓▓▓▓▓▓│░░░░░░░│▓▓▓▓▓│░░░░░░░░░░░░│
│▓▓▓▓▓▓▓│░░░░░░░│▓▓▓▓▓│░░░░░░░░░░░░│
│▓▓▓▓▓▓▓│░░░░░░░│▓▓▓▓▓│░░░░░░░░░░░░│
└────────────────────────────────────┘
    ▓ = polarization UP
    ░ = polarization DOWN
    │ = domain wall (boundary)
```

## Why Domains Exist

The material minimizes its total energy. Having domains reduces the energy at the surfaces.

## The TDGL Equation

**Time-Dependent Ginzburg-Landau** tells us how domains evolve:

```
∂P/∂t = -L × (energy gradient)

In English:
"Polarization changes to reduce total energy,
 at a speed controlled by L"
```

The energy has three parts:
1. **Landau energy:** Prefers P = ±Ps (the two stable states)
2. **Gradient energy:** Penalizes sharp domain walls
3. **Electric energy:** External field pushes P one way

## GPU Simulation

We divide the crystal into a 3D grid and update each point:

```
For each point (x, y, z):
    1. Calculate energy gradient
    2. Update polarization
    3. Repeat

128 × 128 × 128 = 2 million points
GPU does them ALL IN PARALLEL!
```

---

# Part 8: The Material - HZO

## What is HZO?

**H**afnium-**Z**irconium-**O**xide superlattice

```
Stacked layers:
┌─────────────┐
│    HfO₂     │ ← Hafnium oxide
├─────────────┤
│    ZrO₂     │ ← Zirconium oxide
├─────────────┤
│    HfO₂     │
├─────────────┤
│    ZrO₂     │
└─────────────┘
     ↑
  ~10 nm thick total
```

## Why HZO is Special

| Property | Value | Why It's Good |
|----------|-------|---------------|
| Thickness | ~10 nm | Fits in tiny chips! |
| Voltage to switch | ~1-3 V | Works with phone batteries |
| Endurance | 10¹² cycles | Lasts basically forever |
| States | ~30 levels | Stores way more info |
| CMOS compatible | ✅ | Can use existing factories |

## Key Numbers

| Parameter | Symbol | Value | Unit |
|-----------|--------|-------|------|
| Saturation Polarization | Ps | 25 | μC/cm² |
| Coercive Field | Ec | 1.0 | MV/cm |
| Film Thickness | t | 10 | nm |
| States | - | ~30 | - |

---

# Part 9: Neural Networks (Why This Matters for AI)

## What is a Neural Network?

It's inspired by your brain! Layers of "neurons" connected by "weights":

```
Input Layer      Hidden Layer      Output Layer
    ●─────────────────●─────────────────●
    ●─────────────────●─────────────────●
    ●─────────────────●─────────────────●
    ●─────────────────●─────────────────●
    
    ─── = connection with a weight
```

## How It Works

1. Input comes in (like pixels of an image)
2. Each connection multiplies input by its weight
3. Each neuron adds up all the weighted inputs
4. Repeat for each layer
5. Output = answer (like "this is a cat")

**The core operation is matrix-vector multiplication!** (Remember crossbar?)

## Training

Start with random weights → show lots of examples → adjust weights to reduce errors → repeat millions of times

IronLattice can potentially do training 1000× faster than regular computers!

---

# Part 10: The Three Demos

## Demo 1: Hysteresis Visualizer

**Difficulty:** ⭐ Easy (graphics only, no GPU compute)

**What it shows:**
- P-E hysteresis curve in real-time
- Voltage slider you can drag
- 30 analog states visualization
- Different materials to compare

**Technical:**
- CPU-based physics (Preisach model already coded!)
- Vulkan graphics for 2D line plotting
- GLFW window

## Demo 2: Crossbar Visualizer

**Difficulty:** ⭐⭐ Medium (GPU compute + graphics)

**What it shows:**
- Matrix-vector multiply animation
- Currents flowing through grid
- Non-idealities you can toggle (IR drop, sneak paths)
- Weight programming by clicking cells

**Technical:**
- Vulkan compute shaders for MVM
- 2D grid visualization
- Animation timing

## Demo 3: Phase-Field Simulator

**Difficulty:** ⭐⭐⭐ Advanced (heavy GPU compute)

**What it shows:**
- 3D domain structure in the crystal
- Domain walls moving when field applied
- Real-time TDGL solving
- Parameter sweeps (temperature, field)

**Technical:**
- 3D storage buffers
- TDGL compute shader
- Volume rendering
- 128³ grid = 2 million points

---

# Part 11: The Code Structure

```
ironlattice-vis/
│
├── demo1-hysteresis/      ← P-E curve demo
│   ├── cmd/               ← Main program
│   ├── pkg/ferroelectric/ ← Preisach model (DONE!)
│   └── shaders/           ← Graphics shaders
│
├── demo2-crossbar/        ← MVM animation demo
│   ├── cmd/               ← Main program
│   ├── pkg/crossbar/      ← Array model
│   └── shaders/           ← Compute + graphics
│
├── demo3-phasefield/      ← Domain simulation
│   ├── cmd/               ← Main program
│   ├── pkg/physics/       ← TDGL equations
│   └── shaders/           ← Heavy compute
│
├── docs/                  ← Documentation
├── papers/                ← Research papers (23 downloaded!)
├── opensource/            ← Reference projects
└── ELI5.md                ← You are here! 🎉
```

---

# Part 12: What You Need to Build It

## Software

| Tool | Purpose |
|------|---------|
| Go | Programming language |
| Vulkan SDK | GPU graphics and compute |
| GLFW | Window creation |
| go-vk | Go bindings for Vulkan |
| glslangValidator | Compile shaders |

## Install Commands

```bash
# Go
sudo apt install golang-go

# Vulkan
sudo apt install vulkan-tools vulkan-sdk

# GLFW
sudo apt install libglfw3-dev

# Go dependencies
go mod tidy
```

---

# Part 13: Glossary

| Term | Simple Meaning |
|------|----------------|
| **Ferroelectric** | Material that remembers which way you pushed it |
| **Polarization (P)** | How separated the charges are |
| **Hysteresis** | Going up ≠ going down (history matters) |
| **Coercive Field (Ec)** | Push needed to flip the polarization |
| **Saturation (Ps)** | Maximum possible polarization |
| **Remanent (Pr)** | Polarization that remains when you stop pushing |
| **Crossbar** | Grid of wires with memory at each intersection |
| **MVM** | Matrix-vector multiplication (core AI math) |
| **CIM** | Compute-in-Memory (do math where data lives) |
| **Preisach Model** | Simulating hysteresis with tiny switches |
| **TDGL** | Equation for how domains evolve over time |
| **Domain** | Region with same polarization direction |
| **Domain Wall** | Boundary between domains |
| **HZO** | Hafnium-Zirconium-Oxide (the magic material) |
| **Vulkan** | GPU programming interface |
| **GLSL** | Shader programming language |
| **SPIR-V** | Compiled shader format |

---

# Part 14: The Mission

## Current Status
- ✅ Preisach physics code works
- ✅ 23 research papers downloaded
- ✅ Demo structure created
- ✅ Physics documentation complete
- 🔲 Vulkan graphics pipeline (next step!)

## What's Next
1. Get Demo 1 showing a real P-E curve
2. Add interactivity (voltage slider)
3. Move to Demo 2 with compute shaders
4. Build Demo 3 for 3D simulation

## The Dream
Anyone can open these demos and **see** how ferroelectric compute-in-memory works. No PhD required!

---

**Congratulations! You now know enough to be the lead engineer. Go build it! 🚀🧠⚡**
