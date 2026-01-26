# Module 5 UI Improvements - Learnings

## Scroll Indicator Implementation (2026-01-25)

### Problem
Users were not discovering content below the fold in Module 5. The Data Center Calculator and other important sections were hidden below the visible viewport (900px tall window).

### Solution
Added a visual scroll indicator below the hero Energy Race section:
- Text: "▼ Scroll down for Data Center Calculator, Market Analysis, and more ▼"
- Style: Center-aligned, italic, low importance (subtle gray)
- Placement: Between hero section and first row of content

### Location
File: `<local-path>`
Lines: After `heroEnergyRace`, before `row1`

### Pattern
This pattern can be reused in other scrollable content areas where:
1. Important content exists below the fold
2. Users may not realize scrolling is needed
3. A gentle hint improves discoverability without being intrusive

### Fyne Widget Usage
```go
scrollHintLabel := widget.NewLabelWithStyle(
    "▼ Scroll down for more ▼", 
    fyne.TextAlignCenter, 
    fyne.TextStyle{Italic: true}
)
scrollHintLabel.Importance = widget.LowImportance
```

The `LowImportance` setting applies theme-specific subtle coloring (usually gray).

## Module 6 EDA UI Improvements (2026-01-25)

### Enhanced Statistics Display
**Location**: `module6-eda/pkg/gui/tabs/builder_validation_tab.go`

Added density and utilization metrics to the statistics box:
- Density: cells/µm² calculation
- Utilization: percentage of cell area vs total array area
- Added visual separator for better prominence

### Example Preview Content
**Problem**: Preview tabs showed "Generate to see..." which provided no context.

**Solution**: Added example content to all three preview tabs:
1. **Verilog Preview**: Shows example module structure with comments
2. **DEF Preview**: Shows example DEF format with comments
3. **Layout Preview**: Shows example ASCII visualization with explanation

This helps users understand what they'll get before clicking Generate.

### Improved Log Styling
**Changes**:
- Added monospace font to log output for better readability
- Added "Clear Log" button positioned next to "Validation Log" title
- Log section now has proper header with button

### Compact OpenLane Status Panel
**Changes**:
- Changed title to "OpenLane (Optional)" to indicate it's not required
- Added helpful text: "Optional: Enable placement validation if OpenLane/Docker is installed"
- Auto-hides "Pull Image" button when Docker is not available
- Adjusted split ratio: 65% validation results, 35% OpenLane panel

### Layout Improvements
**Before**: 50/50 split gave too much space to OpenLane (which is optional)
**After**: 65/35 split prioritizes validation results (which are primary)

### Pattern: Conditional UI Elements
```go
// Hide button based on status
go func() {
    time.Sleep(500 * time.Millisecond)
    fyne.Do(func() {
        if strings.Contains(dockerStatus.Text, "not available") {
            pullImageBtn.Hide()
        }
    })
}()
```

This pattern shows UI elements only when relevant, reducing clutter.

## 2026-01-25: MNIST Module 3 UI Layout Improvements

Successfully implemented UI layout fixes based on user analysis to improve space utilization and reduce cramped layouts.

### Changes Applied

1. **Increased Canvas Vertical Space** (`dualmode.go` line 190)
   - Changed `leftSplit.SetOffset(0.35)` to `0.50`
   - Drawing canvas now gets 50% of vertical space instead of 35%
   - Gives users more comfortable drawing area

2. **Increased Left Column Width** (`dualmode.go` line 196)
   - Changed `mainSplit.SetOffset(0.35)` to `0.40`
   - Left column (drawing + controls) now gets 40% of horizontal space instead of 35%
   - Better balance between drawing/controls and results/weights

3. **Fixed ADC/DAC/Hidden Layout** (`dualmode.go` lines 565-577)
   - Replaced cramped 6-column grid with 2-row layout
   - Row 1: ADC and DAC (4 columns: label, select, label, select)
   - Row 2: Hidden (2 columns: label, select)
   - Much better spacing and readability

4. **Fixed Preset Buttons Layout** (`dualmode.go` lines 591-600)
   - Split 5-button row into 2 rows (3+2)
   - Row 1: Ideal, QuantCliff, Noisy (3 buttons)
   - Row 2: BrokenADC, Tour (2 buttons)
   - Prevents button cramming

