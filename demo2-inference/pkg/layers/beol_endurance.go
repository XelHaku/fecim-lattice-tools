// Package layers provides BEOL (Back-End-Of-Line) ferroelectric integration
// and endurance/wake-up optimization for IronLattice CIM architectures.
//
// This module simulates:
// - BEOL-compatible FeFET fabrication with low thermal budget (<400°C)
// - Monolithic 3D integration for CIM accelerators
// - Wake-up free ferroelectric materials (superlattice structures)
// - Ultra-high endurance (10^12 cycles) through defect engineering
// - Fatigue recovery mechanisms in HfO2-ZrO2 systems
//
// Based on research from IEEE EDTM 2024, Nature Communications 2024-2025,
// and advanced HfO2/ZrO2 superlattice studies.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// BEOL INTEGRATION CONFIGURATION
// ============================================================================

// BEOLConfig defines parameters for back-end-of-line FeFET fabrication
type BEOLConfig struct {
	// Thermal budget constraints
	MaxTemperatureC     float64 // Maximum processing temperature (typically <400-500°C)
	AnnealingTempC      float64 // Ferroelectric crystallization temperature
	AnnealingTimeSec    float64 // Crystallization anneal duration
	UnidirectionalAnneal bool   // Heating from Pt/ZrO2 interface side

	// Channel material selection
	ChannelMaterial  string  // "IGZO", "ITO", "PolySi", "a-Si"
	ChannelThicknessnm float64 // Channel thickness
	ChannelMobilityCm2Vs float64 // Carrier mobility

	// Ferroelectric layer
	FEMaterial       string  // "HZO", "ZrO2-HfO2-SL", "Y:HfO2"
	FEThicknessnm    float64 // Ferroelectric thickness
	SuperlatticePeriod int   // Number of HfO2/ZrO2 bilayers (0 for solid solution)

	// Integration level
	MetalLayers      int     // Number of metal layers in BEOL stack
	InterlayerDielectric string // "SiO2", "Low-k", "Air-gap"
}

// DefaultBEOLConfig returns optimized BEOL configuration for CIM
func DefaultBEOLConfig() *BEOLConfig {
	return &BEOLConfig{
		MaxTemperatureC:      400,
		AnnealingTempC:       350,
		AnnealingTimeSec:     60,
		UnidirectionalAnneal: true,
		ChannelMaterial:      "IGZO",
		ChannelThicknessnm:   10,
		ChannelMobilityCm2Vs: 15.0,
		FEMaterial:           "ZrO2-HfO2-SL",
		FEThicknessnm:        10,
		SuperlatticePeriod:   5,
		MetalLayers:          8,
		InterlayerDielectric: "Low-k",
	}
}

// EnduranceConfig defines parameters for fatigue and cycling simulation
type EnduranceConfig struct {
	// Target performance
	TargetCycles    float64 // Target endurance (e.g., 1e12)
	PrInitialuCcm2  float64 // Initial remnant polarization
	PrMinimumRatio  float64 // Minimum Pr/Pr0 before failure (e.g., 0.9)

	// Fatigue mechanisms
	OxygenVacancyDensity float64 // Initial [Vo] per cm^3
	TrapDensityCm3       float64 // Interface trap density
	FilamentThreshold    float64 // Critical vacancy density for filament

	// Recovery parameters
	RecoveryTimeSec   float64 // Time for fatigue recovery
	RecoveryEfficiency float64 // Fraction of polarization recovered

	// Superlattice benefits
	DefectReduction   float64 // Factor by which SL reduces defects vs HZO
	InterfaceBlocking float64 // Filament disruption factor
}

// DefaultEnduranceConfig returns high-endurance configuration
func DefaultEnduranceConfig() *EnduranceConfig {
	return &EnduranceConfig{
		TargetCycles:        1e12,
		PrInitialuCcm2:      27.4,
		PrMinimumRatio:      0.9,
		OxygenVacancyDensity: 1e18,
		TrapDensityCm3:       1e11,
		FilamentThreshold:    1e20,
		RecoveryTimeSec:      30,
		RecoveryEfficiency:   0.95,
		DefectReduction:      0.1,
		InterfaceBlocking:    10.0,
	}
}

// ============================================================================
// BEOL-COMPATIBLE FeFET DEVICE
// ============================================================================

// BEOLFeFET represents a back-end-of-line compatible ferroelectric FET
type BEOLFeFET struct {
	Config *BEOLConfig

	// Device characteristics
	MemoryWindowV     float64 // Memory window (V)
	OnOffRatio        float64 // ION/IOFF ratio
	ThresholdVoltageV float64 // Vth
	SubthresholdSwingmV float64 // SS (mV/dec)

	// Channel properties
	ChannelConductance float64 // G (S)
	InterfaceQuality   float64 // 0-1 quality metric

	// Integration state
	ThermalBudgetUsed float64 // Cumulative thermal budget
	Layer             int     // BEOL metal layer position
}

// NewBEOLFeFET creates a BEOL-compatible FeFET device
func NewBEOLFeFET(config *BEOLConfig) *BEOLFeFET {
	device := &BEOLFeFET{
		Config: config,
	}
	device.initialize()
	return device
}

