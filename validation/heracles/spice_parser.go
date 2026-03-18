package heracles

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SpicePEData holds the parsed ascending and descending branch data from
// an ngspice simulation output, expressed in the same units as PEPoint
// (E in MV/cm, P in uC/cm^2).
type SpicePEData struct {
	Ascending  []PEPoint
	Descending []PEPoint
}

// ParseNgspiceOutput parses ngspice print/CSV output containing voltage and
// current data from a Heracles FeCap DC sweep simulation.
//
// The parser looks for tabular data with columns for voltage (V) and current (A).
// It converts charge (integrated current or raw charge values) into polarization
// density (C/m^2) using the provided area (m^2), then converts to the PEPoint
// convention: E in MV/cm, P in uC/cm^2.
//
// thicknessM is the film thickness in meters (used for V -> E conversion).
// areaM2 is the capacitor area in m^2 (used for Q -> P conversion).
//
// The raw output is expected to contain lines like:
//
//	Index   asc_v           asc_i
//	0       -3.000000e+00   -1.234567e-06
//	1       -2.970000e+00   -1.200000e-06
//	...
//
// or ngspice "print" format with similar structure.
func ParseNgspiceOutput(raw string, thicknessM, areaM2 float64) (*SpicePEData, error) {
	if thicknessM <= 0 {
		return nil, fmt.Errorf("thickness must be positive, got %e", thicknessM)
	}
	if areaM2 <= 0 {
		return nil, fmt.Errorf("area must be positive, got %e", areaM2)
	}

	lines := strings.Split(raw, "\n")

	// Detect whether the output is comma-separated (CSV) by checking for
	// data lines containing commas.  CSV format is tried first when detected
	// because the whitespace-based print parser would incorrectly merge
	// comma-separated numeric fields.
	hasCSV := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, ",") {
			parts := strings.Split(trimmed, ",")
			if len(parts) >= 2 {
				hasCSV = true
				break
			}
		}
	}

	if hasCSV {
		// Try CSV first.
		asc, desc, err := parseCSVFormat(lines, thicknessM, areaM2)
		if err == nil && (len(asc) > 0 || len(desc) > 0) {
			return &SpicePEData{Ascending: asc, Descending: desc}, nil
		}
	}

	// Try ngspice tabular print output (whitespace-separated).
	asc, desc, err := parseNgspicePrintFormat(lines, thicknessM, areaM2)
	if err == nil && (len(asc) > 0 || len(desc) > 0) {
		return &SpicePEData{Ascending: asc, Descending: desc}, nil
	}

	// Last resort: try CSV even if comma detection missed it.
	if !hasCSV {
		asc, desc, err = parseCSVFormat(lines, thicknessM, areaM2)
		if err == nil && (len(asc) > 0 || len(desc) > 0) {
			return &SpicePEData{Ascending: asc, Descending: desc}, nil
		}
	}

	return nil, fmt.Errorf("no parseable voltage/current data found in ngspice output")
}

// numericFieldRe matches a floating point number in scientific or decimal notation.
var numericFieldRe = regexp.MustCompile(`[+-]?[0-9]*\.?[0-9]+(?:[eE][+-]?\d+)?`)

