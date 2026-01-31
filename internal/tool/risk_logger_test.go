package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aiq/aiq/internal/config"
)

// TestRiskLogger_Functionality tests risk logger functionality
func TestRiskLogger_Functionality(t *testing.T) {
	t.Run("logs risk assessment decisions to file", func(t *testing.T) {
		// Reset logger to ensure clean state
		riskLogger = nil

		// Get actual config directory (will use real ~/.aiq)
		baseDir, err := config.GetBaseConfigDir()
		if err != nil {
			t.Fatalf("GetBaseConfigDir() failed: %v", err)
		}

		// Log a risk assessment
		LogRiskAssessment("Test: SQL operation assessed as low risk")

		// Verify log file was created
		logPath := filepath.Join(baseDir, "logs", "risk_assessment.log")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			t.Fatal("Expected log file to be created, but it doesn't exist")
		}

		// Read log file (may have other entries, that's fine)
		logContent, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		// Verify log content contains our message
		if !strings.Contains(string(logContent), "Test: SQL operation assessed as low risk") {
			t.Errorf("Expected log to contain test message")
		}

		// Verify log contains timestamp format
		if !strings.Contains(string(logContent), "[") {
			t.Error("Expected log to contain timestamp")
		}
	})

	t.Run("handles multiple log entries", func(t *testing.T) {
		// Reset logger
		riskLogger = nil

		// Log multiple entries
		LogRiskAssessment("Test Entry 1")
		LogRiskAssessment("Test Entry 2")
		LogRiskAssessment("Test Entry 3")

		// Read log file
		baseDir, err := config.GetBaseConfigDir()
		if err != nil {
			t.Fatalf("GetBaseConfigDir() failed: %v", err)
		}

		logPath := filepath.Join(baseDir, "logs", "risk_assessment.log")
		logContent, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		content := string(logContent)
		if !strings.Contains(content, "Test Entry 1") {
			t.Error("Expected log to contain Test Entry 1")
		}
		if !strings.Contains(content, "Test Entry 2") {
			t.Error("Expected log to contain Test Entry 2")
		}
		if !strings.Contains(content, "Test Entry 3") {
			t.Error("Expected log to contain Test Entry 3")
		}
	})

	t.Run("handles logging initialization", func(t *testing.T) {
		// Reset logger
		riskLogger = nil

		// Should not panic on first call
		LogRiskAssessment("Test initialization")
		// If we get here without panic, test passes
	})
}
