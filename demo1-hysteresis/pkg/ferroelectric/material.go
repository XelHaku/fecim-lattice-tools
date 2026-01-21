// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import "math"

// HZOMaterial contains material parameters for Hafnium-Zirconium Oxide.
// Values are based on literature measurements for HfO2-ZrO2 superlattices.
// References:
// - Park et al., Adv. Mater. 2015 (HZO ferroelectricity)
// - Müller et al., Nano Lett. 2012 (HfO2 ferroelectric properties)
// - Pesic et al., Adv. Funct. Mater. 2016 (wake-up, fatigue)
type HZOMaterial struct {
	Name string // Material name/identifier

	// Polarization parameters (C/m²)
	Pr float64 // Remanent polarization
	Ps float64 // Saturation polarization

	// Field parameters (V/m)
	Ec float64 // Coercive field

	// Dielectric properties
	Epsilon   float64 // Relative permittivity (high frequency)
	EpsilonLF float64 // Low frequency permittivity
	LossAngle float64 // Dielectric loss tangent (tan δ)

	// Film properties
	Thickness float64 // Film thickness (m)
	Area      float64 // Active area (m²)

	// Dynamics
	Tau   float64 // Characteristic switching time (s)
	Tau0  float64 // Activation attempt frequency inverse (s)
	Ea    float64 // Activation energy for switching (eV)
	Alpha float64 // Switching exponent (KAI model)

	// Temperature properties
	CurieTemp   float64 // Curie temperature (K)
	TempCoeffEc float64 // Temperature coefficient of Ec (V/m/K)
	TempCoeffPr float64 // Temperature coefficient of Pr (C/m²/K)

	// Reliability parameters
	EnduranceCycles float64 // Endurance limit (cycles)
	RetentionTime   float64 // Retention time at 85°C (s)
	ImrintField     float64 // Imprint field shift (V/m)
}

// DefaultHZO returns material parameters for typical Si-doped HfO2 (Hf0.5Zr0.5O2).
// Based on literature values for 10nm ALD-deposited films.
// Ref: Park et al., Adv. Mater. 27, 1811 (2015)
func DefaultHZO() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HZO (Si-doped)",
		Pr:              25e-2,   // 25 μC/cm² = 0.25 C/m²
		Ps:              30e-2,   // 30 μC/cm² = 0.30 C/m²
		Ec:              1.2e8,   // 1.2 MV/cm = 1.2e8 V/m
		Epsilon:         30,      // High frequency permittivity
		EpsilonLF:       38,      // Low frequency (includes domain contribution)
		LossAngle:       0.02,    // tan δ ≈ 2%
		Thickness:       10e-9,   // 10 nm
		Area:            100e-12, // 100 nm² typical device
		Tau:             1e-9,    // ~1 ns intrinsic switching
		Tau0:            1e-13,   // Attempt frequency ~10 THz
		Ea:              0.7,     // Activation energy ~0.7 eV
		Alpha:           2.0,     // KAI exponent (2D growth)
		CurieTemp:       723,     // ~450°C
		TempCoeffEc:     -2e5,    // Ec decreases with T
		TempCoeffPr:     -5e-5,   // Pr decreases with T
		EnduranceCycles: 1e10,    // 10^10 cycles
		RetentionTime:   3.15e9,  // 100 years at 85°C
		ImrintField:     1e6,     // Small imprint
	}
}

// OptimizedHZO returns parameters for optimized superlattice HZO.
// HfO2/ZrO2 superlattice with enhanced properties.
// Ref: Cheema et al., Nature 580, 478 (2020)
func OptimizedHZO() *HZOMaterial {
	return &HZOMaterial{
		Name:            "HZO Superlattice",
		Pr:              45e-2, // 45 μC/cm² (superlattice enhanced)
		Ps:              50e-2, // 50 μC/cm²
		Ec:              0.8e8, // 0.8 MV/cm (reduced by negative capacitance)
		Epsilon:         35,
		EpsilonLF:       50,
		LossAngle:       0.015,
		Thickness:       10e-9,
		Area:            100e-12,
		Tau:             0.5e-9, // Faster switching
		Tau0:            1e-13,
		Ea:              0.5, // Lower activation energy
		Alpha:           2.5,
		CurieTemp:       773, // Higher Tc
		TempCoeffEc:     -1.5e5,
		TempCoeffPr:     -3e-5,
		EnduranceCycles: 1e12, // Better endurance
		RetentionTime:   1e10, // Excellent retention
		ImrintField:     0.5e6,
	}
}

