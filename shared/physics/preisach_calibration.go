// preisach_calibration.go provides functions to calibrate TanhEverett Preisach
// parameters (Ps, Ec, Delta) against experimental P-E loop data.
//
// The calibration extracts Pr and Ec from measured data, then uses golden-section
// search on Delta to minimize RMSE between a simulated Preisach loop and the
// experimental trace.
package physics

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// PEPoint is defined in worldclass_cv.go with fields:
//   Field_Vm        float64  // Electric field (V/m)
//   Polarization_Cm float64  // Polarization (C/m^2)
// All functions in this file use that existing type.

// PreisachCalibrationResult holds the fitted Preisach parameters and
// goodness-of-fit statistics from CalibratePreisachToData.
type PreisachCalibrationResult struct {
	Ps    float64 // Fitted saturation polarization (C/m^2)
	Ec    float64 // Fitted coercive field (V/m)
	Delta float64 // Fitted Everett distribution width (V/m)
	RMSE  float64 // Root-mean-square error of fit (C/m^2)
}

// CalibratePreisachToData fits (Ps, Ec, Delta) for a TanhEverett Preisach model
// against experimental P-E loop data. The data must form a full bipolar loop
// (ascending then descending, or vice versa) with at least 6 points.
//
// Algorithm:
//  1. Extract Ec from data: |E| at the P=0 crossing on the ascending branch.
//  2. Extract Ps from data: maximum |P| in the dataset.
//  3. Perform golden-section search on Delta in [0.05*Ec, 2.0*Ec] to minimize
//     RMSE between a TanhEverett-driven Preisach loop and the data.
//
// Returns an error if the input has fewer than 6 points, or if Ec or Ps cannot
// be meaningfully extracted (e.g., all-positive field data).
func CalibratePreisachToData(peData []PEPoint) (*PreisachCalibrationResult, error) {
	if len(peData) < 6 {
		return nil, errors.New("need at least 6 P-E points for calibration")
	}

	// --- Extract initial estimates from data ---
	ec, err := extractEc(peData)
	if err != nil {
		return nil, err
	}
	ps := extractPs(peData)
	if ps <= 0 {
		return nil, errors.New("Ps must be positive; no polarization data found")
	}

	// --- Golden-section search on Delta ---
	// Delta controls the squareness of the loop. Narrow Delta -> square loop.
	// Search bounds: [0.05*Ec, 2.0*Ec].
	deltaLow := 0.05 * ec
	deltaHigh := 2.0 * ec

	// Golden-section minimization of RMSE(Delta).
	bestDelta, bestRMSE := goldenSectionMinimize(deltaLow, deltaHigh, func(delta float64) float64 {
		return computeLoopRMSE(peData, ps, ec, delta)
	}, 50) // 50 iterations gives ~1e-11 relative precision

	return &PreisachCalibrationResult{
		Ps:    ps,
		Ec:    ec,
		Delta: bestDelta,
		RMSE:  bestRMSE,
	}, nil
}

// extractEc finds the coercive field from a P-E loop by locating the zero-
// crossing of P on the ascending branch (where E goes from negative to positive
// and P crosses zero from negative to positive).
func extractEc(data []PEPoint) (float64, error) {
	// Find ascending branch zero-crossing: look for consecutive points where
	// P changes sign from negative to positive while E increases.
	for i := 1; i < len(data); i++ {
		prev := data[i-1]
		curr := data[i]
		if prev.Polarization_Cm <= 0 && curr.Polarization_Cm > 0 && curr.Field_Vm > prev.Field_Vm {
			// Linear interpolation for E at P=0.
			frac := -prev.Polarization_Cm / (curr.Polarization_Cm - prev.Polarization_Cm)
			ec := prev.Field_Vm + frac*(curr.Field_Vm-prev.Field_Vm)
			if ec <= 0 {
				continue // Ec should be positive on ascending branch
			}
			return ec, nil
		}
	}
	// Fallback: look for descending branch zero-crossing (P goes + to -)
	for i := 1; i < len(data); i++ {
		prev := data[i-1]
		curr := data[i]
		if prev.Polarization_Cm >= 0 && curr.Polarization_Cm < 0 && curr.Field_Vm < prev.Field_Vm {
			frac := prev.Polarization_Cm / (prev.Polarization_Cm - curr.Polarization_Cm)
			ec := math.Abs(prev.Field_Vm + frac*(curr.Field_Vm-prev.Field_Vm))
			if ec > 0 {
				return ec, nil
			}
		}
	}
	return 0, errors.New("could not extract Ec: no P=0 zero-crossing found")
}

// extractPs returns the maximum absolute polarization in the dataset.
func extractPs(data []PEPoint) float64 {
	maxP := 0.0
	for _, pt := range data {
		if ap := math.Abs(pt.Polarization_Cm); ap > maxP {
			maxP = ap
		}
	}
	return maxP
}

