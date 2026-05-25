//go:build !cgo

package gogpuapp

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModule1HysteresisDefaultGogpuUIEndToEnd(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleDocs)
	docsSignature := harness.renderActiveFrameSignature()

	hysteresisSignature := harness.clickSidebarModule(viewmodel.ModuleHysteresis)
	if got := harness.activeModuleID(); got != viewmodel.ModuleHysteresis {
		t.Fatalf("active module after sidebar navigation = %q, want %q", got, viewmodel.ModuleHysteresis)
	}
	if hysteresisSignature == docsSignature {
		t.Fatal("hysteresis sidebar navigation did not change the rendered gogpu/ui frame")
	}

	snapshot := harness.model.ActivePort().Snapshot()
	if snapshot.Descriptor.ID != viewmodel.ModuleHysteresis {
		t.Fatalf("snapshot descriptor = %q, want %q", snapshot.Descriptor.ID, viewmodel.ModuleHysteresis)
	}
	if !strings.Contains(snapshot.Descriptor.BoundaryNotice, "SIMULATION OUTPUT") || !strings.Contains(snapshot.Descriptor.BoundaryNotice, "Not measured device data") {
		t.Fatalf("hysteresis boundary notice does not clearly frame simulation limits: %q", snapshot.Descriptor.BoundaryNotice)
	}
	plot := snapshotPlotByID(snapshot, "pe_loop")
	if len(plot.Series) == 0 || len(plot.Series[0].Points) < 100 {
		t.Fatalf("hysteresis P-E plot has %d series with too few points", len(plot.Series))
	}
	if snapshotMetricValue(snapshot, "waveform") == "" {
		t.Fatal("hysteresis snapshot omitted waveform metric")
	}

	rootOnly := harness.renderSignature()
	withOverlay := harness.renderActiveFrameSignature()
	if withOverlay == rootOnly {
		t.Fatal("hysteresis gogpu/ui frame did not render the Module 1 overlay")
	}

	img, err := CaptureFrameImage(viewmodel.ModuleHysteresis, 900, 640)
	if err != nil {
		t.Fatalf("CaptureFrameImage(ModuleHysteresis): %v", err)
	}
	_, colors := offscreenImageSignatureAndColorCount(img)
	if colors < 8 {
		t.Fatalf("captured Module 1 frame appears blank: %d unique colors", colors)
	}
}

func TestModule1HysteresisDefaultGogpuUIDiagnosticsEndToEnd(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)
	theme := material3.New(widget.Hex(0x2F5D50))
	app := uiapp.New()
	app.Window().HandleResize(1200, 760)
	var rebuildRoot func()
	rebuildRoot = func() {
		app.SetRoot(buildRootWithSelectAndActions(model, theme, nil, func(action viewmodel.Action) {
			if err := model.ActivePort().ApplyAction(action); err != nil {
				t.Fatalf("ApplyAction(%s): %v", action.ID, err)
			}
			rebuildRoot()
		}))
		app.Frame()
	}
	rebuildRoot()

	before := renderHeadlessAppFrameSignature(t, app, model.ActivePort())
	buttons := collectSidebarButtons(app.Window().Root())
	controlOffset := len(viewmodel.KnownDescriptors())
	if len(buttons) <= controlOffset+9 {
		t.Fatalf("root button count = %d, want sidebar buttons plus PUND/FORC controls", len(buttons))
	}

	clickButton(buttons[controlOffset+8])
	buttons = collectSidebarButtons(app.Window().Root())
	clickButton(buttons[controlOffset+9])
	after := renderHeadlessAppFrameSignature(t, app, model.ActivePort())
	if after == before {
		t.Fatal("running Module 1 diagnostics did not change the rendered gogpu/ui frame")
	}

	snapshot := model.ActivePort().Snapshot()
	if plot := snapshotPlotByID(snapshot, "pund_current_waveforms"); len(plot.Series) == 0 {
		t.Fatal("Run PUND control did not expose PUND current waveform plot")
	}
	if plot := snapshotPlotByID(snapshot, "forc_density_heatmap"); len(plot.Series) == 0 {
		t.Fatal("Run FORC control did not expose FORC density heatmap plot")
	}
	if !snapshotHasSection(snapshot, "diagnostic_pund") || !snapshotHasSection(snapshot, "diagnostic_forc") {
		t.Fatalf("diagnostic controls did not expose PUND and FORC summary sections: %#v", snapshot.Sections)
	}
}

func renderHeadlessAppFrameSignature(t *testing.T, app *uiapp.App, port viewmodel.ModulePort) uint64 {
	t.Helper()
	app.Frame()
	dc := newOffscreenContext(1200, 760)
	defer dc.Close()
	drawAppFrame(dc, app, port, 1200, 760)
	if err := dc.FlushGPU(); err != nil {
		t.Fatalf("flush rendered Module 1 frame: %v", err)
	}
	return imageSignature(dc.Image())
}

func snapshotHasSection(snapshot viewmodel.ModuleSnapshot, id string) bool {
	for _, section := range snapshot.Sections {
		if section.ID == id {
			return true
		}
	}
	return false
}

func TestDefaultGogpuHysteresisSurfaceHasNoFyneDependencies(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "./cmd/fecim-lattice-tools", "./internal/gogpuapp")
	cmd.Dir = "../.."
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list default gogpu/ui surface failed:\n%s", out)
	}

	for _, forbidden := range []string{
		"fyne.io/fyne",
		"fecim-lattice-tools/module1-hysteresis/pkg/gui",
		"fecim-lattice-tools/module1-hysteresis/cmd/hysteresis-fyne",
	} {
		if strings.Contains(string(out), forbidden) {
			t.Fatalf("default gogpu/ui dependency graph contains deprecated Fyne surface %q:\n%s", forbidden, out)
		}
	}
}
