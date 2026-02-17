# Reviews and Surveys

This directory contains comprehensive academic review papers and surveys on emerging memory technologies, providing foundational context for understanding ferroelectric memory's position in the broader landscape of non-volatile memory research.

## Papers in this Directory

### Emerging_Memory_Technologies_Review_arXiv.pdf
Broad survey of emerging non-volatile memory technologies including ReRAM, PCM, STT-MRAM, and ferroelectric memories, comparing performance metrics, scalability, and application domains.

### Flash_Memory_vs_Emerging_NVM_Review_arXiv.pdf
Comparative analysis of traditional Flash memory against emerging NVM technologies, highlighting the physical limits of Flash and opportunities for replacement technologies.

### in_memory_computing_dnn_survey_2023.pdf
2023 survey specifically focused on in-memory computing architectures for deep neural networks, covering analog compute, digital CIM, and hybrid approaches.

### Non_Volatile_Memory_ML_arXiv.pdf
Review of non-volatile memory applications in machine learning, covering both inference and training use cases, with emphasis on energy efficiency and latency benefits.

## Key Findings

1. **Technology Maturity Spectrum**: Ferroelectric memories occupy a unique position with CMOS compatibility superior to ReRAM/PCM and endurance far exceeding Flash.

2. **CIM Suitability Rankings**: Surveys consistently rank ferroelectric devices highly for compute-in-memory due to analog programmability, low-voltage operation, and non-destructive reads.

3. **Performance Trade-offs**: Reviews identify key trade-offs between programming speed, endurance, retention, and analog resolution - ferroelectrics excel in the latter two.

4. **Application-Specific Optimization**: DNN surveys emphasize that different layers (convolution vs fully-connected) benefit from different memory characteristics, suggesting hybrid architectures.

5. **Flash Replacement Urgency**: Flash review papers highlight that NAND scaling is reaching fundamental limits, creating urgent demand for 3D-stackable alternatives.

## Relevance to FeCIM

These survey papers provide essential context for the FeCIM Lattice Tools project:

- **Design Space Validation**: Our 30-level demo baseline aligns with survey findings on optimal analog resolution for DNN workloads
- **Comparison Module Accuracy**: Survey data informs the technology comparison parameters in module5
- **Architecture Decisions**: CIM survey insights guide our crossbar non-ideality modeling (IR drop, sneak paths)
- **Benchmark Selection**: Survey-identified standard benchmarks (MNIST, CIFAR-10) justify our neural network test cases
- **Physical Constraints**: Reviews provide realistic bounds for parameters like endurance (10⁹-10¹² cycles) and retention times

The physics models in module1 (hysteresis, Preisach) and module2 (crossbar non-idealities) are validated against ranges documented in these surveys.

## Related Topics

- **[08-industry-reports](../08-industry-reports/)** - Industry perspective on commercial readiness and market timing
- **[01-material-physics](../01-ferroelectric-materials/)** - Detailed physics underlying the technologies surveyed here
- **[06-neural-networks-inference](../10-cim-compilers-mapping/)** - Specific DNN implementations discussed in CIM surveys
- **[10-cim-compilers-mapping](../10-cim-compilers-mapping/)** - Software tools for mapping workloads to CIM architectures
- **[19-variability-yield](../19-variability-yield/)** - Manufacturing challenges identified across surveyed technologies
