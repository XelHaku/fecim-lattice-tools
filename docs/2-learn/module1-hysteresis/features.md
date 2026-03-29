# Module 1: Hysteresis - Features

**Navigation:** [← Module 1 Index](./README.md) | [Physics](./physics.md) | [Run Modes](./run-modes.md) | [Tools](./tools.md)

---

## Evidence Status

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

---

## What This Module Does

Module 1 simulates ferroelectric hysteresis loops using two physics engines:

1. **Preisach Model** (default) - Phenomenological hysteron-based approach
2. **Landau-Khalatnikov** (optional) - Thermodynamic mean-field model

It provides:
- Real-time P-E curve visualization
- Discrete polarization levels for memory-state modeling (30-level baseline)
- GUI visualization with multiple waveform modes
- Simulated calibration and write/read/verify flows
- Multi-platform support (GUI, TUI, headless, Vulkan)

---

## Primary Components

### Physics Engines
- `module1-hysteresis/pkg/ferroelectric/preisach.go` - Tanh-based Preisach
- `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go` - Mayergoyz Preisach
- `shared/physics/landau.go` - L-K solver
- `shared/physics/preisach.go` - Stack engine for turning points

### GUI/Visualization
- `module1-hysteresis/pkg/gui/gui.go` - Fyne GUI initialization
- `module1-hysteresis/pkg/gui/physics_engine.go` - Engine selector
- `module1-hysteresis/pkg/gui/simulation.go` - Real-time update loop
- `module1-hysteresis/pkg/render/plot.go` - P-E curve rendering

### Controller
- `module1-hysteresis/pkg/controller/writer.go` - ISPP write controller
- `module1-hysteresis/pkg/controller/state_machine.go` - Phase machine
- `module1-hysteresis/pkg/controller/bounds.go` - Binary search logic

### Materials
- `module1-hysteresis/pkg/ferroelectric/material.go` - Material presets
- `shared/physics/material.go` - Shared definitions

### Alternative Interfaces
- `module1-hysteresis/pkg/tui/tui.go` - Terminal UI (Bubble Tea)
- `module1-hysteresis/pkg/render/` - Vulkan GPU renderer (optional)

**See also:** [RUN_MODES.md](./run-modes.md) for GUI/TUI/headless/Vulkan behavior and engine defaults.

---

## Key Workflows

### 1. Interactive Loop Exploration (Manual Mode)

```
User drags E-field slider → Physics engine updates P → Plot redraws
                              ↓
                    Hysteron states update
                              ↓
                    Discrete level calculated
                              ↓
                    UI shows: P, E, Level, Mode
```

**Use case:** Explore hysteresis behavior, understand coercive field

### 2. Waveform Cycling

```
Select waveform → Auto-cycle E-field → Trace complete loops
   (sine, triangle, square, random)
```

**Waveforms:**
- **Sine:** Smooth full-loop traversal
- **Triangle:** Linear ramps showing switching thresholds
- **Square:** Fast switching between extremes
- **Random Walk:** Multi-level programming demonstration

### 3. Write/Read/Verify Flow

```
WRITE Phase:
  ├─ Apply pulse (ISPP controller)
  ├─ Verify readback level
  ├─ Adjust next pulse (binary search)
  └─ Repeat until target reached

HOLD Phase:
  └─ E = 0, observe P retention (non-volatility)

READ Phase:
  ├─ Apply sense field (E < Ec)
  └─ Report level without disturbing state

DISPLAY Phase:
  └─ Show results in log
```

**Use case:** Demonstrate memory operations, study programming convergence

### 4. Temperature Sweep

```
Set temperature → Calibrate Ec(T), Pr(T) → Update loop shape
     ↓
Cache calibration for reuse
```

**Use case:** Study temperature dependence, validate thermal stability

### 5. Material Comparison

```
Load material preset → Compare loop shapes side-by-side
  (HZO, superlattice, AlScN, cryogenic, etc.)
```

**Use case:** Understand how Ec, Pr, Ps affect loop characteristics

---

## Feature Matrix

### Physics Models

| Feature | Preisach | Landau-Khalatnikov |
|---------|----------|-------------------|
| Hysteresis loop | ✅ Hysteron-based | ✅ Free energy minimum |
| Minor loops | ✅ Turning point stack | ✅ Dynamics solver |
| Temperature | ✅ Linear scaling | ✅ Curie transition |
| Switching time | ✅ Defined (not used in viz) | ✅ KAI dynamics |
| Fatigue | ✅ Per-hysteron degradation | ❌ Not implemented |
| Wake-up | ✅ Gradual Pr increase | ❌ Not implemented |

### Visualization Modes

| Mode | GUI | TUI | Headless | Vulkan |
|------|-----|-----|----------|--------|
| Real-time plot | ✅ | ✅ | ❌ | ✅ |
| ASCII rendering | ❌ | ✅ | ✅ | ❌ |
| Interactive controls | ✅ | ✅ | ❌ | ❌ |
| Export data | ✅ | ✅ | ✅ | ✅ |
| Hardware acceleration | ❌ | ❌ | ❌ | ✅ (GPU) |

