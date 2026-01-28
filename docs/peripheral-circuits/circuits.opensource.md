# Open-Source Peripheral Circuit Tools for CIM

**A Comprehensive Guide to ADC, DAC, TIA, and Mixed-Signal Design Tools**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source tools, libraries, and frameworks for designing and simulating peripheral circuits for compute-in-memory systems. It covers ADC/DAC design, analog simulation, mixed-signal verification, and related tools from academic research and the open-source community.

---

## 1. SPICE Simulators

### 1.1 ngspice

**URL:** https://ngspice.sourceforge.io/

**Description:** The gold standard open-source SPICE simulator. Essential for analog circuit design and verification.

**Features:**
- Full SPICE3f5 compatibility
- Verilog-A device model support (via OpenVAF)
- BSIM3/BSIM4 MOSFET models
- Behavioral modeling (ABM)
- Python bindings (PySpice)
- Parallel simulation support

**Installation:**
```bash
# Ubuntu/Debian
sudo apt install ngspice

# macOS
brew install ngspice

# From source
git clone https://git.code.sf.net/p/ngspice/ngspice
cd ngspice && ./autogen.sh && ./configure && make && sudo make install
```

**Example - 5-bit SAR ADC:**
```spice
* 5-bit SAR ADC for CIM
.include "cmos_22nm.lib"

* Comparator
.subckt comparator inp inn out vdd vss
    M1 tail vss inn vss nmos W=1u L=22n
    M2 out vdd inp tail nmos W=1u L=22n
    M3 out vdd vdd vdd pmos W=2u L=22n
    Ibias vdd tail 10u
.ends

* DAC capacitor array (binary weighted)
.subckt cap_dac d4 d3 d2 d1 d0 out vref vss
    C4 d4 out 16p  ; MSB
    C3 d3 out 8p
    C2 d2 out 4p
    C1 d1 out 2p
    C0 d0 out 1p   ; LSB
    C_ref out vss 1p  ; Unit cap
.ends

* Test bench
Vin in 0 PWL(0 0 10n 1)
Vref ref 0 1.0
XDAC d4 d3 d2 d1 d0 dac_out ref 0 cap_dac
XCOMP in dac_out comp_out vdd 0 comparator

.tran 1n 100n
.end
```

**Relevance to FeCIM:** Validates our behavioral ADC/DAC models against transistor-level simulation.

---

### 1.2 Xyce (Sandia)

**URL:** https://xyce.sandia.gov/

**Description:** Parallel SPICE simulator from Sandia National Labs. Excellent for large-scale simulations.

**Features:**
- Massive parallelism (1000s of cores)
- Harmonic balance analysis
- Sensitivity analysis
- Device-level noise simulation
- Verilog-A support

**Installation:**
```bash
# From source (requires Trilinos)
git clone https://github.com/Xyce/Xyce
cd Xyce && mkdir build && cd build
cmake .. && make -j8
```

**Relevance to FeCIM:** Can simulate full crossbar + peripherals at transistor level.

---

### 1.3 QUCS / QUCS-S

**URL:** https://github.com/Qucs/qucs

**Description:** GUI-based circuit simulator with SPICE integration.

**Features:**
- Schematic capture
- S-parameter simulation
- Harmonic balance
- Integration with ngspice/Xyce

**Installation:**
```bash
# Ubuntu
sudo apt install qucs

# With ngspice backend (QUCS-S)
sudo apt install qucs-s
```

---

## 2. ADC Design Tools

### 2.1 ADC Design Framework (Open)

**URL:** https://github.com/ucb-art/adc_generator

**Description:** UC Berkeley's ADC generator using BAG framework.

**Features:**
- SAR ADC generator
- Flash ADC templates
- Layout automation
- Calibration models

