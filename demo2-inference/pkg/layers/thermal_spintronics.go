// Package layers provides thermal management and spintronics-based CIM simulation.
//
// Thermal Management Topics:
// - Hotspot modeling and temperature distribution
// - Joule heating in resistive crossbar arrays
// - Cell-to-cell thermal crosstalk
// - Temperature-aware scheduling and remapping
// - HR3AM-style bitwidth downgrading
// - 3D crossbar thermal considerations
//
// Spintronics CIM Topics:
// - Magnetic Tunnel Junction (MTJ) devices
// - STT-MRAM and SOT-MRAM crossbars
// - Domain wall synapses and neurons
// - Skyrmion-based neuromorphic computing
// - P-bit stochastic computing
// - Probabilistic neural networks
//
// Key findings:
// - Temperature sensitivity: 0.9% accuracy loss per 1K increase
// - HR3AM achieves up to 58% accuracy improvement via thermal optimization
// - SOT-MRAM CIM: 23.7-29.6 TOPS/W at 8-bit precision (28nm)
// - STT-MRAM: 22.4-41.5 TOPS/W (28nm 2Mb macro)
// - P-STT neurons: <1 pJ/bit, ~10ns switching
// - Skyrmion synapse: 64 states, 2.2ns write time
// - Domain wall MTJ: all-spin neuromorphic hardware
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// THERMAL MANAGEMENT FOR CIM CROSSBAR ARRAYS
// =============================================================================

// ThermalConfig configures thermal simulation parameters.
type ThermalConfig struct {
	// Array dimensions
	Rows    int
	Cols    int
	Layers  int // For 3D stacking

	// Material properties
	ThermalConductivity float64 // W/(m·K)
	SpecificHeat        float64 // J/(kg·K)
	Density             float64 // kg/m³

	// Geometry
	CellPitchNm   float64 // nm
	LayerSpacingNm float64 // nm (for 3D)

	// Boundary conditions
	AmbientTempK     float64 // K
	SubstrateThickUm float64 // μm
	EMCThickUm       float64 // Epoxy molding compound thickness

	// Operating parameters
	PulseDurationNs float64 // ns
	PulseCurrentUA  float64 // μA
	DutyCycle       float64 // 0-1

	// Thresholds
	MaxOperatingTempK   float64 // Max safe temperature
	CriticalTempK       float64 // Thermal runaway threshold
	AccuracyLossPerK    float64 // % accuracy loss per K (typ. 0.9%)
}

// DefaultThermalConfig returns typical ReRAM thermal parameters.
func DefaultThermalConfig() *ThermalConfig {
	return &ThermalConfig{
		Rows:               64,
		Cols:               64,
		Layers:             1,
		ThermalConductivity: 1.5, // HfO2 typical
		SpecificHeat:       500,
		Density:            9680, // HfO2
		CellPitchNm:        100,
		LayerSpacingNm:     500,
		AmbientTempK:       300,
		SubstrateThickUm:   500,
		EMCThickUm:         300,
		PulseDurationNs:    50,
		PulseCurrentUA:     100,
		DutyCycle:          0.1,
		MaxOperatingTempK:  358,    // 85°C
		CriticalTempK:      400,    // Thermal runaway risk
		AccuracyLossPerK:   0.009,  // 0.9% per K
	}
}

// ThermalCell represents thermal state of a single cell.
type ThermalCell struct {
	Row         int
	Col         int
	Layer       int
	Temperature float64   // Current temperature (K)
	PowerW      float64   // Instantaneous power dissipation
	Resistance  float64   // Current resistance state
	AccessCount int       // Number of accesses
	History     []float64 // Temperature history
}

// ThermalCrossbar simulates thermal behavior of CIM crossbar.
type ThermalCrossbar struct {
	Config *ThermalConfig
	Cells  [][][]*ThermalCell // [layer][row][col]

	// Thermal analysis
	HotspotTemp   float64
	HotspotLoc    [3]int // [layer, row, col]
	AverageTemp   float64
	ThermalMap    [][][]float64

	// Accumulated statistics
	TotalEnergy     float64
	PeakPower       float64
	ThermalCycles   int
	AccuracyLoss    float64
}

// NewThermalCrossbar creates a new thermal simulation.
func NewThermalCrossbar(config *ThermalConfig) *ThermalCrossbar {
	tc := &ThermalCrossbar{
		Config:     config,
		Cells:      make([][][]*ThermalCell, config.Layers),
		ThermalMap: make([][][]float64, config.Layers),
	}

	// Initialize cells
	for l := 0; l < config.Layers; l++ {
		tc.Cells[l] = make([][]*ThermalCell, config.Rows)
		tc.ThermalMap[l] = make([][]float64, config.Rows)
		for r := 0; r < config.Rows; r++ {
			tc.Cells[l][r] = make([]*ThermalCell, config.Cols)
			tc.ThermalMap[l][r] = make([]float64, config.Cols)
			for c := 0; c < config.Cols; c++ {
				tc.Cells[l][r][c] = &ThermalCell{
					Row:         r,
					Col:         c,
					Layer:       l,
					Temperature: config.AmbientTempK,
					Resistance:  100e3, // 100kΩ default
					History:     make([]float64, 0, 100),
				}
				tc.ThermalMap[l][r][c] = config.AmbientTempK
			}
		}
	}

	return tc
}

// CalculateJouleHeating computes power dissipation in a cell.
func (tc *ThermalCrossbar) CalculateJouleHeating(cell *ThermalCell, voltage float64) float64 {
	// P = V²/R
	power := (voltage * voltage) / cell.Resistance

	// Temperature-dependent resistance (positive coefficient for HfO2)
	// R(T) = R0 * (1 + α*(T - T0))
	alpha := 0.001 // Temperature coefficient
	tempFactor := 1 + alpha*(cell.Temperature-tc.Config.AmbientTempK)
	adjustedPower := power / tempFactor

	cell.PowerW = adjustedPower
	return adjustedPower
}