// initialize sets device parameters based on configuration
func (d *BEOLFeFET) initialize() {
	// Memory window depends on FE thickness and material
	baseWindow := 2.0 // V for 10nm HZO

	// Superlattice enhancement
	if d.Config.SuperlatticePeriod > 0 {
		baseWindow *= 1.0 + 0.1*float64(d.Config.SuperlatticePeriod)
	}

	// Temperature impact on crystallization quality
	if d.Config.AnnealingTempC < 300 {
		baseWindow *= 0.7 // Poor crystallization at low temp
	} else if d.Config.AnnealingTempC > 400 {
		baseWindow *= 1.1 // Better crystallization but BEOL risk
	}

	d.MemoryWindowV = baseWindow

	// On/Off ratio depends on channel material
	switch d.Config.ChannelMaterial {
	case "IGZO":
		d.OnOffRatio = 1e8
		d.SubthresholdSwingmV = 100
	case "ITO":
		d.OnOffRatio = 1e8
		d.SubthresholdSwingmV = 90
	case "PolySi":
		d.OnOffRatio = 1e6
		d.SubthresholdSwingmV = 150
	case "a-Si":
		d.OnOffRatio = 1e5
		d.SubthresholdSwingmV = 200
	default:
		d.OnOffRatio = 1e7
		d.SubthresholdSwingmV = 120
	}

	// Interface quality (oxide semiconductors avoid IL issues)
	if d.Config.ChannelMaterial == "IGZO" || d.Config.ChannelMaterial == "ITO" {
		d.InterfaceQuality = 0.95 // No defective IL
	} else {
		d.InterfaceQuality = 0.7 // Si-based has IL issues
	}

	// Threshold voltage
	d.ThresholdVoltageV = 0.5 + 0.5*rand.Float64()

	// Channel conductance (mobility dependent)
	d.ChannelConductance = d.Config.ChannelMobilityCm2Vs * 1e-6 // Normalized
}

// ProgramState programs the FeFET to a target conductance level
func (d *BEOLFeFET) ProgramState(targetConductance float64, pulseVoltage float64) float64 {
	// Actual conductance with programming noise
	noise := 0.02 * (1 - d.InterfaceQuality) * rand.NormFloat64()
	actualConductance := targetConductance * (1 + noise)

	// Clamp to valid range
	if actualConductance < 0 {
		actualConductance = 0
	}
	if actualConductance > 1 {
		actualConductance = 1
	}

	d.ChannelConductance = actualConductance
	return actualConductance
}

// ReadState reads the current conductance state
func (d *BEOLFeFET) ReadState(readVoltage float64) float64 {
	// Add read noise
	readNoise := 0.01 * rand.NormFloat64()
	return d.ChannelConductance * (1 + readNoise)
}

// ============================================================================
// MONOLITHIC 3D INTEGRATION
// ============================================================================

// M3DLayer represents a single layer in monolithic 3D integration
type M3DLayer struct {
	LayerID        int
	LayerType      string // "Logic", "Memory", "CIM"
	FeFETArray     [][]*BEOLFeFET
	InterconnectRC float64 // RC delay of TSV/via
	ThermalResistance float64 // K/W thermal resistance
}

// M3DStack represents a complete monolithic 3D integrated stack
type M3DStack struct {
	Config     *BEOLConfig
	Layers     []*M3DLayer
	TotalPower float64 // W
	MaxTemp    float64 // Peak temperature

	// Performance metrics
	ComputeDensityTOPSmm2 float64
	EnergyEfficiencyTOPSW float64
	LatencyNs             float64
}

// NewM3DStack creates a monolithic 3D integrated stack
func NewM3DStack(config *BEOLConfig, numLayers int, arraySize int) *M3DStack {
	stack := &M3DStack{
		Config: config,
		Layers: make([]*M3DLayer, numLayers),
	}

	for i := 0; i < numLayers; i++ {
		layer := &M3DLayer{
			LayerID:           i,
			FeFETArray:        make([][]*BEOLFeFET, arraySize),
			InterconnectRC:    0.1 * float64(i+1), // ns, increases with layer
			ThermalResistance: 100 + 20*float64(i), // K/W
		}

		// Alternate layer types
		if i == 0 {
			layer.LayerType = "Logic"
		} else {
			layer.LayerType = "CIM"
		}

		// Create FeFET arrays for CIM layers
		if layer.LayerType == "CIM" {
			for j := 0; j < arraySize; j++ {
				layer.FeFETArray[j] = make([]*BEOLFeFET, arraySize)
				for k := 0; k < arraySize; k++ {
					layer.FeFETArray[j][k] = NewBEOLFeFET(config)
					layer.FeFETArray[j][k].Layer = i
				}
			}
		}

		stack.Layers[i] = layer
	}

	stack.calculateMetrics()
	return stack
}

// calculateMetrics computes stack performance metrics
func (s *M3DStack) calculateMetrics() {
	totalMACs := 0.0
	totalEnergy := 0.0
	maxLatency := 0.0

	for _, layer := range s.Layers {
		if layer.LayerType == "CIM" {
			arraySize := len(layer.FeFETArray)
			macs := float64(arraySize * arraySize)
			totalMACs += macs

			// Energy per MAC (fJ)
			energyPerMAC := 50.0 // fJ for FeFET CIM
			if s.Config.ChannelMaterial == "IGZO" {
				energyPerMAC *= 0.8 // Lower leakage
			}
			totalEnergy += macs * energyPerMAC

			// Latency including interconnect
			latency := 10.0 + layer.InterconnectRC // ns
			if latency > maxLatency {
				maxLatency = latency
			}
		}
	}

	// TOPS = 10^12 OPS/sec
	// If latency is in ns, frequency = 1e9/latency
	freq := 1e9 / maxLatency
	s.ComputeDensityTOPSmm2 = totalMACs * freq / 1e12

	// Energy efficiency
	totalEnergyJ := totalEnergy * 1e-15 // fJ to J
	powerW := totalEnergyJ * freq
	s.EnergyEfficiencyTOPSW = (totalMACs * freq / 1e12) / powerW
	s.TotalPower = powerW
	s.LatencyNs = maxLatency
}

