# Ferroelectric CIM Visualization Project Roadmap

## Project Overview

This project provides educational demonstrations and simulations for Ferroelectric CIM ferroelectric compute-in-memory (CIM) technology, based on HfO₂/ZrO₂ superlattice materials developed at external research institution.

---

## Current State (January 2026)

### Completed Components

#### Demo 1: Hysteresis Visualization
- **Purpose:** Interactive P-E hysteresis curve demonstration
- **Features:**
  - Preisach model simulation
  - GPU-accelerated rendering (Vulkan shaders)
  - Real-time parameter adjustment
- **Files:**
  - `demo1-hysteresis/cmd/hysteresis/main.go`
  - `demo1-hysteresis/pkg/ferroelectric/preisach.go`
  - `demo1-hysteresis/pkg/ferroelectric/material.go`
  - `demo1-hysteresis/pkg/render/render.go`
  - `demo1-hysteresis/pkg/render/plot.go`
  - `demo1-hysteresis/shaders/cell.vert`, `cell.frag`
  - `demo1-hysteresis/shaders/hysteresis.vert`, `hysteresis.frag`

#### Demo 2: Neural Network Inference
- **Purpose:** Demonstrate CIM-based neural network inference
- **Features:**
  - Crossbar array simulation with noise/quantization
  - MNIST dataset loading
  - Accuracy evaluation metrics
  - Heatmap visualization
  - Weight import/export utilities
  - Hardware-aware training simulation
  - Normalization layers (BatchNorm, LayerNorm, RMSNorm)
  - Convolutional layers with im2col mapping
- **Files:**
  - `demo2-inference/cmd/inference/main.go`
  - `demo2-inference/pkg/crossbar/array.go`
  - `demo2-inference/pkg/network/network.go`
  - `demo2-inference/pkg/data/mnist.go`
  - `demo2-inference/pkg/evaluation/accuracy.go`
  - `demo2-inference/pkg/visualization/heatmap.go`
  - `demo2-inference/pkg/weights/weights.go`
  - `demo2-inference/pkg/training/training.go`
  - `demo2-inference/pkg/layers/normalization.go`
  - `demo2-inference/pkg/layers/convolution.go`
  - `demo2-inference/shaders/mvm.comp`
  - `demo2-inference/shaders/activation.comp`

#### Documentation
- **RESEARCH_LOG.md:** 58 sections, ~3,300 lines of research documentation
- **CURRICULUM.md:** 9-area comprehensive curriculum
- **DOWNLOAD_PLAN.md:** Priority paper download list

---

## Development Phases

### Phase 1: Foundation (Completed)
- [x] Project structure setup
- [x] Basic Preisach model implementation
- [x] Crossbar array simulation
- [x] MNIST data loading
- [x] Accuracy metrics

### Phase 2: Visualization (Completed)
- [x] P-E hysteresis curve plotting
- [x] Crossbar heatmap visualization
- [x] GLSL/Vulkan shader framework
- [x] ASCII fallback rendering

### Phase 3: Training & Inference (Completed)
- [x] Weight import/export (JSON, binary)
- [x] Hardware-aware training utilities
- [x] MLC programming simulation
- [x] Quantization support
- [x] Normalization layer fusion

### Phase 4: Advanced Layers (In Progress)
- [x] Batch normalization
- [x] Layer normalization (transformers)
- [x] RMS normalization (LLMs)
- [x] 2D convolution with im2col
- [x] Pooling layers
- [x] Depthwise separable convolution
- [x] Attention mechanism (multi-head, sliding window, crossbar-optimized)
- [x] Regularization (dropout, noise injection, weight decay)
- [x] Transformer blocks (encoder, FFN, positional encoding)
- [x] Reservoir computing layer
- [x] Weight pruning utilities (magnitude, structured, block, SWS)
- [x] Model serialization (JSON, binary, quantization)

### Phase 5: GPU Acceleration (Planned)
- [ ] Vulkan compute pipeline integration
- [ ] GPU-accelerated MVM
- [ ] Real-time inference visualization
- [ ] Multi-crossbar orchestration

### Phase 6: Advanced Simulations (Planned)
- [ ] Spiking neural network simulation
- [ ] Negative capacitance modeling
- [ ] Antiferroelectric neuron simulation
- [ ] 3D FeNAND visualization

---

## Technical Specifications

### Crossbar Simulation Parameters

| Parameter | Default | Range |
|-----------|---------|-------|
| Array size | 64×64 | 16-256 |
| ADC bits | 6 | 4-8 |
| DAC bits | 8 | 4-8 |
| Noise level | 2% | 0-10% |
| Weight bits | 6 | 2-8 |

### Supported Models

| Model Type | Status | Crossbar Mapping |
|------------|--------|------------------|
| MLP | ✅ Ready | Direct |
| CNN | ✅ Ready | Im2col |
| BNN | 🔄 Partial | Direct |
| SNN | 📋 Planned | Event-driven |
| Transformer | 📋 Planned | Attention blocks |

### Key Performance Targets

