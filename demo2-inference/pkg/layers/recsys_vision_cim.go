// Package layers provides recommendation system and neuromorphic vision CIM simulation
// for compute-in-memory architectures.
//
// This module implements:
// - Deep Learning Recommendation Model (DLRM) embedding tables
// - ARCHER ReRAM-based PIM for recommendations
// - Event camera / DVS neuromorphic vision sensors
// - SPIDR-style digital CIM for event-based perception
//
// Based on research from HPCA 2025, Hot Chips 2024, and MDPI Sensors 2025.
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
)

// =============================================================================
// RECOMMENDATION SYSTEM EMBEDDING TABLES
// =============================================================================

// EmbeddingTableConfig configures an embedding table
type EmbeddingTableConfig struct {
	NumEmbeddings    int     // Vocabulary size
	EmbeddingDim     int     // Embedding dimension
	Sparsity         float64 // Access sparsity (0-1)
	CompressionRatio float64 // For decomposed models
	PoolingMode      string  // "sum", "mean", "max"
}

// DefaultEmbeddingTableConfig returns default embedding configuration
func DefaultEmbeddingTableConfig() *EmbeddingTableConfig {
	return &EmbeddingTableConfig{
		NumEmbeddings:    1000000, // 1M embeddings
		EmbeddingDim:     128,
		Sparsity:         0.99, // 99% sparse access
		CompressionRatio: 4.0,  // 4x compression
		PoolingMode:      "sum",
	}
}

// EmbeddingTable represents a standard embedding table
type EmbeddingTable struct {
	Config      *EmbeddingTableConfig
	Weights     [][]float32 // [NumEmbeddings][EmbeddingDim]
	AccessCount []int64     // Per-embedding access frequency
	TotalAccess int64
	mu          sync.RWMutex
}

// NewEmbeddingTable creates a new embedding table
func NewEmbeddingTable(config *EmbeddingTableConfig) *EmbeddingTable {
	et := &EmbeddingTable{
		Config:      config,
		Weights:     make([][]float32, config.NumEmbeddings),
		AccessCount: make([]int64, config.NumEmbeddings),
	}

	// Initialize with random embeddings
	scale := float32(1.0 / math.Sqrt(float64(config.EmbeddingDim)))
	for i := 0; i < config.NumEmbeddings; i++ {
		et.Weights[i] = make([]float32, config.EmbeddingDim)
		for j := 0; j < config.EmbeddingDim; j++ {
			et.Weights[i][j] = float32(rand.NormFloat64()) * scale
		}
	}

	return et
}

// Lookup retrieves embeddings for given indices
func (et *EmbeddingTable) Lookup(indices []int) [][]float32 {
	et.mu.Lock()
	defer et.mu.Unlock()

	result := make([][]float32, len(indices))
	for i, idx := range indices {
		if idx >= 0 && idx < et.Config.NumEmbeddings {
			result[i] = make([]float32, et.Config.EmbeddingDim)
			copy(result[i], et.Weights[idx])
			et.AccessCount[idx]++
			et.TotalAccess++
		}
	}
	return result
}

// LookupAndPool retrieves and pools embeddings
func (et *EmbeddingTable) LookupAndPool(indices []int) []float32 {
	embeddings := et.Lookup(indices)
	if len(embeddings) == 0 {
		return make([]float32, et.Config.EmbeddingDim)
	}

	result := make([]float32, et.Config.EmbeddingDim)

	switch et.Config.PoolingMode {
	case "sum":
		for _, emb := range embeddings {
			for j, v := range emb {
				result[j] += v
			}
		}
	case "mean":
		for _, emb := range embeddings {
			for j, v := range emb {
				result[j] += v
			}
		}
		scale := float32(1.0 / float64(len(embeddings)))
		for j := range result {
			result[j] *= scale
		}
	case "max":
		for j := range result {
			result[j] = float32(math.Inf(-1))
		}
		for _, emb := range embeddings {
			for j, v := range emb {
				if v > result[j] {
					result[j] = v
				}
			}
		}
	}

	return result
}

// GetHotEmbeddings returns most frequently accessed embeddings
func (et *EmbeddingTable) GetHotEmbeddings(topK int) []int {
	et.mu.RLock()
	defer et.mu.RUnlock()

	type indexCount struct {
		idx   int
		count int64
	}

	counts := make([]indexCount, et.Config.NumEmbeddings)
	for i := range counts {
		counts[i] = indexCount{i, et.AccessCount[i]}
	}

	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	result := make([]int, topK)
	for i := 0; i < topK && i < len(counts); i++ {
		result[i] = counts[i].idx
	}
	return result
}

// =============================================================================
// DLRM (DEEP LEARNING RECOMMENDATION MODEL)
// =============================================================================

// DLRMConfig configures a DLRM model
type DLRMConfig struct {
	NumSparseFeatures  int      // Number of categorical features
	NumDenseFeatures   int      // Number of continuous features
	EmbeddingDims      []int    // Embedding dimension per sparse feature
	VocabSizes         []int    // Vocabulary size per sparse feature
	BottomMLPDims      []int    // Bottom MLP dimensions
	TopMLPDims         []int    // Top MLP dimensions
	InteractionType    string   // "dot", "concat"
}

// DefaultDLRMConfig returns default DLRM configuration
func DefaultDLRMConfig() *DLRMConfig {
	return &DLRMConfig{
		NumSparseFeatures:  26,                      // Standard DLRM
		NumDenseFeatures:   13,                      // Standard DLRM
		EmbeddingDims:      []int{128, 128, 128},    // First 3 features
		VocabSizes:         []int{10000, 10000, 10000},
		BottomMLPDims:      []int{13, 512, 256, 128},
		TopMLPDims:         []int{512, 256, 1},
		InteractionType:    "dot",
	}
}

