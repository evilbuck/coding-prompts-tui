package tui

// Example usage of the reusable Dialog component
// This file demonstrates how to use the Dialog with different content types

import tea "github.com/charmbracelet/bubbletea"

// ExampleListDialog shows how to create a list dialog
func ExampleListDialog() *Dialog {
	// Create list items
	items := []ListItem{
		{Label: "Option 1", Value: "opt1", Selected: false},
		{Label: "Option 2", Value: "opt2", Selected: true},
		{Label: "Option 3", Value: "opt3", Selected: false},
	}

	// Create list content with multi-select and callback
	listContent := NewListContent(items, true, func(selectedItems []ListItem) tea.Msg {
		return ListSelectMsg{Items: selectedItems}
	})

	// Create dialog configuration
	config := DefaultDialogConfig()
	config.Title = "Select Options"
	config.HelpText = "Space: Toggle • Enter: Apply • Escape: Cancel"
	// config.Alignment = lipgloss.Center // Optional: center the content

	// Create and return the dialog
	return NewDialog(config, listContent)
}

// ExampleTextDialog shows how to create a text display dialog
func ExampleTextDialog(content string) *Dialog {
	// Create text content
	textContent := NewTextContent(content)

	// Create dialog configuration
	config := DefaultDialogConfig()
	config.Title = "View Content"
	config.Width = 60
	config.HelpText = "↑/↓: Scroll • Escape: Close"

	// Create and return the dialog
	return NewDialog(config, textContent)
}

// ExampleSimpleDialog shows the simplest way to create a dialog
func ExampleSimpleDialog() *Dialog {
	// Create simple text content
	textContent := NewTextContent("This is a simple dialog with default settings.")

	// Create dialog with default config
	return NewSimpleDialog("Simple Dialog", textContent)
}

// ExampleUsageInApp demonstrates how to integrate the dialog into an app
func ExampleUsageInApp() {
	/*
		In your main app struct, add:

		type App struct {
			// ... other fields
			dialog *Dialog
		}

		In NewApp():

		app := &App{
			// ... other initialization
			dialog: ExampleListDialog(),
		}

		In Update():

		// Handle dialog input if visible
		if a.dialog.IsVisible() {
			model, cmd := a.dialog.Update(msg)
			a.dialog = model
			return a, cmd
		}

		// Handle other keys to show dialog
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "d": // Show dialog
				a.dialog.Show()
				return a, nil
			}
		}

		In View():

		// Show dialog if visible
		if a.dialog.IsVisible() {
			return a.dialog.ViewAsSimpleOverlay(mainLayout)
		}
		return mainLayout

		In Init():

		return tea.Batch(
			// ... other initialization commands
			a.dialog.Init(),
		)
	*/
}