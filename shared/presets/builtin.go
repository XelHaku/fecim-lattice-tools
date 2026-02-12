package presets

import "time"

// loadBuiltInPresets loads all the built-in presets into the manager
func (m *Manager) loadBuiltInPresets() {
	builtins := getAllBuiltInPresets()
	for _, p := range builtins {
		m.presets[p.Metadata.ID] = p
	}
}

// getAllBuiltInPresets returns all built-in presets
func getAllBuiltInPresets() []*Preset {
	var all []*Preset

	// Global presets
	all = append(all, getGlobalPresets()...)

	// Module-specific presets
	all = append(all, getHysteresisPresets()...)
	all = append(all, getCrossbarPresets()...)
	all = append(all, getMNISTPresets()...)
	all = append(all, getCircuitsPresets()...)
	all = append(all, getComparisonPresets()...)
	all = append(all, getEDAPresets()...)

	return all
}

// builtInPreset is a helper to create built-in presets with consistent metadata
func builtInPreset(id, name, description string, module Module, category Category, tags []string, config map[string]interface{}) *Preset {
	now := time.Date(2025, 2, 7, 0, 0, 0, 0, time.UTC)
	return &Preset{
		Metadata: Metadata{
			ID:          id,
			Name:        name,
			Description: description,
			Category:    category,
			Module:      module,
			Author:      "FeCIM Team",
			Version:     "1.0.0",
			CreatedAt:   now,
			UpdatedAt:   now,
			Tags:        tags,
			BuiltIn:     true,
		},
		Config: config,
	}
}

// =============================================================================
// Global Presets
// =============================================================================

func getGlobalPresets() []*Preset {
	return []*Preset{
		builtInPreset(
			"global-quick-demo",
			"Quick Demo",
			"Fast settings for quick demonstrations with good visual feedback",
			ModuleGlobal,
			CategoryDemo,
			[]string{"demo", "fast", "presentation"},
			map[string]interface{}{
				"animation_speed": 1.5,
				"auto_demo":       true,
				"show_tooltips":   true,
			},
		),
		builtInPreset(
			"global-educational",
			"Educational Mode",
			"Slower animations with detailed explanations for learning",
			ModuleGlobal,
			CategoryEducational,
			[]string{"learning", "slow", "detailed"},
			map[string]interface{}{
				"animation_speed":   0.5,
				"auto_demo":         true,
				"show_tooltips":     true,
				"show_explanations": true,
				"pause_on_step":     true,
			},
		),
		builtInPreset(
			"global-research",
			"Research Mode",
			"High precision settings for accurate analysis and data collection",
			ModuleGlobal,
			CategoryResearch,
			[]string{"research", "precision", "data"},
			map[string]interface{}{
				"animation_speed":  1.0,
				"auto_demo":        false,
				"high_precision":   true,
				"enable_data_log":  true,
				"export_on_change": false,
			},
		),
	}
}

// =============================================================================
// Hysteresis Module Presets
// =============================================================================

func getHysteresisPresets() []*Preset {
	return []*Preset{
		// Educational presets
		builtInPreset(
			"hysteresis-basic-loop",
			"Basic P-E Loop",
			"Classic hysteresis loop demonstration with sine wave excitation",
			ModuleHysteresis,
			CategoryEducational,
			[]string{"learning", "hysteresis", "P-E loop"},
			map[string]interface{}{
				"waveform":       "Sine Wave",
				"frequency":      0.5,
				"amplitude":      1.5,
				"material":       "HZO (optimized)",
				"physics_engine": "preisach",
				"num_levels":     30,
				"time_scale":     1.0,
			},
		),
		builtInPreset(
			"hysteresis-switching",
			"Ferroelectric Switching",
			"Demonstrates polarization switching dynamics",
			ModuleHysteresis,
			CategoryEducational,
			[]string{"learning", "switching", "dynamics"},
			map[string]interface{}{
				"waveform":       "ISPP (Write/Read)",
				"frequency":      0.3,
				"material":       "HZO (optimized)",
				"physics_engine": "preisach",
				"num_levels":     30,
				"target_level":   15,
			},
		),
		builtInPreset(
			"hysteresis-lk-dynamics",
			"L-K Dynamics",
			"Landau-Khalatnikov time-domain switching analysis",
			ModuleHysteresis,
			CategoryResearch,
			[]string{"research", "L-K", "dynamics", "time-domain"},
			map[string]interface{}{
				"waveform":       "Time-Resolved Switching",
				"frequency":      1.0,
				"material":       "HZO (optimized)",
				"physics_engine": "lk",
				"time_scale":     1.0,
			},
		),

		// Research presets
		builtInPreset(
			"hysteresis-material-comparison",
			"Material Comparison",
			"Compare HZO vs PZT ferroelectric behavior",
			ModuleHysteresis,
			CategoryResearch,
			[]string{"research", "materials", "comparison"},
			map[string]interface{}{
				"waveform":       "Sine Wave",
				"frequency":      0.5,
				"amplitude":      1.5,
				"material":       "HZO (optimized)",
				"physics_engine": "preisach",
				"num_levels":     30,
			},
		),
		builtInPreset(
			"hysteresis-temperature-study",
			"Temperature Effects",
			"Study temperature-dependent hysteresis behavior",
			ModuleHysteresis,
			CategoryResearch,
			[]string{"research", "temperature", "thermal"},
			map[string]interface{}{
				"waveform":       "Sine Wave",
				"frequency":      0.5,
				"material":       "HZO (optimized)",
				"physics_engine": "preisach",
				"temperature":    300.0,
				"num_levels":     30,
			},
		),

		// Demo presets
		builtInPreset(
			"hysteresis-investor-demo",
			"Investor Demo",
			"Impressive visual demo for presentations",
			ModuleHysteresis,
			CategoryDemo,
			[]string{"demo", "presentation", "investor"},
			map[string]interface{}{
				"waveform":       "Sine Wave",
				"frequency":      0.3,
				"amplitude":      1.5,
				"material":       "HZO (optimized)",
				"physics_engine": "preisach",
				"num_levels":     30,
				"time_scale":     0.8,
			},
		),
	}
}

