// recsys_activation.go - Recommendation System CIM and Low-Power Activation Functions
// IronLattice Ferroelectric CIM Educational Simulation
//
// This module implements:
// 1. Recommendation System CIM Acceleration:
//    - DLRM (Deep Learning Recommendation Model) embedding tables
//    - Sparse embedding lookup on crossbar arrays
//    - ARCHER-style ReRAM-based PIM
//    - Embedding table compression and hashing
//
// 2. Low-Power Activation Functions:
//    - Hardware-efficient ReLU variants (D-ReLU)
//    - Piecewise linear sigmoid/tanh approximations
//    - GELU/Swish approximations for transformers
//    - Analog memristor-based activation circuits
//    - LUT-based activation implementations
//
// References:
// - ARCHER (FCS 2024): ReRAM-based recommendation accelerator
// - UpDLRM (DAC 2024): UPMEM PIM for DLRM
// - Nature Comms (2025): Nonlinear activation in analog crossbars
// - IEEE TCAS-II (2024): Hardware-friendly Swish approximation
// - MDPI Electronics (2024): GELU internal symmetry implementation

package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// PART 1: RECOMMENDATION SYSTEM CIM ACCELERATION
// =============================================================================

// EmbeddingTableConfig configures an embedding table
type EmbeddingTableConfig struct {
	Name           string
	NumEmbeddings  int     // Vocabulary size (e.g., 10M users)
	EmbeddingDim   int     // Dimension (e.g., 64, 128)
	SparseFeature  bool    // Is this a sparse categorical feature?
	CompressionType string // "none", "hash", "quotient-remainder", "bloom"
	HashBuckets    int     // For hashing-based compression
}

// EmbeddingTable represents an embedding table
type EmbeddingTable struct {
	Config      *EmbeddingTableConfig
	Weights     [][]float64 // [NumEmbeddings x EmbeddingDim]
	AccessCount []int64     // Hot/cold tracking per embedding
	MemoryBytes int64       // Memory footprint
}

// NewEmbeddingTable creates a new embedding table
func NewEmbeddingTable(config *EmbeddingTableConfig) *EmbeddingTable {
	if config == nil {
		config = &EmbeddingTableConfig{
			Name:          "default",
			NumEmbeddings: 10000,
			EmbeddingDim:  64,
			SparseFeature: true,
			CompressionType: "none",
		}
	}

	table := &EmbeddingTable{
		Config:      config,
		AccessCount: make([]int64, config.NumEmbeddings),
	}

	// Initialize weights
	effectiveSize := config.NumEmbeddings
	if config.CompressionType == "hash" && config.HashBuckets > 0 {
		effectiveSize = config.HashBuckets
	}

	table.Weights = make([][]float64, effectiveSize)
	scale := 1.0 / math.Sqrt(float64(config.EmbeddingDim))
	for i := 0; i < effectiveSize; i++ {
		table.Weights[i] = make([]float64, config.EmbeddingDim)
		for j := 0; j < config.EmbeddingDim; j++ {
			table.Weights[i][j] = rand.NormFloat64() * scale
		}
	}

	// Calculate memory
	table.MemoryBytes = int64(effectiveSize) * int64(config.EmbeddingDim) * 4 // 4 bytes per float32

	return table
}

// Lookup retrieves embeddings for given indices
func (et *EmbeddingTable) Lookup(indices []int) [][]float64 {
	embeddings := make([][]float64, len(indices))

	for i, idx := range indices {
		effectiveIdx := idx

		// Apply hashing if compression enabled
		if et.Config.CompressionType == "hash" && et.Config.HashBuckets > 0 {
			effectiveIdx = idx % et.Config.HashBuckets
		}

		if effectiveIdx < 0 || effectiveIdx >= len(et.Weights) {
			effectiveIdx = 0 // Default to first embedding
		}

		embeddings[i] = make([]float64, et.Config.EmbeddingDim)
		copy(embeddings[i], et.Weights[effectiveIdx])
		et.AccessCount[idx%len(et.AccessCount)]++
	}

	return embeddings
}

// PoolEmbeddings combines multiple embeddings (for multi-hot encoding)
func (et *EmbeddingTable) PoolEmbeddings(embeddings [][]float64, mode string) []float64 {
	if len(embeddings) == 0 {
		return make([]float64, et.Config.EmbeddingDim)
	}

	result := make([]float64, et.Config.EmbeddingDim)

	switch mode {
	case "sum":
		for _, emb := range embeddings {
			for j := 0; j < et.Config.EmbeddingDim; j++ {
				result[j] += emb[j]
			}
		}
	case "mean":
		for _, emb := range embeddings {
			for j := 0; j < et.Config.EmbeddingDim; j++ {
				result[j] += emb[j]
			}
		}
		for j := range result {
			result[j] /= float64(len(embeddings))
		}
	case "max":
		copy(result, embeddings[0])
		for i := 1; i < len(embeddings); i++ {
			for j := 0; j < et.Config.EmbeddingDim; j++ {
				if embeddings[i][j] > result[j] {
					result[j] = embeddings[i][j]
				}
			}
		}
	}

	return result
}

