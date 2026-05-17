//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
// tutorial_controller.go implements an interactive tutorial framework with
// step-by-step explanations, highlighting, and progress tracking.
package widgets

import (
	"context"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// TutorialStep represents a single step in an interactive tutorial.
type TutorialStep struct {
	// Core content
	Title       string // Step title displayed prominently
	Explanation string // Detailed explanation of what's happening
	Action      func() // Action to execute for this step

	// Optional features
	HighlightElement string           // Element ID to highlight (for UI highlighting)
	Animation        *AnimationConfig // Optional animation to play
	UserPrompt       string           // Prompt user for action (empty = auto-advance)
	QuizQuestion     *QuizQuestion    // Optional quiz for comprehension check
	Duration         time.Duration    // Wait time before next step (0 = wait for user)
	SkippedWhenFast  bool             // Skip this step in fast mode
	Tags             []string         // Tags for filtering (e.g., "physics", "engineering")
	DifficultyLevel  TutorialLevel    // Beginner/Intermediate/Advanced
	Annotations      []StepAnnotation // Visual annotations to display
}

// TutorialLevel represents tutorial difficulty.
type TutorialLevel int

const (
	LevelBeginner TutorialLevel = iota
	LevelIntermediate
	LevelAdvanced
	LevelExpert
)

func (l TutorialLevel) String() string {
	switch l {
	case LevelBeginner:
		return "Beginner"
	case LevelIntermediate:
		return "Intermediate"
	case LevelAdvanced:
		return "Advanced"
	case LevelExpert:
		return "Expert"
	default:
		return "Unknown"
	}
}

// StepAnnotation is a visual annotation on the visualization.
type StepAnnotation struct {
	ID       string // Unique identifier
	Type     string // "arrow", "circle", "box", "text"
	Position fyne.Position
	Size     fyne.Size
	Color    string // Hex color
	Text     string // For text annotations
}

// QuizQuestion is an optional comprehension check.
type QuizQuestion struct {
	Question    string
	Options     []string
	CorrectIdx  int
	Explanation string // Shown after answering
}

// AnimationConfig defines an animation to play during a step.
type AnimationConfig struct {
	Name       string        // Animation identifier
	Duration   time.Duration // Total animation duration
	FrameDelay time.Duration // Delay between frames
	Loop       bool          // Whether to loop
	Params     map[string]interface{}
}

// TutorialController manages interactive tutorial playback with rich features.
type TutorialController struct {
	steps       []TutorialStep
	currentStep int
	running     bool
	paused      bool
	fastMode    bool
	levelFilter TutorialLevel
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex

	// Progress tracking
	completedSteps map[int]bool
	quizScores     map[int]bool
	startTime      time.Time
	stepTimes      []time.Duration

	// Callbacks
	onStepStart  func(step int, total int, ts TutorialStep)
	onStepEnd    func(step int, total int, ts TutorialStep)
	onComplete   func(stats TutorialStats)
	onQuizAnswer func(step int, correct bool)
	onHighlight  func(elementID string, highlight bool)
	onAnnotation func(annotations []StepAnnotation, show bool)
}

// TutorialStats contains statistics about a completed tutorial.
type TutorialStats struct {
	TotalSteps      int
	CompletedSteps  int
	SkippedSteps    int
	QuizCorrect     int
	QuizTotal       int
	TotalTime       time.Duration
	AverageStepTime time.Duration
}

// NewTutorialController creates a new tutorial controller.
func NewTutorialController(steps []TutorialStep) *TutorialController {
	return &TutorialController{
		steps:          steps,
		completedSteps: make(map[int]bool),
		quizScores:     make(map[int]bool),
		levelFilter:    LevelExpert, // Show all by default
	}
}

// SetLevelFilter filters steps by difficulty level.
func (t *TutorialController) SetLevelFilter(level TutorialLevel) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.levelFilter = level
}

