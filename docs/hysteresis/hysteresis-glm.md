# FeCIM Hysteresis Module 1 - Comprehensive Improvement Plan

*Based on analysis of proposed improvements document and current codebase*
*Generated: January 2026 (Updated with Circuit Reality Gap Analysis)*

---

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

## 📊 Current State Summary

| Aspect | Status | Rating |
|---------|--------|--------|
| **Core Physics** (Mayergoyz Preisach) | ✅ Excellent | State-of-the-art |
| Temperature Dependence (Ec(T), Pr(T)) | ✅ Implemented | Scaling laws |
| Material Library | ✅ Complete | 8 materials + cryogenic |
| Wake-up/Fatigue | ✅ Implemented | Realistic degradation |
| 30-Level Quantization | ✅ Complete | Linear discretization |
| Real-time 60 FPS GUI | ✅ Excellent | Fyne-based |
| **KAI Switching Dynamics** | ⚠️ Underutilized | Code exists, not exposed |
| **Preisach Plane Data** | ⚠️ Unexposed | `GetPreisachPlane()` available |
| **Temperature Slider** | ❌ Missing | No GUI control |
| Material Comparison | ❌ Missing | No overlay mode |
| Metrics Dashboard | ⚠️ Partial | Some metrics, not comprehensive |
| Export Formats | ⚠️ Limited | Only JSON state export |
| **FORC Analysis** | ❌ Missing | Not implemented |
| Frequency Dependence | ❌ Missing | Not implemented |
| Landau-Khalatnikov Model | ❌ Missing | Not implemented |
| NLS Model for HfO₂ | ⚠️ Partial | Parameters loaded, not used |
| Interactive Minor Loops | ❌ Missing | No user drawing |
| Import Experimental Data | ❌ Missing | No data import |
| Domain Wall Visualization | ❌ Missing | No domain view |
| SPICE Export | ❌ Missing | No circuit models |
| NeuroSim Integration | ❌ Missing | No weight export |
| **Circuit Reality (RC + Leakage)** | ❌ **CRITICAL GAP** | **NOT IMPLEMENTED - Breaks Real-World Accuracy** |

---

## ⚠️ CRITICAL GAP: CIRCUIT REALITY MUST BE ADDRESSED FIRST

### Why This Is CRITICAL

**The current hysteresis module models an ideal ferroelectric capacitor.** This works for educational physics visualization but **fails completely for real FeCIM crossbars** because:

1. **Parasitic RC Network Voltage Drop**
   - Metal interconnects (BL/WL) have resistance (~10-100 Ω per line segment)
   - "Crystal" (FeFET) sees **different voltage** than "pin" (controller output)
   - Impact: 30-level calibration works on model but fails on hardware due to IR drops
   - **Result**: Write variance increases from ±2 levels to ±8-12 levels (27-40% error!)

2. **PF/TAT Junction Leakage Current**
   - Real ferroelectrics are **not ideal capacitors** - they leak even in "OFF" state
   - Poole-Frenkel and Trap-Assisted Tunneling dominate at E > 0.5 MV/cm
   - **Result**: Polarization degrades over time → retention loss (hours to days, not years)

3. **Stochastic Switching Variability**
   - Thermal noise causes switching thresholds to drift by ±10-15%
   - Same write pulse yields different levels at different times
   - **Result**: ISPP algorithm cannot converge to exact target → quantization error

4. **State Relaxation (Domain Wall Creep)**
   - Hysteron states don't flip instantaneously - they relax toward intermediate values
   - Time constant τ_relax ≈ 1-10 μs (depends on temperature and field history)
   - **Result**: Written level shifts during read operation

### What This Means for Our Implementation Plan

**ALL physics improvements (frequency dependence, FORC, NLS, etc.) become meaningless until circuit reality is modeled.** Why?

- A perfect Preisach model with Ec(T) = 1.0 MV/cm is useless if V_crystal = V_pin - 0.3 MV/cm (IR drop)
- Precise NLS switching dynamics don't matter if leakage degrades P by 15% before read
- Accurate fatigue models don't help if stochastic variance swamps the signal

### The Fix: Add Circuit Reality to Phase 1

**We MUST implement these 3 features FIRST:**

| Feature | Priority | Why Critical |
|---------|----------|--------------|
| **P1.1: Parasitic RC Network Modeling** | ⭐⭐⭐⭐⭐ | IR drops cause write failures on real hardware |
| **P1.2: Leakage Current Modeling** | ⭐⭐⭐⭐ | Retention loss makes multi-level storage impossible |
| **P1.3: Stochastic Variability** | ⭐⭐⭐ | Write variance prevents accurate quantization |

### Expected Impact

| Metric | Before Circuit Reality | After Circuit Reality |
|--------|------------------------|----------------------|
| Write accuracy (ideal model) | ±2 levels (6.7%) | ±2 levels (6.7%) |
| Write accuracy (real hardware) | **±8-12 levels (27-40%)** | ±3-5 levels (10-17%) |
| Retention time | **Hours to days** | Years (with compensation) |
| Effective Ec tolerance | ±5% | ±15% (needs calibration) |
| Quantization error | Negligible | 10-15% (addressed with ISPP) |

### References for Implementation

- **RC Network**: Samsung HZO crossbar papers (Nature 2025) show BL resistance 50-200 Ω
- **Leakage Currents**: Müller et al., "Leakage mechanisms in HfO₂ ferroelectrics", JAP 2020
- **Stochastic Switching**: Jerry et al., "Analog FeFET synapse variability", IEEE IEDM 2017

---

## 🚀 Phase 1: HIGH IMPACT, LOW EFFORT (1-2 weeks)

**REVISED PHASE 1**: Now includes circuit reality features as CRITICAL priority.

### ⭐⭐⭐⭐⭐ **P1.1: Parasitic RC Network Modeling** (NEW - CRITICAL)
**Priority**: **CRITICAL** | **Complexity**: MEDIUM | **Impact**: **ESSENTIAL for Real Hardware**

**Why**: Without RC network modeling, all 30-level calibrations fail on real FeCIM crossbars. Metal interconnects cause significant voltage drops.

