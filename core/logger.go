package core

import (
	"io"
	"log/slog"
	"os"
)

// InitLogger initializes the default slog logger with configuration
func InitLogger(config LoggingConfig) {
	// Set defaults
	if config.Level == "" {
		config.Level = "info"
	}
	if config.Format == "" {
		config.Format = "text"
	}
	if config.Output == "" {
		config.Output = "stdout"
	}

	// Parse log level
	var logLevel slog.Level
	switch config.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Determine output destination
	var output io.Writer
	switch config.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// Fallback to stdout if file can't be opened
			output = os.Stdout
		} else {
			output = file
		}
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	switch config.Format {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	// Set the default logger
	slog.SetDefault(slog.New(handler))
}
