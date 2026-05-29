package circuits

import "testing"

func TestHalfSelectStressWorkflowComputesPassiveWriteStress(t *testing.T) {
	updated := newHalfSelectStressWorkflow(CircuitsState{
		Rows:          8,
		OperationMode: OperationWrite,
		Architecture:  ArchitecturePassive,
	}).compute()

	if updated.HalfSelectState != HalfSelectStateColumnWriteActive {
		t.Fatalf("half-select state = %q", updated.HalfSelectState)
	}
	if updated.HalfSelectCells != 7 {
		t.Fatalf("half-select cells = %d, want 7", updated.HalfSelectCells)
	}
	if updated.DisturbVoltage != DefaultDisturbVoltage {
		t.Fatalf("disturb voltage = %.6f, want %.6f", updated.DisturbVoltage, DefaultDisturbVoltage)
	}
	if updated.StressPerPulse != PassiveStressPerPulse {
		t.Fatalf("stress per pulse = %.6f, want %.6f", updated.StressPerPulse, PassiveStressPerPulse)
	}
	wantCycles := int(1 / PassiveStressPerPulse)
	if updated.StressCyclesToLevel != wantCycles {
		t.Fatalf("cycles to level = %d, want %d", updated.StressCyclesToLevel, wantCycles)
	}
}

func TestHalfSelectStressWorkflowClearsNonWriteStress(t *testing.T) {
	updated := newHalfSelectStressWorkflow(CircuitsState{
		Rows:                8,
		OperationMode:       OperationRead,
		Architecture:        ArchitecturePassive,
		HalfSelectState:     HalfSelectStateColumnWriteActive,
		HalfSelectCells:     7,
		DisturbVoltage:      DefaultDisturbVoltage,
		StressPerPulse:      PassiveStressPerPulse,
		StressCyclesToLevel: int(1 / PassiveStressPerPulse),
	}).compute()

	if updated.HalfSelectState != HalfSelectStateInactive || updated.HalfSelectCells != 0 || updated.DisturbVoltage != 0 || updated.StressPerPulse != 0 || updated.StressCyclesToLevel != 0 {
		t.Fatalf("non-write stress not cleared: %+v", updated)
	}
}
