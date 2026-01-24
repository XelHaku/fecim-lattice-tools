# Gio Migration Plan: FeCIM Visualizer

## Executive Summary

Migrate the FeCIM Visualizer from Fyne v2 to Gio UI framework to resolve Wayland resize instability and gain better control over custom widget rendering.

**Why Gio?**
- Native Wayland support (no glfw layer)
- Immediate-mode rendering (no Layout/Refresh cycles)
- Better suited for custom visualizations
- Pure Go, no CGO required
- Active development with good community

---

## Current Architecture (Fyne)

### Module Structure
```
cmd/fecim-visualizer/     # Main entry, tab container
module1-hysteresis/       # P-E curve, cell visualization
module2-crossbar/         # Heatmaps, bar charts, MVM animation
module3-mnist/            # Digit recognition demo
module4-circuits/         # DAC/ADC/TIA diagrams
module5-comparison/       # Technology comparison charts
module6-eda/              # EDA tool integration
shared/                   # Theme, logging, widgets
```

### Key Fyne Patterns Used
1. **widget.BaseWidget + WidgetRenderer** - All custom widgets
2. **canvas.Raster** - Heatmaps, charts (pixel-level drawing)
3. **container.Border/Stack/Grid** - Layouts
4. **container.DocTabs** - Tab navigation
5. **widget.Select/Button/Slider** - Standard controls
6. **fyne.Do()** - Thread-safe UI updates
7. **Preferences API** - Saving window size, last tab

### Custom Widgets to Migrate
| Widget | File | Complexity |
|--------|------|------------|
| CrossbarHeatmap | heatmap.go | High - raster + hover + click |
| ColorLegend | widgets.go | Medium - gradient raster |
| VectorBarChart | vectors.go | Medium - bar chart raster |
| AccuracyWaterfall | widgets.go | Medium - waterfall chart |
| ModeIndicatorBox | liveslide.go | Low - colored box + text |
| PEPlot | gui.go (demo1) | High - animated curve |
| CellVisualization | gui.go (demo1) | High - 3D-ish cell diagram |

---

## Gio Concepts Overview

### Immediate Mode vs Retained Mode

**Fyne (Retained Mode):**
```go
// Create widget once, update state, call Refresh()
heatmap := NewCrossbarHeatmap(8, 8)
heatmap.SetData(data)
heatmap.Refresh() // Triggers Layout() -> Refresh() cycle
```

**Gio (Immediate Mode):**
```go
// Redraw every frame, state is external
func (h *Heatmap) Layout(gtx layout.Context) layout.Dimensions {
    for row := range h.data {
        for col := range h.data[row] {
            // Draw cell directly - no intermediate state
            drawCell(gtx, row, col, h.data[row][col])
        }
    }
    return layout.Dimensions{Size: gtx.Constraints.Max}
}
```

### Key Gio Types

| Gio | Fyne Equivalent | Purpose |
|-----|-----------------|---------|
| `layout.Context` | `fyne.Size` | Current constraints + ops list |
| `layout.Dimensions` | `fyne.Size` | Returned size after layout |
| `op.Ops` | `fyne.Canvas` | Drawing operations buffer |
| `widget.Clickable` | `widget.Button` | Click handling |
| `widget.Bool` | `widget.Check` | Checkbox state |
| `widget.Editor` | `widget.Entry` | Text input |
| `widget.List` | `widget.List` | Scrollable list |
| `paint.ColorOp` | `canvas.Rectangle` | Fill color |
| `clip.Rect` | - | Clipping region |

### Gio Layouts

```go
// Flex (like Fyne Border/HBox/VBox)
layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
    layout.Rigid(sidebar),
    layout.Flexed(1, content),
)

// Stack (like Fyne Stack)
layout.Stack{}.Layout(gtx,
    layout.Expanded(background),
    layout.Stacked(foreground),
)

// Inset (like Fyne Padded)
layout.Inset{Top: unit.Dp(10)}.Layout(gtx, content)
```

---

## Migration Strategy

### Phase 1: Parallel Prototype (1-2 weeks)

Create a minimal Gio prototype alongside existing Fyne app:

```
cmd/fecim-visualizer-gio/    # New Gio entry point
pkg/gio/                      # Gio-specific widgets
  heatmap.go                  # Port CrossbarHeatmap
  barchart.go                 # Port VectorBarChart
  theme.go                    # FeCIM colors in Gio
```

