package libs

import (
	"os"
	"strings"

	"golang.org/x/exp/slog"
)

func InitLogging(level string, logFormat string) {
	var logLevel slog.Leveler

	switch strings.ToUpper(level) {
	case "INFO":
		logLevel = slog.LevelInfo
	case "ERROR":
		logLevel = slog.LevelError
	case "WARN":
		logLevel = slog.LevelWarn
	case "DEBUG":
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}

	opts := slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	if strings.ToLower(logFormat) == "json" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &opts)))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &opts)))
	}
}
