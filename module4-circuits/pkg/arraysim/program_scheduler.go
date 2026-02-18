package arraysim

import "fmt"

// ProgramOrderMode controls write/program sequencing for array cells.
type ProgramOrderMode string

const (
	ProgramOrderRowMajor     ProgramOrderMode = "row-major"
	ProgramOrderCheckerboard ProgramOrderMode = "checkerboard"
	ProgramOrderSerpentine   ProgramOrderMode = "serpentine"
	ProgramOrderAdaptive     ProgramOrderMode = "adaptive"
)

// CellIndex identifies a single array cell.
type CellIndex struct {
	Row int
	Col int
}

// ProgramScheduleResult captures a generated programming sequence and its disturb score.
type ProgramScheduleResult struct {
	Mode              ProgramOrderMode
	Rows              int
	Cols              int
	Order             []CellIndex
	CumulativeDisturb float64
}

// GenerateProgramSchedule generates programming order for an array and computes
// cumulative disturb under a local-neighbor model.
//
// Disturb model: each programmed cell contributes baseDisturb * (neighborsAlreadyProgrammed^2),
// where neighbors are 4-connected (N/S/E/W). Squared neighbor loading makes ordering matter.
func GenerateProgramSchedule(rows, cols int, mode ProgramOrderMode, baseDisturb float64) (ProgramScheduleResult, error) {
	if rows <= 0 || cols <= 0 {
		return ProgramScheduleResult{}, fmt.Errorf("invalid array dimensions rows=%d cols=%d", rows, cols)
	}
	if baseDisturb <= 0 {
		baseDisturb = 1.0
	}

	var order []CellIndex
	switch mode {
	case ProgramOrderRowMajor:
		order = scheduleRowMajor(rows, cols)
	case ProgramOrderCheckerboard:
		order = scheduleCheckerboard(rows, cols)
	case ProgramOrderSerpentine:
		order = scheduleSerpentine(rows, cols)
	case ProgramOrderAdaptive:
		order = scheduleAdaptive(rows, cols)
	default:
		return ProgramScheduleResult{}, fmt.Errorf("unknown scheduling mode: %q", mode)
	}

	return ProgramScheduleResult{
		Mode:              mode,
		Rows:              rows,
		Cols:              cols,
		Order:             order,
		CumulativeDisturb: cumulativeDisturb(order, rows, cols, baseDisturb),
	}, nil
}

func scheduleRowMajor(rows, cols int) []CellIndex {
	order := make([]CellIndex, 0, rows*cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			order = append(order, CellIndex{Row: r, Col: c})
		}
	}
	return order
}

func scheduleCheckerboard(rows, cols int) []CellIndex {
	order := make([]CellIndex, 0, rows*cols)
	for parity := 0; parity < 2; parity++ {
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				if (r+c)%2 == parity {
					order = append(order, CellIndex{Row: r, Col: c})
				}
			}
		}
	}
	return order
}

func scheduleSerpentine(rows, cols int) []CellIndex {
	order := make([]CellIndex, 0, rows*cols)
	for r := 0; r < rows; r++ {
		if r%2 == 0 {
			for c := 0; c < cols; c++ {
				order = append(order, CellIndex{Row: r, Col: c})
			}
			continue
		}
		for c := cols - 1; c >= 0; c-- {
			order = append(order, CellIndex{Row: r, Col: c})
		}
	}
	return order
}

// scheduleAdaptive greedily chooses the unprogrammed cell with the least already-programmed
// neighbor count (4-connected). Ties are broken deterministically by row-major index.
func scheduleAdaptive(rows, cols int) []CellIndex {
	total := rows * cols
	order := make([]CellIndex, 0, total)
	programmed := make([][]bool, rows)
	for r := range programmed {
		programmed[r] = make([]bool, cols)
	}

	for len(order) < total {
		best := CellIndex{Row: -1, Col: -1}
		bestCost := int(^uint(0) >> 1) // max int
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				if programmed[r][c] {
					continue
				}
				cost := programmedNeighborCount(programmed, r, c)
				if cost < bestCost || (cost == bestCost && (best.Row < 0 || r < best.Row || (r == best.Row && c < best.Col))) {
					bestCost = cost
					best = CellIndex{Row: r, Col: c}
				}
			}
		}
		programmed[best.Row][best.Col] = true
		order = append(order, best)
	}
	return order
}

func cumulativeDisturb(order []CellIndex, rows, cols int, baseDisturb float64) float64 {
	programmed := make([][]bool, rows)
	for r := range programmed {
		programmed[r] = make([]bool, cols)
	}
	var total float64
	n := len(order)
	for i, cell := range order {
		k := programmedNeighborCount(programmed, cell.Row, cell.Col)
		// Earlier writes are more disturb-sensitive due weaker shielding/settling.
		timeWeight := float64(n-i) / float64(n)
		total += baseDisturb * float64(k*k) * timeWeight
		programmed[cell.Row][cell.Col] = true
	}
	return total
}

func programmedNeighborCount(programmed [][]bool, r, c int) int {
	rows := len(programmed)
	cols := len(programmed[0])
	count := 0
	if r > 0 && programmed[r-1][c] {
		count++
	}
	if r+1 < rows && programmed[r+1][c] {
		count++
	}
	if c > 0 && programmed[r][c-1] {
		count++
	}
	if c+1 < cols && programmed[r][c+1] {
		count++
	}
	return count
}
