# File Selection State Management Architecture

## Overview

The file selection system in the TUI application manages state across two panels: the File Tree panel and the Selected Files panel. This document describes how the selection state is coordinated between these components.

## Components

### 1. FileTreeModel (`internal/tui/filetree.go`)

**Responsibilities:**
- Display hierarchical file structure
- Handle directory expansion/collapse
- Manage file selection state
- Communicate selection changes to other panels

**Key State:**
- `selected map[string]bool` - Tracks which files are selected (path → selected status)
- `expanded map[string]bool` - Tracks which directories are expanded (path → expanded status)
- `items []filesystem.FileTreeItem` - Flattened tree view for display
- `rootNode *filesystem.FileNode` - Full filesystem tree structure

**Key Methods:**
- `refreshItems()` - Rebuilds the flattened item list from the tree structure
- `sendFileSelectionUpdate()` - Sends selection change messages to other panels
- `GetSelectedFiles()` - Returns current selection map

### 2. SelectedFilesModel (`internal/tui/selected.go`)

**Responsibilities:**
- Display list of selected files
- Allow removal of files from selection
- Communicate deselection back to file tree

**Key State:**
- `files []SelectedFile` - List of currently selected files
- `cursor int` - Current cursor position in the list

**Key Methods:**
- `AddFile(name, path)` - Adds a file to the selected list
- `RemoveFile(path)` - Removes a file by path
- `sendFileDeselectionUpdate(path)` - Sends deselection message when file is removed

### 3. App (`internal/tui/app.go`)

**Responsibilities:**
- Coordinate between panels
- Handle inter-panel communication messages
- Maintain overall application state

**Key Methods:**
- `updateSelectedFilesFromSelection(selectedFiles)` - Syncs SelectedFilesModel with FileTreeModel selection
- Message handling for `FileSelectionMsg` and `FileDeselectionMsg`

## State Flow

### File Selection (Spacebar in File Tree)

1. User presses spacebar on a file in the File Tree panel
2. `FileTreeModel.Update()` toggles the file's selection state in `m.selected[path]`
3. `FileTreeModel` calls `refreshItems()` to update display
4. `FileTreeModel` sends `FileSelectionMsg` containing entire selection map
5. `App.Update()` receives `FileSelectionMsg` and calls `updateSelectedFilesFromSelection()`
6. `App` clears and rebuilds the SelectedFilesModel's file list based on selection map
7. Both panels now reflect the updated selection state

### File Deselection (Delete/Backspace/X in Selected Files)

1. User presses delete key on a file in the Selected Files panel
2. `SelectedFilesModel.Update()` removes the file from its local list
3. `SelectedFilesModel` sends `FileDeselectionMsg` with the removed file's path
4. `App.Update()` receives `FileDeselectionMsg` and updates `FileTreeModel.selected[path] = false`
5. `App` calls `FileTreeModel.refreshItems()` to update display
6. Both panels now reflect the updated selection state

### Directory Expansion (Enter in File Tree)

1. User presses Enter on a directory in the File Tree panel
2. `FileTreeModel.Update()` toggles the directory's expanded state in `m.expanded[path]`
3. `FileTreeModel` calls `refreshItems()` to rebuild the flattened tree view
4. Directory contents become visible/hidden in the tree
5. File Tree display updates to show new structure

## Message Types

### # File Selection State Management Architecture

## Overview

The file selection system in the TUI application manages state across two panels: the File Tree panel and the Selected Files panel. This document describes how the selection state is coordinated between these components.

## Components

### 1. FileTreeModel (`internal/tui/filetree.go`)

**Responsibilities:**
- Display hierarchical file structure
- Handle directory expansion/collapse
- Manage file selection state
- Communicate selection changes to other panels

**Key State:**
- `selected map[string]bool` - Tracks which files are selected (path → selected status)
- `expanded map[string]bool` - Tracks which directories are expanded (path → expanded status)
- `items []filesystem.FileTreeItem` - Flattened tree view for display
- `rootNode *filesystem.FileNode` - Full filesystem tree structure

**Key Methods:**
- `refreshItems()` - Rebuilds the flattened item list from the tree structure
- `sendFileSelectionUpdate()` - Sends selection change messages to other panels
- `GetSelectedFiles()` - Returns current selection map