// =============================================================================
// Crossbar Module Presets
// =============================================================================

func getCrossbarPresets() []*Preset {
	return []*Preset{
		// Educational presets
		builtInPreset(
			"crossbar-basic-mvm",
			"Basic MVM Operation",
			"Simple matrix-vector multiplication demonstration",
			ModuleCrossbar,
			CategoryEducational,
			[]string{"learning", "MVM", "crossbar"},
			map[string]interface{}{
				"array_size":   32,
				"noise_level":  0.02,
				"adc_bits":     6,
				"temperature":  300.0,
				"architecture": "passive",
				"colormap":     "fecim",
			},
		),
		builtInPreset(
			"crossbar-ir-drop",
			"IR Drop Analysis",
			"Visualize voltage drop effects in larger arrays",
			ModuleCrossbar,
			CategoryEducational,
			[]string{"learning", "IR drop", "non-idealities"},
			map[string]interface{}{
				"array_size":   64,
				"noise_level":  0.05,
				"adc_bits":     6,
				"temperature":  300.0,
				"architecture": "passive",
				"show_ir_drop": true,
			},
		),
		builtInPreset(
			"crossbar-sneak-paths",
			"Sneak Path Effects",
			"Demonstrate sneak current issues in passive arrays",
			ModuleCrossbar,
			CategoryEducational,
			[]string{"learning", "sneak paths", "non-idealities"},
			map[string]interface{}{
				"array_size":      32,
				"noise_level":     0.02,
				"architecture":    "passive",
				"show_sneak_path": true,
			},
		),

		// Research presets
		builtInPreset(
			"crossbar-large-array",
			"Large Array Analysis",
			"128×128 array for scalability research",
			ModuleCrossbar,
			CategoryResearch,
			[]string{"research", "scalability", "large"},
			map[string]interface{}{
				"array_size":   128,
				"noise_level":  0.03,
				"adc_bits":     8,
				"temperature":  300.0,
				"architecture": "1T1R",
			},
		),
		builtInPreset(
			"crossbar-high-precision",
			"High Precision Mode",
			"Maximum ADC resolution for accuracy studies",
			ModuleCrossbar,
			CategoryResearch,
			[]string{"research", "precision", "accuracy"},
			map[string]interface{}{
				"array_size":   64,
				"noise_level":  0.01,
				"adc_bits":     10,
				"temperature":  300.0,
				"architecture": "1T1R",
			},
		),
		builtInPreset(
			"crossbar-temperature-sweep",
			"Temperature Study",
			"Configuration for temperature-dependent analysis",
			ModuleCrossbar,
			CategoryResearch,
			[]string{"research", "temperature", "thermal"},
			map[string]interface{}{
				"array_size":   64,
				"noise_level":  0.02,
				"adc_bits":     6,
				"temperature":  350.0,
				"architecture": "passive",
			},
		),

		// Demo presets
		builtInPreset(
			"crossbar-visual-demo",
			"Visual Demo",
			"Colorful animated demonstration for presentations",
			ModuleCrossbar,
			CategoryDemo,
			[]string{"demo", "visual", "presentation"},
			map[string]interface{}{
				"array_size":   48,
				"noise_level":  0.02,
				"adc_bits":     6,
				"temperature":  300.0,
				"architecture": "passive",
				"colormap":     "plasma",
				"auto_demo":    true,
			},
		),
	}
}