// =============================================================================
// DLRM Model
// =============================================================================

// DLRMConfig configures a DLRM model
type DLRMConfig struct {
	// Dense features (continuous)
	DenseInputDim     int
	BottomMLPLayers   []int // e.g., [512, 256, 64]

	// Sparse features (categorical)
	EmbeddingConfigs  []*EmbeddingTableConfig

	// Top MLP
	TopMLPLayers      []int // e.g., [512, 256, 1]

	// Interaction
	InteractionType   string // "dot", "cat"
}

// DefaultDLRMConfig returns a typical DLRM configuration
func DefaultDLRMConfig() *DLRMConfig {
	return &DLRMConfig{
		DenseInputDim:   13,
		BottomMLPLayers: []int{512, 256, 64},
		EmbeddingConfigs: []*EmbeddingTableConfig{
			{Name: "user_id", NumEmbeddings: 1000000, EmbeddingDim: 64, SparseFeature: true},
			{Name: "item_id", NumEmbeddings: 100000, EmbeddingDim: 64, SparseFeature: true},
			{Name: "category", NumEmbeddings: 1000, EmbeddingDim: 64, SparseFeature: true},
		},
		TopMLPLayers:    []int{512, 256, 1},
		InteractionType: "dot",
	}
}

// DLRMModel represents a DLRM model
type DLRMModel struct {
	Config          *DLRMConfig
	EmbeddingTables []*EmbeddingTable
	BottomMLP       [][]float64 // Simplified: just weight matrices
	TopMLP          [][]float64

	// Statistics
	TotalParameters int64
	EmbeddingParams int64
	MLPParams       int64
}

// NewDLRMModel creates a new DLRM model
func NewDLRMModel(config *DLRMConfig) *DLRMModel {
	if config == nil {
		config = DefaultDLRMConfig()
	}

	model := &DLRMModel{
		Config:          config,
		EmbeddingTables: make([]*EmbeddingTable, len(config.EmbeddingConfigs)),
	}

	// Create embedding tables
	for i, ec := range config.EmbeddingConfigs {
		model.EmbeddingTables[i] = NewEmbeddingTable(ec)
		model.EmbeddingParams += model.EmbeddingTables[i].MemoryBytes / 4
	}

	// Create bottom MLP (simplified)
	prevDim := config.DenseInputDim
	for _, dim := range config.BottomMLPLayers {
		model.BottomMLP = append(model.BottomMLP, make([]float64, prevDim*dim))
		model.MLPParams += int64(prevDim * dim)
		prevDim = dim
	}

	// Create top MLP (simplified)
	// Input: num_embeddings * (num_embeddings+1) / 2 for dot interaction + dense
	numEmb := len(config.EmbeddingConfigs) + 1 // +1 for dense bottom output
	interactionDim := numEmb * (numEmb + 1) / 2
	prevDim = interactionDim

	for _, dim := range config.TopMLPLayers {
		model.TopMLP = append(model.TopMLP, make([]float64, prevDim*dim))
		model.MLPParams += int64(prevDim * dim)
		prevDim = dim
	}

	model.TotalParameters = model.EmbeddingParams + model.MLPParams

	return model
}

// GetMemoryBreakdown returns memory usage breakdown
func (model *DLRMModel) GetMemoryBreakdown() map[string]int64 {
	breakdown := make(map[string]int64)

	for i, table := range model.EmbeddingTables {
		breakdown[fmt.Sprintf("embedding_%d_%s", i, table.Config.Name)] = table.MemoryBytes
	}

	breakdown["bottom_mlp"] = 0
	for _, w := range model.BottomMLP {
		breakdown["bottom_mlp"] += int64(len(w)) * 4
	}

	breakdown["top_mlp"] = 0
	for _, w := range model.TopMLP {
		breakdown["top_mlp"] += int64(len(w)) * 4
	}

	return breakdown
}

// =============================================================================
// ReRAM-based PIM for Embeddings (ARCHER-style)
// =============================================================================

