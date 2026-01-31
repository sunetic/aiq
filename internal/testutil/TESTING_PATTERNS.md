# Testing Patterns and Conventions

This document describes testing patterns and conventions used in the aiq project.

## Test Organization

### File Naming
- Test files are co-located with implementation files: `*_test.go`
- Test files are in the same package as the code they test
- This allows access to unexported functions for thorough testing

### Package Structure
```
internal/
  package/
    implementation.go
    implementation_test.go  # Co-located
  testutil/                 # Shared test utilities
    mocks/
    fixtures/
    helpers.go
```

## Test Types

### Unit Tests
- Test individual functions and methods in isolation
- Use mocks for external dependencies
- Fast execution, no external dependencies
- Run with: `go test ./...`

### Integration Tests
- Test interactions between components
- May require external dependencies (test databases)
- Use `//go:build integration` build tag
- Run with: `go test -tags=integration ./...`

## Mocking Strategy

### LLM Client Mock
Use `testutil/mocks.MockLLMClient` for testing without real API calls:

```go
import "github.com/aiq/aiq/internal/testutil/mocks"

mockClient := mocks.NewMockLLMClient()
mockClient.WithResponse(&llm.ChatResponse{
    Choices: []struct{
        Message struct{
            Role: "assistant",
            Content: "Test response",
        },
        FinishReason: "stop",
    },
})
```

### File System Operations
- Use `t.TempDir()` for temporary directories
- Use `testutil.CreateTempDir()` helper for additional cleanup if needed
- Never use real user directories in tests

### Database Operations
- Use in-memory databases (SQLite) for testing
- Mock database connections for unit tests
- Use test databases for integration tests

## Test Helpers

### Common Helpers
- `testutil.CreateTempDir(t)` - Create temporary directory
- `testutil.CreateTempSession(t, source, dbType)` - Create temporary session
- `testutil.LoadTestFixture(name)` - Load test fixture
- `testutil.MustLoadTestFixture(t, name)` - Load fixture, fail test on error

### Fixtures
- Store test data in `testutil/fixtures/`
- Use descriptive names: `legacy_session.json`, `invalid_config.yaml`
- Keep fixtures minimal and focused

## Test Patterns

### Table-Driven Tests
Prefer table-driven tests for multiple similar test cases:

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case1", "input1", "output1", false},
        {"case2", "input2", "output2", false},
        {"error", "bad", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("unexpected error: %v", err)
            }
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

### Test Cleanup
Always clean up resources:

```go
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    defer os.RemoveAll(tmpDir) // Explicit cleanup if needed
    
    // Or use helper
    tmpDir, cleanup := testutil.CreateTempDir(t)
    defer cleanup()
}
```

### Error Testing
Test both success and error cases:

```go
func TestFunction(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Test success case
    })
    
    t.Run("error", func(t *testing.T) {
        // Test error case
    })
}
```

## Coverage Goals

### Critical Paths
- Target minimum 70% coverage for core business logic
- Focus on error handling and edge cases

### Error Handling
- Target 100% coverage for error paths
- Test all error conditions
- Test error propagation

## Running Tests

### Unit Tests Only
```bash
go test ./...
```

### Integration Tests
```bash
go test -tags=integration ./...
```

### With Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Specific Package
```bash
go test ./internal/session/...
```

## Best Practices

1. **Keep tests simple and focused**: One test, one concern
2. **Use descriptive test names**: `TestFunctionName_Scenario_ExpectedResult`
3. **Test behavior, not implementation**: Focus on what, not how
4. **Use mocks for external dependencies**: Keep tests fast and isolated
5. **Clean up resources**: Always clean up temporary files and directories
6. **Document complex tests**: Add comments explaining non-obvious test logic
7. **Test edge cases**: Empty inputs, nil values, boundary conditions
8. **Test error cases**: Don't just test happy path

## Common Pitfalls

1. **Testing implementation details**: Test behavior, not internal structure
2. **Over-mocking**: Only mock external dependencies, not internal functions
3. **Shared state**: Avoid shared state between tests
4. **Flaky tests**: Ensure tests are deterministic and don't depend on timing
5. **Ignoring errors**: Always check and test error conditions
