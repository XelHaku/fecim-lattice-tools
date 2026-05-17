//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
// demo_mode_selector.go provides a UI for selecting and managing different
// demo modes including tutorials, animations, and quick demos.
package widgets

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DemoMode represents the type of demonstration.
type DemoMode int

const (
	DemoModeNone      DemoMode = iota
	DemoModeQuick              // Quick automated demo
	DemoModeTutorial           // Interactive step-by-step tutorial
	DemoModeAnimation          // Educational animation playback
	DemoModeSandbox            // Free exploration with hints
)

func (m DemoMode) String() string {
	switch m {
	case DemoModeNone:
		return "Normal"
	case DemoModeQuick:
		return "Quick Demo"
	case DemoModeTutorial:
		return "Tutorial"
	case DemoModeAnimation:
		return "Animation"
	case DemoModeSandbox:
		return "Sandbox"
	default:
		return "Unknown"
	}
}

// DemoModeInfo provides metadata about a demo mode.
type DemoModeInfo struct {
	Mode        DemoMode
	Name        string
	Description string
	Duration    time.Duration
	Icon        fyne.Resource
	Available   bool
}

// DemoModeSelector provides UI for selecting demo modes.
type DemoModeSelector struct {
	widget.BaseWidget

	currentMode DemoMode
	modes       []DemoModeInfo

	// UI elements
	modeList     *widget.List
	startBtn     *widget.Button
	detailsPanel *fyne.Container

	// State
	selectedIndex int

	// Callbacks
	onModeSelected func(mode DemoMode)
	onStart        func(mode DemoMode)
}

// NewDemoModeSelector creates a new demo mode selector.
func NewDemoModeSelector() *DemoModeSelector {
	s := &DemoModeSelector{
		modes: []DemoModeInfo{
			{
				Mode:        DemoModeQuick,
				Name:        "⚡ Quick Demo",
				Description: "30-second automated demonstration of key concepts. Perfect for first-time visitors.",
				Duration:    30 * time.Second,
				Available:   true,
			},
			{
				Mode:        DemoModeTutorial,
				Name:        "📚 Interactive Tutorial",
				Description: "Step-by-step guided learning with quizzes and hands-on exercises. Best for deep understanding.",
				Duration:    10 * time.Minute,
				Available:   true,
			},
			{
				Mode:        DemoModeAnimation,
				Name:        "🎬 Educational Animation",
				Description: "Watch animated explanations of physics and concepts. Good for visual learners.",
				Duration:    5 * time.Minute,
				Available:   true,
			},
			{
				Mode:        DemoModeSandbox,
				Name:        "🧪 Sandbox Mode",
				Description: "Free exploration with contextual hints. For experienced users who want to experiment.",
				Duration:    0, // No fixed duration
				Available:   true,
			},
		},
		selectedIndex: 0,
	}

	s.ExtendBaseWidget(s)
	return s
}

// SetOnModeSelected sets the callback for when a mode is selected.
func (s *DemoModeSelector) SetOnModeSelected(fn func(mode DemoMode)) {
	s.onModeSelected = fn
}

// SetOnStart sets the callback for when start is pressed.
func (s *DemoModeSelector) SetOnStart(fn func(mode DemoMode)) {
	s.onStart = fn
}

// SetModeAvailable enables/disables a specific mode.
func (s *DemoModeSelector) SetModeAvailable(mode DemoMode, available bool) {
	for i := range s.modes {
		if s.modes[i].Mode == mode {
			s.modes[i].Available = available
			break
		}
	}
	s.Refresh()
}

// GetCurrentMode returns the currently selected mode.
func (s *DemoModeSelector) GetCurrentMode() DemoMode {
	return s.currentMode
}

