## ADDED Requirements

### Requirement: Skills directory structure
The system SHALL support Skills stored in `~/.aiqconfig/skills/<skill-name>/SKILL.md` format.

#### Scenario: Skills directory organization
- **WHEN** user adds a Skill
- **THEN** Skill is stored in its own subdirectory: `~/.aiqconfig/skills/<skill-name>/SKILL.md`

#### Scenario: Multiple Skills support
- **WHEN** user has multiple Skills installed
- **THEN** each Skill is stored in a separate subdirectory under `~/.aiqconfig/skills/`

### Requirement: Skills file format parsing
The system SHALL parse SKILL.md files with YAML frontmatter and Markdown content.

#### Scenario: Parse Skill metadata
- **WHEN** system loads a Skill file
- **THEN** system extracts YAML frontmatter containing `name` and `description` fields

#### Scenario: Parse Skill content
- **WHEN** system loads a Skill file
- **THEN** system extracts Markdown content after YAML frontmatter

#### Scenario: Handle invalid Skill format
- **WHEN** Skill file has invalid YAML frontmatter or missing required fields
- **THEN** system logs error and skips the Skill without crashing

### Requirement: Progressive Skills loading
The system SHALL load Skills progressively based on query relevance, not all at once.

#### Scenario: Load Skills metadata on startup
- **WHEN** application starts
- **THEN** system loads only Skills metadata (name, description) from all Skill files

#### Scenario: Match Skills to user query
- **WHEN** user submits a query
- **THEN** system matches query keywords against Skills metadata to find relevant Skills

#### Scenario: Load matched Skills content
- **WHEN** Skills are matched to user query
- **THEN** system loads full content of matched Skills (default: top 3 most relevant)

#### Scenario: Cache loaded Skills
- **WHEN** Skill content is loaded
- **THEN** system caches loaded content to avoid re-reading from disk

### Requirement: Skills matching algorithm
The system SHALL match user queries to Skills using LLM-based semantic relevance judgment, with keyword-based matching as fallback.

#### Scenario: LLM semantic matching
- **WHEN** user submits a query
- **THEN** system sends query and all Skills metadata to LLM for semantic relevance judgment

#### Scenario: LLM returns relevant Skills
- **WHEN** LLM processes matching request
- **THEN** LLM returns list of relevant Skill names (top N, default 3) based on semantic understanding

#### Scenario: Fallback to keyword matching
- **WHEN** LLM semantic matching fails or is unavailable
- **THEN** system falls back to keyword-based matching algorithm (exact name match > partial name match > description keyword match)

#### Scenario: Select top N Skills
- **WHEN** LLM returns relevant Skills or keyword matching completes
- **THEN** system selects top N most relevant Skills (default: 3) for loading

#### Scenario: Cache matching results
- **WHEN** same query is matched multiple times
- **THEN** system uses cached results (LLM or keyword) to avoid repeated processing

### Requirement: Dynamic Skills management during conversation
The system SHALL intelligently manage loaded Skills during multi-turn conversations, loading new Skills and evicting irrelevant ones.

#### Scenario: Track Skills usage on each query
- **WHEN** user sends a query and Skills are matched
- **THEN** system tracks when each Skill was last matched/used

#### Scenario: Load new Skills on each query
- **WHEN** user sends a new query
- **THEN** system re-matches Skills using LLM semantic matching and loads newly matched Skills

#### Scenario: Evict Skills not matched in recent queries
- **WHEN** Skills have not been matched in last N queries (default: 3)
- **THEN** system evicts these Skills from cache, freeing up tokens

#### Scenario: Keep Skills relevant to current conversation
- **WHEN** determining which Skills to evict
- **THEN** system keeps Skills that are still relevant to current conversation context, even if not matched in current query

### Requirement: Prompt management and compression
The system SHALL manage prompt length with token monitoring and LLM-based semantic compression strategies.

#### Scenario: Estimate token count
- **WHEN** system builds prompt for LLM
- **THEN** system estimates token count for all prompt components (system prompt, conversation history, Skills content, tools)

#### Scenario: LLM compression at 80% threshold
- **WHEN** estimated token count exceeds 80% of context window
- **THEN** system uses LLM to semantically compress conversation history (moderate compression, ~50% reduction) while preserving key decisions, results, and user preferences, falling back to simple truncation if LLM compression fails

#### Scenario: Aggressive LLM compression at 90% threshold
- **WHEN** estimated token count exceeds 90% of context window after compression
- **THEN** system uses LLM for aggressive compression (~70% reduction) and evicts inactive Skills (not referenced in recent queries), keeping only active and relevant Skills

#### Scenario: Maximum LLM compression at 95% threshold
- **WHEN** estimated token count exceeds 95% of context window
- **THEN** system uses LLM for maximum compression (~80% reduction, keep only essential context), compresses both conversation history and Skills content, and keeps only active Skills

#### Scenario: Compression caching
- **WHEN** LLM compresses conversation history or Skills content
- **THEN** system caches compressed results (cache key: content hash, cache value: compressed content) to avoid re-compressing same content

#### Scenario: Skills priority management
- **WHEN** system manages Skills in prompt
- **THEN** system prioritizes: Active (referenced in current/recent queries) > Relevant (matched but not used) > Inactive (not matched)

### Requirement: Skills integration into prompt
The system SHALL integrate Skills content into LLM system prompt dynamically.

#### Scenario: Build prompt with Skills
- **WHEN** system prepares prompt for LLM
- **THEN** system includes matched Skills content in system prompt section

#### Scenario: Format Skills content
- **WHEN** Skills content is added to prompt
- **THEN** Skills are formatted as: `## Skill: <name>\n<description>\n\n<content>`

#### Scenario: Order Skills by priority
- **WHEN** multiple Skills are included in prompt
- **THEN** Skills are ordered by priority (active first, then relevant)

### Requirement: Built-in tools for Skills
The system SHALL provide built-in tools that Skills can reference.

#### Scenario: HTTP request tool
- **WHEN** Skill needs to make HTTP requests
- **THEN** system provides `http_request` tool supporting GET, POST, PUT, DELETE with headers, query parameters, and body

#### Scenario: Command execution tool
- **WHEN** Skill needs to execute shell commands
- **THEN** system provides `execute_command` tool with timeout and working directory support, restricted to safe commands (configurable allowlist)

#### Scenario: File operations tool
- **WHEN** Skill needs file operations
- **THEN** system provides `file_operations` tool for reading, writing, listing files, restricted to user config directory and current working directory

#### Scenario: Database query tool
- **WHEN** Skill needs to query database
- **THEN** system provides `execute_sql` tool that reuses existing database connection

#### Scenario: Register built-in tools
- **WHEN** application starts
- **THEN** system registers all built-in tools in tool registry

### Requirement: Skills lifecycle management
The system SHALL manage Skills loading, caching, and eviction throughout application lifecycle.

#### Scenario: Initialize Skills manager
- **WHEN** application starts
- **THEN** system initializes Skills manager and loads Skills metadata

#### Scenario: Reload Skills on demand
- **WHEN** user adds or modifies Skills
- **THEN** system requires application restart to reload Skills (no hot-reloading)

#### Scenario: Handle Skills errors gracefully
- **WHEN** Skill file is corrupted or unreadable
- **THEN** system logs error, skips the Skill, and continues with other Skills
