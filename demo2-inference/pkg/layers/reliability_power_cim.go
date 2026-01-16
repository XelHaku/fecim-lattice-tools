// reliability_power_cim.go - CIM Reliability, Variability Modeling, and Power Management
// Implements device variation simulation, fatigue/endurance modeling, and DVFS strategies
// Based on FeFET reliability research and CIM power optimization techniques

package layers

import (
	"fmt"
	"math"
	"sort"
)

// ============================================================================
// DEVICE VARIABILITY MODELING
// ============================================================================

// VariabilityConfig configures device variation simulation
type VariabilityConfig struct {
	// Device-to-device (D2D) variation
	D2DConductanceSigma  float64 // σ for conductance variation (%)
	D2DThresholdSigma    float64 // σ for Vth variation (mV)
	D2DOnOffRatioSigma   float64 // σ for on/off ratio variation (%)

	// Cycle-to-cycle (C2C) variation
	C2CConductanceSigma  float64 // σ for write cycle variation (%)
	C2CThresholdSigma    float64 // σ for threshold drift (mV)

	// Read noise
	ReadNoiseSigma       float64 // σ for read noise (%)
	ThermalNoiseKT       float64 // kT for thermal noise (eV)

	// Stuck-at faults
	SAFRate              float64 // Stuck-at fault rate (probability)
	SAFHighRatio         float64 // Ratio of stuck-at-high vs stuck-at-low

	// Retention
	RetentionDecayRate   float64 // Conductance decay rate per year
	RetentionTempCoeff   float64 // Temperature coefficient

	// Device type
	DeviceType           string  // "fefet", "reram", "pcm", "mram"
}

// DefaultFeFETVariabilityConfig returns typical FeFET variation parameters
func DefaultFeFETVariabilityConfig() *VariabilityConfig {
	return &VariabilityConfig{
		D2DConductanceSigma:  5.0,   // 5% D2D
		D2DThresholdSigma:    50.0,  // 50mV Vth spread
		D2DOnOffRatioSigma:   10.0,  // 10% on/off variation
		C2CConductanceSigma:  2.0,   // 2% C2C
		C2CThresholdSigma:    20.0,  // 20mV cycle drift
		ReadNoiseSigma:       1.0,   // 1% read noise
		ThermalNoiseKT:       0.026, // Room temperature
		SAFRate:              0.001, // 0.1% SAF rate
		SAFHighRatio:         0.5,   // Equal high/low SAF
		RetentionDecayRate:   0.01,  // 1% per year
		RetentionTempCoeff:   0.002, // 0.2% per °C
		DeviceType:           "fefet",
	}
}

// DefaultReRAMVariabilityConfig returns typical ReRAM variation parameters
func DefaultReRAMVariabilityConfig() *VariabilityConfig {
	return &VariabilityConfig{
		D2DConductanceSigma:  15.0,  // Higher D2D for ReRAM
		D2DThresholdSigma:    100.0,
		D2DOnOffRatioSigma:   20.0,
		C2CConductanceSigma:  5.0,   // Higher C2C
		C2CThresholdSigma:    50.0,
		ReadNoiseSigma:       3.0,   // Higher read noise
		ThermalNoiseKT:       0.026,
		SAFRate:              0.005, // Higher SAF rate
		SAFHighRatio:         0.6,   // More stuck-at-high
		RetentionDecayRate:   0.02,
		RetentionTempCoeff:   0.005,
		DeviceType:           "reram",
	}
}

// VariabilityModel simulates device variations
type VariabilityModel struct {
	Config       *VariabilityConfig
	D2DMap       [][]float64 // Device-to-device variation factors
	SAFMap       [][]int     // Stuck-at fault map (-1=low, 0=none, 1=high)
	AgeMap       [][]float64 // Device age in cycles
	Stats        *VariabilityStats
}

// VariabilityStats tracks variation statistics
type VariabilityStats struct {
	MeanConductanceError   float64
	MaxConductanceError    float64
	SAFCount               int
	RetentionLoss          float64
	EffectiveBits          float64 // Effective precision after variation
}

// NewVariabilityModel creates a new variability model
func NewVariabilityModel(config *VariabilityConfig, rows, cols int) *VariabilityModel {
	vm := &VariabilityModel{
		Config: config,
		D2DMap: make([][]float64, rows),
		SAFMap: make([][]int, rows),
		AgeMap: make([][]float64, rows),
		Stats:  &VariabilityStats{},
	}

	// Initialize D2D variation map
	for i := 0; i < rows; i++ {
		vm.D2DMap[i] = make([]float64, cols)
		vm.SAFMap[i] = make([]int, cols)
		vm.AgeMap[i] = make([]float64, cols)

		for j := 0; j < cols; j++ {
			// Sample D2D variation (Gaussian)
			vm.D2DMap[i][j] = 1.0 + randGaussian()*config.D2DConductanceSigma/100.0

			// Sample SAF
			if randFloat() < config.SAFRate {
				if randFloat() < config.SAFHighRatio {
					vm.SAFMap[i][j] = 1 // Stuck-at-high
				} else {
					vm.SAFMap[i][j] = -1 // Stuck-at-low
				}
				vm.Stats.SAFCount++
			}
		}
	}

	return vm
}

