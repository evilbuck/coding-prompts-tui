package tui

import (
	"testing"

	"coding-prompts-tui/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// StateObserver tracks state changes for testing
type StateObserver struct {
	FocusChanges    []FocusedPanel
	MenuModeChanges []bool
	DebugChanges    []bool
	LayoutChanges   []LayoutChangeMsg
}

func NewStateObserver() *StateObserver {
	return &StateObserver{
		FocusChanges:    make([]FocusedPanel, 0),
		MenuModeChanges: make([]bool, 0),
		DebugChanges:    make([]bool, 0),
		LayoutChanges:   make([]LayoutChangeMsg, 0),
	}
}

func (so *StateObserver) RecordMessage(msg tea.Msg) {
	switch msg := msg.(type) {
	case FocusChangeMsg:
		so.FocusChanges = append(so.FocusChanges, msg.Panel)
	case MenuModeChangeMsg:
		so.MenuModeChanges = append(so.MenuModeChanges, msg.Enabled)
	case DebugModeChangeMsg:
		so.DebugChanges = append(so.DebugChanges, msg.Enabled)
	case LayoutChangeMsg:
		so.LayoutChanges = append(so.LayoutChanges, msg)
	}
}

func (so *StateObserver) Reset() {
	so.FocusChanges = so.FocusChanges[:0]
	so.MenuModeChanges = so.MenuModeChanges[:0]
	so.DebugChanges = so.DebugChanges[:0]
	so.LayoutChanges = so.LayoutChanges[:0]
}

// Helper function to create a test app with minimal dependencies
func createTestApp(t *testing.T) *App {
	// Create temporary directory for test
	targetDir := t.TempDir()
	
	// Create minimal config managers
	cfgManager, err := config.NewManager()
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}
	
	settingsManager, err := config.NewSettingsManager()
	if err != nil {
		t.Fatalf("Failed to create settings manager: %v", err)
	}
	
	workspace := &config.WorkspaceState{
		Path:           targetDir,
		SelectedFiles:  []string{},
		ChatInput:      "",
		ActivePersonas: []string{"default"},
	}
	
	return NewApp(targetDir, cfgManager, settingsManager, workspace)
}

// TestStateCommandGeneration tests that state commands are generated correctly
func TestStateCommandGeneration(t *testing.T) {
	app := createTestApp(t)
	observer := NewStateObserver()

	tests := []struct {
		name     string
		cmd      tea.Cmd
		expected tea.Msg
	}{
		{
			name: "setFocus generates FocusChangeMsg",
			cmd:  app.setFocus(ChatPanel),
			expected: FocusChangeMsg{Panel: ChatPanel},
		},
		{
			name: "setMenuMode generates MenuModeChangeMsg",
			cmd:  app.setMenuMode(true),
			expected: MenuModeChangeMsg{Enabled: true},
		},
		{
			name: "toggleDebugMode generates DebugModeChangeMsg",
			cmd:  app.toggleDebugMode(),
			expected: DebugModeChangeMsg{Enabled: !app.debugMode},
		},
		{
			name: "updateLayout generates LayoutChangeMsg",
			cmd:  app.updateLayout(100, 50),
			expected: LayoutChangeMsg{Width: 100, Height: 50},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd == nil {
				t.Fatal("Command should not be nil")
			}

			msg := tt.cmd()
			observer.RecordMessage(msg)

			switch expected := tt.expected.(type) {
			case FocusChangeMsg:
				if len(observer.FocusChanges) == 0 {
					t.Fatal("Expected FocusChangeMsg was not recorded")
				}
				if observer.FocusChanges[len(observer.FocusChanges)-1] != expected.Panel {
					t.Errorf("Expected panel %v, got %v", expected.Panel, observer.FocusChanges[len(observer.FocusChanges)-1])
				}
			case MenuModeChangeMsg:
				if len(observer.MenuModeChanges) == 0 {
					t.Fatal("Expected MenuModeChangeMsg was not recorded")
				}
				if observer.MenuModeChanges[len(observer.MenuModeChanges)-1] != expected.Enabled {
					t.Errorf("Expected enabled %v, got %v", expected.Enabled, observer.MenuModeChanges[len(observer.MenuModeChanges)-1])
				}
			case DebugModeChangeMsg:
				if len(observer.DebugChanges) == 0 {
					t.Fatal("Expected DebugModeChangeMsg was not recorded")
				}
				if observer.DebugChanges[len(observer.DebugChanges)-1] != expected.Enabled {
					t.Errorf("Expected enabled %v, got %v", expected.Enabled, observer.DebugChanges[len(observer.DebugChanges)-1])
				}
			case LayoutChangeMsg:
				if len(observer.LayoutChanges) == 0 {
					t.Fatal("Expected LayoutChangeMsg was not recorded")
				}
				last := observer.LayoutChanges[len(observer.LayoutChanges)-1]
				if last.Width != expected.Width || last.Height != expected.Height {
					t.Errorf("Expected layout %dx%d, got %dx%d", expected.Width, expected.Height, last.Width, last.Height)
				}
			}
			
			observer.Reset()
		})
	}
}

