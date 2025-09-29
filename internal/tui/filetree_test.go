package tui

import (
	"strings"
	"testing"

	"coding-prompts-tui/internal/filesystem"
)

func TestFileTreeHeaderCalculation(t *testing.T) {
	// Create a new file tree model
	model := NewFileTreeModel("/tmp", []string{})

	// Set a reasonable width for testing
	model.width = 50

	// Calculate header content and height
	headerContent, headerHeight := model.calculateHeaderContent()

	// Verify header content is not empty
	if headerContent == "" {
		t.Error("Header content should not be empty")
	}

	// Verify header height is reasonable (should be at least 4 lines)
	if headerHeight < 4 {
		t.Errorf("Header height should be at least 4, got %d", headerHeight)
	}

	// Count actual newlines in header content
	actualLines := strings.Count(headerContent, "\n") + 1
	if actualLines != headerHeight {
		t.Errorf("Header height calculation mismatch: calculated %d, actual %d", headerHeight, actualLines)
	}
}

func TestViewportSizing(t *testing.T) {
	model := NewFileTreeModel("/tmp", []string{})

	// Set panel dimensions
	panelWidth := 40
	panelHeight := 20
	model.SetSize(panelWidth, panelHeight)

	// Calculate expected header height
	_, headerHeight := model.calculateHeaderContent()

	// Expected viewport size should be panel size minus header
	expectedViewportHeight := panelHeight - headerHeight
	if expectedViewportHeight < 1 {
		expectedViewportHeight = 1
	}

	// Check viewport dimensions
	if model.viewport.Width != panelWidth {
		t.Errorf("Viewport width mismatch: expected %d, got %d", panelWidth, model.viewport.Width)
	}

	if model.viewport.Height != expectedViewportHeight {
		t.Errorf("Viewport height mismatch: expected %d, got %d", expectedViewportHeight, model.viewport.Height)
	}
}

func TestEnsureVisibleBounds(t *testing.T) {
	model := NewFileTreeModel("/tmp", []string{})
	model.width = 40
	model.height = 20

	// Simulate some items
	model.items = make([]filesystem.FileTreeItem, 50) // More items than can fit in viewport
	for i := range model.items {
		model.items[i] = filesystem.FileTreeItem{
			Name:  "file.txt",
			Path:  "/tmp/file.txt",
			IsDir: false,
			Level: 0,
		}
	}

	// Set viewport size
	_, headerHeight := model.calculateHeaderContent()
	model.ensureViewportSizedWithHeader(headerHeight)

	// Test cursor at beginning
	model.cursor = 0
	model.ensureVisible()
	if model.viewport.YOffset != 0 {
		t.Errorf("YOffset should be 0 when cursor is at beginning, got %d", model.viewport.YOffset)
	}

	// Test cursor at end
	model.cursor = len(model.items) - 1
	model.ensureVisible()
	expectedOffset := len(model.items) - model.viewport.Height
	if expectedOffset < 0 {
		expectedOffset = 0
	}
	if model.viewport.YOffset != expectedOffset {
		t.Errorf("YOffset should be %d when cursor at end, got %d", expectedOffset, model.viewport.YOffset)
	}

	// Test bounds checking - cursor beyond items
	model.cursor = len(model.items) + 10
	model.ensureVisible()
	if model.cursor != len(model.items)-1 {
		t.Errorf("Cursor should be clamped to %d, got %d", len(model.items)-1, model.cursor)
	}
}