// ComputeMVM performs matrix-vector multiplication across 3D stack
func (s *M3DStack) ComputeMVM(input []float64) []float64 {
	var output []float64

	for _, layer := range s.Layers {
		if layer.LayerType != "CIM" {
			continue
		}

		arraySize := len(layer.FeFETArray)
		if len(input) > arraySize {
			input = input[:arraySize]
		}

		layerOutput := make([]float64, arraySize)

		// MVM computation
		for i := 0; i < arraySize; i++ {
			sum := 0.0
			for j := 0; j < len(input); j++ {
				conductance := layer.FeFETArray[i][j].ReadState(0.1)
				sum += input[j] * conductance
			}
			layerOutput[i] = sum
		}

		output = layerOutput
		input = output // Feed to next layer
	}

	return output
}

// ============================================================================
// WAKE-UP AND ENDURANCE SIMULATION
// ============================================================================

// WakeUpState tracks the wake-up behavior of a ferroelectric capacitor
type WakeUpState struct {
	Config *EnduranceConfig

	// Polarization state
	CurrentPruCcm2   float64 // Current remnant polarization
	InitialPruCcm2   float64 // Initial Pr
	SaturatedPruCcm2 float64 // Pr after wake-up complete

	// Cycling state
	TotalCycles      float64 // Accumulated switching cycles
	CyclesSinceReset float64 // Cycles since last recovery

	// Defect state
	OxygenVacancyDensity float64 // Current [Vo]
	TrappedCharge        float64 // Trapped charge density
	FilamentProgress     float64 // 0-1 filament formation progress

	// Wake-up metrics
	WakeUpComplete   bool
	WakeUpCycles     float64 // Cycles needed for wake-up
	IsWakeUpFree     bool    // True for optimized materials
}

// NewWakeUpState creates a new wake-up/fatigue tracker
func NewWakeUpState(config *EnduranceConfig, isSuperlattice bool) *WakeUpState {
	state := &WakeUpState{
		Config:               config,
		InitialPruCcm2:       config.PrInitialuCcm2,
		CurrentPruCcm2:       config.PrInitialuCcm2 * 0.7, // Initial depolarized
		SaturatedPruCcm2:     config.PrInitialuCcm2,
		OxygenVacancyDensity: config.OxygenVacancyDensity,
	}

	// Superlattice materials are wake-up free
	if isSuperlattice {
		state.IsWakeUpFree = true
		state.CurrentPruCcm2 = config.PrInitialuCcm2 // No initial depolarization
		state.OxygenVacancyDensity *= config.DefectReduction
		state.WakeUpComplete = true
	} else {
		state.IsWakeUpFree = false
		state.WakeUpCycles = 1e4 // Typical wake-up cycles for HZO
	}

	return state
}

// ApplyCycles simulates the effect of switching cycles
func (w *WakeUpState) ApplyCycles(numCycles float64) {
	w.TotalCycles += numCycles
	w.CyclesSinceReset += numCycles

	if !w.IsWakeUpFree && !w.WakeUpComplete {
		// Wake-up phase: Pr increases
		wakeUpProgress := w.TotalCycles / w.WakeUpCycles
		if wakeUpProgress >= 1.0 {
			w.WakeUpComplete = true
			w.CurrentPruCcm2 = w.SaturatedPruCcm2
		} else {
			// Gradual wake-up
			w.CurrentPruCcm2 = w.InitialPruCcm2*0.7 + 0.3*w.SaturatedPruCcm2*wakeUpProgress
		}
	} else {
		// Fatigue phase: Pr decreases
		w.applyFatigue(numCycles)
	}
}

// applyFatigue models polarization fatigue
func (w *WakeUpState) applyFatigue(numCycles float64) {
	// Oxygen vacancy accumulation
	vacancyGenRate := 1e12 // [Vo]/cycle
	if w.IsWakeUpFree {
		vacancyGenRate *= w.Config.DefectReduction
	}
	w.OxygenVacancyDensity += vacancyGenRate * numCycles

	// Charge trapping
	trapRate := 1e5 // charges/cycle
	w.TrappedCharge += trapRate * numCycles

	// Filament formation progress
	if w.OxygenVacancyDensity > w.Config.FilamentThreshold*0.5 {
		progressRate := (w.OxygenVacancyDensity - w.Config.FilamentThreshold*0.5) /
			(w.Config.FilamentThreshold * 0.5)
		w.FilamentProgress += progressRate * numCycles / w.Config.TargetCycles

		// Superlattice blocks filaments
		if w.IsWakeUpFree {
			w.FilamentProgress /= w.Config.InterfaceBlocking
		}
	}

	// Polarization degradation
	// Log-based fatigue model: Pr = Pr0 * (1 - A*log10(N))
	if w.TotalCycles > 1 {
		fatigueCoeff := 0.01 // A coefficient
		if w.IsWakeUpFree {
			fatigueCoeff *= 0.1 // 10x better for superlattice
		}
		degradation := fatigueCoeff * math.Log10(w.TotalCycles)
		w.CurrentPruCcm2 = w.SaturatedPruCcm2 * (1 - degradation)
	}

	// Clamp to minimum
	minPr := w.SaturatedPruCcm2 * w.Config.PrMinimumRatio
	if w.CurrentPruCcm2 < minPr {
		w.CurrentPruCcm2 = minPr
	}
}

// RecoverFatigue simulates fatigue recovery (30s rest)
func (w *WakeUpState) RecoverFatigue() {
	if !w.IsWakeUpFree {
		// Partial recovery for HZO
		recovery := (w.SaturatedPruCcm2 - w.CurrentPruCcm2) * 0.5
		w.CurrentPruCcm2 += recovery
	} else {
		// Full recovery for superlattice
		recovery := (w.SaturatedPruCcm2 - w.CurrentPruCcm2) * w.Config.RecoveryEfficiency
		w.CurrentPruCcm2 += recovery
	}

	// Reset trapped charge (detrapping)
	w.TrappedCharge *= 0.1 // 90% detrapping
	w.CyclesSinceReset = 0
}

