package widgets

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestNewEducationalPanel(t *testing.T) {
	config := EducationalPanelConfig{
		Title:   "Test Title",
		Content: "Test content goes here.",
		MinSize: fyne.NewSize(250, 300),
	}
	ep := NewEducationalPanel(config)

	if ep == nil {
		t.Fatal("NewEducationalPanel returned nil")
	}

	title, content := ep.GetContent()
	if title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got '%s'", title)
	}
	if content != "Test content goes here." {
		t.Errorf("Expected content 'Test content goes here.', got '%s'", content)
	}
	if ep.MinSize() != config.MinSize {
		t.Errorf("Expected MinSize %v, got %v", config.MinSize, ep.MinSize())
	}
}

func TestEducationalPanel_SetContent(t *testing.T) {
	ep := NewEducationalPanel(EducationalPanelConfig{})

	ep.SetContent("New Title", "New content")

	title, content := ep.GetContent()
	if title != "New Title" {
		t.Errorf("Expected title 'New Title', got '%s'", title)
	}
	if content != "New content" {
		t.Errorf("Expected content 'New content', got '%s'", content)
	}
}

func TestEducationalPanel_SetTitle(t *testing.T) {
	ep := NewEducationalPanel(EducationalPanelConfig{
		Title:   "Original",
		Content: "Content stays",
	})

	ep.SetTitle("Updated Title")

	title, content := ep.GetContent()
	if title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%s'", title)
	}
	if content != "Content stays" {
		t.Errorf("Content should remain unchanged, got '%s'", content)
	}
}

func TestEducationalPanel_AppendContent(t *testing.T) {
	ep := NewEducationalPanel(EducationalPanelConfig{
		Content: "Line 1",
	})

	ep.AppendContent("Line 2")

	_, content := ep.GetContent()
	if content != "Line 1\nLine 2" {
		t.Errorf("Expected appended content, got '%s'", content)
	}
}

func TestEducationalPanel_DefaultConfig(t *testing.T) {
	// Empty config should use defaults
	ep := NewEducationalPanel(EducationalPanelConfig{})

	if ep.MinSize().Width <= 0 || ep.MinSize().Height <= 0 {
		t.Error("Default MinSize should be positive")
	}

	title, content := ep.GetContent()
	if title == "" {
		t.Error("Default title should not be empty")
	}
	if content == "" {
		t.Error("Default content should not be empty")
	}
}

func TestCreateEducationalAccordion(t *testing.T) {
	sections := []EducationalSection{
		{
			Title:   "Section 1",
			Summary: "Summary 1",
			Details: "Details for section 1",
		},
		{
			Title:     "Section 2",
			Summary:   "Summary 2",
			Details:   "Details for section 2",
			LearnMore: "https://example.com",
		},
	}

	accordion := CreateEducationalAccordion(sections)
	if accordion == nil {
		t.Fatal("CreateEducationalAccordion returned nil")
	}
}

func TestCreateEducationalAccordion_Empty(t *testing.T) {
	accordion := CreateEducationalAccordion([]EducationalSection{})
	if accordion == nil {
		t.Fatal("CreateEducationalAccordion with empty sections returned nil")
	}
}
