package hysteresis

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestBuildLevelCalibrationExportPayloadSummarizesCalibration(t *testing.T) {
	m := New()
	if err := m.ApplyAction(viewmodel.Action{ID: EventRunLevelCalibration, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run level calibration: %v", err)
	}

	payload := buildLevelCalibrationExportPayload(m.state.LevelCalibration)

	if payload.ArtifactType != "level_calibration" {
		t.Fatalf("artifact type = %q", payload.ArtifactType)
	}
	if payload.Status != LevelCalibrationFresh {
		t.Fatalf("status = %q, want fresh", payload.Status)
	}
	if payload.Inputs.LevelCount != m.state.LevelCalibration.Result.LevelCount {
		t.Fatalf("level count = %d, want %d", payload.Inputs.LevelCount, m.state.LevelCalibration.Result.LevelCount)
	}
	if len(payload.Levels) != len(m.state.LevelCalibration.Result.AscendingFields) {
		t.Fatalf("levels = %d, want %d", len(payload.Levels), len(m.state.LevelCalibration.Result.AscendingFields))
	}
	if payload.Summary.Monotonicity == "" {
		t.Fatal("missing monotonicity summary")
	}
}
