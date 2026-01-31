## 1. Error Information Extraction

- [x] 1.1 Create error extraction utility functions in `internal/tool/error_extractor.go`:
  - `ExtractErrorInfo(error error) ErrorInfo` - Main extraction function
  - `extractForeignKeyConstraint(error string) (affectedResource, dependency, constraint string, found bool)`
  - `extractSyntaxError(error string) (affectedResource string, found bool)`
  - `extractPermissionError(error string) (affectedResource string, found bool)`
  - `extractResourceNotFound(error string) (affectedResource string, found bool)`
  - `extractResourceExists(error string) (affectedResource string, found bool)`
  - `categorizeErrorType(error string) string` - Returns error type category

- [x] 1.2 Define `ErrorInfo` struct:
  - `ErrorCode string` - Database error code if available
  - `ErrorType string` - Categorized error type
  - `AffectedResources []string` - List of affected resources
  - `Dependencies []string` - List of dependencies
  - `SuggestedActions []string` - Suggested actions

- [x] 1.3 Implement pattern matching for common error patterns:
  - Foreign key constraint errors (MySQL, PostgreSQL, SeekDB patterns)
  - Syntax errors
  - Permission errors
  - Resource not found errors
  - Resource exists errors

- [ ] 1.4 Add unit tests for error extraction:
  - Test foreign key constraint extraction
  - Test syntax error extraction
  - Test permission error extraction
  - Test resource not found extraction
  - Test resource exists extraction
  - Test unknown error handling

## 2. Structured Error Response Enhancement

- [x] 2.1 Update `internal/sql/tool_handler.go` `ExecuteTool()` method:
  - Extract error information using `ExtractErrorInfo()` when SQL execution fails
  - Build structured error JSON with all available fields
  - Maintain backward compatibility (keep original `error` field)

- [x] 2.2 Update `internal/tool/sql_tool.go` `ExecuteSQL()` function:
  - Pass error to extraction function (extraction happens in tool_handler.go)
  - Return structured error information if available

- [x] 2.3 Update `internal/tool/builtin/command_tool.go` error handling:
  - Extract error information from command execution errors (extraction happens in tool_handler.go)
  - Include structured error fields in error JSON responses

- [ ] 2.4 Test structured error responses:
  - Test with foreign key constraint errors
  - Test with syntax errors
  - Test with permission errors
  - Test with unknown errors
  - Verify backward compatibility

## 3. Tool Execution Tracking

- [x] 3.1 Add tool execution tracking to `ToolHandler` struct in `internal/sql/tool_handler.go`:
  - Add `recentExecutions []ToolExecution` field
  - Define `ToolExecution` struct:
    - `Tool string` - Tool name
    - `Arguments map[string]interface{}` - Tool arguments
    - `Status string` - "success" or "error"
    - `ErrorInfo *ErrorInfo` - Error information if failed
    - `StateChanges []string` - List of state changes

- [x] 3.2 Implement `trackToolExecution()` method:
  - Record tool execution with arguments and status
  - Detect state changes (DDL operations, resource creation/deletion)
  - Maintain last N executions (default: 5)

- [x] 3.3 Implement `resetExecutionTracking()` method:
  - Clear execution tracking when new user query starts
  - Called at beginning of `HandleToolCallLoop()`

- [x] 3.4 Update `HandleToolCallLoop()` to track executions:
  - Call `trackToolExecution()` after each tool execution
  - Pass execution details (tool name, args, status, error info)

## 4. Tool Execution Summary Generation

- [x] 4.1 Implement `generateExecutionSummary()` method in `ToolHandler`:
  - Generate summary of last 3-5 tool executions
  - Include execution status (SUCCESS/FAILED)
  - Include error type and affected resources for failures
  - Highlight state changes explicitly

- [x] 4.2 Format summary in LLM-readable format:
  - Use structured format with clear sections
  - Mark as `<TOOL_EXECUTION_SUMMARY>`
  - Include list of recent executions
  - Include state changes section
  - Include dependencies resolved section (if applicable)

- [x] 4.3 Integrate summary into LLM request:
  - Generate summary before each LLM call in `HandleToolCallLoop()`
  - Include summary as separate section before user query
  - Add to messages array before sending to LLM

- [ ] 4.4 Test summary generation:
  - Test with multiple tool executions
  - Test with mixed success/failure outcomes
  - Test state change detection
  - Test summary format and readability

## 5. State Change Detection

- [x] 5.1 Implement state change detection logic:
  - Detect DDL operations (CREATE/DROP/ALTER) as state changes
  - Detect resource creation (tables, columns)
  - Detect resource deletion (tables, columns)
  - Detect dependency resolution (dependent resource deleted)

- [x] 5.2 Update `trackToolExecution()` to detect state changes:
  - Parse SQL arguments for DDL operations
  - Extract resource names from CREATE/DROP/ALTER statements
  - Track which resources were created/deleted

- [x] 5.3 Track dependency resolution:
  - When resource is deleted, check if it was blocking previous operations
  - Mark as dependency resolution if blocking resource is deleted

- [ ] 5.4 Test state change detection:
  - Test DDL operation detection
  - Test resource creation tracking
  - Test resource deletion tracking
  - Test dependency resolution detection

## 6. Prompt Enhancement

- [x] 6.1 Update `internal/prompt/loader.go` `getBuiltInPromptStrings()`:
  - Add `<ERROR_HANDLING>` section to `database-base.md` prompt
  - Add error analysis and retry guidance instructions
  - Include example of intelligent retry scenario

- [x] 6.2 Add error handling guidance to `common.md` prompt:
  - Add instructions for analyzing structured error information
  - Add guidance for understanding state changes
  - Add instructions for intelligent retry decisions

- [x] 6.3 Update prompt templates with error handling sections:
  - Include instructions for analyzing error structure
  - Include instructions for reviewing tool execution summary
  - Include instructions for automatic retry when dependencies resolved
  - Include example scenarios

- [ ] 6.4 Test prompt updates:
  - Verify prompts include error handling sections
  - Verify prompts are clear and actionable
  - Test LLM behavior with updated prompts

## 7. Integration and Testing

- [ ] 7.1 Integrate all components in `HandleToolCallLoop()`:
  - Ensure error extraction is called for all tool errors
  - Ensure execution tracking happens for all tool calls
  - Ensure summary generation happens before each LLM call
  - Ensure summary is included in LLM request

- [ ] 7.2 Test end-to-end error handling flow:
  - Test foreign key constraint scenario (drop tables with dependencies)
  - Test LLM automatic retry after dependency resolution
  - Test with multiple failed operations
  - Test with mixed success/failure outcomes

- [ ] 7.3 Test backward compatibility:
  - Verify existing error handling still works
  - Verify LLM can handle both old and new error formats
  - Verify no breaking changes to tool response format

- [ ] 7.4 Performance testing:
  - Measure token usage impact of summaries
  - Measure performance impact of error extraction
  - Optimize if needed (summary length, extraction efficiency)

## 8. Documentation and Cleanup

- [ ] 8.1 Add code comments explaining error extraction logic
- [ ] 8.2 Add code comments explaining summary generation
- [ ] 8.3 Add code comments explaining state change detection
- [ ] 8.4 Update any relevant documentation about error handling
- [ ] 8.5 Verify all tests pass
- [ ] 8.6 Code review and cleanup
