// Package layers provides power and area estimation utilities for CIM accelerators.
// Implements models for FeFET crossbar arrays, peripherals, and full-chip estimation.
package layers

import (
	"fmt"
	"math"
)

// ============================================================================
// TECHNOLOGY PARAMETERS
// ============================================================================

// TechnologyNode represents CMOS/FeFET technology parameters.
type TechnologyNode struct {
	Name            string  // e.g., "28nm", "14nm", "7nm"
	NodeNm          float64 // Technology node in nm
	VDD             float64 // Supply voltage (V)
	FeFETArea       float64 // FeFET cell area (µm²)
	SRAMBitArea     float64 // SRAM 6T bit cell area (µm²)
	LogicGateArea   float64 // Average logic gate area (µm²)
	LeakagePower    float64 // Leakage power density (mW/mm²)
	DynamicEnergy   float64 // Dynamic energy per switch (fJ)
	ClockFreq       float64 // Maximum clock frequency (GHz)
	WireCapPerMm    float64 // Wire capacitance (fF/mm)
}

// GetTechnologyNode returns parameters for a given technology node.
func GetTechnologyNode(nodeNm float64) *TechnologyNode {
	switch {
	case nodeNm >= 28:
		return &TechnologyNode{
			Name:          "28nm",
			NodeNm:        28,
			VDD:           0.9,
			FeFETArea:     0.024,  // ~24F² for 1T FeFET
			SRAMBitArea:   0.127,  // ~127F² for 6T SRAM
			LogicGateArea: 0.05,
			LeakagePower:  0.5,
			DynamicEnergy: 10.0,
			ClockFreq:     1.0,
			WireCapPerMm:  200.0,
		}
	case nodeNm >= 14:
		return &TechnologyNode{
			Name:          "14nm",
			NodeNm:        14,
			VDD:           0.8,
			FeFETArea:     0.006,
			SRAMBitArea:   0.032,
			LogicGateArea: 0.012,
			LeakagePower:  0.8,
			DynamicEnergy: 5.0,
			ClockFreq:     2.0,
			WireCapPerMm:  250.0,
		}
	case nodeNm >= 7:
		return &TechnologyNode{
			Name:          "7nm",
			NodeNm:        7,
			VDD:           0.7,
			FeFETArea:     0.0015,
			SRAMBitArea:   0.008,
			LogicGateArea: 0.003,
			LeakagePower:  1.2,
			DynamicEnergy: 2.5,
			ClockFreq:     3.0,
			WireCapPerMm:  300.0,
		}
	default: // 5nm and below
		return &TechnologyNode{
			Name:          "5nm",
			NodeNm:        5,
			VDD:           0.65,
			FeFETArea:     0.0008,
			SRAMBitArea:   0.004,
			LogicGateArea: 0.0015,
			LeakagePower:  1.5,
			DynamicEnergy: 1.5,
			ClockFreq:     4.0,
			WireCapPerMm:  350.0,
		}
	}
}

// ============================================================================
// CROSSBAR ARRAY ESTIMATION
// ============================================================================

// CrossbarEstimationConfig configures crossbar estimation parameters.
type CrossbarEstimationConfig struct {
	Rows           int
	Cols           int
	BitsPerCell    int     // Weight bits per cell
	ADCBits        int
	DACBits        int
	TechNode       float64 // Technology node (nm)
	ArrayUtilization float64 // 0-1, fraction of array used
}

// DefaultCrossbarEstimationConfig returns typical estimation settings.
func DefaultCrossbarEstimationConfig() *CrossbarEstimationConfig {
	return &CrossbarEstimationConfig{
		Rows:            256,
		Cols:            256,
		BitsPerCell:     4,
		ADCBits:         6,
		DACBits:         8,
		TechNode:        28,
		ArrayUtilization: 0.8,
	}
}

// CrossbarEstimator estimates power and area for crossbar arrays.
type CrossbarEstimator struct {
	Config   *CrossbarEstimationConfig
	TechNode *TechnologyNode
}

// NewCrossbarEstimator creates a new estimator.
func NewCrossbarEstimator(config *CrossbarEstimationConfig) *CrossbarEstimator {
	return &CrossbarEstimator{
		Config:   config,
		TechNode: GetTechnologyNode(config.TechNode),
	}
}

