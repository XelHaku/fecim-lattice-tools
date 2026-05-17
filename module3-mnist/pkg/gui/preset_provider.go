//go:build legacy_fyne

package gui

import (
	"fecim-lattice-tools/shared/presets"
)

// Ensure MNISTPresetProvider implements PresetProvider.
var _ presets.PresetProvider = (*MNISTPresetProvider)(nil)

// MNISTPresetProvider implements preset support for the MNIST dual-mode module.
type MNISTPresetProvider struct {
	app *DualModeApp
}

// NewMNISTPresetProvider creates a new preset provider for the MNIST module.
func NewMNISTPresetProvider(app *DualModeApp) *MNISTPresetProvider {
	return &MNISTPresetProvider{app: app}
}

// GetModule returns the module identifier.
func (mp *MNISTPresetProvider) GetModule() presets.Module {
	return presets.ModuleMNIST
}

// GetPresetKeys returns the list of configuration keys supported.
func (mp *MNISTPresetProvider) GetPresetKeys() []string {
	return presets.MNISTPresetKeys
}

// GetCurrentConfig returns the current configuration as a preset config map.
func (mp *MNISTPresetProvider) GetCurrentConfig() map[string]interface{} {
	config := make(map[string]interface{})

	// DualModeApp always runs both FP and CIM paths side-by-side.
	config["mode"] = "cim"

	// Quantization level (number of discrete analog states).
	config["quantization"] = mp.app.networkCtrl.GetNumLevels()

	// Noise
	noiseLevel := FeCIMDefaultNoise
	if mp.app.noiseSlider != nil {
		noiseLevel = mp.app.noiseSlider.Value
	}
	config["noise_level"] = noiseLevel
	config["noise_enabled"] = noiseLevel > 0

	// Weight heatmap visibility (collapsed = hidden).
	config["show_weights"] = !mp.app.weightsCollapsed

	// Inference animation phase overlay.
	config["show_activations"] = mp.app.animationEnabled

	// Energy tracker and comparison card are always present once the UI is built;
	// expose their presence as a snapshot so presets round-trip cleanly.
	config["show_energy"] = mp.app.energyWidget != nil
	config["show_comparison"] = mp.app.comparisonCard != nil

	// Quick demo (auto digit cycling).
	config["auto_recognize"] = mp.app.quickDemoRunning

	// batch_size has no per-session field; return the canonical default.
	config["batch_size"] = 100

	return config
}

// ApplyPreset applies a preset configuration to the DualModeApp.
func (mp *MNISTPresetProvider) ApplyPreset(preset *presets.Preset) error {
	// Determine the target levels and noise, defaulting to current values so
	// a single applyPreset call covers both numeric settings atomically.
	levels := mp.app.networkCtrl.GetNumLevels()
	noiseLevel := FeCIMDefaultNoise
	if mp.app.noiseSlider != nil {
		noiseLevel = mp.app.noiseSlider.Value
	}

	levelsChanged := false
	noiseChanged := false

	if q, ok := preset.GetInt("quantization"); ok {
		levels = q
		levelsChanged = true
	}

	if nl, ok := preset.GetFloat("noise_level"); ok {
		noiseLevel = nl
		noiseChanged = true
	}

	// If noise is explicitly disabled, zero it out regardless of noise_level.
	if enabled, ok := preset.GetBool("noise_enabled"); ok && !enabled {
		noiseLevel = 0
		noiseChanged = true
	}

	if levelsChanged || noiseChanged {
		// applyPreset also updates the level selector, noise slider, and triggers
		// a full re-inference so all widgets stay consistent.
		mp.app.applyPreset(levels, noiseLevel, FeCIMDefaultADC, FeCIMDefaultDAC)
	}

	// Weight heatmap visibility.
	if showWeights, ok := preset.GetBool("show_weights"); ok {
		if showWeights && mp.app.weightsCollapsed {
			mp.app.toggleWeightsCollapsed()
		} else if !showWeights && !mp.app.weightsCollapsed {
			mp.app.toggleWeightsCollapsed()
		}
	}

	// Inference animation overlay.
	if showActivations, ok := preset.GetBool("show_activations"); ok {
		mp.app.animationEnabled = showActivations
	}

	return nil
}
