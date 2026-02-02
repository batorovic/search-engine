package health

import (
	"context"
	"time"

	"search-engine/infra/postgres"
	"search-engine/infra/provider"
	"search-engine/infra/redis"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	db              *postgres.PostgresDB
	cache           *redis.RedisCache
	providerManager *provider.Manager
}

func NewHandler(db *postgres.PostgresDB, cache *redis.RedisCache, pm *provider.Manager) *Handler {
	return &Handler{
		db:              db,
		cache:           cache,
		providerManager: pm,
	}
}

type Response struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (h *Handler) Liveness(c *fiber.Ctx) error {
	return c.JSON(Response{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) Readiness(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	overallStatus := "healthy"

	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = CheckResult{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			overallStatus = "unhealthy"
		} else {
			checks["database"] = CheckResult{Status: "healthy"}
		}
	}

	if h.cache != nil {
		if err := h.cache.Ping(ctx); err != nil {
			checks["redis"] = CheckResult{
				Status:  "unhealthy",
				Message: err.Error(),
			}
			overallStatus = "unhealthy"
		} else {
			checks["redis"] = CheckResult{Status: "healthy"}
		}
	}

	if h.providerManager != nil {
		providerHealth := h.providerManager.HealthCheckAll(ctx)
		for name, err := range providerHealth {
			if err != nil {
				checks["provider_"+name] = CheckResult{
					Status:  "unhealthy",
					Message: err.Error(),
				}
			} else {
				checks["provider_"+name] = CheckResult{Status: "healthy"}
			}
		}
	}

	response := Response{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	statusCode := fiber.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = fiber.StatusServiceUnavailable
	}

	return c.Status(statusCode).JSON(response)
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.Liveness)
	app.Get("/health/ready", h.Readiness)
}
