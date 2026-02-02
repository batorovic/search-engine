package provider

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"search-engine/domain"
	"search-engine/infra/httpclient"

	"go.uber.org/zap"
)

type HTTPProvider struct {
	name    string
	baseURL string
	format  string
	client  *httpclient.Doer
	logger  *zap.Logger
	timeout time.Duration
}

func NewHTTPProvider(name, baseURL, format string, timeout time.Duration, client httpclient.HTTPClient, logger *zap.Logger) (*HTTPProvider, error) {
	if format != "json" && format != "xml" {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	doer := httpclient.NewDoer(client)

	return &HTTPProvider{
		name:    name,
		baseURL: strings.TrimSuffix(baseURL, "/"),
		format:  format,
		client:  doer,
		logger:  logger,
		timeout: timeout,
	}, nil
}

func (p *HTTPProvider) Name() string {
	return p.name
}

func (p *HTTPProvider) Search(ctx context.Context, query string) ([]domain.ProviderContent, error) {
	searchURL := fmt.Sprintf("%s?q=%s", p.baseURL, url.QueryEscape(query))

	p.logger.Debug("searching provider",
		zap.String("provider", p.name),
		zap.String("url", searchURL),
		zap.String("query", query),
	)

	resp, err := p.client.Get(ctx, searchURL, map[string]string{
		"Accept":     p.getAcceptHeader(),
		"User-Agent": "search-engine/1.0",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from provider %s: %w", p.name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider %s returned status %d", p.name, resp.StatusCode)
	}

	switch p.format {
	case "json":
		return p.parseJSONResponse(resp.Body)
	case "xml":
		return p.parseXMLResponse(resp.Body)
	default:
		return nil, fmt.Errorf("unsupported format: %s", p.format)
	}
}

func (p *HTTPProvider) HealthCheck(ctx context.Context) error {
	resp, err := p.client.Head(ctx, p.baseURL)
	if err != nil {
		return fmt.Errorf("provider %s health check failed: %w", p.name, err)
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("provider %s is unhealthy: status %d", p.name, resp.StatusCode)
	}

	return nil
}

func (p *HTTPProvider) getAcceptHeader() string {
	switch p.format {
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	default:
		return "application/json"
	}
}

type JSONResponse struct {
	Contents []JSONContent `json:"contents"`
}

type JSONContent struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Metrics     JSONMetrics `json:"metrics"`
	PublishedAt string      `json:"published_at"`
	Tags        []string    `json:"tags"`
}

type JSONMetrics struct {
	Views    int    `json:"views"`
	Likes    int    `json:"likes"`
	Duration string `json:"duration,omitempty"`
}

func (p *HTTPProvider) parseJSONResponse(body []byte) ([]domain.ProviderContent, error) {
	var response JSONResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	contents := make([]domain.ProviderContent, len(response.Contents))
	for i, item := range response.Contents {
		publishedAt, err := time.Parse(time.RFC3339, item.PublishedAt)
		if err != nil {
			publishedAt = time.Now()
		}

		rawData, _ := json.Marshal(item)

		contents[i] = domain.ProviderContent{
			ExternalID:  item.ID,
			Title:       item.Title,
			Type:        mapContentType(item.Type),
			PublishedAt: publishedAt,
			Views:       item.Metrics.Views,
			Likes:       item.Metrics.Likes,
			Tags:        item.Tags,
			RawData:     rawData,
		}
	}

	return contents, nil
}

type XMLFeed struct {
	Items []XMLItem `xml:"items>item"`
}

type XMLItem struct {
	ID              string        `xml:"id"`
	Headline        string        `xml:"headline"`
	Type            string        `xml:"type"`
	Stats           XMLStats      `xml:"stats"`
	PublicationDate string        `xml:"publication_date"`
	Categories      XMLCategories `xml:"categories"`
}

type XMLStats struct {
	Views       int    `xml:"views"`
	Likes       int    `xml:"likes"`
	Duration    string `xml:"duration,omitempty"`
	ReadingTime int    `xml:"reading_time,omitempty"`
	Reactions   int    `xml:"reactions,omitempty"`
}

type XMLCategories struct {
	Categories []string `xml:"category"`
}

func (p *HTTPProvider) parseXMLResponse(body []byte) ([]domain.ProviderContent, error) {
	var feed XMLFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse XML response: %w", err)
	}

	contents := make([]domain.ProviderContent, len(feed.Items))
	for i, item := range feed.Items {
		publishedAt, err := time.Parse("2006-01-02", item.PublicationDate)
		if err != nil {
			publishedAt = time.Now()
		}

		rawData, _ := json.Marshal(item)

		contents[i] = domain.ProviderContent{
			ExternalID:  item.ID,
			Title:       item.Headline,
			Type:        mapContentType(item.Type),
			PublishedAt: publishedAt,
			Views:       item.Stats.Views,
			Likes:       item.Stats.Likes,
			ReadingTime: item.Stats.ReadingTime,
			Reactions:   item.Stats.Reactions,
			Tags:        item.Categories.Categories,
			RawData:     rawData,
		}
	}

	return contents, nil
}


func mapContentType(t string) string {
	t = strings.ToLower(t)
	switch t {
	case "video":
		return "video"
	case "article", "text":
		return "text"
	default:
		return t
	}
}
