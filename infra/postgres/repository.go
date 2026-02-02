package postgres

import (
	"context"
	"fmt"
	"math/big"

	"search-engine/app/search"
	"search-engine/domain"
	"search-engine/infra/postgres/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type repository struct {
	queries *db.Queries
}

func NewRepository(database *PostgresDB) search.Repository {
	return &repository{
		queries: db.New(database.Pool),
	}
}

func (r *repository) Search(ctx context.Context, query string, tags []string, contentTypes []string, sortBy string, page, perPage int) ([]domain.Content, int64, error) {
	params := db.SearchContentsParams{
		Query:        query,
		Tags:         tags,
		ContentTypes: contentTypes,
		SortBy:       sortBy,
		PageOffset:   int32((page - 1) * perPage),
		PageLimit:    int32(perPage),
	}

	rows, err := r.queries.SearchContents(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search contents: %w", err)
	}

	countParams := db.CountSearchContentsParams{
		Query:        query,
		Tags:         tags,
		ContentTypes: contentTypes,
	}
	total, err := r.queries.CountSearchContents(ctx, countParams)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count search contents: %w", err)
	}

	contents := make([]domain.Content, len(rows))
	for i, row := range rows {
		score, _ := row.Score.Float64Value()
		contents[i] = domain.Content{
			ID:          uuidFromPgtype(row.ID),
			ExternalID:  row.ExternalID,
			Provider:    row.Provider,
			Title:       row.Title,
			Type:        domain.ContentType(row.Type),
			PublishedAt: row.PublishedAt.Time,
			Views:       int(row.Views.Int32),
			Likes:       int(row.Likes.Int32),
			Reactions:   int(row.Reactions.Int32),
			ReadingTime: int(row.ReadingTime.Int32),
			Score:       score.Float64,
			Tags:        row.Tags,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
	}

	return contents, total, nil
}

func (r *repository) Upsert(ctx context.Context, content *domain.Content) error {
	params := db.UpsertContentParams{
		ExternalID:  content.ExternalID,
		Provider:    content.Provider,
		Title:       content.Title,
		Type:        string(content.Type),
		PublishedAt: pgtype.Timestamp{Time: content.PublishedAt, Valid: true},
		RawData:     content.RawData,
		Views:       pgtype.Int4{Int32: int32(content.Views), Valid: true},
		Likes:       pgtype.Int4{Int32: int32(content.Likes), Valid: true},
		Reactions:   pgtype.Int4{Int32: int32(content.Reactions), Valid: true},
		ReadingTime: pgtype.Int4{Int32: int32(content.ReadingTime), Valid: true},
		Score:       floatToNumeric(content.Score),
		Tags:        content.Tags,
	}

	_, err := r.queries.UpsertContent(ctx, params)
	return err
}

func (r *repository) SearchByProvider(ctx context.Context, provider string, query string, page, perPage int) ([]domain.Content, error) {
	params := db.SearchContentsByProviderParams{
		Provider:   provider,
		Query:      query,
		PageOffset: int32((page - 1) * perPage),
		PageLimit:  int32(perPage),
	}

	rows, err := r.queries.SearchContentsByProvider(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search contents by provider: %w", err)
	}

	contents := make([]domain.Content, len(rows))
	for i, row := range rows {
		score, _ := row.Score.Float64Value()
		contents[i] = domain.Content{
			ID:          uuidFromPgtype(row.ID),
			ExternalID:  row.ExternalID,
			Provider:    row.Provider,
			Title:       row.Title,
			Type:        domain.ContentType(row.Type),
			PublishedAt: row.PublishedAt.Time,
			Views:       int(row.Views.Int32),
			Likes:       int(row.Likes.Int32),
			Reactions:   int(row.Reactions.Int32),
			ReadingTime: int(row.ReadingTime.Int32),
			Score:       score.Float64,
			Tags:        row.Tags,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
	}

	return contents, nil
}

func (r *repository) GetByID(ctx context.Context, id string) (*domain.Content, error) {
	return nil, fmt.Errorf("not implemented")
}

func uuidFromPgtype(u pgtype.UUID) uuid.UUID {
	return uuid.UUID(u.Bytes)
}

func floatToNumeric(f float64) pgtype.Numeric {
	scaled := int64(f * 100)
	n := pgtype.Numeric{
		Int:   big.NewInt(scaled),
		Exp:   -2,
		Valid: true,
	}
	return n
}