// ApplyVariation applies device variations to weight matrix
func (vm *VariabilityModel) ApplyVariation(weights [][]float64, temperature float64) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])

	result := make([][]float64, rows)
	totalError := 0.0
	maxError := 0.0

	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			w := weights[i][j]

			// Apply D2D variation
			w *= vm.D2DMap[i][j]

			// Apply C2C variation
			c2cNoise := randGaussian() * vm.Config.C2CConductanceSigma / 100.0
			w *= (1.0 + c2cNoise)

			// Apply read noise
			readNoise := randGaussian() * vm.Config.ReadNoiseSigma / 100.0
			w *= (1.0 + readNoise)

			// Apply thermal noise
			thermalNoise := math.Sqrt(2.0*vm.Config.ThermalNoiseKT*temperature/300.0) * randGaussian() * 0.01
			w += thermalNoise

			// Apply retention loss based on age
			retentionFactor := 1.0 - vm.Config.RetentionDecayRate*vm.AgeMap[i][j]
			retentionFactor *= 1.0 - vm.Config.RetentionTempCoeff*(temperature-300.0)
			w *= math.Max(retentionFactor, 0.5)

			// Apply SAF
			switch vm.SAFMap[i][j] {
			case 1:
				w = 1.0 // Stuck-at-high
			case -1:
				w = 0.0 // Stuck-at-low
			}

			result[i][j] = w

			// Track error statistics
			error := math.Abs(result[i][j] - weights[i][j])
			totalError += error
			if error > maxError {
				maxError = error
			}
		}
	}

	vm.Stats.MeanConductanceError = totalError / float64(rows*cols)
	vm.Stats.MaxConductanceError = maxError
	vm.computeEffectiveBits()

	return result
}

func (vm *VariabilityModel) computeEffectiveBits() {
	// Effective bits = log2(1/σ_total)
	totalSigma := math.Sqrt(
		math.Pow(vm.Config.D2DConductanceSigma/100.0, 2) +
			math.Pow(vm.Config.C2CConductanceSigma/100.0, 2) +
			math.Pow(vm.Config.ReadNoiseSigma/100.0, 2))

	if totalSigma > 0 {
		vm.Stats.EffectiveBits = -math.Log2(totalSigma)
	} else {
		vm.Stats.EffectiveBits = 8.0 // Max precision
	}
}

// ============================================================================
// ENDURANCE AND FATIGUE MODELING
// ============================================================================

// EnduranceConfig configures endurance/fatigue simulation
type EnduranceConfig struct {
	MaxEnduranceCycles   int64   // Maximum endurance cycles
	FatigueOnset         int64   // Cycles before fatigue begins
	FatigueRate          float64 // Degradation rate after onset

	// Fatigue mechanisms
	ChargeTrapDensity    float64 // Initial trap density (cm⁻²)
	TrapGenerationRate   float64 // Trap generation per cycle
	InterfaceDegradation float64 // Interface state degradation

	// Asymmetric behavior
	SetEndurance         int64   // SET operation endurance
	ResetEndurance       int64   // RESET operation endurance

	// Recovery
	RecoveryEnabled      bool
	RecoveryTime         float64 // Recovery time constant (s)
	RecoveryFactor       float64 // Maximum recovery fraction
}

// DefaultFeFETEnduranceConfig returns typical FeFET endurance parameters
func DefaultFeFETEnduranceConfig() *EnduranceConfig {
	return &EnduranceConfig{
		MaxEnduranceCycles:   1e12, // 10^12 cycles for FeFET
		FatigueOnset:         1e10,
		FatigueRate:          0.01,
		ChargeTrapDensity:    1e10,
		TrapGenerationRate:   1e-12,
		InterfaceDegradation: 0.001,
		SetEndurance:         1e12,
		ResetEndurance:       1e12,
		RecoveryEnabled:      true,
		RecoveryTime:         1.0,
		RecoveryFactor:       0.3,
	}
}

// DefaultReRAMEnduranceConfig returns typical ReRAM endurance parameters
func DefaultReRAMEnduranceConfig() *EnduranceConfig {
	return &EnduranceConfig{
		MaxEnduranceCycles:   1e6, // 10^6 cycles for ReRAM
		FatigueOnset:         1e5,
		FatigueRate:          0.05,
		ChargeTrapDensity:    1e11,
		TrapGenerationRate:   1e-6,
		InterfaceDegradation: 0.01,
		SetEndurance:         1e6,
		ResetEndurance:       5e5, // Asymmetric
		RecoveryEnabled:      false,
		RecoveryTime:         10.0,
		RecoveryFactor:       0.1,
	}
}

// EnduranceModel simulates device endurance and fatigue
type EnduranceModel struct {
	Config       *EnduranceConfig
	CycleCount   [][]int64   // Write cycles per device
	SetCount     [][]int64   // SET operation count
	ResetCount   [][]int64   // RESET operation count
	TrapDensity  [][]float64 // Current trap density
	LastAccess   [][]float64 // Last access time for recovery
	Stats        *EnduranceStats
}

// EnduranceStats tracks endurance statistics
type EnduranceStats struct {
	TotalCycles          int64
	MaxCycles            int64
	FailedDevices        int
	AverageTrapDensity   float64
	MemoryWindowLoss     float64 // % of original MW
	PredictedLifetime    float64 // Hours at current rate
}

// NewEnduranceModel creates a new endurance model
func NewEnduranceModel(config *EnduranceConfig, rows, cols int) *EnduranceModel {
	em := &EnduranceModel{
		Config:      config,
		CycleCount:  make([][]int64, rows),
		SetCount:    make([][]int64, rows),
		ResetCount:  make([][]int64, rows),
		TrapDensity: make([][]float64, rows),
		LastAccess:  make([][]float64, rows),
		Stats:       &EnduranceStats{},
	}

	for i := 0; i < rows; i++ {
		em.CycleCount[i] = make([]int64, cols)
		em.SetCount[i] = make([]int64, cols)
		em.ResetCount[i] = make([]int64, cols)
		em.TrapDensity[i] = make([]float64, cols)
		em.LastAccess[i] = make([]float64, cols)

		for j := 0; j < cols; j++ {
			em.TrapDensity[i][j] = config.ChargeTrapDensity
		}
	}

	return em
}

