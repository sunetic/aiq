## ADDED Requirements

### Requirement: Integration test coverage for conversation flows

The system SHALL have integration tests for end-to-end conversation workflows including multi-turn conversations and error recovery.

#### Scenario: Test multi-turn conversation flow
- **WHEN** user has a multi-turn conversation
- **THEN** system maintains conversation context across turns
- **AND** system loads previous messages correctly
- **AND** system handles tool calls and results in conversation history

#### Scenario: Test error recovery flow
- **WHEN** tool execution fails with dependency error
- **THEN** system provides structured error information to LLM
- **AND** LLM receives tool execution summary
- **AND** LLM makes intelligent retry decision based on context
- **AND** system retries operation after dependencies are resolved

#### Scenario: Test conversation history persistence
- **WHEN** conversation spans multiple sessions
- **THEN** system saves conversation history to session file
- **AND** system loads conversation history on next session
- **AND** system maintains full context including tool calls

### Requirement: Integration test coverage for Skills integration

The system SHALL have integration tests for Skills loading, matching, and integration with prompts.

#### Scenario: Test Skills loading and matching
- **WHEN** user submits query
- **THEN** system loads Skills metadata on startup
- **AND** system matches relevant Skills to query
- **AND** system loads matched Skills content
- **AND** system includes Skills content in prompt

#### Scenario: Test Skills with various query types
- **WHEN** user submits different query types (database, command, file)
- **THEN** system matches appropriate Skills for each query type
- **AND** system filters out irrelevant Skills
- **AND** system handles ambiguous queries correctly

#### Scenario: Test Skills precision filtering
- **WHEN** query matches Skills with low precision scores
- **THEN** system filters out low-precision matches
- **AND** system only includes high-precision Skills
- **AND** system avoids loading irrelevant Skills

### Requirement: Integration test coverage for prompt management

The system SHALL have integration tests for prompt management with long conversations and compression.

#### Scenario: Test prompt compression with long conversations
- **WHEN** conversation history exceeds token limit
- **THEN** system compresses conversation history using LLM
- **AND** system preserves important context
- **AND** system evicts low-priority Skills if needed
- **AND** system maintains prompt within token limits

#### Scenario: Test prompt building with Skills
- **WHEN** prompt is built with multiple matched Skills
- **THEN** system includes Skills content in system prompt section
- **AND** system orders Skills by priority
- **AND** system formats prompt correctly

### Requirement: Integration test coverage for risk-based confirmation workflows

The system SHALL have integration tests for risk assessment and confirmation workflows.

#### Scenario: Test low-risk operation automatic execution
- **WHEN** user requests low-risk operation (SELECT, ls, file read)
- **THEN** system executes operation automatically without confirmation
- **AND** system displays results directly to user
- **AND** system logs risk assessment decision

#### Scenario: Test high-risk operation confirmation
- **WHEN** user requests high-risk operation (DROP, rm, file write)
- **THEN** system displays operation details
- **AND** system prompts user for confirmation
- **AND** system executes after user confirms
- **AND** system cancels operation if user denies

#### Scenario: Test LLM-provided risk level
- **WHEN** LLM provides risk_level in tool call
- **THEN** system respects LLM's risk assessment
- **AND** system executes automatically for low risk
- **AND** system requires confirmation for high risk

#### Scenario: Test unknown operation handling
- **WHEN** LLM calls tool for unknown operation (not in whitelist)
- **THEN** system requires confirmation by default
- **AND** system allows LLM to set risk_level="high" for unknown operations
- **AND** system handles LLM uncertainty appropriately

### Requirement: Integration test coverage for tool execution workflows

The system SHALL have integration tests for complete tool execution workflows including error handling.

#### Scenario: Test SQL tool execution workflow
- **WHEN** user requests SQL query execution
- **THEN** system checks risk level
- **AND** system executes automatically for low-risk queries
- **AND** system requires confirmation for high-risk queries
- **AND** system displays results in table format
- **AND** system avoids redundant output

#### Scenario: Test command tool execution workflow
- **WHEN** user requests command execution
- **THEN** system checks risk level
- **AND** system executes command with timeout
- **AND** system displays output to user
- **AND** system handles command errors

#### Scenario: Test file tool execution workflow
- **WHEN** user requests file operation
- **THEN** system checks operation type and risk level
- **AND** system executes read operations automatically
- **AND** system requires confirmation for write operations
- **AND** system handles file errors

### Requirement: Integration test coverage for cross-platform compatibility

The system SHALL have integration tests to verify cross-platform compatibility.

#### Scenario: Test on macOS
- **WHEN** system runs on macOS
- **THEN** all functionality works correctly
- **AND** file paths are handled correctly
- **AND** command execution works with macOS shell

#### Scenario: Test on Linux
- **WHEN** system runs on Linux
- **THEN** all functionality works correctly
- **AND** file paths are handled correctly
- **AND** command execution works with Linux shell

#### Scenario: Test on Windows
- **WHEN** system runs on Windows
- **THEN** all functionality works correctly
- **AND** file paths are handled correctly (Windows paths)
- **AND** command execution works with Windows shell
