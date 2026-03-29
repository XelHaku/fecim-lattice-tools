# Circuit Simulation Tools for FeCIM Peripheral Design

**A comprehensive guide to SPICE simulators and mixed-signal tools for analog circuit design**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source circuit simulation tools essential for designing peripheral circuits (ADC, DAC, TIA, comparators) for Ferroelectric Compute-in-Memory (FeCIM) systems. Whether you're validating behavioral models against transistor-level simulations, optimizing power consumption, or debugging non-ideal circuit behavior, this guide connects you to the right tools.

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

**Why this matters:** Our behavioral models in `module4-circuits` provide fast simulations for teaching, but production circuits require verification against real physics. SPICE simulators bridge that gap.

---

## 1. ngspice - The Industry Standard

**URL:** https://ngspice.sourceforge.io/
**GitHub:** https://git.code.sf.net/p/ngspice/ngspice
**License:** BSD-3-Clause
**Version:** 45.2 (January 2026)
**Language:** C
**Status:** Actively maintained (weekly releases)

### What It Does

ngspice is the open-source reference SPICE simulator. It implements the full SPICE3f5 specification with modern extensions for mixed-signal design.

### Key Features

- **Full SPICE3 compatibility:** All analysis types (DC, AC, transient, noise, Monte Carlo)
- **BSIM3/BSIM4 MOSFET models:** Industry-standard transistor models for SKY130, GF180, 22nm
- **XSPICE extensions:** Mixed-signal simulation (digital + analog)
- **Verilog-A support:** Via OpenVAF compiler for custom device models
- **Python interface:** PySpice library for scripted simulations
- **OSDI models:** Open Source Device Interface for advanced device physics
- **Behavioral modeling:** ABM (A-type) elements for system-level blocks

### Installation

```bash
# Ubuntu/Debian - fastest path
sudo apt install ngspice libngspice0 libngspice0-dev

# macOS
brew install ngspice

# From source (latest features)
git clone https://git.code.sf.net/p/ngspice/ngspice
cd ngspice
./autogen.sh
./configure --with-x=yes --enable-openmp
make -j$(nproc)
sudo make install

# Verify installation
ngspice -v
# Should output: ngspice 45.2
```

### Example 1: Comparator Circuit for CIM

A common peripheral component in FeCIM systems is the comparator. This example simulates a simple sense amplifier comparator used for reading cell states.

```spice
* FeCIM Sense Amplifier Comparator
* 22nm CMOS Technology
.title CIM Sense Amplifier Comparator

* Include technology file (assuming SKY130)
.include "sky130_fd_pr__nfet_01v8.lib"
.include "sky130_fd_pr__pfet_01v8.lib"

* ====== Comparator Subcircuit ======
.subckt comparator inp inn vdd vss out
* Tail current source (NMOS active load)
    Mtail tail vss inn vss nch_svt W=0.5u L=0.15u M=4
* Differential pair
    Mp1 out vdd inp vdd pch_svt W=1.0u L=0.15u M=1
    Mp2 out vdd inn vdd pch_svt W=1.0u L=0.15u M=1
    Mn1 out inp tail vss nch_svt W=0.5u L=0.15u M=1
    Mn2 out inn tail vss nch_svt W=0.5u L=0.15u M=1
* Tail bias (set to 100 µA)
    Ibias vdd tail 100u
* Output load cap (from layout)
    Cout out vss 5f
.ends

* ====== Test Bench ======
* Signal sources
Vdd vdd 0 DC 1.8
Vss vss 0 DC 0

* Test: ramp voltage to see switching
Vinp inp 0 DC 0.9 AC 1 SIN(0.9 0.05 1M)
Vinn inn 0 DC 0.9

* Instantiate comparator
Xcomp inp inn vdd vss out comparator

* Analysis
.tran 0 10u 0 10n

* Measurements
.measure tran delay TRIG v(inn) VAL=0.895 RISE=1 TARG v(out) VAL=0.9 RISE=1
.measure tran prop_delay TRIG v(inp) VAL=0.925 FALL=2 TARG v(out) VAL=0.9 FALL=1

.control
run
set hcopydevtype=postscript
plot v(inp) v(inn) v(out)
quit
.endc

.end
```

**What to expect:** You should see output transitions with ~100ps propagation delay. Adjust W/L ratios to meet speed requirements.

### Example 2: 5-bit SAR ADC Test

SAR (Successive Approximation Register) ADCs are common in mixed-signal chips. This example verifies a SAR comparator chain:

