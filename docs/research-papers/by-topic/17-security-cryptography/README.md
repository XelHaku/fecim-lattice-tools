# Security and Cryptography

This directory explores security applications of ferroelectric devices, particularly Physical Unclonable Functions (PUFs) and hardware security primitives. Ferroelectric devices exhibit inherent manufacturing variations and stochastic switching behaviors that can be leveraged for security rather than compensated as defects.

## Papers in this Directory

### fefet_puf_charge_domain_computing_2025.pdf
2025 demonstration of FeFET-based Physical Unclonable Functions using charge domain computing. Exploits threshold voltage variations in ferroelectric transistors to generate stable, unclonable cryptographic keys.

### memristor_puf_lightweight_crypto_2022.pdf
2022 work on memristor-based PUF implementations for lightweight cryptography in IoT devices. Demonstrates low-power key generation and authentication using resistive switching variability.

## Key Findings

1. **Intrinsic Entropy Source**: Ferroelectric domain nucleation and switching exhibit quantum-scale stochastic behavior, providing true random number generation (TRNG) capabilities.

2. **Manufacturing Variation Advantage**: Rather than being a liability, device-to-device ferroelectric variations become the foundation for unclonable PUF implementations.

3. **Charge Domain Security**: FeFET threshold voltage distributions can be mapped to stable cryptographic keys with bit error rates <1% and inter-chip Hamming distances >48%.

4. **Low-Power Authentication**: PUF-based authentication consumes 100-1000× less energy than traditional cryptographic hardware, critical for edge/IoT devices.

5. **Aging Stability**: Despite endurance cycling, ferroelectric PUF characteristics remain stable over 10⁶+ authentication cycles with error correction codes (ECC).

## Relevance to FeCIM

Security applications represent an important diversification opportunity for ferroelectric technology:

- **Variability Modeling**: Our variability models (currently in module2 for non-idealities) can be repurposed to simulate PUF behavior
- **Device Characterization**: The hysteresis and threshold voltage distributions we model for memory can characterize PUF uniqueness
- **Secure CIM**: Future integration of PUF capabilities into CIM systems for secure edge AI applications
- **TRNG Module**: Physics-based stochastic switching models could implement true random number generation
- **Dual-Use Technology**: Same ferroelectric devices serve both memory/compute and security functions

**Potential Module**: A security module could simulate PUF key extraction, challenge-response protocols, and TRNG implementations using our ferroelectric device models.

## Related Topics

- **[01-material-physics](../01-ferroelectric-materials/)** - Stochastic domain switching physics underlying PUF behavior
- **[19-variability-yield](../19-variability-yield/)** - Device variation that enables PUF uniqueness
- **[06-neural-networks-inference](../10-cim-compilers-mapping/)** - Secure edge AI combining CIM and PUF
- **[02-crossbar-arrays](../04-cim-architectures/)** - Crossbar-based PUF architectures
- **[08-industry-reports](../08-industry-reports/)** - IoT and edge computing market drivers for hardware security
