package search

import (
	"context"

	"search-engine/domain"
)

type Repository interface {
	Search(ctx context.Context, query string, tags []string, contentTypes []string, sortBy string, page, perPage int) ([]domain.Content, int64, error)
	SearchByProvider(ctx context.Context, provider string, query string, page, perPage int) ([]domain.Content, error)
	Upsert(ctx context.Context, content *domain.Content) error
	GetByID(ctx context.Context, id string) (*domain.Content, error)
}
