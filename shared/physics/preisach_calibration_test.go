package physics

import (
	"math"
	"path/filepath"
	"testing"
)

// TestCalibratePreisachToData_Park2015 loads the Park 2015 HZO digitized P-E
// loop, runs the golden-section calibration, and verifies RMSE < 5% of Ps.
func TestCalibratePreisachToData_Park2015(t *testing.T) {
	csvPath := filepath.Join("..", "..", "validation", "literature", "data", "park2015_fig2a_hzo_10nm.csv")
	data, err := LoadPELoopCSV(csvPath)
	if err != nil {
		t.Fatalf("LoadPELoopCSV: %v", err)
	}
	if len(data) < 6 {
		t.Fatalf("expected >= 6 points, got %d", len(data))
	}

	result, err := CalibratePreisachToData(data)
	if err != nil {
		t.Fatalf("CalibratePreisachToData: %v", err)
	}

	t.Logf("Calibrated: Ps=%.4f C/m^2 (%.2f uC/cm^2), Ec=%.4e V/m (%.3f MV/cm), Delta=%.4e V/m (%.3f MV/cm)",
		result.Ps, result.Ps*100, result.Ec, result.Ec/1e8, result.Delta, result.Delta/1e8)
	t.Logf("RMSE=%.6f C/m^2 (%.4f uC/cm^2), RMSE/Ps=%.4f%%",
		result.RMSE, result.RMSE*100, result.RMSE/result.Ps*100)

	// Key assertion: RMSE must be less than 5% of Ps.
	rmseFrac := result.RMSE / result.Ps
	if rmseFrac >= 0.05 {
		t.Fatalf("RMSE/Ps = %.4f%% >= 5%%; calibration quality insufficient", rmseFrac*100)
	}

	// Sanity checks on extracted parameters.
	if result.Ec <= 0 {
		t.Errorf("Ec should be positive, got %e", result.Ec)
	}
	if result.Ps <= 0 {
		t.Errorf("Ps should be positive, got %e", result.Ps)
	}
	if result.Delta <= 0 {
		t.Errorf("Delta should be positive, got %e", result.Delta)
	}
	// Ec should be in the ballpark of 1 MV/cm for Park 2015 HZO.
	ecMVcm := result.Ec / 1e8
	if ecMVcm < 0.5 || ecMVcm > 2.0 {
		t.Errorf("Ec = %.3f MV/cm outside expected [0.5, 2.0] range", ecMVcm)
	}
}

// TestCalibratePreisachToData_TooFewPoints verifies the minimum-data guard.
func TestCalibratePreisachToData_TooFewPoints(t *testing.T) {
	data := []PEPoint{
		{Field_Vm: -1e8, Polarization_Cm: -0.15},
		{Field_Vm: 0, Polarization_Cm: 0},
		{Field_Vm: 1e8, Polarization_Cm: 0.15},
	}
	_, err := CalibratePreisachToData(data)
	if err == nil {
		t.Fatal("expected error for < 6 points, got nil")
	}
}

// TestCalibratePreisachToData_SyntheticRoundTrip generates a P-E loop from a
// known TanhEverett, calibrates back, and checks that parameters are recovered.
func TestCalibratePreisachToData_SyntheticRoundTrip(t *testing.T) {
	// Generate synthetic data with known parameters.
	truePs := 0.20     // C/m^2
	trueEc := 1.0e8    // V/m
	trueDelta := 0.4e8 // V/m

	ev := &TanhEverett{Ps: truePs, Ec: trueEc, Delta: trueDelta}
	satE := 3.0e8
	stack := NewPreisachStack(satE, ev)

	// Pre-condition to negative saturation.
	stack.Update(-satE)

	// Generate ascending + descending loop.
	nPts := 40
	data := make([]PEPoint, 0, 2*nPts)
	for i := 0; i <= nPts; i++ {
		e := -satE + 2*satE*float64(i)/float64(nPts)
		p := stack.Update(e)
		data = append(data, PEPoint{Field_Vm: e, Polarization_Cm: p})
	}
	for i := 0; i <= nPts; i++ {
		e := satE - 2*satE*float64(i)/float64(nPts)
		p := stack.Update(e)
		data = append(data, PEPoint{Field_Vm: e, Polarization_Cm: p})
	}

	result, err := CalibratePreisachToData(data)
	if err != nil {
		t.Fatalf("CalibratePreisachToData: %v", err)
	}

	t.Logf("True:  Ps=%.4f Ec=%.4e Delta=%.4e", truePs, trueEc, trueDelta)
	t.Logf("Fit:   Ps=%.4f Ec=%.4e Delta=%.4e RMSE=%.6f", result.Ps, result.Ec, result.Delta, result.RMSE)

	// Ps should be recovered exactly (it's the max |P| in the data).
	if relErr(result.Ps, truePs) > 0.02 {
		t.Errorf("Ps recovery: got %.6f, want ~%.6f (relErr=%.4f%%)", result.Ps, truePs, relErr(result.Ps, truePs)*100)
	}

	// Ec should be close to the true value.
	if relErr(result.Ec, trueEc) > 0.05 {
		t.Errorf("Ec recovery: got %.4e, want ~%.4e (relErr=%.4f%%)", result.Ec, trueEc, relErr(result.Ec, trueEc)*100)
	}

	// Delta should be reasonably close.
	if relErr(result.Delta, trueDelta) > 0.10 {
		t.Errorf("Delta recovery: got %.4e, want ~%.4e (relErr=%.4f%%)", result.Delta, trueDelta, relErr(result.Delta, trueDelta)*100)
	}

	// RMSE should be very small for a round-trip.
	if result.RMSE/truePs > 0.01 {
		t.Errorf("RMSE/Ps = %.4f%%, expected < 1%% for round-trip", result.RMSE/truePs*100)
	}
}

