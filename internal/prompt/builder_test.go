package prompt

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	// 1. Create a temporary directory for our test file system
	tmpDir := t.TempDir()

	// Create a dummy personas file
	personasDir := filepath.Join(tmpDir, "personas")
	err := os.Mkdir(personasDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create personas dir: %v", err)
	}
	systemPromptContent := "You are a test assistant."
	err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte(systemPromptContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy system prompt: %v", err)
	}

	// Create a dummy overview file
	overviewContent := "This is the project overview."
	err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(overviewContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy overview file: %v", err)
	}

	// 2. Create some dummy files and directories to build a tree
	file1Content := "hello world"
	err = os.WriteFile(filepath.Join(tmpDir, "testfile1.txt"), []byte(file1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy file 1: %v", err)
	}

	err = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "subdir", "testfile2.txt"), []byte("another file"), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy file 2: %v", err)
	}

	// Change to the temp directory to mimic running from the project root
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalWd) // Change back when the test is done

	// 3. Define the inputs for the Build function
	selectedFiles := map[string]bool{
		filepath.Join(tmpDir, "testfile1.txt"):           true,
		filepath.Join(tmpDir, "subdir", "testfile2.txt"): false, // This one is not selected
	}
	userPrompt := "This is a test user prompt."

	// 4. Call the Build function
	// We pass tmpDir as the root path
	xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt, []string{"default"})

	// 5. Assert the output
	if err != nil {
		t.Fatalf("Build() returned an unexpected error: %v", err)
	}

	// Check for root prompt element
	if !strings.Contains(xmlOutput, "<prompt>") || !strings.Contains(xmlOutput, "</prompt>") {
		t.Error("Expected XML to contain root prompt element")
	}

	// Check for file tree
	if !strings.Contains(xmlOutput, "<filetree>") {
		t.Error("Expected XML to contain a filetree element")
	}
	if !strings.Contains(xmlOutput, "- testfile1.txt") {
		t.Error("Expected file tree to contain 'testfile1.txt'")
	}
	if !strings.Contains(xmlOutput, "- subdir/") {
		t.Error("Expected file tree to contain 'subdir/'")
	}
	if !strings.Contains(xmlOutput, "- testfile2.txt") {
		t.Error("Expected file tree to contain 'testfile2.txt' (unselected files should appear in tree)")
	}

	// Check for selected file content
	expectedFileXML := `<file name="testfile1.txt"><![CDATA[hello world]]></file>`
	if !strings.Contains(xmlOutput, expectedFileXML) {
		t.Errorf("Expected XML to contain the selected file content.\nGot:\n%s\nExpected to contain:\n%s", xmlOutput, expectedFileXML)
	}

	// Check that unselected file content is NOT included as a file element
	unselectedFileXML := `<file name="subdir/testfile2.txt">`
	if strings.Contains(xmlOutput, unselectedFileXML) {
		t.Error("XML should not contain file element for unselected file 'subdir/testfile2.txt'")
	}

	// Check for overview system prompt
	expectedOverviewPrompt := `<SystemPrompt type="project-overview"><![CDATA[This is the project overview.]]></SystemPrompt>`
	if !strings.Contains(xmlOutput, expectedOverviewPrompt) {
		t.Errorf("Expected XML to contain the overview prompt.\nGot:\n%s\nExpected to contain:\n%s", xmlOutput, expectedOverviewPrompt)
	}

	// Check for default system prompt
	expectedSystemPrompt := `<SystemPrompt type="default"><![CDATA[You are a test assistant.]]></SystemPrompt>`
	if !strings.Contains(xmlOutput, expectedSystemPrompt) {
		t.Errorf("Expected XML to contain the system prompt.\nGot:\n%s\nExpected to contain:\n%s", xmlOutput, expectedSystemPrompt)
	}

	// Check for user prompt
	expectedUserPrompt := `<UserPrompt><![CDATA[This is a test user prompt.]]></UserPrompt>`
	if !strings.Contains(xmlOutput, expectedUserPrompt) {
		t.Errorf("Expected XML to contain the user prompt.\nGot:\n%s\nExpected to contain:\n%s", xmlOutput, expectedUserPrompt)
	}

	// Validate that the output is well-formed XML
	var prompt Prompt
	err = xml.Unmarshal([]byte(xmlOutput), &prompt)
	if err != nil {
		t.Errorf("Generated XML is not well-formed: %v\nXML:\n%s", err, xmlOutput)
	}

	// Validate the parsed structure
	if prompt.XMLName.Local != "prompt" {
		t.Errorf("Expected root element to be 'prompt', got '%s'", prompt.XMLName.Local)
	}
	if len(prompt.Files) != 1 {
		t.Errorf("Expected 1 file in prompt, got %d", len(prompt.Files))
	}
	if prompt.Files[0].Name != "testfile1.txt" {
		t.Errorf("Expected file name 'testfile1.txt', got '%s'", prompt.Files[0].Name)
	}
	if prompt.Files[0].Content != "hello world" {
		t.Errorf("Expected file content 'hello world', got '%s'", prompt.Files[0].Content)
	}
	if len(prompt.SystemPrompt) != 2 {
		t.Fatalf("Expected 2 system prompts, got %d", len(prompt.SystemPrompt))
	}
	// Note: order is not guaranteed by map iteration, so we check both
	overviewPrompt := prompt.SystemPrompt[0]
	defaultPrompt := prompt.SystemPrompt[1]

	if overviewPrompt.Type != "project-overview" {
		t.Errorf("Expected first system prompt to have type 'project-overview', got '%s'", overviewPrompt.Type)
	}
	if overviewPrompt.Content != "This is the project overview." {
		t.Errorf("Expected overview prompt content 'This is the project overview.', got '%s'", overviewPrompt.Content)
	}
	if defaultPrompt.Type != "default" {
		t.Errorf("Expected second system prompt to have type 'default', got '%s'", defaultPrompt.Type)
	}
	if defaultPrompt.Content != "You are a test assistant." {
		t.Errorf("Expected system prompt 'You are a test assistant.', got '%s'", defaultPrompt.Content)
	}

	if prompt.UserPrompt.Text != "This is a test user prompt." {
		t.Errorf("Expected user prompt 'This is a test user prompt.', got '%s'", prompt.UserPrompt.Text)
	}
}

