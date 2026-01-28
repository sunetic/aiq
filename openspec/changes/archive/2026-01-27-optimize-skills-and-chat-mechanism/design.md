## Context

The current Skills matching system uses a simple keyword-based scoring algorithm that can match Skills with very low relevance scores (e.g., score 10 from a single keyword match in description). This causes irrelevant Skills to be loaded, wasting tokens and potentially confusing the LLM.

The chat mode currently requires a database source to be selected before entering, which prevents users from using chat mode for general conversation or Skills-based operations when no database is configured. This is especially problematic for first-time users who haven't configured a source yet.

The command prompt (`aiq> `) doesn't show which source is currently active, making it unclear what database context the user is working with.

## Goals / Non-Goals

**Goals:**
- Improve Skills matching precision by filtering out low-relevance Skills (score < threshold)
- Allow chat mode to work without a database source (free mode)
- Display current source in command prompt for better context awareness
- Maintain backward compatibility with existing Skills and chat behavior

**Non-Goals:**
- Complete rewrite of Skills matching algorithm (incremental improvement)
- Embedding-based matching or vector databases (use LLM for semantic judgment instead)
- Multi-source switching within a single chat session
- Complex prompt customization

## Decisions

### 1. LLM-Based Semantic Matching for Skills

**Decision**: Use LLM to perform semantic relevance judgment instead of keyword-based scoring with threshold.

**Rationale**: 
- Keyword matching is prone to false positives (e.g., "mysql" matches seekdb-docs because seekdb is MySQL-compatible)
- LLM can understand semantic context and distinguish between "installing MySQL" vs "using seekdb (MySQL-compatible)"
- More accurate matching reduces token waste and improves LLM responses
- Leverages existing LLM infrastructure, no need for embeddings or vector DB

**Alternatives Considered**:
- Keyword threshold (original plan): Simple but inaccurate, still causes false matches
- Embedding-based similarity: Requires vector DB, adds complexity and dependencies
- LLM semantic judgment: Best balance of accuracy and simplicity, uses existing infrastructure

**Implementation**:
- Send user query and all Skills metadata (name + description) to LLM
- Ask LLM to return list of relevant Skill names (top N, default 3)
- LLM response format: JSON array of Skill names or structured response
- Cache LLM results per query to avoid repeated calls
- Fallback to keyword matching if LLM call fails

### 2. Free Chat Mode Without Database Source

**Decision**: Make source selection optional in chat mode. If no source is selected, run in "free mode" with limited capabilities.

**Rationale**:
- Enables general conversation and Skills-based operations without database
- Better first-time user experience
- Allows using Skills for non-database tasks (e.g., system operations, file operations)

**Free Mode Capabilities**:
- General conversation with LLM
- Skills-based operations (execute_command, http_request, file_operations)
- No SQL execution (execute_sql tool not available)
- No database schema context

**Alternatives Considered**:
- Always require source: Too restrictive, blocks legitimate use cases
- Separate "general chat" mode: Adds complexity, splits functionality
- Optional source with graceful degradation: Best balance of flexibility and clarity

**Implementation**:
- Modify `RunSQLMode()` to accept optional source
- If no sources available or user chooses to skip, set `src = nil`
- Conditionally initialize database connection only if source exists
- Pass `nil` connection to tool handler in free mode
- Tool handler skips `execute_sql` tool registration if connection is nil
- Update system prompt to indicate free mode vs database mode

### 3. Enhanced Command Prompt with Source Display

**Decision**: Display current source name in prompt: `aiq[source-name]> ` or `aiq> ` (if no source).

**Rationale**:
- Provides immediate context about active database
- Helps users understand what context they're working in
- Simple visual indicator, no additional complexity

**Alternatives Considered**:
- Full source details in prompt: Too verbose, clutters prompt
- Separate status line: More complex, requires screen management
- Source name only: Clean, concise, sufficient

**Implementation**:
- Modify prompt generation in `RunSQLMode()`: `aiq[%s]> ` format
- Use source name if available, empty string if free mode
- Update readline config to use dynamic prompt

### 4. Fallback Keyword Matching

**Decision**: Keep keyword-based matching as fallback when LLM semantic matching fails or is unavailable.

**Rationale**:
- Provides resilience if LLM call fails
- Faster for simple cases (though less accurate)
- Can be used as initial filter before LLM call to reduce LLM input size

**Implementation**:
- Use existing keyword matching as fallback
- Only invoke LLM for semantic matching when available
- If LLM fails, fall back to keyword matching with existing algorithm

### 5. Dynamic Skills Management During Conversation

**Decision**: When user sends new input, re-match Skills and intelligently manage loaded Skills (load new ones, evict irrelevant ones).

**Rationale**:
- Current implementation accumulates Skills without eviction, leading to token waste
- New queries may require different Skills than previous ones
- Need to balance keeping relevant Skills vs. loading new ones

**Implementation**:
- On each user input, re-match Skills using LLM semantic matching
- Load newly matched Skills
- Evict Skills that are no longer relevant (not matched in recent queries)
- Track Skills usage frequency/recency to determine eviction priority
- Keep Skills that are still relevant to current conversation context

