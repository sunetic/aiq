## Why

The project currently has very low test coverage (approximately 14% - only 9 test files for 63 implementation files), and many critical features lack comprehensive test coverage. Multiple archived changes have incomplete test tasks, indicating systematic gaps in testing. Without proper test coverage, we risk regressions, bugs in production, and difficulty refactoring safely. This change will establish comprehensive test coverage across all major modules to ensure code quality, catch bugs early, and enable confident refactoring.

## What Changes

- **Unit Tests**: Add comprehensive unit tests for all major modules including:
  - Prompt management (loader, version detection, upgrade flow)
  - Session management (message persistence, content normalization, backward compatibility)
  - Risk assessment and confirmation mechanisms
  - Tool implementations (command, file, HTTP, SQL tools)
  - Error extraction and handling
  - Configuration management
  - Skills loading and matching
  - Database operations
  - Chart rendering and visualization

- **Integration Tests**: Add integration tests for:
  - End-to-end conversation flows (multi-turn, error recovery)
  - Skills integration with various query types
  - Prompt management with long conversations
  - Risk-based confirmation workflows
  - Cross-platform compatibility (macOS, Linux, Windows)

- **Test Infrastructure**: Establish testing patterns and utilities:
  - Mock LLM client for testing without API calls
  - Test fixtures for common scenarios
  - Test helpers for session and configuration management
  - Test utilities for tool execution mocking

- **Test Coverage Goals**: Achieve minimum 70% code coverage for critical paths, with 100% coverage for error handling and edge cases.

## Capabilities

### New Capabilities
- `unit-test-coverage`: Comprehensive unit test suite covering all internal packages
- `integration-test-coverage`: Integration tests for end-to-end workflows and cross-component interactions
- `test-infrastructure`: Testing utilities, mocks, and helpers to support efficient test development
- `test-coverage-reporting`: Automated test coverage reporting and tracking

### Modified Capabilities
- None (this is purely adding test coverage, not changing functional requirements)

## Impact

**Affected Code**:
- All packages in `internal/` directory will receive test coverage
- Test files will be added alongside implementation files (`*_test.go`)
- New test utilities may be added to `internal/testutil/` or similar

**Dependencies**:
- Testing framework: Go's built-in `testing` package
- Mocking: May introduce `testify` or similar for assertions and mocks
- Coverage tools: `go test -cover` and `go test -coverprofile`

**Systems**:
- CI/CD pipeline may need updates to run tests and report coverage
- Development workflow: Developers should run tests before committing

**Risk**:
- Low risk - adding tests doesn't change production code behavior
- May reveal existing bugs during test implementation
- Test infrastructure setup may require some initial investment
