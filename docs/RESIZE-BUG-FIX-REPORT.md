# FeCIM Visualizer Resize Bug Fix Report

## Date: 2026-01-24

## Summary

Fixed geometric instability issues in the Fyne-based GUI application running under Wayland/Sway tiling window managers. The root causes were identified using a custom debug infrastructure and addressed through shared layout utilities.

---

## Root Causes Identified

### 1. Negative/Zero Widget Sizes
**Symptom**: Widgets were receiving negative dimensions during layout (e.g., `-84.0x0.0`, `0.0x-258.8`)
**Cause**: Parent containers allocating less space than widget MinSize requirements
**Impact**: Triggers undefined layout behavior and resize cascades

### 2. Rapid Refresh Loops
**Symptom**: `modeIndicatorBoxRenderer Layout() called 52 times in rapid succession`
**Cause**: Refresh() calling Layout() which triggers another Refresh()
**Impact**: CPU spike and visual glitches

### 3. Redundant Layout Recalculations
**Symptom**: Same widget receiving Layout() calls with identical sizes
**Cause**: No size caching in custom renderers
**Impact**: Unnecessary work triggering Fyne's layout system

### 4. 1-Pixel Window Resize (FIXED)
**Symptom**: Window resizing from `2021x1003` to `2022x1004` after initialization
**Cause**: Widget Refresh() calls during startup triggering MinSize recalculations
**Impact**: Triggers tiling window manager to re-tile
**Fix**: Startup stabilization mechanism that suppresses Refresh() calls during first second

---

## Fixes Implemented

### 1. Shared Layout Cache Utility (`shared/widgets/base_renderer.go`)

```go
// LayoutCache tracks the last layout size to avoid redundant layout operations.
type LayoutCache struct {
    LastSize  fyne.Size
    HasLayout bool
}

// ShouldLayout returns true if layout is needed (size changed or first layout).
// Uses integer comparison to avoid floating-point precision issues on Wayland.
func (c *LayoutCache) ShouldLayout(size fyne.Size) bool {
    // Guard against invalid sizes (negative or zero) - critical for Wayland stability
    if size.Width <= 0 || size.Height <= 0 {
        return false
    }
    // Skip if size hasn't changed (use integer comparison to avoid float drift)
    if c.HasLayout && int(size.Width) == int(c.LastSize.Width) && int(size.Height) == int(c.LastSize.Height) {
        return false
    }
    return true
}
```

### 2. Startup Stabilization (`shared/widgets/debug.go`)

```go
// IsStartupStabilizing returns true if we're in the startup stabilization period.
// During this period, widget Refresh() calls are suppressed to prevent resize oscillation.
func IsStartupStabilizing() bool {
    // Returns true for first 1 second after startup
    return time.Since(startupTime) < 1*time.Second
}
```

This is used in widget data update methods:

```go
func (v *VectorBarChart) SetValues(values []float64) {
    // ... update data ...

    // Skip refresh during startup stabilization
    if sharedwidgets.IsStartupStabilizing() {
        return
    }
    fyne.Do(func() {
        v.Refresh()
    })
}
```

### 3. Updated Renderers

All custom renderers now use the shared `LayoutCache`:

| File | Renderer | Change |
|------|----------|--------|
| `module2-crossbar/pkg/gui/heatmap.go` | `heatmapRenderer` | Added LayoutCache, size validation |
| `module2-crossbar/pkg/gui/widgets.go` | `colorLegendRenderer` | Added LayoutCache, value caching |
| `module2-crossbar/pkg/gui/widgets.go` | `waterfallRenderer` | Added LayoutCache |
| `module2-crossbar/pkg/gui/vectors.go` | `vectorBarChartRenderer` | Added LayoutCache, text caching |
| `module2-crossbar/pkg/gui/liveslide.go` | `modeIndicatorBoxRenderer` | Added LayoutCache |
| `module5-comparison/pkg/gui/liveslide.go` | `comparisonModeRenderer` | Added LayoutCache |
| `cmd/fecim-visualizer/launcher.go` | `demoCardRenderer` | Added LayoutCache |

Widgets updated with startup suppression:

| File | Method | Change |
|------|--------|--------|
| `module2-crossbar/pkg/gui/heatmap.go` | `rateLimitedRefresh()` | Startup suppression |
| `module2-crossbar/pkg/gui/vectors.go` | `SetValues()`, `SetLabels()`, `SetUnit()` | Startup suppression |
| `module2-crossbar/pkg/gui/widgets.go` | `SetColormap()`, `SetLabels()`, `UpdateMetrics()`, `UpdateValues()`, `SetSteps()`, `SetTarget()` | Startup suppression |
| `module2-crossbar/pkg/gui/liveslide.go` | `SetMode()` | Startup suppression |

