package comparison

import (
	"math"

	"fecim-lattice-tools/shared/logging"
)

var log = logging.NewLogger("comparison")

// Architecture represents a compute architecture for neural network inference.
type Architecture struct {
	Name        string // Architecture name
	Description string // Brief description
	Technology  string // Underlying technology

	// Physical parameters
	ProcessNode float64 // Process node (nm)
	ChipArea    float64 // Chip area (mm²)
	TDP         float64 // Thermal design power (W)

	// Performance parameters
	PeakTOPS   float64 // Peak operations per second (TOPS)
	MemoryBW   float64 // Memory bandwidth (GB/s)
	MemorySize float64 // Memory size (GB)

	// Efficiency metrics (calculated or specified)
	TOPSPerWatt float64 // Energy efficiency (TOPS/W)
	TOPSPerMM2  float64 // Area efficiency (TOPS/mm²)

	// Cost factors
	ManufactureCost float64 // Relative manufacturing cost
	PowerCost       float64 // Power consumption cost factor

	// Data quality
	IsEstimated bool // True if values are model inputs or estimates (not validated)
}

// TraditionalCPU creates a traditional CPU + DRAM architecture.
func TraditionalCPU() *Architecture {
	return &Architecture{
		Name:            "Traditional CPU+DRAM",
		Description:     "Von Neumann architecture with DDR5 memory",
		Technology:      "CMOS + DRAM",
		ProcessNode:     5,
		ChipArea:        400, // Large CPU die
		TDP:             125, // Typical server CPU
		PeakTOPS:        1.0, // ~1 TOPS for matrix ops
		MemoryBW:        100, // DDR5 bandwidth
		MemorySize:      128, // Server memory
		TOPSPerWatt:     0.008,
		TOPSPerMM2:      0.0025,
		ManufactureCost: 1.0, // Baseline
		PowerCost:       1.0,
	}
}

// GPUAccelerator creates a GPU-based accelerator architecture.
func GPUAccelerator() *Architecture {
	return &Architecture{
		Name:            "GPU Accelerator",
		Description:     "High-performance GPU with HBM memory",
		Technology:      "CMOS + HBM",
		ProcessNode:     4,
		ChipArea:        800,  // Large GPU die
		TDP:             400,  // High-end GPU
		PeakTOPS:        100,  // Modern AI GPU
		MemoryBW:        2000, // HBM bandwidth
		MemorySize:      80,   // HBM capacity
		TOPSPerWatt:     0.25,
		TOPSPerMM2:      0.125,
		ManufactureCost: 3.0, // Expensive
		PowerCost:       3.2,
	}
}

// FeCIMChip creates an FeCIM compute-in-memory architecture.
//
// ⚠️ MODEL INPUTS ONLY (NOT VALIDATED) ⚠️
//
// FeCIM is at TRL 4 (lab validation only). Project materials do not disclose:
//   - TDP (power consumption)
//   - TOPS (performance)
//   - TOPS/W (efficiency)
//   - Chip area
//   - Any chip-level specifications
//
// The values below are demo inputs for visualization purposes only.
// They should NOT be presented as validated device specs or used for investment decisions.
//
// Context-only references (not validation or proof):
//   - 30 discrete analog states (conference claim, pending peer review)
//   - Peer-reviewed ranges for analog states / MNIST accuracy exist, but are
//     not used here as verified device specifications.
//
// See docs/4-research/honesty-audit.md for full analysis.
func FeCIMChip() *Architecture {
	return &Architecture{
		Name:            "FeCIM CIM",
		Description:     "Ferroelectric compute-in-memory with 30-level baseline (model input; TRL4)",
		Technology:      "FeFET Crossbar",
		ProcessNode:     45,   // ESTIMATED - not disclosed
		ChipArea:        50,   // ESTIMATED - not disclosed by FeCIM
		TDP:             5,    // ESTIMATED - not disclosed by FeCIM
		PeakTOPS:        50,   // ESTIMATED - not disclosed by FeCIM
		MemoryBW:        0,    // Compute-in-memory eliminates bottleneck (correct concept)
		MemorySize:      1,    // ESTIMATED - weights stored in-situ
		TOPSPerWatt:     10,   // ESTIMATED - derived from estimates above
		TOPSPerMM2:      1.0,  // ESTIMATED - derived from estimates above
		ManufactureCost: 0.3,  // ESTIMATED - not disclosed
		PowerCost:       0.04, // ESTIMATED - not disclosed
		IsEstimated:     true, // TRL 4 - specs are model inputs/estimates
	}
}

