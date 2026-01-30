## ADDED Requirements

### Requirement: Command timeout user prompt
The system SHALL prompt the user when a command execution times out, allowing them to choose whether to continue waiting or cancel the command.

#### Scenario: Prompt user on idle timeout
- **WHEN** command execution has been idle (no output) for the configured timeout period (default: 60 seconds)
- **THEN** system displays a confirmation prompt asking if the user wants to continue waiting
- **AND** command execution continues in the background while waiting for user response

#### Scenario: User chooses to continue waiting
- **WHEN** user responds "yes" to the timeout prompt
- **THEN** system resets the idle timeout timer
- **AND** command execution continues normally
- **AND** system will prompt again if command becomes idle again for the timeout period

#### Scenario: User chooses to cancel
- **WHEN** user responds "no" to the timeout prompt
- **THEN** system terminates the command process
- **AND** system returns a timeout error indicating the command was cancelled due to timeout

#### Scenario: User interrupts prompt
- **WHEN** user presses Ctrl+C during the timeout prompt
- **THEN** system treats it as cancellation
- **AND** system terminates the command process
- **AND** system returns an error indicating the command was cancelled by user

#### Scenario: Multiple timeout prompts
- **WHEN** user chooses to continue waiting after a timeout
- **AND** command becomes idle again for the timeout period
- **THEN** system prompts the user again
- **AND** user can choose to continue or cancel at each prompt

#### Scenario: Prompt does not interfere with output
- **WHEN** command produces output while timeout prompt is displayed
- **THEN** prompt is displayed on a separate line
- **AND** command output continues to be displayed normally
- **AND** prompt and output do not interfere with each other
