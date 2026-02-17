# Module 4: Peripheral Circuits (DAC/ADC/TIA)

**Navigation:** [← Back to Learn](../README.md) | [ELI5](./eli5.md) | [Physics](./physics.md) | [Features](./features.md)

---

## Overview

Module 4 simulates the peripheral circuits required for crossbar array operation: Digital-to-Analog Converters (DAC), Transimpedance Amplifiers (TIA), and Analog-to-Digital Converters (ADC). It demonstrates the complete signal chain from digital input to analog cell programming to digital readout.

**Key Concept:** Crossbar arrays need peripheral circuits to interface between digital control and analog memory/compute. This module shows how DACs drive inputs, TIAs sense currents, and ADCs quantize outputs.

---

## Quick Links

### For Beginners
- **[ELI5 Explanation](./eli5.md)** - What are DACs and ADCs?

### For Developers
- **[Physics Reference](./physics.md)** - Transfer functions, INL/DNL, noise
- **[Features](./features.md)** - Circuit models, analysis tools

### For Researchers
- **[Open-Source Tools](./tools.md)** - SPICE integration, circuit libraries

---

## Module Contents

```
module4-circuits/
├── shared/peripherals/      # Circuit models (shared)
│   ├── dac.go                # Digital-to-Analog Converter
│   ├── adc.go                # Analog-to-Digital Converter
│   ├── tia.go                # Transimpedance Amplifier
│   ├── chargepump.go         # Voltage multiplier
│   ├── analysis.go           # INL/DNL analysis
│   └── defaults.go           # Default parameters
├── pkg/gui/                  # Fyne visualization
│   ├── app.go                # Main circuit GUI
│   ├── tab_unified_voltage.go # Unified voltage/ISPP tab
│   └── device_state.go       # Device state machine
└── pkg/arraysim/             # Array simulation integration
```

---

## Quick Start

### GUI Mode
```bash
fecim-lattice-tools circuits
```

### Headless Analysis
```bash
fecim-lattice-tools --mode circuits --dac-bits 8 --adc-bits 6
```

---

## What You'll Learn

1. **DAC Operation**
   - Code-to-voltage mapping
   - LSB calculation
   - INL/DNL (Integral/Differential Nonlinearity)

2. **TIA Operation**
   - Current-to-voltage conversion
   - Transimpedance gain
   - Noise analysis

3. **ADC Operation**
   - Voltage-to-code quantization
   - Quantization error (±0.5 LSB)
   - Resolution vs speed trade-offs

4. **Charge Pump**
   - Voltage multiplication
   - Stage calculation
   - Efficiency vs ripple

---

## Signal Chain

```
Digital Input → DAC → Crossbar Cell → TIA → ADC → Digital Output
   (code)      (V)      (G×V=I)        (V)   (code)

Example:
  Input code: 128 (8-bit)
       ↓ DAC (0-1V range)
  Voltage: 0.502V
       ↓ Crossbar cell (G=50µS)
  Current: 25.1µA
       ↓ TIA (R=100kΩ)
  Voltage: 2.51V
       ↓ ADC (6-bit, 0-3.3V range)
  Output code: 48
```

---

## Key Features

- **DAC modeling:** 4-bit to 12-bit resolution
- **ADC modeling:** 4-bit to 10-bit resolution
- **TIA simulation:** Configurable gain, offset, noise
- **Charge pump:** Dickson topology with losses
- **INL/DNL analysis:** Visualize nonlinearity
- **Voltage zone visualization:** 0T1R/1T1R/2T1R schemes
- **ISPP programming:** Incremental Step Pulse Programming
- **Architecture comparison:** Different transistor topologies

---

## Circuit Parameters

### DAC (Digital-to-Analog Converter)

```
Resolution: N bits → 2^N levels
LSB = (VrefHigh - VrefLow) / (2^N - 1)
Vout = VrefLow + code × LSB

Example (8-bit, 0-1V):
  LSB = 1V / 255 = 3.92 mV
  Code 128 → 0.502V
```

### ADC (Analog-to-Digital Converter)

```
Resolution: N bits → 2^N levels
LSB = (VrefHigh - VrefLow) / (2^N - 1)
code = round((Vin - VrefLow) / LSB)

Quantization error: ±0.5 LSB
```

### TIA (Transimpedance Amplifier)

```
Vout = Iin × Rtia + Voffset
Rtia = 10kΩ to 1MΩ (typical)

Noise: Vnoise_rms = Inoise_rms × Rtia × √BW
```

---

## Architecture Voltage Rules

### 0T1R (Passive Crossbar)

```
Selected cell: Full voltage
Unselected (same row): V/2
Unselected (same col): V/2
Other cells: 0V

Half-select disturb issue!
```

### 1T1R (1 Transistor per Cell)

```
Transistor blocks reverse current
Reduces half-select to ~V/10
Standard production architecture
```

### 2T1R (2 Transistors)

```
Complete isolation
No disturb
Higher area cost
```

---

## ISPP (Incremental Step Pulse Programming)

```
Target level: 15 (out of 30)

Step 1: Apply V1 → Read → Level 12 (too low)
Step 2: Apply V1+ΔV → Read → Level 17 (overshoot)
Step 3: Apply V1+ΔV/2 → Read → Level 15 (converged!)

Convergence typically 3-10 iterations
```

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | Circuit basics | Beginners |
| [physics.md](./physics.md) | Transfer functions, analysis | Developers |
| [features.md](./features.md) | Workflows, models | Developers |
| [tools.md](./tools.md) | SPICE integration | Researchers |

---

## Evidence Status

- **Demonstrated:** Circuit models, transfer functions are implemented and testable
- **Modeled:** Default parameters are educational baselines, not measured from silicon
- **Aspirational:** Production-grade peripheral design is future work

---

## Related Modules

- **[Module 1: Hysteresis](../module1-hysteresis/README.md)** - FeFET device model
- **[Module 2: Crossbar](../module2-crossbar/README.md)** - Uses DAC/ADC for I/O
- **[Module 6: EDA](../module6-eda/README.md)** - Layout of peripheral circuits

---

## Testing

```bash
go test ./shared/peripherals
go test ./module4-circuits/pkg/arraysim
```

---

**Last Updated:** 2026-02-16
