## Context

The codebase contains Chinese text in several areas:
- Code comments in Go source files (minimal, ~2 files in test files)
- Spec documents in `openspec/specs/` directory
- Change documents in `openspec/changes/` directory (especially archived changes)
- Test file comments and descriptions

The project's primary language is English, and README_CN.md is explicitly maintained for Chinese-speaking users. All other Chinese text should be translated to English to improve internationalization and maintain consistency.

**Current State:**
- Most code comments are already in English
- Some test files contain Chinese comments (e.g., `internal/tool/builtin/command_tool_test.go`, `internal/llm/client_test.go`)
- Archive directory (`openspec/changes/archive/`) contains many Chinese documents (~19 files)
- Active spec files in `openspec/specs/` may contain Chinese text

**Constraints:**
- Must preserve technical accuracy
- Must maintain code functionality (comments only, no code changes)
- README_CN.md must remain unchanged
- Translation should be natural and idiomatic English

## Goals / Non-Goals

**Goals:**
- Translate all Chinese comments in Go code files to English
- Translate all Chinese text in spec documents to English
- Translate Chinese text in test files to English
- Translate Chinese text in archived change documents to English
- Ensure translations are accurate and maintain technical meaning
- Use consistent terminology throughout the codebase

**Non-Goals:**
- Translating README_CN.md (explicitly excluded)
- Modifying functional code (only comments and documentation)
- Translating user-facing error messages (if they are meant to be in Chinese)
- Creating automated translation tools (manual translation for quality)
- Translating commit messages (historical records)

## Decisions

### 1. Translation Approach: Manual Review and Translation

**Decision**: Use manual translation with review rather than automated translation tools.

**Rationale**: 
- Automated translation tools (e.g., Google Translate) may produce inaccurate technical translations
- Manual translation ensures technical accuracy and natural English
- Code comments require understanding of context and technical concepts

**Alternatives Considered**:
- Automated translation tools: Faster but less accurate for technical content
- Hybrid approach (automated + manual review): Still requires full manual review, so no time savings

### 2. File Discovery Strategy: Systematic Search

**Decision**: Use grep with Unicode pattern matching to find all Chinese text, then translate systematically.

**Rationale**:
- Ensures no Chinese text is missed
- Provides a complete inventory of files to translate
- Can be verified after translation

**Implementation**:
```bash
# Find Chinese characters in code files
grep -r "[\u4e00-\u9fff]" internal/ --include="*.go"

# Find Chinese characters in spec files
grep -r "[\u4e00-\u9fff]" openspec/ --include="*.md"
```

### 3. Translation Order: Code First, Then Documentation

**Decision**: Translate code comments first, then spec documents, then archived changes.

**Rationale**:
- Code comments are most visible to developers
- Spec documents are actively used
- Archived changes are historical and less critical

**Order**:
1. Code comments in `internal/` directory
2. Test file comments
3. Active spec files in `openspec/specs/`
4. Active change documents in `openspec/changes/` (non-archive)
5. Archived change documents in `openspec/changes/archive/`

### 4. Terminology Consistency

**Decision**: Maintain a glossary of technical terms and use consistent translations.

**Rationale**:
- Ensures consistency across translations
- Prevents confusion from multiple translations of the same term
- Improves readability

**Common Terms**:
- 任务 → Task
- 测试 → Test
- 配置 → Configuration
- 会话 → Session
- 提示词 → Prompt
- 工具 → Tool
- 风险 → Risk
- 确认 → Confirmation
- 执行 → Execute/Execution
- 错误 → Error

### 5. Comment Format Preservation

**Decision**: Preserve comment format and structure, only translate content.

**Rationale**:
- Maintains code readability
- Preserves comment style consistency
- Avoids unnecessary formatting changes

**Examples**:
- `// Task 5.4: 验证用户可以通过 timeout 参数自定义超时时间` 
  → `// Task 5.4: Verify that users can customize timeout via timeout parameter`
- `// 基本测试，验证命令执行功能` 
  → `// Basic test to verify command execution functionality`

### 6. Spec Document Translation: Preserve Structure

**Decision**: Translate spec document content while preserving markdown structure, code blocks, and technical formatting.

**Rationale**:
- Maintains document readability
- Preserves technical accuracy
- Keeps document structure intact

## Risks / Trade-offs

### Risk 1: Loss of Technical Accuracy
**Risk**: Translation may lose subtle technical nuances or context.
**Mitigation**: 
- Review translations with technical context in mind
- Preserve technical terms where appropriate
- Use native English speakers or technical reviewers if available

### Risk 2: Inconsistent Terminology
**Risk**: Same Chinese term translated differently in different places.
**Mitigation**:
- Maintain terminology glossary
- Review translations for consistency
- Use grep to find all occurrences of translated terms

### Risk 3: Missing Chinese Text
**Risk**: Some Chinese text may be missed during translation.
**Mitigation**:
- Use systematic grep search before and after translation
- Verify no Chinese characters remain (except README_CN.md)
- Run final verification pass

### Risk 4: Breaking Formatting
**Risk**: Translation may break markdown formatting or code structure.
**Mitigation**:
- Preserve all formatting characters
- Test markdown rendering after translation
- Review code comments don't break line length limits

### Trade-off: Time vs. Quality
**Trade-off**: Manual translation takes longer but ensures quality.
**Decision**: Prioritize quality over speed for technical documentation.

## Migration Plan

### Phase 1: Discovery and Inventory
1. Run grep to find all files containing Chinese text
2. Create inventory list of files to translate
3. Categorize by priority (code > active docs > archived docs)

### Phase 2: Code Comments Translation
1. Translate comments in `internal/` Go source files
2. Translate test file comments
3. Review and verify translations
4. Commit changes

### Phase 3: Active Documentation Translation
1. Translate spec files in `openspec/specs/`
2. Translate active change documents in `openspec/changes/` (non-archive)
3. Review and verify translations
4. Commit changes

### Phase 4: Archived Documentation Translation
1. Translate archived change documents in `openspec/changes/archive/`
2. Review and verify translations
3. Commit changes

### Phase 5: Verification
1. Run final grep to verify no Chinese text remains (except README_CN.md)
2. Review translations for consistency
3. Update terminology glossary if needed
4. Final commit

### Rollback Strategy
- All changes are in comments/documentation only
- Git history preserves original Chinese text
- Can revert individual files if translation issues found
- No functional code changes, so no risk to application behavior

## Open Questions

1. **Should archived change documents be translated?**
   - Decision: Yes, for consistency, but lower priority
   - Rationale: Historical documents still benefit from English translation for future reference

2. **How to handle mixed Chinese/English content?**
   - Decision: Translate Chinese portions, preserve English portions
   - Rationale: Maintains existing English content while translating Chinese

3. **Should commit messages be translated?**
   - Decision: No, historical records should remain as-is
   - Rationale: Commit messages are historical artifacts

4. **What about user-facing strings in code?**
   - Decision: Review case-by-case; if meant for Chinese users, keep Chinese; if meant for developers, translate
   - Rationale: User-facing strings may be intentionally in Chinese
