package gui

import "testing"

func TestM5Data05_ScenarioDiffReport(t *testing.T) {
	a := BuildScenarioRun(ScenarioConservative, PresetScenarioConfig(ScenarioConservative))
	b := BuildScenarioRun(ScenarioOptimistic, PresetScenarioConfig(ScenarioOptimistic))
	d := DiffScenarios(a, b)
	if len(d.ChangedAssumptions) == 0 || len(d.OutputDeltas) == 0 || len(d.Attribution) == 0 {
		t.Fatalf("incomplete diff report: %+v", d)
	}
}
