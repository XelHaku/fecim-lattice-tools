package arraysim

import (
	"fmt"
	"math"

	"fecim-lattice-tools/shared/peripherals"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

const defaultReadMarginThresholdV = 50e-3

// ReadMarginResult summarizes adjacent-level separation after the full read chain.
type ReadMarginResult struct {
	ArraySize      int
	Levels         int
	CouplingMode   string
	MinMarginV     float64   // minimum delta-V between adjacent levels after sense chain
	MarginPerLevel []float64 // margin between each pair of adjacent levels
	Reliable       bool      // margin > threshold (e.g., > 50mV or > 1 LSB)
}

func (m CouplingMode) String() string {
	switch m {
	case CouplingIdeal:
		return "Ideal"
	case CouplingTierA:
		return "Tier-A"
	case CouplingTierB:
		return "Tier-B"
	default:
		return fmt.Sprintf("Unknown(%d)", int(m))
	}
}

// ReadMarginAnalysis computes level-to-level read margin after DAC→array→TIA→ADC.
func ReadMarginAnalysis(config ArrayConfig, levels int) ReadMarginResult {
	cfg := withAnalysisDefaults(config)
	rows, cols := cfg.Rows, cfg.Cols

	levelVout := make([][][]float64, levels)
	for level := 0; level < levels; level++ {
		g := conductanceAtLevel(cfg.Material, level, levels)
		levelVout[level] = readAllCellsADCInput(cfg, g)
	}

	margins := make([]float64, 0, levels-1)
	globalMin := math.Inf(1)
	for level := 0; level < levels-1; level++ {
		pairMin := math.Inf(1)
		for r := 0; r < rows; r++ {
			for c := 0; c < cols; c++ {
				delta := math.Abs(levelVout[level+1][r][c] - levelVout[level][r][c])
				if delta < pairMin {
					pairMin = delta
				}
				if delta < globalMin {
					globalMin = delta
				}
			}
		}
		if math.IsInf(pairMin, 1) {
			pairMin = 0
		}
		margins = append(margins, pairMin)
	}

	if math.IsInf(globalMin, 1) {
		globalMin = 0
	}

	threshold := reliabilityThreshold(cfg.Sense)
	return ReadMarginResult{
		ArraySize:      rows,
		Levels:         levels,
		CouplingMode:   cfg.CouplingMode.String(),
		MinMarginV:     globalMin,
		MarginPerLevel: margins,
		Reliable:       globalMin > threshold,
	}
}

func withAnalysisDefaults(cfg ArrayConfig) ArrayConfig {
	if cfg.Rows <= 0 {
		cfg.Rows = 32
	}
	if cfg.Cols <= 0 {
		cfg.Cols = cfg.Rows
	}
	if cfg.ReadVoltageV == 0 {
		cfg.ReadVoltageV = 0.2
	}
	filmGeom := cfg.Geometry.WithDefaults()
	arrayGeom := DefaultCellGeometry()
	arrayGeom.Film = filmGeom
	cfg.Wire = cfg.Wire.WithDefaults(arrayGeom)
	cfg.Boundary = cfg.Boundary.WithDefaults(cfg.Wire)
	cfg.Geometry = filmGeom
	if cfg.CouplingMode != CouplingIdeal && cfg.CouplingMode != CouplingTierA && cfg.CouplingMode != CouplingTierB {
		cfg.CouplingMode = CouplingTierA
	}
	if cfg.Sense.TIA.Rf <= 0 || cfg.Sense.ADC.Bits <= 0 {
		tia := peripherals.DefaultTIA()
		adc := peripherals.DefaultADC()
		cfg.Sense = SenseChain{
			TIA: TIAConfig{Rf: tia.Gain, Vref: tia.OutputOffset, Vmin: 0, Vmax: tia.MaxOutputVoltage},
			ADC: ADCConfig{Bits: adc.Bits, Vmin: adc.VrefLow, Vmax: adc.VrefHigh},
		}
	}
	if cfg.Material == nil {
		cfg.Material = sharedphysics.FeCIMMaterial()
	}
	return cfg
}

func conductanceAtLevel(mat *sharedphysics.HZOMaterial, level, levels int) float64 {
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	return mat.DiscreteLevel(level, levels)
}

func readAllCellsADCInput(cfg ArrayConfig, gCell float64) [][]float64 {
	vouts := make([][]float64, cfg.Rows)
	for r := 0; r < cfg.Rows; r++ {
		vouts[r] = make([]float64, cfg.Cols)
		for c := 0; c < cfg.Cols; c++ {
			scale := couplingScale(cfg.CouplingMode, cfg.Rows, cfg.Cols, r, c)
			currentA := gCell * cfg.ReadVoltageV * scale
			vouts[r][c] = cfg.Sense.ConvertCurrent(currentA).Vout
		}
	}
	return vouts
}

func couplingScale(mode CouplingMode, rows, cols, r, c int) float64 {
	distance := float64(r+c) / float64(rows+cols)
	switch mode {
	case CouplingIdeal:
		return 1.0
	case CouplingTierA:
		return 1.0 / (1.0 + 1.0*distance + 0.002*float64(rows))
	case CouplingTierB:
		return 1.0 / (1.0 + 1.5*distance + 0.003*float64(rows))
	default:
		return 1.0
	}
}

func solveRead(cfg ArrayConfig, params SolveParams) (SolveResult, bool) {
	switch cfg.CouplingMode {
	case CouplingIdeal:
		rows := len(params.Conductance)
		cols := 0
		if rows > 0 {
			cols = len(params.Conductance[0])
		}
		out := SolveResult{CellVoltages: make([][]float64, rows), CellCurrents: make([][]float64, rows), RowCurrents: make([]float64, rows), ColCurrents: make([]float64, cols)}
		for r := 0; r < rows; r++ {
			out.CellVoltages[r] = make([]float64, cols)
			out.CellCurrents[r] = make([]float64, cols)
			for c := 0; c < cols; c++ {
				v := params.WLVoltages[r] - params.BLVoltages[c]
				i := params.Conductance[r][c] * v
				out.CellVoltages[r][c], out.CellCurrents[r][c] = v, i
				out.RowCurrents[r] += i
				out.ColCurrents[c] += i
			}
		}
		return out, true
	case CouplingTierA:
		res, err := NewTierASolver().Solve(params)
		return res, err == nil
	case CouplingTierB:
		res, err := NewTierBSolver().Solve(params)
		return res, err == nil
	default:
		return SolveResult{}, false
	}
}

func reliabilityThreshold(s SenseChain) float64 {
	if s.ADC.Bits <= 0 {
		return defaultReadMarginThresholdV
	}
	codes := (1 << uint(s.ADC.Bits)) - 1
	if codes <= 0 {
		return defaultReadMarginThresholdV
	}
	lsb := (s.ADC.Vmax - s.ADC.Vmin) / float64(codes)
	if lsb > defaultReadMarginThresholdV {
		return lsb
	}
	return defaultReadMarginThresholdV
}
