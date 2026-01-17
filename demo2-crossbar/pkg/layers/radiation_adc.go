// Package layers provides radiation hardening and ADC optimization for CIM arrays.
// This module simulates radiation effects on ferroelectric memory and various
// ADC architectures for compute-in-memory accelerators.
package layers

import (
	"math"
	"math/rand"
	"time"
)

// =============================================================================
// RADIATION ENVIRONMENT MODELING
// =============================================================================

// RadiationEnvironment defines space radiation conditions
type RadiationEnvironment struct {
	// Total Ionizing Dose parameters
	TIDRate        float64 // krad(Si)/year
	AccumulatedTID float64 // krad(Si) total

	// Single Event Effects parameters
	HeavyIonFlux   float64 // particles/cm²/s
	ProtonFlux     float64 // particles/cm²/s
	LETSpectrum    []float64 // LET values in MeV·cm²/mg

	// Environment type
	OrbitType      string  // "LEO", "MEO", "GEO", "Deep"
	MissionDuration float64 // years
	ShieldingMass  float64 // g/cm² aluminum equivalent
}

// NewLEOEnvironment creates a Low Earth Orbit radiation environment
func NewLEOEnvironment(altitude float64, inclination float64, years float64) *RadiationEnvironment {
	// LEO trapped proton environment (South Atlantic Anomaly dominant)
	tidRate := 5.0 // krad/year typical for LEO
	if inclination > 50 {
		tidRate *= 1.5 // Higher inclination = more SAA exposure
	}

	return &RadiationEnvironment{
		TIDRate:         tidRate,
		AccumulatedTID:  tidRate * years,
		HeavyIonFlux:    1e-8, // particles/cm²/s
		ProtonFlux:      1e4,  // particles/cm²/s (SAA)
		LETSpectrum:     []float64{1.0, 5.0, 10.0, 20.0, 40.0, 80.0},
		OrbitType:       "LEO",
		MissionDuration: years,
		ShieldingMass:   2.0, // g/cm² typical
	}
}

// NewGEOEnvironment creates a Geostationary Orbit radiation environment
func NewGEOEnvironment(years float64) *RadiationEnvironment {
	return &RadiationEnvironment{
		TIDRate:         20.0, // krad/year (higher than LEO)
		AccumulatedTID:  20.0 * years,
		HeavyIonFlux:    1e-6, // Higher galactic cosmic ray flux
		ProtonFlux:      1e2,  // Lower trapped protons
		LETSpectrum:     []float64{1.0, 10.0, 30.0, 60.0, 100.0, 114.0},
		OrbitType:       "GEO",
		MissionDuration: years,
		ShieldingMass:   5.0,
	}
}

// NewDeepSpaceEnvironment creates a deep space (lunar/Mars) environment
func NewDeepSpaceEnvironment(years float64) *RadiationEnvironment {
	return &RadiationEnvironment{
		TIDRate:         50.0, // krad/year (solar particle events)
		AccumulatedTID:  50.0 * years,
		HeavyIonFlux:    1e-5, // Galactic cosmic rays dominant
		ProtonFlux:      1e6,  // During solar particle events
		LETSpectrum:     []float64{1.0, 20.0, 50.0, 100.0, 150.0, 200.0},
		OrbitType:       "Deep",
		MissionDuration: years,
		ShieldingMass:   10.0,
	}
}

// =============================================================================
// RADIATION EFFECTS ON FERROELECTRIC MEMORY
// =============================================================================

// FerroelectricRadHardConfig configures radiation-hardened ferroelectric memory
type FerroelectricRadHardConfig struct {
	// Material properties
	Material         string  // "HZO", "PZT", "AlScN"
	FilmThickness    float64 // nm
	OrthorhombicFrac float64 // Fraction of ferroelectric phase

	// Radiation tolerance specs
	TIDTolerance     float64 // krad(Si) before degradation
	SEUThresholdLET  float64 // MeV·cm²/mg (for SEU immunity)
	SELThresholdLET  float64 // MeV·cm²/mg (for latchup immunity)

	// Device structure
	CellType         string  // "1T1C", "2T2C", "FeFET"
	RedundancyFactor int     // TMR (Triple Modular Redundancy)
}

// DefaultHZORadHardConfig returns radiation-hardened HZO configuration
func DefaultHZORadHardConfig() *FerroelectricRadHardConfig {
	return &FerroelectricRadHardConfig{
		Material:         "HZO",
		FilmThickness:    12.0,          // nm
		OrthorhombicFrac: 0.85,          // 85% orthorhombic phase
		TIDTolerance:     150.0,         // >150 krad(Si) for QML-V
		SEUThresholdLET:  math.Inf(1),   // Immune (ferroelectric)
		SELThresholdLET:  114.0,         // MeV·cm²/mg
		CellType:         "2T2C",
		RedundancyFactor: 1,
	}
}

