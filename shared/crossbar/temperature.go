// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import "math"

// TemperatureEffects models temperature-dependent physics effects.
// FeFET devices show enhanced properties at cryogenic temperatures
// and degraded performance at elevated temperatures.
type TemperatureEffects struct {
	AmbientK float64 // Operating temperature in Kelvin
}

// NewTemperatureEffects creates a temperature effects model.
func NewTemperatureEffects(tempK float64) *TemperatureEffects {
	if tempK < 0 {
		tempK = 300 // Default to room temperature
	}
	return &TemperatureEffects{AmbientK: tempK}
}

// TemperaturePresets provides common operating temperatures.
const (
	TempCryogenic  = 77.0  // Liquid nitrogen temperature
	TempColdSpace  = 4.0   // Deep space / liquid helium
	TempRoom       = 300.0 // Room temperature (27°C)
	TempAutomotive = 400.0 // Automotive Grade 0 (125°C)
	TempIndustrial = 358.0 // Industrial grade (85°C)
)

// AdjustedWireResistance applies temperature coefficient of resistance (TCR) to wire resistance.
// Uses copper TCR = 0.00393 /K (3.93% per Kelvin change from 300K reference).
// Higher temperature → higher resistance → more IR drop.
func (t *TemperatureEffects) AdjustedWireResistance(R0 float64) float64 {
	const copperTCR = 0.00393 // Copper temperature coefficient
	return R0 * (1.0 + copperTCR*(t.AmbientK-300.0))
}

// AdjustedConductanceRange scales Gmin/Gmax conductance window with temperature.
// Physics basis:
//   - Cryogenic (<100K): Enhanced ferroelectric polarization → wider window
//   - High temp (>300K): Thermal noise reduces effective window
//
// Returns (adjustedGmin, adjustedGmax).
func (t *TemperatureEffects) AdjustedConductanceRange(gMin, gMax float64) (float64, float64) {
	if t.AmbientK < 100 {
		// Cryogenic enhancement
		// At 4K, Pr can reach 75 µC/cm² vs 15-34 µC/cm² at RT (Adv. Elec. Mat. 2024)
		// Model as window expansion factor
		enhancementFactor := 1.0 + 0.5*(100-t.AmbientK)/100
		// Gmin decreases (better OFF state), Gmax increases (better ON state)
		return gMin / enhancementFactor, gMax * enhancementFactor
	}

	if t.AmbientK > 300 {
		// High temperature degradation
		// Thermal fluctuations reduce effective polarization
		// Model as window narrowing (conservative estimate)
		degradationFactor := 1.0 - 0.1*(t.AmbientK-300)/100
		if degradationFactor < 0.5 {
			degradationFactor = 0.5 // Cap at 50% degradation
		}
		// Window narrows symmetrically toward center
		gMid := (gMin + gMax) / 2
		gRange := (gMax - gMin) * degradationFactor
		return gMid - gRange/2, gMid + gRange/2
	}

	// Room temperature - no adjustment
	return gMin, gMax
}

// AdjustedDriftRate scales drift rate with temperature using Arrhenius model.
// Drift processes are thermally activated: rate ∝ exp(-Ea/kT)
// Higher temperature → faster drift.
//
// Reference: Thermal activation energy Ea ≈ 0.5 eV for typical ferroelectric switching.
func (t *TemperatureEffects) AdjustedDriftRate(driftCoeff float64) float64 {
	const kB = 1.38e-23 // Boltzmann constant (J/K)
	const Ea = 0.5      // Activation energy (eV)
	const eV = 1.6e-19  // Electron-volt to Joules

	// Reference rate at 300K
	refRate := math.Exp(-Ea * eV / (kB * 300))
	// Rate at operating temperature
	newRate := math.Exp(-Ea * eV / (kB * t.AmbientK))

	// Scale drift coefficient by rate ratio
	return driftCoeff * (newRate / refRate)
}

// AdjustedRetention estimates retention time scaling with temperature.
// Returns a factor to multiply the nominal retention time.
// Lower temperature → exponentially better retention.
func (t *TemperatureEffects) AdjustedRetention() float64 {
	const kB = 1.38e-23
	const Ea = 0.5
	const eV = 1.6e-19

	// Higher activation energy ratio → longer retention
	refRate := math.Exp(-Ea * eV / (kB * 300))
	newRate := math.Exp(-Ea * eV / (kB * t.AmbientK))

	// Retention scales inversely with rate
	return refRate / newRate
}

