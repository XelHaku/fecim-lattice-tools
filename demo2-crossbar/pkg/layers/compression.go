// Package layers provides neural network layer implementations for crossbar-based CIM.
// compression.go implements model compression techniques for efficient CIM deployment.
//
// Compression Techniques:
// - Knowledge distillation (teacher-student)
// - Weight pruning (magnitude, structured, iterative)
// - Quantization-aware compression
// - Low-rank factorization
// - Filter decomposition
//
// CIM-Specific Optimizations:
// - Crossbar-aware pruning patterns
// - Conductance-friendly quantization
// - Tile-aligned compression
//
// References:
// - Frontiers 2025: Survey of compression techniques
// - ScienceDirect 2024: ITMC iterative compression
// - Applied Intelligence 2024: Comprehensive review

package layers

import (
	"math"
	"sort"
)

// ============================================================================
// Compression Configuration
// ============================================================================

// CompressionConfig configures model compression
type CompressionConfig struct {
	// Pruning settings
	PruningRatio     float64
	PruningMethod    PruningMethod
	IterativePruning bool
	PruningSteps     int

	// Quantization settings
	QuantizeBits     int
	QuantizeWeights  bool
	QuantizeActivations bool

	// Knowledge distillation
	UseDistillation  bool
	Temperature      float64
	DistillAlpha     float64

	// Low-rank factorization
	UseLowRank       bool
	RankRatio        float64

	// CIM-specific
	CrossbarSize     int
	TileAligned      bool
}

// PruningMethod defines pruning approach
type PruningMethod int

const (
	// PruneUnstructured removes individual weights
	PruneUnstructured PruningMethod = iota
	// PruneStructured removes entire filters/channels
	PruneStructured
	// PruneNM implements N:M structured sparsity
	PruneNM
	// PruneTileAligned aligns pruning to crossbar tiles
	PruneTileAligned
)

// DefaultCompressionConfig returns default compression settings
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		PruningRatio:       0.5,
		PruningMethod:      PruneUnstructured,
		IterativePruning:   true,
		PruningSteps:       10,
		QuantizeBits:       8,
		QuantizeWeights:    true,
		QuantizeActivations: true,
		UseDistillation:    false,
		Temperature:        4.0,
		DistillAlpha:       0.5,
		UseLowRank:         false,
		RankRatio:          0.5,
		CrossbarSize:       64,
		TileAligned:        true,
	}
}

// ============================================================================
// Model Compressor
// ============================================================================

// ModelCompressor handles model compression
type ModelCompressor struct {
	Config *CompressionConfig
	Stats  *CompressionStats
}

// CompressionStats tracks compression statistics
type CompressionStats struct {
	OriginalParams   int64
	CompressedParams int64
	CompressionRatio float64
	SparsityAchieved float64
	OriginalSize     int64 // bytes
	CompressedSize   int64 // bytes
	AccuracyDrop     float64
}

// NewModelCompressor creates a new compressor
func NewModelCompressor(config *CompressionConfig) *ModelCompressor {
	if config == nil {
		config = DefaultCompressionConfig()
	}
	return &ModelCompressor{
		Config: config,
		Stats:  &CompressionStats{},
	}
}

// ============================================================================
// Weight Pruning
// ============================================================================

// PruneWeights applies pruning to weight matrix
func (c *ModelCompressor) PruneWeights(weights [][]float64) ([][]float64, [][]bool) {
	switch c.Config.PruningMethod {
	case PruneUnstructured:
		return c.pruneUnstructured(weights)
	case PruneStructured:
		return c.pruneStructured(weights)
	case PruneNM:
		return c.pruneNM(weights, 2, 4) // 2:4 default
	case PruneTileAligned:
		return c.pruneTileAligned(weights)
	default:
		return c.pruneUnstructured(weights)
	}
}

