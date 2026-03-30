package literature

// external_benchmarks_test.go validates simulator material presets against
// published measured data. Each test cites a specific DOI and checks that
// our default parameter falls within the reported experimental range.
//
// Methodology: range-check (value within [min, max] from the paper).
// This is the simplest defensible validation: if a parameter is within the
// published measured spread, we do not claim accuracy but we do confirm
// the simulator is not divergent from reality.

import (
	"math"
	"testing"

	physics "fecim-lattice-tools/shared/physics"
)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// checkRange fails the test if val is outside [lo, hi] and logs the comparison.
func checkRange(t *testing.T, name string, val, lo, hi float64, unit, doi string) {
	t.Helper()
	t.Logf("%-30s  val=%.4g  range=[%.4g, %.4g] %s  DOI:%s", name, val, lo, hi, unit, doi)
	if val < lo || val > hi {
		t.Errorf("FAIL %s: %.4g outside [%.4g, %.4g] %s (DOI: %s)", name, val, lo, hi, unit, doi)
	}
}

// ---------------------------------------------------------------------------
// 1. DefaultHZO() Pr against Park 2015
// ---------------------------------------------------------------------------

// TestDefaultHZO_Pr_Park2015 validates that DefaultHZO remnant polarization
// falls within the experimentally measured range for 10nm ALD Hf0.5Zr0.5O2.
//
// Source: Park et al., Adv. Mater. 27, 1811 (2015)
// DOI: 10.1002/adma.201404531
// Table 1 / Fig. 3: Pr ranges from 15 to 34 uC/cm^2 across doping and
// anneal conditions for 10nm HZO films.
func TestDefaultHZO_Pr_Park2015(t *testing.T) {
	mat := physics.DefaultHZO()
	if mat == nil {
		t.Fatal("DefaultHZO returned nil")
	}
	// Pr is stored in C/m^2; convert to uC/cm^2 for comparison.
	// 1 C/m^2 = 100 uC/cm^2
	prUCcm2 := mat.Pr * 100.0

	checkRange(t, "DefaultHZO.Pr", prUCcm2, 15.0, 34.0,
		"uC/cm^2", "10.1002/adma.201404531")
}

// ---------------------------------------------------------------------------
// 2. DefaultHZO() Ec against Park 2015
// ---------------------------------------------------------------------------

// TestDefaultHZO_Ec_Park2015 validates that DefaultHZO coercive field falls
// within the experimentally measured range for 10nm HZO.
//
// Source: Park et al., Adv. Mater. 27, 1811 (2015)
// DOI: 10.1002/adma.201404531
// Fig. 3 / Table 1: Ec ranges from 0.8 to 1.5 MV/cm across compositions.
func TestDefaultHZO_Ec_Park2015(t *testing.T) {
	mat := physics.DefaultHZO()
	if mat == nil {
		t.Fatal("DefaultHZO returned nil")
	}
	// Ec is stored in V/m; convert to MV/cm.
	// 1 MV/cm = 1e8 V/m
	ecMVcm := mat.Ec / 1e8

	checkRange(t, "DefaultHZO.Ec", ecMVcm, 0.8, 1.5,
		"MV/cm", "10.1002/adma.201404531")
}

// ---------------------------------------------------------------------------
// 3. FeCIMMaterial() Pr against FeCIM spec range
// ---------------------------------------------------------------------------

// TestFeCIMMaterial_Pr_Range validates that the FeCIM material preset Pr
// falls within the broader HZO literature range. FeCIM Pr is labeled
// ESTIMATED in the codebase; it should at least be within physically
// plausible HZO bounds.
//
// Source: Park 2015 (DOI: 10.1002/adma.201404531) gives 15-34 uC/cm^2
// for standard HZO. FeCIM targets optimized devices so we allow a slightly
// wider upper bound (40 uC/cm^2) per Cheema et al., Nature 580, 478 (2020),
// DOI: 10.1038/s41586-020-2208-x (reports >40 in ultrathin stacks).
func TestFeCIMMaterial_Pr_Range(t *testing.T) {
	mat := physics.FeCIMMaterial()
	if mat == nil {
		t.Fatal("FeCIMMaterial returned nil")
	}
	prUCcm2 := mat.Pr * 100.0

	checkRange(t, "FeCIMMaterial.Pr", prUCcm2, 15.0, 40.0,
		"uC/cm^2", "10.1002/adma.201404531")
}

