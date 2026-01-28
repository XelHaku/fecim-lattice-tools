# Peripheral Circuits Explained Like I'm 5

**Understanding ADCs, DACs, TIAs, and Charge Pumps Through Simple Analogies + Production Module Specification**

---

## Part 1: The Translator Analogy

### What Are Peripheral Circuits?

Imagine your brain only speaks English, but the computer memory only speaks French. You need **translators** to communicate!

```
Your brain (digital)  ←→  Translators  ←→  Memory (analog)
    0, 1, 2...        │   DAC, ADC   │    voltages, currents
                      │   TIA, etc.  │
```

**Peripheral circuits are the translators** between the digital world (numbers) and the analog world (voltages and currents).

---

## Part 2: The Four Main Translators

### 2.1 DAC - The Number-to-Voltage Translator

**DAC = Digital-to-Analog Converter**

Think of a DAC like a **water faucet with numbered settings**:

```
Turn dial to: 0  → Tiny drip (0.0 V)
Turn dial to: 15 → Medium flow (0.75 V)
Turn dial to: 29 → Full blast (1.5 V)

The number you pick → The voltage you get
```

**In FeCIM:**
- We have 30 settings (levels 0-29)
- Each setting tells the memory cell "store this much"
- The DAC converts "level 15" into "0.75 volts"

```
   DIGITAL                    ANALOG
┌─────────┐               ┌─────────────┐
│   15    │ ──→  DAC  ──→ │   0.75 V    │
└─────────┘               └─────────────┘
  (number)                   (voltage)
```

---

### 2.2 ADC - The Voltage-to-Number Translator

**ADC = Analog-to-Digital Converter**

Think of an ADC like a **measuring cup with marked lines**:

```
Look at the water level:
  Below line 5?  → Report "4"
  Between 5-10?  → Report "7"
  Between 10-15? → Report "12"
  Above line 25? → Report "28"

The voltage you see → The number you report
```

**In FeCIM:**
- We read the memory cell's current
- The ADC converts "0.75 volts" into "level 15"
- Now the computer knows what was stored!

```
   ANALOG                     DIGITAL
┌─────────────┐           ┌─────────┐
│   0.75 V    │ ──→ ADC ──→│   15    │
└─────────────┘           └─────────┘
  (voltage)                 (number)
```

---

### 2.3 TIA - The Current Magnifier

**TIA = Transimpedance Amplifier**

Think of a TIA like a **magnifying glass for electricity**:

```
Memory cell output: Tiny current (1 microamp = ant-sized)
                         │
                         ▼
                   ┌─────────┐
                   │   TIA   │  ← Magnifies by 10,000×
                   └─────────┘
                         │
                         ▼
TIA output:        Big voltage (10 millivolts = elephant-sized)
```

**Why do we need it?**
- Memory cells produce TINY currents (like whispers)
- The ADC needs BIGGER voltages (like normal talking)
- The TIA amplifies the signal so it can be "heard"

```
     TINY                    MAGNIFIER                 BIG
┌───────────┐               ┌─────────┐           ┌───────────┐
│  1 µA     │ ──────────→  │   TIA   │ ─────────→│  10 mV    │
│ (whisper) │               │  ×10k   │           │ (talking) │
└───────────┘               └─────────┘           └───────────┘
```

---

### 2.4 Charge Pump - The Voltage Booster

**Charge Pump = Voltage Step-Up Transformer**

Think of a charge pump like a **ladder for voltage**:

```
Modern chips run on: 1.0 V (not very tall)
FeFET needs:         1.5 V (taller!)

Problem: We need to reach higher!

      ╔═══════╗
      ║  1.5V ║ ← Memory cell "shelf"
      ╠═══════╣
      ║       ║
      ║       ║
──────╩───────╩──── 1.0V power supply "floor"

Solution: Charge Pump = Ladder!

      ╔═══════╗
      ║  1.5V ║ ← Now we can reach!
      ╠═══╪═══╣
      ║   │   ║  Charge Pump
      ║   │   ║  (ladder)
──────╩───┴───╩────
        ↑
    1.0V supply
```

