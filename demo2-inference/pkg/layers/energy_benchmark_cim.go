// Package layers provides energy harvesting integration and benchmarking frameworks
// for Compute-in-Memory systems. This module implements self-powered CIM systems,
// energy harvester interfaces, and standardized benchmarking metrics based on
// MLPerf, AnalogNAS-Bench, and industry evaluation frameworks.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

// =============================================================================
// ENERGY HARVESTING FOR CIM
// =============================================================================

// EnergyHarvesterType represents different energy harvesting technologies
type EnergyHarvesterType string

const (
	HarvesterSolar        EnergyHarvesterType = "solar"
	HarvesterPiezoelectric EnergyHarvesterType = "piezoelectric"
	HarvesterTriboelectric EnergyHarvesterType = "triboelectric"
	HarvesterThermoelectric EnergyHarvesterType = "thermoelectric"
	HarvesterRF            EnergyHarvesterType = "rf"
	HarvesterHybridPTENG   EnergyHarvesterType = "hybrid_pteng" // Piezo-Tribo hybrid
	HarvesterVibration     EnergyHarvesterType = "vibration"
)

// EnergyHarvesterConfig configures an energy harvester
type EnergyHarvesterConfig struct {
	Type               EnergyHarvesterType
	AreaCm2            float64 // Harvester area in cm²
	EfficiencyPercent  float64 // Conversion efficiency
	MaxPowerMW         float64 // Maximum power output in mW
	MinOperatingV      float64 // Minimum operating voltage
	MaxOperatingV      float64 // Maximum operating voltage
	OutputImpedanceOhm float64 // Output impedance
	EnvironmentFactor  float64 // 0.0-1.0 based on environment
}

// NewEnergyHarvesterConfig creates config for specific harvester type
func NewEnergyHarvesterConfig(harvesterType EnergyHarvesterType) *EnergyHarvesterConfig {
	configs := map[EnergyHarvesterType]*EnergyHarvesterConfig{
		HarvesterSolar: {
			Type:               HarvesterSolar,
			AreaCm2:            1.0,
			EfficiencyPercent:  25.0,   // Wide-bandgap solar cell
			MaxPowerMW:         10.0,   // Indoor conditions
			MinOperatingV:      0.7,
			MaxOperatingV:      1.8,
			OutputImpedanceOhm: 100.0,
			EnvironmentFactor:  0.5,    // Indoor lighting
		},
		HarvesterPiezoelectric: {
			Type:               HarvesterPiezoelectric,
			AreaCm2:            4.0,
			EfficiencyPercent:  15.0,
			MaxPowerMW:         5.0,
			MinOperatingV:      1.0,
			MaxOperatingV:      12.0,
			OutputImpedanceOhm: 10000.0, // High impedance
			EnvironmentFactor:  0.3,     // Intermittent motion
		},
		HarvesterTriboelectric: {
			Type:               HarvesterTriboelectric,
			AreaCm2:            10.0,
			EfficiencyPercent:  50.0,   // TENG efficiency
			MaxPowerMW:         2.0,
			MinOperatingV:      10.0,   // High voltage, low current
			MaxOperatingV:      600.0,
			OutputImpedanceOhm: 100000.0,
			EnvironmentFactor:  0.2,
		},
		HarvesterHybridPTENG: {
			Type:               HarvesterHybridPTENG,
			AreaCm2:            5.0,
			EfficiencyPercent:  35.0,
			MaxPowerMW:         8.0,    // 2.33 mW/cm² peak
			MinOperatingV:      1.0,
			MaxOperatingV:      100.0,
			OutputImpedanceOhm: 50000.0,
			EnvironmentFactor:  0.4,
		},
		HarvesterThermoelectric: {
			Type:               HarvesterThermoelectric,
			AreaCm2:            2.0,
			EfficiencyPercent:  5.0,    // TEG efficiency
			MaxPowerMW:         1.0,
			MinOperatingV:      0.3,
			MaxOperatingV:      3.0,
			OutputImpedanceOhm: 50.0,
			EnvironmentFactor:  0.6,    // Body heat
		},
		HarvesterRF: {
			Type:               HarvesterRF,
			AreaCm2:            5.0,    // Antenna area
			EfficiencyPercent:  40.0,
			MaxPowerMW:         0.1,    // RF harvesting is low power
			MinOperatingV:      0.5,
			MaxOperatingV:      3.3,
			OutputImpedanceOhm: 50.0,   // Matched impedance
			EnvironmentFactor:  0.1,    // Sparse RF environment
		},
	}
	if config, ok := configs[harvesterType]; ok {
		return config
	}
	return configs[HarvesterSolar]
}

// EnergyHarvester simulates an energy harvesting device
type EnergyHarvester struct {
	Config           *EnergyHarvesterConfig
	CurrentPowerMW   float64
	AccumulatedMWh   float64
	OutputVoltageMv  float64
	OutputCurrentUA  float64
	IsActive         bool
	LastUpdateTime   time.Time
	PowerHistory     []PowerSample
}

// PowerSample records power output over time
type PowerSample struct {
	Timestamp   time.Time
	PowerMW     float64
	VoltageV    float64
	CurrentUA   float64
	Environment float64
}

// NewEnergyHarvester creates a new energy harvester
func NewEnergyHarvester(config *EnergyHarvesterConfig) *EnergyHarvester {
	return &EnergyHarvester{
		Config:          config,
		CurrentPowerMW:  0.0,
		AccumulatedMWh:  0.0,
		OutputVoltageMv: 0.0,
		OutputCurrentUA: 0.0,
		IsActive:        false,
		LastUpdateTime:  time.Now(),
		PowerHistory:    make([]PowerSample, 0),
	}
}

