package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"coding-prompts-tui/internal/filesystem"
)

// FileTreeModel represents the file tree panel
type FileTreeModel struct {
	targetDir    string
	rootNode     *filesystem.FileNode
	items        []filesystem.FileTreeItem
	cursor       int
	title        string
	expanded     map[string]bool
	selected     map[string]bool
}

// NewFileTreeModel creates a new file tree model
func NewFileTreeModel(targetDir string) *FileTreeModel {
	return &FileTreeModel{
		targetDir: targetDir,
		title:     "📁 File Tree",
		items:     []filesystem.FileTreeItem{},
		cursor:    0,
		expanded:  make(map[string]bool),
		selected:  make(map[string]bool),
	}
}

// Init initializes the file tree model
func (m *FileTreeModel) Init() tea.Cmd {
	// Scan the target directory
	rootNode, err := filesystem.ScanDirectory(m.targetDir)
	if err != nil {
		// If we can't scan the directory, create a simple error item
		m.items = []filesystem.FileTreeItem{
			{Name: "Error: " + err.Error(), Path: "", IsDir: false, Level: 0},
		}
		return nil
	}
	
	m.rootNode = rootNode
	m.refreshItems()
	return nil
}

// refreshItems rebuilds the flattened item list based on current expanded state
func (m *FileTreeModel) refreshItems() {
	if m.rootNode == nil {
		m.items = []filesystem.FileTreeItem{}
		return
	}
	
	// Add the root directory items (not the root itself, but its children)
	m.items = []filesystem.FileTreeItem{}
	for _, child := range m.rootNode.Children {
		childItems := filesystem.FlattenTree(child, 0, m.expanded)
		// Update selected state from our local state
		for i := range childItems {
			childItems[i].Selected = m.selected[childItems[i].Path]
		}
		m.items = append(m.items, childItems...)
	}
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
				currentItem := m.items[m.cursor]
				m.expanded[currentItem.Path] = !m.expanded[currentItem.Path]
				m.refreshItems()
				// Return a file selection message to communicate with other panels
				return m, m.sendFileSelectionUpdate()
			}
		case " ":
			// Toggle file selection (only for files, not directories)
			if m.cursor < len(m.items) && !m.items[m.cursor].IsDir {
				currentItem := m.items[m.cursor]
				m.selected[currentItem.Path] = !m.selected[currentItem.Path]
				m.refreshItems()
				// Return a file selection message to communicate with other panels
				return m, m.sendFileSelectionUpdate()
			}
		}
	}
	return m, nil
}

// FileSelectionMsg represents a message about file selection changes
type FileSelectionMsg struct {
	SelectedFiles map[string]bool
}

// sendFileSelectionUpdate creates a file selection update message
func (m *FileTreeModel) sendFileSelectionUpdate() tea.Cmd {
	return func() tea.Msg {
		return FileSelectionMsg{SelectedFiles: m.selected}
	}
}

// GetSelectedFiles returns the currently selected files
func (m *FileTreeModel) GetSelectedFiles() map[string]bool {
	return m.selected
}

// GetItems returns the current items for testing
func (m *FileTreeModel) GetItems() []filesystem.FileTreeItem {
	return m.items
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
			if m.expanded[item.Path] {
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