```spice
* SAR ADC 5-bit Comparator
.title SAR ADC with 5-bit Comparator

.include "sky130.lib"

* Binary-weighted capacitor DAC
.subckt cap_dac b4 b3 b2 b1 b0 out vref vss
    C4 b4 out 16p  ; MSB = 16 LSBs
    C3 b3 out 8p
    C2 b2 out 4p
    C1 b1 out 2p
    C0 b0 out 1p   ; LSB
    Cunit out vss 1p   ; Unit cap
.ends

* Comparator (from previous example)
.subckt comparator inp inn vdd vss out
    Mtail tail vss inn vss nch_svt W=0.5u L=0.15u
    Mp1 out vdd inp vdd pch_svt W=1.0u L=0.15u
    Mp2 out vdd inn vdd pch_svt W=1.0u L=0.15u
    Mn1 out inp tail vss nch_svt W=0.5u L=0.15u
    Mn2 out inn tail vss nch_svt W=0.5u L=0.15u
    Ibias vdd tail 100u
    Cout out vss 5f
.ends

* Main ADC test
Vdd vdd 0 DC 1.8
Vin in 0 DC 0 SIN(0.9 0.4 100k)
Vref ref 0 DC 1.0

* Capacitor DAC fed with test codes
* (In reality, SAR logic switches these sequentially)
Vb4 b4 0 DC 0
Vb3 b3 0 DC 1  ; Set bit 3
Vb2 b2 0 DC 0
Vb1 b1 0 DC 1  ; Set bit 1
Vb0 b0 0 DC 0

* DAC and comparator
Xdac b4 b3 b2 b1 b0 dac_out ref 0 cap_dac
Xcomp in dac_out vdd 0 cmp_out comparator

* Transient analysis
.tran 0 50u 0 100n

.control
run
plot v(in) v(dac_out) v(cmp_out)
quit
.endc

.end
```

### Key Analysis Types

| Analysis | Use Case | Time |
|----------|----------|------|
| `.dc` | DC transfer characteristics (gain, offset) | <1 sec |
| `.ac` | Frequency response (bandwidth, gain) | <1 sec |
| `.tran` | Transient behavior (settling, noise) | 1-10 sec |
| `.noise` | Noise spectral density (NEF metric) | <1 sec |
| `.meas` | Extract measurements (delay, power) | Auto |

### Integration with FeCIM Module 4

```go
// In module4-circuits/pkg/peripherals/adc.go

// Example: Read SPICE simulation results
func ReadNgspiceResults(filepath string) ([]float64, error) {
    // Parse ngspice rawfile output
    // ngspice -b script.cir -o output.log
    // Extract measurements for validation
}
```

### Performance Tips

- **Convergence issues?** Add `.option gmin=1e-12` and `.option abstol=1e-14`
- **Slow simulation?** Use XSPICE `d_switch` for digital, not full transistor models
- **Large circuits?** Enable OpenMP: `./configure --enable-openmp` then `.option ompnum=4`

---

## 2. Xyce - Parallel SPICE for Large Circuits

**URL:** https://xyce.sandia.gov/
**GitHub:** https://github.com/Xyce/Xyce
**License:** GPL-3.0
**Version:** 7.8 (January 2026)
**Language:** C++
**Status:** Actively maintained (Sandia National Labs)

### What It Does

Xyce is a parallel circuit simulator designed for simulating circuits with hundreds of millions of transistors. While ngspice is serial, Xyce scales across multiple CPU cores and HPC clusters.

### Key Features

- **Massive parallelism:** MPI-based for distributed computing
- **Harmonic balance analysis:** For periodic steady-state circuits
- **Sensitivity analysis:** Device-level parametric sweeps
- **VERILOG-A support:** Same as ngspice
- **Device noise:** Thermal, shot, flicker noise simulation
- **Behavioral modeling:** Arbitrary behavioral sources

### Installation

```bash
# Ubuntu (easier than building)
sudo apt install xyce

# From source (requires Trilinos)
# More complex - see https://xyce.sandia.gov/documentation/
git clone https://github.com/Xyce/Xyce
cd Xyce && mkdir build && cd build
cmake -DCMAKE_INSTALL_PREFIX=/usr/local ..
make -j$(nproc)
sudo make install
```

### Example: Full Crossbar + Peripheral Simulation

Xyce excels when you have a 64×64 crossbar (4096 cells) plus all peripheral circuits. ngspice would struggle; Xyce handles it:

