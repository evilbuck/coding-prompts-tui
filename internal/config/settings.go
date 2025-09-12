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
	Bindings KeyBindings    `toml:"bindings"`
	UI       UserUISettings `toml:"ui"`
	Debug    DebugSettings  `toml:"debug"`
}

// KeyBindings contains all key binding configurations
type KeyBindings struct {
	// Global bindings (always active)
	EscapeToNormal string `toml:"escape_to_normal"`

	// Mode-specific bindings
	MenuMode   ModeBindings `toml:"menu_mode"`
	NormalMode ModeBindings `toml:"normal_mode"`

	// TODO:: remove this. there isn't any legacy applications in the wild
	// Deprecated: Legacy single-character bindings for backward compatibility
	MenuActivation string `toml:"menu_activation,omitempty"`
	PersonaMenu    string `toml:"persona_menu,omitempty"`
}

// ModeBindings represents key bindings for a specific interaction mode
type ModeBindings struct {
	Activation  string `toml:"activation,omitempty"`
	Exit        string `toml:"exit,omitempty"`
	PersonaMenu string `toml:"persona_menu,omitempty"`
	Tab         string `toml:"tab,omitempty"`
	ShiftTab    string `toml:"shift_tab,omitempty"`
}

// UserUISettings represents user interface configuration options from TOML
type UserUISettings struct {
	NotificationTTL int `toml:"notification_ttl"`
}

// DebugSettings represents debug configuration options from TOML
type DebugSettings struct {
	Enabled     bool   `toml:"enabled"`      // Enable debug mode on startup
	ToggleKey   string `toml:"toggle_key"`   // Key binding to toggle debug mode
	FileLogging bool   `toml:"file_logging"` // Enable file logging for debug messages
	LogFile     string `toml:"log_file"`     // Log file path relative to workspace
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

	// Apply defaults for missing values
	m.applyDefaults(&settings)

	// Validate the loaded settings
	if err := m.validate(&settings); err != nil {
		return fmt.Errorf("invalid configuration in %s: %w", m.configPath, err)
	}

	m.settings = &settings
	return nil
}

// applyDefaults applies default values for any missing configuration
func (m *SettingsManager) applyDefaults(settings *UserSettings) {
	defaults := getDefaultSettings()

	// Apply binding defaults
	if settings.Bindings.EscapeToNormal == "" {
		settings.Bindings.EscapeToNormal = defaults.Bindings.EscapeToNormal
	}

	// Apply menu mode defaults
	if settings.Bindings.MenuMode.Activation == "" {
		settings.Bindings.MenuMode.Activation = defaults.Bindings.MenuMode.Activation
	}
	if settings.Bindings.MenuMode.Exit == "" {
		settings.Bindings.MenuMode.Exit = defaults.Bindings.MenuMode.Exit
	}
	if settings.Bindings.MenuMode.PersonaMenu == "" {
		settings.Bindings.MenuMode.PersonaMenu = defaults.Bindings.MenuMode.PersonaMenu
	}

	// Apply normal mode defaults
	if settings.Bindings.NormalMode.Tab == "" {
		settings.Bindings.NormalMode.Tab = defaults.Bindings.NormalMode.Tab
	}
	if settings.Bindings.NormalMode.ShiftTab == "" {
		settings.Bindings.NormalMode.ShiftTab = defaults.Bindings.NormalMode.ShiftTab
	}

	// Apply UI defaults
	if settings.UI.NotificationTTL <= 0 {
		settings.UI.NotificationTTL = defaults.UI.NotificationTTL
	}

	// Apply debug defaults
	if settings.Debug.ToggleKey == "" {
		settings.Debug.ToggleKey = defaults.Debug.ToggleKey
	}
	if settings.Debug.LogFile == "" {
		settings.Debug.LogFile = defaults.Debug.LogFile
	}
	// Note: FileLogging defaults to false (zero value), so we don't override it
	// Note: Enabled defaults to false (zero value), so we don't override it
}

