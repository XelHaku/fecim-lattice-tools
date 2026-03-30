// Package themes provides multiple theme options (light, dark, high-contrast)
// for the FeCIM Lattice Tools Fyne GUI applications.
//
// Usage:
//
//	import "fecim-lattice-tools/shared/themes"
//
//	// In your app initialization:
//	manager := themes.NewManager(fyneApp)
//	manager.LoadPreference() // Load saved preference
//
//	// Add theme switcher to settings:
//	switcher := themes.NewThemeSwitcher(manager)
package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// ThemeType represents the available theme types
type ThemeType string

const (
	// ThemeDark is the default FeCIM dark theme (blue-based)
	ThemeDark ThemeType = "dark"
	// ThemeLight is a light theme for daytime use
	ThemeLight ThemeType = "light"
	// ThemeHighContrast is a high-contrast theme for accessibility
	ThemeHighContrast ThemeType = "high-contrast"
	// ThemePresentation is optimized for conference projection (dark, 1.3x text)
	ThemePresentation ThemeType = "presentation"
)

// AllThemes returns all available theme types
func AllThemes() []ThemeType {
	return []ThemeType{ThemeDark, ThemeLight, ThemeHighContrast, ThemePresentation}
}

// ThemeDisplayName returns a human-readable name for the theme
func ThemeDisplayName(t ThemeType) string {
	switch t {
	case ThemeDark:
		return "Dark (FeCIM)"
	case ThemeLight:
		return "Light"
	case ThemeHighContrast:
		return "High Contrast"
	case ThemePresentation:
		return "Presentation"
	default:
		return string(t)
	}
}

// GetTheme returns a Fyne theme for the given theme type
func GetTheme(t ThemeType) fyne.Theme {
	switch t {
	case ThemeLight:
		return &LightTheme{}
	case ThemeHighContrast:
		return &HighContrastTheme{}
	case ThemePresentation:
		return &PresentationTheme{}
	case ThemeDark:
		fallthrough
	default:
		return &DarkTheme{}
	}
}

// ============================================================================
// Dark Theme (Original FeCIM Theme)
// ============================================================================

// DarkTheme is the original FeCIM dark theme with blue-based colors
type DarkTheme struct{}

// Dark theme color palette
var (
	// Primary brand colors (dark)
	DarkPrimary   = color.RGBA{0, 212, 255, 255}   // Cyan #00D4FF
	DarkSecondary = color.RGBA{255, 107, 107, 255} // Coral red #FF6B6B
	DarkAccent    = color.RGBA{78, 205, 196, 255}  // Teal #4ECDC4

	// Semantic colors (dark)
	DarkWarning = color.RGBA{255, 200, 87, 255}  // Amber #FFC857
	DarkSuccess = color.RGBA{87, 204, 153, 255}  // Green #57CC99
	DarkError   = color.RGBA{255, 107, 107, 255} // Coral red #FF6B6B
	DarkInfo    = color.RGBA{0, 212, 255, 255}   // Cyan #00D4FF

	// Background hierarchy (dark to light)
	DarkBackground = color.RGBA{0, 50, 100, 255} // FeCIM blue #003264
	DarkSurface    = color.RGBA{0, 70, 130, 255} // Medium blue #004682
	DarkSurfaceAlt = color.RGBA{0, 90, 160, 255} // Lighter blue #005AA0
	DarkInput      = color.RGBA{0, 80, 140, 255} // Input background #00508C
	DarkButton     = color.RGBA{0, 70, 130, 255} // Button background

	// Text colors (dark)
	DarkText         = color.RGBA{240, 244, 248, 255} // Off-white #F0F4F8
	DarkTextDim      = color.RGBA{160, 180, 200, 255} // Dim text #A0B4C8
	DarkTextDisabled = color.RGBA{100, 120, 140, 255} // Disabled text #64788C

	// UI structure (dark)
	DarkSeparator = color.RGBA{0, 90, 160, 255} // Separator/border lines
	DarkShadow    = color.RGBA{0, 20, 40, 128}  // Subtle shadows
)

