# Menu Binding Mode - Product Requirements Document

## Problem Statement

The original menu implementation in the TUI had a global keybinding (default "x") that would trigger menu actions anywhere in the application. This created a user experience problem where typing the menu activation key in the chat area would accidentally trigger menu actions instead of being typed as part of the user's prompt.

## Solution Overview

Implement a "menu binding mode" system that requires users to first select the footer menu before menu keybindings become active. This prevents accidental menu activation while preserving the intended functionality.

## Implementation Details

### Core Architecture

1. **Separated Concerns**: Visual selection of the footer and functional menu binding mode are separate but coordinated concerns
2. **Focus Integration**: The footer menu becomes part of the normal focus cycle (tab navigation)
3. **Mode-Based Activation**: Menu keybindings only work when in "menu binding mode"

### User Interaction Flow

1. **Normal Operation**: Users can type freely in chat area without accidentally triggering menu
2. **Menu Activation**: User must first navigate to footer (via Tab or mouse click)
3. **Menu Actions**: Once footer is focused, menu keybindings (like "x") become active
4. **Mode Exit**: Escape key or navigating away from footer exits menu binding mode

### Technical Implementation

#### State Management
- Added `FooterMenuPanel` to `FocusedPanel` enum
- Added `menuBindingMode bool` field to track activation state
- Menu binding mode automatically activates when footer has focus

#### Navigation Updates
- Updated `nextPanel()` and `prevPanel()` to include footer in tab cycle
- Updated `handleMouseClick()` to detect clicks in footer area
- Footer becomes focusable like other panels

#### Visual Feedback
- Footer border changes color when focused (matches other focused panels)
- Uses same focus styling as other panels (color "69" when focused, "240" when not)

#### Keybinding Changes
- Menu activation key (default "x") only works when `menuBindingMode = true`
- Added Escape key handler to exit menu binding mode and return to chat panel
- Preserved all existing keybindings for other functions

## Benefits

1. **Prevents Accidental Activation**: Users can type freely without triggering menu accidentally
2. **Intuitive Navigation**: Footer becomes part of normal tab navigation flow
3. **Visual Clarity**: Clear indication when menu is "active" through focus styling
4. **Keyboard Accessible**: Full keyboard support with Escape key for quick exit
5. **Mouse Support**: Click footer to activate menu mode
6. **Extensible**: Easy to add more menu items in the future

## User Experience

### Before
- User types in chat: "I need to fix the regex pattern"
- Accidentally triggers menu when typing "x" in "fix"
- Menu activation interrupts typing flow

### After
- User types in chat: "I need to fix the regex pattern" 
- No accidental menu activation
- To access menu: Tab to footer (or click) → press "x" → menu activates
- Escape key quickly returns to chat for continued typing

## Future Considerations

This implementation provides a foundation for expanding menu functionality:
- Multiple menu items can be added to footer
- Each can have its own keybinding that only works in menu binding mode
- Visual styling can be enhanced to show available menu options
- Menu mode could display help text or additional context

## Testing

The implementation maintains compatibility with:
- All existing keyboard shortcuts (Ctrl+C, Ctrl+S, Ctrl+Y, Tab, Shift+Tab)
- Mouse interaction with all panels
- Existing focus management and visual styling
- Configuration system for menu activation key

## Files Modified

- `internal/tui/app.go`: Core implementation of menu binding mode system
  - Added FooterMenuPanel enum value
  - Added menuBindingMode state field
  - Updated navigation functions
  - Updated mouse handling
  - Added conditional keybinding logic
  - Updated visual styling for footer focus