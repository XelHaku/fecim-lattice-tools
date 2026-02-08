// Package help provides startup tips and quick hints.
package help

import (
	"fmt"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// QuickTip represents a single tip to show users.
type QuickTip struct {
	Title   string
	Content string
	Icon    fyne.Resource
	Module  string // Optional: which module this tip relates to
}

// quickTips is the collection of startup tips.
var quickTips = []QuickTip{
	{
		Title:   "Press F1 for Help",
		Content: "At any time, press F1 to get context-sensitive help for the current screen. Press Shift+F1 to open the full help browser.",
		Icon:    theme.HelpIcon(),
	},
	{
		Title:   "Explore the Modules",
		Content: "FeCIM Lattice Tools has 6 interactive modules. Start with Hysteresis to understand how ferroelectric memory works, then progress through Crossbar, MNIST, and more.",
		Icon:    theme.ListIcon(),
	},
	{
		Title:   "Search Documentation",
		Content: "In the Documentation tab, press Ctrl+K to quickly search across all docs. The search is fuzzy, so don't worry about exact spelling!",
		Icon:    theme.SearchIcon(),
	},
	{
		Title:   "Interactive Simulations",
		Content: "Most modules have interactive sliders and controls. Experiment freely! You can always reset to defaults. Look for play/pause buttons to control animations.",
		Icon:    theme.MediaPlayIcon(),
	},
	{
		Title:   "Glossary Terms",
		Content: "Technical terms in the documentation are clickable. Click on highlighted terms to see their definitions without leaving the current page.",
		Icon:    theme.InfoIcon(),
	},
	{
		Title:   "Screenshots & Recording",
		Content: "Use the Screenshot button in the toolbar to capture your experiments. The Record button captures video with audio for creating presentations.",
		Icon:    theme.MediaPhotoIcon(),
	},
	{
		Title:   "The Hysteresis Loop",
		Content: "The P-E hysteresis loop is fundamental to FeCIM. Watch how polarization (P) responds to electric field (E) - the loop shape determines memory characteristics.",
		Icon:    theme.DocumentIcon(),
		Module:  "hysteresis",
	},
	{
		Title:   "Matrix Magic",
		Content: "In the Crossbar module, see how physics performs matrix multiplication in a single step. Current = Voltage × Conductance - it's that simple!",
		Icon:    theme.ComputerIcon(),
		Module:  "crossbar",
	},
	{
		Title:   "Draw Your Own Digits",
		Content: "In the MNIST module, draw your own handwritten digits and watch the neural network classify them in real-time. Compare floating-point vs. hardware accuracy.",
		Icon:    theme.ContentPasteIcon(),
		Module:  "mnist",
	},
	{
		Title:   "Energy Efficiency",
		Content: "FeCIM can be 10-1000× more energy efficient than GPUs for AI workloads. Check out the Comparison module to see the numbers.",
		Icon:    theme.SettingsIcon(),
		Module:  "comparison",
	},
	{
		Title:   "Keyboard Navigation",
		Content: "Use Tab to move between controls, Arrow keys to adjust sliders, Space/Enter to activate buttons. The app is fully keyboard accessible.",
		Icon:    theme.NavigateNextIcon(),
	},
	{
		Title:   "ELI5 Documents",
		Content: "Each module has an 'ELI5.md' (Explain Like I'm 5) document with simplified explanations. Great for getting started with complex topics!",
		Icon:    theme.AccountIcon(),
	},
}

// TipManager handles startup tips display.
type TipManager struct {
	tips       []QuickTip
	prefs      fyne.Preferences
	window     fyne.Window
	currentIdx int
	showOnStart bool
}

// NewTipManager creates a new tip manager.
func NewTipManager(prefs fyne.Preferences, window fyne.Window) *TipManager {
	return &TipManager{
		tips:        quickTips,
		prefs:       prefs,
		window:      window,
		showOnStart: prefs.BoolWithFallback("show_tips_on_startup", true),
	}
}

// ShouldShowOnStartup returns whether tips should be shown at startup.
func (tm *TipManager) ShouldShowOnStartup() bool {
	return tm.showOnStart
}

// SetShowOnStartup sets whether tips should be shown at startup.
func (tm *TipManager) SetShowOnStartup(show bool) {
	tm.showOnStart = show
	tm.prefs.SetBool("show_tips_on_startup", show)
}

// GetRandomTip returns a random tip.
func (tm *TipManager) GetRandomTip() QuickTip {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return tm.tips[r.Intn(len(tm.tips))]
}

// GetTipForModule returns a tip relevant to a specific module.
func (tm *TipManager) GetTipForModule(module string) *QuickTip {
	var moduleTips []QuickTip
	for _, tip := range tm.tips {
		if tip.Module == module {
			moduleTips = append(moduleTips, tip)
		}
	}
	if len(moduleTips) == 0 {
		return nil
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tip := moduleTips[r.Intn(len(moduleTips))]
	return &tip
}

// ShowStartupTip shows the startup tip dialog.
func (tm *TipManager) ShowStartupTip() {
	if !tm.showOnStart || tm.window == nil {
		return
	}
	
	tm.currentIdx = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(tm.tips))
	tm.showTipDialog()
}

func (tm *TipManager) showTipDialog() {
	tip := tm.tips[tm.currentIdx]

	// Title with icon
	icon := canvas.NewImageFromResource(tip.Icon)
	icon.SetMinSize(fyne.NewSize(32, 32))
	icon.FillMode = canvas.ImageFillContain

	titleLabel := widget.NewLabelWithStyle(tip.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleRow := container.NewHBox(icon, titleLabel)

	// Content
	contentLabel := widget.NewLabel(tip.Content)
	contentLabel.Wrapping = fyne.TextWrapWord

	// Navigation
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		tm.currentIdx--
		if tm.currentIdx < 0 {
			tm.currentIdx = len(tm.tips) - 1
		}
		tm.showTipDialog()
	})

	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		tm.currentIdx++
		if tm.currentIdx >= len(tm.tips) {
			tm.currentIdx = 0
		}
		tm.showTipDialog()
	})

	tipCounter := widget.NewLabel("")
	tipCounter.SetText(fmt.Sprintf("Tip %d of %d", tm.currentIdx+1, len(tm.tips)))

	navRow := container.NewHBox(
		prevBtn,
		layout.NewSpacer(),
		tipCounter,
		layout.NewSpacer(),
		nextBtn,
	)

	// Show on startup checkbox
	showCheck := widget.NewCheck("Show tips on startup", func(checked bool) {
		tm.SetShowOnStartup(checked)
	})
	showCheck.SetChecked(tm.showOnStart)

	// Full content
	content := container.NewVBox(
		titleRow,
		widget.NewSeparator(),
		contentLabel,
		widget.NewSeparator(),
		navRow,
		showCheck,
	)

	contentBox := container.NewPadded(content)

	d := dialog.NewCustom("💡 Did You Know?", "Got it!", contentBox, tm.window)
	d.Resize(fyne.NewSize(450, 280))
	d.Show()
}

