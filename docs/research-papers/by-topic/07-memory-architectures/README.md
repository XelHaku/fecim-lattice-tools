# Memory Architectures for AI Accelerators

## Overview

This directory contains research on advanced memory architectures for AI workloads, with emphasis on 3D stacking, memory-side acceleration, and High Bandwidth Memory (HBM) thermal/power challenges. While FeCIM embeds computation directly in ferroelectric memory (compute-in-memory), these papers explore complementary approaches: placing compute logic near memory (processing-in-memory, PIM) or optimizing memory hierarchies for AI workloads. Understanding these architectures helps position FeCIM's advantages and identify hybrid opportunities.

## Papers in this Directory

### 3D Memory-Side Acceleration
- **`3d_memory_side_acceleration_2024.pdf`** - 3D-stacked DRAM with logic dies for memory-side acceleration (processing-near-memory)
- **`memory_neural_accelerators_3d_2024.pdf`** - Neural network accelerators with 3D-stacked memory integration

### HBM Challenges for Large Models
- **`hbm_thermal_power_llm_2024.pdf`** - Thermal and power challenges of HBM for large language model (LLM) inference

## Key Findings

### Memory Bottleneck in AI: The "Von Neumann Wall"

**Problem:**
```
Traditional Architecture:
CPU/GPU <------ Memory Bus ------> DRAM
         (Bottleneck: 100 GB/s, 10 pJ/bit)
```

**AI Workload Characteristics:**
- **Data movement dominates energy**: 80-90% of total power for LLM inference
- **Memory bandwidth-limited**: GPUs idle waiting for data (50-70% utilization)
- **Capacity vs. bandwidth trade-off**: HBM provides bandwidth, but limited capacity (80 GB)

**Energy Breakdown (LLM Inference on A100 GPU):**
- Compute (FP16 MAC): 10%
- DRAM access: 60%
- On-chip data movement: 20%
- Other (control, I/O): 10%

### 3D Memory Architectures

#### 1. 3D-Stacked DRAM + Logic (HBM-PIM)

**Architecture:**
```
┌─────────────┐
│   Logic Die │  ← AI accelerator logic
├─────────────┤
│  DRAM Die 8 │
│  DRAM Die 7 │
│     ...     │  ← Through-silicon vias (TSVs)
│  DRAM Die 1 │
└─────────────┘
```

**Benefits:**
- **1000× bandwidth**: Internal TSV bandwidth ~1 TB/s vs. 100 GB/s off-chip
- **100× energy efficiency**: 0.1 pJ/bit vs. 10 pJ/bit off-chip
- **Reduced latency**: <10 ns vs. >100 ns off-chip

**Challenges:**
- **Thermal**: Logic die generates heat, DRAM retention degrades >85°C
- **Yield**: Stacking reduces overall yield (multi-die product)
- **Cost**: TSV adds 20-30% manufacturing cost

**Demonstrated Systems:**
- **Samsung HBM-PIM**: 1.2 TB/s bandwidth, 16 GB capacity
- **SK Hynix GDDR6-AiM**: Processing-in-memory for AI
- **UPMEM**: 2000 cores embedded in DDR4 DIMMs

#### 2. Chiplet-Based Disaggregation

**Concept:** Separate memory dies from compute dies, connect via high-speed interconnect

**Example: AMD MI300X**
```
[GPU Chiplet 1] [GPU Chiplet 2] ... [GPU Chiplet 8]
        ↕              ↕                    ↕
   [HBM Stack 1]  [HBM Stack 2]  ...  [HBM Stack 8]
```

**Benefits:**
- **Scalability**: Add more HBM stacks independently
- **Yield**: Bad chiplet doesn't scrap entire package
- **Flexibility**: Mix different memory types (HBM3, GDDR6, etc.)

### HBM Thermal Challenges for LLMs

**Problem:** LLM inference is memory-bound, not compute-bound

**Thermal Analysis (from hbm_thermal_power_llm_2024.pdf):**
- **HBM power**: 15-20 W per stack (8 stacks = 120-160 W)
- **Hot spots**: TSV regions reach 95-105°C under sustained access
- **DRAM retention**: Requires refresh every 64 ms, increases to 32 ms at 85°C
- **Throttling**: System reduces frequency at >85°C, losing 30-50% performance

