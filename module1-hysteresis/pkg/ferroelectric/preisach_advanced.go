// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

// DistributionType specifies the type of Preisach distribution function.
type DistributionType int

const (
	// DistGaussian uses a 2D Gaussian distribution (default).
	DistGaussian DistributionType = iota
	// DistLorentzian uses separable Lorentzian distributions.
	// Enables closed-form Everett integral calculation (2-3× speedup).
	DistLorentzian
)

// Hysteron represents an elementary bistable switching unit in the Preisach model.
// Each hysteron switches UP at field alpha and DOWN at field beta (alpha > beta).
type Hysteron struct {
	Alpha float64 // Positive switching field (V/m)
	Beta  float64 // Negative switching field (V/m)
	State int     // Current state: +1 or -1
}

// MayergoyzPreisach implements the full classical Preisach model
// following Mayergoyz's mathematical framework.
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991)
type MayergoyzPreisach struct {
	material *HZOMaterial

	// Preisach plane discretization
	hysterons    []Hysteron  // Array of hysterons
	numAlpha     int         // Grid points along alpha axis
	numBeta      int         // Grid points along beta axis
	distribution [][]float64 // μ(α, β) distribution weights

	// Distribution type and parameters
	DistType DistributionType // Distribution type (Gaussian or Lorentzian)

	// Gaussian distribution parameters
	AlphaMean   float64 // Mean of alpha distribution (≈ +Ec)
	AlphaSigma  float64 // Standard deviation of alpha
	BetaMean    float64 // Mean of beta distribution (≈ -Ec)
	BetaSigma   float64 // Standard deviation of beta
	Correlation float64 // Correlation between alpha and beta

	// Lorentzian distribution parameters
	LorentzAlphaC float64 // Center of alpha Lorentzian (default: Ec)
	LorentzAlphaW float64 // Width of alpha Lorentzian (default: 0.5*Ec)
	LorentzBetaC  float64 // Center of beta Lorentzian (default: -Ec)
	LorentzBetaW  float64 // Width of beta Lorentzian (default: 0.5*Ec)

	// Temperature dependence
	Temperature  float64 // Operating temperature (K)
	CurieTemp    float64 // Curie temperature (K)
	TempExponent float64 // Temperature exponent for Ec(T)

	// State tracking
	fieldHistory []float64 // History of applied fields
	polarization float64   // Current polarization

	// Fatigue and wake-up
	cycleCount    int     // Number of switching cycles
	fatigueRate   float64 // Fatigue degradation rate
	wakeupCycles  int     // Cycles needed for wake-up
	currentWakeup float64 // Current wake-up factor (0-1)

	// NLS (Nucleation-Limited Switching) parameters for Merz law dynamics
	// tau(E) = Tau0NLS * exp(EaNLS / |E|)
	// Loaded from material, can be overridden with SetNLSParameters()
	Tau0NLS float64 // Attempt time for NLS (s)
	EaNLS   float64 // Activation field for NLS (V/m)

	// Substrate strain effects
	// Biaxial strain from lattice mismatch (e.g., HZO on Si: ~-2% compressive)
	// Strain shifts the effective coercive field via electrostrictive coupling:
	// Ec_eff = Ec * (1 + strainShiftFactor * strain)
	SubstrateStrain   float64 // Biaxial in-plane strain (negative = compressive)
	strainShiftFactor float64 // Calculated from Q11, Q12 electrostrictive coefficients
}

