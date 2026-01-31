package tool

import (
	"fmt"
	"regexp"
	"strings"
)

// ErrorInfo contains structured error information extracted from error messages
type ErrorInfo struct {
	ErrorCode         string   // Database error code if available
	ErrorType         string   // Categorized error type
	AffectedResources []string // List of affected resources (tables, columns, etc.)
	Dependencies      []string // List of dependencies (resources that must be resolved first)
	SuggestedActions  []string // Suggested actions to resolve the error
}

// ExtractErrorInfo extracts structured error information from an error
func ExtractErrorInfo(err error) ErrorInfo {
	if err == nil {
		return ErrorInfo{}
	}

	errorMsg := err.Error()
	info := ErrorInfo{
		ErrorType:         categorizeErrorType(errorMsg),
		AffectedResources: []string{},
		Dependencies:      []string{},
		SuggestedActions:  []string{},
	}

	// Extract error code if present (e.g., "Error 3730 (HY000)")
	if code := extractErrorCode(errorMsg); code != "" {
		info.ErrorCode = code
	}

	// Extract information based on error type
	switch info.ErrorType {
	case "foreign_key_constraint":
		if affected, dep, constraint, found := extractForeignKeyConstraint(errorMsg); found {
			info.AffectedResources = append(info.AffectedResources, affected)
			info.Dependencies = append(info.Dependencies, dep)
			info.SuggestedActions = append(info.SuggestedActions,
				fmt.Sprintf("Drop the dependent table '%s' first, or drop the foreign key constraint '%s'", dep, constraint))
		}
	case "syntax_error":
		if affected, found := extractSyntaxError(errorMsg); found {
			info.AffectedResources = append(info.AffectedResources, affected)
			info.SuggestedActions = append(info.SuggestedActions, "Check SQL syntax and correct the error")
		}
	case "permission_denied":
		if affected, found := extractPermissionError(errorMsg); found {
			info.AffectedResources = append(info.AffectedResources, affected)
			info.SuggestedActions = append(info.SuggestedActions, "Check user permissions and grant necessary privileges")
		}
	case "resource_not_found":
		if affected, found := extractResourceNotFound(errorMsg); found {
			info.AffectedResources = append(info.AffectedResources, affected)
			info.SuggestedActions = append(info.SuggestedActions, "Verify the resource exists or create it first")
		}
	case "resource_exists":
		if affected, found := extractResourceExists(errorMsg); found {
			info.AffectedResources = append(info.AffectedResources, affected)
			info.SuggestedActions = append(info.SuggestedActions, "Use IF NOT EXISTS clause or drop existing resource first")
		}
	case "connection_error":
		info.SuggestedActions = append(info.SuggestedActions, "Check database connection and network connectivity")
	case "timeout":
		info.SuggestedActions = append(info.SuggestedActions, "Operation timed out, consider increasing timeout or optimizing query")
	}

	return info
}

// extractErrorCode extracts database error code from error message
// Examples: "Error 3730 (HY000)" -> "3730", "ERROR: 42P01" -> "42P01"
func extractErrorCode(errorMsg string) string {
	// MySQL format: "Error 3730 (HY000): message"
	re := regexp.MustCompile(`(?i)Error\s+(\d+)\s*\(`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1]
	}

	// PostgreSQL format: "ERROR: 42P01: message"
	re = regexp.MustCompile(`(?i)ERROR:\s*([0-9A-Z]+):`)
	matches = re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractForeignKeyConstraint extracts foreign key constraint information
