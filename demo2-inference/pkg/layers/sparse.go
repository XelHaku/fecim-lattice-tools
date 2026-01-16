// Package layers provides neural network layer implementations for crossbar-based CIM.
// sparse.go implements sparse weight utilities for efficient crossbar mapping.
//
// Sparse techniques for CIM:
// - Unstructured pruning: Individual weight elimination
// - Structured pruning: Block/channel-level sparsity
// - Sorted Weight Sectioning (SWS): ADC energy reduction
// - Index-free sparse mapping: Dual-FeFET approach
// - QUBO matrix compression: For optimization problems
//
// References:
// - Nature Electronics 2025: Index-free sparse FeFET crossbar
// - arXiv 2024: Sorted Weight Sectioning for CIM
// - Nature Comms 2024: FeFET CiM annealer with sparse QUBO

package layers

import (
	"math"
	"sort"
)

// SparsityFormat defines how sparsity is represented
type SparsityFormat int

const (
	// SparsityDense represents no sparsity (full matrix)
	SparsityDense SparsityFormat = iota
	// SparsityCOO represents Coordinate format (row, col, val)
	SparsityCOO
	// SparsityCSR represents Compressed Sparse Row format
	SparsityCSR
	// SparsityCSC represents Compressed Sparse Column format
	SparsityCSC
	// SparsityBitmap represents binary mask + values
	SparsityBitmap
	// SparsityBlock represents block-sparse format
	SparsityBlock
)

// SparseWeights represents sparse weight matrix
type SparseWeights struct {
	Format     SparsityFormat
	Rows       int
	Cols       int
	NumNonzero int
	Density    float64

	// COO format
	RowIndices []int
	ColIndices []int
	Values     []float64

	// CSR format
	RowPtr []int
	ColIdx []int
	Data   []float64

	// Bitmap format
	Bitmap [][]bool
	Dense  [][]float64

	// Block format
	BlockSize int
	Blocks    []*SparseBlock
}

// SparseBlock represents a non-zero block in block-sparse format
type SparseBlock struct {
	RowStart int
	ColStart int
	Data     [][]float64
}

// SparsifyConfig configures sparsification
type SparsifyConfig struct {
	TargetSparsity float64 // Target sparsity ratio (0.0-1.0)
	Method         SparsifyMethod
	BlockSize      int     // For block sparsity
	Threshold      float64 // For magnitude pruning
	PreserveTop    int     // Preserve top-k weights per row/col
}

// SparsifyMethod defines pruning method
type SparsifyMethod int

const (
	// SparsifyMagnitude prunes by absolute magnitude
	SparsifyMagnitude SparsifyMethod = iota
	// SparsifyRandom prunes randomly
	SparsifyRandom
	// SparsifyStructured prunes entire rows/columns
	SparsifyStructured
	// SparsifyBlock prunes blocks
	SparsifyBlock
	// SparsifySWS uses Sorted Weight Sectioning
	SparsifySWS
)

// DefaultSparsifyConfig returns default sparsification config
func DefaultSparsifyConfig() *SparsifyConfig {
	return &SparsifyConfig{
		TargetSparsity: 0.5,
		Method:         SparsifyMagnitude,
		BlockSize:      4,
		Threshold:      0.0,
		PreserveTop:    0,
	}
}

// Sparsifier handles weight sparsification
type Sparsifier struct {
	Config *SparsifyConfig
}

// NewSparsifier creates a new sparsifier
func NewSparsifier(config *SparsifyConfig) *Sparsifier {
	if config == nil {
		config = DefaultSparsifyConfig()
	}
	return &Sparsifier{Config: config}
}

