# Fyne To gogpu/ui Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first safe migration slice for making `gogpu/ui` the future default while keeping the existing Fyne app stable.

**Architecture:** Keep `cmd/fecim-lattice-tools` as the current Fyne shell and add `cmd/fecim-lattice-tools-next` as the future-default `gogpu/ui` shell. Introduce `shared/viewmodel` as the UI-neutral contract that future module ports will use, then render placeholder module cards from that contract.

**Tech Stack:** Go 1.25, Fyne v2.7.2 kept in place, `github.com/gogpu/ui`, `github.com/gogpu/gogpu`, `github.com/gogpu/gg`, standard Go tests, GitHub Actions.

---

## Scope

This plan implements only the foundation slice:

- Move the repo toolchain to Go 1.25.
- Add `shared/viewmodel` without importing Fyne or `gogpu/ui`.
- Add `cmd/fecim-lattice-tools-next` as a `gogpu/ui` shell with placeholder cards for all seven modules.
- Add a zero-CGO CI check for the future shell.
- Update docs to make the current and future commands explicit.

This plan does not port Module 7 Docs or Module 5 Comparison yet. Those need their own module-specific plans after this foundation lands.

## File Structure

- Create `shared/viewmodel/types.go`
  - Owns UI-neutral module descriptors, snapshots, sections, metrics, actions, and the `ModulePort` interface.
- Create `shared/viewmodel/static_module.go`
  - Provides deterministic placeholder/static module ports for the next shell.
- Create `shared/viewmodel/types_test.go`
  - Proves descriptors, snapshots, and unsupported actions work without any UI framework.
- Create `cmd/fecim-lattice-tools-next/appmodel.go`
  - Owns the future shell app spec and module placeholder registry.
- Create `cmd/fecim-lattice-tools-next/main.go`
  - Real `gogpu/ui` entry point, compiled only when CGO is disabled.
- Create `cmd/fecim-lattice-tools-next/main_cgo.go`
  - Guard entry point that explains `CGO_ENABLED=0`.
- Create `cmd/fecim-lattice-tools-next/appmodel_test.go`
  - Tests app spec and placeholder module coverage without a desktop window.
- Create `cmd/fecim-lattice-tools-next/root_test.go`
  - Headless `gogpu/ui` test for root widget construction, compiled only when CGO is disabled.
- Modify `go.mod` and `go.sum`
  - Move to Go 1.25 and add the `gogpu/ui` dependency stack.
- Modify `.github/workflows/ci.yml`
  - Use Go 1.25 and add `CGO_ENABLED=0` future shell tests.
- Modify `.github/workflows/nightly.yml`
  - Use Go 1.25 and add the same future shell test lane.
- Modify `Makefile`
  - Add `test-next-ui`.
- Modify `README.md`, `AGENTS.md`, and `CONTRIBUTING.md`
  - Document both commands and the zero-CGO future shell path.

---

### Task 1: Add UI-Neutral View Model Package

**Files:**
- Create: `shared/viewmodel/types_test.go`
- Create: `shared/viewmodel/types.go`
- Create: `shared/viewmodel/static_module.go`

- [ ] **Step 1: Write the failing viewmodel tests**

Create `shared/viewmodel/types_test.go`:

