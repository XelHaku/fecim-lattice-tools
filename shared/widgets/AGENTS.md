<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# shared/widgets

## Purpose

Custom Fyne widget library used across all GUI modules. Provides reusable components: material picker, preset selector, color legend, tooltip system, accessibility helpers, theme integration, responsive layout, educational animations, debug tools, and utility widgets. 100+ files (712KB) of widget implementations and tests.

## Key Files

| File | Description |
|------|-------------|
| `material_picker.go` | Material selection widget with search and filtering (12KB). |
| `preset_browser.go` | Browse and load material presets (15KB). |
| `preset_selector.go` | Quick selector for common presets (8.5KB). |
| `color_legend.go` | Color scale legend for heatmaps (10KB). |
| `tooltip_helpers.go` | Tooltip utilities for context-sensitive help (16KB). |
| `tooltips.go` | Pre-built tooltip strings for all UI elements (49KB). |
| `accessibility.go` | Screen reader support and accessible labels (11KB). |
| `accessibility_helpers.go` | Accessibility utility functions (8.4KB). |
| `educational_animations.go` | Teaching animations: hysteresis loops, waveforms (20KB). |
| `interactive_tutorials.go` | Interactive guided tutorials for new users (28KB). |
| `mode_indicator.go` | Display current operation mode (read/write/compute) (5.3KB). |
| `key_stat.go` | Key statistics display widget (4.1KB). |
| `resize_detector.go` | Detect window resize for responsive layout (7.1KB). |
| `responsive_grid_layout.go` | Auto-adjusting grid layout (4.9KB). |
| `adaptive_layout.go` | Responsive container that reflows on resize (14KB). |
| `debug.go` | Debug display tools and diagnostics (10KB). |
| `ui_helpers.go` | Common UI construction utilities (3.8KB). |
| `safe_do.go` | Wrapper for fyne.Do() with safety checks (260 bytes). |
| Base renderer components | `base_renderer.go`, `embedded_base.go` for custom rendering |

## For AI Agents

### Working In This Directory

**Widget Categories:**

1. **Selection Widgets**: Material picker, preset browser, selector
2. **Display Widgets**: Color legend, key statistics, mode indicator
3. **Layout Widgets**: Responsive grid, adaptive layout, resize detector
4. **Help Widgets**: Tooltips, accessibility, educational animations, tutorials
5. **Debug Widgets**: Debug display, diagnostics
6. **Utility Widgets**: Safe do wrapper, UI helpers

**Accessibility First:**

- All widgets support screen reader labels
- Use `accessibility_helpers.go` for ARIA annotations
- Keyboard navigation must work for every widget
- Color must not be sole indicator of state (add icons/text)

**Responsive Design:**

- Use `responsive_grid_layout.go` for auto-reflow
- Use `resize_detector.go` to track window size
- Test at multiple screen sizes (mobile, tablet, desktop)
- Never assume fixed layout

**Theme Integration:**

- All widgets respect `shared/theme` color scheme
- Use theme constants; never hardcode colors
- Implement `CanvasObject.Refresh()` for theme changes

**Fyne Integration:**

- All widgets extend Fyne `widget.BaseWidget`
- Implement `CreateRenderer()` for custom drawing
- Use `fyne.Do()` for goroutine-safe updates (see `safe_do.go`)
- Never block UI thread; long operations go to goroutine

### Testing Requirements

```bash
# Run all widget tests
go test ./shared/widgets -v

# Run material picker tests
go test ./shared/widgets -run TestMaterialPicker -v

# Run preset selector tests
go test ./shared/widgets -run TestPresetSelector -v

# Run color legend tests
go test ./shared/widgets -run TestColorLegend -v

# Run accessibility tests
go test ./shared/widgets -run TestAccessibility -v

# Run educational animation tests
go test ./shared/widgets -run TestEducationalAnimations -v

# Run responsive layout tests
go test ./shared/widgets -run TestResponsiveLayout -v

# Run resize detector tests
go test ./shared/widgets -run TestResizeDetector -v

# Run tooltip tests
go test ./shared/widgets -run TestTooltips -v

# Run base renderer tests
go test ./shared/widgets -run TestBaseRenderer -v

# Run theme integration tests
go test ./shared/widgets -run TestThemeIntegration -v

# Run accessibility helpers tests
go test ./shared/widgets -run TestAccessibilityHelpers -v

# Run widget logic tests
go test ./shared/widgets -run TestWidgetLogic -v
```

### Common Patterns

- **Material picker usage**: `NewMaterialPicker(onSelect func(*Material))` creates picker
- **Tooltip addition**: `AddTooltip(widget, "text", category)` using `tooltips.go` strings
- **Accessible labels**: `SetAccessibleLabel(widget, label)` via `accessibility_helpers.go`
- **Responsive layout**: `NewResponsiveGridLayout(colCount, minWidth)` auto-reflows
- **Color legend**: `NewColorLegend(min, max, colorGradient)` for heatmaps
- **Educational animation**: `NewEducationalAnimation(animationType)` for teaching mode
- **Safe goroutine update**: `SafeDo(func() { widget.Refresh() })` wraps `fyne.Do()`
- **Resize detection**: Use `ResizeDetector.OnResize` callback for layout changes

## Dependencies

### Internal

- `shared/theme` - Application color scheme and theming
- `shared/logging` - Widget operation logging (optional)

### External

- `fyne.io/fyne/v2` - Core Fyne framework
- `image/color` (Go stdlib) - Color definitions
- Standard library: `fmt`, `sync`, `time`

<!-- MANUAL: Last edited 2026-02-13. 100+ files of custom Fyne widgets. Accessibility-first design. -->
