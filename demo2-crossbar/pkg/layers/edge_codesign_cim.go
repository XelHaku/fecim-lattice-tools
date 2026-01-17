// Package layers provides edge deployment and hardware-software co-design for CIM systems.
// This module implements edge deployment case studies, CIM compiler frameworks,
// non-ideality aware training, and dataflow scheduling based on research from
// CMSwitch, PIMCOMP, CIM-MLC, HASTILY, and NORA frameworks.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// =============================================================================
// EDGE DEPLOYMENT CASE STUDIES
// =============================================================================

// EdgeDeviceType represents different edge deployment targets
type EdgeDeviceType string

const (
	EdgeDeviceIoTSensor   EdgeDeviceType = "iot_sensor"
	EdgeDeviceWearable    EdgeDeviceType = "wearable"
	EdgeDeviceAutomotive  EdgeDeviceType = "automotive"
	EdgeDeviceSmartHome   EdgeDeviceType = "smart_home"
	EdgeDeviceIndustrial  EdgeDeviceType = "industrial_iot"
	EdgeDeviceBiomedical  EdgeDeviceType = "biomedical"
)

// EdgeDeploymentConfig configures edge device constraints
type EdgeDeploymentConfig struct {
	DeviceType        EdgeDeviceType
	PowerBudgetMW     float64 // Milliwatts
	LatencyBudgetMs   float64 // Milliseconds
	MemoryBudgetKB    float64 // Kilobytes
	AccuracyThreshold float64 // Minimum acceptable accuracy
	BatteryCapacityMWh float64 // Battery capacity in milliwatt-hours
	DutyCyclePercent  float64 // Active duty cycle percentage
	PrivacyRequired   bool    // On-device inference only
	RealTimeRequired  bool    // Hard real-time constraints
	EnvironmentalTemp float64 // Operating temperature in Celsius
}

// NewEdgeDeploymentConfig creates a deployment config for specific device type
func NewEdgeDeploymentConfig(deviceType EdgeDeviceType) *EdgeDeploymentConfig {
	configs := map[EdgeDeviceType]*EdgeDeploymentConfig{
		EdgeDeviceIoTSensor: {
			DeviceType:        EdgeDeviceIoTSensor,
			PowerBudgetMW:     1.0,     // 1mW for ultra-low power
			LatencyBudgetMs:   100.0,   // 100ms acceptable
			MemoryBudgetKB:    64.0,    // 64KB SRAM
			AccuracyThreshold: 0.85,
			BatteryCapacityMWh: 500.0,  // Coin cell battery
			DutyCyclePercent:  1.0,     // 1% duty cycle
			PrivacyRequired:   false,
			RealTimeRequired:  false,
			EnvironmentalTemp: 25.0,
		},
		EdgeDeviceWearable: {
			DeviceType:        EdgeDeviceWearable,
			PowerBudgetMW:     10.0,    // 10mW for wearables
			LatencyBudgetMs:   50.0,    // 50ms for responsive UI
			MemoryBudgetKB:    256.0,   // 256KB SRAM
			AccuracyThreshold: 0.90,
			BatteryCapacityMWh: 2000.0, // Smartwatch battery
			DutyCyclePercent:  10.0,    // 10% active
			PrivacyRequired:   true,    // Health data privacy
			RealTimeRequired:  false,
			EnvironmentalTemp: 30.0,    // Body heat
		},
		EdgeDeviceAutomotive: {
			DeviceType:        EdgeDeviceAutomotive,
			PowerBudgetMW:     1000.0,  // 1W for automotive
			LatencyBudgetMs:   10.0,    // 10ms for safety-critical
			MemoryBudgetKB:    4096.0,  // 4MB SRAM
			AccuracyThreshold: 0.99,    // High accuracy required
			BatteryCapacityMWh: 0.0,    // Powered by vehicle
			DutyCyclePercent:  100.0,   // Always on
			PrivacyRequired:   false,
			RealTimeRequired:  true,    // Hard real-time
			EnvironmentalTemp: 85.0,    // High temp automotive
		},
		EdgeDeviceSmartHome: {
			DeviceType:        EdgeDeviceSmartHome,
			PowerBudgetMW:     100.0,   // 100mW for smart home
			LatencyBudgetMs:   200.0,   // 200ms acceptable
			MemoryBudgetKB:    1024.0,  // 1MB SRAM
			AccuracyThreshold: 0.92,
			BatteryCapacityMWh: 0.0,    // Wall powered
			DutyCyclePercent:  20.0,    // Wake on event
			PrivacyRequired:   true,    // Home privacy
			RealTimeRequired:  false,
			EnvironmentalTemp: 25.0,
		},
		EdgeDeviceIndustrial: {
			DeviceType:        EdgeDeviceIndustrial,
			PowerBudgetMW:     500.0,   // 500mW for industrial
			LatencyBudgetMs:   5.0,     // 5ms for control loops
			MemoryBudgetKB:    2048.0,  // 2MB SRAM
			AccuracyThreshold: 0.95,
			BatteryCapacityMWh: 0.0,    // Powered
			DutyCyclePercent:  100.0,   // Continuous monitoring
			PrivacyRequired:   false,
			RealTimeRequired:  true,    // Control timing
			EnvironmentalTemp: 60.0,    // Industrial heat
		},
		EdgeDeviceBiomedical: {
			DeviceType:        EdgeDeviceBiomedical,
			PowerBudgetMW:     0.5,     // 0.5mW for implantables
			LatencyBudgetMs:   1000.0,  // 1s for monitoring
			MemoryBudgetKB:    32.0,    // 32KB SRAM
			AccuracyThreshold: 0.95,    // Medical accuracy
			BatteryCapacityMWh: 100.0,  // Small implant battery
			DutyCyclePercent:  0.1,     // Very low duty cycle
			PrivacyRequired:   true,    // HIPAA compliance
			RealTimeRequired:  false,
			EnvironmentalTemp: 37.0,    // Body temperature
		},
	}
	if config, ok := configs[deviceType]; ok {
		return config
	}
	return configs[EdgeDeviceIoTSensor]
}

// IoTSensorNode represents an IoT sensor with in-sensor CIM
type IoTSensorNode struct {
	Config           *EdgeDeploymentConfig
	CIMArray         *InSensorCIMArray
	SensorType       string
	SamplingRateHz   float64
	DataWidthBits    int
	PreprocessStages []PreprocessStage
	InferenceModel   *EdgeNeuralNetwork
	OutputActions    []string
	BatteryLevel     float64 // 0.0 to 1.0
	ActiveTime       float64 // Cumulative active time in hours
}

// InSensorCIMArray represents CIM integrated with sensor
type InSensorCIMArray struct {
	Rows             int
	Cols             int
	CellType         string // "ferroelectric", "rram", "sram"
	BitPrecision     int
	EnergyPerOpPJ    float64 // Picojoules per operation
	ThroughputGOPS   float64 // Giga-operations per second
	ArrayUtilization float64 // 0.0 to 1.0
	Temperature      float64
	NonIdealityLevel float64 // 0.0 to 1.0
}

// PreprocessStage represents a preprocessing stage
type PreprocessStage struct {
	Name         string
	Operation    string // "filter", "normalize", "downsample", "feature_extract"
	EnergyPJ     float64
	LatencyUs    float64
	OutputSize   int
}

// NewIoTSensorNode creates a new IoT sensor node
func NewIoTSensorNode(sensorType string, rows, cols int) *IoTSensorNode {
	config := NewEdgeDeploymentConfig(EdgeDeviceIoTSensor)

	return &IoTSensorNode{
		Config: config,
		CIMArray: &InSensorCIMArray{
			Rows:             rows,
			Cols:             cols,
			CellType:         "ferroelectric",
			BitPrecision:     4,
			EnergyPerOpPJ:    0.5, // 0.5 pJ/op for ferroelectric
			ThroughputGOPS:   1.0,
			ArrayUtilization: 0.0,
			Temperature:      config.EnvironmentalTemp,
			NonIdealityLevel: 0.05,
		},
		SensorType:     sensorType,
		SamplingRateHz: 100.0,
		DataWidthBits:  12,
		PreprocessStages: []PreprocessStage{
			{Name: "LPF", Operation: "filter", EnergyPJ: 10.0, LatencyUs: 1.0, OutputSize: 64},
			{Name: "Norm", Operation: "normalize", EnergyPJ: 5.0, LatencyUs: 0.5, OutputSize: 64},
		},
		InferenceModel: nil, // Set later
		OutputActions:  []string{"alert", "log", "transmit"},
		BatteryLevel:   1.0,
		ActiveTime:     0.0,
	}
}

// ProcessSample processes a sensor sample through in-sensor CIM
func (n *IoTSensorNode) ProcessSample(sample []float64) (*IoTInferenceResult, error) {
	if n.BatteryLevel <= 0 {
		return nil, fmt.Errorf("battery depleted")
	}

	result := &IoTInferenceResult{
		SensorType:    n.SensorType,
		InputSize:     len(sample),
		Preprocessed:  false,
		InferenceRun:  false,
		TotalEnergyPJ: 0.0,
		TotalLatencyUs: 0.0,
	}

	// Preprocessing stages
	data := sample
	for _, stage := range n.PreprocessStages {
		data = n.applyPreprocessing(data, stage)
		result.TotalEnergyPJ += stage.EnergyPJ
		result.TotalLatencyUs += stage.LatencyUs
	}
	result.Preprocessed = true
	result.PreprocessedSize = len(data)

	// CIM inference
	if n.InferenceModel != nil {
		inferenceResult := n.runCIMInference(data)
		result.InferenceRun = true
		result.Classification = inferenceResult.Classification
		result.Confidence = inferenceResult.Confidence
		result.TotalEnergyPJ += inferenceResult.EnergyPJ
		result.TotalLatencyUs += inferenceResult.LatencyUs
	}

	// Update battery consumption
	energyMWh := result.TotalEnergyPJ * 1e-9 / 3600.0 // Convert pJ to mWh
	n.BatteryLevel -= energyMWh / n.Config.BatteryCapacityMWh
	if n.BatteryLevel < 0 {
		n.BatteryLevel = 0
	}

	// Update active time
	n.ActiveTime += result.TotalLatencyUs * 1e-6 / 3600.0 // Convert to hours

	return result, nil
}

