# Data Acquisition & Instrumentation Tools

Python libraries and frameworks for controlling lab equipment in ferroelectric device characterization, including source measurement units (SMUs), oscilloscopes, pulse generators, and lock-in amplifiers.

---

## Overview

Characterizing FeCIM devices requires precise measurement and control of hardware. These open-source tools bridge the gap between Python and physical lab equipment via standard communication protocols (GPIB, USB, Ethernet).

### Key Concepts

| Term | Definition |
|------|-----------|
| **VISA** | Virtual Instrument Software Architecture - standardized protocol for lab equipment communication |
| **SMU** | Source Measurement Unit - simultaneously applies voltage and measures current |
| **DAQ** | Data Acquisition card - high-speed, multi-channel measurement hardware |
| **USBTMC** | USB Test & Measurement Class - USB standard for lab instruments |
| **IVI** | Interchangeable Virtual Instruments - cross-vendor standard drivers |

---

## 1. PyVISA (Essential Foundation)

**GitHub:** https://github.com/pyvisa/pyvisa
**License:** MIT
**Version:** 1.14+
**Backends:** NI-VISA, PyVISA-py (pure Python)

### What It Does
PyVISA is the lowest-level Python interface to lab equipment. It handles communication protocol translation, leaving hardware-specific details to backend drivers.

### Architecture
```
Your Python Code
    ↓
PyVISA (protocol wrapper)
    ↓
Backend (NI-VISA or PyVISA-py)
    ↓
Hardware Driver (GPIB, USB, Ethernet)
    ↓
Physical Equipment
```

### Installation

```bash
# Basic installation
pip install pyvisa

# With pure Python backend (no NI-VISA required)
pip install pyvisa pyvisa-py

# Optional: install usbtmc for USB devices
pip install pyusb
```

### Basic Usage

```python
import pyvisa

# Create resource manager
rm = pyvisa.ResourceManager()

# List all connected instruments
print(rm.list_resources())
# Output: ('GPIB0::24::INSTR', 'USBTMC0::0x05E6::0x2450::...', ...)

# Connect to Keithley SMU on GPIB address 24
smu = rm.open_resource('GPIB0::24::INSTR')

# Identification query
print(smu.query('*IDN?'))
# Output: KEITHLEY INSTRUMENTS INC.,MODEL 2400,3Z457YY,C33...,03

# Set voltage to 5V
smu.write('SOUR:VOLT 5')

# Measure current
current = float(smu.query('MEAS:CURR?'))
print(f"Current: {current:.3e} A")

# Close connection
smu.close()
```

### Why This Layer Matters
- Abstracts GPIB/USB/Ethernet complexity
- Enables vendor-independent instrument drivers
- Foundation for higher-level tools (PyMeasure, QCoDeS)
- Direct SCPI command control when needed

### Limitations
- Requires manual SCPI command learning per instrument
- No built-in error handling or parameter validation
- Low-level: verbose code for complex experiments

---

## 2. PyMeasure (Recommended for Most Users)

**GitHub:** https://github.com/pymeasure/pymeasure
**License:** MIT
**Version:** 0.9+
**Supported Instruments:** 100+ devices with pre-built drivers

### What It Does
PyMeasure wraps PyVISA with high-level, Pythonic instrument drivers. Write clean code without memorizing SCPI commands.

### Key Features

| Feature | Benefit |
|---------|---------|
| **Pre-built drivers** | Keithley, Agilent, HP, Stanford Research, etc. |
| **Parameter API** | Properties, validation, unit handling |
| **GUI Framework** | Qt-based experiment designer |
| **Procedure Pattern** | Structured measurement workflows |
| **Non-blocking** | Thread-safe concurrent measurements |

### Supported Instruments for FeCIM