// SimulateThermalStep performs one thermal simulation timestep.
func (tc *ThermalCrossbar) SimulateThermalStep(dt float64) {
	cfg := tc.Config

	// Calculate thermal diffusivity
	diffusivity := cfg.ThermalConductivity / (cfg.Density * cfg.SpecificHeat)

	// Finite difference thermal simulation
	newTemps := make([][][]float64, cfg.Layers)
	for l := 0; l < cfg.Layers; l++ {
		newTemps[l] = make([][]float64, cfg.Rows)
		for r := 0; r < cfg.Rows; r++ {
			newTemps[l][r] = make([]float64, cfg.Cols)
		}
	}

	pitch := cfg.CellPitchNm * 1e-9 // Convert to meters

	for l := 0; l < cfg.Layers; l++ {
		for r := 0; r < cfg.Rows; r++ {
			for c := 0; c < cfg.Cols; c++ {
				cell := tc.Cells[l][r][c]
				T := cell.Temperature

				// Neighbor temperatures (with boundary conditions)
				var Tn, Ts, Te, Tw, Tu, Td float64

				if r > 0 {
					Tn = tc.Cells[l][r-1][c].Temperature
				} else {
					Tn = cfg.AmbientTempK // Boundary
				}
				if r < cfg.Rows-1 {
					Ts = tc.Cells[l][r+1][c].Temperature
				} else {
					Ts = cfg.AmbientTempK
				}
				if c > 0 {
					Tw = tc.Cells[l][r][c-1].Temperature
				} else {
					Tw = cfg.AmbientTempK
				}
				if c < cfg.Cols-1 {
					Te = tc.Cells[l][r][c+1].Temperature
				} else {
					Te = cfg.AmbientTempK
				}

				// 3D heat flow
				if cfg.Layers > 1 {
					layerPitch := cfg.LayerSpacingNm * 1e-9
					if l > 0 {
						Td = tc.Cells[l-1][r][c].Temperature
					} else {
						Td = cfg.AmbientTempK + 10 // Bottom layer hotter
					}
					if l < cfg.Layers-1 {
						Tu = tc.Cells[l+1][r][c].Temperature
					} else {
						Tu = cfg.AmbientTempK
					}

					// 3D Laplacian
					laplacian := (Tn+Ts+Te+Tw-4*T)/(pitch*pitch) +
						(Tu+Td-2*T)/(layerPitch*layerPitch)
					heatGen := cell.PowerW / (pitch * pitch * layerPitch * cfg.Density * cfg.SpecificHeat)
					newTemps[l][r][c] = T + dt*(diffusivity*laplacian+heatGen)
				} else {
					// 2D Laplacian
					laplacian := (Tn + Ts + Te + Tw - 4*T) / (pitch * pitch)
					heatGen := cell.PowerW / (pitch * pitch * pitch * cfg.Density * cfg.SpecificHeat)
					newTemps[l][r][c] = T + dt*(diffusivity*laplacian+heatGen)
				}

				// Cooling to substrate
				coolingRate := 0.001 * (T - cfg.AmbientTempK)
				newTemps[l][r][c] -= coolingRate * dt
			}
		}
	}

	// Update temperatures
	tc.HotspotTemp = cfg.AmbientTempK
	sumTemp := 0.0
	count := 0

	for l := 0; l < cfg.Layers; l++ {
		for r := 0; r < cfg.Rows; r++ {
			for c := 0; c < cfg.Cols; c++ {
				tc.Cells[l][r][c].Temperature = newTemps[l][r][c]
				tc.ThermalMap[l][r][c] = newTemps[l][r][c]

				if newTemps[l][r][c] > tc.HotspotTemp {
					tc.HotspotTemp = newTemps[l][r][c]
					tc.HotspotLoc = [3]int{l, r, c}
				}
				sumTemp += newTemps[l][r][c]
				count++
			}
		}
	}

	tc.AverageTemp = sumTemp / float64(count)
	tc.ThermalCycles++
}

// EstimateAccuracyLoss calculates accuracy degradation from temperature.
func (tc *ThermalCrossbar) EstimateAccuracyLoss() float64 {
	// 0.9% loss per K above ambient
	deltaT := tc.HotspotTemp - tc.Config.AmbientTempK
	tc.AccuracyLoss = tc.Config.AccuracyLossPerK * deltaT * 100
	return tc.AccuracyLoss
}

// =============================================================================
// THERMAL MITIGATION STRATEGIES
// =============================================================================

// ThermalMitigationConfig configures mitigation strategies.
type ThermalMitigationConfig struct {
	// HR3AM-style bitwidth downgrading
	EnableBitwidthDowngrade bool
	MaxBitwidth             int
	MinBitwidth             int
	TempThresholdDowngrade  float64 // K

	// Tile pairing for load balancing
	EnableTilePairing bool
	IdleTileRatio     float64 // Fraction of tiles kept idle

	// Allocation schemes
	AllocationScheme string // "naive", "strike", "chessboard"

	// Dynamic thermal management
	EnableDTM         bool
	ThrottleThreshold float64 // K
	ThrottleFactor    float64 // Clock reduction factor
}

// DefaultMitigationConfig returns typical mitigation settings.
func DefaultMitigationConfig() *ThermalMitigationConfig {
	return &ThermalMitigationConfig{
		EnableBitwidthDowngrade: true,
		MaxBitwidth:             8,
		MinBitwidth:             4,
		TempThresholdDowngrade:  340, // 67°C
		EnableTilePairing:       true,
		IdleTileRatio:           0.2,
		AllocationScheme:        "chessboard",
		EnableDTM:               true,
		ThrottleThreshold:       350, // 77°C
		ThrottleFactor:          0.5,
	}
}

// ThermalMitigator implements thermal mitigation strategies.
type ThermalMitigator struct {
	Config   *ThermalMitigationConfig
	Crossbar *ThermalCrossbar

	// State tracking
	CurrentBitwidths [][]int
	TilePairings     [][2]int
	ThrottledCells   [][]bool
	IdleTiles        []int
}

// NewThermalMitigator creates a thermal mitigation system.
func NewThermalMitigator(config *ThermalMitigationConfig, crossbar *ThermalCrossbar) *ThermalMitigator {
	rows := crossbar.Config.Rows
	cols := crossbar.Config.Cols

	tm := &ThermalMitigator{
		Config:           config,
		Crossbar:         crossbar,
		CurrentBitwidths: make([][]int, rows),
		ThrottledCells:   make([][]bool, rows),
	}

	for r := 0; r < rows; r++ {
		tm.CurrentBitwidths[r] = make([]int, cols)
		tm.ThrottledCells[r] = make([]bool, cols)
		for c := 0; c < cols; c++ {
			tm.CurrentBitwidths[r][c] = config.MaxBitwidth
		}
	}

	// Initialize tile pairings
	if config.EnableTilePairing {
		tm.initializeTilePairing()
	}

	return tm
}

