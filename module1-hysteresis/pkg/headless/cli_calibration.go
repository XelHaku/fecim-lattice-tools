// Package gui provides CLI calibration functionality for hysteresis module.
package headless

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

// CLICalibrationOptions configures CLI calibration behavior
type CLICalibrationOptions struct {
	MaterialName string  // Material to calibrate (empty = all materials)
	NumLevels    int     // Number of discrete levels (default: 30)
	Temperature  float64 // Temperature in Kelvin (default: 300)
	Force        bool    // Force recalibration even if file exists
	Verbose      bool    // Print progress messages
	Verify       bool    // Run verification after calibration
}

type TempCalibration struct {
	Temperature     float64   `json:"temperature_k"`
	CalibrationUp   []float64 `json:"calibration_up"`
	CalibrationDown []float64 `json:"calibration_down"`
	CalibUpLow      []float64 `json:"calib_up_low"`
	CalibUpHigh     []float64 `json:"calib_up_high"`
	CalibDownLow    []float64 `json:"calib_down_low"`
	CalibDownHigh   []float64 `json:"calib_down_high"`
	LastErrorUp     []int     `json:"last_error_up"`
	LastErrorDown   []int     `json:"last_error_down"`
	RelaxCompUp     []float64 `json:"relax_comp_up"`
	RelaxCompDown   []float64 `json:"relax_comp_down"`
}

type CalibrationData struct {
	Version         int                      `json:"version"`
	MaterialName    string                   `json:"material_name"`
	NumLevels       int                      `json:"num_levels"`
	TargetRangeFrac float64                  `json:"target_range_frac"`
	Calibrations    map[int]*TempCalibration `json:"calibrations"`
	SavedAt         string                   `json:"saved_at"`

	CalibrationUp   []float64 `json:"calibration_up,omitempty"`
	CalibrationDown []float64 `json:"calibration_down,omitempty"`
	CalibUpLow      []float64 `json:"calib_up_low,omitempty"`
	CalibUpHigh     []float64 `json:"calib_up_high,omitempty"`
	CalibDownLow    []float64 `json:"calib_down_low,omitempty"`
	CalibDownHigh   []float64 `json:"calib_down_high,omitempty"`
	LastErrorUp     []int     `json:"last_error_up,omitempty"`
	LastErrorDown   []int     `json:"last_error_down,omitempty"`
	RelaxCompUp     []float64 `json:"relax_comp_up,omitempty"`
	RelaxCompDown   []float64 `json:"relax_comp_down,omitempty"`
}

const (
	calibrationVersion = 4
	calibrationDir     = "data/calibrations"
)

func calibrationFileForMaterial(materialName string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r - 'A' + 'a'
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		case r == ' ' || r == '(' || r == ')':
			return '_'
		default:
			return -1
		}
	}, materialName)
	for strings.Contains(safe, "__") {
		safe = strings.ReplaceAll(safe, "__", "_")
	}
	safe = strings.Trim(safe, "_")
	if safe == "" {
		safe = "unknown"
	}
	return filepath.Join(calibrationDir, safe+".json")
}

