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
			name:      "Button is lighter blue",
			colorName: theme.ColorNameButton,
			want:      color.RGBA{0, 70, 130, 255},
		},
		{
			name:      "InputBackground is darker blue",
			colorName: theme.ColorNameInputBackground,
			want:      color.RGBA{0, 40, 80, 255},
		},
		{
			name:      "Separator is visible",
			colorName: theme.ColorNameSeparator,
			want:      color.RGBA{0, 80, 150, 255},
		},
		{
			name:      "Foreground is light gray",
			colorName: theme.ColorNameForeground,
			want:      color.RGBA{230, 230, 230, 255},
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

	// Verify all colors have full alpha
	colors := []color.RGBA{
		ColorPrimary, ColorSecondary, ColorAccent, ColorWarning,
		ColorBackground, ColorAxis, ColorPositive, ColorNegative,
	}

	for i, c := range colors {
		if c.A != 255 {
			t.Errorf("Color %d has non-full alpha: %d", i, c.A)
		}
	}

	// Grid color should have partial alpha for transparency
	if ColorGrid.A == 255 {
		t.Errorf("ColorGrid should have partial alpha for transparency")
	}
}