```go
package viewmodel

import (
	"errors"
	"testing"
)

func TestKnownDescriptorsCoverSevenModules(t *testing.T) {
	descriptors := KnownDescriptors()

	if len(descriptors) != 7 {
		t.Fatalf("descriptor count = %d, want 7", len(descriptors))
	}

	wantIDs := []ModuleID{
		ModuleHysteresis,
		ModuleCrossbar,
		ModuleMNIST,
		ModuleCircuits,
		ModuleComparison,
		ModuleEDA,
		ModuleDocs,
	}
	for i, want := range wantIDs {
		if descriptors[i].ID != want {
			t.Fatalf("descriptor[%d].ID = %q, want %q", i, descriptors[i].ID, want)
		}
		if descriptors[i].Title == "" {
			t.Fatalf("descriptor[%d] has empty title", i)
		}
		if descriptors[i].Description == "" {
			t.Fatalf("descriptor[%d] has empty description", i)
		}
	}
}

func TestStaticModuleSnapshotIsDeterministic(t *testing.T) {
	descriptor := ModuleDescriptor{
		ID:          ModuleDocs,
		Title:       "Documentation",
		Description: "Documentation and validation references.",
		Status:      StatusPlaceholder,
	}
	port := NewStaticModule(descriptor, []Section{
		{
			ID:    "scope",
			Title: "Scope",
			Body:  "Placeholder card for future gogpu/ui module port.",
		},
	})

	snapshot := port.Snapshot()

	if snapshot.Descriptor != descriptor {
		t.Fatalf("snapshot descriptor = %#v, want %#v", snapshot.Descriptor, descriptor)
	}
	if len(snapshot.Sections) != 1 {
		t.Fatalf("section count = %d, want 1", len(snapshot.Sections))
	}
	if snapshot.Sections[0].ID != "scope" {
		t.Fatalf("section ID = %q, want scope", snapshot.Sections[0].ID)
	}
	if !snapshot.UpdatedAt.IsZero() {
		t.Fatalf("UpdatedAt = %v, want zero for deterministic placeholder", snapshot.UpdatedAt)
	}
}

func TestStaticModuleRejectsActions(t *testing.T) {
	port := NewStaticModule(ModuleDescriptor{
		ID:          ModuleComparison,
		Title:       "Comparison",
		Description: "Comparison placeholder.",
		Status:      StatusPlaceholder,
	}, nil)

	err := port.ApplyAction(Action{ID: "run"})
	if !errors.Is(err, ErrUnsupportedAction) {
		t.Fatalf("ApplyAction error = %v, want ErrUnsupportedAction", err)
	}
}
```

- [ ] **Step 2: Run the test to verify RED**

Run:

```bash
go test ./shared/viewmodel
```

Expected: FAIL because `shared/viewmodel` has no implementation and symbols such as `KnownDescriptors`, `ModuleID`, and `NewStaticModule` are undefined.

- [ ] **Step 3: Add the viewmodel types**

Create `shared/viewmodel/types.go`:

```go
package viewmodel

import "time"

type ModuleID string

const (
	ModuleHysteresis ModuleID = "hysteresis"
	ModuleCrossbar   ModuleID = "crossbar"
	ModuleMNIST      ModuleID = "mnist"
	ModuleCircuits   ModuleID = "circuits"
	ModuleComparison ModuleID = "comparison"
	ModuleEDA        ModuleID = "eda"
	ModuleDocs       ModuleID = "docs"
)

const (
	StatusPlaceholder = "placeholder"
	StatusFunctional  = "functional"
	StatusFallback    = "fyne-fallback"
)

type ModuleDescriptor struct {
	ID          ModuleID
	Title       string
	Description string
	Status      string
}

type Metric struct {
	ID         string
	Label      string
	Value      string
	Unit       string
	Confidence string
}

type Section struct {
	ID    string
	Title string
	Body  string
}

type ActionKind string

const (
	ActionCommand ActionKind = "command"
	ActionToggle  ActionKind = "toggle"
	ActionSelect  ActionKind = "select"
)

type Action struct {
	ID      string
	Label   string
	Kind    ActionKind
	Payload map[string]string
}

type ModuleSnapshot struct {
	Descriptor ModuleDescriptor
	Metrics    []Metric
	Sections   []Section
	Actions    []Action
	UpdatedAt  time.Time
}

type ModulePort interface {
	Descriptor() ModuleDescriptor
	Snapshot() ModuleSnapshot
	ApplyAction(Action) error
	Start()
	Stop()
}

func KnownDescriptors() []ModuleDescriptor {
	return []ModuleDescriptor{
		{
			ID:          ModuleHysteresis,
			Title:       "FeCIM Hysteresis Simulation",
			Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
			Status:      StatusPlaceholder,
		},
		{
			ID:          ModuleCrossbar,
			Title:       "FeCIM Crossbar Array Visualization",
			Description: "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.",
			Status:      StatusPlaceholder,
		},
		{
			ID:          ModuleMNIST,
			Title:       "FeCIM MNIST Neural Network",
			Description: "Educational CIM inference pipeline with quantized weights and reproducible metrics.",
			Status:      StatusPlaceholder,
		},
		{
			ID:          ModuleCircuits,
			Title:       "FeCIM Peripheral Circuits Visualizer",
			Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
			Status:      StatusPlaceholder,
		},
		{
			ID:          ModuleComparison,
			Title:       "FeCIM Comparison",
			Description: "Evidence-first technology comparison and scenario analysis.",
			Status:      StatusPlaceholder,
		},
		{
			ID:          ModuleEDA,
			Title:       "FeCIM EDA Design Suite",
			Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
			Status:      StatusPlaceholder,
		},
		{
			ID:          ModuleDocs,
			Title:       "Documentation",
			Description: "Curriculum, validation references, trust boundaries, and research notes.",
			Status:      StatusPlaceholder,
		},
	}
}
```