// ARCHERConfig configures ARCHER-style PIM accelerator
type ARCHERConfig struct {
	CrossbarRows      int
	CrossbarCols      int
	NumCrossbars      int
	NumTiles          int
	OnChipBufferKB    int
	CompressionRatio  float64 // Embedding compression
	HashCollisions    bool    // Handle hash collisions
}

// DefaultARCHERConfig returns default ARCHER settings
func DefaultARCHERConfig() *ARCHERConfig {
	return &ARCHERConfig{
		CrossbarRows:     256,
		CrossbarCols:     256,
		NumCrossbars:     64,
		NumTiles:         4,
		OnChipBufferKB:   512,
		CompressionRatio: 10.0, // 10x compression target
		HashCollisions:   true,
	}
}

// ARCHERAccelerator implements ARCHER-style ReRAM PIM
type ARCHERAccelerator struct {
	Config              *ARCHERConfig
	EmbeddingCrossbars  [][][]float64 // Embeddings stored in crossbars
	LookupBuffer        [][]float64   // On-chip buffer for lookups
	CollisionTable      map[int][]int // Hash collision handling

	// Performance metrics
	OnChipHits          int64
	OffChipAccesses     int64
	TotalLookups        int64
	SpeedupOverGPU      float64
}

// NewARCHERAccelerator creates a new ARCHER accelerator
func NewARCHERAccelerator(config *ARCHERConfig) *ARCHERAccelerator {
	if config == nil {
		config = DefaultARCHERConfig()
	}

	acc := &ARCHERAccelerator{
		Config:         config,
		CollisionTable: make(map[int][]int),
		SpeedupOverGPU: 3.5, // Reported speedup
	}

	// Initialize crossbars
	acc.EmbeddingCrossbars = make([][][]float64, config.NumCrossbars)
	for i := 0; i < config.NumCrossbars; i++ {
		acc.EmbeddingCrossbars[i] = make([][]float64, config.CrossbarRows)
		for j := 0; j < config.CrossbarRows; j++ {
			acc.EmbeddingCrossbars[i][j] = make([]float64, config.CrossbarCols)
		}
	}

	// Initialize on-chip buffer
	bufferEntries := config.OnChipBufferKB * 1024 / (config.CrossbarCols * 4)
	acc.LookupBuffer = make([][]float64, bufferEntries)
	for i := range acc.LookupBuffer {
		acc.LookupBuffer[i] = make([]float64, config.CrossbarCols)
	}

	return acc
}

// MapEmbeddingTable maps an embedding table to crossbars
func (acc *ARCHERAccelerator) MapEmbeddingTable(table *EmbeddingTable) error {
	numEmbeddings := len(table.Weights)
	embDim := table.Config.EmbeddingDim

	// Calculate required crossbars
	embsPerCrossbar := acc.Config.CrossbarRows
	crossbarsForDim := (embDim + acc.Config.CrossbarCols - 1) / acc.Config.CrossbarCols
	totalCrossbars := ((numEmbeddings + embsPerCrossbar - 1) / embsPerCrossbar) * crossbarsForDim

	if totalCrossbars > acc.Config.NumCrossbars {
		// Need compression
		compressionNeeded := float64(totalCrossbars) / float64(acc.Config.NumCrossbars)
		if compressionNeeded > acc.Config.CompressionRatio {
			return fmt.Errorf("embedding table too large: need %.1fx compression, max %.1fx",
				compressionNeeded, acc.Config.CompressionRatio)
		}
	}

	// Map embeddings to crossbars (simplified)
	crossbarIdx := 0
	rowIdx := 0
	for embIdx := 0; embIdx < numEmbeddings && crossbarIdx < acc.Config.NumCrossbars; embIdx++ {
		for d := 0; d < embDim && d < acc.Config.CrossbarCols; d++ {
			acc.EmbeddingCrossbars[crossbarIdx][rowIdx][d] = table.Weights[embIdx][d]
		}
		rowIdx++
		if rowIdx >= acc.Config.CrossbarRows {
			rowIdx = 0
			crossbarIdx++
		}
	}

	return nil
}

// LookupEmbedding performs embedding lookup on crossbar
func (acc *ARCHERAccelerator) LookupEmbedding(indices []int, embDim int) [][]float64 {
	results := make([][]float64, len(indices))
	acc.TotalLookups += int64(len(indices))

	for i, idx := range indices {
		results[i] = make([]float64, embDim)

		// Check on-chip buffer first
		bufferIdx := idx % len(acc.LookupBuffer)
		if acc.isInBuffer(bufferIdx) {
			copy(results[i], acc.LookupBuffer[bufferIdx][:embDim])
			acc.OnChipHits++
		} else {
			// Off-chip access to crossbar
			crossbarIdx := idx / acc.Config.CrossbarRows
			rowIdx := idx % acc.Config.CrossbarRows

			if crossbarIdx < len(acc.EmbeddingCrossbars) &&
				rowIdx < len(acc.EmbeddingCrossbars[crossbarIdx]) {
				copy(results[i], acc.EmbeddingCrossbars[crossbarIdx][rowIdx][:embDim])
				// Update buffer
				copy(acc.LookupBuffer[bufferIdx], results[i])
			}
			acc.OffChipAccesses++
		}
	}

	return results
}

