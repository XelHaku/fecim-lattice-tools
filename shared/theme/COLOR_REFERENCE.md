# FeCIM Theme Color Reference

Quick visual reference for all theme colors.

## Primary Palette

```
ColorPrimary (Cyan Accent)
█████████ #00D4FF  RGB(0, 212, 255)
Use for: CTAs, accents, links, focus states

ColorSecondary (Coral Red)
█████████ #FF6B6B  RGB(255, 107, 107)
Use for: Error states, delete actions, negative feedback

ColorAccent (Teal)
█████████ #4ECDC4  RGB(78, 205, 196)
Use for: Alternative highlights, secondary accents
```

## Semantic Status Colors

```
ColorSuccess (Green)
█████████ #57CC99  RGB(87, 204, 153)
Use for: Success messages, checkmarks, confirmations

ColorError (Coral Red)
█████████ #FF6B6B  RGB(255, 107, 107)
Use for: Error messages, validation failures

ColorWarning (Amber)
█████████ #FFC857  RGB(255, 200, 87)
Use for: Warnings, cautions, important notices

ColorInfo (Cyan)
█████████ #00D4FF  RGB(0, 212, 255)
Use for: Informational messages, tips, help text
```

## Background Hierarchy (Darkest → Lightest)

```
ColorBackground (FeCIM Blue - Darkest)
█████████ #003264  RGB(0, 50, 100)
Use for: Main application background
Contrast with ColorText: 15:1 (AAA)

ColorSurface (Medium Blue)
█████████ #004682  RGB(0, 70, 130)
Use for: Cards, panels, elevated components
Contrast with ColorText: 13:1 (AAA)

ColorInput (Lighter Blue)
█████████ #00508C  RGB(0, 80, 140)
Use for: Input fields, text areas, editable content
Contrast with ColorText: 11:1 (AAA)

ColorSurfaceAlt (Lightest Interactive)
█████████ #005AA0  RGB(0, 90, 160)
Use for: Hover states, active elements, pressed buttons
Contrast with ColorText: 9:1 (AAA)
```

## Text Colors

```
ColorText (Off-White)
█████████ #F0F4F8  RGB(240, 244, 248)
Use for: Primary text, headings, body content
Best on: ColorBackground, ColorSurface, ColorInput

ColorTextDim (Light Gray)
█████████ #A0B4C8  RGB(160, 180, 200)
Use for: Secondary text, labels, placeholders, helper text
Contrast on ColorBackground: 7.5:1 (AAA)

ColorTextDisabled (Medium Gray)
█████████ #64788C  RGB(100, 120, 140)
Use for: Disabled text, inactive elements
Contrast on ColorBackground: 3.5:1 (AA large text)
```

## Interactive Element Colors

```
ColorButton (Default Button)
█████████ #004682  RGB(0, 70, 130)
Use for: Default button background (same as ColorSurface)

ColorButtonHover (Button Hover)
█████████ #005AA0  RGB(0, 90, 160)
Use for: Button hover state (same as ColorSurfaceAlt)

ColorInputDisabled (Disabled State)
████░░░░░ #002850  RGB(0, 40, 80, 128) - 50% transparent
Use for: Disabled inputs, inactive fields
```

## Visualization Colors

```
ColorGrid (Grid Lines)
████░░░░░ #005096  RGB(0, 80, 150, 128) - 50% transparent
Use for: Graph grid lines, subtle guides

ColorAxis (Axis Lines)
█████████ #A0B4C8  RGB(160, 180, 200)
Use for: Chart axes, scale markers, dimension lines

ColorPositive (Positive Values)
█████████ #FF7878  RGB(255, 120, 120)
Use for: Positive polarization, gains, upward trends

ColorNegative (Negative Values)
█████████ #78A0FF  RGB(120, 160, 255)
Use for: Negative polarization, losses, downward trends
```

## UI Structure Colors

```
ColorSeparator (Borders/Dividers)
█████████ #005AA0  RGB(0, 90, 160)
Use for: Borders, dividers, separators between sections

ColorShadow (Subtle Shadows)
████░░░░░ #001428  RGB(0, 20, 40, 128) - 50% transparent
Use for: Drop shadows, elevation indicators
```

## Color Combinations

### Recommended Pairings

```
✅ GOOD CONTRAST (Readable)
═══════════════════════════════
ColorText on ColorBackground     15:1 (AAA)
ColorText on ColorSurface         13:1 (AAA)
ColorText on ColorInput           11:1 (AAA)
ColorTextDim on ColorBackground    7.5:1 (AAA)
ColorPrimary on ColorBackground    4.5:1 (AA)
ColorWarning on ColorBackground    5.5:1 (AA)
```

### Avoid These Combinations

```
❌ POOR CONTRAST (Hard to Read)
═══════════════════════════════
ColorTextDisabled on ColorSurfaceAlt  2.8:1 (FAIL)
ColorPrimary on ColorSurfaceAlt       3.2:1 (FAIL)
ColorTextDim on ColorInput            5.2:1 (marginal for small text)
```

## Usage by Context

### Data Visualization
- Background: `ColorBackground`
- Grid: `ColorGrid`
- Axes: `ColorAxis`
- Data series: `ColorPrimary`, `ColorAccent`, `ColorSecondary`
- Positive/Negative: `ColorPositive`, `ColorNegative`

### Form Inputs
- Background: `ColorInput`
- Border: `ColorSeparator`
- Text: `ColorText`
- Placeholder: `ColorTextDim`
- Disabled: `ColorInputDisabled`
- Focus ring: `ColorPrimary`

### Buttons
- Default: `ColorButton` background, `ColorText` foreground
- Hover: `ColorButtonHover` background
- Primary: `ColorPrimary` background, `ColorBackground` foreground
- Disabled: `ColorInputDisabled` background, `ColorTextDisabled` foreground

### Status Messages
- Success: `ColorSuccess` text or background
- Error: `ColorError` text or background
- Warning: `ColorWarning` text or background
- Info: `ColorInfo` text or background

### Cards and Panels
- Background: `ColorSurface`
- Border: `ColorSeparator`
- Shadow: `ColorShadow`
- Text: `ColorText`
- Heading: `ColorText` or `ColorPrimary`

## Accessibility Checklist

- [ ] All body text uses `ColorText` (15:1 contrast)
- [ ] Secondary text uses `ColorTextDim` (7.5:1 contrast)
- [ ] Disabled text only for non-critical content
- [ ] `ColorPrimary` used sparingly (4.5:1 contrast - AA only)
- [ ] Status colors have sufficient contrast when used for text
- [ ] Interactive elements have clear hover/focus states
- [ ] Color is not the only indicator of meaning (use icons too)

## Quick Reference Card

```
┌───────────────────────────────────────────────────────┐
│ FECIM THEME QUICK REFERENCE                           │
├───────────────────────────────────────────────────────┤
│ BACKGROUNDS (dark → light)                            │
│  #003264  #004682  #00508C  #005AA0                   │
│  Base     Card     Input    Hover                     │
├───────────────────────────────────────────────────────┤
│ TEXT                                                  │
│  #F0F4F8  #A0B4C8  #64788C                           │
│  Primary  Dim      Disabled                           │
├───────────────────────────────────────────────────────┤
│ ACCENTS                                               │
│  #00D4FF  #FF6B6B  #4ECDC4                           │
│  Primary  Error    Alt                                │
├───────────────────────────────────────────────────────┤
│ STATUS                                                │
│  #57CC99  #FF6B6B  #FFC857  #00D4FF                  │
│  Success  Error    Warning  Info                      │
└───────────────────────────────────────────────────────┘
```
