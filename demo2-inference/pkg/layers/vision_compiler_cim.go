// Package layers provides neuromorphic vision sensor and CIM compiler/mapping optimization
// for compute-in-memory neural network acceleration.
//
// Research context:
// - Event cameras (DVS) produce sparse asynchronous event streams
// - CIM compilers map neural network layers to crossbar arrays
// - Key challenges: tiling, weight replication, mixed precision
//
// Key metrics from literature:
// - Speck chip: <1mW idle power for event-driven processing
// - SpiDR: reconfigurable CIM SNN accelerator
// - Memristor-CMOS hybrid: 75-79% power savings
// - LRMP: layer replication with mixed precision optimization
// - 256x256 tile size with 9-row activation for non-ideality mitigation
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// NEUROMORPHIC VISION SENSOR SIMULATION
// ============================================================================

// EventPolarity represents the polarity of a DVS event
type EventPolarity int

const (
	PolarityOff EventPolarity = -1 // Brightness decrease
	PolarityOn  EventPolarity = 1  // Brightness increase
)

// DVSEvent represents a single event from a Dynamic Vision Sensor
type DVSEvent struct {
	X         int           // X coordinate (pixel)
	Y         int           // Y coordinate (pixel)
	Timestamp float64       // Timestamp in microseconds
	Polarity  EventPolarity // ON (+1) or OFF (-1) event
}

// DVSConfig configures the Dynamic Vision Sensor simulation
type DVSConfig struct {
	Width            int     // Sensor width in pixels
	Height           int     // Sensor height in pixels
	ContrastThreshold float64 // Log intensity change threshold (typically 0.1-0.3)
	RefractoryPeriod float64 // Minimum time between events per pixel (us)
	TemporalNoise    float64 // Noise rate (events/pixel/second)
	DynamicRange     float64 // Dynamic range in dB (typically 120-140 dB)
	Latency          float64 // Pixel latency in microseconds (typically 1-10)
}

// DefaultDVSConfig returns typical DVS sensor parameters
func DefaultDVSConfig() *DVSConfig {
	return &DVSConfig{
		Width:            346,   // DAVIS346 sensor
		Height:           260,
		ContrastThreshold: 0.15, // 15% intensity change
		RefractoryPeriod: 1.0,   // 1 microsecond
		TemporalNoise:    0.1,   // Low noise rate
		DynamicRange:     130,   // 130 dB
		Latency:          3.0,   // 3 microseconds
	}
}

// DVSSensor simulates a Dynamic Vision Sensor
type DVSSensor struct {
	Config       *DVSConfig
	LogIntensity [][]float64 // Log of last intensity per pixel
	LastEventTime [][]float64 // Time of last event per pixel
	EventBuffer  []DVSEvent   // Output event buffer
}

// NewDVSSensor creates a new DVS sensor simulator
func NewDVSSensor(config *DVSConfig) *DVSSensor {
	logInt := make([][]float64, config.Height)
	lastTime := make([][]float64, config.Height)
	for y := 0; y < config.Height; y++ {
		logInt[y] = make([]float64, config.Width)
		lastTime[y] = make([]float64, config.Width)
	}
	return &DVSSensor{
		Config:       config,
		LogIntensity: logInt,
		LastEventTime: lastTime,
		EventBuffer:  make([]DVSEvent, 0, 10000),
	}
}

// ProcessFrame generates events from intensity change between frames
func (d *DVSSensor) ProcessFrame(intensity [][]float64, timestamp float64) []DVSEvent {
	d.EventBuffer = d.EventBuffer[:0]

	for y := 0; y < d.Config.Height; y++ {
		for x := 0; x < d.Config.Width; x++ {
			// Convert to log intensity
			newLogInt := math.Log(intensity[y][x] + 1e-10)
			oldLogInt := d.LogIntensity[y][x]

			// Check refractory period
			if timestamp-d.LastEventTime[y][x] < d.Config.RefractoryPeriod {
				continue
			}

			// Calculate intensity change
			delta := newLogInt - oldLogInt

			// Generate events based on threshold crossings
			for delta > d.Config.ContrastThreshold {
				d.EventBuffer = append(d.EventBuffer, DVSEvent{
					X:         x,
					Y:         y,
					Timestamp: timestamp + d.Config.Latency,
					Polarity:  PolarityOn,
				})
				delta -= d.Config.ContrastThreshold
				d.LogIntensity[y][x] += d.Config.ContrastThreshold
				d.LastEventTime[y][x] = timestamp
			}

			for delta < -d.Config.ContrastThreshold {
				d.EventBuffer = append(d.EventBuffer, DVSEvent{
					X:         x,
					Y:         y,
					Timestamp: timestamp + d.Config.Latency,
					Polarity:  PolarityOff,
				})
				delta += d.Config.ContrastThreshold
				d.LogIntensity[y][x] -= d.Config.ContrastThreshold
				d.LastEventTime[y][x] = timestamp
			}
		}
	}

	// Add temporal noise
	numNoiseEvents := int(d.Config.TemporalNoise * float64(d.Config.Width*d.Config.Height) * 1e-6)
	for i := 0; i < numNoiseEvents; i++ {
		polarity := PolarityOn
		if rand.Float64() < 0.5 {
			polarity = PolarityOff
		}
		d.EventBuffer = append(d.EventBuffer, DVSEvent{
			X:         rand.Intn(d.Config.Width),
			Y:         rand.Intn(d.Config.Height),
			Timestamp: timestamp + rand.Float64()*1000, // Random within 1ms
			Polarity:  polarity,
		})
	}

	// Sort by timestamp
	sort.Slice(d.EventBuffer, func(i, j int) bool {
		return d.EventBuffer[i].Timestamp < d.EventBuffer[j].Timestamp
	})

	return d.EventBuffer
}