// CustomArchitecture creates a custom architecture with specified parameters.
func CustomArchitecture(name string, tops, power, area float64) *Architecture {
	var topsPerWatt, topsPerMM2 float64
	if power > 0 {
		topsPerWatt = tops / power
	}
	if area > 0 {
		topsPerMM2 = tops / area
	}
	return &Architecture{
		Name:        name,
		ProcessNode: 7,
		ChipArea:    area,
		TDP:         power,
		PeakTOPS:    tops,
		TOPSPerWatt: topsPerWatt,
		TOPSPerMM2:  topsPerMM2,
	}
}

// InferenceResult contains results from running inference on an architecture.
type InferenceResult struct {
	Architecture string  // Architecture name
	ModelOps     int     // Operations in model (MACs)
	BatchSize    int     // Batch size
	Latency      float64 // Latency (ms)
	Throughput   float64 // Throughput (inferences/sec)
	Energy       float64 // Energy per inference (mJ)
	PowerUsed    float64 // Power used (W)
}

// RunInference simulates running inference on the architecture.
func (a *Architecture) RunInference(modelOps int, batchSize int) InferenceResult {
	log.Input("RunInference", map[string]interface{}{
		"architecture": a.Name,
		"modelOps":     modelOps,
		"batchSize":    batchSize,
	})

	// Operations in TOPS
	opsInTOPS := float64(modelOps) * float64(batchSize) / 1e12

	// Latency = ops / throughput
	// Account for memory bottleneck in traditional architectures
	computeLatency := opsInTOPS / a.PeakTOPS * 1000 // ms

	memoryLatency := 0.0
	if a.MemoryBW > 0 && a.Technology != "FeFET Crossbar" {
		// Weight movement overhead (assumes weights larger than on-chip cache)
		weightBytes := float64(modelOps) * 2 / float64(batchSize) // FP16
		memoryTime := weightBytes / (a.MemoryBW * 1e9) * 1000     // ms
		memoryLatency = memoryTime
	}

	totalLatency := computeLatency + memoryLatency

	// Energy = power * time
	// W * ms = mJ
	energy := a.TDP * totalLatency // mJ

	// Throughput
	throughput := float64(batchSize) / (totalLatency / 1000) // inferences/sec

	result := InferenceResult{
		Architecture: a.Name,
		ModelOps:     modelOps,
		BatchSize:    batchSize,
		Latency:      totalLatency,
		Throughput:   throughput,
		Energy:       energy,
		PowerUsed:    a.TDP,
	}

	log.Calculation("RunInference", map[string]interface{}{
		"computeLatency": computeLatency,
		"memoryLatency":  memoryLatency,
		"totalLatency":   totalLatency,
		"energy":         energy,
		"throughput":     throughput,
	}, result)

	return result
}

// Workload defines a neural network workload.
type Workload struct {
	Name        string // Workload name
	Description string // Description
	TotalOps    int    // Total operations (MACs)
	Layers      int    // Number of layers
	Parameters  int    // Number of parameters
}

// MNISTWorkload creates an MNIST classification workload.
func MNISTWorkload() Workload {
	// 784-128-10 network
	// Layer 1: 784*128 = 100352 MACs
	// Layer 2: 128*10 = 1280 MACs
	return Workload{
		Name:        "MNIST",
		Description: "Handwritten digit recognition",
		TotalOps:    101632, // 100352 + 1280
		Layers:      2,
		Parameters:  101632,
	}
}

// ResNet50Workload creates a ResNet-50 workload.
func ResNet50Workload() Workload {
	return Workload{
		Name:        "ResNet-50",
		Description: "Image classification (ImageNet)",
		TotalOps:    4e9, // ~4 billion MACs
		Layers:      50,
		Parameters:  25600000, // 25.6M parameters
	}
}

// BERTBaseWorkload creates a BERT-Base workload.
func BERTBaseWorkload() Workload {
	return Workload{
		Name:        "BERT-Base",
		Description: "Natural language processing",
		TotalOps:    11e9, // ~11 billion MACs per sequence
		Layers:      12,
		Parameters:  110000000, // 110M parameters
	}
}

