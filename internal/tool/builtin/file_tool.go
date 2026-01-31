package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aiq/aiq/internal/config"
)

// FileTool handles file operations
type FileTool struct {
	allowedDirs []string // Allowed directories for file operations
}

// NewFileTool creates a new file tool with path restrictions
func NewFileTool() (*FileTool, error) {
	// Get allowed directories
	configDir, err := config.GetBaseConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	return &FileTool{
		allowedDirs: []string{configDir, cwd},
	}, nil
}

// FileReadParams represents parameters for file read
type FileReadParams struct {
	Path string `json:"path"`
}

// FileWriteParams represents parameters for file write
type FileWriteParams struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// FileListParams represents parameters for directory listing
type FileListParams struct {
	Path string `json:"path"`
}

// FileExistsParams represents parameters for file existence check
type FileExistsParams struct {
	Path string `json:"path"`
}

// FileResult represents file operation result
type FileResult struct {
	Success bool     `json:"success"`
	Message string   `json:"message,omitempty"`
	Content string   `json:"content,omitempty"`
	Files   []string `json:"files,omitempty"`
	Exists  bool     `json:"exists,omitempty"`
}

// validatePath checks if a path is within allowed directories
func (t *FileTool) validatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	for _, allowedDir := range t.allowedDirs {
		allowedAbs, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}

		rel, err := filepath.Rel(allowedAbs, absPath)
		if err != nil {
			continue
		}

		// Check if path is within allowed directory (not going up with ..)
		if !strings.HasPrefix(rel, "..") {
			return nil
		}
	}

	return fmt.Errorf("path '%s' is not within allowed directories: %v", path, t.allowedDirs)
}

// ReadFile reads a file
func (t *FileTool) ReadFile(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var fileParams FileReadParams
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	if err := json.Unmarshal(paramsJSON, &fileParams); err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}

	if fileParams.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	if err := t.validatePath(fileParams.Path); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(fileParams.Path)
	if err != nil {
		return FileResult{
			Success: false,
			Message: fmt.Sprintf("failed to read file: %v", err),
		}, nil
	}

	return FileResult{
		Success: true,
		Content: string(content),
	}, nil
}

// WriteFile writes to a file
func (t *FileTool) WriteFile(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var fileParams FileWriteParams
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	if err := json.Unmarshal(paramsJSON, &fileParams); err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}

	if fileParams.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	if err := t.validatePath(fileParams.Path); err != nil {
		return nil, err
	}

	// Create directory if needed
	dir := filepath.Dir(fileParams.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return FileResult{
			Success: false,
			Message: fmt.Sprintf("failed to create directory: %v", err),
		}, nil
	}

	if err := os.WriteFile(fileParams.Path, []byte(fileParams.Content), 0644); err != nil {
		return FileResult{
			Success: false,
			Message: fmt.Sprintf("failed to write file: %v", err),
		}, nil
	}

	return FileResult{
		Success: true,
		Message: "File written successfully",
	}, nil
}

// ListDirectory lists files in a directory
func (t *FileTool) ListDirectory(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var fileParams FileListParams
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	if err := json.Unmarshal(paramsJSON, &fileParams); err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}

	path := fileParams.Path
	if path == "" {
		// Default to current working directory
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	if err := t.validatePath(path); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return FileResult{
			Success: false,
			Message: fmt.Sprintf("failed to list directory: %v", err),
		}, nil
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		files = append(files, entry.Name())
	}

	return FileResult{
		Success: true,
		Files:   files,
	}, nil
}

// FileExists checks if a file exists
func (t *FileTool) FileExists(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	var fileParams FileExistsParams
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	if err := json.Unmarshal(paramsJSON, &fileParams); err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}

	if fileParams.Path == "" {
		return nil, fmt.Errorf("path is required")
	}

	if err := t.validatePath(fileParams.Path); err != nil {
		return nil, err
	}

	_, err = os.Stat(fileParams.Path)
	exists := err == nil

	return FileResult{
		Success: true,
		Exists:  exists,
	}, nil
}

// Execute executes a file operation based on operation type
func (t *FileTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation is required (read, write, list, exists)")
	}

	switch operation {
	case "read":
		return t.ReadFile(ctx, params)
	case "write":
		return t.WriteFile(ctx, params)
	case "list":
		return t.ListDirectory(ctx, params)
	case "exists":
		return t.FileExists(ctx, params)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}

// GetDefinition returns the tool definition for LLM
func (t *FileTool) GetDefinition() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "file_operations",
			"description": "Perform file operations (read, write, list, check existence). Restricted to user config directory and current working directory.",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"read", "write", "list", "exists"},
						"description": "Operation to perform",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File or directory path",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to write (for write operation)",
					},
					"risk_level": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"low", "medium", "high"},
						"description": "Optional: Risk level assessment for this operation. 'low' = safe to execute automatically (e.g., read, list, exists), 'medium'/'high' = requires user confirmation (e.g., write). If not provided, system will assess risk conservatively.",
					},
					"task_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"definitive", "exploratory"},
						"description": "Optional: Task type classification. 'definitive' = task is clear and complete, 'exploratory' = task requires information gathering or multi-step process. If not provided, system will infer from context.",
					},
					"output_mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"full", "streaming"},
						"description": "Optional: Output display mode. 'full' = display all results to user (for definitive tasks), 'streaming' = real-time streaming with truncation (for exploratory tasks). If not provided, inferred from task_type.",
					},
				},
				"required": []string{"operation", "path"},
			},
		},
	}
}
