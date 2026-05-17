//go:build legacy_fyne

package gui

import (
	"fecim-lattice-tools/shared/presets"
)

// Ensure ComparisonPresetProvider implements PresetProvider.
var _ presets.PresetProvider = (*ComparisonPresetProvider)(nil)

// ComparisonPresetProvider implements preset support for the technology comparison module.
type ComparisonPresetProvider struct {
	app *ComparisonApp
}

// NewComparisonPresetProvider creates a new preset provider for the comparison module.
func NewComparisonPresetProvider(app *ComparisonApp) *ComparisonPresetProvider {
	return &ComparisonPresetProvider{app: app}
}

// GetModule returns the module identifier.
func (cp *ComparisonPresetProvider) GetModule() presets.Module {
	return presets.ModuleComparison
}

// GetPresetKeys returns the list of configuration keys supported.
func (cp *ComparisonPresetProvider) GetPresetKeys() []string {
	return presets.ComparisonPresetKeys
}

// GetCurrentConfig returns the current configuration as a preset config map.
func (cp *ComparisonPresetProvider) GetCurrentConfig() map[string]interface{} {
	cp.app.animMu.RLock()
	defer cp.app.animMu.RUnlock()

	config := make(map[string]interface{})

	config["comparison_mode"] = cp.app.uiMode
	config["workload"] = cp.app.currentWorkload
	config["inferences"] = cp.app.currentInferences
	config["animation"] = cp.app.running

	return config
}

// ApplyPreset applies a preset configuration to the ComparisonApp.
func (cp *ComparisonPresetProvider) ApplyPreset(preset *presets.Preset) error {
	if workload, ok := preset.GetString("workload"); ok {
		// onWorkloadChanged updates currentWorkload and schedules recalculation.
		cp.app.onWorkloadChanged(workload)
		if cp.app.workloadSelect != nil {
			cp.app.workloadSelect.SetSelected(workload)
		}
	}

	if inferences, ok := preset.GetFloat("inferences"); ok {
		cp.app.animMu.Lock()
		cp.app.currentInferences = inferences
		cp.app.animMu.Unlock()
		if cp.app.inferencesSlider != nil {
			cp.app.inferencesSlider.SetValue(inferences)
		}
		cp.app.scheduleRecompute()
	}

	if mode, ok := preset.GetString("comparison_mode"); ok {
		cp.app.animMu.Lock()
		cp.app.uiMode = mode
		cp.app.animMu.Unlock()
	}

	return nil
}
