# Fyne Conventions for FeCIM Lattice Tools

---

## FeCIM Project Rules

These rules are project-specific and take precedence over the generic Fyne guidance below.

### 1. Abbreviation Prefixes (Project-Specific)

- `wrd*` prefix = "Write-Read-Demo" -- fields managing the ISPP write-verify-read demo cycle (WaveformWriteReadDemo). OK to use in Module 1 where established, but always add the section comment: `// Write-Read-Demo (wrd) fields`
- `ispp*` prefix = "Incremental Step Pulse Programming" -- fields for ISPP algorithm state
- New abbreviations require approval and documentation in this file

### 2. Threading Contract (CRITICAL)

- Every goroutine MUST have a comment explaining what closes it:
  - Simulation loops: `// Runs until a.running.Store(false) on window close.`
  - One-shot goroutines: `// Runs once at startup; exits after calibration completes.`
- ALL widget/canvas method calls from goroutines MUST be wrapped in `fyne.Do()`:
  - `widget.SetText()`, `widget.Refresh()`, `Select.SetSelected()`, etc.
  - EXCEPTION: sends to a channel consumed by the UI goroutine are OK
  - Reading/writing state protected by `a.mu` does NOT need `fyne.Do()`
- Never use `time.Sleep()` inside a goroutine to defer a UI update. Use dialog callbacks, channels, or immediate synchronous calls instead.

### 3. Enum Constant Naming

- ALL enum constants must be prefixed with the type name:
  - CORRECT: `WaveformTypeSine` or `WaveformSine` (when type is `WaveformType`)
  - WRONG: `ModeWrite` when the type is `OperationMode`
- This applies even for internal/unexported types

### 4. App Struct Field Organization

Order fields in this section sequence with header comments:

```go
// Core logic components
// Physics / simulation state
// Simulation state (ephemeral per-step)
// Write-Read-Demo (wrd) fields       <-- use this exact header
// ISPP convergence algorithm state
// Calibration data
// UI state
// UI components                       <-- widget fields go here
// UI update throttling
// Demo/statistics metrics
```

### 5. Layout Conventions

- Every `SetOffset()` call must have an inline comment: `// X% panel-name, Y% panel-name`
- Every `SetMinSize()` call must have a comment if the size is not obvious
- Always use `shared/widgets.NewAdaptiveLayout` for modules that need both desktop and mobile layouts
- Prefer `container.NewBorder` for fixed-header/footer patterns over VBox+Spacer

### 6. File Size Guidelines

- Warn at 800 lines; aim to split before 1,000 lines
- Hard limit: 1,200 lines per file (document exceptions with a reason comment at top)
- `module4-circuits/pkg/gui/tab_unified.go` (2,205 lines) is a known exception -- tracked as tech debt

### 7. Shared Widgets First

- Before creating any new widget, check `shared/widgets/` (70+ files)
- Key available widgets: `AdaptiveLayout`, `ColorLegend`, `ModeIndicator`, `StatusBar`, `EducationalPanel`, `Tooltips`, `DebugPanel`

### 8. Constructor Pattern

- All exported types: `NewFoo() *Foo`
- Alternative constructors: `NewFooWithBar(bar Bar) *Foo`
- Never export a type without a `New*` constructor

---

## Fyne Core Naming Conventions

Fyne projects follow a consistent naming style that aligns with Go idioms. Adhering to these makes your code predictable and easier to collaborate on.

- **General Rules**
  - Use **American English** for all identifiers.
  - Prefer **complete words** over abbreviations, except for very local variables like `a` (for `fyne.App`), `w` (for `fyne.Window`), or loop indices.
  - Be descriptive: `temperatureEntry` is better than `tempEnt`.

- **Constructors**
  - Every exported type should have a constructor function prefixed with `New`.
    ```go
    func NewTemperatureChart() *TemperatureChart { ... }
    ```
  - For alternative creation patterns, use `With...` suffixes (e.g., `NewChartWithBounds`).

- **Callbacks and Event Handlers**
  - Use the `On` prefix followed by a past-tense verb to indicate "when this happens, do that".
    Examples:
    - `OnTapped` for a button click
    - `OnChanged` for a widget's data change
    - `OnClosed` for a window close event

