package provider

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"search-engine/domain"
	"search-engine/infra/httpclient"

	"go.uber.org/zap"
)

func init() {
	Register("xml", func(name, url string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) ContentProvider {
		return NewProvider2(name, url, timeout, client, logger)
	})
}

type Provider2 struct {
	BaseHTTPProvider
	logger *zap.Logger
}

type Provider2Feed struct {
	XMLName xml.Name       `xml:"feed"`
	Items   Provider2Items `xml:"items"`
	Meta    Provider2Meta  `xml:"meta"`
}

type Provider2Items struct {
	Items []Provider2Item `xml:"item"`
}

type Provider2Item struct {
	Type            string              `xml:"type,attr"`
	ID              string              `xml:"id"`
	Headline        string              `xml:"headline"`
	ContentType     string              `xml:"type"`
	Stats           Provider2Stats      `xml:"stats"`
	PublicationDate string              `xml:"publication_date"`
	Categories      Provider2Categories `xml:"categories"`
}

type Provider2Stats struct {
	Views    int    `xml:"views"`
	Likes    int    `xml:"likes"`
	Duration string `xml:"duration"`
	ReadingTime int `xml:"reading_time"`
	Reactions   int `xml:"reactions"`
	Comments    int `xml:"comments"`
}

type Provider2Categories struct {
	Categories []string `xml:"category"`
}

type Provider2Meta struct {
	TotalCount   int `xml:"total_count"`
	CurrentPage  int `xml:"current_page"`
	ItemsPerPage int `xml:"items_per_page"`
}

func NewProvider2(name, baseURL string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) *Provider2 {
	return &Provider2{
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

func (p *Provider2) Search(ctx context.Context, query string) ([]domain.ProviderContent, error) {
	body, err := p.FetchData(ctx, "application/xml")
	if err != nil {
		return nil, err
	}

	var feed Provider2Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	contents := make([]domain.ProviderContent, 0, len(feed.Items.Items))
	query = strings.ToLower(query)

	for _, item := range feed.Items.Items {
		if query != "" && !strings.Contains(strings.ToLower(item.Headline), query) {
			continue
		}

		content := p.mapToProviderContent(item)

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

func (p *Provider2) SearchWithPagination(ctx context.Context, query string, page, perPage int) (*SearchResponse, error) {
	body, err := p.FetchData(ctx, "application/xml")
	if err != nil {
		return nil, err
	}

	var feed Provider2Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	allContents := make([]domain.ProviderContent, 0)
	query = strings.ToLower(query)

	for _, item := range feed.Items.Items {
		if query != "" && !strings.Contains(strings.ToLower(item.Headline), query) {
			continue
		}

		content := p.mapToProviderContent(item)

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

func (p *Provider2) mapToProviderContent(item Provider2Item) domain.ProviderContent {
	publishedAt, _ := time.Parse("2006-01-02", item.PublicationDate)
	rawData, _ := json.Marshal(item)

	contentType := mapContentType(item.ContentType)

	content := domain.ProviderContent{
		ExternalID:  item.ID,
		Title:       item.Headline,
		Type:        contentType,
		PublishedAt: publishedAt,
		Tags:        item.Categories.Categories,
		RawData:     rawData,
	}

	if contentType == "video" {
		content.Views = item.Stats.Views
		content.Likes = item.Stats.Likes
	} else {
		content.ReadingTime = item.Stats.ReadingTime
		content.Reactions = item.Stats.Reactions
	}

	return content
}

