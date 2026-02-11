//go:build !ci
// +build !ci

package main

import (
	"fmt"
	"reflect"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ptrID returns the pointer address of an interface value for deduplication.
func ptrID(o any) uintptr {
	v := reflect.ValueOf(o)
	if !v.IsValid() {
		return 0
	}
	if v.Kind() != reflect.Pointer {
		return 0
	}
	return v.Pointer()
}

// Advanced UI discovery functions for the screenshot crawler

// findInteractiveElements discovers all interactive UI elements that could trigger overlays or dialogs
func findInteractiveElements(root fyne.CanvasObject) []interactiveElement {
	seen := map[uintptr]bool{}
	elements := make([]interactiveElement, 0) // Ensure non-nil slice
	
	// Return empty slice if root is nil
	if root == nil {
		return elements
	}
	
	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		
		ptr := ptrID(o)
		if ptr != 0 && seen[ptr] {
			return
		}
		if ptr != 0 {
			seen[ptr] = true
		}
		
		// Check for different types of interactive elements
		switch obj := o.(type) {
		case *widget.Button:
			if obj.OnTapped != nil && obj.Text != "" {
				elements = append(elements, interactiveElement{
					Type:   "button",
					Text:   obj.Text,
					Object: obj,
					Action: "tap",
				})
			}
		case *widget.Hyperlink:
			if obj.OnTapped != nil && obj.Text != "" {
				elements = append(elements, interactiveElement{
					Type:   "hyperlink", 
					Text:   obj.Text,
					Object: obj,
					Action: "tap",
				})
			}
		case *widget.Entry:
			elements = append(elements, interactiveElement{
				Type:   "entry",
				Text:   obj.PlaceHolder,
				Object: obj,
				Action: "focus",
			})
		case *widget.Select:
			elements = append(elements, interactiveElement{
				Type:   "select",
				Text:   fmt.Sprintf("Select (%d options)", len(obj.Options)),
				Object: obj,
				Action: "expand",
			})
		case *widget.Check:
			elements = append(elements, interactiveElement{
				Type:   "check",
				Text:   obj.Text,
				Object: obj,
				Action: "toggle",
			})
		case *widget.RadioGroup:
			elements = append(elements, interactiveElement{
				Type:   "radio",
				Text:   fmt.Sprintf("Radio (%d options)", len(obj.Options)),
				Object: obj,
				Action: "select",
			})
		case *container.AppTabs:
			for i, item := range obj.Items {
				elements = append(elements, interactiveElement{
					Type:   "tab",
					Text:   item.Text,
					Object: obj,
					Action: fmt.Sprintf("select-%d", i),
				})
			}
		}
		
		// Recurse into containers
		if tabs, ok := o.(*container.AppTabs); ok {
			for _, item := range tabs.Items {
				walk(item.Content)
			}
		} else if container, ok := o.(*fyne.Container); ok {
			for _, child := range container.Objects {
				walk(child)
			}
		} else {
			// Handle wrapped content via reflection
			v := reflect.ValueOf(o)
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}
			if v.IsValid() && v.Kind() == reflect.Struct {
				for _, fieldName := range []string{"Content", "content", "Objects", "objects"} {
					f := v.FieldByName(fieldName)
					if f.IsValid() && f.CanInterface() {
						if child, ok := f.Interface().(fyne.CanvasObject); ok {
							walk(child)
						} else if children, ok := f.Interface().([]fyne.CanvasObject); ok {
							for _, child := range children {
								walk(child)
							}
						}
					}
				}
			}
		}
	}
	
	walk(root)
	return elements
}

// interactiveElement represents a UI element that can be interacted with
type interactiveElement struct {
	Type   string
	Text   string
	Object fyne.CanvasObject
	Action string
}

// findPopupTriggers identifies buttons likely to trigger popups or dialogs
func findPopupTriggers(elements []interactiveElement) []interactiveElement {
	triggers := make([]interactiveElement, 0) // Ensure non-nil slice
	
	// Keywords that commonly indicate popup/dialog triggers
	triggerKeywords := map[string]bool{
		// Dialog triggers
		"about":        true,
		"settings":     true,
		"preferences":  true,
		"config":       true,
		"configuration": true,
		"options":      true,
		"tools":        true,
		"help":         true,
		"info":         true,
		"details":      true,
		"properties":   true,
		
		// Modal triggers
		"export":       true,
		"import":       true,
		"save":         true,
		"load":         true,
		"open":         true,
		"new":          true,
		"create":       true,
		"add":          true,
		"edit":         true,
		"delete":       true,
		"remove":       true,
		
		// Info/overlay triggers
		"learn":        true,
		"more":         true,
		"show":         true,
		"view":         true,
		"display":      true,
		"advanced":     true,
		"expert":       true,
		
		// Material/component pickers
		"material":     true,
		"select":       true,
		"choose":       true,
		"pick":         true,
		"browse":       true,
	}
	
	for _, element := range elements {
		if element.Type != "button" && element.Type != "hyperlink" {
			continue
		}
		
		text := strings.ToLower(strings.TrimSpace(element.Text))
		words := strings.Fields(text)
		
		for _, word := range words {
			// Clean word of punctuation
			word = strings.Trim(word, ".,!?()[]{}:;")
			if triggerKeywords[word] {
				triggers = append(triggers, element)
				break
			}
		}
	}
	
	return triggers
}