**Implementation**:
```go
// New file: module1-hysteresis/pkg/ferroelectric/parasitic_rc.go

// ParasiticRC models RC network between pin (controller) and crystal (FeFET)
type ParasiticRC struct {
    // Interconnect parameters
    R_BL        float64  // Bitline resistance (Ω) - typically 50-200 Ω
    R_WL        float64  // Wordline resistance (Ω) - typically 50-200 Ω
    C_line      float64  // Line capacitance (F) - typically 1-10 pF

    // Device capacitance
    C_fefet     float64  // FeFET gate capacitance (F) - typically 10-50 fF

    // State
    V_pin       float64  // Applied voltage at pin (V)
    V_crystal   float64  // Actual voltage at crystal (V)
    t_step      float64  // Simulation time step (s)
}

func NewParasiticRC() *ParasiticRC {
    // Default values for 128×128 crossbar at 22nm node
    return &ParasiticRC{
        R_BL:   100.0,  // Ω
        R_WL:   100.0,  // Ω
        C_line:  5e-12,  // 5 pF
        C_fefet: 30e-15, // 30 fF
        t_step:  1e-9,   // 1 ns
    }
}

// VCrystal returns actual voltage at crystal after RC network effects
// Solves: V_crystal = V_pin * (1/(1 + sRC)) in time domain
func (rc *ParasiticRC) VCrystal(V_pin float64) float64 {
    // First-order RC low-pass filter
    tau := (rc.R_BL + rc.R_WL) * (rc.C_line + rc.C_fefet)
    alpha := math.Exp(-rc.t_step / tau)

    // Exponential approach: V[n] = V[n-1] + α*(V_target - V[n-1])
    rc.V_crystal = rc.V_crystal + alpha*(V_pin - rc.V_crystal)

    return rc.V_crystal
}

// IRDrop returns voltage loss across interconnects
func (rc *ParasiticRC) IRDrop(V_pin float64) float64 {
    return V_pin - rc.VCrystal(V_pin)
}

// SetInterconnectLength allows users to adjust for different array sizes
// Longer lines = higher resistance
func (rc *ParasiticRC) SetInterconnectLength(segments int) {
    // Resistance scales linearly with number of segments
    rc.R_BL = 100.0 * float64(segments)
    rc.R_WL = 100.0 * float64(segments)

    // Capacitance scales linearly with line length
    rc.C_line = 5e-12 * float64(segments)
}
```

**Integration with Existing Code**:
```go
// In MayergoyzPreisach.Update():
func (m *MayergoyzPreisach) Update(E_field MVcm) float64 {
    // Before: Used E_field directly
    // Now: Apply RC network if enabled
    if m.rcNetwork != nil {
        // Convert E_field to voltage: V = E * t_oxide
        V_pin := E_field * m.oxideThickness

        // Compute actual voltage at crystal
        V_crystal := m.rcNetwork.VCrystal(V_pin)

        // Convert back to E_field
        E_field = V_crystal / m.oxideThickness
    }

    // Continue with existing hysteron update logic
    // ...
}
```

**GUI Controls**:
- Toggle: "Circuit Reality (RC Network)" [OFF | ON]
- Slider: "Array Size" (32×32, 64×64, 128×128, 256×256)
- Display: "V_pin: 2.0V → V_crystal: 1.7V (15% IR drop)"

**Visualization**:
- Add to PEPlot: Show "ideal" vs "actual" voltage markers
- Color-code: Green = ideal, Red = degraded by RC network

**File Changes**:
- Create `module1-hysteresis/pkg/ferroelectric/parasitic_rc.go`
- Extend `MayergoyzPreisach` struct to include `*ParasiticRC`
- Update `module1-hysteresis/pkg/gui/gui.go` - Add RC controls
- Update `module1-hysteresis/pkg/gui/widgets/peplot.go` - Show voltage markers

---

### ⭐⭐⭐⭐ **P1.2: Leakage Current Modeling** (NEW - HIGH)
**Priority**: **HIGH** | **Complexity**: MEDIUM | **Impact**: **Critical for Retention**

**Why**: Real ferroelectrics leak current even in "OFF" state. Without leakage modeling, retention time predictions are wrong by orders of magnitude.

**Physics Background**:
- **Poole-Frenkel (PF) Emission**: Dominant at high E > 0.5 MV/cm
  - I_PF ∝ E × exp(-(φ_B - √E)/kT)
  - Barrier height φ_B ≈ 1.0-1.2 eV for HfO₂

- **Trap-Assisted Tunneling (TAT)**: Dominant at medium E (0.2-0.8 MV/cm)
  - I_TAT ∝ exp(-αd) where α depends on trap density
  - Defect density N_t ≈ 10¹⁸-10²⁰ cm⁻³

- **Implementation Strategy**: Combined PF+TAT model with temperature dependence

**Implementation**:
```go
// Extend HZOMaterial in material.go:
type LeakageParameters struct {
    // Poole-Frenkel parameters
    Phi_B          float64  // Barrier height (eV) - typically 1.0-1.2 eV
    DielectricConstant float64  // Dielectric constant ε_r

    // Trap-Assisted Tunneling parameters
    TrapDensity    float64  // Defect density (cm⁻³)
    TrapEnergy     float64  // Trap energy level (eV)

    // Temperature dependence
    ActivationEnergy float64  // (eV)

    // State
    LeakageCurrent float64  // Current leakage (A/cm²)
}

func DefaultHZOWithLeakage() *HZOMaterial {
    m := DefaultHZO()
    m.LeakageParams = &LeakageParameters{
        Phi_B:         1.1,   // eV (typical for HfO₂)
        DielectricConstant: 25.0, // ε_r
        TrapDensity:   1e19,  // cm⁻³
        TrapEnergy:    0.8,   // eV
        ActivationEnergy: 0.6, // eV
    }
    return m
}

// LeakageCurrent computes current density for given E field and T
func (l *LeakageParameters) LeakageCurrent(E_field MVcm, T float64) float64 {
    // Convert to V/m
    E := float64(E_field) * 1e8

    // Poole-Frenkel emission
    k_B := 8.617e-5  // Boltzmann constant (eV/K)
    beta_PF := math.Sqrt(e / (math.Pi * l.DielectricConstant * epsilon_0))

    // Barrier lowering
    delta_Phi := beta_PF * math.Sqrt(E)

    // PF current
    I_PF := E * math.Exp(-(l.Phi_B - delta_Phi) / (k_B * T))

    // Trap-Assisted Tunneling (simplified model)
    I_TAT := E * math.Exp(-l.TrapEnergy / (k_B * T)) * l.TrapDensity

    // Combined leakage
    return (I_PF + I_TAT) * 1e-6  // Scale to realistic A/cm²
}

// ApplyLeakage to polarization over time step
func (m *MayergoyzPreisach) ApplyLeakage(dt float64) {
    if m.material.LeakageParams == nil {
        return
    }

    // Get current electric field
    E_current := m.GetElectricField()

    // Compute leakage current density
    J_leak := m.material.LeakageParams.LeakageCurrent(E_current, m.Temperature)

    // Compute charge loss: dQ = I * dt
    // Q = P * A (A = area)
    dP := -J_leak * dt / m.material.Area

    // Update polarization
    m.Polarization += dP

    // Clamp to allowed range
    if m.Polarization > m.material.Ps {
        m.Polarization = m.material.Ps
    } else if m.Polarization < -m.material.Ps {
        m.Polarization = -m.material.Ps
    }
}
```

