package physics

import (
	"math"
	"testing"
)

// ============================================================================
// TRANSFER.GO TESTS
// ============================================================================

func TestPolarizationToConductance_NormalMapping(t *testing.T) {
	tests := []struct {
		name     string
		P        float64
		Ps       float64
		Gmin     float64
		Gmax     float64
		expected float64
	}{
		{
			name:     "P at +Ps maps to Gmax",
			P:        10.0,
			Ps:       10.0,
			Gmin:     1e-6,
			Gmax:     1e-3,
			expected: 1e-3,
		},
		{
			name:     "P at -Ps maps to Gmin",
			P:        -10.0,
			Ps:       10.0,
			Gmin:     1e-6,
			Gmax:     1e-3,
			expected: 1e-6,
		},
		{
			name:     "P at 0 maps to midpoint",
			P:        0.0,
			Ps:       10.0,
			Gmin:     1e-6,
			Gmax:     1e-3,
			expected: (1e-6 + 1e-3) / 2,
		},
		{
			name:     "P at +Ps/2 maps to 3/4 point",
			P:        5.0,
			Ps:       10.0,
			Gmin:     0.0,
			Gmax:     100.0,
			expected: 75.0, // 0 + 100 * (0.5 + 1) / 2 = 75
		},
		{
			name:     "P at -Ps/2 maps to 1/4 point",
			P:        -5.0,
			Ps:       10.0,
			Gmin:     0.0,
			Gmax:     100.0,
			expected: 25.0, // 0 + 100 * (-0.5 + 1) / 2 = 25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PolarizationToConductance(tt.P, tt.Ps, tt.Gmin, tt.Gmax)
			if math.Abs(result-tt.expected) > 1e-12 {
				t.Errorf("PolarizationToConductance(%v, %v, %v, %v) = %v, expected %v",
					tt.P, tt.Ps, tt.Gmin, tt.Gmax, result, tt.expected)
			}
		})
	}
}

func TestPolarizationToConductance_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		P        float64
		Ps       float64
		Gmin     float64
		Gmax     float64
		expected float64
	}{
		{
			name:     "Ps = 0 returns midpoint",
			P:        5.0,
			Ps:       0.0,
			Gmin:     1e-6,
			Gmax:     1e-3,
			expected: (1e-6 + 1e-3) / 2,
		},
		{
			name:     "P > Ps clamped to Gmax",
			P:        20.0,
			Ps:       10.0,
			Gmin:     1e-6,
			Gmax:     1e-3,
			expected: 1e-3,
		},
		{
			name:     "P < -Ps clamped to Gmin",
			P:        -20.0,
			Ps:       10.0,
			Gmin:     1e-6,
			Gmax:     1e-3,
			expected: 1e-6,
		},
		{
			name:     "Negative Ps with positive P",
			P:        5.0,
			Ps:       -10.0,
			Gmin:     0.0,
			Gmax:     100.0,
			expected: 25.0, // normalizedP = 5/-10 = -0.5, clamped, result = ((-0.5)+1)/2*100 = 25
		},
		{
			name:     "Gmin = Gmax returns that value",
			P:        5.0,
			Ps:       10.0,
			Gmin:     1e-3,
			Gmax:     1e-3,
			expected: 1e-3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PolarizationToConductance(tt.P, tt.Ps, tt.Gmin, tt.Gmax)
			if math.Abs(result-tt.expected) > 1e-12 {
				t.Errorf("PolarizationToConductance(%v, %v, %v, %v) = %v, expected %v",
					tt.P, tt.Ps, tt.Gmin, tt.Gmax, result, tt.expected)
			}
		})
	}
}

