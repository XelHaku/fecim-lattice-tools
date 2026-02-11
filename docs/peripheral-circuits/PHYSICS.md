# Peripheral Circuits Physics: Deep Technical Reference

**FeCIM Lattice Tools - Module 4: Peripheral Circuits**

> Start here for comprehensive physics understanding of ADCs, DACs, TIAs, and charge pumps that support ferroelectric compute-in-memory operations.

---

## Overview

This document provides the physics foundation for peripheral circuits in FeCIM systems. It covers:

- **Analog-to-Digital Converters (ADCs)**: 5-bit SAR architecture, quantization, linearity errors
- **Digital-to-Analog Converters (DACs)**: 5-bit output, voltage ranges, nonlinearity modeling
- **Transimpedance Amplifiers (TIAs)**: Current-to-voltage conversion, noise, bandwidth
- **Charge Pumps**: Voltage boosting for write operations, efficiency modeling
- **Timing Analysis**: Read/write cycle timing, throughput
- **System Integration**: How peripheral circuits work together in a FeCIM crossbar

**Note:** Unless explicitly cited, numeric values in this document are simulation defaults or illustrative assumptions from `shared/peripherals` (not measured hardware data) ([CITATION NEEDED - placeholder value]). Equations are standard references.

## Model Defaults Map (Doc ↔ Code)

Quick map of numeric values used in this document to code defaults (simulation inputs). Values below are **not** measured unless explicitly cited.

| Doc value (where used) | Code default | Source |
|---|---|---|
| ADC bits/levels (5‑bit, 32 levels; 30 used) | `DefaultBits=5`, `DefaultLevels=32`, `DefaultADC().Bits=5` | `shared/peripherals/defaults.go`, `shared/peripherals/adc.go` |
| ADC Vref high/low (1.0V / 0.0V) | `ADCVrefHigh`, `ADCVrefLow`, `DefaultADC().Vref*` | `shared/peripherals/defaults.go`, `shared/peripherals/adc.go` |
| ADC conversion time (50 ns) | `ADCConversionTime`, `DefaultADC().ConversionTime` | `shared/peripherals/defaults.go`, `shared/peripherals/adc.go` |
| ADC INL/DNL (0.5 / 0.25 LSB) | `DefaultINL`, `DefaultDNL`, `DefaultADC().INL/DNL` | `shared/peripherals/defaults.go`, `shared/peripherals/adc.go` |
| ADC energy per conversion (~25 fJ) | `ADC.EnergyPerConversion()` (`5e-15 * Bits` for SAR) | `shared/peripherals/adc.go` |
| DAC Vref high/low (±1.5V) | `DACVrefHigh`, `DACVrefLow`, `DefaultDAC().Vref*` | `shared/peripherals/defaults.go`, `shared/peripherals/dac.go` |
| DAC settle time (10 ns) | `DACSettleTime`, `DefaultDAC().SettleTime` | `shared/peripherals/defaults.go`, `shared/peripherals/dac.go` |
| DAC INL/DNL (0.5 / 0.25 LSB) | `DefaultINL`, `DefaultDNL`, `DefaultDAC().INL/DNL` | `shared/peripherals/defaults.go`, `shared/peripherals/dac.go` |
| DAC energy per conversion (~15 fJ) | `DAC.EnergyPerConversion()` (uses `capacitance=0.2e-15`) | `shared/peripherals/dac.go` |
| TIA gain/bandwidth/noise/offset/max (10 kΩ, 100 MHz, 1 pA/√Hz, 5 mV, 100 µA, 1.0 V) | `DefaultTIA()` fields | `shared/peripherals/tia.go` |
| TIA settling time (~11 ns) | `TIA.SettlingTime()` derived from bandwidth | `shared/peripherals/tia.go` |
| Charge pump defaults (Vin 1.0V, Vout 1.5V, 2 stages, 0.3V drop, 50 MHz, 10 µA, 100 pF, 70%) | `DefaultChargePump()` fields | `shared/peripherals/chargepump.go` |
| Charge pump rise time (~88 ns) | `ChargePump.RiseTime()` derived from stages/clock | `shared/peripherals/chargepump.go` |
| Array settle (sneak/RC) 5 ns | `arraySettle := 5e-9` (constant; not parameterized) | `shared/peripherals/analysis.go` |
| Program pulse 100 ns | `writePulse := 100e-9` (constant; not parameterized) | `shared/peripherals/analysis.go` |

---

## Part 1: Analog-to-Digital Converters (ADCs)

### What Does an ADC Do?

In a crossbar read operation, sensing cells produce **tiny currents** (microamps to nanoamps). These must be converted to digital levels (0-31) for the neural network.

```
Crossbar Read:
    Cell current (I)  →  [TIA]  →  Voltage (V)  →  [ADC]  →  Digital Code (0-31)
      100 nA                        1 mV                     → to computation
```

