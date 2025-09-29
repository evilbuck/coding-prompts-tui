package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGitignoreMatcher(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create a test .gitignore file
	gitignoreContent := `# Test gitignore
*.log
*.tmp
build/
.git/
node_modules/
!important.log
`
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .gitignore: %v", err)
	}

	// Create the gitignore matcher
	matcher, err := NewGitignoreMatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create gitignore matcher: %v", err)
	}

	// Test cases
	tests := []struct {
		path         string
		isDir        bool
		shouldIgnore bool
		description  string
	}{
		{filepath.Join(tmpDir, "test.log"), false, true, "*.log pattern should match"},
		{filepath.Join(tmpDir, "test.txt"), false, false, "regular files should not be ignored"},
		{filepath.Join(tmpDir, "important.log"), false, false, "negation pattern should override"},
		{filepath.Join(tmpDir, "build"), true, true, "build/ directory should be ignored"},
		{filepath.Join(tmpDir, "build", "output.txt"), false, true, "files in ignored directories should be ignored"},
		{filepath.Join(tmpDir, ".git"), true, true, ".git/ directory should be ignored"},
		{filepath.Join(tmpDir, ".git", "config"), false, true, "files in .git should be ignored"},
		{filepath.Join(tmpDir, "node_modules"), true, true, "node_modules/ should be ignored"},
		{filepath.Join(tmpDir, "src", "main.go"), false, false, "normal source files should not be ignored"},
		{filepath.Join(tmpDir, "temp.tmp"), false, true, "*.tmp files should be ignored"},
	}

	for _, test := range tests {
		result := matcher.ShouldIgnore(test.path, test.isDir)
		if result != test.shouldIgnore {
			t.Errorf("%s: expected %t, got %t for path %s",
				test.description, test.shouldIgnore, result, test.path)
		}
	}
}

func TestGitignoreMatcherWithoutFile(t *testing.T) {
	// Create a temporary directory without .gitignore
	tmpDir := t.TempDir()

	// Create the gitignore matcher - should use defaults
	matcher, err := NewGitignoreMatcher(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create gitignore matcher: %v", err)
	}

	// Test that default patterns are applied
	tests := []struct {
		path         string
		isDir        bool
		shouldIgnore bool
		description  string
	}{
		{filepath.Join(tmpDir, ".git"), true, true, "default .git/ should be ignored"},
		{filepath.Join(tmpDir, "node_modules"), true, true, "default node_modules/ should be ignored"},
		{filepath.Join(tmpDir, ".DS_Store"), false, true, "default .DS_Store should be ignored"},
		{filepath.Join(tmpDir, "main.go"), false, false, "normal files should not be ignored"},
		{filepath.Join(tmpDir, "build"), true, true, "default build/ should be ignored"},
	}

	for _, test := range tests {
		result := matcher.ShouldIgnore(test.path, test.isDir)
		if result != test.shouldIgnore {
			t.Errorf("%s: expected %t, got %t for path %s",
				test.description, test.shouldIgnore, result, test.path)
		}
	}
}

func TestGitignorePatternParsing(t *testing.T) {
	tmpDir := t.TempDir()
	matcher := &GitignoreMatcher{rootPath: tmpDir}

	// Test pattern parsing
	tests := []struct {
		line        string
		expectNil   bool
		isNegative  bool
		isDir       bool
		description string
	}{
		{"*.log", false, false, false, "simple wildcard pattern"},
		{"!important.log", false, true, false, "negation pattern"},
		{"build/", false, false, true, "directory pattern"},
		{"# comment", true, false, false, "comment should be ignored"},
		{"", true, false, false, "empty line should be ignored"},
		{"/absolute", false, false, false, "absolute path pattern"},
	}

	for _, test := range tests {
		pattern := matcher.parsePattern(test.line)

		if test.expectNil {
			if pattern != nil {
				t.Errorf("%s: expected nil pattern, got %+v", test.description, pattern)
			}
			continue
		}

		if pattern == nil {
			t.Errorf("%s: expected non-nil pattern", test.description)
			continue
		}

		if pattern.IsNegative != test.isNegative {
			t.Errorf("%s: expected IsNegative=%t, got %t", test.description, test.isNegative, pattern.IsNegative)
		}

		if pattern.IsDir != test.isDir {
			t.Errorf("%s: expected IsDir=%t, got %t", test.description, test.isDir, pattern.IsDir)
		}
	}
}