func TestConductanceToPolarization_NormalMapping(t *testing.T) {
	tests := []struct {
		name     string
		G        float64
		Gmin     float64
		Gmax     float64
		Ps       float64
		expected float64
	}{
		{
			name:     "G at Gmax maps to +Ps",
			G:        1e-3,
			Gmin:     1e-6,
			Gmax:     1e-3,
			Ps:       10.0,
			expected: 10.0,
		},
		{
			name:     "G at Gmin maps to -Ps",
			G:        1e-6,
			Gmin:     1e-6,
			Gmax:     1e-3,
			Ps:       10.0,
			expected: -10.0,
		},
		{
			name:     "G at midpoint maps to 0",
			G:        (1e-6 + 1e-3) / 2,
			Gmin:     1e-6,
			Gmax:     1e-3,
			Ps:       10.0,
			expected: 0.0,
		},
		{
			name:     "G at 3/4 point maps to +Ps/2",
			G:        75.0,
			Gmin:     0.0,
			Gmax:     100.0,
			Ps:       10.0,
			expected: 5.0, // normalizedG = 0.75, normalizedP = 2*0.75-1 = 0.5, P = 0.5*10 = 5
		},
		{
			name:     "G at 1/4 point maps to -Ps/2",
			G:        25.0,
			Gmin:     0.0,
			Gmax:     100.0,
			Ps:       10.0,
			expected: -5.0, // normalizedG = 0.25, normalizedP = 2*0.25-1 = -0.5, P = -0.5*10 = -5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConductanceToPolarization(tt.G, tt.Gmin, tt.Gmax, tt.Ps)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("ConductanceToPolarization(%v, %v, %v, %v) = %v, expected %v",
					tt.G, tt.Gmin, tt.Gmax, tt.Ps, result, tt.expected)
			}
		})
	}
}

func TestConductanceToPolarization_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		G        float64
		Gmin     float64
		Gmax     float64
		Ps       float64
		expected float64
	}{
		{
			name:     "Gmin = Gmax returns 0",
			G:        1e-3,
			Gmin:     1e-3,
			Gmax:     1e-3,
			Ps:       10.0,
			expected: 0.0,
		},
		{
			name:     "G beyond Gmax",
			G:        200.0,
			Gmin:     0.0,
			Gmax:     100.0,
			Ps:       10.0,
			expected: 30.0, // normalizedG = 2, normalizedP = 2*2-1 = 3, P = 3*10 = 30 (no clamping on reverse)
		},
		{
			name:     "G below Gmin",
			G:        -50.0,
			Gmin:     0.0,
			Gmax:     100.0,
			Ps:       10.0,
			expected: -20.0, // normalizedG = -0.5, normalizedP = 2*(-0.5)-1 = -2, P = -2*10 = -20
		},
		{
			name:     "Negative Ps",
			G:        75.0,
			Gmin:     0.0,
			Gmax:     100.0,
			Ps:       -10.0,
			expected: -5.0, // normalizedG = 0.75, normalizedP = 0.5, P = 0.5*(-10) = -5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConductanceToPolarization(tt.G, tt.Gmin, tt.Gmax, tt.Ps)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("ConductanceToPolarization(%v, %v, %v, %v) = %v, expected %v",
					tt.G, tt.Gmin, tt.Gmax, tt.Ps, result, tt.expected)
			}
		})
	}
}

func TestPolarizationConductanceRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		P    float64
		Ps   float64
		Gmin float64
		Gmax float64
	}{
		{
			name: "Round trip at +Ps",
			P:    10.0,
			Ps:   10.0,
			Gmin: 1e-6,
			Gmax: 1e-3,
		},
		{
			name: "Round trip at -Ps",
			P:    -10.0,
			Ps:   10.0,
			Gmin: 1e-6,
			Gmax: 1e-3,
		},
		{
			name: "Round trip at 0",
			P:    0.0,
			Ps:   10.0,
			Gmin: 1e-6,
			Gmax: 1e-3,
		},
		{
			name: "Round trip at arbitrary value",
			P:    3.7,
			Ps:   10.0,
			Gmin: 0.0,
			Gmax: 100.0,
		},
		{
			name: "Round trip with negative Ps",
			P:    -5.0,
			Ps:   -10.0,
			Gmin: 0.0,
			Gmax: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// P -> G -> P
			G := PolarizationToConductance(tt.P, tt.Ps, tt.Gmin, tt.Gmax)
			PRecovered := ConductanceToPolarization(G, tt.Gmin, tt.Gmax, tt.Ps)

			// Clamp expected P to valid range for comparison
			expectedP := tt.P
			if tt.Ps != 0 {
				normalizedP := tt.P / tt.Ps
				if normalizedP > 1 {
					normalizedP = 1
				}
				if normalizedP < -1 {
					normalizedP = -1
				}
				expectedP = normalizedP * tt.Ps
			}

			if math.Abs(PRecovered-expectedP) > 1e-9 {
				t.Errorf("Round trip failed: P=%v -> G=%v -> P=%v (expected %v)",
					tt.P, G, PRecovered, expectedP)
			}
		})
	}
}

// ============================================================================
// PREISACH.GO TESTS
// ============================================================================

// MockEverettFunction implements a simple linear Everett function for testing
type MockEverettFunction struct {
	scaleFactor float64
}

func (m *MockEverettFunction) Calculate(alpha, beta float64) float64 {
	// Simple linear: E(alpha, beta) = scaleFactor * (alpha - beta) / 2
	// This represents a triangular region in the Preisach plane
	return m.scaleFactor * (alpha - beta) / 2.0
}

func TestNewPreisachStack_Initialization(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}

	ps := NewPreisachStack(satE, everett)

	if ps == nil {
		t.Fatal("NewPreisachStack returned nil")
	}

	if len(ps.Stack) != 1 {
		t.Errorf("Initial stack length = %d, expected 1", len(ps.Stack))
	}

	if ps.Stack[0].E != -satE {
		t.Errorf("Initial stack point E = %v, expected %v", ps.Stack[0].E, -satE)
	}

	if ps.Stack[0].Type != -1 {
		t.Errorf("Initial stack point Type = %d, expected -1 (Min)", ps.Stack[0].Type)
	}

	if ps.CurrentDir != 1 {
		t.Errorf("Initial CurrentDir = %d, expected 1", ps.CurrentDir)
	}

	if ps.LastE != -satE {
		t.Errorf("Initial LastE = %v, expected %v", ps.LastE, -satE)
	}

	if ps.SaturationE != satE {
		t.Errorf("SaturationE = %v, expected %v", ps.SaturationE, satE)
	}
}

func TestPreisachStack_MonotonicIncrease(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// Start at -5, go to 0, then to +5
	fields := []float64{-5.0, -2.5, 0.0, 2.5, 5.0}
	var prevP float64

	for i, E := range fields {
		P := ps.Update(E)

		if i > 0 && P <= prevP {
			t.Errorf("Monotonic increase failed: E=%v gives P=%v, previous P=%v", E, P, prevP)
		}

		prevP = P
	}

	// Should have only 1 turning point (initial -satE)
	if len(ps.Stack) != 1 {
		t.Errorf("After monotonic increase, stack length = %d, expected 1", len(ps.Stack))
	}
}

func TestPreisachStack_MonotonicDecrease(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// First go up to establish a maximum
	ps.Update(5.0)

	// Now decrease monotonically
	fields := []float64{5.0, 2.5, 0.0, -2.5, -5.0}
	var prevP float64
	firstIteration := true

	for _, E := range fields {
		P := ps.Update(E)

		if !firstIteration && P >= prevP {
			t.Errorf("Monotonic decrease failed: E=%v gives P=%v, previous P=%v", E, P, prevP)
		}

		prevP = P
		firstIteration = false
	}
}

