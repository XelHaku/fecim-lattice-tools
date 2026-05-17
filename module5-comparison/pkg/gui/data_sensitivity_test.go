//go:build legacy_fyne

package gui

import "testing"

func TestM5Data04_SensitivityRankingSorted(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	r := SensitivityRanking(run, "tco_reduction_pct")
	if len(r) == 0 {
		t.Fatal("expected non-empty ranking")
	}
	if r[0].ImpactPct < r[len(r)-1].ImpactPct {
		t.Fatal("expected sorted descending by impact")
	}
}