// Pattern: "Cannot drop table 'X' referenced by foreign key constraint 'Y' on table 'Z'"
// Returns: affected resource (X), dependency (Z), constraint name (Y), found (bool)
func extractForeignKeyConstraint(errorMsg string) (affectedResource, dependency, constraint string, found bool) {
	// MySQL/SeekDB pattern: "Cannot drop table 'X' referenced by foreign key constraint 'Y' on table 'Z'"
	re := regexp.MustCompile(`(?i)Cannot drop (?:table|column)\s+['"]([^'"]+)['"].*?foreign key constraint\s+['"]([^'"]+)['"].*?(?:on|in) (?:table|column)\s+['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) >= 4 {
		return matches[1], matches[3], matches[2], true
	}

	// Alternative pattern: "Cannot drop table 'X' because it is referenced by foreign key 'Y' from table 'Z'"
	re = regexp.MustCompile(`(?i)Cannot drop (?:table|column)\s+['"]([^'"]+)['"].*?referenced by foreign key\s+['"]([^'"]+)['"].*?from (?:table|column)\s+['"]([^'"]+)['"]`)
	matches = re.FindStringSubmatch(errorMsg)
	if len(matches) >= 4 {
		return matches[1], matches[3], matches[2], true
	}

	// PostgreSQL pattern: "update or delete on table \"X\" violates foreign key constraint \"Y\" on table \"Z\""
	re = regexp.MustCompile(`(?i)(?:update|delete|drop).*?table\s+['"]([^'"]+)['"].*?violates foreign key constraint\s+['"]([^'"]+)['"].*?on table\s+['"]([^'"]+)['"]`)
	matches = re.FindStringSubmatch(errorMsg)
	if len(matches) >= 4 {
		return matches[1], matches[3], matches[2], true
	}

	return "", "", "", false
}

// extractSyntaxError extracts syntax error information
// Pattern: "You have an error in your SQL syntax near 'X'"
func extractSyntaxError(errorMsg string) (affectedResource string, found bool) {
	// MySQL pattern: "You have an error in your SQL syntax; check the manual..."
	re := regexp.MustCompile(`(?i)(?:You have an error in your SQL syntax|syntax error).*?(?:near|at)\s+['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// PostgreSQL pattern: "syntax error at or near \"X\""
	re = regexp.MustCompile(`(?i)syntax error.*?(?:at or near|near)\s+['"]([^'"]+)['"]`)
	matches = re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// Generic pattern: "syntax error"
	if strings.Contains(strings.ToLower(errorMsg), "syntax error") {
		return "SQL query", true
	}

	return "", false
}

// extractPermissionError extracts permission error information
// Pattern: "Access denied for user 'X'"
func extractPermissionError(errorMsg string) (affectedResource string, found bool) {
	// MySQL pattern: "Access denied for user 'X'@'Y' to database 'Z'"
	re := regexp.MustCompile(`(?i)Access denied.*?(?:for user|to database|to table)\s+['"]([^'"]+)['"]`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// PostgreSQL pattern: "permission denied for table X"
	re = regexp.MustCompile(`(?i)permission denied.*?(?:for|on)\s+(?:table|database|schema)\s+['"]?([^'"]+)['"]?`)
	matches = re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// Generic patterns
	patterns := []string{
		"permission denied",
		"access denied",
		"insufficient privileges",
		"privilege",
	}
	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(errorMsg), pattern) {
			return "resource", true
		}
	}

	return "", false
}

// extractResourceNotFound extracts resource not found information
// Pattern: "Table 'X' doesn't exist"
func extractResourceNotFound(errorMsg string) (affectedResource string, found bool) {
	// MySQL pattern: "Table 'X' doesn't exist"
	re := regexp.MustCompile(`(?i)(?:Table|Column|Database|Schema)\s+['"]([^'"]+)['"].*?doesn't exist`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// PostgreSQL pattern: "relation \"X\" does not exist"
	re = regexp.MustCompile(`(?i)(?:relation|table|column|database|schema)\s+['"]([^'"]+)['"].*?does not exist`)
	matches = re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// Generic pattern
	if strings.Contains(strings.ToLower(errorMsg), "doesn't exist") ||
		strings.Contains(strings.ToLower(errorMsg), "does not exist") {
		// Try to extract resource name
		re = regexp.MustCompile(`['"]([^'"]+)['"]`)
		matches = re.FindStringSubmatch(errorMsg)
		if len(matches) > 1 {
			return matches[1], true
		}
		return "resource", true
	}

	return "", false
}

// extractResourceExists extracts resource exists information
// Pattern: "Table 'X' already exists"
func extractResourceExists(errorMsg string) (affectedResource string, found bool) {
	// MySQL pattern: "Table 'X' already exists"
	re := regexp.MustCompile(`(?i)(?:Table|Column|Database|Schema)\s+['"]([^'"]+)['"].*?already exists`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// PostgreSQL pattern: "relation \"X\" already exists"
	re = regexp.MustCompile(`(?i)(?:relation|table|column|database|schema)\s+['"]([^'"]+)['"].*?already exists`)
	matches = re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1], true
	}

	// Generic pattern
	if strings.Contains(strings.ToLower(errorMsg), "already exists") {
		// Try to extract resource name
		re = regexp.MustCompile(`['"]([^'"]+)['"]`)
		matches = re.FindStringSubmatch(errorMsg)
		if len(matches) > 1 {
			return matches[1], true
		}
		return "resource", true
	}

	return "", false
}

// categorizeErrorType categorizes error into standard types
func categorizeErrorType(errorMsg string) string {
	errorLower := strings.ToLower(errorMsg)

	// Foreign key constraint
	if strings.Contains(errorLower, "foreign key") ||
		strings.Contains(errorLower, "referenced by") ||
		strings.Contains(errorLower, "violates foreign key") {
		return "foreign_key_constraint"
	}

	// Syntax error
	if strings.Contains(errorLower, "syntax error") ||
		strings.Contains(errorLower, "error in your sql syntax") {
		return "syntax_error"
	}

	// Permission denied
	if strings.Contains(errorLower, "permission denied") ||
		strings.Contains(errorLower, "access denied") ||
		strings.Contains(errorLower, "insufficient privileges") ||
		strings.Contains(errorLower, "privilege") {
		return "permission_denied"
	}

	// Resource not found
	if strings.Contains(errorLower, "doesn't exist") ||
		strings.Contains(errorLower, "does not exist") ||
		strings.Contains(errorLower, "not found") {
		return "resource_not_found"
	}

	// Resource exists
	if strings.Contains(errorLower, "already exists") {
		return "resource_exists"
	}

	// Connection error
	if strings.Contains(errorLower, "connection") ||
		strings.Contains(errorLower, "connect") ||
		strings.Contains(errorLower, "network") {
		return "connection_error"
	}

	// Timeout
	if strings.Contains(errorLower, "timeout") ||
		strings.Contains(errorLower, "timed out") {
		return "timeout"
	}

	// Unknown
	return "unknown"
}