```spice
* 64x64 Memristor Crossbar + ADC/DAC Peripherals
* Xyce netlist (can run on GPU cluster)

.title FeCIM Crossbar System

* Include parametric models
.param W=64 H=64
.param Roff=10k Ron=1k

* Crossbar array (4096 resistors)
* Simplified generation - in practice, use Python script
.subckt crossbar_64x64
* ... define 64 rows × 64 columns of memristors
* ... each: Rxy row_x col_y memristor_model
.ends

* ADC array (64 outputs, one per column)
* ... 64 SAR ADCs in parallel

* DAC array (64 inputs, one per row)
* ... 64 DACs in parallel

.tran 0 1u 0 1n

.end
```

**When to use Xyce:**
- Simulating full 64×64+ crossbars with peripherals
- HPC access available (typically academic)
- Need sensitivity analysis across thousands of parameters
- Harmonic balance for RF/analog blocks

**Limitation:** steeper learning curve, less documentation than ngspice.

---

## 3. QUCS-S - GUI for ngspice/Xyce

**URL:** https://ra3xdh.github.io/
**GitHub:** https://github.com/ra3xdh/qucs-s
**License:** GPL-2.0
**Version:** 1.8 (January 2026)
**Language:** C++ with Qt6

### What It Does

QUCS-S provides a graphical schematic editor and waveform viewer, using ngspice or Xyce as the backend simulator.

### Key Features

- **Schematic capture:** Drag-and-drop circuit design
- **Component library:** Resistors, capacitors, transistors, op-amps
- **Waveform viewer:** Interactive plots with cursor measurements
- **Model integration:** Easily load .lib files
- **S-parameter simulation:** For RF applications
- **Multiple backends:** Use ngspice, Xyce, or SPICE OPUS

### Installation

```bash
# Ubuntu/Debian
sudo apt install qucs-s

# macOS
brew install qucs-s

# Build from source
git clone https://github.com/ra3xdh/qucs-s.git
cd qucs-s
mkdir build && cd build
cmake .. -DCMAKE_INSTALL_PREFIX=/usr/local
make -j$(nproc)
sudo make install
```

### Workflow Example

1. **Launch QUCS-S**
   ```bash
   qucs-s &
   ```

2. **Create new project** → Choose ngspice backend

3. **Draw schematic:**
   - Add voltage source (Vdd = 1.8V)
   - Add resistor (1kΩ)
   - Add capacitor (1µF)
   - Connect to ground

4. **Add simulation:**
   - Right-click → Insert Simulation
   - Choose: Transient analysis, 0 to 10ms

5. **Run:** Click "Simulate" button

6. **View results:** Waveform window shows voltage vs. time

7. **Export:** Save as .cir for batch processing

### FeCIM Integration

```bash
# Generate schematic for SAR ADC
# Use QUCS-S GUI, then export to:
# Export → SPICE Netlist → sar_adc.cir

# Then run batch ngspice:
ngspice -b sar_adc.cir -o results.log
```

**Advantage:** Non-engineers can draw circuits without writing SPICE syntax.

---

## 4. PySpice - Python Scripted Simulations

**URL:** https://github.com/PySpice-org/PySpice
**License:** GPL-3.0
**Version:** 1.6 (January 2026)
**Language:** Python
**Status:** Actively maintained

### What It Does

PySpice is a Python wrapper around ngspice/Xyce. Instead of writing `.cir` files, you write Python. Perfect for parametric sweeps and automated analysis.

### Installation

```bash
# Via pip
pip install PySpice

# Verify
python -c "from PySpice.Spice.Netlist import Circuit; print('OK')"
```

### Example 1: Transient Behavior of DAC + TIA

This Python script simulates a 5-bit DAC driving a transimpedance amplifier:

