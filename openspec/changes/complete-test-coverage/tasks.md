## 1. Test Infrastructure Setup

- [x] 1.1 Create `internal/testutil/` directory structure
- [x] 1.2 Implement mock LLM client interface and basic implementation
- [x] 1.3 Create test fixtures directory structure (sessions, configs, prompts, skills)
- [x] 1.4 Implement helper functions for creating temporary sessions
- [x] 1.5 Implement helper functions for creating temporary directories
- [x] 1.6 Implement helper functions for loading test fixtures
- [ ] 1.7 Create mock implementations for UI components (if needed)
- [x] 1.8 Document testing patterns and conventions

## 2. Unit Tests for Prompt Management

- [x] 2.1 Test prompt loader initialization (creates defaults, loads from user dir)
- [x] 2.2 Test prompt version detection (SHA256 hash comparison)
- [x] 2.3 Test version-based choice persistence (store/retrieve per version)
- [x] 2.4 Test prompt upgrade flow (overwrite vs keep choice)
- [x] 2.5 Test prompt file error handling (missing files, invalid content)
- [x] 2.6 Test prompt loader with various prompt file scenarios

## 3. Unit Tests for Session Management

- [x] 3.1 Test complete message array storage (RawMessages serialization)
- [x] 3.2 Test complete message array loading (deserialization, conversion)
- [x] 3.3 Test content normalization (string conversion, null handling)
- [x] 3.4 Test backward compatibility (legacy Messages format)
- [x] 3.5 Test message trimming (when array exceeds limit)
- [x] 3.6 Test session save/load with various message types (system, user, assistant, tool)
- [x] 3.7 Test session manager operations (create, load, save, delete)

## 4. Unit Tests for Risk Assessment

- [x] 4.1 Test LLM-provided risk level handling (low/medium/high)
- [x] 4.2 Test code-level whitelist fallback (SQL, command, file, HTTP)
- [x] 4.3 Test risk assessment priority (LLM overrides whitelist)
- [x] 4.4 Test high-risk operation confirmation requirement
- [x] 4.5 Test risk logger functionality (logging decisions to file)
- [x] 4.6 Test risk assessor for each tool type (SQL, command, file, HTTP)

## 5. Unit Tests for Tool Implementations

- [x] 5.1 Test command tool execution (success, error, timeout scenarios)
- [x] 5.2 Test command tool timeout parameter handling
- [x] 5.3 Test command tool output capture (stdout, stderr)
- [x] 5.4 Test file tool read operations
- [x] 5.5 Test file tool write operations (with confirmation)
- [x] 5.6 Test file tool error handling (file not found, permission errors)
- [x] 5.7 Test HTTP tool GET requests (low risk, automatic execution)
- [x] 5.8 Test HTTP tool POST/DELETE requests (high risk, confirmation)
- [x] 5.9 Test HTTP tool error handling (network errors, timeouts)
- [ ] 5.10 Test SQL tool SELECT queries (low risk, automatic execution)
- [ ] 5.11 Test SQL tool DROP/TRUNCATE (high risk, confirmation)
- [ ] 5.12 Test SQL tool error handling (syntax errors, constraint violations)
- [ ] 5.13 Test SQL tool result formatting

## 6. Unit Tests for Error Extraction

- [ ] 6.1 Test structured error extraction (error code, type, affected resources)
- [ ] 6.2 Test MySQL error parsing (foreign key constraints, syntax errors)
- [ ] 6.3 Test PostgreSQL error parsing (constraint violations, syntax errors)
- [ ] 6.4 Test error dependency identification
- [ ] 6.5 Test suggested actions generation

## 7. Unit Tests for Configuration Management

- [ ] 7.1 Test configuration loading (valid/invalid configs)
- [ ] 7.2 Test configuration validation
- [ ] 7.3 Test directory structure creation
- [ ] 7.4 Test version choices storage and retrieval
- [ ] 7.5 Test configuration wizard (if testable without UI)

