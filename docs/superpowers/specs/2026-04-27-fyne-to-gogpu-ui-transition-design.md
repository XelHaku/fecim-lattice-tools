# Fyne To gogpu/ui Transition Design

## Purpose

FeCIM Lattice Tools will transition from the current Fyne desktop shell to a future-default `gogpu/ui` shell without breaking the existing public demo path. The migration must preserve scientific correctness, module behavior, documentation honesty, and reproducibility while improving rendering portability and long-term UI control.

The current Fyne application remains the stable release path until the successor shell reaches measured parity. The successor command is named `cmd/fecim-lattice-tools-next` during migration, then becomes `cmd/fecim-lattice-tools` after parity.

## Current State

Before the migration foundation work, the repository targeted Go 1.24 with `toolchain go1.24.12` and depended on Fyne v2.7.2. The repository now requires Go 1.25 or newer. The main Fyne shell lives in `cmd/fecim-lattice-tools/main.go`.

Module embedding is centralized through `shared/widgets/embedded_app.go`:

```go
type EmbeddedApp interface {
	BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject
	Start()
	Stop()
}
```

That interface makes every embeddable module return Fyne widgets directly. Fyne usage is spread across the main shell, module GUI packages, shared widgets, recording widgets, accessibility helpers, theme helpers, recent files, and test utilities. A direct rewrite would be high risk because it would mix UI framework migration with module behavior changes.

## Target State

The future default app uses `gogpu/ui` with a zero-CGO execution path:

```bash
CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next
```

The stable command remains available during transition:

```bash
go run ./cmd/fecim-lattice-tools
```

After parity:

```text
cmd/fecim-lattice-tools/       future default gogpu/ui shell
cmd/fecim-lattice-tools-fyne/  legacy fallback shell for one release
```

The final architecture separates scientific/module state from UI framework code. Physics, validation, export, and simulation packages must not import Fyne or `gogpu/ui`.

## Migration Principles

- Keep Fyne working until the `gogpu/ui` shell has parity.
- Do not port all modules at once.
- Do not change physics behavior while migrating UI.
- Add UI-neutral state and action contracts before rewriting screens.
- Test parity with deterministic state snapshots, exported data, and targeted visual checks.
- Use Go 1.25 or newer for the `gogpu/ui` path.
- Run `gogpu/ui` checks with `CGO_ENABLED=0`.

## Architecture

### Existing Fyne Layer

The Fyne shell continues to use `shared/widgets.EmbeddedApp`. This layer is not deleted or heavily refactored during early migration.

Responsibilities:

- Keep the current public demo stable.
- Continue supporting screenshots, recording, theme preferences, recent files, and current module lifecycle.
- Provide behavioral reference outputs for parity checks.

### New UI-Neutral View Model Layer

Create a new package such as `shared/viewmodel` to describe module state without binding to either Fyne or `gogpu/ui`.

Initial concepts:

```go
type ModuleID string

type ModuleDescriptor struct {
	ID          ModuleID
	Title       string
	Description string
	Status      string
}

type ModuleSnapshot struct {
	Descriptor ModuleDescriptor
	Metrics    []Metric
	Sections   []Section
	Actions    []Action
	UpdatedAt  time.Time
}

type Action struct {
	ID      string
	Label   string
	Kind    ActionKind
	Payload map[string]string
}

type ModulePort interface {
	Descriptor() ModuleDescriptor
	Snapshot() ModuleSnapshot
	ApplyAction(Action) error
	Start()
	Stop()
}
```

The exact types should stay small at first. The goal is not to model every possible widget; the goal is to expose enough state for the next shell to render a real module without importing Fyne.

### Future Default Shell

Create `cmd/fecim-lattice-tools-next` as the successor app.

Responsibilities:

- Own `gogpu.NewApp(...)`, `ui/app.New(...)`, `ggcanvas`, and rendering setup.
- Provide the main navigation, toolbar, module switcher, simulation boundary banner, and help/docs entry point.
- Render module snapshots from `shared/viewmodel`.
- Use `gogpu/ui` primitives and module-specific `gogpu/ui` adapters.

The shell should use event-driven rendering with `WithContinuousRender(false)` unless a module explicitly needs animation.

### Module Adapters

Each module gets an adapter that converts existing module behavior into UI-neutral snapshots. The adapter is introduced before the `gogpu/ui` screen.

Example structure:

```text
module5-comparison/
  pkg/
    model/        UI-neutral scenario logic if available
    gui/          existing Fyne UI
    uiport/       ModulePort adapter and snapshot tests
    gogpu/        future gogpu/ui screen
```

Adapters should avoid duplicating physics or business logic. If existing logic is trapped inside Fyne callbacks, move the logic into a UI-neutral package first, then call it from both UIs.

## Module Port Order

1. **Module 7 Docs**
   - Lowest physics risk.
   - Exercises text, navigation, scroll, search, references, and documentation trust boundaries.
   - Good first test for `gogpu/ui` layout and headless frame testing.

