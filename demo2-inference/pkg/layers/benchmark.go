// Package layers provides neural network layer implementations for CIM inference.
// benchmark.go implements comprehensive benchmarking suite for CIM models.
//
// References:
// - MLPerf Inference benchmark suite
// - NeuroBench framework for neuromorphic computing
// - IBM NorthPole analog AI accelerator benchmarks
// - FTJ memristor crossbar evaluation methodologies

package layers

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

// =============================================================================
// Benchmark Categories - Based on MLPerf and NeuroBench
// =============================================================================

// BenchmarkCategory defines different benchmark domains
type BenchmarkCategory string

const (
	CategoryImageClassification BenchmarkCategory = "image_classification"
	CategoryObjectDetection     BenchmarkCategory = "object_detection"
	CategorySpeechRecognition   BenchmarkCategory = "speech_recognition"
	CategoryNLP                 BenchmarkCategory = "nlp"
	CategoryRecommendation      BenchmarkCategory = "recommendation"
	CategoryTimeSeries          BenchmarkCategory = "time_series"
	CategoryReasoning           BenchmarkCategory = "reasoning"
	CategorySNN                 BenchmarkCategory = "snn"
)

// BenchmarkDataset defines standard datasets for evaluation
type BenchmarkDataset struct {
	Name         string
	Category     BenchmarkCategory
	InputShape   []int
	NumClasses   int
	TrainSize    int
	TestSize     int
	Description  string
}

// StandardDatasets provides canonical benchmark datasets
var StandardDatasets = map[string]*BenchmarkDataset{
	"mnist": {
		Name:        "MNIST",
		Category:    CategoryImageClassification,
		InputShape:  []int{1, 28, 28},
		NumClasses:  10,
		TrainSize:   60000,
		TestSize:    10000,
		Description: "Handwritten digit classification",
	},
	"cifar10": {
		Name:        "CIFAR-10",
		Category:    CategoryImageClassification,
		InputShape:  []int{3, 32, 32},
		NumClasses:  10,
		TrainSize:   50000,
		TestSize:    10000,
		Description: "Natural image classification (10 classes)",
	},
	"cifar100": {
		Name:        "CIFAR-100",
		Category:    CategoryImageClassification,
		InputShape:  []int{3, 32, 32},
		NumClasses:  100,
		TrainSize:   50000,
		TestSize:    10000,
		Description: "Fine-grained image classification",
	},
	"imagenet": {
		Name:        "ImageNet-1K",
		Category:    CategoryImageClassification,
		InputShape:  []int{3, 224, 224},
		NumClasses:  1000,
		TrainSize:   1281167,
		TestSize:    50000,
		Description: "Large-scale image classification",
	},
	"speech_commands": {
		Name:        "Speech Commands",
		Category:    CategorySpeechRecognition,
		InputShape:  []int{1, 16000},
		NumClasses:  35,
		TrainSize:   84843,
		TestSize:    11005,
		Description: "Keyword spotting (35 commands)",
	},
	"shd": {
		Name:        "Spiking Heidelberg Digits",
		Category:    CategorySNN,
		InputShape:  []int{700, 100},
		NumClasses:  20,
		TrainSize:   8156,
		TestSize:    2264,
		Description: "Spike-encoded spoken digits",
	},
	"dvs_gesture": {
		Name:        "DVS128 Gesture",
		Category:    CategorySNN,
		InputShape:  []int{2, 128, 128},
		NumClasses:  11,
		TrainSize:   1077,
		TestSize:    264,
		Description: "Event camera gesture recognition",
	},
}

// =============================================================================
// Benchmark Metrics
// =============================================================================

// AccuracyMetrics holds classification accuracy results
type AccuracyMetrics struct {
	TopK           map[int]float64 // Top-K accuracy (k -> accuracy)
	Precision      float64
	Recall         float64
	F1Score        float64
	ConfusionMatrix [][]int
}

// LatencyMetrics holds timing measurements
type LatencyMetrics struct {
	MeanLatencyMs   float64
	P50LatencyMs    float64
	P90LatencyMs    float64
	P99LatencyMs    float64
	MaxLatencyMs    float64
	MinLatencyMs    float64
	StdDevMs        float64
	ThroughputIPS   float64 // Inferences per second
}

// EnergyMetrics holds power and energy measurements
type EnergyMetrics struct {
	TotalEnergyJ      float64
	EnergyPerInfJ     float64
	EnergyPerMACfJ    float64
	AveragePowerW     float64
	PeakPowerW        float64
	// CIM-specific metrics
	ArrayReadEnergy   float64
	ADCEnergy         float64
	DACEnergy         float64
	DigitalEnergy     float64
}

// AreaMetrics holds silicon area estimates
type AreaMetrics struct {
	TotalAreaMM2      float64
	ArrayAreaMM2      float64
	PeripheralAreaMM2 float64
	DigitalAreaMM2    float64
	MemoryAreaMM2     float64
}