// EventStats computes statistics about an event stream
type EventStats struct {
	TotalEvents    int
	OnEvents       int
	OffEvents      int
	EventRate      float64 // Events per second
	SpatialDensity float64 // Active pixels ratio
	TemporalSpan   float64 // Time span in microseconds
}

// ComputeEventStats analyzes an event stream
func ComputeEventStats(events []DVSEvent, width, height int) *EventStats {
	if len(events) == 0 {
		return &EventStats{}
	}

	stats := &EventStats{TotalEvents: len(events)}
	activePixels := make(map[int]bool)

	minTime := events[0].Timestamp
	maxTime := events[0].Timestamp

	for _, e := range events {
		if e.Polarity == PolarityOn {
			stats.OnEvents++
		} else {
			stats.OffEvents++
		}
		activePixels[e.Y*width+e.X] = true
		if e.Timestamp < minTime {
			minTime = e.Timestamp
		}
		if e.Timestamp > maxTime {
			maxTime = e.Timestamp
		}
	}

	stats.TemporalSpan = maxTime - minTime
	stats.SpatialDensity = float64(len(activePixels)) / float64(width*height)
	if stats.TemporalSpan > 0 {
		stats.EventRate = float64(len(events)) / (stats.TemporalSpan * 1e-6)
	}

	return stats
}

// ============================================================================
// EVENT-DRIVEN SNN PROCESSING
// ============================================================================

// SpikingNeuronCIM represents a spiking neuron for CIM-based processing
type SpikingNeuronCIM struct {
	Membrane     float64 // Membrane potential
	Threshold    float64 // Spike threshold
	LeakFactor   float64 // Leak time constant
	ResetPotential float64
	LastSpikeTime float64
	RefractoryTime float64
}

// SpikingLayerCIM implements a spiking neural network layer on CIM
type SpikingLayerCIM struct {
	Neurons      []*SpikingNeuronCIM
	Weights      [][]float64 // Stored in crossbar
	NumInputs    int
	NumOutputs   int
	Crossbar     *CrossbarTileCIM // CIM crossbar for weights
}

// CrossbarTileCIM represents a CIM crossbar tile for SNN processing
type CrossbarTileCIM struct {
	Rows          int
	Cols          int
	Conductances  [][]float64 // Weight conductances
	MaxConductance float64
	MinConductance float64
	RowsPerActivation int // Rows activated at once (for non-ideality mitigation)
	ADCBits       int
	DACBits       int
}

// NewCrossbarTileCIM creates a new CIM crossbar tile
func NewCrossbarTileCIM(rows, cols int) *CrossbarTileCIM {
	cond := make([][]float64, rows)
	for i := range cond {
		cond[i] = make([]float64, cols)
	}
	return &CrossbarTileCIM{
		Rows:              rows,
		Cols:              cols,
		Conductances:      cond,
		MaxConductance:    1.0,
		MinConductance:    0.0,
		RowsPerActivation: 9, // As per literature
		ADCBits:           4,
		DACBits:           1,
	}
}

// ProgramWeights maps weights to crossbar conductances
func (c *CrossbarTileCIM) ProgramWeights(weights [][]float64) error {
	if len(weights) > c.Rows || (len(weights) > 0 && len(weights[0]) > c.Cols) {
		return fmt.Errorf("weights exceed crossbar dimensions")
	}

	// Find weight range for normalization
	minW, maxW := weights[0][0], weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < minW {
				minW = weights[i][j]
			}
			if weights[i][j] > maxW {
				maxW = weights[i][j]
			}
		}
	}

	// Map to conductance range
	wRange := maxW - minW
	if wRange < 1e-10 {
		wRange = 1.0
	}

	for i := range weights {
		for j := range weights[i] {
			normalized := (weights[i][j] - minW) / wRange
			c.Conductances[i][j] = c.MinConductance + normalized*(c.MaxConductance-c.MinConductance)
		}
	}

	return nil
}

// ComputeMVM performs matrix-vector multiplication with row activation batching
func (c *CrossbarTileCIM) ComputeMVM(input []float64) []float64 {
	output := make([]float64, c.Cols)

	// Process in batches of RowsPerActivation
	for startRow := 0; startRow < c.Rows; startRow += c.RowsPerActivation {
		endRow := startRow + c.RowsPerActivation
		if endRow > c.Rows {
			endRow = c.Rows
		}

		// Partial sum for this batch
		for col := 0; col < c.Cols; col++ {
			partialSum := 0.0
			for row := startRow; row < endRow; row++ {
				if row < len(input) {
					// Quantize input (DAC)
					quantizedInput := quantize(input[row], c.DACBits)
					partialSum += quantizedInput * c.Conductances[row][col]
				}
			}
			output[col] += partialSum
		}
	}

	// Quantize output (ADC)
	for i := range output {
		output[i] = quantize(output[i], c.ADCBits)
	}

	return output
}

// quantize applies quantization based on bit precision
func quantize(value float64, bits int) float64 {
	levels := float64(1 << bits)
	return math.Round(value*levels) / levels
}