The ADC performs **quantization**: mapping continuous voltages to discrete digital levels.

### 5-Bit SAR Architecture

The Ferroelectric CIM system uses a **5-bit Successive Approximation Register (SAR)** ADC:

- **5 bits** = 2^5 = **32 possible codes** (0 to 31)
- **Demo baseline uses 30 levels** (simulation baseline; codes 0-29, reserving 30-31)
- **Architecture:** SAR compares input voltage against a DAC reference iteratively

#### SAR Conversion Process

```
SAR Algorithm (searching for the right digital code):

Input: Vanalog
Setup: Vref_high = 1.0V, Vref_low = 0.0V

Bit 4 (16s place):  Set MSB, compare Vanalog vs 0.5V
                    If Vanalog > 0.5V, keep bit. Otherwise, clear.

Bit 3 (8s place):   Compare against next level (either 0.75V or 0.25V)
                    If match, keep bit.

Bit 2 (4s place):   Continue binary search...
Bit 1 (2s place):
Bit 0 (1s place):   Final bit set

Result: 5-bit digital code (0-31)

Time: 50 ns (model default; `ADCConversionTime` / `DefaultADC().ConversionTime`; per‑bit timing is illustrative)
```

### Ideal Conversion Formula

For a given analog voltage, the ideal digital code is (mid‑tread quantization, then clamped to 0…31):

$$\text{Code} = \text{round}\left(\frac{V_{\text{analog}} - V_{\text{ref,low}}}{LSB}\right)$$

Where the **Least Significant Bit (LSB)** is:

$$LSB = \frac{V_{\text{ref,high}} - V_{\text{ref,low}}}{2^{\text{bits}} - 1} = \frac{1.0V - 0.0V}{31} \approx 32.3 \text{ mV}$$

#### Example Conversions

| Input Voltage | Ideal Code | Physical Meaning |
|---|---|---|
| 0.0 V | 0 | Minimum (no current) |
| 0.325 V | 10 | ~1/3 scale |
| 0.5 V | 16 | Mid-range (rounded) |
| 1.0 V | 31 | Maximum (full scale) |

### Real ADCs Have Imperfections

Ideal ADCs don't exist. Real hardware introduces two main errors:

#### 1. Integral Non-Linearity (INL)

**Definition:** Maximum deviation of the actual transfer function from a straight line.

```
Ideal vs Real ADC Response:

Code │
  31 │                                     ●
     │                                  ●
  20 │      Ideal ━━━━━━━                ●
     │      Real  ⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯          ●
     │                    ╲        ╱
  10 │                     ╲    ╱
     │                       ╲╱
   0 │●─────────────────────────────────────
     └─────────────────────────────────────→ Voltage
       0V                                1V

   INL = maximum vertical offset between real and ideal = 0.5 LSB
```

**Impact:** Certain voltage ranges compress or expand, reducing effective resolution.

**Typical value:** 0.5 LSB (default in FeCIM)

#### 2. Differential Non-Linearity (DNL)

**Definition:** Variation in step size between adjacent codes.

```
Step Size Variation:

Ideal steps:    32.3 mV ║ 32.3 mV ║ 32.3 mV ║ 32.3 mV
                         ║         ║         ║

Real steps:     30 mV ║ 35 mV ║ 31 mV ║ 34 mV
                      ║       ║       ║

DNL = (actual_step - ideal_step) / LSB
DNL at code 2 = (35 - 32.3) / 32.3 = 0.08 LSB
```

**Impact:** Some codes cover wider voltage ranges than others. Worst case: **missing codes** (DNL < -1 LSB).

**Typical value:** 0.25 LSB (default in FeCIM)

### Modeling INL and DNL in Code

The FeCIM ADC adds sinusoidal INL and pseudo-random DNL:

```go
// Input voltage with INL/DNL errors applied
func (a *ADC) ConvertWithNonlinearity(voltage float64) int {
    lsb := (a.VrefHigh - a.VrefLow) / float64(a.Levels()-1)

    // Get ideal code first
    idealLevel := a.Convert(voltage)

    // INL error: varies sinusoidally with code (periodic device mismatch)
    inlOffset := a.INL * lsb * sin(π * level / 31)

    // DNL error: pseudo-random per level (random device variations)
    dnlOffset := a.DNL * lsb * (0.5 - level%5 / 4)

    // Re-convert with errors applied
    return a.Convert(voltage + inlOffset + dnlOffset)
}
```

### ADC Performance Metrics

#### Effective Number of Bits (ENOB)

Accounting for INL and DNL, the actual resolution is less than 5 bits:

$$ENOB = \text{Bits} - \log_2\sqrt{1 + INL^2 + DNL^2}$$

For default FeCIM ADC (INL=0.5 LSB, DNL=0.25 LSB):

