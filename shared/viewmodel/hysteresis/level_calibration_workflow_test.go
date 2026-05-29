package hysteresis

import (
	"testing"

	"fecim-lattice-tools/shared/physics"
)

func TestLevelCalibrationWorkflowRunProducesFreshSimulationSummary(t *testing.T) {
	mat := physics.DefaultHZO()
	workflow := newLevelCalibrationWorkflow(defaultLevelCalibrationForMaterial(mat), mat)

	state, err := workflow.run()
	if err != nil {
		t.Fatalf("run level calibration workflow: %v", err)
	}
	if !state.HasResult {
		t.Fatal("workflow state has no result")
	}
	if state.Status != LevelCalibrationFresh {
		t.Fatalf("workflow status = %q, want %q", state.Status, LevelCalibrationFresh)
	}
	if state.Result.MaterialName != mat.Name {
		t.Fatalf("workflow material = %q, want %q", state.Result.MaterialName, mat.Name)
	}
	if state.Error != "" {
		t.Fatalf("workflow error = %q, want empty", state.Error)
	}
}

func TestLevelCalibrationWorkflowSetLevelCountMarksFreshResultStale(t *testing.T) {
	mat := physics.DefaultHZO()
	fresh := freshLevelCalibrationState(t, mat)

	updated, err := newLevelCalibrationWorkflow(fresh, mat).setLevelCount("16")
	if err != nil {
		t.Fatalf("set level count: %v", err)
	}
	if updated.LevelCount != 16 {
		t.Fatalf("level count = %d, want 16", updated.LevelCount)
	}
	if updated.Status != LevelCalibrationStale {
		t.Fatalf("status = %q, want %q", updated.Status, LevelCalibrationStale)
	}
	if !updated.HasResult || updated.Result.LevelCount != 30 {
		t.Fatalf("stale transition should preserve previous 30-level result: %+v", updated.Result)
	}
}

func TestLevelCalibrationWorkflowSetTargetRangeMarksFreshResultStale(t *testing.T) {
	mat := physics.DefaultHZO()
	fresh := freshLevelCalibrationState(t, mat)

	updated, err := newLevelCalibrationWorkflow(fresh, mat).setTargetRange("0.70")
	if err != nil {
		t.Fatalf("set target range: %v", err)
	}
	if updated.TargetRangeFrac != 0.70 {
		t.Fatalf("target range = %.2f, want 0.70", updated.TargetRangeFrac)
	}
	if updated.Status != LevelCalibrationStale {
		t.Fatalf("status = %q, want %q", updated.Status, LevelCalibrationStale)
	}
	if !updated.HasResult || updated.Result.TargetRangeFrac == updated.TargetRangeFrac {
		t.Fatalf("stale transition should preserve previous result target range: state=%+v result=%+v", updated, updated.Result)
	}
}

func TestLevelCalibrationWorkflowSetTemperatureMarksFreshResultStale(t *testing.T) {
	mat := physics.DefaultHZO()
	fresh := freshLevelCalibrationState(t, mat)

	updated, err := newLevelCalibrationWorkflow(fresh, mat).setTemperature("350")
	if err != nil {
		t.Fatalf("set temperature: %v", err)
	}
	if updated.TemperatureK != 350 {
		t.Fatalf("temperature = %.1f, want 350", updated.TemperatureK)
	}
	if updated.Status != LevelCalibrationStale {
		t.Fatalf("status = %q, want %q", updated.Status, LevelCalibrationStale)
	}
	if !updated.HasResult || updated.Result.TemperatureK == updated.TemperatureK {
		t.Fatalf("stale transition should preserve previous result temperature: state=%+v result=%+v", updated, updated.Result)
	}
}

func freshLevelCalibrationState(t *testing.T, mat *physics.HZOMaterial) LevelCalibrationState {
	t.Helper()
	workflow := newLevelCalibrationWorkflow(defaultLevelCalibrationForMaterial(mat), mat)
	fresh, err := workflow.run()
	if err != nil {
		t.Fatalf("run level calibration workflow: %v", err)
	}
	return fresh
}
