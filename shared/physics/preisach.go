package physics

import ()

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

// PreisachStack stores the Preisach turning-point history used to evaluate
// hysteretic polarization P(E) in ferroelectric media.
//
// Physics context:
//   - E is electric field in V/m.
//   - P returned by Update/ComputePolarization is polarization in C/m^2.
//   - The stack keeps only reversal extrema and enforces the classical
//     wipe-out property, which removes eclipsed minor loops while preserving
//     major-loop memory.
type PreisachStack struct {
	Stack       []TurningPoint
	CurrentDir  int // +1 (Increasing E), -1 (Decreasing E)
	LastE       float64
	SaturationE float64 // Field required to reach saturation
	Everett     EverettFunction
}

// NewPreisachStack constructs a Preisach history initialized at negative
// saturation (-SaturationE). saturationE is the field magnitude in V/m needed
// to reach saturated polarization, and everett provides the Everett integral
// over Preisach plane regions.
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

// Update ingests a new applied field E (V/m), updates reversal history, applies
// wipe-out compression, and returns the resulting polarization P (C/m^2).
//
// This is the main dynamic entry point for rate-independent Preisach hysteresis
// evolution when the excitation waveform is provided sample-by-sample.
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

// ComputePolarization evaluates polarization P (C/m^2) at currentE (V/m) from
// the current turning-point geometry and Everett function.
//
// The implementation follows the alternating-segment Preisach construction and
// is allocation-free to support high-rate simulation loops.
func (ps *PreisachStack) ComputePolarization(currentE float64) float64 {
	// P = -Ps + 2 * Sum
	// Sum = E(M1, m0) - E(M1, m1) + E(M2, m1) - E(M2, m2) + ...
	//
	// Use compensated accumulation (Kahan summation) to reduce loss of
	// significance when long turning-point histories produce many alternating
	// positive/negative terms.
	sum := 0.0
	compensation := 0.0
	addCompensated := func(v float64) {
		y := v - compensation
		t := sum + y
		compensation = (t - sum) - y
		sum = t
	}

	n := len(ps.Stack)

	// Initial branch: stack only has m0, so currentE acts as first max segment.
	if n == 1 {
		addCompensated(ps.Everett.Calculate(currentE, ps.Stack[0].E))
		return -ps.Everett.Calculate(ps.SaturationE, -ps.SaturationE) + 2.0*sum
	}

	// Stack points are [m0, M1, m1, M2, m2, ...]
	for i := 1; i < n; i += 2 {
		maxVal := ps.Stack[i].E
		minPrev := ps.Stack[i-1].E
		addCompensated(ps.Everett.Calculate(maxVal, minPrev))

		// Next min in stack, or currentE if this is the last max segment.
		if i+1 < n {
			addCompensated(-ps.Everett.Calculate(maxVal, ps.Stack[i+1].E))
		} else {
			addCompensated(-ps.Everett.Calculate(maxVal, currentE))
		}
	}

	// When ascending (stack ends with a Min, n is odd), the loop above only
	// covers segments between stored Max/Min pairs. It misses the open
	// ascending segment from the last Min to currentE. Without this term,
	// P is frozen during ascending ramps and jumps discontinuously at
	// wipe-out events (the "Pr teleportation" bug).
	//
	// The n==1 case (initial ascending from m0) is handled by the early
	// return above, so this only fires for n >= 3.
	if n%2 == 1 {
		addCompensated(ps.Everett.Calculate(currentE, ps.Stack[n-1].E))
	}

	return -ps.Everett.Calculate(ps.SaturationE, -ps.SaturationE) + 2.0*sum
}