// RecordWrite records a write operation and updates fatigue
func (em *EnduranceModel) RecordWrite(row, col int, isSet bool, currentTime float64) bool {
	if row >= len(em.CycleCount) || col >= len(em.CycleCount[0]) {
		return false
	}

	// Apply recovery if enabled
	if em.Config.RecoveryEnabled {
		timeSinceAccess := currentTime - em.LastAccess[row][col]
		recoveryFactor := 1.0 - math.Exp(-timeSinceAccess/em.Config.RecoveryTime)
		em.TrapDensity[row][col] *= (1.0 - em.Config.RecoveryFactor*recoveryFactor)
	}

	// Increment cycle counts
	em.CycleCount[row][col]++
	if isSet {
		em.SetCount[row][col]++
	} else {
		em.ResetCount[row][col]++
	}

	// Check for endurance failure
	if isSet && em.SetCount[row][col] > em.Config.SetEndurance {
		em.Stats.FailedDevices++
		return false
	}
	if !isSet && em.ResetCount[row][col] > em.Config.ResetEndurance {
		em.Stats.FailedDevices++
		return false
	}

	// Update trap density (fatigue)
	if em.CycleCount[row][col] > em.Config.FatigueOnset {
		em.TrapDensity[row][col] += em.Config.TrapGenerationRate
	}

	em.LastAccess[row][col] = currentTime
	em.Stats.TotalCycles++

	if em.CycleCount[row][col] > em.Stats.MaxCycles {
		em.Stats.MaxCycles = em.CycleCount[row][col]
	}

	return true
}

// ComputeFatigueFactor returns the degradation factor for a device
func (em *EnduranceModel) ComputeFatigueFactor(row, col int) float64 {
	cycles := em.CycleCount[row][col]

	if cycles < em.Config.FatigueOnset {
		return 1.0
	}

	// Exponential degradation after fatigue onset
	fatigueAmount := float64(cycles-em.Config.FatigueOnset) * em.Config.FatigueRate / 1e9
	return math.Max(1.0-fatigueAmount, 0.1)
}

// ApplyFatigue applies fatigue effects to weight matrix
func (em *EnduranceModel) ApplyFatigue(weights [][]float64) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])

	result := make([][]float64, rows)
	totalTrapDensity := 0.0
	totalMWLoss := 0.0

	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			fatigueFactor := em.ComputeFatigueFactor(i, j)
			result[i][j] = weights[i][j] * fatigueFactor

			totalTrapDensity += em.TrapDensity[i][j]
			totalMWLoss += (1.0 - fatigueFactor)
		}
	}

	em.Stats.AverageTrapDensity = totalTrapDensity / float64(rows*cols)
	em.Stats.MemoryWindowLoss = totalMWLoss / float64(rows*cols) * 100.0

	return result
}

// PredictLifetime estimates remaining device lifetime
func (em *EnduranceModel) PredictLifetime(writesPerSecond float64) float64 {
	remainingCycles := float64(em.Config.MaxEnduranceCycles - em.Stats.MaxCycles)
	if writesPerSecond <= 0 {
		return math.Inf(1)
	}
	return remainingCycles / writesPerSecond / 3600.0 // Hours
}

// ============================================================================
// NEURAL NETWORK ACCURACY UNDER VARIATION
// ============================================================================

// AccuracyAnalyzer analyzes NN accuracy under device variations
type AccuracyAnalyzer struct {
	VariabilityModel *VariabilityModel
	EnduranceModel   *EnduranceModel
	BaselineAccuracy float64
	Stats            *AccuracyStats
}

// AccuracyStats tracks accuracy degradation
type AccuracyStats struct {
	CurrentAccuracy      float64
	AccuracyDrop         float64
	NoiseTolerance       float64 // Maximum tolerable noise
	EffectivePrecision   int     // Bits of effective precision
	RetrainingRecommended bool
}

// NewAccuracyAnalyzer creates an accuracy analyzer
func NewAccuracyAnalyzer(vm *VariabilityModel, em *EnduranceModel) *AccuracyAnalyzer {
	return &AccuracyAnalyzer{
		VariabilityModel: vm,
		EnduranceModel:   em,
		BaselineAccuracy: 0.95, // 95% baseline
		Stats:            &AccuracyStats{},
	}
}

// EstimateAccuracyDrop estimates accuracy drop due to variations
func (aa *AccuracyAnalyzer) EstimateAccuracyDrop() float64 {
	// Empirical model based on research
	// Accuracy drops approximately quadratically with total variation

	totalVariation := aa.VariabilityModel.Stats.MeanConductanceError
	safPenalty := float64(aa.VariabilityModel.Stats.SAFCount) * 0.001
	fatiguePenalty := aa.EnduranceModel.Stats.MemoryWindowLoss * 0.01

	// Accuracy = baseline * (1 - k1*σ² - k2*SAF - k3*fatigue)
	k1 := 0.5  // Variation coefficient
	k2 := 1.0  // SAF coefficient
	k3 := 0.1  // Fatigue coefficient

	accuracyDrop := k1*totalVariation*totalVariation + k2*safPenalty + k3*fatiguePenalty
	accuracyDrop = math.Min(accuracyDrop, 0.5) // Cap at 50% drop

	aa.Stats.CurrentAccuracy = aa.BaselineAccuracy * (1.0 - accuracyDrop)
	aa.Stats.AccuracyDrop = accuracyDrop * 100.0

	// Determine if retraining is needed (>5% drop)
	aa.Stats.RetrainingRecommended = accuracyDrop > 0.05

	return aa.Stats.AccuracyDrop
}

// ComputeNoiseTolerance finds maximum tolerable noise level
func (aa *AccuracyAnalyzer) ComputeNoiseTolerance(targetAccuracyDrop float64) float64 {
	// Binary search for noise tolerance
	low, high := 0.0, 0.5

	for high-low > 0.001 {
		mid := (low + high) / 2.0

		// Temporary variation model
		testConfig := *aa.VariabilityModel.Config
		testConfig.D2DConductanceSigma = mid * 100.0

		drop := mid * mid * 0.5 // Simplified model
		if drop < targetAccuracyDrop {
			low = mid
		} else {
			high = mid
		}
	}

	aa.Stats.NoiseTolerance = (low + high) / 2.0 * 100.0
	return aa.Stats.NoiseTolerance
}

// ============================================================================
// POWER MANAGEMENT - DYNAMIC VOLTAGE FREQUENCY SCALING (DVFS)
// ============================================================================

