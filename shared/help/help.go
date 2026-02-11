// Package help provides an in-app help system with contextual help (F1),
// searchable documentation browser, and startup tips.
package help

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// HelpTopic represents a single help topic with searchable content.
type HelpTopic struct {
	ID          string   // Unique identifier (e.g., "module.hysteresis.materials")
	Title       string   // Display title
	Summary     string   // Short description for search results
	Content     string   // Full markdown content
	Keywords    []string // Additional search keywords
	Category    string   // Category for grouping (e.g., "Getting Started", "Modules")
	Module      string   // Associated module (e.g., "hysteresis", "crossbar")
	ContextKeys []string // Keys that trigger this topic for contextual help
}

// HelpSystem manages the in-app help system.
type HelpSystem struct {
	mu            sync.RWMutex
	topics        map[string]*HelpTopic
	contextMap    map[string]string // context key -> topic ID
	currentCtx    string            // Current help context
	window        fyne.Window
	onShowHelp    func(topic *HelpTopic)
	onShowBrowser func()
}

// globalHelpSystem is the singleton help system instance.
var (
	globalHelpSystem *HelpSystem
	globalOnce       sync.Once
)

// GetHelpSystem returns the global help system instance.
func GetHelpSystem() *HelpSystem {
	globalOnce.Do(func() {
		globalHelpSystem = &HelpSystem{
			topics:     make(map[string]*HelpTopic),
			contextMap: make(map[string]string),
		}
		// Register embedded help topics
		registerEmbeddedTopics(globalHelpSystem)
	})
	return globalHelpSystem
}

// Init initializes the help system with a window for dialogs.
func (hs *HelpSystem) Init(window fyne.Window) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.window = window
}

// SetCallbacks sets the callbacks for showing help.
func (hs *HelpSystem) SetCallbacks(onShowHelp func(topic *HelpTopic), onShowBrowser func()) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.onShowHelp = onShowHelp
	hs.onShowBrowser = onShowBrowser
}

// RegisterTopic adds a help topic to the system.
func (hs *HelpSystem) RegisterTopic(topic *HelpTopic) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.topics[topic.ID] = topic
	for _, key := range topic.ContextKeys {
		hs.contextMap[key] = topic.ID
	}
}

// GetTopic retrieves a topic by ID.
func (hs *HelpSystem) GetTopic(id string) *HelpTopic {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return hs.topics[id]
}

// GetAllTopics returns all registered topics.
func (hs *HelpSystem) GetAllTopics() []*HelpTopic {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	topics := make([]*HelpTopic, 0, len(hs.topics))
	for _, t := range hs.topics {
		topics = append(topics, t)
	}
	// Sort by category then title
	sort.Slice(topics, func(i, j int) bool {
		if topics[i].Category != topics[j].Category {
			return topics[i].Category < topics[j].Category
		}
		return topics[i].Title < topics[j].Title
	})
	return topics
}

// GetTopicsByCategory returns topics grouped by category.
func (hs *HelpSystem) GetTopicsByCategory() map[string][]*HelpTopic {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	result := make(map[string][]*HelpTopic)
	for _, t := range hs.topics {
		result[t.Category] = append(result[t.Category], t)
	}
	// Sort within each category
	for cat := range result {
		sort.Slice(result[cat], func(i, j int) bool {
			return result[cat][i].Title < result[cat][j].Title
		})
	}
	return result
}

// SetContext sets the current help context (e.g., when switching modules).
func (hs *HelpSystem) SetContext(contextKey string) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.currentCtx = contextKey
}

// GetContextualTopic returns the help topic for the current context.
func (hs *HelpSystem) GetContextualTopic() *HelpTopic {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	if topicID, ok := hs.contextMap[hs.currentCtx]; ok {
		return hs.topics[topicID]
	}
	// Fall back to overview
	if t, ok := hs.topics["overview"]; ok {
		return t
	}
	return nil
}

// ShowContextualHelp shows help for the current context (F1 behavior).
func (hs *HelpSystem) ShowContextualHelp() {
	topic := hs.GetContextualTopic()
	hs.mu.RLock()
	callback := hs.onShowHelp
	browserCallback := hs.onShowBrowser
	hs.mu.RUnlock()

	if topic != nil && callback != nil {
		callback(topic)
	} else if browserCallback != nil {
		// No contextual help, show browser
		browserCallback()
	}
}

// ShowBrowser shows the help browser.
func (hs *HelpSystem) ShowBrowser() {
	hs.mu.RLock()
	callback := hs.onShowBrowser
	hs.mu.RUnlock()
	if callback != nil {
		callback()
	}
}

// Search performs a fuzzy search across all topics.
func (hs *HelpSystem) Search(query string) []*HelpTopic {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	if query == "" {
		return hs.GetAllTopics()
	}

	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)
	var results []*HelpTopic
	scores := make(map[string]int)

	for _, topic := range hs.topics {
		score := 0

		// Title match (highest priority)
		titleLower := strings.ToLower(topic.Title)
		if strings.Contains(titleLower, queryLower) {
			score += 100
		}
		for _, term := range queryTerms {
			if strings.Contains(titleLower, term) {
				score += 50
			}
		}

		// Summary match
		summaryLower := strings.ToLower(topic.Summary)
		for _, term := range queryTerms {
			if strings.Contains(summaryLower, term) {
				score += 20
			}
		}

		// Keywords match
		for _, keyword := range topic.Keywords {
			kwLower := strings.ToLower(keyword)
			if strings.Contains(kwLower, queryLower) {
				score += 30
			}
			for _, term := range queryTerms {
				if strings.Contains(kwLower, term) {
					score += 15
				}
			}
		}

		// Content match
		contentLower := strings.ToLower(topic.Content)
		for _, term := range queryTerms {
			if strings.Contains(contentLower, term) {
				score += 5
			}
		}

		if score > 0 {
			results = append(results, topic)
			scores[topic.ID] = score
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return scores[results[i].ID] > scores[results[j].ID]
	})

	return results
}

// SetupF1Shortcut configures keyboard help shortcuts.
func SetupF1Shortcut(window fyne.Window, helpSystem *HelpSystem) {
	if window == nil || helpSystem == nil {
		return
	}

	// F1: accessibility keyboard navigation help (A11Y-4)
	f1Shortcut := &desktop.CustomShortcut{KeyName: fyne.KeyF1}
	window.Canvas().AddShortcut(f1Shortcut, func(shortcut fyne.Shortcut) {
		sharedwidgets.ShowKeyboardHelp(window)
	})

	// Shift+F1: contextual module help
	shiftF1Shortcut := &desktop.CustomShortcut{
		KeyName:  fyne.KeyF1,
		Modifier: fyne.KeyModifierShift,
	}
	window.Canvas().AddShortcut(shiftF1Shortcut, func(shortcut fyne.Shortcut) {
		helpSystem.ShowContextualHelp()
	})

	// Ctrl+F1: full help browser
	ctrlF1Shortcut := &desktop.CustomShortcut{
		KeyName:  fyne.KeyF1,
		Modifier: fyne.KeyModifierControl,
	}
	window.Canvas().AddShortcut(ctrlF1Shortcut, func(shortcut fyne.Shortcut) {
		helpSystem.ShowBrowser()
	})
}

// ModuleHelp returns the main help topic for a module.
func (hs *HelpSystem) ModuleHelp(module string) *HelpTopic {
	id := fmt.Sprintf("module.%s", module)
	return hs.GetTopic(id)
}
