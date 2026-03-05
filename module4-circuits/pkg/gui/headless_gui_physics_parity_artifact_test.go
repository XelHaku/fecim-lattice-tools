package gui

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func writeParityArtifact(t interface{ Logf(string, ...any) }, art *parityArtifact) {
	if art == nil {
		return
	}
	// Default output path is predictable for CI archiving.
	// Allow override for experiments.
	path := os.Getenv("FECIM_PARITY_JSON_PATH")
	if path == "" {
		base := os.Getenv("FECIM_M4_REGRESSION_JSON_DIR")
		if base == "" {
			base = filepath.Join("output", "regression", "module4")
		}
		path = filepath.Join(base, "gui_vs_headless_parity.json")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Logf("parity artifact: mkdir failed: %v", err)
		return
	}

	// Ensure profile populated even if not set.
	if art.Profile == "" {
		art.Profile = os.Getenv("FECIM_MATERIAL_PROFILE")
		if art.Profile == "" {
			art.Profile = "pr"
		}
	}
	if art.Version == "" {
		art.Version = "v1"
	}
	// Force deterministic metadata for regression artifacts.
	art.GeneratedUnix = 0

	b, err := json.MarshalIndent(art, "", "  ")
	if err != nil {
		t.Logf("parity artifact: marshal failed: %v", err)
		return
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Logf("parity artifact: write failed: %v", err)
		return
	}
	// Best-effort log.
	t.Logf("parity artifact written: %s (%d bytes)", path, len(b))
}

func compareSnapshotAndRecord(t interface {
	Errorf(string, ...any)
}, art *parityArtifact, material, scenario, step string, gui, headless paritySnapshot, tolV, tolI float64) []parityStepArtifact {
	records := make([]parityStepArtifact, 0, 3)

	// Voltage
	dV := math.Abs(gui.EffectiveCellVoltage - headless.EffectiveCellVoltage)
	passV := dV <= tolV
	if !passV {
		t.Errorf("parity(%s/%s/%s): effective voltage delta %.6g V > tol %.6g V", material, scenario, step, dV, tolV)
	}
	records = append(records, parityStepArtifact{
		Material:         material,
		Scenario:         scenario,
		StepType:         step,
		Tolerance:        tolV,
		MaxObservedDelta: dV,
		Pass:             passV,
		FailingIndex:     0,
		DeltaKind:        "voltage",
	})

	// Row currents
	maxDI := 0.0
	idxI := -1
	for i := range gui.RowCurrentsUA {
		if i >= len(headless.RowCurrentsUA) {
			break
		}
		d := math.Abs(gui.RowCurrentsUA[i] - headless.RowCurrentsUA[i])
		if d > maxDI {
			maxDI = d
			idxI = i
		}
	}
	passI := maxDI <= tolI
	if !passI {
		t.Errorf("parity(%s/%s/%s): row current max delta %.6g uA > tol %.6g uA (row=%d)", material, scenario, step, maxDI, tolI, idxI)
	}
	records = append(records, parityStepArtifact{
		Material:         material,
		Scenario:         scenario,
		StepType:         step,
		Tolerance:        tolI,
		MaxObservedDelta: maxDI,
		Pass:             passI,
		FailingIndex:     idxI,
		DeltaKind:        "current_ua",
	})

	// Levels (exact match)
	idxL := -1
	for i := range gui.RowLevels {
		if i >= len(headless.RowLevels) {
			break
		}
		if gui.RowLevels[i] != headless.RowLevels[i] {
			idxL = i
			break
		}
	}
	passL := idxL == -1
	if !passL {
		t.Errorf("parity(%s/%s/%s): row level mismatch at row=%d (gui=%d headless=%d)", material, scenario, step, idxL, gui.RowLevels[idxL], headless.RowLevels[idxL])
	}
	records = append(records, parityStepArtifact{
		Material:  material,
		Scenario:  scenario,
		StepType:  step,
		Tolerance: 0,
		MaxObservedDelta: func() float64 {
			if passL {
				return 0
			}
			return 1
		}(),
		Pass:         passL,
		FailingIndex: idxL,
		DeltaKind:    "level",
	})

	_ = art // art kept for signature stability
	return records
}

func compareWriteAndRecord(t interface {
	Errorf(string, ...any)
}, material, scenario string, gui, headless parityWriteTrace, tol float64) parityStepArtifact {
	// For write-step we fold multiple comparisons into one record.
	maxD := 0.0
	maxD = math.Max(maxD, math.Abs(gui.RequestedVoltage-headless.RequestedVoltage))
	maxD = math.Max(maxD, math.Abs(gui.AppliedVoltage-headless.AppliedVoltage))
	maxD = math.Max(maxD, math.Abs(gui.EffectiveVoltage-headless.EffectiveVoltage))

	pass := true
	if maxD > tol {
		pass = false
		t.Errorf("parity(%s/%s/WRITE_STEP): voltage delta %.6g V > tol %.6g V", material, scenario, maxD, tol)
	}
	if gui.DACCode != headless.DACCode {
		pass = false
		t.Errorf("parity(%s/%s/WRITE_STEP): DACCode mismatch gui=%d headless=%d", material, scenario, gui.DACCode, headless.DACCode)
	}
	if gui.Trajectory != headless.Trajectory {
		pass = false
		t.Errorf("parity(%s/%s/WRITE_STEP): trajectory mismatch gui=%v headless=%v", material, scenario, gui.Trajectory, headless.Trajectory)
	}
	if gui.Result != headless.Result {
		pass = false
		t.Errorf("parity(%s/%s/WRITE_STEP): ISPP result mismatch gui=%v headless=%v", material, scenario, gui.Result, headless.Result)
	}

	return parityStepArtifact{
		Material:           material,
		Scenario:           scenario,
		StepType:           "WRITE_STEP",
		Tolerance:          tol,
		MaxObservedDelta:   maxD,
		Pass:               pass,
		FailingIndex:       -1,
		DeltaKind:          "write",
		AdditionalMetadata: fmt.Sprintf("dacCode=%d", gui.DACCode),
	}
}
