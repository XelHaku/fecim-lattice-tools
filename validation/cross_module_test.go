//go:build legacy_fyne

package validation

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	mnistcore "fecim-lattice-tools/module3-mnist/pkg/core"
	circuitsgui "fecim-lattice-tools/module4-circuits/pkg/gui"
	"fecim-lattice-tools/shared/crossbar"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

func TestCrossModule_MaterialSharedMatchesModule1(t *testing.T) {
	sharedMat := sharedphysics.FeCIMMaterial()
	module1Mat := ferroelectric.FeCIMMaterial()

	if sharedMat == nil || module1Mat == nil {
		t.Fatalf("nil material: shared=%v module1=%v", sharedMat, module1Mat)
	}

	if sharedMat.Name != module1Mat.Name {
		t.Fatalf("material name mismatch: shared=%q module1=%q", sharedMat.Name, module1Mat.Name)
	}
	if sharedMat.NumLevels != module1Mat.NumLevels {
		t.Fatalf("NumLevels mismatch: shared=%d module1=%d", sharedMat.NumLevels, module1Mat.NumLevels)
	}
	if sharedMat.Ec != module1Mat.Ec {
		t.Fatalf("Ec mismatch: shared=%g module1=%g", sharedMat.Ec, module1Mat.Ec)
	}
	if sharedMat.Ps != module1Mat.Ps {
		t.Fatalf("Ps mismatch: shared=%g module1=%g", sharedMat.Ps, module1Mat.Ps)
	}
	if sharedMat.Gmin != module1Mat.Gmin || sharedMat.Gmax != module1Mat.Gmax {
		t.Fatalf("conductance bounds mismatch: shared=[%g,%g] module1=[%g,%g]",
			sharedMat.Gmin, sharedMat.Gmax, module1Mat.Gmin, module1Mat.Gmax)
	}

	for _, level := range []int{0, sharedMat.NumLevels / 2, sharedMat.NumLevels - 1} {
		gotShared := sharedMat.DiscreteLevel(level, sharedMat.NumLevels)
		gotModule1 := module1Mat.DiscreteLevel(level, module1Mat.NumLevels)
		if gotShared != gotModule1 {
			t.Fatalf("DiscreteLevel mismatch at level %d: shared=%g module1=%g", level, gotShared, gotModule1)
		}
	}
}

func TestCrossModule_CrossbarConductanceMatchesModule3CIMExpectation(t *testing.T) {
	arr, err := crossbar.NewArray(&crossbar.Config{Rows: 1, Cols: 1, ADCBits: 16, DACBits: 16, NoiseLevel: 0})
	if err != nil {
		t.Fatalf("crossbar.NewArray failed: %v", err)
	}

	const programmedConductance = 22.0 / (sharedphysics.DefaultLevels - 1) // exact DefaultLevels grid point
	if err := arr.ProgramWeight(0, 0, programmedConductance); err != nil {
		t.Fatalf("ProgramWeight failed: %v", err)
	}
	gStored := arr.GetConductanceMatrix()[0][0]

	// Module3 legacy CIM expectation for crossbar conductance decoding:
	// effective_weight = (conductance - 0.5) * 4
	legacyExpectedWeight := (gStored - 0.5) * 4.0

	net := mnistcore.NewDualModeNetwork(1, 1, 1)
	net.Config.SingleLayer = true
	net.Config.NoiseLevel = 0
	net.Config.NoiseADC = 0
	net.Config.NoiseThermal = 0
	net.Config.NoiseFlicker = 0
	net.Config.NoiseCellVariation = 0
	net.Config.ADCBits = 16
	net.Config.DACBits = 16
	net.QuantSingleLayerWeights = [][]float64{{legacyExpectedWeight}}
	net.QuantSingleLayerBias = []float64{0}

	pred, _, probs := net.InferCIMOnly([]float64{1.0})
	if pred != 0 || len(probs) != 1 {
		t.Fatalf("unexpected CIM inference shape: pred=%d probs=%v", pred, probs)
	}

	// Crossbar stores normalized conductance on the same DefaultLevels grid used by CIM path.
	level := crossbar.GetLevel(gStored)
	wantFromLevel := (float64(level)/float64(sharedphysics.DefaultLevels-1) - 0.5) * 4.0
	if math.Abs(legacyExpectedWeight-wantFromLevel) > 1e-12 {
		t.Fatalf("legacy mapping drift: got %.15f want %.15f (level=%d g=%.15f)", legacyExpectedWeight, wantFromLevel, level, gStored)
	}

	// Also pin physical conductance range contract used across crossbar and CIM path.
	if crossbar.GMin != sharedphysics.GMin || crossbar.GMax != sharedphysics.GMax {
		t.Fatalf("crossbar physical range drifted from shared physics: crossbar=[%g,%g] shared=[%g,%g]",
			crossbar.GMin, crossbar.GMax, sharedphysics.GMin, sharedphysics.GMax)
	}
}

func TestCrossModule_ISPPSharedCalculatorMatchesModule4Path(t *testing.T) {
	ds := circuitsgui.NewDeviceState(4, 4, nil, nil)
	mat := sharedphysics.FeCIMMaterial()
	ds.SetMaterial(mat)

	const (
		currentLevel = 5
		targetLevel  = 20
	)
	ds.StartISPP(0, 0, targetLevel, currentLevel)
	status := ds.GetISPPStatus()
	if !status.Active {
		t.Fatal("ISPP should be active after StartISPP")
	}

	calc := sharedphysics.NewISPPCalculator(mat.CoerciveVoltage(), mat.GetNumLevels())
	if status.Voltage <= 0 {
		t.Fatalf("invalid start voltage from module4 ISPP path: %g", status.Voltage)
	}

	result := ds.ISPPIterate(currentLevel + 1) // still below target in ascending direction
	if result != circuitsgui.ISPPResultContinue {
		t.Fatalf("expected ISPPResultContinue, got %v", result)
	}
	next := ds.GetISPPStatus()

	wantNext := calc.CalculateNextVoltage(status.Voltage, sharedphysics.DirectionAscending)
	if math.Abs(next.Voltage-wantNext) > 1e-12 {
		t.Fatalf("next voltage mismatch vs shared ISPP calculator: got=%g want=%g", next.Voltage, wantNext)
	}
}