// Update updates harvester state based on environment
func (h *EnergyHarvester) Update(environmentFactor float64) {
	h.Config.EnvironmentFactor = environmentFactor

	// Calculate current power output
	basePower := h.Config.MaxPowerMW * h.Config.EfficiencyPercent / 100.0
	h.CurrentPowerMW = basePower * environmentFactor

	// Add noise/variation
	variation := 1.0 + (rand.Float64()-0.5)*0.2
	h.CurrentPowerMW *= variation

	// Clamp to valid range
	if h.CurrentPowerMW < 0 {
		h.CurrentPowerMW = 0
	}
	if h.CurrentPowerMW > h.Config.MaxPowerMW {
		h.CurrentPowerMW = h.Config.MaxPowerMW
	}

	// Calculate voltage and current
	h.OutputVoltageMv = h.Config.MinOperatingV * 1000.0 * math.Sqrt(h.CurrentPowerMW/h.Config.MaxPowerMW+0.01)
	if h.CurrentPowerMW > 0 && h.OutputVoltageMv > 0 {
		h.OutputCurrentUA = h.CurrentPowerMW * 1e6 / h.OutputVoltageMv
	} else {
		h.OutputCurrentUA = 0
	}

	// Update accumulated energy
	now := time.Now()
	elapsed := now.Sub(h.LastUpdateTime).Hours()
	h.AccumulatedMWh += h.CurrentPowerMW * elapsed
	h.LastUpdateTime = now

	h.IsActive = h.CurrentPowerMW > 0.01 // Minimum threshold

	// Record sample
	h.PowerHistory = append(h.PowerHistory, PowerSample{
		Timestamp:   now,
		PowerMW:     h.CurrentPowerMW,
		VoltageV:    h.OutputVoltageMv / 1000.0,
		CurrentUA:   h.OutputCurrentUA,
		Environment: environmentFactor,
	})

	// Keep history bounded
	if len(h.PowerHistory) > 1000 {
		h.PowerHistory = h.PowerHistory[len(h.PowerHistory)-1000:]
	}
}

// GetAveragePower returns average power over history
func (h *EnergyHarvester) GetAveragePower() float64 {
	if len(h.PowerHistory) == 0 {
		return 0
	}
	total := 0.0
	for _, sample := range h.PowerHistory {
		total += sample.PowerMW
	}
	return total / float64(len(h.PowerHistory))
}

// PowerManagementUnit manages power from harvester to CIM
type PowerManagementUnit struct {
	Harvester          *EnergyHarvester
	StorageCapacityMWh float64
	StoredEnergyMWh    float64
	OutputVoltageMv    float64
	RegulationEfficiency float64
	MinOperatingMWh    float64 // Minimum energy to operate
	LoadPowerMW        float64
	ChargingRate       float64 // mW
	DischargingRate    float64 // mW
	IsCharging         bool
	CanOperate         bool
}

// NewPowerManagementUnit creates a PMU
func NewPowerManagementUnit(harvester *EnergyHarvester) *PowerManagementUnit {
	return &PowerManagementUnit{
		Harvester:          harvester,
		StorageCapacityMWh: 0.1,  // 100 μWh storage (supercapacitor)
		StoredEnergyMWh:    0.0,
		OutputVoltageMv:    1000.0, // 1V regulated output
		RegulationEfficiency: 0.85,
		MinOperatingMWh:    0.01,  // 10 μWh minimum
		LoadPowerMW:        0.5,   // 500 μW CIM load
		ChargingRate:       0.0,
		DischargingRate:    0.0,
		IsCharging:         false,
		CanOperate:         false,
	}
}

// Update updates PMU state
func (p *PowerManagementUnit) Update(dtHours float64) {
	// Get harvester power
	harvesterPower := p.Harvester.CurrentPowerMW * p.RegulationEfficiency

	// Net power = harvested - load
	netPower := harvesterPower - p.LoadPowerMW

	if netPower > 0 {
		// Charging
		p.IsCharging = true
		p.ChargingRate = netPower
		p.DischargingRate = 0
		p.StoredEnergyMWh += netPower * dtHours
	} else {
		// Discharging
		p.IsCharging = false
		p.ChargingRate = 0
		p.DischargingRate = -netPower
		p.StoredEnergyMWh += netPower * dtHours // netPower is negative
	}

	// Clamp storage
	if p.StoredEnergyMWh < 0 {
		p.StoredEnergyMWh = 0
	}
	if p.StoredEnergyMWh > p.StorageCapacityMWh {
		p.StoredEnergyMWh = p.StorageCapacityMWh
	}

	// Determine if CIM can operate
	p.CanOperate = p.StoredEnergyMWh >= p.MinOperatingMWh || harvesterPower >= p.LoadPowerMW
}

// GetStateOfCharge returns storage SOC percentage
func (p *PowerManagementUnit) GetStateOfCharge() float64 {
	if p.StorageCapacityMWh == 0 {
		return 0
	}
	return (p.StoredEnergyMWh / p.StorageCapacityMWh) * 100.0
}

// SelfPoweredCIMSystem represents a self-powered CIM system
type SelfPoweredCIMSystem struct {
	Config           *SelfPoweredConfig
	Harvester        *EnergyHarvester
	PMU              *PowerManagementUnit
	CIMArray         *BatterylessCIMArray
	InferenceEngine  *BatterylessInference
	SystemState      SystemPowerState
	InferencesCompleted int
	FailedInferences    int
	TotalEnergyUsedMWh  float64
}

// SelfPoweredConfig configures self-powered CIM system
type SelfPoweredConfig struct {
	HarvesterType      EnergyHarvesterType
	ArrayRows          int
	ArrayCols          int
	BitPrecision       int
	UseNearMemoryCompute bool // vs analog in-memory
	BinaryWeights      bool   // BNN for robustness
	CheckpointEnabled  bool   // Save state on power loss
}

// BatterylessCIMArray represents CIM array for batteryless operation
type BatterylessCIMArray struct {
	Rows              int
	Cols              int
	MinOperatingV     float64
	MaxOperatingV     float64
	CellType          string // "memristor", "fefet", "sram"
	IsNonVolatile     bool
	WeightsStored     bool
	EnergyPerInfPJ    float64
	CurrentUtilization float64
}

// BatterylessInference manages inference in batteryless context
type BatterylessInference struct {
	ModelLoaded       bool
	CheckpointValid   bool
	LastCheckpointIdx int
	InferenceInProgress bool
	CurrentBatchIdx   int
	TotalBatches      int
	ResultsBuffer     []float64
}

// SystemPowerState represents system power state
type SystemPowerState string

const (
	PowerStateOff       SystemPowerState = "off"
	PowerStateStartup   SystemPowerState = "startup"
	PowerStateActive    SystemPowerState = "active"
	PowerStateLowPower  SystemPowerState = "low_power"
	PowerStateCheckpoint SystemPowerState = "checkpoint"
)

