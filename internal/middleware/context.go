package middleware

import (
	"context"
	"log/slog"
)

type contextKey int

const (
	loggerKey    contextKey = iota
	requestIDKey
)

// LoggerFrom returns the request-scoped logger, or slog.Default() if none.
func LoggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// RequestIDFrom returns the request ID from context, or empty string.
func RequestIDFrom(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
