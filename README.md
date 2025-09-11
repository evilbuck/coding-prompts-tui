# LLM Coding Prompt Builder

A terminal user interface (TUI) tool that helps developers build structured prompts with code context for AI assistants like Claude, ChatGPT, and Gemini.

## Overview

This tool allows you to:
- Browse your project files in an interactive file tree
- Select specific files to include in your prompt
- Write custom user prompts in a dedicated chat area
- Generate XML-formatted prompts with file context for better AI interactions

## Features

- **Three-Panel Grid Layout**: File tree, selected files, and chat input areas
- **Interactive File Selection**: Browse directories and select individual files
- **Tab Navigation**: Seamless switching between interface panels
- **XML Output**: Structured prompts with file tree, file contents, system prompts, and user input
- **Smart File Filtering**: Automatically ignores common build artifacts and hidden files

## Installation

### Prerequisites

- Go 1.21 or later
- Terminal with color support

### Build from Source

1. Clone or download this repository:
   ```bash
   git clone <repository-url>
   cd coding-prompts-tui
   ```

2. Build the application:
   ```bash
   go build -o prompter main.go
   ```

3. (Optional) Install globally:
   ```bash
   go install .
   ```

### Dependencies

The application uses the following Go modules:
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/charmbracelet/lipgloss` - Styling and layout

Dependencies will be automatically downloaded when you run `go build` or `go mod tidy`.

## Usage

### Basic Usage

Run the application with a target directory:

```bash
./prompter <directory>
```

Examples:
```bash
# Use current directory
./prompter .

# Use specific project directory
./prompter /path/to/your/project

# Use relative path
./prompter ../my-project
```

### Interface Layout

The TUI consists of three main panels:

```
┌─────────────────────────────────────────────────────────┐
│  📁 File Tree          │  ✅ Selected Files             │
│                        │                                │
│  Browse and select     │  View and manage               │
│  project files         │  selected files                │
│                        │                                │
│                        │                                │
├─────────────────────────────────────────────────────────┤
│  💬 User Prompt                                         │
│                                                         │
│  Enter your prompt for the LLM here...                 │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Navigation

#### Panel Navigation
- **Tab** - Move to next panel (File Tree → Selected Files → Chat → File Tree)
- **Shift+Tab** - Move to previous panel

#### File Tree Panel
- **↑/↓ Arrow Keys** - Navigate up/down through files and folders
- **Enter** - Expand/collapse folders
- **Space** - Select/deselect files (files only, not folders)

#### Selected Files Panel
- **↑/↓ Arrow Keys** - Navigate through selected files
- **x** or **Delete/Backspace** - Remove file from selection

#### Chat Panel
- **Type** - Enter your prompt text
- **Ctrl+S** - Generate XML prompt (future feature)

#### Global Controls
- **Ctrl+C** or **q** - Quit the application

### File Selection

1. **Navigate** to files using arrow keys in the File Tree panel
2. **Expand folders** by pressing Enter on directory items
3. **Select files** by pressing Space on file items (not directories)
4. **View selections** in the Selected Files panel
5. **Remove files** from selection using x/Delete in the Selected Files panel

### Generated Output Format

The application generates XML-structured prompts in this format:

```xml
<filetree>
[Complete directory tree structure]
</filetree>

<file name="path/to/selected/file.go">
[File contents]
</file>

<file name="path/to/another/file.js">
[Another file's contents]
</file>

<SystemPrompt>
[Content from personas/default.md]
</SystemPrompt>

<UserPrompt>
[Your custom prompt text]
</UserPrompt>
```

## File Filtering

The application automatically ignores common files and directories:
- Hidden files (starting with `.`)
- `node_modules/`
- `.git/`
- Build artifacts (`build/`, `dist/`, `target/`)
- Package caches (`__pycache__/`, `vendor/`)
- OS files (`.DS_Store`, `Thumbs.db`)

## System Requirements

- **Operating System**: Linux, macOS, Windows
- **Terminal**: Any terminal with ANSI color support
- **Memory**: Minimal (typically < 50MB)
- **Disk Space**: < 10MB for the binary

## Troubleshooting

### Build Issues

If you encounter build errors:

1. Verify Go version:
   ```bash
   go version  # Should be 1.21+
   ```

2. Clean and rebuild:
   ```bash
   go clean
   go mod tidy
   go build -o prompter main.go
   ```

### Runtime Issues

**"No such file or directory" error:**
- Ensure the target directory exists
- Use absolute paths if relative paths don't work
- Check directory permissions

**Display issues:**
- Ensure your terminal supports ANSI colors
- Try a different terminal if the interface appears garbled
- Resize terminal window if layout seems cramped

**Performance issues with large projects:**
- The application filters common build artifacts automatically
- For very large codebases, consider targeting specific subdirectories

## Development

### Project Structure

```
.
├── main.go                     # Application entry point
├── go.mod                      # Go module definition
├── internal/
│   ├── tui/                   # TUI components
│   │   ├── app.go            # Main application model
│   │   ├── filetree.go       # File tree panel
│   │   ├── selected.go       # Selected files panel
│   │   └── chat.go           # Chat input panel
│   └── filesystem/
│       └── scanner.go        # Directory scanning utilities
├── personas/
│   └── default.md           # Default system prompt
└── README.md               # This file
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

[Add your license information here]

## Roadmap

- [ ] XML prompt generation and export
- [ ] Custom system prompt selection
- [ ] File content preview
- [ ] Search functionality
- [ ] Configuration file support
- [ ] Multiple export formats
- [ ] Syntax highlighting in file preview
# coding-prompts-tui
# coding-prompts-tui