// RadHardFerroelectricCell models a radiation-hardened ferroelectric memory cell
type RadHardFerroelectricCell struct {
	Config           *FerroelectricRadHardConfig
	Polarization     float64 // Remanent polarization (μC/cm²)
	BasePolarization float64 // Original polarization
	AccumulatedTID   float64 // krad(Si)
	HeavyIonHits     int     // Number of heavy ion interactions
	DegradationState float64 // 0.0 = pristine, 1.0 = failed
}

// NewRadHardFerroelectricCell creates a new rad-hard ferroelectric cell
func NewRadHardFerroelectricCell(config *FerroelectricRadHardConfig, pr float64) *RadHardFerroelectricCell {
	return &RadHardFerroelectricCell{
		Config:           config,
		Polarization:     pr,
		BasePolarization: pr,
		AccumulatedTID:   0,
		HeavyIonHits:     0,
		DegradationState: 0,
	}
}

// ApplyTID simulates Total Ionizing Dose effects on the cell
func (cell *RadHardFerroelectricCell) ApplyTID(dose float64) {
	cell.AccumulatedTID += dose

	// HZO shows only ~5% degradation at 10 Mrad
	// Model: exponential degradation with material-specific rate
	var degradationRate float64
	switch cell.Config.Material {
	case "HZO":
		degradationRate = 5e-6 // 5% per 10 Mrad = 5e-6 per krad
	case "PZT":
		degradationRate = 2e-5 // PZT less rad-hard
	case "AlScN":
		degradationRate = 1e-5 // AlScN moderate
	default:
		degradationRate = 1e-5
	}

	// Polarization degradation
	degradation := 1.0 - math.Exp(-degradationRate*cell.AccumulatedTID)
	cell.Polarization = cell.BasePolarization * (1.0 - degradation)
	cell.DegradationState = degradation

	// Heavy ion damage reduces orthorhombic phase fraction
	if cell.HeavyIonHits > 0 {
		phaseLoss := float64(cell.HeavyIonHits) * 0.001 // 0.1% per hit
		cell.Config.OrthorhombicFrac = math.Max(0.5, cell.Config.OrthorhombicFrac-phaseLoss)
	}
}

// ApplyHeavyIon simulates a heavy ion strike on the cell
func (cell *RadHardFerroelectricCell) ApplyHeavyIon(let float64) bool {
	cell.HeavyIonHits++

	// Ferroelectric cells are inherently SEU immune
	// but heavy ions can cause phase transitions at high LET
	if let > 100 { // Very high LET
		// Possible orthorhombic → monoclinic/tetragonal transition
		phaseLoss := (let - 100) * 0.0001
		cell.Config.OrthorhombicFrac = math.Max(0.5, cell.Config.OrthorhombicFrac-phaseLoss)
		return true // Transient effect occurred
	}

	return false // No effect
}

// IsFunctional checks if the cell is still operational
func (cell *RadHardFerroelectricCell) IsFunctional() bool {
	// Memory window must be sufficient for reliable read
	minPolarization := cell.BasePolarization * 0.5 // 50% margin
	return cell.Polarization >= minPolarization && cell.Config.OrthorhombicFrac > 0.6
}

// =============================================================================
// RADIATION-HARDENED CIM CROSSBAR
// =============================================================================

// RadHardCrossbarConfig configures a radiation-hardened CIM crossbar
type RadHardCrossbarConfig struct {
	Rows              int
	Cols              int
	CellConfig        *FerroelectricRadHardConfig
	UseTMR            bool    // Triple Modular Redundancy
	UseECC            bool    // Error Correction Coding
	ECCBits           int     // ECC overhead bits
	ScrubInterval     float64 // Seconds between scrub cycles
	SEFIRecoveryTime  float64 // Seconds to recover from SEFI
}

// RadHardCrossbar implements a radiation-hardened CIM crossbar array
type RadHardCrossbar struct {
	Config          *RadHardCrossbarConfig
	Cells           [][]*RadHardFerroelectricCell
	TMRCells        [][][]*RadHardFerroelectricCell // [copy][row][col]
	Environment     *RadiationEnvironment
	MissionTime     float64 // Elapsed mission time (seconds)
	SEFICount       int
	SEUCount        int     // Should be 0 for ferroelectric
	TotalIonHits    int
	LastScrubTime   float64
}