// Sparsify converts dense weights to sparse format
func (s *Sparsifier) Sparsify(weights [][]float64) *SparseWeights {
	rows := len(weights)
	if rows == 0 {
		return nil
	}
	cols := len(weights[0])

	switch s.Config.Method {
	case SparsifyMagnitude:
		return s.sparsifyMagnitude(weights)
	case SparsifyBlock:
		return s.sparsifyBlock(weights)
	case SparsifySWS:
		return s.sparsifySWS(weights)
	default:
		return s.sparsifyMagnitude(weights)
	}

	sparse := &SparseWeights{
		Rows: rows,
		Cols: cols,
	}
	return sparse
}

// sparsifyMagnitude prunes weights by absolute magnitude
func (s *Sparsifier) sparsifyMagnitude(weights [][]float64) *SparseWeights {
	rows := len(weights)
	cols := len(weights[0])

	// Collect all weights with positions
	type weightPos struct {
		row, col int
		value    float64
		absVal   float64
	}
	allWeights := make([]weightPos, 0, rows*cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			allWeights = append(allWeights, weightPos{
				row:    i,
				col:    j,
				value:  weights[i][j],
				absVal: math.Abs(weights[i][j]),
			})
		}
	}

	// Sort by absolute magnitude (descending)
	sort.Slice(allWeights, func(i, j int) bool {
		return allWeights[i].absVal > allWeights[j].absVal
	})

	// Keep top (1-sparsity) weights
	keepCount := int(float64(rows*cols) * (1.0 - s.Config.TargetSparsity))
	if keepCount < 1 {
		keepCount = 1
	}

	// Build sparse representation
	sparse := &SparseWeights{
		Format:     SparsityBitmap,
		Rows:       rows,
		Cols:       cols,
		NumNonzero: keepCount,
		Density:    float64(keepCount) / float64(rows*cols),
		Bitmap:     make([][]bool, rows),
		Dense:      make([][]float64, rows),
	}

	for i := 0; i < rows; i++ {
		sparse.Bitmap[i] = make([]bool, cols)
		sparse.Dense[i] = make([]float64, cols)
	}

	// Mark kept weights
	for k := 0; k < keepCount && k < len(allWeights); k++ {
		w := allWeights[k]
		sparse.Bitmap[w.row][w.col] = true
		sparse.Dense[w.row][w.col] = w.value
	}

	// Also build COO format
	sparse.RowIndices = make([]int, 0, keepCount)
	sparse.ColIndices = make([]int, 0, keepCount)
	sparse.Values = make([]float64, 0, keepCount)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if sparse.Bitmap[i][j] {
				sparse.RowIndices = append(sparse.RowIndices, i)
				sparse.ColIndices = append(sparse.ColIndices, j)
				sparse.Values = append(sparse.Values, sparse.Dense[i][j])
			}
		}
	}

	return sparse
}

