package crossbar

// LIT-P3-02: Multi-hop sneak path model for passive crossbar arrays > 128×128.
//
// The existing 3-cell (1-hop) model captures the dominant sneak mechanism in
// small arrays, but for large passive arrays (>128×128) there are many more
// 5-cell (2-hop) paths that collectively contribute significant parasitic
// current even though each individual path is weaker (5 resistors in series
// vs 3).
//
// Topology:
//
//	3-cell (1-hop): WL_sR → G(sR,j1) → BL_j1 → G(i1,j1) → WL_i1 → G(i1,sC) → BL_sC
//	5-cell (2-hop): ... → G(i1,j2) → BL_j2 → G(i2,j2) → WL_i2 → G(i2,sC) → BL_sC
//
// Path counts for N×N array:
//
//	3-cell: (N-1)² paths  →  O(N²)
//	5-cell: (N-1)²(N-2)² paths  →  O(N⁴)
//
// For N=128 there are ~261 million 5-cell paths — exact enumeration is
// infeasible. Arrays with rows×cols ≤ 1024 use exact O(N⁴) enumeration;
// larger arrays use Monte Carlo sampling with extrapolation.
//
// Reference: IEEE Trans. Electron Devices (Linn et al. 2010), CrossSim arXiv 2025.

import (
	"math/rand"
)

// multiHopSneakFields stores the extra fields injected into SneakPathAnalysis
// by AnalyzeSneakPathsMultiHop. They are zero-valued when the struct is
// created by the single-hop AnalyzeSneakPathsWithIsolation call.
//
// These fields are embedded in SneakPathAnalysis via the MultiHopFields
// embedded struct (see nonidealities.go). Adding them here keeps the
// nonidealities.go changes minimal.

// SneakMultiHopResult holds the multi-hop contribution returned by
// AnalyzeSneakPathsMultiHop in addition to the base SneakPathAnalysis.
type SneakMultiHopResult struct {
	// Base analysis (3-cell paths only, same as AnalyzeSneakPathsWithIsolation).
	*SneakPathAnalysis

	// FiveHopSneak is the total conductance-weighted sneak current from all
	// 5-cell (2-hop) paths to the selected output bit line (post-isolation).
	FiveHopSneak float64

	// FiveHopRatio is FiveHopSneak / base TotalSneak (3-cell).
	// Values > 1 indicate that 5-cell paths dominate for this array size.
	FiveHopRatio float64

	// TotalSneakMultiHop is FiveHopSneak + base TotalSneak (all hops combined).
	TotalSneakMultiHop float64

	// IsSampled is true when the 5-cell result was obtained by Monte Carlo
	// sampling (array too large for exact enumeration: rows×cols > 1024).
	IsSampled bool

	// SampleCount is the number of random 5-cell paths sampled (0 = exact).
	SampleCount int
}

// sneakFiveCellTotal computes the total conductance-weighted 5-cell sneak
// contribution to BL_sC from WL_sR for a passive crossbar.
//
// For arrays with rows×cols ≤ 1024 (roughly ≤32×32) the result is exact.
// For larger arrays, 1000 random paths are sampled and the result is scaled
// to the total path count.
//
// Returns (total, sampled bool, sampleCount int).
func (a *Array) sneakFiveCellTotal(sR, sC int, isolFactor float64) (float64, bool, int) {
	rows := a.config.Rows
	cols := a.config.Cols

	// Minimum array size required: at least 2 distinct rows and 3 distinct cols
	// (sR, one i1; sC, j1, j2). Below this threshold no 5-cell paths exist.
	if rows < 3 || cols < 3 {
		return 0, false, 0
	}

	// Exact for small arrays; sampled for large.
	useExact := rows*cols <= 1024

	if useExact {
		return a.sneakFiveCellExact(sR, sC, isolFactor)
	}
	return a.sneakFiveCellSampled(sR, sC, isolFactor, 1000)
}

// sneakFiveCellExact enumerates all 5-cell paths exactly.
// Complexity: O((C-1)(R-1)(C-2)(R-2)) ≈ O(N⁴) — safe only for small N.
func (a *Array) sneakFiveCellExact(sR, sC int, isolFactor float64) (float64, bool, int) {
	rows := a.config.Rows
	cols := a.config.Cols

	var total float64
	var count int

	for j1 := 0; j1 < cols; j1++ {
		if j1 == sC {
			continue
		}
		g1 := a.cells[sR][j1].Conductance
		if g1 <= 0 {
			continue
		}
		for i1 := 0; i1 < rows; i1++ {
			if i1 == sR {
				continue
			}
			g2 := a.cells[i1][j1].Conductance
			if g2 <= 0 {
				continue
			}
			for j2 := 0; j2 < cols; j2++ {
				if j2 == sC || j2 == j1 {
					continue
				}
				g3 := a.cells[i1][j2].Conductance
				if g3 <= 0 {
					continue
				}
				for i2 := 0; i2 < rows; i2++ {
					if i2 == sR || i2 == i1 {
						continue
					}
					g4 := a.cells[i2][j2].Conductance
					if g4 <= 0 {
						continue
					}
					g5 := a.cells[i2][sC].Conductance
					if g5 <= 0 {
						continue
					}
					gSeries := 1.0 / (1/g1 + 1/g2 + 1/g3 + 1/g4 + 1/g5)
					total += gSeries * isolFactor
					count++
				}
			}
		}
	}

	return total, false, count
}