// TipBanner creates a small tip banner widget for inline display.
type TipBanner struct {
	widget.BaseWidget
	tip       QuickTip
	onDismiss func()
}

// NewTipBanner creates a dismissable tip banner.
func NewTipBanner(tip QuickTip, onDismiss func()) *TipBanner {
	tb := &TipBanner{
		tip:       tip,
		onDismiss: onDismiss,
	}
	tb.ExtendBaseWidget(tb)
	return tb
}

// CreateRenderer implements fyne.Widget.
func (tb *TipBanner) CreateRenderer() fyne.WidgetRenderer {
	icon := widget.NewIcon(tb.tip.Icon)
	title := widget.NewLabelWithStyle(tb.tip.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	content := widget.NewLabel(tb.tip.Content)
	content.Wrapping = fyne.TextWrapWord

	dismissBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		if tb.onDismiss != nil {
			tb.onDismiss()
		}
	})
	dismissBtn.Importance = widget.LowImportance

	leftContent := container.NewHBox(icon, container.NewVBox(title, content))
	c := container.NewBorder(nil, nil, nil, dismissBtn, leftContent)
	
	// Add background
	bg := canvas.NewRectangle(theme.OverlayBackgroundColor())
	full := container.NewStack(bg, container.NewPadded(c))

	return widget.NewSimpleRenderer(full)
}
