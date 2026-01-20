// nc_2dferro.go - Negative Capacitance FET and 2D Ferroelectric Materials for CIM
//
// This module implements:
// 1. NC-FET device physics with sub-60mV/decade subthreshold swing
// 2. Transient negative capacitance stabilization
// 3. 2D ferroelectric materials (CuInP2S6, α-In2Se3, CuCrP2S6)
// 4. Van der Waals heterostructure synaptic devices
// 5. 1T1M architecture for neuromorphic crossbar arrays
//
// References:
// - Terra Quantum NC-FET: <30 mV/decade, 0.5 fJ/operation (2025)
// - HZO MoS2 NC-FET: 6.07 mV/decade minimum SS, 28× voltage gain
// - CuCrP2S6 memristor: 20,000 cycles, 120°C stability
// - 1T1M architecture: all-vdW neuromorphic arrays

package layers

import (
	"math"
	"math/rand"
)

// =============================================================================
// NEGATIVE CAPACITANCE FET MODELS
// =============================================================================

// NCFETConfig configures negative capacitance field-effect transistor
type NCFETConfig struct {
	// Ferroelectric layer properties
	FerroelectricThickness float64 // nm (typical: 5-20)
	FerroelectricMaterial  string  // "HZO", "AlHfO2", "PZT"
	Permittivity           float64 // relative permittivity
	CoerciveField          float64 // MV/cm
	RemanentPolarization   float64 // μC/cm²

	// Dielectric buffer layer
	DielectricThickness float64 // nm (typical: 1-5)
	DielectricMaterial  string  // "Al2O3", "SiO2", "HfO2"
	DielectricK         float64 // relative permittivity

	// Transistor parameters
	ChannelLength  float64 // nm
	ChannelWidth   float64 // nm
	SupplyVoltage  float64 // V (ultra-low: 0.3-0.5V)
	ThresholdVolt  float64 // V
	MobilityFactor float64 // cm²/V·s

	// NC stabilization
	CapacitanceMatching bool    // enable Cfe/Cox matching
	StabilizationFactor float64 // Cfe/Cox ratio target
}

// DefaultNCFETConfig returns optimized NC-FET configuration
func DefaultNCFETConfig() *NCFETConfig {
	return &NCFETConfig{
		FerroelectricThickness: 12.0,   // nm (HZO optimal)
		FerroelectricMaterial:  "HZO",  // Hf0.5Zr0.5O2
		Permittivity:           30.0,   // HZO relative permittivity
		CoerciveField:          1.0,    // MV/cm
		RemanentPolarization:   15.0,   // μC/cm²
		DielectricThickness:    2.0,    // nm
		DielectricMaterial:     "Al2O3",
		DielectricK:            9.0,    // Al2O3
		ChannelLength:          20.0,   // nm
		ChannelWidth:           100.0,  // nm
		SupplyVoltage:          0.3,    // V ultra-low
		ThresholdVolt:          0.15,   // V
		MobilityFactor:         200.0,  // cm²/V·s
		CapacitanceMatching:    true,
		StabilizationFactor:    1.5,    // Cfe/Cox optimal
	}
}

// NCFET represents a negative capacitance field-effect transistor
type NCFET struct {
	Config *NCFETConfig

	// Computed capacitances
	FerroCapacitance float64 // fF/μm²
	OxideCapacitance float64 // fF/μm²
	TotalCapacitance float64 // fF/μm²

	// NC effect parameters
	VoltageGain       float64 // internal voltage amplification
	SubthresholdSwing float64 // mV/decade (target: <60)
	BodyFactor        float64 // body effect coefficient

	// Performance metrics
	OnCurrent    float64 // μA/μm
	OffCurrent   float64 // nA/μm
	OnOffRatio   float64 // typically 10⁵-10⁶
	EnergyPerOp  float64 // fJ/operation
	SwitchSpeed  float64 // ns
}

// NewNCFET creates a new negative capacitance FET
func NewNCFET(config *NCFETConfig) *NCFET {
	if config == nil {
		config = DefaultNCFETConfig()
	}

	nc := &NCFET{
		Config: config,
	}

	nc.calculateCapacitances()
	nc.calculateNCEffect()
	nc.calculatePerformance()

	return nc
}

// calculateCapacitances computes gate stack capacitances
func (nc *NCFET) calculateCapacitances() {
	eps0 := 8.854e-3 // fF/μm (vacuum permittivity)

	// Ferroelectric capacitance: C = ε₀ε_r A / d
	// Convert thickness from nm to μm
	tfe := nc.Config.FerroelectricThickness * 1e-3 // μm
	nc.FerroCapacitance = eps0 * nc.Config.Permittivity / tfe

	// Oxide capacitance
	tox := nc.Config.DielectricThickness * 1e-3 // μm
	nc.OxideCapacitance = eps0 * nc.Config.DielectricK / tox

	// For NC effect, ferroelectric acts as negative capacitance
	// Total: 1/Ctot = 1/Cfe + 1/Cox (with Cfe negative)
	// When |Cfe| > Cox, we get voltage amplification

	if nc.Config.CapacitanceMatching {
		// Stabilized NC: Cfe matched to provide gain without instability
		// Effective capacitance enhanced by NC factor
		ncFactor := nc.Config.StabilizationFactor
		nc.TotalCapacitance = nc.OxideCapacitance * ncFactor
	} else {
		// Simple series combination (Cfe treated as magnitude)
		nc.TotalCapacitance = (nc.FerroCapacitance * nc.OxideCapacitance) /
			(nc.FerroCapacitance + nc.OxideCapacitance)
	}
}

// calculateNCEffect computes negative capacitance amplification
func (nc *NCFET) calculateNCEffect() {
	// Voltage gain from NC effect: Av = Cox / (Cox - |Cfe|)
	// For stabilized NC, gain is controlled by matching

	cfeAbs := nc.FerroCapacitance
	cox := nc.OxideCapacitance

	// NC voltage gain (when Cfe approaches Cox)
	if cfeAbs < cox {
		nc.VoltageGain = cox / (cox - cfeAbs*0.9) // 90% matching
	} else {
		// Full amplification regime
		nc.VoltageGain = nc.Config.StabilizationFactor * 10.0
	}

	// Cap voltage gain to realistic values (measured: up to 28×)
	if nc.VoltageGain > 28.0 {
		nc.VoltageGain = 28.0
	}

	// Subthreshold swing: SS = (kT/q) * ln(10) * (1 + Cd/Cg)
	// With NC: SS = 60mV/decade × (1 - |Cfe|/Cox)
	// For |Cfe| → Cox: SS → 0

	kTq := 25.85 // mV at room temperature
	ln10 := math.Log(10)
	thermalLimit := kTq * ln10 // 59.5 mV/decade

	// NC reduction factor
	ncReduction := 1.0 - (cfeAbs / cox) * 0.95

	if ncReduction < 0.1 {
		ncReduction = 0.1 // minimum 6 mV/decade (as demonstrated)
	}

	nc.SubthresholdSwing = thermalLimit * ncReduction

	// Body factor modification by NC
	nc.BodyFactor = 1.0 / nc.VoltageGain
}