// sparsifyBlock performs block-sparse pruning
func (s *Sparsifier) sparsifyBlock(weights [][]float64) *SparseWeights {
	rows := len(weights)
	cols := len(weights[0])
	blockSize := s.Config.BlockSize
	if blockSize < 1 {
		blockSize = 4
	}

	// Calculate block grid dimensions
	blockRows := (rows + blockSize - 1) / blockSize
	blockCols := (cols + blockSize - 1) / blockSize

	// Calculate block magnitudes (L2 norm)
	type blockInfo struct {
		rowIdx, colIdx int
		magnitude      float64
		data           [][]float64
	}
	blocks := make([]blockInfo, 0, blockRows*blockCols)

	for bi := 0; bi < blockRows; bi++ {
		for bj := 0; bj < blockCols; bj++ {
			startRow := bi * blockSize
			startCol := bj * blockSize
			endRow := min(startRow+blockSize, rows)
			endCol := min(startCol+blockSize, cols)

			// Extract block and compute magnitude
			blockData := make([][]float64, endRow-startRow)
			sumSq := 0.0
			for i := startRow; i < endRow; i++ {
				blockData[i-startRow] = make([]float64, endCol-startCol)
				for j := startCol; j < endCol; j++ {
					blockData[i-startRow][j-startCol] = weights[i][j]
					sumSq += weights[i][j] * weights[i][j]
				}
			}

			blocks = append(blocks, blockInfo{
				rowIdx:    bi,
				colIdx:    bj,
				magnitude: math.Sqrt(sumSq),
				data:      blockData,
			})
		}
	}

	// Sort blocks by magnitude
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].magnitude > blocks[j].magnitude
	})

	// Keep top blocks
	keepBlocks := int(float64(len(blocks)) * (1.0 - s.Config.TargetSparsity))
	if keepBlocks < 1 {
		keepBlocks = 1
	}

	// Build sparse representation
	sparse := &SparseWeights{
		Format:    SparsityBlock,
		Rows:      rows,
		Cols:      cols,
		BlockSize: blockSize,
		Blocks:    make([]*SparseBlock, keepBlocks),
		Bitmap:    make([][]bool, rows),
		Dense:     make([][]float64, rows),
	}

	for i := 0; i < rows; i++ {
		sparse.Bitmap[i] = make([]bool, cols)
		sparse.Dense[i] = make([]float64, cols)
	}

	nonzeroCount := 0
	for k := 0; k < keepBlocks; k++ {
		b := blocks[k]
		sparse.Blocks[k] = &SparseBlock{
			RowStart: b.rowIdx * blockSize,
			ColStart: b.colIdx * blockSize,
			Data:     b.data,
		}

		// Copy to dense/bitmap
		for i := range b.data {
			for j := range b.data[i] {
				ri := b.rowIdx*blockSize + i
				ci := b.colIdx*blockSize + j
				if ri < rows && ci < cols {
					sparse.Dense[ri][ci] = b.data[i][j]
					if b.data[i][j] != 0 {
						sparse.Bitmap[ri][ci] = true
						nonzeroCount++
					}
				}
			}
		}
	}

	sparse.NumNonzero = nonzeroCount
	sparse.Density = float64(nonzeroCount) / float64(rows*cols)

	return sparse
}

// sparsifySWS implements Sorted Weight Sectioning for CIM
// This places sorted weight sections on bit-sliced crossbars to reduce ADC energy
func (s *Sparsifier) sparsifySWS(weights [][]float64) *SparseWeights {
	rows := len(weights)
	cols := len(weights[0])

	// SWS sorts weights within each column to leverage small magnitude weights
	sparse := &SparseWeights{
		Format: SparsityBitmap,
		Rows:   rows,
		Cols:   cols,
		Bitmap: make([][]bool, rows),
		Dense:  make([][]float64, rows),
	}

	for i := 0; i < rows; i++ {
		sparse.Bitmap[i] = make([]bool, cols)
		sparse.Dense[i] = make([]float64, cols)
	}

	// For each column, sort by magnitude and prune smallest
	keepPerCol := int(float64(rows) * (1.0 - s.Config.TargetSparsity))
	if keepPerCol < 1 {
		keepPerCol = 1
	}

	type colWeight struct {
		row    int
		value  float64
		absVal float64
	}

	nonzeroCount := 0
	for j := 0; j < cols; j++ {
		// Collect column weights
		colWeights := make([]colWeight, rows)
		for i := 0; i < rows; i++ {
			colWeights[i] = colWeight{
				row:    i,
				value:  weights[i][j],
				absVal: math.Abs(weights[i][j]),
			}
		}

		// Sort by magnitude (descending)
		sort.Slice(colWeights, func(a, b int) bool {
			return colWeights[a].absVal > colWeights[b].absVal
		})

		// Keep top weights
		for k := 0; k < keepPerCol && k < rows; k++ {
			cw := colWeights[k]
			sparse.Bitmap[cw.row][j] = true
			sparse.Dense[cw.row][j] = cw.value
			nonzeroCount++
		}
	}

	sparse.NumNonzero = nonzeroCount
	sparse.Density = float64(nonzeroCount) / float64(rows*cols)

	return sparse
}

// ============================================================================
// Sparse Matrix Operations
// ============================================================================

