package recentfiles

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// MenuBuilder helps construct File menu with recent files
type MenuBuilder struct {
	manager       *Manager
	onOpenRecent  func(file *RecentFile)
	onClearRecent func()
	maxItems      int
}

// NewMenuBuilder creates a menu builder for recent files
func NewMenuBuilder(manager *Manager) *MenuBuilder {
	return &MenuBuilder{
		manager:  manager,
		maxItems: 10,
	}
}

// SetMaxItems sets the maximum number of recent files to show
func (b *MenuBuilder) SetMaxItems(n int) *MenuBuilder {
	b.maxItems = n
	return b
}

// OnOpenRecent sets the callback for when a recent file is selected
func (b *MenuBuilder) OnOpenRecent(callback func(file *RecentFile)) *MenuBuilder {
	b.onOpenRecent = callback
	return b
}

// OnClearRecent sets the callback for clearing recent files
func (b *MenuBuilder) OnClearRecent(callback func()) *MenuBuilder {
	b.onClearRecent = callback
	return b
}

// BuildRecentFilesMenu creates the recent files submenu
func (b *MenuBuilder) BuildRecentFilesMenu() *fyne.Menu {
	items := b.buildRecentItems(FileTypeAny)
	return fyne.NewMenu("Recent Files", items...)
}

// BuildRecentConfigsMenu creates a submenu for recent configs only
func (b *MenuBuilder) BuildRecentConfigsMenu() *fyne.Menu {
	items := b.buildRecentItems(FileTypeConfig)
	return fyne.NewMenu("Recent Configs", items...)
}

// BuildRecentExportsMenu creates a submenu for recent exports only
func (b *MenuBuilder) BuildRecentExportsMenu() *fyne.Menu {
	items := b.buildRecentItems(FileTypeExport)
	return fyne.NewMenu("Recent Exports", items...)
}

// BuildRecentProjectsMenu creates a submenu for recent projects only
func (b *MenuBuilder) BuildRecentProjectsMenu() *fyne.Menu {
	items := b.buildRecentItems(FileTypeProject)
	return fyne.NewMenu("Recent Projects", items...)
}

// BuildRecentPresetsMenu creates a submenu for recent presets only  
func (b *MenuBuilder) BuildRecentPresetsMenu() *fyne.Menu {
	items := b.buildRecentItems(FileTypePreset)
	return fyne.NewMenu("Recent Presets", items...)
}

// buildRecentItems creates menu items for recent files of a given type
func (b *MenuBuilder) buildRecentItems(fileType FileType) []*fyne.MenuItem {
	files := b.manager.List(fileType)

	// Limit number of items
	if len(files) > b.maxItems {
		files = files[:b.maxItems]
	}

	if len(files) == 0 {
		return []*fyne.MenuItem{
			fyne.NewMenuItem("(No recent files)", nil),
		}
	}

	items := make([]*fyne.MenuItem, 0, len(files)+2)

	for _, f := range files {
		file := f // Capture for closure
		label := b.formatMenuItem(file)
		item := fyne.NewMenuItem(label, func() {
			if b.onOpenRecent != nil {
				b.onOpenRecent(file)
			}
		})

		// Mark missing files
		if !file.Exists {
			item.Disabled = true
		}

		items = append(items, item)
	}

	// Add separator and Clear action
	items = append(items, fyne.NewMenuItemSeparator())
	clearItem := fyne.NewMenuItem("Clear Recent Files", func() {
		if fileType == FileTypeAny {
			b.manager.ClearAll()
		} else {
			b.manager.Clear(fileType)
		}
		if b.onClearRecent != nil {
			b.onClearRecent()
		}
	})
	items = append(items, clearItem)

	return items
}

// formatMenuItem creates a display label for a recent file
func (b *MenuBuilder) formatMenuItem(file *RecentFile) string {
	// Shorten path for display
	name := file.Name
	dir := filepath.Dir(file.Path)
	shortDir := shortenPath(dir, 30)

	if shortDir != "" && shortDir != "." {
		return fmt.Sprintf("%s - %s", name, shortDir)
	}
	return name
}

// shortenPath truncates a path to fit within maxLen characters
func shortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}

	// Try to keep the last part of the path
	parts := filepath.SplitList(path)
	if len(parts) == 0 {
		parts = []string{path}
	}

	// For single path, just truncate
	if len(path) > maxLen {
		return "..." + path[len(path)-maxLen+3:]
	}
	return path
}