// DLRM implements Deep Learning Recommendation Model
type DLRM struct {
	Config          *DLRMConfig
	EmbeddingTables []*EmbeddingTable
	BottomMLP       *MLPLayer
	TopMLP          *MLPLayer
	Stats           *DLRMStats
}

// MLPLayer represents a multi-layer perceptron
type MLPLayer struct {
	Weights [][]float32
	Biases  [][]float32
	Dims    []int
}

// DLRMStats tracks DLRM statistics
type DLRMStats struct {
	TotalInferences     int64
	EmbeddingLookups    int64
	MLPOps              int64
	InteractionOps      int64
	MemoryAccessBytes   int64
	ComputeTimeMs       float64
	MemoryBoundPercent  float64
}

// NewDLRM creates a new DLRM model
func NewDLRM(config *DLRMConfig) *DLRM {
	dlrm := &DLRM{
		Config:          config,
		EmbeddingTables: make([]*EmbeddingTable, len(config.VocabSizes)),
		Stats:           &DLRMStats{},
	}

	// Create embedding tables
	for i := range config.VocabSizes {
		embConfig := &EmbeddingTableConfig{
			NumEmbeddings: config.VocabSizes[i],
			EmbeddingDim:  config.EmbeddingDims[i%len(config.EmbeddingDims)],
			PoolingMode:   "sum",
		}
		dlrm.EmbeddingTables[i] = NewEmbeddingTable(embConfig)
	}

	// Create MLPs
	dlrm.BottomMLP = newMLPLayer(config.BottomMLPDims)
	dlrm.TopMLP = newMLPLayer(config.TopMLPDims)

	return dlrm
}

// newMLPLayer creates a new MLP layer
func newMLPLayer(dims []int) *MLPLayer {
	mlp := &MLPLayer{
		Weights: make([][]float32, len(dims)-1),
		Biases:  make([][]float32, len(dims)-1),
		Dims:    dims,
	}

	for i := 0; i < len(dims)-1; i++ {
		inDim := dims[i]
		outDim := dims[i+1]
		scale := float32(math.Sqrt(2.0 / float64(inDim)))

		mlp.Weights[i] = make([]float32, inDim*outDim)
		for j := range mlp.Weights[i] {
			mlp.Weights[i][j] = float32(rand.NormFloat64()) * scale
		}

		mlp.Biases[i] = make([]float32, outDim)
	}

	return mlp
}

// Forward runs MLP forward pass
func (mlp *MLPLayer) Forward(input []float32) []float32 {
	current := input

	for layer := 0; layer < len(mlp.Dims)-1; layer++ {
		inDim := mlp.Dims[layer]
		outDim := mlp.Dims[layer+1]

		output := make([]float32, outDim)
		for j := 0; j < outDim; j++ {
			sum := mlp.Biases[layer][j]
			for i := 0; i < inDim && i < len(current); i++ {
				sum += current[i] * mlp.Weights[layer][i*outDim+j]
			}
			// ReLU activation (except last layer)
			if layer < len(mlp.Dims)-2 {
				if sum < 0 {
					sum = 0
				}
			}
			output[j] = sum
		}
		current = output
	}

	return current
}

// Inference runs DLRM inference
func (dlrm *DLRM) Inference(denseFeatures []float32, sparseIndices [][]int) float32 {
	dlrm.Stats.TotalInferences++

	// Bottom MLP on dense features
	denseEmb := dlrm.BottomMLP.Forward(denseFeatures)
	dlrm.Stats.MLPOps += int64(len(dlrm.Config.BottomMLPDims) * len(denseFeatures))

	// Embedding lookups
	sparseEmbs := make([][]float32, len(sparseIndices))
	for i, indices := range sparseIndices {
		if i < len(dlrm.EmbeddingTables) {
			sparseEmbs[i] = dlrm.EmbeddingTables[i].LookupAndPool(indices)
			dlrm.Stats.EmbeddingLookups += int64(len(indices))
			dlrm.Stats.MemoryAccessBytes += int64(len(indices) * dlrm.Config.EmbeddingDims[i%len(dlrm.Config.EmbeddingDims)] * 4)
		}
	}

	// Feature interaction
	var interactionOutput []float32
	if dlrm.Config.InteractionType == "dot" {
		// Dot product interaction
		allEmbs := append([][]float32{denseEmb}, sparseEmbs...)
		interactionOutput = dlrm.dotInteraction(allEmbs)
	} else {
		// Concatenation
		interactionOutput = denseEmb
		for _, emb := range sparseEmbs {
			interactionOutput = append(interactionOutput, emb...)
		}
	}

	// Top MLP
	output := dlrm.TopMLP.Forward(interactionOutput)
	dlrm.Stats.MLPOps += int64(len(dlrm.Config.TopMLPDims) * len(interactionOutput))

	// Sigmoid
	return float32(1.0 / (1.0 + math.Exp(-float64(output[0]))))
}

// dotInteraction computes pairwise dot products
func (dlrm *DLRM) dotInteraction(embeddings [][]float32) []float32 {
	n := len(embeddings)
	// Lower triangular dot products
	numInteractions := n * (n - 1) / 2

	result := make([]float32, len(embeddings[0])+numInteractions)

	// Concatenate first embedding
	copy(result, embeddings[0])

	// Pairwise dot products
	idx := len(embeddings[0])
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			dot := float32(0)
			minLen := len(embeddings[i])
			if len(embeddings[j]) < minLen {
				minLen = len(embeddings[j])
			}
			for k := 0; k < minLen; k++ {
				dot += embeddings[i][k] * embeddings[j][k]
			}
			if idx < len(result) {
				result[idx] = dot
				idx++
			}
			dlrm.Stats.InteractionOps++
		}
	}

	return result
}

// =============================================================================
// ARCHER - ReRAM-BASED PIM FOR RECOMMENDATIONS
// =============================================================================