// ============================================================================
// IN-SENSOR COMPUTING ARCHITECTURE
// ============================================================================

// InSensorConfig configures in-sensor computing
type InSensorConfig struct {
	SensorWidth    int
	SensorHeight   int
	ProcessingUnits int // Number of near-sensor compute units
	LocalMemoryKB  int
	AnalogCompute  bool // Enable analog in-pixel processing
	EventDriven    bool
}

// InSensorProcessor simulates in-sensor computing
type InSensorProcessor struct {
	Config     *InSensorConfig
	PixelPE    [][]float64 // Per-pixel processing element state
	LocalMem   []float64   // Local SRAM buffer
	Crossbar   *CrossbarTileCIM
}

// NewInSensorProcessor creates an in-sensor compute unit
func NewInSensorProcessor(config *InSensorConfig) *InSensorProcessor {
	pixelPE := make([][]float64, config.SensorHeight)
	for y := range pixelPE {
		pixelPE[y] = make([]float64, config.SensorWidth)
	}

	return &InSensorProcessor{
		Config:   config,
		PixelPE:  pixelPE,
		LocalMem: make([]float64, config.LocalMemoryKB*1024/8),
		Crossbar: NewCrossbarTileCIM(64, 64), // Small crossbar for first layer
	}
}

// ProcessEventStream processes DVS events with in-sensor computing
func (p *InSensorProcessor) ProcessEventStream(events []DVSEvent, timeWindow float64) []float64 {
	// Accumulate events into spatial bins
	for y := range p.PixelPE {
		for x := range p.PixelPE[y] {
			p.PixelPE[y][x] = 0
		}
	}

	for _, e := range events {
		if e.X < p.Config.SensorWidth && e.Y < p.Config.SensorHeight {
			p.PixelPE[e.Y][e.X] += float64(e.Polarity)
		}
	}

	// Flatten to input vector
	input := make([]float64, p.Config.SensorWidth*p.Config.SensorHeight)
	idx := 0
	for y := range p.PixelPE {
		for x := range p.PixelPE[y] {
			input[idx] = p.PixelPE[y][x]
			idx++
		}
	}

	// Early feature extraction via crossbar
	if p.Config.AnalogCompute {
		return p.Crossbar.ComputeMVM(input[:p.Crossbar.Rows])
	}

	return input
}

// SpeckChipConfig models the Speck neuromorphic chip parameters
type SpeckChipConfig struct {
	NumCores       int
	NeuronsPerCore int
	SynapsesBits   int
	IdlePowerMW    float64 // <1mW idle
	ActivePowerMW  float64
	EventLatencyUS float64
}

// DefaultSpeckConfig returns Speck chip parameters from literature
func DefaultSpeckConfig() *SpeckChipConfig {
	return &SpeckChipConfig{
		NumCores:       8,
		NeuronsPerCore: 1024,
		SynapsesBits:   8,
		IdlePowerMW:    0.5,
		ActivePowerMW:  10.0,
		EventLatencyUS: 1.0,
	}
}

// ============================================================================
// CIM COMPILER AND MAPPING OPTIMIZATION
// ============================================================================

// LayerSpec describes a neural network layer for mapping
type LayerSpec struct {
	Name        string
	Type        LayerTypeCIM
	InputShape  []int
	OutputShape []int
	WeightShape []int // e.g., [OutFeatures, InFeatures] or [OutC, InC, KH, KW]
	Quantization int   // Bits for weights
	Sparsity    float64 // Weight sparsity ratio
}

// LayerTypeCIM defines layer types for CIM mapping
type LayerTypeCIM int

const (
	LayerLinear LayerTypeCIM = iota
	LayerConv2D
	LayerDepthwiseConv
	LayerAttention
	LayerEmbedding
	LayerBatchNorm
)

// CrossbarSpec describes a physical crossbar array
type CrossbarSpec struct {
	Rows     int
	Cols     int
	Bits     int     // Weight precision
	MaxFreq  float64 // Operating frequency MHz
	EnergyPerMAC float64 // Energy per MAC in pJ
	AreaMM2  float64
}

// TileMapping describes how a weight matrix is mapped to tiles
type TileMapping struct {
	LayerName   string
	TileID      int
	RowStart    int
	RowEnd      int
	ColStart    int
	ColEnd      int
	Replication int // Number of tile replicas for throughput
}

// CIMCompilerConfig configures the CIM compiler
type CIMCompilerConfig struct {
	TileSize        int     // Crossbar tile size (e.g., 256)
	NumTiles        int     // Total available tiles
	RowsPerActivation int   // Rows activated per cycle
	DefaultBits     int     // Default weight precision
	EnableMixedPrec bool    // Enable mixed precision
	EnableReplication bool  // Enable layer replication
	OptimizeFor     string  // "latency", "throughput", "energy"
}

// DefaultCIMCompilerConfig returns typical compiler settings
func DefaultCIMCompilerConfig() *CIMCompilerConfig {
	return &CIMCompilerConfig{
		TileSize:          256,
		NumTiles:          5688, // As per literature
		RowsPerActivation: 9,
		DefaultBits:       4,
		EnableMixedPrec:   true,
		EnableReplication: true,
		OptimizeFor:       "throughput",
	}
}

// CIMCompiler maps neural networks to CIM crossbar arrays
type CIMCompiler struct {
	Config      *CIMCompilerConfig
	Crossbar    *CrossbarSpec
	LayerMappings map[string][]*TileMapping
	TileUsage   int
}

