// Package widgets provides shared widget utilities for Fyne GUI development.
// This file implements a reusable OperationLog widget for displaying
// timestamped operation history in demo applications.
package widgets

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// OperationLog is a reusable widget for displaying timestamped operation history.
// It maintains a scrolling log of recent operations with configurable capacity.
type OperationLog struct {
	widget.BaseWidget

	mu           sync.RWMutex
	title        string
	entries      []string
	maxEntries   int
	minSize      fyne.Size
	startTime    time.Time
	useMonospace bool
	emptyText    string
	titleLabel   *widget.Label
	contentLabel *widget.Label
}

// OperationLogConfig holds configuration for creating an OperationLog.
type OperationLogConfig struct {
	Title        string
	MaxEntries   int
	MinSize      fyne.Size
	UseMonospace bool
	EmptyText    string
}

// NewOperationLog creates a new operation log widget.
func NewOperationLog(config OperationLogConfig) *OperationLog {
	if config.Title == "" {
		config.Title = "Operation Log"
	}
	if config.MaxEntries <= 0 {
		config.MaxEntries = 10
	}
	if config.MinSize.Width <= 0 {
		config.MinSize.Width = 150
	}
	if config.MinSize.Height <= 0 {
		config.MinSize.Height = 120
	}
	if config.EmptyText == "" {
		config.EmptyText = "Waiting for operations..."
	}

	o := &OperationLog{
		title:        config.Title,
		maxEntries:   config.MaxEntries,
		minSize:      config.MinSize,
		startTime:    time.Now(),
		entries:      make([]string, 0, config.MaxEntries),
		useMonospace: config.UseMonospace,
		emptyText:    config.EmptyText,
	}
	o.ExtendBaseWidget(o)
	return o
}

// Add adds a new log entry with a timestamp.
func (o *OperationLog) Add(entry string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	elapsed := time.Since(o.startTime)
	timestamp := fmt.Sprintf("[%5.1fs]", elapsed.Seconds())
	fullEntry := timestamp + " " + entry

	o.entries = append(o.entries, fullEntry)
	if len(o.entries) > o.maxEntries {
		o.entries = o.entries[1:]
	}

	o.updateContent()
}

// AddWithPrefix adds a log entry with a custom prefix instead of timestamp.
func (o *OperationLog) AddWithPrefix(prefix, entry string) {
	o.mu.Lock()
	defer o.mu.Unlock()

	fullEntry := prefix + " " + entry
	o.entries = append(o.entries, fullEntry)
	if len(o.entries) > o.maxEntries {
		o.entries = o.entries[1:]
	}

	o.updateContent()
}

// Clear clears all log entries and resets the start time.
func (o *OperationLog) Clear() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.entries = o.entries[:0]
	o.startTime = time.Now()
	o.updateContent()
}

// GetEntries returns a copy of all current entries.
func (o *OperationLog) GetEntries() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	result := make([]string, len(o.entries))
	copy(result, o.entries)
	return result
}

// SetTitle updates the log title.
func (o *OperationLog) SetTitle(title string) {
	o.mu.Lock()
	o.title = title
	o.mu.Unlock()
	if o.titleLabel != nil {
		fyne.Do(func() {
			o.titleLabel.SetText(title)
		})
	}
}

func (o *OperationLog) updateContent() {
	if o.contentLabel == nil {
		return
	}

	text := o.emptyText
	if len(o.entries) > 0 {
		text = ""
		for _, entry := range o.entries {
			if text != "" {
				text += "\n"
			}
			text += entry
		}
	}

	fyne.Do(func() {
		o.contentLabel.SetText(text)
	})
}

// MinSize returns the minimum size for the widget.
func (o *OperationLog) MinSize() fyne.Size {
	return o.minSize
}

// CreateRenderer implements fyne.Widget.
func (o *OperationLog) CreateRenderer() fyne.WidgetRenderer {
	o.mu.RLock()
	title := o.title
	o.mu.RUnlock()

	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	contentLabel := widget.NewLabel(o.emptyText)
	if o.useMonospace {
		contentLabel.TextStyle = fyne.TextStyle{Monospace: true}
	}
	contentLabel.Wrapping = fyne.TextWrapWord

	contentScroll := container.NewScroll(contentLabel)
	contentScroll.SetMinSize(fyne.NewSize(o.minSize.Width-20, o.minSize.Height-40))

	box := container.NewBorder(
		container.NewVBox(titleLabel, widget.NewSeparator()),
		nil, nil, nil,
		container.NewPadded(contentScroll),
	)

	o.mu.Lock()
	o.titleLabel = titleLabel
	o.contentLabel = contentLabel
	o.mu.Unlock()

	// Update content to show any existing entries
	o.updateContent()

	return widget.NewSimpleRenderer(box)
}

// FormattedLogEntry creates a formatted log entry with a result type.
type FormattedLogEntry struct {
	Type    string // e.g., "SUCCESS", "ERROR", "INFO"
	Message string
}

// AddFormatted adds a formatted log entry with type prefix.
func (o *OperationLog) AddFormatted(entry FormattedLogEntry) {
	prefix := fmt.Sprintf("[%s]", entry.Type)
	o.AddWithPrefix(prefix, entry.Message)
}

// AddSuccess adds a success-type log entry.
func (o *OperationLog) AddSuccess(message string) {
	o.AddFormatted(FormattedLogEntry{Type: "✓", Message: message})
}

// AddError adds an error-type log entry.
func (o *OperationLog) AddError(message string) {
	o.AddFormatted(FormattedLogEntry{Type: "✗", Message: message})
}

// AddInfo adds an info-type log entry.
func (o *OperationLog) AddInfo(message string) {
	o.AddFormatted(FormattedLogEntry{Type: "ℹ", Message: message})
}