// GetEnduranceMargin returns remaining endurance as fraction
func (w *WakeUpState) GetEnduranceMargin() float64 {
	return 1.0 - w.TotalCycles/w.Config.TargetCycles
}

// IsFailed checks if device has failed (Pr below threshold or breakdown)
func (w *WakeUpState) IsFailed() bool {
	if w.CurrentPruCcm2 < w.SaturatedPruCcm2*w.Config.PrMinimumRatio {
		return true
	}
	if w.FilamentProgress >= 1.0 {
		return true
	}
	return false
}

// ============================================================================
// SUPERLATTICE OPTIMIZATION
// ============================================================================

// SuperlatticeConfig defines HfO2/ZrO2 superlattice parameters
type SuperlatticeConfig struct {
	NumBilayers      int     // Number of HfO2/ZrO2 bilayers
	HfO2Thicknessnm  float64 // HfO2 layer thickness
	ZrO2Thicknessnm  float64 // ZrO2 layer thickness
	TotalThicknessnm float64 // Total film thickness

	// Crystallographic
	PreferredPhase  string  // "Orthorhombic", "Monoclinic", "Rhombohedral"
	PhaseUniformity float64 // 0-1 phase purity

	// Interface properties
	InterfaceDensity   float64 // Number of interfaces per nm
	OxygenVacancyProfile string // "Uniform", "Graded", "Blocked"
}

// DefaultSuperlatticeConfig returns optimized SL configuration
func DefaultSuperlatticeConfig() *SuperlatticeConfig {
	return &SuperlatticeConfig{
		NumBilayers:        10,
		HfO2Thicknessnm:    0.5,
		ZrO2Thicknessnm:    0.5,
		TotalThicknessnm:   10,
		PreferredPhase:     "Orthorhombic",
		PhaseUniformity:    0.95,
		InterfaceDensity:   1.0,
		OxygenVacancyProfile: "Blocked",
	}
}

// SuperlatticeOptimizer optimizes SL structure for endurance
type SuperlatticeOptimizer struct {
	Config *SuperlatticeConfig

	// Optimization state
	BestEndurance   float64
	BestNumBilayers int
	BestThickness   float64

	// Search results
	EnduranceMap map[int]float64 // bilayers -> endurance
}

// NewSuperlatticeOptimizer creates a new optimizer
func NewSuperlatticeOptimizer(config *SuperlatticeConfig) *SuperlatticeOptimizer {
	return &SuperlatticeOptimizer{
		Config:       config,
		EnduranceMap: make(map[int]float64),
	}
}

// OptimizeBilayers finds optimal number of bilayers
func (o *SuperlatticeOptimizer) OptimizeBilayers(minBilayers, maxBilayers int) int {
	o.BestEndurance = 0

	for n := minBilayers; n <= maxBilayers; n++ {
		endurance := o.predictEndurance(n)
		o.EnduranceMap[n] = endurance

		if endurance > o.BestEndurance {
			o.BestEndurance = endurance
			o.BestNumBilayers = n
		}
	}

	return o.BestNumBilayers
}

// predictEndurance estimates endurance for given bilayer count
func (o *SuperlatticeOptimizer) predictEndurance(numBilayers int) float64 {
	// Base endurance for solid solution HZO: ~10^9 cycles
	baseEndurance := 1e9

	// Superlattice enhancement factors:
	// 1. Interface blocking of filaments
	interfaceEnhancement := math.Pow(2, float64(numBilayers)/5)

	// 2. Reduced oxygen vacancy mobility
	vacancyReduction := 1.0 + 0.5*float64(numBilayers)

	// 3. Strain engineering benefit (peaks around 10 bilayers)
	strainOptimum := 10.0
	strainFactor := 1.5 * math.Exp(-math.Pow(float64(numBilayers)-strainOptimum, 2)/50)

	// 4. Too many bilayers increases interface defects
	if numBilayers > 15 {
		interfaceEnhancement *= math.Exp(-float64(numBilayers-15) / 10)
	}

	totalEnhancement := interfaceEnhancement * vacancyReduction * (1 + strainFactor)
	return baseEndurance * totalEnhancement
}

// GenerateVacancyProfile generates oxygen vacancy distribution
func (o *SuperlatticeOptimizer) GenerateVacancyProfile(numPoints int) []float64 {
	profile := make([]float64, numPoints)
	bilayerThickness := o.Config.TotalThicknessnm / float64(o.Config.NumBilayers)

	for i := 0; i < numPoints; i++ {
		position := float64(i) / float64(numPoints-1) * o.Config.TotalThicknessnm
		bilayerIndex := int(position / bilayerThickness)
		posInBilayer := math.Mod(position, bilayerThickness)

		// Vacancy concentration varies: lower in HfO2, higher in ZrO2 interfaces
		halfThickness := bilayerThickness / 2
		if posInBilayer < halfThickness {
			// HfO2 layer - lower vacancy
			profile[i] = 0.3 + 0.2*rand.Float64()
		} else {
			// ZrO2 layer - higher vacancy but blocked at interface
			distToInterface := math.Min(posInBilayer-halfThickness,
				bilayerThickness-posInBilayer)
			blocking := 1.0 - math.Exp(-distToInterface/0.2)
			profile[i] = (0.5 + 0.3*rand.Float64()) * blocking
		}

		// Add bilayer index variation
		profile[i] *= (1.0 + 0.1*float64(bilayerIndex%2))
	}

	return profile
}

// ============================================================================
// INTEGRATED BEOL + ENDURANCE SYSTEM
// ============================================================================