// ARCHERConfig configures ARCHER PIM accelerator
type ARCHERConfig struct {
	NumCrossbarArrays int     // Number of ReRAM arrays
	ArrayRows         int     // Rows per array
	ArrayCols         int     // Columns per array
	ADCBits           int     // ADC precision
	DecompositionRank int     // Low-rank decomposition
	OnChipCapacityMB  int     // On-chip memory
	OffChipBandwidth  float64 // GB/s
}

// DefaultARCHERConfig returns default ARCHER configuration
func DefaultARCHERConfig() *ARCHERConfig {
	return &ARCHERConfig{
		NumCrossbarArrays: 64,
		ArrayRows:         256,
		ArrayCols:         256,
		ADCBits:           8,
		DecompositionRank: 32,
		OnChipCapacityMB:  64,
		OffChipBandwidth:  64.0, // GB/s
	}
}

// ARCHER implements ReRAM-based PIM for recommendations
// Based on: Frontiers of Computer Science 2024
type ARCHER struct {
	Config            *ARCHERConfig
	Crossbars         []*ReRAMCrossbar
	DecomposedWeights *DecomposedEmbedding
	Stats             *ARCHERStats
}

// ReRAMCrossbar represents a ReRAM crossbar array
type ReRAMCrossbar struct {
	Rows       int
	Cols       int
	Conductance [][]float32 // Stored as conductance
	ADCBits    int
}

// DecomposedEmbedding stores decomposed embedding weights
type DecomposedEmbedding struct {
	LeftMatrix  [][]float32 // [NumEmbeddings][Rank]
	RightMatrix [][]float32 // [Rank][EmbeddingDim]
	Rank        int
}

// ARCHERStats tracks ARCHER statistics
type ARCHERStats struct {
	OnChipLookups      int64
	OffChipLookups     int64
	CrossbarOps        int64
	DecompressionOps   int64
	EnergyPJ           float64
	LatencyNS          float64
	OnChipHitRate      float64
}

// NewARCHER creates a new ARCHER accelerator
func NewARCHER(config *ARCHERConfig) *ARCHER {
	archer := &ARCHER{
		Config:    config,
		Crossbars: make([]*ReRAMCrossbar, config.NumCrossbarArrays),
		Stats:     &ARCHERStats{},
	}

	for i := 0; i < config.NumCrossbarArrays; i++ {
		archer.Crossbars[i] = &ReRAMCrossbar{
			Rows:        config.ArrayRows,
			Cols:        config.ArrayCols,
			Conductance: make([][]float32, config.ArrayRows),
			ADCBits:     config.ADCBits,
		}
		for j := 0; j < config.ArrayRows; j++ {
			archer.Crossbars[i].Conductance[j] = make([]float32, config.ArrayCols)
		}
	}

	return archer
}

// DecomposeEmbedding performs low-rank decomposition
func (a *ARCHER) DecomposeEmbedding(table *EmbeddingTable) {
	// SVD-based decomposition: W ≈ L × R
	numEmb := table.Config.NumEmbeddings
	embDim := table.Config.EmbeddingDim
	rank := a.Config.DecompositionRank

	a.DecomposedWeights = &DecomposedEmbedding{
		LeftMatrix:  make([][]float32, numEmb),
		RightMatrix: make([][]float32, rank),
		Rank:        rank,
	}

	// Initialize with random projection (simplified)
	for i := 0; i < numEmb; i++ {
		a.DecomposedWeights.LeftMatrix[i] = make([]float32, rank)
		for j := 0; j < rank; j++ {
			a.DecomposedWeights.LeftMatrix[i][j] = float32(rand.NormFloat64()) * 0.1
		}
	}

	for i := 0; i < rank; i++ {
		a.DecomposedWeights.RightMatrix[i] = make([]float32, embDim)
		for j := 0; j < embDim; j++ {
			a.DecomposedWeights.RightMatrix[i][j] = float32(rand.NormFloat64()) * 0.1
		}
	}
}

// LookupDecomposed performs embedding lookup with decompression
func (a *ARCHER) LookupDecomposed(idx int) []float32 {
	if a.DecomposedWeights == nil {
		return nil
	}

	rank := a.DecomposedWeights.Rank
	embDim := len(a.DecomposedWeights.RightMatrix[0])

	// Get left vector
	leftVec := a.DecomposedWeights.LeftMatrix[idx%len(a.DecomposedWeights.LeftMatrix)]

	// Compute embedding: e = L[idx] × R
	embedding := make([]float32, embDim)
	for j := 0; j < embDim; j++ {
		sum := float32(0)
		for k := 0; k < rank; k++ {
			sum += leftVec[k] * a.DecomposedWeights.RightMatrix[k][j]
		}
		embedding[j] = sum
	}

	a.Stats.DecompressionOps += int64(rank * embDim)
	a.Stats.OnChipLookups++

	return embedding
}

// ComputeOnCrossbar performs MVM on ReRAM crossbar
func (a *ARCHER) ComputeOnCrossbar(arrayIdx int, input []float32) []float32 {
	if arrayIdx >= len(a.Crossbars) {
		return nil
	}

	xbar := a.Crossbars[arrayIdx]
	output := make([]float32, xbar.Rows)

	for i := 0; i < xbar.Rows; i++ {
		sum := float32(0)
		for j := 0; j < xbar.Cols && j < len(input); j++ {
			sum += xbar.Conductance[i][j] * input[j]
		}
		// ADC quantization
		maxVal := float32(1 << xbar.ADCBits)
		output[i] = float32(int(sum*maxVal)) / maxVal
	}

	a.Stats.CrossbarOps += int64(xbar.Rows * xbar.Cols)
	a.Stats.EnergyPJ += float64(xbar.Rows*xbar.Cols) * 0.1 // ~0.1 pJ/MAC
	a.Stats.LatencyNS += 100 // ~100ns per MVM

	return output
}

// =============================================================================
// META MTIA - TABLE BRANCH EMBEDDING
// =============================================================================

