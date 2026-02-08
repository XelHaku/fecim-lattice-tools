package presets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Manager handles preset storage, retrieval, and organization
type Manager struct {
	mu           sync.RWMutex
	presets      map[string]*Preset // ID -> Preset
	presetsDir   string
	providers    map[Module]PresetProvider
	changeHooks  []func(preset *Preset, action string)
}

// NewManager creates a new preset manager
func NewManager(presetsDir string) *Manager {
	m := &Manager{
		presets:    make(map[string]*Preset),
		presetsDir: presetsDir,
		providers:  make(map[Module]PresetProvider),
	}

	// Ensure directory exists
	os.MkdirAll(presetsDir, 0755)

	// Load built-in presets first
	m.loadBuiltInPresets()

	// Load user presets (will override built-ins with same ID if any)
	m.loadUserPresets()

	return m
}

// DefaultPresetsDir returns the default presets directory
func DefaultPresetsDir() string {
	// Look for the config directory relative to executable or in common locations
	candidates := []string{
		"presets",
		"config/presets",
		filepath.Join(os.Getenv("HOME"), ".config/fecim-lattice-tools/presets"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	// Default to local presets directory
	return "presets"
}

// RegisterProvider registers a preset provider for a module
func (m *Manager) RegisterProvider(provider PresetProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.GetModule()] = provider
}

// GetProvider returns the provider for a module
func (m *Manager) GetProvider(module Module) (PresetProvider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.providers[module]
	return p, ok
}

// AddChangeHook adds a callback that's called when presets change
func (m *Manager) AddChangeHook(hook func(preset *Preset, action string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.changeHooks = append(m.changeHooks, hook)
}

// notifyHooks calls all registered change hooks
func (m *Manager) notifyHooks(preset *Preset, action string) {
	m.mu.RLock()
	hooks := make([]func(preset *Preset, action string), len(m.changeHooks))
	copy(hooks, m.changeHooks)
	m.mu.RUnlock()

	for _, hook := range hooks {
		hook(preset, action)
	}
}

// Save persists a preset to disk
func (m *Manager) Save(preset *Preset) error {
	if err := preset.Validate(); err != nil {
		return err
	}

	m.mu.Lock()
	m.presets[preset.Metadata.ID] = preset
	m.mu.Unlock()

	// Don't save built-in presets to disk
	if preset.Metadata.BuiltIn {
		return nil
	}

	// Save to file
	filename := filepath.Join(m.presetsDir, preset.Metadata.ID+".json")
	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preset: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write preset file: %w", err)
	}

	m.notifyHooks(preset, "save")
	return nil
}

// Load retrieves a preset by ID
func (m *Manager) Load(id string) (*Preset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	preset, ok := m.presets[id]
	if !ok {
		return nil, fmt.Errorf("preset not found: %s", id)
	}
	return preset.Clone(), nil
}

// Delete removes a preset
func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	preset, ok := m.presets[id]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("preset not found: %s", id)
	}

	if preset.Metadata.BuiltIn {
		m.mu.Unlock()
		return fmt.Errorf("cannot delete built-in preset")
	}

	delete(m.presets, id)
	m.mu.Unlock()

	// Delete file
	filename := filepath.Join(m.presetsDir, id+".json")
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete preset file: %w", err)
	}

	m.notifyHooks(preset, "delete")
	return nil
}

// List returns all presets, optionally filtered
func (m *Manager) List(filters ...ListFilter) []*Preset {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Preset
	for _, p := range m.presets {
		include := true
		for _, f := range filters {
			if !f(p) {
				include = false
				break
			}
		}
		if include {
			result = append(result, p.Clone())
		}
	}

	// Sort by name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Metadata.Name < result[j].Metadata.Name
	})

	return result
}

// ListFilter is a function that filters presets
type ListFilter func(*Preset) bool

// FilterByModule returns a filter for a specific module
func FilterByModule(module Module) ListFilter {
	return func(p *Preset) bool {
		return p.Metadata.Module == module || p.Metadata.Module == ModuleGlobal
	}
}

