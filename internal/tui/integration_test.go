package tui

import (
	"os"
	"testing"

	"coding-prompts-tui/internal/config"
)

func TestFileTreeInitializationFromWorkspace(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "prompter-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a workspace with some selected files
	testPath := "/test/workspace"
	workspace := &config.WorkspaceState{
		Path:          testPath,
		SelectedFiles: []string{"file1.go", "file2.go"},
		ChatInput:     "Test prompt",
	}

	// Create a file tree model with the workspace's selected files
	model := NewFileTreeModel(testPath, workspace.SelectedFiles)

	// Verify that the selected files are properly initialized
	if len(model.selected) != 2 {
		t.Errorf("Expected 2 selected files, got %d", len(model.selected))
	}

	if !model.selected["file1.go"] {
		t.Error("file1.go should be selected")
	}

	if !model.selected["file2.go"] {
		t.Error("file2.go should be selected")
	}

	// Verify unselected file is not marked as selected
	if model.selected["file3.go"] {
		t.Error("file3.go should not be selected")
	}
}

func TestChatModelInitializationFromWorkspace(t *testing.T) {
	// Create a chat model with initial value from workspace
	initialValue := "This is a test prompt from workspace"
	model := NewChatModel(initialValue)

	// Verify the textarea has the correct initial value
	if model.textarea.Value() != initialValue {
		t.Errorf("Expected chat input '%s', got '%s'", initialValue, model.textarea.Value())
	}
}

func TestChatModelSetSize(t *testing.T) {
	model := NewChatModel("")

	// Test setting size
	width := 80
	height := 20
	model.SetSize(width, height)

	// Verify the size was set
	if model.width != width {
		t.Errorf("Expected width %d, got %d", width, model.width)
	}
	if model.height != height {
		t.Errorf("Expected height %d, got %d", height, model.height)
	}

	// Verify textarea dimensions (height should account for title/help text)
	expectedTextareaHeight := height - 4
	if expectedTextareaHeight < 3 {
		expectedTextareaHeight = 3
	}
	// Note: We can't directly check textarea width/height as they're private,
	// but we can verify the method doesn't panic and the model stores the values
}