// MTIAConfig configures Meta MTIA accelerator
type MTIAConfig struct {
	NumCores          int     // Number of compute cores
	MemoryGB          int     // LPDDR5 memory
	TDPWatts          int     // Thermal design power
	TBEEnabled        bool    // Table Branch Embedding
	ProcessNode       string  // e.g., "TSMC 5nm"
}

// DefaultMTIAConfig returns default MTIA configuration
// Based on: Hot Chips 2024
func DefaultMTIAConfig() *MTIAConfig {
	return &MTIAConfig{
		NumCores:    64,
		MemoryGB:    128,
		TDPWatts:    90,
		TBEEnabled:  true,
		ProcessNode: "TSMC 5nm",
	}
}

// MTIA implements Meta Training and Inference Accelerator
type MTIA struct {
	Config        *MTIAConfig
	TBEUnits      []*TBEUnit
	EmbeddingCache map[int][]float32
	Stats         *MTIAStats
	mu            sync.RWMutex
}

// TBEUnit implements Table Branch Embedding
type TBEUnit struct {
	ID             int
	CacheSize      int
	HotEmbeddings  map[int][]float32
	AccessPattern  []int
}

// MTIAStats tracks MTIA statistics
type MTIAStats struct {
	TBEHits         int64
	TBEMisses       int64
	Speedup         float64 // vs baseline
	PowerEfficiency float64 // inferences/watt
}

// NewMTIA creates a new MTIA accelerator
func NewMTIA(config *MTIAConfig) *MTIA {
	mtia := &MTIA{
		Config:         config,
		TBEUnits:       make([]*TBEUnit, config.NumCores),
		EmbeddingCache: make(map[int][]float32),
		Stats:          &MTIAStats{},
	}

	for i := 0; i < config.NumCores; i++ {
		mtia.TBEUnits[i] = &TBEUnit{
			ID:            i,
			CacheSize:     1024,
			HotEmbeddings: make(map[int][]float32),
			AccessPattern: make([]int, 0),
		}
	}

	return mtia
}

// LookupWithTBE performs embedding lookup with TBE optimization
func (m *MTIA) LookupWithTBE(tableIdx, embIdx int, table *EmbeddingTable) []float32 {
	m.mu.RLock()
	cacheKey := tableIdx*1000000 + embIdx
	if cached, ok := m.EmbeddingCache[cacheKey]; ok {
		m.mu.RUnlock()
		m.Stats.TBEHits++
		return cached
	}
	m.mu.RUnlock()

	// Cache miss - fetch from table
	m.Stats.TBEMisses++
	embedding := table.LookupAndPool([]int{embIdx})

	// Update cache
	m.mu.Lock()
	m.EmbeddingCache[cacheKey] = embedding
	m.mu.Unlock()

	return embedding
}

// ComputeSpeedup calculates TBE speedup
func (m *MTIA) ComputeSpeedup() float64 {
	total := m.Stats.TBEHits + m.Stats.TBEMisses
	if total == 0 {
		return 1.0
	}
	hitRate := float64(m.Stats.TBEHits) / float64(total)
	// TBE provides 2-3x speedup at high hit rates
	m.Stats.Speedup = 1.0 + 2.0*hitRate
	return m.Stats.Speedup
}

// =============================================================================
// EVENT CAMERA / DVS NEUROMORPHIC VISION
// =============================================================================

// DVSConfig configures a Dynamic Vision Sensor
type DVSConfig struct {
	ResolutionX      int     // Horizontal resolution
	ResolutionY      int     // Vertical resolution
	TemporalResUS    float64 // Temporal resolution (µs)
	ContrastThreshold float64 // Log intensity change threshold
	RefractoryPeriod float64 // Minimum time between events (µs)
	DynamicRangeDB   float64 // Dynamic range in dB
	PowerMW          float64 // Power consumption (mW)
}

// DefaultDVSConfig returns default DVS configuration
// Based on: Prophesee Gen4, Samsung DVS
func DefaultDVSConfig() *DVSConfig {
	return &DVSConfig{
		ResolutionX:       1280,
		ResolutionY:       720,
		TemporalResUS:     1.0,    // 1 µs
		ContrastThreshold: 0.15,   // 15% log intensity change
		RefractoryPeriod:  1.0,    // 1 µs
		DynamicRangeDB:    120.0,  // 120 dB
		PowerMW:           100.0,  // 100 mW
	}
}

// Event represents a single DVS event
type Event struct {
	X         int     // X coordinate
	Y         int     // Y coordinate
	Timestamp float64 // Time in microseconds
	Polarity  int     // +1 (ON) or -1 (OFF)
}

// EventStream represents a stream of DVS events
type EventStream struct {
	Events     []Event
	StartTime  float64
	EndTime    float64
	EventRate  float64 // Events per second
}

// DVS implements a Dynamic Vision Sensor simulation
type DVS struct {
	Config          *DVSConfig
	PixelStates     [][]float64 // Log intensity per pixel
	LastEventTime   [][]float64 // Last event timestamp per pixel
	EventBuffer     []Event
	Stats           *DVSStats
	mu              sync.Mutex
}

// DVSStats tracks DVS statistics
type DVSStats struct {
	TotalEvents      int64
	ONEvents         int64
	OFFEvents        int64
	EventRate        float64 // Events/second
	DataReduction    float64 // vs frame-based
	AvgLatencyUS     float64
}

// NewDVS creates a new DVS sensor
func NewDVS(config *DVSConfig) *DVS {
	dvs := &DVS{
		Config:        config,
		PixelStates:   make([][]float64, config.ResolutionY),
		LastEventTime: make([][]float64, config.ResolutionY),
		EventBuffer:   make([]Event, 0),
		Stats:         &DVSStats{},
	}

	for y := 0; y < config.ResolutionY; y++ {
		dvs.PixelStates[y] = make([]float64, config.ResolutionX)
		dvs.LastEventTime[y] = make([]float64, config.ResolutionX)
	}

	return dvs
}

