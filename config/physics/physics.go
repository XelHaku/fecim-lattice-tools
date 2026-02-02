// Package physics provides access to physical parameters from split YAML configs.
//
// Config files live under config/:
//   - constants.yaml
//   - materials.yaml
//   - crossbar.yaml
//   - training.yaml
//   - energy.yaml
//   - timing.yaml
//   - preisach.yaml
//   - calibration.yaml
//   - simulation.yaml
//   - mnist.yaml
//   - benchmarks.yaml
//
// Legacy monolith config/physics.yaml is still supported.
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

//go:embed defaults/*.yaml
var embeddedConfig embed.FS

// Config holds all physics configuration aggregated from split YAML files.
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
	Benchmarks  map[string]any       `yaml:"benchmarks"`
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
	AnalogStates   int  `yaml:"analog_states,omitempty"`   // Number of discrete states (e.g., 30, 32, 140)
	TRLLevel       int  `yaml:"trl_level,omitempty"`       // Technology Readiness Level (1-9)
	CMOSCompatible bool `yaml:"cmos_compatible,omitempty"` // CMOS fabrication compatible

	// Depolarization (Polycrystalline Analog Behavior)
	DepolarizationFactorVMC float64 `yaml:"depolarization_factor_vm_c,omitempty"` // V*m/C - Creates "slant" for 30-level operation

	// Polarization (C/m²)
	PrCM2 float64 `yaml:"pr_c_m2"`
	PsCM2 float64 `yaml:"ps_c_m2"`

	// Field (V/m)
	EcVM          float64 `yaml:"ec_v_m"`
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
	CurieTempK     float64 `yaml:"curie_temp_k"`
	TempCoeffEc    float64 `yaml:"temp_coeff_ec"`
	TempCoeffPr    float64 `yaml:"temp_coeff_pr"`
	OperatingTempK float64 `yaml:"operating_temp_k,omitempty"` // For cryogenic operation

	// Reliability
	EnduranceCycles float64 `yaml:"endurance_cycles"`
	RetentionTimeS  float64 `yaml:"retention_time_s"`
	ImprintFieldVM  float64 `yaml:"imprint_field_v_m"`

	// FTJ-specific (Ferroelectric Tunnel Junction)
	TERRatio      float64 `yaml:"ter_ratio,omitempty"`       // Tunneling electroresistance ratio
	GmaxGminRatio float64 `yaml:"gmax_gmin_ratio,omitempty"` // Conductance on/off ratio

	// AlScN-specific
	ScFraction float64 `yaml:"sc_fraction,omitempty"` // Scandium fraction in AlScN

	// In2Se3-specific (2D ferroelectric)
	PrInplaneCM2       float64 `yaml:"pr_inplane_c_m2,omitempty"`       // In-plane Pr for 2D materials
	EcThinVM           float64 `yaml:"ec_thin_v_m,omitempty"`           // Ec for ultrathin films
	EcThickVM          float64 `yaml:"ec_thick_v_m,omitempty"`          // Ec for thicker films
	BandgapEV          float64 `yaml:"bandgap_ev,omitempty"`            // Bandgap for semiconductor FE
	MinThicknessM      float64 `yaml:"min_thickness_m,omitempty"`       // Minimum viable thickness
	QuintupleLayerNm   float64 `yaml:"quintuple_layer_nm,omitempty"`    // QL thickness for 2D
	VdWMaterial        bool    `yaml:"vdw_material,omitempty"`          // Van der Waals layered
	Stacking           string  `yaml:"stacking,omitempty"`              // Stacking type (3R, 2H)
	Phase              string  `yaml:"phase,omitempty"`                 // Crystal phase (alpha, beta)
	AlphaToBetaTempK   float64 `yaml:"alpha_to_beta_temp_k,omitempty"`  // Phase transition temp
	LinearityImprovement bool  `yaml:"linearity_improvement,omitempty"` // Better linearity at cryo

	// Synaptic device parameters
	Synaptic MaterialSynaptic `yaml:"synaptic,omitempty"`

	// Synthesis method
	Synthesis MaterialSynthesis `yaml:"synthesis,omitempty"`

	// Landau-Khalatnikov / Thermodynamics
	Thermodynamics MaterialThermodynamics `yaml:"thermodynamics,omitempty"`

	// Electrostriction / Stress coupling
	Coupling MaterialCoupling `yaml:"coupling,omitempty"`

	// Circuit parasitics
	Circuit MaterialCircuit `yaml:"circuit,omitempty"`

	// Nucleation-Limited Switching (Merz law)
	NLS MaterialNLS `yaml:"nls,omitempty"`

	// Conductance transfer (P -> G)
	Conductance MaterialConductance `yaml:"conductance,omitempty"`
}

