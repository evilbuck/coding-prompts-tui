package config

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// KeyCombination represents a parsed key combination
type KeyCombination struct {
	Key   string
	Ctrl  bool
	Alt   bool
	Shift bool
}

// ParseKeyBinding parses a key binding string into a KeyCombination
func ParseKeyBinding(binding string) (*KeyCombination, error) {
	if binding == "" {
		return nil, fmt.Errorf("key binding cannot be empty")
	}

	// Split on '+' to get modifiers and key
	parts := strings.Split(strings.ToLower(binding), "+")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid key binding format: %q", binding)
	}

	combo := &KeyCombination{}

	// The last part is always the key
	keyPart := parts[len(parts)-1]

	// Process modifiers (all parts except the last one)
	for i := 0; i < len(parts)-1; i++ {
		modifier := strings.TrimSpace(parts[i])
		switch modifier {
		case "ctrl", "control":
			combo.Ctrl = true
		case "alt":
			combo.Alt = true
		case "shift":
			combo.Shift = true
		default:
			return nil, fmt.Errorf("unknown modifier: %q", modifier)
		}
	}

	// Set the key
	combo.Key = strings.TrimSpace(keyPart)
	if combo.Key == "" {
		return nil, fmt.Errorf("key cannot be empty")
	}

	return combo, nil
}

// MatchesKeyMsg checks if a tea.KeyMsg matches this key combination
func (kc *KeyCombination) MatchesKeyMsg(msg tea.KeyMsg) bool {
	// First, try exact string matching for complex combinations
	expectedStr := kc.String()
	actualStr := msg.String()

	// Handle common terminal variations of complex key combinations
	if expectedStr == "ctrl+shift+m" {
		// Try different ways terminals might encode ctrl+shift+m
		variations := []string{
			"ctrl+shift+m",
			"ctrl+M",     // Some terminals send uppercase
			"ctrl+alt+m", // Some terminals map ctrl+shift to ctrl+alt
		}

		for _, variation := range variations {
			if actualStr == variation {
				return true
			}
		}

		// Check if it's sent as alt+m with special encoding
		if msg.Alt && len(msg.Runes) > 0 && strings.ToLower(string(msg.Runes[0])) == "m" {
			return true
		}
	}

	// Try exact string match first
	if actualStr == expectedStr {
		return true
	}

	// Fallback to component-wise matching
	key := strings.ToLower(kc.Key)

	// Handle special keys that don't use runes
	isSpecialKey := key == "tab" || key == "esc" || key == "escape" ||
		key == "enter" || key == "space" || key == "backspace" ||
		key == "delete" || key == "up" || key == "down" ||
		key == "left" || key == "right" || key == "home" ||
		key == "end" || strings.HasPrefix(key, "f")

	if isSpecialKey {
		// For special keys, check modifiers carefully
		if kc.Alt != msg.Alt {
			return false
		}

		// Check ctrl modifier for special keys
		if kc.Ctrl && !hasCtrlModifier(msg) {
			return false
		}

		if !kc.Ctrl && hasCtrlModifier(msg) {
			return false
		}
	} else {
		// For regular character keys, check all modifiers
		if kc.Ctrl && !hasCtrlModifier(msg) {
			return false
		}

		if !kc.Ctrl && hasCtrlModifier(msg) {
			return false
		}

		if kc.Alt != msg.Alt {
			return false
		}
	}

	// Check the key itself
	return kc.matchesKey(msg)
}

// matchesKey checks if the key part matches
func (kc *KeyCombination) matchesKey(msg tea.KeyMsg) bool {
	key := strings.ToLower(kc.Key)

	// Handle special keys
	switch key {
	case "esc", "escape":
		return msg.Type == tea.KeyEsc
	case "tab":
		return msg.Type == tea.KeyTab
	case "enter", "return":
		return msg.Type == tea.KeyEnter
	case "space":
		return msg.Type == tea.KeySpace
	case "backspace":
		return msg.Type == tea.KeyBackspace
	case "delete", "del":
		return msg.Type == tea.KeyDelete
	case "up":
		return msg.Type == tea.KeyUp
	case "down":
		return msg.Type == tea.KeyDown
	case "left":
		return msg.Type == tea.KeyLeft
	case "right":
		return msg.Type == tea.KeyRight
	case "home":
		return msg.Type == tea.KeyHome
	case "end":
		return msg.Type == tea.KeyEnd
	case "pgup", "pageup":
		return msg.Type == tea.KeyPgUp
	case "pgdn", "pagedown":
		return msg.Type == tea.KeyPgDown
	}

	// Handle function keys
	if strings.HasPrefix(key, "f") && len(key) > 1 {
		switch key {
		case "f1":
			return msg.Type == tea.KeyF1
		case "f2":
			return msg.Type == tea.KeyF2
		case "f3":
			return msg.Type == tea.KeyF3
		case "f4":
			return msg.Type == tea.KeyF4
		case "f5":
			return msg.Type == tea.KeyF5
		case "f6":
			return msg.Type == tea.KeyF6
		case "f7":
			return msg.Type == tea.KeyF7
		case "f8":
			return msg.Type == tea.KeyF8
		case "f9":
			return msg.Type == tea.KeyF9
		case "f10":
			return msg.Type == tea.KeyF10
		case "f11":
			return msg.Type == tea.KeyF11
		case "f12":
			return msg.Type == tea.KeyF12
		}
	}

	// Handle regular characters
	if len(key) == 1 {
		// For single characters, check against the runes
		if len(msg.Runes) > 0 {
			return strings.ToLower(string(msg.Runes[0])) == key
		}

		// Also check against the String representation
		return strings.ToLower(msg.String()) == key
	}

	return false
}

// hasCtrlModifier checks if the key message has a ctrl modifier
func hasCtrlModifier(msg tea.KeyMsg) bool {
	// This is a heuristic based on common ctrl key combinations
	switch msg.Type {
	case tea.KeyCtrlA, tea.KeyCtrlB, tea.KeyCtrlC, tea.KeyCtrlD, tea.KeyCtrlE,
		tea.KeyCtrlF, tea.KeyCtrlG, tea.KeyCtrlH, tea.KeyCtrlI, tea.KeyCtrlJ,
		tea.KeyCtrlK, tea.KeyCtrlL, tea.KeyCtrlM, tea.KeyCtrlN, tea.KeyCtrlO,
		tea.KeyCtrlP, tea.KeyCtrlQ, tea.KeyCtrlR, tea.KeyCtrlS, tea.KeyCtrlT,
		tea.KeyCtrlU, tea.KeyCtrlV, tea.KeyCtrlW, tea.KeyCtrlX, tea.KeyCtrlY,
		tea.KeyCtrlZ:
		return true
	}

	// Check string representation for more complex combinations
	msgStr := msg.String()
	if strings.Contains(msgStr, "ctrl+") {
		return true
	}

	// Check for ctrl+shift combinations that might come through differently
	if msg.Alt && len(msg.Runes) > 0 {
		// In some terminals, ctrl+shift+key combinations come through as alt+key
		return true
	}

	return false
}

// String returns a string representation of the key combination
func (kc *KeyCombination) String() string {
	parts := []string{}

	if kc.Ctrl {
		parts = append(parts, "ctrl")
	}
	if kc.Alt {
		parts = append(parts, "alt")
	}
	if kc.Shift {
		parts = append(parts, "shift")
	}

	parts = append(parts, kc.Key)

	return strings.Join(parts, "+")
}