// pruneUnstructured removes smallest magnitude weights
func (c *ModelCompressor) pruneUnstructured(weights [][]float64) ([][]float64, [][]bool) {
	rows := len(weights)
	if rows == 0 {
		return weights, nil
	}
	cols := len(weights[0])

	// Collect all weights
	type weightInfo struct {
		row, col int
		absVal   float64
	}
	allWeights := make([]weightInfo, 0, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			allWeights = append(allWeights, weightInfo{
				row:    i,
				col:    j,
				absVal: math.Abs(weights[i][j]),
			})
		}
	}

	// Sort by magnitude
	sort.Slice(allWeights, func(i, j int) bool {
		return allWeights[i].absVal < allWeights[j].absVal
	})

	// Calculate threshold
	pruneCount := int(float64(len(allWeights)) * c.Config.PruningRatio)

	// Create mask
	mask := make([][]bool, rows)
	pruned := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		mask[i] = make([]bool, cols)
		pruned[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			mask[i][j] = true
			pruned[i][j] = weights[i][j]
		}
	}

	// Apply pruning
	for k := 0; k < pruneCount; k++ {
		w := allWeights[k]
		mask[w.row][w.col] = false
		pruned[w.row][w.col] = 0
	}

	// Update stats
	c.Stats.SparsityAchieved = float64(pruneCount) / float64(rows*cols)

	return pruned, mask
}

// pruneStructured removes entire rows/columns
func (c *ModelCompressor) pruneStructured(weights [][]float64) ([][]float64, [][]bool) {
	rows := len(weights)
	if rows == 0 {
		return weights, nil
	}
	cols := len(weights[0])

	// Calculate row importance (L2 norm)
	rowImportance := make([]float64, rows)
	for i := 0; i < rows; i++ {
		sumSq := 0.0
		for j := 0; j < cols; j++ {
			sumSq += weights[i][j] * weights[i][j]
		}
		rowImportance[i] = math.Sqrt(sumSq)
	}

	// Sort rows by importance
	type rowInfo struct {
		idx        int
		importance float64
	}
	sortedRows := make([]rowInfo, rows)
	for i := 0; i < rows; i++ {
		sortedRows[i] = rowInfo{idx: i, importance: rowImportance[i]}
	}
	sort.Slice(sortedRows, func(i, j int) bool {
		return sortedRows[i].importance < sortedRows[j].importance
	})

	// Prune least important rows
	pruneRows := int(float64(rows) * c.Config.PruningRatio)

	mask := make([][]bool, rows)
	pruned := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		mask[i] = make([]bool, cols)
		pruned[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			mask[i][j] = true
			pruned[i][j] = weights[i][j]
		}
	}

	// Zero out pruned rows
	for k := 0; k < pruneRows; k++ {
		rowIdx := sortedRows[k].idx
		for j := 0; j < cols; j++ {
			mask[rowIdx][j] = false
			pruned[rowIdx][j] = 0
		}
	}

	c.Stats.SparsityAchieved = float64(pruneRows) / float64(rows)

	return pruned, mask
}

// pruneNM implements N:M structured sparsity
func (c *ModelCompressor) pruneNM(weights [][]float64, n, m int) ([][]float64, [][]bool) {
	rows := len(weights)
	if rows == 0 {
		return weights, nil
	}
	cols := len(weights[0])

	mask := make([][]bool, rows)
	pruned := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		mask[i] = make([]bool, cols)
		pruned[i] = make([]float64, cols)
	}

	// Process each row
	for i := 0; i < rows; i++ {
		// Process in blocks of M
		for blockStart := 0; blockStart < cols; blockStart += m {
			blockEnd := min(blockStart+m, cols)
			blockLen := blockEnd - blockStart

			// Collect block weights with indices
			type blockWeight struct {
				col    int
				absVal float64
			}
			block := make([]blockWeight, blockLen)
			for j := blockStart; j < blockEnd; j++ {
				block[j-blockStart] = blockWeight{
					col:    j,
					absVal: math.Abs(weights[i][j]),
				}
			}

			// Sort by magnitude (descending)
			sort.Slice(block, func(a, b int) bool {
				return block[a].absVal > block[b].absVal
			})

			// Keep top N
			keepCount := min(n, blockLen)
			for k := 0; k < keepCount; k++ {
				j := block[k].col
				mask[i][j] = true
				pruned[i][j] = weights[i][j]
			}
		}
	}

	c.Stats.SparsityAchieved = 1.0 - float64(n)/float64(m)

	return pruned, mask
}

