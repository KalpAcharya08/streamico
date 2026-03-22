package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// =============================================================================
// HANDLER LAYER (HTTP API)
// =============================================================================
// The Handler receives HTTP requests, decodes the JSON, calls the Service, 
// and writes the HTTP response. It does NOT contain business logic.

type Handler struct {
	service *Service
	cache   *Cache
}

// NewHandler creates a new auth handler.
func NewHandler(service *Service, cache *Cache) *Handler {
	return &Handler{service: service, cache: cache}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	// 1. Decode the JSON body into our struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	// 2. Call the service layer to do the actual work
	result, err := h.service.Register(r.Context(), req)

	// 3. Handle errors based on type
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, ErrUsernameExists):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, ErrValidationFailed):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			// For unexpected errors, log them implicitly by returning a generic 500 error
			writeError(w, http.StatusInternalServerError, "an internal error occurred")
		}
		return
	}

	// 4. Send the successful response (201 Created)
	writeJSON(w, http.StatusCreated, result)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	result, err := h.service.Login(r.Context(), req)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCreds):
			writeError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, ErrValidationFailed):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "an internal error occurred")
		}
		return
	}

	// Send success response (200 OK)
	writeJSON(w, http.StatusOK, result)
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusBadRequest, "missing authorization header")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeError(w, http.StatusBadRequest, "invalid authorization header")
		return
	}
	tokenString := parts[1]

	// Blacklist the token with a fixed TTL. 
	// In a real scenario, use duration until the token actually expires.
	err := h.cache.BlacklistToken(r.Context(), tokenString, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to process logout")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "successfully logged out"})
}

// =============================================================================
// HTTP HELPERS
// =============================================================================

// writeJSON is a helper to encode and send JSON HTTP responses.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonBytes)
}

// writeError is a helper specifically for sending JSON error messages.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