// NewRadHardCrossbar creates a new radiation-hardened crossbar
func NewRadHardCrossbar(config *RadHardCrossbarConfig, env *RadiationEnvironment) *RadHardCrossbar {
	cb := &RadHardCrossbar{
		Config:      config,
		Environment: env,
		MissionTime: 0,
		Cells:       make([][]*RadHardFerroelectricCell, config.Rows),
	}

	// Initialize primary cells
	for i := 0; i < config.Rows; i++ {
		cb.Cells[i] = make([]*RadHardFerroelectricCell, config.Cols)
		for j := 0; j < config.Cols; j++ {
			cb.Cells[i][j] = NewRadHardFerroelectricCell(config.CellConfig, 20.0) // 20 μC/cm²
		}
	}

	// Initialize TMR copies if enabled
	if config.UseTMR {
		cb.TMRCells = make([][][]*RadHardFerroelectricCell, 3)
		for copy := 0; copy < 3; copy++ {
			cb.TMRCells[copy] = make([][]*RadHardFerroelectricCell, config.Rows)
			for i := 0; i < config.Rows; i++ {
				cb.TMRCells[copy][i] = make([]*RadHardFerroelectricCell, config.Cols)
				for j := 0; j < config.Cols; j++ {
					cb.TMRCells[copy][i][j] = NewRadHardFerroelectricCell(config.CellConfig, 20.0)
				}
			}
		}
	}

	return cb
}

// SimulateRadiation simulates radiation exposure over a time period
func (cb *RadHardCrossbar) SimulateRadiation(duration float64) *RadiationSimResult {
	result := &RadiationSimResult{
		Duration:    duration,
		StartTID:    cb.Environment.AccumulatedTID,
		HeavyIonHits: 0,
		CellFailures: 0,
	}

	// Calculate TID accumulation
	tidDose := cb.Environment.TIDRate * (duration / (365.25 * 24 * 3600)) // Convert seconds to years

	// Apply TID to all cells
	for i := 0; i < cb.Config.Rows; i++ {
		for j := 0; j < cb.Config.Cols; j++ {
			cb.Cells[i][j].ApplyTID(tidDose)
			if !cb.Cells[i][j].IsFunctional() {
				result.CellFailures++
			}
		}
	}

	// Simulate heavy ion strikes (Poisson process)
	cellArea := 1e-10 // cm² per cell (typical)
	totalArea := cellArea * float64(cb.Config.Rows*cb.Config.Cols)
	expectedHits := cb.Environment.HeavyIonFlux * totalArea * duration

	numHits := int(rand.NormFloat64()*math.Sqrt(expectedHits) + expectedHits)
	if numHits < 0 {
		numHits = 0
	}

	for h := 0; h < numHits; h++ {
		// Random cell selection
		row := rand.Intn(cb.Config.Rows)
		col := rand.Intn(cb.Config.Cols)

		// Random LET from spectrum
		letIdx := rand.Intn(len(cb.Environment.LETSpectrum))
		let := cb.Environment.LETSpectrum[letIdx]

		if cb.Cells[row][col].ApplyHeavyIon(let) {
			result.HeavyIonHits++
			cb.TotalIonHits++
		}
	}

	cb.MissionTime += duration
	cb.Environment.AccumulatedTID += tidDose
	result.EndTID = cb.Environment.AccumulatedTID
	result.FunctionalCells = cb.Config.Rows*cb.Config.Cols - result.CellFailures

	return result
}

// RadiationSimResult contains simulation results
type RadiationSimResult struct {
	Duration        float64
	StartTID        float64
	EndTID          float64
	HeavyIonHits    int
	CellFailures    int
	FunctionalCells int
}

// TMRVote performs Triple Modular Redundancy voting for a cell
func (cb *RadHardCrossbar) TMRVote(row, col int) float64 {
	if !cb.Config.UseTMR {
		return cb.Cells[row][col].Polarization
	}

	values := make([]float64, 3)
	for i := 0; i < 3; i++ {
		values[i] = cb.TMRCells[i][row][col].Polarization
	}

	// Majority voting (median for analog)
	if values[0] <= values[1] && values[1] <= values[2] {
		return values[1]
	} else if values[1] <= values[0] && values[0] <= values[2] {
		return values[0]
	} else {
		return values[2]
	}
}

// =============================================================================
// ADC ARCHITECTURES FOR CIM
// =============================================================================

// ADCType defines the type of ADC architecture
type ADCType int

const (
	ADCTypeSAR ADCType = iota
	ADCTypeFlash
	ADCTypePipeline
	ADCTypeInMemory
	ADCTypeAdaptive
	ADCTypeTimeDomain
)

// ADCConfig configures an ADC for CIM
type ADCConfig struct {
	Type           ADCType
	Resolution     int     // bits
	SamplingRate   float64 // Hz
	Technology     int     // nm process node
	VDD            float64 // Supply voltage (V)

	// SAR-specific
	SARCycles      int     // Number of SAR cycles
	CapacitorDAC   bool    // Use capacitor DAC

	// Flash-specific
	FlashComparators int   // Number of comparators (2^N - 1)

	// Adaptive-specific
	AdaptiveMinBits int    // Minimum bits in adaptive mode
	AdaptiveMaxBits int    // Maximum bits in adaptive mode

	// Area/Energy targets
	TargetFoM      float64 // fJ/conversion target
}

