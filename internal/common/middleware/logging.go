package middleware

// =============================================================================
// 📖 LEARNING NOTES — Request Logger Middleware
// =============================================================================
//
// WHAT IS MIDDLEWARE?
// Middleware is a function that "wraps" your handlers. It runs code BEFORE
// and/or AFTER each request — like a security guard at a door.
//
//   Request → [Logging MW] → [CORS MW] → [Your Handler] → Response
//              ↑ starts timer                               ↑ logs duration
//
// THE MIDDLEWARE PATTERN IN GO:
//   func(next http.Handler) http.Handler
//
//   - "next" is the handler AFTER this middleware (could be another middleware
//     or the actual handler)
//   - You call next.ServeHTTP(w, r) to pass the request along the chain
//   - Code BEFORE next.ServeHTTP runs before the request is handled
//   - Code AFTER  next.ServeHTTP runs after the response is sent
// =============================================================================

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// RequestLogger logs every HTTP request with method, path, and duration.
//
// Why return a function that returns a function? This is called a "closure" —
// it lets us inject the logger once, and Chi calls the inner function for
// every request automatically.
func RequestLogger(log *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ── BEFORE the request ──────────────────────────────
			start := time.Now()

			// Pass the request to the next handler in the chain
			next.ServeHTTP(w, r)

			// ── AFTER the request ───────────────────────────────
			duration := time.Since(start)

			// Log the request details using structured fields
			log.Info("http request",
				zap.String("method", r.Method),         // GET, POST, etc.
				zap.String("path", r.URL.Path),          // /health, /api/v1/...
				zap.String("remote", r.RemoteAddr),      // Client IP address
				zap.Duration("duration", duration),      // How long it took
			)
		})
	}
}
