// pkg/compiler/compiler.go
package compiler

import (
	"fmt"
	"math"

	"fecim-lattice-tools/shared/logging"
	"fecim-lattice-tools/shared/mathutil"
)

var log = logging.NewLogger("eda-compiler")

// Compiler constants
const (
	// conductanceToResistance converts µS to Ω: R = 1e6 / G_µS.
	conductanceToResistance = 1e6

	// standbyPowerMW is the estimated standby power for a blank array (10 µW).
	standbyPowerMW = 0.01

	// activePowerMW is the estimated active power for a programmed array (100 µW).
	activePowerMW = 0.1

	// cellAreaConversion converts µm² to mm² (1e-6).
	cellAreaConversion = 1e-6

	// maxPSNR is the PSNR ceiling when MSE is negligible (100 dB).
	maxPSNR = 100.0

	// minMSEThreshold is the MSE below which PSNR is capped to maxPSNR.
	minMSEThreshold = 1e-9
)

// GenerateDesign is the main entry point.
// Transforms configuration into a physical array design.
// If config.ComputeConfig.InitialWeights is provided, performs mapping.
// Otherwise, generates a blank initialized array.
func GenerateDesign(config *ArrayConfig) (*ArrayDesign, error) {
	log.Input("GenerateDesign", map[string]interface{}{
		"mode":       config.Mode,
		"arrayRows":  config.ArrayRows,
		"arrayCols":  config.ArrayCols,
		"levels":     config.Levels,
		"hasWeights": config.ComputeConfig != nil && config.ComputeConfig.InitialWeights != nil,
	})

	var design *ArrayDesign
	var err error

	// Check for Compute Mode with Weights
	if config.Mode == ModeCompute &&
		config.ComputeConfig != nil &&
		config.ComputeConfig.InitialWeights != nil {
		design, err = mapWeights(config)
	} else {
		// Default: Generate Blank Array
		design = GenerateBlank(config)
	}

	if err != nil {
		log.ErrorContext("GenerateDesign", err, map[string]interface{}{
			"mode": config.Mode,
		})
		return nil, err
	}

	log.Calculation("GenerateDesign", map[string]interface{}{
		"totalCells":  design.Stats.TotalCells,
		"activeCells": design.Stats.ActiveCells,
		"areaMM2":     design.Stats.AreaMM2,
		"powerMW":     design.Stats.PowerMW,
	}, design)

	return design, nil
}

// GenerateBlank creates an initialized array without weights
func GenerateBlank(config *ArrayConfig) *ArrayDesign {
	log.Input("GenerateBlank", map[string]interface{}{
		"rows":      config.ArrayRows,
		"cols":      config.ArrayCols,
		"cellPitch": config.CellPitch,
		"rowHeight": config.RowHeight,
	})

	// Use config dimensions
	rows := config.ArrayRows
	cols := config.ArrayCols

	// Pre-allocate for performance
	cells := make([]CellAssignment, 0, rows*cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Initialize to reset state (Level 0)
			resistance := 0.0
			if config.GMin > 0 {
				resistance = conductanceToResistance / config.GMin
			}
			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				Level:       0,
				Conductance: config.GMin,
				Resistance:  resistance,
				ProgramV:    config.VProgMin,
			})
		}
	}

	// Calculate area estimates
	// Standard cell area (SKY130 ref) ~ 0.25um^2 = 0.25e-6 mm^2
	// Pitch is in microns. Area = (Pitch * Height) * 1e-6 mm^2
	cellAreaMM2 := (config.CellPitch * config.RowHeight) * cellAreaConversion
	totalAreaMM2 := float64(rows*cols) * cellAreaMM2

	design := &ArrayDesign{
		Config: config,
		Cells:  cells,
		Stats: DesignStats{
			TotalCells:     rows * cols,
			ActiveCells:    0, // None programmed/used
			AreaMM2:        totalAreaMM2,
			PowerMW:        standbyPowerMW,
			ThroughputGOPS: 0.0,
		},
	}

	log.Calculation("GenerateBlank", map[string]interface{}{
		"totalCells": rows * cols,
		"areaMM2":    totalAreaMM2,
	}, design)

	return design
}