- [ ] **Step 4: Add the static module implementation**

Create `shared/viewmodel/static_module.go`:

```go
package viewmodel

import (
	"errors"
)

var ErrUnsupportedAction = errors.New("viewmodel: unsupported action")

type StaticModule struct {
	descriptor ModuleDescriptor
	sections   []Section
	metrics    []Metric
	actions    []Action
}

func NewStaticModule(descriptor ModuleDescriptor, sections []Section) *StaticModule {
	return &StaticModule{
		descriptor: descriptor,
		sections:   append([]Section(nil), sections...),
	}
}

func (m *StaticModule) Descriptor() ModuleDescriptor {
	return m.descriptor
}

func (m *StaticModule) Snapshot() ModuleSnapshot {
	return ModuleSnapshot{
		Descriptor: m.descriptor,
		Metrics:    append([]Metric(nil), m.metrics...),
		Sections:   append([]Section(nil), m.sections...),
		Actions:    append([]Action(nil), m.actions...),
	}
}

func (m *StaticModule) ApplyAction(Action) error {
	return ErrUnsupportedAction
}

func (m *StaticModule) Start() {}

func (m *StaticModule) Stop() {}
```

- [ ] **Step 5: Verify GREEN**

Run:

```bash
go test ./shared/viewmodel
```

Expected: PASS.

- [ ] **Step 6: Commit Task 1**

Run:

```bash
git add shared/viewmodel
git commit -m "feat(viewmodel): add UI-neutral module contract"
```

---

### Task 2: Add Future Shell App Model Without UI Dependencies

**Files:**
- Create: `cmd/fecim-lattice-tools-next/appmodel_test.go`
- Create: `cmd/fecim-lattice-tools-next/appmodel.go`

- [ ] **Step 1: Write the failing app model tests**

Create `cmd/fecim-lattice-tools-next/appmodel_test.go`:

