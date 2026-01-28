## 1. LLM-Based Semantic Matching for Skills

- [x] 1.1 Create `MatchWithLLM()` method in `internal/skills/matcher.go` that sends user query + Skills metadata to LLM
- [x] 1.2 Design prompt template for LLM semantic matching (include user query and Skills list with names/descriptions)
- [x] 1.3 Implement LLM response parsing to extract list of relevant Skill names (JSON array format)
- [x] 1.4 Add caching mechanism for LLM matching results (cache key: query hash, cache value: Skill names)
- [x] 1.5 Update `Match()` method to call `MatchWithLLM()` first, fallback to keyword matching on failure
- [x] 1.6 Add error handling for LLM API failures, timeouts, and invalid response formats
- [x] 1.7 Keep existing keyword matching as fallback (no changes to keyword extraction logic)
- [x] 1.8 Add unit tests for LLM semantic matching (mock LLM client)
- [x] 1.9 Add unit tests for fallback to keyword matching when LLM fails
- [x] 1.10 Add integration tests for semantic matching accuracy (e.g., "install mysql" should not match seekdb-docs)

## 2. Free Chat Mode Implementation

- [x] 2.1 Modify `RunSQLMode()` in `internal/sql/mode.go` to make source selection optional
- [x] 2.2 Update source selection logic to allow skipping (add "Skip (free mode)" option to menu)
- [x] 2.3 Handle case when no sources are configured: enter free mode automatically
- [x] 2.4 Conditionally initialize database connection only if source exists
- [x] 2.5 Update `NewToolHandler()` to accept optional connection (nil for free mode)
- [x] 2.6 Modify `GetLLMFunctionsWithBuiltin()` to conditionally exclude `execute_sql` tool when connection is nil
- [x] 2.7 Update system prompt generation to indicate free mode vs database mode
- [x] 2.8 Add error handling for SQL execution attempts in free mode
- [x] 2.9 Update session management to handle free mode (no source in session metadata)

## 3. Enhanced Command Prompt

- [x] 3.1 Modify prompt generation in `RunSQLMode()` to include source name: `aiq[source-name]> ` format
- [x] 3.2 Update readline config to use dynamic prompt based on source availability
- [x] 3.3 Handle prompt display when source is nil (free mode): show `aiq> ` without source name
- [x] 3.4 Ensure prompt updates correctly when source changes (if applicable in future)

## 4. Testing and Validation

- [x] 4.1 Test LLM semantic matching accuracy (verify irrelevant Skills are filtered, e.g., "install mysql" should not match seekdb-docs)
- [x] 4.2 Test fallback to keyword matching when LLM fails
- [x] 4.3 Test caching mechanism for LLM matching results
- [x] 4.4 Test free chat mode functionality (general conversation, Skills operations, no SQL)
- [x] 4.5 Test prompt display in both database mode and free mode
- [x] 4.6 Test error handling when SQL is attempted in free mode
- [x] 4.7 Test backward compatibility (existing Skills still work with LLM matching)
- [x] 4.8 Integration test: enter chat mode without sources, verify free mode works
- [x] 4.9 Integration test: enter chat mode with sources, verify database mode works
- [x] 4.10 Performance test: measure LLM matching latency vs keyword matching

## 5. Dynamic Skills Management During Conversation

- [x] 5.1 Add Skills usage tracking in `Manager` (track when each Skill was last matched/used)
- [x] 5.2 Modify `HandleToolCallLoop()` to track Skills usage on each query
- [x] 5.3 Implement eviction logic: evict Skills not matched in last N queries (default: 3)
- [x] 5.4 Update Skills priority management: keep Skills relevant to current conversation context
- [x] 5.5 Add method to determine if a Skill is still relevant to current conversation (check recent queries)
- [x] 5.6 Modify Skills loading logic: load new Skills, evict irrelevant ones before loading
- [x] 5.7 Add unit tests for Skills usage tracking and eviction logic
- [x] 5.8 Add integration tests for multi-turn conversation with Skills loading/eviction

## 6. LLM-Based Context Compression

- [x] 6.1 Create `CompressWithLLM()` method in `internal/prompt/compressor.go` that uses LLM for semantic compression
- [x] 6.2 Design compression prompt template for conversation history (preserve key decisions, results, preferences)
- [x] 6.3 Design compression prompt template for Skills content (extract relevant parts, remove redundancy)
- [x] 6.4 Implement LLM compression for conversation history (summarize while preserving key context)
- [x] 6.5 Implement LLM compression for Skills content (compress while maintaining essential information)
- [x] 6.6 Update `Compress()` method to trigger LLM compression at 80% threshold (moderate compression)
- [x] 6.7 Add aggressive LLM compression at 90% threshold (more aggressive summarization)
- [x] 6.8 Add maximum LLM compression at 95% threshold (keep only essential context)
- [x] 6.9 Add caching for LLM compression results (avoid re-compressing same content)
- [x] 6.10 Add fallback to simple truncation if LLM compression fails
- [x] 6.11 Add unit tests for LLM compression (mock LLM client, verify compression quality)
- [x] 6.12 Add integration tests for context compression during long conversations

## 7. Documentation Updates

- [x] 7.1 Update README.md to document free chat mode
- [x] 7.2 Update README.md to explain LLM-based semantic matching for Skills
- [x] 7.3 Update README.md to document dynamic Skills management and LLM compression
- [x] 7.4 Update README_CN.md with same changes
- [x] 7.5 Document LLM matching behavior and fallback mechanism
- [x] 7.6 Document Skills eviction strategy and compression thresholds