// CreateRenderer implements fyne.Widget.
func (s *DemoModeSelector) CreateRenderer() fyne.WidgetRenderer {
	// Mode list
	s.modeList = widget.NewList(
		func() int { return len(s.modes) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				layout.NewSpacer(),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			box := obj.(*fyne.Container)
			nameLabel := box.Objects[0].(*widget.Label)
			durationLabel := box.Objects[2].(*widget.Label)

			mode := s.modes[id]
			nameLabel.SetText(mode.Name)

			if mode.Duration > 0 {
				durationLabel.SetText(formatDuration(mode.Duration))
			} else {
				durationLabel.SetText("∞")
			}

			if !mode.Available {
				nameLabel.Importance = widget.LowImportance
			} else {
				nameLabel.Importance = widget.MediumImportance
			}
		},
	)

	s.modeList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(s.modes) {
			s.selectedIndex = int(id)
			s.currentMode = s.modes[id].Mode
			s.updateDetails()
			if s.onModeSelected != nil {
				s.onModeSelected(s.currentMode)
			}
		}
	}

	// Details panel
	s.detailsPanel = container.NewVBox(
		widget.NewLabel("Select a demo mode"),
	)

	// Start button
	s.startBtn = widget.NewButtonWithIcon("Start Demo", theme.MediaPlayIcon(), func() {
		if s.onStart != nil && s.selectedIndex >= 0 && s.selectedIndex < len(s.modes) {
			s.onStart(s.modes[s.selectedIndex].Mode)
		}
	})
	s.startBtn.Importance = widget.HighImportance

	// Layout
	content := container.NewBorder(
		widget.NewLabelWithStyle("Demo Modes", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		container.NewVBox(
			widget.NewSeparator(),
			s.detailsPanel,
			s.startBtn,
		),
		nil, nil,
		s.modeList,
	)

	return widget.NewSimpleRenderer(content)
}

func (s *DemoModeSelector) updateDetails() {
	if s.detailsPanel == nil || s.selectedIndex < 0 || s.selectedIndex >= len(s.modes) {
		return
	}

	mode := s.modes[s.selectedIndex]

	fyne.Do(func() {
		descLabel := widget.NewLabel(mode.Description)
		descLabel.Wrapping = fyne.TextWrapWord

		var durationText string
		if mode.Duration > 0 {
			durationText = fmt.Sprintf("Duration: ~%s", formatDuration(mode.Duration))
		} else {
			durationText = "Duration: Open-ended"
		}

		s.detailsPanel.Objects = []fyne.CanvasObject{
			widget.NewLabelWithStyle(mode.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			descLabel,
			widget.NewLabel(durationText),
		}
		s.detailsPanel.Refresh()

		s.startBtn.SetText(fmt.Sprintf("Start %s", mode.Mode.String()))
		if mode.Available {
			s.startBtn.Enable()
		} else {
			s.startBtn.Disable()
		}
	})
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

// DemoProgressPanel shows progress during demo playback.
type DemoProgressPanel struct {
	widget.BaseWidget

	title       *widget.Label
	step        *widget.Label
	progress    *widget.ProgressBar
	explanation *widget.Label
	controls    *fyne.Container

	// Control buttons
	pauseBtn *widget.Button
	stopBtn  *widget.Button
	skipBtn  *widget.Button

	// State
	isPaused bool

	// Callbacks
	onPause  func()
	onResume func()
	onStop   func()
	onSkip   func()
}

// NewDemoProgressPanel creates a new progress panel.
func NewDemoProgressPanel() *DemoProgressPanel {
	p := &DemoProgressPanel{
		title:       widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		step:        widget.NewLabel(""),
		progress:    widget.NewProgressBar(),
		explanation: widget.NewLabel(""),
	}

	p.explanation.Wrapping = fyne.TextWrapWord

	p.pauseBtn = widget.NewButtonWithIcon("", theme.MediaPauseIcon(), func() {
		p.togglePause()
	})

	p.stopBtn = widget.NewButtonWithIcon("", theme.MediaStopIcon(), func() {
		if p.onStop != nil {
			p.onStop()
		}
	})

	p.skipBtn = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		if p.onSkip != nil {
			p.onSkip()
		}
	})

	p.controls = container.NewHBox(
		p.pauseBtn,
		p.stopBtn,
		layout.NewSpacer(),
		p.skipBtn,
	)

	p.ExtendBaseWidget(p)
	return p
}

