// Package theme provides the shared FeCIM theme for consistent branding across all demos.
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// FeCIM theme colors - shared across all demos
var (
	ColorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan
	ColorSecondary  = color.RGBA{255, 107, 107, 255} // Coral red
	ColorAccent     = color.RGBA{78, 205, 196, 255}  // Teal
	ColorWarning    = color.RGBA{255, 230, 109, 255} // Yellow
	ColorBackground = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
	ColorGrid       = color.RGBA{0, 70, 130, 128}    // Grid lines (lighter blue)
	ColorAxis       = color.RGBA{150, 180, 200, 255} // Axis lines
	ColorPositive   = color.RGBA{255, 100, 100, 255} // Positive polarization
	ColorNegative   = color.RGBA{100, 150, 255, 255} // Negative polarization
)

// FeCIMTheme implements fyne.Theme for consistent FeCIM branding
type FeCIMTheme struct{}

// Color returns the appropriate color for the given theme color name
func (t *FeCIMTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorBackground // FeCIM blue #003264
	case theme.ColorNameForeground:
		return color.RGBA{230, 230, 230, 255}
	case theme.ColorNamePrimary:
		return ColorPrimary
	case theme.ColorNameButton:
		return color.RGBA{0, 70, 130, 255} // Slightly lighter blue
	case theme.ColorNameInputBackground:
		return color.RGBA{0, 40, 80, 255} // Darker blue for inputs
	case theme.ColorNameSeparator:
		return color.RGBA{0, 80, 150, 255} // Separator lines
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
