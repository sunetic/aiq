## ADDED Requirements

### Requirement: Prompt file content detection
The system SHALL detect if user has modified prompt files by comparing content hashes between built-in prompt files and user-modified prompt files in `~/.aiq/prompts/` directory.

#### Scenario: Calculate hash for built-in prompts
- **WHEN** system initializes prompt loader
- **THEN** system calculates content hash (SHA256) for each built-in prompt string in code

#### Scenario: Calculate hash for user prompt files
- **WHEN** system checks prompt files in `~/.aiq/prompts/` directory
- **THEN** system reads file content and calculates content hash (SHA256) for each existing prompt file

#### Scenario: Compare content hashes on startup
- **WHEN** application starts and prompt loader initializes
- **THEN** system compares built-in prompt content hashes with user directory prompt file content hashes for each prompt file

#### Scenario: Detect content modification
- **WHEN** built-in prompt content hash differs from user directory prompt file content hash for any file
- **THEN** system marks that prompt file as having been modified by user

### Requirement: Prompt content upgrade prompt
The system SHALL prompt user when prompt file content has been modified, asking whether to overwrite user-modified files with new versions.

#### Scenario: Prompt user on content modification detected
- **WHEN** system detects content modification (hash mismatch) and user has not made a choice for current application version
- **THEN** system displays prompt asking user whether to overwrite modified prompt files with new versions

#### Scenario: User chooses to overwrite
- **WHEN** user chooses to overwrite prompt files
- **THEN** system overwrites existing prompt files in `~/.aiq/prompts/` with built-in versions

#### Scenario: User chooses to keep existing files
- **WHEN** user chooses to keep existing prompt files
- **THEN** system preserves user-modified prompt files and displays instruction message

#### Scenario: Display instruction message
- **WHEN** user chooses to keep existing prompt files
- **THEN** system displays message informing user they can manually delete files in `~/.aiq/prompts/` to trigger rebuild on next startup

### Requirement: Content hash calculation
The system SHALL calculate content hash (SHA256) for prompt files to detect modifications.

#### Scenario: Calculate hash for built-in prompts
- **WHEN** system needs to detect if user has modified prompts
- **THEN** system calculates SHA256 hash of built-in prompt string content

#### Scenario: Calculate hash for user prompt files
- **WHEN** system reads prompt file from `~/.aiq/prompts/` directory
- **THEN** system calculates SHA256 hash of file content

#### Scenario: Handle missing file
- **WHEN** prompt file does not exist in user directory
- **THEN** system treats file as not modified (will be created, no prompt needed)

#### Scenario: Handle file read errors
- **WHEN** prompt file cannot be read (permission error, corruption, etc.)
- **THEN** system treats file as modified (triggers upgrade prompt to restore valid file)

### Requirement: Version check per application version
The system SHALL only prompt user once per application version, storing user's choice to avoid repeated prompts.

#### Scenario: Check stored choice
- **WHEN** system detects content modification (hash mismatch)
- **THEN** system checks if user has already made a choice for current application version

#### Scenario: Skip prompt if choice exists
- **WHEN** user has already made a choice for current application version
- **THEN** system uses stored choice without prompting user again

#### Scenario: Store user choice
- **WHEN** user makes a choice (overwrite or keep)
- **THEN** system stores choice associated with current application version in version choices file
