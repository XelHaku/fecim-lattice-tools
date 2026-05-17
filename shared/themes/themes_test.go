//go:build legacy_fyne

package themes

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func init() {
	test.NewApp()
}

func TestAllThemes(t *testing.T) {
	themes := AllThemes()
	if len(themes) != 4 {
		t.Errorf("Expected 4 themes, got %d", len(themes))
	}

	expected := []ThemeType{ThemeDark, ThemeLight, ThemeHighContrast, ThemePresentation}
	for i, th := range themes {
		if th != expected[i] {
			t.Errorf("Theme %d: expected %s, got %s", i, expected[i], th)
		}
	}
}

func TestThemeDisplayName(t *testing.T) {
	tests := []struct {
		theme ThemeType
		want  string
	}{
		{ThemeDark, "Dark (FeCIM)"},
		{ThemeLight, "Light"},
		{ThemeHighContrast, "High Contrast"},
		{ThemePresentation, "Presentation"},
		{ThemeType("unknown"), "unknown"},
	}

	for _, tt := range tests {
		got := ThemeDisplayName(tt.theme)
		if got != tt.want {
			t.Errorf("ThemeDisplayName(%s) = %s, want %s", tt.theme, got, tt.want)
		}
	}
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		themeType ThemeType
		wantType  string
	}{
		{ThemeDark, "*themes.DarkTheme"},
		{ThemeLight, "*themes.LightTheme"},
		{ThemeHighContrast, "*themes.HighContrastTheme"},
		{ThemePresentation, "*themes.PresentationTheme"},
		{ThemeType("unknown"), "*themes.DarkTheme"}, // Falls back to dark
	}

	for _, tt := range tests {
		got := GetTheme(tt.themeType)
		gotType := getTypeName(got)
		if gotType != tt.wantType {
			t.Errorf("GetTheme(%s) returned %s, want %s", tt.themeType, gotType, tt.wantType)
		}
	}
}

func getTypeName(v interface{}) string {
	switch v.(type) {
	case *DarkTheme:
		return "*themes.DarkTheme"
	case *LightTheme:
		return "*themes.LightTheme"
	case *HighContrastTheme:
		return "*themes.HighContrastTheme"
	case *PresentationTheme:
		return "*themes.PresentationTheme"
	default:
		return "unknown"
	}
}

func TestDarkThemeColors(t *testing.T) {
	th := &DarkTheme{}

	tests := []struct {
		name      string
		colorName fyne.ThemeColorName
		want      color.Color
	}{
		{"Background", theme.ColorNameBackground, DarkBackground},
		{"Foreground", theme.ColorNameForeground, DarkText},
		{"Primary", theme.ColorNamePrimary, DarkPrimary},
		{"Success", theme.ColorNameSuccess, DarkSuccess},
		{"Error", theme.ColorNameError, DarkError},
		{"Warning", theme.ColorNameWarning, DarkWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := th.Color(tt.colorName, theme.VariantDark)
			if !colorEqual(got, tt.want) {
				t.Errorf("Color(%s) = %v, want %v", tt.colorName, got, tt.want)
			}
		})
	}
}

func TestLightThemeColors(t *testing.T) {
	th := &LightTheme{}

	tests := []struct {
		name      string
		colorName fyne.ThemeColorName
		want      color.Color
	}{
		{"Background", theme.ColorNameBackground, LightBackground},
		{"Foreground", theme.ColorNameForeground, LightText},
		{"Primary", theme.ColorNamePrimary, LightPrimary},
		{"Success", theme.ColorNameSuccess, LightSuccess},
		{"Error", theme.ColorNameError, LightError},
		{"Warning", theme.ColorNameWarning, LightWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := th.Color(tt.colorName, theme.VariantLight)
			if !colorEqual(got, tt.want) {
				t.Errorf("Color(%s) = %v, want %v", tt.colorName, got, tt.want)
			}
		})
	}
}

func TestHighContrastThemeColors(t *testing.T) {
	th := &HighContrastTheme{}

	tests := []struct {
		name      string
		colorName fyne.ThemeColorName
		want      color.Color
	}{
		{"Background", theme.ColorNameBackground, HCBackground},
		{"Foreground", theme.ColorNameForeground, HCText},
		{"Primary", theme.ColorNamePrimary, HCPrimary},
		{"Success", theme.ColorNameSuccess, HCSuccess},
		{"Error", theme.ColorNameError, HCError},
		{"Warning", theme.ColorNameWarning, HCWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := th.Color(tt.colorName, theme.VariantDark)
			if !colorEqual(got, tt.want) {
				t.Errorf("Color(%s) = %v, want %v", tt.colorName, got, tt.want)
			}
		})
	}
}