// NewCIMCompiler creates a new CIM compiler
func NewCIMCompiler(config *CIMCompilerConfig) *CIMCompiler {
	return &CIMCompiler{
		Config: config,
		Crossbar: &CrossbarSpec{
			Rows:         config.TileSize,
			Cols:         config.TileSize,
			Bits:         config.DefaultBits,
			MaxFreq:      192, // MHz
			EnergyPerMAC: 0.5, // pJ
			AreaMM2:      0.01,
		},
		LayerMappings: make(map[string][]*TileMapping),
	}
}

// MapLayer maps a single layer to crossbar tiles
func (c *CIMCompiler) MapLayer(layer *LayerSpec) ([]*TileMapping, error) {
	var mappings []*TileMapping

	switch layer.Type {
	case LayerLinear:
		mappings = c.mapLinearLayer(layer)
	case LayerConv2D:
		mappings = c.mapConv2DLayer(layer)
	case LayerAttention:
		mappings = c.mapAttentionLayer(layer)
	default:
		mappings = c.mapLinearLayer(layer) // Default to linear mapping
	}

	c.LayerMappings[layer.Name] = mappings
	return mappings, nil
}

// mapLinearLayer maps a fully connected layer
func (c *CIMCompiler) mapLinearLayer(layer *LayerSpec) []*TileMapping {
	var mappings []*TileMapping

	outFeatures := layer.WeightShape[0]
	inFeatures := layer.WeightShape[1]

	tileID := c.TileUsage

	// Tile across both dimensions
	for row := 0; row < inFeatures; row += c.Config.TileSize {
		for col := 0; col < outFeatures; col += c.Config.TileSize {
			rowEnd := min(row+c.Config.TileSize, inFeatures)
			colEnd := min(col+c.Config.TileSize, outFeatures)

			mappings = append(mappings, &TileMapping{
				LayerName:   layer.Name,
				TileID:      tileID,
				RowStart:    row,
				RowEnd:      rowEnd,
				ColStart:    col,
				ColEnd:      colEnd,
				Replication: 1,
			})
			tileID++
			c.TileUsage++
		}
	}

	return mappings
}

// mapConv2DLayer maps a convolutional layer using im2col
func (c *CIMCompiler) mapConv2DLayer(layer *LayerSpec) []*TileMapping {
	var mappings []*TileMapping

	// [OutC, InC, KH, KW]
	outChannels := layer.WeightShape[0]
	inChannels := layer.WeightShape[1]
	kernelH := layer.WeightShape[2]
	kernelW := layer.WeightShape[3]

	// Im2col transforms conv to matmul: [OutC, InC*KH*KW]
	cols := inChannels * kernelH * kernelW
	rows := outChannels

	tileID := c.TileUsage

	for row := 0; row < rows; row += c.Config.TileSize {
		for col := 0; col < cols; col += c.Config.TileSize {
			rowEnd := min(row+c.Config.TileSize, rows)
			colEnd := min(col+c.Config.TileSize, cols)

			mappings = append(mappings, &TileMapping{
				LayerName:   layer.Name,
				TileID:      tileID,
				RowStart:    row,
				RowEnd:      rowEnd,
				ColStart:    col,
				ColEnd:      colEnd,
				Replication: 1,
			})
			tileID++
			c.TileUsage++
		}
	}

	return mappings
}

// mapAttentionLayer maps attention weights (Q, K, V, O projections)
func (c *CIMCompiler) mapAttentionLayer(layer *LayerSpec) []*TileMapping {
	var mappings []*TileMapping

	// Attention has 4 weight matrices: Q, K, V, O
	dim := layer.WeightShape[0]

	projections := []string{"Q", "K", "V", "O"}
	tileID := c.TileUsage

	for _, proj := range projections {
		for row := 0; row < dim; row += c.Config.TileSize {
			for col := 0; col < dim; col += c.Config.TileSize {
				rowEnd := min(row+c.Config.TileSize, dim)
				colEnd := min(col+c.Config.TileSize, dim)

				mappings = append(mappings, &TileMapping{
					LayerName:   fmt.Sprintf("%s_%s", layer.Name, proj),
					TileID:      tileID,
					RowStart:    row,
					RowEnd:      rowEnd,
					ColStart:    col,
					ColEnd:      colEnd,
					Replication: 1,
				})
				tileID++
				c.TileUsage++
			}
		}
	}

	return mappings
}

// ============================================================================
// LAYER REPLICATION WITH MIXED PRECISION (LRMP)
// ============================================================================

// LRMPConfig configures LRMP optimization
type LRMPConfig struct {
	MaxTiles          int
	MinPrecision      int   // Minimum bits (e.g., 2)
	MaxPrecision      int   // Maximum bits (e.g., 8)
	TargetAccuracy    float64
	OptimizeLatency   bool
	OptimizeThroughput bool
}

// LRMPOptimizer implements Layer Replication with Mixed Precision
type LRMPOptimizer struct {
	Config       *LRMPConfig
	Compiler     *CIMCompiler
	LayerStats   map[string]*LayerPerfStats
}

// LayerPerfStats tracks per-layer performance metrics
type LayerPerfStats struct {
	Name           string
	Latency        float64 // Cycles
	Throughput     float64 // Operations per second
	TilesUsed      int
	Precision      int
	Replication    int
	AccuracyImpact float64 // Estimated accuracy degradation
}

