# Module 6 EDA - UI Improvement Proposal

> Generated: 2026-01-30
> Status: Proposal for review
> Note: Targets a broader multi-panel GUI than the current 2-view implementation.

## Executive Summary

This document proposes improvements to the Module 6 EDA GUI to address identified usability issues, enhance visual design, and improve accessibility compliance.

## Current State Assessment

### Pain Points
1. **Ugly Layout** - User feedback indicates visual design needs work
2. **Low Contrast** - White on dark blue strains eyes
3. **Dense Information** - Config panels feel cramped
4. **Small Preview Areas** - Code and images need more space
5. **Hard to Read Status** - Validation indicators unclear

### Strengths to Preserve
- Dark theme aesthetic (professional EDA look)
- Comprehensive validation feedback
- Educational Learn tab content
- Real-time statistics updates

## Proposed Improvements

### 1. Layout Architecture (HIGH Priority)

**Current:** Dense 4-column grids, cramped panels
**Proposed:** Card-based design with visual hierarchy

```
┌─────────────────────────────────────────────────────┐
│ Header: View Selector + Status                       │
├─────────────────────────────────────────────────────┤
│ ┌─────────────────┐  ┌─────────────────────────────┐│
│ │  Cell Config    │  │    Array Config             ││
│ │  ┌───────────┐  │  │  ┌───────────┐ ┌─────────┐ ││
│ │  │ Geometry  │  │  │  │ Dimensions │ │ Stats   │ ││
│ │  └───────────┘  │  │  └───────────┘ └─────────┘ ││
│ │  ┌───────────┐  │  │  ┌─────────────────────────┐││
│ │  │ Timing    │  │  │  │ Architecture Selector   │││
│ │  └───────────┘  │  │  └─────────────────────────┘││
│ └─────────────────┘  └─────────────────────────────┘│
├─────────────────────────────────────────────────────┤
│        [ Generate All ] [ Validate ] [ Export ]      │
├─────────────────────────────────────────────────────┤
│ Preview Tabs (larger height - 60% of remaining)      │
│ ┌─────────────────────────────────────────────────┐ │
│ │  Verilog | DEF | Layout                         │ │
│ │  ┌─────────────────────────────────────────┐    │ │
│ │  │ Code preview with syntax highlighting   │    │ │
│ │  │ Line numbers | Copy button              │    │ │
│ │  └─────────────────────────────────────────┘    │ │
│ └─────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────┤
│ Validation: [✓ Yosys] [✓ DEF] [✓ Cross] [⊝ Place]  │
├─────────────────────────────────────────────────────┤
│ Log (collapsible, 100px default)                    │
└─────────────────────────────────────────────────────┘
```

### 2. Color Scheme Enhancement (HIGH Priority)

**Current Palette:**
- Background: #0d3a5c (too dark)
- Text: #FFFFFF (high contrast but harsh)
- Accent: Cyan only

**Proposed Palette:**
```
Background Layers:
- BG0: #002852 (deepest - page background)
- BG1: #003264 (main panels)
- BG2: #004178 (elevated cards)
- BG3: #00508C (hover/focus states)

Text:
- Primary: #F0F4F8 (softer white)
- Secondary: #94A3B8 (muted for labels)
- Code: #E2E8F0 (code preview)

Status Colors (WCAG compliant):
- Pass: #34D399 (emerald green) + ✓ icon
- Fail: #F87171 (soft red) + ✗ icon
- Pending: #FBBF24 (amber) + ○ icon
- Skip: #9CA3AF (gray) + ⊝ icon

Accent:
- Primary: #22D3EE (cyan)
- Secondary: #818CF8 (indigo)
```

### 3. Status Indicators (HIGH Priority)

**Current:** Text labels with color
**Proposed:** Badges with icons + color + text

```go
// Example badge component
func StatusBadge(status string) fyne.CanvasObject {
    var icon, color, text = getStatusStyle(status)
    return container.NewHBox(
        canvas.NewText(icon, color),  // ✓/✗/○/⊝
        widget.NewLabel(text),         // "PASS"/"FAIL"
    )
}
```

### 4. Code Preview Enhancement (MEDIUM Priority)

**Current:** Plain MultiLineEntry
**Proposed:** Styled code block with features

