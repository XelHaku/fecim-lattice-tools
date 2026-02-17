package crossbar

import (
	"testing"
)

// TestUnifiedSolver_Bidirectional verifies that the solver handles
// reverse currents (negative effective voltage) correctly without clamping.
func TestUnifiedSolver_Bidirectional(t *testing.T) {
	// Setup a small 3x3 array
	rows, cols := 3, 3
	sim := NewIRDropSimulator(rows, cols)

	// Create a scenario where reverse current MUST flow if physics is correct.
	//
	// Driver: Row 0 = 1.0V
	// Sense:  Col 0 = 0.0V
	//
	// Heavy blockage:
	// Cell(0,0) is HIGH resistance (G ~ 0)
	//
	// Sneak path:
	// Row 0 -> Cell(0,1) -> Col 1 -> Cell(1,1) -> Row 1 -> Cell(1,0) -> Col 0
	//
	// To force this:
	// G(0,1) = High, G(1,1) = High, G(1,0) = High
	// G(0,0) = Low
	// All other G = Low
	//
	// In the middle of this path (e.g. at Row 1), the voltage potential
	// must allow current to flow "backwards" or "sideways" relative to the main drive.
	// Specifically, for current to flow Col 1 -> Row 1, V_col1 must be > V_row1.
	// This implies V_row1 - V_col1 < 0.
	// We expect Cell(1,1) to have negative effective voltage/current if visualized as Row-Col.

	highG := 100e-6 // 100 uS
	lowG := 1e-9    // 1 nS (effectively open)

	// Initialize all to low
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			sim.SetConductance(i, j, lowG)
		}
	}

	// Set sneak path cells to high
	sim.SetConductance(0, 1, highG) // Path segment A
	sim.SetConductance(1, 1, highG) // Path segment B (Vertical leak)
	sim.SetConductance(1, 0, highG) // Path segment C

	// Set blockage
	sim.SetConductance(0, 0, lowG)

	// High wire resistance to accentuate IR drop effects and make voltages diverge
	sim.RowResist = 10.0
	sim.ColResist = 10.0

	// Apply input
	sim.SetInputVoltage(0, 1.0)
	sim.SetInputVoltage(1, 0.0) // Floating/Ground? If driver is 0, it sinks.
	// To test pure sneak, we might want floating rows, but the solver assumes driven rows.
	// If Row 1 is driven to 0V, current will flow into it.
	// The test is: Does current flow through Cell(1,1)?

	sim.Simulate(100)

	// Check Cell(1,1) current.
	// V_row0_col1 should be High (~1V - drops)
	// V_col1_row1 should be High (connected to row0 via G)
	// V_row1_col1 is 0V (driven).
	// So V_row1 - V_col1 should be negative (0 - High).
	// Current should be negative.

	curr11 := sim.CellCurrents[1][1]

	t.Logf("Current at (1,1): %v", curr11)

	if curr11 >= 0 {
		t.Errorf("Expected negative current (sneak path) at (1,1), got %v. Clamping might still be active.", curr11)
	}

	// Check that we have non-zero output at Col 0 despite blockage at (0,0)
	// Output at Col 0 comes from the sneak path ending at Cell(1,0)
	outCurrs := sim.GetOutputCurrents()
	col0Curr := outCurrs[0]

	t.Logf("Output Current Col 0: %v", col0Curr)

	if col0Curr <= 1e-9 {
		t.Errorf("Expected significant sneak current at Col 0, got %v", col0Curr)
	}
}

func TestUnifiedSolver_ShortCircuit(t *testing.T) {
	rows, cols := 3, 3
	sim := NewIRDropSimulator(rows, cols)

	// Short circuit at (0,0)
	// Should see massive voltage drop on the wire leading to it.
	sim.SetConductance(0, 0, 1.0) // 1 Siemens (Short)
	sim.RowResist = 10.0
	sim.SetInputVoltage(0, 1.0)

	sim.Simulate(100)

	// Voltage at (0,0) row node should be very low due to IR drop across the short
	// V_out = V_in * (R_cell / (R_wire + R_cell))
	// R_cell = 1 ohm, R_wire = 10 ohm -> V_out ~ 1/11 ~ 0.09V
	// Wait, grid is more complex, but qualitative drop should be huge.

	vRow00 := sim.RowVoltages[0][0]
	t.Logf("Voltage at Row(0,0): %v", vRow00)

	if vRow00 > 0.5 {
		t.Errorf("Expected significant voltage drop at shorted cell, got %v", vRow00)
	}
}

func TestUnifiedSolver_NoClamping(t *testing.T) {
	// Explicitly test that V_row < V_col results in negative effective voltage
	// independent of the mesh connectivity (force it if possible, or infer from result).

	// We can re-use the bidirectional test logic but focus on the effective voltage/IR drop map.
	rows, cols := 2, 2
	sim := NewIRDropSimulator(rows, cols)

	// Set typical conductances
	sim.SetConductance(0, 0, 50e-6)
	sim.SetConductance(0, 1, 50e-6)
	sim.SetConductance(1, 0, 50e-6)
	sim.SetConductance(1, 1, 50e-6)

	sim.VoltageIn[0] = 0.5
	sim.VoltageOut[0] = 1.0 // Sense node at 1.0V (unusual but valid for test)

	sim.Simulate(10)

	// At (0,0), V_row ~ 0.5, V_col ~ 1.0
	// effV = 0.5 - 1.0 = -0.5
	// If clamped, would be 0.

	cellI := sim.CellCurrents[0][0]
	if cellI >= 0 {
		t.Errorf("Expected negative current with V_in=0.5, V_out=1.0, got %v", cellI)
	}
}
