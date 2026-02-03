package physics

import (
	"math"
	"math/rand"

	"fecim-lattice-tools/shared/logging"
)

// Package-level logger for Landau-Khalatnikov solver diagnostics.
var lkLog *logging.Logger

// LKSolver implements the First-Order Landau-Khalatnikov dynamics
// for ferroelectric polarization evolution.
type LKSolver struct {
	// Static Material Properties (from calibration)
	Beta  float64 // First-order barrier coefficient (Negative)
	Gamma float64 // Stability coefficient (Positive)
	Rho   float64 // Viscosity / Damping (Ohm-meters)

	// Electrostriction & Stress
	Q12    float64 // Electrostriction coefficient (m^4/C^2)
	Stress float64 // In-plane tensile stress (Pa)

	// Circuit parasitics
	SeriesResistance float64 // Series resistance (Ohms)
	Thickness        float64 // Film thickness (m)
	Area             float64 // Active area (m^2)

	// Depolarization (Polycrystalline Analog Behavior)
	K_dep float64 // Depolarization coefficient (V*m/C) - creates "slant" for analog levels

	// NLS Parameters
	UseNLS          bool
	ActivationField float64 // Activation field (V/m) for Merz's Law
	TauInf          float64 // Infinite field switching time (s)
	IncubationEnd   float64 // Time when switching can start (s),

	// Thermodynamic constants
	CurieTemp  float64 // Curie temperature (K)
	CurieConst float64 // Curie constant (K)

	// Noise (Langevin Dynamics)
	EnableNoise bool

	// Series-resistance aggregation (ρ_eff = ρ + R_series*A/d)
	UseEffectiveViscosity bool

	// Dynamic State
	Alpha       float64 // Calculated stiffness (Temperature + Stress dependent)
	P           float64 // Current Polarization (C/m^2)
	PMax        float64 // Saturation polarization clamp for numerical stability (C/m^2)
	Temperature float64 // Current Temperature (K)
	Time        float64 // Simulation time

	// Internal logging controls
	logCount int
	logLimit int

	// Numerical stability logging (rate-limited)
	nanCount int
	nanLimit int
}

// NewLKSolver creates a new solver with default "Golden Set" parameters for 10nm HZO.
func NewLKSolver() *LKSolver {
	return &LKSolver{
		// Default to "Golden Set" (Set I)
		Beta:   -2.160e8,
		Gamma:  1.653e10,
		Rho:    0.05,
		Q12:    -0.026,
		Stress: 1.0e9, // 1 GPa

		// Circuit parasitics (from hysteresis-gemini.md)
		SeriesResistance: 50.0,          // Ohms
		Thickness:        10e-9,         // 10 nm
		Area:             45e-9 * 45e-9, // 45 nm x 45 nm (FeCIM default cell)

		// Depolarization for Polycrystalline Analog Behavior
		K_dep: 2.5e8, // V*m/C - Default value (matches physics.yaml, within recommended 1-5×10⁸ range)

		UseNLS:          true,
		ActivationField: 1.9e9, // 19 MV/cm (Merz activation field)
		TauInf:          1.0e-13,

		CurieTemp:  723.0,
		CurieConst: 1.5e5,

		EnableNoise:           false,
		UseEffectiveViscosity: true,

		Temperature: 300.0,

		// CRITICAL: Initialize to negative saturation (-Pr)
		// If P0, then E_dep=K_dep*P=0, and depolarization has no effect!
		// This causes binary switching instead of analog slope.
		P:    -0.30, // C/m² (Approximate -Pr for FeCIM HZO at negative saturation)
		PMax: 0.30,  // C/m² (Default clamp, overridden by material config)

		logLimit: 25,
		nanLimit: 5,
	}
}

// UpdateParams recalculates Alpha based on current Temperature and Stress using
// the Unified Coefficient Formula: alpha = alpha_t(T) - 2*Q12*Stress
func (s *LKSolver) UpdateParams() {
	const (
		Eps0 = 8.854e-12 // Vacuum Permittivity (F/m)
	)

	// Thermodynamic contribution (Curie-Weiss)
	alphaT := (s.Temperature - s.CurieTemp) / (2 * Eps0 * s.CurieConst)

	// Mechanical contribution (Electrostriction)
	// Alpha(T,σ) = (T-Tc)/(2ε0C) - 2*Q12*σ
	alphaMech := 2 * s.Q12 * s.Stress

	s.Alpha = alphaT - alphaMech
}

