## ADDED Requirements

### Requirement: Prompt file modification detection

The system SHALL detect when user has modified prompt files in `~/.aiq/prompts` by comparing content hashes.

#### Scenario: User modifies prompt file
- **WHEN** user modifies a prompt file in `~/.aiq/prompts`
- **THEN** system detects modification by comparing SHA256 hash of file content with built-in prompt hash
- **AND** system identifies which files have been modified

#### Scenario: No modifications detected
- **WHEN** user prompt files match built-in prompts (same content hash)
- **THEN** system proceeds without prompting user
- **AND** system initializes default prompts if files don't exist

### Requirement: Version-based choice tracking

The system SHALL store user's choice (overwrite/keep) per application version and only prompt once per version.

#### Scenario: First run with modified prompts
- **WHEN** system detects modified prompts and no choice exists for current version
- **THEN** system prompts user: "Prompt files have been modified. Overwrite with new versions? (Your modifications will be lost)"
- **AND** user can choose "overwrite" or "keep"
- **AND** system stores choice for current version

#### Scenario: Subsequent runs with same version
- **WHEN** system detects modified prompts but choice already exists for current version
- **THEN** system uses stored choice without prompting user
- **AND** system respects user's previous choice (overwrite or keep)

#### Scenario: New version upgrade
- **WHEN** application version changes (e.g., v0.0.5 → v0.0.6)
- **THEN** system checks for modified prompts again
- **AND** if modifications detected, prompts user again (new version = new prompt)
- **AND** stores choice for new version separately

### Requirement: User guidance for manual upgrade

The system SHALL provide clear guidance when user chooses to keep modified prompts.

#### Scenario: User chooses to keep modified prompts
- **WHEN** user chooses "keep" when prompted about prompt upgrades
- **THEN** system displays message: "To upgrade prompt files later, delete files in ~/.aiq/prompts/ and restart aiq"
- **AND** system respects user's choice and doesn't overwrite files

#### Scenario: User deletes prompt files
- **WHEN** user manually deletes files in `~/.aiq/prompts/`
- **THEN** system detects missing files on next startup
- **AND** system creates default prompt files automatically
- **AND** system doesn't prompt user (files don't exist, so no conflict)

## MODIFIED Requirements

### Requirement: Default prompt initialization

The system SHALL respect user's version choice when initializing default prompts.

#### Scenario: User chose overwrite
- **WHEN** user chose "overwrite" for current version
- **THEN** system overwrites existing prompt files with built-in versions
- **AND** system creates files if they don't exist

#### Scenario: User chose keep
- **WHEN** user chose "keep" for current version
- **THEN** system skips file creation if files already exist
- **AND** system only creates files if they don't exist