// ProcessFrame processes a frame and generates events
func (d *DVS) ProcessFrame(frame [][]float64, timestamp float64) []Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	events := make([]Event, 0)

	for y := 0; y < d.Config.ResolutionY && y < len(frame); y++ {
		for x := 0; x < d.Config.ResolutionX && x < len(frame[y]); x++ {
			// Compute log intensity
			intensity := frame[y][x]
			if intensity < 1e-10 {
				intensity = 1e-10
			}
			logIntensity := math.Log(intensity)

			// Check refractory period
			if timestamp-d.LastEventTime[y][x] < d.Config.RefractoryPeriod {
				continue
			}

			// Check contrast threshold
			diff := logIntensity - d.PixelStates[y][x]
			if math.Abs(diff) >= d.Config.ContrastThreshold {
				polarity := 1
				if diff < 0 {
					polarity = -1
					d.Stats.OFFEvents++
				} else {
					d.Stats.ONEvents++
				}

				event := Event{
					X:         x,
					Y:         y,
					Timestamp: timestamp,
					Polarity:  polarity,
				}
				events = append(events, event)
				d.Stats.TotalEvents++

				d.PixelStates[y][x] = logIntensity
				d.LastEventTime[y][x] = timestamp
			}
		}
	}

	d.EventBuffer = append(d.EventBuffer, events...)
	return events
}

// GetEventStream returns accumulated events as a stream
func (d *DVS) GetEventStream() *EventStream {
	d.mu.Lock()
	defer d.mu.Unlock()

	stream := &EventStream{
		Events: make([]Event, len(d.EventBuffer)),
	}
	copy(stream.Events, d.EventBuffer)

	if len(stream.Events) > 0 {
		stream.StartTime = stream.Events[0].Timestamp
		stream.EndTime = stream.Events[len(stream.Events)-1].Timestamp
		duration := stream.EndTime - stream.StartTime
		if duration > 0 {
			stream.EventRate = float64(len(stream.Events)) / (duration / 1e6)
		}
	}

	return stream
}

// ComputeDataReduction calculates data reduction vs frame-based
func (d *DVS) ComputeDataReduction(frameRate float64, duration float64) float64 {
	// Frame-based data: resolution × frames × bytes_per_pixel
	frameData := float64(d.Config.ResolutionX*d.Config.ResolutionY) * frameRate * duration * 1.0

	// Event-based data: events × bytes_per_event (typically 8 bytes)
	eventData := float64(d.Stats.TotalEvents) * 8.0

	if eventData > 0 {
		d.Stats.DataReduction = frameData / eventData
	}
	return d.Stats.DataReduction
}

// =============================================================================
// SPIDR - DIGITAL CIM FOR EVENT-BASED PERCEPTION
// =============================================================================

// SPIDRConfig configures SPIDR accelerator
type SPIDRConfig struct {
	NumPEs            int     // Processing elements
	SRAMBanks         int     // SRAM banks for CIM
	BitsPerWeight     int     // Weight precision
	BitsPerActivation int     // Activation precision
	ClockMHz          float64 // Clock frequency
	SpikingEnabled    bool    // Use spiking neurons
}

// DefaultSPIDRConfig returns default SPIDR configuration
// Based on: SPIDR (arXiv 2024)
func DefaultSPIDRConfig() *SPIDRConfig {
	return &SPIDRConfig{
		NumPEs:            64,
		SRAMBanks:         16,
		BitsPerWeight:     8,
		BitsPerActivation: 8,
		ClockMHz:          500.0,
		SpikingEnabled:    true,
	}
}

// SPIDR implements digital CIM for event-based perception
type SPIDR struct {
	Config       *SPIDRConfig
	PEs          []*SPIDRPE
	EventQueue   chan Event
	OutputSpikes []Spike
	Stats        *SPIDRStats
	mu           sync.Mutex
}

// SPIDRPE represents a SPIDR processing element
type SPIDRPE struct {
	ID              int
	Weights         [][]int8
	MembranePotential []float32
	Threshold       float32
	Decay           float32
}

// Spike represents an output spike
type Spike struct {
	NeuronID  int
	Timestamp float64
}

// SPIDRStats tracks SPIDR statistics
type SPIDRStats struct {
	EventsProcessed   int64
	SpikesGenerated   int64
	CIMOps            int64
	EnergyPJ          float64
	LatencyUS         float64
	Throughput        float64 // Events/second
	TOPSW             float64 // TOPS/W
}

// NewSPIDR creates a new SPIDR accelerator
func NewSPIDR(config *SPIDRConfig) *SPIDR {
	spidr := &SPIDR{
		Config:       config,
		PEs:          make([]*SPIDRPE, config.NumPEs),
		EventQueue:   make(chan Event, 10000),
		OutputSpikes: make([]Spike, 0),
		Stats:        &SPIDRStats{},
	}

	neuronsPerPE := 256
	inputsPerNeuron := 128

	for i := 0; i < config.NumPEs; i++ {
		pe := &SPIDRPE{
			ID:                i,
			Weights:           make([][]int8, neuronsPerPE),
			MembranePotential: make([]float32, neuronsPerPE),
			Threshold:         1.0,
			Decay:             0.9,
		}
		for j := 0; j < neuronsPerPE; j++ {
			pe.Weights[j] = make([]int8, inputsPerNeuron)
			for k := 0; k < inputsPerNeuron; k++ {
				pe.Weights[j][k] = int8(rand.Intn(256) - 128)
			}
		}
		spidr.PEs[i] = pe
	}

	return spidr
}