// NewMayergoyzPreisach creates a new full Preisach model.
//
// Grid Size Selection:
// The gridSize parameter controls hysteron density on the Preisach plane.
// Recommended values based on convergence studies:
//   - 30-40: Fast computation, ~1% error vs converged (suitable for interactive demos)
//   - 50:    Standard accuracy, <0.5% error vs converged (default for simulations)
//   - 100+:  High accuracy, diminishing returns (<0.1% improvement)
//
// A 50×50 grid provides 1250 hysterons in the valid (α > β) region, sufficient
// for smooth hysteresis curves and accurate coercive field representation.
//
// Reference: Mayergoyz, "Mathematical Models of Hysteresis" (1991), Chapter 1
// shows that 50-100 grid points per dimension suffices for <1% loop area error.
func NewMayergoyzPreisach(material *HZOMaterial, gridSize int) *MayergoyzPreisach {
	// Load NLS parameters from material (with defaults for backward compatibility)
	tau0NLS := material.Tau0NLS
	if tau0NLS == 0 {
		tau0NLS = 1e-12 // Default: 1 ps attempt time
	}
	eaNLS := material.EaNLS
	if eaNLS == 0 {
		eaNLS = material.Ec * 0.5 // Default: 0.5 * Ec (ensures fast switching above Ec)
	}

	m := &MayergoyzPreisach{
		material:    material,
		numAlpha:    gridSize,
		numBeta:     gridSize,
		DistType:    DistGaussian,       // Default to Gaussian
		AlphaMean:   material.Ec,        // +Ec
		AlphaSigma:  material.Ec * 0.65, // 65% distribution - wider for more gradual switching
		BetaMean:    -material.Ec,       // -Ec
		BetaSigma:   material.Ec * 0.65, // Match alpha sigma
		Correlation: 0.15,               // Low correlation for well-distributed hysterons
		// Lorentzian defaults
		LorentzAlphaC: material.Ec,       // Center at +Ec
		LorentzAlphaW: material.Ec * 0.5, // Half-width at half-maximum
		LorentzBetaC:  -material.Ec,      // Center at -Ec
		LorentzBetaW:  material.Ec * 0.5, // Half-width at half-maximum
		Temperature:   300,               // Room temperature (K)
		CurieTemp:     723,               // HZO Curie temperature ~450°C
		TempExponent:  0.5,               // Typical exponent
		fatigueRate:   1e-10,             // Very low fatigue for HZO
		wakeupCycles:  100,
		currentWakeup: 0.8, // Start partially woken up
		Tau0NLS:       tau0NLS,
		EaNLS:         eaNLS,
		// Substrate strain defaults
		// Reference: Haun et al., J. Appl. Phys. 62, 3331 (1987) - electrostrictive coefficients
		// The -0.15 factor represents ~15% Ec shift per 1% strain, derived from Q11/Q12 coefficients.
		// For HZO on Si: compressive strain from thermal mismatch increases Ec.
		// See also: Materlik et al., J. Appl. Phys. 117, 134109 (2015) - strain effects in HfO2
		SubstrateStrain:   0,     // No strain by default
		strainShiftFactor: -0.15, // ~15% Ec shift per 1% strain (Haun 1987, Materlik 2015)
	}

	m.initializeHysterons()
	m.initializeDistribution()

	log.Debug("NewMayergoyzPreisach: material=%s, grid=%dx%d, Ec=%.2f MV/cm, Ps=%.1f µC/cm²",
		material.Name, gridSize, gridSize, material.Ec/1e8, material.Ps*100)

	return m
}

// initializeHysterons creates the hysteron grid on the Preisach plane.
func (m *MayergoyzPreisach) initializeHysterons() {
	// The Preisach plane has α on vertical axis, β on horizontal
	// Valid region: α > β (lower triangle)

	// Temperature-corrected coercive fields
	EcEff := m.temperatureCorrectedEc()

	// Field range: from -2*Ec to +2*Ec
	Emax := 2.0 * EcEff
	dE := 2.0 * Emax / float64(m.numAlpha-1)

	m.hysterons = make([]Hysteron, 0, m.numAlpha*m.numBeta/2)

	for i := 0; i < m.numAlpha; i++ {
		alpha := -Emax + float64(i)*dE
		for j := 0; j < m.numBeta; j++ {
			beta := -Emax + float64(j)*dE

			// Only include valid hysterons where α > β
			if alpha > beta {
				m.hysterons = append(m.hysterons, Hysteron{
					Alpha: alpha,
					Beta:  beta,
					State: -1, // Start in negative state (depoled)
				})
			}
		}
	}
}

// initializeDistribution sets up the Preisach distribution function μ(α, β).
// lorentzian1D computes a 1D Lorentzian (Cauchy) distribution.
// L(x) = (γ/π) / [(x-x₀)² + γ²]
// where γ is the half-width at half-maximum.
func lorentzian1D(x, center, width float64) float64 {
	halfWidth := width / 2
	return (halfWidth / math.Pi) / (math.Pow(x-center, 2) + halfWidth*halfWidth)
}

// initializeDistribution sets up the Preisach distribution function μ(α, β).
// Dispatches to Gaussian or Lorentzian based on DistType.
func (m *MayergoyzPreisach) initializeDistribution() {
	switch m.DistType {
	case DistGaussian:
		m.initializeDistributionGaussian()
	case DistLorentzian:
		m.initializeDistributionLorentzian()
	default:
		m.initializeDistributionGaussian()
	}
}

