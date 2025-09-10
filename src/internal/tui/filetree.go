package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FileItem represents a file or directory in the tree
type FileItem struct {
	Name     string
	Path     string
	IsDir    bool
	Selected bool
	Expanded bool
	Level    int
	Children []*FileItem
}

// FileTreeModel represents the file tree panel
type FileTreeModel struct {
	targetDir string
	items     []*FileItem
	cursor    int
	title     string
}

// NewFileTreeModel creates a new file tree model
func NewFileTreeModel(targetDir string) *FileTreeModel {
	return &FileTreeModel{
		targetDir: targetDir,
		title:     "📁 File Tree",
		items:     []*FileItem{},
		cursor:    0,
	}
}

// Init initializes the file tree model
func (m *FileTreeModel) Init() tea.Cmd {
	// TODO: Load directory structure
	// For now, add some placeholder items
	m.items = []*FileItem{
		{Name: "src", Path: "src", IsDir: true, Level: 0, Expanded: false},
		{Name: "main.go", Path: "main.go", IsDir: false, Level: 0},
		{Name: "README.md", Path: "README.md", IsDir: false, Level: 0},
		{Name: "go.mod", Path: "go.mod", IsDir: false, Level: 0},
	}
	return nil
}

// Update handles messages for the file tree
func (m *FileTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			// Toggle directory expansion
			if m.cursor < len(m.items) && m.items[m.cursor].IsDir {
				m.items[m.cursor].Expanded = !m.items[m.cursor].Expanded
				// TODO: Load/hide children
			}
		case " ":
			// Toggle file selection (only for files, not directories)
			if m.cursor < len(m.items) && !m.items[m.cursor].IsDir {
				m.items[m.cursor].Selected = !m.items[m.cursor].Selected
			}
		}
	}
	return m, nil
}

// View renders the file tree
func (m *FileTreeModel) View() string {
	var b strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205"))
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	b.WriteString(helpStyle.Render("↑/↓: navigate, Enter: expand/collapse, Space: select file"))
	b.WriteString("\n\n")

	// File list
	for i, item := range m.items {
		var line strings.Builder

		// Indentation for tree structure
		indent := strings.Repeat("  ", item.Level)
		line.WriteString(indent)

		// Cursor indicator
		if i == m.cursor {
			line.WriteString("▶ ")
		} else {
			line.WriteString("  ")
		}

		// Icon and expansion indicator
		if item.IsDir {
			if item.Expanded {
				line.WriteString("📂 ")
			} else {
				line.WriteString("📁 ")
			}
		} else {
			// File selection indicator
			if item.Selected {
				line.WriteString("☑️ ")
			} else {
				line.WriteString("📄 ")
			}
		}

		// Item name
		itemStyle := lipgloss.NewStyle()
		if i == m.cursor {
			itemStyle = itemStyle.Foreground(lipgloss.Color("69")).Bold(true)
		}
		if item.Selected {
			itemStyle = itemStyle.Foreground(lipgloss.Color("10"))
		}
		
		line.WriteString(itemStyle.Render(item.Name))

		b.WriteString(line.String())
		b.WriteString("\n")
	}

	return b.String()
}