```python
from PySpice.Spice.Netlist import Circuit
from PySpice.Unit import *
from PySpice.Spice.NgSpice.Shared import NgSpiceShared
import numpy as np
import matplotlib.pyplot as plt

# Create circuit
circuit = Circuit('DAC + TIA Test')

# Include technology file
circuit.include('/path/to/sky130.lib')

# Power supplies
circuit.V('dd', 'vdd', circuit.gnd, 1.8@u_V)
circuit.V('ss', 'vss', circuit.gnd, 0@u_V)

# DAC (5-bit binary-weighted capacitor DAC)
# Capacitor DAC with reference voltage
circuit.V('ref', 'vref', circuit.gnd, 1.0@u_V)
circuit.C(1, 'dac_out', circuit.gnd, 10@u_pF)  # Output cap

# Simulate DAC output as voltage source
# (In practice, capacitor DAC would be detailed circuit)
circuit.V('dac', 'dac_out', circuit.gnd,
          'PWL(0 0.5V 100n 0.6V 200n 0.7V 300n 0.8V 400n 0.9V)')

# Op-amp (ideal behavioral model)
# TIA feedback network
circuit.R('f', 'dac_out', 'tia_out', 10@u_kΩ)  # Feedback resistor
circuit.C('f', 'dac_out', 'tia_out', 0.1@u_pF)  # Feedback cap for compensation

# Ideal op-amp behavioral model (VCVS)
# Gain = 10^5, single pole at 100 MHz
circuit.VCVS('opamp', 'tia_out', circuit.gnd, 'dac_out', circuit.gnd, 100000)

# Transient simulation
simulator = circuit.simulator(temperature=25@u_degC, nominal_temperature=25@u_degC)
analysis = simulator.transient(step_time=1@u_ns, end_time=500@u_ns)

# Extract and plot
time = np.array(analysis.time)
v_dac = np.array(analysis['dac_out'])
v_out = np.array(analysis['tia_out'])

plt.figure(figsize=(10, 6))
plt.subplot(2, 1, 1)
plt.plot(time * 1e9, v_dac, 'b-', label='DAC Output')
plt.ylabel('Voltage (V)')
plt.legend()
plt.grid(True)

plt.subplot(2, 1, 2)
plt.plot(time * 1e9, v_out, 'r-', label='TIA Output')
plt.xlabel('Time (ns)')
plt.ylabel('Voltage (V)')
plt.legend()
plt.grid(True)

plt.tight_layout()
plt.savefig('dac_tia_response.png', dpi=150)
print("Plot saved to dac_tia_response.png")
```

**Run it:**
```bash
python dac_tia_simulation.py
```

### Example 2: ADC INL/DNL Analysis

Characterize ADC linearity across all codes:

```python
from PySpice.Spice.Netlist import Circuit
from PySpice.Unit import *
import numpy as np

def simulate_adc_linearity(n_bits=5):
    """Characterize ADC INL/DNL"""

    circuit = Circuit('ADC Linearity Test')
    circuit.include('/path/to/sky130.lib')

    # Power supplies
    circuit.V('dd', 'vdd', circuit.gnd, 1.8@u_V)

    # Reference voltage
    circuit.V('ref', 'vref', circuit.gnd, 1.0@u_V)

    # Simulate ADC as behavioral block
    # Input voltage sweep
    circuit.V('in', 'vin', circuit.gnd, 'DC 0 AC 1')

    # ADC output (simulate as multi-bit DAC)
    # This is where you'd instantiate actual ADC schematic
    circuit.VCVS('adc_b0', 'b0', circuit.gnd, 'vin', circuit.gnd, 1)  # Simplified

    # Simulate at each input code
    results = {}
    n_codes = 2 ** n_bits

    for code in range(n_codes):
        vin_ideal = code / (n_codes - 1) * 1.0  # 0 to Vref

        # Modify circuit and simulate
        # (In real usage, sweep voltage and extract output transitions)

        results[code] = vin_ideal

    # Calculate INL/DNL
    inl = np.zeros(n_codes)
    dnl = np.zeros(n_codes)

    for code in range(n_codes):
        # INL = actual_code - ideal_code
        inl[code] = (results[code] - code / (n_codes - 1) * 1.0)

        if code > 0:
            # DNL = (step_i - step_ideal)
            dnl[code] = (results[code] - results[code-1]) - (1.0 / (n_codes - 1))

    print(f"ADC {n_bits}-bit Linearity Analysis")
    print(f"MAX INL: {np.max(np.abs(inl)):.3f} LSB")
    print(f"MAX DNL: {np.max(np.abs(dnl)):.3f} LSB")
    print(f"Monotonic: {np.all(dnl > -1.0)}")

    return inl, dnl

# Run analysis
inl, dnl = simulate_adc_linearity(n_bits=5)
```

### PySpice + Module 4 Integration

```go
// In module4-circuits/pkg/peripherals/validation.go

// Call Python simulation from Go
func ValidateADCAgainstPySpice(config ADCConfig) error {
    // Write config to JSON
    // Shell out to Python script:
    //   python validate_adc_pyspice.py --config adc.json
    // Read back measurement results
    return nil
}
```