// sneakFiveCellSampled estimates the 5-cell sneak total by Monte Carlo sampling.
// nSamples random 5-cell paths are drawn uniformly; the mean is extrapolated
// to the total path count.
func (a *Array) sneakFiveCellSampled(sR, sC int, isolFactor float64, nSamples int) (float64, bool, int) {
	rows := a.config.Rows
	cols := a.config.Cols

	// Build candidate index slices to avoid repeated modular arithmetic.
	jCands := make([]int, 0, cols-1) // all j ≠ sC
	for j := 0; j < cols; j++ {
		if j != sC {
			jCands = append(jCands, j)
		}
	}
	iCands := make([]int, 0, rows-1) // all i ≠ sR
	for i := 0; i < rows; i++ {
		if i != sR {
			iCands = append(iCands, i)
		}
	}
	if len(jCands) < 2 || len(iCands) < 2 {
		return 0, true, 0 // Not enough dimensions for 5-cell paths
	}

	// Total number of (j1,i1,j2,i2) 4-tuples:
	//   (C-1) * (R-1) * (C-2) * (R-2)
	totalPaths := int64(len(jCands)) * int64(len(iCands)) *
		int64(len(jCands)-1) * int64(len(iCands)-1)

	rng := rand.New(rand.NewSource(42))

	var sumG float64
	var hit int // samples with all-positive conductances

	for s := 0; s < nSamples; s++ {
		// Pick j1 from jCands.
		j1 := jCands[rng.Intn(len(jCands))]
		// Pick i1 from iCands.
		i1 := iCands[rng.Intn(len(iCands))]
		// Pick j2 from jCands, j2 ≠ j1 — rejection sample.
		j2Idx := rng.Intn(len(jCands) - 1)
		j1Pos := 0
		for k, jv := range jCands {
			if jv == j1 {
				j1Pos = k
				break
			}
		}
		if j2Idx >= j1Pos {
			j2Idx++
		}
		j2 := jCands[j2Idx]

		// Pick i2 from iCands, i2 ≠ i1 — rejection sample.
		i2Idx := rng.Intn(len(iCands) - 1)
		i1Pos := 0
		for k, iv := range iCands {
			if iv == i1 {
				i1Pos = k
				break
			}
		}
		if i2Idx >= i1Pos {
			i2Idx++
		}
		i2 := iCands[i2Idx]

		g1 := a.cells[sR][j1].Conductance
		g2 := a.cells[i1][j1].Conductance
		g3 := a.cells[i1][j2].Conductance
		g4 := a.cells[i2][j2].Conductance
		g5 := a.cells[i2][sC].Conductance

		if g1 <= 0 || g2 <= 0 || g3 <= 0 || g4 <= 0 || g5 <= 0 {
			continue
		}
		sumG += 1.0 / (1/g1 + 1/g2 + 1/g3 + 1/g4 + 1/g5)
		hit++
	}

	if hit == 0 {
		return 0, true, 0
	}

	// Mean series conductance per sampled path (including zero-conductance paths).
	meanG := sumG / float64(nSamples) // denominator: ALL samples, not just hits
	total := meanG * float64(totalPaths) * isolFactor
	return total, true, hit
}

// AnalyzeSneakPathsMultiHop performs multi-hop sneak path analysis.
//
// maxHops controls which path lengths are included:
//   - maxHops = 1: 3-cell paths only (identical to AnalyzeSneakPathsWithIsolation)
//   - maxHops = 2: 3-cell and 5-cell paths
//
// For arrays with rows×cols ≤ 1024 the 5-cell result is exact; larger arrays
// use Monte Carlo sampling with 1000 random paths (IsSampled=true).
func (a *Array) AnalyzeSneakPathsMultiHop(selectedRow, selectedCol int, isolationFactor float64, maxHops int) *SneakMultiHopResult {
	base := a.AnalyzeSneakPathsWithIsolation(selectedRow, selectedCol, isolationFactor)
	result := &SneakMultiHopResult{
		SneakPathAnalysis:  base,
		TotalSneakMultiHop: base.TotalSneak,
	}

	if maxHops < 2 {
		return result
	}

	fiveHop, sampled, nSampled := a.sneakFiveCellTotal(selectedRow, selectedCol, isolationFactor)
	result.FiveHopSneak = fiveHop
	result.IsSampled = sampled
	result.SampleCount = nSampled
	result.TotalSneakMultiHop = base.TotalSneak + fiveHop
	if base.TotalSneak > 0 {
		result.FiveHopRatio = fiveHop / base.TotalSneak
	}

	getLog().Calculation("AnalyzeSneakPathsMultiHop", map[string]interface{}{
		"selectedRow":        selectedRow,
		"selectedCol":        selectedCol,
		"maxHops":            maxHops,
		"threeHopSneak":      base.TotalSneak,
		"fiveHopSneak":       fiveHop,
		"fiveHopRatio":       result.FiveHopRatio,
		"totalSneakMultiHop": result.TotalSneakMultiHop,
		"isSampled":          sampled,
	}, nil)

	return result
}

// FiveHopScalingFactor returns the ratio of 5-cell sneak to 3-cell sneak for
// the given array size, assuming uniform conductance G.
//
// Analytically:
//
//	3-cell sneak total ≈ (R-1)(C-1) × G/3
//	5-cell sneak total ≈ (R-1)(R-2)(C-1)(C-2) × G/5
//	ratio = (R-2)(C-2)/5 × 3 = 3(R-2)(C-2)/5
//
// For a 128×128 array: ratio ≈ 3×126×126/5 ≈ 9525  (5-cell paths dominate!).
// For a 8×8 array:    ratio ≈ 3×6×6/5 ≈ 21.6.
// This is why multi-hop matters for large passive arrays.
func FiveHopScalingFactor(rows, cols int) float64 {
	if rows < 3 || cols < 3 {
		return 0
	}
	return 3.0 * float64(rows-2) * float64(cols-2) / 5.0
}
