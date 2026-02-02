package search

import (
	"strconv"
	"strings"

	"search-engine/domain"
	"search-engine/pkg/apierror"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	logger  *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Search(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	var req domain.SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, apierror.NewValidationError(err.Error()), requestID)
	}

	req.SetDefaults()

	if req.OrderBy != "popularity" && req.OrderBy != "relevant_score" {
		return h.errorResponse(c, apierror.ErrInvalidSortField, requestID)
	}

	params := SearchParams{
		Query:        req.Query,
		Tags:         req.Tags,
		ContentTypes: req.ContentTypes,
		SortBy:       req.OrderBy,
		Page:         req.Page,
		PerPage:      req.PerPage,
	}

	result, err := h.service.Search(c.Context(), params)
	if err != nil {
		h.logger.Error("search failed",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
		return h.errorResponse(c, apierror.ErrInternalServer, requestID)
	}

	response := domain.NewSuccessResponse(
		domain.SearchData{Items: result.Items},
		&domain.Meta{
			Page:       result.Page,
			PerPage:    result.PerPage,
			Total:      result.Total,
			TotalPages: result.TotalPages,
			RequestID:  requestID,
		},
	)

	return c.JSON(response)
}

func (h *Handler) SearchGET(c *fiber.Ctx) error {
	requestID := c.Locals("requestid").(string)

	query := c.Query("q", "")
	tagsParam := c.Query("tags", "")
	contentType := c.Query("type", "")
	sortBy := c.Query("sort", "relevant_score")
	pageStr := c.Query("page", "1")
	perPageStr := c.Query("per_page", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil || perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	var tags []string
	if tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	var contentTypes []string
	if contentType != "" {
		contentTypes = strings.Split(contentType, ",")
		for i := range contentTypes {
			contentTypes[i] = strings.TrimSpace(contentTypes[i])
		}
	}

	if sortBy != "popularity" && sortBy != "relevant_score" {
		return h.errorResponse(c, apierror.ErrInvalidSortField, requestID)
	}

	params := SearchParams{
		Query:        query,
		Tags:         tags,
		ContentTypes: contentTypes,
		SortBy:       sortBy,
		Page:         page,
		PerPage:      perPage,
	}

	result, err := h.service.Search(c.Context(), params)
	if err != nil {
		h.logger.Error("search failed",
			zap.Error(err),
			zap.String("request_id", requestID),
		)
		return h.errorResponse(c, apierror.ErrInternalServer, requestID)
	}

	response := domain.NewSuccessResponse(
		domain.SearchData{Items: result.Items},
		&domain.Meta{
			Page:       result.Page,
			PerPage:    result.PerPage,
			Total:      result.Total,
			TotalPages: result.TotalPages,
			RequestID:  requestID,
		},
	)

	return c.JSON(response)
}

func (h *Handler) errorResponse(c *fiber.Ctx, apiErr *apierror.APIError, requestID string) error {
	response := domain.NewErrorResponse(apiErr.Code, apiErr.Message, requestID)
	return c.Status(apiErr.StatusCode).JSON(response)
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	v1 := app.Group("/api/v1")
	v1.Post("/search", h.Search)
}