// ConfigureFromMaterial updates solver parameters from HZOMaterial.
// This should be called after NewLKSolver() to override defaults with material-specific values.
// Critical for ensuring the depolarization field (K_dep) matches the material configuration.
func (s *LKSolver) ConfigureFromMaterial(mat *HZOMaterial) {
	if mat == nil {
		return
	}

	if mat.BetaLandau != 0 {
		s.Beta = mat.BetaLandau
	}
	if mat.GammaLandau != 0 {
		s.Gamma = mat.GammaLandau
	}
	if mat.RhoViscosity != 0 {
		s.Rho = mat.RhoViscosity
	}
	if mat.Q12 != 0 {
		s.Q12 = mat.Q12
	}
	if mat.StressGPa != 0 {
		s.Stress = mat.StressGPa * 1e9
	}
	if mat.K_dep > 0 {
		s.K_dep = mat.K_dep
	}
	if mat.Thickness > 0 {
		s.Thickness = mat.Thickness
	}
	if mat.Area > 0 {
		s.Area = mat.Area
	}
	if mat.CurieTemp > 0 {
		s.CurieTemp = mat.CurieTemp
	}
	if mat.CurieConst > 0 {
		s.CurieConst = mat.CurieConst
	}
	if mat.SeriesResistanceOhm > 0 {
		s.SeriesResistance = mat.SeriesResistanceOhm
	}
	if mat.Tau0NLS > 0 {
		s.TauInf = mat.Tau0NLS
	}
	if mat.EaNLS > 0 {
		s.ActivationField = mat.EaNLS
	}

	// Initialize P to negative remanent polarization if provided
	if mat.Pr != 0 {
		s.P = -math.Abs(mat.Pr)
	}
	// Configure saturation clamp using material Ps/Pr when available.
	pMax := math.Max(math.Abs(mat.Ps), math.Abs(mat.Pr))
	if pMax > 0 {
		s.PMax = pMax
	}
	if s.PMax <= 0 {
		s.PMax = math.Abs(s.P)
	}

	// Debug logging to confirm configuration
	matLog.Input("ConfigureFromMaterial", map[string]interface{}{
		"Beta":             s.Beta,
		"Gamma":            s.Gamma,
		"Rho":              s.Rho,
		"Q12":              s.Q12,
		"Stress_Pa":        s.Stress,
		"K_dep":            s.K_dep,
		"Thickness":        s.Thickness,
		"Area":             s.Area,
		"CurieTemp":        s.CurieTemp,
		"CurieConst":       s.CurieConst,
		"SeriesResistance": s.SeriesResistance,
		"ActivationField":  s.ActivationField,
		"TauInf":           s.TauInf,
		"InitPolarization": s.P,
	})
}

// dPdT is the Master Differential Equation with Depolarization:
// rho * dP/dt = E_eff - dG/dP
// E_eff = E_applied - E_depolarization
// E_depolarization = K_dep * P
// dG/dP = 2*alpha*P + 4*beta*P^3 + 6*gamma*P^5
func (s *LKSolver) dPdT(t, P, E_applied, noise, rhoEff float64) float64 {
	// Depolarization Field: creates "slant" for 30-level analog operation
	// This term opposes polarization, braking the switching to enable intermediate states
	E_depolarization := s.K_dep * P

	// Effective Field
	E_eff := E_applied - E_depolarization

	// Deterministic Force (Gradient of Free Energy)
	P2 := P * P
	P3 := P2 * P
	P5 := P3 * P2
	dG_dP := (2 * s.Alpha * P) + (4 * s.Beta * P3) + (6 * s.Gamma * P5)

	return (E_eff + noise - dG_dP) / rhoEff
}

// Step performs one Runge-Kutta 4 (RK4) integration step.
// Returns the new Polarization P.
func (s *LKSolver) Step(E, dt float64) float64 {
	s.UpdateParams() // Ensure Alpha is current

	rhoEff := s.effectiveRho()
	noise := s.noiseTerm(dt, rhoEff)
	if invalidFloat(s.P) {
		s.logNumericalIssue("state", E, dt, rhoEff, noise, s.P)
		s.P = 0
	}
	if s.PMax > 0 {
		s.P = s.clampP(s.P)
	}

	// Nucleation-Limited Switching (NLS) Check
	if s.UseNLS {
		if !s.checkIncubation(E, dt) {
			// If still incubating, P does not change significantly (only dielectric response)
			// For simplicity, we define dP/dt = E / Rho (linear response) or just 0
			// A better approx is P = P_prev + epsilon * E
			return s.P
		}
	}

	// RK4 Integration
	prevP := s.P
	k1 := s.dPdT(0, s.P, E, noise, rhoEff)
	k2 := s.dPdT(dt/2, s.P+0.5*dt*k1, E, noise, rhoEff)
	k3 := s.dPdT(dt/2, s.P+0.5*dt*k2, E, noise, rhoEff)
	k4 := s.dPdT(dt, s.P+dt*k3, E, noise, rhoEff)

	if invalidFloat(k1) || invalidFloat(k2) || invalidFloat(k3) || invalidFloat(k4) {
		s.logNumericalIssue("k", E, dt, rhoEff, noise, prevP)
		s.Time += dt
		return s.P
	}

	dP := (dt / 6.0) * (k1 + 2*k2 + 2*k3 + k4)
	if invalidFloat(dP) {
		s.logNumericalIssue("dP", E, dt, rhoEff, noise, prevP)
		s.Time += dt
		return s.P
	}
	nextP := prevP + dP
	if invalidFloat(nextP) {
		s.logNumericalIssue("P", E, dt, rhoEff, noise, prevP)
		s.Time += dt
		return s.P
	}
	nextP = s.clampP(nextP)

	s.P = nextP
	s.Time += dt

	s.logStep(E, dt, rhoEff, noise, k1)

	return s.P
}

