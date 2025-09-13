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
	visible           bool
	availablePersonas []string
	selectedPersonas  map[string]bool // map of persona name to selected state
	cursor            int
	width             int
	height            int
	debugLogger       *log.Logger
}

// PersonaSelectionMsg is sent when personas are selected/deselected
type PersonaSelectionMsg struct {
	ActivePersonas []string
}

// NewPersonaDialogModel creates a new persona dialog model
func NewPersonaDialogModel() *PersonaDialogModel {
	return &PersonaDialogModel{
		visible:          false,
		selectedPersonas: make(map[string]bool),
		cursor:           0,
		debugLogger:      nil,
	}
}

// SetAvailablePersonas sets the list of available personas
func (m *PersonaDialogModel) SetAvailablePersonas(personas []string) {
	m.availablePersonas = personas
	// Ensure cursor is within bounds
	if m.cursor >= len(m.availablePersonas) {
		m.cursor = 0
	}
}

// SetActivePersonas sets the currently active personas
func (m *PersonaDialogModel) SetActivePersonas(personas []string) {
	// Clear current selection
	m.selectedPersonas = make(map[string]bool)
	// Set active personas
	for _, persona := range personas {
		m.selectedPersonas[persona] = true
	}
}

// Show displays the dialog
func (m *PersonaDialogModel) Show() {
	m.visible = true
}

// Hide closes the dialog
func (m *PersonaDialogModel) Hide() {
	m.visible = false
}

// IsVisible returns whether the dialog is visible
func (m *PersonaDialogModel) IsVisible() bool {
	return m.visible
}

// SetSize sets the dialog size for centering
func (m *PersonaDialogModel) SetSize(width, height int) {
	m.width = width
	m.height = height
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
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Debug logging for all key presses in persona dialog
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
		case "down", "j":
			if m.cursor < len(m.availablePersonas)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case " ": // Space bar to toggle selection
			if m.cursor >= 0 && m.cursor < len(m.availablePersonas) {
				persona := m.availablePersonas[m.cursor]
				m.selectedPersonas[persona] = !m.selectedPersonas[persona]
			}
		case "enter":
			// Apply selection and close dialog
			activePersonas := m.getActivePersonasList()
			m.Hide()
			return m, func() tea.Msg {
				return PersonaSelectionMsg{ActivePersonas: activePersonas}
			}
		case "escape":
			// Debug logging for escape key
			if m.debugLogger != nil {
				m.debugLogger.Printf("PERSONA_DIALOG: Escape key pressed - hiding dialog")
			}
			// Cancel and close dialog without applying changes
			m.Hide()
			return m, nil
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
	// If no personas selected, default to "default"
	if len(active) == 0 {
		active = []string{"default"}
	}
	return active
}

// View renders the dialog
func (m *PersonaDialogModel) View() string {
	if !m.visible {
		return ""
	}

	// Create dialog content
	var content strings.Builder
	content.WriteString("Select Active Personas:\n\n")

	// Render persona list with checkboxes
	for i, persona := range m.availablePersonas {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		checkbox := "☐"
		if m.selectedPersonas[persona] {
			checkbox = "☑"
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, persona)
		if i == m.cursor {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("69")).
				Render(line)
		}
		content.WriteString(line + "\n")
	}

	content.WriteString("\n")
	content.WriteString("Space: Toggle • Enter: Apply • Escape: Cancel")

	// Calculate dialog dimensions with safety checks
	lines := strings.Split(content.String(), "\n")
	dialogWidth := 30 // Minimum width
	for _, line := range lines {
		if len(line) > dialogWidth {
			dialogWidth = len(line)
		}
	}
	dialogWidth += 4 // Add padding
	dialogHeight := len(lines) + 2 // Add padding
	
	// Ensure minimum dimensions
	if dialogWidth < 30 {
		dialogWidth = 30
	}
	if dialogHeight < 5 {
		dialogHeight = 5
	}

	// Create dialog box style
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Padding(1, 2).
		Width(dialogWidth).
		Height(dialogHeight)

	dialog := dialogStyle.Render(content.String())

	// Center the dialog (only if we have valid dimensions)
	if m.width > 0 && m.height > 0 {
		dialogLines := strings.Split(dialog, "\n")
		actualDialogHeight := len(dialogLines)
		actualDialogWidth := 0
		for _, line := range dialogLines {
			// Calculate actual width without ANSI codes
			cleanLine := lipgloss.NewStyle().Render(line)
			if len(cleanLine) > actualDialogWidth {
				actualDialogWidth = len(cleanLine)
			}
		}

		// Calculate centering position with bounds checking
		topPadding := (m.height - actualDialogHeight) / 2
		if topPadding < 0 {
			topPadding = 0
		}
		leftPadding := (m.width - actualDialogWidth) / 2
		if leftPadding < 0 {
			leftPadding = 0
		}

		// Add top padding
		var centeredDialog strings.Builder
		for i := 0; i < topPadding; i++ {
			centeredDialog.WriteString("\n")
		}

		// Add left padding to each line
		for _, line := range dialogLines {
			centeredDialog.WriteString(strings.Repeat(" ", leftPadding) + line + "\n")
		}

		return centeredDialog.String()
	}

	return dialog
}