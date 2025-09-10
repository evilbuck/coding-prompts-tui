package tui

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FocusedPanel represents which panel currently has focus
type FocusedPanel int

const (
	FileTreePanel FocusedPanel = iota
	SelectedFilesPanel
	ChatPanel
)

// App represents the main application model
type App struct {
	targetDir    string
	width        int
	height       int
	focused      FocusedPanel
	fileTree     *FileTreeModel
	selectedFiles *SelectedFilesModel
	chat         *ChatModel
}

// NewApp creates a new application instance
func NewApp(targetDir string) *App {
	return &App{
		targetDir:     targetDir,
		focused:       FileTreePanel,
		fileTree:      NewFileTreeModel(targetDir),
		selectedFiles: NewSelectedFilesModel(),
		chat:          NewChatModel(),
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.fileTree.Init(),
		a.selectedFiles.Init(),
		a.chat.Init(),
	)
}

// Update handles messages and updates the application state
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case FileSelectionMsg:
		// Update selected files panel when file selection changes
		a.updateSelectedFilesFromSelection(msg.SelectedFiles)
		return a, nil

	case FileDeselectionMsg:
		// Update file tree selection state when file is removed from selected files
		a.fileTree.selected[msg.FilePath] = false
		a.fileTree.refreshItems()
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			a.nextPanel()
			return a, nil
		case "shift+tab":
			a.prevPanel()
			return a, nil
		}
	}

	// Update the focused panel
	switch a.focused {
	case FileTreePanel:
		model, cmd := a.fileTree.Update(msg)
		a.fileTree = model.(*FileTreeModel)
		cmds = append(cmds, cmd)
	case SelectedFilesPanel:
		model, cmd := a.selectedFiles.Update(msg)
		a.selectedFiles = model.(*SelectedFilesModel)
		cmds = append(cmds, cmd)
	case ChatPanel:
		model, cmd := a.chat.Update(msg)
		a.chat = model.(*ChatModel)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View renders the application
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	// Calculate panel dimensions
	topHeight := int(float64(a.height) * 0.66)
	bottomHeight := a.height - topHeight
	leftWidth := a.width / 2
	rightWidth := a.width - leftWidth

	// Create styles for panels
	focusedBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69"))

	normalBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	// File tree panel (top-left)
	fileTreeStyle := normalBorder
	if a.focused == FileTreePanel {
		fileTreeStyle = focusedBorder
	}
	fileTreePanel := fileTreeStyle.
		Width(leftWidth - 2).
		Height(topHeight - 2).
		Render(a.fileTree.View())

	// Selected files panel (top-right)
	selectedStyle := normalBorder
	if a.focused == SelectedFilesPanel {
		selectedStyle = focusedBorder
	}
	selectedPanel := selectedStyle.
		Width(rightWidth - 2).
		Height(topHeight - 2).
		Render(a.selectedFiles.View())

	// Chat panel (bottom)
	chatStyle := normalBorder
	if a.focused == ChatPanel {
		chatStyle = focusedBorder
	}
	chatPanel := chatStyle.
		Width(a.width - 2).
		Height(bottomHeight - 2).
		Render(a.chat.View())

	// Layout the panels
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, fileTreePanel, selectedPanel)
	
	return lipgloss.JoinVertical(lipgloss.Left, topRow, chatPanel)
}

// nextPanel moves focus to the next panel
func (a *App) nextPanel() {
	switch a.focused {
	case FileTreePanel:
		a.focused = SelectedFilesPanel
	case SelectedFilesPanel:
		a.focused = ChatPanel
	case ChatPanel:
		a.focused = FileTreePanel
	}
}

// prevPanel moves focus to the previous panel
func (a *App) prevPanel() {
	switch a.focused {
	case FileTreePanel:
		a.focused = ChatPanel
	case SelectedFilesPanel:
		a.focused = FileTreePanel
	case ChatPanel:
		a.focused = SelectedFilesPanel
	}
}

// updateSelectedFilesFromSelection synchronizes the selected files panel with file tree selection
func (a *App) updateSelectedFilesFromSelection(selectedFiles map[string]bool) {
	// Clear current selection
	a.selectedFiles.files = []SelectedFile{}
	
	// Add all currently selected files
	for path, selected := range selectedFiles {
		if selected {
			a.selectedFiles.AddFile(filepath.Base(path), path)
		}
	}
	
	// Reset cursor if needed
	if len(a.selectedFiles.files) == 0 {
		a.selectedFiles.cursor = 0
	} else if a.selectedFiles.cursor >= len(a.selectedFiles.files) {
		a.selectedFiles.cursor = len(a.selectedFiles.files) - 1
	}
}