### 2. SelectedFilesModel (`internal/tui/selected.go`)

**Responsibilities:**
- Display list of selected files
- Allow removal of files from selection
- Communicate deselection back to file tree

**Key State:**
- `files []SelectedFile` - List of currently selected files
- `cursor int` - Current cursor position in the list

**Key Methods:**
- `AddFile(name, path)` - Adds a file to the selected list
- `RemoveFile(path)` - Removes a file by path
- `sendFileDeselectionUpdate(path)` - Sends deselection message when file is removed

### 3. App (`internal/tui/app.go`)

**Responsibilities:**
- Coordinate between panels
- Handle inter-panel communication messages
- Maintain overall application state
- **Orchestrate state persistence**

**Key Methods:**
- `updateSelectedFilesFromSelection(selectedFiles)` - Syncs SelectedFilesModel with FileTreeModel selection
- Message handling for `FileSelectionMsg`, `FileDeselectionMsg`, and `ChatInputMsg`

## Application State Persistence

A new `internal/config` package manages saving and loading the application's state to a JSON file, ensuring that user selections and prompts are preserved across sessions.

### Components

- **`ConfigManager` (`internal/config/manager.go`)**: A singleton responsible for all configuration-related operations. It loads the config on startup, provides the current workspace state to the TUI, and saves the config whenever the state changes.
- **`AppConfig` (`internal/config/config.go`)**: The root configuration struct, holding a map of all known workspaces.
- **`WorkspaceState` (`internal/config/config.go`)**: Stores the state for a single directory, including the list of `SelectedFiles` and the `ChatInput` text.

### Storage

- The configuration is stored in a JSON file located at `~/.config/prompter/config.json`. This provides a simple, human-readable, and dependency-free persistence mechanism.

## State Flow

### File Selection (Spacebar in File Tree)

1. User presses spacebar on a file in the File Tree panel
2. `FileTreeModel.Update()` toggles the file's selection state in `m.selected[path]`
3. `FileTreeModel` calls `refreshItems()` to update display
4. `FileTreeModel` sends `FileSelectionMsg` containing entire selection map
5. `App.Update()` receives `FileSelectionMsg`:
    a. It calls `updateSelectedFilesFromSelection()` to sync the UI.
    b. It updates the `workspace.SelectedFiles` slice.
    c. It calls `configManager.Save()` to persist the changes to disk.
6. Both panels now reflect the updated selection state.

### File Deselection (Delete/Backspace/X in Selected Files)

1. User presses delete key on a file in the Selected Files panel
2. `SelectedFilesModel.Update()` removes the file from its local list
3. `SelectedFilesModel` sends `FileDeselectionMsg` with the removed file's path
4. `App.Update()` receives `FileDeselectionMsg`:
    a. It updates `FileTreeModel.selected[path] = false`.
    b. It calls `FileTreeModel.refreshItems()` to update the UI.
    c. It removes the file from the `workspace.SelectedFiles` slice.
    d. It calls `configManager.Save()` to persist the changes.
5. Both panels now reflect the updated selection state.

### Chat Input Change

1. The user types in the `ChatModel`'s textarea.
2. The `ChatModel.Update()` method updates the textarea's internal value.
3. In `App.Update()`, after the `ChatModel` is updated, a check is performed to see if the textarea's value differs from the `workspace.ChatInput`.
4. If it has changed, the `App` model dispatches a `ChatInputMsg` to itself.
5. `App.Update()` receives the `ChatInputMsg`:
    a. It updates `workspace.ChatInput` with the new content.
    b. It calls `configManager.Save()` to persist the change.

### Application Startup

1. `main()` initializes the `ConfigManager`.
2. `ConfigManager` loads `~/.config/prompter/config.json` into the `AppConfig` struct. If the file doesn't exist, a default one is created.
3. The absolute path of the target directory is determined.
4. `configManager.GetWorkspace(path)` is called to get the `WorkspaceState` for the current directory, creating a new one if it's the first time.
5. The `App` model is initialized with the `ConfigManager` and the `WorkspaceState`.
6. `NewApp` passes the `workspace.SelectedFiles` to `NewFileTreeModel` and `workspace.ChatInput` to `NewChatModel`, restoring the previous session's state.

## Message Types

