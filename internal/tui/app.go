package tui

import (
	"path/filepath"
	"time"

	"coding-prompts-tui/internal/prompt"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/timer"
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

// Notification message types
type ShowNotificationMsg struct {
	text string
}

type HideNotificationMsg struct{}

// App represents the main application model
type App struct {
	targetDir           string
	width               int
	height              int
	focused             FocusedPanel
	fileTree            *FileTreeModel
	selectedFiles       *SelectedFilesModel
	chat                *ChatModel
	promptDialog        *PromptDialogModel
	notificationText    string
	notificationVisible bool
	notificationTimer   timer.Model
}

// NewApp creates a new application instance
func NewApp(targetDir string) *App {
	return &App{
		targetDir:     targetDir,
		focused:       FileTreePanel,
		fileTree:      NewFileTreeModel(targetDir),
		selectedFiles: NewSelectedFilesModel(),
		chat:          NewChatModel(),
		promptDialog:  NewPromptDialogModel(),
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
		a.promptDialog.SetSize(msg.Width, msg.Height)
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

	case ShowNotificationMsg:
		a.notificationText = msg.text
		a.notificationVisible = true
		a.notificationTimer = timer.New(750 * time.Millisecond)
		return a, a.notificationTimer.Init()

	case timer.TimeoutMsg:
		if a.notificationTimer.ID() == msg.ID {
			a.notificationVisible = false
		}
		return a, nil

	case HideNotificationMsg:
		a.notificationVisible = false
		return a, nil

	case tea.KeyMsg:
		// Handle global clipboard copy first
		if msg.String() == "ctrl+shift+c" {
			var promptToCopy string
			if a.promptDialog.IsVisible() && a.promptDialog.GetContent() != "" {
				promptToCopy = a.promptDialog.GetContent()
			} else {
				generatedPrompt, err := prompt.Build(a.targetDir, a.fileTree.selected, a.chat.textarea.Value())
				if err != nil {
					// Show error notification
					return a, tea.Cmd(func() tea.Msg {
						return ShowNotificationMsg{text: "error building prompt"}
					})
				}
				promptToCopy = generatedPrompt
			}
			
			err := clipboard.WriteAll(promptToCopy)
			if err != nil {
				// Show error notification
				return a, tea.Cmd(func() tea.Msg {
					return ShowNotificationMsg{text: "clipboard error"}
				})
			}
			
			// Show success notification
			return a, tea.Cmd(func() tea.Msg {
				return ShowNotificationMsg{text: "prompt copied"}
			})
		}

		// Handle prompt dialog input if visible
		if a.promptDialog.IsVisible() {
			model, cmd := a.promptDialog.Update(msg)
			a.promptDialog = model
			return a, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			a.nextPanel()
			return a, nil
		case "shift+tab":
			a.prevPanel()
			return a, nil
		case "ctrl+s":
			generatedPrompt, err := prompt.Build(a.targetDir, a.fileTree.selected, a.chat.textarea.Value())
			if err != nil {
				// Handle error, maybe show an error message
				// For now, we'll just log it
				// log.Printf("Error building prompt: %v", err)
			} else {
				a.promptDialog.Show(generatedPrompt)
			}
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

	// Main layout
	mainLayout := a.mainLayout()

	// Show prompt dialog if visible
	if a.promptDialog.IsVisible() {
		dialogView := a.promptDialog.View()
		// Overlay notification on dialog if visible
		if a.notificationVisible {
			return a.overlayNotification(dialogView)
		}
		return dialogView
	}

	// Overlay notification on main layout if visible
	if a.notificationVisible {
		return a.overlayNotification(mainLayout)
	}

	return mainLayout
}

func (a *App) mainLayout() string {
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

// overlayNotification renders the notification overlay on top of the given content
func (a *App) overlayNotification(content string) string {
	notification := lipgloss.NewStyle().
		Background(lipgloss.Color("2")).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1).
		Bold(true).
		Render(a.notificationText)
	
	// Position notification at the top center
	return lipgloss.Place(a.width, a.height,
		lipgloss.Center, lipgloss.Top,
		notification,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.AdaptiveColor{Light: "0", Dark: "0"}),
	) + "\n" + content
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