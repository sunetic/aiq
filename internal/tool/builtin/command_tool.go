package builtin

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aiq/aiq/internal/ui"
)

// CommandTool handles command execution
type CommandTool struct {
	blockedCommands     map[string]bool // Blocklist of dangerous commands
	interactiveCommands map[string]bool // Commands that require interactive input
}

// NewCommandTool creates a new command tool with security blocklist
// By default, all commands are allowed except those in the blocklist
func NewCommandTool() *CommandTool {
	// Default blocklist: dangerous commands that should be blocked
	blocked := map[string]bool{
		"rm":       true, // Remove files/directories - dangerous
		"sudo":     true, // Execute as superuser - dangerous
		"dd":       true, // Disk dump - can destroy data
		"mkfs":     true, // Make filesystem - can destroy data
		"fdisk":    true, // Disk partitioning - can destroy data
		"shutdown": true, // Shutdown system - dangerous
		"reboot":   true, // Reboot system - dangerous
		"halt":     true, // Halt system - dangerous
		"poweroff": true, // Power off system - dangerous
		"init":     true, // Init system - dangerous
		"killall":  true, // Kill all processes - dangerous
		"kill":     true, // Kill processes - dangerous (especially kill -9)
	}

	// Commands that require interactive input (cannot be executed non-interactively)
	interactive := map[string]bool{
		"mysql_secure_installation": true,
		"passwd":                    true,
		"ssh":                       true, // Without -o BatchMode=yes
		"ftp":                       true,
		"telnet":                    true,
		"less":                      true,
		"more":                      true,
		"vi":                        true,
		"vim":                       true,
		"nano":                      true,
		"emacs":                     true,
	}

	return &CommandTool{
		blockedCommands:     blocked,
		interactiveCommands: interactive,
	}
}

// SetBlockedCommands sets the blocklist of blocked commands
func (t *CommandTool) SetBlockedCommands(commands []string) {
	t.blockedCommands = make(map[string]bool)
	for _, cmd := range commands {
		t.blockedCommands[cmd] = true
	}
}

// CommandParams represents parameters for command execution
type CommandParams struct {
	Command    string   `json:"command"`
	Args       []string `json:"args,omitempty"`
	WorkingDir string   `json:"working_dir,omitempty"`
	Timeout    int      `json:"timeout,omitempty"` // Timeout in seconds, default 60
}

// OutputCallback is called when command produces output (for real-time display)
type OutputCallback func(line string)

// CommandResult represents command execution result
type CommandResult struct {
	Stdout          string `json:"stdout"`           // Full stdout output
	Stderr          string `json:"stderr"`           // Full stderr output
	TruncatedStdout string `json:"truncated_stdout"` // Truncated stdout for LLM (last N lines)
	TruncatedStderr string `json:"truncated_stderr"` // Truncated stderr for LLM (last N lines)
	ExitCode        int    `json:"exit_code"`
}

// truncateOutput truncates output to last N lines
func truncateOutput(output string, maxLines int) string {
	if maxLines <= 0 {
		return output
	}

	lines := strings.Split(output, "\n")
	if len(lines) <= maxLines {
		return output
	}

	// Return last maxLines lines
	start := len(lines) - maxLines
	return strings.Join(lines[start:], "\n")
}

// truncateOutputBySize limits output to maximum size (10MB)
func truncateOutputBySize(output string, maxSize int) string {
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // 10MB default
	}

	if len(output) <= maxSize {
		return output
	}

	// If output exceeds maxSize, keep only the last portion
	// Try to keep whole lines
	truncated := output[len(output)-maxSize:]
	// Find first newline after truncation point
	if idx := strings.Index(truncated, "\n"); idx >= 0 {
		truncated = truncated[idx+1:]
	}
	return truncated
}

// Execute executes a shell command with streaming output and truncation
func (t *CommandTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	return t.ExecuteWithCallback(ctx, params, nil)
}

