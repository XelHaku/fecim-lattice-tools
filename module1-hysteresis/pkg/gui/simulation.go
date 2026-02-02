package gui

import (
	"encoding/json"
	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
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
	Version      int                      `json:"version"`       // Schema version (4 = Everett clamp fix)
	MaterialName string                   `json:"material_name"` // Material these calibrations are for
	NumLevels    int                      `json:"num_levels"`    // Number of discrete levels
	Calibrations map[int]*TempCalibration `json:"calibrations"`  // Key: temperature in Kelvin (rounded)
	SavedAt      string                   `json:"saved_at"`      // Timestamp

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

// maxWrdRetries is the maximum number of WRD retry attempts before accepting current level
// This prevents infinite loops when targeting difficult levels (e.g., level 26)
// Reduced from 25 to 15 since boundary oscillation tolerance kicks in at retry 10
const maxWrdRetries = 15

// calibrationFileForMaterial returns the calibration file path for a given material.
// Material names are sanitized to be filesystem-safe.
func calibrationFileForMaterial(materialName string) string {
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

	return filepath.Join(calibrationDir, safe+".json")
}

// saveCalibration persists calibration data to disk (v2: multi-temperature)
func (a *App) saveCalibration() error {
	if a.material == nil || !a.calibrated {
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
	Ec := a.material.Ec
	if a.preisach != nil {
		Ec = a.preisach.GetEffectiveEc()
	}
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
		Version:      calibrationVersion,
		MaterialName: a.material.Name,
		NumLevels:    a.numLevels,
		Calibrations: calibrations,
		SavedAt:      time.Now().Format(time.RFC3339),
	}

	// Get per-material calibration file path
	calibFile := calibrationFileForMaterial(a.material.Name)

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
	calibFile := calibrationFileForMaterial(a.material.Name)

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
	currentTemp := a.preisach.Temperature
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
	Ec := a.material.Ec
	if a.preisach != nil {
		Ec = a.preisach.GetEffectiveEc()
	}
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
	a.preisach.SetTemperature(newTemp)
	if a.lkSolver != nil {
		a.lkSolver.Temperature = newTemp
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

// simulationLoop runs the main simulation loop at ~60 FPS
// simulationLoop runs the main simulation loop at ~60 FPS with adaptive physics stepping
func (a *App) simulationLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS targeting
	defer ticker.Stop()

	lastTime := time.Now()

	// Adaptive Time-Stepping Constants
	const (
		dtMax     = 0.025 // 25ms cap to prevent explosion after pause
		dtNominal = 1e-4  // 0.1ms nominal physics step (good for standard loop)
		dtMin     = 1e-6  // 1µs minimum step near critical field (Ec)
	)

	for a.running {
		<-ticker.C

		now := time.Now()
		frameDt := now.Sub(lastTime).Seconds()
		lastTime = now

		if a.paused {
			continue
		}

		if frameDt > dtMax {
			frameDt = dtMax
		}

		// Lock ONCE for the entire frame's physics burst
		a.mu.Lock()

		// --- Adaptive Sub-Stepping Loop ---
		remainingDt := frameDt

		// Safety break to prevent infinite loops if calculation is too slow
		const maxSubSteps = 1000
		subSteps := 0

		matEc := 0.0
		if a.material != nil {
			matEc = a.material.Ec
		}

		for remainingDt > 0 && subSteps < maxSubSteps {
			// Determine step size based on physics state
			// If E-field is near Ec, use smaller steps to capture switching dynamics

			currentStep := dtNominal

			// Check proximity to Ec (switching region)
			// Switching happens at +Ec (increasing) and -Ec (decreasing)
			// But effective Ec varies. Use material Ec as baseline proxy.
			if matEc > 0 {
				distPlus := math.Abs(a.electricField - matEc)
				distMinus := math.Abs(a.electricField + matEc)
				minDist := math.Min(distPlus, distMinus)

				// User requirement: If |E - Ec| < 0.1 MV/cm: dt = dt_min
				// 0.1 MV/cm = 0.1e6 V/cm = 1e5 V/m (Units in material are V/m? Wait.
				// Ec is ~1 MV/cm = 1e8 V/m. 0.1 MV/cm = 1e7 V/m.
				// User said: "0.1 MV/cm". 1 MV/cm = 10^6 V/cm = 10^8 V/m.
				// So 0.1 MV/cm = 10^7 V/m.
				// Let's use 10 MV/m (1e7) as threshold.

				threshold := 1e7 // 0.1 MV/cm
				if minDist < threshold {
					currentStep = dtMin
				}
			}

			// Don't step past the frame time
			if currentStep > remainingDt {
				currentStep = remainingDt
			}

			a.updatePhysics(currentStep)
			remainingDt -= currentStep
			subSteps++
		}

		// Update UI once per frame with the final state
		a.updateUI()

		a.mu.Unlock()
	}
}

// updatePhysics handles the physics and state transitions.
// MUST be called with a.mu held.
func (a *App) updatePhysics(dt float64) {
	// a.mu.Lock() -> Removed, caller holds lock

	mat := a.material
	if mat == nil {
		return
	}

	Emax := mat.Ec * 2.5 // Scale for full loop traversal
	Ec := mat.Ec
	a.simTime += dt
	phaseDuration := 1.0 / a.frequency
	rampRate := 4.0 * Emax * a.frequency

	// Generate E-field based on waveform
	if a.waveform == WaveformManual {
		if a.manualAnimating {
			a.manualPhaseTime += dt
			targetLevel := a.manualTargetLevel // 1-indexed (1-N)
			startLevel := a.manualStartLevel   // Captured at animation start
			maxLevelIdx := a.numLevels - 1

			switch a.manualPhase {
			case 0: // RESET phase
				var resetE float64
				if targetLevel > startLevel {
					resetE = -2.0 * Ec
				} else {
					resetE = 2.0 * Ec
				}

				// Ramp to prep field
				diff := resetE - a.electricField
				step := rampRate * dt
				if math.Abs(diff) < step {
					a.electricField = resetE
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				// Transition when field reached and held briefly
				if a.manualPhaseTime > phaseDuration*0.3 && math.Abs(a.electricField-resetE) < 0.01*Emax {
					a.manualPhase = 1
					a.manualPhaseTime = 0
				}

			case 1: // HOLD_RESET - return to zero
				step := rampRate * dt
				if math.Abs(a.electricField) < step {
					a.electricField = 0
				} else if a.electricField > 0 {
					a.electricField -= step
				} else {
					a.electricField += step
				}

				if a.manualPhaseTime > phaseDuration*0.2 && math.Abs(a.electricField) < 0.01*Emax {
					a.manualPhase = 2
					a.manualPhaseTime = 0
				}

			case 2: // WRITE - apply calibrated field
				var writeE float64
				targetIdx := targetLevel - 1
				midLevel := a.numLevels / 2
				goingUp := targetLevel > midLevel

				if targetIdx < 0 || targetIdx >= len(a.calibrationUp) {
					// Out of bounds - use fallback
					ratio := float64(targetLevel-1) / float64(maxLevelIdx)
					if goingUp {
						writeE = Ec * (0.9 + ratio*1.1)
					} else {
						writeE = -Ec * (0.9 + ratio*1.1)
					}
				} else if goingUp {
					writeE = a.calibrationUp[targetIdx]
					if writeE == 0 {
						ratio := float64(targetLevel-1) / float64(maxLevelIdx)
						writeE = Ec * (0.9 + ratio*1.1)
					}
				} else {
					writeE = a.calibrationDown[targetIdx]
					if writeE == 0 {
						ratio := float64(a.numLevels-targetLevel) / float64(maxLevelIdx)
						writeE = -Ec * (0.9 + ratio*1.1)
					}
				}

				// Ramp to write field
				diff := writeE - a.electricField
				step := rampRate * dt
				if math.Abs(diff) < step {
					a.electricField = writeE
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				if a.manualPhaseTime > phaseDuration*0.4 && math.Abs(a.electricField-writeE) < 0.01*Emax {
					a.manualPhase = 3
					a.manualPhaseTime = 0
				}

			case 3: // HOLD_WRITE - return to zero, polarization persists
				step := rampRate * dt
				if math.Abs(a.electricField) < step {
					a.electricField = 0
				} else if a.electricField > 0 {
					a.electricField -= step
				} else {
					a.electricField += step
				}

				// Animation complete
				if a.manualPhaseTime > phaseDuration*0.3 && math.Abs(a.electricField) < 0.01*Emax {
					finalLevel := a.discreteLevel + 1
					levelError := finalLevel - targetLevel
					adjIdx := targetLevel - 1
					Ec := mat.Ec

					// Log animation result with detailed state
					log.Printf("MANUAL ANIMATION COMPLETE: target=%d, final=%d, error=%d",
						targetLevel, finalLevel, levelError)

					if levelError != 0 && a.calibrated && adjIdx >= 0 && adjIdx < len(a.calibrationUp) {
						if targetLevel > startLevel { // startLevel defined outside case
							// ASCENDING calibration adjustment
							if a.calibManager != nil {
								a.calibManager.UpdateCalibrationUp(adjIdx, levelError, Ec)
								a.calibrationUp[adjIdx] = a.calibManager.CalibrationUp[adjIdx]
							}
						} else {
							// DESCENDING calibration adjustment
							if a.calibManager != nil {
								a.calibManager.UpdateCalibrationDown(adjIdx, levelError, Ec)
								a.calibrationDown[adjIdx] = a.calibManager.CalibrationDown[adjIdx]
							}
						}
					}

				}
			}
		}
	} else if a.autoMode {
		Emax := mat.Ec * 2 // Use local copy for thread safety
		// Wrap phase to prevent floating-point precision loss over long times
		phase := math.Mod(2*math.Pi*a.frequency*a.simTime, 2*math.Pi)

		switch a.waveform {
		case WaveformSine:
			a.electricField = Emax * math.Sin(phase)
		case WaveformTriangle:
			p := phase / (2 * math.Pi)
			if p < 0.25 {
				a.electricField = Emax * (4 * p)
			} else if p < 0.75 {
				a.electricField = Emax * (2 - 4*p)
			} else {
				a.electricField = Emax * (4*p - 4)
			}
		case WaveformWriteReadDemo:
			// Write/read demo with directional pre-bias + ISPP:
			// - Do NOT saturate every cycle.
			// - Apply ±Ec in direction of next target, return to 0.
			// - Full saturation is only used on overshoot recovery inside WriteController.
			//
			// Phase mapping:
			// 0 = PREP (apply ±Ec bias toward target)
			// 1 = HOLD_PREP (return to zero)
			// 2 = WRITE (apply calibrated/ISPP field toward target)
			// 3 = HOLD_WRITE (return to zero, polarization persists)
			// 4 = READ (small sense pulse below Ec)
			// 5 = DISPLAY (show result, pick next target)

			a.wrdPhaseTimer += dt
			phaseDuration := 1.0 / a.frequency
			rampRate := 3.0 * Emax * a.frequency
			Ec := mat.Ec // Use local copy for thread safety
			midLevel := a.numLevels / 2

			targetLevel := a.wrdTargetLevel // 1-indexed
			// Note: startLevel (a.wrdStartLevel) no longer used for direction - we use absolute position

			switch a.wrdPhase {
			case 0: // PREP - apply ±Ec in direction of target (no full saturation)
				currentLevel := a.discreteLevel + 1
				var prepE float64
				if targetLevel > currentLevel {
					prepE = Ec
				} else if targetLevel < currentLevel {
					prepE = -Ec
				} else {
					prepE = 0
				}
				a.wrdPrepE = prepE // Store for logging

				// Ramp to reset field
				diff := prepE - a.electricField
				step := rampRate * dt
				if math.Abs(diff) < step {
					a.electricField = prepE
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				// Transition when field reached and held briefly
				if a.wrdPhaseTimer > phaseDuration*0.25 && math.Abs(a.electricField-prepE) < 0.01*Emax {
					// Capture end-of-PREP state for logging
					a.wrdResetEndP = a.polarization * 100 // Convert to µC/cm²
					a.wrdResetEndLvl = a.discreteLevel + 1
					log.Printf("WRD PHASE 0→1: PREP done | E=%.3f MV/cm | P=%.2f→%.2f µC/cm² | L=%d→%d | target=%d",
						a.electricField/1e8, a.wrdResetStartP, a.wrdResetEndP, a.wrdStartLevel, a.wrdResetEndLvl, targetLevel)
					a.wrdPhase = 1
					a.wrdPhaseTimer = 0
				}

			case 1: // HOLD_RESET - return to zero (now at known remanent state)
				step := rampRate * dt
				if math.Abs(a.electricField) < step {
					a.electricField = 0
				} else if a.electricField > 0 {
					a.electricField -= step
				} else {
					a.electricField += step
				}

				// Now at known remanent state (level 1 or level N)
				if a.wrdPhaseTimer > phaseDuration*0.15 && math.Abs(a.electricField) < 0.01*Emax {
					// Capture start-of-WRITE state for logging
					a.wrdWriteStartP = a.polarization * 100 // Convert to µC/cm²
					log.Printf("WRD PHASE 1→2: SETTLE done | E=%.3f MV/cm | P=%.2f µC/cm² | L=%d | ready to write target=%d",
						a.electricField/1e8, a.wrdWriteStartP, a.discreteLevel+1, targetLevel)

					// Initialize WriteController (Refactoring)
					// Determine if fromSaturation: Level 1 or MaxLevel
					currentLevel := a.discreteLevel + 1
					fromSaturation := currentLevel <= 2 || currentLevel >= a.numLevels-1

					// SYNC TIMING: Pulse duration should be ~40% of phase duration to allow ramp-up
					a.writeController.PulseDuration = phaseDuration * 0.4
					a.writeController.Start(targetLevel, fromSaturation)

					a.wrdPhase = 2
					a.wrdPhaseTimer = 0
				}

			case 2: // WRITE - Delegated to WriteController
				// Refactored to use pkg/controller/WriteController
				// This consolidates Write, Wait, Verify, and Retry logic

				currentLevel := a.discreteLevel + 1
				targetField, done := a.writeController.Update(dt, a.electricField, currentLevel)

				// Autonomous runtime recalibration trigger (overshoots or too many pulses)
				if a.autoRecalibrate && !a.recalibratePending {
					if a.writeController.OvershootCount >= a.recalibrateOvershootMax {
						a.recalibratePending = true
						a.recalibrateReason = fmt.Sprintf("overshoots=%d target=%d", a.writeController.OvershootCount, targetLevel)
						log.Printf("AUTO RECAL scheduled: %s", a.recalibrateReason)
					} else if a.writeController.PulseCount >= a.recalibratePulseMax {
						a.recalibratePending = true
						a.recalibrateReason = fmt.Sprintf("pulses=%d target=%d", a.writeController.PulseCount, targetLevel)
						log.Printf("AUTO RECAL scheduled: %s", a.recalibrateReason)
					}
				}

				// Apply Voltage Ramp (Controller outputs target, Simulation handles physics ramp)
				step := rampRate * 1.5 * dt // Faster ramp for write pulses
				diff := targetField - a.electricField
				if math.Abs(diff) < step {
					a.electricField = targetField
				} else if diff > 0 {
					a.electricField += step
				} else {
					a.electricField -= step
				}

				if done {
					switch a.writeController.State {
					case controller.StateSuccess:
						// SUCCESS: Proceed to DISPLAY phase
						a.wrdPhase = 5
						a.wrdPhaseTimer = 0

						// Logging and Metrics
						a.wrdTotalWrites++
						a.wrdSuccessWrites++
						successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100

						log.Printf("WRD SUCCESS via Controller: target=%d, tries=%d", targetLevel, a.writeController.RetryCount)

						// Log result (replaces Case 4 success log)
						log.Printf("WRD PHASE 4→5: TARGET HIT | L_read=%d L_target=%d | rate=%.1f%% (%d/%d)",
							a.writeController.LastVerifyLevel, targetLevel,
							successRate, a.wrdSuccessWrites, a.wrdTotalWrites)

						// CRITICAL FIX: Learn from the Servo using CalibrationManager
						// Instead of naïve 100% overwrite, we use binary search + monotonicity
						learnedE := a.writeController.CurrentField
						Ec := mat.Ec
						targetIdx := targetLevel - 1 // 0-indexed for CM

						if a.calibManager != nil {
							if targetLevel > midLevel {
								// Ascending (written from reset negative)
								a.calibManager.UpdateCalibrationUp(targetIdx, 0, Ec)
								// Override the midpoint with the actual successful servo voltage to anchor the search
								a.calibManager.CalibrationUp[targetIdx] = learnedE
								a.calibrationUp[targetIdx] = learnedE
							} else {
								// Descending (written from reset positive)
								a.calibManager.UpdateCalibrationDown(targetIdx, 0, Ec)
								a.calibManager.CalibrationDown[targetIdx] = learnedE
								a.calibrationDown[targetIdx] = learnedE
							}
							log.Printf("CALIB LEARN: target=%d learnedE=%.3f×Ec (updated via CM)", targetLevel, learnedE/Ec)
						}

					case controller.StateForceReset:
						// RETRY LIMIT: Trigger Full Reset
						log.Printf("WRD RETRY LIMIT: Full RESET to saturation (retry %d)", a.writeController.RetryCount)
						a.wrdPhase = 0 // RESET phase
						a.wrdPhaseTimer = 0

					case controller.StateFailed:
						// Generic failure (shouldn't happen with infinite retries enabled, but just in case)
						log.Printf("WRD FAILED: Controller gave up on target %d", targetLevel)
						a.wrdPhase = 5 // Display anyway
						a.wrdPhaseTimer = 0
						a.wrdTotalWrites++ // Count as attempt
					}
				}

				// Cases 3 (HOLD) and 4 (VERIFY) are now internal to the Controller
				// We stay in Case 2 until 'done' is true.

			case 3, 4:
				// Legacy states - should not be reached with new controller
				// Jump to 2 just in case
				a.wrdPhase = 2
			case 5: // DISPLAY phase - return to zero, show result
				step := rampRate * 0.4 * dt
				if math.Abs(a.electricField) < step {
					a.electricField = 0
				} else if a.electricField > 0 {
					a.electricField -= step
				} else {
					a.electricField += step
				}
				// Transition to next cycle
				if a.wrdPhaseTimer > phaseDuration*0.6 {
					// Runtime auto recalibration (run between targets to avoid state disruption)
					if a.autoRecalibrate && a.recalibratePending {
						log.Printf("AUTO RECAL start: %s", a.recalibrateReason)
						a.calibrateLevelsAtTemperature(a.preisach.Temperature)
						if err := a.saveCalibration(); err != nil {
							log.Printf("Warning: failed to save calibration: %v", err)
						}
						a.needsCalibration = false
						a.recalibratePending = false
						a.recalibrateReason = ""
					}

					// Record start level for next cycle
					a.wrdStartLevel = a.discreteLevel + 1

					// Add comparison callout every 5 cycles
					if a.wrdTotalWrites > 0 && a.wrdTotalWrites%5 == 0 {
						fecimEnergy := a.wrdTotalEnergyfJ / 1000 // pJ
						// NOTE: 10M× is Dr. Tour's unverified claim. Peer-reviewed: 25-100× (Samsung Nature 2025)
						nandEquiv := fecimEnergy * 50   // 25-100× better (conservative: use 50)
						dramEquiv := fecimEnergy * 1000 // 1000× worse
						bitsStored := float64(a.wrdTotalWrites) * 4.91
						a.addLogEntry("━━ ENERGY COMPARISON ━━")
						a.addLogEntry(fmt.Sprintf("FeCIM: %.0f pJ total", fecimEnergy))
						a.addLogEntry(fmt.Sprintf("NAND:  %.0f pJ (50×!)", nandEquiv))
						a.addLogEntry(fmt.Sprintf("DRAM:  %.0f pJ (1000×)", dramEquiv))
						a.addLogEntry(fmt.Sprintf("Bits stored: %.0f (%.1f×binary)", bitsStored, 4.91))
						a.addLogEntry("━━━━━━━━━━━━━━━━━━━━━━")
					}

					// Milestone celebrations
					switch a.wrdTotalWrites {
					case 10:
						a.addLogEntry("★★ 10 ops! ~49 bits stored ★★")
					case 25:
						a.addLogEntry("★★★ 25 ops! ~123 bits stored ★★★")
					case 50:
						a.addLogEntry("★★★★ 50 ops! ~245 bits stored ★★★★")
						a.addLogEntry("Binary would need 245 cells!")
						a.addLogEntry("FeCIM: only 50 cells! (5× denser)")
					case 100:
						a.addLogEntry("★★★★★ 100 OPERATIONS! ★★★★★")
						a.addLogEntry("~491 bits in 100 FeCIM cells")
						a.addLogEntry("Binary: 491 cells needed!")
						successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100
						a.addLogEntry(fmt.Sprintf("Accuracy: %.0f%%", successRate))
					}

					// Pick new target - alternate between high and low
					midLvl := a.numLevels / 2
					rangeSize := a.numLevels / 3
					if rangeSize < 2 {
						rangeSize = 2
					}
					if a.wrdTargetLevel > midLvl {
						// Low range: 2 to rangeSize+1 (avoid extremes)
						a.wrdTargetLevel = rand.Intn(rangeSize) + 2
						if a.wrdTargetLevel > a.numLevels-1 {
							a.wrdTargetLevel = a.numLevels - 1
						}
					} else {
						// High range: (numLevels - rangeSize) to numLevels-1
						a.wrdTargetLevel = a.numLevels - rangeSize + rand.Intn(rangeSize)
						if a.wrdTargetLevel < 2 {
							a.wrdTargetLevel = 2
						}
					}
					// Capture start state for next cycle logging
					a.wrdWriteStartP = a.polarization * 100 // Convert to µC/cm²
					log.Printf("WRD CYCLE START: cycle=%d | startLevel=%d | newTarget=%d | P=%.2f µC/cm²",
						a.wrdTotalWrites+1, a.discreteLevel+1, a.wrdTargetLevel, a.wrdWriteStartP)
					// NOTE: Don't clear history - let the trail accumulate to show full hysteresis loop
					// Spike detection in plot widget handles any discontinuities
					// Go to RESET phase to ensure clean state and proper initialization
					a.wrdPhase = 0
					a.wrdPhaseTimer = 0
					a.wrdCycleEnergy = 0  // Reset energy accumulator for next cycle
					a.isppTotalPulses = 0 // Reset ISPP pulse counter for next target

					// Reset ISPP widget to idle state for new target
					if a.isppWidget != nil {
						// UI update must be on main thread
						fyne.Do(func() {
							a.isppWidget.SetAnimationState(0, 0, a.discreteLevel+1, 0, false)
						})
					}
				}

			case 6: // Legacy BOOST phase - redirect to Controller
				// The generic WriteController handles retries/boost internally or via re-entry to Case 2
				a.wrdPhase = 2
			}
		case WaveformTimeResolved:
			if !a.timeResAnimating {
				a.addLogEntry("NOTE: Time-Resolved Switching")
				a.addLogEntry("requires legacy physics engine.")
				a.timeResAnimating = true // To prevent repeated logging
			}
		}
	}

	// Update physics
	prevP := a.polarization
	// Use standard Update (Dynamic update removed in clean-up)
	a.polarization = a.preisach.Update(a.electricField)
	a.normalizedP = a.preisach.NormalizedPolarization()
	maxLevel := a.numLevels - 1
	a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * float64(maxLevel)))
	if a.discreteLevel < 0 {
		a.discreteLevel = 0
	}
	if a.discreteLevel > maxLevel {
		a.discreteLevel = maxLevel
	}

	// Calculate energy: integral of E·dP ≈ |E| * |ΔP|
	// During write/read cycles, accumulate energy for the cycle (phases 0-4)
	if a.waveform == WaveformWriteReadDemo && a.wrdPhase >= 0 && a.wrdPhase <= 4 {
		deltaP := a.polarization - prevP
		// Energy per unit volume: E·dP in J/m³
		// Use actual cell dimensions from material (use local copy for thread safety)
		cellVolume := mat.Area * mat.Thickness
		// Fallback if material doesn't have dimensions
		if cellVolume <= 0 {
			cellVolume = 2e-22 // Default: 100nm x 100nm x 20nm
		}
		energyJ := math.Abs(a.electricField*deltaP) * cellVolume
		energyfJ := energyJ * 1e15 // Convert J to fJ
		a.wrdCycleEnergy += energyfJ
	}

	// Record history (skip RESET and BOOST phases in WRD mode to avoid visual spikes)
	// WRD phases: 0=RESET, 1=HOLD_RESET, 2=WRITE, 3=HOLD_WRITE, 4=READ, 5=DISPLAY, 6=BOOST
	skipHistory := a.waveform == WaveformWriteReadDemo && (a.wrdPhase <= 1 || a.wrdPhase == 6)
	if !skipHistory {
		a.eHistory = append(a.eHistory, a.electricField)
		a.pHistory = append(a.pHistory, a.polarization)
		if len(a.eHistory) > a.maxHistory {
			a.eHistory = a.eHistory[1:]
			a.pHistory = a.pHistory[1:]
		}
	}
}

// updateUI prepares data and calls refreshGUI.
// MUST be called with a.mu held.
func (a *App) updateUI() {
	// Update UI (must be on main thread)
	fE := a.electricField
	pV := a.polarization
	dL := a.discreteLevel
	materialEc := a.material.Ec
	eHist := make([]float64, len(a.eHistory))
	pHist := make([]float64, len(a.pHistory))
	copy(eHist, a.eHistory)
	copy(pHist, a.pHistory)

	// Release lock temporarily if needed?
	// No, refreshGUI uses fyne.Do which schedules on main thread.
	// The copy operations above are safe under lock.
	// We invoke refreshGUI which takes VALUES (copies).

	a.refreshGUI(fE, pV, dL, materialEc, eHist, pHist)
}

// refreshGUI updates all UI elements with the latest simulation data.
func (a *App) refreshGUI(fE float64, pV float64, dL int, eC float64, hE []float64, hP []float64) {
	fyne.Do(func() {
		// Update labels
		a.eFieldLabel.SetText(fmt.Sprintf("E-field: %.3f MV/cm", fE/1e8))
		a.pLabel.SetText(fmt.Sprintf("%.2f µC/cm²", pV*100))
		a.mu.RLock()
		numLevels := a.numLevels
		a.mu.RUnlock()
		a.levelLabel.SetText(fmt.Sprintf("%d/%d", dL+1, numLevels))

		// Update state descriptor (divide into thirds)
		var stateText string
		lowThird := numLevels / 3
		highThird := numLevels * 2 / 3
		if dL < lowThird {
			stateText = "Negative P"
		} else if dL >= highThird {
			stateText = "Positive P"
		} else {
			stateText = "Intermediate"
		}
		if a.stateLabel != nil {
			a.stateLabel.SetText(stateText)
		}

		// Update stability indicator (M12)
		if a.stabilityIndicator != nil {
			a.stabilityIndicator.SetLevel(dL+1, numLevels)
		}

		// Update wake-up/fatigue labels (Dr. Tour recommendation)
		// Update wake-up/fatigue labels (Dr. Tour recommendation)
		cycles, degradation, wakeup := 0, 0.0, 1.0 // Placeholder for clean-up
		if a.cyclesLabel != nil {
			if cycles >= 1000000 {
				a.cyclesLabel.SetText(fmt.Sprintf("%.1fM", float64(cycles)/1e6))
			} else if cycles >= 1000 {
				a.cyclesLabel.SetText(fmt.Sprintf("%.1fK", float64(cycles)/1e3))
			} else {
				a.cyclesLabel.SetText(fmt.Sprintf("%d", cycles))
			}
		}
		if a.wakeupLabel != nil {
			a.wakeupLabel.SetText(fmt.Sprintf("%.1f%%", wakeup*100))
		}
		if a.fatigueLabel != nil {
			a.fatigueLabel.SetText(fmt.Sprintf("%.4f%%", degradation*100))
		}
		// L01: Update cycle phase indicator based on wakeup and degradation
		if a.cyclePhaseLabel != nil {
			var phase string
			if wakeup < 0.95 {
				phase = "WAKE-UP"
			} else if degradation < 0.0001 { // < 0.01% degradation
				phase = "STABLE"
			} else {
				phase = "FATIGUE"
			}
			a.cyclePhaseLabel.SetText(phase)
		}

		// Update temperature-dependent metrics (must hold lock during preisach access)
		a.mu.RLock()
		effEc := a.preisach.GetEffectiveEc()
		// Use material's nominal Pr for plot delimiters (not GetEffectivePr which recalculates from current state)
		effPr := a.material.Pr
		switchedFraction := (a.normalizedP + 1) / 2

		// Calculate squareness (Pr/Ps ratio)
		squareness := 0.0
		if a.material != nil && a.material.Ps > 0 {
			squareness = effPr / a.material.Ps
		}
		a.mu.RUnlock()

		if a.effEcLabel != nil {
			// Show Ec with ±15% uncertainty (typical device-to-device variation)
			ecVal := effEc / 1e8
			a.effEcLabel.SetText(fmt.Sprintf("Ec(T): %.2f±%.2f MV/cm", ecVal, ecVal*0.15))
		}
		if a.effPrLabel != nil {
			// Show Pr with ±20% uncertainty (typical device-to-device variation)
			prVal := effPr * 100
			a.effPrLabel.SetText(fmt.Sprintf("Pr(T): %.1f±%.1f µC/cm²", prVal, prVal*0.20))
		}
		if a.squarenessLabel != nil {
			a.squarenessLabel.SetText(fmt.Sprintf("Squareness: %.2f", squareness))
		}
		if a.switchedLabel != nil {
			a.switchedLabel.SetText(fmt.Sprintf("Switched: %.0f%%", switchedFraction*100))
		}

		// Update phase indicator based on current mode and phase
		a.mu.RLock()
		waveform := a.waveform
		wrdPhaseVal := a.wrdPhase
		manPhaseVal := a.manualPhase
		animating := a.manualAnimating
		a.mu.RUnlock()

		if a.phaseIndicator != nil {
			switch waveform {
			case WaveformWriteReadDemo:
				a.phaseIndicator.SetPhase(wrdPhaseVal, "wrd")
			case WaveformManual:
				if animating {
					a.phaseIndicator.SetPhase(manPhaseVal, "manual")
				} else {
					a.phaseIndicator.SetPhase(-1, "") // Idle
				}
			default:
				a.phaseIndicator.SetPhase(-1, "") // Idle for sine/triangle
			}
		}

		// Update slider to match current E-field (only if not being manually controlled)
		// During Manual animation, the slider reflects the animated E-field
		// Normalize by Ec for display (-1.5 to +1.5 range)
		a.mu.RLock()
		shouldUpdateSlider := a.waveform != WaveformManual || a.manualAnimating
		a.mu.RUnlock()
		if shouldUpdateSlider {
			a.eFieldSlider.SetValue(fE / eC)
		}

		// Update status and logging
		if a.paused {
			a.statusLabel.SetText("⏸ Paused")
		} else {
			a.mu.RLock()
			currentWaveform := a.waveform
			wrdPhase := a.wrdPhase
			wrdTarget := a.wrdTargetLevel
			wrdRead := a.wrdReadLevel
			lastPhase := a.lastLogPhase
			wrdTotalWrites := a.wrdTotalWrites
			wrdSuccessWrites := a.wrdSuccessWrites
			wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
			midLevel := a.numLevels / 2 // Dynamic middle level for direction logic
			a.mu.RUnlock()

			switch currentWaveform {
			case WaveformWriteReadDemo:
				var phaseStr string
				// Log phase transitions (6 phases: RESET, HOLD_RESET, WRITE, HOLD_WRITE, READ, DISPLAY)
				if wrdPhase != lastPhase {
					a.mu.Lock()
					a.lastLogPhase = wrdPhase
					switch wrdPhase {
					case 0:
						// RESET: Saturate in opposite direction
						direction := "-sat"
						if wrdTarget <= midLevel {
							direction = "+sat"
						}
						a.addLogEntry(fmt.Sprintf("◆◆ RESET   | %s | prep", direction))
					case 1:
						// HOLD_RESET: Return to zero (known state)
						a.addLogEntry("░░ SETTLE  | E=0 | prep done")
					case 2:
						// WRITE: Apply calibrated field to reach target
						direction := "+"
						if wrdTarget <= midLevel {
							direction = "-"
						}
						a.addLogEntry(fmt.Sprintf("▓▓ WRITE L%d | %sE>Ec | ~10fJ", wrdTarget, direction))
					case 3:
						// HOLD_WRITE: Return to zero, polarization persists
						a.addLogEntry(fmt.Sprintf("░░ HOLD L%d | E=0 | 0 fJ!", dL+1))
					case 4:
						// READ: Non-destructive sense
						a.addLogEntry("▒▒ READ    | E<Ec | ~1fJ")
					case 5:
						// DISPLAY: Show result
						status := "✓ MATCH"
						if wrdRead != wrdTarget {
							diff := abs(wrdRead - wrdTarget)
							if diff == 1 {
								status = fmt.Sprintf("△ ±1 (got %d)", wrdRead)
							} else {
								status = fmt.Sprintf("✗ miss (got %d)", wrdRead)
							}
						}
						successRate := 0.0
						if wrdTotalWrites > 0 {
							successRate = float64(wrdSuccessWrites) / float64(wrdTotalWrites) * 100
						}
						a.addLogEntry(fmt.Sprintf("●● L%d %s [%.0f%% rate]", wrdTarget, status, successRate))
					}
					a.mu.Unlock()
				}

				// Enhanced status with energy metrics (using local copies from RLock above)
				energyTotal := wrdTotalEnergyfJ
				writeCount := wrdTotalWrites

				switch wrdPhase {
				case 0:
					direction := "-sat"
					if wrdTarget <= midLevel {
						direction = "+sat"
					}
					phaseStr = fmt.Sprintf("◆ RESET | %s | preparing", direction)
				case 1:
					phaseStr = "░ SETTLE | E=0 | at known state"
				case 2:
					direction := "+"
					if wrdTarget <= midLevel {
						direction = "-"
					}
					phaseStr = fmt.Sprintf("▓ WRITE L%d | %sE>Ec | ~10fJ", wrdTarget, direction)
				case 3:
					phaseStr = fmt.Sprintf("░ HOLD L%d | E=0 | ZERO POWER", dL+1)
				case 4:
					phaseStr = fmt.Sprintf("▒ READ | Sense L%d | ~1fJ", dL+1)
				case 5:
					successRate := 0.0
					if writeCount > 0 {
						successRate = float64(wrdSuccessWrites) / float64(writeCount) * 100
					}
					if wrdRead == wrdTarget {
						phaseStr = fmt.Sprintf("● L%d ✓ | Ops:%d | %.0f%% | %.0fpJ", wrdRead, writeCount, successRate, energyTotal/1000)
					} else {
						phaseStr = fmt.Sprintf("● L%d (want %d) | Ops:%d | %.0f%%", wrdRead, wrdTarget, writeCount, successRate)
					}
				}
				a.statusLabel.SetText(fmt.Sprintf("⚡ FeCIM Write/Read | %s", phaseStr))
			case WaveformManual:
				// Manual mode status with RESET-AND-RETRY physics
				a.mu.RLock()
				animAnimating := a.manualAnimating
				manPhase := a.manualPhase
				manTarget := a.manualTargetLevel
				manStart := a.manualStartLevel
				a.mu.RUnlock()

				if animAnimating {
					var phaseStr string
					switch manPhase {
					case 0:
						// RESET phase
						if manTarget > manStart {
							phaseStr = "RESET -sat..."
						} else {
							phaseStr = "RESET +sat..."
						}
					case 1:
						phaseStr = "SETTLE E=0..."
					case 2:
						phaseStr = fmt.Sprintf("WRITE → L%d...", manTarget)
					case 3:
						phaseStr = fmt.Sprintf("HOLD L%d...", dL+1)
					default:
						phaseStr = fmt.Sprintf("Current: L%d", dL+1)
					}
					a.statusLabel.SetText(fmt.Sprintf("TARGET L%d | %s", manTarget, phaseStr))
				} else {
					a.statusLabel.SetText(fmt.Sprintf("Manual L%d | Click level bar", dL+1))
				}
			case WaveformTimeResolved:
				a.mu.RLock()
				animating := a.timeResAnimating
				idx := a.timeResIndex
				dataLen := len(a.timeResDataTimes)
				a.mu.RUnlock()

				if animating && dataLen > 0 && idx < dataLen {
					a.mu.RLock()
					currentTime := a.timeResDataTimes[idx]
					switchedCount := a.timeResDataSwitch[idx]
					totalHysterons := len(a.timeResDataSwitch)
					a.mu.RUnlock()

					switchedFrac := float64(switchedCount) / float64(totalHysterons) * 100
					a.statusLabel.SetText(fmt.Sprintf("⚡ Time-Resolved | t=%.1f ns | %.0f%% switched | L%d",
						currentTime*1e9, switchedFrac, dL+1))
				} else {
					a.statusLabel.SetText("⚡ Time-Resolved Switching (KAI Dynamics)")
				}
			default:
				frac := (a.normalizedP + 1) / 2 * 100
				a.statusLabel.SetText(fmt.Sprintf("● Running | t=%.2fs | Switched: %.1f%%", a.simTime, frac))
			}
		}

		// Slide panel removed - was distracting and flickering

		// Update log text
		a.mu.RLock()
		logText := a.getLogText()
		a.mu.RUnlock()
		a.logText.SetText(logText)

		// Update plot
		a.plot.SetData(hE, hP, fE, pV)
		a.plot.Refresh()

		// Update level indicator (level is 0-indexed, display is 1-indexed)
		a.levelIndicator.SetLevel(dL)

		// Highlight target level during animations
		a.mu.RLock()
		currentMode := a.waveform
		currentWrdPh := a.wrdPhase
		currentWrdTrg := a.wrdTargetLevel
		manAnim := a.manualAnimating
		manTrg := a.manualTargetLevel
		matEcVal := a.material.Ec // For settled threshold
		a.mu.RUnlock()

		if currentMode == WaveformWriteReadDemo {
			// Show target until point SETTLES at target level (not just crosses it)
			// Settled = level matches target AND E-field is near zero
			atTarget := (dL + 1) == currentWrdTrg                     // level is 0-indexed, target is 1-indexed
			eFieldSettled := math.Abs(fE) < 0.01*matEcVal             // E-field near zero (1% of Ec)
			settled := atTarget && eFieldSettled && currentWrdPh >= 3 // Must be past WRITE phase

			// Keep highlight on during active phases OR until settled at target
			highlight := currentWrdPh >= 0 && currentWrdPh <= 5 && !settled
			a.levelIndicator.SetTargetLevel(currentWrdTrg, highlight)
		} else if currentMode == WaveformManual && manAnim {
			// Show target until point SETTLES at target level in Manual mode
			atTarget := (dL + 1) == manTrg // level is 0-indexed, target is 1-indexed
			eFieldSettled := math.Abs(fE) < 0.01*matEcVal
			settled := atTarget && eFieldSettled

			// Keep highlight on until settled at target
			a.levelIndicator.SetTargetLevel(manTrg, !settled)
		} else {
			// Clear target highlight
			a.levelIndicator.SetTargetLevel(0, false)
		}

		a.levelIndicator.Refresh()

		// Update cell visualizer
		a.cellViz.SetLevel(dL)
		a.cellViz.Refresh()
	})
}

// calibrateLevelsAtTemperature performs calibration at a specific temperature.
// Sets Preisach temperature before calibrating, stores result in cache.
// MUST be called with a.mu held.
func (a *App) calibrateLevelsAtTemperature(tempK float64) {
	if a.preisach == nil || a.material == nil {
		return
	}

	// Set Preisach temperature before calibrating
	a.preisach.SetTemperature(tempK)
	a.calibrationTemp = tempK

	// Perform calibration
	a.calibrateLevels()

	// Store in cache
	tempKRounded := int(math.Round(tempK))
	if a.tempCalibrations == nil {
		a.tempCalibrations = make(map[int]*TempCalibration)
	}
	a.tempCalibrations[tempKRounded] = &TempCalibration{
		Temperature:     tempK,
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

	log.Printf("Level calibration complete for material: %s at %.0fK", a.material.Name, tempK)
}

// calibrateLevels performs a calibration sweep to map field→level relationship.
// This mimics how real ferroelectric memory controllers characterize each device
// and build lookup tables for programming. Called at startup and when material changes.
// MUST be called with a.mu held.
//
// FIXED: Uses binary search with fresh saturation for each level to avoid
// Preisach state corruption from the previous sweep-based approach.
func (a *App) calibrateLevels() {
	if a.preisach == nil || a.material == nil {
		return
	}

	// Use temperature-corrected Ec from Preisach model
	Ec := a.preisach.GetEffectiveEc()
	if Ec == 0 {
		// Fallback to material Ec if temperature correction returns 0
		Ec = a.material.Ec
	}
	Emax := 2.0 * Ec // Go well beyond saturation for reliable calibration
	numLevels := a.numLevels
	maxLevel := numLevels - 1
	if maxLevel < 1 {
		maxLevel = 1 // Prevent division by zero when numLevels=1
	}

	// Record calibration temperature
	a.calibrationTemp = a.preisach.Temperature

	// Resize calibration arrays if needed (check all arrays to handle partial initialization)
	if len(a.calibrationUp) != numLevels {
		a.calibrationUp = make([]float64, numLevels)
	}
	if len(a.calibrationDown) != numLevels {
		a.calibrationDown = make([]float64, numLevels)
	}
	if len(a.calibUpLow) != numLevels {
		a.calibUpLow = make([]float64, numLevels)
	}
	if len(a.calibUpHigh) != numLevels {
		a.calibUpHigh = make([]float64, numLevels)
	}
	if len(a.calibDownLow) != numLevels {
		a.calibDownLow = make([]float64, numLevels)
	}
	if len(a.calibDownHigh) != numLevels {
		a.calibDownHigh = make([]float64, numLevels)
	}
	if len(a.lastErrorUp) != numLevels {
		a.lastErrorUp = make([]int, numLevels)
	}
	if len(a.lastErrorDown) != numLevels {
		a.lastErrorDown = make([]int, numLevels)
	}
	if len(a.relaxCompUp) != numLevels {
		a.relaxCompUp = make([]float64, numLevels)
	}
	if len(a.relaxCompDown) != numLevels {
		a.relaxCompDown = make([]float64, numLevels)
	}

	// Initialize bounds based on temperature-corrected Ec
	for i := 0; i < numLevels; i++ {
		// Initial bounds: full range (will be narrowed by runtime feedback)
		a.calibUpLow[i] = Ec * 0.3
		a.calibUpHigh[i] = Ec * 2.0
		a.calibDownLow[i] = -Ec * 2.0
		a.calibDownHigh[i] = -Ec * 0.3

		// No static relaxation compensation - not needed in quasistatic Preisach simulation.
		// Switching time (10ns) is negligible vs pulse duration (300ms), so no drift occurs.
		// The adaptive runtime feedback system handles any corrections needed.
		a.relaxCompUp[i] = 0.0
		a.relaxCompDown[i] = 0.0
	}

	// Normalize using Ps to match runtime discrete-level mapping
	effPs := a.material.Ps
	if effPs <= 0 {
		effPs = a.material.Pr
	}
	if effPs <= 0 {
		effPs = 1.0
	}

	// Helper function to test what level results from a given field
	// starting from negative saturation (for ascending calibration)
	testLevelAscending := func(testE float64) int {
		// Reset and saturate negative
		a.preisach.Reset()
		for i := 0; i < 50; i++ {
			a.preisach.Update(-Emax)
		}
		a.preisach.Update(0) // At level 1 (negative remanent)

		// Apply test field and return to zero
		a.preisach.Update(testE)
		p := a.preisach.Update(0)

		// Normalize by Ps to match runtime discrete-level mapping
		normalizedP := p / effPs
		if normalizedP > 1.0 {
			normalizedP = 1.0
		}
		if normalizedP < -1.0 {
			normalizedP = -1.0
		}
		level := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if level < 0 {
			level = 0
		}
		if level > maxLevel {
			level = maxLevel
		}
		return level
	}

	// Helper function for descending calibration (from positive saturation)
	testLevelDescending := func(testE float64) int {
		// Reset and saturate positive
		a.preisach.Reset()
		for i := 0; i < 50; i++ {
			a.preisach.Update(Emax)
		}
		a.preisach.Update(0) // At level N (positive remanent)

		// Apply test field (negative) and return to zero
		a.preisach.Update(testE)
		p := a.preisach.Update(0)

		// Normalize by Ps to match runtime discrete-level mapping
		normalizedP := p / effPs
		if normalizedP > 1.0 {
			normalizedP = 1.0
		}
		if normalizedP < -1.0 {
			normalizedP = -1.0
		}
		level := int(math.Round((normalizedP + 1) / 2 * float64(maxLevel)))
		if level < 0 {
			level = 0
		}
		if level > maxLevel {
			level = maxLevel
		}
		return level
	}

	// Calibrate ASCENDING using binary search for each level
	// Start with initial estimates based on linear interpolation
	for targetLevel := 1; targetLevel < numLevels; targetLevel++ {
		// Initial estimate: linear interpolation between Ec and 2*Ec
		ratio := float64(targetLevel) / float64(maxLevel)
		initialGuess := Ec * (0.8 + ratio*1.2) // Range: 0.8*Ec to 2.0*Ec

		// Binary search to find exact field - use full bounds range
		lowE := Ec * 0.3 // Match initialized bounds for full coverage
		highE := Emax
		bestE := initialGuess
		bestDiff := numLevels // Start with worst case

		// Binary search with 15 iterations (precision: ~0.003% of range)
		for iter := 0; iter < 15; iter++ {
			midE := (lowE + highE) / 2
			resultLevel := testLevelAscending(midE)

			diff := resultLevel - targetLevel
			if abs(diff) < abs(bestDiff) {
				bestDiff = diff
				bestE = midE
			}

			if resultLevel == targetLevel {
				// Found exact match
				break
			} else if resultLevel < targetLevel {
				// Need higher field
				lowE = midE
			} else {
				// Need lower field
				highE = midE
			}
		}

		a.calibrationUp[targetLevel] = bestE
	}
	// Level 0 is at negative saturation, no field needed from that state
	a.calibrationUp[0] = 0

	// Calibrate DESCENDING using binary search for each level
	for targetLevel := maxLevel - 1; targetLevel >= 0; targetLevel-- {
		// Initial estimate: linear interpolation between -Ec and -2*Ec
		ratio := float64(maxLevel-targetLevel) / float64(maxLevel)
		initialGuess := -Ec * (0.8 + ratio*1.2) // Range: -0.8*Ec to -2.0*Ec

		// Binary search to find exact field (negative values) - use full bounds range
		lowE := -Emax      // More negative
		highE := -Ec * 0.3 // Less negative - match initialized bounds
		bestE := initialGuess
		bestDiff := numLevels

		for iter := 0; iter < 15; iter++ {
			midE := (lowE + highE) / 2
			resultLevel := testLevelDescending(midE)

			diff := resultLevel - targetLevel
			if abs(diff) < abs(bestDiff) {
				bestDiff = diff
				bestE = midE
			}

			if resultLevel == targetLevel {
				break
			} else if resultLevel > targetLevel {
				// Need more negative field (lower E)
				highE = midE
			} else {
				// Need less negative field (higher E)
				lowE = midE
			}
		}

		a.calibrationDown[targetLevel] = bestE
	}
	// Level maxLevel is at positive saturation, no field needed from that state
	a.calibrationDown[maxLevel] = 0

	// MONOTONICITY ENFORCEMENT: Fix any non-monotonic values (spikes)
	// This is critical for preventing large errors when Preisach resolution
	// causes binary search to converge to non-monotonic values.
	//
	// For ascending (calibrationUp): E-field must increase with level
	// For descending (calibrationDown): E-field must decrease (more negative) with decreasing level

	// Fix ascending calibration: ensure calibrationUp[i] <= calibrationUp[i+1]
	for i := 1; i < numLevels-1; i++ {
		if a.calibrationUp[i+1] < a.calibrationUp[i] {
			// Spike detected - interpolate from neighbors
			// Find next valid (higher) value
			nextValid := a.calibrationUp[i]
			for j := i + 1; j < numLevels; j++ {
				if a.calibrationUp[j] > a.calibrationUp[i] {
					nextValid = a.calibrationUp[j]
					break
				}
			}
			// Interpolate
			a.calibrationUp[i+1] = (a.calibrationUp[i] + nextValid) / 2
		}
	}

	// Second pass: ensure strict monotonicity from bottom up
	for i := 1; i < numLevels; i++ {
		if a.calibrationUp[i] <= a.calibrationUp[i-1] {
			// Add small increment to maintain monotonicity
			step := Ec * 0.02 // 2% of Ec per level minimum step
			a.calibrationUp[i] = a.calibrationUp[i-1] + step
		}
	}

	// Fix descending calibration: ensure calibrationDown[i] >= calibrationDown[i-1] (less negative for higher levels)
	for i := numLevels - 2; i > 0; i-- {
		if a.calibrationDown[i-1] > a.calibrationDown[i] {
			// Spike detected - interpolate
			prevValid := a.calibrationDown[i]
			for j := i - 1; j >= 0; j-- {
				if a.calibrationDown[j] < a.calibrationDown[i] {
					prevValid = a.calibrationDown[j]
					break
				}
			}
			a.calibrationDown[i-1] = (a.calibrationDown[i] + prevValid) / 2
		}
	}

	// Second pass: ensure strict monotonicity from top down
	for i := numLevels - 2; i >= 0; i-- {
		if a.calibrationDown[i] >= a.calibrationDown[i+1] {
			// Add small decrement to maintain monotonicity
			step := Ec * 0.02 // 2% of Ec per level minimum step
			a.calibrationDown[i] = a.calibrationDown[i+1] - step
		}
	}

	// Reset Preisach to neutral state after calibration
	a.preisach.Reset()
	a.electricField = 0
	a.polarization = 0

	// Sync to CalibrationManager (Refactoring)
	if a.calibManager != nil {
		a.calibManager.CalibrationUp = append([]float64(nil), a.calibrationUp...)
		a.calibManager.CalibrationDown = append([]float64(nil), a.calibrationDown...)
		a.calibManager.CalibUpLow = append([]float64(nil), a.calibUpLow...)
		a.calibManager.CalibUpHigh = append([]float64(nil), a.calibUpHigh...)
		a.calibManager.CalibDownLow = append([]float64(nil), a.calibDownLow...)
		a.calibManager.CalibDownHigh = append([]float64(nil), a.calibDownHigh...)
		// LastError and RelaxComp reset during calibration usually, so copying current state (likely zeros) is fine
		a.calibManager.LastErrorUp = append([]int(nil), a.lastErrorUp...)
		a.calibManager.LastErrorDown = append([]int(nil), a.lastErrorDown...)
		a.calibManager.RelaxCompUp = append([]float64(nil), a.relaxCompUp...)
		a.calibManager.RelaxCompDown = append([]float64(nil), a.relaxCompDown...)
	}

	a.calibrated = true
}

// enforceMonotonicityUp ensures calibrationUp[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
// MUST be called with a.mu held.
func (a *App) enforceMonotonicityUp(idx int, Ec float64) {
	if idx <= 0 || idx >= len(a.calibrationUp) {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is greater than previous level
	if a.calibrationUp[idx] <= a.calibrationUp[idx-1] {
		a.calibrationUp[idx] = a.calibrationUp[idx-1] + step
	}

	// Ensure value is less than next level (if not at end)
	if idx < len(a.calibrationUp)-1 && a.calibrationUp[idx] >= a.calibrationUp[idx+1] {
		// Push next value up to maintain monotonicity
		a.calibrationUp[idx+1] = a.calibrationUp[idx] + step
	}
}

// enforceMonotonicityDown ensures calibrationDown[idx] maintains local monotonicity.
// Called after runtime calibration updates to prevent spikes.
// MUST be called with a.mu held.
func (a *App) enforceMonotonicityDown(idx int, Ec float64) {
	if idx <= 0 || idx >= len(a.calibrationDown)-1 {
		return
	}
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	// Ensure value is more negative than next level (lower index = more negative)
	if a.calibrationDown[idx] >= a.calibrationDown[idx+1] {
		a.calibrationDown[idx] = a.calibrationDown[idx+1] - step
	}

	// Ensure value is less negative than previous level
	if idx > 0 && a.calibrationDown[idx] <= a.calibrationDown[idx-1] {
		// Push previous value down to maintain monotonicity
		a.calibrationDown[idx-1] = a.calibrationDown[idx] - step
	}
}

// updateCalibrationUp updates ascending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
// MUST be called with a.mu held.
func (a *App) updateCalibrationUp(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(a.calibrationUp) {
		return
	}

	currentE := a.calibrationUp[targetIdx]
	lastErr := a.lastErrorUp[targetIdx]

	// Update bounds based on error direction
	if levelError > 0 {
		// Overshot (read level > target) → field too strong → update upper bound
		if currentE < a.calibUpHigh[targetIdx] {
			a.calibUpHigh[targetIdx] = currentE
		}
	} else {
		// Undershot (read level < target) → field too weak → update lower bound
		if currentE > a.calibUpLow[targetIdx] {
			a.calibUpLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if a.calibUpLow[targetIdx] < a.calibUpHigh[targetIdx] {
		newVal = (a.calibUpLow[targetIdx] + a.calibUpHigh[targetIdx]) / 2
	} else {
		// Bounds crossed - use direct error-proportional adjustment
		// This is more aggressive than midpoint to escape stuck states
		adjustment := float64(levelError) * Ec * 0.05 // 5% per level error
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Soft constraints: guide towards reasonable range but don't hard-clamp
	// This allows the binary search to explore beyond if physics requires it
	minE := Ec * 0.3 // Absolute minimum: 0.3×Ec
	maxE := Ec * 2.5 // Absolute maximum: 2.5×Ec

	if newVal < minE {
		newVal = minE
	} else if newVal > maxE {
		newVal = maxE
	}

	oldVal := a.calibrationUp[targetIdx]
	a.calibrationUp[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For ascending calibration: higher indices = higher fields
	//   calibrationUp[0] lowest, calibrationUp[29] highest
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal < oldVal {
		// Made field LOWER - cascade to lower indices (they must be even lower)
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := a.calibrationUp[i+1] - step // Must be lower than next
			if a.calibrationUp[i] > maxAllowed {
				a.calibrationUp[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal > oldVal {
		// Made field HIGHER - cascade to higher indices (they must be even higher)
		for i := targetIdx + 1; i < len(a.calibrationUp); i++ {
			minAllowed := a.calibrationUp[i-1] + step // Must be higher than previous
			if a.calibrationUp[i] < minAllowed {
				a.calibrationUp[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call enforceMonotonicityUp - cascade handles it and won't fight with newVal
	a.lastErrorUp[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(a.relaxCompUp) {
		relaxAdjust := 0.02 // 2% target adjustment per retry (more aggressive for faster convergence)
		oldRelaxComp := a.relaxCompUp[targetIdx]
		var targetRelax float64
		if levelError > 0 {
			// Overshot: read level > target, reduce compensation (we overshot too much)
			targetRelax = oldRelaxComp - relaxAdjust*float64(levelError) // Scale by error magnitude
		} else {
			// Undershot: read level < target, increase compensation (need more overshoot)
			targetRelax = oldRelaxComp - relaxAdjust*float64(levelError) // levelError is negative, so this adds
		}
		// Exponential moving average: 70% old + 30% new for smooth convergence
		a.relaxCompUp[targetIdx] = 0.7*oldRelaxComp + 0.3*targetRelax
		// Clamp to reasonable bounds [-0.05, 0.25]
		if a.relaxCompUp[targetIdx] < -0.05 {
			a.relaxCompUp[targetIdx] = -0.05
		} else if a.relaxCompUp[targetIdx] > 0.25 {
			a.relaxCompUp[targetIdx] = 0.25
		}
		if a.relaxCompUp[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_UP[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, a.relaxCompUp[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_UP[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		a.calibUpLow[targetIdx]/1e8, a.calibUpHigh[targetIdx]/1e8)

	// Refactoring: Sync to CalibrationManager
	if a.calibManager != nil {
		copy(a.calibManager.CalibrationUp, a.calibrationUp)
	}
}

// updateCalibrationDown updates descending calibration using binary search with bounds tracking.
// Called after write-verify fails to refine the field value for the given level.
// MUST be called with a.mu held.
func (a *App) updateCalibrationDown(targetIdx int, levelError int, Ec float64) {
	if targetIdx < 0 || targetIdx >= len(a.calibrationDown) {
		return
	}

	currentE := a.calibrationDown[targetIdx]
	lastErr := a.lastErrorDown[targetIdx]

	// Update bounds based on error direction (descending uses negative fields)
	// calibDownLow = most negative limit (e.g., -2.5×Ec)
	// calibDownHigh = least negative limit (e.g., -0.3×Ec)
	// midpoint searches between them
	if levelError > 0 {
		// Positive error = didn't go down enough = field not negative enough
		// Current field is too weak → becomes new UPPER bound (least negative limit)
		// Next search will be between Low and currentE (more negative)
		if currentE < a.calibDownHigh[targetIdx] {
			a.calibDownHigh[targetIdx] = currentE
		}
	} else {
		// Negative error = went too far down = field too negative
		// Current field is too strong → becomes new LOWER bound (most negative limit)
		// Next search will be between currentE and High (less negative)
		if currentE > a.calibDownLow[targetIdx] {
			a.calibDownLow[targetIdx] = currentE
		}
	}

	// Binary search: use midpoint of bounds
	var newVal float64
	if a.calibDownLow[targetIdx] < a.calibDownHigh[targetIdx] {
		newVal = (a.calibDownLow[targetIdx] + a.calibDownHigh[targetIdx]) / 2
	} else {
		// Bounds crossed - use direct error-proportional adjustment
		// For descending, positive error means we need MORE negative field
		adjustment := float64(levelError) * Ec * 0.05 // 5% per level error
		newVal = currentE - adjustment
	}

	// Detect oscillation and dampen
	if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
		newVal = currentE*0.7 + newVal*0.3
	}

	// Soft constraints: guide towards reasonable range but don't hard-clamp
	minE := -Ec * 2.5 // Absolute minimum (most negative): -2.5×Ec
	maxE := -Ec * 0.3 // Absolute maximum (least negative): -0.3×Ec

	if newVal > maxE {
		newVal = maxE
	} else if newVal < minE {
		newVal = minE
	}

	oldVal := a.calibrationDown[targetIdx]
	a.calibrationDown[targetIdx] = newVal

	// CASCADE to maintain monotonicity across all levels
	// For descending calibration: lower indices = more negative (stronger field)
	//   calibrationDown[0] most negative, calibrationDown[29] least negative (or 0)
	step := Ec * 0.02 // 2% of Ec minimum step between levels

	if newVal > oldVal {
		// Made field LESS negative (weaker) - cascade to higher indices
		// Higher indices must be even less negative to maintain monotonicity
		for i := targetIdx + 1; i < len(a.calibrationDown); i++ {
			minAllowed := a.calibrationDown[i-1] + step // Must be less negative than previous
			if a.calibrationDown[i] < minAllowed {
				a.calibrationDown[i] = minAllowed
			} else {
				break // Already satisfied
			}
		}
	} else if newVal < oldVal {
		// Made field MORE negative (stronger) - cascade to lower indices
		// Lower indices must be even more negative to maintain monotonicity
		for i := targetIdx - 1; i >= 0; i-- {
			maxAllowed := a.calibrationDown[i+1] - step // Must be more negative than next
			if a.calibrationDown[i] > maxAllowed {
				a.calibrationDown[i] = maxAllowed
			} else {
				break // Already satisfied
			}
		}
	}

	// Don't call enforceMonotonicityDown - cascade handles it and won't fight with newVal
	a.lastErrorDown[targetIdx] = levelError

	// Update relaxation compensation based on error direction (exponential moving average for smooth convergence)
	if targetIdx < len(a.relaxCompDown) {
		relaxAdjust := 0.02 // 2% target adjustment per retry (more aggressive for faster convergence)
		oldRelaxComp := a.relaxCompDown[targetIdx]
		var targetRelax float64
		if levelError < 0 {
			// Went too far down (error < 0), reduce compensation
			targetRelax = oldRelaxComp + relaxAdjust*float64(levelError) // levelError is negative, so this subtracts
		} else {
			// Didn't go down enough (error > 0), increase compensation
			targetRelax = oldRelaxComp + relaxAdjust*float64(levelError) // Scale by error magnitude
		}
		// Exponential moving average: 70% old + 30% new for smooth convergence
		a.relaxCompDown[targetIdx] = 0.7*oldRelaxComp + 0.3*targetRelax
		// Clamp to reasonable bounds [-0.05, 0.25]
		if a.relaxCompDown[targetIdx] < -0.05 {
			a.relaxCompDown[targetIdx] = -0.05
		} else if a.relaxCompDown[targetIdx] > 0.25 {
			a.relaxCompDown[targetIdx] = 0.25
		}
		if a.relaxCompDown[targetIdx] != oldRelaxComp {
			log.Printf("RELAX_DOWN[%d]: %.4f → %.4f (err=%+d)", targetIdx, oldRelaxComp, a.relaxCompDown[targetIdx], levelError)
		}
	}

	log.Printf("CALIB_DOWN[%d]: old=%.3f new=%.3f MV/cm, err=%+d, bounds=[%.3f,%.3f]",
		targetIdx, oldVal/1e8, newVal/1e8, levelError,
		a.calibDownLow[targetIdx]/1e8, a.calibDownHigh[targetIdx]/1e8)

	// Refactoring: Sync to CalibrationManager
	if a.calibManager != nil {
		copy(a.calibManager.CalibrationDown, a.calibrationDown)
	}
}
