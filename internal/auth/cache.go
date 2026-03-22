package auth

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache manages authentication-related data in Redis
type Cache struct {
	client *redis.Client
}

// NewCache constructs an Auth Cache
func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

// BlacklistToken adds a token to the blacklist with the given expiration
func (c *Cache) BlacklistToken(ctx context.Context, token string, expiration time.Duration) error {
	return c.client.Set(ctx, "blacklist:"+token, "true", expiration).Err()
}

// IsBlacklisted checks if the provided token exists in the blacklist
func (c *Cache) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	err := c.client.Get(ctx, "blacklist:"+token).Err()
	if err == redis.Nil {
		return false, nil // Not in blacklist
	} else if err != nil {
		return false, err // Redis error
	}
	// Value exists, token is blacklisted
	return true, nil
}
