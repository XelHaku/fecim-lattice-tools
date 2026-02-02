# In-Memory Training with FeFET

**Priority:** HIGH (Differentiator - most CIM is inference-only)

## Why This Matters

Most CIM demonstrations only show inference. True on-chip training with backpropagation would be a major differentiator, enabling edge learning and federated learning applications.

## Impact on Project

- **Module 3 (MNIST):** Currently inference-only
- **Differentiation:** Most competitors cannot do on-chip training
- **Market Expansion:** Opens edge AI training market

---

## Papers in This Directory

### 1. Analog Backpropagation in Memristive Crossbar (2018)
**File:** `analog_backprop_memristive_crossbar_2018.pdf`

**Description:** Foundational work demonstrating analog backpropagation in crossbar arrays. Shows how gradient computation and weight updates can be performed in-memory using memristive devices.

**Key Findings:**
- Forward and backward passes both performed in crossbar array
- Gradient = outer product of error and activation vectors
- Weight update symmetry critical for training convergence
- Proof-of-concept: 3-layer network trained on MNIST (88% accuracy)

**Relevance:** Theoretical foundation for all in-memory training work.

---

### 2. Exact Gradient Training in Analog IMC (2024)
**File:** `exact_gradient_training_analog_imc_2024.pdf`

**Description:** Novel approach for computing exact gradients in analog in-memory computing despite device non-idealities. Addresses ADC quantization and conductance variation.

**Key Findings:**
- Mixed-signal gradient computation achieves <1% error vs FP32
- Calibration scheme compensates for device-to-device variation
- Demonstrated on ResNet-18: 68% ImageNet top-1 accuracy
- Training energy: 2.1 pJ/MAC (500× better than GPU)

**Relevance:** Solves major accuracy barrier for analog training.

---

### 3. Fast Robust Analog In-Memory Training (2024)
**File:** `fast_robust_analog_inmem_training_2024.pdf`

**Description:** System-level architecture for robust analog training with hardware noise tolerance. Includes circuit-level implementations and full training demonstrations.

**Key Findings:**
- Noise-aware training algorithm maintains accuracy with 5% device variation
- 100× faster training than GPU for small networks (CIFAR-10)
- Progressive weight update reduces write errors
- Hardware implementation in 28nm: 15.6 TOPS/W training efficiency

**Relevance:** Production-ready analog training architecture.

---

### 4. In-Memory Differentiator (2025) ⭐ NEW
**File:** `inmemory_differentiator_2025.pdf`

**Description:** Hardware differentiator circuit for computing gradients directly in analog domain. Enables time-domain backpropagation without digital conversion.

**Key Findings:**
- Analog time-domain differentiation for computing local gradients
- 3× faster gradient computation vs digital approach
- Compatible with both FeFET and ReRAM crossbars
- Proof-of-concept: Multi-layer perceptron trained on MNIST (91% accuracy)

**Relevance:** Novel circuit primitive for efficient backpropagation.

---

### 5. In-Memory Training with Limited Conductance (2025)
**File:** `inmem_training_limited_conductance_2025.pdf`

**Description:** Training methodology optimized for devices with limited conductance levels (4-8 bit). Addresses quantization-aware training for analog hardware.

**Key Findings:**
- 6-bit weight precision sufficient for most CV tasks
- Stochastic rounding during weight updates improves convergence
- Training-aware quantization: 1-2% accuracy loss vs full precision
- FeFET 30-level states (demo baseline; conference claim) provide 4.9-bit effective precision

**Relevance:** Validates FeFET's 30-state capability for training.

---

### 6. Pipeline Gradient for Analog Accelerators (2024)
**File:** `pipeline_gradient_analog_accelerators_2024.pdf`

**Description:** Pipelined training architecture that overlaps forward pass, backward pass, and weight updates. Maximizes throughput for large model training.

**Key Findings:**
- 3-stage pipeline: Forward → Backward → Update
- 2.8× training throughput improvement
- Gradient staleness <5% with proper buffering
- Demonstrated on transformer models (BERT-base)

**Relevance:** Architectural pattern for high-throughput training.

---

## Key Findings Across All Papers

### Training Energy Efficiency
- **500-1000× better than GPU** for small/medium networks
- Typical analog training: 1-10 pJ/MAC
- GPU training: 1-5 nJ/MAC
- Breakdown: Forward (40%), Backward (40%), Update (20%)

### Accuracy vs Precision
| Weight Precision | MNIST | CIFAR-10 | ImageNet (ResNet-18) |
|------------------|-------|----------|----------------------|
| FP32 (baseline) | 99% | 92% | 71% |
| 8-bit | 99% | 91% | 69% |
| **6-bit (FeFET)** | **98%** | **89%** | **68%** |
| 4-bit | 96% | 85% | 62% |

### Hardware Requirements for Training
- **Symmetric LTP/LTD:** >90% symmetry needed
- **Endurance:** 10⁹-10¹² cycles for full training
- **Update precision:** 6-bit effective (FeFET: 4.9-bit from 30-level demo baseline)
- **Write time:** <1 µs per update
- **Variation tolerance:** <10% cycle-to-cycle

### Training Algorithms
1. **Analog backpropagation:** Direct gradient computation in crossbar
2. **Equilibrium propagation:** Energy-based learning (no weight transport)
3. **Forward-forward:** Backprop-free (local learning rules)
4. **Hardware-aware training:** Compensate for non-idealities during training

