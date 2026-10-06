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

	"github.com/joho/godotenv"
)

func main() {
	ingest := flag.Bool("ingest", false, "Run RAG ingestion and exit")
	flag.Parse()

	_ = godotenv.Load(".env")

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

	// Wire dependencies based on config
	deps, err := app.Wire(cfg)
	if err != nil {
		logger.Error("wiring failed", "error", err)
		os.Exit(1)
	}

	// Build HTTP handler
	handler := app.BuildHandler(deps)

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting server", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

func generateRequestID() string {
	return "req-" + time.Now().Format("20060102150405")
}
