package help

import (
	"strings"
	"testing"
)

func TestGetHelpSystem(t *testing.T) {
	// GetHelpSystem should return a singleton
	hs1 := GetHelpSystem()
	hs2 := GetHelpSystem()
	if hs1 != hs2 {
		t.Error("GetHelpSystem should return the same instance")
	}
}

func TestEmbeddedTopicsRegistered(t *testing.T) {
	hs := GetHelpSystem()
	
	// Check that embedded topics are registered
	topics := hs.GetAllTopics()
	if len(topics) == 0 {
		t.Error("Expected embedded topics to be registered")
	}
	
	// Check for key topics
	expectedIDs := []string{
		"overview",
		"module.hysteresis",
		"module.crossbar",
		"module.mnist",
		"module.circuits",
		"module.comparison",
		"module.eda",
		"module.docs",
		"shortcuts",
		"troubleshooting",
		"about.fecim",
	}
	
	for _, id := range expectedIDs {
		topic := hs.GetTopic(id)
		if topic == nil {
			t.Errorf("Expected topic '%s' to be registered", id)
		}
	}
}

func TestGetTopicsByCategory(t *testing.T) {
	hs := GetHelpSystem()
	categories := hs.GetTopicsByCategory()
	
	if len(categories) == 0 {
		t.Error("Expected categories to be present")
	}
	
	// Check for expected categories
	expectedCategories := []string{"Getting Started", "Modules", "Reference", "About"}
	for _, cat := range expectedCategories {
		if _, ok := categories[cat]; !ok {
			t.Errorf("Expected category '%s' to exist", cat)
		}
	}
}

func TestSetContext(t *testing.T) {
	hs := GetHelpSystem()
	
	// Set context to hysteresis
	hs.SetContext("hysteresis")
	topic := hs.GetContextualTopic()
	if topic == nil {
		t.Error("Expected to get a contextual topic for hysteresis")
	}
	if topic != nil && topic.ID != "module.hysteresis" {
		t.Errorf("Expected module.hysteresis topic, got %s", topic.ID)
	}
	
	// Set context to home
	hs.SetContext("home")
	topic = hs.GetContextualTopic()
	if topic == nil {
		t.Error("Expected to get a contextual topic for home")
	}
	if topic != nil && topic.ID != "overview" {
		t.Errorf("Expected overview topic, got %s", topic.ID)
	}
}

func TestSearch(t *testing.T) {
	hs := GetHelpSystem()
	
	// Search for "hysteresis"
	results := hs.Search("hysteresis")
	if len(results) == 0 {
		t.Error("Expected search results for 'hysteresis'")
	}
	
	// First result should be the hysteresis module
	if len(results) > 0 {
		found := false
		for _, r := range results {
			if r.ID == "module.hysteresis" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected hysteresis module to be in search results")
		}
	}
	
	// Search for "F1"
	results = hs.Search("F1")
	if len(results) == 0 {
		t.Error("Expected search results for 'F1'")
	}
	
	// Search for nonsense should return nothing
	results = hs.Search("xyzzy12345")
	if len(results) != 0 {
		t.Error("Expected no results for nonsense query")
	}
}

func TestTopicContent(t *testing.T) {
	hs := GetHelpSystem()
	
	// Each topic should have content
	topics := hs.GetAllTopics()
	for _, topic := range topics {
		if topic.Title == "" {
			t.Errorf("Topic %s has empty title", topic.ID)
		}
		if topic.Content == "" {
			t.Errorf("Topic %s has empty content", topic.ID)
		}
		if topic.Category == "" {
			t.Errorf("Topic %s has empty category", topic.ID)
		}
		// Content should be markdown (start with #)
		if !strings.HasPrefix(strings.TrimSpace(topic.Content), "#") {
			t.Errorf("Topic %s content should start with markdown heading", topic.ID)
		}
	}
}

func TestModuleHelp(t *testing.T) {
	hs := GetHelpSystem()
	
	modules := []string{"hysteresis", "crossbar", "mnist", "circuits", "comparison", "eda", "docs"}
	for _, mod := range modules {
		topic := hs.ModuleHelp(mod)
		if topic == nil {
			t.Errorf("Expected help topic for module '%s'", mod)
		}
	}
}
