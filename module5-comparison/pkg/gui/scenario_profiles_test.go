//go:build legacy_fyne

package gui

import "testing"

func TestPresetScenarioConfig(t *testing.T) {
	for _, profile := range []ScenarioProfile{ScenarioConservative, ScenarioBaseline, ScenarioOptimistic} {
		cfg := PresetScenarioConfig(profile)
		if err := cfg.Validate(); err != nil {
			t.Fatalf("profile %s invalid: %v", profile, err)
		}
		if cfg.Name == "" {
			t.Fatalf("profile %s missing name", profile)
		}
	}
}
