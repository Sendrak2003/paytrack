package main

import (
	_ "embed"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"payments-dashboard/internal/ai"
	"payments-dashboard/internal/config"
	"payments-dashboard/internal/handlers"
	"payments-dashboard/internal/logging"
	"payments-dashboard/internal/models"
	"payments-dashboard/internal/repository"
	"payments-dashboard/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed docs/swagger.html
var swaggerHTML []byte

//go:embed docs/openapi.yaml
var openapiSpec []byte

func main() {
	cfg := config.Load()
	log := logging.New(cfg.LogLevel, cfg.LogFormat)
	slog.SetDefault(log)

	log.Info("starting paytrack",
		"port", cfg.Port,
		"db_driver", cfg.DBDriver,
		"allowed_origins", cfg.AllowedOrigins,
	)

	db, err := connectDB(cfg, log)
	if err != nil {
		log.Error("database connection failed", "error", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(
		&models.Client{},
		&models.Project{},
		&models.Payment{},
		&models.Act{},
	); err != nil {
		log.Error("migration failed", "error", err)
		os.Exit(1)
	}
	log.Info("migrations applied")

	seedDatabase(db)

	repo := repository.New(db)
	extractor := ai.NewExtractor(ai.Config{
		Provider: cfg.AIProvider,
		BaseURL:  cfg.AIBaseURL,
		APIKey:   cfg.AIAPIKey,
		Model:    cfg.AIModel,
		Log:      log,
	})
	importSvc := services.NewImportService(db, log, extractor)
	h := handlers.New(repo, importSvc, log)

	if cfg.LogLevel != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db handle error", "error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			log.Warn("readiness check failed", "error", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db unreachable", "error": err.Error()})
			return
		}
		stats := sqlDB.Stats()
		c.JSON(http.StatusOK, gin.H{
			"status":        "ready",
			"db_open_conns": stats.OpenConnections,
			"db_in_use":     stats.InUse,
			"db_idle":       stats.Idle,
		})
	})

	h.RegisterRoutes(r)

	r.GET("/swagger", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", swaggerHTML)
	})
	r.GET("/swagger/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", openapiSpec)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
		// Timeouts protect against slow/stuck clients (slowloris) and leaked
		// connections. Tunable via env (see config).
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	go func() {
		log.Info("server listening", "addr", srv.Addr, "swagger", "/swagger")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Info("server stopped")
}

// connectDB opens the DB, configures the connection pool, and verifies
// connectivity with a retry loop (the DB container may start after the app).
func connectDB(cfg config.Config, log *slog.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}

	dialector, err := openDialector(cfg.DBDriver, cfg.DBDSN)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	log.Info("connection pool configured",
		"max_open", cfg.MaxOpenConns,
		"max_idle", cfg.MaxIdleConns,
		"conn_max_lifetime", cfg.ConnMaxLifetime.String(),
	)

	if cfg.DBDriver == "sqlite" {
		db.Exec("PRAGMA journal_mode=WAL;")
		db.Exec("PRAGMA busy_timeout=5000;")
	}

	const maxAttempts = 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = sqlDB.PingContext(ctx)
		cancel()
		if err == nil {
			log.Info("database connected", "attempt", attempt)
			return db, nil
		}
		log.Warn("database ping failed, retrying",
			"attempt", attempt, "max", maxAttempts, "error", err)
		time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
	}
	return nil, err
}

func requestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()

		level := slog.LevelInfo
		status := c.Writer.Status()
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}
		if path == "/health" || path == "/ready" {
			level = slog.LevelDebug
		}

		log.Log(c.Request.Context(), level, "http_request",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
