## Context

The project currently has minimal test coverage (9 test files for 63 implementation files, ~14% coverage). Existing tests follow Go's standard testing patterns using the built-in `testing` package. The codebase includes:

- **Core modules**: CLI, configuration, session management, prompt management
- **LLM integration**: Client for API calls, tool calling, message handling
- **Tools**: Built-in tools (command, file, HTTP, SQL) with risk assessment
- **Skills system**: Loading, matching, and integration with prompts
- **Database operations**: Connection management, query execution, schema fetching
- **UI components**: Menu system, table formatting, chart rendering

Many features were implemented without tests, and archived changes show incomplete test tasks. The project needs systematic test coverage to enable safe refactoring and catch regressions.

## Goals / Non-Goals

**Goals:**
- Achieve minimum 70% code coverage for critical paths (core business logic)
- 100% coverage for error handling and edge cases
- Comprehensive unit tests for all packages in `internal/`
- Integration tests for end-to-end workflows (conversation flows, tool execution)
- Establish reusable test infrastructure (mocks, fixtures, helpers)
- Enable fast, reliable test execution in CI/CD

**Non-Goals:**
- 100% code coverage (unrealistic, focus on critical paths)
- Testing external dependencies (LLM APIs, databases) - use mocks
- Testing UI rendering details (focus on logic, not visual output)
- Performance/load testing (separate concern)
- Manual testing scenarios (automated tests only)

## Decisions

### 1. Testing Framework: Go `testing` + `testify` (Optional)

**Decision**: Use Go's built-in `testing` package as primary framework. Consider `testify` for assertions and mocks only where it significantly improves readability.

**Rationale**: 
- Go's `testing` package is standard and sufficient for most cases
- `testify` can help with complex assertions and mocking, but adds dependency
- Prefer table-driven tests (Go idiom) over assertion libraries for simple cases

**Alternatives Considered**:
- Pure `testing` package: More verbose but no dependencies
- Full `testify` adoption: Cleaner assertions but adds dependency and learning curve

**Decision**: Start with `testing` package, introduce `testify` selectively for complex mocking scenarios (LLM client, UI interactions).

### 2. Mock Strategy: Interface-based Mocking

**Decision**: Use interface-based mocking for external dependencies (LLM client, UI, file system operations).

**Rationale**:
- Go's interface system enables easy mocking without external libraries
- Allows testing without real API calls or file system operations
- Enables dependency injection for testability

**Implementation**:
- Create test interfaces for `llm.Client`, `ui.UI`, file operations
- Provide mock implementations in `internal/testutil/mocks/`
- Use dependency injection or constructor injection for testable code

**Example**:
```go
// internal/testutil/mocks/llm_client.go
type MockLLMClient struct {
    ChatWithToolsFunc func(ctx context.Context, messages []interface{}, tools []llm.Function) (*llm.ChatResponse, error)
}
```

### 3. Test Organization: Co-located Test Files

**Decision**: Place test files (`*_test.go`) alongside implementation files in the same package.

**Rationale**:
- Standard Go convention
- Tests can access unexported functions/types for thorough testing
- Easy to find and maintain tests

**Structure**:
```
internal/
  tool/
    builtin/
      command_tool.go
      command_tool_test.go  # Co-located
  testutil/                # Shared test utilities
    mocks/
      llm_client.go
      ui.go
    fixtures/
      sessions/
      configs/
    helpers.go
```

### 4. Integration Test Strategy: Separate `integration` Build Tag

**Decision**: Use build tags to separate unit tests from integration tests.

**Rationale**:
- Integration tests may require external dependencies (test databases, network)
- Allows running fast unit tests separately from slower integration tests
- CI can run both, developers can run unit tests locally

**Implementation**:
- Unit tests: No build tag, run with `go test ./...`
- Integration tests: `//go:build integration` tag, run with `go test -tags=integration ./...`

### 5. Test Data Management: Fixtures and Helpers

**Decision**: Create test fixtures and helper functions in `internal/testutil/` for common test scenarios.

**Rationale**:
- Reduces test code duplication
- Provides consistent test data
- Makes tests more readable and maintainable

**Fixtures**:
- Session files (legacy and new formats)
- Configuration files (valid and invalid)
- Prompt files (various versions)
- Skills metadata

