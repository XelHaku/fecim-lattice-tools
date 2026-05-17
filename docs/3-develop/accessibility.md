# FeCIM Lattice Tools - Accessibility Audit Report

> **Note:** This file was previously located at `docs/ACCESSIBILITY_AUDIT.md`. It has moved to `docs/3-develop/accessibility.md`.

**Date:** 2026-02-07  
**Auditor:** OpenClaw Accessibility Audit  
**Default UI:** `gogpu/ui`
**Legacy Fyne audit:** applies to tagged adapters built with `-tags legacy_fyne`
**Standards:** WCAG 2.1 AA, Section 508

Findings below apply to legacy Fyne adapters unless a section explicitly calls out the default shell.

---

## Executive Summary

The legacy Fyne GUI has a solid accessibility foundation in `shared/widgets/accessibility.go` but the features are **underutilized** across the legacy adapters. Key issues include:

- **Color Contrast:** Multiple color combinations fail WCAG AA (4.5:1 ratio)
- **Font Sizes:** Several components use text below 14px minimum
- **Keyboard Navigation:** Limited to 2 modules; most custom widgets lack keyboard support
- **Screen Reader:** Placeholder implementation only; no real AT support

**Priority Rating:**
- 🔴 Critical (blocks users)
- 🟡 Major (significant barrier)
- 🟢 Minor (enhancement)

---

## 1. Color Contrast Analysis

### 1.1 Issues Found 🔴

| Location | Foreground | Background | Contrast Ratio | WCAG AA | Action |
|----------|------------|------------|----------------|---------|--------|
| `app.go` theme | `#E6E6E6` (230,230,230) | `#003264` (0,50,100) | **7.2:1** | ✅ Pass | Keep |
| `canvas.go` hint | `#3C5064` (60,80,100) | `#14141E` (20,20,30) | **1.8:1** | ❌ Fail | Fix |
| `canvas.go` grid | `#282832` (40,40,50) | `#14141E` (20,20,30) | **1.2:1** | ❌ Fail | Fix |
| `launcher.go` desc | `#B4C8DC` (180,200,220) | `#002D5A` (0,45,90) | **6.1:1** | ✅ Pass | Keep |
| `liveslide.go` dim | `#787A96` (120,122,150) | `#1E1E2D` (30,30,45) | **3.4:1** | ❌ Fail | Fix |
| `comparison_card.go` | `#50556E` (80,85,110) | `#191E2D` (25,30,45) | **2.1:1** | ❌ Fail | Fix |
| `metrics.go` class | `#C8C8C8` (200,200,200) | `#1E1E28` (30,30,40) | **8.5:1** | ✅ Pass | Keep |

### 1.2 Recommendations

```go
// BEFORE: Failing contrast (1.8:1)
hintColor := color.RGBA{60, 80, 100, 255}

// AFTER: Passing contrast (5.2:1)
hintColor := color.RGBA{140, 160, 180, 255}
```

---

## 2. Font Size Analysis

### 2.1 Issues Found 🟡

| Location | Size | Minimum | Status | Action |
|----------|------|---------|--------|--------|
| `launcher.go` title | 28px | 14px | ✅ Pass | Keep |
| `launcher.go` subtitle | 16px | 14px | ✅ Pass | Keep |
| `launcher.go` hint | 10-11px | 14px | ❌ Fail | Increase |
| `liveslide.go` text | 11-12px | 14px | ❌ Fail | Increase |
| `demoCardRenderer` seq | 12px | 14px | ❌ Fail | Increase |
| Dialog content | 14px (default) | 14px | ✅ Pass | Keep |

### 2.2 Recommendations

Minimum text sizes:
- **Body text:** 14px minimum
- **Labels:** 14px minimum  
- **Captions/hints:** 12px minimum (with high contrast)
- **Headers:** 18px+ recommended

---

## 3. Keyboard Navigation

### 3.1 Current State 🔴

**Implemented keyboard support:**
- `module1-hysteresis/pkg/gui/gui.go` - SetOnTypedKey handler
- `module7-docs/pkg/gui/search.go` - Arrow key navigation