// findCloseElements identifies elements that likely close dialogs or overlays
func findCloseElements(elements []interactiveElement) []interactiveElement {
	closers := make([]interactiveElement, 0) // Ensure non-nil slice
	
	closeKeywords := map[string]bool{
		"close":    true,
		"cancel":   true,
		"back":     true,
		"dismiss":  true,
		"ok":       true,
		"done":     true,
		"finish":   true,
		"exit":     true,
		"quit":     true,
		"hide":     true,
	}
	
	for _, element := range elements {
		if element.Type != "button" {
			continue
		}
		
		text := strings.ToLower(strings.TrimSpace(element.Text))
		if closeKeywords[text] {
			closers = append(closers, element)
		}
	}
	
	return closers
}

// triggerInteractiveElement safely triggers an interactive element
func triggerInteractiveElement(element interactiveElement) bool {
	switch element.Type {
	case "button":
		if btn, ok := element.Object.(*widget.Button); ok && btn.OnTapped != nil {
			btn.OnTapped()
			return true
		}
	case "hyperlink":
		if link, ok := element.Object.(*widget.Hyperlink); ok && link.OnTapped != nil {
			link.OnTapped()
			return true
		}
	case "tab":
		if tabs, ok := element.Object.(*container.AppTabs); ok {
			// Parse the tab index from action
			var tabIndex int
			if n, err := fmt.Sscanf(element.Action, "select-%d", &tabIndex); n == 1 && err == nil {
				if tabIndex >= 0 && tabIndex < len(tabs.Items) {
					tabs.SelectIndex(tabIndex)
					return true
				}
			}
		}
	case "check":
		if check, ok := element.Object.(*widget.Check); ok {
			check.SetChecked(!check.Checked)
			return true
		}
	}
	
	return false
}

// findMenus discovers menu-like structures that might contain additional options
func findMenus(root fyne.CanvasObject) []menuInfo {
	seen := map[uintptr]bool{}
	menus := make([]menuInfo, 0) // Ensure non-nil slice
	
	// Return empty slice if root is nil
	if root == nil {
		return menus
	}
	
	var walk func(o fyne.CanvasObject)
	walk = func(o fyne.CanvasObject) {
		if o == nil {
			return
		}
		
		ptr := ptrID(o)
		if ptr != 0 && seen[ptr] {
			return
		}
		if ptr != 0 {
			seen[ptr] = true
		}
		
		// Look for menu-like structures
		switch obj := o.(type) {
		case *widget.Select:
			menus = append(menus, menuInfo{
				Type:    "select",
				Options: obj.Options,
				Object:  obj,
			})
		case *widget.RadioGroup:
			menus = append(menus, menuInfo{
				Type:    "radio",
				Options: obj.Options,
				Object:  obj,
			})
		}
		
		// Recurse
		if tabs, ok := o.(*container.AppTabs); ok {
			for _, item := range tabs.Items {
				walk(item.Content)
			}
		} else if container, ok := o.(*fyne.Container); ok {
			for _, child := range container.Objects {
				walk(child)
			}
		} else {
			v := reflect.ValueOf(o)
			if v.Kind() == reflect.Pointer {
				v = v.Elem()
			}
			if v.IsValid() && v.Kind() == reflect.Struct {
				for _, fieldName := range []string{"Content", "content"} {
					f := v.FieldByName(fieldName)
					if f.IsValid() && f.CanInterface() {
						if child, ok := f.Interface().(fyne.CanvasObject); ok {
							walk(child)
						}
					}
				}
			}
		}
	}
	
	walk(root)
	return menus
}

// menuInfo represents a menu-like UI structure
type menuInfo struct {
	Type    string
	Options []string
	Object  fyne.CanvasObject
}

// exploreMenuOptions triggers different states of menu components
func exploreMenuOptions(menu menuInfo) []string {
	var states []string
	
	switch menu.Type {
	case "select":
		if sel, ok := menu.Object.(*widget.Select); ok {
			original := sel.Selected
			for i, option := range sel.Options {
				sel.SetSelectedIndex(i)
				states = append(states, fmt.Sprintf("select-option-%d-%s", i, safeCrawlerNameHelper(option)))
			}
			// Restore original
			sel.SetSelected(original)
		}
	case "radio":
		if radio, ok := menu.Object.(*widget.RadioGroup); ok {
			original := radio.Selected
			for i, option := range radio.Options {
				radio.SetSelected(option)
				states = append(states, fmt.Sprintf("radio-option-%d-%s", i, safeCrawlerNameHelper(option)))
			}
			// Restore original
			radio.SetSelected(original)
		}
	}
	
	return states
}

// findToolTips discovers elements that might have tooltips (future implementation)
func findToolTips(root fyne.CanvasObject) []tooltipInfo {
	tips := make([]tooltipInfo, 0) // Ensure non-nil slice
	
	// Placeholder for tooltip discovery
	// Fyne tooltips are not easily discoverable via reflection
	// This would require platform-specific implementations
	
	return tips
}

// tooltipInfo represents a UI element that might have a tooltip
type tooltipInfo struct {
	Element fyne.CanvasObject
	Text    string
}

// safeCrawlerName is already defined in the main file, but we include it here for completeness
func safeCrawlerNameHelper(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	s = strings.ReplaceAll(s, ":", "-")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	if s == "" {
		s = "unnamed"
	}
	return s
}