// IronLatticeBEOLSystem combines BEOL integration with endurance optimization
type IronLatticeBEOLSystem struct {
	BEOLConfig      *BEOLConfig
	EnduranceConfig *EnduranceConfig
	SLConfig        *SuperlatticeConfig

	// Components
	M3DStack    *M3DStack
	Optimizer   *SuperlatticeOptimizer
	WakeUpStates [][]*WakeUpState

	// System metrics
	TotalDevices      int
	HealthyDevices    int
	AveragePruCcm2    float64
	SystemEndurance   float64
	ThermalHeadroom   float64
}

// NewIronLatticeBEOLSystem creates an integrated BEOL + endurance system
func NewIronLatticeBEOLSystem(numLayers, arraySize int) *IronLatticeBEOLSystem {
	beolConfig := DefaultBEOLConfig()
	endConfig := DefaultEnduranceConfig()
	slConfig := DefaultSuperlatticeConfig()

	system := &IronLatticeBEOLSystem{
		BEOLConfig:      beolConfig,
		EnduranceConfig: endConfig,
		SLConfig:        slConfig,
	}

	// Create M3D stack
	system.M3DStack = NewM3DStack(beolConfig, numLayers, arraySize)

	// Create optimizer
	system.Optimizer = NewSuperlatticeOptimizer(slConfig)
	system.Optimizer.OptimizeBilayers(5, 20)

	// Create wake-up states for each device
	isSuperlattice := beolConfig.SuperlatticePeriod > 0
	system.WakeUpStates = make([][]*WakeUpState, arraySize)
	for i := 0; i < arraySize; i++ {
		system.WakeUpStates[i] = make([]*WakeUpState, arraySize)
		for j := 0; j < arraySize; j++ {
			system.WakeUpStates[i][j] = NewWakeUpState(endConfig, isSuperlattice)
		}
	}

	system.TotalDevices = arraySize * arraySize
	system.HealthyDevices = system.TotalDevices
	system.calculateSystemMetrics()

	return system
}

// calculateSystemMetrics updates system-level metrics
func (s *IronLatticeBEOLSystem) calculateSystemMetrics() {
	totalPr := 0.0
	healthy := 0
	minEndurance := 1.0

	for i := range s.WakeUpStates {
		for j := range s.WakeUpStates[i] {
			state := s.WakeUpStates[i][j]
			totalPr += state.CurrentPruCcm2
			if !state.IsFailed() {
				healthy++
			}
			enduranceMargin := state.GetEnduranceMargin()
			if enduranceMargin < minEndurance {
				minEndurance = enduranceMargin
			}
		}
	}

	s.HealthyDevices = healthy
	s.AveragePruCcm2 = totalPr / float64(s.TotalDevices)
	s.SystemEndurance = minEndurance

	// Thermal headroom based on BEOL constraints
	usedTemp := s.BEOLConfig.AnnealingTempC
	s.ThermalHeadroom = (s.BEOLConfig.MaxTemperatureC - usedTemp) /
		s.BEOLConfig.MaxTemperatureC
}

// RunInference runs inference with cycling/fatigue simulation
func (s *IronLatticeBEOLSystem) RunInference(input []float64, numIterations int) [][]float64 {
	results := make([][]float64, numIterations)

	cyclesPerInference := 100.0 // Cycles per inference operation

	for iter := 0; iter < numIterations; iter++ {
		// Apply cycles to all devices
		for i := range s.WakeUpStates {
			for j := range s.WakeUpStates[i] {
				s.WakeUpStates[i][j].ApplyCycles(cyclesPerInference)
			}
		}

		// Run MVM computation
		results[iter] = s.M3DStack.ComputeMVM(input)

		// Periodic recovery (every 1000 iterations)
		if iter > 0 && iter%1000 == 0 {
			s.applyRecovery()
		}
	}

	s.calculateSystemMetrics()
	return results
}

// applyRecovery applies fatigue recovery to all devices
func (s *IronLatticeBEOLSystem) applyRecovery() {
	for i := range s.WakeUpStates {
		for j := range s.WakeUpStates[i] {
			s.WakeUpStates[i][j].RecoverFatigue()
		}
	}
}

// GetHealthReport returns detailed system health information
func (s *IronLatticeBEOLSystem) GetHealthReport() map[string]interface{} {
	// Calculate Pr distribution
	prValues := make([]float64, 0, s.TotalDevices)
	for i := range s.WakeUpStates {
		for j := range s.WakeUpStates[i] {
			prValues = append(prValues, s.WakeUpStates[i][j].CurrentPruCcm2)
		}
	}
	sort.Float64s(prValues)

	// Percentiles
	p5 := prValues[int(0.05*float64(len(prValues)))]
	p50 := prValues[int(0.50*float64(len(prValues)))]
	p95 := prValues[int(0.95*float64(len(prValues)))]

	return map[string]interface{}{
		"total_devices":      s.TotalDevices,
		"healthy_devices":    s.HealthyDevices,
		"health_percentage":  100.0 * float64(s.HealthyDevices) / float64(s.TotalDevices),
		"average_pr_uC_cm2":  s.AveragePruCcm2,
		"pr_p5_uC_cm2":       p5,
		"pr_p50_uC_cm2":      p50,
		"pr_p95_uC_cm2":      p95,
		"system_endurance":   s.SystemEndurance,
		"thermal_headroom":   s.ThermalHeadroom,
		"beol_temp_limit_C":  s.BEOLConfig.MaxTemperatureC,
		"optimal_bilayers":   s.Optimizer.BestNumBilayers,
		"predicted_endurance": s.Optimizer.BestEndurance,
	}
}

// ============================================================================
// FATIGUE PREDICTION MODEL
// ============================================================================