```go
package main

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestDefaultAppSpecNamesFutureDefaultShell(t *testing.T) {
	spec := DefaultAppSpec()

	if spec.Title != "FeCIM Lattice Tools Next" {
		t.Fatalf("Title = %q, want FeCIM Lattice Tools Next", spec.Title)
	}
	if spec.Command != "fecim-lattice-tools-next" {
		t.Fatalf("Command = %q, want fecim-lattice-tools-next", spec.Command)
	}
	if spec.Width < 1200 || spec.Height < 800 {
		t.Fatalf("unexpected window size: %dx%d", spec.Width, spec.Height)
	}
}

func TestBuildPlaceholderPortsCoversAllKnownDescriptors(t *testing.T) {
	ports := BuildPlaceholderPorts()
	descriptors := viewmodel.KnownDescriptors()

	if len(ports) != len(descriptors) {
		t.Fatalf("port count = %d, want %d", len(ports), len(descriptors))
	}

	for i, port := range ports {
		got := port.Descriptor()
		want := descriptors[i]
		if got != want {
			t.Fatalf("port[%d] descriptor = %#v, want %#v", i, got, want)
		}
		snapshot := port.Snapshot()
		if len(snapshot.Sections) == 0 {
			t.Fatalf("port[%d] snapshot has no sections", i)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify RED**

Run:

```bash
go test ./cmd/fecim-lattice-tools-next
```

Expected: FAIL because `DefaultAppSpec` and `BuildPlaceholderPorts` are undefined.

- [ ] **Step 3: Add the app model implementation**

Create `cmd/fecim-lattice-tools-next/appmodel.go`:

```go
package main

import "fecim-lattice-tools/shared/viewmodel"

type AppSpec struct {
	Title   string
	Command string
	Width   int
	Height  int
}

func DefaultAppSpec() AppSpec {
	return AppSpec{
		Title:   "FeCIM Lattice Tools Next",
		Command: "fecim-lattice-tools-next",
		Width:   1400,
		Height:  900,
	}
}

func BuildPlaceholderPorts() []viewmodel.ModulePort {
	descriptors := viewmodel.KnownDescriptors()
	ports := make([]viewmodel.ModulePort, 0, len(descriptors))
	for _, descriptor := range descriptors {
		ports = append(ports, viewmodel.NewStaticModule(descriptor, []viewmodel.Section{
			{
				ID:    "migration-status",
				Title: "Migration Status",
				Body:  "This module is represented by a UI-neutral placeholder while the gogpu/ui shell reaches parity with the current Fyne implementation.",
			},
		}))
	}
	return ports
}
```

- [ ] **Step 4: Verify GREEN**

Run:

```bash
go test ./cmd/fecim-lattice-tools-next
```

Expected: PASS.

- [ ] **Step 5: Commit Task 2**

Run:

```bash
git add cmd/fecim-lattice-tools-next/appmodel.go cmd/fecim-lattice-tools-next/appmodel_test.go
git commit -m "feat(next-ui): add future shell app model"
```

---

### Task 3: Upgrade Toolchain And Add gogpu Dependencies

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`

- [ ] **Step 1: Update the Go directive**

Run:

```bash
go mod edit -go=1.25.0 -toolchain=go1.25.0
```

Expected: `go.mod` changes from `go 1.24.0` and `toolchain go1.24.12` to Go 1.25.

- [ ] **Step 2: Add the gogpu dependency stack**

Run:

```bash
go get github.com/gogpu/ui@v0.1.13 github.com/gogpu/gogpu@v0.29.4 github.com/gogpu/gg@v0.43.2
go mod tidy
```

Expected: `go.mod` includes direct requirements for:

```go
github.com/gogpu/gg v0.43.2
github.com/gogpu/gogpu v0.29.4
github.com/gogpu/ui v0.1.13
```

- [ ] **Step 3: Verify the existing app still builds under the default path**

Run:

```bash
go test -short ./cmd/fecim-lattice-tools ./shared/viewmodel/... ./cmd/fecim-lattice-tools-next/...
```

Expected: PASS.

- [ ] **Step 4: Commit Task 3**

Run:

```bash
git add go.mod go.sum
git commit -m "chore: upgrade Go toolchain for gogpu UI"
```

---

### Task 4: Add the gogpu/ui Successor Shell

**Files:**
- Create: `cmd/fecim-lattice-tools-next/main.go`
- Create: `cmd/fecim-lattice-tools-next/main_cgo.go`
- Create: `cmd/fecim-lattice-tools-next/root_test.go`

- [ ] **Step 1: Write the failing headless root test**