---

## 5. Ahkab - Pure Python SPICE (Advanced)

**URL:** https://github.com/ahkab/ahkab
**License:** GPL-2.0
**Version:** 0.18
**Language:** Python
**Status:** Unmaintained (last update 2015)

### What It Does

Ahkab is a SPICE simulator written entirely in Python. No C/C++ compilation needed.

### Features

- **Pure Python:** Cross-platform, no build required
- **Symbolic analysis:** Can derive transfer functions symbolically
- **DC, AC, transient, noise:** Full analysis suite
- **Educational:** Easy to understand internals

### When to Use

- Educational purposes (learning SPICE internals)
- Quick prototyping without compilation
- Symbolic analysis of small circuits

### Example

```python
from ahkab import Circuit, printing

# Create circuit
circuit = Circuit('RC Filter')

# Add components
circuit.add_resistor('R1', 'in', 'out', 10@u_kΩ)
circuit.add_capacitor('C1', 'out', circuit.gnd, 1.59@u_nF)

# Voltage source
circuit.add_voltage_source('Vin', 'in', circuit.gnd, 1@u_V)

# AC analysis
printing.print_circuit(circuit)
```

**Note:** Ahkab is no longer actively maintained. Use ngspice for production work.

---

## 6. GnuCap - Interactive SPICE

**URL:** http://gnucap.org/
**License:** GPL-3.0
**Language:** C++
**Status:** Maintained (academic project)

### What It Does

GnuCap is an interactive SPICE simulator with a focus on extensibility. You can define custom device models via plugins.

### Key Features

- **Interactive command line:** Real-time circuit modifications
- **Plugin architecture:** Load custom device models
- **Verilog-A support:** Via external compilers
- **Behavioral blocks:** Complex behavioral modeling

### Example

```bash
# Interactive session
$ gnucap
gnucap> read circuit.cir
gnucap> transient 0 10u
gnucap> plot v(out)
gnucap> set seed 42 ; set noise=true
gnucap> transient 0 10u
gnucap> plot v(out)
gnucap> quit
```

**Use case:** When you need fine-grained control and want to experiment interactively. Less common than ngspice but valuable for research.

---

## 7. CircuitJS - Browser-Based Simulator

**URL:** https://www.falstad.com/circuit/circuitjs.html
**License:** GPL-2.0
**Language:** JavaScript
**Platform:** Web browser (no installation)

### What It Does

CircuitJS is a real-time circuit simulator that runs in your browser. Simulate, drag components, and watch behavior instantly.

### Features

- **Real-time visualization:** See current/voltage flowing instantly
- **Component library:** Resistors, capacitors, transistors, logic gates
- **Animation:** Electron flow visualization
- **Educational:** Perfect for teaching

### Example Workflow

1. Open https://www.falstad.com/circuit/circuitjs.html
2. Add components (drag from left panel)
3. Draw circuit (connect with wires)
4. Run simulation (automatic)
5. Oscilloscope probe to measure voltages
6. Export as JSON or PNG

### FeCIM Use Case

Good for educational demos but not suitable for:
- Large circuits (>100 components)
- Detailed device physics
- Accurate parasitics

**Best for:** Visual explanations in presentations, student projects.

---

## 8. KiCad + ngspice Integration

**URL:** https://www.kicad.org/
**License:** GPL-3.0
**Version:** KiCad 9.0 (January 2026)
**Language:** C++ with Python scripting

### What It Does

KiCad is a professional-grade EDA suite. Draw schematics, simulate with ngspice backend, and generate PCB layouts—all in one tool.

### Key Features

- **Schematic editor:** Professional component placement and routing
- **Symbol/footprint libraries:** Thousands of standard components
- **ngspice integration:** Seamless simulation from schematic
- **Python API:** Automate schematic generation
- **PCB design:** Full layout tools (beyond scope here)

### Installation

```bash
# Ubuntu
sudo apt install kicad

# macOS
brew install kicad

# Verify
kicad --version
```

### Workflow: SAR ADC Schematic → Simulation

**Step 1: Draw Schematic in KiCad**

1. File → New Schematic
2. Add components:
   - Voltage sources (Vdd, Vin, Vref)
   - Comparator symbol
   - Logic gates (for SAR control)
   - DAC symbol (library)

3. Connect with wires

**Step 2: Add Simulation Annotations**

```
# Right-click on component → Properties → SPICE node sequence
# For comparator: set model to "comparator_macro"
```