// EfficiencyMetrics holds derived efficiency numbers
type EfficiencyMetrics struct {
	TOPSMM2           float64 // TOPS per mm²
	TOPSW             float64 // TOPS per Watt
	AccuracyPerTOP    float64 // Accuracy achieved per TOP
	InferencesPerJ    float64 // Inferences per Joule
	// Comparison ratios vs GPU baseline
	SpeedupVsGPU      float64
	EnergyReduction   float64
}

// HardwareConstraints for CIM deployment
type HardwareConstraints struct {
	MaxArraySize      int
	MaxArrayCount     int
	ADCBits           int
	DACBits           int
	WeightBits        int
	NoiseLevel        float64
	StuckAtFaultRate  float64
	VariationSigma    float64
}

// BenchmarkResult aggregates all metrics for a benchmark run
type BenchmarkResult struct {
	// Metadata
	ModelName         string
	DatasetName       string
	Timestamp         time.Time
	HardwareConfig    HardwareConstraints

	// Results
	Accuracy          AccuracyMetrics
	Latency           LatencyMetrics
	Energy            EnergyMetrics
	Area              AreaMetrics
	Efficiency        EfficiencyMetrics

	// CIM-specific
	ArrayUtilization  float64
	WeightMappingEff  float64
	QuantizationLoss  float64

	// Raw data
	PredictedLabels   []int
	GroundTruthLabels []int
	InferenceTimesMs  []float64
}

// =============================================================================
// CIM Benchmark Suite
// =============================================================================

// CIMBenchmarkSuite orchestrates comprehensive benchmarking
type CIMBenchmarkSuite struct {
	Config            BenchmarkConfig
	Results           map[string]*BenchmarkResult
	BaselineResults   map[string]*BenchmarkResult // GPU/CPU baselines
}

// BenchmarkConfig configures the benchmark suite
type BenchmarkConfig struct {
	// Hardware settings
	Hardware          HardwareConstraints
	TechnologyNode    string // "28nm", "14nm", "7nm", etc.

	// Benchmark settings
	NumWarmupRuns     int
	NumBenchmarkRuns  int
	BatchSize         int

	// Evaluation settings
	ComputeEnergy     bool
	ComputeArea       bool
	CompareToGPU      bool

	// Noise injection for realistic simulation
	EnableNoise       bool
	EnableVariation   bool
	EnableFaults      bool
}

// NewCIMBenchmarkSuite creates a new benchmark suite
func NewCIMBenchmarkSuite(config BenchmarkConfig) *CIMBenchmarkSuite {
	if config.NumWarmupRuns == 0 {
		config.NumWarmupRuns = 10
	}
	if config.NumBenchmarkRuns == 0 {
		config.NumBenchmarkRuns = 100
	}
	if config.BatchSize == 0 {
		config.BatchSize = 1
	}

	return &CIMBenchmarkSuite{
		Config:          config,
		Results:         make(map[string]*BenchmarkResult),
		BaselineResults: make(map[string]*BenchmarkResult),
	}
}

// =============================================================================
// Model Definitions for Benchmarking
// =============================================================================

// BenchmarkModel defines a model architecture for testing
type BenchmarkModel struct {
	Name          string
	Layers        []LayerSpec
	TotalMACs     int64
	TotalParams   int64
	TargetDataset string
}

// LayerSpec defines a single layer
type LayerSpec struct {
	Type          string // "conv", "fc", "attention", etc.
	InputShape    []int
	OutputShape   []int
	KernelSize    []int // For conv layers
	NumHeads      int   // For attention
	MACs          int64
	Params        int64
}

