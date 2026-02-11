# Crossbar Arrays Explained Like I'm 5

**Understanding Compute-in-Memory Through Simple Analogies + Production Module Specification**

---

**Note:** References to “30 levels” refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

## Part 1: The Water Park Analogy

### What's a Crossbar Array?

Imagine a **water park with a grid of water slides**. Each slide has a valve that controls how much water can flow through it:

```
         WATER TOWERS (Input)
            │    │    │    │
            5    2    4    8   ← How much water we pour in
            ▼    ▼    ▼    ▼
         ═══╬════╬════╬════╬═══→ Pool 1 collects: 23 gallons
            ║    ║    ║    ║
         ═══╬════╬════╬════╬═══→ Pool 2 collects: 17 gallons
            ║    ║    ║    ║
         ═══╬════╬════╬════╬═══→ Pool 3 collects: 31 gallons

            ╬ = One water slide with adjustable valve
```

**How it works:**
1. Pour water into the towers at the top (these are your INPUT numbers)
2. Each slide valve is set to a different openness (these are WEIGHT numbers)
3. Water × Valve = How much flows through that slide
4. Pools at the end collect all the water from their row
5. The pool totals ARE your answer!

**The magic:** ALL slides flow at the same time! You do ALL the math INSTANTLY!

---

### Why Is This Amazing for Computers?

**Normal computer** = One calculator doing problems one at a time:
```
5 × 3 = 15   ✓ Done, next...
2 × 7 = 14   ✓ Done, next...
4 × 6 = 24   ✓ Done, next...
(sooooo slow when you have millions!)
```

**Crossbar computer** = Millions of calculators working together:
```
5 × 3 = 15  ─┐
2 × 7 = 14   │
4 × 6 = 24   ├── ALL DONE AT ONCE!
8 × 2 = 16   │
...        ─┘
```

---

## Part 2: Real Crossbars Use Electricity

### From Water to Electricity

| Water Park | Real Crossbar |
|------------|---------------|
| Water pressure | Voltage (how hard we push) |
| Valve openness | Conductance (how easily current flows) |
| Water flow | Current (what comes out) |

**The physics equation is simple:**
```
Current = Voltage × Conductance

I = V × G

This is Ohm's Law - the universe does this automatically!
```

---

### 30 Levels = More Precise Valves

Remember from Demo 1, each cell can hold **30 different values**?

In our water park:
- Level 0 = Valve almost closed (tiny drip)
- Level 15 = Valve half open (medium flow)
- Level 29 = Valve wide open (maximum flow)

```
Level 0:   │░│  → barely any flow      (weight ≈ 0)
Level 10:  │▒│  → some flow            (weight ≈ 0.33)
Level 20:  │▓│  → good flow            (weight ≈ 0.67)
Level 29:  │█│  → maximum flow         (weight = 1.0)
```

**Why 30 levels matters:** More precise "valve settings" = more accurate AI!

---

## Part 3: Problems in the Real World

Real water parks aren't perfect, and neither are real crossbars!

### Problem 1: Pressure Drop (IR Drop)

Like water pressure dropping as you get farther from the pump:

```
PUMP HERE
    ↓
    █ → █ → █ → █ → █
    │    │    │    │
   100%  98%  95%  90%  85%  ← Less pressure at the end!
```

**In the chip:** Cells far from the voltage source see lower voltage.

**How bad is it?**
| Array Size | Pressure Loss at Corner | Accuracy Impact |
|------------|------------------------|-----------------|
| 64×64 | ~5% | Small (~0.3%) |
| 128×128 | ~12% | Noticeable (~1.2%) |
| 256×256 | ~22% | Significant (~3.5%) |

**Solutions:**
- Make the pipes (wires) thicker
- Add pressure boosters along the way
- Use smaller arrays and tile them together

---

### Problem 2: Sneak Paths (Water Taking Shortcuts)

Like water finding sneaky shortcuts through pipes you didn't want:

```
    Wanted path:        Sneaky path:
        ↓                   ↓
      ┌───┐               ┌───┐
      │ █ │               │ █ │←──┐
      └───┘               └───┘   │
        ↓                   ↓     │ Oops!
      ┌───┐               ┌───┐   │
      │ █ │               │ █ │───┘
      └───┘               └───┘
```

**In the chip:** Current flows through cells you didn't select, adding noise to your answer.