$$ENOB = 5 - \log_2\sqrt{1 + 0.5^2 + 0.25^2} = 5 - 0.20 = 4.80 \text{ bits}$$

This means the ADC actually provides **~4.80 bits of resolution** instead of 5.

#### Signal-to-Noise Ratio (SNR)

**Theoretical SNR** (perfect ADC, quantization noise only):

$$SNR_{\text{theo}} = 6.02 \times N + 1.76 \text{ dB}$$

For 5-bit ADC:
$$SNR_{\text{theo}} = 6.02 \times 5 + 1.76 = 31.86 \text{ dB}$$

**Effective SNR** (including nonlinearity):

$$SNR_{\text{eff}} = 6.02 \times ENOB + 1.76 = 6.02 \times 4.80 + 1.76 = 30.7 \text{ dB}$$

The nonlinearity reduces SNR by **1.16 dB** (about 23% noise increase).

### ADC Conversion Time

The SAR ADC requires time for:
1. **Comparator delay:** ~5 ns per bit
2. **DAC settling:** ~10 ns
3. **Latch settling:** ~5 ns

**Total: ~50 ns** per conversion

This is critical for **throughput**:
- Read path: 50 ns ADC conversion
- Write path: 10 ns DAC settling + 88 ns charge pump rise time

---

## Part 2: Digital-to-Analog Converters (DACs)

### What Does a DAC Do?

In a crossbar write operation, neural network levels (0-29) must be converted to **write voltages** to program ferroelectric cells.

```
Crossbar Write:
    Digital Level (0-29)  →  [DAC]  →  Voltage (±1.5V)  →  [Charge Pump]  →  to cell
```

The DAC performs **interpolation**: mapping discrete levels to analog voltages that will set the cell conductance.

### 5-Bit Architecture and Voltage Range

The FeCIM DAC uses 5-bit resolution with a **bipolar output range**:

- **Input:** 5-bit digital code (0-31, using 0-29 for FeCIM; codes 30-31 reserved as headroom)
- **Output range:** -1.5V to +1.5V (3V span)
- **Total levels:** 32, with 30 used

#### DAC Voltage Mapping

$$V_{\text{out}} = V_{\text{ref,low}} + \frac{\text{code}}{2^N - 1} \times (V_{\text{ref,high}} - V_{\text{ref,low}})$$

For FeCIM:
$$V_{\text{out}} = -1.5V + \frac{\text{code}}{31} \times 3.0V$$

#### Example DAC Conversions

| Digital Code | Output Voltage | Cell Effect |
|---|---|---|
| 0 | -1.50 V | Minimum conductance (erase) |
| 15 | -0.05 V | Near mid-level (slightly negative) |
| 16 | +0.05 V | Near mid-level (slightly positive) |
| 29 | +1.31 V | Max FeCIM level (code 29) |
| 31 | +1.50 V | Full-scale (reserved headroom) |

### DAC Nonlinearity: INL and DNL

DACs have the same error sources as ADCs, with similar impacts:

#### INL in DACs

```
Ideal vs Real DAC Output:

Vout │
  1.5V │                                     ●
       │                                  ●
  0.9V │      Ideal ━━━━━━━                ●
       │      Real  ⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯          ●
       │                    ╲        ╱
  0.3V │                     ╲    ╱
       │                       ╲╱
 -1.5V │●─────────────────────────────────────
       └─────────────────────────────────────→ Code
         0                                 31

   INL = 0.5 LSB = 0.5 × 96.8mV = 48.4mV
   (Max voltage output deviation from ideal line)
```

#### DNL in DACs

```
Output step variation:

Ideal steps:     96.8mV ║ 96.8mV ║ 96.8mV ║ 96.8mV
Real steps:      94mV ║ 99mV ║ 96mV ║ 98mV
                      ║       ║       ║

DNL = (actual_step - ideal_step) / LSB
```

**FeCIM DAC defaults:** INL = 0.5 LSB, DNL = 0.25 LSB

### DAC Energy Consumption

Energy per conversion (writing one level):

$$E_{\text{DAC}} = C_{\text{eff}} \times \left(\frac{V_{\text{span}}}{2}\right)^2 \times 2^N$$

For FeCIM (effective switched‑cap model; `DAC.EnergyPerConversion()` uses `C_eff = 0.2 fF` as a simulation input):
- Effective unit capacitor: 0.2 fF
- Vspan: 3.0 V (±1.5 V rails)
- Number of levels: 32

$$E_{\text{DAC}} = 0.2 \times 10^{-15}F \times \left(\frac{3.0}{2}\right)^2 \times 32 \approx 14.4 \text{ fJ}$$

**Typical: ~15 fJ per conversion** (model default; architecture-dependent)

---

## Part 3: Transimpedance Amplifiers (TIAs)

### What Does a TIA Do?

