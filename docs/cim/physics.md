# Compute-in-Memory Physics: Why Analog Computing Matters

**Document Purpose**: Explain the physical and mathematical foundations of Compute-in-Memory (CIM) architectures, with focus on memristor/FeFET crossbar arrays.

**Target Audience**: Engineers, researchers, and students seeking to understand why CIM offers fundamental advantages over von Neumann architectures for AI workloads.

---

## Table of Contents

1. [The von Neumann Bottleneck](#1-the-von-neumann-bottleneck)
2. [Ohm's Law as Multiplication](#2-ohms-law-as-multiplication)
3. [Kirchhoff's Current Law as Addition](#3-kirchhoffs-current-law-as-addition)
4. [Complete Crossbar Circuit Model](#4-complete-crossbar-circuit-model)
5. [Energy Analysis](#5-energy-analysis)
6. [References](#6-references)

---

## 1. The von Neumann Bottleneck

### 1.1 Physical Separation Problem

Modern computers follow the von Neumann architecture, where **processing (CPU/GPU) and memory (DRAM) are physically separated**. This separation creates fundamental inefficiencies:

```
┌─────────────┐         Data Bus        ┌─────────────┐
│             │ ◄────────────────────► │             │
│  Processor  │                         │   Memory    │
│   (CPU)     │  ← Bottleneck here!    │   (DRAM)    │
│             │                         │             │
└─────────────┘                         └─────────────┘
```

**Key problem**: Data must shuttle back and forth between CPU and memory for every operation.

### 1.2 Energy Cost of Data Movement

The energy hierarchy reveals the problem:

| Operation | Energy Cost | Reference |
|-----------|-------------|-----------|
| 32-bit integer ADD | 0.9 pJ | Sze et al., 2017 |
| MAC operation (multiply-accumulate) | 0.25 pJ | Sze et al., 2017 |
| **DRAM access (64-bit)** | **~200 pJ** | Sze et al., 2017 |
| **Moving data from DRAM** | **~500 pJ** | IBM Research, 2023 |
| **Ratio (DRAM access / MAC)** | **220-1000×** | - |

**Critical insight**: **Accessing memory costs 220-1000× more energy than the computation itself.**

### 1.3 The Memory Wall

For AI workloads (deep learning inference/training):

- **90% of energy** is spent moving data between memory and processor
- **10% of energy** is spent on actual computation (MAC operations)
- **Performance bottleneck**: Memory bandwidth limits throughput, not computation speed

**Example**: A single matrix multiplication with a 1024×1024 weight matrix requires:
- **1,048,576 weight loads** from DRAM
- **1,048,576 MAC operations**
- If each DRAM access costs 200 pJ and each MAC costs 0.25 pJ:
  - Memory energy: 1,048,576 × 200 pJ = **210 mJ**
  - Computation energy: 1,048,576 × 0.25 pJ = **0.26 mJ**
  - **Memory dominates by 800×**

### 1.4 Why CIM Solves This

**Compute-in-Memory (CIM)** eliminates data movement by **storing weights inside the computational array**:

```
┌─────────────────────────────┐
│   Crossbar Array            │
│                             │
│   Weights stored in         │
│   memristors/FeFETs         │
│                             │
│   Computation happens       │
│   where data lives          │
└─────────────────────────────┘
```

**Result**: No DRAM access needed for weights during inference → 100-1000× energy reduction.

---

## 2. Ohm's Law as Multiplication

### 2.1 The Core Equation

**Ohm's Law for conductance:**

$$I = G \times V$$

Where:
- $I$ = output current (Amperes)
- $G$ = conductance (Siemens) = $1/R$ (stored weight)
- $V$ = input voltage (Volts)

**Physical insight**: A memristor/FeFET naturally performs multiplication because current flow is proportional to both conductance (stored value) and applied voltage (input signal).

### 2.2 Memristor as Analog Multiplier

```
    V_in (input voltage)
      |
      ▼
    ┌───┐
    │ G │  ← Conductance = stored weight W
    └───┘
      |
      ▼
    I_out = G × V = W × V_in
```

**Comparison to digital multiplication:**

| Digital Approach | Analog Memristor Approach |
|------------------|---------------------------|
| Read weight from memory | Weight is already encoded as G |
| Load into ALU | No movement needed |
| Execute multiply instruction | Physics does multiplication |
| Multiple clock cycles | Single timestep (~ns) |
| ~100+ transistors per multiplier | 1 memristor device |

### 2.3 Programming Conductance States

For analog computing, memristors/FeFETs store weights as **discrete conductance levels**:

$$G_i = G_{\text{min}} + i \times \frac{G_{\text{max}} - G_{\text{min}}}{N - 1}, \quad i = 0, 1, \ldots, N-1$$

Where:
- $N$ = number of analog states (e.g., 30 for HfO₂-ZrO₂ FeFETs)
- $G_{\text{min}}$, $G_{\text{max}}$ = minimum/maximum conductance
- $i$ = programmed state index

**Example for FeFET with 30 states:**
- $G_{\text{min}} = 1 \mu S$, $G_{\text{max}} = 100 \mu S$
- State 0: $G_0 = 1 \mu S$
- State 15: $G_{15} = 52.1 \mu S$
- State 29: $G_{29} = 100 \mu S$
- **Effective precision**: $\log_2(30) \approx 4.9$ bits per cell

---

## 3. Kirchhoff's Current Law as Addition

### 3.1 The Core Equation

**Kirchhoff's Current Law (KCL)** at a node:

$$\sum I_{\text{in}} = \sum I_{\text{out}}$$

**Applied to a crossbar column** (bit line):

$$I_j = \sum_{i=1}^{N} I_{ij} = \sum_{i=1}^{N} G_{ij} \times V_i$$

Where:
- $I_j$ = total output current on column $j$
- $I_{ij}$ = current contribution from crosspoint $(i, j)$
- $G_{ij}$ = conductance at crosspoint $(i, j)$ (weight $W_{ij}$)
- $V_i$ = input voltage on row $i$ (input $X_i$)

**Physical insight**: Currents automatically sum along the wire due to charge conservation. **No adder circuits needed**.

### 3.2 Physical Implementation

```
         V₁    V₂    V₃   (input voltages)
          │     │     │
    ──────┼─────┼─────┼───── Row 1
          ┆G₁₁  ┆G₁₂  ┆G₁₃
          ↓     ↓     ↓
    ──────┼─────┼─────┼───── Row 2
          ┆G₂₁  ┆G₂₂  ┆G₂₃
          ↓     ↓     ↓
    ──────┴─────┴─────┴───── Row 3
          │     │     │
          ↓     ↓     ↓
         I₁    I₂    I₃   (output currents)

    I₁ = G₁₁V₁ + G₂₁V₂ + G₃₁V₃
    I₂ = G₁₂V₁ + G₂₂V₂ + G₃₂V₃
    I₃ = G₁₃V₁ + G₂₃V₂ + G₃₃V₃
```

**Matrix-vector multiplication in one timestep:**

$$\mathbf{I} = \mathbf{G} \times \mathbf{V}$$

Where:
- $\mathbf{I} = [I_1, I_2, I_3]^T$ (output vector)
- $\mathbf{G}$ = 3×3 conductance matrix (stored weights)
- $\mathbf{V} = [V_1, V_2, V_3]^T$ (input vector)

### 3.3 Comparison to Digital Approach

**Digital matrix-vector multiply (3×3 example):**
1. Load $W_{11}$ from memory
2. Multiply $W_{11} \times X_1$
3. Store partial result
4. Load $W_{21}$ from memory
5. Multiply $W_{21} \times X_2$
6. Add to partial result
7. Repeat 9 times (3×3 elements)
8. **Total**: 9 memory loads, 9 multiplies, 6 adds

**Analog crossbar (3×3 example):**
1. Apply $V_1, V_2, V_3$ to rows
2. Measure $I_1, I_2, I_3$ on columns
3. **Done in ~10 ns**

**Speedup**: Digital requires 24 operations sequentially; analog completes in parallel in one physical timestep.

---

## 4. Complete Crossbar Circuit Model

### 4.1 Ideal Crossbar (No Parasitics)

For an $N \times M$ crossbar (N rows, M columns):

**Circuit equation:**

$$\mathbf{I}_{\text{out}} = \mathbf{G} \times \mathbf{V}_{\text{in}}$$

Where:
- $\mathbf{I}_{\text{out}} \in \mathbb{R}^{M}$ (M output currents on bit lines)
- $\mathbf{G} \in \mathbb{R}^{N \times M}$ (N×M conductance matrix)
- $\mathbf{V}_{\text{in}} \in \mathbb{R}^{N}$ (N input voltages on word lines)

**Component count:**
- $N \times M$ memristors/FeFETs
- $N$ voltage sources (word line drivers)
- $M$ current sensors (TIA/ADC)

### 4.2 Real Crossbar (With Wire Resistance)

**Parasitic elements:**
- **Word line resistance**: $R_{\text{WL}}$ per cell segment (typically 2.5 Ω @ 45nm)
- **Bit line resistance**: $R_{\text{BL}}$ per cell segment (typically 2.5 Ω @ 45nm)
- **Interconnect capacitance**: $C_{\text{wire}}$ (affects switching speed)

**Modified circuit model:**

```
                     V_in(1)              V_in(2)              V_in(3)
                       │                    │                    │
                       R_WL                 R_WL                 R_WL
         ┌─────────────┴───────┬────────────┴───────┬────────────┴──────┐
         │                     │                    │                   │
    ┌────┴────┐          ┌─────┴────┐         ┌─────┴────┐             │ WL1
    │  G₁₁    │          │  G₁₂     │         │  G₁₃     │             │
    └────┬────┘          └─────┬────┘         └─────┬────┘             │
         │ R_BL                │ R_BL               │ R_BL              │
    ┌────┴────┐          ┌─────┴────┐         ┌─────┴────┐             │
    │  G₂₁    │          │  G₂₂     │         │  G₂₃     │             │ WL2
    └────┬────┘          └─────┬────┘         └─────┬────┘             │
         │                     │                    │                   │
         ↓                     ↓                    ↓                   ↓
       I_out(1)              I_out(2)             I_out(3)           I_leak
        (BL1)                 (BL2)                (BL3)
```

**Key non-idealities:**

1. **IR Drop (Voltage Attenuation)**:
   - Voltage decreases along word line: $V_{\text{actual}}(k) = V_{\text{in}} - I_{\text{row}} \times k \times R_{\text{WL}}$
   - Far crosspoints see lower voltage → computation error
   - **Error grows with array size** (100× worse for 1024×1024 vs 32×32)

2. **Sneak Paths**:
   - Current flows through unintended paths in large arrays
   - **Mitigation**: Add selector devices (1T1R, 1D1R) at each crosspoint

3. **Line Resistance Equation**:

$$I_j = \sum_{i=1}^{N} G_{ij} \times \left( V_i - I_{\text{row},i} \times R_{\text{WL},i} \right)$$

Where:
- $I_{\text{row},i} = \sum_{j=1}^{M} I_{ij}$ (total current on row $i$)
- $R_{\text{WL},i} = i \times R_{\text{WL,segment}}$ (cumulative WL resistance)

### 4.3 Circuit Analysis Complexity

**Node count for $N \times M$ crossbar with parasitics:**
- **Voltage nodes**: $2 \times M \times N$ (top and bottom of each crosspoint)
- **Resistors**: $3 \times M \times N$
  - $M \times N$ memristors
  - $M \times N$ WL segments
  - $M \times N$ BL segments

**Example**: 128×128 crossbar
- Voltage nodes: $2 \times 128 \times 128 = 32,768$
- Resistors: $3 \times 128 \times 128 = 49,152$
- **Requires circuit simulator (SPICE) for accurate analysis**

---

## 5. Energy Analysis

### 5.1 Why Data Movement Dominates

**Energy hierarchy** (from Horowitz, 2014 and Sze et al., 2017):

| Operation | Energy (pJ) | Relative Cost |
|-----------|-------------|---------------|
| 32-bit register access | 0.05 | 1× |
| 32-bit integer ADD | 0.9 | 18× |
| 32-bit FP multiply | 3.7 | 74× |
| 32-bit MAC operation | 0.25 | 5× |
| SRAM access (32 KB) | 5 | 100× |
| DRAM access (64-bit) | 200 | 4,000× |
| **Moving data from DRAM** | **500** | **10,000×** |

**Critical observation**: Memory access costs **800-4000× more** than computation.

### 5.2 CIM Energy Equation

**Energy per MAC in analog crossbar:**

$$E_{\text{MAC}} = C \times V^2 + G \times V \times t$$

Where:
- $C \times V^2$: Capacitive switching energy (charging bit/word lines)
  - $C$: Total capacitance (wire + device)
  - $V$: Operating voltage
- $G \times V \times t$: Resistive current flow (steady-state energy)
  - $G$: Device conductance
  - $t$: Integration time

**Typical values (HfO₂-ZrO₂ FeFET):**
- $V = 2$ V (operating voltage)
- $C = 1$ fF (femtofarad, device + local wire)
- $G = 50 \mu S$ (mid-state conductance)
- $t = 10$ ns (read time)

$$E_{\text{MAC}} = (1 \times 10^{-15}) \times 4 + (50 \times 10^{-6}) \times 2 \times (10 \times 10^{-9})$$
$$E_{\text{MAC}} = 4 \times 10^{-15} + 1 \times 10^{-12} \approx 1 \text{ pJ}$$

**Comparison:**
- Digital MAC (DRAM access): ~200 pJ (memory access) + 0.25 pJ (computation) ≈ **200 pJ**
- Analog CIM MAC: **1 pJ**
- **Savings**: 200× lower energy

### 5.3 Experimental Results from Literature

| Platform | Efficiency | Speedup vs CPU | Reference |
|----------|------------|----------------|-----------|
| **Digital GPU (NVIDIA A100)** | ~1 TOPS/W | 1× (baseline) | NVIDIA, 2020 |
| **Analog CIM (PCM, IBM)** | 2,900 TOPS/W | 2,900× | Joshi et al., Nature, 2020 |
| **FeFET CIM (3-bit)** | 36.5 TOPS/W | 36.5× | Cheng et al., IEEE, 2021 |
| **IBM NorthPole (digital CIM)** | 47× faster, 73× more efficient | 47× speed | IBM Research, 2023 |
| **HfO₂-ZrO₂ FeFET (30-state)** | ~100-500 TOPS/W (estimated) | 100-500× | Dr. external research group, 2025 |

**Key insight**: Even **digital** CIM (NorthPole) achieves 47× speedup by eliminating data movement. **Analog** CIM with memristors/FeFETs can reach 100-2900× efficiency gains.

### 5.4 Scaling Analysis

**Energy breakdown for 1024×1024 matrix-vector multiply:**

| Component | Digital (GPU) | Analog CIM | Savings |
|-----------|---------------|------------|---------|
| **Memory access** | 1,048,576 × 200 pJ = **210 mJ** | 0 (weights in-situ) | ∞ |
| **Computation** | 1,048,576 × 0.25 pJ = 0.26 mJ | 1,048,576 × 1 pJ = 1.05 mJ | 0.25× |
| **ADC (CIM only)** | - | 1024 × 10 pJ = 10.2 µJ | - |
| **Total** | **210.26 mJ** | **1.06 mJ** | **198×** |

**Caveats:**
- Assumes single matrix-vector multiply
- ADC energy depends on resolution (higher bits = more energy)
- Wire resistance increases energy for large arrays

### 5.5 Technology Comparison

**Energy efficiency vs precision tradeoff:**

```
Efficiency (TOPS/W)
    ↑
3000│                    ● Analog PCM (1-bit)
    │
1000│         ● FeFET 30-state (estimated)
    │
 100│     ● FeFET 3-bit
    │
  10│  ● Digital CIM (NorthPole)
    │
   1│● GPU (A100)
    └─────────────────────────────────────→
     1-bit   3-bit   5-bit   8-bit   16-bit
                 Precision (bits)
```

**Observations:**
- **Low precision (1-3 bits)**: Analog CIM dominates (1000-3000 TOPS/W)
- **Medium precision (4-6 bits)**: FeFET multi-level cells competitive (100-500 TOPS/W)
- **High precision (8+ bits)**: Digital CIM or GPU better (1-50 TOPS/W)

**Application mapping:**
- **Edge AI inference** (low precision OK): Analog CIM ideal
- **Training** (high precision needed): Digital GPU/TPU better
- **Mixed-signal hybrid**: Analog for inference, digital for training

---

## 6. References

### Foundational Papers

1. **Sze, V., Chen, Y.-H., Yang, T.-J., & Emer, J. S.** (2017). "Efficient Processing of Deep Neural Networks: A Tutorial and Survey." *Proceedings of the IEEE*, 105(12), 2295–2329. https://doi.org/10.1109/JPROC.2017.2761740
   - **Energy hierarchy data** (DRAM vs MAC costs)

2. **Horowitz, M.** (2014). "Computing's Energy Problem (and what we can do about it)." *IEEE International Solid-State Circuits Conference (ISSCC)*.
   - **Memory wall analysis**

3. **Joshi, V., et al.** (2020). "Accurate deep neural network inference using computational phase-change memory." *Nature Communications*, 11, 2473. https://doi.org/10.1038/s41467-020-16108-9
   - **2,900 TOPS/W analog CIM result**

4. **IBM Research** (2023). "NorthPole: A 12nm Digital AI Processor."
   - **Digital CIM: 47× faster, 73× more efficient than GPU**

### FeFET-Specific Research

5. **Tour, J. R.** (2025). "Ferroelectric Compute-in-Memory: A New Paradigm for AI Hardware." *COSM 2025 Conference*.
   - **30 analog states in HfO₂-ZrO₂ superlattice**

6. **Cheng, C.-H., et al.** (2021). "Ferroelectric FET for Analog Computing." *IEEE Transactions on Electron Devices*.
   - **FeFET CIM: 36.5 TOPS/W (3-bit)**

### Wire Resistance Analysis

7. **Gokmen, T., & Vlasov, Y.** (2016). "Acceleration of Deep Neural Network Training with Resistive Cross-Point Devices: Design Considerations." *Frontiers in Neuroscience*, 10, 333. https://doi.org/10.3389/fnins.2016.00333
   - **IR drop modeling in crossbar arrays**

8. **Shafiee, A., et al.** (2016). "ISAAC: A Convolutional Neural Network Accelerator with In-Situ Analog Arithmetic in Crossbars." *ACM SIGARCH Computer Architecture News*, 44(3), 14–26.
   - **Circuit-level crossbar simulation**

---

## Appendix A: Notation Summary

| Symbol | Meaning | Units |
|--------|---------|-------|
| $I$ | Current | Amperes (A) |
| $V$ | Voltage | Volts (V) |
| $G$ | Conductance | Siemens (S) |
| $R$ | Resistance | Ohms (Ω) |
| $C$ | Capacitance | Farads (F) |
| $E$ | Energy | Joules (J) or pJ |
| $N$ | Number of rows (inputs) | - |
| $M$ | Number of columns (outputs) | - |
| $t$ | Time | Seconds (s) |
| $\mathbf{G}$ | Conductance matrix | S |
| $\mathbf{V}_{\text{in}}$ | Input voltage vector | V |
| $\mathbf{I}_{\text{out}}$ | Output current vector | A |

---

## Appendix B: Energy Hierarchy Summary

**From lowest to highest energy cost:**

1. **Register access**: 0.05 pJ
2. **MAC operation**: 0.25 pJ
3. **Integer ADD**: 0.9 pJ
4. **FP multiply**: 3.7 pJ
5. **SRAM access**: 5 pJ
6. **DRAM access**: 200 pJ (800× MAC)
7. **DRAM data movement**: 500 pJ (2000× MAC)

**Key takeaway**: Memory hierarchy dominates energy budget → CIM eliminates steps 6-7 by storing weights in-situ.

---

**Document Version**: 1.0
**Last Updated**: 2026-01-25
**Author**: FeCIM Lattice Tools Project
**License**: MIT
