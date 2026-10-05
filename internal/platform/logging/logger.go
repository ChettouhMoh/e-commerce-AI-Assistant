package logging

import (
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(env string) *Logger {
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	return &Logger{Logger: slog.New(handler)}
}