// DefaultSARConfig returns a default SAR ADC configuration for CIM
func DefaultSARConfig(bits int) *ADCConfig {
	return &ADCConfig{
		Type:         ADCTypeSAR,
		Resolution:   bits,
		SamplingRate: 100e6, // 100 MS/s
		Technology:   28,    // 28nm
		VDD:          0.9,   // 0.9V
		SARCycles:    bits,
		CapacitorDAC: true,
		TargetFoM:    10.0,  // 10 fJ/conv
	}
}

// DefaultInMemoryADCConfig returns IMADC configuration
func DefaultInMemoryADCConfig(bits int) *ADCConfig {
	return &ADCConfig{
		Type:         ADCTypeInMemory,
		Resolution:   bits,
		SamplingRate: 500e6, // 500 MS/s
		Technology:   28,
		VDD:          0.9,
		TargetFoM:    1.0, // 1 fJ/conv (much lower)
	}
}

// CIMADCUnit models an ADC unit for CIM arrays
type CIMADCUnit struct {
	Config      *ADCConfig

	// Performance metrics
	Area        float64 // mm²
	EnergyPerOp float64 // fJ/conversion
	Latency     float64 // ns
	FoM         float64 // fJ/conversion-step (Walden FoM)

	// State
	ConversionCount int64
	TotalEnergy     float64 // pJ total
}

// NewCIMADCUnit creates a new ADC unit with calculated metrics
func NewCIMADCUnit(config *ADCConfig) *CIMADCUnit {
	adc := &CIMADCUnit{
		Config: config,
	}
	adc.calculateMetrics()
	return adc
}

// calculateMetrics calculates area, energy, and performance metrics
func (adc *CIMADCUnit) calculateMetrics() {
	// Technology scaling factors
	techScale := math.Pow(float64(adc.Config.Technology)/28.0, 2)

	switch adc.Config.Type {
	case ADCTypeSAR:
		// SAR ADC: Area scales with 2^N capacitors
		// Typical: 4.3 fJ/conv for 6-bit at 28nm
		baseArea := 0.000159 // mm² for 6-bit at 28nm
		adc.Area = baseArea * math.Pow(2, float64(adc.Config.Resolution-6)) * techScale

		// Energy: ~4-50 fJ/conv depending on resolution
		baseFoM := 4.3 // fJ/conv-step for state-of-art
		adc.FoM = baseFoM * math.Pow(1.5, float64(adc.Config.Resolution-6))
		adc.EnergyPerOp = adc.FoM * float64(adc.Config.Resolution)

		// Latency: N cycles for N-bit SAR
		clockPeriod := 1e9 / adc.Config.SamplingRate // ns
		adc.Latency = clockPeriod * float64(adc.Config.SARCycles)

	case ADCTypeFlash:
		// Flash ADC: 2^N - 1 comparators
		numComparators := int(math.Pow(2, float64(adc.Config.Resolution))) - 1
		adc.Config.FlashComparators = numComparators

		// Area: linear with comparators
		compArea := 0.00001 * techScale // mm² per comparator
		adc.Area = compArea * float64(numComparators)

		// Energy: ~100-500 fJ/conv (much higher than SAR)
		adc.FoM = 50.0 * math.Pow(1.8, float64(adc.Config.Resolution-4))
		adc.EnergyPerOp = adc.FoM * float64(adc.Config.Resolution)

		// Latency: single cycle
		adc.Latency = 1e9 / adc.Config.SamplingRate

	case ADCTypeInMemory:
		// In-Memory ADC (IMADC): uses NVM devices for conversion
		// Area: 45 μm² = 0.000045 mm² (from research)
		adc.Area = 0.000045 * techScale

		// Energy: 29.6 fJ (from research)
		adc.EnergyPerOp = 29.6
		adc.FoM = adc.EnergyPerOp / float64(adc.Config.Resolution)

		// Latency: comparable to SAR
		adc.Latency = float64(adc.Config.Resolution) * 2 // ns

	case ADCTypeAdaptive:
		// Adaptive resolution ADC
		avgBits := float64(adc.Config.AdaptiveMinBits+adc.Config.AdaptiveMaxBits) / 2
		baseArea := 0.0002 * techScale
		adc.Area = baseArea * math.Pow(2, avgBits-6)

		// Energy varies with actual resolution used
		adc.FoM = 8.0 // Slightly higher overhead for adaptation
		adc.EnergyPerOp = adc.FoM * avgBits

		// Latency: worst case
		adc.Latency = float64(adc.Config.AdaptiveMaxBits) * 1e9 / adc.Config.SamplingRate

	case ADCTypeTimeDomain:
		// Time-domain ADC (VCO-based)
		adc.Area = 0.0001 * techScale
		adc.FoM = 6.0 // 6.02 fJ/conv from research
		adc.EnergyPerOp = adc.FoM * float64(adc.Config.Resolution)
		adc.Latency = 10.0 // ns
	}
}

