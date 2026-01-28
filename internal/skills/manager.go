package skills

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager manages Skills lifecycle: loading, caching, and eviction
type Manager struct {
	// Metadata loaded on startup (lightweight)
	metadataList []*Metadata

	// Cache of loaded Skills (full content)
	cache map[string]*Skill

	// Usage tracking: track when each Skill was last matched/used
	usageHistory map[string][]time.Time // Skill name -> list of usage timestamps
	queryHistory []string                // Recent queries for context relevance

	// Mutex for thread-safe access
	mu sync.RWMutex
}

const (
	// DefaultEvictionQueries is the default number of queries to check for eviction
	DefaultEvictionQueries = 3
	// MaxQueryHistory is the maximum number of recent queries to keep for context
	MaxQueryHistory = 10
)

// NewManager creates a new Skills manager
func NewManager() *Manager {
	return &Manager{
		metadataList: []*Metadata{},
		cache:        make(map[string]*Skill),
		usageHistory: make(map[string][]time.Time),
		queryHistory: make([]string, 0, MaxQueryHistory),
	}
}

// Initialize loads Skills metadata on startup
func (m *Manager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	metadata, err := LoadSkillsMetadata()
	if err != nil {
		return fmt.Errorf("failed to load skills metadata: %w", err)
	}

	m.metadataList = metadata
	if len(metadata) > 0 {
		log.Printf("Loaded %d skill(s) metadata (progressive loading enabled)", len(metadata))
		for _, md := range metadata {
			log.Printf("  - %s: %s", md.Name, md.Description)
		}
	}

	return nil
}

// GetMetadata returns all Skills metadata
func (m *Manager) GetMetadata() []*Metadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*Metadata, len(m.metadataList))
	copy(result, m.metadataList)
	return result
}

// LoadSkill loads a Skill by name, using cache if available
func (m *Manager) LoadSkill(name string) (*Skill, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check cache first
	if skill, exists := m.cache[name]; exists && skill.Loaded {
		return skill, nil
	}

	// Find metadata
	var metadata *Metadata
	for _, md := range m.metadataList {
		if md.Name == name {
			metadata = md
			break
		}
	}

	if metadata == nil {
		return nil, fmt.Errorf("skill '%s' not found", name)
	}

	// Load full content (progressive loading: only now do we read the full file)
	log.Printf("Loading full content for skill: %s", name)
	skill, err := LoadSkillContent(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to load skill content: %w", err)
	}

	// Cache it
	m.cache[name] = skill
	log.Printf("Skill '%s' loaded and cached (content length: %d chars)", name, len(skill.Content))

	return skill, nil
}

// LoadSkills loads multiple Skills by name
func (m *Manager) LoadSkills(names []string) ([]*Skill, error) {
	var skills []*Skill
	var errors []error

	for _, name := range names {
		skill, err := m.LoadSkill(name)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		skills = append(skills, skill)
	}

	if len(errors) > 0 {
		return skills, fmt.Errorf("failed to load some skills: %v", errors)
	}

	return skills, nil
}

// SetPriority sets the priority of a Skill
func (m *Manager) SetPriority(name string, priority Priority) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if skill, exists := m.cache[name]; exists {
		skill.Priority = priority
	}
}

// TrackUsage records that a Skill was matched/used for a query
func (m *Manager) TrackUsage(skillName string, query string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if _, exists := m.usageHistory[skillName]; !exists {
		m.usageHistory[skillName] = make([]time.Time, 0)
	}
	m.usageHistory[skillName] = append(m.usageHistory[skillName], now)

	// Track query history for context relevance
	m.queryHistory = append(m.queryHistory, query)
	if len(m.queryHistory) > MaxQueryHistory {
		m.queryHistory = m.queryHistory[1:]
	}
}

// EvictUnusedSkills evicts Skills that haven't been matched in the last N queries
func (m *Manager) EvictUnusedSkills(numQueries int) []string {
	if numQueries <= 0 {
		numQueries = DefaultEvictionQueries
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	evicted := make([]string, 0)

	// Check each cached Skill
	for name, skill := range m.cache {
		// Check if Skill was used in recent queries
		usageTimes := m.usageHistory[name]
		if len(usageTimes) == 0 {
			// Never used, evict if low priority
			if skill.Priority == PriorityInactive {
				delete(m.cache, name)
				delete(m.usageHistory, name)
				evicted = append(evicted, name)
			}
			continue
		}

		// Check if used in last N queries
		// We approximate this by checking if last usage was recent enough
		// (assuming queries happen within reasonable time)
		recentUsageCount := 0
		for i := len(usageTimes) - 1; i >= 0 && recentUsageCount < numQueries; i-- {
			if time.Since(usageTimes[i]) < 10*time.Minute { // Consider queries within 10 minutes as "recent"
				recentUsageCount++
			} else {
				break
			}
		}

		// Also check if Skill is still relevant to current conversation context
		isRelevant := m.isRelevantToContext(name)

		// Evict if not used recently AND not relevant to context AND low priority
		if recentUsageCount == 0 && !isRelevant && skill.Priority < PriorityActive {
			delete(m.cache, name)
			// Keep usage history for a while (don't delete immediately)
			evicted = append(evicted, name)
		}
	}

	return evicted
}

// isRelevantToContext checks if a Skill is still relevant to current conversation context
func (m *Manager) isRelevantToContext(skillName string) bool {
	// Simple heuristic: check if Skill name or keywords appear in recent queries
	skillLower := fmt.Sprintf("%s", skillName) // Could enhance with metadata keywords
	for _, query := range m.queryHistory {
		queryLower := fmt.Sprintf("%s", query)
		// Simple substring match (could be enhanced with better matching)
		if len(queryLower) > 0 && len(skillLower) > 0 {
			// Check if skill name appears in query (case-insensitive substring)
			if len(queryLower) >= len(skillLower) {
				for i := 0; i <= len(queryLower)-len(skillLower); i++ {
					if queryLower[i:i+len(skillLower)] == skillLower {
						return true
					}
				}
			}
		}
	}
	return false
}

// GetUsageCount returns the number of times a Skill was used
func (m *Manager) GetUsageCount(skillName string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if usageTimes, exists := m.usageHistory[skillName]; exists {
		return len(usageTimes)
	}
	return 0
}

// EvictSkill removes a Skill from cache (for prompt compression)
func (m *Manager) EvictSkill(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.cache, name)
}

// EvictLowPrioritySkills removes Skills with priority lower than the given priority
func (m *Manager) EvictLowPrioritySkills(minPriority Priority) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, skill := range m.cache {
		if skill.Priority < minPriority {
			delete(m.cache, name)
		}
	}
}

// GetCachedSkills returns all currently cached Skills
func (m *Manager) GetCachedSkills() []*Skill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	skills := make([]*Skill, 0, len(m.cache))
	for _, skill := range m.cache {
		skills = append(skills, skill)
	}
	return skills
}