// MaterialThermodynamics holds Landau-Khalatnikov coefficients.
type MaterialThermodynamics struct {
	BetaLandau   float64 `yaml:"beta_landau"`
	GammaLandau  float64 `yaml:"gamma_landau"`
	RhoViscosity float64 `yaml:"rho_viscosity"`
	CurieConstK  float64 `yaml:"curie_const_k"`
}

// MaterialCoupling holds electrostriction and stress coupling parameters.
type MaterialCoupling struct {
	Q11Electrostriction float64 `yaml:"q11_electrostriction"`
	Q12Electrostriction float64 `yaml:"q12_electrostriction"`
	StressGPa           float64 `yaml:"stress_gpa"`
}

// MaterialCircuit holds circuit parasitic parameters.
type MaterialCircuit struct {
	SeriesResistanceOhm float64 `yaml:"series_resistance_ohm"`
}

// MaterialNLS holds Nucleation-Limited Switching parameters.
type MaterialNLS struct {
	ActivationFieldVM float64 `yaml:"activation_field_v_m"`
	TauInfS           float64 `yaml:"tau_inf_s"`
}

// MaterialConductance holds conductance mapping parameters.
type MaterialConductance struct {
	GminS      float64 `yaml:"gmin_s"`
	GmaxS      float64 `yaml:"gmax_s"`
	OnOffRatio float64 `yaml:"on_off_ratio,omitempty"` // Gmax/Gmin ratio
}

// MaterialSynaptic holds synaptic device parameters for neuromorphic computing.
type MaterialSynaptic struct {
	PotentiationPulses int     `yaml:"potentiation_pulses,omitempty"` // Pulses for LTP
	DepressionPulses   int     `yaml:"depression_pulses,omitempty"`   // Pulses for LTD
	PulseWidthS        float64 `yaml:"pulse_width_s,omitempty"`       // Programming pulse width
	PulseVoltageV      float64 `yaml:"pulse_voltage_v,omitempty"`     // Programming voltage
	NonlinearityLTP    float64 `yaml:"nonlinearity_ltp,omitempty"`    // LTP nonlinearity
	NonlinearityLTD    float64 `yaml:"nonlinearity_ltd,omitempty"`    // LTD nonlinearity
}

// MaterialSynthesis holds synthesis method information.
type MaterialSynthesis struct {
	Method          string   `yaml:"method,omitempty"`           // Synthesis method (FWF, ALD, etc.)
	Scale           string   `yaml:"scale,omitempty"`            // Production scale
	Precursors      []string `yaml:"precursors,omitempty"`       // Precursor materials
	SynthesisTimeS  float64  `yaml:"synthesis_time_s,omitempty"` // Synthesis time
	EnergyReduction float64  `yaml:"energy_reduction,omitempty"` // Energy reduction vs conventional
}

// Crossbar holds crossbar array configuration.
type Crossbar struct {
	DefaultRows int `yaml:"default_rows"`
	DefaultCols int `yaml:"default_cols"`

	QuantizationLevels int `yaml:"quantization_levels"`
	ADCBits            int `yaml:"adc_bits"`
	DACBits            int `yaml:"dac_bits"`

	ConductanceMinS  float64 `yaml:"conductance_min_s"`
	ConductanceMaxS  float64 `yaml:"conductance_max_s"`
	ConductanceRatio float64 `yaml:"conductance_ratio"`

	DeviceVariation float64 `yaml:"device_variation"`
	ReadNoise       float64 `yaml:"read_noise"`
	WriteNoise      float64 `yaml:"write_noise"`

	WordLineResistanceOhm float64 `yaml:"word_line_resistance_ohm"`
	BitLineResistanceOhm  float64 `yaml:"bit_line_resistance_ohm"`

	SneakPathEnabled      bool    `yaml:"sneak_path_enabled"`
	SneakConductanceRatio float64 `yaml:"sneak_conductance_ratio"`
}

