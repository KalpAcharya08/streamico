package logger

// =============================================================================
// 📖 LEARNING NOTES — Structured Logger
// =============================================================================
// Your logic was correct! Here's what was added:
//   - "package logger" → every Go file needs this
//   - Import for "go.uber.org/zap"
//
// Zap is one of the fastest loggers in Go. It has two modes:
//   Development → human-readable, colorful output in your terminal
//   Production  → JSON format, perfect for log aggregation (ELK, Grafana)
// =============================================================================

import "go.uber.org/zap"

// New creates a Zap logger configured for the given environment.
//
//   - "development" → pretty-printed, human-readable logs
//   - "production"  → JSON-formatted, machine-readable logs
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		// JSON logs like: {"level":"info","ts":1234,"msg":"server started","port":"8080"}
		return zap.NewProduction()
	}
	// Pretty logs like: 2026-03-13T10:00:00  INFO  server started  {"port": "8080"}
	return zap.NewDevelopment()
}