// SparseMVM performs sparse matrix-vector multiplication
func SparseMVM(sparse *SparseWeights, input []float64) []float64 {
	if sparse == nil {
		return nil
	}

	output := make([]float64, sparse.Rows)

	switch sparse.Format {
	case SparsityBitmap:
		for i := 0; i < sparse.Rows; i++ {
			sum := 0.0
			for j := 0; j < sparse.Cols && j < len(input); j++ {
				if sparse.Bitmap[i][j] {
					sum += sparse.Dense[i][j] * input[j]
				}
			}
			output[i] = sum
		}

	case SparsityCOO:
		for k := range sparse.Values {
			i := sparse.RowIndices[k]
			j := sparse.ColIndices[k]
			if j < len(input) {
				output[i] += sparse.Values[k] * input[j]
			}
		}

	case SparsityCSR:
		for i := 0; i < sparse.Rows; i++ {
			sum := 0.0
			for k := sparse.RowPtr[i]; k < sparse.RowPtr[i+1]; k++ {
				j := sparse.ColIdx[k]
				if j < len(input) {
					sum += sparse.Data[k] * input[j]
				}
			}
			output[i] = sum
		}

	case SparsityBlock:
		for _, block := range sparse.Blocks {
			for i := range block.Data {
				for j := range block.Data[i] {
					ri := block.RowStart + i
					ci := block.ColStart + j
					if ri < sparse.Rows && ci < len(input) {
						output[ri] += block.Data[i][j] * input[ci]
					}
				}
			}
		}

	default:
		// Fall back to dense
		for i := 0; i < sparse.Rows; i++ {
			sum := 0.0
			for j := 0; j < sparse.Cols && j < len(input); j++ {
				sum += sparse.Dense[i][j] * input[j]
			}
			output[i] = sum
		}
	}

	return output
}

// ============================================================================
// Crossbar Sparse Mapping
// ============================================================================

// CrossbarSparseConfig configures sparse crossbar mapping
type CrossbarSparseConfig struct {
	ArraySize      int
	UseCompression bool
	CompressionRatio float64
}

// SparseCrossbarMapper maps sparse weights to crossbar arrays
type SparseCrossbarMapper struct {
	Config   *CrossbarSparseConfig
	Mappings []*CrossbarMapping
}

// CrossbarMapping represents weight mapping to a crossbar tile
type CrossbarMapping struct {
	TileRow      int
	TileCol      int
	SparseWeights *SparseWeights
	Conductances [][]float64
	ActiveRows   []int // Non-zero rows (for selective activation)
	ActiveCols   []int // Non-zero columns
	Utilization  float64
}

// NewSparseCrossbarMapper creates a new sparse crossbar mapper
func NewSparseCrossbarMapper(arraySize int) *SparseCrossbarMapper {
	return &SparseCrossbarMapper{
		Config: &CrossbarSparseConfig{
			ArraySize:        arraySize,
			UseCompression:   true,
			CompressionRatio: 0.75, // 75% area savings for 85% sparse QUBO
		},
		Mappings: make([]*CrossbarMapping, 0),
	}
}