// calculatePerformance computes transistor performance metrics
func (nc *NCFET) calculatePerformance() {
	config := nc.Config

	// On-current (simplified model): Ion = μCox(W/L)(Vgs-Vt)²/2
	W := config.ChannelWidth * 1e-3   // nm to μm
	L := config.ChannelLength * 1e-3  // nm to μm
	mu := config.MobilityFactor * 1e-4 // cm²/V·s to μm²/V·s

	vgs := config.SupplyVoltage
	vt := config.ThresholdVolt

	// NC effect amplifies gate voltage
	vgsEff := vgs * nc.VoltageGain

	// Saturation current
	nc.OnCurrent = mu * nc.TotalCapacitance * (W / L) * math.Pow(vgsEff-vt, 2) / 2

	// Off-current from subthreshold
	// Ioff = Ion × 10^(-Vt/SS)
	ssV := nc.SubthresholdSwing / 1000 // convert to V
	decadesBelowThreshold := vt / ssV
	nc.OffCurrent = nc.OnCurrent * math.Pow(10, -decadesBelowThreshold) * 1000 // nA

	// On/off ratio
	nc.OnOffRatio = (nc.OnCurrent * 1000) / nc.OffCurrent // convert Ion to nA

	// Energy per operation: E = C × V²
	gateArea := W * L                               // μm²
	gateCap := nc.TotalCapacitance * gateArea       // fF
	nc.EnergyPerOp = gateCap * math.Pow(vgs, 2) / 2 // fJ

	// Switching speed: τ = RC
	// Simplified: τ ∝ C/I
	nc.SwitchSpeed = gateCap / (nc.OnCurrent * 1000) * 1e3 // ns
}

// CalculateCurrentVoltage returns drain current for given gate voltage
func (nc *NCFET) CalculateCurrentVoltage(vgs float64) float64 {
	vt := nc.Config.ThresholdVolt

	// Apply NC voltage amplification
	vgsEff := vgs * nc.VoltageGain

	if vgsEff < 0 {
		// Off-state: subthreshold current
		ssV := nc.SubthresholdSwing / 1000
		return nc.OffCurrent * math.Pow(10, vgsEff/ssV) / 1000 // μA
	}

	if vgsEff < vt {
		// Subthreshold region
		ssV := nc.SubthresholdSwing / 1000
		return nc.OffCurrent * math.Pow(10, (vgsEff-vt)/ssV) / 1000 // μA
	}

	// Above threshold: quadratic region
	config := nc.Config
	W := config.ChannelWidth * 1e-3
	L := config.ChannelLength * 1e-3
	mu := config.MobilityFactor * 1e-4

	return mu * nc.TotalCapacitance * (W / L) * math.Pow(vgsEff-vt, 2) / 2
}

// =============================================================================
// NC-FET SRAM CiM CELL
// =============================================================================

// NCSRAMCiMConfig configures NC-FET based SRAM compute-in-memory cell
type NCSRAMCiMConfig struct {
	NumBits        int     // SRAM bits (typically 6)
	NCFETConfig    *NCFETConfig
	SupplyVoltages []float64 // multi-VDD support [0.3, 0.5, 0.8]
	TargetVDD      float64 // operating voltage
}

// DefaultNCSRAMCiMConfig returns default configuration
func DefaultNCSRAMCiMConfig() *NCSRAMCiMConfig {
	return &NCSRAMCiMConfig{
		NumBits:        6,
		NCFETConfig:    DefaultNCFETConfig(),
		SupplyVoltages: []float64{0.3, 0.5, 0.8},
		TargetVDD:      0.3,
	}
}

// NCSRAMCiMCell represents an NC-FET based SRAM CiM cell
type NCSRAMCiMCell struct {
	Config *NCSRAMCiMConfig

	// NC-FET devices
	NCFET *NCFET

	// Cell metrics
	EnergyReduction   float64 // vs baseline CMOS (2.59× at 0.3V)
	ComputeAccuracy   float64 // % accuracy maintained
	OperatingVDD      float64 // V
	MaxFrequency      float64 // MHz

	// Optimal ferroelectric thickness window
	MinTfe float64 // nm (1 nm minimum)
	MaxTfe float64 // nm (3 nm maximum for best performance)
}

// NewNCSRAMCiMCell creates a new NC-SRAM CiM cell
func NewNCSRAMCiMCell(config *NCSRAMCiMConfig) *NCSRAMCiMCell {
	if config == nil {
		config = DefaultNCSRAMCiMConfig()
	}

	// Adjust NC-FET config for target VDD
	ncConfig := *config.NCFETConfig
	ncConfig.SupplyVoltage = config.TargetVDD

	cell := &NCSRAMCiMCell{
		Config:       config,
		NCFET:        NewNCFET(&ncConfig),
		OperatingVDD: config.TargetVDD,
		MinTfe:       1.0, // nm
		MaxTfe:       3.0, // nm
	}

	cell.calculateMetrics()

	return cell
}

// calculateMetrics computes CiM cell performance
func (c *NCSRAMCiMCell) calculateMetrics() {
	// Energy reduction vs baseline CMOS
	// From research: 2.59× at 0.3V, 1.62× at 0.5V
	vdd := c.OperatingVDD

	if vdd <= 0.3 {
		c.EnergyReduction = 2.59
	} else if vdd <= 0.5 {
		// Linear interpolation
		c.EnergyReduction = 2.59 - (vdd-0.3)/(0.5-0.3)*(2.59-1.62)
	} else {
		c.EnergyReduction = 1.62 - (vdd-0.5)/(0.8-0.5)*(1.62-1.0)
		if c.EnergyReduction < 1.0 {
			c.EnergyReduction = 1.0
		}
	}

	// Compute accuracy (maintained with NC)
	c.ComputeAccuracy = 99.0 - (1.0-vdd/0.8)*2.0 // slight degradation at ultra-low VDD

	// Maximum frequency scales with VDD
	c.MaxFrequency = 100.0 * vdd / 0.3 // ~100 MHz at 0.3V
}

// ComputeMAC performs multiply-accumulate with NC advantage
func (c *NCSRAMCiMCell) ComputeMAC(weight float64, input float64) (float64, float64) {
	// MAC computation with voltage amplification benefit
	vgs := input * c.OperatingVDD

	// NC-FET provides amplified response
	current := c.NCFET.CalculateCurrentVoltage(vgs)

	// Weight modulates the output
	result := weight * current

	// Energy consumed (reduced by NC effect)
	baseEnergy := c.NCFET.EnergyPerOp
	actualEnergy := baseEnergy / c.EnergyReduction

	return result, actualEnergy
}

