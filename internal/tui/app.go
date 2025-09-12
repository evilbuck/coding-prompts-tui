package tui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"coding-prompts-tui/internal/config"
	"coding-prompts-tui/internal/persona"
	"coding-prompts-tui/internal/prompt"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.dalton.dog/bubbleup"
)

// FocusedPanel represents which panel currently has focus
type FocusedPanel int

const (
	FileTreePanel FocusedPanel = iota
	SelectedFilesPanel
	ChatPanel
	FooterMenuPanel
)

// App represents the main application model
type App struct {
	targetDir       string
	width           int
	height          int
	focused         FocusedPanel
	menuBindingMode bool
	fileTree        *FileTreeModel
	selectedFiles   *SelectedFilesModel
	chat            *ChatModel
	promptDialog    *PromptDialogModel
	personaDialog   *PersonaDialogModel
	alertModel      bubbleup.AlertModel
	configManager   *config.ConfigManager
	settingsManager *config.SettingsManager
	personaManager  *persona.Manager
	workspace       *config.WorkspaceState
	debugMode       bool
	lastDebugInfo   string
	debugLogger     *log.Logger
}

// NewApp creates a new application instance
func NewApp(targetDir string, cfgManager *config.ConfigManager, settingsManager *config.SettingsManager, workspace *config.WorkspaceState) *App {
	fileTree := NewFileTreeModel(targetDir, workspace.SelectedFiles)
	selectedFiles := NewSelectedFilesModel(cfgManager)
	chat := NewChatModel(workspace.ChatInput)
	
	// Initialize persona manager and discover personas
	personaManager := persona.NewManager(targetDir)
	personaManager.DiscoverPersonas()
	
	// Initialize persona dialog
	personaDialog := NewPersonaDialogModel()
	personaDialog.SetAvailablePersonas(personaManager.GetAvailablePersonas())
	personaDialog.SetActivePersonas(workspace.ActivePersonas)

	// Initialize debug logger
	debugLogger := initializeDebugLogger(targetDir)

	app := &App{
		targetDir:       targetDir,
		focused:         FileTreePanel,
		fileTree:        fileTree,
		selectedFiles:   selectedFiles,
		chat:            chat,
		promptDialog:    NewPromptDialogModel(),
		personaDialog:   personaDialog,
		alertModel:      *bubbleup.NewAlertModel(40, true), // Will be updated dynamically on window resize
		configManager:   cfgManager,
		settingsManager: settingsManager,
		personaManager:  personaManager,
		workspace:       workspace,
		debugLogger:     debugLogger,
	}
	app.updateSelectedFilesFromSelection(fileTree.selected)
	return app
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.fileTree.Init(),
		a.selectedFiles.Init(),
		a.chat.Init(),
		a.personaDialog.Init(),
		a.alertModel.Init(),
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
		a.personaDialog.SetSize(msg.Width, msg.Height)

		// Update notification width to 30% of interface width, with reasonable bounds
		notificationWidth := int(float64(msg.Width) * 0.3)
		if notificationWidth < 20 {
			notificationWidth = 20 // Minimum width for readability
		} else if notificationWidth > 80 {
			notificationWidth = 80 // Maximum width to prevent overly wide notifications
		}

		// Create new AlertModel with updated width
		a.alertModel = *bubbleup.NewAlertModel(notificationWidth, true)

		// Propagate calculated panel sizes to sub-models that need them
		// These calculations must match exactly what mainLayout() gives to the border
		headerHeight := 3 // Single line header with padding
		footerHeight := 3 // Single line footer with padding
		availableHeight := a.height - headerHeight - footerHeight
		topHeight := int(float64(availableHeight) * 0.66)
		leftWidth := a.width / 2
		// The border style sets Width(leftWidth-2) and Height(topHeight-2)
		// So the content area inside the border is even smaller
		// We need to account for the border padding (typically 1 char on each side)
		contentWidth := leftWidth - 2 - 2  // border width minus border padding
		contentHeight := topHeight - 2 - 2 // border height minus border padding
		a.fileTree.SetSize(contentWidth, contentHeight)
		return a, nil

	case tea.MouseMsg:
		// Handle mouse clicks for panel focus
		if msg.Type == tea.MouseLeft {
			a.handleMouseClick(msg.X, msg.Y)
		}
		// Let the currently focused panel handle the mouse event
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
			chatModel, chatCmd := a.chat.Update(msg)
			a.chat = chatModel.(*ChatModel)
			cmds = append(cmds, chatCmd)
		}
		return a, tea.Batch(cmds...)

	case FileSelectionMsg:
		// Update selected files panel when file selection changes
		a.updateSelectedFilesFromSelection(msg.SelectedFiles)
		a.workspace.SelectedFiles = []string{}
		for path, selected := range msg.SelectedFiles {
			if selected {
				a.workspace.SelectedFiles = append(a.workspace.SelectedFiles, path)
			}
		}
		a.configManager.Save()
		return a, nil

	case ChatInputMsg:
		a.workspace.ChatInput = msg.Content
		a.configManager.Save()
		return a, nil

	case FileDeselectionMsg:
		// Update file tree selection state when file is removed from selected files
		a.fileTree.selected[msg.FilePath] = false
		a.fileTree.refreshItems()
		// Also update workspace state
		var newSelected []string
		for _, f := range a.workspace.SelectedFiles {
			if f != msg.FilePath {
				newSelected = append(newSelected, f)
			}
		}
		a.workspace.SelectedFiles = newSelected
		a.configManager.Save()
		return a, nil

	case PersonaSelectionMsg:
		// Update workspace state with new active personas
		a.workspace.ActivePersonas = msg.ActivePersonas
		a.configManager.Save()
		return a, nil

	case tea.KeyMsg:
		// Handle global clipboard copy first
		if msg.String() == "ctrl+y" {
			var promptToCopy string
			if a.promptDialog.IsVisible() && a.promptDialog.GetContent() != "" {
				promptToCopy = a.promptDialog.GetContent()
			} else {
				generatedPrompt, err := prompt.Build(a.targetDir, a.fileTree.selected, a.chat.textarea.Value(), a.workspace.ActivePersonas)
				if err != nil {
					// Show error notification
					alertCmd := a.createAlert(bubbleup.ErrorKey, "error building prompt")
					return a, alertCmd
				}
				promptToCopy = generatedPrompt
			}

			err := clipboard.WriteAll(promptToCopy)
			if err != nil {
				// Show error notification
				alertCmd := a.createAlert(bubbleup.ErrorKey, "clipboard error")
				return a, alertCmd
			}

			// Show success notification
			alertCmd := a.createAlert(bubbleup.InfoKey, "prompt copied")
			return a, alertCmd
		}

		// Handle persona dialog input if visible
		if a.personaDialog.IsVisible() {
			model, cmd := a.personaDialog.Update(msg)
			a.personaDialog = model
			return a, cmd
		}

		// Handle prompt dialog input if visible
		if a.promptDialog.IsVisible() {
			model, cmd := a.promptDialog.Update(msg)
			a.promptDialog = model
			return a, cmd
		}

		// Handle menu activation first (supports both legacy and new modes)
		if a.handleMenuActivation(msg) {
			alertCmd := a.createAlert(bubbleup.InfoKey, "menu mode activated")
			return a, alertCmd
		}

		// Debug mode: show key information (after menu handling so we can see if activation works)
		if a.debugMode {
			debugInfo := fmt.Sprintf("Key: %q, Type: %v, Alt: %v, Runes: %v", msg.String(), msg.Type, msg.Alt, msg.Runes)
			if a.lastDebugInfo != "" {
				debugInfo = a.lastDebugInfo + " | " + debugInfo
				a.lastDebugInfo = ""
			}
			
			// Log to file
			if a.debugLogger != nil {
				a.debugLogger.Printf("DEBUG: %s", debugInfo)
			}
			
			// Also show as notification in TUI (but don't return immediately - let other handlers run)
			alertCmd := a.createAlert(bubbleup.InfoKey, debugInfo)
			cmds = append(cmds, alertCmd)
		}

		// Handle other key commands
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "f11":
			// Toggle debug mode
			a.debugMode = !a.debugMode
			var message string
			if a.debugMode {
				message = "Debug mode ON - keys will be shown"
			} else {
				message = "Debug mode OFF"
			}
			alertCmd := a.createAlert(bubbleup.InfoKey, message)
			return a, alertCmd
		case "tab":
			a.nextPanel()
			return a, nil
		case "shift+tab":
			a.prevPanel()
			return a, nil
		case "escape":
			// If in menu binding mode, exit to normal mode
			if a.menuBindingMode {
				a.exitMenuMode()
				return a, nil
			}
		case "ctrl+s":
			generatedPrompt, err := prompt.Build(a.targetDir, a.fileTree.selected, a.chat.textarea.Value(), a.workspace.ActivePersonas)
			if err != nil {
				// Handle error, maybe show an error message
				// For now, we'll just log it
				// log.Printf("Error building prompt: %v", err)
			} else {
				a.promptDialog.Show(generatedPrompt)
			}
			return a, nil
		}

		// Handle menu-specific commands (only active in menu binding mode)
		if a.menuBindingMode {
			switch msg.String() {
			case a.settingsManager.GetPersonaMenuKey():
				// Show persona selection dialog
				a.personaDialog.SetActivePersonas(a.workspace.ActivePersonas)
				a.personaDialog.Show()
				return a, nil
			}
		}
	}

	// Update the alert model
	outAlert, outCmd := a.alertModel.Update(msg)
	a.alertModel = outAlert.(bubbleup.AlertModel)
	cmds = append(cmds, outCmd)

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
		chatModel, chatCmd := a.chat.Update(msg)
		a.chat = chatModel.(*ChatModel)
		// Check if the chat input has changed
		if a.workspace.ChatInput != a.chat.textarea.Value() {
			cmds = append(cmds, func() tea.Msg {
				return ChatInputMsg{Content: a.chat.textarea.Value()}
			})
		}
		cmds = append(cmds, chatCmd)
	default:
		// Handle invalid focus state - reset to FileTreePanel
		a.focused = FileTreePanel
		a.menuBindingMode = false
		model, cmd := a.fileTree.Update(msg)
		a.fileTree = model.(*FileTreeModel)
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

	// Show persona dialog if visible (takes priority over prompt dialog)
	if a.personaDialog.IsVisible() {
		dialogView := a.personaDialog.View()
		// Render with alert notifications
		return a.alertModel.Render(dialogView)
	}

	// Show prompt dialog if visible
	if a.promptDialog.IsVisible() {
		dialogView := a.promptDialog.View()
		// Render with alert notifications
		return a.alertModel.Render(dialogView)
	}

	// Render main layout with alert notifications
	return a.alertModel.Render(mainLayout)
}