**How bad is it?**
| Array Type | Sneak Error | Fix |
|------------|-------------|-----|
| Passive (no transistors) | 5-20% | Add one-way valves |
| With transistors (1T1R) | ~0% | Problem solved! |
| With selectors (1S1R) | 0.5-2% | Good compromise |

**FeFET Advantage:** Ferroelectric cells can be "self-rectifying" (built-in one-way valve).

---

### Problem 3: Every Valve is Slightly Different

Like pipes that are slightly different from each other, even when they should be the same:

```
    What we wanted:     What we got:
      │  │  │            │  │  │
      █  █  █            █  ▓  █   ← Middle one is
      │  │  │            │  │  │      slightly different!
```

**In the chip:** Manufacturing isn't perfect, so each cell varies by ~5-10%.

**Solutions:**
- Program carefully with "write-verify" (program, check, adjust)
- Train AI to be robust to noise
- Use error-tolerant algorithms

---

### Problem 4: The Output Measuring Tool (ADC) Has Limited Precision

The tool that reads how much water came out can only give approximate readings:

```
Actual output: 0.7328451...

Our measuring cup only shows: 0, 0.25, 0.50, 0.75, 1.00

We read: 0.75  (rounded)
              ↑
         Lost precision!
```

**The ADC Trade-off:**
| ADC Bits | Precision | Power Cost |
|----------|-----------|------------|
| 4-bit | 16 levels | 1× (baseline) |
| 6-bit | 64 levels | 4× |
| 8-bit | 256 levels | 16× |

**Fun fact:** The ADC uses **50-80% of all the power** in a CIM chip! That's why "ADC-less" designs are a hot research area.

---

## Part 4: Why Ferroelectric Crossbars Win

| Memory Type | Good Things | Bad Things |
|-------------|-------------|------------|
| RRAM | Simple, small | Gets hot, sneak paths |
| PCM | Stores lots | VERY hot when switching |
| MRAM | Fast, durable | Complex, big |
| **FeFET** | Low power, fits with normal chips | Newer technology |

**FeCIM uses ferroelectric because:**
- ✅ Barely any heat (displacement current, not burning)
- ✅ Can be self-rectifying (built-in sneak path fix)
- ✅ Works in normal chip factories
- ⚠️ 30 analog states (demo baseline; very precise)
- ✅ Lasts 10¹² cycles (basically forever)

---

## Part 5: The One Sentence Summary

> **A crossbar array is a grid where physics does all the multiplications at once, making AI 100× more energy efficient than regular computers.**

---

# Part 6: Perfect Crossbar Module Specification for FeCIM

## What the Production Module MUST Do

Based on research from 40+ papers and our implementation analysis, here's the specification for a production-ready crossbar simulation module:

---

### 6.1 Core MVM Operation

```go
// REQUIRED: Physics-accurate Matrix-Vector Multiplication
// y[i] = Σ_j (G[i][j] × V[j])

type CrossbarArray interface {
    // Core operations
    MVM(input []float64) ([]float64, error)   // Forward: y = G × x
    VMM(input []float64) ([]float64, error)   // Transpose: y = x × G

    // Programming
    ProgramWeight(row, col int, weight float64) error
    ProgramMatrix(weights [][]float64) error

    // Readback
    GetConductance(row, col int) float64
    GetConductanceMatrix() [][]float64
}
```

**Requirements:**
- Support arrays from 8×8 to 1024×1024
- Quantize all weights to exactly 30 discrete levels (demo baseline; simulation baseline)
- Apply Ohm's Law: I = G × V at each cell
- Apply Kirchhoff's Current Law: Sum currents on each row

---

### 6.2 Non-Ideality Models

#### IR Drop Model

```go
type IRDropSimulator interface {
    // Solve resistive network for actual node voltages
    Simulate(appliedVoltages []float64) *IRDropResult

    // Parameters
    SetWireResistance(rowR, colR float64)  // Ω per segment
    SetContactResistance(R float64)         // Ω per cell
}

type IRDropResult struct {
    ActualVoltages [][]float64  // Voltage at each cell
    VoltageLoss    [][]float64  // Loss from ideal
    MaxLossPercent float64      // Worst case
}
```

**Requirements:**
- Model wire resistance: ~2.5 Ω per cell pitch (45nm node)
- Model contact resistance: ~50 Ω
- Use iterative relaxation or direct matrix solve
- Report voltage loss at each cell location

#### Sneak Path Model

