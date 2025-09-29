package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettingsManager_Load_DefaultSettings(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	manager := &SettingsManager{
		configPath: configPath,
	}

	err := manager.load()
	if err != nil {
		t.Fatalf("Expected no error loading default settings, got: %v", err)
	}

	// In the new default settings, legacy bindings should be empty
	if manager.settings.Bindings.MenuActivation != "" {
		t.Errorf("Expected default menu_activation to be empty (new mode), got: %q", manager.settings.Bindings.MenuActivation)
	}

	// Check new mode-based default
	if manager.settings.Bindings.MenuMode.Activation != "alt+m" {
		t.Errorf("Expected default menu_mode.activation to be 'alt+m', got: %q", manager.settings.Bindings.MenuMode.Activation)
	}

	// Check UI settings default
	if manager.settings.UI.NotificationTTL != 3 {
		t.Errorf("Expected default notification_ttl to be 3, got: %d", manager.settings.UI.NotificationTTL)
	}
}

func TestSettingsManager_Load_ValidTOML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	// Create a valid TOML file
	validTOML := `[bindings]
menu_activation = "m"
persona_menu = "p"`

	err := os.WriteFile(configPath, []byte(validTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	manager := &SettingsManager{
		configPath: configPath,
	}

	err = manager.load()
	if err != nil {
		t.Fatalf("Expected no error loading valid TOML, got: %v", err)
	}

	if manager.settings.Bindings.MenuActivation != "m" {
		t.Errorf("Expected menu_activation to be 'm', got: %q", manager.settings.Bindings.MenuActivation)
	}
}

func TestSettingsManager_Load_InvalidTOML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	// Create an invalid TOML file
	invalidTOML := `[bindings
menu_activation = "x"`

	err := os.WriteFile(configPath, []byte(invalidTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	manager := &SettingsManager{
		configPath: configPath,
	}

	err = manager.load()
	if err == nil {
		t.Fatal("Expected error loading invalid TOML, got nil")
	}

	if !strings.Contains(err.Error(), "invalid TOML format") {
		t.Errorf("Expected error message to contain 'invalid TOML format', got: %v", err)
	}
}

func TestSettingsManager_Validate_EmptyMenuActivation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	// Create TOML with empty menu_mode.activation (new format)
	emptyTOML := `[bindings.menu_mode]
activation = ""`

	err := os.WriteFile(configPath, []byte(emptyTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	manager := &SettingsManager{
		configPath: configPath,
	}

	err = manager.load()
	if err == nil {
		t.Fatal("Expected validation error for empty menu_activation, got nil")
	}

	if !strings.Contains(err.Error(), "menu_mode.activation cannot be empty") {
		t.Errorf("Expected error message about empty menu_mode.activation, got: %v", err)
	}
}

func TestSettingsManager_Validate_MultiCharacterMenuActivation(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	// Create TOML with multi-character menu_activation
	multiCharTOML := `[bindings]
menu_activation = "ctrl-x"`

	err := os.WriteFile(configPath, []byte(multiCharTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	manager := &SettingsManager{
		configPath: configPath,
	}

	err = manager.load()
	if err == nil {
		t.Fatal("Expected validation error for multi-character menu_activation, got nil")
	}

	if !strings.Contains(err.Error(), "must be a single character") {
		t.Errorf("Expected error message about single character requirement, got: %v", err)
	}
}

func TestSettingsManager_GetMenuActivationKey(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	validTOML := `[bindings]
menu_activation = "z"
persona_menu = "p"`

	err := os.WriteFile(configPath, []byte(validTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	manager := &SettingsManager{
		configPath: configPath,
	}

	err = manager.load()
	if err != nil {
		t.Fatalf("Expected no error loading valid TOML, got: %v", err)
	}

	key := manager.GetMenuActivationKey()
	if key != "z" {
		t.Errorf("Expected GetMenuActivationKey() to return 'z', got: %q", key)
	}
}

func TestSettingsManager_FileReadError(t *testing.T) {
	// Use a path that doesn't exist or can't be read
	nonExistentPath := filepath.Join("/nonexistent/path/coding_prompts.toml")

	manager := &SettingsManager{
		configPath: nonExistentPath,
	}

	// This should succeed with default settings when file doesn't exist
	err := manager.load()
	if err != nil {
		t.Fatalf("Expected no error when file doesn't exist, got: %v", err)
	}

	// Should use new default settings (legacy field should be empty)
	if manager.settings.Bindings.MenuActivation != "" {
		t.Errorf("Expected default menu_activation to be empty (new mode), got: %q", manager.settings.Bindings.MenuActivation)
	}

	// Check new mode default
	if manager.settings.Bindings.MenuMode.Activation != "alt+m" {
		t.Errorf("Expected default menu_mode.activation 'alt+m', got: %q", manager.settings.Bindings.MenuMode.Activation)
	}
}

func TestSettingsManager_Reload(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	// Create initial config
	initialTOML := `[bindings]
menu_activation = "x"
persona_menu = "p"`

	err := os.WriteFile(configPath, []byte(initialTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	manager := &SettingsManager{
		configPath: configPath,
	}

	// Load initial config
	err = manager.load()
	if err != nil {
		t.Fatalf("Failed to load initial config: %v", err)
	}

	if manager.GetMenuActivationKey() != "x" {
		t.Errorf("Expected initial key 'x', got: %q", manager.GetMenuActivationKey())
	}

	// Update the config file
	updatedTOML := `[bindings]
menu_activation = "y"
persona_menu = "p"`

	err = os.WriteFile(configPath, []byte(updatedTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}

	// Reload should pick up the change
	err = manager.Reload()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	if manager.GetMenuActivationKey() != "y" {
		t.Errorf("Expected updated key 'y', got: %q", manager.GetMenuActivationKey())
	}
}

func TestSettingsManager_OnChangeCallback(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "coding_prompts.toml")

	// Create initial config
	initialTOML := `[bindings]
menu_activation = "a"
persona_menu = "p"`

	err := os.WriteFile(configPath, []byte(initialTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	manager := &SettingsManager{
		configPath: configPath,
	}

	err = manager.load()
	if err != nil {
		t.Fatalf("Failed to load initial config: %v", err)
	}

	// Set up callback
	callbackCalled := false
	var callbackSettings *UserSettings

	manager.SetOnChange(func(newSettings *UserSettings) {
		callbackCalled = true
		callbackSettings = newSettings
	})

	// Update the config file
	updatedTOML := `[bindings]
menu_activation = "b"
persona_menu = "q"`

	err = os.WriteFile(configPath, []byte(updatedTOML), 0644)
	if err != nil {
		t.Fatalf("Failed to write updated config: %v", err)
	}

	// Trigger reload with notification
	err = manager.reloadAndNotify()
	if err != nil {
		t.Fatalf("Failed to reload and notify: %v", err)
	}

	if !callbackCalled {
		t.Error("Expected onChange callback to be called")
	}

	if callbackSettings == nil || callbackSettings.Bindings.MenuActivation != "b" {
		t.Errorf("Expected callback settings to have menu_activation 'b', got: %v", callbackSettings)
	}
}
