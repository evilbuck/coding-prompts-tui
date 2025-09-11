# Debugging Guide

This guide provides instructions for debugging the chat input connection issue where user input from the chat area isn't properly flowing to the generated prompt.

## Current Issue

The application is showing "Test prompt" in the `<UserPrompt/>` section instead of the actual text typed by the user in the chat area.

## Debug Version

The application has been instrumented with comprehensive debug logging to trace the flow of user input from the chat textarea through to the prompt builder.

## Testing Instructions

### 1. Run the Application
```bash
./prompter .
```

### 2. Test the Complete Workflow

Follow this sequence to test the user input flow:

1. **Navigate to Chat Panel**
   - Use Tab key to switch focus to the chat panel
   - **Expected Debug Output:** `DEBUG (Focus): Switched to panel: 2` (ChatPanel = 2)

2. **Type Text in Chat Area**
   - Type some test text in the chat textarea
   - **Expected Debug Output:**
     ```
     DEBUG (App): Sending key to ChatPanel: [character]
     DEBUG (Chat): Received key runes: [character]
     DEBUG (Chat): Current value before update: [previous text]
     DEBUG (Chat): Current value after update: [updated text]
     ```

3. **Generate Prompt with Ctrl+S**
   - Press Ctrl+S to generate and display the prompt
   - **Expected Debug Output:**
     ```
     DEBUG (Ctrl+S): Chat textarea value: [your typed text]
     DEBUG (Ctrl+S): Chat textarea value length: [length]
     DEBUG (Builder): Received userPrompt: [your typed text]
     DEBUG (Builder): userPrompt length: [length]
     ```

4. **Copy Prompt with Ctrl+Shift+C**
   - Press Ctrl+Shift+C to copy the prompt to clipboard
   - **Expected Debug Output:**
     ```
     DEBUG: Chat textarea value: [your typed text]
     DEBUG: Chat textarea value length: [length]
     DEBUG (Builder): Received userPrompt: [your typed text]
     DEBUG (Builder): userPrompt length: [length]
     ```

## Troubleshooting by Debug Output

### Focus Issues
- **Missing:** `DEBUG (Focus): Switched to panel: 2`
- **Problem:** Tab navigation not switching to chat panel
- **Solution:** Check focus management in app.go

### Input Routing Issues
- **Missing:** `DEBUG (App): Sending key to ChatPanel: [character]`
- **Problem:** Keystrokes not being routed to chat panel when focused
- **Solution:** Check panel update logic in app.go

### Textarea Update Issues
- **Missing:** `DEBUG (Chat): Received key runes:` or value not updating
- **Problem:** Chat textarea not receiving or processing input
- **Solution:** Check chat.go Update method and textarea integration

### Value Retrieval Issues
- **Wrong Value:** Debug shows different text than what was typed
- **Problem:** State management issue between typing and retrieval
- **Solution:** Check textarea.Value() method and state consistency

### Prompt Builder Issues
- **Wrong Value in Builder:** Builder receives different text than app sent
- **Problem:** Parameter passing issue between app and builder
- **Solution:** Check function call parameters and argument passing

## Expected Behavior

The debug output should show a clear flow:
1. User switches to chat panel (focus change logged)
2. User types characters (each keystroke logged)
3. Textarea value updates with each character (before/after values logged)
4. When generating prompt, textarea.Value() returns the typed text
5. Prompt builder receives the correct user text
6. Generated XML contains the user's actual input in `<UserPrompt/>`

## Debug Code Locations

### App-level Debug (internal/tui/app.go)
- Focus switching debug (lines ~152, ~156)
- Panel update routing debug (lines ~188-190)
- Clipboard copy debug (lines ~112-114)
- Ctrl+S prompt generation debug (lines ~159-161)

### Chat Panel Debug (internal/tui/chat.go)
- Key input reception debug (lines ~39-43)
- Textarea value changes debug (lines ~48-52)

### Prompt Builder Debug (internal/prompt/builder.go)
- User prompt parameter debug (lines ~32-34)

## Cleaning Up Debug Code

Once the issue is identified and fixed, remove all debug `println()` statements from:
- `internal/tui/app.go`
- `internal/tui/chat.go` 
- `internal/prompt/builder.go`

Then rebuild the application:
```bash
go build -o prompter ./
```