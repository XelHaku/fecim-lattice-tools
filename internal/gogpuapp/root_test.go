//go:build !cgo

package gogpuapp

import (
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
		t.Fatal("BuildPlaceholderPorts did not include a comparison port")
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
