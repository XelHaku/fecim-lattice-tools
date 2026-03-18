package crossbar

import (
	"fmt"
	"math"
	"testing"
)

// simpleMVM performs a naive matrix-vector multiply: y[i] = Sum_j W[i][j]*x[j].
// Used as a reference oracle for tiled MVM tests.
func simpleMVM(weights [][]float64, input []float64) ([]float64, error) {
	rows := len(weights)
	output := make([]float64, rows)
	for i := 0; i < rows; i++ {
		for j := 0; j < len(input); j++ {
			output[i] += weights[i][j] * input[j]
		}
	}
	return output, nil
}

// makeWeightMatrix creates a deterministic rows x cols weight matrix.
func makeWeightMatrix(rows, cols int) [][]float64 {
	w := make([][]float64, rows)
	for i := range w {
		w[i] = make([]float64, cols)
		for j := range w[i] {
			// Deterministic pseudo-random values in [0, 1)
			w[i][j] = float64((i*cols+j)*7%97) / 97.0
		}
	}
	return w
}

// makeInputVector creates a deterministic input vector of length n.
func makeInputVector(n int) []float64 {
	v := make([]float64, n)
	for i := range v {
		v[i] = float64((i*13+5)%53) / 53.0
	}
	return v
}

func TestTileWeightMatrix_SingleTile(t *testing.T) {
	// A 64x64 matrix with 256x256 tiles fits in a single tile.
	weights := makeWeightMatrix(64, 64)
	cfg := TilingConfig{MaxRows: 256, MaxCols: 256, Padding: "zero"}

	tw, err := TileWeightMatrix(weights, cfg)
	if err != nil {
		t.Fatalf("TileWeightMatrix failed: %v", err)
	}

	if tw.TotalTiles() != 1 {
		t.Errorf("expected 1 tile, got %d", tw.TotalTiles())
	}
	if tw.TileGrid != [2]int{1, 1} {
		t.Errorf("expected grid [1,1], got %v", tw.TileGrid)
	}
	if tw.OrigShape != [2]int{64, 64} {
		t.Errorf("expected origShape [64,64], got %v", tw.OrigShape)
	}

	// Verify tile dimensions
	tile := tw.Tiles[0]
	if tile.Rows != 64 || tile.Cols != 64 {
		t.Errorf("expected tile 64x64, got %dx%d", tile.Rows, tile.Cols)
	}
	if tile.RowOffset != 0 || tile.ColOffset != 0 {
		t.Errorf("expected offset (0,0), got (%d,%d)", tile.RowOffset, tile.ColOffset)
	}

	// Verify weight data
	for i := 0; i < 64; i++ {
		for j := 0; j < 64; j++ {
			if tile.Weights[i][j] != weights[i][j] {
				t.Fatalf("weight mismatch at (%d,%d): %.6f != %.6f", i, j, tile.Weights[i][j], weights[i][j])
			}
		}
	}

	// Efficiency should be 1.0 since matrix exactly fills the tile
	eff := tw.Efficiency()
	if math.Abs(eff-1.0) > 1e-9 {
		t.Errorf("expected efficiency 1.0, got %.6f", eff)
	}
}

func TestTileWeightMatrix_2x2Grid(t *testing.T) {
	// A 512x512 matrix with 256x256 tiles yields exactly 4 tiles.
	weights := makeWeightMatrix(512, 512)
	cfg := TilingConfig{MaxRows: 256, MaxCols: 256, Padding: "zero"}

	tw, err := TileWeightMatrix(weights, cfg)
	if err != nil {
		t.Fatalf("TileWeightMatrix failed: %v", err)
	}

	if tw.TotalTiles() != 4 {
		t.Errorf("expected 4 tiles, got %d", tw.TotalTiles())
	}
	if tw.TileGrid != [2]int{2, 2} {
		t.Errorf("expected grid [2,2], got %v", tw.TileGrid)
	}

	// Verify tile layout
	expectedOffsets := [][2]int{{0, 0}, {0, 256}, {256, 0}, {256, 256}}
	for i, tile := range tw.Tiles {
		if tile.RowOffset != expectedOffsets[i][0] || tile.ColOffset != expectedOffsets[i][1] {
			t.Errorf("tile %d: expected offset (%d,%d), got (%d,%d)",
				i, expectedOffsets[i][0], expectedOffsets[i][1], tile.RowOffset, tile.ColOffset)
		}
		if tile.Rows != 256 || tile.Cols != 256 {
			t.Errorf("tile %d: expected 256x256, got %dx%d", i, tile.Rows, tile.Cols)
		}
	}

	// Efficiency should be 1.0 since 512 = 2*256 exactly
	eff := tw.Efficiency()
	if math.Abs(eff-1.0) > 1e-9 {
		t.Errorf("expected efficiency 1.0, got %.6f", eff)
	}
}