```
SMU:
  - Keithley 2400, 2450, 2600 series
  - Keithley 2601A, 2601B (dual channels)

Oscilloscope:
  - Agilent/Keysight DSO/MSO 7000/8000/9000 series
  - Tektronix scope models

Function Generator:
  - HP/Agilent 33500B series
  - Agilent 81150A

Lock-In Amplifier:
  - Stanford Research SR830, SR865, SR860
```

### Installation

```bash
pip install pymeasure

# With GUI support
pip install pymeasure[UI]

# Development install
git clone https://github.com/pymeasure/pymeasure
cd pymeasure
pip install -e .
```

### Basic Usage: P-E Loop Measurement

```python
from pymeasure.instruments.keithley import Keithley2400
import numpy as np
import matplotlib.pyplot as plt

# Create instrument object
smu = Keithley2400("GPIB0::24::INSTR")

# Configure for voltage source operation
smu.apply_voltage()
smu.source_voltage_range = 10  # Set to 10V range
smu.compliance_current = 0.1   # 100 mA compliance

try:
    # Enable the source
    smu.enable_source()

    # Define voltage sweep: 0 → +5V → -5V → 0
    voltages = np.concatenate([
        np.linspace(0, 5, 50),      # Forward
        np.linspace(5, -5, 100),    # Reverse
        np.linspace(-5, 0, 50)      # Return
    ])

    currents = []

    # Measure current at each voltage step
    for voltage in voltages:
        smu.source_voltage = voltage
        current = smu.current
        currents.append(current)
        print(f"V={voltage:6.2f}V, I={current:.3e}A")

    # Plot P-E loop (current vs voltage proxy for polarization)
    plt.figure(figsize=(10, 6))
    plt.plot(voltages, np.array(currents) * 1e6, 'b-', linewidth=2)
    plt.xlabel('Voltage (V)', fontsize=12)
    plt.ylabel('Current (μA)', fontsize=12)
    plt.title('P-E Loop Measurement (SMU)', fontsize=14)
    plt.grid(True, alpha=0.3)
    plt.show()

finally:
    # Clean shutdown
    smu.disable_source()
    smu.shutdown()
```

### Using Measurement Procedures (Structured Experiments)

```python
from pymeasure.experiment import Procedure, Results, Worker
from pymeasure.instruments.keithley import Keithley2400
import numpy as np
from datetime import datetime

class VoltageSweeproc(Procedure):
    """Sweep voltage and measure current."""

    # Define parameters with units and ranges
    start_voltage = FloatParameter('Start Voltage', units='V', minimum=-10, maximum=10, default=0)
    stop_voltage = FloatParameter('Stop Voltage', units='V', minimum=-10, maximum=10, default=5)
    num_points = IntegerParameter('Number of Points', minimum=10, maximum=1000, default=50)
    compliance = FloatParameter('Compliance Current', units='mA', minimum=0.1, maximum=1000, default=100)

    # Output columns
    DATA_COLUMNS = ['Voltage (V)', 'Current (A)']

    def startup(self):
        """Initialize instrument."""
        self.smu = Keithley2400(self.instrument_id)
        self.smu.apply_voltage()
        self.smu.source_voltage_range = 10
        self.smu.compliance_current = self.compliance / 1000  # Convert mA to A
        self.smu.enable_source()

    def execute(self):
        """Run the measurement loop."""
        voltages = np.linspace(self.start_voltage, self.stop_voltage, self.num_points)

        for i, v in enumerate(voltages):
            if self.should_stop():
                break

            self.smu.source_voltage = v
            i_meas = self.smu.current

            self.emit('results', {
                'Voltage (V)': v,
                'Current (A)': i_meas
            })

            self.emit('progress', 100 * i / len(voltages))

    def shutdown(self):
        """Clean up."""
        self.smu.disable_source()
        self.smu.shutdown()

# Run the procedure with GUI
if __name__ == '__main__':
    procedure = VoltageSweepProcedure()
    results = Results(procedure, "data/sweep_results.csv")
    worker = Worker(procedure, results)
    worker.start()
```

