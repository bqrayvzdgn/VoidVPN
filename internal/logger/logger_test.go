package logger

import (
	"context"
	"log/slog"
	"testing"
)

func TestInitDebug(t *testing.T) {
	Init("debug")
	h := slog.Default().Handler()
	if !h.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("debug level should be enabled after Init(\"debug\")")
	}
}

func TestInitInfo(t *testing.T) {
	Init("info")
	h := slog.Default().Handler()
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("info level should be enabled after Init(\"info\")")
	}
	if h.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("debug level should not be enabled after Init(\"info\")")
	}
}

func TestInitWarn(t *testing.T) {
	Init("warn")
	h := slog.Default().Handler()
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("warn level should be enabled after Init(\"warn\")")
	}
	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("info level should not be enabled after Init(\"warn\")")
	}
}

func TestInitError(t *testing.T) {
	Init("error")
	h := slog.Default().Handler()
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Error("error level should be enabled after Init(\"error\")")
	}
	if h.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("warn level should not be enabled after Init(\"error\")")
	}
}

func TestInitInvalidDefaultsToInfo(t *testing.T) {
	Init("garbage")
	h := slog.Default().Handler()
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("info level should be enabled after Init(\"garbage\") (default)")
	}
	if h.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("debug level should not be enabled for default level")
	}
}

func TestPauseSuppressesLogs(t *testing.T) {
	Init("debug")
	h := slog.Default().Handler()

	// Before pause, debug should be enabled
	if !h.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("debug should be enabled before Pause()")
	}

	Pause()
	defer Resume()

	// While paused, all levels should be suppressed
	if h.Enabled(context.Background(), slog.LevelError) {
		t.Error("error level should be suppressed while paused")
	}
	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("info level should be suppressed while paused")
	}
}

func TestResumeRestoresLogs(t *testing.T) {
	Init("info")
	Pause()
	Resume()

	h := slog.Default().Handler()
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("info should be enabled after Resume()")
	}
}