// validate performs validation on the loaded settings
func (m *SettingsManager) validate(settings *UserSettings) error {
	// Check for backward compatibility mode (legacy single-character bindings)
	if settings.Bindings.MenuActivation != "" || settings.Bindings.PersonaMenu != "" {
		return m.validateLegacyBindings(settings)
	}

	// Validate new mode-based bindings
	return m.validateModeBindings(settings)
}

// validateLegacyBindings validates the old single-character binding format
func (m *SettingsManager) validateLegacyBindings(settings *UserSettings) error {
	// Apply legacy defaults if missing
	if settings.Bindings.MenuActivation == "" {
		settings.Bindings.MenuActivation = "x" // Default legacy menu activation
	}
	if settings.Bindings.PersonaMenu == "" {
		settings.Bindings.PersonaMenu = "p" // Default legacy persona menu
	}

	if len(settings.Bindings.MenuActivation) != 1 {
		return fmt.Errorf("bindings.menu_activation must be a single character, got: %q", settings.Bindings.MenuActivation)
	}

	if len(settings.Bindings.PersonaMenu) != 1 {
		return fmt.Errorf("bindings.persona_menu must be a single character, got: %q", settings.Bindings.PersonaMenu)
	}

	return nil
}

// validateModeBindings validates the new mode-based binding format
func (m *SettingsManager) validateModeBindings(settings *UserSettings) error {
	// Validate menu mode activation key
	if settings.Bindings.MenuMode.Activation == "" {
		return fmt.Errorf("bindings.menu_mode.activation cannot be empty")
	}

	if err := validateKeyBinding(settings.Bindings.MenuMode.Activation); err != nil {
		return fmt.Errorf("invalid bindings.menu_mode.activation: %w", err)
	}

	// Validate menu mode exit key
	if settings.Bindings.MenuMode.Exit != "" {
		if err := validateKeyBinding(settings.Bindings.MenuMode.Exit); err != nil {
			return fmt.Errorf("invalid bindings.menu_mode.exit: %w", err)
		}
	}

	// Validate persona menu key (if specified)
	if settings.Bindings.MenuMode.PersonaMenu != "" {
		if err := validateKeyBinding(settings.Bindings.MenuMode.PersonaMenu); err != nil {
			return fmt.Errorf("invalid bindings.menu_mode.persona_menu: %w", err)
		}
	}

	return nil
}

// validateKeyBinding validates a key binding string (supports modifier combinations)
func validateKeyBinding(binding string) error {
	if binding == "" {
		return fmt.Errorf("key binding cannot be empty")
	}

	// Parse the key binding to validate its format
	_, err := ParseKeyBinding(binding)
	if err != nil {
		return fmt.Errorf("invalid key binding format: %w", err)
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
// For backward compatibility, returns legacy binding if present, otherwise new mode-based binding
func (m *SettingsManager) GetMenuActivationKey() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check for legacy binding first
	if m.settings.Bindings.MenuActivation != "" {
		return m.settings.Bindings.MenuActivation
	}

	// Return new mode-based activation key
	return m.settings.Bindings.MenuMode.Activation
}

// GetPersonaMenuKey returns the persona menu key binding (thread-safe)
// For backward compatibility, returns legacy binding if present, otherwise new mode-based binding
func (m *SettingsManager) GetPersonaMenuKey() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Check for legacy binding first
	if m.settings.Bindings.PersonaMenu != "" {
		return m.settings.Bindings.PersonaMenu
	}

	// Return new mode-based persona menu key
	return m.settings.Bindings.MenuMode.PersonaMenu
}

// GetMenuModeActivation returns the menu mode activation key (thread-safe)
func (m *SettingsManager) GetMenuModeActivation() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings.Bindings.MenuMode.Activation
}

// GetMenuModeExit returns the menu mode exit key (thread-safe)
func (m *SettingsManager) GetMenuModeExit() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings.Bindings.MenuMode.Exit
}

// GetMenuModePersonaMenu returns the persona menu key for menu mode (thread-safe)
func (m *SettingsManager) GetMenuModePersonaMenu() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings.Bindings.MenuMode.PersonaMenu
}

