package gui

import (
	"encoding/json"
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
	Temperature     float64   `json:"temperature_k"`     // Temperature in Kelvin
	CalibrationUp   []float64 `json:"calibration_up"`    // Ascending calibration values
	CalibrationDown []float64 `json:"calibration_down"`  // Descending calibration values
	CalibUpLow      []float64 `json:"calib_up_low"`      // Binary search lower bounds (ascending)
	CalibUpHigh     []float64 `json:"calib_up_high"`     // Binary search upper bounds (ascending)
	CalibDownLow    []float64 `json:"calib_down_low"`    // Binary search lower bounds (descending)
	CalibDownHigh   []float64 `json:"calib_down_high"`   // Binary search upper bounds (descending)
	LastErrorUp     []int     `json:"last_error_up"`     // Last error for oscillation detection
	LastErrorDown   []int     `json:"last_error_down"`   // Last error for oscillation detection
	RelaxCompUp     []float64 `json:"relax_comp_up"`     // Relaxation compensation factors (ascending)
	RelaxCompDown   []float64 `json:"relax_comp_down"`   // Relaxation compensation factors (descending)
}

// CalibrationData holds persistent calibration state (v2: multi-temperature support)
type CalibrationData struct {
	Version      int                        `json:"version"`       // Schema version (2 = multi-temp)
	MaterialName string                     `json:"material_name"` // Material these calibrations are for
	NumLevels    int                        `json:"num_levels"`    // Number of discrete levels
	Calibrations map[int]*TempCalibration   `json:"calibrations"`  // Key: temperature in Kelvin (rounded)
	SavedAt      string                     `json:"saved_at"`      // Timestamp

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

const calibrationVersion = 2

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

	// v2 format: multi-temperature calibration
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
func (a *App) simulationLoop() {
	ticker := time.NewTicker(16 * time.Millisecond) // ~60 FPS
	defer ticker.Stop()

	lastTime := time.Now()

	for a.running {
		<-ticker.C

		if a.paused {
			continue
		}

		a.mu.Lock()

		// Check material inside lock to prevent race condition
		if a.material == nil {
			a.mu.Unlock()
			continue
		}

		// Copy material reference under lock for safe access
		mat := a.material

		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		// Clamp dt to prevent animation "catch-up" after window loses focus
		// Max 100ms per frame keeps animation smooth when returning from background
		if dt > 0.1 {
			dt = 0.1
		}
		a.simTime += dt
		// Wrap simTime to prevent floating-point issues after long runs
		if a.simTime > 1000 {
			a.simTime = math.Mod(a.simTime, 1000)
		}

		// Generate E-field based on waveform
		if a.waveform == WaveformManual {
			// Manual mode: slider control or click-to-level animation
			//
			// PHYSICS: Hysteresis is PATH-DEPENDENT and NON-REVERSIBLE.
			// If you overshoot a target level, you CANNOT correct by applying less field
			// or opposite field (that's a different branch of the hysteresis loop).
			// You MUST reset to a known saturation state and try again.
			//
			// Phases:
			// 0: RESET - saturate in opposite direction to target
			// 1: HOLD_RESET - return to zero (now at known remanent: level 1 or N)
			// 2: WRITE - apply calibrated field toward target
			// 3: HOLD_WRITE - return to zero, polarization persists at target
			if a.manualAnimating {
				Ec := mat.Ec // Use local copy for thread safety
				Emax := Ec * 1.5 // Match calibration saturation field
				phaseDuration := 0.6 / a.frequency
				rampRate := 4.0 * Emax * a.frequency
				maxLevelIdx := a.numLevels - 1

				a.manualPhaseTime += dt

				targetLevel := a.manualTargetLevel // 1-indexed (1-N)
				startLevel := a.manualStartLevel   // Captured at animation start

				switch a.manualPhase {
				case 0: // RESET phase - always saturate to known state before writing
					// NOTE: ISPP skip-reset optimization removed (incompatible with Preisach model).
					// Preisach hysteresis is path-dependent - reliable level targeting requires
					// starting from a known saturation state. Incremental writes with sub-Ec
					// fields do not produce predictable switching.

					var resetE float64
					if targetLevel > startLevel {
						// Going UP: first saturate negative (reach level 1)
						resetE = -2.0 * Ec // Match calibration saturation (2.0×Ec)
					} else {
						// Going DOWN: first saturate positive (reach level N)
						resetE = 2.0 * Ec // Match calibration saturation (2.0×Ec)
					}

					// Ramp to reset field
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
					if a.manualPhaseTime > phaseDuration*0.2 && math.Abs(a.electricField) < 0.01*Emax {
						// Log the write field that will be used (only once at phase transition)
						targetIdx := targetLevel - 1
						goingUp := targetLevel > startLevel
						if targetIdx >= 0 && targetIdx < len(a.calibrationUp) {
							if goingUp {
								log.Printf("WRITE PHASE START: target=%d, calibUp[%d]=%.4f*Ec", targetLevel, targetIdx, a.calibrationUp[targetIdx]/Ec)
							} else {
								log.Printf("WRITE PHASE START: target=%d, calibDown[%d]=%.4f*Ec", targetLevel, targetIdx, a.calibrationDown[targetIdx]/Ec)
							}
						}
						a.manualPhase = 2
						a.manualPhaseTime = 0
					}

				case 2: // WRITE - apply calibrated field for target
					var writeE float64
					targetIdx := targetLevel - 1
					midLevel := a.numLevels / 2
					goingUp := targetLevel > midLevel

					// Use calibrated fields for reliable level targeting
					// (ISPP incremental writes removed - incompatible with Preisach model)
					if targetIdx < 0 || targetIdx >= len(a.calibrationUp) {
						// Out of bounds - use fallback
						ratio := float64(targetLevel-1) / float64(maxLevelIdx)
						if goingUp {
							writeE = Ec * (1.0 + ratio*1.0)
						} else {
							writeE = -Ec * (1.0 + ratio*1.0)
						}
					} else if goingUp {
						// Going UP from level 1: use ascending calibration
						writeE = a.calibrationUp[targetIdx]
						if writeE == 0 {
							// Fallback: interpolate based on target position
							ratio := float64(targetLevel-1) / float64(maxLevelIdx)
							writeE = Ec * (1.0 + ratio*1.0) // Ec to 2*Ec
						}
					} else {
						// Going DOWN from level N: use descending calibration
						writeE = a.calibrationDown[targetIdx]
						if writeE == 0 {
							ratio := float64(a.numLevels-targetLevel) / float64(maxLevelIdx)
							writeE = -Ec * (1.0 + ratio*1.0) // -Ec to -2*Ec
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

					// Transition when field applied and held
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

						// Log animation result with detailed state
						log.Printf("ANIMATION COMPLETE: target=%d, final=%d, error=%d, normalizedP=%.4f",
							targetLevel, finalLevel, levelError, a.normalizedP)

						// Update calibration using binary search with bounds tracking
						// This approach converges much faster and avoids oscillation
						adjIdx := targetLevel - 1
						if levelError != 0 && a.calibrated && adjIdx >= 0 && adjIdx < len(a.calibrationUp) {
							midLevel := a.numLevels / 2
							if targetLevel > midLevel {
								// ASCENDING calibration adjustment
								currentE := a.calibrationUp[adjIdx]
								lastErr := a.lastErrorUp[adjIdx]

								// Update bounds based on error direction
								if levelError > 0 {
									// Overshot (went too high) → field was too strong → update upper bound
									if currentE < a.calibUpHigh[adjIdx] {
										a.calibUpHigh[adjIdx] = currentE
									}
								} else {
									// Undershot (too low) → field was too weak → update lower bound
									if currentE > a.calibUpLow[adjIdx] {
										a.calibUpLow[adjIdx] = currentE
									}
								}

								// Binary search: use midpoint of bounds
								var newVal float64
								if a.calibUpLow[adjIdx] < a.calibUpHigh[adjIdx] {
									// Valid bounds: use midpoint
									newVal = (a.calibUpLow[adjIdx] + a.calibUpHigh[adjIdx]) / 2
								} else {
									// Bounds crossed (shouldn't happen) - use small adjustment
									adjustment := float64(levelError) * Ec * 0.01
									newVal = currentE - adjustment
								}

								// Detect oscillation (error sign flipped) and dampen
								if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
									// Oscillating - use weighted average closer to current
									newVal = currentE*0.7 + newVal*0.3
								}

								// Level-dependent minimum field: higher levels need stronger fields
								// Linear interpolation: level 1 needs ~0.4×Ec, level 29 needs ~1.4×Ec
								maxLevel := float64(a.numLevels - 1)
								levelRatio := float64(adjIdx) / maxLevel
								minE := Ec * (0.4 + levelRatio*1.0) // Range: 0.4×Ec to 1.4×Ec
								maxE := Ec * (0.6 + levelRatio*1.2) // Range: 0.6×Ec to 1.8×Ec

								if newVal < minE {
									newVal = minE
								} else if newVal > maxE {
									newVal = maxE
								}
								a.calibrationUp[adjIdx] = newVal
								a.enforceMonotonicityUp(adjIdx, Ec) // Prevent spikes from runtime updates
								a.lastErrorUp[adjIdx] = levelError
								log.Printf("CALIB UP[%d]: bounds=[%.4f,%.4f]*Ec, new=%.4f*Ec, err=%d",
									adjIdx, a.calibUpLow[adjIdx]/Ec, a.calibUpHigh[adjIdx]/Ec, newVal/Ec, levelError)
							} else {
								// DESCENDING calibration adjustment
								currentE := a.calibrationDown[adjIdx]
								lastErr := a.lastErrorDown[adjIdx]

								// Update bounds (note: descending uses negative fields)
								if levelError > 0 {
									// Overshot UP (didn't go down enough) → field not negative enough → update lower bound
									if currentE > a.calibDownLow[adjIdx] {
										a.calibDownLow[adjIdx] = currentE
									}
								} else {
									// Undershot (went too far down) → field too negative → update upper bound
									if currentE < a.calibDownHigh[adjIdx] {
										a.calibDownHigh[adjIdx] = currentE
									}
								}

								// Binary search: use midpoint of bounds
								var newVal float64
								if a.calibDownLow[adjIdx] < a.calibDownHigh[adjIdx] {
									// Valid bounds: use midpoint
									newVal = (a.calibDownLow[adjIdx] + a.calibDownHigh[adjIdx]) / 2
								} else {
									// Bounds crossed - use small adjustment
									adjustment := float64(levelError) * Ec * 0.01
									newVal = currentE - adjustment
								}

								// Detect oscillation and dampen
								if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
									newVal = currentE*0.7 + newVal*0.3
								}

								// Level-dependent minimum field: lower levels need stronger (more negative) fields
								// For descending: level 29 needs ~-0.4×Ec, level 1 needs ~-1.4×Ec
								maxLevel := float64(a.numLevels - 1)
								levelRatio := float64(adjIdx) / maxLevel
								// Invert: low level index = strong negative field
								minE := -Ec * (0.6 + (1-levelRatio)*1.2) // Range: -1.8×Ec to -0.6×Ec
								maxE := -Ec * (0.4 + (1-levelRatio)*1.0) // Range: -1.4×Ec to -0.4×Ec

								if newVal > maxE {
									newVal = maxE
								} else if newVal < minE {
									newVal = minE
								}
								a.calibrationDown[adjIdx] = newVal
								a.enforceMonotonicityDown(adjIdx, Ec) // Prevent spikes from runtime updates
								a.lastErrorDown[adjIdx] = levelError
								log.Printf("CALIB DOWN[%d]: bounds=[%.4f,%.4f]*Ec, new=%.4f*Ec, err=%d",
									adjIdx, a.calibDownLow[adjIdx]/Ec, a.calibDownHigh[adjIdx]/Ec, newVal/Ec, levelError)
							}
						}

						a.manualAnimating = false
						a.manualPhase = 0
						a.addLogEntry(fmt.Sprintf("→ Level %d (target %d)", finalLevel, targetLevel))
					}
				}
			}
			// If not animating, electric field is already set by slider in controls.go
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
				// Correct ferroelectric write/read physics with RESET-AND-RETRY approach:
				//
				// PHYSICS: Hysteresis is PATH-DEPENDENT and NON-REVERSIBLE.
				// If you overshoot a target level, you CANNOT correct by applying less field
				// or opposite field (that's a different branch of the hysteresis loop).
				// You MUST reset to a known saturation state and apply precise programming pulse.
				//
				// Phase mapping:
				// 0 = RESET (saturate in opposite direction to target)
				// 1 = HOLD_RESET (return to zero - now at known remanent: level 1 or N)
				// 2 = WRITE (apply calibrated field toward target)
				// 3 = HOLD_WRITE (return to zero, polarization persists)
				// 4 = READ (small sense pulse below Ec)
				// 5 = DISPLAY (show result, pick next target)

				a.wrdPhaseTimer += dt
				phaseDuration := 1.0 / a.frequency
				rampRate := 3.0 * Emax * a.frequency
				Ec := mat.Ec // Use local copy for thread safety
				midLevel := a.numLevels / 2
				maxLevelIdx := a.numLevels - 1

				targetLevel := a.wrdTargetLevel // 1-indexed
				// Note: startLevel (a.wrdStartLevel) no longer used for direction - we use absolute position

				switch a.wrdPhase {
				case 0: // RESET - saturate in opposite direction to target
					var resetE float64
					// Use absolute level position (not relative to start) for consistent calibration
					// Calibration was measured from saturated states, so we must match that
					if targetLevel > midLevel {
						// Target in upper half: saturate negative first (reach level 1), then apply ascending cal
						resetE = -2.0 * Ec // Match calibration saturation (2.0×Ec)
					} else {
						// Target in lower half: saturate positive first (reach level N), then apply descending cal
						resetE = 2.0 * Ec // Match calibration saturation (2.0×Ec)
					}
					a.wrdSaturateE = resetE // Store for logging

					// Ramp to reset field
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
					if a.wrdPhaseTimer > phaseDuration*0.25 && math.Abs(a.electricField-resetE) < 0.01*Emax {
						// Capture end-of-RESET state for logging
						a.wrdResetEndP = a.polarization * 100 // Convert to µC/cm²
						a.wrdResetEndLvl = a.discreteLevel + 1
						log.Printf("WRD PHASE 0→1: RESET done | E=%.3f MV/cm | P=%.2f→%.2f µC/cm² | L=%d→%d | target=%d",
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
						a.wrdPhase = 2
						a.wrdPhaseTimer = 0
					}

				case 2: // WRITE - apply calibrated field for target
					var writeE float64
					// Use absolute level position to match calibration measurement conditions
				goingUp := targetLevel > midLevel
					wrdTargetIdx := targetLevel - 1
					currentLevel := a.discreteLevel + 1 // 1-indexed

					// Check if this is a retry from undershoot (skipped RESET)
					// In undershoot retry, we need incremental field from current position
					isUndershootRetry := a.wrdRetryCount > 0 && a.wrdResetEndLvl == 0

					// Bounds check for calibration array access
					if wrdTargetIdx < 0 || wrdTargetIdx >= len(a.calibrationUp) {
						// Out of bounds - use fallback
						ratio := float64(targetLevel-1) / float64(maxLevelIdx)
						if goingUp {
							writeE = Ec * (1.0 + ratio*1.0)
						} else {
							writeE = -Ec * (1.0 + ratio*1.0)
						}
					} else if isUndershootRetry {
						// UNDERSHOOT RETRY: We're on the correct branch, just apply more field
						// No need to saturate - just push harder toward target
						if goingUp {
							// Get field for target level
							targetE := a.calibrationUp[wrdTargetIdx]
							if targetE == 0 {
								ratio := float64(targetLevel-1) / float64(maxLevelIdx)
								targetE = Ec * (1.0 + ratio*1.0)
							}
							// Apply slightly more field than the difference to ensure we reach target
							writeE = targetE * 1.05 // 5% extra to overcome any drift
						} else {
							// Get field for target level
							targetE := a.calibrationDown[wrdTargetIdx]
							if targetE == 0 {
								ratio := float64(a.numLevels-targetLevel) / float64(maxLevelIdx)
								targetE = -Ec * (1.0 + ratio*1.0)
							}
							writeE = targetE * 1.05 // 5% extra (more negative)
						}
						log.Printf("WRD UNDERSHOOT WRITE: currentLvl=%d targetLvl=%d | applying E=%.3f MV/cm",
							currentLevel, targetLevel, writeE/1e8)
					} else if goingUp {
						// Going UP from level 1: use ascending calibration
						baseE := a.calibrationUp[wrdTargetIdx]
						if baseE == 0 {
							// Fallback: interpolate based on target position
							ratio := float64(targetLevel-1) / float64(maxLevelIdx)
							baseE = Ec * (1.0 + ratio*1.0) // Ec to 2*Ec
						}
						// Apply relaxation compensation (overshoot to counteract relaxation drift)
						compFactor := 1.0
						if wrdTargetIdx < len(a.relaxCompUp) {
							compFactor = 1.0 + a.relaxCompUp[wrdTargetIdx]
						}
						writeE = baseE * compFactor
					} else {
						// Going DOWN from level N: use descending calibration
						baseE := a.calibrationDown[wrdTargetIdx]
						if baseE == 0 {
							ratio := float64(a.numLevels-targetLevel) / float64(maxLevelIdx)
							baseE = -Ec * (1.0 + ratio*1.0) // -Ec to -2*Ec
						}
						// Apply relaxation compensation
						compFactor := 1.0
						if wrdTargetIdx < len(a.relaxCompDown) {
							compFactor = 1.0 + a.relaxCompDown[wrdTargetIdx]
						}
						writeE = baseE * compFactor
					}
					a.wrdWriteE = writeE // Store for logging

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

					// Transition when field applied and held
					if a.wrdPhaseTimer > phaseDuration*0.3 && math.Abs(a.electricField-writeE) < 0.01*Emax {
						// Capture end-of-WRITE state for logging
						a.wrdWriteEndP = a.polarization * 100 // Convert to µC/cm²
						a.wrdWriteEndLvl = a.discreteLevel + 1
						log.Printf("WRD PHASE 2→3: WRITE done | E=%.3f MV/cm (%.2f×Ec) | P=%.2f→%.2f µC/cm² | L=%d→%d | target=%d",
							a.electricField/1e8, a.electricField/Ec, a.wrdWriteStartP, a.wrdWriteEndP, a.wrdResetEndLvl, a.wrdWriteEndLvl, targetLevel)
						a.wrdPhase = 3
						a.wrdPhaseTimer = 0
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

					// Transition to READ phase
					if a.wrdPhaseTimer > phaseDuration*0.2 && math.Abs(a.electricField) < 0.01*Emax {
						// Capture start-of-READ state for logging
						a.wrdReadStartP = a.polarization * 100 // Convert to µC/cm²
						log.Printf("WRD PHASE 3→4: HOLD done | E=%.3f MV/cm | P=%.2f µC/cm² | L=%d | ready to read",
							a.electricField/1e8, a.wrdReadStartP, a.discreteLevel+1)
						a.wrdPhase = 4
						a.wrdPhaseTimer = 0
					}

				case 4: // READ phase - small sense pulse below Ec
					// READ pulse direction should match WRITE direction to minimize disturbing the state
					// Applying opposite polarity read would push polarization further away from written state
					var readE float64
					if targetLevel > midLevel {
						// Ascending calibration used positive E - read with same polarity
						readE = Ec * 0.3
					} else {
						// Descending calibration used negative E - read with same polarity
						readE = -Ec * 0.3
					}
					step := rampRate * 0.4 * dt
					diff := readE - a.electricField
					if math.Abs(diff) < step {
						a.electricField = readE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Capture read level and VERIFY
					if a.wrdPhaseTimer > phaseDuration*0.3 {
						a.wrdReadLevel = a.discreteLevel + 1
						levelError := a.wrdReadLevel - a.wrdTargetLevel
						success := levelError == 0 // Exact match required - retry until we hit the exact target

						// WRITE-VERIFY-RETRY LOOP (INFINITE UNTIL SUCCESS)
						// MUST hit the target - no giving up!
						if success {
							// SUCCESS: Proceed to DISPLAY phase
							a.wrdPhase = 5
							a.wrdPhaseTimer = 0

							// Track metrics - 100% success rate guaranteed
							a.wrdTotalWrites++
							a.wrdSuccessWrites++
							successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100

							if a.wrdRetryCount > 0 {
								log.Printf("WRD PHASE 4→5: TARGET HIT after %d retries | L_read=%d L_target=%d | rate=%.1f%% (%d/%d)",
									a.wrdRetryCount, a.wrdReadLevel, a.wrdTargetLevel,
									successRate, a.wrdSuccessWrites, a.wrdTotalWrites)
							} else {
								log.Printf("WRD PHASE 4→5: TARGET HIT (1st try) | L_read=%d L_target=%d | rate=%.1f%% (%d/%d)",
									a.wrdReadLevel, a.wrdTargetLevel,
									successRate, a.wrdSuccessWrites, a.wrdTotalWrites)
							}

							// Reset retry count for next target
							a.wrdRetryCount = 0

							// Add accumulated energy for this cycle
							a.wrdTotalEnergyfJ += a.wrdCycleEnergy
							a.wrdCycleEnergy = 0

							// Log this cycle for debugging
							if a.wrdDebugLog != nil {
								cycle := WriteReadCycle{
									CycleNum:    len(a.wrdDebugLog.Cycles) + 1,
									TargetLevel: a.wrdTargetLevel,
									StartLevel:  a.wrdStartLevel,
									ReadLevel:   a.wrdReadLevel,
									Success:     true, // Always true now - we only get here on success
									Phases: []WriteReadPhase{
										{
											Phase:      "RESET",
											EFieldPeak: a.wrdSaturateE / 1e8,
											PStart:     a.wrdResetStartP,
											PEnd:       a.wrdResetEndP,
											LevelStart: a.wrdStartLevel,
											LevelEnd:   a.wrdResetEndLvl,
										},
										{
											Phase:      "WRITE",
											EFieldPeak: a.wrdWriteE / 1e8,
											PStart:     a.wrdWriteStartP,
											PEnd:       a.wrdWriteEndP,
											LevelStart: a.wrdResetEndLvl,
											LevelEnd:   a.wrdWriteEndLvl,
										},
										{
											Phase:      "READ",
											EFieldPeak: readE / 1e8,
											PStart:     a.wrdReadStartP,
											PEnd:       a.wrdReadStartP,
											LevelStart: a.wrdWriteEndLvl,
											LevelEnd:   a.wrdReadLevel,
										},
									},
								}
								a.wrdDebugLog.Cycles = append(a.wrdDebugLog.Cycles, cycle)

								if len(a.wrdDebugLog.Cycles) > 100 {
									a.wrdDebugLog.Cycles = a.wrdDebugLog.Cycles[len(a.wrdDebugLog.Cycles)-100:]
								}

								if len(a.wrdDebugLog.Cycles)%5 == 0 {
									go a.saveDebugLog()
								}
							}
						} else {
							// FAILED: Update calibration and RETRY (NO LIMIT - must hit target!)
							a.wrdRetryCount++

							// Determine if overshoot or undershoot based on direction
							// goingUp: positive field, higher levels
							// goingDown: negative field, lower levels
							goingUp := a.wrdTargetLevel > midLevel

							// Undershoot: didn't apply enough field, might continue without reset
							// Overshoot: went past target, MUST reset (hysteresis is path-dependent)
							var isUndershoot bool
							absError := abs(levelError)
							if goingUp {
								// Going up (positive field): undershoot means read < target (levelError < 0)
								isUndershoot = levelError < 0
							} else {
								// Going down (negative field): undershoot means read > target (levelError > 0)
								isUndershoot = levelError > 0
							}

							// Only skip RESET for small undershoots (1-2 levels) on first retry
							// READ phase can disturb polarization, so larger errors need full RESET
							canSkipReset := isUndershoot && absError <= 2 && a.wrdRetryCount == 1

							if canSkipReset {
								log.Printf("WRD VERIFY UNDERSHOOT: L_read=%d L_target=%d err=%+d | RETRY #%d (skip RESET, small error)",
									a.wrdReadLevel, a.wrdTargetLevel, levelError, a.wrdRetryCount)
							} else if isUndershoot {
								log.Printf("WRD VERIFY UNDERSHOOT: L_read=%d L_target=%d err=%+d | RETRY #%d (RESET needed, error too large or repeated)",
									a.wrdReadLevel, a.wrdTargetLevel, levelError, a.wrdRetryCount)
							} else {
								log.Printf("WRD VERIFY OVERSHOOT: L_read=%d L_target=%d err=%+d | RETRY #%d (must RESET)",
									a.wrdReadLevel, a.wrdTargetLevel, levelError, a.wrdRetryCount)
							}

							// Update calibration BEFORE retry to converge on correct field
							targetIdx := a.wrdTargetLevel - 1
							if a.calibrated && targetIdx >= 0 && targetIdx < len(a.calibrationUp) {
								if goingUp {
									a.updateCalibrationUp(targetIdx, levelError, Ec)
								} else {
									a.updateCalibrationDown(targetIdx, levelError, Ec)
								}
							}

							// Clear history trail to prevent visual spikes during retry
							a.eHistory = a.eHistory[:0]
							a.pHistory = a.pHistory[:0]

							if canSkipReset {
								// SMALL UNDERSHOOT: Skip RESET, use BOOST phase (phase 6)
								// We're still on the same branch of the hysteresis loop
								a.wrdWriteStartP = a.polarization * 100
								a.wrdResetEndLvl = 0 // Mark that we skipped RESET (used by WRITE phase)
								a.wrdPhase = 6       // BOOST phase (undershoot retry)
								a.wrdPhaseTimer = 0
							} else {
								// OVERSHOOT or LARGE UNDERSHOOT: Must fully reset
								a.wrdResetStartP = a.polarization * 100
								a.wrdPhase = 0 // Full RESET
								a.wrdPhaseTimer = 0
							}
						}
					}

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
						// Record start level for next cycle
						a.wrdStartLevel = a.discreteLevel + 1

						// Add comparison callout every 5 cycles
						if a.wrdTotalWrites > 0 && a.wrdTotalWrites%5 == 0 {
							fecimEnergy := a.wrdTotalEnergyfJ / 1000 // pJ
							// NOTE: 10M× is Dr. Tour's unverified claim. Peer-reviewed: 25-100× (Samsung Nature 2025)
							nandEquiv := fecimEnergy * 50            // 25-100× better (conservative: use 50)
							dramEquiv := fecimEnergy * 1000          // 1000× worse
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
						// Capture start-of-RESET state for next cycle logging
						a.wrdResetStartP = a.polarization * 100 // Convert to µC/cm²
						log.Printf("WRD CYCLE START: cycle=%d | startLevel=%d | newTarget=%d | P=%.2f µC/cm²",
							a.wrdTotalWrites+1, a.discreteLevel+1, a.wrdTargetLevel, a.wrdResetStartP)
						// Clear history when starting new cycle to avoid vertical spike artifacts
						// The trail from the previous cycle (opposite direction) would create visual spikes
						a.eHistory = a.eHistory[:0]
						a.pHistory = a.pHistory[:0]
						a.wrdPhase = 0
						a.wrdPhaseTimer = 0
						a.wrdCycleEnergy = 0 // Reset energy accumulator for next cycle
					}

				case 6: // BOOST phase - undershoot retry, skip RESET and apply more field
					// This phase handles small undershoots by directly applying more field
					// without going through the full RESET cycle
					var writeE float64
					goingUp := targetLevel > midLevel
					wrdTargetIdx := targetLevel - 1

					if wrdTargetIdx < 0 || wrdTargetIdx >= len(a.calibrationUp) {
						ratio := float64(targetLevel-1) / float64(maxLevelIdx)
						if goingUp {
							writeE = Ec * (1.0 + ratio*1.0) * 1.08 // 8% extra for boost
						} else {
							writeE = -Ec * (1.0 + ratio*1.0) * 1.08
						}
					} else if goingUp {
						targetE := a.calibrationUp[wrdTargetIdx]
						if targetE == 0 {
							ratio := float64(targetLevel-1) / float64(maxLevelIdx)
							targetE = Ec * (1.0 + ratio*1.0)
						}
						writeE = targetE * 1.08 // 8% extra for boost
					} else {
						targetE := a.calibrationDown[wrdTargetIdx]
						if targetE == 0 {
							ratio := float64(a.numLevels-targetLevel) / float64(maxLevelIdx)
							targetE = -Ec * (1.0 + ratio*1.0)
						}
						writeE = targetE * 1.08 // 8% extra for boost
					}
					a.wrdWriteE = writeE

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

					// Transition to HOLD phase when field applied
					if a.wrdPhaseTimer > phaseDuration*0.25 && math.Abs(a.electricField-writeE) < 0.01*Emax {
						a.wrdWriteEndP = a.polarization * 100
						a.wrdWriteEndLvl = a.discreteLevel + 1
						log.Printf("WRD PHASE 6→3: BOOST done | E=%.3f MV/cm (%.2f×Ec) | P=%.2f→%.2f µC/cm² | L=%d | target=%d",
							a.electricField/1e8, a.electricField/Ec, a.wrdWriteStartP, a.wrdWriteEndP, a.wrdWriteEndLvl, targetLevel)
						a.wrdPhase = 3 // Go to HOLD phase
						a.wrdPhaseTimer = 0
					}
				}
			case WaveformTimeResolved:
				// Time-resolved switching dynamics visualization
				// Shows KAI (Kolmogorov-Avrami-Ishibashi) stretched exponential switching
				if !a.timeResAnimating {
					// Start new animation - simulate domain switching dynamics
					Eapplied := 2.0 * a.material.Ec       // Write pulse at 2×Ec
					duration := 100e-9         // 100 nanoseconds
					steps := 100               // 100 time points

					times, pols, switched := a.preisach.SimulateDomainSwitching(Eapplied, duration, steps)

					a.timeResDataTimes = times
					a.timeResDataPols = pols
					a.timeResDataSwitch = switched
					a.timeResIndex = 0
					a.timeResAnimating = true

					// Clear history for clean display
					a.eHistory = a.eHistory[:0]
					a.pHistory = a.pHistory[:0]

					a.addLogEntry("━━ TIME-RESOLVED SWITCHING ━━")
					a.addLogEntry(fmt.Sprintf("E = %.1f MV/cm (2×Ec)", Eapplied/1e8))
					a.addLogEntry(fmt.Sprintf("Duration: %.0f ns", duration*1e9))
					a.addLogEntry("KAI stretched exponential")
					a.addLogEntry("P(t)=Ps(1-exp(-(t/τ)^n))")
				}

				// Animate through the precomputed data
				if a.timeResAnimating && len(a.timeResDataTimes) > 0 {
					// Advance at a rate of ~2 samples per frame (controlled animation speed)
					a.timeResIndex += 2
					if a.timeResIndex >= len(a.timeResDataTimes) {
						// Loop back to start for continuous demonstration
						a.timeResIndex = 0
						a.eHistory = a.eHistory[:0]
						a.pHistory = a.pHistory[:0]
						a.addLogEntry("─── Loop ───")
					}

					// Set current state from precomputed data
					idx := a.timeResIndex
					currentTime := a.timeResDataTimes[idx]
					a.polarization = a.timeResDataPols[idx]
					a.electricField = 2.0 * a.material.Ec * (1.0 - math.Exp(-math.Pow(currentTime/(100e-9/10), 2.0)))

					// Update discrete level
					a.normalizedP = a.polarization / a.material.Ps
					maxLevel := a.numLevels - 1
					a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * float64(maxLevel)))
					if a.discreteLevel < 0 {
						a.discreteLevel = 0
					}
					if a.discreteLevel > maxLevel {
						a.discreteLevel = maxLevel
					}

					// Log progress at key milestones
					if idx == 10 {
						switchedFrac := float64(a.timeResDataSwitch[idx]) / float64(len(a.timeResDataSwitch)) * 100
						a.addLogEntry(fmt.Sprintf("10 ns: %.0f%% switched", switchedFrac))
					} else if idx == 50 {
						switchedFrac := float64(a.timeResDataSwitch[idx]) / float64(len(a.timeResDataSwitch)) * 100
						a.addLogEntry(fmt.Sprintf("50 ns: %.0f%% switched", switchedFrac))
					} else if idx == 90 {
						switchedFrac := float64(a.timeResDataSwitch[idx]) / float64(len(a.timeResDataSwitch)) * 100
						a.addLogEntry(fmt.Sprintf("90 ns: %.0f%% switched", switchedFrac))
						a.addLogEntry("τ = switching time constant")
						a.addLogEntry("n = Avrami exponent")
					}
				}
			}
		}

		// Update physics
		prevP := a.polarization
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
			energyJ := math.Abs(a.electricField * deltaP) * cellVolume
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

		// Copy data for UI update
		eField := a.electricField
		pol := a.polarization
		level := a.discreteLevel
		materialEc := mat.Ec // Capture Ec under lock for thread safety
		eHist := make([]float64, len(a.eHistory))
		pHist := make([]float64, len(a.pHistory))
		copy(eHist, a.eHistory)
		copy(pHist, a.pHistory)

		a.mu.Unlock()

		// Update UI (must be on main thread)
		a.updateUI(eField, pol, level, materialEc, eHist, pHist)
	}
}