// NewSelfPoweredCIMSystem creates a self-powered CIM system
func NewSelfPoweredCIMSystem(config *SelfPoweredConfig) *SelfPoweredCIMSystem {
	harvesterConfig := NewEnergyHarvesterConfig(config.HarvesterType)
	harvester := NewEnergyHarvester(harvesterConfig)
	pmu := NewPowerManagementUnit(harvester)

	// Adjust PMU for system needs
	energyPerInf := float64(config.ArrayRows*config.ArrayCols) * 0.5 // 0.5 pJ per MAC
	pmu.LoadPowerMW = energyPerInf * 1e-9 * 1000 // Convert to mW assuming 1M inf/s

	return &SelfPoweredCIMSystem{
		Config:    config,
		Harvester: harvester,
		PMU:       pmu,
		CIMArray: &BatterylessCIMArray{
			Rows:              config.ArrayRows,
			Cols:              config.ArrayCols,
			MinOperatingV:     0.7,
			MaxOperatingV:     1.8,
			CellType:          "memristor",
			IsNonVolatile:     true,
			WeightsStored:     true,
			EnergyPerInfPJ:    energyPerInf,
			CurrentUtilization: 0.0,
		},
		InferenceEngine: &BatterylessInference{
			ModelLoaded:       true,
			CheckpointValid:   false,
			LastCheckpointIdx: 0,
			InferenceInProgress: false,
			CurrentBatchIdx:   0,
			TotalBatches:      0,
			ResultsBuffer:     make([]float64, 0),
		},
		SystemState:        PowerStateOff,
		InferencesCompleted: 0,
		FailedInferences:   0,
		TotalEnergyUsedMWh: 0.0,
	}
}

// RunInference attempts to run inference with available power
func (s *SelfPoweredCIMSystem) RunInference(inputData []float64) (*SelfPoweredInferenceResult, error) {
	result := &SelfPoweredInferenceResult{
		Success:       false,
		PoweredBy:     string(s.Config.HarvesterType),
		EnergyUsedPJ:  0,
		LatencyUs:     0,
		CheckpointUsed: false,
	}

	// Check if we can operate
	if !s.PMU.CanOperate {
		s.FailedInferences++
		return result, fmt.Errorf("insufficient power for inference")
	}

	// Update state
	s.SystemState = PowerStateActive
	s.InferenceEngine.InferenceInProgress = true

	// Calculate energy needed
	energyNeededPJ := s.CIMArray.EnergyPerInfPJ * float64(len(inputData))
	energyNeededMWh := energyNeededPJ * 1e-9 / 3600.0

	// Check if we have enough energy
	if s.PMU.StoredEnergyMWh < energyNeededMWh && s.Harvester.CurrentPowerMW*0.001 < energyNeededMWh {
		// Try checkpoint-based partial inference
		if s.Config.CheckpointEnabled {
			result.CheckpointUsed = true
			// Partial inference logic would go here
		} else {
			s.FailedInferences++
			s.SystemState = PowerStateLowPower
			return result, fmt.Errorf("insufficient energy for full inference")
		}
	}

	// Perform inference (simplified)
	output := make([]float64, s.CIMArray.Cols)
	for i := range output {
		sum := 0.0
		for j := 0; j < len(inputData) && j < s.CIMArray.Rows; j++ {
			// Simulated MAC with noise for analog non-ideality
			weight := rand.Float64()*2 - 1 // Random weights for simulation
			noise := rand.NormFloat64() * 0.05
			sum += inputData[j] * weight * (1 + noise)
		}
		output[i] = sum
	}

	// Calculate latency
	latencyUs := float64(s.CIMArray.Rows*s.CIMArray.Cols) / 1000.0 // 1 ns per MAC

	// Update energy accounting
	s.TotalEnergyUsedMWh += energyNeededMWh
	s.PMU.StoredEnergyMWh -= energyNeededMWh
	if s.PMU.StoredEnergyMWh < 0 {
		s.PMU.StoredEnergyMWh = 0
	}

	// Success
	result.Success = true
	result.Output = output
	result.EnergyUsedPJ = energyNeededPJ
	result.LatencyUs = latencyUs
	s.InferencesCompleted++
	s.InferenceEngine.InferenceInProgress = false

	return result, nil
}

// SelfPoweredInferenceResult holds inference results
type SelfPoweredInferenceResult struct {
	Success        bool
	Output         []float64
	PoweredBy      string
	EnergyUsedPJ   float64
	LatencyUs      float64
	CheckpointUsed bool
}

// GetSystemMetrics returns system performance metrics
func (s *SelfPoweredCIMSystem) GetSystemMetrics() *SelfPoweredMetrics {
	successRate := 0.0
	totalAttempts := s.InferencesCompleted + s.FailedInferences
	if totalAttempts > 0 {
		successRate = float64(s.InferencesCompleted) / float64(totalAttempts) * 100.0
	}

	return &SelfPoweredMetrics{
		InferencesCompleted: s.InferencesCompleted,
		FailedInferences:    s.FailedInferences,
		SuccessRate:         successRate,
		TotalEnergyUsedMWh:  s.TotalEnergyUsedMWh,
		AverageHarvestPowerMW: s.Harvester.GetAveragePower(),
		StorageSOC:          s.PMU.GetStateOfCharge(),
		SystemState:         string(s.SystemState),
	}
}

// SelfPoweredMetrics holds system metrics
type SelfPoweredMetrics struct {
	InferencesCompleted   int
	FailedInferences      int
	SuccessRate           float64
	TotalEnergyUsedMWh    float64
	AverageHarvestPowerMW float64
	StorageSOC            float64
	SystemState           string
}

// =============================================================================
// HYBRID PIEZO-TRIBO NANOGENERATOR (PT-HNG)
// =============================================================================

// PTHybridNanogenerator represents a piezoelectric-triboelectric hybrid
type PTHybridNanogenerator struct {
	Config              *PTHNGConfig
	PiezoComponent      *PiezoelectricLayer
	TriboComponent      *TriboelectricLayer
	OutputPowerMW       float64
	OutputVoltageV      float64
	OutputCurrentUA     float64
	CombinedEfficiency  float64
}

// PTHNGConfig configures the hybrid nanogenerator
type PTHNGConfig struct {
	PiezoMaterial       string  // "PZT", "PVDF", "BaTiO3"
	TriboMaterialPos    string  // Positive triboelectric material
	TriboMaterialNeg    string  // Negative triboelectric material
	AreaCm2             float64
	ThicknessUm         float64
	FerroelectricDoped  bool    // Enhanced with ferroelectric nanoparticles
	AgBCZTNanowires     bool    // Ag-BCZT nanowire enhancement
}