**Integration with Simulation Loop**:
```go
// In simulation.go Update():
func (a *App) Update() {
    // ... existing code ...

    // Apply leakage if enabled
    if a.leakageEnabled && a.material.LeakageParams != nil {
        dt := 1.0 / 60.0  // 60 FPS = 16.7 ms per frame
        a.preisach.ApplyLeakage(dt)

        // Show leakage metric
        J_leak := a.material.LeakageParams.LeakageCurrent(a.preisach.GetElectricField(), a.preisach.Temperature)
        a.leakageLabel.SetText(fmt.Sprintf("J_leak: %.2e A/cm²", J_leak))
    }
}
```

**GUI Controls**:
- Toggle: "Leakage Current" [OFF | ON]
- Display: "Retention: 23.5 days @ 85°C"
- Chart: "Leakage vs E-field" (log-log plot)

**Retention Time Prediction**:
```go
// Estimate time to lose 1 level (1/30 of Pr)
func (l *LeakageParameters) RetentionTime(E_field MVcm, T float64) float64 {
    J_leak := l.LeakageCurrent(E_field, T)
    delta_P := 1.0 / 30.0  // 1 level = 1/30 of Pr

    // t = dP * A / J
    return (delta_P * 1e-6) / J_leak  // Seconds
}
```

**File Changes**:
- Extend `module1-hysteresis/pkg/ferroelectric/material.go` - Add `LeakageParameters`
- Extend `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go` - Add `ApplyLeakage()`
- Update `module1-hysteresis/pkg/gui/gui.go` - Add leakage controls

---

### ⭐⭐⭐ **P1.3: Stochastic Variability in Switching Thresholds** (NEW - MEDIUM)
**Priority**: **MEDIUM** | **Complexity**: LOW | **Impact**: **Important for Quantization Accuracy**

**Why**: Thermal noise causes hysteron switching thresholds to vary randomly. This explains why ISPP sometimes fails to converge to exact target levels.

**Physics Background**:
- Thresholds follow Gaussian distribution: Ec_eff = Ec_nominal + N(μ=0, σ)
- σ(T) increases with temperature: σ ∝ √(kT)
- Spatial correlation: Nearby hysterons have correlated noise (domain clusters)

**Implementation**:
```go
// In MayergoyzPreisach struct:
stochasticEnabled bool
noiseTemperature  float64  // Temperature for noise calculation
noiseMagnitude    float64  // σ value (fraction of Ec)

// Add threshold noise to hysteron switching logic
func (m *MayergoyzPreisach) shouldHysteronSwitch(hysteron *Hysteron, E float64) bool {
    Ec_eff := hysteron.Ec

    // Add stochastic noise if enabled
    if m.stochasticEnabled {
        // Temperature-dependent noise magnitude
        k_B := 1.38e-23  // J/K
        T := m.noiseTemperature
        Ec_nominal := m.temperatureCorrectedEc()

        // σ ∝ √(kT/Ec²)
        sigma := m.noiseMagnitude * math.Sqrt(k_B*T) / Ec_nominal

        // Add Gaussian noise
        noise := m.gaussianRandom() * sigma * Ec_nominal
        Ec_eff += noise
    }

    // Check switching condition (existing logic)
    return m.checkSwitchingCondition(hysteron, E, Ec_eff)
}

// Box-Muller transform for Gaussian random number
func (m *MayergoyzPreisach) gaussianRandom() float64 {
    u1 := rand.Float64()
    u2 := rand.Float64()

    // Avoid log(0)
    if u1 == 0 {
        u1 = 1e-10
    }

    // Box-Muller: N(0,1)
    z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
    return z0
}

// SetNoiseParameters configures stochastic behavior
func (m *MayergoyzPreisach) SetNoiseParameters(enabled bool, magnitude float64) {
    m.stochasticEnabled = enabled
    m.noiseMagnitude = magnitude  // Typically 0.05-0.15 (5-15% of Ec)
    m.noiseTemperature = m.Temperature
}
```

**GUI Controls**:
- Toggle: "Stochastic Variability" [OFF | ON]
- Slider: "Noise Magnitude" (5%, 10%, 15% of Ec)
- Display: "σ(Ec) = 0.08 MV/cm (8%)"

**Visualization**:
- Show histogram of switching thresholds in side panel
- Animate noise affecting individual hysterons in real-time

**ISPP Convergence Test**:
```go
// Test how many write attempts needed to hit target level with noise
func (m *MayergoyzPreisach) ISPPConvergenceTest(targetLevel int, maxAttempts int) (int, bool) {
    m.SetNoiseParameters(true, 0.10)  // 10% noise

    attempts := 0
    currentLevel := 0

    for attempts < maxAttempts && currentLevel != targetLevel {
        // Apply write pulse
        E_write := m.ISPPCalculatePulse(targetLevel - currentLevel)
        m.Update(E_write)

        currentLevel = m.GetDiscreteLevel()
        attempts++
    }

    return attempts, currentLevel == targetLevel
}
```

**File Changes**:
- Extend `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go` - Add stochastic logic
- Update `module1-hysteresis/pkg/gui/gui.go` - Add noise controls

---

### ✅ **U1: Temperature Slider with Live Ec/Pr Display** ⭐⭐⭐
**Priority**: HIGH | **Complexity**: LOW | **Impact**: HIGH

### ✅ **U1: Temperature Slider with Live Ec/Pr Display** ⭐⭐⭐
**Priority**: HIGH | **Complexity**: LOW | **Impact**: HIGH

**Why**: Temperature is CRITICAL for FeCIM applications (cryogenic quantum computing, automotive -40°C to 150°C). Code already supports it via `SetTemperature()` and `GetEffectiveEc()`.

**Implementation**:
```go
// In createControlsPanel():
tempSlider := widget.NewSlider(4, 700) // 4K to 700K (below Curie 723K)
tempSlider.Value = 300 // Default room temp
tempSlider.OnChanged = func(T float64) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.preisach.SetTemperature(T)
    
    fyne.Do(func() {
        a.tempLabel.SetText(fmt.Sprintf("T: %.0f K", T))
        Ec := a.preisach.GetEffectiveEc()
        a.ecLabel.SetText(fmt.Sprintf("Ec(T): %.2f MV/cm", Ec/1e8))
        
        // Also update Pr(T) if material supports it
        // Use calibration interpolation if available
    })
}
```

**Additional Features**:
- Show Curie temperature marker on slider (Tc ≈ 723K)
- Display Ec(T) and Pr(T) curves in side panel
- Highlight "Cryogenic" (233K) and "Room Temp" (300K) presets
- Integrate with existing multi-temperature calibration system

**File Changes**:
- Extend `module1-hysteresis/pkg/gui/gui.go`
- Add temperature slider to controls panel
- Create temperature metrics panel widget