func (a *App) mainLayout() string {
	// Calculate panel dimensions with header and footer
	headerHeight := 3 // Single line header with padding
	footerHeight := 3 // Single line footer with padding
	availableHeight := a.height - headerHeight - footerHeight
	topHeight := int(float64(availableHeight) * 0.66)
	bottomHeight := availableHeight - topHeight
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

	// Create header with persona information
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(a.width-2).
		Height(1).
		Padding(0, 2).
		BorderForeground(lipgloss.Color("240"))

	// Get active personas, default to "default" if none set
	activePersonas := a.workspace.ActivePersonas
	if len(activePersonas) == 0 {
		activePersonas = []string{"default"}
	}
	
	var headerContent string
	if len(activePersonas) == 1 {
		headerContent = "Persona: " + activePersonas[0]
	} else {
		headerContent = "Personas: " + strings.Join(activePersonas, ", ")
	}
	header := headerStyle.Render(headerContent)

	// Create footer with menu button
	footerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(a.width-2).
		Height(1).
		Padding(0, 2)
	
	// Apply focused style to footer if it has focus
	if a.focused == FooterMenuPanel {
		footerStyle = footerStyle.BorderForeground(lipgloss.Color("69"))
	} else {
		footerStyle = footerStyle.BorderForeground(lipgloss.Color("240"))
	}

	// Display appropriate menu activation key based on mode
	var menuActivationDisplay string
	if a.settingsManager.IsLegacyMode() {
		menuActivationDisplay = a.settingsManager.GetMenuActivationKey()
	} else {
		menuActivationDisplay = a.settingsManager.GetMenuModeActivation()
	}
	
	var debugInfo string
	if a.debugMode {
		debugInfo = " • F11: debug OFF"
	} else {
		debugInfo = " • F11: debug"
	}
	
	footerContent := "menu (" + menuActivationDisplay + ") • personas (" + a.settingsManager.GetPersonaMenuKey() + ")" + debugInfo
	footer := footerStyle.Render(footerContent)

	// Layout the panels
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, fileTreePanel, selectedPanel)

	return lipgloss.JoinVertical(lipgloss.Left, header, topRow, chatPanel, footer)
}