5. **Moved P1 Widgets to Weight Zone Tabs** (`dualmode.go` lines 611-635, 714-728)
   - Removed Quantization and Energy widgets from controls zone
   - Added them as new tabs in weight zone
   - Now: "Quantized", "FP vs Quant", "Side-by-Side", "Quantization", "Energy"
   - Controls zone is now compact and focused
   - P1 widgets get proper full-panel space instead of being cramped

6. **Increased Canvas MinSize** (`canvas.go` line 148)
   - Changed `fyne.NewSize(280, 280)` to `fyne.NewSize(350, 350)`
   - Canvas can now expand larger, especially with 50% vertical allocation
   - Comment updated to reflect new size

### Build Verification

- All changes compile successfully
- MNIST module builds cleanly
- Full project builds without errors

### Layout Philosophy

- Controls should be compact but not cramped
- Widgets that need space should use tabs or expanding containers
- Split offsets should be tuned based on actual content needs
- Drawing canvas is primary interaction point, should be prominent

## 2026-01-25: Module 5 Comparison Simplification

Successfully simplified Module 5 by removing confusing and duplicate elements.

### Changes Applied

**File**: `<local-path>`

1. **Removed Unused Struct Fields**
   - Removed: `educationalPanel`, `operationLog`, `modeIndicator`
   - Removed: `modeSelect`, `pauseBtn`
   - Removed: `memoryWall` (animation widget)
   - Removed: `presentationMode`, `currentPhase`, `phaseTimer` (animation state)

2. **Simplified Header** (lines 410-425)
   - Removed: Mode selector dropdown and Pause button
   - Added: Simple "Reset Animation" button
   - Layout: Title + spacer + Reset button

3. **Removed Right Panel** (lines 530-546)
   - Eliminated: Educational panel (duplicated tab content)
   - Eliminated: Operation log (not useful)
   - Eliminated: Sources hyperlink (info already in footer)
   - Result: More horizontal space for center content

4. **Simplified Energy Comparison Tab** (lines 449-468)
   - Removed: Memory Wall animation card (confusing)
   - Kept: Hero headline, Energy Race, Analog States
   - Layout is cleaner and more focused

5. **Simplified Market & Strategy Tab** (lines 471-499)
   - Removed: "Technology Readiness Level 4" card (duplicated footer disclaimer)
   - Kept: Market chart, Competitive matrix, Phased strategy

6. **Cleaned Up Footer** (lines 548-566)
   - Removed: Mode indicator widget
   - Kept: Status label + disclaimer

7. **Updated Main Layout** (lines 568-585)
   - Changed: 12%/68%/20% (left/center/right) to 15%/85% (left/center)
   - Removed: Right panel entirely
   - Left panel increased from 150px to 180px minimum width
   - Center content now gets 85% of horizontal space

8. **Removed Animation Mode Code**
   - Removed: Auto-demo mode handling (lines 212-227)
   - Removed: `onPhaseChanged()` function
   - Removed: `updateStatusForMode()` function
   - Removed: `SetPresentationMode()` function
   - Removed: Educational panel updates in `updateCalculations()`
   - Removed: Operation log entries

### Build Verification

- Module 5 compiles successfully
- Full project builds without errors
- All removed code was unused or redundant

### UI Design Lessons

- Remove duplicate content - if it's in tabs, don't repeat in sidebar
- Remove unused features - presentation modes were never used
- Focus on primary content - give tabs maximum space
- Simple is better - fewer controls = less confusion
- Footer disclaimers don't need duplication in tabs


## BUG-M2-002: Missing fyne.Do wrapper in updateConductanceDisplay

**Fixed:** 2026-01-25

**Location:** `<local-path>`

**Issue:** The `updateConductanceDisplay()` method was calling `ca.conductanceHeatmap.SetData(matrix)` without wrapping it in `fyne.Do()`. This caused thread-safety issues when called from goroutines (e.g., from `resetArray()` in analysis.go:198).

**Fix:** Wrapped the `SetData()` call in `fyne.Do(func() { ... })` to ensure thread-safe UI updates:

```go
func (ca *CrossbarApp) updateConductanceDisplay() {
	matrix := ca.array.GetConductanceMatrix()
	fyne.Do(func() {
		ca.conductanceHeatmap.SetData(matrix)
	})
}
```

**Impact:** Prevents race conditions and potential crashes when updating the conductance display from background operations.

