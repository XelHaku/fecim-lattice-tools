//go:build legacy_fyne

package themes

import (
	"image/color"

	"fyne.io/fyne/v2"
)

// ScaledTheme wraps a base theme and scales text/control sizes.
type ScaledTheme struct {
	base  fyne.Theme
	scale float32
}

func NewScaledTheme(base fyne.Theme, scale float32) *ScaledTheme {
	if scale <= 0 {
		scale = 1.0
	}
	return &ScaledTheme{base: base, scale: scale}
}

func (t *ScaledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return t.base.Color(name, variant)
}

func (t *ScaledTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t *ScaledTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t *ScaledTheme) Size(name fyne.ThemeSizeName) float32 {
	return t.base.Size(name) * t.scale
}