func (acc *ARCHERAccelerator) isInBuffer(idx int) bool {
	// Simplified: check if any non-zero values
	for _, v := range acc.LookupBuffer[idx] {
		if v != 0 {
			return true
		}
	}
	return false
}

// GetMetrics returns performance metrics
func (acc *ARCHERAccelerator) GetMetrics() map[string]float64 {
	hitRate := float64(acc.OnChipHits) / float64(acc.TotalLookups+1) * 100
	return map[string]float64{
		"on_chip_hit_rate":   hitRate,
		"off_chip_accesses":  float64(acc.OffChipAccesses),
		"total_lookups":      float64(acc.TotalLookups),
		"speedup_over_gpu":   acc.SpeedupOverGPU,
	}
}

// =============================================================================
// PART 2: LOW-POWER ACTIVATION FUNCTIONS
// =============================================================================

// ActivationType defines activation function types
type ActivationType int

const (
	ActivationReLU ActivationType = iota
	ActivationLeakyReLU
	ActivationDReLU       // Dynamic ReLU
	ActivationSigmoid
	ActivationTanh
	ActivationGELU
	ActivationSwish
	ActivationSiLU
	ActivationHardSigmoid
	ActivationHardSwish
)

// ActivationConfig configures an activation function
type ActivationConfig struct {
	Type              ActivationType
	ApproximationType string  // "exact", "piecewise", "lut", "analog"
	NumPieces         int     // For piecewise approximation
	LUTSize           int     // For LUT-based implementation
	BitPrecision      int     // Hardware bit width
	LeakySlope        float64 // For LeakyReLU
}

// DefaultActivationConfig returns default activation settings
func DefaultActivationConfig(actType ActivationType) *ActivationConfig {
	return &ActivationConfig{
		Type:              actType,
		ApproximationType: "piecewise",
		NumPieces:         16,
		LUTSize:           256,
		BitPrecision:      8,
		LeakySlope:        0.01,
	}
}

// ActivationUnit implements hardware-efficient activation functions
type ActivationUnit struct {
	Config          *ActivationConfig
	LUT             []float64         // Lookup table
	PiecewiseParams []PiecewiseSegment // Piecewise linear params

	// Performance
	AreaReduction   float64 // vs full sigmoid
	PowerReduction  float64 // vs full sigmoid
	AccuracyLoss    float64 // % accuracy loss
}

// PiecewiseSegment defines a segment for piecewise approximation
type PiecewiseSegment struct {
	XStart float64
	XEnd   float64
	Slope  float64
	Offset float64
}

// NewActivationUnit creates a new activation unit
func NewActivationUnit(config *ActivationConfig) *ActivationUnit {
	if config == nil {
		config = DefaultActivationConfig(ActivationReLU)
	}

	unit := &ActivationUnit{
		Config: config,
	}

	// Initialize based on approximation type
	switch config.ApproximationType {
	case "lut":
		unit.initLUT()
	case "piecewise":
		unit.initPiecewise()
	}

	// Set hardware metrics based on activation type
	switch config.Type {
	case ActivationReLU:
		unit.AreaReduction = 1.0    // Baseline
		unit.PowerReduction = 1.0
		unit.AccuracyLoss = 0.0
	case ActivationDReLU:
		unit.AreaReduction = 0.7    // 30% less area than sigmoid
		unit.PowerReduction = 0.65
		unit.AccuracyLoss = 0.0
	case ActivationSigmoid:
		if config.ApproximationType == "piecewise" {
			unit.AreaReduction = 0.4  // 60% reduction with piecewise
			unit.PowerReduction = 0.5
			unit.AccuracyLoss = 0.5
		} else {
			unit.AreaReduction = 1.0
			unit.PowerReduction = 1.0
			unit.AccuracyLoss = 0.0
		}
	case ActivationGELU:
		if config.ApproximationType == "piecewise" {
			unit.AreaReduction = 0.5
			unit.PowerReduction = 0.3 // 70% power reduction
			unit.AccuracyLoss = 0.1
		}
	case ActivationHardSigmoid:
		unit.AreaReduction = 0.2
		unit.PowerReduction = 0.15
		unit.AccuracyLoss = 1.0
	case ActivationHardSwish:
		unit.AreaReduction = 0.25
		unit.PowerReduction = 0.2
		unit.AccuracyLoss = 0.5
	}

	return unit
}