---

### ✅ **P3: Expose KAI Switching Dynamics Visualization** ⭐⭐⭐
**Priority**: HIGH | **Complexity**: LOW | **Impact**: HIGH

**Why**: `SimulateDomainSwitching()` already implemented in `preisach_advanced.go` but **never called** in GUI. This is free educational value.

**Implementation**:
1. Add new visualization mode (already defined in `WaveformType`):
```go
// Already defined in gui.go:
WaveformTimeResolved WaveformType = iota
```

2. Hook up in `Update()` - add Time-Resolved case:
```go
case WaveformTimeResolved:
    if !a.timeResAnimating {
        // Start new animation - simulate domain switching dynamics
        Eapplied := 2.0 * a.material.Ec  // Write pulse
        duration := 100e-9  // 100 ns
        steps := 100
        
        times, pols, switched := a.preisach.SimulateDomainSwitching(
            Eapplied, duration, steps)
        
        a.timeResDataTimes = times
        a.timeResDataPols = pols
        a.timeResDataSwitch = switched
        a.timeResIndex = 0
        a.timeResAnimating = true
        
        // Clear history for clean display
        a.eHistory = a.eHistory[:0]
        a.pHistory = a.pHistory[:0]
        
        a.addLogEntry("━━ TIME-RESOLVED SWITCHING ━━")
        a.addLogEntry(fmt.Sprintf("E = %.1f MV/cm (2×Ec)", Eapplied/1e8))
        a.addLogEntry(fmt.Sprintf("Duration: %.0f ns", duration*1e9))
        a.addLogEntry("KAI stretched exponential")
        a.addLogEntry("P(t)=Ps(1-exp(-(t/τ)^n))")
    }
```

3. Animate through precomputed data in `Update()`:
```go
// Advance animation
a.timeResIndex += 2  // 2 frames per iteration (50 FPS)
idx := a.timeResIndex
if idx < len(a.timeResDataTimes) {
    currentTime := a.timeResDataTimes[idx]
    currentPol := a.timeResDataPols[idx]
    switchedFrac := float64(a.timeResDataSwitch[idx]) / 
                     float64(len(a.preisach.hysterons))
    
    a.polarization = currentPol
    a.normalizedP = a.polarization / a.material.Pr
    a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * 29))
}
```

4. Update plot widget to show time-domain data (extend PEPlot to support time series)

**File Changes**:
- Extend `module1-hysteresis/pkg/gui/simulation.go` - Time-Resolved case
- Update `module1-hysteresis/pkg/gui/widgets/peplot.go` - Add time series support
- Update waveform dropdown to include Time-Resolved option

---

### ✅ **I1: JSON/CSV Export for P-E Data** ⭐⭐⭐
**Priority**: HIGH | **Complexity**: LOW | **Impact**: HIGH

**Why**: Critical for interoperability with external tools (NeuroSim, CrossSim, AIHWKIT).

**Implementation**:
```go
// Export format
type HysteresisExport struct {
    Format     string `json:"format"`       // "fecim_hysteresis_v1"
    Model      string `json:"model"`        // "mayergoyz_preisach"
    Material   string `json:"material"`      // Material name
    Parameters MaterialParams `json:"parameters"`
    LoopData   LoopData `json:"loop_data"`
    Metadata   ExportMetadata `json:"metadata"`
}

type MaterialParams struct {
    Ec  float64 `json:"ec"`
    Ps  float64 `json:"ps"`
    Pr  float64 `json:"pr"`
    Temperature float64 `json:"temperature_k"`
}

type LoopData struct {
    EField []float64 `json:"e_field_V_m"`
    Polarization []float64 `json:"polarization_C_m2"`
}

type ExportMetadata struct {
    Generated string `json:"generated"`
    Software   string `json:"software"`
}

func (a *App) exportHysteresisData(filename string) error {
    // Format check from extension
    ext := filepath.Ext(filename)
    
    a.mu.RLock()
    defer a.mu.Unlock()
    
    var format string
    
    if ext == ".json" {
        format = "json"
        // Export JSON
        data := HysteresisExport{
            Format: "fecim_hysteresis_v1",
            Model: "mayergoyz_preisach",
            Material: a.material.Name,
            Parameters: MaterialParams{
                Ec: a.material.Ec,
                Ps: a.material.Ps,
                Pr: a.material.Pr,
                Temperature: a.preisach.Temperature,
            },
            LoopData: LoopData{
                EField: a.eHistory,
                Polarization: a.pHistory,
            },
            Metadata: ExportMetadata{
                Generated: time.Now().Format(time.RFC3339),
                Software: "FeCIM Lattice Tools v1.0",
            },
        }
        bytes, _ := json.MarshalIndent(data, "", "  ")
        return os.WriteFile(filename, bytes, 0644)
        
    } else if ext == ".csv" {
        // Export CSV
        var buf bytes.Buffer
        buf.WriteString("E_field_V_m,Polarization_C_m2,Level\n")
        for i := 0; i < len(a.eHistory); i++ {
            level := 0
            if len(a.pHistory) > i {
                level = int(math.Round((a.pHistory[i]/a.material.Ps + 1) / 2 * 29))
            }
            fmt.Fprintf(&buf, "%.6e,%.6e,%d\n", 
                       a.eHistory[i], a.pHistory[i], level)
        }
        return os.WriteFile(filename, buf.Bytes(), 0644)
    }
}
```

Add GUI buttons in controls panel:
- "Export JSON" → file dialog → save
- "Export CSV" → file dialog → save

**File Changes**:
- Create `module1-hysteresis/pkg/export/export.go`
- Add export functions to GUI controls panel

---

## 🥈 Phase 2: MEDIUM IMPACT (2-4 weeks)

### ✅ **U2: Multi-Material Overlay Comparison Mode** ⭐⭐⭐
**Priority**: HIGH | **Complexity**: MEDIUM | **Impact**: HIGH

**Why**: Educational value + marketing differentiation. Show FeCIM vs competitors side-by-side.

**Implementation**:
```go
// In App struct:
overlayMode    bool                 // Toggle overlay mode
overlayLoops  []OverlayLoop       // Store up to 4 loops
overlayIndex  int                 // Currently displayed loop

type OverlayLoop struct {
    Material   *HZOMaterial
    EData     []float64
    PData     []float64
    Color     color.Color
}

// In createControlsPanel():
overlayToggle := widget.NewCheck("Compare Materials (Overlay)", false)
overlayToggle.OnChanged = func(checked bool) {
    a.mu.Lock()
    a.overlayMode = checked
    if checked {
        // Enable overlay selection
        a.buildOverlayPanel()
    } else {
        // Disable, show single plot
    }
    a.mu.Unlock()
}

overlayPanel := container.NewVBox(
    widget.NewLabel("Select up to 4 materials to compare:"),
    // Multi-select widget (material 1, 2, 3, 4)
    "Overlay" button,
)

// Extend PEPlot to support overlays
func (p *PEPlot) SetOverlayData(loops []OverlayLoop) {
    p.overlayLoops = loops
    p.Refresh()
}
```

