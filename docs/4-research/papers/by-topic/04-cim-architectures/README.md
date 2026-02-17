# Compute-in-Memory (CIM) Architectures

## Overview

This directory contains research on compute-in-memory architectures, with emphasis on analog crossbar arrays for matrix-vector multiplication (MVM), peripheral circuit design (ADC/DAC/TIA), and system-level integration. These architectures exploit physical laws (Ohm's law, Kirchhoff's current law) to perform massively parallel computations in the analog domain, achieving 10-1000× energy efficiency improvements over digital approaches. Critical for understanding how FeCIM's 30-level demo baseline translates to system-level performance.

## Papers in this Directory

### Survey and Overview Papers
- **`cim_landscape_overview_2024.pdf`** - Comprehensive 2024 overview of CIM landscape across memory technologies
- **`Analog_AI_Accelerators_Survey_arXiv.pdf`** - Survey of analog AI accelerator architectures and design trade-offs
- **`Analog_AI_Promise_arXiv.pdf`** - Vision paper on analog AI's potential and challenges
- **`Analog_CIM_Energy_Efficiency_arXiv.pdf`** - Energy efficiency analysis of analog CIM vs. digital accelerators
- **`FeFET_CIM_Energy_Efficiency_arXiv_2024.pdf`** - FeFET CIM energy efficiency study (arXiv 2024)
- **`ferroelectric_CIM_review_2023.pdf`** - Comprehensive review of ferroelectric-based CIM systems
- **`In_Memory_Computing_Deep_Learning_arXiv.pdf`** - In-memory computing for deep learning applications
- **`Memristor_CIM_Survey_arXiv.pdf`** - Memristive CIM architectures and programming strategies
- **`Neuromorphic_Computing_Hardware_arXiv.pdf`** - Neuromorphic hardware architectures overview

### Crossbar Array Fundamentals
- **`Crossbar_Sneak_Path_Analysis_arXiv.pdf`** - Analysis of sneak path currents and mitigation strategies
- **`sneak_path_self_rectifying_arrays_2022.pdf`** - Self-rectifying memory cells to suppress sneak paths
- **`RRAM_Crossbar_Programming_arXiv.pdf`** - Programming protocols for resistive crossbar arrays

### FeFET/Ferroelectric-Specific Architectures
- **`FeFET_Crossbar_Impact_arXiv.pdf`** - Impact of FeFET device parameters on crossbar performance
- **`FeFET_Crossbar_MNIST_Hardware_arXiv.pdf`** - Experimental 96.6% MNIST accuracy on FeFET hardware
- **`fecap_fefet_cim_elements_2024.pdf`** - Ferroelectric capacitor and FeFET elements for CIM
- **`multilevel_fefet_crossbar_2023.pdf`** - Multi-level FeFET crossbar design and characterization
- **`Multi_Level_FeFET_Programming_arXiv.pdf`** - Programming algorithms for multi-level FeFET states
- **`3D_FeFET_Architectures_2025.pdf`** - 3D integrated FeFET architectures for high-density CIM

### Experimental FTJ (Ferroelectric Tunnel Junction) Implementations
- **`FTJ_Crossbar_Experiment_arXiv.pdf`** - Experimental characterization of FTJ crossbar arrays

### Peripheral Circuit Design
- **`Mixed_Signal_DNN_Accelerator_arXiv.pdf`** - Mixed-signal accelerator design with ADC/DAC integration
- **`adc_precision_cim_accuracy_2024.pdf`** - Impact of ADC precision on CIM accuracy
- **`pruning_adc_efficiency_crossbar_2024.pdf`** - Network pruning to reduce ADC requirements
- **`HCiM_ADC_Less_2024.pdf`** - Hybrid CIM architecture without ADCs (digital accumulation)
- **`hcim_adcless_hybrid_cim_2024.pdf`** - ADC-less hybrid CIM for energy efficiency

### Optimization Techniques
- **`memory_tech_crossbar_dnn_accuracy_2024.pdf`** - Memory technology impact on DNN accuracy
- **`simple_packing_algorithm_nvm_2024.pdf`** - Weight packing algorithms for non-volatile memory
- **`Bit_Slicing_Techniques_arXiv_2024.pdf`** - Bit-slicing techniques for CIM accuracy/efficiency
- **`Temperature_Resilient_FeFET_CIM_2024.pdf`** - Temperature compensation for FeFET CIM (-40°C to 125°C)

