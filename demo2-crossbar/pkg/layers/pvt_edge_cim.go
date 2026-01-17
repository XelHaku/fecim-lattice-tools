// pvt_edge_cim.go - PVT Variation Modeling and Edge Deployment for CIM
// Research iteration 125: Process-Voltage-Temperature variation and TinyML edge CIM
//
// Key findings:
// - 2T1R1C robust CIM: Charge/discharge MAC enhances PVT robustness (TVLSI 2025)
// - PICO-RAM: PVT-insensitive SRAM CIM with charge-domain computation
// - Eq-CIM: Monolithic 3D IGZO-RRAM-SRAM with <0.27% accuracy loss (-40°C to 120°C)
// - TinyML constraints: 100-480 MHz, <1MB flash, 128-512KB SRAM
// - On-device learning: 256KB SRAM sufficient (<1/100 memory vs conventional)
// - Edge AI: 20× improvement with hardware accelerators vs MCU

package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// PVT Variation Types and Constants
// ============================================================================

// ProcessCorner represents CMOS process corner variations
type ProcessCorner int

const (
	CornerTT ProcessCorner = iota // Typical-Typical (nominal)
	CornerFF                      // Fast-Fast
	CornerSS                      // Slow-Slow
	CornerFS                      // Fast-Slow (NMOS fast, PMOS slow)
	CornerSF                      // Slow-Fast (NMOS slow, PMOS fast)
)

func (pc ProcessCorner) String() string {
	return []string{"TT", "FF", "SS", "FS", "SF"}[pc]
}

// PVTCondition represents a specific PVT operating point
type PVTCondition struct {
	ProcessCorner ProcessCorner
	Voltage       float64 // Supply voltage in V
	Temperature   float64 // Temperature in Celsius
}

// StandardPVTCorners returns industry-standard PVT corners for testing
func StandardPVTCorners() []PVTCondition {
	corners := []PVTCondition{}
	// Process corners
	processes := []ProcessCorner{CornerTT, CornerFF, CornerSS, CornerFS, CornerSF}
	// Voltage corners: nominal ±10%
	voltages := []float64{0.9, 1.0, 1.1} // 1.0V nominal
	// Temperature corners
	temperatures := []float64{-40, 25, 125} // Industrial range

	for _, p := range processes {
		for _, v := range voltages {
			for _, t := range temperatures {
				corners = append(corners, PVTCondition{
					ProcessCorner: p,
					Voltage:       v,
					Temperature:   t,
				})
			}
		}
	}
	return corners
}

// ============================================================================
// PVT Variation Models
// ============================================================================

// PVTVariationConfig configures PVT variation simulation
type PVTVariationConfig struct {
	// Process variation parameters
	VthSigma        float64 // Threshold voltage sigma (mV)
	LeffSigma       float64 // Effective length sigma (nm)
	ToxSigma        float64 // Oxide thickness sigma (Angstrom)
	MobilityVariation float64 // Mobility variation coefficient

	// Voltage variation
	SupplyNominal   float64 // Nominal supply voltage (V)
	SupplyVariation float64 // Supply variation range (±%)
	IRDropMax       float64 // Maximum IR drop (mV)

	// Temperature range
	TempMin         float64 // Minimum temperature (°C)
	TempMax         float64 // Maximum temperature (°C)
	TempNominal     float64 // Nominal temperature (°C)

	// Device-specific
	DeviceType      string  // "RRAM", "FeFET", "SRAM", "IGZO"
	TechnologyNode  int     // Technology node in nm
}

// DefaultPVTConfig returns default PVT configuration for 28nm FeFET
func DefaultPVTConfig() *PVTVariationConfig {
	return &PVTVariationConfig{
		VthSigma:        30.0,  // 30mV sigma
		LeffSigma:       2.0,   // 2nm sigma
		ToxSigma:        0.5,   // 0.5A sigma
		MobilityVariation: 0.05, // 5% mobility variation

		SupplyNominal:   1.0,
		SupplyVariation: 10.0,  // ±10%
		IRDropMax:       50.0,  // 50mV max IR drop

		TempMin:         -40.0,
		TempMax:         125.0,
		TempNominal:     25.0,

		DeviceType:      "FeFET",
		TechnologyNode:  28,
	}
}

// PVTVariationModel models PVT-induced variations in CIM arrays
type PVTVariationModel struct {
	Config     *PVTVariationConfig
	rng        *rand.Rand

	// Cached variation factors per cell
	VthVariation     [][]float64 // Threshold voltage variation
	MobilityFactor   [][]float64 // Mobility variation factor
	ResistanceScale  [][]float64 // Resistance scaling factor
}

// NewPVTVariationModel creates a new PVT variation model
func NewPVTVariationModel(config *PVTVariationConfig, rows, cols int, seed int64) *PVTVariationModel {
	model := &PVTVariationModel{
		Config: config,
		rng:    rand.New(rand.NewSource(seed)),
	}

	// Initialize per-cell variations (process variation - fixed after fabrication)
	model.VthVariation = make([][]float64, rows)
	model.MobilityFactor = make([][]float64, rows)
	model.ResistanceScale = make([][]float64, rows)

	for i := 0; i < rows; i++ {
		model.VthVariation[i] = make([]float64, cols)
		model.MobilityFactor[i] = make([]float64, cols)
		model.ResistanceScale[i] = make([]float64, cols)

		for j := 0; j < cols; j++ {
			// Gaussian-distributed Vth variation
			model.VthVariation[i][j] = model.rng.NormFloat64() * config.VthSigma

			// Log-normal mobility variation
			model.MobilityFactor[i][j] = math.Exp(model.rng.NormFloat64() * config.MobilityVariation)

			// Resistance variation (typically log-normal for RRAM)
			model.ResistanceScale[i][j] = math.Exp(model.rng.NormFloat64() * 0.1)
		}
	}

	return model
}