Based on literature:
- **Accuracy:** >95% MNIST at 6-bit weights
- **Energy:** <100 fJ/MAC (simulated)
- **Throughput:** >100 TOPS/W (theoretical)

---

## Research Integration

### Ferroelectric CIM Technology Mapping

| Ferroelectric CIM Feature | Demo Component |
|---------------------|----------------|
| HZO superlattice | Material parameters in preisach.go |
| Analog CIM | Crossbar array simulation |
| Multi-level cell | MLC programmer in training.go |
| STDP learning | Planned for SNN demo |
| Flash Joule synthesis | Documentation only |

### Key Papers Referenced

1. **HfO₂/ZrO₂ Superlattice FTJ** (Materials Horizons 2024)
2. **MLC FeFET Crossbar** (Nature Communications 2023)
3. **AIMC Attention** (Nature Computational Science 2025)
4. **Flash In₂Se₃ Synthesis** (Adv. Electronic Materials 2025)
5. **All-Ferroelectric SNN** (Advanced Science 2024)

---

## File Structure

```
multilayer-ferroelectric-cim-visualizer/
├── demo1-hysteresis/
│   ├── cmd/hysteresis/main.go
│   ├── pkg/
│   │   ├── ferroelectric/
│   │   │   ├── material.go
│   │   │   └── preisach.go
│   │   ├── render/
│   │   │   ├── render.go
│   │   │   └── plot.go
│   │   └── simulation/
│   │       └── engine.go
│   └── shaders/
│       ├── cell.vert, cell.frag
│       └── hysteresis.vert, hysteresis.frag
│
├── demo2-inference/
│   ├── cmd/inference/main.go
│   ├── pkg/
│   │   ├── crossbar/array.go
│   │   ├── network/network.go
│   │   ├── data/mnist.go
│   │   ├── evaluation/accuracy.go
│   │   ├── visualization/heatmap.go
│   │   ├── weights/
│   │   │   ├── weights.go
│   │   │   └── serialization.go
│   │   ├── training/training.go
│   │   └── layers/
│   │       ├── normalization.go
│   │       ├── convolution.go
│   │       ├── attention.go
│   │       ├── regularization.go
│   │       ├── transformer.go
│   │       ├── pruning.go
│   │       ├── activations.go
│   │       ├── recurrent.go
│   │       ├── loss.go
│   │       ├── optimizer.go
│   │       ├── init.go
│   │       ├── augmentation.go
│   │       ├── batch.go
│   │       ├── checkpoint.go
│   │       ├── export.go
│   │       ├── quantization.go
│   │       ├── pipeline.go
│   │       ├── sparse.go
│   │       ├── snn.go
│   │       ├── tiling.go
│   │       ├── compression.go
│   │       ├── orchestration.go
│   │       ├── training_loop.go
│   │       ├── estimation.go
│   │       ├── calibration.go
│   │       ├── profiling.go
│   │       ├── visualization.go
│   │       ├── benchmark.go
│   │       ├── photonic.go
│   │       ├── attention_cim.go
│   │       ├── domain_wall.go
│   │       ├── continual_learning.go
│   │       ├── in_sensor.go
│   │       ├── audio_distillation.go
│   │       ├── olfaction_nas.go
│   │       ├── tactile_federated.go
│   │       ├── vision_security.go
│   │       ├── multimodal_hil.go
│   │       ├── error_cosim.go
│   │       ├── learning_thermal.go
│   │       ├── power_3d.go
│   │       ├── memory_latency.go
│   │       ├── neuromorphic_devices.go
│   │       ├── llm_reliability.go
│   │       ├── diffusion_cryogenic.go
│   │       ├── vision_interconnect.go
│   │       ├── optical_hbm4.go
│   │       ├── noise_adc.go
│   │       ├── thermal_nmp.go
│   │       ├── ecc_power.go
│   │       ├── audio_cnn.go
│   │       ├── gnn_nas.go
│   │       ├── rl_multichip.go
│   │       ├── spiking_diffusion_ecc.go
│   │       ├── optical_clustering.go
│   │       ├── bnn_onnx.go
│   │       ├── tnn_thermal.go
│   │       ├── mixedprec_tiling.go
│   │       ├── inmem_training.go
│   │       ├── continual_distill.go
│   │       ├── federated_adversarial.go
│   │       ├── nas_reinforcement.go
│   │       ├── snn_gnn.go
│   │       ├── neuromorphic_transformer.go
│   │       ├── cosim_deploy.go
│   │       ├── model_zoo_demo.go
│   │       ├── visualization_edge.go
│   │       ├── hil_multichip.go
│   │       ├── ferro3d_snn.go
│   │       ├── kvcache_fjh.go
│   │       ├── reservoir_insensor.go
│   │       ├── optical_benchmark.go
│   │       ├── ftj_ndr.go
│   │       ├── domain_fatigue.go
│   │       ├── memcapacitor_hdc.go
│   │       ├── stochastic_hybrid.go
│   │       ├── nc_2dferro.go
│   │       ├── cam_attention.go
│   │       ├── snn_mixedprec.go
│   │       ├── edge_optical.go
│   │       ├── federated_vision.go
│   │       ├── hil_cam.go
│   │       ├── cryo_3dfenand.go
│   │       ├── beol_endurance.go
│   │       ├── memcap_stochastic.go
│   │       ├── plasticity_sensorfusion.go
│   │       ├── domainwall_hdc.go
│   │       ├── cam_compiler.go
│   │       ├── ftj_onchip.go
│   │       ├── writeverify_a2s.go
│   │       ├── interconnect_ecc.go
│   │       ├── ncfet_hybrid.go
│   │       ├── thermal_spintronics.go
│   │       ├── optical_compiler.go
│   │       ├── training_attention.go
│   │       ├── radiation_adc.go
│   │       ├── security_mixedsignal.go
│   │       ├── sensory_compiler.go
│   │       ├── gnn_endurance.go
│   │       ├── recsys_activation.go
│   │       ├── testing_detection.go
│   │       ├── materials_endurance.go
│   │       ├── calibration_transformer_cim.go
│   │       ├── power_spiking_cim.go
│   │       ├── security_hybrid_cim.go
│   │       ├── edge_spike_cim.go
│   │       ├── reservoir_multidie_cim.go
│   │       ├── optical_inmem_training.go
│   │       ├── gnn_mram_pcm.go
│   │       ├── recsys_vision_cim.go
│   │       ├── security_ftj_cim.go
│   │       ├── cryo_reram_cim.go
│   │       ├── transformer_adc_cim.go
│   │       ├── compiler_power_cim.go
│   │       ├── pvt_edge_cim.go
│   │       ├── neuromorphic_reliability_cim.go
│   │       ├── onchip_security_cim.go
│   │       ├── multimodal_gnn_cim.go
│   │       ├── emerging_analog_cim.go
│   │       ├── federated_recsys_cim.go
│   │       ├── vision_compiler_cim.go
│   │       ├── hetero_llm_cim.go
│   │       ├── radiation_ecc_cim.go
│   │       ├── scheduling_diffusion_cim.go
│   │       ├── cryo_olfactory_cim.go
│   │       ├── rl_3dchiplet_cim.go
│   │       ├── photonic_gnn_cim.go
│   │       ├── timeseries_hdc_cim.go
│   │       ├── autonomous_quantum_cim.go
│   │       ├── edge_neuromorphic_cim.go
│   │       ├── compiler_codesign_cim.go
│   │       ├── reliability_power_cim.go
│   │       ├── deployment_testing_cim.go
│   │       ├── security_multichip_cim.go
│   │       ├── application_insitu_cim.go
│   │       ├── emerging_neuromorphic_cim.go
│   │       ├── edge_codesign_cim.go
│   │       ├── energy_benchmark_cim.go
│   │       ├── scientific_recsys_cim.go
│   │       ├── hybrid_reliability_cim.go
│   │       ├── llm_3d_cim.go
│   │       ├── neuromorphic_mram_cim.go
│   │       └── compiler_onchip_cim.go
│   └── shaders/
│       ├── mvm.comp
│       └── activation.comp
│
├── docs/
│   ├── RESEARCH_LOG.md
│   ├── CURRICULUM.md
│   ├── PROJECT_ROADMAP.md
│   ├── DEMO_IMPLEMENTATION.md
│   ├── HZO_PARAMETERS.md
│   └── IRONLATTICE_PARADIGM.md
│
└── papers/
    └── DOWNLOAD_PLAN.md
```

