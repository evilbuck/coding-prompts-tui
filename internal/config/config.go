package config

import "time"

// AppConfig represents the complete application state
type AppConfig struct {
	RecentWorkspaces map[string]*WorkspaceState `json:"recent_workspaces"`
	Metadata         ConfigMetadata             `json:"metadata"`
}

// WorkspaceState represents a previously loaded folder and its state
type WorkspaceState struct {
	Path          string    `json:"path"`            // Absolute path to workspace
	LastAccessed  time.Time `json:"last_accessed"`   // When last opened
	SelectedFiles []string  `json:"selected_files"`  // Relative paths of selected files
	ChatInput     string    `json:"chat_input"`      // Saved chat input
}

// ConfigMetadata stores application metadata
type ConfigMetadata struct {
	Version      string    `json:"version"`       // Config schema version
	AppVersion   string    `json:"app_version"`   // App version that created this config
	CreatedAt    time.Time `json:"created_at"`
	LastModified time.Time `json:"last_modified"`
}
