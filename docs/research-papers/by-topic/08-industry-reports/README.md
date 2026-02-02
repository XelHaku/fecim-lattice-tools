# Industry Reports

This directory contains industry surveys, roadmaps, and large-scale implementation reports that provide context for the commercial viability and competitive landscape of ferroelectric computing technologies.

## Papers in this Directory

### Edge_AI_Accelerators_Survey_arXiv.pdf
Survey of edge AI accelerator architectures and deployment strategies for on-device machine learning inference.

### Hardware_Accelerators_Deep_Learning_arXiv.pdf
Comprehensive review of hardware acceleration approaches for deep learning, comparing ASIC, FPGA, GPU, and emerging memory-based architectures.

### ITRS_IRDS_Roadmap_2025.pdf
International Roadmap for Devices and Systems (IRDS) 2025 edition, outlining semiconductor industry projections for emerging memory and compute technologies.

### Tsinghua_Face_Classification_NatureComms_2025.pdf
Nature Communications 2025 report on large-scale face classification system implementation using emerging memory technologies, demonstrating real-world deployment capabilities.

### Wafer_Scale_Integration_Review_arXiv.pdf
Review of wafer-scale integration techniques and challenges, relevant for understanding manufacturing scalability of ferroelectric memory arrays.

## Key Findings

1. **Industry Momentum**: The IRDS roadmap identifies emerging non-volatile memories (including ferroelectrics) as critical for post-CMOS scaling.

2. **Edge AI Market**: Edge AI accelerators are driving demand for energy-efficient, high-density memory technologies that can perform compute-in-memory operations.

3. **Real-World Validation**: The Tsinghua face classification system demonstrates that ferroelectric-based systems can achieve commercial-grade performance on practical workloads.

4. **Manufacturing Readiness**: Wafer-scale integration reviews indicate that ferroelectric devices benefit from CMOS-compatible processing, enabling rapid technology transfer.

5. **Performance Benchmarks**: Hardware accelerator surveys provide competitive context, showing where ferroelectric CIM fits in the landscape of deep learning accelerators.

## Relevance to FeCIM

These industry reports inform the FeCIM Lattice Tools project in several critical ways:

**Note:** References to 30 levels refer to the demo baseline (conference claim; pending peer review). Peer‑reviewed devices report 32–140 states.

- **Competitive Positioning**: Understanding where 30-level demo baseline ferroelectric cells fit against commercial alternatives (Flash, ReRAM, PCM)
- **Use Case Validation**: Real-world deployments (face classification) validate the MNIST and neural network modules in this toolkit
- **Manufacturing Context**: IRDS roadmap data helps ground our simulations in realistic process assumptions
- **Market Timing**: Industry surveys indicate current market readiness for ferroelectric CIM solutions
- **Performance Targets**: Accelerator benchmarks provide targets for our crossbar and circuit simulations

The comparison module (module5) directly leverages insights from these reports to provide accurate technology comparisons.

## Related Topics

- **[09-reviews-surveys](../09-reviews-surveys/)** - Academic surveys providing technical depth on emerging memory technologies
- **[15-3d-stacking-architectures](../15-3d-stacking-architectures/)** - 3D integration crucial for wafer-scale implementations
- **[19-variability-yield](../19-variability-yield/)** - Manufacturing yield challenges highlighted in industry roadmaps
- **[06-neural-networks-inference](../06-neural-networks-inference/)** - Neural network implementations validated by real-world deployments