// StandardModels provides canonical benchmark models
var StandardModels = map[string]*BenchmarkModel{
	"lenet5": {
		Name:          "LeNet-5",
		TargetDataset: "mnist",
		TotalMACs:     416000,
		TotalParams:   61706,
		Layers: []LayerSpec{
			{Type: "conv", InputShape: []int{1, 28, 28}, OutputShape: []int{6, 24, 24}, KernelSize: []int{5, 5}},
			{Type: "pool", InputShape: []int{6, 24, 24}, OutputShape: []int{6, 12, 12}},
			{Type: "conv", InputShape: []int{6, 12, 12}, OutputShape: []int{16, 8, 8}, KernelSize: []int{5, 5}},
			{Type: "pool", InputShape: []int{16, 8, 8}, OutputShape: []int{16, 4, 4}},
			{Type: "fc", InputShape: []int{256}, OutputShape: []int{120}},
			{Type: "fc", InputShape: []int{120}, OutputShape: []int{84}},
			{Type: "fc", InputShape: []int{84}, OutputShape: []int{10}},
		},
	},
	"mlp_mnist": {
		Name:          "MLP-MNIST",
		TargetDataset: "mnist",
		TotalMACs:     669706,
		TotalParams:   669706,
		Layers: []LayerSpec{
			{Type: "fc", InputShape: []int{784}, OutputShape: []int{512}},
			{Type: "fc", InputShape: []int{512}, OutputShape: []int{256}},
			{Type: "fc", InputShape: []int{256}, OutputShape: []int{10}},
		},
	},
	"vgg8": {
		Name:          "VGG-8",
		TargetDataset: "cifar10",
		TotalMACs:     87000000,
		TotalParams:   7000000,
		Layers: []LayerSpec{
			{Type: "conv", InputShape: []int{3, 32, 32}, OutputShape: []int{64, 32, 32}, KernelSize: []int{3, 3}},
			{Type: "conv", InputShape: []int{64, 32, 32}, OutputShape: []int{64, 32, 32}, KernelSize: []int{3, 3}},
			{Type: "pool", InputShape: []int{64, 32, 32}, OutputShape: []int{64, 16, 16}},
			{Type: "conv", InputShape: []int{64, 16, 16}, OutputShape: []int{128, 16, 16}, KernelSize: []int{3, 3}},
			{Type: "conv", InputShape: []int{128, 16, 16}, OutputShape: []int{128, 16, 16}, KernelSize: []int{3, 3}},
			{Type: "pool", InputShape: []int{128, 16, 16}, OutputShape: []int{128, 8, 8}},
			{Type: "fc", InputShape: []int{8192}, OutputShape: []int{1024}},
			{Type: "fc", InputShape: []int{1024}, OutputShape: []int{10}},
		},
	},
	"resnet18": {
		Name:          "ResNet-18",
		TargetDataset: "imagenet",
		TotalMACs:     1800000000,
		TotalParams:   11700000,
	},
	"bert_tiny": {
		Name:          "BERT-Tiny",
		TargetDataset: "nlp",
		TotalMACs:     200000000,
		TotalParams:   4400000,
	},
	"snn_shd": {
		Name:          "SNN-SHD",
		TargetDataset: "shd",
		TotalMACs:     5000000,
		TotalParams:   50000,
		Layers: []LayerSpec{
			{Type: "lif", InputShape: []int{700}, OutputShape: []int{256}},
			{Type: "lif", InputShape: []int{256}, OutputShape: []int{256}},
			{Type: "lif", InputShape: []int{256}, OutputShape: []int{20}},
		},
	},
}

// =============================================================================
// Benchmark Execution
// =============================================================================

// RunBenchmark executes a full benchmark for a model/dataset pair
func (suite *CIMBenchmarkSuite) RunBenchmark(modelName, datasetName string) (*BenchmarkResult, error) {
	model, ok := StandardModels[modelName]
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", modelName)
	}

	dataset, ok := StandardDatasets[datasetName]
	if !ok {
		return nil, fmt.Errorf("unknown dataset: %s", datasetName)
	}

	result := &BenchmarkResult{
		ModelName:      model.Name,
		DatasetName:    dataset.Name,
		Timestamp:      time.Now(),
		HardwareConfig: suite.Config.Hardware,
	}

	// Simulate inference and collect metrics
	result.Latency = suite.benchmarkLatency(model, dataset)
	result.Accuracy = suite.benchmarkAccuracy(model, dataset)

	if suite.Config.ComputeEnergy {
		result.Energy = suite.benchmarkEnergy(model, dataset)
	}

	if suite.Config.ComputeArea {
		result.Area = suite.benchmarkArea(model)
	}

	// Compute efficiency metrics
	result.Efficiency = suite.computeEfficiency(model, result)

	// CIM-specific metrics
	result.ArrayUtilization = suite.computeArrayUtilization(model)
	result.WeightMappingEff = suite.computeMappingEfficiency(model)
	result.QuantizationLoss = suite.estimateQuantizationLoss(model)

	// Store result
	key := fmt.Sprintf("%s_%s", modelName, datasetName)
	suite.Results[key] = result

	return result, nil
}

// benchmarkLatency measures inference timing
func (suite *CIMBenchmarkSuite) benchmarkLatency(model *BenchmarkModel, dataset *BenchmarkDataset) LatencyMetrics {
	times := make([]float64, suite.Config.NumBenchmarkRuns)

	// Simulate inference timing based on model complexity
	// Real implementation would run actual inference
	baseLatencyUs := float64(model.TotalMACs) / 1e9 * 1000 // 1 TOPS baseline

	// CIM acceleration factor (typically 10-100x vs digital)
	cimSpeedup := 20.0
	adjustedLatencyUs := baseLatencyUs / cimSpeedup

	for i := range times {
		// Add realistic variation (±10%)
		variation := 1.0 + (float64(i%20)-10)/100.0
		times[i] = adjustedLatencyUs * variation / 1000.0 // Convert to ms
	}

	sort.Float64s(times)

	// Calculate statistics
	var sum, sumSq float64
	for _, t := range times {
		sum += t
		sumSq += t * t
	}
	mean := sum / float64(len(times))
	variance := sumSq/float64(len(times)) - mean*mean

	n := len(times)
	return LatencyMetrics{
		MeanLatencyMs: mean,
		P50LatencyMs:  times[n/2],
		P90LatencyMs:  times[n*90/100],
		P99LatencyMs:  times[n*99/100],
		MaxLatencyMs:  times[n-1],
		MinLatencyMs:  times[0],
		StdDevMs:      math.Sqrt(variance),
		ThroughputIPS: 1000.0 / mean,
	}
}