// Training holds training configuration.
type Training struct {
	LearningRate float64 `yaml:"learning_rate"`
	WeightDecay  float64 `yaml:"weight_decay"`
	Momentum     float64 `yaml:"momentum"`

	DefaultBatchSize int     `yaml:"default_batch_size"`
	GradientClip     float64 `yaml:"gradient_clip"`

	WeightClipMin    float64 `yaml:"weight_clip_min"`
	WeightClipMax    float64 `yaml:"weight_clip_max"`
	UpdateNoiseSigma float64 `yaml:"update_noise_sigma"`
	AsymmetryRatio   float64 `yaml:"asymmetry_ratio"`

	QuantizeForward  bool `yaml:"quantize_forward"`
	QuantizeBackward bool `yaml:"quantize_backward"`
	StraightThrough  bool `yaml:"straight_through"`
}

// Energy holds energy parameters.
type Energy struct {
	ReadEnergyJ  float64 `yaml:"read_energy_j"`
	WriteEnergyJ float64 `yaml:"write_energy_j"`
	MACEnergyJ   float64 `yaml:"mac_energy_j"`

	NANDWriteEnergyJ  float64 `yaml:"nand_write_energy_j"`
	DRAMAccessEnergyJ float64 `yaml:"dram_access_energy_j"`
	SRAMAccessEnergyJ float64 `yaml:"sram_access_energy_j"`

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

const (
	constantsFile   = "constants.yaml"
	materialsFile   = "materials.yaml"
	crossbarFile    = "crossbar.yaml"
	trainingFile    = "training.yaml"
	energyFile      = "energy.yaml"
	timingFile      = "timing.yaml"
	preisachFile    = "preisach.yaml"
	calibrationFile = "calibration.yaml"
	simulationFile  = "simulation.yaml"
	mnistFile       = "mnist.yaml"
	benchmarksFile  = "benchmarks.yaml"

	legacyConfigFile = "physics.yaml"
)

var (
	configRoots = []string{
		"config",
		"../config",
		"../../config",
	}
	// Split file probe list (materials.yaml is intentionally excluded to avoid
	// treating legacy setups as split configs).
	splitProbeFiles = []string{
		constantsFile,
		crossbarFile,
		trainingFile,
		energyFile,
		timingFile,
		preisachFile,
		calibrationFile,
		simulationFile,
		mnistFile,
		benchmarksFile,
	}
)

type constantsWrapper struct {
	Constants Constants `yaml:"constants"`
}

type materialsWrapper struct {
	Materials map[string]*Material `yaml:"materials"`
}

type crossbarWrapper struct {
	Crossbar Crossbar `yaml:"crossbar"`
}

type trainingWrapper struct {
	Training Training `yaml:"training"`
}

type energyWrapper struct {
	Energy Energy `yaml:"energy"`
}

type timingWrapper struct {
	Timing Timing `yaml:"timing"`
}

type preisachWrapper struct {
	Preisach Preisach `yaml:"preisach"`
}

type calibrationWrapper struct {
	Calibration Calibration `yaml:"calibration"`
}

type simulationWrapper struct {
	Simulation Simulation `yaml:"simulation"`
}

type mnistWrapper struct {
	MNIST MNIST `yaml:"mnist"`
}

type benchmarksWrapper struct {
	Benchmarks map[string]any `yaml:"benchmarks"`
}

// Load loads the physics configuration from split YAML files.
// It first tries to load split configs from config/*.yaml relative to the working
// directory, falls back to legacy config/physics.yaml if present, and finally
// uses embedded defaults.
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
	if root, ok := findSplitConfigRoot(); ok {
		return loadSplitConfig(root)
	}

	legacyData, err := readLegacyConfig()
	if err != nil {
		return nil, err
	}
	if legacyData != nil {
		var cfg Config
		if err := yaml.Unmarshal(legacyData, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", legacyConfigFile, err)
		}
		return &cfg, nil
	}

	// Fall back to embedded split defaults.
	return loadSplitConfig("")
}

