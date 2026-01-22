// pkg/compiler/compiler.go
package compiler

import (
	"fmt"
	"math"
)

// Compile transforms a weight matrix into crossbar cell assignments
func Compile(weights [][]float64, config CompileConfig) (*CrossbarMapping, error) {
	// Validate
	if len(weights) == 0 || len(weights[0]) == 0 {
		return nil, fmt.Errorf("empty weight matrix")
	}

	rows := len(weights)
	cols := len(weights[0])

	if rows > config.ArrayRows || cols > config.ArrayCols {
		return nil, fmt.Errorf("weights %dx%d exceed array %dx%d",
			rows, cols, config.ArrayRows, config.ArrayCols)
	}

	// Find weight range for symmetric quantization
	wMin, wMax := weights[0][0], weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < wMin {
				wMin = weights[i][j]
			}
			if weights[i][j] > wMax {
				wMax = weights[i][j]
			}
		}
	}
	wAbsMax := math.Max(math.Abs(wMin), math.Abs(wMax))

	// Compile each weight
	var cells []CellAssignment
	var mseSum float64
	levelsUsed := make(map[int]bool)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			w := weights[i][j]

			// Quantize: map [-wAbsMax, +wAbsMax] to [0, Levels-1]
			normalized := (w + wAbsMax) / (2 * wAbsMax) // 0 to 1
			level := int(math.Round(normalized * float64(config.Levels-1)))
			level = clamp(level, 0, config.Levels-1)
			levelsUsed[level] = true

			// Dequantize to get quantized value
			qNorm := float64(level) / float64(config.Levels-1)
			qValue := -wAbsMax + qNorm*(2*wAbsMax)
			mseSum += (w - qValue) * (w - qValue)

			// Map to physical parameters
			gNorm := float64(level) / float64(config.Levels-1)
			conductance := config.GMin + gNorm*(config.GMax-config.GMin)
			progV := config.VProgMin + gNorm*(config.VProgMax-config.VProgMin)

			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				WeightValue: w,
				QuantLevel:  level,
				Conductance: conductance,
				ProgramV:    progV,
			})
		}
	}

	// Calculate statistics
	numCells := rows * cols
	mse := mseSum / float64(numCells)
	psnr := 100.0
	if mse > 0 {
		psnr = 10 * math.Log10((wAbsMax*wAbsMax)/mse)
	}

	return &CrossbarMapping{
		Config: config,
		Cells:  cells,
		Stats: Stats{
			TotalCells:   config.ArrayRows * config.ArrayCols,
			UsedCells:    numCells,
			Utilization:  float64(numCells) / float64(config.ArrayRows*config.ArrayCols),
			WeightMin:    wMin,
			WeightMax:    wMax,
			QuantMSE:     mse,
			QuantPSNR:    psnr,
			UniqueLevels: len(levelsUsed),
		},
	}, nil
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