**Step 3: Export to SPICE**

```
Tools → Simulate → Export SPICE Netlist
# Generates: schematic.cir
```

**Step 4: Run ngspice**

```bash
ngspice schematic.cir
```

### Python API Example

Programmatically generate ADC schematics:

```python
#!/usr/bin/env python3
import pcbnew
import os

# Create new PCB
board = pcbnew.CreateEmptyBoard()

# Add ADC schematic components programmatically
# (This is schematic-level, would use schematic API)

# For now, KiCad schematic API is limited
# Easier to hand-draw, then export
```

---

## 9. OpenVAF - Verilog-A Device Models

**URL:** https://github.com/pascalkuthe/OpenVAF
**License:** GPL-3.0
**Version:** 0.23 (January 2026)
**Language:** Rust

### What It Does

OpenVAF compiles Verilog-A device models to binary formats that ngspice and Xyce can load. Enables custom ferroelectric cell models.

### Key Features

- **Verilog-A compilation:** Standard HDL for analog models
- **ngspice integration:** Via OSDI (Open Source Device Interface)
- **Xyce support:** Similar integration
- **Physics modeling:** Implement Preisach, Tanh, or custom hysteresis

### Installation

```bash
# Download binary (easiest)
wget https://github.com/pascalkuthe/OpenVAF/releases/download/v0.23.0/openvaf-linux-x86_64.tar.gz
tar xzf openvaf-linux-x86_64.tar.gz
sudo mv openvaf /usr/local/bin/

# Verify
openvaf --version
```

### Example: FeCIM Cell Model

Create a behavioral ferroelectric cell model in Verilog-A:

**File: fecim_cell.va**

```verilog
// FeCIM Ferroelectric Memory Cell Model
// Behavioral hysteresis with 30 analog states

`include "constants.vams"
`include "disciplines.vams"

module fecim_cell(p, n);
    inout p, n;
    electrical p, n;

    // Material parameters (HfO₂-ZrO₂ superlattice)
    parameter real Pr = 20e-6 from [10e-6 : 40e-6];      // Remnant polarization (C/cm²)
    parameter real Ec = 1.2e6 from [0.6e6 : 1.5e6];      // Coercive field (V/cm)
    parameter real thickness = 10e-9 from [5e-9 : 50e-9]; // Film thickness (m)
    parameter real area = 100e-12 from [1e-12 : 1e-9];    // Cell area (m²)
    parameter real levels = 30 from [10 : 100];            // Analog states

    // Physical parameters
    real P;           // Polarization state (0 to Pr)
    real E;           // Electric field
    real dP;          // Polarization change rate
    real Q;           // Stored charge

    analog begin
        // Electric field from applied voltage
        E = V(p, n) / thickness;

        // Polarization update (first-order Landau-Khalatnikov)
        // dP/dt = -Ec * (P - tanh(α*E)) with saturation
        dP = -Ec * (P - Pr * tanh(E / Ec)) / (1e-9);  // 1ns time constant

        // Integrate polarization with state quantization
        P = idt(dP, Pr/2);  // Initialize to Pr/2
        P = max(0, min(Pr, P));  // Clip to [0, Pr]

        // Quantize to 30 levels
        Q = round(P / Pr * (levels - 1)) / (levels - 1) * Pr;

        // Output: charge is integral of polarization
        I(p, n) <+ ddt(Q * area * 1e4);  // Convert to Amperes
    end
endmodule
```

**Compile to OSDI:**

```bash
openvaf fecim_cell.va --output fecim_cell.osdi
```

**Use in ngspice:**

```spice
* Load OSDI model
.osdi fecim_cell.osdi

* Instantiate FeCIM cell
Xcell p n fecim_cell

* Apply voltage pulse to program/read
Vpulse p 0 PULSE(0 2V 100n 10n 10n 40n 100n)

.tran 0 1u 0 1n
.end
```

### FeCIM Module 1 Integration

```go
// In module1-hysteresis/pkg/ferroelectric/verilog_a.go

func CompileVerilogA(vaFile string) (string, error) {
    // Shell to OpenVAF
    cmd := exec.Command("openvaf", vaFile, "--output", "model.osdi")
    return cmd.Output()
}

// Then use OSDI model in ngspice for validation
```

---

## 10. Tool Comparison Matrix