**Helpers**:
- `CreateTempSession(t *testing.T) *session.Session`
- `CreateMockLLMClient() *testutil.MockLLMClient`
- `LoadTestFixture(name string) []byte`

### 6. Coverage Measurement: `go test -coverprofile`

**Decision**: Use Go's built-in coverage tools with `-coverprofile` for detailed analysis.

**Rationale**:
- Built-in, no external dependencies
- Provides line-by-line coverage
- Can generate HTML reports with `go tool cover`

**Workflow**:
- Run tests with coverage: `go test -coverprofile=coverage.out ./...`
- Generate HTML report: `go tool cover -html=coverage.out`
- CI can track coverage trends over time

### 7. Testing External Dependencies: Mock Everything

**Decision**: Mock all external dependencies (LLM API, databases, file system, network).

**Rationale**:
- Tests should be fast, isolated, and deterministic
- External services may be unavailable or rate-limited
- Focuses tests on application logic, not external services

**What to Mock**:
- LLM API calls (use mock client with predefined responses)
- Database connections (use in-memory SQLite or mocks)
- File system operations (use `os` package with temp directories)
- Network requests (mock HTTP client)
- User input (mock UI functions)

## Risks / Trade-offs

**[Risk] Test Maintenance Burden**
- **Mitigation**: Focus on testing critical paths and error cases. Use table-driven tests to reduce duplication. Establish clear testing patterns early.

**[Risk] Tests May Reveal Existing Bugs**
- **Mitigation**: Document discovered bugs, fix critical ones immediately, create issues for non-critical bugs. Don't let bugs block test implementation.

**[Risk] Mock Complexity**
- **Mitigation**: Start with simple mocks, refactor as needed. Use interfaces to keep mocking straightforward. Document mock usage patterns.

**[Risk] Integration Tests Require External Dependencies**
- **Mitigation**: Use build tags to separate integration tests. Provide clear documentation on setting up test environment. Use Docker or test containers for consistent environments.

**[Risk] Coverage Goals May Be Too Ambitious**
- **Mitigation**: Start with critical modules, iterate. Focus on quality over quantity. 70% coverage is a goal, not a hard requirement - prioritize critical paths.

**[Trade-off] Test Speed vs. Coverage**
- **Decision**: Prioritize fast unit tests. Use integration tests sparingly for critical workflows. Mock external dependencies to keep tests fast.

**[Trade-off] Test Readability vs. Coverage**
- **Decision**: Prefer readable, maintainable tests over achieving 100% coverage. Use table-driven tests and clear test names. Document complex test scenarios.

## Migration Plan

**Phase 1: Test Infrastructure (Week 1)**
1. Create `internal/testutil/` directory structure
2. Implement mock LLM client
3. Create test fixtures and helpers
4. Document testing patterns

**Phase 2: Core Module Tests (Weeks 2-3)**
1. Session management tests (message persistence, normalization)
2. Prompt management tests (loader, version detection)
3. Configuration management tests
4. Risk assessment tests

**Phase 3: Tool Tests (Week 4)**
1. Command tool tests
2. File tool tests
3. HTTP tool tests
4. SQL tool tests

**Phase 4: Integration Tests (Week 5)**
1. End-to-end conversation flow tests
2. Skills integration tests
3. Error handling and retry tests

**Phase 5: Coverage Analysis and Gaps (Week 6)**
1. Run coverage analysis
2. Identify gaps in critical paths
3. Fill remaining gaps
4. Set up CI coverage reporting

**Rollback Strategy**: N/A - Adding tests doesn't affect production code. If tests reveal bugs, fix bugs rather than removing tests.

## Open Questions

1. **Should we use `testify` for all tests or selectively?**
   - **Decision needed**: Start selective, evaluate after initial tests

2. **How to handle testing UI components that require terminal interaction?**
   - **Decision needed**: Mock UI functions, focus on logic testing

3. **Should integration tests run in CI by default?**
   - **Decision needed**: Yes, but allow skipping with flag if dependencies unavailable

4. **What's the minimum coverage threshold for CI to pass?**
   - **Decision needed**: Start with 50%, increase to 70% over time

5. **How to test LLM prompt engineering without real API calls?**
   - **Decision needed**: Mock LLM client, verify prompt structure and content
