package gui

import "testing"

func TestM5Data03_ConfidenceIntervalsOnKeyOutputs(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	for _, k := range []string{"energy_reduction_pct", "latency_reduction_pct", "tco_reduction_pct", "co2_reduction_pct"} {
		if run.Outputs[k].Uncertainty <= 0 {
			t.Fatalf("expected positive uncertainty for %s", k)
		}
	}
}