// PiezoelectricLayer represents piezo component
type PiezoelectricLayer struct {
	Material            string
	D33PmPerV           float64 // Piezoelectric coefficient
	PolarizationUCcm2   float64
	Capacitance         float64
	OpenCircuitVoltage  float64
	ShortCircuitCurrent float64
}

// TriboelectricLayer represents TENG component
type TriboelectricLayer struct {
	PositiveMaterial    string
	NegativeMaterial    string
	SurfaceChargeDensity float64 // μC/m²
	ContactArea         float64
	SeparationGap       float64
	Mode                string  // "contact-separation", "sliding", "single-electrode"
}

// NewPTHybridNanogenerator creates a hybrid nanogenerator
func NewPTHybridNanogenerator(config *PTHNGConfig) *PTHybridNanogenerator {
	// Piezoelectric coefficient based on material
	d33 := 100.0 // Default
	switch config.PiezoMaterial {
	case "PZT":
		d33 = 585.0 // High performance PZT
	case "PVDF":
		d33 = 33.0
	case "BaTiO3":
		d33 = 190.0
	}

	// Enhancement from Ag-BCZT nanowires
	if config.AgBCZTNanowires {
		d33 *= 1.5 // 50% enhancement
	}

	return &PTHybridNanogenerator{
		Config: config,
		PiezoComponent: &PiezoelectricLayer{
			Material:            config.PiezoMaterial,
			D33PmPerV:           d33,
			PolarizationUCcm2:   25.0,
			Capacitance:         0.0,
			OpenCircuitVoltage:  0.0,
			ShortCircuitCurrent: 0.0,
		},
		TriboComponent: &TriboelectricLayer{
			PositiveMaterial:    config.TriboMaterialPos,
			NegativeMaterial:    config.TriboMaterialNeg,
			SurfaceChargeDensity: 50.0, // μC/m²
			ContactArea:         config.AreaCm2,
			SeparationGap:       0.5,   // mm
			Mode:                "contact-separation",
		},
		OutputPowerMW:      0.0,
		OutputVoltageV:     0.0,
		OutputCurrentUA:    0.0,
		CombinedEfficiency: 0.35, // 35% for hybrid
	}
}

// Harvest harvests energy from mechanical input
func (p *PTHybridNanogenerator) Harvest(forceN, frequencyHz float64) *HarvestResult {
	// Piezoelectric contribution
	// V_piezo = d33 * force * thickness / (permittivity * area)
	piezoVoltage := p.PiezoComponent.D33PmPerV * forceN * 0.001 // Simplified
	piezoCurrent := piezoVoltage / 10000.0 * 1e6 // High impedance, μA

	// Triboelectric contribution
	// V_TENG = σ * d / ε₀
	triboVoltage := p.TriboComponent.SurfaceChargeDensity * p.TriboComponent.SeparationGap / 8.85e-6
	triboCurrent := triboVoltage / 100000.0 * 1e6 // Very high impedance

	// Combine outputs (they're out of phase, so use RMS combination)
	p.OutputVoltageV = math.Sqrt(piezoVoltage*piezoVoltage + triboVoltage*triboVoltage)
	p.OutputCurrentUA = math.Sqrt(piezoCurrent*piezoCurrent + triboCurrent*triboCurrent)

	// Power calculation
	piezoPower := piezoVoltage * piezoCurrent * 1e-6 // mW
	triboPower := triboVoltage * triboCurrent * 1e-6 // mW

	// Synergistic enhancement from hybrid (research shows 1.3-2x improvement)
	synergyFactor := 1.5
	if p.Config.FerroelectricDoped {
		synergyFactor = 2.0
	}

	p.OutputPowerMW = (piezoPower + triboPower) * synergyFactor * frequencyHz

	// Peak power density: 2.33 mW/cm² achieved with Ag-BCZT
	maxPowerDensity := 2.33 // mW/cm²
	if p.Config.AgBCZTNanowires {
		maxPowerDensity = 2.33
	} else {
		maxPowerDensity = 1.0
	}
	maxPower := maxPowerDensity * p.Config.AreaCm2
	if p.OutputPowerMW > maxPower {
		p.OutputPowerMW = maxPower
	}

	return &HarvestResult{
		PiezoPowerMW:     piezoPower * frequencyHz,
		TriboPowerMW:     triboPower * frequencyHz,
		TotalPowerMW:     p.OutputPowerMW,
		VoltageV:         p.OutputVoltageV,
		CurrentUA:        p.OutputCurrentUA,
		PowerDensityMWcm2: p.OutputPowerMW / p.Config.AreaCm2,
		Efficiency:       p.CombinedEfficiency,
	}
}

// HarvestResult holds harvesting results
type HarvestResult struct {
	PiezoPowerMW      float64
	TriboPowerMW      float64
	TotalPowerMW      float64
	VoltageV          float64
	CurrentUA         float64
	PowerDensityMWcm2 float64
	Efficiency        float64
}

// =============================================================================
// CIM BENCHMARKING FRAMEWORK
// =============================================================================

// CIMBenchmarkSuite represents a comprehensive benchmark suite
type CIMBenchmarkSuite struct {
	Config          *BenchmarkConfig
	Benchmarks      []*CIMBenchmark
	Results         []*BenchmarkResult
	Comparisons     []*TechnologyComparison
}

// BenchmarkConfig configures the benchmark suite
type BenchmarkConfig struct {
	Name              string
	Version           string
	IncludeMLPerf     bool
	IncludeAnalogNAS  bool
	IncludeCustom     bool
	EvaluateAccuracy  bool
	EvaluateEnergy    bool
	EvaluateLatency   bool
	EvaluateDrift     bool
	DriftIntervals    []time.Duration
}

// CIMBenchmark represents a single benchmark
type CIMBenchmark struct {
	Name              string
	Category          string // "inference", "training", "edge", "datacenter"
	Model             string
	Dataset           string
	BatchSize         int
	TargetAccuracy    float64
	TargetLatencyMs   float64
	TargetEnergyMJ    float64
	InputShape        []int
	Workload          *BenchmarkWorkload
}

// BenchmarkWorkload represents the computational workload
type BenchmarkWorkload struct {
	TotalMACs         int64
	TotalParams       int64
	MemoryAccessBytes int64
	LayerTypes        []string
	ComputeIntensity  float64 // MACs per byte
}

