//go:build legacy_fyne

package gui

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TempCalibration holds calibration data for a specific temperature
type TempCalibration struct {
	Temperature     float64   `json:"temperature_k"`    // Temperature in Kelvin
	CalibrationUp   []float64 `json:"calibration_up"`   // Ascending calibration values
	CalibrationDown []float64 `json:"calibration_down"` // Descending calibration values
	CalibUpLow      []float64 `json:"calib_up_low"`     // Binary search lower bounds (ascending)
	CalibUpHigh     []float64 `json:"calib_up_high"`    // Binary search upper bounds (ascending)
	CalibDownLow    []float64 `json:"calib_down_low"`   // Binary search lower bounds (descending)
	CalibDownHigh   []float64 `json:"calib_down_high"`  // Binary search upper bounds (descending)
	LastErrorUp     []int     `json:"last_error_up"`    // Last error for oscillation detection
	LastErrorDown   []int     `json:"last_error_down"`  // Last error for oscillation detection
	RelaxCompUp     []float64 `json:"relax_comp_up"`    // Relaxation compensation factors (ascending)
	RelaxCompDown   []float64 `json:"relax_comp_down"`  // Relaxation compensation factors (descending)
}

// CalibrationData holds persistent calibration state (v2+: multi-temperature support)
type CalibrationData struct {
	Version         int                      `json:"version"`           // Schema version (4 = Everett clamp fix)
	MaterialName    string                   `json:"material_name"`     // Material these calibrations are for
	NumLevels       int                      `json:"num_levels"`        // Number of discrete levels
	TargetRangeFrac float64                  `json:"target_range_frac"` // Effective Ps range used for level mapping
	Calibrations    map[int]*TempCalibration `json:"calibrations"`      // Key: temperature in Kelvin (rounded)
	SavedAt         string                   `json:"saved_at"`          // Timestamp

	// Legacy v1 fields (for migration, not used in v2)
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

const calibrationVersion = 4

// Key temperatures for automotive range calibration (Kelvin)
var keyTemperatures = []float64{
	233, // -40C (automotive cold)
	273, // 0C
	300, // 27C (room temp, default)
	373, // 100C
	423, // 150C (automotive hot)
}

// temperatureTolerance is the max distance from a cached calibration to interpolate
const temperatureTolerance = 25.0 // Kelvin
const calibrationDir = "data/calibrations"
const lkHistorySampleInterval = 5e-4 // 0.5ms sim-time between history points (L-K)

// maxWrdRetries is the maximum number of WRD retry attempts before accepting current level
// This prevents infinite loops when targeting difficult levels (e.g., level 26)
// Reduced from 25 to 15 since boundary oscillation tolerance kicks in at retry 10
const maxWrdRetries = 15

// calibrationFileForMaterial returns the calibration file path for a given material and engine.
// Material names are sanitized to be filesystem-safe.
func calibrationFileForMaterial(materialName string, engine PhysicsEngine) string {
	// Sanitize material name for use as filename
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r - 'A' + 'a' // lowercase
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		case r == ' ' || r == '(' || r == ')':
			return '_'
		default:
			return -1 // drop
		}
	}, materialName)

	// Remove consecutive underscores and trim
	for strings.Contains(safe, "__") {
		safe = strings.ReplaceAll(safe, "__", "_")
	}
	safe = strings.Trim(safe, "_")

	if safe == "" {
		safe = "unknown"
	}

	filename := safe
	if engine == PhysicsLandau {
		filename = safe + "-lk"
	}
	return filepath.Join(calibrationDir, filename+".json")
}