### GUI Experiment Designer
PyMeasure provides Qt-based GUI for parameter input, live plotting, and data export without writing UI code.

---

## 3. QCoDeS (Microsoft - Advanced Labs)

**GitHub:** https://github.com/microsoft/Qcodes
**License:** MIT
**Version:** 0.50+
**Integration:** Jupyter, SQLite, xarray

### What It Does
Enterprise-grade data acquisition framework with built-in data management, live plotting, and Jupyter integration. Ideal for research labs running complex multi-instrument experiments.

### Key Features

| Feature | Details |
|---------|---------|
| **Parameter-based API** | Type-safe, unit-aware parameters |
| **SQLite backend** | Automatic experiment logging and querying |
| **Jupyter-native** | Live plotting in notebooks |
| **Xarray integration** | Multi-dimensional data handling |
| **100+ drivers** | Comprehensive instrument library |
| **Measurement loops** | High-performance data collection |

### Installation

```bash
pip install qcodes

# With plotting support
pip install qcodes[matplotlib,jupyter]

# For Slack notifications
pip install qcodes[slack]
```

### Example: Multi-Instrument P-E Loop Measurement

```python
from qcodes import Instrument, VisaInstrument, Parameter
from qcodes.instrument_drivers.stanford_research import SR830
from qcodes.instrument_drivers.keithley import Keithley2450
from qcodes.dataset import Measurement
from qcodes.plots.qcodes_pyqtgraph import QtPlot
import numpy as np
import matplotlib.pyplot as plt

# Initialize instruments
smu = Keithley2450('smu', 'GPIB0::24::INSTR')
lockin = SR830('lockin', 'GPIB0::8::INSTR')

# Configure SMU
smu.voltage_range(10)
smu.current_compliance(0.1)

# Configure lock-in
lockin.frequency(1000)  # 1 kHz drive
lockin.sensitivity(0.01)  # 10 mV sensitivity
lockin.time_constant(0.1)  # 100 ms time constant

# Create measurement with data storage
meas = Measurement()
meas.register_parameter(smu.voltage)
meas.register_parameter(smu.current)
meas.register_parameter(lockin.X)  # In-phase component
meas.register_parameter(lockin.Y)  # Quadrature component

# Run with context manager (auto-saves to SQLite)
with meas.run(name='pe_loop_measurement') as datasaver:
    voltages = np.linspace(0, 5, 100)

    for v in voltages:
        smu.voltage(v)
        time.sleep(0.1)  # Wait for settling

        # Log all parameters together
        datasaver.add_result(
            (smu.voltage, smu.voltage()),
            (smu.current, smu.current()),
            (lockin.X, lockin.X()),
            (lockin.Y, lockin.Y())
        )

# Access saved data from SQLite
dataset = datasaver.dataset
df = dataset.to_pandas_dataframe()
print(df.head())
```

### Real-Time Plotting in Jupyter

```python
from qcodes.plots.qcodes_pyqtgraph import QtPlot

# Create plot context
plot = QtPlot(
    [smu.voltage.set_cmd],
    [smu.current],
    title="Live P-E Loop"
)

# Plot updates in real-time as data is logged
with meas.run(name='live_pe_loop') as datasaver:
    for v in np.linspace(0, 5, 100):
        smu.voltage(v)
        datasaver.add_result(
            (smu.voltage, smu.voltage()),
            (smu.current, smu.current())
        )
        plot.update()
```

### Why Use QCoDeS?
- **Reproducibility:** SQLite stores all parameters, settings, timestamps
- **Data organization:** Hierarchical experiment structure
- **Scalability:** Handles 100+ device measurements per minute
- **Integration:** Works seamlessly with Jupyter for interactive analysis

---

## 4. python-ivi (IVI Standard Compliance)

**GitHub:** https://github.com/python-ivi/python-ivi
**License:** MIT
**Version:** 0.52+
**Driver Count:** 200+ (all IVI-compliant)

