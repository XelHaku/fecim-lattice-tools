// Package presets provides a unified preset system for saving/loading named configurations
// across all FeCIM Lattice Tools modules.
package presets

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Category represents the type of preset
type Category string

const (
	CategoryEducational Category = "educational"
	CategoryResearch    Category = "research"
	CategoryDemo        Category = "demo"
	CategoryCustom      Category = "custom"
	CategoryBenchmark   Category = "benchmark"
)

// Module identifies which module a preset applies to
type Module string

const (
	ModuleGlobal     Module = "global"
	ModuleHysteresis Module = "hysteresis"
	ModuleCrossbar   Module = "crossbar"
	ModuleMNIST      Module = "mnist"
	ModuleCircuits   Module = "circuits"
	ModuleComparison Module = "comparison"
	ModuleEDA        Module = "eda"
)

// AllModules returns all available module identifiers
func AllModules() []Module {
	return []Module{
		ModuleGlobal,
		ModuleHysteresis,
		ModuleCrossbar,
		ModuleMNIST,
		ModuleCircuits,
		ModuleComparison,
		ModuleEDA,
	}
}

// Metadata contains common information about a preset
type Metadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    Category  `json:"category"`
	Module      Module    `json:"module"`
	Author      string    `json:"author,omitempty"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags,omitempty"`
	BuiltIn     bool      `json:"built_in"`
}

// Preset represents a complete preset configuration
type Preset struct {
	Metadata Metadata               `json:"metadata"`
	Config   map[string]interface{} `json:"config"`
}

// NewPreset creates a new preset with the given metadata and configuration
func NewPreset(name, description string, module Module, category Category, config map[string]interface{}) *Preset {
	now := time.Now()
	return &Preset{
		Metadata: Metadata{
			ID:          generateID(name, module),
			Name:        name,
			Description: description,
			Category:    category,
			Module:      module,
			Version:     "1.0.0",
			CreatedAt:   now,
			UpdatedAt:   now,
			BuiltIn:     false,
		},
		Config: config,
	}
}

// generateID creates a unique identifier from name and module
func generateID(name string, module Module) string {
	return fmt.Sprintf("%s-%s-%d", module, sanitizeName(name), time.Now().UnixNano())
}

// sanitizeName converts a name to a valid ID component
func sanitizeName(name string) string {
	result := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		} else if c == ' ' || c == '-' || c == '_' {
			result += "-"
		}
	}
	return result
}

// Clone creates a deep copy of the preset
func (p *Preset) Clone() *Preset {
	configCopy := make(map[string]interface{})
	data, _ := json.Marshal(p.Config)
	json.Unmarshal(data, &configCopy)

	return &Preset{
		Metadata: p.Metadata,
		Config:   configCopy,
	}
}

// GetFloat returns a float64 value from config
func (p *Preset) GetFloat(key string) (float64, bool) {
	if v, ok := p.Config[key]; ok {
		switch val := v.(type) {
		case float64:
			return val, true
		case float32:
			return float64(val), true
		case int:
			return float64(val), true
		case int64:
			return float64(val), true
		}
	}
	return 0, false
}

// GetInt returns an int value from config
func (p *Preset) GetInt(key string) (int, bool) {
	if v, ok := p.Config[key]; ok {
		switch val := v.(type) {
		case int:
			return val, true
		case int64:
			return int(val), true
		case float64:
			return int(val), true
		case float32:
			return int(val), true
		}
	}
	return 0, false
}

// GetString returns a string value from config
func (p *Preset) GetString(key string) (string, bool) {
	if v, ok := p.Config[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// GetBool returns a bool value from config
func (p *Preset) GetBool(key string) (bool, bool) {
	if v, ok := p.Config[key]; ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// SetValue sets a value in the config
func (p *Preset) SetValue(key string, value interface{}) {
	p.Config[key] = value
	p.Metadata.UpdatedAt = time.Now()
}

// Merge combines another preset's config into this one (other takes precedence)
func (p *Preset) Merge(other *Preset) {
	for k, v := range other.Config {
		p.Config[k] = v
	}
	p.Metadata.UpdatedAt = time.Now()
}

// Validate checks if the preset has required fields
func (p *Preset) Validate() error {
	if p.Metadata.Name == "" {
		return fmt.Errorf("preset name is required")
	}
	if p.Metadata.Module == "" {
		return fmt.Errorf("preset module is required")
	}
	if p.Config == nil {
		return fmt.Errorf("preset config is required")
	}
	if p.Metadata.ID == "" {
		return fmt.Errorf("preset id is required")
	}
	if strings.Contains(p.Metadata.ID, "..") || filepath.Base(p.Metadata.ID) != p.Metadata.ID || strings.ContainsAny(p.Metadata.ID, `/\\`) {
		return fmt.Errorf("preset id contains unsafe path characters")
	}
	return nil
}

// PresetProvider is an interface for modules that support presets
type PresetProvider interface {
	// GetCurrentConfig returns the current configuration as a preset config map
	GetCurrentConfig() map[string]interface{}

	// ApplyPreset applies a preset configuration
	ApplyPreset(preset *Preset) error

	// GetModule returns the module identifier
	GetModule() Module

	// GetPresetKeys returns the list of configuration keys supported
	GetPresetKeys() []string
}
