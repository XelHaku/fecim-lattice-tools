package gui

import (
	"strconv"

	"fecim-lattice-tools/shared/presets"
)

// Ensure HysteresisPresetProvider implements PresetProvider
var _ presets.PresetProvider = (*HysteresisPresetProvider)(nil)

// HysteresisPresetProvider implements preset support for the hysteresis module
type HysteresisPresetProvider struct {
	app *App
}

// NewHysteresisPresetProvider creates a new preset provider for the hysteresis module
func NewHysteresisPresetProvider(app *App) *HysteresisPresetProvider {
	return &HysteresisPresetProvider{app: app}
}

// GetModule returns the module identifier
func (hpp *HysteresisPresetProvider) GetModule() presets.Module {
	return presets.ModuleHysteresis
}

// GetPresetKeys returns the list of configuration keys supported
func (hpp *HysteresisPresetProvider) GetPresetKeys() []string {
	return presets.HysteresisPresetKeys
}

// GetCurrentConfig returns the current configuration as a preset config map
func (hpp *HysteresisPresetProvider) GetCurrentConfig() map[string]interface{} {
	hpp.app.mu.RLock()
	defer hpp.app.mu.RUnlock()

	config := make(map[string]interface{})

	// Waveform
	config["waveform"] = hpp.getWaveformName()

	// Frequency
	config["frequency"] = hpp.app.frequency

	// Amplitude (as ratio of Ec)
	if hpp.app.material != nil && hpp.app.material.Ec > 0 {
		config["amplitude"] = hpp.app.electricField / hpp.app.material.Ec
	}

	// Material
	if hpp.app.material != nil {
		config["material"] = hpp.app.material.Name
	}

	// Physics engine
	config["physics_engine"] = hpp.getPhysicsEngineName()

	// Number of levels
	config["num_levels"] = hpp.app.numLevels

	// Time scale
	config["time_scale"] = hpp.app.timeScale

	// Target level (for ISPP mode)
	config["target_level"] = hpp.app.wrdTargetLevel

	// Temperature (if available)
	config["temperature"] = hpp.app.currentTemperature()

	// Target range fraction
	config["target_range"] = hpp.app.wrdRangeFrac

	// Guard fraction
	config["guard_fraction"] = hpp.app.wrdGuardFrac

	return config
}

// ApplyPreset applies a preset configuration
func (hpp *HysteresisPresetProvider) ApplyPreset(preset *presets.Preset) error {
	hpp.app.mu.Lock()
	defer hpp.app.mu.Unlock()

	// Waveform
	if waveform, ok := preset.GetString("waveform"); ok {
		hpp.setWaveform(waveform)
	}

	// Frequency
	if freq, ok := preset.GetFloat("frequency"); ok {
		hpp.app.frequency = freq
		// Note: frequency is updated through the control callbacks
	}

	// Material
	if material, ok := preset.GetString("material"); ok {
		hpp.setMaterial(material)
	}

	// Physics engine
	if engine, ok := preset.GetString("physics_engine"); ok {
		hpp.setPhysicsEngine(engine)
	}

	// Number of levels
	if levels, ok := preset.GetInt("num_levels"); ok {
		hpp.app.numLevels = levels
		if hpp.app.levelsEntry != nil {
			hpp.app.levelsEntry.SetText(strconv.Itoa(levels))
		}
	}

	// Time scale
	if scale, ok := preset.GetFloat("time_scale"); ok {
		hpp.app.timeScale = scale
	}

	// Target level
	if level, ok := preset.GetInt("target_level"); ok {
		hpp.app.wrdTargetLevel = level
		if hpp.app.levelIndicator != nil {
			hpp.app.levelIndicator.SetTargetLevel(level, false)
		}
	}

	// Target range
	if frac, ok := preset.GetFloat("target_range"); ok {
		hpp.app.wrdRangeFrac = frac
		if hpp.app.wrdRangeSlider != nil {
			hpp.app.wrdRangeSlider.SetValue(frac * 100)
		}
	}

	// Guard fraction
	if frac, ok := preset.GetFloat("guard_fraction"); ok {
		hpp.app.wrdGuardFrac = frac
	}

	// Refresh UI
	hpp.app.refreshUI()

	return nil
}

// Helper methods

func (hpp *HysteresisPresetProvider) getWaveformName() string {
	switch hpp.app.waveform {
	case WaveformManual:
		return "Manual"
	case WaveformSine:
		return "Sine Wave"
	case WaveformTriangle:
		return "Triangle Wave"
	case WaveformWriteReadDemo:
		return "ISPP (Write/Read)"
	case WaveformTimeResolved:
		return "Time-Resolved Switching"
	default:
		return "Unknown"
	}
}

func (hpp *HysteresisPresetProvider) setWaveform(name string) {
	var waveform WaveformType
	switch name {
	case "Manual":
		waveform = WaveformManual
	case "Sine Wave":
		waveform = WaveformSine
	case "Triangle Wave":
		waveform = WaveformTriangle
	case "ISPP (Write/Read)":
		waveform = WaveformWriteReadDemo
	case "Time-Resolved Switching":
		waveform = WaveformTimeResolved
	default:
		return
	}

	hpp.app.waveform = waveform
	if hpp.app.waveformSelect != nil {
		hpp.app.waveformSelect.SetSelected(name)
	}
}

func (hpp *HysteresisPresetProvider) getPhysicsEngineName() string {
	switch hpp.app.physicsEngine {
	case PhysicsPreisach:
		return "preisach"
	case PhysicsLandau:
		return "lk"
	default:
		return "preisach"
	}
}

func (hpp *HysteresisPresetProvider) setPhysicsEngine(name string) {
	switch name {
	case "preisach":
		hpp.app.physicsEngine = PhysicsPreisach
	case "lk":
		hpp.app.physicsEngine = PhysicsLandau
	}
	// Update physics select UI if available
	if hpp.app.physicsSelect != nil {
		if name == "preisach" {
			hpp.app.physicsSelect.SetSelected("Preisach (Quasi-Static)")
		} else if name == "lk" {
			hpp.app.physicsSelect.SetSelected("L-K (Dynamic)")
		}
	}
}

func (hpp *HysteresisPresetProvider) setMaterial(name string) {
	for i, mat := range hpp.app.materials {
		if mat.Name == name {
			hpp.app.matIndex = i
			hpp.app.material = mat
			// Note: preisach model needs to be recreated for new material
			// This is handled by the app's material change logic

			if hpp.app.materialBtn != nil {
				hpp.app.materialBtn.SetText(name)
			}
			break
		}
	}
}

// refreshUI updates all UI elements (called after preset is applied)
func (a *App) refreshUI() {
	// This method should be called after all settings are applied
	// to update any dependent UI elements

	if a.plot != nil {
		a.plot.Refresh()
	}

	if a.levelIndicator != nil {
		a.levelIndicator.Refresh()
	}
}