// ProcessEvent processes a single event through the network
func (s *SPIDR) ProcessEvent(event Event) []Spike {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Stats.EventsProcessed++
	spikes := make([]Spike, 0)

	// Map event to input neurons based on spatial location
	inputIdx := event.Y*s.Config.NumPEs + event.X
	peIdx := inputIdx % s.Config.NumPEs

	pe := s.PEs[peIdx]

	// Update membrane potentials using CIM
	for neuron := 0; neuron < len(pe.MembranePotential); neuron++ {
		// Decay
		pe.MembranePotential[neuron] *= pe.Decay

		// Accumulate input weighted by event polarity
		weightIdx := inputIdx % len(pe.Weights[neuron])
		weight := float32(pe.Weights[neuron][weightIdx]) / 128.0
		pe.MembranePotential[neuron] += weight * float32(event.Polarity)

		s.Stats.CIMOps++

		// Check threshold
		if pe.MembranePotential[neuron] >= pe.Threshold {
			spike := Spike{
				NeuronID:  peIdx*len(pe.MembranePotential) + neuron,
				Timestamp: event.Timestamp,
			}
			spikes = append(spikes, spike)
			pe.MembranePotential[neuron] = 0 // Reset

			s.Stats.SpikesGenerated++
		}
	}

	s.OutputSpikes = append(s.OutputSpikes, spikes...)
	return spikes
}

// ProcessEventStream processes a stream of events
func (s *SPIDR) ProcessEventStream(stream *EventStream) {
	for _, event := range stream.Events {
		s.ProcessEvent(event)
	}

	// Calculate statistics
	if len(stream.Events) > 0 {
		duration := stream.EndTime - stream.StartTime
		if duration > 0 {
			s.Stats.Throughput = float64(s.Stats.EventsProcessed) / (duration / 1e6)
			s.Stats.LatencyUS = duration / float64(len(stream.Events))
		}
	}

	// Estimate energy and efficiency
	s.Stats.EnergyPJ = float64(s.Stats.CIMOps) * 0.01 // ~10 fJ/op for digital CIM
	if s.Stats.EnergyPJ > 0 {
		ops := float64(s.Stats.CIMOps)
		energyJ := s.Stats.EnergyPJ * 1e-12
		s.Stats.TOPSW = (ops / 1e12) / energyJ
	}
}

// =============================================================================
// NEURO-CIM ACCELERATOR
// =============================================================================

// NeuroCIMConfig configures Neuro-CIM accelerator
type NeuroCIMConfig struct {
	ArraySize         int     // CIM array size
	NumArrays         int     // Number of arrays
	ADCBits           int     // ADC precision
	NeuronType        string  // "LIF", "IF", "analog"
	MixedMode         bool    // Digital-analog mixed mode
	TOPSW             float64 // Target efficiency
}

// DefaultNeuroCIMConfig returns default Neuro-CIM configuration
// Based on: Neuro-CIM (VLSI 2022), 310.4 TOPS/W
func DefaultNeuroCIMConfig() *NeuroCIMConfig {
	return &NeuroCIMConfig{
		ArraySize:  256,
		NumArrays:  64,
		ADCBits:    4,
		NeuronType: "LIF",
		MixedMode:  true,
		TOPSW:      310.4,
	}
}

// NeuroCIM implements neuromorphic CIM processor
type NeuroCIM struct {
	Config       *NeuroCIMConfig
	Arrays       []*CIMArray
	Neurons      []*LIFNeuron
	Stats        *NeuroCIMStats
}

// CIMArray represents a compute-in-memory array
type CIMArray struct {
	ID          int
	Size        int
	Weights     [][]int8
	WLActivity  float64 // Word line activity
	BLActivity  float64 // Bit line activity
}

// LIFNeuron represents a Leaky Integrate-and-Fire neuron
type LIFNeuron struct {
	ID                int
	MembranePotential float32
	Threshold         float32
	LeakRate          float32
	RefractoryTime    float64
	LastSpike         float64
}

// NeuroCIMStats tracks Neuro-CIM statistics
type NeuroCIMStats struct {
	TotalOps          int64
	SpikeCount        int64
	AvgWLActivity     float64
	AvgBLActivity     float64
	EnergyPJ          float64
	ActualTOPSW       float64
}

// NewNeuroCIM creates a new Neuro-CIM accelerator
func NewNeuroCIM(config *NeuroCIMConfig) *NeuroCIM {
	nc := &NeuroCIM{
		Config:  config,
		Arrays:  make([]*CIMArray, config.NumArrays),
		Neurons: make([]*LIFNeuron, config.NumArrays*config.ArraySize),
		Stats:   &NeuroCIMStats{},
	}

	for i := 0; i < config.NumArrays; i++ {
		nc.Arrays[i] = &CIMArray{
			ID:      i,
			Size:    config.ArraySize,
			Weights: make([][]int8, config.ArraySize),
		}
		for j := 0; j < config.ArraySize; j++ {
			nc.Arrays[i].Weights[j] = make([]int8, config.ArraySize)
			for k := 0; k < config.ArraySize; k++ {
				nc.Arrays[i].Weights[j][k] = int8(rand.Intn(256) - 128)
			}
		}
	}

	for i := range nc.Neurons {
		nc.Neurons[i] = &LIFNeuron{
			ID:                i,
			MembranePotential: 0,
			Threshold:         1.0,
			LeakRate:          0.1,
			RefractoryTime:    1.0,
			LastSpike:         -100,
		}
	}

	return nc
}