**Pattern:** All UI widget updates from goroutines MUST be wrapped in `fyne.Do()` - see `docs/development/FYNE_NOTES.md` for full guidance.

## BUG-M5-002: Status Label Cache Bypass

**Fixed:** 2026-01-25

**Location:** `<local-path>`

**Issue:** The `statusLabel` had a `lastStatusText` cache field (line 80) that was declared but never used. The `updateStatus()` method was calling `SetText()` directly without checking the cache, resulting in redundant UI updates even when the status text hadn't changed.

**Fix:** Implemented cache checking in the `updateStatus()` method:

```go
func (ca *ComparisonApp) updateStatus(status string) {
	if ca.statusLabel == nil {
		return
	}
	newText := "Status: " + status
	// Only update if text has actually changed (cache bypass prevention)
	if ca.lastStatusText == newText {
		return
	}
	ca.lastStatusText = newText
	ca.statusLabel.SetText(newText)
}
```

**Documentation:** Added comments near the struct field (lines 79-80) explaining that all status updates MUST go through `updateStatus()` to use the cache properly.

**Impact:**
- Prevents redundant `SetText()` calls when status hasn't changed
- Reduces unnecessary UI refresh cycles
- Respects Fyne's internal widget caching

**Pattern:** For frequently-updated UI labels, implement application-level caching:
1. Declare cache field in struct (`lastStatusText string`)
2. Check cache before calling `SetText()`
3. Document that all updates must go through the central method
4. Verify with grep that no direct `SetText()` calls bypass the cache

**Verification:**
- `grep -r "statusLabel\.SetText"` shows only one occurrence (in `updateStatus()`)
- Tests pass: `go test ./module5-comparison/...`
- Build succeeds: `go build ./module5-comparison/...`

## BUG-M4-004: Missing sharedArrayCanvas refresh in Start()

**Fixed:** 2026-01-25

**Location:** `<local-path>`

**Issue:** The `Start()` method was refreshing multiple canvases when the tab is selected, but was missing a call to `refreshSharedArray()` for the shared array canvas. This meant the Operations Tab array visualization wasn't updated when switching to the Module 4 tab.

**Fix:** Added `e.refreshSharedArray()` call to the list of refresh operations in the `Start()` method:

```go
func (e *EmbeddedCircuitsApp) Start() {
	// Refresh all canvases when tab is selected
	e.refreshWriteArray()
	e.refreshWritePulse()
	e.refreshReadZone()
	e.refreshTimingDiagrams()
	e.refreshSharedArray()  // <- Added this line
	fyne.Do(func() {
		if e.computeArrayCanvas != nil {
			e.computeArrayCanvas.Refresh()
		}
		// ... other canvas refreshes
	})
}
```

**Impact:** The shared array canvas in the Operations Tab now properly refreshes when the Module 4 tab becomes visible.

**Pattern:** When implementing the embedded app interface (`Start()`, `Stop()`, `BuildContent()`), ensure ALL canvases and visualizations are refreshed in `Start()` to guarantee up-to-date display when the tab is selected.

## BUG-M4-001: Operations Panel Visibility Sync on Mode Change

**Fixed:** 2026-01-25

**Location:** `<local-path>`

**Issue:** When switching operation modes (WRITE/READ/COMPUTE), panels were being toggled individually with if/else logic. This created a race condition where multiple panels could be visible during the transition, causing leftover UI elements from the previous mode to persist.

**Root Cause:** Sequential show/hide operations without clearing all states first:
```go
// BROKEN: Shows new panel before hiding old ones
if mode == ModeWrite {
    ca.writeConfigPanel.Show()
} else {
    ca.writeConfigPanel.Hide()
}
// If ModeRead, readConfigPanel shows before writeConfigPanel hides
```

**Fix:** Hide ALL panels first, THEN show only the selected panel:

```go
// updateOperationsPanels() - CORRECTED PATTERN
func (ca *CircuitsApp) updateOperationsPanels() {
    ca.mu.RLock()
    mode := ca.currentMode
    ca.mu.RUnlock()

    // CRITICAL: Hide ALL panels first to prevent leftover UI
    if ca.writeConfigPanel != nil {
        ca.writeConfigPanel.Hide()
    }
    if ca.readConfigPanel != nil {
        ca.readConfigPanel.Hide()
    }
    if ca.computeConfigPanel != nil {
        ca.computeConfigPanel.Hide()
    }

    // THEN show only the selected panel
    switch mode {
    case ModeWrite:
        if ca.writeConfigPanel != nil {
            ca.writeConfigPanel.Show()
        }
    case ModeRead:
        if ca.readConfigPanel != nil {
            ca.readConfigPanel.Show()
        }
    case ModeCompute:
        if ca.computeConfigPanel != nil {
            ca.computeConfigPanel.Show()
        }
    }

    // ... rest of function
}
```