### What It Does
Implements IVI (Interchangeable Virtual Instruments) standard, allowing instrument substitution without code changes.

### Instrument Categories

```
Scopes:
  - Agilent/Keysight DSOX, 54855, Infiniium
  - Tektronix TDS, MSO, DPO series
  - LeCroy Waverunner, HDO

AWG/Function Generator:
  - Agilent/Keysight 81150, 81160, 81160A
  - Tektronix AFG3000 series

DMM:
  - Agilent/Keysight 34410, 34411
  - Keithley 2000, 2010, 2015

Power Supply:
  - Agilent/Keysight E36xx series
  - Sorensen XG, Kepco
```

### Installation

```bash
git clone https://github.com/python-ivi/python-ivi
cd python-ivi
pip install .
```

### Example: Oscilloscope Waveform Capture

```python
import ivi

# Open scope (automatic driver selection by model number)
scope = ivi.agilent.agilentMSO7104A("TCPIP::192.168.1.104::INSTR")

# Configure acquisition (IVI-standard parameters)
scope.acquisition.type = 'normal'
scope.acquisition.number_of_points_minimum = 1000
scope.timebase.range = 0.001  # 1 ms/div

# Set up trigger
scope.trigger.type = 'edge'
scope.trigger.source = scope.channels[0]
scope.trigger.level = 1.0  # 1V trigger threshold

# Enable channels
scope.channels[0].enabled = True
scope.channels[1].enabled = True

# Capture waveform
scope.acquisition.initiate()
scope.acquisition.wait_for_acquisition_complete(timeout=10)

# Fetch data
waveform_ch1 = scope.channels[0].measurement.fetch_waveform()
waveform_ch2 = scope.channels[1].measurement.fetch_waveform()

# Access waveform arrays
print(f"Time axis: {waveform_ch1.time}")
print(f"Ch1 voltage: {waveform_ch1.voltage}")
print(f"Ch2 voltage: {waveform_ch2.voltage}")

# Plot
import matplotlib.pyplot as plt
plt.plot(waveform_ch1.time * 1e3, waveform_ch1.voltage, 'b-', label='Ch1')
plt.plot(waveform_ch2.time * 1e3, waveform_ch2.voltage, 'r-', label='Ch2')
plt.xlabel('Time (ms)')
plt.ylabel('Voltage (V)')
plt.legend()
plt.grid(True, alpha=0.3)
plt.show()
```

### Key Advantage
Swap from Agilent to Tektronix scope with single line change:
```python
scope = ivi.agilent.agilentMSO7104A(...)  # Change this line
# All IVI properties remain identical
```

---

## 5. nidaqmx (National Instruments DAQ)

**GitHub:** https://github.com/ni/nidaqmx-python
**License:** MIT
**Version:** 1.0+
**Hardware:** NI USB/PCIe DAQ cards

### What It Does
High-speed, multi-channel analog and digital I/O. Essential for synchronized measurements on multiple FeCIM cells.

### Installation

```bash
# Requires NI-DAQmx runtime (installed separately)
pip install nidaqmx

# Download runtime:
# https://www.ni.com/en/support/downloads/drivers/download.nidaqmx.html
```

### Example: Multi-Channel Voltage Measurement

```python
import nidaqmx
from nidaqmx.constants import AcquisitionType
import numpy as np

# Create task for analog input
with nidaqmx.Task() as task:
    # Add 4 voltage input channels
    task.ai_channels.add_ai_voltage_chan(
        "Dev1/ai0:3",
        name_to_assign_to_channel="Crossbar_Cells",
        min_val=-5,
        max_val=5
    )

    # Configure timing: 10 kHz sampling, continuous
    task.timing.cfg_samp_clk_timing(
        rate=10000,  # 10 kS/s
        sample_mode=AcquisitionType.CONTINUOUS
    )

    # Read 1000 samples per channel (100 ms at 10 kHz)
    data = task.read(number_of_samples_per_channel=1000)

    # data is list of arrays, one per channel
    ch0_data = np.array(data[0])
    ch1_data = np.array(data[1])
    ch2_data = np.array(data[2])
    ch3_data = np.array(data[3])

    print(f"Ch0 mean: {ch0_data.mean():.4f} V")
    print(f"Ch1 std: {ch1_data.std():.6f} V")
```

