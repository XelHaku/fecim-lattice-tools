package physics

import (

)

// EverettFunction abstract interface for the density distribution
type EverettFunction interface {
	// Calculate returns the integral of density distribution over the region defined by alpha, beta
	Calculate(alpha, beta float64) float64
}

// TurningPoint represents a reversal in the field history
type TurningPoint struct {
	E    float64 // Electric Field Value
	Type int     // +1 for Max (Ascending->Descending), -1 for Min (Descending->Ascending)
}

// PreisachStack implements the memory structure for the Preisach Model
// utilizing the "Wipe-Out" property to perfectly compress history.
type PreisachStack struct {
	Stack       []TurningPoint
	CurrentDir  int // +1 (Increasing E), -1 (Decreasing E)
	LastE       float64
	SaturationE float64 // Field required to reach saturation
	Everett     EverettFunction
}

// NewPreisachStack creates a new history stack initialized at negative saturation
func NewPreisachStack(saturationE float64, everett EverettFunction) *PreisachStack {
	// Initial state: Deep negative saturation
	// Stack has one point: {-Sat, Min}
	ps := &PreisachStack{
		Stack:       make([]TurningPoint, 0),
		CurrentDir:  1, // Next move will be increasing from negative sat
		LastE:       -saturationE,
		SaturationE: saturationE,
		Everett:     everett,
	}
	
	// Seed with negative saturation
	ps.Stack = append(ps.Stack, TurningPoint{E: -saturationE, Type: -1})
	
	return ps
}

// Update processes a new input field value E.
// It applies the Wipe-Out logic and updates the stack.
// Returns the new Polarization P.
func (ps *PreisachStack) Update(E float64) float64 {
	// 1. Determine Direction
	direction := 0
	if E > ps.LastE {
		direction = 1
	} else if E < ps.LastE {
		direction = -1
	} else {
		return ps.ComputePolarization(E) // No change
	}

	// 2. Check for Reversal (Creation of new turning point)
	if direction != ps.CurrentDir {
		// We just turned! Push the *previous* point onto the stack
		// If we were increasing (Dir=1), LastE is a local Max
		// If we were decreasing (Dir=-1), LastE is a local Min
		
		tpType := 0
		if ps.CurrentDir == 1 {
			tpType = 1 // Max
		} else {
			tpType = -1 // Min
		}
		
		ps.Stack = append(ps.Stack, TurningPoint{E: ps.LastE, Type: tpType})
		ps.CurrentDir = direction
	}

	// 3. Wipe-Out Logic
	// Erase any historical turning points that are "engulfed" by the new excursion
	
	if direction == 1 { // Ascending
		// If E > previous Max on stack, pop that Max (and its paired Min)
		for len(ps.Stack) >= 2 {
			top := ps.Stack[len(ps.Stack)-1]
			
			// We only care if top is a MAX that we are exceeding
			if top.Type == 1 && E >= top.E {
				// Wipe out this Max/Min pair
				ps.Stack = ps.Stack[:len(ps.Stack)-2] 
				// Note: In standard Preisach, stack is alternating Min/Max.
				// If we pop Max, we also pop the Min below it to merge the major loop.
			} else {
				break
			}
		}
	} else { // Descending
		// If E < previous Min on stack, pop that Min (and its paired Max)
		for len(ps.Stack) >= 2 {
			top := ps.Stack[len(ps.Stack)-1]
			
			if top.Type == -1 && E <= top.E {
				ps.Stack = ps.Stack[:len(ps.Stack)-2]
			} else {
				break
			}
		}
	}
	
	ps.LastE = E
	return ps.ComputePolarization(E)
}

// ComputePolarization sums the hysterons based on the geometric shape of the stack
func (ps *PreisachStack) ComputePolarization(currentE float64) float64 {
	// Start with negative saturation
	// P = -Ps + 2 * Sum( Area(Triangle_k) )
	
	// This simplified "Sum of Areas" relies on the Everett function properties
	// P(t) = -Ps + 2 * [ E_func(alpha1, beta0) - E_func(alpha1, beta1) + E_func(alpha2, beta1) ... ]
	
	// Construct temporary full list of corners for calculation
	// The "Active" corner is the current E
	
	// Create a temporary list of points including current E
	points := make([]float64, 0, len(ps.Stack)+1)
	for _, tp := range ps.Stack {
		points = append(points, tp.E)
	}
	points = append(points, currentE)
	
	// Summation
	// P = -Ps + 2 * Sum[ E(M_i, m_i-1) - E(M_i, m_i) ]
	// Requires standardizing the sequence.
	// We assume stack starts with Min (-Sat).
	
	// Placeholder for Ps (Saturation Polarization)
	// Typically Everett(Sat, -Sat) = Ps
	// So -Ps is the baseline.
	
	// For standard HZO, we can assume normalized output [-1, 1] and scale later,
	// or assume Everett returns actual Polarization units.
	
	sum := 0.0
	
	// Iterate through pairs (Min, Max)
	// Points: [Min0, Max1, Min2, Max3, ... Current]
	
	for i := 1; i < len(points); i += 2 {
		val_max := points[i]
		val_min := points[i-1]
		
		// If this is the last segment and it's incomplete (currentE is a Min going down)
		// We might need special handling. But generally the sequence defines the boundary.
		
		// Everett Formula: P = sum ( E(alpha_k, beta_k-1) - E(alpha_k, beta_k) )
		// Here alpha is 'up switching' (vertical axis), beta is 'down switching' (horizontal)
		// Triangle T(alpha, beta) is the area of hysterons switched UP.
		
		// Let's use the property:
		// Contribution = Everett(Max, Min) - Everett(Max, Max) ??
		// Ideally Everett(x, y) gives integral over triangle defined by tip (x,y)
		
		sum += ps.Everett.Calculate(val_max, val_min)
	}
	
	// Scale: Base is -Ps. Double the sum because we are switching from -1 to +1?
	// Depends on Everett definition.
	// Usually P = -Ps + 2 * Sum
	
	// Let's rely on the interface to handle the magnitude. 
	// If Calculate returns P_switched_volume, then:
	return -ps.Everett.Calculate(ps.SaturationE, -ps.SaturationE) + 2.0*sum
}
