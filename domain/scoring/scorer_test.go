package scoring

import (
	"testing"
	"time"

	"search-engine/domain"
)

func TestScorer_CalculateScore(t *testing.T) {
	scorer := NewScorer()

	tests := []struct {
		name     string
		content  domain.ProviderContent
		minScore float64
	}{
		{
			name: "high engagement video",
			content: domain.ProviderContent{
				Type:        "video",
				Views:       100000,
				Likes:       5000,
				PublishedAt: time.Now().AddDate(0, 0, -1),
			},
			minScore: 50.0,
		},
		{
			name: "low engagement video",
			content: domain.ProviderContent{
				Type:        "video",
				Views:       100,
				Likes:       5,
				PublishedAt: time.Now().AddDate(0, 0, -1),
			},
			minScore: 0.0,
		},
		{
			name: "text content",
			content: domain.ProviderContent{
				Type:        "text",
				Reactions:   500,
				ReadingTime: 10,
				PublishedAt: time.Now().AddDate(0, 0, -1),
			},
			minScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.CalculateScore(tt.content)

			if score < tt.minScore {
				t.Errorf("CalculateScore() = %v, want >= %v", score, tt.minScore)
			}

			if score < 0 {
				t.Errorf("CalculateScore() = %v, scores should never be negative", score)
			}
		})
	}
}

func TestScorer_FreshnessDecay(t *testing.T) {
	scorer := NewScorer()

	baseContent := domain.ProviderContent{
		Type:  "video",
		Views: 1000,
		Likes: 100,
	}

	freshContent := baseContent
	freshContent.PublishedAt = time.Now().AddDate(0, 0, -1)

	oldContent := baseContent
	oldContent.PublishedAt = time.Now().AddDate(-2, 0, 0)

	freshScore := scorer.CalculateScore(freshContent)
	oldScore := scorer.CalculateScore(oldContent)

	if freshScore <= oldScore {
		t.Errorf("Fresh content score (%v) should be higher than old content score (%v)", freshScore, oldScore)
	}
}