// AdjustedNoise scales thermal noise with temperature.
// Thermal noise voltage Vn ∝ sqrt(kT), so noise power ∝ T.
// Returns a noise multiplier relative to 300K.
func (t *TemperatureEffects) AdjustedNoise() float64 {
	// RMS voltage noise scales as sqrt(T)
	return math.Sqrt(t.AmbientK / 300.0)
}

// AdjustedSwitchingEnergy estimates switching energy scaling with temperature.
// Ferroelectric switching energy scales roughly linearly with coercive field,
// which can vary with temperature.
func (t *TemperatureEffects) AdjustedSwitchingEnergy(baseEnergy float64) float64 {
	// At cryogenic temperatures, coercive field can be slightly lower
	// At high temperatures, enhanced ionic motion can reduce Ec
	// This is a simplified model - actual behavior is material-dependent

	if t.AmbientK < 100 {
		// Slight reduction at cryo
		return baseEnergy * (0.9 + 0.1*t.AmbientK/100)
	}
	if t.AmbientK > 300 {
		// Slight reduction at high temp (but reliability concerns)
		factor := 1.0 - 0.05*(t.AmbientK-300)/100
		if factor < 0.8 {
			factor = 0.8
		}
		return baseEnergy * factor
	}
	return baseEnergy
}

// GetTemperatureLabel returns a human-readable label for the temperature.
func (t *TemperatureEffects) GetTemperatureLabel() string {
	switch {
	case t.AmbientK < 10:
		return "Deep Cryogenic"
	case t.AmbientK < 100:
		return "Cryogenic"
	case t.AmbientK < 273:
		return "Cold"
	case t.AmbientK < 323:
		return "Room Temperature"
	case t.AmbientK < 373:
		return "Industrial"
	case t.AmbientK < 423:
		return "Automotive"
	default:
		return "Extreme Heat"
	}
}

// TemperatureEffectsForMVM returns temperature effects configured for MVM simulation.
// Adjusts wire parameters and provides conductance scaling.
type TemperatureAdjustedParams struct {
	WireResistanceFactor float64 // Multiply nominal wire R by this
	GminAdjusted         float64 // Adjusted minimum conductance
	GmaxAdjusted         float64 // Adjusted maximum conductance
	DriftRateFactor      float64 // Multiply nominal drift rate by this
	NoiseFactor          float64 // Multiply nominal noise by this
	RetentionFactor      float64 // Multiply nominal retention by this
}

// GetAdjustedParams returns all temperature-adjusted parameters.
func (t *TemperatureEffects) GetAdjustedParams() *TemperatureAdjustedParams {
	adjGmin, adjGmax := t.AdjustedConductanceRange(GMin, GMax)

	return &TemperatureAdjustedParams{
		WireResistanceFactor: t.AdjustedWireResistance(1.0), // Factor for 1Ω base
		GminAdjusted:         adjGmin,
		GmaxAdjusted:         adjGmax,
		DriftRateFactor:      t.AdjustedDriftRate(1.0), // Factor for base rate
		NoiseFactor:          t.AdjustedNoise(),
		RetentionFactor:      t.AdjustedRetention(),
	}
}

// ============================================================================
// H15: AUTOMOTIVE THERMAL PHYSICS (25°C to 85°C)
// Per Dr. Tour critique - implement full thermal physics with retention curves
// ============================================================================

// AutomotiveTemperatureGrades defines industry-standard temperature grades.
const (
	TempGrade0Min = 233.0 // -40°C (Grade 0 automotive min)
	TempGrade0Max = 423.0 // +150°C (Grade 0 automotive max)
	TempGrade1Min = 233.0 // -40°C (Grade 1 min)
	TempGrade1Max = 398.0 // +125°C (Grade 1 max)
	TempGrade2Min = 233.0 // -40°C (Grade 2 min)
	TempGrade2Max = 378.0 // +105°C (Grade 2 max)
	TempGrade3Min = 273.0 // 0°C (Grade 3 min)
	TempGrade3Max = 343.0 // +70°C (Grade 3 max)
)

// RetentionCurve represents retention time vs temperature data.
type RetentionCurve struct {
	TemperaturesK []float64 // Temperature points (Kelvin)
	RetentionS    []float64 // Retention time at each temperature (seconds)
	ActivationEV  float64   // Activation energy used for Arrhenius fit (eV)
}