// MapToTiles maps sparse weights to crossbar tiles
func (m *SparseCrossbarMapper) MapToTiles(sparse *SparseWeights) []*CrossbarMapping {
	arraySize := m.Config.ArraySize

	// Calculate number of tiles needed
	rowTiles := (sparse.Rows + arraySize - 1) / arraySize
	colTiles := (sparse.Cols + arraySize - 1) / arraySize

	mappings := make([]*CrossbarMapping, 0, rowTiles*colTiles)

	for ti := 0; ti < rowTiles; ti++ {
		for tj := 0; tj < colTiles; tj++ {
			startRow := ti * arraySize
			startCol := tj * arraySize
			endRow := min(startRow+arraySize, sparse.Rows)
			endCol := min(startCol+arraySize, sparse.Cols)

			// Extract tile weights
			tileWeights := make([][]float64, endRow-startRow)
			tileBitmap := make([][]bool, endRow-startRow)
			activeRows := make([]int, 0)
			activeCols := make([]int, 0)
			nonzero := 0

			for i := startRow; i < endRow; i++ {
				tileWeights[i-startRow] = make([]float64, endCol-startCol)
				tileBitmap[i-startRow] = make([]bool, endCol-startCol)
				rowHasNonzero := false
				for j := startCol; j < endCol; j++ {
					if sparse.Bitmap[i][j] {
						tileWeights[i-startRow][j-startCol] = sparse.Dense[i][j]
						tileBitmap[i-startRow][j-startCol] = true
						nonzero++
						rowHasNonzero = true
					}
				}
				if rowHasNonzero {
					activeRows = append(activeRows, i-startRow)
				}
			}

			// Find active columns
			colActive := make([]bool, endCol-startCol)
			for j := startCol; j < endCol; j++ {
				for i := startRow; i < endRow; i++ {
					if sparse.Bitmap[i][j] {
						colActive[j-startCol] = true
						break
					}
				}
				if colActive[j-startCol] {
					activeCols = append(activeCols, j-startCol)
				}
			}

			// Calculate utilization
			tileSize := (endRow - startRow) * (endCol - startCol)
			utilization := float64(nonzero) / float64(tileSize)

			// Skip empty tiles if compression enabled
			if m.Config.UseCompression && nonzero == 0 {
				continue
			}

			mapping := &CrossbarMapping{
				TileRow: ti,
				TileCol: tj,
				SparseWeights: &SparseWeights{
					Format:  SparsityBitmap,
					Rows:    endRow - startRow,
					Cols:    endCol - startCol,
					Bitmap:  tileBitmap,
					Dense:   tileWeights,
					Density: utilization,
				},
				ActiveRows:  activeRows,
				ActiveCols:  activeCols,
				Utilization: utilization,
			}

			mappings = append(mappings, mapping)
		}
	}

	m.Mappings = mappings
	return mappings
}

// GetTotalUtilization returns average utilization across all tiles
func (m *SparseCrossbarMapper) GetTotalUtilization() float64 {
	if len(m.Mappings) == 0 {
		return 0
	}
	sum := 0.0
	for _, mapping := range m.Mappings {
		sum += mapping.Utilization
	}
	return sum / float64(len(m.Mappings))
}

// GetAreaSavings returns area savings from sparsity compression
func (m *SparseCrossbarMapper) GetAreaSavings(originalTiles int) float64 {
	if originalTiles == 0 {
		return 0
	}
	return 1.0 - float64(len(m.Mappings))/float64(originalTiles)
}

// ============================================================================
// Index-Free Sparse Representation (Nature Electronics 2025)
// ============================================================================

// IndexFreeSparse represents index-free sparse format using dual-FeFET
type IndexFreeSparse struct {
	Rows          int
	Cols          int
	SparsityMask  [][]bool    // Digital sparsity pattern
	WeightValues  [][]float64 // Analog weight values
	CombinedWeights [][]float64 // Element-wise product (Hadamard)
}

// NewIndexFreeSparse creates index-free sparse representation
// This follows the dual-FeFET approach where sparsity is stored digitally
// and weights are stored in analog, then combined via Hadamard product
func NewIndexFreeSparse(weights [][]float64, sparsityMask [][]bool) *IndexFreeSparse {
	rows := len(weights)
	if rows == 0 {
		return nil
	}
	cols := len(weights[0])

	ifs := &IndexFreeSparse{
		Rows:           rows,
		Cols:           cols,
		SparsityMask:   sparsityMask,
		WeightValues:   weights,
		CombinedWeights: make([][]float64, rows),
	}

	// Compute Hadamard product
	for i := 0; i < rows; i++ {
		ifs.CombinedWeights[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			if i < len(sparsityMask) && j < len(sparsityMask[i]) && sparsityMask[i][j] {
				ifs.CombinedWeights[i][j] = weights[i][j]
			}
		}
	}

	return ifs
}