**In FeCIM:**
- The chip's power supply is only 1.0V
- FeFET cells need 1.5V to switch polarization
- The charge pump "boosts" 1.0V up to 1.5V
- It also makes -1.5V for switching the other direction

---

## Part 3: The Complete Signal Chain

### How They All Work Together

When you **WRITE** to memory:

```
Step 1: Pick a number (0-29)
     ↓
Step 2: DAC converts to voltage
     ↓
Step 3: Charge Pump boosts voltage
     ↓
Step 4: Voltage programs the cell

COMPUTER         DAC          PUMP         MEMORY
┌─────┐       ┌─────┐       ┌─────┐       ┌─────┐
│ 15  │ ────→ │0.5V │ ────→ │1.5V │ ────→ │ ⬛  │
└─────┘       └─────┘       └─────┘       └─────┘
```

When you **READ** from memory:

```
Step 1: Apply small voltage to cell
     ↓
Step 2: Cell produces current based on stored value
     ↓
Step 3: TIA amplifies tiny current to measurable voltage
     ↓
Step 4: ADC converts voltage back to number

MEMORY          TIA           ADC         COMPUTER
┌─────┐       ┌─────┐       ┌─────┐       ┌─────┐
│ ⬛  │ ────→ │×10k │ ────→ │READ │ ────→ │ 15  │
└─────┘       └─────┘       └─────┘       └─────┘
  5µA          50mV         level 15
```

---

## Part 4: Why This Matters

### The Power Problem

Here's the shocking truth: **The translators use MORE power than the memory itself!**

```
Power Usage Breakdown:
┌────────────────────────────────────────────────────────────┐
│ ████████████████████████████████████████████████  80%     │ ← ADC (!!)
│ ████████                                          20%     │ ← Everything else
└────────────────────────────────────────────────────────────┘

The ADC is a power HOG!
```

**Why?**
- The ADC must compare the voltage against MANY levels
- Each comparison uses energy
- More precision (bits) = more comparisons = more power

**The good news:** FeCIM only needs 5-bit precision (30 levels), not 8 or 10 bits. This saves LOTS of power!

---

### The Speed Problem

Each translator takes time:

```
Write Operation Timeline:
│────────│────────────│──────────────────────────│
   DAC      Pump Rise        Write Pulse
   10ns       40ns              100ns
                                            Total: ~150 ns

Read Operation Timeline:
│────────│────────────│
   TIA      ADC Convert
   10ns       50ns
                       Total: ~60 ns

Full Cycle: ~210 ns = ~5 million operations per second
```

**The bottleneck:** The charge pump and write pulse are the slowest. But since we read MUCH more than write, overall speed is good!

---

## Part 5: Problems and Solutions

### Problem 1: Imprecise Translators (INL/DNL)

The DAC and ADC aren't perfect—they make small mistakes:

```
Ideal DAC:              Real DAC:
Level 10 → 0.500V       Level 10 → 0.503V  (0.003V error!)
Level 11 → 0.533V       Level 11 → 0.530V  (uneven step!)
Level 12 → 0.567V       Level 12 → 0.569V

INL = How far off from perfect line (Integral Non-Linearity)
DNL = How uneven the steps are (Differential Non-Linearity)
```

**Solution:** Keep INL and DNL under 0.5 LSB (half a step). Our design: INL=0.5, DNL=0.25 ✓

---

### Problem 2: Noisy TIA

The TIA adds noise when amplifying:

```
Ideal signal:  ─────────────────────────────

Real signal:   ─────∼∼∼∼∼∼∼∼∼∼∼∼∼∼∼──────
                         ↑
                    Noise (wiggles)
```

**Solution:** Use wider bandwidth (faster settling) and lower input noise. Our design: 100MHz BW, 1 pA/√Hz noise ✓

---

### Problem 3: Charge Pump Ripple

The charge pump output isn't perfectly smooth:

```
Ideal output:    ────────────────────
                 1.5V

Real output:     ─╱╲╱╲╱╲╱╲╱╲╱╲─
                 1.5V ± ripple
```

**Solution:** Add big output capacitors and use fast clock. Our design: 100pF fly caps, 50MHz clock ✓

---