### Example: Synchronized Analog Output + Input

```python
import nidaqmx
import numpy as np

# Voltage pulse waveform for ferroelectric switching
pulse_voltage = np.concatenate([
    np.zeros(100),              # Wait 100 samples
    np.ones(200) * 5.0,        # +5V pulse, 200 samples
    np.zeros(100)              # Wait 100 samples
])

with nidaqmx.Task() as task_ao, nidaqmx.Task() as task_ai:

    # Analog Output: Write pulse to all rows
    task_ao.ao_channels.add_ao_voltage_chan("Dev1/ao0:7")
    task_ao.timing.cfg_samp_clk_timing(rate=10000)

    # Analog Input: Measure all columns
    task_ai.ai_channels.add_ai_voltage_chan("Dev1/ai0:7", max_val=5, min_val=-5)
    task_ai.timing.cfg_samp_clk_timing(rate=10000)

    # Write output (same waveform to all 8 rows)
    voltage_matrix = np.tile(pulse_voltage, (8, 1))  # 8 rows × 400 samples
    task_ao.write(voltage_matrix, auto_start=True)

    # Read current response (8 columns × 400 samples)
    current_data = task_ai.read(number_of_samples_per_channel=len(pulse_voltage))

    print(f"Measured current shape: {np.array(current_data).shape}")
```

### Why Use for FeCIM?
- **Sync:** Sample all 30 crossbar cells simultaneously
- **Speed:** 100+ kS/s per channel
- **Isolation:** Analog I/O isolated from PC noise
- **Trigger:** Hardware trigger for precise timing

---

## 6. Lock-In Amplifiers: SR830 & SR865

**PyMeasure Driver:** Available
**Direct VISA:** Supported
**QCoDeS Driver:** Microsoft Instruments

### Applications for FeCIM

| Measurement | Technique | Frequency |
|-------------|-----------|-----------|
| **Capacitance-Voltage** | 1 MHz AC bridge | 100 kHz - 1 MHz |
| **Conductance** | AC conductivity | 100 Hz - 100 kHz |
| **Leakage Current** | Phase-sensitive | 1-10 kHz |
| **Hysteresis Losses** | Ellipticity | DC + 1 kHz |

### PyMeasure Example: AC Measurement

```python
from pymeasure.instruments.stanford_research import SR830
import numpy as np
import matplotlib.pyplot as plt

lockin = SR830("GPIB0::8::INSTR")

# Configure lock-in amplifier
lockin.frequency = 10000  # 10 kHz drive
lockin.amplitude = 0.1    # 100 mV AC stimulus
lockin.input = 'A'        # Input channel A
lockin.sensitivity = 0.001  # 1 mV sensitivity
lockin.time_constant = 1.0  # 1 second averaging

# Measure X (in-phase) and Y (quadrature)
x = lockin.x
y = lockin.y
r = lockin.r      # Magnitude
theta = lockin.theta  # Phase

print(f"X (In-phase): {x:.6f} V")
print(f"Y (Quadrature): {y:.6f} V")
print(f"Magnitude: {r:.6f} V")
print(f"Phase: {theta:.2f}°")

# Complex impedance calculation
Z = (x + 1j * y) / 0.1  # V_measured / V_applied
R = Z.real  # Resistance
X_L = Z.imag  # Reactance

print(f"R = {R:.2e} Ω, X = {X_L:.2e} Ω")

lockin.shutdown()
```

---

## Communication Interfaces

