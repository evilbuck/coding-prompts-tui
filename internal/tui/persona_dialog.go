package tui

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PersonaDialogModel represents the persona selection dialog
type PersonaDialogModel struct {
	promptDialog      *PromptDialogModel
	availablePersonas []string
	selectedPersonas  map[string]bool
	cursor            int
	debugLogger       *log.Logger
}

// PersonaSelectionMsg is sent when personas are selected/deselected
type PersonaSelectionMsg struct {
	ActivePersonas []string
}

// NewPersonaDialogModel creates a new persona dialog model
func NewPersonaDialogModel() *PersonaDialogModel {
	return &PersonaDialogModel{
		promptDialog:     NewPromptDialogModel(),
		selectedPersonas: make(map[string]bool),
		cursor:           0,
		debugLogger:      nil,
	}
}

// SetAvailablePersonas sets the list of available personas
func (m *PersonaDialogModel) SetAvailablePersonas(personas []string) {
	m.availablePersonas = personas
	if m.cursor >= len(m.availablePersonas) {
		m.cursor = 0
	}
}

// SetActivePersonas sets the currently active personas
func (m *PersonaDialogModel) SetActivePersonas(personas []string) {
	m.selectedPersonas = make(map[string]bool)
	for _, persona := range personas {
		m.selectedPersonas[persona] = true
	}
}

// Show displays the dialog
func (m *PersonaDialogModel) Show() {
	content := m.generateDialogContent()
	m.promptDialog.Show(content)
}

// Hide closes the dialog
func (m *PersonaDialogModel) Hide() {
	m.promptDialog.Hide()
}

// IsVisible returns whether the dialog is visible
func (m *PersonaDialogModel) IsVisible() bool {
	return m.promptDialog.IsVisible()
}

// SetSize sets the dialog size for centering
func (m *PersonaDialogModel) SetSize(width, height int) {
	m.promptDialog.SetSize(width, height)
}

// SetDebugLogger sets the debug logger for the dialog
func (m *PersonaDialogModel) SetDebugLogger(logger *log.Logger) {
	m.debugLogger = logger
}

// Init initializes the dialog
func (m *PersonaDialogModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the dialog
func (m *PersonaDialogModel) Update(msg tea.Msg) (*PersonaDialogModel, tea.Cmd) {
	if !m.IsVisible() {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.debugLogger != nil {
			m.debugLogger.Printf("PERSONA_DIALOG: Key pressed: %q", msg.String())
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.availablePersonas) - 1
			}
			m.updateDialogContent()
		case "down", "j":
			if m.cursor < len(m.availablePersonas)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
			m.updateDialogContent()
		case " ":
			if m.cursor >= 0 && m.cursor < len(m.availablePersonas) {
				persona := m.availablePersonas[m.cursor]
				m.selectedPersonas[persona] = !m.selectedPersonas[persona]
				m.updateDialogContent()
			}
		case "enter":
			activePersonas := m.getActivePersonasList()
			m.Hide()
			return m, func() tea.Msg {
				return PersonaSelectionMsg{ActivePersonas: activePersonas}
			}
		case "esc":
			if m.debugLogger != nil {
				m.debugLogger.Printf("PERSONA_DIALOG: Escape key pressed - hiding dialog")
			}
			m.Hide()
			return m, nil
		default:
			// Let the underlying prompt dialog handle other keys (like scrolling)
			var cmd tea.Cmd
			m.promptDialog, cmd = m.promptDialog.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// getActivePersonasList returns the currently selected personas as a slice
func (m *PersonaDialogModel) getActivePersonasList() []string {
	var active []string
	for _, persona := range m.availablePersonas {
		if m.selectedPersonas[persona] {
			active = append(active, persona)
		}
	}
	if len(active) == 0 {
		active = []string{"default"}
	}
	return active
}

// View renders the dialog
func (m *PersonaDialogModel) View() string {
	return m.promptDialog.View()
}

// generateDialogContent creates the persona selection content
func (m *PersonaDialogModel) generateDialogContent() string {
	var content strings.Builder
	content.WriteString("Select Active Personas:\n\n")

	// Render persona list with checkboxes
	for i, persona := range m.availablePersonas {
		cursor := " "
		if i == m.cursor {
			cursor = "▶"
		}

		checkbox := "☐"
		if m.selectedPersonas[persona] {
			checkbox = "☑"
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, persona)

		// Highlight current selection
		if i == m.cursor {
			line = lipgloss.NewStyle().
				Background(lipgloss.Color("69")).
				Foreground(lipgloss.Color("0")).
				Render(" " + line + " ")
		} else {
			line = " " + line + " "
		}

		content.WriteString(line + "\n")
	}

	content.WriteString("\n")
	content.WriteString("Space: Toggle • Enter: Apply • Escape: Cancel")

	return content.String()
}

// updateDialogContent refreshes the dialog content after changes
func (m *PersonaDialogModel) updateDialogContent() {
	if m.IsVisible() {
		content := m.generateDialogContent()
		m.promptDialog.Show(content)
	}
}