Create `cmd/fecim-lattice-tools-next/root_test.go`:

```go
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
```

- [ ] **Step 2: Run the test to verify RED**

Run:

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next
```

Expected: FAIL because `buildRoot` is undefined.

- [ ] **Step 3: Add the zero-CGO gogpu/ui entry point**

Create `cmd/fecim-lattice-tools-next/main.go`:

```go
//go:build !cgo

package main

import (
	"log"

	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	uiapp "github.com/gogpu/ui/app"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/render"
	uitheme "github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func main() {
	spec := DefaultAppSpec()
	ports := BuildPlaceholderPorts()

	gpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle(spec.Title).
		WithSize(spec.Width, spec.Height).
		WithContinuousRender(false))

	seed := widget.Hex(0x2F5D50)
	appTheme := uitheme.DefaultLight()
	appTheme.Colors.Primary = seed
	appTheme.Colors.PrimaryDark = widget.Hex(0x1F463C)
	appTheme.Colors.PrimaryLight = widget.Hex(0x6F9C8D)
	materialTheme := material3.New(seed)

	app := uiapp.New(
		uiapp.WithWindowProvider(gpuApp),
		uiapp.WithPlatformProvider(gpuApp),
		uiapp.WithEventSource(gpuApp.EventSource()),
		uiapp.WithTheme(appTheme),
	)
	app.SetRoot(buildRoot(spec, ports, materialTheme))

	var canvas *ggcanvas.Canvas
	gpuApp.OnDraw(func(dc *gogpu.Context) {
		w, h := dc.Width(), dc.Height()
		if w <= 0 || h <= 0 {
			return
		}
		if canvas == nil {
			provider := gpuApp.GPUContextProvider()
			if provider == nil {
				return
			}
			var err error
			canvas, err = ggcanvas.New(provider, w, h)
			if err != nil {
				log.Printf("ggcanvas: %v", err)
				return
			}
		}

		app.Frame()
		cw, ch := canvas.Size()
		if cw != w || ch != h {
			if err := canvas.Resize(w, h); err != nil {
				log.Printf("resize: %v", err)
				return
			}
			cw, ch = w, h
		}

		if err := canvas.Draw(func(cc *gg.Context) {
			cc.SetRGBA(0.96, 0.97, 0.96, 1)
			cc.DrawRectangle(0, 0, float64(cw), float64(ch))
			cc.Fill()
			app.Window().DrawTo(render.NewCanvas(cc, cw, ch))
		}); err != nil {
			log.Printf("draw: %v", err)
			return
		}
		if err := canvas.Render(dc.RenderTarget()); err != nil {
			log.Printf("render: %v", err)
		}
	})
	gpuApp.OnClose(func() { gg.CloseAccelerator() })

	if err := gpuApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func buildRoot(spec AppSpec, ports []viewmodel.ModulePort, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(spec.Title).FontSize(28).Bold(),
		primitives.Text("Future default gogpu/ui shell. Current module cards are placeholders until parity with the Fyne app is reached.").FontSize(15),
		primitives.Text("Stable fallback remains: go run ./cmd/fecim-lattice-tools").FontSize(13),
	}

	for _, port := range ports {
		snapshot := port.Snapshot()
		children = append(children, moduleCard(snapshot, theme))
	}

	return primitives.Box(children...).
		Padding(28).
		Gap(14).
		Background(theme.Colors.Surface)
}