// GPT2Workload creates a GPT-2 workload.
func GPT2Workload() Workload {
	return Workload{
		Name:        "GPT-2",
		Description: "Language model inference",
		TotalOps:    35e9, // ~35 billion MACs per token
		Layers:      12,
		Parameters:  1500000000, // 1.5B parameters
	}
}

// LLMWorkload creates a large language model workload.
func LLMWorkload() Workload {
	return Workload{
		Name:        "LLM-70B",
		Description: "Large language model (70B params)",
		TotalOps:    140e12, // ~140 trillion MACs
		Layers:      80,
		Parameters:  70000000000, // 70B parameters
	}
}

// DataCenterMetrics contains metrics for data center scale.
type DataCenterMetrics struct {
	Architecture     string  // Architecture name
	ChipsRequired    int     // Number of chips needed
	TotalPower       float64 // Total power (kW)
	RackSpace        int     // Rack units required
	InferencesPerSec float64 // Total throughput
	CostPerInference float64 // Cost per inference ($)
	TCO              float64 // Total cost of ownership ($/year)
	CO2Emissions     float64 // kg CO2 per day
}

// ScaleToDataCenter calculates data center scale metrics.
func ScaleToDataCenter(arch *Architecture, targetThroughput float64, workload Workload) DataCenterMetrics {
	log.Input("ScaleToDataCenter", map[string]interface{}{
		"architecture":     arch.Name,
		"targetThroughput": targetThroughput,
		"workload":         workload.Name,
		"workloadTotalOps": workload.TotalOps,
	})

	// Run single chip inference
	result := arch.RunInference(workload.TotalOps, 1)

	// Chips needed for target throughput
	chipsRequired := 1
	if result.Throughput > 0 {
		chipsRequired = int(math.Ceil(targetThroughput / result.Throughput))
	}
	if chipsRequired < 1 {
		chipsRequired = 1
	}

	// Total power (kW)
	totalPower := arch.TDP * float64(chipsRequired) / 1000

	// Rack space (assume 1 chip = 1U for simplicity, GPUs take more)
	rackUnits := chipsRequired
	if arch.Technology == "CMOS + HBM" {
		rackUnits = chipsRequired * 2 // GPUs need more space
	}

	// Actual throughput
	actualThroughput := result.Throughput * float64(chipsRequired)

	// Cost per inference (electricity cost ~$0.10/kWh)
	electricityCost := 0.10
	energyPerInference := result.Energy / 3.6e9 // Convert mJ to kWh
	costPerInference := energyPerInference * electricityCost

	// TCO (assume 3 year depreciation, include cooling 1.5x PUE)
	chipCost := arch.ManufactureCost * 10000 // Relative cost
	electricityCostYear := totalPower * 24 * 365 * electricityCost * 1.5
	depreciation := chipCost * float64(chipsRequired) / 3
	tco := electricityCostYear + depreciation

	// CO2 emissions (assume 0.4 kg CO2/kWh)
	co2PerKWh := 0.4
	co2PerDay := totalPower * 24 * co2PerKWh

	metrics := DataCenterMetrics{
		Architecture:     arch.Name,
		ChipsRequired:    chipsRequired,
		TotalPower:       totalPower,
		RackSpace:        rackUnits,
		InferencesPerSec: actualThroughput,
		CostPerInference: costPerInference,
		TCO:              tco,
		CO2Emissions:     co2PerDay,
	}

	log.Calculation("ScaleToDataCenter", map[string]interface{}{
		"chipsRequired":    chipsRequired,
		"totalPower_kW":    totalPower,
		"rackUnits":        rackUnits,
		"actualThroughput": actualThroughput,
		"tco":              tco,
		"co2PerDay":        co2PerDay,
	}, metrics)

	return metrics
}

// ComparisonResult contains full comparison between architectures.
type ComparisonResult struct {
	Workload      Workload
	BatchSize     int
	Architectures []*Architecture
	Results       []InferenceResult
	DataCenter    []DataCenterMetrics
}