// mapWeights handles the compute-mode mapping logic
func mapWeights(config *ArrayConfig) (*ArrayDesign, error) {
	weights := config.ComputeConfig.InitialWeights

	if len(weights) == 0 || len(weights[0]) == 0 {
		err := fmt.Errorf("empty weight matrix provided")
		log.ErrorContext("mapWeights", err, nil)
		return nil, err
	}

	rows := len(weights)
	cols := len(weights[0])

	log.Input("mapWeights", map[string]interface{}{
		"weightRows": rows,
		"weightCols": cols,
		"arrayRows":  config.ArrayRows,
		"arrayCols":  config.ArrayCols,
		"levels":     config.Levels,
	})

	if rows > config.ArrayRows || cols > config.ArrayCols {
		err := fmt.Errorf("weights %dx%d exceed array %dx%d",
			rows, cols, config.ArrayRows, config.ArrayCols)
		log.ErrorContext("mapWeights", err, map[string]interface{}{
			"weightRows": rows,
			"weightCols": cols,
			"arrayRows":  config.ArrayRows,
			"arrayCols":  config.ArrayCols,
		})
		return nil, err
	}

	// Find weight range for quantization
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
	if wAbsMax == 0 {
		wAbsMax = 1.0
	}

	var cells []CellAssignment
	var mseSum float64
	levelsUsed := make(map[int]bool)

	// Iterate over physical array dimensions
	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			var w float64 = 0.0
			var level int = 0
			var conductance float64 = config.GMin
			var progV float64 = config.VProgMin

			// Map if within weight matrix bounds
			if i < rows && j < cols {
				w = weights[i][j]

				// Quantize
				normalized := (w + wAbsMax) / (2 * wAbsMax)
				level = int(math.Round(normalized * float64(config.Levels-1)))
				level = mathutil.ClampInt(level, 0, config.Levels-1)
				levelsUsed[level] = true

				// Dequantize for stats
				var qNorm float64
				if config.Levels > 1 {
					qNorm = float64(level) / float64(config.Levels-1)
				}
				qValue := -wAbsMax + qNorm*(2*wAbsMax)
				mseSum += (w - qValue) * (w - qValue)

				conductance = config.GMin + qNorm*(config.GMax-config.GMin)
				progV = config.VProgMin + qNorm*(config.VProgMax-config.VProgMin)
			}

			var resistance float64
			if conductance > 0 {
				resistance = conductanceToResistance / conductance
			}
			cells = append(cells, CellAssignment{
				Row:           i,
				Col:           j,
				Level:         level,
				Conductance:   conductance,
				Resistance:    resistance,
				ProgramV:      progV,
				InitialWeight: w,
			})
		}
	}

	// Stats
	numWeights := rows * cols
	mse := mseSum / float64(numWeights)
	psnr := maxPSNR
	if mse > minMSEThreshold {
		psnr = 10 * math.Log10((wAbsMax*wAbsMax)/mse)
	}

	totalAreaMM2 := float64(config.ArrayRows*config.ArrayCols) * (config.CellPitch * config.RowHeight) * cellAreaConversion

	design := &ArrayDesign{
		Config: config,
		Cells:  cells,
		Stats: DesignStats{
			TotalCells:  config.ArrayRows * config.ArrayCols,
			ActiveCells: numWeights,
			UsedCells:   numWeights, // Deprecated: kept for backward compatibility
			AreaMM2:     totalAreaMM2,
			PowerMW:     activePowerMW,
			QuantMSE:    mse,
			QuantPSNR:   psnr,
			WeightMin:   wMin,
			WeightMax:   wMax,
		},
	}

	log.Calculation("mapWeights", map[string]interface{}{
		"totalCells":    design.Stats.TotalCells,
		"activeWeights": numWeights,
		"levelsUsed":    len(levelsUsed),
		"quantMSE":      mse,
		"quantPSNR":     psnr,
		"areaMM2":       totalAreaMM2,
	}, design)

	return design, nil
}

// Compile is the Legacy wrapper for backward compatibility
// Deprecated: Use GenerateDesign instead
func Compile(weights [][]float64, legacyConfig CompileConfig) (*CrossbarMapping, error) {
	// Convert legacy config to new ArrayConfig
	// Note: CompileConfig is aliased to ArrayConfig in types.go, so casting works directly
	// or we just use it as is if it's the exact same struct layout.
	// However, since ArrayConfig added fields (StorageConfig, etc), we should be careful.
	// Since types.go says 'type CompileConfig = ArrayConfig', they are identical types.

	config := &legacyConfig // Pointer to the config

	// If weights provided, attach them to the config for mapping
	if weights != nil {
		config.Mode = ModeCompute
		config.ComputeConfig = &ComputeArrayConfig{
			InitialWeights: weights,
			QuantLevels:    config.Levels,
		}
	}

	return GenerateDesign(config)
}

