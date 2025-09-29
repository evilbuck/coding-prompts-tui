package config

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestParseKeyBinding(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *KeyCombination
		expectError bool
	}{
		{
			name:  "simple key",
			input: "m",
			expected: &KeyCombination{
				Key:   "m",
				Ctrl:  false,
				Alt:   false,
				Shift: false,
			},
			expectError: false,
		},
		{
			name:  "ctrl+shift+m",
			input: "ctrl+shift+m",
			expected: &KeyCombination{
				Key:   "m",
				Ctrl:  true,
				Alt:   false,
				Shift: true,
			},
			expectError: false,
		},
		{
			name:  "alt+f12",
			input: "alt+f12",
			expected: &KeyCombination{
				Key:   "f12",
				Ctrl:  false,
				Alt:   true,
				Shift: false,
			},
			expectError: false,
		},
		{
			name:  "escape key",
			input: "esc",
			expected: &KeyCombination{
				Key:   "esc",
				Ctrl:  false,
				Alt:   false,
				Shift: false,
			},
			expectError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "invalid modifier",
			input:       "super+m",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "missing key",
			input:       "ctrl+",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseKeyBinding(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.Key != tt.expected.Key {
				t.Errorf("Expected key %q, got %q", tt.expected.Key, result.Key)
			}
			if result.Ctrl != tt.expected.Ctrl {
				t.Errorf("Expected Ctrl %v, got %v", tt.expected.Ctrl, result.Ctrl)
			}
			if result.Alt != tt.expected.Alt {
				t.Errorf("Expected Alt %v, got %v", tt.expected.Alt, result.Alt)
			}
			if result.Shift != tt.expected.Shift {
				t.Errorf("Expected Shift %v, got %v", tt.expected.Shift, result.Shift)
			}
		})
	}
}

func TestKeyCombinationString(t *testing.T) {
	tests := []struct {
		name     string
		combo    *KeyCombination
		expected string
	}{
		{
			name: "simple key",
			combo: &KeyCombination{
				Key:   "m",
				Ctrl:  false,
				Alt:   false,
				Shift: false,
			},
			expected: "m",
		},
		{
			name: "ctrl+shift+m",
			combo: &KeyCombination{
				Key:   "m",
				Ctrl:  true,
				Alt:   false,
				Shift: true,
			},
			expected: "ctrl+shift+m",
		},
		{
			name: "alt+f12",
			combo: &KeyCombination{
				Key:   "f12",
				Ctrl:  false,
				Alt:   true,
				Shift: false,
			},
			expected: "alt+f12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.combo.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMatchesKeyMsg(t *testing.T) {
	tests := []struct {
		name     string
		combo    *KeyCombination
		keyMsg   tea.KeyMsg
		expected bool
	}{
		{
			name: "simple key match",
			combo: &KeyCombination{
				Key:   "m",
				Ctrl:  false,
				Alt:   false,
				Shift: false,
			},
			keyMsg: tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'m'},
			},
			expected: true,
		},
		{
			name: "escape key match",
			combo: &KeyCombination{
				Key:   "esc",
				Ctrl:  false,
				Alt:   false,
				Shift: false,
			},
			keyMsg: tea.KeyMsg{
				Type: tea.KeyEsc,
			},
			expected: true,
		},
		{
			name: "tab key match",
			combo: &KeyCombination{
				Key:   "tab",
				Ctrl:  false,
				Alt:   false,
				Shift: false,
			},
			keyMsg: tea.KeyMsg{
				Type: tea.KeyTab,
				Alt:  false,
			},
			expected: true,
		},
		{
			name: "simple key mismatch",
			combo: &KeyCombination{
				Key:   "m",
				Ctrl:  false,
				Alt:   false,
				Shift: false,
			},
			keyMsg: tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{'n'},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.combo.MatchesKeyMsg(tt.keyMsg)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
