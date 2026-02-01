// Package ferroelectric provides physics models for ferroelectric materials.
package ferroelectric

import (
	"math"

	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/physics"
)

// Package-level logger
var log *logging.Logger

func init() {
	log = logging.NewLogger("preisach")
}

// TanhEverett implements the EverettFunction interface using Bo Jiang's Tanh model.
// This preserves the existing "S-shape" characteristic of the simulation.
type TanhEverett struct {
	Ps     float64
	Ec     float64
	Delta  float64 // Distribution width
}

// Calculate returns the polarization contribution for a hysteron switching region.
// Note: The shared/physics.PreisachStack uses this to sum triangular areas.
// For the Tanh model, we treat the major loop function P_major(E) as the
// integral of the Everett function along the line beta = -E_sat.
// Simplification: We return P_major(alpha) - P_major(beta) ??
//
// Actually, the shared stack implementation computes:
// P = -Ps + 2 * Sum( Everett(Max, Min) )
// If we want to match the analytic Tanh model:
// P(E) = Ps * Tanh((E - Ec)/Delta)
//
// We need to map the "Everett(Max, Min)" call to something that reconstructs this.
// A common approximation for independent hysterons (like Tanh) is:
// Everett(alpha, beta) = (F(alpha) - F(beta)) / 2 * Ps
// where F is the cumulative distribution function (the major loop normalized to 0..1).
//
// Let's implement this factorization.
func (t *TanhEverett) Calculate(alpha, beta float64) float64 {
	// F(x) = Tanh((x - Ec) / Delta) ?
	// No, Tanh is symmetric around Ec (if we account for bias).
	// Let's assume standard centered loop for the hysteron distribution Basis.
	
	// Major ascending branch: P_asc(E)
	// Major descending branch: P_desc(E)
	
	// Factorized Everett function formulation:
	// E(alpha, beta) ~ (P_asc(alpha) - P_desc(beta)) / 2
	
	valAlpha := math.Tanh((alpha - t.Ec) / t.Delta)
	valBeta  := math.Tanh((beta + t.Ec) / t.Delta) // Note offset for descending
	
	// We return the "volume" of dipoles in the triangle (alpha, beta)
	// Scaling factor Ps is handled outside or here? 
	// The stack sums 2 * Everett. 
	// P = -Ps + 2 * Sum.
	// If Sum = (valAlpha - valBeta)/2 * Ps
	// Then 2*Sum = (valAlpha - valBeta) * Ps
	
	return (valAlpha - valBeta) * 0.5 * t.Ps
}

// PreisachModel implements the Preisach hysteresis model for ferroelectrics.
// Wraps the shared/physics.PreisachStack engine.
// PreisachModel implements the Preisach hysteresis model for ferroelectrics.
// Wraps the shared/physics.PreisachStack engine.
type PreisachModel struct {
	material *HZOMaterial
	stack    *physics.PreisachStack
	everett  *TanhEverett
	
	Temperature float64
	Stress      float64 // GPa
}

// NewPreisachModel creates a new Preisach model with the given material.
func NewPreisachModel(material *HZOMaterial) *PreisachModel {
	log.Input("NewPreisachModel", map[string]interface{}{
		"material_name": material.Name,
		"Ec":            material.Ec,
		"Ps":            material.Ps,
	})

	// Configure Everett function based on material
	everett := &TanhEverett{
		Ps:    material.Ps,
		Ec:    material.Ec,
		Delta: material.Ec * 0.25, // 25% distribution width
	}
	
	// E_saturation should be > Ec. typically 3-5x Ec.
	E_sat := material.Ec * 5.0

	return &PreisachModel{
		material:    material,
		stack:       physics.NewPreisachStack(E_sat, everett),
		everett:     everett,
		Temperature: 300.0,
		Stress:      1.0, // Default 1 GPa
	}
}

// DiscreteState represents a single programmable state.
type DiscreteState struct {
	Level       int
	Polarization float64
	NormalizedP float64
	Voltage     float64
	Conductance float64
}

// DiscreteStates returns the polarization values for n evenly spaced discrete states.
// This is a helper for testing and visualization.
func (p *PreisachModel) DiscreteStates(n int) []float64 {
	poles := make([]float64, n)
	step := 2.0 * p.material.Ps / float64(n-1)
	for i := 0; i < n; i++ {
		poles[i] = -p.material.Ps + float64(i)*step
	}
	return poles
}

// Reset clears the history and sets polarization to negative saturation.
func (p *PreisachModel) Reset() {
	// Re-initialize stack
	E_sat := p.material.Ec * 5.0
	everett := p.stack.Everett
	p.stack = physics.NewPreisachStack(E_sat, everett)
}

// Update applies a new electric field and returns the resulting polarization.
func (p *PreisachModel) Update(E float64) float64 {
	log.Input("Update", map[string]interface{}{"E": E})
	
	P := p.stack.Update(E)
	
	log.Calculation("Update", map[string]interface{}{"E": E}, P)
	return P
}