// benchmarkAccuracy evaluates classification accuracy
func (suite *CIMBenchmarkSuite) benchmarkAccuracy(model *BenchmarkModel, dataset *BenchmarkDataset) AccuracyMetrics {
	// Simulated accuracy based on model/dataset and hardware constraints
	baselineAccuracy := getBaselineAccuracy(model.Name, dataset.Name)

	// Accuracy degradation from hardware constraints
	degradation := 0.0

	// Quantization loss (lower bits = more loss)
	if suite.Config.Hardware.WeightBits < 8 {
		degradation += float64(8-suite.Config.Hardware.WeightBits) * 0.5
	}

	// Noise impact
	if suite.Config.EnableNoise {
		degradation += suite.Config.Hardware.NoiseLevel * 100 * 0.3
	}

	// Fault impact (stuck-at faults)
	if suite.Config.EnableFaults {
		degradation += suite.Config.Hardware.StuckAtFaultRate * 100 * 2.0
	}

	// Variation impact
	if suite.Config.EnableVariation {
		degradation += suite.Config.Hardware.VariationSigma * 100 * 0.5
	}

	actualAccuracy := math.Max(0, baselineAccuracy-degradation)

	return AccuracyMetrics{
		TopK: map[int]float64{
			1: actualAccuracy,
			5: math.Min(100, actualAccuracy+4),
		},
		Precision: actualAccuracy / 100,
		Recall:    actualAccuracy / 100,
		F1Score:   actualAccuracy / 100,
	}
}

// getBaselineAccuracy returns expected accuracy for model/dataset pair
func getBaselineAccuracy(modelName, datasetName string) float64 {
	baselines := map[string]map[string]float64{
		"lenet5":    {"mnist": 99.2},
		"mlp_mnist": {"mnist": 98.0},
		"vgg8":      {"cifar10": 91.5, "cifar100": 68.0},
		"resnet18":  {"imagenet": 69.7, "cifar10": 95.0, "cifar100": 77.0},
		"bert_tiny": {"nlp": 82.0},
		"snn_shd":   {"shd": 92.5, "dvs_gesture": 95.0},
	}

	if modelBaselines, ok := baselines[modelName]; ok {
		if acc, ok := modelBaselines[datasetName]; ok {
			return acc
		}
	}
	return 85.0 // Default fallback
}

// benchmarkEnergy estimates energy consumption
func (suite *CIMBenchmarkSuite) benchmarkEnergy(model *BenchmarkModel, dataset *BenchmarkDataset) EnergyMetrics {
	// Energy model based on technology node and operations
	var energyPerMACfJ float64
	switch suite.Config.TechnologyNode {
	case "28nm":
		energyPerMACfJ = 200
	case "14nm":
		energyPerMACfJ = 80
	case "7nm":
		energyPerMACfJ = 30
	case "5nm":
		energyPerMACfJ = 15
	default:
		energyPerMACfJ = 100
	}

	// CIM reduces MAC energy significantly
	cimEnergyReduction := 0.1 // 10x reduction
	effectiveEnergyPerMAC := energyPerMACfJ * cimEnergyReduction

	totalMACEnergy := float64(model.TotalMACs) * effectiveEnergyPerMAC * 1e-15

	// Peripheral energy (ADC, DAC, digital logic)
	// ADC energy scales with bits and array reads
	numArrayReads := model.TotalMACs / int64(suite.Config.Hardware.MaxArraySize*suite.Config.Hardware.MaxArraySize)
	if numArrayReads == 0 {
		numArrayReads = 1
	}

	adcEnergyPerRead := math.Pow(2, float64(suite.Config.Hardware.ADCBits)) * 0.1e-12 // ~100 fJ per level
	totalADCEnergy := float64(numArrayReads) * adcEnergyPerRead * float64(suite.Config.Hardware.MaxArraySize)

	dacEnergy := float64(numArrayReads) * 50e-15 * float64(suite.Config.Hardware.MaxArraySize)
	digitalEnergy := totalMACEnergy * 0.2 // 20% overhead for control logic

	totalEnergy := totalMACEnergy + totalADCEnergy + dacEnergy + digitalEnergy

	// Power from latency
	latency := suite.benchmarkLatency(model, dataset)
	avgPower := totalEnergy / (latency.MeanLatencyMs * 1e-3)

	return EnergyMetrics{
		TotalEnergyJ:    totalEnergy,
		EnergyPerInfJ:   totalEnergy,
		EnergyPerMACfJ:  effectiveEnergyPerMAC,
		AveragePowerW:   avgPower,
		PeakPowerW:      avgPower * 1.5,
		ArrayReadEnergy: totalMACEnergy,
		ADCEnergy:       totalADCEnergy,
		DACEnergy:       dacEnergy,
		DigitalEnergy:   digitalEnergy,
	}
}

