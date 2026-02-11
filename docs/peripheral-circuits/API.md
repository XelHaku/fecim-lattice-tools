# Peripheral Circuits API Reference

**Version:** 1.0
**Date:** 2026-01-27
**Package:** `fecim-lattice-tools/shared/peripherals`

---

## Table of Contents

1. [Overview](#overview)
2. [ADC - Analog-to-Digital Converter](#adc---analog-to-digital-converter)
3. [DAC - Digital-to-Analog Converter](#dac---digital-to-analog-converter)
4. [TIA - Transimpedance Amplifier](#tia---transimpedance-amplifier)
5. [ChargePump - Voltage Booster](#chargepump---voltage-booster)
6. [Analysis Functions](#analysis-functions)
7. [Type Reference](#type-reference)
8. [Complete Examples](#complete-examples)

---

## Overview

The `peripherals` package provides models for key analog circuits used in FeCIM systems:

- **ADC**: Converts sensed voltages to discrete digital levels (0-31 for 5-bit)
- **DAC**: Converts digital levels to analog write voltages (±1.5V)
- **TIA**: Amplifies tiny sense currents to measurable voltages
- **ChargePump**: Boosts supply voltage to write voltage levels
- **Analysis**: Functions to characterize system performance

All peripherals are designed for the 30-level FeCIM demo baseline (simulation baseline; 5-bit storage with 2 levels reserved).

**Note:** References to 30 levels refer to this demo baseline (configurable). Literature reports multi-level states (not verified here).

### Default Configuration

All peripheral types provide `Default*()` constructors with practical values for FeCIM:

```
ADC:  5-bit, 0-1V range, 50ns conversion
DAC:  5-bit, -1.5V to +1.5V range, 10ns settle
TIA:  10kΩ gain, 100MHz bandwidth, 1pA/sqrt(Hz) noise
Pump: 2-stage, 1V input, 1.5V output, 70% efficiency
```

---

## ADC - Analog-to-Digital Converter

### Type Definition

```go
type ADC struct {
    Bits           int     // Resolution in bits (typically 5)
    VrefHigh       float64 // High reference voltage (V)
    VrefLow        float64 // Low reference voltage (V)
    INL            float64 // Integral nonlinearity (LSB)
    DNL            float64 // Differential nonlinearity (LSB)
    ConversionTime float64 // Conversion time (ns)
    Type           ADCType // Architecture: SAR, Flash, SigmaDelta
}

type ADCType int
// Constants: ADCTypeSAR, ADCTypeFlash, ADCTypeSigmaDelta
```

### Functions

#### DefaultADC

Returns an ADC configured for FeCIM read operations.

```go
func DefaultADC() *ADC
```

**Returns:** ADC with 5-bit resolution, 0-1V range, 0.5 LSB INL, 0.25 LSB DNL

**Example:**
```go
adc := peripherals.DefaultADC()
// ADC configured: 32 levels (0-31), SAR type, 50ns conversion
```

#### Convert

Performs ideal ADC conversion from voltage to digital level.

```go
func (a *ADC) Convert(voltage float64) int
```

**Parameters:**
- `voltage`: Input voltage (clamped to VrefLow-VrefHigh)

**Returns:** Digital level (0 to Levels-1)

**Behavior:**
- Input values below VrefLow clamp to 0
- Input values above VrefHigh clamp to max level
- Linear quantization between reference voltages

**Example:**
```go
adc := peripherals.DefaultADC()
// ADC range: 0V to 1V → levels 0 to 31

level := adc.Convert(0.5)  // Returns 16 (mid-scale)
level := adc.Convert(-1.0) // Returns 0 (clamped to VrefLow)
level := adc.Convert(2.0)  // Returns 31 (clamped to VrefHigh)
```

#### ConvertWithNonlinearity

Adds realistic INL/DNL errors to conversion.

```go
func (a *ADC) ConvertWithNonlinearity(voltage float64) int
```

**Parameters:**
- `voltage`: Input voltage

**Returns:** Digital level with simulated nonlinearity errors

**Nonlinearity Model:**
- INL (Integral NonLinearity): Code-dependent sine-wave error
- DNL (Differential NonLinearity): Step-to-step variation
- Both scale with LSB size and ADC parameters

**Example:**
```go
adc := peripherals.DefaultADC()

ideal := adc.Convert(0.5)              // 16
withError := adc.ConvertWithNonlinearity(0.5) // ~16 (±1 LSB typical)

// Compare ideal vs. real behavior
```

#### ENOB

Calculates Effective Number of Bits considering nonlinearity.

```go
func (a *ADC) ENOB() float64
```

**Returns:** Effective resolution in bits (Bits - log2(sqrt(1 + INL² + DNL²)))

**Formula:**
```
ENOB = Bits - log₂(√(1 + INL² + DNL²))
```

**Interpretation:**
- With 0.5 LSB INL and 0.25 LSB DNL: ENOB ≈ 4.80 bits
- Ideal ADC (INL=DNL=0): ENOB = Bits

**Example:**
```go
adc := peripherals.DefaultADC()
enob := adc.ENOB() // ≈ 4.80 bits

// Quality assessment
if enob < 4.5 {
    fmt.Println("Poor ADC linearity")
} else if enob > 4.9 {
    fmt.Println("Excellent ADC performance")
}
```

#### EffectiveSNR

Signal-to-Noise Ratio accounting for nonlinearity.

```go
func (a *ADC) EffectiveSNR() float64
```

**Returns:** SNR in dB at effective bit depth

**Formula:**
```
SNR = 6.02 * ENOB + 1.76 dB
```

**Comparison with theoretical:**
- TheoreticalSNR() uses ideal bits: `6.02 * Bits + 1.76`
- EffectiveSNR() uses ENOB: accounts for real imperfections

**Example:**
```go
adc := peripherals.DefaultADC()

theoretical := adc.TheoreticalSNR() // ~31.86 dB (ideal 5-bit)
effective := adc.EffectiveSNR()     // ~30.66 dB (with INL/DNL)

loss := theoretical - effective      // ~1.2 dB loss
```

#### EnergyPerConversion

Energy consumed per ADC conversion.

```go
func (a *ADC) EnergyPerConversion() float64
```

**Returns:** Energy in Joules per conversion

**Architecture-Dependent:**
- SAR: ~1-10 fJ per conversion-step → ~25 fJ total for 5-bit
- Flash: ~50-100 fJ per level → ~1.6 pJ for 32 levels
- Sigma-Delta: Higher due to oversampling

**Example:**
```go
adc := peripherals.DefaultADC() // Default is SAR

energy := adc.EnergyPerConversion() // ~25 fJ

// System power estimate (assuming 1M conversions/sec)
convRate := 1e6
power := energy * convRate // ~25 nW
```

#### Levels

Number of discrete output levels.

```go
func (a *ADC) Levels() int
```

**Returns:** 2^Bits

**Example:**
```go
adc := peripherals.DefaultADC()
levels := adc.Levels() // 32
```

#### Resolution

Voltage per LSB.

```go
func (a *ADC) Resolution() float64
```

**Returns:** (VrefHigh - VrefLow) / (Levels - 1)

**Example:**
```go
adc := peripherals.DefaultADC()
resolution := adc.Resolution() // 1.0 / 31 ≈ 32.3 mV per LSB
```

---

## DAC - Digital-to-Analog Converter

### Type Definition

```go
type DAC struct {
    Bits       int     // Resolution in bits (typically 5)
    VrefHigh   float64 // High reference voltage (V)
    VrefLow    float64 // Low reference voltage (V)
    INL        float64 // Integral nonlinearity (LSB)
    DNL        float64 // Differential nonlinearity (LSB)
    SettleTime float64 // Settling time (ns)
}
```

### Functions

#### DefaultDAC

Returns a DAC configured for FeCIM 30-level write operations.

```go
func DefaultDAC() *DAC
```

**Returns:** DAC with 5-bit resolution, -1.5V to +1.5V range, 10ns settle time

**Example:**
```go
dac := peripherals.DefaultDAC()
// Output range: -1.5V to +1.5V (32 levels, using 30)
```

#### Convert

Maps digital level to analog voltage.

```go
func (d *DAC) Convert(level int) float64
```

**Parameters:**
- `level`: Digital level (0 to Levels-1), clamped if out of range

**Returns:** Analog voltage

**Behavior:**
- Linear interpolation: V = VrefLow + (level/(Levels-1)) * (VrefHigh - VrefLow)
- Level 0 → VrefLow
- Level (Levels-1) → VrefHigh

**Example:**
```go
dac := peripherals.DefaultDAC()

v0 := dac.Convert(0)   // -1.5V (minimum)
v15 := dac.Convert(15) // ≈ -0.05V (near mid)
v16 := dac.Convert(16) // ≈ +0.05V (near mid)
v31 := dac.Convert(31) // +1.5V (maximum)
```

#### ConvertWithNonlinearity

Adds realistic INL/DNL errors to conversion.

```go
func (d *DAC) ConvertWithNonlinearity(level int) float64
```

**Parameters:**
- `level`: Digital level

**Returns:** Voltage with nonlinearity errors

**Nonlinearity Model:**
- INL: Code-dependent sine-wave distortion
- DNL: Level-to-level step variation
- Simulates realistic DAC mismatches

**Example:**
```go
dac := peripherals.DefaultDAC()

ideal := dac.Convert(16)              // ≈ +0.05V
withError := dac.ConvertWithNonlinearity(16) // ≈ +0.05V ± error

// Simulate write voltage uncertainty
```

#### Resolution

Voltage per LSB.

```go
func (d *DAC) Resolution() float64
```

**Returns:** (VrefHigh - VrefLow) / (Levels - 1)

**Example:**
```go
dac := peripherals.DefaultDAC()
resolution := dac.Resolution() // 3.0 / 31 ≈ 96.8 mV per LSB
```

#### EnergyPerConversion

Energy consumed per DAC conversion.

```go
func (d *DAC) EnergyPerConversion() float64
```

**Returns:** Energy in Joules per conversion

**Estimation:**
- Based on switched‑capacitor topology
- Energy ~ C_eff * (Vspan/2)² * 2^N
- Typical: ~15 fJ per conversion (C_eff≈0.2 fF, Vspan=3.0V)

**Example:**
```go
dac := peripherals.DefaultDAC()
energy := dac.EnergyPerConversion() // ~15 fJ
```

#### VoltageRange

Full output voltage span.

```go
func (d *DAC) VoltageRange() (min, max float64)
```

**Returns:** (VrefLow, VrefHigh)

**Example:**
```go
dac := peripherals.DefaultDAC()
min, max := dac.VoltageRange()
fmt.Printf("Output range: %.2fV to %.2fV\n", min, max)
// Output: -1.50V to +1.50V
```

#### Levels

Number of discrete output levels.

```go
func (d *DAC) Levels() int
```

**Returns:** 2^Bits

---

## TIA - Transimpedance Amplifier

### Type Definition

```go
type TIA struct {
    Gain             float64 // Transimpedance gain (Ω)
    Bandwidth        float64 // -3dB bandwidth (Hz)
    InputNoiseRMS    float64 // Input-referred noise (A/√Hz)
    OutputOffset     float64 // Output offset voltage (V)
    MaxInputCurrent  float64 // Maximum input current (A)
    MaxOutputVoltage float64 // Maximum output voltage (V)
}
```

### Functions

#### DefaultTIA

Returns a TIA configured for crossbar sense operations.

```go
func DefaultTIA() *TIA
```

**Returns:** TIA with 10kΩ gain, 100MHz bandwidth, 1pA/√Hz noise, 1V max output

**Example:**
```go
tia := peripherals.DefaultTIA()
// Typical FeCIM sense amplifier configuration
```

#### Convert

Performs ideal current-to-voltage conversion.

```go
func (t *TIA) Convert(current float64) float64
```

**Parameters:**
- `current`: Input current (A)

**Returns:** Output voltage (clamped to 0 - MaxOutputVoltage)

**Formula:**
```
V_out = I_in * Gain + Offset
```

**Behavior:**
- Output clamps to MaxOutputVoltage (saturation)
- Negative outputs clamp to 0

**Example:**
```go
tia := peripherals.DefaultTIA()

i1 := 10e-6  // 10 µA
v1 := tia.Convert(i1)
// v1 = 10e-6 * 10e3 + 0.005 = 0.105V

i2 := 100e-6 // 100 µA (exceeds max input)
v2 := tia.Convert(i2)
// v2 = saturated at 1.0V
```

#### ConvertWithNoise

Current-to-voltage conversion with simulated thermal noise.

```go
func (t *TIA) ConvertWithNoise(current float64) float64
```

**Parameters:**
- `current`: Input current

**Returns:** Voltage with noise contribution

**Noise Model:**
- Thermal noise: V_noise = I_noise * Gain * √(Bandwidth)
- Added at ~10% RMS for demonstration
- In hardware: use actual random noise generation

**Example:**
```go
tia := peripherals.DefaultTIA()

ideal := tia.Convert(10e-6)       // Clean conversion
noisy := tia.ConvertWithNoise(10e-6) // With noise added

diff := noisy - ideal // Small deviation due to noise
```

#### SNR

Signal-to-Noise Ratio for a given input current.

```go
func (t *TIA) SNR(current float64) float64
```

**Parameters:**
- `current`: Input signal current

**Returns:** SNR in dB

**Formula:**
```
SNR = 20 * log₁₀(Signal / Noise)
    = 20 * log₁₀(I_in * Gain / (I_noise * Gain * √BW))
```

**Interpretation:**
- Higher input current → Higher SNR
- Limited by thermal noise floor

**Example:**
```go
tia := peripherals.DefaultTIA()

snr10 := tia.SNR(10e-6)  // 10 µA signal
snr100 := tia.SNR(100e-6) // 100 µA signal
// snr100 > snr10 by ~20 dB (10× current increase)
```

#### MinDetectableCurrent

Minimum input current detectable (SNR = 1).

```go
func (t *TIA) MinDetectableCurrent() float64
```

**Returns:** Minimum current in Amps

**Formula:**
```
I_min = I_noise * √(Bandwidth)
```

**Interpretation:**
- Below this current, signal lost in noise
- Defines amplifier sensitivity limit

**Example:**
```go
tia := peripherals.DefaultTIA()
imin := tia.MinDetectableCurrent() // ~100 pA typical
```

#### DynamicRange

Dynamic range in dB (ratio of max to min detectable current).

```go
func (t *TIA) DynamicRange() float64
```

**Returns:** Dynamic range in dB

**Formula:**
```
DR = 20 * log₁₀(I_max / I_min)
```

**Example:**
```go
tia := peripherals.DefaultTIA()
dr := tia.DynamicRange() // ~120 dB typical
```

#### SettlingTime

Estimated step response settling time (10%-90% criterion).

```go
func (t *TIA) SettlingTime() float64
```

**Returns:** Settling time in seconds (typically nanoseconds)

**Formula:**
```
t_settle = ln(1/accuracy) / (2π * Bandwidth)
         = 6.9 / (2π * BW) for 0.1% accuracy
```

**Example:**
```go
tia := peripherals.DefaultTIA()
tsettle := tia.SettlingTime() // ~11 ns typical

// Verify TIA is fast enough for read operations
if tsettle < 76e-9 { // 76 ns budget
    fmt.Println("TIA settling acceptable for 76ns read")
}
```

#### PowerConsumption

Estimated power consumption.

```go
func (t *TIA) PowerConsumption() float64
```

**Returns:** Power in Watts

**Estimation:**
- Based on: P ≈ 2 * kT * Bandwidth * Gain / efficiency
- kT at 300K: 4.14e-21 J
- Efficiency: 0.1 (10% - typical for TIA)

**Example:**
```go
tia := peripherals.DefaultTIA()
power := tia.PowerConsumption() // Watts (very small, nW range)
```

---

## ChargePump - Voltage Booster

### Type Definition

```go
type ChargePump struct {
    InputVoltage   float64 // Supply voltage (V)
    OutputVoltage  float64 // Target output voltage (V)
    Stages         int     // Number of pump stages
    DiodeDrop      float64 // Effective diode/switch drop per stage (V)
    ClockFrequency float64 // Pump clock frequency (Hz)
    LoadCurrent    float64 // Maximum load current (A)
    FlyCapacitance float64 // Flying capacitor value (F)
    Efficiency     float64 // Power conversion efficiency (0-1)
}
```

### Functions

#### DefaultChargePump

Returns a charge pump for FeCIM write operations.

```go
func DefaultChargePump() *ChargePump
```

**Returns:** 2-stage Dickson pump, 1V→1.5V boost, 70% efficiency

**Example:**
```go
pump := peripherals.DefaultChargePump()
// Boosts 1V supply to 1.5V for positive writes
```

#### NegativePump

Creates charge pump for negative write voltage.

```go
func NegativePump() *ChargePump
```

**Returns:** ChargePump configured for -1.5V output

**Example:**
```go
posPump := peripherals.DefaultChargePump()    // +1.5V
negPump := peripherals.NegativePump()         // -1.5V

// Dual pump system for bipolar writes
```

#### IdealOutputVoltage

Theoretical maximum output voltage (no losses).

```go
func (c *ChargePump) IdealOutputVoltage() float64
```

**Returns:** Voltage in Volts

**Formula (Dickson pump):**
```
V_ideal = (N+1) * V_in
        = (Stages + 1) * InputVoltage
```

**Example:**
```go
pump := peripherals.DefaultChargePump()
videal := pump.IdealOutputVoltage() // (2+1) * 1.0 = 3.0V
```

#### ActualOutputVoltage

Realistic output voltage accounting for losses.

```go
func (c *ChargePump) ActualOutputVoltage() float64
```

**Returns:** Voltage in Volts

**Loss Mechanisms:**
- Threshold voltage drops (MOS switches)
- IR drops in capacitors and switches
- Incomplete charge transfer

**Formula:**
```
V_unreg = (N+1)*V_in - N*V_drop - I_load*R_out
R_out ≈ 1 / (C_fly * f_clk)
V_actual = sign(V_target) * min(|V_target|, |V_unreg|)
```

**Example:**
```go
pump := peripherals.DefaultChargePump()
vactual := pump.ActualOutputVoltage() // ≈1.5V (regulated; unreg ≈2.4V)
```

#### OutputRipple

Peak-to-peak voltage ripple on output.

```go
func (c *ChargePump) OutputRipple() float64
```

**Returns:** Ripple voltage in Volts

**Formula:**
```
ΔV = I_load / (C_out * f_clock)
```

**Where:**
- C_out ≈ 10 * FlyCapacitance (assumed storage cap)

**Example:**
```go
pump := peripherals.DefaultChargePump()
ripple := pump.OutputRipple() // ≈0.2 mV with default caps
```

#### BoostFactor

Voltage multiplication ratio.

```go
func (c *ChargePump) BoostFactor() float64
```

**Returns:** V_actual / V_in

**Example:**
```go
pump := peripherals.DefaultChargePump()
boost := pump.BoostFactor() // ≈1.5 (regulated output / input)
```

#### MaxCurrentCapability

Maximum sustainable output current.

```go
func (c *ChargePump) MaxCurrentCapability() float64
```

**Returns:** Current in Amps

**Formula:**
```
I_max = C_fly * f_clock * (N+1) * V_in / V_out
```

**Interpretation:**
- Limited by capacitor charge delivery rate
- Higher frequency → Higher current capability

**Example:**
```go
pump := peripherals.DefaultChargePump()
imax := pump.MaxCurrentCapability() // ≈10 mA with default caps
```

#### PowerInput / PowerOutput / PowerLoss

Power analysis.

```go
func (c *ChargePump) PowerInput() float64
func (c *ChargePump) PowerOutput() float64
func (c *ChargePump) PowerLoss() float64
```

**Returns:** Power in Watts

**Relationship:**
```
P_in = P_out / Efficiency
P_loss = P_in - P_out
```

**Example:**
```go
pump := peripherals.DefaultChargePump()

pout := pump.PowerOutput() // V_out * I_load
pin := pump.PowerInput()   // pout / efficiency
ploss := pump.PowerLoss()  // pin - pout

efficiency := pout / pin   // Should match pump.Efficiency
```

#### RiseTime

Output voltage rise time (10%-90%).

```go
func (c *ChargePump) RiseTime() float64
```

**Returns:** Time in seconds

**Formula:**
```
t_rise = (Stages * 2.2) / f_clock
```

**Interpretation:**
- How fast pump reaches output voltage
- Important for write timing constraints

**Example:**
```go
pump := peripherals.DefaultChargePump()
trise := pump.RiseTime() // ≈90 ns typical
```

#### EnergyPerOperation

Energy for one write voltage pulse.

```go
func (c *ChargePump) EnergyPerOperation(pulseDuration float64) float64
```

**Parameters:**
- `pulseDuration`: Write pulse width (seconds)

**Returns:** Energy in Joules

**Formula:**
```
E = P_in * t_pulse
```

**Example:**
```go
pump := peripherals.DefaultChargePump()

// Typical write pulse: 100 ns
energy := pump.EnergyPerOperation(100e-9)
```

#### ChargeTransferEfficiency

Per-stage charge transfer efficiency.

```go
func (c *ChargePump) ChargeTransferEfficiency() float64
```

**Returns:** Efficiency ratio (0-1)

**Formula:**
```
η = V_actual / V_ideal
```

**Example:**
```go
pump := peripherals.DefaultChargePump()
eta := pump.ChargeTransferEfficiency() // ≈0.65-0.75
```

#### Area

Estimated silicon area (capacitor-dominated).

```go
func (c *ChargePump) Area() float64
```

**Returns:** Area in µm²

**Example:**
```go
pump := peripherals.DefaultChargePump()
area := pump.Area() // Typically 100-1000 µm²
```

#### SupportsLevel

Check if pump can generate voltage for a given state.

```go
func (c *ChargePump) SupportsLevel(level int, maxLevel int) bool
```

**Parameters:**
- `level`: Target state (0 to maxLevel)
- `maxLevel`: Maximum state value

**Returns:** true if pump can reach required voltage

**Example:**
```go
pump := peripherals.DefaultChargePump()

// For 30-level system
supports20 := pump.SupportsLevel(20, 30) // true
supports30 := pump.SupportsLevel(30, 30) // true/false (depends on losses)
```

---

## Analysis Functions

### AnalyzeINLDNL

Detailed INL/DNL analysis for DAC or ADC.

```go
func (d *DAC) AnalyzeINLDNL() *INLDNLAnalysis
func (a *ADC) AnalyzeINLDNL() *INLDNLAnalysis
```

**Returns:** `*INLDNLAnalysis` with comprehensive nonlinearity metrics

**Example:**
```go
dac := peripherals.DefaultDAC()
analysis := dac.AnalyzeINLDNL()

fmt.Printf("Max INL: %.2f LSB at code %d\n",
    analysis.MaxINL, analysis.WorstCode)
fmt.Printf("Max DNL: %.2f LSB\n", analysis.MaxDNL)
fmt.Printf("Min DNL: %.2f LSB\n", analysis.MinDNL)
```

### AnalyzeTiming

Complete timing analysis for peripheral system.

```go
func AnalyzeTiming(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TimingAnalysis
```

**Parameters:** All four peripheral types

**Returns:** `*TimingAnalysis` with read/write/cycle timing

**Calculated Values:**
- DACSettle: Time for DAC output to stabilize
- ArraySettle: Array RC/sneak settling time
- PumpRise: Charge pump voltage rise time
- WritePulse: Program pulse width
- WriteTime: Total write operation duration
- TIASettle: Transimpedance amp settling
- ADCConvert: ADC conversion time
- ReadTime: Total read operation duration
- CycleTime: Full read+write cycle
- MaxThroughput: Operations per second

**Example:**
```go
dac := peripherals.DefaultDAC()
adc := peripherals.DefaultADC()
tia := peripherals.DefaultTIA()
pump := peripherals.DefaultChargePump()

timing := peripherals.AnalyzeTiming(dac, adc, tia, pump)

fmt.Printf("Write time: %.1f ns\n", timing.WriteTime*1e9)
fmt.Printf("Read time: %.1f ns\n", timing.ReadTime*1e9)
fmt.Printf("Throughput: %.1f kops/s\n", timing.MaxThroughput/1e3)
```

### AnalyzePower

Power consumption breakdown.

```go
func AnalyzePower(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump,
                  timing *TimingAnalysis) *PowerBreakdown
```

**Parameters:**
- Peripheral types
- TimingAnalysis (for cycle time)

**Returns:** `*PowerBreakdown` with component and total power

**Calculated Values:**
- Energy per operation for each component
- Power averaged over cycle time
- Fractional contribution (%)
- PumpEnergy corresponds to the write pulse; total energy is a full read+write cycle

**Example:**
```go
timing := peripherals.AnalyzeTiming(dac, adc, tia, pump)
power := peripherals.AnalyzePower(dac, adc, tia, pump, timing)

fmt.Printf("DAC Energy: %.2f fJ (%.1f%%)\n",
    power.DACEnergy*1e15, power.DACFraction*100)
fmt.Printf("ADC Energy: %.2f fJ (%.1f%%)\n",
    power.ADCEnergy*1e15, power.ADCFraction*100)
fmt.Printf("TIA Energy: %.2f fJ (%.1f%%)\n",
    power.TIAEnergy*1e15, power.TIAFraction*100)
fmt.Printf("Pump Energy: %.2f fJ (%.1f%%)\n",
    power.PumpEnergy*1e15, power.PumpFraction*100)
fmt.Printf("Total: %.2f fJ/op\n", power.TotalEnergy*1e15)
```

### ComputeTransferFunction

Trace signal through entire peripheral chain.

```go
func ComputeTransferFunction(dac *DAC, adc *ADC, tia *TIA,
                             pump *ChargePump) *TransferFunction
```

**Parameters:** All four peripheral types

**Returns:** `*TransferFunction` with signal levels at each stage

**Output Arrays (30 entries each):**
- InputLevels: Digital input (0-29)
- DACVoltages: After D/A conversion
- PumpVoltages: After charge pump (if applicable)
- TIAVoltages: After transimpedance amp
- ADCLevels: Final digital output
- Errors: Output - Input

**Example:**
```go
tf := peripherals.ComputeTransferFunction(dac, adc, tia, pump)

// Check round-trip fidelity
maxError := 0
for _, err := range tf.Errors {
    if err < 0 {
        err = -err
    }
    if err > maxError {
        maxError = err
    }
}
fmt.Printf("Maximum round-trip error: ±%d levels\n", maxError)

// Analyze system linearity
for level := 0; level < 30; level++ {
    fmt.Printf("Level %2d: DAC=%.3fV → TIA=%.3fV → ADC=%2d (err=%+2d)\n",
        tf.InputLevels[level],
        tf.DACVoltages[level],
        tf.TIAVoltages[level],
        tf.ADCLevels[level],
        tf.Errors[level])
}
```

---

## Type Reference

### INLDNLAnalysis

```go
type INLDNLAnalysis struct {
    Levels    int       // Number of levels analyzed
    INLValues []float64 // INL at each code (in LSB)
    DNLValues []float64 // DNL at each code (in LSB)
    MaxINL    float64   // Worst-case INL
    MaxDNL    float64   // Worst-case positive DNL
    MinDNL    float64   // Worst-case negative DNL
    WorstCode int       // Code with maximum INL
}
```

**Usage:**
- INL: Deviation from ideal straight line
- DNL: Deviation of step size from ideal
- Negative DNL below -1 LSB causes missing codes
- Positive DNL causes code compression

### TimingAnalysis

```go
type TimingAnalysis struct {
    DACSettle     float64 // DAC settling time (s)
    ArraySettle   float64 // Array RC/sneak settling (s)
    PumpRise      float64 // Charge pump rise time (s)
    WritePulse    float64 // Write pulse width (s)
    WriteTime     float64 // Total write time (s)
    TIASettle     float64 // TIA settling time (s)
    ADCConvert    float64 // ADC conversion time (s)
    ReadTime      float64 // Total read time (s)
    CycleTime     float64 // Full read+write cycle (s)
    MaxThroughput float64 // Maximum operations per second
}
```

### PowerBreakdown

```go
type PowerBreakdown struct {
    // Power (W)
    DACPower   float64
    ADCPower   float64
    TIAPower   float64
    PumpPower  float64
    TotalPower float64

    // Energy per operation (J)
    DACEnergy   float64
    ADCEnergy   float64
    TIAEnergy   float64
    PumpEnergy  float64
    TotalEnergy float64

    // Fractional contribution (0-1)
    DACFraction  float64
    ADCFraction  float64
    TIAFraction  float64
    PumpFraction float64
}
```

### TransferFunction

```go
type TransferFunction struct {
    InputLevels  []int     // Digital input (0-29)
    DACVoltages  []float64 // After D/A (V)
    PumpVoltages []float64 // After pump (V)
    TIAVoltages  []float64 // After TIA (V)
    ADCLevels    []int     // Final digital output
    Errors       []int     // Output - Input
}
```

---

## Complete Examples

### Example 1: Characterize a FeCIM Read Path

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/peripherals"
)

func main() {
    // Create default FeCIM peripherals
    tia := peripherals.DefaultTIA()
    adc := peripherals.DefaultADC()

    // Configure ADC to match TIA output range
    adc.VrefLow = tia.OutputOffset
    adc.VrefHigh = tia.MaxOutputVoltage

    // Simulate reading different currents
    testCurrents := []float64{1e-6, 10e-6, 50e-6, 100e-6}

    fmt.Println("=== FeCIM Read Path Characterization ===")
    fmt.Println("Current (µA) | TIA Output (V) | ADC Level | SNR (dB)")
    fmt.Println("---------------------------------------------------")

    for _, i := range testCurrents {
        vout := tia.Convert(i)
        level := adc.Convert(vout)
        snr := tia.SNR(i)

        fmt.Printf("   %6.1f    |    %.4f       |   %2d     | %.1f\n",
            i*1e6, vout, level, snr)
    }

    // System characterization
    fmt.Println("\n=== System Metrics ===")
    fmt.Printf("TIA Gain: %.1f kΩ\n", tia.Gain/1e3)
    fmt.Printf("TIA Bandwidth: %.0f MHz\n", tia.Bandwidth/1e6)
    fmt.Printf("TIA Min Detectable Current: %.1f pA\n",
        tia.MinDetectableCurrent()*1e12)
    fmt.Printf("ADC ENOB: %.2f bits\n", adc.ENOB())
    fmt.Printf("ADC Effective SNR: %.1f dB\n", adc.EffectiveSNR())
}
```

### Example 2: Write Path with Charge Pump

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/peripherals"
)

func main() {
    dac := peripherals.DefaultDAC()
    posPump := peripherals.DefaultChargePump()
    negPump := peripherals.NegativePump()

    fmt.Println("=== Write Voltage Generation ===\n")

    // Test positive pump
    fmt.Println("Positive Pump (for levels 15-29):")
    fmt.Printf("Supply: %.2fV → Target: %.2fV\n",
        posPump.InputVoltage, posPump.OutputVoltage)
    fmt.Printf("Ideal Output: %.2fV\n", posPump.IdealOutputVoltage())
    fmt.Printf("Actual Output: %.2fV\n", posPump.ActualOutputVoltage())
    fmt.Printf("Boost Factor: %.2f×\n", posPump.BoostFactor())
    fmt.Printf("Output Ripple: %.1f mV\n", posPump.OutputRipple()*1e3)
    fmt.Printf("Max Current: %.1f µA\n",
        posPump.MaxCurrentCapability()*1e6)
    fmt.Printf("Rise Time: %.1f ns\n", posPump.RiseTime()*1e9)

    // Test negative pump
    fmt.Println("\nNegative Pump (for levels 0-15):")
    fmt.Printf("Output: %.2fV\n", negPump.ActualOutputVoltage())

    // DAC conversion for all 30 levels
    fmt.Println("\n=== Level-to-Voltage Mapping ===")
    fmt.Println("Level | DAC Voltage (V)")
    fmt.Println("------------------------")
    for level := 0; level < 30; level += 5 {
        v := dac.Convert(level)
        fmt.Printf(" %2d   |    %.3f\n", level, v)
    }
}
```

### Example 3: Complete System Characterization

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/peripherals"
)

func main() {
    // Create peripheral set
    dac := peripherals.DefaultDAC()
    adc := peripherals.DefaultADC()
    tia := peripherals.DefaultTIA()
    pump := peripherals.DefaultChargePump()

    // Configure ADC for read
    adc.VrefLow = tia.OutputOffset
    adc.VrefHigh = tia.MaxOutputVoltage

    // Analyze complete system
    timing := peripherals.AnalyzeTiming(dac, adc, tia, pump)
    power := peripherals.AnalyzePower(dac, adc, tia, pump, timing)

    fmt.Println("=== FeCIM Peripheral System Summary ===\n")

    // Timing
    fmt.Println("TIMING:")
    fmt.Printf("  Write Path:  %6.1f ns (DAC: %.1f, Pump: %.1f)\n",
        timing.WriteTime*1e9,
        timing.DACSettle*1e9,
        timing.PumpRise*1e9)
    fmt.Printf("  Read Path:   %6.1f ns (TIA: %.1f, ADC: %.1f)\n",
        timing.ReadTime*1e9,
        timing.TIASettle*1e9,
        timing.ADCConvert*1e9)
    fmt.Printf("  Cycle Time:  %6.1f ns\n", timing.CycleTime*1e9)
    fmt.Printf("  Throughput:  %6.1f kops/s\n", timing.MaxThroughput/1e3)

    // Power
    fmt.Println("\nPOWER PER CYCLE:")
    fmt.Printf("  DAC:   %8.2f fJ (%.1f%%)\n",
        power.DACEnergy*1e15, power.DACFraction*100)
    fmt.Printf("  ADC:   %8.2f fJ (%.1f%%)\n",
        power.ADCEnergy*1e15, power.ADCFraction*100)
    fmt.Printf("  TIA:   %8.2f fJ (%.1f%%)\n",
        power.TIAEnergy*1e15, power.TIAFraction*100)
    fmt.Printf("  Pump:  %8.2f fJ (%.1f%%)\n",
        power.PumpEnergy*1e15, power.PumpFraction*100)
    fmt.Printf("  Total: %8.2f fJ/cycle\n", power.TotalEnergy*1e15)
    fmt.Printf("  Power: %8.2f nW (avg)\n", power.TotalPower*1e9)

    // Transfer function
    tf := peripherals.ComputeTransferFunction(dac, adc, tia, pump)
    maxErr := 0
    for _, err := range tf.Errors {
        if err < 0 {
            err = -err
        }
        if err > maxErr {
            maxErr = err
        }
    }

    fmt.Println("\nLINEARITY:")
    fmt.Printf("  Round-trip error: ±%d levels (max)\n", maxErr)

    // DAC/ADC quality
    dacINL := dac.AnalyzeINLDNL()
    adcINL := adc.AnalyzeINLDNL()

    fmt.Println("\nCONVERTER QUALITY:")
    fmt.Printf("  DAC: MaxINL=%.2f LSB, MaxDNL=%.2f LSB\n",
        dacINL.MaxINL, dacINL.MaxDNL)
    fmt.Printf("  ADC: MaxINL=%.2f LSB, MaxDNL=%.2f LSB, ENOB=%.2f bits\n",
        adcINL.MaxINL, adcINL.MaxDNL, adc.ENOB())
}
```

---

## Notes and Best Practices

### Configuration Guidelines

- **ADC Range**: Match to expected sense amplifier output voltage
- **DAC Range**: ±1.5V standard for FeCIM bipolar writes
- **TIA Gain**: 10kΩ typical for 1µA-100µA currents
- **Pump Frequency**: 50MHz minimizes output ripple vs. power

### Nonlinearity Interpretation

- **INL < 0.5 LSB**: Excellent
- **0.5-1.0 LSB**: Good
- **1.0-2.0 LSB**: Acceptable
- **> 2.0 LSB**: Poor, verify design

- **DNL < -1.0 LSB**: Non-monotonic (missing codes)
- **DNL < 0.5 LSB**: Excellent
- **0.5-1.0 LSB**: Good
- **> 1.0 LSB**: Poor

### Timing Constraints (Model Defaults)

For read operation target:
- DAC settle: 10ns
- Array settle: 5ns
- TIA settle: ~11ns
- ADC convert: 50ns
- Total read: ~76ns

For complete cycle (read + write):
- Write: ~203ns
- Read: ~76ns
- Total: ~279ns

### Power Estimation (Model Defaults)

Typical per-operation breakdown:
- DAC: ~14.4 fJ (switching)
- ADC: ~25 fJ (SAR conversion)
- TIA: ~6.3 fJ per read window
- Pump: ~2.14 pJ per write (input energy)
- Total read: ~46 fJ
- Total write: ~2.19 pJ

---

## Related Documentation

- Physics background: [PHYSICS.md](PHYSICS.md)
- Operation guide: [circuits.operations.md](circuits.operations.md)
- ELI5 explanation: [circuits.ELI5.md](circuits.ELI5.md)
- Research papers: [circuits.research.md](circuits.research.md)