// IsLegacyMode returns true if using legacy single-character bindings
func (m *SettingsManager) IsLegacyMode() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings.Bindings.MenuActivation != "" || m.settings.Bindings.PersonaMenu != ""
}

// GetNotificationTTL returns the notification time-to-live in seconds
func (m *SettingsManager) GetNotificationTTL() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.settings.UI.NotificationTTL <= 0 {
		return 3 // Default 3 seconds
	}
	return m.settings.UI.NotificationTTL
}

// Debug settings accessors

// IsDebugEnabled returns whether debug mode should be enabled on startup
func (m *SettingsManager) IsDebugEnabled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings.Debug.Enabled
}

// GetDebugToggleKey returns the key binding for toggling debug mode
func (m *SettingsManager) GetDebugToggleKey() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.settings.Debug.ToggleKey == "" {
		return "f11" // Default F11
	}
	return m.settings.Debug.ToggleKey
}

// IsDebugFileLoggingEnabled returns whether file logging should be enabled for debug
func (m *SettingsManager) IsDebugFileLoggingEnabled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.settings.Debug.FileLogging
}

// GetDebugLogFile returns the log file path for debug messages
func (m *SettingsManager) GetDebugLogFile() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	if m.settings.Debug.LogFile == "" {
		return "logs/error.log" // Default path
	}
	return m.settings.Debug.LogFile
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
	if err := os.MkdirAll(configDir, 0o755); err != nil {
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
	if onChange != nil && (m.hasBindingsChanged(&oldSettings.Bindings, &newSettings.Bindings) ||
		m.hasUIChanged(&oldSettings.UI, &newSettings.UI) ||
		m.hasDebugChanged(&oldSettings.Debug, &newSettings.Debug)) {
		onChange(newSettings)
	}

	return nil
}

// hasBindingsChanged checks if any key bindings have changed
func (m *SettingsManager) hasBindingsChanged(old, new *KeyBindings) bool {
	// Check legacy bindings
	if old.MenuActivation != new.MenuActivation || old.PersonaMenu != new.PersonaMenu {
		return true
	}

	// Check global bindings
	if old.EscapeToNormal != new.EscapeToNormal {
		return true
	}

	// Check menu mode bindings
	if old.MenuMode.Activation != new.MenuMode.Activation ||
		old.MenuMode.Exit != new.MenuMode.Exit ||
		old.MenuMode.PersonaMenu != new.MenuMode.PersonaMenu {
		return true
	}

	// Check normal mode bindings
	if old.NormalMode.Tab != new.NormalMode.Tab ||
		old.NormalMode.ShiftTab != new.NormalMode.ShiftTab {
		return true
	}

	return false
}

// hasUIChanged checks if any UI settings have changed
func (m *SettingsManager) hasUIChanged(old, new *UserUISettings) bool {
	return old.NotificationTTL != new.NotificationTTL
}

// hasDebugChanged checks if any debug settings have changed
func (m *SettingsManager) hasDebugChanged(old, new *DebugSettings) bool {
	return old.Enabled != new.Enabled ||
		old.ToggleKey != new.ToggleKey ||
		old.FileLogging != new.FileLogging ||
		old.LogFile != new.LogFile
}

// getDefaultSettings returns the default settings
func getDefaultSettings() *UserSettings {
	return &UserSettings{
		Bindings: KeyBindings{
			EscapeToNormal: "esc",
			MenuMode: ModeBindings{
				Activation:  "alt+m",
				Exit:        "esc",
				PersonaMenu: "p",
			},
			NormalMode: ModeBindings{
				Tab:      "tab",
				ShiftTab: "shift+tab",
			},
			// Legacy defaults for backward compatibility
			MenuActivation: "",
			PersonaMenu:    "",
		},
		UI: UserUISettings{
			NotificationTTL: 3, // Default 3 seconds
		},
		Debug: DebugSettings{
			Enabled:     false,            // Debug disabled by default
			ToggleKey:   "f11",            // F11 to toggle
			FileLogging: true,             // Enable file logging when debug is on
			LogFile:     "logs/error.log", // Default log file path
		},
	}
}

