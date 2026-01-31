package tool

import (
	"errors"
	"strings"
	"testing"
)

// TestErrorExtractor_StructuredExtraction tests structured error extraction
func TestErrorExtractor_StructuredExtraction(t *testing.T) {
	t.Run("extracts error code, type, and affected resources", func(t *testing.T) {
		err := errors.New("Error 3730 (HY000): Cannot drop table 'users' referenced by foreign key constraint 'fk_sales_user' on table 'sales'")

		info := ExtractErrorInfo(err)

		if info.ErrorCode == "" {
			t.Error("Expected error code to be extracted")
		}

		if info.ErrorType == "" {
			t.Error("Expected error type to be extracted")
		}

		if len(info.AffectedResources) == 0 {
			t.Error("Expected affected resources to be extracted")
		}

		if len(info.Dependencies) == 0 {
			t.Error("Expected dependencies to be extracted")
		}

		if len(info.SuggestedActions) == 0 {
			t.Error("Expected suggested actions to be generated")
		}
	})

	t.Run("handles nil error", func(t *testing.T) {
		info := ExtractErrorInfo(nil)

		if info.ErrorCode != "" {
			t.Error("Expected empty error code for nil error")
		}

		if info.ErrorType != "" {
			t.Error("Expected empty error type for nil error")
		}
	})

	t.Run("extracts MySQL error code", func(t *testing.T) {
		err := errors.New("Error 3730 (HY000): Some error message")
		info := ExtractErrorInfo(err)

		if info.ErrorCode != "3730" {
			t.Errorf("Expected error code '3730', got %q", info.ErrorCode)
		}
	})

	t.Run("extracts PostgreSQL error code", func(t *testing.T) {
		err := errors.New("ERROR: 42P01: relation \"users\" does not exist")
		info := ExtractErrorInfo(err)

		if info.ErrorCode != "42P01" {
			t.Errorf("Expected error code '42P01', got %q", info.ErrorCode)
		}
	})
}

// TestErrorExtractor_MySQLErrorParsing tests MySQL error parsing
func TestErrorExtractor_MySQLErrorParsing(t *testing.T) {
	t.Run("parses foreign key constraint error", func(t *testing.T) {
		err := errors.New("Error 3730 (HY000): Cannot drop table 'users' referenced by foreign key constraint 'fk_sales_user' on table 'sales'")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "foreign_key_constraint" {
			t.Errorf("Expected error type 'foreign_key_constraint', got %q", info.ErrorType)
		}

		if len(info.AffectedResources) == 0 {
			t.Error("Expected affected resource 'users' to be extracted")
		}

		if len(info.Dependencies) == 0 {
			t.Error("Expected dependency 'sales' to be extracted")
		}
	})

	t.Run("parses syntax error", func(t *testing.T) {
		err := errors.New("You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'SELCT'")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "syntax_error" {
			t.Errorf("Expected error type 'syntax_error', got %q", info.ErrorType)
		}

		if len(info.AffectedResources) == 0 {
			t.Error("Expected affected resource to be extracted")
		}
	})

	t.Run("parses table not found error", func(t *testing.T) {
		err := errors.New("Table 'nonexistent' doesn't exist")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "resource_not_found" {
			t.Errorf("Expected error type 'resource_not_found', got %q", info.ErrorType)
		}

		if len(info.AffectedResources) == 0 {
			t.Error("Expected affected resource 'nonexistent' to be extracted")
		}
	})

	t.Run("parses permission denied error", func(t *testing.T) {
		err := errors.New("Access denied for user 'test'@'localhost' to database 'mydb'")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "permission_denied" {
			t.Errorf("Expected error type 'permission_denied', got %q", info.ErrorType)
		}
	})
}

// TestErrorExtractor_PostgreSQLErrorParsing tests PostgreSQL error parsing
func TestErrorExtractor_PostgreSQLErrorParsing(t *testing.T) {
	t.Run("parses foreign key constraint violation", func(t *testing.T) {
		err := errors.New("update or delete on table \"users\" violates foreign key constraint \"fk_sales_user\" on table \"sales\"")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "foreign_key_constraint" {
			t.Errorf("Expected error type 'foreign_key_constraint', got %q", info.ErrorType)
		}

		if len(info.AffectedResources) == 0 {
			t.Error("Expected affected resource to be extracted")
		}

		if len(info.Dependencies) == 0 {
			t.Error("Expected dependency to be extracted")
		}
	})

	t.Run("parses constraint violation", func(t *testing.T) {
		err := errors.New("ERROR: 23505: duplicate key value violates unique constraint \"users_pkey\"")

		info := ExtractErrorInfo(err)

		if info.ErrorCode != "23505" {
			t.Errorf("Expected error code '23505', got %q", info.ErrorCode)
		}
	})

	t.Run("parses relation does not exist", func(t *testing.T) {
		err := errors.New("ERROR: relation \"nonexistent\" does not exist")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "resource_not_found" {
			t.Errorf("Expected error type 'resource_not_found', got %q", info.ErrorType)
		}
	})

	t.Run("parses syntax error", func(t *testing.T) {
		err := errors.New("ERROR: syntax error at or near \"SELCT\"")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "syntax_error" {
			t.Errorf("Expected error type 'syntax_error', got %q", info.ErrorType)
		}
	})
}