### Material Library

| Material | Pr (µC/cm²) | Ec (MV/cm) | τ (ns) | Tc (°C) |
|----------|-------------|------------|--------|---------|
| Standard HZO | 25 | 1.2 | 1 | 450 |
| FeCIM HZO | 30 | 1.0 | 10 | 450 |
| Superlattice | 22 | 0.85 | 0.36 | 500 |
| Cryogenic HZO | 75 | 1.5 | 1 | 450 |
| AlScN | 120 | 5.0 | 10 | 1000 |
| HZO-32 | 20 | 1.0 | 10 | 450 |
| FTJ-140 | 25 | 1.2 | 20 | 450 |

See [materials.md](./materials.md) for complete parameter tables.

### Programming Features

| Feature | Status | Notes |
|---------|--------|-------|
| ISPP write controller | ✅ | Binary search with overshoot recovery |
| Binary search bounds | ✅ | Adaptive [VMin, VMax] adjustment |
| Guard-band correction | ✅ | Max 2 pulses for boundary states |
| Overshoot detection | ✅ | Triggers bounds widening |
| ACCEPT ±1 logic | ✅ | After 8+ overshoots (physics-limited) |
| Reset shortcut | ✅ | Fast path when overshoot detected |
| Verification logging | ✅ | Per-target metrics in UI |
| DCC (Direct Cell Control) | ✅ | Single-pulse write, `shared/physics/dcc_write.go` |

### World-Class Physics Features

These features are implemented in `shared/physics/worldclass_*.go` and provide research-grade characterization:

| Feature | Package | Description |
|---------|---------|-------------|
| **PUND measurement** | `worldclass_pund.go` | Separate switching from non-switching charge via P/U/N/D pulse sequence; uses trapezoidal current integration |
| **Retention model** | `worldclass_retention.go` | Power-law decay P(t) = P₀(t/t₀)^(-β); β ≈ 0.01–0.05 for HZO; also exponential model for comparison |
| **Wake-up + Fatigue** | `worldclass_wakeup.go` | Two-phase Pr(N) model: wake-up (exponential rise) then fatigue (exponential decay after onset threshold) |
| **FORC density** | `worldclass_forc.go` | First-order reversal curve sweeps with Preisach density extraction via -½ d²P/(dEa dEb) |
| **FORC raster GUI** | `module1-hysteresis/pkg/gui/forc_panel.go` | Interactive hot-colormap canvas.Raster visualization |
| **C2C variation** | `worldclass_c2c.go` | State-dependent cycle-to-cycle noise: σ_G = k_abs × G_ref/G (higher noise at low G); calibrated for HZO FeFET |
| **CV characterization** | `worldclass_cv.go` | Capacitance-voltage sweep simulation |
| **Frequency dispersion** | `worldclass_frequency_dispersion.go` | Pr and Ec vs measurement frequency |

**Key literature basis for world-class features:**
- PUND: Standard characterization protocol (Merz 1954, Scott 2000)
- Wake-up/Fatigue: Park et al., APL 2013; Yurchuk et al., IEEE TED 2014
- C2C scaling: IEEE EDL 2023 (state-dependent 1/G scaling)
- FORC density: Mayergoyz (1986), Roberts et al., J Appl. Phys. 2000

---

## Extension Points

### Adding New Materials

1. Define parameters in `material.go`:
```go
FeCIMMaterial{
    Name: "MyMaterial",
    Pr:   0.35,  // C/m²
    Ps:   0.40,
    Ec:   1.5e8, // V/m
    Tau:  5e-9,  // s
    // ... other params
}
```

2. Add to material selector dropdown in GUI

### Custom Hysteron Distributions

Extend `preisach_advanced.go`:
```go
func CustomDistribution(alpha, beta, Ec, sigma float64) float64 {
    // Your distribution function
    return weight
}
```

### Alternative Renderers

Implement `Renderer` interface:
```go
type Renderer interface {
    DrawLoop(E, P []float64)
    DrawPoint(E, P float64)
    Update()
}
```

### New Waveform Modes

Add to `waveform.go`:
```go
case "custom":
    eField = yourCustomFunction(time)
```

---

## Known Limitations

### Current Implementation

1. **Landau-Khalatnikov:** Educational, not device-calibrated
2. **Wake-up/Fatigue:** UI-level placeholders in Preisach, not physics-accurate
3. **GPU Renderer:** Vulkan path is specialized, may not work on all platforms
4. **Material Presets:** Include illustrative/unverified values; cite before external use
5. **Switching Time (τ):** Not used in real-time visualization (quasistatic approximation)
6. **FORC Calibration:** Uses tanh Everett approximation, not experimental FORC data