// saveCalibration persists calibration data to disk (v2: multi-temperature)
func (a *App) saveCalibration() error {
	if a.material == nil || !a.calibrated {
		return nil
	}
	// Tests and CI sometimes need to run without mutating tracked calibration JSON.
	if os.Getenv("FECIM_DISABLE_CALIBRATION_SAVE") == "1" {
		return nil
	}

	// Build calibrations map from active calibration and cache
	calibrations := make(map[int]*TempCalibration)

	// Copy existing cached calibrations
	for tempK, cal := range a.tempCalibrations {
		calibrations[tempK] = cal
	}

	// Save current active calibration at its temperature
	tempK := int(math.Round(a.calibrationTemp))

	// MONOTONICITY ENFORCEMENT before saving: fix any runtime corruption
	// This ensures saved calibration values are always monotonic.
	Ec := a.effectiveEc()
	numLevels := len(a.calibrationUp)
	step := Ec * 0.02 // 2% of Ec minimum step

	// Fix ascending: ensure calibrationUp[i] < calibrationUp[i+1]
	for i := 1; i < numLevels; i++ {
		if a.calibrationUp[i] <= a.calibrationUp[i-1] {
			a.calibrationUp[i] = a.calibrationUp[i-1] + step
		}
	}

	// Fix descending: ensure calibrationDown[i] > calibrationDown[i-1] (less negative)
	for i := numLevels - 2; i >= 0; i-- {
		if a.calibrationDown[i] >= a.calibrationDown[i+1] {
			a.calibrationDown[i] = a.calibrationDown[i+1] - step
		}
	}

	calibrations[tempK] = &TempCalibration{
		Temperature:     a.calibrationTemp,
		CalibrationUp:   append([]float64(nil), a.calibrationUp...),
		CalibrationDown: append([]float64(nil), a.calibrationDown...),
		CalibUpLow:      append([]float64(nil), a.calibUpLow...),
		CalibUpHigh:     append([]float64(nil), a.calibUpHigh...),
		CalibDownLow:    append([]float64(nil), a.calibDownLow...),
		CalibDownHigh:   append([]float64(nil), a.calibDownHigh...),
		LastErrorUp:     append([]int(nil), a.lastErrorUp...),
		LastErrorDown:   append([]int(nil), a.lastErrorDown...),
		RelaxCompUp:     append([]float64(nil), a.relaxCompUp...),
		RelaxCompDown:   append([]float64(nil), a.relaxCompDown...),
	}

	data := CalibrationData{
		Version:         calibrationVersion,
		MaterialName:    a.material.Name,
		NumLevels:       a.numLevels,
		TargetRangeFrac: a.wrdRangeFrac,
		Calibrations:    calibrations,
		SavedAt:         time.Now().Format(time.RFC3339),
	}

	// Get per-material calibration file path
	calibFile := calibrationFileForMaterial(a.material.Name, a.physicsEngine)

	// Ensure calibration directory exists
	dir := filepath.Dir(calibFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create calibration dir: %w", err)
	}

	// Write JSON file
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal calibration: %w", err)
	}

	if err := os.WriteFile(calibFile, jsonData, 0644); err != nil {
		return fmt.Errorf("write calibration file: %w", err)
	}

	log.Printf("Calibration saved: %s (%d levels, %d temperatures)", calibFile, a.numLevels, len(calibrations))
	return nil
}

