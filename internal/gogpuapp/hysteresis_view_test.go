//go:build !cgo

package gogpuapp

import (
	"strings"
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildHysteresisView(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildHysteresisView(snapshot, theme)
	if w == nil {
		t.Fatal("buildHysteresisView returned nil")
	}
}

func TestBuildHysteresisViewContainsMaterialSections(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Sections) == 0 {
		t.Error("No material sections in snapshot")
	}
}

func TestDefaultHysteresisLoop(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	if len(snapshot.Plots) == 0 {
		t.Error("No plots in hysteresis snapshot")
		return
	}
	plot := snapshot.Plots[0]
	if len(plot.Series) == 0 {
		t.Error("No series in hysteresis plot")
		return
	}
	points := plot.Series[0].Points
	if len(points) != 200 {
		t.Errorf("Loop points len = %d, want 200", len(points))
	}
}

func TestHysteresisBoundaryNotice(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	if snapshot.Descriptor.BoundaryNotice == "" {
		t.Error("Expected boundary notice in hysteresis descriptor")
	}
}

func TestHysteresisOverlayLayoutLeavesHeaderAndAxisLabelSpace(t *testing.T) {
	for _, tc := range []struct {
		w, h int
		minY int
	}{
		{w: 1280, h: 820, minY: 340},
		{w: 1024, h: 768, minY: 330},
		{w: 900, h: 640, minY: 310},
	} {
		x, y, width, height := hysteresisOverlayPlotRegion(tc.w, tc.h)
		if y < float64(tc.minY) {
			t.Fatalf("%dx%d hysteresis overlay y = %.1f, want at least %d so the module title and boundary notice remain visible", tc.w, tc.h, y, tc.minY)
		}
		if y+height > float64(tc.h-40) {
			t.Fatalf("%dx%d hysteresis overlay bottom = %.1f, want <= %d", tc.w, tc.h, y+height, tc.h-40)
		}
		if x < 300 {
			t.Fatalf("%dx%d hysteresis overlay x = %.1f, want at least 300 for sidebar and Y-axis label space", tc.w, tc.h, x)
		}
		if width > 940 {
			t.Fatalf("%dx%d hysteresis overlay width = %.1f, want <= 940", tc.w, tc.h, width)
		}
	}
}

func TestCompactSimulationNoticeFitsDefaultScreenshotWidth(t *testing.T) {
	notice := compactSimulationNotice()
	if len([]rune(notice)) > 85 {
		t.Fatalf("compact simulation notice is %d runes, want <= 85 for screenshot readability: %q", len([]rune(notice)), notice)
	}
	for _, want := range []string{"EDUCATIONAL SIMULATION", "Published", "not silicon measurements"} {
		if !strings.Contains(notice, want) {
			t.Fatalf("compact simulation notice missing %q: %q", want, notice)
		}
	}
}

func TestCompactBoundaryNoticeFitsModuleHeader(t *testing.T) {
	vm := hysteresisvm.New()
	notice := compactBoundaryNotice(vm.Snapshot().Descriptor.BoundaryNotice)
	if len([]rune(notice)) > 90 {
		t.Fatalf("compact boundary notice is %d runes, want <= 90 for Module 1 header readability: %q", len([]rune(notice)), notice)
	}
	for _, want := range []string{"SIMULATION OUTPUT", "not measured device data", "cite"} {
		if !strings.Contains(notice, want) {
			t.Fatalf("compact boundary notice missing %q: %q", want, notice)
		}
	}
}

func TestHysteresisPlotData(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	found := false
	for _, plot := range snapshot.Plots {
		if plot.ID == "pe_loop" {
			found = true
			if plot.XLabel == "" || plot.YLabel == "" {
				t.Error("Plot axis labels missing")
			}
		}
	}
	if !found {
		t.Error("No pe_loop plot found in hysteresis snapshot")
	}
}

func TestHysteresisRetention(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	found := false
	for _, plot := range snapshot.Plots {
		if plot.ID == "retention" {
			found = true
			if len(plot.Series) == 0 || len(plot.Series[0].Points) == 0 {
				t.Error("Retention plot has no data points")
			}
		}
	}
	if !found {
		t.Error("No retention plot found in hysteresis snapshot")
	}
}

