//go:build legacy_fyne

package gui

import (
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module3-mnist/pkg/core"
)

// DualModeAppTestHooks exposes a minimal, stable surface for headless GUI integration tests.
// These hooks are intentionally narrow: they let tests verify widget wiring and parameter binding
// without relying on fragile widget-tree searches.
//
// This is used by module3-mnist/tests/gui_integration_test.go.
type DualModeAppTestHooks struct {
	LevelsSelect *widget.Select
	NoiseSlider  *widget.Slider

	StatusLabel    *widget.Label
	FPPredLabel    *widget.Label
	CIMPredLabel   *widget.Label
	AgreementLabel *widget.Label
}

func (app *DualModeApp) TestHooks() DualModeAppTestHooks {
	return DualModeAppTestHooks{
		LevelsSelect:   app.levelsSelect,
		NoiseSlider:    app.noiseSlider,
		StatusLabel:    app.statusLabel,
		FPPredLabel:    app.fpPredLabel,
		CIMPredLabel:   app.cimPredLabel,
		AgreementLabel: app.agreementLabel,
	}
}

// NetworkConfig returns the active network configuration (shared by FP/CIM paths).
func (app *DualModeApp) NetworkConfig() *core.NetworkConfig {
	return app.network().Config
}

// RunInferenceForTest triggers inference and UI update. Intended for headless integration tests.
func (app *DualModeApp) RunInferenceForTest(pixels []float64) {
	app.runInference(pixels)
}
