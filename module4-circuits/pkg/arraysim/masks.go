package arraysim

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

func maskAt(mask [][]bool, r, c int) bool {
	if r < 0 || r >= len(mask) {
		return false
	}
	if c < 0 || c >= len(mask[r]) {
		return false
	}
	return mask[r][c]
}