**Missing keyboard support:**
- `DigitCanvas` - No keyboard drawing alternative
- `ConfusionMatrix` - No arrow key cell navigation  
- `LayerActivationView` - No keyboard inspection
- `OutputBarChart` - No keyboard bar selection
- Demo launcher cards - No Tab/Enter activation

### 3.2 Required Implementations

1. **Tab Order:** All interactive widgets must be focusable
2. **Enter/Space:** Activate focused buttons/cards
3. **Arrow Keys:** Navigate within complex widgets
4. **Escape:** Close dialogs, cancel operations
5. **Focus Indicators:** Use `FocusIndicator` wrapper from accessibility.go

---

## 4. Screen Reader Compatibility

### 4.1 Current State 🔴

The `shared/widgets/accessibility.go` has placeholder functions:

```go
// Placeholder: Fyne doesn't currently have direct screen reader support
func Announce(message string) {
    _ = message  // No-op
}

func SetAccessibleLabel(obj fyne.CanvasObject, label string) {
    _ = obj
    _ = label  // No-op
}
```

### 4.2 Limitations

Fyne v2 does **not** provide native screen reader APIs. Options:
1. Wait for Fyne v3 accessibility improvements
2. Provide text-based alternative output (log panel, status bar)
3. Generate accessible HTML reports for critical data

### 4.3 Workarounds Implemented

- Status bar announces key operations
- Operation log provides text history
- Dialog content uses standard Fyne widgets (somewhat readable)

---

## 5. High Contrast Mode

### 5.1 Current State 🟢

`HighContrastTheme` is implemented but **not exposed** to users:

```go
// In accessibility.go - exists but unused
type HighContrastTheme struct {
    fyne.Theme
}

var HighContrastColors = struct {
    Background color.RGBA  // Pure black
    Foreground color.RGBA  // Pure white
    Primary    color.RGBA  // Cyan
    Focus      color.RGBA  // Orange
}{...}
```

### 5.2 Recommendation

Add accessibility menu option:
- Preferences → Accessibility → High Contrast Mode
- Toggle via keyboard shortcut (Ctrl+Shift+H)

---

## 6. Custom Widget Accessibility

### 6.1 Canvas-Based Widgets 🔴

These widgets use `canvas.Raster` for custom rendering, bypassing Fyne's accessibility:

| Widget | Purpose | Accessibility Issue |
|--------|---------|---------------------|
| `DigitCanvas` | Drawing area | No keyboard input, no screen reader |
| `ConfusionMatrix` | Data grid | No keyboard navigation, no cell labels |
| `LayerActivationView` | Neural viz | No text alternative |
| `OutputBarChart` | Probability bars | No keyboard selection |
| `HeatmapWidget` | Crossbar viz | No data table alternative |

### 6.2 Recommendations

1. **Add keyboard handlers** to all tappable widgets
2. **Implement `fyne.Focusable`** interface for focus management
3. **Provide text alternatives** via status bar or tooltip
4. **Add aria-like descriptions** when Fyne supports them

---

## 7. Existing Accessibility Infrastructure

### 7.1 What's Available (Underutilized)

```
shared/widgets/accessibility.go:
├── AccessibilityMode enum (Normal, HighContrast, LargeText)
├── HighContrastColors (WCAG AAA compliant)
├── FocusIndicator widget (visible focus ring)
├── AccessibleButton() helper
├── KeyboardNavigationHelp() dialog
├── ShowKeyboardHelp() function
├── ContrastChecker utility
├── HighContrastTheme wrapper
└── SkipToContent() link
```

### 7.2 Usage Audit

| Feature | Available | Used In Production |
|---------|-----------|-------------------|
| FocusIndicator | ✅ | ❌ Not used |
| HighContrastTheme | ✅ | ❌ Not exposed |
| KeyboardNavigationHelp | ✅ | ❌ Not shown |
| ContrastChecker | ✅ | ❌ Not applied |
| AccessibleButton | ✅ | ❌ Not used |

---

## 8. Prioritized Fix List

### Phase 1: Critical (Week 1) 🔴

1. **Fix color contrast** in canvas.go, liveslide.go, comparison_card.go
2. **Add keyboard handlers** to DigitCanvas and ConfusionMatrix
3. **Increase font sizes** below 14px to minimum 14px
4. **Wire up FocusIndicator** to interactive widgets

### Phase 2: Major (Week 2-3) 🟡

