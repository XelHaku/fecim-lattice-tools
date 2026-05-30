package peripherals

import (
	"testing"

	"fecim-lattice-tools/shared/testutil"
)

// TestDACEnergyPerConversion_Theoretical verifies E_DAC = C * V^2 * N.
func TestDACEnergyPerConversion_Theoretical(t *testing.T) {
	dac := DefaultDAC()
	got := dac.EnergyPerConversion()

	const cEff = 0.2e-15
	v := (dac.VrefHigh - dac.VrefLow) / 2.0
	n := float64(dac.Levels())
	want := cEff * v * v * n

	if testutil.RelErr(got, want) > 1e-15 {
		t.Fatalf("DAC energy mismatch: got=%e J want=%e J rel_err=%e", got, want, testutil.RelErr(got, want))
	}
}

// TestCrossbarArrayStaticPowerSum verifies P_array = sum(G_cell * V^2).
func TestCrossbarArrayStaticPowerSum(t *testing.T) {
	vRead := 0.2
	cells := []float64{5e-6, 10e-6, 15e-6, 20e-6} // conductances in S

	var got float64
	for _, g := range cells {
		got += g * vRead * vRead
	}

	want := (5e-6 + 10e-6 + 15e-6 + 20e-6) * vRead * vRead
	if testutil.RelErr(got, want) > 1e-15 {
		t.Fatalf("crossbar static power mismatch: got=%e W want=%e W rel_err=%e", got, want, testutil.RelErr(got, want))
	}
}

// TestTIAPowerConsumptionModel verifies the implemented TIA power model.
func TestTIAPowerConsumptionModel(t *testing.T) {
	tia := DefaultTIA()
	got := tia.PowerConsumption()

	const efficiency = 0.1
	want := 2 * kBT300 * tia.Bandwidth * tia.Gain / efficiency

	if testutil.RelErr(got, want) > 1e-15 {
		t.Fatalf("TIA power mismatch: got=%e W want=%e W rel_err=%e", got, want, testutil.RelErr(got, want))
	}
}

// TestTotalInferenceEnergyComponents verifies E_total = E_DAC + E_array + E_TIA + E_ADC.
func TestTotalInferenceEnergyComponents(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()

	const (
		cells    = 128
		vRead    = 0.2
		gCell    = 20e-6
		readTime = 20e-9
	)

	dacEnergy := float64(cells) * dac.EnergyPerConversion()
	arrayPower := float64(cells) * gCell * vRead * vRead
	arrayEnergy := arrayPower * readTime
	tiaEnergy := float64(cells) * tia.PowerConsumption() * readTime
	adcEnergy := float64(cells) * adc.EnergyPerConversion()

	total := dacEnergy + arrayEnergy + tiaEnergy + adcEnergy
	components := []float64{dacEnergy, arrayEnergy, tiaEnergy, adcEnergy}

	var sum float64
	for _, e := range components {
		sum += e
	}

	if testutil.RelErr(total, sum) > 1e-15 {
		t.Fatalf("total energy mismatch: total=%e J sum_components=%e J rel_err=%e", total, sum, testutil.RelErr(total, sum))
	}
	if total <= 0 {
		t.Fatalf("total inference energy must be positive, got=%e J", total)
	}
}

// TestInferenceEnergyLinearScaling verifies energy scales linearly with array size (cell count).
func TestInferenceEnergyLinearScaling(t *testing.T) {
	dac := DefaultDAC()
	adc := DefaultADC()
	tia := DefaultTIA()

	energyForCells := func(cells int) float64 {
		const (
			vRead    = 0.2
			gCell    = 20e-6
			readTime = 20e-9
		)
		dacEnergy := float64(cells) * dac.EnergyPerConversion()
		arrayEnergy := (float64(cells) * gCell * vRead * vRead) * readTime
		tiaEnergy := float64(cells) * tia.PowerConsumption() * readTime
		adcEnergy := float64(cells) * adc.EnergyPerConversion()
		return dacEnergy + arrayEnergy + tiaEnergy + adcEnergy
	}

	small := 32 * 32
	large := 64 * 64
	eSmall := energyForCells(small)
	eLarge := energyForCells(large)

	wantRatio := float64(large) / float64(small)
	gotRatio := eLarge / eSmall

	if testutil.RelErr(gotRatio, wantRatio) > 1e-12 {
		t.Fatalf("energy scaling mismatch: got_ratio=%f want_ratio=%f (small=%e J large=%e J)", gotRatio, wantRatio, eSmall, eLarge)
	}
}