// Polarization returns the current polarization state.
func (p *PreisachModel) Polarization() float64 {
	// We need to expose P from stack or cache it.
	// Implementing a getter on stack would be cleaner, but Update returns it.
	// For now, assume state is consistent or store last P if needed.
	// But Update(LastE) is idempotent in theory (no change).
	return p.stack.Update(p.stack.LastE) 
}

// NormalizedPolarization returns polarization as fraction of Ps (-1 to +1).
func (p *PreisachModel) NormalizedPolarization() float64 {
	return p.Polarization() / p.material.Ps
}

// GetHysteresisLoop generates a full P-E hysteresis loop.
func (p *PreisachModel) GetHysteresisLoop(Emax float64, points int) ([]float64, []float64) {
	// Temporarily reset stack to generate loop
	// Ideally we clone, but for GUI generation we typically want a fresh loop
	p.Reset()

	E := make([]float64, 0, points*4)
	PVal := make([]float64, 0, points*4) // renamed to avoid collision with P() method

	// Saturation start
	p.Update(-Emax)

	// Ascending
	for i := 0; i <= points*2; i++ {
		e := -Emax + 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		PVal = append(PVal, pol)
	}

	// Descending
	for i := 1; i <= points*2; i++ {
		e := Emax - 2*Emax*float64(i)/float64(points*2)
		pol := p.Update(e)
		E = append(E, e)
		PVal = append(PVal, pol)
	}

	return E, PVal
}

// SetTemperature updates the simulation temperature and scales material parameters.
func (p *PreisachModel) SetTemperature(tempK float64) {
	p.Temperature = tempK
	
	// Scale Ec and Ps based on linear temperature coefficients
	// Ec(T) = Ec0 + Coeff * (T - 300)
	// Note: We use 300K as reference, assuming params are defined at RT
	
	deltaT := tempK - 300.0
	
	// Retrieve coefficients from material config (HZOMaterial needs these fields exposed)
	// Assuming HZOMaterial has TempCoeffEc and TempCoeffPr
	
	newEc := p.material.Ec + p.material.TempCoeffEc * deltaT
	newPs := p.material.Ps + p.material.TempCoeffPr * deltaT
	
	// Safety clamps
	if newEc < 1e5 { newEc = 1e5 } // Minimum 0.001 MV/cm
	if newPs < 1e-6 { newPs = 1e-6 } // Minimum polarization
	
	// Update Everett function
	p.everett.Ec = newEc
	p.everett.Ps = newPs
	p.everett.Delta = newEc * 0.25
}

// GetEffectiveEc returns the current temperature-scaled Coercive Field.
func (p *PreisachModel) GetEffectiveEc() float64 {
	return p.everett.Ec
}

// SetStress updates the mechanical stress and scales Ec accordingly.
// Stress is in GPa.
// Scaling Logic: Ec ~ sqrt(|Alpha|)
// Alpha = AlphaT - 2*Q12*Stress
func (p *PreisachModel) SetStress(stressGPa float64) {
	p.Stress = stressGPa
	
	// Recalculate everything (Temperature and Stress)
	p.updateEffectiveParameters()
}

// updateEffectiveParameters recalculates Ec and Ps based on T and Stress
func (p *PreisachModel) updateEffectiveParameters() {
	// Base parameters at 300K, 1GPa (if calibrated there) or 0GPa?
	// Let's assume material.Ec is at 300K and defined Stress (usually 1GPa for HZO).
	// For simplicity, we use linear temp scaling + alpha-based stress scaling relative to baseline.
	
	// 1. Temperature Scaling
	deltaT := p.Temperature - 300.0
	ec_T := p.material.Ec + p.material.TempCoeffEc * deltaT
	ps_T := p.material.Ps + p.material.TempCoeffPr * deltaT
	
	// 2. Stress Scaling
	// Relative change in Alpha
	// Alpha_ref = Alpha(300K, 1GPa)
	// Alpha_new = Alpha(T, Stress)
	
	// Need material Landau coefficients. If not available, assume default.
	// HZO defaults:
	// Q12 ~ -0.026
	// Alpha_0 ~ -1e9 (at 300K)
	
	// If material struct doesn't have Q12, we can't do accurate stress scaling.
	// We will skip stress scaling if Q12 is 0 or missing, to avoid breaking logic.
	
	// Assuming linear stress effect on Ec for now if we lack full Landau params:
	// Ec increases with tensile stress (negative Q12 makes alpha more negative).
	// Sensitivity: dEc/dStress approx 0.1 MV/cm per GPa?
	
	// Let's implement a simplified linear sensitivity based on literature if coefficients missing
	// dEc/dSigma ~ +5% per GPa for HZO
	
	stressFactor := 1.0 + 0.05 * (p.Stress - 1.0) // Relative to 1GPa baseline
	
	newEc := ec_T * stressFactor
	newPs := ps_T // Ps is less sensitive to stress in first order

	// Safety
	if newEc < 1e5 { newEc = 1e5 }
	if newPs < 1e-6 { newPs = 1e-6 }
	
	// Update Everett
	p.everett.Ec = newEc
	p.everett.Ps = newPs
	p.everett.Delta = newEc * 0.25
}
