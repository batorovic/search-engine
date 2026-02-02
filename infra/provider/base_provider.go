package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"search-engine/infra/httpclient"

	"go.uber.org/zap"
)

type BaseHTTPProvider struct {
	name    string
	baseURL string
	doer    *httpclient.Doer
}

type BaseHTTPProviderConfig struct {
	Name    string
	BaseURL string
	Timeout time.Duration
	Client  httpclient.HTTPClient
	Logger  *zap.Logger
}

func NewBaseHTTPProvider(config BaseHTTPProviderConfig) BaseHTTPProvider {
	client := config.Client
	if client == nil {
		client = httpclient.NewDefaultHTTPClient(
			httpclient.WithTimeout(config.Timeout),
		)
	}

	return BaseHTTPProvider{
		name:    config.Name,
		baseURL: config.BaseURL,
		doer:    httpclient.NewDoer(client),
	}
}

func (b *BaseHTTPProvider) Name() string {
	return b.name
}

func (b *BaseHTTPProvider) BaseURL() string {
	return b.baseURL
}

func (b *BaseHTTPProvider) HealthCheck(ctx context.Context) error {
	resp, err := b.doer.Head(ctx, b.baseURL)
	if err != nil {
		return fmt.Errorf("health check request failed for %s: %w", b.name, err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("health check for %s returned status: %d", b.name, resp.StatusCode)
	}

	return nil
}

func (b *BaseHTTPProvider) FetchData(ctx context.Context, acceptHeader string) ([]byte, error) {
	headers := make(map[string]string)
	if acceptHeader != "" {
		headers["Accept"] = acceptHeader
	}

	resp, err := b.doer.Get(ctx, b.baseURL, headers)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", b.name, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s unexpected status code: %d", b.name, resp.StatusCode)
	}

	return resp.Body, nil
}
