package themes

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestScaledThemeDefaultsAndScaling(t *testing.T) {
	base := GetTheme(ThemeDark)
	scaled := NewScaledTheme(base, 0)
	if scaled.scale != 1.0 {
		t.Fatalf("expected non-positive scale to clamp to 1.0, got %f", scaled.scale)
	}

	scaled = NewScaledTheme(base, 1.5)
	if got, want := scaled.Size("text"), base.Size("text")*1.5; got != want {
		t.Fatalf("scaled size = %f, want %f", got, want)
	}
}

func TestThemePreviewCardAndRenderer(t *testing.T) {
	card := NewThemePreviewCard(ThemeDark)
	if got := card.MinSize(); got != fyne.NewSize(280, 80) {
		t.Fatalf("unexpected min size: %+v", got)
	}

	r := card.CreateRenderer().(*themePreviewRenderer)
	card.Resize(fyne.NewSize(320, 100))
	r.Layout(card.Size())
	objs := r.Objects()
	if len(objs) < 9 {
		t.Fatalf("expected preview objects to be built, got %d", len(objs))
	}

	if _, ok := objs[0].(*canvas.Rectangle); !ok {
		t.Fatal("first object should be background rectangle")
	}

	card.SetTheme(ThemeHighContrast)
	r.Refresh()
	if len(r.Objects()) == 0 {
		t.Fatal("refresh should rebuild preview objects")
	}
	r.Destroy()
}

func TestThemeSwitcherAndCompactSwitcher(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	m := NewManager(app)

	ts := NewThemeSwitcher(m)
	if ts.selector.Selected != ThemeDisplayName(ThemeDark) {
		t.Fatalf("expected selector to start at dark theme, got %q", ts.selector.Selected)
	}

	ts.selector.SetSelected(ThemeDisplayName(ThemeLight))
	if got := m.CurrentTheme(); got != ThemeLight {
		t.Fatalf("selector selection should update manager theme, got %s", got)
	}

	_ = ts.CreateRenderer()
	ts.Destroy()

	cts := NewCompactThemeSwitcher(m)
	if cts.getIcon() != "☀" {
		t.Fatalf("expected light icon, got %q", cts.getIcon())
	}
	m.SetTheme(ThemeHighContrast)
	if cts.getIcon() != "◐" {
		t.Fatalf("expected high contrast icon, got %q", cts.getIcon())
	}
	_ = cts.CreateRenderer()
	cts.Destroy()
}

func TestCreateSettingsSectionAndQuickToggle(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	m := NewManager(app)

	obj := CreateSettingsSection(m)
	if obj == nil {
		t.Fatal("CreateSettingsSection returned nil")
	}

	btn := CreateQuickToggle(m)
	if btn == nil {
		t.Fatal("CreateQuickToggle returned nil")
	}
	initial := btn.Text
	btn.OnTapped()
	if btn.Text == initial {
		t.Fatalf("quick toggle should update button label, still %q", btn.Text)
	}
}

func TestWithAlphaAndGetThemedColorDefault(t *testing.T) {
	base := color.RGBA{R: 1, G: 2, B: 3, A: 255}
	got := withAlpha(base, 7).(color.RGBA)
	if got.R != 1 || got.G != 2 || got.B != 3 || got.A != 7 {
		t.Fatalf("withAlpha mismatch: got %#v", got)
	}

	dark := color.RGBA{10, 10, 10, 255}
	light := color.RGBA{240, 240, 240, 255}
	hc := color.RGBA{0, 255, 255, 255}
	if c := GetThemedColor(ThemeType("unknown"), dark, light, hc); c != dark {
		t.Fatalf("unknown theme should default to dark color")
	}
}

func TestCreateQuickToggleReturnsButton(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	m := NewManager(app)

	btn := CreateQuickToggle(m)
	if _, ok := interface{}(btn).(*widget.Button); !ok {
		t.Fatal("expected widget.Button")
	}
}