// initializeTilePairing sets up hot-cold tile pairs.
func (tm *ThermalMitigator) initializeTilePairing() {
	// Simple pairing: bottom rows (hotter in 3D) with top rows
	rows := tm.Crossbar.Config.Rows
	numPairs := rows / 2

	tm.TilePairings = make([][2]int, numPairs)
	for i := 0; i < numPairs; i++ {
		tm.TilePairings[i] = [2]int{i, rows - 1 - i}
	}

	// Mark some tiles as idle for thermal buffering
	numIdle := int(float64(rows) * tm.Config.IdleTileRatio)
	tm.IdleTiles = make([]int, numIdle)
	for i := 0; i < numIdle; i++ {
		tm.IdleTiles[i] = rows/2 + i - numIdle/2
	}
}

// ApplyBitwidthDowngrading reduces precision for hot cells (HR3AM).
func (tm *ThermalMitigator) ApplyBitwidthDowngrading() {
	if !tm.Config.EnableBitwidthDowngrade {
		return
	}

	for r := 0; r < tm.Crossbar.Config.Rows; r++ {
		for c := 0; c < tm.Crossbar.Config.Cols; c++ {
			cell := tm.Crossbar.Cells[0][r][c]
			temp := cell.Temperature

			if temp > tm.Config.TempThresholdDowngrade {
				// Progressive downgrading based on temperature
				overTemp := temp - tm.Config.TempThresholdDowngrade
				reduction := int(overTemp / 10) // 1 bit per 10K above threshold

				newBitwidth := tm.Config.MaxBitwidth - reduction
				if newBitwidth < tm.Config.MinBitwidth {
					newBitwidth = tm.Config.MinBitwidth
				}
				tm.CurrentBitwidths[r][c] = newBitwidth
			} else {
				tm.CurrentBitwidths[r][c] = tm.Config.MaxBitwidth
			}
		}
	}
}

// ApplyDynamicThermalManagement throttles hot regions.
func (tm *ThermalMitigator) ApplyDynamicThermalManagement() float64 {
	if !tm.Config.EnableDTM {
		return 1.0
	}

	throttled := 0
	total := tm.Crossbar.Config.Rows * tm.Crossbar.Config.Cols

	for r := 0; r < tm.Crossbar.Config.Rows; r++ {
		for c := 0; c < tm.Crossbar.Config.Cols; c++ {
			cell := tm.Crossbar.Cells[0][r][c]
			if cell.Temperature > tm.Config.ThrottleThreshold {
				tm.ThrottledCells[r][c] = true
				throttled++
			} else {
				tm.ThrottledCells[r][c] = false
			}
		}
	}

	// Return effective clock factor
	throttledRatio := float64(throttled) / float64(total)
	return 1.0 - throttledRatio*(1-tm.Config.ThrottleFactor)
}

// GetAllocationPattern returns cell activation pattern for scheme.
func (tm *ThermalMitigator) GetAllocationPattern(step int) [][]bool {
	rows := tm.Crossbar.Config.Rows
	cols := tm.Crossbar.Config.Cols
	active := make([][]bool, rows)

	for r := 0; r < rows; r++ {
		active[r] = make([]bool, cols)
	}

	switch tm.Config.AllocationScheme {
	case "chessboard":
		// Alternating pattern reduces peak temperature
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				active[r][c] = (r+c+step)%2 == 0
			}
		}

	case "strike":
		// Stripe pattern
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				active[r][c] = (r+step)%2 == 0
			}
		}

	default: // "naive"
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				active[r][c] = true
			}
		}
	}

	return active
}

// =============================================================================
// SPINTRONICS-BASED CIM: MTJ DEVICES
// =============================================================================

// MTJState represents magnetic tunnel junction states.
type MTJState int

const (
	MTJParallel     MTJState = iota // Low resistance (P)
	MTJAntiParallel                 // High resistance (AP)
)

// MTJConfig configures MTJ device parameters.
type MTJConfig struct {
	// Device type
	DeviceType string // "STT", "SOT", "VCMA"

	// Resistance states
	ResistanceP  float64 // Parallel state (Ω)
	ResistanceAP float64 // Anti-parallel state (Ω)
	TMR          float64 // Tunnel magnetoresistance ratio

	// Switching parameters
	SwitchingCurrentUA float64 // Critical switching current
	SwitchingTimeNs    float64 // Typical switching time
	ThermalStability   float64 // Δ = E_b / k_B T

	// Stochastic properties
	SwitchingProbability float64 // At critical current
	ThermalFluctuations  bool

	// Energy
	SwitchingEnergyFJ float64 // fJ per switch

	// Process node
	DiameterNm float64
	ProcessNm  int
}

// DefaultSTTMRAMConfig returns typical STT-MRAM parameters.
func DefaultSTTMRAMConfig() *MTJConfig {
	return &MTJConfig{
		DeviceType:           "STT",
		ResistanceP:          5e3,   // 5 kΩ
		ResistanceAP:         10e3,  // 10 kΩ (TMR ~100%)
		TMR:                  1.0,   // 100%
		SwitchingCurrentUA:   50,
		SwitchingTimeNs:      10,
		ThermalStability:     60, // Typical for retention
		SwitchingProbability: 0.99,
		ThermalFluctuations:  true,
		SwitchingEnergyFJ:    100,
		DiameterNm:           40,
		ProcessNm:            28,
	}
}

// DefaultSOTMRAMConfig returns typical SOT-MRAM parameters.
func DefaultSOTMRAMConfig() *MTJConfig {
	return &MTJConfig{
		DeviceType:           "SOT",
		ResistanceP:          4e3,
		ResistanceAP:         10e3, // 150% TMR
		TMR:                  1.5,
		SwitchingCurrentUA:   200, // Higher write current
		SwitchingTimeNs:      2,   // 4× faster than STT
		ThermalStability:     50,
		SwitchingProbability: 0.999,
		ThermalFluctuations:  true,
		SwitchingEnergyFJ:    350,
		DiameterNm:           50,
		ProcessNm:            28,
	}
}