**Applied to:**
1. `updateOperationsPanels()` (lines 580-622) - Mode-specific config panels
2. `updateOperationsButtons()` (lines 1473-1502) - Mode-specific action buttons

**Pattern:** When toggling visibility of mutually exclusive UI elements:
1. **HIDE ALL** elements first
2. **SHOW ONLY** the selected element
3. Use a `switch` statement instead of if/else chains for clarity

**Impact:** Prevents UI confusion where users see overlapping content from multiple modes. Ensures clean mode transitions in unified multi-mode interfaces.

**Related:** This pattern applies to any stacked container with toggled visibility (tabs, mode panels, wizard steps, etc.).

## BUG-M2-004: Educational Content Label Layout Trigger

**Fixed:** 2026-01-25

**Location:** `<local-path>`

**Issue:** The educational content label (`ca.eduContentLabel`) had wrapping disabled (line 228: `Wrapping = fyne.TextWrapOff`) but content changes still triggered parent layout recalculations. When switching tabs, the `setEducationalContent()` method would update the label text, causing the left panel to resize and triggering a cascade of layout updates across the entire window.

**Root Cause:** Even with wrapping disabled, Fyne's label widget recalculates `MinSize()` when text content changes. The parent VBox container respects these changes, propagating them up the layout tree.

**Fix:** Wrapped the educational content label in a fixed-size container using `container.NewGridWrap()`:

```go
// Left panel - wrap educational content in fixed-size container to prevent layout changes
// Educational content wrapper: fixed height prevents parent layout recalculation on text changes
eduContentWrapper := container.NewGridWrap(fyne.NewSize(200, 300), ca.eduContentLabel)

leftPanelContent := container.NewVBox(
	ca.eduTitleLabel,
	widget.NewSeparator(),
	eduContentWrapper,  // <- Fixed-size wrapper instead of direct label
	widget.NewSeparator(),
	ca.keyStatLabel,
	ca.keyStatValue,
)
leftPanel := container.NewVScroll(leftPanelContent)
```

**Dimensions:**
- Width: 200 pixels (matches left panel width allocation ~15% of 1280px)
- Height: 300 pixels (sufficient for longest educational text blocks)

**Impact:**
- Prevents parent layout recalculation when tab changes update educational content
- Eliminates visual "jitter" when switching between tabs
- Content changes are now contained within the fixed-size wrapper
- Scrolling still works via the outer `NewVScroll` container if content exceeds 300px

**Pattern:** For frequently-updated labels that can cause layout thrashing:
1. Set `Wrapping = fyne.TextWrapOff` (prevents wrapping changes)
2. Wrap label in `container.NewGridWrap(fyne.NewSize(width, height), label)` (prevents MinSize changes)
3. Choose dimensions based on:
   - Width: Parent container's allocation
   - Height: Longest expected content + buffer
4. Place fixed-size wrapper inside a scroll container if overflow is possible

**Alternative Solutions Considered:**
- `container.NewMax()` - Doesn't fix height, still allows dynamic resizing
- Fixed `MinSize` on label - Doesn't prevent parent propagation
- Manual size calculation - Fragile, breaks on font/theme changes

**Verification:**
- Compiles successfully: `go build ./module2-crossbar/pkg/gui`
- No diagnostics errors
- Pattern documented for reuse in other modules

## BUG-M2-005: Auto Demo Context Leak on Window Close

**Fixed:** 2026-01-25

**Location:** `<local-path>`

**Issue:** When a user closed the window with the auto demo running in Module 2 (Crossbar+), the auto demo goroutine's context was not cancelled. The `Stop()` method on embedded apps (which calls `stopAutoDemoLoop()`) was only executed after `ShowAndRun()` returned (lines 768-776), not during the window close intercept.

