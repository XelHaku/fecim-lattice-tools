// Package layers provides neural network layer implementations for crossbar arrays.
package layers

import (
	"math"
	"sort"
)

// PruningConfig holds configuration for weight pruning.
type PruningConfig struct {
	Sparsity       float64 // Target sparsity (0.0 - 1.0)
	Method         string  // "magnitude", "structured", "block"
	BlockSize      int     // Block size for block pruning
	GranularityCol bool    // Prune columns (structured)
	GranularityRow bool    // Prune rows (structured)
}

// DefaultPruningConfig returns default pruning configuration.
func DefaultPruningConfig(sparsity float64) *PruningConfig {
	return &PruningConfig{
		Sparsity:       sparsity,
		Method:         "magnitude",
		BlockSize:      8,
		GranularityCol: false,
		GranularityRow: false,
	}
}

// WeightPruner implements various pruning strategies for crossbar arrays.
type WeightPruner struct {
	config *PruningConfig
}

// NewWeightPruner creates a new weight pruner.
func NewWeightPruner(config *PruningConfig) *WeightPruner {
	if config == nil {
		config = DefaultPruningConfig(0.5)
	}
	return &WeightPruner{config: config}
}

// MagnitudePrune applies magnitude-based (unstructured) pruning.
// Zeroes out the smallest weights by magnitude.
func (wp *WeightPruner) MagnitudePrune(weights [][]float64) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])
	totalWeights := rows * cols

	// Collect all weights with positions
	type weightPos struct {
		row, col int
		value    float64
		absValue float64
	}

	allWeights := make([]weightPos, 0, totalWeights)
	for i := range weights {
		for j := range weights[i] {
			allWeights = append(allWeights, weightPos{
				row:      i,
				col:      j,
				value:    weights[i][j],
				absValue: math.Abs(weights[i][j]),
			})
		}
	}

	// Sort by absolute value
	sort.Slice(allWeights, func(i, j int) bool {
		return allWeights[i].absValue < allWeights[j].absValue
	})

	// Determine pruning threshold
	pruneCount := int(float64(totalWeights) * wp.config.Sparsity)

	// Create pruned weights
	pruned := make([][]float64, rows)
	for i := range pruned {
		pruned[i] = make([]float64, cols)
		copy(pruned[i], weights[i])
	}

	// Zero out smallest weights
	for i := 0; i < pruneCount && i < len(allWeights); i++ {
		wp := allWeights[i]
		pruned[wp.row][wp.col] = 0
	}

	return pruned
}

// StructuredPruneColumns prunes entire columns (output neurons).
// More efficient for crossbar arrays as it allows column deactivation.
func (wp *WeightPruner) StructuredPruneColumns(weights [][]float64) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])

	// Compute L2 norm for each column
	type colNorm struct {
		col  int
		norm float64
	}

	colNorms := make([]colNorm, cols)
	for j := 0; j < cols; j++ {
		sumSq := 0.0
		for i := 0; i < rows; i++ {
			sumSq += weights[i][j] * weights[i][j]
		}
		colNorms[j] = colNorm{col: j, norm: math.Sqrt(sumSq)}
	}

	// Sort columns by norm
	sort.Slice(colNorms, func(i, j int) bool {
		return colNorms[i].norm < colNorms[j].norm
	})

	// Determine columns to prune
	pruneCount := int(float64(cols) * wp.config.Sparsity)
	prunedCols := make(map[int]bool)
	for i := 0; i < pruneCount; i++ {
		prunedCols[colNorms[i].col] = true
	}

	// Create pruned weights
	pruned := make([][]float64, rows)
	for i := range pruned {
		pruned[i] = make([]float64, cols)
		for j := range pruned[i] {
			if prunedCols[j] {
				pruned[i][j] = 0
			} else {
				pruned[i][j] = weights[i][j]
			}
		}
	}

	return pruned
}

