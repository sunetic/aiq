## ADDED Requirements

### Requirement: Runtime version detection
The system SHALL dynamically detect application version at runtime using multiple fallback mechanisms.

#### Scenario: Get version from build-time injection
- **WHEN** application starts
- **THEN** system first attempts to get version from build-time injected variable (via `-ldflags`)

#### Scenario: Fallback to git describe
- **WHEN** build-time version is not available and `.git` directory exists
- **THEN** system attempts to get version using `git describe --tags --always`

#### Scenario: Use default version
- **WHEN** build-time version and git describe are both unavailable
- **THEN** system uses default version value "dev"

### Requirement: Runtime commit ID detection
The system SHALL dynamically detect commit ID at runtime using multiple fallback mechanisms.

#### Scenario: Get commit ID from build-time injection
- **WHEN** application starts
- **THEN** system first attempts to get commit ID from build-time injected variable (via `-ldflags`)

#### Scenario: Fallback to git rev-parse
- **WHEN** build-time commit ID is not available and `.git` directory exists
- **THEN** system attempts to get commit ID using `git rev-parse HEAD`

#### Scenario: Use default commit ID
- **WHEN** build-time commit ID and git rev-parse are both unavailable
- **THEN** system uses default commit ID value "unknown"

### Requirement: Version command line flag
The system SHALL support `-v` and `--version` command line flags to display version information.

#### Scenario: Display version with -v flag
- **WHEN** user runs `aiq -v`
- **THEN** system displays version information and exits without entering interactive mode

#### Scenario: Display version with --version flag
- **WHEN** user runs `aiq --version`
- **THEN** system displays version information and exits without entering interactive mode

#### Scenario: Version output format
- **WHEN** user runs version command
- **THEN** system displays formatted string: "aiq <version> (commit: <commit-id>)"

#### Scenario: Version command exits immediately
- **WHEN** user runs version command
- **THEN** system exits with code 0 after displaying version, without initializing configuration or entering menu

### Requirement: Version information formatting
The system SHALL provide formatted version information string combining version and commit ID.

#### Scenario: Format version info
- **WHEN** system needs to display version information
- **THEN** system formats string as "aiq <version> (commit: <commit-id>)" where version and commit ID are current runtime values

#### Scenario: Handle missing commit ID
- **WHEN** commit ID is "unknown"
- **THEN** system still displays version information with "unknown" commit ID

### Requirement: Build-time version injection
The system SHALL support build-time injection of version and commit ID via linker flags.

#### Scenario: Inject version via ldflags
- **WHEN** binary is built with `-ldflags "-X github.com/aiq/aiq/internal/version.Version=<value>"`
- **THEN** system uses injected version value at runtime

#### Scenario: Inject commit ID via ldflags
- **WHEN** binary is built with `-ldflags "-X github.com/aiq/aiq/internal/version.CommitID=<value>"`
- **THEN** system uses injected commit ID value at runtime

#### Scenario: CI/CD version injection
- **WHEN** binary is built in CI/CD pipeline
- **THEN** build process injects version from git tag and commit ID from CI environment variables