**Goal:** Verify Gio works on Wayland without resize bugs

**Deliverable:** Single-screen demo with one heatmap

### Phase 2: Core Widgets (2-3 weeks)

Port all custom visualization widgets:

1. **Heatmap** - Most complex, do first
2. **ColorLegend** - Pairs with heatmap
3. **VectorBarChart** - Common pattern
4. **PEPlot** - Animation handling
5. **AccuracyWaterfall** - Similar to bar chart

### Phase 3: Module Migration (3-4 weeks)

Migrate one module at a time:

1. **Demo 2 (Crossbar)** - Most widgets, good test
2. **Demo 1 (Hysteresis)** - Animation patterns
3. **Demo 5 (Comparison)** - Charts
4. **Demo 3 (MNIST)** - Image handling
5. **Demo 4 (Circuits)** - Diagrams
6. **Demo 6 (EDA)** - Can wait, WIP anyway

### Phase 4: Integration (1-2 weeks)

- Tab navigation
- Preferences (window size, last tab)
- Keyboard shortcuts
- Recording/screenshot

---

## Widget Migration Examples

### Heatmap (Fyne → Gio)

**Current Fyne:**
```go
type CrossbarHeatmap struct {
    widget.BaseWidget
    data     [][]float64
    raster   *canvas.Raster
}

func (h *CrossbarHeatmap) CreateRenderer() fyne.WidgetRenderer {
    h.raster = canvas.NewRaster(h.generateImage)
    return &heatmapRenderer{heatmap: h, raster: h.raster}
}

func (r *heatmapRenderer) Refresh() {
    r.raster.Refresh()
}
```

**Gio Equivalent:**
```go
type Heatmap struct {
    Data     [][]float64
    Rows     int
    Cols     int
    MinVal   float64
    MaxVal   float64
    Colormap func(float64) color.NRGBA

    // Interaction state
    hoveredCell image.Point
    clickable   widget.Clickable
}

func (h *Heatmap) Layout(gtx layout.Context) layout.Dimensions {
    // Calculate cell size from constraints
    cellW := gtx.Constraints.Max.X / h.Cols
    cellH := gtx.Constraints.Max.Y / h.Rows

    // Handle clicks
    if h.clickable.Clicked(gtx) {
        pos := h.clickable.Position()
        col := int(pos.X) / cellW
        row := int(pos.Y) / cellH
        // Handle cell click...
    }

    // Draw cells
    for row := 0; row < h.Rows; row++ {
        for col := 0; col < h.Cols; col++ {
            value := h.Data[row][col]
            normalized := (value - h.MinVal) / (h.MaxVal - h.MinVal)
            cellColor := h.Colormap(normalized)

            rect := image.Rect(
                col*cellW, row*cellH,
                (col+1)*cellW, (row+1)*cellH,
            )
            paint.FillShape(gtx.Ops, cellColor, clip.Rect(rect).Op())
        }
    }

    return layout.Dimensions{Size: gtx.Constraints.Max}
}
```

### Bar Chart (Fyne → Gio)

**Gio Implementation:**
```go
type BarChart struct {
    Values []float64
    Labels []string
    Color  color.NRGBA
    Title  string
}

func (b *BarChart) Layout(gtx layout.Context) layout.Dimensions {
    return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
        // Title
        layout.Rigid(func(gtx layout.Context) layout.Dimensions {
            return widget.Label{}.Layout(gtx, shaper, font, unit.Sp(14), b.Title)
        }),
        // Bars
        layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
            return b.drawBars(gtx)
        }),
    )
}

func (b *BarChart) drawBars(gtx layout.Context) layout.Dimensions {
    if len(b.Values) == 0 {
        return layout.Dimensions{}
    }

    maxVal := b.Values[0]
    for _, v := range b.Values {
        if v > maxVal {
            maxVal = v
        }
    }

    barWidth := gtx.Constraints.Max.X / len(b.Values)

    for i, val := range b.Values {
        height := int(float64(gtx.Constraints.Max.Y) * (val / maxVal))
        rect := image.Rect(
            i*barWidth, gtx.Constraints.Max.Y-height,
            (i+1)*barWidth-2, gtx.Constraints.Max.Y,
        )
        paint.FillShape(gtx.Ops, b.Color, clip.Rect(rect).Op())
    }

    return layout.Dimensions{Size: gtx.Constraints.Max}
}
```

### Tab Container

