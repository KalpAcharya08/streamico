package video

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// UploadInput represents the data needed to upload a video.
type UploadInput struct {
	UserID      string
	Title       string
	Description string
	File        io.Reader
	Filename    string
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// generateID creates a simple secure hex ID for videos instead of requiring google/uuid
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback block if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

// UploadVideo processes the video upload, saves the file, and stores metadata to the DB.
func (s *Service) UploadVideo(ctx context.Context, input UploadInput) (*VideoResponse, error) {
	// 1. Generate unique file IDs
	videoID := generateID()
	ext := filepath.Ext(input.Filename)
	if ext == "" {
		ext = ".mp4" // Default extension
	}
	newFilename := fmt.Sprintf("%s%s", videoID, ext)
	
	// 2. Define storage path
	uploadDir := filepath.Join("uploads", "videos")
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	
	savePath := filepath.Join(uploadDir, newFilename)

	// 3. Save file to disk
	out, err := os.Create(savePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, input.File); err != nil {
		return nil, fmt.Errorf("failed to save video file: %w", err)
	}

	// 4. Save metadata to database
	now := time.Now().UTC()
	video := &Video{
		ID:          videoID,
		UserID:      input.UserID,
		Title:       input.Title,
		Description: input.Description,
		FilePath:    savePath,
		Status:      StatusProcessing, // Start as processing while transcoding happens
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateVideo(ctx, video); err != nil {
		// Clean up the file if DB fails
		os.Remove(savePath) 
		return nil, fmt.Errorf("failed to save video metadata: %w", err)
	}

	// 5. Start background transcoding job (Simulated)
	// VERY IMPORTANT INTERN LESSON: 
	// We pass context.Background() because the original `ctx` will be cancelled the moment 
	// we send the HTTP Response! The goroutine needs a clear context to keep running.
	go s.processVideo(context.Background(), video.ID, savePath)

	// 6. Return safe response immediately!
	resp := &VideoResponse{
		ID:          video.ID,
		UserID:      video.UserID,
		Title:       video.Title,
		Description: video.Description,
		Status:      string(video.Status),
		CreatedAt:   video.CreatedAt,
	}

	return resp, nil
}

// processVideo transcodes the video into HLS chunks and an m3u8 playlist using FFmpeg.
func (s *Service) processVideo(ctx context.Context, videoID string, rawFilePath string) {
	fmt.Printf("🎬 [FFMPEG WORKER] Starting HLS Transcoding for video %s...\n", videoID)
	
	// 1. Create a dedicated directory for the HLS files
	hlsDir := filepath.Join("uploads", "videos", videoID)
	if err := os.MkdirAll(hlsDir, os.ModePerm); err != nil {
		fmt.Printf("❌ [FFMPEG WORKER] Failed to create HLS directory: %v\n", err)
		s.repo.UpdateVideoStatus(ctx, videoID, StatusFailed)
		return
	}

	// 2. Define the playlist file path
	playlistPath := filepath.Join(hlsDir, "playlist.m3u8")

	// 3. Build the FFmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", rawFilePath,
		"-profile:v", "baseline",
		"-level", "3.0",
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		playlistPath,
	)

	// We capture combined output (stdout and stderr) where FFmpeg writes its logs
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ [FFMPEG WORKER] Transcoding failed for %s: %v\nLogs:\n%s\n", videoID, err, string(output))
		s.repo.UpdateVideoStatus(ctx, videoID, StatusFailed)
		return
	}

	// 4. Update the Database with the new playlist path and mark as ready!
	if err := s.repo.UpdateVideoPath(ctx, videoID, playlistPath); err != nil {
		fmt.Printf("❌ [FFMPEG WORKER] Failed to update video path in DB: %v\n", err)
	}

	if err := s.repo.UpdateVideoStatus(ctx, videoID, StatusReady); err != nil {
		fmt.Printf("❌ [FFMPEG WORKER] Failed to update video status in DB: %v\n", err)
	}

	// 5. Cleanup: Delete the huge initial .mp4 file to save space on the server disk
	os.Remove(rawFilePath)

	fmt.Printf("✅ [FFMPEG WORKER] Transcoding complete! %s is ready via HLS at %s\n", videoID, playlistPath)
}

func (s *Service) GetVideo(ctx context.Context, id string) (*VideoResponse, error) {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	return &VideoResponse{
		ID:          v.ID,
		UserID:      v.UserID,
		Title:       v.Title,
		Description: v.Description,
		Status:      string(v.Status),
		CreatedAt:   v.CreatedAt,
	}, nil
}

func (s *Service) ListUserVideos(ctx context.Context, userID string) ([]*VideoResponse, error) {
	videos, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var responses []*VideoResponse
	for _, v := range videos {
		responses = append(responses, &VideoResponse{
			ID:          v.ID,
			UserID:      v.UserID,
			Title:       v.Title,
			Description: v.Description,
			Status:      string(v.Status),
			CreatedAt:   v.CreatedAt,
		})
	}
	return responses, nil
}