### GPIB (IEEE 488.2)
Traditional lab standard. Requires GPIB controller.

| Adapter | Cost | Speed |
|---------|------|-------|
| NI GPIB-USB-HS | $500-800 | 1 MB/s |
| Prologix GPIB-USB | $150-200 | 200 kB/s |
| NI PCI GPIB | $400-600 | 8 MB/s |

### USB-TMC (Test & Measurement Class)
Modern instruments, direct USB connection.

```python
import pyvisa
rm = pyvisa.ResourceManager()
print(rm.list_resources())  # USBTMC0::0x05E6::0x2450::...
```

### Ethernet (LXI / VXI-11)
Networked labs, remote access.

```python
# Device on lab network
scope = ivi.agilent.agilentMSO7104A("TCPIP::192.168.1.104::INSTR")
```

---

## Instrument Selection Matrix for FeCIM

### Tier 1: Essential (Minimum Setup)

| Device | Purpose | Recommended | Driver |
|--------|---------|------------|--------|
| **SMU** | Apply V, measure I | Keithley 2400 | PyMeasure |
| **Oscilloscope** | Transient capture | Agilent DSOX | python-ivi |
| **Function Generator** | Pulse waveforms | Agilent 81150 | python-ivi |

### Tier 2: Advanced (Research Lab)

| Device | Purpose | Recommended | Driver |
|--------|---------|------------|--------|
| **Lock-In** | AC measurements | SR830 | PyMeasure/QCoDeS |
| **DAQ Card** | Multi-channel sync | NI USB-6341 | nidaqmx |
| **Parameter Analyzer** | Fast C-V | Agilent 4294A | Custom |

### Tier 3: Production

| Device | Purpose | Recommended | Driver |
|--------|---------|------------|--------|
| **Multi-site Handler** | Parallel testing | LTX-Credence | Custom |
| **Pattern Generator** | Complex sequences | Marvin Test | Custom |
| **Thermal Chamber** | Temperature sweep | Oven + PID | Custom |

---

## Recommended Installation Paths

### Path 1: Getting Started (Beginner)

```bash
# Minimal dependencies
pip install pyvisa pyvisa-py

# For Keithley measurements
pip install pymeasure

# Total setup: 5 minutes
```

### Path 2: Research Lab (Recommended)

```bash
# Full data acquisition stack
pip install pyvisa pyvisa-py
pip install pymeasure[UI]
pip install qcodes[jupyter,matplotlib]

# Optional: NI DAQ support
pip install nidaqmx
# (Requires NI-DAQmx runtime from ni.com)

# Total setup: 15 minutes
```

### Path 3: Advanced (Multi-Protocol)

```bash
# All tools
pip install pyvisa pyvisa-py
pip install pymeasure[UI]
pip install qcodes[jupyter,matplotlib,slack]
pip install python-ivi
pip install nidaqmx

# Optional development
git clone https://github.com/pymeasure/pymeasure
cd pymeasure && pip install -e .

# Total setup: 30 minutes
```

---

## Complete Workflow Example: FeCIM Characterization

