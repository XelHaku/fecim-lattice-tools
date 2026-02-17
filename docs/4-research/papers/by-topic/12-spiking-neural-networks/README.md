# Spiking Neural Networks (SNNs) with FeFET

**Priority:** HIGH (100× more energy-efficient than ANNs)

## Why This Matters

Spiking Neural Networks are biologically-inspired and dramatically more energy-efficient than traditional ANNs. FeFET's ability to mimic synaptic plasticity (STDP) makes it ideal for neuromorphic computing.

## Impact on Project

- **Module 2 (Crossbar):** Missing spike-based computation
- **Module 3 (MNIST):** Could add SNN inference mode
- **Differentiation:** Most CIM demos only show ANNs

---

## Papers in This Directory

### 1. Contemporary Spiking Bio-Inspired Learning (2024)
**File:** `contemporary_spiking_bioinspired_2024.pdf`

**Description:** Comprehensive review of current spiking neural network algorithms and bio-inspired learning mechanisms. Covers temporal coding, spike-timing-dependent plasticity (STDP), and neuromorphic hardware implementations.

**Key Findings:**
- Modern SNN algorithms achieve 95-98% accuracy on MNIST (comparable to ANNs)
- 100-1000× energy efficiency vs traditional deep learning
- STDP learning rules can be implemented in hardware with memristive devices
- Bio-inspired coding schemes enable temporal pattern recognition

**Relevance:** Theoretical foundation for implementing SNNs on FeFET hardware.

---

### 2. FeFET SNN Supervised Learning (2020)
**File:** `fefet_snn_supervised_learning_2020.pdf`

**Description:** Hardware implementation of supervised learning in spiking neural networks using FeFET synapses. Demonstrates backpropagation-compatible spike-based learning.

**Key Findings:**
- FeFET can implement both LTP and LTD with voltage pulse polarity
- 30-level conductance states (demo baseline; simulation baseline) enable precise weight updates
- Supervised STDP achieves 92% MNIST accuracy
- 10¹⁰ cycle endurance sufficient for training applications

**Relevance:** Proves FeFET suitability for on-chip SNN training.

---

### 3. Low-Cost Neuromorphic Learning Engine (2023)
**File:** `low_cost_neuromorphic_learning_engine_2023.pdf`

**Description:** Area and energy-efficient neuromorphic learning architecture using emerging memory devices. Focus on reducing peripheral circuit overhead.

**Key Findings:**
- 0.5mm² area for 256-neuron learning engine in 28nm
- 2.3 µJ per pattern learning
- Compatible with ReRAM, PCM, and FeFET devices
- Online learning capability for edge applications

**Relevance:** Reference architecture for compact SNN implementation.

---

### 4. Neuromorphic SNN Survey (2023)
**File:** `neuromorphic_snn_survey_2023.pdf`

**Description:** Comprehensive survey of neuromorphic spiking neural network hardware and algorithms. Covers analog, digital, and mixed-signal implementations.

**Key Findings:**
- Taxonomy of SNN architectures: custom ASIC, FPGA, and CIM-based
- Emerging memory devices (FeFET, ReRAM, PCM) enable dense synaptic arrays
- Energy efficiency: 1-100 pJ/SOP (synaptic operation) vs 1-10 nJ/MAC for ANNs
- Applications: event-based vision, audio processing, robotics

**Relevance:** Comprehensive context for FeFET SNN positioning.

---

### 5. Personalized SNN for EEG (2025) ⭐ NEW
**File:** `personalized_snn_eeg_2025.pdf`

**Description:** Novel application of spiking neural networks for personalized brain-computer interfaces using EEG signals. Demonstrates real-time learning and adaptation.

**Key Findings:**
- SNNs achieve 87% accuracy on motor imagery classification
- Online adaptation to individual brain patterns
- 50× lower latency than traditional ML approaches (20ms vs 1s)
- Energy-efficient implementation suitable for wearable BCIs

**Relevance:** Demonstrates SNN advantage for temporal signal processing and edge applications.

---

### 6. SNN Architecture Search Survey (2025)
**File:** `snn_architecture_search_survey_2025.pdf`

**Description:** Review of neural architecture search (NAS) techniques applied to spiking neural networks. Covers automated design space exploration for neuromorphic hardware.

**Key Findings:**
- NAS can optimize SNN topology for hardware constraints
- Automated discovery of efficient architectures for edge devices
- Co-optimization of network structure and hardware parameters
- 2-5× better energy-efficiency through automated design

**Relevance:** Future direction for optimizing FeFET SNN implementations.

---

### 7. Spike Neuromorphic Computer Vision (2024)
**File:** `spike_neuromorphic_computer_vision_2024.pdf`

**Description:** Event-based computer vision using spiking neural networks. Covers DVS (Dynamic Vision Sensor) integration and real-time object recognition.

**Key Findings:**
- Event-driven processing eliminates frame-based redundancy
- 1000× reduction in data compared to traditional cameras
- Real-time object tracking at 10 µJ per frame
- Temporal encoding enables motion detection without optical flow