// BenchmarkResult holds benchmark execution results
type BenchmarkResult struct {
	Benchmark         *CIMBenchmark
	Technology        string  // "sram_cim", "rram_cim", "fefet_cim", "digital"
	BaselineAccuracy  float64 // Full precision accuracy
	NoisyAccuracy     float64 // Accuracy with hardware noise
	AnalogAccuracy    float64 // Hardware-aware trained accuracy
	DriftAccuracies   map[string]float64 // Accuracy over time
	LatencyMs         float64
	EnergyMJ          float64
	TOPSW             float64 // Energy efficiency
	TOPSmm2           float64 // Area efficiency
	MacroUtilization  float64
	ADCOverhead       float64 // ADC energy/area fraction
	CSNR              float64 // Compute signal-to-noise ratio
}

// TechnologyComparison compares different CIM technologies
type TechnologyComparison struct {
	Technologies     []string
	Metric           string
	Values           map[string]float64
	Winner           string
	WinnerMargin     float64
}

// NewCIMBenchmarkSuite creates a benchmark suite
func NewCIMBenchmarkSuite(config *BenchmarkConfig) *CIMBenchmarkSuite {
	suite := &CIMBenchmarkSuite{
		Config:      config,
		Benchmarks:  make([]*CIMBenchmark, 0),
		Results:     make([]*BenchmarkResult, 0),
		Comparisons: make([]*TechnologyComparison, 0),
	}

	// Add MLPerf-style benchmarks
	if config.IncludeMLPerf {
		suite.addMLPerfBenchmarks()
	}

	// Add AnalogNAS-Bench style benchmarks
	if config.IncludeAnalogNAS {
		suite.addAnalogNASBenchmarks()
	}

	// Add custom CIM benchmarks
	if config.IncludeCustom {
		suite.addCustomBenchmarks()
	}

	return suite
}

func (s *CIMBenchmarkSuite) addMLPerfBenchmarks() {
	// MLPerf Mobile benchmarks adapted for CIM
	mlperfBenchmarks := []*CIMBenchmark{
		{
			Name:            "MLPerf-Mobile-ImageClassification",
			Category:        "edge",
			Model:           "MobileNetV2",
			Dataset:         "ImageNet",
			BatchSize:       1,
			TargetAccuracy:  0.75,
			TargetLatencyMs: 10.0,
			InputShape:      []int{1, 224, 224, 3},
			Workload: &BenchmarkWorkload{
				TotalMACs:        300000000,
				TotalParams:      3500000,
				MemoryAccessBytes: 14000000,
				LayerTypes:       []string{"conv2d", "depthwise_conv", "dense"},
				ComputeIntensity: 21.4,
			},
		},
		{
			Name:            "MLPerf-Mobile-ObjectDetection",
			Category:        "edge",
			Model:           "SSD-MobileNetV2",
			Dataset:         "COCO",
			BatchSize:       1,
			TargetAccuracy:  0.22, // mAP
			TargetLatencyMs: 30.0,
			InputShape:      []int{1, 320, 320, 3},
			Workload: &BenchmarkWorkload{
				TotalMACs:        800000000,
				TotalParams:      5500000,
				MemoryAccessBytes: 22000000,
				LayerTypes:       []string{"conv2d", "depthwise_conv", "dense", "nms"},
				ComputeIntensity: 36.4,
			},
		},
		{
			Name:            "MLPerf-Mobile-ImageSegmentation",
			Category:        "edge",
			Model:           "DeepLabV3",
			Dataset:         "ADE20K",
			BatchSize:       1,
			TargetAccuracy:  0.31, // mIoU
			TargetLatencyMs: 50.0,
			InputShape:      []int{1, 512, 512, 3},
			Workload: &BenchmarkWorkload{
				TotalMACs:        2700000000,
				TotalParams:      26000000,
				MemoryAccessBytes: 104000000,
				LayerTypes:       []string{"conv2d", "atrous_conv", "dense"},
				ComputeIntensity: 26.0,
			},
		},
		{
			Name:            "MLPerf-LLM-Summarization",
			Category:        "datacenter",
			Model:           "Llama3.1-8B",
			Dataset:         "CNN-DailyMail",
			BatchSize:       1,
			TargetAccuracy:  0.30, // ROUGE-L
			TargetLatencyMs: 1000.0,
			InputShape:      []int{1, 2048}, // Token sequence
			Workload: &BenchmarkWorkload{
				TotalMACs:        16000000000000, // 16T MACs
				TotalParams:      8000000000,
				MemoryAccessBytes: 32000000000,
				LayerTypes:       []string{"attention", "ffn", "embedding"},
				ComputeIntensity: 500.0,
			},
		},
	}

	s.Benchmarks = append(s.Benchmarks, mlperfBenchmarks...)
}

func (s *CIMBenchmarkSuite) addAnalogNASBenchmarks() {
	// AnalogNAS-Bench style benchmarks for AIMC evaluation
	analogBenchmarks := []*CIMBenchmark{
		{
			Name:            "AnalogNAS-CIFAR10-VGG8",
			Category:        "inference",
			Model:           "VGG8",
			Dataset:         "CIFAR-10",
			BatchSize:       64,
			TargetAccuracy:  0.8955, // Validated accuracy
			TargetLatencyMs: 5.0,
			InputShape:      []int{64, 32, 32, 3},
			Workload: &BenchmarkWorkload{
				TotalMACs:        150000000,
				TotalParams:      7500000,
				MemoryAccessBytes: 30000000,
				LayerTypes:       []string{"conv2d", "dense", "pool"},
				ComputeIntensity: 5.0,
			},
		},
		{
			Name:            "AnalogNAS-ImageNet-ResNet18",
			Category:        "inference",
			Model:           "ResNet18",
			Dataset:         "ImageNet",
			BatchSize:       32,
			TargetAccuracy:  0.70,
			TargetLatencyMs: 20.0,
			InputShape:      []int{32, 224, 224, 3},
			Workload: &BenchmarkWorkload{
				TotalMACs:        1800000000,
				TotalParams:      11700000,
				MemoryAccessBytes: 46800000,
				LayerTypes:       []string{"conv2d", "dense", "batchnorm", "residual"},
				ComputeIntensity: 38.5,
			},
		},
	}

	s.Benchmarks = append(s.Benchmarks, analogBenchmarks...)
}