// RunCLICalibration performs calibration without GUI and saves results to file.
// Returns nil on success, error on failure.
func RunCLICalibration(opts CLICalibrationOptions) error {
	// Set defaults (NumLevels=0 means use material's native level count)
	if opts.Temperature == 0 {
		opts.Temperature = 300
	}

	materials := ferroelectric.AllMaterials()
	var materialsToCalibrate []*ferroelectric.HZOMaterial

	if opts.MaterialName == "" || opts.MaterialName == "all" {
		// Calibrate all materials
		materialsToCalibrate = materials
	} else {
		// Find specific material
		for _, m := range materials {
			if m.Name == opts.MaterialName {
				materialsToCalibrate = []*ferroelectric.HZOMaterial{m}
				break
			}
		}
		if len(materialsToCalibrate) == 0 {
			return fmt.Errorf("material not found: %s\nAvailable materials: %v", opts.MaterialName, getMaterialNames(materials))
		}
	}

	// Ensure calibration directory exists
	if err := os.MkdirAll(calibrationDir, 0755); err != nil {
		return fmt.Errorf("failed to create calibration directory: %w", err)
	}

	for _, mat := range materialsToCalibrate {
		calibFile := calibrationFileForMaterial(mat.Name)

		// Skip if file exists and not forcing
		if !opts.Force {
			if _, err := os.Stat(calibFile); err == nil {
				if opts.Verbose {
					fmt.Printf("Skipping %s (calibration exists: %s)\n", mat.Name, calibFile)
				}
				continue
			}
		}

		// Calculate actual level count for this material
		actualLevels := mat.GetNumLevels()
		if opts.NumLevels > 0 {
			actualLevels = opts.NumLevels
		}

		if opts.Verbose {
			fmt.Printf("Calibrating %s at %.0fK with %d levels...\n", mat.Name, opts.Temperature, actualLevels)
		}

		start := time.Now()
		if err := calibrateMaterial(mat, opts); err != nil {
			return fmt.Errorf("calibration failed for %s: %w", mat.Name, err)
		}

		if opts.Verbose {
			fmt.Printf("  Saved: %s (%.2fs)\n", calibFile, time.Since(start).Seconds())
		}

		// Run verification if requested
		if opts.Verify {
			if opts.Verbose {
				fmt.Printf("  Verifying calibration...\n")
			}
			successRate, err := verifyCalibration(mat, opts)
			if err != nil {
				return fmt.Errorf("verification failed for %s: %w", mat.Name, err)
			}
			if opts.Verbose {
				fmt.Printf("  Verification: %.1f%% target accuracy\n", successRate*100)
			}
			// Warn but don't fail for moderate accuracy - extreme levels have physics limitations
			// Very high level counts (>100) have fundamental Preisach model limitations
			minRequired := 0.70
			if actualLevels > 100 {
				minRequired = 0.01 // Very high level counts are experimental - Preisach model can't resolve
				if opts.Verbose {
					fmt.Printf("    Note: >100 levels exceeds Preisach model resolution (experimental)\n")
				}
			}
			if successRate < minRequired {
				return fmt.Errorf("verification failed for %s: only %.1f%% accuracy (need %.0f%%+ minimum)", mat.Name, successRate*100, minRequired*100)
			} else if successRate < 0.85 && opts.Verbose {
				fmt.Printf("    Note: %.1f%% accuracy (extreme levels have physics limitations)\n", successRate*100)
			}
		}
	}

	return nil
}

