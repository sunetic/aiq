package tool

import (
	"regexp"
	"strings"
)

// RiskLevel represents the risk level of a tool operation
type RiskLevel int

const (
	// RiskLow indicates the operation is safe and can be executed automatically
	RiskLow RiskLevel = iota
	// RiskHigh indicates the operation is potentially dangerous and requires user confirmation
	RiskHigh
)

// RiskAssessor interface for assessing risk of tool operations
// Each tool type can implement its own risk assessor
type RiskAssessor interface {
	// AssessRisk evaluates the risk level of a tool operation
	// Priority: (1) LLM-provided risk_level, (2) Code whitelist, (3) Conservative default (require confirmation)
	AssessRisk(toolName string, args map[string]interface{}) RiskLevel
}

// extractRiskLevel extracts LLM-provided risk_level from tool arguments
// Returns the risk level string ("low", "medium", "high") and whether it was provided
func extractRiskLevel(args map[string]interface{}) (string, bool) {
	if riskLevel, ok := args["risk_level"].(string); ok {
		return strings.ToLower(riskLevel), true
	}
	return "", false
}

// assessRiskFromLLM converts LLM-provided risk_level string to RiskLevel
// "low" -> RiskLow, "medium"/"high" -> RiskHigh
func assessRiskFromLLM(riskLevelStr string) RiskLevel {
	switch strings.ToLower(riskLevelStr) {
	case "low":
		return RiskLow
	case "medium", "high":
		return RiskHigh
	default:
		// Unknown risk level -> conservative default (require confirmation)
		return RiskHigh
	}
}

// SQLRiskAssessor assesses risk for SQL operations
type SQLRiskAssessor struct{}

// NewSQLRiskAssessor creates a new SQL risk assessor
func NewSQLRiskAssessor() *SQLRiskAssessor {
	return &SQLRiskAssessor{}
}

// AssessRisk evaluates SQL operation risk
// Priority: (1) LLM-provided risk_level, (2) Code whitelist (SELECT, SHOW, DESCRIBE, EXPLAIN), (3) Require confirmation
func (r *SQLRiskAssessor) AssessRisk(toolName string, args map[string]interface{}) RiskLevel {
	// Priority 1: Check LLM-provided risk_level
	if riskLevelStr, ok := extractRiskLevel(args); ok {
		LogRiskAssessment("SQL: LLM provided risk_level=%s", riskLevelStr)
		return assessRiskFromLLM(riskLevelStr)
	}

	// Priority 2: Check code whitelist
	if sql, ok := args["sql"].(string); ok {
		if isSQLWhitelisted(sql) {
			LogRiskAssessment("SQL: Whitelisted operation (low risk), SQL starts with: %s", getSQLFirstKeyword(sql))
			return RiskLow
		}
		LogRiskAssessment("SQL: Not whitelisted, SQL starts with: %s", getSQLFirstKeyword(sql))
	}

	// Priority 3: Conservative default (require confirmation)
	LogRiskAssessment("SQL: Default to high risk (requires confirmation)")
	return RiskHigh
}

// getSQLFirstKeyword extracts the first SQL keyword for logging
func getSQLFirstKeyword(sql string) string {
	sql = strings.TrimSpace(sql)
	parts := strings.Fields(sql)
	if len(parts) > 0 {
		return strings.ToUpper(parts[0])
	}
	return "unknown"
}

// isSQLWhitelisted checks if SQL statement matches whitelist patterns (SELECT, SHOW, DESCRIBE, EXPLAIN, CREATE TABLE)
// CREATE TABLE is considered low-risk as it only creates new tables without modifying existing data
func isSQLWhitelisted(sql string) bool {
	// Normalize SQL: trim whitespace and convert to uppercase for comparison
	normalized := strings.TrimSpace(strings.ToUpper(sql))

	// Check for CREATE TABLE first (two-word pattern)
	if strings.HasPrefix(normalized, "CREATE TABLE") {
		return true
	}

	// Check for single-word patterns
	whitelistPatterns := []string{
		"SELECT",
		"SHOW",
		"DESCRIBE",
		"DESC",
		"EXPLAIN",
	}

	for _, pattern := range whitelistPatterns {
		// Use regex to match at start of statement (after optional whitespace/comments)
		// Match pattern followed by space, semicolon, or end of string
		regex := regexp.MustCompile(`(?i)^\s*` + regexp.QuoteMeta(pattern) + `\s`)
		if regex.MatchString(sql) {
			return true
		}
	}

	return false
}

