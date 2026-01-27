// Package physics provides access to physical parameters from config/physics.yaml.
//
// Usage:
//
//	cfg, err := physics.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	material := cfg.GetMaterial("fecim_hzo")
//	fmt.Printf("Ec = %.2f MV/cm\n", material.EcVM/1e8)
package physics

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed physics.yaml
var embeddedConfig embed.FS

// Config holds all physics configuration from physics.yaml.
type Config struct {
	Constants   Constants            `yaml:"constants"`
	Materials   map[string]*Material `yaml:"materials"`
	Crossbar    Crossbar             `yaml:"crossbar"`
	Training    Training             `yaml:"training"`
	Energy      Energy               `yaml:"energy"`
	Timing      Timing               `yaml:"timing"`
	Preisach    Preisach             `yaml:"preisach"`
	Calibration Calibration          `yaml:"calibration"`
	Simulation  Simulation           `yaml:"simulation"`
	MNIST       MNIST                `yaml:"mnist"`
}

// Constants holds global physics constants.
type Constants struct {
	FeCIMLevels     int     `yaml:"fecim_levels"`
	BitsPerCell     float64 `yaml:"bits_per_cell"`
	BoltzmannEV     float64 `yaml:"boltzmann_ev"`
	Epsilon0        float64 `yaml:"epsilon_0"`
	RoomTemperature float64 `yaml:"room_temperature"`
}

// Material holds ferroelectric material parameters.
type Material struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Reference   string `yaml:"reference"`

	// Multi-level capability
	AnalogStates int  `yaml:"analog_states,omitempty"` // Number of discrete states (e.g., 30, 32, 140)
	TRLLevel     int  `yaml:"trl_level,omitempty"`     // Technology Readiness Level (1-9)
	CMOSCompatible bool `yaml:"cmos_compatible,omitempty"` // CMOS fabrication compatible

	// Polarization (C/m²)
	PrCM2 float64 `yaml:"pr_c_m2"`
	PsCM2 float64 `yaml:"ps_c_m2"`

	// Field (V/m)
	EcVM         float64 `yaml:"ec_v_m"`
	MemoryWindowV float64 `yaml:"memory_window_v,omitempty"` // Memory window voltage

	// Dielectric
	EpsilonHF   float64 `yaml:"epsilon_hf"`
	EpsilonLF   float64 `yaml:"epsilon_lf"`
	LossTangent float64 `yaml:"loss_tangent"`

	// Geometry
	ThicknessM  float64 `yaml:"thickness_m"`
	AreaM2      float64 `yaml:"area_m2"`
	CellPitchNm float64 `yaml:"cell_pitch_nm,omitempty"`

	// Dynamics
	TauS               float64 `yaml:"tau_s"`
	Tau0S              float64 `yaml:"tau0_s"`
	ActivationEnergyEV float64 `yaml:"activation_energy_ev"`
	KAIExponent        float64 `yaml:"kai_exponent"`

	// Temperature
	CurieTempK       float64 `yaml:"curie_temp_k"`
	TempCoeffEc      float64 `yaml:"temp_coeff_ec"`
	TempCoeffPr      float64 `yaml:"temp_coeff_pr"`
	OperatingTempK   float64 `yaml:"operating_temp_k,omitempty"` // For cryogenic operation

	// Reliability
	EnduranceCycles float64 `yaml:"endurance_cycles"`
	RetentionTimeS  float64 `yaml:"retention_time_s"`
	ImprintFieldVM  float64 `yaml:"imprint_field_v_m"`

	// FTJ-specific (Ferroelectric Tunnel Junction)
	TERRatio       float64 `yaml:"ter_ratio,omitempty"`        // Tunneling electroresistance ratio
	GmaxGminRatio  float64 `yaml:"gmax_gmin_ratio,omitempty"`  // Conductance on/off ratio

	// AlScN-specific
	ScFraction float64 `yaml:"sc_fraction,omitempty"` // Scandium fraction in AlScN
}