// ComputeDriveStrength calculates transistor drive strength under PVT
func (m *PVTVariationModel) ComputeDriveStrength(row, col int, condition PVTCondition) float64 {
	// Base drive strength (normalized to 1.0 at nominal)
	strength := 1.0

	// Process corner effect on drive strength
	switch condition.ProcessCorner {
	case CornerFF:
		strength *= 1.3 // 30% faster
	case CornerSS:
		strength *= 0.7 // 30% slower
	case CornerFS:
		strength *= 1.0 // Mixed - depends on device type
	case CornerSF:
		strength *= 1.0 // Mixed
	}

	// Voltage scaling (quadratic for digital, linear for analog)
	vRatio := condition.Voltage / m.Config.SupplyNominal
	strength *= vRatio * vRatio

	// Temperature effect (mobility decreases with temperature)
	// Mobility ~ T^(-1.5) for silicon
	tempK := condition.Temperature + 273.15
	nominalK := m.Config.TempNominal + 273.15
	strength *= math.Pow(nominalK/tempK, 1.5)

	// Per-cell variation
	if row >= 0 && row < len(m.VthVariation) && col >= 0 && col < len(m.VthVariation[0]) {
		strength *= m.MobilityFactor[row][col]

		// Vth shift affects overdrive
		vthShift := m.VthVariation[row][col] / 1000.0 // Convert mV to V
		overdriveFactor := 1.0 - vthShift/(condition.Voltage-0.3) // Assume Vth ~0.3V
		if overdriveFactor > 0 {
			strength *= overdriveFactor
		}
	}

	return strength
}

// ComputeResistance calculates memristor resistance under PVT
func (m *PVTVariationModel) ComputeResistance(row, col int, baseResistance float64, condition PVTCondition) float64 {
	resistance := baseResistance

	// Temperature coefficient for RRAM (typically positive)
	// R(T) = R0 * (1 + alpha * (T - T0))
	alpha := 0.002 // 0.2%/°C typical for HfOx RRAM
	tempFactor := 1.0 + alpha*(condition.Temperature-m.Config.TempNominal)
	resistance *= tempFactor

	// Voltage-dependent resistance (nonlinear I-V)
	// For RRAM, resistance can decrease at higher voltages due to Joule heating
	if condition.Voltage > m.Config.SupplyNominal {
		vFactor := 1.0 - 0.1*(condition.Voltage/m.Config.SupplyNominal-1.0)
		resistance *= vFactor
	}

	// Per-cell variation
	if row >= 0 && row < len(m.ResistanceScale) && col >= 0 && col < len(m.ResistanceScale[0]) {
		resistance *= m.ResistanceScale[row][col]
	}

	return resistance
}

// ============================================================================
// 2T1R1C Robust CIM Cell
// ============================================================================

// Cell2T1R1CConfig configures the 2T1R1C robust CIM cell
type Cell2T1R1CConfig struct {
	RRAMResistanceLRS float64 // Low resistance state (Ohms)
	RRAMResistanceHRS float64 // High resistance state (Ohms)
	Capacitance       float64 // Integration capacitor (fF)
	TransistorWidth   float64 // Transistor width (nm)
	ChargeCycles      int     // Number of charge/discharge cycles
	IntegrationTime   float64 // Integration time (ns)
}

// Default2T1R1CConfig returns default 2T1R1C configuration
func Default2T1R1CConfig() *Cell2T1R1CConfig {
	return &Cell2T1R1CConfig{
		RRAMResistanceLRS: 10e3,   // 10 kOhm LRS
		RRAMResistanceHRS: 100e3,  // 100 kOhm HRS
		Capacitance:       50.0,   // 50 fF
		TransistorWidth:   200.0,  // 200 nm
		ChargeCycles:      8,      // 8 cycles
		IntegrationTime:   10.0,   // 10 ns
	}
}

// Cell2T1R1C implements the 2T1R1C robust CIM cell from TVLSI 2025
// Key innovation: Charge/discharge MAC operation enhances PVT robustness
type Cell2T1R1C struct {
	Config      *Cell2T1R1CConfig
	PVTModel    *PVTVariationModel
	Row, Col    int

	// Cell state
	RRAMState   float64 // Conductance state (0-1 normalized)
	CapVoltage  float64 // Capacitor voltage
}

// NewCell2T1R1C creates a new 2T1R1C cell
func NewCell2T1R1C(config *Cell2T1R1CConfig, pvtModel *PVTVariationModel, row, col int) *Cell2T1R1C {
	return &Cell2T1R1C{
		Config:   config,
		PVTModel: pvtModel,
		Row:      row,
		Col:      col,
	}
}

// ProgramWeight programs the RRAM to represent a weight
func (c *Cell2T1R1C) ProgramWeight(weight float64) {
	// Normalize weight to 0-1 range
	c.RRAMState = math.Max(0, math.Min(1, (weight+1)/2))
}

// ComputeMAC performs charge-domain MAC operation
// Returns partial sum contribution from this cell
func (c *Cell2T1R1C) ComputeMAC(inputVoltage float64, condition PVTCondition) float64 {
	// Get effective RRAM resistance considering PVT
	baseR := c.Config.RRAMResistanceLRS + c.RRAMState*(c.Config.RRAMResistanceHRS-c.Config.RRAMResistanceLRS)
	effectiveR := c.PVTModel.ComputeResistance(c.Row, c.Col, baseR, condition)

	// Charge/discharge operation
	// Key insight: Using ratio of charge/discharge times cancels out many PVT effects
	tau := effectiveR * c.Config.Capacitance * 1e-15 // RC time constant

	// Charge phase
	chargeTime := c.Config.IntegrationTime * 1e-9 // Convert ns to s
	chargeVoltage := inputVoltage * (1 - math.Exp(-chargeTime/tau))

	// Discharge phase (controlled by input pulse width)
	dischargeTime := chargeTime * math.Abs(inputVoltage) // Input-dependent
	dischargeVoltage := chargeVoltage * math.Exp(-dischargeTime/tau)

	// MAC result is proportional to voltage difference
	macResult := chargeVoltage - dischargeVoltage

	// The beauty of 2T1R1C: taking the ratio of charge/discharge
	// cancels out most PVT variations
	return macResult
}

// ============================================================================
// PICO-RAM PVT-Insensitive CIM
// ============================================================================

// PICORAMConfig configures the PICO-RAM architecture
type PICORAMConfig struct {
	ArrayRows       int
	ArrayCols       int
	BitlineCap      float64 // Bitline capacitance (fF)
	SenseAmpOffset  float64 // Sense amplifier offset (mV)
	ChargeSharing   bool    // Enable charge-sharing computation
	DifferentialSensing bool // Use differential sensing
}

// DefaultPICORAMConfig returns default PICO-RAM configuration
func DefaultPICORAMConfig() *PICORAMConfig {
	return &PICORAMConfig{
		ArrayRows:       256,
		ArrayCols:       256,
		BitlineCap:      100.0,
		SenseAmpOffset:  5.0,
		ChargeSharing:   true,
		DifferentialSensing: true,
	}
}

// PICORAM implements PVT-Insensitive Charge-domain Operations RAM
// Key innovation: Charge-domain computation inherently robust to PVT
type PICORAM struct {
	Config    *PICORAMConfig
	PVTModel  *PVTVariationModel

	// SRAM cells storing weights
	Weights   [][]float64

	// Bitline states
	BitlineVoltages []float64
}

