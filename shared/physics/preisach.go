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
		// The stack ends with a Min (where we turned to start ascending)
		// So the previous Max is at len-2.
		for len(ps.Stack) >= 2 {
			maxPoint := ps.Stack[len(ps.Stack)-2]
			
			// We only care if maxPoint is a MAX that we are exceeding
			if maxPoint.Type == 1 && E >= maxPoint.E {
				// Wipe out this Max/Min pair (pop the Max and the Min BEFORE it? No, pop Max and Min After it?)
				// Stack: ... Min_prev, Max, Min_last
				// We fuse Min_prev and Min_last?
				// Actually, standard wipeout removes the nested loop (Max, Min_last).
				// We pop the top two elements (Min_last and Max).
				ps.Stack = ps.Stack[:len(ps.Stack)-2] 
				
				// Now the top of stack is Min_prev. We continue ascending from there.
			} else {
				break
			}
		}
	} else { // Descending
		// If E < previous Min on stack, pop that Min (and its paired Max)
		// Stack ends with Max. Previous Min is at len-2.
		for len(ps.Stack) >= 2 {
			minPoint := ps.Stack[len(ps.Stack)-2]
			
			if minPoint.Type == -1 && E <= minPoint.E {
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
	// P = -Ps + 2 * Sum
	// Sum = E(M1, m0) - E(M1, m1) + E(M2, m1) - E(M2, m2) + ...
	
	// Create a temporary list of points including current E
	points := make([]float64, 0, len(ps.Stack)+1)
	for _, tp := range ps.Stack {
		points = append(points, tp.E)
	}
	points = append(points, currentE)
	
	// points[0] is always Min0 (-Sat)
	// points[1] is Max1
	// points[2] is Min1
	// ...
	
	sum := 0.0
	
	// Iterate over Max points (indices 1, 3, 5...)
	for i := 1; i < len(points); i += 2 {
		Max := points[i]
		MinPrev := points[i-1] // always exists
		
		// Add the UP switching triangle
		sum += ps.Everett.Calculate(Max, MinPrev)
		
		// Check for DOWN switching triangle (subtraction)
		if i+1 < len(points) {
			MinNext := points[i+1]
			sum -= ps.Everett.Calculate(Max, MinNext)
		}
	}
	
	return -ps.Everett.Calculate(ps.SaturationE, -ps.SaturationE) + 2.0*sum
}