// SetFastMode enables/disables fast mode (skips optional steps).
func (t *TutorialController) SetFastMode(fast bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fastMode = fast
}

// FastMode returns whether fast mode is enabled.
func (t *TutorialController) FastMode() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.fastMode
}

// SetOnStepStart sets callback for when a step starts.
func (t *TutorialController) SetOnStepStart(fn func(step int, total int, ts TutorialStep)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onStepStart = fn
}

// SetOnStepEnd sets callback for when a step ends.
func (t *TutorialController) SetOnStepEnd(fn func(step int, total int, ts TutorialStep)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onStepEnd = fn
}

// SetOnComplete sets callback for tutorial completion.
func (t *TutorialController) SetOnComplete(fn func(stats TutorialStats)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onComplete = fn
}

// SetOnQuizAnswer sets callback for quiz responses.
func (t *TutorialController) SetOnQuizAnswer(fn func(step int, correct bool)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onQuizAnswer = fn
}

// SetOnHighlight sets callback for UI highlighting.
func (t *TutorialController) SetOnHighlight(fn func(elementID string, highlight bool)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onHighlight = fn
}

// SetOnAnnotation sets callback for annotations.
func (t *TutorialController) SetOnAnnotation(fn func(annotations []StepAnnotation, show bool)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onAnnotation = fn
}

// Start begins the tutorial.
func (t *TutorialController) Start() {
	t.mu.Lock()
	if t.running {
		t.mu.Unlock()
		return
	}
	t.running = true
	t.paused = false
	t.currentStep = 0
	t.startTime = time.Now()
	t.stepTimes = make([]time.Duration, len(t.steps))
	t.completedSteps = make(map[int]bool)
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.mu.Unlock()

	go t.run()
}

// Stop halts the tutorial.
func (t *TutorialController) Stop() {
	t.mu.Lock()
	if !t.running {
		t.mu.Unlock()
		return
	}
	t.running = false
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.mu.Unlock()
}

// Pause pauses the tutorial.
func (t *TutorialController) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running {
		t.paused = true
	}
}

// Resume resumes a paused tutorial.
func (t *TutorialController) Resume() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.running {
		t.paused = false
	}
}

// NextStep advances to the next step manually.
func (t *TutorialController) NextStep() {
	t.mu.Lock()
	defer t.mu.Unlock()
	// Signal to advance (implementation depends on step design)
}

// PreviousStep goes back one step.
func (t *TutorialController) PreviousStep() {
	t.mu.Lock()
	if !t.running || t.currentStep <= 0 {
		t.mu.Unlock()
		return
	}
	t.currentStep--
	t.mu.Unlock()
}

// JumpToStep jumps to a specific step.
func (t *TutorialController) JumpToStep(step int) {
	t.mu.Lock()
	if !t.running || step < 0 || step >= len(t.steps) {
		t.mu.Unlock()
		return
	}
	t.currentStep = step
	t.mu.Unlock()
}

// IsRunning returns true if the tutorial is running.
func (t *TutorialController) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.running
}

// IsPaused returns true if the tutorial is paused.
func (t *TutorialController) IsPaused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.paused
}

// CurrentStep returns the current step index.
func (t *TutorialController) CurrentStep() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentStep
}

// TotalSteps returns the total number of steps.
func (t *TutorialController) TotalSteps() int {
	return len(t.steps)
}

// GetProgress returns progress as a fraction (0.0 - 1.0).
func (t *TutorialController) GetProgress() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.steps) == 0 {
		return 0
	}
	return float64(t.currentStep) / float64(len(t.steps))
}

// GetStats returns current tutorial statistics.
func (t *TutorialController) GetStats() TutorialStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := TutorialStats{
		TotalSteps:     len(t.steps),
		CompletedSteps: len(t.completedSteps),
		TotalTime:      time.Since(t.startTime),
	}

	for _, correct := range t.quizScores {
		stats.QuizTotal++
		if correct {
			stats.QuizCorrect++
		}
	}

	if stats.CompletedSteps > 0 {
		stats.AverageStepTime = stats.TotalTime / time.Duration(stats.CompletedSteps)
	}

	return stats
}

