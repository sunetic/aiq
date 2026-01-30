## MODIFIED Requirements

### Requirement: Configuration file management
The system SHALL store LLM configuration in YAML format at `~/.aiq/config/config.yaml` and manage additional configuration files in the config directory.

#### Scenario: Read configuration
- **WHEN** application starts and configuration file exists
- **THEN** system loads configuration from `~/.aiq/config/config.yaml`

#### Scenario: Configuration file format
- **WHEN** configuration is saved
- **THEN** file contains LLM URL and API key in YAML format

#### Scenario: Store version choice records
- **WHEN** user makes a choice regarding prompt file upgrades
- **THEN** system stores choice in `~/.aiq/config/prompt-version-choices.yaml` file

#### Scenario: Read version choice records
- **WHEN** system needs to check if user has made a choice for current application version
- **THEN** system reads version choices from `~/.aiq/config/prompt-version-choices.yaml` file

#### Scenario: Version choices file format
- **WHEN** version choices are saved
- **THEN** file contains YAML structure mapping application versions to user choices (overwrite/keep)

#### Scenario: Handle missing version choices file
- **WHEN** version choices file does not exist
- **THEN** system treats it as no previous choices made and prompts user if needed
