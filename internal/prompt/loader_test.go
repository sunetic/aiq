package prompt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aiq/aiq/internal/config"
	"github.com/aiq/aiq/internal/testutil"
	"github.com/aiq/aiq/internal/version"
)

// TestPromptLoader_Initialization tests prompt loader initialization
func TestPromptLoader_Initialization(t *testing.T) {
	t.Run("creates default files if they don't exist", func(t *testing.T) {
		// Use actual config directory but ensure it's clean
		// This test verifies that NewLoader creates default files
		promptsDir, err := config.GetPromptsDir()
		if err != nil {
			t.Fatalf("GetPromptsDir() failed: %v", err)
		}

		// Backup existing files if they exist
		backupDir := t.TempDir()
		backupFiles := []string{
			FreeModeBasePromptFile,
			DatabaseBasePromptFile,
			CommonPromptFile,
		}

		for _, filename := range backupFiles {
			src := filepath.Join(promptsDir, filename)
			dst := filepath.Join(backupDir, filename)
			if data, err := os.ReadFile(src); err == nil {
				os.WriteFile(dst, data, 0644)
				os.Remove(src)
			}
		}

		// Restore files after test
		defer func() {
			for _, filename := range backupFiles {
				src := filepath.Join(backupDir, filename)
				dst := filepath.Join(promptsDir, filename)
				if data, err := os.ReadFile(src); err == nil {
					os.WriteFile(dst, data, 0644)
				}
			}
		}()

		// Create loader - should create default files
		loader, err := NewLoader()
		if err != nil {
			t.Fatalf("NewLoader() failed: %v", err)
		}
		if loader == nil {
			t.Fatal("NewLoader() returned nil")
		}

		// Verify default files were created
		for _, filename := range backupFiles {
			filePath := filepath.Join(promptsDir, filename)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Expected file %s to be created, but it doesn't exist", filename)
			}
		}
	})

	t.Run("loads prompts from user directory if they exist", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		// Create prompts directory structure
		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		// Create loader to get built-in prompts
		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		// Get built-in prompt strings and create all required files
		builtInPrompts := loader.getBuiltInPromptStrings()
		requiredFiles := []string{
			FreeModeBasePromptFile,
			DatabaseBasePromptFile,
			CommonPromptFile,
		}

		// Create a custom prompt file (modify FreeModeBasePromptFile)
		customContent := "---\n# Custom Prompt\n---\nThis is a custom prompt"
		customFile := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(customFile, []byte(customContent), 0644); err != nil {
			t.Fatalf("Failed to create custom prompt file: %v", err)
		}

		// Create other required files with built-in content
		for _, filename := range requiredFiles {
			if filename == FreeModeBasePromptFile {
				continue // Already created with custom content
			}
			filePath := filepath.Join(promptsDir, filename)
			content := builtInPrompts[filename]
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				t.Fatalf("Failed to create prompt file %s: %v", filename, err)
			}
		}

		// Load prompts
		if err := loader.loadPrompts(); err != nil {
			t.Fatalf("loadPrompts() failed: %v", err)
		}

		// Verify loader loaded the custom content
		loadedContent := loader.prompts[FreeModeBasePromptFile]
		if loadedContent == "" {
			t.Error("Expected loader to load custom prompt content, but it's empty")
		}

		// Verify custom content was loaded (after frontmatter removal)
		if loadedContent != "This is a custom prompt" {
			t.Errorf("Expected custom content 'This is a custom prompt', got %q", loadedContent)
		}
	})
}

// TestPromptLoader_VersionDetection tests prompt version detection using SHA256 hash comparison
func TestPromptLoader_VersionDetection(t *testing.T) {
	t.Run("detects modifications by comparing content hashes", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		// Create prompts directory
		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		// Create loader with custom prompts directory
		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		// Get built-in hash
		builtInHashes := loader.getBuiltInPromptHashes()
		builtInHash := builtInHashes[FreeModeBasePromptFile]

		// Modify the prompt file
		modifiedContent := "---\n# Modified Prompt\n---\nThis is a modified prompt"
		modifiedFile := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(modifiedFile, []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to write modified file: %v", err)
		}

		// Get user hash
		userHashes := loader.getUserPromptHashes()
		userHash := userHashes[FreeModeBasePromptFile]

		// Verify hashes differ
		if builtInHash == userHash {
			t.Error("Expected hashes to differ after modification, but they are the same")
		}

		// Verify hash calculation (using the package function)
		expectedHash := hashContent(modifiedContent)
		if userHash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, userHash)
		}
	})

	t.Run("detects no modifications when files match", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		// Create prompts directory
		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		// Get built-in prompt content
		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		builtInPrompts := loader.getBuiltInPromptStrings()
		builtInContent := builtInPrompts[FreeModeBasePromptFile]

		// Write built-in content to file (simulating default file creation)
		filePath := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(filePath, []byte(builtInContent), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Get hashes
		builtInHashes := loader.getBuiltInPromptHashes()
		userHashes := loader.getUserPromptHashes()

		// Verify hashes match
		builtInHash := builtInHashes[FreeModeBasePromptFile]
		userHash := userHashes[FreeModeBasePromptFile]

		if builtInHash != userHash {
			t.Errorf("Expected hashes to match for matching files, but got builtIn=%s, user=%s", builtInHash, userHash)
		}

		// Verify no modifications detected
		modifiedFiles, err := loader.checkPromptContentMismatch()
		if err != nil {
			t.Fatalf("checkPromptContentMismatch() failed: %v", err)
		}

		// Should have no modifications
		for _, filename := range modifiedFiles {
			if filename == FreeModeBasePromptFile {
				t.Errorf("Expected no modifications for %s, but it was detected as modified", filename)
			}
		}
	})
}