// pruneTileAligned aligns pruning to crossbar tile boundaries
func (c *ModelCompressor) pruneTileAligned(weights [][]float64) ([][]float64, [][]bool) {
	rows := len(weights)
	if rows == 0 {
		return weights, nil
	}
	cols := len(weights[0])
	tileSize := c.Config.CrossbarSize

	// Calculate tile importance
	tileRows := (rows + tileSize - 1) / tileSize
	tileCols := (cols + tileSize - 1) / tileSize

	type tileInfo struct {
		tileRow, tileCol int
		importance       float64
	}
	tiles := make([]tileInfo, 0, tileRows*tileCols)

	for ti := 0; ti < tileRows; ti++ {
		for tj := 0; tj < tileCols; tj++ {
			rowStart := ti * tileSize
			rowEnd := min((ti+1)*tileSize, rows)
			colStart := tj * tileSize
			colEnd := min((tj+1)*tileSize, cols)

			// Calculate tile importance (Frobenius norm)
			sumSq := 0.0
			for i := rowStart; i < rowEnd; i++ {
				for j := colStart; j < colEnd; j++ {
					sumSq += weights[i][j] * weights[i][j]
				}
			}

			tiles = append(tiles, tileInfo{
				tileRow:    ti,
				tileCol:    tj,
				importance: math.Sqrt(sumSq),
			})
		}
	}

	// Sort tiles by importance
	sort.Slice(tiles, func(i, j int) bool {
		return tiles[i].importance < tiles[j].importance
	})

	// Prune least important tiles
	pruneTiles := int(float64(len(tiles)) * c.Config.PruningRatio)

	mask := make([][]bool, rows)
	pruned := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		mask[i] = make([]bool, cols)
		pruned[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			mask[i][j] = true
			pruned[i][j] = weights[i][j]
		}
	}

	// Zero out pruned tiles
	for k := 0; k < pruneTiles; k++ {
		tile := tiles[k]
		rowStart := tile.tileRow * tileSize
		rowEnd := min((tile.tileRow+1)*tileSize, rows)
		colStart := tile.tileCol * tileSize
		colEnd := min((tile.tileCol+1)*tileSize, cols)

		for i := rowStart; i < rowEnd; i++ {
			for j := colStart; j < colEnd; j++ {
				mask[i][j] = false
				pruned[i][j] = 0
			}
		}
	}

	c.Stats.SparsityAchieved = float64(pruneTiles) / float64(len(tiles))

	return pruned, mask
}

// ============================================================================
// Iterative Pruning
// ============================================================================

// IterativePruner implements gradual pruning
type IterativePruner struct {
	Config       *CompressionConfig
	CurrentStep  int
	CurrentRatio float64
	FinalRatio   float64
}

// NewIterativePruner creates iterative pruner
func NewIterativePruner(finalRatio float64, steps int) *IterativePruner {
	return &IterativePruner{
		Config: &CompressionConfig{
			PruningRatio:     0.0,
			PruningMethod:    PruneUnstructured,
			IterativePruning: true,
			PruningSteps:     steps,
		},
		CurrentStep:  0,
		CurrentRatio: 0.0,
		FinalRatio:   finalRatio,
	}
}

// Step performs one pruning step
func (p *IterativePruner) Step(weights [][]float64) ([][]float64, [][]bool) {
	p.CurrentStep++

	// Cubic schedule (gradual increase)
	progress := float64(p.CurrentStep) / float64(p.Config.PruningSteps)
	p.CurrentRatio = p.FinalRatio * math.Pow(progress, 3)

	p.Config.PruningRatio = p.CurrentRatio

	compressor := NewModelCompressor(p.Config)
	return compressor.PruneWeights(weights)
}

// IsDone returns true if pruning is complete
func (p *IterativePruner) IsDone() bool {
	return p.CurrentStep >= p.Config.PruningSteps
}

// ============================================================================
// Knowledge Distillation
// ============================================================================

// DistillationConfig configures knowledge distillation
type DistillationConfig struct {
	Temperature float64 // Softmax temperature
	Alpha       float64 // Weight for distillation loss
	Beta        float64 // Weight for hard label loss
}

// DefaultDistillationConfig returns default distillation config
func DefaultDistillationConfig() *DistillationConfig {
	return &DistillationConfig{
		Temperature: 4.0,
		Alpha:       0.7,
		Beta:        0.3,
	}
}

// KnowledgeDistiller handles knowledge distillation
type KnowledgeDistiller struct {
	Config *DistillationConfig
}