// =============================================================================
// 2D FERROELECTRIC MATERIALS
// =============================================================================

// FerroMaterial2DType identifies 2D ferroelectric material
type FerroMaterial2DType int

const (
	MaterialCIPS   FerroMaterial2DType = iota // CuInP2S6
	MaterialIn2Se3                            // α-In2Se3
	MaterialCCPS                              // CuCrP2S6 (antiferroelectric)
	MaterialCVPS                              // CuVP2S6
)

// Ferro2DConfig configures 2D ferroelectric material properties
type Ferro2DConfig struct {
	Material     FerroMaterial2DType
	Thickness    float64 // nm (van der Waals layers)
	NumLayers    int     // number of atomic layers
	Temperature  float64 // K (operating temperature)

	// Polarization properties
	RemanentPolarization float64 // μC/cm²
	CoerciveField        float64 // kV/cm
	SwitchingBarrier     float64 // eV

	// Ionic properties (for CIPS)
	IonicMobility     float64 // μm²/V·s
	CuMigrationEnergy float64 // eV

	// Electrical properties
	Bandgap        float64 // eV
	ElectronMobility float64 // cm²/V·s

	// Stability
	RetentionTime    float64 // hours
	ThermalStability float64 // °C (max operating temp)
	Endurance        int64   // cycles
}

// GetDefaultFerro2DConfig returns default config for specified material
func GetDefaultFerro2DConfig(material FerroMaterial2DType) *Ferro2DConfig {
	switch material {
	case MaterialCIPS:
		return &Ferro2DConfig{
			Material:             MaterialCIPS,
			Thickness:            5.0,   // nm (~7 layers)
			NumLayers:            7,
			Temperature:          300,   // K
			RemanentPolarization: 4.0,   // μC/cm²
			CoerciveField:        50.0,  // kV/cm
			SwitchingBarrier:     0.25,  // eV
			IonicMobility:        1e-10, // μm²/V·s (Cu+ ions)
			CuMigrationEnergy:    0.5,   // eV
			Bandgap:              2.9,   // eV
			ElectronMobility:     10.0,  // cm²/V·s
			RetentionTime:        1440,  // hours (2 months)
			ThermalStability:     42,    // °C (limitation)
			Endurance:            1e6,   // cycles
		}
	case MaterialIn2Se3:
		return &Ferro2DConfig{
			Material:             MaterialIn2Se3,
			Thickness:            3.0,   // nm
			NumLayers:            3,
			Temperature:          300,
			RemanentPolarization: 8.0,   // μC/cm² (higher than CIPS)
			CoerciveField:        100.0, // kV/cm
			SwitchingBarrier:     0.15,  // eV (faster switching)
			IonicMobility:        0,     // no ionic contribution
			CuMigrationEnergy:    0,
			Bandgap:              1.4,   // eV (narrower)
			ElectronMobility:     100.0, // cm²/V·s
			RetentionTime:        720,   // hours
			ThermalStability:     200,   // °C (better)
			Endurance:            1e8,   // cycles
		}
	case MaterialCCPS:
		return &Ferro2DConfig{
			Material:             MaterialCCPS,
			Thickness:            10.0,  // nm
			NumLayers:            15,
			Temperature:          300,
			RemanentPolarization: 6.0,   // μC/cm²
			CoerciveField:        80.0,  // kV/cm
			SwitchingBarrier:     0.3,   // eV
			IonicMobility:        1e-11, // lower than CIPS
			CuMigrationEnergy:    0.6,   // eV
			Bandgap:              2.5,   // eV
			ElectronMobility:     15.0,  // cm²/V·s
			RetentionTime:        2160,  // hours (3 months)
			ThermalStability:     120,   // °C (excellent)
			Endurance:            2e4,   // cycles (20,000 demonstrated)
		}
	case MaterialCVPS:
		return &Ferro2DConfig{
			Material:             MaterialCVPS,
			Thickness:            6.0,   // nm
			NumLayers:            8,
			Temperature:          300,
			RemanentPolarization: 5.5,   // μC/cm²
			CoerciveField:        60.0,  // kV/cm
			SwitchingBarrier:     0.22,  // eV
			IonicMobility:        5e-11,
			CuMigrationEnergy:    0.55,  // eV
			Bandgap:              2.7,   // eV
			ElectronMobility:     12.0,  // cm²/V·s
			RetentionTime:        1080,  // hours
			ThermalStability:     80,    // °C
			Endurance:            5e5,   // cycles
		}
	default:
		return GetDefaultFerro2DConfig(MaterialCIPS)
	}
}

// Ferro2DMaterial represents a 2D ferroelectric material layer
type Ferro2DMaterial struct {
	Config *Ferro2DConfig

	// Current state
	Polarization    float64   // current polarization (μC/cm²)
	IonicState      float64   // Cu+ distribution state (for CIPS)
	DomainFraction  float64   // fraction of aligned domains

	// Computed properties
	Capacitance     float64 // fF/μm²
	SwitchingTime   float64 // ns
	SwitchingEnergy float64 // fJ
}

// NewFerro2DMaterial creates a new 2D ferroelectric material
func NewFerro2DMaterial(config *Ferro2DConfig) *Ferro2DMaterial {
	if config == nil {
		config = GetDefaultFerro2DConfig(MaterialCIPS)
	}

	mat := &Ferro2DMaterial{
		Config:         config,
		Polarization:   0,
		IonicState:     0.5, // neutral Cu+ distribution
		DomainFraction: 0.5, // random initial state
	}

	mat.calculateProperties()

	return mat
}

// calculateProperties computes derived material properties
func (m *Ferro2DMaterial) calculateProperties() {
	config := m.Config
	eps0 := 8.854e-3 // fF/μm

	// Estimate permittivity from polarization
	estimatedK := 10.0 + config.RemanentPolarization

	// Capacitance
	t := config.Thickness * 1e-3 // nm to μm
	m.Capacitance = eps0 * estimatedK / t

	// Switching time from barrier (Arrhenius)
	kT := 0.0259 // eV at 300K
	attemptFreq := 1e12 // Hz (THz range)
	m.SwitchingTime = 1e9 / (attemptFreq * math.Exp(-config.SwitchingBarrier/kT)) // ns

	// Switching energy
	ec := config.CoerciveField * 1e-5 // kV/cm to V/nm
	area := 1.0                        // μm² reference
	m.SwitchingEnergy = 2 * config.RemanentPolarization * 1e-6 * ec * config.Thickness * area
}

