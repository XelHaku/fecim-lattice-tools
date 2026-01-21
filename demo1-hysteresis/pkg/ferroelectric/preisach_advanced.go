// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"math"
	"math/rand"
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

	// Distribution parameters (Gaussian model)
	AlphaMean   float64 // Mean of alpha distribution (≈ +Ec)
	AlphaSigma  float64 // Standard deviation of alpha
	BetaMean    float64 // Mean of beta distribution (≈ -Ec)
	BetaSigma   float64 // Standard deviation of beta
	Correlation float64 // Correlation between alpha and beta

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
}

// NewMayergoyzPreisach creates a new full Preisach model.
func NewMayergoyzPreisach(material *HZOMaterial, gridSize int) *MayergoyzPreisach {
	m := &MayergoyzPreisach{
		material:      material,
		numAlpha:      gridSize,
		numBeta:       gridSize,
		AlphaMean:     material.Ec,       // +Ec
		AlphaSigma:    material.Ec * 0.2, // 20% distribution
		BetaMean:      -material.Ec,      // -Ec
		BetaSigma:     material.Ec * 0.2,
		Correlation:   0.5,   // Some correlation between α and β
		Temperature:   300,   // Room temperature (K)
		CurieTemp:     723,   // HZO Curie temperature ~450°C
		TempExponent:  0.5,   // Typical exponent
		fatigueRate:   1e-10, // Very low fatigue for HZO
		wakeupCycles:  100,
		currentWakeup: 0.8, // Start partially woken up
	}

	m.initializeHysterons()
	m.initializeDistribution()

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
func (m *MayergoyzPreisach) initializeDistribution() {
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

// temperatureCorrectedEc returns the coercive field at current temperature.
// Ec(T) = Ec0 * (1 - T/Tc)^β
func (m *MayergoyzPreisach) temperatureCorrectedEc() float64 {
	if m.Temperature >= m.CurieTemp {
		return 0 // Above Curie temperature, no ferroelectricity
	}

	ratio := m.Temperature / m.CurieTemp
	return m.material.Ec * math.Pow(1-ratio, m.TempExponent)
}

// SetTemperature updates the operating temperature and recalculates distributions.
func (m *MayergoyzPreisach) SetTemperature(T float64) {
	m.Temperature = T
	m.initializeHysterons()
	m.initializeDistribution()
}

// Update applies a new electric field and returns the resulting polarization.
func (m *MayergoyzPreisach) Update(E float64) float64 {
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

	// Record history
	m.fieldHistory = append(m.fieldHistory, E)

	return m.polarization
}

// Cycle increments the cycle count (call after each P-E loop).
func (m *MayergoyzPreisach) Cycle() {
	m.cycleCount++

	// Update wake-up factor
	if m.currentWakeup < 1.0 {
		wakeupRate := 1.0 - math.Exp(-float64(m.cycleCount)/float64(m.wakeupCycles))
		m.currentWakeup = 0.8 + 0.2*wakeupRate
		m.initializeDistribution() // Recalculate with new wake-up
	}
}

// Reset clears the model to initial state.
func (m *MayergoyzPreisach) Reset() {
	for i := range m.hysterons {
		m.hysterons[i].State = -1
	}
	m.polarization = 0
	m.fieldHistory = m.fieldHistory[:0]
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

// GetEffectiveEc returns the temperature-corrected coercive field.
func (m *MayergoyzPreisach) GetEffectiveEc() float64 {
	return m.temperatureCorrectedEc()
}

// SimulateDomainSwitching returns domain switching dynamics over time.
// Returns time, polarization, and number of switched domains.
func (m *MayergoyzPreisach) SimulateDomainSwitching(Eapplied float64, duration float64, steps int) ([]float64, []float64, []int) {
	times := make([]float64, steps)
	pols := make([]float64, steps)
	switched := make([]int, steps)

	dt := duration / float64(steps-1)
	tau := m.material.Tau // Switching time constant

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