func (n *IoTSensorNode) applyPreprocessing(data []float64, stage PreprocessStage) []float64 {
	switch stage.Operation {
	case "filter":
		// Simple low-pass filter simulation
		filtered := make([]float64, len(data))
		alpha := 0.3
		filtered[0] = data[0]
		for i := 1; i < len(data); i++ {
			filtered[i] = alpha*data[i] + (1-alpha)*filtered[i-1]
		}
		return filtered
	case "normalize":
		// Min-max normalization
		minVal, maxVal := data[0], data[0]
		for _, v := range data {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
		normalized := make([]float64, len(data))
		rangeVal := maxVal - minVal
		if rangeVal == 0 {
			rangeVal = 1
		}
		for i, v := range data {
			normalized[i] = (v - minVal) / rangeVal
		}
		return normalized
	case "downsample":
		// Downsample by factor of 2
		downsampled := make([]float64, len(data)/2)
		for i := 0; i < len(downsampled); i++ {
			downsampled[i] = (data[2*i] + data[2*i+1]) / 2.0
		}
		return downsampled
	default:
		return data
	}
}

func (n *IoTSensorNode) runCIMInference(data []float64) *CIMInferenceResult {
	// Simulate CIM inference with non-idealities
	ops := n.CIMArray.Rows * n.CIMArray.Cols
	energyPJ := float64(ops) * n.CIMArray.EnergyPerOpPJ
	latencyUs := float64(ops) / (n.CIMArray.ThroughputGOPS * 1e3)

	// Add non-ideality effects
	noiseLevel := n.CIMArray.NonIdealityLevel * rand.Float64()
	confidence := 0.95 - noiseLevel

	return &CIMInferenceResult{
		Classification: 0, // Placeholder
		Confidence:     confidence,
		EnergyPJ:       energyPJ,
		LatencyUs:      latencyUs,
	}
}

// IoTInferenceResult holds inference results from IoT sensor
type IoTInferenceResult struct {
	SensorType      string
	InputSize       int
	Preprocessed    bool
	PreprocessedSize int
	InferenceRun    bool
	Classification  int
	Confidence      float64
	TotalEnergyPJ   float64
	TotalLatencyUs  float64
}

// CIMInferenceResult holds CIM-specific inference metrics
type CIMInferenceResult struct {
	Classification int
	Confidence     float64
	EnergyPJ       float64
	LatencyUs      float64
}

// GetBatteryLifeHours estimates remaining battery life
func (n *IoTSensorNode) GetBatteryLifeHours() float64 {
	if n.ActiveTime == 0 {
		return math.Inf(1)
	}
	consumedMWh := (1.0 - n.BatteryLevel) * n.Config.BatteryCapacityMWh
	consumptionRate := consumedMWh / n.ActiveTime // mWh per hour
	remainingMWh := n.BatteryLevel * n.Config.BatteryCapacityMWh
	if consumptionRate == 0 {
		return math.Inf(1)
	}
	return remainingMWh / consumptionRate
}

// WearableHealthDevice represents a wearable health monitoring device
type WearableHealthDevice struct {
	Config           *EdgeDeploymentConfig
	CIMProcessor     *WearableCIMProcessor
	Sensors          []WearableSensor
	HealthModels     []*HealthInferenceModel
	PrivacyModule    *OnDevicePrivacy
	AlertThresholds  map[string]float64
	UserProfile      *HealthUserProfile
}

// WearableCIMProcessor represents CIM processor in wearable
type WearableCIMProcessor struct {
	ArraySize        int
	NumArrays        int
	BitPrecision     int
	PeakTOPS         float64
	EfficiencyTOPSW  float64 // TOPS per Watt
	SupportedOps     []string
	CurrentPowerMW   float64
	ThermalThrottle  bool
}

// WearableSensor represents a sensor in wearable device
type WearableSensor struct {
	Name           string
	Type           string // "ppg", "ecg", "imu", "temperature", "spo2"
	SamplingRateHz float64
	Resolution     int
	PowerUW        float64 // Microwatts
	DataRateBps    float64
}

// HealthInferenceModel represents a health inference model
type HealthInferenceModel struct {
	Name           string
	Task           string // "heart_rate", "arrhythmia", "activity", "stress", "sleep"
	InputSensors   []string
	ModelSizeKB    float64
	AccuracyTarget float64
	LatencyTargetMs float64
	QuantizationBits int
}

// OnDevicePrivacy handles privacy-preserving computation
type OnDevicePrivacy struct {
	Enabled           bool
	DataEncryption    bool
	FederatedLearning bool
	DifferentialPrivacy bool
	EpsilonBudget     float64
	LocalProcessing   bool
}

// HealthUserProfile stores user health baseline
type HealthUserProfile struct {
	Age              int
	RestingHeartRate float64
	BaselineMetrics  map[string]float64
	HealthConditions []string
}

// NewWearableHealthDevice creates a new wearable health device
func NewWearableHealthDevice() *WearableHealthDevice {
	config := NewEdgeDeploymentConfig(EdgeDeviceWearable)

	return &WearableHealthDevice{
		Config: config,
		CIMProcessor: &WearableCIMProcessor{
			ArraySize:       256,
			NumArrays:       4,
			BitPrecision:    8,
			PeakTOPS:        0.5,
			EfficiencyTOPSW: 50.0, // 50 TOPS/W
			SupportedOps:    []string{"conv2d", "dense", "pooling", "attention"},
			CurrentPowerMW:  0.0,
			ThermalThrottle: false,
		},
		Sensors: []WearableSensor{
			{Name: "PPG", Type: "ppg", SamplingRateHz: 100.0, Resolution: 16, PowerUW: 200.0, DataRateBps: 1600.0},
			{Name: "ECG", Type: "ecg", SamplingRateHz: 250.0, Resolution: 12, PowerUW: 500.0, DataRateBps: 3000.0},
			{Name: "IMU", Type: "imu", SamplingRateHz: 50.0, Resolution: 16, PowerUW: 100.0, DataRateBps: 4800.0},
			{Name: "Temp", Type: "temperature", SamplingRateHz: 1.0, Resolution: 12, PowerUW: 10.0, DataRateBps: 12.0},
		},
		HealthModels: []*HealthInferenceModel{
			{Name: "HeartRate", Task: "heart_rate", InputSensors: []string{"ppg"}, ModelSizeKB: 50.0, AccuracyTarget: 0.95, LatencyTargetMs: 100.0, QuantizationBits: 8},
			{Name: "Arrhythmia", Task: "arrhythmia", InputSensors: []string{"ecg"}, ModelSizeKB: 100.0, AccuracyTarget: 0.90, LatencyTargetMs: 500.0, QuantizationBits: 8},
			{Name: "Activity", Task: "activity", InputSensors: []string{"imu"}, ModelSizeKB: 75.0, AccuracyTarget: 0.92, LatencyTargetMs: 200.0, QuantizationBits: 4},
		},
		PrivacyModule: &OnDevicePrivacy{
			Enabled:           true,
			DataEncryption:    true,
			FederatedLearning: true,
			DifferentialPrivacy: true,
			EpsilonBudget:     1.0,
			LocalProcessing:   true,
		},
		AlertThresholds: map[string]float64{
			"heart_rate_high": 150.0,
			"heart_rate_low":  40.0,
			"arrhythmia_risk": 0.8,
			"temperature_fever": 38.0,
		},
		UserProfile: &HealthUserProfile{
			Age:              30,
			RestingHeartRate: 70.0,
			BaselineMetrics:  make(map[string]float64),
			HealthConditions: []string{},
		},
	}
}

// AutomotiveEdgeSystem represents automotive edge AI system
type AutomotiveEdgeSystem struct {
	Config           *EdgeDeploymentConfig
	CIMAccelerator   *AutomotiveCIMAccelerator
	VisionPipeline   *VisionProcessingPipeline
	SensorFusion     *MultiSensorFusion
	SafetyMonitor    *FunctionalSafetyMonitor
	PerceptionModels []*PerceptionModel
}

// AutomotiveCIMAccelerator represents high-performance automotive CIM
type AutomotiveCIMAccelerator struct {
	NumCores         int
	CoreArraySize    int
	TotalTOPS        float64
	EfficiencyTOPSW  float64
	OperatingTempC   float64
	MaxTempC         float64
	RedundancyLevel  int // For safety
	ErrorCorrection  bool
	ASILLevel        string // "A", "B", "C", "D"
}

// VisionProcessingPipeline represents vision processing pipeline
type VisionProcessingPipeline struct {
	InputResolution  [2]int
	FrameRateFPS     float64
	ColorDepth       int
	Stages           []VisionStage
	OutputFormat     string
}

// VisionStage represents a stage in vision pipeline
type VisionStage struct {
	Name          string
	Operation     string
	LatencyMs     float64
	PowerMW       float64
	OnCIM         bool
}

// MultiSensorFusion handles multi-modal sensor fusion
type MultiSensorFusion struct {
	Cameras        int
	LiDARs         int
	Radars         int
	Ultrasonics    int
	FusionMethod   string // "early", "late", "hybrid"
	TemporalWindow float64 // seconds
}

// FunctionalSafetyMonitor ensures automotive safety
type FunctionalSafetyMonitor struct {
	ASILLevel          string
	WatchdogEnabled    bool
	RedundantCompute   bool
	ErrorDetection     bool
	FailsafeActions    []string
	DiagnosticCoverage float64
}

// PerceptionModel represents an automotive perception model
type PerceptionModel struct {
	Name            string
	Task            string // "object_detection", "lane_detection", "depth_estimation", "semantic_seg"
	InputModalities []string
	LatencyTargetMs float64
	AccuracyTarget  float64 // mAP or IoU
	SafetyCritical  bool
}

// NewAutomotiveEdgeSystem creates automotive edge system
func NewAutomotiveEdgeSystem() *AutomotiveEdgeSystem {
	config := NewEdgeDeploymentConfig(EdgeDeviceAutomotive)

	return &AutomotiveEdgeSystem{
		Config: config,
		CIMAccelerator: &AutomotiveCIMAccelerator{
			NumCores:        16,
			CoreArraySize:   512,
			TotalTOPS:       100.0, // 100 TOPS
			EfficiencyTOPSW: 100.0, // 100 TOPS/W
			OperatingTempC:  config.EnvironmentalTemp,
			MaxTempC:        125.0,
			RedundancyLevel: 2,     // Dual redundancy
			ErrorCorrection: true,
			ASILLevel:       "D",   // Highest safety level
		},
		VisionPipeline: &VisionProcessingPipeline{
			InputResolution: [2]int{1920, 1080},
			FrameRateFPS:    30.0,
			ColorDepth:      8,
			Stages: []VisionStage{
				{Name: "Debayer", Operation: "image_processing", LatencyMs: 0.5, PowerMW: 50.0, OnCIM: false},
				{Name: "Resize", Operation: "resize", LatencyMs: 0.2, PowerMW: 20.0, OnCIM: false},
				{Name: "Backbone", Operation: "conv_backbone", LatencyMs: 5.0, PowerMW: 500.0, OnCIM: true},
				{Name: "Neck", Operation: "feature_pyramid", LatencyMs: 1.0, PowerMW: 100.0, OnCIM: true},
				{Name: "Head", Operation: "detection_head", LatencyMs: 2.0, PowerMW: 200.0, OnCIM: true},
			},
			OutputFormat: "bboxes",
		},
		SensorFusion: &MultiSensorFusion{
			Cameras:        8,
			LiDARs:         1,
			Radars:         5,
			Ultrasonics:    12,
			FusionMethod:   "hybrid",
			TemporalWindow: 0.1, // 100ms
		},
		SafetyMonitor: &FunctionalSafetyMonitor{
			ASILLevel:          "D",
			WatchdogEnabled:    true,
			RedundantCompute:   true,
			ErrorDetection:     true,
			FailsafeActions:    []string{"emergency_brake", "safe_stop", "driver_alert"},
			DiagnosticCoverage: 0.99, // 99% diagnostic coverage
		},
		PerceptionModels: []*PerceptionModel{
			{Name: "YOLO-CIM", Task: "object_detection", InputModalities: []string{"camera"}, LatencyTargetMs: 10.0, AccuracyTarget: 0.85, SafetyCritical: true},
			{Name: "LaneNet-CIM", Task: "lane_detection", InputModalities: []string{"camera"}, LatencyTargetMs: 15.0, AccuracyTarget: 0.95, SafetyCritical: true},
			{Name: "DepthNet-CIM", Task: "depth_estimation", InputModalities: []string{"camera", "lidar"}, LatencyTargetMs: 20.0, AccuracyTarget: 0.90, SafetyCritical: false},
		},
	}
}

// =============================================================================
// CIM COMPILER AND MAPPER FRAMEWORK
// =============================================================================

// CIMCompilerConfig configures the CIM compiler
type CIMCompilerConfig struct {
	TargetHardware    *CIMHardwareSpec
	OptimizationLevel int    // 0-3
	MemoryConstraint  int    // KB
	LatencyConstraint float64 // ms
	EnergyConstraint  float64 // mJ
	EnableFusion      bool
	EnableTiling      bool
	EnablePipelining  bool
	MappingStrategy   string // "greedy", "dp", "ilp", "ml_guided"
}

// CIMHardwareSpec specifies target CIM hardware
type CIMHardwareSpec struct {
	Name             string
	NumTiles         int
	TileRows         int
	TileCols         int
	BitPrecision     int
	MemoryBanks      int
	MemoryPerBankKB  float64
	InterconnectType string // "mesh", "ring", "crossbar"
	PeakTOPS         float64
	EnergyPerMACpJ   float64
	SupportsSparsity bool
	SupportedOps     []string
}

// CIMCompiler compiles neural networks for CIM execution
type CIMCompiler struct {
	Config          *CIMCompilerConfig
	IRGraph         *CIMComputeGraph
	MappedGraph     *MappedCIMGraph
	ScheduledOps    []*ScheduledCIMOp
	CompilationStats *CompilationStatistics
}

// CIMComputeGraph represents intermediate representation
type CIMComputeGraph struct {
	Nodes    []*CIMGraphNode
	Edges    []*CIMGraphEdge
	Inputs   []int
	Outputs  []int
	Metadata map[string]interface{}
}

// CIMGraphNode represents a node in compute graph
type CIMGraphNode struct {
	ID            int
	OpType        string
	Name          string
	InputShapes   [][]int
	OutputShape   []int
	Weights       []float64
	Attributes    map[string]interface{}
	ComputeOps    int64
	MemoryBytes   int64
	MappingHint   string
}

// CIMGraphEdge represents data flow edge
type CIMGraphEdge struct {
	SrcNode   int
	DstNode   int
	TensorID  int
	DataType  string
	Shape     []int
}

// MappedCIMGraph represents hardware-mapped graph
type MappedCIMGraph struct {
	TileAssignments map[int]*TileAssignment
	MemoryAlloc     map[int]*MemoryAllocation
	DataMovements   []*DataMovement
	PipelineStages  [][]*CIMGraphNode
}

// TileAssignment represents tile assignment for a node
type TileAssignment struct {
	NodeID       int
	TileIDs      []int
	Utilization  float64
	RowMapping   []int
	ColMapping   []int
	Duplications int
}

// MemoryAllocation represents memory allocation
type MemoryAllocation struct {
	TensorID    int
	BankID      int
	Offset      int
	Size        int
	Lifetime    [2]int // [start_cycle, end_cycle]
	Reusable    bool
}

// DataMovement represents data movement between tiles
type DataMovement struct {
	SrcTile     int
	DstTile     int
	TensorID    int
	SizeBytes   int
	LatencyCycles int
	EnergyPJ    float64
}

// ScheduledCIMOp represents a scheduled operation
type ScheduledCIMOp struct {
	NodeID        int
	TileID        int
	StartCycle    int
	EndCycle      int
	Dependencies  []int
	MemReads      []*MemoryAllocation
	MemWrites     []*MemoryAllocation
}

// CompilationStatistics tracks compilation metrics
type CompilationStatistics struct {
	TotalNodes        int
	MappedNodes       int
	FusedNodes        int
	TileUtilization   float64
	MemoryUtilization float64
	EstimatedCycles   int64
	EstimatedEnergyMJ float64
	EstimatedLatencyMs float64
	OptimizationPasses []string
}

// NewCIMCompiler creates a new CIM compiler
func NewCIMCompiler(config *CIMCompilerConfig) *CIMCompiler {
	return &CIMCompiler{
		Config:          config,
		IRGraph:         nil,
		MappedGraph:     nil,
		ScheduledOps:    nil,
		CompilationStats: &CompilationStatistics{},
	}
}

// Compile compiles a neural network model
func (c *CIMCompiler) Compile(modelPath string) error {
	// Step 1: Parse model and build IR
	if err := c.buildIRGraph(modelPath); err != nil {
		return fmt.Errorf("IR build failed: %w", err)
	}

	// Step 2: Graph optimizations
	c.optimizeGraph()

	// Step 3: Map to hardware
	if err := c.mapToHardware(); err != nil {
		return fmt.Errorf("mapping failed: %w", err)
	}

	// Step 4: Schedule operations
	if err := c.scheduleOperations(); err != nil {
		return fmt.Errorf("scheduling failed: %w", err)
	}

	// Step 5: Generate code
	c.updateStatistics()

	return nil
}

func (c *CIMCompiler) buildIRGraph(modelPath string) error {
	// Simplified IR graph construction
	c.IRGraph = &CIMComputeGraph{
		Nodes:    make([]*CIMGraphNode, 0),
		Edges:    make([]*CIMGraphEdge, 0),
		Inputs:   []int{0},
		Outputs:  []int{},
		Metadata: make(map[string]interface{}),
	}

	// Add example nodes (in real implementation, parse from model file)
	c.IRGraph.Nodes = append(c.IRGraph.Nodes, &CIMGraphNode{
		ID:          0,
		OpType:      "input",
		Name:        "input_0",
		OutputShape: []int{1, 224, 224, 3},
		ComputeOps:  0,
		MemoryBytes: 224 * 224 * 3 * 4,
	})

	c.CompilationStats.TotalNodes = len(c.IRGraph.Nodes)
	return nil
}

func (c *CIMCompiler) optimizeGraph() {
	passes := []string{}

	if c.Config.EnableFusion {
		c.fuseOperations()
		passes = append(passes, "op_fusion")
	}

	if c.Config.EnableTiling {
		c.applyTiling()
		passes = append(passes, "tiling")
	}

	c.CompilationStats.OptimizationPasses = passes
}

func (c *CIMCompiler) fuseOperations() {
	// Identify fusable patterns: Conv+BN+ReLU, MatMul+Add, etc.
	fusedCount := 0
	for i := 0; i < len(c.IRGraph.Nodes)-2; i++ {
		node := c.IRGraph.Nodes[i]
		if node.OpType == "conv2d" && i+2 < len(c.IRGraph.Nodes) {
			next1 := c.IRGraph.Nodes[i+1]
			next2 := c.IRGraph.Nodes[i+2]
			if next1.OpType == "batch_norm" && next2.OpType == "relu" {
				// Mark for fusion
				node.Attributes["fused_ops"] = []string{"batch_norm", "relu"}
				fusedCount += 2
			}
		}
	}
	c.CompilationStats.FusedNodes = fusedCount
}

func (c *CIMCompiler) applyTiling() {
	// Apply tiling for large layers
	hw := c.Config.TargetHardware
	for _, node := range c.IRGraph.Nodes {
		if len(node.OutputShape) >= 2 {
			h, w := 1, 1
			if len(node.OutputShape) >= 2 {
				h = node.OutputShape[len(node.OutputShape)-2]
				w = node.OutputShape[len(node.OutputShape)-1]
			}
			if h > hw.TileRows || w > hw.TileCols {
				// Compute tiling factors
				tileH := (h + hw.TileRows - 1) / hw.TileRows
				tileW := (w + hw.TileCols - 1) / hw.TileCols
				node.Attributes["tile_factors"] = []int{tileH, tileW}
			}
		}
	}
}

func (c *CIMCompiler) mapToHardware() error {
	c.MappedGraph = &MappedCIMGraph{
		TileAssignments: make(map[int]*TileAssignment),
		MemoryAlloc:     make(map[int]*MemoryAllocation),
		DataMovements:   make([]*DataMovement, 0),
		PipelineStages:  make([][]*CIMGraphNode, 0),
	}

	hw := c.Config.TargetHardware
	currentTile := 0
	totalUtilization := 0.0

	for _, node := range c.IRGraph.Nodes {
		if node.OpType == "input" || node.OpType == "output" {
			continue
		}

		// Greedy tile assignment
		assignment := &TileAssignment{
			NodeID:      node.ID,
			TileIDs:     []int{currentTile % hw.NumTiles},
			Utilization: c.estimateUtilization(node, hw),
			RowMapping:  []int{0, hw.TileRows},
			ColMapping:  []int{0, hw.TileCols},
		}
		c.MappedGraph.TileAssignments[node.ID] = assignment
		totalUtilization += assignment.Utilization
		currentTile++
		c.CompilationStats.MappedNodes++
	}

	if c.CompilationStats.MappedNodes > 0 {
		c.CompilationStats.TileUtilization = totalUtilization / float64(c.CompilationStats.MappedNodes)
	}

	return nil
}

func (c *CIMCompiler) estimateUtilization(node *CIMGraphNode, hw *CIMHardwareSpec) float64 {
	totalCells := hw.TileRows * hw.TileCols
	if len(node.OutputShape) < 2 {
		return 0.5
	}
	usedCells := node.OutputShape[len(node.OutputShape)-2] * node.OutputShape[len(node.OutputShape)-1]
	util := float64(usedCells) / float64(totalCells)
	if util > 1.0 {
		util = 1.0
	}
	return util
}

func (c *CIMCompiler) scheduleOperations() error {
	c.ScheduledOps = make([]*ScheduledCIMOp, 0)
	currentCycle := 0

	for _, node := range c.IRGraph.Nodes {
		if node.OpType == "input" || node.OpType == "output" {
			continue
		}

		assignment := c.MappedGraph.TileAssignments[node.ID]
		if assignment == nil {
			continue
		}

		// Estimate operation duration
		duration := c.estimateOpDuration(node)

		op := &ScheduledCIMOp{
			NodeID:       node.ID,
			TileID:       assignment.TileIDs[0],
			StartCycle:   currentCycle,
			EndCycle:     currentCycle + duration,
			Dependencies: []int{},
		}
		c.ScheduledOps = append(c.ScheduledOps, op)
		currentCycle += duration
	}

	c.CompilationStats.EstimatedCycles = int64(currentCycle)
	return nil
}

func (c *CIMCompiler) estimateOpDuration(node *CIMGraphNode) int {
	// Simplified duration estimation
	ops := node.ComputeOps
	if ops == 0 {
		ops = 1000000 // Default 1M ops
	}
	hw := c.Config.TargetHardware
	opsPerCycle := int64(hw.TileRows * hw.TileCols)
	return int(ops / opsPerCycle) + 1
}

func (c *CIMCompiler) updateStatistics() {
	hw := c.Config.TargetHardware
	cycles := c.CompilationStats.EstimatedCycles

	// Clock frequency assumption: 1 GHz
	clockFreqGHz := 1.0
	c.CompilationStats.EstimatedLatencyMs = float64(cycles) / (clockFreqGHz * 1e6)

	// Energy estimation
	totalOps := int64(0)
	for _, node := range c.IRGraph.Nodes {
		totalOps += node.ComputeOps
	}
	c.CompilationStats.EstimatedEnergyMJ = float64(totalOps) * hw.EnergyPerMACpJ * 1e-9
}

// CMSwitchCompiler implements CMSwitch dual-mode compilation (ASPLOS 2025)
type CMSwitchCompiler struct {
	BaseCompiler    *CIMCompiler
	ComputeMode     bool // true = compute-centric, false = memory-centric
	ModeThreshold   float64
	SwitchPoints    []int
	ModeHistory     []bool
}

// NewCMSwitchCompiler creates CMSwitch compiler
func NewCMSwitchCompiler(config *CIMCompilerConfig) *CMSwitchCompiler {
	return &CMSwitchCompiler{
		BaseCompiler:  NewCIMCompiler(config),
		ComputeMode:   true,
		ModeThreshold: 0.7, // Switch mode if utilization drops below 70%
		SwitchPoints:  make([]int, 0),
		ModeHistory:   make([]bool, 0),
	}
}

// CompileWithModeSwitching compiles with adaptive mode switching
func (c *CMSwitchCompiler) CompileWithModeSwitching(modelPath string) error {
	// Build IR first
	if err := c.BaseCompiler.buildIRGraph(modelPath); err != nil {
		return err
	}

	// Analyze each layer for optimal mode
	for i, node := range c.BaseCompiler.IRGraph.Nodes {
		optimalMode := c.determineOptimalMode(node)
		if optimalMode != c.ComputeMode {
			c.SwitchPoints = append(c.SwitchPoints, i)
			c.ComputeMode = optimalMode
		}
		c.ModeHistory = append(c.ModeHistory, c.ComputeMode)
	}

	// Continue with regular compilation
	c.BaseCompiler.optimizeGraph()
	if err := c.BaseCompiler.mapToHardware(); err != nil {
		return err
	}
	return c.BaseCompiler.scheduleOperations()
}

func (c *CMSwitchCompiler) determineOptimalMode(node *CIMGraphNode) bool {
	// Compute-centric: high arithmetic intensity
	// Memory-centric: memory-bound operations

	if node.ComputeOps == 0 || node.MemoryBytes == 0 {
		return true // Default to compute mode
	}

	arithmeticIntensity := float64(node.ComputeOps) / float64(node.MemoryBytes)

	// Threshold: 10 ops/byte is compute-bound
	return arithmeticIntensity > 10.0
}

// =============================================================================
// NON-IDEALITY AWARE TRAINING (NORA Framework)
// =============================================================================

// NonIdealityConfig configures non-ideality modeling
type NonIdealityConfig struct {
	// Conductance variation
	ConductanceVariation float64 // Standard deviation as fraction
	ProgrammingNoise     float64
	ReadNoise            float64

	// IR drop
	IRDropEnabled        bool
	WireResistanceOhm    float64
	MaxIRDropPercent     float64

	// Stuck-at faults
	StuckAtFaultRate     float64
	StuckAtHigh          float64
	StuckAtLow           float64

	// Temperature effects
	TempCoefficient      float64 // Conductance change per degree C
	OperatingTempC       float64
	ReferenceTempC       float64

	// Quantization
	WeightBits           int
	ActivationBits       int
	ADCBits              int
	DACBits              int
}

// NewNonIdealityConfig creates default non-ideality config
func NewNonIdealityConfig() *NonIdealityConfig {
	return &NonIdealityConfig{
		ConductanceVariation: 0.05,  // 5% variation
		ProgrammingNoise:     0.02,  // 2% programming error
		ReadNoise:            0.01,  // 1% read noise
		IRDropEnabled:        true,
		WireResistanceOhm:    0.1,
		MaxIRDropPercent:     10.0,
		StuckAtFaultRate:     0.001, // 0.1% fault rate
		StuckAtHigh:          0.5,   // 50% stuck-high
		StuckAtLow:           0.5,   // 50% stuck-low
		TempCoefficient:      -0.002, // -0.2%/C
		OperatingTempC:       85.0,
		ReferenceTempC:       25.0,
		WeightBits:           4,
		ActivationBits:       8,
		ADCBits:              8,
		DACBits:              4,
	}
}

// HardwareAwareTrainer implements hardware-aware training
type HardwareAwareTrainer struct {
	Config              *NonIdealityConfig
	Model               *EdgeNeuralNetwork
	NoiseInjector       *NoiseInjector
	QuantizationAware   bool
	ClippingEnabled     bool
	RescalingFactors    []float64
	TrainingHistory     *TrainingHistory
}

// NoiseInjector injects hardware non-idealities during training
type NoiseInjector struct {
	Config            *NonIdealityConfig
	InjectionMode     string // "forward", "backward", "both"
	NoiseSchedule     string // "constant", "annealing", "cyclic"
	CurrentNoiseScale float64
	Epoch             int
}

// TrainingHistory tracks training metrics
type TrainingHistory struct {
	Epochs          []int
	TrainLoss       []float64
	ValLoss         []float64
	TrainAccuracy   []float64
	ValAccuracy     []float64
	HardwareAccuracy []float64 // Accuracy with hardware noise
	NoiseScales     []float64
}

// EdgeNeuralNetwork represents a neural network for edge deployment
type EdgeNeuralNetwork struct {
	Layers           []*EdgeNNLayer
	TotalParams      int64
	TotalOps         int64
	MemoryFootprintKB float64
	QuantizedWeights [][]float64
}

// EdgeNNLayer represents a layer in edge neural network
type EdgeNNLayer struct {
	Name           string
	Type           string
	InputShape     []int
	OutputShape    []int
	Weights        []float64
	Bias           []float64
	NumParams      int64
	NumOps         int64
	Quantized      bool
	QuantBits      int
}

// NewHardwareAwareTrainer creates hardware-aware trainer
func NewHardwareAwareTrainer(config *NonIdealityConfig) *HardwareAwareTrainer {
	return &HardwareAwareTrainer{
		Config: config,
		Model:  nil,
		NoiseInjector: &NoiseInjector{
			Config:            config,
			InjectionMode:     "both",
			NoiseSchedule:     "annealing",
			CurrentNoiseScale: 1.0,
			Epoch:             0,
		},
		QuantizationAware: true,
		ClippingEnabled:   true,
		RescalingFactors:  make([]float64, 0),
		TrainingHistory: &TrainingHistory{
			Epochs:           make([]int, 0),
			TrainLoss:        make([]float64, 0),
			ValLoss:          make([]float64, 0),
			TrainAccuracy:    make([]float64, 0),
			ValAccuracy:      make([]float64, 0),
			HardwareAccuracy: make([]float64, 0),
			NoiseScales:      make([]float64, 0),
		},
	}
}

// TrainWithNonIdealities trains model with hardware noise injection
func (t *HardwareAwareTrainer) TrainWithNonIdealities(epochs int) error {
	if t.Model == nil {
		return fmt.Errorf("model not set")
	}

	for epoch := 0; epoch < epochs; epoch++ {
		t.NoiseInjector.Epoch = epoch
		t.updateNoiseScale(epoch, epochs)

		// Forward pass with noise
		trainLoss, trainAcc := t.trainEpoch()

		// Validation without noise
		valLoss, valAcc := t.validateEpoch(false)

		// Validation with hardware noise
		_, hwAcc := t.validateEpoch(true)

		// Record history
		t.TrainingHistory.Epochs = append(t.TrainingHistory.Epochs, epoch)
		t.TrainingHistory.TrainLoss = append(t.TrainingHistory.TrainLoss, trainLoss)
		t.TrainingHistory.ValLoss = append(t.TrainingHistory.ValLoss, valLoss)
		t.TrainingHistory.TrainAccuracy = append(t.TrainingHistory.TrainAccuracy, trainAcc)
		t.TrainingHistory.ValAccuracy = append(t.TrainingHistory.ValAccuracy, valAcc)
		t.TrainingHistory.HardwareAccuracy = append(t.TrainingHistory.HardwareAccuracy, hwAcc)
		t.TrainingHistory.NoiseScales = append(t.TrainingHistory.NoiseScales, t.NoiseInjector.CurrentNoiseScale)
	}

	// Compute optimal rescaling factors (NORA approach)
	t.computeRescalingFactors()

	return nil
}

func (t *HardwareAwareTrainer) updateNoiseScale(epoch, totalEpochs int) {
	switch t.NoiseInjector.NoiseSchedule {
	case "constant":
		t.NoiseInjector.CurrentNoiseScale = 1.0
	case "annealing":
		// Start with low noise, increase over time
		progress := float64(epoch) / float64(totalEpochs)
		t.NoiseInjector.CurrentNoiseScale = 0.1 + 0.9*progress
	case "cyclic":
		// Cyclic noise schedule
		cycle := math.Sin(2 * math.Pi * float64(epoch) / 10.0)
		t.NoiseInjector.CurrentNoiseScale = 0.5 + 0.5*cycle
	}
}

func (t *HardwareAwareTrainer) trainEpoch() (float64, float64) {
	// Simulated training epoch with noise injection
	baseLoss := 0.5 + rand.Float64()*0.2
	baseAcc := 0.7 + rand.Float64()*0.15

	// Noise reduces effective accuracy during training
	noiseEffect := t.NoiseInjector.CurrentNoiseScale * t.Config.ConductanceVariation
	loss := baseLoss * (1 + noiseEffect)
	acc := baseAcc * (1 - noiseEffect*0.5)

	return loss, acc
}

func (t *HardwareAwareTrainer) validateEpoch(withNoise bool) (float64, float64) {
	baseLoss := 0.3 + rand.Float64()*0.1
	baseAcc := 0.85 + rand.Float64()*0.1

	if withNoise {
		// Apply all non-ideality effects
		noiseEffect := t.Config.ConductanceVariation
		irDropEffect := 0.0
		if t.Config.IRDropEnabled {
			irDropEffect = t.Config.MaxIRDropPercent / 100.0
		}
		quantEffect := math.Pow(2, -float64(t.Config.WeightBits)) * 0.5

		totalEffect := noiseEffect + irDropEffect + quantEffect
		baseAcc *= (1 - totalEffect)
	}

	return baseLoss, baseAcc
}

func (t *HardwareAwareTrainer) computeRescalingFactors() {
	// NORA: Noise-Optimized Rescaling for Analog inference
	// Compute per-layer rescaling to minimize noise impact

	if t.Model == nil || len(t.Model.Layers) == 0 {
		return
	}

	t.RescalingFactors = make([]float64, len(t.Model.Layers))

	for i, layer := range t.Model.Layers {
		// Compute optimal scale based on weight distribution
		if len(layer.Weights) == 0 {
			t.RescalingFactors[i] = 1.0
			continue
		}

		// Find weight range
		maxAbs := 0.0
		for _, w := range layer.Weights {
			if math.Abs(w) > maxAbs {
				maxAbs = math.Abs(w)
			}
		}

		// Scale to use full ADC range while minimizing quantization error
		adcRange := math.Pow(2, float64(t.Config.ADCBits)) - 1
		if maxAbs > 0 {
			t.RescalingFactors[i] = adcRange / (2 * maxAbs)
		} else {
			t.RescalingFactors[i] = 1.0
		}
	}
}

// InjectNoise applies hardware non-idealities to weights
func (n *NoiseInjector) InjectNoise(weights []float64) []float64 {
	noisy := make([]float64, len(weights))
	scale := n.CurrentNoiseScale

	for i, w := range weights {
		// Conductance variation
		condNoise := rand.NormFloat64() * n.Config.ConductanceVariation * scale

		// Programming noise
		progNoise := rand.NormFloat64() * n.Config.ProgrammingNoise * scale

		// Read noise
		readNoise := rand.NormFloat64() * n.Config.ReadNoise * scale

		// Temperature effect
		tempDelta := n.Config.OperatingTempC - n.Config.ReferenceTempC
		tempEffect := 1.0 + n.Config.TempCoefficient*tempDelta

		// Stuck-at faults
		faultValue := w
		if rand.Float64() < n.Config.StuckAtFaultRate {
			if rand.Float64() < n.Config.StuckAtHigh {
				faultValue = 1.0 // Stuck high
			} else {
				faultValue = 0.0 // Stuck low
			}
		} else {
			faultValue = w * tempEffect * (1 + condNoise + progNoise + readNoise)
		}

		noisy[i] = faultValue
	}

	return noisy
}

// IRDropSimulator simulates IR drop in crossbar arrays
type IRDropSimulator struct {
	Config         *NonIdealityConfig
	ArrayRows      int
	ArrayCols      int
	WireResistance [][]float64 // Row and column wire resistance
	CellResistance [][]float64 // Cell resistance matrix
}

// NewIRDropSimulator creates IR drop simulator
func NewIRDropSimulator(config *NonIdealityConfig, rows, cols int) *IRDropSimulator {
	sim := &IRDropSimulator{
		Config:         config,
		ArrayRows:      rows,
		ArrayCols:      cols,
		WireResistance: make([][]float64, rows),
		CellResistance: make([][]float64, rows),
	}

	for i := 0; i < rows; i++ {
		sim.WireResistance[i] = make([]float64, cols)
		sim.CellResistance[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			sim.WireResistance[i][j] = config.WireResistanceOhm
			sim.CellResistance[i][j] = 10000.0 // Default 10k ohm
		}
	}

	return sim
}

// ComputeIRDrop computes IR drop for given input voltages
func (s *IRDropSimulator) ComputeIRDrop(inputVoltages []float64) ([][]float64, float64) {
	// Simplified IR drop computation
	// In practice, would solve Kirchhoff's laws iteratively

	irDropMatrix := make([][]float64, s.ArrayRows)
	maxDrop := 0.0

	for i := 0; i < s.ArrayRows; i++ {
		irDropMatrix[i] = make([]float64, s.ArrayCols)
		for j := 0; j < s.ArrayCols; j++ {
			// IR drop increases with distance from voltage source
			rowDrop := float64(i) * s.Config.WireResistanceOhm * 0.001 // Simplified
			colDrop := float64(j) * s.Config.WireResistanceOhm * 0.001
			totalDrop := rowDrop + colDrop

			if totalDrop > maxDrop {
				maxDrop = totalDrop
			}

			irDropMatrix[i][j] = totalDrop
		}
	}

	return irDropMatrix, maxDrop
}

// =============================================================================
// DATAFLOW SCHEDULER
// =============================================================================

// DataflowScheduler schedules operations across CIM tiles
type DataflowScheduler struct {
	Config          *DataflowConfig
	Hardware        *CIMHardwareSpec
	Schedule        *ExecutionSchedule
	ResourceTracker *ResourceTracker
}

// DataflowConfig configures dataflow scheduling
type DataflowConfig struct {
	Strategy         string // "weight_stationary", "output_stationary", "row_stationary", "hybrid"
	PipeliningDepth  int
	BufferSizeKB     float64
	PrefetchEnabled  bool
	RecomputeEnabled bool
}

// ExecutionSchedule represents the execution schedule
type ExecutionSchedule struct {
	Stages          [][]*ScheduledOp
	TotalCycles     int64
	PeakMemoryKB    float64
	EnergyBreakdown *EnergyBreakdown
	UtilizationMap  map[int][]float64 // Tile ID -> utilization over time
}

// ScheduledOp represents a scheduled operation
type ScheduledOp struct {
	NodeID         int
	TileID         int
	StartCycle     int64
	EndCycle       int64
	Dataflow       string
	InputTensors   []int
	OutputTensor   int
	MemoryReads    int64
	MemoryWrites   int64
	ComputeOps     int64
}

// EnergyBreakdown shows energy consumption breakdown
type EnergyBreakdown struct {
	ComputeEnergyMJ  float64
	MemoryEnergyMJ   float64
	InterconnectMJ   float64
	LeakageEnergyMJ  float64
	TotalEnergyMJ    float64
}

// ResourceTracker tracks resource utilization
type ResourceTracker struct {
	TileStatus      []TileStatus
	MemoryStatus    []MemoryBankStatus
	InterconnectBW  float64
}

// TileStatus tracks tile usage
type TileStatus struct {
	TileID       int
	InUse        bool
	CurrentOp    int
	Utilization  float64
	Temperature  float64
}

// MemoryBankStatus tracks memory bank usage
type MemoryBankStatus struct {
	BankID       int
	UsedKB       float64
	TotalKB      float64
	ReadBW       float64
	WriteBW      float64
}

// NewDataflowScheduler creates a dataflow scheduler
func NewDataflowScheduler(config *DataflowConfig, hw *CIMHardwareSpec) *DataflowScheduler {
	scheduler := &DataflowScheduler{
		Config:   config,
		Hardware: hw,
		Schedule: &ExecutionSchedule{
			Stages:          make([][]*ScheduledOp, 0),
			TotalCycles:     0,
			PeakMemoryKB:    0,
			EnergyBreakdown: &EnergyBreakdown{},
			UtilizationMap:  make(map[int][]float64),
		},
		ResourceTracker: &ResourceTracker{
			TileStatus:   make([]TileStatus, hw.NumTiles),
			MemoryStatus: make([]MemoryBankStatus, hw.MemoryBanks),
		},
	}

	// Initialize resource tracker
	for i := 0; i < hw.NumTiles; i++ {
		scheduler.ResourceTracker.TileStatus[i] = TileStatus{
			TileID:      i,
			InUse:       false,
			CurrentOp:   -1,
			Utilization: 0.0,
			Temperature: 25.0,
		}
	}

	for i := 0; i < hw.MemoryBanks; i++ {
		scheduler.ResourceTracker.MemoryStatus[i] = MemoryBankStatus{
			BankID:  i,
			UsedKB:  0,
			TotalKB: hw.MemoryPerBankKB,
		}
	}

	return scheduler
}

// ScheduleGraph schedules a compute graph
func (s *DataflowScheduler) ScheduleGraph(graph *CIMComputeGraph) error {
	// Topological sort
	sortedNodes := s.topologicalSort(graph)

	// Schedule each node
	currentCycle := int64(0)
	currentStage := make([]*ScheduledOp, 0)

	for _, nodeID := range sortedNodes {
		node := graph.Nodes[nodeID]
		if node.OpType == "input" || node.OpType == "output" {
			continue
		}

		// Find available tile
		tileID := s.findAvailableTile()
		if tileID < 0 {
			// No tile available, start new stage
			if len(currentStage) > 0 {
				s.Schedule.Stages = append(s.Schedule.Stages, currentStage)
				currentStage = make([]*ScheduledOp, 0)
				s.resetTiles()
				tileID = s.findAvailableTile()
			}
		}

		// Schedule operation
		duration := s.estimateOpDuration(node)
		op := &ScheduledOp{
			NodeID:       nodeID,
			TileID:       tileID,
			StartCycle:   currentCycle,
			EndCycle:     currentCycle + duration,
			Dataflow:     s.Config.Strategy,
			ComputeOps:   node.ComputeOps,
			MemoryReads:  node.MemoryBytes,
			MemoryWrites: int64(product(node.OutputShape) * 4),
		}

		currentStage = append(currentStage, op)
		s.ResourceTracker.TileStatus[tileID].InUse = true
		s.ResourceTracker.TileStatus[tileID].CurrentOp = nodeID
	}

	// Add final stage
	if len(currentStage) > 0 {
		s.Schedule.Stages = append(s.Schedule.Stages, currentStage)
	}

	// Compute total cycles
	s.computeTotalCycles()

	// Compute energy breakdown
	s.computeEnergyBreakdown()

	return nil
}

func (s *DataflowScheduler) topologicalSort(graph *CIMComputeGraph) []int {
	// Simple topological sort based on node IDs (assuming sequential dependencies)
	sorted := make([]int, len(graph.Nodes))
	for i := range sorted {
		sorted[i] = i
	}
	return sorted
}

func (s *DataflowScheduler) findAvailableTile() int {
	for i, status := range s.ResourceTracker.TileStatus {
		if !status.InUse {
			return i
		}
	}
	return -1
}

func (s *DataflowScheduler) resetTiles() {
	for i := range s.ResourceTracker.TileStatus {
		s.ResourceTracker.TileStatus[i].InUse = false
		s.ResourceTracker.TileStatus[i].CurrentOp = -1
	}
}

func (s *DataflowScheduler) estimateOpDuration(node *CIMGraphNode) int64 {
	ops := node.ComputeOps
	if ops == 0 {
		ops = 1000000
	}
	opsPerCycle := int64(s.Hardware.TileRows * s.Hardware.TileCols)
	return ops / opsPerCycle + 1
}

func (s *DataflowScheduler) computeTotalCycles() {
	maxEnd := int64(0)
	for _, stage := range s.Schedule.Stages {
		for _, op := range stage {
			if op.EndCycle > maxEnd {
				maxEnd = op.EndCycle
			}
		}
	}
	s.Schedule.TotalCycles = maxEnd
}

func (s *DataflowScheduler) computeEnergyBreakdown() {
	totalCompute := int64(0)
	totalMemRead := int64(0)
	totalMemWrite := int64(0)

	for _, stage := range s.Schedule.Stages {
		for _, op := range stage {
			totalCompute += op.ComputeOps
			totalMemRead += op.MemoryReads
			totalMemWrite += op.MemoryWrites
		}
	}

	// Energy per operation (simplified)
	s.Schedule.EnergyBreakdown.ComputeEnergyMJ = float64(totalCompute) * s.Hardware.EnergyPerMACpJ * 1e-9
	s.Schedule.EnergyBreakdown.MemoryEnergyMJ = float64(totalMemRead+totalMemWrite) * 0.5e-9 // 0.5 pJ/byte
	s.Schedule.EnergyBreakdown.InterconnectMJ = float64(s.Schedule.TotalCycles) * 0.001e-9  // 1 fJ/cycle
	s.Schedule.EnergyBreakdown.LeakageEnergyMJ = float64(s.Schedule.TotalCycles) * float64(s.Hardware.NumTiles) * 0.01e-9

	s.Schedule.EnergyBreakdown.TotalEnergyMJ = s.Schedule.EnergyBreakdown.ComputeEnergyMJ +
		s.Schedule.EnergyBreakdown.MemoryEnergyMJ +
		s.Schedule.EnergyBreakdown.InterconnectMJ +
		s.Schedule.EnergyBreakdown.LeakageEnergyMJ
}

// Helper function
func product(shape []int) int {
	if len(shape) == 0 {
		return 0
	}
	p := 1
	for _, s := range shape {
		p *= s
	}
	return p
}

// =============================================================================
// EDGE DEPLOYMENT OPTIMIZER
// =============================================================================

// EdgeDeploymentOptimizer optimizes models for edge deployment
type EdgeDeploymentOptimizer struct {
	Config           *EdgeDeploymentConfig
	OriginalModel    *EdgeNeuralNetwork
	OptimizedModel   *EdgeNeuralNetwork
	Optimizations    []AppliedOptimization
	PerformanceGains *PerformanceGains
}

// AppliedOptimization records an applied optimization
type AppliedOptimization struct {
	Name             string
	Type             string // "pruning", "quantization", "distillation", "arch_search"
	TargetLayers     []string
	CompressionRatio float64
	AccuracyImpact   float64
	SpeedupFactor    float64
	EnergyReduction  float64
}

// PerformanceGains tracks optimization gains
type PerformanceGains struct {
	OriginalSizeKB    float64
	OptimizedSizeKB   float64
	SizeReduction     float64
	OriginalLatencyMs float64
	OptimizedLatencyMs float64
	LatencyReduction  float64
	OriginalEnergyMJ  float64
	OptimizedEnergyMJ float64
	EnergyReduction   float64
	AccuracyDelta     float64
}

// NewEdgeDeploymentOptimizer creates edge deployment optimizer
func NewEdgeDeploymentOptimizer(config *EdgeDeploymentConfig) *EdgeDeploymentOptimizer {
	return &EdgeDeploymentOptimizer{
		Config:           config,
		OriginalModel:    nil,
		OptimizedModel:   nil,
		Optimizations:    make([]AppliedOptimization, 0),
		PerformanceGains: &PerformanceGains{},
	}
}

// OptimizeForDevice optimizes model for target edge device
func (o *EdgeDeploymentOptimizer) OptimizeForDevice(model *EdgeNeuralNetwork) (*EdgeNeuralNetwork, error) {
	o.OriginalModel = model
	o.recordOriginalMetrics()

	// Apply optimizations based on device constraints
	optimized := o.cloneModel(model)

	// 1. Quantization (always applied for edge)
	optimized = o.applyQuantization(optimized)

	// 2. Pruning if memory constrained
	if o.needsPruning() {
		optimized = o.applyPruning(optimized)
	}

	// 3. Knowledge distillation if accuracy critical
	if o.Config.AccuracyThreshold > 0.95 {
		optimized = o.applyDistillation(optimized)
	}

	o.OptimizedModel = optimized
	o.recordOptimizedMetrics()

	return optimized, nil
}

func (o *EdgeDeploymentOptimizer) recordOriginalMetrics() {
	if o.OriginalModel == nil {
		return
	}
	o.PerformanceGains.OriginalSizeKB = o.OriginalModel.MemoryFootprintKB
	o.PerformanceGains.OriginalLatencyMs = float64(o.OriginalModel.TotalOps) / 1e9 // Simplified
	o.PerformanceGains.OriginalEnergyMJ = float64(o.OriginalModel.TotalOps) * 0.5e-9
}

func (o *EdgeDeploymentOptimizer) recordOptimizedMetrics() {
	if o.OptimizedModel == nil {
		return
	}
	o.PerformanceGains.OptimizedSizeKB = o.OptimizedModel.MemoryFootprintKB
	o.PerformanceGains.OptimizedLatencyMs = float64(o.OptimizedModel.TotalOps) / 1e9
	o.PerformanceGains.OptimizedEnergyMJ = float64(o.OptimizedModel.TotalOps) * 0.5e-9

	// Compute reductions
	if o.PerformanceGains.OriginalSizeKB > 0 {
		o.PerformanceGains.SizeReduction = 1.0 - o.PerformanceGains.OptimizedSizeKB/o.PerformanceGains.OriginalSizeKB
	}
	if o.PerformanceGains.OriginalLatencyMs > 0 {
		o.PerformanceGains.LatencyReduction = 1.0 - o.PerformanceGains.OptimizedLatencyMs/o.PerformanceGains.OriginalLatencyMs
	}
	if o.PerformanceGains.OriginalEnergyMJ > 0 {
		o.PerformanceGains.EnergyReduction = 1.0 - o.PerformanceGains.OptimizedEnergyMJ/o.PerformanceGains.OriginalEnergyMJ
	}
}

func (o *EdgeDeploymentOptimizer) cloneModel(model *EdgeNeuralNetwork) *EdgeNeuralNetwork {
	cloned := &EdgeNeuralNetwork{
		Layers:           make([]*EdgeNNLayer, len(model.Layers)),
		TotalParams:      model.TotalParams,
		TotalOps:         model.TotalOps,
		MemoryFootprintKB: model.MemoryFootprintKB,
	}
	for i, layer := range model.Layers {
		cloned.Layers[i] = &EdgeNNLayer{
			Name:        layer.Name,
			Type:        layer.Type,
			InputShape:  append([]int{}, layer.InputShape...),
			OutputShape: append([]int{}, layer.OutputShape...),
			Weights:     append([]float64{}, layer.Weights...),
			Bias:        append([]float64{}, layer.Bias...),
			NumParams:   layer.NumParams,
			NumOps:      layer.NumOps,
			Quantized:   layer.Quantized,
			QuantBits:   layer.QuantBits,
		}
	}
	return cloned
}

func (o *EdgeDeploymentOptimizer) applyQuantization(model *EdgeNeuralNetwork) *EdgeNeuralNetwork {
	// Determine quantization bits based on device
	bits := 8
	if o.Config.PowerBudgetMW < 1.0 {
		bits = 4 // Ultra-low power needs aggressive quantization
	} else if o.Config.AccuracyThreshold > 0.95 {
		bits = 8 // High accuracy needs higher precision
	}

	for _, layer := range model.Layers {
		layer.Quantized = true
		layer.QuantBits = bits

		// Quantize weights
		scale := math.Pow(2, float64(bits-1)) - 1
		for i, w := range layer.Weights {
			quantized := math.Round(w * scale) / scale
			layer.Weights[i] = quantized
		}
	}

	// Update memory footprint
	compressionRatio := 32.0 / float64(bits)
	model.MemoryFootprintKB /= compressionRatio

	o.Optimizations = append(o.Optimizations, AppliedOptimization{
		Name:             fmt.Sprintf("INT%d Quantization", bits),
		Type:             "quantization",
		CompressionRatio: compressionRatio,
		AccuracyImpact:   -0.01 * (8.0 - float64(bits)), // ~1% per bit reduced
		SpeedupFactor:    1.5,
		EnergyReduction:  0.3,
	})

	return model
}

func (o *EdgeDeploymentOptimizer) needsPruning() bool {
	if o.OriginalModel == nil {
		return false
	}
	return o.OriginalModel.MemoryFootprintKB > o.Config.MemoryBudgetKB
}

func (o *EdgeDeploymentOptimizer) applyPruning(model *EdgeNeuralNetwork) *EdgeNeuralNetwork {
	// Target sparsity based on memory constraint
	targetSize := o.Config.MemoryBudgetKB
	currentSize := model.MemoryFootprintKB
	targetSparsity := 1.0 - targetSize/currentSize
	if targetSparsity < 0 {
		targetSparsity = 0
	}
	if targetSparsity > 0.9 {
		targetSparsity = 0.9 // Max 90% sparsity
	}

	for _, layer := range model.Layers {
		// Magnitude-based pruning
		threshold := o.computePruningThreshold(layer.Weights, targetSparsity)
		prunedCount := 0
		for i, w := range layer.Weights {
			if math.Abs(w) < threshold {
				layer.Weights[i] = 0
				prunedCount++
			}
		}
		actualSparsity := float64(prunedCount) / float64(len(layer.Weights))
		layer.NumParams = int64(float64(layer.NumParams) * (1 - actualSparsity))
	}

	// Update model metrics
	model.TotalParams = int64(float64(model.TotalParams) * (1 - targetSparsity))
	model.MemoryFootprintKB *= (1 - targetSparsity)

	o.Optimizations = append(o.Optimizations, AppliedOptimization{
		Name:             fmt.Sprintf("%.0f%% Sparsity Pruning", targetSparsity*100),
		Type:             "pruning",
		CompressionRatio: 1.0 / (1 - targetSparsity),
		AccuracyImpact:   -targetSparsity * 0.05, // ~5% accuracy loss per 100% sparsity
		SpeedupFactor:    1.0 / (1 - targetSparsity*0.5),
		EnergyReduction:  targetSparsity * 0.4,
	})

	return model
}

func (o *EdgeDeploymentOptimizer) computePruningThreshold(weights []float64, sparsity float64) float64 {
	if len(weights) == 0 {
		return 0
	}

	// Sort absolute values
	absWeights := make([]float64, len(weights))
	for i, w := range weights {
		absWeights[i] = math.Abs(w)
	}
	sort.Float64s(absWeights)

	// Find threshold at target percentile
	idx := int(sparsity * float64(len(absWeights)))
	if idx >= len(absWeights) {
		idx = len(absWeights) - 1
	}
	return absWeights[idx]
}

func (o *EdgeDeploymentOptimizer) applyDistillation(model *EdgeNeuralNetwork) *EdgeNeuralNetwork {
	// Knowledge distillation: smaller model trained to match teacher
	// Simulated: reduce model size while maintaining accuracy

	compressionFactor := 0.75 // 25% size reduction
	model.TotalParams = int64(float64(model.TotalParams) * compressionFactor)
	model.TotalOps = int64(float64(model.TotalOps) * compressionFactor)
	model.MemoryFootprintKB *= compressionFactor

	o.Optimizations = append(o.Optimizations, AppliedOptimization{
		Name:             "Knowledge Distillation",
		Type:             "distillation",
		CompressionRatio: 1.0 / compressionFactor,
		AccuracyImpact:   -0.005, // Only 0.5% accuracy loss with distillation
		SpeedupFactor:    1.0 / compressionFactor,
		EnergyReduction:  1 - compressionFactor,
	})

	return model
}

// GetDeploymentReport generates deployment readiness report
func (o *EdgeDeploymentOptimizer) GetDeploymentReport() *DeploymentReport {
	report := &DeploymentReport{
		DeviceType:         o.Config.DeviceType,
		MeetsConstraints:   true,
		ConstraintStatus:   make(map[string]ConstraintStatus),
		Optimizations:      o.Optimizations,
		PerformanceGains:   o.PerformanceGains,
		Recommendations:    make([]string, 0),
	}

	// Check power constraint
	powerStatus := ConstraintStatus{
		Name:     "Power Budget",
		Target:   o.Config.PowerBudgetMW,
		Actual:   o.PerformanceGains.OptimizedEnergyMJ * 1000 / o.PerformanceGains.OptimizedLatencyMs,
		Unit:     "mW",
		Met:      false,
	}
	powerStatus.Met = powerStatus.Actual <= powerStatus.Target
	report.ConstraintStatus["power"] = powerStatus
	if !powerStatus.Met {
		report.MeetsConstraints = false
		report.Recommendations = append(report.Recommendations, "Consider more aggressive quantization or pruning to reduce power")
	}

	// Check memory constraint
	memStatus := ConstraintStatus{
		Name:   "Memory Budget",
		Target: o.Config.MemoryBudgetKB,
		Actual: o.PerformanceGains.OptimizedSizeKB,
		Unit:   "KB",
		Met:    false,
	}
	memStatus.Met = memStatus.Actual <= memStatus.Target
	report.ConstraintStatus["memory"] = memStatus
	if !memStatus.Met {
		report.MeetsConstraints = false
		report.Recommendations = append(report.Recommendations, "Model exceeds memory budget - apply pruning or use smaller architecture")
	}

	// Check latency constraint
	latStatus := ConstraintStatus{
		Name:   "Latency Budget",
		Target: o.Config.LatencyBudgetMs,
		Actual: o.PerformanceGains.OptimizedLatencyMs,
		Unit:   "ms",
		Met:    false,
	}
	latStatus.Met = latStatus.Actual <= latStatus.Target
	report.ConstraintStatus["latency"] = latStatus
	if !latStatus.Met {
		report.MeetsConstraints = false
		report.Recommendations = append(report.Recommendations, "Latency exceeds budget - consider layer fusion or architecture optimization")
	}

	return report
}

// DeploymentReport summarizes deployment readiness
type DeploymentReport struct {
	DeviceType       EdgeDeviceType
	MeetsConstraints bool
	ConstraintStatus map[string]ConstraintStatus
	Optimizations    []AppliedOptimization
	PerformanceGains *PerformanceGains
	Recommendations  []string
}

// ConstraintStatus tracks individual constraint status
type ConstraintStatus struct {
	Name   string
	Target float64
	Actual float64
	Unit   string
	Met    bool
}

// =============================================================================
// EDGE DEPLOYMENT CASE STUDY RUNNER
// =============================================================================

// EdgeCaseStudyRunner runs edge deployment case studies
type EdgeCaseStudyRunner struct {
	CaseStudies []*EdgeCaseStudy
	Results     []*CaseStudyResult
}

// EdgeCaseStudy represents an edge deployment case study
type EdgeCaseStudy struct {
	Name           string
	Description    string
	DeviceType     EdgeDeviceType
	Application    string
	ModelArch      string
	DatasetSize    int
	InputShape     []int
	TargetMetrics  map[string]float64
}

// CaseStudyResult holds case study results
type CaseStudyResult struct {
	CaseStudy        *EdgeCaseStudy
	Success          bool
	DeployedModel    *EdgeNeuralNetwork
	Report           *DeploymentReport
	ActualMetrics    map[string]float64
	ExecutionTimeMs  float64
}

// NewEdgeCaseStudyRunner creates case study runner
func NewEdgeCaseStudyRunner() *EdgeCaseStudyRunner {
	runner := &EdgeCaseStudyRunner{
		CaseStudies: make([]*EdgeCaseStudy, 0),
		Results:     make([]*CaseStudyResult, 0),
	}

	// Add predefined case studies
	runner.CaseStudies = append(runner.CaseStudies,
		&EdgeCaseStudy{
			Name:          "Smart Agriculture Sensor",
			Description:   "Crop health monitoring using leaf image classification",
			DeviceType:    EdgeDeviceIoTSensor,
			Application:   "agriculture",
			ModelArch:     "MobileNetV3-Small",
			DatasetSize:   10000,
			InputShape:    []int{1, 96, 96, 3},
			TargetMetrics: map[string]float64{"accuracy": 0.85, "latency_ms": 100, "energy_mj": 0.001},
		},
		&EdgeCaseStudy{
			Name:          "ECG Arrhythmia Detection",
			Description:   "Real-time arrhythmia detection on smartwatch",
			DeviceType:    EdgeDeviceWearable,
			Application:   "healthcare",
			ModelArch:     "1D-CNN",
			DatasetSize:   50000,
			InputShape:    []int{1, 256, 1},
			TargetMetrics: map[string]float64{"accuracy": 0.95, "latency_ms": 50, "energy_mj": 0.01},
		},
		&EdgeCaseStudy{
			Name:          "Pedestrian Detection ADAS",
			Description:   "Safety-critical pedestrian detection for autonomous driving",
			DeviceType:    EdgeDeviceAutomotive,
			Application:   "automotive",
			ModelArch:     "YOLO-CIM",
			DatasetSize:   100000,
			InputShape:    []int{1, 416, 416, 3},
			TargetMetrics: map[string]float64{"mAP": 0.85, "latency_ms": 10, "fps": 30},
		},
		&EdgeCaseStudy{
			Name:          "Voice Command Recognition",
			Description:   "Always-on keyword spotting for smart home",
			DeviceType:    EdgeDeviceSmartHome,
			Application:   "smart_home",
			ModelArch:     "DS-CNN",
			DatasetSize:   30000,
			InputShape:    []int{1, 40, 49, 1},
			TargetMetrics: map[string]float64{"accuracy": 0.92, "latency_ms": 200, "energy_mj": 0.1},
		},
		&EdgeCaseStudy{
			Name:          "Vibration Anomaly Detection",
			Description:   "Predictive maintenance for industrial machinery",
			DeviceType:    EdgeDeviceIndustrial,
			Application:   "industrial",
			ModelArch:     "Autoencoder",
			DatasetSize:   20000,
			InputShape:    []int{1, 512, 3},
			TargetMetrics: map[string]float64{"f1_score": 0.90, "latency_ms": 5, "energy_mj": 0.5},
		},
	)

	return runner
}

// RunAllCaseStudies executes all case studies
func (r *EdgeCaseStudyRunner) RunAllCaseStudies() {
	for _, cs := range r.CaseStudies {
		result := r.runCaseStudy(cs)
		r.Results = append(r.Results, result)
	}
}

func (r *EdgeCaseStudyRunner) runCaseStudy(cs *EdgeCaseStudy) *CaseStudyResult {
	result := &CaseStudyResult{
		CaseStudy:     cs,
		Success:       false,
		ActualMetrics: make(map[string]float64),
	}

	// Create model for case study
	model := r.createModelForCaseStudy(cs)

	// Create optimizer for target device
	config := NewEdgeDeploymentConfig(cs.DeviceType)
	optimizer := NewEdgeDeploymentOptimizer(config)

	// Optimize model
	optimized, err := optimizer.OptimizeForDevice(model)
	if err != nil {
		return result
	}

	result.DeployedModel = optimized
	result.Report = optimizer.GetDeploymentReport()
	result.Success = result.Report.MeetsConstraints

	// Record actual metrics
	result.ActualMetrics["size_kb"] = optimized.MemoryFootprintKB
	result.ActualMetrics["params"] = float64(optimized.TotalParams)
	result.ActualMetrics["ops"] = float64(optimized.TotalOps)

	return result
}

func (r *EdgeCaseStudyRunner) createModelForCaseStudy(cs *EdgeCaseStudy) *EdgeNeuralNetwork {
	// Create simplified model based on architecture
	model := &EdgeNeuralNetwork{
		Layers: make([]*EdgeNNLayer, 0),
	}

	inputSize := 1
	for _, s := range cs.InputShape {
		inputSize *= s
	}

	// Add layers based on architecture type
	switch cs.ModelArch {
	case "MobileNetV3-Small":
		model.Layers = append(model.Layers,
			&EdgeNNLayer{Name: "conv1", Type: "conv2d", NumParams: 16 * 3 * 3 * 3, NumOps: 16 * 96 * 96 * 3 * 3 * 3},
			&EdgeNNLayer{Name: "bneck1", Type: "inverted_residual", NumParams: 16 * 16 * 3, NumOps: 16 * 48 * 48 * 16 * 3},
			&EdgeNNLayer{Name: "bneck2", Type: "inverted_residual", NumParams: 24 * 72 * 3, NumOps: 24 * 24 * 24 * 72 * 3},
			&EdgeNNLayer{Name: "fc", Type: "dense", NumParams: 576 * 1024, NumOps: 576 * 1024},
		)
		model.TotalParams = 2900000
		model.TotalOps = 56000000
		model.MemoryFootprintKB = float64(model.TotalParams*4) / 1024

	case "1D-CNN":
		model.Layers = append(model.Layers,
			&EdgeNNLayer{Name: "conv1", Type: "conv1d", NumParams: 32 * 5 * 1, NumOps: 32 * 256 * 5},
			&EdgeNNLayer{Name: "conv2", Type: "conv1d", NumParams: 64 * 5 * 32, NumOps: 64 * 128 * 5 * 32},
			&EdgeNNLayer{Name: "fc", Type: "dense", NumParams: 64 * 64 * 5, NumOps: 64 * 64 * 5},
		)
		model.TotalParams = 150000
		model.TotalOps = 2000000
		model.MemoryFootprintKB = float64(model.TotalParams*4) / 1024

	default:
		// Generic model
		model.Layers = append(model.Layers,
			&EdgeNNLayer{Name: "layer1", Type: "generic", NumParams: int64(inputSize * 64), NumOps: int64(inputSize * 64)},
			&EdgeNNLayer{Name: "layer2", Type: "generic", NumParams: 64 * 32, NumOps: 64 * 32},
			&EdgeNNLayer{Name: "output", Type: "dense", NumParams: 32 * 10, NumOps: 32 * 10},
		)
		model.TotalParams = int64(inputSize*64 + 64*32 + 32*10)
		model.TotalOps = model.TotalParams
		model.MemoryFootprintKB = float64(model.TotalParams*4) / 1024
	}

	return model
}

// GetSummaryReport generates summary across all case studies
func (r *EdgeCaseStudyRunner) GetSummaryReport() string {
	successCount := 0
	for _, result := range r.Results {
		if result.Success {
			successCount++
		}
	}

	return fmt.Sprintf("Edge Deployment Case Studies: %d/%d successful",
		successCount, len(r.Results))
}
