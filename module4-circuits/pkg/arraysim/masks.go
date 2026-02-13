package arraysim

import "math"

func selectorEnabled(params SolveParams, r, c int) bool {
	switch params.SelectorMode {
	case SelectorRead:
		if params.ReadMask == nil {
			return true
		}
		return maskAt(params.ReadMask, r, c)
	case SelectorWrite:
		if params.WriteMask == nil {
			return true
		}
		return maskAt(params.WriteMask, r, c)
	default:
		return true
	}
}

func effectiveCellConductance(params SolveParams, r, c int) float64 {
	gCell := 0.0
	if r >= 0 && r < len(params.Conductance) && c >= 0 && c < len(params.Conductance[r]) {
		gCell = params.Conductance[r][c]
	}
	if gCell <= 0 {
		return 0
	}

	enabled := selectorEnabled(params, r, c)
	gEff := gCell
	if !params.Selector.Enabled {
		if !enabled {
			return 0
		}
	} else {
		gSel := params.Selector.OffConductance
		if enabled {
			gSel = params.Selector.OnConductance
		}
		gEff = seriesConductance(gEff, gSel)
	}

	if params.SelectorEnabled && params.SelectorRon > 0 {
		if gEff <= 0 {
			return 0
		}
		gEff = 1.0 / (1.0/gEff + params.SelectorRon)
	}
	return gEff
}

func seriesConductance(g1, g2 float64) float64 {
	if g1 <= 0 || g2 <= 0 {
		return 0
	}
	if math.IsInf(g2, 1) {
		return g1
	}
	if math.IsInf(g1, 1) {
		return g2
	}
	return 1.0 / (1.0/g1 + 1.0/g2)
}

func maskAt(mask [][]bool, r, c int) bool {
	if r < 0 || r >= len(mask) {
		return false
	}
	if c < 0 || c >= len(mask[r]) {
		return false
	}
	return mask[r][c]
}
