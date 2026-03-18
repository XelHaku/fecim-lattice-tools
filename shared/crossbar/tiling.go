// Package crossbar implements ferroelectric crossbar array simulation.
// This file provides a weight tiling/mapping engine for splitting large neural
// network weight matrices across multiple crossbar tiles. When a layer exceeds
// a single crossbar array (e.g., 784x128 MNIST layer on 256x256 tiles), the
// tiling engine partitions the weight matrix and later merges partial MVM
// outputs via accumulation.
package crossbar

import (
	"fmt"
	"math"
)

// TilingConfig configures how large weight matrices are tiled onto crossbar arrays.
type TilingConfig struct {
	MaxRows int    // Maximum rows per tile (e.g., 256)
	MaxCols int    // Maximum columns per tile (e.g., 256)
	Overlap int    // Overlap rows/cols for boundary handling (0 = none)
	Padding string // "zero" or "replicate" for partial tiles
}

// DefaultTilingConfig returns a sensible default for 256x256 crossbar arrays.
func DefaultTilingConfig() TilingConfig {
	return TilingConfig{
		MaxRows: 256,
		MaxCols: 256,
		Overlap: 0,
		Padding: "zero",
	}
}

// TiledWeight represents a weight matrix split across multiple crossbar tiles.
type TiledWeight struct {
	Tiles     []WeightTile // All tiles in row-major order
	OrigShape [2]int       // Original matrix dimensions [rows, cols]
	TileGrid  [2]int       // Number of tiles in [row, col] direction
	Config    TilingConfig
}

// WeightTile is one crossbar array's worth of weights.
type WeightTile struct {
	RowOffset int         // Starting row in original matrix
	ColOffset int         // Starting column in original matrix
	Rows      int         // Tile height (may be < MaxRows for last row of tiles)
	Cols      int         // Tile width (may be < MaxCols for last col of tiles)
	Weights   [][]float64 // Quantized weights for this tile
}

// TileWeightMatrix splits a large weight matrix into tiles that each fit within
// config.MaxRows x config.MaxCols. Tiles are ordered row-major: the tile at
// grid position (tr, tc) is at index tr*TileGrid[1]+tc.
//
// For MVM (y = W*x), the input dimension is cols and the output dimension is rows.
// A tiled MVM works by:
//  1. Slicing the input vector for each column-tile
//  2. Computing partial MVM per tile
//  3. Accumulating outputs from tiles in the same tile-row
func TileWeightMatrix(weights [][]float64, config TilingConfig) (*TiledWeight, error) {
	if err := validateTilingConfig(config); err != nil {
		return nil, err
	}
	if len(weights) == 0 {
		return nil, fmt.Errorf("tiling: weight matrix is empty")
	}

	totalRows := len(weights)
	totalCols := len(weights[0])
	for i, row := range weights {
		if len(row) != totalCols {
			return nil, fmt.Errorf("tiling: jagged matrix at row %d (expected %d cols, got %d)", i, totalCols, len(row))
		}
	}

	// Compute effective step size (tile stride) accounting for overlap.
	rowStep := config.MaxRows - config.Overlap
	colStep := config.MaxCols - config.Overlap
	if rowStep <= 0 || colStep <= 0 {
		return nil, fmt.Errorf("tiling: overlap (%d) must be less than tile size (rows=%d, cols=%d)",
			config.Overlap, config.MaxRows, config.MaxCols)
	}

	// Number of tiles in each direction.
	tileRows := ceilDiv(totalRows, rowStep)
	tileCols := ceilDiv(totalCols, colStep)

	tiles := make([]WeightTile, 0, tileRows*tileCols)

	for tr := 0; tr < tileRows; tr++ {
		rowStart := tr * rowStep
		rowEnd := rowStart + config.MaxRows
		if rowEnd > totalRows {
			rowEnd = totalRows
		}
		tileH := rowEnd - rowStart

		for tc := 0; tc < tileCols; tc++ {
			colStart := tc * colStep
			colEnd := colStart + config.MaxCols
			if colEnd > totalCols {
				colEnd = totalCols
			}
			tileW := colEnd - colStart

			// Extract sub-matrix for this tile.
			tileWeights := make([][]float64, tileH)
			for r := 0; r < tileH; r++ {
				tileWeights[r] = make([]float64, tileW)
				copy(tileWeights[r], weights[rowStart+r][colStart:colEnd])
			}

			tiles = append(tiles, WeightTile{
				RowOffset: rowStart,
				ColOffset: colStart,
				Rows:      tileH,
				Cols:      tileW,
				Weights:   tileWeights,
			})
		}
	}

	return &TiledWeight{
		Tiles:     tiles,
		OrigShape: [2]int{totalRows, totalCols},
		TileGrid:  [2]int{tileRows, tileCols},
		Config:    config,
	}, nil
}