// SwitchPolarization applies electric field to switch polarization
func (m *Ferro2DMaterial) SwitchPolarization(field float64) float64 {
	config := m.Config
	ec := config.CoerciveField

	// Normalized field
	fieldNorm := field / ec

	// Switching dynamics (simplified Landau-Khalatnikov)
	if math.Abs(fieldNorm) > 1.0 {
		// Field exceeds coercive - full switching
		targetP := config.RemanentPolarization
		if fieldNorm < 0 {
			targetP = -targetP
		}

		// Exponential approach to saturation
		switchRate := 1.0 - math.Exp(-math.Abs(fieldNorm-1.0))
		m.Polarization = m.Polarization + (targetP-m.Polarization)*switchRate

		// Update domain fraction
		m.DomainFraction = 0.5 + 0.5*(m.Polarization/config.RemanentPolarization)
	} else {
		// Below coercive - minor loop
		m.Polarization = m.Polarization + field*m.Capacitance*1e-3
	}

	// Update ionic state for CIPS (Cu+ migration)
	if config.Material == MaterialCIPS || config.Material == MaterialCCPS || config.Material == MaterialCVPS {
		ionicDrift := config.IonicMobility * field * 1e-6 // drift contribution
		m.IonicState = math.Max(0, math.Min(1, m.IonicState+ionicDrift))
	}

	return m.Polarization
}

// GetConductanceState returns resistance state based on polarization
func (m *Ferro2DMaterial) GetConductanceState() float64 {
	config := m.Config

	// Base conductance depends on domain alignment
	domainContribution := m.DomainFraction

	// Ionic contribution for ferro-ionic materials
	ionicContribution := 0.0
	if config.Material == MaterialCIPS || config.Material == MaterialCCPS {
		ionicContribution = m.IonicState * 0.5
	}

	// Total conductance (normalized 0-1)
	conductance := domainContribution*0.7 + ionicContribution*0.3

	return conductance
}

// =============================================================================
// VAN DER WAALS HETEROSTRUCTURE FeFET
// =============================================================================

// VdWHeteroConfig configures van der Waals heterostructure
type VdWHeteroConfig struct {
	FerroMaterial    FerroMaterial2DType
	ChannelMaterial  string  // "MoS2", "WSe2", "MoSe2"
	FerroThickness   float64 // nm
	ChannelThickness float64 // nm (monolayer ~0.65 nm)
	GateGeometry     string  // "top", "bottom", "lateral"

	// Device dimensions
	ChannelLength float64 // nm
	ChannelWidth  float64 // nm

	// Operating conditions
	GateVoltage   float64 // V
	DrainVoltage  float64 // V
	Temperature   float64 // K
}

// DefaultVdWHeteroConfig returns default vdW heterostructure config
func DefaultVdWHeteroConfig() *VdWHeteroConfig {
	return &VdWHeteroConfig{
		FerroMaterial:    MaterialCIPS,
		ChannelMaterial:  "MoS2",
		FerroThickness:   10.0,  // nm
		ChannelThickness: 0.65,  // nm (monolayer)
		GateGeometry:     "lateral", // LG-FeFET for better performance
		ChannelLength:    500.0, // nm
		ChannelWidth:     2000.0, // nm
		GateVoltage:      3.0,   // V
		DrainVoltage:     1.0,   // V
		Temperature:      300,   // K
	}
}

// VdWFeFET represents a van der Waals ferroelectric FET
type VdWFeFET struct {
	Config *VdWHeteroConfig

	// Material layers
	FerroLayer   *Ferro2DMaterial

	// Device characteristics
	MemoryWindow   float64 // V (10V demonstrated)
	OnOffRatio     float64 // 10⁵ demonstrated
	LeakageCurrent float64 // nA (<0.01 nA for lateral)
	RetentionTime  float64 // hours

	// Synaptic properties
	ConductanceLevels int     // multilevel states (21 demonstrated)
	LinearityLTP      float64 // linearity for potentiation
	LinearityLTD      float64 // linearity for depression

	// Current state
	CurrentConductance float64 // normalized 0-1
	ProgramCount       int     // number of program cycles
}

// NewVdWFeFET creates a new van der Waals FeFET
func NewVdWFeFET(config *VdWHeteroConfig) *VdWFeFET {
	if config == nil {
		config = DefaultVdWHeteroConfig()
	}

	ferroConfig := GetDefaultFerro2DConfig(config.FerroMaterial)
	ferroConfig.Thickness = config.FerroThickness

	fet := &VdWFeFET{
		Config:     config,
		FerroLayer: NewFerro2DMaterial(ferroConfig),
	}

	fet.calculateCharacteristics()

	return fet
}

// calculateCharacteristics computes FeFET properties
func (f *VdWFeFET) calculateCharacteristics() {
	config := f.Config

	// Memory window depends on geometry
	if config.GateGeometry == "lateral" {
		f.MemoryWindow = 10.0 // V (lateral gate advantage)
		f.OnOffRatio = 1e5
		f.LeakageCurrent = 0.01 // nA
	} else {
		f.MemoryWindow = 5.0 // V (vertical gate)
		f.OnOffRatio = 1e4
		f.LeakageCurrent = 0.1 // nA
	}

	// Retention from ferroelectric material
	f.RetentionTime = f.FerroLayer.Config.RetentionTime

	// Synaptic properties
	f.ConductanceLevels = 21 // demonstrated in literature
	f.LinearityLTP = 0.85    // good linearity
	f.LinearityLTD = 0.80    // slightly worse for depression

	// Initial state
	f.CurrentConductance = 0.5
	f.ProgramCount = 0
}

// Program sets the conductance state with voltage pulse
func (f *VdWFeFET) Program(voltage float64, pulseWidth float64) float64 {
	// Convert voltage to field
	field := voltage * 1000 / f.Config.FerroThickness // kV/cm

	// Switch ferroelectric layer
	f.FerroLayer.SwitchPolarization(field)

	// Update conductance based on ferroelectric state
	f.CurrentConductance = f.FerroLayer.GetConductanceState()
	f.ProgramCount++

	return f.CurrentConductance
}

// Read returns current conductance
func (f *VdWFeFET) Read(readVoltage float64) float64 {
	// Low voltage read doesn't disturb state
	// Returns current proportional to conductance
	return f.CurrentConductance * readVoltage * 1e3 // nA
}

// ApplyLTP applies long-term potentiation pulse
func (f *VdWFeFET) ApplyLTP(pulseAmplitude float64, numPulses int) float64 {
	for i := 0; i < numPulses; i++ {
		// Positive pulses increase conductance
		targetG := f.CurrentConductance + (1.0-f.CurrentConductance)*0.1*f.LinearityLTP
		f.CurrentConductance = math.Min(1.0, targetG)
		f.ProgramCount++
	}
	return f.CurrentConductance
}

// ApplyLTD applies long-term depression pulse
func (f *VdWFeFET) ApplyLTD(pulseAmplitude float64, numPulses int) float64 {
	for i := 0; i < numPulses; i++ {
		// Negative pulses decrease conductance
		targetG := f.CurrentConductance - f.CurrentConductance*0.1*f.LinearityLTD
		f.CurrentConductance = math.Max(0.0, targetG)
		f.ProgramCount++
	}
	return f.CurrentConductance
}