// loadCalibration loads calibration data from disk if valid for current material
// Returns true if calibration was loaded successfully
// Supports both v1 (single-temp) and v2 (multi-temp) formats
// Each material has its own calibration file in data/calibrations/
func (a *App) loadCalibration() bool {
	if a.material == nil {
		return false
	}

	// Get per-material calibration file path
	calibFile := calibrationFileForMaterial(a.material.Name, a.physicsEngine)

	jsonData, err := os.ReadFile(calibFile)
	if err != nil {
		// File doesn't exist yet - that's fine, will calibrate fresh
		log.Printf("No calibration file for %s, will calibrate", a.material.Name)
		return false
	}

	var data CalibrationData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Printf("Invalid calibration file %s, will recalibrate: %v", calibFile, err)
		return false
	}

	// Warn if material name doesn't match (shouldn't happen with per-material files)
	if data.MaterialName != a.material.Name {
		log.Printf("Warning: calibration file material mismatch (file=%s, expected=%s)", data.MaterialName, a.material.Name)
	}
	if data.NumLevels != a.numLevels {
		log.Printf("Calibration levels mismatch (got %d, want %d), will recalibrate", data.NumLevels, a.numLevels)
		return false
	}
	if data.TargetRangeFrac <= 0 {
		log.Printf("Calibration missing target_range_frac, will recalibrate")
		return false
	}
	if math.Abs(data.TargetRangeFrac-a.wrdRangeFrac) > 1e-6 {
		log.Printf("Calibration target range mismatch (file=%.3f, current=%.3f), will recalibrate", data.TargetRangeFrac, a.wrdRangeFrac)
		return false
	}

	// Handle version migration
	if data.Version == 1 {
		// v1 format: single-temp calibration at 300K (room temp)
		log.Printf("Migrating v1 calibration file to v2 format (treating as 300K)")

		// Validate v1 array lengths
		if len(data.CalibrationUp) != a.numLevels || len(data.CalibrationDown) != a.numLevels {
			log.Printf("Calibration array size mismatch, will recalibrate")
			return false
		}

		// Create TempCalibration from v1 data
		tempCal := &TempCalibration{
			Temperature:     300,
			CalibrationUp:   data.CalibrationUp,
			CalibrationDown: data.CalibrationDown,
			CalibUpLow:      data.CalibUpLow,
			CalibUpHigh:     data.CalibUpHigh,
			CalibDownLow:    data.CalibDownLow,
			CalibDownHigh:   data.CalibDownHigh,
			LastErrorUp:     data.LastErrorUp,
			LastErrorDown:   data.LastErrorDown,
		}

		// Initialize auxiliary arrays if needed (backward compat with older v1 files)
		a.initializeTempCalibrationBounds(tempCal)

		// Store in cache
		a.tempCalibrations = make(map[int]*TempCalibration)
		a.tempCalibrations[300] = tempCal
		a.calibrationTemp = 300

		// Load into active arrays
		a.loadTempCalibration(tempCal)

		log.Printf("Calibration migrated from v1: %s (%d levels at 300K)", calibFile, data.NumLevels)
		return true
	}

	if data.Version != calibrationVersion {
		log.Printf("Calibration version mismatch (got %d, want %d), will recalibrate", data.Version, calibrationVersion)
		return false
	}

	// v2+ format: multi-temperature calibration
	if data.Calibrations == nil || len(data.Calibrations) == 0 {
		log.Printf("Empty calibrations map in v2 file, will recalibrate")
		return false
	}

	// Load all temperature calibrations into cache
	a.tempCalibrations = make(map[int]*TempCalibration)
	for tempK, cal := range data.Calibrations {
		if cal == nil {
			continue
		}
		// Validate array lengths for each calibration
		if len(cal.CalibrationUp) != a.numLevels || len(cal.CalibrationDown) != a.numLevels {
			log.Printf("Calibration array size mismatch at %dK, skipping", tempK)
			continue
		}
		// Initialize bounds if needed
		a.initializeTempCalibrationBounds(cal)
		a.tempCalibrations[tempK] = cal
	}

	if len(a.tempCalibrations) == 0 {
		log.Printf("No valid calibrations in file, will recalibrate")
		return false
	}

	// Load calibration for current temperature (interpolate or use nearest)
	currentTemp := a.currentTemperature()
	a.loadCalibrationForTemperature(currentTemp)

	log.Printf("Calibration loaded: %s (%d levels, %d temperatures)", calibFile, a.numLevels, len(a.tempCalibrations))
	return true
}

// initializeTempCalibrationBounds initializes auxiliary arrays if they're nil
func (a *App) initializeTempCalibrationBounds(cal *TempCalibration) {
	if len(cal.CalibUpLow) != a.numLevels {
		ec := a.material.Ec
		emax := ec * 1.5
		cal.CalibUpLow = make([]float64, a.numLevels)
		cal.CalibUpHigh = make([]float64, a.numLevels)
		cal.CalibDownLow = make([]float64, a.numLevels)
		cal.CalibDownHigh = make([]float64, a.numLevels)
		for i := 0; i < a.numLevels; i++ {
			cal.CalibUpLow[i] = ec * 0.5
			cal.CalibUpHigh[i] = emax
			cal.CalibDownLow[i] = -emax
			cal.CalibDownHigh[i] = -ec * 0.5
		}
	}
	if len(cal.LastErrorUp) != a.numLevels {
		cal.LastErrorUp = make([]int, a.numLevels)
		cal.LastErrorDown = make([]int, a.numLevels)
	}
	if len(cal.RelaxCompUp) != a.numLevels || len(cal.RelaxCompDown) != a.numLevels {
		cal.RelaxCompUp = make([]float64, a.numLevels)
		cal.RelaxCompDown = make([]float64, a.numLevels)
		maxLevel := a.numLevels - 1
		if maxLevel < 1 {
			maxLevel = 1 // Prevent division by zero when numLevels=1
		}
		for i := 0; i < a.numLevels; i++ {
			// Initialize with parabolic profile (peak 5% at middle, zero at edges)
			normalizedPos := float64(i) / float64(maxLevel)
			cal.RelaxCompUp[i] = 0.05 * 4 * normalizedPos * (1 - normalizedPos)
			cal.RelaxCompDown[i] = 0.05 * 4 * normalizedPos * (1 - normalizedPos)
		}
	}
}