// Convert performs ADC conversion with noise modeling
func (adc *CIMADCUnit) Convert(analogValue float64, fullScale float64) (int, float64) {
	adc.ConversionCount++
	adc.TotalEnergy += adc.EnergyPerOp / 1000 // Convert fJ to pJ

	// Normalize to full scale
	normalized := analogValue / fullScale
	if normalized < 0 {
		normalized = 0
	} else if normalized > 1 {
		normalized = 1
	}

	// Quantization
	levels := int(math.Pow(2, float64(adc.Config.Resolution)))
	quantized := int(normalized * float64(levels-1))

	// Add quantization noise
	lsb := fullScale / float64(levels)
	quantError := analogValue - (float64(quantized)+0.5)*lsb

	// Add thermal noise (simplified model)
	thermalNoise := rand.NormFloat64() * lsb * 0.1 // 10% LSB noise

	return quantized, quantError + thermalNoise
}

// =============================================================================
// CIM ADC OPTIMIZATION FRAMEWORK
// =============================================================================

// ADCOptimizationConfig configures ADC optimization for CIM
type ADCOptimizationConfig struct {
	// Crossbar parameters
	CrossbarRows    int
	CrossbarCols    int
	WeightBits      int
	InputBits       int

	// Output range
	MaxPartialSum   float64 // Maximum possible partial sum
	OutputBits      int     // Required output bits

	// Constraints
	MaxArea         float64 // mm² budget for ADCs
	MaxPower        float64 // mW budget
	MinThroughput   float64 // TOPS requirement

	// Optimization targets
	OptimizeFor     string  // "energy", "area", "throughput", "balanced"
}

// ADCOptimizationResult contains optimization results
type ADCOptimizationResult struct {
	OptimalType     ADCType
	OptimalBits     int
	NumADCs         int
	TotalArea       float64 // mm²
	TotalPower      float64 // mW
	Throughput      float64 // TOPS
	EnergyEff       float64 // TOPS/W
	AreaEff         float64 // TOPS/mm²
}

// ADCOptimizer optimizes ADC configuration for CIM arrays
type ADCOptimizer struct {
	Config *ADCOptimizationConfig
}

// NewADCOptimizer creates a new ADC optimizer
func NewADCOptimizer(config *ADCOptimizationConfig) *ADCOptimizer {
	return &ADCOptimizer{Config: config}
}

// Optimize finds optimal ADC configuration
func (opt *ADCOptimizer) Optimize() *ADCOptimizationResult {
	best := &ADCOptimizationResult{}
	bestScore := math.Inf(-1)

	// Explore design space
	adcTypes := []ADCType{ADCTypeSAR, ADCTypeFlash, ADCTypeInMemory, ADCTypeAdaptive}
	bitOptions := []int{4, 5, 6, 7, 8}

	for _, adcType := range adcTypes {
		for _, bits := range bitOptions {
			// Skip invalid combinations
			if bits > opt.Config.OutputBits {
				continue
			}

			result := opt.evaluateConfig(adcType, bits)

			// Check constraints
			if result.TotalArea > opt.Config.MaxArea {
				continue
			}
			if result.TotalPower > opt.Config.MaxPower {
				continue
			}
			if result.Throughput < opt.Config.MinThroughput {
				continue
			}

			// Calculate score based on optimization target
			score := opt.calculateScore(result)
			if score > bestScore {
				bestScore = score
				best = result
			}
		}
	}

	return best
}

// evaluateConfig evaluates a specific ADC configuration
func (opt *ADCOptimizer) evaluateConfig(adcType ADCType, bits int) *ADCOptimizationResult {
	var config *ADCConfig
	switch adcType {
	case ADCTypeSAR:
		config = DefaultSARConfig(bits)
	case ADCTypeInMemory:
		config = DefaultInMemoryADCConfig(bits)
	default:
		config = &ADCConfig{
			Type:       adcType,
			Resolution: bits,
			Technology: 28,
			VDD:        0.9,
		}
	}

	adc := NewCIMADCUnit(config)

	// Number of ADCs = number of columns (one per output)
	numADCs := opt.Config.CrossbarCols

	// Calculate total metrics
	totalArea := adc.Area * float64(numADCs)

	// Power: energy per op × operations per second
	opsPerSecond := 1e9 / adc.Latency * float64(numADCs)
	totalPower := adc.EnergyPerOp * opsPerSecond / 1e9 // fJ × ops/s / 1e9 = mW

	// Throughput: MACs per second
	macsPerOp := float64(opt.Config.CrossbarRows) // Each ADC conversion = rows MACs
	throughput := opsPerSecond * macsPerOp / 1e12 // TOPS

	return &ADCOptimizationResult{
		OptimalType: adcType,
		OptimalBits: bits,
		NumADCs:     numADCs,
		TotalArea:   totalArea,
		TotalPower:  totalPower,
		Throughput:  throughput,
		EnergyEff:   throughput / (totalPower / 1000), // TOPS/W
		AreaEff:     throughput / totalArea,           // TOPS/mm²
	}
}