// DVFSConfig configures dynamic voltage/frequency scaling
type DVFSConfig struct {
	// Voltage levels
	VoltageMin       float64 // Minimum operating voltage (V)
	VoltageNom       float64 // Nominal voltage (V)
	VoltageMax       float64 // Maximum voltage (V)
	VoltageLevels    []float64

	// Frequency levels
	FrequencyMin     float64 // Minimum frequency (MHz)
	FrequencyNom     float64 // Nominal frequency (MHz)
	FrequencyMax     float64 // Maximum frequency (MHz)
	FrequencyLevels  []float64

	// Power model
	DynamicPowerCoeff float64 // C_eff * f (pF)
	LeakagePowerBase  float64 // Leakage power at Vnom (mW)
	LeakageVoltageExp float64 // Leakage voltage exponent

	// Timing constraints
	TransitionTimeUS  float64 // Voltage transition time (µs)
	SettlingTimeUS    float64 // Frequency settling time (µs)

	// Policy
	Policy            string  // "performance", "balanced", "powersave"
	TargetUtilization float64 // Target utilization for balanced mode
}

// DefaultDVFSConfig returns default DVFS configuration
func DefaultDVFSConfig() *DVFSConfig {
	return &DVFSConfig{
		VoltageMin:        0.6,
		VoltageNom:        1.0,
		VoltageMax:        1.1,
		VoltageLevels:     []float64{0.6, 0.7, 0.8, 0.9, 1.0, 1.1},
		FrequencyMin:      100.0,
		FrequencyNom:      500.0,
		FrequencyMax:      1000.0,
		FrequencyLevels:   []float64{100, 200, 300, 400, 500, 750, 1000},
		DynamicPowerCoeff: 0.1,
		LeakagePowerBase:  10.0,
		LeakageVoltageExp: 2.0,
		TransitionTimeUS:  10.0,
		SettlingTimeUS:    5.0,
		Policy:            "balanced",
		TargetUtilization: 0.7,
	}
}

// DVFSController manages voltage/frequency scaling
type DVFSController struct {
	Config           *DVFSConfig
	CurrentVoltage   float64
	CurrentFrequency float64
	OperatingPoint   int     // Index into V/F levels
	Stats            *DVFSStats
}

// DVFSStats tracks DVFS statistics
type DVFSStats struct {
	AveragePower     float64
	PeakPower        float64
	EnergySaved      float64 // Compared to max V/F
	Transitions      int
	TimeInEachState  []float64
}

// NewDVFSController creates a DVFS controller
func NewDVFSController(config *DVFSConfig) *DVFSController {
	return &DVFSController{
		Config:           config,
		CurrentVoltage:   config.VoltageNom,
		CurrentFrequency: config.FrequencyNom,
		OperatingPoint:   len(config.VoltageLevels) / 2,
		Stats: &DVFSStats{
			TimeInEachState: make([]float64, len(config.VoltageLevels)),
		},
	}
}

// ComputePower calculates power at given V/F point
func (dvfs *DVFSController) ComputePower(voltage, frequency float64) float64 {
	// Dynamic power: P_dyn = C_eff * V² * f
	dynamicPower := dvfs.Config.DynamicPowerCoeff * voltage * voltage * frequency

	// Leakage power: P_leak = P_leak_base * (V/V_nom)^exp
	leakagePower := dvfs.Config.LeakagePowerBase *
		math.Pow(voltage/dvfs.Config.VoltageNom, dvfs.Config.LeakageVoltageExp)

	return dynamicPower + leakagePower
}

// SelectOperatingPoint chooses optimal V/F based on workload
func (dvfs *DVFSController) SelectOperatingPoint(utilization float64, deadline float64) (float64, float64) {
	switch dvfs.Config.Policy {
	case "performance":
		return dvfs.Config.VoltageMax, dvfs.Config.FrequencyMax
	case "powersave":
		return dvfs.Config.VoltageMin, dvfs.Config.FrequencyMin
	case "balanced":
		return dvfs.selectBalancedPoint(utilization, deadline)
	default:
		return dvfs.Config.VoltageNom, dvfs.Config.FrequencyNom
	}
}

func (dvfs *DVFSController) selectBalancedPoint(utilization, deadline float64) (float64, float64) {
	// Find minimum V/F that meets deadline while minimizing power

	bestPower := math.Inf(1)
	bestV, bestF := dvfs.Config.VoltageNom, dvfs.Config.FrequencyNom

	for _, v := range dvfs.Config.VoltageLevels {
		for _, f := range dvfs.Config.FrequencyLevels {
			// Check if this V/F can meet deadline
			// Assume execution time scales inversely with frequency
			execTime := utilization * (dvfs.Config.FrequencyNom / f)

			if execTime <= deadline {
				power := dvfs.ComputePower(v, f)
				if power < bestPower {
					bestPower = power
					bestV, bestF = v, f
				}
			}
		}
	}

	return bestV, bestF
}

// TransitionTo transitions to new V/F point
func (dvfs *DVFSController) TransitionTo(newVoltage, newFrequency float64) float64 {
	// Voltage scaling: raise voltage before frequency, lower after
	transitionEnergy := 0.0

	if newVoltage > dvfs.CurrentVoltage {
		// Scaling up: voltage first
		dvfs.CurrentVoltage = newVoltage
		transitionEnergy += dvfs.Config.TransitionTimeUS * dvfs.ComputePower(newVoltage, dvfs.CurrentFrequency) / 1000.0
		dvfs.CurrentFrequency = newFrequency
	} else {
		// Scaling down: frequency first
		dvfs.CurrentFrequency = newFrequency
		transitionEnergy += dvfs.Config.SettlingTimeUS * dvfs.ComputePower(dvfs.CurrentVoltage, newFrequency) / 1000.0
		dvfs.CurrentVoltage = newVoltage
	}

	dvfs.Stats.Transitions++
	return transitionEnergy
}

// ============================================================================
// ADC POWER OPTIMIZATION
// ============================================================================