// loadTempCalibration loads a TempCalibration into the active calibration arrays
func (a *App) loadTempCalibration(cal *TempCalibration) {
	a.calibrationUp = append([]float64(nil), cal.CalibrationUp...)
	a.calibrationDown = append([]float64(nil), cal.CalibrationDown...)
	a.calibUpLow = append([]float64(nil), cal.CalibUpLow...)
	a.calibUpHigh = append([]float64(nil), cal.CalibUpHigh...)
	a.calibDownLow = append([]float64(nil), cal.CalibDownLow...)
	a.calibDownHigh = append([]float64(nil), cal.CalibDownHigh...)
	a.lastErrorUp = append([]int(nil), cal.LastErrorUp...)
	a.lastErrorDown = append([]int(nil), cal.LastErrorDown...)
	a.relaxCompUp = append([]float64(nil), cal.RelaxCompUp...)
	a.relaxCompDown = append([]float64(nil), cal.RelaxCompDown...)
	a.calibrationTemp = cal.Temperature

	// Refactoring: Sync to CalibrationManager
	if a.calibManager != nil {
		copy(a.calibManager.CalibrationUp, a.calibrationUp)
		copy(a.calibManager.CalibrationDown, a.calibrationDown)
	}

	// MONOTONICITY ENFORCEMENT on load: fix any corrupted values from file
	Ec := a.effectiveEc()
	numLevels := len(a.calibrationUp)
	step := Ec * 0.02 // 2% of Ec minimum step

	// Fix ascending: ensure calibrationUp[i] < calibrationUp[i+1]
	for i := 1; i < numLevels; i++ {
		if a.calibrationUp[i] <= a.calibrationUp[i-1] {
			a.calibrationUp[i] = a.calibrationUp[i-1] + step
		}
	}

	// Fix descending: ensure calibrationDown[i] > calibrationDown[i-1] (less negative)
	for i := numLevels - 2; i >= 0; i-- {
		if a.calibrationDown[i] >= a.calibrationDown[i+1] {
			a.calibrationDown[i] = a.calibrationDown[i+1] - step
		}
	}

	// Sync to CalibrationManager (Refactoring)
	if a.calibManager != nil {
		a.calibManager.CalibrationUp = append([]float64(nil), a.calibrationUp...)
		a.calibManager.CalibrationDown = append([]float64(nil), a.calibrationDown...)
		a.calibManager.CalibUpLow = append([]float64(nil), a.calibUpLow...)
		a.calibManager.CalibUpHigh = append([]float64(nil), a.calibUpHigh...)
		a.calibManager.CalibDownLow = append([]float64(nil), a.calibDownLow...)
		a.calibManager.CalibDownHigh = append([]float64(nil), a.calibDownHigh...)
		a.calibManager.LastErrorUp = append([]int(nil), a.lastErrorUp...)
		a.calibManager.LastErrorDown = append([]int(nil), a.lastErrorDown...)
		a.calibManager.RelaxCompUp = append([]float64(nil), a.relaxCompUp...)
		a.calibManager.RelaxCompDown = append([]float64(nil), a.relaxCompDown...)
	}

	a.calibrated = true

	// Validate calibration quality
	a.validateCalibration()

	// Log critical calibration quality issues
	upDupes := countDuplicates(a.calibrationUp)
	downDupes := countDuplicates(a.calibrationDown)
	if upDupes > 10 || downDupes > 10 {
		log.Printf("CRITICAL: Calibration has %d/%d duplicate E-fields.", upDupes, downDupes)
		log.Printf("  Consider: increasing grid size or widening distribution (σ)")
	}
}

