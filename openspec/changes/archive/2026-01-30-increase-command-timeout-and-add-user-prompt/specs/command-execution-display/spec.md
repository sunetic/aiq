## MODIFIED Requirements

### Requirement: Command execution display optimization
The system SHALL display command execution output in a user-friendly manner that avoids screen flooding, provides real-time feedback, and intelligently truncates output for LLM processing.

#### Scenario: Display tool call with normal text
- **WHEN** LLM calls `execute_command` tool
- **THEN** system displays tool call information (command name and arguments) using normal text color with loading icon (⏳)

#### Scenario: Show real-time output in rolling window
- **WHEN** command is executing and producing output
- **THEN** system displays output in a rolling window of 3 lines using gray/dimmed text color, updating in real-time without scrolling the entire screen
- **AND** uses ANSI escape codes to clear and redraw the rolling window area

#### Scenario: Avoid screen flooding
- **WHEN** command produces large amounts of output
- **THEN** system does not flood the screen with all output lines, but maintains the rolling window display showing only the latest 3 lines

#### Scenario: Use idle timeout instead of fixed timeout
- **WHEN** command is executing and producing output
- **THEN** system resets the 60-second idle timer on each output line
- **AND** only times out if no output is received for 60 seconds continuously

#### Scenario: Truncate successful command output for LLM
- **WHEN** command executes successfully (exit code 0) and produces output
- **THEN** system truncates output to last 20 lines before sending to LLM, reducing token consumption

#### Scenario: Include more output for failed commands
- **WHEN** command execution fails (exit code non-zero) and produces output
- **THEN** system truncates output to last 100 lines before sending to LLM, providing sufficient context for error diagnosis

#### Scenario: Display execution status with icon
- **WHEN** command execution completes successfully
- **THEN** system displays success status with ✓ icon and execution duration
- **WHEN** command execution fails
- **THEN** system displays failure status with ✗ icon, exit code, and execution duration

#### Scenario: Show output summary after completion
- **WHEN** command execution completes and there were many output lines
- **THEN** system displays "... (N more lines above)" hint below the rolling window
