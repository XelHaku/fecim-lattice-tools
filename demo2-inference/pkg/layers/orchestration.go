// Package layers provides neural network layer implementations for crossbar-based CIM.
// orchestration.go implements multi-array coordination, pipelining, and benchmarking.
//
// Multi-Array Features:
// - Layer-to-array mapping
// - Pipelined execution
// - Weight replication for load balancing
// - Partial sum accumulation
//
// Benchmarking:
// - Throughput measurement (TOPS, TOPS/W)
// - Latency profiling
// - Energy estimation
// - Accuracy tracking
//
// References:
// - arXiv 2501.06780: COMPASS compiler framework
// - Nature 2022: RRAM compute-in-memory chip
// - Science Advances 2024: In-situ training

package layers

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// ============================================================================
// Multi-Array Configuration
// ============================================================================

// OrchestrationConfig configures multi-array system
type OrchestrationConfig struct {
	NumArrays        int
	ArrayRows        int
	ArrayCols        int
	PipelineDepth    int
	BufferSizeKB     int
	ClockFreqMHz     float64
	EnergyPerMAC     float64 // fJ
	EnergyPerAccess  float64 // fJ per byte
	WeightReplication bool
}

// DefaultOrchestrationConfig returns default orchestration settings
func DefaultOrchestrationConfig() *OrchestrationConfig {
	return &OrchestrationConfig{
		NumArrays:        16,
		ArrayRows:        64,
		ArrayCols:        64,
		PipelineDepth:    4,
		BufferSizeKB:     64,
		ClockFreqMHz:     1000.0,
		EnergyPerMAC:     100.0,
		EnergyPerAccess:  10.0,
		WeightReplication: true,
	}
}

// ============================================================================
// Array Manager
// ============================================================================

// ArrayManager coordinates multiple crossbar arrays
type ArrayManager struct {
	Config       *OrchestrationConfig
	Arrays       []*CrossbarArray
	Assignments  map[int][]*ArrayAssignment // layer -> array assignments
	Pipeline     *ExecutionPipeline
	Stats        *OrchestrationStats
	mu           sync.Mutex
}

// CrossbarArray represents a single crossbar array
type CrossbarArray struct {
	ID           int
	Rows         int
	Cols         int
	Weights      [][]float64
	Conductances [][]float64
	Occupied     bool
	LayerID      int
	TileRow      int
	TileCol      int
}

// ArrayAssignment maps layer tiles to arrays
type ArrayAssignment struct {
	LayerID   int
	ArrayID   int
	TileRow   int
	TileCol   int
	RowStart  int
	RowEnd    int
	ColStart  int
	ColEnd    int
}

// OrchestrationStats tracks orchestration statistics
type OrchestrationStats struct {
	TotalInferences    int64
	TotalMACs          int64
	TotalCycles        int64
	TotalEnergyFJ      float64
	PipelineStalls     int64
	WeightReloads      int64
	BufferOverflows    int64
	Throughput         float64 // TOPS
	EnergyEfficiency   float64 // TOPS/W
	AverageLatencyUs   float64
}

// NewArrayManager creates a new array manager
func NewArrayManager(config *OrchestrationConfig) *ArrayManager {
	if config == nil {
		config = DefaultOrchestrationConfig()
	}

	manager := &ArrayManager{
		Config:      config,
		Arrays:      make([]*CrossbarArray, config.NumArrays),
		Assignments: make(map[int][]*ArrayAssignment),
		Stats:       &OrchestrationStats{},
	}

	// Initialize arrays
	for i := 0; i < config.NumArrays; i++ {
		manager.Arrays[i] = &CrossbarArray{
			ID:       i,
			Rows:     config.ArrayRows,
			Cols:     config.ArrayCols,
			Occupied: false,
		}
	}

	return manager
}

// ============================================================================
// Layer Mapping
// ============================================================================