Crossbar sensing produces **currents** (I). Peripheral electronics need **voltages** (V) for the ADC. The TIA converts:

$$V_{\text{out}} = I_{\text{in}} \times R_f$$

where Rf is the **feedback resistance (transimpedance gain)**.

```
TIA Block Diagram:

  ┌─────────────────────────┐
  │                         │
  │  ┌──────┐               │
  │  │  +   │               │
I─→─┤  OA   ├──────┬────────┤V_out
  │  │  -   │      │        │
  │  └──────┘     Rf        │
  │               │         │
  │               ●         │
  │               │         │
  └─────────────────────────┘

Feedback: V_out = I × Rf  (virtual ground at OA input)
```

### TIA Specifications for FeCIM

Default configuration from `DefaultTIA()`:

| Parameter | Value | Meaning |
|---|---|---|
| **Transimpedance Gain** | 10 kΩ | 100 nA input → 1 mV output |
| **Bandwidth** | 100 MHz | Frequency response flat to 100 MHz |
| **Input Noise** | 1 pA/√Hz | Thermal noise density |
| **Output Offset** | 5 mV | DC offset voltage |
| **Max Input Current** | 100 µA | ADC input range limit |
| **Max Output Voltage** | 1.0 V | ADC reference voltage |

### TIA Noise Analysis

#### Thermal Noise

The TIA input transistor (and feedback resistor) generate **Johnson noise**:

$$i_{\text{noise,density}} = \sqrt{\frac{4kT}{R_f}}$$

where:
- k = 1.38 × 10^-23 J/K (Boltzmann constant)
- T = 300 K (room temperature)
- Rf = feedback resistance (Ω)

For FeCIM TIA (Rf = 10 kΩ):
$$i_{\text{noise,density}} \approx 1.3 \text{ pA/√Hz}$$

We model this as **1 pA/√Hz** (same order of magnitude). Over 100 MHz:
$$I_{\text{noise,rms}} \approx 1 \text{ pA/√Hz} \times \sqrt{100MHz} \approx 10 \text{ nA}$$

**Output noise voltage:**
$$V_{\text{noise,density}} = I_{\text{noise}} \times R_f = 1 \text{ pA/√Hz} \times 10k\Omega = 10 \text{ nV/√Hz}$$
$$V_{\text{noise,rms}} = V_{\text{noise,density}} \times \sqrt{BW} = 10 \text{ nV/√Hz} \times \sqrt{100MHz} \approx 100 \text{ µV}$$

#### Minimum Detectable Current

The smallest current we can reliably measure (SNR = 1):

$$I_{\text{min}} = I_{\text{noise}} \times \sqrt{BW} = 1 \text{ pA/√Hz} \times \sqrt{100 \times 10^6} = 10 \text{ nA}$$

Any input below 10 nA is buried in noise.

### SNR and Dynamic Range

#### Signal-to-Noise Ratio

For a given input current:

$$SNR = 20 \log_{10}\left(\frac{I_{\text{signal}} \times R_f}{V_{\text{noise}}}\right)$$

Example: 100 nA input
$$SNR = 20 \log_{10}\left(\frac{100 \times 10^{-9} \times 10^4}{1 \times 10^{-12} \times \sqrt{100 \times 10^6}}\right) = 20 \log_{10}(10) = 20 \text{ dB}$$

#### Dynamic Range

Ratio of maximum to minimum detectable signals:

$$DR = 20 \log_{10}\left(\frac{I_{\text{max}}}{I_{\text{min}}}\right) = 20 \log_{10}\left(\frac{100 \times 10^{-6}}{10 \times 10^{-9}}\right) = 80 \text{ dB}$$

This means the TIA can simultaneously handle signals spanning **100 million to 1** range—excellent for variable crossbar conductances.

### TIA Bandwidth and Settling

#### Frequency Response

Real TIAs have limited bandwidth due to the op-amp pole frequency. The transfer function is:

$$H(f) = \frac{R_f}{1 + j(f/f_p)}$$

where fp is the pole frequency (typically equal to the specified bandwidth).

```
Magnitude Response:

|H(f)| │
       │●●●●●●●●●●●●●
 (dB) │
       │            ●●●
       │               ●●●
-3dB  ┼─────────────────●●●
       │                   ●●●
       └──────────────────────→ Frequency
                             f_p = 100MHz
```

At the **-3dB point** (100 MHz), the gain is reduced by a factor of √2.

#### Settling Time

When the input current suddenly changes, the output voltage settles exponentially:

$$V_{\text{out}}(t) = V_{\text{final}} \times (1 - e^{-t/\tau})$$

For 0.1% settling accuracy:
$$\tau = \frac{\ln(1000)}{2\pi \times BW} = \frac{6.91}{2\pi \times 100 \times 10^6} = 11 \text{ ns}$$