// calculateScore calculates optimization score
func (opt *ADCOptimizer) calculateScore(result *ADCOptimizationResult) float64 {
	switch opt.Config.OptimizeFor {
	case "energy":
		return result.EnergyEff
	case "area":
		return result.AreaEff
	case "throughput":
		return result.Throughput
	case "balanced":
		// Geometric mean of normalized metrics
		energyNorm := result.EnergyEff / 100   // Normalize to ~1
		areaNorm := result.AreaEff / 1000      // Normalize to ~1
		throughputNorm := result.Throughput / 10 // Normalize to ~1
		return math.Pow(energyNorm*areaNorm*throughputNorm, 1.0/3.0)
	default:
		return result.EnergyEff
	}
}

// =============================================================================
// ADAPTIVE RESOLUTION ADC
// =============================================================================

// AdaptiveADCConfig configures an adaptive resolution ADC
type AdaptiveADCConfig struct {
	MinResolution    int     // Minimum bits (e.g., 4)
	MaxResolution    int     // Maximum bits (e.g., 8)
	LayerResolutions map[string]int // Per-layer resolution requirements

	// Adaptation strategy
	Strategy         string  // "layer-wise", "activation-based", "error-aware"
	ErrorThreshold   float64 // For error-aware adaptation
}

// AdaptiveADC implements an adaptive resolution ADC
type AdaptiveADC struct {
	Config         *AdaptiveADCConfig
	CurrentRes     int

	// Statistics
	TotalConversions  int64
	ResolutionHist    map[int]int64 // Histogram of resolutions used
	EnergySaved       float64       // fJ saved vs max resolution
}

// NewAdaptiveADC creates a new adaptive ADC
func NewAdaptiveADC(config *AdaptiveADCConfig) *AdaptiveADC {
	return &AdaptiveADC{
		Config:         config,
		CurrentRes:     config.MaxResolution,
		ResolutionHist: make(map[int]int64),
	}
}

// SetResolutionForLayer sets resolution based on layer requirements
func (adc *AdaptiveADC) SetResolutionForLayer(layerName string) {
	if res, ok := adc.Config.LayerResolutions[layerName]; ok {
		adc.CurrentRes = res
	} else {
		adc.CurrentRes = adc.Config.MaxResolution
	}
}

// AdaptResolution adapts resolution based on input statistics
func (adc *AdaptiveADC) AdaptResolution(inputRange float64, targetError float64) int {
	// Calculate minimum bits needed for target error
	// Error ≈ 1 / (2^bits) → bits = log2(1/error)
	minBitsNeeded := int(math.Ceil(math.Log2(1.0 / targetError)))

	// Clamp to allowed range
	if minBitsNeeded < adc.Config.MinResolution {
		minBitsNeeded = adc.Config.MinResolution
	}
	if minBitsNeeded > adc.Config.MaxResolution {
		minBitsNeeded = adc.Config.MaxResolution
	}

	adc.CurrentRes = minBitsNeeded
	return minBitsNeeded
}

// Convert performs adaptive conversion
func (adc *AdaptiveADC) Convert(value float64, fullScale float64) (int, float64) {
	adc.TotalConversions++
	adc.ResolutionHist[adc.CurrentRes]++

	// Calculate energy savings
	maxEnergy := 10.0 * float64(adc.Config.MaxResolution) // Assume 10 fJ/bit
	actualEnergy := 10.0 * float64(adc.CurrentRes)
	adc.EnergySaved += maxEnergy - actualEnergy

	// Perform quantization
	levels := int(math.Pow(2, float64(adc.CurrentRes)))
	normalized := value / fullScale
	if normalized < 0 {
		normalized = 0
	} else if normalized > 1 {
		normalized = 1
	}

	quantized := int(normalized * float64(levels-1))
	lsb := fullScale / float64(levels)
	error := value - (float64(quantized)+0.5)*lsb

	return quantized, error
}

// GetStatistics returns adaptive ADC statistics
func (adc *AdaptiveADC) GetStatistics() map[string]interface{} {
	avgResolution := 0.0
	for res, count := range adc.ResolutionHist {
		avgResolution += float64(res) * float64(count)
	}
	if adc.TotalConversions > 0 {
		avgResolution /= float64(adc.TotalConversions)
	}

	return map[string]interface{}{
		"total_conversions": adc.TotalConversions,
		"average_resolution": avgResolution,
		"energy_saved_fJ":   adc.EnergySaved,
		"resolution_histogram": adc.ResolutionHist,
	}
}

// =============================================================================
// ADC-LESS CIM APPROACHES
// =============================================================================