// MTJDevice represents a single magnetic tunnel junction.
type MTJDevice struct {
	Config      *MTJConfig
	State       MTJState
	Resistance  float64
	Temperature float64

	// Stochastic state for p-bits
	FluctuationRate float64 // Random telegraph noise frequency
	LastFlipTime    float64

	// Statistics
	WriteCount    int
	ReadCount     int
	TotalEnergy   float64
	WriteErrors   int
}

// NewMTJDevice creates a new MTJ device.
func NewMTJDevice(config *MTJConfig) *MTJDevice {
	mtj := &MTJDevice{
		Config:          config,
		State:           MTJParallel,
		Resistance:      config.ResistanceP,
		Temperature:     300,
		FluctuationRate: 1e6, // 1 MHz default
	}
	return mtj
}

// Read returns the resistance with noise.
func (mtj *MTJDevice) Read() float64 {
	mtj.ReadCount++

	// Add read noise (typically ~1-5%)
	noise := 1.0
	if mtj.Config.ThermalFluctuations {
		noise = 1.0 + (rand.Float64()-0.5)*0.02
	}

	return mtj.Resistance * noise
}

// Write attempts to switch MTJ state.
func (mtj *MTJDevice) Write(targetState MTJState, currentUA float64) bool {
	mtj.WriteCount++
	mtj.TotalEnergy += mtj.Config.SwitchingEnergyFJ * 1e-15

	if mtj.State == targetState {
		return true // Already in target state
	}

	// Stochastic switching model
	// P = 1 - exp(-t/τ * (I/I_c - 1))
	normalizedCurrent := currentUA / mtj.Config.SwitchingCurrentUA
	if normalizedCurrent < 1.0 {
		mtj.WriteErrors++
		return false // Below threshold
	}

	switchProb := 1.0 - math.Exp(-1.0*(normalizedCurrent-1.0))
	if mtj.Config.ThermalFluctuations {
		switchProb *= mtj.Config.SwitchingProbability
	}

	if rand.Float64() < switchProb {
		mtj.State = targetState
		if targetState == MTJParallel {
			mtj.Resistance = mtj.Config.ResistanceP
		} else {
			mtj.Resistance = mtj.Config.ResistanceAP
		}
		return true
	}

	mtj.WriteErrors++
	return false
}

// GetPBitOutput returns stochastic p-bit output (0 or 1).
func (mtj *MTJDevice) GetPBitOutput(biasVoltage float64) int {
	// P-bit: Low barrier magnet with controllable switching probability
	// Sigmoid activation based on bias
	probability := 1.0 / (1.0 + math.Exp(-biasVoltage*10))

	if rand.Float64() < probability {
		return 1
	}
	return 0
}

// =============================================================================
// MRAM CROSSBAR ARRAY
// =============================================================================

// MRAMCrossbarConfig configures MRAM-based CIM crossbar.
type MRAMCrossbarConfig struct {
	Rows     int
	Cols     int
	MTJType  string // "STT", "SOT"
	BitWidth int

	// Architecture
	Architecture string // "1T1MTJ", "2T1MTJ" (SOT)

	// Compute precision
	InputBits  int
	OutputBits int
	ADCBits    int

	// Reference
	ReferenceResistance float64
}

// DefaultMRAMCrossbarConfig returns typical MRAM CIM settings.
func DefaultMRAMCrossbarConfig() *MRAMCrossbarConfig {
	return &MRAMCrossbarConfig{
		Rows:                64,
		Cols:                64,
		MTJType:             "SOT",
		BitWidth:            1, // Binary weights typical
		Architecture:        "2T1MTJ",
		InputBits:           8,
		OutputBits:          8,
		ADCBits:             6,
		ReferenceResistance: 7.5e3, // Midpoint
	}
}

// MRAMCrossbar implements MRAM-based CIM array.
type MRAMCrossbar struct {
	Config  *MRAMCrossbarConfig
	Devices [][]*MTJDevice

	// Weight storage
	Weights [][]float64

	// Performance metrics
	TotalMACs     int
	TotalEnergy   float64
	EnergyPerMAC  float64
	Throughput    float64 // TOPS
	WriteLatency  float64 // ns
	ReadLatency   float64 // ns
}

// NewMRAMCrossbar creates a new MRAM crossbar array.
func NewMRAMCrossbar(config *MRAMCrossbarConfig) *MRAMCrossbar {
	var mtjConfig *MTJConfig
	if config.MTJType == "SOT" {
		mtjConfig = DefaultSOTMRAMConfig()
	} else {
		mtjConfig = DefaultSTTMRAMConfig()
	}

	mc := &MRAMCrossbar{
		Config:  config,
		Devices: make([][]*MTJDevice, config.Rows),
		Weights: make([][]float64, config.Rows),
	}

	for r := 0; r < config.Rows; r++ {
		mc.Devices[r] = make([]*MTJDevice, config.Cols)
		mc.Weights[r] = make([]float64, config.Cols)
		for c := 0; c < config.Cols; c++ {
			mc.Devices[r][c] = NewMTJDevice(mtjConfig)
		}
	}

	// Calculate performance metrics
	if config.MTJType == "SOT" {
		mc.WriteLatency = 2   // ns
		mc.ReadLatency = 1    // ns
		mc.EnergyPerMAC = 0.3 // pJ (23.7-29.6 TOPS/W @ 8b)
	} else {
		mc.WriteLatency = 10
		mc.ReadLatency = 2
		mc.EnergyPerMAC = 0.5
	}

	return mc
}

// ProgramWeights writes weight matrix to MRAM.
func (mc *MRAMCrossbar) ProgramWeights(weights [][]float64) error {
	if len(weights) != mc.Config.Rows || len(weights[0]) != mc.Config.Cols {
		return fmt.Errorf("weight dimensions mismatch")
	}

	for r := 0; r < mc.Config.Rows; r++ {
		for c := 0; c < mc.Config.Cols; c++ {
			mc.Weights[r][c] = weights[r][c]

			// Binary: threshold at 0
			var targetState MTJState
			if weights[r][c] >= 0 {
				targetState = MTJParallel // Low R = positive
			} else {
				targetState = MTJAntiParallel
			}

			currentUA := 100.0 // Use sufficient current
			if mc.Config.MTJType == "SOT" {
				currentUA = 250.0
			}

			mc.Devices[r][c].Write(targetState, currentUA)
		}
	}

	return nil
}

