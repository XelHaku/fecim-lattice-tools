// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedtheme "fecim-lattice-tools/shared/theme"
)

// Global notification manager for demos to use
var (
	globalNotificationManager *NotificationManager
	globalNotificationMu      sync.RWMutex
)

// SetGlobalNotificationManager sets the global notification manager
// that can be accessed by demos to show notifications.
func SetGlobalNotificationManager(nm *NotificationManager) {
	globalNotificationMu.Lock()
	defer globalNotificationMu.Unlock()
	globalNotificationManager = nm
}

// GetGlobalNotificationManager returns the global notification manager.
// Returns nil if not set.
func GetGlobalNotificationManager() *NotificationManager {
	globalNotificationMu.RLock()
	defer globalNotificationMu.RUnlock()
	return globalNotificationManager
}

// ShowNotification is a convenience function to show a notification
// using the global notification manager. Does nothing if not set.
func ShowNotification(notifType NotificationType, title, message string) {
	nm := GetGlobalNotificationManager()
	if nm != nil {
		nm.Notify(notifType, title, message)
	}
}

// ShowInfo shows an info notification using the global manager.
func ShowInfo(title, message string) {
	ShowNotification(NotificationInfo, title, message)
}

// ShowSuccess shows a success notification using the global manager.
func ShowSuccess(title, message string) {
	ShowNotification(NotificationSuccess, title, message)
}

// ShowWarning shows a warning notification using the global manager.
func ShowWarning(title, message string) {
	ShowNotification(NotificationWarning, title, message)
}

// ShowError shows an error notification using the global manager.
func ShowError(title, message string) {
	ShowNotification(NotificationError, title, message)
}

// NotificationType represents the type of notification
type NotificationType int

const (
	// NotificationInfo is for informational messages
	NotificationInfo NotificationType = iota
	// NotificationSuccess is for success messages
	NotificationSuccess
	// NotificationWarning is for warning messages
	NotificationWarning
	// NotificationError is for error messages
	NotificationError
)

