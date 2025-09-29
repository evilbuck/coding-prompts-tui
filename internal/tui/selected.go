package tui

import (
	"fmt"
	"strings"

	"coding-prompts-tui/internal/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SelectedFile represents a file that has been selected for inclusion
type SelectedFile struct {
	Name string
	Path string
}

// SelectedFilesModel represents the selected files panel
type SelectedFilesModel struct {
	files         []SelectedFile
	cursor        int
	title         string
	configManager *config.ConfigManager
}

// NewSelectedFilesModel creates a new selected files model
func NewSelectedFilesModel(configManager *config.ConfigManager) *SelectedFilesModel {
	return &SelectedFilesModel{
		title:         "✅ Selected Files",
		files:         []SelectedFile{},
		cursor:        0,
		configManager: configManager,
	}
}

// Init initializes the selected files model
func (m *SelectedFilesModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the selected files panel
func (m *SelectedFilesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		default:
			// Check if this key is configured for file removal
			settings := m.configManager.GetSelectedFilesPanelSettings()
			for _, removalKey := range settings.RemovalKeys {
				if msg.String() == removalKey {
					// Remove selected file
					if len(m.files) > 0 && m.cursor < len(m.files) {
						removedFile := m.files[m.cursor]
						m.removeFile(m.cursor)
						// Send a message to update the file tree selection state
						return m, m.sendFileDeselectionUpdate(removedFile.Path)
					}
					break
				}
			}
		}
	}
	return m, nil
}

// View renders the selected files panel
func (m *SelectedFilesModel) View() string {
	var b strings.Builder

	// Title row
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10"))

	titleText := titleStyle.Render(m.title)

	b.WriteString(titleText)
	b.WriteString("\n\n")

	// Help text - contextual based on whether files exist and are selected
	settings := m.configManager.GetSelectedFilesPanelSettings()
	if settings.ShowHelpText {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)

		if len(m.files) == 0 {
			b.WriteString(helpStyle.Render("No files selected"))
		} else {
			// Format the removal keys for display
			keyNames := m.formatKeysForDisplay(settings.RemovalKeys)
			helpText := fmt.Sprintf(settings.HelpText, keyNames)
			b.WriteString(helpStyle.Render(helpText))
		}
		b.WriteString("\n\n")
	}

	// Selected files list
	if len(m.files) == 0 {
		// Empty state is already shown in help text
		// No additional content needed here
	} else {
		for i, file := range m.files {
			var line strings.Builder

			// Cursor indicator
			if i == m.cursor {
				line.WriteString("▶ ")
			} else {
				line.WriteString("  ")
			}

			// File icon
			line.WriteString("📄 ")

			// File name
			fileStyle := lipgloss.NewStyle()
			if i == m.cursor {
				fileStyle = fileStyle.Foreground(lipgloss.Color("69")).Bold(true)
			}

			line.WriteString(fileStyle.Render(file.Name))

			b.WriteString(line.String())
			b.WriteString("\n")
		}
	}

	// Count
	b.WriteString("\n")
	countStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	b.WriteString(countStyle.Render(fmt.Sprintf("Total: %d files", len(m.files))))

	return b.String()
}

// AddFile adds a file to the selected files list
func (m *SelectedFilesModel) AddFile(name, path string) {
	// Check if file is already selected
	for _, file := range m.files {
		if file.Path == path {
			return // Already selected
		}
	}

	m.files = append(m.files, SelectedFile{
		Name: name,
		Path: path,
	})
}

// RemoveFile removes a file from the selected files list by path
func (m *SelectedFilesModel) RemoveFile(path string) {
	for i, file := range m.files {
		if file.Path == path {
			m.removeFile(i)
			return
		}
	}
}

// removeFile removes a file at the given index
func (m *SelectedFilesModel) removeFile(index int) {
	if index < 0 || index >= len(m.files) {
		return
	}

	m.files = append(m.files[:index], m.files[index+1:]...)

	// Adjust cursor if necessary
	if m.cursor >= len(m.files) && len(m.files) > 0 {
		m.cursor = len(m.files) - 1
	}
	if len(m.files) == 0 {
		m.cursor = 0
	}
}

// GetSelectedFiles returns the list of selected files
func (m *SelectedFilesModel) GetSelectedFiles() []SelectedFile {
	return m.files
}

// ClearAllFiles removes all files from the selected files list
func (m *SelectedFilesModel) ClearAllFiles() tea.Cmd {
	// Clear all files
	m.files = []SelectedFile{}
	m.cursor = 0

	// Create a command that will notify the app about the clear action
	return func() tea.Msg {
		return ClearAllFilesMsg{}
	}
}

// FileDeselectionMsg represents a message about file deselection
type FileDeselectionMsg struct {
	FilePath string
}

// ClearAllFilesMsg represents a message about clearing all selected files
type ClearAllFilesMsg struct{}

// sendFileDeselectionUpdate creates a file deselection update message
func (m *SelectedFilesModel) sendFileDeselectionUpdate(filePath string) tea.Cmd {
	return func() tea.Msg {
		return FileDeselectionMsg{FilePath: filePath}
	}
}

// formatKeysForDisplay formats the removal keys for display in help text
func (m *SelectedFilesModel) formatKeysForDisplay(keys []string) string {
	if len(keys) == 0 {
		return ""
	}

	// Convert key names to display names
	displayKeys := make([]string, len(keys))
	for i, key := range keys {
		switch key {
		case " ":
			displayKeys[i] = "space"
		case "delete":
			displayKeys[i] = "del"
		default:
			displayKeys[i] = key
		}
	}

	// Join with slashes
	return strings.Join(displayKeys, "/")
}
