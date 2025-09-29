package prompt

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"coding-prompts-tui/internal/filesystem"
)

type cdata struct {
	Text string `xml:",cdata"`
}

type File struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	Content string   `xml:",cdata"`
}

type SystemPrompt struct {
	XMLName xml.Name `xml:"SystemPrompt"`
	Type    string   `xml:"type,attr,omitempty"`
	Content string   `xml:",cdata"`
}

type Prompt struct {
	XMLName      xml.Name       `xml:"prompt"`
	FileTree     cdata          `xml:"filetree"`
	Files        []File         `xml:"file"`
	SystemPrompt []SystemPrompt `xml:"SystemPrompt"`
	UserPrompt   cdata          `xml:"UserPrompt"`
}

func Build(rootPath string, selectedFiles map[string]bool, userPrompt string, activePersonas []string) (string, error) {
	// 1. Generate file tree
	fileTree, err := generateFileTree(rootPath)
	if err != nil {
		return "", fmt.Errorf("error generating file tree: %w", err)
	}

	// 2. Get selected file contents
	var files []File
	for path, selected := range selectedFiles {
		if selected {
			content, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("error reading file %s: %w", path, err)
			}
			relativePath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return "", fmt.Errorf("error getting relative path for %s: %w", path, err)
			}
			files = append(files, File{Name: relativePath, Content: string(content)})
		}
	}

	var systemPrompts []SystemPrompt

	// 3. Get project overview
	overviewContent, err := getProjectOverview(rootPath)
	if err == nil && overviewContent != "" {
		systemPrompts = append(systemPrompts, SystemPrompt{
			Type:    "project-overview",
			Content: overviewContent,
		})
	}

	// 4. Get system prompts for active personas
	if len(activePersonas) == 0 {
		activePersonas = []string{"default"}
	}

	for _, persona := range activePersonas {
		personaPath := filepath.Join(rootPath, "personas", persona+".md")
		systemPromptContent, err := os.ReadFile(personaPath)
		if err != nil {
			// If persona file doesn't exist, use a fallback
			systemPromptContent = []byte(fmt.Sprintf("You are a helpful AI assistant with the %s persona.", persona))
		}
		systemPrompts = append(systemPrompts, SystemPrompt{
			Type:    persona,
			Content: string(systemPromptContent),
		})
	}

	// 5. Construct the prompt struct
	prompt := Prompt{
		FileTree:     cdata{Text: fileTree},
		Files:        files,
		SystemPrompt: systemPrompts,
		UserPrompt:   cdata{Text: userPrompt},
	}

	// 6. Marshal to XML
	xmlOutput, err := xml.MarshalIndent(prompt, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling to xml: %w", err)
	}

	return string(xmlOutput), nil
}

func getProjectOverview(rootPath string) (string, error) {
	overviewFiles := []string{"CLAUDE.md", "GEMINI.md", "README.md"}
	for _, filename := range overviewFiles {
		path := filepath.Join(rootPath, filename)
		if _, err := os.Stat(path); err == nil {
			content, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("error reading overview file %s: %w", path, err)
			}
			return string(content), nil
		}
	}
	return "", nil // No overview file found
}

func generateFileTree(rootPath string) (string, error) {
	// Try to use gitignore-aware generation
	tree, err := generateFileTreeWithGitignore(rootPath)
	if err != nil {
		// Fall back to legacy generation if gitignore fails
		return generateFileTreeLegacy(rootPath)
	}
	return tree, nil
}

func generateFileTreeWithGitignore(rootPath string) (string, error) {
	// Create gitignore matcher
	matcher, err := filesystem.NewGitignoreMatcher(rootPath)
	if err != nil {
		return "", err
	}

	var tree strings.Builder
	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Don't apply ignore rules to the root directory itself
		if path != rootPath {
			// Skip ignored files and directories using gitignore patterns
			if matcher.ShouldIgnore(path, info.IsDir()) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)
		if info.IsDir() {
			fmt.Fprintf(&tree, "%s- %s/\n", indent, info.Name())
		} else {
			fmt.Fprintf(&tree, "%s- %s\n", indent, info.Name())
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return tree.String(), nil
}

func generateFileTreeLegacy(rootPath string) (string, error) {
	var tree strings.Builder
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Don't apply ignore rules to the root directory itself
		if path != rootPath {
			// Skip ignored files and directories
			if filesystem.ShouldIgnore(info.Name()) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		depth := strings.Count(relPath, string(os.PathSeparator))
		indent := strings.Repeat("  ", depth)
		if info.IsDir() {
			fmt.Fprintf(&tree, "%s- %s/\n", indent, info.Name())
		} else {
			fmt.Fprintf(&tree, "%s- %s\n", indent, info.Name())
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return tree.String(), nil
}
