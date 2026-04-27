//go:build !cgo

package main

import (
	"testing"

	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildRootInstallsInHeadlessApp(t *testing.T) {
	spec := DefaultAppSpec()
	ports := BuildPlaceholderPorts()
	root := buildRoot(spec, ports, material3.New(widget.Hex(0x2F5D50)))

	app := uiapp.New()
	app.SetRoot(root)
	app.Frame()

	if app.Window().Root() == nil {
		t.Fatal("root widget was not installed")
	}
}