// ---------------------------------------------------------------------------
// 4. DefaultHZO() EnduranceCycles against literature range
// ---------------------------------------------------------------------------

// TestDefaultHZO_Endurance_LiteratureRange validates that the default
// endurance value is within the experimentally demonstrated range for HZO.
//
// Lower bound: 10^9 cycles (conservative baseline for thin-film HZO).
//   Source: Mueller et al., IEEE IEDM 2013, DOI: 10.1109/IEDM.2013.6724605
//
// Upper bound: 10^12 cycles (best-in-class HZO+MoS2 hybrid).
//   Source: Science Advances 2024, DOI: 10.1126/sciadv.adp0174
func TestDefaultHZO_Endurance_LiteratureRange(t *testing.T) {
	mat := physics.DefaultHZO()
	if mat == nil {
		t.Fatal("DefaultHZO returned nil")
	}

	logEndurance := math.Log10(mat.EnduranceCycles)
	checkRange(t, "DefaultHZO.Endurance", logEndurance, 9.0, 12.0,
		"log10(cycles)", "10.1109/IEDM.2013.6724605")
}

// ---------------------------------------------------------------------------
// 5. C2C variation against Soliman 2023
// ---------------------------------------------------------------------------

// TestDefaultHZO_C2C_Soliman2023 validates cycle-to-cycle variation against
// the measured Vth sigma in a 28nm HKMG 4-state FeFET crossbar.
//
// Source: Soliman et al., Nature Communications 14, 6348 (2023)
// DOI: 10.1038/s41467-023-42110-y
// Measured: sigma_Vth = 38-40 mV for 4 states, corresponding to ~3-5% C2C.
// Our default: C2CSigmaBase = 0.03 (3%).
//
// We accept 1% to 8% as the plausible range: 1% is the best reported for
// tightly controlled devices; 8% covers multi-level arrays with >8 states.
func TestDefaultHZO_C2C_Soliman2023(t *testing.T) {
	mat := physics.DefaultHZO()
	if mat == nil {
		t.Fatal("DefaultHZO returned nil")
	}

	c2cPct := mat.C2CSigmaBase * 100.0 // convert fraction to percent
	checkRange(t, "DefaultHZO.C2CSigmaBase", c2cPct, 1.0, 8.0,
		"%", "10.1038/s41467-023-42110-y")
}

// ---------------------------------------------------------------------------
// 6. FeCIMMaterial() Endurance (demonstrated) against literature
// ---------------------------------------------------------------------------

// TestFeCIMMaterial_Endurance_Demonstrated validates that the FeCIM
// demonstrated endurance (10^9) is within the broader literature range.
//
// Source: Mueller et al., IEEE IEDM 2013 (lower bound 10^9)
//   DOI: 10.1109/IEDM.2013.6724605
// Source: Science Advances 2024 (upper bound 10^12)
//   DOI: 10.1126/sciadv.adp0174
func TestFeCIMMaterial_Endurance_Demonstrated(t *testing.T) {
	mat := physics.FeCIMMaterial()
	if mat == nil {
		t.Fatal("FeCIMMaterial returned nil")
	}

	logEndurance := math.Log10(mat.EnduranceCycles)
	checkRange(t, "FeCIM.Endurance(demo)", logEndurance, 9.0, 12.0,
		"log10(cycles)", "10.1109/IEDM.2013.6724605")
}

// ---------------------------------------------------------------------------
// 7. MaterlikHfO2() Pr against Materlik 2015 operating target
// ---------------------------------------------------------------------------