// FatiguePredictionModel predicts device lifetime and degradation
type FatiguePredictionModel struct {
	// Model parameters (fitted to experimental data)
	AlphaFatigue    float64 // Fatigue rate coefficient
	BetaRecovery    float64 // Recovery efficiency
	GammaFilament   float64 // Filament formation rate
	DeltaVacancy    float64 // Vacancy generation rate

	// Operating conditions
	OperatingTempC   float64
	OperatingVoltageV float64
	DutyCycle        float64 // Fraction of time under bias
}

// NewFatiguePredictionModel creates a calibrated fatigue model
func NewFatiguePredictionModel() *FatiguePredictionModel {
	return &FatiguePredictionModel{
		AlphaFatigue:     0.01,
		BetaRecovery:     0.8,
		GammaFilament:    1e-15,
		DeltaVacancy:     1e-10,
		OperatingTempC:   85,
		OperatingVoltageV: 3.0,
		DutyCycle:        0.5,
	}
}

// PredictLifetime predicts cycles to failure for given conditions
func (m *FatiguePredictionModel) PredictLifetime(isSuperlattice bool, numBilayers int) float64 {
	// Arrhenius temperature acceleration
	kB := 8.617e-5 // eV/K
	Ea := 0.7      // Activation energy (eV)
	tempK := m.OperatingTempC + 273.15
	tempFactor := math.Exp(-Ea / (kB * tempK))

	// Voltage acceleration (power law)
	Vref := 2.5
	voltFactor := math.Pow(m.OperatingVoltageV/Vref, 3)

	// Base lifetime for HZO: 10^9 cycles
	baseLifetime := 1e9

	// Superlattice enhancement
	slEnhancement := 1.0
	if isSuperlattice {
		// Each bilayer adds ~2x enhancement up to optimum
		slEnhancement = math.Pow(2, math.Min(float64(numBilayers), 10)/5)
		// Additional oxygen vacancy blocking
		slEnhancement *= (1 + float64(numBilayers)*0.1)
	}

	// Combined prediction
	lifetime := baseLifetime / (tempFactor * voltFactor * m.DutyCycle) * slEnhancement

	return lifetime
}

// PredictDegradationCurve generates Pr vs cycles curve
func (m *FatiguePredictionModel) PredictDegradationCurve(Pr0 float64, cycles []float64,
	isSuperlattice bool) []float64 {
	prValues := make([]float64, len(cycles))

	for i, n := range cycles {
		if n <= 0 {
			prValues[i] = Pr0
			continue
		}

		// Log-based degradation: Pr = Pr0 * (1 - alpha * log10(N))
		alpha := m.AlphaFatigue
		if isSuperlattice {
			alpha *= 0.1 // 10x better for SL
		}

		degradation := alpha * math.Log10(n)
		pr := Pr0 * (1 - degradation)

		// Clamp to minimum (90% of Pr0)
		if pr < Pr0*0.9 {
			pr = Pr0 * 0.9
		}

		prValues[i] = pr
	}

	return prValues
}

// ============================================================================
// BEOL PROCESS WINDOW ANALYSIS
// ============================================================================

// ProcessWindow defines acceptable process parameter ranges
type ProcessWindow struct {
	TempMinC    float64
	TempMaxC    float64
	TimeMinSec  float64
	TimeMaxSec  float64
	VoltageMinV float64
	VoltageMaxV float64
}

// ProcessWindowAnalyzer finds optimal BEOL process conditions
type ProcessWindowAnalyzer struct {
	// Target device characteristics
	TargetMemoryWindowV float64
	TargetEndurance     float64
	TargetOnOffRatio    float64

	// Acceptable windows
	CrystallizationWindow *ProcessWindow
	ProgrammingWindow     *ProcessWindow

	// Analysis results
	OptimalTempC     float64
	OptimalTimeSec   float64
	ProcessYield     float64
}

// NewProcessWindowAnalyzer creates a new analyzer
func NewProcessWindowAnalyzer() *ProcessWindowAnalyzer {
	return &ProcessWindowAnalyzer{
		TargetMemoryWindowV: 2.5,
		TargetEndurance:     1e10,
		TargetOnOffRatio:    1e7,
	}
}

// AnalyzeCrystallization finds optimal crystallization conditions
func (a *ProcessWindowAnalyzer) AnalyzeCrystallization(channelMaterial string) *ProcessWindow {
	window := &ProcessWindow{}

	switch channelMaterial {
	case "IGZO", "ITO":
		// Oxide semiconductors allow lower temp
		window.TempMinC = 300
		window.TempMaxC = 400
		window.TimeMinSec = 30
		window.TimeMaxSec = 120
	case "PolySi":
		// Poly-Si needs higher temp for recrystallization
		window.TempMinC = 400
		window.TempMaxC = 500
		window.TimeMinSec = 60
		window.TimeMaxSec = 300
	default:
		window.TempMinC = 350
		window.TempMaxC = 450
		window.TimeMinSec = 60
		window.TimeMaxSec = 180
	}

	a.CrystallizationWindow = window

	// Find optimal point (center of window typically best)
	a.OptimalTempC = (window.TempMinC + window.TempMaxC) / 2
	a.OptimalTimeSec = (window.TimeMinSec + window.TimeMaxSec) / 2

	return window
}

// EstimateYield estimates process yield within window
func (a *ProcessWindowAnalyzer) EstimateYield(tempVariationC, timeVariationSec float64) float64 {
	if a.CrystallizationWindow == nil {
		return 0
	}

	// Gaussian yield model
	tempWindow := a.CrystallizationWindow.TempMaxC - a.CrystallizationWindow.TempMinC
	timeWindow := a.CrystallizationWindow.TimeMaxSec - a.CrystallizationWindow.TimeMinSec

	tempYield := math.Erf(tempWindow / (2 * math.Sqrt(2) * tempVariationC))
	timeYield := math.Erf(timeWindow / (2 * math.Sqrt(2) * timeVariationSec))

	a.ProcessYield = tempYield * timeYield
	return a.ProcessYield
}

