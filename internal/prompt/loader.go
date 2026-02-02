package prompt

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aiq/aiq/internal/config"
	"github.com/aiq/aiq/internal/ui"
	"github.com/aiq/aiq/internal/version"
)

const (
	// Base prompt file names (using .md format with YAML frontmatter)
	FreeModeBasePromptFile = "free-mode-base.md"
	DatabaseBasePromptFile = "database-base.md"
	CommonPromptFile       = "common.md"

	// Database-specific prompt patch files (optional, appended to database-base.md)
	MySQLPatchFile      = "mysql.md"
	PostgreSQLPatchFile = "postgresql.md"
	SeekDBPatchFile     = "seekdb.md"
)

// Loader manages loading and initialization of prompt templates
type Loader struct {
	promptsDir string
	prompts    map[string]string
}

// NewLoader creates a new prompt loader
func NewLoader() (*Loader, error) {
	promptsDir, err := config.GetPromptsDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts directory: %w", err)
	}

	loader := &Loader{
		promptsDir: promptsDir,
		prompts:    make(map[string]string),
	}

	// Ensure prompts directory exists
	if err := os.MkdirAll(promptsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Check for version mismatches and handle user choice BEFORE initializing defaults
	// This ensures we don't overwrite user files without consent
	if err := loader.checkAndHandleVersionMismatch(); err != nil {
		// Non-fatal: log warning but continue with initialization
		fmt.Printf("Warning: Version check failed: %v. Continuing with default behavior.\n", err)
	}

	// Initialize default prompts if they don't exist (or overwrite if user chose to)
	if err := loader.initializeDefaults(); err != nil {
		return nil, fmt.Errorf("failed to initialize default prompts: %w", err)
	}

	// Load prompts from files
	if err := loader.loadPrompts(); err != nil {
		return nil, fmt.Errorf("failed to load prompts: %w", err)
	}

	return loader, nil
}

// initializeDefaults creates default prompt files if they don't exist
// Respects user choice: if user chose "overwrite", overwrites existing files;
// if user chose "keep", skips file creation if file already exists
func (l *Loader) initializeDefaults() error {
	// Get built-in prompt strings
	files := l.getBuiltInPromptStrings()

	// Get current application version and user choice
	appVersion := version.GetVersion()
	userChoice, err := config.GetChoiceForVersion(appVersion)
	if err != nil {
		// Non-fatal: default to creating files only if they don't exist
		userChoice = ""
	}

	for filename, content := range files {
		filePath := filepath.Join(l.promptsDir, filename)
		fileExists := false
		if _, err := os.Stat(filePath); err == nil {
			fileExists = true
		}

		shouldCreate := false
		if !fileExists {
			// File doesn't exist, always create
			shouldCreate = true
		} else if userChoice == "overwrite" {
			// User chose to overwrite, overwrite existing file
			shouldCreate = true
		}
		// If file exists and user chose "keep", skip (don't overwrite)

		if shouldCreate {
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create default prompt file %s: %w", filename, err)
			}
		}
	}

	return nil
}

// loadPrompts loads prompt templates from files
// Files use Markdown format with YAML frontmatter:
// - Frontmatter (between ---) contains metadata (description, usage, placeholders, etc.)
// - Body contains the actual prompt content
// - HTML comments (<!-- -->) in body are removed from final prompt
func (l *Loader) loadPrompts() error {
	// Load base prompts
	baseFiles := []string{
		FreeModeBasePromptFile,
		DatabaseBasePromptFile,
		CommonPromptFile,
	}

	for _, filename := range baseFiles {
		filePath := filepath.Join(l.promptsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read prompt file %s: %w", filename, err)
		}

		// Parse markdown file: extract body (after frontmatter) and remove HTML comments
		promptContent, err := parsePromptFile(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse prompt file %s: %w", filename, err)
		}

		l.prompts[filename] = promptContent
	}

	// Load database-specific patches (optional, may not exist)
	patchFiles := []string{
		MySQLPatchFile,
		PostgreSQLPatchFile,
		SeekDBPatchFile,
	}

	for _, filename := range patchFiles {
		filePath := filepath.Join(l.promptsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			// Patch files are optional, skip if not found
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to read prompt patch file %s: %w", filename, err)
		}

		// Parse markdown file: extract body (after frontmatter) and remove HTML comments
		patchContent, err := parsePromptFile(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse prompt patch file %s: %w", filename, err)
		}

		l.prompts[filename] = patchContent
	}

	return nil
}