2. **Module 5 Comparison**
   - Mostly dashboard state, scenario controls, evidence panels, and charts.
   - Good candidate for snapshot parity because outputs are already structured.

3. **Module 2 Crossbar**
   - High visual value and strong GPU payoff.
   - Needs heatmaps, IR drop, sneak paths, and deterministic image/pixel checks.

4. **Module 1 Hysteresis**
   - Plot-heavy and physics-sensitive.
   - Requires strict parity against existing golden physics outputs.

5. **Module 4 Circuits**
   - Many controls, logs, overlays, and circuit-state interactions.
   - Port after common controls and status panels are stable.

6. **Module 3 MNIST**
   - Interactive inference workflow and dataset assets.
   - Port after drawing/input patterns are mature.

7. **Module 6 EDA**
   - Export workflows and file validation.
   - Port after dialogs, file operations, progress reporting, and artifact views are stable.

## Testing Strategy

### Unit Tests

Every `shared/viewmodel` type and module adapter gets normal Go tests. These tests must run without Fyne, `gogpu/ui`, GPU access, or a desktop session.

### Parity Tests

For each migrated module:

- Same initial configuration produces the same UI-neutral snapshot.
- Same action sequence produces the same exported data.
- Same simulation seed produces the same metrics.
- Existing Fyne tests remain green.

### `gogpu/ui` Tests

Use headless `ui/app.New()` tests where possible. The docs indicate `gogpu/ui` supports a headless app without window providers for layout/state tests.

Windowed smoke tests should be separated from normal CI unless a runner with display/GPU is available.

### CI

Early migration CI should include:

```bash
go test ./...
CGO_ENABLED=0 go test ./shared/viewmodel/... ./cmd/fecim-lattice-tools-next/...
git diff --check
```

After the Go 1.25 upgrade is complete and stable, the broader zero-CGO path can be expanded.

## Go And Dependency Plan

`gogpu/ui` requires Go 1.25 or newer. The migration foundation moved this repository from its previous Go 1.24 baseline to Go 1.25.

The dependency change must be isolated:

1. Update `go.mod` to Go 1.25.
2. Add `gogpu/ui`, `gogpu`, and `gg`.
3. Add CI coverage for `CGO_ENABLED=0`.
4. Do not remove Fyne.
5. Verify existing Fyne app and tests still pass.

If Fyne or another dependency has an issue with Go 1.25, stop and resolve that before adding the successor shell.

## Cutover Criteria

The `gogpu/ui` shell can become the default only when:

- All seven modules have working `gogpu/ui` screens or accepted replacements.
- All module `Start()` and `Stop()` lifecycle behavior works in the new shell.
- Module navigation and keyboard ownership are deterministic.
- Documentation, simulation boundary banner, and Learn More entry point are present.
- Screenshot or export workflows have replacements or documented exclusions.
- Existing validation and physics regression tests pass.
- A public README command points to the future shell.
- Fyne fallback remains available for one release.

## Deprecation Plan

After cutover:

1. Move the current Fyne shell to `cmd/fecim-lattice-tools-fyne`.
2. Move the `gogpu/ui` shell to `cmd/fecim-lattice-tools`.
3. Update README, AGENTS, CONTRIBUTING, and testing docs.
4. Keep Fyne dependencies for one release.
5. Remove Fyne only after the fallback period and after users confirm the successor shell covers their workflows.

## Risks

- Fyne UI logic is mixed with module behavior in several packages.
- `gogpu/ui` examples and dependency versions may move quickly.
- Interactive GPU tests may not be reliable on generic CI runners.
- Migrating recording, screenshots, dialogs, recent files, accessibility preferences, and shortcuts requires careful replacement.
- A full visual rewrite can accidentally change scientific claims or confidence framing.

## Non-Goals

- No visual redesign before module parity.
- No physics changes bundled with UI migration.
- No immediate Fyne removal.
- No attempt to preserve every Fyne widget abstraction.
- No framework-specific types in physics, validation, export, or simulation packages.

## Initial Implementation Decisions

- The future shell starts with placeholder cards for every module, but only Docs and Comparison become functional in the first module-port phase.
- Screenshots and recording remain Fyne-only until core module parity is reached. The `gogpu/ui` shell must show that these features are not yet ported rather than hiding the gap.
- Visual regression starts with deterministic layout/state tests and selected pixel probes. Generated PNG snapshot comparison is added after the first visual module has a stable render path.
- Module-specific `gogpu/ui` adapters live under each module at `pkg/gogpu`. Shared widget primitives live under `shared/gogpuui` only after duplication appears in at least two module ports.

## Approval

Approved direction:

- `gogpu/ui` is the future default.
- Use `cmd/fecim-lattice-tools-next` during transition.
- Keep current Fyne app stable until parity.
- Cut over only after measured parity.