func TestHighContrastThemeLargerText(t *testing.T) {
	hc := &HighContrastTheme{}
	dark := &DarkTheme{}

	// High contrast should have larger text
	hcTextSize := hc.Size(theme.SizeNameText)
	darkTextSize := dark.Size(theme.SizeNameText)

	if hcTextSize <= darkTextSize {
		t.Errorf("High contrast text size (%f) should be larger than dark theme (%f)",
			hcTextSize, darkTextSize)
	}
}

func TestGetContrastColor(t *testing.T) {
	tests := []struct {
		name      string
		bgColor   color.Color
		wantLight bool // true = expect white, false = expect black
	}{
		{"Dark background", color.RGBA{0, 50, 100, 255}, true},
		{"Light background", color.RGBA{240, 244, 248, 255}, false},
		{"Pure black", color.RGBA{0, 0, 0, 255}, true},
		{"Pure white", color.RGBA{255, 255, 255, 255}, false},
		{"Mid gray", color.RGBA{128, 128, 128, 255}, false}, // Slightly above 0.5 luminance
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetContrastColor(tt.bgColor)
			r, g, b, _ := got.RGBA()
			isWhite := r == 0xFFFF && g == 0xFFFF && b == 0xFFFF

			if tt.wantLight && !isWhite {
				t.Errorf("GetContrastColor(%v) should return white for dark background", tt.bgColor)
			}
			if !tt.wantLight && isWhite {
				t.Errorf("GetContrastColor(%v) should return black for light background", tt.bgColor)
			}
		})
	}
}

func TestGetThemedColor(t *testing.T) {
	dark := color.RGBA{0, 50, 100, 255}
	light := color.RGBA{240, 244, 248, 255}
	hc := color.RGBA{0, 255, 255, 255}

	tests := []struct {
		themeType ThemeType
		want      color.Color
	}{
		{ThemeDark, dark},
		{ThemeLight, light},
		{ThemeHighContrast, hc},
	}

	for _, tt := range tests {
		t.Run(string(tt.themeType), func(t *testing.T) {
			got := GetThemedColor(tt.themeType, dark, light, hc)
			if !colorEqual(got, tt.want) {
				t.Errorf("GetThemedColor(%s) = %v, want %v", tt.themeType, got, tt.want)
			}
		})
	}
}

func TestThemeFallbackToDefault(t *testing.T) {
	themes := []fyne.Theme{&DarkTheme{}, &LightTheme{}, &HighContrastTheme{}, &PresentationTheme{}}
	unknownColor := fyne.ThemeColorName("unknown_color_xyz")

	for _, th := range themes {
		got := th.Color(unknownColor, theme.VariantDark)
		expected := theme.DefaultTheme().Color(unknownColor, theme.VariantDark)
		if !colorEqual(got, expected) {
			t.Errorf("Unknown color should fall back to default theme")
		}
	}
}

func TestThemeFont(t *testing.T) {
	themes := []fyne.Theme{&DarkTheme{}, &LightTheme{}, &HighContrastTheme{}, &PresentationTheme{}}
	defaultFont := theme.DefaultTheme().Font(fyne.TextStyle{})

	for _, th := range themes {
		got := th.Font(fyne.TextStyle{})
		if got != defaultFont {
			t.Errorf("Font should delegate to default theme")
		}
	}
}

func TestThemeIcon(t *testing.T) {
	themes := []fyne.Theme{&DarkTheme{}, &LightTheme{}, &HighContrastTheme{}, &PresentationTheme{}}
	iconName := fyne.ThemeIconName("home")
	defaultIcon := theme.DefaultTheme().Icon(iconName)

	for _, th := range themes {
		got := th.Icon(iconName)
		if got != defaultIcon {
			t.Errorf("Icon should delegate to default theme")
		}
	}
}

func TestColorConstants(t *testing.T) {
	// Verify key color constants
	tests := []struct {
		name       string
		color      color.RGBA
		r, g, b, a uint8
	}{
		{"DarkPrimary", DarkPrimary, 0, 212, 255, 255},
		{"DarkBackground", DarkBackground, 0, 50, 100, 255},
		{"LightPrimary", LightPrimary, 0, 120, 180, 255},
		{"LightBackground", LightBackground, 250, 252, 255, 255},
		{"HCPrimary", HCPrimary, 0, 255, 255, 255},
		{"HCBackground", HCBackground, 0, 0, 0, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color.R != tt.r || tt.color.G != tt.g || tt.color.B != tt.b || tt.color.A != tt.a {
				t.Errorf("%s = RGBA(%d,%d,%d,%d), want RGBA(%d,%d,%d,%d)",
					tt.name, tt.color.R, tt.color.G, tt.color.B, tt.color.A,
					tt.r, tt.g, tt.b, tt.a)
			}
		})
	}
}

// colorEqual compares two colors for equality
func colorEqual(a, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
