package arraysim

import (
	"math"
	"testing"
)

const testEpsilon = 1e-6

func TestTierA_PassiveHalfSelectPattern(t *testing.T) {
	solver := NewTierASolver()

	conductance := [][]float64{
		{10e-6, 10e-6},
		{10e-6, 10e-6},
	}
	params := SolveParams{
		WLVoltages:  []float64{0.5, 0.0},
		BLVoltages:  []float64{-0.5, 0.0},
		Conductance: conductance,
		ActiveRows:  []bool{true, true},
		Geometry:    CellGeometry{},
		Wire: WireParams{
			RWordLine: 1000,
			RBitLine:  1000,
		},
	}

	result, err := solver.Solve(params)
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}
	if len(result.CellVoltages) != 2 || len(result.CellVoltages[0]) != 2 {
		t.Fatalf("unexpected cell voltage size: %#v", result.CellVoltages)
	}

	want := [][]float64{
		{0.971057, 0.477403},
		{0.478505, -0.012409},
	}

	const eps = 5e-3
	for r := range want {
		for c := range want[r] {
			got := result.CellVoltages[r][c]
			if math.Abs(got-want[r][c]) > eps {
				t.Fatalf("cell (%d,%d) voltage: got %.6f, want %.6f", r, c, got, want[r][c])
			}
		}
	}

	if result.CellVoltages[0][0] <= result.CellVoltages[0][1] {
		t.Fatalf("target cell should have highest voltage, got %.6f <= %.6f", result.CellVoltages[0][0], result.CellVoltages[0][1])
	}
	if math.Abs(result.CellVoltages[1][1]) > 0.05 {
		t.Fatalf("diagonal cell should remain near 0V under coupling, got %.6f", result.CellVoltages[1][1])
	}
}

func TestTierA_RwireZeroMatchesIdeal(t *testing.T) {
	solver := NewTierASolver()

	conductance := [][]float64{
		{1e-3, 2e-3},
		{3e-3, 4e-3},
	}
	params := SolveParams{
		WLVoltages:  []float64{0.7, -0.1},
		BLVoltages:  []float64{0.2, -0.3},
		Conductance: conductance,
		ActiveRows:  []bool{true, true},
		// WireParams.WithDefaults treats <=0 as "unset", so use a tiny value
		// to approximate the Rwire->0 limit.
		Wire: WireParams{RWordLine: 1e-18, RBitLine: 1e-18},
	}

	result, err := solver.Solve(params)
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}

	for r := range conductance {
		for c := range conductance[r] {
			wantV := params.WLVoltages[r] - params.BLVoltages[c]
			gotV := result.CellVoltages[r][c]
			if math.Abs(gotV-wantV) > testEpsilon {
				t.Fatalf("cell (%d,%d) voltage: got %.9f, want %.9f", r, c, gotV, wantV)
			}

			wantI := conductance[r][c] * wantV
			gotI := result.CellCurrents[r][c]
			if math.Abs(gotI-wantI) > testEpsilon {
				t.Fatalf("cell (%d,%d) current: got %.9g, want %.9g", r, c, gotI, wantI)
			}
		}
	}
}

