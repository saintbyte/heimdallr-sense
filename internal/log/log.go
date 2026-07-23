package log

import (
	"log/slog"
	"os"
)

var logger *slog.Logger
var enabled bool

func Init(e bool) {
	enabled = e
	if enabled {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
}

func Info(msg string, args ...any) {
	if enabled {
		logger.Info(msg, args...)
	}
}

func Error(msg string, args ...any) {
	if enabled {
		logger.Error(msg, args...)
	}
}

func Fatal(msg string, args ...any) {
	if enabled {
		logger.Error(msg, args...)
	}
	os.Exit(1)
}