**Example (Python API):**
```python
from bag.layout import RoutingGrid
from adc_generator import SARGenerator

# Configure 5-bit SAR ADC
config = {
    'resolution': 5,
    'sampling_rate': 100e6,  # 100 MS/s
    'supply': 1.0,           # 1V
    'input_range': 1.0,      # 1V p-p
}

sar = SARGenerator(config)
sar.generate_schematic()
sar.run_simulation()
print(f"ENOB: {sar.enob:.2f} bits")
print(f"Power: {sar.power * 1e6:.2f} µW")
```

---

### 2.2 ADCpy (Python ADC Modeling)

**URL:** Custom / various implementations

**Description:** Python-based ADC behavioral modeling.

**Example Implementation:**
```python
import numpy as np

class SARADC:
    """5-bit SAR ADC behavioral model matching our Go implementation."""

    def __init__(self, bits=5, vref=1.0, inl=0.5, dnl=0.25):
        self.bits = bits
        self.levels = 2 ** bits
        self.vref = vref
        self.inl = inl  # LSB
        self.dnl = dnl  # LSB
        self.lsb = vref / (self.levels - 1)

    def convert(self, voltage):
        """Ideal conversion."""
        voltage = np.clip(voltage, 0, self.vref)
        level = int(round(voltage / self.lsb))
        return min(level, self.levels - 1)

    def convert_with_nonlinearity(self, voltage):
        """Conversion with INL/DNL errors."""
        ideal = self.convert(voltage)

        # INL: sinusoidal error across codes
        inl_error = self.inl * self.lsb * np.sin(np.pi * ideal / (self.levels - 1))

        # DNL: code-dependent step error
        dnl_error = self.dnl * self.lsb * (0.5 - (ideal % 5) / 4.0)

        return self.convert(voltage + inl_error + dnl_error)

    def enob(self):
        """Effective number of bits."""
        noise_factor = np.sqrt(1 + self.inl**2 + self.dnl**2)
        return self.bits - np.log2(noise_factor)

    def energy_per_conversion(self):
        """Energy in femtojoules."""
        return 5e-15 * self.bits  # ~5 fJ/bit for SAR

# Usage
adc = SARADC(bits=5)
print(f"Levels: {adc.levels}")
print(f"ENOB: {adc.enob():.2f}")
print(f"Energy: {adc.energy_per_conversion() * 1e15:.1f} fJ")
```

---

### 2.3 pySPICE

**URL:** https://github.com/PySpice-org/PySpice

**Description:** Python interface to ngspice/Xyce.

**Installation:**
```bash
pip install PySpice
```

**Example - ADC Characterization:**
```python
from PySpice.Spice.NgSpice.Shared import NgSpiceShared
from PySpice.Probe.Plot import plot

# Create circuit
circuit = Circuit('SAR_ADC_Test')
circuit.include('cmos_22nm.lib')

# Add components
circuit.V('in', 'input', circuit.gnd, 'SIN(0.5 0.5 1k)')
circuit.X('adc', 'sar_adc_5bit', 'input', 'd4', 'd3', 'd2', 'd1', 'd0')

# Simulate
simulator = circuit.simulator(temperature=25, nominal_temperature=25)
analysis = simulator.transient(step_time=1@u_ns, end_time=1@u_ms)

# Plot results
plot(analysis.time, analysis['d4'], analysis['d3'], analysis['d2'], analysis['d1'], analysis['d0'])
```

---

## 3. DAC Design Tools

### 3.1 BAG Framework (Berkeley Analog Generator)

**URL:** https://github.com/ucb-art/BAG_framework

**Description:** Python-based analog IC generator.

**Features:**
- Schematic generator
- Layout generator
- Parasitic extraction
- Process-portable designs