## Part 6: Summary for Kids

| Component | What It Does | Like a... |
|-----------|--------------|-----------|
| **DAC** | Number → Voltage | Water faucet with settings |
| **ADC** | Voltage → Number | Measuring cup with lines |
| **TIA** | Amplifies current | Magnifying glass |
| **Charge Pump** | Boosts voltage | Ladder to reach higher |

---

## One Sentence Summary

> **Peripheral circuits are the translators that let computers talk to memory cells: DAC and charge pump for writing, TIA and ADC for reading.**

---

## Want to Go Deeper?

For the physics behind CIM operations, see **[circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md)** which explains:
- How Ohm's Law enables in-memory multiplication
- How Kirchhoff's Current Law enables accumulation
- Non-destructive read mechanisms
- Partial polarization for multi-level states

---

# Part 7: Perfect Peripheral Module Specification for FeCIM

## What the Production Module MUST Do

Based on research from 30+ papers and our implementation analysis, here's the specification for a production-ready peripheral circuit simulation module:

---

### 7.1 ADC Specification

```go
// REQUIRED: SAR ADC for crossbar read operations
type ADC interface {
    // Core conversion
    Convert(voltage float64) int                    // Ideal
    ConvertWithNonlinearity(voltage float64) int    // With errors

    // Configuration
    SetBits(n int)           // Resolution (default: 5)
    SetVref(low, high float64)  // Reference range

    // Analysis
    ENOB() float64           // Effective number of bits
    SNR() float64            // Signal-to-noise ratio (dB)
    EnergyPerConversion() float64  // Energy (J)
}

type ADCConfig struct {
    Bits           int       // 5 bits (32 levels ≥ 30 FeCIM levels)
    VrefLow        float64   // 0.0 V
    VrefHigh       float64   // 1.0 V
    INL            float64   // <0.5 LSB
    DNL            float64   // <0.5 LSB
    ConversionTime float64   // 50 ns (SAR)
    Type           ADCType   // SAR (recommended)
}
```

**Requirements:**
- 5-bit resolution minimum (matches 30 FeCIM levels)
- SAR architecture for best energy efficiency
- INL < 0.5 LSB, DNL < 0.5 LSB
- Conversion time < 100 ns
- Energy < 50 fJ per conversion

---

### 7.2 DAC Specification

```go
// REQUIRED: DAC for crossbar write operations
type DAC interface {
    // Core conversion
    Convert(level int) float64                    // Ideal
    ConvertWithNonlinearity(level int) float64    // With errors

    // Configuration
    SetBits(n int)
    SetVref(low, high float64)  // ±1.5V for FeFET

    // Analysis
    Resolution() float64        // V per LSB
    EnergyPerConversion() float64
    IsMonotonic() bool          // DNL > -1 LSB
}

type DACConfig struct {
    Bits       int       // 5 bits
    VrefLow    float64   // -1.5 V (for negative write)
    VrefHigh   float64   // +1.5 V (for positive write)
    INL        float64   // <0.5 LSB
    DNL        float64   // <0.5 LSB (guarantee monotonicity)
    SettleTime float64   // 10 ns
}
```

**Requirements:**
- 5-bit resolution
- Bipolar output: ±1.5V for FeFET switching
- Monotonic (no missing codes)
- Settling time < 20 ns
- Energy < 20 fJ per conversion

---

### 7.3 TIA Specification

```go
// REQUIRED: Transimpedance amplifier for current sensing
type TIA interface {
    // Core conversion
    Convert(current float64) float64              // I → V
    ConvertWithNoise(current float64) float64     // With noise

    // Analysis
    SNR(current float64) float64    // For given input
    MinDetectableCurrent() float64  // Noise floor
    DynamicRange() float64          // dB
    SettlingTime() float64          // 10-90%
}

type TIAConfig struct {
    Gain             float64   // 10 kΩ transimpedance
    Bandwidth        float64   // 100 MHz
    InputNoiseRMS    float64   // 1 pA/√Hz
    OutputOffset     float64   // <10 mV
    MaxInputCurrent  float64   // 100 µA
    MaxOutputVoltage float64   // 1.0 V (ADC range)
}
```