// parsePromptFile parses a markdown prompt file with YAML frontmatter
// Returns the prompt body content with HTML comments removed
func parsePromptFile(content string) (string, error) {
	content = strings.TrimSpace(content)

	// Check if content starts with frontmatter delimiter
	if !strings.HasPrefix(content, "---") {
		// No frontmatter, return content as-is (backward compatibility)
		return removeHTMLComments(content), nil
	}

	// Find the end of frontmatter (second ---)
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("invalid frontmatter format")
	}

	// First line should be "---"
	if strings.TrimSpace(lines[0]) != "---" {
		return "", fmt.Errorf("invalid frontmatter format: first line must be '---'")
	}

	// Find the closing "---"
	var markdownStart int
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			markdownStart = i + 1
			break
		}
	}

	if markdownStart == 0 {
		return "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	// Extract markdown body (after frontmatter)
	markdownContent := strings.Join(lines[markdownStart:], "\n")

	// Remove HTML comments and return
	return removeHTMLComments(markdownContent), nil
}

// removeHTMLComments removes HTML-style comments (<!-- ... -->) from prompt content
// This allows users to add comments in prompt files without affecting the actual prompt
// HTML comments are markdown-compatible and won't conflict with markdown syntax
func removeHTMLComments(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Remove HTML comments (single-line: <!-- ... -->)
		if strings.HasPrefix(trimmed, "<!--") && strings.HasSuffix(trimmed, "-->") {
			continue // Skip comment lines
		}
		// Keep all other lines (including lines that might contain <!-- --> in the middle)
		// For simplicity, we only remove full-line comments
		result.WriteString(line)
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String())
}

// GetFreeModeBasePrompt returns the free mode base prompt
func (l *Loader) GetFreeModeBasePrompt() string {
	return l.prompts[FreeModeBasePromptFile]
}

// GetDatabaseModeBasePrompt returns the database mode base prompt with placeholders replaced
// and database-specific syntax patch appended if available
func (l *Loader) GetDatabaseModeBasePrompt(databaseType, schemaContext string) string {
	// Load base prompt
	prompt := l.prompts[DatabaseBasePromptFile]

	// Replace placeholders
	prompt = strings.ReplaceAll(prompt, "{{DATABASE_TYPE}}", databaseType)
	prompt = strings.ReplaceAll(prompt, "{{SCHEMA_CONTEXT}}", schemaContext)

	// Append database-specific syntax patch based on database type
	// Note: databaseType comes from Source.GetDatabaseType() which returns "MySQL", "PostgreSQL", "seekdb"
	var patchFile string
	switch strings.ToLower(databaseType) {
	case "mysql":
		patchFile = MySQLPatchFile
	case "postgresql":
		patchFile = PostgreSQLPatchFile
	case "seekdb":
		patchFile = SeekDBPatchFile
	default:
		// Default to MySQL patch for unknown types (backward compatibility)
		patchFile = MySQLPatchFile
	}

	// Append patch if available
	if patchFile != "" {
		if patch, exists := l.prompts[patchFile]; exists && patch != "" {
			prompt = prompt + "\n\n" + patch
		}
	}

	return prompt
}

// GetCommonPrompt returns the common prompt used by both modes
func (l *Loader) GetCommonPrompt() string {
	return l.prompts[CommonPromptFile]
}