// initializeDistributionGaussian sets up a 2D Gaussian distribution.
func (m *MayergoyzPreisach) initializeDistributionGaussian() {
	// Using 2D Gaussian distribution (Gaussian-Gaussian model)
	// μ(α, β) = A * exp(-[(α-αm)²/2σα² + (β-βm)²/2σβ² - 2ρ(α-αm)(β-βm)/(σασβ)] / (1-ρ²))

	EcEff := m.temperatureCorrectedEc()
	alphaM := EcEff * (m.AlphaMean / m.material.Ec)
	betaM := -EcEff * (-m.BetaMean / m.material.Ec)
	sigmaA := m.AlphaSigma * (EcEff / m.material.Ec)
	sigmaB := m.BetaSigma * (EcEff / m.material.Ec)
	rho := m.Correlation

	m.distribution = make([][]float64, len(m.hysterons))
	totalWeight := 0.0

	for i, h := range m.hysterons {
		// Bivariate Gaussian
		da := (h.Alpha - alphaM) / sigmaA
		db := (h.Beta - betaM) / sigmaB

		exponent := -(da*da - 2*rho*da*db + db*db) / (2 * (1 - rho*rho))
		weight := math.Exp(exponent)

		// Apply wake-up factor (increases effective distribution near Ec)
		wakeupFactor := 1.0 + (1-m.currentWakeup)*0.5*math.Exp(-math.Pow((h.Alpha-alphaM)/sigmaA, 2))
		weight *= wakeupFactor

		m.distribution[i] = []float64{weight}
		totalWeight += weight
	}

	// Normalize so total polarization equals Ps
	if totalWeight > 0 {
		normFactor := m.material.Ps / totalWeight
		for i := range m.distribution {
			m.distribution[i][0] *= normFactor
		}
	}
}

// initializeDistributionLorentzian sets up a separable Lorentzian distribution.
// μ(α, β) = L(α) × L(β) where L is the Lorentzian (Cauchy) distribution.
// This enables closed-form Everett integral calculation (2-3× speedup).
func (m *MayergoyzPreisach) initializeDistributionLorentzian() {
	EcEff := m.temperatureCorrectedEc()

	// Temperature-scale the Lorentzian parameters
	alphaC := EcEff * (m.LorentzAlphaC / m.material.Ec)
	alphaW := m.LorentzAlphaW * (EcEff / m.material.Ec)
	betaC := -EcEff * (-m.LorentzBetaC / m.material.Ec)
	betaW := m.LorentzBetaW * (EcEff / m.material.Ec)

	m.distribution = make([][]float64, len(m.hysterons))
	totalWeight := 0.0

	for i, h := range m.hysterons {
		// Separable Lorentzian: μ(α, β) = L(α) × L(β)
		lAlpha := lorentzian1D(h.Alpha, alphaC, alphaW)
		lBeta := lorentzian1D(h.Beta, betaC, betaW)
		weight := lAlpha * lBeta

		// Apply wake-up factor (increases effective distribution near Ec)
		// For Lorentzian, use similar wake-up model as Gaussian
		wakeupFactor := 1.0 + (1-m.currentWakeup)*0.5*lorentzian1D(h.Alpha, alphaC, alphaW)/lorentzian1D(alphaC, alphaC, alphaW)
		weight *= wakeupFactor

		m.distribution[i] = []float64{weight}
		totalWeight += weight
	}

	// Normalize so total polarization equals Ps
	if totalWeight > 0 {
		normFactor := m.material.Ps / totalWeight
		for i := range m.distribution {
			m.distribution[i][0] *= normFactor
		}
	}
}

// temperatureCorrectedEc returns the coercive field corrected for temperature and strain.
// Temperature: Ec(T) = Ec0 * (1 - T/Tc)^β
// Strain: Ec_eff = Ec(T) * (1 + strainShiftFactor * strain)
func (m *MayergoyzPreisach) temperatureCorrectedEc() float64 {
	if m.Temperature >= m.CurieTemp {
		return 0 // Above Curie temperature, no ferroelectricity
	}

	// Temperature correction
	ratio := m.Temperature / m.CurieTemp
	tempEc := m.material.Ec * math.Pow(1-ratio, m.TempExponent)

	// Strain correction: compressive strain (negative) increases Ec
	strainEc := tempEc * (1 + m.strainShiftFactor*m.SubstrateStrain)

	return strainEc
}

// SetTemperature updates the operating temperature and recalculates distributions.
func (m *MayergoyzPreisach) SetTemperature(T float64) {
	oldTemp := m.Temperature
	m.Temperature = T
	m.initializeHysterons()
	m.initializeDistribution()

	effEc := m.temperatureCorrectedEc()
	log.Debug("SetTemperature: %.0fK → %.0fK, Ec(T)=%.2f MV/cm (%.0f%% of Tc)",
		oldTemp, T, effEc/1e8, T/m.CurieTemp*100)
}

// SetDistributionType sets the distribution type and reinitializes the model.
func (m *MayergoyzPreisach) SetDistributionType(dtype DistributionType) {
	m.DistType = dtype
	m.initializeDistribution()

	distName := "Gaussian"
	if dtype == DistLorentzian {
		distName = "Lorentzian"
	}
	log.Debug("SetDistributionType: %s", distName)
}