// benchmarkArea estimates silicon area
func (suite *CIMBenchmarkSuite) benchmarkArea(model *BenchmarkModel) AreaMetrics {
	// Array area based on technology node
	var cellArea float64 // µm²
	switch suite.Config.TechnologyNode {
	case "28nm":
		cellArea = 0.1
	case "14nm":
		cellArea = 0.04
	case "7nm":
		cellArea = 0.015
	case "5nm":
		cellArea = 0.008
	default:
		cellArea = 0.05
	}

	// Number of cells needed
	numCells := model.TotalParams
	arrayArea := float64(numCells) * cellArea * 1e-6 // Convert to mm²

	// Peripheral overhead (typically 30-50% of array)
	peripheralArea := arrayArea * 0.4

	// Digital logic
	digitalArea := arrayArea * 0.1

	return AreaMetrics{
		TotalAreaMM2:      arrayArea + peripheralArea + digitalArea,
		ArrayAreaMM2:      arrayArea,
		PeripheralAreaMM2: peripheralArea,
		DigitalAreaMM2:    digitalArea,
		MemoryAreaMM2:     arrayArea * 0.1, // Buffer memory
	}
}

// computeEfficiency derives efficiency metrics
func (suite *CIMBenchmarkSuite) computeEfficiency(model *BenchmarkModel, result *BenchmarkResult) EfficiencyMetrics {
	// TOPS calculation
	tops := float64(model.TotalMACs) * result.Latency.ThroughputIPS * 2 / 1e12

	eff := EfficiencyMetrics{}

	if result.Area.TotalAreaMM2 > 0 {
		eff.TOPSMM2 = tops / result.Area.TotalAreaMM2
	}

	if result.Energy.AveragePowerW > 0 {
		eff.TOPSW = tops / result.Energy.AveragePowerW
	}

	if tops > 0 {
		eff.AccuracyPerTOP = result.Accuracy.TopK[1] / tops
	}

	if result.Energy.EnergyPerInfJ > 0 {
		eff.InferencesPerJ = 1.0 / result.Energy.EnergyPerInfJ
	}

	// GPU comparison (A100 baseline: ~300 TFLOPS, 400W, 826 mm²)
	gpuTOPSW := 0.75   // 300 TFLOPS / 400W
	gpuTOPSMM2 := 0.36 // 300 TFLOPS / 826 mm²

	if eff.TOPSW > 0 {
		eff.EnergyReduction = eff.TOPSW / gpuTOPSW
	}
	if eff.TOPSMM2 > 0 {
		eff.SpeedupVsGPU = eff.TOPSMM2 / gpuTOPSMM2
	}

	return eff
}

// computeArrayUtilization calculates how well arrays are utilized
func (suite *CIMBenchmarkSuite) computeArrayUtilization(model *BenchmarkModel) float64 {
	arraySize := suite.Config.Hardware.MaxArraySize
	totalCells := arraySize * arraySize

	// Check if model weights fit well into arrays
	paramsPerArray := int64(totalCells)
	numArraysNeeded := (model.TotalParams + paramsPerArray - 1) / paramsPerArray
	actualParamsStored := numArraysNeeded * paramsPerArray

	return float64(model.TotalParams) / float64(actualParamsStored) * 100
}

// computeMappingEfficiency measures weight mapping quality
func (suite *CIMBenchmarkSuite) computeMappingEfficiency(model *BenchmarkModel) float64 {
	// Simplified: check if layer dimensions match array dimensions
	arraySize := suite.Config.Hardware.MaxArraySize

	var totalScore float64
	for _, layer := range model.Layers {
		if layer.Type == "fc" && len(layer.InputShape) > 0 && len(layer.OutputShape) > 0 {
			// FC layers: input x output should fit arrays
			inSize := layer.InputShape[0]
			outSize := layer.OutputShape[0]

			rowUtil := math.Min(float64(inSize), float64(arraySize)) / float64(arraySize)
			colUtil := math.Min(float64(outSize), float64(arraySize)) / float64(arraySize)

			totalScore += (rowUtil + colUtil) / 2
		} else {
			totalScore += 0.7 // Default for other layer types
		}
	}

	if len(model.Layers) > 0 {
		return totalScore / float64(len(model.Layers)) * 100
	}
	return 80.0
}

// estimateQuantizationLoss predicts accuracy loss from quantization
func (suite *CIMBenchmarkSuite) estimateQuantizationLoss(model *BenchmarkModel) float64 {
	bits := suite.Config.Hardware.WeightBits

	// Empirical quantization loss model
	// 8-bit: ~0%, 6-bit: ~0.5%, 4-bit: ~2%, 2-bit: ~10%
	switch {
	case bits >= 8:
		return 0.0
	case bits >= 6:
		return 0.5
	case bits >= 4:
		return 2.0
	case bits >= 2:
		return 10.0
	default:
		return 25.0
	}
}