- **Enumerations**
  - Define a dedicated type for constants created with `iota`.
  - Prefix each constant with the type name for namespace-like clarity.
    ```go
    type SensorType int

    const (
        SensorTypeTemperature SensorType = iota
        SensorTypeHumidity
        SensorTypePressure
    )
    ```

---

## Project Architecture & Structural Standards

For a multi-screen app with complex state and plots, a clean architecture separates UI from logic and makes testing easier.

### Recommended Pattern: MVVM with Fyne Bindings

Fyne's `binding` package supports the Model-View-ViewModel pattern. This keeps views reactive and business logic independent.

- **Model**: Contains data structures, business logic, and external communication. Should have no dependency on Fyne.
- **ViewModel**: Holds the state of a screen using `binding` types (`binding.String`, `binding.Float`, `binding.Bool`, etc.). Exposes methods the View can call. When the ViewModel updates a binding, the UI automatically refreshes.
  ```go
  type DashboardViewModel struct {
      Temperature binding.Float
      Humidity    binding.Float
      IsLoading   binding.Bool
  }

  func NewDashboardViewModel() *DashboardViewModel {
      return &DashboardViewModel{
          Temperature: binding.NewFloat(),
          Humidity:    binding.NewFloat(),
          IsLoading:   binding.NewBool(),
      }
  }

  func (vm *DashboardViewModel) RefreshData() {
      vm.IsLoading.Set(true)
      go func() {
          // fetch data from model
          vm.Temperature.Set(23.5)
          vm.Humidity.Set(60.0)
          vm.IsLoading.Set(false)
      }()
  }
  ```

- **View**: Built with Fyne widgets and containers. Receives a ViewModel and binds UI elements to its properties. Contains no business logic -- only layout and binding code.
  ```go
  func MakeDashboardScreen(vm *DashboardViewModel) fyne.CanvasObject {
      tempLabel := widget.NewLabelWithData(
          binding.FloatToStringWithFormat(vm.Temperature, "%.1f °C"))
      refreshBtn := widget.NewButton("Refresh", vm.RefreshData)
      return container.NewVBox(tempLabel, refreshBtn)
  }
  ```

### Package Organization

Separate concerns by splitting code into packages to avoid circular dependencies and improve readability.

```
your-app/
├── cmd/
│   └── your-app/               # main package
│       └── main.go
├── internal/                    # private application code
│   ├── model/                   # data models and business logic
│   ├── viewmodel/               # ViewModels for each screen
│   └── ui/                      # UI layer (views and reusable components)
│       ├── components/          # custom widgets (e.g., gauges, charts)
│       └── screens/             # screen builders
│           ├── dashboard.go
│           ├── settings.go
│           └── ...
├── assets/                      # images, fonts, etc.
├── go.mod
└── FyneApp.toml                 # metadata (app ID, name, version)
```

- `cmd/your-app/main.go` is the only place where `fyne.App` is created and windows are shown.
- `internal/` prevents other projects from importing your internal packages.
- `ui/screens/` contains files like `dashboard.go` that export `MakeDashboardScreen(vm *DashboardViewModel) fyne.CanvasObject`.
- `ui/components/` holds reusable pieces (e.g., a `Plot` widget) that can be used across screens.

### Managing State Across Screens

- Use **ViewModels per screen**, but share common data through a **global app state** (e.g., an `AppModel` struct) injected into each ViewModel.
- For inter-screen communication, consider using events (e.g., a simple pub/sub bus) or rely on the shared model.

---

## Practical Development Tips

1. **Use `FyneApp.toml`**: Create this file in your project root with your app's metadata. Required for proper app packaging and versioning.

2. **Run `goimports`**: Always format code with `goimports -w .` before committing. Ensures consistent style and sorted imports.

3. **Keep UI Construction Functions Pure**: Functions like `MakeDashboardScreen` should only create and return a `fyne.CanvasObject`. All logic belongs in the ViewModel or Model.

4. **Organize Files Consistently**: Within each Go file, follow the standard order: `package` -> `import` -> `const` -> `var` -> `type` -> `func` (constructor first, then exported methods, then unexported methods).

5. **Leverage Fyne's Layout System**: For responsive designs, use `container.NewBorder`, `container.NewGridWithColumns`, etc. Avoid absolute positioning.

6. **Performance Considerations with Plots**: Use `canvas.Raster` or `widget.Canvas` for custom drawing, but refresh only when data changes. Reuse existing objects rather than recreating them.

