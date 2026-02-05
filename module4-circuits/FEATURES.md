# Module 4: Circuits - Features

Peripheral-circuit demo: DAC -> charge pump -> FeFET array -> TIA -> ADC.

---

## Features

- **Three Operation Modes** - READ, WRITE, COMPUTE
- **Signal-Chain Visualization** - DAC, charge pump, FeFET array, TIA, ADC
- **Architecture Modes (UI)** - 0T1R (passive), 1T1R, 2T1R voltage-rule visualization
- **Linearity Analysis** - INL/DNL plots for DAC and ADC
- **Timing & Power Models** - Derived from peripheral defaults (not silicon-calibrated)
- **Voltage-Zone Panels** - Safe read vs write windows
- **Array State View** - Programmed level heatmap and step-through compute view
- **GPU Batch Peripherals (Experimental)** - Optional GPU batch conversion utilities (not wired into GUI)

---

## Default Circuit Models (From `shared/peripherals`)

| Circuit | Default Specs |
|---|---|
| **DAC** | 5-bit, +/-1.5 V range, 10 ns settle, INL 0.5 LSB, DNL 0.25 LSB |
| **ADC** | 5-bit SAR, 0-1.0 V range, 50 ns convert, INL 0.5 LSB, DNL 0.25 LSB |
| **TIA** | 10 kOhm gain, 100 MHz BW, 1 pA/sqrtHz input noise |
| **Charge Pump** | 2-stage Dickson, 1.0 V -> 1.5 V, 70% efficiency |

---

Defaults above are simulation parameters from `shared/peripherals`, not measured silicon values (DOI: (add)).

## Key Parameters

| Parameter | Default | Notes |
|---|---:|---|
| DAC / ADC bits | 5 / 5 | 32 levels, demo uses 30 |
| Read zone | 0 to 0.5xVc | Safe sensing window |
| Write zone | Vc to 2.5xVc | Programming window |

---

## Notes

- Architecture modes affect voltage rules and visualization; device-level transistor physics is not modeled.
- Timing and power outputs are model-based estimates for education.