// FeCIMMaterial returns parameters matching FeCIM specifications.
// Based on Dr. Tour's Nov 2024 presentation data.
//
// ⚠️ HONESTY NOTE: Some parameters are TARGETS not DEMONSTRATED values.
// Dr. Tour's presentation (Nov 2024) disclosed:
//
// VERIFIED/DEMONSTRATED:
//   - 30 discrete analog states (VERIFIED)
//   - 87% MNIST accuracy with 88% theoretical max (VERIFIED)
//   - 10^9 endurance cycles DEMONSTRATED
//   - 10^7 seconds retention DEMONSTRATED (~116 days)
//
// TARGET (NOT YET ACHIEVED):
//   - 10^12 endurance cycles (stated as target, not demonstrated)
//   - 100 year retention (not demonstrated)
//
// ESTIMATED (NOT DISCLOSED):
//   - Pr, Ps, Ec (using standard HZO literature values)
//   - Tau (slides showed 10ns, using 10ns not 1ns)
//   - All other parameters are literature-based estimates
//
// See: opensource/papers/08_Documentation/HONESTY_AUDIT.md
func FeCIMMaterial() *HZOMaterial {
	return &HZOMaterial{
		Name:            "FeCIM HZO",
		Pr:              30e-2,         // 30 μC/cm² - ESTIMATED (not disclosed)
		Ps:              35e-2,         // 35 μC/cm² - ESTIMATED (not disclosed)
		Ec:              1.0e8,         // 1.0 MV/cm - ESTIMATED (not disclosed)
		Epsilon:         32,            // ESTIMATED
		EpsilonLF:       40,            // ESTIMATED
		LossAngle:       0.01,          // Low loss - ESTIMATED
		Thickness:       10e-9,         // ESTIMATED
		Area:            45e-9 * 45e-9, // 45nm pitch cell - ESTIMATED
		Tau:             10e-9,         // 10 ns (from Tour's slides, not 1ns)
		Tau0:            1e-13,
		Ea:              0.6,
		Alpha:           2.0,
		CurieTemp:       723,
		TempCoeffEc:     -1.8e5,
		TempCoeffPr:     -4e-5,
		EnduranceCycles: 1e9, // 10^9 DEMONSTRATED (10^12 is TARGET)
		RetentionTime:   1e7, // 10^7 sec (~116 days) DEMONSTRATED
		ImrintField:     1e6,
	}
}

// FeCIMMaterialTarget returns FeCIM TARGET specifications.
// ⚠️ WARNING: These are GOALS not demonstrated performance.
// Dr. Tour explicitly stated: "We still have to get this up to the
// required 10^12 cycles."
func FeCIMMaterialTarget() *HZOMaterial {
	mat := FeCIMMaterial()
	mat.Name = "FeCIM HZO (TARGET - NOT DEMONSTRATED)"
	mat.EnduranceCycles = 1e12 // TARGET - not yet achieved
	mat.RetentionTime = 3.15e9 // TARGET - 100 years (not demonstrated)
	mat.Tau = 1e-9             // TARGET - 1ns (demonstrated ~10ns)
	return mat
}

// CoerciveVoltage returns the coercive voltage for a given thickness.
func (m *HZOMaterial) CoerciveVoltage() float64 {
	return m.Ec * m.Thickness
}

// RemanentCharge returns the remanent charge per unit area.
func (m *HZOMaterial) RemanentCharge() float64 {
	return m.Pr
}

// Capacitance returns the capacitance of the ferroelectric layer.
func (m *HZOMaterial) Capacitance() float64 {
	epsilon0 := 8.854e-12 // F/m
	return epsilon0 * m.Epsilon * m.Area / m.Thickness
}

// SwitchingEnergy returns the energy required for complete switching.
// E = 2 * Pr * Ec * Volume
func (m *HZOMaterial) SwitchingEnergy() float64 {
	volume := m.Area * m.Thickness
	return 2 * m.Pr * m.Ec * volume
}

// SwitchingTime returns temperature-dependent switching time.
// τ(T) = τ0 * exp(Ea / kT)
func (m *HZOMaterial) SwitchingTime(T float64) float64 {
	kB := 8.617e-5 // eV/K
	return m.Tau0 * math.Exp(m.Ea/(kB*T))
}

// CoerciveFieldAtTemp returns temperature-dependent coercive field.
// Ec(T) = Ec0 * (1 - T/Tc)^0.5
func (m *HZOMaterial) CoerciveFieldAtTemp(T float64) float64 {
	if T >= m.CurieTemp {
		return 0
	}
	return m.Ec * math.Pow(1-T/m.CurieTemp, 0.5)
}

// PolarizationAtTemp returns temperature-dependent remanent polarization.
// Pr(T) = Pr0 * (1 - T/Tc)^0.5
func (m *HZOMaterial) PolarizationAtTemp(T float64) float64 {
	if T >= m.CurieTemp {
		return 0
	}
	return m.Pr * math.Pow(1-T/m.CurieTemp, 0.5)
}

// EnduranceAtCycles returns estimated Pr degradation after N cycles.
// Using stretched exponential model: Pr(N) = Pr0 * exp(-(N/N0)^β)
func (m *HZOMaterial) EnduranceAtCycles(N float64) float64 {
	beta := 0.3 // Stretched exponential factor
	return m.Pr * math.Exp(-math.Pow(N/m.EnduranceCycles, beta))
}

// RetentionAtTime returns estimated Pr retention after time t at temperature T.
// Using Arrhenius model for retention loss.
func (m *HZOMaterial) RetentionAtTime(t, T float64) float64 {
	// Acceleration factor for temperature
	Tref := 358.0 // 85°C reference
	kB := 8.617e-5
	Ea := 1.0 // Retention activation energy (eV)

	accel := math.Exp(Ea / kB * (1/Tref - 1/T))
	effectiveTime := t * accel

	if effectiveTime >= m.RetentionTime {
		return 0.9 * m.Pr // 10% loss at end of life
	}
	return m.Pr * (1 - 0.1*effectiveTime/m.RetentionTime)
}

// DiscreteLevel returns the conductance for a given state level (0-29).
func (m *HZOMaterial) DiscreteLevel(level int, totalLevels int) float64 {
	// Linear mapping from level to normalized polarization
	normalizedP := -1.0 + 2.0*float64(level)/float64(totalLevels-1)

	// Map to conductance (1-100 µS range)
	Gmin := 1e-6   // 1 µS
	Gmax := 100e-6 // 100 µS
	return Gmin + (Gmax-Gmin)*(normalizedP+1)/2
}