// NewLRMPOptimizer creates an LRMP optimizer
func NewLRMPOptimizer(config *LRMPConfig, compiler *CIMCompiler) *LRMPOptimizer {
	return &LRMPOptimizer{
		Config:     config,
		Compiler:   compiler,
		LayerStats: make(map[string]*LayerPerfStats),
	}
}

// OptimizeNetwork applies LRMP optimization to a network
func (o *LRMPOptimizer) OptimizeNetwork(layers []*LayerSpec) (*OptimizationResult, error) {
	result := &OptimizationResult{
		LayerPrecisions:    make(map[string]int),
		LayerReplications:  make(map[string]int),
		TotalTiles:         0,
	}

	// Phase 1: Compute baseline latencies
	layerLatencies := make(map[string]float64)
	for _, layer := range layers {
		latency := o.estimateLatency(layer, o.Config.MaxPrecision, 1)
		layerLatencies[layer.Name] = latency
	}

	// Find bottleneck (maximum latency layer)
	var maxLatency float64
	var bottleneck string
	for name, lat := range layerLatencies {
		if lat > maxLatency {
			maxLatency = lat
			bottleneck = name
		}
	}

	// Phase 2: Allocate precision and replication
	remainingTiles := o.Config.MaxTiles

	for _, layer := range layers {
		// Start with max precision
		precision := o.Config.MaxPrecision
		replication := 1

		// Calculate tiles needed
		tilesNeeded := o.tilesForLayer(layer, precision)

		// If bottleneck, try replication
		if layer.Name == bottleneck && o.Config.OptimizeThroughput {
			maxRep := remainingTiles / tilesNeeded
			if maxRep > 4 {
				maxRep = 4 // Cap replication
			}
			replication = maxRep
		}

		// If over budget, reduce precision
		for tilesNeeded*replication > remainingTiles && precision > o.Config.MinPrecision {
			precision--
			tilesNeeded = o.tilesForLayer(layer, precision)
		}

		result.LayerPrecisions[layer.Name] = precision
		result.LayerReplications[layer.Name] = replication
		result.TotalTiles += tilesNeeded * replication
		remainingTiles -= tilesNeeded * replication

		// Track stats
		o.LayerStats[layer.Name] = &LayerPerfStats{
			Name:        layer.Name,
			Latency:     o.estimateLatency(layer, precision, replication),
			TilesUsed:   tilesNeeded * replication,
			Precision:   precision,
			Replication: replication,
		}
	}

	// Calculate final metrics
	result.EstimatedLatency = o.calculateTotalLatency(layers, result)
	result.EstimatedThroughput = 1.0 / result.EstimatedLatency

	return result, nil
}

// estimateLatency estimates layer latency based on precision and replication
func (o *LRMPOptimizer) estimateLatency(layer *LayerSpec, precision, replication int) float64 {
	// Base latency from weight size
	var ops int
	switch layer.Type {
	case LayerLinear:
		ops = layer.WeightShape[0] * layer.WeightShape[1]
	case LayerConv2D:
		ops = layer.WeightShape[0] * layer.WeightShape[1] * layer.WeightShape[2] * layer.WeightShape[3]
	default:
		ops = 1
		for _, d := range layer.WeightShape {
			ops *= d
		}
	}

	// Cycles based on tile size and row activation
	tileSize := o.Compiler.Config.TileSize
	rowsPerActivation := o.Compiler.Config.RowsPerActivation

	cyclesPerTile := float64(tileSize) / float64(rowsPerActivation)
	numTiles := float64(ops) / float64(tileSize*tileSize)

	// Precision affects cycles (more bits = more cycles)
	precisionFactor := float64(precision) / 4.0

	return (cyclesPerTile * numTiles * precisionFactor) / float64(replication)
}

// tilesForLayer calculates tiles needed for a layer at given precision
func (o *LRMPOptimizer) tilesForLayer(layer *LayerSpec, precision int) int {
	var elements int
	for _, d := range layer.WeightShape {
		if elements == 0 {
			elements = d
		} else {
			elements *= d
		}
	}

	tileCapacity := o.Compiler.Config.TileSize * o.Compiler.Config.TileSize
	tiles := (elements + tileCapacity - 1) / tileCapacity

	// Higher precision may require more tiles
	precisionTiles := (precision + 3) / 4 // Assume 4 bits per cell

	return tiles * precisionTiles
}

// calculateTotalLatency computes network latency with pipeline consideration
func (o *LRMPOptimizer) calculateTotalLatency(layers []*LayerSpec, result *OptimizationResult) float64 {
	var maxLatency float64
	for _, layer := range layers {
		lat := o.estimateLatency(layer, result.LayerPrecisions[layer.Name], result.LayerReplications[layer.Name])
		if lat > maxLatency {
			maxLatency = lat
		}
	}
	return maxLatency
}

// OptimizationResult holds LRMP optimization results
type OptimizationResult struct {
	LayerPrecisions    map[string]int
	LayerReplications  map[string]int
	TotalTiles         int
	EstimatedLatency   float64
	EstimatedThroughput float64
}

// ============================================================================
// DATAFLOW OPTIMIZATION
// ============================================================================

// DataflowType defines CIM dataflow patterns
type DataflowType int

const (
	DataflowWeightStationary DataflowType = iota
	DataflowOutputStationary
	DataflowInputStationary
	DataflowRowStationary
)