// ADCPowerConfig configures ADC power optimization
type ADCPowerConfig struct {
	BaseResolution    int     // Base ADC resolution (bits)
	MaxResolution     int     // Maximum resolution
	MinResolution     int     // Minimum resolution
	PowerPerBit       float64 // Power scaling per bit (mW)
	AreaPerBit        float64 // Area scaling per bit (mm²)
	LatencyPerBit     float64 // Latency scaling per bit (ns)

	// Adaptive schemes
	AdaptiveEnabled   bool
	ConfidenceThresh  float64 // Confidence threshold for early termination
	SparsityThresh    float64 // Sparsity threshold for skipping
}

// DefaultADCPowerConfig returns default ADC configuration
func DefaultADCPowerConfig() *ADCPowerConfig {
	return &ADCPowerConfig{
		BaseResolution:   6,
		MaxResolution:    8,
		MinResolution:    2,
		PowerPerBit:      0.5, // 0.5mW per bit
		AreaPerBit:       0.001,
		LatencyPerBit:    5.0,
		AdaptiveEnabled:  true,
		ConfidenceThresh: 0.9,
		SparsityThresh:   0.5,
	}
}

// AdaptiveADC implements adaptive precision ADC
type AdaptiveADC struct {
	Config    *ADCPowerConfig
	Stats     *ADCStats
}

// ADCStats tracks ADC statistics
type ADCStats struct {
	TotalConversions   int64
	FullResConversions int64
	EarlyTerminations  int64
	SkippedConversions int64
	AverageResolution  float64
	TotalEnergy        float64
	EnergySaved        float64
}

// NewAdaptiveADC creates an adaptive ADC
func NewAdaptiveADC(config *ADCPowerConfig) *AdaptiveADC {
	return &AdaptiveADC{
		Config: config,
		Stats:  &ADCStats{},
	}
}

// Convert performs ADC conversion with adaptive precision
func (adc *AdaptiveADC) Convert(analogValue float64, confidence float64, sparsity float64) (int, float64) {
	adc.Stats.TotalConversions++

	// Check for sparsity-based skipping
	if adc.Config.AdaptiveEnabled && sparsity > adc.Config.SparsityThresh {
		adc.Stats.SkippedConversions++
		adc.Stats.EnergySaved += adc.computeEnergy(adc.Config.BaseResolution)
		return 0, 0
	}

	// Determine resolution based on confidence
	resolution := adc.Config.BaseResolution
	if adc.Config.AdaptiveEnabled && confidence > adc.Config.ConfidenceThresh {
		// Use lower resolution for high-confidence values
		resolution = adc.Config.MinResolution +
			int(float64(adc.Config.BaseResolution-adc.Config.MinResolution)*(1.0-confidence))
		adc.Stats.EarlyTerminations++
	} else {
		adc.Stats.FullResConversions++
	}

	// Quantize value
	levels := 1 << resolution
	quantized := int(math.Round(analogValue * float64(levels-1)))
	quantized = max(0, min(quantized, levels-1))

	// Compute energy
	energy := adc.computeEnergy(resolution)
	adc.Stats.TotalEnergy += energy
	adc.Stats.EnergySaved += adc.computeEnergy(adc.Config.BaseResolution) - energy

	// Update average resolution
	n := float64(adc.Stats.TotalConversions)
	adc.Stats.AverageResolution = (adc.Stats.AverageResolution*(n-1) + float64(resolution)) / n

	return quantized, energy
}

func (adc *AdaptiveADC) computeEnergy(resolution int) float64 {
	// Energy scales exponentially with resolution
	return adc.Config.PowerPerBit * math.Pow(2, float64(resolution-adc.Config.MinResolution)) *
		adc.Config.LatencyPerBit / 1e6 // Convert to µJ
}

// ============================================================================
// CIM POWER MODEL
// ============================================================================

// CIMPowerModel models overall CIM power consumption
type CIMPowerModel struct {
	DVFSController *DVFSController
	AdaptiveADC    *AdaptiveADC
	ArrayConfig    *CIMArrayPowerConfig
	Stats          *CIMPowerStats
}

// CIMArrayPowerConfig configures array-level power
type CIMArrayPowerConfig struct {
	NumArrays        int
	ArrayRows        int
	ArrayCols        int
	CellReadPower    float64 // Power per cell read (nW)
	CellWritePower   float64 // Power per cell write (nW)
	WordlineDriver   float64 // Wordline driver power (µW)
	BitlineDriver    float64 // Bitline driver power (µW)
	SenseAmpPower    float64 // Sense amplifier power (µW)
	PeripheralPower  float64 // Peripheral circuits (mW)
	LeakagePower     float64 // Static leakage (mW)
}

// DefaultCIMArrayPowerConfig returns default array power config
func DefaultCIMArrayPowerConfig() *CIMArrayPowerConfig {
	return &CIMArrayPowerConfig{
		NumArrays:       64,
		ArrayRows:       128,
		ArrayCols:       128,
		CellReadPower:   0.1,  // 0.1 nW
		CellWritePower:  10.0, // 10 nW
		WordlineDriver:  10.0, // 10 µW
		BitlineDriver:   5.0,  // 5 µW
		SenseAmpPower:   20.0, // 20 µW
		PeripheralPower: 5.0,  // 5 mW
		LeakagePower:    2.0,  // 2 mW
	}
}

// CIMPowerStats tracks CIM power statistics
type CIMPowerStats struct {
	TotalEnergyNJ     float64
	AveragePowerMW    float64
	PeakPowerMW       float64
	ArrayEnergyNJ     float64
	ADCEnergyNJ       float64
	PeripheralEnergyNJ float64
	TOPSW             float64 // TOPS/W efficiency
}

// NewCIMPowerModel creates a CIM power model
func NewCIMPowerModel(dvfsConfig *DVFSConfig, adcConfig *ADCPowerConfig, arrayConfig *CIMArrayPowerConfig) *CIMPowerModel {
	return &CIMPowerModel{
		DVFSController: NewDVFSController(dvfsConfig),
		AdaptiveADC:    NewAdaptiveADC(adcConfig),
		ArrayConfig:    arrayConfig,
		Stats:          &CIMPowerStats{},
	}
}