// ============================================================================
// VISUALIZATION HELPERS
// ============================================================================

// GenerateEndurancePlot creates ASCII visualization of endurance vs bilayers
func GenerateEndurancePlot(optimizer *SuperlatticeOptimizer) string {
	if len(optimizer.EnduranceMap) == 0 {
		return "No optimization data available"
	}

	// Find range
	minBilayers, maxBilayers := 100, 0
	maxEndurance := 0.0
	for n, e := range optimizer.EnduranceMap {
		if n < minBilayers {
			minBilayers = n
		}
		if n > maxBilayers {
			maxBilayers = n
		}
		if e > maxEndurance {
			maxEndurance = e
		}
	}

	// Build plot
	plot := "Endurance vs Superlattice Bilayers\n"
	plot += "═══════════════════════════════════════════\n"
	plot += fmt.Sprintf("Target: %.0e cycles\n\n", 1e12)

	height := 10
	width := maxBilayers - minBilayers + 1

	for row := height; row >= 0; row-- {
		threshold := maxEndurance * float64(row) / float64(height)
		line := fmt.Sprintf("%8.0e │", threshold)

		for n := minBilayers; n <= maxBilayers; n++ {
			e := optimizer.EnduranceMap[n]
			if e >= threshold {
				if n == optimizer.BestNumBilayers {
					line += "█"
				} else {
					line += "▓"
				}
			} else {
				line += " "
			}
		}
		plot += line + "\n"
	}

	// X-axis
	plot += "         └" + fmt.Sprintf("%s", repeatChar("─", width)) + "\n"
	plot += "          "
	for n := minBilayers; n <= maxBilayers; n++ {
		if n%5 == 0 {
			plot += fmt.Sprintf("%d", n%10)
		} else {
			plot += " "
		}
	}
	plot += "\n"
	plot += fmt.Sprintf("          Number of Bilayers (%d-%d)\n", minBilayers, maxBilayers)
	plot += fmt.Sprintf("\n★ Optimal: %d bilayers → %.2e cycles\n",
		optimizer.BestNumBilayers, optimizer.BestEndurance)

	return plot
}

// repeatChar repeats a character n times
func repeatChar(char string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += char
	}
	return result
}

// GenerateFatigueCurve creates ASCII visualization of fatigue behavior
func GenerateFatigueCurve(isSuperlattice bool, initialPr float64) string {
	model := NewFatiguePredictionModel()

	// Generate cycle points (log scale)
	cycles := []float64{1, 10, 100, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12}
	prValuesHZO := model.PredictDegradationCurve(initialPr, cycles, false)
	prValuesSL := model.PredictDegradationCurve(initialPr, cycles, true)

	plot := "Polarization Fatigue: HZO vs Superlattice\n"
	plot += "═══════════════════════════════════════════\n\n"

	height := 10
	width := len(cycles)

	for row := height; row >= 0; row-- {
		threshold := initialPr * float64(row) / float64(height)
		line := fmt.Sprintf("%5.1f │", threshold)

		for i := 0; i < width; i++ {
			hzoAbove := prValuesHZO[i] >= threshold
			slAbove := prValuesSL[i] >= threshold

			if hzoAbove && slAbove {
				line += "█" // Both above
			} else if slAbove {
				line += "▓" // Only SL above
			} else if hzoAbove {
				line += "░" // Only HZO above (shouldn't happen often)
			} else {
				line += " " // Neither
			}
		}
		plot += line + "\n"
	}

	// X-axis
	plot += "      └" + repeatChar("─", width) + "\n"
	plot += "       "
	for i := range cycles {
		if i%3 == 0 {
			plot += fmt.Sprintf("%.0e", cycles[i])[:1]
		} else {
			plot += " "
		}
	}
	plot += "\n"
	plot += "       Cycles (log scale)\n\n"
	plot += "Legend: █ Both OK  ▓ SL only  ░ HZO only\n"
	plot += fmt.Sprintf("Initial Pr: %.1f µC/cm²\n", initialPr)

	return plot
}

// GenerateM3DStackDiagram creates ASCII visualization of 3D stack
func GenerateM3DStackDiagram(stack *M3DStack) string {
	diagram := "Monolithic 3D FeFET CIM Stack\n"
	diagram += "═══════════════════════════════════════════\n\n"

	numLayers := len(stack.Layers)

	for i := numLayers - 1; i >= 0; i-- {
		layer := stack.Layers[i]

		// Layer separator
		diagram += "  ┌──────────────────────────────┐\n"

		// Layer content
		switch layer.LayerType {
		case "Logic":
			diagram += fmt.Sprintf("  │  Layer %d: CMOS Logic (FEOL)  │\n", i)
			diagram += "  │  ┌───┐ ┌───┐ ┌───┐ ┌───┐   │\n"
			diagram += "  │  │AND│ │ OR│ │INV│ │MUX│   │\n"
			diagram += "  │  └───┘ └───┘ └───┘ └───┘   │\n"
		case "CIM":
			arraySize := len(layer.FeFETArray)
			diagram += fmt.Sprintf("  │  Layer %d: FeFET CIM (%dx%d)  │\n", i, arraySize, arraySize)
			diagram += "  │  ╔═══╤═══╤═══╤═══╗        │\n"
			diagram += "  │  ║ G │ G │ G │ G ║ ← WL   │\n"
			diagram += "  │  ╟───┼───┼───┼───╢        │\n"
			diagram += "  │  ║ G │ G │ G │ G ║        │\n"
			diagram += "  │  ╚═══╧═══╧═══╧═══╝        │\n"
			diagram += "  │      ↓   ↓   ↓   ↓        │\n"
			diagram += "  │     BL  BL  BL  BL        │\n"
		default:
			diagram += fmt.Sprintf("  │  Layer %d: %s             │\n", i, layer.LayerType)
		}

		diagram += "  └──────────────────────────────┘\n"

		// Inter-layer via
		if i > 0 {
			diagram += "         │ │ │ │ │ │ ← TSV/Via\n"
			diagram += fmt.Sprintf("         (RC: %.2f ns)\n", layer.InterconnectRC)
		}
	}

	diagram += "\n"
	diagram += "  ════════════════════════════════\n"
	diagram += "         Silicon Substrate\n"
	diagram += "  ════════════════════════════════\n\n"

	// Performance summary
	diagram += fmt.Sprintf("Performance Metrics:\n")
	diagram += fmt.Sprintf("  • Compute Density: %.2f TOPS/mm²\n", stack.ComputeDensityTOPSmm2)
	diagram += fmt.Sprintf("  • Energy Efficiency: %.0f TOPS/W\n", stack.EnergyEfficiencyTOPSW)
	diagram += fmt.Sprintf("  • Latency: %.1f ns\n", stack.LatencyNs)
	diagram += fmt.Sprintf("  • BEOL Layers: %d\n", stack.Config.MetalLayers)

	return diagram
}

