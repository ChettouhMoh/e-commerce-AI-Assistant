package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-ai-assistant/internal/app"
	"ecommerce-ai-assistant/internal/platform/config"
	"ecommerce-ai-assistant/internal/platform/logging"

	"github.com/gin-gonic/gin"
)

func main() {
	ingest := flag.Bool("ingest", false, "Run RAG ingestion and exit")
	flag.Parse()

	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	logger := logging.New(cfg.AppEnv)

	if *ingest {
		if err := runIngestion(context.Background(), cfg, logger); err != nil {
			logger.Error("ingestion failed", "error", err)
			os.Exit(1)
		}
		logger.Info("ingestion complete")
		return
	}

	deps, err := app.Wire(context.Background(), cfg)
	if err != nil {
		logger.Error("wiring failed", "error", err)
		os.Exit(1)
	}

	handler, err := buildHandler(deps, cfg, logger)
	if err != nil {
		logger.Error("handler build failed", "error", err)
		os.Exit(1)
	}

	srv := app.NewServer(cfg.HTTPAddr, handler, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	logger.Info("server stopped")
}

func runIngestion(ctx context.Context, cfg *config.Config, logger *logging.Logger) error {
	return nil
}

func buildHandler(deps *app.Dependencies, cfg *config.Config, logger *logging.Logger) (http.Handler, error) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			id = generateRequestID()
		}
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-Id", id)
		c.Next()
	})
	r.GET("/healthz", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/readyz", func(c *gin.Context) { c.String(http.StatusOK, "ready") })
	return r, nil
}

func generateRequestID() string {
	return "req-" + time.Now().Format("20060102150405")
}
