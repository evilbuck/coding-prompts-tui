# Application State

## Goal

Save the state of the application between sessions. This includes the saving of preferences and previously loaded folders.

**Technology**

This might require some research. I don't know if it makes sense to embed sqlite in here. Though, it might. I want to be able to do things like save presets of selected files.

What other technologies could I use. Preferably would like to embed it into this application. Is there a way to run sqlite using only Go libraries?

## MVP

Design a preferences data structure. Figure out the technology to use.
When selecting files for context, remember the parent folder and the selected files.

## Research Results & Technology Decision

### Technology Recommendation: JSON-based Configuration Storage

**Selected Technology**: JSON with file-based storage using Go's standard library

**Rationale**:
- **Simplicity**: Zero dependencies, built into Go standard library
- **Human-readable**: Users can manually edit config if needed
- **Deployment**: No external database files to manage
- **Performance**: Sufficient for expected data volumes (<1MB)
- **Development speed**: Fastest to implement and debug

**Storage Location**: `~/.config/prompter/config.json`

### Alternative Technologies Considered

1. **SQLite Options (Pure Go)**:
   - `glebarez/go-sqlite` - Pure Go, no CGO required
   - `modernc.org/sqlite` - High performance pure Go implementation
   - **Verdict**: Overkill for simple configuration data

2. **Key-Value Databases**:
   - `BoltDB/bbolt` - Simple, reliable, single-writer
   - `BadgerDB` - High performance, concurrent access
   - **Verdict**: Good options but unnecessary complexity for this use case

3. **Simple File Formats**:
   - JSON (chosen)
   - YAML with comments support
   - TOML for configuration-focused syntax

## Data Structure Design

```go
// AppConfig represents the complete application state
type AppConfig struct {
    Preferences      UserPreferences   `json:"preferences"`
    RecentWorkspaces []WorkspaceState  `json:"recent_workspaces"`
    Presets          map[string]FilePreset `json:"presets"`
    Metadata         ConfigMetadata    `json:"metadata"`
}

// UserPreferences stores user interface and behavior settings
type UserPreferences struct {
    // UI preferences
    Theme           string `json:"theme"`           // "dark", "light", "auto"
    ShowHiddenFiles bool   `json:"show_hidden"`     // Show .files and .directories
    FileTreeWidth   int    `json:"tree_width"`      // Width of file tree panel
    
    // Default behavior
    AutoSelectReadme     bool `json:"auto_select_readme"`      // Auto-select README files
    RememberLastPosition bool `json:"remember_position"`       // Remember cursor position
    MaxRecentWorkspaces  int  `json:"max_recent_workspaces"`   // Limit recent folders
    
    // Output preferences
    IncludeFileTree     bool   `json:"include_file_tree"`     // Include <filetree> in output
    DefaultSystemPrompt string `json:"default_system_prompt"` // Path to default persona
    
    // Performance preferences
    MaxFileSize       int64    `json:"max_file_size"`         // Skip files larger than this (bytes)
    ExcludePatterns   []string `json:"exclude_patterns"`      // Gitignore-style patterns
    IncludeExtensions []string `json:"include_extensions"`    // Only include these extensions (if set)
}

// WorkspaceState represents a previously loaded folder and its state
type WorkspaceState struct {
    Path           string    `json:"path"`              // Absolute path to workspace
    Name           string    `json:"name"`              // Display name (usually folder name)
    LastAccessed   time.Time `json:"last_accessed"`     // When last opened
    
    // Saved state for this workspace
    SelectedFiles  []string  `json:"selected_files"`    // Relative paths of selected files
    CursorPosition string    `json:"cursor_position"`   // Last cursor position in file tree
    ChatHistory    []string  `json:"chat_history"`      // Recent chat inputs for this workspace
    
    // Workspace metadata
    FileCount    int       `json:"file_count"`        // Total files in workspace (for display)
    LastScanTime time.Time `json:"last_scan_time"`    // When file tree was last scanned
}

// FilePreset represents a saved selection of files that can be reused
type FilePreset struct {
    Name          string    `json:"name"`              // User-defined name
    Description   string    `json:"description"`       // Optional description
    WorkspacePath string    `json:"workspace_path"`    // Associated workspace (optional)
    Files         []string  `json:"files"`             // Relative file paths
    SystemPrompt  string    `json:"system_prompt"`     // Associated persona/prompt
    CreatedAt     time.Time `json:"created_at"`
    LastUsed      time.Time `json:"last_used"`
    UseCount      int       `json:"use_count"`         // Track popularity
}

// ConfigMetadata stores application metadata
type ConfigMetadata struct {
    Version      string    `json:"version"`           // Config schema version
    AppVersion   string    `json:"app_version"`       // App version that created this config
    CreatedAt    time.Time `json:"created_at"`
    LastModified time.Time `json:"last_modified"`
    ConfigPath   string    `json:"config_path"`       // Where config is stored
}
```

## Implementation Plan

### 1. Create Configuration Package (`internal/config/`)
- **config.go**: Main configuration types and structures
- **manager.go**: ConfigManager with Load/Save operations
- **defaults.go**: Default configuration values and initialization

### 2. ConfigManager Implementation
```go
type ConfigManager struct {
    configPath string
    config     *AppConfig
    mutex      sync.RWMutex  // Thread safety for TUI
}

// Key operations:
// - LoadConfig() - Load on startup with fallback to defaults
// - SaveConfig() - Save when changes occur
// - AutoSave() - Periodic saves during long sessions
// - BackupConfig() - Create backups before major changes
```

### 3. Storage Strategy
```
~/.config/prompter/
├── config.json          # Main configuration
├── config.json.backup   # Automatic backup
└── presets/             # Optional: separate preset files
    ├── frontend.json
    └── backend.json
```

### 4. Integration Points
- **TUI State**: Connect Bubble Tea models to persistent state
- **File Selection**: Persist and restore file selections per workspace
- **User Preferences**: Apply saved UI preferences on startup
- **Workspace Memory**: Remember recent folders and their states
- **Preset Management**: Save/load file selection templates

### 5. Features to Implement
- **Configuration Loading**: Startup config loading with migration support
- **Auto-Save**: Periodic saving during application use
- **Workspace Memory**: Remember selected files per folder
- **Preset Management**: Save/load file selection presets
- **Backup & Recovery**: Automatic config backups and corruption recovery