// ComputeMVM performs matrix-vector multiplication.
func (mc *MRAMCrossbar) ComputeMVM(input []float64) ([]float64, error) {
	if len(input) != mc.Config.Rows {
		return nil, fmt.Errorf("input dimension mismatch")
	}

	output := make([]float64, mc.Config.Cols)

	// Analog MVM via current summation
	for c := 0; c < mc.Config.Cols; c++ {
		sum := 0.0
		for r := 0; r < mc.Config.Rows; r++ {
			// Read conductance and compute
			resistance := mc.Devices[r][c].Read()
			conductance := 1.0 / resistance
			sum += input[r] * conductance
		}
		output[c] = sum

		mc.TotalMACs += mc.Config.Rows
	}

	// ADC quantization
	for i := range output {
		// Scale and quantize
		levels := float64(int(1) << mc.Config.ADCBits)
		output[i] = math.Round(output[i]*levels) / levels
	}

	mc.TotalEnergy += mc.EnergyPerMAC * float64(mc.Config.Rows*mc.Config.Cols) * 1e-12

	return output, nil
}

// GetEfficiency returns TOPS/W metric.
func (mc *MRAMCrossbar) GetEfficiency() float64 {
	// TOPS/W = (MACs per second * 2) / Power
	// Typical: 22.4-41.5 TOPS/W for STT, 23.7-29.6 for SOT
	if mc.Config.MTJType == "SOT" {
		return 26.0 // TOPS/W average
	}
	return 32.0 // TOPS/W average for STT
}

// =============================================================================
// DOMAIN WALL SYNAPSES AND NEURONS
// =============================================================================

// DomainWallConfig configures domain wall device.
type DomainWallConfig struct {
	// Track parameters
	TrackLengthNm float64
	TrackWidthNm  float64
	ThicknessNm   float64

	// Material: CoFeB/MgO typical
	DMIStrength    float64 // Dzyaloshinskii-Moriya interaction
	SOTEfficiency  float64 // Spin-orbit torque efficiency

	// Synaptic states
	NumStates       int     // Number of analog levels
	StateResolution float64 // nm per state

	// Switching
	DWVelocity     float64 // m/s at 1 MA/cm²
	SwitchingTimeNs float64
}

// DefaultDomainWallConfig returns typical DW synapse parameters.
func DefaultDomainWallConfig() *DomainWallConfig {
	return &DomainWallConfig{
		TrackLengthNm:  500,
		TrackWidthNm:   80,
		ThicknessNm:    1.5,
		DMIStrength:    1.5, // mJ/m²
		SOTEfficiency:  0.3,
		NumStates:      16,
		StateResolution: 500.0 / 16, // ~31 nm per state
		DWVelocity:     100,         // m/s
		SwitchingTimeNs: 10,
	}
}

// DomainWallSynapse implements a domain wall-based synapse.
type DomainWallSynapse struct {
	Config *DomainWallConfig

	// Domain wall position (0 = left, TrackLength = right)
	DWPosition float64

	// Resistance based on position
	MinResistance float64
	MaxResistance float64

	// Statistics
	PulseCount int
	TotalDrift float64
}

// NewDomainWallSynapse creates a DW synapse.
func NewDomainWallSynapse(config *DomainWallConfig) *DomainWallSynapse {
	return &DomainWallSynapse{
		Config:        config,
		DWPosition:    config.TrackLengthNm / 2, // Start at center
		MinResistance: 5e3,
		MaxResistance: 15e3,
	}
}

// GetWeight returns current synaptic weight (0-1).
func (dw *DomainWallSynapse) GetWeight() float64 {
	return dw.DWPosition / dw.Config.TrackLengthNm
}

// GetResistance returns current resistance state.
func (dw *DomainWallSynapse) GetResistance() float64 {
	weight := dw.GetWeight()
	return dw.MinResistance + weight*(dw.MaxResistance-dw.MinResistance)
}

// ApplyPulse moves domain wall with current pulse.
func (dw *DomainWallSynapse) ApplyPulse(currentMA float64, durationNs float64, direction int) {
	dw.PulseCount++

	// Domain wall velocity: v ∝ J (current density)
	velocity := dw.Config.DWVelocity * currentMA // Simplified linear model

	// Distance moved
	distance := velocity * (durationNs * 1e-9) * 1e9 // Convert to nm

	// Apply movement
	dw.DWPosition += float64(direction) * distance

	// Clamp to track bounds
	if dw.DWPosition < 0 {
		dw.DWPosition = 0
	}
	if dw.DWPosition > dw.Config.TrackLengthNm {
		dw.DWPosition = dw.Config.TrackLengthNm
	}
}

// Potentiate increases synaptic weight (LTP).
func (dw *DomainWallSynapse) Potentiate(strength float64) {
	dw.ApplyPulse(strength, dw.Config.SwitchingTimeNs, 1)
}

// Depress decreases synaptic weight (LTD).
func (dw *DomainWallSynapse) Depress(strength float64) {
	dw.ApplyPulse(strength, dw.Config.SwitchingTimeNs, -1)
}

// =============================================================================
// SKYRMION NEUROMORPHIC COMPUTING
// =============================================================================

// SkyrmionConfig configures skyrmion-based device.
type SkyrmionConfig struct {
	// Track dimensions
	TrackLengthNm float64
	TrackWidthNm  float64

	// Skyrmion properties
	SkyrmionRadiusNm float64
	MaxSkyrmions     int

	// Dynamics
	VelocityMperS     float64 // at 1e11 A/m² current
	CreationEnergyFJ  float64
	AnnihilationEnergyFJ float64

	// Synapse parameters
	WriteTimeNs float64
}

// DefaultSkyrmionConfig returns typical skyrmion parameters.
func DefaultSkyrmionConfig() *SkyrmionConfig {
	return &SkyrmionConfig{
		TrackLengthNm:        500,
		TrackWidthNm:         100,
		SkyrmionRadiusNm:     10,
		MaxSkyrmions:         64,
		VelocityMperS:        100,
		CreationEnergyFJ:     10,
		AnnihilationEnergyFJ: 5,
		WriteTimeNs:          2.2, // From literature
	}
}

// SkyrmionSynapse implements skyrmion-based synapse.
type SkyrmionSynapse struct {
	Config *SkyrmionConfig

	// Skyrmion count represents weight
	NumSkyrmions int

	// Position tracking for weighted sum
	Positions []float64 // Position of each skyrmion

	// Statistics
	CreationCount     int
	AnnihilationCount int
	TotalEnergy       float64
}

