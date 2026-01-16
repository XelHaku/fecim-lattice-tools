// Package layers provides model profiling tools for CIM deployment analysis.
// Implements layer-by-layer profiling, bottleneck detection, and optimization suggestions.
package layers

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ============================================================================
// LAYER PROFILING
// ============================================================================

// LayerType represents different neural network layer types.
type LayerType int

const (
	LayerDense LayerType = iota
	LayerConv2D
	LayerDepthwiseConv
	LayerPooling
	LayerNormalization
	LayerActivation
	LayerAttention
	LayerEmbedding
	LayerRecurrent
	LayerReservoir
)

// String returns layer type name.
func (t LayerType) String() string {
	names := []string{
		"Dense", "Conv2D", "DepthwiseConv", "Pooling",
		"Normalization", "Activation", "Attention",
		"Embedding", "Recurrent", "Reservoir",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return "Unknown"
}

// LayerProfile contains profiling data for a single layer.
type LayerProfile struct {
	Name          string
	Type          LayerType
	Index         int

	// Shape information
	InputShape    []int
	OutputShape   []int
	WeightShape   []int

	// Computation metrics
	MACs          int64   // Multiply-accumulate operations
	Parameters    int64   // Number of parameters
	MemoryBytes   int64   // Memory footprint

	// CIM-specific metrics
	RequiredArrays    int
	ArrayUtilization  float64 // 0-1
	TilingFactor      int     // Number of tiles needed
	SparsityPotential float64 // How much sparsity could help

	// Timing (simulated)
	ComputeTimeNs   float64
	MemoryTimeNs    float64
	TotalTimeNs     float64

	// Energy (simulated)
	ComputeEnergyPJ float64
	MemoryEnergyPJ  float64
	TotalEnergyPJ   float64

	// Bottleneck indicators
	IsComputeBound bool
	IsMemoryBound  bool
	BottleneckScore float64 // Higher = more of a bottleneck
}

// LayerProfiler analyzes individual layers.
type LayerProfiler struct {
	ArrayRows     int
	ArrayCols     int
	ClockFreqGHz  float64
	EnergyPerMAC  float64 // pJ
	MemBandwidth  float64 // GB/s
}

// NewLayerProfiler creates a new profiler.
func NewLayerProfiler(arrayRows, arrayCols int) *LayerProfiler {
	return &LayerProfiler{
		ArrayRows:    arrayRows,
		ArrayCols:    arrayCols,
		ClockFreqGHz: 1.0,
		EnergyPerMAC: 1.0, // 1 pJ per MAC
		MemBandwidth: 100, // 100 GB/s
	}
}

// ProfileDenseLayer profiles a fully connected layer.
func (p *LayerProfiler) ProfileDenseLayer(name string, inputSize, outputSize int) *LayerProfile {
	profile := &LayerProfile{
		Name:        name,
		Type:        LayerDense,
		InputShape:  []int{inputSize},
		OutputShape: []int{outputSize},
		WeightShape: []int{outputSize, inputSize},
	}

	// Compute MACs and parameters
	profile.MACs = int64(inputSize) * int64(outputSize)
	profile.Parameters = int64(inputSize)*int64(outputSize) + int64(outputSize) // weights + bias
	profile.MemoryBytes = profile.Parameters * 4 // float32

	// CIM mapping
	profile.RequiredArrays = int(math.Ceil(float64(outputSize)/float64(p.ArrayCols))) *
		int(math.Ceil(float64(inputSize)/float64(p.ArrayRows)))
	profile.ArrayUtilization = float64(inputSize*outputSize) /
		float64(profile.RequiredArrays*p.ArrayRows*p.ArrayCols)
	profile.TilingFactor = profile.RequiredArrays

	// Sparsity potential (dense layers often sparse after training)
	profile.SparsityPotential = 0.7

	// Timing
	cyclesPerMVM := 1 + 6 // 1 compute + ADC cycles
	profile.ComputeTimeNs = float64(profile.RequiredArrays*cyclesPerMVM) / p.ClockFreqGHz
	profile.MemoryTimeNs = float64(profile.MemoryBytes) / (p.MemBandwidth * 1e9) * 1e9
	profile.TotalTimeNs = math.Max(profile.ComputeTimeNs, profile.MemoryTimeNs)

	// Energy
	profile.ComputeEnergyPJ = float64(profile.MACs) * p.EnergyPerMAC
	profile.MemoryEnergyPJ = float64(profile.MemoryBytes) * 0.1 // 0.1 pJ/byte
	profile.TotalEnergyPJ = profile.ComputeEnergyPJ + profile.MemoryEnergyPJ

	// Bottleneck analysis
	profile.IsComputeBound = profile.ComputeTimeNs > profile.MemoryTimeNs
	profile.IsMemoryBound = !profile.IsComputeBound
	profile.BottleneckScore = profile.TotalTimeNs / 1000 // Normalize to µs

	return profile
}

// ProfileConv2DLayer profiles a 2D convolutional layer.
func (p *LayerProfiler) ProfileConv2DLayer(name string, inC, outC, H, W, K int, stride int) *LayerProfile {
	profile := &LayerProfile{
		Name:        name,
		Type:        LayerConv2D,
		InputShape:  []int{inC, H, W},
		OutputShape: []int{outC, H / stride, W / stride},
		WeightShape: []int{outC, inC, K, K},
	}

	// Output dimensions
	outH := H / stride
	outW := W / stride

	// MACs: output_pixels × kernel_ops
	profile.MACs = int64(outH) * int64(outW) * int64(outC) * int64(inC) * int64(K) * int64(K)
	profile.Parameters = int64(outC) * int64(inC) * int64(K) * int64(K) + int64(outC)
	profile.MemoryBytes = profile.Parameters * 4

	// Im2col mapping to crossbar
	im2colRows := inC * K * K
	im2colCols := outH * outW
	profile.RequiredArrays = int(math.Ceil(float64(im2colRows)/float64(p.ArrayRows))) *
		int(math.Ceil(float64(outC)/float64(p.ArrayCols)))
	profile.ArrayUtilization = float64(im2colRows*outC) /
		float64(profile.RequiredArrays*p.ArrayRows*p.ArrayCols)
	profile.TilingFactor = profile.RequiredArrays

	// Conv layers have moderate sparsity potential
	profile.SparsityPotential = 0.5

	// Timing
	batchesPerArray := int(math.Ceil(float64(im2colCols) / float64(p.ArrayRows)))
	profile.ComputeTimeNs = float64(profile.RequiredArrays*batchesPerArray*7) / p.ClockFreqGHz
	profile.MemoryTimeNs = float64(profile.MemoryBytes) / (p.MemBandwidth * 1e9) * 1e9
	profile.TotalTimeNs = math.Max(profile.ComputeTimeNs, profile.MemoryTimeNs)

	// Energy
	profile.ComputeEnergyPJ = float64(profile.MACs) * p.EnergyPerMAC
	profile.MemoryEnergyPJ = float64(profile.MemoryBytes) * 0.1
	profile.TotalEnergyPJ = profile.ComputeEnergyPJ + profile.MemoryEnergyPJ

	// Bottleneck
	profile.IsComputeBound = profile.ComputeTimeNs > profile.MemoryTimeNs
	profile.IsMemoryBound = !profile.IsComputeBound
	profile.BottleneckScore = profile.TotalTimeNs / 1000

	return profile
}

// ProfileAttentionLayer profiles a multi-head attention layer.
func (p *LayerProfiler) ProfileAttentionLayer(name string, seqLen, dModel, numHeads int) *LayerProfile {
	profile := &LayerProfile{
		Name:        name,
		Type:        LayerAttention,
		InputShape:  []int{seqLen, dModel},
		OutputShape: []int{seqLen, dModel},
		WeightShape: []int{4, dModel, dModel}, // Q, K, V, O projections
	}

	headDim := dModel / numHeads

	// Q, K, V projections: 3 × seqLen × dModel × dModel
	projectionMACs := int64(3) * int64(seqLen) * int64(dModel) * int64(dModel)

	// Attention scores: seqLen × seqLen × headDim × numHeads
	attentionMACs := int64(seqLen) * int64(seqLen) * int64(headDim) * int64(numHeads)

	// Output projection: seqLen × dModel × dModel
	outputMACs := int64(seqLen) * int64(dModel) * int64(dModel)

	profile.MACs = projectionMACs + attentionMACs + outputMACs
	profile.Parameters = 4 * int64(dModel) * int64(dModel)
	profile.MemoryBytes = profile.Parameters*4 + int64(seqLen)*int64(seqLen)*int64(numHeads)*4

	// CIM mapping for attention
	profile.RequiredArrays = 4 * int(math.Ceil(float64(dModel)/float64(p.ArrayCols))) *
		int(math.Ceil(float64(dModel)/float64(p.ArrayRows)))
	profile.ArrayUtilization = 0.6 // Attention has overhead
	profile.TilingFactor = profile.RequiredArrays

	// Attention can benefit from sparse patterns
	profile.SparsityPotential = 0.8

	// Timing (attention is often memory-bound due to score matrix)
	profile.ComputeTimeNs = float64(profile.MACs) / (float64(profile.RequiredArrays) * float64(p.ArrayRows*p.ArrayCols) * p.ClockFreqGHz * 1e9) * 1e9
	profile.MemoryTimeNs = float64(profile.MemoryBytes) / (p.MemBandwidth * 1e9) * 1e9
	profile.TotalTimeNs = math.Max(profile.ComputeTimeNs, profile.MemoryTimeNs)

	// Energy
	profile.ComputeEnergyPJ = float64(profile.MACs) * p.EnergyPerMAC
	profile.MemoryEnergyPJ = float64(profile.MemoryBytes) * 0.1
	profile.TotalEnergyPJ = profile.ComputeEnergyPJ + profile.MemoryEnergyPJ

	// Attention often memory-bound
	profile.IsComputeBound = profile.ComputeTimeNs > profile.MemoryTimeNs
	profile.IsMemoryBound = !profile.IsComputeBound
	profile.BottleneckScore = profile.TotalTimeNs / 1000

	return profile
}

// ============================================================================
// MODEL PROFILING
// ============================================================================

// ModelProfile contains complete model analysis.
type ModelProfile struct {
	Name           string
	Layers         []*LayerProfile
	TotalMACs      int64
	TotalParams    int64
	TotalMemory    int64
	TotalArrays    int
	TotalTimeNs    float64
	TotalEnergyPJ  float64
	Bottlenecks    []*BottleneckInfo
	Recommendations []string
}

// ModelProfiler analyzes complete models.
type ModelProfiler struct {
	LayerProfiler *LayerProfiler
	Profiles      map[string]*ModelProfile
}

// NewModelProfiler creates a new model profiler.
func NewModelProfiler(arrayRows, arrayCols int) *ModelProfiler {
	return &ModelProfiler{
		LayerProfiler: NewLayerProfiler(arrayRows, arrayCols),
		Profiles:      make(map[string]*ModelProfile),
	}
}

// ProfileModel analyzes a model defined by layer specs.
func (p *ModelProfiler) ProfileModel(name string, layers []LayerSpec) *ModelProfile {
	profile := &ModelProfile{
		Name:   name,
		Layers: make([]*LayerProfile, 0, len(layers)),
	}

	for i, spec := range layers {
		var layerProfile *LayerProfile

		switch spec.Type {
		case LayerDense:
			layerProfile = p.LayerProfiler.ProfileDenseLayer(
				spec.Name, spec.InputSize, spec.OutputSize)
		case LayerConv2D:
			layerProfile = p.LayerProfiler.ProfileConv2DLayer(
				spec.Name, spec.InChannels, spec.OutChannels,
				spec.Height, spec.Width, spec.KernelSize, spec.Stride)
		case LayerAttention:
			layerProfile = p.LayerProfiler.ProfileAttentionLayer(
				spec.Name, spec.SeqLen, spec.DModel, spec.NumHeads)
		default:
			// Generic layer
			layerProfile = &LayerProfile{
				Name:  spec.Name,
				Type:  spec.Type,
				Index: i,
			}
		}

		layerProfile.Index = i
		profile.Layers = append(profile.Layers, layerProfile)

		// Accumulate totals
		profile.TotalMACs += layerProfile.MACs
		profile.TotalParams += layerProfile.Parameters
		profile.TotalMemory += layerProfile.MemoryBytes
		profile.TotalArrays += layerProfile.RequiredArrays
		profile.TotalTimeNs += layerProfile.TotalTimeNs
		profile.TotalEnergyPJ += layerProfile.TotalEnergyPJ
	}

	// Identify bottlenecks
	profile.Bottlenecks = p.identifyBottlenecks(profile)

	// Generate recommendations
	profile.Recommendations = p.generateRecommendations(profile)

	p.Profiles[name] = profile
	return profile
}

// LayerSpec defines a layer for profiling.
type LayerSpec struct {
	Name        string
	Type        LayerType
	InputSize   int
	OutputSize  int
	InChannels  int
	OutChannels int
	Height      int
	Width       int
	KernelSize  int
	Stride      int
	SeqLen      int
	DModel      int
	NumHeads    int
}

// ============================================================================
// BOTTLENECK ANALYSIS
// ============================================================================

// BottleneckInfo describes a performance bottleneck.
type BottleneckInfo struct {
	LayerIndex    int
	LayerName     string
	Type          string // "compute", "memory", "array_count", "utilization"
	Severity      float64 // 0-1, higher = worse
	Description   string
	Suggestion    string
}

// identifyBottlenecks finds performance bottlenecks.
func (p *ModelProfiler) identifyBottlenecks(profile *ModelProfile) []*BottleneckInfo {
	bottlenecks := make([]*BottleneckInfo, 0)

	// Sort layers by time contribution
	sortedLayers := make([]*LayerProfile, len(profile.Layers))
	copy(sortedLayers, profile.Layers)
	sort.Slice(sortedLayers, func(i, j int) bool {
		return sortedLayers[i].TotalTimeNs > sortedLayers[j].TotalTimeNs
	})

	// Top 3 time-consuming layers are bottlenecks
	for i := 0; i < min(3, len(sortedLayers)); i++ {
		layer := sortedLayers[i]
		fraction := layer.TotalTimeNs / profile.TotalTimeNs

		if fraction > 0.1 { // More than 10% of time
			bottleneck := &BottleneckInfo{
				LayerIndex: layer.Index,
				LayerName:  layer.Name,
				Severity:   fraction,
			}

			if layer.IsMemoryBound {
				bottleneck.Type = "memory"
				bottleneck.Description = fmt.Sprintf("Layer consumes %.1f%% time, memory-bound", fraction*100)
				bottleneck.Suggestion = "Consider weight quantization or on-chip weight caching"
			} else {
				bottleneck.Type = "compute"
				bottleneck.Description = fmt.Sprintf("Layer consumes %.1f%% time, compute-bound", fraction*100)
				bottleneck.Suggestion = "Consider using more arrays or structured pruning"
			}

			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	// Check for low utilization
	for _, layer := range profile.Layers {
		if layer.ArrayUtilization < 0.5 && layer.RequiredArrays > 0 {
			bottlenecks = append(bottlenecks, &BottleneckInfo{
				LayerIndex:  layer.Index,
				LayerName:   layer.Name,
				Type:        "utilization",
				Severity:    1 - layer.ArrayUtilization,
				Description: fmt.Sprintf("Array utilization only %.1f%%", layer.ArrayUtilization*100),
				Suggestion:  "Consider layer fusion or weight reordering",
			})
		}
	}

	// Check for excessive array count
	maxArrays := 64 // Typical limit
	for _, layer := range profile.Layers {
		if layer.RequiredArrays > maxArrays {
			bottlenecks = append(bottlenecks, &BottleneckInfo{
				LayerIndex:  layer.Index,
				LayerName:   layer.Name,
				Type:        "array_count",
				Severity:    float64(layer.RequiredArrays-maxArrays) / float64(maxArrays),
				Description: fmt.Sprintf("Requires %d arrays, exceeds limit of %d", layer.RequiredArrays, maxArrays),
				Suggestion:  "Use weight tiling with partial sum accumulation",
			})
		}
	}

	return bottlenecks
}

// generateRecommendations creates optimization suggestions.
func (p *ModelProfiler) generateRecommendations(profile *ModelProfile) []string {
	recs := make([]string, 0)

	// Overall model size
	if profile.TotalMemory > 100*1024*1024 { // >100 MB
		recs = append(recs, "Model exceeds 100MB - consider quantization to 8-bit or 4-bit")
	}

	// Sparsity potential
	totalSparsityPotential := 0.0
	for _, layer := range profile.Layers {
		totalSparsityPotential += layer.SparsityPotential * float64(layer.MACs)
	}
	avgSparsity := totalSparsityPotential / float64(profile.TotalMACs)
	if avgSparsity > 0.5 {
		recs = append(recs, fmt.Sprintf("High sparsity potential (%.0f%%) - consider structured pruning", avgSparsity*100))
	}

	// Memory bandwidth
	memoryFraction := 0.0
	for _, layer := range profile.Layers {
		if layer.IsMemoryBound {
			memoryFraction += layer.TotalTimeNs
		}
	}
	memoryFraction /= profile.TotalTimeNs
	if memoryFraction > 0.5 {
		recs = append(recs, fmt.Sprintf("%.0f%% time memory-bound - consider on-chip weight caching", memoryFraction*100))
	}

	// Attention layers
	attentionCount := 0
	for _, layer := range profile.Layers {
		if layer.Type == LayerAttention {
			attentionCount++
		}
	}
	if attentionCount > 0 {
		recs = append(recs, fmt.Sprintf("%d attention layers - consider sparse attention patterns (sliding window, block sparse)", attentionCount))
	}

	// Array efficiency
	avgUtilization := 0.0
	for _, layer := range profile.Layers {
		avgUtilization += layer.ArrayUtilization
	}
	avgUtilization /= float64(len(profile.Layers))
	if avgUtilization < 0.6 {
		recs = append(recs, fmt.Sprintf("Low average array utilization (%.0f%%) - consider layer fusion", avgUtilization*100))
	}

	return recs
}

// ============================================================================
// PROFILING REPORT
// ============================================================================

// GenerateReport creates a detailed profiling report.
func (profile *ModelProfile) GenerateReport() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("=== Model Profile: %s ===\n\n", profile.Name))

	// Summary
	sb.WriteString("--- Summary ---\n")
	sb.WriteString(fmt.Sprintf("Total MACs:      %.2f M\n", float64(profile.TotalMACs)/1e6))
	sb.WriteString(fmt.Sprintf("Total Params:    %.2f M\n", float64(profile.TotalParams)/1e6))
	sb.WriteString(fmt.Sprintf("Total Memory:    %.2f MB\n", float64(profile.TotalMemory)/1e6))
	sb.WriteString(fmt.Sprintf("Total Arrays:    %d\n", profile.TotalArrays))
	sb.WriteString(fmt.Sprintf("Est. Latency:    %.2f µs\n", profile.TotalTimeNs/1000))
	sb.WriteString(fmt.Sprintf("Est. Energy:     %.2f µJ\n\n", profile.TotalEnergyPJ/1e6))

	// Layer breakdown
	sb.WriteString("--- Layer Breakdown ---\n")
	sb.WriteString(fmt.Sprintf("%-20s %-12s %10s %10s %8s %8s\n",
		"Layer", "Type", "MACs", "Params", "Arrays", "Time%"))
	sb.WriteString(strings.Repeat("-", 70) + "\n")

	for _, layer := range profile.Layers {
		timePct := layer.TotalTimeNs / profile.TotalTimeNs * 100
		sb.WriteString(fmt.Sprintf("%-20s %-12s %10.2fK %10.2fK %8d %7.1f%%\n",
			truncate(layer.Name, 20),
			layer.Type.String(),
			float64(layer.MACs)/1000,
			float64(layer.Parameters)/1000,
			layer.RequiredArrays,
			timePct))
	}
	sb.WriteString("\n")

	// Bottlenecks
	if len(profile.Bottlenecks) > 0 {
		sb.WriteString("--- Bottlenecks ---\n")
		for _, b := range profile.Bottlenecks {
			sb.WriteString(fmt.Sprintf("[%s] %s (severity: %.1f%%)\n",
				b.Type, b.LayerName, b.Severity*100))
			sb.WriteString(fmt.Sprintf("  → %s\n", b.Description))
			sb.WriteString(fmt.Sprintf("  💡 %s\n", b.Suggestion))
		}
		sb.WriteString("\n")
	}

	// Recommendations
	if len(profile.Recommendations) > 0 {
		sb.WriteString("--- Recommendations ---\n")
		for i, rec := range profile.Recommendations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, rec))
		}
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// RUNTIME PROFILER
// ============================================================================

