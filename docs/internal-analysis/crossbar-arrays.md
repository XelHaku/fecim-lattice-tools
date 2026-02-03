# Research Synthesis: Crossbar Arrays and Matrix-Vector Multiplication (MVM)

> **Note:** Internal analysis note. Values are reported/illustrative and not validated by this codebase.

**Internal Analysis Document - FeCIM Lattice Tools Project**

## 1. Executive Summary

Ferroelectric Compute-in-Memory (FeCIM) leverages crossbar array architectures to perform analog Matrix-Vector Multiplication (MVM) directly within the memory fabric. This project uses a 30-level demo baseline (configurable) while literature reports multi-level states (unverified). While passive (0T1R) architectures offer maximum density (4F²), non-idealities such as IR drop and sneak paths necessitate the use of isolated architectures (1T1R, 2T1R) for large-scale deployments. Literature reports high-accuracy MNIST results and energy improvements vs NAND, but those values are **not** validated here.

---

## 2. Matrix-Vector Multiplication (MVM) Physics

The core advantage of crossbar arrays is the ability to perform MVM in $O(1)$ time complexity using fundamental physical laws.

### 2.1 Ohm's Law: I = G × V
Each cell in the crossbar acts as a programmable conductance ($G_{ij}$). When an input voltage ($V_j$) is applied to the column, the resulting current through the cell is proportional to the stored weight:
$$I_{ij} = G_{ij} \times V_j$$

### 2.2 Kirchhoff's Current Law: I_out = Σ(G_ij × V_j)
Currents from all cells in a row are summed along the horizontal metal line:
$$I_{i} = \sum_{j=0}^{N-1} G_{ij} \times V_j$$
This physical summation performs the "accumulate" part of the Multiply-Accumulate (MAC) operation simultaneously for all elements in a row.

### 2.3 O(1) Computation Advantage
In digital systems, MVM is $O(N^2)$ or $O(N \cdot M)$. In a crossbar, the entire operation is completed in a single clock cycle (determined by RC delay), regardless of matrix size, provided the peripheral circuits (DAC/ADC) can handle the parallel load.

---

## 3. Array Architectures

Selection of architecture involves a trade-off between density, isolation, and accuracy.

| Feature | 0T1R (Passive) | 1T1R | 2T1R |
|---------|----------------|------|------|
| **Access Device** | None | 1 NMOS Transistor | 2 NMOS Transistors |
| **Cell Area** | 4F² | 8-12F² | 10-12F² |
| **Isolation** | None | 1000:1 (Transistor OFF) | Best (Separate R/W) |
| **Sneak Error** | 5-20% | <0.1% | Negligible |
| **Density** | Highest | Medium | Lowest |
| **Best For** | Small Demo (<32x32) | Standard Production | High-Precision |

### 3.1 0T1R (Passive)
- **Density**: 4F² (e.g., 1936 nm² at 22nm node).
- **Challenge**: Significant sneak path errors (up to 20%) without self-rectifying materials.

### 3.2 1T1R (Standard)
- **Isolation**: Transistor OFF-state provides ~1000:1 isolation.
- **Accuracy**: Industry standard for arrays up to 256x256.

### 3.3 2T1R (High Precision)
- **Features**: Separate read/write paths eliminate read disturb and maximize SNR (80-100 dB).

---

## 4. Non-Idealities

Real-world crossbars suffer from parasitic effects that degrade MVM accuracy.

### 4.1 IR Drop
Voltage loss along metal wires due to line resistance ($R_{line} \approx 2.5 \Omega$ per cell pitch).
- **Formula**: $V_{eff}(i,j) = V_{in} - \sum (I \cdot R_{line})$.
- **Impact**: 10-20% voltage reduction in 64x64 arrays; becomes critical beyond 128x128.
- **Mitigation**: Thicker metal, array tiling, or pre-distortion compensation.

### 4.2 Sneak Paths
Unintended current loops through unselected cells in passive arrays.

**Sneak Path Analysis Diagram (3x3 Example):**
```text
     BL0     BL1     BL2
      │       │       │
      ↓       ↓       ↓
WL0 ──●───────●───────●──  (0V)
      │       │       │
WL1 ──●───────●═══════●──  (1V)  <-- Selected Row
      │       ║       ║
     G10    ║G11║    G12   <-- Target Cell (G11)
            ║   ║
WL2 ──●─────╬───╬─────●──  (0V)
      │     ║   ║     │
     G20    G21  G22
      │     ║   ║     │
     GND   ═╩═══╝    GND
           SIGNAL +
         SNEAK CURRENT
```
- **Error**: $I_{sneak} \approx V \times G_{avg} \times (N_{rows} \times N_{cols}) / 3$.

### 4.3 Device Variation & Drift
- **Variation**: 3-10% $\sigma$ (Device-to-device). FeFET typically 5-10%.
- **Drift**: Logarithmic decay $G(t) = G_0 (t/t_0)^{-\nu}$.
- **Coefficient**: $\nu \approx 0.0005 - 0.001$ for FeFET (superior to RRAM's 0.05).

---

## 5. Conductance Modeling

FeFET devices map analog states to specific conductance values.

- **Range**: $G_{min} = 10 \mu S$ (OFF), $G_{max} = 100 \mu S$ (ON).
- **Linear Model**: $G = G_{min} + (G_{max} - G_{min}) \times \frac{level}{29}$.
- **Exponential Model**: $G = G_{min} \times \exp(\ln(G_{max}/G_{min}) \times \frac{level}{29})$.
- **Physics**: Exponential scaling more accurately reflects subthreshold FeFET behavior.

---

## 6. Array Size Guidelines

Scaling limits are determined primarily by IR drop and peripheral circuit overhead.

- **8×8 to 16×16**: Ideal for educational demos and basic logic gates.
- **64×64**: Standard size for real-time inference (IR drop manageable at 5-10%).
- **128×128**: Extreme IR drop (12-20%); 1T1R architecture becomes mandatory to maintain SNR.

---

## 7. Accuracy Metrics

FeCIM performance is benchmarked using standard neural network tasks.

- **Reported MNIST results**: See cited literature in references (not validated here).
- **Inference Accuracy**: Can drop with quantization and non-idealities; magnitude is model-dependent (illustrative).

---

## 8. Simulation Tools

Professional tools used for crossbar validation and co-design.

- **CrossSim (Sandia)**: 40× faster than SPICE; PyTorch/TensorFlow interface for large-scale DNN accuracy.
- **badcrossbar (UCL)**: Nodal analysis solver for exact IR drop and parasitic resistance visualization.
- **NeuroSim (Georgia Tech)**: Comprehensive toolchain covering device-circuit-architecture benchmarks.

---

## 9. References

- **[10.1016/j.softx.2020.100617]** Joksas, D. & Mehonic, A. (2020). *badcrossbar: A Python tool for computing and plotting currents and voltages in passive crossbar arrays*. SoftwareX.
- **[10.3389/frai.2021.659060]** Chen, P.Y., et al. (2021). *NeuroSim: Circuit-Level Crossbar Simulator*. Frontiers in AI.
- **[SAND2021-12318C]** Xiao, T.P., et al. (2021). *CrossSim: GPU-Accelerated Simulation of Analog Neural Networks*. Sandia National Labs.
- **[10.1038/s41467-023-42110-y]** Nature Communications 2023 (FeFET 96.6% MNIST).
- **[10.1016/j.jallcom.2025.034309]** ScienceDirect 2025 (HZO Reservoir 98.24% MNIST).
