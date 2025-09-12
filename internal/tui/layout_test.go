package tui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestLayoutConfig_Calculations(t *testing.T) {
	config := NewLayoutConfig()
	
	// Test with known dimensions
	totalWidth := 100
	totalHeight := 50
	
	// Test available height calculation
	expectedAvailableHeight := totalHeight - config.HeaderHeight - config.FooterHeight
	availableHeight := config.AvailableHeight(totalHeight)
	if availableHeight != expectedAvailableHeight {
		t.Errorf("AvailableHeight: expected %d, got %d", expectedAvailableHeight, availableHeight)
	}
	
	// Test top panel height calculation
	expectedTopHeight := int(float64(expectedAvailableHeight) * config.TopHeightRatio)
	topHeight := config.TopPanelHeight(totalHeight)
	if topHeight != expectedTopHeight {
		t.Errorf("TopPanelHeight: expected %d, got %d", expectedTopHeight, topHeight)
	}
	
	// Test bottom panel height calculation
	expectedBottomHeight := expectedAvailableHeight - expectedTopHeight
	bottomHeight := config.BottomPanelHeight(totalHeight)
	if bottomHeight != expectedBottomHeight {
		t.Errorf("BottomPanelHeight: expected %d, got %d", expectedBottomHeight, bottomHeight)
	}
	
	// Test left panel width calculation
	expectedLeftWidth := totalWidth / 2
	leftWidth := config.LeftPanelWidth(totalWidth)
	if leftWidth != expectedLeftWidth {
		t.Errorf("LeftPanelWidth: expected %d, got %d", expectedLeftWidth, leftWidth)
	}
	
	// Test right panel width calculation
	expectedRightWidth := totalWidth - expectedLeftWidth
	rightWidth := config.RightPanelWidth(totalWidth)
	if rightWidth != expectedRightWidth {
		t.Errorf("RightPanelWidth: expected %d, got %d", expectedRightWidth, rightWidth)
	}
}

func TestStretchWidth(t *testing.T) {
	containerWidth := 100
	
	// Test without border compensation
	width := StretchWidth(containerWidth, false)
	if width != containerWidth {
		t.Errorf("StretchWidth without borders: expected %d, got %d", containerWidth, width)
	}
	
	// Test with border compensation
	expectedWidth := containerWidth - 2
	width = StretchWidth(containerWidth, true)
	if width != expectedWidth {
		t.Errorf("StretchWidth with borders: expected %d, got %d", expectedWidth, width)
	}
}

func TestStretchHeight(t *testing.T) {
	containerHeight := 50
	
	// Test without border compensation
	height := StretchHeight(containerHeight, false)
	if height != containerHeight {
		t.Errorf("StretchHeight without borders: expected %d, got %d", containerHeight, height)
	}
	
	// Test with border compensation
	expectedHeight := containerHeight - 2
	height = StretchHeight(containerHeight, true)
	if height != expectedHeight {
		t.Errorf("StretchHeight with borders: expected %d, got %d", expectedHeight, height)
	}
}

func TestFocusStyle(t *testing.T) {
	normalStyle := lipgloss.NewStyle().Background(lipgloss.Color("gray"))
	focusedStyle := lipgloss.NewStyle().Background(lipgloss.Color("blue"))
	
	// Test focused state - verify it returns a style
	result := FocusStyle(true, normalStyle, focusedStyle)
	_ = result // Just verify it doesn't panic and returns something
	
	// Test normal state - verify it returns a style
	result = FocusStyle(false, normalStyle, focusedStyle)
	_ = result // Just verify it doesn't panic and returns something
}

func TestPanelStyle(t *testing.T) {
	baseStyle := lipgloss.NewStyle().Background(lipgloss.Color("gray"))
	width := 50
	height := 25
	
	result := PanelStyle(baseStyle, width, height)
	
	// Verify the style has the correct width and height applied
	// Note: We can't directly inspect the lipgloss style properties,
	// but we can verify the function doesn't panic and returns a style
	if result.GetWidth() != width {
		t.Errorf("PanelStyle width: expected %d, got %d", width, result.GetWidth())
	}
	
	if result.GetHeight() != height {
		t.Errorf("PanelStyle height: expected %d, got %d", height, result.GetHeight())
	}
}

func TestCreatePanel(t *testing.T) {
	content := "test content"
	normalStyle := lipgloss.NewStyle().Background(lipgloss.Color("gray"))
	focusedStyle := lipgloss.NewStyle().Background(lipgloss.Color("blue"))
	width := 50
	height := 25
	
	// Test focused panel
	result := CreatePanel(content, true, normalStyle, focusedStyle, width, height)
	if result == "" {
		t.Error("CreatePanel should return rendered content")
	}
	
	// Test normal panel
	result = CreatePanel(content, false, normalStyle, focusedStyle, width, height)
	if result == "" {
		t.Error("CreatePanel should return rendered content")
	}
}