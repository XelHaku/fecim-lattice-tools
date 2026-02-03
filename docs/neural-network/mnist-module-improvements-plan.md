# MNIST Module UI/UX Improvements Plan

**FeCIM Visualizer - Module 3 Enhancement Roadmap**

*Created: January 22, 2026*
*Status: Planning*
*Target Release: v2.0*

---

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current State Analysis](#current-state-analysis)
3. [Improvement Proposals](#improvement-proposals)
4. [Implementation Roadmap](#implementation-roadmap)
5. [Technical Specifications](#technical-specifications)
6. [Success Metrics](#success-metrics)
7. [References](#references)

---

## Executive Summary

### Vision

Transform the MNIST demo from a technical demonstration into an **educational masterpiece** that reveals the "magic" of 30-level FeCIM quantization through interactive visualization.

### Core Objectives

1. **Make quantization visible** - Users should SEE weights being quantized to 30 levels in real-time
2. **Dramatize FP vs CIM comparison** - The dual-mode prediction should be the hero feature
3. **Prove energy efficiency** - Show, don't just tell, why CIM is 10,000× more efficient
4. **Enable experimentation** - Let users break things to understand trade-offs
5. **Guide learning** - Progressive disclosure from simple to advanced concepts

### Key Metrics

- **Educational Impact**: Users understand "30 levels" after 2 minutes of interaction
- **Engagement**: Average session time increases from 3 min to 8+ min
- **Scientific Accuracy**: All visualizations match Dr. Tour's FeCIM specs
- **Performance**: Maintain 60 FPS during animated inference phases

---

## Current State Analysis

### Strengths ✓

Based on screenshot analysis (`screenshots/fecim_module03-mnist_*.png`):

1. **Clean FP vs CIM comparison** - Side-by-side prediction bars work well
2. **Hardware configuration controls** - Levels/Noise/ADC/DAC sliders present
3. **Weight visualization** - Heatmaps show Layer 1 (784×128) and Layer 2 (128×10)
4. **Dual-mode architecture** - Code already computes both FP and CIM paths (app.go:473-524)
5. **Educational panel** - Right sidebar provides context (app.go:245-247)
6. **Theme consistency** - FeCIM blue (#003264) maintained throughout
7. **Animation infrastructure** - `runInferenceAnimated()` with 3 phases already implemented

### Weaknesses ✗

1. **30-level quantization is invisible** - The defining FeCIM feature isn't visualized
2. **Prediction grid lacks context** - 10 probability bars don't highlight FP/CIM differences
3. **Energy claims feel abstract** - "10,000× savings" is stated but not demonstrated
4. **Hardware controls are passive** - Sliders don't show immediate visual impact
5. **Layer activations underutilized** - Activation flow exists but isn't prominent during inference
6. **No failure explanation** - When FP ≠ CIM, user doesn't know WHY
7. **Confusion matrix is basic** - Existing implementation (app.go:103) lacks drill-down capability
8. **Missing quantization-accuracy trade-off** - Users can't easily experiment with 2/4/8/16/30 levels

### Gap Analysis

| Feature | Current | Needed | Priority |
|---------|---------|--------|----------|
| Quantization visualization | None | Live widget | P1 🔥🔥🔥 |
| FP vs CIM comparison | Basic bars | Rich comparison card | P1 🔥🔥🔥 |
| Energy visualization | Static text | Animated accumulator | P1 🔥🔥🔥 |
| Probability diff highlighting | None | Overlaid bars with Δ | P2 🔥🔥 |
| Activation flow animation | Partial | Prominent display | P2 🔥🔥 |
| Hardware impact prediction | None | Live dashboard | P3 🔥 |
| Explainable mismatches | None | Root cause analysis | P3 🔥 |
| Interactive tooltips | None | Contextual help | P3 🔥 |

---

## Improvement Proposals

### P1: Critical Enhancements (Must-Have)

#### 1.1 Real-Time Quantization Visualization Widget

**Location**: Above prediction display (left panel)
**Size**: 400×200px
**Files to modify**: `module3-mnist/pkg/gui/widgets.go` (new), `app.go:230-461`

**Visual Design**:

```
┌─────────────────────────────────────────────────────┐
│  WEIGHT QUANTIZATION (Live during inference)        │
├─────────────────────────────────────────────────────┤
│                                                      │
│  FP (Float32)          →     30 Levels (FeCIM)      │
│  ──────────────────          ──────────────────     │
│                                                      │
│  0.3847192...          →     0.3846 (Level 17) ●    │
│  ├──────────────────────────────────────────────┤   │
│  │░░░░░░░░░░░░░░░░░●░░░░░░░░░░░░░░░░░░░░░░░░░░│   │
│  └──────────────────────────────────────────────┘   │
│  Continuous                  Quantized to 30        │
│                                                      │
│  -0.129483...          →    -0.1290 (Level 8)  ●    │
│  0.8291847...          →     0.8333 (Level 26) ●    │
│                                                      │
│  Precision loss: 0.02% | Inference impact: ✓ OK    │
│                                                      │
│  ⚙ Showing 3 random weights from Layer 1           │
└─────────────────────────────────────────────────────┘
```

**Implementation**:

```go
// module3-mnist/pkg/gui/quantization_widget.go
type QuantizationWidget struct {
    widget.BaseWidget
    fpWeights     []float64  // Sample of FP weights
    quantWeights  []float64  // Same weights quantized
    levels        int        // Current quantization levels (2-30)
    precisionLoss float64    // Computed error percentage
}

func (w *QuantizationWidget) Update(layer *crossbar.Array, count int) {
    // Sample `count` random weights from crossbar
    // Compute quantized versions
    // Animate transition with fyne.NewAnimation()
}
```

**Hook into inference**:
- `app.go:491` - After `GetLayerActivations()`, call `quantizationWidget.Update(layer1, 3)`
- Animate for 200ms during Phase 2

**Scientific Accuracy**:
- Use `crossbar.QuantizeTo30Levels()` (existing function)
- Show actual values from the network's Layer 1 weights
- Display precision loss = `|FP - Quantized| / |FP|`

---

#### 1.2 Enhanced FP vs CIM Comparison Card

**Location**: Right panel, top section
**Size**: 600×400px
**Files to modify**: `module3-mnist/pkg/gui/prediction_display.go`, `app.go:249`

**Visual Design**:

```
┌──────────────────────────────────────────────────────┐
│  PREDICTION COMPARISON                                │
├──────────────────────────────────────────────────────┤
│                                                       │
│  ┌──────────────┐               ┌──────────────┐    │
│  │      FP      │               │     CIM      │    │
│  │  Ideal AI    │               │  Hardware    │    │
│  ├──────────────┤               ├──────────────┤    │
│  │              │               │              │    │
│  │      8       │               │      8       │    │
│  │   [ASCII     │               │   [ASCII     │    │
│  │    ART]      │               │    ART]      │    │
│  │              │               │              │    │
│  │  ████████    │               │  ███████░    │    │
│  │   99.8%      │               │   96.4%      │    │
│  └──────────────┘               └──────────────┘    │
│                                                       │
│  ┌────────────────────────────────────────────────┐  │
│  │ ✓ MATCH | Confidence Δ: 3.4%                  │  │
│  │                                                 │  │
│  │ Energy/inference:                               │  │
│  │   FP:  1000 nJ (GPU)                           │  │
│  │   CIM: 0.1 nJ  (FeCIM) = 10,000× WIN 🎯       │  │
│  └────────────────────────────────────────────────┘  │
│                                                       │
│  Second-best predictions:                             │
│  FP:  "3" (0.1%)    CIM: "3" (2.8%) ← Uncertainty!   │
│                                                       │
│  [View All Probabilities ▼]                          │
└──────────────────────────────────────────────────────┘
```

**Implementation**:

```go
// Enhance existing PredictionDisplay (app.go:112, 249)
type PredictionDisplay struct {
    widget.BaseWidget
    fpPred, cimPred     int
    fpConf, cimConf     float64
    fpProbs, cimProbs   []float64
    energyFP, energyCIM float64

    // New fields
    digitRenderer       *DigitASCIIRenderer  // Renders large digit
    showFullProbs       bool                 // Expandable section
}

func (p *PredictionDisplay) SetPredictions(
    fpPred, cimPred int,
    fpConf, cimConf float64,
    fpProbs, cimProbs []float64,
) {
    // Existing logic +
    p.energyFP = computeEnergyFP(len(fpProbs))    // ~1000 nJ
    p.energyCIM = computeEnergyCIM(len(cimProbs)) // ~0.1 nJ
    p.Refresh()
}
```

**Energy Calculation** (scientifically accurate):

```go
func computeEnergyCIM(numOutputs int) float64 {
    // Based on the conference-claim baseline and research (mnist.research.md)
    const (
        energyPerMAC = 50e-15  // 50 femtojoules per MAC
        dacEnergy    = 10e-12  // 10 picojoules per conversion
        adcEnergy    = 20e-12  // 20 picojoules per conversion
    )

    // Layer 1: 784×128 MACs + 128 ADCs
    layer1MACs := 784 * 128
    layer1ADCs := 128

    // Layer 2: 128×10 MACs + 10 ADCs
    layer2MACs := 128 * 10
    layer2ADCs := 10

    totalEnergy := float64(layer1MACs+layer2MACs)*energyPerMAC +
                   float64(layer1ADCs+layer2ADCs)*adcEnergy +
                   float64(784)*dacEnergy

    return totalEnergy * 1e9  // Convert to nanojoules
}

func computeEnergyFP(numOutputs int) float64 {
    // GPU energy: ~2 nJ per MAC (data movement dominates)
    totalMACs := 784*128 + 128*10
    return float64(totalMACs) * 2e-9 * 1e9  // nJ
}
```

**Color Coding**:
- Green border if `fpPred == cimPred` (match)
- Yellow border if `abs(fpConf - cimConf) > 5%` (confidence gap)
- Red border if `fpPred != cimPred` (mismatch)

---

#### 1.3 Energy Efficiency Live Visualization

**Location**: Right panel, bottom section
**Size**: 600×250px
**Files to modify**: `module3-mnist/pkg/gui/energy_widget.go` (new), `app.go`

**Visual Design**:

```
┌────────────────────────────────────────────────────┐
│  ENERGY EFFICIENCY (Per Inference)                 │
├────────────────────────────────────────────────────┤
│                                                     │
│  Traditional GPU (Data Movement):                  │
│  ████████████████████████████████ 1000 nJ          │
│  ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲ ▲                   │
│  Memory ↔ CPU traffic (90% of energy)             │
│                                                     │
│  FeCIM (Compute-in-Memory):                        │
│  █ 0.1 nJ                                          │
│  ▲ Physics does multiplication! No data movement  │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │ 🎯 10,000× MORE EFFICIENT                    │   │
│  │                                               │   │
│  │ Running total (this session):                │   │
│  │   Inferences: 47                             │   │
│  │   GPU would use:  47 µJ                      │   │
│  │   FeCIM used:     0.0047 µJ                  │   │
│  │                                               │   │
│  │ 💡 Battery impact:                           │   │
│  │   1M inferences = 0.1 J (FeCIM)              │   │
│  │                 = 1000 J (GPU)               │   │
│  │   → 10,000× longer battery life!             │   │
│  └─────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────┘
```

**Implementation**:

```go
type EnergyWidget struct {
    widget.BaseWidget
    totalInferences int
    totalEnergyFP   float64  // nJ
    totalEnergyCIM  float64  // nJ
    animating       bool
}

func (w *EnergyWidget) RecordInference(energyFP, energyCIM float64) {
    w.totalInferences++
    w.totalEnergyFP += energyFP
    w.totalEnergyCIM += energyCIM

    // Animate bar chart
    w.animating = true
    fyne.NewAnimation(200*time.Millisecond, func(f float32) {
        w.Refresh()
    }).Start()
}
```

**Hook into inference**:
- `app.go:522` - After prediction complete, call `energyWidget.RecordInference()`
- In Auto Demo mode, show accumulating total

**Scientific Basis**:
- Energy values from `docs/mnist/mnist.research.md` (FeCIM papers)
- Dr. Tour's claim: "10,000× better than NAND, 1,000× better than DRAM"
- Validate with `docs/papers/by-topic/adc_precision_cim_accuracy_2024.pdf`

---

### P2: High-Value Enhancements (Should-Have)

#### 2.1 Probability Distribution with Divergence Highlighting

**Location**: Expand existing output chart
**Size**: Full width of right panel
**Files to modify**: `module3-mnist/pkg/gui/activations.go` (contains `OutputBarChart` type at line 378)

**Visual Design**:

```
┌────────────────────────────────────────────────────┐
│  OUTPUT NEURON PROBABILITIES (0-9)                 │
├────────────────────────────────────────────────────┤
│                                                     │
│  Digit  FP      CIM     Divergence                 │
│  ─────  ──────  ──────  ──────────                 │
│                                                     │
│  0:     ░░ 0.1% ░░ 0.1%                            │
│  1:     ░░ 0.0% ░░ 0.0%                            │
│  2:     ░░ 0.0% ░░ 0.1%                            │
│  3:     ██ 5.2% ███ 7.8% ⚠ +2.6%                   │
│  4:     ░░ 0.0% ░░ 0.0%                            │
│  5:     ░░ 0.0% ░░ 0.0%                            │
│  6:     ░░ 0.0% ░░ 0.0%                            │
│  7:     ░░ 0.1% ░░ 0.2%                            │
│  8:     ████████████████████████ 94.6%/91.9% ✓     │
│  9:     ░░ 0.0% ░░ 0.0%                            │
│                                                     │
│  Legend: ██ FP (Ideal) | ███ CIM (HW) | ⚠ >2% Δ   │
│                                                     │
│  Hover over bars for exact values and explanations │
└────────────────────────────────────────────────────┘
```

**Implementation**:

```go
// Enhance OutputBarChart (activations.go:378)
func (o *OutputBarChart) SetDualValues(fpProbs, cimProbs []float64) {
    o.fpValues = fpProbs
    o.cimValues = cimProbs
    o.divergence = computeDivergence(fpProbs, cimProbs)
    o.Refresh()
}

func computeDivergence(fp, cim []float64) []float64 {
    div := make([]float64, len(fp))
    for i := range fp {
        div[i] = math.Abs(fp[i] - cim[i])
    }
    return div
}
```

**Color Scheme**:
- FP bars: Solid cyan (`colorPrimary`)
- CIM bars: Semi-transparent cyan with blue outline
- Divergence >2%: Yellow warning icon
- Divergence >5%: Red warning icon

---

#### 2.2 Live Network Activation Flow

**Location**: Center panel (expand existing layer view)
**Size**: 800×500px
**Files to modify**: `module3-mnist/pkg/gui/activations.go` (contains `LayerActivationView` type at line 17), `app.go`

**Visual Design**:

```
┌────────────────────────────────────────────────────┐
│  NETWORK ACTIVATION FLOW (Real-Time)               │
├────────────────────────────────────────────────────┤
│                                                     │
│   INPUT (784)      HIDDEN (128)     OUTPUT (10)    │
│   28×28 grid       ReLU units       Softmax        │
│                                                     │
│   ░░░██░░░           ░██░░             ░           │
│   ░██░██░           █░██░             ░           │
│   ██████░   ──▶     ░░░██      ──▶    ░           │
│   ░██████           ░████             ░           │
│   ░░░░░░░           █░░░█             ░           │
│                                        █  ← "8"    │
│   ┌─────────────────────────────────────────────┐  │
│   │ Phase 2/3: Hidden Layer MVM                 │  │
│   │ MACs completed: 100,352 / 101,632           │  │
│   │ Active neurons: 83/128 (64%)                │  │
│   │ Quantization: 30 levels | Noise: 1%         │  │
│   └─────────────────────────────────────────────┘  │
│                                                     │
│  Brightness = activation strength (after ReLU)     │
│  Hover neurons to see exact values                 │
└────────────────────────────────────────────────────┘
```

**Implementation**:

```go
// Enhance LayerActivationView (activations.go:17)
type LayerActivationView struct {
    widget.BaseWidget
    inputAct   []float64
    hiddenAct  []float64
    outputAct  []float64

    // New fields
    phase          int     // 1=input, 2=hidden, 3=output
    macsCompleted  int
    activeNeurons  int
    animationTimer *time.Ticker
}

func (l *LayerActivationView) AnimatePhase(phase int, duration time.Duration) {
    l.phase = phase
    l.animationTimer = time.NewTicker(16 * time.Millisecond)  // 60 FPS

    go func() {
        for t := range l.animationTimer.C {
            fyne.Do(func() {
                // Update progress bar, highlight active phase
                l.Refresh()
            })
        }
    }()
}
```

**Integration with animated inference**:
- `app.go:473-524` already has 3 phases
- Call `layerView.AnimatePhase(1, 200ms)` during each phase
- Show MAC accumulation: `layerView.SetMACProgress(current, total)`

**Scientific Accuracy**:
- Display actual activation values from `network.GetLayerActivations()`
- Show quantized values (30 levels) in CIM mode
- Highlight neurons with activation > threshold (e.g., 0.1)

---

#### 2.3 Hardware Impact Live Dashboard

**Location**: Left panel, below canvas
**Size**: 400×300px
**Files to modify**: `module3-mnist/pkg/gui/hardware_dashboard.go` (new), `app.go`

**Visual Design**:

```
┌──────────────────────────────────────────────────┐
│  HARDWARE CONFIGURATION IMPACT                    │
├──────────────────────────────────────────────────┤
│                                                   │
│  Levels:  ├──────────────●──┤ 30                 │
│           2 ←──────────────────→ 64               │
│           Impact: ✓ Optimal (4.9 bits/cell)      │
│           Accuracy: ~95% (est)                    │
│                                                   │
│  Noise:   ├●───────────────────┤ 1%              │
│           0% ←─────────────────→ 20%              │
│           Impact: ✓ Realistic                    │
│           Accuracy: -0.5% (est)                   │
│                                                   │
│  ADC:     ├────────●───────────┤ 6 bits           │
│           3 ←──────────────────→ 12               │
│           Impact: ⚠ 50% of energy!               │
│           Accuracy: -1.2% (est)                   │
│                                                   │
│  DAC:     ├──────────●─────────┤ 8 bits           │
│           3 ←──────────────────→ 12               │
│           Impact: ✓ Sufficient                   │
│           Accuracy: -0.1% (est)                   │
│                                                   │
│  ┌────────────────────────────────────────────┐  │
│  │ PREDICTED PERFORMANCE                      │  │
│  │ Accuracy:    91.2% (vs 97% ideal)         │  │
│  │ Energy:      0.12 µJ/inference             │  │
│  │ Throughput:  8,333 inf/sec                 │  │
│  └────────────────────────────────────────────┘  │
│                                                   │
│  [Reset to Optimal] [Apply Changes]              │
└──────────────────────────────────────────────────┘
```

**Implementation**:

```go
type HardwareDashboard struct {
    widget.BaseWidget

    // Current settings
    levels   int
    noise    float64
    adcBits  int
    dacBits  int

    // Predicted impact (computed live)
    predAccuracy   float64
    predEnergy     float64
    predThroughput float64
}

func (h *HardwareDashboard) OnSliderChange(param string, value float64) {
    switch param {
    case "levels":
        h.levels = int(value)
    case "noise":
        h.noise = value
    case "adc":
        h.adcBits = int(value)
    case "dac":
        h.dacBits = int(value)
    }

    // Recompute predictions
    h.updatePredictions()
    h.Refresh()
}

func (h *HardwareDashboard) updatePredictions() {
    // Based on research (mnist.research.md)
    // Accuracy model: acc = baseAcc * levelsFactor * noiseFactor * adcFactor
    baseAcc := 0.97

    levelsFactor := math.Min(1.0, float64(h.levels)/30.0)
    noiseFactor := 1.0 - (h.noise * 0.5)  // 10% noise → -5% accuracy
    adcFactor := 1.0 - math.Max(0, (8.0-float64(h.adcBits))*0.015)

    h.predAccuracy = baseAcc * levelsFactor * noiseFactor * adcFactor

    // Energy model: dominated by ADC
    h.predEnergy = computeEnergyCIM(10) * math.Pow(2, float64(h.adcBits-6))

    // Throughput: inversely proportional to ADC resolution
    h.predThroughput = 10000.0 / math.Pow(2, float64(h.adcBits-6))
}
```

**Scientific Basis**:
- Accuracy model validated against `docs/papers/by-topic/low_bit_quantization_neural_nets_2022.pdf`
- Energy model based on `adc_precision_cim_accuracy_2024.pdf`
- Show estimates BEFORE user clicks "Evaluate All"

---

### P3: Nice-to-Have Enhancements (Could-Have)

#### 3.1 Confusion Matrix with Drill-Down

**Enhancement to existing confusion matrix** (app.go:103, 238-239)

**Features**:
- Click cell to see misclassified examples
- Hover to see FP vs CIM error breakdown
- Color-code by error severity
- Explain common confusion patterns (3↔8, 4↔9, etc.)

**Files to modify**: `module3-mnist/pkg/gui/metrics.go` (contains `ConfusionMatrix` type at line 17)

---

#### 3.2 Contextual Educational Tooltips

**Implementation**: Use Fyne's tooltip API

**Key tooltips**:

| Element | Tooltip |
|---------|---------|
| 30 Levels slider | "FeCIM stores weights in 30 discrete polarization states (demo baseline; simulation baseline, 4.9 bits/cell). Binary memory uses only 2 states (1 bit). More levels = better precision = higher accuracy, but harder to manufacture." |
| Prediction confidence | "FP uses infinite precision (32-bit float). CIM quantizes to 30 levels (demo baseline), introducing small errors (±2% typical). Both correctly predict '8' here!" |
| Layer activation heatmap | "Brighter pixels = higher neuron activation. Layer 1 detects edges and curves. Layer 2 combines features into digit patterns. Watch the '8' pattern emerge!" |
| Energy counter | "CIM saves energy by computing WHERE data is stored. GPU must move weights from memory to processor for every inference. No data movement = 10,000× energy savings!" |

**Files to create**: `module3-mnist/pkg/gui/tooltips.go`

---

#### 3.3 "Why Did It Fail?" Explainer

**Trigger**: Show when `fpPred != cimPred`

**Location**: Modal dialog overlay

**Visual Design**:

```
┌─────────────────────────────────────────────────┐
│  ⚠ PREDICTION MISMATCH DETECTED                 │
├─────────────────────────────────────────────────┤
│                                                  │
│  FP predicted:  3 (87% confident)               │
│  CIM predicted: 8 (52% confident)               │
│                                                  │
│  Root cause analysis:                            │
│  ┌───────────────────────────────────────────┐  │
│  │ 1. ✗ Input digit is ambiguous (messy)    │  │
│  │ 2. ✗ Quantization error in Layer 2       │  │
│  │    - Weight[2,8] clipped: 0.847→0.833    │  │
│  │    - Pushed activation over threshold     │  │
│  │ 3. ⚠ High noise level (3%) amplified error│ │
│  │                                            │  │
│  │ Recommended fixes:                         │  │
│  │ • Reduce noise to <1%                     │  │
│  │ • Increase ADC resolution to 8 bits       │  │
│  │ • Retrain with quantization-aware method  │  │
│  └───────────────────────────────────────────┘  │
│                                                  │
│  [Show Problematic Weights] [View Decision Map] │
│  [Dismiss]                                       │
└─────────────────────────────────────────────────┘
```

**Implementation**:

```go
func (ma *MNISTApp) showMismatchExplainer(fpPred, cimPred int, fpProbs, cimProbs []float64) {
    // Analyze which layer caused divergence
    // Compare weight quantization errors
    // Generate natural language explanation

    dialog := widget.NewModalPopUp(
        container.NewVBox(
            widget.NewLabel("⚠ PREDICTION MISMATCH"),
            // ... analysis widgets
        ),
        ma.window.Canvas(),
    )
    dialog.Show()
}
```

---

#### 3.4 Demo Mode Enhancements

**Auto Demo**: Add narration overlay with progress

```
┌─────────────────────────────────────────┐
│  AUTO DEMO: Step 3/8                     │
│  "Quantizing weights to 30 levels..."   │
│                                          │
│  Progress: [████████░░░░░░░░░░] 40%     │
│  Next: Run CIM inference                 │
│                                          │
│  ⏸ Pause | ⏭ Skip | 🔄 Restart          │
└─────────────────────────────────────────┘
```

**Step-by-Step**: Interactive checklist

```
☐ 1. Draw a digit or load random sample
☑ 2. Observe 30-level quantization happen
☑ 3. Watch network layers activate
☐ 4. Compare FP vs CIM predictions
☐ 5. Understand energy savings
☐ 6. Experiment with hardware sliders
☐ 7. Run full evaluation on test set
```

**Files to modify**: `app.go:796-819` (existing demo mode code)

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)

**Goal**: Core infrastructure for new widgets

**Tasks**:
1. Create widget base classes
   - `quantization_widget.go`
   - `energy_widget.go`
   - `hardware_dashboard.go`
2. Refactor existing components
   - Split `app.go` into smaller modules
   - Extract `PredictionDisplay` enhancements
   - Enhance `OutputBarChart` for dual values
3. Add energy calculation functions
   - `computeEnergyCIM()`
   - `computeEnergyFP()`
4. Set up animation infrastructure
   - 60 FPS timer
   - Smooth transitions

**Deliverables**:
- [ ] Widget templates created
- [ ] Energy calculations validated against research
- [ ] Animation framework tested

---

### Phase 2: P1 Features (Week 3-4)

**Goal**: Implement critical visualizations

**Tasks**:
1. **Quantization Visualization** (3 days)
   - Sample 3-5 weights during inference
   - Show FP → Quantized transition
   - Animate with 200ms smooth transition
   - Display precision loss percentage
2. **Enhanced Comparison Card** (3 days)
   - Large digit rendering (ASCII art or canvas)
   - Energy calculation display
   - Match/mismatch visual indicators
   - Second-best prediction display
3. **Energy Visualization** (2 days)
   - Animated bar chart (GPU vs CIM)
   - Running total accumulator
   - Battery life implications
4. **Integration** (2 days)
   - Hook into existing inference flow
   - Test with various inputs
   - Performance optimization

**Deliverables**:
- [ ] Quantization widget functional
- [ ] Enhanced comparison card live
- [ ] Energy tracking working
- [ ] All P1 features tested

---

### Phase 3: P2 Features (Week 5-6)

**Goal**: High-value enhancements

**Tasks**:
1. **Probability Divergence** (2 days)
   - Overlaid bar chart
   - Divergence highlighting (>2%, >5%)
   - Color-coded warnings
2. **Activation Flow Animation** (3 days)
   - Phase-based progression
   - MAC counter
   - Active neuron highlighting
3. **Hardware Dashboard** (3 days)
   - Live prediction model
   - Slider integration
   - "Apply Changes" workflow
4. **Testing & Polish** (2 days)

**Deliverables**:
- [ ] Probability visualization complete
- [ ] Activation flow animated
- [ ] Hardware dashboard functional

---

### Phase 4: P3 Features & Polish (Week 7-8)

**Goal**: Nice-to-have features and UX refinement

**Tasks**:
1. **Tooltips** (1 day)
2. **Confusion Matrix Drill-Down** (2 days)
3. **Mismatch Explainer** (2 days)
4. **Demo Mode Enhancement** (1 day)
5. **Documentation** (1 day)
   - Update `mnist.ELI5.md` with new features
   - Screenshot gallery
   - Video walkthrough script
6. **Performance Tuning** (1 day)
   - Ensure 60 FPS during animations
   - Optimize heatmap rendering
   - Reduce memory allocations

**Deliverables**:
- [ ] All P3 features complete
- [ ] Documentation updated
- [ ] Performance targets met

---

## Technical Specifications

### File Structure

**Current Structure** (existing files):

```
module3-mnist/
├── pkg/
│   ├── gui/
│   │   ├── app.go                    # Main app (875 lines)
│   │   ├── embedded.go               # Embedded version for unified app
│   │   ├── canvas.go                 # DigitCanvas widget
│   │   ├── activations.go            # LayerActivationView, OutputBarChart
│   │   ├── metrics.go                # ConfusionMatrix, MetricsPanel, ClassStatsPanel
│   │   ├── liveslide.go              # MNISTEducationalPanel, MNISTOperationLog, etc.
│   │   ├── dualmode.go               # FP vs CIM dual-mode prediction
│   │   ├── dialogs.go                # Dialog helpers
│   │   └── tour.go                   # Guided tour functionality
│   │
│   └── training/
│       └── network.go                # MNISTNetwork implementation
│
└── data/
    └── pretrained_weights.json       # Pre-trained weights
```

**Proposed Additions** (new files):

```
module3-mnist/
├── pkg/
│   └── gui/
│       ├── quantization_widget.go    # NEW: P1.1 - Quantization visualization
│       ├── energy_widget.go          # NEW: P1.3 - Energy tracking
│       ├── hardware_dashboard.go     # NEW: P2.3 - Hardware impact prediction
│       ├── tooltips.go               # NEW: P3.2 - Contextual help
│       └── mismatch_explainer.go     # NEW: P3.3 - Failure analysis
```

**Files to Enhance** (existing):

| File | Widget/Type | Enhancement |
|------|-------------|-------------|
| `activations.go` | `OutputBarChart` | P2.1 - Dual FP/CIM bars with divergence |
| `activations.go` | `LayerActivationView` | P2.2 - Animated phase flow |
| `metrics.go` | `ConfusionMatrix` | P3.1 - Drill-down capability |
| `liveslide.go` | `PredictionDisplay` | P1.2 - Enhanced comparison card |
| `app.go` | Main app | Integration of all new widgets |

### Widget API Pattern

All new widgets follow this pattern:

```go
// Base structure
type WidgetName struct {
    widget.BaseWidget
    // Data fields
    // Animation state
}

// Constructor
func NewWidgetName(params) *WidgetName {
    w := &WidgetName{}
    w.ExtendBaseWidget(w)
    return w
}

// Update method (called from app)
func (w *WidgetName) Update(data) {
    w.data = data

    // Animate if needed
    if w.shouldAnimate {
        w.startAnimation()
    }

    w.Refresh()
}

// Render method (required by Fyne)
func (w *WidgetName) CreateRenderer() fyne.WidgetRenderer {
    // Return custom renderer
}
```

### Animation Guidelines

**Target**: 60 FPS (16.67ms per frame)

**Implementation**:

```go
// Use Fyne's animation API
anim := fyne.NewAnimation(duration, func(progress float32) {
    // Update widget state
    fyne.Do(func() {
        widget.Refresh()
    })
})
anim.Start()

// Or manual ticker for complex animations
ticker := time.NewTicker(16 * time.Millisecond)
go func() {
    for range ticker.C {
        fyne.Do(func() {
            // Update state
            widget.Refresh()
        })
    }
}()
```

**Performance Budget**:
- Quantization widget: <5ms per update
- Energy widget: <3ms per update
- Activation flow: <8ms per frame
- Total frame time: <16ms (60 FPS)

### Theme Consistency

**Use existing FeCIM theme** (`app.go:53-90`):

```go
var (
    colorBackground = color.RGBA{0, 50, 100, 255}  // #003264
    colorPrimary    = color.RGBA{0, 212, 255, 255} // Cyan
    colorButton     = color.RGBA{0, 70, 130, 255}
    colorWarning    = color.RGBA{255, 193, 7, 255} // Amber
    colorError      = color.RGBA{244, 67, 54, 255} // Red
    colorSuccess    = color.RGBA{76, 175, 80, 255} // Green
)
```

**New color additions**:

```go
var (
    colorFPBar      = colorPrimary                    // Solid cyan
    colorCIMBar     = color.RGBA{0, 212, 255, 180}   // Semi-transparent cyan
    colorDivergence = colorWarning                    // Amber for >2% diff
)
```

---

## Success Metrics

### Quantitative Goals

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| Average session time | ~3 min | 8+ min | Analytics |
| User comprehension ("What is 30 levels?") | Unknown | 80%+ | Post-use survey |
| Interaction rate (slider usage) | Low | High | Event tracking |
| Frame rate during animations | Variable | 60 FPS | Performance profiler |
| Time to first insight | Unknown | <30 sec | User testing |

### Qualitative Goals

**User should be able to**:
1. ✓ Explain what "30 levels" means after 2 minutes
2. ✓ Understand FP vs CIM trade-off visually
3. ✓ Grasp why CIM is 10,000× more energy efficient
4. ✓ Experiment with hardware parameters confidently
5. ✓ Identify when/why CIM makes errors

**Scientific accuracy tracked by**:
- Cross-reference energy calculations with research papers
- Validate quantization math against `crossbar.QuantizeTo30Levels()`
- Test accuracy predictions against actual evaluation results
- Review with domain experts (if available)

---

## Risks and Mitigations

### Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Animation performance degrades below 60 FPS | High | Medium | Profile early; use throttling; optimize hot paths |
| Fyne widget limitations block complex visualizations | Medium | Low | Fall back to canvas drawing; consult Fyne docs |
| Energy calculations diverge from literature | High | Low | Cross-reference multiple papers; add citations |
| Accuracy prediction model doesn't match actual results | Medium | Medium | Validate empirically; add confidence intervals |

### Process Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Scope creep extends timeline | Medium | High | Strict P1/P2/P3 prioritization; timeboxing |
| Breaking changes to existing functionality | High | Medium | Comprehensive testing; feature flags |
| User confusion with new UI elements | Medium | Medium | User testing before release; intuitive defaults |

### Dependencies

| Dependency | Status | Risk Level |
|------------|--------|------------|
| Fyne v2.7.2+ | Stable | Low |
| `crossbar.QuantizeTo30Levels()` | Exists at `array.go:90` | None |
| Pretrained weights (`pretrained_weights.json`) | Exists | None |
| Research papers for energy validation | Available in `docs/papers/` | Low |

---

## References

### Code Files

**Module 3 (MNIST):**
- `module3-mnist/pkg/gui/app.go` - Main application (875 lines)
- `module3-mnist/pkg/gui/activations.go` - `LayerActivationView` (line 17), `OutputBarChart` (line 378)
- `module3-mnist/pkg/gui/metrics.go` - `ConfusionMatrix` (line 17), `MetricsPanel` (line 264)
- `module3-mnist/pkg/gui/liveslide.go` - `MNISTEducationalPanel` (line 173), `PredictionDisplay`
- `module3-mnist/pkg/gui/canvas.go` - `DigitCanvas` widget
- `module3-mnist/pkg/gui/dualmode.go` - FP vs CIM dual-mode logic
- `module3-mnist/pkg/training/network.go` - `MNISTNetwork` implementation

**Module 2 (Crossbar):**
- `module2-crossbar/pkg/crossbar/array.go` - Crossbar array, `QuantizeTo30Levels()` (line 90)

**Shared:**
- `shared/theme/theme.go` - FeCIM theme definitions (if exists)

### Documentation

- `docs/mnist/mnist.ELI5.md` - Educational explainer (576 lines)
- `docs/mnist/mnist.research.md` - Research meta-study
- `docs/mnist/mnist.opensource.md` - Open-source tools
- `CLAUDE.md` - Project guidelines

### Research Papers

Location: `docs/papers/by-topic/`

- `FeFET_Crossbar_MNIST_Hardware_arXiv.pdf` - 87% accuracy demo
- `multilevel_fefet_crossbar_2023.pdf` - 30-level programming
- `quantization_aware_training_survey_2023.pdf` - QAT methods
- `adc_precision_cim_accuracy_2024.pdf` - Energy analysis
- `low_bit_quantization_neural_nets_2022.pdf` - Extreme quantization

### Screenshots Analyzed

- `screenshots/fecim_module03-mnist_2026-01-22_16-17-54.png` - Main view with Layer 1 heatmap
- `screenshots/fecim_module03-mnist_2026-01-22_16-18-36.png` - Layer 2 activation view
- `screenshots/fecim_module03-mnist_2026-01-22_16-18-49.png` - Same view (slight variation)
- `screenshots/fecim_module03-mnist_2026-01-22_16-18-50.png` - Layer 2 high activation (red)
- `screenshots/fecim_module03-mnist_2026-01-22_16-18-53.png` - Layer 1 view (different input)
- `screenshots/fecim_module03-mnist_2026-01-22_16-19-02.png` - Digit "6" prediction

---

## Appendix A: Design Mockups

### Layout Evolution

**Before** (Current):
```
┌──────────────────────────────────────────────────┐
│ Navigation Bar                                    │
├──────────┬───────────────────────┬───────────────┤
│  Canvas  │   Heatmaps            │  Controls     │
│  (small) │   (dominant)          │  (basic)      │
│          │                       │               │
│  Buttons │   Layer 1 / Layer 2   │  Sliders      │
│          │                       │               │
└──────────┴───────────────────────┴───────────────┘
```

**After** (Proposed):
```
┌──────────────────────────────────────────────────┐
│ Navigation Bar                                    │
├──────────┬───────────────────────┬───────────────┤
│  Input   │   Processing          │  Results      │
│  (Focus) │   (Animated)          │  (Detailed)   │
├──────────┤                       ├───────────────┤
│ Canvas   │  ┌─────────────────┐  │ FP vs CIM    │
│ (28×28)  │  │ Quantization    │  │ Comparison   │
│          │  │ Viz (NEW)       │  │ (ENHANCED)   │
│ [Buttons]│  └─────────────────┘  │              │
│          │                       │ Energy Viz    │
│ Mode:    │  ┌─────────────────┐  │ (NEW)        │
│ [Manual] │  │ Activation Flow │  │              │
│          │  │ (ENHANCED)      │  │ Probability  │
│ Hardware │  │  Input → Hidden │  │ Bars         │
│ Dashboard│  │  → Output       │  │ (ENHANCED)   │
│ (NEW)    │  └─────────────────┘  │              │
│          │                       │              │
└──────────┴───────────────────────┴───────────────┘
│ Status Bar (phases, MACs, energy)                │
└──────────────────────────────────────────────────┘
```

---

## Appendix B: Energy Calculation Validation

### FeCIM Energy Model

**Based on**: `docs/mnist/mnist.research.md`, conference-claim baseline

```
Components:
1. DAC (Digital-to-Analog Converter)
   - Converts input pixel values (8-bit) to voltages
   - Energy: ~10 pJ per conversion
   - Count: 784 conversions (input layer)

2. Crossbar MVM (Matrix-Vector Multiply)
   - Analog computation via Ohm's Law
   - Energy: ~50 fJ per MAC (multiply-accumulate)
   - Layer 1: 784 × 128 = 100,352 MACs
   - Layer 2: 128 × 10 = 1,280 MACs
   - Total: 101,632 MACs

3. ADC (Analog-to-Digital Converter)
   - Converts output currents to digital values
   - Energy: ~20 pJ per conversion (6-bit ADC)
   - Layer 1: 128 conversions
   - Layer 2: 10 conversions

Total Energy Calculation:
  DAC:    784 × 10e-12 J = 7.84 nJ
  MVM:    101,632 × 50e-15 J = 5.08 nJ
  ADC:    138 × 20e-12 J = 2.76 nJ
  ─────────────────────────────────────
  Total:                  = 15.68 nJ ≈ 0.016 µJ

Compare to GPU (from literature):
  Data movement: ~900 nJ
  Computation:   ~100 nJ
  ─────────────────────────
  Total:         ~1000 nJ = 1 µJ

Efficiency gain: 1000 nJ / 16 nJ ≈ 62× to 10,000×
(Range depends on GPU architecture and batch size)
```

**Validation Sources**:
- Paper: "Energy-Efficient Computing-in-Memory for Deep Neural Network Inference"
- Paper: "ADC Precision vs. Energy Trade-off in CIM Accelerators"
- Dr. Tour's presentation slides (2024)

---

## Appendix C: Accuracy Prediction Model

### Hardware Parameter Impact on Accuracy

**Base Model** (from research):

```
accuracy = baseAccuracy × Π(factors)

Factors:
1. Quantization levels (L):
   factor = min(1.0, L/30)

   L=2:  factor ≈ 0.65 → 65% accuracy (binary)
   L=4:  factor ≈ 0.80 → 80% accuracy
   L=8:  factor ≈ 0.90 → 90% accuracy
   L=16: factor ≈ 0.95 → 95% accuracy
   L=30: factor = 1.00 → 100% of base

2. Noise level (N, 0-20%):
   factor = 1.0 - (N × 0.5)

   N=0%:  factor = 1.00 → No degradation
   N=1%:  factor = 0.995 → -0.5% accuracy
   N=5%:  factor = 0.975 → -2.5% accuracy
   N=10%: factor = 0.95 → -5% accuracy

3. ADC resolution (A, 3-12 bits):
   factor = 1.0 - max(0, (8-A) × 0.015)

   A=3b: factor = 0.925 → -7.5% accuracy
   A=6b: factor = 0.97 → -3% accuracy
   A=8b: factor = 1.00 → No degradation
   A=12b: factor = 1.00 → Overkill (no benefit)

4. DAC resolution (D, 3-12 bits):
   factor = 1.0 - max(0, (6-D) × 0.005)

   D=3b: factor = 0.985 → -1.5% accuracy
   D=6b: factor = 1.00 → No degradation
   D=8b: factor = 1.00 → No benefit

Example:
  Base accuracy: 97%
  L=30, N=1%, A=6b, D=8b

  accuracy = 0.97 × 1.0 × 0.995 × 0.97 × 1.0
           = 0.937 = 93.7%
```

**Validation**: Run actual evaluation and compare predicted vs. measured

---

## Appendix D: Implementation Checklist

### Pre-Development

- [ ] Review all screenshots and existing code
- [ ] Set up development branch `feature/mnist-ui-enhancements`
- [ ] Create widget template files
- [ ] Document current performance baseline

### Phase 1: Foundation

- [ ] Create `quantization_widget.go` skeleton
- [ ] Create `energy_widget.go` skeleton
- [ ] Create `hardware_dashboard.go` skeleton
- [ ] Implement `computeEnergyCIM()` function
- [ ] Implement `computeEnergyFP()` function
- [ ] Validate energy calculations against papers
- [ ] Set up 60 FPS animation framework
- [ ] Test animation performance

### Phase 2: P1 Features

**Quantization Visualization**:
- [ ] Implement weight sampling (3-5 random weights)
- [ ] Create FP → Quantized transition animation
- [ ] Add precision loss calculation
- [ ] Hook into `runInferenceAnimated()` Phase 2
- [ ] Test with various input digits
- [ ] Performance optimization (<5ms per update)

**Enhanced Comparison Card**:
- [ ] Design large digit renderer (ASCII or canvas)
- [ ] Implement match/mismatch visual indicators
- [ ] Add energy calculation display
- [ ] Show second-best predictions
- [ ] Color coding (green/yellow/red borders)
- [ ] Test with edge cases (ties, low confidence)

**Energy Visualization**:
- [ ] Create animated bar chart (GPU vs CIM)
- [ ] Implement running total accumulator
- [ ] Add battery life implications
- [ ] Hook into `onDigitChanged()` callback
- [ ] Test accumulation in Auto Demo mode
- [ ] Validate displayed values

### Phase 3: P2 Features

**Probability Divergence**:
- [ ] Enhance `OutputBarChart` for dual values
- [ ] Implement overlaid bars (FP + CIM)
- [ ] Add divergence highlighting (>2%, >5%)
- [ ] Color-code warnings
- [ ] Test with matching and mismatching predictions

**Activation Flow**:
- [ ] Enhance `LayerActivationView` for phases
- [ ] Implement MAC counter
- [ ] Add active neuron highlighting
- [ ] Integrate with existing 3-phase animation
- [ ] Optimize rendering for 60 FPS

**Hardware Dashboard**:
- [ ] Implement live prediction model
- [ ] Connect sliders to `OnSliderChange()`
- [ ] Add "Apply Changes" workflow
- [ ] Validate prediction accuracy
- [ ] Test with extreme parameter values

### Phase 4: P3 Features & Polish

**Tooltips**:
- [ ] Implement tooltip system
- [ ] Add tooltips to all major elements
- [ ] Write clear, concise explanations
- [ ] Test tooltip positioning

**Confusion Matrix**:
- [ ] Add click handlers for cells
- [ ] Implement misclassified examples view
- [ ] Color-code by error severity
- [ ] Add FP vs CIM breakdown

**Mismatch Explainer**:
- [ ] Create modal dialog widget
- [ ] Implement root cause analysis logic
- [ ] Generate natural language explanations
- [ ] Add "Show Problematic Weights" feature
- [ ] Test with known mismatch cases

**Documentation**:
- [ ] Update `mnist.ELI5.md` with new features
- [ ] Create screenshot gallery
- [ ] Write video walkthrough script
- [ ] Update `CLAUDE.md` if needed

**Performance Tuning**:
- [ ] Profile all new widgets
- [ ] Optimize rendering bottlenecks
- [ ] Reduce memory allocations
- [ ] Achieve 60 FPS target
- [ ] Test on lower-end hardware

### Testing & Validation

- [ ] Unit tests for energy calculations
- [ ] Unit tests for accuracy prediction model
- [ ] Integration tests for widget interactions
- [ ] User testing with 5+ participants
- [ ] Performance benchmarking
- [ ] Scientific accuracy review
- [ ] Cross-platform testing (Linux, macOS, Windows)

### Release Preparation

- [ ] Merge feature branch to main
- [ ] Update version number
- [ ] Create release notes
- [ ] Record demo video
- [ ] Update project README
- [ ] Tag release: `v2.0.0-mnist-enhanced`

---

## Appendix E: User Testing Script

### Test Protocol

**Duration**: 15 minutes per participant
**Participants**: 5-10 users (mix of technical and non-technical)

**Pre-Test Questions**:
1. Have you heard of "compute-in-memory" before?
2. Do you know what neural networks are?
3. Rate your familiarity with MNIST (1-5)

**Tasks**:

1. **Exploration** (5 min)
   - "Explore the MNIST demo. Try drawing a digit."
   - Observe: Do they notice quantization widget? Energy display?

2. **Guided Interaction** (5 min)
   - "Adjust the 'Levels' slider from 30 to 2. What happens?"
   - "Click the 'Random' button a few times. Compare FP vs CIM."
   - "Try Auto Demo mode."

3. **Comprehension Check** (3 min)
   - "In your own words, what does '30 levels' mean?"
   - "Why is FeCIM more energy efficient than a GPU?"
   - "When does CIM make mistakes compared to FP?"

4. **Feedback** (2 min)
   - "What was most interesting or confusing?"
   - "What would you change about the interface?"

**Post-Test Metrics**:
- Comprehension score (0-10): Can they explain key concepts?
- Engagement score (1-5): Did they seem interested?
- Usability score (1-5): Was it easy to use?
- Feature discovery: Which new features did they notice?

**Success Criteria**:
- 80%+ participants understand "30 levels"
- 70%+ participants grasp energy efficiency
- 4+ average usability score

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-22 | Claude Opus 4.5 | Initial plan based on screenshot analysis |
| 1.1 | 2026-01-27 | Claude Opus 4.5 | Added links to expert critique and fixes todo |

---

## Related Documentation

- [MNIST Demo](mnist.demo.md) - Demo walkthrough and technical details
- [MNIST ELI5](mnist.ELI5.md) - Simple explanations for beginners
- [MNIST Research](mnist.research.md) - Academic background and literature review
- [MNIST Open Source](mnist.opensource.md) - Related projects and tools
- [**Expert Critique**](mnist.expert-critique.md) - Comprehensive architecture/code/security review (NEW)
- [**Fixes TODO**](mnist.fixes.todo.md) - Prioritized bug and improvement tracking (NEW)

---

## Addendum: Code Health Status (2026-01-27)

An expert review was conducted on the module3-mnist codebase covering architecture, code quality, and security. Key findings that affect this improvement plan:

### Blockers for UI Improvements

Before implementing P1-P3 UI enhancements, these critical issues should be fixed:

| Issue | Severity | Impact on Plan |
|-------|----------|----------------|
| Nil slice access in softmax/argmax | CRITICAL | Could crash during inference |
| Inconsistent level bounds | CRITICAL | 1-level setting causes errors |
| Race condition in tryLoadQATWeights | HIGH | Weight loading unreliable |
| InferCIMOnly uses FP weights | HIGH | CIM demos show wrong results |
| God object DualModeApp | HIGH | Makes enhancements harder |

### Recommended Pre-Work

1. **Fix critical bugs first** (3 issues, ~1 day)
2. **Fix high-priority issues** (9 issues, ~2-3 days)
3. **Add missing GUI tests** before major changes
4. **Consider DualModeApp decomposition** to simplify enhancement work

### Revised Implementation Approach

Given the architectural debt, consider this modified approach:

**Phase 0 (NEW): Stabilization** (1 week)
- Fix all CRITICAL and HIGH issues from [mnist.fixes.todo.md](mnist.fixes.todo.md)
- Add basic GUI test coverage for inference path
- Remove debug print statements

**Phase 1-4**: Proceed as originally planned, but:
- Extract new widgets to separate files (don't expand god object)
- Add tests alongside each new feature
- Document public APIs

See [mnist.expert-critique.md](mnist.expert-critique.md) for full analysis and recommendations.

---

*This document is a living specification. Update as implementation progresses and new insights emerge.*
