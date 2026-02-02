package search

import (
	"context"
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"time"

	"search-engine/domain"
	"search-engine/infra/provider"
	"search-engine/infra/redis"
	"search-engine/domain/scoring"

	"go.uber.org/zap"
)

type Service struct {
	repo            Repository
	providerManager *provider.Manager
	cache           *redis.RedisCache
	logger          *zap.Logger
	scorer          *scoring.Scorer
	cacheTTL        time.Duration
}

func NewService(repo Repository, pm *provider.Manager, cache *redis.RedisCache, logger *zap.Logger) *Service {
	return &Service{
		repo:            repo,
		providerManager: pm,
		cache:           cache,
		logger:          logger,
		scorer:          scoring.NewScorer(),
		cacheTTL:        5 * time.Minute,
	}
}

type SearchParams struct {
	Query        string
	Tags         []string
	ContentTypes []string
	SortBy       string
	Page         int
	PerPage      int
}

type SearchResult struct {
	Items      []domain.Content
	Total      int64
	Page       int
	PerPage    int
	TotalPages int
}

func (s *Service) Search(ctx context.Context, params SearchParams) (*SearchResult, error) {
	cacheKey := s.generateCacheKey(params)

	var cachedResult SearchResult
	if s.cache != nil {
		if err := s.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
			s.logger.Debug("cache hit", zap.String("cache_key", cacheKey))
			return &cachedResult, nil
		}
	}

	s.logger.Info("fetching from providers with pagination",
		zap.String("query", params.Query),
		zap.Int("page", params.Page),
		zap.Int("per_page", params.PerPage),
	)

	providerResults := s.providerManager.SearchAllWithPagination(ctx, params.Query, params.Page, params.PerPage)

	var allContents []domain.Content

	for _, result := range providerResults {
		if result.Error != nil {
			s.logger.Warn("provider failed, falling back to database",
				zap.String("provider", result.Provider),
				zap.Error(result.Error),
			)

			dbContents, err := s.repo.SearchByProvider(ctx, result.Provider, params.Query, params.Page, params.PerPage)
			if err != nil {
				s.logger.Error("database fallback also failed",
					zap.String("provider", result.Provider),
					zap.Error(err),
				)
				continue
			}

			s.logger.Info("served from database (circuit breaker fallback)",
				zap.String("provider", result.Provider),
				zap.Int("count", len(dbContents)),
			)

			allContents = append(allContents, dbContents...)
		} else {
			for _, pc := range result.Contents {
				score := s.scorer.CalculateScore(pc)

				content := domain.Content{
					ID:          domain.NewUUID(),
					ExternalID:  pc.ExternalID,
					Provider:    result.Provider,
					Title:       pc.Title,
					Type:        domain.ContentType(pc.Type),
					PublishedAt: pc.PublishedAt,
					Views:       pc.Views,
					Likes:       pc.Likes,
					Reactions:   pc.Reactions,
					ReadingTime: pc.ReadingTime,
					Score:       score,
					Tags:        pc.Tags,
					RawData:     pc.RawData,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				allContents = append(allContents, content)
			}

			go s.persistContentsToDatabase(context.Background(), result.Contents, result.Provider)
		}
	}

	filteredContents := s.applyFiltersAndSorting(allContents, params)

	paginatedContents, total := s.paginateResults(filteredContents, params.Page, params.PerPage)

	totalPages := int(total) / params.PerPage
	if int(total)%params.PerPage != 0 {
		totalPages++
	}

	result := &SearchResult{
		Items:      paginatedContents,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
	}

	if s.cache != nil {
		if err := s.cache.Set(ctx, cacheKey, result, s.cacheTTL); err != nil {
			s.logger.Warn("failed to cache result",
				zap.Error(err),
				zap.String("cache_key", cacheKey),
			)
		}
	}

	return result, nil
}

func (s *Service) generateCacheKey(params SearchParams) string {
	keyData := fmt.Sprintf("q=%s&tags=%v&types=%v&sort=%s&page=%d&per_page=%d",
		params.Query,
		params.Tags,
		params.ContentTypes,
		params.SortBy,
		params.Page,
		params.PerPage,
	)

	hash := md5.Sum([]byte(keyData))
	return fmt.Sprintf("search:%x", hash)
}