**Requirements:**
- Gain: 10-50 kΩ (map 1-100 µA → 10 mV - 1 V)
- Bandwidth: 50-200 MHz (fast settling)
- Input noise: < 5 pA/√Hz
- Dynamic range: > 40 dB
- Settling time: < 20 ns

---

### 7.4 Charge Pump Specification

```go
// REQUIRED: Charge pump for write voltage generation
type ChargePump interface {
    // Output characteristics
    IdealOutputVoltage() float64
    ActualOutputVoltage(load float64) float64
    OutputRipple(load float64) float64

    // Power analysis
    PowerInput() float64
    PowerOutput() float64
    Efficiency() float64

    // Timing
    RiseTime() float64
    MaxCurrentCapability() float64
}

type ChargePumpConfig struct {
    InputVoltage   float64   // 1.0 V (CMOS supply)
    OutputVoltage  float64   // 1.5 V (or -1.5 V)
    Stages         int       // 2 (Dickson)
    ClockFrequency float64   // 50 MHz
    FlyCapacitance float64   // 100 pF
    LoadCurrent    float64   // 10 µA
    Efficiency     float64   // 70%
}
```

**Requirements:**
- Input: 1.0 V CMOS supply
- Output: ±1.5 V for FeFET switching
- Efficiency: > 60%
- Rise time: < 100 ns
- Ripple: < 50 mV
- Max current: > 10 µA

---

### 7.5 INL/DNL Analysis

```go
// REQUIRED: Linearity analysis for DAC and ADC
type INLDNLAnalysis struct {
    Levels    int         // Number of levels analyzed
    INLValues []float64   // INL at each code (LSB)
    DNLValues []float64   // DNL at each code (LSB)
    MaxINL    float64     // Worst-case INL
    MaxDNL    float64     // Worst-case DNL
    MinDNL    float64     // Check for missing codes
    WorstCode int         // Code with worst INL
}

func (d *DAC) AnalyzeINLDNL() *INLDNLAnalysis
func (a *ADC) AnalyzeINLDNL() *INLDNLAnalysis
```

**Requirements:**
- Analyze all 30 FeCIM levels
- Report max INL and DNL
- Flag any missing codes (DNL < -1)
- Identify worst-performing code

---

### 7.6 Timing Analysis

```go
// REQUIRED: System timing analysis
type TimingAnalysis struct {
    // Individual component times
    DACSettle     float64   // DAC settling time
    PumpRise      float64   // Charge pump rise time
    WriteTime     float64   // Total write (DAC + pump + pulse)
    TIASettle     float64   // TIA settling time
    ADCConvert    float64   // ADC conversion time
    ReadTime      float64   // Total read (TIA + ADC)

    // System metrics
    CycleTime     float64   // Full read+write cycle
    MaxThroughput float64   // Operations per second
}

func AnalyzeTiming(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TimingAnalysis
```

**Requirements:**
- Track timing through full signal chain
- Calculate total cycle time
- Report maximum throughput

---

### 7.7 Power Analysis

```go
// REQUIRED: Power and energy breakdown
type PowerBreakdown struct {
    // Energy per component (J)
    DACEnergy   float64
    ADCEnergy   float64
    TIAEnergy   float64
    PumpEnergy  float64
    TotalEnergy float64

    // Power (W)
    TotalPower  float64

    // Fractions (%)
    DACFraction  float64
    ADCFraction  float64   // Should be 50-80% (dominant)
    TIAFraction  float64
    PumpFraction float64
}

func AnalyzePower(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump, timing *TimingAnalysis) *PowerBreakdown
```

**Requirements:**
- Track energy per component
- Calculate power fractions
- Identify dominant consumer (ADC)

---

### 7.8 Transfer Function Analysis

```go
// REQUIRED: End-to-end signal chain analysis
type TransferFunction struct {
    InputLevels  []int       // Digital input (0-29)
    DACVoltages  []float64   // After DAC
    PumpVoltages []float64   // After charge pump
    TIAVoltages  []float64   // After TIA (readback)
    ADCLevels    []int       // Final ADC output
    Errors       []int       // Output - Input error
}

func ComputeTransferFunction(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TransferFunction
```