## 8. Unit Tests for Skills System

- [ ] 8.1 Test Skills metadata loading (from directories)
- [ ] 8.2 Test Skills parser (YAML frontmatter parsing)
- [ ] 8.3 Test Skills matcher (scoring, precision filtering)
- [ ] 8.4 Test Skills loader (progressive loading, error handling)
- [ ] 8.5 Test Skills manager operations

## 9. Unit Tests for Database Operations

- [ ] 9.1 Test database connection establishment
- [ ] 9.2 Test database connection error handling
- [ ] 9.3 Test query execution (success, error scenarios)
- [ ] 9.4 Test schema fetching (tables, columns)
- [ ] 9.5 Test database-specific schema queries (MySQL vs PostgreSQL)

## 10. Unit Tests for Chart Rendering

- [ ] 10.1 Test chart type detection (bar, line, pie, scatter)
- [ ] 10.2 Test chart configuration (axis labels, colors, styling)
- [ ] 10.3 Test chart customizer logic
- [ ] 10.4 Test chart detector functionality

## 11. Unit Tests for LLM Client

- [ ] 11.1 Test LLM client message normalization
- [ ] 11.2 Test LLM client tool calling
- [ ] 11.3 Test LLM client error handling
- [ ] 11.4 Test LLM client response parsing

## 12. Integration Tests for Conversation Flows

- [ ] 12.1 Test multi-turn conversation flow (context preservation)
- [ ] 12.2 Test error recovery flow (dependency resolution, retry)
- [ ] 12.3 Test conversation history persistence across sessions
- [ ] 12.4 Test complete message array in conversation flow

## 13. Integration Tests for Skills Integration

- [ ] 13.1 Test Skills loading and matching with various query types
- [ ] 13.2 Test Skills precision filtering (avoid irrelevant matches)
- [ ] 13.3 Test Skills integration with prompts (content inclusion)
- [ ] 13.4 Test Skills with ambiguous queries

## 14. Integration Tests for Prompt Management

- [ ] 14.1 Test prompt compression with long conversations
- [ ] 14.2 Test prompt building with multiple Skills
- [ ] 14.3 Test prompt management with Skills eviction

## 15. Integration Tests for Risk-Based Confirmation

- [ ] 15.1 Test low-risk operation automatic execution (SELECT, ls, file read)
- [ ] 15.2 Test high-risk operation confirmation (DROP, rm, file write)
- [ ] 15.3 Test LLM-provided risk level handling
- [ ] 15.4 Test unknown operation handling (not in whitelist)

## 16. Integration Tests for Tool Execution Workflows

- [ ] 16.1 Test SQL tool execution workflow (risk check, execution, display)
- [ ] 16.2 Test command tool execution workflow
- [ ] 16.3 Test file tool execution workflow
- [ ] 16.4 Test HTTP tool execution workflow
- [ ] 16.5 Test redundant output reduction (verify no duplicate summaries)

## 17. Integration Tests for Cross-Platform Compatibility

- [ ] 17.1 Test on macOS (file paths, command execution)
- [ ] 17.2 Test on Linux (file paths, command execution)
- [ ] 17.3 Test on Windows (file paths, command execution) - if applicable

## 18. Test Coverage Reporting Setup

- [ ] 18.1 Set up coverage profile generation (`go test -coverprofile`)
- [ ] 18.2 Create script to generate HTML coverage report
- [ ] 18.3 Set up coverage tracking in CI/CD (if applicable)
- [ ] 18.4 Document coverage goals and thresholds
- [ ] 18.5 Create coverage analysis script (identify gaps)

## 19. Test Documentation and Cleanup

- [ ] 19.1 Document testing patterns and best practices
- [ ] 19.2 Document how to run tests (unit vs integration)
- [ ] 19.3 Document how to add new tests
- [ ] 19.4 Review and refactor test code for consistency
- [ ] 19.5 Verify all tests pass
- [ ] 19.6 Run coverage analysis and identify remaining gaps