**Example - 5-bit DAC:**
```python
from bag import BagProject
from analog_templates import CapDACGenerator

prj = BagProject()

# 5-bit capacitor DAC for CIM
params = {
    'resolution': 5,
    'unit_cap': 1e-15,  # 1 fF
    'vref': 1.5,        # ±1.5V for FeFET
    'architecture': 'binary_weighted',
}

gen = CapDACGenerator(prj, params)
gen.design()
gen.extract()

print(f"DNL: {gen.dnl} LSB")
print(f"INL: {gen.inl} LSB")
print(f"Settling: {gen.settling_time * 1e9} ns")
```

---

### 3.2 Open-Source DAC IP

**URL:** Various academic repositories

**Python Behavioral Model:**
```python
import numpy as np

class DAC:
    """5-bit DAC behavioral model matching our Go implementation."""

    def __init__(self, bits=5, vref_low=-1.5, vref_high=1.5, inl=0.5, dnl=0.25):
        self.bits = bits
        self.levels = 2 ** bits
        self.vref_low = vref_low
        self.vref_high = vref_high
        self.inl = inl
        self.dnl = dnl
        self.lsb = (vref_high - vref_low) / (self.levels - 1)
        self.settle_time = 10e-9  # 10 ns

    def convert(self, level):
        """Ideal conversion."""
        level = np.clip(level, 0, self.levels - 1)
        fraction = level / (self.levels - 1)
        return self.vref_low + fraction * (self.vref_high - self.vref_low)

    def convert_with_nonlinearity(self, level):
        """Conversion with INL/DNL errors."""
        ideal = self.convert(level)

        # INL error
        inl_error = self.inl * self.lsb * np.sin(np.pi * level / (self.levels - 1))

        # DNL error
        dnl_error = self.dnl * self.lsb * (0.5 - (level % 3) / 2.0)

        return ideal + inl_error + dnl_error

    def is_monotonic(self):
        """Check monotonicity (DNL > -1 LSB)."""
        return self.dnl > -1.0

# Usage
dac = DAC(bits=5)
for level in range(30):
    v = dac.convert_with_nonlinearity(level)
    print(f"Level {level:2d} -> {v:+.4f} V")
```

---

## 4. TIA / Amplifier Tools

### 4.1 Open-Source Op-Amp Library

**URL:** https://github.com/bmurmann/EE-627

**Description:** Stanford's analog IC design course materials with op-amp designs.

**Example - TIA Design in SPICE:**
```spice
* Transimpedance Amplifier for CIM readout
* Target: 10 kOhm gain, 100 MHz bandwidth

.param Rf=10k
.param Cf=0.1p

* Op-amp macro model
.subckt opamp inp inn out vdd vss
    Gm out 0 inp inn 1m
    Rout out 0 10k
    Cout out 0 0.1p
.ends

* TIA topology
Xamp inp inn out vdd vss opamp
Rf out inn {Rf}
Cf out inn {Cf}
Cin inp 0 0.5p  ; Input capacitance
Iin inp 0 AC 1u ; 1 µA input current

* Analysis
.ac dec 100 1k 1G
.noise V(out) Iin

.end
```

---

### 4.2 Analog-HDL

**URL:** https://github.com/cogwheel-hdl/cogwheel

**Description:** Analog behavioral modeling in HDL.

**Verilog-A TIA Model:**
```verilog
// TIA Verilog-A model for ngspice/Xyce
`include "constants.vams"
`include "disciplines.vams"

module tia(in, out);
    inout in, out;
    electrical in, out;

    parameter real Gain = 10e3;        // 10 kOhm
    parameter real Bandwidth = 100e6;  // 100 MHz
    parameter real Inoise = 1e-12;     // 1 pA/sqrt(Hz)
    parameter real Voffset = 5e-3;     // 5 mV offset
    parameter real Vmax = 1.0;         // Max output

    real v_out, noise_rms;

    analog begin
        // Ideal gain
        v_out = I(in) * Gain + Voffset;

        // Bandwidth limiting (single pole)
        V(out) <+ laplace_nd(v_out, {1}, {1, 1/(2*`M_PI*Bandwidth)});

        // Clamp output
        if (V(out) > Vmax)
            V(out) <+ Vmax;
        if (V(out) < 0)
            V(out) <+ 0;

        // Noise (white noise model)
        noise_rms = Inoise * Gain * sqrt(Bandwidth);
        V(out) <+ white_noise(noise_rms*noise_rms, "thermal");
    end
