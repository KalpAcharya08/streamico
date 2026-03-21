package middleware

// =============================================================================
// 📖 LEARNING NOTES — CORS Middleware
// =============================================================================
//
// WHAT IS CORS?
// CORS = Cross-Origin Resource Sharing
//
// When your frontend (localhost:3000) calls your backend (localhost:8080),
// the browser blocks the request by default for security. CORS headers tell
// the browser: "it's okay, I allow requests from this origin."
//
// Without CORS middleware → browser blocks frontend-to-backend calls
// With CORS middleware    → browser allows the requests
//
// HOW IT WORKS:
// 1. Browser sends a "preflight" OPTIONS request first
// 2. Server responds with allowed origins/methods/headers
// 3. Browser checks the response — if allowed, sends the real request
// =============================================================================

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
// This allows frontend apps on different domains/ports to call our API.
func CORS() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		// AllowedOrigins: which domains can call your API
		// "*" = allow everyone (fine for development, restrict in production)
		AllowedOrigins: []string{"*"},

		// AllowedMethods: which HTTP methods are allowed
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},

		// AllowedHeaders: which headers the client can send
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},

		// AllowCredentials: allow cookies/auth headers to be sent
		AllowCredentials: true,

		// MaxAge: how long (seconds) the browser caches preflight results
		// 300 = 5 minutes (reduces preflight requests)
		MaxAge: 300,
	})
}
