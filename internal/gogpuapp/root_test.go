//go:build !cgo

package gogpuapp

import (
	"encoding/binary"
	"hash/fnv"
	"image"
	"testing"

	"github.com/gogpu/gg"
	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	uirender "github.com/gogpu/ui/render"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
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

func TestHeadlessRootClickingSidebarSwitchesRenderedModule(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleDocs)
	previousSignature := harness.renderSignature()
	seen := map[uint64]viewmodel.ModuleID{}

	for _, descriptor := range viewmodel.KnownDescriptors() {
		signature := harness.clickSidebarModule(descriptor.ID)

		if got := harness.activeModuleID(); got != descriptor.ID {
			t.Fatalf("active module after click = %q, want %q", got, descriptor.ID)
		}
		if signature == previousSignature {
			t.Fatalf("render signature did not change after clicking %q", descriptor.ID)
		}
		if prior, exists := seen[signature]; exists {
			t.Fatalf("render signature for %q matched %q", descriptor.ID, prior)
		}
		seen[signature] = descriptor.ID
		previousSignature = signature
	}
}

func TestHeadlessRootClickingCircuitsControlAppliesAction(t *testing.T) {
	model := NewAppModel(viewmodel.ModuleCircuits)
	theme := material3.New(widget.Hex(0x2F5D50))
	app := uiapp.New()
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

	buttons := collectSidebarButtons(app.Window().Root())
	controlOffset := len(viewmodel.KnownDescriptors())
	if len(buttons) <= controlOffset+2 {
		t.Fatalf("root button count = %d, want sidebar buttons plus circuits controls", len(buttons))
	}

	clickButton(buttons[controlOffset+2])

	if got := snapshotMetricValue(model.ActivePort().Snapshot(), "mode"); got != "COMPUTE" {
		t.Fatalf("mode metric after clicking compute = %q, want COMPUTE", got)
	}
}

type headlessModuleSwitchHarness struct {
	t      *testing.T
	app    *uiapp.App
	model  AppModel
	theme  *material3.Theme
	width  int
	height int
}

func newHeadlessModuleSwitchHarness(t *testing.T, active viewmodel.ModuleID) *headlessModuleSwitchHarness {
	t.Helper()

	h := &headlessModuleSwitchHarness{
		t:      t,
		app:    uiapp.New(),
		model:  NewAppModel(active),
		theme:  material3.New(widget.Hex(0x2F5D50)),
		width:  1200,
		height: 760,
	}
	h.app.Window().HandleResize(h.width, h.height)
	h.rebuildRoot()
	return h
}

func (h *headlessModuleSwitchHarness) rebuildRoot() {
	h.app.SetRoot(buildRootWithSelect(h.model, h.theme, func(id viewmodel.ModuleID) {
		if !h.model.SelectModule(id) {
			h.t.Fatalf("select module %q returned false", id)
		}
		h.rebuildRoot()
	}))
}

func (h *headlessModuleSwitchHarness) activeModuleID() viewmodel.ModuleID {
	return h.model.ActiveModuleID
}

func (h *headlessModuleSwitchHarness) clickSidebarModule(id viewmodel.ModuleID) uint64 {
	h.t.Helper()

	buttons := collectSidebarButtons(h.app.Window().Root())
	descriptors := viewmodel.KnownDescriptors()
	if len(buttons) < len(descriptors) {
		h.t.Fatalf("button count = %d, want at least %d sidebar buttons", len(buttons), len(descriptors))
	}

	buttonIndex := -1
	for i, descriptor := range descriptors {
		if descriptor.ID == id {
			buttonIndex = i
			break
		}
	}
	if buttonIndex < 0 {
		h.t.Fatalf("unknown module %q", id)
	}

	bounds := buttons[buttonIndex].ScreenBounds()
	center := geometry.Pt(bounds.Min.X+bounds.Width()/2, bounds.Min.Y+bounds.Height()/2)
	h.app.HandleEvent(event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, center, center, event.ModNone))
	h.app.HandleEvent(event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, center, center, event.ModNone))
	return h.renderSignature()
}

func (h *headlessModuleSwitchHarness) renderSignature() uint64 {
	h.t.Helper()

	h.app.Frame()
	dc := newOffscreenContext(h.width, h.height)
	defer dc.Close()
	dc.ClearWithColor(gg.RGBA{R: 0.96, G: 0.97, B: 0.96, A: 1})
	h.app.Window().DrawTo(uirender.NewCanvas(dc, h.width, h.height))
	if err := dc.FlushGPU(); err != nil {
		h.t.Fatalf("flush rendered root: %v", err)
	}
	return imageSignature(dc.Image())
}

func (h *headlessModuleSwitchHarness) renderActiveFrameSignature() uint64 {
	h.t.Helper()
	return h.renderFrame(func(dc *gg.Context) {
		drawAppFrame(dc, h.app, h.model.ActivePort(), h.width, h.height)
	})
}