```python
"""
Complete ferroelectric device characterization workflow.
Measures P-E loops, capacitance, and leakage current.
"""

import numpy as np
import matplotlib.pyplot as plt
from pymeasure.instruments.keithley import Keithley2400
from pymeasure.instruments.stanford_research import SR830
from datetime import datetime

class FeCIMCharacterizer:
    def __init__(self, smu_addr='GPIB0::24::INSTR',
                 lockin_addr='GPIB0::8::INSTR'):
        self.smu = Keithley2400(smu_addr)
        self.lockin = SR830(lockin_addr)
        self.setup_instruments()

    def setup_instruments(self):
        """Configure both instruments."""
        # SMU configuration
        self.smu.apply_voltage()
        self.smu.source_voltage_range = 10
        self.smu.compliance_current = 0.1
        self.smu.enable_source()

        # Lock-in configuration
        self.lockin.frequency = 100000  # 100 kHz
        self.lockin.amplitude = 0.05    # 50 mV AC
        self.lockin.sensitivity = 0.001
        self.lockin.time_constant = 0.5

    def measure_pe_loop(self, v_max=5.0, num_points=100):
        """Measure P-E hysteresis loop."""
        voltages = np.concatenate([
            np.linspace(0, v_max, num_points),
            np.linspace(v_max, -v_max, 2*num_points),
            np.linspace(-v_max, 0, num_points)
        ])

        currents = []

        for v in voltages:
            self.smu.source_voltage = v
            i = self.smu.current
            currents.append(i)

        return voltages, np.array(currents)

    def measure_capacitance(self, voltages):
        """Measure C-V at fixed AC frequency."""
        capacitances = []

        for v in voltages:
            self.smu.source_voltage = v
            # Lock-in measures impedance at 100 kHz
            x = self.lockin.x
            y = self.lockin.y

            # Z = V / I at lock-in frequency
            z = complex(x, y) / self.lockin.amplitude

            # C = Im(Y) / ω where Y = 1/Z
            omega = 2 * np.pi * self.lockin.frequency
            y_imag = z.imag / (abs(z)**2 * omega)
            c_meas = y_imag

            capacitances.append(c_meas)

        return np.array(capacitances)

    def measure_leakage(self, voltage, duration_seconds=10, sampling_rate=100):
        """Measure DC leakage at fixed voltage."""
        self.smu.source_voltage = voltage

        samples = int(duration_seconds * sampling_rate)
        currents = []
        times = []

        for i in range(samples):
            current = self.smu.current
            currents.append(current)
            times.append(i / sampling_rate)

        return np.array(times), np.array(currents)

    def plot_results(self, v, i, v_cv, c, times, i_leak):
        """Plot all measurement results."""
        fig, axes = plt.subplots(2, 2, figsize=(12, 10))

        # P-E Loop
        axes[0, 0].plot(v, i*1e6, 'b-', linewidth=2)
        axes[0, 0].set_xlabel('Voltage (V)')
        axes[0, 0].set_ylabel('Current (μA)')
        axes[0, 0].set_title('P-E Hysteresis Loop')
        axes[0, 0].grid(True, alpha=0.3)

        # C-V Curve
        axes[0, 1].plot(v_cv, c*1e12, 'r-', linewidth=2)
        axes[0, 1].set_xlabel('Voltage (V)')
        axes[0, 1].set_ylabel('Capacitance (pF)')
        axes[0, 1].set_title('Capacitance-Voltage (100 kHz)')
        axes[0, 1].grid(True, alpha=0.3)

        # Leakage over time
        axes[1, 0].semilogy(times, np.abs(i_leak)*1e9, 'g-', linewidth=2)
        axes[1, 0].set_xlabel('Time (s)')
        axes[1, 0].set_ylabel('Current (nA, log scale)')
        axes[1, 0].set_title('DC Leakage Current (10s)')
        axes[1, 0].grid(True, alpha=0.3, which='both')

        # Statistics
        ax = axes[1, 1]
        ax.axis('off')
        stats = f"""
        Device Metrics:
        ──────────────────
        Max Current: {i.max()*1e6:.2f} μA
        Min Current: {i.min()*1e6:.2f} μA
        Leakage (avg): {i_leak.mean()*1e9:.3f} nA
        Capacitance range: {c.min()*1e12:.2f} - {c.max()*1e12:.2f} pF

        Measurement Info:
        ──────────────────
        Timestamp: {datetime.now().isoformat()}
        Frequency: 100 kHz
        """
        ax.text(0.1, 0.5, stats, fontfamily='monospace', fontsize=10,
                verticalalignment='center')

        plt.tight_layout()
        plt.savefig('fecim_characterization.png', dpi=300, bbox_inches='tight')
        plt.show()

    def shutdown(self):
        """Clean shutdown."""
        self.smu.disable_source()
        self.smu.shutdown()
        self.lockin.shutdown()

# Run complete characterization
if __name__ == '__main__':
    char = FeCIMCharacterizer()

    try:
        # Measure P-E loop
        print("Measuring P-E loop...")
        v, i = char.measure_pe_loop(v_max=5.0, num_points=100)

        # Measure C-V
        print("Measuring C-V...")
        v_cv = np.linspace(-5, 5, 50)
        c = char.measure_capacitance(v_cv)

        # Measure leakage
        print("Measuring leakage current...")
        times, i_leak = char.measure_leakage(voltage=0.0, duration_seconds=10)

        # Plot everything
        char.plot_results(v, i, v_cv, c, times, i_leak)

    finally:
        char.shutdown()
```