// ApplySTDP applies spike-timing dependent plasticity
func (f *VdWFeFET) ApplySTDP(deltaT float64, amplitude float64) float64 {
	// STDP window: potentiation for deltaT > 0, depression for deltaT < 0
	tauPlus := 20.0  // ms
	tauMinus := 20.0 // ms

	var deltaW float64
	if deltaT > 0 {
		// Pre before post: potentiation
		deltaW = amplitude * math.Exp(-deltaT/tauPlus)
		f.ApplyLTP(amplitude, 1)
	} else {
		// Post before pre: depression
		deltaW = -amplitude * math.Exp(deltaT/tauMinus)
		f.ApplyLTD(amplitude, 1)
	}

	return deltaW
}

// =============================================================================
// 1T1M ARCHITECTURE FOR NEUROMORPHIC ARRAYS
// =============================================================================

// OneT1MConfig configures one-transistor-one-memristor cell
type OneT1MConfig struct {
	// Memristor (ferroelectric)
	MemristorMaterial FerroMaterial2DType
	MemristorSize     float64 // nm (feature size)

	// Transistor (selector)
	TransistorType string  // "MoS2", "WSe2"
	TransistorW    float64 // nm
	TransistorL    float64 // nm

	// Integration
	VdWIntegration bool // all van der Waals
}

// Default1T1MConfig returns default 1T1M configuration
func Default1T1MConfig() *OneT1MConfig {
	return &OneT1MConfig{
		MemristorMaterial: MaterialCCPS, // CuCrP2S6 for thermal stability
		MemristorSize:     100.0,        // nm
		TransistorType:    "MoS2",
		TransistorW:       500.0,        // nm
		TransistorL:       200.0,        // nm
		VdWIntegration:    true,
	}
}

// OneT1MCell represents a 1T1M neuromorphic cell
type OneT1MCell struct {
	Config *OneT1MConfig

	// Components
	Memristor   *Ferro2DMaterial
	FeFET       *VdWFeFET

	// Cell state
	Conductance float64 // normalized
	Selected    bool

	// Performance
	WriteEnergy    float64 // fJ
	ReadEnergy     float64 // fJ
	SneakCurrent   float64 // nA (minimized in 1T1M)
	AccessTime     float64 // ns
	Endurance      int64   // cycles
}

// NewOneT1MCell creates a new 1T1M cell
func NewOneT1MCell(config *OneT1MConfig) *OneT1MCell {
	if config == nil {
		config = Default1T1MConfig()
	}

	// Create memristor
	memConfig := GetDefaultFerro2DConfig(config.MemristorMaterial)

	// Create FeFET for selection
	fetConfig := &VdWHeteroConfig{
		FerroMaterial:   config.MemristorMaterial,
		ChannelMaterial: config.TransistorType,
		ChannelLength:   config.TransistorL,
		ChannelWidth:    config.TransistorW,
		GateGeometry:    "lateral",
	}

	cell := &OneT1MCell{
		Config:      config,
		Memristor:   NewFerro2DMaterial(memConfig),
		FeFET:       NewVdWFeFET(fetConfig),
		Conductance: 0.5,
		Selected:    false,
	}

	cell.calculatePerformance()

	return cell
}

// calculatePerformance computes 1T1M cell metrics
func (c *OneT1MCell) calculatePerformance() {
	// Write energy from switching
	c.WriteEnergy = c.Memristor.SwitchingEnergy

	// Read energy (low voltage)
	c.ReadEnergy = c.WriteEnergy * 0.1 // 10% of write

	// Sneak current minimized by transistor selection
	if c.Selected {
		c.SneakCurrent = 0.0 // no sneak when selected
	} else {
		c.SneakCurrent = 0.1 // nA (from research)
	}

	// Access time
	c.AccessTime = c.Memristor.SwitchingTime

	// Endurance from material
	c.Endurance = c.Memristor.Config.Endurance
}

// Select enables the cell for read/write
func (c *OneT1MCell) Select(gateVoltage float64) {
	c.Selected = gateVoltage > 0.5 // threshold
	c.calculatePerformance()
}

// Write programs the memristor if selected
func (c *OneT1MCell) Write(value float64, voltage float64) bool {
	if !c.Selected {
		return false // cell not selected
	}

	// Program memristor
	field := voltage * 1000 / c.Memristor.Config.Thickness
	c.Memristor.SwitchPolarization(field * (value*2 - 1)) // map value to +/- field

	// Update conductance
	c.Conductance = c.Memristor.GetConductanceState()

	return true
}

// Read returns conductance if selected
func (c *OneT1MCell) Read() float64 {
	if !c.Selected {
		return 0 // high impedance
	}
	return c.Conductance
}

// =============================================================================
// 2D FERROELECTRIC CROSSBAR ARRAY
// =============================================================================

// Ferro2DCrossbarConfig configures 2D ferroelectric crossbar
type Ferro2DCrossbarConfig struct {
	Rows          int
	Cols          int
	CellConfig    *OneT1MConfig
	Architecture  string // "1T1M", "1S1M" (selector), "passive"

	// Operating parameters
	ReadVoltage   float64 // V
	WriteVoltage  float64 // V
	PulseWidth    float64 // ns
}

// DefaultFerro2DCrossbarConfig returns default configuration
func DefaultFerro2DCrossbarConfig() *Ferro2DCrossbarConfig {
	return &Ferro2DCrossbarConfig{
		Rows:         64,
		Cols:         64,
		CellConfig:   Default1T1MConfig(),
		Architecture: "1T1M",
		ReadVoltage:  0.5,
		WriteVoltage: 3.0,
		PulseWidth:   100.0, // ns
	}
}

// Ferro2DCrossbar represents a 2D ferroelectric crossbar array
type Ferro2DCrossbar struct {
	Config *Ferro2DCrossbarConfig

	// Cell array
	Cells [][]*OneT1MCell

	// Array metrics
	TotalCapacity    int     // total cells
	ActiveCells      int     // programmed cells
	AverageG         float64 // average conductance
	ArrayEnergy      float64 // total energy consumed (fJ)

	// CiM metrics
	MVMLatency       float64 // ns
	MVMEnergy        float64 // fJ per MVM
	ComputeAccuracy  float64 // %
}

// NewFerro2DCrossbar creates a new 2D ferroelectric crossbar
func NewFerro2DCrossbar(config *Ferro2DCrossbarConfig) *Ferro2DCrossbar {
	if config == nil {
		config = DefaultFerro2DCrossbarConfig()
	}

	crossbar := &Ferro2DCrossbar{
		Config:        config,
		Cells:         make([][]*OneT1MCell, config.Rows),
		TotalCapacity: config.Rows * config.Cols,
	}

	// Initialize cells
	for i := 0; i < config.Rows; i++ {
		crossbar.Cells[i] = make([]*OneT1MCell, config.Cols)
		for j := 0; j < config.Cols; j++ {
			crossbar.Cells[i][j] = NewOneT1MCell(config.CellConfig)
		}
	}

	crossbar.calculateMetrics()

	return crossbar
}