// ComputeMVMEnergy computes energy for one MVM operation
func (pm *CIMPowerModel) ComputeMVMEnergy(activeRows, activeCols int) float64 {
	cfg := pm.ArrayConfig

	// Array energy
	cellEnergy := float64(activeRows*activeCols) * cfg.CellReadPower * 1e-6 // nW to mW
	wordlineEnergy := float64(activeRows) * cfg.WordlineDriver * 1e-3
	bitlineEnergy := float64(activeCols) * cfg.BitlineDriver * 1e-3

	arrayEnergy := cellEnergy + wordlineEnergy + bitlineEnergy

	// ADC energy (one conversion per column)
	adcEnergy := float64(activeCols) * pm.AdaptiveADC.Config.PowerPerBit *
		float64(pm.AdaptiveADC.Config.BaseResolution) * 1e-3

	// Peripheral energy
	peripheralEnergy := cfg.PeripheralPower * 0.001 // 1µs operation

	totalEnergy := (arrayEnergy + adcEnergy + peripheralEnergy) * 1e6 // Convert to nJ

	pm.Stats.ArrayEnergyNJ += arrayEnergy * 1e6
	pm.Stats.ADCEnergyNJ += adcEnergy * 1e6
	pm.Stats.PeripheralEnergyNJ += peripheralEnergy * 1e6
	pm.Stats.TotalEnergyNJ += totalEnergy

	return totalEnergy
}

// ComputeTOPSW computes energy efficiency
func (pm *CIMPowerModel) ComputeTOPSW(macs int64, timeUS float64) float64 {
	if pm.Stats.TotalEnergyNJ == 0 {
		return 0
	}

	// TOPS = MACs / time / 1e12
	tops := float64(macs) / (timeUS * 1e-6) / 1e12

	// Power = Energy / time
	powerW := pm.Stats.TotalEnergyNJ * 1e-9 / (timeUS * 1e-6)

	pm.Stats.TOPSW = tops / powerW
	pm.Stats.AveragePowerMW = powerW * 1000

	return pm.Stats.TOPSW
}

// ============================================================================
// WRITE VERIFY AND CALIBRATION
// ============================================================================

// WriteVerifyConfig configures write-verify schemes
type WriteVerifyConfig struct {
	MaxIterations    int     // Maximum write-verify iterations
	TargetPrecision  float64 // Target conductance precision
	VerifyThreshold  float64 // Verify pass threshold
	StepSize         float64 // Programming step size
	AdaptiveStep     bool    // Use adaptive step size
}

// WriteVerifyController implements iterative write-verify
type WriteVerifyController struct {
	Config *WriteVerifyConfig
	Stats  *WriteVerifyStats
}

// WriteVerifyStats tracks write-verify statistics
type WriteVerifyStats struct {
	TotalWrites       int64
	VerifyPasses      int64
	VerifyFails       int64
	AverageIterations float64
	EnergyOverhead    float64 // % overhead from verify
}

// NewWriteVerifyController creates a write-verify controller
func NewWriteVerifyController(config *WriteVerifyConfig) *WriteVerifyController {
	return &WriteVerifyController{
		Config: config,
		Stats:  &WriteVerifyStats{},
	}
}

// ProgramWithVerify programs a cell with verification
func (wv *WriteVerifyController) ProgramWithVerify(current, target float64) (float64, int, bool) {
	value := current
	stepSize := wv.Config.StepSize

	for iter := 0; iter < wv.Config.MaxIterations; iter++ {
		// Check if within tolerance
		error := math.Abs(value - target)
		if error < wv.Config.TargetPrecision {
			wv.Stats.VerifyPasses++
			wv.Stats.TotalWrites++
			wv.updateAverageIterations(iter + 1)
			return value, iter + 1, true
		}

		// Adjust step size adaptively
		if wv.Config.AdaptiveStep {
			stepSize = math.Min(error, wv.Config.StepSize)
		}

		// Apply programming pulse
		if value < target {
			value += stepSize * (0.8 + 0.4*randFloat()) // With variation
		} else {
			value -= stepSize * (0.8 + 0.4*randFloat())
		}
	}

	wv.Stats.VerifyFails++
	wv.Stats.TotalWrites++
	wv.updateAverageIterations(wv.Config.MaxIterations)
	return value, wv.Config.MaxIterations, false
}

func (wv *WriteVerifyController) updateAverageIterations(iterations int) {
	n := float64(wv.Stats.TotalWrites)
	wv.Stats.AverageIterations = (wv.Stats.AverageIterations*(n-1) + float64(iterations)) / n
}

// ============================================================================
// RELIABILITY-AWARE INFERENCE
// ============================================================================

// ReliabilityAwareInference combines all reliability features
type ReliabilityAwareInference struct {
	VariabilityModel  *VariabilityModel
	EnduranceModel    *EnduranceModel
	AccuracyAnalyzer  *AccuracyAnalyzer
	PowerModel        *CIMPowerModel
	WriteVerify       *WriteVerifyController
	Stats             *ReliabilityStats
}

// ReliabilityStats combines all reliability statistics
type ReliabilityStats struct {
	InferenceCount    int64
	AccuracyHistory   []float64
	PowerHistory      []float64
	PredictedFailures int
	MaintenanceNeeded bool
}