// StructuredPruneRows prunes entire rows (input features).
// Useful when inputs can be masked.
func (wp *WeightPruner) StructuredPruneRows(weights [][]float64) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])

	// Compute L2 norm for each row
	type rowNorm struct {
		row  int
		norm float64
	}

	rowNorms := make([]rowNorm, rows)
	for i := 0; i < rows; i++ {
		sumSq := 0.0
		for j := 0; j < cols; j++ {
			sumSq += weights[i][j] * weights[i][j]
		}
		rowNorms[i] = rowNorm{row: i, norm: math.Sqrt(sumSq)}
	}

	// Sort rows by norm
	sort.Slice(rowNorms, func(i, j int) bool {
		return rowNorms[i].norm < rowNorms[j].norm
	})

	// Determine rows to prune
	pruneCount := int(float64(rows) * wp.config.Sparsity)
	prunedRows := make(map[int]bool)
	for i := 0; i < pruneCount; i++ {
		prunedRows[rowNorms[i].row] = true
	}

	// Create pruned weights
	pruned := make([][]float64, rows)
	for i := range pruned {
		pruned[i] = make([]float64, cols)
		if prunedRows[i] {
			// Row is pruned, all zeros
			continue
		}
		copy(pruned[i], weights[i])
	}

	return pruned
}

// BlockPrune applies block-sparse pruning.
// Prunes entire blocks, which can map to crossbar tiles.
func (wp *WeightPruner) BlockPrune(weights [][]float64) [][]float64 {
	rows := len(weights)
	if rows == 0 {
		return weights
	}
	cols := len(weights[0])
	blockSize := wp.config.BlockSize

	// Number of blocks
	numBlocksRow := (rows + blockSize - 1) / blockSize
	numBlocksCols := (cols + blockSize - 1) / blockSize
	totalBlocks := numBlocksRow * numBlocksCols

	// Compute L2 norm for each block
	type blockNorm struct {
		blockRow, blockCol int
		norm               float64
	}

	blockNorms := make([]blockNorm, 0, totalBlocks)

	for br := 0; br < numBlocksRow; br++ {
		for bc := 0; bc < numBlocksCols; bc++ {
			sumSq := 0.0
			startRow := br * blockSize
			startCol := bc * blockSize
			endRow := min(startRow+blockSize, rows)
			endCol := min(startCol+blockSize, cols)

			for i := startRow; i < endRow; i++ {
				for j := startCol; j < endCol; j++ {
					sumSq += weights[i][j] * weights[i][j]
				}
			}

			blockNorms = append(blockNorms, blockNorm{
				blockRow: br,
				blockCol: bc,
				norm:     math.Sqrt(sumSq),
			})
		}
	}

	// Sort blocks by norm
	sort.Slice(blockNorms, func(i, j int) bool {
		return blockNorms[i].norm < blockNorms[j].norm
	})

	// Determine blocks to prune
	pruneCount := int(float64(totalBlocks) * wp.config.Sparsity)
	prunedBlocks := make(map[[2]int]bool)
	for i := 0; i < pruneCount && i < len(blockNorms); i++ {
		key := [2]int{blockNorms[i].blockRow, blockNorms[i].blockCol}
		prunedBlocks[key] = true
	}

	// Create pruned weights
	pruned := make([][]float64, rows)
	for i := range pruned {
		pruned[i] = make([]float64, cols)
		for j := range pruned[i] {
			br := i / blockSize
			bc := j / blockSize
			key := [2]int{br, bc}
			if prunedBlocks[key] {
				pruned[i][j] = 0
			} else {
				pruned[i][j] = weights[i][j]
			}
		}
	}

	return pruned
}

// Prune applies pruning based on configuration.
func (wp *WeightPruner) Prune(weights [][]float64) [][]float64 {
	switch wp.config.Method {
	case "magnitude":
		return wp.MagnitudePrune(weights)
	case "structured":
		if wp.config.GranularityCol {
			return wp.StructuredPruneColumns(weights)
		}
		return wp.StructuredPruneRows(weights)
	case "block":
		return wp.BlockPrune(weights)
	default:
		return wp.MagnitudePrune(weights)
	}
}