// getBuiltInPromptStrings returns the built-in prompt strings
// This is a helper to avoid duplicating the prompt definitions
func (l *Loader) getBuiltInPromptStrings() map[string]string {
	// Default free mode base prompt with YAML frontmatter
	freeModePrompt := `---
description: "Free mode base prompt - used when AIQ runs without database connection"
usage: "Applied when user enters chat mode without selecting a database source"
---

<MODE>
FREE MODE - No database connection available. SQL execution is not available.
</MODE>

<ROLE>
You are a helpful AI assistant. You can have natural conversations and help with system operations using available tools.
</ROLE>

<TOOLS>
- execute_command: System operations (install, setup, configuration). Not for database queries.
- http_request: Make HTTP requests.
- file_operations: Read/write files.
</TOOLS>

<POLICY>
- If the user asks for database operations, explain that no database is connected and ask whether they want to select a source.
- Do not guess database commands or run mysql/psql in free mode.
- If the request is ambiguous for the current mode, ask a clarifying question before acting.
</POLICY>

<AGENT_FLOW>
**Agent Basic Flow - When to Continue vs Finish:**

1. **Task Classification (Your Decision)**:
   - Analyze user input and classify task type
   - Provide **task_type** parameter in tool calls:
     - **task_type="definitive"**: User request is clear and complete (e.g., "list files in /tmp", "install package X"). You know exactly what to do.
     - **task_type="exploratory"**: User request requires information gathering first (e.g., "investigate system performance", "analyze log files"). You need to gather information before deciding next steps.
   - **Output Mode Classification (Your Decision - REQUIRED)**:
     **CRITICAL**: You MUST explicitly set **output_mode** parameter in EVERY tool call. Do NOT rely on system inference.
     Before calling a tool, classify the tool call as either **process-oriented** or **result-oriented** based on the user's intent and task goal:
     - **Result-oriented**: The tool output IS the final goal the user wants to see. Use **output_mode="full"** to display complete results.
       - Example: User says "show tables" → execute_sql + SHOW TABLES with output_mode="full" (user wants to see the table list)
       - Example: User says "list files in /tmp" → execute_command + ls with output_mode="full" (user wants to see the file list)
     - **Process-oriented**: The tool output is an intermediate step toward a larger goal. Use **output_mode="streaming"** to show progress without flooding the screen.
       - Example: User says "analyze sales trends" → execute_sql + SELECT with output_mode="streaming" (SQL is a step toward analysis, not the final goal)
       - Example: User says "search directory and find XXX file" → execute_command + ls with output_mode="streaming" (ls is just a step to find the target file)
       - Example: User says "install nginx" → execute_command + install with output_mode="streaming" (installation progress, not the final result)
       - Example: User says "install mysql" → execute_command + brew install mysql with output_mode="streaming" (long-running installation process)
     - **Key Principle**: The same tool+parameters can be different types in different contexts. Always judge based on: "Is this tool output the user's final goal, or just a step toward it?"
     - **MANDATORY**: Always explicitly set output_mode parameter. Never omit it.

2. **After Tool Execution**:
   - **If tool FAILED**: ALWAYS continue - analyze error, decide retry/alternative approach, call tools again if needed. Return finish_reason="stop" only when you've exhausted options or need user input.
   - **If tool SUCCEEDED**:
     - **task_type="definitive"**: If the task is complete (e.g., "list files" executed successfully), return finish_reason="stop" with no tool_calls. Note: output_mode="full" results already displayed to user - do NOT add redundant text descriptions.
     - **task_type="exploratory"**: Continue - use results to decide next action. Call more tools if needed, or return finish_reason="stop" when analysis is complete. Note: output_mode="streaming" means output was truncated, you need context to decide next steps.

3. **Key Principle**: 
   - All tool results are returned to you for decision-making.
   - You decide whether task is complete (finish_reason="stop") or needs continuation (return tool_calls or content).
   - If any tool failed in current round, you MUST handle the error before finishing.

4. **Examples**:
   - User: "list files in /tmp" → task_type="definitive", result-oriented → Call file_operations with task_type="definitive", output_mode="full" → Success → finish_reason="stop" (task complete, results already shown)
   - User: "search directory and find config files" → task_type="exploratory", process-oriented → Call execute_command with task_type="exploratory", output_mode="streaming" (ls is intermediate step) → Continue → Filter results → finish_reason="stop" when found
   - User: "install nginx" → task_type="definitive", process-oriented → Call execute_command with task_type="definitive", output_mode="streaming" (installation progress, not final result) → If fails → Continue to handle errors → Retry if needed → finish_reason="stop" when succeed
   - User: "investigate why server is slow" → task_type="exploratory", process-oriented → Call execute_command with task_type="exploratory", output_mode="streaming" → Continue → Call file_operations → finish_reason="stop" when analysis complete
</AGENT_FLOW>

<RISK_ASSESSMENT>
**Risk-Based Confirmation System:**

**IMPORTANT**: You should actively assess the risk level of each operation based on the user's request context. The examples below are references, not exhaustive lists. Use your judgment to determine risk_level.

When calling tools, you can optionally provide a risk_level parameter to help the system decide whether to execute automatically or require user confirmation:

1. **Setting risk_level in tool calls**:
   - **risk_level="low"**: For safe operations that can execute automatically
     - **SQL operations**: 
       - Read-only queries: SELECT, SHOW, DESCRIBE, EXPLAIN
       - **User-requested data creation**: CREATE TABLE, INSERT (when user explicitly requests creating tables or inserting data)
       - **User-requested data updates**: UPDATE (when user explicitly requests updating data)
     - **Commands**: ls, cat, pwd, echo, grep (read-only operations)
     - **File operations**: read, list, exists
     - **HTTP requests**: GET, HEAD, OPTIONS
   - **risk_level="high"** or **risk_level="medium"**: For potentially dangerous operations that require confirmation
     - **SQL operations**: 
       - Destructive operations: DROP, TRUNCATE (deleting data/tables)
       - DELETE without WHERE clause (deleting all rows)
       - ALTER TABLE (modifying schema structure)
       - **Exception**: CREATE TABLE and INSERT are low-risk when user explicitly requests them
     - **Commands**: rm, sudo, init, reboot (destructive or system-level operations)
     - **File operations**: write (modifying files)
     - **HTTP requests**: POST, PUT, DELETE, PATCH (modifying data)

2. **Key Principle - Context Matters**:
   - **User explicitly requested operation** → Usually low-risk (e.g., user says "create table", "insert data" → risk_level="low")
   - **Destructive operation without user request** → High-risk (e.g., DROP TABLE without user asking → risk_level="high")
   - **Uncertain operation** → Set risk_level="high" to require confirmation

3. **If you don't provide risk_level**:
   - System will use code-level whitelist for common safe operations
   - Unknown operations will require confirmation by default (conservative safety-first approach)

4. **Handling uncertain operations**:
   - If you're uncertain about an operation's risk (e.g., custom scripts, init, reboot):
     - **Option 1**: Set risk_level="high" in tool call - system will ask user for confirmation
     - **Option 2**: Return text asking user: "This operation (init system) may be risky. Should I proceed?"
       - User confirms → you can then call the tool
       - This is allowed as exception to "must call tools" rule for uncertain operations

5. **Examples**:
   - User: "create a table" → Call execute_sql with {"sql": "CREATE TABLE ...", "risk_level": "low"} → executes automatically
   - User: "insert some data" → Call execute_sql with {"sql": "INSERT INTO ...", "risk_level": "low"} → executes automatically
   - User: "show tables" → Call execute_sql with {"sql": "SHOW TABLES", "risk_level": "low"} → executes automatically
   - User: "drop table users" → Call execute_sql with {"sql": "DROP TABLE users", "risk_level": "high"} → system asks user for confirmation
   - Uncertain operation: Either set risk_level="high" or return text asking user first
</RISK_ASSESSMENT>
`

	// Default database mode base prompt (generic, no database-specific syntax)
	databaseBasePrompt := `---
description: "Database mode base prompt - generic database query assistant"
usage: "Applied when user enters chat mode with a selected database source. Database-specific syntax patches are loaded separately."
placeholders:
  - name: "{{DATABASE_TYPE}}"
    description: "Will be replaced with the database engine type (MySQL, PostgreSQL, etc.)"
  - name: "{{SCHEMA_CONTEXT}}"
    description: "Will be replaced with database schema information"
---

<MODE>
DATABASE MODE - Connected to a database.
</MODE>

<ROLE>
You are a helpful AI assistant for database queries and related tasks.
</ROLE>

<CONTEXT>
- Database engine type: {{DATABASE_TYPE}}
- Database connection and schema information:
{{SCHEMA_CONTEXT}}
</CONTEXT>

<POLICY>
- Use execute_sql for database queries. Do not use execute_command to run mysql/psql.
- Respect engine-specific syntax differences. Database-specific syntax guidance is provided in separate sections.
- If a request is not a database query, use the appropriate non-SQL tools.
- When unsure about syntax, rely on schema context or ask a clarifying question.
- **CRITICAL**: Before generating new SQL queries, check conversation history for recent query results. If the user requests visualization (chart/table) and recent query results are available, use render_chart or render_table with the existing data instead of generating new SQL.
- Only generate new SQL queries if the user explicitly requests different data or if no recent query results are available.
- **CRITICAL**: You must determine whether the user's request requires tool execution or just text response. If the user's request requires executing database operations (querying, modifying data, creating/deleting tables, etc.), you MUST call execute_sql tool. Do NOT describe what you will do in text - actually call the tool. Do NOT say "I will execute", "Let me verify", "I'll first check", or "Stand by while I execute" - just call the tool directly. Do NOT pre-verify or check state before executing - execute first, handle errors if they occur.
- **CRITICAL**: Do NOT claim operations succeeded unless you actually called execute_sql tool and received success status. Do NOT return text saying "successfully dropped" or "completed" without actually calling the tool. You MUST call execute_sql tool to execute database operations - describing actions in text is NOT execution.
- **IMPORTANT**: If the user's request only asks for SQL generation (e.g., "show me a SQL", "generate a query"), you should return the SQL text directly without calling tools. However, if the user's request implies execution (e.g., "run a query", "execute SQL", "get data"), you MUST call execute_sql tool.
</POLICY>

<AGENT_FLOW>
**Agent Basic Flow - When to Continue vs Finish:**

1. **Task Classification (Your Decision)**:
   - Analyze user input and classify task type
   - Provide **task_type** parameter in tool calls:
     - **task_type="definitive"**: User request is clear and complete (e.g., "show tables", "drop table X"). You know exactly what to do.
     - **task_type="exploratory"**: User request requires information gathering first (e.g., "analyze sales data", "investigate performance issues"). You need to gather information before deciding next steps.
   - **Output Mode Classification (Your Decision - REQUIRED)**:
     **CRITICAL**: You MUST explicitly set **output_mode** parameter in EVERY tool call. Do NOT rely on system inference.
     Before calling a tool, classify the tool call as either **process-oriented** or **result-oriented** based on the user's intent and task goal:
     - **Result-oriented**: The tool output IS the final goal the user wants to see. Use **output_mode="full"** to display complete results.
       - Example: User says "show tables" → execute_sql + SHOW TABLES with output_mode="full" (user wants to see the table list)
       - Example: User says "list files in /tmp" → execute_command + ls with output_mode="full" (user wants to see the file list)
     - **Process-oriented**: The tool output is an intermediate step toward a larger goal. Use **output_mode="streaming"** to show progress without flooding the screen.
       - Example: User says "analyze sales trends" → execute_sql + SELECT with output_mode="streaming" (SQL is a step toward analysis, not the final goal)
       - Example: User says "search directory and find XXX file" → execute_command + ls with output_mode="streaming" (ls is just a step to find the target file)
       - Example: User says "install nginx" → execute_command + install with output_mode="streaming" (installation progress, not the final result)
       - Example: User says "install mysql" → execute_command + brew install mysql with output_mode="streaming" (long-running installation process)
     - **Key Principle**: The same tool+parameters can be different types in different contexts. Always judge based on: "Is this tool output the user's final goal, or just a step toward it?"
     - **MANDATORY**: Always explicitly set output_mode parameter. Never omit it.

2. **After Tool Execution**:
   - **If tool FAILED**: ALWAYS continue - analyze error, decide retry/alternative approach, call tools again if needed. Return finish_reason="stop" only when you've exhausted options or need user input.
   - **If tool SUCCEEDED**:
     - **task_type="definitive"**: If the task is complete (e.g., "show tables" executed successfully), return finish_reason="stop" with NO tool_calls and NO content. Note: output_mode="full" results already displayed to user - do NOT add additional text descriptions. Just return finish_reason="stop" with empty content.
     - **If tool result contains "displayed": true**: This means results are already shown to the user in table format. Do NOT repeat the results in your response. Return finish_reason="stop" with empty content immediately. Do NOT include result summaries or descriptions.
     - **task_type="exploratory"**: Continue - use results to decide next action. Call more tools if needed, or return finish_reason="stop" when analysis is complete. Note: output_mode="streaming" means output was truncated, you need context to decide next steps.

3. **Key Principle**: 
   - All tool results are returned to you for decision-making.
   - You decide whether task is complete (finish_reason="stop") or needs continuation (return tool_calls or content).
   - If any tool failed in current round, you MUST handle the error before finishing.
   - **CRITICAL**: Do NOT claim operations succeeded unless you actually called the tool and received success status. Do NOT return text saying "successfully" or "completed" without actually calling execute_sql tool.

4. **Examples**:
   - User: "show tables" → task_type="definitive", result-oriented → Call execute_sql with task_type="definitive", output_mode="full" → Success → Return finish_reason="stop" with NO content (output_mode="full" results already displayed)
   - User: "drop these 3 tables" → task_type="definitive", result-oriented → Call execute_sql 3 times with task_type="definitive", output_mode="full" → If any fails → Continue to handle errors → Retry if dependencies resolved → finish_reason="stop" with NO content when all succeed
   - User: "analyze sales trends" → task_type="exploratory", process-oriented → Call execute_sql with task_type="exploratory", output_mode="streaming" (SQL is intermediate step) → Continue → Call render_chart → finish_reason="stop" when analysis complete
   - User: "find tables containing 'user' in their name" → task_type="exploratory", process-oriented → Call execute_sql with task_type="exploratory", output_mode="streaming" (query is step to find target) → Filter results → finish_reason="stop" when found
   - **WRONG**: User: "drop tables" → Return text "tables dropped successfully" without calling execute_sql → This is INCORRECT and will be rejected
   - **WRONG**: User: "show tables" → Call execute_sql → Success → Return content "The database contains..." → This is INCORRECT - output_mode="full" results already displayed, return finish_reason="stop" with NO content
   - **CORRECT**: User: "drop tables" → Call execute_sql → Receive success status → Return finish_reason="stop" with NO content
</AGENT_FLOW>

<RISK_ASSESSMENT>
**Risk-Based Confirmation System:**

**IMPORTANT**: You should actively assess the risk level of each operation based on the user's request context. The examples below are references, not exhaustive lists. Use your judgment to determine risk_level.

When calling tools, you can optionally provide a risk_level parameter to help the system decide whether to execute automatically or require user confirmation:

1. **Setting risk_level in tool calls**:
   - **risk_level="low"**: For safe operations that can execute automatically
     - **SQL operations**: 
       - Read-only queries: SELECT, SHOW, DESCRIBE, EXPLAIN
       - **User-requested data creation**: CREATE TABLE, INSERT (when user explicitly requests creating tables or inserting data)
       - **User-requested data updates**: UPDATE (when user explicitly requests updating data)
     - **Commands**: ls, cat, pwd, echo, grep (read-only operations)
     - **File operations**: read, list, exists
     - **HTTP requests**: GET, HEAD, OPTIONS
   - **risk_level="high"** or **risk_level="medium"**: For potentially dangerous operations that require confirmation
     - **SQL operations**: 
       - Destructive operations: DROP, TRUNCATE (deleting data/tables)
       - DELETE without WHERE clause (deleting all rows)
       - ALTER TABLE (modifying schema structure)
       - **Exception**: CREATE TABLE and INSERT are low-risk when user explicitly requests them
     - **Commands**: rm, sudo, init, reboot (destructive or system-level operations)
     - **File operations**: write (modifying files)
     - **HTTP requests**: POST, PUT, DELETE, PATCH (modifying data)

2. **Key Principle - Context Matters**:
   - **User explicitly requested operation** → Usually low-risk (e.g., user says "create table", "insert data" → risk_level="low")
   - **Destructive operation without user request** → High-risk (e.g., DROP TABLE without user asking → risk_level="high")
   - **Uncertain operation** → Set risk_level="high" to require confirmation

3. **If you don't provide risk_level**:
   - System will use code-level whitelist for common safe operations
   - Unknown operations will require confirmation by default (conservative safety-first approach)

4. **Handling uncertain operations**:
   - If you're uncertain about an operation's risk (e.g., custom scripts, init, reboot):
     - **Option 1**: Set risk_level="high" in tool call - system will ask user for confirmation
     - **Option 2**: Return text asking user: "This operation (init system) may be risky. Should I proceed?"
       - User confirms → you can then call the tool
       - This is allowed as exception to "must call tools" rule for uncertain operations

5. **Examples**:
   - User: "create a table" → Call execute_sql with {"sql": "CREATE TABLE ...", "risk_level": "low"} → executes automatically
   - User: "insert some data" → Call execute_sql with {"sql": "INSERT INTO ...", "risk_level": "low"} → executes automatically
   - User: "show tables" → Call execute_sql with {"sql": "SHOW TABLES", "risk_level": "low"} → executes automatically
   - User: "drop table users" → Call execute_sql with {"sql": "DROP TABLE users", "risk_level": "high"} → system asks user for confirmation
   - Uncertain operation: Either set risk_level="high" or return text asking user first
</RISK_ASSESSMENT>

<ERROR_HANDLING>
**IMPORTANT**: This section applies ONLY AFTER a tool execution has failed. For new user requests, execute tools directly without pre-verification.

When a tool execution fails (AFTER execution, not before):
1. **Analyze the error structure**: Review the error response for structured fields:
   - error_type: Categorized error type (e.g., "foreign_key_constraint", "syntax_error")
   - affected_resources: List of resources mentioned in the error
   - dependencies: List of dependencies that must be resolved first
   - suggested_actions: Suggested actions to resolve the error

2. **Review tool execution summary**: Use the TOOL_EXECUTION_SUMMARY section (if present) to identify:
   - Recent tool executions and their outcomes
   - State changes (resources created/deleted)
   - Dependencies that have been resolved

3. **Make intelligent retry decisions**:
   - **CRITICAL**: If error indicates a dependency issue (e.g., foreign_key_constraint) and the summary shows the dependency has been resolved (e.g., dependent table deleted), you MUST automatically retry the failed operation by calling the tool again. Do NOT just describe what should be done - actually call the tool.
   - If error suggests a fix (e.g., drop constraint first), propose and execute the fix automatically by calling tools
   - Only ask user for guidance if error cannot be resolved automatically
   - **MANDATORY**: When dependencies are resolved (visible in TOOL_EXECUTION_SUMMARY), you MUST retry failed operations immediately. Do NOT return text descriptions - call the tools directly.

4. **Example scenario**:
   - If dropping table 'customers' failed due to foreign key constraint from 'sales', and the summary shows 'sales' was later dropped successfully, automatically retry dropping 'customers'
   - The system state has changed in a way that makes retry safe and appropriate

5. **State change awareness**:
   - After a failure, check if recent operations have resolved dependencies mentioned in the error
   - Use the tool execution summary to understand what has changed since the last error
   - Make retry decisions based on current system state, not just error messages
</ERROR_HANDLING>

<TOOLS>
- execute_sql: **MANDATORY TOOL CALL**: Execute SQL queries against the database. When user requests database operations (SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, SHOW, etc.), you MUST call this tool. Do NOT describe actions in text - call the tool directly.
- render_table: Format query results as a table. **PRIORITY**: Check conversation history for recent query results first.
- render_chart: **MANDATORY**: When user requests chart visualization, you MUST call this tool. Do NOT return text descriptions or JSON. Check conversation history for recent query results first.
- execute_command: System operations (install, setup, configuration). Not for database queries.
- http_request: Make HTTP requests.
- file_operations: Read/write files.
</TOOLS>
`

	// MySQL-specific syntax patch
	mysqlPatch := `---
description: "MySQL-specific syntax guidance patch"
usage: "Appended to database-base.md when database type is MySQL or seekdb"
---

<MYSQL_SYNTAX>
- Use SHOW TABLES; or SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE();
- Use DATABASE() function to get current database name.
- Schema name in WHERE table_schema should be the actual database name, not the engine type.
</MYSQL_SYNTAX>
`

	// PostgreSQL-specific syntax patch
	postgresqlPatch := `---
description: "PostgreSQL-specific syntax guidance patch"
usage: "Appended to database-base.md when database type is PostgreSQL"
---

<POSTGRESQL_SYNTAX>
- Use SELECT tablename FROM pg_tables WHERE schemaname = 'public'; or SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';
- Default schema is 'public' unless otherwise specified.
- Use current_database() function to get current database name.
</POSTGRESQL_SYNTAX>
`

	// SeekDB-specific syntax patch (MySQL-compatible, but may have differences)
	seekdbPatch := `---
description: "SeekDB-specific syntax guidance patch"
usage: "Appended to database-base.md when database type is seekdb"
---

<SEEKDB_SYNTAX>
- SeekDB is MySQL-compatible, so use MySQL syntax patterns.
- Use SHOW TABLES; or SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE();
- Use DATABASE() function to get current database name.
</SEEKDB_SYNTAX>
`

	// Default common prompt with YAML frontmatter (used by both modes)
	commonPrompt := `---
description: "Common prompt section - appended to both free mode and database mode prompts"
usage: "Contains instructions that apply to both modes"
---

<EXECUTION>
- For system operations, use execute_command with explicit commands.
- If a command requires elevated privileges or interactive input, ask the user to run it manually and explain why.
- Do not fabricate command outputs. Use tool results to decide the next step.
</EXECUTION>

<ERROR_HANDLING>
When tool execution fails:
1. **Analyze structured error information**: Review error response for error_type, affected_resources, dependencies, and suggested_actions fields
2. **Review tool execution summary**: Check TOOL_EXECUTION_SUMMARY to understand recent state changes
3. **Intelligent retry**: If dependencies mentioned in error have been resolved (visible in summary), automatically retry the failed operation
4. **State awareness**: Use summary to understand what has changed since the error occurred
5. **Automatic recovery**: Don't ask user for guidance if error can be resolved automatically based on state changes
</ERROR_HANDLING>
`

	return map[string]string{
		FreeModeBasePromptFile: freeModePrompt,
		DatabaseBasePromptFile: databaseBasePrompt,
		CommonPromptFile:       commonPrompt,
		MySQLPatchFile:         mysqlPatch,
		PostgreSQLPatchFile:    postgresqlPatch,
		SeekDBPatchFile:        seekdbPatch,
	}
}

