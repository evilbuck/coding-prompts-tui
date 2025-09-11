package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

const (
	SettingsDir  = "coding-prompts"
	SettingsFile = "coding_prompts.toml"
)

// UserSettings represents user-configurable settings loaded from TOML
type UserSettings struct {
	Bindings KeyBindings `toml:"bindings"`
}

// KeyBindings contains all key binding configurations
type KeyBindings struct {
	MenuActivation string `toml:"menu_activation"`
}

// SettingsManager handles loading and validation of user settings from TOML
type SettingsManager struct {
	configPath string
	settings   *UserSettings
	mutex      sync.RWMutex
	watcher    *fsnotify.Watcher
	onChange   func(*UserSettings) // Callback when settings change
}

// NewSettingsManager creates a new SettingsManager
func NewSettingsManager() (*SettingsManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	
	configPath := filepath.Join(homeDir, ".config", SettingsDir, SettingsFile)
	
	m := &SettingsManager{
		configPath: configPath,
	}
	
	if err := m.load(); err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	
	return m, nil
}

// load reads the TOML configuration file with validation (thread-safe)
func (m *SettingsManager) load() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.loadUnsafe()
}

// loadUnsafe reads the TOML configuration file with validation (not thread-safe)
func (m *SettingsManager) loadUnsafe() error {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Use default settings if file doesn't exist
		m.settings = getDefaultSettings()
		return nil
	}
	
	// Read the TOML file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", m.configPath, err)
	}
	
	var settings UserSettings
	if err := toml.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("invalid TOML format in %s: %w", m.configPath, err)
	}
	
	// Validate the loaded settings
	if err := m.validate(&settings); err != nil {
		return fmt.Errorf("invalid configuration in %s: %w", m.configPath, err)
	}
	
	m.settings = &settings
	return nil
}

// validate performs validation on the loaded settings
func (m *SettingsManager) validate(settings *UserSettings) error {
	// Validate menu activation key binding
	if settings.Bindings.MenuActivation == "" {
		return fmt.Errorf("bindings.menu_activation cannot be empty")
	}
	
	// Check if it's a valid single character
	if len(settings.Bindings.MenuActivation) != 1 {
		return fmt.Errorf("bindings.menu_activation must be a single character, got: %q", settings.Bindings.MenuActivation)
	}
	
	return nil
}

// GetSettings returns the current user settings (thread-safe)
func (m *SettingsManager) GetSettings() *UserSettings {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	// Return a copy to prevent external modification
	settingsCopy := *m.settings
	return &settingsCopy
}

// GetMenuActivationKey returns the menu activation key binding (thread-safe)
func (m *SettingsManager) GetMenuActivationKey() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings.Bindings.MenuActivation
}

// Reload reloads the configuration from disk
func (m *SettingsManager) Reload() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.loadUnsafe()
}

// SetOnChange sets a callback function that gets called when settings change
func (m *SettingsManager) SetOnChange(callback func(*UserSettings)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.onChange = callback
}

// StartWatching starts watching the config file for changes
func (m *SettingsManager) StartWatching() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.watcher != nil {
		return fmt.Errorf("already watching config file")
	}
	
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	
	// Watch the config directory (not the file directly, as file might not exist yet)
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		watcher.Close()
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	if err := watcher.Add(configDir); err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch config directory: %w", err)
	}
	
	m.watcher = watcher
	
	// Start watching in a goroutine
	go m.watchLoop()
	
	return nil
}

// StopWatching stops watching the config file
func (m *SettingsManager) StopWatching() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.watcher == nil {
		return nil
	}
	
	err := m.watcher.Close()
	m.watcher = nil
	return err
}

// watchLoop runs the file watcher loop
func (m *SettingsManager) watchLoop() {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			
			// Only care about writes to our config file
			if event.Name == m.configPath && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				if err := m.reloadAndNotify(); err != nil {
					// In a real app, you might want to log this error
					// For now, we'll silently continue
					continue
				}
			}
			
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			// Handle error - in a real app you might want to log this
			_ = err
		}
	}
}

// reloadAndNotify reloads settings and calls the onChange callback
func (m *SettingsManager) reloadAndNotify() error {
	m.mutex.Lock()
	oldSettings := *m.settings // Make a copy
	err := m.loadUnsafe()
	newSettings := m.settings
	onChange := m.onChange
	m.mutex.Unlock()
	
	if err != nil {
		return err
	}
	
	// Call onChange callback if settings actually changed
	if onChange != nil && (oldSettings.Bindings.MenuActivation != newSettings.Bindings.MenuActivation) {
		onChange(newSettings)
	}
	
	return nil
}

// getDefaultSettings returns the default settings
func getDefaultSettings() *UserSettings {
	return &UserSettings{
		Bindings: KeyBindings{
			MenuActivation: "x",
		},
	}
}