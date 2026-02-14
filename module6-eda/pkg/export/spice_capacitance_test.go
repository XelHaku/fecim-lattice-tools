package export

import (
	"math"
	"strings"
	"testing"
)

const (
	epsilon0 = 8.854187817e-12 // Vacuum permittivity (F/m)
)

// TestM6_SPICE_03_Capacitance_Formula validates C_fe calculation:
// C_fe = ε₀ × εᵣ × A / d
// Test case: HZO with εᵣ=30, A=100nm², d=10nm → expected ~26.56 fF
func TestM6_SPICE_03_Capacitance_Formula_HZO(t *testing.T) {
	// Test parameters
	epsR := 30.0           // Relative permittivity for HZO
	area_m2 := 100e-18     // 100 nm² in m²
	thickness_m := 10e-9   // 10 nm in m

	// Expected capacitance: C = ε₀ × εᵣ × A / d
	expectedCap_F := epsilon0 * epsR * area_m2 / thickness_m
	expectedCap_fF := expectedCap_F * 1e15 // Convert to fF

	// Generate FECAP_HZO subcircuit with test parameters
	mat := FeFETMaterial{
		RelativePermittivity: epsR,
		ThicknessM:           thickness_m,
		AreaM2:               area_m2,
		CoerciveFieldVM:      1e8,
	}
	subckt := GenerateFeFETSubcircuit(mat)

	// Extract capacitance value from subcircuit
	extractedCap_fF, epsRExtracted := extractCapacitanceAndEpsRFromSPICE(t, subckt)

	// Verify capacitance matches formula to < 1%
	deltaPct := math.Abs(extractedCap_fF-expectedCap_fF) / expectedCap_fF * 100.0
	if deltaPct >= 1.0 {
		t.Errorf("M6-SPICE-03: Capacitance mismatch — extracted=%.3f fF, expected=%.3f fF (formula: ε₀×εᵣ×A/d), delta=%.2f%% (tolerance: <1%%)",
			extractedCap_fF, expectedCap_fF, deltaPct)
	}

	// Verify εᵣ is correctly embedded
	if math.Abs(epsRExtracted-epsR) > 0.1 {
		t.Errorf("M6-SPICE-03: εᵣ mismatch — extracted=%.1f, expected=%.1f", epsRExtracted, epsR)
	}

	t.Logf("M6-SPICE-03 PASS: HZO capacitance validated — C_fe=%.3f fF (expected %.3f fF from εᵣ=%.1f, A=%.0f nm², d=%.1f nm), delta=%.3f%%",
		extractedCap_fF, expectedCap_fF, epsR, area_m2*1e18, thickness_m*1e9, deltaPct)
}

// TestM6_SPICE_03_Capacitance_Formula_DefaultHZO validates default HZO material
func TestM6_SPICE_03_Capacitance_Formula_DefaultHZO(t *testing.T) {
	mat := DefaultHzoFeFETMaterial()

	// Expected capacitance: C = ε₀ × εᵣ × A / d
	expectedCap_F := epsilon0 * mat.RelativePermittivity * mat.AreaM2 / mat.ThicknessM
	expectedCap_fF := expectedCap_F * 1e15

	subckt := GenerateFeFETSubcircuit(mat)
	extractedCap_fF, _ := extractCapacitanceAndEpsRFromSPICE(t, subckt)

	deltaPct := math.Abs(extractedCap_fF-expectedCap_fF) / expectedCap_fF * 100.0
	if deltaPct >= 1.0 {
		t.Errorf("M6-SPICE-03: Default HZO capacitance mismatch — extracted=%.3f fF, expected=%.3f fF, delta=%.2f%%",
			extractedCap_fF, expectedCap_fF, deltaPct)
	}

	t.Logf("M6-SPICE-03 PASS: Default HZO capacitance validated — C_fe=%.3f fF (εᵣ=%.1f, A=%.3e m², d=%.1e m), delta=%.3f%%",
		extractedCap_fF, mat.RelativePermittivity, mat.AreaM2, mat.ThicknessM, deltaPct)
}

