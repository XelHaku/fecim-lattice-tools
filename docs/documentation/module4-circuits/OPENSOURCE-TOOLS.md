<!-- Category: Open-Source Tools | Module: module4-circuits | Reading time: ~3 min -->
# Module 4 Open-Source Tools: Circuit Simulation

> Tools for simulating and verifying the peripheral circuits that
> interface with CIM crossbar arrays.

---

## Tools Used in This Module

| Tool | Role |
|------|------|
| Go toolchain | All simulation models (DAC, TIA, ADC, charge pump, noise) |
| Fyne | GUI rendering for the circuit interface |

The peripheral circuit models are implemented as Go functions in
`shared/peripherals/`. They use idealized transfer functions, not
transistor-level simulation.

---

## External Circuit Simulation Tools

These open-source tools are relevant for deeper circuit analysis
beyond what the FeCIM simulator provides.

### SPICE Simulators

| Tool | Description |
|------|-------------|
| **ngspice** | Open-source SPICE simulator. Handles transient, AC, DC analysis. Can import SPICE netlists exported by Module 6. |
| **Xyce** | Parallel SPICE from Sandia National Labs. Good for large-array simulations where ngspice runs out of memory. |
| **QUCS-S** | GUI front-end for ngspice/Xyce. Schematic capture and waveform viewer. |

### Compact Model Libraries

| Resource | Description |
|----------|-------------|
| **BSIM** | Industry-standard MOSFET compact models (BSIM3, BSIM4, BSIM-CMG for FinFET). |
| **PSP** | Surface-potential-based MOSFET model, alternative to BSIM. |
| **Verilog-A models** | Behavioral device models that can represent FeFET characteristics in SPICE. |

### Waveform and Analysis

| Tool | Description |
|------|-------------|
| **GTKWave** | Waveform viewer for digital and analog simulation outputs. |
| **Python + matplotlib** | Post-processing SPICE output (.raw files) for publication-quality plots. |
| **LTspice** | Free (not open-source) SPICE with integrated schematic editor. Widely used for quick analog verification. |

---

## Integration Path

Module 6 exports SPICE netlists from the compiled crossbar design:

```
Module 6 export --> .sp netlist --> ngspice/Xyce --> waveform analysis
```

The exported netlist represents cells as resistors with conductance
values. To add FeFET physics, replace the resistor with a Verilog-A
compact model.

---

## Code Locations

| Path | Purpose |
|------|---------|
| `shared/peripherals/dac.go` | DAC model |
| `shared/peripherals/tia.go` | TIA model |
| `shared/peripherals/adc.go` | ADC model |
| `shared/peripherals/chargepump.go` | Charge pump model |
| `shared/peripherals/noise.go` | Noise composition |
| `module6-eda/pkg/export/spice.go` | SPICE netlist export |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