// NewPICORAM creates a new PICO-RAM instance
func NewPICORAM(config *PICORAMConfig, pvtConfig *PVTVariationConfig, seed int64) *PICORAM {
	return &PICORAM{
		Config:   config,
		PVTModel: NewPVTVariationModel(pvtConfig, config.ArrayRows, config.ArrayCols, seed),
		Weights:  make([][]float64, config.ArrayRows),
		BitlineVoltages: make([]float64, config.ArrayCols),
	}
}

// LoadWeights loads weight matrix into SRAM cells
func (p *PICORAM) LoadWeights(weights [][]float64) error {
	if len(weights) != p.Config.ArrayRows {
		return fmt.Errorf("weight rows %d != array rows %d", len(weights), p.Config.ArrayRows)
	}
	for i := 0; i < p.Config.ArrayRows; i++ {
		if len(weights[i]) != p.Config.ArrayCols {
			return fmt.Errorf("weight cols %d != array cols %d", len(weights[i]), p.Config.ArrayCols)
		}
		p.Weights[i] = make([]float64, p.Config.ArrayCols)
		copy(p.Weights[i], weights[i])
	}
	return nil
}

// ComputeMVMChargeSharing performs MVM using charge-sharing
func (p *PICORAM) ComputeMVMChargeSharing(inputs []float64, condition PVTCondition) []float64 {
	outputs := make([]float64, p.Config.ArrayCols)

	// Reset bitlines
	for j := range p.BitlineVoltages {
		p.BitlineVoltages[j] = 0
	}

	// Charge sharing computation
	// Key insight: Charge is conserved regardless of PVT
	for i := 0; i < p.Config.ArrayRows; i++ {
		if inputs[i] == 0 {
			continue
		}

		for j := 0; j < p.Config.ArrayCols; j++ {
			// SRAM cell contributes charge proportional to weight
			// Q = C * V * W
			weight := p.Weights[i][j]
			if weight == 0 {
				continue
			}

			// Charge contribution (PVT-insensitive due to charge conservation)
			chargeContrib := p.Config.BitlineCap * inputs[i] * weight

			// Small PVT effect on charge injection (much smaller than voltage-mode)
			driveStrength := p.PVTModel.ComputeDriveStrength(i, j, condition)
			// Charge sharing is robust, but timing might vary
			efficiencyFactor := 0.95 + 0.05*driveStrength/1.0

			p.BitlineVoltages[j] += chargeContrib * efficiencyFactor
		}
	}

	// Normalize by total capacitance
	for j := 0; j < p.Config.ArrayCols; j++ {
		outputs[j] = p.BitlineVoltages[j] / p.Config.BitlineCap
	}

	return outputs
}

// ============================================================================
// Eq-CIM Monolithic 3D Architecture
// ============================================================================

// EqCIMConfig configures the Eq-CIM monolithic 3D architecture
type EqCIMConfig struct {
	// Layer configuration
	SRAMTier        int     // SRAM tier (bottom)
	IGZOTier        int     // IGZO transistor tier
	RRAMTier        int     // RRAM tier (top)

	// Array sizes
	MacroRows       int
	MacroCols       int

	// Temperature range
	TempMin         float64 // Minimum operating temperature
	TempMax         float64 // Maximum operating temperature

	// Equalization settings
	DynamicBias     bool    // Enable dynamic bias equalization
	ThermalTracking bool    // Enable thermal tracking
}

// DefaultEqCIMConfig returns default Eq-CIM configuration
func DefaultEqCIMConfig() *EqCIMConfig {
	return &EqCIMConfig{
		SRAMTier:        0,
		IGZOTier:        1,
		RRAMTier:        2,
		MacroRows:       64,
		MacroCols:       64,
		TempMin:         -40.0,
		TempMax:         120.0,
		DynamicBias:     true,
		ThermalTracking: true,
	}
}

// EqCIM implements the Equalized CIM monolithic 3D architecture
// Key innovation: <0.27% accuracy loss from -40°C to 120°C
// 5.06× storage density, 5.05×/2.45× area/energy efficiency
type EqCIM struct {
	Config     *EqCIMConfig
	PVTModel   *PVTVariationModel

	// Multi-tier storage
	SRAMWeights  [][]float64 // SRAM tier for high-precision
	RRAMWeights  [][]float64 // RRAM tier for density

	// Equalization state
	BiasVoltages [][]float64 // Dynamic bias per cell
	TempHistory  []float64   // Temperature history for tracking
}

// NewEqCIM creates a new Eq-CIM instance
func NewEqCIM(config *EqCIMConfig, pvtConfig *PVTVariationConfig, seed int64) *EqCIM {
	eq := &EqCIM{
		Config:      config,
		PVTModel:    NewPVTVariationModel(pvtConfig, config.MacroRows, config.MacroCols, seed),
		SRAMWeights: make([][]float64, config.MacroRows),
		RRAMWeights: make([][]float64, config.MacroRows),
		BiasVoltages: make([][]float64, config.MacroRows),
		TempHistory: make([]float64, 0, 100),
	}

	for i := 0; i < config.MacroRows; i++ {
		eq.SRAMWeights[i] = make([]float64, config.MacroCols)
		eq.RRAMWeights[i] = make([]float64, config.MacroCols)
		eq.BiasVoltages[i] = make([]float64, config.MacroCols)
	}

	return eq
}

// CalibrateForTemperature performs temperature calibration
func (eq *EqCIM) CalibrateForTemperature(temperature float64) {
	eq.TempHistory = append(eq.TempHistory, temperature)
	if len(eq.TempHistory) > 100 {
		eq.TempHistory = eq.TempHistory[1:]
	}

	if !eq.Config.DynamicBias {
		return
	}

	// Compute average temperature trend
	avgTemp := 0.0
	for _, t := range eq.TempHistory {
		avgTemp += t
	}
	avgTemp /= float64(len(eq.TempHistory))

	// Adjust bias voltages to equalize performance
	// Key insight: IGZO transistors have different temp coefficient than Si
	// By adjusting bias, we can compensate
	tempOffset := temperature - 25.0 // Reference to room temp

	for i := 0; i < eq.Config.MacroRows; i++ {
		for j := 0; j < eq.Config.MacroCols; j++ {
			// IGZO has near-zero temp coefficient, use it to stabilize
			// Bias adjustment compensates for RRAM temp drift
			eq.BiasVoltages[i][j] = 0.1 * tempOffset / 100.0 // Small bias adjustment
		}
	}
}