// Crossbar holds crossbar array configuration.
type Crossbar struct {
	DefaultRows int `yaml:"default_rows"`
	DefaultCols int `yaml:"default_cols"`

	QuantizationLevels int `yaml:"quantization_levels"`
	ADCBits            int `yaml:"adc_bits"`
	DACBits            int `yaml:"dac_bits"`

	ConductanceMinS   float64 `yaml:"conductance_min_s"`
	ConductanceMaxS   float64 `yaml:"conductance_max_s"`
	ConductanceRatio  float64 `yaml:"conductance_ratio"`

	DeviceVariation float64 `yaml:"device_variation"`
	ReadNoise       float64 `yaml:"read_noise"`
	WriteNoise      float64 `yaml:"write_noise"`

	WordLineResistanceOhm float64 `yaml:"word_line_resistance_ohm"`
	BitLineResistanceOhm  float64 `yaml:"bit_line_resistance_ohm"`

	SneakPathEnabled       bool    `yaml:"sneak_path_enabled"`
	SneakConductanceRatio  float64 `yaml:"sneak_conductance_ratio"`
}

// Training holds training configuration.
type Training struct {
	LearningRate float64 `yaml:"learning_rate"`
	WeightDecay  float64 `yaml:"weight_decay"`
	Momentum     float64 `yaml:"momentum"`

	DefaultBatchSize int     `yaml:"default_batch_size"`
	GradientClip     float64 `yaml:"gradient_clip"`

	WeightClipMin     float64 `yaml:"weight_clip_min"`
	WeightClipMax     float64 `yaml:"weight_clip_max"`
	UpdateNoiseSigma  float64 `yaml:"update_noise_sigma"`
	AsymmetryRatio    float64 `yaml:"asymmetry_ratio"`

	QuantizeForward  bool `yaml:"quantize_forward"`
	QuantizeBackward bool `yaml:"quantize_backward"`
	StraightThrough  bool `yaml:"straight_through"`
}

// Energy holds energy parameters.
type Energy struct {
	ReadEnergyJ  float64 `yaml:"read_energy_j"`
	WriteEnergyJ float64 `yaml:"write_energy_j"`
	MACEnergyJ   float64 `yaml:"mac_energy_j"`

	NANDWriteEnergyJ   float64 `yaml:"nand_write_energy_j"`
	DRAMAccessEnergyJ  float64 `yaml:"dram_access_energy_j"`
	SRAMAccessEnergyJ  float64 `yaml:"sram_access_energy_j"`

	StandbyPowerWPerCell float64 `yaml:"standby_power_w_per_cell"`
}

// Timing holds timing parameters.
type Timing struct {
	ReadLatencyS  float64 `yaml:"read_latency_s"`
	WriteLatencyS float64 `yaml:"write_latency_s"`
	MACLatencyS   float64 `yaml:"mac_latency_s"`

	PulseWidthS      float64 `yaml:"pulse_width_s"`
	VerifyDelayS     float64 `yaml:"verify_delay_s"`
	MaxProgramCycles int     `yaml:"max_program_cycles"`
}

// Preisach holds Preisach model parameters.
type Preisach struct {
	GridSize int `yaml:"grid_size"`

	AlphaSigmaRatio float64 `yaml:"alpha_sigma_ratio"`
	BetaSigmaRatio  float64 `yaml:"beta_sigma_ratio"`
	Correlation     float64 `yaml:"correlation"`

	FatigueRate   float64 `yaml:"fatigue_rate"`
	WakeupCycles  int     `yaml:"wakeup_cycles"`
	InitialWakeup float64 `yaml:"initial_wakeup"`
}

// Calibration holds calibration parameters.
type Calibration struct {
	Iterations    int     `yaml:"iterations"`
	FieldMinRatio float64 `yaml:"field_min_ratio"`
	FieldMaxRatio float64 `yaml:"field_max_ratio"`

	AdjustmentRate float64 `yaml:"adjustment_rate"`
	LevelTolerance int     `yaml:"level_tolerance"`
}

// Simulation holds simulation parameters.
type Simulation struct {
	FrameRateHz int     `yaml:"frame_rate_hz"`
	DtS         float64 `yaml:"dt_s"`

	MaxHistoryPoints int `yaml:"max_history_points"`

	DefaultFrequencyHz    float64 `yaml:"default_frequency_hz"`
	DefaultAmplitudeRatio float64 `yaml:"default_amplitude_ratio"`
}

