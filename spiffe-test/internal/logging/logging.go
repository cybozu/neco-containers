package logging

import (
	"log/slog"
	"os"
)

// Setup configures the default slog logger with JSON output.
// If debug is true, log level is set to Debug; otherwise Info.
func Setup(debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
}
