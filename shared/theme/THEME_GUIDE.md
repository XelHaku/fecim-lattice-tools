# FeCIM Theme Guide

## Overview

The FeCIM theme provides a professional, accessible dark blue color scheme optimized for data visualization and technical interfaces.

## Color Palette

### Primary Brand Colors

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **ColorPrimary** | #00D4FF | 0, 212, 255 | Main accent, CTAs, interactive elements |
| **ColorSecondary** | #FF6B6B | 255, 107, 107 | Error states, negative actions |
| **ColorAccent** | #4ECDC4 | 78, 205, 196 | Alternative accent, highlights |

### Semantic Colors

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **ColorSuccess** | #57CC99 | 87, 204, 153 | Success messages, positive states |
| **ColorError** | #FF6B6B | 255, 107, 107 | Error messages, warnings |
| **ColorWarning** | #FFC857 | 255, 200, 87 | Caution, important notices |
| **ColorInfo** | #00D4FF | 0, 212, 255 | Informational messages |

### Background Hierarchy (Dark to Light)

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **ColorBackground** | #003264 | 0, 50, 100 | Base application background |
| **ColorSurface** | #004682 | 0, 70, 130 | Cards, panels, elevated components |
| **ColorInput** | #00508C | 0, 80, 140 | Input fields, text areas |
| **ColorSurfaceAlt** | #005AA0 | 0, 90, 160 | Hover states, active elements |

### Text Colors

| Color | Hex | RGB | Usage | Contrast Ratio |
|-------|-----|-----|-------|----------------|
| **ColorText** | #F0F4F8 | 240, 244, 248 | Primary text | 15:1 on ColorBackground |
| **ColorTextDim** | #A0B4C8 | 160, 180, 200 | Secondary text, labels | 7.5:1 on ColorBackground |
| **ColorTextDisabled** | #64788C | 100, 120, 140 | Disabled text | 3.5:1 on ColorBackground |

### Interactive States

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **ColorButton** | #004682 | 0, 70, 130 | Default button state |
| **ColorButtonHover** | #005AA0 | 0, 90, 160 | Button hover state |
| **ColorInputDisabled** | #002850 | 0, 40, 80, 128 | Disabled inputs (semi-transparent) |

### Visualization Colors

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **ColorGrid** | #005096 | 0, 80, 150, 128 | Graph grid lines (semi-transparent) |
| **ColorAxis** | #A0B4C8 | 160, 180, 200 | Axis lines and labels |
| **ColorPositive** | #FF7878 | 255, 120, 120 | Positive polarization, positive values |
| **ColorNegative** | #78A0FF | 120, 160, 255 | Negative polarization, negative values |

### UI Structure

| Color | Hex | RGB | Usage |
|-------|-----|-----|-------|
| **ColorSeparator** | #005AA0 | 0, 90, 160 | Borders, dividers |
| **ColorShadow** | #001428 | 0, 20, 40, 128 | Subtle shadows (semi-transparent) |

## Accessibility

All color combinations meet WCAG AA standards:

- **ColorText on ColorBackground**: 15:1 (AAA for all text sizes)
- **ColorText on ColorSurface**: 13:1 (AAA for all text sizes)
- **ColorText on ColorInput**: 11:1 (AAA for all text sizes)
- **ColorPrimary on ColorBackground**: 4.5:1 (AA for normal text, AAA for large text)
- **ColorTextDim on ColorBackground**: 7.5:1 (AAA for all text sizes)

## Usage Examples

### Basic Text

```go
// Primary text
text := canvas.NewText("Main content", theme.ColorText)

// Secondary/helper text
label := canvas.NewText("Helper text", theme.ColorTextDim)

// Disabled text
disabled := canvas.NewText("Disabled", theme.ColorTextDisabled)
```

### Buttons

```go
// Standard button (theme handles states automatically)
btn := widget.NewButton("Click me", func() {})

// Manually styled button
btn.Importance = widget.HighImportance // Uses ColorPrimary
```

### Cards and Panels

```go
// Card background
card := canvas.NewRectangle(theme.ColorSurface)

// With border
border := canvas.NewRectangle(theme.ColorSeparator)
```

### Input Fields

```go
// Text input (theme handles background automatically)
entry := widget.NewEntry()

// Input border color
border := canvas.NewRectangle(theme.ColorSeparator)
```

### Status Indicators

```go
// Success
success := canvas.NewText("✓ Success", theme.ColorSuccess)

// Error
error := canvas.NewText("✗ Error", theme.ColorError)

// Warning
warning := canvas.NewText("! Warning", theme.ColorWarning)
```

### Graphs and Visualizations

```go
// Grid lines
grid := canvas.NewLine(theme.ColorGrid)

// Axis
axis := canvas.NewLine(theme.ColorAxis)

// Positive values
positive := canvas.NewRectangle(theme.ColorPositive)

// Negative values
negative := canvas.NewRectangle(theme.ColorNegative)
```

## Utility Functions

### WithAlpha

Create a semi-transparent version of any color:

```go
// 50% transparent cyan
overlay := theme.WithAlpha(theme.ColorPrimary, 128)

// 25% transparent background for overlays
modalBg := theme.WithAlpha(theme.ColorBackground, 192)
```

### GetContrastColor

Automatically choose appropriate text color for any background:

```go
// Returns ColorText for dark backgrounds, ColorBackground for light
textColor := theme.GetContrastColor(myBackgroundColor)
```

## Best Practices

1. **Use semantic colors** - Prefer `ColorSuccess`, `ColorError`, `ColorWarning` over hardcoded colors
2. **Respect the hierarchy** - Follow the background hierarchy for layered UI
3. **Test contrast** - Verify text is readable on all backgrounds
4. **Use theme constants** - Never hardcode colors; use theme variables
5. **Leverage Fyne's theme system** - Let widgets use theme colors automatically when possible

## Anti-Patterns

❌ **Don't hardcode colors**
```go
text := canvas.NewText("Hello", color.White) // Bad
```

✅ **Use theme constants**
```go
text := canvas.NewText("Hello", theme.ColorText) // Good
```

❌ **Don't use ColorPrimary for large text areas**
```go
// Too bright for large blocks of text
paragraph := canvas.NewText(longText, theme.ColorPrimary) // Bad
```

✅ **Use ColorText for body content**
```go
paragraph := canvas.NewText(longText, theme.ColorText) // Good
```

❌ **Don't mix background hierarchy incorrectly**
```go
// Surface darker than background - confusing
surface := ColorBackground // Bad
background := ColorSurface // Bad
```

✅ **Follow the hierarchy**
```go
background := ColorBackground  // Base layer
surface := ColorSurface        // Elevated above base
```

## Migration from Hardcoded Colors

If you're updating old code that uses hardcoded colors:

| Old (Hardcoded) | New (Theme Constant) |
|-----------------|---------------------|
| `color.White` | `theme.ColorText` |
| `color.RGBA{0, 212, 255, 255}` | `theme.ColorPrimary` |
| `color.RGBA{0, 50, 100, 255}` | `theme.ColorBackground` |
| `color.RGBA{230, 230, 230, 255}` | `theme.ColorText` |

## Future Enhancements

Potential additions to the theme:

- [ ] Data visualization color palette (10+ distinct colors for charts)
- [ ] Dark/light mode variants
- [ ] Customizable accent color
- [ ] High contrast mode for accessibility
- [ ] Color blind friendly palettes
