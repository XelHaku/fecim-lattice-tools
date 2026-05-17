//go:build legacy_fyne

package gui

import (
	"testing"

	"fecim-lattice-tools/shared/physics"
)

func TestM5Data02_ProvenanceTagsPresent(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	for name, in := range run.Inputs {
		if in.Tag.Provenance == "" {
			t.Fatalf("missing provenance for input %s", name)
		}
	}
	if run.Inputs["fecim_energy_pj_per_mac"].Tag.Provenance != physics.ProvenancePlaceholder {
		t.Fatalf("expected placeholder provenance for FeCIM energy")
	}
}
