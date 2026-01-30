## Context

### Current State
- Prompt files are stored in `~/.aiq/prompts/` directory for user customization
- Default prompt files are created on first run via `prompt.NewLoader()` if they don't exist
- Prompt files use YAML frontmatter format with `version` field (e.g., `version: "1.1"`)
- Current implementation in `internal/prompt/loader.go`:
  - `initializeDefaults()`: Creates default prompt files if missing
  - `parsePromptFile()`: Parses markdown files but only extracts body content, not frontmatter metadata
  - No content modification detection or upgrade mechanism exists

### Problem
When users modify prompt files, they typically modify the content directly without updating the version number. When AIQ is upgraded, the system doesn't know whether to:
- Use user-modified prompt files (may be outdated or incompatible)
- Overwrite with new built-in versions (may lose user customizations)

Comparing version numbers is unreliable because users don't typically update version numbers when modifying content. We need to detect actual content modifications.

### Constraints
- Must preserve user customizations when desired
- Must allow users to get latest prompt improvements
- Should not interrupt normal workflow unnecessarily
- Must work in both development (with git) and production (binary) environments

## Goals / Non-Goals

**Goals:**
- Detect content modifications in user prompt files (by comparing content hashes)
- Prompt user once per application version when modifications detected
- Store user choice to avoid repeated prompts
- Support version command (`-v` / `--version`) for debugging
- Automatically detect version from build-time injection or git

**Non-Goals:**
- Automatic prompt file migration or transformation
- Version comparison logic (simple string comparison is sufficient)
- Prompt file diff visualization
- Rollback mechanism for prompt files
- Version history tracking beyond current choice

## Decisions

### 1. Version Detection Strategy: Multi-level Fallback

**Decision**: Use three-tier fallback for version detection:
1. Build-time injection via `-ldflags` (production builds)
2. Runtime git commands (`git describe`, `git rev-parse`) (development)
3. Default values ("dev", "unknown")

**Rationale**:
- Production binaries should have version injected at build time (no git dependency)
- Development builds can use git commands for convenience
- Default values ensure system always works, even without git or build flags

**Alternatives Considered**:
- **Hardcoded version**: Rejected - requires manual updates on every release
- **Git-only approach**: Rejected - production binaries don't have git
- **Config file version**: Rejected - adds complexity, version should come from build

### 2. Version Storage: Separate YAML File

**Decision**: Store version choices in `~/.aiq/config/prompt-version-choices.yaml` separate from main config.

**Rationale**:
- Keeps main config file focused on LLM/source configuration
- Allows independent management of version choices
- Easy to reset by deleting single file
- YAML format matches existing config pattern

**Alternatives Considered**:
- **Store in main config.yaml**: Rejected - mixes concerns, harder to reset
- **JSON format**: Rejected - project uses YAML consistently
- **Database**: Rejected - overkill for simple key-value storage

### 3. Content Hash Comparison: Detect Modifications by Hash

**Decision**: Calculate SHA256 hash of prompt content to detect if user has modified files, instead of comparing version numbers.

**Rationale**:
- Users modify prompt content directly without updating version numbers
- Content hash comparison accurately detects any modifications
- More reliable than version number comparison
- SHA256 is fast and provides good collision resistance

**Alternatives Considered**:
- **Version number comparison**: Rejected - users don't update version numbers when modifying content
- **File modification time**: Rejected - unreliable, can be affected by system time changes
- **Diff comparison**: Rejected - too slow and complex for startup check

### 4. User Prompt Timing: Before File Operations

**Decision**: Check version and prompt user in `NewLoader()` before `initializeDefaults()` creates/overwrites files.

**Rationale**:
- Must get user consent before overwriting files
- Early check prevents unnecessary file operations
- Clear separation: check → prompt → act

**Alternatives Considered**:
- **After file creation**: Rejected - would overwrite user files before asking
- **Lazy check on first use**: Rejected - inconsistent behavior, harder to reason about

### 5. Version Command: Early Exit

**Decision**: Handle `-v` / `--version` flags in `main.go` before any initialization, exit immediately after printing.

