package domain

import (
	"testing"
	"time"
)

func TestContentType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		ct       ContentType
		expected bool
	}{
		{"valid video", ContentTypeVideo, true},
		{"valid text", ContentTypeText, true},
		{"invalid empty", ContentType(""), false},
		{"invalid random", ContentType("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidateProviderContent(t *testing.T) {
	validContent := ProviderContent{
		ExternalID:  "test-123",
		Title:       "Test Title",
		Type:        "video",
		PublishedAt: time.Now(),
		Views:       100,
		Likes:       50,
	}

	tests := []struct {
		name    string
		content ProviderContent
		wantErr bool
	}{
		{
			name:    "valid content",
			content: validContent,
			wantErr: false,
		},
		{
			name: "missing external_id",
			content: ProviderContent{
				Title:       "Test",
				Type:        "video",
				PublishedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing title",
			content: ProviderContent{
				ExternalID:  "test-123",
				Type:        "video",
				PublishedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			content: ProviderContent{
				ExternalID:  "test-123",
				Title:       "Test",
				Type:        "invalid",
				PublishedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "negative views",
			content: ProviderContent{
				ExternalID:  "test-123",
				Title:       "Test",
				Type:        "video",
				PublishedAt: time.Now(),
				Views:       -10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProviderContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProviderContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
