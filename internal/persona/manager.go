package persona

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Manager handles persona discovery and management
type Manager struct {
	personasDir string
	personas    []string
}

// NewManager creates a new persona manager
func NewManager(rootDir string) *Manager {
	return &Manager{
		personasDir: filepath.Join(rootDir, "personas"),
	}
}

// DiscoverPersonas scans the personas directory for available personas
func (m *Manager) DiscoverPersonas() error {
	// Check if personas directory exists
	if _, err := os.Stat(m.personasDir); os.IsNotExist(err) {
		return fmt.Errorf("personas directory not found: %s", m.personasDir)
	}

	entries, err := os.ReadDir(m.personasDir)
	if err != nil {
		return fmt.Errorf("failed to read personas directory: %w", err)
	}

	var personas []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only consider .md files
		if strings.HasSuffix(entry.Name(), ".md") {
			// Remove .md extension to get persona name
			personaName := strings.TrimSuffix(entry.Name(), ".md")
			personas = append(personas, personaName)
		}
	}

	// Sort personas alphabetically
	sort.Strings(personas)
	m.personas = personas

	return nil
}

// GetAvailablePersonas returns the list of discovered personas
func (m *Manager) GetAvailablePersonas() []string {
	return append([]string{}, m.personas...) // Return a copy
}

// ValidatePersonas checks if the given personas exist
func (m *Manager) ValidatePersonas(personas []string) []string {
	var valid []string
	personaSet := make(map[string]bool)

	// Create a set of available personas for quick lookup
	for _, p := range m.personas {
		personaSet[p] = true
	}

	for _, persona := range personas {
		if personaSet[persona] {
			valid = append(valid, persona)
		}
	}

	// If no valid personas, return default
	if len(valid) == 0 {
		if personaSet["default"] {
			return []string{"default"}
		}
		// If even default doesn't exist, return first available or empty
		if len(m.personas) > 0 {
			return []string{m.personas[0]}
		}
		return []string{}
	}

	return valid
}

// GetPersonaPath returns the file path for a given persona
func (m *Manager) GetPersonaPath(persona string) string {
	return filepath.Join(m.personasDir, persona+".md")
}

// PersonaExists checks if a specific persona file exists
func (m *Manager) PersonaExists(persona string) bool {
	path := m.GetPersonaPath(persona)
	_, err := os.Stat(path)
	return err == nil
}

// ReadPersonaContent reads the content of a persona file
func (m *Manager) ReadPersonaContent(persona string) (string, error) {
	path := m.GetPersonaPath(persona)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read persona %s: %w", persona, err)
	}
	return string(content), nil
}