// GenerateRetentionCurve creates a retention curve from 25°C to 85°C.
// Uses Arrhenius extrapolation based on reference retention time.
//
// Physics basis:
//   - Retention follows Arrhenius law: t_ret ∝ exp(Ea/kT)
//   - Reference: 10^7 seconds (116 days) at 85°C demonstrated for FeCIM
//   - Activation energy Ea ≈ 1.0-1.2 eV for HZO (from IEEE IRPS 2022)
//
// Parameters:
//   - refRetentionS: Reference retention time at 85°C (seconds)
//   - activationEV: Activation energy (eV), typically 1.0-1.2 for HZO
func GenerateRetentionCurve(refRetentionS, activationEV float64) *RetentionCurve {
	const kB = 8.617e-5    // Boltzmann constant in eV/K
	const refTempK = 358.0 // 85°C reference

	// Temperature points from 25°C to 150°C
	temps := []float64{
		298, // 25°C (room temp)
		313, // 40°C
		328, // 55°C
		343, // 70°C
		358, // 85°C (industrial reference)
		373, // 100°C
		398, // 125°C (automotive grade 1)
		423, // 150°C (automotive grade 0)
	}

	retentions := make([]float64, len(temps))

	for i, T := range temps {
		// Arrhenius scaling: t(T) = t_ref * exp(Ea/k * (1/T - 1/T_ref))
		exponent := (activationEV / kB) * (1/T - 1/refTempK)
		retentions[i] = refRetentionS * math.Exp(exponent)
	}

	return &RetentionCurve{
		TemperaturesK: temps,
		RetentionS:    retentions,
		ActivationEV:  activationEV,
	}
}

// DefaultRetentionCurve returns the FeCIM retention curve based on demonstrated values.
func DefaultRetentionCurve() *RetentionCurve {
	// FeCIM demonstrated: 10^7 seconds at 85°C
	// Using Ea = 1.1 eV (middle of literature range for HZO)
	return GenerateRetentionCurve(1e7, 1.1)
}

// RetentionAt returns interpolated retention time at a given temperature.
func (rc *RetentionCurve) RetentionAt(tempK float64) float64 {
	const kB = 8.617e-5

	// Use Arrhenius formula directly from reference point at 85°C
	refTempK := 358.0
	refRetention := rc.RetentionS[4] // 85°C index

	exponent := (rc.ActivationEV / kB) * (1/tempK - 1/refTempK)
	return refRetention * math.Exp(exponent)
}

// RetentionYearsAt returns retention time in years at given temperature.
func (rc *RetentionCurve) RetentionYearsAt(tempK float64) float64 {
	secondsPerYear := 365.25 * 24 * 3600
	return rc.RetentionAt(tempK) / secondsPerYear
}

// MeetsAutomotiveGrade checks if retention meets a specific automotive grade.
func (rc *RetentionCurve) MeetsAutomotiveGrade(grade int, requiredYears float64) bool {
	var maxTempK float64
	switch grade {
	case 0:
		maxTempK = TempGrade0Max
	case 1:
		maxTempK = TempGrade1Max
	case 2:
		maxTempK = TempGrade2Max
	case 3:
		maxTempK = TempGrade3Max
	default:
		maxTempK = TempGrade1Max
	}

	return rc.RetentionYearsAt(maxTempK) >= requiredYears
}

// ThermalPhysicsModel provides comprehensive temperature-dependent physics.
type ThermalPhysicsModel struct {
	Material       string  // Material name
	CurieTempK     float64 // Curie temperature (K)
	RefPr          float64 // Reference Pr at 300K (C/m²)
	RefEc          float64 // Reference Ec at 300K (V/m)
	ActivationEV   float64 // Activation energy for retention (eV)
	RetentionCurve *RetentionCurve
}

// NewThermalPhysicsModel creates a thermal physics model for a material.
func NewThermalPhysicsModel(material string, curieTempK, refPr, refEc, activationEV, refRetentionS float64) *ThermalPhysicsModel {
	return &ThermalPhysicsModel{
		Material:       material,
		CurieTempK:     curieTempK,
		RefPr:          refPr,
		RefEc:          refEc,
		ActivationEV:   activationEV,
		RetentionCurve: GenerateRetentionCurve(refRetentionS, activationEV),
	}
}

