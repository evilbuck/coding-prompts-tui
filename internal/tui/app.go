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

// State change messages for reactive system
type FocusChangeMsg struct {
	Panel FocusedPanel
}

type MenuModeChangeMsg struct {
	Enabled bool
}

type DebugModeChangeMsg struct {
	Enabled bool
}

type LayoutChangeMsg struct {
	Width  int
	Height int
}

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
	layoutConfig    *LayoutConfig
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
	debugLogger := initializeDebugLogger(targetDir, settingsManager)

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
		debugMode:       settingsManager.IsDebugEnabled(), // Set from config
		debugLogger:     debugLogger,
		layoutConfig:    NewLayoutConfig(),
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

	// Handle state change messages first with centralized validation
	if stateModel, stateCmd := a.handleStateChange(msg); stateCmd != nil {
		cmds = append(cmds, stateCmd)
		return stateModel, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Use reactive pattern for layout changes
		return a, a.updateLayout(msg.Width, msg.Height)

	case tea.MouseMsg:
		// Handle mouse clicks for panel focus
		if msg.Type == tea.MouseLeft {
			if mouseCmd := a.handleMouseClick(msg.X, msg.Y); mouseCmd != nil {
				cmds = append(cmds, mouseCmd)
			}
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

	// Bindings
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
		if menuCmd := a.handleMenuActivation(msg); menuCmd != nil {
			return a, menuCmd
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

		// Check for debug toggle key
		debugToggleKey := a.settingsManager.GetDebugToggleKey()
		if debugKeyCombination, err := config.ParseKeyBinding(debugToggleKey); err == nil && debugKeyCombination.MatchesKeyMsg(msg) {
			// Toggle debug mode using reactive pattern
			return a, a.toggleDebugMode()
		}

		// Handle other key commands
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		case "tab":
			return a, a.nextPanel()
		case "shift+tab":
			return a, a.prevPanel()
		case "escape":
			// If in menu binding mode, exit to normal mode
			if a.menuBindingMode {
				return a, a.exitMenuMode()
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
	// Calculate panel dimensions using layout config
	topHeight := a.layoutConfig.TopPanelHeight(a.height)
	bottomHeight := a.layoutConfig.BottomPanelHeight(a.height)
	leftWidth := a.layoutConfig.LeftPanelWidth(a.width)
	rightWidth := a.layoutConfig.RightPanelWidth(a.width)

	// Create styles for panels
	focusedBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69"))

	normalBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	// File tree panel (top-left)
	fileTreePanel := CreatePanel(
		a.fileTree.View(),
		a.focused == FileTreePanel,
		normalBorder,
		focusedBorder,
		StretchWidth(leftWidth, true),
		StretchHeight(topHeight, true),
	)

	// Chat panel (top-right)
	chatPanel := CreatePanel(
		a.chat.View(),
		a.focused == ChatPanel,
		normalBorder,
		focusedBorder,
		StretchWidth(rightWidth, true),
		StretchHeight(topHeight, true),
	)

	// Selected files panel (middle row, full width)
	selectedPanel := CreatePanel(
		a.selectedFiles.View(),
		a.focused == SelectedFilesPanel,
		normalBorder,
		focusedBorder,
		StretchWidth(a.width, true),
		StretchHeight(bottomHeight, true),
	)

	// Create header with persona information
	headerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(StretchWidth(a.width, true)).
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
		Width(StretchWidth(a.width, true)).
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
	debugToggleKey := a.settingsManager.GetDebugToggleKey()
	if a.debugMode {
		debugInfo = fmt.Sprintf(" • %s: debug OFF", debugToggleKey)
	} else {
		debugInfo = fmt.Sprintf(" • %s: debug", debugToggleKey)
	}

	footerContent := "menu (" + menuActivationDisplay + ") • personas (" + a.settingsManager.GetPersonaMenuKey() + ")" + debugInfo
	footer := footerStyle.Render(footerContent)

	// Layout the panels
	// topRow := lipgloss.JoinHorizontal(lipgloss.Top, fileTreePanel, selectedPanel)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, fileTreePanel, chatPanel)

	// return lipgloss.JoinVertical(lipgloss.Left, header, topRow, chatPanel, footer)
	return lipgloss.JoinVertical(lipgloss.Left, header, topRow, selectedPanel, footer)
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

// nextPanel returns a command to move focus to the next panel
func (a *App) nextPanel() tea.Cmd {
	var nextFocus FocusedPanel
	switch a.focused {
	case FileTreePanel:
		nextFocus = SelectedFilesPanel
	case SelectedFilesPanel:
		nextFocus = ChatPanel
	case ChatPanel:
		nextFocus = FooterMenuPanel
	case FooterMenuPanel:
		nextFocus = FileTreePanel
	default:
		// Reset to FileTreePanel if focus state is invalid
		nextFocus = FileTreePanel
	}
	return a.setFocus(nextFocus)
}

// prevPanel returns a command to move focus to the previous panel
func (a *App) prevPanel() tea.Cmd {
	var prevFocus FocusedPanel
	switch a.focused {
	case FileTreePanel:
		prevFocus = FooterMenuPanel
	case SelectedFilesPanel:
		prevFocus = FileTreePanel
	case ChatPanel:
		prevFocus = SelectedFilesPanel
	case FooterMenuPanel:
		prevFocus = ChatPanel
	default:
		// Reset to FileTreePanel if focus state is invalid
		prevFocus = FileTreePanel
	}
	return a.setFocus(prevFocus)
}

// handleMouseClick determines which panel was clicked and returns a command to set focus
func (a *App) handleMouseClick(x, y int) tea.Cmd {
	// Calculate panel dimensions using layout config - matches mainLayout()
	headerHeight := a.layoutConfig.HeaderHeight
	footerHeight := a.layoutConfig.FooterHeight
	topHeight := a.layoutConfig.TopPanelHeight(a.height)
	bottomHeight := a.layoutConfig.BottomPanelHeight(a.height)
	leftWidth := a.layoutConfig.LeftPanelWidth(a.width)

	// Check if click is in the header area
	if y < headerHeight {
		// Header clicked - could add header focus support in the future
		return nil
	}

	var targetFocus FocusedPanel
	// Check if click is in the top panel area (file tree or selected files panels)
	if y < headerHeight+topHeight {
		// Check if click is in the left half (file tree panel)
		if x < leftWidth {
			targetFocus = FileTreePanel
		} else {
			// Click is in the right half (selected files panel)
			targetFocus = SelectedFilesPanel
		}
	} else if y < headerHeight+topHeight+bottomHeight {
		// Click is in the chat area
		targetFocus = ChatPanel
	} else if y < headerHeight+topHeight+bottomHeight+footerHeight {
		// Click is in the footer area
		targetFocus = FooterMenuPanel
	} else {
		// Click outside all areas
		return nil
	}

	return a.setFocus(targetFocus)
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
// Returns a command if menu mode should be activated, nil otherwise
func (a *App) handleMenuActivation(msg tea.KeyMsg) tea.Cmd {
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
			return a.createAlert("info", "menu mode activated")
		}
		return nil
	}

	// New mode: check for modifier-based activation
	activationKey := a.settingsManager.GetMenuModeActivation()
	if activationKey == "" {
		return nil
	}

	keyCombination, err := config.ParseKeyBinding(activationKey)
	if err != nil {
		// Invalid key combination, ignore
		return nil
	}

	if keyCombination.MatchesKeyMsg(msg) {
		return tea.Batch(
			a.enterMenuMode(),
			a.createAlert("info", "menu mode activated"),
		)
	}

	return nil
}

// enterMenuMode returns a command to activate menu binding mode
func (a *App) enterMenuMode() tea.Cmd {
	return a.setMenuMode(true)
}

// exitMenuMode returns a command to deactivate menu binding mode
func (a *App) exitMenuMode() tea.Cmd {
	return a.setMenuMode(false)
}

// initializeDebugLogger creates and configures a debug logger based on settings
func initializeDebugLogger(targetDir string, settingsManager *config.SettingsManager) *log.Logger {
	// Only initialize logger if file logging is enabled
	if !settingsManager.IsDebugFileLoggingEnabled() {
		return nil
	}

	// Get log file path from config
	logFilePath := settingsManager.GetDebugLogFile()
	fullLogPath := filepath.Join(targetDir, logFilePath)
	logDir := filepath.Dir(fullLogPath)

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		// If we can't create the logs directory, return nil logger
		return nil
	}

	// Open log file in append mode, create if it doesn't exist
	file, err := os.OpenFile(fullLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
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

// State command generators for reactive system
func (a *App) setFocus(panel FocusedPanel) tea.Cmd {
	return func() tea.Msg {
		return FocusChangeMsg{Panel: panel}
	}
}

func (a *App) setMenuMode(enabled bool) tea.Cmd {
	return func() tea.Msg {
		return MenuModeChangeMsg{Enabled: enabled}
	}
}

func (a *App) toggleDebugMode() tea.Cmd {
	return func() tea.Msg {
		return DebugModeChangeMsg{Enabled: !a.debugMode}
	}
}

func (a *App) updateLayout(width, height int) tea.Cmd {
	return func() tea.Msg {
		return LayoutChangeMsg{Width: width, Height: height}
	}
}

// handleStateChange processes all state change messages with validation
func (a *App) handleStateChange(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case FocusChangeMsg:
		// Validate focus change
		if !a.isValidPanel(msg.Panel) {
			if a.debugMode {
				cmds = append(cmds, a.createAlert("error", "Invalid focus panel"))
			}
			return a, tea.Batch(cmds...)
		}

		// Only change if different
		if a.focused != msg.Panel {
			oldFocus := a.focused
			a.focused = msg.Panel

			// Update dependent state: menu binding mode in legacy mode
			if a.settingsManager.IsLegacyMode() {
				oldMenuMode := a.menuBindingMode
				a.menuBindingMode = (msg.Panel == FooterMenuPanel)

				// Debug log state changes
				if a.debugMode && a.debugLogger != nil {
					a.debugLogger.Printf("STATE: Focus changed %v→%v, MenuMode %v→%v",
						oldFocus, a.focused, oldMenuMode, a.menuBindingMode)
				}
			}
		}

	case MenuModeChangeMsg:
		// Only change if different
		if a.menuBindingMode != msg.Enabled {
			oldMenuMode := a.menuBindingMode
			a.menuBindingMode = msg.Enabled

			// Update dependent state: focus to footer when enabling menu mode
			if msg.Enabled {
				oldFocus := a.focused
				a.focused = FooterMenuPanel

				// Debug log state changes
				if a.debugMode && a.debugLogger != nil {
					a.debugLogger.Printf("STATE: MenuMode %v→%v, Focus %v→%v",
						oldMenuMode, a.menuBindingMode, oldFocus, a.focused)
				}
			}
		}

	case DebugModeChangeMsg:
		// Only change if different
		if a.debugMode != msg.Enabled {
			oldDebugMode := a.debugMode
			a.debugMode = msg.Enabled

			// Debug log state changes (before mode is disabled)
			if a.debugLogger != nil {
				a.debugLogger.Printf("STATE: DebugMode %v→%v", oldDebugMode, a.debugMode)
			}

			// Show notification about debug mode change
			var message string
			if msg.Enabled {
				message = "Debug mode ON - keys will be shown"
			} else {
				message = "Debug mode OFF"
			}
			cmds = append(cmds, a.createAlert("info", message))
		}

	case LayoutChangeMsg:
		// Validate layout dimensions
		if msg.Width <= 0 || msg.Height <= 0 {
			if a.debugMode {
				cmds = append(cmds, a.createAlert("error", "Invalid layout dimensions"))
			}
			return a, tea.Batch(cmds...)
		}

		// Only update if different
		if a.width != msg.Width || a.height != msg.Height {
			oldWidth, oldHeight := a.width, a.height
			a.width = msg.Width
			a.height = msg.Height

			// Update dialogs with new size
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
			headerHeight := 3 // Single line header with padding
			footerHeight := 3 // Single line footer with padding
			availableHeight := a.height - headerHeight - footerHeight
			topHeight := int(float64(availableHeight) * 0.66)
			leftWidth := a.width / 2
			contentWidth := leftWidth - 2 - 2  // border width minus border padding
			contentHeight := topHeight - 2 - 2 // border height minus border padding
			a.fileTree.SetSize(contentWidth, contentHeight)

			// Debug log state changes
			if a.debugMode && a.debugLogger != nil {
				a.debugLogger.Printf("STATE: Layout changed %dx%d→%dx%d",
					oldWidth, oldHeight, a.width, a.height)
			}
		}

	default:
		// Not a state change message, return unchanged
		return a, nil
	}

	return a, tea.Batch(cmds...)
}

// State validation helpers
func (a *App) isValidPanel(panel FocusedPanel) bool {
	return panel >= FileTreePanel && panel <= FooterMenuPanel
}

// validateStateInvariants checks that the current state is consistent
func (a *App) validateStateInvariants() error {
	// Check focus is valid
	if !a.isValidPanel(a.focused) {
		return fmt.Errorf("invalid focus panel: %v", a.focused)
	}

	// Check menu binding mode consistency in legacy mode
	if a.settingsManager.IsLegacyMode() {
		expectedMenuMode := (a.focused == FooterMenuPanel)
		if a.menuBindingMode != expectedMenuMode {
			return fmt.Errorf("menu binding mode %v inconsistent with focus %v in legacy mode",
				a.menuBindingMode, a.focused)
		}
	}

	// Check layout dimensions are valid
	if a.width < 0 || a.height < 0 {
		return fmt.Errorf("invalid layout dimensions: %dx%d", a.width, a.height)
	}

	return nil
}