// TestLoadPELoopCSV_Park2015 validates the CSV loader with the actual data file.
func TestLoadPELoopCSV_Park2015(t *testing.T) {
	csvPath := filepath.Join("..", "..", "validation", "literature", "data", "park2015_fig2a_hzo_10nm.csv")
	data, err := LoadPELoopCSV(csvPath)
	if err != nil {
		t.Fatalf("LoadPELoopCSV: %v", err)
	}
	// The CSV has 61 data rows (1 header + 61 data lines = 62 total lines).
	if len(data) != 61 {
		t.Errorf("expected 61 data points, got %d", len(data))
	}
	// First point should be E=-3 MV/cm = -3e8 V/m, P ~ -19.38 uC/cm^2 = -0.1938 C/m^2.
	if math.Abs(data[0].Field_Vm-(-3e8)) > 1e5 {
		t.Errorf("first E = %e, expected ~ -3e8", data[0].Field_Vm)
	}
	if math.Abs(data[0].Polarization_Cm-(-19.38e-2)) > 1e-3 {
		t.Errorf("first P = %e, expected ~ -0.1938", data[0].Polarization_Cm)
	}
}

// TestSplitAscendingDescending verifies branch splitting.
func TestSplitAscendingDescending(t *testing.T) {
	data := []PEPoint{
		{Field_Vm: -2e8, Polarization_Cm: -0.15},
		{Field_Vm: -1e8, Polarization_Cm: -0.10},
		{Field_Vm: 0, Polarization_Cm: 0.05},
		{Field_Vm: 1e8, Polarization_Cm: 0.12},
		{Field_Vm: 2e8, Polarization_Cm: 0.18},
		{Field_Vm: 1e8, Polarization_Cm: 0.14},
		{Field_Vm: 0, Polarization_Cm: -0.03},
		{Field_Vm: -1e8, Polarization_Cm: -0.11},
		{Field_Vm: -2e8, Polarization_Cm: -0.16},
	}
	asc, desc := SplitAscendingDescending(data)
	if len(asc) != 5 {
		t.Errorf("ascending: expected 5 points, got %d", len(asc))
	}
	if len(desc) != 5 {
		t.Errorf("descending: expected 5 points, got %d", len(desc))
	}
	// Ascending should end at maxE, descending should start at maxE.
	if asc[len(asc)-1].Field_Vm != 2e8 {
		t.Errorf("ascending last E = %e, expected 2e8", asc[len(asc)-1].Field_Vm)
	}
	if desc[0].Field_Vm != 2e8 {
		t.Errorf("descending first E = %e, expected 2e8", desc[0].Field_Vm)
	}
}

// TestGoldenSectionMinimize verifies convergence on a known parabola.
func TestGoldenSectionMinimize(t *testing.T) {
	// f(x) = (x-3)^2, minimum at x=3.
	x, fv := goldenSectionMinimize(0, 10, func(x float64) float64 {
		return (x - 3) * (x - 3)
	}, 50)
	if math.Abs(x-3) > 1e-10 {
		t.Errorf("minimum at x=%.12f, expected 3.0", x)
	}
	if fv > 1e-20 {
		t.Errorf("f(xMin)=%.12e, expected ~0", fv)
	}
}

// TestPark2015PreisachEverettPreset verifies the hardcoded preset runs a loop
// with reasonable output.
func TestPark2015PreisachEverettPreset(t *testing.T) {
	ev := Park2015PreisachEverett()
	stack := NewPreisachStack(3e8, ev)
	if stack == nil {
		t.Fatal("NewPreisachStack returned nil")
	}
	// Drive to negative saturation, then sweep up.
	stack.Update(-3e8)
	pAtZero := stack.Update(0)
	pAtPos := stack.Update(3e8)
	t.Logf("P(E=0)=%.4f C/m^2, P(E=3MV/cm)=%.4f C/m^2", pAtZero, pAtPos)
	// P at positive saturation should be close to Ps.
	if math.Abs(pAtPos-ev.Ps) > 0.02 {
		t.Errorf("P(3MV/cm) = %.4f, expected ~Ps=%.4f", pAtPos, ev.Ps)
	}
}

func relErr(got, want float64) float64 {
	if want == 0 {
		return math.Abs(got)
	}
	return math.Abs(got-want) / math.Abs(want)
}