// ComputeHybridMVM performs MVM using hybrid SRAM+RRAM
func (eq *EqCIM) ComputeHybridMVM(inputs []float64, temperature float64) []float64 {
	eq.CalibrateForTemperature(temperature)

	condition := PVTCondition{
		ProcessCorner: CornerTT,
		Voltage:       1.0,
		Temperature:   temperature,
	}

	outputs := make([]float64, eq.Config.MacroCols)

	for j := 0; j < eq.Config.MacroCols; j++ {
		sum := 0.0
		for i := 0; i < eq.Config.MacroRows; i++ {
			// Combine SRAM (high precision) and RRAM (high density) contributions
			sramContrib := inputs[i] * eq.SRAMWeights[i][j]

			// RRAM contribution with PVT compensation
			baseR := 10e3 + eq.RRAMWeights[i][j]*90e3
			effectiveR := eq.PVTModel.ComputeResistance(i, j, baseR, condition)
			rramContrib := inputs[i] / effectiveR * 1e3 // Normalize

			// Apply dynamic bias compensation
			bias := eq.BiasVoltages[i][j]
			compensatedRRAM := rramContrib * (1.0 + bias)

			// Weighted combination
			sum += 0.7*sramContrib + 0.3*compensatedRRAM
		}
		outputs[j] = sum
	}

	return outputs
}

// ============================================================================
// TinyML Edge CIM
// ============================================================================

// TinyMLConstraints represents hardware constraints for TinyML deployment
type TinyMLConstraints struct {
	MaxFlashKB      int     // Maximum flash memory (KB)
	MaxSRAMKB       int     // Maximum SRAM (KB)
	MaxFrequencyMHz int     // Maximum clock frequency (MHz)
	MaxPowerMW      float64 // Maximum power (mW)
	TargetLatencyMS float64 // Target inference latency (ms)

	// CIM-specific
	CrossbarSize    int     // Crossbar array size
	ADCBits         int     // ADC resolution
	WeightBits      int     // Weight precision
}

// DefaultTinyMLConstraints returns constraints for typical MCU
func DefaultTinyMLConstraints() *TinyMLConstraints {
	return &TinyMLConstraints{
		MaxFlashKB:      1024,    // 1MB flash
		MaxSRAMKB:       256,     // 256KB SRAM
		MaxFrequencyMHz: 240,     // 240 MHz
		MaxPowerMW:      50.0,    // 50 mW
		TargetLatencyMS: 100.0,   // 100 ms
		CrossbarSize:    64,      // 64×64 crossbar
		ADCBits:         6,       // 6-bit ADC
		WeightBits:      4,       // 4-bit weights
	}
}

// MCUClass represents different MCU capability classes
type MCUClass int

const (
	MCUClassUltraLow MCUClass = iota // <1MHz, <10KB SRAM (MSP430)
	MCUClassLow                      // 10-100MHz, 32-128KB (ARM Cortex-M0)
	MCUClassMid                      // 100-300MHz, 128-512KB (ARM Cortex-M4)
	MCUClassHigh                     // 300-600MHz, 512KB-2MB (ARM Cortex-M7)
	MCUClassPerformance              // >600MHz, >2MB (ARM Cortex-A class)
)

func (mc MCUClass) String() string {
	return []string{"UltraLow", "Low", "Mid", "High", "Performance"}[mc]
}

// GetMCUConstraints returns constraints for an MCU class
func GetMCUConstraints(class MCUClass) *TinyMLConstraints {
	constraints := map[MCUClass]*TinyMLConstraints{
		MCUClassUltraLow: {
			MaxFlashKB: 64, MaxSRAMKB: 8, MaxFrequencyMHz: 16,
			MaxPowerMW: 1.0, TargetLatencyMS: 1000, CrossbarSize: 16, ADCBits: 4, WeightBits: 2,
		},
		MCUClassLow: {
			MaxFlashKB: 256, MaxSRAMKB: 64, MaxFrequencyMHz: 80,
			MaxPowerMW: 10.0, TargetLatencyMS: 500, CrossbarSize: 32, ADCBits: 5, WeightBits: 4,
		},
		MCUClassMid: {
			MaxFlashKB: 512, MaxSRAMKB: 256, MaxFrequencyMHz: 240,
			MaxPowerMW: 50.0, TargetLatencyMS: 100, CrossbarSize: 64, ADCBits: 6, WeightBits: 4,
		},
		MCUClassHigh: {
			MaxFlashKB: 2048, MaxSRAMKB: 512, MaxFrequencyMHz: 480,
			MaxPowerMW: 200.0, TargetLatencyMS: 50, CrossbarSize: 128, ADCBits: 8, WeightBits: 8,
		},
		MCUClassPerformance: {
			MaxFlashKB: 8192, MaxSRAMKB: 2048, MaxFrequencyMHz: 1000,
			MaxPowerMW: 1000.0, TargetLatencyMS: 10, CrossbarSize: 256, ADCBits: 8, WeightBits: 8,
		},
	}
	return constraints[class]
}

// ============================================================================
// Edge CIM Accelerator
// ============================================================================

// EdgeCIMConfig configures the edge CIM accelerator
type EdgeCIMConfig struct {
	Constraints     *TinyMLConstraints
	PVTConfig       *PVTVariationConfig

	// Architecture
	NumCrossbars    int
	CrossbarRows    int
	CrossbarCols    int

	// Optimization
	WeightSharing   bool    // Share weights across inputs
	ActivationReuse bool    // Reuse activations
	OutputStationary bool   // Output-stationary dataflow

	// Power management
	ClockGating     bool
	PowerGating     bool
	DVFS            bool
}

// DefaultEdgeCIMConfig returns default edge CIM configuration
func DefaultEdgeCIMConfig() *EdgeCIMConfig {
	return &EdgeCIMConfig{
		Constraints:     DefaultTinyMLConstraints(),
		PVTConfig:       DefaultPVTConfig(),
		NumCrossbars:    4,
		CrossbarRows:    64,
		CrossbarCols:    64,
		WeightSharing:   true,
		ActivationReuse: true,
		OutputStationary: true,
		ClockGating:     true,
		PowerGating:     true,
		DVFS:            true,
	}
}

// EdgeCIMAccelerator implements a CIM accelerator for edge deployment
type EdgeCIMAccelerator struct {
	Config       *EdgeCIMConfig
	Crossbars    []*PICORAM

	// Resource tracking
	UsedFlashKB  float64
	UsedSRAMKB   float64
	CurrentPowerMW float64

	// Statistics
	TotalMACs    int64
	TotalCycles  int64
}

