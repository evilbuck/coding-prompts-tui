package tui

import "github.com/charmbracelet/lipgloss"

// LayoutConfig holds centralized layout configuration
type LayoutConfig struct {
	HeaderHeight       int
	FooterHeight       int
	BorderCompensation int
	TopHeightRatio     float64
}

// NewLayoutConfig creates a default layout configuration
func NewLayoutConfig() *LayoutConfig {
	return &LayoutConfig{
		HeaderHeight:       3,
		FooterHeight:       3,
		BorderCompensation: 2, // 1 pixel border on each side
		TopHeightRatio:     0.66,
	}
}

// AvailableHeight calculates the height available for main content panels
func (lc *LayoutConfig) AvailableHeight(totalHeight int) int {
	return totalHeight - lc.HeaderHeight - lc.FooterHeight
}

// TopPanelHeight calculates the height for top panels
func (lc *LayoutConfig) TopPanelHeight(totalHeight int) int {
	availableHeight := lc.AvailableHeight(totalHeight)
	return int(float64(availableHeight) * lc.TopHeightRatio)
}

// BottomPanelHeight calculates the height for bottom panels
func (lc *LayoutConfig) BottomPanelHeight(totalHeight int) int {
	availableHeight := lc.AvailableHeight(totalHeight)
	topHeight := lc.TopPanelHeight(totalHeight)
	return availableHeight - topHeight
}

func (lc *LayoutConfig) CalcPanelWidth(totalWidth int, percentage float64) int {
	return int(float64(totalWidth) * percentage / 100)
}

// LeftPanelWidth calculates the width for left panels (50% split)
func (lc *LayoutConfig) LeftPanelWidth(totalWidth int) int {
	return lc.CalcPanelWidth(totalWidth, 30)
}

// RightPanelWidth calculates the width for right panels
func (lc *LayoutConfig) RightPanelWidth(totalWidth int) int {
	leftWidth := lc.LeftPanelWidth(totalWidth)
	return totalWidth - leftWidth
}

// StretchWidth calculates width that stretches to container, optionally accounting for borders
func StretchWidth(containerWidth int, accountForBorders bool) int {
	if accountForBorders {
		return containerWidth - 2 // Standard border compensation
	}
	return containerWidth
}

// StretchHeight calculates height that stretches to container, optionally accounting for borders
func StretchHeight(containerHeight int, accountForBorders bool) int {
	if accountForBorders {
		return containerHeight - 2 // Standard border compensation
	}
	return containerHeight
}

// PanelStyle applies width and height to a base style
func PanelStyle(base lipgloss.Style, width, height int) lipgloss.Style {
	return base.Width(width).Height(height)
}

// FocusStyle returns the appropriate style based on focus state
func FocusStyle(focused bool, normalStyle, focusedStyle lipgloss.Style) lipgloss.Style {
	if focused {
		return focusedStyle
	}
	return normalStyle
}

// CreatePanel is a convenience function that combines style selection and sizing
func CreatePanel(content string, focused bool, normalStyle, focusedStyle lipgloss.Style, width, height int) string {
	style := FocusStyle(focused, normalStyle, focusedStyle)
	return PanelStyle(style, width, height).Render(content)
}

