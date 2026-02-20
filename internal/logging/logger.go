package logging

import (
	"fmt"
	"gate-limiter/config/settings"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Setup configures the global slog logger and returns a closer function.
func Setup(cfg settings.LoggingConfig) (func() error, error) {
	level := parseLevel(cfg.Level)

	var w io.Writer
	var closer func() error

	switch cfg.Output {
	case "stderr":
		w = os.Stderr
		closer = func() error { return nil }
	case "file":
		rw, err := NewRotatingWriter(cfg.File)
		if err != nil {
			return nil, fmt.Errorf("rotating writer: %w", err)
		}
		w = rw
		closer = rw.Close
	default: // "stdout"
		w = os.Stdout
		closer = func() error { return nil }
	}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	default: // "text"
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	}

	slog.SetDefault(slog.New(handler))
	return closer, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