func (h *headlessModuleSwitchHarness) renderFrameSignatureWithOverlays(ids ...viewmodel.ModuleID) uint64 {
	h.t.Helper()
	return h.renderFrame(func(dc *gg.Context) {
		drawAppFrame(dc, h.app, nil, h.width, h.height)
		for _, id := range ids {
			drawModuleOverlays(dc, h.portFor(id).Snapshot(), h.width, h.height)
		}
	})
}

func (h *headlessModuleSwitchHarness) renderFrame(draw func(*gg.Context)) uint64 {
	h.t.Helper()

	h.app.Frame()
	dc := newOffscreenContext(h.width, h.height)
	defer dc.Close()
	draw(dc)
	if err := dc.FlushGPU(); err != nil {
		h.t.Fatalf("flush rendered frame: %v", err)
	}
	return imageSignature(dc.Image())
}

func (h *headlessModuleSwitchHarness) portFor(id viewmodel.ModuleID) viewmodel.ModulePort {
	h.t.Helper()
	for _, port := range h.model.Ports {
		if port.Descriptor().ID == id {
			return port
		}
	}
	h.t.Fatalf("missing module port %q", id)
	return nil
}

func imageSignature(img image.Image) uint64 {
	bounds := img.Bounds()
	hash := fnv.New64a()
	var buf [8]byte
	binary.LittleEndian.PutUint32(buf[0:4], uint32(bounds.Dx()))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(bounds.Dy()))
	_, _ = hash.Write(buf[:])

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			binary.LittleEndian.PutUint16(buf[0:2], uint16(r))
			binary.LittleEndian.PutUint16(buf[2:4], uint16(g))
			binary.LittleEndian.PutUint16(buf[4:6], uint16(b))
			binary.LittleEndian.PutUint16(buf[6:8], uint16(a))
			_, _ = hash.Write(buf[:])
		}
	}
	return hash.Sum64()
}

func snapshotMetricValue(snapshot viewmodel.ModuleSnapshot, id string) string {
	for _, metric := range snapshot.Metrics {
		if metric.ID == id {
			return metric.Value
		}
	}
	return ""
}

func TestDrawAppFrameUsesOnlyActiveModuleOverlay(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleDocs)

	harness.clickSidebarModule(viewmodel.ModuleHysteresis)
	hysteresisActiveOverlay := harness.renderActiveFrameSignature()
	hysteresisWithInactiveCrossbar := harness.renderFrameSignatureWithOverlays(viewmodel.ModuleHysteresis, viewmodel.ModuleCrossbar)
	if hysteresisActiveOverlay == hysteresisWithInactiveCrossbar {
		t.Fatal("hysteresis active frame matched a frame with the inactive crossbar overlay")
	}

	harness.clickSidebarModule(viewmodel.ModuleCrossbar)
	crossbarActiveOverlay := harness.renderActiveFrameSignature()
	crossbarWithInactiveHysteresis := harness.renderFrameSignatureWithOverlays(viewmodel.ModuleCrossbar, viewmodel.ModuleHysteresis)
	if crossbarActiveOverlay == crossbarWithInactiveHysteresis {
		t.Fatal("crossbar active frame matched a frame with the inactive hysteresis overlay")
	}
}

func TestDrawAppFrameDrawsCircuitsOverlay(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleCircuits)

	root := harness.renderSignature()
	withOverlay := harness.renderActiveFrameSignature()
	if withOverlay == root {
		t.Fatal("circuits active frame did not draw a module overlay")
	}
}

func TestDrawAppFrameDrawsHysteresisOverlay(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleHysteresis)

	root := harness.renderSignature()
	withOverlay := harness.renderActiveFrameSignature()
	if withOverlay == root {
		t.Fatal("hysteresis active frame did not draw a module overlay")
	}
}

func TestDrawAppFrameDrawsCrossbarOverlay(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleCrossbar)

	root := harness.renderSignature()
	withOverlay := harness.renderActiveFrameSignature()
	if withOverlay == root {
		t.Fatal("crossbar active frame did not draw a module overlay")
	}
}

func TestDrawModuleOverlaysDrawsCircuitsCanvas(t *testing.T) {
	const width = 900
	const height = 640
	dc := newOffscreenContext(width, height)
	defer dc.Close()
	dc.SetRGBA(0.96, 0.97, 0.96, 1)
	dc.DrawRectangle(0, 0, width, height)
	dc.Fill()
	if err := dc.FlushGPU(); err != nil {
		t.Fatalf("flush base frame: %v", err)
	}
	before := imageSignature(dc.Image())

	model := NewAppModel(viewmodel.ModuleCircuits)
	drawModuleOverlays(dc, model.ActivePort().Snapshot(), width, height)
	if err := dc.FlushGPU(); err != nil {
		t.Fatalf("flush circuits overlay: %v", err)
	}
	after := imageSignature(dc.Image())

	if after == before {
		t.Fatal("drawModuleOverlays did not draw the circuits canvas")
	}
}