// DataflowAnalyzer analyzes optimal dataflow for layers
type DataflowAnalyzer struct {
	TileSize   int
	BandwidthGBps float64
	OnChipMemKB   int
}

// DataflowChoice represents a dataflow decision
type DataflowChoice struct {
	Type           DataflowType
	ReuseFactor    float64
	MemoryTraffic  float64 // GB
	ComputeUtil    float64 // Utilization percentage
}

// AnalyzeLayer determines optimal dataflow for a layer
func (d *DataflowAnalyzer) AnalyzeLayer(layer *LayerSpec) *DataflowChoice {
	// Compute data sizes
	var weightSize, inputSize, outputSize int

	switch layer.Type {
	case LayerLinear:
		weightSize = layer.WeightShape[0] * layer.WeightShape[1]
		inputSize = layer.InputShape[0] * layer.InputShape[1]
		outputSize = layer.OutputShape[0] * layer.OutputShape[1]

	case LayerConv2D:
		weightSize = layer.WeightShape[0] * layer.WeightShape[1] * layer.WeightShape[2] * layer.WeightShape[3]
		inputSize = 1
		for _, dim := range layer.InputShape {
			inputSize *= dim
		}
		outputSize = 1
		for _, dim := range layer.OutputShape {
			outputSize *= dim
		}
	}

	// Choose dataflow based on data ratios
	choice := &DataflowChoice{}

	// Weight stationary if weights fit in tiles
	tileCapacity := d.TileSize * d.TileSize
	if weightSize <= tileCapacity*100 { // Fits in reasonable number of tiles
		choice.Type = DataflowWeightStationary
		choice.ReuseFactor = float64(inputSize) // Each weight reused for all inputs
		choice.MemoryTraffic = float64(inputSize+outputSize) * 4 / 1e9 // Assume 4 bytes
	} else if outputSize < inputSize {
		choice.Type = DataflowOutputStationary
		choice.ReuseFactor = float64(weightSize)
		choice.MemoryTraffic = float64(weightSize+inputSize) * 4 / 1e9
	} else {
		choice.Type = DataflowInputStationary
		choice.ReuseFactor = float64(weightSize)
		choice.MemoryTraffic = float64(weightSize+outputSize) * 4 / 1e9
	}

	// Estimate compute utilization
	ops := float64(weightSize * inputSize / max(layer.InputShape[len(layer.InputShape)-1], 1))
	peakOps := float64(d.TileSize * d.TileSize * 192e6) // At 192 MHz
	choice.ComputeUtil = min(ops/peakOps, 1.0)

	return choice
}

// ============================================================================
// WEIGHT TILING AND PARTITIONING
// ============================================================================

// TilingStrategy defines weight matrix tiling approaches
type TilingStrategy int

const (
	TilingRowMajor TilingStrategy = iota
	TilingColMajor
	TilingBlocked
	TilingZOrder // Space-filling curve
)

// WeightTiler handles weight matrix partitioning
type WeightTiler struct {
	TileRows    int
	TileCols    int
	Strategy    TilingStrategy
	MaxParallelTiles int
}

// TilePartition represents a partitioned weight region
type TilePartition struct {
	TileIndex   int
	RowRange    [2]int // [start, end)
	ColRange    [2]int
	DataOffset  int    // Offset in flattened weight array
}

// PartitionWeights divides a weight matrix into tiles
func (t *WeightTiler) PartitionWeights(rows, cols int) []*TilePartition {
	var partitions []*TilePartition
	idx := 0

	switch t.Strategy {
	case TilingRowMajor:
		for r := 0; r < rows; r += t.TileRows {
			for c := 0; c < cols; c += t.TileCols {
				partitions = append(partitions, &TilePartition{
					TileIndex: idx,
					RowRange:  [2]int{r, min(r+t.TileRows, rows)},
					ColRange:  [2]int{c, min(c+t.TileCols, cols)},
				})
				idx++
			}
		}

	case TilingColMajor:
		for c := 0; c < cols; c += t.TileCols {
			for r := 0; r < rows; r += t.TileRows {
				partitions = append(partitions, &TilePartition{
					TileIndex: idx,
					RowRange:  [2]int{r, min(r+t.TileRows, rows)},
					ColRange:  [2]int{c, min(c+t.TileCols, cols)},
				})
				idx++
			}
		}

	case TilingBlocked:
		// 2D blocked for better locality
		blockRows := t.TileRows * 2
		blockCols := t.TileCols * 2
		for br := 0; br < rows; br += blockRows {
			for bc := 0; bc < cols; bc += blockCols {
				// Tiles within block
				for r := br; r < min(br+blockRows, rows); r += t.TileRows {
					for c := bc; c < min(bc+blockCols, cols); c += t.TileCols {
						partitions = append(partitions, &TilePartition{
							TileIndex: idx,
							RowRange:  [2]int{r, min(r+t.TileRows, rows)},
							ColRange:  [2]int{c, min(c+t.TileCols, cols)},
						})
						idx++
					}
				}
			}
		}
	}

	return partitions
}

// ComputePartialSums aggregates partial sums from tiled computation
func (t *WeightTiler) ComputePartialSums(partitions []*TilePartition, results map[int][]float64, outputSize int) []float64 {
	output := make([]float64, outputSize)

	for _, part := range partitions {
		if result, ok := results[part.TileIndex]; ok {
			for i, v := range result {
				colIdx := part.ColRange[0] + i
				if colIdx < outputSize {
					output[colIdx] += v
				}
			}
		}
	}

	return output
}

