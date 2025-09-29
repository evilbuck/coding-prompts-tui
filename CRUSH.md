# CRUSH.md - Coding Prompts TUI Development Guide

## Build/Test Commands

### Building

```bash
# Build the application
go build -o prompter main.go
# or
go build -o prompter ./cmd/prompter

# Install globally
go install .
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/tui/
go test ./internal/config/
go test ./internal/prompt/

# Run a single test
go test -run TestParseKeyBinding ./internal/config/
go test -run TestStateCommandGeneration ./internal/tui/
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Tidy dependencies
go mod tidy
```

## Code Style Guidelines

### Imports

- Standard library imports first
- Third-party imports second
- Internal imports last
- Blank line between groups

```go
import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/charmbracelet/bubbletea"

    "coding-prompts-tui/internal/config"
)
```

### Naming Conventions

- **Exported functions/types**: PascalCase (`NewApp`, `ConfigManager`)
- **Unexported functions/types**: camelCase (`newApp`, `configManager`)
- **Constants**: PascalCase for exported, camelCase for internal
- **Variables**: camelCase, descriptive names
- **Receiver names**: Single letter (`m` for models, `a` for app, etc.)

### Error Handling

- Return errors with `fmt.Errorf` and `%w` for wrapping
- Check errors immediately after operations
- Use descriptive error messages

```go
content, err := os.ReadFile(path)
if err != nil {
    return "", fmt.Errorf("error reading file %s: %w", path, err)
}
```

### Structs and Types

- Clear, descriptive field names
- Use appropriate JSON/XML tags
- Group related fields together

```go
type App struct {
    targetDir       string
    width           int
    height          int
    focused         FocusedPanel
    configManager   *config.ConfigManager
}
```

### Functions

- Clear, descriptive names
- Keep functions focused on single responsibility
- Use early returns for error conditions

### Testing

- Use table-driven tests for multiple test cases
- Use `t.Run` for subtests with descriptive names
- Test both success and error paths
- Mock external dependencies when needed

```go
func TestParseKeyBinding(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    *KeyCombination
        expectError bool
    }{
        // test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### File Organization

- `internal/` for non-reusable code
- Clear package names matching directory structure
- Separate concerns into appropriate packages (`tui`, `config`, `prompt`, etc.)

### Comments

- Document exported functions and types
- Use clear, concise comments
- Explain complex logic or edge cases

### Dependencies

- Use Go modules for dependency management
- Keep dependencies minimal and well-maintained
- Use specific versions in go.mod</content>
  <parameter name="file_path">CRUSH.md

### Rules

This project also uses claude code. Look for the rules and commands defined in the following:

- @include CLAUDE.md
- @include .claude/\*.md

