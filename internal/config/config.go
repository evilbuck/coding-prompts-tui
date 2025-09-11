package config

import "time"

// AppConfig represents the complete application state
type AppConfig struct {
	RecentWorkspaces map[string]*WorkspaceState `json:"recent_workspaces"`
	UISettings       UISettings                 `json:"ui_settings"`
	Metadata         ConfigMetadata             `json:"metadata"`
}

// WorkspaceState represents a previously loaded folder and its state
type WorkspaceState struct {
	Path          string    `json:"path"`            // Absolute path to workspace
	LastAccessed  time.Time `json:"last_accessed"`   // When last opened
	SelectedFiles []string  `json:"selected_files"`  // Relative paths of selected files
	ChatInput     string    `json:"chat_input"`      // Saved chat input
	CurrentPersona string   `json:"current_persona"` // Active persona name (defaults to "default")
}

// ConfigMetadata stores application metadata
type ConfigMetadata struct {
	Version      string    `json:"version"`       // Config schema version
	AppVersion   string    `json:"app_version"`   // App version that created this config
	CreatedAt    time.Time `json:"created_at"`
	LastModified time.Time `json:"last_modified"`
}

// UISettings contains user interface configuration options
type UISettings struct {
	SelectedFilesPanel SelectedFilesPanelSettings `json:"selected_files_panel"`
}

// SelectedFilesPanelSettings configures the behavior of the selected files panel
type SelectedFilesPanelSettings struct {
	RemovalKeys    []string `json:"removal_keys"`    // Keys that remove selected files
	ShowHelpText   bool     `json:"show_help_text"`  // Whether to show help text
	HelpText       string   `json:"help_text"`       // Custom help text format
	ConfirmRemoval bool     `json:"confirm_removal"` // Whether to confirm before removing files
}
