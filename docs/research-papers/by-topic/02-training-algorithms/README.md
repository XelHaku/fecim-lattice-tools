# Training Algorithms for Analog AI Hardware

## Overview

This directory contains research on training algorithms specifically designed for analog AI hardware, with emphasis on quantization-aware training (QAT), low-precision neural networks, and hardware-software co-design approaches. These techniques are critical for deploying neural networks on analog compute-in-memory (CIM) systems where continuous weight values must be discretized to finite analog states, and where hardware non-idealities affect accuracy.

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

## Papers in this Directory

### Quantization-Aware Training
- **`Quantization_Aware_Training_arXiv.pdf`** - Foundational QAT techniques for training networks robust to weight quantization
- **`quantization_aware_training_survey_2023.pdf`** - Comprehensive 2023 survey of QAT methods and best practices
- **`Advanced_Quantization_Training_arXiv.pdf`** - Advanced techniques including mixed-precision and gradient approximation methods
- **`low_bit_quantization_neural_nets_2022.pdf`** - Specialized methods for extreme low-bit (2-4 bit) quantization regimes

### Low-Precision Neural Networks
- **`Low_Precision_Neural_Networks_arXiv.pdf`** - General survey of low-precision NN architectures and training
- **`Binarized_Neural_Networks_Hardware_arXiv.pdf`** - Binary (1-bit) neural networks for extreme hardware efficiency
- **`Ternary_Neural_Networks_Hardware_arXiv.pdf`** - Ternary (2-bit) networks balancing accuracy and efficiency

### Hardware-Aware Training
- **`Analog_AI_Hardware_Codesign_arXiv.pdf`** - Co-design methodologies integrating hardware constraints into training
- **`aimc_accuracy_post_training_2024.pdf`** - Post-training techniques for improving accuracy on analog IMC hardware
- **`tiki_taka_analog_training_2024.pdf`** - Tiki-Taka algorithm for in-situ training on analog hardware with asymmetric updates
- **`Variation_Resilient_FeFET_BNN_MNIST_2024.pdf`** - Training methods resilient to FeFET device variation for binary neural networks

## Key Findings

### Quantization Strategies
1. **QAT is essential for analog hardware**: Networks trained with quantization awareness achieve 2-5% higher accuracy than post-training quantization when deployed on analog systems
2. **Straight-through estimators** enable gradient flow through discrete quantization operations during backpropagation
3. **Mixed-precision quantization** allocates more bits to sensitive layers (typically first/last layers) while using extreme quantization elsewhere

### Hardware-Specific Considerations
1. **Device variation modeling**: Training must account for cycle-to-cycle and device-to-device variation (σ/μ typically 5-15% for ferroelectric devices)
2. **Asymmetric updates**: Tiki-Taka and similar algorithms handle asymmetric potentiation/depression characteristics of analog devices
3. **Conductance drift**: Post-training calibration can compensate for time-dependent conductance drift without retraining

### Bit-Depth vs. Accuracy Trade-offs
- **8-bit**: Near full-precision accuracy (<0.5% degradation) for most networks
- **4-bit**: 1-3% accuracy loss with QAT, suitable for FeCIM's 30-state operation (~4.9 bits)
- **2-bit (ternary)**: 3-7% accuracy loss, requires architectural adaptation
- **1-bit (binary)**: 5-15% accuracy loss, best for edge inference applications

### FeCIM-Specific Insights
1. **30 analog states (4.9 bits; demo baseline, simulation baseline)** exceeds requirements for most QAT algorithms designed for 4-bit systems
2. **High endurance (10⁹-10¹² cycles)** enables true in-situ training, not just inference
3. **Variation resilience** in FeFET BNNs suggests robust operation even with device-to-device variation

## Relevance to FeCIM

### Direct Applications
1. **30-state quantization**: QAT methods can be adapted to quantize weights to 30 discrete ferroelectric polarization levels
2. **Training pipeline**: Implement QAT during software training → map trained weights to FeCIM conductance states
3. **In-situ learning**: High endurance enables on-chip learning using Tiki-Taka or similar analog-aware algorithms

### Implementation in FeCIM Lattice Tools
- **Module 3 (MNIST)**: Apply QAT to achieve 96-98% accuracy with 30-level quantization
- **Crossbar simulation**: Model quantization effects, device variation, and drift during inference
- **Training workflow**: Provide QAT utilities to generate FeCIM-optimized weight mappings

### Performance Expectations
Based on research literature:
- **30-level quantization**: Expected <1% accuracy degradation vs. full precision (well within 4-bit QAT performance)
- **With device variation (σ/μ = 10%)**: Additional 1-2% degradation, mitigated by variation-aware training
- **Long-term drift**: Periodic recalibration every 10⁴-10⁶ inference cycles recommended

## Related Topics

- **[01-ferroelectric-materials](../01-ferroelectric-materials/)** - Physical device characteristics (Pr, Ec, endurance) that constrain training
- **[03-simulation-tools](../03-simulation-tools/)** - Tools like AIHWKit and CrossSim that implement these algorithms
- **[04-cim-architectures](../04-cim-architectures/)** - Hardware architectures these algorithms target
- **[05-neuromorphic](../05-neuromorphic/)** - Spiking neural network training with similar constraints
- **[08-hardware-measurements](../20-manufacturing-integration/)** - Validation of trained networks on actual hardware

## References for FeCIM Development

**High Priority for Implementation:**
1. `quantization_aware_training_survey_2023.pdf` - Comprehensive methodology guide
2. `aimc_accuracy_post_training_2024.pdf` - Practical deployment techniques
3. `Variation_Resilient_FeFET_BNN_MNIST_2024.pdf` - FeFET-specific training insights

**For Advanced Features:**
4. `tiki_taka_analog_training_2024.pdf` - In-situ training on analog hardware
5. `Analog_AI_Hardware_Codesign_arXiv.pdf` - Co-design optimization strategies
