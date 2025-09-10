package tui

import (
	"fmt"
	"strings"

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
	files  []SelectedFile
	cursor int
	title  string
}

// NewSelectedFilesModel creates a new selected files model
func NewSelectedFilesModel() *SelectedFilesModel {
	return &SelectedFilesModel{
		title:  "✅ Selected Files",
		files:  []SelectedFile{},
		cursor: 0,
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
		case "delete", "backspace", "x":
			// Remove selected file
			if len(m.files) > 0 && m.cursor < len(m.files) {
				removedFile := m.files[m.cursor]
				m.removeFile(m.cursor)
				// Send a message to update the file tree selection state
				return m, m.sendFileDeselectionUpdate(removedFile.Path)
			}
		}
	}
	return m, nil
}

// View renders the selected files panel
func (m *SelectedFilesModel) View() string {
	var b strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("10"))
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	b.WriteString(helpStyle.Render("↑/↓: navigate, x/del: remove file"))
	b.WriteString("\n\n")

	// Selected files list
	if len(m.files) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		b.WriteString(emptyStyle.Render("No files selected"))
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

// FileDeselectionMsg represents a message about file deselection
type FileDeselectionMsg struct {
	FilePath string
}

// sendFileDeselectionUpdate creates a file deselection update message
func (m *SelectedFilesModel) sendFileDeselectionUpdate(filePath string) tea.Cmd {
	return func() tea.Msg {
		return FileDeselectionMsg{FilePath: filePath}
	}
}