## ADDED Requirements

### Requirement: Test comment translation
The system SHALL translate all Chinese comments and descriptions in test files to English, preserving test functionality and readability.

#### Scenario: Translate test function comments
- **WHEN** a test file contains Chinese comments describing test functions
- **THEN** all Chinese comments SHALL be translated to English
- **AND** test function names SHALL remain unchanged
- **AND** test logic SHALL remain unchanged

#### Scenario: Translate test scenario descriptions
- **WHEN** a test contains Chinese descriptions (e.g., in `t.Run()` calls or comments)
- **THEN** all Chinese descriptions SHALL be translated to English
- **AND** test scenario names SHALL be translated
- **AND** test assertions and error messages SHALL remain in English (or be translated if in Chinese)

#### Scenario: Preserve test structure
- **WHEN** translating test comments
- **THEN** test function structure SHALL be preserved
- **AND** test table structures (for table-driven tests) SHALL be maintained
- **AND** test helper functions SHALL remain unchanged

#### Scenario: Translate task references in tests
- **WHEN** a test comment contains task references (e.g., "Task 5.4: ...")
- **THEN** the task reference SHALL be preserved
- **AND** only the Chinese description SHALL be translated
- **Example**: `// Task 5.4: 验证用户可以通过 timeout 参数自定义超时时间` → `// Task 5.4: Verify that users can customize timeout via timeout parameter`

#### Scenario: Maintain test readability
- **WHEN** translating test comments
- **THEN** translated comments SHALL be clear and descriptive
- **AND** test intent SHALL be preserved
- **AND** comments SHALL help developers understand test purpose

#### Scenario: Handle test error messages
- **WHEN** test error messages contain Chinese text
- **THEN** error messages SHALL be translated to English
- **AND** error message format SHALL be preserved
- **AND** technical accuracy SHALL be maintained
