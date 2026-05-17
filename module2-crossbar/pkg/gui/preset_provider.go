//go:build legacy_fyne

package gui

import (
	"fecim-lattice-tools/shared/presets"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// Ensure CrossbarPresetProvider implements PresetProvider
var _ presets.PresetProvider = (*CrossbarPresetProvider)(nil)

// CrossbarPresetProvider implements preset support for the crossbar module
type CrossbarPresetProvider struct {
	app *CrossbarApp
}

// NewCrossbarPresetProvider creates a new preset provider for the crossbar module
func NewCrossbarPresetProvider(app *CrossbarApp) *CrossbarPresetProvider {
	return &CrossbarPresetProvider{app: app}
}

// GetModule returns the module identifier
func (cpp *CrossbarPresetProvider) GetModule() presets.Module {
	return presets.ModuleCrossbar
}

// GetPresetKeys returns the list of configuration keys supported
func (cpp *CrossbarPresetProvider) GetPresetKeys() []string {
	return presets.CrossbarPresetKeys
}

// GetCurrentConfig returns the current configuration as a preset config map
func (cpp *CrossbarPresetProvider) GetCurrentConfig() map[string]interface{} {
	config := make(map[string]interface{})

	// Array size
	if cpp.app.config != nil {
		config["array_size"] = cpp.app.config.Rows
		config["noise_level"] = cpp.app.config.NoiseLevel
		config["adc_bits"] = cpp.app.config.ADCBits
		config["dac_bits"] = cpp.app.config.DACBits
	}

	// Temperature
	config["temperature"] = cpp.app.currentTemperatureK()

	// Architecture
	config["architecture"] = cpp.getArchitectureName()

	// Colormap
	config["colormap"] = cpp.app.condColormap

	// Auto demo state
	cpp.app.stateMu.RLock()
	config["auto_demo"] = cpp.app.autoDemo
	cpp.app.stateMu.RUnlock()

	return config
}

// ApplyPreset applies a preset configuration
func (cpp *CrossbarPresetProvider) ApplyPreset(preset *presets.Preset) error {
	// Array size
	if size, ok := preset.GetInt("array_size"); ok {
		noiseLevel := cpp.app.config.NoiseLevel
		adcBits := cpp.app.config.ADCBits

		if nl, ok := preset.GetFloat("noise_level"); ok {
			noiseLevel = nl
		}
		if bits, ok := preset.GetInt("adc_bits"); ok {
			adcBits = bits
		}

		cpp.app.recreateArray(size, noiseLevel, adcBits)

		if cpp.app.arraySizeSlider != nil {
			cpp.app.arraySizeSlider.SetValue(float64(size))
		}
	} else {
		// Apply noise level separately if array size not changed
		if nl, ok := preset.GetFloat("noise_level"); ok {
			cpp.app.config.NoiseLevel = nl
			if cpp.app.noiseSlider != nil {
				cpp.app.noiseSlider.SetValue(nl * 100)
			}
		}

		if bits, ok := preset.GetInt("adc_bits"); ok {
			cpp.app.config.ADCBits = bits
			if cpp.app.adcBitsSlider != nil {
				cpp.app.adcBitsSlider.SetValue(float64(bits))
			}
		}
	}

	// DAC bits
	if bits, ok := preset.GetInt("dac_bits"); ok {
		cpp.app.config.DACBits = bits
	}

	// Temperature
	if temp, ok := preset.GetFloat("temperature"); ok {
		cpp.app.setTemperatureK(temp)
		if cpp.app.temperatureSlider != nil {
			cpp.app.temperatureSlider.SetValue(temp)
		}
	}

	// Architecture
	if arch, ok := preset.GetString("architecture"); ok {
		cpp.setArchitecture(arch)
	}

	// Colormap
	if cmap, ok := preset.GetString("colormap"); ok {
		cpp.app.condColormap = cmap
		if cpp.app.colormapSelect != nil {
			cpp.app.colormapSelect.SetSelected(cmap)
		}
		if cpp.app.conductanceHeatmap != nil {
			cpp.app.conductanceHeatmap.SetColormap(cmap)
		}
		if cpp.app.condLegend != nil {
			cpp.app.condLegend.SetColormap(cmap)
		}
	}

	// Auto demo
	if autoDemo, ok := preset.GetBool("auto_demo"); ok {
		if autoDemo {
			cpp.app.startAutoDemoLoop()
		} else {
			cpp.app.stopAutoDemoLoop()
		}
	}

	// Refresh displays
	cpp.app.runEnhancedMVMInstant()

	return nil
}

// Helper methods

func (cpp *CrossbarPresetProvider) getArchitectureName() string {
	switch cpp.app.architecture {
	case sharedwidgets.Architecture0T1R:
		return "passive"
	case sharedwidgets.Architecture1T1R:
		return "1T1R"
	case sharedwidgets.Architecture2T1R:
		return "2T1R"
	default:
		return "passive"
	}
}

func (cpp *CrossbarPresetProvider) setArchitecture(name string) {
	var arch string
	switch name {
	case "passive":
		arch = sharedwidgets.Architecture0T1R
	case "1T1R":
		arch = sharedwidgets.Architecture1T1R
	case "2T1R":
		arch = sharedwidgets.Architecture2T1R
	default:
		return
	}

	cpp.app.architecture = arch
	// Note: The architecture toggle widget state is managed by the widget itself
	// Trigger a refresh to apply the new architecture
	cpp.app.runEnhancedMVMInstant()
}