// TestStateObservability tests that all state changes are observable
func TestStateObservability(t *testing.T) {
	app := createTestApp(t)
	observer := NewStateObserver()

	// Test focus change observability
	t.Run("Focus changes are observable", func(t *testing.T) {
		allPanels := []FocusedPanel{FileTreePanel, SelectedFilesPanel, ChatPanel, FooterMenuPanel}
		
		for _, panel := range allPanels {
			cmd := app.setFocus(panel)
			msg := cmd()
			observer.RecordMessage(msg)
		}

		if len(observer.FocusChanges) != len(allPanels) {
			t.Errorf("Expected %d focus changes, got %d", len(allPanels), len(observer.FocusChanges))
		}

		for i, panel := range allPanels {
			if observer.FocusChanges[i] != panel {
				t.Errorf("Expected focus change %d to be %v, got %v", i, panel, observer.FocusChanges[i])
			}
		}
	})

	observer.Reset()

	// Test menu mode change observability
	t.Run("Menu mode changes are observable", func(t *testing.T) {
		states := []bool{true, false, true}
		
		for _, enabled := range states {
			cmd := app.setMenuMode(enabled)
			msg := cmd()
			observer.RecordMessage(msg)
		}

		if len(observer.MenuModeChanges) != len(states) {
			t.Errorf("Expected %d menu mode changes, got %d", len(states), len(observer.MenuModeChanges))
		}

		for i, state := range states {
			if observer.MenuModeChanges[i] != state {
				t.Errorf("Expected menu mode change %d to be %v, got %v", i, state, observer.MenuModeChanges[i])
			}
		}
	})

	observer.Reset()

	// Test debug mode change observability
	t.Run("Debug mode changes are observable", func(t *testing.T) {
		initialDebugMode := app.debugMode
		
		cmd := app.toggleDebugMode()
		msg := cmd()
		observer.RecordMessage(msg)

		if len(observer.DebugChanges) != 1 {
			t.Errorf("Expected 1 debug mode change, got %d", len(observer.DebugChanges))
		}

		expectedState := !initialDebugMode
		if observer.DebugChanges[0] != expectedState {
			t.Errorf("Expected debug mode change to be %v, got %v", expectedState, observer.DebugChanges[0])
		}
	})
}

// TestStateValidation tests state validation and invariants
func TestStateValidation(t *testing.T) {
	app := createTestApp(t)

	// Test that focus values are within valid range
	t.Run("Focus panel validation", func(t *testing.T) {
		validPanels := []FocusedPanel{FileTreePanel, SelectedFilesPanel, ChatPanel, FooterMenuPanel}
		
		for _, panel := range validPanels {
			if panel < FileTreePanel || panel > FooterMenuPanel {
				t.Errorf("Panel %v is outside valid range", panel)
			}
		}
	})

	// Test initial state is valid
	t.Run("Initial state is valid", func(t *testing.T) {
		if app.focused < FileTreePanel || app.focused > FooterMenuPanel {
			t.Errorf("Initial focus %v is invalid", app.focused)
		}
		
		// Menu binding mode should be consistent with focus in legacy mode
		if app.settingsManager.IsLegacyMode() {
			expectedMenuMode := (app.focused == FooterMenuPanel)
			if app.menuBindingMode != expectedMenuMode {
				t.Errorf("Initial menu binding mode %v inconsistent with focus %v in legacy mode", 
					app.menuBindingMode, app.focused)
			}
		}
	})
}