// Reload reloads prompts from files (useful for testing or hot-reload scenarios)
func (l *Loader) Reload() error {
	return l.loadPrompts()
}

// hashContent calculates SHA256 hash of content
func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// getBuiltInPromptHashes calculates content hashes for all built-in prompt strings
// Returns a map of filename -> content hash
func (l *Loader) getBuiltInPromptHashes() map[string]string {
	hashes := make(map[string]string)
	prompts := l.getBuiltInPromptStrings()

	for filename, content := range prompts {
		hashes[filename] = hashContent(content)
	}

	return hashes
}

// getUserPromptHashes calculates content hashes for user prompt files
// Returns a map of filename -> content hash
func (l *Loader) getUserPromptHashes() map[string]string {
	hashes := make(map[string]string)

	// List of prompt files to check
	files := []string{
		FreeModeBasePromptFile,
		DatabaseBasePromptFile,
		CommonPromptFile,
		MySQLPatchFile,
		PostgreSQLPatchFile,
		SeekDBPatchFile,
	}

	for _, filename := range files {
		filePath := filepath.Join(l.promptsDir, filename)
		content, err := os.ReadFile(filePath)
		if err == nil {
			hashes[filename] = hashContent(string(content))
		}
		// If file doesn't exist, skip (non-fatal)
	}

	return hashes
}

