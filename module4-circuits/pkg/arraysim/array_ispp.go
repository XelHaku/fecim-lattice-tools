package arraysim

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"fecim-lattice-tools/shared/physics"
)

// ProgramOpts controls array-level ISPP execution.
type ProgramOpts struct {
	Order        string // "row-major", "col-major", "checkerboard"
	MaxPulses    int    // per cell budget (default 30)
	VerifyAfter  bool   // verify read after each cell
	AccumDisturb bool   // track cumulative half-select disturb
}

// CellResult stores per-cell programming outcomes.
type CellResult struct {
	Row, Col    int
	TargetLevel int
	FinalLevel  int
	LevelError  int
	PulsesUsed  int
	DisturbRecv float64 // cumulative disturb received (in level units)
	DisturbSent float64 // disturb caused to others
}

// ProgramResult stores array-level and per-cell summary metrics.
type ProgramResult struct {
	Cells          [][]CellResult
	TotalPulses    int
	WorstError     int
	MaxDisturb     float64
	AllConverged   bool
	ProgramTimeNs  float64
	TotalEnergy_fJ float64
}

// ProgramArray runs array-level ISPP with sequential updates and V/2 disturb tracking.
func ProgramArray(config ArrayConfig, targetLevels [][]int, opts ProgramOpts) (*ProgramResult, error) {
	cfg := withAnalysisDefaults(config)
	if cfg.Rows <= 0 || cfg.Cols <= 0 {
		return nil, errors.New("invalid array dimensions")
	}
	if len(targetLevels) != cfg.Rows {
		return nil, fmt.Errorf("targetLevels rows mismatch: got %d want %d", len(targetLevels), cfg.Rows)
	}
	for r := range targetLevels {
		if len(targetLevels[r]) != cfg.Cols {
			return nil, fmt.Errorf("targetLevels[%d] cols mismatch: got %d want %d", r, len(targetLevels[r]), cfg.Cols)
		}
	}
	if opts.MaxPulses <= 0 {
		opts.MaxPulses = 30
	}
	if opts.Order == "" {
		opts.Order = "row-major"
	}

	levels := physics.DefaultLevels
	if cfg.Material != nil {
		levels = cfg.Material.GetNumLevels()
	}
	if levels < 2 {
		levels = physics.DefaultLevels
	}

	// Initialize all cells at level 0.
	levelState := make([][]float64, cfg.Rows)
	conductance := make([][]float64, cfg.Rows)
	selfDisturb := make([][]float64, cfg.Rows)
	results := make([][]CellResult, cfg.Rows)
	for r := 0; r < cfg.Rows; r++ {
		levelState[r] = make([]float64, cfg.Cols)
		conductance[r] = make([]float64, cfg.Cols)
		selfDisturb[r] = make([]float64, cfg.Cols)
		results[r] = make([]CellResult, cfg.Cols)
		for c := 0; c < cfg.Cols; c++ {
			results[r][c] = CellResult{Row: r, Col: c, TargetLevel: clampLevel(targetLevels[r][c], levels)}
			conductance[r][c] = conductanceAtLevel(cfg.Material, 0, levels)
		}
	}

	order, err := programOrder(cfg.Rows, cfg.Cols, opts.Order)
	if err != nil {
		return nil, err
	}

	const pulseWidthNs = 100.0
	globalDisturb := 0.0
	totalPulses := 0
	totalEnergyFJ := 0.0

	for _, rc := range order {
		r, c := rc[0], rc[1]
		target := float64(clampLevel(targetLevels[r][c], levels))

		for pulse := 0; pulse < opts.MaxPulses; pulse++ {
			current := levelState[r][c]
			errLevel := target - current
			if math.Abs(errLevel) <= 0.5 {
				break
			}

			step := math.Copysign(math.Min(math.Abs(errLevel), 3.0), errLevel)
			writeV := 1.6 + 0.06*math.Abs(errLevel)
			if writeV > 3.3 {
				writeV = 3.3
			}

			wl := make([]float64, cfg.Rows)
			bl := make([]float64, cfg.Cols)
			for i := 0; i < cfg.Rows; i++ {
				wl[i] = writeV / 2
			}
			for j := 0; j < cfg.Cols; j++ {
				bl[j] = writeV / 2
			}
			wl[r] = writeV
			bl[c] = 0

			solveRes, ok := solveRead(cfg, SolveParams{
				WLVoltages:  wl,
				BLVoltages:  bl,
				Conductance: conductance,
				Geometry:    DefaultCellGeometry(),
				Wire:        cfg.Wire,
				Boundary:    cfg.Boundary,
			})
			eff := writeV
			if ok && r < len(solveRes.CellVoltages) && c < len(solveRes.CellVoltages[r]) {
				eff = math.Abs(solveRes.CellVoltages[r][c])
			}
			irFactor := 1.0
			if writeV > 1e-12 {
				irFactor = eff / writeV
			}

			actualStep := step * irFactor
			levelState[r][c] = clampLevelFloat(levelState[r][c]+actualStep, levels)
			conductance[r][c] = conductanceAtLevel(cfg.Material, int(math.Round(levelState[r][c])), levels)
			results[r][c].PulsesUsed++
			totalPulses++

			if ok && r < len(solveRes.CellCurrents) && c < len(solveRes.CellCurrents[r]) {
				i := math.Abs(solveRes.CellCurrents[r][c])
				totalEnergyFJ += writeV * i * pulseWidthNs * 1e6
			}

			if opts.AccumDisturb {
				delta := 0.0002 * (writeV / 2.0) // small V/2 disturb in level units
				globalDisturb += delta
				selfDisturb[r][c] += delta
				results[r][c].DisturbSent += delta * float64(cfg.Rows*cfg.Cols-1)
			}
		}

		if opts.VerifyAfter {
			results[r][c].FinalLevel = int(math.Round(levelState[r][c]))
			results[r][c].LevelError = results[r][c].FinalLevel - results[r][c].TargetLevel
		}
	}

	worstErr := 0
	maxDisturb := 0.0
	allConverged := true
	for r := 0; r < cfg.Rows; r++ {
		for c := 0; c < cfg.Cols; c++ {
			if opts.AccumDisturb {
				recv := globalDisturb - selfDisturb[r][c]
				results[r][c].DisturbRecv = recv
				levelState[r][c] = clampLevelFloat(levelState[r][c]+recv, levels)
			}
			final := int(math.Round(levelState[r][c]))
			results[r][c].FinalLevel = final
			results[r][c].LevelError = final - results[r][c].TargetLevel
			if absInt(results[r][c].LevelError) > 1 {
				allConverged = false
			}
			if absInt(results[r][c].LevelError) > worstErr {
				worstErr = absInt(results[r][c].LevelError)
			}
			if results[r][c].DisturbRecv > maxDisturb {
				maxDisturb = results[r][c].DisturbRecv
			}
		}
	}

	programTimeNs := float64(totalPulses) * pulseWidthNs

	return &ProgramResult{
		Cells:          results,
		TotalPulses:    totalPulses,
		WorstError:     worstErr,
		MaxDisturb:     maxDisturb,
		AllConverged:   allConverged,
		ProgramTimeNs:  programTimeNs,
		TotalEnergy_fJ: totalEnergyFJ,
	}, nil
}

func programOrder(rows, cols int, mode string) ([][2]int, error) {
	order := make([][2]int, 0, rows*cols)
	switch mode {
	case "row-major":
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				order = append(order, [2]int{r, c})
			}
		}
	case "col-major":
		for c := 0; c < cols; c++ {
			for r := 0; r < rows; r++ {
				order = append(order, [2]int{r, c})
			}
		}
	case "checkerboard":
		for parity := 0; parity < 2; parity++ {
			for r := 0; r < rows; r++ {
				for c := 0; c < cols; c++ {
					if (r+c)%2 == parity {
						order = append(order, [2]int{r, c})
					}
				}
			}
		}
	default:
		return nil, fmt.Errorf("unknown program order %q", mode)
	}
	return order, nil
}

func clampLevel(v, levels int) int {
	if v < 0 {
		return 0
	}
	if v >= levels {
		return levels - 1
	}
	return v
}

func clampLevelFloat(v float64, levels int) float64 {
	if v < 0 {
		return 0
	}
	max := float64(levels - 1)
	if v > max {
		return max
	}
	return v
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func (p *ProgramResult) MarshalJSONSummary() string {
	if p == nil {
		return "{}"
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