func TestBuildErrorConditions(t *testing.T) {
	t.Run("invalid root path", func(t *testing.T) {
		selectedFiles := map[string]bool{}
		userPrompt := "test prompt"

		_, err := Build("/nonexistent/path", selectedFiles, userPrompt, []string{"default"})
		if err == nil {
			t.Error("Expected error for invalid root path, got nil")
		}
	})

	t.Run("invalid selected file path", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create personas directory with default.md
		personasDir := filepath.Join(tmpDir, "personas")
		err := os.Mkdir(personasDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create personas dir: %v", err)
		}
		err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to write dummy system prompt: %v", err)
		}

		selectedFiles := map[string]bool{
			"/nonexistent/file.txt": true,
		}
		userPrompt := "test prompt"

		_, err = Build(tmpDir, selectedFiles, userPrompt, []string{"default"})
		if err == nil {
			t.Error("Expected error for nonexistent selected file, got nil")
		}
	})

	t.Run("unreadable selected file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create personas directory with default.md
		personasDir := filepath.Join(tmpDir, "personas")
		err := os.Mkdir(personasDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create personas dir: %v", err)
		}
		err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to write dummy system prompt: %v", err)
		}

		// Create a file with no read permissions
		unreadableFile := filepath.Join(tmpDir, "unreadable.txt")
		err = os.WriteFile(unreadableFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Remove read permissions
		err = os.Chmod(unreadableFile, 0000)
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}
		defer os.Chmod(unreadableFile, 0644) // Restore for cleanup

		selectedFiles := map[string]bool{
			unreadableFile: true,
		}
		userPrompt := "test prompt"

		_, err = Build(tmpDir, selectedFiles, userPrompt, []string{"default"})
		if err == nil {
			t.Error("Expected error for unreadable selected file, got nil")
		}
	})
}