// updateUI updates all UI elements with the latest simulation data.
// materialEc is passed as parameter to avoid race conditions with a.material access.
func (a *App) updateUI(eField, pol float64, level int, materialEc float64, eHist, pHist []float64) {
	fyne.Do(func() {
		// Update labels
		a.eFieldLabel.SetText(fmt.Sprintf("E-field: %.3f MV/cm", eField/1e8))
		a.pLabel.SetText(fmt.Sprintf("%.2f µC/cm²", pol*100))
		a.mu.RLock()
		numLevels := a.numLevels
		a.mu.RUnlock()
		a.levelLabel.SetText(fmt.Sprintf("%d/%d", level+1, numLevels))

		// Update state descriptor (divide into thirds)
		var stateText string
		lowThird := numLevels / 3
		highThird := numLevels * 2 / 3
		if level < lowThird {
			stateText = "Negative P"
		} else if level >= highThird {
			stateText = "Positive P"
		} else {
			stateText = "Intermediate"
		}
		if a.stateLabel != nil {
			a.stateLabel.SetText(stateText)
		}

		// Update wake-up/fatigue labels (Dr. Tour recommendation)
		cycles, degradation, wakeup := a.preisach.GetFatigueState()
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

		// Update temperature-dependent metrics
		effEc := a.preisach.GetEffectiveEc()
		effPr := a.preisach.GetEffectivePr()
		switchedFraction := a.preisach.GetSwitchedFraction()

		// Calculate squareness (Pr/Ps ratio)
		squareness := 0.0
		a.mu.RLock()
		if a.material != nil && a.material.Ps > 0 {
			squareness = effPr / a.material.Ps
		}
		a.mu.RUnlock()

		if a.effEcLabel != nil {
			a.effEcLabel.SetText(fmt.Sprintf("Ec(T): %.2f MV/cm", effEc/1e8))
		}
		if a.effPrLabel != nil {
			a.effPrLabel.SetText(fmt.Sprintf("Pr(T): %.1f µC/cm²", effPr*100))
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
			a.eFieldSlider.SetValue(eField / materialEc)
		}

		// Update status and logging
		if a.paused {
			a.statusLabel.SetText("⏸ Paused")
		} else {
			a.mu.RLock()
			waveform := a.waveform
			wrdPhase := a.wrdPhase
			wrdTarget := a.wrdTargetLevel
			wrdRead := a.wrdReadLevel
			lastPhase := a.lastLogPhase
			wrdTotalWrites := a.wrdTotalWrites
			wrdSuccessWrites := a.wrdSuccessWrites
			wrdTotalEnergyfJ := a.wrdTotalEnergyfJ
			midLevel := a.numLevels / 2 // Dynamic middle level for direction logic
			a.mu.RUnlock()

			switch waveform {
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
						a.addLogEntry(fmt.Sprintf("░░ HOLD L%d | E=0 | 0 fJ!", level+1))
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
					phaseStr = fmt.Sprintf("░ HOLD L%d | E=0 | ZERO POWER", level+1)
				case 4:
					phaseStr = fmt.Sprintf("▒ READ | Sense L%d | ~1fJ", level+1)
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
				animating := a.manualAnimating
				manPhase := a.manualPhase
				manTarget := a.manualTargetLevel
				manStart := a.manualStartLevel
				a.mu.RUnlock()

				if animating {
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
						phaseStr = fmt.Sprintf("HOLD L%d...", level+1)
					default:
						phaseStr = fmt.Sprintf("Current: L%d", level+1)
					}
					a.statusLabel.SetText(fmt.Sprintf("TARGET L%d | %s", manTarget, phaseStr))
				} else {
					a.statusLabel.SetText(fmt.Sprintf("Manual L%d | Click level bar", level+1))
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
						currentTime*1e9, switchedFrac, level+1))
				} else {
					a.statusLabel.SetText("⚡ Time-Resolved Switching (KAI Dynamics)")
				}
			default:
				frac := a.preisach.GetSwitchedFraction() * 100
				a.statusLabel.SetText(fmt.Sprintf("● Running | t=%.2fs | Switched: %.1f%%", a.simTime, frac))
			}
		}

		// Update slide text based on current waveform
		a.slideText.SetText(a.getSlideText())

		// Update log text
		a.mu.RLock()
		logText := a.getLogText()
		a.mu.RUnlock()
		a.logText.SetText(logText)

		// Update plot
		a.plot.SetData(eHist, pHist, eField, pol)
		a.plot.Refresh()

		// Update level indicator (level is 0-indexed, display is 1-indexed)
		a.levelIndicator.SetLevel(level)

		// Highlight target level during animations
		a.mu.RLock()
		currentWaveform := a.waveform
		currentWrdPhase := a.wrdPhase
		currentWrdTarget := a.wrdTargetLevel
		manualAnim := a.manualAnimating
		manualTarget := a.manualTargetLevel
		a.mu.RUnlock()

		if currentWaveform == WaveformWriteReadDemo {
			// Show target during phases 0-4 (RESET/SETTLE/WRITE/HOLD/READ)
			highlight := currentWrdPhase >= 0 && currentWrdPhase <= 4
			a.levelIndicator.SetTargetLevel(currentWrdTarget, highlight)
		} else if currentWaveform == WaveformManual && manualAnim {
			// Show target during Manual mode click animation
			a.levelIndicator.SetTargetLevel(manualTarget, true)
		} else {
			// Clear target highlight
			a.levelIndicator.SetTargetLevel(0, false)
		}

		a.levelIndicator.Refresh()

		// Update cell visualizer
		a.cellViz.SetLevel(level)
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

		normalizedP := p / a.material.Ps
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

		normalizedP := p / a.material.Ps
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
		lowE := Ec * 0.3  // Match initialized bounds for full coverage
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
	minE := Ec * 0.3  // Absolute minimum: 0.3×Ec
	maxE := Ec * 2.5  // Absolute maximum: 2.5×Ec

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
}