// parseNgspicePrintFormat parses the standard ngspice "print" output format.
// It looks for sections containing voltage and current columns and extracts
// ascending and descending branch data.
func parseNgspicePrintFormat(lines []string, thicknessM, areaM2 float64) ([]PEPoint, []PEPoint, error) {
	type dataBlock struct {
		voltages []float64
		currents []float64
	}

	var blocks []dataBlock
	var current *dataBlock

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if current != nil && len(current.voltages) > 0 {
				blocks = append(blocks, *current)
				current = nil
			}
			continue
		}

		// Skip header lines and separator lines.
		if strings.HasPrefix(trimmed, "Index") || strings.HasPrefix(trimmed, "---") ||
			strings.HasPrefix(trimmed, "No.") || strings.HasPrefix(trimmed, "x") {
			if current != nil && len(current.voltages) > 0 {
				blocks = append(blocks, *current)
			}
			current = &dataBlock{}
			continue
		}

		// Try to extract numeric fields from the line.
		fields := numericFieldRe.FindAllString(trimmed, -1)
		if len(fields) < 2 {
			continue
		}

		// We need at least an index + voltage + current (3 fields)
		// or voltage + current (2 fields).
		var vIdx, iIdx int
		if len(fields) >= 3 {
			// Assume: index, voltage, current
			vIdx = 1
			iIdx = 2
		} else {
			// Assume: voltage, current
			vIdx = 0
			iIdx = 1
		}

		voltage, err := strconv.ParseFloat(fields[vIdx], 64)
		if err != nil {
			continue
		}
		currentA, err := strconv.ParseFloat(fields[iIdx], 64)
		if err != nil {
			continue
		}

		if current == nil {
			current = &dataBlock{}
		}
		current.voltages = append(current.voltages, voltage)
		current.currents = append(current.currents, currentA)
	}
	if current != nil && len(current.voltages) > 0 {
		blocks = append(blocks, *current)
	}

	if len(blocks) == 0 {
		return nil, nil, fmt.Errorf("no data blocks found")
	}

	convertBlock := func(blk dataBlock) []PEPoint {
		pts := make([]PEPoint, 0, len(blk.voltages))
		for i := range blk.voltages {
			v := blk.voltages[i]
			iA := blk.currents[i]

			// Convert voltage to electric field: E = V / thickness
			// then to MV/cm: 1 V/m = 1e-8 MV/cm
			eMVcm := (v / thicknessM) * 1e-8

			// Current from DC sweep is displacement current dQ/dV * dV.
			// For a ferroelectric capacitor in DC sweep, the "current" effectively
			// represents charge displacement.  Convert to polarization density:
			// P = Q / Area = I / Area  (where I here is charge-like from DC)
			// Then to uC/cm^2: 1 C/m^2 = 100 uC/cm^2
			pUCcm2 := (iA / areaM2) * 100.0

			pts = append(pts, PEPoint{E_MVcm: eMVcm, P_uCcm: pUCcm2})
		}
		return pts
	}

	// First block is ascending, second (if present) is descending.
	var asc, desc []PEPoint
	if len(blocks) >= 1 {
		asc = convertBlock(blocks[0])
	}
	if len(blocks) >= 2 {
		desc = convertBlock(blocks[1])
	}

	return asc, desc, nil
}

// parseCSVFormat attempts to parse comma-separated value lines.
func parseCSVFormat(lines []string, thicknessM, areaM2 float64) ([]PEPoint, []PEPoint, error) {
	var asc, desc []PEPoint
	seenHeader := false
	branchCol := -1
	vCol := -1
	iCol := -1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		fields := strings.Split(trimmed, ",")
		if len(fields) < 2 {
			continue
		}

		// Detect header row.
		if !seenHeader {
			lower := strings.ToLower(trimmed)
			if strings.Contains(lower, "branch") || strings.Contains(lower, "voltage") ||
				strings.Contains(lower, "v(") || strings.Contains(lower, "current") {
				for i, f := range fields {
					fl := strings.ToLower(strings.TrimSpace(f))
					if strings.Contains(fl, "branch") {
						branchCol = i
					}
					if strings.Contains(fl, "voltage") || strings.Contains(fl, "v(") || fl == "v" {
						vCol = i
					}
					if strings.Contains(fl, "current") || strings.Contains(fl, "i(") || strings.Contains(fl, "i") {
						iCol = i
					}
				}
				seenHeader = true
				continue
			}
		}

		// Parse data row.
		if vCol < 0 && len(fields) >= 3 {
			// Default layout: branch, voltage, current
			branchCol = 0
			vCol = 1
			iCol = 2
		}
		if vCol < 0 && len(fields) >= 2 {
			vCol = 0
			iCol = 1
		}
		if vCol < 0 || iCol < 0 || vCol >= len(fields) || iCol >= len(fields) {
			continue
		}

		v, err := strconv.ParseFloat(strings.TrimSpace(fields[vCol]), 64)
		if err != nil {
			continue
		}
		iA, err := strconv.ParseFloat(strings.TrimSpace(fields[iCol]), 64)
		if err != nil {
			continue
		}

		eMVcm := (v / thicknessM) * 1e-8
		pUCcm2 := (iA / areaM2) * 100.0
		pt := PEPoint{E_MVcm: eMVcm, P_uCcm: pUCcm2}

		branch := "asc"
		if branchCol >= 0 && branchCol < len(fields) {
			branch = strings.ToLower(strings.TrimSpace(fields[branchCol]))
		}

		switch branch {
		case "desc", "descending":
			desc = append(desc, pt)
		default:
			asc = append(asc, pt)
		}
	}

	if len(asc) == 0 && len(desc) == 0 {
		return nil, nil, fmt.Errorf("no CSV data rows found")
	}
	return asc, desc, nil
}
