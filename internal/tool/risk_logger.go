package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aiq/aiq/internal/config"
)

var riskLogger *riskLoggerInstance

type riskLoggerInstance struct {
	logFile *os.File
}

// initRiskLogger initializes the risk assessment logger
func initRiskLogger() error {
	if riskLogger != nil {
		return nil
	}

	// Get log directory (~/.aiq/logs)
	baseDir, err := config.GetBaseConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get base config dir: %w", err)
	}

	logDir := filepath.Join(baseDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file (append mode)
	logFile, err := os.OpenFile(
		filepath.Join(logDir, "risk_assessment.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	riskLogger = &riskLoggerInstance{
		logFile: logFile,
	}

	return nil
}

// LogRiskAssessment writes risk assessment information to log file
// Logs are written to ~/.aiq/logs/risk_assessment.log
func LogRiskAssessment(format string, args ...interface{}) {
	if riskLogger == nil {
		if err := initRiskLogger(); err != nil {
			// Silently fail if logging can't be initialized
			return
		}
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)

	riskLogger.logFile.WriteString(logLine)
	riskLogger.logFile.Sync()
}
