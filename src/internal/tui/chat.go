package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// ChatModel represents the chat input panel
type ChatModel struct {
	title    string
	textarea textarea.Model
}

// NewChatModel creates a new chat model
func NewChatModel() *ChatModel {
	ta := textarea.New()
	ta.Placeholder = "Enter your prompt for the LLM here..."
	ta.CharLimit = 5000
	ta.SetWidth(80)
	ta.SetHeight(8)
	ta.Focus()

	return &ChatModel{
		title:    "💬 User Prompt",
		textarea: ta,
	}
}

// Init initializes the chat model
func (m *ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the chat panel
func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			// TODO: Generate and save prompt
			return m, nil
		}
	}
	
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the chat input panel
func (m *ChatModel) View() string {
	var b strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	b.WriteString(helpStyle.Render("Enter your prompt below. Ctrl+S to generate XML prompt"))
	b.WriteString("\n\n")

	// Textarea
	b.WriteString(m.textarea.View())

	// Character count
	b.WriteString("\n")
	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	charCount := len(m.textarea.Value())
	maxChars := m.textarea.CharLimit
	b.WriteString(countStyle.Render(sprintf("Characters: %d/%d", charCount, maxChars)))

	return b.String()
}

// GetPrompt returns the current user prompt text
func (m *ChatModel) GetPrompt() string {
	return m.textarea.Value()
}

// SetPrompt sets the user prompt text
func (m *ChatModel) SetPrompt(prompt string) {
	m.textarea.SetValue(prompt)
}

// Focus focuses the textarea
func (m *ChatModel) Focus() tea.Cmd {
	return m.textarea.Focus()
}

// Blur removes focus from the textarea
func (m *ChatModel) Blur() {
	m.textarea.Blur()
}

// sprintf is a helper function (we'll import fmt later)
func sprintf(format string, args ...interface{}) string {
	// Simple implementation for now
	if len(args) == 2 {
		if format == "Characters: %d/%d" {
			char := args[0].(int)
			max := args[1].(int)
			return "Characters: " + itoa(char) + "/" + itoa(max)
		}
	}
	return format
}

// Simple integer to string conversion
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	
	var result strings.Builder
	negative := i < 0
	if negative {
		i = -i
	}
	
	for i > 0 {
		result.WriteByte(byte('0' + i%10))
		i /= 10
	}
	
	if negative {
		result.WriteByte('-')
	}
	
	// Reverse the string
	s := result.String()
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	
	return string(runes)
}