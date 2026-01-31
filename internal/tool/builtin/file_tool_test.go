package builtin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestFileTool_ReadOperations tests file tool read operations
func TestFileTool_ReadOperations(t *testing.T) {
	t.Run("reads file content", func(t *testing.T) {
		tool, err := NewFileTool()
		if err != nil {
			t.Fatalf("NewFileTool() failed: %v", err)
		}

		// Use current working directory (allowed directory)
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() failed: %v", err)
		}

		testFile := filepath.Join(cwd, "test_read.txt")
		testContent := "test file content"
		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile) // Clean up

		params := map[string]interface{}{
			"operation": "read",
			"path":      testFile,
		}

		ctx := context.Background()
		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		fileResult, ok := result.(FileResult)
		if !ok {
			t.Fatalf("Expected FileResult, got %T", result)
		}

		if !fileResult.Success {
			t.Errorf("Expected success, got failure: %s", fileResult.Message)
		}

		if fileResult.Content != testContent {
			t.Errorf("Expected content %q, got %q", testContent, fileResult.Content)
		}
	})

	t.Run("handles file not found", func(t *testing.T) {
		tool, err := NewFileTool()
		if err != nil {
			t.Fatalf("NewFileTool() failed: %v", err)
		}

		// Use current working directory (allowed directory)
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() failed: %v", err)
		}

		params := map[string]interface{}{
			"operation": "read",
			"path":      filepath.Join(cwd, "nonexistent_file_12345.txt"),
		}

		ctx := context.Background()
		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		fileResult, ok := result.(FileResult)
		if !ok {
			t.Fatalf("Expected FileResult, got %T", result)
		}

		if fileResult.Success {
			t.Error("Expected failure for nonexistent file")
		}
	})

	t.Run("checks file existence", func(t *testing.T) {
		tool, err := NewFileTool()
		if err != nil {
			t.Fatalf("NewFileTool() failed: %v", err)
		}

		// Use current working directory (allowed directory)
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() failed: %v", err)
		}

		testFile := filepath.Join(cwd, "test_exists.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer os.Remove(testFile) // Clean up

		params := map[string]interface{}{
			"operation": "exists",
			"path":      testFile,
		}

		ctx := context.Background()
		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		fileResult, ok := result.(FileResult)
		if !ok {
			t.Fatalf("Expected FileResult, got %T", result)
		}

		if !fileResult.Success {
			t.Errorf("Expected success, got failure: %s", fileResult.Message)
		}

		if !fileResult.Exists {
			t.Error("Expected file to exist")
		}
	})

	t.Run("lists directory contents", func(t *testing.T) {
		tool, err := NewFileTool()
		if err != nil {
			t.Fatalf("NewFileTool() failed: %v", err)
		}

		// Use current working directory (allowed directory)
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() failed: %v", err)
		}

		// Create test files in current directory
		testFile1 := filepath.Join(cwd, "test_list1.txt")
		testFile2 := filepath.Join(cwd, "test_list2.txt")
		os.WriteFile(testFile1, []byte("content1"), 0644)
		os.WriteFile(testFile2, []byte("content2"), 0644)
		defer os.Remove(testFile1)
		defer os.Remove(testFile2)

		params := map[string]interface{}{
			"operation": "list",
			"path":      cwd,
		}

		ctx := context.Background()
		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		fileResult, ok := result.(FileResult)
		if !ok {
			t.Fatalf("Expected FileResult, got %T", result)
		}

		if !fileResult.Success {
			t.Errorf("Expected success, got failure: %s", fileResult.Message)
		}

		if len(fileResult.Files) == 0 {
			t.Error("Expected files to be listed")
		}
	})
}

// TestFileTool_WriteOperations tests file tool write operations
func TestFileTool_WriteOperations(t *testing.T) {
	t.Run("writes file content", func(t *testing.T) {
		tool, err := NewFileTool()
		if err != nil {
			t.Fatalf("NewFileTool() failed: %v", err)
		}

		// Use current working directory (allowed directory)
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Getwd() failed: %v", err)
		}

		testFile := filepath.Join(cwd, "test_write.txt")
		testContent := "written content"
		defer os.Remove(testFile) // Clean up

		params := map[string]interface{}{
			"operation": "write",
			"path":      testFile,
			"content":   testContent,
		}

		ctx := context.Background()
		result, err := tool.Execute(ctx, params)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		fileResult, ok := result.(FileResult)
		if !ok {
			t.Fatalf("Expected FileResult, got %T", result)
		}

		if !fileResult.Success {
			t.Errorf("Expected success, got failure: %s", fileResult.Message)
		}

		// Verify file was written
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("Expected content %q, got %q", testContent, string(content))
		}
	})
}

// TestFileTool_ErrorHandling tests file tool error handling
func TestFileTool_ErrorHandling(t *testing.T) {
	t.Run("handles permission errors", func(t *testing.T) {
		tool, err := NewFileTool()
		if err != nil {
			t.Fatalf("NewFileTool() failed: %v", err)
		}

		// Try to read from restricted path (if possible)
		// Note: This may not work on all systems, so we just verify it doesn't panic
		params := map[string]interface{}{
			"operation": "read",
			"path":      "/root/restricted_file.txt",
		}

		ctx := context.Background()
		_, err = tool.Execute(ctx, params)
		// Error is acceptable for restricted paths
		_ = err
	})

	t.Run("handles invalid operation", func(t *testing.T) {
		tool, err := NewFileTool()
		if err != nil {
			t.Fatalf("NewFileTool() failed: %v", err)
		}

		params := map[string]interface{}{
			"operation": "invalid_operation",
			"path":      "/tmp/test.txt",
		}

		ctx := context.Background()
		result, err := tool.Execute(ctx, params)
		if err != nil {
			// Error is acceptable for invalid operation
			return
		}

		fileResult, ok := result.(FileResult)
		if ok && fileResult.Success {
			t.Error("Expected failure for invalid operation")
		}
	})
}