// =============================================================================
// Comparison and Reporting
// =============================================================================

// CompareModels compares multiple models on the same benchmark
func (suite *CIMBenchmarkSuite) CompareModels(modelNames []string, datasetName string) map[string]*BenchmarkResult {
	results := make(map[string]*BenchmarkResult)

	for _, modelName := range modelNames {
		result, err := suite.RunBenchmark(modelName, datasetName)
		if err == nil {
			results[modelName] = result
		}
	}

	return results
}

// CompareTechnologyNodes compares performance across technology nodes
func (suite *CIMBenchmarkSuite) CompareTechnologyNodes(modelName, datasetName string, nodes []string) map[string]*BenchmarkResult {
	results := make(map[string]*BenchmarkResult)
	originalNode := suite.Config.TechnologyNode

	for _, node := range nodes {
		suite.Config.TechnologyNode = node
		result, err := suite.RunBenchmark(modelName, datasetName)
		if err == nil {
			results[node] = result
		}
	}

	suite.Config.TechnologyNode = originalNode
	return results
}

// GenerateReport creates a benchmark report
func (suite *CIMBenchmarkSuite) GenerateReport() string {
	report := "# CIM Benchmark Report\n\n"
	report += fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339))
	report += fmt.Sprintf("Technology Node: %s\n", suite.Config.TechnologyNode)
	report += fmt.Sprintf("Array Size: %dx%d\n", suite.Config.Hardware.MaxArraySize, suite.Config.Hardware.MaxArraySize)
	report += fmt.Sprintf("Weight Bits: %d\n\n", suite.Config.Hardware.WeightBits)

	report += "## Results Summary\n\n"
	report += "| Model | Dataset | Accuracy | Latency (ms) | Energy (µJ) | TOPS/W |\n"
	report += "|-------|---------|----------|--------------|-------------|--------|\n"

	for key, result := range suite.Results {
		report += fmt.Sprintf("| %s | %s | %.2f%% | %.3f | %.3f | %.1f |\n",
			result.ModelName,
			result.DatasetName,
			result.Accuracy.TopK[1],
			result.Latency.MeanLatencyMs,
			result.Energy.TotalEnergyJ*1e6,
			result.Efficiency.TOPSW)
		_ = key
	}

	return report
}

// ExportResults exports results to JSON
func (suite *CIMBenchmarkSuite) ExportResults() ([]byte, error) {
	return json.MarshalIndent(suite.Results, "", "  ")
}

// =============================================================================
// FTJ-Specific Benchmarks
// =============================================================================

// FTJBenchmarkConfig configures FTJ-specific parameters
type FTJBenchmarkConfig struct {
	// Device parameters
	OnOffRatio        float64 // Typical: 100-1000
	NumStates         int     // Multi-level states (2-256)
	SwitchingSpeedNs  float64 // Typical: 10-100 ns
	EnduranceCycles   float64 // Typical: 10^8 - 10^12
	RetentionSeconds  float64 // Typical: 10^5 s

	// Self-rectifying properties
	RectifyingRatio   float64 // For sneak path suppression

	// Variability
	CycleVariation    float64 // Percentage
	DeviceVariation   float64 // Percentage
}

// FTJBenchmark evaluates FTJ-based CIM systems
type FTJBenchmark struct {
	Config     FTJBenchmarkConfig
	BaseConfig BenchmarkConfig
}

// NewFTJBenchmark creates an FTJ benchmark instance
func NewFTJBenchmark(ftjConfig FTJBenchmarkConfig, baseConfig BenchmarkConfig) *FTJBenchmark {
	return &FTJBenchmark{
		Config:     ftjConfig,
		BaseConfig: baseConfig,
	}
}

// EvaluateCrossbarSize determines maximum crossbar size for given read margin
func (b *FTJBenchmark) EvaluateCrossbarSize(targetReadMargin float64) int {
	// Based on self-rectifying ratio, calculate max array size
	// Higher rectifying ratio enables larger arrays

	// Sneak path analysis: N = sqrt(RR * RM / (1 - RM))
	// where RR = rectifying ratio, RM = read margin
	rr := b.Config.RectifyingRatio
	rm := targetReadMargin

	if rm >= 1.0 || rm <= 0 {
		return 64 // Default
	}

	maxSize := int(math.Sqrt(rr * rm / (1 - rm)))

	// Clamp to reasonable range
	if maxSize < 16 {
		maxSize = 16
	}
	if maxSize > 4096 {
		maxSize = 4096
	}

	return maxSize
}

