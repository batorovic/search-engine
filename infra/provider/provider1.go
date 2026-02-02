package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"search-engine/domain"
	"search-engine/infra/httpclient"

	"go.uber.org/zap"
)

func init() {
	Register("json", func(name, url string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) ContentProvider {
		return NewProvider1(name, url, timeout, client, logger)
	})
}

type Provider1 struct {
	BaseHTTPProvider
	logger *zap.Logger
}

type Provider1Response struct {
	Contents   []Provider1Content `json:"contents"`
	Pagination struct {
		Total   int `json:"total"`
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
	} `json:"pagination"`
}

type Provider1Content struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Type        string           `json:"type"`
	Metrics     Provider1Metrics `json:"metrics"`
	PublishedAt string           `json:"published_at"`
	Tags        []string         `json:"tags"`
}

type Provider1Metrics struct {
	Views    int    `json:"views"`
	Likes    int    `json:"likes"`
	Duration string `json:"duration"`
}

func NewProvider1(name, baseURL string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) *Provider1 {
	return &Provider1{
		BaseHTTPProvider: NewBaseHTTPProvider(BaseHTTPProviderConfig{
			Name:    name,
			BaseURL: baseURL,
			Timeout: timeout,
			Client:  client,
			Logger:  logger,
		}),
		logger: logger,
	}
}

func (p *Provider1) Search(ctx context.Context, query string) ([]domain.ProviderContent, error) {
	body, err := p.FetchData(ctx, "application/json")
	if err != nil {
		return nil, err
	}

	var apiResp Provider1Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	contents := make([]domain.ProviderContent, 0, len(apiResp.Contents))
	query = strings.ToLower(query)

	for _, item := range apiResp.Contents {
		if query != "" && !strings.Contains(strings.ToLower(item.Title), query) {
			continue
		}

		publishedAt, _ := time.Parse(time.RFC3339, item.PublishedAt)
		rawData, _ := json.Marshal(item)

		content := domain.ProviderContent{
			ExternalID:  item.ID,
			Title:       item.Title,
			Type:        mapContentType(item.Type),
			PublishedAt: publishedAt,
			Views:       item.Metrics.Views,
			Likes:       item.Metrics.Likes,
			Tags:        item.Tags,
			RawData:     rawData,
		}

		if err := domain.ValidateProviderContent(content); err != nil {
			if p.logger != nil {
				p.logger.Warn("invalid content from provider, skipping",
					zap.String("provider", p.Name()),
					zap.String("external_id", content.ExternalID),
					zap.Error(err),
				)
			}
			continue
		}

		contents = append(contents, content)
	}

	return contents, nil
}

func (p *Provider1) FetchAll(ctx context.Context) ([]domain.ProviderContent, error) {
	body, err := p.FetchData(ctx, "application/json")
	if err != nil {
		return nil, err
	}

	var apiResp Provider1Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	contents := make([]domain.ProviderContent, 0, len(apiResp.Contents))

	for _, item := range apiResp.Contents {
		publishedAt, _ := time.Parse(time.RFC3339, item.PublishedAt)
		rawData, _ := json.Marshal(item)

		content := domain.ProviderContent{
			ExternalID:  item.ID,
			Title:       item.Title,
			Type:        mapContentType(item.Type),
			PublishedAt: publishedAt,
			Views:       item.Metrics.Views,
			Likes:       item.Metrics.Likes,
			Tags:        item.Tags,
			RawData:     rawData,
		}

		if err := domain.ValidateProviderContent(content); err != nil {
			if p.logger != nil {
				p.logger.Warn("invalid content from provider, skipping",
					zap.String("provider", p.Name()),
					zap.String("external_id", content.ExternalID),
					zap.Error(err),
				)
			}
			continue
		}

		contents = append(contents, content)
	}

	return contents, nil
}

func (p *Provider1) SearchWithPagination(ctx context.Context, query string, page, perPage int) (*SearchResponse, error) {
	body, err := p.FetchData(ctx, "application/json")
	if err != nil {
		return nil, err
	}

	var apiResp Provider1Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	allContents := make([]domain.ProviderContent, 0)
	query = strings.ToLower(query)

	for _, item := range apiResp.Contents {
		if query != "" && !strings.Contains(strings.ToLower(item.Title), query) {
			continue
		}

		publishedAt, _ := time.Parse(time.RFC3339, item.PublishedAt)
		rawData, _ := json.Marshal(item)

		content := domain.ProviderContent{
			ExternalID:  item.ID,
			Title:       item.Title,
			Type:        mapContentType(item.Type),
			PublishedAt: publishedAt,
			Views:       item.Metrics.Views,
			Likes:       item.Metrics.Likes,
			Tags:        item.Tags,
			RawData:     rawData,
		}

		if err := domain.ValidateProviderContent(content); err != nil {
			if p.logger != nil {
				p.logger.Warn("invalid content from provider, skipping",
					zap.String("provider", p.Name()),
					zap.String("external_id", content.ExternalID),
					zap.Error(err),
				)
			}
			continue
		}

		allContents = append(allContents, content)
	}

	total := len(allContents)
	totalPages := (total + perPage - 1) / perPage

	start := (page - 1) * perPage
	end := start + perPage

	if start >= total {
		return &SearchResponse{
			Contents: []domain.ProviderContent{},
			Pagination: PaginationInfo{
				CurrentPage: page,
				PerPage:     perPage,
				Total:       total,
				TotalPages:  totalPages,
			},
		}, nil
	}

	if end > total {
		end = total
	}

	pagedContents := allContents[start:end]

	return &SearchResponse{
		Contents: pagedContents,
		Pagination: PaginationInfo{
			CurrentPage: page,
			PerPage:     perPage,
			Total:       total,
			TotalPages:  totalPages,
		},
	}, nil
}

