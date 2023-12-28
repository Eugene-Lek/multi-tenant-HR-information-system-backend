package routes

import (
	"io"
	"log"
	"log/slog"
	"os"
)

// Embed the slog.Logger struct to inherit all its methods & fields
type tailoredLogger struct {
	*slog.Logger
}

// Overwrite the below methods in order to implement my preferred versions
func (l tailoredLogger) With(args ...any) *tailoredLogger {
	return &tailoredLogger{l.Logger.With(args...)}
}

func (l tailoredLogger) Info(msg string, args ...any) {
	l.Logger.Info(msg, slog.Group("details", args...))
}

func (l tailoredLogger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, slog.Group("details", args...))
}

func (l tailoredLogger) Error(msg string, args ...any) {
	l.Logger.Error(msg, slog.Group("details", args...))
}

func (l tailoredLogger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, slog.Group("details", args...))
}

const serviceName = "hr-information-system-backend"

// Returns a root logger with the chosen output medium
func NewRootLogger(w io.Writer) *tailoredLogger {
	// Instantiate logger with JSON format & output medium
	rootLogger := slog.New(slog.NewJSONHandler(w, nil))

	// Add metadata that apply to all requests
	hostName, err := os.Hostname()
	if err != nil {
		log.Fatalf("Logger could not retrieve hostname: %s", err)
	}

	rootLogger = rootLogger.With("name", serviceName, "hostname", hostName)

	return &tailoredLogger{rootLogger}
}