// shouldSkipStep determines if a step should be skipped.
func (t *TutorialController) shouldSkipStep(step TutorialStep) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Skip based on difficulty level
	if step.DifficultyLevel > t.levelFilter {
		return true
	}

	// Skip optional steps in fast mode
	if t.fastMode && step.SkippedWhenFast {
		return true
	}

	return false
}

// run executes the tutorial.
func (t *TutorialController) run() {
	defer func() {
		t.mu.Lock()
		t.running = false
		t.mu.Unlock()
	}()

	total := len(t.steps)

	for {
		t.mu.RLock()
		if !t.running {
			t.mu.RUnlock()
			return
		}
		if t.currentStep >= total {
			t.mu.RUnlock()
			break
		}
		step := t.steps[t.currentStep]
		stepIdx := t.currentStep
		ctx := t.ctx
		t.mu.RUnlock()

		// Check if should skip
		if t.shouldSkipStep(step) {
			t.mu.Lock()
			t.currentStep++
			t.mu.Unlock()
			continue
		}

		// Wait while paused
		for t.IsPaused() {
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}

		stepStart := time.Now()

		// Notify step start
		t.mu.RLock()
		onStepStart := t.onStepStart
		onHighlight := t.onHighlight
		onAnnotation := t.onAnnotation
		t.mu.RUnlock()

		if onStepStart != nil {
			onStepStart(stepIdx, total, step)
		}

		// Apply highlighting
		if step.HighlightElement != "" && onHighlight != nil {
			onHighlight(step.HighlightElement, true)
		}

		// Show annotations
		if len(step.Annotations) > 0 && onAnnotation != nil {
			onAnnotation(step.Annotations, true)
		}

		// Execute step action
		if step.Action != nil {
			step.Action()
		}

		// Wait for duration or user interaction
		if step.Duration > 0 {
			if !t.waitOrStop(ctx, step.Duration) {
				return
			}
		}

		// Remove highlighting
		if step.HighlightElement != "" && onHighlight != nil {
			onHighlight(step.HighlightElement, false)
		}

		// Hide annotations
		if len(step.Annotations) > 0 && onAnnotation != nil {
			onAnnotation(step.Annotations, false)
		}

		// Record step completion
		t.mu.Lock()
		t.completedSteps[stepIdx] = true
		t.stepTimes[stepIdx] = time.Since(stepStart)
		t.mu.Unlock()

		// Notify step end
		t.mu.RLock()
		onStepEnd := t.onStepEnd
		t.mu.RUnlock()
		if onStepEnd != nil {
			onStepEnd(stepIdx, total, step)
		}

		// Advance to next step
		t.mu.Lock()
		t.currentStep++
		t.mu.Unlock()
	}

	// Tutorial complete
	t.mu.RLock()
	onComplete := t.onComplete
	t.mu.RUnlock()

	if onComplete != nil {
		onComplete(t.GetStats())
	}
}

// waitOrStop waits for duration or returns false if stopped.
func (t *TutorialController) waitOrStop(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return t.IsRunning()
	}
}

// TutorialProgressWidget displays tutorial progress visually.
type TutorialProgressWidget struct {
	widget.BaseWidget

	controller  *TutorialController
	titleLabel  *widget.Label
	stepLabel   *widget.Label
	progressBar *widget.ProgressBar
	contentText *widget.RichText
	container   *fyne.Container
}

