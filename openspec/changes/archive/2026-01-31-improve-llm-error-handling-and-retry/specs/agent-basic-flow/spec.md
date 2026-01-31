## ADDED Requirements

### Requirement: Agent Basic Flow - Tool Execution and LLM Decision Making

The system SHALL implement a fundamental Agent flow where LLM makes all decisions about task classification (exploratory vs definitive) and continuation (when to finish vs continue), while the code layer simply executes tools and returns all results to LLM for decision-making.

**Key Principle**: Exploratory vs Definitive is a **conceptual distinction made by LLM during planning**, not a code-level classification based on tool types or SQL keywords. The code layer should not attempt to classify tasks - it should trust LLM's decision-making capability.

#### Scenario: User Input Processing
- **WHEN** user submits a request
- **THEN** system sends user request to LLM
- **AND** LLM analyzes the request and decides whether it's exploratory or definitive
- **AND** LLM decides what tools to call based on its understanding

#### Scenario: Exploratory Task Classification (LLM Decision)
- **WHEN** user request requires information gathering before deciding next steps (e.g., "analyze sales trends", "investigate performance issues")
- **THEN** LLM classifies this as exploratory task
- **AND** LLM calls tools to gather information
- **AND** system returns all tool results (success or failure) to LLM for next step planning
- **AND** LLM continues calling tools or analyzing results until task is complete
- **AND** LLM returns finish_reason="stop" when analysis is complete

#### Scenario: Definitive Task Classification (LLM Decision)
- **WHEN** user request is clear and complete (e.g., "show tables", "drop table X", "list files in /tmp")
- **THEN** LLM classifies this as definitive task
- **AND** LLM calls appropriate tools directly
- **AND** system returns all tool results to LLM
- **AND** if tools succeed, LLM returns finish_reason="stop" to complete the task
- **AND** if tools fail, LLM analyzes errors and decides retry/alternative approach

#### Scenario: Tool Execution Result Handling
- **WHEN** any tool is executed (regardless of tool type: execute_sql, execute_command, http_request, file_operations, etc.)
- **THEN** system returns tool result to LLM (both success and failure cases)
- **AND** system does NOT make code-level decisions about whether result should be returned to LLM
- **AND** system does NOT classify tools as exploratory/definitive based on SQL keywords or tool types
- **AND** LLM receives all tool results and decides next action based on task context

#### Scenario: Failed Tool Execution Handling
- **WHEN** any tool execution fails in current round
- **THEN** system MUST return error result to LLM
- **AND** LLM MUST analyze the error and decide next action (retry, alternative approach, or ask user)
- **AND** LLM MUST NOT return finish_reason="stop" until errors are handled or user input is needed
- **AND** system continues loop to let LLM process error and decide retry/alternative

#### Scenario: Successful Tool Execution Handling
- **WHEN** tool execution succeeds
- **THEN** system returns success result to LLM
- **AND** LLM decides whether task is complete based on task classification:
  - **Definitive task**: If task is complete, LLM returns finish_reason="stop" with no tool_calls
  - **Exploratory task**: LLM continues processing results, calls more tools if needed, or returns finish_reason="stop" when analysis complete
- **AND** system does NOT make code-level decisions about whether to continue or finish

#### Scenario: LLM Decision on Task Completion
- **WHEN** LLM receives tool execution results
- **THEN** LLM decides whether to:
  - Continue: Return tool_calls for next step (if more information needed or errors to handle)
  - Finish: Return finish_reason="stop" with no tool_calls (if task is complete)
  - Respond: Return content with finish_reason="stop" (if user interaction needed)
- **AND** system respects LLM's decision and acts accordingly
- **AND** system does NOT override LLM's decision based on code-level heuristics