func TestBuildEmptyInputs(t *testing.T) {
	t.Run("no selected files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Change to temp directory for personas/default.md lookup
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current working directory: %v", err)
		}
		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		defer os.Chdir(originalWd)

		// Create personas directory with default.md
		personasDir := filepath.Join(tmpDir, "personas")
		err = os.Mkdir(personasDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create personas dir: %v", err)
		}
		systemPromptContent := "You are a test assistant."
		err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte(systemPromptContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write dummy system prompt: %v", err)
		}

		// Create a test file (but don't select it)
		err = os.WriteFile(filepath.Join(tmpDir, "testfile.txt"), []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		selectedFiles := map[string]bool{} // No files selected
		userPrompt := "This is a test prompt with no files."

		xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt, []string{"default"})
		if err != nil {
			t.Fatalf("Build() returned an unexpected error: %v", err)
		}

		// Should still contain filetree and prompts, but no file elements
		if !strings.Contains(xmlOutput, "<filetree>") {
			t.Error("Expected XML to contain filetree even with no selected files")
		}
		if !strings.Contains(xmlOutput, "- testfile.txt") {
			t.Error("Expected filetree to contain unselected file")
		}
		if strings.Contains(xmlOutput, "<file name=") {
			t.Error("Expected no file elements when no files are selected")
		}
		if !strings.Contains(xmlOutput, systemPromptContent) {
			t.Error("Expected system prompt to be included")
		}
		if !strings.Contains(xmlOutput, userPrompt) {
			t.Error("Expected user prompt to be included")
		}

		// Validate XML structure
		var prompt Prompt
		err = xml.Unmarshal([]byte(xmlOutput), &prompt)
		if err != nil {
			t.Errorf("Generated XML is not well-formed: %v", err)
		}
		if len(prompt.Files) != 0 {
			t.Errorf("Expected 0 files in prompt, got %d", len(prompt.Files))
		}
	})

	t.Run("empty user prompt", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Change to temp directory for personas/default.md lookup
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current working directory: %v", err)
		}
		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		defer os.Chdir(originalWd)

		// Create personas directory with default.md
		personasDir := filepath.Join(tmpDir, "personas")
		err = os.Mkdir(personasDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create personas dir: %v", err)
		}
		systemPromptContent := "You are a test assistant."
		err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte(systemPromptContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write dummy system prompt: %v", err)
		}

		// Create and select a test file
		testFile := filepath.Join(tmpDir, "testfile.txt")
		err = os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		selectedFiles := map[string]bool{
			testFile: true,
		}
		userPrompt := "" // Empty user prompt

		xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt, []string{"default"})
		if err != nil {
			t.Fatalf("Build() returned an unexpected error: %v", err)
		}

		// Should contain all elements including empty user prompt
		// Empty user prompt generates <UserPrompt></UserPrompt> (no CDATA when empty)
		if !strings.Contains(xmlOutput, "<UserPrompt></UserPrompt>") {
			t.Error("Expected empty user prompt element")
		}

		// Validate XML structure
		var prompt Prompt
		err = xml.Unmarshal([]byte(xmlOutput), &prompt)
		if err != nil {
			t.Errorf("Generated XML is not well-formed: %v", err)
		}
		if prompt.UserPrompt.Text != "" {
			t.Errorf("Expected empty user prompt, got '%s'", prompt.UserPrompt.Text)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Change to temp directory for personas/default.md lookup
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current working directory: %v", err)
		}
		err = os.Chdir(tmpDir)
		if err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}
		defer os.Chdir(originalWd)

		// Create only personas directory with default.md
		personasDir := filepath.Join(tmpDir, "personas")
		err = os.Mkdir(personasDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create personas dir: %v", err)
		}
		systemPromptContent := "You are a test assistant."
		err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte(systemPromptContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write dummy system prompt: %v", err)
		}

		selectedFiles := map[string]bool{}
		userPrompt := "Test with empty directory"

		xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt, []string{"default"})
		if err != nil {
			t.Fatalf("Build() returned an unexpected error: %v", err)
		}

		// Should contain minimal filetree with just personas
		if !strings.Contains(xmlOutput, "<filetree>") {
			t.Error("Expected XML to contain filetree")
		}
		if !strings.Contains(xmlOutput, "- personas/") {
			t.Error("Expected filetree to contain personas directory")
		}

		// Validate XML structure
		var prompt Prompt
		err = xml.Unmarshal([]byte(xmlOutput), &prompt)
		if err != nil {
			t.Errorf("Generated XML is not well-formed: %v", err)
		}
	})
}

func TestFileTreeFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory for personas/default.md lookup
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Create personas directory with default.md
	personasDir := filepath.Join(tmpDir, "personas")
	err = os.Mkdir(personasDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create personas dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy system prompt: %v", err)
	}

	// Create a complex directory structure
	err = os.WriteFile(filepath.Join(tmpDir, "root-file.txt"), []byte("root content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}

	// Create nested directories and files
	level1Dir := filepath.Join(tmpDir, "level1")
	err = os.Mkdir(level1Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create level1 dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(level1Dir, "level1-file.txt"), []byte("level1 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create level1 file: %v", err)
	}

	level2Dir := filepath.Join(level1Dir, "level2")
	err = os.Mkdir(level2Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create level2 dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(level2Dir, "level2-file.txt"), []byte("level2 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create level2 file: %v", err)
	}

	// Create another top-level directory
	siblingDir := filepath.Join(tmpDir, "sibling")
	err = os.Mkdir(siblingDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sibling dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(siblingDir, "sibling-file.txt"), []byte("sibling content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sibling file: %v", err)
	}

	selectedFiles := map[string]bool{
		filepath.Join(tmpDir, "root-file.txt"): true,
	}
	userPrompt := "Test file tree format"

	xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt, []string{"default"})
	if err != nil {
		t.Fatalf("Build() returned an unexpected error: %v", err)
	}

	// Extract the filetree content
	start := strings.Index(xmlOutput, "<filetree><![CDATA[")
	end := strings.Index(xmlOutput, "]]></filetree>")
	if start == -1 || end == -1 {
		t.Fatal("Could not find filetree CDATA section in XML")
	}

	filetreeContent := xmlOutput[start+len("<filetree><![CDATA[") : end]
	lines := strings.Split(strings.TrimSpace(filetreeContent), "\n")

	// Validate file tree structure and indentation
	expectedStructure := map[string]int{
		"- level1/":          0, // No indentation for top level
		"- level1-file.txt":  2, // 2 spaces indentation for level1 content
		"- level2/":          2, // 2 spaces indentation for level1 content
		"- level2-file.txt":  4, // 4 spaces indentation for level2 content
		"- personas/":        0, // No indentation for top level
		"- default.md":       2, // 2 spaces indentation for personas content
		"- root-file.txt":    0, // No indentation for top level
		"- sibling/":         0, // No indentation for top level
		"- sibling-file.txt": 2, // 2 spaces indentation for sibling content
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Count leading spaces
		leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))

		// Check if this line matches our expected structure
		if expectedIndent, exists := expectedStructure[trimmed]; exists {
			if leadingSpaces != expectedIndent {
				t.Errorf("Expected %d spaces for '%s', got %d", expectedIndent, trimmed, leadingSpaces)
			}
		}
	}

	// Validate that directories end with "/"
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasSuffix(trimmed, "/") {
			// This should be a directory entry
			if !strings.HasPrefix(trimmed, "- ") {
				t.Errorf("Directory entry should start with '- ': %s", trimmed)
			}
		} else {
			// This should be a file entry
			if !strings.HasPrefix(trimmed, "- ") {
				t.Errorf("File entry should start with '- ': %s", trimmed)
			}
		}
	}

	// Validate that all expected entries are present
	allContent := strings.Join(lines, "\n")
	expectedEntries := []string{
		"- level1/",
		"- level1-file.txt",
		"- level2/",
		"- level2-file.txt",
		"- personas/",
		"- default.md",
		"- root-file.txt",
		"- sibling/",
		"- sibling-file.txt",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(allContent, entry) {
			t.Errorf("Expected file tree to contain '%s'", entry)
		}
	}
}