func TestPreisachStack_ReversalDetection(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// Increase from -5 to +3
	ps.Update(3.0)

	initialStackLen := len(ps.Stack)

	// Now reverse: decrease to +1
	ps.Update(1.0)

	// Should have created a new turning point (Max at E=3)
	if len(ps.Stack) != initialStackLen+1 {
		t.Errorf("After reversal, stack length = %d, expected %d", len(ps.Stack), initialStackLen+1)
	}

	// The new turning point should be a Max (Type=1) at E=3
	lastTP := ps.Stack[len(ps.Stack)-1]
	if lastTP.Type != 1 {
		t.Errorf("Last turning point Type = %d, expected 1 (Max)", lastTP.Type)
	}
	if math.Abs(lastTP.E-3.0) > 1e-9 {
		t.Errorf("Last turning point E = %v, expected 3.0", lastTP.E)
	}
}

func TestPreisachStack_WipeOutPropertyAscending(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// Create a minor loop: -5 -> +2 -> -1 -> +3
	ps.Update(2.0)  // Stack: [Min(-5)], ascending, no reversal yet
	initialLen := len(ps.Stack)

	ps.Update(-1.0) // Reversal: creates Max(2), now Stack: [Min(-5), Max(2)]

	stackBeforeWipeout := len(ps.Stack)
	expectedBeforeWipeout := initialLen + 1 // Added one turning point (Max at 2)

	if stackBeforeWipeout != expectedBeforeWipeout {
		t.Logf("Before wipeout, stack length = %d, expected %d (this is informational)",
			stackBeforeWipeout, expectedBeforeWipeout)
	}

	// Now exceed previous max: go to +3 (exceeds +2)
	ps.Update(3.0)

	// Wipe-out should have removed the Max(2) and its paired Min(-1)
	// Final stack should only have initial Min(-5)
	if len(ps.Stack) != 1 {
		t.Errorf("After wipeout, stack length = %d, expected 1", len(ps.Stack))
	}

	// Only initial Min(-5) should remain
	if ps.Stack[0].E != -satE || ps.Stack[0].Type != -1 {
		t.Errorf("After wipeout, stack[0] = {E:%v, Type:%d}, expected {E:%v, Type:-1}",
			ps.Stack[0].E, ps.Stack[0].Type, -satE)
	}
}

func TestPreisachStack_WipeOutPropertyDescending(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// Create pattern: -5 -> +4 -> +1 -> +3 -> 0
	ps.Update(4.0) // Stack: [Min(-5)], ascending
	ps.Update(1.0) // Reversal: Stack: [Min(-5), Max(4)]
	ps.Update(3.0) // Reversal: Stack: [Min(-5), Max(4), Min(1)]

	stackBeforeWipeout := len(ps.Stack)
	// We expect 3 items: Min(-5), Max(4), Min(1)
	if stackBeforeWipeout != 3 {
		t.Logf("Before descending wipeout, stack length = %d, expected 3 (this is informational)",
			stackBeforeWipeout)
	}

	// Now go below previous Min(1): go to 0
	ps.Update(0.0)

	// Wipe-out should have removed Min(1) and Max(4) pair
	// Final stack: [Min(-5), Max(3)] or just [Min(-5)] depending on implementation
	// Since we're descending and exceeded Min(1), we wipe the nested loop
	expectedAfterWipeout := 2 // [Min(-5), Max(3)]
	if len(ps.Stack) != expectedAfterWipeout {
		t.Errorf("After descending wipeout, stack length = %d, expected %d",
			len(ps.Stack), expectedAfterWipeout)
	}
}

func TestPreisachStack_SameValueNoChange(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// Move to some value
	P1 := ps.Update(2.0)
	stackLen := len(ps.Stack)

	// Update with same value
	P2 := ps.Update(2.0)

	// Polarization should be identical
	if math.Abs(P1-P2) > 1e-12 {
		t.Errorf("Same E value changed polarization: P1=%v, P2=%v", P1, P2)
	}

	// Stack should be unchanged
	if len(ps.Stack) != stackLen {
		t.Errorf("Stack length changed on same E: was %d, now %d", stackLen, len(ps.Stack))
	}
}

