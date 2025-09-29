# Available TUI components

## PromptDialogModel

**Location**: `internal/tui/prompt_dialog.go`

A reusable scrollable dialog component for displaying content in a modal overlay. This component provides a consistent way to show text content that may be too long to fit on screen at once.

### Features

- **Scrollable Content**: Automatically handles content that exceeds the viewport size
- **Auto-sizing**: Dialog takes 80% of terminal width/height for optimal readability
- **Word Wrapping**: Content automatically wraps to fit the viewport width
- **Scroll Indicators**: Shows percentage indicator when content is scrollable
- **Centered Display**: Dialog appears centered on screen
- **Modal Behavior**: Captures all input when visible and overlays the main UI

### Usage

```go
// Create and initialize
dialog := NewPromptDialogModel()
dialog.SetSize(terminalWidth, terminalHeight)

// Display content
dialog.Show("Your content here...")

// Handle in update loop
if dialog.IsVisible() {
    var cmd tea.Cmd
    dialog, cmd = dialog.Update(msg)
    return model, cmd
}

// Render in view
if dialog.IsVisible() {
    return dialog.View()
}
return normalView
```

### Key Bindings

When the dialog is visible:
- **Arrow Keys, Page Up/Down, Home/End**: Navigate through content
- **Ctrl+C, Q, Enter, Escape**: Close the dialog

### API Reference

#### Constructor
- `NewPromptDialogModel() *PromptDialogModel`: Creates a new dialog instance

#### Methods
- `SetSize(width, height int)`: Set dialog dimensions for proper centering
- `Show(content string)`: Display the dialog with the given content
- `Hide()`: Close the dialog
- `IsVisible() bool`: Check if dialog is currently shown
- `GetContent() string`: Get the current dialog content
- `Update(tea.Msg) (*PromptDialogModel, tea.Cmd)`: Handle Bubble Tea messages
- `View() string`: Render the dialog

### Implementation Notes

- Uses Bubble Tea's viewport component for scrolling functionality
- Automatically resets scroll position to top when new content is shown
- Content is word-wrapped using Lipgloss styling
- Dialog has a thick border with yellow accent color (Color "228")
- Scroll percentage appears in bottom-right when content overflows

### Use Cases

- Displaying generated prompts for review
- Showing persona descriptions or help text
- Any text content that needs scrollable, modal presentation
- Content that may vary in length and needs consistent presentation