// EvaluateSynapticPerformance benchmarks FTJ as artificial synapse
func (b *FTJBenchmark) EvaluateSynapticPerformance() map[string]float64 {
	results := make(map[string]float64)

	// Linearity (based on number of states)
	// More states = better linearity for analog computing
	maxLinearBits := math.Log2(float64(b.Config.NumStates))
	results["effective_bits"] = maxLinearBits

	// Symmetry (estimated from cycle variation)
	symmetry := math.Max(0, 100-b.Config.CycleVariation*10)
	results["weight_symmetry"] = symmetry

	// Retention reliability
	retentionScore := math.Min(100, math.Log10(b.Config.RetentionSeconds)*20)
	results["retention_score"] = retentionScore

	// Endurance score
	enduranceScore := math.Min(100, math.Log10(b.Config.EnduranceCycles)*10)
	results["endurance_score"] = enduranceScore

	// Speed score (lower is better, normalized)
	speedScore := math.Max(0, 100-b.Config.SwitchingSpeedNs)
	results["speed_score"] = speedScore

	// Overall synapse quality
	results["synapse_quality"] = (symmetry + retentionScore + enduranceScore + speedScore) / 4

	return results
}

// =============================================================================
// NC-FET Benchmarks
// =============================================================================

// NCFETConfig configures negative capacitance FET parameters
type NCFETConfig struct {
	// Subthreshold characteristics
	SubthresholdSwingMV float64 // mV/decade (< 60 = NC effect)
	OnCurrentUA         float64 // On-state current per µm
	OffCurrentPA        float64 // Off-state leakage per µm

	// Ferroelectric layer
	FerroelectricMaterial string // "HZO", "SBT", etc.
	FerroelectricThickNm  float64

	// Gate stack
	GateLengthNm        float64
	EOTNm               float64 // Equivalent oxide thickness

	// Hysteresis and reliability
	HysteresisWindowMV  float64
	EnduranceCycles     float64
}

// NCFETBenchmark evaluates NC-FET performance
type NCFETBenchmark struct {
	Config NCFETConfig
}

// NewNCFETBenchmark creates an NC-FET benchmark instance
func NewNCFETBenchmark(config NCFETConfig) *NCFETBenchmark {
	return &NCFETBenchmark{Config: config}
}

// EvaluateSwitchingEnergy estimates switching energy advantage
func (b *NCFETBenchmark) EvaluateSwitchingEnergy() map[string]float64 {
	results := make(map[string]float64)

	// Standard MOSFET SS = 60 mV/dec at room temp
	standardSS := 60.0
	ncSS := b.Config.SubthresholdSwingMV

	// Voltage scaling advantage
	if ncSS > 0 && ncSS < standardSS {
		voltageReduction := ncSS / standardSS
		// Energy scales with V^2
		energyReduction := voltageReduction * voltageReduction
		results["energy_reduction_factor"] = 1.0 / energyReduction
		results["voltage_scaling"] = voltageReduction
	} else {
		results["energy_reduction_factor"] = 1.0
		results["voltage_scaling"] = 1.0
	}

	// On/off ratio improvement
	if b.Config.OffCurrentPA > 0 {
		onOffRatio := (b.Config.OnCurrentUA * 1e6) / b.Config.OffCurrentPA
		results["on_off_ratio"] = onOffRatio
		results["on_off_decades"] = math.Log10(onOffRatio)
	}

	// Leakage reduction
	// Steeper SS means less leakage at same operating point
	leakageReduction := math.Pow(10, (standardSS-ncSS)/standardSS*2)
	results["leakage_reduction"] = leakageReduction

	return results
}

// EvaluateCIMIntegration assesses suitability for CIM
func (b *NCFETBenchmark) EvaluateCIMIntegration() map[string]float64 {
	results := make(map[string]float64)

	// Area efficiency (smaller gate = denser arrays)
	areaScore := 100 * (7.0 / b.Config.GateLengthNm) // Normalized to 7nm
	if areaScore > 100 {
		areaScore = 100
	}
	results["area_score"] = areaScore

	// Endurance for weight updates
	enduranceScore := math.Min(100, math.Log10(b.Config.EnduranceCycles)*10)
	results["endurance_score"] = enduranceScore

	// Hysteresis penalty (affects precision)
	if b.Config.HysteresisWindowMV > 0 {
		hysteresisPenalty := math.Min(50, b.Config.HysteresisWindowMV/10)
		results["hysteresis_penalty"] = hysteresisPenalty
		results["precision_score"] = 100 - hysteresisPenalty
	} else {
		results["precision_score"] = 100
	}

	// Overall CIM suitability
	results["cim_suitability"] = (areaScore + enduranceScore + results["precision_score"]) / 3

	return results
}

// =============================================================================
// NeuroBench-Style Benchmarks
// =============================================================================

// NeuroBenchMetrics for neuromorphic systems
type NeuroBenchMetrics struct {
	// Algorithmic metrics
	TaskAccuracy       float64
	SpikingEfficiency  float64 // Sparsity of spike activity
	TemporalPrecision  float64 // For time-based coding

	// System metrics
	SynapticOpsPerSec  float64
	EnergyPerSynOp     float64
	Latency            float64

	// Comparison to SoTA
	AccuracyVsSoTA     float64
	EfficiencyVsSoTA   float64
}

// NeuroBenchSuite for neuromorphic algorithm evaluation
type NeuroBenchSuite struct {
	Config BenchmarkConfig
}

