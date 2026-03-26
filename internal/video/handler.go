package video

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"streaming-backend/internal/auth"
)

// Handler handles video HTTP endpoints
type Handler struct {
	service *Service
}

// NewHandler creates a new video handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// UploadVideo handles POST /api/v1/videos
func (h *Handler) UploadVideo(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the Context
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized access")
		return
	}

	// Parse multipart form (max 100MB memory buffer)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large or invalid form")
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		writeError(w, http.StatusBadRequest, "video file is required in 'video' field")
		return
	}
	defer file.Close()

	input := UploadInput{
		UserID:      userID,
		Title:       title,
		Description: description,
		File:        file,
		Filename:    filepath.Base(header.Filename),
	}

	resp, err := h.service.UploadVideo(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload video")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// GetVideo handles GET /api/v1/videos/{id}
func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	if videoID == "" {
		writeError(w, http.StatusBadRequest, "video id is required")
		return
	}

	resp, err := h.service.GetVideo(r.Context(), videoID)
	if err != nil {
		writeError(w, http.StatusNotFound, "video not found")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListUserVideos handles GET /api/v1/videos
func (h *Handler) ListUserVideos(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized access")
		return
	}

	videos, err := h.service.ListUserVideos(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch videos")
		return
	}

	writeJSON(w, http.StatusOK, videos)
}

// ExploreVideos handles GET /api/v1/videos/explore
func (h *Handler) ExploreVideos(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	
	videos, err := h.service.ExploreVideos(r.Context(), query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch explore videos: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, videos)
}

// --- HTTP Helpers ---
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonBytes)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