// CreateFileMenu creates a complete File menu with standard items and recent files
// This is a convenience function that combines common File menu items with recent files
func CreateFileMenu(
	manager *Manager,
	onNew func(),
	onOpen func(),
	onSave func(),
	onSaveAs func(),
	onExport func(),
	onOpenRecent func(file *RecentFile),
	onQuit func(),
) *fyne.Menu {
	items := make([]*fyne.MenuItem, 0)

	// New
	if onNew != nil {
		items = append(items, fyne.NewMenuItem("New", onNew))
	}

	// Open
	if onOpen != nil {
		items = append(items, fyne.NewMenuItem("Open...", onOpen))
	}

	// Recent Files submenu
	if manager != nil {
		builder := NewMenuBuilder(manager)
		builder.OnOpenRecent(onOpenRecent)
		recentMenu := builder.BuildRecentFilesMenu()
		recentItem := fyne.NewMenuItem("Recent Files", nil)
		recentItem.ChildMenu = recentMenu
		items = append(items, recentItem)
	}

	items = append(items, fyne.NewMenuItemSeparator())

	// Save
	if onSave != nil {
		items = append(items, fyne.NewMenuItem("Save", onSave))
	}

	// Save As
	if onSaveAs != nil {
		items = append(items, fyne.NewMenuItem("Save As...", onSaveAs))
	}

	// Export
	if onExport != nil {
		items = append(items, fyne.NewMenuItemSeparator())
		items = append(items, fyne.NewMenuItem("Export...", onExport))
	}

	// Quit
	if onQuit != nil {
		items = append(items, fyne.NewMenuItemSeparator())
		items = append(items, fyne.NewMenuItem("Quit", onQuit))
	}

	return fyne.NewMenu("File", items...)
}

// RefreshableRecentMenu wraps a recent files menu that can be refreshed
type RefreshableRecentMenu struct {
	manager     *Manager
	menu        *fyne.Menu
	menuItem    *fyne.MenuItem
	fileType    FileType
	onOpen      func(file *RecentFile)
	onClear     func()
	maxItems    int
}

// NewRefreshableRecentMenu creates a recent files menu that can refresh its contents
func NewRefreshableRecentMenu(manager *Manager, title string, fileType FileType) *RefreshableRecentMenu {
	r := &RefreshableRecentMenu{
		manager:  manager,
		fileType: fileType,
		maxItems: 10,
	}

	// Create initial menu
	r.menu = fyne.NewMenu(title)
	r.menuItem = fyne.NewMenuItem(title, nil)
	r.menuItem.ChildMenu = r.menu

	// Register for updates
	manager.OnChange(func([]*RecentFile) {
		r.Refresh()
	})

	r.Refresh()
	return r
}

// SetOnOpen sets the callback for opening a recent file
func (r *RefreshableRecentMenu) SetOnOpen(callback func(file *RecentFile)) {
	r.onOpen = callback
}

// SetOnClear sets the callback for when recent files are cleared
func (r *RefreshableRecentMenu) SetOnClear(callback func()) {
	r.onClear = callback
}

// SetMaxItems sets the maximum number of items to show
func (r *RefreshableRecentMenu) SetMaxItems(n int) {
	r.maxItems = n
}

// MenuItem returns the menu item for use in a parent menu
func (r *RefreshableRecentMenu) MenuItem() *fyne.MenuItem {
	return r.menuItem
}

// Menu returns the underlying menu
func (r *RefreshableRecentMenu) Menu() *fyne.Menu {
	return r.menu
}

// Refresh updates the menu items from the manager
func (r *RefreshableRecentMenu) Refresh() {
	files := r.manager.List(r.fileType)

	if len(files) > r.maxItems {
		files = files[:r.maxItems]
	}

	// Clear existing items
	r.menu.Items = nil

	if len(files) == 0 {
		r.menu.Items = append(r.menu.Items, fyne.NewMenuItem("(No recent files)", nil))
	} else {
		for _, f := range files {
			file := f // Capture
			label := formatMenuLabel(file)
			item := fyne.NewMenuItem(label, func() {
				if r.onOpen != nil {
					r.onOpen(file)
				}
			})
			if !file.Exists {
				item.Disabled = true
			}
			r.menu.Items = append(r.menu.Items, item)
		}
	}

	// Add clear option
	r.menu.Items = append(r.menu.Items, fyne.NewMenuItemSeparator())
	r.menu.Items = append(r.menu.Items, fyne.NewMenuItem("Clear", func() {
		r.manager.Clear(r.fileType)
		if r.onClear != nil {
			r.onClear()
		}
	}))

	r.menu.Refresh()
}

func formatMenuLabel(file *RecentFile) string {
	name := file.Name
	dir := filepath.Dir(file.Path)

	// Shorten directory path
	if len(dir) > 25 {
		dir = "..." + dir[len(dir)-22:]
	}

	if dir != "." && dir != "" {
		return fmt.Sprintf("%s (%s)", name, dir)
	}
	return name
}
