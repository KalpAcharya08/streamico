package user

import (
	"encoding/json"
	"net/http"

	"streaming-backend/internal/auth" // We import auth to extract the Context keys
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetMe handles GET /api/v1/users/me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the Context (put there by Auth Middleware)
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized access")
		return
	}

	// Call the service
	profile, err := h.service.GetProfileByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user profile not found")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// --- HTTP Helpers (copying locally for the package) ---
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonBytes)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
