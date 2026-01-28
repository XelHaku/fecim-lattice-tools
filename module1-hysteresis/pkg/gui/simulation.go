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
	}

	data := CalibrationData{
		Version:      calibrationVersion,
		MaterialName: a.material.Name,
		NumLevels:    a.numLevels,
		Calibrations: calibrations,
		SavedAt:      time.Now().Format(time.RFC3339),
	}

	// Ensure data directory exists
	dir := filepath.Dir(calibrationFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	// Write JSON file
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal calibration: %w", err)
	}

	if err := os.WriteFile(calibrationFile, jsonData, 0644); err != nil {
		return fmt.Errorf("write calibration file: %w", err)
	}

	log.Printf("Calibration saved for material: %s (%d levels, %d temperatures)", a.material.Name, a.numLevels, len(calibrations))
	return nil
}

// loadCalibration loads calibration data from disk if valid for current material
// Returns true if calibration was loaded successfully
// Supports both v1 (single-temp) and v2 (multi-temp) formats
func (a *App) loadCalibration() bool {
	if a.material == nil {
		return false
	}

	jsonData, err := os.ReadFile(calibrationFile)
	if err != nil {
		// File doesn't exist yet - that's fine
		return false
	}

	var data CalibrationData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		log.Printf("Invalid calibration file, will recalibrate: %v", err)
		return false
	}

	// Validate material and levels must match
	if data.MaterialName != a.material.Name {
		log.Printf("Calibration material mismatch (got %s, want %s), will recalibrate", data.MaterialName, a.material.Name)
		return false
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

		log.Printf("Calibration migrated from v1: %s (%d levels at 300K, saved %s)", data.MaterialName, data.NumLevels, data.SavedAt)
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

	log.Printf("Calibration loaded for material: %s (%d levels, %d temperatures, saved %s)", data.MaterialName, data.NumLevels, len(a.tempCalibrations), data.SavedAt)
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
	a.calibrationTemp = cal.Temperature
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
						resetE = -1.5 * Ec // Match calibration saturation
					} else {
						// Going DOWN: first saturate positive (reach level N)
						resetE = 1.5 * Ec // Match calibration saturation
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

								// Clamp to valid range (allow weaker fields for mid-levels)
								if newVal < Ec*0.3 {
									newVal = Ec * 0.3
								} else if newVal > Ec*2.5 {
									newVal = Ec * 1.5
								}
								a.calibrationUp[adjIdx] = newVal
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

								// Clamp to valid range (allow weaker fields for mid-levels)
								if newVal > -Ec*0.3 {
									newVal = -Ec * 0.3
								} else if newVal < -Ec*2.5 {
									newVal = -Ec * 1.5
								}
								a.calibrationDown[adjIdx] = newVal
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
						resetE = -1.5 * Ec // Match calibration saturation
					} else {
						// Target in lower half: saturate positive first (reach level N), then apply descending cal
						resetE = 1.5 * Ec // Match calibration saturation
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

					// Bounds check for calibration array access
					if wrdTargetIdx < 0 || wrdTargetIdx >= len(a.calibrationUp) {
						// Out of bounds - use fallback
						ratio := float64(targetLevel-1) / float64(maxLevelIdx)
						if goingUp {
							writeE = Ec * (1.0 + ratio*1.0)
						} else {
							writeE = -Ec * (1.0 + ratio*1.0)
						}
					} else if goingUp {
						// Going UP from level 1: use ascending calibration
						writeE = a.calibrationUp[wrdTargetIdx]
						if writeE == 0 {
							// Fallback: interpolate based on target position
							ratio := float64(targetLevel-1) / float64(maxLevelIdx)
							writeE = Ec * (1.0 + ratio*1.0) // Ec to 2*Ec
						}
					} else {
						// Going DOWN from level N: use descending calibration
						writeE = a.calibrationDown[wrdTargetIdx]
						if writeE == 0 {
							ratio := float64(a.numLevels-targetLevel) / float64(maxLevelIdx)
							writeE = -Ec * (1.0 + ratio*1.0) // -Ec to -2*Ec
						}
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
					readE := Ec * 0.3 // Well below Ec - won't switch
					step := rampRate * 0.4 * dt
					diff := readE - a.electricField
					if math.Abs(diff) < step {
						a.electricField = readE
					} else if diff > 0 {
						a.electricField += step
					} else {
						a.electricField -= step
					}
					// Capture read level and transition
					if a.wrdPhaseTimer > phaseDuration*0.3 {
						a.wrdReadLevel = a.discreteLevel + 1
						a.wrdPhase = 5
						a.wrdPhaseTimer = 0

						// Track Dr. Tour demo metrics
						a.wrdTotalWrites++
						// Success if within ±2 levels (realistic analog tolerance for 30-level memory)
						levelError := a.wrdReadLevel - a.wrdTargetLevel
						success := abs(levelError) <= 2
						if success {
							a.wrdSuccessWrites++
						}
						successRate := float64(a.wrdSuccessWrites) / float64(a.wrdTotalWrites) * 100
						log.Printf("WRD PHASE 4→5: READ done | readE=%.3f MV/cm | L_read=%d | L_target=%d | error=%+d | success=%v | rate=%.1f%% (%d/%d)",
							readE/1e8, a.wrdReadLevel, a.wrdTargetLevel, levelError, success, successRate, a.wrdSuccessWrites, a.wrdTotalWrites)

						// Update calibration using binary search with bounds tracking
						// Bounds check: ensure target is within calibration array bounds
						targetIdx := a.wrdTargetLevel - 1
						if levelError != 0 && a.calibrated && targetIdx >= 0 && targetIdx < len(a.calibrationUp) {
							midLevel := a.numLevels / 2
							// Use absolute level position to match write phase logic
							goingUp := a.wrdTargetLevel > midLevel
							if goingUp {
								// ASCENDING calibration with binary search
								currentE := a.calibrationUp[targetIdx]
								lastErr := a.lastErrorUp[targetIdx]

								if levelError > 0 {
									if currentE < a.calibUpHigh[targetIdx] {
										a.calibUpHigh[targetIdx] = currentE
									}
								} else {
									if currentE > a.calibUpLow[targetIdx] {
										a.calibUpLow[targetIdx] = currentE
									}
								}

								var newVal float64
								if a.calibUpLow[targetIdx] < a.calibUpHigh[targetIdx] {
									newVal = (a.calibUpLow[targetIdx] + a.calibUpHigh[targetIdx]) / 2
								} else {
									adjustment := float64(levelError) * Ec * 0.01
									newVal = currentE - adjustment
								}

								if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
									newVal = currentE*0.7 + newVal*0.3
								}

								if newVal < Ec*0.3 {
									newVal = Ec * 0.3
								} else if newVal > Ec*2.5 {
									newVal = Ec * 1.5
								}
								a.calibrationUp[targetIdx] = newVal
								a.lastErrorUp[targetIdx] = levelError
							} else {
								// DESCENDING calibration with binary search
								currentE := a.calibrationDown[targetIdx]
								lastErr := a.lastErrorDown[targetIdx]

								if levelError > 0 {
									if currentE > a.calibDownLow[targetIdx] {
										a.calibDownLow[targetIdx] = currentE
									}
								} else {
									if currentE < a.calibDownHigh[targetIdx] {
										a.calibDownHigh[targetIdx] = currentE
									}
								}

								var newVal float64
								if a.calibDownLow[targetIdx] < a.calibDownHigh[targetIdx] {
									newVal = (a.calibDownLow[targetIdx] + a.calibDownHigh[targetIdx]) / 2
								} else {
									adjustment := float64(levelError) * Ec * 0.01
									newVal = currentE - adjustment
								}

								if lastErr != 0 && ((lastErr > 0 && levelError < 0) || (lastErr < 0 && levelError > 0)) {
									newVal = currentE*0.7 + newVal*0.3
								}

								if newVal > -Ec*0.3 {
									newVal = -Ec * 0.3
								} else if newVal < -Ec*2.5 {
									newVal = -Ec * 1.5
								}
								a.calibrationDown[targetIdx] = newVal
								a.lastErrorDown[targetIdx] = levelError
							}
						}

						// Add accumulated energy for this cycle (calculated from E·dP integration)
						a.wrdTotalEnergyfJ += a.wrdCycleEnergy
						a.wrdCycleEnergy = 0 // Reset for next cycle

						// Log this cycle for debugging with complete phase data
						if a.wrdDebugLog != nil {
							cycle := WriteReadCycle{
								CycleNum:    len(a.wrdDebugLog.Cycles) + 1,
								TargetLevel: a.wrdTargetLevel,
								StartLevel:  a.wrdStartLevel,
								ReadLevel:   a.wrdReadLevel,
								Success:     abs(a.wrdReadLevel-a.wrdTargetLevel) <= 2,
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
										PEnd:       a.wrdReadStartP, // Non-destructive read - P doesn't change
										LevelStart: a.wrdWriteEndLvl,
										LevelEnd:   a.wrdReadLevel,
									},
								},
							}
							a.wrdDebugLog.Cycles = append(a.wrdDebugLog.Cycles, cycle)

							// Cap debug log to 100 cycles to prevent memory leak
							if len(a.wrdDebugLog.Cycles) > 100 {
								a.wrdDebugLog.Cycles = a.wrdDebugLog.Cycles[len(a.wrdDebugLog.Cycles)-100:]
							}

							// Save after every 5 cycles
							if len(a.wrdDebugLog.Cycles)%5 == 0 {
								go a.saveDebugLog()
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
						a.wrdPhase = 0
						a.wrdPhaseTimer = 0
						a.wrdCycleEnergy = 0 // Reset energy accumulator for next cycle
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

		// Record history
		a.eHistory = append(a.eHistory, a.electricField)
		a.pHistory = append(a.pHistory, a.polarization)
		if len(a.eHistory) > a.maxHistory {
			a.eHistory = a.eHistory[1:]
			a.pHistory = a.pHistory[1:]
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
	Emax := 1.5 * Ec // Go slightly beyond saturation
	numLevels := a.numLevels
	maxLevel := numLevels - 1

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

	// Initialize bounds based on temperature-corrected Ec
	for i := 0; i < numLevels; i++ {
		// Initial bounds: full range (will be narrowed by runtime feedback)
		a.calibUpLow[i] = Ec * 0.3
		a.calibUpHigh[i] = Ec * 2.0
		a.calibDownLow[i] = -Ec * 2.0
		a.calibDownHigh[i] = -Ec * 0.3
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

		// Binary search to find exact field
		lowE := Ec * 0.5
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

		// Binary search to find exact field (negative values)
		lowE := -Emax      // More negative
		highE := -Ec * 0.5 // Less negative
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

	// Reset Preisach to neutral state after calibration
	a.preisach.Reset()
	a.electricField = 0
	a.polarization = 0
	a.calibrated = true
}