// CommandRiskAssessor assesses risk for command operations
type CommandRiskAssessor struct{}

// NewCommandRiskAssessor creates a new command risk assessor
func NewCommandRiskAssessor() *CommandRiskAssessor {
	return &CommandRiskAssessor{}
}

// AssessRisk evaluates command operation risk
// Priority: (1) LLM-provided risk_level, (2) Code whitelist (ls, cat, pwd, echo, grep, etc.), (3) Require confirmation
func (r *CommandRiskAssessor) AssessRisk(toolName string, args map[string]interface{}) RiskLevel {
	// Priority 1: Check LLM-provided risk_level
	if riskLevelStr, ok := extractRiskLevel(args); ok {
		LogRiskAssessment("Command: LLM provided risk_level=%s", riskLevelStr)
		return assessRiskFromLLM(riskLevelStr)
	}

	// Priority 2: Check code whitelist
	if command, ok := args["command"].(string); ok {
		if isCommandWhitelisted(command) {
			LogRiskAssessment("Command: Whitelisted operation (low risk), command: %s", getCommandName(command))
			return RiskLow
		}
		LogRiskAssessment("Command: Not whitelisted, command: %s", getCommandName(command))
	}

	// Priority 3: Conservative default (require confirmation)
	LogRiskAssessment("Command: Default to high risk (requires confirmation)")
	return RiskHigh
}

// getCommandName extracts command name for logging
func getCommandName(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "unknown"
	}
	// Skip environment variable assignments
	firstCmd := parts[0]
	if strings.Contains(firstCmd, "=") {
		if len(parts) > 1 {
			firstCmd = parts[1]
		} else {
			return "unknown"
		}
	}
	// Remove path prefix
	if idx := strings.LastIndex(firstCmd, "/"); idx >= 0 {
		return firstCmd[idx+1:]
	}
	return firstCmd
}

// isCommandWhitelisted checks if command is in whitelist (read-only safe commands)
func isCommandWhitelisted(command string) bool {
	// Extract first command (handle env vars like VAR=value command)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}

	// Skip environment variable assignments (VAR=value)
	firstCmd := parts[0]
	if strings.Contains(firstCmd, "=") {
		if len(parts) > 1 {
			firstCmd = parts[1]
		} else {
			return false
		}
	}

	// Remove path prefix (e.g., /usr/bin/ls -> ls)
	cmdName := firstCmd
	if idx := strings.LastIndex(firstCmd, "/"); idx >= 0 {
		cmdName = firstCmd[idx+1:]
	}

	// Whitelist of safe read-only commands
	whitelist := map[string]bool{
		"ls":       true,
		"cat":      true,
		"pwd":      true,
		"echo":     true,
		"grep":     true,
		"head":     true,
		"tail":     true,
		"wc":       true,
		"find":     true, // read-only find (without -delete)
		"which":    true,
		"type":     true,
		"whereis":  true,
		"locate":   true,
		"stat":     true,
		"file":     true,
		"date":     true,
		"uptime":   true,
		"whoami":   true,
		"id":       true,
		"env":      true,
		"printenv": true,
	}

	return whitelist[cmdName]
}

// FileOperationRiskAssessor assesses risk for file operations
type FileOperationRiskAssessor struct{}

// NewFileOperationRiskAssessor creates a new file operation risk assessor
func NewFileOperationRiskAssessor() *FileOperationRiskAssessor {
	return &FileOperationRiskAssessor{}
}

