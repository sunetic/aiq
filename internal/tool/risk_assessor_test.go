package tool

import (
	"testing"
)

// TestRiskAssessor_LLMRiskLevel tests LLM-provided risk level handling
func TestRiskAssessor_LLMRiskLevel(t *testing.T) {
	t.Run("low risk level executes automatically", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		args := map[string]interface{}{
			"sql":        "DROP TABLE users",
			"risk_level": "low",
		}

		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow for risk_level='low', got %v", risk)
		}
	})

	t.Run("high risk level requires confirmation", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		args := map[string]interface{}{
			"sql":        "SELECT * FROM users",
			"risk_level": "high",
		}

		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh for risk_level='high', got %v", risk)
		}
	})

	t.Run("medium risk level requires confirmation", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		args := map[string]interface{}{
			"sql":        "SELECT * FROM users",
			"risk_level": "medium",
		}

		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh for risk_level='medium', got %v", risk)
		}
	})

	t.Run("unknown risk level defaults to high", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		args := map[string]interface{}{
			"sql":        "SELECT * FROM users",
			"risk_level": "unknown",
		}

		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh for unknown risk_level, got %v", risk)
		}
	})
}

// TestRiskAssessor_CodeWhitelist tests code-level whitelist fallback
func TestRiskAssessor_CodeWhitelist(t *testing.T) {
	t.Run("SQL whitelist fallback", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		tests := []struct {
			name     string
			sql      string
			expected RiskLevel
		}{
			{"SELECT query", "SELECT * FROM users", RiskLow},
			{"SHOW tables", "SHOW TABLES", RiskLow},
			{"DESCRIBE table", "DESCRIBE users", RiskLow},
			{"EXPLAIN query", "EXPLAIN SELECT * FROM users", RiskLow},
			{"CREATE TABLE", "CREATE TABLE test (id INT)", RiskLow},
			{"DROP TABLE", "DROP TABLE users", RiskHigh},
			{"DELETE", "DELETE FROM users", RiskHigh},
			{"TRUNCATE", "TRUNCATE TABLE users", RiskHigh},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				args := map[string]interface{}{
					"sql": tt.sql,
				}
				risk := sqlAssessor.AssessRisk("execute_sql", args)
				if risk != tt.expected {
					t.Errorf("Expected %v for SQL '%s', got %v", tt.expected, tt.sql, risk)
				}
			})
		}
	})

	t.Run("command whitelist fallback", func(t *testing.T) {
		cmdAssessor := NewCommandRiskAssessor()

		tests := []struct {
			name     string
			command  string
			expected RiskLevel
		}{
			{"ls command", "ls -la", RiskLow},
			{"cat command", "cat file.txt", RiskLow},
			{"pwd command", "pwd", RiskLow},
			{"echo command", "echo test", RiskLow},
			{"grep command", "grep pattern file.txt", RiskLow},
			{"rm command", "rm file.txt", RiskHigh},
			{"sudo command", "sudo reboot", RiskHigh},
			{"unknown command", "custom_script.sh", RiskHigh},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				args := map[string]interface{}{
					"command": tt.command,
				}
				risk := cmdAssessor.AssessRisk("execute_command", args)
				if risk != tt.expected {
					t.Errorf("Expected %v for command '%s', got %v", tt.expected, tt.command, risk)
				}
			})
		}
	})

	t.Run("file operation whitelist fallback", func(t *testing.T) {
		fileAssessor := NewFileOperationRiskAssessor()

		tests := []struct {
			name      string
			operation string
			expected  RiskLevel
		}{
			{"read operation", "read", RiskLow},
			{"list operation", "list", RiskLow},
			{"exists operation", "exists", RiskLow},
			{"write operation", "write", RiskHigh},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				args := map[string]interface{}{
					"operation": tt.operation,
				}
				risk := fileAssessor.AssessRisk("file_operations", args)
				if risk != tt.expected {
					t.Errorf("Expected %v for operation '%s', got %v", tt.expected, tt.operation, risk)
				}
			})
		}
	})

	t.Run("HTTP method whitelist fallback", func(t *testing.T) {
		httpAssessor := NewHTTPRequestRiskAssessor()

		tests := []struct {
			name     string
			method   string
			expected RiskLevel
		}{
			{"GET request", "GET", RiskLow},
			{"HEAD request", "HEAD", RiskLow},
			{"OPTIONS request", "OPTIONS", RiskLow},
			{"POST request", "POST", RiskHigh},
			{"PUT request", "PUT", RiskHigh},
			{"DELETE request", "DELETE", RiskHigh},
			{"PATCH request", "PATCH", RiskHigh},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				args := map[string]interface{}{
					"method": tt.method,
				}
				risk := httpAssessor.AssessRisk("http_request", args)
				if risk != tt.expected {
					t.Errorf("Expected %v for method '%s', got %v", tt.expected, tt.method, risk)
				}
			})
		}
	})
}

// TestRiskAssessor_Priority tests risk assessment priority
func TestRiskAssessor_Priority(t *testing.T) {
	t.Run("LLM risk_level takes priority over whitelist", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		// SELECT is whitelisted (low risk), but LLM says high risk
		args := map[string]interface{}{
			"sql":        "SELECT * FROM users",
			"risk_level": "high",
		}

		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh (LLM priority), got %v", risk)
		}
	})

	t.Run("LLM low risk overrides non-whitelisted operation", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		// DROP is not whitelisted (high risk), but LLM says low risk
		args := map[string]interface{}{
			"sql":        "DROP TABLE users",
			"risk_level": "low",
		}

		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow (LLM priority), got %v", risk)
		}
	})

	t.Run("conservative default for unknown operations", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		// Unknown SQL operation without risk_level
		args := map[string]interface{}{
			"sql": "CUSTOM_SQL_COMMAND",
		}

		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh (conservative default), got %v", risk)
		}
	})
}