// CrossbarEstimate contains power and area estimates.
type CrossbarEstimate struct {
	// Area breakdown (mm²)
	ArrayArea      float64
	ADCArea        float64
	DACArea        float64
	DriverArea     float64
	BufferArea     float64
	TotalArea      float64

	// Power breakdown (mW)
	ArrayPower     float64
	ADCPower       float64
	DACPower       float64
	DriverPower    float64
	LeakagePower   float64
	TotalPower     float64

	// Performance metrics
	ThroughputTOPS float64 // Tera-ops per second
	EfficiencyTOPSW float64 // TOPS per Watt
	EnergyPerMAC   float64 // fJ per MAC operation
	LatencyNs      float64 // Latency per MVM operation
}

// Estimate calculates power and area for the crossbar configuration.
func (e *CrossbarEstimator) Estimate() *CrossbarEstimate {
	cfg := e.Config
	tech := e.TechNode
	est := &CrossbarEstimate{}

	// =========== AREA ESTIMATION ===========

	// FeFET array area
	numCells := cfg.Rows * cfg.Cols
	cellsPerWeight := int(math.Ceil(float64(cfg.BitsPerCell) / 2.0)) // 2 bits per MLC cell
	arrayAreaUm2 := float64(numCells*cellsPerWeight) * tech.FeFETArea
	est.ArrayArea = arrayAreaUm2 / 1e6 // Convert to mm²

	// ADC area (one per column)
	// ADC area scales exponentially with bits: ~2^(bits-4) × base_area
	adcBaseArea := 0.001 // mm² for 4-bit ADC
	adcScale := math.Pow(2, float64(cfg.ADCBits-4))
	est.ADCArea = float64(cfg.Cols) * adcBaseArea * adcScale

	// DAC area (one per row)
	dacBaseArea := 0.0005 // mm² for 8-bit DAC
	dacScale := math.Pow(2, float64(cfg.DACBits-8))
	est.DACArea = float64(cfg.Rows) * dacBaseArea * dacScale

	// Row/column drivers
	driverAreaPerLine := 0.0001 // mm² per driver
	est.DriverArea = float64(cfg.Rows+cfg.Cols) * driverAreaPerLine

	// Buffers and control logic
	est.BufferArea = 0.1 * (est.ArrayArea + est.ADCArea + est.DACArea)

	est.TotalArea = est.ArrayArea + est.ADCArea + est.DACArea + est.DriverArea + est.BufferArea

	// =========== POWER ESTIMATION ===========

	// Clock frequency for this technology
	freqGHz := tech.ClockFreq

	// Array power (dominated by sensing current)
	// ~1 pJ per MAC for FeFET crossbar
	macPerCycle := float64(cfg.Rows * cfg.Cols)
	energyPerMACpJ := 1.0 * (tech.VDD / 0.9) * (tech.VDD / 0.9) // Scales with V²
	est.ArrayPower = macPerCycle * energyPerMACpJ * freqGHz / 1000 // mW

	// ADC power (dominant peripheral)
	// ~0.5 mW per 6-bit ADC at 1 GHz sample rate
	adcBasePower := 0.5 * math.Pow(2, float64(cfg.ADCBits-6))
	est.ADCPower = float64(cfg.Cols) * adcBasePower * freqGHz

	// DAC power
	dacBasePower := 0.1 * math.Pow(2, float64(cfg.DACBits-8))
	est.DACPower = float64(cfg.Rows) * dacBasePower * freqGHz

	// Driver power
	est.DriverPower = float64(cfg.Rows+cfg.Cols) * 0.01 * freqGHz

	// Leakage power
	est.LeakagePower = est.TotalArea * tech.LeakagePower

	est.TotalPower = est.ArrayPower + est.ADCPower + est.DACPower + est.DriverPower + est.LeakagePower

	// =========== PERFORMANCE METRICS ===========

	// Throughput: MACs per second
	macsPerSecond := macPerCycle * freqGHz * 1e9
	est.ThroughputTOPS = macsPerSecond / 1e12

	// Efficiency
	est.EfficiencyTOPSW = est.ThroughputTOPS / (est.TotalPower / 1000) // TOPS/W

	// Energy per MAC
	est.EnergyPerMAC = (est.TotalPower / 1000) / macsPerSecond * 1e15 // fJ

	// Latency per MVM
	cyclesPerMVM := 1 + cfg.ADCBits // 1 cycle compute + ADC conversion
	est.LatencyNs = float64(cyclesPerMVM) / freqGHz

	return est
}

