# Module 2 Crossbar Physics Models

This document describes the physics models implemented in module2-crossbar, covering conductance models, matrix operations, non-idealities, and device physics.

> **Note:** These are simplified models for simulation and visualization. Numeric values are model defaults or reported ranges, not validated hardware specs.
> **Literature values are unverified in this project; add peer-reviewed citations before external use ([CITATION NEEDED - placeholder value]).**

**Table of Contents**
- [Conductance Models](#conductance-models)
- [Matrix-Vector Multiplication](#matrix-vector-multiplication)
- [IR Drop Model](#ir-drop-model)
- [Sneak Path Model](#sneak-path-model)
- [Drift Model](#drift-model)
- [Temperature Effects](#temperature-effects)
- [Process Variation](#process-variation)
- [Verification & Validation](#verification--validation)

---

## Conductance Models

The crossbar array represents analog states as conductance values between minimum and maximum bounds.

### Constants

| Parameter | Value | Unit | Source |
|-----------|-------|------|--------|
| **Gmin** | 10 | µS | Design choice (OFF state) |
| **Gmax** | 100 | µS | Design choice (ON state) |
| **Ratio** | 10:1 | dimensionless | Gmax/Gmin |
| **Quantization Levels** | 30 | discrete states | Demo baseline (simulation baseline) |

### Linear Model

$$G = G_{\min} + (G_{\max} - G_{\min}) \times \frac{\text{level}}{29}$$

Where:
- `level` ∈ [0, 29] discrete analog state
- Level 0 → Gmin = 10 µS (OFF)
- Level 15 → ~55 µS (midpoint)
- Level 29 → Gmax = 100 µS (ON)

**Characteristics:**
- Simple linear interpolation
- Good for most applications
- Less accurate at extremes (low and high conductance regions)

### Exponential Model

$$G = G_{\min} \times \exp\left(\ln\left(\frac{G_{\max}}{G_{\min}}\right) \times \frac{\text{level}}{29}\right)$$

Equivalently:

$$G = G_{\min} \times \left(\frac{G_{\max}}{G_{\min}}\right)^{\text{level}/29}$$

Where at level = 14.5 (geometric mean):

$$G_{\text{mid}} = \sqrt{G_{\min} \times G_{\max}} = \sqrt{10 \times 100} \approx 31.6 \text{ µS}$$

**Characteristics:**
- Exponential scaling more realistic for FeFET behavior
- Better distribution across conductance range
- Can be a closer approximation for some ferroelectric devices (model assumption)
- Use as a comparative option, not a validated device model

### Lookup Table Model

$$G = \text{ConductanceTable}[\text{level}]$$

Where `ConductanceTable` is a pre-calibrated 30-entry table from measured device data.

**Characteristics:**
- Uses empirical calibration data
- Most accurate if calibration data available
- Falls back to linear if table unavailable
- Allows material-specific conductance curves
- Calibration data is not bundled in this repo; provide your own table

---

## Matrix-Vector Multiplication

The core operation: **y = W × x** where W is the weight matrix stored as conductances.

### Ohm's Law (Single Cell)

Current through a single cell:

$$I_{ij} = G_{ij} \times V_j$$

Where:
- $G_{ij}$ = normalized conductance [0, 1] (baseline MVM uses normalized; physical conversion is used in IR drop/drift analysis)
- $V_j$ = input voltage (DAC quantized)

### Column Summation (Single Output)

Current summed on word line i:

$$I_i = \sum_{j=0}^{N_{\text{cols}}-1} G_{ij} \times V_j$$

### Full Matrix-Vector Multiplication

```
For each row i:
  sum = 0
  For each column j:
    V_quantized = DAC_quantize(input[j])
    G_effective = conductance[i][j] × variationFactor  // normalized
    sum += G_effective × V_quantized
  output[i] = ADC_quantize(sum / N_cols)
```

**Note:** The physical mapping (linear/exponential/lookup via `GetPhysicalConductance`) is used in IR drop and device-physics analysis paths. The baseline MVM path remains normalized for speed and stable stacking across layers.

### Quantization Effects

**DAC Quantization (Input)**
- Input voltages quantized to DAC levels
- Module 2 GUI/CLI defaults: 8-bit DAC (configurable); tests often use 16-bit for near-ideal behavior
- Formula: $V_{\text{dac}} = \text{round}(V \times (2^{n}-1)) / (2^{n}-1)$

**ADC Quantization (Output)**
- Output currents quantized to ADC levels
- Module 2 GUI/CLI defaults: 6-bit ADC (configurable); tests often use 16-bit for near-ideal behavior
- Reduces output precision

**Impact on Accuracy:**
- ~1% accuracy loss per 3% RMSE for neural networks (illustrative; [CITATION NEEDED - placeholder value])
- 8-bit ADC/DAC: acceptable for inference (assumption; depends on model)
- 12+ bits: suitable for training (assumption; depends on model)

---

## IR Drop Model

Voltage drop due to wire resistance and cumulative current flow along metal lines.

### Wire Parameters

| Parameter | Default | Unit | Range | Notes |
|-----------|---------|------|-------|-------|
| $R_{\text{wl}}$ | 2.5 | Ω/pitch | 1-10 | Word line resistance per cell pitch |
| $R_{\text{bl}}$ | 2.5 | Ω/pitch | 1-10 | Bit line resistance per cell pitch |
| $R_{\text{contact}}$ | 50 | Ω | 10-100 | Contact resistance |

Based on a 45nm-era assumption (illustrative; [CITATION NEEDED - placeholder value]). Scales with feature size.

### Voltage Drop Formula

For cell (i, j) with row driver on left and sense amp/ground at top:

**Word Line Voltage (cumulative drop):**
$$V_{\text{wl}}(i,j) = V_{\text{input}} - I_i \times R_{\text{wl}} \times j$$

Where $I_i$ is current on word line i at column j.

**Bit Line Voltage (rise from ground):**
$$V_{\text{bl}}(i,j) = I_j \times R_{\text{bl}} \times i$$

Where $I_j$ is cumulative current in column j.

**Effective Cell Voltage:**
$$V_{\text{eff}}(i,j) = V_{\text{wl}}(i,j) - V_{\text{bl}}(i,j)$$

### Worst-Case Cell Location

IR drop is **maximum at far corner** (high row, high column):
- Both row and column experience maximum cumulative current
- Cell (rows-1, cols-1) typically worst case
- Can cause voltage reduction; 10-15% is an illustrative model range ([CITATION NEEDED - placeholder value])

### Iterative Relaxation Solver

IR drop is computed iteratively:

1. Initialize: $V_{\text{wl}}(i,j) = V_{\text{input}}$, $V_{\text{bl}}(i,j) = 0$
2. For each iteration (typically 100):
   - Calculate cell currents: $I_{ij} = V_{\text{eff}} \times G_{ij}$
   - Update word line voltages from cumulative current
   - Update bit line voltages from cumulative current
3. Converges when voltage changes < tolerance

### Architecture Effects

**0T1R (Passive Crossbar):**
- Sneak currents add to main current
- Effective resistance higher (~1.5× nominal)
- 10-15% IR drop is an illustrative range ([CITATION NEEDED - placeholder value])

**1T1R (Transistor-Isolated):**
- Transistor isolation reduces parasitic currents
- Lower effective current on lines
- 5-10% IR drop is an illustrative range ([CITATION NEEDED - placeholder value])

---

## Sneak Path Model

Unintended current paths through unselected cells that add noise to the selected cell signal.

### Physical Model

When reading/writing cell (selected_row, selected_col):

**Same Row Paths:** Current leaks from selected word line through unselected cells to unselected bit lines.

**Same Column Paths:** Current from unselected word lines leaks through unselected cells to the selected bit line.

**Off-Diagonal Paths:** Three-cell series paths form complete loops:
- Cell(selected_row, j) → Cell(i, j) → Cell(i, selected_col)

### Three-Cell Series Conductance

For off-diagonal sneak path with three cells:

$$G_{\text{path}} = \frac{1}{\frac{1}{G_1} + \frac{1}{G_2} + \frac{1}{G_3}}$$

Sneak current through path:

$$I_{\text{sneak}} = V \times G_{\text{path}}$$

### Sneak-to-Signal Ratio

For a selected cell with conductance $G_s$:

$$\text{SNR (dB)} = 20 \log_{10}\left(\frac{I_{\text{signal}}}{I_{\text{sneak}}}\right)$$

Where:
- Signal current: $I_s = V \times G_s$
- Sneak current: sum of all parasitic paths

### Architecture Comparison

**0T1R (Passive):**
- No transistor isolation
- All sneak paths contribute equally
- Sneak-to-signal ratio: 1-2 (0-6 dB SNR) is illustrative ([CITATION NEEDED - placeholder value])
- 5-20% accuracy loss is illustrative ([CITATION NEEDED - placeholder value])

**1T1R (Transistor Isolation):**
- Transistor OFF-state provides ~10^6 ON/OFF ratio
- Conservative estimate: 0.001 isolation factor
- Sneak-to-signal ratio: 10^-3 (60 dB SNR) is illustrative ([CITATION NEEDED - placeholder value])
- Sneak path effect is reduced but not eliminated (model assumption)

### Computation Modes

**Simplified Mode** (fast, less accurate):
- Fixed sneak factor (0.01 for 0T1R, 0.00001 for 1T1R)
- Used for arrays > 32×32
- Suitable for large-scale simulations

**Full Mode** (slower, accurate):
- Enumerate all three-cell paths
- Accurate sneak path magnitudes
- Used for arrays ≤ 32×32
- O(n^4) complexity

---

## Drift Model

Conductance drift over time, causing accuracy degradation.

### Physics Basis

Literature reports long retention for FeFETs (e.g., >10 years at 85°C), but this is unverified here ([CITATION NEEDED - placeholder value]). Drift is modeled logarithmically:

$$G(t) = G_0 \times \left(1 - \beta \times \ln\left(\frac{t}{t_0}\right)\right)$$

Or with power law:

$$G(t) = G_0 \times \left(\frac{t}{t_0}\right)^{\alpha}$$

Where:
- $G_0$ = initial conductance
- $\alpha$ = drift coefficient (very small for FeFET)
- $\beta$ = logarithmic drift coefficient
- $t_0$ = reference time (1 second)

### FeFET Drift Coefficient

Status labels below reflect literature reporting only; this project does not verify them.

| Source | Value | Status | Notes |
|--------|-------|--------|-------|
| **Assumed (default)** | 0.001 | ⚠️ Unverified | Estimated from >10 year retention requirement |
| **Literature-derived** | 0.0005 | ⚠️ Estimated | Implies <0.5 level drift per decade |
| **RRAM (comparison)** | 0.05 | ⚠️ Reported | Literature: 5-10% drift typical |
| **PCM (comparison)** | 0.1 | ⚠️ Reported | Highest among emerging memories |
| **Flash (comparison)** | 0.02 | ⚠️ Reported | Well-characterized in industry |

**Important:** FeFET drift coefficient 0.001 is **derived** from retention requirements, not directly measured. No reported in literature source provides explicit FeFET drift coefficients in the same format as RRAM/PCM papers.

### Thermal Activation (Arrhenius)

Drift processes are thermally activated:

$$\text{drift rate} \propto \exp\left(-\frac{E_a}{k_B T}\right)$$

Where:
- $E_a$ = activation energy ≈ 0.5 eV (assumed; [CITATION NEEDED - placeholder value])
- $k_B$ = Boltzmann constant = 1.38×10^-23 J/K
- $T$ = absolute temperature (K)

**Temperature Doubling Rule:** Drift accelerates ~2× per 20K temperature increase.

### 10-Year Retention Prediction

Estimated retention at room temperature (300K):

$$\text{Retention} = \left(1 - \text{drift coefficient} \times \ln(10 \text{ years})\right) \times 100\%$$

With 0.001 coefficient:
- ~99.9% retention at 10 years (excellent)

---

## Temperature Effects

Temperature affects multiple physics parameters.

**Implementation note:** `TemperatureEffects` provides these scalings. Current MVM-with-non-idealities applies temperature to wire resistance (IR drop); other temperature scalings (conductance window, noise, drift rate) are exposed via helpers and the drift simulator but are not auto-applied in the baseline MVM path.

**Note:** Numeric temperature effects below are illustrative or literature-reported and are not verified here ([CITATION NEEDED - placeholder value]).

### Wire Resistance

Copper temperature coefficient: α ≈ 0.004 /K

$$R(T) = R_0 \times (1 + \alpha(T - T_0))$$

Where $T_0 = 300K$ (room temperature).

**Effect:** Higher T → higher resistance → more IR drop.

### Conductance Range Scaling

**Cryogenic (T < 100K):**
- Enhanced ferroelectric polarization
- Pr at 4K can reach 75 µC/cm² vs 15-34 µC/cm² at room temperature
- Model: window expansion ~50% at 4K
- Gmin decreases (better OFF state)
- Gmax increases (better ON state)

**Elevated Temperature (T > 300K):**
- Thermal noise reduces effective polarization
- Window narrowing ~10% per 100K
- Both Gmin and Gmax shift toward center
- Degradation caps at 50% window loss

### Drift Rate Acceleration

Temperature-dependent drift scaling:

$$\text{drift rate factor} = \frac{\exp(-E_a / (k_B T))}{\exp(-E_a / (k_B T_0))}$$

Higher temperature significantly accelerates drift (exponential).

### Noise Scaling

Thermal noise voltage scales with $\sqrt{T}$:

$$V_{\text{noise}} \propto \sqrt{T}$$

Noise multiplier relative to 300K:

$$\text{noise factor} = \sqrt{\frac{T}{300K}}$$

### Temperature Presets

| Scenario | Temperature | Label |
|----------|-------------|-------|
| Deep Space | 4 K | Cryogenic |
| Liquid Nitrogen | 77 K | Cryogenic |
| Room Temperature | 300 K (27°C) | Standard |
| Industrial | 358 K (85°C) | Elevated |
| Automotive Grade 0 | 400 K (125°C) | High |

---

## Process Variation

Device-to-device variation and systematic spatial gradients.

### Device-to-Device Variation

Random variations in cell conductance due to fabrication randomness:

$$G_{\text{cell}} = G_{\text{nominal}} \times (1 + \sigma_{\text{device}} \times \mathcal{N}(0,1))$$

Default: $\sigma_{\text{device}} = 0.02$ (2% standard deviation)

### Spatial Gradients

Systematic spatial variation across the array (e.g., due to temperature gradient during processing):

**Horizontal Gradient:**
$$G(x) = G_0 \times \left(1 + \text{GradientX} \times (x - x_{\text{center}})\right)$$

**Vertical Gradient:**
$$G(y) = G_0 \times \left(1 + \text{GradientY} \times (y - y_{\text{center}})\right)$$

Default gradients: 0.001 per cell pitch (0.1% variation per cell from center).

### Edge Effects

Boundary cells often degrade due to:
- Proximity to die edge
- Thermal gradients
- Packaging stress

Edge degradation factor: 5% reduction (default)

```
Variation factor = Random × GradientX × GradientY × EdgeFactor
```

---

## Endurance & Fatigue

Cycle-dependent degradation of the conductance window.

### Degradation Model

After fatigue threshold, window narrows toward midpoint:

$$\text{degradation factor} = 1 - 0.5 \times (1 - e^{-3 \times \text{fatigue ratio}})$$

Where:
$$\text{fatigue ratio} = \frac{\text{cycles} - \text{fatigue threshold}}{\text{failure threshold} - \text{fatigue threshold}}$$

### Default Parameters

| Parameter | Value | Source |
|-----------|-------|--------|
| Fatigue Threshold | 10^8 cycles | IEEE IRPS ([CITATION NEEDED - placeholder value]) |
| Failure Threshold | 10^12 cycles | Claim requires peer-reviewed support ([CITATION NEEDED - placeholder value]) |
| Window Narrowing | 50% at failure | Assumed FeFET behavior ([CITATION NEEDED - placeholder value]) |

**Note:** Endurance modeling off by default (performance optimization). Enable for cycle-specific studies.

---

## Process Variation Configuration

### Device-to-Device Sigma

Random variation per cell:

```
G_measured = G_programmed × (1 + sigma × random_factor)
```

Default: 2% (1 level variation per ~14 levels on average)

### Systematic Gradients

Linear spatial variation across array:

- GradientX: 0.1% per cell horizontally
- GradientY: 0.1% per cell vertically
- Captures non-uniform processing, thermal effects

### Edge Effects

Boundary cells degrade by 5% (tunable).

Cells at array corners affected most due to both horizontal and vertical gradients plus edge effect.

---

## Verification & Validation

Physics models are checked against:

1. **Ohm's Law Tests** - IR drop follows V = I×R
2. **Conductance Scaling** - Sneak paths scale correctly with G
3. **Temperature Physics** - Arrhenius behavior modeled
4. **Architectural Comparison** - 1T1R vs 0T1R differences are qualitatively modeled (not validated)
5. **Technology Comparison** - Relative drift comparison (model-based, unverified)

### Test Coverage

Tests in `pkg/crossbar/physics_test.go`:

- IR drop increases toward array corner ✓
- IR drop scales with wire resistance ✓
- Sneak paths form three-cell series paths ✓
- Temperature effects on resistance/drift ✓
- Full MVM computation accuracy ✓

Run tests:
```bash
go test ./module2-crossbar/pkg/crossbar -v -run Physics
```

---

## Implementation Reference

### Key Files

| File | Purpose |
|------|---------|
| `array.go` | Array initialization, conductance models |
| `nonidealities.go` | IR drop, sneak path analysis |
| `irdrop.go` | IR drop simulator and analysis |
| `sneakpath.go` | Sneak path calculator |
| `drift.go` | Drift simulator and Arrhenius model |
| `temperature.go` | Temperature effect calculations |
| `enhanced.go` | MVMWithNonIdealities comprehensive analysis |

### Key Structures

```go
// Conductance model selection
type ConductanceModel int
const (
    ConductanceLinear      // Linear: G = Gmin + (Gmax-Gmin)*level/29
    ConductanceExponential // Exponential: G = Gmin*exp(ln(Gmax/Gmin)*level/29)
    ConductanceLookup      // Lookup table from calibration
)

// Wire parameters for IR drop
type WireParams struct {
    RwordLine float64 // Ω per cell pitch
    RbitLine  float64 // Ω per cell pitch
    Rcontact  float64 // Contact resistance (Ω)
}

// MVM operation options
type MVMOptions struct {
    EnableIRDrop     bool
    EnableSneakPaths bool
    EnableVariation  bool
    EnableDrift      bool
    Temperature      float64 // Kelvin
    Architecture     string  // "1T1R" or "0T1R"
}
```

---

## Physical Constants

| Constant | Value | Unit | Source |
|----------|-------|------|--------|
| Boltzmann constant | 1.38×10^-23 | J/K | Physics |
| Electron volt | 1.6×10^-19 | J | Physics |
| Copper TCR | 0.00393 | /K | Material data |
| FeFET activation energy | 0.5 | eV | Typical ferroelectric |

---

## References

**Peer-Reviewed Sources ([CITATION NEEDED - placeholder value]):**
- Add specific endurance/retention citations for FeFET/FE devices used to justify default thresholds.
- Avoid future-dated or placeholder citations; include full reference + DOI when verified.

**Implementation Notes:**
- Physics models are literature-informed but not fully validated
- FeFET drift coefficient estimated, not directly measured
- Temperature models use Arrhenius activation energy (assumed)
- Sneak path model follows a three-cell series formulation (literature-inspired)