// MapLayer maps a layer to available arrays
func (m *ArrayManager) MapLayer(layerID int, weights [][]float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rows := len(weights)
	if rows == 0 {
		return fmt.Errorf("empty weights")
	}
	cols := len(weights[0])

	// Calculate tiles needed
	tilesRow := (rows + m.Config.ArrayRows - 1) / m.Config.ArrayRows
	tilesCol := (cols + m.Config.ArrayCols - 1) / m.Config.ArrayCols
	totalTiles := tilesRow * tilesCol

	// Find available arrays
	available := m.getAvailableArrays()
	if len(available) < totalTiles {
		// Need weight reloading
		m.Stats.WeightReloads++
	}

	assignments := make([]*ArrayAssignment, 0, totalTiles)

	tileIdx := 0
	for ti := 0; ti < tilesRow; ti++ {
		for tj := 0; tj < tilesCol; tj++ {
			arrayID := tileIdx % m.Config.NumArrays

			rowStart := ti * m.Config.ArrayRows
			rowEnd := min((ti+1)*m.Config.ArrayRows, rows)
			colStart := tj * m.Config.ArrayCols
			colEnd := min((tj+1)*m.Config.ArrayCols, cols)

			// Extract tile weights
			tileWeights := make([][]float64, rowEnd-rowStart)
			for i := rowStart; i < rowEnd; i++ {
				tileWeights[i-rowStart] = make([]float64, colEnd-colStart)
				for j := colStart; j < colEnd; j++ {
					tileWeights[i-rowStart][j-colStart] = weights[i][j]
				}
			}

			// Assign to array
			m.Arrays[arrayID].Weights = tileWeights
			m.Arrays[arrayID].Occupied = true
			m.Arrays[arrayID].LayerID = layerID
			m.Arrays[arrayID].TileRow = ti
			m.Arrays[arrayID].TileCol = tj

			assignments = append(assignments, &ArrayAssignment{
				LayerID:  layerID,
				ArrayID:  arrayID,
				TileRow:  ti,
				TileCol:  tj,
				RowStart: rowStart,
				RowEnd:   rowEnd,
				ColStart: colStart,
				ColEnd:   colEnd,
			})

			tileIdx++
		}
	}

	m.Assignments[layerID] = assignments
	return nil
}

// getAvailableArrays returns list of unoccupied arrays
func (m *ArrayManager) getAvailableArrays() []int {
	available := make([]int, 0)
	for i, arr := range m.Arrays {
		if !arr.Occupied {
			available = append(available, i)
		}
	}
	return available
}

// ReleaseLayer releases arrays used by a layer
func (m *ArrayManager) ReleaseLayer(layerID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	assignments, exists := m.Assignments[layerID]
	if !exists {
		return
	}

	for _, assign := range assignments {
		m.Arrays[assign.ArrayID].Occupied = false
		m.Arrays[assign.ArrayID].Weights = nil
	}

	delete(m.Assignments, layerID)
}

// ============================================================================
// Pipelined Execution
// ============================================================================

// ExecutionPipeline manages pipelined layer execution
type ExecutionPipeline struct {
	Config     *OrchestrationConfig
	Stages     []*PipelineStage
	Buffers    []*InterstageBuffer
	CurrentCycle int64
}

// PipelineStage represents one pipeline stage
type PipelineStage struct {
	StageID    int
	LayerID    int
	Arrays     []int // Array IDs assigned to this stage
	InputReady bool
	OutputReady bool
	Busy       bool
}

// InterstageBuffer holds data between pipeline stages
type InterstageBuffer struct {
	Data     [][]float64
	Capacity int
	Size     int
}

