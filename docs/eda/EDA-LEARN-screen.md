# The Mission: Bridging AI and Silicon

## Why This Module Exists

You have designed a neural network in Python. You have trained it to recognize images or process language. **But how does it become physical reality?**

In standard computing, we rely on **Abstraction**. You write code, and a compiler turns it into binary for a generic CPU.

In **Compute-in-Memory (CIM)**, we break that abstraction. We don't just run code; we map the distinct "weights" of your neural network directly onto the physical "conductance" of Ferroelectric transistors (FeFETs).

**This module is the bridge between your abstract Python model and the concrete laws of physics.**

---

## The "Gap" in Modern Tools

Why can't we just use standard chip design tools?

### 1. The Digital vs. Analog problem
*   **Standard Tools (OpenLane/Cadence):** Think in "0" and "1".
*   **FeCIM Reality:** We need **30+ discrete analog levels** to store a weight. Standard tools see this as noise.

### 2. The "Memory Wall"
*   **Standard Architecture:** Memory and Compute are separate (Von Neumann bottleneck).
*   **FeCIM Reality:** Memory **IS** the Compute. We do math *inside* the storage array using Ohm's Law ($I = G \cdot V$) and Kirchhoff's Law ($\sum I$).

### 3. The Missing Component
*   **PDKs (Process Design Kits):** The factory provides standard transistors (NMOS/PMOS).
*   **FeCIM Reality:** Factories do not yet have standard "FeFET" cells. We must model them ourselves using physics equations (Landau-Khalatnikov).

---

## Goals of This Module

This visualizations and design suite has three clear goals:

### Goal 1: TRANSFORM
**Map Abstract Math to Physical States**
We take your floating-point weights (e.g., `0.735`) and "quantize" them into the limited physical states your hardware supports (e.g., `Conductance Level 22`).

### Goal 2: SIMULATE
**Prove It Works Before Building It**
We generate industry-standard **SPICE** netlists. This allows you to run a physics-accurate simulation of your customized array using **ngspice**, verifying that your "Level 22" actually produces the correct current.

### Goal 3: REALIZE
**Generate the Blueprint**
Ultimately, we aim to produce a **GDSII** layout. This is the geometric file sent to a fab (like IHP or SkyWater) to pattern the actual silicon.

---

## Your Workflow

1.  **Design** your crossbar parameters (Size, Tech Node, FeFET Model).
2.  **Import** your weights from PyTorch/ONNX.
3.  **Visualize** the mapping efficiency and quantization error.
4.  **Export** to SPICE for verification or GDSII for fabrication.