**Eviction Strategy**:
- Skills not matched in last N queries (configurable, default 3) are candidates for eviction
- Low-priority Skills (PriorityRelevant) are evicted before high-priority ones (PriorityActive)
- Skills that are still relevant to current conversation are kept even if not matched in current query

### 6. LLM-Based Context Compression

**Decision**: When context reaches a threshold (e.g., 80% of context window), use LLM to semantically compress conversation history and Skills content.

**Rationale**:
- Current compression only truncates history and evicts Skills, losing important context
- LLM can summarize and compress information while preserving key details
- Better than simple truncation: maintains conversation continuity and important context

**Implementation**:
- When token usage exceeds threshold (e.g., 80% of context window), trigger LLM compression
- Send conversation history and loaded Skills content to LLM
- Ask LLM to:
  - Summarize conversation history (preserve key decisions, results, user preferences)
  - Compress Skills content (extract only relevant parts, remove redundant information)
  - Return compressed version that maintains essential context
- Replace original history/Skills with compressed versions
- Cache compression results to avoid re-compressing same content

**Compression Prompt Design**:
- Clear instructions: compress while preserving key information
- Specify target compression ratio (e.g., reduce to 50% of original)
- Ask LLM to preserve: user preferences, important decisions, query results, active context
- Ask LLM to remove: redundant information, outdated context, irrelevant details

**Compression Thresholds**:
- 80% threshold: Start LLM compression (moderate compression)
- 90% threshold: Aggressive compression (more aggressive summarization)
- 95% threshold: Maximum compression (keep only essential context)

## Risks / Trade-offs

**[Risk] LLM semantic matching adds latency**
- **Mitigation**: Cache LLM results per query, use lightweight prompt, consider async matching
- **Trade-off**: Slightly slower than keyword matching, but much more accurate

**[Risk] LLM matching fails or returns invalid format**
- **Mitigation**: Fallback to keyword matching, validate LLM response format, handle errors gracefully
- **Trade-off**: Need robust error handling, but provides resilience

**[Risk] Free mode confusion**
- **Mitigation**: Clear messaging about mode limitations, prompt indicates free mode, error messages explain when SQL is attempted
- **Trade-off**: Users might try SQL in free mode, but error messages will guide them

**[Risk] Prompt display complexity**
- **Mitigation**: Simple string formatting, fallback to default prompt if source name unavailable
- **Trade-off**: Slightly more complex prompt generation, but minimal impact

**[Risk] Backward compatibility**
- **Mitigation**: Default threshold allows existing behavior, free mode is opt-in (no source = free mode)
- **Trade-off**: Existing Skills may need better descriptions to meet threshold, but improves overall quality

**[Risk] LLM compression may lose important context**
- **Mitigation**: Clear compression prompts, preserve key information, test compression quality
- **Trade-off**: Some detail may be lost, but maintains conversation continuity better than truncation

**[Risk] Dynamic Skills eviction may remove still-relevant Skills**
- **Mitigation**: Track usage frequency/recency, keep Skills relevant to current conversation context
- **Trade-off**: May keep some Skills longer than needed, but avoids losing relevant context

## Migration Plan

1. **Phase 1: LLM-Based Semantic Matching**
   - Create LLM-based matching function that sends query + Skills metadata to LLM
   - Implement prompt for LLM to judge relevance and return top N Skills
   - Add caching for LLM matching results
   - Implement fallback to keyword matching if LLM fails
   - Test semantic matching accuracy vs keyword matching

2. **Phase 2: Free Chat Mode**
   - Modify `RunSQLMode()` to make source optional
   - Update tool handler to conditionally register `execute_sql`
   - Update system prompt for free mode
   - Test free mode functionality

3. **Phase 3: Enhanced Prompt**
   - Update prompt generation to include source name
   - Test prompt display in both modes
   - Ensure graceful fallback

4. **Phase 4: Dynamic Skills Management**
   - Implement Skills usage tracking (frequency/recency)
   - Add eviction logic for irrelevant Skills
   - Test Skills loading/eviction during multi-turn conversations

5. **Phase 5: LLM-Based Context Compression**
   - Implement LLM compression function for conversation history
   - Implement LLM compression function for Skills content
   - Add compression triggers at token thresholds
   - Test compression quality and context preservation

6. **Rollback Strategy**:
   - If issues arise, revert LLM matching to keyword matching
   - Revert source requirement check (require source again)
   - Revert prompt changes (back to `aiq> `)
   - Disable LLM compression, fall back to simple truncation

## Open Questions

1. Should minimum score threshold be configurable per-Skill or globally?
   - **Decision**: Start with global threshold, consider per-Skill if needed

2. Should free mode allow switching to database mode mid-session?
   - **Decision**: Not in initial implementation, can be added later if needed

3. What should happen if user tries to execute SQL in free mode?
   - **Decision**: Clear error message suggesting to select a source or configure one