// NewExecutionPipeline creates a new execution pipeline
func NewExecutionPipeline(config *OrchestrationConfig, numLayers int) *ExecutionPipeline {
	depth := min(config.PipelineDepth, numLayers)

	pipeline := &ExecutionPipeline{
		Config:  config,
		Stages:  make([]*PipelineStage, depth),
		Buffers: make([]*InterstageBuffer, depth-1),
	}

	for i := 0; i < depth; i++ {
		pipeline.Stages[i] = &PipelineStage{
			StageID: i,
			LayerID: -1,
		}
	}

	bufferSize := config.BufferSizeKB * 1024 / 4 // float32
	for i := 0; i < depth-1; i++ {
		pipeline.Buffers[i] = &InterstageBuffer{
			Capacity: bufferSize,
		}
	}

	return pipeline
}

// Execute runs one pipeline cycle
func (p *ExecutionPipeline) Execute(inputs [][]float64, manager *ArrayManager) [][]float64 {
	p.CurrentCycle++

	// Process stages in reverse order (to allow forwarding)
	var outputs [][]float64

	for i := len(p.Stages) - 1; i >= 0; i-- {
		stage := p.Stages[i]

		if !stage.Busy || stage.LayerID < 0 {
			continue
		}

		// Get input for this stage
		var stageInput [][]float64
		if i == 0 {
			stageInput = inputs
		} else if p.Buffers[i-1].Size > 0 {
			stageInput = p.Buffers[i-1].Data
		}

		if stageInput == nil {
			continue
		}

		// Execute layer on assigned arrays
		stageOutput := p.executeStage(stage, stageInput, manager)

		// Store output
		if i == len(p.Stages)-1 {
			outputs = stageOutput
		} else if i < len(p.Buffers) {
			p.Buffers[i].Data = stageOutput
			p.Buffers[i].Size = len(stageOutput)
		}
	}

	return outputs
}

// executeStage executes one pipeline stage
func (p *ExecutionPipeline) executeStage(stage *PipelineStage, inputs [][]float64, manager *ArrayManager) [][]float64 {
	assignments, exists := manager.Assignments[stage.LayerID]
	if !exists || len(inputs) == 0 {
		return nil
	}

	// Determine output size
	maxRow := 0
	for _, assign := range assignments {
		if assign.RowEnd > maxRow {
			maxRow = assign.RowEnd
		}
	}

	outputs := make([][]float64, len(inputs))
	for b := range outputs {
		outputs[b] = make([]float64, maxRow)
	}

	// Execute on each assigned array and accumulate
	for _, assign := range assignments {
		array := manager.Arrays[assign.ArrayID]
		if array.Weights == nil {
			continue
		}

		// MVM for each input in batch
		for b, input := range inputs {
			for i := range array.Weights {
				sum := 0.0
				for j := range array.Weights[i] {
					colIdx := assign.ColStart + j
					if colIdx < len(input) {
						sum += array.Weights[i][j] * input[colIdx]
					}
				}
				outputs[b][assign.RowStart+i] += sum
			}
		}

		// Update stats
		macs := len(array.Weights) * len(array.Weights[0]) * len(inputs)
		manager.Stats.TotalMACs += int64(macs)
	}

	return outputs
}

// ============================================================================
// Benchmarking
// ============================================================================

// Benchmark represents a benchmark result
type Benchmark struct {
	Name            string
	Throughput      float64 // TOPS
	EnergyEff       float64 // TOPS/W
	Latency         float64 // microseconds
	Accuracy        float64
	MACs            int64
	Energy          float64 // picojoules
	MemoryUsage     int64   // bytes
}

// Benchmarker runs performance benchmarks
type Benchmarker struct {
	Config  *OrchestrationConfig
	Results []*Benchmark
}

// NewBenchmarker creates a new benchmarker
func NewBenchmarker(config *OrchestrationConfig) *Benchmarker {
	if config == nil {
		config = DefaultOrchestrationConfig()
	}
	return &Benchmarker{
		Config:  config,
		Results: make([]*Benchmark, 0),
	}
}