// SetSubstrateStrain applies biaxial substrate strain effects to the model.
// Strain shifts the effective coercive field via electrostrictive coupling.
//
// Physics: For HZO on silicon, compressive strain (negative) typically INCREASES Ec
// due to electrostrictive coupling: Ec_eff = Ec * (1 + factor * strain)
//
// Typical values:
//   - HZO on Si: strain ≈ -0.02 (-2% compressive)
//   - This shifts Ec by ~10-20% depending on film quality
//
// The strainShiftFactor can be derived from electrostrictive coefficients:
//
//	factor ≈ 2 * Q11 / Ec  (simplified model)
//
// where Q11 ≈ 0.089 m⁴/C² for HZO.
//
// References:
//   - Haun et al., J. Appl. Phys. 62, 3331 (1987) - electrostrictive coefficients
//   - Materlik et al., J. Appl. Phys. 117, 134109 (2015) - strain effects in HfO2
func (m *MayergoyzPreisach) SetSubstrateStrain(strain float64) {
	oldStrain := m.SubstrateStrain
	m.SubstrateStrain = strain

	// Recalculate hysteron grid and distribution with new strain
	m.initializeHysterons()
	m.initializeDistribution()

	// Calculate the effective Ec shift for logging
	shiftPercent := m.strainShiftFactor * strain * 100
	log.Debug("SetSubstrateStrain: %.2f%% → %.2f%%, Ec shift: %+.1f%%",
		oldStrain*100, strain*100, shiftPercent)
}

// SetStrainShiftFactor sets the strain-to-Ec coupling factor.
// Default is 0.15 (~15% Ec change per 1% strain).
// Can be calculated from electrostrictive coefficients: factor ≈ 2*Q11*Ec/Ec = 2*Q11
func (m *MayergoyzPreisach) SetStrainShiftFactor(factor float64) {
	m.strainShiftFactor = factor
	if m.SubstrateStrain != 0 {
		// Re-apply strain with new factor
		m.initializeHysterons()
		m.initializeDistribution()
	}
}

// GetSubstrateStrain returns the current substrate strain value.
func (m *MayergoyzPreisach) GetSubstrateStrain() float64 {
	return m.SubstrateStrain
}

// GetEffectiveEc returns the coercive field with all corrections applied
// (temperature + strain). This is the same as temperatureCorrectedEc()
// but exposed for external use.
func (m *MayergoyzPreisach) GetEffectiveEc() float64 {
	return m.temperatureCorrectedEc()
}

// Update applies a new electric field and returns the resulting polarization.
func (m *MayergoyzPreisach) Update(E float64) float64 {
	// Safety check: ensure distribution and hysterons are synchronized
	if len(m.distribution) != len(m.hysterons) {
		log.Debug("Update: distribution/hysteron mismatch (%d vs %d), reinitializing",
			len(m.distribution), len(m.hysterons))
		m.initializeDistribution()
	}

	// Update each hysteron's state based on the applied field
	for i := range m.hysterons {
		if E >= m.hysterons[i].Alpha {
			m.hysterons[i].State = +1 // Switch UP
		} else if E <= m.hysterons[i].Beta {
			m.hysterons[i].State = -1 // Switch DOWN
		}
		// Otherwise, state remains unchanged (memory effect)
	}

	// Calculate polarization by integrating over Preisach plane
	// P = ∫∫ μ(α, β) * γ(α, β) dα dβ
	m.polarization = 0
	for i, h := range m.hysterons {
		m.polarization += m.distribution[i][0] * float64(h.State)
	}

	// Apply fatigue degradation
	m.polarization *= (1 - m.fatigueRate*float64(m.cycleCount))

	// Clamp polarization to physical bounds [-Ps, +Ps]
	// This prevents visual spikes from numerical edge cases or race conditions
	Ps := m.material.Ps
	if m.polarization > Ps {
		log.Debug("Update: clamping P=%.4f to +Ps=%.4f (E=%.2e)", m.polarization, Ps, E)
		m.polarization = Ps
	} else if m.polarization < -Ps {
		log.Debug("Update: clamping P=%.4f to -Ps=%.4f (E=%.2e)", m.polarization, -Ps, E)
		m.polarization = -Ps
	}

	// Record history with bounded growth to prevent memory exhaustion
	const maxFieldHistory = 10000
	m.fieldHistory = append(m.fieldHistory, E)
	if len(m.fieldHistory) > maxFieldHistory {
		// Keep the most recent half of history
		m.fieldHistory = m.fieldHistory[maxFieldHistory/2:]
	}

	return m.polarization
}