// ProcessSpikes processes input spikes through CIM arrays
func (nc *NeuroCIM) ProcessSpikes(inputSpikes []Spike, currentTime float64) []Spike {
	outputSpikes := make([]Spike, 0)

	// Count active inputs per array
	for _, array := range nc.Arrays {
		activeInputs := 0
		for _, spike := range inputSpikes {
			if spike.NeuronID%nc.Config.NumArrays == array.ID {
				activeInputs++
			}
		}
		array.WLActivity = float64(activeInputs) / float64(array.Size)
	}

	// Process each array
	for arrayIdx, array := range nc.Arrays {
		// Compute partial sums using CIM
		for _, spike := range inputSpikes {
			inputIdx := spike.NeuronID % array.Size

			for outIdx := 0; outIdx < array.Size; outIdx++ {
				neuronIdx := arrayIdx*array.Size + outIdx
				neuron := nc.Neurons[neuronIdx]

				// Check refractory period
				if currentTime-neuron.LastSpike < neuron.RefractoryTime {
					continue
				}

				// Leak
				neuron.MembranePotential -= neuron.LeakRate * neuron.MembranePotential

				// Integrate
				weight := float32(array.Weights[inputIdx][outIdx]) / 128.0
				neuron.MembranePotential += weight

				nc.Stats.TotalOps++

				// Fire
				if neuron.MembranePotential >= neuron.Threshold {
					outputSpikes = append(outputSpikes, Spike{
						NeuronID:  neuronIdx,
						Timestamp: currentTime,
					})
					neuron.MembranePotential = 0
					neuron.LastSpike = currentTime
					nc.Stats.SpikeCount++
				}
			}
		}

		// Track BL activity
		activeOutputs := 0
		for i := 0; i < array.Size; i++ {
			if nc.Neurons[arrayIdx*array.Size+i].MembranePotential > 0.1 {
				activeOutputs++
			}
		}
		array.BLActivity = float64(activeOutputs) / float64(array.Size)
	}

	// Update statistics
	totalWL := 0.0
	totalBL := 0.0
	for _, array := range nc.Arrays {
		totalWL += array.WLActivity
		totalBL += array.BLActivity
	}
	nc.Stats.AvgWLActivity = totalWL / float64(nc.Config.NumArrays)
	nc.Stats.AvgBLActivity = totalBL / float64(nc.Config.NumArrays)

	// Low activity = low energy
	activityFactor := (nc.Stats.AvgWLActivity + nc.Stats.AvgBLActivity) / 2.0
	nc.Stats.EnergyPJ = float64(nc.Stats.TotalOps) * 0.01 * activityFactor
	if nc.Stats.EnergyPJ > 0 {
		nc.Stats.ActualTOPSW = (float64(nc.Stats.TotalOps) / 1e12) / (nc.Stats.EnergyPJ * 1e-12)
	}

	return outputSpikes
}

// =============================================================================
// INTEGRATED VISION + RECOMMENDATION SYSTEM
// =============================================================================

// VisionRecSysConfig configures integrated vision-recommendation system
type VisionRecSysConfig struct {
	DVSConfig      *DVSConfig
	DLRMConfig     *DLRMConfig
	SPIDRConfig    *SPIDRConfig
	FusionMode     string // "early", "late", "hybrid"
}

// VisionRecSys integrates neuromorphic vision with recommendations
type VisionRecSys struct {
	Config       *VisionRecSysConfig
	DVS          *DVS
	SPIDR        *SPIDR
	DLRM         *DLRM
	Stats        *VisionRecSysStats
}

// VisionRecSysStats tracks integrated system statistics
type VisionRecSysStats struct {
	VisualEventsProcessed int64
	RecommendationsMade   int64
	FusionLatencyMS       float64
	TotalEnergyMJ         float64
}

// NewVisionRecSys creates a new integrated system
func NewVisionRecSys(config *VisionRecSysConfig) *VisionRecSys {
	return &VisionRecSys{
		Config: config,
		DVS:    NewDVS(config.DVSConfig),
		SPIDR:  NewSPIDR(config.SPIDRConfig),
		DLRM:   NewDLRM(config.DLRMConfig),
		Stats:  &VisionRecSysStats{},
	}
}

// ProcessVisualInput processes visual input for recommendations
func (vrs *VisionRecSys) ProcessVisualInput(frame [][]float64, timestamp float64, userFeatures []float32, itemIndices [][]int) float32 {
	// Process frame through DVS
	events := vrs.DVS.ProcessFrame(frame, timestamp)
	vrs.Stats.VisualEventsProcessed += int64(len(events))

	// Process events through SPIDR
	eventStream := &EventStream{Events: events}
	vrs.SPIDR.ProcessEventStream(eventStream)

	// Extract visual features from spike patterns
	visualFeatures := vrs.extractVisualFeatures()

	// Combine with user features for recommendation
	combinedFeatures := append(userFeatures, visualFeatures...)

	// Run DLRM inference
	score := vrs.DLRM.Inference(combinedFeatures, itemIndices)
	vrs.Stats.RecommendationsMade++

	return score
}

// extractVisualFeatures extracts features from spike patterns
func (vrs *VisionRecSys) extractVisualFeatures() []float32 {
	// Simple spike rate encoding
	features := make([]float32, vrs.Config.SPIDRConfig.NumPEs)

	spikeRates := make(map[int]int)
	for _, spike := range vrs.SPIDR.OutputSpikes {
		peIdx := spike.NeuronID / 256
		spikeRates[peIdx]++
	}

	maxRate := 1
	for _, rate := range spikeRates {
		if rate > maxRate {
			maxRate = rate
		}
	}

	for pe, rate := range spikeRates {
		if pe < len(features) {
			features[pe] = float32(rate) / float32(maxRate)
		}
	}

	return features
}

// =============================================================================
// DEMONSTRATION FUNCTIONS
// =============================================================================