// RunNeuroBench executes neuromorphic benchmarks
func (suite *NeuroBenchSuite) RunNeuroBench(modelName, datasetName string) (*NeuroBenchMetrics, error) {
	model, ok := StandardModels[modelName]
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", modelName)
	}

	dataset, ok := StandardDatasets[datasetName]
	if !ok {
		return nil, fmt.Errorf("unknown dataset: %s", datasetName)
	}

	metrics := &NeuroBenchMetrics{}

	// Task accuracy (use baseline or SNN-specific)
	metrics.TaskAccuracy = getBaselineAccuracy(modelName, datasetName)

	// Spiking efficiency (sparsity)
	// For SNN models, estimate spike sparsity
	if dataset.Category == CategorySNN {
		metrics.SpikingEfficiency = 95.0 // Typical 5% activity
	} else {
		metrics.SpikingEfficiency = 70.0 // ANN-to-SNN conversion
	}

	// Temporal precision
	metrics.TemporalPrecision = 90.0 // Placeholder

	// Synaptic operations
	synOpsPerInference := model.TotalMACs * 2 // Approximate
	latencyS := 0.001                         // 1ms inference time
	metrics.SynapticOpsPerSec = float64(synOpsPerInference) / latencyS
	metrics.Latency = latencyS * 1000 // ms

	// Energy (FeFET-based neuromorphic)
	energyPerSynOp := 10e-15 // 10 fJ per synapse
	metrics.EnergyPerSynOp = energyPerSynOp

	// Comparison to state-of-the-art
	// SHD SoTA: ~95%, DVS Gesture SoTA: ~98%
	sotaAccuracy := map[string]float64{
		"shd":         95.0,
		"dvs_gesture": 98.0,
	}
	if sota, ok := sotaAccuracy[datasetName]; ok {
		metrics.AccuracyVsSoTA = metrics.TaskAccuracy / sota * 100
	} else {
		metrics.AccuracyVsSoTA = 90.0
	}

	// Efficiency comparison (vs Loihi: ~20 pJ/SOP)
	loihiEnergy := 20e-12
	metrics.EfficiencyVsSoTA = loihiEnergy / metrics.EnergyPerSynOp

	return metrics, nil
}

// =============================================================================
// Utility Functions
// =============================================================================

// PrintBenchmarkResult formats a result for display
func PrintBenchmarkResult(result *BenchmarkResult) string {
	output := fmt.Sprintf("Benchmark: %s on %s\n", result.ModelName, result.DatasetName)
	output += fmt.Sprintf("Timestamp: %s\n\n", result.Timestamp.Format(time.RFC3339))

	output += "Accuracy:\n"
	output += fmt.Sprintf("  Top-1: %.2f%%\n", result.Accuracy.TopK[1])
	output += fmt.Sprintf("  Top-5: %.2f%%\n", result.Accuracy.TopK[5])

	output += "\nLatency:\n"
	output += fmt.Sprintf("  Mean: %.3f ms\n", result.Latency.MeanLatencyMs)
	output += fmt.Sprintf("  P99:  %.3f ms\n", result.Latency.P99LatencyMs)
	output += fmt.Sprintf("  Throughput: %.1f IPS\n", result.Latency.ThroughputIPS)

	if result.Energy.TotalEnergyJ > 0 {
		output += "\nEnergy:\n"
		output += fmt.Sprintf("  Per Inference: %.3f µJ\n", result.Energy.EnergyPerInfJ*1e6)
		output += fmt.Sprintf("  Per MAC: %.1f fJ\n", result.Energy.EnergyPerMACfJ)
		output += fmt.Sprintf("  Average Power: %.3f mW\n", result.Energy.AveragePowerW*1000)
	}

	if result.Area.TotalAreaMM2 > 0 {
		output += "\nArea:\n"
		output += fmt.Sprintf("  Total: %.4f mm²\n", result.Area.TotalAreaMM2)
		output += fmt.Sprintf("  Array: %.4f mm²\n", result.Area.ArrayAreaMM2)
	}

	output += "\nEfficiency:\n"
	output += fmt.Sprintf("  TOPS/W: %.1f\n", result.Efficiency.TOPSW)
	output += fmt.Sprintf("  TOPS/mm²: %.2f\n", result.Efficiency.TOPSMM2)
	output += fmt.Sprintf("  vs GPU Speedup: %.1fx\n", result.Efficiency.SpeedupVsGPU)
	output += fmt.Sprintf("  vs GPU Energy: %.1fx better\n", result.Efficiency.EnergyReduction)

	output += "\nCIM Metrics:\n"
	output += fmt.Sprintf("  Array Utilization: %.1f%%\n", result.ArrayUtilization)
	output += fmt.Sprintf("  Mapping Efficiency: %.1f%%\n", result.WeightMappingEff)
	output += fmt.Sprintf("  Quantization Loss: %.1f%%\n", result.QuantizationLoss)

	return output
}