endmodule
```

---

## 5. Charge Pump Tools

### 5.1 Charge Pump Generator

**URL:** Academic tools / custom

**Python Model:**
```python
import numpy as np

class DicksonChargePump:
    """Dickson charge pump model matching our Go implementation."""

    def __init__(self, stages=2, vin=1.0, vout_target=1.5,
                 fly_cap=100e-12, freq=50e6, vth=0.3):
        self.stages = stages
        self.vin = vin
        self.vout_target = vout_target
        self.fly_cap = fly_cap
        self.freq = freq
        self.vth = vth
        self.efficiency = 0.7

    def ideal_vout(self):
        """Theoretical maximum output."""
        return (self.stages + 1) * self.vin

    def actual_vout(self, i_load=10e-6):
        """Output with losses."""
        vth_drop = self.vth * self.stages
        ir_drop = i_load / (self.fly_cap * self.freq)
        return self.ideal_vout() - vth_drop - ir_drop

    def max_current(self):
        """Maximum output current capability."""
        return self.fly_cap * self.freq * (self.stages + 1) * self.vin / self.vout_target

    def ripple(self, i_load=10e-6, c_out=1e-9):
        """Output ripple voltage."""
        return i_load / (c_out * self.freq)

    def power_input(self, i_load=10e-6):
        """Input power consumption."""
        return self.vout_target * i_load / self.efficiency

    def rise_time(self):
        """Approximate rise time."""
        return self.stages * 2.2 / self.freq

# Usage
pump = DicksonChargePump(stages=2, vin=1.0, vout_target=1.5)
print(f"Ideal Vout: {pump.ideal_vout():.2f} V")
print(f"Actual Vout: {pump.actual_vout():.2f} V")
print(f"Max current: {pump.max_current() * 1e6:.1f} µA")
print(f"Rise time: {pump.rise_time() * 1e9:.1f} ns")
```

---

### 5.2 SPICE Charge Pump Model

```spice
* 2-stage Dickson Charge Pump for FeCIM write operations

.param Cfly=100p
.param Cout=1n
.param Fclk=50Meg

* Clock generators (non-overlapping)
Vclk1 clk1 0 PULSE(0 1 0 1n 1n {0.5/Fclk} {1/Fclk})
Vclk2 clk2 0 PULSE(0 1 {0.5/Fclk} 1n 1n {0.5/Fclk} {1/Fclk})

* Input supply
Vin in 0 1.0

* Stage 1
C1 clk1 n1 {Cfly}
D1 in n1 dmod
D2 n1 n2 dmod

* Stage 2
C2 clk2 n2 {Cfly}
D3 n2 n3 dmod
D4 n3 out dmod

* Output cap and load
Cout out 0 {Cout}
Rload out 0 150k  ; 10 µA load at 1.5V

* Diode model (or use NMOS switches)
.model dmod D(Is=1e-14 N=1 Vj=0.3)

.tran 1n 10u
.end
```

---

## 6. Full System Tools

### 6.1 OpenFASoC

**URL:** https://github.com/idea-fasoc/OpenFASOC

**Description:** Open-source fully-autonomous SoC design framework. Includes ADC/DAC generators.

**Features:**
- Temperature sensor generator
- LDO generator
- ADC generator (SAR, Flash)
- DAC generator
- Fully automated flow

**Installation:**
```bash
git clone https://github.com/idea-fasoc/OpenFASOC
cd OpenFASOC
pip install -r requirements.txt
```

---

### 6.2 MAGICAL (Analog Layout)

**URL:** https://github.com/MAGICAL-EDA/MAGICAL

**Description:** Machine learning-assisted analog layout tool.

**Features:**
- Automated analog placement
- Routing with symmetry constraints
- DRC-clean layouts
- Open PDK support

---

### 6.3 Analog-AI Toolkit

**URL:** Custom / academic

**Description:** Tools for CIM peripheral design optimization.

**Example - Full System Analysis:**
```python
import numpy as np
from dataclasses import dataclass

