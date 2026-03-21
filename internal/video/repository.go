package video

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateVideo inserts a new video record into the database.
func (r *Repository) CreateVideo(ctx context.Context, v *Video) error {
	query := `INSERT INTO videos (id, user_id, title, description, file_path, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := r.db.ExecContext(ctx, query, v.ID, v.UserID, v.Title, v.Description, v.FilePath, v.Status, v.CreatedAt, v.UpdatedAt)
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}
	return nil
}

// GetByID retrieves a video by its UUID.
func (r *Repository) GetByID(ctx context.Context, id string) (*Video, error) {
	query := `
		SELECT id, user_id, title, description, file_path, status, created_at, updated_at
		FROM videos
		WHERE id = ?
	`
	var v Video
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&v.ID, &v.UserID, &v.Title, &v.Description, &v.FilePath, &v.Status, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("video not found: %w", err)
		}
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	return &v, nil
}

// ListByUserID retrieves all videos uploaded by a specific user.
func (r *Repository) ListByUserID(ctx context.Context, userID string) ([]*Video, error) {
	query := `
		SELECT id, user_id, title, description, file_path, status, created_at, updated_at
		FROM videos
		WHERE user_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		var v Video
		if err := rows.Scan(&v.ID, &v.UserID, &v.Title, &v.Description, &v.FilePath, &v.Status, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan video row: %w", err)
		}
		videos = append(videos, &v)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return videos, nil
}

// UpdateVideoStatus updates the database record state (e.g., from processing to ready).
func (r *Repository) UpdateVideoStatus(ctx context.Context, id string, status VideoStatus) error {
	query := `UPDATE videos SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}
	return nil
}

// UpdateVideoPath updates the file path to point to the new HLS playlist.
func (r *Repository) UpdateVideoPath(ctx context.Context, id string, path string) error {
	query := `UPDATE videos SET file_path = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, path, id)
	if err != nil {
		return fmt.Errorf("failed to update video path: %w", err)
	}
	return nil
}
