# CIM Compilers and Mapping Tools

This directory focuses on software compilation frameworks and mapping algorithms that translate neural network models onto compute-in-memory (CIM) hardware substrates. These tools bridge the gap between high-level ML frameworks (PyTorch, TensorFlow) and low-level hardware constraints.

## Papers in this Directory

### cim_explorer_bnn_tnn_rram_2025.pdf
CIM-Explorer (2025): Design space exploration tool for binary neural networks (BNN) and ternary neural networks (TNN) on RRAM-based CIM accelerators. Includes quantization-aware training and hardware-software co-optimization.

### compass_crossbar_compiler_2025.pdf
COMPASS (2025): Crossbar compiler that automatically maps DNN models to analog CIM arrays while accounting for non-idealities including IR drop, device variation, ADC/DAC precision, and stuck-at faults.

## Key Findings

1. **Quantization-Aware Mapping**: Both tools emphasize that quantization decisions (binary, ternary, or multi-bit) must be co-designed with hardware mapping to achieve optimal accuracy-efficiency trade-offs.

2. **Non-Ideality Compensation**: COMPASS demonstrates automatic insertion of compensation techniques for IR drop and device variation, improving inference accuracy by 5-15% on real hardware.

3. **Design Space Exploration**: CIM-Explorer shows that systematic exploration of configuration spaces (crossbar size, ADC bits, parallelism) can reduce energy by 10× while maintaining accuracy.

4. **Fault Tolerance**: Mapping algorithms must account for stuck-at faults and yield issues, with graceful degradation strategies for partially defective arrays.

5. **Layer-Specific Optimization**: Different DNN layers (convolution, pooling, fully-connected) benefit from different mapping strategies and crossbar configurations.

## Relevance to FeCIM

These compiler tools are directly relevant to the FeCIM Lattice Tools project:

- **Module3 MNIST Implementation**: Our neural network module can benefit from mapping algorithms that optimize layer assignments to 30-level demo baseline ferroelectric cells
- **Crossbar Simulation Validation**: COMPASS IR drop models validate our module2 crossbar non-ideality simulations
- **Quantization Strategy**: CIM-Explorer's quantization methods inform how we leverage the 30-level demo baseline versus binary/ternary alternatives
- **Future Compiler Integration**: These papers provide roadmap for adding automatic model-to-hardware compilation to our toolkit
- **Accuracy Prediction**: Compiler fault models help predict real-world accuracy given manufacturing defects and variability

**Potential Integration**: A future module could integrate COMPASS-like mapping algorithms to automatically optimize DNN deployment on our simulated ferroelectric crossbars.

## Related Topics

- **[06-neural-networks-inference](../10-cim-compilers-mapping/)** - Neural network workloads being mapped by these compilers
- **[02-crossbar-arrays](../04-cim-architectures/)** - Physical crossbar architectures targeted by mapping algorithms
- **[09-reviews-surveys](../09-reviews-surveys/)** - Broader CIM architecture surveys providing context
- **[19-variability-yield](../19-variability-yield/)** - Device variation that compilers must compensate for
- **[04-peripheral-circuits](../07-memory-architectures/)** - ADC/DAC/TIA constraints that affect mapping decisions