func TestGetProjectOverview(t *testing.T) {
	t.Run("CLAUDE.md exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		claudeContent := "This is the CLAUDE.md file."
		err := os.WriteFile(filepath.Join(tmpDir, "CLAUDE.md"), []byte(claudeContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write CLAUDE.md: %v", err)
		}

		// Also create GEMINI.md and README.md to test priority
		err = os.WriteFile(filepath.Join(tmpDir, "GEMINI.md"), []byte("This is GEMINI.md"), 0644)
		if err != nil {
			t.Fatalf("Failed to write GEMINI.md: %v", err)
		}
		err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("This is README.md"), 0644)
		if err != nil {
			t.Fatalf("Failed to write README.md: %v", err)
		}

		content, err := getProjectOverview(tmpDir)
		if err != nil {
			t.Fatalf("getProjectOverview returned error: %v", err)
		}

		if content != claudeContent {
			t.Errorf("Expected content '%s', got '%s'", claudeContent, content)
		}
	})

	t.Run("GEMINI.md exists (no CLAUDE.md)", func(t *testing.T) {
		tmpDir := t.TempDir()

		geminiContent := "This is the GEMINI.md file."
		err := os.WriteFile(filepath.Join(tmpDir, "GEMINI.md"), []byte(geminiContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write GEMINI.md: %v", err)
		}

		// Also create README.md to test priority
		err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("This is README.md"), 0644)
		if err != nil {
			t.Fatalf("Failed to write README.md: %v", err)
		}

		content, err := getProjectOverview(tmpDir)
		if err != nil {
			t.Fatalf("getProjectOverview returned error: %v", err)
		}

		if content != geminiContent {
			t.Errorf("Expected content '%s', got '%s'", geminiContent, content)
		}
	})

	t.Run("README.md exists (no CLAUDE.md or GEMINI.md)", func(t *testing.T) {
		tmpDir := t.TempDir()

		readmeContent := "This is the README.md file."
		err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readmeContent), 0644)
		if err != nil {
			t.Fatalf("Failed to write README.md: %v", err)
		}

		content, err := getProjectOverview(tmpDir)
		if err != nil {
			t.Fatalf("getProjectOverview returned error: %v", err)
		}

		if content != readmeContent {
			t.Errorf("Expected content '%s', got '%s'", readmeContent, content)
		}
	})

	t.Run("no overview files exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		content, err := getProjectOverview(tmpDir)
		if err != nil {
			t.Fatalf("getProjectOverview returned error: %v", err)
		}

		if content != "" {
			t.Errorf("Expected empty content, got '%s'", content)
		}
	})

	t.Run("overview file exists but is unreadable", func(t *testing.T) {
		tmpDir := t.TempDir()

		claudeFile := filepath.Join(tmpDir, "CLAUDE.md")
		err := os.WriteFile(claudeFile, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write CLAUDE.md: %v", err)
		}

		// Remove read permissions
		err = os.Chmod(claudeFile, 0000)
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}
		defer os.Chmod(claudeFile, 0644) // Restore for cleanup

		_, err = getProjectOverview(tmpDir)
		if err == nil {
			t.Error("Expected error for unreadable overview file, got nil")
		}
	})
}

