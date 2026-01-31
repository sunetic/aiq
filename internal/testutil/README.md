# Test Utilities

This directory contains shared test utilities, mocks, and fixtures for the aiq project.

## Structure

```
testutil/
├── mocks/          # Mock implementations for external dependencies
│   └── llm_client.go
├── fixtures/        # Test data fixtures
│   ├── sessions/    # Session file fixtures
│   ├── configs/    # Configuration file fixtures
│   ├── prompts/    # Prompt file fixtures
│   └── skills/     # Skills fixtures
├── helpers.go      # Helper functions for common test operations
└── README.md        # This file
```

## Usage

### Mock LLM Client

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

### Test Helpers

```go
import "github.com/aiq/aiq/internal/testutil"

// Create temporary directory
tmpDir, cleanup := testutil.CreateTempDir(t)
defer cleanup()

// Create temporary session
sess, sessionPath := testutil.CreateTempSession(t, "test_source", "mysql")

// Load test fixture
data := testutil.MustLoadTestFixture(t, "sessions/legacy_session.json")
```

## Testing Patterns

### Unit Tests

- Use mocks for external dependencies (LLM API, databases)
- Use fixtures for test data
- Use helpers for common operations (temp dirs, sessions)

### Integration Tests

- Use `//go:build integration` build tag
- May require external dependencies (test databases)
- Run with: `go test -tags=integration ./...`

## Adding New Fixtures

1. Place fixture files in appropriate subdirectory under `fixtures/`
2. Use descriptive names (e.g., `legacy_session.json`, `invalid_config.yaml`)
3. Document fixture purpose in comments or this README

## Adding New Mocks

1. Create mock in `mocks/` directory
2. Implement same interface as real implementation
3. Provide methods for setting responses and verifying calls
4. Document usage in this README