// String returns the string representation of the notification type
func (t NotificationType) String() string {
	switch t {
	case NotificationInfo:
		return "INFO"
	case NotificationSuccess:
		return "SUCCESS"
	case NotificationWarning:
		return "WARNING"
	case NotificationError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// NotificationEntry represents a single notification
type NotificationEntry struct {
	ID        int64
	Type      NotificationType
	Title     string
	Message   string
	Timestamp time.Time
	Read      bool
}

// DefaultToastDuration is the default duration for toast notifications
const DefaultToastDuration = 4 * time.Second

// NotificationManager manages notifications and toasts
type NotificationManager struct {
	mu            sync.RWMutex
	history       []NotificationEntry
	maxHistory    int
	nextID        int64
	toastStack    *fyne.Container
	toasts        map[int64]*Toast
	window        fyne.Window
	onChange      func() // Callback when history changes
	onToastAdd    func(*Toast)
	onToastRemove func(*Toast)
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(window fyne.Window, maxHistory int) *NotificationManager {
	if maxHistory <= 0 {
		maxHistory = 100
	}
	nm := &NotificationManager{
		history:    make([]NotificationEntry, 0, maxHistory),
		maxHistory: maxHistory,
		toasts:     make(map[int64]*Toast),
		window:     window,
	}
	return nm
}

// SetOnChange sets the callback for when notification history changes
func (nm *NotificationManager) SetOnChange(fn func()) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.onChange = fn
}

// SetToastCallbacks sets callbacks for toast add/remove events
func (nm *NotificationManager) SetToastCallbacks(onAdd, onRemove func(*Toast)) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.onToastAdd = onAdd
	nm.onToastRemove = onRemove
}

// Notify creates a notification and shows a toast
func (nm *NotificationManager) Notify(notifType NotificationType, title, message string) {
	nm.NotifyWithDuration(notifType, title, message, DefaultToastDuration)
}

// NotifyWithDuration creates a notification with custom toast duration
func (nm *NotificationManager) NotifyWithDuration(notifType NotificationType, title, message string, duration time.Duration) {
	nm.mu.Lock()
	id := nm.nextID
	nm.nextID++

	entry := NotificationEntry{
		ID:        id,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
		Read:      false,
	}

	// Add to history, removing oldest if at capacity
	nm.history = append(nm.history, entry)
	if len(nm.history) > nm.maxHistory {
		nm.history = nm.history[1:]
	}

	onChange := nm.onChange
	onToastAdd := nm.onToastAdd
	nm.mu.Unlock()

	if onChange != nil {
		onChange()
	}

	// Create and show toast
	toast := NewToast(entry, duration, func(t *Toast) {
		nm.removeToast(t.entry.ID)
	})

	nm.mu.Lock()
	nm.toasts[id] = toast
	nm.mu.Unlock()

	if onToastAdd != nil {
		fyne.Do(func() {
			onToastAdd(toast)
		})
	}
}

// removeToast removes a toast from tracking
func (nm *NotificationManager) removeToast(id int64) {
	nm.mu.Lock()
	toast, exists := nm.toasts[id]
	if exists {
		delete(nm.toasts, id)
	}
	onRemove := nm.onToastRemove
	nm.mu.Unlock()

	if exists && onRemove != nil {
		fyne.Do(func() {
			onRemove(toast)
		})
	}
}

// Info shows an info notification
func (nm *NotificationManager) Info(title, message string) {
	nm.Notify(NotificationInfo, title, message)
}

// Success shows a success notification
func (nm *NotificationManager) Success(title, message string) {
	nm.Notify(NotificationSuccess, title, message)
}

// Warning shows a warning notification
func (nm *NotificationManager) Warning(title, message string) {
	nm.Notify(NotificationWarning, title, message)
}

// Error shows an error notification
func (nm *NotificationManager) Error(title, message string) {
	nm.Notify(NotificationError, title, message)
}

// Infof shows an info notification with formatted message
func (nm *NotificationManager) Infof(title, format string, args ...interface{}) {
	nm.Info(title, fmt.Sprintf(format, args...))
}

// Successf shows a success notification with formatted message
func (nm *NotificationManager) Successf(title, format string, args ...interface{}) {
	nm.Success(title, fmt.Sprintf(format, args...))
}

// Warningf shows a warning notification with formatted message
func (nm *NotificationManager) Warningf(title, format string, args ...interface{}) {
	nm.Warning(title, fmt.Sprintf(format, args...))
}

// Errorf shows an error notification with formatted message
func (nm *NotificationManager) Errorf(title, format string, args ...interface{}) {
	nm.Error(title, fmt.Sprintf(format, args...))
}

// History returns a copy of the notification history (newest first)
func (nm *NotificationManager) History() []NotificationEntry {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Return in reverse order (newest first)
	result := make([]NotificationEntry, len(nm.history))
	for i, entry := range nm.history {
		result[len(nm.history)-1-i] = entry
	}
	return result
}

// UnreadCount returns the number of unread notifications
func (nm *NotificationManager) UnreadCount() int {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	count := 0
	for _, entry := range nm.history {
		if !entry.Read {
			count++
		}
	}
	return count
}

// MarkAllRead marks all notifications as read
func (nm *NotificationManager) MarkAllRead() {
	nm.mu.Lock()
	for i := range nm.history {
		nm.history[i].Read = true
	}
	onChange := nm.onChange
	nm.mu.Unlock()

	if onChange != nil {
		onChange()
	}
}

// ClearHistory clears all notification history
func (nm *NotificationManager) ClearHistory() {
	nm.mu.Lock()
	nm.history = nm.history[:0]
	onChange := nm.onChange
	nm.mu.Unlock()

	if onChange != nil {
		onChange()
	}
}

// =============================================================================
// Toast Widget
// =============================================================================

// Toast represents a single toast notification widget
type Toast struct {
	widget.BaseWidget

	entry     NotificationEntry
	duration  time.Duration
	onDismiss func(*Toast)
	dismissed bool
	mu        sync.Mutex
}

// NewToast creates a new toast notification widget
func NewToast(entry NotificationEntry, duration time.Duration, onDismiss func(*Toast)) *Toast {
	t := &Toast{
		entry:     entry,
		duration:  duration,
		onDismiss: onDismiss,
	}
	t.ExtendBaseWidget(t)

	// Auto-dismiss after duration
	if duration > 0 {
		go func() {
			time.Sleep(duration)
			t.Dismiss()
		}()
	}

	return t
}

// Entry returns the notification entry
func (t *Toast) Entry() NotificationEntry {
	return t.entry
}

// Dismiss dismisses the toast
func (t *Toast) Dismiss() {
	t.mu.Lock()
	if t.dismissed {
		t.mu.Unlock()
		return
	}
	t.dismissed = true
	onDismiss := t.onDismiss
	t.mu.Unlock()

	if onDismiss != nil {
		onDismiss(t)
	}
}

// CreateRenderer creates the renderer for the toast widget
func (t *Toast) CreateRenderer() fyne.WidgetRenderer {
	// Get colors based on notification type
	bgColor, iconRes := t.getTypeStyle()

	// Background
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 8
	bg.StrokeColor = sharedtheme.WithAlpha(sharedtheme.ColorText, 40).(color.Color)
	bg.StrokeWidth = 1

	// Icon
	icon := canvas.NewImageFromResource(iconRes)
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(20, 20))

	// Title
	title := canvas.NewText(t.entry.Title, sharedtheme.ColorText)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 13

	// Message
	message := canvas.NewText(t.entry.Message, sharedtheme.ColorTextDim)
	message.TextSize = 12

	// Timestamp
	timestamp := canvas.NewText(t.entry.Timestamp.Format("15:04:05"), sharedtheme.ColorTextDim)
	timestamp.TextSize = 10

	// Close button (x)
	closeBtn := widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
		t.Dismiss()
	})
	closeBtn.Importance = widget.LowImportance

	return &toastRenderer{
		toast:     t,
		bg:        bg,
		icon:      icon,
		title:     title,
		message:   message,
		timestamp: timestamp,
		closeBtn:  closeBtn,
	}
}