**FeCIM TIA settles in ~11 ns** to 0.1% accuracy. This is critical for fast read operations.

### TIA Power Consumption

The simulation uses a **lower‑bound dynamic power** estimate tied to noise and bandwidth:

$$P \approx 2 \times \frac{kT \times BW \times R_f}{\eta}$$

This uses the feedback gain $R_f$ as a proxy for the front‑end transconductance to scale with bandwidth
(a deliberate simplification for repeatable, low‑overhead modeling). For the default TIA:

- $P \approx 8.3 \times 10^{-8} \text{ W}$ (**~83 nW**)

Energy per read window (using the default read time from `AnalyzeTiming`, ~76 ns):
$$E \approx 83 \text{ nW} \times 76 \text{ ns} \approx 6.3 \text{ fJ}$$

**Note:** Real high‑speed TIAs often draw **mW‑level bias power**; that static bias is **not** modeled here and can be layered on if needed.

---

## Part 4: Charge Pumps

### What Does a Charge Pump Do?

Ferroelectric cells require **high write voltages** (±1.5V) but the chip runs on **1V supply**. The charge pump boosts voltage:

```
Charge Pump (2-stage):

Input:  1V supply
Output: 1.5V write voltage
Boost:  1.5× (50% voltage increase)

Energy cost: 70% efficient (typical)
```

### Dickson Topology (2-Stage)

The FeCIM charge pump uses the **Dickson configuration**, the most common topology for on-chip voltage boosting.

#### How It Works

**Stage 1:** Charge flying capacitor C1 to Vin in clock phase 1

```
Clock Phase 1 (Charge up):

        ┌─────────C1─────────┐
        │                    │
   Vin─┤                     ├─→ Vout (floating)
        │                    │
        └───────────────────┘

   C1 charges to Vin volts
```

**Stage 2:** Connect C1 in series with Vin in clock phase 2, boosting output

```
Clock Phase 2 (Boost):

  Vin ─────┬───────C1───────┬─ Vout
           │                │
         ╱ │ ╱Diode         │ ╱Diode
        ╱  │╱               │╱
       ────┴────────────────┴──── GND

   C1 voltage (Vin) + Vin = 2×Vin connected in series
   But diode drops reduce output
   Actual: Vout ≈ (N+1)×Vin - N×Vth
           where N = number of stages
           Vth = diode drop per stage
```

#### Ideal Output Voltage

For N=2 stages:
$$V_{\text{out,ideal}} = (N+1) \times V_{in} = 3 \times 1.0V = 3.0V$$

But this assumes **perfect charge transfer**. Real losses reduce this.

#### Actual Output Voltage (With Losses)

The pump loses voltage in:
1. **Diode drops:** ~0.3V per MOS diode × stages
2. **IR drops:** Load current through effective output resistance
3. **Leakage:** Reverse current through pass devices

$$V_{\text{out,unreg}} = V_{\text{out,ideal}} - (N \times V_{th}) - I_{load} \times R_{out}$$
$$R_{out} \approx \frac{1}{C_{fly} \times f_{clk}}$$

For FeCIM defaults (model inputs from `DefaultChargePump()`: Vin=1V, N=2, Vth=0.3V, Cfly=100 pF, f=50 MHz, Iload=10 µA):
$$R_{out} \approx \frac{1}{100pF \times 50MHz} \approx 200 \Omega$$
$$V_{\text{out,unreg}} \approx 3.0V - 0.6V - 0.002V = 2.398V$$

The model then **regulates/clamps** to the target output rail:
$$V_{\text{out,actual}} = \text{sign}(V_{target}) \times \min(|V_{target}|, |V_{\text{out,unreg}}|)$$
For the default +1.5V rail, $V_{\text{out,actual}} \approx 1.5V$.

**Boost factor:**
$$\text{Boost} = \frac{V_{\text{out}}}{V_{in}} = \frac{1.5V}{1.0V} = 1.5\times$$

### Output Ripple and Regulation

#### Ripple Voltage

The pump cannot deliver voltage continuously. It oscillates due to discrete charging phases:

$$\Delta V = \frac{I_{\text{load}}}{C_{\text{out}} \times f}$$

where Cout is the output storage capacitor and f is clock frequency.

For FeCIM:
- Iload = 10 µA
- Cout = 1 nF (10× flying cap, with Cfly=100 pF)
- f = 50 MHz

$$\Delta V = \frac{10 \times 10^{-6}}{1 \times 10^{-9} \times 50 \times 10^6} = 0.2 \text{ mV}$$

**Ripple is small** with large output capacitance. Smaller C\_out values increase ripple proportionally.

Typical solution: Add external smoothing capacitor (0.1-1 nF).

#### Load Regulation