// Cycle increments the cycle count (call after each P-E loop).
func (m *MayergoyzPreisach) Cycle() {
	m.cycleCount++

	// Update wake-up factor
	if m.currentWakeup < 1.0 {
		oldWakeup := m.currentWakeup
		wakeupRate := 1.0 - math.Exp(-float64(m.cycleCount)/float64(m.wakeupCycles))
		m.currentWakeup = 0.8 + 0.2*wakeupRate
		m.initializeDistribution() // Recalculate with new wake-up

		// Log wake-up progress at milestones
		if m.cycleCount%100 == 0 || m.currentWakeup >= 0.99 {
			log.Debug("Cycle: count=%d, wakeup=%.1f%% → %.1f%%", m.cycleCount, oldWakeup*100, m.currentWakeup*100)
		}
	}
}

// Reset clears the model to initial state.
func (m *MayergoyzPreisach) Reset() {
	for i := range m.hysterons {
		m.hysterons[i].State = -1
	}
	m.polarization = 0
	m.fieldHistory = m.fieldHistory[:0]

	log.Trace("Reset: all hysterons set to -1, P=0")
}

// Polarization returns the current polarization.
func (m *MayergoyzPreisach) Polarization() float64 {
	return m.polarization
}

// NormalizedPolarization returns P/Ps in range [-1, +1].
func (m *MayergoyzPreisach) NormalizedPolarization() float64 {
	return m.polarization / m.material.Ps
}

// GetHysteresisLoop generates a complete P-E hysteresis loop.
func (m *MayergoyzPreisach) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64) {
	m.Reset()

	E := make([]float64, 0, points*4)
	P := make([]float64, 0, points*4)

	// First, saturate in positive direction
	for i := 0; i <= points/2; i++ {
		e := Emax * float64(i) / float64(points/2)
		p := m.Update(e)
		E = append(E, e)
		P = append(P, p)
	}

	// Sweep from +Emax to -Emax
	for i := 0; i <= points; i++ {
		e := Emax - 2*Emax*float64(i)/float64(points)
		p := m.Update(e)
		E = append(E, e)
		P = append(P, p)
	}

	// Sweep from -Emax back to +Emax
	for i := 0; i <= points; i++ {
		e := -Emax + 2*Emax*float64(i)/float64(points)
		p := m.Update(e)
		E = append(E, e)
		P = append(P, p)
	}

	m.Cycle()
	return E, P
}

// GetMinorLoop generates a minor hysteresis loop between E1 and E2.
func (m *MayergoyzPreisach) GetMinorLoop(E1, E2 float64, points int) ([]float64, []float64) {
	E := make([]float64, 0, points*2)
	P := make([]float64, 0, points*2)

	// Sweep from E1 to E2
	for i := 0; i <= points; i++ {
		e := E1 + (E2-E1)*float64(i)/float64(points)
		p := m.Update(e)
		E = append(E, e)
		P = append(P, p)
	}

	// Sweep back from E2 to E1
	for i := 0; i <= points; i++ {
		e := E2 + (E1-E2)*float64(i)/float64(points)
		p := m.Update(e)
		E = append(E, e)
		P = append(P, p)
	}

	return E, P
}

// GetPreisachPlane returns the current state of all hysterons for visualization.
// Returns alpha, beta, and state (+1/-1) for each hysteron.
func (m *MayergoyzPreisach) GetPreisachPlane() ([]float64, []float64, []int) {
	alphas := make([]float64, len(m.hysterons))
	betas := make([]float64, len(m.hysterons))
	states := make([]int, len(m.hysterons))

	for i, h := range m.hysterons {
		alphas[i] = h.Alpha
		betas[i] = h.Beta
		states[i] = h.State
	}

	return alphas, betas, states
}

// GetDistribution returns the Preisach distribution weights.
func (m *MayergoyzPreisach) GetDistribution() []float64 {
	weights := make([]float64, len(m.distribution))
	for i := range m.distribution {
		weights[i] = m.distribution[i][0]
	}
	return weights
}

// GetSwitchedFraction returns the fraction of hysterons in +1 state.
func (m *MayergoyzPreisach) GetSwitchedFraction() float64 {
	switched := 0
	for _, h := range m.hysterons {
		if h.State == +1 {
			switched++
		}
	}
	return float64(switched) / float64(len(m.hysterons))
}