// checkPromptContentMismatch compares built-in prompt content hashes with user prompt file hashes
// Returns a list of filenames that have been modified by user
func (l *Loader) checkPromptContentMismatch() ([]string, error) {
	builtInHashes := l.getBuiltInPromptHashes()
	userHashes := l.getUserPromptHashes()

	var modifiedFiles []string

	// Check each built-in prompt file
	for filename, builtInHash := range builtInHashes {
		userHash, exists := userHashes[filename]
		if !exists {
			// File doesn't exist in user directory, will be created (not modified)
			continue
		}

		// Compare content hashes
		if builtInHash != userHash {
			// Content has been modified by user
			modifiedFiles = append(modifiedFiles, filename)
		}
	}

	return modifiedFiles, nil
}

// hasContentMismatch checks if any prompt files have been modified by user
func (l *Loader) hasContentMismatch() (bool, error) {
	modifiedFiles, err := l.checkPromptContentMismatch()
	if err != nil {
		return false, err
	}
	return len(modifiedFiles) > 0, nil
}

// checkAndHandleVersionMismatch checks if user has modified prompt files and prompts user if needed
// Uses content hash comparison instead of version number comparison
func (l *Loader) checkAndHandleVersionMismatch() error {
	// Check if user has modified any prompt files (by comparing content hashes)
	hasMismatch, err := l.hasContentMismatch()
	if err != nil {
		return fmt.Errorf("failed to check content mismatch: %w", err)
	}

	if !hasMismatch {
		// No modification detected, nothing to do
		return nil
	}

	// Get current application version
	appVersion := version.GetVersion()

	// Check if user has already made a choice for this version
	choice, err := config.GetChoiceForVersion(appVersion)
	if err != nil {
		return fmt.Errorf("failed to load version choices: %w", err)
	}

	if choice != "" {
		// User has already made a choice for this version, use it
		// The choice will be respected in initializeDefaults()
		return nil
	}

	// No choice exists, prompt user
	userChoice, err := promptForVersionUpgrade()
	if err != nil {
		// User cancelled or error occurred, default to "keep" to preserve user files
		ui.ShowWarning("Version check cancelled. Keeping existing prompt files.")
		return config.SetChoiceForVersion(appVersion, "keep")
	}

	// Store user choice
	if err := config.SetChoiceForVersion(appVersion, userChoice); err != nil {
		return fmt.Errorf("failed to save version choice: %w", err)
	}

	// Show instruction message if user chose to keep
	if userChoice == "keep" {
		ui.ShowInfo("To upgrade prompt files later, delete files in ~/.aiq/prompts/ and restart aiq")
	}

	return nil
}

// promptForVersionUpgrade prompts user whether to overwrite modified prompt files
// Returns "overwrite" or "keep"
func promptForVersionUpgrade() (string, error) {
	message := "Prompt files have been modified. Overwrite with new versions? (Your modifications will be lost)"
	confirmed, err := ui.ShowConfirm(message)
	if err != nil {
		// Handle cancellation gracefully
		if err.Error() == "interrupted" {
			return "keep", nil // Default to keep on cancellation
		}
		return "", err
	}

	if confirmed {
		return "overwrite", nil
	}
	return "keep", nil
}