**Root Cause:** The window close intercept (lines 736-762) saved state and stopped recording, but did NOT call `Stop()` on demos before closing. This meant:
1. User clicks close button → close intercept runs
2. Window preferences saved
3. Recording stopped (if active)
4. Window closes immediately
5. Cleanup code after `ShowAndRun()` (lines 768-776) runs AFTER window is destroyed
6. Auto demo context leaks because cancellation never happened

**Fix:** Call `Stop()` on ALL demos inside the close intercept, BEFORE `window.Close()`:

```go
// Set close intercept to save final state
window.SetCloseIntercept(func() {
	log.Info("Application closing, saving state...")

	// Save final window size
	finalSize := window.Canvas().Size()
	saveWindowSize(prefs, finalSize)
	log.Debug("Final window size saved: %.0fx%.0f", finalSize.Width, finalSize.Height)

	// Save current tab index
	if tabs.Selected() != nil {
		for i, tab := range tabs.Items {
			if tab == tabs.Selected() {
				saveLastTab(prefs, i)
				log.Debug("Last tab saved: %d (%s)", i, tab.Text)
				break
			}
		}
	}

	// Stop recording if active
	if recordingState.IsRecording() {
		recordingState.stopRecording()
	}

	// Stop all demos to clean up resources (e.g., auto demo contexts)
	log.Debug("Stopping all demos before window close...")
	demos.demo1.Stop()
	if demos.demo2 != nil {
		demos.demo2.Stop()  // <- This calls stopAutoDemoLoop()
	}
	demos.demo3.Stop()
	demos.demo4.Stop()
	demos.demo5.Stop()
	demos.demo6.Stop()

	// Close the window
	window.Close()
})
```

**Impact:**
- Auto demo goroutines are properly cancelled before window close
- No context leaks when user closes window with demo running
- Cleanup happens in correct order: stop resources → close window → cleanup after ShowAndRun

**Pattern:** When using embedded app interface with lifecycle methods (`Start()`, `Stop()`, `BuildContent()`):
1. Call `Start()` when tab is selected (already done via `tabs.OnSelected`)
2. Call `Stop()` when tab is deselected (already done via `tabs.OnSelected`)
3. **CRITICAL:** Call `Stop()` on ALL embedded apps in window close intercept
4. The cleanup after `ShowAndRun()` is a fallback - don't rely on it for active cleanup

**Related Files:**
- `module2-crossbar/pkg/gui/embedded.go:96-99` - `Stop()` method that calls `stopAutoDemoLoop()`
- `module2-crossbar/pkg/gui/animation.go:194-207` - `stopAutoDemoLoop()` implementation

**Verification:**
- Compiles successfully: `go build ./cmd/fecim-visualizer`
- Pattern ensures all background goroutines are cancelled before window destruction
- Cleanup order: preferences → recording → demos → window close

## Error Handling - Quantization Functions (2026-01-25)

**Changed**: `QuantizeWeights` and `QuantizeBias` in `module3-mnist/pkg/core/quantize.go`

**Before**: Functions used `panic("levels must be >= 2")`

**After**: Functions return `(result, error)` instead
- `QuantizeWeights(weights, levels)` → `QuantizeWeights(weights, levels) ([][]float64, error)`
- `QuantizeBias(bias, levels)` → `QuantizeBias(bias, levels) ([]float64, error)`
- Error message: `"quantize: levels must be >= 2, got %d"`

**All callers updated**:
- `module3-mnist/pkg/core/network.go`: Uses `_` to discard error (levels are pre-clamped to [2, 31])
- `module3-mnist/pkg/core/quantize_test.go`: Proper error handling with `t.Fatalf()`
- `module3-mnist/pkg/core/physics_test.go`: Proper error handling with `t.Fatalf()`

**Pattern**: When hardcoded valid levels (like 30), safe to use `_, _ = QuantizeWeights(...)` with comment explaining why.

**Verification**:
- ✅ All tests pass
- ✅ Build succeeds
- ✅ Error paths work correctly (tested invalid levels 0, 1)

## Goroutine Panic Recovery - SafeGo Helper (2026-01-25)

**Created**: `shared/utils/recover.go`

**Purpose**: Prevent application crashes when goroutines panic. Wraps goroutines with automatic panic recovery and stack trace logging.

**Implementation**:
```go
// SafeGo runs a function in a goroutine with panic recovery.
func SafeGo(name string, fn func()) {
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("[PANIC] %s: %v\n%s", name, r, debug.Stack())
            }
        }()
        fn()
    }()
}
```