func moduleCard(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	descriptor := snapshot.Descriptor
	body := descriptor.Description
	if len(snapshot.Sections) > 0 && snapshot.Sections[0].Body != "" {
		body = body + "\n" + snapshot.Sections[0].Body
	}

	return primitives.Box(
		primitives.Text(descriptor.Title).FontSize(18).Bold(),
		primitives.Text(string(descriptor.ID)+" | "+descriptor.Status).FontSize(12),
		primitives.Text(body).FontSize(14),
	).
		Padding(16).
		Gap(8).
		Background(theme.Colors.SurfaceContainer)
}
```

- [ ] **Step 4: Add the CGO guard entry point**

Create `cmd/fecim-lattice-tools-next/main_cgo.go`:

```go
//go:build cgo

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "fecim-lattice-tools-next uses gogpu/ui through the zero-CGO WebGPU stack.")
	fmt.Fprintln(os.Stderr, "Run with: CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next")
	os.Exit(2)
}
```

- [ ] **Step 5: Verify GREEN**

Run:

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next
```

Expected: PASS.

- [ ] **Step 6: Verify the zero-CGO shell builds**

Run:

```bash
CGO_ENABLED=0 go build -o /tmp/fecim-lattice-tools-next ./cmd/fecim-lattice-tools-next
```

Expected: PASS and `/tmp/fecim-lattice-tools-next` exists.

- [ ] **Step 7: Verify the CGO guard path**

Run:

```bash
go run ./cmd/fecim-lattice-tools-next
```

Expected: exit code 2 with:

```text
fecim-lattice-tools-next uses gogpu/ui through the zero-CGO WebGPU stack.
Run with: CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next
```

- [ ] **Step 8: Commit Task 4**

Run:

```bash
git add cmd/fecim-lattice-tools-next
git commit -m "feat(next-ui): add gogpu shell placeholders"
```

---

### Task 5: Add CI And Makefile Future Shell Checks

**Files:**
- Modify: `Makefile`
- Modify: `.github/workflows/ci.yml`
- Modify: `.github/workflows/nightly.yml`

- [ ] **Step 1: Add the Makefile target**

Modify `Makefile` by adding `test-next-ui` to `.PHONY` and adding the target after `test-short`:

```make
test-next-ui:
	CGO_ENABLED=0 $(GO) test ./shared/viewmodel/... ./cmd/fecim-lattice-tools-next/...
```

Also add this line to the `help` target output:

```make
	@echo "  make test-next-ui    Run zero-CGO tests for the future gogpu/ui shell"
```

- [ ] **Step 2: Verify the Makefile target**

Run:

```bash
make test-next-ui
```

Expected: PASS.

- [ ] **Step 3: Update CI to Go 1.25 and test the future shell**

Modify `.github/workflows/ci.yml`:

```yaml
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
```

Add this step after `Run make ci`:

```yaml
      - name: Future gogpu UI shell tests
        run: make test-next-ui
```

Modify the `external-validation` job Go setup to:

```yaml
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
```

- [ ] **Step 4: Update nightly to Go 1.25 and test the future shell**

Modify `.github/workflows/nightly.yml` Go setup to:

```yaml
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
```

Add this step after `Build & vet`:

```yaml
      - name: Future gogpu UI shell tests
        run: make test-next-ui
```

- [ ] **Step 5: Verify workflow YAML and Makefile changes locally**

Run:

```bash
make test-next-ui
git diff --check
```

Expected: PASS with no whitespace errors.

- [ ] **Step 6: Commit Task 5**

Run:

```bash
git add Makefile .github/workflows/ci.yml .github/workflows/nightly.yml
git commit -m "ci: add zero-CGO future UI checks"
```

---

### Task 6: Document Current And Future UI Commands

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `CONTRIBUTING.md`

- [ ] **Step 1: Update README**

Add this section after the current Quick Start or Getting Started commands:

````markdown
## Future UI Shell

The current default desktop app is still the Fyne shell:

```bash
go run ./cmd/fecim-lattice-tools
```

The future default shell is being built with `gogpu/ui`:

```bash
CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next
```

The `next` shell currently shows UI-neutral placeholder cards for all modules. Fyne remains the stable public demo path until the `gogpu/ui` shell reaches measured parity.
````

- [ ] **Step 2: Update AGENTS.md**

Add this note near the build/test instructions:

````markdown
## Future UI Migration