func TestTierA_ActiveRowsMaskMatchesZeroedRows(t *testing.T) {
	solver := NewTierASolver()

	conductance := [][]float64{
		{1e-3},
		{50e-3}, // large "would matter" row we will mask out
		{1e-3},
	}
	base := SolveParams{
		WLVoltages:  []float64{0.4, 0.4, 0.4},
		BLVoltages:  []float64{0.0},
		Conductance: conductance,
		Wire:        WireParams{RWordLine: 100, RBitLine: 100},
	}

	masked := base
	masked.ActiveRows = []bool{true, false, true}
	resMasked, err := solver.Solve(masked)
	if err != nil {
		t.Fatalf("Solve(masked) returned error: %v", err)
	}

	zeroed := base
	zeroed.Conductance = [][]float64{
		{1e-3},
		{0.0},
		{1e-3},
	}
	zeroed.ActiveRows = nil // all active, but row 1 contributes zero current
	resZeroed, err := solver.Solve(zeroed)
	if err != nil {
		t.Fatalf("Solve(zeroed) returned error: %v", err)
	}

	if len(resMasked.ColCurrents) != 1 || len(resZeroed.ColCurrents) != 1 {
		t.Fatalf("unexpected col current size")
	}
	if math.Abs(resMasked.ColCurrents[0]-resZeroed.ColCurrents[0]) > testEpsilon {
		t.Fatalf("col current mismatch: masked %.9g vs zeroed %.9g", resMasked.ColCurrents[0], resZeroed.ColCurrents[0])
	}

	for _, r := range []int{0, 2} {
		if math.Abs(resMasked.CellVoltages[r][0]-resZeroed.CellVoltages[r][0]) > testEpsilon {
			t.Fatalf("row %d voltage mismatch: masked %.9f vs zeroed %.9f", r, resMasked.CellVoltages[r][0], resZeroed.CellVoltages[r][0])
		}
		if math.Abs(resMasked.CellCurrents[r][0]-resZeroed.CellCurrents[r][0]) > testEpsilon {
			t.Fatalf("row %d current mismatch: masked %.9g vs zeroed %.9g", r, resMasked.CellCurrents[r][0], resZeroed.CellCurrents[r][0])
		}
	}

	if resMasked.CellCurrents[1][0] != 0 {
		t.Fatalf("inactive row should produce 0 current, got I=%.9g", resMasked.CellCurrents[1][0])
	}
}

func TestTierA_NegativeVoltagePreservesSign(t *testing.T) {
	solver := NewTierASolver()

	params := SolveParams{
		WLVoltages:  []float64{0.0},
		BLVoltages:  []float64{0.5},
		Conductance: [][]float64{{10e-6}},
		ActiveRows:  []bool{true},
		Wire:        WireParams{RWordLine: 1000, RBitLine: 1000},
	}

	result, err := solver.Solve(params)
	if err != nil {
		t.Fatalf("Solve returned error: %v", err)
	}

	ideal := params.WLVoltages[0] - params.BLVoltages[0]
	gotV := result.CellVoltages[0][0]
	gotI := result.CellCurrents[0][0]

	if ideal >= 0 {
		t.Fatalf("test setup: expected negative ideal voltage, got %.6f", ideal)
	}
	if gotV >= 0 {
		t.Fatalf("expected negative cell voltage, got %.9f", gotV)
	}
	if math.Abs(gotV) > math.Abs(ideal)+testEpsilon {
		t.Fatalf("IR-drop should not increase |V|: got %.9f, ideal %.9f", gotV, ideal)
	}
	if gotI >= 0 {
		t.Fatalf("expected negative cell current, got %.9g", gotI)
	}
}

func TestTierA_MatchesDenseReferenceSolve(t *testing.T) {
	solver := NewTierASolver()
	params := SolveParams{
		WLVoltages: []float64{0.8, 0.4, -0.1},
		BLVoltages: []float64{0.2, -0.2, 0.1},
		Conductance: [][]float64{
			{2e-3, 1e-3, 0.5e-3},
			{3e-3, 0.0, 1.5e-3},
			{1e-3, 2.5e-3, 4e-3},
		},
		ActiveRows: []bool{true, true, false},
		Wire:       WireParams{RWordLine: 42.0, RBitLine: 57.0},
	}

	got, err := solver.Solve(params)
	if err != nil {
		t.Fatalf("TierA solve error: %v", err)
	}
	want, err := referenceSolveDense(params)
	if err != nil {
		t.Fatalf("reference solve error: %v", err)
	}

	const eps = 1e-12
	for r := range want.CellVoltages {
		for c := range want.CellVoltages[r] {
			if math.Abs(got.CellVoltages[r][c]-want.CellVoltages[r][c]) > eps {
				t.Fatalf("cell voltage mismatch (%d,%d): got %.15g want %.15g", r, c, got.CellVoltages[r][c], want.CellVoltages[r][c])
			}
			if math.Abs(got.CellCurrents[r][c]-want.CellCurrents[r][c]) > eps {
				t.Fatalf("cell current mismatch (%d,%d): got %.15g want %.15g", r, c, got.CellCurrents[r][c], want.CellCurrents[r][c])
			}
		}
		if math.Abs(got.RowCurrents[r]-want.RowCurrents[r]) > eps {
			t.Fatalf("row current mismatch (%d): got %.15g want %.15g", r, got.RowCurrents[r], want.RowCurrents[r])
		}
	}
	for c := range want.ColCurrents {
		if math.Abs(got.ColCurrents[c]-want.ColCurrents[c]) > eps {
			t.Fatalf("col current mismatch (%d): got %.15g want %.15g", c, got.ColCurrents[c], want.ColCurrents[c])
		}
	}

	if maxResidual := kclMaxResidual(params, want); maxResidual > 1e-9 {
		t.Fatalf("unexpected KCL residual from dense reference: %g", maxResidual)
	}
}