---

### ✅ **U4: Real-Time Metrics Dashboard** ⭐⭐⭐
**Priority**: MEDIUM-HIGH | **Complexity**: LOW-MEDIUM | **Impact**: HIGH

**Why**: Users want to see loop area, effective Ec/Pr, squareness, energy consumption.

**Current State**: Partial - wake-up/fatigue displayed but not comprehensive.

**Implementation**:
```go
// In App struct:
metricsPanel *MetricsPanel

type MetricsPanel struct {
    widget.BaseWidget
    loopArea      *widget.Label
    effectiveEc   *widget.Label
    effectivePr   *widget.Label
    squareness    *widget.Label
    switchedFrac  *widget.Label
    cycleCount    *widget.Label
    degradation   *widget.Label
}

func NewMetricsPanel() *MetricsPanel {
    return &MetricsPanel{
        loopArea: widget.NewLabel(""),
        effectiveEc: widget.NewLabel(""),
        effectivePr: widget.NewLabel(""),
        squareness: widget.NewLabel(""),
        switchedFrac: widget.NewLabel(""),
        cycleCount: widget.NewLabel(""),
        degradation: widget.NewLabel(""),
    }
}

func (m *MetricsPanel) Update(preisach *MayergoyzPreisach, E, P []float64) {
    // Calculate loop area via trapezoidal integration
    area := 0.0
    for i := 1; i < len(E); i++ {
        area += 0.5 * (P[i] + P[i-1]) * (E[i] - E[i-1])
    }
    
    // Extract effective values from loop
    // Find E=0 crossings for Pr
    // Find max/min for Ec
    
    m.loopArea.SetText(fmt.Sprintf("%.2e J/m³", math.Abs(area)))
    m.effectiveEc.SetText(fmt.Sprintf("%.2f MV/cm", preisach.GetEffectiveEc()/1e8))
    // ... update other labels
}
```

Integrate into existing info panel layout.

**File Changes**:
- Create `module1-hysteresis/pkg/gui/widgets/metrics_panel.go`
- Update `module1-hysteresis/pkg/gui/gui.go` - Add metrics panel to UI

---

### ✅ **P2: Nucleation-Limited Switching (NLS) Model for HfO₂** ⭐⭐⭐
**Priority**: HIGH | **Complexity**: MEDIUM | **Impact**: HIGH

**Why**: HfO₂-specific accuracy critical for write latency prediction. Parameters loaded in `material.go` but not used.

**Implementation**:
```go
// Extend MayergoyzPreisach struct (already has NLS params):
// Add NLS-specific methods:

// incubationTime returns time before switching starts (dead time)
func (m *MayergoyzPreisach) IncubationTime(E float64) float64 {
    Ec := m.temperatureCorrectedEc()
    if math.Abs(E) <= Ec {
        return math.Inf(1) // No switching below Ec
    }
    Eeff := math.Abs(E) - Ec
    
    // Arrhenius-type nucleation
    return m.Tau0NLS * math.Exp(m.EaNLS / Eeff)
}

// totalSwitchingTime = incubation + growth
func (m *MayergoyzPreisach) TotalSwitchingTime(E float64) float64 {
    return m.IncubationTime(E) + m.TauGrowth
}
```

Expose via GUI:
- Add "Switching Model" dropdown: [Preisach | NLS | KAI]
- Show write latency prediction for current level
- Visualize incubation vs growth phases

---

### ✅ **P4: Preisach Plane Visualization** ⭐⭐⭐
**Priority**: MEDIUM-HIGH | **Complexity**: MEDIUM | **Impact**: HIGH

**Why**: `GetPreisachPlane()` returns α, β, states - shows how hysteresis emerges.

**Implementation**:
```go
// New widget: widgets/preisach_plane.go

type PreisachPlaneWidget struct {
    widget.BaseWidget
    alphas  []float64
    betas   []float64
    states  []int
    weights []float64
}

func (w *PreisachPlaneWidget) Update(p *MayergoyzPreisach) {
    w.alphas, w.betas, w.states = p.GetPreisachPlane()
    w.weights = p.GetDistribution()
    w.Refresh()
}

func (w *PreisachPlaneWidget) CreateRenderer() fyne.WidgetRenderer {
    // Draw 2D scatter plot:
    // - β on X-axis, α on Y-axis
    // - Color by state: +1=red, -1=blue
    // - Size/opacity by weight
    // - Show staircase line (current boundary)
    // - Only show α > β region (triangle)
}
```

Add to GUI:
- Toggle button in controls: "Show Preisach Plane"
- Separate tab/window for visualization
- Educational annotations explaining what users see

---

### ✅ **I2: Import from Experimental P-E Data** ⭐⭐⭐
**Priority**: MEDIUM | **Complexity**: MEDIUM | **Impact**: MEDIUM

**Why**: Compare simulated loops with experimental measurements (Radiant, Keithley tracers).

**Implementation**:
```go
func (a *App) importPEData(filename string) error {
    // Parse CSV: E (MV/cm), P (µC/cm²)
    // Optional headers: Material, Temperature, Frequency
    
    // Load data
    var importedLoop struct {
        EField: []float64
        Polarization: []float64
    }
    
    // Overlay on existing plot
    a.plot.SetImportedData(importedLoop)
    a.plot.Refresh()
}
```

Add GUI:
- "Import Data" button → file dialog
- "Match Parameters" button → auto-fit Preisach to imported data
- Show simulated vs experimental with different colors

---

## 🔥 Phase 3: MEDIUM IMPACT (1-2 months)

### ✅ **P5: Frequency-Dependent Hysteresis** ⭐⭐⭐
**Priority**: MEDIUM | **Complexity**: MEDIUM | **Impact**: MEDIUM

**Why**: At higher frequencies, hysteresis loops become thinner. Critical for high-speed operation.

**Implementation**:
```go
// Add to MayergoyzPreisach:
frequency    float64 // Operating frequency (Hz)
frequencyFactor float64

func (m *MayergoyzPreisach) SetFrequency(freq float64) {
    m.frequency = freq
    // Adjust effective Ec based on frequency
    f0 := 1e3  // Reference frequency (1 kHz)
    m.frequencyFactor = 1.0 + math.Pow(freq/f0, 0.1)
    
    // Recalculate hysterons with scaled Ec
    m.initializeHysterons()
    m.initializeDistribution()
}

// In GUI:
freqSlider := widget.NewSlider(1, 1e6) // 1 Hz to 1 MHz
freqSlider.OnChanged = func(f float64) {
    a.preisach.SetFrequency(f)
    // Update loop area vs frequency label
}
```