// ============================================================================
// FULL CHIP ESTIMATION
// ============================================================================

// ChipConfig configures full chip estimation.
type ChipConfig struct {
	NumArrays        int
	ArrayConfig      *CrossbarEstimationConfig
	GlobalBufferKB   int
	InterconnectType string // "mesh", "bus", "noc"
	IOBandwidthGBps  float64
	ControlOverhead  float64 // Fraction for control logic
}

// DefaultChipConfig returns typical chip settings.
func DefaultChipConfig() *ChipConfig {
	return &ChipConfig{
		NumArrays:        64,
		ArrayConfig:      DefaultCrossbarEstimationConfig(),
		GlobalBufferKB:   512,
		InterconnectType: "mesh",
		IOBandwidthGBps:  100,
		ControlOverhead:  0.15,
	}
}

// ChipEstimate contains full-chip power and area estimates.
type ChipEstimate struct {
	// Array breakdown
	TotalArrayArea    float64
	TotalArrayPower   float64

	// Memory
	GlobalBufferArea  float64
	GlobalBufferPower float64

	// Interconnect
	InterconnectArea  float64
	InterconnectPower float64

	// I/O
	IOArea           float64
	IOPower          float64

	// Control
	ControlArea      float64
	ControlPower     float64

	// Totals
	TotalArea        float64 // mm²
	TotalPower       float64 // W
	DieSizeMm        float64 // Die edge length (mm)

	// Performance
	PeakThroughputTOPS float64
	PeakEfficiencyTOPSW float64
	SustainedThroughput float64 // With memory bottleneck
}

// ChipEstimator estimates full-chip metrics.
type ChipEstimator struct {
	Config          *ChipConfig
	CrossbarEst     *CrossbarEstimator
}

// NewChipEstimator creates a new chip estimator.
func NewChipEstimator(config *ChipConfig) *ChipEstimator {
	return &ChipEstimator{
		Config:      config,
		CrossbarEst: NewCrossbarEstimator(config.ArrayConfig),
	}
}

// Estimate calculates full-chip metrics.
func (e *ChipEstimator) Estimate() *ChipEstimate {
	cfg := e.Config
	tech := GetTechnologyNode(cfg.ArrayConfig.TechNode)
	est := &ChipEstimate{}

	// Get single array estimate
	arrayEst := e.CrossbarEst.Estimate()

	// =========== ARRAY TOTALS ===========
	est.TotalArrayArea = arrayEst.TotalArea * float64(cfg.NumArrays)
	est.TotalArrayPower = arrayEst.TotalPower * float64(cfg.NumArrays) / 1000 // W

	// =========== GLOBAL BUFFER ===========
	bufferBits := float64(cfg.GlobalBufferKB * 1024 * 8)
	est.GlobalBufferArea = bufferBits * tech.SRAMBitArea / 1e6 // mm²
	// SRAM power: ~0.5 pJ/bit access at 28nm
	bufferAccessPerSec := cfg.IOBandwidthGBps * 1e9 * 8 // bits/sec
	est.GlobalBufferPower = bufferAccessPerSec * 0.5e-12 // W

	// =========== INTERCONNECT ===========
	// Mesh interconnect scales with sqrt(arrays)
	meshSide := math.Sqrt(float64(cfg.NumArrays))
	interconnectLength := meshSide * math.Sqrt(est.TotalArrayArea) // mm
	est.InterconnectArea = interconnectLength * 0.1 // 0.1 mm wide channels
	// Wire power
	wireCap := interconnectLength * tech.WireCapPerMm * 1e-15 // F
	switchFreq := tech.ClockFreq * 1e9
	est.InterconnectPower = 0.5 * wireCap * tech.VDD * tech.VDD * switchFreq

	// =========== I/O ===========
	// Assume SerDes at 25 Gbps per lane
	numLanes := int(math.Ceil(cfg.IOBandwidthGBps * 8 / 25))
	est.IOArea = float64(numLanes) * 0.1 // 0.1 mm² per SerDes lane
	est.IOPower = float64(numLanes) * 0.1 // 100 mW per lane

	// =========== CONTROL ===========
	coreArea := est.TotalArrayArea + est.GlobalBufferArea + est.InterconnectArea
	est.ControlArea = coreArea * cfg.ControlOverhead
	est.ControlPower = (est.TotalArrayPower + est.GlobalBufferPower) * cfg.ControlOverhead

	// =========== TOTALS ===========
	est.TotalArea = est.TotalArrayArea + est.GlobalBufferArea + est.InterconnectArea +
		est.IOArea + est.ControlArea
	est.TotalPower = est.TotalArrayPower + est.GlobalBufferPower + est.InterconnectPower +
		est.IOPower + est.ControlPower

	est.DieSizeMm = math.Sqrt(est.TotalArea)

	// =========== PERFORMANCE ===========
	est.PeakThroughputTOPS = arrayEst.ThroughputTOPS * float64(cfg.NumArrays)
	est.PeakEfficiencyTOPSW = est.PeakThroughputTOPS / est.TotalPower

	// Sustained throughput limited by memory bandwidth
	bytesPerMAC := 2.0 // Input + output
	memBoundMACs := cfg.IOBandwidthGBps * 1e9 / bytesPerMAC
	est.SustainedThroughput = math.Min(est.PeakThroughputTOPS, memBoundMACs/1e12)

	return est
}