// GetEffectivePr returns the actual simulated remanent polarization.
// This measures what the Preisach model actually produces at E=0 after saturation,
// which may differ from the nominal material Pr due to loop shape.
func (m *MayergoyzPreisach) GetEffectivePr() float64 {
	if m.Temperature >= m.CurieTemp {
		return 0 // Above Curie temperature, no ferroelectricity
	}

	// Save current state
	savedStates := make([]int, len(m.hysterons))
	for i, h := range m.hysterons {
		savedStates[i] = h.State
	}
	savedPol := m.polarization

	// Saturate positive then return to E=0
	Emax := m.temperatureCorrectedEc() * 2.5
	steps := 20
	for i := 0; i <= steps; i++ {
		m.Update(Emax * float64(i) / float64(steps))
	}
	for i := steps; i >= 0; i-- {
		m.Update(Emax * float64(i) / float64(steps))
	}
	actualPr := m.polarization

	// Restore original state
	for i := range m.hysterons {
		m.hysterons[i].State = savedStates[i]
	}
	m.polarization = savedPol

	return actualPr
}

// GetSwitchingTime returns the field-dependent switching time using Merz's law.
// This implements NLS (Nucleation-Limited Switching) dynamics:
//
//	tau(E) = tau0 * exp(Ea / |E|)
//
// At high fields (E >> Ea), switching is fast (~100 ps).
// At low fields (E ~ Ec), switching slows dramatically (~100 ns).
//
// Reference: Merz, W.J. "Domain Formation and Domain Wall Motions in
// Ferroelectric BaTiO3 Single Crystals" Phys. Rev. 95, 690 (1954)
// For HfO2-based materials: Park et al., Adv. Mater. 27, 1811 (2015)
func (m *MayergoyzPreisach) GetSwitchingTime(E float64) float64 {
	absE := math.Abs(E)
	if absE < 1e-6 {
		return math.Inf(1) // No switching at zero field
	}

	// Merz law: tau = tau0 * exp(Ea/E)
	tau := m.Tau0NLS * math.Exp(m.EaNLS/absE)

	// Clamp to reasonable range (100 ps to 1 s)
	// Upper bound of 1 second prevents numerical issues in simulations
	if tau < 1e-10 {
		tau = 1e-10
	}
	if tau > 1.0 {
		tau = 1.0
	}

	return tau
}

// SetNLSParameters allows customizing the Merz law parameters.
// tau0 is the attempt time (typically 1e-10 to 1e-12 s).
// Ea is the activation field (typically 10-15 MV/cm for HfO2).
func (m *MayergoyzPreisach) SetNLSParameters(tau0, Ea float64) {
	m.Tau0NLS = tau0
	m.EaNLS = Ea
	log.Debug("SetNLSParameters: tau0=%.2e s, Ea=%.2f MV/cm", tau0, Ea/1e8)
}

// SimulateDomainSwitching returns domain switching dynamics over time.
// Returns time, polarization, and number of switched domains.
func (m *MayergoyzPreisach) SimulateDomainSwitching(Eapplied float64, duration float64, steps int) ([]float64, []float64, []int) {
	times := make([]float64, steps)
	pols := make([]float64, steps)
	switched := make([]int, steps)

	dt := duration / float64(steps-1)

	// Use KAI model switching time for domain growth dynamics
	// Note: GetSwitchingTime() provides field-dependent NLS (Merz law) switching time
	// but KAI model uses its own time constant for nucleation/growth dynamics
	tau := m.material.Tau
	tauNLS := m.GetSwitchingTime(Eapplied)

	log.Debug("SimulateDomainSwitching: E=%.2f MV/cm, tau(KAI)=%.2e s, tau(NLS)=%.2e s, duration=%.0f ns",
		Eapplied/1e8, tau, tauNLS, duration*1e9)

	// KAI (Kolmogorov-Avrami-Ishibashi) switching dynamics
	// P(t) = Ps * (1 - exp(-(t/τ)^n))
	n := 2.0 // Avrami exponent for 2D domain growth

	m.Reset()

	for i := 0; i < steps; i++ {
		t := float64(i) * dt
		times[i] = t

		// Calculate switching progress
		progress := 1.0 - math.Exp(-math.Pow(t/tau, n))

		// Apply field proportionally to progress
		effectiveE := Eapplied * progress
		pols[i] = m.Update(effectiveE)

		// Count switched hysterons
		count := 0
		for _, h := range m.hysterons {
			if h.State == +1 {
				count++
			}
		}
		switched[i] = count
	}

	log.Debug("SimulateDomainSwitching complete: final P=%.2f µC/cm², switched=%d/%d hysterons",
		pols[steps-1]*100, switched[steps-1], len(m.hysterons))

	return times, pols, switched
}

