## ADDED Requirements

### Requirement: Mock LLM client for testing

The system SHALL provide a mock LLM client implementation for testing without real API calls.

#### Scenario: Test with mock LLM client
- **WHEN** test uses mock LLM client
- **THEN** mock client provides predefined responses
- **AND** test can verify LLM API calls without network requests
- **AND** test can simulate various LLM response scenarios (success, error, tool calls)

#### Scenario: Test LLM client interface
- **WHEN** mock LLM client implements LLM client interface
- **THEN** mock can be substituted for real client in tests
- **AND** tests can verify message structure sent to LLM
- **AND** tests can verify tool definitions passed to LLM

#### Scenario: Test LLM response scenarios
- **WHEN** test needs to simulate specific LLM response
- **THEN** mock client allows setting response content
- **AND** mock client allows setting tool calls
- **AND** mock client allows simulating errors

### Requirement: Test fixtures for common scenarios

The system SHALL provide test fixtures for common test scenarios including sessions, configurations, and prompts.

#### Scenario: Test with session fixtures
- **WHEN** test needs session data
- **THEN** system provides fixtures for legacy session format
- **AND** system provides fixtures for new session format with RawMessages
- **AND** fixtures include various message types (user, assistant, tool)

#### Scenario: Test with configuration fixtures
- **WHEN** test needs configuration data
- **THEN** system provides fixtures for valid configurations
- **AND** system provides fixtures for invalid configurations
- **AND** fixtures cover various configuration scenarios

#### Scenario: Test with prompt fixtures
- **WHEN** test needs prompt data
- **THEN** system provides fixtures for different prompt versions
- **AND** system provides fixtures for modified prompts
- **AND** fixtures include various prompt file formats

#### Scenario: Test with Skills fixtures
- **WHEN** test needs Skills data
- **THEN** system provides fixtures for valid Skills
- **AND** system provides fixtures for invalid Skills
- **AND** fixtures include Skills with various metadata

### Requirement: Test helpers for common operations

The system SHALL provide helper functions for common test operations.

#### Scenario: Test helper for creating temporary sessions
- **WHEN** test needs a session
- **THEN** helper function creates temporary session file
- **AND** helper function cleans up after test
- **AND** helper function supports both legacy and new formats

#### Scenario: Test helper for creating mock LLM client
- **WHEN** test needs mock LLM client
- **THEN** helper function creates configured mock client
- **AND** helper function allows setting default responses
- **AND** helper function allows verifying calls

#### Scenario: Test helper for loading test fixtures
- **WHEN** test needs test data
- **THEN** helper function loads fixture by name
- **AND** helper function returns fixture content
- **AND** helper function handles missing fixtures gracefully

#### Scenario: Test helper for temporary directories
- **WHEN** test needs temporary directory
- **THEN** helper function creates temporary directory
- **AND** helper function cleans up after test
- **AND** helper function handles cleanup errors

### Requirement: Test utilities for tool execution mocking

The system SHALL provide utilities for mocking tool execution in tests.

#### Scenario: Test with mocked tool execution
- **WHEN** test needs to mock tool execution
- **THEN** utility allows setting tool execution result
- **AND** utility allows simulating tool execution errors
- **AND** utility allows verifying tool execution calls

#### Scenario: Test tool execution with various scenarios
- **WHEN** test needs to test different tool execution scenarios
- **THEN** utility supports success scenarios
- **AND** utility supports error scenarios
- **AND** utility supports timeout scenarios

### Requirement: Test organization and structure

The system SHALL organize tests in a consistent structure following Go testing conventions.

#### Scenario: Test files co-located with implementation
- **WHEN** test file is created
- **THEN** test file is placed alongside implementation file (`*_test.go`)
- **AND** test file is in same package as implementation
- **AND** test file can access unexported functions for thorough testing

#### Scenario: Test utilities in shared location
- **WHEN** test utilities are created
- **THEN** utilities are placed in `internal/testutil/` directory
- **AND** utilities are organized by category (mocks, fixtures, helpers)
- **AND** utilities are reusable across test files

#### Scenario: Integration tests separated by build tag
- **WHEN** integration test is created
- **THEN** integration test uses `//go:build integration` tag
- **AND** integration tests can be run separately from unit tests
- **AND** unit tests run faster without integration tests