// ============================================================================
// MODEL ESTIMATION
// ============================================================================

// ModelEstimationConfig configures model-level estimation.
type ModelEstimationConfig struct {
	LayerSizes    []int   // Neurons per layer
	BatchSize     int
	Sparsity      float64 // Weight sparsity (0-1)
	Quantization  int     // Bits per weight
}

// ModelEstimate contains model deployment estimates.
type ModelEstimate struct {
	// Resource requirements
	TotalWeights     int
	TotalMACs        int64
	RequiredArrays   int
	MemoryBytes      int64

	// Per-inference metrics
	InferenceLatencyUs float64
	InferenceEnergyUJ  float64

	// Throughput
	InferencesPerSec   float64
	BatchLatencyUs     float64
}

// ModelEstimator estimates model deployment on CIM.
type ModelEstimator struct {
	ModelConfig  *ModelEstimationConfig
	ChipConfig   *ChipConfig
	ChipEstimate *ChipEstimate
}

// NewModelEstimator creates a new model estimator.
func NewModelEstimator(modelConfig *ModelEstimationConfig, chipConfig *ChipConfig) *ModelEstimator {
	chipEst := NewChipEstimator(chipConfig)
	return &ModelEstimator{
		ModelConfig:  modelConfig,
		ChipConfig:   chipConfig,
		ChipEstimate: chipEst.Estimate(),
	}
}

// Estimate calculates model deployment metrics.
func (e *ModelEstimator) Estimate() *ModelEstimate {
	cfg := e.ModelConfig
	chip := e.ChipConfig
	est := &ModelEstimate{}

	// Count weights and MACs
	for i := 0; i < len(cfg.LayerSizes)-1; i++ {
		layerWeights := cfg.LayerSizes[i] * cfg.LayerSizes[i+1]
		est.TotalWeights += layerWeights
		est.TotalMACs += int64(layerWeights) * int64(cfg.BatchSize)
	}

	// Account for sparsity
	effectiveMACs := float64(est.TotalMACs) * (1 - cfg.Sparsity)

	// Memory for weights
	bitsPerWeight := cfg.Quantization
	est.MemoryBytes = int64(est.TotalWeights * bitsPerWeight / 8)

	// Required arrays
	arraySize := chip.ArrayConfig.Rows * chip.ArrayConfig.Cols
	est.RequiredArrays = int(math.Ceil(float64(est.TotalWeights) / float64(arraySize)))

	// Check if model fits
	if est.RequiredArrays > chip.NumArrays {
		// Model needs to be tiled across time
		tilingFactor := float64(est.RequiredArrays) / float64(chip.NumArrays)
		effectiveMACs *= tilingFactor
	}

	// Inference latency
	// Assume pipelined execution: latency = layers × array_latency
	numLayers := len(cfg.LayerSizes) - 1
	arrayLatencyNs := 10.0 // ~10ns per MVM
	est.InferenceLatencyUs = float64(numLayers) * arrayLatencyNs / 1000

	// Energy per inference
	energyPerMAC := e.ChipEstimate.TotalPower / (e.ChipEstimate.PeakThroughputTOPS * 1e12) // J
	est.InferenceEnergyUJ = effectiveMACs * energyPerMAC * 1e6 / float64(cfg.BatchSize)

	// Throughput
	est.InferencesPerSec = e.ChipEstimate.SustainedThroughput * 1e12 / effectiveMACs * float64(cfg.BatchSize)
	est.BatchLatencyUs = float64(cfg.BatchSize) / est.InferencesPerSec * 1e6

	return est
}