// initLUT initializes the lookup table
func (unit *ActivationUnit) initLUT() {
	size := unit.Config.LUTSize
	unit.LUT = make([]float64, size)

	// Range: typically [-8, 8] for most activations
	xMin, xMax := -8.0, 8.0
	step := (xMax - xMin) / float64(size-1)

	for i := 0; i < size; i++ {
		x := xMin + float64(i)*step
		unit.LUT[i] = unit.computeExact(x)
	}
}

// initPiecewise initializes piecewise linear approximation
func (unit *ActivationUnit) initPiecewise() {
	numPieces := unit.Config.NumPieces
	unit.PiecewiseParams = make([]PiecewiseSegment, numPieces)

	// Range for approximation
	xMin, xMax := -8.0, 8.0
	pieceWidth := (xMax - xMin) / float64(numPieces)

	for i := 0; i < numPieces; i++ {
		xStart := xMin + float64(i)*pieceWidth
		xEnd := xStart + pieceWidth

		// Compute slope and offset for linear fit
		y1 := unit.computeExact(xStart)
		y2 := unit.computeExact(xEnd)

		slope := (y2 - y1) / pieceWidth
		offset := y1 - slope*xStart

		unit.PiecewiseParams[i] = PiecewiseSegment{
			XStart: xStart,
			XEnd:   xEnd,
			Slope:  slope,
			Offset: offset,
		}
	}
}

// computeExact computes the exact activation value
func (unit *ActivationUnit) computeExact(x float64) float64 {
	switch unit.Config.Type {
	case ActivationReLU:
		if x > 0 {
			return x
		}
		return 0
	case ActivationLeakyReLU:
		if x > 0 {
			return x
		}
		return unit.Config.LeakySlope * x
	case ActivationDReLU:
		// Dynamic ReLU: adaptive threshold
		threshold := 0.0
		if x > threshold {
			return x
		}
		return 0.1 * x
	case ActivationSigmoid:
		return 1.0 / (1.0 + math.Exp(-x))
	case ActivationTanh:
		return math.Tanh(x)
	case ActivationGELU:
		// GELU = x * Φ(x) ≈ x * sigmoid(1.702 * x)
		return x * (1.0 / (1.0 + math.Exp(-1.702*x)))
	case ActivationSwish:
		// Swish = x * sigmoid(x)
		return x / (1.0 + math.Exp(-x))
	case ActivationSiLU:
		// SiLU = Swish with β=1
		return x / (1.0 + math.Exp(-x))
	case ActivationHardSigmoid:
		// HardSigmoid = clip((x + 3) / 6, 0, 1)
		y := (x + 3.0) / 6.0
		if y < 0 {
			return 0
		}
		if y > 1 {
			return 1
		}
		return y
	case ActivationHardSwish:
		// HardSwish = x * HardSigmoid(x)
		hs := (x + 3.0) / 6.0
		if hs < 0 {
			hs = 0
		} else if hs > 1 {
			hs = 1
		}
		return x * hs
	default:
		return x
	}
}

// Apply applies the activation function
func (unit *ActivationUnit) Apply(x float64) float64 {
	switch unit.Config.ApproximationType {
	case "exact":
		return unit.computeExact(x)
	case "lut":
		return unit.applyLUT(x)
	case "piecewise":
		return unit.applyPiecewise(x)
	default:
		return unit.computeExact(x)
	}
}

// applyLUT applies activation using lookup table
func (unit *ActivationUnit) applyLUT(x float64) float64 {
	xMin, xMax := -8.0, 8.0

	// Clamp to range
	if x <= xMin {
		return unit.LUT[0]
	}
	if x >= xMax {
		return unit.LUT[len(unit.LUT)-1]
	}

	// Linear interpolation
	normalized := (x - xMin) / (xMax - xMin)
	idx := normalized * float64(len(unit.LUT)-1)
	idxLow := int(idx)
	idxHigh := idxLow + 1
	if idxHigh >= len(unit.LUT) {
		idxHigh = len(unit.LUT) - 1
	}

	frac := idx - float64(idxLow)
	return unit.LUT[idxLow]*(1-frac) + unit.LUT[idxHigh]*frac
}