```go
type SneakPathAnalyzer interface {
    // Analyze sneak paths for a target cell
    Analyze(targetRow, targetCol int) *SneakPathResult

    // Configure mitigation
    EnableSelector(onOffRatio float64)
    EnableHalfSelect(voltage float64)
}

type SneakPathResult struct {
    TargetCurrent  float64     // Desired current
    SneakCurrent   float64     // Parasitic current
    SneakRatio     float64     // Sneak/Target ratio
    WorstPaths     []SneakPath // Top contributing paths
}
```

**Requirements:**
- Implement three-cell sneak path model minimum
- Support passive (0T1R), selector (1S1R), and active (1T1R) modes
- Identify and enumerate top sneak current contributors
- Calculate error contribution to MVM output

#### Device Variation Model

```go
type VariationModel interface {
    // Apply variation to conductance
    ApplyVariation(nominal float64) float64

    // Configuration
    SetD2DVariation(sigma float64)  // Device-to-device %
    SetC2CVariation(sigma float64)  // Cycle-to-cycle %
}
```

**Requirements:**
- Gaussian distribution with configurable σ
- Separate device-to-device (fixed per cell) and cycle-to-cycle (per operation)
- Default: 5% D2D, 3% C2C for FeFET

#### Drift Model

```go
type DriftSimulator interface {
    // Simulate time evolution
    SimulateTime(dt float64)

    // Check retention
    GetRetention(elapsedTime float64) float64
}
```

**Requirements:**
- Model: G(t) = G₀ × (t/t₀)^ν
- FeFET drift: Assumed ν ≈ 0.001 for simulation (no reported in literature source; qualitatively 'mild')
- Temperature-aware retention projection

---

### 6.3 ADC/DAC Quantization

```go
type Peripherals interface {
    // DAC: Convert digital input to analog voltage
    QuantizeDAC(digital float64) float64

    // ADC: Convert analog current to digital output
    QuantizeADC(analog float64) float64

    // Configuration
    SetADCBits(bits int)  // Default: 6
    SetDACBits(bits int)  // Default: 6
}
```

**Requirements:**
- Uniform quantization to 2^bits levels
- Clamp to valid range [0, 1] before quantization
- Support 4-bit to 10-bit precision

---

### 6.4 30-Level Quantization

```go
const FeCIMLevels = 30

// REQUIRED: All weights must be quantized to exactly 30 levels
func QuantizeTo30Levels(value float64) float64 {
    value = clamp(value, 0, 1)
    level := round(value * 29)  // 0-29
    return level / 29           // Back to normalized
}

func GetDiscreteLevel(conductance float64) int {
    return int(round(conductance * 29))
}
```

**Requirements:**
- Linear mapping from [0, 1] to {0, 1, 2, ..., 29}
- Conductance range: 1 µS to 100 µS (linear)
- Consistent with the demo's 30‑level baseline (simulation baseline)

---

### 6.5 Write-Verify Programming

```go
type WriteVerifyProgrammer interface {
    // Program with verification
    Program(row, col int, targetLevel int) error

    // Configuration
    SetMaxIterations(n int)
    SetTolerance(t float64)
}
```

**Requirements:**
- Iterative: write pulse → read → adjust → repeat
- Configurable tolerance (default: ±0.5 levels)
- Track number of iterations for statistics
- Support incremental pulse programming (Scheme C)

---

### 6.6 Performance Metrics

```go
type Metrics struct {
    // Accuracy
    IdealOutput    []float64
    ActualOutput   []float64
    MSE            float64
    MaxError       float64

    // Energy
    ArrayEnergy    float64  // MVM computation (pJ)
    ADCEnergy      float64  // ADC conversion (pJ)
    DACEnergy      float64  // DAC conversion (pJ)
    TotalEnergy    float64

    // Statistics
    TotalReads     int64
    TotalWrites    int64
    CyclesSinceRefresh int64
}
```

---

### 6.7 Visualization Requirements

```go
type Visualizer interface {
    // Heatmaps
    RenderConductanceHeatmap() image.Image
    RenderIRDropHeatmap() image.Image
    RenderSneakPathHeatmap() image.Image

    // Animation
    AnimateMVM(input []float64, fps int)

    // Interactive
    OnCellClick(row, col int, callback func(CellInfo))
}
```

**Requirements:**
- Real-time update at 30+ FPS
- Color scale: viridis or similar perceptually uniform
- Show discrete levels (0-29) on hover
- Highlight selected cell and its sneak paths

---

### 6.8 Configuration Interface