7. **Testing**: ViewModels can be tested with standard Go tests because they don't depend on Fyne. UI components can be tested using `test.NewApp()` and `test.NewWindow()` from `fyne.io/fyne/v2/test`.

---

## Advanced Architecture: The App Controller Pattern

As your app grows, `main.go` shouldn't just create windows; it should initialize a Controller or Router that manages transitions between screens and holds the long-lived application context.

### The App Context

Instead of passing many arguments to every screen, create an `AppContext` struct.

```go
type AppContext struct {
    App         fyne.App
    MainWindow  fyne.Window
    Store       *model.DataStore
    Preferences *model.UserPrefs
    Theme       *ui.CustomTheme
}
```

### The Router/Navigator

In a multi-screen app, you need a way to swap the main content without closing the window.

```go
type Router struct {
    ctx     *AppContext
    topBar  *fyne.Container
    content *fyne.Container
}

func (r *Router) NavigateTo(screenName string) {
    var newScreen fyne.CanvasObject
    switch screenName {
    case "dashboard":
        vm := viewmodel.NewDashboardVM(r.ctx.Store)
        newScreen = screens.MakeDashboard(vm)
    }
    r.content.Objects = []fyne.CanvasObject{newScreen}
    r.content.Refresh()
}
```

---

## Plotting and High-Performance Custom Widgets

For plots and many inputs, standard widgets might not be enough. You may need custom widgets.

### Custom Widget Structural Convention

When creating a plot widget, separate the **Widget** (state) from the **Renderer** (drawing logic). This is Fyne's internal requirement for performance.

- **Widget Struct**: Holds properties (data points, min/max values).
- **Renderer Struct**: Holds the actual `canvas.Line` or `canvas.Raster` objects.
- **Naming**: `PlotWidget` and `plotRenderer`.

For plots with thousands of points, do not use `canvas.Line` for every point. Use `canvas.Raster` to draw pixels directly to a buffer -- significantly faster for real-time data.

---

## ViewModel & Data Binding

For many inputs, manual synchronization is error-prone. Expand your `binding` strategy to include validation.

### Validation Convention

Prefix validation functions with `Validate`. Use Fyne's `widget.Entry` with `Validator`.

```go
func (vm *SettingsViewModel) ValidateIPAddress(input string) error {
    if !isValidIP(input) {
        return errors.New("invalid IP format")
    }
    return nil
}

// In the View:
entry := widget.NewEntryWithData(vm.IPBinding)
entry.Validator = vm.ValidateIPAddress
```

### List Data Binding

For screens with many repetitive items (e.g., a list of sensors), use `binding.UntypedList`. Use the suffix `List` for the binding and `Item` for the individual objects.

---

## Asset & Theme Management

Large apps should not hardcode colors or icons. Use Go's `embed` package.

```
assets/
├── icons/
│   ├── check.svg
│   └── warning.svg
├── fonts/
│   └── Roboto-Regular.ttf
└── bundled.go
```

```go
//go:embed icons/*.svg
var iconFS embed.FS

func GetIcon(name string) fyne.Resource {
    data, _ := iconFS.ReadFile("icons/" + name + ".svg")
    return fyne.NewStaticResource(name, data)
}
```

---

## Summary of Key Naming Conventions

| Element           | Convention                     | Example                          |
|-------------------|--------------------------------|----------------------------------|
| Constructor       | `New` + TypeName               | `NewTemperatureChart`            |
| Callback field    | `On` + PastTenseVerb           | `OnTapped`, `OnChanged`          |
| Enum type         | Descriptive type name          | `type Unit string`               |
| Enum constants    | TypeName + CamelCase           | `UnitCelsius`, `UnitFahrenheit`  |
| Binding variable  | Use binding types              | `binding.Float`                  |
| File name         | lower_case_with_underscores    | `dashboard_screen.go`            |

## Expanded Conventions

| Category | Convention | Reasoning |
|----------|-----------|-----------|
| **Interfaces** | `internal/interfaces` | Define "Repository" or "Service" interfaces here to mock for testing. |
| **Long Tasks** | `func (vm *VM) Action(ctx context.Context)` | Always pass a context to allow cancelling background data fetches. |
| **Async UI** | `fyne.Do(func())` | Use this inside goroutines when you need to update a non-bound UI element. |
| **Files** | `_test.go` per package | Aim for 80% coverage on `viewmodel` and `model` packages. |