**Applied to**:
1. `module2-crossbar/pkg/gui/app_enhanced.go:461` - Enhanced MVM animation goroutine
2. `cmd/fecim-visualizer/main.go:206` - FFmpeg capture loop
3. `cmd/fecim-visualizer/main.go:536` - Screenshot notification timer
4. `cmd/fecim-visualizer/main.go:575` - Recording saved notification timer
5. `cmd/fecim-visualizer/main.go:589` - Recording error notification timer
6. `cmd/fecim-visualizer/main.go:602` - Recording timer (display + screenshot)
7. `cmd/fecim-visualizer/main.go:732` - Startup content loader
8. `cmd/fecim-visualizer/main.go:754` - Resize debugger

**Pattern**:
```go
// Before: Unprotected goroutine
go func() {
    // Code that might panic
}()

// After: Protected with SafeGo
utils.SafeGo("descriptive-name", func() {
    // Same code, now panic-safe
})
```

**Benefits**:
- Application stays running even if a background task panics
- Stack traces logged for debugging
- Named goroutines make logs easier to trace
- Consistent error handling across all background operations

**Verification**:
- ✅ All packages compile successfully
- ✅ No change in functionality (transparent wrapper)
- ✅ Ready to catch panics in production

**Future Work**: Consider applying to other goroutines in:
- `module1-hysteresis/pkg/gui/` - Simulation loops
- `module3-mnist/pkg/gui/` - Inference workers
- `module4-circuits/pkg/gui/` - Animation timers
- `module5-comparison/pkg/gui/` - Auto-update loops

## BUG-M3-001: Race Condition in DualModeNetwork.Infer() with Shared RNG

**Fixed:** 2026-01-25

**Location:** `<local-path>`

**Issue:** The `Infer()` method held an `RLock` (read lock) but called `AddGaussianNoise()` which mutates the `net.rng` field. Since multiple goroutines can hold `RLock` simultaneously, this created a data race on the shared random number generator.

**Root Cause:**
1. `Infer()` acquires `RLock` (line 567) to allow concurrent reads
2. Multiple goroutines can hold `RLock` at the same time
3. `AddGaussianNoise(data, level, net.rng)` mutates `net.rng` state (lines 592, 616, 622, 724, 729)
4. Concurrent mutation = data race

**Fix:** Added a separate mutex for RNG access:

1. **Added RNG mutex** (line 76):
   ```go
   // Separate mutex for RNG access to prevent races under RLock
   rngMu sync.Mutex
   ```

2. **Created thread-safe helper method** (lines 564-574):
   ```go
   // safeNoise applies Gaussian noise with thread-safe RNG access.
   // This prevents data races when multiple goroutines hold RLock on the network.
   func (net *DualModeNetwork) safeNoise(data []float64, noiseLevel float64) []float64 {
       if noiseLevel <= 0 {
           return data
       }
       net.rngMu.Lock()
       result := AddGaussianNoise(data, noiseLevel, net.rng)
       net.rngMu.Unlock()
       return result
   }
   ```

3. **Replaced all `AddGaussianNoise` calls**:
   - Line 592: Tour Mode CIM path
   - Line 616: Standard Mode Layer 1
   - Line 622: Standard Mode Layer 2
   - Line 724: InferCIMOnly Layer 1
   - Line 729: InferCIMOnly Layer 2

**Verified Safe:** Line 303 in `LoadWeights()` uses `net.rng.Float64()` under a write lock (`mu.Lock()`), so it's already thread-safe.

**Pattern:**
When a method holds a read lock (`RLock`) but needs to mutate shared state:
1. Identify the specific mutable field (e.g., `rng`)
2. Add a separate mutex for that field (`rngMu sync.Mutex`)
3. Create a helper method that locks the specific mutex
4. Replace all direct mutations with calls to the helper method

**Benefits:**
- Allows concurrent reads of immutable fields (weights, config) via `RLock`
- Serializes mutations to the RNG via `rngMu`
- Finer-grained locking than upgrading to full `Lock()`

**Verification:**
- ✅ `go test -race ./module3-mnist/pkg/core` passes
- ✅ `go test -race ./module3-mnist/...` passes
- ✅ `go test -race ./...` passes (all 117 tests)
- ✅ No race conditions detected

**Impact:** Multiple goroutines can now safely call `Infer()` concurrently without data races on the RNG.