// TestRiskAssessor_HighRiskConfirmation tests high-risk operation confirmation requirement
func TestRiskAssessor_HighRiskConfirmation(t *testing.T) {
	t.Run("high-risk SQL requires confirmation", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		highRiskSQLs := []string{
			"DROP TABLE users",
			"TRUNCATE TABLE users",
			"DELETE FROM users",
			"ALTER TABLE users DROP COLUMN id",
		}

		for _, sql := range highRiskSQLs {
			args := map[string]interface{}{
				"sql": sql,
			}
			risk := sqlAssessor.AssessRisk("execute_sql", args)
			if risk != RiskHigh {
				t.Errorf("Expected RiskHigh for SQL '%s', got %v", sql, risk)
			}
		}
	})

	t.Run("high-risk commands require confirmation", func(t *testing.T) {
		cmdAssessor := NewCommandRiskAssessor()

		highRiskCommands := []string{
			"rm -rf /",
			"sudo reboot",
			"init 0",
			"custom_destructive_script.sh",
		}

		for _, cmd := range highRiskCommands {
			args := map[string]interface{}{
				"command": cmd,
			}
			risk := cmdAssessor.AssessRisk("execute_command", args)
			if risk != RiskHigh {
				t.Errorf("Expected RiskHigh for command '%s', got %v", cmd, risk)
			}
		}
	})
}

// TestRiskAssessor_EachToolType tests risk assessor for each tool type
func TestRiskAssessor_EachToolType(t *testing.T) {
	t.Run("SQL risk assessor", func(t *testing.T) {
		assessor := GetRiskAssessor("execute_sql")
		if _, ok := assessor.(*SQLRiskAssessor); !ok {
			t.Errorf("Expected SQLRiskAssessor, got %T", assessor)
		}

		// Test it works
		args := map[string]interface{}{
			"sql": "SELECT * FROM users",
		}
		risk := assessor.AssessRisk("execute_sql", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow, got %v", risk)
		}
	})

	t.Run("command risk assessor", func(t *testing.T) {
		assessor := GetRiskAssessor("execute_command")
		if _, ok := assessor.(*CommandRiskAssessor); !ok {
			t.Errorf("Expected CommandRiskAssessor, got %T", assessor)
		}

		args := map[string]interface{}{
			"command": "ls -la",
		}
		risk := assessor.AssessRisk("execute_command", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow, got %v", risk)
		}
	})

	t.Run("file operation risk assessor", func(t *testing.T) {
		assessor := GetRiskAssessor("file_operations")
		if _, ok := assessor.(*FileOperationRiskAssessor); !ok {
			t.Errorf("Expected FileOperationRiskAssessor, got %T", assessor)
		}

		args := map[string]interface{}{
			"operation": "read",
		}
		risk := assessor.AssessRisk("file_operations", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow, got %v", risk)
		}
	})

	t.Run("HTTP request risk assessor", func(t *testing.T) {
		assessor := GetRiskAssessor("http_request")
		if _, ok := assessor.(*HTTPRequestRiskAssessor); !ok {
			t.Errorf("Expected HTTPRequestRiskAssessor, got %T", assessor)
		}

		args := map[string]interface{}{
			"method": "GET",
		}
		risk := assessor.AssessRisk("http_request", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow, got %v", risk)
		}
	})

	t.Run("unknown tool uses default risk assessor", func(t *testing.T) {
		assessor := GetRiskAssessor("unknown_tool")
		if _, ok := assessor.(*DefaultRiskAssessor); !ok {
			t.Errorf("Expected DefaultRiskAssessor, got %T", assessor)
		}

		args := map[string]interface{}{}
		risk := assessor.AssessRisk("unknown_tool", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh for unknown tool, got %v", risk)
		}
	})
}

// TestRiskAssessor_EdgeCases tests edge cases
func TestRiskAssessor_EdgeCases(t *testing.T) {
	t.Run("handles missing sql parameter", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		args := map[string]interface{}{}
		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh for missing sql parameter, got %v", risk)
		}
	})

	t.Run("handles empty SQL string", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		args := map[string]interface{}{
			"sql": "",
		}
		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskHigh {
			t.Errorf("Expected RiskHigh for empty SQL, got %v", risk)
		}
	})

	t.Run("handles SQL with leading whitespace", func(t *testing.T) {
		sqlAssessor := NewSQLRiskAssessor()

		args := map[string]interface{}{
			"sql": "   SELECT * FROM users",
		}
		risk := sqlAssessor.AssessRisk("execute_sql", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow for SELECT with leading whitespace, got %v", risk)
		}
	})

	t.Run("handles command with environment variables", func(t *testing.T) {
		cmdAssessor := NewCommandRiskAssessor()

		args := map[string]interface{}{
			"command": "VAR=value ls -la",
		}
		risk := cmdAssessor.AssessRisk("execute_command", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow for ls with env var, got %v", risk)
		}
	})

	t.Run("handles command with path prefix", func(t *testing.T) {
		cmdAssessor := NewCommandRiskAssessor()

		args := map[string]interface{}{
			"command": "/usr/bin/ls -la",
		}
		risk := cmdAssessor.AssessRisk("execute_command", args)
		if risk != RiskLow {
			t.Errorf("Expected RiskLow for /usr/bin/ls, got %v", risk)
		}
	})
}
