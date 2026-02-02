package provider

import (
	"context"
	"time"

	"search-engine/domain"
)

type ContentProvider interface {
	Search(ctx context.Context, query string) ([]domain.ProviderContent, error)
	Name() string
	HealthCheck(ctx context.Context) error
}

type FetchableProvider interface {
	ContentProvider
	FetchAll(ctx context.Context) ([]domain.ProviderContent, error)
}

type PaginatableProvider interface {
	ContentProvider
	SearchWithPagination(ctx context.Context, query string, page, perPage int) (*SearchResponse, error)
}

type SearchResponse struct {
	Contents   []domain.ProviderContent
	Pagination PaginationInfo
}

type PaginationInfo struct {
	CurrentPage int
	PerPage     int
	Total       int
	TotalPages  int
}

type ProviderResult struct {
	Provider string
	Contents []domain.ProviderContent
	Error    error
	Duration time.Duration
}

type Manager struct {
	providers []ContentProvider
	timeout   time.Duration
}

func NewManager(timeout time.Duration) *Manager {
	return &Manager{
		providers: make([]ContentProvider, 0),
		timeout:   timeout,
	}
}

func (m *Manager) Register(provider ContentProvider) {
	m.providers = append(m.providers, provider)
}

func (m *Manager) SearchAll(ctx context.Context, query string) []ProviderResult {
	results := make(chan ProviderResult, len(m.providers))

	for _, p := range m.providers {
		go func(provider ContentProvider) {
			start := time.Now()

			providerCtx, cancel := context.WithTimeout(ctx, m.timeout)
			defer cancel()

			contents, err := provider.Search(providerCtx, query)
			results <- ProviderResult{
				Provider: provider.Name(),
				Contents: contents,
				Error:    err,
				Duration: time.Since(start),
			}
		}(p)
	}

	collected := make([]ProviderResult, 0, len(m.providers))
	for i := 0; i < len(m.providers); i++ {
		select {
		case result := <-results:
			collected = append(collected, result)
		case <-ctx.Done():
			return collected
		}
	}

	return collected
}

func (m *Manager) SearchAllWithPagination(ctx context.Context, query string, page, perPage int) []ProviderResult {
	results := make(chan ProviderResult, len(m.providers))

	for _, p := range m.providers {
		go func(provider ContentProvider) {
			start := time.Now()

			providerCtx, cancel := context.WithTimeout(ctx, m.timeout)
			defer cancel()

			var contents []domain.ProviderContent
			var err error

			if paginatable, ok := provider.(PaginatableProvider); ok {
				resp, paginationErr := paginatable.SearchWithPagination(providerCtx, query, page, perPage)
				if paginationErr != nil {
					err = paginationErr
				} else {
					contents = resp.Contents
				}
			} else {
				contents, err = provider.Search(providerCtx, query)
			}

			results <- ProviderResult{
				Provider: provider.Name(),
				Contents: contents,
				Error:    err,
				Duration: time.Since(start),
			}
		}(p)
	}

	collected := make([]ProviderResult, 0, len(m.providers))
	for i := 0; i < len(m.providers); i++ {
		select {
		case result := <-results:
			collected = append(collected, result)
		case <-ctx.Done():
			return collected
		}
	}

	return collected
}

func (m *Manager) HealthCheckAll(ctx context.Context) map[string]error {
	healthResults := make(map[string]error)

	for _, p := range m.providers {
		healthResults[p.Name()] = p.HealthCheck(ctx)
	}

	return healthResults
}

func (m *Manager) GetProviders() []ContentProvider {
	return m.providers
}
