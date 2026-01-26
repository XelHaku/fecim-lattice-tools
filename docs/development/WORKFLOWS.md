# Development Workflows - FeCIM Lattice Tools

This document outlines development workflows for the FeCIM Lattice Tools project, a Go-based visualization suite for Ferroelectric Compute-in-Memory (FeCIM) systems built with Fyne GUI framework.

## Table of Contents

1. [Development Setup](#development-setup)
2. [Building and Running](#building-and-running)
3. [Adding a New Module](#adding-a-new-module)
4. [Testing Workflow](#testing-workflow)
5. [Debugging](#debugging)
6. [Code Review Checklist](#code-review-checklist)
7. [Git Conventions](#git-conventions)
8. [Troubleshooting](#troubleshooting)

---

## Development Setup

### Prerequisites

**Go:**
- Go 1.21+ (supports modern language features)
- Check version: `go version`

**System Dependencies:**

Linux (Ubuntu/Debian):
```bash
sudo apt-get install -y \
  gcc \
  libgl1-mesa-dev \
  libx11-dev \
  libxinerama-dev \
  libxrandr-dev \
  libxcursor-dev \
  libxi-dev \
  libxext-dev \
  libxfixes-dev
```

macOS:
```bash
# Xcode Command Line Tools (includes clang)
xcode-select --install
```

Windows:
- Install MinGW-w64 or Microsoft C++ Build Tools
- Ensure `gcc` is in PATH

**Optional Tools:**
- `ffmpeg` (for video recording feature in GUI)
- `git` (version control)

### Clone and Build

```bash
# Clone the repository
git clone <repository-url>
cd fecim-lattice-tools

# Build the main visualizer
go build -o fecim-visualizer ./cmd/fecim-visualizer

# Or use the launch script (convenience wrapper)
./launch.sh
```

### Directory Structure

```
fecim-lattice-tools/
├── cmd/
│   └── fecim-visualizer/      # Main unified app entry point
├── module1-hysteresis/         # P-E curve simulation, Preisach model
├── module2-crossbar/           # Crossbar array, MVM, non-idealities
├── module3-mnist/              # MNIST neural network demo
├── module4-circuits/           # Peripheral circuits (DAC/ADC/TIA)
├── module5-comparison/         # Technology comparison & market analysis
├── module6-eda/                # EDA design suite
├── shared/                     # Shared theme, widgets, logging
└── docs/
    ├── development/            # Developer documentation
    ├── videos/                 # Video transcripts & references
    └── archive/                # Historical notes
```

### Development Environment Setup

Create `.env` file (optional, not committed):
```bash
# Logging verbosity: 0|off, 1|info, 2|debug, 3|trace
export FECIM_VERBOSITY=2

# Debug GUI layout issues (Fyne-specific)
export FYNE_DEBUG_RESIZE=0

# Disable GPU rendering (useful for headless environments)
export FYNE_NO_GL=0

# Theme (native system theme preferred)
export FYNE_THEME=dark
```

Load before running:
```bash
source .env
./fecim-visualizer -verbosity debug
```

---

## Building and Running

### Build Variants

**Release Build (Optimized):**
```bash
go build -ldflags="-s -w" -o fecim-visualizer ./cmd/fecim-visualizer
```

**Debug Build (With Symbols):**
```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer
```

**Build with Race Detector:**
```bash
# Detects concurrent memory access issues
go build -race -o fecim-visualizer ./cmd/fecim-visualizer
```

### Run the Application

**Direct Execution:**
```bash
./fecim-visualizer
```

**With Logging:**
```bash
./fecim-visualizer -verbosity info    # Info level logging
./fecim-visualizer -verbosity debug   # Debug level logging
./fecim-visualizer -verbosity trace   # Trace level logging
```

**Via Launch Script:**
```bash
./launch.sh
```

The script handles:
- Automatic rebuild if sources changed
- Sets reasonable defaults
- Captures exit codes

### Application Features

Once running, the unified visualizer provides:
- **Module 1:** Hysteresis simulation (P-E curves, write/read demo)
- **Module 2:** Crossbar array visualization (MVM, IR drop, sneak paths)
- **Module 3:** MNIST neural network (inference, training comparison)
- **Module 4:** Peripheral circuits (DAC/ADC/TIA simulation)
- **Module 5:** Technology comparison (market positioning)
- **Module 6:** EDA design suite (WIP)

**Recording & Screenshots:**
- Screenshot button: Captures current view as PNG
- Record button: Starts/stops video recording (20 FPS, H.264)
- Files saved to `./screenshots/` and `./recordings/`

---

## Adding a New Module

### Step 1: Create Module Directory Structure

```bash
moduleN-name/
├── cmd/                        # CLI entry points (optional)
│   └── moduleN-name-gui/
│       └── main.go
├── pkg/
│   ├── physics/                # Physics simulation package
│   │   ├── physics.go
│   │   ├── physics_test.go
│   │   └── constants.go
│   ├── gui/                    # Fyne GUI components
│   │   ├── embedded.go         # REQUIRED: Embedded app interface
│   │   ├── app.go              # Main UI logic
│   │   ├── widgets/            # Custom UI widgets
│   │   │   ├── plot.go
│   │   │   └── controls.go
│   │   └── gui_test.go         # UI tests
│   └── core/                   # Core algorithms (optional)
│       └── core.go
└── README.md                   # Module-specific documentation
```

### Step 2: Implement the Embedded App Interface

Create `/moduleN-name/pkg/gui/embedded.go`:

```go
package gui

import "fyne.io/fyne/v2"

// EmbeddedApp wraps the module's main app for embedding in the unified visualizer
type EmbeddedApp struct {
	*App  // Compose your main App struct
}

// NewEmbeddedApp creates and initializes the embedded demo
func NewEmbeddedApp() *EmbeddedApp {
	// Initialize physics simulation
	// Create default UI state
	// Return wrapped instance
	return &EmbeddedApp{
		App: &App{
			// Initialize fields
		},
	}
}

// BuildContent constructs the UI canvas for the tab
// Called by the main visualizer with the shared Fyne app and window
func (e *EmbeddedApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject {
	e.fyneApp = fyneApp
	e.mainWindow = parentWindow

	// Build and return the UI container
	// Example:
	// return container.NewVBox(
	//     e.buildControls(),
	//     e.buildVisualization(),
	// )

	return nil // Replace with actual UI
}

// Start activates the simulation/demo
// Called when the tab is selected
// Initialize goroutines, timers, animation loops here
func (e *EmbeddedApp) Start() {
	// Start simulation loop
	// Begin animations
	// Initialize any goroutines
}

// Stop deactivates the simulation/demo
// Called when switching away from the tab or closing the app
// Cleanup: stop goroutines, free resources
func (e *EmbeddedApp) Stop() {
	// Stop all goroutines
	// Clean up resources
	// Flush any pending state
}
```

### Step 3: Register Module in Main Visualizer

Edit `/cmd/fecim-visualizer/main.go`:

**1. Import the module:**
```go
import (
	// ... existing imports
	moduleNgui "multilayer-ferroelectric-cim-visualizer/moduleN-name/pkg/gui"
)
```

**2. Add to DemoApp struct:**
```go
type DemoApp struct {
	demo1 *demo1gui.EmbeddedApp
	// ... existing demos
	demoN *moduleNgui.EmbeddedApp  // Add this
}
```

**3. Create the demo instance (in main()):**
```go
// Around line 410, after creating other demos:
fmt.Println("[STARTUP] Creating demoN (module name)...")
demoN := moduleNgui.NewEmbeddedApp()
fmt.Println("[STARTUP] demoN created")

// Update DemoApp struct:
demos := &DemoApp{
	demo1: d1,
	// ...
	demoN: demoN,  // Add this
}
```

**4. Add to view names and build UI:**
```go
// Line ~423, add to viewNames slice:
viewNames := []string{
	"Home",
	"FeCIM Hysteresis Simulation",
	// ... existing names
	"Module N: Your Module Name",  // Add this
}

// Line ~484, build content:
fmt.Println("[STARTUP] Building demoN content...")
demoNContent := demos.demoN.BuildContent(fyneApp, window)
fmt.Println("[STARTUP] demoN content built")

// Line ~498, add to views slice:
views = []fyne.CanvasObject{
	launcherContent,
	container.NewMax(demo1Content),
	// ... existing views
	container.NewMax(demoNContent),  // Add this
}
```

**5. Add to onViewChange callback (line ~646-702):**
```go
// In the stop switch:
case N:
	log.Debug("Stopping demoN")
	demos.demoN.Stop()

// In the start switch:
case N:
	log.Debug("Starting demoN")
	demos.demoN.Start()
```

**6. Add cleanup on app close (line ~784-793):**
```go
// In SetCloseIntercept callback:
demos.demoN.Stop()

// In cleanup on exit:
demos.demoN.Stop()
```

### Step 4: Write Tests

Create `/moduleN-name/pkg/physics/physics_test.go`:

```go
package physics

import (
	"testing"
)

// TestBasicFunctionality verifies core physics
func TestBasicFunctionality(t *testing.T) {
	// Test the main physics computation
}

// TestEdgeCases covers boundary conditions
func TestEdgeCases(t *testing.T) {
	// Test extreme values, zero, negative, etc.
}
```

**Run tests:**
```bash
go test ./moduleN-name/...
```

### Step 5: Update Documentation

Create `/moduleN-name/README.md`:

```markdown
# Module N: Your Module Name

## Description

Brief description of what this module demonstrates.

## Physics Model

Explain the underlying physics:
- Key parameters and equations
- Constants and their sources
- Verification status

## UI Components

- **Control 1:** Description
- **Chart 1:** Description

## Files

- `pkg/physics/physics.go` - Core physics simulation
- `pkg/gui/embedded.go` - Embedded app interface
- `pkg/gui/app.go` - Main UI logic

## Testing

```bash
go test ./...
```

## References

- [Source 1](url)
- [Source 2](url)
```

---

## Testing Workflow

### Test Structure

The project uses Go's standard `testing` package with 117 tests across 13 packages:

```
✅ module1-hysteresis/pkg/ferroelectric       7 tests
✅ module1-hysteresis/pkg/simulation           4 tests
✅ module2-crossbar/pkg/crossbar              29 tests
✅ module3-mnist/pkg/core                     33 tests
✅ module3-mnist/pkg/training                  3 tests
✅ module4-circuits/pkg/peripherals            9 tests
✅ module5-comparison/pkg/comparison          19 tests
✅ shared/logging                              6 tests
✅ shared/theme                                5 tests
✅ shared/widgets                              1 test
✅ module6-eda/pkg/export                      3 tests
✅ cmd/fecim-visualizer                        2 tests
────────────────────────────────────
✅ TOTAL                                     117 tests (100% PASS)
```

### Running Tests

**Run All Tests:**
```bash
go test ./...
```

**Run With Verbose Output:**
```bash
go test -v ./...
```

**Run Specific Package:**
```bash
go test -v ./module2-crossbar/pkg/crossbar
```

**Run Specific Test:**
```bash
go test -v -run TestQuantizationCliff ./module3-mnist/pkg/core
```

**Run With Race Detector (Concurrency Safety):**
```bash
go test -race ./...
```

**Run Benchmarks:**
```bash
go test -bench=. ./module2-crossbar/pkg/crossbar
```

**Run Tests With Coverage:**
```bash
go test -cover ./...
```

**Generate Coverage Report:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Opens in browser
```

### GUI Test Pattern

GUI tests use `t.Skip()` for headless CI environments:

```go
func TestGUIComponent(t *testing.T) {
	// Skip in headless CI (no display server)
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("Skipping GUI test in headless environment")
	}

	// GUI test logic...
}
```

### Test Categories

#### Physics Verification Tests

**Location:** `module2-crossbar/pkg/crossbar/physics_test.go`, etc.

Verify FeCIM physics accuracy:
- 30-level quantization specification
- IR drop calculations (Ohm's law)
- Sneak path signal-to-noise ratio
- Drift and retention characteristics

Example test structure:
```go
func TestIRDropOhmsLaw(t *testing.T) {
	// Arrange: Set up circuit with known resistance
	// Act: Calculate voltage drop
	// Assert: Verify V = IR
}
```

#### Integration Tests

**Location:** `module3-mnist/pkg/core/integration_test.go`

Test multiple systems together:
- Full inference pipeline (784→128→10)
- Concurrent operations
- Configuration presets

#### Mock Tests

For GUI components, create mocks where needed:
```go
type MockFyneApp struct {
	// Minimal interface implementation for testing
}

func (m *MockFyneApp) Preferences() fyne.Preferences {
	// Return test preferences
}
```

### Before Committing

**Required:**
1. Run all tests: `go test ./...`
2. Run with race detector: `go test -race ./...`
3. Verify no new test failures
4. Check coverage for changed code

**Command to run before commit:**
```bash
go test -race -v ./... && echo "✅ All tests passed"
```

---

## Debugging

### Logging

The project uses a shared logging package (`shared/logging/`):

```go
import "multilayer-ferroelectric-cim-visualizer/shared/logging"

// Get logger for your package
var log = logging.NewLogger("module-name")

// Use logging methods:
log.Debug("Debug message: %v", value)
log.Info("Info message")
log.Warn("Warning message")
log.Error("Error: %v", err)
```

**Set verbosity at runtime:**
```bash
./fecim-visualizer -verbosity debug
./fecim-visualizer -verbosity trace
```

**Programmatically:**
```go
logging.SetVerbosity(logging.DebugLevel)
```

### GUI Layout Debugging

Enable Fyne layout debugging with environment variable:

```bash
FYNE_DEBUG_RESIZE=1 ./fecim-visualizer
```

This logs window resize events and layout calculations. Useful for:
- Tracking down oscillation loops on Wayland/Sway
- Verifying MinSize calculations
- Debugging custom layout implementations

### Race Detector

Catch concurrent memory access issues during development:

```bash
# Build with race detector
go build -race -o fecim-visualizer ./cmd/fecim-visualizer

# Run and reproduce the issue
./fecim-visualizer

# Race detector will output stack traces of data races
```

**Key areas to watch for races:**
- UI updates from goroutines (wrap in `fyne.Do()`)
- Physics simulation updates
- State mutations in demo lifecycle

### Thread-Safe UI Updates

**CRITICAL: All UI updates from goroutines MUST use `fyne.Do()`**

```go
// ❌ WRONG - Data race
go func() {
	label.SetText("Updated")  // Race condition!
}()

// ✅ CORRECT - Thread-safe
go func() {
	fyne.Do(func() {
		label.SetText("Updated")  // Safe on main thread
	})
}()
```

### Debugging Common Issues

#### Window Resize Loops

**Symptom:** Window constantly resizing on Wayland/Sway

**Cause:** MinSize calculations causing feedback loops

**Fix:** Use `ForceMinSizeLayout` (see `cmd/fecim-visualizer/main.go`, line 52)

#### Graphics Driver Issues

**Symptom:** Black screen, rendering errors

**Solution:**
```bash
# Try software rendering (disables GPU)
FYNE_NO_GL=1 ./fecim-visualizer

# Check OpenGL support
glxinfo | grep "OpenGL version"
```

#### FFmpeg Recording Failures

**Symptom:** Recording starts but produces empty file

**Check:**
```bash
# Verify ffmpeg installed
which ffmpeg
ffmpeg -version

# Check file permissions in recordings/ directory
ls -la recordings/
```

#### Goroutine Leaks

**Detect:**
```bash
# Use pprof to inspect goroutines at runtime
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

**Profile current state:**
```go
import _ "net/http/pprof"

// In main():
go func() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Then visit: http://localhost:6060/debug/pprof/
```

---

## Code Review Checklist

Before submitting a pull request, verify:

### Thread Safety

- [ ] All UI updates from goroutines wrapped in `fyne.Do()`?
- [ ] No unguarded access to shared mutable state?
- [ ] Use sync.Mutex for critical sections?
- [ ] Tests pass with `-race` flag?

### Physics & Accuracy

- [ ] Physics constants have citations (Dr. Tour, Nature Commun., etc.)?
- [ ] 30-level quantization verified?
- [ ] Edge cases tested (zero, negative, extreme values)?
- [ ] Units documented (MV/cm, µC/cm², etc.)?

### Error Handling

- [ ] No panics for user input errors?
- [ ] Errors wrapped with context: `fmt.Errorf("context: %w", err)`?
- [ ] Error messages are user-friendly?
- [ ] Recoverable errors logged, not fataled?

### Testing

- [ ] New code has tests?
- [ ] All 117 tests pass?
- [ ] No test failures with `-race`?
- [ ] Coverage maintained or improved?

### Code Style

- [ ] Follows Go conventions (CamelCase, etc.)?
- [ ] Comments explain "why", not "what"?
- [ ] Package-level comments present?
- [ ] Imports organized: stdlib, external, internal?

### Documentation

- [ ] Changes documented in README or CLAUDE.md?
- [ ] Complex algorithms explained?
- [ ] Physics model documented with sources?
- [ ] New modules in `docs/development/WORKFLOWS.md`?

### Commits

- [ ] Commits follow format: `type: description`
  - `feat:` new feature
  - `fix:` bug fix
  - `docs:` documentation
  - `refactor:` code restructuring
  - `test:` test additions
  - `chore:` build, dependencies, etc.
- [ ] One logical change per commit?
- [ ] Commit messages describe intent?

### Performance

- [ ] No new memory leaks?
- [ ] Reasonable CPU usage?
- [ ] UI remains responsive?
- [ ] Long-running operations in goroutines?

---

## Git Conventions

### Commit Format

```
type: brief description (50 chars max)

Optional longer explanation (wrap at 72 chars).
Can include:
- Motivation for change
- Implementation details
- References to issues: Fixes #123
```

**Types:**
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation change
- `refactor:` Code restructuring (no behavior change)
- `test:` Test additions/fixes
- `chore:` Build, dependencies, tooling
- `perf:` Performance improvement

**Examples:**
```
feat: add module 4 peripheral circuits visualization

docs: update WORKFLOWS.md with testing procedures

fix: resolve race condition in hysteresis demo stop

test: add 15 new physics validation tests

refactor: extract quantization logic to shared package

chore: update go.mod dependencies
```

### Branch Naming

Branches should be descriptive and lowercase:

```
feature/module-name          # New module
feature/improve-performance  # Enhancement
bugfix/fix-description       # Bug fix
docs/update-workflows        # Documentation
```

### Pull Request Process

1. **Branch off main:**
   ```bash
   git checkout -b feature/your-feature
   ```

2. **Make changes with regular commits:**
   ```bash
   git add .
   git commit -m "feat: description"
   ```

3. **Run tests before pushing:**
   ```bash
   go test -race ./...
   ```

4. **Push and create PR:**
   ```bash
   git push origin feature/your-feature
   ```

5. **PR checklist:**
   - [ ] Tests pass locally
   - [ ] No race condition warnings
   - [ ] Follows code review checklist
   - [ ] Commit messages descriptive
   - [ ] Ready for review

---

## Troubleshooting

### Build Errors

**"Package not found"**
```bash
# Update module cache
go mod tidy

# Verify dependencies
go mod graph | head -20
```

**"Undefined: something"**
```bash
# Likely import path issue - check module name in go.mod
cat go.mod | head -5

# All module paths should match
# multilayer-ferroelectric-cim-visualizer/module1-hysteresis
```

**CGO/gcc errors**
```bash
# Verify system dependencies installed
sudo apt-get install -y gcc libgl1-mesa-dev libx11-dev

# Or on macOS:
xcode-select --install
```

### Runtime Errors

**"Segmentation fault"**
- Usually graphics driver issue
- Try: `FYNE_NO_GL=1 ./fecim-visualizer`
- Update GPU drivers

**"panic: runtime error: index out of range"**
- Add logging to identify the exact line
- Use `go build` without optimizations for better error messages
- Check array bounds before access

**Demo fails to initialize**
- Check logs: `./fecim-visualizer -verbosity debug`
- Verify package imports are correct
- Ensure embedded app interface fully implemented

### Performance Issues

**High CPU usage:**
```bash
# Profile CPU usage
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Check for busy loops without sleep
# Look for 50ms frame rates: time.Sleep(50 * time.Millisecond)
```

**Memory growth:**
```bash
# Profile memory
go tool pprof http://localhost:6060/debug/pprof/heap

# Look for goroutine leaks
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

**Slow simulation:**
- Check quantization step (should use precomputed 30 levels)
- Verify history buffers don't exceed needed size
- Profile hot paths with `go test -bench`

### GUI Issues

**Blurry text on high-DPI displays:**
- Fyne handles this automatically
- Update to latest Fyne version if issues persist

**Theme not applying:**
```go
// Verify in main():
fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})

// Check theme resources are loaded
```

**Tab content disappears:**
- Ensure `BuildContent()` returns non-nil CanvasObject
- Verify Start/Stop don't prematurely hide UI
- Check container sizing is explicit

### Recording Not Working

**FFmpeg not found:**
```bash
# Install ffmpeg
sudo apt-get install ffmpeg    # Linux
brew install ffmpeg            # macOS

# Verify
ffmpeg -version
```

**Recording produces silent video:**
- Project only records video, not audio
- Audio sync would require audio stream integration

**Recording is too slow/choppy:**
- Canvas capture happens on main thread
- Currently uses 20 FPS balance
- Reduce recording frequency if needed

---

## Getting Help

### Documentation References

| Need | Location |
|------|----------|
| Function lookups | `docs/development/scriptReference.md` |
| Error resolution | `docs/development/scriptReference.md#error-resolution-guide` |
| Physics details | `docs/development/scriptReference.md` |
| Fyne GUI patterns | `docs/development/FYNE_NOTES.md` |
| Testing procedures | `docs/development/TESTING.md` |
| Physics accuracy report | `docs/development/HYPER_ANALYSIS_REPORT.md` |
| Project rules | `CLAUDE.md` |

### Quick Checks

```bash
# Verify environment is ready
go version                    # Should be 1.21+
gcc --version                 # Should work
git status                    # Should show clean or expected changes

# Run quick test suite
go test ./... -race           # All tests should pass

# Check formatting
go fmt ./...                  # Format all code
go vet ./...                  # Static analysis
```

### Common Workflows

**Add feature to module:**
1. Edit `moduleN/pkg/physics/physics.go` or `gui/app.go`
2. Add tests in `*_test.go`
3. Run: `go test -race ./moduleN/...`
4. Commit: `git commit -m "feat: description"`

**Fix physics bug:**
1. Add failing test that reproduces issue
2. Fix implementation
3. Verify test passes: `go test -run TestName`
4. Commit: `git commit -m "fix: description"`

**Add new visualization:**
1. Create widget in `moduleN/pkg/gui/widgets/`
2. Implement fyne.CanvasObject interface
3. Add to UI in `BuildContent()`
4. Test manually: `./fecim-visualizer`
5. Commit: `git commit -m "feat: description"`

**Performance tuning:**
1. Profile with pprof: `go test -cpuprofile=cpu.prof -bench=.`
2. Analyze: `go tool pprof cpu.prof`
3. Make targeted optimization
4. Verify improvement with benchmark
5. Commit: `git commit -m "perf: description"`

---

## Related Documentation

- **CLAUDE.md** - Project-level rules and constants
- **TESTING.md** - Detailed test documentation
- **FYNE_NOTES.md** - Fyne GUI framework patterns
- **scriptReference.md** - Function and error reference
- **README.md** (each module) - Module-specific details

---

**Last Updated:** January 2025

**Maintained By:** FeCIM Development Team

**Questions?** See CLAUDE.md or relevant module README.md for context-specific guidance.
