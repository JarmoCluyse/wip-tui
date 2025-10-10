package logging

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func Init() {
	// Create a file for logging
	logFile, err := os.OpenFile("wip-tui.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Fallback to stderr if file creation fails
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		logger.Error("failed to create log file, using stderr", "error", err)
		return
	}

	// Create structured logger with debug level
	logger = slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Log that logger was initialized
	logger.Info("logger initialized successfully", "log_file", "wip-tui.log")
}

func Get() *slog.Logger {
	if logger == nil {
		Init()
	}
	return logger
}