// ============================================================================
// COMPARISON UTILITIES
// ============================================================================

// AcceleratorType represents different accelerator types.
type AcceleratorType int

const (
	AcceleratorFeFETCIM AcceleratorType = iota
	AcceleratorRRAMCIM
	AcceleratorGPU
	AcceleratorTPU
	AcceleratorCPU
)

// AcceleratorProfile contains typical accelerator characteristics.
type AcceleratorProfile struct {
	Type           AcceleratorType
	Name           string
	PeakTOPS       float64
	TDP            float64 // Watts
	EfficiencyTOPSW float64
	AreaMm2        float64
	CostUSD        float64
	MemoryGB       float64
}

// GetAcceleratorProfile returns typical characteristics.
func GetAcceleratorProfile(accType AcceleratorType) *AcceleratorProfile {
	switch accType {
	case AcceleratorFeFETCIM:
		return &AcceleratorProfile{
			Type:            AcceleratorFeFETCIM,
			Name:            "FeFET CIM (64 arrays)",
			PeakTOPS:        100,
			TDP:             5,
			EfficiencyTOPSW: 20,
			AreaMm2:         25,
			CostUSD:         50,
			MemoryGB:        0.5,
		}
	case AcceleratorRRAMCIM:
		return &AcceleratorProfile{
			Type:            AcceleratorRRAMCIM,
			Name:            "RRAM CIM (64 arrays)",
			PeakTOPS:        80,
			TDP:             8,
			EfficiencyTOPSW: 10,
			AreaMm2:         30,
			CostUSD:         60,
			MemoryGB:        0.5,
		}
	case AcceleratorGPU:
		return &AcceleratorProfile{
			Type:            AcceleratorGPU,
			Name:            "NVIDIA A100",
			PeakTOPS:        312, // INT8
			TDP:             400,
			EfficiencyTOPSW: 0.78,
			AreaMm2:         826,
			CostUSD:         10000,
			MemoryGB:        80,
		}
	case AcceleratorTPU:
		return &AcceleratorProfile{
			Type:            AcceleratorTPU,
			Name:            "Google TPU v4",
			PeakTOPS:        275,
			TDP:             175,
			EfficiencyTOPSW: 1.57,
			AreaMm2:         400,
			CostUSD:         5000,
			MemoryGB:        32,
		}
	case AcceleratorCPU:
		return &AcceleratorProfile{
			Type:            AcceleratorCPU,
			Name:            "Intel Xeon (AVX-512)",
			PeakTOPS:        3,
			TDP:             250,
			EfficiencyTOPSW: 0.012,
			AreaMm2:         700,
			CostUSD:         3000,
			MemoryGB:        512,
		}
	default:
		return GetAcceleratorProfile(AcceleratorFeFETCIM)
	}
}

// CompareAccelerators returns comparison of all accelerator types.
func CompareAccelerators() []*AcceleratorProfile {
	return []*AcceleratorProfile{
		GetAcceleratorProfile(AcceleratorFeFETCIM),
		GetAcceleratorProfile(AcceleratorRRAMCIM),
		GetAcceleratorProfile(AcceleratorGPU),
		GetAcceleratorProfile(AcceleratorTPU),
		GetAcceleratorProfile(AcceleratorCPU),
	}
}

// ============================================================================
// REPORT GENERATION
// ============================================================================

// EstimationReport contains complete estimation results.
type EstimationReport struct {
	CrossbarEst *CrossbarEstimate
	ChipEst     *ChipEstimate
	ModelEst    *ModelEstimate
	Comparison  []*AcceleratorProfile
}

