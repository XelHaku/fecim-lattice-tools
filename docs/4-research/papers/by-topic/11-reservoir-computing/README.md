# Reservoir Computing

This directory explores reservoir computing (RC) implementations using ferroelectric devices. Reservoir computing is a recurrent neural network paradigm where a fixed, randomly connected "reservoir" layer performs temporal feature extraction, with only the output layer trained. Ferroelectric devices are particularly well-suited for RC due to their natural nonlinear dynamics and temporal memory effects.

## Papers in this Directory

### all_ferroelectric_reservoir_computing_2023.pdf
2023 demonstration of fully ferroelectric reservoir computing system where both reservoir and readout layers use ferroelectric devices. Achieves state-of-the-art performance on temporal classification tasks with minimal training overhead.

### analog_rc_ferroelectric_mpb_transistors_2024.pdf
2024 work on analog reservoir computing using ferroelectric transistors based on morphotropic phase boundary (MPB) compositions. Exploits enhanced piezoelectric response at MPB for richer temporal dynamics.

## Key Findings

1. **Temporal Dynamics Advantage**: Ferroelectric devices exhibit natural temporal memory effects (domain switching dynamics, relaxation) that provide "free" computational power for sequence processing tasks.

2. **Training Efficiency**: Reservoir computing with ferroelectric devices requires training only linear output weights, reducing training energy by 100-1000× compared to fully-trained RNNs.

3. **MPB Enhancement**: Morphotropic phase boundary compositions (like HfO₂-ZrO₂) offer superior nonlinear response curves, improving reservoir expressivity.

4. **Speech/Time-Series Performance**: Ferroelectric RC systems achieve competitive accuracy on speech recognition and time-series prediction tasks while consuming orders of magnitude less energy than GPU-based RNNs.

5. **All-Ferroelectric Integration**: Using ferroelectric devices for both reservoir and readout simplifies fabrication and enables monolithic integration.

## Relevance to FeCIM

Reservoir computing represents an important extension of the FeCIM Lattice Tools project:

- **HZO Material Synergy**: Our focus on HfO₂-ZrO₂ superlattices aligns perfectly with MPB-based reservoir computing research
- **Temporal Computing**: RC adds time-series processing capabilities beyond the feedforward DNNs in our current module3
- **Hysteresis Exploitation**: Our Preisach hysteresis model (module1) captures the nonlinear dynamics central to reservoir computing
- **Energy Efficiency**: RC's training efficiency complements our compute-in-memory focus on inference efficiency
- **Future Module Opportunity**: A dedicated reservoir computing module could leverage our existing physics models to simulate RC workloads

**Potential Extension**: The physics modeling in module1 (P-E curves, domain switching) provides a foundation for simulating reservoir dynamics. A future module7-reservoir could implement speech recognition or time-series forecasting using ferroelectric RC.

## Related Topics

- **[01-material-physics](../01-ferroelectric-materials/)** - HZO material properties (MPB compositions) enabling superior RC performance
- **[06-neural-networks-inference](../10-cim-compilers-mapping/)** - Feedforward DNNs that RC extends with temporal processing
- **[03-analog-in-memory-compute](../04-cim-architectures/)** - Analog computing principles leveraged by RC
- **[12-benchmarking-datasets](../08-industry-reports/)** - Time-series and speech datasets for RC evaluation
- **[19-variability-yield](../19-variability-yield/)** - Device variation affects reservoir richness and requires robustness analysis