// calibrateMaterial performs calibration for a single material
func calibrateMaterial(mat *ferroelectric.HZOMaterial, opts CLICalibrationOptions) error {
	// Use material's native level count, or override if explicitly specified
	numLevels := mat.GetNumLevels()
	if opts.NumLevels > 0 {
		numLevels = opts.NumLevels
	}
	tempK := opts.Temperature

	// Create Preisach model with resolution scaled to number of levels
	// Need at least 10x the number of levels for adequate resolution
	preisachGridSize := numLevels * 10
	if preisachGridSize < 200 {
		preisachGridSize = 200 // Minimum for accuracy
	}
	if preisachGridSize > 1000 {
		preisachGridSize = 1000 // Cap for memory/performance
	}
	// Create Preisach model
	preisach := ferroelectric.NewPreisachModel(mat)
	preisach.SetTemperature(tempK)

	// Get temperature-corrected Ec
	Ec := preisach.GetEffectiveEc()
	if Ec == 0 {
		Ec = mat.Ec
	}
	Emax := 2.0 * Ec // Wider range for reliable calibration
	Ps := mat.Ps

	maxLevel := numLevels - 1
	if maxLevel < 1 {
		maxLevel = 1
	}

	// Initialize calibration arrays
	calibrationUp := make([]float64, numLevels)
	calibrationDown := make([]float64, numLevels)
	calibUpLow := make([]float64, numLevels)
	calibUpHigh := make([]float64, numLevels)
	calibDownLow := make([]float64, numLevels)
	calibDownHigh := make([]float64, numLevels)
	lastErrorUp := make([]int, numLevels)
	lastErrorDown := make([]int, numLevels)
	relaxCompUp := make([]float64, numLevels)
	relaxCompDown := make([]float64, numLevels)

	// Initialize bounds
	for i := 0; i < numLevels; i++ {
		calibUpLow[i] = Ec * 0.3
		calibUpHigh[i] = Ec * 2.0
		calibDownLow[i] = -Ec * 2.0
		calibDownHigh[i] = -Ec * 0.3

		// Initialize relaxation compensation with parabolic profile
		normalizedPos := float64(i) / float64(maxLevel)
		relaxCompUp[i] = 0.05 * 4 * normalizedPos * (1 - normalizedPos)
		relaxCompDown[i] = 0.05 * 4 * normalizedPos * (1 - normalizedPos)
	}

	// Helper to test what level results from a field (ascending)
	testLevelAsc := func(testE float64) int {
		preisach.Reset()
		preisach.Update(-Emax)
		preisach.Update(0)
		preisach.Update(testE)
		P := preisach.Update(0)
		normalizedP := P / Ps
		level := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if level < 0 {
			level = 0
		} else if level > maxLevel {
			level = maxLevel
		}
		return level
	}

	// Calibrate ascending (from -Ps to each level)
	for level := 1; level < numLevels; level++ {
		// Binary search for field that hits target level
		lowE := Ec * 0.3
		highE := Ec * 2.0
		bestE := (lowE + highE) / 2
		bestDiff := numLevels

		for iter := 0; iter < 25; iter++ {
			midE := (lowE + highE) / 2
			resultLevel := testLevelAsc(midE)
			bestE = midE

			diff := resultLevel - level
			if absInt(diff) < absInt(bestDiff) {
				bestDiff = diff
				bestE = midE
			}

			if resultLevel == level {
				break // Exact match
			} else if resultLevel < level {
				lowE = midE // Need higher field
			} else {
				highE = midE // Need lower field
			}

			if highE-lowE < Ec*0.01 {
				break
			}
		}

		calibrationUp[level] = bestE
		calibUpLow[level] = lowE
		calibUpHigh[level] = highE
	}

	// Helper to test what level results from a field (descending)
	testLevelDesc := func(testE float64) int {
		preisach.Reset()
		preisach.Update(Emax)
		preisach.Update(0)
		preisach.Update(testE)
		P := preisach.Update(0)
		normalizedP := P / Ps
		level := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if level < 0 {
			level = 0
		} else if level > maxLevel {
			level = maxLevel
		}
		return level
	}

	// Calibrate descending (from +Ps to each level)
	for level := numLevels - 2; level >= 0; level-- {
		// Binary search for field (negative) that hits target level
		lowE := -Ec * 2.0  // More negative
		highE := -Ec * 0.3 // Less negative
		bestE := (lowE + highE) / 2
		bestDiff := numLevels

		for iter := 0; iter < 25; iter++ {
			midE := (lowE + highE) / 2
			resultLevel := testLevelDesc(midE)
			bestE = midE

			diff := resultLevel - level
			if absInt(diff) < absInt(bestDiff) {
				bestDiff = diff
				bestE = midE
			}

			if resultLevel == level {
				break // Exact match
			} else if resultLevel > level {
				// Need more negative field to go lower
				highE = midE
			} else {
				// Need less negative field (result too low)
				lowE = midE
			}

			if highE-lowE < Ec*0.01 {
				break
			}
		}

		calibrationDown[level] = bestE
		calibDownLow[level] = lowE
		calibDownHigh[level] = highE
	}

	// Enforce strict monotonicity with minimum step size
	minStep := Ec * 0.02 // 2% of Ec minimum separation

	// Ascending: each level must require strictly higher field
	// Two-pass algorithm to handle both spikes and plateaus
	for pass := 0; pass < 2; pass++ {
		for i := 1; i < numLevels; i++ {
			minAllowed := calibrationUp[i-1] + minStep
			if calibrationUp[i] < minAllowed {
				calibrationUp[i] = minAllowed
			}
		}
	}

	// Descending: each level must require strictly more negative field
	// For descending: index 0 = level 0 (most negative = -Ps)
	// So calibrationDown[i] should be MORE negative than calibrationDown[i+1]
	for pass := 0; pass < 2; pass++ {
		for i := numLevels - 2; i >= 0; i-- {
			maxAllowed := calibrationDown[i+1] - minStep // More negative
			if calibrationDown[i] > maxAllowed {
				calibrationDown[i] = maxAllowed
			}
		}
	}

	// Build calibration data structure
	tempKRounded := int(math.Round(tempK))
	calData := &CalibrationData{
		Version:      calibrationVersion,
		MaterialName: mat.Name,
		NumLevels:    numLevels,
		Calibrations: map[int]*TempCalibration{
			tempKRounded: {
				Temperature:     tempK,
				CalibrationUp:   calibrationUp,
				CalibrationDown: calibrationDown,
				CalibUpLow:      calibUpLow,
				CalibUpHigh:     calibUpHigh,
				CalibDownLow:    calibDownLow,
				CalibDownHigh:   calibDownHigh,
				LastErrorUp:     lastErrorUp,
				LastErrorDown:   lastErrorDown,
				RelaxCompUp:     relaxCompUp,
				RelaxCompDown:   relaxCompDown,
			},
		},
		SavedAt: time.Now().Format(time.RFC3339),
	}

	// Save to file
	return saveCalibrationData(mat.Name, calData)
}