---

## Next Steps (Priority Order)

### Immediate
1. ~~Implement transformer block (attention + FFN + residual connections)~~ ✅ Done
2. Complete Vulkan compute pipeline for GPU-accelerated MVM
3. End-to-end CNN inference demo with visualization

### Short-term
4. ~~Weight pruning for sparse crossbar mapping~~ ✅ Done
5. Spiking neural network simulation
6. Interactive web-based demo (WebGPU)
7. Pre-trained model zoo (MNIST, CIFAR-10)
8. ~~Recurrent layers (LSTM, GRU) for time-series~~ ✅ Done

### Long-term
9. Negative capacitance device simulation
10. 3D FeNAND array visualization
11. Hardware-in-the-loop testing framework
12. Educational video/animation generation
13. Optical/photonic ferroelectric integration

---

## Dependencies

### Go Packages
- Standard library (math, encoding, compress)
- Vulkan bindings (planned: vgpu/cogentcore)

### External Tools
- GLSL compiler (glslangValidator)
- Vulkan SDK
- MNIST dataset (auto-download)

---

## Contributing

This project is part of the Ferroelectric CIM educational initiative. Key areas for contribution:
- Additional neural network architectures
- Visualization improvements
- Documentation and tutorials
- Hardware validation data

---

## References

- external research institution Ferroelectric CIM: [One Small Step Grant](https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants)
- Dr. external research group Lab: [Google Scholar](https://scholar.google.com/citations?user=YwoecRMAAAAJ)
- HfO₂ Ferroelectric Roadmap: [APL Materials](https://pubs.aip.org/aip/apm/article/11/8/089201/2908480)

---

*Last updated: January 2026 - Iteration 153*