// ADCLessConfig configures ADC-less CIM approaches
type ADCLessConfig struct {
	Approach        string  // "time-domain", "stochastic", "binary-neural", "spiking"

	// Time-domain parameters
	VCOFrequency    float64 // Hz for VCO-based
	CounterBits     int     // Counter resolution

	// Stochastic parameters
	BitStreamLength int     // Length of stochastic bit streams

	// Binary neural network
	UseBinaryWeights bool
	UseBinaryActivations bool
}

// ADCLessCIM implements ADC-less compute-in-memory
type ADCLessCIM struct {
	Config          *ADCLessConfig

	// Energy metrics
	EnergyPerMAC    float64 // fJ/MAC
	AreaPerCell     float64 // mm²

	// Accuracy impact
	AccuracyLoss    float64 // Percentage accuracy loss vs full precision
}

// NewADCLessCIM creates a new ADC-less CIM unit
func NewADCLessCIM(config *ADCLessConfig) *ADCLessCIM {
	cim := &ADCLessCIM{
		Config: config,
	}
	cim.calculateMetrics()
	return cim
}

// calculateMetrics calculates energy and area metrics
func (cim *ADCLessCIM) calculateMetrics() {
	switch cim.Config.Approach {
	case "time-domain":
		// Time-domain: uses VCO + counter instead of ADC
		// Energy: ~28× reduction vs 7-bit ADC (from HCiM paper)
		cim.EnergyPerMAC = 0.5  // fJ/MAC
		cim.AreaPerCell = 0.00001 // mm²
		cim.AccuracyLoss = 1.0  // ~1% accuracy loss

	case "stochastic":
		// Stochastic computing: bit-serial multiplication
		cim.EnergyPerMAC = 0.1  // fJ/MAC (very low)
		cim.AreaPerCell = 0.000005
		cim.AccuracyLoss = 3.0  // Higher accuracy loss due to variance

	case "binary-neural":
		// Binary Neural Networks: 1-bit weights and activations
		cim.EnergyPerMAC = 0.05 // fJ/MAC (XNOR + popcount)
		cim.AreaPerCell = 0.000002
		cim.AccuracyLoss = 5.0  // Significant accuracy loss

	case "spiking":
		// Spiking Neural Networks: temporal encoding
		cim.EnergyPerMAC = 0.2  // fJ/spike
		cim.AreaPerCell = 0.00001
		cim.AccuracyLoss = 2.0  // Moderate accuracy loss
	}
}

// ComputeMAC performs ADC-less MAC operation
func (cim *ADCLessCIM) ComputeMAC(input, weight float64) float64 {
	switch cim.Config.Approach {
	case "time-domain":
		// Convert to time domain and back
		period := 1.0 / cim.Config.VCOFrequency
		cycles := int(input * weight / period)
		return float64(cycles) * period

	case "stochastic":
		// Stochastic multiplication via AND gate
		result := 0.0
		for i := 0; i < cim.Config.BitStreamLength; i++ {
			inBit := rand.Float64() < input
			wBit := rand.Float64() < weight
			if inBit && wBit {
				result += 1.0
			}
		}
		return result / float64(cim.Config.BitStreamLength)

	case "binary-neural":
		// XNOR multiplication
		inSign := input >= 0
		wSign := weight >= 0
		if inSign == wSign {
			return 1.0
		}
		return -1.0

	case "spiking":
		// Rate-coded spiking
		spikes := int(input * weight * 100) // Scale to spike count
		return float64(spikes) / 100.0

	default:
		return input * weight
	}
}

// =============================================================================
// INTEGRATED RADIATION-HARDENED CIM WITH ADC OPTIMIZATION
// =============================================================================

// RadHardCIMSystemConfig configures an integrated rad-hard CIM system
type RadHardCIMSystemConfig struct {
	// Crossbar config
	CrossbarRows    int
	CrossbarCols    int
	FerroConfig     *FerroelectricRadHardConfig

	// ADC config
	ADCConfig       *ADCConfig
	UseAdaptiveADC  bool
	UseADCLess      bool
	ADCLessConfig   *ADCLessConfig

	// Radiation config
	Environment     *RadiationEnvironment

	// System config
	UseTMR          bool
	UseECC          bool
	ScrubEnabled    bool
	ScrubInterval   time.Duration
}

// RadHardCIMSystem implements an integrated rad-hard CIM system
type RadHardCIMSystem struct {
	Config          *RadHardCIMSystemConfig
	Crossbar        *RadHardCrossbar
	ADC             *CIMADCUnit
	AdaptiveADC     *AdaptiveADC
	ADCLess         *ADCLessCIM

	// Statistics
	TotalMACs       int64
	TotalEnergy     float64 // pJ
	RadiationEvents int
	LastScrub       time.Time
}

