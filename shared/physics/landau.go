package physics

import (
	"math"
	"math/rand"
)

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
	
	// Depolarization (Polycrystalline Analog Behavior)
	K_dep float64 // Depolarization coefficient (V*m/C) - creates "slant" for analog levels

	// NLS Parameters
	UseNLS        bool
	ActivationE   float64 // Activation Energy (eV) for Merz's Law
	TauInf        float64 // Infinite field switching time (s)
	IncubationEnd float64 // Time when switching can start (s)

	// Dynamic State
	Alpha       float64 // Calculated stiffness (Temperature + Stress dependent)
	P           float64 // Current Polarization (C/m^2)
	Temperature float64 // Current Temperature (K)
	Time        float64 // Simulation time
}

// NewLKSolver creates a new solver with default "Golden Set" parameters for 10nm HZO.
func NewLKSolver() *LKSolver {
	return &LKSolver{
		// Default to "Golden Set" (Set I)
		Beta:  -2.160e8,
		Gamma: 1.653e10,
		Rho:   0.05,
		Q12:   -0.026,
		Stress: 1.0e9, // 1 GPa
		
		// Depolarization for Polycrystalline Analog Behavior
		K_dep: 2.5e8, // V*m/C - Creates slanted loop for 30-level operation
		
		UseNLS:      true,
		ActivationE: 0.7,
		TauInf:      1.0e-13,
		
		Temperature: 300.0,
	}
}

// UpdateParams recalculates Alpha based on current Temperature and Stress using
// the Unified Coefficient Formula: alpha = alpha_t(T) - 2*Q12*Stress
func (s *LKSolver) UpdateParams() {
	const (
		Tc         = 723.0     // Curie Temperature (K)
		CurieConst = 1.5e5     // Curie Constant (K)
		Eps0       = 8.854e-12 // Vacuum Permittivity (F/m)
	)

	// Thermodynamic contribution (Curie-Weiss)
	alphaT := (s.Temperature - Tc) / (2 * Eps0 * CurieConst)

	// Mechanical contribution (Electrostriction)
	// Tensile stress (positive) with negative Q12 makes Alpha more negative (stable)
	alphaMech := 2 * s.Q12 * s.Stress

	s.Alpha = alphaT - alphaMech
}

// dPdT is the Master Differential Equation:
// rho * dP/dt = E_total - dG/dP
// where E_total = E_applied - E_dep (E_dep = K_dep * P)
// dG/dP = 2*alpha*P + 4*beta*P^3 + 6*gamma*P^5
//
// The depolarization field (E_dep) represents the collective effect of grain boundaries
// and domain interactions in a polycrystalline device. It opposes polarization growth,
// creating a "slanted" hysteresis loop that enables analog (multi-level) operation.
func (s *LKSolver) dPdT(t, P, E float64) float64 {
	// 1. Calculate Depolarization Field
	// This represents the average effect of grain boundaries/interfacial layers
	// Higher K_dep → More slanted loop → More analog levels
	E_dep := s.K_dep * P
	
	// 2. Effective Field (Applied minus Depolarization)
	E_total := E - E_dep
	
	// 3. Deterministic Force (Gradient of Free Energy)
	dG_dP := (2 * s.Alpha * P) + (4 * s.Beta * math.Pow(P, 3)) + (6 * s.Gamma * math.Pow(P, 5))

	// 4. Stochastic Noise (Langevin) - Optional, scaled by Rho
	noise := 0.0
	// TODO: Add EnableNoise flag and Box-Muller generator

	// Use E_total (with depolarization) instead of raw E_applied
	return (E_total + noise - dG_dP) / s.Rho
}

// Step performs one Runge-Kutta 4 (RK4) integration step.
// Returns the new Polarization P.
func (s *LKSolver) Step(E, dt float64) float64 {
	s.UpdateParams() // Ensure Alpha is current

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
	k1 := s.dPdT(0, s.P, E)
	k2 := s.dPdT(dt/2, s.P+0.5*dt*k1, E)
	k3 := s.dPdT(dt/2, s.P+0.5*dt*k2, E)
	k4 := s.dPdT(dt, s.P+dt*k3, E)

	dP := (dt / 6.0) * (k1 + 2*k2 + 2*k3 + k4)
	s.P += dP
	s.Time += dt
	
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
	// Note: We use the absolute field magnitude relative to activation field
	// E should be in V/m. ActivationE is in eV, need to convert or use calibrated ActivationField (V/m)
	// Let's assume ActivationE is actually ActivationField in V/m for this calculation to keep units consistent
	// Typically Ea ~ 19 MV/cm = 1.9e9 V/m
	
	const ActivationField = 1.9e9 
	
	tNum := s.TauInf * math.Exp(ActivationField/E_mag)
	
	// Add to incubation time tracker (simplified for now)
	// Real NLS requires tracking state for each domain. 
	// For a single solver instance (mean field), we can just use a delay counter.
	
	// For this Phase 1 implementation, we will perform a probabilistic check
	// Probability of nucleation in dt: P_nuc = 1 - exp(-dt / t_inc)
	
	prob := 1.0 - math.Exp(-dt/tNum)
	return rand.Float64() < prob
}

func (s *LKSolver) SetState(P float64) {
	s.P = P
}

func (s *LKSolver) GetState() float64 {
	return s.P
}