**Mitigation Strategies:**
1. **Thermal-aware scheduling**: Rotate HBM access to distribute heat
2. **Liquid cooling**: Direct liquid cooling of HBM stacks (data center only)
3. **Underclocking**: Reduce HBM frequency from 3.2 GT/s to 2.4 GT/s (-25% power, -20% perf)
4. **Model compression**: Quantization reduces memory footprint and accesses

### 3D Integration vs. Compute-in-Memory

| Approach | Architecture | Bandwidth | Energy/Access | Non-Volatile | CMOS Compatible |
|----------|--------------|-----------|---------------|--------------|-----------------|
| **HBM (Baseline)** | 2.5D (interposer) | 1 TB/s | 10 pJ/bit | No | Yes |
| **HBM-PIM** | 3D (TSV) | 1 TB/s | 0.1 pJ/bit | No | Yes |
| **FeCIM** | Monolithic CIM | N/A (in-situ) | 0.001 pJ/bit | Yes | Yes |

**Key Insight:** FeCIM eliminates data movement entirely by computing where data is stored, achieving 1000× better energy efficiency than even 3D-stacked DRAM.

### Processing-in-Memory (PIM) vs. Compute-in-Memory (CIM)

**Processing-in-Memory (PIM):**
- Adds digital logic near DRAM banks
- Performs operations on data after reading from DRAM
- Energy savings: 10-100× (reduced data movement)
- Example: Samsung HBM-PIM for matrix multiplication