| Feature | ngspice | Xyce | PySpice | QUCS-S | Ahkab | OpenVAF |
|---------|---------|------|---------|--------|-------|---------|
| **SPICE Standard** | Full | Full | Partial | Full | Partial | N/A |
| **License** | BSD-3 | GPL-3 | GPL-3 | GPL-2 | GPL-2 | GPL-3 |
| **Installation Ease** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Large Circuits** | Medium | ⭐⭐⭐⭐⭐ | Medium | Small | Small | N/A |
| **GUI** | CLI | CLI | No | Yes | CLI | No |
| **Python Integration** | PySpice | Xyce API | Native | No | Native | Via script |
| **Verilog-A** | OSDI | Yes | Yes | Yes | No | Native |
| **Noise Analysis** | Yes | Yes | Yes | Yes | Yes | Via model |
| **Active Development** | Yes | Yes | Yes | Yes | No | Yes |
| **Production Ready** | ✅ | ✅ | ✅ | ✅ | ⚠️ | ✅ |

---

## 11. Recommended Workflows

### Workflow 1: Quick ADC Validation (30 minutes)

**Goal:** Verify SAR ADC transient response against ngspice

1. **Write SPICE netlist** (`sar_adc.cir`) with comparator + DAC
2. **Run ngspice:**
   ```bash
   ngspice -b sar_adc.cir -o results.log
   ```
3. **Parse results:**
   ```python
   # Python script to extract timing
   with open('results.log') as f:
       data = [line for line in f if 'measure' in line]
   ```
4. **Compare to Go model:** Update `module4-circuits/adc.go` if needed

**Tools:** ngspice + text editor + Python

---

### Workflow 2: Full Peripheral Design (1-2 days)

**Goal:** Design complete ADC + DAC + TIA system in 130nm technology

1. **KiCad:** Draw professional schematics (comparator, op-amp, DAC)
2. **Export:** Save SPICE netlist
3. **PySpice:** Automate parametric sweeps (W/L ratios, bias currents)
4. **Xyce:** Simulate with extracted parasitics from layout
5. **Results:** Power, area, performance metrics

**Tools:** KiCad → PySpice → Xyce

---

### Workflow 3: Custom FeCIM Model (Advanced, 1 week)

**Goal:** Create ferroelectric cell model matching experimental data

1. **Literature review:** Find Preisach parameters from papers
2. **Verilog-A:** Write hysteresis model in `fecim_cell.va`
3. **OpenVAF:** Compile to OSDI
4. **ngspice:** Validate against measured P-E curves
5. **Integration:** Import into `module1-hysteresis`

**Tools:** OpenVAF → ngspice → Go integration

---

## 12. Installation Quick Start

### Install All Tools (Ubuntu/Debian)

```bash
# SPICE simulators
sudo apt install ngspice xyce qucs-s

# Python interfaces
pip install PySpice ngspice

# KiCad (full EDA suite)
sudo apt install kicad

# OpenVAF (Verilog-A compiler)
wget https://github.com/pascalkuthe/OpenVAF/releases/download/v0.23.0/openvaf-linux-x86_64.tar.gz
tar xzf openvaf-linux-x86_64.tar.gz
sudo mv openvaf /usr/local/bin/

# Verify everything
ngspice --version
pip show PySpice
kicad --version
openvaf --version
```

### Install on macOS

```bash
# Homebrew
brew install ngspice xyce kicad

# Python
pip install PySpice

# OpenVAF (manual download)
# wget ... (same as above)
```

### Install on Windows

- **ngspice:** https://ngspice.sourceforge.io/ (binary installer)
- **KiCad:** https://www.kicad.org/download/
- **PySpice:** `pip install PySpice`
- **OpenVAF:** Build from source (Rust required)

---

## 13. Integration with FeCIM Modules

### Module 1 (Hysteresis)

**Validation:** Compare Preisach Go implementation against:
- ngspice with Verilog-A hysteresis model
- Published experimental P-E curves

```bash
# Generate P-E curve from ngspice
ngspice -b preisach_model.cir -o preisach_out.log

# Python: extract and compare
python validate_hysteresis.py --spice preisach_out.log --go module1_output.json
```

### Module 2 (Crossbar)

**Integration:** Validate MVM (matrix-vector multiply) against:
- ngspice crossbar array with real transistor sneak-path losses
- Compare to module2 behavioral model

```python
# In module2-crossbar/pkg/crossbar/validation.py
import subprocess

def run_spice_crossbar_simulation():
    """Simulate 8x8 crossbar in ngspice, extract MVM accuracy"""
    result = subprocess.run(['ngspice', '-b', 'crossbar_8x8.cir'],
                          capture_output=True, text=True)
    return parse_output(result.stdout)
```