// =============================================================================
// MNIST Module Presets
// =============================================================================

func getMNISTPresets() []*Preset {
	return []*Preset{
		// Educational presets
		builtInPreset(
			"mnist-fp-baseline",
			"FP32 Baseline",
			"Standard floating-point inference for comparison",
			ModuleMNIST,
			CategoryEducational,
			[]string{"learning", "baseline", "FP32"},
			map[string]interface{}{
				"mode":          "fp",
				"quantization":  32,
				"noise_enabled": false,
				"show_weights":  true,
			},
		),
		builtInPreset(
			"mnist-cim-intro",
			"CIM Introduction",
			"Introduction to compute-in-memory inference",
			ModuleMNIST,
			CategoryEducational,
			[]string{"learning", "CIM", "introduction"},
			map[string]interface{}{
				"mode":          "cim",
				"quantization":  8,
				"noise_enabled": false,
				"show_weights":  true,
				"show_energy":   true,
			},
		),

		// Research presets
		builtInPreset(
			"mnist-noise-analysis",
			"Noise Impact Analysis",
			"Study noise effects on inference accuracy",
			ModuleMNIST,
			CategoryResearch,
			[]string{"research", "noise", "accuracy"},
			map[string]interface{}{
				"mode":          "cim",
				"quantization":  8,
				"noise_enabled": true,
				"noise_level":   0.05,
				"show_weights":  true,
			},
		),
		builtInPreset(
			"mnist-quantization-study",
			"Quantization Study",
			"Compare different quantization levels",
			ModuleMNIST,
			CategoryResearch,
			[]string{"research", "quantization", "precision"},
			map[string]interface{}{
				"mode":            "cim",
				"quantization":    4,
				"noise_enabled":   false,
				"show_comparison": true,
			},
		),

		// Demo presets
		builtInPreset(
			"mnist-live-demo",
			"Live Demo",
			"Interactive digit recognition demonstration",
			ModuleMNIST,
			CategoryDemo,
			[]string{"demo", "interactive", "recognition"},
			map[string]interface{}{
				"mode":             "cim",
				"quantization":     8,
				"noise_enabled":    false,
				"auto_recognize":   true,
				"show_activations": true,
			},
		),

		// Benchmark presets
		builtInPreset(
			"mnist-benchmark-accuracy",
			"Accuracy Benchmark",
			"Full test set accuracy measurement",
			ModuleMNIST,
			CategoryBenchmark,
			[]string{"benchmark", "accuracy", "full-test"},
			map[string]interface{}{
				"mode":          "cim",
				"quantization":  8,
				"noise_enabled": true,
				"noise_level":   0.02,
				"batch_size":    100,
			},
		),
	}
}

// =============================================================================
// Circuits Module Presets
// =============================================================================

func getCircuitsPresets() []*Preset {
	return []*Preset{
		// Educational presets
		builtInPreset(
			"circuits-basic-1t1r",
			"Basic 1T1R Cell",
			"Understanding the 1T1R memory cell structure",
			ModuleCircuits,
			CategoryEducational,
			[]string{"learning", "1T1R", "cell"},
			map[string]interface{}{
				"cell_type":      "1T1R",
				"show_voltages":  true,
				"show_currents":  true,
				"animation_mode": "step",
			},
		),
		builtInPreset(
			"circuits-read-operation",
			"Read Operation",
			"Visualize memory cell read operation",
			ModuleCircuits,
			CategoryEducational,
			[]string{"learning", "read", "operation"},
			map[string]interface{}{
				"operation":      "read",
				"show_voltages":  true,
				"show_timing":    true,
				"animation_mode": "continuous",
			},
		),
		builtInPreset(
			"circuits-write-operation",
			"Write Operation",
			"Visualize memory cell write operation",
			ModuleCircuits,
			CategoryEducational,
			[]string{"learning", "write", "operation"},
			map[string]interface{}{
				"operation":      "write",
				"show_voltages":  true,
				"show_timing":    true,
				"animation_mode": "continuous",
			},
		),

		// Research presets
		builtInPreset(
			"circuits-timing-analysis",
			"Timing Analysis",
			"Detailed timing diagram analysis",
			ModuleCircuits,
			CategoryResearch,
			[]string{"research", "timing", "analysis"},
			map[string]interface{}{
				"show_timing":     true,
				"timing_mode":     "detailed",
				"show_margins":    true,
				"export_waveform": true,
			},
		),
		builtInPreset(
			"circuits-power-analysis",
			"Power Analysis",
			"Power consumption during operations",
			ModuleCircuits,
			CategoryResearch,
			[]string{"research", "power", "energy"},
			map[string]interface{}{
				"show_power":       true,
				"show_energy":      true,
				"accumulate_power": true,
			},
		),

		// Demo presets
		builtInPreset(
			"circuits-animated-demo",
			"Animated Demo",
			"Full animated circuit demonstration",
			ModuleCircuits,
			CategoryDemo,
			[]string{"demo", "animated", "visual"},
			map[string]interface{}{
				"animation_mode":  "continuous",
				"animation_speed": 0.5,
				"auto_cycle":      true,
				"show_all":        true,
			},
		),
	}
}