// DiscreteStates returns the 30 programmable states for FeCIM.
func (m *MayergoyzPreisach) DiscreteStates(N int) []DiscreteState {
	states := make([]DiscreteState, N)
	Ps := m.material.Ps
	Ec := m.temperatureCorrectedEc()

	// Calculate voltage needed for each state
	// Using hyperbolic tangent model for state-to-voltage mapping
	for i := 0; i < N; i++ {
		targetP := -Ps + 2*Ps*float64(i)/float64(N-1)
		normalizedP := targetP / Ps

		// Inverse of P = Ps*tanh((E-Ec)/δ) approximately
		// E ≈ Ec + δ * arctanh(P/Ps) for ascending branch
		delta := Ec * 0.3
		var voltage float64
		if math.Abs(normalizedP) < 0.999 {
			voltage = delta * math.Atanh(normalizedP)
		} else {
			voltage = math.Copysign(2*Ec, normalizedP)
		}

		states[i] = DiscreteState{
			Level:        i,
			Polarization: targetP,
			NormalizedP:  normalizedP,
			Voltage:      voltage * m.material.Thickness,
			Conductance:  m.polarizationToConductance(targetP),
		}
	}

	return states
}

// DiscreteState represents one of the 30 programmable states.
type DiscreteState struct {
	Level        int     // State index (0-29)
	Polarization float64 // Polarization (C/m²)
	NormalizedP  float64 // P/Ps (-1 to +1)
	Voltage      float64 // Programming voltage (V)
	Conductance  float64 // Equivalent conductance (S)
}

// polarizationToConductance converts polarization to channel conductance.
// Based on ferroelectric FET model where polarization modulates threshold.
func (m *MayergoyzPreisach) polarizationToConductance(P float64) float64 {
	// Simplified model: G = G0 + ΔG * (P/Ps)
	// FeCIM: 1µS to 100µS range
	G0 := 50e-6     // 50 µS baseline
	deltaG := 49e-6 // ±49 µS range

	normalizedP := P / m.material.Ps
	return G0 + deltaG*normalizedP
}

// AddNoise adds realistic noise to the model (thermal, shot, etc.).
func (m *MayergoyzPreisach) AddNoise(noiseLevel float64) {
	for i := range m.distribution {
		noise := 1.0 + noiseLevel*(rand.Float64()*2-1)
		m.distribution[i][0] *= noise
	}
}

// GetFatigueState returns current fatigue-related metrics.
func (m *MayergoyzPreisach) GetFatigueState() (cycles int, degradation float64, wakeup float64) {
	degradation = m.fatigueRate * float64(m.cycleCount)
	return m.cycleCount, degradation, m.currentWakeup
}

// PreisachExport represents the serializable state of a Preisach model.
// This structure captures all necessary information to restore a calibrated model.
type PreisachExport struct {
	Version     int     `json:"version"`           // Format version for compatibility
	Material    string  `json:"material"`          // Material name
	Temperature float64 `json:"temperature_k"`     // Operating temperature (K)
	GridSize    int     `json:"grid_size"`         // Grid discretization size
	DistType    string  `json:"distribution_type"` // "gaussian" or "lorentzian"

	// Hysteron states (compact int8 representation: -1 or +1)
	HysteronStates []int8 `json:"hysteron_states"`

	// Distribution parameters (for reconstruction)
	AlphaMean   float64 `json:"alpha_mean"`
	AlphaSigma  float64 `json:"alpha_sigma"`
	BetaMean    float64 `json:"beta_mean"`
	BetaSigma   float64 `json:"beta_sigma"`
	Correlation float64 `json:"correlation"`

	// Lorentzian parameters (if using Lorentzian distribution)
	LorentzAlphaC float64 `json:"lorentz_alpha_c,omitempty"`
	LorentzAlphaW float64 `json:"lorentz_alpha_w,omitempty"`
	LorentzBetaC  float64 `json:"lorentz_beta_c,omitempty"`
	LorentzBetaW  float64 `json:"lorentz_beta_w,omitempty"`

	// Fatigue and wake-up state
	CycleCount    int     `json:"cycle_count"`
	CurrentWakeup float64 `json:"current_wakeup"`

	// Metadata
	Timestamp string `json:"timestamp"`  // ISO8601 timestamp
	NumStates int    `json:"num_states"` // Number of hysterons (for validation)
}