// RuntimeProfiler measures actual execution times.
type RuntimeProfiler struct {
	Measurements map[string][]time.Duration
	StartTimes   map[string]time.Time
}

// NewRuntimeProfiler creates a new runtime profiler.
func NewRuntimeProfiler() *RuntimeProfiler {
	return &RuntimeProfiler{
		Measurements: make(map[string][]time.Duration),
		StartTimes:   make(map[string]time.Time),
	}
}

// Start begins timing a section.
func (p *RuntimeProfiler) Start(name string) {
	p.StartTimes[name] = time.Now()
}

// Stop ends timing and records the duration.
func (p *RuntimeProfiler) Stop(name string) time.Duration {
	if start, ok := p.StartTimes[name]; ok {
		duration := time.Since(start)
		p.Measurements[name] = append(p.Measurements[name], duration)
		delete(p.StartTimes, name)
		return duration
	}
	return 0
}

// GetStats returns timing statistics.
func (p *RuntimeProfiler) GetStats(name string) (avg, min, max time.Duration) {
	measurements := p.Measurements[name]
	if len(measurements) == 0 {
		return 0, 0, 0
	}

	var total time.Duration
	minDur := measurements[0]
	maxDur := measurements[0]

	for _, d := range measurements {
		total += d
		if d < minDur {
			minDur = d
		}
		if d > maxDur {
			maxDur = d
		}
	}

	return total / time.Duration(len(measurements)), minDur, maxDur
}

