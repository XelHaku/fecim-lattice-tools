# Fyne Window Resize Bug on Wayland/Sway

## WORKAROUND (Recommended)

Add this to your Sway config (`~/.config/sway/config`):

```bash
# Force FeCIM Visualizer to float instead of tile
for_window [app_id="com.fecim.visualizer"] floating enable
```

Find app_id with: `swaymsg -t get_tree | grep app_id`

This resolves the conflict between Fyne's MinSize and Sway's tiling.

---

## Bug Description

A Fyne v2 application experiences a persistent 1-pixel window resize oscillation during startup when running on Wayland with the Sway tiling window manager.

**Observed behavior:**
```
[RESIZE] Window: 0x0 -> 1493x1003    # Initial size from preferences
[RESIZE] Window: 1493x1003 -> 1494x1004   # Unwanted 1-pixel resize
[RESIZE] Window: 1494x1004 -> 455x830     # Tiling WM resize (expected)
```

The 1-pixel resize (line 2) should not happen.

## Environment

- **OS**: Linux (Ubuntu 22.04+)
- **Display Server**: Wayland
- **Window Manager**: Sway (i3-compatible tiling WM)
- **Go Version**: 1.21+
- **Fyne Version**: v2.7.2
- **Application**: Custom multi-tab GUI with heatmaps, bar charts, and custom widgets

## Symptoms

1. Window resizes by exactly 1 pixel in both width and height after initial display
2. Happens immediately after first layout pass completes
3. Causes tiling WM to recalculate tile positions (visual flicker)
4. Only occurs on Wayland/Sway, not on X11 or other compositors

## Investigation Findings

### What we tried that didn't fully work:

1. **LayoutCache in renderers** - Skip Layout() if size unchanged
   - Helped with rapid refresh loops but not the 1-pixel resize

2. **Startup suppression of Refresh() calls** - Skip widget Refresh() for first 1 second
   - Reduced unnecessary refreshes but Fyne still calls renderer.Refresh() internally

3. **Integer size comparison in LayoutCache** - Use `int(size.Width)` instead of float32
   - Prevented some redundant layouts but didn't fix the MinSize calculation

4. **Rounding window size to even integers** - Round saved/loaded sizes
   - Didn't prevent Fyne from requesting the +1 pixel adjustment

### Stack traces before resize

The refresh calls immediately before the 1-pixel resize come from custom widget renderers:
- `heatmapRenderer.Refresh()`
- `colorLegendRenderer.Refresh()`
- `vectorBarChartRenderer.Refresh()`

These are called by Fyne's internal layout system, not by our application code.

## Root Cause (Confirmed)

This is a **Layout Feedback Loop** between Fyne's MinSize calculation and Sway's tiling constraints:

1. **Interaction/Refresh:** App starts or user interacts
2. **Layout Calculation:** Fyne calculates MinSize from all widgets
3. **Resize Request:** Fyne requests window size = MinSize
4. **Rejection:** Sway forces window to tile dimensions
5. **Loop:** Fyne detects size < MinSize, requests resize again

The 1-pixel difference is caused by float32→int rounding during MinSize aggregation.

**GitHub Issues:**
- #4031: "Window size oscillation 1-pixel" - float/int rounding
- #5357: "Laggy resize on Wayland" - too many resize events
- #5449: "Infinite Loop in Layout" - recursive layout triggers

## Possible Causes

### 1. MinSize Floating-Point Rounding

Fyne calculates the window's content MinSize by recursively calling MinSize() on all widgets. If any widget returns a size with sub-pixel values (e.g., 100.4 × 200.6), the accumulated rounding could result in a 1-pixel difference.

**Evidence**: The resize is exactly +1 in both dimensions, suggesting consistent rounding up.

**Research keywords**: `fyne MinSize float32 rounding`, `fyne window minimum size calculation`

### 2. Wayland Geometry Negotiation

Wayland's window geometry protocol allows the compositor and client to negotiate the actual window size. The 1-pixel difference could be caused by:
- Server-side decorations adding to the size
- HiDPI scaling causing fractional pixels
- Compositor-enforced minimum sizes

**Research keywords**: `wayland xdg_toplevel geometry`, `sway window size negotiation`, `wayland fractional scaling`

### 3. Fyne-Wayland Driver Issue

Fyne's Wayland driver (using go-gl/glfw) might have a bug in how it handles initial window sizing or responds to configure events.

**Research keywords**: `fyne wayland resize`, `glfw wayland window size`, `fyne issue window resize`

### 4. Layout Pass Triggering Resize

During the first full layout pass, Fyne computes the actual sizes of all widgets. If the total content size is slightly larger than the requested window size, Fyne requests a resize to accommodate.

**Evidence**: The resize happens after all widgets have been laid out, suggesting the content dictates the size.

**Research keywords**: `fyne layout content size`, `fyne window auto resize`, `fyne SetFixedSize`

### 5. glfw/Wayland Scaling Issues

If using a HiDPI display or non-integer scaling, the conversion between device pixels and logical pixels could cause 1-pixel discrepancies.

**Research keywords**: `glfw wayland hidpi`, `fyne wayland scale factor`, `wayland output scale`

## Potential Solutions to Research

1. **Use SetFixedSize()** - Prevent window from resizing based on content
   ```go
   window.SetFixedSize(true)
   ```
   Trade-off: User can't resize the window

2. **Override content MinSize** - Wrap content in a container that returns a fixed MinSize
   ```go
   container.NewMax(content) // Might help?
   ```

3. **Delay content loading** - Don't add widgets until after window is shown
   ```go
   window.Show()
   time.Sleep(100 * time.Millisecond)
   window.SetContent(content)
   ```

4. **Use Canvas size callback** - Detect the stabilized size and prevent further resizes
   ```go
   window.Canvas().SetOnTypedKey(...) // No resize callback exists
   ```

5. **File a Fyne bug report** - This might be a known issue
   - https://github.com/fyne-io/fyne/issues
   - Search for: "wayland resize" "1 pixel" "sway"

## Files Modified During Investigation

- `shared/widgets/debug.go` - Added resize debugging infrastructure
- `shared/widgets/base_renderer.go` - Added LayoutCache utility
- `module2-crossbar/pkg/gui/heatmap.go` - Added size validation
- `module2-crossbar/pkg/gui/widgets.go` - Added startup suppression
- `module2-crossbar/pkg/gui/vectors.go` - Added startup suppression
- `module2-crossbar/pkg/gui/liveslide.go` - Added startup suppression
- `cmd/fecim-visualizer/main.go` - Window size rounding

## Debug Commands

```bash
# Enable resize debugging
FYNE_DEBUG_RESIZE=1 ./fecim-visualizer 2>&1 | grep RESIZE

# Enable full layout debugging
FYNE_DEBUG_LAYOUT=1 FYNE_DEBUG_RESIZE=1 ./fecim-visualizer 2>&1

# Check Fyne version
go list -m fyne.io/fyne/v2

# Check glfw version
go list -m github.com/go-gl/glfw/v3.3/glfw
```

## Related Links

- Fyne GitHub Issues: https://github.com/fyne-io/fyne/issues
- Fyne Wayland Driver: https://github.com/nicois/wayland (may be outdated)
- glfw Wayland Issues: https://github.com/glfw/glfw/issues?q=wayland+resize
- Sway Issues: https://github.com/swaywm/sway/issues
