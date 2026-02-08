package widgets

import (
	"fmt"
	"testing"
	"time"
)

func TestNotificationType_String(t *testing.T) {
	tests := []struct {
		notifType NotificationType
		expected  string
	}{
		{NotificationInfo, "INFO"},
		{NotificationSuccess, "SUCCESS"},
		{NotificationWarning, "WARNING"},
		{NotificationError, "ERROR"},
		{NotificationType(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.notifType.String(); got != tt.expected {
				t.Errorf("NotificationType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNotificationManager_Basic(t *testing.T) {
	// Create manager without window (nil is ok for tests)
	nm := NewNotificationManager(nil, 10)

	// Test initial state
	if len(nm.History()) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(nm.History()))
	}
	if nm.UnreadCount() != 0 {
		t.Errorf("Expected 0 unread, got %d", nm.UnreadCount())
	}

	// Add notifications (without toast callbacks set)
	nm.Info("Test", "Info message")
	nm.Success("Test", "Success message")
	nm.Warning("Test", "Warning message")
	nm.Error("Test", "Error message")

	// Check history
	history := nm.History()
	if len(history) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(history))
	}

	// History should be newest first
	if history[0].Type != NotificationError {
		t.Errorf("Expected newest (Error) first, got %v", history[0].Type)
	}
	if history[3].Type != NotificationInfo {
		t.Errorf("Expected oldest (Info) last, got %v", history[3].Type)
	}

	// Check unread count
	if nm.UnreadCount() != 4 {
		t.Errorf("Expected 4 unread, got %d", nm.UnreadCount())
	}
}

func TestNotificationManager_MaxHistory(t *testing.T) {
	nm := NewNotificationManager(nil, 3)

	// Add more than max
	nm.Info("1", "First")
	nm.Info("2", "Second")
	nm.Info("3", "Third")
	nm.Info("4", "Fourth")
	nm.Info("5", "Fifth")

	// Should only have max entries
	history := nm.History()
	if len(history) != 3 {
		t.Errorf("Expected 3 entries (max), got %d", len(history))
	}

	// Should have newest entries
	if history[0].Title != "5" || history[2].Title != "3" {
		t.Errorf("Expected entries 5,4,3, got %s,%s,%s",
			history[0].Title, history[1].Title, history[2].Title)
	}
}

func TestNotificationManager_MarkAllRead(t *testing.T) {
	nm := NewNotificationManager(nil, 10)

	nm.Info("1", "First")
	nm.Info("2", "Second")
	nm.Info("3", "Third")

	if nm.UnreadCount() != 3 {
		t.Errorf("Expected 3 unread, got %d", nm.UnreadCount())
	}

	nm.MarkAllRead()

	if nm.UnreadCount() != 0 {
		t.Errorf("Expected 0 unread after mark all, got %d", nm.UnreadCount())
	}
}

func TestNotificationManager_ClearHistory(t *testing.T) {
	nm := NewNotificationManager(nil, 10)

	nm.Info("1", "First")
	nm.Info("2", "Second")

	nm.ClearHistory()

	if len(nm.History()) != 0 {
		t.Errorf("Expected empty history after clear, got %d", len(nm.History()))
	}
}

func TestNotificationManager_OnChange(t *testing.T) {
	nm := NewNotificationManager(nil, 10)

	callCount := 0
	nm.SetOnChange(func() {
		callCount++
	})

	nm.Info("1", "First")
	nm.MarkAllRead()
	nm.ClearHistory()

	if callCount != 3 {
		t.Errorf("Expected 3 onChange calls, got %d", callCount)
	}
}

func TestNotificationManager_FormattedMessages(t *testing.T) {
	nm := NewNotificationManager(nil, 10)

	nm.Infof("Info", "Value: %d", 42)
	nm.Successf("Success", "File: %s", "test.txt")
	nm.Warningf("Warning", "Temp: %.1f°C", 75.5)
	nm.Errorf("Error", "Code: %d, Msg: %s", 500, "Internal Error")

	history := nm.History()
	if len(history) != 4 {
		t.Fatalf("Expected 4 entries, got %d", len(history))
	}

	if history[3].Message != "Value: 42" {
		t.Errorf("Expected 'Value: 42', got '%s'", history[3].Message)
	}
	if history[0].Message != "Code: 500, Msg: Internal Error" {
		t.Errorf("Expected formatted error, got '%s'", history[0].Message)
	}
}

func TestNotificationEntry_Fields(t *testing.T) {
	nm := NewNotificationManager(nil, 10)

	before := time.Now()
	nm.Info("Test Title", "Test Message")
	after := time.Now()

	history := nm.History()
	if len(history) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(history))
	}

	entry := history[0]
	if entry.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", entry.Title)
	}
	if entry.Message != "Test Message" {
		t.Errorf("Expected message 'Test Message', got '%s'", entry.Message)
	}
	if entry.Type != NotificationInfo {
		t.Errorf("Expected type Info, got %v", entry.Type)
	}
	if entry.Read {
		t.Error("Expected unread entry")
	}
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Errorf("Timestamp out of range: %v not between %v and %v",
			entry.Timestamp, before, after)
	}
	if entry.ID < 0 {
		t.Error("Expected valid ID")
	}
}