// Summary returns a summary of all measurements.
func (p *RuntimeProfiler) Summary() string {
	var sb strings.Builder
	sb.WriteString("=== Runtime Profile ===\n")
	sb.WriteString(fmt.Sprintf("%-30s %12s %12s %12s %8s\n",
		"Section", "Avg", "Min", "Max", "Count"))
	sb.WriteString(strings.Repeat("-", 76) + "\n")

	for name := range p.Measurements {
		avg, minD, maxD := p.GetStats(name)
		count := len(p.Measurements[name])
		sb.WriteString(fmt.Sprintf("%-30s %12v %12v %12v %8d\n",
			truncate(name, 30), avg, minD, maxD, count))
	}

	return sb.String()
}

// ============================================================================
// IN-SENSOR COMPUTING PROFILE
// ============================================================================

// InSensorProfile contains in-sensor computing analysis.
type InSensorProfile struct {
	SensorType      string
	Resolution      [2]int
	FrameRate       float64
	EventRate       float64
	ProcessingMode  string // "frame", "event", "hybrid"

	// Power breakdown
	SensorPowerMW   float64
	ProcessingPowerMW float64
	CommunicationPowerMW float64
	TotalPowerMW    float64

	// Latency
	SensorLatencyUs  float64
	ProcessingLatencyUs float64
	TotalLatencyUs   float64

	// Data rates
	RawDataRateMbps     float64
	ProcessedDataRateMbps float64
	CompressionRatio    float64
}