// TestM6_SPICE_03_Capacitance_ScalingLaw validates capacitance scales correctly with εᵣ
func TestM6_SPICE_03_Capacitance_ScalingLaw_Permittivity(t *testing.T) {
	baseEpsR := 25.0
	baseMat := FeFETMaterial{
		RelativePermittivity: baseEpsR,
		ThicknessM:           10e-9,
		AreaM2:               2.025e-15,
		CoerciveFieldVM:      1e8,
	}

	baseSubckt := GenerateFeFETSubcircuit(baseMat)
	baseCap_fF, _ := extractCapacitanceAndEpsRFromSPICE(t, baseSubckt)

	// Double permittivity → double capacitance
	scaledMat := baseMat
	scaledMat.RelativePermittivity = baseEpsR * 2.0

	scaledSubckt := GenerateFeFETSubcircuit(scaledMat)
	scaledCap_fF, _ := extractCapacitanceAndEpsRFromSPICE(t, scaledSubckt)

	expectedScaledCap_fF := baseCap_fF * 2.0
	deltaPct := math.Abs(scaledCap_fF-expectedScaledCap_fF) / expectedScaledCap_fF * 100.0

	if deltaPct >= 1.0 {
		t.Errorf("M6-SPICE-03: Capacitance scaling with εᵣ failed — base=%.3f fF (εᵣ=%.1f), scaled=%.3f fF (εᵣ=%.1f), expected=%.3f fF, delta=%.2f%%",
			baseCap_fF, baseMat.RelativePermittivity, scaledCap_fF, scaledMat.RelativePermittivity, expectedScaledCap_fF, deltaPct)
	}

	t.Logf("M6-SPICE-03 PASS: Capacitance scales with εᵣ — base=%.3f fF (εᵣ=%.1f) → scaled=%.3f fF (εᵣ=%.1f, 2× ratio), delta=%.3f%%",
		baseCap_fF, baseMat.RelativePermittivity, scaledCap_fF, scaledMat.RelativePermittivity, deltaPct)
}

// TestM6_SPICE_03_Capacitance_ScalingLaw_Area validates C ∝ Area
func TestM6_SPICE_03_Capacitance_ScalingLaw_Area(t *testing.T) {
	baseMat := FeFETMaterial{
		RelativePermittivity: 25.0,
		ThicknessM:           10e-9,
		AreaM2:               100e-18, // 100 nm²
		CoerciveFieldVM:      1e8,
	}

	baseSubckt := GenerateFeFETSubcircuit(baseMat)
	baseCap_fF, _ := extractCapacitanceAndEpsRFromSPICE(t, baseSubckt)

	// 4× area → 4× capacitance
	scaledMat := baseMat
	scaledMat.AreaM2 = baseMat.AreaM2 * 4.0

	scaledSubckt := GenerateFeFETSubcircuit(scaledMat)
	scaledCap_fF, _ := extractCapacitanceAndEpsRFromSPICE(t, scaledSubckt)

	expectedScaledCap_fF := baseCap_fF * 4.0
	deltaPct := math.Abs(scaledCap_fF-expectedScaledCap_fF) / expectedScaledCap_fF * 100.0

	if deltaPct >= 1.0 {
		t.Errorf("M6-SPICE-03: Capacitance scaling with area failed — base=%.3f fF (A=%.0f nm²), scaled=%.3f fF (A=%.0f nm²), expected=%.3f fF, delta=%.2f%%",
			baseCap_fF, baseMat.AreaM2*1e18, scaledCap_fF, scaledMat.AreaM2*1e18, expectedScaledCap_fF, deltaPct)
	}

	t.Logf("M6-SPICE-03 PASS: Capacitance scales with area — base=%.3f fF (A=%.0f nm²) → scaled=%.3f fF (A=%.0f nm², 4× ratio), delta=%.3f%%",
		baseCap_fF, baseMat.AreaM2*1e18, scaledCap_fF, scaledMat.AreaM2*1e18, deltaPct)
}

// TestM6_SPICE_03_Capacitance_ScalingLaw_Thickness validates C ∝ 1/d
func TestM6_SPICE_03_Capacitance_ScalingLaw_Thickness(t *testing.T) {
	baseMat := FeFETMaterial{
		RelativePermittivity: 25.0,
		ThicknessM:           10e-9, // 10 nm
		AreaM2:               100e-18,
		CoerciveFieldVM:      1e8,
	}

	baseSubckt := GenerateFeFETSubcircuit(baseMat)
	baseCap_fF, _ := extractCapacitanceAndEpsRFromSPICE(t, baseSubckt)

	// 2× thickness → 0.5× capacitance
	scaledMat := baseMat
	scaledMat.ThicknessM = baseMat.ThicknessM * 2.0

	scaledSubckt := GenerateFeFETSubcircuit(scaledMat)
	scaledCap_fF, _ := extractCapacitanceAndEpsRFromSPICE(t, scaledSubckt)

	expectedScaledCap_fF := baseCap_fF * 0.5
	deltaPct := math.Abs(scaledCap_fF-expectedScaledCap_fF) / expectedScaledCap_fF * 100.0

	if deltaPct >= 1.0 {
		t.Errorf("M6-SPICE-03: Capacitance scaling with thickness failed — base=%.3f fF (d=%.1f nm), scaled=%.3f fF (d=%.1f nm), expected=%.3f fF, delta=%.2f%%",
			baseCap_fF, baseMat.ThicknessM*1e9, scaledCap_fF, scaledMat.ThicknessM*1e9, expectedScaledCap_fF, deltaPct)
	}

	t.Logf("M6-SPICE-03 PASS: Capacitance scales inversely with thickness — base=%.3f fF (d=%.1f nm) → scaled=%.3f fF (d=%.1f nm, 2× thickness → 0.5× cap), delta=%.3f%%",
		baseCap_fF, baseMat.ThicknessM*1e9, scaledCap_fF, scaledMat.ThicknessM*1e9, deltaPct)
}

