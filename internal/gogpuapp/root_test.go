//go:build !cgo

package gogpuapp

import (
	"os"
	"strings"
	"testing"

	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestBuildRootInstallsInHeadlessApp(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleHysteresis)
	root := buildRoot(model, material3.New(widget.Hex(0x2F5D50)))

	app := uiapp.New()
	app.SetRoot(root)
	app.Frame()

	if app.Window().Root() == nil {
		t.Fatal("root widget was not installed")
	}
}

func TestBuildRoot_RendersWithRealComparisonPort(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleComparison)

	var foundComparison bool
	for _, p := range model.Ports {
		if p.Descriptor().ID == viewmodel.ModuleComparison {
			foundComparison = true
			break
		}
	}
	if !foundComparison {
		t.Fatal("BuildAppPorts did not include a comparison port")
	}

	root := buildRoot(model, material3.New(widget.Hex(0x2F5D50)))
	if root == nil {
		t.Fatal("buildRoot returned nil")
	}

	app := uiapp.New()
	app.SetRoot(root)
	app.Frame()
	if app.Window().Root() == nil {
		t.Fatal("dispatch through buildComparisonView dropped the root widget")
	}
}

func TestBuildRootUsesRequestedActiveModule(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleCrossbar)
	if got := model.ActivePort().Descriptor().ID; got != viewmodel.ModuleCrossbar {
		t.Fatalf("ActivePort = %q, want %q", got, viewmodel.ModuleCrossbar)
	}

	root := buildRoot(model, material3.New(widget.Hex(0x2F5D50)))
	if root == nil {
		t.Fatal("buildRoot returned nil")
	}
}

func TestOverlayRenderingUsesActiveModuleSnapshot(t *testing.T) {
	body, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	text := string(body)

	forbidden := []string{
		"var globalPorts",
		"globalPorts = model.Ports",
		"for _, port := range globalPorts",
	}
	for _, phrase := range forbidden {
		if strings.Contains(text, phrase) {
			t.Errorf("overlay rendering must not use package-global module ports: found %q", phrase)
		}
	}

	required := []string{
		"activePort := model.ActivePort()",
		"drawModuleOverlays(cc, activePort.Snapshot(), cw, ch)",
		"func drawModuleOverlays(cc *gg.Context, snapshot viewmodel.ModuleSnapshot, w, h int)",
		"func drawCrossbarOverlay(cc *gg.Context, snapshot viewmodel.ModuleSnapshot, rows, cols, w, h int)",
	}
	for _, phrase := range required {
		if !strings.Contains(text, phrase) {
			t.Errorf("overlay rendering must be active-snapshot scoped: missing %q", phrase)
		}
	}
}