Features to add:
- Line numbers (left gutter)
- Copy-to-clipboard button (top-right)
- Monospace font with proper sizing
- Subtle syntax highlighting (keywords in cyan)
- Better scrolling behavior

### 5. Statistics Display (MEDIUM Priority)

**Current:** Horizontal row with pipes "Total: 64 | Area: 11.85μm²"
**Proposed:** Compact metric cards

```
┌──────────┐ ┌──────────┐ ┌──────────┐
│ 64       │ │ 11.85    │ │ 3.68     │
│ Total    │ │ Area μm² │ │ WL μm    │
└──────────┘ └──────────┘ └──────────┘
┌──────────┐ ┌──────────┐ ┌──────────┐
│ 10.88    │ │ 5.40     │ │ 100.0%   │
│ BL μm    │ │ Density  │ │ Util     │
└──────────┘ └──────────┘ └──────────┘
```

### 6. Action Buttons (MEDIUM Priority)

**Current:** 3 buttons in row
**Proposed:** Centered button group with better sizing

- Minimum 44x44px touch targets
- Clear hover states (lighten background)
- Disabled state (50% opacity)
- Loading state (spinner icon)

### 7. Log Section (LOW Priority)

**Current:** Always visible, fixed height
**Proposed:** Collapsible with expand/collapse toggle

- Default: Collapsed (shows last line only)
- Expanded: 150px with scroll
- Toggle button: "▼ Log (3 messages)"
- Clear button moves inline

### 8. Learn Tab Improvements (LOW Priority)

**Current:** Narrow sidebar, long scroll
**Proposed:**
- Wider sidebar (240px vs 180px)
- Topic cards instead of plain list
- Sticky sidebar during scroll
- Progress indicators for completed topics

## Implementation Roadmap

| Phase | Focus | Effort |
|-------|-------|--------|
| 1 | Color scheme + Status badges | 1 week |
| 2 | Layout restructure (card-based) | 1 week |
| 3 | Code preview enhancement | 3 days |
| 4 | Statistics cards + Actions | 2 days |
| 5 | Collapsible log + Learn tab | 3 days |
| 6 | Polish + accessibility audit | 3 days |

## Fyne Implementation Notes

### Elevated Cards
```go
func ElevatedCard(title string, content fyne.CanvasObject) fyne.CanvasObject {
    bg := canvas.NewRectangle(color.NRGBA{0, 65, 120, 255}) // BG2
    shadow := canvas.NewRectangle(color.NRGBA{0, 0, 0, 50})
    shadow.Move(fyne.NewPos(2, 2))

    return container.NewStack(
        shadow,
        bg,
        container.NewPadded(content),
    )
}
```

### Status Badge
```go
func StatusBadge(status ValidationStatus) fyne.CanvasObject {
    icons := map[ValidationStatus]string{
        Pass:    "✓",
        Fail:    "✗",
        Pending: "○",
        Skip:    "⊝",
    }
    colors := map[ValidationStatus]color.Color{
        Pass:    color.NRGBA{52, 211, 153, 255},
        Fail:    color.NRGBA{248, 113, 113, 255},
        Pending: color.NRGBA{251, 191, 36, 255},
        Skip:    color.NRGBA{156, 163, 175, 255},
    }

    icon := canvas.NewText(icons[status], colors[status])
    label := widget.NewLabel(status.String())

    return container.NewHBox(icon, label)
}
```

### Threading Safety
All UI updates from goroutines must use:
```go
fyne.Do(func() {
    label.SetText("Updated value")
    widget.Refresh()
})
```

## Accessibility Checklist

- [ ] Color contrast ratio ≥ 4.5:1 (AA)
- [ ] Status not conveyed by color alone
- [ ] Touch targets ≥ 44x44px
- [ ] Keyboard navigation support
- [ ] Focus indicators visible
- [ ] Screen reader labels (ARIA)

## Success Metrics

1. **User feedback** - "Ugly layout" no longer applies
2. **Readability** - Extended use without eye strain
3. **Accessibility** - WCAG 2.1 AA compliant
4. **Efficiency** - Fewer clicks for common workflows

## References

- Fyne UI Toolkit: https://fyne.io
- WCAG Guidelines: https://www.w3.org/WAI/WCAG21/quickref/
- EDA Tool UI Patterns: KiCad, OpenLane, Magic VLSI
