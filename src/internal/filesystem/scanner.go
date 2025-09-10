package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

// FileNode represents a file or directory in the filesystem
type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Children []*FileNode
}

// ScanDirectory recursively scans a directory and returns a tree structure
func ScanDirectory(rootPath string) (*FileNode, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}

	root := &FileNode{
		Name:     filepath.Base(rootPath),
		Path:     rootPath,
		IsDir:    info.IsDir(),
		Children: []*FileNode{},
	}

	if !info.IsDir() {
		return root, nil
	}

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return root, err
	}

	for _, entry := range entries {
		// Skip hidden files and common ignore patterns
		if shouldIgnore(entry.Name()) {
			continue
		}

		childPath := filepath.Join(rootPath, entry.Name())
		child, err := ScanDirectory(childPath)
		if err != nil {
			// Skip files we can't read
			continue
		}

		root.Children = append(root.Children, child)
	}

	return root, nil
}

// shouldIgnore determines if a file or directory should be ignored
func shouldIgnore(name string) bool {
	// Common ignore patterns
	ignorePatterns := []string{
		".",        // Hidden files (starting with .)
		"node_modules",
		".git",
		".svn",
		".hg",
		"vendor",
		"target",
		"build",
		"dist",
		"__pycache__",
		".DS_Store",
		"Thumbs.db",
	}

	// Check if starts with dot (hidden files)
	if strings.HasPrefix(name, ".") {
		return true
	}

	// Check against ignore patterns
	for _, pattern := range ignorePatterns {
		if name == pattern {
			return true
		}
	}

	return false
}

// FlattenTree converts a tree structure to a flat list for display
func FlattenTree(root *FileNode, level int, expanded map[string]bool) []FileTreeItem {
	var items []FileTreeItem

	item := FileTreeItem{
		Name:     root.Name,
		Path:     root.Path,
		IsDir:    root.IsDir,
		Level:    level,
		Expanded: expanded[root.Path],
	}
	items = append(items, item)

	// If it's a directory and expanded, add children
	if root.IsDir && expanded[root.Path] {
		for _, child := range root.Children {
			childItems := FlattenTree(child, level+1, expanded)
			items = append(items, childItems...)
		}
	}

	return items
}

// FileTreeItem represents an item in the flattened tree view
type FileTreeItem struct {
	Name     string
	Path     string
	IsDir    bool
	Level    int
	Expanded bool
	Selected bool
}

// GetFileContent reads and returns the content of a file
func GetFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// IsTextFile determines if a file is likely to be a text file
func IsTextFile(filePath string) bool {
	// Simple check based on file extension
	textExtensions := []string{
		".txt", ".md", ".rst", ".log",
		".go", ".js", ".ts", ".py", ".rb", ".php", ".java", ".c", ".cpp", ".h", ".hpp",
		".html", ".css", ".scss", ".sass", ".xml", ".json", ".yaml", ".yml", ".toml",
		".sh", ".bash", ".zsh", ".fish",
		".sql", ".dockerfile", ".makefile",
		".vue", ".jsx", ".tsx",
		".rs", ".elm", ".hs", ".ml", ".fs",
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	for _, validExt := range textExtensions {
		if ext == validExt {
			return true
		}
	}

	// Check for files without extensions that are commonly text
	base := strings.ToLower(filepath.Base(filePath))
	textFiles := []string{
		"makefile", "dockerfile", "readme", "license", "changelog",
		"authors", "contributors", "copying", "install", "news",
	}

	for _, validFile := range textFiles {
		if base == validFile {
			return true
		}
	}

	return false
}