// applyPiecewise applies piecewise linear approximation
func (unit *ActivationUnit) applyPiecewise(x float64) float64 {
	// Find the right segment
	for _, seg := range unit.PiecewiseParams {
		if x >= seg.XStart && x < seg.XEnd {
			return seg.Slope*x + seg.Offset
		}
	}

	// Handle boundaries
	if x < unit.PiecewiseParams[0].XStart {
		seg := unit.PiecewiseParams[0]
		return seg.Slope*x + seg.Offset
	}
	seg := unit.PiecewiseParams[len(unit.PiecewiseParams)-1]
	return seg.Slope*x + seg.Offset
}

// ApplyBatch applies activation to a batch of values
func (unit *ActivationUnit) ApplyBatch(inputs []float64) []float64 {
	outputs := make([]float64, len(inputs))
	for i, x := range inputs {
		outputs[i] = unit.Apply(x)
	}
	return outputs
}

// =============================================================================
// GELU Hardware Implementation (ISPA - Internal Symmetry Piecewise Approximation)
// =============================================================================

// GELUISPAConfig configures ISPA-based GELU
type GELUISPAConfig struct {
	NumPositivePieces int     // Pieces for positive x
	MaxX              float64 // Maximum input value
	BitPrecision      int
}

// GELUISPAUnit implements GELU using internal symmetry
type GELUISPAUnit struct {
	Config           *GELUISPAConfig
	PositiveSegments []PiecewiseSegment

	// Performance (vs standard GELU)
	AreaReduction    float64 // ~50%
	PowerReduction   float64 // ~70%
	MSE              float64 // Mean squared error
}

// NewGELUISPAUnit creates an ISPA-based GELU unit
func NewGELUISPAUnit(config *GELUISPAConfig) *GELUISPAUnit {
	if config == nil {
		config = &GELUISPAConfig{
			NumPositivePieces: 16,
			MaxX:              4.0,
			BitPrecision:      16,
		}
	}

	unit := &GELUISPAUnit{
		Config:         config,
		AreaReduction:  0.5,
		PowerReduction: 0.7,
	}

	// Initialize positive axis piecewise approximation
	unit.initPositiveSegments()
	unit.calculateMSE()

	return unit
}

// initPositiveSegments initializes segments for positive x
func (unit *GELUISPAUnit) initPositiveSegments() {
	numPieces := unit.Config.NumPositivePieces
	maxX := unit.Config.MaxX
	unit.PositiveSegments = make([]PiecewiseSegment, numPieces)

	pieceWidth := maxX / float64(numPieces)

	for i := 0; i < numPieces; i++ {
		xStart := float64(i) * pieceWidth
		xEnd := xStart + pieceWidth

		// GELU(x) = x * Φ(x) where Φ is standard normal CDF
		y1 := unit.geluExact(xStart)
		y2 := unit.geluExact(xEnd)

		slope := (y2 - y1) / pieceWidth
		offset := y1 - slope*xStart

		unit.PositiveSegments[i] = PiecewiseSegment{
			XStart: xStart,
			XEnd:   xEnd,
			Slope:  slope,
			Offset: offset,
		}
	}
}

func (unit *GELUISPAUnit) geluExact(x float64) float64 {
	// GELU(x) = x * Φ(x) ≈ 0.5 * x * (1 + tanh(sqrt(2/π) * (x + 0.044715 * x³)))
	return 0.5 * x * (1 + math.Tanh(math.Sqrt(2/math.Pi)*(x+0.044715*x*x*x)))
}

// Apply applies ISPA-GELU
func (unit *GELUISPAUnit) Apply(x float64) float64 {
	// Use internal symmetry: GELU(-x) = -x + GELU(x) for the positive part
	if x >= 0 {
		return unit.applyPositive(x)
	}

	// For negative x: GELU(-|x|) = -|x| * Φ(-|x|) = -|x| * (1 - Φ(|x|))
	// Using symmetry: GELU(-x) ≈ -x + GELU(x) - x (approximation)
	absX := -x
	positiveResult := unit.applyPositive(absX)
	// GELU(-x) = -x * (1 - Φ(x)) ≈ -x + x * Φ(x) = -x + positiveResult (for large x)
	// More accurate: GELU(-x) = -x * Φ(-x) = -x * (1 - Φ(x))
	return -absX * (1 - positiveResult/absX) // Approximation using symmetry
}

func (unit *GELUISPAUnit) applyPositive(x float64) float64 {
	if x >= unit.Config.MaxX {
		return x // GELU(x) ≈ x for large positive x
	}

	for _, seg := range unit.PositiveSegments {
		if x >= seg.XStart && x < seg.XEnd {
			return seg.Slope*x + seg.Offset
		}
	}

	// Last segment
	seg := unit.PositiveSegments[len(unit.PositiveSegments)-1]
	return seg.Slope*x + seg.Offset
}