func (s *CIMBenchmarkSuite) addCustomBenchmarks() {
	// Custom CIM-specific benchmarks
	customBenchmarks := []*CIMBenchmark{
		{
			Name:            "CIM-BNN-MNIST",
			Category:        "edge",
			Model:           "BinaryNet",
			Dataset:         "MNIST",
			BatchSize:       1,
			TargetAccuracy:  0.98,
			TargetLatencyMs: 0.1,
			InputShape:      []int{1, 28, 28, 1},
			Workload: &BenchmarkWorkload{
				TotalMACs:        500000,
				TotalParams:      100000,
				MemoryAccessBytes: 400000,
				LayerTypes:       []string{"binary_conv", "binary_dense"},
				ComputeIntensity: 1.25,
			},
		},
		{
			Name:            "CIM-Attention-GPT2",
			Category:        "datacenter",
			Model:           "GPT-2-Small",
			Dataset:         "WikiText",
			BatchSize:       1,
			TargetAccuracy:  0.0, // Perplexity metric
			TargetLatencyMs: 100.0,
			InputShape:      []int{1, 1024}, // Tokens
			Workload: &BenchmarkWorkload{
				TotalMACs:        124000000000,
				TotalParams:      124000000,
				MemoryAccessBytes: 496000000,
				LayerTypes:       []string{"attention", "ffn"},
				ComputeIntensity: 250.0,
			},
		},
		{
			Name:            "CIM-SNN-DVS",
			Category:        "edge",
			Model:           "SpikingResNet",
			Dataset:         "DVS-Gesture",
			BatchSize:       1,
			TargetAccuracy:  0.95,
			TargetLatencyMs: 10.0,
			InputShape:      []int{1, 128, 128, 2},
			Workload: &BenchmarkWorkload{
				TotalMACs:        10000000, // Sparse ops
				TotalParams:      500000,
				MemoryAccessBytes: 2000000,
				LayerTypes:       []string{"spiking_conv", "lif_neuron"},
				ComputeIntensity: 5.0,
			},
		},
	}

	s.Benchmarks = append(s.Benchmarks, customBenchmarks...)
}

// RunBenchmark runs a benchmark on specified technology
func (s *CIMBenchmarkSuite) RunBenchmark(benchmark *CIMBenchmark, technology string) *BenchmarkResult {
	result := &BenchmarkResult{
		Benchmark:        benchmark,
		Technology:       technology,
		DriftAccuracies:  make(map[string]float64),
	}

	// Simulate benchmark execution based on technology
	switch technology {
	case "sram_cim":
		result = s.runSRAMCIMBenchmark(benchmark, result)
	case "rram_cim":
		result = s.runRRAMCIMBenchmark(benchmark, result)
	case "fefet_cim":
		result = s.runFeFETCIMBenchmark(benchmark, result)
	case "digital":
		result = s.runDigitalBenchmark(benchmark, result)
	default:
		result = s.runGenericCIMBenchmark(benchmark, result)
	}

	s.Results = append(s.Results, result)
	return result
}

func (s *CIMBenchmarkSuite) runSRAMCIMBenchmark(benchmark *CIMBenchmark, result *BenchmarkResult) *BenchmarkResult {
	// SRAM CIM characteristics
	// High accuracy, high speed, limited density, no drift

	result.BaselineAccuracy = benchmark.TargetAccuracy
	result.NoisyAccuracy = benchmark.TargetAccuracy * 0.99 // Minimal noise
	result.AnalogAccuracy = benchmark.TargetAccuracy * 0.995

	// SRAM CIM energy efficiency: ~1000-5000 TOPS/W for digital SRAM CIM
	result.TOPSW = 2500.0
	result.TOPSmm2 = 3854.0 // DREAM-CIM result

	// Latency based on array operations
	result.LatencyMs = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e9) * 1000

	// Energy calculation
	result.EnergyMJ = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e12) * 1000

	// No drift for SRAM
	for _, interval := range s.Config.DriftIntervals {
		result.DriftAccuracies[interval.String()] = result.AnalogAccuracy
	}

	result.MacroUtilization = 0.85
	result.ADCOverhead = 0.15 // 15% for digital SRAM CIM
	result.CSNR = 40.0 // High SNR

	return result
}

func (s *CIMBenchmarkSuite) runRRAMCIMBenchmark(benchmark *CIMBenchmark, result *BenchmarkResult) *BenchmarkResult {
	// RRAM CIM characteristics
	// Higher density, analog compute, drift over time, ADC bottleneck

	result.BaselineAccuracy = benchmark.TargetAccuracy
	result.NoisyAccuracy = benchmark.TargetAccuracy * 0.92 // Significant noise
	result.AnalogAccuracy = benchmark.TargetAccuracy * 0.97 // With HAT

	// RRAM CIM: 10-200 TOPS/W typical
	result.TOPSW = 100.0
	result.TOPSmm2 = 500.0

	result.LatencyMs = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e9) * 1000
	result.EnergyMJ = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e12) * 1000

	// Drift modeling
	driftRate := 0.001 // 0.1% per log(time)
	for _, interval := range s.Config.DriftIntervals {
		timeFactor := math.Log10(float64(interval.Seconds()) + 1)
		drift := 1.0 - driftRate*timeFactor
		result.DriftAccuracies[interval.String()] = result.AnalogAccuracy * drift
	}

	result.MacroUtilization = 0.70
	result.ADCOverhead = 0.50 // 50% ADC overhead for analog RRAM
	result.CSNR = 20.0

	return result
}

func (s *CIMBenchmarkSuite) runFeFETCIMBenchmark(benchmark *CIMBenchmark, result *BenchmarkResult) *BenchmarkResult {
	// FeFET CIM characteristics
	// Multi-level cells, moderate drift, good endurance, CMOS compatible

	result.BaselineAccuracy = benchmark.TargetAccuracy
	result.NoisyAccuracy = benchmark.TargetAccuracy * 0.95
	result.AnalogAccuracy = benchmark.TargetAccuracy * 0.98

	// FeFET CIM: REMNA achieves 26.72 TOPS/W
	result.TOPSW = 26.72
	result.TOPSmm2 = 200.0 // Higher density than SRAM

	result.LatencyMs = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e9) * 1000
	result.EnergyMJ = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e12) * 1000

	// Moderate drift
	driftRate := 0.0005
	for _, interval := range s.Config.DriftIntervals {
		timeFactor := math.Log10(float64(interval.Seconds()) + 1)
		drift := 1.0 - driftRate*timeFactor
		result.DriftAccuracies[interval.String()] = result.AnalogAccuracy * drift
	}

	result.MacroUtilization = 0.80
	result.ADCOverhead = 0.30
	result.CSNR = 25.0

	return result
}