// validateCalibration checks for degenerate calibration values (duplicate E-fields)
// and logs warnings. This helps diagnose level consistency issues.
func (a *App) validateCalibration() {
	// Check ascending calibration for duplicates
	upDupes := countDuplicates(a.calibrationUp)
	downDupes := countDuplicates(a.calibrationDown)

	totalLevels := len(a.calibrationUp)
	if upDupes > 0 || downDupes > 0 {
		log.Printf("WARNING: Calibration quality issue - %d/%d ascending and %d/%d descending levels share E-field values",
			upDupes, totalLevels, downDupes, totalLevels)
		log.Printf("  This indicates hysteresis staircase regions where multiple levels map to same E-field")
		log.Printf("  Write accuracy may be reduced. Consider recalibrating with different parameters.")
	}
}

// countDuplicates returns how many values in the slice share the same value as another
func countDuplicates(vals []float64) int {
	if len(vals) == 0 {
		return 0
	}
	// Count values that appear more than once
	seen := make(map[float64]int)
	tolerance := 1e-10 // Values within this tolerance are considered equal
	for _, v := range vals {
		// Round to tolerance for comparison
		rounded := math.Round(v/tolerance) * tolerance
		seen[rounded]++
	}

	dupeCount := 0
	for _, count := range seen {
		if count > 1 {
			dupeCount += count // All instances of duplicated values
		}
	}
	return dupeCount
}

// loadCalibrationForTemperature loads or interpolates calibration for the given temperature
// MUST be called with a.mu held
func (a *App) loadCalibrationForTemperature(tempK float64) {
	// Check if we have an exact match (within 1K)
	tempKRounded := int(math.Round(tempK))
	if cal, ok := a.tempCalibrations[tempKRounded]; ok {
		a.loadTempCalibration(cal)
		log.Printf("Loaded exact calibration for %dK", tempKRounded)
		return
	}

	// Find nearest calibrations for interpolation
	lowerTemp, upperTemp, lowerCal, upperCal := a.findNearestCalibrations(tempK)

	if lowerCal == nil && upperCal == nil {
		// No calibrations available - should not happen if loaded correctly
		log.Printf("No calibrations available for interpolation at %.0fK", tempK)
		return
	}

	// If only one calibration available, use it directly
	if lowerCal == nil {
		a.loadTempCalibration(upperCal)
		log.Printf("Using nearest calibration from %.0fK for %.0fK", upperTemp, tempK)
		return
	}
	if upperCal == nil {
		a.loadTempCalibration(lowerCal)
		log.Printf("Using nearest calibration from %.0fK for %.0fK", lowerTemp, tempK)
		return
	}

	// Interpolate between two calibrations
	a.interpolateCalibrations(tempK, lowerTemp, upperTemp, lowerCal, upperCal)
	log.Printf("Interpolated calibration for %.0fK from %.0fK and %.0fK", tempK, lowerTemp, upperTemp)
}

// findNearestCalibrations finds the two calibrations nearest to the given temperature
// Returns (lowerTemp, upperTemp, lowerCal, upperCal) where lower < tempK < upper
func (a *App) findNearestCalibrations(tempK float64) (float64, float64, *TempCalibration, *TempCalibration) {
	var lowerTemp, upperTemp float64 = -1e9, 1e9
	var lowerCal, upperCal *TempCalibration

	for t, cal := range a.tempCalibrations {
		temp := float64(t)
		if temp <= tempK && temp > lowerTemp {
			lowerTemp = temp
			lowerCal = cal
		}
		if temp >= tempK && temp < upperTemp {
			upperTemp = temp
			upperCal = cal
		}
	}

	// Handle edge cases
	if lowerTemp < 0 {
		lowerTemp = 0
		lowerCal = nil
	}
	if upperTemp > 1e6 {
		upperTemp = 0
		upperCal = nil
	}

	return lowerTemp, upperTemp, lowerCal, upperCal
}