// MVM performs matrix-vector multiplication on index-free sparse
func (ifs *IndexFreeSparse) MVM(input []float64) []float64 {
	output := make([]float64, ifs.Rows)
	for i := 0; i < ifs.Rows; i++ {
		sum := 0.0
		for j := 0; j < ifs.Cols && j < len(input); j++ {
			sum += ifs.CombinedWeights[i][j] * input[j]
		}
		output[i] = sum
	}
	return output
}

// ============================================================================
// Sparsity Analysis
// ============================================================================

// SparsityStats contains sparsity statistics
type SparsityStats struct {
	TotalElements  int
	NonzeroCount   int
	ZeroCount      int
	Sparsity       float64
	Density        float64
	RowSparsity    []float64
	ColSparsity    []float64
	BlockSparsity  float64
	AvgMagnitude   float64
	MaxMagnitude   float64
}

// AnalyzeSparsity computes sparsity statistics
func AnalyzeSparsity(weights [][]float64, threshold float64) *SparsityStats {
	if len(weights) == 0 {
		return nil
	}
	rows := len(weights)
	cols := len(weights[0])

	stats := &SparsityStats{
		TotalElements: rows * cols,
		RowSparsity:   make([]float64, rows),
		ColSparsity:   make([]float64, cols),
	}

	colNonzero := make([]int, cols)
	sumMag := 0.0

	for i := 0; i < rows; i++ {
		rowNonzero := 0
		for j := 0; j < cols; j++ {
			absVal := math.Abs(weights[i][j])
			if absVal > threshold {
				stats.NonzeroCount++
				rowNonzero++
				colNonzero[j]++
			}
			sumMag += absVal
			if absVal > stats.MaxMagnitude {
				stats.MaxMagnitude = absVal
			}
		}
		stats.RowSparsity[i] = 1.0 - float64(rowNonzero)/float64(cols)
	}

	for j := 0; j < cols; j++ {
		stats.ColSparsity[j] = 1.0 - float64(colNonzero[j])/float64(rows)
	}

	stats.ZeroCount = stats.TotalElements - stats.NonzeroCount
	stats.Sparsity = float64(stats.ZeroCount) / float64(stats.TotalElements)
	stats.Density = 1.0 - stats.Sparsity
	stats.AvgMagnitude = sumMag / float64(stats.TotalElements)

	return stats
}

// AnalyzeBlockSparsity computes block-level sparsity
func AnalyzeBlockSparsity(weights [][]float64, blockSize int, threshold float64) float64 {
	if len(weights) == 0 {
		return 0
	}
	rows := len(weights)
	cols := len(weights[0])

	blockRows := (rows + blockSize - 1) / blockSize
	blockCols := (cols + blockSize - 1) / blockSize
	totalBlocks := blockRows * blockCols
	zeroBlocks := 0

	for bi := 0; bi < blockRows; bi++ {
		for bj := 0; bj < blockCols; bj++ {
			startRow := bi * blockSize
			startCol := bj * blockSize
			endRow := min(startRow+blockSize, rows)
			endCol := min(startCol+blockSize, cols)

			// Check if entire block is zero
			blockZero := true
			for i := startRow; i < endRow && blockZero; i++ {
				for j := startCol; j < endCol && blockZero; j++ {
					if math.Abs(weights[i][j]) > threshold {
						blockZero = false
					}
				}
			}
			if blockZero {
				zeroBlocks++
			}
		}
	}

	return float64(zeroBlocks) / float64(totalBlocks)
}

// ============================================================================
// Energy Estimation for Sparse CIM
// ============================================================================