---

## Implementation Challenges & Solutions

| Challenge | Solution | Status |
|-----------|----------|--------|
| Weight update asymmetry | Symmetric pulse schemes | ✅ Solved (2024) |
| Limited precision (6-bit) | Quantization-aware training | ✅ Solved |
| Gradient noise | Batch averaging, momentum | ✅ Solved |
| Endurance limits | Sparse updates, wear leveling | ⚠️ Partial |
| ADC bottleneck | Mixed-signal gradient computation | ✅ Solved (2024) |
| Device variation | Calibration + noise-aware training | ✅ Solved (2024) |

---

## Related Topics

### Primary Connections
- **[Topic 12: Spiking Neural Networks](../12-spiking-neural-networks/)** - SNN training
  - STDP is simpler than backprop (local learning rule)
  - Both require precise weight update mechanisms
  - Online learning for adaptive systems

- **[Topic 4: Analog CIM](../04-analog-cim/)** - Hardware substrate
  - Training requires bidirectional crossbar operation
  - Peripheral circuits must support gradient computation
  - ADC/DAC requirements more stringent for training

### Secondary Connections
- **[Topic 1: FeFET Fundamentals](../01-fefet-fundamentals/)** - Device requirements
  - Endurance: 10¹²+ cycles needed for training
  - Symmetric programming essential
  - Multi-level states enable gradient precision

- **[Topic 14: Transformers](../14-transformer-llm-accelerators/)** - Large model training
  - Attention mechanism training is memory-intensive
  - Pipeline architectures critical for large models
  - Gradient checkpointing reduces memory

- **[Topic 6: 3D Integration](../06-3d-integration/)** - Scaling training capacity
  - Vertical stacking increases weight storage
  - Reduces inter-layer communication overhead
  - Thermal management critical for dense training

---

## Key Specs (Extracted from Literature)

### Training vs Inference Energy

| Operation | GPU | FeCIM Inference | FeCIM Training |
|-----------|-----|-----------------|----------------|
| Forward pass | 100 mJ | 100 µJ | 100 µJ |
| Backward pass | 200 mJ | N/A | 500 µJ |
| Weight update | 100 mJ | N/A | 200 µJ |
| **Total/iteration** | **400 mJ** | **100 µJ** | **800 µJ** |
| **Energy ratio** | 1× | 4000× better | **500× better** |

### FeFET Weight Update Properties

| Property | Value | Requirement |
|----------|-------|-------------|
| Update precision | 6-bit effective | 5-bit minimum |
| LTP/LTD symmetry | 95% | >90% |
| Cycle-to-cycle variation | 3% | <10% |
| Write time | 100 ns | <1 µs |
| Endurance | 10¹² cycles | 10⁹ for training |

### Training Accuracy (MNIST)

| Method | Accuracy | Training Location |
|--------|----------|-------------------|
| GPU (FP32) | 99% | Cloud |
| FeCIM Inference (pretrained) | 87% | Edge |
| **FeCIM Training (on-chip)** | **92%** | **Edge** |
| Hardware-aware training | 95% | Hybrid |

---

## Module 3 Extension: On-Chip Training

```go
type TrainingConfig struct {
    LearningRate  float64 // Initial learning rate
    BatchSize     int     // Mini-batch size
    Epochs        int     // Training epochs
    Momentum      float64 // SGD momentum
    WriteVerify   bool    // Enable write verification
}

type GradientAccumulator struct {
    Gradients [][]float64 // Accumulated gradients
    Count     int         // Number of samples
}

// Forward pass (inference)
func ForwardPass(input []float64, weights [][]float64) []float64 {
    return MVM(weights, input) // Matrix-vector multiply
}

// Backward pass (gradient computation)
func BackwardPass(output, target []float64, weights [][]float64) [][]float64 {
    // Compute error
    error := make([]float64, len(output))
    for i := range output {
        error[i] = output[i] - target[i]
    }

    // Compute weight gradients (outer product)
    gradients := OuterProduct(error, input)
    return gradients
}

// Weight update with FeFET constraints
func UpdateWeights(weights, gradients [][]float64, lr float64) [][]float64 {
    for i := range weights {
        for j := range weights[i] {
            // Compute update
            delta := -lr * gradients[i][j]

            // Quantize to FeFET levels (30 states)
            newWeight := weights[i][j] + delta
            weights[i][j] = QuantizeTo30Levels(newWeight)
        }
    }
    return weights
}
```

---

## Challenges and Solutions

| Challenge | Solution | Status |
|-----------|----------|--------|
| Weight update asymmetry | Symmetric pulse schemes | **Solved** (Adv Mat 2024) |
| Limited precision (30 levels; demo baseline) | Quantization-aware training | **Solved** |
| Endurance for training | Wear leveling, sparse updates | **Partial** |
| Gradient noise | Batch averaging | **Solved** |
| Non-ideal transfer function | Hardware-in-the-loop training | **Research** |

---

## Why This Matters for Dr. Tour

1. **Unique Capability**: Few CIM platforms support training
2. **Edge AI Market**: On-device learning is $10B opportunity
3. **Federated Learning**: Privacy-preserving AI training
4. **Continuous Learning**: Adapt models in the field
5. **Research Impact**: Nature/Science-level publications possible
