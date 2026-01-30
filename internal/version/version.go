package version

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Version and CommitID are set at build time via -ldflags
// Example: go build -ldflags "-X github.com/aiq/aiq/internal/version.Version=v1.0.0"
//
// For testing: You can override these at build time:
//
//	go build -ldflags "-X github.com/aiq/aiq/internal/version.Version=v1.0.0 -X github.com/aiq/aiq/internal/version.CommitID=abc1234"
//
// Or test version detection by modifying prompt file versions in ~/.aiq/prompts/
var (
	Version  = "dev"
	CommitID = "unknown"
)

const (
	// gitCommandTimeout is the timeout for git commands
	gitCommandTimeout = 2 * time.Second
)

// GetVersion returns the application version using three-tier fallback:
// 1. Build-time injected Version variable (via -ldflags)
// 2. Git describe command (if git available)
// 3. Default value "dev"
func GetVersion() string {
	// If version was injected at build time, use it
	if Version != "dev" {
		return Version
	}

	// Try to get version from git
	if v := getGitVersion(); v != "" {
		return v
	}

	// Fallback to default
	return "dev"
}

// GetCommitID returns the commit ID using three-tier fallback:
// 1. Build-time injected CommitID variable (via -ldflags)
// 2. Git rev-parse command (if git available)
// 3. Default value "unknown"
func GetCommitID() string {
	// If commit ID was injected at build time, use it
	if CommitID != "unknown" {
		return CommitID
	}

	// Try to get commit ID from git
	if c := getGitCommitID(); c != "" {
		return c
	}

	// Fallback to default
	return "unknown"
}

// GetVersionInfo returns a formatted version string: "aiq <version> (commit: <commit-id>)"
func GetVersionInfo() string {
	version := GetVersion()
	commitID := GetCommitID()
	return fmt.Sprintf("aiq %s (commit: %s)", version, commitID)
}

// getGitVersion attempts to get version from git describe command
// Returns empty string if git is unavailable or command fails
func getGitVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--always")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(output))
	if version == "" {
		return ""
	}

	return version
}

// getGitCommitID attempts to get commit ID from git rev-parse command
// Returns empty string if git is unavailable or command fails
func getGitCommitID() string {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	commitID := strings.TrimSpace(string(output))
	if commitID == "" {
		return ""
	}

	// Return short commit ID (first 7 characters) for readability
	if len(commitID) > 7 {
		return commitID[:7]
	}

	return commitID
}