// SparseEnergyModel estimates energy for sparse CIM operations
type SparseEnergyModel struct {
	MACEnergy        float64 // Energy per MAC (fJ)
	ADCEnergy        float64 // ADC energy per conversion (fJ)
	DACEnergy        float64 // DAC energy per conversion (fJ)
	RoutingEnergy    float64 // Inter-tile routing (fJ/bit)
	LeakagePower     float64 // Leakage power (nW/cell)
	CompressionRatio float64 // Effective compression from sparsity
}

// DefaultSparseEnergyModel returns typical CIM energy parameters
func DefaultSparseEnergyModel() *SparseEnergyModel {
	return &SparseEnergyModel{
		MACEnergy:        100,  // 100 fJ/MAC for FeFET CIM
		ADCEnergy:        500,  // 500 fJ per ADC conversion
		DACEnergy:        200,  // 200 fJ per DAC conversion
		RoutingEnergy:    10,   // 10 fJ/bit for on-chip routing
		LeakagePower:     0.1,  // 0.1 nW/cell
		CompressionRatio: 1.0,
	}
}

// EstimateEnergy estimates total energy for sparse MVM
func (em *SparseEnergyModel) EstimateEnergy(sparse *SparseWeights, batchSize int) float64 {
	if sparse == nil {
		return 0
	}

	// MAC energy (only for non-zero elements)
	macEnergy := float64(sparse.NumNonzero) * em.MACEnergy * float64(batchSize)

	// ADC energy (proportional to rows with non-zero outputs)
	adcEnergy := float64(sparse.Rows) * em.ADCEnergy * float64(batchSize)

	// DAC energy (for each input)
	dacEnergy := float64(sparse.Cols) * em.DACEnergy * float64(batchSize)

	// Apply compression ratio if using sparse tiles
	totalEnergy := macEnergy + adcEnergy + dacEnergy
	totalEnergy *= em.CompressionRatio

	return totalEnergy // in fJ
}

// EstimateEnergyReduction estimates energy reduction vs dense
func (em *SparseEnergyModel) EstimateEnergyReduction(sparse *SparseWeights) float64 {
	if sparse == nil {
		return 0
	}

	denseMACs := sparse.Rows * sparse.Cols
	sparseMACs := sparse.NumNonzero

	// Energy reduction primarily from MAC reduction
	return 1.0 - float64(sparseMACs)/float64(denseMACs)
}

// ============================================================================
// QUBO Matrix Compression (Nature Communications 2024)
// ============================================================================

// QUBOMatrix represents QUBO problem matrix for CIM annealer
type QUBOMatrix struct {
	Size       int
	Matrix     [][]float64
	Sparsity   float64
	Compressed *SparseWeights
}

// NewQUBOMatrix creates a QUBO matrix
func NewQUBOMatrix(matrix [][]float64) *QUBOMatrix {
	if len(matrix) == 0 {
		return nil
	}
	size := len(matrix)

	// Analyze sparsity
	nonzero := 0
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			if matrix[i][j] != 0 {
				nonzero++
			}
		}
	}

	return &QUBOMatrix{
		Size:     size,
		Matrix:   matrix,
		Sparsity: 1.0 - float64(nonzero)/float64(size*size),
	}
}

// Compress compresses QUBO matrix for FeFET crossbar
// Achieves up to 75% chip size savings for sparse QUBO (>85% sparsity)
func (q *QUBOMatrix) Compress() *SparseWeights {
	sparsifier := NewSparsifier(&SparsifyConfig{
		TargetSparsity: 0.0, // Keep all non-zero elements
		Method:         SparsifyMagnitude,
	})

	// Convert to sparse format
	q.Compressed = sparsifier.Sparsify(q.Matrix)
	return q.Compressed
}

// GetChipSavings returns estimated chip area savings
func (q *QUBOMatrix) GetChipSavings() float64 {
	// Based on Nature Communications 2024: 75% savings for 85%+ sparse QUBO
	if q.Sparsity >= 0.85 {
		return 0.75
	}
	// Linear interpolation for lower sparsity
	return q.Sparsity * 0.75 / 0.85
}
