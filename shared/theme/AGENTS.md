<!-- Parent: ../../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# shared/theme

## Purpose

Provides consistent theming for all GUI modules. Defines color palettes, typography, spacing, and widget appearance. Single source of truth for visual design. Includes color reference guide, improvement notes, and test validation to ensure theme consistency across application.

## Key Files

| File | Description |
|------|-------------|
| `theme.go` | Main theme implementation (8.5KB). Color definitions, font settings, spacing constants. |
| `theme_test.go` | Theme validation and consistency tests (6.8KB). |
| `COLOR_REFERENCE.md` | Comprehensive color palette documentation with hex codes and usage guidelines. |
| `THEME_GUIDE.md` | Design guide for implementing new widgets with theme colors. |
| `IMPROVEMENTS.md` | Historical design improvements and rationale. |

## For AI Agents

### Working In This Directory

**Theme Color Palette:**

Defined in `theme.go`:
- Primary colors: Brand color, accent, success, warning, error
- Neutrals: Background, surface, text (light/dark)
- Semantic: Link, disabled, muted
- Specialized: Hysteresis (curve colors), crossbar (heatmap range), circuit (component colors)

**Typography:**

- Font families: Default, monospace
- Sizes: Display, heading, body, caption
- Weight: Regular, medium, bold (limited by Fyne)

**Spacing:**

- Unit: `unit.Dp` (density-independent pixels)
- Common spacers: Small (4dp), medium (8dp), large (16dp)

**Usage Pattern:**

```go
import sharedtheme "fecim-lattice-tools/shared/theme"

// Use theme colors directly
color := sharedtheme.ColorPrimary
background := sharedtheme.ColorBackground

// Use theme fonts
font := sharedtheme.FontBody

// Use theme spacing
padding := sharedtheme.SpaceM
```

**Color Reference:**

- Consult `COLOR_REFERENCE.md` before adding new colors
- All colors must be accessible (WCAG AA minimum for text)
- Document rationale in `IMPROVEMENTS.md`

**Design Guide:**

- Follow `THEME_GUIDE.md` for widget styling
- Consistent spacing and padding across modules
- Semantic colors for status (green=success, red=error, etc.)

### Testing Requirements

```bash
# Run all theme tests
go test ./shared/theme -v

# Run theme validation
go test ./shared/theme -run TestTheme -v

# Run consistency checks
go test ./shared/theme -run TestConsistency -v
```

### Common Patterns

- **Widget styling**: Apply theme colors and spacing via `Create*Color()` and `Space*()` functions
- **Color gradients**: Use theme color range for heatmaps (e.g., `ColorMin` to `ColorMax`)
- **Dynamic themes**: Theme can be switched at runtime; all widgets auto-update via `Refresh()`
- **Accessibility**: Always check contrast ratio for text + background

## Dependencies

### Internal

- `shared/logging` (optional) - Theme loading diagnostics

### External

- `fyne.io/fyne/v2` - Fyne framework for color and theme interfaces
- `image/color` (Go stdlib) - Color type definitions

<!-- MANUAL: Last edited 2026-02-13. Theming is stable; colors documented in COLOR_REFERENCE.md. -->
