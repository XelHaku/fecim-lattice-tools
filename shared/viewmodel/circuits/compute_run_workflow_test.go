package circuits

import "testing"

func TestComputeRunWorkflowBuildsDetailedMVMSummary(t *testing.T) {
	log := newComputeRunWorkflow(CircuitsState{
		Rows:          2,
		Cols:          2,
		QuantLevels:   DefaultQuantLevels,
		Architecture:  ArchitecturePassive,
		CouplingTier:  CouplingTierA,
		TIAGain:       1e4,
		SupplyVoltage: 1.8,
		ADCResolution: 5,
	}).buildLog()

	if log == nil {
		t.Fatal("log = nil")
	}
	if log.Schema != "fecim.circuits.compute_run.v1" || log.ArraySize != "2x2" || log.ExportedCells != 4 {
		t.Fatalf("unexpected header: %+v", log)
	}
	if len(log.InputVector) != 2 || !near(log.InputVector[0], 0.2, 1e-9) || !near(log.InputVector[1], 0.5, 1e-9) {
		t.Fatalf("input vector = %+v", log.InputVector)
	}
	if log.Weights[0][0] != 0 || log.Weights[0][1] != 1 || log.Weights[1][0] != 2 || log.Weights[1][1] != 3 {
		t.Fatalf("weights = %+v", log.Weights)
	}
	row0 := log.RowResults[0]
	if row0.Row != 0 || !row0.Active || row0.Saturated || row0.ADCLevel != 0 {
		t.Fatalf("unexpected row0 flags: %+v", row0)
	}
	if !near(row0.CurrentUA, 2.4068965517, 1e-9) || !near(row0.TIAVoltage, 0.0240689655, 1e-9) {
		t.Fatalf("unexpected row0 analog terms: %+v", row0)
	}
	row1 := log.RowResults[1]
	if row1.ADCLevel != 1 || !near(row1.CurrentUA, 7.1862068966, 1e-9) {
		t.Fatalf("unexpected row1: %+v", row1)
	}
}