// createAlert creates an alert command with configured TTL
func (a *App) createAlert(alertType string, message string) tea.Cmd {
	// TODO: The bubbleup library doesn't currently support configurable TTL
	// For now, we just create the standard alert. The TTL configuration is ready
	// for when we either:
	// 1. Find a way to configure TTL in bubbleup
	// 2. Switch to a different notification library
	// 3. Create a custom notification system
	
	// Get the configured TTL (currently unused but ready)
	ttlSeconds := a.settingsManager.GetNotificationTTL()
	_ = ttlSeconds // Silence unused variable warning
	
	return a.alertModel.NewAlertCmd(alertType, message)
}

// nextPanel moves focus to the next panel
func (a *App) nextPanel() {
	switch a.focused {
	case FileTreePanel:
		a.focused = SelectedFilesPanel
	case SelectedFilesPanel:
		a.focused = ChatPanel
	case ChatPanel:
		a.focused = FooterMenuPanel
	case FooterMenuPanel:
		a.focused = FileTreePanel
	default:
		// Reset to FileTreePanel if focus state is invalid
		a.focused = FileTreePanel
	}
	// Update menu binding mode based on footer focus (legacy mode only)
	if a.settingsManager.IsLegacyMode() {
		a.menuBindingMode = (a.focused == FooterMenuPanel)
	}
}