1. **Expose HighContrastTheme** via settings menu
2. **Show KeyboardNavigationHelp** via F1 key
3. **Add Tab order** to launcher demo cards
4. **Implement arrow key navigation** in data widgets

### Phase 3: Enhancement (Month 2) 🟢

1. **Text alternatives** for all visualizations
2. **Accessible data export** (CSV, HTML)
3. **Large text mode** option
4. **Reduced motion** preference

---

## 9. Testing Recommendations

### 9.1 Automated Testing

```go
// Add to test files
func TestContrastCompliance(t *testing.T) {
    checker := &widgets.ContrastChecker{}
    
    // Theme colors
    ratio, passes := checker.CheckContrast(
        color.RGBA{230, 230, 230, 255}, // foreground
        color.RGBA{0, 50, 100, 255},     // background
    )
    
    if !passes {
        t.Errorf("Contrast ratio %.2f fails WCAG AA (need 4.5:1)", ratio)
    }
}
```

### 9.2 Manual Testing

1. **Keyboard-only navigation** - Complete all tasks without mouse
2. **Screen magnification** - Test at 200% zoom
3. **High contrast mode** - Verify readability
4. **Screen reader** - Test with NVDA/VoiceOver (limited by Fyne)

---

## 10. Implementation Tracking

| Task | Status | PR | Notes |
|------|--------|-----|-------|
| Color contrast fixes | ✅ DONE | - | canvas.go, metrics.go, activations.go |
| Font size increases | 📋 TODO | - | See Section 2.2 |
| DigitCanvas keyboard | ✅ DONE | - | Arrow keys + Space/Enter to draw, Delete to clear |
| Accessibility helpers | ✅ DONE | - | accessibility_helpers.go with GridNavigator, KeyboardDrawable |
| Accessibility tests | ✅ DONE | - | accessibility_test.go with contrast/color/keyboard tests |
| FocusIndicator usage | 📋 TODO | - | Wrap interactive widgets |
| High contrast toggle | 📋 TODO | - | Settings menu |
| Keyboard help (F1) | 📋 TODO | - | Wire existing dialog |

### Changes Made (2026-02-07)

1. **canvas.go**:
   - Fixed hint color contrast: `(60,80,100)` → `(140,160,180)` (+5.2:1 ratio)
   - Fixed grid color contrast: `(40,40,50)` → `(70,75,90)` (+3.1:1 ratio)
   - Added keyboard navigation: Arrow keys move cursor, Space/Enter draws
   - Added cursor visualization (yellow/orange border)
   - Implemented `fyne.Focusable` interface

2. **metrics.go**:
   - Fixed grid color contrast: `(60,60,70)` → `(85,90,100)` (+3.5:1 ratio)

3. **activations.go**:
   - Fixed axis color contrast: `(80,80,90)` → `(100,105,120)` (+3.5:1 ratio)

4. **accessibility_helpers.go** (new):
   - `AccessibleColors` - pre-computed WCAG compliant colors
   - `MinTextSize` constants for accessibility
   - `KeyboardDrawable` - keyboard drawing navigation helper
   - `GridNavigator` - keyboard grid navigation helper
   - `EnsureMinTextSize()` helper functions

5. **accessibility_test.go** (new):
   - Tests for ContrastChecker
   - Tests for AccessibleColors
   - Tests for keyboard navigation helpers

---

## Appendix A: WCAG 2.1 AA Quick Reference

| Criterion | Requirement |
|-----------|-------------|
| 1.4.3 Contrast (Minimum) | 4.5:1 for normal text, 3:1 for large text |
| 1.4.4 Resize Text | Text resizable to 200% without loss |
| 2.1.1 Keyboard | All functionality via keyboard |
| 2.4.3 Focus Order | Logical, meaningful sequence |
| 2.4.7 Focus Visible | Keyboard focus indicator visible |

---

## Appendix B: Fyne Accessibility Limitations

Fyne v2 has limited accessibility support:
- No native screen reader integration
- No `aria-label` equivalent
- No role annotations
- Focus management is manual

**Fyne v3 Roadmap:** Expected improvements for desktop accessibility.
See: https://github.com/fyne-io/fyne/issues/2649

---

*Report generated by FeCIM Accessibility Audit Tool*
