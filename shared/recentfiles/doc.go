// Package recentfiles provides tracking of recently opened and saved files
// across FeCIM Lattice Tools sessions.
//
// # Overview
//
// The recentfiles package tracks files that users open, save, export, or import
// in the application. It supports multiple file types (configs, exports, projects,
// presets) and persists the list across sessions using Fyne preferences.
//
// # Usage
//
// Initialize the global manager during app startup:
//
//	prefs := fyneApp.Preferences()
//	recentFilesManager := recentfiles.InitGlobal(prefs)
//
// Then from any module, track file access:
//
//	// When loading a config file
//	recentfiles.GlobalAddConfig(filePath, "hysteresis")
//
//	// When exporting data
//	recentfiles.GlobalAddExport(exportPath, "crossbar")
//
//	// When opening a project
//	recentfiles.GlobalAddProject(projectPath, "eda")
//
// # Creating Recent Files Menu
//
// The package provides helpers to create File menu items:
//
//	// Create a refreshable recent files submenu
//	recentMenu := recentfiles.NewRefreshableRecentMenu(manager, "Recent Files", recentfiles.FileTypeAny)
//	recentMenu.SetOnOpen(func(file *recentfiles.RecentFile) {
//		// Handle opening the file
//		log.Printf("Opening: %s", file.Path)
//	})
//
//	// Add to File menu
//	fileMenu := fyne.NewMenu("File",
//		fyne.NewMenuItem("Open...", onOpen),
//		recentMenu.MenuItem(),
//		// ... other items
//	)
//
// # File Types
//
// The package supports these file types:
//   - FileTypeConfig: Configuration and preset files
//   - FileTypeExport: Exported data (CSV, JSON, PNG)
//   - FileTypeProject: Project files or directories
//   - FileTypePreset: Named preset configurations
//   - FileTypeAny: Used for listing/filtering - matches all types
//
// # Persistence
//
// Recent files are automatically persisted using Fyne preferences.
// The list is loaded on manager creation and saved after each modification.
// Files that no longer exist can be cleaned up using CleanupMissing().
package recentfiles