// NewKnowledgeDistiller creates a distiller
func NewKnowledgeDistiller(config *DistillationConfig) *KnowledgeDistiller {
	if config == nil {
		config = DefaultDistillationConfig()
	}
	return &KnowledgeDistiller{Config: config}
}

// SoftmaxWithTemperature applies softmax with temperature
func (d *KnowledgeDistiller) SoftmaxWithTemperature(logits []float64) []float64 {
	T := d.Config.Temperature
	scaled := make([]float64, len(logits))

	// Find max for numerical stability
	maxVal := logits[0]
	for _, v := range logits {
		if v > maxVal {
			maxVal = v
		}
	}

	// Compute softmax
	sum := 0.0
	for i, v := range logits {
		scaled[i] = math.Exp((v - maxVal) / T)
		sum += scaled[i]
	}

	for i := range scaled {
		scaled[i] /= sum
	}

	return scaled
}

// ComputeDistillationLoss computes KL divergence loss
func (d *KnowledgeDistiller) ComputeDistillationLoss(studentLogits, teacherLogits []float64) float64 {
	studentSoft := d.SoftmaxWithTemperature(studentLogits)
	teacherSoft := d.SoftmaxWithTemperature(teacherLogits)

	// KL divergence: sum(teacher * log(teacher/student))
	klDiv := 0.0
	for i := range teacherSoft {
		if teacherSoft[i] > 1e-10 && studentSoft[i] > 1e-10 {
			klDiv += teacherSoft[i] * math.Log(teacherSoft[i]/studentSoft[i])
		}
	}

	// Scale by T^2 as per Hinton et al.
	return klDiv * d.Config.Temperature * d.Config.Temperature
}

// ComputeTotalLoss combines distillation and hard label losses
func (d *KnowledgeDistiller) ComputeTotalLoss(studentLogits, teacherLogits []float64, hardLabel int) float64 {
	// Distillation loss
	distillLoss := d.ComputeDistillationLoss(studentLogits, teacherLogits)

	// Hard label cross-entropy loss
	studentProbs := d.SoftmaxWithTemperature(studentLogits)
	hardLoss := -math.Log(studentProbs[hardLabel] + 1e-10)

	return d.Config.Alpha*distillLoss + d.Config.Beta*hardLoss
}

// ============================================================================
// Low-Rank Factorization
// ============================================================================

// LowRankFactorizer decomposes weight matrices
type LowRankFactorizer struct {
	RankRatio float64 // Fraction of original rank to keep
}

// NewLowRankFactorizer creates a factorizer
func NewLowRankFactorizer(rankRatio float64) *LowRankFactorizer {
	return &LowRankFactorizer{RankRatio: rankRatio}
}

// Factorize decomposes W into U * V^T using truncated SVD approximation
func (f *LowRankFactorizer) Factorize(weights [][]float64) ([][]float64, [][]float64) {
	rows := len(weights)
	if rows == 0 {
		return nil, nil
	}
	cols := len(weights[0])

	// Target rank
	maxRank := min(rows, cols)
	targetRank := int(float64(maxRank) * f.RankRatio)
	if targetRank < 1 {
		targetRank = 1
	}

	// Simple power iteration for approximate SVD
	// For production, would use proper SVD library
	U := make([][]float64, rows)
	V := make([][]float64, cols)

	for i := range U {
		U[i] = make([]float64, targetRank)
	}
	for j := range V {
		V[j] = make([]float64, targetRank)
	}

	// Initialize with random values
	for k := 0; k < targetRank; k++ {
		// Random initialization for U column k
		for i := 0; i < rows; i++ {
			U[i][k] = weights[i][k%cols]
		}

		// Power iteration (simplified)
		for iter := 0; iter < 10; iter++ {
			// V = W^T * U
			for j := 0; j < cols; j++ {
				sum := 0.0
				for i := 0; i < rows; i++ {
					sum += weights[i][j] * U[i][k]
				}
				V[j][k] = sum
			}

			// Normalize V
			norm := 0.0
			for j := 0; j < cols; j++ {
				norm += V[j][k] * V[j][k]
			}
			norm = math.Sqrt(norm)
			if norm > 1e-10 {
				for j := 0; j < cols; j++ {
					V[j][k] /= norm
				}
			}

			// U = W * V
			for i := 0; i < rows; i++ {
				sum := 0.0
				for j := 0; j < cols; j++ {
					sum += weights[i][j] * V[j][k]
				}
				U[i][k] = sum
			}

			// Normalize U
			norm = 0.0
			for i := 0; i < rows; i++ {
				norm += U[i][k] * U[i][k]
			}
			norm = math.Sqrt(norm)
			if norm > 1e-10 {
				for i := 0; i < rows; i++ {
					U[i][k] /= norm
				}
			}
		}
	}

	return U, V
}