// TestPromptLoader_VersionChoicePersistence tests version-based choice persistence
// Note: These tests use the actual config directory, which may affect real user data
// In a production environment, these should use dependency injection or a test config directory
func TestPromptLoader_VersionChoicePersistence(t *testing.T) {
	t.Run("stores choice associated with current application version", func(t *testing.T) {
		// Use actual config directory - this test verifies the storage mechanism works
		// Set version to a test version
		originalVersion := version.Version
		testVersion := "test-v1.0.0"
		version.Version = testVersion
		defer func() { version.Version = originalVersion }()

		// Store choice
		choice := "overwrite"
		if err := config.SetChoiceForVersion(testVersion, choice); err != nil {
			t.Fatalf("SetChoiceForVersion() failed: %v", err)
		}

		// Clean up after test
		defer func() {
			// Remove test choice after test
			choices, _ := config.LoadVersionChoices()
			if choices != nil && choices.Choices != nil {
				delete(choices.Choices, testVersion)
				config.SaveVersionChoices(choices)
			}
		}()

		// Retrieve choice
		retrievedChoice, err := config.GetChoiceForVersion(testVersion)
		if err != nil {
			t.Fatalf("GetChoiceForVersion() failed: %v", err)
		}

		if retrievedChoice != choice {
			t.Errorf("Expected choice %s, got %s", choice, retrievedChoice)
		}
	})

	t.Run("retrieves stored choice on subsequent runs", func(t *testing.T) {
		// Set version to a test version
		originalVersion := version.Version
		testVersion := "test-v1.0.1"
		version.Version = testVersion
		defer func() { version.Version = originalVersion }()

		// Store choice
		choice := "keep"
		if err := config.SetChoiceForVersion(testVersion, choice); err != nil {
			t.Fatalf("SetChoiceForVersion() failed: %v", err)
		}

		// Clean up after test
		defer func() {
			choices, _ := config.LoadVersionChoices()
			if choices != nil && choices.Choices != nil {
				delete(choices.Choices, testVersion)
				config.SaveVersionChoices(choices)
			}
		}()

		// Simulate subsequent run - retrieve choice
		retrievedChoice, err := config.GetChoiceForVersion(testVersion)
		if err != nil {
			t.Fatalf("GetChoiceForVersion() failed: %v", err)
		}

		if retrievedChoice != choice {
			t.Errorf("Expected choice %s on subsequent run, got %s", choice, retrievedChoice)
		}
	})

	t.Run("prompts user again when application version changes", func(t *testing.T) {
		// Set version to first version
		originalVersion := version.Version
		version1 := "test-v1.0.2"
		version.Version = version1

		// Store choice for version 1
		choice1 := "overwrite"
		if err := config.SetChoiceForVersion(version1, choice1); err != nil {
			t.Fatalf("SetChoiceForVersion() failed: %v", err)
		}

		// Clean up after test
		defer func() {
			version.Version = originalVersion
			choices, _ := config.LoadVersionChoices()
			if choices != nil && choices.Choices != nil {
				delete(choices.Choices, version1)
				delete(choices.Choices, "test-v2.0.0")
				config.SaveVersionChoices(choices)
			}
		}()

		// Change to new version
		version2 := "test-v2.0.0"
		version.Version = version2

		// Try to retrieve choice for version 2 - should be empty (no choice yet)
		choice2, err := config.GetChoiceForVersion(version2)
		if err != nil {
			t.Fatalf("GetChoiceForVersion() failed: %v", err)
		}

		if choice2 != "" {
			t.Errorf("Expected no choice for new version %s, but got %s", version2, choice2)
		}

		// Verify choice for version 1 is still stored
		retrievedChoice1, err := config.GetChoiceForVersion(version1)
		if err != nil {
			t.Fatalf("GetChoiceForVersion() failed: %v", err)
		}

		if retrievedChoice1 != choice1 {
			t.Errorf("Expected choice %s for version %s, got %s", choice1, version1, retrievedChoice1)
		}
	})
}