// CompareArchitectures runs a full comparison.
func CompareArchitectures(workload Workload, batchSize int, targetThroughput float64) ComparisonResult {
	log.Input("CompareArchitectures", map[string]interface{}{
		"workload":         workload.Name,
		"batchSize":        batchSize,
		"targetThroughput": targetThroughput,
	})

	architectures := []*Architecture{
		TraditionalCPU(),
		GPUAccelerator(),
		FeCIMChip(),
	}

	results := make([]InferenceResult, len(architectures))
	dcMetrics := make([]DataCenterMetrics, len(architectures))

	for i, arch := range architectures {
		results[i] = arch.RunInference(workload.TotalOps, batchSize)
		dcMetrics[i] = ScaleToDataCenter(arch, targetThroughput, workload)
	}

	comparison := ComparisonResult{
		Workload:      workload,
		BatchSize:     batchSize,
		Architectures: architectures,
		Results:       results,
		DataCenter:    dcMetrics,
	}

	log.Calculation("CompareArchitectures", map[string]interface{}{
		"architectures": len(architectures),
		"cpu_latency":   results[0].Latency,
		"gpu_latency":   results[1].Latency,
		"fecim_latency": results[2].Latency,
	}, comparison)

	return comparison
}

// FeCIMAdvantage calculates FeCIM advantages over other architectures.
type FeCIMAdvantage struct {
	VsCPU struct {
		EnergyReduction    float64
		LatencyReduction   float64
		ThroughputIncrease float64
		PowerReduction     float64
		CostReduction      float64
	}
	VsGPU struct {
		EnergyReduction  float64
		LatencyReduction float64
		AreaReduction    float64
		PowerReduction   float64
		CostReduction    float64
	}
}

// CalculateAdvantages calculates FeCIM advantages.
// safeRatio returns a/b, or 0 if b is zero (prevents div-by-zero in comparisons).
func safeRatio(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

func CalculateAdvantages(comparison ComparisonResult) FeCIMAdvantage {
	log.Input("CalculateAdvantages", map[string]interface{}{
		"workload": comparison.Workload.Name,
	})

	var adv FeCIMAdvantage

	var cpuResult, gpuResult, fecimResult InferenceResult
	var cpuDC, gpuDC, fecimDC DataCenterMetrics
	var cpuArch, gpuArch, fecimArch *Architecture

	for i, arch := range comparison.Architectures {
		switch arch.Name {
		case "Traditional CPU+DRAM":
			cpuResult = comparison.Results[i]
			cpuDC = comparison.DataCenter[i]
			cpuArch = arch
		case "GPU Accelerator":
			gpuResult = comparison.Results[i]
			gpuDC = comparison.DataCenter[i]
			gpuArch = arch
		case "FeCIM CIM":
			fecimResult = comparison.Results[i]
			fecimDC = comparison.DataCenter[i]
			fecimArch = arch
		}
	}

	// vs CPU (all ratios guarded against zero denominators)
	if cpuArch != nil && fecimArch != nil {
		adv.VsCPU.EnergyReduction = safeRatio(cpuResult.Energy, fecimResult.Energy)
		adv.VsCPU.LatencyReduction = safeRatio(cpuResult.Latency, fecimResult.Latency)
		adv.VsCPU.ThroughputIncrease = safeRatio(fecimResult.Throughput, cpuResult.Throughput)
		adv.VsCPU.PowerReduction = safeRatio(cpuArch.TDP, fecimArch.TDP)
		adv.VsCPU.CostReduction = safeRatio(cpuDC.TCO, fecimDC.TCO)
	}

	// vs GPU (all ratios guarded against zero denominators)
	if gpuArch != nil && fecimArch != nil {
		adv.VsGPU.EnergyReduction = safeRatio(gpuResult.Energy, fecimResult.Energy)
		adv.VsGPU.LatencyReduction = safeRatio(gpuResult.Latency, fecimResult.Latency)
		adv.VsGPU.AreaReduction = safeRatio(gpuArch.ChipArea, fecimArch.ChipArea)
		adv.VsGPU.PowerReduction = safeRatio(gpuArch.TDP, fecimArch.TDP)
		adv.VsGPU.CostReduction = safeRatio(gpuDC.TCO, fecimDC.TCO)
	}

	log.Calculation("CalculateAdvantages", map[string]interface{}{
		"vsCPU_energy":     adv.VsCPU.EnergyReduction,
		"vsCPU_throughput": adv.VsCPU.ThroughputIncrease,
		"vsGPU_energy":     adv.VsGPU.EnergyReduction,
		"vsGPU_area":       adv.VsGPU.AreaReduction,
	}, adv)

	return adv
}
