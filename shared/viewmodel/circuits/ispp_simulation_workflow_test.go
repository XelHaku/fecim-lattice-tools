package circuits

import "testing"

func TestISPPSimulationWorkflowComputesLevelAttempts(t *testing.T) {
	updated := newISPPSimulationWorkflow(CircuitsState{QuantLevels: 3}).compute()

	if !updated.ISPPExecuted {
		t.Fatal("ISPPExecuted = false, want true")
	}
	if updated.QuantLevels != 3 {
		t.Fatalf("QuantLevels = %d, want 3", updated.QuantLevels)
	}
	if len(updated.ISPPAttempts) != 3 || len(updated.ISPPConverged) != 3 {
		t.Fatalf("ISPP result lengths = attempts %d converged %d, want 3/3", len(updated.ISPPAttempts), len(updated.ISPPConverged))
	}
	if updated.ISPPTotalAttempts <= 0 {
		t.Fatalf("ISPPTotalAttempts = %d, want positive", updated.ISPPTotalAttempts)
	}
	if updated.ISPPAvgAttempts != float64(updated.ISPPTotalAttempts)/3 {
		t.Fatalf("ISPPAvgAttempts = %.6f, want total/3 from %d", updated.ISPPAvgAttempts, updated.ISPPTotalAttempts)
	}
	if updated.ISPPConvergedCount < 0 || updated.ISPPConvergedCount > 3 {
		t.Fatalf("ISPPConvergedCount = %d, want within level count", updated.ISPPConvergedCount)
	}
	for level, attempts := range updated.ISPPAttempts {
		if attempts <= 0 {
			t.Fatalf("attempts[%d] = %d, want positive", level, attempts)
		}
	}
}
