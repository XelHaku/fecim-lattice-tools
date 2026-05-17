//go:build legacy_fyne

package theme

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func init() {
	// Initialize test app to ensure Fyne is available for tests
	test.NewApp()
}

func TestFeCIMThemeColors(t *testing.T) {
	th := &FeCIMTheme{}

	tests := []struct {
		name      string
		colorName fyne.ThemeColorName
		want      color.Color
	}{
		{
			name:      "Background is FeCIM blue",
			colorName: theme.ColorNameBackground,
			want:      ColorBackground,
		},
		{
			name:      "Primary is cyan",
			colorName: theme.ColorNamePrimary,
			want:      ColorPrimary,
		},
		{
			name:      "Button uses ColorButton",
			colorName: theme.ColorNameButton,
			want:      ColorButton,
		},
		{
			name:      "InputBackground uses ColorInput",
			colorName: theme.ColorNameInputBackground,
			want:      ColorInput,
		},
		{
			name:      "Separator uses ColorSeparator",
			colorName: theme.ColorNameSeparator,
			want:      ColorSeparator,
		},
		{
			name:      "Foreground uses ColorText",
			colorName: theme.ColorNameForeground,
			want:      ColorText,
		},
		{
			name:      "Success is green",
			colorName: theme.ColorNameSuccess,
			want:      ColorSuccess,
		},
		{
			name:      "Error is red",
			colorName: theme.ColorNameError,
			want:      ColorError,
		},
		{
			name:      "Warning is amber",
			colorName: theme.ColorNameWarning,
			want:      ColorWarning,
		},
		{
			name:      "Hover uses ColorButtonHover",
			colorName: theme.ColorNameHover,
			want:      ColorButtonHover,
		},
		{
			name:      "Placeholder uses ColorTextDim",
			colorName: theme.ColorNamePlaceHolder,
			want:      ColorTextDim,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := th.Color(tt.colorName, theme.VariantDark)
			if got != tt.want {
				t.Errorf("Color(%v) = %v, want %v", tt.colorName, got, tt.want)
			}
		})
	}
}

func TestFeCIMThemeFallback(t *testing.T) {
	th := &FeCIMTheme{}

	// Test that unknown color names fall back to default theme
	unknownColor := th.Color("unknown_color", theme.VariantDark)
	defaultColor := theme.DefaultTheme().Color("unknown_color", theme.VariantDark)

	if unknownColor != defaultColor {
		t.Errorf("Unknown color should fall back to default theme")
	}
}

func TestFeCIMTheme_VariantAwareColors(t *testing.T) {
	th := &FeCIMTheme{}

	darkBg := th.Color(theme.ColorNameBackground, theme.VariantDark)
	lightBg := th.Color(theme.ColorNameBackground, theme.VariantLight)
	if darkBg == lightBg {
		t.Fatalf("background color should differ between dark and light variants")
	}

	darkFg := th.Color(theme.ColorNameForeground, theme.VariantDark)
	lightFg := th.Color(theme.ColorNameForeground, theme.VariantLight)
	if darkFg == lightFg {
		t.Fatalf("foreground color should differ between dark and light variants")
	}

	darkMenu := th.Color(theme.ColorNameMenuBackground, theme.VariantDark)
	lightMenu := th.Color(theme.ColorNameMenuBackground, theme.VariantLight)
	if darkMenu == lightMenu {
		t.Fatalf("menu background should differ between dark and light variants")
	}
}

func TestFeCIMThemeFont(t *testing.T) {
	th := &FeCIMTheme{}

	// Should return same font as default theme
	defaultFont := theme.DefaultTheme().Font(fyne.TextStyle{})
	themeFont := th.Font(fyne.TextStyle{})

	if defaultFont != themeFont {
		t.Errorf("Font should delegate to default theme")
	}
}

func TestFeCIMThemeIcon(t *testing.T) {
	th := &FeCIMTheme{}

	// Should return same icon as default theme
	iconName := fyne.ThemeIconName("home")
	defaultIcon := theme.DefaultTheme().Icon(iconName)
	themeIcon := th.Icon(iconName)

	if defaultIcon != themeIcon {
		t.Errorf("Icon should delegate to default theme")
	}
}

func TestFeCIMThemeSize(t *testing.T) {
	th := &FeCIMTheme{}

	// Should return same size as default theme
	defaultSize := theme.DefaultTheme().Size(theme.SizeNamePadding)
	themeSize := th.Size(theme.SizeNamePadding)

	if defaultSize != themeSize {
		t.Errorf("Size should delegate to default theme")
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color constants have expected values
	if ColorPrimary.R != 0 || ColorPrimary.G != 212 || ColorPrimary.B != 255 {
		t.Errorf("ColorPrimary unexpected value: %v", ColorPrimary)
	}

	if ColorBackground.R != 0 || ColorBackground.G != 50 || ColorBackground.B != 100 {
		t.Errorf("ColorBackground unexpected value: %v", ColorBackground)
	}

	// Verify most colors have full alpha
	colors := []color.RGBA{
		ColorPrimary, ColorSecondary, ColorAccent, ColorWarning,
		ColorBackground, ColorAxis, ColorPositive, ColorNegative,
		ColorText, ColorSuccess, ColorError,
	}

	for i, c := range colors {
		if c.A != 255 {
			t.Errorf("Color %d has non-full alpha: %d", i, c.A)
		}
	}

	// Semi-transparent colors should have partial alpha
	semiTransparent := []color.RGBA{ColorGrid, ColorInputDisabled, ColorShadow}
	for i, c := range semiTransparent {
		if c.A == 255 {
			t.Errorf("Semi-transparent color %d should have partial alpha", i)
		}
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test WithAlpha
	testColor := color.RGBA{100, 150, 200, 255}
	alphaColor := WithAlpha(testColor, 128)
	alphaRGBA := alphaColor.(color.RGBA)
	if alphaRGBA.R != 100 || alphaRGBA.G != 150 || alphaRGBA.B != 200 || alphaRGBA.A != 128 {
		t.Errorf("WithAlpha failed: got %v, want RGBA{100,150,200,128}", alphaRGBA)
	}

	// Test GetContrastColor with dark background
	darkBG := color.RGBA{0, 50, 100, 255}
	if GetContrastColor(darkBG) != ColorText {
		t.Errorf("GetContrastColor should return ColorText for dark background")
	}

	// Test GetContrastColor with light background
	lightBG := color.RGBA{240, 244, 248, 255}
	if GetContrastColor(lightBG) != ColorBackground {
		t.Errorf("GetContrastColor should return ColorBackground for light background")
	}
}

func TestColorHierarchy(t *testing.T) {
	// Test that background hierarchy is correctly ordered (dark to light)
	// Extract green channel as proxy for brightness in blue color scheme
	bgGreen := ColorBackground.G
	surfaceGreen := ColorSurface.G
	inputGreen := ColorInput.G
	surfaceAltGreen := ColorSurfaceAlt.G

	if !(bgGreen < surfaceGreen && surfaceGreen < inputGreen && inputGreen < surfaceAltGreen) {
		t.Errorf("Background hierarchy not ordered correctly: BG=%d, Surface=%d, Input=%d, SurfaceAlt=%d",
			bgGreen, surfaceGreen, inputGreen, surfaceAltGreen)
	}
}