func TestTileWeightMatrix_Rectangular(t *testing.T) {
	// A 300x500 matrix with 256x256 tiles yields a 2x2 grid of tiles.
	// Tile sizes: row tiles [256, 44], col tiles [256, 244]
	weights := makeWeightMatrix(300, 500)
	cfg := TilingConfig{MaxRows: 256, MaxCols: 256, Padding: "zero"}

	tw, err := TileWeightMatrix(weights, cfg)
	if err != nil {
		t.Fatalf("TileWeightMatrix failed: %v", err)
	}

	if tw.TileGrid != [2]int{2, 2} {
		t.Errorf("expected grid [2,2], got %v", tw.TileGrid)
	}
	if tw.TotalTiles() != 4 {
		t.Errorf("expected 4 tiles, got %d", tw.TotalTiles())
	}

	// Check partial tile sizes
	// Tile (0,0): 256x256, (0,1): 256x244, (1,0): 44x256, (1,1): 44x244
	expected := []struct{ r, c int }{
		{256, 256}, {256, 244}, {44, 256}, {44, 244},
	}
	for i, tile := range tw.Tiles {
		if tile.Rows != expected[i].r || tile.Cols != expected[i].c {
			t.Errorf("tile %d: expected %dx%d, got %dx%d",
				i, expected[i].r, expected[i].c, tile.Rows, tile.Cols)
		}
	}

	// Efficiency should be < 1.0 due to partial tiles
	eff := tw.Efficiency()
	// Used = 300*500 = 150000
	// Allocated = 256*256 + 256*244 + 44*256 + 44*244 = 65536 + 62464 + 11264 + 10736 = 150000
	// Actually the tiles exactly cover the original — no padding waste, so efficiency = 1.0
	if math.Abs(eff-1.0) > 1e-9 {
		t.Errorf("expected efficiency 1.0 for exact cover, got %.6f", eff)
	}
}

func TestTiledMVM_MatchesMonolithic(t *testing.T) {
	// Verify that tiled MVM produces the same result as monolithic MVM.
	sizes := []struct {
		rows, cols, tileSize int
	}{
		{64, 64, 32},    // 2x2 tiles
		{100, 200, 64},  // 2x4 tiles
		{512, 512, 256}, // 2x2 tiles
		{300, 500, 128}, // 3x4 tiles
	}

	for _, s := range sizes {
		t.Run(
			func() string {
				return fmt.Sprintf("%dx%d_tile%d", s.rows, s.cols, s.tileSize)
			}(),
			func(t *testing.T) {
				weights := makeWeightMatrix(s.rows, s.cols)
				input := makeInputVector(s.cols)

				// Monolithic reference
				expected, err := simpleMVM(weights, input)
				if err != nil {
					t.Fatalf("simpleMVM failed: %v", err)
				}

				// Tiled MVM
				cfg := TilingConfig{MaxRows: s.tileSize, MaxCols: s.tileSize, Padding: "zero"}
				tw, err := TileWeightMatrix(weights, cfg)
				if err != nil {
					t.Fatalf("TileWeightMatrix failed: %v", err)
				}

				got, err := tw.TiledMVM(input, simpleMVM)
				if err != nil {
					t.Fatalf("TiledMVM failed: %v", err)
				}

				if len(got) != len(expected) {
					t.Fatalf("output length mismatch: %d != %d", len(got), len(expected))
				}

				maxErr := 0.0
				for i := range got {
					err := math.Abs(got[i] - expected[i])
					if err > maxErr {
						maxErr = err
					}
				}

				// Should match to floating-point precision (summation order may differ slightly)
				tol := 1e-10
				if maxErr > tol {
					t.Errorf("tiled MVM max error %.2e exceeds tolerance %.0e", maxErr, tol)
				}
				t.Logf("Tiles: %d, max error: %.2e", tw.TotalTiles(), maxErr)
			},
		)
	}
}

func TestTiledWeight_Efficiency(t *testing.T) {
	tests := []struct {
		name     string
		rows     int
		cols     int
		tileSize int
		wantEff  float64 // expected efficiency (approximate)
	}{
		{"exact_fit", 256, 256, 256, 1.0},
		{"exact_2x2", 512, 512, 256, 1.0},
		// 100x100 on 256x256 tile: used=10000, allocated=10000 (single tile, trimmed)
		{"small_single", 100, 100, 256, 1.0},
		// 257x257 on 256x256: used=66049, allocated=256*256+256*1+1*256+1*1=65536+256+256+1=66049
		{"one_over", 257, 257, 256, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weights := makeWeightMatrix(tt.rows, tt.cols)
			cfg := TilingConfig{MaxRows: tt.tileSize, MaxCols: tt.tileSize, Padding: "zero"}

			tw, err := TileWeightMatrix(weights, cfg)
			if err != nil {
				t.Fatalf("TileWeightMatrix failed: %v", err)
			}

			eff := tw.Efficiency()
			if math.Abs(eff-tt.wantEff) > 0.01 {
				t.Errorf("efficiency: got %.4f, want ~%.4f", eff, tt.wantEff)
			}
		})
	}
}