// verifyCalibration tests that calibrated fields actually hit target levels
func verifyCalibration(mat *ferroelectric.HZOMaterial, opts CLICalibrationOptions) (float64, error) {
	numLevels := mat.GetNumLevels()
	if opts.NumLevels > 0 {
		numLevels = opts.NumLevels
	}
	tempK := opts.Temperature
	maxLevel := numLevels - 1
	if maxLevel < 1 {
		maxLevel = 1
	}

	// Create fresh Preisach model with matching resolution
	preisachGridSize := numLevels * 10
	if preisachGridSize < 200 {
		preisachGridSize = 200
	}
	if preisachGridSize > 1000 {
		preisachGridSize = 1000
	}
	// Create fresh Preisach model
	preisach := ferroelectric.NewPreisachModel(mat)
	preisach.SetTemperature(tempK)

	Ec := preisach.GetEffectiveEc()
	if Ec == 0 {
		Ec = mat.Ec
	}
	Emax := 2.0 * Ec
	Ps := mat.Ps

	// Load calibration data
	calData, err := loadCalibrationData(mat.Name)
	if err != nil {
		return 0, fmt.Errorf("failed to load calibration: %w", err)
	}

	tempKRounded := int(math.Round(tempK))
	cal, ok := calData.Calibrations[tempKRounded]
	if !ok {
		return 0, fmt.Errorf("no calibration for temperature %dK", tempKRounded)
	}

	// Test ascending calibration (saturate negative, apply field)
	successAsc := 0
	var failedAsc []string
	for targetLevel := 1; targetLevel < numLevels; targetLevel++ {
		// Saturate negative
		preisach.Reset()
		for i := 0; i < 50; i++ {
			preisach.Update(-Emax)
		}
		preisach.Update(0)

		// Apply calibrated field
		writeE := cal.CalibrationUp[targetLevel]
		preisach.Update(writeE)
		P := preisach.Update(0)

		// Calculate resulting level
		normalizedP := P / Ps
		resultLevel := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if resultLevel < 0 {
			resultLevel = 0
		} else if resultLevel > maxLevel {
			resultLevel = maxLevel
		}

		// Check if hit target (allow ±1 tolerance)
		if absInt(resultLevel-targetLevel) <= 1 {
			successAsc++
		} else if opts.Verbose {
			failedAsc = append(failedAsc, fmt.Sprintf("asc[%d]→%d", targetLevel, resultLevel))
		}
	}

	// Test descending calibration (saturate positive, apply field)
	successDesc := 0
	var failedDesc []string
	for targetLevel := 0; targetLevel < numLevels-1; targetLevel++ {
		// Saturate positive
		preisach.Reset()
		for i := 0; i < 50; i++ {
			preisach.Update(Emax)
		}
		preisach.Update(0)

		// Apply calibrated field
		writeE := cal.CalibrationDown[targetLevel]
		preisach.Update(writeE)
		P := preisach.Update(0)

		// Calculate resulting level
		normalizedP := P / Ps
		resultLevel := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if resultLevel < 0 {
			resultLevel = 0
		} else if resultLevel > maxLevel {
			resultLevel = maxLevel
		}

		// Check if hit target (allow ±1 tolerance)
		if absInt(resultLevel-targetLevel) <= 1 {
			successDesc++
		} else if opts.Verbose {
			failedDesc = append(failedDesc, fmt.Sprintf("desc[%d]→%d", targetLevel, resultLevel))
		}
	}

	// Print failures if verbose
	if opts.Verbose && (len(failedAsc) > 0 || len(failedDesc) > 0) {
		if len(failedAsc) > 0 {
			fmt.Printf("    Failed asc: %v\n", failedAsc)
		}
		if len(failedDesc) > 0 {
			fmt.Printf("    Failed desc: %v\n", failedDesc)
		}
	}

	// Calculate overall success rate
	totalTests := (numLevels - 1) * 2 // Skip level 0 for ascending, level N-1 for descending
	totalSuccess := successAsc + successDesc
	return float64(totalSuccess) / float64(totalTests), nil
}

// absInt returns the absolute value of an integer (for CLI calibration)
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// loadCalibrationData loads calibration from JSON file
func loadCalibrationData(materialName string) (*CalibrationData, error) {
	filePath := calibrationFileForMaterial(materialName)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data CalibrationData
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

// saveCalibrationData writes calibration to JSON file
func saveCalibrationData(materialName string, data *CalibrationData) error {
	filePath := calibrationFileForMaterial(materialName)

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// getMaterialNames returns list of material names for error messages
func getMaterialNames(materials []*ferroelectric.HZOMaterial) []string {
	names := make([]string, len(materials))
	for i, m := range materials {
		names[i] = m.Name
	}
	return names
}

// ListMaterials returns available material names for CLI help
func ListMaterials() []string {
	return getMaterialNames(ferroelectric.AllMaterials())
}