```go
type Config struct {
    // Array dimensions
    Rows int  // 8 to 1024
    Cols int  // 8 to 1024

    // Quantization
    ConductanceLevels int     // Default: 30
    ADCBits          int      // Default: 6
    DACBits          int      // Default: 6

    // Non-idealities (all can be disabled)
    EnableIRDrop     bool
    EnableSneakPaths bool
    EnableVariation  bool
    EnableDrift      bool

    // Device parameters
    GMin           float64  // 1e-6 S
    GMax           float64  // 100e-6 S
    WireResistance float64  // 2.5 Ω
    D2DVariation   float64  // 0.05 (5%)
    DriftCoeff     float64  // 0.001

    // Architecture
    CellType       string   // "0T1R", "1T1R", "1S1R"
    SelectorOnOff  float64  // On/off ratio for 1S1R
}
```

---

### 6.9 Test Cases for Validation

```go
func TestMVMAccuracy(t *testing.T) {
    // Create array with known weights
    weights := [][]float64{
        {1, 0, 0.5},
        {0, 1, 0.5},
        {0.5, 0.5, 1},
    }
    arr := NewArray(3, 3)
    arr.ProgramMatrix(weights)

    // Test MVM (no non-idealities)
    input := []float64{1, 1, 1}
    expected := []float64{1.5, 1.5, 2.0}  // Matrix multiply
    output, _ := arr.MVM(input)

    assert.InDelta(expected, output, 0.01)
}

func TestIRDropScaling(t *testing.T) {
    // IR drop should increase with array size
    sizes := []int{32, 64, 128, 256}
    maxDrops := []float64{}

    for _, size := range sizes {
        arr := NewArrayWithIRDrop(size, size)
        result := arr.SimulateIRDrop(uniformVoltages)
        maxDrops = append(maxDrops, result.MaxLossPercent)
    }

    // Each doubling should roughly double max drop
    for i := 1; i < len(maxDrops); i++ {
        ratio := maxDrops[i] / maxDrops[i-1]
        assert.InDelta(2.0, ratio, 0.5)
    }
}

func Test30LevelQuantization(t *testing.T) {
    // All conductances should snap to exactly 30 levels
    for _, input := range []float64{0.0, 0.017, 0.5, 0.983, 1.0} {
        quantized := QuantizeTo30Levels(input)
        level := GetDiscreteLevel(quantized)
        assert.True(level >= 0 && level <= 29)
        assert.Equal(quantized, float64(level)/29)
    }
}
```

---

### 6.10 Production Checklist

| Requirement | Description | Priority |
|-------------|-------------|----------|
| ✅ MVM/VMM | Matrix operations with Ohm's Law | CRITICAL |
| ✅ 30 levels | Demo baseline (simulation baseline (configurable) | CRITICAL |
| ✅ IR drop | Resistive network solver | HIGH |
| ✅ Sneak paths | Three-cell model minimum | HIGH |
| ✅ Variation | 5% D2D Gaussian | HIGH |
| ✅ ADC/DAC | 6-bit default quantization | HIGH |
| ⬜ Drift | Time-domain conductance evolution | MEDIUM |
| ⬜ Write-verify | Iterative programming | MEDIUM |
| ⬜ Temperature | T-aware IR drop and drift | LOW |
| ⬜ Differential | 2T2R for signed weights | LOW |

---

## Summary: What Makes a Perfect Crossbar Module

1. **Physics Accuracy:** Ohm's Law (I=GV) and Kirchhoff's Law (currents sum)
2. **30-Level Quantization:** Demo baseline (simulation baseline; reported multi-level range)
3. **Non-Idealities:** IR drop, sneak paths, variation, drift (all toggleable)
4. **Peripheral Models:** Realistic ADC/DAC quantization
5. **Visualization:** Real-time heatmaps and MVM animation
6. **Configurability:** Array size, architecture, all parameters adjustable
7. **Validation:** Test suite against analytical and published results

**The goal:** Enable researchers and engineers to understand, design, and optimize ferroelectric CIM systems through accurate, visual, and interactive simulation.

---

## Related Documentation

- **[Crossbar Physics](../educational/crossbar.physics.md)** - Deep technical reference on crossbar physics
- **[Demo Guide](../educational/crossbar.demo.md)** - How to run the interactive visualization
- **[Research Papers](../educational/crossbar.research.md)** - Academic citations and meta-study
- **[Open Source Tools](./crossbar.opensource.md)** - Comparison with other simulators

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
**Source:** Dr. external research group's HfO₂-ZrO₂ superlattice research (COSM 2025)