**Gio Tabs:**
```go
type TabContainer struct {
    tabs     []Tab
    selected int
    list     widget.List
}

type Tab struct {
    Title   string
    Content func(gtx layout.Context) layout.Dimensions
    click   widget.Clickable
}

func (t *TabContainer) Layout(gtx layout.Context) layout.Dimensions {
    return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
        // Tab bar
        layout.Rigid(t.layoutTabBar),
        // Content
        layout.Flexed(1, t.tabs[t.selected].Content),
    )
}

func (t *TabContainer) layoutTabBar(gtx layout.Context) layout.Dimensions {
    return layout.Flex{}.Layout(gtx,
        // Map tabs to rigid elements
        // Handle clicks to switch tabs
    )
}
```

---

## Gio Project Structure

```
cmd/
  fecim-visualizer/          # Keep Fyne version during migration
  fecim-visualizer-gio/      # New Gio entry point
    main.go

pkg/gio/
  app.go                     # Main application state
  theme.go                   # FeCIM colors, fonts

  widgets/
    heatmap.go               # Heatmap widget
    colorlegend.go           # Color legend
    barchart.go              # Bar chart
    waterfall.go             # Waterfall chart
    peplot.go                # P-E curve plot
    mode_indicator.go        # Mode indicator box

  layouts/
    tabs.go                  # Tab container
    sidebar.go               # Control panel layout

  demos/
    demo1_hysteresis.go      # Demo 1 screen
    demo2_crossbar.go        # Demo 2 screen
    demo3_mnist.go           # Demo 3 screen
    demo4_circuits.go        # Demo 4 screen
    demo5_comparison.go      # Demo 5 screen
    demo6_eda.go             # Demo 6 screen
```

---

## Dependencies

### Remove (Fyne)
```go
// go.mod - remove
fyne.io/fyne/v2 v2.7.2
```

### Add (Gio)
```go
// go.mod - add
gioui.org v0.5.0
gioui.org/x v0.5.0              // Extra widgets
golang.org/x/image v0.15.0       // Image processing
```

### Keep
```go
// Unchanged - these work with any UI
multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar  // Core logic
multilayer-ferroelectric-cim-visualizer/shared/logging                  // Logging
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Learning curve | High | Medium | Start with simple widgets, use examples |
| Missing Gio widgets | Medium | Medium | Build custom or use gioui.org/x |
| Animation complexity | Medium | High | Study Gio animation examples |
| Font rendering differences | Low | Low | Test early with actual text |
| Performance regression | Low | Medium | Profile during development |
| Timeline slip | Medium | High | Keep Fyne version working as fallback |

---

## Milestones

| Milestone | Target | Criteria |
|-----------|--------|----------|
| M1: Prototype | Week 2 | Single heatmap renders in Gio, no Wayland bugs |
| M2: Core Widgets | Week 5 | All visualization widgets ported |
| M3: Demo 2 Complete | Week 7 | Crossbar demo fully functional in Gio |
| M4: All Demos | Week 11 | All 6 demos working |
| M5: Polish | Week 13 | Preferences, shortcuts, recording |
| M6: Fyne Removal | Week 14 | Remove Fyne dependency |

---

## Getting Started

### 1. Install Gio dependencies
```bash
go get gioui.org@latest
go get gioui.org/x@latest
```

### 2. Create prototype entry point
```bash
mkdir -p cmd/fecim-visualizer-gio
```

### 3. Run Gio example to verify setup
```bash
go run gioui.org/example/hello@latest
```

### 4. Start with heatmap widget
```bash
# Create minimal heatmap test
touch pkg/gio/widgets/heatmap.go
```

---

## Resources

- **Gio Documentation**: https://gioui.org/doc
- **Gio Examples**: https://github.com/gioui/gio-example
- **Gio Architecture**: https://gioui.org/doc/architecture
- **Community Chat**: https://gophers.slack.com #gioui
- **Video Tutorial**: https://www.youtube.com/watch?v=PxnL3-Sex3o

---

## Decision Required

Before starting migration:

1. **Full migration or hybrid?**
   - Full: Replace Fyne entirely
   - Hybrid: Use Gio for problem widgets only (complex, may not work)

2. **Timeline acceptable?**
   - ~14 weeks for full migration
   - Can run both versions in parallel during transition

3. **Fallback plan?**
   - Keep Fyne version tagged/branched
   - Can revert if Gio doesn't meet needs
