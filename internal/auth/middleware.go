package auth

import (
	"context"
	"net/http"
	"strings"

	"streaming-backend/pkg/jwt"
)

// Define custom types for Context keys to prevent collisions.
type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

// Middleware struct holds dependencies needed by our middleware functions.
type Middleware struct {
	jwtSecret string
}

// NewMiddleware creates a new auth middleware instance.
func NewMiddleware(jwtSecret string) *Middleware {
	return &Middleware{jwtSecret: jwtSecret}
}

// RequireAuth is a Chi middleware that enforces a valid JWT token.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the Authorization header from the request
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// 2. Extract the token string (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}
		tokenString := parts[1]

		// 3. Validate the token
		claims, err := jwt.ValidateToken(tokenString, m.jwtSecret)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// 4. Inject the UserID and Role into the HTTP Request Context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		// 5. Create a new request with the updated context and pass it to the next handler
		reqWithContext := r.WithContext(ctx)
		next.ServeHTTP(w, reqWithContext)
	})
}
