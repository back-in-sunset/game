package limit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTest(t *testing.T) (*Limiter, *miniredis.Miniredis, context.Context) {
	// Start a mini Redis server for testing
	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	// Create a Redis client connected to the mini server
	client := redis.NewClient(&redis.Options{
		Addr: mini.Addr(),
	})

	// Create a limiter with test parameters
	limiter := NewLimiter(client, "test-key", 3, time.Second)
	ctx := context.Background()

	return limiter, mini, ctx
}

func TestLimiterAllow(t *testing.T) {
	limiter, mini, ctx := setupTest(t)
	defer mini.Close()

	// First three requests should be allowed
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx)
		if err != nil {
			t.Fatalf("Failed to check rate limit: %v", err)
		}
		if !allowed {
			t.Errorf("Request %d should be allowed but was denied", i+1)
		}
	}

	// Fourth request should be denied
	allowed, err := limiter.Allow(ctx)
	if err != nil {
		t.Fatalf("Failed to check rate limit: %v", err)
	}
	if allowed {
		t.Error("Fourth request should be denied but was allowed")
	}

	// After window passes, requests should be allowed again
	mini.FastForward(time.Second)
	allowed, err = limiter.Allow(ctx)
	if err != nil {
		t.Fatalf("Failed to check rate limit: %v", err)
	}
	if !allowed {
		t.Error("Request after window expiration should be allowed but was denied")
	}
}

func TestLimiterReset(t *testing.T) {
	limiter, mini, ctx := setupTest(t)
	defer mini.Close()

	// Fill up the limit
	for i := 0; i < 5; i++ {
		_, err := limiter.Allow(ctx)
		if err != nil {
			t.Fatalf("Failed to check rate limit: %v", err)
		}
	}

	// Verify limit is reached
	allowed, err := limiter.Allow(ctx)
	if err != nil {
		t.Fatalf("Failed to check rate limit: %v", err)
	}
	if allowed {
		t.Error("Request should be denied but was allowed")
	}

	// Reset the limiter
	err = limiter.Reset(ctx)
	if err != nil {
		t.Fatalf("Failed to reset limiter: %v", err)
	}

	// After reset, request should be allowed
	allowed, err = limiter.Allow(ctx)
	if err != nil {
		t.Fatalf("Failed to check rate limit: %v", err)
	}
	if !allowed {
		t.Error("Request after reset should be allowed but was denied")
	}
}

func TestLimiterWithRedisFailure(t *testing.T) {
	limiter, mini, ctx := setupTest(t)

	// Simulate Redis server going down
	mini.Close()

	// Attempt to use the limiter should return an error
	_, err := limiter.Allow(ctx)
	if err == nil {
		t.Error("Expected error when Redis is down, but got nil")
	}

	// Reset should also return an error
	err = limiter.Reset(ctx)
	if err == nil {
		t.Error("Expected error when Redis is down, but got nil")
	}
}