// getTypeStyle returns the background color and icon for the notification type
func (t *Toast) getTypeStyle() (color.Color, fyne.Resource) {
	switch t.entry.Type {
	case NotificationSuccess:
		return sharedtheme.WithAlpha(sharedtheme.ColorSuccess, 220).(color.Color), theme.ConfirmIcon()
	case NotificationWarning:
		return sharedtheme.WithAlpha(sharedtheme.ColorWarning, 220).(color.Color), theme.WarningIcon()
	case NotificationError:
		return sharedtheme.WithAlpha(sharedtheme.ColorError, 220).(color.Color), theme.ErrorIcon()
	default: // Info
		return sharedtheme.WithAlpha(sharedtheme.ColorInfo, 220).(color.Color), theme.InfoIcon()
	}
}

// toastRenderer implements the fyne.WidgetRenderer interface for Toast
type toastRenderer struct {
	toast     *Toast
	bg        *canvas.Rectangle
	icon      *canvas.Image
	title     *canvas.Text
	message   *canvas.Text
	timestamp *canvas.Text
	closeBtn  *widget.Button
}

func (r *toastRenderer) Layout(size fyne.Size) {
	padding := float32(12)
	iconSize := float32(20)
	closeBtnSize := float32(24)

	// Background fills entire size
	r.bg.Resize(size)

	// Icon on the left
	r.icon.Resize(fyne.NewSize(iconSize, iconSize))
	r.icon.Move(fyne.NewPos(padding, padding))

	// Close button on the right
	r.closeBtn.Resize(fyne.NewSize(closeBtnSize, closeBtnSize))
	r.closeBtn.Move(fyne.NewPos(size.Width-closeBtnSize-padding/2, padding/2))

	// Title next to icon
	titleX := padding + iconSize + 8
	titleWidth := size.Width - titleX - closeBtnSize - padding
	r.title.Move(fyne.NewPos(titleX, padding))
	r.title.Resize(fyne.NewSize(titleWidth, 16))

	// Message below title
	r.message.Move(fyne.NewPos(titleX, padding+18))
	r.message.Resize(fyne.NewSize(titleWidth, 14))

	// Timestamp at bottom right
	r.timestamp.Move(fyne.NewPos(size.Width-60-padding, size.Height-18))
}

func (r *toastRenderer) MinSize() fyne.Size {
	return fyne.NewSize(280, 60)
}

func (r *toastRenderer) Refresh() {
	canvas.Refresh(r.toast)
}

func (r *toastRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.icon, r.title, r.message, r.timestamp, r.closeBtn}
}

func (r *toastRenderer) Destroy() {}

// =============================================================================
// Toast Container - Overlay for displaying toasts
// =============================================================================

// ToastContainer manages the display of toast notifications
type ToastContainer struct {
	widget.BaseWidget

	mu          sync.RWMutex
	toasts      []*Toast
	maxVisible  int
	container   *fyne.Container
	spacingY    float32
}

// NewToastContainer creates a new toast container
func NewToastContainer(maxVisible int) *ToastContainer {
	if maxVisible <= 0 {
		maxVisible = 5
	}
	tc := &ToastContainer{
		toasts:     make([]*Toast, 0),
		maxVisible: maxVisible,
		spacingY:   8,
	}
	tc.container = container.NewVBox()
	tc.ExtendBaseWidget(tc)
	return tc
}

// Add adds a toast to the container
func (tc *ToastContainer) Add(toast *Toast) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.toasts = append(tc.toasts, toast)

	// Limit visible toasts
	if len(tc.toasts) > tc.maxVisible {
		// Dismiss oldest toast
		tc.toasts[0].Dismiss()
	}

	tc.rebuildContainer()
}

