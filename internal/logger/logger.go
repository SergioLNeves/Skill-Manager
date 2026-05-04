package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Init configures the default slog logger to write JSON to logDir/app.log
// and also to stderr. Returns a cleanup function that flushes and closes the file.
func Init(logDir string) (func(), error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return func() {}, err
	}

	logPath := filepath.Join(logDir, "app.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return func() {}, err
	}

	w := io.MultiWriter(f, os.Stderr)
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(h))

	return func() { f.Close() }, nil
}
