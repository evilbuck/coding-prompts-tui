package tui

import (
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DialogConfig contains configuration options for a dialog
type DialogConfig struct {
	Title       string
	Width       int    // 0 = auto-size
	Height      int    // 0 = auto-size
	Dismissible bool   // Can be dismissed with escape key
	ShowHelp    bool   // Show help text at bottom
	HelpText    string // Custom help text, defaults to "Escape: Close"
}

// DefaultDialogConfig returns sensible defaults for a dialog
func DefaultDialogConfig() DialogConfig {
	return DialogConfig{
		Title:       "",
		Width:       0,
		Height:      0,
		Dismissible: true,
		ShowHelp:    true,
		HelpText:    "Escape: Close",
	}
}

// Dialog represents a reusable modal dialog component
type Dialog struct {
	config      DialogConfig
	content     DialogContent
	visible     bool
	width       int
	height      int
	debugLogger *log.Logger
}

// NewDialog creates a new dialog with the specified configuration and content
func NewDialog(config DialogConfig, content DialogContent) *Dialog {
	return &Dialog{
		config:  config,
		content: content,
		visible: false,
	}
}

// NewSimpleDialog creates a dialog with default config and the given title and content
func NewSimpleDialog(title string, content DialogContent) *Dialog {
	config := DefaultDialogConfig()
	config.Title = title
	return NewDialog(config, content)
}

// SetContent updates the dialog content
func (d *Dialog) SetContent(content DialogContent) {
	d.content = content
}

// SetConfig updates the dialog configuration
func (d *Dialog) SetConfig(config DialogConfig) {
	d.config = config
}

// Show displays the dialog
func (d *Dialog) Show() {
	d.visible = true
	if d.content != nil {
		d.content.OnShow()
	}
}

// Hide closes the dialog
func (d *Dialog) Hide() {
	d.visible = false
	if d.content != nil {
		d.content.OnHide()
	}
}

// IsVisible returns whether the dialog is currently shown
func (d *Dialog) IsVisible() bool {
	return d.visible
}

// SetSize sets the dialog size for centering calculations
func (d *Dialog) SetSize(width, height int) {
	d.width = width
	d.height = height
	if d.content != nil {
		// Calculate available content area (minus borders and padding)
		contentWidth := width - 6 // Account for borders and padding
		contentHeight := height - 6
		d.content.SetSize(contentWidth, contentHeight)
	}
}

// SetDebugLogger sets the debug logger for the dialog
func (d *Dialog) SetDebugLogger(logger *log.Logger) {
	d.debugLogger = logger
}

// Init initializes the dialog
func (d *Dialog) Init() tea.Cmd {
	if d.content != nil {
		return d.content.Init()
	}
	return nil
}

// Update handles messages for the dialog
func (d *Dialog) Update(msg tea.Msg) (*Dialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if d.debugLogger != nil {
			d.debugLogger.Printf("DIALOG: Key pressed: %q", msg.String())
		}

		// Handle escape key dismissal if enabled
		if d.config.Dismissible && msg.String() == "esc" {
			if d.debugLogger != nil {
				d.debugLogger.Printf("DIALOG: Escape key pressed - hiding dialog")
			}
			d.Hide()
			return d, nil
		}

		// Let content handle other keys
		if d.content != nil {
			cmd := d.content.Update(msg)
			return d, cmd
		}
	}

	return d, nil
}

// View renders the dialog (for backwards compatibility)
func (d *Dialog) View() string {
	if !d.visible {
		return ""
	}
	return d.renderDialog()
}

// ViewAsOverlay renders the dialog as a full-screen overlay with dimmed backdrop
func (d *Dialog) ViewAsOverlay(backgroundContent string) string {
	if !d.visible {
		return backgroundContent
	}

	// Create dimmed backdrop
	backdrop := d.createDimmedBackdrop(backgroundContent)

	// Get the dialog content
	dialog := d.renderDialog()

	// Center the dialog over the backdrop
	return d.centerDialog(backdrop, dialog)
}

// ViewAsSimpleOverlay renders the dialog using a simpler overlay method
func (d *Dialog) ViewAsSimpleOverlay(backgroundContent string) string {
	if !d.visible {
		return backgroundContent
	}

	return d.overlayDialog(backgroundContent, d.renderDialog())
}

// renderDialog creates the dialog content with styling
func (d *Dialog) renderDialog() string {
	if d.content == nil {
		return ""
	}

	// Get content from the content provider
	contentStr := d.content.Render()

	// Add title if specified
	if d.config.Title != "" {
		contentStr = d.config.Title + "\n\n" + contentStr
	}

	// Add help text if enabled
	if d.config.ShowHelp && d.config.HelpText != "" {
		contentStr = contentStr + "\n\n" + d.config.HelpText
	}

	// Calculate dialog width
	dialogWidth := d.config.Width
	if dialogWidth == 0 {
		// Auto-size based on content or screen
		if d.width > 0 {
			dialogWidth = int(float64(d.width) * 0.6) // 60% of screen width
		} else {
			dialogWidth = 50 // Fallback
		}
	}

	// Create dialog box style
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("69")).
		Background(lipgloss.Color("0")).
		Foreground(lipgloss.Color("15")).
		Padding(1, 2).
		Width(dialogWidth).
		Align(lipgloss.Left)

	return dialogStyle.Render(contentStr)
}

// createDimmedBackdrop applies a dimming effect to background content
func (d *Dialog) createDimmedBackdrop(content string) string {
	lines := strings.Split(content, "\n")
	dimmedLines := make([]string, len(lines))

	// Style to dim the background content
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")) // Dark gray

	for i, line := range lines {
		dimmedLines[i] = dimStyle.Render(line)
	}

	return strings.Join(dimmedLines, "\n")
}

// centerDialog positions the dialog in the center of the backdrop
func (d *Dialog) centerDialog(backdrop, dialog string) string {
	return lipgloss.Place(d.width, d.height, lipgloss.Center, lipgloss.Center,
		dialog,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("237")),
	)
}

// overlayDialog overlays the dialog onto the background using the simple method
func (d *Dialog) overlayDialog(backgroundContent, dialog string) string {
	lines := strings.Split(backgroundContent, "\n")

	// Dim all background lines
	for i, line := range lines {
		lines[i] = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(line)
	}

	// Get dialog lines
	dialogLines := strings.Split(dialog, "\n")

	// Calculate where to place the dialog (center)
	startRow := (len(lines) - len(dialogLines)) / 2
	if startRow < 0 {
		startRow = 0
	}

	// Overlay dialog lines
	for i, dialogLine := range dialogLines {
		targetRow := startRow + i
		if targetRow < len(lines) {
			// Center the dialog line
			if d.width > 0 {
				dialogWidth := lipgloss.Width(dialogLine)
				padding := (d.width - dialogWidth) / 2
				if padding > 0 {
					lines[targetRow] = strings.Repeat(" ", padding) + dialogLine
				} else {
					lines[targetRow] = dialogLine
				}
			} else {
				lines[targetRow] = dialogLine
			}
		}
	}

	return strings.Join(lines, "\n")
}