#### Scenario: Multi-Tool Execution in Single Round
- **WHEN** LLM calls multiple tools in a single round (e.g., dropping multiple tables)
- **THEN** system executes all tools sequentially
- **AND** system returns all tool results to LLM (both successes and failures)
- **AND** if any tool fails, LLM analyzes all results and decides retry strategy
- **AND** LLM may retry failed operations if dependencies are resolved (visible in tool execution summary)
- **AND** LLM returns finish_reason="stop" only when all operations complete successfully or user input is needed

#### Scenario: State Change Awareness
- **WHEN** tool execution causes state changes (e.g., table dropped, file created)
- **THEN** system tracks state changes in tool execution summary
- **AND** system includes summary in next LLM call
- **AND** LLM uses summary to understand current system state
- **AND** LLM makes retry decisions based on state changes (e.g., retry dropping table if dependent table was dropped)

### Requirement: Code Layer Simplicity

The code layer SHALL NOT attempt to classify tasks or make decisions about continuation. It should only:
- Execute tools as requested by LLM
- Return all tool results to LLM
- Respect LLM's finish_reason and tool_calls decisions
- Track tool executions for summary generation

#### Scenario: No Code-Level Task Classification
- **WHEN** system processes tool execution results
- **THEN** system does NOT classify tools as exploratory/definitive based on:
  - SQL keywords (SELECT, SHOW, DROP, CREATE, etc.)
  - Tool types (execute_sql, execute_command, etc.)
  - Execution success/failure status
- **AND** system returns all results to LLM regardless of classification

#### Scenario: No Code-Level Continuation Decisions
- **WHEN** all tools in a round execute successfully
- **THEN** system does NOT make code-level decision to exit early
- **AND** system continues loop to let LLM decide whether to finish or continue
- **AND** system only exits when LLM explicitly returns finish_reason="stop" with no tool_calls

#### Scenario: Trust LLM Decision-Making
- **WHEN** LLM returns finish_reason="stop" with no tool_calls
- **THEN** system exits loop and returns LLM's response
- **AND** system does NOT second-guess LLM's decision
- **AND** system does NOT force continuation even if code-level heuristics suggest otherwise

### Requirement: Prompt Guidance for Agent Flow

The system SHALL provide clear prompt guidance to LLM about Agent flow, including:
- Task classification (exploratory vs definitive)
- When to continue vs finish
- How to handle tool execution results
- Examples of different scenarios

#### Scenario: Prompt Includes Agent Flow Section
- **WHEN** system loads prompts for LLM
- **THEN** prompts include `<AGENT_FLOW>` section with:
  - Task classification guidance (definitive vs exploratory)
  - Tool execution result handling rules
  - Decision-making principles
  - Examples for different scenarios
- **AND** guidance applies to all tool types (not just SQL)
- **AND** guidance emphasizes LLM's decision-making responsibility

#### Scenario: Prompt Examples Cover All Tool Types
- **WHEN** system provides Agent flow examples in prompts
- **THEN** examples include:
  - SQL operations (execute_sql)
  - Command execution (execute_command)
  - File operations (file_operations)
  - HTTP requests (http_request)
- **AND** examples demonstrate both definitive and exploratory scenarios
- **AND** examples show proper finish_reason="stop" usage

## Rationale

This requirement ensures that:
1. **LLM makes all strategic decisions**: Task classification and continuation decisions are made by LLM based on context, not by code-level heuristics
2. **Code layer remains simple**: Code only executes tools and returns results, without attempting to classify or decide
3. **Consistent behavior across all tools**: Same flow applies to SQL, commands, files, HTTP requests, etc.
4. **Better error handling**: LLM can intelligently retry operations when state changes make retry safe
5. **Flexible task handling**: LLM can adapt to different task types without code changes

## Impact

- **Affected code**:
  - `internal/sql/tool_handler.go`: Remove code-level exploratory/definitive classification, return all tool results to LLM
  - `internal/prompt/loader.go`: Add `<AGENT_FLOW>` section to prompts with clear guidance
  - All tool execution handlers: Ensure all results are returned to LLM

- **User experience**: More intelligent and consistent behavior across all tool types, better error recovery