### 4. Debug Infrastructure (`shared/widgets/debug.go`)

Added comprehensive debugging controlled by environment variables:

- `FYNE_DEBUG_LAYOUT=1` - Log all Layout()/Refresh() calls
- `FYNE_DEBUG_RESIZE=1` - Track window resize events and correlate with refresh calls

Features:
- Stack trace capture for rapid refresh detection
- Window resize tracking with before/after comparison
- Interaction logging (tab changes, dropdown selections, button clicks)

---

## Files Modified

### Core Fixes
- `shared/widgets/base_renderer.go` - Added `LayoutCache`, `ValidateSize()`, `SafeResize()`
- `shared/widgets/debug.go` - Added resize debugging infrastructure

### Renderer Updates
- `module2-crossbar/pkg/gui/heatmap.go`
- `module2-crossbar/pkg/gui/widgets.go`
- `module2-crossbar/pkg/gui/vectors.go`
- `module2-crossbar/pkg/gui/liveslide.go`
- `module5-comparison/pkg/gui/liveslide.go`
- `cmd/fecim-visualizer/launcher.go`

### Debug Hooks Added
- `cmd/fecim-visualizer/main.go` - Window resize tracking
- `module2-crossbar/pkg/gui/controls.go` - Select widget interaction logging
- `module6-eda/pkg/gui/tabs/hdl_tab.go` - Button interaction logging
- `module6-eda/pkg/gui/tabs/compiler_tab.go` - Button interaction logging
- `module6-eda/pkg/gui/tabs/export_tab.go` - Button interaction logging
- `module4-circuits/pkg/gui/app.go` - View selector logging

---

## Testing Results

### Before Fix
```
[LAYOUT] WARNING: modeIndicatorBoxRenderer Layout() called 52 times in rapid succession (9.94ms)
[RESIZE-BUG] RAPID REFRESH: heatmapRenderer called 10 times in 100ms!
[RESIZE-BUG] RAPID REFRESH: colorLegendRenderer called 8 times in 100ms!
[LAYOUT] heatmapRenderer Layout(-84.0x0.0) - NEGATIVE SIZE
```

### After Fix
```
[RESIZE] Window: 0x0 -> 2022x1004        # Clean startup
[RESIZE] Window: 2022x1004 -> 683x830    # Normal tiling
No WARNING or RAPID REFRESH messages
No 1-pixel oscillation!
```

---

## All Issues Resolved

### 1-Pixel Startup Resize - FIXED
The 1-pixel resize that occurred during initialization has been eliminated by:
1. Adding startup stabilization mechanism in `shared/widgets/debug.go`
2. Suppressing widget Refresh() calls during the first second after startup
3. Using integer-based size comparison in LayoutCache to avoid floating-point drift

---

## Recommendations for Future Development

### 1. Use the Shared LayoutCache
When creating new custom renderers:

```go
type myRenderer struct {
    widget  *MyWidget
    objects []fyne.CanvasObject
    cache   sharedwidgets.LayoutCache  // ADD THIS
}

func (r *myRenderer) Layout(size fyne.Size) {
    if !r.cache.ShouldLayout(size) {  // CHECK FIRST
        return
    }
    // ... do layout ...
    r.cache.MarkLayout(size)  // MARK AFTER
}
```

### 2. Cache Text Values
When updating labels in Refresh():

```go
// BAD - triggers layout recalculation every time
r.label.SetText(fmt.Sprintf("Value: %.2f", value))

// GOOD - only update when changed
newText := fmt.Sprintf("Value: %.2f", value)
if newText != r.lastText {
    r.label.SetText(newText)
    r.lastText = newText
}
```

### 3. Use Debug Mode for Troubleshooting
```bash
FYNE_DEBUG_RESIZE=1 ./fecim-visualizer 2>&1 | grep -E "(RESIZE|WARNING|RAPID)"
```

### 4. Wrap Dynamic Content in Scroll Containers
```go
// Prevents content from forcing window resize
scrollable := container.NewScroll(dynamicContent)
scrollable.SetMinSize(fyne.NewSize(400, 300))
```

---

## Conclusion

All resize instability issues have been resolved by:
1. Adding size validation to prevent negative dimension handling
2. Implementing layout caching (with integer comparison) to avoid redundant recalculations
3. Breaking the Refresh() -> Layout() -> Refresh() cycle
4. **Adding startup stabilization to suppress Refresh() calls during initialization**

The application now runs stably under Sway/Wayland tiling window managers:
- No geometry feedback loops
- No 1-pixel oscillation during startup
- Clean window initialization and tiling behavior
