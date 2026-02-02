## Why

The codebase currently contains Chinese comments in code files and Chinese text in spec documents. To improve internationalization and maintain consistency with the English codebase, all Chinese text (except README_CN.md) should be translated to English. This ensures that developers who don't read Chinese can understand the codebase, and maintains consistency with the project's primary language.

## What Changes

- Translate all Chinese comments in Go code files to English
- Translate all Chinese text in spec documents (`openspec/specs/` and `openspec/changes/`) to English
- Translate Chinese text in test files (comments, test descriptions) to English
- Translate Chinese text in documentation files (except README_CN.md) to English
- Keep README_CN.md unchanged as it is specifically for Chinese-speaking users
- Ensure technical accuracy is maintained during translation

**BREAKING**: None - this is a documentation/comment translation change, no functional changes.

## Capabilities

### New Capabilities
- `code-comment-translation`: Translate all Chinese comments in Go source files to English
- `spec-document-translation`: Translate all Chinese text in spec documents to English
- `test-comment-translation`: Translate Chinese comments and descriptions in test files to English

### Modified Capabilities
None - this change only affects documentation and comments, not functional requirements.

## Impact

**Affected Areas:**
- All Go source files in `internal/` directory (code comments)
- All spec files in `openspec/specs/` directory
- All change documents in `openspec/changes/` directory (proposals, designs, specs, tasks)
- Test files in `internal/*/` directories (test comments and descriptions)
- Documentation files (except README_CN.md)

**Files to Translate:**
- Code comments in Go files (estimated: ~50-100 comments)
- Spec documents (estimated: ~20-30 files)
- Test file comments (estimated: ~10-20 files)
- Change documents in archive (estimated: ~30-50 files)

**No Impact On:**
- README_CN.md (explicitly excluded)
- Functional code logic
- API contracts
- Database schemas
- Configuration formats