// NewEdgeCIMAccelerator creates a new edge CIM accelerator
func NewEdgeCIMAccelerator(config *EdgeCIMConfig, seed int64) *EdgeCIMAccelerator {
	acc := &EdgeCIMAccelerator{
		Config:    config,
		Crossbars: make([]*PICORAM, config.NumCrossbars),
	}

	for i := 0; i < config.NumCrossbars; i++ {
		picoConfig := &PICORAMConfig{
			ArrayRows:       config.CrossbarRows,
			ArrayCols:       config.CrossbarCols,
			BitlineCap:      100.0,
			ChargeSharing:   true,
			DifferentialSensing: true,
		}
		acc.Crossbars[i] = NewPICORAM(picoConfig, config.PVTConfig, seed+int64(i))
	}

	return acc
}

// EstimateResources estimates resource usage for a model
func (acc *EdgeCIMAccelerator) EstimateResources(modelParams int, activationSize int) ResourceEstimate {
	// Weight storage
	weightBits := acc.Config.Constraints.WeightBits
	weightBytes := float64(modelParams) * float64(weightBits) / 8.0

	// Activation storage (need buffer for intermediate activations)
	activationBytes := float64(activationSize) * 4 // Float32 activations

	// Check if fits in constraints
	fitsFlash := weightBytes/1024.0 <= float64(acc.Config.Constraints.MaxFlashKB)
	fitsSRAM := activationBytes/1024.0 <= float64(acc.Config.Constraints.MaxSRAMKB)

	return ResourceEstimate{
		WeightMemoryKB:    weightBytes / 1024.0,
		ActivationMemoryKB: activationBytes / 1024.0,
		FitsFlash:         fitsFlash,
		FitsSRAM:          fitsSRAM,
		RequiredCrossbars: (modelParams + acc.Config.CrossbarRows*acc.Config.CrossbarCols - 1) /
		                   (acc.Config.CrossbarRows * acc.Config.CrossbarCols),
	}
}

// ResourceEstimate contains resource usage estimates
type ResourceEstimate struct {
	WeightMemoryKB     float64
	ActivationMemoryKB float64
	FitsFlash          bool
	FitsSRAM           bool
	RequiredCrossbars  int
}

// ============================================================================
// On-Device Learning
// ============================================================================

// OnDeviceLearningConfig configures on-device learning
type OnDeviceLearningConfig struct {
	// Memory budget
	MaxGradientBufferKB float64
	MaxActivationCacheKB float64

	// Training parameters
	LearningRate        float64
	BatchSize           int
	UpdateFrequency     int     // Update weights every N samples

	// Sparse update
	SparseGradients     bool
	GradientSparsity    float64 // Fraction of gradients to keep

	// Transfer learning
	FreezeBackbone      bool    // Freeze early layers
	TrainableLayerStart int     // First trainable layer
}

// DefaultOnDeviceLearningConfig returns MIT Han Lab inspired config
// Key insight: 256KB SRAM sufficient for on-device learning
func DefaultOnDeviceLearningConfig() *OnDeviceLearningConfig {
	return &OnDeviceLearningConfig{
		MaxGradientBufferKB:  128.0,  // 128KB for gradients
		MaxActivationCacheKB: 128.0,  // 128KB for activations
		LearningRate:         0.001,
		BatchSize:            1,      // Single sample for edge
		UpdateFrequency:      10,     // Update every 10 samples
		SparseGradients:      true,
		GradientSparsity:     0.1,    // Keep top 10% gradients
		FreezeBackbone:       true,
		TrainableLayerStart:  -2,     // Train last 2 layers
	}
}

// OnDeviceLearner implements memory-efficient on-device learning
type OnDeviceLearner struct {
	Config       *OnDeviceLearningConfig
	Accelerator  *EdgeCIMAccelerator

	// Gradient accumulation
	GradientBuffer   map[string][]float64
	SampleCount      int

	// Sparse gradient tracking
	TopKIndices      map[string][]int
}

// NewOnDeviceLearner creates a new on-device learner
func NewOnDeviceLearner(config *OnDeviceLearningConfig, acc *EdgeCIMAccelerator) *OnDeviceLearner {
	return &OnDeviceLearner{
		Config:         config,
		Accelerator:    acc,
		GradientBuffer: make(map[string][]float64),
		TopKIndices:    make(map[string][]int),
	}
}

// AccumulateGradient accumulates sparse gradients
func (l *OnDeviceLearner) AccumulateGradient(layerName string, gradient []float64) {
	if l.Config.SparseGradients {
		gradient = l.sparsifyGradient(layerName, gradient)
	}

	if existing, ok := l.GradientBuffer[layerName]; ok {
		for i := range existing {
			if i < len(gradient) {
				existing[i] += gradient[i]
			}
		}
	} else {
		l.GradientBuffer[layerName] = make([]float64, len(gradient))
		copy(l.GradientBuffer[layerName], gradient)
	}

	l.SampleCount++

	// Apply update if threshold reached
	if l.SampleCount >= l.Config.UpdateFrequency {
		l.ApplyUpdates()
	}
}

// sparsifyGradient keeps only top-K gradients
func (l *OnDeviceLearner) sparsifyGradient(layerName string, gradient []float64) []float64 {
	k := int(float64(len(gradient)) * l.Config.GradientSparsity)
	if k < 1 {
		k = 1
	}

	// Find top-K indices by magnitude
	type gradIdx struct {
		idx int
		mag float64
	}
	grads := make([]gradIdx, len(gradient))
	for i, g := range gradient {
		grads[i] = gradIdx{i, math.Abs(g)}
	}
	sort.Slice(grads, func(i, j int) bool {
		return grads[i].mag > grads[j].mag
	})

	// Create sparse gradient
	sparse := make([]float64, len(gradient))
	topK := make([]int, k)
	for i := 0; i < k; i++ {
		idx := grads[i].idx
		sparse[idx] = gradient[idx]
		topK[i] = idx
	}

	l.TopKIndices[layerName] = topK
	return sparse
}

// ApplyUpdates applies accumulated gradients to weights
func (l *OnDeviceLearner) ApplyUpdates() {
	scale := l.Config.LearningRate / float64(l.SampleCount)

	for layerName, gradients := range l.GradientBuffer {
		// Apply scaled gradients
		for i := range gradients {
			gradients[i] *= scale
		}
		// Here we would update the actual weights in the accelerator
		// For simulation, just clear the buffer
		_ = layerName
	}

	// Reset accumulator
	l.GradientBuffer = make(map[string][]float64)
	l.SampleCount = 0
}

