package help

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestNewHelpBrowserAndSearchCategorySelection(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := test.NewWindow(nil)

	hs := GetHelpSystem()
	hb := NewHelpBrowser(hs, window)
	if hb == nil {
		t.Fatal("NewHelpBrowser returned nil")
	}
	if len(hb.categories) == 0 || hb.categories[0] != "All Topics" {
		t.Fatalf("expected categories with All Topics first, got %#v", hb.categories)
	}

	all := hs.GetAllTopics()
	hb.onSearch("")
	if len(hb.topics) != len(all) {
		t.Fatalf("empty search should return all topics: got %d want %d", len(hb.topics), len(all))
	}

	hb.onSearch("hysteresis")
	if len(hb.topics) == 0 {
		t.Fatal("expected non-empty search results for hysteresis")
	}

	// Select first category => all topics.
	hb.searchEntry.SetText("temp")
	hb.onCategorySelected(0)
	if hb.searchEntry.Text != "" {
		t.Fatal("category selection should clear search entry")
	}
	if len(hb.topics) != len(all) {
		t.Fatalf("all-topics category should return all topics: got %d want %d", len(hb.topics), len(all))
	}

	// Out-of-range category id should panic (documents current behavior).
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for out-of-range category selection")
		}
	}()
	hb.onCategorySelected(len(hb.categories))
}

func TestHelpBrowserTopicSelectionAndShowTopic(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := test.NewWindow(nil)

	hs := GetHelpSystem()
	hb := NewHelpBrowser(hs, window)

	hb.onTopicSelected(len(hb.topics)) // out-of-range should be ignored
	if hb.currentTopic != nil {
		t.Fatal("out-of-range topic selection should not set current topic")
	}

	topic := hs.GetTopic("module.crossbar")
	if topic == nil {
		t.Fatal("expected embedded topic module.crossbar")
	}
	hb.ShowTopic(nil)
	hb.ShowTopic(topic)
	if hb.currentTopic == nil || hb.currentTopic.ID != "module.crossbar" {
		t.Fatalf("ShowTopic should set current topic to module.crossbar, got %#v", hb.currentTopic)
	}

	hb.topics = []*HelpTopic{topic}
	hb.onTopicSelected(0)
	if hb.currentTopic == nil || hb.currentTopic.ID != "module.crossbar" {
		t.Fatal("onTopicSelected should set current topic")
	}
}

func TestContextualHelpDialogAndTruncate(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	window := test.NewWindow(nil)

	if got := truncate("abc", 10); got != "abc" {
		t.Fatalf("truncate short string = %q, want %q", got, "abc")
	}
	if got := truncate("abcdefghij", 6); got != "abc..." {
		t.Fatalf("truncate long string = %q, want %q", got, "abc...")
	}

	// nil topic should return early without panic
	NewContextualHelpDialog(nil, window).Show()

	topic := GetHelpSystem().GetTopic("overview")
	if topic == nil {
		t.Fatal("expected overview topic")
	}
	NewContextualHelpDialog(topic, window).Show()
}
