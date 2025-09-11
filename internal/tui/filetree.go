package tui

import (
    "strings"

    "github.com/charmbracelet/bubbles/viewport"
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
    // viewport to enable scrolling when content exceeds available space
    viewport     viewport.Model
    width        int
    height       int
}

// NewFileTreeModel creates a new file tree model
func NewFileTreeModel(targetDir string, initialSelection []string) *FileTreeModel {
	selected := make(map[string]bool)
	for _, f := range initialSelection {
		selected[f] = true
	}

	return &FileTreeModel{
		targetDir: targetDir,
		title:     "📁 File Tree",
		items:     []filesystem.FileTreeItem{},
		cursor:    0,
		expanded:  make(map[string]bool),
		selected:  selected,
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
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
            m.ensureVisible()
        case "down", "j":
            if m.cursor < len(m.items)-1 {
                m.cursor++
            }
            m.ensureVisible()
        case "pgdown", "ctrl+f":
            if m.viewport.Height > 0 {
                m.cursor += m.viewport.Height
                if m.cursor >= len(m.items) {
                    m.cursor = len(m.items) - 1
                }
                m.ensureVisible()
            }
        case "pgup", "ctrl+b":
            if m.viewport.Height > 0 {
                m.cursor -= m.viewport.Height
                if m.cursor < 0 {
                    m.cursor = 0
                }
                m.ensureVisible()
            }
        case "home", "g":
            m.cursor = 0
            m.ensureVisible()
        case "end", "G":
            if len(m.items) > 0 {
                m.cursor = len(m.items) - 1
            }
            m.ensureVisible()
        case "enter":
            // Toggle directory expansion
            if m.cursor < len(m.items) && m.items[m.cursor].IsDir {
                currentItem := m.items[m.cursor]
                m.expanded[currentItem.Path] = !m.expanded[currentItem.Path]
                m.refreshItems()
                m.ensureVisible()
                // Return a file selection message to communicate with other panels
                return m, m.sendFileSelectionUpdate()
            }
        case " ":
            // Toggle file selection (only for files, not directories)
            if m.cursor < len(m.items) && !m.items[m.cursor].IsDir {
                currentItem := m.items[m.cursor]
                m.selected[currentItem.Path] = !m.selected[currentItem.Path]
                m.refreshItems()
                m.ensureVisible()
                // Return a file selection message to communicate with other panels
                return m, m.sendFileSelectionUpdate()
            }
        }
    case tea.MouseMsg:
        // Let viewport handle mouse wheel scrolling
        m.viewport, cmd = m.viewport.Update(msg)
        return m, cmd
    }
    return m, cmd
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

// calculateHeaderContent builds the header content and returns both the content and height
func (m *FileTreeModel) calculateHeaderContent() (string, int) {
    var header strings.Builder

    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("205"))
    header.WriteString(titleStyle.Render(m.title))
    header.WriteString("\n\n")

    helpStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("240")).
        Italic(true)
    header.WriteString(helpStyle.Render("↑/↓: navigate, PgUp/PgDn: page, Enter: expand/collapse, Space: select file, g/G: top/bottom"))
    header.WriteString("\n\n")

    // Compute rendered header height with wrapping against current width
    renderedHeader := lipgloss.NewStyle().Width(max(1, m.width)).Render(header.String())
    headerLineCount := 0
    if renderedHeader != "" {
        headerLineCount = 1 + strings.Count(renderedHeader, "\n")
    }
    
    return renderedHeader, headerLineCount
}

// View renders the file tree
func (m *FileTreeModel) View() string {
    // Get header content and height
    renderedHeader, headerLineCount := m.calculateHeaderContent()

    // Update viewport size first with correct header height
    m.ensureViewportSizedWithHeader(headerLineCount)

    // Build scrollable content
    var content strings.Builder
    for i, item := range m.items {
        var line strings.Builder

        indent := strings.Repeat("  ", item.Level)
        line.WriteString(indent)

        if i == m.cursor {
            line.WriteString("▶ ")
        } else {
            line.WriteString("  ")
        }

        if item.IsDir {
            if m.expanded[item.Path] {
                line.WriteString("📂 ")
            } else {
                line.WriteString("📁 ")
            }
        } else {
            if item.Selected {
                line.WriteString("☑️ ")
            } else {
                line.WriteString("📄 ")
            }
        }

        itemStyle := lipgloss.NewStyle()
        if i == m.cursor {
            itemStyle = itemStyle.Foreground(lipgloss.Color("69")).Bold(true)
        }
        if item.Selected {
            itemStyle = itemStyle.Foreground(lipgloss.Color("10"))
        }
        line.WriteString(itemStyle.Render(item.Name))

        content.WriteString(line.String())
        content.WriteString("\n")
    }

    // Set content and ensure proper scrolling
    m.viewport.SetContent(content.String())
    m.ensureVisible()

    return renderedHeader + m.viewport.View()
}

// SetSize sets the available width and height for the panel (including header).
func (m *FileTreeModel) SetSize(width, height int) {
    m.width = width
    m.height = height
    // Calculate proper header height and resize viewport accordingly
    _, headerHeight := m.calculateHeaderContent()
    m.ensureViewportSizedWithHeader(headerHeight)
    m.ensureVisible()
}


// ensureViewportSizedWithHeader sizes the viewport using the provided header height.
func (m *FileTreeModel) ensureViewportSizedWithHeader(headerHeight int) {
    if m.width <= 0 || m.height <= 0 {
        return
    }
    vpHeight := m.height - headerHeight
    if vpHeight < 1 {
        vpHeight = 1
    }
    if m.viewport.Width != m.width || m.viewport.Height != vpHeight {
        if m.viewport.Width == 0 && m.viewport.Height == 0 {
            m.viewport = viewport.New(m.width, vpHeight)
        }
        m.viewport.Width = m.width
        m.viewport.Height = vpHeight
    }
}

// max helper
func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

// ensureVisible scrolls the viewport so the cursor is within the visible window.
func (m *FileTreeModel) ensureVisible() {
    if m.viewport.Height <= 0 || len(m.items) == 0 {
        return
    }
    
    // Clamp cursor to valid bounds
    if m.cursor < 0 {
        m.cursor = 0
    }
    if m.cursor >= len(m.items) {
        m.cursor = len(m.items) - 1
    }

    top := m.viewport.YOffset
    bottom := m.viewport.YOffset + m.viewport.Height - 1

    // Scroll to show cursor if it's outside visible area
    if m.cursor < top {
        m.viewport.YOffset = m.cursor
    } else if m.cursor > bottom {
        m.viewport.YOffset = m.cursor - m.viewport.Height + 1
    }

    // Clamp YOffset to valid bounds with better bounds checking
    maxOffset := len(m.items) - m.viewport.Height
    if maxOffset < 0 {
        maxOffset = 0
    }
    
    if m.viewport.YOffset < 0 {
        m.viewport.YOffset = 0
    }
    if m.viewport.YOffset > maxOffset {
        m.viewport.YOffset = maxOffset
    }
    
    // Double-check that YOffset doesn't cause content to be cut off
    if m.viewport.YOffset > 0 && m.viewport.YOffset + m.viewport.Height > len(m.items) {
        m.viewport.YOffset = max(0, len(m.items) - m.viewport.Height)
    }
}