// prevPanel moves focus to the previous panel
func (a *App) prevPanel() {
	switch a.focused {
	case FileTreePanel:
		a.focused = FooterMenuPanel
	case SelectedFilesPanel:
		a.focused = FileTreePanel
	case ChatPanel:
		a.focused = SelectedFilesPanel
	case FooterMenuPanel:
		a.focused = ChatPanel
	default:
		// Reset to FileTreePanel if focus state is invalid
		a.focused = FileTreePanel
	}
	// Update menu binding mode based on footer focus (legacy mode only)
	if a.settingsManager.IsLegacyMode() {
		a.menuBindingMode = (a.focused == FooterMenuPanel)
	}
}

// handleMouseClick determines which panel was clicked and sets focus accordingly
func (a *App) handleMouseClick(x, y int) {
	// Calculate panel dimensions - these must match mainLayout()
	headerHeight := 3 // Single line header with padding
	footerHeight := 3 // Single line footer with padding
	availableHeight := a.height - headerHeight - footerHeight
	topHeight := int(float64(availableHeight) * 0.66)
	bottomHeight := availableHeight - topHeight
	leftWidth := a.width / 2

	// Check if click is in the header area
	if y < headerHeight {
		// Header clicked - could add header focus support in the future
		return
	}
	// Check if click is in the top panel area (file tree or selected files panels)
	if y < headerHeight+topHeight {
		// Check if click is in the left half (file tree panel)
		if x < leftWidth {
			a.focused = FileTreePanel
		} else {
			// Click is in the right half (selected files panel)
			a.focused = SelectedFilesPanel
		}
	} else if y < headerHeight+topHeight+bottomHeight {
		// Click is in the chat area
		a.focused = ChatPanel
	} else if y < headerHeight+topHeight+bottomHeight+footerHeight {
		// Click is in the footer area
		a.focused = FooterMenuPanel
	}
	// Update menu binding mode based on footer focus (legacy mode only)
	if a.settingsManager.IsLegacyMode() {
		a.menuBindingMode = (a.focused == FooterMenuPanel)
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

// handleMenuActivation checks if the given key message should activate menu mode
// Returns true if menu mode was activated, false otherwise
func (a *App) handleMenuActivation(msg tea.KeyMsg) bool {
	// Debug: Always show what we're checking for
	if a.debugMode {
		legacyMode := a.settingsManager.IsLegacyMode()
		activationKey := a.settingsManager.GetMenuModeActivation()
		debugMsg := fmt.Sprintf("Menu check: Legacy=%v, Expected=%q, Got=%q", legacyMode, activationKey, msg.String())
		// Store debug info to show later since we can't return alert here
		a.lastDebugInfo = debugMsg
	}

	// Check for legacy mode (focus-based activation)
	if a.settingsManager.IsLegacyMode() {
		// Legacy mode: menu binding only works when footer has focus
		if a.focused == FooterMenuPanel && msg.String() == a.settingsManager.GetMenuActivationKey() {
			// In legacy mode, this just shows a notification since menu is already "active"
			return true
		}
		return false
	}

	// New mode: check for modifier-based activation
	activationKey := a.settingsManager.GetMenuModeActivation()
	if activationKey == "" {
		return false
	}

	keyCombination, err := config.ParseKeyBinding(activationKey)
	if err != nil {
		// Invalid key combination, ignore
		return false
	}

	if keyCombination.MatchesKeyMsg(msg) {
		a.enterMenuMode()
		return true
	}

	return false
}

// enterMenuMode activates menu binding mode
func (a *App) enterMenuMode() {
	a.menuBindingMode = true
	// Optionally focus the footer to provide visual feedback
	a.focused = FooterMenuPanel
}

// exitMenuMode deactivates menu binding mode  
func (a *App) exitMenuMode() {
	a.menuBindingMode = false
	// Return to chat panel for continued typing
	a.focused = ChatPanel
}

// initializeDebugLogger creates and configures a debug logger that writes to logs/error.log
func initializeDebugLogger(targetDir string) *log.Logger {
	logDir := filepath.Join(targetDir, "logs")
	logFile := filepath.Join(logDir, "error.log")
	
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// If we can't create the logs directory, return nil logger
		return nil
	}
	
	// Open log file in append mode, create if it doesn't exist
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// If we can't open the log file, return nil logger
		return nil
	}
	
	// Create logger with timestamp prefix
	logger := log.New(file, "", log.LstdFlags)
	
	// Log initialization message
	logger.Printf("=== Debug session started at %s ===", time.Now().Format("2006-01-02 15:04:05"))
	
	return logger
}