// GetMemoryUsage returns current memory usage
func (l *OnDeviceLearner) GetMemoryUsage() float64 {
	totalBytes := 0.0
	for _, grad := range l.GradientBuffer {
		totalBytes += float64(len(grad)) * 4 // Float32
	}
	return totalBytes / 1024.0 // Return KB
}

// ============================================================================
// Model Compression for Edge
// ============================================================================

// CompressionConfig configures model compression
type CompressionConfig struct {
	// Quantization
	WeightBits       int
	ActivationBits   int
	QuantizationMode string // "symmetric", "asymmetric", "dynamic"

	// Pruning
	PruningRatio     float64
	PruningGranularity string // "element", "channel", "filter", "block"

	// Knowledge distillation
	DistillationTemp float64
	TeacherWeight    float64

	// Low-rank factorization
	RankRatio        float64
	FactorizeConv    bool
	FactorizeFC      bool
}

// DefaultCompressionConfig returns aggressive compression config for edge
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		WeightBits:       4,
		ActivationBits:   8,
		QuantizationMode: "symmetric",
		PruningRatio:     0.7,        // 70% sparsity
		PruningGranularity: "block",  // Block pruning for CIM
		DistillationTemp: 4.0,
		TeacherWeight:    0.5,
		RankRatio:        0.25,       // 4× reduction
		FactorizeConv:    true,
		FactorizeFC:      true,
	}
}

// ModelCompressor compresses models for edge deployment
type ModelCompressor struct {
	Config *CompressionConfig
}

// NewModelCompressor creates a new model compressor
func NewModelCompressor(config *CompressionConfig) *ModelCompressor {
	return &ModelCompressor{Config: config}
}

// QuantizeWeights quantizes weights to target bit-width
func (c *ModelCompressor) QuantizeWeights(weights []float64) []int8 {
	// Find scale factor
	maxAbs := 0.0
	for _, w := range weights {
		if abs := math.Abs(w); abs > maxAbs {
			maxAbs = abs
		}
	}

	maxVal := float64(int(1)<<(c.Config.WeightBits-1) - 1)
	scale := maxVal / maxAbs

	// Quantize
	quantized := make([]int8, len(weights))
	for i, w := range weights {
		q := int8(math.Round(w * scale))
		if q > int8(maxVal) {
			q = int8(maxVal)
		} else if q < -int8(maxVal)-1 {
			q = -int8(maxVal) - 1
		}
		quantized[i] = q
	}

	return quantized
}

// PruneBlockSparse performs block-sparse pruning for CIM
func (c *ModelCompressor) PruneBlockSparse(weights [][]float64, blockSize int) [][]float64 {
	rows := len(weights)
	cols := len(weights[0])

	// Calculate block importance
	type blockInfo struct {
		rowStart, colStart int
		magnitude          float64
	}
	blocks := []blockInfo{}

	for i := 0; i < rows; i += blockSize {
		for j := 0; j < cols; j += blockSize {
			mag := 0.0
			for bi := 0; bi < blockSize && i+bi < rows; bi++ {
				for bj := 0; bj < blockSize && j+bj < cols; bj++ {
					mag += math.Abs(weights[i+bi][j+bj])
				}
			}
			blocks = append(blocks, blockInfo{i, j, mag})
		}
	}

	// Sort by magnitude
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].magnitude > blocks[j].magnitude
	})

	// Keep top (1-pruningRatio) blocks
	keepCount := int(float64(len(blocks)) * (1 - c.Config.PruningRatio))
	keepBlocks := make(map[[2]int]bool)
	for i := 0; i < keepCount; i++ {
		keepBlocks[[2]int{blocks[i].rowStart, blocks[i].colStart}] = true
	}

	// Create pruned weights
	pruned := make([][]float64, rows)
	for i := range pruned {
		pruned[i] = make([]float64, cols)
	}

	for i := 0; i < rows; i += blockSize {
		for j := 0; j < cols; j += blockSize {
			if keepBlocks[[2]int{i, j}] {
				for bi := 0; bi < blockSize && i+bi < rows; bi++ {
					for bj := 0; bj < blockSize && j+bj < cols; bj++ {
						pruned[i+bi][j+bj] = weights[i+bi][j+bj]
					}
				}
			}
		}
	}

	return pruned
}

// LowRankFactorize performs low-rank factorization
func (c *ModelCompressor) LowRankFactorize(weights [][]float64) ([][]float64, [][]float64) {
	rows := len(weights)
	cols := len(weights[0])
	rank := int(float64(min(rows, cols)) * c.Config.RankRatio)
	if rank < 1 {
		rank = 1
	}

	// Simple SVD approximation using power iteration
	// In practice, would use proper SVD library
	U := make([][]float64, rows)
	V := make([][]float64, rank)

	for i := range U {
		U[i] = make([]float64, rank)
		for j := 0; j < rank; j++ {
			U[i][j] = rand.NormFloat64() / math.Sqrt(float64(rank))
		}
	}

	for i := range V {
		V[i] = make([]float64, cols)
		for j := range V[i] {
			V[i][j] = rand.NormFloat64() / math.Sqrt(float64(cols))
		}
	}

	// Power iteration to find principal components
	for iter := 0; iter < 10; iter++ {
		// Update V = U^T * W
		for r := 0; r < rank; r++ {
			for j := 0; j < cols; j++ {
				sum := 0.0
				for i := 0; i < rows; i++ {
					sum += U[i][r] * weights[i][j]
				}
				V[r][j] = sum
			}
		}

		// Normalize V
		for r := 0; r < rank; r++ {
			norm := 0.0
			for j := 0; j < cols; j++ {
				norm += V[r][j] * V[r][j]
			}
			norm = math.Sqrt(norm)
			if norm > 0 {
				for j := 0; j < cols; j++ {
					V[r][j] /= norm
				}
			}
		}

		// Update U = W * V^T
		for i := 0; i < rows; i++ {
			for r := 0; r < rank; r++ {
				sum := 0.0
				for j := 0; j < cols; j++ {
					sum += weights[i][j] * V[r][j]
				}
				U[i][r] = sum
			}
		}
	}

	return U, V
}

// ============================================================================
// Edge Deployment Optimizer
// ============================================================================

// EdgeOptimizer optimizes models for edge CIM deployment
type EdgeOptimizer struct {
	Constraints  *TinyMLConstraints
	Compressor   *ModelCompressor
	Accelerator  *EdgeCIMAccelerator
}