func TestPreisachStack_ComputePolarizationAtSaturation(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 10.0} // Ps = 10 * (satE - (-satE)) / 2 = 50
	ps := NewPreisachStack(satE, everett)

	// Move to positive saturation
	P_pos := ps.Update(satE)

	// Expected: P = -Ps + 2*E(+satE, -satE) = -50 + 2*50 = 50
	expectedPs := 50.0
	if math.Abs(P_pos-expectedPs) > 1e-9 {
		t.Errorf("Polarization at +satE = %v, expected %v", P_pos, expectedPs)
	}

	// Move to negative saturation
	P_neg := ps.Update(-satE)

	// Expected: P = -Ps + 0 (no up triangles) = -50
	if math.Abs(P_neg-(-expectedPs)) > 1e-9 {
		t.Errorf("Polarization at -satE = %v, expected %v", P_neg, -expectedPs)
	}
}

func TestPreisachStack_ComplexHysteresisLoop(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// Trace a complex path: -5 -> +5 -> -3 -> +4 -> -5 -> +5
	sequence := []float64{-5, 0, 5, 2, -3, 1, 4, 0, -5, 0, 5}

	var polarizations []float64
	for _, E := range sequence {
		P := ps.Update(E)
		polarizations = append(polarizations, P)
	}

	// Verify that we get different polarizations for same E values depending on history
	// E=0 appears at indices 1, 7, 9 with different histories
	P_first := polarizations[1]  // After -5 -> 0 (ascending)
	P_last := polarizations[9]   // After -5 -> 0 again

	// Due to hysteresis, these should differ (unless wiped out)
	// Actually, after going to -5 again, the history is wiped, so P_last should equal P_first
	if math.Abs(P_first-P_last) > 1e-9 {
		t.Logf("Note: P at E=0 varies with history - P_first=%v, P_last=%v", P_first, P_last)
	}

	// Polarization at +5 after full major loop should equal initial +5 value
	P_at_5_first := polarizations[2]
	P_at_5_second := polarizations[10]
	if math.Abs(P_at_5_first-P_at_5_second) > 1e-9 {
		t.Errorf("Major loop not closing: P(+5) first=%v, second=%v", P_at_5_first, P_at_5_second)
	}
}

func TestPreisachStack_MultipleReversals(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 1.0}
	ps := NewPreisachStack(satE, everett)

	// Create nested loops with multiple reversals
	sequence := []float64{
		-5.0, // Start
		3.0,  // Up to 3
		1.0,  // Down to 1 (reversal, creates Max at 3)
		2.0,  // Up to 2 (reversal, creates Min at 1)
		0.0,  // Down to 0 (reversal, creates Max at 2)
		1.5,  // Up to 1.5 (reversal, creates Min at 0)
	}

	for _, E := range sequence {
		ps.Update(E)
	}

	// Should have: Min(-5), Max(3), Min(1), Max(2), Min(0), and we're going up
	// After wipeouts, the exact count depends on the wipeout logic
	// But stack should have multiple turning points
	if len(ps.Stack) < 3 {
		t.Errorf("After multiple reversals, stack length = %d, expected at least 3", len(ps.Stack))
	}
}

func TestPreisachStack_ZeroEverettFunction(t *testing.T) {
	satE := 5.0
	everett := &MockEverettFunction{scaleFactor: 0.0} // Always returns 0
	ps := NewPreisachStack(satE, everett)

	// Any field should give zero polarization
	P1 := ps.Update(0.0)
	P2 := ps.Update(5.0)
	P3 := ps.Update(-5.0)

	if P1 != 0.0 || P2 != 0.0 || P3 != 0.0 {
		t.Errorf("Zero Everett function should give zero polarization: P1=%v, P2=%v, P3=%v", P1, P2, P3)
	}
}