// TestMaterlikHfO2_Pr_Materlik2015 validates the Materlik preset Pr against
// the operating target stated in the preset docstring (Pr ~ 20 uC/cm^2).
// The paper reports Pr in the range of 10-30 uC/cm^2 depending on LGD
// coefficients and thickness.
//
// Source: Materlik et al., J. Appl. Phys. 117, 134109 (2015)
// DOI: 10.1063/1.4916229
func TestMaterlikHfO2_Pr_Materlik2015(t *testing.T) {
	mat := physics.MaterlikHfO2()
	if mat == nil {
		t.Fatal("MaterlikHfO2 returned nil")
	}
	prUCcm2 := mat.Pr * 100.0

	checkRange(t, "MaterlikHfO2.Pr", prUCcm2, 10.0, 30.0,
		"uC/cm^2", "10.1063/1.4916229")
}

// ---------------------------------------------------------------------------
// 8. FeCIMMaterial() C2C against Soliman 2023
// ---------------------------------------------------------------------------

// TestFeCIMMaterial_C2C_Soliman2023 validates the FeCIM C2C variation
// against the same Soliman 2023 benchmark used for DefaultHZO.
//
// Source: Soliman et al., Nature Communications 14, 6348 (2023)
// DOI: 10.1038/s41467-023-42110-y
func TestFeCIMMaterial_C2C_Soliman2023(t *testing.T) {
	mat := physics.FeCIMMaterial()
	if mat == nil {
		t.Fatal("FeCIMMaterial returned nil")
	}

	c2cPct := mat.C2CSigmaBase * 100.0
	checkRange(t, "FeCIM.C2CSigmaBase", c2cPct, 1.0, 8.0,
		"%", "10.1038/s41467-023-42110-y")
}

// ---------------------------------------------------------------------------
// 9. DefaultHZO() LGD coefficients against Materlik 2015
// ---------------------------------------------------------------------------

// TestDefaultHZO_LGD_Materlik2015 validates that DefaultHZO uses Materlik
// 2015 LGD beta/gamma coefficients exactly (these are not ranges but exact
// published values).
//
// Source: Materlik et al., J. Appl. Phys. 117, 134109 (2015)
// DOI: 10.1063/1.4916229
// Table I: beta = -6.72e8 J*m^5/C^4, gamma = 1.95e10 J*m^9/C^6
func TestDefaultHZO_LGD_Materlik2015(t *testing.T) {
	mat := physics.DefaultHZO()
	if mat == nil {
		t.Fatal("DefaultHZO returned nil")
	}

	const doi = "10.1063/1.4916229"

	// Beta: exact match within floating-point tolerance
	betaExpected := -6.720e8
	betaRelErr := math.Abs(mat.BetaLandau-betaExpected) / math.Abs(betaExpected)
	t.Logf("%-30s  val=%.4e  expected=%.4e  relErr=%.2e  DOI:%s",
		"DefaultHZO.BetaLandau", mat.BetaLandau, betaExpected, betaRelErr, doi)
	if betaRelErr > 1e-6 {
		t.Errorf("FAIL BetaLandau: %.4e != %.4e (relErr=%.2e > 1e-6)", mat.BetaLandau, betaExpected, betaRelErr)
	}

	// Gamma: exact match within floating-point tolerance
	gammaExpected := 1.950e10
	gammaRelErr := math.Abs(mat.GammaLandau-gammaExpected) / math.Abs(gammaExpected)
	t.Logf("%-30s  val=%.4e  expected=%.4e  relErr=%.2e  DOI:%s",
		"DefaultHZO.GammaLandau", mat.GammaLandau, gammaExpected, gammaRelErr, doi)
	if gammaRelErr > 1e-6 {
		t.Errorf("FAIL GammaLandau: %.4e != %.4e (relErr=%.2e > 1e-6)", mat.GammaLandau, gammaExpected, gammaRelErr)
	}
}
