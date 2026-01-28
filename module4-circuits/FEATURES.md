# Module 4: Circuits - Features

## Features

- **Three Operation Modes** — READ (sense), WRITE (program), COMPUTE (MVM)
- **Complete Signal Chain** — DAC → ChargePump → FeFET → TIA → ADC
- **Architecture Support** — 0T1R (passive), 1T1R (gated)
- **Material Calibration** — Voltage ranges from physics.yaml (Ec, thickness)
- **INL/DNL Analysis** — Linearity metrics for converters
- **GPU Acceleration** — Vulkan compute shaders for batch operations

## Circuit Models

| Circuit | Specs |
|---------|-------|
| **DAC** | 5-bit, ±1.5V range, 10ns settle, 15 fJ |
| **ADC** | 5-bit SAR, 0-1V range, 50ns convert, 25 fJ |
| **TIA** | 10 kΩ gain, 100 MHz BW, 1 pA/√Hz noise |
| **Charge Pump** | 2-stage Dickson, 1V→1.5V, 70% efficiency |

## Key Parameters

| Parameter | Value | Notes |
|-----------|-------|-------|
| DAC Resolution | 5-bit (32 levels) | Uses 30 for FeCIM |
| ADC Resolution | 5-bit | ENOB: 4.87 bits |
| Read Zone | 0 to 0.5×Vc | Safe sensing |
| Write Zone | Vc to 2.5×Vc | Programming window |
| Max Voltage | 3.0V | Hardware limit |

## Timing

| Operation | Duration |
|-----------|----------|
| Write Cycle | ~170 ns |
| Read Cycle | ~65 ns |
| Throughput | ~4.7 GOPS |

## Energy Budget

| Component | Energy |
|-----------|--------|
| DAC | 15 fJ |
| ADC | 25 fJ |
| TIA | ~5 fJ |
| Pump | ~10 fJ |
| **Total** | ~50 fJ/op |
