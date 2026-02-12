package help

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestHelpSystemCallbacksAndBrowserPaths(t *testing.T) {
	hs := &HelpSystem{topics: map[string]*HelpTopic{}, contextMap: map[string]string{}}
	overview := &HelpTopic{ID: "overview", Title: "Overview", Category: "Getting Started", Content: "# Overview"}
	hysteresis := &HelpTopic{ID: "module.hysteresis", Title: "Hysteresis", Category: "Modules", Content: "# H", ContextKeys: []string{"hysteresis"}}
	hs.RegisterTopic(overview)
	hs.RegisterTopic(hysteresis)

	var shownTopic *HelpTopic
	browserShown := 0
	hs.SetCallbacks(func(topic *HelpTopic) { shownTopic = topic }, func() { browserShown++ })

	hs.SetContext("hysteresis")
	hs.ShowContextualHelp()
	if shownTopic == nil || shownTopic.ID != "module.hysteresis" {
		t.Fatalf("expected contextual topic callback for hysteresis, got %#v", shownTopic)
	}

	shownTopic = nil
	hs.onShowHelp = nil // force browser fallback path
	hs.SetContext("unknown")
	hs.ShowContextualHelp()
	if browserShown == 0 {
		t.Fatal("expected browser fallback callback to be invoked")
	}

	hs.ShowBrowser()
	if browserShown < 2 {
		t.Fatal("expected explicit ShowBrowser callback invocation")
	}
}

func TestHelpSystemInitAndContextFallbacks(t *testing.T) {
	hs := &HelpSystem{topics: map[string]*HelpTopic{}, contextMap: map[string]string{}}

	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("help")
	hs.Init(w)
	if hs.window == nil {
		t.Fatal("Init should store window")
	}

	// No topic and no overview: nil fallback
	hs.SetContext("missing")
	if got := hs.GetContextualTopic(); got != nil {
		t.Fatalf("expected nil contextual topic when no topics exist, got %#v", got)
	}

	overview := &HelpTopic{ID: "overview", Title: "Overview", Category: "Getting Started", Content: "# Overview"}
	hs.RegisterTopic(overview)
	if got := hs.GetContextualTopic(); got == nil || got.ID != "overview" {
		t.Fatalf("expected overview fallback topic, got %#v", got)
	}

	// Empty search should return all topics path
	if got := hs.Search(""); len(got) != 1 {
		t.Fatalf("empty search should return all topics, got %d", len(got))
	}
}

func TestSetupF1ShortcutNilAndValidWindow(t *testing.T) {
	// nil guard branches
	SetupF1Shortcut(nil, nil)

	app := test.NewApp()
	defer app.Quit()
	w := app.NewWindow("shortcuts")
	hs := &HelpSystem{topics: map[string]*HelpTopic{}, contextMap: map[string]string{}}

	SetupF1Shortcut(w, hs)

	// Ensure helper functions are callable from this setup
	hs.SetCallbacks(func(*HelpTopic) {}, func() {})
	hs.ShowBrowser()
}

func TestTipManagerStartupPreferencesAndModuleTips(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	prefs := app.Preferences()
	w := app.NewWindow("tips")
	tm := NewTipManager(prefs, w)

	if !tm.ShouldShowOnStartup() {
		t.Fatal("expected default show-on-startup=true")
	}

	tm.SetShowOnStartup(false)
	if tm.ShouldShowOnStartup() {
		t.Fatal("expected show-on-startup=false after SetShowOnStartup")
	}
	if prefs.BoolWithFallback("show_tips_on_startup", true) {
		t.Fatal("expected preference persisted as false")
	}

	// Guard branch: should return early with startup tips disabled.
	tm.ShowStartupTip()

	// Module-specific and missing-module paths.
	if tip := tm.GetTipForModule("hysteresis"); tip == nil {
		t.Fatal("expected module tip for hysteresis")
	}
	if tip := tm.GetTipForModule("does-not-exist"); tip != nil {
		t.Fatalf("expected nil tip for unknown module, got %#v", tip)
	}

	if tip := tm.GetRandomTip(); tip.Title == "" || tip.Content == "" {
		t.Fatalf("GetRandomTip returned invalid tip: %#v", tip)
	}
}

func TestTipBannerRendererAndDismiss(t *testing.T) {
	called := 0
	tip := QuickTip{Title: "T", Content: "C", Icon: fyne.NewStaticResource("x", []byte{1, 2, 3})}
	banner := NewTipBanner(tip, func() { called++ })
	r := banner.CreateRenderer()
	if r == nil {
		t.Fatal("CreateRenderer returned nil")
	}
	if len(r.Objects()) == 0 {
		t.Fatal("renderer should expose objects")
	}
	// invoke dismiss callback directly to validate state management path
	banner.onDismiss()
	if called != 1 {
		t.Fatalf("expected dismiss callback once, got %d", called)
	}
}
