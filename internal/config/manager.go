package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	AppName    = "prompter"
	ConfigName = "config.json"
	AppVersion = "0.1.0" // This should be updated with the actual app version
)

// ConfigManager handles loading and saving the application configuration.
type ConfigManager struct {
	configPath string
	config     *AppConfig
	mutex      sync.RWMutex
}

// NewManager creates a new ConfigManager.
func NewManager() (*ConfigManager, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(cfgDir, AppName, ConfigName)

	m := &ConfigManager{
		configPath: configPath,
	}

	err = m.load()
	if err != nil {
		return nil, err
	}

	return m, nil
}

// load reads the configuration from disk, or creates a default one.
func (m *ConfigManager) load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.config = newDefaultConfig()
		return m.save()
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// If unmarshalling fails, create a new default config
		m.config = newDefaultConfig()
		return m.save()
	}
	m.config = &cfg
	if m.config.RecentWorkspaces == nil {
		m.config.RecentWorkspaces = make(map[string]*WorkspaceState)
	}

	return nil
}

// save writes the current configuration to disk.
func (m *ConfigManager) save() error {
	m.config.Metadata.LastModified = time.Now()
	m.config.Metadata.AppVersion = AppVersion

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(m.configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(m.configPath, data, 0644)
}

// Save saves the configuration. It's a thread-safe wrapper around save().
func (m *ConfigManager) Save() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.save()
}

// GetWorkspace returns the state for a given workspace path.
func (m *ConfigManager) GetWorkspace(path string) *WorkspaceState {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ws, ok := m.config.RecentWorkspaces[path]
	if !ok {
		ws = &WorkspaceState{
			Path:          path,
			SelectedFiles: []string{},
		}
		m.config.RecentWorkspaces[path] = ws
	}
	ws.LastAccessed = time.Now()
	// Save the workspace immediately to persist the new workspace or updated LastAccessed
	m.save()
	return ws
}

// newDefaultConfig creates a new AppConfig with default values.
func newDefaultConfig() *AppConfig {
	return &AppConfig{
		RecentWorkspaces: make(map[string]*WorkspaceState),
		Metadata: ConfigMetadata{
			Version:      "1",
			AppVersion:   AppVersion,
			CreatedAt:    time.Now(),
			LastModified: time.Now(),
		},
	}
}
