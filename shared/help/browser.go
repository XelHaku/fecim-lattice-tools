// Package help provides the help browser dialog.
package help

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// HelpBrowser provides a searchable help dialog.
type HelpBrowser struct {
	helpSystem   *HelpSystem
	window       fyne.Window
	dialog       dialog.Dialog
	searchEntry  *widget.Entry
	categoryList *widget.List
	topicList    *widget.List
	contentText  *widget.RichText
	categories   []string
	topics       []*HelpTopic
	currentTopic *HelpTopic
}

// NewHelpBrowser creates a new help browser.
func NewHelpBrowser(hs *HelpSystem, window fyne.Window) *HelpBrowser {
	hb := &HelpBrowser{
		helpSystem: hs,
		window:     window,
		categories: []string{"All Topics"},
		topics:     hs.GetAllTopics(),
	}

	// Build category list
	catMap := hs.GetTopicsByCategory()
	for cat := range catMap {
		hb.categories = append(hb.categories, cat)
	}

	hb.createUI()
	return hb
}

func (hb *HelpBrowser) createUI() {
	// Search entry
	hb.searchEntry = widget.NewEntry()
	hb.searchEntry.SetPlaceHolder("Search help topics...")
	hb.searchEntry.OnChanged = hb.onSearch

	// Category list
	hb.categoryList = widget.NewList(
		func() int { return len(hb.categories) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Category")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(hb.categories[id])
		},
	)
	hb.categoryList.OnSelected = hb.onCategorySelected

	// Topic list
	hb.topicList = widget.NewList(
		func() int { return len(hb.topics) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabelWithStyle("Topic Title", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Summary text..."),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(hb.topics) {
				return
			}
			topic := hb.topics[id]
			box := obj.(*fyne.Container)
			title := box.Objects[0].(*widget.Label)
			summary := box.Objects[1].(*widget.Label)
			title.SetText(topic.Title)
			summary.SetText(truncate(topic.Summary, 60))
		},
	)
	hb.topicList.OnSelected = hb.onTopicSelected

	// Content display
	hb.contentText = widget.NewRichTextFromMarkdown("# Help\n\nSelect a topic from the list.")
	hb.contentText.Wrapping = fyne.TextWrapWord
	contentScroll := container.NewVScroll(hb.contentText)

	// Build layout
	leftPanel := container.NewBorder(
		widget.NewLabelWithStyle("Categories", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		hb.categoryList,
	)

	middlePanel := container.NewBorder(
		hb.searchEntry,
		nil, nil, nil,
		hb.topicList,
	)

	rightPanel := contentScroll

	// Split view
	split := container.NewHSplit(
		container.NewHSplit(leftPanel, middlePanel),
		rightPanel,
	)
	split.SetOffset(0.4)

	// Create dialog
	hb.dialog = dialog.NewCustom("Help Browser (Shift+F1)", "Close", split, hb.window)
	hb.dialog.Resize(fyne.NewSize(900, 600))
}

func (hb *HelpBrowser) onSearch(query string) {
	if query == "" {
		hb.topics = hb.helpSystem.GetAllTopics()
	} else {
		hb.topics = hb.helpSystem.Search(query)
	}
	hb.topicList.Refresh()
}

func (hb *HelpBrowser) onCategorySelected(id widget.ListItemID) {
	if id == 0 {
		// All topics
		hb.topics = hb.helpSystem.GetAllTopics()
	} else {
		cat := hb.categories[id]
		catMap := hb.helpSystem.GetTopicsByCategory()
		if topics, ok := catMap[cat]; ok {
			hb.topics = topics
		} else {
			hb.topics = nil
		}
	}
	hb.topicList.Refresh()
	hb.searchEntry.SetText("")
}

func (hb *HelpBrowser) onTopicSelected(id widget.ListItemID) {
	if id >= len(hb.topics) {
		return
	}
	topic := hb.topics[id]
	hb.currentTopic = topic
	hb.contentText.ParseMarkdown(topic.Content)
}

// Show displays the help browser.
func (hb *HelpBrowser) Show() {
	hb.dialog.Show()
	hb.window.Canvas().Focus(hb.searchEntry)
}

// ShowTopic displays a specific help topic.
func (hb *HelpBrowser) ShowTopic(topic *HelpTopic) {
	if topic == nil {
		return
	}
	hb.currentTopic = topic
	hb.contentText.ParseMarkdown(topic.Content)
	hb.dialog.Show()
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ContextualHelpDialog shows a quick contextual help popup.
type ContextualHelpDialog struct {
	topic  *HelpTopic
	window fyne.Window
}

// NewContextualHelpDialog creates a popup for contextual help.
func NewContextualHelpDialog(topic *HelpTopic, window fyne.Window) *ContextualHelpDialog {
	return &ContextualHelpDialog{
		topic:  topic,
		window: window,
	}
}

// Show displays the contextual help popup.
func (chd *ContextualHelpDialog) Show() {
	if chd.topic == nil {
		return
	}

	content := widget.NewRichTextFromMarkdown(chd.topic.Content)
	content.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	// Add a "More Help" button
	moreBtn := widget.NewButtonWithIcon("Browse All Help", theme.HelpIcon(), func() {
		GetHelpSystem().ShowBrowser()
	})

	fullContent := container.NewBorder(nil, moreBtn, nil, nil, scroll)

	d := dialog.NewCustom(chd.topic.Title+" (F1)", "Close", fullContent, chd.window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}