As load current increases, output voltage drops (pump can't supply enough current fast enough):

```
Output Voltage vs Load Current:

Vout │
1.5V ├●●●●●
     │     ●●●
1.4V │        ●●●
     │           ●●●
     │              ●
1.3V │               ●●
     │                 ●●
     └────────────────────→ Iload
       0 µA           100 µA

   Load regulation: -0.2V/100µA = -0.002V/µA
```

### Charge Pump Efficiency

**Efficiency** is the fraction of input power that becomes useful output power:

$$\eta = \frac{P_{\text{out}}}{P_{\text{in}}} = \frac{V_{\text{out}} \times I_{\text{load}}}{V_{\text{in}} \times I_{\text{in}}}$$

For ideal charge transfer (unregulated):
$$\eta_{\text{ideal}} = \frac{V_{\text{out,unreg}}}{(N+1) \times V_{in}}$$

With regulation to the 1.5V rail, we assume **70% power efficiency** as a practical switched‑cap value for this model.

$$P_{\text{in}} = \frac{P_{\text{out}}}{\eta}$$

For the regulated/clamped case (default target rail +1.5V with achievable unregulated headroom), the model uses:
$$P_{\text{out}} \approx V_{\text{out,actual}} \times I_{\text{load}} \approx 1.5V \times 10\mu A = 15 \text{ µW}$$
$$P_{\text{in}} \approx \frac{15\text{ µW}}{0.7} = 21.4 \text{ µW}$$

If the pump cannot reach the target rail (unregulated output below target), the code computes $P_{\text{out}}$ using the **actual achievable output voltage** returned by `ChargePump.ActualOutputVoltage()`.

Most power is dissipated in:
1. Diode drops: 60% of loss
2. IR drops: 30% of loss
3. Leakage: 10% of loss

### Rise Time

When the pump suddenly needs to supply voltage (before a write operation), the output rises from 0V to final value over time:

$$t_{\text{rise}} = \frac{N \times 2.2}{f_{clk}}$$

For 2-stage pump at 50 MHz:
$$t_{\text{rise}} = \frac{2 \times 2.2}{50 \times 10^6} = 88 \text{ ns}$$

**10%-90% rise time: ~88 ns**

This is the time needed before a write voltage is stable enough. Must be considered in write cycle timing.

### Maximum Current Capability

The pump has a physical limit on how much current it can source before voltage collapses:

$$I_{\text{max}} = C_{fly} \times f_{clk} \times (N+1) \times \frac{V_{in}}{V_{\text{out}}}$$

For FeCIM:
$$I_{\text{max}} = 100pF \times 50MHz \times 3 \times \frac{1.0}{1.5} = 10 \text{ mA}$$

Above this, the pump cannot maintain voltage—write operation fails.

### Negative Pump for Erase

For erasing cells (conductance reduction), we need **negative write voltage** (-1.5V). This uses an inverted Dickson pump:

```
Negative Pump Configuration:

Ground connected where Vin connects in positive pump
Vin connected where ground connects in positive pump
Result: -1.5V output (mirror image of positive pump)

Specifications identical to positive pump:
- Boost: 1.5×
- Efficiency: 70%
- Rise time: 88 ns
- Ripple: 0.2 mV
```

---

## Part 5: System Timing Analysis

Timing values below are derived from model defaults in `shared/peripherals` and are illustrative. Array settle (sneak/RC) 5 ns and program pulse 100 ns are fixed constants in `AnalyzeTiming` (placeholders, not parameterized) ([CITATION NEEDED - placeholder value]).

### FeCIM Operation Cycles

FeCIM performs three main operations, each with different timing:

#### 1. Read Cycle

Reading a cell's current and digitizing it:

```
Read Path Timeline:

DAC Settle:    ├────────────────┤  10 ns (set sense voltage)

Sneak/RC:      ├────────────────┤   5 ns (array settling)

TIA Settle:    ├────────────────────────┤  11 ns (current to voltage)

ADC Convert:   ├───────────────────────────────┤  50 ns (SAR conversion)

Total Read:    ├─────────────────────────────────────────┤  76 ns
```

**Read Time = 76 ns (13.1 MHz throughput if read-only)**

#### 2. Write Cycle

Programming a cell to a new conductance level:

```
Write Path Timeline:

DAC Settle:    ├────────────────┤  10 ns

Pump Rise:     ├──────────────────────────────────────┤  88 ns

Program Pulse: ├─────────────────┤  100 ns (apply Vc to cell)

Settling:      ├──────┤  5 ns (charge redistribution)

Total Write:   ├────────────────────────────────────────────────┤  203 ns
```

**Write Time = 203 ns (4.9 MHz throughput if write-only)**

#### 3. Compute Cycle (Inference)

Typical neural network inference: multiple reads followed by accumulation (conceptual, actual computation in crossbar):

```
One Compute Step (one row read):

    Read Phase A:
    DAC→TIA→ADC:     ├────┤  76 ns (read column 0)

    Read Phase B:
    DAC→TIA→ADC:     ├────┤  76 ns (read column 1)

    ...64 parallel ADCs (all in parallel!)

    Accumulate:       ├──┤  negligible (on-chip)

    Full Row Read:    ├────┤  76 ns (all columns in parallel!)
```

**Key:** ADCs are **fully parallel**. All 64 columns convert simultaneously.
- **One row read:** 76 ns
- **Full 64×64 inference:** 64 × 76 ns = 4.9 µs

### Throughput Implications

With realistic peripherals:

| Operation | Time | Throughput |
|---|---|---|
| Single read | 76 ns | 13.1 MHz |
| Single write | 203 ns | 4.9 MHz |
| Full row read (64 parallel) | 76 ns | 64× parallelism |
| Full inference (64×64 weights) | 4.9 µs | 204 inferences/ms |

---

## Part 6: System Energy Analysis

Energy numbers below are model-based estimates using `shared/peripherals` defaults, not measured device values ([CITATION NEEDED - placeholder value]).

### Energy Breakdown Per Operation

#### Per Read Operation

| Component | Energy | Fraction |
|---|---|---|
| DAC conversion | 14.4 fJ | 31% |
| TIA amplification | 6.3 fJ | 14% |
| ADC conversion | 25 fJ | 55% |
| **Total** | **~46 fJ** | 100% |

Energy normalized to 100 fJ baseline: **~0.46×**

#### Per Write Operation

| Component | Energy | Fraction |
|---|---|---|
| DAC conversion | 14.4 fJ | ~1% |
| Charge pump (input) | 2.14 pJ | ~99% |
| **Total** | **~2.15 pJ** | 100% |

Write is **~50× more energy‑expensive** than read (model default).  
**Note:** Cell‑internal write energy is not modeled here; add separately if needed.

#### Full Inference (64×64 MNIST)

64 reads (one for each weight):
$$E_{\text{inference}} = 64 \times 46fJ \approx 2.9pJ$$

This is the **computation energy only**. Does not include:
- Static power (leakage)
- Row/column decoding
- Off-chip memory I/O

---

## Part 7: Real-World Non-Idealities

### Non-Ideal Effects Not Fully Modeled

The physics documentation covers ideal/idealized circuits. Real devices show additional effects:

#### 1. Nonlinear Device Conductance

Cells don't have perfectly linear I-V curves:

```
Ideal vs Real Cell Response:

I │                    Real device
  │                    (exponential/
  │        ╱ Ideal     saturation)
  │    ╱╱╱
  │  ╱╱         ╱╱╱╱╱
  │╱╱         ╱╱╱
  └──────────────────→ V
```

This causes **10-20% error** in high/low conductance states. See **MODULE4-PHYSICS-IMPROVEMENTS.md** for enhancement proposals.

#### 2. Sneak Paths in Passive Arrays

In 0T1R passive crossbars, unwanted current flows through parallel paths:

```
Sneak Path Example (0T1R):

     BL_j (read)   BL_k   BL_l
         │          │      │
    ─────●──────────●──────●───→ target cell
         │      ╱   │      │
    ─────●─────●────●──────●───
         │         ╱ │      │
    ─────●────────●──●──────●───
             Sneak paths add 5-20% error
```

#### 3. IR Drop in Large Arrays

Resistance in row/column lines causes voltage loss:

$$V_{\text{actual}} = V_{\text{nominal}} - I \times R_{\text{line}} \times L$$

In 64×64 arrays: **1-5% voltage loss** at far corner.

#### 4. Write Disturb in Passive Mode

Unselected cells experience V/2 pulses during writes, causing gradual drift.

#### 5. Temperature Variation

Coercive field Ec and polarization Pr vary 10-15% across -40°C to +125°C operating range.

---

## Part 8: Design Equations Summary

### ADC

**Quantization (mid‑tread):**
$$\text{Code} = \text{round}\left(\frac{V - V_{ref,low}}{(V_{ref,high} - V_{ref,low})/(2^N-1)}\right)$$

**ENOB:**
$$ENOB = N - \log_2\sqrt{1 + INL^2 + DNL^2}$$

**SNR (effective):**
$$SNR_{eff} = 6.02 \times ENOB + 1.76 \text{ dB}$$

### DAC

**Output voltage:**
$$V_{out} = V_{ref,low} + \frac{\text{code}}{2^N - 1}(V_{ref,high} - V_{ref,low})$$

**Energy:**
$$E = C_{eff} \times \left(\frac{V_{span}}{2}\right)^2 \times 2^N$$

### TIA

**Output voltage:**
$$V_{out} = I_{in} \times R_f + V_{offset}$$

**Settling time (0.1% accuracy):**
$$t = \frac{\ln(1000)}{2\pi \times BW} \approx \frac{7}{2\pi \times BW}$$

**SNR:**
$$SNR = 20 \log_{10}\left(\frac{I \times R_f}{I_{noise} \times \sqrt{BW}}\right)$$

**Dynamic range:**
$$DR = 20 \log_{10}\left(\frac{I_{max}}{I_{min}}\right)$$

### Charge Pump

**Ideal output:**
$$V_{out,ideal} = (N+1) \times V_{in}$$

**Actual output (with losses):**
$$V_{out,unreg} = (N+1) \times V_{in} - N \times V_{th} - I_{load} \times R_{out}$$
$$R_{out} \approx \frac{1}{C_{fly} \times f_{clk}}$$
$$V_{out,actual} = \text{sign}(V_{target}) \times \min(|V_{target}|, |V_{out,unreg}|)$$

**Ripple:**
$$\Delta V = \frac{I_{load}}{C_{out} \times f}$$

**Boost factor:**
$$\text{Boost} = \frac{V_{out}}{V_{in}}$$

**Efficiency:**
$$\eta = \frac{V_{out}}{(N+1) \times V_{in}}$$

**Rise time (10%-90%):**
$$t_{rise} = \frac{N \times 2.2}{f_{clk}}$$

### Timing

**Read cycle (as implemented in `AnalyzeTiming`):**
$$t_{read} = t_{DAC} + t_{array\_settle} + t_{TIA} + t_{ADC}$$

**Write cycle (as implemented in `AnalyzeTiming`):**
$$t_{write} = t_{DAC} + t_{pump\_rise} + t_{pulse} + t_{array\_settle}$$

**Full inference (64×64 weight matrix):**
$$t_{inference} = 64 \times t_{read}$$

---

## Part 9: Practical Design Trade-Offs

### ADC Resolution vs Speed

```
ADC Architecture Comparison:

              Resolution    Speed      Power     Use Case
SAR           Good (N bits) Good       Low       Crossbar read path ✓
Flash         Excellent     Excellent  Very High Monitor/debug only
Sigma-Delta   Excellent     Slow       Medium    Not suitable
```

FeCIM uses SAR because it balances speed (50 ns) with power efficiency.

### TIA Gain vs Noise

**Higher gain (Rf):**
- Advantage: Better SNR, can sense smaller currents
- Disadvantage: Larger noise voltage, needs higher BW for settling
- Trade-off: FeCIM chose Rf = 10 kΩ for 100 nA sensitivity

**Lower gain:**
- Advantage: Faster settling, lower output noise
- Disadvantage: Can't sense small signals
- Not suitable for variable crossbar conductances

### Charge Pump Frequency vs Ripple

**Higher clock frequency:**
- Advantage: Faster rise time, more current capability
- Disadvantage: Higher switching noise/EMI, higher power
- FeCIM chose 50 MHz (good balance)

**Lower frequency:**
- Advantage: Less EMI
- Disadvantage: Slower write operations, larger ripple
- Too slow for interactive inference

---

## References and Further Reading

### In This Suite

- **[circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md)** — How CIM actually computes
- **[circuits.operations.md](circuits.operations.md)** — 0T1R vs 1T1R array architectures
- **[circuits.research.md](circuits.research.md)** — Recent peripheral circuit papers
- **[MODULE4-PHYSICS-IMPROVEMENTS.md](MODULE4-PHYSICS-IMPROVEMENTS.md)** — Physics enhancement proposals
- **[../development/TESTING.md](../development/TESTING.md)** — Testing peripheral circuits

### Related Physics

- **[../hysteresis/../hysteresis/hysteresis.physics.md](../hysteresis/../hysteresis/hysteresis.physics.md)** — Ferroelectric hysteresis
- **[../crossbar/reference/PHYSICS.md](../crossbar/reference/PHYSICS.md)** — Crossbar array physics

### External References

Standard references for equations and terminology. Unless explicitly cited above, **repo numbers are simulation defaults** (see `shared/peripherals`).

- J. F. Dickson, "On-chip high-voltage generation in MNOS integrated circuits using an improved voltage multiplier technique," *IEEE Journal of Solid-State Circuits*, 1976. (charge pumps)
- B. Razavi, *Data Conversion System Design*, IEEE Press/Wiley, 1995. (ADC architecture, noise/ENOB)
- IEEE Std 1241-2010, *IEEE Standard for Terminology and Test Methods for ADCs*. (ADC performance metrics)

---

## Code Implementation Reference

Physics models implemented in: `/shared/peripherals/`

- `adc.go` — ADC conversion, ENOB, SNR calculations
- `dac.go` — DAC conversion, settling time, energy models
- `tia.go` — Current-to-voltage, noise, settling, power
- `chargepump.go` — Voltage boosting, ripple, rise time, efficiency
- `analysis.go` — System-level timing, power, transfer functions

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite

**Last Updated:** January 2026
