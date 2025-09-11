package main

import (
	"fmt"
	"os"
	"path/filepath"

	"coding-prompts-tui/internal/config"
	"coding-prompts-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
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

	// Get absolute path for workspace management
	absPath, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting absolute path: %v\n", err)
		os.Exit(1)
	}

	// Initialize config manager
	cfgManager, err := config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config manager: %v\n", err)
		os.Exit(1)
	}

	// Get the workspace state
	workspace := cfgManager.GetWorkspace(absPath)

	// Initialize TUI application
	app := tui.NewApp(absPath, cfgManager, workspace)

	// Create Bubble Tea program with alt screen and mouse support
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}
