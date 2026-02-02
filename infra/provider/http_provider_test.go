package provider

import (
	"testing"
)

func TestMapContentType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"video lowercase", "video", "video"},
		{"video uppercase", "VIDEO", "video"},
		{"video mixed", "ViDeO", "video"},
		{"text lowercase", "text", "text"},
		{"text uppercase", "TEXT", "text"},
		{"article to text", "article", "text"},
		{"article uppercase", "ARTICLE", "text"},
		{"unknown type", "unknown", "unknown"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapContentType(tt.input)
			if got != tt.expected {
				t.Errorf("mapContentType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMapContentType_Normalization(t *testing.T) {
	articleVariants := []string{"article", "Article", "ARTICLE", "aRtIcLe"}

	for _, variant := range articleVariants {
		result := mapContentType(variant)
		if result != "text" {
			t.Errorf("mapContentType(%q) = %q, want 'text'", variant, result)
		}
	}
}