// SetCallbacks sets all control callbacks.
func (p *DemoProgressPanel) SetCallbacks(onPause, onResume, onStop, onSkip func()) {
	p.onPause = onPause
	p.onResume = onResume
	p.onStop = onStop
	p.onSkip = onSkip
}

// UpdateStep updates the display for a new step.
func (p *DemoProgressPanel) UpdateStep(stepNum, totalSteps int, title, explanation string) {
	fyne.Do(func() {
		p.title.SetText(title)
		p.step.SetText(fmt.Sprintf("Step %d of %d", stepNum+1, totalSteps))
		p.explanation.SetText(explanation)
		p.progress.SetValue(float64(stepNum+1) / float64(totalSteps))
	})
}

// SetProgress sets the progress bar value directly.
func (p *DemoProgressPanel) SetProgress(value float64) {
	fyne.Do(func() {
		p.progress.SetValue(value)
	})
}

func (p *DemoProgressPanel) togglePause() {
	p.isPaused = !p.isPaused
	if p.isPaused {
		p.pauseBtn.SetIcon(theme.MediaPlayIcon())
		if p.onPause != nil {
			p.onPause()
		}
	} else {
		p.pauseBtn.SetIcon(theme.MediaPauseIcon())
		if p.onResume != nil {
			p.onResume()
		}
	}
}

// CreateRenderer implements fyne.Widget.
func (p *DemoProgressPanel) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewVBox(
		p.title,
		p.step,
		p.progress,
		widget.NewSeparator(),
		p.explanation,
		layout.NewSpacer(),
		p.controls,
	)

	return widget.NewSimpleRenderer(container.NewPadded(content))
}

// TutorialCard displays a tutorial with its metadata.
type TutorialCard struct {
	widget.BaseWidget

	tutorial *InteractiveTutorial
	onClick  func()
}

// NewTutorialCard creates a new tutorial card.
func NewTutorialCard(tutorial InteractiveTutorial) *TutorialCard {
	c := &TutorialCard{
		tutorial: &tutorial,
	}
	c.ExtendBaseWidget(c)
	return c
}

// SetOnClick sets the click handler.
func (c *TutorialCard) SetOnClick(fn func()) {
	c.onClick = fn
}