**Rationale**:
- Version command should be fast and non-intrusive
- No need to initialize config or prompts for version display
- Matches standard CLI tool behavior

**Alternatives Considered**:
- **After initialization**: Rejected - slower, unnecessary overhead
- **In CLI menu**: Rejected - requires entering interactive mode

### 6. Version Choice Storage Format

**Decision**: Use YAML structure:
```yaml
choices:
  "v1.0.0": "overwrite"  # or "keep"
  "v1.1.0": "keep"
```

**Rationale**:
- Simple key-value mapping (version → choice)
- Easy to query: check if version exists in choices map
- Human-readable for debugging
- Extensible if needed later

**Alternatives Considered**:
- **Single global choice**: Rejected - doesn't track per-version choices
- **Timestamp-based**: Rejected - unnecessary complexity
- **Boolean flag**: Rejected - less clear than explicit "overwrite"/"keep"

## Risks / Trade-offs

### Risk: Git Commands Fail in Production
**Mitigation**: 
- Production builds use `-ldflags` injection (no git dependency)
- Fallback to default values ensures system still works
- Log warnings if git commands fail (non-fatal)

### Risk: Hash Collision (Very Low Probability)
**Mitigation**:
- SHA256 provides excellent collision resistance (2^256 possible values)
- Probability of collision is astronomically low for prompt files
- If collision occurs, worst case is false positive (prompts user unnecessarily)

### Risk: User Chooses "Keep" but Wants Upgrade Later
**Mitigation**:
- Display clear instruction message on "keep" choice
- User can manually delete `~/.aiq/prompts/` files to trigger rebuild
- Document in help/README

### Risk: Version Choice File Corruption
**Mitigation**:
- Treat missing/corrupted file as "no choice made" (safe default)
- YAML parsing errors are non-fatal (fallback to prompting)
- File is small and can be manually edited/deleted

### Risk: Performance Impact from Git Commands
**Mitigation**:
- Git commands only run if build-time injection unavailable
- Cache results in memory (version doesn't change during runtime)
- Timeout on git commands (fallback to defaults if slow)

### Trade-off: Prompting User vs. Silent Behavior
**Decision**: Prompt user (explicit choice)
**Rationale**: 
- User customizations are valuable, shouldn't be silently overwritten
- One-time prompt per version is acceptable UX trade-off
- Clear instruction message helps users understand options

## Migration Plan

### Phase 1: Add Version Management
1. Create `internal/version/version.go` with version detection functions
2. Add `-v` / `--version` flag handling in `cmd/aiq/main.go`
3. Update CI/CD workflow to inject version via `-ldflags`

### Phase 2: Add Content Hash Calculation
1. Add content hash calculation function (SHA256)
2. Calculate hashes for built-in prompt strings
3. Calculate hashes for user prompt files
4. Compare hashes to detect modifications

### Phase 3: Add Version Detection Logic
1. Add version comparison logic in `prompt.NewLoader()`
2. Create version choice storage in `internal/config/`
3. Add user prompt UI in `internal/ui/` (reuse existing prompt functions)

### Phase 4: Integration
1. Integrate version check into `NewLoader()` initialization flow
2. Test with various scenarios (new install, upgrade, user-modified files)
3. Update documentation

### Rollback Strategy
- If issues arise, can disable version check via feature flag
- Version choice file can be deleted to reset behavior
- No database migrations or irreversible changes

## Open Questions

1. **Should we support partial upgrades?** (e.g., upgrade some files but not others)
   - **Current decision**: No - all-or-nothing approach is simpler
   - **Future consideration**: Could add per-file version tracking if needed

2. **How to handle version format changes?** (e.g., "1.1" → "v1.1.0")
   - **Current decision**: Exact string match (simple, predictable)
   - **Future consideration**: Could add normalization if format changes

3. **Should version check be optional/configurable?**
   - **Current decision**: Always check (ensures users get updates)
   - **Future consideration**: Could add config flag to disable if needed

4. **What if user modifies file but version stays same?**
   - **Current decision**: Prompt user (content hash detects modification)
   - **Rationale**: Content hash comparison accurately detects any content changes, regardless of version number