// NewSkyrmionSynapse creates a skyrmion synapse.
func NewSkyrmionSynapse(config *SkyrmionConfig) *SkyrmionSynapse {
	return &SkyrmionSynapse{
		Config:    config,
		Positions: make([]float64, 0, config.MaxSkyrmions),
	}
}

// GetWeight returns normalized weight (0-1).
func (ss *SkyrmionSynapse) GetWeight() float64 {
	return float64(ss.NumSkyrmions) / float64(ss.Config.MaxSkyrmions)
}

// CreateSkyrmion nucleates a new skyrmion.
func (ss *SkyrmionSynapse) CreateSkyrmion() bool {
	if ss.NumSkyrmions >= ss.Config.MaxSkyrmions {
		return false
	}

	ss.NumSkyrmions++
	ss.Positions = append(ss.Positions, 0) // Created at track start
	ss.CreationCount++
	ss.TotalEnergy += ss.Config.CreationEnergyFJ * 1e-15

	return true
}

// AnnihilateSkyrmion removes a skyrmion.
func (ss *SkyrmionSynapse) AnnihilateSkyrmion() bool {
	if ss.NumSkyrmions <= 0 {
		return false
	}

	ss.NumSkyrmions--
	if len(ss.Positions) > 0 {
		ss.Positions = ss.Positions[:len(ss.Positions)-1]
	}
	ss.AnnihilationCount++
	ss.TotalEnergy += ss.Config.AnnihilationEnergyFJ * 1e-15

	return true
}

// SetWeight programs target number of skyrmions.
func (ss *SkyrmionSynapse) SetWeight(targetWeight float64) {
	targetCount := int(targetWeight * float64(ss.Config.MaxSkyrmions))

	for ss.NumSkyrmions < targetCount {
		ss.CreateSkyrmion()
	}
	for ss.NumSkyrmions > targetCount {
		ss.AnnihilateSkyrmion()
	}
}

// SkyrmionWeightedSum performs neuromorphic weighted sum.
type SkyrmionWeightedSum struct {
	Synapses []*SkyrmionSynapse
	NumInputs int
}

// NewSkyrmionWeightedSum creates weighted sum circuit.
func NewSkyrmionWeightedSum(numInputs int) *SkyrmionWeightedSum {
	sws := &SkyrmionWeightedSum{
		Synapses:  make([]*SkyrmionSynapse, numInputs),
		NumInputs: numInputs,
	}

	config := DefaultSkyrmionConfig()
	for i := 0; i < numInputs; i++ {
		sws.Synapses[i] = NewSkyrmionSynapse(config)
	}

	return sws
}

// Compute performs weighted sum: Σ(input[i] * weight[i]).
func (sws *SkyrmionWeightedSum) Compute(inputs []float64) float64 {
	if len(inputs) != sws.NumInputs {
		return 0
	}

	sum := 0.0
	for i := 0; i < sws.NumInputs; i++ {
		weight := sws.Synapses[i].GetWeight()
		sum += inputs[i] * weight
	}

	return sum
}

// =============================================================================
// P-BIT STOCHASTIC COMPUTING
// =============================================================================

// PBitConfig configures p-bit device.
type PBitConfig struct {
	// Device parameters
	DeviceType string // "LBM" (low barrier magnet), "SOT-MTJ"

	// Fluctuation properties
	FluctuationRateHz float64 // Random telegraph noise frequency
	ThermalStability  float64 // Δ (low for fast fluctuation)

	// Control
	BiasRange   float64 // Voltage range for probability control
	Sensitivity float64 // Probability change per mV
}

// DefaultPBitConfig returns typical p-bit parameters.
func DefaultPBitConfig() *PBitConfig {
	return &PBitConfig{
		DeviceType:        "SOT-MTJ",
		FluctuationRateHz: 1e9, // 1 GHz for fast p-bits
		ThermalStability:  1.0, // Very low for stochasticity
		BiasRange:         100, // mV
		Sensitivity:       0.01,
	}
}

// PBit represents a probabilistic bit.
type PBit struct {
	Config *PBitConfig

	// Current state
	State int // 0 or 1

	// Bias input
	Bias float64

	// Statistics
	FlipCount   int
	SampleCount int
}

// NewPBit creates a p-bit device.
func NewPBit(config *PBitConfig) *PBit {
	return &PBit{
		Config: config,
		State:  0,
	}
}

// SetBias sets the input bias (controls probability).
func (pb *PBit) SetBias(bias float64) {
	pb.Bias = bias
}

// Sample returns stochastic output with current bias.
func (pb *PBit) Sample() int {
	pb.SampleCount++

	// Sigmoid probability: P(1) = 1 / (1 + exp(-β * bias))
	beta := pb.Config.Sensitivity * 100
	prob := 1.0 / (1.0 + math.Exp(-beta*pb.Bias))

	if rand.Float64() < prob {
		if pb.State == 0 {
			pb.FlipCount++
		}
		pb.State = 1
	} else {
		if pb.State == 1 {
			pb.FlipCount++
		}
		pb.State = 0
	}

	return pb.State
}

// PBitNetwork implements network of coupled p-bits.
type PBitNetwork struct {
	PBits    []*PBit
	Weights  [][]float64 // Coupling weights
	NumBits  int
	Sync     bool // Synchronous updates
}

// NewPBitNetwork creates a p-bit network.
func NewPBitNetwork(numBits int) *PBitNetwork {
	config := DefaultPBitConfig()

	pbn := &PBitNetwork{
		PBits:   make([]*PBit, numBits),
		Weights: make([][]float64, numBits),
		NumBits: numBits,
		Sync:    false,
	}

	for i := 0; i < numBits; i++ {
		pbn.PBits[i] = NewPBit(config)
		pbn.Weights[i] = make([]float64, numBits)
	}

	return pbn
}

// SetWeights configures coupling matrix (Ising/Boltzmann).
func (pbn *PBitNetwork) SetWeights(weights [][]float64) {
	for i := 0; i < pbn.NumBits; i++ {
		for j := 0; j < pbn.NumBits; j++ {
			pbn.Weights[i][j] = weights[i][j]
		}
	}
}