func (s *CIMBenchmarkSuite) runDigitalBenchmark(benchmark *CIMBenchmark, result *BenchmarkResult) *BenchmarkResult {
	// Traditional digital baseline (GPU/TPU)

	result.BaselineAccuracy = benchmark.TargetAccuracy
	result.NoisyAccuracy = benchmark.TargetAccuracy // No noise in digital
	result.AnalogAccuracy = benchmark.TargetAccuracy

	// Digital: GPU ~1-10 TOPS/W
	result.TOPSW = 5.0
	result.TOPSmm2 = 50.0

	result.LatencyMs = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e9) * 1000
	result.EnergyMJ = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e12) * 1000

	// No drift
	for _, interval := range s.Config.DriftIntervals {
		result.DriftAccuracies[interval.String()] = result.AnalogAccuracy
	}

	result.MacroUtilization = 0.60
	result.ADCOverhead = 0.0
	result.CSNR = 100.0 // Effectively infinite for digital

	return result
}

func (s *CIMBenchmarkSuite) runGenericCIMBenchmark(benchmark *CIMBenchmark, result *BenchmarkResult) *BenchmarkResult {
	// Generic CIM benchmark
	result.BaselineAccuracy = benchmark.TargetAccuracy
	result.NoisyAccuracy = benchmark.TargetAccuracy * 0.93
	result.AnalogAccuracy = benchmark.TargetAccuracy * 0.96

	result.TOPSW = 50.0
	result.TOPSmm2 = 200.0

	result.LatencyMs = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e9) * 1000
	result.EnergyMJ = float64(benchmark.Workload.TotalMACs) / (result.TOPSW * 1e12) * 1000

	for _, interval := range s.Config.DriftIntervals {
		result.DriftAccuracies[interval.String()] = result.AnalogAccuracy * 0.99
	}

	result.MacroUtilization = 0.75
	result.ADCOverhead = 0.35
	result.CSNR = 22.0

	return result
}

// CompareTechnologies compares technologies across a metric
func (s *CIMBenchmarkSuite) CompareTechnologies(technologies []string, metric string) *TechnologyComparison {
	comparison := &TechnologyComparison{
		Technologies: technologies,
		Metric:       metric,
		Values:       make(map[string]float64),
	}

	for _, tech := range technologies {
		values := make([]float64, 0)
		for _, result := range s.Results {
			if result.Technology == tech {
				switch metric {
				case "accuracy":
					values = append(values, result.AnalogAccuracy)
				case "energy_efficiency":
					values = append(values, result.TOPSW)
				case "area_efficiency":
					values = append(values, result.TOPSmm2)
				case "latency":
					values = append(values, result.LatencyMs)
				case "csnr":
					values = append(values, result.CSNR)
				}
			}
		}
		if len(values) > 0 {
			avg := 0.0
			for _, v := range values {
				avg += v
			}
			comparison.Values[tech] = avg / float64(len(values))
		}
	}

	// Determine winner (highest is better except for latency)
	bestValue := -math.MaxFloat64
	if metric == "latency" {
		bestValue = math.MaxFloat64
	}

	for tech, value := range comparison.Values {
		if metric == "latency" {
			if value < bestValue {
				bestValue = value
				comparison.Winner = tech
			}
		} else {
			if value > bestValue {
				bestValue = value
				comparison.Winner = tech
			}
		}
	}

	// Calculate margin
	if len(comparison.Values) >= 2 {
		values := make([]float64, 0)
		for _, v := range comparison.Values {
			values = append(values, v)
		}
		sort.Float64s(values)
		if len(values) >= 2 {
			comparison.WinnerMargin = (values[len(values)-1] - values[len(values)-2]) / values[len(values)-2] * 100
		}
	}

	s.Comparisons = append(s.Comparisons, comparison)
	return comparison
}

// =============================================================================
// ACCURACY AND NOISE METRICS
// =============================================================================

// AccuracyMetrics represents comprehensive accuracy evaluation
type AccuracyMetrics struct {
	BaselineAccuracy    float64
	NoisyAccuracy       float64
	AnalogAccuracy      float64
	DriftAccuracies     map[time.Duration]float64
	AccuracyDegradation float64
	RobustnessScore     float64
}

// NoiseModel represents hardware noise characteristics
type NoiseModel struct {
	ConductanceVariation float64 // σ/μ ratio
	ProgramNoise         float64
	ReadNoise            float64
	ADCQuantization      int
	IRDropFactor         float64
	ThermalNoise         float64
}

// CSNRCalculator calculates Compute Signal-to-Noise Ratio
type CSNRCalculator struct {
	NoiseModel      *NoiseModel
	ArraySize       int
	BitPrecision    int
	InputVariance   float64
	WeightVariance  float64
}

// NewCSNRCalculator creates CSNR calculator
func NewCSNRCalculator(noiseModel *NoiseModel, arraySize, bitPrecision int) *CSNRCalculator {
	return &CSNRCalculator{
		NoiseModel:    noiseModel,
		ArraySize:     arraySize,
		BitPrecision:  bitPrecision,
		InputVariance: 1.0,
		WeightVariance: 1.0,
	}
}

// CalculateCSNR calculates the compute signal-to-noise ratio
func (c *CSNRCalculator) CalculateCSNR() float64 {
	// CSNR = Signal Power / Noise Power
	// Signal = N * σ_x² * σ_w² (expected MAC output variance)
	signalPower := float64(c.ArraySize) * c.InputVariance * c.WeightVariance

	// Noise contributions
	conductanceNoise := c.NoiseModel.ConductanceVariation * c.NoiseModel.ConductanceVariation
	programNoise := c.NoiseModel.ProgramNoise * c.NoiseModel.ProgramNoise
	readNoise := c.NoiseModel.ReadNoise * c.NoiseModel.ReadNoise
	quantNoise := 1.0 / (12.0 * math.Pow(2, float64(2*c.NoiseModel.ADCQuantization))) // Quantization noise
	thermalNoise := c.NoiseModel.ThermalNoise * c.NoiseModel.ThermalNoise

	totalNoise := conductanceNoise + programNoise + readNoise + quantNoise + thermalNoise
	noisePower := float64(c.ArraySize) * totalNoise * c.WeightVariance * c.InputVariance

	if noisePower == 0 {
		return 100.0 // Effectively infinite SNR
	}

	csnr := signalPower / noisePower
	return 10 * math.Log10(csnr) // Return in dB
}

