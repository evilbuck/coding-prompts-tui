package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
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
	b.WriteString(helpStyle.Render("Enter your prompt below. Ctrl+S to generate XML prompt, Ctrl+Y to copy"))
	b.WriteString("\n\n")

	// Textarea
	b.WriteString(m.textarea.View())

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