// ExecuteWithCallback executes a shell command with optional output callback for real-time display
func (t *CommandTool) ExecuteWithCallback(ctx context.Context, params map[string]interface{}, callback OutputCallback) (interface{}, error) {
	// Parse parameters
	var cmdParams CommandParams
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	if err := json.Unmarshal(paramsJSON, &cmdParams); err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}

	// Validate command
	if cmdParams.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Check if command requires sudo (commands with sudo need user interaction for password)
	if strings.Contains(cmdParams.Command, "sudo ") || strings.HasPrefix(cmdParams.Command, "sudo ") {
		return nil, fmt.Errorf("command requires sudo privileges and cannot be executed automatically. Please run this command manually in your terminal: %s", cmdParams.Command)
	}

	// Extract environment variables from command (e.g., "LC_ALL=C command")
	var envVars []string
	parts := strings.Fields(cmdParams.Command)
	for _, part := range parts {
		if strings.Contains(part, "=") && !strings.HasPrefix(part, "-") {
			envVars = append(envVars, part)
		} else {
			break
		}
	}

	// Check if first command (after env vars) is blocked
	if len(parts) > len(envVars) {
		firstCmd := parts[len(envVars)]
		if t.blockedCommands[firstCmd] {
			return nil, fmt.Errorf("command '%s' is blocked for security reasons. Blocked commands: %v", firstCmd, t.getBlockedCommandList())
		}
		if t.interactiveCommands[firstCmd] {
			return nil, fmt.Errorf("command '%s' requires interactive input and cannot be executed non-interactively. Please run this command manually in your terminal, or use a non-interactive alternative if available", firstCmd)
		}
	}

	// Use shell to execute all commands - simpler and handles all shell features
	// Detect shell: prefer sh, fallback to bash if sh not available
	shell := "/bin/sh"
	if _, err := exec.LookPath("sh"); err != nil {
		shell = "/bin/bash"
	}
	cmdName := shell
	cmdArgs := []string{"-c", cmdParams.Command}

	// Set idle timeout (timeout resets when there's output)
	idleTimeout := 60 * time.Second
	if cmdParams.Timeout > 0 {
		idleTimeout = time.Duration(cmdParams.Timeout) * time.Second
	}

	// Create command (no fixed context timeout - we use idle timeout instead)
	cmd := exec.Command(cmdName, cmdArgs...)

	// Set environment variables if any were specified
	if len(envVars) > 0 {
		// Start with current environment
		cmd.Env = os.Environ()
		// Add or override with specified environment variables
		for _, envVar := range envVars {
			cmd.Env = append(cmd.Env, envVar)
		}
	}

	// Set working directory
	if cmdParams.WorkingDir != "" {
		cmd.Dir = cmdParams.WorkingDir
	}

	// Use separate pipes for stdout and stderr to enable streaming
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Buffers to collect output
	var stdoutBuilder, stderrBuilder strings.Builder
	maxOutputSize := 10 * 1024 * 1024 // 10MB limit

	// Channels for coordination
	done := make(chan error, 1)
	activity := make(chan struct{}, 100) // Buffered channel for activity signals

	// Sync for goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout in goroutine
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			// Signal activity (non-blocking)
			select {
			case activity <- struct{}{}:
			default:
			}
			// Call callback for real-time display
			if callback != nil {
				callback(line)
			}
			// Store in buffer
			if stdoutBuilder.Len()+len(line)+1 <= maxOutputSize {
				stdoutBuilder.WriteString(line + "\n")
			}
		}
	}()

	// Read stderr in goroutine
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			// Signal activity (non-blocking)
			select {
			case activity <- struct{}{}:
			default:
			}
			// Call callback for real-time display
			if callback != nil {
				callback(line)
			}
			// Store in buffer
			if stderrBuilder.Len()+len(line)+1 <= maxOutputSize {
				stderrBuilder.WriteString(line + "\n")
			}
		}
	}()

	// Wait for command to complete
	go func() {
		wg.Wait() // Wait for readers to finish
		done <- cmd.Wait()
	}()

	// Wait for command completion with idle timeout
	// Timer resets whenever there's output activity
	idleTimer := time.NewTimer(idleTimeout)
	defer idleTimer.Stop()

	for {
		select {
		case err := <-done:
			// Command completed
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					return nil, fmt.Errorf("command execution failed: %w", err)
				}
			}

			// Get full output
			stdout := stdoutBuilder.String()
			stderr := stderrBuilder.String()

			// Apply size limit if exceeded
			stdout = truncateOutputBySize(stdout, maxOutputSize)
			stderr = truncateOutputBySize(stderr, maxOutputSize)

			// Truncate for LLM based on exit code
			// Success: 20 lines, Failure: 100 lines
			truncateLines := 20
			if exitCode != 0 {
				truncateLines = 100
			}

			truncatedStdout := truncateOutput(stdout, truncateLines)
			truncatedStderr := truncateOutput(stderr, truncateLines)

			result := CommandResult{
				Stdout:          stdout,
				Stderr:          stderr,
				TruncatedStdout: truncatedStdout,
				TruncatedStderr: truncatedStderr,
				ExitCode:        exitCode,
			}

			return result, nil

		case <-activity:
			// Activity detected - reset idle timer
			if !idleTimer.Stop() {
				select {
				case <-idleTimer.C:
				default:
				}
			}
			idleTimer.Reset(idleTimeout)

		case <-idleTimer.C:
			// Idle timeout - no output for too long
			// Ask user if they want to continue waiting
			continueWaiting, err := ui.ShowConfirm(
				fmt.Sprintf("Command has been idle for %v. Continue waiting?", idleTimeout))
			if err != nil {
				// User interrupted (Ctrl+C), treat as cancellation
				cmd.Process.Kill()
				return nil, fmt.Errorf("command execution cancelled by user")
			}
			if !continueWaiting {
				// User chose not to continue
				cmd.Process.Kill()
				return nil, fmt.Errorf("command execution timeout: no output for %v", idleTimeout)
			}
			// User chose to continue, reset the timer
			idleTimer.Reset(idleTimeout)

		case <-ctx.Done():
			// Parent context cancelled
			cmd.Process.Kill()
			return nil, fmt.Errorf("command execution cancelled: %w", ctx.Err())
		}
	}
}