### FileSelectionMsg
```go
type FileSelectionMsg struct {
    SelectedFiles map[string]bool
}
```
Sent from FileTreeModel to App when file selection changes. Contains complete selection state.

### FileDeselectionMsg
```go
type FileDeselectionMsg struct {
    FilePath string
}
```
Sent from SelectedFilesModel to App when a file is removed from the selected list.

### ChatInputMsg
```go
type ChatInputMsg struct {
    Content string
}
```
Sent from the App model to itself when the chat textarea's content has changed.

## Data Flow Diagram

```
FileTreeModel                    App                    SelectedFilesModel
     |                          |                            |
     | [spacebar pressed]       |                            |
     |---------------------->   |                            |
     | FileSelectionMsg         |                            |
     |                          |---------->                 |
     |                          | updateSelectedFilesFrom... |
     |                          |                            |
     |                          |       [delete pressed]    |
     |                          |   <------------------------|
     |                          |   FileDeselectionMsg       |
     |   <----------------------|                            |
     |   update selected[path]  |                            |
     |   refreshItems()         |                            |
```

**Persistence Flow Diagram**
```
   TUI Models                  App Model                  ConfigManager
(FileTree, Chat)                 |                            |
       |                         |                            |
       |--- (User action) ---->  |                            |
       |                         |                            |
       |                         |--- (State change) -------> |
       |                         | workspace.ChatInput = ...  |
       |                         | workspace.SelectedFiles = ...|
       |                         |                            |
       |                         | configManager.Save() ------>| (Writes to config.json)
       |                         |                            |
```

## Key Design Decisions

### Single Source of Truth
The `FileTreeModel.selected` map serves as the single source of truth for file selection state *within the UI*. The `WorkspaceState` struct is the source of truth for *persistent storage*.

### Message-Based Communication
Panels communicate through Bubble Tea messages rather than direct method calls, maintaining clean separation of concerns and following the Bubble Tea architecture pattern.

### Path-Based Identification
Files are identified by their full filesystem paths, ensuring unique identification even when files have the same name in different directories.

### Lazy Tree Building
The file tree is built lazily - directories are only scanned when expanded, improving performance for large directory structures.

### State Synchronization
Selection state is kept in sync between panels through a message-passing system that ensures both views reflect the same underlying state.

### Simple, Embedded Persistence
State is saved to a simple JSON file in the user's config directory. This avoids external dependencies like databases and keeps the application self-contained and easy to manage.

## Performance Considerations

- **Lazy Loading**: Directories are only scanned when expanded
- **Efficient Updates**: Only changed parts of the tree are rebuilt
- **Message Batching**: Bubble Tea naturally batches updates for smooth rendering
- **Path-Based Lookups**: Using maps for O(1) selection and expansion state lookups


## Data Flow Diagram

```
FileTreeModel                    App                    SelectedFilesModel
     |                          |                            |
     | [spacebar pressed]       |                            |
     |---------------------->   |                            |
     | FileSelectionMsg         |                            |
     |                          |---------->                 |
     |                          | updateSelectedFilesFrom... |
     |                          |                            |
     |                          |       [delete pressed]    |
     |                          |   <------------------------|
     |                          |   FileDeselectionMsg       |
     |   <----------------------|                            |
     |   update selected[path]  |                            |
     |   refreshItems()         |                            |
```

## Key Design Decisions

### Single Source of Truth
The `FileTreeModel.selected` map serves as the single source of truth for file selection state. The `SelectedFilesModel` is a derived view that gets updated via messages.

### Message-Based Communication
Panels communicate through Bubble Tea messages rather than direct method calls, maintaining clean separation of concerns and following the Bubble Tea architecture pattern.

### Path-Based Identification
Files are identified by their full filesystem paths, ensuring unique identification even when files have the same name in different directories.

### Lazy Tree Building
The file tree is built lazily - directories are only scanned when expanded, improving performance for large directory structures.

### State Synchronization
Selection state is kept in sync between panels through a message-passing system that ensures both views reflect the same underlying state.

## Performance Considerations

- **Lazy Loading**: Directories are only scanned when expanded
- **Efficient Updates**: Only changed parts of the tree are rebuilt
- **Message Batching**: Bubble Tea naturally batches updates for smooth rendering
- **Path-Based Lookups**: Using maps for O(1) selection and expansion state lookups