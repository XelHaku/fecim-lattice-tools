package gui_test

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/module2-crossbar/pkg/gui/tabs"
	"fecim-lattice-tools/shared/crossbar"
)

// M2-GUI-01: Headless GUI smoke test (tab lifecycle).
func TestM2GUI01_TabLifecycle_NoPanics(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := app.NewWindow("m2-gui-lifecycle")
	defer w.Close()

	arr, err := crossbar.NewArray(&crossbar.Config{
		Rows:             16,
		Cols:             16,
		NoiseLevel:       0,
		ADCBits:          8,
		DACBits:          8,
		UseGPU:           false,
		ConductanceModel: crossbar.ConductanceLinear,
		Endurance:        crossbar.DefaultEnduranceConfig(),
		HalfSelect:       crossbar.DefaultHalfSelectConfig(),
	})
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	ideal := tabs.NewIdealTab(arr, func() {})
	w.SetContent(ideal.Content())
	w.Content().Refresh()

	ird := tabs.NewIRDropTab(16)
	w.SetContent(ird.Content())
	w.Content().Refresh()

	snk := tabs.NewSneakTab(16)
	w.SetContent(snk.Content())
	w.Content().Refresh()

	dft := tabs.NewDriftTab(16)
	w.SetContent(dft.Content())
	w.Content().Refresh()
}