### Module 4 (Circuits)

**Direct integration:** Replace behavioral models with ngspice validation

```go
// module4-circuits/pkg/peripherals/validation.go

func ValidateADCAgainstNgspice(bits int, inl, dnl float64) error {
    // Generate SPICE netlist with given parameters
    netlist := generateSARADCNetlist(bits, inl, dnl)

    // Run ngspice
    cmd := exec.Command("ngspice", "-b", "-")
    cmd.Stdin = strings.NewReader(netlist)
    output, err := cmd.Output()

    // Parse output and verify
    measurements := parseNgspiceOutput(output)
    if measurements.PropagationDelay > MAX_DELAY {
        return fmt.Errorf("ADC delay too high: %v", measurements.PropagationDelay)
    }
    return nil
}
```

---

## 14. Common Issues & Solutions

### Issue: ngspice "convergence failed"

**Cause:** Circuit has instability or unrealistic component values

**Solution:**
```spice
.option gmin=1e-12       ; Reduce minimum conductance
.option abstol=1e-14     ; Tighter absolute tolerance
.option reltol=1e-5      ; Tighter relative tolerance
.option maxstep=1n       ; Force smaller time steps
```

### Issue: Slow simulation (hours instead of seconds)

**Cause:** Model complexity or parasitic extraction

**Solution:**
1. Use simplified models (resistors only, no diode physics)
2. Enable parallel processing (ngspice OpenMP):
   ```
   .option ompnum=4
   ```
3. Reduce circuit size for testing

### Issue: "Unknown device type: osdi"

**Cause:** ngspice compiled without OSDI support

**Solution:**
```bash
# Recompile ngspice with OSDI
./configure --enable-osdi
make -j$(nproc)
sudo make install
```

### Issue: PySpice import errors

**Cause:** ngspice library not found

**Solution:**
```bash
# Install ngspice first
sudo apt install libngspice0 libngspice0-dev

# Then reinstall PySpice
pip install --upgrade PySpice
```

---

## 15. Further Reading

### Official Documentation

| Tool | URL |
|------|-----|
| **ngspice** | https://ngspice.sourceforge.io/docs.html |
| **Xyce** | https://xyce.sandia.gov/documentation/ |
| **PySpice** | https://pyspice.fabrice-salvaire.fr/ |
| **KiCad** | https://docs.kicad.org/ |
| **OpenVAF** | https://openvaf.semimod.de/ |
| **QUCS-S** | https://ra3xdh.github.io/ |

### Key Papers & Resources

- **SPICE Algorithm:** L. W. Nagel, D. O. Pederson (1973) - "SPICE (Simulation Program with Integrated Circuit Emphasis)"
- **Ferroelectric Modeling:** Preisach, Tanh, and TDGL models in `docs/2-learn/module1-hysteresis/physics.md`
- **CIM Circuits:** `docs/4-research/internal-analysis/circuits.CIM-fundamentals.md`

### Example Repositories

- **Berkeley Analog Generator:** https://github.com/ucb-art/BAG_framework
- **OpenFASoC (ADC/DAC generators):** https://github.com/idea-fasoc/OpenFASOC
- **SKY130 PDK:** https://github.com/google/skywater-pdk

---

## 16. Related Documentation

- **[circuits.CIM-fundamentals.md](../internal-analysis/circuits.CIM-fundamentals.md)** — Physics of CIM peripheral circuits
- **[Module 1 Physics](../../2-learn/module1-hysteresis/physics.md)** — Ferroelectric physics models
- **[Module 6 EDA](../../2-learn/module6-eda/README.md)** — EDA tools and learning resources

---

## Quick Reference Table

| Task | Best Tool | Time | Difficulty |
|------|-----------|------|------------|
| Draw circuit GUI | KiCad / QUCS-S | 15 min | Easy |
| Simulate transient | ngspice | 1 min | Easy |
| Parametric sweep | PySpice | 5 min | Medium |
| Large circuit (1000+) | Xyce | 10 min | Hard |
| Custom device model | OpenVAF | 30 min | Hard |
| Real-time animation | CircuitJS | 5 min | Easy |
| Production design | ngspice + KiCad | 2-3 days | Hard |

---

**Last Updated:** January 27, 2026
**Maintained by:** FeCIM Lattice Tools Project
**Feedback:** See `docs/about/Contributing.md`