// ProfileInSensorSystem analyzes an in-sensor computing system.
func ProfileInSensorSystem(resolution [2]int, frameRate float64, featureExtraction bool) *InSensorProfile {
	profile := &InSensorProfile{
		SensorType: "FeFET Image Sensor",
		Resolution: resolution,
		FrameRate:  frameRate,
	}

	pixels := resolution[0] * resolution[1]

	if featureExtraction {
		profile.ProcessingMode = "hybrid"
		// Event-based with in-sensor processing
		profile.EventRate = float64(pixels) * frameRate * 0.1 // 10% pixel change rate

		// Power
		profile.SensorPowerMW = float64(pixels) * 0.005   // 5 µW per pixel
		profile.ProcessingPowerMW = float64(pixels) * 0.01 // 10 µW per pixel for CIM
		profile.CommunicationPowerMW = profile.EventRate * 8 * 1e-6 // Event transmission

		// Latency (event processing is faster)
		profile.SensorLatencyUs = 10.0
		profile.ProcessingLatencyUs = 5.0 // In-sensor is fast

		// Data rates
		profile.RawDataRateMbps = float64(pixels) * frameRate * 8 / 1e6
		profile.ProcessedDataRateMbps = profile.EventRate * 64 / 1e6 // Features only
		profile.CompressionRatio = profile.RawDataRateMbps / profile.ProcessedDataRateMbps

	} else {
		profile.ProcessingMode = "frame"
		// Traditional frame-based

		// Power
		profile.SensorPowerMW = float64(pixels) * 0.01 // 10 µW per pixel
		profile.CommunicationPowerMW = float64(pixels) * frameRate * 8 / 1e6 * 10 // 10 pJ/bit

		// Latency
		profile.SensorLatencyUs = 1000 / frameRate * 1000 // Frame period
		profile.ProcessingLatencyUs = 0                   // Off-chip

		// Data rates
		profile.RawDataRateMbps = float64(pixels) * frameRate * 8 / 1e6
		profile.ProcessedDataRateMbps = profile.RawDataRateMbps
		profile.CompressionRatio = 1.0
	}

	profile.TotalPowerMW = profile.SensorPowerMW + profile.ProcessingPowerMW + profile.CommunicationPowerMW
	profile.TotalLatencyUs = profile.SensorLatencyUs + profile.ProcessingLatencyUs

	return profile
}

