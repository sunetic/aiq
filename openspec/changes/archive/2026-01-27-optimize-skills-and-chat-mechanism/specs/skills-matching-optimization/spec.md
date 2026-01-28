## ADDED Requirements

### Requirement: LLM-based semantic matching for Skills
The system SHALL use LLM to perform semantic relevance judgment for matching Skills to user queries, instead of relying solely on keyword-based scoring.

#### Scenario: Semantic matching via LLM
- **WHEN** user submits a query and system needs to match Skills
- **THEN** system sends user query and all Skills metadata (name + description) to LLM for semantic relevance judgment

#### Scenario: LLM returns relevant Skills
- **WHEN** LLM receives user query and Skills metadata
- **THEN** LLM returns list of relevant Skill names (top N, default 3) based on semantic understanding

#### Scenario: LLM distinguishes semantic context
- **WHEN** user query is "install mysql" and both "init-mysql-mac" and "seekdb-docs" Skills exist
- **THEN** LLM understands that "install mysql" is about installation, not database usage, and returns only "init-mysql-mac"

#### Scenario: Fallback to keyword matching
- **WHEN** LLM semantic matching fails (API error, timeout, invalid response)
- **THEN** system falls back to keyword-based matching algorithm

#### Scenario: Cache LLM matching results
- **WHEN** same query is matched multiple times
- **THEN** system uses cached LLM results to avoid repeated API calls

### Requirement: LLM matching prompt design
The system SHALL provide a clear prompt to LLM for semantic matching that includes user query and Skills metadata.

#### Scenario: Prompt includes user query
- **WHEN** system prepares LLM matching request
- **THEN** prompt includes the user's query text

#### Scenario: Prompt includes Skills metadata
- **WHEN** system prepares LLM matching request
- **THEN** prompt includes all available Skills with their names and descriptions

#### Scenario: LLM returns structured response
- **WHEN** LLM processes matching request
- **THEN** LLM returns JSON array of relevant Skill names or structured response format

### Requirement: Keyword matching as fallback
The system SHALL maintain keyword-based matching as a fallback mechanism when LLM matching is unavailable.

#### Scenario: Fallback on LLM failure
- **WHEN** LLM matching fails or times out
- **THEN** system uses existing keyword-based matching algorithm

#### Scenario: Fallback maintains functionality
- **WHEN** keyword matching is used as fallback
- **THEN** system still returns top N most relevant Skills based on keyword scores