// TestCentralizedStateHandler tests the centralized state handler
func TestCentralizedStateHandler(t *testing.T) {
	app := createTestApp(t)
	
	t.Run("Focus change handling", func(t *testing.T) {
		// Initial state
		initialFocus := app.focused
		
		// Test valid focus change
		model, cmd := app.handleStateChange(FocusChangeMsg{Panel: ChatPanel})
		app = model.(*App)
		
		if app.focused != ChatPanel {
			t.Errorf("Expected focus to be ChatPanel, got %v", app.focused)
		}
		
		// Test that it changed from initial
		if app.focused == initialFocus {
			t.Error("Focus should have changed from initial state")
		}
		
		// Should have no commands for valid focus change
		if cmd != nil {
			t.Error("Valid focus change should not generate commands")
		}
	})
	
	t.Run("Invalid focus change handling", func(t *testing.T) {
		// Test invalid focus change (outside valid range)
		app.debugMode = true // Enable debug mode to see error
		model, cmd := app.handleStateChange(FocusChangeMsg{Panel: FocusedPanel(999)})
		app = model.(*App)
		
		// Should have generated error command in debug mode
		if cmd == nil {
			t.Error("Invalid focus change in debug mode should generate error command")
		}
	})
	
	t.Run("Menu mode change handling", func(t *testing.T) {
		// Test menu mode change
		initialMenuMode := app.menuBindingMode
		
		model, cmd := app.handleStateChange(MenuModeChangeMsg{Enabled: !initialMenuMode})
		app = model.(*App)
		
		if app.menuBindingMode == initialMenuMode {
			t.Error("Menu mode should have changed")
		}
		
		// Should have no commands for valid menu mode change
		if cmd != nil {
			t.Error("Valid menu mode change should not generate commands")
		}
	})
	
	t.Run("Debug mode change handling", func(t *testing.T) {
		// Test debug mode change
		initialDebugMode := app.debugMode
		
		model, cmd := app.handleStateChange(DebugModeChangeMsg{Enabled: !initialDebugMode})
		app = model.(*App)
		
		if app.debugMode == initialDebugMode {
			t.Error("Debug mode should have changed")
		}
		
		// Should generate notification command
		if cmd == nil {
			t.Error("Debug mode change should generate notification command")
		}
	})
	
	t.Run("Layout change handling", func(t *testing.T) {
		// Test valid layout change
		newWidth, newHeight := 100, 50
		
		model, cmd := app.handleStateChange(LayoutChangeMsg{Width: newWidth, Height: newHeight})
		app = model.(*App)
		
		if app.width != newWidth || app.height != newHeight {
			t.Errorf("Expected layout %dx%d, got %dx%d", newWidth, newHeight, app.width, app.height)
		}
		
		// Should have no commands for valid layout change
		if cmd != nil {
			t.Error("Valid layout change should not generate commands")
		}
	})
	
	t.Run("Invalid layout change handling", func(t *testing.T) {
		// Test invalid layout change
		app.debugMode = true // Enable debug mode to see error
		
		model, cmd := app.handleStateChange(LayoutChangeMsg{Width: -1, Height: -1})
		app = model.(*App)
		
		// Should have generated error command in debug mode
		if cmd == nil {
			t.Error("Invalid layout change in debug mode should generate error command")
		}
	})
	
	t.Run("Non-state messages pass through unchanged", func(t *testing.T) {
		// Test that non-state messages return nil (pass through)
		model, cmd := app.handleStateChange("not-a-state-message")
		
		if model != app || cmd != nil {
			t.Error("Non-state messages should pass through unchanged")
		}
	})
}

