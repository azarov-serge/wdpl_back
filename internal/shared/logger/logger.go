package logger

import (
	"log/slog"
	"os"

	"wdpl_back/internal/shared/config"
)

// Логгер оборачиваем, чтобы при необходимости заменить реализацию.
type Logger = *slog.Logger

func New(cfg *config.Config) Logger {
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}