func TestDrawModuleOverlaysDrawsHysteresisCanvas(t *testing.T) {
	const width = 900
	const height = 640
	dc := newOffscreenContext(width, height)
	defer dc.Close()
	dc.SetRGBA(0.96, 0.97, 0.96, 1)
	dc.DrawRectangle(0, 0, width, height)
	dc.Fill()
	if err := dc.FlushGPU(); err != nil {
		t.Fatalf("flush base frame: %v", err)
	}
	before := imageSignature(dc.Image())

	model := NewAppModel(viewmodel.ModuleHysteresis)
	drawModuleOverlays(dc, model.ActivePort().Snapshot(), width, height)
	if err := dc.FlushGPU(); err != nil {
		t.Fatalf("flush hysteresis overlay: %v", err)
	}
	after := imageSignature(dc.Image())

	if after == before {
		t.Fatal("drawModuleOverlays did not draw the hysteresis canvas")
	}
}

func TestCircuitsOverlayRespondsToViewmodelState(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleCircuits)
	port := harness.portFor(viewmodel.ModuleCircuits)

	before := harness.renderActiveFrameSignature()
	if err := port.ApplyAction(viewmodel.Action{
		ID:   circuitsvm.ActionResizeArray,
		Kind: viewmodel.ActionSelect,
		Payload: map[string]string{
			"rows": "32",
			"cols": "32",
		},
	}); err != nil {
		t.Fatalf("resize circuits array: %v", err)
	}
	if err := port.ApplyAction(viewmodel.Action{
		ID:   circuitsvm.ActionSelectCell,
		Kind: viewmodel.ActionSelect,
		Payload: map[string]string{
			"row": "31",
			"col": "30",
		},
	}); err != nil {
		t.Fatalf("select circuits cell: %v", err)
	}
	if err := port.ApplyAction(viewmodel.Action{ID: circuitsvm.ActionRunCompute, Kind: viewmodel.ActionCommand}); err != nil {
		t.Fatalf("run circuits compute: %v", err)
	}

	after := harness.renderActiveFrameSignature()
	if after == before {
		t.Fatal("circuits overlay did not change after array, selected-cell, and mode state changed")
	}
}

func TestCircuitsOverlayRespondsToHalfSelectStressState(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleCircuits)
	port := harness.portFor(viewmodel.ModuleCircuits)
	if err := port.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetOperationMode,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"mode": circuitsvm.OperationWrite},
	}); err != nil {
		t.Fatalf("set write mode: %v", err)
	}

	passive := harness.renderActiveFrameSignature()
	if err := port.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetArchitecture,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"architecture": circuitsvm.Architecture2T1R},
	}); err != nil {
		t.Fatalf("set 2T1R: %v", err)
	}
	isolated := harness.renderActiveFrameSignature()

	if isolated == passive {
		t.Fatal("circuits overlay did not change after half-select stress state changed from passive to isolated")
	}
}

func TestCircuitsOverlayRespondsToPVTInvestigationState(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleCircuits)
	port := harness.portFor(viewmodel.ModuleCircuits)

	before := harness.renderActiveFrameSignature()
	if err := port.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetADCBits,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"bits": "7"},
	}); err != nil {
		t.Fatalf("set ADC bits: %v", err)
	}
	after := harness.renderActiveFrameSignature()

	if after == before {
		t.Fatal("circuits overlay did not change after PVT ENOB summary changed")
	}
}

func TestCircuitsOverlayRespondsToReferenceSpecState(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleCircuits)
	port := harness.portFor(viewmodel.ModuleCircuits)

	before := harness.renderActiveFrameSignature()
	if err := port.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetDACBits,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"bits": "4"},
	}); err != nil {
		t.Fatalf("set DAC bits: %v", err)
	}
	after := harness.renderActiveFrameSignature()

	if after == before {
		t.Fatal("circuits overlay did not change after reference spec compliance changed")
	}
}

func TestCircuitsOverlayRespondsToReferenceTimingState(t *testing.T) {
	harness := newHeadlessModuleSwitchHarness(t, viewmodel.ModuleCircuits)
	port := harness.portFor(viewmodel.ModuleCircuits)

	before := harness.renderActiveFrameSignature()
	if err := port.ApplyAction(viewmodel.Action{
		ID:      circuitsvm.ActionSetTimingOperation,
		Kind:    viewmodel.ActionSelect,
		Payload: map[string]string{"operation": "WRITE"},
	}); err != nil {
		t.Fatalf("set write timing operation: %v", err)
	}
	after := harness.renderActiveFrameSignature()

	if after == before {
		t.Fatal("circuits overlay did not change after reference timing state changed")
	}
}