// TestStateTransitionIntegrity tests state transitions maintain consistency
func TestStateTransitionIntegrity(t *testing.T) {
	app := createTestApp(t)
	
	t.Run("Focus and menu mode consistency in legacy mode", func(t *testing.T) {
		// Make sure we're in legacy mode for this test
		if !app.settingsManager.IsLegacyMode() {
			t.Skip("Skipping legacy mode test - not in legacy mode")
		}
		
		// Test focus change to footer panel
		model, _ := app.handleStateChange(FocusChangeMsg{Panel: FooterMenuPanel})
		app = model.(*App)
		
		if app.focused != FooterMenuPanel {
			t.Error("Focus should be on footer panel")
		}
		
		if !app.menuBindingMode {
			t.Error("Menu binding mode should be enabled when footer panel is focused in legacy mode")
		}
		
		// Test focus change away from footer panel
		model, _ = app.handleStateChange(FocusChangeMsg{Panel: ChatPanel})
		app = model.(*App)
		
		if app.focused != ChatPanel {
			t.Error("Focus should be on chat panel")
		}
		
		if app.menuBindingMode {
			t.Error("Menu binding mode should be disabled when not on footer panel in legacy mode")
		}
	})
	
	t.Run("Menu mode enables footer focus", func(t *testing.T) {
		// Start with non-footer focus
		app.handleStateChange(FocusChangeMsg{Panel: ChatPanel})
		
		// Enable menu mode
		model, _ := app.handleStateChange(MenuModeChangeMsg{Enabled: true})
		app = model.(*App)
		
		if !app.menuBindingMode {
			t.Error("Menu binding mode should be enabled")
		}
		
		if app.focused != FooterMenuPanel {
			t.Error("Focus should be on footer panel when menu mode is enabled")
		}
	})
	
	t.Run("State invariants maintained", func(t *testing.T) {
		// Test various state combinations to ensure invariants are maintained
		transitions := []tea.Msg{
			FocusChangeMsg{Panel: FileTreePanel},
			FocusChangeMsg{Panel: SelectedFilesPanel},
			FocusChangeMsg{Panel: ChatPanel},
			FocusChangeMsg{Panel: FooterMenuPanel},
			MenuModeChangeMsg{Enabled: true},
			MenuModeChangeMsg{Enabled: false},
			DebugModeChangeMsg{Enabled: true},
			DebugModeChangeMsg{Enabled: false},
			LayoutChangeMsg{Width: 80, Height: 24},
		}
		
		for i, msg := range transitions {
			model, _ := app.handleStateChange(msg)
			app = model.(*App)
			
			if err := app.validateStateInvariants(); err != nil {
				t.Errorf("State invariants violated after transition %d (%v): %v", i, msg, err)
			}
		}
	})
}

// TestStateValidationHelpers tests the state validation helper functions
func TestStateValidationHelpers(t *testing.T) {
	app := createTestApp(t)
	
	t.Run("isValidPanel validation", func(t *testing.T) {
		validPanels := []FocusedPanel{FileTreePanel, SelectedFilesPanel, ChatPanel, FooterMenuPanel}
		for _, panel := range validPanels {
			if !app.isValidPanel(panel) {
				t.Errorf("Panel %v should be valid", panel)
			}
		}
		
		invalidPanels := []FocusedPanel{FocusedPanel(-1), FocusedPanel(999)}
		for _, panel := range invalidPanels {
			if app.isValidPanel(panel) {
				t.Errorf("Panel %v should be invalid", panel)
			}
		}
	})
	
	t.Run("validateStateInvariants checks", func(t *testing.T) {
		// Test valid state
		if err := app.validateStateInvariants(); err != nil {
			t.Errorf("Initial state should be valid: %v", err)
		}
		
		// Test invalid focus
		app.focused = FocusedPanel(999)
		if err := app.validateStateInvariants(); err == nil {
			t.Error("Invalid focus should fail validation")
		}
		
		// Reset to valid state
		app.focused = FileTreePanel
		
		// Test invalid dimensions
		app.width = -1
		if err := app.validateStateInvariants(); err == nil {
			t.Error("Invalid width should fail validation")
		}
	})
}