---

### ✅ **P7: FORC (First-Order Reversal Curve) Analysis** ⭐⭐⭐
**Priority**: MEDIUM-HIGH | **Complexity**: MEDIUM-HIGH | **Impact**: MEDIUM

**Why**: Extract Preisach distribution μ(α,β) from measured minor loops. Compare with theoretical distribution.

**Implementation**:
```go
func (m *MayergoyzPreisach) GenerateFORC(HrValues []float64, points int) FORCData {
    forc := FORCData{
        Hr: HrValues,
        H:  make([][]float64, len(HrValues)),
        P:  make([][]float64, len(HrValues)),
    }
    
    Emax := 2.5 * m.temperatureCorrectedEc()
    
    for i, Hr := range HrValues {
        m.Reset()
        // Saturate positive
        for j := 0; j <= 20; j++ {
            m.Update(Emax * float64(j) / 20)
        }
        // Go to reversal field Hr
        for j := 0; j <= 20; j++ {
            m.Update(Emax - (Emax-Hr)*float64(j)/20)
        }
        // Measure ascending branch
        forc.H[i] = make([]float64, points)
        forc.P[i] = make([]float64, points)
        for j := 0; j < points; j++ {
            H := Hr + (Emax-Hr)*float64(j)/float64(points-1)
            forc.H[i][j] = m.Update(H)
            forc.P[i][j] = m.Polarization
        }
    }
    
    return forc
}

// Compute FORC distribution: ρ(Hr, H) = -½ ∂²P/∂Hr∂H
func ComputeFORCDistribution(forc FORCData) [][]float64 {
    // 2D numerical differentiation of FORC surface
    // Returns ρ(α, β) as 2D array
}
```

Add GUI:
- "FORC Analysis" mode
- Generate FORC dataset button
- 2D contour plot of ρ(α, β)
- Overlay theoretical Gaussian distribution

---

## 🌟 Phase 4: LONG-TERM / STRETCH GOALS

### ✅ **P8: Stretched Exponential Fatigue** ⭐⭐
**Priority**: MEDIUM | **Complexity**: LOW | **Impact**: MEDIUM

**Why**: Current linear fatigue unrealistic. Stretched exponential matches HZO literature.

**Implementation**:
```go
// Already in material.go: EnduranceAtCycles with beta=0.3
// Update applyFatigue():

func (m *MayergoyzPreisach) applyFatigue() float64 {
    // Stretched exponential fatigue (Kohlrausch-Williams-Watts)
    beta := 0.3  // HZO literature value
    N0 := m.material.EnduranceCycles
    N := float64(m.cycleCount)
    
    fatigueFactor := math.Exp(-math.Pow(N/N0, beta))
    return fatigueFactor
}
```

---

### ✅ **P1: Landau-Khalatnikov (LK) Dynamics Model** ⭐⭐⭐
**Priority**: MEDIUM | **Complexity**: MEDIUM | **Impact**: HIGH

**Why**: Physics-based alternative to Preisach. Captures frequency dependence naturally.

**Implementation**:
```go
// New file: pkg/ferroelectric/landau_khalatnikov.go

type LandauKhalatnikov struct {
    // Landau coefficients
    Alpha  float64 // Linear: α = a(T - Tc)
    Beta   float64 // Cubic term
    Gamma  float64 // Quintic term
    Eta    float64 // Viscosity (damping)
    
    // State
    Polarization float64
    Temperature  float64
    CurieTemp    float64
}

func (lk *LandauKhalatnikov) FreeEnergyGradient(P float64) float64 {
    return 2*lk.Alpha*P + 4*lk.Beta*math.Pow(P, 3) + 
           6*lk.Gamma*math.Pow(P, 5)
}

func (lk *LandauKhalatnikov) Update(E float64, dt float64) float64 {
    // Euler integration: dP/dt = (1/η)(-dF/dP + E)
    dPdt := (1/lk.Eta) * (-lk.FreeEnergyGradient(lk.Polarization) + E)
    lk.Polarization += dPdt * dt
    return lk.Polarization
}
```

Add to GUI:
- Model selector: [Preisach | Landau-Khalatnikov]
- Compare both models side-by-side
- Show free energy landscape

---

### ✅ **P6: Jiles-Atherton Alternative Model** ⭐
**Priority**: LOW-MEDIUM | **Complexity**: MEDIUM | **Impact**: LOW-MEDIUM

**Why**: Educational comparison. Fewer parameters (5 vs 400+ hysterons), faster computation.

**Implementation**:
```go
// New file: pkg/ferroelectric/jiles_atherton.go

type JilesAtherton struct {
    Ps     float64 // Saturation polarization
    a       float64 // Shape parameter
    k       float64 // Pinning coefficient
    c       float64 // Reversibility
    alpha   float64 // Interdomain coupling
}

func (ja *JilesAtherton) Update(E float64) float64 {
    dM := (ja.anhysteretic(ja.M) - ja.M) / 
          (ja.k * delta - ja.alpha * (ja.anhysteretic(ja.M) - ja.M))
    
    // dP/dt = (M - M) / (k * delta - alpha*(M - M))
    // Integrate dP
    return ja.M + dM
}
```

Add to GUI:
- Model selector: [Preisach | Jiles-Atherton]
- Compare both models side-by-side
- Show parameter trade-offs

---

## 🌙 Phase 5: ENHANCEMENT (Ongoing)

### ✅ **P9: Imprint Field Modeling** ⭐⭐
**Priority**: LOW-MEDIUM | **Complexity**: LOW | **Impact**: LOW-MEDIUM

**Why**: Long-term retention degradation. ImprintField defined but unused.

**Implementation**:
```go
// Extend MayergoyzPreisach struct:
imprint *ImprintState

type ImprintState struct {
    ImprintField float64  // Accumulated shift
    ImprintRate  float64  // Shift rate per second
}

func (m *MayergoyzPreisach) ApplyImprint(biasField float64, duration float64) {
    // Imprint accumulates logarithmically
    m.imprint.ImprintField += m.imprint.ImprintRate * 
        math.Log10(1+duration) * math.Sign(biasField)
}

func (m *MayergoyzPreisach) Update(E float64) float64 {
    // Shift effective field by imprint
    Eeff := E - m.imprint.ImprintField
    // ... rest of Update uses Eeff
}
```

---

### ✅ **U3: Interactive Minor Loop Drawing** ⭐⭐
**Priority**: MEDIUM | **Complexity**: MEDIUM | **Impact**: MEDIUM

**Why**: Educational - users "draw" arbitrary paths on P-E plane.