// GenerateReport creates a complete estimation report.
func GenerateReport(modelLayers []int, batchSize int) *EstimationReport {
	// Crossbar estimation
	crossbarConfig := DefaultCrossbarEstimationConfig()
	crossbarEst := NewCrossbarEstimator(crossbarConfig)

	// Chip estimation
	chipConfig := DefaultChipConfig()
	chipEst := NewChipEstimator(chipConfig)

	// Model estimation
	modelConfig := &ModelEstimationConfig{
		LayerSizes:   modelLayers,
		BatchSize:    batchSize,
		Sparsity:     0.5,
		Quantization: 6,
	}
	modelEst := NewModelEstimator(modelConfig, chipConfig)

	return &EstimationReport{
		CrossbarEst: crossbarEst.Estimate(),
		ChipEst:     chipEst.Estimate(),
		ModelEst:    modelEst.Estimate(),
		Comparison:  CompareAccelerators(),
	}
}

// PrintReport outputs the estimation report.
func (r *EstimationReport) PrintReport() string {
	report := "=== CIM Accelerator Estimation Report ===\n\n"

	// Crossbar section
	report += "--- Single Crossbar Array ---\n"
	report += fmt.Sprintf("Array Area:     %.4f mm²\n", r.CrossbarEst.TotalArea)
	report += fmt.Sprintf("Array Power:    %.2f mW\n", r.CrossbarEst.TotalPower)
	report += fmt.Sprintf("Throughput:     %.2f TOPS\n", r.CrossbarEst.ThroughputTOPS)
	report += fmt.Sprintf("Efficiency:     %.1f TOPS/W\n", r.CrossbarEst.EfficiencyTOPSW)
	report += fmt.Sprintf("Energy/MAC:     %.2f fJ\n", r.CrossbarEst.EnergyPerMAC)
	report += fmt.Sprintf("Latency:        %.1f ns\n\n", r.CrossbarEst.LatencyNs)

	// Chip section
	report += "--- Full Chip ---\n"
	report += fmt.Sprintf("Total Area:     %.2f mm² (%.1f mm × %.1f mm)\n",
		r.ChipEst.TotalArea, r.ChipEst.DieSizeMm, r.ChipEst.DieSizeMm)
	report += fmt.Sprintf("Total Power:    %.2f W\n", r.ChipEst.TotalPower)
	report += fmt.Sprintf("Peak TOPS:      %.1f\n", r.ChipEst.PeakThroughputTOPS)
	report += fmt.Sprintf("Peak Eff:       %.1f TOPS/W\n", r.ChipEst.PeakEfficiencyTOPSW)
	report += fmt.Sprintf("Sustained TOPS: %.1f\n\n", r.ChipEst.SustainedThroughput)

	// Model section
	report += "--- Model Deployment ---\n"
	report += fmt.Sprintf("Total Weights:  %d\n", r.ModelEst.TotalWeights)
	report += fmt.Sprintf("Total MACs:     %d\n", r.ModelEst.TotalMACs)
	report += fmt.Sprintf("Required Arrays: %d\n", r.ModelEst.RequiredArrays)
	report += fmt.Sprintf("Latency:        %.2f µs\n", r.ModelEst.InferenceLatencyUs)
	report += fmt.Sprintf("Energy:         %.2f µJ\n", r.ModelEst.InferenceEnergyUJ)
	report += fmt.Sprintf("Throughput:     %.0f inf/s\n\n", r.ModelEst.InferencesPerSec)

	// Comparison section
	report += "--- Accelerator Comparison ---\n"
	report += fmt.Sprintf("%-20s %8s %8s %10s\n", "Accelerator", "TOPS", "TDP(W)", "TOPS/W")
	for _, acc := range r.Comparison {
		report += fmt.Sprintf("%-20s %8.1f %8.0f %10.2f\n",
			acc.Name, acc.PeakTOPS, acc.TDP, acc.EfficiencyTOPSW)
	}

	return report
}

// ============================================================================
// NEUROMORPHIC VISION ESTIMATION
// ============================================================================

// VisionSensorConfig configures neuromorphic vision sensor estimation.
type VisionSensorConfig struct {
	ResolutionX     int
	ResolutionY     int
	PixelPitchUm    float64
	EventRateHz     float64 // Average events per second
	ProcessingMode  string  // "frame", "event", "hybrid"
	IntegratedCIM   bool    // CIM in sensor
}