### Performance Considerations

- **GUI:** 60 FPS on typical hardware (depends on hysteron count)
- **Hysteron Count:** Default 900 (30×30 grid), configurable
- **Minor Loop Overhead:** Stack grows with turning points (typically < 100 points)

### Platform Support

| Platform | GUI | TUI | Headless | Vulkan |
|----------|-----|-----|----------|--------|
| Linux | ✅ | ✅ | ✅ | ✅ (with drivers) |
| macOS | ✅ | ✅ | ✅ | ❌ (Metal not supported) |
| Windows | ✅ | ✅ | ✅ | ✅ (with drivers) |

---

## Performance Tips

### For Smooth Visualization

```bash
# Reduce hysteron count for faster updates
fecim-lattice-tools hysteresis --grid-size 20  # 400 hysterons instead of 900

# Use TUI for lower overhead
fecim-lattice-tools hysteresis --tui

# Use Vulkan for GPU acceleration (if available)
fecim-lattice-tools hysteresis --vulkan
```

### For Accurate Physics

```bash
# Increase hysteron count for smoother loops
fecim-lattice-tools hysteresis --grid-size 50  # 2500 hysterons

# Use L-K for thermodynamic accuracy
fecim-lattice-tools --mode hysteresis --engine lk
```

---

## Data Export

### JSON Export

```json
{
  "material": "superlattice",
  "temperature": 300,
  "E_field": [-2e8, -1.5e8, ..., 2e8],
  "polarization": [-0.5, -0.45, ..., 0.5],
  "levels": [0, 2, ..., 29]
}
```

### CSV Export

```csv
E_field_V_m,polarization_C_m2,discrete_level
-2e8,-0.5,0
-1.5e8,-0.45,2
...
```

**Use case:** Import into MATLAB, Python, or spreadsheet for analysis

---

## Testing

### Unit Tests

```bash
# Controller tests
go test ./module1-hysteresis/pkg/controller/...

# Physics tests
go test ./module1-hysteresis/pkg/ferroelectric/...

# Ensemble tests (9 materials × 2 engines)
go test ./cmd/fecim-lattice-tools/ -run TestISPPConverges
```

### Integration Tests

```bash
# Headless ISPP validation
FECIM_MATERIAL=superlattice fecim-lattice-tools --mode hysteresis --headless

# Preisach target progression
go test ./cmd/fecim-lattice-tools/ -run TestPreisachTargetProgression
```

### Regression Tests

```bash
# Update golden data (when physics model changes)
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...

# Verify against golden
go test ./... -v
```

Golden files location: `validation/testdata/physics_regression/`

---

## Integration with Other Modules

### Module 2: Crossbar

```go
// Export conductance-polarization mapping
conductance := crossbar.PolarizationToConductance(level, Gmin, Gmax)
```

### Module 3: MNIST

```go
// Export discrete levels for weight quantization
weight := float64(level) / 29.0  // Normalize to [0, 1]
```

### Module 4: Circuits

```go
// Export Ec, Pr for peripheral design
dac.SetVoltageRange(0, 2*material.Ec*thickness)
```

---

## Troubleshooting

### GUI Won't Start

```bash
# Try TUI fallback
fecim-lattice-tools hysteresis --tui

# Or headless
fecim-lattice-tools hysteresis --headless
```

### Loop Looks Wrong

1. Check material parameters match expectations
2. Verify engine (Preisach vs L-K) is correct for use case
3. Try default material: `--material fecim_hzo`

### ISPP Not Converging

- Check target level is within [0, 29]
- Verify material Ec is reasonable (0.5 - 6 MV/cm)
- Increase max iterations if physics-limited (sharp switching materials)

### Performance Issues

- Reduce grid size: `--grid-size 20`
- Use headless mode for batch processing
- Disable real-time plotting during data export

---

## Future Enhancements (Roadmap)

- [x] FORC measurement import for calibrated Preisach distribution — implemented in `worldclass_forc.go`
- [x] Wake-up physics model — implemented in `worldclass_wakeup.go`
- [x] C2C state-dependent variation — implemented in `worldclass_c2c.go`
- [x] PUND characterization — implemented in `worldclass_pund.go`
- [x] Retention power-law model — implemented in `worldclass_retention.go`
- [ ] Domain wall dynamics visualization
- [ ] Substrate strain effects
- [ ] Multi-domain microstructure simulation
- [ ] Fatigue mechanism (oxygen vacancy migration) — placeholder only; no physics model yet
- [ ] Piezoelectric coupling
- [ ] Real-time switching dynamics (KAI model in viz loop)

---

## See Also

- **[Physics Reference](./physics.md)** - Equations and model details
- **[Run Modes](./run-modes.md)** - CLI usage and mode selection
- **[Materials](./materials.md)** - Parameter tables and selection guide
- **[Tools](./tools.md)** - Integration with external simulators

---

**Last Updated:** 2026-02-16
