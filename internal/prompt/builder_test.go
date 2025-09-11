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
		filepath.Join(tmpDir, "testfile1.txt"): true,
		filepath.Join(tmpDir, "subdir", "testfile2.txt"): false, // This one is not selected
	}
	userPrompt := "This is a test user prompt."

	// 4. Call the Build function
	// We pass tmpDir as the root path
	xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt)

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


	// Check for system prompt
	expectedSystemPrompt := `<SystemPrompt><![CDATA[You are a test assistant.]]></SystemPrompt>`
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
	if prompt.SystemPrompt.Text != "You are a test assistant." {
		t.Errorf("Expected system prompt 'You are a test assistant.', got '%s'", prompt.SystemPrompt.Text)
	}
	if prompt.UserPrompt.Text != "This is a test user prompt." {
		t.Errorf("Expected user prompt 'This is a test user prompt.', got '%s'", prompt.UserPrompt.Text)
	}
}

func TestBuildErrorConditions(t *testing.T) {
	t.Run("invalid root path", func(t *testing.T) {
		selectedFiles := map[string]bool{}
		userPrompt := "test prompt"
		
		_, err := Build("/nonexistent/path", selectedFiles, userPrompt)
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
		
		_, err = Build(tmpDir, selectedFiles, userPrompt)
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
		
		_, err = Build(tmpDir, selectedFiles, userPrompt)
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
		
		xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt)
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
		
		xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt)
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
		
		xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt)
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

	xmlOutput, err := Build(tmpDir, selectedFiles, userPrompt)
	if err != nil {
		t.Fatalf("Build() returned an unexpected error: %v", err)
	}

	// Extract the filetree content
	start := strings.Index(xmlOutput, "<filetree><![CDATA[")
	end := strings.Index(xmlOutput, "]]></filetree>")
	if start == -1 || end == -1 {
		t.Fatal("Could not find filetree CDATA section in XML")
	}
	
	filetreeContent := xmlOutput[start+len("<filetree><![CDATA["):end]
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
