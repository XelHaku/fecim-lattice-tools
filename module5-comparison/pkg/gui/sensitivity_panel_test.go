package gui

import (
	"testing"

	"fecim-lattice-tools/shared/physics"
)

func TestSensitivityPanel_NewPanelNotNil(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	sp := NewSensitivityPanel(run, "tco_reduction_pct")
	if sp == nil {
		t.Fatal("expected non-nil SensitivityPanel")
	}
}

func TestSensitivityPanel_BuildContentWithEmptyRun(t *testing.T) {
	sp := NewSensitivityPanel(ScenarioRun{}, "tco_reduction_pct")
	content := sp.buildContent()
	if content == nil {
		t.Fatal("expected non-nil content even with empty run")
	}
}

func TestSensitivityPanel_BuildContentWithBaselineRun(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	sp := NewSensitivityPanel(run, "tco_reduction_pct")
	content := sp.buildContent()
	if content == nil {
		t.Fatal("expected non-nil content for baseline run")
	}
}

func TestSensitivityPanel_BuildParameterList(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	sp := NewSensitivityPanel(run, "energy_reduction_pct")
	list := sp.buildParameterList()
	if list == nil {
		t.Fatal("expected non-nil parameter list")
	}
}

func TestSensitivityPanel_BuildTornadoChart(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	sp := NewSensitivityPanel(run, "tco_reduction_pct")
	chart := sp.buildTornadoChart()
	if chart == nil {
		t.Fatal("expected non-nil tornado chart")
	}
}

func TestSensitivityPanel_TornadoChartNoOutput(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	sp := NewSensitivityPanel(run, "nonexistent_output")
	chart := sp.buildTornadoChart()
	if chart == nil {
		t.Fatal("expected non-nil tornado chart even with unknown output")
	}
}

func TestSensitivityPanel_SetRunUpdates(t *testing.T) {
	run1 := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	sp := NewSensitivityPanel(run1, "tco_reduction_pct")

	run2 := BuildScenarioRun(ScenarioOptimistic, PresetScenarioConfig(ScenarioOptimistic))
	sp.SetRun(run2, "energy_reduction_pct")

	if sp.outputName != "energy_reduction_pct" {
		t.Fatalf("expected outputName updated, got %s", sp.outputName)
	}
	if sp.run.Profile != ScenarioOptimistic {
		t.Fatalf("expected profile updated to optimistic, got %s", sp.run.Profile)
	}
}

func TestSensitivityPanel_AllOutputMetrics(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	outputs := []string{"energy_reduction_pct", "latency_reduction_pct", "tco_reduction_pct", "co2_reduction_pct"}
	for _, out := range outputs {
		sp := NewSensitivityPanel(run, out)
		chart := sp.buildTornadoChart()
		if chart == nil {
			t.Fatalf("nil tornado chart for output %s", out)
		}
	}
}

func TestSensitivityPanel_ProvenanceColorCoverage(t *testing.T) {
	cases := []physics.Provenance{
		physics.ProvenanceMeasured,
		physics.ProvenanceCalibrated,
		physics.ProvenanceEstimated,
		physics.ProvenancePlaceholder,
		physics.Provenance("unknown"),
	}
	for _, p := range cases {
		c := provenanceColor(p)
		if c == nil {
			t.Fatalf("nil color for provenance %s", p)
		}
	}
}

func TestSensitivityPanel_TornadoBarColorThresholds(t *testing.T) {
	cases := []float64{0.5, 3.0, 5.0, 8.0, 15.0}
	for _, pct := range cases {
		c := tornadoBarColor(pct)
		if c == nil {
			t.Fatalf("nil color for impact %.1f%%", pct)
		}
	}
}

func TestSensitivityPanel_MinSize(t *testing.T) {
	sp := NewSensitivityPanel(ScenarioRun{}, "tco_reduction_pct")
	size := sp.MinSize()
	if size.Width < 300 || size.Height < 200 {
		t.Fatalf("MinSize too small: %v", size)
	}
}

func TestSensitivityPanel_ConservativeScenario(t *testing.T) {
	run := BuildScenarioRun(ScenarioConservative, PresetScenarioConfig(ScenarioConservative))
	sp := NewSensitivityPanel(run, "tco_reduction_pct")
	content := sp.buildContent()
	if content == nil {
		t.Fatal("expected non-nil content for conservative scenario")
	}
}