// Remove removes a toast from the container
func (tc *ToastContainer) Remove(toast *Toast) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	for i, t := range tc.toasts {
		if t == toast {
			tc.toasts = append(tc.toasts[:i], tc.toasts[i+1:]...)
			break
		}
	}

	tc.rebuildContainer()
}

// rebuildContainer rebuilds the container with current toasts
func (tc *ToastContainer) rebuildContainer() {
	tc.container.RemoveAll()
	for _, toast := range tc.toasts {
		tc.container.Add(toast)
	}
	tc.container.Refresh()
	tc.Refresh()
}

// CreateRenderer creates the renderer for the toast container
func (tc *ToastContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(tc.container)
}

// =============================================================================
// Notification History Panel
// =============================================================================

// NotificationHistoryPanel displays the notification history
type NotificationHistoryPanel struct {
	widget.BaseWidget

	manager   *NotificationManager
	list      *widget.List
	container *fyne.Container
	visible   bool
}

// NewNotificationHistoryPanel creates a new notification history panel
func NewNotificationHistoryPanel(manager *NotificationManager) *NotificationHistoryPanel {
	nhp := &NotificationHistoryPanel{
		manager: manager,
	}
	nhp.ExtendBaseWidget(nhp)

	// Create header with clear button
	headerLabel := widget.NewLabel("Notification History")
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}

	markReadBtn := widget.NewButtonWithIcon("Mark All Read", theme.MailComposeIcon(), func() {
		manager.MarkAllRead()
	})
	markReadBtn.Importance = widget.LowImportance

	clearBtn := widget.NewButtonWithIcon("Clear", theme.DeleteIcon(), func() {
		manager.ClearHistory()
	})
	clearBtn.Importance = widget.LowImportance

	header := container.NewBorder(nil, nil, headerLabel, container.NewHBox(markReadBtn, clearBtn))

	// Create list
	nhp.list = widget.NewList(
		func() int {
			return len(manager.History())
		},
		func() fyne.CanvasObject {
			return nhp.createListItem()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			nhp.updateListItem(id, obj)
		},
	)

	// Update list when history changes
	manager.SetOnChange(func() {
		fyne.Do(func() {
			nhp.list.Refresh()
		})
	})

	nhp.container = container.NewBorder(header, nil, nil, nil, nhp.list)

	return nhp
}

// createListItem creates a template list item
func (nhp *NotificationHistoryPanel) createListItem() fyne.CanvasObject {
	icon := canvas.NewImageFromResource(theme.InfoIcon())
	icon.SetMinSize(fyne.NewSize(16, 16))

	title := widget.NewLabel("Title")
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Truncation = fyne.TextTruncateEllipsis

	message := widget.NewLabel("Message")
	message.Truncation = fyne.TextTruncateEllipsis

	timestamp := widget.NewLabel("00:00:00")
	timestamp.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewBorder(
		nil, nil,
		icon,
		timestamp,
		container.NewVBox(title, message),
	)
}

// updateListItem updates a list item with notification data
func (nhp *NotificationHistoryPanel) updateListItem(id widget.ListItemID, obj fyne.CanvasObject) {
	history := nhp.manager.History()
	if id >= len(history) {
		return
	}

	entry := history[id]
	c := obj.(*fyne.Container)

	// Icon is first object in border (left side)
	icon := c.Objects[0].(*canvas.Image)
	_, iconRes := nhp.getTypeStyle(entry.Type)
	icon.Resource = iconRes
	icon.Refresh()

	// Timestamp is second object in border (right side)
	timestamp := c.Objects[1].(*widget.Label)
	timestamp.SetText(entry.Timestamp.Format("15:04:05"))

	// Content is the center VBox
	content := c.Objects[2].(*fyne.Container)
	title := content.Objects[0].(*widget.Label)
	message := content.Objects[1].(*widget.Label)

	titleText := entry.Title
	if !entry.Read {
		titleText = "● " + titleText // Unread indicator
	}
	title.SetText(titleText)
	message.SetText(entry.Message)
}

// getTypeStyle returns color and icon for notification type
func (nhp *NotificationHistoryPanel) getTypeStyle(notifType NotificationType) (color.Color, fyne.Resource) {
	switch notifType {
	case NotificationSuccess:
		return sharedtheme.ColorSuccess, theme.ConfirmIcon()
	case NotificationWarning:
		return sharedtheme.ColorWarning, theme.WarningIcon()
	case NotificationError:
		return sharedtheme.ColorError, theme.ErrorIcon()
	default:
		return sharedtheme.ColorInfo, theme.InfoIcon()
	}
}