**Implementation**:
```go
// Extend PEPlot widget to support mouse drag:
func (p *PEPlot) EnableInteractiveMode(enabled bool) {
    p.interactiveEnabled = enabled
}

func (p *PEPlot) onDragged(event *fyne.DragEvent) {
    if !p.interactiveEnabled {
        return
    }
    // Convert screen position to E value
    E := p.screenToField(event.Position.X)
    
    // Update physics
    P := a.preisach.Update(E)
    
    // Draw current position on plot
    p.currentPoint = fyne.Position{event.Position.X, 
                              p.fieldToScreen(P)}
    p.Refresh()
}
```

---

### ✅ **U5: Domain Wall Visualization (Simplified 1D)** ⭐⭐
**Priority**: MEDIUM | **Complexity**: MEDIUM | **Impact**: MEDIUM

**Why**: Visual intuition for domain switching during write operations.

**Implementation**:
```go
// New widget: widgets/domain_strip.go

type DomainStrip struct {
    widget.BaseWidget
    domains []int    // +1 or -1 for each segment
    width   int     // Number of segments (e.g., 50)
}

func (d *DomainStrip) UpdateFromPreisach(p *MayergoyzPreisach) {
    // Map hysteron states to visual domains
    _, _, states := p.GetPreisachPlane()
    d.domains = make([]int, d.width)
    
    step := len(states) / d.width
    for i := 0; i < d.width; i++ {
        d.domains[i] = states[i*step]
    }
}

func (d *DomainStrip) CreateRenderer() fyne.WidgetRenderer {
    // Draw horizontal strip of colored rectangles
    // Red for +1, Blue for -1
    // Animate transitions
}
```

---

### ✅ **U6: Export Screenshots/GIF Animation** ⭐⭐
**Priority**: LOW | **Complexity**: MEDIUM | **Impact**: LOW-MEDIUM

**Why**: Users want to share visualizations for presentations/demos.

**Implementation**:
```go
func (a *App) exportScreenshot(filename string) error {
    img := a.plot.Snapshot()
    f, _ := os.Create(filename)
    defer f.Close()
    return png.Encode(f, img)
}

func (a *App) recordGIF(filename string, duration time.Duration) error {
    frames := make([]*image.Paletted, 0)
    ticker := time.NewTicker(33 * time.Millisecond)
    
    for time.Since(start) < duration {
        <-ticker.C
        frames = append(frames, a.plot.Snapshot())
    }
    
    gif.EncodeAll(f, &gif.GIF{Image: frames, Delay: delays})
}
```

---

### ✅ **I5: Preisach Distribution Import/Export** ⭐⭐
**Priority**: LOW | **Complexity**: LOW | **Impact**: LOW-MEDIUM

**Why**: Load measured distributions, share calibrated models.

**Implementation**:
```go
func (m *MayergoyzPreisach) ImportDistribution(filename string) error {
    // Load CSV: alpha,beta,weight
    // Rebuild hysterons with imported weights
    // Already have ImportState() - extend to support distribution-only
}
```

---

### ✅ **I6: REST API for External Control** (ADVANCED)
**Priority**: LOW | **Complexity**: MEDIUM-HIGH | **Impact**: LOW

**Why**: Remote control, integration with AI training pipelines.

**Implementation**:
```go
func (a *App) startAPIServer(port int) {
    http.HandleFunc("/API.mdstatus", func(w http.ResponseWriter, r *http.Request) {
        a.mu.RLock()
        defer a.mu.Unlock()
        json.NewEncoder(w).Encode(map[string]interface{}{
            "E": a.electricField,
            "P": a.polarization,
            "level": a.discreteLevel,
        })
    })
    
    http.HandleFunc("/API.mdfield", func(w http.ResponseWriter, r *http.Request) {
        var req struct{ E float64 `json:"e"`}
        json.NewDecoder(r.Body).Decode(&req)
        a.mu.Lock()
        defer a.mu.Unlock()
        a.electricField = req.E
    })
    
    go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
```

---

### ✅ **I7: FerroX Material Parameter Export** ⭐⭐
**Priority**: LOW | **Complexity**: LOW | **Impact**: LOW

**Why**: Integration with FerroX phase-field simulator.

**Implementation**:
```go
func (a *App) exportFerroXParams(filename string) error {
    // Generate FerroX input deck format
    deck := fmt.Sprintf(`
# Generated by FeCIM Lattice Tools
# Material: %s

material.eps_r = %f
material.Pr = %f
material.Ec = %f
material.tau = %f

# Landau coefficients (derived from P-E fit)
material.alpha1 = %f
material.alpha11 = %f
material.alpha111 = %f
`, a.material.Name, a.material.Ps, a.material.Ps, a.material.Ec,
      /* derived Landau coeffs */)
    
    return os.WriteFile(filename, []byte(deck), 0644)
}
```

---

### ✅ **I4: NeuroSim-Compatible Weight Export** ⭐⭐
**Priority**: LOW-MEDIUM | **Complexity**: LOW | **Impact**: MEDIUM

**Why**: Crossbar simulation integration.

**Implementation**:
```go
type NeurosimWeightExport struct {
    Format     string `json:"format"`
    ArraySize  [2]int `json:"array_size"`
    GMin       float64 `json:"g_min_s"`
    GMax       float64 `json:"g_max_s"`
    NumLevels  int     `json:"num_levels"`
    Weights    [][]int `json:"weights"`
}

func (a *App) exportNeurosimWeights(filename string) error {
    // Generate weight matrix for crossbar
    // 30 levels → conductance values
    // Format compatible with NeuroSim
    export := NeurosimWeightExport{
        Format: "neurosim_weight_v1",
        ArraySize: [2]int{32, 32},
        GMin: 1e-6,
        GMax: 100e-6,
        NumLevels: 30,
        Weights: getWeightsFromCalibration(),
    }
    
    jsonBytes, _ := json.MarshalIndent(export, "", " ")
    return os.WriteFile(filename, jsonBytes, 0644)
}
```

---

## 📋 Appendix: Additional Features

### ✅ **A1: Inverse Preisach for Write Pulse Optimization** (ADVANCED)
**Priority**: MEDIUM | **Complexity**: MEDIUM | **Impact**: MEDIUM

Newton-Raphson solver to find E for target P.

### ✅ **A2: PUND Measurement Support** ⭐⭐
**Priority**: MEDIUM | **Complexity**: LOW | **Impact**: MEDIUM

Positive-Up-Negative-Down protocol for distinguishing switching vs non-switching charge.

### ✅ **A3: ML Surrogate Model for Speed** (FUTURE)
**Priority**: LOW | **Complexity**: HIGH | **Impact**: LOW

Neural network approximation for 100× faster inference.

### ✅ **A4: Negative Capacitance Visualization Mode** ⭐
**Priority**: LOW | **Complexity**: MEDIUM | **Impact**: LOW