// BenchmarkInference runs inference benchmark
func (b *Benchmarker) BenchmarkInference(name string, model *ModelCheckpoint, inputs [][]float64, labels []int) *Benchmark {
	startTime := time.Now()

	// Create pipeline
	pipe := NewInferencePipeline(model, &PipelineConfig{
		UseCIMSimulation: true,
		WeightBits:       6,
		NoiseLevel:       0.02,
		CrossbarSize:     b.Config.ArrayRows,
	})

	// Run inference
	predictions := pipe.PredictBatch(inputs)

	elapsed := time.Since(startTime)

	// Calculate accuracy
	correct := 0
	for i, pred := range predictions {
		if pred == labels[i] {
			correct++
		}
	}
	accuracy := float64(correct) / float64(len(labels))

	// Calculate metrics
	stats := pipe.GetStats()
	macs := stats.MACOperations

	// Throughput: TOPS = MACs / time / 1e12
	throughput := float64(macs) / elapsed.Seconds() / 1e12

	// Energy (fJ)
	energy := float64(macs) * b.Config.EnergyPerMAC

	// Energy efficiency: TOPS/W = TOPS / (power in W)
	// Power = Energy / Time
	powerW := energy / 1e15 / elapsed.Seconds() // Convert fJ to J
	energyEff := 0.0
	if powerW > 0 {
		energyEff = throughput / powerW
	}

	benchmark := &Benchmark{
		Name:        name,
		Throughput:  throughput,
		EnergyEff:   energyEff,
		Latency:     elapsed.Seconds() * 1e6 / float64(len(inputs)), // per sample
		Accuracy:    accuracy,
		MACs:        macs,
		Energy:      energy / 1e3, // picojoules
		MemoryUsage: int64(len(model.Layers)) * int64(b.Config.ArrayRows*b.Config.ArrayCols*4),
	}

	b.Results = append(b.Results, benchmark)
	return benchmark
}

// BenchmarkThroughput measures peak throughput
func (b *Benchmarker) BenchmarkThroughput(name string, rows, cols, batchSize, iterations int) *Benchmark {
	// Create random weights
	weights := make([][]float64, rows)
	for i := range weights {
		weights[i] = make([]float64, cols)
		for j := range weights[i] {
			weights[i][j] = 0.1
		}
	}

	// Create random inputs
	inputs := make([][]float64, batchSize)
	for i := range inputs {
		inputs[i] = make([]float64, cols)
		for j := range inputs[i] {
			inputs[i][j] = 0.5
		}
	}

	// Time MVM operations
	startTime := time.Now()

	for iter := 0; iter < iterations; iter++ {
		for _, input := range inputs {
			output := make([]float64, rows)
			for i := 0; i < rows; i++ {
				sum := 0.0
				for j := 0; j < cols; j++ {
					sum += weights[i][j] * input[j]
				}
				output[i] = sum
			}
		}
	}

	elapsed := time.Since(startTime)

	// Calculate metrics
	totalMACs := int64(rows) * int64(cols) * int64(batchSize) * int64(iterations)
	throughput := float64(totalMACs) / elapsed.Seconds() / 1e12

	// Theoretical peak
	peakMACs := float64(b.Config.NumArrays) * float64(b.Config.ArrayRows) * float64(b.Config.ArrayCols)
	peakThroughput := peakMACs * b.Config.ClockFreqMHz * 1e6 / 1e12

	benchmark := &Benchmark{
		Name:       name,
		Throughput: throughput,
		MACs:       totalMACs,
		Latency:    elapsed.Seconds() * 1e6 / float64(iterations),
	}

	b.Results = append(b.Results, benchmark)

	// Print comparison
	fmt.Printf("Peak theoretical: %.2f TOPS\n", peakThroughput)
	fmt.Printf("Achieved: %.4f TOPS (%.2f%% of peak)\n", throughput, throughput/peakThroughput*100)

	return benchmark
}