// GetSparsity computes actual sparsity of weight matrix.
func GetSparsity(weights [][]float64) float64 {
	total := 0
	zeros := 0
	for i := range weights {
		for j := range weights[i] {
			total++
			if weights[i][j] == 0 {
				zeros++
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(zeros) / float64(total)
}

// GetPruneMask returns binary mask indicating non-pruned weights.
func GetPruneMask(weights [][]float64) [][]bool {
	mask := make([][]bool, len(weights))
	for i := range mask {
		mask[i] = make([]bool, len(weights[i]))
		for j := range mask[i] {
			mask[i][j] = weights[i][j] != 0
		}
	}
	return mask
}

// ApplyMask applies a pruning mask to weights.
func ApplyMask(weights [][]float64, mask [][]bool) [][]float64 {
	result := make([][]float64, len(weights))
	for i := range result {
		result[i] = make([]float64, len(weights[i]))
		for j := range result[i] {
			if i < len(mask) && j < len(mask[i]) && mask[i][j] {
				result[i][j] = weights[i][j]
			}
		}
	}
	return result
}

// GradualPruner implements iterative pruning schedule.
// Gradually increases sparsity during training.
type GradualPruner struct {
	InitialSparsity float64
	FinalSparsity   float64
	StartStep       int
	EndStep         int
	PruneFrequency  int
	CurrentStep     int
	pruner          *WeightPruner
}

// NewGradualPruner creates a gradual pruner.
func NewGradualPruner(initialSparsity, finalSparsity float64, startStep, endStep, pruneFreq int) *GradualPruner {
	return &GradualPruner{
		InitialSparsity: initialSparsity,
		FinalSparsity:   finalSparsity,
		StartStep:       startStep,
		EndStep:         endStep,
		PruneFrequency:  pruneFreq,
		CurrentStep:     0,
		pruner:          NewWeightPruner(DefaultPruningConfig(initialSparsity)),
	}
}

// GetCurrentSparsity computes sparsity for current training step.
// Uses cubic schedule: s(t) = sf + (si - sf)(1 - (t-t0)/(tn-t0))^3
func (gp *GradualPruner) GetCurrentSparsity() float64 {
	if gp.CurrentStep < gp.StartStep {
		return gp.InitialSparsity
	}
	if gp.CurrentStep >= gp.EndStep {
		return gp.FinalSparsity
	}

	progress := float64(gp.CurrentStep-gp.StartStep) / float64(gp.EndStep-gp.StartStep)
	// Cubic schedule
	factor := math.Pow(1-progress, 3)
	return gp.FinalSparsity + (gp.InitialSparsity-gp.FinalSparsity)*factor
}

// ShouldPrune returns true if pruning should happen at current step.
func (gp *GradualPruner) ShouldPrune() bool {
	if gp.CurrentStep < gp.StartStep || gp.CurrentStep > gp.EndStep {
		return false
	}
	return gp.CurrentStep%gp.PruneFrequency == 0
}

// Prune applies pruning at current sparsity level.
func (gp *GradualPruner) Prune(weights [][]float64) [][]float64 {
	gp.pruner.config.Sparsity = gp.GetCurrentSparsity()
	return gp.pruner.Prune(weights)
}

// Step increments the training step.
func (gp *GradualPruner) Step() {
	gp.CurrentStep++
}

// SortedWeightSectioning implements SWS for efficient crossbar mapping.
// Sorts weights by magnitude and sections them for bit-sliced crossbars.
type SortedWeightSectioning struct {
	NumSections int     // Number of sections (bit slices)
	Threshold   float64 // Threshold for zero-skipping
}

// NewSortedWeightSectioning creates SWS instance.
func NewSortedWeightSectioning(numSections int) *SortedWeightSectioning {
	return &SortedWeightSectioning{
		NumSections: numSections,
		Threshold:   1e-6,
	}
}

// Section applies sorted weight sectioning.
// Returns sections and indices for reconstruction.
func (sws *SortedWeightSectioning) Section(weights [][]float64) (sections [][][]float64, indices [][]int) {
	rows := len(weights)
	if rows == 0 {
		return nil, nil
	}
	cols := len(weights[0])

	// Flatten and sort by magnitude
	type weightIdx struct {
		row, col int
		value    float64
		absValue float64
	}

	flat := make([]weightIdx, 0, rows*cols)
	for i := range weights {
		for j := range weights[i] {
			flat = append(flat, weightIdx{
				row:      i,
				col:      j,
				value:    weights[i][j],
				absValue: math.Abs(weights[i][j]),
			})
		}
	}

	// Sort by absolute value (descending)
	sort.Slice(flat, func(i, j int) bool {
		return flat[i].absValue > flat[j].absValue
	})

	// Divide into sections
	sectionSize := len(flat) / sws.NumSections
	if sectionSize == 0 {
		sectionSize = 1
	}

	sections = make([][][]float64, sws.NumSections)
	indices = make([][]int, sws.NumSections)

	for s := 0; s < sws.NumSections; s++ {
		sections[s] = make([][]float64, rows)
		for i := range sections[s] {
			sections[s][i] = make([]float64, cols)
		}

		startIdx := s * sectionSize
		endIdx := startIdx + sectionSize
		if s == sws.NumSections-1 {
			endIdx = len(flat)
		}

		for idx := startIdx; idx < endIdx && idx < len(flat); idx++ {
			w := flat[idx]
			sections[s][w.row][w.col] = w.value
			indices[s] = append(indices[s], w.row*cols+w.col)
		}
	}

	return sections, indices
}

// GetADCRequirements estimates ADC bits needed for each section.
func (sws *SortedWeightSectioning) GetADCRequirements(sections [][][]float64) []int {
	adcBits := make([]int, len(sections))

	for s, section := range sections {
		// Find max accumulation in section
		maxAccum := 0.0
		for i := range section {
			rowSum := 0.0
			for j := range section[i] {
				rowSum += math.Abs(section[i][j])
			}
			if rowSum > maxAccum {
				maxAccum = rowSum
			}
		}

		// Estimate ADC bits based on dynamic range
		// Lower sections have smaller weights, need fewer bits
		if maxAccum > 0 {
			adcBits[s] = int(math.Ceil(math.Log2(maxAccum))) + 1
		} else {
			adcBits[s] = 1
		}
	}

	return adcBits
}

// CrossbarSparsityMapper maps sparse weights to crossbar efficiently.
type CrossbarSparsityMapper struct {
	TileRows int
	TileCols int
}

// NewCrossbarSparsityMapper creates a sparsity mapper.
func NewCrossbarSparsityMapper(tileRows, tileCols int) *CrossbarSparsityMapper {
	return &CrossbarSparsityMapper{
		TileRows: tileRows,
		TileCols: tileCols,
	}
}

// MapToTiles maps sparse weight matrix to crossbar tiles.
// Returns tile assignments and skip masks.
func (csm *CrossbarSparsityMapper) MapToTiles(weights [][]float64) (tiles [][][]float64, skipMasks [][][]bool) {
	rows := len(weights)
	if rows == 0 {
		return nil, nil
	}
	cols := len(weights[0])

	numTileRows := (rows + csm.TileRows - 1) / csm.TileRows
	numTileCols := (cols + csm.TileCols - 1) / csm.TileCols

	tiles = make([][][]float64, numTileRows*numTileCols)
	skipMasks = make([][][]bool, numTileRows*numTileCols)

	tileIdx := 0
	for tr := 0; tr < numTileRows; tr++ {
		for tc := 0; tc < numTileCols; tc++ {
			tiles[tileIdx] = make([][]float64, csm.TileRows)
			skipMasks[tileIdx] = make([][]bool, csm.TileRows)

			for i := 0; i < csm.TileRows; i++ {
				tiles[tileIdx][i] = make([]float64, csm.TileCols)
				skipMasks[tileIdx][i] = make([]bool, csm.TileCols)

				srcRow := tr*csm.TileRows + i
				for j := 0; j < csm.TileCols; j++ {
					srcCol := tc*csm.TileCols + j

					if srcRow < rows && srcCol < cols {
						tiles[tileIdx][i][j] = weights[srcRow][srcCol]
						skipMasks[tileIdx][i][j] = weights[srcRow][srcCol] == 0
					} else {
						// Padding
						skipMasks[tileIdx][i][j] = true
					}
				}
			}
			tileIdx++
		}
	}

	return tiles, skipMasks
}

// GetTileUtilization computes utilization (non-zero ratio) per tile.
func (csm *CrossbarSparsityMapper) GetTileUtilization(tiles [][][]float64) []float64 {
	utilization := make([]float64, len(tiles))

	for t, tile := range tiles {
		total := 0
		nonZero := 0
		for i := range tile {
			for j := range tile[i] {
				total++
				if tile[i][j] != 0 {
					nonZero++
				}
			}
		}
		if total > 0 {
			utilization[t] = float64(nonZero) / float64(total)
		}
	}

	return utilization
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