// ReconstructOutput merges tile-level MVM outputs back into the full output vector.
// Each tileOutput[i] corresponds to Tiles[i] and contains the partial sums from
// that tile's MVM. Tiles sharing the same output rows (same tile-row) have their
// partial sums accumulated (column-direction summation via Kirchhoff's current law).
//
// The tileOutputs slice must have len == len(tw.Tiles), and each element must have
// length equal to the corresponding tile's Rows.
func (tw *TiledWeight) ReconstructOutput(tileOutputs [][]float64) ([]float64, error) {
	if len(tileOutputs) != len(tw.Tiles) {
		return nil, fmt.Errorf("tiling: expected %d tile outputs, got %d",
			len(tw.Tiles), len(tileOutputs))
	}

	output := make([]float64, tw.OrigShape[0])

	for i, tile := range tw.Tiles {
		tileOut := tileOutputs[i]
		if len(tileOut) != tile.Rows {
			return nil, fmt.Errorf("tiling: tile %d output length %d != tile rows %d",
				i, len(tileOut), tile.Rows)
		}

		// If there is overlap, only accumulate non-overlapped portion except
		// for the first tile in each row direction. With Overlap==0, all rows
		// are accumulated (the standard case for CIM tiling).
		for r := 0; r < tile.Rows; r++ {
			output[tile.RowOffset+r] += tileOut[r]
		}
	}

	return output, nil
}

// TotalTiles returns the total number of crossbar tiles used.
func (tw *TiledWeight) TotalTiles() int {
	return len(tw.Tiles)
}

// Efficiency returns the ratio of used weight cells to total allocated cells
// across all tiles. A value of 1.0 means every cell in every tile holds a
// meaningful weight; values below 1.0 indicate wasted cells in partial tiles.
func (tw *TiledWeight) Efficiency() float64 {
	usedCells := tw.OrigShape[0] * tw.OrigShape[1]
	if usedCells == 0 {
		return 0
	}

	totalAllocated := 0
	for _, tile := range tw.Tiles {
		totalAllocated += tile.Rows * tile.Cols
	}
	if totalAllocated == 0 {
		return 0
	}

	return float64(usedCells) / float64(totalAllocated)
}

// TileIndex returns the tile grid coordinates (tileRow, tileCol) for a flat tile index.
func (tw *TiledWeight) TileIndex(flat int) (tileRow, tileCol int) {
	tileCol = flat % tw.TileGrid[1]
	tileRow = flat / tw.TileGrid[1]
	return
}

// TiledMVM performs a full matrix-vector multiplication using tiled computation.
// It slices the input vector for each column-tile, performs per-tile MVM using
// the provided mvmFn, and accumulates partial sums.
//
// mvmFn is called for each tile with the tile's weight matrix and the
// corresponding input slice. It should return the partial output vector.
// This allows the caller to use crossbar.Array.MVM or a simple matmul.
func (tw *TiledWeight) TiledMVM(input []float64, mvmFn func(weights [][]float64, inp []float64) ([]float64, error)) ([]float64, error) {
	if len(input) != tw.OrigShape[1] {
		return nil, fmt.Errorf("tiling: input length %d != original cols %d",
			len(input), tw.OrigShape[1])
	}

	tileOutputs := make([][]float64, len(tw.Tiles))

	for i, tile := range tw.Tiles {
		// Slice the input vector for this tile's column range.
		tileInput := input[tile.ColOffset : tile.ColOffset+tile.Cols]

		out, err := mvmFn(tile.Weights, tileInput)
		if err != nil {
			return nil, fmt.Errorf("tiling: MVM on tile %d failed: %w", i, err)
		}
		tileOutputs[i] = out
	}

	return tw.ReconstructOutput(tileOutputs)
}

// validateTilingConfig checks that the config fields are sensible.
func validateTilingConfig(cfg TilingConfig) error {
	if cfg.MaxRows <= 0 {
		return fmt.Errorf("tiling: MaxRows must be > 0, got %d", cfg.MaxRows)
	}
	if cfg.MaxCols <= 0 {
		return fmt.Errorf("tiling: MaxCols must be > 0, got %d", cfg.MaxCols)
	}
	if cfg.Overlap < 0 {
		return fmt.Errorf("tiling: Overlap must be >= 0, got %d", cfg.Overlap)
	}
	if cfg.Padding != "" && cfg.Padding != "zero" && cfg.Padding != "replicate" {
		return fmt.Errorf("tiling: Padding must be \"zero\" or \"replicate\", got %q", cfg.Padding)
	}
	return nil
}

// ceilDiv returns ceil(a/b) for positive integers.
func ceilDiv(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
}
