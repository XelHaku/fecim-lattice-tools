//go:build legacy_fyne

package gui

import (
	"fmt"
	"math"
	"sort"
)

const sneakCurrentEpsilonA = 1e-12

type SneakCellImpact struct {
	Row      int     `json:"row"`
	Col      int     `json:"col"`
	CurrentA float64 `json:"current_a"`
}

type SneakPathMetrics struct {
	TotalSneakCurrentA float64           `json:"total_sneak_current_a"`
	MaxSneakCurrentA   float64           `json:"max_sneak_current_a"`
	AffectedCells      int               `json:"affected_cells"`
	HalfSelectCells    int               `json:"half_select_cells"`
	SneakOnlyCells     int               `json:"sneak_only_cells"`
	TopAffectedCells   []SneakCellImpact `json:"top_affected_cells"`
}

func computeSneakPathMetrics(currents [][]float64, selectedRow, selectedCol int) SneakPathMetrics {
	metrics := SneakPathMetrics{}
	if len(currents) == 0 {
		return metrics
	}

	top := make([]SneakCellImpact, 0, 8)
	for r := range currents {
		for c, current := range currents[r] {
			if r == selectedRow && c == selectedCol {
				continue
			}
			mag := math.Abs(current)
			if mag <= sneakCurrentEpsilonA {
				continue
			}
			metrics.TotalSneakCurrentA += mag
			metrics.AffectedCells++
			if mag > metrics.MaxSneakCurrentA {
				metrics.MaxSneakCurrentA = mag
			}
			if r == selectedRow || c == selectedCol {
				metrics.HalfSelectCells++
			} else {
				metrics.SneakOnlyCells++
			}
			top = append(top, SneakCellImpact{Row: r, Col: c, CurrentA: mag})
		}
	}

	sort.Slice(top, func(i, j int) bool {
		if top[i].CurrentA == top[j].CurrentA {
			if top[i].Row == top[j].Row {
				return top[i].Col < top[j].Col
			}
			return top[i].Row < top[j].Row
		}
		return top[i].CurrentA > top[j].CurrentA
	})
	if len(top) > 3 {
		top = top[:3]
	}
	metrics.TopAffectedCells = top
	return metrics
}

func formatSneakPathSummary(metrics SneakPathMetrics) string {
	if metrics.AffectedCells == 0 {
		return "0T1R: sneak current 0 A (0 affected cells)"
	}
	summary := fmt.Sprintf(
		"0T1R: sneak=%s, affected=%d (half-select=%d, sneak-only=%d)",
		formatCurrentA(metrics.TotalSneakCurrentA),
		metrics.AffectedCells,
		metrics.HalfSelectCells,
		metrics.SneakOnlyCells,
	)
	if len(metrics.TopAffectedCells) == 0 {
		return summary
	}
	top := metrics.TopAffectedCells[0]
	return fmt.Sprintf("%s, top=[%d,%d] %s", summary, top.Row, top.Col, formatCurrentA(top.CurrentA))
}
