package core

import (
	"io"
	"log/slog"
	"os"
)

// InitLogger initializes the default slog logger with configuration
func InitLogger(config *FrameworkConfig) {
	// Set defaults
	logLevel := config.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}

	logFormat := config.LogFormat
	if logFormat == "" {
		logFormat = "text"
	}

	logOutput := config.LogOutput
	if logOutput == "" {
		logOutput = "stdout"
	}

	// Parse log level
	var slogLevel slog.Level
	switch logLevel {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Determine output destination
	var output io.Writer
	switch logOutput {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// Assume it's a file path
		file, err := os.OpenFile(logOutput, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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
		Level: slogLevel,
	}

	switch logFormat {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	// Set the default logger
	slog.SetDefault(slog.New(handler))
}