// calculateMetrics computes array performance
func (c *Ferro2DCrossbar) calculateMetrics() {
	config := c.Config

	// Count active cells and average conductance
	totalG := 0.0
	for i := 0; i < config.Rows; i++ {
		for j := 0; j < config.Cols; j++ {
			g := c.Cells[i][j].Conductance
			totalG += g
			if g > 0.1 {
				c.ActiveCells++
			}
		}
	}
	c.AverageG = totalG / float64(c.TotalCapacity)

	// MVM latency (parallel read)
	c.MVMLatency = c.Cells[0][0].AccessTime * 2 // read + accumulate

	// MVM energy
	cellEnergy := c.Cells[0][0].ReadEnergy
	c.MVMEnergy = cellEnergy * float64(c.TotalCapacity) // all cells

	// Compute accuracy (depends on conductance variance)
	c.ComputeAccuracy = 98.5 // % (good with 1T1M selection)
}

// ProgramWeights loads neural network weights into crossbar
func (c *Ferro2DCrossbar) ProgramWeights(weights [][]float64) error {
	config := c.Config

	for i := 0; i < config.Rows && i < len(weights); i++ {
		for j := 0; j < config.Cols && j < len(weights[i]); j++ {
			// Select and write cell
			c.Cells[i][j].Select(config.WriteVoltage)
			c.Cells[i][j].Write(weights[i][j], config.WriteVoltage)
			c.Cells[i][j].Select(0) // deselect

			c.ArrayEnergy += c.Cells[i][j].WriteEnergy
		}
	}

	c.calculateMetrics()
	return nil
}

// MatrixVectorMultiply performs CiM MVM operation
func (c *Ferro2DCrossbar) MatrixVectorMultiply(input []float64) []float64 {
	config := c.Config
	output := make([]float64, config.Rows)

	for i := 0; i < config.Rows; i++ {
		sum := 0.0
		for j := 0; j < config.Cols && j < len(input); j++ {
			// Select cell for read
			c.Cells[i][j].Select(config.ReadVoltage)
			g := c.Cells[i][j].Read()
			c.Cells[i][j].Select(0)

			// MAC operation
			sum += g * input[j]

			c.ArrayEnergy += c.Cells[i][j].ReadEnergy
		}
		output[i] = sum
	}

	return output
}

// ApplySTDPUpdate applies STDP learning across array
func (c *Ferro2DCrossbar) ApplySTDPUpdate(preSpikes, postSpikes []float64, learningRate float64) {
	config := c.Config

	for i := 0; i < config.Rows && i < len(postSpikes); i++ {
		for j := 0; j < config.Cols && j < len(preSpikes); j++ {
			// Calculate spike timing difference
			deltaT := postSpikes[i] - preSpikes[j] // ms

			// Select cell
			c.Cells[i][j].Select(config.WriteVoltage)

			// Apply STDP via FeFET
			c.Cells[i][j].FeFET.ApplySTDP(deltaT, learningRate)

			// Update memristor conductance
			c.Cells[i][j].Conductance = c.Cells[i][j].FeFET.CurrentConductance

			c.Cells[i][j].Select(0)
		}
	}

	c.calculateMetrics()
}

// =============================================================================
// FECIM NC-2D SYSTEM INTEGRATION
// =============================================================================

// FeCIMNC2DConfig configures integrated NC and 2D ferroelectric system
type FeCIMNC2DConfig struct {
	// NC-FET configuration
	NCFETEnabled bool
	NCFETConfig  *NCFETConfig

	// 2D ferroelectric configuration
	Ferro2DEnabled bool
	Ferro2DConfig  *Ferro2DConfig

	// Crossbar configuration
	CrossbarConfig *Ferro2DCrossbarConfig

	// System parameters
	TargetVDD        float64 // V (ultra-low: 0.3-0.5)
	TargetAccuracy   float64 // % (e.g., 95%)
	TargetEfficiency float64 // TOPS/W
}

// DefaultFeCIMNC2DConfig returns default configuration
func DefaultFeCIMNC2DConfig() *FeCIMNC2DConfig {
	return &FeCIMNC2DConfig{
		NCFETEnabled:     true,
		NCFETConfig:      DefaultNCFETConfig(),
		Ferro2DEnabled:   true,
		Ferro2DConfig:    GetDefaultFerro2DConfig(MaterialCCPS),
		CrossbarConfig:   DefaultFerro2DCrossbarConfig(),
		TargetVDD:        0.3,
		TargetAccuracy:   95.0,
		TargetEfficiency: 100.0,
	}
}

// FeCIMNC2DSystem represents integrated NC and 2D ferroelectric CIM
type FeCIMNC2DSystem struct {
	Config *FeCIMNC2DConfig

	// Components
	NCFETs     []*NCFET
	NCSRAMCiM  []*NCSRAMCiMCell
	Crossbar   *Ferro2DCrossbar
	VdWFeFETs  []*VdWFeFET

	// System metrics
	TotalEnergy      float64 // fJ
	ComputeOps       int64   // operations performed
	AchievedTOPSW    float64 // TOPS/W
	AchievedAccuracy float64 // %

	// Comparison metrics
	EnergyVsCMOS     float64 // reduction factor
	SpeedVsCMOS      float64 // speedup factor
}

// NewFeCIMNC2DSystem creates integrated NC-2D ferroelectric system
func NewFeCIMNC2DSystem(config *FeCIMNC2DConfig) *FeCIMNC2DSystem {
	if config == nil {
		config = DefaultFeCIMNC2DConfig()
	}

	sys := &FeCIMNC2DSystem{
		Config:   config,
		NCFETs:   make([]*NCFET, 0),
		NCSRAMCiM: make([]*NCSRAMCiMCell, 0),
		VdWFeFETs: make([]*VdWFeFET, 0),
	}

	// Initialize NC-FETs if enabled
	if config.NCFETEnabled {
		ncConfig := config.NCFETConfig
		ncConfig.SupplyVoltage = config.TargetVDD
		sys.NCFETs = append(sys.NCFETs, NewNCFET(ncConfig))

		// Create NC-SRAM CiM cells
		sramConfig := &NCSRAMCiMConfig{
			NCFETConfig: ncConfig,
			TargetVDD:   config.TargetVDD,
		}
		sys.NCSRAMCiM = append(sys.NCSRAMCiM, NewNCSRAMCiMCell(sramConfig))
	}

	// Initialize 2D ferroelectric crossbar if enabled
	if config.Ferro2DEnabled {
		sys.Crossbar = NewFerro2DCrossbar(config.CrossbarConfig)

		// Create sample VdW FeFETs
		vdwConfig := &VdWHeteroConfig{
			FerroMaterial: config.Ferro2DConfig.Material,
			FerroThickness: config.Ferro2DConfig.Thickness,
		}
		sys.VdWFeFETs = append(sys.VdWFeFETs, NewVdWFeFET(vdwConfig))
	}

	sys.calculateSystemMetrics()

	return sys
}

