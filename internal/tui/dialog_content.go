package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DialogContent represents the content that can be displayed in a dialog
type DialogContent interface {
	// Render returns the string representation of the content
	Render() string

	// Update handles input messages and returns a command
	Update(msg tea.Msg) tea.Cmd

	// Init initializes the content and returns a command
	Init() tea.Cmd

	// SetSize sets the available size for the content
	SetSize(width, height int)

	// OnShow is called when the dialog becomes visible
	OnShow()

	// OnHide is called when the dialog is hidden
	OnHide()
}

// ListItem represents an item in a list dialog
type ListItem struct {
	Label    string
	Value    interface{}
	Selected bool
}

// ListSelectMsg is sent when items are selected in a list dialog
type ListSelectMsg struct {
	Items []ListItem
}

// ListContent provides an interactive list for dialog content
type ListContent struct {
	items       []ListItem
	cursor      int
	multiSelect bool
	onSelect    func([]ListItem) tea.Msg // Callback when selection is made
	width       int
	height      int
}

// NewListContent creates a new list content with the given items
func NewListContent(items []ListItem, multiSelect bool, onSelect func([]ListItem) tea.Msg) *ListContent {
	return &ListContent{
		items:       items,
		cursor:      0,
		multiSelect: multiSelect,
		onSelect:    onSelect,
	}
}

// Render returns the string representation of the list
func (lc *ListContent) Render() string {
	var content strings.Builder

	// Render list items
	for i, item := range lc.items {
		cursor := " "
		if i == lc.cursor {
			cursor = "▶"
		}

		var checkbox string
		if lc.multiSelect {
			if item.Selected {
				checkbox = "☑ "
			} else {
				checkbox = "☐ "
			}
		} else {
			if item.Selected {
				checkbox = "● "
			} else {
				checkbox = "○ "
			}
		}

		line := fmt.Sprintf("%s %s%s", cursor, checkbox, item.Label)

		// Highlight current selection
		if i == lc.cursor {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("69")).
				Foreground(lipgloss.Color("0")).
				Render(" " + line + " ")
		} else {
			line = " " + line + " "
		}

		content.WriteString(line + "\n")
	}

	// Add help text based on selection mode
	content.WriteString("\n")
	if lc.multiSelect {
		content.WriteString("↑/↓: Navigate • Space: Toggle • Enter: Apply")
	} else {
		content.WriteString("↑/↓: Navigate • Enter: Select")
	}

	return content.String()
}

// Update handles input for the list content
func (lc *ListContent) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if lc.cursor > 0 {
				lc.cursor--
			} else {
				lc.cursor = len(lc.items) - 1
			}
		case "down", "j":
			if lc.cursor < len(lc.items)-1 {
				lc.cursor++
			} else {
				lc.cursor = 0
			}
		case " ":
			if lc.multiSelect && lc.cursor >= 0 && lc.cursor < len(lc.items) {
				lc.items[lc.cursor].Selected = !lc.items[lc.cursor].Selected
			}
		case "enter":
			if !lc.multiSelect && lc.cursor >= 0 && lc.cursor < len(lc.items) {
				// For single select, set only the current item as selected
				for i := range lc.items {
					lc.items[i].Selected = false
				}
				lc.items[lc.cursor].Selected = true
			}

			// Call the selection callback
			if lc.onSelect != nil {
				return func() tea.Msg {
					return lc.onSelect(lc.items)
				}
			}
		}
	}
	return nil
}

// Init initializes the list content
func (lc *ListContent) Init() tea.Cmd {
	return nil
}

// SetSize sets the available size for the list content
func (lc *ListContent) SetSize(width, height int) {
	lc.width = width
	lc.height = height
}

// OnShow is called when the dialog becomes visible
func (lc *ListContent) OnShow() {
	// No special action needed for list content
}

// OnHide is called when the dialog is hidden
func (lc *ListContent) OnHide() {
	// No special action needed for list content
}

// SetItems updates the list items
func (lc *ListContent) SetItems(items []ListItem) {
	lc.items = items
	if lc.cursor >= len(lc.items) {
		lc.cursor = 0
	}
}

// GetSelectedItems returns the currently selected items
func (lc *ListContent) GetSelectedItems() []ListItem {
	var selected []ListItem
	for _, item := range lc.items {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// TextContent provides scrollable text display for dialog content
type TextContent struct {
	text     string
	viewport viewport.Model
	width    int
	height   int
}

// NewTextContent creates a new text content with the given text
func NewTextContent(text string) *TextContent {
	vp := viewport.New(0, 0)
	vp.KeyMap = viewport.DefaultKeyMap()

	return &TextContent{
		text:     text,
		viewport: vp,
	}
}

// Render returns the string representation of the text content
func (tc *TextContent) Render() string {
	content := tc.viewport.View()

	// Add scroll indicators if content is scrollable
	if tc.viewport.TotalLineCount() > tc.viewport.Height {
		scrollPercent := tc.viewport.ScrollPercent()
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Render(fmt.Sprintf("%.0f%%", scrollPercent*100))

		// Add scroll percentage to the bottom right of content
		contentLines := strings.Split(content, "\n")
		if len(contentLines) > 0 {
			lastLineIdx := len(contentLines) - 1
			lastLine := contentLines[lastLineIdx]

			// Pad the last line and add scroll info
			padding := tc.viewport.Width - lipgloss.Width(lastLine) - lipgloss.Width(scrollInfo)
			if padding > 0 {
				contentLines[lastLineIdx] = lastLine + strings.Repeat(" ", padding) + scrollInfo
			}
		}
		content = strings.Join(contentLines, "\n")
	}

	return content
}

// Update handles input for the text content (scrolling)
func (tc *TextContent) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	tc.viewport, cmd = tc.viewport.Update(msg)
	return cmd
}

// Init initializes the text content
func (tc *TextContent) Init() tea.Cmd {
	return nil
}

// SetSize sets the available size for the text content
func (tc *TextContent) SetSize(width, height int) {
	tc.width = width
	tc.height = height
	tc.viewport.Width = width
	tc.viewport.Height = height

	// Update viewport content with word wrapping
	if tc.text != "" {
		wrappedText := lipgloss.NewStyle().Width(width).Render(tc.text)
		tc.viewport.SetContent(wrappedText)
	}
}

// OnShow is called when the dialog becomes visible
func (tc *TextContent) OnShow() {
	// Reset scroll position to top when shown
	tc.viewport.GotoTop()
}

// OnHide is called when the dialog is hidden
func (tc *TextContent) OnHide() {
	// No special action needed for text content
}

// SetText updates the text content
func (tc *TextContent) SetText(text string) {
	tc.text = text
	if tc.width > 0 {
		wrappedText := lipgloss.NewStyle().Width(tc.width).Render(text)
		tc.viewport.SetContent(wrappedText)
	}
}

// GetText returns the current text content
func (tc *TextContent) GetText() string {
	return tc.text
}