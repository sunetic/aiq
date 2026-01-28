## ADDED Requirements

### Requirement: LLM-based semantic compression for conversation history
The system SHALL use LLM to semantically compress conversation history when context reaches compression thresholds, preserving key information while reducing token usage.

#### Scenario: Trigger LLM compression at 80% threshold
- **WHEN** estimated token count exceeds 80% of context window
- **THEN** system sends conversation history to LLM for semantic compression, preserving key decisions, results, and user preferences

#### Scenario: LLM compresses conversation history
- **WHEN** LLM receives conversation history for compression
- **THEN** LLM returns compressed summary that maintains essential context while reducing token count (target: ~50% reduction)

#### Scenario: Preserve key information during compression
- **WHEN** LLM compresses conversation history
- **THEN** LLM preserves: user preferences, important decisions, query results, active context, and removes: redundant information, outdated context, irrelevant details

#### Scenario: Replace original history with compressed version
- **WHEN** LLM compression completes successfully
- **THEN** system replaces original conversation history with compressed version in prompt

#### Scenario: Fallback to simple truncation
- **WHEN** LLM compression fails (API error, timeout, invalid response)
- **THEN** system falls back to simple truncation (keep last N messages)

### Requirement: LLM-based semantic compression for Skills content
The system SHALL use LLM to compress Skills content when context reaches compression thresholds, extracting only relevant parts while maintaining essential information.

#### Scenario: Compress Skills content at threshold
- **WHEN** estimated token count exceeds compression threshold and Skills content is large
- **THEN** system sends Skills content to LLM for compression, extracting relevant parts and removing redundancy

#### Scenario: LLM compresses Skills content
- **WHEN** LLM receives Skills content for compression
- **THEN** LLM returns compressed version that maintains essential information while reducing token count

#### Scenario: Preserve essential Skills information
- **WHEN** LLM compresses Skills content
- **THEN** LLM preserves: relevant instructions, key examples, important context, and removes: redundant explanations, outdated information, irrelevant details

### Requirement: Dynamic Skills management during conversation
The system SHALL intelligently manage loaded Skills during multi-turn conversations, loading new Skills and evicting irrelevant ones.

#### Scenario: Track Skills usage
- **WHEN** user sends a query and Skills are matched
- **THEN** system tracks when each Skill was last matched/used

#### Scenario: Load new Skills on each query
- **WHEN** user sends a new query
- **THEN** system re-matches Skills and loads newly matched Skills

#### Scenario: Evict irrelevant Skills
- **WHEN** Skills have not been matched in last N queries (default: 3)
- **THEN** system evicts these Skills from cache, freeing up tokens

#### Scenario: Keep Skills relevant to current conversation
- **WHEN** determining which Skills to evict
- **THEN** system keeps Skills that are still relevant to current conversation context, even if not matched in current query

#### Scenario: Prioritize Skills by usage
- **WHEN** managing Skills during conversation
- **THEN** system prioritizes: frequently used Skills > recently used Skills > low-priority Skills

### Requirement: Compression caching
The system SHALL cache LLM compression results to avoid re-compressing the same content.

#### Scenario: Cache compression results
- **WHEN** LLM compresses conversation history or Skills content
- **THEN** system caches the compressed result (cache key: content hash, cache value: compressed content)

#### Scenario: Use cached compression
- **WHEN** same content needs to be compressed again
- **THEN** system uses cached compressed version instead of calling LLM again

### Requirement: Compression prompt design
The system SHALL provide clear prompts to LLM for compression that specify target compression ratio and information to preserve.

#### Scenario: Compression prompt includes instructions
- **WHEN** system prepares LLM compression request
- **THEN** prompt includes: target compression ratio (e.g., reduce to 50% of original), information to preserve (user preferences, decisions, results, active context), information to remove (redundancy, outdated context, irrelevant details)

#### Scenario: LLM returns structured compression
- **WHEN** LLM processes compression request
- **THEN** LLM returns compressed content in structured format (JSON or plain text) that maintains essential context
