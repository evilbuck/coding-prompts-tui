package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"coding-prompts-tui/internal/tui"
)

func main() {
	// Check for directory argument
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s .\n", os.Args[0])
		os.Exit(1)
	}

	targetDir := os.Args[1]
	
	// Verify directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Directory '%s' does not exist\n", targetDir)
		os.Exit(1)
	}

	// Initialize TUI application
	app := tui.NewApp(targetDir)
	
	// Create Bubble Tea program
	p := tea.NewProgram(app, tea.WithAltScreen())
	
	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}