package presets

// BaseProvider implements common PresetProvider functionality
type BaseProvider struct {
	module Module
	keys   []string
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(module Module, keys []string) *BaseProvider {
	return &BaseProvider{
		module: module,
		keys:   keys,
	}
}

// GetModule returns the module identifier
func (bp *BaseProvider) GetModule() Module {
	return bp.module
}

// GetPresetKeys returns the list of configuration keys supported
func (bp *BaseProvider) GetPresetKeys() []string {
	return bp.keys
}

// HysteresisPresetKeys defines the configuration keys for the hysteresis module
var HysteresisPresetKeys = []string{
	"waveform",
	"frequency",
	"amplitude",
	"material",
	"physics_engine",
	"num_levels",
	"time_scale",
	"target_level",
	"temperature",
	"target_range",
	"guard_fraction",
}

// CrossbarPresetKeys defines the configuration keys for the crossbar module
var CrossbarPresetKeys = []string{
	"array_size",
	"noise_level",
	"adc_bits",
	"dac_bits",
	"temperature",
	"architecture",
	"colormap",
	"show_ir_drop",
	"show_sneak_path",
	"auto_demo",
	"animation_speed",
}

// MNISTPresetKeys defines the configuration keys for the MNIST module
var MNISTPresetKeys = []string{
	"mode",
	"quantization",
	"noise_enabled",
	"noise_level",
	"show_weights",
	"show_energy",
	"show_activations",
	"show_comparison",
	"auto_recognize",
	"batch_size",
}

// CircuitsPresetKeys defines the configuration keys for the circuits module
var CircuitsPresetKeys = []string{
	"cell_type",
	"operation",
	"show_voltages",
	"show_currents",
	"show_timing",
	"show_power",
	"show_energy",
	"animation_mode",
	"animation_speed",
	"timing_mode",
	"auto_cycle",
}

// ComparisonPresetKeys defines the configuration keys for the comparison module
var ComparisonPresetKeys = []string{
	"comparison_mode",
	"show_energy",
	"show_area",
	"show_speed",
	"show_market",
	"show_roadmap",
	"highlight",
	"animation",
}

// EDAPresetKeys defines the configuration keys for the EDA module
var EDAPresetKeys = []string{
	"pdk",
	"target_freq",
	"opt_level",
	"show_flow",
	"show_tools",
	"show_layout",
	"show_layers",
	"show_timing",
	"show_slack",
	"show_paths",
	"animate_flow",
	"guided_mode",
	"demo_mode",
}

// GlobalPresetKeys defines the configuration keys for global settings
var GlobalPresetKeys = []string{
	"animation_speed",
	"auto_demo",
	"show_tooltips",
	"show_explanations",
	"pause_on_step",
	"high_precision",
	"enable_data_log",
	"export_on_change",
}

// GetModuleKeys returns the preset keys for a given module
func GetModuleKeys(module Module) []string {
	switch module {
	case ModuleHysteresis:
		return HysteresisPresetKeys
	case ModuleCrossbar:
		return CrossbarPresetKeys
	case ModuleMNIST:
		return MNISTPresetKeys
	case ModuleCircuits:
		return CircuitsPresetKeys
	case ModuleComparison:
		return ComparisonPresetKeys
	case ModuleEDA:
		return EDAPresetKeys
	case ModuleGlobal:
		return GlobalPresetKeys
	default:
		return nil
	}
}