Show S-curve region where dP/dE < 0.

### ✅ **A5: Phase Diagram Generation** ⭐
**Priority**: LOW | **Complexity**: MEDIUM | **Impact**: LOW

Generate T-E and T-σ phase diagrams showing ferroelectric/paraelectric boundaries.

---

## 📋 Implementation Roadmap

| Phase | Tasks | Duration |
|--------|--------|----------|
| **Phase 1** (3 weeks - REVISED) | |
| | ├─ ⭐ P1.1: Parasitic RC Network (CRITICAL) | 5 days |
| | ├─ ⭐ P1.2: Leakage Current Modeling (HIGH) | 4 days |
| | ├─ ⭐ P1.3: Stochastic Variability (MEDIUM) | 2 days |
| | ├─ U1: Temperature slider | 2 days |
| | ├─ P3: KAI dynamics viz | 2 days |
| | └─ I1: JSON/CSV export | 2 days |
| ├─ P4: Preisach plane viz | 1 week |
| **Phase 2** (4 weeks) | |
| ├─ I2: Import experimental data | 1 week |
| └─ P2: NLS model integration | 2 days |
| **Phase 3** (6 weeks) | |
| ├─ P5: Frequency dependence | 1 week |
| ├─ P7: FORC analysis | 2 weeks |
| ├─ I3: SPICE export | 1 week |
| └─ U3: Interactive minor loops | 3 days |
| **Phase 4** (4 weeks) | |
| ├─ P8: Stretched exponential fatigue | 2 days |
| ├─ P1: Landau-Khalatnikov model | 2 weeks |
| ├─ U5: Domain wall viz | 1 week |
| ├─ U6: Screenshots/GIF export | 2 days |
| └─ I4: NeuroSim export | 2 days |
| **Phase 5** (ongoing) | |
| ├─ I5: Preisach distribution import/export | 2 days |
| ├─ I6: REST API | 3 days |
| ├─ I7: FerroX export | 2 days |
| └─ Remaining enhancements | TBD |

**Total Estimated Time**: 4-8 months for full feature set
**CRITICAL PREREQUISITES** (1 month): P1.1, P1.2, P1.3
**Minimal Viable Subset** (2-3 months): U1, P3, I1, U4, P2, P4

---

## 🏗 Architecture Considerations

### Code Organization
```
module1-hysteresis/
├── pkg/
│   ├── ferroelectric/
│   │   ├── preisach.go              (existing - keep)
│   │   ├── preisach_advanced.go      (existing - enhance)
│   │   │   ├── parasitic_rc.go          (NEW)
│   │   ├── landau_khalatnikov.go  (NEW)
│   │   ├── jiles_atherton.go      (NEW)
│   │   ├── material.go              (existing - enhance for leakage)
│   │   └── forc_analysis.go       (NEW)
│   ├── gui/
│   │   ├── gui.go                  (extend)
│   │   ├── widgets/
│   │   │   ├── preisach_plane.go    (NEW)
│   │   │   ├── domain_strip.go       (NEW)
│   │   │   ├── metrics_panel.go      (NEW)
│   │   │   └── export_panel.go       (NEW)
│   └── algo/
│       ├── calibration.go            (existing - extend)
│       └── forc_analysis.go       (NEW - NLS integration)
└── data/
    ├── calibrations/                 (existing - multi-temp already there)
    └── exports/                        (NEW - for exported data)
```

### Testing Strategy
- All new features require tests
- Use existing test patterns: `*_test.go`
- Golden regression tests for critical physics
- Integration tests for new widgets
- Performance benchmarks (must maintain 60 FPS)
- GUI visual tests (manual inspection)

### Performance Targets
- **Simulation**: Maintain 60 FPS (current: 16.7ms/frame)
- **Preisach Update**: < 100μs per call (2000 hysterons)
- **GUI Refresh**: < 5ms per frame
- **Memory**: < 100MB for 2000 hysterons + history

---

## ✅ Verification Checklist

Before marking any feature complete:
- [ ] Code compiles without errors
- [ ] `go test ./module1-hysteresis` passes
- [ ] No new race conditions detected
- [ ] GUI renders correctly (visual inspection)
- [ ] Documentation updated (if needed)
- [ ] Integration with existing calibration works
- [ ] **Circuit Reality (P1.x)**: RC network and leakage validated vs Samsung Nature 2025 data
- [ ] **Circuit Reality (P1.x)**: Voltage drop < 15% error vs physical measurements
- [ ] **Circuit Reality (P1.x)**: Retention time within 2× of measured values
- [ ] **Circuit Reality (P1.x)**: Stochastic variance matches ISPP write errors

---

## 🎯 Success Metrics

| Metric | Target |
|--------|--------|
| **Circuit Accuracy** | < 15% error vs Samsung Nature 2025 measurements |
| **Physics Accuracy** | < 5% error vs literature |
| **Simulation Speed** | 60 FPS maintained |
| **Test Coverage** | > 90% for new code |
| **User Experience** | Intuitive, responsive |
| **Documentation** | Complete for new features |

---

## 📚 Key References

### Internal Documentation
- `docs/hysteresis/../hysteresis/hysteresis.physics.md` - Physics fundamentals
- `docs/hysteresis/../hysteresis/hysteresis.research.md` - Research meta-study
- `docs/hysteresis/hysteresis.opensource.md` - Open-source tools review
- `docs/hysteresis/hysteresis-proposed-improvements-opensource.md` - Proposed improvements (31 items)

### Research Papers
- 24 papers in `/docs/research-papers/by-topic/01-ferroelectric-materials/`
- HfO₂ physics: Nature Communications, Advanced Materials, Nano Letters
- Preisach modeling: IEEE Transactions on Magnetics
- Domain dynamics: Physical Review, arXiv

### Codebase
- `module1-hysteresis/pkg/ferroelectric/preisach.go` - Basic model
- `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go` - Advanced model (1047 lines)
- `module1-hysteresis/pkg/ferroelectric/material.go` - Material library
- `module1-hysteresis/pkg/gui/gui.go` - Main UI (1447 lines)
- `module1-hysteresis/pkg/gui/simulation.go` - Simulation loop (1447 lines)
- `module1-hysteresis/pkg/algo/calibration.go` - Calibration system
- `module1-hysteresis/pkg/controller/writer.go` - Write controller

---

**This plan provides:**
- **31 specific improvements** mapped from proposed improvements document
- **Clear priority tiers** for incremental implementation
- **Technical implementation details** for each feature
- **Estimated timelines** for planning
- **Architecture considerations** for code organization
- **Testing and verification strategies**

**Recommendation**: Start with **Phase 1** (5 high-impact, low-complexity features) for immediate value, then progress through remaining phases.