**Requirements:**
- Trace signal through all components
- Calculate end-to-end error
- Verify level-in = level-out (ideally)

---

### 7.9 Visualization Requirements

```go
type Visualizer interface {
    // Transfer functions
    PlotADCTransfer() image.Image
    PlotDACTransfer() image.Image
    PlotINLDNL() image.Image

    // Signal flow
    AnimateWritePath(level int)
    AnimateReadPath(level int)

    // Power
    PlotPowerBreakdown() image.Image
    PlotTimingDiagram() image.Image

    // Interactive
    OnComponentClick(component string, callback func(info ComponentInfo))
}
```

**Requirements:**
- ADC/DAC transfer function plots
- INL/DNL bar charts
- Power pie chart (show ADC dominance)
- Timing diagram (show write vs read)
- Interactive signal flow animation

---

### 7.10 Production Checklist

| Requirement | Description | Priority |
|-------------|-------------|----------|
| ✅ ADC model | 5-bit SAR with INL/DNL | CRITICAL |
| ✅ DAC model | 5-bit bipolar with INL/DNL | CRITICAL |
| ✅ TIA model | Current-to-voltage with noise | CRITICAL |
| ✅ Charge pump | Dickson 2-stage | CRITICAL |
| ✅ INL/DNL analysis | Per-code linearity | HIGH |
| ✅ Timing analysis | Full cycle time | HIGH |
| ✅ Power breakdown | Energy per component | HIGH |
| ✅ Transfer function | End-to-end error | HIGH |
| ⬜ Temperature model | T-dependent specs | MEDIUM |
| ⬜ Process corners | Fast/slow/typical | MEDIUM |
| ⬜ ADC-less option | Time-domain alternative | LOW |
| ⬜ Calibration | Digital INL/DNL correction | LOW |

---

### 7.11 Test Cases

```go
func TestADCQuantization(t *testing.T) {
    adc := DefaultADC()

    // Test 30 levels map correctly
    for level := 0; level < 30; level++ {
        voltage := float64(level) / 29.0  // Normalized
        result := adc.Convert(voltage)

        // Should map back to same level (within ±1)
        assert.InDelta(level, result, 1)
    }
}

func TestDACMonotonicity(t *testing.T) {
    dac := DefaultDAC()

    // Each level should produce higher voltage than previous
    prevV := dac.Convert(0)
    for level := 1; level < 30; level++ {
        currV := dac.Convert(level)
        assert.Greater(currV, prevV, "DAC must be monotonic")
        prevV = currV
    }
}

func TestTransferFunction(t *testing.T) {
    // Full write → read cycle should preserve level
    dac := DefaultDAC()
    adc := DefaultADC()
    tia := DefaultTIA()
    pump := DefaultChargePump()

    tf := ComputeTransferFunction(dac, adc, tia, pump)

    // Most levels should read back correctly
    errors := 0
    for i := 0; i < 30; i++ {
        if tf.Errors[i] != 0 {
            errors++
        }
    }
    assert.Less(errors, 3, "At most 2 levels should have errors")
}

func TestPowerBreakdown(t *testing.T) {
    // ADC should dominate power
    breakdown := AnalyzePower(DefaultDAC(), DefaultADC(),
                             DefaultTIA(), DefaultChargePump(), timing)

    assert.Greater(breakdown.ADCFraction, 0.3,
                  "ADC should be significant power consumer")
}
```

---

## Summary: What Makes a Perfect Peripheral Module

1. **All Four Components:** ADC, DAC, TIA, Charge Pump
2. **Realistic Nonlinearity:** INL/DNL models for ADC/DAC
3. **Noise Models:** TIA thermal noise, charge pump ripple
4. **Timing Analysis:** Full cycle time calculation
5. **Power Breakdown:** Show ADC dominance
6. **Transfer Function:** End-to-end level accuracy
7. **Visualization:** Signal flow, INL/DNL plots, power pie chart

**The goal:** Enable researchers and engineers to understand the complete signal chain from digital input to analog memory and back, with realistic non-idealities and power/timing trade-offs.
