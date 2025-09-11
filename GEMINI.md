# Project: LLM Coding Prompt Builder

## Project Overview

This project is a terminal user interface (TUI) tool built with Go. Its purpose is to help developers create structured, context-rich prompts for large language models (LLMs) like Gemini. The tool allows users to navigate their project's file tree, select relevant files, and write a user prompt. It then combines this information into a well-formatted XML structure, which is ideal for providing context to an LLM.

The application is built using the following key technologies:
*   **Go**: The primary programming language.
*   **Bubble Tea**: A Go library for building TUI applications.
*   **Bubbles**: A library of TUI components for Bubble Tea.
*   **Lipgloss**: A library for styling TUI elements.

The architecture is straightforward:
*   `main.go` is the application entry point.
*   The `internal/tui` package manages the user interface, with separate files for the main application model (`app.go`), the file tree (`filetree.go`), the list of selected files (`selected.go`), and the chat input (`chat.go`).
*   The `internal/filesystem` package handles file system operations, such as scanning directories and reading files.

## Building and Running

### Prerequisites

*   Go 1.21 or later

### Building the Application

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd coding-prompts-tui
    ```

2.  **Build the executable:**
    ```bash
    go build -o prompter src/main.go
    ```

### Running the Application

To run the application, provide a directory path as an argument:

```bash
./prompter <directory>
```

For example, to run it in the current directory:

```bash
./prompter .
```

### Running Tests

To run the project's tests, use the following command:

```bash
go test ./...
```

## Development Conventions

*   **Styling**: The project uses `lipgloss` for styling, with styles defined in the TUI components.
*   **Structure**: The code is organized into `internal` packages, which is a standard Go practice for non-reusable code.
*   **Dependencies**: Go modules are used for dependency management. Dependencies are listed in the `go.mod` file.
*   **File Filtering**: The application is designed to ignore common non-source files, such as `node_modules`, `.git`, and build artifacts. This logic is located in `internal/filesystem/scanner.go`.
