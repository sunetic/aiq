## 1. Version Management Module

- [x] 1.1 Create `internal/version/version.go` with package structure
- [x] 1.2 Add `Version` and `CommitID` variables (default values: "dev", "unknown")
- [x] 1.3 Implement `GetVersion()` function with three-tier fallback:
  - Check build-time injected `Version` variable
  - Fallback to `git describe --tags --always` if git available
  - Return "dev" as default
- [x] 1.4 Implement `GetCommitID()` function with three-tier fallback:
  - Check build-time injected `CommitID` variable
  - Fallback to `git rev-parse HEAD` if git available
  - Return "unknown" as default
- [x] 1.5 Implement `GetVersionInfo()` function to format version string: "aiq <version> (commit: <commit-id>)"
- [x] 1.6 Add error handling and timeout for git commands (fallback to defaults on failure)
- [ ] 1.7 Add unit tests for version detection functions

## 2. Version Command Support

- [x] 2.1 Add `-v` and `--version` flag parsing in `cmd/aiq/main.go` (before any initialization)
- [x] 2.2 Implement version command handler that:
  - Calls `version.GetVersionInfo()`
  - Prints version information to stdout
  - Exits with code 0 immediately
- [ ] 2.3 Test version command works without initializing config or prompts
- [ ] 2.4 Update help text/documentation for version flags

## 3. CI/CD Build Configuration

- [x] 3.1 Update `.github/workflows/go.yml` build step to inject version via `-ldflags`:
  - `-X github.com/aiq/aiq/internal/version.Version=${{ github.ref_name }}`
  - `-X github.com/aiq/aiq/internal/version.CommitID=${{ github.sha }}`
- [ ] 3.2 Test build process injects version correctly in CI/CD
- [x] 3.3 Update local build documentation/scripts if needed (created Makefile and LOCAL_TESTING.md)

## 4. Content Hash Calculation

- [x] 4.1 Add `hashContent()` function in `internal/prompt/loader.go`:
  - Calculate SHA256 hash of prompt content
  - Return hash string
- [x] 4.2 Add `getBuiltInPromptHashes()` function:
  - Calculate content hashes for all built-in prompt strings
  - Return map of filename → content hash
- [x] 4.3 Add `getUserPromptHashes()` function:
  - Read prompt files from `~/.aiq/prompts/` directory
  - Calculate content hash for each file
  - Return map of filename → content hash
- [x] 4.4 Add `checkPromptContentMismatch()` function:
  - Compare built-in content hashes with user file content hashes
  - Return list of files that have been modified
- [ ] 4.5 Add unit tests for frontmatter parsing functions

## 5. Version Choice Storage

- [x] 5.1 Create `internal/config/version_choices.go` file
- [x] 5.2 Define `VersionChoices` struct:
  - `Choices map[string]string` (version → "overwrite"/"keep")
- [x] 5.3 Implement `GetVersionChoicesFilePath()` function:
  - Returns `~/.aiq/config/prompt-version-choices.yaml`
- [x] 5.4 Implement `LoadVersionChoices()` function:
  - Reads YAML file if exists
  - Returns empty choices map if file missing/corrupted (non-fatal)
- [x] 5.5 Implement `SaveVersionChoices()` function:
  - Writes choices to YAML file
  - Creates config directory if needed
- [x] 5.6 Implement `GetChoiceForVersion()` function:
  - Checks if choice exists for given version
  - Returns choice ("overwrite"/"keep") or empty string if not found
- [x] 5.7 Implement `SetChoiceForVersion()` function:
  - Stores choice for given version
  - Saves to file immediately
- [ ] 5.8 Add unit tests for version choices storage

## 6. Content Detection and Comparison Logic

- [x] 6.1 Add `checkPromptContentMismatch()` function in `internal/prompt/loader.go`:
  - Gets built-in prompt content hashes
  - Gets user prompt file content hashes
  - Compares hashes for each file
  - Returns list of files that have been modified
- [x] 6.2 Add `hasContentMismatch()` helper function:
  - Checks if any prompt files have been modified (hash mismatch)
  - Returns boolean
- [x] 6.3 Integrate content check into `NewLoader()`:
  - Call `checkPromptContentMismatch()` before `initializeDefaults()`
  - Store modification results for later use
- [ ] 6.4 Add unit tests for version comparison logic

## 7. User Prompt UI

- [x] 7.1 Add `PromptForVersionUpgrade()` function in `internal/prompt/loader.go` or `internal/ui/`:
  - Displays message about version mismatch
  - Asks user: "Overwrite prompt files with new versions? (y/n)"
  - Returns user choice ("overwrite" or "keep")
- [x] 7.2 Add instruction message display:
  - When user chooses "keep", show message:
    "To upgrade prompt files later, delete files in ~/.aiq/prompts/ and restart aiq"
- [x] 7.3 Integrate user prompt into version check flow:
  - Check if user has already made choice for current app version
  - If no choice, prompt user
  - Store choice for current app version
- [x] 7.4 Handle user cancellation (Ctrl+C) gracefully
- [ ] 7.5 Add unit tests for user prompt flow (mock user input)

## 8. Integration and Flow Control

- [x] 8.1 Modify `NewLoader()` initialization flow:
  - Get current application version
  - Check for version mismatches
  - If mismatch and no stored choice, prompt user
  - Apply user choice (overwrite or keep)
  - Then proceed with normal initialization
- [x] 8.2 Update `initializeDefaults()` to respect user choice:
  - If user chose "overwrite", overwrite existing files
  - If user chose "keep", skip file creation if exists
- [x] 8.3 Ensure version check happens before any file operations
- [x] 8.4 Add error handling for version check failures (non-fatal, continue with defaults)

## 9. Testing

- [ ] 9.1 Test new installation scenario (no prompt files exist)
- [ ] 9.2 Test upgrade scenario (prompt files exist with modified content)
- [ ] 9.3 Test user chooses "overwrite" (files get updated)
- [ ] 9.4 Test user chooses "keep" (files preserved, instruction shown)
- [ ] 9.5 Test version choice persistence (same version doesn't prompt again)
- [ ] 9.6 Test different application versions (new version prompts again)
- [x] 9.7 Test version command (`-v` / `--version`) in various scenarios
- [x] 9.8 Test git fallback when build-time injection unavailable
- [ ] 9.9 Test error scenarios (corrupted files, missing git, etc.)

## 10. Documentation and Cleanup

- [x] 10.1 Update README with version command usage
- [x] 10.2 Document prompt file upgrade process (created TESTING.md)
- [ ] 10.3 Add code comments explaining version detection logic
- [ ] 10.4 Update any relevant documentation about prompt customization
- [ ] 10.5 Verify all tests pass
- [ ] 10.6 Code review and cleanup