// NewRadHardCIMSystem creates a new integrated rad-hard CIM system
func NewRadHardCIMSystem(config *RadHardCIMSystemConfig) *RadHardCIMSystem {
	cbConfig := &RadHardCrossbarConfig{
		Rows:       config.CrossbarRows,
		Cols:       config.CrossbarCols,
		CellConfig: config.FerroConfig,
		UseTMR:     config.UseTMR,
		UseECC:     config.UseECC,
	}

	system := &RadHardCIMSystem{
		Config:   config,
		Crossbar: NewRadHardCrossbar(cbConfig, config.Environment),
		LastScrub: time.Now(),
	}

	if config.UseADCLess {
		system.ADCLess = NewADCLessCIM(config.ADCLessConfig)
	} else if config.UseAdaptiveADC {
		system.AdaptiveADC = NewAdaptiveADC(&AdaptiveADCConfig{
			MinResolution: 4,
			MaxResolution: 8,
		})
	} else {
		system.ADC = NewCIMADCUnit(config.ADCConfig)
	}

	return system
}

// ComputeMVM performs matrix-vector multiplication with radiation effects
func (sys *RadHardCIMSystem) ComputeMVM(input []float64) []float64 {
	output := make([]float64, sys.Config.CrossbarCols)

	for j := 0; j < sys.Config.CrossbarCols; j++ {
		sum := 0.0
		for i := 0; i < sys.Config.CrossbarRows && i < len(input); i++ {
			// Get weight with TMR voting if enabled
			weight := sys.Crossbar.TMRVote(i, j) / 20.0 // Normalize from polarization

			if sys.Config.UseADCLess && sys.ADCLess != nil {
				sum += sys.ADCLess.ComputeMAC(input[i], weight)
			} else {
				sum += input[i] * weight
			}
		}

		// ADC conversion (if not ADC-less)
		if !sys.Config.UseADCLess {
			if sys.AdaptiveADC != nil {
				quantized, _ := sys.AdaptiveADC.Convert(sum, 10.0)
				output[j] = float64(quantized) * 10.0 / math.Pow(2, float64(sys.AdaptiveADC.CurrentRes))
			} else if sys.ADC != nil {
				quantized, _ := sys.ADC.Convert(sum, 10.0)
				output[j] = float64(quantized) * 10.0 / math.Pow(2, float64(sys.ADC.Config.Resolution))
			} else {
				output[j] = sum
			}
		} else {
			output[j] = sum
		}

		sys.TotalMACs += int64(sys.Config.CrossbarRows)
	}

	return output
}

// SimulateMission simulates radiation exposure during a mission
func (sys *RadHardCIMSystem) SimulateMission(duration time.Duration) *MissionSimResult {
	result := &MissionSimResult{
		Duration:    duration,
		StartTime:   time.Now(),
	}

	// Simulate radiation
	radResult := sys.Crossbar.SimulateRadiation(duration.Seconds())

	result.TIDAccumulated = radResult.EndTID - radResult.StartTID
	result.HeavyIonHits = radResult.HeavyIonHits
	result.CellFailures = radResult.CellFailures
	result.FunctionalCells = radResult.FunctionalCells
	result.SystemFunctional = float64(result.FunctionalCells) > 0.9*float64(sys.Config.CrossbarRows*sys.Config.CrossbarCols)

	return result
}

// MissionSimResult contains mission simulation results
type MissionSimResult struct {
	Duration        time.Duration
	StartTime       time.Time
	TIDAccumulated  float64
	HeavyIonHits    int
	CellFailures    int
	FunctionalCells int
	SystemFunctional bool
}

// GetSystemMetrics returns comprehensive system metrics
func (sys *RadHardCIMSystem) GetSystemMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"total_macs":      sys.TotalMACs,
		"total_energy_pJ": sys.TotalEnergy,
		"crossbar_size":   []int{sys.Config.CrossbarRows, sys.Config.CrossbarCols},
		"mission_time_s":  sys.Crossbar.MissionTime,
		"accumulated_TID": sys.Crossbar.Environment.AccumulatedTID,
		"total_ion_hits":  sys.Crossbar.TotalIonHits,
	}

	if sys.ADC != nil {
		metrics["adc_type"] = "standard"
		metrics["adc_bits"] = sys.ADC.Config.Resolution
		metrics["adc_energy_fJ"] = sys.ADC.EnergyPerOp
		metrics["adc_area_mm2"] = sys.ADC.Area
	} else if sys.AdaptiveADC != nil {
		metrics["adc_type"] = "adaptive"
		stats := sys.AdaptiveADC.GetStatistics()
		for k, v := range stats {
			metrics["adaptive_"+k] = v
		}
	} else if sys.ADCLess != nil {
		metrics["adc_type"] = "adc-less"
		metrics["adcless_approach"] = sys.ADCLess.Config.Approach
		metrics["energy_per_mac_fJ"] = sys.ADCLess.EnergyPerMAC
	}

	return metrics
}
