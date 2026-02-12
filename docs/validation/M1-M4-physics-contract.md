# M1–M4 Physics Contract: Preisach Polarization ↔ Conductance Mapping

## Scope
This document validates the shared physics contract between:
- **Module 1 (hysteresis)**: Preisach polarization state generation
- **Module 4 (circuits)**: conductance-level mapping used for read/compute paths

## 1) Shared types and locations

### Module 1 polarization path
- `module1-hysteresis/pkg/ferroelectric/preisach.go`
  - `type PreisachModel`
  - `func (p *PreisachModel) Update(E float64) float64` → returns **P in C/m²**
  - `func (p *PreisachModel) DiscreteStates(n int) []DiscreteState`
- `shared/physics/preisach.go`
  - `type PreisachStack`
  - `func (ps *PreisachStack) Update(E float64) float64` → **P in C/m²**

### Shared material + transfer definitions (single source of truth)
- `shared/physics/material.go`
  - `type HZOMaterial` (contains `Pr`, `Ps`, `Gmin`, `Gmax`, `NumLevels`, etc.)
  - `func DefaultHZO() *HZOMaterial`
  - `func (m *HZOMaterial) DiscreteLevel(level, totalLevels int) float64`
- `shared/physics/transfer.go`
  - `func PolarizationToConductance(P, Ps, Gmin, Gmax float64) float64`
  - `func ConductanceToPolarization(G, Gmin, Gmax, Ps float64) float64`

### Module 4 conductance path
- `module4-circuits/pkg/gui/device_state.go`
  - imports `sharedphysics "fecim-lattice-tools/shared/physics"`
  - `func (ds *DeviceState) levelToConductance(level, levels int) float64`
    - uses `mat.DiscreteLevel(level, levels)` from **shared material**

## 2) Unit conventions
- Polarization `P`, `Pr`, `Ps`: **C/m²** internally
  - conversion: `1 C/m² = 100 µC/cm²`
- Conductance `G`, `Gmin`, `Gmax`: **S** (Siemens)
  - display often in µS (`S × 1e6`)
- Electric field `E`, `Ec`: **V/m**

## 3) Mapping equations

### Polarization → Conductance (shared/physics/transfer.go)
\[
G = G_{min} + (G_{max}-G_{min})\frac{\left(\frac{P}{P_s}+1\right)}{2}
\]
with `P/Ps` clamped to `[-1, +1]`.

### Discrete level → Conductance (shared/physics/material.go)
- normalized polarization:
\[
P_n = -1 + 2\frac{\text{level}}{N-1}
\]
- then:
\[
G = G_{min} + (G_{max}-G_{min})\frac{(P_n+1)}{2}
\]

These are algebraically consistent with the P→G transfer above.

## 4) Tolerances used in validation
- Equality checks between transfer-function result and Module 4 level mapping:
  - absolute tolerance = `max(|G|·1e-12, 1e-18)` S
- Monotonicity check:
  - enforce non-decreasing conductance with level index (within `1e-18` S slack)
- Range check:
  - each level must stay within `[Gmin, Gmax]` (with `1e-18` S slack)

## 5) Copy-paste drift check (import/source consistency)
- `module1-hysteresis/pkg/ferroelectric/material.go` is a re-export layer of `shared/physics` (`type HZOMaterial = sharedphysics.HZOMaterial`, `DefaultHZO() -> sharedphysics.DefaultHZO()` etc.)
- `module4-circuits/pkg/gui/device_state.go` imports and uses `shared/physics` material/mapping directly.

**Conclusion:** both modules rely on the same shared physics package and mapping definitions; no independent duplicate mapping implementation is required for the validated path.

## 6) Material parameter consistency (Module 1 vs Module 4)
For default HZO (`DefaultHZO()`), Module 4 receives the exact same material object values used by Module 1:
- `Ps = 0.30 C/m²` (30 µC/cm²)
- `Pr = 0.25 C/m²` (25 µC/cm²)
- `Gmin = 1e-6 S`
- `Gmax = 100e-6 S`
- `NumLevels = 30`

Validated by test: `validation/m1_m4_physics_consistency_test.go`.
