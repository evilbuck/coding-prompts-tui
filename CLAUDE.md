# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based TUI (Terminal User Interface) application called "LLM Coding Prompt Builder" that helps developers create structured prompts with code context for AI assistants. The tool uses Bubble Tea for the TUI interface and allows users to select files from their project directory to include in XML-formatted prompts.

## Purpose

The application enables developers to:
- Navigate and select files from their codebase using a TUI interface
- Generate structured XML prompts containing file trees, file contents, system prompts, and user prompts
- Create context-rich prompts for AI assistants (Claude, ChatGPT, Gemini) with relevant codebase information

## Architecture

### Core Components

**TUI Interface Structure**:
- **Chat Area**: User input area for writing custom prompts
- **File Selection**: Tree-based file browser with checkboxes for individual file selection
- **Navigation**: Tab-based switching between interface areas
- **Output**: XML-structured prompt generation

**Expected Directory Structure**:
```
.
├── main.go              # Application entry point
├── cmd/                 # CLI command structure
├── internal/
│   ├── tui/            # Bubble Tea TUI components
│   │   ├── filetree.go # File selection tree component
│   │   ├── chat.go     # Chat/prompt input component
│   │   └── app.go      # Main TUI application
│   ├── prompt/         # Prompt generation logic
│   │   └── builder.go  # XML prompt builder
│   └── filesystem/     # File system utilities
│       └── scanner.go  # Directory scanning utilities
├── personas/           # System prompt templates
│   └── default.md      # Default system prompt
├── go.mod
├── go.sum
└── README.md
```

## Common Development Commands

### Building and Running
```bash
# Build the application
go build -o prompter ./cmd/prompter

# Run directly
go run ./cmd/prompter .

# Install locally
go install ./cmd/prompter

# Run with current directory
./prompter .
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/tui/
go test ./internal/prompt/
```

### Development
```bash
# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Tidy dependencies
go mod tidy

# Vendor dependencies (if needed)
go mod vendor
```

## Key Implementation Details

### TUI Navigation
- **Tab Navigation**: Use Tab key to switch between Chat Area and File Selection
- **File Selection**: Arrow keys for navigation, Space bar for file selection
- **No Folder Selection**: MVP requires explicit file selection only, folders cannot be checked

### XML Output Structure
The application generates XML in this format:
```xml
<filetree>
[Directory tree structure]
</filetree>

<file name="relative/path/to/file.go">
[File contents]
</file>

<SystemPrompt>
[Content from personas/default.md]
</SystemPrompt>

<UserPrompt>
[User input from Chat Area]
</UserPrompt>
```

### Dependencies
Expected Go dependencies:
- **Bubble Tea**: For TUI framework (`github.com/charmbracelet/bubbletea`)
- **Bubbles**: For TUI components (`github.com/charmbracelet/bubbles`)
- **Lipgloss**: For styling (`github.com/charmbracelet/lipgloss`)

### File System Integration
- **Working Directory**: Application uses current working directory as project root
- **File Tree Generation**: Recursively scans directories to build file tree
- **Path Resolution**: All file paths are relative to the project root
- **File Reading**: Selected files are read and embedded in XML output

## Development Workflow

1. **Start with TUI Framework**: Set up Bubble Tea application structure
2. **Implement File Tree**: Create file system scanning and tree display
3. **Add File Selection**: Implement checkbox selection for individual files
4. **Create Chat Interface**: Add text input area for user prompts
5. **Build XML Generator**: Implement prompt builder with XML structure
6. **Add System Prompts**: Integrate personas/default.md system prompt
7. **Polish Navigation**: Ensure smooth tab navigation between areas

## Testing Considerations

- **TUI Testing**: Focus on model state changes and update logic
- **File System Mocking**: Mock file system operations for consistent tests
- **XML Generation**: Validate XML structure and content accuracy
- **Edge Cases**: Handle empty directories, unreadable files, large files

## Configuration

### System Prompts
- Default system prompt stored in `personas/default.md`
- System should be extensible for additional personas
- Prompts should be configurable and easily modified

### File Filtering
- Consider implementing file type filtering (e.g., ignore binary files)
- Respect common ignore patterns (.git, node_modules, etc.)
- Allow configuration of maximum file size limits