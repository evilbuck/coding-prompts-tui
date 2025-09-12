# Debugging Guide

This guide provides instructions for using the configurable debug system to troubleshoot issues in the coding-prompts TUI application.

## Debug System Overview

The application includes a configurable debug system that can be enabled through configuration or toggled during runtime. When enabled, it provides comprehensive logging of key events, menu activation, and internal state changes.

## Debug Configuration

Debug settings are configured in `~/.config/coding-prompts/coding_prompts.toml`:

```toml
[debug]
# Enable debug mode on startup (default: false)
enabled = false
# Key binding to toggle debug mode (default: "f11")
toggle_key = "f11"
# Enable file logging for debug messages (default: true when debug enabled)
file_logging = true
# Log file path relative to workspace (default: "logs/error.log")
log_file = "logs/error.log"
```

## Using Debug Mode

### 1. Enable Debug Mode

You can enable debug mode in several ways:

**Persistent (via configuration):**
```toml
[debug]
enabled = true
```

**Runtime Toggle:**
- Press the configured toggle key (default: F11) to toggle debug mode
- Look for "Debug mode ON/OFF" notification
- Footer will show the current toggle key

### 2. Debug Information

When debug mode is active, you'll see:

1. **Key Event Logging**
   - All key presses are captured and displayed
   - Shows key type, modifiers, and runes
   - Helps debug key binding issues

2. **Menu Activation Debugging**
   - Shows menu activation attempts and results
   - Displays expected vs. actual key combinations
   - Useful for troubleshooting modal key bindings

3. **Visual Feedback**
   - Debug information appears as notifications
   - Footer shows debug toggle status
   - Current key information is displayed in real-time

### 3. File Logging

When `file_logging` is enabled:
- Debug messages are written to the configured log file
- Default location: `logs/error.log` in workspace
- Log includes timestamps and session markers
- Useful for persistent debugging across sessions

## Troubleshooting Common Issues

### Key Binding Issues

When debug mode is active, key information is displayed for every key press:
```
Key: "alt+m", Type: KeyRune, Alt: true, Runes: [109]
```

**Common Problems:**
- **Menu not activating:** Check if the debug output shows the correct key combination
- **Wrong key detected:** Terminal may interpret key combinations differently
- **Modifier issues:** Alt, Ctrl, Shift detection varies by terminal

### Menu Mode Problems

Debug output shows menu activation attempts:
```
Menu check: Legacy=false, Expected="alt+m", Got="alt+m"
```

**Troubleshooting:**
- Compare "Expected" vs "Got" values
- Legacy mode may affect key parsing
- Terminal compatibility issues with complex key combinations

### Configuration Issues

**Debug not starting:** Check configuration syntax and file location
**Toggle key not working:** Verify key binding syntax in config
**File logging not working:** Check file permissions and directory existence

### Log File Issues

**Log file location:** Relative to workspace directory
**Permissions:** Ensure write access to log directory
**Disk space:** Check available space for log files

## Configuration Examples

### Enable Debug on Startup
```toml
[debug]
enabled = true
toggle_key = "f11"
file_logging = true
log_file = "debug.log"
```

### Custom Debug Key
```toml
[debug]
enabled = false
toggle_key = "ctrl+d"
file_logging = true
log_file = "logs/debug.log"
```

### Debug Without File Logging
```toml
[debug]
enabled = false
toggle_key = "f12"
file_logging = false
```

## Debug Architecture

The debug system is implemented across several components:

### Configuration (`internal/config/settings.go`)
- `DebugSettings` struct for TOML configuration
- Settings validation and defaults
- Thread-safe access methods

### TUI Application (`internal/tui/app.go`)
- Runtime debug mode toggle
- Key event logging and display
- Debug logger initialization
- Integration with settings system

### Key Components
- Debug mode can be enabled via config or runtime toggle
- File logging is optional and configurable
- Toggle key is customizable through configuration
- Debug information appears as notifications in the UI

## Best Practices

1. **Use file logging** for complex debugging sessions
2. **Configure a comfortable toggle key** for your workflow
3. **Disable debug in production** configurations
4. **Monitor log file size** for long debugging sessions
5. **Use notifications for real-time feedback** during development