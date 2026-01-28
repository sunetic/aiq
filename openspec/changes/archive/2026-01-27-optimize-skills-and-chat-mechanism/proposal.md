## Why

The current Skills matching mechanism is not precise enough, causing irrelevant Skills to be loaded (e.g., loading seekdb-docs when installing MySQL). Additionally, the chat mode requires a database source to be selected, preventing users from entering chat mode when no sources are configured. The command prompt also lacks visibility into the current source, making it unclear which database context is active.

During multi-turn conversations, Skills accumulate without eviction, leading to token waste. The current compression mechanism only truncates conversation history and evicts Skills, losing important context. There's no intelligent mechanism to compress context using LLM semantic understanding.

## What Changes

- **Improve Skills matching precision**: Use LLM-based semantic matching instead of keyword-based scoring to better filter irrelevant Skills (e.g., "install mysql" should not match seekdb-docs).
- **Add free chat mode**: Allow users to enter chat mode without requiring a database source. In free mode, only general conversation and Skills-based operations are available (no SQL execution).
- **Enhance command prompt**: Display current source information in the prompt (e.g., `aiq[source-name]> `) to provide better context awareness.
- **Dynamic Skills management**: Intelligently manage Skills during conversation - load new Skills on each query, evict Skills not matched in recent queries, keep Skills relevant to current conversation context.
- **LLM-based context compression**: When context reaches thresholds, use LLM to semantically compress conversation history and Skills content, preserving key information while reducing token usage.

## Capabilities

### New Capabilities
- `skills-matching-optimization`: LLM-based semantic matching for Skills with improved relevance judgment
- `free-chat-mode`: Chat mode that works without requiring a database source, supporting general conversation and Skills-based operations
- `llm-context-compression`: LLM-based semantic compression for conversation history and Skills content
- `dynamic-skills-management`: Intelligent Skills loading and eviction during multi-turn conversations

### Modified Capabilities
- `claude-skills-support`: Update Skills matching algorithm to use LLM semantic matching, add dynamic Skills management during conversation
- `sql-interactive-mode`: Add optional source selection, support free mode without database connection, and enhance prompt display with source information

## Impact

- **Skills matching**: Changes to `internal/skills/matcher.go` to implement LLM-based semantic matching
- **Skills management**: Changes to `internal/skills/manager.go` to add usage tracking and eviction logic
- **Prompt compression**: Changes to `internal/prompt/compressor.go` to implement LLM-based semantic compression
- **Chat mode**: Changes to `internal/sql/mode.go` to support optional source selection and free mode
- **Tool handler**: Changes to `internal/sql/tool_handler.go` to implement dynamic Skills management during conversation
- **UI prompt**: Changes to `internal/ui/` or `internal/sql/mode.go` to display source information in prompt
- **User experience**: Better Skills relevance, more flexible chat access, intelligent context compression, and efficient token usage
