<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# shared/physics

## Purpose

Core physics package (80+ files, 90KB+) providing ferroelectric material models, Landau-Khalatnikov solver, ISPP WriteController, Preisach hysteresis model, quantization utilities, conductance models, technology comparison frameworks, and comprehensive physics validation. Single source of truth for all physical simulations across modules.

## Key Files

| File | Description |
|------|-------------|
| `material.go` | HZOMaterial type and factory functions (34KB). Pr, Ps, Ec, delta, alpha/beta/gamma coefficients. THREE-TIER SYSTEM: DefaultHZO, FeCIMMaterial (conservative), FeCIMMaterialTarget (aspirational), LiteratureSuperlattice (best-case). |
| `landau.go` | Landau-Khalatnikov solver (19KB). Runge-Kutta-4 integration of polarization dynamics under applied field. Nucleation-limited switching model. |
| `ispp_write.go` | WriteController for ISPP (14KB). Binary search convergence, guard-band correction, overshoot recovery. Bridges module1 and module4. |
| `preisach.go` | Preisach hysteresis adapter (6.3KB). Product-form Everett function (fixed in Feb 2026). Provides minor loop calculations. |
| `tanh_everett.go` | Everett function adapter (611 bytes). Wraps shared Preisach Everett for compatibility. |
| `calibration.go` | Material parameter calibration (9.3KB). Tunes material coefficients to match literature targets. |
| `conductance.go` | Conductance model: Linear, Exponential, Lookup (4.9KB). Maps polarization to device conductance with quantization. |
| `quantization.go` | Discrete level quantization (3KB). 30-state discretization and level-to-conductance mapping. |
| `technology.go` | Technology node comparison (3.6KB). Scales device parameters to different nodes. |
| `power.go` | Power dissipation calculation (3.6KB). From field, current, switching energy. |
| `transfer.go` | Transfer function models for nonlinear device behavior (2.7KB). |
| `units.go` | Unit conversion and scaling utilities (9.9KB). V↔V/m, SI unit conversions. |
| `worldclass_*.go` | "World-class" performance benchmarks (8 files, ~9KB). CV, FORC, PUND, retention, frequency dispersion, wakeup validation. |
| Extensive test suite | 80+ test files. Regression, golden data, property-based, stress, e2e pipeline tests. |

## For AI Agents

### Working In This Directory

**THREE-TIER MATERIAL SYSTEM (CRITICAL):**

For honest scientific communication:

1. **DefaultHZO()** - Baseline Si-doped HfO2/ZrO2 from literature (Park et al., Müller et al.)
2. **FeCIMMaterial()** - Conservative: Only values demonstrated in Dr. Tour's published work
3. **FeCIMMaterialTarget()** - Aspirational: Conference-claimed targets (not yet peer-reviewed)
4. **LiteratureSuperlattice()** - Best-case: HfO2/ZrO2 superlattices (Cheema et al. Nature 2020)

Use FeCIMMaterial() for educational visualization. Use LiteratureSuperlattice() to show theoretical potential.

**Everett Function (FIXED Feb 2026):**

Product form is correct:
```
Everett = [1+tanh((α-Ec)/Δ)] * [1-tanh((β+Ec)/Δ)] * Ps/4
```

Never revert to factorized-difference form. Golden regression data regenerated.

**Landau-Khalatnikov Solver:**

Integrates: `rho_eff * dP/dt = E - K_dep*P - dG/dP + noise`

- State: polarization P (C/m²), cumulative NLS progress
- RK-4 integration with adaptive stepping
- Handles nucleation-limited switching under E < Ec
- Thread-safe for concurrent cell simulation

**WriteController (ISPP):**

Shared across Module1 and Module4:

- Binary search on voltage: VMin (not enough), VMax (too much)
- Guard pulses (max 2) during VERIFY to nudge stuck states
- Accept ±1 logic: Accept when within 1 level of target AFTER > 8 overshoots
- Overshoot limit (30): Triggers SUCCESS for physics-limited materials
- Reset shortcut: Rapid reset when > 3 overshoots (recovers bounds)

**Conductance Models:**

Three choices: Linear, Exponential, or LookUp table

- Linear: `G = GMin + (GMax - GMin) * (P - Pmin) / (Pmax - Pmin)`
- Exponential: `G = GMin * exp(k * P)` for nonlinear switching
- Lookup: Interpolate from pre-computed table (fastest for large arrays)

**Quantization:**

Default 30 discrete levels. Maps continuous conductance to level:

```
level = round((G - GMin) / (GMax - GMin) * 29)
```

Inverse: `G = GMin + (GMax - GMin) * level / 29`

**Worldclass Validation:**

Eight worldclass benchmarks validate against literature claims:

- CV (Capacitance-Voltage): Hysteresis shape, coercivity
- FORC (First-Order Reversal Curves): Switching distribution
- PUND (Positive-Up Negative-Down): Polarization switching
- Retention: P decay over time at elevated temperature
- Frequency dispersion: Frequency-dependent permittivity
- Wakeup: Wake-up effect recovery over time

Use for claims validation.

### Testing Requirements

```bash
# Run all physics tests
go test ./shared/physics -v

# Run material tests
go test ./shared/physics -run TestMaterial -v

# Run Landau-Khalatnikov tests
go test ./shared/physics -run TestLandau -v

# Run ISPP WriteController tests (critical)
go test ./shared/physics -run TestISPPWrite -v

# Run Preisach tests
go test ./shared/physics -run TestPreisach -v

# Run golden regression tests
go test ./shared/physics -run TestGoldenRegression -v

# Run conductance model tests
go test ./shared/physics -run TestConductance -v

# Run worldclass validation tests
go test ./shared/physics -run TestWorldclass -v

# Run e2e pipeline test
go test ./shared/physics -run TestE2EPipeline -v

# Run stress tests
go test ./shared/physics -run TestStress -v

# Run property-based tests
go test ./shared/physics -run TestProperty -v

# Run benchmarks
go test ./shared/physics -bench Bench -benchmem
```

### Common Patterns

- **Material creation**: Pick factory function (DefaultHZO, FeCIMMaterial, etc.)
- **Field conversion**: `field = voltage / thickness` where thickness from material
- **LK integration**: Instantiate solver with material, step via RK-4
- **ISPP convergence**: Use WriteController; call WriteTarget(level) to program
- **Preisach loops**: Use product-form Everett; calibrate Delta to match Pr
- **Quantization**: QuantizeTo30Levels(conductance) for discrete levels
- **Conductance mapping**: ConductanceModel determines G from P
- **Worldclass validation**: Run before publishing device claims

## Dependencies

### Internal

- `shared/logging` - Physics operation logging
- `config/physics` - External configuration overrides (optional)

### External

- `math` (Go stdlib) - RK-4 integration, transcendental functions
- `math/rand` (Go stdlib) - Stochastic switching, device variation
- Standard library: `fmt`, `sync`, `time`, `os`

### Known Issues

- None reported; physics is stable and well-tested
- Everett function fix (Feb 2026) is complete; golden data regenerated
- 80+ tests ensure robustness across edge cases

<!-- MANUAL: Last edited 2026-02-13. Core physics package; single source of truth. THREE-TIER MATERIAL SYSTEM is critical for honesty. -->