func TestFileTreeRespectsGitignore(t *testing.T) {
	tmpDir := t.TempDir()

	// Create personas directory with default.md
	personasDir := filepath.Join(tmpDir, "personas")
	err := os.Mkdir(personasDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create personas dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(personasDir, "default.md"), []byte("You are a test assistant."), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy system prompt: %v", err)
	}

	// Create a .gitignore file with common patterns
	gitignoreContent := `# Test gitignore patterns
*.log
*.tmp
*.test
build/
dist/
node_modules/
.DS_Store
debug_*.go
test_*.go
`
	err = os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	// Create files that should be IGNORED according to .gitignore
	ignoredFiles := []struct {
		path    string
		content string
	}{
		{"error.log", "error log content"},
		{"temp.tmp", "temporary content"},
		{"app.test", "test binary"},
		{".DS_Store", "mac metadata"},
		{"debug_main.go", "debug file"},
		{"test_helper.go", "test helper"},
	}

	for _, file := range ignoredFiles {
		err = os.WriteFile(filepath.Join(tmpDir, file.path), []byte(file.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create ignored file %s: %v", file.path, err)
		}
	}

	// Create directories that should be IGNORED
	ignoredDirs := []string{"build", "dist", "node_modules"}
	for _, dir := range ignoredDirs {
		dirPath := filepath.Join(tmpDir, dir)
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create ignored directory %s: %v", dir, err)
		}

		// Add files inside ignored directories
		err = os.WriteFile(filepath.Join(dirPath, "content.txt"), []byte("should be ignored"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file in ignored directory %s: %v", dir, err)
		}
	}

	// Create files that should NOT be ignored
	allowedFiles := []struct {
		path    string
		content string
	}{
		{"main.go", "package main"},
		{"README.md", "project readme"},
		{"config.json", "configuration"},
		{"src/app.go", "source code"},
	}

	// Create src directory for nested file
	err = os.Mkdir(filepath.Join(tmpDir, "src"), 0755)
	if err != nil {
		t.Fatalf("Failed to create src directory: %v", err)
	}

	for _, file := range allowedFiles {
		err = os.WriteFile(filepath.Join(tmpDir, file.path), []byte(file.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create allowed file %s: %v", file.path, err)
		}
	}

	// Change to temp directory for running Build
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}
	defer os.Chdir(originalWd)

	// Build the prompt with no selected files (we just want to test the file tree)
	selectedFiles := map[string]bool{}
	userPrompt := "Test gitignore filtering"

	xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt, []string{"default"})
	if err != nil {
		t.Fatalf("Build() returned an unexpected error: %v", err)
	}

	// Extract the filetree content for analysis
	start := strings.Index(xmlOutput, "<filetree><![CDATA[")
	end := strings.Index(xmlOutput, "]]></filetree>")
	if start == -1 || end == -1 {
		t.Fatal("Could not find filetree CDATA section in XML")
	}

	filetreeContent := xmlOutput[start+len("<filetree><![CDATA[") : end]

	// Verify that IGNORED files are NOT in the file tree
	ignoredItems := []string{
		"error.log",
		"temp.tmp",
		"app.test",
		".DS_Store",
		"debug_main.go",
		"test_helper.go",
		"build/",
		"dist/",
		"node_modules/",
		"content.txt", // files inside ignored directories
	}

	for _, item := range ignoredItems {
		if strings.Contains(filetreeContent, item) {
			t.Errorf("File tree should NOT contain ignored item '%s', but it was found in:\n%s", item, filetreeContent)
		}
	}

	// Verify that ALLOWED files ARE in the file tree
	allowedItems := []string{
		"main.go",
		"README.md",
		"config.json",
		"src/",
		"app.go",
		"personas/",
		"default.md",
		".gitignore", // The .gitignore file itself should be in the tree
	}

	for _, item := range allowedItems {
		if !strings.Contains(filetreeContent, item) {
			t.Errorf("File tree should contain allowed item '%s', but it was not found in:\n%s", item, filetreeContent)
		}
	}

	// Validate that the XML is well-formed
	var prompt Prompt
	err = xml.Unmarshal([]byte(xmlOutput), &prompt)
	if err != nil {
		t.Errorf("Generated XML is not well-formed: %v\nXML:\n%s", err, xmlOutput)
	}
}