// BenchmarkEnergy estimates energy consumption
func (b *Benchmarker) BenchmarkEnergy(name string, macs int64) *Benchmark {
	// MAC energy
	macEnergy := float64(macs) * b.Config.EnergyPerMAC

	// Memory access energy (simplified)
	// Assume 1 byte per MAC for weights + activations
	accessEnergy := float64(macs) * 2 * b.Config.EnergyPerAccess

	totalEnergy := macEnergy + accessEnergy

	// Calculate TOPS/W
	// Assuming 1 inference takes N cycles at ClockFreqMHz
	cyclesPerInference := float64(macs) / (float64(b.Config.NumArrays) * float64(b.Config.ArrayRows))
	timePerInference := cyclesPerInference / (b.Config.ClockFreqMHz * 1e6)
	powerW := totalEnergy / 1e15 / timePerInference

	throughput := float64(macs) / timePerInference / 1e12
	energyEff := 0.0
	if powerW > 0 {
		energyEff = throughput / powerW
	}

	benchmark := &Benchmark{
		Name:       name,
		EnergyEff:  energyEff,
		Throughput: throughput,
		MACs:       macs,
		Energy:     totalEnergy / 1e3, // picojoules
	}

	b.Results = append(b.Results, benchmark)
	return benchmark
}

// CompareBenchmarks compares multiple benchmark results
func (b *Benchmarker) CompareBenchmarks() {
	fmt.Println("\n=== Benchmark Comparison ===")
	fmt.Printf("%-20s %12s %12s %12s %12s\n", "Name", "TOPS", "TOPS/W", "Latency(us)", "Accuracy")
	fmt.Println(string(make([]byte, 72)))

	for _, bm := range b.Results {
		fmt.Printf("%-20s %12.4f %12.2f %12.2f %12.2f%%\n",
			bm.Name, bm.Throughput, bm.EnergyEff, bm.Latency, bm.Accuracy*100)
	}
}

// ============================================================================
// Weight Replication
// ============================================================================

// WeightReplicator handles weight replication for load balancing
type WeightReplicator struct {
	Config          *OrchestrationConfig
	ReplicationFactor map[int]int // layer -> replication factor
}

// NewWeightReplicator creates a weight replicator
func NewWeightReplicator(config *OrchestrationConfig) *WeightReplicator {
	return &WeightReplicator{
		Config:            config,
		ReplicationFactor: make(map[int]int),
	}
}

// ComputeReplication computes replication factors for throughput balancing
func (r *WeightReplicator) ComputeReplication(layerMACs []int64) map[int]int {
	if len(layerMACs) == 0 {
		return nil
	}

	// Find bottleneck (layer with most MACs)
	maxMACs := layerMACs[0]
	for _, macs := range layerMACs {
		if macs > maxMACs {
			maxMACs = macs
		}
	}

	// Compute replication to balance throughput
	for i, macs := range layerMACs {
		if macs > 0 {
			factor := int(math.Ceil(float64(maxMACs) / float64(macs)))
			// Limit replication by available arrays
			maxFactor := r.Config.NumArrays / len(layerMACs)
			if factor > maxFactor {
				factor = maxFactor
			}
			if factor < 1 {
				factor = 1
			}
			r.ReplicationFactor[i] = factor
		} else {
			r.ReplicationFactor[i] = 1
		}
	}

	return r.ReplicationFactor
}

// ReplicateWeights creates replicated weight copies
func (r *WeightReplicator) ReplicateWeights(layerID int, weights [][]float64) [][][]float64 {
	factor, exists := r.ReplicationFactor[layerID]
	if !exists || factor <= 1 {
		return [][][]float64{weights}
	}

	replicated := make([][][]float64, factor)
	for i := 0; i < factor; i++ {
		// Deep copy
		copy := make([][]float64, len(weights))
		for j := range weights {
			copy[j] = make([]float64, len(weights[j]))
			for k := range weights[j] {
				copy[j][k] = weights[j][k]
			}
		}
		replicated[i] = copy
	}

	return replicated
}

// ============================================================================
// Partial Sum Accumulation
// ============================================================================