// Toggle toggles visibility of the panel
func (nhp *NotificationHistoryPanel) Toggle() {
	nhp.visible = !nhp.visible
	if nhp.visible {
		nhp.Show()
	} else {
		nhp.Hide()
	}
}

// SetVisible sets the visibility of the panel
func (nhp *NotificationHistoryPanel) SetVisible(visible bool) {
	nhp.visible = visible
	if visible {
		nhp.Show()
	} else {
		nhp.Hide()
	}
}

// IsVisible returns whether the panel is visible
func (nhp *NotificationHistoryPanel) IsVisible() bool {
	return nhp.visible
}

// CreateRenderer creates the renderer for the history panel
func (nhp *NotificationHistoryPanel) CreateRenderer() fyne.WidgetRenderer {
	// Background
	bg := canvas.NewRectangle(sharedtheme.ColorSurface)
	bg.CornerRadius = 4
	bg.StrokeColor = sharedtheme.ColorSeparator
	bg.StrokeWidth = 1

	return &historyPanelRenderer{
		panel:     nhp,
		bg:        bg,
		container: nhp.container,
	}
}

type historyPanelRenderer struct {
	panel     *NotificationHistoryPanel
	bg        *canvas.Rectangle
	container *fyne.Container
}

func (r *historyPanelRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.container.Resize(size)
}

func (r *historyPanelRenderer) MinSize() fyne.Size {
	return fyne.NewSize(300, 200)
}

func (r *historyPanelRenderer) Refresh() {
	r.container.Refresh()
}

func (r *historyPanelRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.bg, r.container}
}

func (r *historyPanelRenderer) Destroy() {}

// =============================================================================
// Notification Button - Shows unread count badge
// =============================================================================

// NotificationButton is a button that shows the notification history panel
type NotificationButton struct {
	widget.BaseWidget

	manager    *NotificationManager
	badge      *canvas.Text
	badgeBg    *canvas.Circle
	iconWidget *widget.Button
	onTap      func()
}

// NewNotificationButton creates a new notification button
func NewNotificationButton(manager *NotificationManager, onTap func()) *NotificationButton {
	nb := &NotificationButton{
		manager: manager,
		onTap:   onTap,
	}
	nb.ExtendBaseWidget(nb)

	nb.badge = canvas.NewText("0", sharedtheme.ColorText)
	nb.badge.TextSize = 10
	nb.badge.TextStyle = fyne.TextStyle{Bold: true}

	nb.badgeBg = canvas.NewCircle(sharedtheme.ColorError)

	nb.iconWidget = widget.NewButtonWithIcon("", theme.MailComposeIcon(), func() {
		if nb.onTap != nil {
			nb.onTap()
		}
	})

	// Update badge when history changes
	manager.SetOnChange(func() {
		fyne.Do(func() {
			nb.updateBadge()
		})
	})

	return nb
}

// updateBadge updates the badge count
func (nb *NotificationButton) updateBadge() {
	count := nb.manager.UnreadCount()
	if count > 0 {
		if count > 99 {
			nb.badge.Text = "99+"
		} else {
			nb.badge.Text = fmt.Sprintf("%d", count)
		}
		nb.badge.Show()
		nb.badgeBg.Show()
	} else {
		nb.badge.Hide()
		nb.badgeBg.Hide()
	}
	nb.Refresh()
}

// CreateRenderer creates the renderer for the notification button
func (nb *NotificationButton) CreateRenderer() fyne.WidgetRenderer {
	nb.updateBadge() // Initial update
	return &notificationButtonRenderer{
		button:  nb,
		icon:    nb.iconWidget,
		badge:   nb.badge,
		badgeBg: nb.badgeBg,
	}
}

type notificationButtonRenderer struct {
	button  *NotificationButton
	icon    *widget.Button
	badge   *canvas.Text
	badgeBg *canvas.Circle
}

func (r *notificationButtonRenderer) Layout(size fyne.Size) {
	r.icon.Resize(size)

	// Badge in top-right corner
	badgeSize := float32(14)
	r.badgeBg.Resize(fyne.NewSize(badgeSize, badgeSize))
	r.badgeBg.Move(fyne.NewPos(size.Width-badgeSize+2, -2))

	r.badge.Move(fyne.NewPos(size.Width-badgeSize+5, -1))
}

func (r *notificationButtonRenderer) MinSize() fyne.Size {
	return r.icon.MinSize()
}

func (r *notificationButtonRenderer) Refresh() {
	r.icon.Refresh()
	r.badge.Refresh()
	r.badgeBg.Refresh()
}

func (r *notificationButtonRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.icon, r.badgeBg, r.badge}
}

func (r *notificationButtonRenderer) Destroy() {}