// AssessRisk evaluates file operation risk
// Priority: (1) LLM-provided risk_level, (2) Code whitelist (read, list, exists), (3) Require confirmation
func (r *FileOperationRiskAssessor) AssessRisk(toolName string, args map[string]interface{}) RiskLevel {
	// Priority 1: Check LLM-provided risk_level
	if riskLevelStr, ok := extractRiskLevel(args); ok {
		LogRiskAssessment("FileOperation: LLM provided risk_level=%s", riskLevelStr)
		return assessRiskFromLLM(riskLevelStr)
	}

	// Priority 2: Check code whitelist
	if operation, ok := args["operation"].(string); ok {
		if isFileOperationWhitelisted(operation) {
			LogRiskAssessment("FileOperation: Whitelisted operation (low risk), operation: %s", operation)
			return RiskLow
		}
		LogRiskAssessment("FileOperation: Not whitelisted, operation: %s", operation)
	}

	// Priority 3: Conservative default (require confirmation)
	LogRiskAssessment("FileOperation: Default to high risk (requires confirmation)")
	return RiskHigh
}

// isFileOperationWhitelisted checks if file operation is in whitelist (read-only operations)
func isFileOperationWhitelisted(operation string) bool {
	whitelist := map[string]bool{
		"read":   true,
		"list":   true,
		"exists": true,
	}
	return whitelist[strings.ToLower(operation)]
}

// HTTPRequestRiskAssessor assesses risk for HTTP request operations
type HTTPRequestRiskAssessor struct{}

// NewHTTPRequestRiskAssessor creates a new HTTP request risk assessor
func NewHTTPRequestRiskAssessor() *HTTPRequestRiskAssessor {
	return &HTTPRequestRiskAssessor{}
}

// AssessRisk evaluates HTTP request risk
// Priority: (1) LLM-provided risk_level, (2) Code whitelist (GET, HEAD, OPTIONS), (3) Require confirmation
func (r *HTTPRequestRiskAssessor) AssessRisk(toolName string, args map[string]interface{}) RiskLevel {
	// Priority 1: Check LLM-provided risk_level
	if riskLevelStr, ok := extractRiskLevel(args); ok {
		LogRiskAssessment("HTTPRequest: LLM provided risk_level=%s", riskLevelStr)
		return assessRiskFromLLM(riskLevelStr)
	}

	// Priority 2: Check code whitelist
	method := "GET" // Default method
	if m, ok := args["method"].(string); ok {
		method = strings.ToUpper(m)
	}
	if isHTTPMethodWhitelisted(method) {
		LogRiskAssessment("HTTPRequest: Whitelisted method (low risk), method: %s", method)
		return RiskLow
	}
	LogRiskAssessment("HTTPRequest: Not whitelisted, method: %s", method)

	// Priority 3: Conservative default (require confirmation)
	LogRiskAssessment("HTTPRequest: Default to high risk (requires confirmation)")
	return RiskHigh
}

// isHTTPMethodWhitelisted checks if HTTP method is in whitelist (safe read-only methods)
func isHTTPMethodWhitelisted(method string) bool {
	whitelist := map[string]bool{
		"GET":     true,
		"HEAD":    true,
		"OPTIONS": true,
	}
	return whitelist[strings.ToUpper(method)]
}

// GetRiskAssessor returns the appropriate risk assessor for a tool
func GetRiskAssessor(toolName string) RiskAssessor {
	switch toolName {
	case "execute_sql":
		return NewSQLRiskAssessor()
	case "execute_command":
		return NewCommandRiskAssessor()
	case "file_operations":
		return NewFileOperationRiskAssessor()
	case "http_request":
		return NewHTTPRequestRiskAssessor()
	default:
		// Default: conservative assessor that always requires confirmation
		return &DefaultRiskAssessor{}
	}
}

// DefaultRiskAssessor is a conservative risk assessor that requires confirmation for unknown tools
type DefaultRiskAssessor struct{}

// AssessRisk always requires confirmation for unknown tools (conservative default)
func (r *DefaultRiskAssessor) AssessRisk(toolName string, args map[string]interface{}) RiskLevel {
	// Check if LLM provided risk_level
	if riskLevelStr, ok := extractRiskLevel(args); ok {
		LogRiskAssessment("UnknownTool: LLM provided risk_level=%s, tool: %s", riskLevelStr, toolName)
		return assessRiskFromLLM(riskLevelStr)
	}
	// Unknown tool without LLM risk_level -> require confirmation
	LogRiskAssessment("UnknownTool: Default to high risk (requires confirmation), tool: %s", toolName)
	return RiskHigh
}