// ============================================================================
// EXAMPLE USAGE AND DEMO
// ============================================================================

// RunBEOLEnduranceDemo demonstrates BEOL + endurance simulation
func RunBEOLEnduranceDemo() {
	fmt.Println("╔═══════════════════════════════════════════════════════╗")
	fmt.Println("║  IronLattice BEOL Integration & Endurance Simulation  ║")
	fmt.Println("╚═══════════════════════════════════════════════════════╝")
	fmt.Println()

	// 1. Create integrated system
	fmt.Println("1. Creating Monolithic 3D FeFET CIM System...")
	system := NewIronLatticeBEOLSystem(4, 64)

	fmt.Printf("   • %d total FeFET devices\n", system.TotalDevices)
	fmt.Printf("   • BEOL thermal limit: %.0f°C\n", system.BEOLConfig.MaxTemperatureC)
	fmt.Printf("   • Channel material: %s\n", system.BEOLConfig.ChannelMaterial)
	fmt.Printf("   • Superlattice periods: %d\n", system.BEOLConfig.SuperlatticePeriod)
	fmt.Println()

	// 2. Optimize superlattice structure
	fmt.Println("2. Optimizing Superlattice Structure...")
	fmt.Println(GenerateEndurancePlot(system.Optimizer))
	fmt.Println()

	// 3. Show M3D stack
	fmt.Println("3. Monolithic 3D Stack Architecture:")
	fmt.Println(GenerateM3DStackDiagram(system.M3DStack))
	fmt.Println()

	// 4. Run inference with fatigue simulation
	fmt.Println("4. Running Inference with Fatigue Simulation...")
	input := make([]float64, 64)
	for i := range input {
		input[i] = rand.Float64()
	}

	results := system.RunInference(input, 10000)
	fmt.Printf("   • Completed 10,000 inference iterations\n")
	fmt.Printf("   • Total cycles applied: 1M\n")
	fmt.Printf("   • Final output sample: [%.3f, %.3f, %.3f, ...]\n",
		results[9999][0], results[9999][1], results[9999][2])
	fmt.Println()

	// 5. Health report
	fmt.Println("5. System Health Report:")
	report := system.GetHealthReport()
	fmt.Printf("   • Healthy devices: %d/%d (%.1f%%)\n",
		report["healthy_devices"], report["total_devices"],
		report["health_percentage"])
	fmt.Printf("   • Average Pr: %.2f µC/cm²\n", report["average_pr_uC_cm2"])
	fmt.Printf("   • Pr distribution: p5=%.2f, p50=%.2f, p95=%.2f µC/cm²\n",
		report["pr_p5_uC_cm2"], report["pr_p50_uC_cm2"], report["pr_p95_uC_cm2"])
	fmt.Printf("   • System endurance margin: %.2f%%\n", report["system_endurance"].(float64)*100)
	fmt.Printf("   • Thermal headroom: %.1f%%\n", report["thermal_headroom"].(float64)*100)
	fmt.Println()

	// 6. Fatigue comparison
	fmt.Println("6. Fatigue Comparison (HZO vs Superlattice):")
	fmt.Println(GenerateFatigueCurve(true, 27.4))
	fmt.Println()

	// 7. Process window analysis
	fmt.Println("7. BEOL Process Window Analysis:")
	analyzer := NewProcessWindowAnalyzer()
	window := analyzer.AnalyzeCrystallization("IGZO")
	yield := analyzer.EstimateYield(10, 5)
	fmt.Printf("   • Crystallization window: %.0f-%.0f°C, %.0f-%.0fs\n",
		window.TempMinC, window.TempMaxC, window.TimeMinSec, window.TimeMaxSec)
	fmt.Printf("   • Optimal conditions: %.0f°C, %.0fs\n",
		analyzer.OptimalTempC, analyzer.OptimalTimeSec)
	fmt.Printf("   • Estimated yield (±10°C, ±5s): %.1f%%\n", yield*100)
	fmt.Println()

	// 8. Lifetime prediction
	fmt.Println("8. Device Lifetime Prediction:")
	model := NewFatiguePredictionModel()
	lifetimeHZO := model.PredictLifetime(false, 0)
	lifetimeSL := model.PredictLifetime(true, 10)
	fmt.Printf("   • HZO solid solution: %.2e cycles\n", lifetimeHZO)
	fmt.Printf("   • HfO2/ZrO2 superlattice (10 bilayers): %.2e cycles\n", lifetimeSL)
	fmt.Printf("   • Improvement factor: %.0fx\n", lifetimeSL/lifetimeHZO)
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("BEOL + Endurance simulation complete!")
}