// FilterByCategory returns a filter for a specific category
func FilterByCategory(category Category) ListFilter {
	return func(p *Preset) bool {
		return p.Metadata.Category == category
	}
}

// FilterByBuiltIn returns a filter for built-in or custom presets
func FilterByBuiltIn(builtIn bool) ListFilter {
	return func(p *Preset) bool {
		return p.Metadata.BuiltIn == builtIn
	}
}

// FilterByTag returns a filter for presets with a specific tag
func FilterByTag(tag string) ListFilter {
	return func(p *Preset) bool {
		for _, t := range p.Metadata.Tags {
			if strings.EqualFold(t, tag) {
				return true
			}
		}
		return false
	}
}

// SearchByName returns a filter matching preset names
func SearchByName(query string) ListFilter {
	query = strings.ToLower(query)
	return func(p *Preset) bool {
		return strings.Contains(strings.ToLower(p.Metadata.Name), query) ||
			strings.Contains(strings.ToLower(p.Metadata.Description), query)
	}
}

// Apply applies a preset to its registered provider
func (m *Manager) Apply(id string) error {
	preset, err := m.Load(id)
	if err != nil {
		return err
	}

	provider, ok := m.GetProvider(preset.Metadata.Module)
	if !ok {
		return fmt.Errorf("no provider registered for module: %s", preset.Metadata.Module)
	}

	if err := provider.ApplyPreset(preset); err != nil {
		return fmt.Errorf("failed to apply preset: %w", err)
	}

	m.notifyHooks(preset, "apply")
	return nil
}

// CreateFromCurrent creates a new preset from the current state of a provider
func (m *Manager) CreateFromCurrent(name, description string, module Module, category Category) (*Preset, error) {
	provider, ok := m.GetProvider(module)
	if !ok {
		return nil, fmt.Errorf("no provider registered for module: %s", module)
	}

	config := provider.GetCurrentConfig()
	preset := NewPreset(name, description, module, category, config)

	if err := m.Save(preset); err != nil {
		return nil, err
	}

	return preset, nil
}

// loadUserPresets loads all presets from the presets directory
func (m *Manager) loadUserPresets() {
	files, err := os.ReadDir(m.presetsDir)
	if err != nil {
		return // Directory might not exist yet
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
			m.loadPresetFile(filepath.Join(m.presetsDir, f.Name()))
		}
	}
}

// loadPresetFile loads a single preset file
func (m *Manager) loadPresetFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	var preset Preset
	if err := json.Unmarshal(data, &preset); err != nil {
		return
	}

	if err := preset.Validate(); err != nil {
		return
	}

	m.presets[preset.Metadata.ID] = &preset
}

// Export exports a preset to a file
func (m *Manager) Export(id, filename string) error {
	preset, err := m.Load(id)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal preset: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}

// Import imports a preset from a file
func (m *Manager) Import(filename string) (*Preset, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read preset file: %w", err)
	}

	var preset Preset
	if err := json.Unmarshal(data, &preset); err != nil {
		return nil, fmt.Errorf("failed to parse preset: %w", err)
	}

	// Clear built-in flag on import
	preset.Metadata.BuiltIn = false

	// Generate new ID to avoid conflicts
	preset.Metadata.ID = generateID(preset.Metadata.Name, preset.Metadata.Module)

	if err := m.Save(&preset); err != nil {
		return nil, err
	}

	return &preset, nil
}

// GetCategories returns all unique categories
func (m *Manager) GetCategories() []Category {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[Category]bool)
	for _, p := range m.presets {
		seen[p.Metadata.Category] = true
	}

	result := make([]Category, 0, len(seen))
	for c := range seen {
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool {
		return string(result[i]) < string(result[j])
	})
	return result
}

// GetTags returns all unique tags
func (m *Manager) GetTags() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]bool)
	for _, p := range m.presets {
		for _, t := range p.Metadata.Tags {
			seen[t] = true
		}
	}

	result := make([]string, 0, len(seen))
	for t := range seen {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}