// ============================================================================
// CIM-MLC MULTI-LEVEL COMPILATION
// ============================================================================

// CIMMlcCompiler implements multi-level CIM compilation
type CIMMlcCompiler struct {
	// Level 1: Network level
	NetworkGraph *ComputeGraph

	// Level 2: Core level
	CoreMapping  map[string]int // Layer -> Core ID

	// Level 3: Crossbar level
	CrossbarMaps map[int][]*TileMapping // Core -> Tiles

	// Level 4: Wordline level
	WordlineSchedule map[int][]int // Tile -> Wordline activation order

	Config       *CIMMlcConfig
}

// CIMMlcConfig configures multi-level compilation
type CIMMlcConfig struct {
	NumCores        int
	TilesPerCore    int
	WordlinesPerTile int
	PipelineDepth   int
	EnableFusion    bool
}

// ComputeGraph represents the network as a DAG
type ComputeGraph struct {
	Nodes []*ComputeNode
	Edges map[int][]int // Node ID -> successor IDs
}

// ComputeNode represents a computation in the graph
type ComputeNode struct {
	ID          int
	LayerSpec   *LayerSpec
	Dependencies []int
	CoreID      int
	Scheduled   bool
}

// NewCIMMlcCompiler creates a multi-level compiler
func NewCIMMlcCompiler(config *CIMMlcConfig) *CIMMlcCompiler {
	return &CIMMlcCompiler{
		NetworkGraph:     &ComputeGraph{Edges: make(map[int][]int)},
		CoreMapping:      make(map[string]int),
		CrossbarMaps:     make(map[int][]*TileMapping),
		WordlineSchedule: make(map[int][]int),
		Config:           config,
	}
}

// CompileNetwork performs multi-level compilation
func (c *CIMMlcCompiler) CompileNetwork(layers []*LayerSpec) error {
	// Level 1: Build compute graph
	for i, layer := range layers {
		node := &ComputeNode{
			ID:        i,
			LayerSpec: layer,
		}
		if i > 0 {
			node.Dependencies = []int{i - 1}
		}
		c.NetworkGraph.Nodes = append(c.NetworkGraph.Nodes, node)
		if i < len(layers)-1 {
			c.NetworkGraph.Edges[i] = []int{i + 1}
		}
	}

	// Level 2: Map to cores (load balancing)
	c.mapToCores()

	// Level 3: Map to crossbars
	c.mapToCrossbars()

	// Level 4: Schedule wordlines
	c.scheduleWordlines()

	return nil
}

// mapToCores assigns layers to cores for load balancing
func (c *CIMMlcCompiler) mapToCores() {
	coreLoads := make([]int, c.Config.NumCores)

	for _, node := range c.NetworkGraph.Nodes {
		// Estimate layer compute cost
		cost := 1
		for _, d := range node.LayerSpec.WeightShape {
			cost *= d
		}

		// Assign to least loaded core
		minCore := 0
		minLoad := coreLoads[0]
		for i := 1; i < c.Config.NumCores; i++ {
			if coreLoads[i] < minLoad {
				minCore = i
				minLoad = coreLoads[i]
			}
		}

		c.CoreMapping[node.LayerSpec.Name] = minCore
		node.CoreID = minCore
		coreLoads[minCore] += cost
	}
}

// mapToCrossbars maps layers within each core to crossbar tiles
func (c *CIMMlcCompiler) mapToCrossbars() {
	compiler := NewCIMCompiler(DefaultCIMCompilerConfig())

	for coreID := 0; coreID < c.Config.NumCores; coreID++ {
		var coreMappings []*TileMapping

		for _, node := range c.NetworkGraph.Nodes {
			if node.CoreID == coreID {
				mappings, _ := compiler.MapLayer(node.LayerSpec)
				coreMappings = append(coreMappings, mappings...)
			}
		}

		c.CrossbarMaps[coreID] = coreMappings
	}
}

// scheduleWordlines creates wordline activation schedules
func (c *CIMMlcCompiler) scheduleWordlines() {
	for coreID, mappings := range c.CrossbarMaps {
		for _, mapping := range mappings {
			// Create activation schedule for this tile
			numRows := mapping.RowEnd - mapping.RowStart
			rowsPerActivation := 9 // Standard

			var schedule []int
			for r := 0; r < numRows; r += rowsPerActivation {
				schedule = append(schedule, r)
			}

			c.WordlineSchedule[mapping.TileID] = schedule
		}
		_ = coreID // Used in loop
	}
}

// ============================================================================
// PERFORMANCE ESTIMATION
// ============================================================================

// PerfEstimator estimates CIM system performance
type PerfEstimator struct {
	ClockFreqMHz  float64
	TileLatency   int // Cycles per tile MVM
	ADCLatency    int
	DACLatency    int
	InterconnectLat int
}

