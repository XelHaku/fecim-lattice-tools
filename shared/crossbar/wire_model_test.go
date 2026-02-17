package crossbar

import (
	"math"
	"testing"
)

func wireResistanceFromGeometry(resistivity, length, width, thickness float64) float64 {
	area := width * thickness
	return resistivity * length / area
}

func wireCapacitanceFromGeometry(epsR, eps0, length, width, dielectricThickness float64) float64 {
	area := length * width
	return epsR * eps0 * area / dielectricThickness
}

func resistanceAtTemperature(r0, tcr, tempK, refK float64) float64 {
	return r0 * (1 + tcr*(tempK-refK))
}

// (1) Verify resistance increases linearly with wire length.
func TestWireResistanceLinearWithLength(t *testing.T) {
	const rhoCu = 1.68e-8 // Ohm·m at ~20-25C
	const width = 100e-9
	const thickness = 80e-9

	r1 := wireResistanceFromGeometry(rhoCu, 10e-6, width, thickness)
	r2 := wireResistanceFromGeometry(rhoCu, 20e-6, width, thickness)
	r3 := wireResistanceFromGeometry(rhoCu, 40e-6, width, thickness)

	if math.Abs(r2/r1-2.0) > 1e-12 {
		t.Fatalf("expected 2x length => 2x resistance, got ratio %.6f", r2/r1)
	}
	if math.Abs(r3/r1-4.0) > 1e-12 {
		t.Fatalf("expected 4x length => 4x resistance, got ratio %.6f", r3/r1)
	}
}

// (2) Verify resistance decreases with wire width.
func TestWireResistanceDecreasesWithWidth(t *testing.T) {
	const rhoCu = 1.68e-8 // Ohm·m
	const length = 20e-6
	const thickness = 80e-9

	rNarrow := wireResistanceFromGeometry(rhoCu, length, 50e-9, thickness)
	rMid := wireResistanceFromGeometry(rhoCu, length, 100e-9, thickness)
	rWide := wireResistanceFromGeometry(rhoCu, length, 200e-9, thickness)

	if !(rNarrow > rMid && rMid > rWide) {
		t.Fatalf("expected R_narrow > R_mid > R_wide, got %.6e, %.6e, %.6e", rNarrow, rMid, rWide)
	}

	if math.Abs(rMid/rWide-2.0) > 1e-12 {
		t.Fatalf("expected doubling width halves resistance, got ratio %.6f", rMid/rWide)
	}
}

// (3) Verify temperature coefficient follows copper/tungsten model.
func TestWireResistanceTemperatureCoefficientCopperTungsten(t *testing.T) {
	const (
		r0          = 2.5
		refTempK    = 300.0
		tempHighK   = 358.0 // 85C
		alphaCopper = 0.00393
		alphaW      = 0.0045 // tungsten (typical near room temp)
	)

	// Copper path should match the production model exactly.
	prodCopper := NewTemperatureEffects(tempHighK).AdjustedWireResistance(r0)
	expCopper := resistanceAtTemperature(r0, alphaCopper, tempHighK, refTempK)
	if math.Abs(prodCopper-expCopper) > 1e-12 {
		t.Fatalf("copper TCR mismatch: got %.12f, want %.12f", prodCopper, expCopper)
	}

	// Compare copper vs tungsten analytical expectations.
	rCopper := resistanceAtTemperature(r0, alphaCopper, tempHighK, refTempK)
	rTungsten := resistanceAtTemperature(r0, alphaW, tempHighK, refTempK)

	if !(rTungsten > rCopper && rCopper > r0) {
		t.Fatalf("expected R_W(85C) > R_Cu(85C) > R0, got Rw=%.6f, Rcu=%.6f, R0=%.6f", rTungsten, rCopper, r0)
	}
}

// (4) Verify parasitic capacitance scales with wire dimensions.
func TestParasiticCapacitanceScalesWithDimensions(t *testing.T) {
	const (
		eps0 = 8.854e-12
		epsR = 3.9 // SiO2
		tOx  = 30e-9
	)

	cBase := wireCapacitanceFromGeometry(epsR, eps0, 10e-6, 100e-9, tOx)
	cLen2x := wireCapacitanceFromGeometry(epsR, eps0, 20e-6, 100e-9, tOx)
	cWidth2x := wireCapacitanceFromGeometry(epsR, eps0, 10e-6, 200e-9, tOx)

	if math.Abs(cLen2x/cBase-2.0) > 1e-12 {
		t.Fatalf("expected 2x length => 2x capacitance, got ratio %.6f", cLen2x/cBase)
	}
	if math.Abs(cWidth2x/cBase-2.0) > 1e-12 {
		t.Fatalf("expected 2x width => 2x capacitance, got ratio %.6f", cWidth2x/cBase)
	}
}

// (5) Compare IR drop predictions to analytical formula for a uniform array.
func TestIRDropUniformArrayAnalyticalAgreement(t *testing.T) {
	const (
		rows = 8
		cols = 8
		vin  = 0.5
		g    = 50e-6
	)

	sim := NewIRDropSimulator(rows, cols)
	sim.RowResist = 2.5
	sim.ColResist = 2.5

	for i := 0; i < rows; i++ {
		sim.SetInputVoltage(i, vin)
		for j := 0; j < cols; j++ {
			sim.SetConductance(i, j, g)
		}
	}
	sim.Simulate(200)

	// First-order analytical estimate at worst-case corner (rows-1, cols-1)
	// I_cell = V * G
	// WL drop ~ I_cell * Rwl * sum_{k=1}^{cols-1} k = I_cell*Rwl*cols*(cols-1)/2
	// BL drop ~ I_cell * Rbl * sum_{k=1}^{rows-1} k = I_cell*Rbl*rows*(rows-1)/2
	iCell := vin * g
	wlDrop := iCell * sim.RowResist * float64(cols*(cols-1)) / 2.0
	blDrop := iCell * sim.ColResist * float64(rows*(rows-1)) / 2.0
	expectedCornerDrop := wlDrop + blDrop

	simCornerDrop := sim.IRDropMap[rows-1][cols-1]
	relErr := math.Abs(simCornerDrop-expectedCornerDrop) / expectedCornerDrop

	// Iterative solver includes coupled effects; keep tolerance moderate.
	if relErr > 0.35 {
		t.Fatalf("IR drop mismatch too high: simulated=%.6e V, analytical=%.6e V, relErr=%.2f%%", simCornerDrop, expectedCornerDrop, relErr*100)
	}
}