// Update performs one network update.
func (pbn *PBitNetwork) Update() {
	// Calculate effective fields
	fields := make([]float64, pbn.NumBits)

	for i := 0; i < pbn.NumBits; i++ {
		h := 0.0
		for j := 0; j < pbn.NumBits; j++ {
			state := float64(pbn.PBits[j].State)*2 - 1 // Convert to ±1
			h += pbn.Weights[i][j] * state
		}
		fields[i] = h
	}

	// Update p-bits
	if pbn.Sync {
		// Synchronous update
		for i := 0; i < pbn.NumBits; i++ {
			pbn.PBits[i].SetBias(fields[i])
		}
		for i := 0; i < pbn.NumBits; i++ {
			pbn.PBits[i].Sample()
		}
	} else {
		// Asynchronous (random order)
		order := rand.Perm(pbn.NumBits)
		for _, i := range order {
			pbn.PBits[i].SetBias(fields[i])
			pbn.PBits[i].Sample()
		}
	}
}

// RunBoltzmannSampling performs Boltzmann machine sampling.
func (pbn *PBitNetwork) RunBoltzmannSampling(iterations int) [][]int {
	samples := make([][]int, iterations)

	for iter := 0; iter < iterations; iter++ {
		pbn.Update()

		samples[iter] = make([]int, pbn.NumBits)
		for i := 0; i < pbn.NumBits; i++ {
			samples[iter][i] = pbn.PBits[i].State
		}
	}

	return samples
}

// =============================================================================
// INTEGRATION: THERMALLY-AWARE SPINTRONICS CIM
// =============================================================================

// SpintronicsCIMConfig configures integrated system.
type SpintronicsCIMConfig struct {
	// Array configuration
	Rows    int
	Cols    int
	MTJType string

	// Thermal parameters
	EnableThermalModel bool
	ThermalConfig      *ThermalConfig

	// Stochastic features
	EnablePBits      bool
	StochasticLayers []int // Which layers use p-bits
}

// SpintronicsCIMSystem integrates spintronics with thermal management.
type SpintronicsCIMSystem struct {
	Config *SpintronicsCIMConfig

	// MRAM crossbar
	MRAMArray *MRAMCrossbar

	// Thermal model
	ThermalModel    *ThermalCrossbar
	ThermalMitigator *ThermalMitigator

	// P-bit networks for stochastic layers
	PBitNetworks []*PBitNetwork

	// Performance
	TotalInferences int
	TotalEnergy     float64
	AverageAccuracy float64
}

// NewSpintronicsCIMSystem creates an integrated system.
func NewSpintronicsCIMSystem(config *SpintronicsCIMConfig) *SpintronicsCIMSystem {
	system := &SpintronicsCIMSystem{
		Config: config,
	}

	// Initialize MRAM crossbar
	mramConfig := &MRAMCrossbarConfig{
		Rows:       config.Rows,
		Cols:       config.Cols,
		MTJType:    config.MTJType,
		BitWidth:   1,
		InputBits:  8,
		OutputBits: 8,
		ADCBits:    6,
	}
	system.MRAMArray = NewMRAMCrossbar(mramConfig)

	// Initialize thermal model
	if config.EnableThermalModel {
		if config.ThermalConfig == nil {
			config.ThermalConfig = DefaultThermalConfig()
			config.ThermalConfig.Rows = config.Rows
			config.ThermalConfig.Cols = config.Cols
		}
		system.ThermalModel = NewThermalCrossbar(config.ThermalConfig)
		system.ThermalMitigator = NewThermalMitigator(
			DefaultMitigationConfig(),
			system.ThermalModel,
		)
	}

	// Initialize p-bit networks for stochastic layers
	if config.EnablePBits {
		system.PBitNetworks = make([]*PBitNetwork, len(config.StochasticLayers))
		for i, size := range config.StochasticLayers {
			system.PBitNetworks[i] = NewPBitNetwork(size)
		}
	}

	return system
}

// RunInference performs thermally-aware inference.
func (s *SpintronicsCIMSystem) RunInference(input []float64) ([]float64, error) {
	s.TotalInferences++

	// Thermal simulation step
	if s.Config.EnableThermalModel {
		// Apply current heating
		for r := 0; r < s.Config.Rows; r++ {
			for c := 0; c < s.Config.Cols; c++ {
				s.ThermalModel.CalculateJouleHeating(
					s.ThermalModel.Cells[0][r][c],
					0.5, // Typical read voltage
				)
			}
		}

		// Simulate thermal diffusion
		s.ThermalModel.SimulateThermalStep(1e-9) // 1ns step

		// Apply mitigation
		s.ThermalMitigator.ApplyBitwidthDowngrading()
		clockFactor := s.ThermalMitigator.ApplyDynamicThermalManagement()

		// Adjust for throttling
		if clockFactor < 1.0 {
			// Would slow down computation
		}
	}

	// Run MVM on MRAM
	output, err := s.MRAMArray.ComputeMVM(input)
	if err != nil {
		return nil, err
	}

	s.TotalEnergy += s.MRAMArray.EnergyPerMAC * float64(len(input)*len(output)) * 1e-12

	return output, nil
}

// GetThermalStatus returns current thermal state.
func (s *SpintronicsCIMSystem) GetThermalStatus() map[string]float64 {
	if s.ThermalModel == nil {
		return nil
	}

	return map[string]float64{
		"hotspot_temp_K":    s.ThermalModel.HotspotTemp,
		"average_temp_K":    s.ThermalModel.AverageTemp,
		"accuracy_loss_pct": s.ThermalModel.EstimateAccuracyLoss(),
		"hotspot_row":       float64(s.ThermalModel.HotspotLoc[1]),
		"hotspot_col":       float64(s.ThermalModel.HotspotLoc[2]),
	}
}

// =============================================================================
// BENCHMARK AND ANALYSIS UTILITIES
// =============================================================================

// ThermalBenchmark benchmarks thermal behavior.
type ThermalBenchmark struct {
	Results []ThermalBenchmarkResult
}

// ThermalBenchmarkResult stores benchmark data.
type ThermalBenchmarkResult struct {
	Scheme         string
	PeakTempK      float64
	AverageTempK   float64
	AccuracyLoss   float64
	ThermalCycles  int
}

