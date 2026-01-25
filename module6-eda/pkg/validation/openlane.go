package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/openlane"
)

// PlacementResult contains the results of placement validation
type PlacementResult struct {
	Passed         bool
	ViolationCount int
	Violations     []PlacementViolation
	RawOutput      string
}

// PlacementViolation represents a single placement issue
type PlacementViolation struct {
	Cell     string
	Issue    string // "overlap", "out_of_bounds", "unplaced"
	Location string // "x,y"
	Message  string
}

// CellUsageResult contains cell usage statistics
type CellUsageResult struct {
	TotalCells     int
	CellTypes      map[string]int
	UtilizationPct float64
	RawOutput      string
}

// checkPlacementScript is the TCL script for placement validation
// Note: FeCIM crossbar is clockless - no STA, just placement check
const checkPlacementScript = `# check_placement.tcl - For clockless FeCIM crossbar
# NOTE: No STA since design has no clock

read_lef $env(TECH_LEF)
read_lef $env(CELL_LEF)
read_def $env(DEF_FILE)

# Placement validation (this is the DRC for placement)
puts "=== PLACEMENT CHECK ==="
check_placement -verbose

# Cell usage report
puts "=== CELL USAGE ==="
report_cell_usage

# Design summary
puts "=== DESIGN SUMMARY ==="
report_design

exit
`

// RunPlacementCheck validates placement using OpenROAD
func RunPlacementCheck(defPath string, manager *openlane.Manager, config *openlane.Config) (*PlacementResult, error) {
	// Check if DEF file exists
	if _, err := os.Stat(defPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("DEF file not found: %s", defPath)
	}

	// Check mode
	mode := manager.DetectMode()
	if mode == openlane.ModeNone {
		return nil, fmt.Errorf("OpenROAD not available (install Docker with OpenLane image or native OpenROAD)")
	}

	// Create work directory with script
	workDir := filepath.Dir(defPath)
	scriptPath := filepath.Join(workDir, "check_placement.tcl")
	if err := os.WriteFile(scriptPath, []byte(checkPlacementScript), 0644); err != nil {
		return nil, fmt.Errorf("failed to write TCL script: %v", err)
	}
	defer os.Remove(scriptPath)

	// Set up environment variables
	envVars := map[string]string{
		"DEF_FILE": "/design/" + filepath.Base(defPath),
	}

	// Run OpenROAD
	runner := openlane.NewRunner(manager, config)
	result, err := runner.RunOpenROAD("check_placement.tcl", workDir, envVars)

	placementResult := &PlacementResult{
		Passed:     true,
		Violations: []PlacementViolation{},
	}

	if result != nil {
		placementResult.RawOutput = result.Stdout + "\n" + result.Stderr
	}

	if err != nil {
		placementResult.Passed = false
		placementResult.Violations = append(placementResult.Violations, PlacementViolation{
			Issue:   "error",
			Message: err.Error(),
		})
		return placementResult, nil // Return result with error info, not the error
	}

	// Parse output for violations
	placementResult.Violations, placementResult.ViolationCount = parsePlacementOutput(placementResult.RawOutput)
	placementResult.Passed = placementResult.ViolationCount == 0

	return placementResult, nil
}

// parsePlacementOutput extracts violations from OpenROAD output
func parsePlacementOutput(output string) ([]PlacementViolation, int) {
	var violations []PlacementViolation

	// Look for common placement issues
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for overlap errors
		if strings.Contains(strings.ToLower(line), "overlap") {
			violations = append(violations, PlacementViolation{
				Issue:   "overlap",
				Message: line,
			})
		}

		// Check for placement errors
		if strings.Contains(strings.ToLower(line), "error") && strings.Contains(strings.ToLower(line), "placement") {
			violations = append(violations, PlacementViolation{
				Issue:   "placement_error",
				Message: line,
			})
		}

		// Check for unplaced cells
		if strings.Contains(strings.ToLower(line), "unplaced") {
			violations = append(violations, PlacementViolation{
				Issue:   "unplaced",
				Message: line,
			})
		}
	}

	return violations, len(violations)
}

// RunCellUsageReport gets cell usage statistics
func RunCellUsageReport(defPath string, manager *openlane.Manager, config *openlane.Config) (*CellUsageResult, error) {
	// Run placement check which includes cell usage
	placementResult, err := RunPlacementCheck(defPath, manager, config)
	if err != nil {
		return nil, err
	}

	// Parse cell usage from output
	return parseCellUsageOutput(placementResult.RawOutput), nil
}

// parseCellUsageOutput extracts cell usage from OpenROAD output
func parseCellUsageOutput(output string) *CellUsageResult {
	result := &CellUsageResult{
		CellTypes: make(map[string]int),
		RawOutput: output,
	}

	lines := strings.Split(output, "\n")
	inCellUsage := false

	// Regex for cell usage lines: "cell_name count"
	cellRe := regexp.MustCompile(`^\s*(\S+)\s+(\d+)\s*$`)
	totalRe := regexp.MustCompile(`Total\s+(\d+)`)
	utilRe := regexp.MustCompile(`Utilization\s*[:\s]+(\d+\.?\d*)\s*%?`)

	for _, line := range lines {
		if strings.Contains(line, "CELL USAGE") {
			inCellUsage = true
			continue
		}
		if strings.Contains(line, "===") && inCellUsage {
			inCellUsage = false
			continue
		}

		if inCellUsage {
			if matches := cellRe.FindStringSubmatch(line); matches != nil {
				count, _ := strconv.Atoi(matches[2])
				result.CellTypes[matches[1]] = count
				result.TotalCells += count
			}
		}

		// Look for total cell count
		if matches := totalRe.FindStringSubmatch(line); matches != nil {
			result.TotalCells, _ = strconv.Atoi(matches[1])
		}

		// Look for utilization
		if matches := utilRe.FindStringSubmatch(line); matches != nil {
			result.UtilizationPct, _ = strconv.ParseFloat(matches[1], 64)
		}
	}

	return result
}

// IsOpenROADAvailable checks if OpenROAD validation is available
func IsOpenROADAvailable(manager *openlane.Manager) bool {
	return manager.DetectMode() != openlane.ModeNone
}
