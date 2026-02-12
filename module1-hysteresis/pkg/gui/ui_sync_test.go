package gui

import (
	"image/color"
	"reflect"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/gui/widgets"
)

func newSyncTestApp(t *testing.T) *App {
	t.Helper()
	test.NewApp()
	return &App{
		plot:           widgets.NewPEPlot(1, 1, color.RGBA{}, color.RGBA{}, color.RGBA{}, color.RGBA{}, color.RGBA{}, color.RGBA{}),
		levelIndicator: widgets.NewLevelIndicator(),
		cellViz:        widgets.NewCellVisualizer(),
		eFieldSlider:   widget.NewSlider(-2, 2),
		eFieldLabel:    widget.NewLabel(""),
		pLabel:         widget.NewLabel(""),
		levelLabel:     widget.NewLabel(""),
		statusLabel:    widget.NewLabel(""),
		logText:        widget.NewLabel(""),
		phaseIndicator: widgets.NewPhaseIndicator(),
		numLevels:      30,
		logEntries:     []string{},
	}
}

func TestRefreshGUI_SnapshotStaysInSyncWithRenderedValues(t *testing.T) {
	a := newSyncTestApp(t)

	hE := []float64{-1.2e8, 0, 1.1e8}
	hP := []float64{-0.21, 0.02, 0.19}
	s := uiSnapshot{
		fE:            1.1e8,
		pV:            0.19,
		dL:            22,
		eC:            1.0e8,
		hE:            hE,
		hP:            hP,
		numLevels:     30,
		waveform:      WaveformSine,
		physicsEngine: PhysicsPreisach,
	}

	a.refreshGUI(s)

	if got := a.eFieldLabel.Text; got != "E-field: 1.100 MV/cm" {
		t.Fatalf("e-field label desync: got %q", got)
	}
	if got := a.pLabel.Text; got != "19.00 µC/cm²" {
		t.Fatalf("polarization label desync: got %q", got)
	}
	if got := a.levelLabel.Text; got != "23/30" {
		t.Fatalf("level label desync: got %q", got)
	}

	gotHE, gotHP, gotE, gotP, gotFilter := a.plot.DataSnapshot()
	if !reflect.DeepEqual(gotHE, hE) || !reflect.DeepEqual(gotHP, hP) {
		t.Fatalf("plot history desync: got hE=%v hP=%v", gotHE, gotHP)
	}
	if gotE != s.fE || gotP != s.pV {
		t.Fatalf("plot cursor desync: got E=%g P=%g want E=%g P=%g", gotE, gotP, s.fE, s.pV)
	}
	if !gotFilter {
		t.Fatal("preisach mode should keep spike filtering enabled")
	}
	if got := a.levelIndicator.Level(); got != s.dL {
		t.Fatalf("level indicator desync: got=%d want=%d", got, s.dL)
	}
}

func TestRefreshGUI_WRDTargetUsesControllerStateAsTruth(t *testing.T) {
	a := newSyncTestApp(t)
	a.lastTargetMismatchLog = time.Now()
	s := uiSnapshot{
		fE:                    0.4e8,
		pV:                    0.05,
		dL:                    10,
		eC:                    1.0e8,
		hE:                    []float64{0.1e8, 0.2e8},
		hP:                    []float64{0.0, 0.05},
		numLevels:             30,
		waveform:              WaveformWriteReadDemo,
		physicsEngine:         PhysicsLandau,
		wrdPhase:              3,
		wrdTargetLevel:        11,
		controllerState:       controller.StateVerify,
		controllerTargetLevel: 17,
		widgets:               (&App{}).buildWidgetSnapshot(0.4e8, 10, 1.0e8, WaveformWriteReadDemo, 3, 11, false, 0, 0, controller.StateVerify, 17, -1),
		logText:               "snapshot log",
	}

	a.refreshGUI(s)

	target, highlight, mode := a.levelIndicator.TargetState()
	if target != 17 || !highlight || mode != widgets.TargetModeVerify {
		t.Fatalf("target highlight desync: got level=%d highlight=%v mode=%v", target, highlight, mode)
	}
	if got := a.logText.Text; got != "snapshot log" {
		t.Fatalf("log text should come from snapshot; got %q", got)
	}

	phase, phaseMode := a.phaseIndicator.CurrentPhase()
	if phase != widgets.PhaseVerify || phaseMode != "wrd" {
		t.Fatalf("phase indicator desync: got phase=%d mode=%q, want VERIFY/wrd", phase, phaseMode)
	}
}

func TestBuildWidgetSnapshot_WRDSettledUsesControllerTarget(t *testing.T) {
	a := &App{}
	ws := a.buildWidgetSnapshot(0.0, 16, 1.0e8, WaveformWriteReadDemo, 3, 11, false, 0, 0, controller.StateVerify, 17, -1)
	if ws.target.highlight {
		t.Fatalf("expected settled target to clear highlight when level matches controller target")
	}
	if ws.target.level != 17 {
		t.Fatalf("expected controller target level 17, got %d", ws.target.level)
	}
	if ws.phase.phase != 1 {
		t.Fatalf("expected WRD verify phase (1) when controller in StateVerify, got %d", ws.phase.phase)
	}
}

func TestBuildWidgetSnapshot_WRDIdleDoesNotUseControllerTarget(t *testing.T) {
	a := &App{}
	ws := a.buildWidgetSnapshot(0.0, 10, 1.0e8, WaveformWriteReadDemo, 5, 11, false, 0, 0, controller.StateIdle, 17, -1)
	if ws.target.level != 11 {
		t.Fatalf("expected WRD target while controller idle; got %d", ws.target.level)
	}
}
