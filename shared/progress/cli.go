// Package progress provides CLI progress bar for terminal output.
package progress

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// CLIProgress provides a terminal-based progress bar with ETA and status.
type CLIProgress struct {
	progress *Progress
	writer   io.Writer
	width    int
	ticker   *time.Ticker
	done     chan struct{}
	mu       sync.Mutex
	lastLine string
	showRate bool
	showETA  bool
	prefix   string
}

// CLIOption configures a CLIProgress
type CLIOption func(*CLIProgress)

// WithWriter sets the output writer (default: os.Stderr)
func WithWriter(w io.Writer) CLIOption {
	return func(c *CLIProgress) {
		c.writer = w
	}
}

// WithWidth sets the progress bar width (default: 40)
func WithWidth(width int) CLIOption {
	return func(c *CLIProgress) {
		c.width = width
	}
}

// WithShowRate enables/disables rate display
func WithShowRate(show bool) CLIOption {
	return func(c *CLIProgress) {
		c.showRate = show
	}
}

// WithShowETA enables/disables ETA display
func WithShowETA(show bool) CLIOption {
	return func(c *CLIProgress) {
		c.showETA = show
	}
}

// WithPrefix sets a prefix string for the progress line
func WithPrefix(prefix string) CLIOption {
	return func(c *CLIProgress) {
		c.prefix = prefix
	}
}

// NewCLIProgress creates a new CLI progress bar
func NewCLIProgress(p *Progress, opts ...CLIOption) *CLIProgress {
	c := &CLIProgress{
		progress: p,
		writer:   os.Stderr,
		width:    40,
		showRate: true,
		showETA:  true,
		done:     make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Start begins rendering the progress bar
func (c *CLIProgress) Start() {
	c.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-c.done:
				c.ticker.Stop()
				return
			case <-c.ticker.C:
				c.render()
			}
		}
	}()
}

// Stop stops the progress bar and prints a final line
func (c *CLIProgress) Stop() {
	close(c.done)
	c.renderFinal()
}

// render draws the current progress state
func (c *CLIProgress) render() {
	c.mu.Lock()
	defer c.mu.Unlock()

	info := c.progress.Info()
	line := c.buildLine(info)

	// Clear previous line and write new one
	if c.lastLine != "" {
		clearLen := len(c.lastLine)
		fmt.Fprint(c.writer, "\r"+strings.Repeat(" ", clearLen)+"\r")
	}
	fmt.Fprint(c.writer, line)
	c.lastLine = line
}

// renderFinal draws the final state with newline
func (c *CLIProgress) renderFinal() {
	c.mu.Lock()
	defer c.mu.Unlock()

	info := c.progress.Info()

	// Clear previous line
	if c.lastLine != "" {
		clearLen := len(c.lastLine)
		fmt.Fprint(c.writer, "\r"+strings.Repeat(" ", clearLen)+"\r")
	}

	// Build final line with status
	var status string
	switch info.State {
	case StateCompleted:
		status = "✓ Complete"
	case StateCancelled:
		status = "✗ Cancelled"
	case StateFailed:
		status = fmt.Sprintf("✗ Failed: %v", info.Error)
	default:
		status = info.State.String()
	}

	line := fmt.Sprintf("%s %s (%s)\n", c.prefix, status, formatDuration(info.Elapsed))
	fmt.Fprint(c.writer, line)
}

// buildLine constructs the progress bar line
func (c *CLIProgress) buildLine(info ProgressInfo) string {
	var sb strings.Builder

	// Prefix
	if c.prefix != "" {
		sb.WriteString(c.prefix)
		sb.WriteString(" ")
	}

	// State indicator
	switch info.State {
	case StateRunning:
		sb.WriteString("⠋ ") // Spinner char (could animate)
	case StatePaused:
		sb.WriteString("⏸ ")
	case StateCancelled:
		sb.WriteString("✗ ")
	case StateCompleted:
		sb.WriteString("✓ ")
	case StateFailed:
		sb.WriteString("✗ ")
	default:
		sb.WriteString("  ")
	}

	// Phase (truncated)
	phase := info.Phase
	if len(phase) > 20 {
		phase = phase[:17] + "..."
	}
	if phase != "" {
		sb.WriteString(phase)
		sb.WriteString(" ")
	}

	// Progress bar
	if info.Total > 0 {
		sb.WriteString(c.buildBar(info))
	} else {
		// Indeterminate spinner
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		idx := int(time.Now().UnixMilli()/100) % len(spinner)
		sb.WriteString(spinner[idx])
		sb.WriteString(" ")
	}

	// Percentage
	if info.Total > 0 {
		sb.WriteString(fmt.Sprintf(" %5.1f%%", info.Percent))
	}

	// Rate
	if c.showRate && info.Rate > 0 {
		sb.WriteString(fmt.Sprintf(" [%s]", formatRate(info.Rate, info.Total)))
	}

	// ETA
	if c.showETA && info.ETA > 0 {
		sb.WriteString(fmt.Sprintf(" ETA: %s", formatDuration(info.ETA)))
	}

	// Detail (truncated)
	if info.Detail != "" {
		detail := info.Detail
		if len(detail) > 30 {
			detail = detail[:27] + "..."
		}
		sb.WriteString(" - ")
		sb.WriteString(detail)
	}

	return sb.String()
}