// NewReliabilityAwareInference creates a reliability-aware inference engine
func NewReliabilityAwareInference(rows, cols int) *ReliabilityAwareInference {
	vmConfig := DefaultFeFETVariabilityConfig()
	emConfig := DefaultFeFETEnduranceConfig()

	vm := NewVariabilityModel(vmConfig, rows, cols)
	em := NewEnduranceModel(emConfig, rows, cols)

	return &ReliabilityAwareInference{
		VariabilityModel: vm,
		EnduranceModel:   em,
		AccuracyAnalyzer: NewAccuracyAnalyzer(vm, em),
		PowerModel: NewCIMPowerModel(
			DefaultDVFSConfig(),
			DefaultADCPowerConfig(),
			DefaultCIMArrayPowerConfig(),
		),
		WriteVerify: NewWriteVerifyController(&WriteVerifyConfig{
			MaxIterations:   10,
			TargetPrecision: 0.01,
			VerifyThreshold: 0.02,
			StepSize:        0.05,
			AdaptiveStep:    true,
		}),
		Stats: &ReliabilityStats{
			AccuracyHistory: make([]float64, 0),
			PowerHistory:    make([]float64, 0),
		},
	}
}

// RunInference performs reliability-aware inference
func (rai *ReliabilityAwareInference) RunInference(weights [][]float64, inputs []float64) ([]float64, *InferenceReport) {
	rows := len(weights)
	if rows == 0 {
		return nil, nil
	}
	cols := len(weights[0])

	// Apply device variations
	variedWeights := rai.VariabilityModel.ApplyVariation(weights, 300.0) // Room temp

	// Apply fatigue effects
	fatiguedWeights := rai.EnduranceModel.ApplyFatigue(variedWeights)

	// Perform MVM
	outputs := make([]float64, cols)
	for j := 0; j < cols; j++ {
		sum := 0.0
		for i := 0; i < rows && i < len(inputs); i++ {
			sum += fatiguedWeights[i][j] * inputs[i]
		}
		outputs[j] = sum
	}

	// Compute power
	energy := rai.PowerModel.ComputeMVMEnergy(rows, cols)

	// Update statistics
	rai.Stats.InferenceCount++
	accuracyDrop := rai.AccuracyAnalyzer.EstimateAccuracyDrop()

	rai.Stats.AccuracyHistory = append(rai.Stats.AccuracyHistory, 100.0-accuracyDrop)
	rai.Stats.PowerHistory = append(rai.Stats.PowerHistory, energy)

	// Check maintenance needs
	if accuracyDrop > 10.0 || rai.EnduranceModel.Stats.FailedDevices > 10 {
		rai.Stats.MaintenanceNeeded = true
	}

	report := &InferenceReport{
		AccuracyDrop:     accuracyDrop,
		EnergyNJ:         energy,
		EffectiveBits:    rai.VariabilityModel.Stats.EffectiveBits,
		SAFCount:         rai.VariabilityModel.Stats.SAFCount,
		MemoryWindowLoss: rai.EnduranceModel.Stats.MemoryWindowLoss,
		FailedDevices:    rai.EnduranceModel.Stats.FailedDevices,
	}

	return outputs, report
}

// InferenceReport summarizes inference results
type InferenceReport struct {
	AccuracyDrop     float64
	EnergyNJ         float64
	EffectiveBits    float64
	SAFCount         int
	MemoryWindowLoss float64
	FailedDevices    int
}

// Helper functions
func randGaussian() float64 {
	// Box-Muller transform
	u1 := randFloat()
	u2 := randFloat()
	if u1 < 1e-10 {
		u1 = 1e-10
	}
	return math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
}

// SummaryReport generates a summary of reliability status
func (rai *ReliabilityAwareInference) SummaryReport() string {
	// Compute averages
	avgAccuracy := 0.0
	for _, acc := range rai.Stats.AccuracyHistory {
		avgAccuracy += acc
	}
	if len(rai.Stats.AccuracyHistory) > 0 {
		avgAccuracy /= float64(len(rai.Stats.AccuracyHistory))
	}

	avgPower := 0.0
	for _, p := range rai.Stats.PowerHistory {
		avgPower += p
	}
	if len(rai.Stats.PowerHistory) > 0 {
		avgPower /= float64(len(rai.Stats.PowerHistory))
	}

	return fmt.Sprintf(
		"Reliability Summary:\n"+
			"  Inferences: %d\n"+
			"  Avg Accuracy: %.2f%%\n"+
			"  Avg Energy: %.2f nJ\n"+
			"  SAF Count: %d\n"+
			"  Failed Devices: %d\n"+
			"  Maintenance Needed: %v",
		rai.Stats.InferenceCount,
		avgAccuracy,
		avgPower,
		rai.VariabilityModel.Stats.SAFCount,
		rai.EnduranceModel.Stats.FailedDevices,
		rai.Stats.MaintenanceNeeded,
	)
}

// ============================================================================
// MAPPING METHOD FOR SAF AND VARIATION TOLERANCE
// ============================================================================

// SAFTolerantMapper implements fault-tolerant weight mapping
type SAFTolerantMapper struct {
	SAFMap        [][]int     // Known SAF locations
	VariationMap  [][]float64 // Known variation factors
	Config        *MappingConfig
}

// MappingConfig configures fault-tolerant mapping
type MappingConfig struct {
	ArrayRows        int
	ArrayCols        int
	RedundantRows    int     // Spare rows for remapping
	RedundantCols    int     // Spare columns
	VariationThresh  float64 // Threshold to mark device as unreliable
}

// NewSAFTolerantMapper creates a fault-tolerant mapper
func NewSAFTolerantMapper(safMap [][]int, variationMap [][]float64, config *MappingConfig) *SAFTolerantMapper {
	return &SAFTolerantMapper{
		SAFMap:       safMap,
		VariationMap: variationMap,
		Config:       config,
	}
}