func (s *Service) mergeAndScoreResults(providerResults []provider.ProviderResult) []domain.Content {
	var allContents []domain.Content

	for _, result := range providerResults {
		if result.Error != nil {
			s.logger.Warn("provider search failed",
				zap.String("provider", result.Provider),
				zap.Error(result.Error),
			)
			continue
		}

		for _, providerContent := range result.Contents {
			content := domain.Content{
				ID:          domain.NewUUID(),
				ExternalID:  providerContent.ExternalID,
				Provider:    result.Provider,
				Title:       providerContent.Title,
				Type:        domain.ContentType(providerContent.Type),
				PublishedAt: providerContent.PublishedAt,
				Views:       providerContent.Views,
				Likes:       providerContent.Likes,
				Reactions:   providerContent.Reactions,
				ReadingTime: providerContent.ReadingTime,
				Tags:        providerContent.Tags,
				RawData:     providerContent.RawData,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			content.Score = s.scorer.CalculateScore(providerContent)
			allContents = append(allContents, content)
		}
	}

	return allContents
}

func (s *Service) applyFiltersAndSorting(contents []domain.Content, params SearchParams) []domain.Content {
	if len(params.ContentTypes) > 0 {
		typeMap := make(map[string]bool)
		for _, ct := range params.ContentTypes {
			typeMap[ct] = true
		}

		var filtered []domain.Content
		for _, content := range contents {
			if typeMap[string(content.Type)] {
				filtered = append(filtered, content)
			}
		}
		contents = filtered
	}

	if len(params.Tags) > 0 {
		tagMap := make(map[string]bool)
		for _, tag := range params.Tags {
			tagMap[strings.ToLower(tag)] = true
		}

		var filtered []domain.Content
		for _, content := range contents {
			hasMatchingTag := false
			for _, contentTag := range content.Tags {
				if tagMap[strings.ToLower(contentTag)] {
					hasMatchingTag = true
					break
				}
			}
			if hasMatchingTag {
				filtered = append(filtered, content)
			}
		}
		contents = filtered
	}

	sort.Slice(contents, func(i, j int) bool {
		switch params.SortBy {
		case "popularity":
			if contents[i].Type == "video" && contents[j].Type == "video" {
				return contents[i].Views > contents[j].Views
			}
			if contents[i].Type == "text" && contents[j].Type == "text" {
				return contents[i].Reactions > contents[j].Reactions
			}
			return contents[i].Score > contents[j].Score
		default:
			return contents[i].Score > contents[j].Score
		}
	})

	return contents
}

func (s *Service) paginateResults(contents []domain.Content, page, perPage int) ([]domain.Content, int64) {
	total := int64(len(contents))

	if total == 0 {
		return []domain.Content{}, 0
	}

	start := (page - 1) * perPage
	end := start + perPage

	if start >= len(contents) {
		return []domain.Content{}, total
	}

	if end > len(contents) {
		end = len(contents)
	}

	return contents[start:end], total
}

func (s *Service) persistContentsToDatabase(ctx context.Context, providerContents []domain.ProviderContent, providerName string) {
	s.logger.Debug("persisting contents to database",
		zap.String("provider", providerName),
		zap.Int("count", len(providerContents)),
	)

	successCount := 0
	for _, pc := range providerContents {
		score := s.scorer.CalculateScore(pc)

		content := &domain.Content{
			ID:          domain.NewUUID(),
			ExternalID:  pc.ExternalID,
			Provider:    providerName,
			Title:       pc.Title,
			Type:        domain.ContentType(pc.Type),
			PublishedAt: pc.PublishedAt,
			Views:       pc.Views,
			Likes:       pc.Likes,
			Reactions:   pc.Reactions,
			ReadingTime: pc.ReadingTime,
			Score:       score,
			Tags:        pc.Tags,
			RawData:     pc.RawData,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.repo.Upsert(ctx, content); err != nil {
			s.logger.Error("failed to upsert content",
				zap.Error(err),
				zap.String("provider", providerName),
				zap.String("external_id", pc.ExternalID),
			)
		} else {
			successCount++
		}
	}

	s.logger.Info("contents persisted to database",
		zap.String("provider", providerName),
		zap.Int("total", len(providerContents)),
		zap.Int("success", successCount),
	)
}