// VisionSensorEstimate contains vision sensor metrics.
type VisionSensorEstimate struct {
	SensorAreaMm2      float64
	PowerMW            float64
	EventBandwidthMbps float64
	LatencyUs          float64
	EnergyPerEventPJ   float64
}

// EstimateVisionSensor calculates vision sensor metrics.
func EstimateVisionSensor(config *VisionSensorConfig) *VisionSensorEstimate {
	est := &VisionSensorEstimate{}

	// Sensor area
	pixelAreaUm2 := config.PixelPitchUm * config.PixelPitchUm
	totalPixels := config.ResolutionX * config.ResolutionY
	est.SensorAreaMm2 = float64(totalPixels) * pixelAreaUm2 / 1e6

	// Event bandwidth (assume 8 bytes per event: x, y, timestamp, polarity)
	bytesPerEvent := 8.0
	est.EventBandwidthMbps = config.EventRateHz * bytesPerEvent * 8 / 1e6

	// Power estimation
	basePowerPerPixel := 0.001 // 1 µW per pixel baseline
	eventPowerPerPixel := 0.01 // 10 µW per pixel at max event rate
	eventFraction := math.Min(1.0, config.EventRateHz/1e6)
	est.PowerMW = float64(totalPixels) * (basePowerPerPixel + eventFraction*eventPowerPerPixel)

	if config.IntegratedCIM {
		// Add CIM processing power
		est.PowerMW *= 1.5
	}

	// Latency (event detection + readout)
	est.LatencyUs = 1.0 // ~1 µs event latency

	// Energy per event
	est.EnergyPerEventPJ = est.PowerMW * 1000 / config.EventRateHz * 1e6

	return est
}

// ============================================================================
// GNN ESTIMATION
// ============================================================================

// GNNConfig configures GNN estimation on CIM.
type GNNConfig struct {
	NumNodes       int
	NumEdges       int
	FeatureDim     int
	HiddenDim      int
	NumLayers      int
	AggregationType string // "mean", "sum", "max"
}

// GNNEstimate contains GNN deployment metrics.
type GNNEstimate struct {
	AggregationMACs  int64
	TransformMACs    int64
	TotalMACs        int64
	MemoryBytes      int64
	RequiredArrays   int
	SparsityUtilization float64
}

// EstimateGNN calculates GNN deployment on CIM.
func EstimateGNN(config *GNNConfig, chipConfig *ChipConfig) *GNNEstimate {
	est := &GNNEstimate{}

	// Aggregation: for each node, aggregate neighbors
	avgDegree := float64(config.NumEdges) / float64(config.NumNodes)
	aggregationOpsPerLayer := float64(config.NumNodes) * avgDegree * float64(config.FeatureDim)
	est.AggregationMACs = int64(aggregationOpsPerLayer) * int64(config.NumLayers)

	// Transform: W × features for each node
	transformOpsPerLayer := float64(config.NumNodes) * float64(config.FeatureDim) * float64(config.HiddenDim)
	est.TransformMACs = int64(transformOpsPerLayer) * int64(config.NumLayers)

	est.TotalMACs = est.AggregationMACs + est.TransformMACs

	// Memory: node features + edge list + weights
	featureBytes := config.NumNodes * config.HiddenDim * 4 // float32
	edgeBytes := config.NumEdges * 8                       // 2 × int32 per edge
	weightBytes := config.FeatureDim * config.HiddenDim * config.NumLayers * 4
	est.MemoryBytes = int64(featureBytes + edgeBytes + weightBytes)

	// Required arrays (for weight matrices)
	arraySize := chipConfig.ArrayConfig.Rows * chipConfig.ArrayConfig.Cols
	weightsTotal := config.FeatureDim * config.HiddenDim * config.NumLayers
	est.RequiredArrays = int(math.Ceil(float64(weightsTotal) / float64(arraySize)))

	// Sparsity utilization (adjacency matrix sparsity)
	densityAdjMatrix := float64(config.NumEdges) / float64(config.NumNodes*config.NumNodes)
	est.SparsityUtilization = densityAdjMatrix

	return est
}
