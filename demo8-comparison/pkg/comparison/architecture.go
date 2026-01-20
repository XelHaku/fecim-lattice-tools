package comparison

import (
	"math"
)

// Architecture represents a compute architecture for neural network inference.
type Architecture struct {
	Name        string  // Architecture name
	Description string  // Brief description
	Technology  string  // Underlying technology

	// Physical parameters
	ProcessNode float64 // Process node (nm)
	ChipArea    float64 // Chip area (mm²)
	TDP         float64 // Thermal design power (W)

	// Performance parameters
	PeakTOPS    float64 // Peak operations per second (TOPS)
	MemoryBW    float64 // Memory bandwidth (GB/s)
	MemorySize  float64 // Memory size (GB)

	// Efficiency metrics (calculated or specified)
	TOPSPerWatt float64 // Energy efficiency (TOPS/W)
	TOPSPerMM2  float64 // Area efficiency (TOPS/mm²)

	// Cost factors
	ManufactureCost float64 // Relative manufacturing cost
	PowerCost       float64 // Power consumption cost factor
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
		ChipArea:        800, // Large GPU die
		TDP:             400, // High-end GPU
		PeakTOPS:        100, // Modern AI GPU
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
// ⚠️ WARNING: ESTIMATED VALUES - NO BASIS IN DR. TOUR'S PRESENTATION ⚠️
//
// FeCIM is at TRL 4 (lab validation only). Dr. Tour did NOT disclose:
//   - TDP (power consumption)
//   - TOPS (performance)
//   - TOPS/W (efficiency)
//   - Chip area
//   - Any chip-level specifications
//
// The values below are ESTIMATES/GUESSES for visualization purposes only.
// They should NOT be presented as facts or used for investment decisions.
//
// VERIFIED claims from Dr. Tour (Nov 2024):
//   - 30 discrete analog states (VERIFIED)
//   - 87% MNIST accuracy with 88% theoretical max (VERIFIED)
//   - "10M× lower energy than NAND" (CLAIMED, NOT VERIFIED)
//   - "1M× faster than NAND" (CLAIMED, NOT VERIFIED)
//   - "80-90% data center energy reduction" (CLAIMED, NOT VERIFIED)
//
// See opensource/papers/08_Documentation/HONESTY_AUDIT.md for full analysis.
func FeCIMChip() *Architecture {
	return &Architecture{
		Name:            "FeCIM CIM",
		Description:     "Ferroelectric compute-in-memory with 30-level cells (ESTIMATED SPECS - TRL4)",
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
	}
}

// CustomArchitecture creates a custom architecture with specified parameters.
func CustomArchitecture(name string, tops, power, area float64) *Architecture {
	return &Architecture{
		Name:        name,
		ProcessNode: 7,
		ChipArea:    area,
		TDP:         power,
		PeakTOPS:    tops,
		TOPSPerWatt: tops / power,
		TOPSPerMM2:  tops / area,
	}
}

// CalculateEfficiency calculates efficiency metrics.
func (a *Architecture) CalculateEfficiency() {
	if a.TDP > 0 {
		a.TOPSPerWatt = a.PeakTOPS / a.TDP
	}
	if a.ChipArea > 0 {
		a.TOPSPerMM2 = a.PeakTOPS / a.ChipArea
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
	energy := a.TDP * totalLatency / 1000 // mJ (convert ms to s)

	// Throughput
	throughput := float64(batchSize) / (totalLatency / 1000) // inferences/sec

	return InferenceResult{
		Architecture: a.Name,
		ModelOps:     modelOps,
		BatchSize:    batchSize,
		Latency:      totalLatency,
		Throughput:   throughput,
		Energy:       energy,
		PowerUsed:    a.TDP,
	}
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
	// 784-128-64-10 network
	// Layer 1: 784*128 = 100352 MACs
	// Layer 2: 128*64 = 8192 MACs
	// Layer 3: 64*10 = 640 MACs
	return Workload{
		Name:        "MNIST",
		Description: "Handwritten digit recognition",
		TotalOps:    109184, // 100352 + 8192 + 640
		Layers:      3,
		Parameters:  109184,
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
	Architecture   string  // Architecture name
	ChipsRequired  int     // Number of chips needed
	TotalPower     float64 // Total power (kW)
	RackSpace      int     // Rack units required
	InferencesPerSec float64 // Total throughput
	CostPerInference float64 // Cost per inference ($)
	TCO            float64 // Total cost of ownership ($/year)
	CO2Emissions   float64 // kg CO2 per day
}

// ScaleToDataCenter calculates data center scale metrics.
func ScaleToDataCenter(arch *Architecture, targetThroughput float64, workload Workload) DataCenterMetrics {
	// Run single chip inference
	result := arch.RunInference(workload.TotalOps, 1)

	// Chips needed for target throughput
	chipsRequired := int(math.Ceil(targetThroughput / result.Throughput))
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
	energyPerInference := result.Energy / 1000 / 3600 // Convert mJ to kWh
	costPerInference := energyPerInference * electricityCost

	// TCO (assume 3 year depreciation, include cooling 1.5x PUE)
	chipCost := arch.ManufactureCost * 10000 // Relative cost
	electricityCostYear := totalPower * 24 * 365 * electricityCost * 1.5
	depreciation := chipCost * float64(chipsRequired) / 3
	tco := electricityCostYear + depreciation

	// CO2 emissions (assume 0.4 kg CO2/kWh)
	co2PerKWh := 0.4
	co2PerDay := totalPower * 24 * co2PerKWh

	return DataCenterMetrics{
		Architecture:     arch.Name,
		ChipsRequired:    chipsRequired,
		TotalPower:       totalPower,
		RackSpace:        rackUnits,
		InferencesPerSec: actualThroughput,
		CostPerInference: costPerInference,
		TCO:              tco,
		CO2Emissions:     co2PerDay,
	}
}

// ComparisonResult contains full comparison between architectures.
type ComparisonResult struct {
	Workload    Workload
	BatchSize   int
	Architectures []*Architecture
	Results     []InferenceResult
	DataCenter  []DataCenterMetrics
}

// CompareArchitectures runs a full comparison.
func CompareArchitectures(workload Workload, batchSize int, targetThroughput float64) ComparisonResult {
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

	return ComparisonResult{
		Workload:      workload,
		BatchSize:     batchSize,
		Architectures: architectures,
		Results:       results,
		DataCenter:    dcMetrics,
	}
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
		EnergyReduction    float64
		LatencyReduction   float64
		AreaReduction      float64
		PowerReduction     float64
		CostReduction      float64
	}
}

// CalculateAdvantages calculates FeCIM advantages.
func CalculateAdvantages(comparison ComparisonResult) FeCIMAdvantage {
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

	// vs CPU
	if cpuArch != nil && fecimArch != nil {
		adv.VsCPU.EnergyReduction = cpuResult.Energy / fecimResult.Energy
		adv.VsCPU.LatencyReduction = cpuResult.Latency / fecimResult.Latency
		adv.VsCPU.ThroughputIncrease = fecimResult.Throughput / cpuResult.Throughput
		adv.VsCPU.PowerReduction = cpuArch.TDP / fecimArch.TDP
		adv.VsCPU.CostReduction = cpuDC.TCO / fecimDC.TCO
	}

	// vs GPU
	if gpuArch != nil && fecimArch != nil {
		adv.VsGPU.EnergyReduction = gpuResult.Energy / fecimResult.Energy
		adv.VsGPU.LatencyReduction = gpuResult.Latency / fecimResult.Latency
		adv.VsGPU.AreaReduction = gpuArch.ChipArea / fecimArch.ChipArea
		adv.VsGPU.PowerReduction = gpuArch.TDP / fecimArch.TDP
		adv.VsGPU.CostReduction = gpuDC.TCO / fecimDC.TCO
	}

	return adv
}