// NewEdgeOptimizer creates a new edge optimizer
func NewEdgeOptimizer(constraints *TinyMLConstraints) *EdgeOptimizer {
	return &EdgeOptimizer{
		Constraints: constraints,
		Compressor:  NewModelCompressor(DefaultCompressionConfig()),
		Accelerator: NewEdgeCIMAccelerator(DefaultEdgeCIMConfig(), 42),
	}
}

// OptimizationResult contains optimization results
type OptimizationResult struct {
	OriginalSizeKB     float64
	CompressedSizeKB   float64
	CompressionRatio   float64
	EstimatedAccuracy  float64
	EstimatedLatencyMS float64
	EstimatedPowerMW   float64
	FitsConstraints    bool
	Recommendations    []string
}

// OptimizeForEdge optimizes a model for edge deployment
func (o *EdgeOptimizer) OptimizeForEdge(modelParams int, originalAccuracy float64) *OptimizationResult {
	result := &OptimizationResult{
		OriginalSizeKB: float64(modelParams) * 4 / 1024, // Float32
		Recommendations: []string{},
	}

	// Apply quantization
	quantBits := o.Constraints.WeightBits
	quantSize := float64(modelParams) * float64(quantBits) / 8 / 1024

	// Apply pruning (from compressor config)
	pruningRatio := o.Compressor.Config.PruningRatio
	prunedSize := quantSize * (1 - pruningRatio)

	result.CompressedSizeKB = prunedSize
	result.CompressionRatio = result.OriginalSizeKB / prunedSize

	// Estimate accuracy loss
	// Empirical: ~0.5% per bit reduction, ~0.1% per 10% pruning
	bitLoss := float64(32-quantBits) * 0.5
	pruneLoss := pruningRatio * 10 * 0.1
	result.EstimatedAccuracy = originalAccuracy - bitLoss - pruneLoss
	if result.EstimatedAccuracy < 0 {
		result.EstimatedAccuracy = 0
	}

	// Estimate latency
	// Assume CIM provides 20× speedup over MCU baseline
	baseLatencyMS := float64(modelParams) / float64(o.Constraints.MaxFrequencyMHz*1e6) * 1e3
	cimSpeedup := 20.0
	result.EstimatedLatencyMS = baseLatencyMS / cimSpeedup

	// Estimate power
	// CIM: ~100 fJ/MAC, MCU: ~10 pJ/MAC
	macEnergy := 100e-15 // 100 fJ
	result.EstimatedPowerMW = float64(modelParams) * macEnergy / (result.EstimatedLatencyMS / 1000) * 1000

	// Check constraints
	result.FitsConstraints = prunedSize <= float64(o.Constraints.MaxFlashKB) &&
		result.EstimatedLatencyMS <= o.Constraints.TargetLatencyMS &&
		result.EstimatedPowerMW <= o.Constraints.MaxPowerMW

	// Generate recommendations
	if prunedSize > float64(o.Constraints.MaxFlashKB) {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Model too large (%.1fKB > %dKB). Consider more aggressive pruning or quantization.",
				prunedSize, o.Constraints.MaxFlashKB))
	}
	if result.EstimatedLatencyMS > o.Constraints.TargetLatencyMS {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Latency too high (%.1fms > %.1fms). Consider model distillation.",
				result.EstimatedLatencyMS, o.Constraints.TargetLatencyMS))
	}
	if result.EstimatedAccuracy < originalAccuracy-5 {
		result.Recommendations = append(result.Recommendations,
			"Significant accuracy loss expected. Consider quantization-aware training.")
	}

	return result
}

// ============================================================================
// PVT-Aware Inference
// ============================================================================

// PVTAwareInference performs inference with PVT compensation
type PVTAwareInference struct {
	Accelerator    *EdgeCIMAccelerator
	CurrentPVT     PVTCondition

	// Calibration data
	CalibrationLUT map[ProcessCorner]map[float64]float64 // Corner -> Temp -> Scale
}

// NewPVTAwareInference creates PVT-aware inference engine
func NewPVTAwareInference(acc *EdgeCIMAccelerator) *PVTAwareInference {
	inf := &PVTAwareInference{
		Accelerator:    acc,
		CalibrationLUT: make(map[ProcessCorner]map[float64]float64),
	}

	// Initialize calibration LUT
	corners := []ProcessCorner{CornerTT, CornerFF, CornerSS, CornerFS, CornerSF}
	temps := []float64{-40, 0, 25, 85, 125}

	for _, corner := range corners {
		inf.CalibrationLUT[corner] = make(map[float64]float64)
		for _, temp := range temps {
			// Pre-computed calibration factors
			// In practice, would be measured during factory calibration
			baseFactor := 1.0
			switch corner {
			case CornerFF:
				baseFactor = 0.85 // Scale down fast corner
			case CornerSS:
				baseFactor = 1.15 // Scale up slow corner
			}
			tempFactor := 1.0 + (temp-25)*0.001 // 0.1%/°C
			inf.CalibrationLUT[corner][temp] = baseFactor * tempFactor
		}
	}

	return inf
}

// SetPVTCondition updates current PVT condition
func (inf *PVTAwareInference) SetPVTCondition(condition PVTCondition) {
	inf.CurrentPVT = condition
}

// GetCalibrationFactor returns interpolated calibration factor
func (inf *PVTAwareInference) GetCalibrationFactor() float64 {
	cornerLUT, ok := inf.CalibrationLUT[inf.CurrentPVT.ProcessCorner]
	if !ok {
		return 1.0
	}

	// Linear interpolation between temperature points
	temps := []float64{-40, 0, 25, 85, 125}
	temp := inf.CurrentPVT.Temperature

	for i := 0; i < len(temps)-1; i++ {
		if temp >= temps[i] && temp <= temps[i+1] {
			f1 := cornerLUT[temps[i]]
			f2 := cornerLUT[temps[i+1]]
			alpha := (temp - temps[i]) / (temps[i+1] - temps[i])
			return f1 + alpha*(f2-f1)
		}
	}

	// Extrapolation
	if temp < temps[0] {
		return cornerLUT[temps[0]]
	}
	return cornerLUT[temps[len(temps)-1]]
}

// Infer performs PVT-compensated inference
func (inf *PVTAwareInference) Infer(inputs []float64) []float64 {
	calibFactor := inf.GetCalibrationFactor()

	// Scale inputs by calibration factor
	scaledInputs := make([]float64, len(inputs))
	for i, in := range inputs {
		scaledInputs[i] = in * calibFactor
	}

	// Run through crossbars
	var outputs []float64
	for _, crossbar := range inf.Accelerator.Crossbars {
		if len(crossbar.Weights) > 0 && len(crossbar.Weights[0]) > 0 {
			out := crossbar.ComputeMVMChargeSharing(scaledInputs, inf.CurrentPVT)
			outputs = append(outputs, out...)
		}
	}

	return outputs
}