// PartialSumAccumulator handles partial sum accumulation from multiple arrays
type PartialSumAccumulator struct {
	PartialSums map[int][]float64 // output index -> accumulated values
	Counts      map[int]int       // number of contributions
}

// NewPartialSumAccumulator creates an accumulator
func NewPartialSumAccumulator() *PartialSumAccumulator {
	return &PartialSumAccumulator{
		PartialSums: make(map[int][]float64),
		Counts:      make(map[int]int),
	}
}

// Accumulate adds partial sums from an array
func (a *PartialSumAccumulator) Accumulate(outputIdx int, values []float64) {
	existing, exists := a.PartialSums[outputIdx]
	if !exists {
		a.PartialSums[outputIdx] = make([]float64, len(values))
		copy(a.PartialSums[outputIdx], values)
		a.Counts[outputIdx] = 1
	} else {
		for i := range values {
			if i < len(existing) {
				existing[i] += values[i]
			}
		}
		a.Counts[outputIdx]++
	}
}

// GetFinal returns final accumulated values
func (a *PartialSumAccumulator) GetFinal(outputIdx int) []float64 {
	return a.PartialSums[outputIdx]
}

// Reset clears all accumulated values
func (a *PartialSumAccumulator) Reset() {
	a.PartialSums = make(map[int][]float64)
	a.Counts = make(map[int]int)
}

// ============================================================================
// System-Level Orchestrator
// ============================================================================

// SystemOrchestrator coordinates the entire CIM system
type SystemOrchestrator struct {
	Config      *OrchestrationConfig
	ArrayMgr    *ArrayManager
	Pipeline    *ExecutionPipeline
	Replicator  *WeightReplicator
	Benchmarker *Benchmarker
}

// NewSystemOrchestrator creates a complete system orchestrator
func NewSystemOrchestrator(config *OrchestrationConfig) *SystemOrchestrator {
	if config == nil {
		config = DefaultOrchestrationConfig()
	}

	return &SystemOrchestrator{
		Config:      config,
		ArrayMgr:    NewArrayManager(config),
		Replicator:  NewWeightReplicator(config),
		Benchmarker: NewBenchmarker(config),
	}
}

// LoadModel loads a model onto the arrays
func (s *SystemOrchestrator) LoadModel(model *ModelCheckpoint) error {
	// Compute replication factors
	layerMACs := make([]int64, len(model.Layers))
	for i, layer := range model.Layers {
		if len(layer.Shape) >= 2 {
			layerMACs[i] = int64(layer.Shape[0] * layer.Shape[1])
		}
	}
	s.Replicator.ComputeReplication(layerMACs)

	// Map each layer to arrays
	for i, layer := range model.Layers {
		if len(layer.Weights) > 0 {
			err := s.ArrayMgr.MapLayer(i, layer.Weights)
			if err != nil {
				return fmt.Errorf("failed to map layer %d: %v", i, err)
			}
		}
	}

	// Create pipeline
	s.Pipeline = NewExecutionPipeline(s.Config, len(model.Layers))

	return nil
}

// RunInference runs inference through the orchestrated system
func (s *SystemOrchestrator) RunInference(inputs [][]float64) [][]float64 {
	if s.Pipeline == nil {
		return nil
	}

	return s.Pipeline.Execute(inputs, s.ArrayMgr)
}

// GetStats returns orchestration statistics
func (s *SystemOrchestrator) GetStats() *OrchestrationStats {
	stats := s.ArrayMgr.Stats

	// Calculate derived metrics
	if stats.TotalCycles > 0 {
		cycleTime := 1.0 / (s.Config.ClockFreqMHz * 1e6) // seconds
		totalTime := float64(stats.TotalCycles) * cycleTime
		stats.Throughput = float64(stats.TotalMACs) / totalTime / 1e12

		if stats.TotalEnergyFJ > 0 {
			power := stats.TotalEnergyFJ / 1e15 / totalTime
			stats.EnergyEfficiency = stats.Throughput / power
		}
	}

	return stats
}