// MNIST holds MNIST demo parameters.
type MNIST struct {
	InputSize   int   `yaml:"input_size"`
	HiddenSizes []int `yaml:"hidden_sizes"`
	OutputSize  int   `yaml:"output_size"`

	Epochs       int     `yaml:"epochs"`
	BatchSize    int     `yaml:"batch_size"`
	LearningRate float64 `yaml:"learning_rate"`

	BaselineAccuracy    float64 `yaml:"baseline_accuracy"`
	TourClaimedAccuracy float64 `yaml:"tour_claimed_accuracy"`
}

// Global config singleton
var (
	globalConfig *Config
	configOnce   sync.Once
	configErr    error
)

// Load loads the physics configuration from physics.yaml.
// It first tries to load from config/physics.yaml relative to the working directory,
// then falls back to the embedded config.
func Load() (*Config, error) {
	configOnce.Do(func() {
		globalConfig, configErr = loadConfig()
	})
	return globalConfig, configErr
}

// MustLoad loads the config and panics on error.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load physics config: %v", err))
	}
	return cfg
}

// Reload forces a reload of the configuration (useful for testing).
func Reload() (*Config, error) {
	configOnce = sync.Once{}
	return Load()
}

func loadConfig() (*Config, error) {
	var data []byte
	var err error

	// Try to load from filesystem first (allows customization)
	paths := []string{
		"config/physics.yaml",
		"../config/physics.yaml",
		"../../config/physics.yaml",
	}

	for _, path := range paths {
		if _, statErr := os.Stat(path); statErr == nil {
			data, err = os.ReadFile(path)
			if err == nil {
				break
			}
		}
	}

	// Fall back to embedded config
	if data == nil {
		data, err = embeddedConfig.ReadFile("physics.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded config: %w", err)
		}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse physics.yaml: %w", err)
	}

	return &cfg, nil
}

// GetMaterial returns a material by name.
// Valid names (all CMOS compatible):
//   - "default_hzo"           - Baseline Si-doped HZO (30 states)
//   - "fecim_hzo"             - FeCIM demonstrated values (30 states)
//   - "fecim_hzo_target"      - FeCIM aspirational targets (30 states)
//   - "literature_superlattice" - Best academic results (30+ states)
//   - "cryogenic_hzo"         - HZO at 4K for quantum computing
//   - "hzo_standard_32"       - Oh IEEE EDL 2017 (32 states)
//   - "hzo_ftj_140"           - Song Adv.Science 2024 FTJ (140 states)
//   - "alscn"                 - AlScN high-Pr material (8-16 states)
func (c *Config) GetMaterial(name string) *Material {
	return c.Materials[name]
}

// DefaultMaterial returns the default HZO material.
func (c *Config) DefaultMaterial() *Material {
	return c.Materials["default_hzo"]
}

// FeCIMMaterial returns the FeCIM HZO material (demonstrated values).
func (c *Config) FeCIMMaterial() *Material {
	return c.Materials["fecim_hzo"]
}

// MaterialNames returns a list of all available material names.
func (c *Config) MaterialNames() []string {
	names := make([]string, 0, len(c.Materials))
	for name := range c.Materials {
		names = append(names, name)
	}
	return names
}

// --- Convenience methods on Material ---

// CoerciveVoltage returns Ec * thickness (V).
func (m *Material) CoerciveVoltage() float64 {
	return m.EcVM * m.ThicknessM
}

// EcMVcm returns coercive field in MV/cm.
func (m *Material) EcMVcm() float64 {
	return m.EcVM / 1e8
}

// PrMicroCcm2 returns Pr in µC/cm².
func (m *Material) PrMicroCcm2() float64 {
	return m.PrCM2 * 100 // C/m² to µC/cm² = x100
}

// PsMicroCcm2 returns Ps in µC/cm².
func (m *Material) PsMicroCcm2() float64 {
	return m.PsCM2 * 100
}

// ThicknessNm returns thickness in nanometers.
func (m *Material) ThicknessNm() float64 {
	return m.ThicknessM * 1e9
}

// SaveToFile saves the current config to a YAML file.
func (c *Config) SaveToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