// checkIncubation determines if the domain has finished incubating based on Merz's Law.
// Returns true if switching can proceed.
func (s *LKSolver) checkIncubation(E, dt float64) bool {
	// Simplified NLS:
	// Calculate Incubation Time t_inc = tau_inf * exp(Ea / E)
	// If cumulative time under field E > t_inc, then switch.

	E_mag := math.Abs(E)
	const MinField = 1.0e6 // 0.01 MV/cm threshold

	if E_mag < MinField {
		return false // Field too small to drive nucleation
	}

	// Merz's Law for Incubation Time
	// E should be in V/m. ActivationField is in V/m.
	activationField := s.ActivationField
	if activationField <= 0 {
		activationField = 1.9e9
	}

	tNum := s.TauInf * math.Exp(activationField/E_mag)

	// Add to incubation time tracker (simplified for now)
	// Real NLS requires tracking state for each domain.
	// For a single solver instance (mean field), we can just use a delay counter.

	// For this Phase 1 implementation, we will perform a probabilistic check
	// Probability of nucleation in dt: P_nuc = 1 - exp(-dt / t_inc)

	prob := 1.0 - math.Exp(-dt/tNum)
	return rand.Float64() < prob
}

func (s *LKSolver) effectiveRho() float64 {
	rhoEff := s.Rho
	if s.UseEffectiveViscosity && s.SeriesResistance > 0 && s.Thickness > 0 && s.Area > 0 {
		rhoEff += (s.SeriesResistance * s.Area / s.Thickness)
	}
	return rhoEff
}

func (s *LKSolver) noiseTerm(dt, rhoEff float64) float64 {
	if !s.EnableNoise || dt <= 0 {
		return 0
	}

	const kB = 1.380649e-23 // J/K
	sigma := math.Sqrt(2 * kB * s.Temperature * rhoEff / dt)
	return rand.NormFloat64() * sigma
}

func (s *LKSolver) logStep(E, dt, rhoEff, noise, dPdt float64) {
	if !logging.IsVerbose(logging.VerbosityTrace) {
		return
	}
	if lkLog == nil {
		lkLog = logging.NewLogger("lk-solver")
	}
	if lkLog == nil {
		return
	}
	if s.logLimit > 0 && s.logCount >= s.logLimit {
		return
	}
	s.logCount++

	E_dep := s.K_dep * s.P
	E_eff := E - E_dep
	dG_dP := (2 * s.Alpha * s.P) + (4 * s.Beta * math.Pow(s.P, 3)) + (6 * s.Gamma * math.Pow(s.P, 5))

	lkLog.Calculation("LKStep", map[string]interface{}{
		"E_applied":   E,
		"E_dep":       E_dep,
		"E_eff":       E_eff,
		"Alpha":       s.Alpha,
		"Beta":        s.Beta,
		"Gamma":       s.Gamma,
		"K_dep":       s.K_dep,
		"P":           s.P,
		"dG_dP":       dG_dP,
		"rho_eff":     rhoEff,
		"noise":       noise,
		"dt":          dt,
		"Temperature": s.Temperature,
		"Stress_Pa":   s.Stress,
	}, dPdt)
}

func invalidFloat(v float64) bool {
	return math.IsNaN(v) || math.IsInf(v, 0)
}

func (s *LKSolver) clampP(P float64) float64 {
	if s.PMax <= 0 {
		return P
	}
	limit := s.PMax * 1.2
	if limit <= 0 {
		return P
	}
	if P > limit {
		return limit
	}
	if P < -limit {
		return -limit
	}
	return P
}

func (s *LKSolver) logNumericalIssue(stage string, E, dt, rhoEff, noise, prevP float64) {
	if !logging.IsVerbose(logging.VerbosityDebug) {
		return
	}
	if s.nanLimit > 0 && s.nanCount >= s.nanLimit {
		return
	}
	s.nanCount++
	if lkLog == nil {
		lkLog = logging.NewLogger("lk-solver")
	}
	if lkLog == nil {
		return
	}
	lkLog.Debug("LK numerical issue (%s): E=%.3e dt=%.3e P=%.3e rho=%.3e noise=%.3e alpha=%.3e beta=%.3e gamma=%.3e",
		stage, E, dt, prevP, rhoEff, noise, s.Alpha, s.Beta, s.Gamma)
}

func (s *LKSolver) SetState(P float64) {
	if invalidFloat(P) {
		return
	}
	s.P = s.clampP(P)
}

func (s *LKSolver) GetState() float64 {
	return s.P
}
