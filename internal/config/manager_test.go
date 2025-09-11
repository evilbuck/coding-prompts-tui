package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigManagerSaveRestore(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "prompter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary config path
	configPath := filepath.Join(tmpDir, "config.json")

	// Create manager with custom config path
	manager := &ConfigManager{
		configPath: configPath,
	}

	// Load initial config (should create default)
	err = manager.load()
	if err != nil {
		t.Fatalf("Failed to load initial config: %v", err)
	}

	// Test workspace creation and retrieval
	testPath := "/test/workspace"
	workspace := manager.GetWorkspace(testPath)

	if workspace.Path != testPath {
		t.Errorf("Expected workspace path %s, got %s", testPath, workspace.Path)
	}

	// Modify workspace data
	workspace.SelectedFiles = []string{"file1.go", "file2.go"}
	workspace.ChatInput = "Test prompt content"
	
	// Save the changes
	err = manager.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Create a new manager to test loading from disk
	manager2 := &ConfigManager{
		configPath: configPath,
	}
	err = manager2.load()
	if err != nil {
		t.Fatalf("Failed to load config from disk: %v", err)
	}

	// Retrieve the same workspace
	workspace2 := manager2.GetWorkspace(testPath)

	// Verify data was restored correctly
	if len(workspace2.SelectedFiles) != 2 {
		t.Errorf("Expected 2 selected files, got %d", len(workspace2.SelectedFiles))
	}
	if workspace2.SelectedFiles[0] != "file1.go" || workspace2.SelectedFiles[1] != "file2.go" {
		t.Errorf("Selected files not restored correctly: %v", workspace2.SelectedFiles)
	}
	if workspace2.ChatInput != "Test prompt content" {
		t.Errorf("Expected chat input 'Test prompt content', got '%s'", workspace2.ChatInput)
	}
}

func TestConfigManagerMultipleWorkspaces(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "prompter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manager
	manager := &ConfigManager{
		configPath: filepath.Join(tmpDir, "config.json"),
	}
	
	err = manager.load()
	if err != nil {
		t.Fatal(err)
	}

	// Create multiple workspaces
	ws1 := manager.GetWorkspace("/workspace1")
	ws1.SelectedFiles = []string{"file1.go"}
	ws1.ChatInput = "Prompt 1"

	ws2 := manager.GetWorkspace("/workspace2")
	ws2.SelectedFiles = []string{"file2.go", "file3.go"}
	ws2.ChatInput = "Prompt 2"

	// Save
	err = manager.Save()
	if err != nil {
		t.Fatal(err)
	}

	// Create new manager and verify both workspaces exist
	manager2 := &ConfigManager{
		configPath: filepath.Join(tmpDir, "config.json"),
	}
	err = manager2.load()
	if err != nil {
		t.Fatal(err)
	}

	// Check workspace 1
	ws1_restored := manager2.GetWorkspace("/workspace1")
	if ws1_restored.ChatInput != "Prompt 1" {
		t.Errorf("Workspace 1 chat input not restored correctly")
	}
	if len(ws1_restored.SelectedFiles) != 1 || ws1_restored.SelectedFiles[0] != "file1.go" {
		t.Errorf("Workspace 1 selected files not restored correctly")
	}

	// Check workspace 2
	ws2_restored := manager2.GetWorkspace("/workspace2")
	if ws2_restored.ChatInput != "Prompt 2" {
		t.Errorf("Workspace 2 chat input not restored correctly")
	}
	if len(ws2_restored.SelectedFiles) != 2 {
		t.Errorf("Workspace 2 should have 2 selected files, got %d", len(ws2_restored.SelectedFiles))
	}
}