func TestToast_Dismiss(t *testing.T) {
	entry := NotificationEntry{
		ID:      1,
		Type:    NotificationInfo,
		Title:   "Test",
		Message: "Test message",
	}

	dismissed := false
	toast := NewToast(entry, 0, func(toast *Toast) {
		dismissed = true
	})

	toast.Dismiss()

	if !dismissed {
		t.Error("Expected onDismiss to be called")
	}

	// Double dismiss should be safe
	toast.Dismiss()
}

func TestToast_Entry(t *testing.T) {
	entry := NotificationEntry{
		ID:      42,
		Type:    NotificationSuccess,
		Title:   "Success",
		Message: "Operation completed",
	}

	toast := NewToast(entry, 0, nil)

	got := toast.Entry()
	if got.ID != entry.ID || got.Type != entry.Type ||
		got.Title != entry.Title || got.Message != entry.Message {
		t.Error("Entry() returned different data than provided")
	}
}

func TestToastContainer_AddRemove(t *testing.T) {
	// Note: Full widget functionality requires a running Fyne app.
	// This test validates the container's internal state management only.
	tc := &ToastContainer{
		toasts:     make([]*Toast, 0),
		maxVisible: 3,
		spacingY:   8,
	}

	entry1 := NotificationEntry{ID: 1, Title: "Toast 1"}
	entry2 := NotificationEntry{ID: 2, Title: "Toast 2"}

	toast1 := &Toast{entry: entry1}
	toast2 := &Toast{entry: entry2}

	// Manually add to toasts slice (bypassing widget operations)
	tc.mu.Lock()
	tc.toasts = append(tc.toasts, toast1)
	tc.toasts = append(tc.toasts, toast2)
	tc.mu.Unlock()

	tc.mu.RLock()
	count := len(tc.toasts)
	tc.mu.RUnlock()

	if count != 2 {
		t.Errorf("Expected 2 toasts, got %d", count)
	}

	// Manually remove from toasts slice
	tc.mu.Lock()
	tc.toasts = tc.toasts[1:]
	tc.mu.Unlock()

	tc.mu.RLock()
	count = len(tc.toasts)
	tc.mu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 toast after remove, got %d", count)
	}
}

func TestToastContainer_MaxVisible(t *testing.T) {
	// Test the max visible logic directly without widget operations
	tc := &ToastContainer{
		toasts:     make([]*Toast, 0),
		maxVisible: 2,
		spacingY:   8,
	}

	// Simulate adding toasts and respecting maxVisible
	for i := 1; i <= 3; i++ {
		entry := NotificationEntry{ID: int64(i), Title: fmt.Sprintf("Toast %d", i)}
		toast := &Toast{entry: entry}

		tc.mu.Lock()
		tc.toasts = append(tc.toasts, toast)
		// Check if we need to remove oldest
		if len(tc.toasts) > tc.maxVisible {
			tc.toasts = tc.toasts[1:]
		}
		tc.mu.Unlock()
	}

	tc.mu.RLock()
	count := len(tc.toasts)
	tc.mu.RUnlock()

	if count != 2 {
		t.Errorf("Expected 2 toasts (maxVisible), got %d", count)
	}

	// Verify we have the newest ones
	tc.mu.RLock()
	if tc.toasts[0].entry.Title != "Toast 2" || tc.toasts[1].entry.Title != "Toast 3" {
		t.Errorf("Expected Toast 2 and Toast 3, got %s and %s",
			tc.toasts[0].entry.Title, tc.toasts[1].entry.Title)
	}
	tc.mu.RUnlock()
}

func TestNotificationHistoryPanel_Creation(t *testing.T) {
	nm := NewNotificationManager(nil, 10)
	nm.Info("Test", "Test message")

	panel := NewNotificationHistoryPanel(nm)

	// Basic creation should work
	if panel == nil {
		t.Fatal("Expected non-nil panel")
	}

	// Visibility toggle
	if panel.IsVisible() {
		t.Error("Expected panel to be initially hidden")
	}

	panel.SetVisible(true)
	if !panel.IsVisible() {
		t.Error("Expected panel to be visible after SetVisible(true)")
	}

	panel.Toggle()
	if panel.IsVisible() {
		t.Error("Expected panel to be hidden after toggle")
	}
}