// GetOptimalADCBits determines optimal ADC precision for given CSNR target
func (c *CSNRCalculator) GetOptimalADCBits(targetCSNRdB float64) int {
	for bits := 1; bits <= 12; bits++ {
		c.NoiseModel.ADCQuantization = bits
		csnr := c.CalculateCSNR()
		if csnr >= targetCSNRdB {
			return bits
		}
	}
	return 12 // Maximum
}

// =============================================================================
// BENCHMARK REPORT GENERATION
// =============================================================================

// BenchmarkReport generates comprehensive report
type BenchmarkReport struct {
	SuiteName           string
	Timestamp           time.Time
	TotalBenchmarks     int
	TechnologiesTested  []string
	BenchmarkSummaries  []*BenchmarkSummary
	TechnologyRankings  map[string][]string
	KeyFindings         []string
}

// BenchmarkSummary summarizes results for a benchmark
type BenchmarkSummary struct {
	BenchmarkName       string
	BestTechnology      string
	BestAccuracy        float64
	BestEfficiency      float64
	WorstDrift          float64
	Recommendations     []string
}

// GenerateReport creates a comprehensive benchmark report
func (s *CIMBenchmarkSuite) GenerateReport() *BenchmarkReport {
	report := &BenchmarkReport{
		SuiteName:          s.Config.Name,
		Timestamp:          time.Now(),
		TotalBenchmarks:    len(s.Benchmarks),
		TechnologiesTested: make([]string, 0),
		BenchmarkSummaries: make([]*BenchmarkSummary, 0),
		TechnologyRankings: make(map[string][]string),
		KeyFindings:        make([]string, 0),
	}

	// Collect unique technologies
	techMap := make(map[string]bool)
	for _, result := range s.Results {
		techMap[result.Technology] = true
	}
	for tech := range techMap {
		report.TechnologiesTested = append(report.TechnologiesTested, tech)
	}

	// Generate summaries per benchmark
	for _, benchmark := range s.Benchmarks {
		summary := &BenchmarkSummary{
			BenchmarkName:   benchmark.Name,
			Recommendations: make([]string, 0),
		}

		bestAcc := 0.0
		bestEff := 0.0
		worstDrift := 1.0

		for _, result := range s.Results {
			if result.Benchmark == benchmark {
				if result.AnalogAccuracy > bestAcc {
					bestAcc = result.AnalogAccuracy
					summary.BestTechnology = result.Technology
				}
				if result.TOPSW > bestEff {
					bestEff = result.TOPSW
				}
				for _, driftAcc := range result.DriftAccuracies {
					if driftAcc < worstDrift {
						worstDrift = driftAcc
					}
				}
			}
		}

		summary.BestAccuracy = bestAcc
		summary.BestEfficiency = bestEff
		summary.WorstDrift = worstDrift

		// Generate recommendations
		if bestAcc < benchmark.TargetAccuracy {
			summary.Recommendations = append(summary.Recommendations,
				"Consider hardware-aware training to improve accuracy")
		}
		if worstDrift < bestAcc*0.95 {
			summary.Recommendations = append(summary.Recommendations,
				"Implement drift compensation or periodic recalibration")
		}

		report.BenchmarkSummaries = append(report.BenchmarkSummaries, summary)
	}

	// Generate rankings
	metrics := []string{"accuracy", "energy_efficiency", "area_efficiency"}
	for _, metric := range metrics {
		comparison := s.CompareTechnologies(report.TechnologiesTested, metric)

		// Sort technologies by value
		type techValue struct {
			tech  string
			value float64
		}
		tvs := make([]techValue, 0)
		for tech, value := range comparison.Values {
			tvs = append(tvs, techValue{tech, value})
		}
		sort.Slice(tvs, func(i, j int) bool {
			return tvs[i].value > tvs[j].value
		})

		ranking := make([]string, len(tvs))
		for i, tv := range tvs {
			ranking[i] = tv.tech
		}
		report.TechnologyRankings[metric] = ranking
	}

	// Generate key findings
	report.KeyFindings = append(report.KeyFindings,
		fmt.Sprintf("Evaluated %d benchmarks across %d technologies",
			report.TotalBenchmarks, len(report.TechnologiesTested)))

	if len(report.TechnologyRankings["energy_efficiency"]) > 0 {
		report.KeyFindings = append(report.KeyFindings,
			fmt.Sprintf("Most energy-efficient technology: %s",
				report.TechnologyRankings["energy_efficiency"][0]))
	}

	return report
}

// =============================================================================
// INDUSTRY BENCHMARK SPECIFICATIONS
// =============================================================================

// IndustryBenchmarkSpec represents industry standard specifications
type IndustryBenchmarkSpec struct {
	Name             string
	Organization     string
	Version          string
	ReleaseDate      time.Time
	Categories       []string
	RequiredMetrics  []string
	OptionalMetrics  []string
	ReportingFormat  string
}

// GetMLPerfSpec returns MLPerf benchmark specification
func GetMLPerfSpec() *IndustryBenchmarkSpec {
	return &IndustryBenchmarkSpec{
		Name:         "MLPerf",
		Organization: "MLCommons",
		Version:      "5.1",
		ReleaseDate:  time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		Categories:   []string{"Training", "Inference-Datacenter", "Inference-Edge", "Mobile"},
		RequiredMetrics: []string{
			"samples_per_second",
			"time_to_train",
			"latency_p99",
			"accuracy",
		},
		OptionalMetrics: []string{
			"power_consumption",
			"energy_per_sample",
			"cost_per_inference",
		},
		ReportingFormat: "JSON",
	}
}

// GetAnalogNASBenchSpec returns AnalogNAS-Bench specification
func GetAnalogNASBenchSpec() *IndustryBenchmarkSpec {
	return &IndustryBenchmarkSpec{
		Name:         "AnalogNAS-Bench",
		Organization: "IBM Research",
		Version:      "1.0",
		ReleaseDate:  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		Categories:   []string{"AIMC-NAS", "Hardware-Aware"},
		RequiredMetrics: []string{
			"baseline_accuracy",
			"noisy_accuracy",
			"analog_accuracy",
			"noisy_drift_60s",
			"noisy_drift_1h",
			"noisy_drift_24h",
			"noisy_drift_30d",
		},
		OptionalMetrics: []string{
			"energy_efficiency",
			"latency",
			"area",
		},
		ReportingFormat: "HDF5",
	}
}