// computeLoopRMSE simulates a full P-E loop using TanhEverett + PreisachStack
// and returns the RMSE against the experimental data points.
//
// The simulation sweeps the field through the same sequence of E values present
// in the data (pre-conditioned by a full negative-saturation excursion).
func computeLoopRMSE(data []PEPoint, ps, ec, delta float64) float64 {
	everett := &TanhEverett{Ps: ps, Ec: ec, Delta: delta}

	// Determine the maximum field excursion from the data to set SaturationE.
	maxE := 0.0
	for _, pt := range data {
		if ae := math.Abs(pt.Field_Vm); ae > maxE {
			maxE = ae
		}
	}
	// Saturation field should exceed the data range to ensure full coverage.
	satE := maxE * 1.2
	if satE < 2*ec {
		satE = 2 * ec
	}

	stack := NewPreisachStack(satE, everett)
	if stack == nil {
		return math.Inf(1)
	}

	// Pre-condition: drive to negative saturation, then to positive saturation,
	// then back to negative saturation. This initializes the Preisach memory
	// into the negative-saturation state consistent with typical P-E loop
	// measurement protocol.
	preConditionSteps := 20
	for i := 0; i <= preConditionSteps; i++ {
		e := -satE + 2*satE*float64(i)/float64(preConditionSteps)
		stack.Update(e)
	}
	for i := 0; i <= preConditionSteps; i++ {
		e := satE - 2*satE*float64(i)/float64(preConditionSteps)
		stack.Update(e)
	}

	// Now simulate the same field sequence as the data.
	sumSqErr := 0.0
	for _, pt := range data {
		pSim := stack.Update(pt.Field_Vm)
		diff := pSim - pt.Polarization_Cm
		sumSqErr += diff * diff
	}

	return math.Sqrt(sumSqErr / float64(len(data)))
}

// goldenSectionMinimize finds the minimum of f on [a, b] using golden-section
// search. Returns (xMin, fMin). maxIter controls precision.
func goldenSectionMinimize(a, b float64, f func(float64) float64, maxIter int) (float64, float64) {
	phi := (math.Sqrt(5) - 1) / 2 // ~0.618
	resphi := 1.0 - phi           // ~0.382

	x1 := a + resphi*(b-a)
	x2 := a + phi*(b-a)
	f1 := f(x1)
	f2 := f(x2)

	for i := 0; i < maxIter; i++ {
		if f1 < f2 {
			b = x2
			x2 = x1
			f2 = f1
			x1 = a + resphi*(b-a)
			f1 = f(x1)
		} else {
			a = x1
			x1 = x2
			f1 = f2
			x2 = a + phi*(b-a)
			f2 = f(x2)
		}
	}

	if f1 < f2 {
		return x1, f1
	}
	return x2, f2
}

// LoadPELoopCSV loads a two-column CSV file (E_MV_cm, P_uC_cm2) and converts
// to SI PEPoints. The first line is treated as a header and skipped.
//
// Unit conversions applied:
//   - E: MV/cm -> V/m (multiply by 1e8)
//   - P: uC/cm^2 -> C/m^2 (multiply by 1e-2)
func LoadPELoopCSV(csvPath string) ([]PEPoint, error) {
	cleanPath := filepath.Clean(csvPath)
	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("open PE-loop CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse PE-loop CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, errors.New("PE-loop CSV must have a header row and at least one data row")
	}

	// Skip header row.
	points := make([]PEPoint, 0, len(records)-1)
	for lineNum, row := range records[1:] {
		if len(row) < 2 {
			return nil, fmt.Errorf("line %d: expected 2 columns, got %d", lineNum+2, len(row))
		}
		eVal, err := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
		if err != nil {
			return nil, fmt.Errorf("line %d col 1: %w", lineNum+2, err)
		}
		pVal, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("line %d col 2: %w", lineNum+2, err)
		}
		points = append(points, PEPoint{
			Field_Vm:        eVal * 1e8,  // MV/cm -> V/m
			Polarization_Cm: pVal * 1e-2, // uC/cm^2 -> C/m^2
		})
	}
	return points, nil
}

// SplitAscendingDescending splits a full P-E loop trace into ascending
// (E increasing) and descending (E decreasing) branches. Useful for
// branch-specific analysis and visualization.
func SplitAscendingDescending(data []PEPoint) (ascending, descending []PEPoint) {
	if len(data) < 2 {
		return data, nil
	}
	// Find the turning point (maximum E).
	maxIdx := 0
	maxE := data[0].Field_Vm
	for i, pt := range data {
		if pt.Field_Vm > maxE {
			maxE = pt.Field_Vm
			maxIdx = i
		}
	}
	// Everything up to and including max is ascending; rest is descending.
	ascending = make([]PEPoint, maxIdx+1)
	copy(ascending, data[:maxIdx+1])
	descending = make([]PEPoint, len(data)-maxIdx)
	copy(descending, data[maxIdx:])
	return
}

// SortByField sorts P-E points by electric field value for plotting.
func SortByField(data []PEPoint) {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Field_Vm < data[j].Field_Vm
	})
}
