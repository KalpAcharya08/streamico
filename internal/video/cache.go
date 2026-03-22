package video

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache manages video metadata in Redis
type Cache struct {
	client *redis.Client
}

// NewCache constructs a Video Cache
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// GetVideo fetches a single video's metadata from cache
func (c *Cache) GetVideo(ctx context.Context, id string) (*VideoResponse, error) {
	val, err := c.client.Get(ctx, "video:"+id).Result()
	if err != nil {
		return nil, err // Can be redis.Nil for cache miss
	}
	var v VideoResponse
	if err := json.Unmarshal([]byte(val), &v); err != nil {
		return nil, err
	}
	return &v, nil
}

// SetVideo saves a single video's metadata to cache
func (c *Cache) SetVideo(ctx context.Context, v *VideoResponse, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "video:"+v.ID, data, ttl).Err()
}

// GetUserVideos fetches a user's list of videos from cache
func (c *Cache) GetUserVideos(ctx context.Context, userID string) ([]*VideoResponse, error) {
	val, err := c.client.Get(ctx, "user_videos:"+userID).Result()
	if err != nil {
		return nil, err
	}
	var videos []*VideoResponse
	if err := json.Unmarshal([]byte(val), &videos); err != nil {
		return nil, err
	}
	return videos, nil
}

// SetUserVideos saves a user's list of videos to cache
func (c *Cache) SetUserVideos(ctx context.Context, userID string, videos []*VideoResponse, ttl time.Duration) error {
	data, err := json.Marshal(videos)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "user_videos:"+userID, data, ttl).Err()
}

// InvalidateUserVideos removes a user's cached video list
func (c *Cache) InvalidateUserVideos(ctx context.Context, userID string) error {
	return c.client.Del(ctx, "user_videos:"+userID).Err()
}