// CreateRenderer implements fyne.Widget.
func (c *TutorialCard) CreateRenderer() fyne.WidgetRenderer {
	title := widget.NewLabelWithStyle(c.tutorial.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Importance = widget.HighImportance

	desc := widget.NewLabel(c.tutorial.Description)
	desc.Wrapping = fyne.TextWrapWord

	levelBadge := widget.NewLabel(c.tutorial.Difficulty.String())
	levelBadge.Importance = widget.LowImportance

	durationLabel := widget.NewLabel(fmt.Sprintf("~%s", formatDuration(c.tutorial.Duration)))

	meta := container.NewHBox(
		levelBadge,
		widget.NewLabel("•"),
		durationLabel,
		layout.NewSpacer(),
	)

	startBtn := widget.NewButtonWithIcon("Start", theme.MediaPlayIcon(), func() {
		if c.onClick != nil {
			c.onClick()
		}
	})

	// Learning goals
	goalsLabel := widget.NewLabel("You'll learn:")
	goalsLabel.TextStyle.Bold = true

	goalsList := container.NewVBox()
	for _, goal := range c.tutorial.LearningGoals {
		goalsList.Add(widget.NewLabel("• " + goal))
	}

	content := container.NewVBox(
		title,
		desc,
		widget.NewSeparator(),
		meta,
		widget.NewSeparator(),
		goalsLabel,
		goalsList,
		container.NewHBox(layout.NewSpacer(), startBtn),
	)

	return widget.NewSimpleRenderer(container.NewPadded(content))
}

// TutorialListWidget displays a list of available tutorials.
type TutorialListWidget struct {
	widget.BaseWidget

	tutorials []InteractiveTutorial
	onSelect  func(tutorial InteractiveTutorial)
}

// NewTutorialListWidget creates a new tutorial list.
func NewTutorialListWidget(tutorials []InteractiveTutorial) *TutorialListWidget {
	w := &TutorialListWidget{
		tutorials: tutorials,
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetOnSelect sets the selection callback.
func (w *TutorialListWidget) SetOnSelect(fn func(tutorial InteractiveTutorial)) {
	w.onSelect = fn
}

// CreateRenderer implements fyne.Widget.
func (w *TutorialListWidget) CreateRenderer() fyne.WidgetRenderer {
	list := container.NewVBox()

	for _, t := range w.tutorials {
		tutorial := t // Capture for closure
		card := NewTutorialCard(tutorial)
		card.SetOnClick(func() {
			if w.onSelect != nil {
				w.onSelect(tutorial)
			}
		})
		list.Add(card)
		list.Add(widget.NewSeparator())
	}

	scroll := container.NewScroll(list)
	scroll.SetMinSize(fyne.NewSize(300, 400))

	return widget.NewSimpleRenderer(scroll)
}

// AnimationListWidget displays available animations.
type AnimationListWidget struct {
	widget.BaseWidget

	animations []AnimationPreset
	onSelect   func(anim AnimationPreset)
}

// NewAnimationListWidget creates a new animation list.
func NewAnimationListWidget(animations []AnimationPreset) *AnimationListWidget {
	w := &AnimationListWidget{
		animations: animations,
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetOnSelect sets the selection callback.
func (w *AnimationListWidget) SetOnSelect(fn func(anim AnimationPreset)) {
	w.onSelect = fn
}

// CreateRenderer implements fyne.Widget.
func (w *AnimationListWidget) CreateRenderer() fyne.WidgetRenderer {
	list := widget.NewList(
		func() int { return len(w.animations) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(""),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			box := obj.(*fyne.Container)
			title := box.Objects[0].(*widget.Label)
			desc := box.Objects[1].(*widget.Label)

			anim := w.animations[id]
			title.SetText(anim.Name)
			desc.SetText(anim.Description)
			desc.Truncation = fyne.TextTruncateEllipsis
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		if w.onSelect != nil && id >= 0 && id < len(w.animations) {
			w.onSelect(w.animations[id])
		}
	}

	return widget.NewSimpleRenderer(list)
}

// DemoHintOverlay displays contextual hints during sandbox mode.
type DemoHintOverlay struct {
	widget.BaseWidget

	hints    []string
	current  int
	visible  bool
	hintText *widget.Label
}

// NewDemoHintOverlay creates a new hint overlay.
func NewDemoHintOverlay() *DemoHintOverlay {
	h := &DemoHintOverlay{
		hints:    []string{},
		current:  0,
		visible:  true,
		hintText: widget.NewLabel(""),
	}
	h.hintText.Wrapping = fyne.TextWrapWord
	h.hintText.Importance = widget.LowImportance
	h.ExtendBaseWidget(h)
	return h
}

// SetHints sets the available hints.
func (h *DemoHintOverlay) SetHints(hints []string) {
	h.hints = hints
	h.current = 0
	h.updateHint()
}

// NextHint advances to the next hint.
func (h *DemoHintOverlay) NextHint() {
	if len(h.hints) > 0 {
		h.current = (h.current + 1) % len(h.hints)
		h.updateHint()
	}
}

// SetVisible shows/hides the hints.
func (h *DemoHintOverlay) SetVisible(visible bool) {
	h.visible = visible
	h.Refresh()
}

func (h *DemoHintOverlay) updateHint() {
	if len(h.hints) > 0 {
		fyne.Do(func() {
			h.hintText.SetText(fmt.Sprintf("💡 Hint %d/%d: %s", h.current+1, len(h.hints), h.hints[h.current]))
		})
	}
}

// CreateRenderer implements fyne.Widget.
func (h *DemoHintOverlay) CreateRenderer() fyne.WidgetRenderer {
	nextBtn := widget.NewButton("Next Hint", func() {
		h.NextHint()
	})
	nextBtn.Importance = widget.LowImportance

	hideBtn := widget.NewButton("Hide", func() {
		h.SetVisible(false)
	})
	hideBtn.Importance = widget.LowImportance

	content := container.NewVBox(
		h.hintText,
		container.NewHBox(nextBtn, hideBtn),
	)

	return widget.NewSimpleRenderer(container.NewPadded(content))
}