// Color returns the color for the given theme color name
func (t *DarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return DarkBackground
	case theme.ColorNameForeground:
		return DarkText
	case theme.ColorNamePrimary:
		return DarkPrimary
	case theme.ColorNameButton:
		return DarkButton
	case theme.ColorNameHover:
		return DarkSurfaceAlt
	case theme.ColorNamePressed:
		return DarkSurfaceAlt
	case theme.ColorNameDisabledButton:
		return color.RGBA{0, 40, 80, 128}
	case theme.ColorNameInputBackground:
		return DarkInput
	case theme.ColorNameInputBorder:
		return DarkSeparator
	case theme.ColorNameDisabled:
		return color.RGBA{0, 40, 80, 128}
	case theme.ColorNamePlaceHolder:
		return DarkTextDim
	case theme.ColorNameSuccess:
		return DarkSuccess
	case theme.ColorNameError:
		return DarkError
	case theme.ColorNameWarning:
		return DarkWarning
	case theme.ColorNameSeparator:
		return DarkSeparator
	case theme.ColorNameShadow:
		return DarkShadow
	case theme.ColorNameOverlayBackground:
		return withAlpha(DarkBackground, 230)
	case theme.ColorNameMenuBackground:
		return DarkSurface
	case theme.ColorNameHeaderBackground:
		return DarkSurface
	case theme.ColorNameSelection:
		return withAlpha(DarkPrimary, 80)
	case theme.ColorNameFocus:
		return DarkPrimary
	case theme.ColorNameHyperlink:
		return DarkPrimary
	case theme.ColorNameForegroundOnPrimary:
		return DarkBackground
	case theme.ColorNameForegroundOnSuccess:
		return DarkBackground
	case theme.ColorNameForegroundOnWarning:
		return DarkBackground
	case theme.ColorNameForegroundOnError:
		return DarkText
	case theme.ColorNameScrollBar:
		return DarkSeparator
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *DarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *DarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *DarkTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// ============================================================================
// Light Theme
// ============================================================================

// LightTheme is a light theme for daytime use
type LightTheme struct{}

// Light theme color palette
var (
	// Primary brand colors (light)
	LightPrimary   = color.RGBA{0, 120, 180, 255}  // Deep blue #0078B4
	LightSecondary = color.RGBA{220, 80, 80, 255}  // Muted red #DC5050
	LightAccent    = color.RGBA{50, 160, 150, 255} // Teal #32A096

	// Semantic colors (light)
	LightWarning = color.RGBA{200, 140, 40, 255} // Darker amber #C88C28
	LightSuccess = color.RGBA{50, 160, 100, 255} // Darker green #32A064
	LightError   = color.RGBA{200, 60, 60, 255}  // Darker red #C83C3C
	LightInfo    = color.RGBA{0, 120, 180, 255}  // Deep blue #0078B4

	// Background hierarchy (light to dark)
	LightBackground = color.RGBA{250, 252, 255, 255} // Near-white #FAFCFF
	LightSurface    = color.RGBA{240, 244, 250, 255} // Slight blue tint #F0F4FA
	LightSurfaceAlt = color.RGBA{230, 236, 245, 255} // Hover state #E6ECF5
	LightInput      = color.RGBA{255, 255, 255, 255} // Pure white for inputs
	LightButton     = color.RGBA{0, 120, 180, 255}   // Primary color for buttons

	// Text colors (light)
	LightText         = color.RGBA{20, 35, 50, 255}    // Near-black #142332
	LightTextDim      = color.RGBA{80, 100, 120, 255}  // Dim text #506478
	LightTextDisabled = color.RGBA{140, 155, 170, 255} // Disabled text #8C9BAA

	// UI structure (light)
	LightSeparator = color.RGBA{200, 210, 220, 255} // Light separator
	LightShadow    = color.RGBA{0, 20, 40, 40}      // Very subtle shadows
)

// Color returns the color for the given theme color name
func (t *LightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return LightBackground
	case theme.ColorNameForeground:
		return LightText
	case theme.ColorNamePrimary:
		return LightPrimary
	case theme.ColorNameButton:
		return LightButton
	case theme.ColorNameHover:
		return LightSurfaceAlt
	case theme.ColorNamePressed:
		return color.RGBA{200, 210, 225, 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{200, 210, 220, 128}
	case theme.ColorNameInputBackground:
		return LightInput
	case theme.ColorNameInputBorder:
		return LightSeparator
	case theme.ColorNameDisabled:
		return color.RGBA{200, 210, 220, 128}
	case theme.ColorNamePlaceHolder:
		return LightTextDim
	case theme.ColorNameSuccess:
		return LightSuccess
	case theme.ColorNameError:
		return LightError
	case theme.ColorNameWarning:
		return LightWarning
	case theme.ColorNameSeparator:
		return LightSeparator
	case theme.ColorNameShadow:
		return LightShadow
	case theme.ColorNameOverlayBackground:
		return withAlpha(LightBackground, 240)
	case theme.ColorNameMenuBackground:
		return LightSurface
	case theme.ColorNameHeaderBackground:
		return LightSurface
	case theme.ColorNameSelection:
		return withAlpha(LightPrimary, 60)
	case theme.ColorNameFocus:
		return LightPrimary
	case theme.ColorNameHyperlink:
		return LightPrimary
	case theme.ColorNameForegroundOnPrimary:
		return color.RGBA{255, 255, 255, 255}
	case theme.ColorNameForegroundOnSuccess:
		return color.RGBA{255, 255, 255, 255}
	case theme.ColorNameForegroundOnWarning:
		return color.RGBA{30, 30, 30, 255}
	case theme.ColorNameForegroundOnError:
		return color.RGBA{255, 255, 255, 255}
	case theme.ColorNameScrollBar:
		return color.RGBA{180, 190, 200, 200}
	default:
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}
}

func (t *LightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *LightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *LightTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// ============================================================================
// High Contrast Theme
// ============================================================================

// HighContrastTheme is an accessibility-focused theme with maximum contrast
type HighContrastTheme struct{}

// High contrast theme color palette
var (
	// Primary brand colors (high contrast)
	HCPrimary   = color.RGBA{0, 255, 255, 255}   // Pure cyan
	HCSecondary = color.RGBA{255, 100, 100, 255} // Bright red
	HCAccent    = color.RGBA{0, 255, 180, 255}   // Bright teal

	// Semantic colors (high contrast)
	HCWarning = color.RGBA{255, 255, 0, 255} // Pure yellow
	HCSuccess = color.RGBA{0, 255, 0, 255}   // Pure green
	HCError   = color.RGBA{255, 0, 0, 255}   // Pure red
	HCInfo    = color.RGBA{0, 255, 255, 255} // Pure cyan

	// Background hierarchy (maximum contrast)
	HCBackground = color.RGBA{0, 0, 0, 255}     // Pure black
	HCSurface    = color.RGBA{20, 20, 30, 255}  // Near-black
	HCSurfaceAlt = color.RGBA{40, 40, 60, 255}  // Dark gray
	HCInput      = color.RGBA{10, 10, 20, 255}  // Very dark
	HCButton     = color.RGBA{0, 200, 200, 255} // Cyan button

	// Text colors (high contrast)
	HCText         = color.RGBA{255, 255, 255, 255} // Pure white
	HCTextDim      = color.RGBA{200, 200, 200, 255} // Light gray
	HCTextDisabled = color.RGBA{120, 120, 120, 255} // Medium gray

	// UI structure (high contrast)
	HCSeparator = color.RGBA{255, 255, 255, 255} // White separators
	HCShadow    = color.RGBA{0, 0, 0, 200}       // Strong shadow
)

// Color returns the color for the given theme color name
func (t *HighContrastTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return HCBackground
	case theme.ColorNameForeground:
		return HCText
	case theme.ColorNamePrimary:
		return HCPrimary
	case theme.ColorNameButton:
		return HCButton
	case theme.ColorNameHover:
		return HCSurfaceAlt
	case theme.ColorNamePressed:
		return color.RGBA{60, 60, 90, 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{60, 60, 60, 128}
	case theme.ColorNameInputBackground:
		return HCInput
	case theme.ColorNameInputBorder:
		return HCText // White border for visibility
	case theme.ColorNameDisabled:
		return color.RGBA{60, 60, 60, 128}
	case theme.ColorNamePlaceHolder:
		return HCTextDim
	case theme.ColorNameSuccess:
		return HCSuccess
	case theme.ColorNameError:
		return HCError
	case theme.ColorNameWarning:
		return HCWarning
	case theme.ColorNameSeparator:
		return HCSeparator
	case theme.ColorNameShadow:
		return HCShadow
	case theme.ColorNameOverlayBackground:
		return withAlpha(HCBackground, 245)
	case theme.ColorNameMenuBackground:
		return HCSurface
	case theme.ColorNameHeaderBackground:
		return HCSurface
	case theme.ColorNameSelection:
		return withAlpha(HCPrimary, 100)
	case theme.ColorNameFocus:
		return HCPrimary
	case theme.ColorNameHyperlink:
		return HCPrimary
	case theme.ColorNameForegroundOnPrimary:
		return HCBackground
	case theme.ColorNameForegroundOnSuccess:
		return HCBackground
	case theme.ColorNameForegroundOnWarning:
		return HCBackground
	case theme.ColorNameForegroundOnError:
		return HCText
	case theme.ColorNameScrollBar:
		return HCTextDim
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *HighContrastTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *HighContrastTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *HighContrastTheme) Size(name fyne.ThemeSizeName) float32 {
	// Slightly larger text for better readability
	switch name {
	case theme.SizeNameText:
		return theme.DefaultTheme().Size(name) * 1.1
	case theme.SizeNameCaptionText:
		return theme.DefaultTheme().Size(name) * 1.1
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// ============================================================================
// Presentation Theme (Conference / Projection)
// ============================================================================

// PresentationTheme is optimized for conference projection: dark background,
// high-contrast colors, and 1.3x enlarged text for readability at a distance.
type PresentationTheme struct{}

// Presentation theme color palette — high-contrast on dark, slightly warmer
// than the accessibility HC theme to remain easy on the eyes during talks.
var (
	PresBackground = color.RGBA{10, 10, 18, 255}     // Near-black #0A0A12
	PresSurface    = color.RGBA{24, 24, 36, 255}     // Dark surface #181824
	PresSurfaceAlt = color.RGBA{40, 40, 56, 255}     // Hover #282838
	PresInput      = color.RGBA{18, 18, 28, 255}     // Input bg #12121C
	PresPrimary    = color.RGBA{80, 200, 255, 255}   // Bright sky blue #50C8FF
	PresSecondary  = color.RGBA{255, 120, 120, 255}  // Soft red #FF7878
	PresSuccess    = color.RGBA{100, 230, 160, 255}  // Bright green #64E6A0
	PresWarning    = color.RGBA{255, 220, 80, 255}   // Bright amber #FFDC50
	PresError      = color.RGBA{255, 90, 90, 255}    // Bright red #FF5A5A
	PresText       = color.RGBA{248, 248, 255, 255}  // Almost-white #F8F8FF
	PresTextDim    = color.RGBA{180, 185, 200, 255}  // Dim text #B4B9C8
	PresSeparator  = color.RGBA{60, 60, 80, 255}     // Subtle separator #3C3C50
	PresShadow     = color.RGBA{0, 0, 0, 180}        // Strong shadow
)

func (t *PresentationTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return PresBackground
	case theme.ColorNameForeground:
		return PresText
	case theme.ColorNamePrimary:
		return PresPrimary
	case theme.ColorNameButton:
		return PresSurface
	case theme.ColorNameHover:
		return PresSurfaceAlt
	case theme.ColorNamePressed:
		return PresSurfaceAlt
	case theme.ColorNameDisabledButton:
		return color.RGBA{30, 30, 40, 128}
	case theme.ColorNameInputBackground:
		return PresInput
	case theme.ColorNameInputBorder:
		return PresSeparator
	case theme.ColorNameDisabled:
		return color.RGBA{30, 30, 40, 128}
	case theme.ColorNamePlaceHolder:
		return PresTextDim
	case theme.ColorNameSuccess:
		return PresSuccess
	case theme.ColorNameError:
		return PresError
	case theme.ColorNameWarning:
		return PresWarning
	case theme.ColorNameSeparator:
		return PresSeparator
	case theme.ColorNameShadow:
		return PresShadow
	case theme.ColorNameOverlayBackground:
		return withAlpha(PresBackground, 235)
	case theme.ColorNameMenuBackground:
		return PresSurface
	case theme.ColorNameHeaderBackground:
		return PresSurface
	case theme.ColorNameSelection:
		return withAlpha(PresPrimary, 90)
	case theme.ColorNameFocus:
		return PresPrimary
	case theme.ColorNameHyperlink:
		return PresPrimary
	case theme.ColorNameForegroundOnPrimary:
		return PresBackground
	case theme.ColorNameForegroundOnSuccess:
		return PresBackground
	case theme.ColorNameForegroundOnWarning:
		return PresBackground
	case theme.ColorNameForegroundOnError:
		return PresText
	case theme.ColorNameScrollBar:
		return PresSeparator
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *PresentationTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *PresentationTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *PresentationTheme) Size(name fyne.ThemeSizeName) float32 {
	// 1.3x text for projection readability; other sizes unchanged
	switch name {
	case theme.SizeNameText, theme.SizeNameCaptionText, theme.SizeNameSubHeadingText, theme.SizeNameHeadingText:
		return theme.DefaultTheme().Size(name) * 1.3
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// ============================================================================
// Utility Functions
// ============================================================================

// withAlpha returns a new color with the specified alpha channel (0-255)
func withAlpha(c color.Color, alpha uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: alpha,
	}
}

// GetContrastColor returns either white or black text based on background luminance
func GetContrastColor(bgColor color.Color) color.Color {
	r, g, b, _ := bgColor.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

	// Calculate relative luminance (simplified)
	luminance := (0.299*float64(r8) + 0.587*float64(g8) + 0.114*float64(b8)) / 255.0

	if luminance < 0.5 {
		return color.RGBA{255, 255, 255, 255}
	}
	return color.RGBA{0, 0, 0, 255}
}

// GetThemedColor returns a color that adapts to the current theme type
func GetThemedColor(t ThemeType, darkColor, lightColor, hcColor color.Color) color.Color {
	switch t {
	case ThemeLight:
		return lightColor
	case ThemeHighContrast:
		return hcColor
	default:
		return darkColor
	}
}
