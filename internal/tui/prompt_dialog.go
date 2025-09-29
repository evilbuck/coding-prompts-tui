package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PromptDialogModel represents the scrollable prompt dialog
type PromptDialogModel struct {
	viewport viewport.Model
	width    int
	height   int
	content  string
	visible  bool
}

// NewPromptDialogModel creates a new prompt dialog model
func NewPromptDialogModel() *PromptDialogModel {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.DefaultKeyMap()

	return &PromptDialogModel{
		viewport: vp,
		visible:  false,
	}
}

// SetSize updates the dialog dimensions
func (m *PromptDialogModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate dialog dimensions (80% of screen)
	dialogWidth := int(float64(width) * 0.8)
	dialogHeight := int(float64(height) * 0.8)

	// Calculate viewport dimensions (minus borders and padding)
	viewportWidth := dialogWidth - 4
	viewportHeight := dialogHeight - 4

	m.viewport.Width = viewportWidth
	m.viewport.Height = viewportHeight
}

// Show displays the dialog with the given content
func (m *PromptDialogModel) Show(content string) {
	m.content = content
	m.visible = true

	// Word wrap content to fit viewport width
	wrappedContent := lipgloss.NewStyle().Width(m.viewport.Width).Render(content)
	m.viewport.SetContent(wrappedContent)

	// Reset scroll position to top
	m.viewport.GotoTop()
}

// Hide closes the dialog
func (m *PromptDialogModel) Hide() {
	m.visible = false
}

// IsVisible returns whether the dialog is currently shown
func (m *PromptDialogModel) IsVisible() bool {
	return m.visible
}

// GetContent returns the current prompt content
func (m *PromptDialogModel) GetContent() string {
	return m.content
}

// Update handles messages for the prompt dialog
func (m *PromptDialogModel) Update(msg tea.Msg) (*PromptDialogModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "enter", "esc":
			m.Hide()
			return m, nil
		}

		// Pass scroll controls to viewport
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the prompt dialog
func (m *PromptDialogModel) View() string {
	if !m.visible {
		return ""
	}

	// Calculate dialog dimensions
	dialogWidth := int(float64(m.width) * 0.8)
	dialogHeight := int(float64(m.height) * 0.8)

	// Create dialog style
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("228")).
		Padding(1, 2).
		Width(dialogWidth).
		Height(dialogHeight)

	// Render the scrollable content
	content := m.viewport.View()

	// Add scroll indicators if content is scrollable
	if m.viewport.TotalLineCount() > m.viewport.Height {
		scrollPercent := m.viewport.ScrollPercent()
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render(fmt.Sprintf("%.0f%%", scrollPercent*100))

		// Add scroll percentage to the bottom right of content
		contentLines := strings.Split(content, "\n")
		if len(contentLines) > 0 {
			lastLineIdx := len(contentLines) - 1
			lastLine := contentLines[lastLineIdx]

			// Pad the last line and add scroll info
			padding := m.viewport.Width - lipgloss.Width(lastLine) - lipgloss.Width(scrollInfo)
			if padding > 0 {
				contentLines[lastLineIdx] = lastLine + strings.Repeat(" ", padding) + scrollInfo
			}
		}
		content = strings.Join(contentLines, "\n")
	}

	dialog := dialogStyle.Render(content)

	// Center the dialog on screen
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("237")),
	)
}