**Compute-in-Memory (CIM) - FeCIM:**
- Computation happens within memory array itself
- Exploits analog physics (Ohm's law, KCL)
- Energy savings: 1000-10000× (no data movement)
- Example: FeCIM crossbar for MVM

**Complementary Roles:**
- **PIM**: Best for irregular access patterns, general-purpose computing
- **CIM**: Best for regular MVM operations (DNNs, transformers)

## Relevance to FeCIM

### Positioning FeCIM in Memory Hierarchy

**Traditional AI System:**
```
GPU Compute ← 10 pJ/bit → HBM ← 100 pJ/bit → DRAM ← 1000 pJ/bit → SSD
```

**FeCIM-Augmented System:**
```
GPU (complex ops) ← 10 pJ/bit → HBM (temporary data)
                                     ↕
                              FeCIM Crossbar (MVM: 0.01 pJ/bit, non-volatile)
```

**Hybrid Strategy:**
1. **Frequent MVM operations**: FeCIM crossbar (30-level demo baseline; conference claim, non-volatile)
2. **Irregular access**: HBM-PIM (flexible, high bandwidth)
3. **Long-term storage**: Flash/SSD (high capacity)

### 3D Integration Roadmap for FeCIM

**Near-Term (Demonstrated):**
- **22nm BEOL FeFET** (CEA-Leti 2024): Single layer above logic transistors
- **Crossbar size**: 128×128 demonstrated

**Medium-Term (2-3 years):**
- **Multi-layer FeFET**: 4-8 layers in BEOL (similar to 3D NAND)
- **Crossbar capacity**: 512×512 per layer, 8 layers = 2 MB analog weights
- **Target**: Transformer models (GPT-2 small: 117M parameters)

**Long-Term (5-10 years):**
- **128-layer FeFET**: Vertical FeFET arrays (inspired by 3D NAND with 200+ layers)
- **Crossbar capacity**: >1 GB analog weights
- **Target**: LLM inference (7B parameter models)

### Thermal Management Lessons for FeCIM

**From HBM Thermal Study:**
1. **Hot spot avoidance**: Distribute MVM operations across multiple crossbars
2. **Thermal-aware mapping**: Place frequently accessed weights in cooler regions
3. **Active cooling**: For high-density 3D FeCIM, consider micro-channel cooling

**FeCIM Advantages:**
- **Lower power density**: 1 mW/mm² (FeCIM) vs. 10 mW/mm² (HBM)
- **Wider temperature range**: -40°C to 125°C (automotive grade FeFET)
- **Thermal compensation**: On-chip temperature sensing + weight calibration

### FeCIM as HBM Alternative for Inference

**LLM Inference Challenge:** 7B parameter model requires:
- **Weight storage**: 14 GB (FP16) or 7 GB (INT8)
- **HBM capacity**: 80 GB (fits, but expensive)
- **HBM bandwidth**: 3 TB/s (underutilized for inference)
- **HBM power**: 150 W (thermal challenge)

**FeCIM Solution (Speculative):**
- **Weight storage**: 7 GB in FeCIM (30 states demo baseline ≈ INT8)
- **Power**: <1 W (1000× less data movement)
- **Instant-on**: Non-volatile, no weight loading
- **Footprint**: <100 mm² (128-layer 3D FeCIM)

**Challenge:** Current FeCIM is 0.5 mm² for 128×128 = 16 KB. Scaling to 7 GB requires:
- 437,000× capacity increase
- 3D stacking: 100 layers = 4,370× per layer
- Crossbar scaling: 10× per dimension = 100×
- Total: 100 (3D) × 100 (scaling) = 10,000× (short by 40×)

**Conclusion:** FeCIM alone cannot replace HBM for LLMs, but can accelerate MVM-heavy layers.

## Related Topics

- **[01-ferroelectric-materials](../01-ferroelectric-materials/)** - 3D BEOL FeFET integration
- **[04-cim-architectures](../04-cim-architectures/)** - CIM vs. PIM trade-offs
- **[15-benchmarking](../15-benchmarking/)** - Energy/performance comparisons
- **[13-low-power-design](../13-low-power-design/)** - Power optimization strategies
- **[16-scaling-and-integration](../16-scaling-and-integration/)** - 3D integration roadmap

## Implementation Considerations for FeCIM

### Short-Term (Current FeCIM Lattice Tools)
No direct implementation needed, but use insights for:
1. **Energy modeling**: Compare FeCIM vs. HBM for data movement energy
2. **Thermal simulation**: Add thermal effects to crossbar model (future module)
3. **3D visualization**: Show potential 3D FeFET stacking (educational)

### Medium-Term (Research Direction)
Hybrid FeCIM + HBM architecture:
1. **Model partitioning**: Identify MVM-heavy layers for FeCIM
2. **Weight mapping**: Assign weights to FeCIM vs. HBM based on access frequency
3. **Performance estimation**: Compare hybrid vs. HBM-only for transformers

### Long-Term (Beyond MNIST)
3D FeCIM for LLM Inference:
1. **Capacity roadmap**: Track 3D NAND scaling trends (now at 200+ layers)
2. **Thermal modeling**: Simulate heat dissipation in 128-layer FeFET stack
3. **Yield analysis**: Multi-layer yield and redundancy strategies

## References for FeCIM Context

**High Priority (Understanding Trade-offs):**
1. `3d_memory_side_acceleration_2024.pdf` - PIM vs. CIM comparison
2. `memory_neural_accelerators_3d_2024.pdf` - 3D integration strategies

**Medium Priority (Thermal Management):**
3. `hbm_thermal_power_llm_2024.pdf` - Thermal challenges for scaled FeCIM

## Conclusion for FeCIM Roadmap

**Key Insights:**

1. **Memory bottleneck is real**: 80-90% of LLM inference energy is data movement
2. **3D integration is mature**: HBM demonstrates commercial viability of 3D stacking
3. **FeCIM's advantage**: Eliminating data movement entirely beats even best 3D DRAM

**Strategic Positioning:**
- **FeCIM ≠ HBM replacement**: Different roles (CIM vs. general memory)
- **FeCIM = Accelerator**: Best for MVM-dominated workloads (MNIST, CNNs, transformer layers)
- **Hybrid future**: FeCIM for compute + HBM for flexibility

**3D FeCIM Scaling Path:**
- Short-term: 1-8 layers (2-32 MB analog weights)
- Long-term: 64-128 layers (1-2 GB analog weights)
- Ultimate: Competitive with HBM for inference-only workloads