// DefaultHZOThermalModel returns the thermal model for standard HZO.
func DefaultHZOThermalModel() *ThermalPhysicsModel {
	return NewThermalPhysicsModel(
		"HZO",
		723,   // Curie temp ~450°C
		25e-2, // Pr = 25 µC/cm²
		1.0e8, // Ec = 1.0 MV/cm
		1.1,   // Ea = 1.1 eV
		1e7,   // 10^7 s retention at 85°C
	)
}

// PrAtTemperature returns temperature-dependent remanent polarization.
// Uses empirical relation: Pr(T) = Pr(0) * (1 - T/Tc)^β
// where β ≈ 0.5 for typical ferroelectrics.
func (m *ThermalPhysicsModel) PrAtTemperature(tempK float64) float64 {
	if tempK >= m.CurieTempK {
		return 0 // Above Curie temperature
	}

	// Reference Pr is at 300K, not 0K
	// Pr(300K) = Pr(0K) * (1 - 300/Tc)^0.5
	// So Pr(0K) = Pr(300K) / (1 - 300/Tc)^0.5
	refFactor := math.Pow(1-300/m.CurieTempK, 0.5)
	Pr0 := m.RefPr / refFactor

	return Pr0 * math.Pow(1-tempK/m.CurieTempK, 0.5)
}

// EcAtTemperature returns temperature-dependent coercive field.
// Ec typically decreases with temperature due to thermal activation.
// Uses: Ec(T) = Ec(0) * (1 - T/Tc)^0.5
func (m *ThermalPhysicsModel) EcAtTemperature(tempK float64) float64 {
	if tempK >= m.CurieTempK {
		return 0
	}

	refFactor := math.Pow(1-300/m.CurieTempK, 0.5)
	Ec0 := m.RefEc / refFactor

	return Ec0 * math.Pow(1-tempK/m.CurieTempK, 0.5)
}

// EffectiveLevelsAtTemperature estimates the number of distinguishable levels.
// Higher temperature → more noise → fewer reliable levels.
func (m *ThermalPhysicsModel) EffectiveLevelsAtTemperature(tempK float64, nominalLevels int) int {
	// At reference (300K), we have nominal levels
	// Noise scales as sqrt(T), so level resolution degrades
	noiseFactor := math.Sqrt(tempK / 300)

	// Also account for Pr reduction
	prFactor := m.PrAtTemperature(tempK) / m.RefPr
	if prFactor < 0.1 {
		prFactor = 0.1 // Minimum
	}

	effectiveLevels := float64(nominalLevels) * prFactor / noiseFactor
	result := int(math.Round(effectiveLevels))

	if result < 2 {
		result = 2 // Minimum binary
	}
	if result > nominalLevels {
		result = nominalLevels
	}

	return result
}

// GetAutomotiveReport generates a summary of automotive grade compliance.
func (m *ThermalPhysicsModel) GetAutomotiveReport(nominalLevels int) AutomotiveComplianceReport {
	return AutomotiveComplianceReport{
		Material: m.Material,

		// Grade 0: -40°C to +150°C
		Grade0LevelsAt150C: m.EffectiveLevelsAtTemperature(423, nominalLevels),
		Grade0RetentionYrs: m.RetentionCurve.RetentionYearsAt(423),
		Grade0Pass10Year:   m.RetentionCurve.MeetsAutomotiveGrade(0, 10),

		// Grade 1: -40°C to +125°C
		Grade1LevelsAt125C: m.EffectiveLevelsAtTemperature(398, nominalLevels),
		Grade1RetentionYrs: m.RetentionCurve.RetentionYearsAt(398),
		Grade1Pass10Year:   m.RetentionCurve.MeetsAutomotiveGrade(1, 10),

		// Industrial: 85°C
		IndustrialLevels:    m.EffectiveLevelsAtTemperature(358, nominalLevels),
		IndustrialRetention: m.RetentionCurve.RetentionYearsAt(358),
	}
}

// AutomotiveComplianceReport summarizes automotive grade compliance.
type AutomotiveComplianceReport struct {
	Material string

	// Grade 0 (AEC-Q100 Grade 0): -40°C to +150°C
	Grade0LevelsAt150C int
	Grade0RetentionYrs float64
	Grade0Pass10Year   bool

	// Grade 1: -40°C to +125°C
	Grade1LevelsAt125C int
	Grade1RetentionYrs float64
	Grade1Pass10Year   bool

	// Industrial: up to 85°C
	IndustrialLevels    int
	IndustrialRetention float64
}
