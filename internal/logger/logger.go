// Package logger provides structured logging for VoidVPN using log/slog.
package logger

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
)

var paused atomic.Bool

// Init configures the global slog logger based on the log level string.
// Valid levels: "debug", "info", "warn", "error". Defaults to "info".
func Init(level string) {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	inner := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slogLevel,
	})
	slog.SetDefault(slog.New(&pauseHandler{inner: inner}))
}

// Pause suppresses all log output. Use during TUI rendering to prevent
// log lines from interleaving with the spinner animation.
func Pause() {
	paused.Store(true)
}

// Resume re-enables log output after a Pause.
func Resume() {
	paused.Store(false)
}

// pauseHandler wraps an slog.Handler and drops records while paused.
type pauseHandler struct {
	inner slog.Handler
}

func (h *pauseHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if paused.Load() {
		return false
	}
	return h.inner.Enabled(ctx, level)
}

func (h *pauseHandler) Handle(ctx context.Context, r slog.Record) error {
	if paused.Load() {
		return nil
	}
	return h.inner.Handle(ctx, r)
}

func (h *pauseHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &pauseHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *pauseHandler) WithGroup(name string) slog.Handler {
	return &pauseHandler{inner: h.inner.WithGroup(name)}
}
