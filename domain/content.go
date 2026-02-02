package domain

import (
	"time"

	"github.com/google/uuid"
)

type ContentType string

const (
	ContentTypeVideo ContentType = "video"
	ContentTypeText  ContentType = "text"
)

func (c ContentType) IsValid() bool {
	return c == ContentTypeVideo || c == ContentTypeText
}

type Content struct {
	ID          uuid.UUID   `json:"id"`
	ExternalID  string      `json:"external_id"`
	Provider    string      `json:"provider"`
	Title       string      `json:"title"`
	Type        ContentType `json:"type"`
	PublishedAt time.Time   `json:"published_at"`
	RawData     []byte      `json:"-"`

	Views       int      `json:"views,omitempty"`
	Likes       int      `json:"likes,omitempty"`
	Reactions   int      `json:"reactions,omitempty"`
	ReadingTime int      `json:"reading_time,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	Score float64 `json:"score"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUUID() uuid.UUID {
	return uuid.New()
}

type ScoreBreakdown struct {
	BaseScore       float64 `json:"base_score"`
	TypeMultiplier  float64 `json:"type_multiplier"`
	FreshnessScore  float64 `json:"freshness_score"`
	EngagementScore float64 `json:"engagement_score"`
}

type ContentWithScore struct {
	Content
	ScoreBreakdown ScoreBreakdown  `json:"score_breakdown"`
	Metadata       ContentMetadata `json:"metadata"`
}

type ContentMetadata struct {
	Views       int `json:"views,omitempty"`
	Likes       int `json:"likes,omitempty"`
	Reactions   int `json:"reactions,omitempty"`
	ReadingTime int `json:"reading_time,omitempty"`
}

type ProviderContent struct {
	ExternalID  string    `validate:"required,min=1"`
	Title       string    `validate:"required,min=1,max=500"`
	Type        string    `validate:"required,oneof=video text"`
	PublishedAt time.Time `validate:"required"`
	Views       int       `validate:"gte=0"`
	Likes       int       `validate:"gte=0"`
	ReadingTime int       `validate:"gte=0"`
	Reactions   int       `validate:"gte=0"`
	Tags        []string
	RawData     []byte
}