// buildBar constructs the progress bar graphics
func (c *CLIProgress) buildBar(info ProgressInfo) string {
	width := c.width
	filled := int(info.Percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < width; i++ {
		if i < filled {
			sb.WriteString("█")
		} else if i == filled {
			sb.WriteString("▓")
		} else {
			sb.WriteString("░")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// MultiCLIProgress manages multiple concurrent progress bars
type MultiCLIProgress struct {
	progresses []*CLIProgress
	writer     io.Writer
	ticker     *time.Ticker
	done       chan struct{}
	mu         sync.Mutex
	lineCount  int
}

// NewMultiCLIProgress creates a manager for multiple progress bars
func NewMultiCLIProgress(writer io.Writer) *MultiCLIProgress {
	if writer == nil {
		writer = os.Stderr
	}
	return &MultiCLIProgress{
		writer: writer,
		done:   make(chan struct{}),
	}
}

// Add adds a progress to track
func (m *MultiCLIProgress) Add(p *Progress, opts ...CLIOption) *CLIProgress {
	m.mu.Lock()
	defer m.mu.Unlock()

	opts = append(opts, WithWriter(m.writer))
	cp := NewCLIProgress(p, opts...)
	m.progresses = append(m.progresses, cp)
	return cp
}

// Start begins rendering all progress bars
func (m *MultiCLIProgress) Start() {
	m.ticker = time.NewTicker(100 * time.Millisecond)

	go func() {
		for {
			select {
			case <-m.done:
				m.ticker.Stop()
				return
			case <-m.ticker.C:
				m.render()
			}
		}
	}()
}

// Stop stops all progress bars
func (m *MultiCLIProgress) Stop() {
	close(m.done)
	m.renderFinal()
}

// render draws all progress bars
func (m *MultiCLIProgress) render() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Move cursor up to first line
	if m.lineCount > 0 {
		fmt.Fprintf(m.writer, "\033[%dA", m.lineCount)
	}

	m.lineCount = 0
	for _, cp := range m.progresses {
		info := cp.progress.Info()
		line := cp.buildLine(info)
		// Clear line and print
		fmt.Fprintf(m.writer, "\033[2K%s\n", line)
		m.lineCount++
	}
}

// renderFinal draws the final state
func (m *MultiCLIProgress) renderFinal() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Move cursor up
	if m.lineCount > 0 {
		fmt.Fprintf(m.writer, "\033[%dA", m.lineCount)
	}

	for _, cp := range m.progresses {
		cp.renderFinal()
	}
}

// SimpleProgress provides a simple one-line progress API for scripts
func SimpleProgress(operation string, total int64, work func(p *Progress) error) error {
	p := NewProgress(operation, total)
	cli := NewCLIProgress(p)

	p.Start()
	cli.Start()

	err := work(p)

	if err != nil {
		p.Fail(err)
	} else {
		p.Complete()
	}

	cli.Stop()
	return err
}

// SpinnerProgress provides a simple spinner for indeterminate operations
func SpinnerProgress(message string, work func(p *Progress) error) error {
	p := NewProgress(message, 0) // 0 = indeterminate
	cli := NewCLIProgress(p)

	p.Start()
	p.SetPhase(message)
	cli.Start()

	err := work(p)

	if err != nil {
		p.Fail(err)
	} else {
		p.Complete()
	}

	cli.Stop()
	return err
}

// IterProgress wraps iteration with progress tracking
type IterProgress struct {
	progress *Progress
	cli      *CLIProgress
	started  bool
}

// NewIterProgress creates an iterator progress tracker
func NewIterProgress(operation string, total int64) *IterProgress {
	p := NewProgress(operation, total)
	return &IterProgress{
		progress: p,
		cli:      NewCLIProgress(p),
	}
}

// Start begins progress tracking
func (ip *IterProgress) Start() *IterProgress {
	ip.progress.Start()
	ip.cli.Start()
	ip.started = true
	return ip
}

// Update updates progress
func (ip *IterProgress) Update(current int64) {
	ip.progress.Update(current)
}

// Increment increments progress by 1
func (ip *IterProgress) Increment() {
	ip.progress.Increment()
}

// SetPhase sets the current phase
func (ip *IterProgress) SetPhase(phase string) {
	ip.progress.SetPhase(phase)
}

// SetDetail sets the detail message
func (ip *IterProgress) SetDetail(detail string) {
	ip.progress.SetDetail(detail)
}

// IsCancelled checks if cancelled
func (ip *IterProgress) IsCancelled() bool {
	return ip.progress.IsCancelled()
}

// Context returns the cancellation context
func (ip *IterProgress) Context() context.Context {
	return ip.progress.Context()
}

// Complete marks as complete
func (ip *IterProgress) Complete() {
	ip.progress.Complete()
	ip.cli.Stop()
}

// Fail marks as failed
func (ip *IterProgress) Fail(err error) {
	ip.progress.Fail(err)
	ip.cli.Stop()
}

// Cancel cancels the operation
func (ip *IterProgress) Cancel() {
	ip.progress.Cancel()
	ip.cli.Stop()
}

// Progress returns the underlying Progress
func (ip *IterProgress) Progress() *Progress {
	return ip.progress
}