// Report generates an in-sensor profile report.
func (p *InSensorProfile) Report() string {
	var sb strings.Builder

	sb.WriteString("=== In-Sensor Computing Profile ===\n\n")
	sb.WriteString(fmt.Sprintf("Sensor:     %s (%dx%d)\n", p.SensorType, p.Resolution[0], p.Resolution[1]))
	sb.WriteString(fmt.Sprintf("Mode:       %s\n", p.ProcessingMode))
	sb.WriteString(fmt.Sprintf("Frame Rate: %.1f fps\n\n", p.FrameRate))

	sb.WriteString("--- Power ---\n")
	sb.WriteString(fmt.Sprintf("Sensor:     %.2f mW\n", p.SensorPowerMW))
	sb.WriteString(fmt.Sprintf("Processing: %.2f mW\n", p.ProcessingPowerMW))
	sb.WriteString(fmt.Sprintf("Comm:       %.2f mW\n", p.CommunicationPowerMW))
	sb.WriteString(fmt.Sprintf("Total:      %.2f mW\n\n", p.TotalPowerMW))

	sb.WriteString("--- Latency ---\n")
	sb.WriteString(fmt.Sprintf("Sensor:     %.1f µs\n", p.SensorLatencyUs))
	sb.WriteString(fmt.Sprintf("Processing: %.1f µs\n", p.ProcessingLatencyUs))
	sb.WriteString(fmt.Sprintf("Total:      %.1f µs\n\n", p.TotalLatencyUs))

	sb.WriteString("--- Data Rates ---\n")
	sb.WriteString(fmt.Sprintf("Raw:        %.2f Mbps\n", p.RawDataRateMbps))
	sb.WriteString(fmt.Sprintf("Processed:  %.2f Mbps\n", p.ProcessedDataRateMbps))
	sb.WriteString(fmt.Sprintf("Compression: %.1fx\n", p.CompressionRatio))

	return sb.String()
}