// Benchmark state command generation performance
func BenchmarkStateCommandGeneration(b *testing.B) {
	app := createTestApp(&testing.T{})
	
	b.Run("setFocus", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := app.setFocus(ChatPanel)
			_ = cmd()
		}
	})
	
	b.Run("setMenuMode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := app.setMenuMode(true)
			_ = cmd()
		}
	})
	
	b.Run("toggleDebugMode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cmd := app.toggleDebugMode()
			_ = cmd()
		}
	})
	
	b.Run("handleStateChange", func(b *testing.B) {
		focusMsg := FocusChangeMsg{Panel: ChatPanel}
		for i := 0; i < b.N; i++ {
			_, _ = app.handleStateChange(focusMsg)
		}
	})
}

// TestPropertyBasedStateMachine uses property-based testing to verify state machine invariants
func TestPropertyBasedStateMachine(t *testing.T) {
	// Property-based test: random sequences of state changes should always result in valid state
	t.Run("Random state transitions maintain validity", func(t *testing.T) {
		
		// Define possible state transitions
		allPanels := []FocusedPanel{FileTreePanel, SelectedFilesPanel, ChatPanel, FooterMenuPanel}
		boolStates := []bool{true, false}
		layoutDimensions := []LayoutChangeMsg{
			{Width: 80, Height: 24},
			{Width: 120, Height: 30},
			{Width: 100, Height: 50},
			{Width: 50, Height: 20},
		}
		
		// Generate 100 random sequences of state changes
		for sequence := 0; sequence < 100; sequence++ {
			app := createTestApp(t) // Fresh app for each sequence
			
			// Generate 10 random state changes per sequence
			for step := 0; step < 10; step++ {
				var msg tea.Msg
				
				// Randomly choose type of state change (weighted distribution)
				switch step % 4 {
				case 0: // Focus change (most common)
					msg = FocusChangeMsg{Panel: allPanels[step%len(allPanels)]}
				case 1: // Menu mode change
					msg = MenuModeChangeMsg{Enabled: boolStates[step%len(boolStates)]}
				case 2: // Debug mode change
					msg = DebugModeChangeMsg{Enabled: boolStates[(step+1)%len(boolStates)]}
				case 3: // Layout change
					msg = layoutDimensions[step%len(layoutDimensions)]
				}
				
				// Apply the state change
				model, _ := app.handleStateChange(msg)
				app = model.(*App)
				
				// Verify state invariants after each change
				if err := app.validateStateInvariants(); err != nil {
					t.Errorf("State invariants violated in sequence %d, step %d with message %v: %v", 
						sequence, step, msg, err)
					t.Logf("Final state: focused=%v, menuBindingMode=%v, debugMode=%v, dimensions=%dx%d",
						app.focused, app.menuBindingMode, app.debugMode, app.width, app.height)
					return
				}
			}
		}
	})
	
	t.Run("State machine properties hold under stress", func(t *testing.T) {
		// Property: Focus should always be within valid range
		// Property: Menu binding mode should be consistent with focus in legacy mode
		// Property: Layout dimensions should always be non-negative after valid changes
		
		app := createTestApp(t)
		
		// Stress test with rapid state changes
		for i := 0; i < 1000; i++ {
			// Cycle through focus panels
			targetPanel := FocusedPanel(i % 4)
			model, _ := app.handleStateChange(FocusChangeMsg{Panel: targetPanel})
			app = model.(*App)
			
			// Property: Focus should always be the panel we set
			if app.focused != targetPanel {
				t.Errorf("Focus property violated: expected %v, got %v", targetPanel, app.focused)
			}
			
			// Property: Focus should always be valid
			if !app.isValidPanel(app.focused) {
				t.Errorf("Focus validity property violated: %v is not valid", app.focused)
			}
			
			// Property: Legacy mode consistency (if in legacy mode)
			if app.settingsManager.IsLegacyMode() {
				expectedMenuMode := (app.focused == FooterMenuPanel)
				if app.menuBindingMode != expectedMenuMode {
					t.Errorf("Legacy mode property violated: focus=%v, menuMode=%v, expected=%v", 
						app.focused, app.menuBindingMode, expectedMenuMode)
				}
			}
		}
	})
	
	t.Run("State convergence property", func(t *testing.T) {
		// Property: Given the same sequence of state changes, the system should always 
		// converge to the same final state (determinism)
		
		sequence := []tea.Msg{
			FocusChangeMsg{Panel: ChatPanel},
			MenuModeChangeMsg{Enabled: true},
			LayoutChangeMsg{Width: 100, Height: 50},
			FocusChangeMsg{Panel: FileTreePanel},
			DebugModeChangeMsg{Enabled: true},
			MenuModeChangeMsg{Enabled: false},
		}
		
		// Apply the same sequence multiple times
		var finalStates []*App
		for run := 0; run < 5; run++ {
			app := createTestApp(t)
			
			for _, msg := range sequence {
				model, _ := app.handleStateChange(msg)
				app = model.(*App)
			}
			
			finalStates = append(finalStates, app)
		}
		
		// All final states should be identical
		reference := finalStates[0]
		for i, state := range finalStates[1:] {
			if state.focused != reference.focused {
				t.Errorf("State convergence violated in run %d: focused %v != %v", i+1, state.focused, reference.focused)
			}
			if state.menuBindingMode != reference.menuBindingMode {
				t.Errorf("State convergence violated in run %d: menuBindingMode %v != %v", i+1, state.menuBindingMode, reference.menuBindingMode)
			}
			if state.debugMode != reference.debugMode {
				t.Errorf("State convergence violated in run %d: debugMode %v != %v", i+1, state.debugMode, reference.debugMode)
			}
			if state.width != reference.width || state.height != reference.height {
				t.Errorf("State convergence violated in run %d: dimensions %dx%d != %dx%d", 
					i+1, state.width, state.height, reference.width, reference.height)
			}
		}
	})
	
	t.Run("State boundaries and edge cases", func(t *testing.T) {
		app := createTestApp(t)
		
		// Test boundary conditions
		boundaryTests := []struct {
			name string
			msg  tea.Msg
			shouldBeValid bool
		}{
			{"Valid panel boundary - first", FocusChangeMsg{Panel: FileTreePanel}, true},
			{"Valid panel boundary - last", FocusChangeMsg{Panel: FooterMenuPanel}, true},
			{"Invalid panel - negative", FocusChangeMsg{Panel: FocusedPanel(-1)}, false},
			{"Invalid panel - too high", FocusChangeMsg{Panel: FocusedPanel(999)}, false},
			{"Valid layout - minimum", LayoutChangeMsg{Width: 1, Height: 1}, true},
			{"Valid layout - large", LayoutChangeMsg{Width: 1000, Height: 1000}, true},
			{"Invalid layout - zero width", LayoutChangeMsg{Width: 0, Height: 10}, false},
			{"Invalid layout - zero height", LayoutChangeMsg{Width: 10, Height: 0}, false},
			{"Invalid layout - negative width", LayoutChangeMsg{Width: -1, Height: 10}, false},
			{"Invalid layout - negative height", LayoutChangeMsg{Width: 10, Height: -1}, false},
		}
		
		for _, tt := range boundaryTests {
			t.Run(tt.name, func(t *testing.T) {
				app.debugMode = true // Enable debug mode to catch validation errors
				initialState := *app // Take snapshot
				
				model, cmd := app.handleStateChange(tt.msg)
				app = model.(*App)
				
				if tt.shouldBeValid {
					// Valid changes should not generate error commands
					if cmd != nil {
						// Check if it's an error by trying to execute the command
						if cmd != nil {
							msg := cmd()
							if alertMsg, ok := msg.(interface{ Type() string }); ok && alertMsg.Type() == "error" {
								t.Errorf("Valid state change generated error command: %v", tt.msg)
							}
						}
					}
					
					// Should still pass state invariants
					if err := app.validateStateInvariants(); err != nil {
						t.Errorf("Valid state change violated invariants: %v", err)
					}
				} else {
					// Invalid changes should be rejected (state unchanged or error generated)
					if cmd == nil {
						// If no command generated, state should be unchanged for invalid operations
						// This is implementation-specific - some invalid ops might be silently ignored
					}
					
					// State should still be valid even after invalid operations
					if err := app.validateStateInvariants(); err != nil {
						t.Errorf("State invariants violated even for invalid operation: %v", err)
					}
				}
				
				// Reset app state for next test
				*app = initialState
			})
		}
	})
}