`cmd/fecim-lattice-tools` remains the stable Fyne shell. `cmd/fecim-lattice-tools-next` is the future-default `gogpu/ui` shell and must be checked with:

```bash
make test-next-ui
CGO_ENABLED=0 go build -o /tmp/fecim-lattice-tools-next ./cmd/fecim-lattice-tools-next
```

Do not import Fyne or `gogpu/ui` from physics, validation, export, or simulation packages. Use `shared/viewmodel` for UI-neutral module state.
````

- [ ] **Step 3: Update CONTRIBUTING.md**

Add this note to local checks:

````markdown
For changes touching the future `gogpu/ui` shell or UI-neutral module contracts, also run:

```bash
make test-next-ui
CGO_ENABLED=0 go build -o /tmp/fecim-lattice-tools-next ./cmd/fecim-lattice-tools-next
```
````

- [ ] **Step 4: Verify docs formatting**

Run:

```bash
git diff --check
```

Expected: PASS.

- [ ] **Step 5: Commit Task 6**

Run:

```bash
git add README.md AGENTS.md CONTRIBUTING.md
git commit -m "docs: document future gogpu UI shell"
```

---

### Task 7: Final Verification And PR Preparation

**Files:**
- No new files unless verification reveals a defect.

- [ ] **Step 1: Run focused tests**

Run:

```bash
go test ./shared/viewmodel ./cmd/fecim-lattice-tools-next
CGO_ENABLED=0 go test ./shared/viewmodel/... ./cmd/fecim-lattice-tools-next/...
```

Expected: PASS.

- [ ] **Step 2: Run repository CI path locally**

Run:

```bash
make ci
```

Expected: PASS.

- [ ] **Step 3: Run future shell build check**

Run:

```bash
CGO_ENABLED=0 go build -o /tmp/fecim-lattice-tools-next ./cmd/fecim-lattice-tools-next
```

Expected: PASS.

- [ ] **Step 4: Run whitespace check**

Run:

```bash
git diff --check
```

Expected: PASS.

- [ ] **Step 5: Review dependency impact**

Run:

```bash
go list -m all | rg 'gogpu|go-webgpu|fyne'
```

Expected: output includes Fyne and the new `gogpu` stack. Fyne must still be present.

- [ ] **Step 6: Commit any verification-only fixes**

If Task 7 required fixes, commit them:

```bash
git add .
git commit -m "fix(next-ui): stabilize foundation checks"
```

If Task 7 required no fixes, do not create an empty commit.

- [ ] **Step 7: Push branch**

Run:

```bash
git push
```

Expected: branch pushes successfully.

- [ ] **Step 8: Include PR evidence**

Use this PR evidence format:

```markdown
## TDD Evidence

- RED:
  - `go test ./shared/viewmodel` failed before `shared/viewmodel` implementation.
  - `go test ./cmd/fecim-lattice-tools-next` failed before app model implementation.
  - `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next` failed before `buildRoot` implementation.
- GREEN:
  - `go test ./shared/viewmodel`
  - `go test ./cmd/fecim-lattice-tools-next`
  - `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next`
- Final verification:
  - `go test ./shared/viewmodel ./cmd/fecim-lattice-tools-next`
  - `CGO_ENABLED=0 go test ./shared/viewmodel/... ./cmd/fecim-lattice-tools-next/...`
  - `make ci`
  - `CGO_ENABLED=0 go build -o /tmp/fecim-lattice-tools-next ./cmd/fecim-lattice-tools-next`
  - `git diff --check`
```

---

## Self-Review Checklist

- This plan keeps the Fyne app as the stable default.
- This plan adds `gogpu/ui` only to the future shell path.
- This plan introduces a UI-neutral contract before any module UI rewrite.
- This plan does not change physics behavior.
- This plan provides RED/GREEN checks for behavior changes.
- This plan avoids GPU/windowed smoke tests in normal CI.
- This plan makes the first migration PR reviewable and reversible.