func loadSplitConfig(root string) (*Config, error) {
	cfg := &Config{}

	var constants constantsWrapper
	if err := loadSection(root, constantsFile, &constants); err != nil {
		return nil, err
	}
	cfg.Constants = constants.Constants

	var materials materialsWrapper
	if err := loadSection(root, materialsFile, &materials); err != nil {
		return nil, err
	}
	cfg.Materials = materials.Materials

	var crossbar crossbarWrapper
	if err := loadSection(root, crossbarFile, &crossbar); err != nil {
		return nil, err
	}
	cfg.Crossbar = crossbar.Crossbar

	var training trainingWrapper
	if err := loadSection(root, trainingFile, &training); err != nil {
		return nil, err
	}
	cfg.Training = training.Training

	var energy energyWrapper
	if err := loadSection(root, energyFile, &energy); err != nil {
		return nil, err
	}
	cfg.Energy = energy.Energy

	var timing timingWrapper
	if err := loadSection(root, timingFile, &timing); err != nil {
		return nil, err
	}
	cfg.Timing = timing.Timing

	var preisach preisachWrapper
	if err := loadSection(root, preisachFile, &preisach); err != nil {
		return nil, err
	}
	cfg.Preisach = preisach.Preisach

	var calibration calibrationWrapper
	if err := loadSection(root, calibrationFile, &calibration); err != nil {
		return nil, err
	}
	cfg.Calibration = calibration.Calibration

	var simulation simulationWrapper
	if err := loadSection(root, simulationFile, &simulation); err != nil {
		return nil, err
	}
	cfg.Simulation = simulation.Simulation

	var mnist mnistWrapper
	if err := loadSection(root, mnistFile, &mnist); err != nil {
		return nil, err
	}
	cfg.MNIST = mnist.MNIST

	var benchmarks benchmarksWrapper
	if err := loadSection(root, benchmarksFile, &benchmarks); err != nil {
		return nil, err
	}
	cfg.Benchmarks = benchmarks.Benchmarks

	return cfg, nil
}

func loadSection(root, filename string, out any) error {
	data, err := readConfigFile(root, filename)
	if err != nil {
		return err
	}
	if data == nil {
		embeddedPath := filepath.ToSlash(filepath.Join("defaults", filename))
		data, err = embeddedConfig.ReadFile(embeddedPath)
		if err != nil {
			return fmt.Errorf("failed to read embedded %s: %w", filename, err)
		}
	}
	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filename, err)
	}
	return nil
}

func findSplitConfigRoot() (string, bool) {
	for _, root := range configRoots {
		for _, filename := range splitProbeFiles {
			path := filepath.Join(root, filename)
			if fileExists(path) {
				return root, true
			}
		}
	}
	return "", false
}

func readLegacyConfig() ([]byte, error) {
	for _, root := range configRoots {
		path := filepath.Join(root, legacyConfigFile)
		if fileExists(path) {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read %s: %w", path, err)
			}
			return data, nil
		}
	}
	return nil, nil
}

func readConfigFile(root, filename string) ([]byte, error) {
	if root == "" {
		return nil, nil
	}
	path := filepath.Join(root, filename)
	if !fileExists(path) {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return data, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
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

// GetNumLevels returns the number of discrete analog states for this material.
// Returns the material's AnalogStates if set, otherwise falls back to the
// global FeCIMLevels constant (default 30).
func (m *Material) GetNumLevels(cfg *Config) int {
	if m.AnalogStates > 0 {
		return m.AnalogStates
	}
	if cfg != nil && cfg.Constants.FeCIMLevels > 0 {
		return cfg.Constants.FeCIMLevels
	}
	return 30 // Ultimate fallback
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
