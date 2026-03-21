package video

import "time"

// VideoStatus represents the state of an uploaded video
type VideoStatus string

const (
	StatusProcessing VideoStatus = "processing"
	StatusReady      VideoStatus = "ready"
	StatusFailed     VideoStatus = "failed"
)

// Video represents the internal database record.
type Video struct {
	ID          string
	UserID      string
	Title       string
	Description string
	FilePath    string
	Status      VideoStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// VideoResponse is the safe JSON representation of a video.
type VideoResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