// ComputeCompressionRatio calculates size reduction from factorization
func (f *LowRankFactorizer) ComputeCompressionRatio(rows, cols, rank int) float64 {
	originalParams := rows * cols
	factorizedParams := rows*rank + cols*rank
	return float64(originalParams) / float64(factorizedParams)
}

// ============================================================================
// CIM-Aware Compression
// ============================================================================

// CIMCompressionConfig configures CIM-specific compression
type CIMCompressionConfig struct {
	CrossbarSize     int
	ConductanceRange [2]float64 // [Gmin, Gmax]
	ADCBits          int
	TargetSparsity   float64
	PreserveTiles    bool // Don't prune entire tiles
}

// DefaultCIMCompressionConfig returns CIM-optimized settings
func DefaultCIMCompressionConfig() *CIMCompressionConfig {
	return &CIMCompressionConfig{
		CrossbarSize:     64,
		ConductanceRange: [2]float64{0.1, 1.0},
		ADCBits:          6,
		TargetSparsity:   0.5,
		PreserveTiles:    true,
	}
}

// CIMCompressor handles CIM-aware compression
type CIMCompressor struct {
	Config *CIMCompressionConfig
}

// NewCIMCompressor creates CIM-aware compressor
func NewCIMCompressor(config *CIMCompressionConfig) *CIMCompressor {
	if config == nil {
		config = DefaultCIMCompressionConfig()
	}
	return &CIMCompressor{Config: config}
}

// CompressForCIM applies combined compression for CIM deployment
func (c *CIMCompressor) CompressForCIM(weights [][]float64) *CIMCompressedModel {
	rows := len(weights)
	if rows == 0 {
		return nil
	}
	cols := len(weights[0])

	result := &CIMCompressedModel{
		OriginalRows: rows,
		OriginalCols: cols,
	}

	// Step 1: Magnitude pruning with tile awareness
	compConfig := &CompressionConfig{
		PruningRatio:  c.Config.TargetSparsity,
		PruningMethod: PruneTileAligned,
		CrossbarSize:  c.Config.CrossbarSize,
		TileAligned:   true,
	}
	compressor := NewModelCompressor(compConfig)
	prunedWeights, mask := compressor.PruneWeights(weights)

	// Step 2: Quantize to conductance levels
	numLevels := 1 << c.Config.ADCBits
	gMin := c.Config.ConductanceRange[0]
	gMax := c.Config.ConductanceRange[1]

	quantizedWeights := make([][]float64, rows)
	conductances := make([][]float64, rows)

	// Find weight range
	wMin, wMax := math.MaxFloat64, -math.MaxFloat64
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if mask[i][j] {
				if prunedWeights[i][j] < wMin {
					wMin = prunedWeights[i][j]
				}
				if prunedWeights[i][j] > wMax {
					wMax = prunedWeights[i][j]
				}
			}
		}
	}

	wRange := wMax - wMin
	if wRange < 1e-10 {
		wRange = 1.0
	}

	for i := 0; i < rows; i++ {
		quantizedWeights[i] = make([]float64, cols)
		conductances[i] = make([]float64, cols)

		for j := 0; j < cols; j++ {
			if mask[i][j] {
				// Normalize to [0, 1]
				normalized := (prunedWeights[i][j] - wMin) / wRange

				// Quantize to levels
				level := int(normalized * float64(numLevels-1))
				if level < 0 {
					level = 0
				}
				if level >= numLevels {
					level = numLevels - 1
				}

				// Dequantize
				quantizedWeights[i][j] = wMin + float64(level)/float64(numLevels-1)*wRange

				// Map to conductance
				conductances[i][j] = gMin + float64(level)/float64(numLevels-1)*(gMax-gMin)
			}
		}
	}

	result.Weights = quantizedWeights
	result.Mask = mask
	result.Conductances = conductances
	result.Sparsity = compressor.Stats.SparsityAchieved

	// Calculate compression stats
	nonzero := 0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if mask[i][j] {
				nonzero++
			}
		}
	}
	result.CompressionRatio = float64(rows*cols) / float64(nonzero)

	return result
}