// =============================================================================
// Comparison Module Presets
// =============================================================================

func getComparisonPresets() []*Preset {
	return []*Preset{
		// Educational presets
		builtInPreset(
			"comparison-fecim-vs-sram",
			"FeCIM vs SRAM",
			"Compare FeCIM and traditional SRAM technology",
			ModuleComparison,
			CategoryEducational,
			[]string{"learning", "comparison", "SRAM"},
			map[string]interface{}{
				"comparison_mode": "fecim-sram",
				"show_energy":     true,
				"show_area":       true,
				"show_speed":      true,
			},
		),
		builtInPreset(
			"comparison-energy-focus",
			"Energy Advantage",
			"Focus on energy efficiency comparison",
			ModuleComparison,
			CategoryEducational,
			[]string{"learning", "energy", "efficiency"},
			map[string]interface{}{
				"comparison_mode": "energy",
				"show_joules":     true,
				"show_ops":        true,
				"highlight":       "energy",
			},
		),

		// Demo presets
		builtInPreset(
			"comparison-investor-pitch",
			"Technical Briefing",
			"Key metrics for investor presentations",
			ModuleComparison,
			CategoryDemo,
			[]string{"demo", "investor", "business"},
			map[string]interface{}{
				"comparison_mode": "all",
				"show_market":     true,
				"show_roadmap":    true,
				"highlight":       "advantage",
			},
		),
		builtInPreset(
			"comparison-tech-overview",
			"Technology Overview",
			"Comprehensive technology comparison",
			ModuleComparison,
			CategoryDemo,
			[]string{"demo", "overview", "technology"},
			map[string]interface{}{
				"comparison_mode": "comprehensive",
				"show_all":        true,
				"animation":       true,
			},
		),
	}
}

// =============================================================================
// EDA Module Presets
// =============================================================================

func getEDAPresets() []*Preset {
	return []*Preset{
		// Educational presets
		builtInPreset(
			"eda-intro-flow",
			"EDA Introduction",
			"Overview of the EDA design flow",
			ModuleEDA,
			CategoryEducational,
			[]string{"learning", "EDA", "introduction"},
			map[string]interface{}{
				"show_flow":      true,
				"show_tools":     true,
				"guided_mode":    true,
				"highlight_step": "synthesis",
			},
		),
		builtInPreset(
			"eda-layout-basics",
			"Layout Basics",
			"Introduction to physical layout concepts",
			ModuleEDA,
			CategoryEducational,
			[]string{"learning", "layout", "basics"},
			map[string]interface{}{
				"show_layout": true,
				"show_layers": true,
				"zoom_level":  1.0,
				"guided_mode": true,
			},
		),

		// Research presets
		builtInPreset(
			"eda-sky130-flow",
			"SKY130 PDK Flow",
			"Complete flow with SkyWater 130nm PDK",
			ModuleEDA,
			CategoryResearch,
			[]string{"research", "SKY130", "PDK"},
			map[string]interface{}{
				"pdk":         "sky130",
				"target_freq": 100.0,
				"opt_level":   "high",
			},
		),
		builtInPreset(
			"eda-timing-analysis",
			"EDA Timing Analysis",
			"Static timing analysis configuration",
			ModuleEDA,
			CategoryResearch,
			[]string{"research", "timing", "STA"},
			map[string]interface{}{
				"analysis_mode": "timing",
				"show_slack":    true,
				"show_paths":    true,
				"num_paths":     10,
			},
		),

		// Demo presets
		builtInPreset(
			"eda-visual-flow",
			"Visual Flow Demo",
			"Animated EDA flow demonstration",
			ModuleEDA,
			CategoryDemo,
			[]string{"demo", "visual", "flow"},
			map[string]interface{}{
				"show_flow":    true,
				"animate_flow": true,
				"show_outputs": true,
				"demo_mode":    true,
			},
		),
	}
}