// DemoRecommendationCIM demonstrates CIM for recommendation systems
func DemoRecommendationCIM() {
	fmt.Println("=== Recommendation System CIM Demo ===")
	fmt.Println()

	// Create DLRM
	dlrmConfig := DefaultDLRMConfig()
	dlrm := NewDLRM(dlrmConfig)

	// Create ARCHER accelerator
	archerConfig := DefaultARCHERConfig()
	archer := NewARCHER(archerConfig)

	// Decompose first embedding table
	archer.DecomposeEmbedding(dlrm.EmbeddingTables[0])

	fmt.Printf("DLRM Configuration:\n")
	fmt.Printf("  Sparse features: %d\n", dlrmConfig.NumSparseFeatures)
	fmt.Printf("  Dense features: %d\n", dlrmConfig.NumDenseFeatures)
	fmt.Printf("  Embedding dim: %d\n", dlrmConfig.EmbeddingDims[0])
	fmt.Println()

	// Run inference
	denseFeatures := make([]float32, dlrmConfig.NumDenseFeatures)
	for i := range denseFeatures {
		denseFeatures[i] = rand.Float32()
	}

	sparseIndices := make([][]int, len(dlrmConfig.VocabSizes))
	for i := range sparseIndices {
		sparseIndices[i] = []int{rand.Intn(dlrmConfig.VocabSizes[i])}
	}

	score := dlrm.Inference(denseFeatures, sparseIndices)

	fmt.Printf("DLRM Statistics:\n")
	fmt.Printf("  Prediction score: %.4f\n", score)
	fmt.Printf("  Embedding lookups: %d\n", dlrm.Stats.EmbeddingLookups)
	fmt.Printf("  MLP ops: %d\n", dlrm.Stats.MLPOps)
	fmt.Printf("  Memory access: %d bytes\n", dlrm.Stats.MemoryAccessBytes)
	fmt.Println()

	fmt.Printf("ARCHER Statistics:\n")
	fmt.Printf("  Decomposition rank: %d\n", archerConfig.DecompositionRank)
	fmt.Printf("  Compression ratio: %.1fx\n", float64(dlrmConfig.EmbeddingDims[0])/float64(archerConfig.DecompositionRank))
	fmt.Printf("  On-chip capacity: %d MB\n", archerConfig.OnChipCapacityMB)
}

// DemoNeuromorphicVision demonstrates neuromorphic vision processing
func DemoNeuromorphicVision() {
	fmt.Println("=== Neuromorphic Vision CIM Demo ===")
	fmt.Println()

	// Create DVS
	dvsConfig := DefaultDVSConfig()
	dvsConfig.ResolutionX = 128
	dvsConfig.ResolutionY = 128
	dvs := NewDVS(dvsConfig)

	// Create SPIDR accelerator
	spidrConfig := DefaultSPIDRConfig()
	spidr := NewSPIDR(spidrConfig)

	fmt.Printf("DVS Configuration:\n")
	fmt.Printf("  Resolution: %dx%d\n", dvsConfig.ResolutionX, dvsConfig.ResolutionY)
	fmt.Printf("  Temporal resolution: %.1f µs\n", dvsConfig.TemporalResUS)
	fmt.Printf("  Dynamic range: %.0f dB\n", dvsConfig.DynamicRangeDB)
	fmt.Println()

	// Simulate moving edge
	numFrames := 10
	for f := 0; f < numFrames; f++ {
		frame := make([][]float64, dvsConfig.ResolutionY)
		for y := 0; y < dvsConfig.ResolutionY; y++ {
			frame[y] = make([]float64, dvsConfig.ResolutionX)
			for x := 0; x < dvsConfig.ResolutionX; x++ {
				// Moving vertical edge
				edgePos := f * 10
				if x > edgePos && x < edgePos+5 {
					frame[y][x] = 1.0
				} else {
					frame[y][x] = 0.1
				}
			}
		}
		dvs.ProcessFrame(frame, float64(f)*1000) // 1ms per frame
	}

	// Get event stream and process
	stream := dvs.GetEventStream()
	spidr.ProcessEventStream(stream)

	// Calculate data reduction
	dvs.ComputeDataReduction(1000, float64(numFrames)/1000)

	fmt.Printf("DVS Statistics:\n")
	fmt.Printf("  Total events: %d\n", dvs.Stats.TotalEvents)
	fmt.Printf("  ON events: %d\n", dvs.Stats.ONEvents)
	fmt.Printf("  OFF events: %d\n", dvs.Stats.OFFEvents)
	fmt.Printf("  Data reduction: %.1fx vs frame-based\n", dvs.Stats.DataReduction)
	fmt.Println()

	fmt.Printf("SPIDR Statistics:\n")
	fmt.Printf("  Events processed: %d\n", spidr.Stats.EventsProcessed)
	fmt.Printf("  Spikes generated: %d\n", spidr.Stats.SpikesGenerated)
	fmt.Printf("  CIM ops: %d\n", spidr.Stats.CIMOps)
	fmt.Printf("  Efficiency: %.1f TOPS/W\n", spidr.Stats.TOPSW)
}

// DemoNeuroCIM demonstrates Neuro-CIM accelerator
func DemoNeuroCIM() {
	fmt.Println("=== Neuro-CIM Accelerator Demo ===")
	fmt.Println()

	config := DefaultNeuroCIMConfig()
	nc := NewNeuroCIM(config)

	fmt.Printf("Neuro-CIM Configuration:\n")
	fmt.Printf("  Array size: %d\n", config.ArraySize)
	fmt.Printf("  Num arrays: %d\n", config.NumArrays)
	fmt.Printf("  Target: %.1f TOPS/W\n", config.TOPSW)
	fmt.Println()

	// Generate random input spikes
	inputSpikes := make([]Spike, 100)
	for i := range inputSpikes {
		inputSpikes[i] = Spike{
			NeuronID:  rand.Intn(config.NumArrays * config.ArraySize),
			Timestamp: float64(i) * 0.1,
		}
	}

	// Process spikes
	outputSpikes := nc.ProcessSpikes(inputSpikes, 10.0)

	fmt.Printf("Processing Results:\n")
	fmt.Printf("  Input spikes: %d\n", len(inputSpikes))
	fmt.Printf("  Output spikes: %d\n", len(outputSpikes))
	fmt.Printf("  Total ops: %d\n", nc.Stats.TotalOps)
	fmt.Printf("  Avg WL activity: %.2f%%\n", nc.Stats.AvgWLActivity*100)
	fmt.Printf("  Avg BL activity: %.2f%%\n", nc.Stats.AvgBLActivity*100)
	fmt.Printf("  Actual efficiency: %.1f TOPS/W\n", nc.Stats.ActualTOPSW)
}