// TestErrorExtractor_DependencyIdentification tests error dependency identification
func TestErrorExtractor_DependencyIdentification(t *testing.T) {
	t.Run("identifies foreign key dependencies", func(t *testing.T) {
		err := errors.New("Cannot drop table 'users' referenced by foreign key constraint 'fk_sales_user' on table 'sales'")

		info := ExtractErrorInfo(err)

		if len(info.Dependencies) == 0 {
			t.Error("Expected dependencies to be identified")
		}

		// Verify dependency is 'sales'
		found := false
		for _, dep := range info.Dependencies {
			if dep == "sales" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected dependency 'sales' to be identified, got %v", info.Dependencies)
		}
	})

	t.Run("identifies multiple dependencies if present", func(t *testing.T) {
		// Note: Current implementation extracts one dependency per error
		// This test verifies the extraction works
		err := errors.New("Cannot drop table 'users' referenced by foreign key constraint 'fk_sales_user' on table 'sales'")

		info := ExtractErrorInfo(err)

		if len(info.Dependencies) == 0 {
			t.Error("Expected at least one dependency to be identified")
		}
	})
}

// TestErrorExtractor_SuggestedActions tests suggested actions generation
func TestErrorExtractor_SuggestedActions(t *testing.T) {
	t.Run("generates suggested actions for foreign key constraint", func(t *testing.T) {
		err := errors.New("Cannot drop table 'users' referenced by foreign key constraint 'fk_sales_user' on table 'sales'")

		info := ExtractErrorInfo(err)

		if len(info.SuggestedActions) == 0 {
			t.Error("Expected suggested actions to be generated")
		}

		// Verify action mentions dropping dependent table or constraint
		found := false
		for _, action := range info.SuggestedActions {
			if contains(action, "drop") || contains(action, "constraint") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected suggested action to mention drop or constraint, got %v", info.SuggestedActions)
		}
	})

	t.Run("generates suggested actions for syntax error", func(t *testing.T) {
		err := errors.New("You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use near 'SELCT'")

		info := ExtractErrorInfo(err)

		if len(info.SuggestedActions) == 0 {
			t.Error("Expected suggested actions to be generated")
		}

		// Verify action mentions syntax check
		found := false
		for _, action := range info.SuggestedActions {
			if contains(action, "syntax") || contains(action, "check") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected suggested action to mention syntax or check, got %v", info.SuggestedActions)
		}
	})

	t.Run("generates suggested actions for permission denied", func(t *testing.T) {
		err := errors.New("Access denied for user 'test'")

		info := ExtractErrorInfo(err)

		if len(info.SuggestedActions) == 0 {
			t.Error("Expected suggested actions to be generated")
		}

		// Verify action mentions permissions
		found := false
		for _, action := range info.SuggestedActions {
			if contains(action, "permission") || contains(action, "privilege") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected suggested action to mention permission or privilege, got %v", info.SuggestedActions)
		}
	})

	t.Run("generates suggested actions for resource not found", func(t *testing.T) {
		err := errors.New("Table 'nonexistent' doesn't exist")

		info := ExtractErrorInfo(err)

		if len(info.SuggestedActions) == 0 {
			t.Error("Expected suggested actions to be generated")
		}
	})

	t.Run("generates suggested actions for connection error", func(t *testing.T) {
		err := errors.New("Connection refused")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "connection_error" {
			t.Errorf("Expected error type 'connection_error', got %q", info.ErrorType)
		}

		if len(info.SuggestedActions) == 0 {
			t.Error("Expected suggested actions to be generated")
		}
	})

	t.Run("generates suggested actions for timeout", func(t *testing.T) {
		err := errors.New("Query execution timed out")

		info := ExtractErrorInfo(err)

		if info.ErrorType != "timeout" {
			t.Errorf("Expected error type 'timeout', got %q", info.ErrorType)
		}

		if len(info.SuggestedActions) == 0 {
			t.Error("Expected suggested actions to be generated")
		}
	})
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}