func (unit *GELUISPAUnit) calculateMSE() {
	// Calculate MSE over test points
	numPoints := 1000
	sumSquaredError := 0.0

	for i := 0; i < numPoints; i++ {
		x := -4.0 + 8.0*float64(i)/float64(numPoints)
		exact := unit.geluExact(x)
		approx := unit.Apply(x)
		err := exact - approx
		sumSquaredError += err * err
	}

	unit.MSE = sumSquaredError / float64(numPoints)
}

// =============================================================================
// Analog Memristor-Based Activation (Nature Comms 2025)
// =============================================================================

// AnalogActivationConfig configures analog activation circuit
type AnalogActivationConfig struct {
	NumMemristors     int     // Memristors for nonlinear ramp
	ConductanceRange  [2]float64 // Min/max conductance (µS)
	RampClockCycles   int     // Clock cycles for ramp ADC
	TargetFunction    ActivationType
}

// AnalogActivationUnit implements memristor-based analog activation
type AnalogActivationUnit struct {
	Config              *AnalogActivationConfig
	MemristorConductances []float64 // Programmed for inverse function
	RampVoltages        []float64   // Generated ramp

	// Performance vs digital
	AreaSaving          float64 // % area saved
	EnergySaving        float64 // % energy saved
	Latency             float64 // ns per activation
}

// NewAnalogActivationUnit creates an analog activation unit
func NewAnalogActivationUnit(config *AnalogActivationConfig) *AnalogActivationUnit {
	if config == nil {
		config = &AnalogActivationConfig{
			NumMemristors:    32,
			ConductanceRange: [2]float64{1.0, 100.0}, // 1-100 µS
			RampClockCycles:  32,
			TargetFunction:   ActivationSigmoid,
		}
	}

	unit := &AnalogActivationUnit{
		Config:     config,
		AreaSaving: 40.0,   // ~40% area saving
		EnergySaving: 60.0, // ~60% energy saving
		Latency:    10.0,   // 10 ns typical
	}

	unit.programMemristors()

	return unit
}

// programMemristors programs memristor conductances for inverse function
func (unit *AnalogActivationUnit) programMemristors() {
	n := unit.Config.NumMemristors
	unit.MemristorConductances = make([]float64, n)
	unit.RampVoltages = make([]float64, n)

	gMin := unit.Config.ConductanceRange[0]
	gMax := unit.Config.ConductanceRange[1]

	// Generate inverse function shape for conductances
	for i := 0; i < n; i++ {
		// Time step
		t := float64(i) / float64(n-1)

		// For sigmoid: inverse is logit, shape the ramp accordingly
		var targetVoltage float64
		switch unit.Config.TargetFunction {
		case ActivationSigmoid:
			// Ramp should follow inverse sigmoid (logit) shape
			y := t // Output from 0 to 1
			if y <= 0.01 {
				y = 0.01
			}
			if y >= 0.99 {
				y = 0.99
			}
			// logit(y) = log(y / (1-y))
			logit := math.Log(y / (1 - y))
			// Normalize to voltage range
			targetVoltage = (logit + 4.0) / 8.0 // Normalize from [-4,4] to [0,1]
		case ActivationTanh:
			y := 2*t - 1 // Output from -1 to 1
			if y <= -0.99 {
				y = -0.99
			}
			if y >= 0.99 {
				y = 0.99
			}
			// arctanh(y) = 0.5 * log((1+y)/(1-y))
			arctanh := 0.5 * math.Log((1+y)/(1-y))
			targetVoltage = (arctanh + 3.0) / 6.0
		default:
			targetVoltage = t
		}

		// Conductance determines ramp speed at this step
		// Higher conductance = faster ramp
		unit.MemristorConductances[i] = gMin + (gMax-gMin)*targetVoltage
		unit.RampVoltages[i] = t * 1.0 // 1V max ramp
	}
}

// Apply applies the analog activation
func (unit *AnalogActivationUnit) Apply(inputCurrent float64) float64 {
	// Simulate ramp ADC with nonlinear ramp
	// The crossover point determines the output

	// Normalize input to expected range
	inputNorm := (inputCurrent + 4.0) / 8.0 // Assume input in [-4, 4]
	if inputNorm < 0 {
		inputNorm = 0
	}
	if inputNorm > 1 {
		inputNorm = 1
	}

	// Find crossover point
	for i, rampV := range unit.RampVoltages {
		if rampV >= inputNorm {
			// Output is the normalized time index
			output := float64(i) / float64(len(unit.RampVoltages)-1)

			// Map back to activation range
			switch unit.Config.TargetFunction {
			case ActivationSigmoid:
				return output // Already in [0, 1]
			case ActivationTanh:
				return 2*output - 1 // Map to [-1, 1]
			default:
				return output
			}
		}
	}

	return 1.0 // Max output if no crossover found
}