### Safety-Critical and Specialized Applications
- **`CIM_Safety_Critical_ATRICE_arXiv_2312.01633.pdf`** - CIM for safety-critical automotive applications (ATRICE architecture)
- **`Neuromorphic_Hardware_Vision_Systems_arXiv.pdf`** - Vision-specific neuromorphic hardware
- **`Photonics_Neuromorphic_Computing_arXiv.pdf`** - Photonic approaches to neuromorphic computing
- **`Spiking_Neural_Networks_Hardware_arXiv.pdf`** - Hardware architectures for spiking neural networks

## Key Findings

### Fundamental CIM Principles

**Matrix-Vector Multiplication (MVM) in Crossbars:**
```
I_out = G · V_in
where G is conductance matrix (stored weights)
      V_in is input voltage vector (activations)
      I_out is output current (weighted sum)
```

**Energy Efficiency Sources:**
1. **Parallel computation**: All matrix elements compute simultaneously
2. **Analog domain**: No digital multiply-accumulate operations
3. **In-memory**: Zero data movement between memory and compute
4. **Physical laws**: Computation is "free" (Ohm's law)

### Architecture Design Space

| Architecture | Energy Efficiency | Accuracy | Complexity | Best For |
|--------------|-------------------|----------|------------|----------|
| Pure analog CIM | 1000× | Medium | Low | Edge inference |
| Hybrid CIM (digital accumulation) | 100× | High | Medium | Balanced applications |
| ADC-less CIM | 500× | Medium-High | Low | Energy-constrained |
| 3D stacked CIM | 10000× | High | Very High | Data center inference |

### Critical Design Parameters

**1. Array Size Trade-offs:**
- **Small (32×32)**: Low IR drop, high accuracy, limited capacity
- **Medium (128×128)**: Balanced, suitable for MNIST (FeCIM target)
- **Large (512×512)**: High capacity, significant IR drop (>5% accuracy loss)

**2. ADC Resolution:**
- **4-bit**: 3-5% accuracy loss, 10× energy savings
- **6-bit**: 1-2% accuracy loss, 5× energy savings
- **8-bit**: <0.5% accuracy loss (recommended for FeCIM)

**3. Cell Conductance Levels:**
- **2-4 levels**: Binary/ternary, extreme efficiency, low accuracy
- **8-16 levels**: Good accuracy, manageable peripheral circuits
- **30 levels (FeCIM demo baseline; simulation baseline)**: Near-optimal accuracy, requires 5-bit DAC/8-bit ADC
- **64+ levels**: Marginal accuracy gain, complex peripherals

### Non-Ideality Hierarchy (Impact on Accuracy)

From most to least significant:
1. **ADC quantization** (3-8% loss with 4-bit ADC)
2. **Device variation** (1-3% loss with σ/μ = 10%)
3. **IR drop** (0.5-2% loss for 128×128 array)
4. **Sneak paths** (0.3-1% loss without selectors)
5. **Conductance drift** (0.2-0.5% loss over 10⁶ cycles)

### FeFET-Specific Insights

**Advantages for CIM:**
1. **Non-destructive read**: Unlike FeCAP, FeFET read doesn't disturb state
2. **3-terminal device**: Gate-controlled, easier peripheral integration
3. **CMOS compatibility**: Direct integration with logic transistors
4. **Multi-level capability**: multi-level states (reported) demonstrated in reported in literature devices; demo baseline uses 30 levels (simulation baseline)

**Challenges:**
1. **Read disturb**: Small but non-zero, mitigated by read voltage control
2. **Retention**: 10 years at 85°C (automotive grade achieved)
3. **Variation**: σ/μ = 5-15% (requires variation-aware training)

### Experimental Validation

**Published Hardware Results:**
- **FeFET Crossbar MNIST**: 96.6% accuracy (Nature Commun. 2023)
- **FTJ Reservoir Computing**: 98.24% MNIST accuracy (ScienceDirect 2025)
- **3D FeFET (CEA-Leti)**: 22nm BEOL integration demonstrated
- **Automotive FeFET**: Grade 0 qualification (Fraunhofer IPMS 2024)

### Temperature Resilience

Recent findings (Temperature_Resilient_FeFET_CIM_2024.pdf):
- **-40°C to 125°C** operation with <2% accuracy variation
- **Cryogenic (4K-77K)**: Enhanced Pr (75 µC/cm² at 4K), improved accuracy
- **Compensation circuits**: On-chip temperature sensing + weight adjustment

## Relevance to FeCIM

### Direct Applications

**1. 128×128 Crossbar Design (Module 2)**
- Target array size for MNIST (784 inputs × 128 hidden × 10 outputs)
- IR drop analysis: <1% accuracy impact with proper line sizing
- Sneak path mitigation: Selector-less OK for inference (write uses verify)

**2. 30-State Operation (Module 3)**
- Peripheral circuits: 5-bit DAC input, 8-bit ADC output
- Expected accuracy: 96-98% MNIST (validated by reported in literature 96.6% result)
- Device variation: σ/μ = 10% tolerable with QAT

**3. 3D Integration Roadmap**
- CEA-Leti 22nm BEOL demonstrated → Path to high-density FeCIM
- Target: 8-16 layers for transformer inference (future work)

**4. Automotive and Safety-Critical**
- Temperature resilience (-40°C to 125°C) enables automotive deployment
- Grade 0 qualification achieved (Fraunhofer IPMS)
- ATRICE architecture principles applicable to FeCIM

### System-Level Performance Estimates

**Energy per Inference (MNIST, 128×128 crossbar):**
- **FeCIM (30-state, 8-bit ADC)**: ~10 µJ/inference
- **SRAM-based CIM**: ~1 mJ/inference (100× worse)
- **GPU (NVIDIA T4)**: ~10 mJ/inference (1000× worse)

**Area (28nm CMOS + FeFET BEOL):**
- **Weight storage**: 128×128×30 states (demo baseline; simulation baseline) = 80 kb (FeFET)
- **Peripheral circuits**: ~2× area overhead (DAC/ADC/TIA)
- **Total**: <0.5 mm² (vs. 10 mm² for SRAM equivalent)

**Throughput:**
- **MVM latency**: ~10 ns (analog computation) + 100 ns (ADC) = 110 ns
- **MNIST inference**: ~500 ns (3 layers × 2 MVMs/layer × 110 ns)
- **Throughput**: ~2 M inferences/second

## Related Topics

- **[01-ferroelectric-materials](../01-ferroelectric-materials/)** - Device physics underlying CIM cells
- **[02-training-algorithms](../02-training-algorithms/)** - Training for quantized hardware
- **[03-simulation-tools](../03-simulation-tools/)** - Tools to simulate these architectures
- **[05-neuromorphic](../05-neuromorphic/)** - Alternative computing paradigms
- **[07-memory-architectures](../07-memory-architectures/)** - 3D memory integration
- **[11-peripheral-circuits](../07-memory-architectures/)** - ADC/DAC/TIA design details
- **[15-benchmarking](../08-industry-reports/)** - Performance comparison methodologies

## Implementation Priorities for FeCIM Lattice Tools

**High Priority (Implement Now):**
1. `FeFET_Crossbar_MNIST_Hardware_arXiv.pdf` - Validation of 96.6% target
2. `adc_precision_cim_accuracy_2024.pdf` - ADC bit-depth optimization
3. `Crossbar_Sneak_Path_Analysis_arXiv.pdf` - Sneak path modeling

**Medium Priority (Next Phase):**
4. `3D_FeFET_Architectures_2025.pdf` - Roadmap for scaling beyond MNIST
5. `Temperature_Resilient_FeFET_CIM_2024.pdf` - Automotive deployment
6. `HCiM_ADC_Less_2024.pdf` - Alternative architectures for energy optimization

**Advanced Topics (Future Research):**
7. `multilevel_fefet_crossbar_2023.pdf` - Pushing beyond 30 states
8. `CIM_Safety_Critical_ATRICE_arXiv_2312.01633.pdf` - Safety-critical validation
