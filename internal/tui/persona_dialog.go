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
	selectedPersonas  map[string]bool
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
		case " ":
			if m.cursor >= 0 && m.cursor < len(m.availablePersonas) {
				persona := m.availablePersonas[m.cursor]
				m.selectedPersonas[persona] = !m.selectedPersonas[persona]
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

// View renders the dialog (for backwards compatibility)
func (m *PersonaDialogModel) View() string {
	if !m.visible {
		return ""
	}
	return m.renderPopover()
}

// ViewAsOverlay renders the dialog as a full-screen overlay with dimmed backdrop
func (m *PersonaDialogModel) ViewAsOverlay(backgroundContent string) string {
	if !m.visible {
		return backgroundContent
	}

	// Create dimmed backdrop
	backdrop := m.createDimmedBackdrop(backgroundContent)
	
	// Get the popover dialog
	popover := m.renderPopover()
	
	// Center the popover over the backdrop
	return m.centerPopover(backdrop, popover)
}

// renderPopover creates the dialog content with enhanced styling
func (m *PersonaDialogModel) renderPopover() string {
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

	// Create dialog box style with shadow effect
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Background(lipgloss.Color("0")).     // Solid dark background
		Foreground(lipgloss.Color("15")).    // Bright text
		Padding(1, 2).
		Width(40).                           // Fixed width for consistency
		Align(lipgloss.Center)

	// Add a subtle shadow effect
	// shadowStyle := lipgloss.NewStyle().
	// 	Background(lipgloss.Color("8")).
	// 	MarginLeft(1).
	// 	MarginTop(1)

	dialog := dialogStyle.Render(content.String())
	
	// Create shadow by rendering a slightly offset version
	// shadow := shadowStyle.Render(strings.Repeat(" ", lipgloss.Width(dialog)))
	
	// Combine dialog with shadow (simplified version)
	return dialog
}

// createDimmedBackdrop applies a dimming effect to background content
func (m *PersonaDialogModel) createDimmedBackdrop(content string) string {
	lines := strings.Split(content, "\n")
	dimmedLines := make([]string, len(lines))
	
	// Style to dim the background content
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))  // Dark gray - makes content visible but dimmed
	
	for i, line := range lines {
		dimmedLines[i] = dimStyle.Render(line)
	}
	
	return strings.Join(dimmedLines, "\n")
}

// centerPopover positions the popover in the center of the backdrop
func (m *PersonaDialogModel) centerPopover(backdrop, popover string) string {
	backdropLines := strings.Split(backdrop, "\n")
	popoverLines := strings.Split(popover, "\n")
	
	// Calculate dimensions
	backdropHeight := len(backdropLines)
	popoverHeight := len(popoverLines)
	
	// Find the widest line in the popover for centering
	popoverWidth := 0
	for _, line := range popoverLines {
		// Use visual width (accounting for ANSI codes)
		width := lipgloss.Width(line)
		if width > popoverWidth {
			popoverWidth = width
		}
	}
	
	// Calculate center position
	startRow := (backdropHeight - popoverHeight) / 2
	if startRow < 0 {
		startRow = 0
	}
	
	// Create result by overlaying popover onto backdrop
	result := make([]string, len(backdropLines))
	copy(result, backdropLines)
	
	// Place popover lines
	for i, popoverLine := range popoverLines {
		targetRow := startRow + i
		if targetRow >= len(result) {
			break
		}
		
		if m.width > 0 {
			// Center horizontally
			startCol := (m.width - popoverWidth) / 2
			if startCol < 0 {
				startCol = 0
			}
			
			// For simplicity, replace the entire line with centered popover content
			padding := strings.Repeat(" ", startCol)
			result[targetRow] = padding + popoverLine
		} else {
			// No width info, just place at start
			result[targetRow] = popoverLine
		}
	}
	
	return strings.Join(result, "\n")
}

// Alternative simpler overlay approach for easier integration
func (m *PersonaDialogModel) ViewAsSimpleOverlay(backgroundContent string) string {
	if !m.visible {
		return backgroundContent
	}
	
	// Split content into lines
	lines := strings.Split(backgroundContent, "\n")
	
	// Dim all background lines
	for i, line := range lines {
		lines[i] = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(line)
	}
	
	// Get popover content
	popover := m.renderPopover()
	popoverLines := strings.Split(popover, "\n")
	
	// Calculate where to place the popover (center)
	startRow := (len(lines) - len(popoverLines)) / 2
	if startRow < 0 {
		startRow = 0
	}
	
	// Overlay popover lines
	for i, popoverLine := range popoverLines {
		targetRow := startRow + i
		if targetRow < len(lines) {
			// Center the popover line
			if m.width > 0 {
				popoverWidth := lipgloss.Width(popoverLine)
				padding := (m.width - popoverWidth) / 2
				if padding > 0 {
					lines[targetRow] = strings.Repeat(" ", padding) + popoverLine
				} else {
					lines[targetRow] = popoverLine
				}
			} else {
				lines[targetRow] = popoverLine
			}
		}
	}
	
	return strings.Join(lines, "\n")
}
