package filesystem

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GitignorePattern represents a single pattern from .gitignore
type GitignorePattern struct {
	Pattern    string
	IsNegative bool
	IsDir      bool
	Regex      *regexp.Regexp
}

// GitignoreMatcher handles .gitignore pattern matching
type GitignoreMatcher struct {
	patterns []GitignorePattern
	rootPath string
}

// NewGitignoreMatcher creates a new gitignore matcher for the given root path
func NewGitignoreMatcher(rootPath string) (*GitignoreMatcher, error) {
	matcher := &GitignoreMatcher{
		rootPath: rootPath,
	}

	// Try to load .gitignore from the root path
	gitignorePath := filepath.Join(rootPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); err == nil {
		if err := matcher.loadGitignore(gitignorePath); err != nil {
			return nil, err
		}
	}

	// If no .gitignore or it's empty, add some sensible defaults
	if len(matcher.patterns) == 0 {
		matcher.addDefaultPatterns()
	}

	return matcher, nil
}

// loadGitignore parses a .gitignore file and loads the patterns
func (gm *GitignoreMatcher) loadGitignore(gitignorePath string) error {
	file, err := os.Open(gitignorePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern := gm.parsePattern(line)
		if pattern != nil {
			gm.patterns = append(gm.patterns, *pattern)
		}
	}

	return scanner.Err()
}

// parsePattern converts a gitignore pattern line into a GitignorePattern
func (gm *GitignoreMatcher) parsePattern(line string) *GitignorePattern {
	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	pattern := GitignorePattern{}

	// Handle negation (lines starting with !)
	if strings.HasPrefix(line, "!") {
		pattern.IsNegative = true
		line = line[1:]
	}

	// Handle directory-only patterns (ending with /)
	if strings.HasSuffix(line, "/") {
		pattern.IsDir = true
		line = line[:len(line)-1]
	}

	// Escape special regex characters except for gitignore wildcards
	pattern.Pattern = line
	
	// Convert gitignore patterns to regex
	regexPattern := gm.gitignoreToRegex(line)
	
	// Compile the regex
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		// If regex compilation fails, skip this pattern
		return nil
	}
	
	pattern.Regex = regex
	return &pattern
}

// gitignoreToRegex converts gitignore patterns to regex patterns
func (gm *GitignoreMatcher) gitignoreToRegex(pattern string) string {
	// Escape regex special characters except for gitignore wildcards
	escaped := regexp.QuoteMeta(pattern)
	
	// Convert gitignore wildcards to regex
	escaped = strings.ReplaceAll(escaped, `\*\*`, ".*")  // ** matches any number of directories
	escaped = strings.ReplaceAll(escaped, `\*`, "[^/]*") // * matches anything except /
	escaped = strings.ReplaceAll(escaped, `\?`, "[^/]")  // ? matches any single character except /
	
	// Handle leading slash (absolute path from repo root)
	if strings.HasPrefix(pattern, "/") {
		escaped = "^" + escaped[1:] // Remove leading slash and anchor to start
	} else {
		// Pattern can match at any level
		escaped = "(^|/)" + escaped
	}
	
	// For directory patterns, also match anything inside the directory
	// For file patterns, also match if it's inside an ignored directory
	escaped = escaped + "(/.*)?$"
	
	return escaped
}

// addDefaultPatterns adds sensible default ignore patterns when no .gitignore exists
func (gm *GitignoreMatcher) addDefaultPatterns() {
	defaultPatterns := []string{
		".git/",
		".svn/",
		".hg/",
		"node_modules/",
		"vendor/",
		"target/",
		"build/",
		"dist/",
		"__pycache__/",
		".DS_Store",
		"Thumbs.db",
		"*.tmp",
		"*.log",
	}

	for _, pattern := range defaultPatterns {
		parsed := gm.parsePattern(pattern)
		if parsed != nil {
			gm.patterns = append(gm.patterns, *parsed)
		}
	}
}

// ShouldIgnore determines if a file or directory should be ignored based on .gitignore patterns
func (gm *GitignoreMatcher) ShouldIgnore(path string, isDir bool) bool {
	// Convert absolute path to relative path from root
	relPath, err := filepath.Rel(gm.rootPath, path)
	if err != nil {
		// If we can't get relative path, use the basename
		relPath = filepath.Base(path)
	}

	// Normalize path separators for cross-platform compatibility
	relPath = filepath.ToSlash(relPath)

	// Don't ignore the root directory itself
	if relPath == "." {
		return false
	}

	// Check against all patterns
	ignored := false
	for _, pattern := range gm.patterns {
		// Check if pattern matches
		matches := pattern.Regex.MatchString(relPath)
		
		// For directory-only patterns, only apply to actual directories
		// But if a file is inside an ignored directory, it should also be ignored
		if pattern.IsDir && !isDir {
			// Check if this file is inside the ignored directory
			// The regex should handle this with the (/.*)?$ suffix
			matches = matches // Keep the matches as-is since regex handles subdirectories
		}
		
		if matches {
			if pattern.IsNegative {
				ignored = false // Negation pattern overrides previous ignore
			} else {
				ignored = true
			}
		}
	}

	return ignored
}