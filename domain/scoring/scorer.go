package scoring

import (
	"math"
	"time"

	"search-engine/domain"
)

type Scorer struct{}

func NewScorer() *Scorer {
	return &Scorer{}
}

func (s *Scorer) CalculateScore(content domain.ProviderContent) float64 {
	baseScore := s.calculateBaseScore(content)
	typeMultiplier := s.getTypeMultiplier(content.Type)
	freshnessScore := s.calculateFreshnessScore(content.PublishedAt)
	engagementScore := s.calculateEngagementScore(content)

	finalScore := (baseScore * typeMultiplier) + freshnessScore + engagementScore
	return roundTo2Decimals(finalScore)
}

func (s *Scorer) calculateBaseScore(content domain.ProviderContent) float64 {
	switch content.Type {
	case "video":
		return float64(content.Views)/1000.0 + float64(content.Likes)/100.0
	case "text":
		return float64(content.ReadingTime) + float64(content.Reactions)/50.0
	default:
		return 0
	}
}

func (s *Scorer) getTypeMultiplier(contentType string) float64 {
	switch contentType {
	case "video":
		return 1.5
	case "text":
		return 1.0
	default:
		return 1.0
	}
}

func (s *Scorer) calculateFreshnessScore(publishedAt time.Time) float64 {
	daysSincePublished := time.Since(publishedAt).Hours() / 24

	switch {
	case daysSincePublished <= 7:
		return 5.0
	case daysSincePublished <= 30:
		return 3.0
	case daysSincePublished <= 90:
		return 1.0
	default:
		return 0.0
	}
}

func (s *Scorer) calculateEngagementScore(content domain.ProviderContent) float64 {
	switch content.Type {
	case "video":
		if content.Views == 0 {
			return 0
		}
		return float64(content.Likes) / float64(content.Views) * 10.0
	case "text":
		if content.ReadingTime == 0 {
			return 0
		}
		return float64(content.Reactions) / float64(content.ReadingTime) * 5.0
	default:
		return 0
	}
}

func roundTo2Decimals(val float64) float64 {
	return math.Round(val*100) / 100
}