---

## Troubleshooting

### Connection Issues

```python
import pyvisa

# List all instruments
rm = pyvisa.ResourceManager()
resources = rm.list_resources()
print(resources)

# Verbose diagnostics
rm.visalib.log_to_screen()

# Test connection
try:
    instr = rm.open_resource('GPIB0::24::INSTR')
    print(instr.query('*IDN?'))
    instr.close()
except pyvisa.errors.VisaIOError as e:
    print(f"Connection error: {e}")
```

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `ConnectionError: No VISA` | Backend not installed | `pip install pyvisa-py` |
| `VisaIOError: timeout` | Instrument not responding | Check GPIB/USB cable, reboot instrument |
| `ValueError: Unknown command` | Wrong SCPI syntax | Check instrument manual for exact command |
| `Permission denied` | USB permissions (Linux) | Add udev rule or run with `sudo` |

---

## Performance Benchmarks

### Data Acquisition Speed

| Tool | Single Channel | 8 Channels | Overhead |
|------|---|---|---|
| **PyVISA** | 50 Hz | N/A | ~20 ms per query |
| **PyMeasure** | 100 Hz | N/A | ~10 ms (driver wrapper) |
| **QCoDeS** | 100 Hz | N/A | ~10 ms + SQLite |
| **nidaqmx** | 100+ kHz | 100+ kHz | <1 ms (hardware) |

For FeCIM: Use nidaqmx for fast transients (<100 μs), QCoDeS for slow measurements (<10 Hz).

---

## Resources

### Documentation
- **PyVISA:** https://pyvisa.readthedocs.io/
- **PyMeasure:** https://pymeasure.readthedocs.io/
- **QCoDeS:** https://microsoft.github.io/Qcodes/
- **python-ivi:** https://python-ivi.readthedocs.io/
- **nidaqmx:** https://nidaqmx-python.readthedocs.io/

### SCPI References
- **SCPI Standard:** https://www.ivifoundation.org/
- **Keithley SMU Commands:** https://www.tek.com/en/products/keithley/source-measure-units
- **Agilent/Keysight Scopes:** https://www.keysight.com/en/pd-x201804-pna/network-analyzer

### Tutorials
- PyVISA Getting Started: https://pyvisa.readthedocs.io/en/latest/getting_started/index.html
- PyMeasure Examples: https://github.com/pymeasure/pymeasure/tree/master/examples

---

## Summary

| Tool | Best For | Difficulty | Learning Curve |
|------|----------|-----------|-----------------|
| **PyVISA** | Direct instrument control | Low | 30 min |
| **PyMeasure** | Keithley, practical measurements | Medium | 1-2 hours |
| **QCoDeS** | Research labs, data management | High | 2-4 hours |
| **python-ivi** | Multiple vendor instruments | Medium | 1-2 hours |
| **nidaqmx** | High-speed, multi-channel | High | 2-4 hours |

**Recommendation for FeCIM:** Start with PyMeasure + Keithley SMU. Upgrade to QCoDeS + nidaqmx for advanced multi-device characterization.
