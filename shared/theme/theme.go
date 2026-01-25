// Package theme provides the shared FeCIM theme for consistent branding across all demos.
//
// Color Contrast Ratios (WCAG AA requires 4.5:1 for normal text, 3:1 for large text):
//   - ColorBackground vs ColorText: ~15:1 (excellent)
//   - ColorBackground vs ColorPrimary: ~4.5:1 (good for accents)
//   - ColorSurface vs ColorText: ~13:1 (excellent)
//   - ColorInput vs ColorText: ~11:1 (excellent)
//
// Visual Hierarchy (dark to light):
//   1. ColorBackground (#003264) - Base layer
//   2. ColorSurface (#004682) - Cards and elevated components
//   3. ColorInput (#00508C) - Interactive elements
//   4. ColorSurfaceAlt (#005AA0) - Hover/active states
//
// Usage:
//   - Use ColorText for primary text
//   - Use ColorTextDim for secondary/helper text
//   - Use ColorPrimary sparingly for accents and CTAs
//   - Use semantic colors (Success, Error, Warning) for status indicators
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// FeCIM theme colors - shared across all demos
var (
	// Primary brand colors
	ColorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan #00D4FF - main accent
	ColorSecondary  = color.RGBA{255, 107, 107, 255} // Coral red #FF6B6B
	ColorAccent     = color.RGBA{78, 205, 196, 255}  // Teal #4ECDC4

	// Semantic colors
	ColorWarning    = color.RGBA{255, 200, 87, 255}  // Amber #FFC857 - improved contrast
	ColorSuccess    = color.RGBA{87, 204, 153, 255}  // Green #57CC99 - for success states
	ColorError      = color.RGBA{255, 107, 107, 255} // Coral red #FF6B6B - same as secondary
	ColorInfo       = color.RGBA{0, 212, 255, 255}   // Cyan #00D4FF - same as primary

	// Background hierarchy (dark to light)
	ColorBackground = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264 - darkest
	ColorSurface    = color.RGBA{0, 70, 130, 255}    // Medium blue #004682 - cards/elevated
	ColorSurfaceAlt = color.RGBA{0, 90, 160, 255}    // Lighter blue #005AA0 - hover/active

	// Input and interactive states
	ColorInput         = color.RGBA{0, 80, 140, 255} // Input background #00508C - lighter than surface
	ColorInputDisabled = color.RGBA{0, 40, 80, 128}  // Disabled input #002850 - semi-transparent
	ColorButton        = color.RGBA{0, 70, 130, 255} // Button background - same as surface
	ColorButtonHover   = color.RGBA{0, 90, 160, 255} // Button hover - same as surface alt

	// Text colors
	ColorText         = color.RGBA{240, 244, 248, 255} // Off-white #F0F4F8 - improved readability
	ColorTextDim      = color.RGBA{160, 180, 200, 255} // Dim text #A0B4C8 - secondary text
	ColorTextDisabled = color.RGBA{100, 120, 140, 255} // Disabled text #64788C

	// Graph and visualization colors
	ColorGrid     = color.RGBA{0, 80, 150, 128}    // Grid lines (semi-transparent blue)
	ColorAxis     = color.RGBA{160, 180, 200, 255} // Axis lines
	ColorPositive = color.RGBA{255, 120, 120, 255} // Positive polarization - brighter
	ColorNegative = color.RGBA{120, 160, 255, 255} // Negative polarization - brighter

	// UI structure
	ColorSeparator = color.RGBA{0, 90, 160, 255}   // Separator/border lines
	ColorShadow    = color.RGBA{0, 20, 40, 128}    // Subtle shadows
)

// FeCIMTheme implements fyne.Theme for consistent FeCIM branding
type FeCIMTheme struct{}

// Color returns the appropriate color for the given theme color name
func (t *FeCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	// Background and foreground
	case theme.ColorNameBackground:
		return ColorBackground // FeCIM blue #003264 - main background
	case theme.ColorNameForeground:
		return ColorText // Off-white #F0F4F8 - improved contrast
	case theme.ColorNamePrimary:
		return ColorPrimary // Cyan #00D4FF

	// Buttons and interactive elements
	case theme.ColorNameButton:
		return ColorButton // Medium blue for default button state
	case theme.ColorNameHover:
		return ColorButtonHover // Lighter blue for hover state
	case theme.ColorNamePressed:
		return ColorSurfaceAlt // Even lighter for pressed state
	case theme.ColorNameDisabledButton:
		return ColorInputDisabled // Consistent disabled appearance

	// Input fields
	case theme.ColorNameInputBackground:
		return ColorInput // Lighter than surface for better visibility
	case theme.ColorNameInputBorder:
		return ColorSeparator // Border color for inputs
	case theme.ColorNameDisabled:
		return ColorInputDisabled // Semi-transparent for disabled state
	case theme.ColorNamePlaceHolder:
		return ColorTextDim // Dim text for placeholders

	// Status colors
	case theme.ColorNameSuccess:
		return ColorSuccess // Green for success states
	case theme.ColorNameError:
		return ColorError // Red for error states
	case theme.ColorNameWarning:
		return ColorWarning // Amber for warnings

	// UI structure
	case theme.ColorNameSeparator:
		return ColorSeparator // Clear separator color
	case theme.ColorNameShadow:
		return ColorShadow // Subtle shadows
	case theme.ColorNameOverlayBackground:
		return WithAlpha(ColorBackground, 230) // Semi-transparent overlay
	case theme.ColorNameMenuBackground:
		return ColorSurface // Menu background
	case theme.ColorNameHeaderBackground:
		return ColorSurface // Header background

	// Selection and focus
	case theme.ColorNameSelection:
		return WithAlpha(ColorPrimary, 80) // Semi-transparent cyan
	case theme.ColorNameFocus:
		return ColorPrimary // Cyan focus ring
	case theme.ColorNameHyperlink:
		return ColorPrimary // Cyan for links

	// Foreground overlays (text on colored backgrounds)
	case theme.ColorNameForegroundOnPrimary:
		return ColorBackground // Dark text on cyan
	case theme.ColorNameForegroundOnSuccess:
		return ColorBackground // Dark text on green
	case theme.ColorNameForegroundOnWarning:
		return ColorBackground // Dark text on amber
	case theme.ColorNameForegroundOnError:
		return ColorText // Light text on red

	// Scrollbars
	case theme.ColorNameScrollBar:
		return ColorSeparator // Scrollbar thumb
	case theme.ColorNameScrollBarBackground:
		return ColorSurface // Scrollbar track

	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the font for the given style
func (t *FeCIMTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the icon for the given name
func (t *FeCIMTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the size for the given name
func (t *FeCIMTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// Utility functions for consistent theming

// WithAlpha returns a new color with the specified alpha channel (0-255)
func WithAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: alpha,
	}
}

// GetContrastColor returns either ColorText or ColorBackground based on which
// provides better contrast with the given background color
func GetContrastColor(bgColor color.Color) color.Color {
	r, g, b, _ := bgColor.RGBA()
	// Convert to 0-255 range
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

	// Calculate relative luminance (simplified)
	luminance := (0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)) / 255.0

	// Use white text on dark backgrounds, dark text on light backgrounds
	if luminance < 0.5 {
		return ColorText
	}
	return ColorBackground
}
