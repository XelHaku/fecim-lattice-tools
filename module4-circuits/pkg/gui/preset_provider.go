//go:build legacy_fyne

package gui

import (
	"fecim-lattice-tools/shared/presets"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Ensure CircuitsPresetProvider implements PresetProvider.
var _ presets.PresetProvider = (*CircuitsPresetProvider)(nil)

// CircuitsPresetProvider implements preset support for the peripheral circuits module.
type CircuitsPresetProvider struct {
	app *CircuitsApp
}

// NewCircuitsPresetProvider creates a new preset provider for the circuits module.
func NewCircuitsPresetProvider(app *CircuitsApp) *CircuitsPresetProvider {
	return &CircuitsPresetProvider{app: app}
}

// GetModule returns the module identifier.
func (cp *CircuitsPresetProvider) GetModule() presets.Module {
	return presets.ModuleCircuits
}

// GetPresetKeys returns the list of configuration keys supported.
func (cp *CircuitsPresetProvider) GetPresetKeys() []string {
	return presets.CircuitsPresetKeys
}

// GetCurrentConfig returns the current configuration as a preset config map.
func (cp *CircuitsPresetProvider) GetCurrentConfig() map[string]interface{} {
	cp.app.mu.RLock()
	defer cp.app.mu.RUnlock()

	config := make(map[string]interface{})

	config["adc_bits"] = cp.app.adcBits
	config["dac_bits"] = cp.app.dacBits
	config["num_levels"] = cp.app.quantLevels
	config["architecture"] = cp.app.architecture

	return config
}

// ApplyPreset applies a preset configuration to the CircuitsApp.
func (cp *CircuitsPresetProvider) ApplyPreset(preset *presets.Preset) error {
	cp.app.mu.Lock()

	if bits, ok := preset.GetInt("adc_bits"); ok {
		cp.app.adcBits = bits
	}
	if bits, ok := preset.GetInt("dac_bits"); ok {
		cp.app.dacBits = bits
	}
	if levels, ok := preset.GetInt("num_levels"); ok {
		cp.app.quantLevels = levels
	}
	if arch, ok := preset.GetString("architecture"); ok {
		switch arch {
		case "1T1R":
			cp.app.architecture = sharedwidgets.Architecture1T1R
		case "0T1R", "passive":
			cp.app.architecture = sharedwidgets.Architecture0T1R
		}
	}

	cp.app.mu.Unlock()

	// Re-render the write array to reflect field changes.
	cp.app.refreshWriteArray()

	return nil
}
