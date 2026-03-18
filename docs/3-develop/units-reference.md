# Units Reference

All internal computations use SI units. Display/literature units are converted at the presentation layer via `shared/physics/units.go`.

## Master Quantity Table

| Quantity | Internal (SI) | Display | Conversion | `units.go` Function |
|----------|--------------|---------|------------|---------------------|
| Polarization | C/m² | µC/cm² | × 100 | `FormatPolarization()`, `Cm2ToMuCcm2()`, `MuCcm2ToCm2()` |
| Electric field | V/m | MV/cm, kV/cm | ÷ 10⁸ (MV/cm) | `FormatElectricField()`, `VPerMToMVPerCm()`, `MVPerCmToVPerM()` |
| Stress | Pa | GPa | ÷ 10⁹ | Config uses GPa; solver uses Pa |
| Conductance | S | nS–S | auto-scaled | `FormatConductance()` |
| Energy | J | fJ–J | auto-scaled | `FormatEnergy()`, `FormatEnergyMJ()`, `FormatEnergyUJ()` |
| Current | A | pA–A | auto-scaled | `FormatCurrent()` |
| Voltage | V | µV–V | auto-scaled | `FormatVoltage()` |
| Time | s | ps–s | auto-scaled | `FormatTime()` |
| Frequency | Hz | Hz–GHz | auto-scaled | `FormatFrequency()` |
| Resistance | Ω | mΩ–GΩ | auto-scaled | `FormatResistance()` |
| Capacitance | F | aF–F | auto-scaled | `FormatCapacitance()` |
| Power | W | fW–kW | auto-scaled | `FormatPower()` |
| Charge | C | fC–C | auto-scaled | `FormatCharge()` |
| Temperature | K | K (or °C for display) | K − 273.15 = °C | — |
| Thickness | m | nm | × 10⁹ | `Material.ThicknessNm()` |
| Area | m² | — | — | — |
| Viscosity (ρ) | Ω·m | — | — | — |
| Depolarization (K_dep) | V·m/C | — | — | — |
| Thermal resistance | K/W | — | — | `ThermalConfig.ThermalResistance` |
| Thermal capacitance | J/K | — | — | `ThermalConfig.ThermalCapacitance` |
| Activation energy | eV | eV | — | `Material.ActivationEnergyEV` |
| Electrostriction (Q12) | m⁴/C² | — | — | — |
| Cell pitch | nm (config) | nm | — | `Material.CellPitchNm` |
| Conductance range | S | µS | GMin = 10 µS, GMax = 100 µS | `physics.GMin`, `physics.GMax` |
| Power density | W/m² | — | — | `ThermalState.PowerDensityWm2` |

## Physical Constants

| Constant | Symbol | Value | Unit | Location |
|----------|--------|-------|------|----------|
| Vacuum permittivity | ε₀ | 8.854 × 10⁻¹² | F/m | `landau.go` (local const) |
| Boltzmann constant | k_B | 1.380649 × 10⁻²³ | J/K | `landau.go` (Langevin noise) |
| Boltzmann constant | k_B | 8.617 × 10⁻⁵ | eV/K | `material.go` (Arrhenius) |
| E-field conversion | — | 10⁸ | V/m per MV/cm | `units.go` (`VPerMPerMVPerCm`) |

## Known Mixed-Unit Conventions

### Stress: GPa (config) vs Pa (solver)

- `config/physics/physics.go`: `MaterialCoupling.StressGPa` stores in **GPa**
- `shared/physics/landau.go`: `LKSolver.Stress` uses **Pa** (SI)
- Conversion: `LKSolver.Stress = StressGPa × 10⁹`
- Sign convention: positive = tensile, negative = compressive

### Polarization: µC/cm² (literature) vs C/m² (internal)

- Literature commonly reports Pr in µC/cm²
- Internal storage in C/m² (SI)
- Conversion: `1 µC/cm² = 0.01 C/m²` (divide by 100)
- Use `MuCcm2ToCm2()` for input, `Cm2ToMuCcm2()` for output

### Electric Field: MV/cm (literature) vs V/m (internal)

- Literature reports Ec in MV/cm
- Internal storage in V/m (SI)
- Conversion: `1 MV/cm = 10⁸ V/m`
- `FormatElectricField()` auto-selects MV/cm or kV/cm based on magnitude
- Use `MVPerCmToVPerM()` for input, `VPerMToMVPerCm()` for output

### Thickness: nm (config/display) vs m (internal)

- Config YAML and Material struct store thickness in **m** (`thickness_m`)
- Display via `Material.ThicknessNm()` converts to **nm** (× 10⁹)
- Cell pitch stored directly in nm (`cell_pitch_nm`)

### Conductance: normalized [0,1] vs physical (S)

- Crossbar cells store normalized conductance in [0, 1]
- Physical mapping: `GMin = 10 µS`, `GMax = 100 µS`, ratio = 10×
- Three models: linear, exponential (realistic FeFET), lookup (calibrated)
- `GetPhysicalConductance()` maps normalized to physical Siemens

### Activation Energy: eV (config) vs J (Boltzmann)

- Material config stores activation energy in **eV** (`activation_energy_ev`)
- Arrhenius calculations use `k_B = 8.617 × 10⁻⁵ eV/K` to stay in eV throughout
- Langevin noise uses `k_B = 1.380649 × 10⁻²³ J/K` (SI)

## Auto-Scaling Ranges

Each `FormatX()` function selects the most readable SI prefix:

| Function | Prefix range | Thresholds |
|----------|-------------|------------|
| `FormatEnergy` | fJ, pJ, nJ, µJ, mJ, J | 10⁻¹², 10⁻⁹, 10⁻⁶, 10⁻³, 1 |
| `FormatConductance` | nS, µS, mS, S | 10⁻⁶, 10⁻³, 1 |
| `FormatCurrent` | pA, nA, µA, mA, A | 10⁻⁹, 10⁻⁶, 10⁻³, 1 |
| `FormatVoltage` | µV, mV, V | 10⁻³, 1 |
| `FormatTime` | ps, ns, µs, ms, s | 10⁻⁹, 10⁻⁶, 10⁻³, 1 |
| `FormatFrequency` | Hz, kHz, MHz, GHz | 10³, 10⁶, 10⁹ |
| `FormatResistance` | mΩ, Ω, kΩ, MΩ, GΩ | 1, 10³, 10⁶, 10⁹ |
| `FormatCapacitance` | aF, fF, pF, nF, µF, mF, F | 10⁻¹⁵, 10⁻¹², 10⁻⁹, 10⁻⁶, 10⁻³, 1 |
| `FormatPower` | fW, pW, nW, µW, mW, W, kW | 10⁻¹², 10⁻⁹, 10⁻⁶, 10⁻³, 1, 10³ |
| `FormatCharge` | fC, pC, nC, µC, mC, C | 10⁻¹², 10⁻⁹, 10⁻⁶, 10⁻³, 1 |

## Guidelines for Contributors

1. All new physics quantities **MUST** use SI units internally
2. Display conversions go through `shared/physics/units.go`
3. Config YAML may use convenient units (GPa, nm, eV) -- document the conversion in the struct tag or doc comment
4. When in doubt, check `FormatX()` functions in `units.go`
5. Conductance constants live in `shared/physics/conductance.go` (GMin, GMax)
6. Thermal parameters live in `shared/crossbar/thermal.go` (K/W, J/K)
