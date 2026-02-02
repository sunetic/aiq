## ADDED Requirements

### Requirement: Code comments translation
The system SHALL translate all Chinese comments in Go source files to English, preserving technical accuracy and comment formatting.

#### Scenario: Translate Chinese comments in source files
- **WHEN** a Go source file contains Chinese comments (identified by Unicode range \u4e00-\u9fff)
- **THEN** all Chinese comments SHALL be translated to English
- **AND** the translation SHALL preserve technical meaning and accuracy
- **AND** comment formatting (line breaks, indentation) SHALL be preserved
- **AND** code functionality SHALL remain unchanged

#### Scenario: Preserve comment structure
- **WHEN** translating Chinese comments
- **THEN** comment markers (`//`, `/* */`) SHALL be preserved
- **AND** comment position relative to code SHALL be maintained
- **AND** multi-line comment structure SHALL be preserved

#### Scenario: Handle task references in comments
- **WHEN** a comment contains task references (e.g., "Task 5.4: ...")
- **THEN** the task reference SHALL be preserved
- **AND** only the Chinese description SHALL be translated
- **Example**: `// Task 5.4: 验证用户可以通过 timeout 参数自定义超时时间` → `// Task 5.4: Verify that users can customize timeout via timeout parameter`

#### Scenario: Exclude README_CN.md
- **WHEN** processing files for translation
- **THEN** README_CN.md SHALL be excluded from translation
- **AND** Chinese text in README_CN.md SHALL remain unchanged

#### Scenario: Maintain terminology consistency
- **WHEN** translating technical terms
- **THEN** consistent English translations SHALL be used across all files
- **AND** a terminology glossary SHALL be maintained
- **AND** common terms SHALL use standard translations (e.g., 任务 → Task, 测试 → Test, 配置 → Configuration)