// extractCapacitanceFromSPICE parses SPICE subcircuit and calculates capacitance from formula.
// Returns (capacitance_fF, epsR).
// SPICE format: Cfe pos n1 {eps0*epsR*area/thick}
func extractCapacitanceAndEpsRFromSPICE(t *testing.T, subckt string) (float64, float64) {
	t.Helper()
	lines := strings.Split(subckt, "\n")

	// Extract parameters from .subckt FECAP_HZO header
	var eps0Val, areaVal, thickVal float64
	var epsRVal float64 = 25.0 // default

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "*") {
			continue
		}

		// Parse .subckt FECAP_HZO pos neg PARAMS: ... eps0=... area=... thick=...
		if strings.HasPrefix(trimmed, ".subckt FECAP_HZO") && strings.Contains(trimmed, "PARAMS:") {
			eps0Val = extractSPICEParam(trimmed, "eps0")
			areaVal = extractSPICEParam(trimmed, "area")
			thickVal = extractSPICEParam(trimmed, "thick")
		}

		// Parse capacitor line: Cfe pos n1 {eps0*epsR*area/thick}
		if strings.HasPrefix(trimmed, "Cfe ") || strings.HasPrefix(trimmed, "CFE ") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 4 {
				formula := fields[3]
				// Extract εᵣ from formula: {eps0*25*area/thick}
				if strings.HasPrefix(formula, "{") && strings.HasSuffix(formula, "}") {
					formulaStr := formula[1 : len(formula)-1]
					// Parse eps0*epsR*area/thick
					parts := strings.Split(formulaStr, "*")
					if len(parts) >= 3 {
						// parts[1] is epsR
						epsRVal = parseFloatSimple(parts[1])
					}
				}
			}
		}
	}

	// Calculate capacitance: C = eps0 * epsR * area / thick
	if eps0Val > 0 && areaVal > 0 && thickVal > 0 {
		capF := eps0Val * epsRVal * areaVal / thickVal
		return capF * 1e15, epsRVal
	}

	return 0.0, epsRVal
}

// extractSPICEParam extracts a parameter value from a SPICE PARAMS: line
func extractSPICEParam(line, paramName string) float64 {
	// Look for paramName=value
	paramKey := paramName + "="
	if idx := strings.Index(line, paramKey); idx != -1 {
		start := idx + len(paramKey)
		end := start
		for end < len(line) && (line[end] != ' ' && line[end] != '\t') {
			end++
		}
		valStr := line[start:end]
		return parseFloatSimple(valStr)
	}
	return 0.0
}

// parseFloatSimple parses a simple float (including exponential notation)
func parseFloatSimple(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0.0
	}

	// Handle exponential notation: 1.234e-15
	var mantissa, exponent float64
	var sign float64 = 1.0

	// Split on 'e' or 'E'
	parts := strings.Split(s, "e")
	if len(parts) == 1 {
		parts = strings.Split(s, "E")
	}

	// Parse mantissa
	mStr := parts[0]
	if strings.HasPrefix(mStr, "-") {
		sign = -1.0
		mStr = mStr[1:]
	} else if strings.HasPrefix(mStr, "+") {
		mStr = mStr[1:]
	}

	dotPos := strings.Index(mStr, ".")
	if dotPos == -1 {
		// Integer mantissa
		for _, c := range mStr {
			if c >= '0' && c <= '9' {
				mantissa = mantissa*10 + float64(c-'0')
			}
		}
	} else {
		// Decimal mantissa
		intPart := mStr[:dotPos]
		fracPart := mStr[dotPos+1:]
		
		for _, c := range intPart {
			if c >= '0' && c <= '9' {
				mantissa = mantissa*10 + float64(c-'0')
			}
		}
		
		frac := 0.0
		divisor := 1.0
		for _, c := range fracPart {
			if c >= '0' && c <= '9' {
				divisor *= 10
				frac = frac*10 + float64(c-'0')
			}
		}
		mantissa += frac / divisor
	}

	// Parse exponent (if present)
	if len(parts) >= 2 {
		expStr := parts[1]
		expSign := 1.0
		if strings.HasPrefix(expStr, "-") {
			expSign = -1.0
			expStr = expStr[1:]
		} else if strings.HasPrefix(expStr, "+") {
			expStr = expStr[1:]
		}
		
		for _, c := range expStr {
			if c >= '0' && c <= '9' {
				exponent = exponent*10 + float64(c-'0')
			}
		}
		exponent *= expSign
	}

	result := sign * mantissa * math.Pow(10, exponent)
	return result
}