func TestHysteresisComputedMetrics(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	hasPr := false
	hasEc := false
	for _, m := range snapshot.Metrics {
		if m.ID == "pr" {
			hasPr = true
		}
		if m.ID == "ec_plus" {
			hasEc = true
		}
	}
	if !hasPr {
		t.Error("No Pr metric in hysteresis snapshot")
	}
	if !hasEc {
		t.Error("No Ec metric in hysteresis snapshot")
	}
}

func TestHysteresisViewActionButtonsDispatchActions(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	var actions []viewmodel.Action

	w := buildHysteresisViewWithActions(snapshot, theme, func(action viewmodel.Action) {
		actions = append(actions, action)
	})
	buttons := collectSidebarButtons(w)
	if len(buttons) < 4 {
		t.Fatalf("hysteresis button count = %d, want command and waveform controls", len(buttons))
	}

	clickButton(buttons[0])
	clickButton(buttons[1])
	clickButton(buttons[3])

	if len(actions) != 3 {
		t.Fatalf("dispatched action count = %d, want 3", len(actions))
	}
	if actions[0].ID != hysteresisvm.EventToggleSimulation {
		t.Fatalf("action[0].ID = %q, want %q", actions[0].ID, hysteresisvm.EventToggleSimulation)
	}
	if actions[1].ID != hysteresisvm.EventExportCSV {
		t.Fatalf("action[1].ID = %q, want %q", actions[1].ID, hysteresisvm.EventExportCSV)
	}
	if got := actions[2]; got.ID != hysteresisvm.EventSetWaveform || got.Payload["waveform"] != "triangle" {
		t.Fatalf("action[2] = %#v, want triangle waveform action", got)
	}
}

func TestHysteresisViewDiagnosticExportButtonsDispatchActions(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	var actions []viewmodel.Action

	w := buildHysteresisViewWithActions(snapshot, theme, func(action viewmodel.Action) {
		actions = append(actions, action)
	})
	buttons := collectSidebarButtons(w)
	if len(buttons) < 14 {
		t.Fatalf("hysteresis button count = %d, want diagnostic export controls", len(buttons))
	}

	clickButton(buttons[10])
	clickButton(buttons[11])
	clickButton(buttons[12])
	clickButton(buttons[13])

	wantIDs := []string{"export_pund_csv", "export_forc_sweep_csv", "export_forc_matrix_csv", "export_forc_metadata_json"}
	if len(actions) != len(wantIDs) {
		t.Fatalf("dispatched action count = %d, want %d", len(actions), len(wantIDs))
	}
	for i, want := range wantIDs {
		if actions[i].ID != want {
			t.Fatalf("action[%d].ID = %q, want %q", i, actions[i].ID, want)
		}
	}
}

func TestHysteresisDiagnosticPanelStateFollowsPUNDAndFORC(t *testing.T) {
	vm := hysteresisvm.New()
	if err := vm.ApplyAction(viewmodel.Action{ID: "run_pund", Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run_pund: %v", err)
	}
	if err := vm.ApplyAction(viewmodel.Action{
		ID:      "run_forc",
		Kind:    viewmodel.ActionCommand,
		Payload: map[string]string{"reversals": "13"},
	}); err != nil {
		t.Fatalf("run_forc: %v", err)
	}

	state := hysteresisDiagnosticPanelStateFromSnapshot(vm.Snapshot())
	if !state.pundAvailable {
		t.Fatal("PUND diagnostic state unavailable")
	}
	if !state.forcAvailable {
		t.Fatal("FORC diagnostic state unavailable")
	}
	if !strings.Contains(state.pundSummary, "Switching ratio") {
		t.Fatalf("PUND summary = %q, want switching ratio", state.pundSummary)
	}
	if !strings.Contains(state.forcSummary, "peak_density=") {
		t.Fatalf("FORC summary = %q, want peak density", state.forcSummary)
	}
}
