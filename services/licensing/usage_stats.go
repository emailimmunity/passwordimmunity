package licensing

import (
	"sync"
	"time"
)

// FeatureUsageStats tracks usage statistics for licensed features
type FeatureUsageStats struct {
	FeatureID      string    `json:"feature_id"`
	LastUsed       time.Time `json:"last_used"`
	UsageCount     int64     `json:"usage_count"`
	ActiveSessions int       `json:"active_sessions"`
}

// UsageTracker manages feature usage statistics
type UsageTracker struct {
	mu    sync.RWMutex
	stats map[string]map[string]*FeatureUsageStats // orgID -> featureID -> stats
}

// NewUsageTracker creates a new usage tracking instance
func NewUsageTracker() *UsageTracker {
	return &UsageTracker{
		stats: make(map[string]map[string]*FeatureUsageStats),
	}
}

// TrackFeatureUsage records feature usage for an organization
func (t *UsageTracker) TrackFeatureUsage(orgID, featureID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.stats[orgID]; !exists {
		t.stats[orgID] = make(map[string]*FeatureUsageStats)
	}

	if _, exists := t.stats[orgID][featureID]; !exists {
		t.stats[orgID][featureID] = &FeatureUsageStats{
			FeatureID: featureID,
		}
	}

	stats := t.stats[orgID][featureID]
	stats.LastUsed = time.Now()
	stats.UsageCount++
	stats.ActiveSessions++
}

// EndFeatureUsage records the end of a feature usage session
func (t *UsageTracker) EndFeatureUsage(orgID, featureID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if stats, exists := t.stats[orgID][featureID]; exists {
		if stats.ActiveSessions > 0 {
			stats.ActiveSessions--
		}
	}
}

// GetUsageStats returns usage statistics for an organization
func (t *UsageTracker) GetUsageStats(orgID string) map[string]*FeatureUsageStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if stats, exists := t.stats[orgID]; exists {
		// Create a copy to avoid external modifications
		result := make(map[string]*FeatureUsageStats)
		for featureID, stats := range stats {
			statsCopy := *stats
			result[featureID] = &statsCopy
		}
		return result
	}
	return make(map[string]*FeatureUsageStats)
}