// EstimatePerformance computes expected performance metrics
func (p *PerfEstimator) EstimatePerformance(compiler *CIMMlcCompiler) *PerfMetrics {
	metrics := &PerfMetrics{}

	// Count total tiles
	for _, mappings := range compiler.CrossbarMaps {
		metrics.TotalTiles += len(mappings)
	}

	// Estimate cycles
	var maxCycles int
	for _, mappings := range compiler.CrossbarMaps {
		coreCycles := 0
		for _, mapping := range mappings {
			numActivations := (mapping.RowEnd - mapping.RowStart + 8) / 9
			tileCycles := numActivations * (p.TileLatency + p.ADCLatency)
			coreCycles += tileCycles
		}
		if coreCycles > maxCycles {
			maxCycles = coreCycles
		}
	}

	metrics.TotalCycles = maxCycles
	metrics.LatencyUS = float64(maxCycles) / p.ClockFreqMHz
	metrics.ThroughputTOPS = float64(metrics.TotalTiles*256*256) / (metrics.LatencyUS * 1e6)

	// Energy estimation (pJ/MAC typical values)
	pjPerMAC := 0.5
	totalMACs := float64(metrics.TotalTiles * 256 * 256)
	metrics.EnergyUJ = totalMACs * pjPerMAC / 1e6
	metrics.EfficiencyTOPSW = metrics.ThroughputTOPS / (metrics.EnergyUJ * 1e6 / metrics.LatencyUS)

	return metrics
}

// PerfMetrics holds performance estimation results
type PerfMetrics struct {
	TotalTiles      int
	TotalCycles     int
	LatencyUS       float64
	ThroughputTOPS  float64
	EnergyUJ        float64
	EfficiencyTOPSW float64
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ============================================================================
// SIMULATION RUNNER
// ============================================================================

// VisionCompilerDemo demonstrates neuromorphic vision and CIM compilation
func VisionCompilerDemo() {
	fmt.Println("=== Neuromorphic Vision + CIM Compiler Demo ===")

	// 1. DVS Sensor simulation
	fmt.Println("\n1. DVS Sensor Simulation:")
	dvsConfig := DefaultDVSConfig()
	dvs := NewDVSSensor(dvsConfig)

	// Simulate moving edge
	frame := make([][]float64, dvsConfig.Height)
	for y := range frame {
		frame[y] = make([]float64, dvsConfig.Width)
		for x := range frame[y] {
			if x > 100 && x < 200 {
				frame[y][x] = 1.0
			} else {
				frame[y][x] = 0.1
			}
		}
	}

	events := dvs.ProcessFrame(frame, 0)
	stats := ComputeEventStats(events, dvsConfig.Width, dvsConfig.Height)
	fmt.Printf("   Events generated: %d (ON: %d, OFF: %d)\n", stats.TotalEvents, stats.OnEvents, stats.OffEvents)
	fmt.Printf("   Spatial density: %.3f\n", stats.SpatialDensity)

	// 2. CIM Compiler
	fmt.Println("\n2. CIM Compiler Mapping:")
	compiler := NewCIMCompiler(DefaultCIMCompilerConfig())

	layers := []*LayerSpec{
		{Name: "conv1", Type: LayerConv2D, WeightShape: []int{32, 1, 3, 3}},
		{Name: "conv2", Type: LayerConv2D, WeightShape: []int{64, 32, 3, 3}},
		{Name: "fc1", Type: LayerLinear, WeightShape: []int{256, 1024}},
		{Name: "fc2", Type: LayerLinear, WeightShape: []int{10, 256}},
	}

	for _, layer := range layers {
		mappings, _ := compiler.MapLayer(layer)
		fmt.Printf("   %s: %d tiles\n", layer.Name, len(mappings))
	}

	// 3. LRMP Optimization
	fmt.Println("\n3. LRMP Optimization:")
	lrmpConfig := &LRMPConfig{
		MaxTiles:       1000,
		MinPrecision:   2,
		MaxPrecision:   8,
		TargetAccuracy: 0.95,
		OptimizeThroughput: true,
	}

	optimizer := NewLRMPOptimizer(lrmpConfig, compiler)
	result, _ := optimizer.OptimizeNetwork(layers)

	fmt.Printf("   Total tiles used: %d\n", result.TotalTiles)
	for name, prec := range result.LayerPrecisions {
		fmt.Printf("   %s: %d-bit, %dx replication\n", name, prec, result.LayerReplications[name])
	}

	// 4. Multi-level compilation
	fmt.Println("\n4. CIM-MLC Multi-level Compilation:")
	mlcConfig := &CIMMlcConfig{
		NumCores:        8,
		TilesPerCore:    128,
		WordlinesPerTile: 256,
		PipelineDepth:   4,
		EnableFusion:    true,
	}

	mlcCompiler := NewCIMMlcCompiler(mlcConfig)
	mlcCompiler.CompileNetwork(layers)

	for coreID, mappings := range mlcCompiler.CrossbarMaps {
		if len(mappings) > 0 {
			fmt.Printf("   Core %d: %d tiles mapped\n", coreID, len(mappings))
		}
	}

	// 5. Performance estimation
	fmt.Println("\n5. Performance Estimation:")
	estimator := &PerfEstimator{
		ClockFreqMHz:  192,
		TileLatency:   10,
		ADCLatency:    2,
		DACLatency:    1,
		InterconnectLat: 5,
	}

	perf := estimator.EstimatePerformance(mlcCompiler)
	fmt.Printf("   Latency: %.2f us\n", perf.LatencyUS)
	fmt.Printf("   Throughput: %.2f TOPS\n", perf.ThroughputTOPS)
	fmt.Printf("   Efficiency: %.2f TOPS/W\n", perf.EfficiencyTOPSW)

	fmt.Println("\n=== Demo Complete ===")
}