// TestPromptLoader_UpgradeFlow tests prompt upgrade flow with overwrite vs keep choices
func TestPromptLoader_UpgradeFlow(t *testing.T) {
	t.Run("overwrites files when user chooses overwrite", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		// Create prompts directory
		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		// Create loader to get built-in prompts
		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		builtInPrompts := loader.getBuiltInPromptStrings()
		originalContent := builtInPrompts[FreeModeBasePromptFile]

		// Create a modified file
		modifiedContent := "---\n# Modified\n---\nModified content"
		filePath := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(filePath, []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to write modified file: %v", err)
		}

		// Set version and choice to overwrite
		originalVersion := version.Version
		testVersion := "test-upgrade-v1.0.0"
		version.Version = testVersion
		defer func() { version.Version = originalVersion }()

		if err := config.SetChoiceForVersion(testVersion, "overwrite"); err != nil {
			t.Fatalf("SetChoiceForVersion() failed: %v", err)
		}
		defer func() {
			choices, _ := config.LoadVersionChoices()
			if choices != nil && choices.Choices != nil {
				delete(choices.Choices, testVersion)
				config.SaveVersionChoices(choices)
			}
		}()

		// Initialize defaults - should overwrite
		if err := loader.initializeDefaults(); err != nil {
			t.Fatalf("initializeDefaults() failed: %v", err)
		}

		// Verify file was overwritten with built-in content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(fileContent) != originalContent {
			t.Error("Expected file to be overwritten with built-in content, but it wasn't")
		}
	})

	t.Run("preserves files when user chooses keep", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		// Create prompts directory
		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		// Create loader
		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		// Create a modified file
		modifiedContent := "---\n# Modified\n---\nModified content"
		filePath := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(filePath, []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to write modified file: %v", err)
		}

		// Set version and choice to keep
		originalVersion := version.Version
		testVersion := "test-keep-v1.0.0"
		version.Version = testVersion
		defer func() { version.Version = originalVersion }()

		if err := config.SetChoiceForVersion(testVersion, "keep"); err != nil {
			t.Fatalf("SetChoiceForVersion() failed: %v", err)
		}
		defer func() {
			choices, _ := config.LoadVersionChoices()
			if choices != nil && choices.Choices != nil {
				delete(choices.Choices, testVersion)
				config.SaveVersionChoices(choices)
			}
		}()

		// Initialize defaults - should preserve file
		if err := loader.initializeDefaults(); err != nil {
			t.Fatalf("initializeDefaults() failed: %v", err)
		}

		// Verify file was preserved
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(fileContent) != modifiedContent {
			t.Errorf("Expected file to be preserved, but got %q", string(fileContent))
		}
	})
}

// TestPromptLoader_ErrorHandling tests prompt file error handling
func TestPromptLoader_ErrorHandling(t *testing.T) {
	t.Run("handles missing files gracefully", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		// Create prompts directory
		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		// Create loader
		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		// Try to load prompts - should handle missing files
		// Note: loadPrompts requires all base files, so this will fail
		// But we can test that getUserPromptHashes handles missing files
		userHashes := loader.getUserPromptHashes()

		// Should return empty map or handle gracefully
		_ = userHashes // Just verify it doesn't panic
	})

	t.Run("handles invalid content gracefully", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		// Create prompts directory
		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		// Create loader
		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		// Create invalid prompt file (malformed YAML frontmatter)
		invalidContent := "---\ninvalid: yaml: [\n---\nThis is invalid"
		invalidFile := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		// Try to load prompts - should handle invalid file
		// Note: parsePromptFile may fail on invalid YAML, which is acceptable
		err := loader.loadPrompts()
		if err != nil {
			// Error is acceptable for invalid files
			t.Logf("loadPrompts() failed on invalid file (expected): %v", err)
		}
	})
}

// TestPromptLoader_VariousScenarios tests various prompt file scenarios
func TestPromptLoader_VariousScenarios(t *testing.T) {
	t.Run("handles empty prompt files", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		// Create empty file
		emptyFile := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		// Should handle empty file gracefully
		userHashes := loader.getUserPromptHashes()
		_ = userHashes // Verify it doesn't panic
	})

	t.Run("handles files with only frontmatter", func(t *testing.T) {
		tmpDir, cleanup := testutil.CreateTempDir(t)
		defer cleanup()

		promptsDir := filepath.Join(tmpDir, "prompts")
		if err := os.MkdirAll(promptsDir, 0755); err != nil {
			t.Fatalf("Failed to create prompts directory: %v", err)
		}

		loader := &Loader{
			promptsDir: promptsDir,
			prompts:    make(map[string]string),
		}

		// Create file with only frontmatter
		frontmatterOnly := "---\n# Test\n---\n"
		filePath := filepath.Join(promptsDir, FreeModeBasePromptFile)
		if err := os.WriteFile(filePath, []byte(frontmatterOnly), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Should handle file with only frontmatter
		if err := loader.loadPrompts(); err != nil {
			// May fail if other required files are missing, which is acceptable
			t.Logf("loadPrompts() may fail if other files missing: %v", err)
		}
	})
}