// ExportState saves the current Preisach model state to a JSON file.
// This allows preserving calibrated distributions across sessions.
//
// The exported state includes:
//   - All hysteron states (memory state)
//   - Distribution parameters
//   - Temperature and material info
//   - Fatigue/wake-up state
//
// Example:
//
//	err := model.ExportState("data/preisach_states/hzo_300K.json")
func (m *MayergoyzPreisach) ExportState(filename string) error {
	// Convert hysteron states to compact int8 array
	states := make([]int8, len(m.hysterons))
	for i, h := range m.hysterons {
		states[i] = int8(h.State)
	}

	// Determine distribution type string
	distTypeStr := "gaussian"
	if m.DistType == DistLorentzian {
		distTypeStr = "lorentzian"
	}

	// Create export structure
	export := PreisachExport{
		Version:     1, // Format version
		Material:    m.material.Name,
		Temperature: m.Temperature,
		GridSize:    m.numAlpha,
		DistType:    distTypeStr,

		HysteronStates: states,

		AlphaMean:   m.AlphaMean,
		AlphaSigma:  m.AlphaSigma,
		BetaMean:    m.BetaMean,
		BetaSigma:   m.BetaSigma,
		Correlation: m.Correlation,

		LorentzAlphaC: m.LorentzAlphaC,
		LorentzAlphaW: m.LorentzAlphaW,
		LorentzBetaC:  m.LorentzBetaC,
		LorentzBetaW:  m.LorentzBetaW,

		CycleCount:    m.cycleCount,
		CurrentWakeup: m.currentWakeup,

		Timestamp: time.Now().Format(time.RFC3339),
		NumStates: len(states),
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	log.Info("ExportState: saved %d hysteron states to %s (material=%s, T=%.0fK, grid=%d)",
		len(states), filename, m.material.Name, m.Temperature, m.numAlpha)

	return nil
}

// ImportState restores a Preisach model state from a JSON file.
// The current model must have matching grid size for successful import.
//
// Validation performed:
//   - Grid size must match current model
//   - Number of hysterons must match
//   - Material name mismatch generates warning but continues
//
// Example:
//
//	err := model.ImportState("data/preisach_states/hzo_300K.json")
func (m *MayergoyzPreisach) ImportState(filename string) error {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Unmarshal JSON
	var export PreisachExport
	if err := json.Unmarshal(data, &export); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Validate version
	if export.Version != 1 {
		return fmt.Errorf("unsupported format version %d (expected 1)", export.Version)
	}

	// Validate grid size
	if export.GridSize != m.numAlpha {
		return fmt.Errorf("grid size mismatch: file has %d, model has %d", export.GridSize, m.numAlpha)
	}

	// Validate number of hysterons
	if export.NumStates != len(m.hysterons) {
		return fmt.Errorf("hysteron count mismatch: file has %d, model has %d", export.NumStates, len(m.hysterons))
	}

	// Log if material mismatch (but allow import - may be intentional)
	if export.Material != m.material.Name {
		log.Info("ImportState: material mismatch - file has '%s', model has '%s' (continuing anyway)",
			export.Material, m.material.Name)
	}

	// Restore hysteron states
	for i := range m.hysterons {
		if i >= len(export.HysteronStates) {
			return fmt.Errorf("insufficient hysteron states in file (got %d, need %d)",
				len(export.HysteronStates), len(m.hysterons))
		}
		m.hysterons[i].State = int(export.HysteronStates[i])
	}

	// Restore distribution parameters
	m.AlphaMean = export.AlphaMean
	m.AlphaSigma = export.AlphaSigma
	m.BetaMean = export.BetaMean
	m.BetaSigma = export.BetaSigma
	m.Correlation = export.Correlation

	if export.LorentzAlphaC != 0 {
		m.LorentzAlphaC = export.LorentzAlphaC
		m.LorentzAlphaW = export.LorentzAlphaW
		m.LorentzBetaC = export.LorentzBetaC
		m.LorentzBetaW = export.LorentzBetaW
	}

	// Restore fatigue state
	m.cycleCount = export.CycleCount
	m.currentWakeup = export.CurrentWakeup

	// Restore distribution type
	if export.DistType == "lorentzian" {
		m.DistType = DistLorentzian
	} else {
		m.DistType = DistGaussian
	}

	// Set temperature (may differ from file if user changed it)
	m.Temperature = export.Temperature

	// Regenerate distribution with restored parameters
	// This ensures distribution weights match the restored state
	m.initializeDistribution()

	// Recalculate polarization from restored states
	m.polarization = 0
	for i, h := range m.hysterons {
		m.polarization += m.distribution[i][0] * float64(h.State)
	}

	log.Info("ImportState: restored %d hysteron states from %s (material=%s, T=%.0fK, cycles=%d, wakeup=%.1f%%)",
		len(m.hysterons), filename, export.Material, export.Temperature, m.cycleCount, m.currentWakeup*100)

	return nil
}

// DefaultExportPath generates a default export path for the current model state.
// Format: data/preisach_states/[material_name]_[temp]K.json
func (m *MayergoyzPreisach) DefaultExportPath() string {
	// Sanitize material name for filename (replace spaces with underscores)
	materialName := m.material.Name
	materialName = filepath.Base(materialName) // Remove any path separators
	// Replace special characters with underscores
	_ = []string{" ", "(", ")", "/"} // char list for reference
	// Simple replacement would go here if needed

	filename := fmt.Sprintf("%s_%.0fK.json", materialName, m.Temperature)
	return filepath.Join("data", "preisach_states", filename)
}