// NewTutorialProgressWidget creates a new progress widget.
func NewTutorialProgressWidget(ctrl *TutorialController) *TutorialProgressWidget {
	w := &TutorialProgressWidget{
		controller:  ctrl,
		titleLabel:  widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		stepLabel:   widget.NewLabel(""),
		progressBar: widget.NewProgressBar(),
		contentText: widget.NewRichText(),
	}

	w.contentText.Wrapping = fyne.TextWrapWord

	w.container = container.NewVBox(
		w.titleLabel,
		w.progressBar,
		w.stepLabel,
		widget.NewSeparator(),
		container.NewScroll(w.contentText),
	)

	w.ExtendBaseWidget(w)
	return w
}

// UpdateFromStep updates the widget from a tutorial step.
func (w *TutorialProgressWidget) UpdateFromStep(step int, total int, ts TutorialStep) {
	fyne.Do(func() {
		w.titleLabel.SetText(ts.Title)
		w.stepLabel.SetText(ts.Explanation)
		w.progressBar.SetValue(float64(step+1) / float64(total))

		// Build rich text segments
		segments := []widget.RichTextSegment{
			&widget.TextSegment{
				Text:  ts.Explanation,
				Style: widget.RichTextStyleParagraph,
			},
		}

		if ts.UserPrompt != "" {
			segments = append(segments,
				&widget.TextSegment{
					Text:  "\n\n💡 " + ts.UserPrompt,
					Style: widget.RichTextStyleStrong,
				},
			)
		}

		w.contentText.Segments = segments
		w.contentText.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (w *TutorialProgressWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

// TutorialControlBar provides play/pause/navigation buttons.
type TutorialControlBar struct {
	widget.BaseWidget

	controller *TutorialController
	playBtn    *widget.Button
	pauseBtn   *widget.Button
	prevBtn    *widget.Button
	nextBtn    *widget.Button
	speedBtn   *widget.Button
	container  *fyne.Container
}

// NewTutorialControlBar creates a tutorial control bar.
func NewTutorialControlBar(ctrl *TutorialController) *TutorialControlBar {
	bar := &TutorialControlBar{
		controller: ctrl,
	}

	bar.playBtn = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if ctrl.IsRunning() {
			ctrl.Resume()
		} else {
			ctrl.Start()
		}
	})

	bar.pauseBtn = widget.NewButtonWithIcon("", theme.MediaPauseIcon(), func() {
		ctrl.Pause()
	})

	bar.prevBtn = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		ctrl.PreviousStep()
	})

	bar.nextBtn = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		ctrl.NextStep()
	})

	bar.speedBtn = widget.NewButton("1x", func() {
		ctrl.SetFastMode(!ctrl.FastMode())
	})

	bar.container = container.NewHBox(
		bar.prevBtn,
		bar.playBtn,
		bar.pauseBtn,
		bar.nextBtn,
		widget.NewSeparator(),
		bar.speedBtn,
	)

	bar.ExtendBaseWidget(bar)
	return bar
}

// CreateRenderer implements fyne.Widget.
func (bar *TutorialControlBar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(bar.container)
}

// AnimationFrame represents a single animation frame.
type AnimationFrame struct {
	Title     string
	Content   string
	Highlight string
	Duration  time.Duration
	DrawFunc  func(canvas fyne.Canvas) // Custom drawing function
}

// AnimationController manages educational animations.
type AnimationController struct {
	frames       []AnimationFrame
	currentFrame int
	running      bool
	loop         bool
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex

	onFrame func(frame int, af AnimationFrame)
}

// NewAnimationController creates a new animation controller.
func NewAnimationController(frames []AnimationFrame) *AnimationController {
	return &AnimationController{
		frames: frames,
		loop:   false,
	}
}

// SetLoop enables or disables animation looping.
func (a *AnimationController) SetLoop(loop bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loop = loop
}

// SetOnFrame sets the frame change callback.
func (a *AnimationController) SetOnFrame(fn func(frame int, af AnimationFrame)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.onFrame = fn
}

// Start begins the animation.
func (a *AnimationController) Start() {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	a.running = true
	a.currentFrame = 0
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.mu.Unlock()

	go a.run()
}