@dataclass
class CIMPeripherals:
    """Full CIM peripheral system model."""

    # ADC
    adc_bits: int = 5
    adc_energy: float = 25e-15  # 25 fJ

    # DAC
    dac_bits: int = 5
    dac_energy: float = 15e-15  # 15 fJ

    # TIA
    tia_gain: float = 10e3
    tia_bw: float = 100e6
    tia_power: float = 100e-6  # 100 µW

    # Charge Pump
    pump_vin: float = 1.0
    pump_vout: float = 1.5
    pump_eff: float = 0.7

    def total_energy_per_op(self, t_read=60e-9, t_write=150e-9):
        """Total energy per read+write operation."""
        e_adc = self.adc_energy
        e_dac = self.dac_energy
        e_tia = self.tia_power * t_read
        e_pump = self.pump_vout * 10e-6 * t_write / self.pump_eff
        return e_adc + e_dac + e_tia + e_pump

    def power_breakdown(self, t_cycle=210e-9):
        """Power breakdown by component."""
        e_total = self.total_energy_per_op()
        return {
            'ADC': self.adc_energy / e_total * 100,
            'DAC': self.dac_energy / e_total * 100,
            'TIA': (self.tia_power * 60e-9) / e_total * 100,
            'Pump': (self.total_energy_per_op() - self.adc_energy -
                    self.dac_energy - self.tia_power * 60e-9) / e_total * 100,
        }

# Usage
periph = CIMPeripherals()
print(f"Energy per op: {periph.total_energy_per_op() * 1e15:.1f} fJ")
print("Power breakdown:")
for comp, pct in periph.power_breakdown().items():
    print(f"  {comp}: {pct:.1f}%")
```

---

## 7. Open PDKs for Circuit Design

### 7.1 SkyWater SKY130

**URL:** https://github.com/google/skywater-pdk

**Description:** Open-source 130nm PDK from SkyWater via Google.

**ADC/DAC Cells:**
- Standard cells for SAR logic
- Capacitors for DAC
- Comparators available

**Installation:**
```bash
pip install volare
volare enable --pdk sky130 --version latest
```

---

### 7.2 GlobalFoundries GF180MCU

**URL:** https://github.com/google/gf180mcu-pdk

**Description:** Open 180nm MCU PDK.

**Features:**
- Op-amp cells
- Analog primitives
- Mixed-signal IP

---

### 7.3 IHP SG13G2

**URL:** https://github.com/IHP-GmbH/IHP-Open-PDK

**Description:** 130nm BiCMOS with high-frequency options.

**Features:**
- HBT devices for high-speed
- RF passives
- Precision resistors

---

## 8. Our FeCIM Implementation

### 8.1 Module Structure

```
module4-circuits/
├── pkg/peripherals/
│   ├── adc.go          # ADC model (SAR, Flash, Sigma-Delta)
│   ├── dac.go          # DAC model (5-bit for 30 levels)
│   ├── tia.go          # TIA model (current-to-voltage)
│   ├── chargepump.go   # Charge pump (1V → 1.5V)
│   └── analysis.go     # INL/DNL, timing, power analysis
├── pkg/gui/
│   ├── app.go          # Standalone GUI
│   ├── embedded.go     # Embeddable component
│   └── signalflow.go   # Signal chain visualization
└── cmd/
    └── circuits-gui/main.go
