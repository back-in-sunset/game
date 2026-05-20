package limit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Limiter represents a distributed rate limiter using Redis
type Limiter struct {
	client *redis.Client
	key    string
	limit  int
	window time.Duration
}

// NewLimiter creates a new distributed rate limiter
func NewLimiter(redisClient *redis.Client, key string, limit int, window time.Duration) *Limiter {
	return &Limiter{
		client: redisClient,
		key:    key,
		limit:  limit,
		window: window,
	}
}

// Allow checks if the action is allowed under the rate limit
func (l *Limiter) Allow(ctx context.Context) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - l.window.Nanoseconds()

	pipe := l.client.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, l.key, "0", fmt.Sprint(windowStart))

	// Add current request
	pipe.ZAdd(ctx, l.key, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// Get count of requests in window
	pipe.ZCard(ctx, l.key)

	// Set expiration
	pipe.Expire(ctx, l.key, l.window)

	cmders, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute redis pipeline: %w", err)
	}

	count := cmders[2].(*redis.IntCmd).Val()
	return count <= int64(l.limit), nil
}

// Reset resets the rate limiter for the given key
func (l *Limiter) Reset(ctx context.Context) error {
	return l.client.Del(ctx, l.key).Err()
}