// RunThermalBenchmark compares allocation schemes.
func RunThermalBenchmark(config *ThermalConfig, iterations int) *ThermalBenchmark {
	benchmark := &ThermalBenchmark{}
	schemes := []string{"naive", "strike", "chessboard"}

	for _, scheme := range schemes {
		crossbar := NewThermalCrossbar(config)
		mitigatorConfig := DefaultMitigationConfig()
		mitigatorConfig.AllocationScheme = scheme
		mitigator := NewThermalMitigator(mitigatorConfig, crossbar)

		// Run simulation
		for step := 0; step < iterations; step++ {
			pattern := mitigator.GetAllocationPattern(step)

			// Apply heat to active cells
			for r := 0; r < config.Rows; r++ {
				for c := 0; c < config.Cols; c++ {
					if pattern[r][c] {
						crossbar.CalculateJouleHeating(crossbar.Cells[0][r][c], 0.8)
					}
				}
			}

			crossbar.SimulateThermalStep(1e-8) // 10ns step
		}

		benchmark.Results = append(benchmark.Results, ThermalBenchmarkResult{
			Scheme:        scheme,
			PeakTempK:     crossbar.HotspotTemp,
			AverageTempK:  crossbar.AverageTemp,
			AccuracyLoss:  crossbar.EstimateAccuracyLoss(),
			ThermalCycles: crossbar.ThermalCycles,
		})
	}

	return benchmark
}

// SpintronicsBenchmark benchmarks MRAM technologies.
type SpintronicsBenchmark struct {
	Results []SpintronicsBenchmarkResult
}

// SpintronicsBenchmarkResult stores spintronics benchmark data.
type SpintronicsBenchmarkResult struct {
	Technology    string
	Rows          int
	Cols          int
	EnergyPerMAC  float64 // pJ
	Throughput    float64 // TOPS
	TOPSW         float64 // TOPS/W
	WriteLatency  float64 // ns
	ReadLatency   float64 // ns
	TMR           float64 // %
	WriteErrors   int
}

// RunSpintronicsBenchmark compares MRAM technologies.
func RunSpintronicsBenchmark(rows, cols, iterations int) *SpintronicsBenchmark {
	benchmark := &SpintronicsBenchmark{}

	technologies := []string{"STT", "SOT"}

	for _, tech := range technologies {
		config := &MRAMCrossbarConfig{
			Rows:       rows,
			Cols:       cols,
			MTJType:    tech,
			BitWidth:   1,
			InputBits:  8,
			OutputBits: 8,
			ADCBits:    6,
		}

		crossbar := NewMRAMCrossbar(config)

		// Program random weights
		weights := make([][]float64, rows)
		for r := 0; r < rows; r++ {
			weights[r] = make([]float64, cols)
			for c := 0; c < cols; c++ {
				weights[r][c] = rand.Float64()*2 - 1
			}
		}
		crossbar.ProgramWeights(weights)

		// Run inference iterations
		for i := 0; i < iterations; i++ {
			input := make([]float64, rows)
			for j := range input {
				input[j] = rand.Float64()
			}
			crossbar.ComputeMVM(input)
		}

		// Count write errors
		totalWriteErrors := 0
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				totalWriteErrors += crossbar.Devices[r][c].WriteErrors
			}
		}

		var mtjConfig *MTJConfig
		if tech == "SOT" {
			mtjConfig = DefaultSOTMRAMConfig()
		} else {
			mtjConfig = DefaultSTTMRAMConfig()
		}

		benchmark.Results = append(benchmark.Results, SpintronicsBenchmarkResult{
			Technology:   tech,
			Rows:         rows,
			Cols:         cols,
			EnergyPerMAC: crossbar.EnergyPerMAC,
			TOPSW:        crossbar.GetEfficiency(),
			WriteLatency: crossbar.WriteLatency,
			ReadLatency:  crossbar.ReadLatency,
			TMR:          mtjConfig.TMR * 100,
			WriteErrors:  totalWriteErrors,
		})
	}

	return benchmark
}

// PrintThermalBenchmark outputs thermal benchmark results.
func (b *ThermalBenchmark) Print() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║        THERMAL ALLOCATION SCHEME BENCHMARK RESULTS             ║")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-12s │ %8s │ %8s │ %10s │ %6s ║\n",
		"Scheme", "Peak(K)", "Avg(K)", "AccLoss(%)", "Cycles")
	fmt.Println("╠════════════════════════════════════════════════════════════════╣")

	for _, r := range b.Results {
		fmt.Printf("║ %-12s │ %8.1f │ %8.1f │ %10.2f │ %6d ║\n",
			r.Scheme, r.PeakTempK, r.AverageTempK, r.AccuracyLoss, r.ThermalCycles)
	}
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
}

// PrintSpintronicsBenchmark outputs spintronics benchmark results.
func (b *SpintronicsBenchmark) Print() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              SPINTRONICS MRAM CIM BENCHMARK RESULTS                       ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║ %-6s │ %5s │ %9s │ %8s │ %8s │ %6s │ %6s ║\n",
		"Tech", "Size", "E/MAC(pJ)", "TOPS/W", "TMR(%)", "WrLat", "Errors")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════╣")

	for _, r := range b.Results {
		size := fmt.Sprintf("%dx%d", r.Rows, r.Cols)
		fmt.Printf("║ %-6s │ %5s │ %9.3f │ %8.1f │ %8.1f │ %6.1f │ %6d ║\n",
			r.Technology, size, r.EnergyPerMAC, r.TOPSW, r.TMR, r.WriteLatency, r.WriteErrors)
	}
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════╝")
}

// AnalyzeHotspotDistribution finds hotspot patterns.
func AnalyzeHotspotDistribution(crossbar *ThermalCrossbar) map[string]interface{} {
	temps := make([]float64, 0, crossbar.Config.Rows*crossbar.Config.Cols)

	for r := 0; r < crossbar.Config.Rows; r++ {
		for c := 0; c < crossbar.Config.Cols; c++ {
			temps = append(temps, crossbar.Cells[0][r][c].Temperature)
		}
	}

	sort.Float64s(temps)

	n := len(temps)
	return map[string]interface{}{
		"min":         temps[0],
		"max":         temps[n-1],
		"median":      temps[n/2],
		"p90":         temps[int(float64(n)*0.9)],
		"p99":         temps[int(float64(n)*0.99)],
		"range":       temps[n-1] - temps[0],
		"hotspot_loc": crossbar.HotspotLoc,
	}
}