// Stop halts the animation.
func (a *AnimationController) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.running = false
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
}

// IsRunning returns true if animation is running.
func (a *AnimationController) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}

func (a *AnimationController) run() {
	defer func() {
		a.mu.Lock()
		a.running = false
		a.mu.Unlock()
	}()

	for {
		for i, frame := range a.frames {
			a.mu.RLock()
			if !a.running {
				a.mu.RUnlock()
				return
			}
			ctx := a.ctx
			onFrame := a.onFrame
			a.mu.RUnlock()

			a.mu.Lock()
			a.currentFrame = i
			a.mu.Unlock()

			if onFrame != nil {
				onFrame(i, frame)
			}

			// Wait for frame duration
			duration := frame.Duration
			if duration == 0 {
				duration = 500 * time.Millisecond
			}

			timer := time.NewTimer(duration)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}

		// Check loop
		a.mu.RLock()
		shouldLoop := a.loop && a.running
		a.mu.RUnlock()

		if !shouldLoop {
			return
		}
	}
}

// HighlightOverlay provides visual highlighting for UI elements.
type HighlightOverlay struct {
	widget.BaseWidget

	highlights map[string]HighlightConfig
	mu         sync.RWMutex
}

// HighlightConfig defines how an element should be highlighted.
type HighlightConfig struct {
	Position    fyne.Position
	Size        fyne.Size
	Color       string
	BorderWidth float32
	Pulse       bool
	Label       string
}

// NewHighlightOverlay creates a new highlight overlay.
func NewHighlightOverlay() *HighlightOverlay {
	h := &HighlightOverlay{
		highlights: make(map[string]HighlightConfig),
	}
	h.ExtendBaseWidget(h)
	return h
}

// AddHighlight adds a highlight for an element.
func (h *HighlightOverlay) AddHighlight(id string, config HighlightConfig) {
	h.mu.Lock()
	h.highlights[id] = config
	h.mu.Unlock()
	h.Refresh()
}

// RemoveHighlight removes a highlight.
func (h *HighlightOverlay) RemoveHighlight(id string) {
	h.mu.Lock()
	delete(h.highlights, id)
	h.mu.Unlock()
	h.Refresh()
}

// ClearHighlights removes all highlights.
func (h *HighlightOverlay) ClearHighlights() {
	h.mu.Lock()
	h.highlights = make(map[string]HighlightConfig)
	h.mu.Unlock()
	h.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (h *HighlightOverlay) CreateRenderer() fyne.WidgetRenderer {
	return &highlightRenderer{overlay: h}
}

type highlightRenderer struct {
	overlay *HighlightOverlay
	objects []fyne.CanvasObject
}

func (r *highlightRenderer) Layout(size fyne.Size) {
	// Highlights are positioned absolutely
}

func (r *highlightRenderer) MinSize() fyne.Size {
	return fyne.NewSize(0, 0)
}

func (r *highlightRenderer) Refresh() {
	r.overlay.mu.RLock()
	defer r.overlay.mu.RUnlock()

	r.objects = nil

	for _, config := range r.overlay.highlights {
		// Create highlight rectangle
		rect := canvas.NewRectangle(nil)
		rect.StrokeColor = theme.PrimaryColor()
		rect.StrokeWidth = config.BorderWidth
		if config.BorderWidth == 0 {
			rect.StrokeWidth = 2
		}
		rect.Resize(config.Size)
		rect.Move(config.Position)

		r.objects = append(r.objects, rect)

		// Add label if specified
		if config.Label != "" {
			label := canvas.NewText(config.Label, theme.PrimaryColor())
			label.TextStyle.Bold = true
			label.Move(fyne.NewPos(config.Position.X, config.Position.Y-20))
			r.objects = append(r.objects, label)
		}
	}
}

func (r *highlightRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *highlightRenderer) Destroy() {}