// getBlockedCommandList returns list of blocked commands
func (t *CommandTool) getBlockedCommandList() []string {
	var list []string
	for cmd := range t.blockedCommands {
		list = append(list, cmd)
	}
	return list
}

// GetDefinition returns the tool definition for LLM
func (t *CommandTool) GetDefinition() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "execute_command",
			"description": "Execute shell commands for system operations (installation, setup, configuration). Use for system operations, NOT for database queries. Most commands are allowed, but dangerous commands (like rm, sudo, dd) are blocked for security.",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Command to execute (e.g., 'ls -la')",
					},
					"args": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Command arguments",
					},
					"working_dir": map[string]interface{}{
						"type":        "string",
						"description": "Working directory for command execution",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in seconds (default: 60)",
					},
					"risk_level": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"low", "medium", "high"},
						"description": "Optional: Risk level assessment for this operation. 'low' = safe to execute automatically (e.g., ls, cat, pwd), 'medium'/'high' = requires user confirmation (e.g., rm, sudo). If not provided, system will assess risk conservatively.",
					},
					"task_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"definitive", "exploratory"},
						"description": "Optional: Task type classification. 'definitive' = task is clear and complete, 'exploratory' = task requires information gathering or multi-step process. If not provided, system will infer from context.",
					},
					"output_mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"full", "streaming"},
						"description": "**REQUIRED**: Output display mode. Classify tool call as process-oriented or result-oriented: 'full' = result-oriented (tool output IS the final goal, e.g., 'list files'), 'streaming' = process-oriented (tool output is intermediate step, e.g., 'install mysql', 'search and find'). **MANDATORY**: Always explicitly set this parameter. For long-running commands (install, build, compile, etc.), use 'streaming' even if task_type is 'definitive'.",
					},
				},
				"required": []string{"command"},
			},
		},
	}
}