```

### 8.2 Key Features

| Feature | Implementation | File |
|---------|---------------|------|
| 5-bit ADC | SAR model with INL/DNL | adc.go |
| 5-bit DAC | Binary-weighted capacitor | dac.go |
| TIA | 10kΩ gain, 100MHz BW | tia.go |
| Charge Pump | 2-stage Dickson | chargepump.go |
| Transfer Function | Full chain analysis | analysis.go |
| Power Breakdown | Component-wise energy | analysis.go |

---

## 9. Comparison Table

| Tool | Type | ADC | DAC | TIA | Simulation | Open PDK |
|------|------|-----|-----|-----|------------|----------|
| **ngspice** | SPICE | ✅ | ✅ | ✅ | Transistor | ✅ |
| **Xyce** | SPICE | ✅ | ✅ | ✅ | Transistor | ✅ |
| **PySpice** | Python+SPICE | ✅ | ✅ | ✅ | Behavioral | ✅ |
| **OpenFASoC** | Generator | ✅ | ✅ | ❌ | Full | ✅ |
| **BAG** | Generator | ✅ | ✅ | ✅ | Full | ✅ |
| **FeCIM (Ours)** | Behavioral | ✅ | ✅ | ✅ | Behavioral | N/A |

---

## 10. Integration Recommendations

### 10.1 Validation Strategy

1. **Behavioral → SPICE:** Validate Go models against ngspice
2. **SPICE → Silicon:** Compare against published data
3. **Cross-check:** Use multiple tools for critical specs

### 10.2 Recommended Tool Stack

| Task | Tool | Rationale |
|------|------|-----------|
| Quick analysis | FeCIM (Go) | Real-time visualization |
| SPICE validation | ngspice | Open, proven |
| Layout | OpenFASoC | Automated flow |
| Process porting | BAG | Generator approach |

### 10.3 Data Exchange Format

```json
{
  "peripheral_config": {
    "adc": {
      "bits": 5,
      "type": "SAR",
      "inl_lsb": 0.5,
      "dnl_lsb": 0.25,
      "conversion_time_ns": 50
    },
    "dac": {
      "bits": 5,
      "vref_low": -1.5,
      "vref_high": 1.5,
      "settle_time_ns": 10
    },
    "tia": {
      "gain_ohm": 10000,
      "bandwidth_hz": 100e6,
      "noise_pa_sqrthz": 1
    },
    "charge_pump": {
      "stages": 2,
      "vin": 1.0,
      "vout": 1.5,
      "efficiency": 0.7
    }
  }
}
```

---

## 11. Community Resources

### 11.1 Forums and Discussions

- **Analog Designers Community:** https://www.reddit.com/r/ECE/
- **EDAboard:** https://www.edaboard.com/
- **SkyWater Slack:** Open PDK discussions

### 11.2 Conferences

- **ISSCC:** ADC/DAC state-of-the-art
- **CICC:** Custom integrated circuits
- **VLSI Symposium:** Mixed-signal design
- **A-SSCC:** Asian solid-state circuits

### 11.3 Key Publications

| Venue | Focus |
|-------|-------|
| IEEE JSSC | Solid-state circuit design |
| IEEE TCAS-I/II | Circuits and systems |
| IEEE TED | Electron devices |

---

---

## Related Documentation

- **[circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md)** — CIM physics: how read/write/compute works
- **[circuits.operations.md](circuits.operations.md)** — 0T1R vs 1T1R architecture operations
- **[circuits.research.md](circuits.research.md)** — Peripheral circuits meta-study
- **[circuits.ELI5.md](circuits.ELI5.md)** — Simple explanations for beginners

---

## Appendix: Quick Installation Guide

```bash
# Core simulation tools
sudo apt install ngspice               # SPICE simulator
pip install PySpice                     # Python interface

# Layout and generation
pip install volare                      # PDK management
git clone https://github.com/idea-fasoc/OpenFASOC

# Our FeCIM visualizer
cd module4-circuits
go build -o circuits-gui ./cmd/circuits-gui
./circuits-gui
```