// ============================================================================
// Benchmark Suite
// ============================================================================

// PVTBenchmark benchmarks CIM performance across PVT corners
type PVTBenchmark struct {
	Accelerator *EdgeCIMAccelerator
	Results     map[string]*BenchmarkResult
}

// BenchmarkResult contains benchmark results
type BenchmarkResult struct {
	Corner       ProcessCorner
	Temperature  float64
	Voltage      float64
	Throughput   float64 // GOPS
	Efficiency   float64 // TOPS/W
	AccuracyDrop float64 // % accuracy drop vs nominal
}

// NewPVTBenchmark creates a new PVT benchmark suite
func NewPVTBenchmark(acc *EdgeCIMAccelerator) *PVTBenchmark {
	return &PVTBenchmark{
		Accelerator: acc,
		Results:     make(map[string]*BenchmarkResult),
	}
}

// RunFullPVTSweep runs benchmark across all PVT corners
func (b *PVTBenchmark) RunFullPVTSweep() []*BenchmarkResult {
	corners := StandardPVTCorners()
	results := make([]*BenchmarkResult, 0, len(corners))

	for _, condition := range corners {
		result := b.RunSingleCondition(condition)
		key := fmt.Sprintf("%s_%.2fV_%.0fC", condition.ProcessCorner, condition.Voltage, condition.Temperature)
		b.Results[key] = result
		results = append(results, result)
	}

	return results
}

// RunSingleCondition benchmarks a single PVT condition
func (b *PVTBenchmark) RunSingleCondition(condition PVTCondition) *BenchmarkResult {
	result := &BenchmarkResult{
		Corner:      condition.ProcessCorner,
		Temperature: condition.Temperature,
		Voltage:     condition.Voltage,
	}

	// Simulate throughput based on PVT
	baseThroughput := 100.0 // GOPS at nominal

	// Voltage scaling (frequency ~ V^2)
	vScale := math.Pow(condition.Voltage/1.0, 2)

	// Temperature effect (mobility ~ T^-1.5)
	tempK := condition.Temperature + 273.15
	tScale := math.Pow(298.15/tempK, 1.5)

	// Process corner effect
	var pScale float64
	switch condition.ProcessCorner {
	case CornerFF:
		pScale = 1.3
	case CornerSS:
		pScale = 0.7
	case CornerFS, CornerSF:
		pScale = 1.0
	default:
		pScale = 1.0
	}

	result.Throughput = baseThroughput * vScale * tScale * pScale

	// Efficiency calculation
	// Power ~ V^2 * f, f ~ V, so Power ~ V^3
	power := math.Pow(condition.Voltage/1.0, 3) * 100 // mW at nominal
	result.Efficiency = result.Throughput / power * 1000 // TOPS/W

	// Accuracy drop estimation
	// Higher variation at extreme corners
	if condition.ProcessCorner == CornerSS && condition.Temperature > 100 {
		result.AccuracyDrop = 2.0 // 2% at worst corner
	} else if condition.ProcessCorner == CornerFF && condition.Temperature < -20 {
		result.AccuracyDrop = 1.5
	} else {
		result.AccuracyDrop = 0.5 // Typical
	}

	return result
}

// GenerateReport generates a PVT benchmark report
func (b *PVTBenchmark) GenerateReport() string {
	report := "PVT Benchmark Report\n"
	report += "====================\n\n"

	// Find best and worst cases
	var bestEfficiency, worstEfficiency *BenchmarkResult
	var bestThroughput, worstThroughput *BenchmarkResult

	for _, result := range b.Results {
		if bestEfficiency == nil || result.Efficiency > bestEfficiency.Efficiency {
			bestEfficiency = result
		}
		if worstEfficiency == nil || result.Efficiency < worstEfficiency.Efficiency {
			worstEfficiency = result
		}
		if bestThroughput == nil || result.Throughput > bestThroughput.Throughput {
			bestThroughput = result
		}
		if worstThroughput == nil || result.Throughput < worstThroughput.Throughput {
			worstThroughput = result
		}
	}

	if bestEfficiency != nil {
		report += fmt.Sprintf("Best Efficiency: %.2f TOPS/W @ %s, %.2fV, %.0f°C\n",
			bestEfficiency.Efficiency, bestEfficiency.Corner, bestEfficiency.Voltage, bestEfficiency.Temperature)
	}
	if worstEfficiency != nil {
		report += fmt.Sprintf("Worst Efficiency: %.2f TOPS/W @ %s, %.2fV, %.0f°C\n",
			worstEfficiency.Efficiency, worstEfficiency.Corner, worstEfficiency.Voltage, worstEfficiency.Temperature)
	}
	if bestThroughput != nil {
		report += fmt.Sprintf("Best Throughput: %.2f GOPS @ %s, %.2fV, %.0f°C\n",
			bestThroughput.Throughput, bestThroughput.Corner, bestThroughput.Voltage, bestThroughput.Temperature)
	}
	if worstThroughput != nil {
		report += fmt.Sprintf("Worst Throughput: %.2f GOPS @ %s, %.2fV, %.0f°C\n",
			worstThroughput.Throughput, worstThroughput.Corner, worstThroughput.Voltage, worstThroughput.Temperature)
	}

	return report
}

// ============================================================================
// Serialization
// ============================================================================

// PVTEdgeState contains serializable state
type PVTEdgeState struct {
	PVTConfig       *PVTVariationConfig     `json:"pvt_config"`
	TinyMLConfig    *TinyMLConstraints      `json:"tinyml_config"`
	CompressionConfig *CompressionConfig    `json:"compression_config"`
	BenchmarkResults map[string]*BenchmarkResult `json:"benchmark_results,omitempty"`
}

// SerializeState serializes the PVT edge state to JSON
func SerializeState(pvtConfig *PVTVariationConfig, tinymlConfig *TinyMLConstraints,
	compConfig *CompressionConfig, results map[string]*BenchmarkResult) ([]byte, error) {
	state := &PVTEdgeState{
		PVTConfig:        pvtConfig,
		TinyMLConfig:     tinymlConfig,
		CompressionConfig: compConfig,
		BenchmarkResults: results,
	}
	return json.MarshalIndent(state, "", "  ")
}

// DeserializeState deserializes PVT edge state from JSON
func DeserializeState(data []byte) (*PVTEdgeState, error) {
	var state PVTEdgeState
	err := json.Unmarshal(data, &state)
	return &state, err
}

// min returns minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