// interpolateCalibrations linearly interpolates between two calibrations
// MUST be called with a.mu held
func (a *App) interpolateCalibrations(tempK, lowerTemp, upperTemp float64, lowerCal, upperCal *TempCalibration) {
	// Calculate interpolation factor (0 = use lowerCal, 1 = use upperCal)
	t := (tempK - lowerTemp) / (upperTemp - lowerTemp)

	// Initialize arrays if needed
	if len(a.calibrationUp) != a.numLevels {
		a.calibrationUp = make([]float64, a.numLevels)
		a.calibrationDown = make([]float64, a.numLevels)
		a.calibUpLow = make([]float64, a.numLevels)
		a.calibUpHigh = make([]float64, a.numLevels)
		a.calibDownLow = make([]float64, a.numLevels)
		a.calibDownHigh = make([]float64, a.numLevels)
		a.lastErrorUp = make([]int, a.numLevels)
		a.lastErrorDown = make([]int, a.numLevels)
		a.relaxCompUp = make([]float64, a.numLevels)
		a.relaxCompDown = make([]float64, a.numLevels)
	}

	// Interpolate field values for each level
	for i := 0; i < a.numLevels; i++ {
		a.calibrationUp[i] = lowerCal.CalibrationUp[i]*(1-t) + upperCal.CalibrationUp[i]*t
		a.calibrationDown[i] = lowerCal.CalibrationDown[i]*(1-t) + upperCal.CalibrationDown[i]*t

		// For bounds, use the more conservative (wider) bounds
		a.calibUpLow[i] = math.Min(lowerCal.CalibUpLow[i], upperCal.CalibUpLow[i])
		a.calibUpHigh[i] = math.Max(lowerCal.CalibUpHigh[i], upperCal.CalibUpHigh[i])
		a.calibDownLow[i] = math.Min(lowerCal.CalibDownLow[i], upperCal.CalibDownLow[i])
		a.calibDownHigh[i] = math.Max(lowerCal.CalibDownHigh[i], upperCal.CalibDownHigh[i])

		// Interpolate relaxation compensation factors (with bounds check for backward compatibility)
		if i < len(lowerCal.RelaxCompUp) && i < len(upperCal.RelaxCompUp) {
			a.relaxCompUp[i] = lowerCal.RelaxCompUp[i]*(1-t) + upperCal.RelaxCompUp[i]*t
		} else {
			// Static compensation removed - rely on adaptive system
			// No physics justification for static overshoot in quasistatic regime
			a.relaxCompUp[i] = 0.0
		}
		if i < len(lowerCal.RelaxCompDown) && i < len(upperCal.RelaxCompDown) {
			a.relaxCompDown[i] = lowerCal.RelaxCompDown[i]*(1-t) + upperCal.RelaxCompDown[i]*t
		} else {
			// Static compensation removed - rely on adaptive system
			a.relaxCompDown[i] = 0.0
		}

		// Reset error tracking for interpolated calibration
		a.lastErrorUp[i] = 0
		a.lastErrorDown[i] = 0
	}

	a.calibrationTemp = tempK
	a.calibrated = true
}

// hasCalibrationNear checks if we have a calibration within tolerance of the given temperature
func (a *App) hasCalibrationNear(tempK float64) bool {
	for t := range a.tempCalibrations {
		if math.Abs(float64(t)-tempK) <= temperatureTolerance {
			return true
		}
	}
	return false
}

// onTemperatureChanged handles temperature changes and triggers recalibration if needed
// MUST be called with a.mu held
func (a *App) onTemperatureChanged(newTemp float64) {
	// Update Preisach model temperature
	if a.preisach != nil {
		a.preisach.SetTemperature(newTemp)
	}
	if a.lkSolver != nil {
		a.lkSolver.Temperature = newTemp
		// Recalculate Alpha from updated temperature via Curie-Weiss + electrostriction.
		// Without this, temperature slider changes have no effect on LK dynamics.
		if !a.lkSolver.UseMaterialAlpha {
			a.lkSolver.UpdateParams()
		}
	}

	// Check if we need new calibration or can use existing/interpolated
	if a.hasCalibrationNear(newTemp) {
		// Use existing or interpolate from cached calibrations
		a.loadCalibrationForTemperature(newTemp)
	} else {
		// Need to calibrate for this temperature
		log.Printf("Temperature %.0fK exceeds tolerance from cached calibrations, recalibrating...", newTemp)
		a.calibrateLevelsAtTemperature(newTemp)

		// Save updated calibration
		if err := a.saveCalibration(); err != nil {
			log.Printf("Warning: failed to save calibration: %v", err)
		}
	}
}
