//go:build legacy_fyne

package gui

import "testing"

func TestComposedSenseSNRdB_DegradesWithNoise(t *testing.T) {
	embedded, app, win := setupUnifiedTestApp(t)
	defer app.Quit()
	defer win.Close()
	defer embedded.Stop()

	ca := embedded.CircuitsApp
	sense, ok := ca.senseChainConfig()
	if !ok {
		t.Fatal("expected sense chain")
	}

	currentA := 5e-6
	ca.tia.InputNoiseRMS = 1e-12
	snrLowNoise := ca.composedSenseSNRdB(currentA, sense)

	ca.tia.InputNoiseRMS = 8e-12
	snrHighNoise := ca.composedSenseSNRdB(currentA, sense)

	if !(snrHighNoise < snrLowNoise) {
		t.Fatalf("expected SNR degradation with higher noise: low=%.3f dB high=%.3f dB", snrLowNoise, snrHighNoise)
	}
}