func TestTileWeightMatrix_MNIST(t *testing.T) {
	// MNIST layer 1: 784 inputs -> 128 hidden units.
	// For MVM y = W*x, weight matrix is [output x input] = [128 x 784].
	// With 256x256 tiles: ceil(128/256)=1 row tile, ceil(784/256)=4 col tiles.
	weights := makeWeightMatrix(128, 784)
	cfg := TilingConfig{MaxRows: 256, MaxCols: 256, Padding: "zero"}

	tw, err := TileWeightMatrix(weights, cfg)
	if err != nil {
		t.Fatalf("TileWeightMatrix failed: %v", err)
	}

	// Expect 1x4 = 4 tiles (1 row of tiles, 4 columns of tiles)
	if tw.TileGrid != [2]int{1, 4} {
		t.Errorf("expected grid [1,4], got %v", tw.TileGrid)
	}
	if tw.TotalTiles() != 4 {
		t.Errorf("expected 4 tiles for 128x784 on 256x256, got %d", tw.TotalTiles())
	}

	// Verify tile sizes:
	// Tile 0: rows 0-127, cols 0-255 => 128x256
	// Tile 1: rows 0-127, cols 256-511 => 128x256
	// Tile 2: rows 0-127, cols 512-767 => 128x256
	// Tile 3: rows 0-127, cols 768-783 => 128x16
	expectedTiles := []struct {
		rowOff, colOff, rows, cols int
	}{
		{0, 0, 128, 256},
		{0, 256, 128, 256},
		{0, 512, 128, 256},
		{0, 768, 128, 16},
	}
	for i, tile := range tw.Tiles {
		exp := expectedTiles[i]
		if tile.RowOffset != exp.rowOff || tile.ColOffset != exp.colOff {
			t.Errorf("tile %d offset: got (%d,%d), want (%d,%d)",
				i, tile.RowOffset, tile.ColOffset, exp.rowOff, exp.colOff)
		}
		if tile.Rows != exp.rows || tile.Cols != exp.cols {
			t.Errorf("tile %d size: got %dx%d, want %dx%d",
				i, tile.Rows, tile.Cols, exp.rows, exp.cols)
		}
	}

	// Efficiency: used = 128*784 = 100352
	// allocated = 3*128*256 + 128*16 = 98304 + 2048 = 100352
	eff := tw.Efficiency()
	if math.Abs(eff-1.0) > 1e-9 {
		t.Errorf("MNIST layer efficiency: got %.6f, want 1.0", eff)
	}

	// Verify tiled MVM matches monolithic for MNIST-like input (784 pixels)
	input := makeInputVector(784)
	expected, _ := simpleMVM(weights, input)
	got, err := tw.TiledMVM(input, simpleMVM)
	if err != nil {
		t.Fatalf("TiledMVM failed: %v", err)
	}

	maxErr := 0.0
	for i := range got {
		e := math.Abs(got[i] - expected[i])
		if e > maxErr {
			maxErr = e
		}
	}
	if maxErr > 1e-10 {
		t.Errorf("MNIST tiled MVM max error %.2e exceeds tolerance", maxErr)
	}
	t.Logf("MNIST 128x784: %d tiles, efficiency=%.2f%%, max_error=%.2e",
		tw.TotalTiles(), eff*100, maxErr)
}

func TestTileWeightMatrix_ValidationErrors(t *testing.T) {
	// Empty matrix
	_, err := TileWeightMatrix(nil, DefaultTilingConfig())
	if err == nil {
		t.Error("expected error for nil matrix")
	}

	// Zero MaxRows
	_, err = TileWeightMatrix(makeWeightMatrix(4, 4), TilingConfig{MaxRows: 0, MaxCols: 256})
	if err == nil {
		t.Error("expected error for MaxRows=0")
	}

	// Overlap >= MaxRows
	_, err = TileWeightMatrix(makeWeightMatrix(4, 4), TilingConfig{MaxRows: 8, MaxCols: 8, Overlap: 8})
	if err == nil {
		t.Error("expected error when overlap >= tile size")
	}

	// Invalid padding
	_, err = TileWeightMatrix(makeWeightMatrix(4, 4), TilingConfig{MaxRows: 8, MaxCols: 8, Padding: "invalid"})
	if err == nil {
		t.Error("expected error for invalid padding")
	}
}