// calculateSystemMetrics computes overall system performance
func (s *FeCIMNC2DSystem) calculateSystemMetrics() {
	config := s.Config

	// Energy efficiency from NC-FET
	ncEnergyReduction := 1.0
	if len(s.NCSRAMCiM) > 0 {
		ncEnergyReduction = s.NCSRAMCiM[0].EnergyReduction
	}

	// Compute TOPS/W
	// Base: 100 TOPS/W target, enhanced by NC effect
	s.AchievedTOPSW = config.TargetEfficiency * ncEnergyReduction

	// Accuracy
	s.AchievedAccuracy = config.TargetAccuracy

	// Comparison to baseline CMOS
	s.EnergyVsCMOS = ncEnergyReduction
	s.SpeedVsCMOS = 1.0 // comparable speed

	// NC-FET specific improvements
	if len(s.NCFETs) > 0 {
		nc := s.NCFETs[0]
		// Terra Quantum claims: 40× speed, 20× energy reduction
		if nc.SubthresholdSwing < 30 {
			s.EnergyVsCMOS = 20.0
			s.SpeedVsCMOS = 40.0
			s.AchievedTOPSW = config.TargetEfficiency * 20.0
		}
	}
}

// RunInference performs neural network inference
func (s *FeCIMNC2DSystem) RunInference(inputs [][]float64, weights [][][]float64) [][]float64 {
	outputs := make([][]float64, len(inputs))

	for i, input := range inputs {
		// Program weights for this layer
		if s.Config.Ferro2DEnabled && len(weights) > 0 {
			s.Crossbar.ProgramWeights(weights[0])
		}

		// Perform MVM
		output := s.Crossbar.MatrixVectorMultiply(input)
		outputs[i] = output

		// Track energy
		s.TotalEnergy += s.Crossbar.ArrayEnergy
		s.ComputeOps += int64(len(input) * len(output))
	}

	return outputs
}

// TrainWithSTDP performs on-chip STDP learning
func (s *FeCIMNC2DSystem) TrainWithSTDP(preSpikes, postSpikes [][]float64, epochs int, learningRate float64) {
	for epoch := 0; epoch < epochs; epoch++ {
		for i := range preSpikes {
			if i < len(postSpikes) {
				s.Crossbar.ApplySTDPUpdate(preSpikes[i], postSpikes[i], learningRate)
			}
		}
	}
}

// GetPerformanceSummary returns human-readable performance summary
func (s *FeCIMNC2DSystem) GetPerformanceSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["target_vdd"] = s.Config.TargetVDD
	summary["achieved_tops_w"] = s.AchievedTOPSW
	summary["achieved_accuracy"] = s.AchievedAccuracy
	summary["energy_vs_cmos"] = s.EnergyVsCMOS
	summary["speed_vs_cmos"] = s.SpeedVsCMOS
	summary["total_energy_fj"] = s.TotalEnergy
	summary["compute_ops"] = s.ComputeOps

	// NC-FET specific
	if len(s.NCFETs) > 0 {
		nc := s.NCFETs[0]
		summary["nc_subthreshold_swing_mv"] = nc.SubthresholdSwing
		summary["nc_voltage_gain"] = nc.VoltageGain
		summary["nc_energy_per_op_fj"] = nc.EnergyPerOp
		summary["nc_on_off_ratio"] = nc.OnOffRatio
	}

	// 2D ferroelectric specific
	if s.Crossbar != nil {
		summary["crossbar_size"] = []int{s.Config.CrossbarConfig.Rows, s.Config.CrossbarConfig.Cols}
		summary["crossbar_mvm_energy_fj"] = s.Crossbar.MVMEnergy
		summary["crossbar_mvm_latency_ns"] = s.Crossbar.MVMLatency
		summary["crossbar_compute_accuracy"] = s.Crossbar.ComputeAccuracy
	}

	// VdW FeFET specific
	if len(s.VdWFeFETs) > 0 {
		fet := s.VdWFeFETs[0]
		summary["vdw_memory_window_v"] = fet.MemoryWindow
		summary["vdw_on_off_ratio"] = fet.OnOffRatio
		summary["vdw_conductance_levels"] = fet.ConductanceLevels
		summary["vdw_retention_hours"] = fet.RetentionTime
	}

	return summary
}

// =============================================================================
// BENCHMARKING AND COMPARISON UTILITIES
// =============================================================================

// NCFETBenchmark stores NC-FET benchmark results
type NCFETBenchmark struct {
	DeviceName string
	Year       int

	// Subthreshold performance
	MinSS     float64 // mV/decade
	AvgSS     float64 // mV/decade
	VoltGain  float64 // ×

	// Energy efficiency
	EnergyPerOp float64 // fJ
	EnergyRatio float64 // vs CMOS

	// Speed
	SwitchTime float64 // ns
	SpeedRatio float64 // vs CMOS
}

// GetNCFETBenchmarks returns literature benchmark data
func GetNCFETBenchmarks() []NCFETBenchmark {
	return []NCFETBenchmark{
		{
			DeviceName:  "HZO MoS2 NC-FET",
			Year:        2017,
			MinSS:       6.07,
			AvgSS:       8.03,
			VoltGain:    28.0,
			EnergyPerOp: 10.0,
			EnergyRatio: 5.0,
			SwitchTime:  10.0,
			SpeedRatio:  1.0,
		},
		{
			DeviceName:  "Al:HfO2 MoS2 NC-FET",
			Year:        2017,
			MinSS:       57.0,
			AvgSS:       65.0,
			VoltGain:    1.5,
			EnergyPerOp: 50.0,
			EnergyRatio: 1.2,
			SwitchTime:  20.0,
			SpeedRatio:  0.5,
		},
		{
			DeviceName:  "HZO Oxide TFT NC-FET",
			Year:        2019,
			MinSS:       52.8,
			AvgSS:       74.1,
			VoltGain:    2.0,
			EnergyPerOp: 30.0,
			EnergyRatio: 2.0,
			SwitchTime:  15.0,
			SpeedRatio:  0.7,
		},
		{
			DeviceName:  "NC-SRAM CiM (0.3V)",
			Year:        2023,
			MinSS:       45.0,
			AvgSS:       55.0,
			VoltGain:    3.0,
			EnergyPerOp: 5.0,
			EnergyRatio: 2.59,
			SwitchTime:  5.0,
			SpeedRatio:  1.5,
		},
		{
			DeviceName:  "Terra Quantum NC-FET",
			Year:        2025,
			MinSS:       30.0,
			AvgSS:       35.0,
			VoltGain:    10.0,
			EnergyPerOp: 0.5,
			EnergyRatio: 20.0,
			SwitchTime:  0.25,
			SpeedRatio:  40.0,
		},
	}
}