// =============================================================================
// Integrated CIM Activation Layer
// =============================================================================

// CIMActivationLayerConfig configures a CIM activation layer
type CIMActivationLayerConfig struct {
	InputSize         int
	ActivationType    ActivationType
	Implementation    string // "digital", "analog", "hybrid"
	BitPrecision      int
}

// CIMActivationLayer represents a hardware activation layer
type CIMActivationLayer struct {
	Config            *CIMActivationLayerConfig
	DigitalUnit       *ActivationUnit
	AnalogUnit        *AnalogActivationUnit
	GELUUnit          *GELUISPAUnit

	// Performance
	ThroughputGOPS    float64
	EnergyPerAct      float64 // pJ per activation
	AreaMM2           float64
}

// NewCIMActivationLayer creates a new CIM activation layer
func NewCIMActivationLayer(config *CIMActivationLayerConfig) *CIMActivationLayer {
	if config == nil {
		config = &CIMActivationLayerConfig{
			InputSize:      1024,
			ActivationType: ActivationReLU,
			Implementation: "digital",
			BitPrecision:   8,
		}
	}

	layer := &CIMActivationLayer{
		Config: config,
	}

	switch config.Implementation {
	case "digital":
		layer.DigitalUnit = NewActivationUnit(&ActivationConfig{
			Type:              config.ActivationType,
			ApproximationType: "piecewise",
			NumPieces:         16,
			BitPrecision:      config.BitPrecision,
		})
		layer.EnergyPerAct = 0.1  // pJ
		layer.AreaMM2 = 0.001
	case "analog":
		layer.AnalogUnit = NewAnalogActivationUnit(&AnalogActivationConfig{
			TargetFunction: config.ActivationType,
		})
		layer.EnergyPerAct = 0.05 // Lower energy
		layer.AreaMM2 = 0.0015    // Slightly larger
	case "hybrid":
		// Use GELU ISPA for transformers
		if config.ActivationType == ActivationGELU {
			layer.GELUUnit = NewGELUISPAUnit(nil)
			layer.EnergyPerAct = 0.08
		} else {
			layer.DigitalUnit = NewActivationUnit(DefaultActivationConfig(config.ActivationType))
			layer.EnergyPerAct = 0.1
		}
		layer.AreaMM2 = 0.001
	}

	// Calculate throughput
	clockFreq := 1e9 // 1 GHz
	layer.ThroughputGOPS = clockFreq / 1e9 * float64(config.InputSize)

	return layer
}

// Forward applies the activation layer
func (layer *CIMActivationLayer) Forward(inputs []float64) []float64 {
	outputs := make([]float64, len(inputs))

	switch layer.Config.Implementation {
	case "digital":
		if layer.DigitalUnit != nil {
			outputs = layer.DigitalUnit.ApplyBatch(inputs)
		}
	case "analog":
		if layer.AnalogUnit != nil {
			for i, x := range inputs {
				outputs[i] = layer.AnalogUnit.Apply(x)
			}
		}
	case "hybrid":
		if layer.GELUUnit != nil {
			for i, x := range inputs {
				outputs[i] = layer.GELUUnit.Apply(x)
			}
		} else if layer.DigitalUnit != nil {
			outputs = layer.DigitalUnit.ApplyBatch(inputs)
		}
	}

	return outputs
}

// GetStats returns layer statistics
func (layer *CIMActivationLayer) GetStats() map[string]float64 {
	stats := map[string]float64{
		"throughput_gops": layer.ThroughputGOPS,
		"energy_per_act_pj": layer.EnergyPerAct,
		"area_mm2": layer.AreaMM2,
	}

	if layer.DigitalUnit != nil {
		stats["area_reduction"] = layer.DigitalUnit.AreaReduction
		stats["power_reduction"] = layer.DigitalUnit.PowerReduction
		stats["accuracy_loss"] = layer.DigitalUnit.AccuracyLoss
	}

	if layer.GELUUnit != nil {
		stats["gelu_mse"] = layer.GELUUnit.MSE
		stats["gelu_area_reduction"] = layer.GELUUnit.AreaReduction
		stats["gelu_power_reduction"] = layer.GELUUnit.PowerReduction
	}

	return stats
}
