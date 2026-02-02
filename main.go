package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"search-engine/app/health"
	"search-engine/app/search"
	"search-engine/infra/httpclient"
	"search-engine/infra/postgres"
	"search-engine/infra/provider"
	"search-engine/infra/redis"
	"search-engine/pkg/config"
	"search-engine/pkg/log"
	"search-engine/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"go.uber.org/zap"

	_ "search-engine/docs"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logger := log.NewLogger(cfg.App.Env)
	defer logger.Sync()

	logger.Info("starting application",
		zap.String("app", cfg.App.Name),
		zap.String("env", cfg.App.Env),
		zap.String("port", cfg.Server.Port),
	)

	logger.Info("available provider formats",
		zap.Strings("formats", provider.ListRegisteredFormats()),
	)

	db, err := postgres.NewPostgresDB(cfg.Database.ConnectionString())
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	logger.Info("database connected")

	redisCache, err := redis.NewRedisCache(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer redisCache.Close()
	logger.Info("redis connected")

	providerManager := provider.NewManager(cfg.Provider.Timeout)

	httpClient := httpclient.NewDefaultHTTPClient(
		httpclient.WithTimeout(cfg.Provider.Timeout),
	)

	for _, p := range cfg.Providers {
		providerFormat := "http_" + p.Format

		contentProvider, err := provider.CreateProvider(providerFormat, p.Name, p.URL, cfg.Provider.Timeout, httpClient, logger)
		if err != nil {
			logger.Warn("failed to create provider",
				zap.String("name", p.Name),
				zap.String("format", providerFormat),
				zap.Error(err),
			)
			continue
		}

		wrappedProvider := provider.NewCircuitBreakerProvider(
			contentProvider,
			cfg.Provider.CircuitBreakerThreshold,
			cfg.Provider.CircuitBreakerTimeout,
		)

		providerManager.Register(wrappedProvider)
		logger.Info("registered provider",
			zap.String("name", p.Name),
			zap.String("format", providerFormat),
			zap.String("url", p.URL),
		)
	}

	logger.Info("providers registered", zap.Int("count", len(cfg.Providers)))

	searchRepo := postgres.NewRepository(db)

	searchService := search.NewService(searchRepo, providerManager, redisCache, logger)

	healthHandler := health.NewHandler(db, redisCache, providerManager)
	searchHandler := search.NewHandler(searchService, logger)

	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: middleware.NewErrorHandler(logger),
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(middleware.NewLoggerMiddleware(logger))

	app.Use(limiter.New(limiter.Config{
		Max:        cfg.Provider.RateLimitMax,
		Expiration: cfg.Provider.RateLimitWindow,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Rate limit exceeded. Please try again later.",
				},
				"meta": fiber.Map{
					"request_id": c.Locals("requestid"),
				},
			})
		},
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	healthHandler.RegisterRoutes(app)
	searchHandler.RegisterRoutes(app)

	go func() {
		if err := app.Listen(":" + cfg.Server.Port); err != nil {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}