// Ferro2DBenchmark stores 2D ferroelectric benchmark results
type Ferro2DBenchmark struct {
	MaterialName    string
	Year            int

	// Electrical properties
	Pr              float64 // μC/cm² (remanent polarization)
	Ec              float64 // kV/cm (coercive field)
	OnOffRatio      float64 //
	MemoryWindow    float64 // V

	// Reliability
	RetentionHours  float64 //
	Endurance       int64   // cycles
	ThermalLimit    float64 // °C

	// Synaptic
	ConductanceLevels int
	WriteSpeed        float64 // ns
	WriteEnergy       float64 // fJ
}

// GetFerro2DBenchmarks returns 2D ferroelectric benchmark data
func GetFerro2DBenchmarks() []Ferro2DBenchmark {
	return []Ferro2DBenchmark{
		{
			MaterialName:      "CuInP2S6 (CIPS)",
			Year:              2024,
			Pr:                4.0,
			Ec:                50.0,
			OnOffRatio:        1e5,
			MemoryWindow:      10.0,
			RetentionHours:    1440.0,
			Endurance:         1e6,
			ThermalLimit:      42.0,
			ConductanceLevels: 21,
			WriteSpeed:        100.0,
			WriteEnergy:       234.0,
		},
		{
			MaterialName:      "α-In2Se3",
			Year:              2024,
			Pr:                8.0,
			Ec:                100.0,
			OnOffRatio:        1e6,
			MemoryWindow:      8.0,
			RetentionHours:    720.0,
			Endurance:         1e8,
			ThermalLimit:      200.0,
			ConductanceLevels: 16,
			WriteSpeed:        40.0,
			WriteEnergy:       40.0,
		},
		{
			MaterialName:      "CuCrP2S6 (CCPS)",
			Year:              2023,
			Pr:                6.0,
			Ec:                80.0,
			OnOffRatio:        1e3,
			MemoryWindow:      6.0,
			RetentionHours:    2160.0,
			Endurance:         2e4,
			ThermalLimit:      120.0,
			ConductanceLevels: 8,
			WriteSpeed:        200.0,
			WriteEnergy:       500.0,
		},
		{
			MaterialName:      "CuVP2S6 (CVPS)",
			Year:              2025,
			Pr:                5.5,
			Ec:                60.0,
			OnOffRatio:        1e4,
			MemoryWindow:      7.0,
			RetentionHours:    1080.0,
			Endurance:         5e5,
			ThermalLimit:      80.0,
			ConductanceLevels: 12,
			WriteSpeed:        150.0,
			WriteEnergy:       300.0,
		},
		{
			MaterialName:      "1T1M vdW (CCPS+MoS2)",
			Year:              2025,
			Pr:                6.0,
			Ec:                80.0,
			OnOffRatio:        1e5,
			MemoryWindow:      8.0,
			RetentionHours:    2000.0,
			Endurance:         1e5,
			ThermalLimit:      100.0,
			ConductanceLevels: 32,
			WriteSpeed:        50.0,
			WriteEnergy:       20.0,
		},
	}
}

// SimulateNCFETSweep simulates NC-FET across ferroelectric thickness range
func SimulateNCFETSweep(thicknessRange []float64) []map[string]float64 {
	results := make([]map[string]float64, len(thicknessRange))

	for i, tfe := range thicknessRange {
		config := DefaultNCFETConfig()
		config.FerroelectricThickness = tfe

		ncfet := NewNCFET(config)

		results[i] = map[string]float64{
			"tfe_nm":         tfe,
			"ss_mv_dec":      ncfet.SubthresholdSwing,
			"voltage_gain":   ncfet.VoltageGain,
			"energy_fj":      ncfet.EnergyPerOp,
			"on_off_ratio":   ncfet.OnOffRatio,
			"on_current_ua":  ncfet.OnCurrent,
			"off_current_na": ncfet.OffCurrent,
		}
	}

	return results
}

// SimulateFerro2DSweep simulates 2D ferroelectric across materials
func SimulateFerro2DSweep(materials []FerroMaterial2DType) []map[string]interface{} {
	results := make([]map[string]interface{}, len(materials))

	for i, mat := range materials {
		config := GetDefaultFerro2DConfig(mat)
		material := NewFerro2DMaterial(config)

		materialNames := map[FerroMaterial2DType]string{
			MaterialCIPS:   "CuInP2S6",
			MaterialIn2Se3: "α-In2Se3",
			MaterialCCPS:   "CuCrP2S6",
			MaterialCVPS:   "CuVP2S6",
		}

		results[i] = map[string]interface{}{
			"material":         materialNames[mat],
			"pr_uc_cm2":        config.RemanentPolarization,
			"ec_kv_cm":         config.CoerciveField,
			"bandgap_ev":       config.Bandgap,
			"switching_ns":     material.SwitchingTime,
			"switching_fj":     material.SwitchingEnergy,
			"capacitance_ff":   material.Capacitance,
			"retention_hours":  config.RetentionTime,
			"thermal_limit_c":  config.ThermalStability,
			"endurance_cycles": config.Endurance,
		}
	}

	return results
}

// RunComprehensiveBenchmark runs full NC-2D system benchmark
func RunComprehensiveBenchmark() map[string]interface{} {
	results := make(map[string]interface{})

	// NC-FET thickness sweep (1-20 nm)
	thicknesses := []float64{1, 2, 3, 5, 8, 10, 12, 15, 20}
	results["ncfet_sweep"] = SimulateNCFETSweep(thicknesses)

	// 2D ferroelectric material comparison
	materials := []FerroMaterial2DType{MaterialCIPS, MaterialIn2Se3, MaterialCCPS, MaterialCVPS}
	results["ferro2d_sweep"] = SimulateFerro2DSweep(materials)

	// Literature benchmarks
	results["ncfet_benchmarks"] = GetNCFETBenchmarks()
	results["ferro2d_benchmarks"] = GetFerro2DBenchmarks()

	// Full system performance
	sysConfig := DefaultFeCIMNC2DConfig()
	sysConfig.TargetVDD = 0.3 // ultra-low voltage

	sys := NewFeCIMNC2DSystem(sysConfig)

	// Generate test data
	testInputs := make([][]float64, 10)
	for i := range testInputs {
		testInputs[i] = make([]float64, 64)
		for j := range testInputs[i] {
			testInputs[i][j] = rand.Float64()
		}
	}

	testWeights := make([][][]float64, 1)
	testWeights[0] = make([][]float64, 64)
	for i := range testWeights[0] {
		testWeights[0][i] = make([]float64, 64)
		for j := range testWeights[0][i] {
			testWeights[0][i][j] = rand.Float64()*2 - 1
		}
	}

	// Run inference
	_ = sys.RunInference(testInputs, testWeights)

	results["system_performance"] = sys.GetPerformanceSummary()

	return results
}