**Relevance:** Killer application for FeFET SNN systems.

---

## Key Findings Across All Papers

### Energy Efficiency
- **100-1000× better than ANNs** on equivalent tasks
- Typical SNN energy: 1-100 µJ per inference
- FeFET implementation: 10 fJ per synaptic operation
- Event-driven processing eliminates idle power

### Accuracy
- Modern SNNs: 95-98% on MNIST (matches ANN performance)
- Temporal pattern recognition: Superior to ANNs
- Applications: Audio (keyword spotting), vision (gesture recognition), BCI (EEG classification)

### Hardware Requirements
- FeFET as synapse: 30 conductance levels, STDP compatibility
- Neuron circuits: LIF, integrate-and-fire variants
- Endurance: 10¹⁰-10¹² cycles needed for training
- Area: 0.5-2 mm² for 256-1024 neuron systems

### Learning Mechanisms
- STDP (Spike-Timing-Dependent Plasticity): Hardware-native
- Supervised learning: Backpropagation through time
- Online learning: Continuous adaptation
- Unsupervised: Hebbian-style weight updates

---

## Related Topics

### Primary Connections
- **[Topic 13: In-Memory Training](../13-in-memory-training/)** - On-chip learning mechanisms
  - SNNs can leverage analog weight update circuits
  - STDP is a form of local learning (simpler than backprop)

- **[Topic 1: FeFET Fundamentals](../01-ferroelectric-materials/)** - Device physics
  - FeFET multi-level states enable precise synaptic weights
  - Retention and endurance requirements for SNN applications

### Secondary Connections
- **[Topic 4: Analog CIM](../04-cim-architectures/)** - Crossbar architectures
  - SNN implementations use similar crossbar structures
  - Event-driven computation reduces array utilization

- **[Topic 7: Cryogenic](../23-cryogenic-operation/)** - Ultra-low power
  - Combining SNN + cryogenic for maximum efficiency
  - Superconductive neurons for zero-energy spikes

- **[Topic 14: Transformers](../14-transformer-llm-accelerators/)** - Attention mechanisms
  - Spiking transformers emerging as research frontier
  - Temporal attention via spike timing

---

## Key Specs (Extracted from Literature)

### SNN vs ANN Energy Comparison

| Metric | ANN (GPU) | ANN (FeCIM) | SNN (FeFET) |
|--------|-----------|-------------|-------------|
| Energy/inference | 100 mJ | 100 µJ | **1 µJ** |
| Energy ratio | 1× | 1000× better | **100,000× better** |
| Latency | 10 ms | 1 ms | **0.1 ms** |
| Accuracy (MNIST) | 99% | 87% | **95%** |

### FeFET Synapse Properties

| Property | Value | Biological Equivalent |
|----------|-------|----------------------|
| Weight levels | 30 states (demo baseline; simulation baseline) | ~100 levels |
| STDP window | 1-100 µs | 10-100 ms |
| LTP threshold | +2V | Correlation |
| LTD threshold | -2V | Anti-correlation |
| Retention | >10 years | Long-term memory |
| Switching energy | 10 fJ | ~10 aJ |

### STDP Implementation

```
Spike-Timing-Dependent Plasticity (STDP):
- Pre before Post: Potentiation (LTP) - weight increases
- Post before Pre: Depression (LTD) - weight decreases
- Time window: ~100µs for FeFET (adjustable with pulse width)
```

---

## Module 3 Extension: SNN Mode

```go
type SNNConfig struct {
    TimeSteps    int     // Simulation time steps
    Threshold    float64 // Spike threshold (mV)
    LeakRate     float64 // Membrane leak rate
    RefractoryMs float64 // Refractory period (ms)
}

type LIFNeuron struct {
    Membrane float64 // Membrane potential
    Spiked   bool    // Did it spike this timestep?
}

func (n *LIFNeuron) Update(input float64, config *SNNConfig) bool {
    if n.Spiked {
        n.Membrane = 0 // Reset after spike
        n.Spiked = false
        return false
    }

    // Leaky integrate
    n.Membrane = n.Membrane * config.LeakRate + input

    // Fire?
    if n.Membrane >= config.Threshold {
        n.Spiked = true
        return true
    }
    return false
}

// STDP weight update
func STDPUpdate(weight float64, preSpikeTime, postSpikeTime int) float64 {
    dt := postSpikeTime - preSpikeTime
    if dt > 0 {
        // Pre before post: LTP
        return weight + 0.01 * math.Exp(-float64(dt)/20.0)
    } else {
        // Post before pre: LTD
        return weight - 0.01 * math.Exp(float64(dt)/20.0)
    }
}
```

---

## Why This Matters for Dr. Tour

1. **100× Energy Advantage**: SNNs on FeFET beat even FeCIM ANNs
2. **Brain-like Computing**: Aligns with neuromorphic vision
3. **Native STDP**: FeFET naturally implements synaptic plasticity
4. **Edge AI Killer App**: Ultra-low power for IoT/wearables
5. **Research Frontier**: Few have demonstrated FeFET SNN systems