func TestTierA_SelectorRonReducesReadCurrentAndMargin(t *testing.T) {
	solver := NewTierASolver()

	buildParams := func(targetG float64, selectorOn bool) SolveParams {
		return SolveParams{
			WLVoltages:      []float64{0.2},
			BLVoltages:      []float64{0.0},
			Conductance:     [][]float64{{targetG}},
			ActiveRows:      []bool{true},
			Wire:            WireParams{RWordLine: 1e-18, RBitLine: 1e-18},
			SelectorEnabled: selectorOn,
			SelectorRon:     20e3,
		}
	}

	highOff, err := solver.Solve(buildParams(80e-6, false))
	if err != nil {
		t.Fatalf("highOff solve error: %v", err)
	}
	lowOff, err := solver.Solve(buildParams(20e-6, false))
	if err != nil {
		t.Fatalf("lowOff solve error: %v", err)
	}
	highOn, err := solver.Solve(buildParams(80e-6, true))
	if err != nil {
		t.Fatalf("highOn solve error: %v", err)
	}
	lowOn, err := solver.Solve(buildParams(20e-6, true))
	if err != nil {
		t.Fatalf("lowOn solve error: %v", err)
	}

	iHighOff := math.Abs(highOff.RowCurrents[0])
	iLowOff := math.Abs(lowOff.RowCurrents[0])
	iHighOn := math.Abs(highOn.RowCurrents[0])
	iLowOn := math.Abs(lowOn.RowCurrents[0])

	if !(iHighOn < iHighOff && iLowOn < iLowOff) {
		t.Fatalf("selector should reduce read current: high off/on=%g/%g low off/on=%g/%g", iHighOff, iHighOn, iLowOff, iLowOn)
	}

	marginOff := iHighOff - iLowOff
	marginOn := iHighOn - iLowOn
	if !(marginOn < marginOff) {
		t.Fatalf("selector should degrade read margin: off=%g on=%g", marginOff, marginOn)
	}
}

func TestSenseChain_ConvertCurrent(t *testing.T) {
	sense := SenseChain{
		TIA: TIAConfig{
			Rf:   10e3,
			Vref: 0.1,
			Vmin: 0.0,
			Vmax: 1.0,
		},
		ADC: ADCConfig{
			Bits: 4,
			Vmin: 0.0,
			Vmax: 1.0,
		},
	}

	tests := []struct {
		name       string
		currentA   float64
		wantVout   float64
		wantCode   int
		wantTIASat bool
		wantADCSat bool
	}{
		{
			name:       "in-range",
			currentA:   40e-6,
			wantVout:   0.5,
			wantCode:   8,
			wantTIASat: false,
			wantADCSat: false,
		},
		{
			name:       "high-saturation",
			currentA:   200e-6,
			wantVout:   1.0,
			wantCode:   15,
			wantTIASat: true,
			wantADCSat: true,
		},
		{
			name:       "low-saturation",
			currentA:   -20e-6,
			wantVout:   0.0,
			wantCode:   0,
			wantTIASat: true,
			wantADCSat: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := sense.ConvertCurrent(tc.currentA)
			if math.Abs(result.Vout-tc.wantVout) > testEpsilon {
				t.Fatalf("Vout: got %.6f, want %.6f", result.Vout, tc.wantVout)
			}
			if result.Code != tc.wantCode {
				t.Fatalf("Code: got %d, want %d", result.Code, tc.wantCode)
			}
			if result.TIASaturated != tc.wantTIASat {
				t.Fatalf("TIASaturated: got %v, want %v", result.TIASaturated, tc.wantTIASat)
			}
			if result.ADCSaturated != tc.wantADCSat {
				t.Fatalf("ADCSaturated: got %v, want %v", result.ADCSaturated, tc.wantADCSat)
			}
		})
	}
}
