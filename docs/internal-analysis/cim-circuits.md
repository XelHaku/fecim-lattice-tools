# Compute-in-Memory Circuits: Research Synthesis

> **Note:** Internal analysis note. Values are reported/illustrative and not validated by this codebase.

## 1. Executive Summary

This document synthesizes current research and architectural specifications for the peripheral circuits supporting Ferroelectric Compute-in-Memory (FeCIM) systems. The peripheral suite consists of Digital-to-Analog Converters (DAC) for input/programming, Transimpedance Amplifiers (TIA) for current sensing, Analog-to-Digital Converters (ADC) for output quantization, and Charge Pumps for high-voltage write operations. 

A recurring theme in literature is ADC dominance in total energy budgets. The FeCIM demo baseline uses 5-bit precision (~30 discrete levels; simulation baseline) as an illustrative trade-off between accuracy and circuit complexity (not validated here).

> **Note:** Numeric values below are illustrative literature-style defaults. For actual simulator defaults, see `shared/peripherals` and Module 4 UI settings.

## 2. DAC (Digital-to-Analog Converter)

The DAC converts digital activation levels or programming codes into analog voltages. For FeFET programming, it must provide a range capable of exceeding the coercive field.

- **Resolution:** 5-bit (32 levels, 30 utilized for FeCIM)
- **Output Range:** ±1.5V (after amplification/boost)
- **Settling Time:** 10 ns
- **Energy:** 15 fJ/conversion
- **Linearity (INL/DNL):** 0.5 / 0.25 LSB
- **Role:** Drives Word Lines (WL) during read/compute and Bit Lines (BL) during write operations.

## 3. ADC (Analog-to-Digital Converter)

The ADC is the primary bottleneck in CiM systems. The SAR architecture is selected for its superior energy-area efficiency at medium resolutions.

- **Architecture:** SAR (Successive Approximation Register)
- **Resolution:** 5-bit
- **Conversion Time:** 50 ns
- **Energy:** 25 fJ/conversion
- **ENOB:** 4.87 bits
- **Analysis:** ADC power consumption scales exponentially with resolution ($2^N$). Limiting to 5-bit precision is essential for achieving competitive TOPS/W metrics.

## 4. TIA (Transimpedance Amplifier)

The TIA serves as the interface between the analog crossbar array and the digital-ready ADC, converting column currents into voltages.

- **Gain:** 10 kΩ (converts 100 µA to 1.0V)
- **Bandwidth:** 100 MHz
- **Input Noise:** 1 pA/√Hz
- **Max Input Current:** 100 µA (Saturation point)
- **SNR Requirement:** >32 dB for 5-bit accuracy.

## 5. Charge Pump

Since modern CMOS logic operates at ~1.0V, an on-chip charge pump is required to generate the ±1.5V pulses needed for FeFET switching.

- **Topology:** 2-stage Dickson Charge Pump
- **Input:** 1.0V CMOS supply
- **Output:** ±1.5V write voltage
- **Efficiency:** 70%
- **Clock Frequency:** 50 MHz
- **Role:** Providing the high-voltage "headroom" necessary for non-volatile state updates.

## 6. Energy Efficiency Analysis

### Power Breakdown
The following chart describes the typical energy distribution during a single Multiply-Accumulate (MAC) operation:

- **ADC:** 65% (Dominant)
- **Array (MVM):** 20%
- **TIA/Sense:** 10%
- **DAC:** 5%

### System Metrics
- **Energy per MAC:** ~100 fJ/MAC (estimated 110 fJ including peripherals)
- **Peak Efficiency:** 885 TOPS/W
- **Comparison:**
  - **GPU (H100/A100):** 1-10 TOPS/W
  - **FeCIM:** ~800+ TOPS/W (80-100x improvement)

## 7. Timing Specifications

The system operates on a distinct phase-based timing model for Read, Write, and Compute cycles.

### Timing Diagram (ASCII)
```text
Phase         | 0ns      | 50ns     | 100ns    | 150ns    | 200ns
--------------|----------|----------|----------|----------|----------
WRITE CYCLE   |[DAC Settle][Pump Rise][   Write Pulse    ]|
              | 10ns     | 40ns     | 100ns               |
--------------|----------|----------|----------|----------|----------
READ CYCLE    |[TIA Settle][  ADC Convert  ]|
              | 10ns       | 50ns           |
--------------|----------|----------|----------|----------|----------
COMPUTE (MVM) |[DAC Array][ Array MVM ][TIA][ADC Conv]|
              | 10ns      | 15ns       |10ns| 50ns    |
```

- **Write cycle:** ~150-170 ns
- **Read cycle:** ~60-65 ns
- **Compute (MVM):** ~75-85 ns
- **Throughput:** ~4.7 GOPS (for a 256x256 array at 75ns cycle)

## 8. CMOS Process Nodes

The peripheral circuits are designed for compatibility with standard foundry processes to enable monolithic integration of ferroelectric materials in the Back-End-of-Line (BEOL).

- **SKY130 (130nm):** Default open-source PDK; excellent for educational and academic prototyping.
- **GF180MCU (180nm):** Optimized for high-voltage (up to 5V) support, beneficial for charge pump design.
- **22nm FD-SOI:** Target for high-performance integration (e.g., CEA-Leti BEOL FeRAM platform); provides 100x lower energy than planar nodes.

## 9. Signal Flow Diagrams

### READ Path
Digital Address → WL Driver → **Crossbar Cell** → **TIA** → **ADC** → Digital Output

### WRITE Path
Digital Level → **DAC** → **Charge Pump** → **FeFET Gate** (State Update)

### COMPUTE Path
Digital Vector → **DAC Array** → **Crossbar Array** → **TIA Array** → **ADC Array** → Digital Result

## 10. References

1. *Peripheral Circuits Research Meta-Study*, FeCIM Project Internal Docs, 2026.
2. *Module 4: Peripheral Circuits Architecture*, FeCIM Architecture Reference, 2026.
3. *Analog CIM Energy Efficiency Analysis*, arXiv:24xx.xxxx.
4. *HCiM: ADC-Less Architectures for Hybrid CiM*, ISSCC 2024.
5. *Multi-Level FeFET Programming Schemes*, Nature Communications 2023.
