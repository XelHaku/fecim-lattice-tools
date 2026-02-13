package arraysim

import "testing"

func TestRunWriteVerifyLoop_Converges(t *testing.T) {
	model := LinearProgramModel{Gain: 2e-6, Gmin: 1e-6, Gmax: 150e-6}
	cfg := WriteVerifyConfig{TargetG: 60e-6, Tolerance: 1e-6, StartVoltage: 0.2, StepVoltage: 0.05, MaxIters: 50}
	res := RunWriteVerifyLoop(20e-6, model, cfg)
	if !res.Converged {
		t.Fatalf("expected converged, got %+v", res)
	}
	if res.FinalG < 59e-6 || res.FinalG > 61e-6 {
		t.Fatalf("final conductance out of band: %.9f", res.FinalG)
	}
}

func TestRunWriteVerifyLoop_StopsWhenNoProgress(t *testing.T) {
	model := LinearProgramModel{Gain: 0, Gmin: 1e-6, Gmax: 150e-6}
	cfg := WriteVerifyConfig{TargetG: 100e-6, Tolerance: 1e-6, StartVoltage: 0.2, StepVoltage: 0.05, MaxIters: 5}
	res := RunWriteVerifyLoop(20e-6, model, cfg)
	if res.Converged {
		t.Fatalf("unexpected convergence: %+v", res)
	}
	if res.Iterations != 5 {
		t.Fatalf("iterations=%d want 5", res.Iterations)
	}
}