// CIMCompressedModel represents compressed model for CIM
type CIMCompressedModel struct {
	OriginalRows     int
	OriginalCols     int
	Weights          [][]float64
	Mask             [][]bool
	Conductances     [][]float64
	Sparsity         float64
	CompressionRatio float64
}

// ============================================================================
// Compression Pipeline
// ============================================================================

// CompressionPipeline chains multiple compression techniques
type CompressionPipeline struct {
	Stages []CompressionStage
	Config *CompressionConfig
}

// CompressionStage represents one compression step
type CompressionStage struct {
	Name   string
	Method string
	Config interface{}
}

// NewCompressionPipeline creates a compression pipeline
func NewCompressionPipeline(config *CompressionConfig) *CompressionPipeline {
	if config == nil {
		config = DefaultCompressionConfig()
	}
	return &CompressionPipeline{
		Config: config,
		Stages: make([]CompressionStage, 0),
	}
}

// AddPruning adds pruning stage
func (p *CompressionPipeline) AddPruning(ratio float64, method PruningMethod) *CompressionPipeline {
	p.Stages = append(p.Stages, CompressionStage{
		Name:   "pruning",
		Method: "magnitude",
		Config: map[string]interface{}{"ratio": ratio, "method": method},
	})
	return p
}

// AddQuantization adds quantization stage
func (p *CompressionPipeline) AddQuantization(bits int) *CompressionPipeline {
	p.Stages = append(p.Stages, CompressionStage{
		Name:   "quantization",
		Method: "uniform",
		Config: map[string]interface{}{"bits": bits},
	})
	return p
}

// AddDistillation adds knowledge distillation stage
func (p *CompressionPipeline) AddDistillation(temperature, alpha float64) *CompressionPipeline {
	p.Stages = append(p.Stages, CompressionStage{
		Name:   "distillation",
		Method: "kl_divergence",
		Config: map[string]interface{}{"temperature": temperature, "alpha": alpha},
	})
	return p
}

// Execute runs the compression pipeline
func (p *CompressionPipeline) Execute(weights [][]float64) ([][]float64, *CompressionStats) {
	result := weights
	stats := &CompressionStats{
		OriginalParams: int64(len(weights) * len(weights[0])),
	}

	for _, stage := range p.Stages {
		switch stage.Name {
		case "pruning":
			cfg := stage.Config.(map[string]interface{})
			ratio := cfg["ratio"].(float64)
			method := cfg["method"].(PruningMethod)

			compressor := NewModelCompressor(&CompressionConfig{
				PruningRatio:  ratio,
				PruningMethod: method,
			})
			result, _ = compressor.PruneWeights(result)
			stats.SparsityAchieved = compressor.Stats.SparsityAchieved

		case "quantization":
			cfg := stage.Config.(map[string]interface{})
			bits := cfg["bits"].(int)
			result = quantizeWeights(result, bits)
		}
	}

	// Calculate final stats
	nonzero := int64(0)
	for i := range result {
		for j := range result[i] {
			if result[i][j] != 0 {
				nonzero++
			}
		}
	}
	stats.CompressedParams = nonzero
	stats.CompressionRatio = float64(stats.OriginalParams) / float64(nonzero)

	return result, stats
}

// quantizeWeights applies uniform quantization
func quantizeWeights(weights [][]float64, bits int) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])

	levels := float64(1 << bits)

	// Find range
	minVal, maxVal := math.MaxFloat64, -math.MaxFloat64
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if weights[i][j] < minVal {
				minVal = weights[i][j]
			}
			if weights[i][j] > maxVal {
				maxVal = weights[i][j]
			}
		}
	}

	valRange := maxVal - minVal
	if valRange < 1e-10 {
		valRange = 1.0
	}

	result := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			normalized := (weights[i][j] - minVal) / valRange
			level := math.Round(normalized * (levels - 1))
			result[i][j] = minVal + level/(levels-1)*valRange
		}
	}

	return result
}