// MapWeights maps weights avoiding SAF and high-variation cells
func (m *SAFTolerantMapper) MapWeights(weights [][]float64) ([][]float64, [][]int) {
	rows := len(weights)
	if rows == 0 {
		return weights, nil
	}
	cols := len(weights[0])

	// Create mapping: logical -> physical
	logicalToPhysical := make([][]int, rows)
	mappedWeights := make([][]float64, m.Config.ArrayRows)

	for i := range mappedWeights {
		mappedWeights[i] = make([]float64, m.Config.ArrayCols)
	}

	physicalRow := 0
	for logicalRow := 0; logicalRow < rows; logicalRow++ {
		logicalToPhysical[logicalRow] = make([]int, cols)

		for logicalCol := 0; logicalCol < cols; logicalCol++ {
			// Find suitable physical location
			placed := false
			for pr := physicalRow; pr < m.Config.ArrayRows && !placed; pr++ {
				for pc := 0; pc < m.Config.ArrayCols && !placed; pc++ {
					// Check if cell is usable
					if m.isCellUsable(pr, pc) {
						mappedWeights[pr][pc] = weights[logicalRow][logicalCol]
						logicalToPhysical[logicalRow][logicalCol] = pr*m.Config.ArrayCols + pc
						placed = true
					}
				}
			}

			if !placed {
				// Use redundant cells
				// Simplified: just use next available
				logicalToPhysical[logicalRow][logicalCol] = -1 // Mark as unmapped
			}
		}
	}

	return mappedWeights, logicalToPhysical
}

func (m *SAFTolerantMapper) isCellUsable(row, col int) bool {
	if row >= len(m.SAFMap) || col >= len(m.SAFMap[0]) {
		return false
	}

	// Check for SAF
	if m.SAFMap[row][col] != 0 {
		return false
	}

	// Check for high variation
	if m.VariationMap != nil && row < len(m.VariationMap) && col < len(m.VariationMap[0]) {
		if math.Abs(m.VariationMap[row][col]-1.0) > m.Config.VariationThresh {
			return false
		}
	}

	return true
}

// ComputeSumWeightVariation computes sum weight variation for quality assessment
func (m *SAFTolerantMapper) ComputeSumWeightVariation(weights [][]float64) float64 {
	totalVariation := 0.0
	count := 0

	rows := len(weights)
	if rows == 0 {
		return 0
	}
	cols := len(weights[0])

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i < len(m.VariationMap) && j < len(m.VariationMap[0]) {
				variation := math.Abs(m.VariationMap[i][j] - 1.0)
				totalVariation += variation * math.Abs(weights[i][j])
				count++
			}
		}
	}

	if count == 0 {
		return 0
	}
	return totalVariation / float64(count)
}

// ============================================================================
// NOISE INJECTION TRAINING
// ============================================================================

// NoiseInjectionTrainer implements noise-aware training
type NoiseInjectionTrainer struct {
	Config    *NoiseTrainingConfig
	Stats     *NoiseTrainingStats
}

// NoiseTrainingConfig configures noise injection training
type NoiseTrainingConfig struct {
	NoiseLevelStart  float64 // Starting noise level
	NoiseLevelEnd    float64 // Ending noise level (for curriculum)
	NoiseSchedule    string  // "constant", "linear", "curriculum"
	MaxEpochs        int
	NoiseType        string  // "gaussian", "uniform", "quantization"
}

// NoiseTrainingStats tracks training statistics
type NoiseTrainingStats struct {
	CurrentEpoch      int
	CurrentNoiseLevel float64
	AccuracyWithNoise float64
	RobustnessGain    float64
}

// NewNoiseInjectionTrainer creates a noise injection trainer
func NewNoiseInjectionTrainer(config *NoiseTrainingConfig) *NoiseInjectionTrainer {
	return &NoiseInjectionTrainer{
		Config: config,
		Stats:  &NoiseTrainingStats{},
	}
}

// GetNoiseLevel returns current noise level based on schedule
func (nit *NoiseInjectionTrainer) GetNoiseLevel(epoch int) float64 {
	nit.Stats.CurrentEpoch = epoch

	switch nit.Config.NoiseSchedule {
	case "constant":
		nit.Stats.CurrentNoiseLevel = nit.Config.NoiseLevelStart
	case "linear":
		progress := float64(epoch) / float64(nit.Config.MaxEpochs)
		nit.Stats.CurrentNoiseLevel = nit.Config.NoiseLevelStart +
			progress*(nit.Config.NoiseLevelEnd-nit.Config.NoiseLevelStart)
	case "curriculum":
		// Start with low noise, gradually increase
		progress := float64(epoch) / float64(nit.Config.MaxEpochs)
		nit.Stats.CurrentNoiseLevel = nit.Config.NoiseLevelStart *
			math.Pow(nit.Config.NoiseLevelEnd/nit.Config.NoiseLevelStart, progress)
	default:
		nit.Stats.CurrentNoiseLevel = nit.Config.NoiseLevelStart
	}

	return nit.Stats.CurrentNoiseLevel
}

// InjectNoise injects noise into weights during training
func (nit *NoiseInjectionTrainer) InjectNoise(weights [][]float64) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])

	noisy := make([][]float64, rows)
	noiseLevel := nit.Stats.CurrentNoiseLevel

	for i := 0; i < rows; i++ {
		noisy[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			switch nit.Config.NoiseType {
			case "gaussian":
				noisy[i][j] = weights[i][j] * (1.0 + noiseLevel*randGaussian())
			case "uniform":
				noisy[i][j] = weights[i][j] * (1.0 + noiseLevel*(2*randFloat()-1))
			case "quantization":
				// Simulate quantization noise
				levels := 64.0 // 6-bit
				noisy[i][j] = math.Round(weights[i][j]*levels) / levels
			default:
				noisy[i][j] = weights[i][j]
			}
		}
	}

	return noisy
}

// ============================================================================
// HELPER: Sorting for optimization
// ============================================================================

type byVariation struct {
	indices   []int
	variation []float64
}

func (b byVariation) Len() int           { return len(b.indices) }
func (b byVariation) Swap(i, j int)      { b.indices[i], b.indices[j] = b.indices[j], b.indices[i] }
func (b byVariation) Less(i, j int) bool { return b.variation[b.indices[i]] < b.variation[b.indices[j]] }

// SortByVariation sorts cell indices by variation level
func SortByVariation(variation []float64) []int {
	indices := make([]int, len(variation))
	for i := range indices {
		indices[i] = i
	}
	sort.Sort(byVariation{indices: indices, variation: variation})
	return indices
}
