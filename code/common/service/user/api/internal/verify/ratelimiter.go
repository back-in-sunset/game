package verify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RateLimiter interface {
	AllowSend(ctx context.Context, scene string, target string, cooldown time.Duration, dailyLimit int64) error
}

type memoryRateLimiter struct {
	mu   sync.Mutex
	meta map[string]rateMeta
}

type rateMeta struct {
	lastSent time.Time
	day      string
	count    int64
}

func NewMemoryRateLimiter() RateLimiter {
	return &memoryRateLimiter{
		meta: make(map[string]rateMeta),
	}
}

func (l *memoryRateLimiter) AllowSend(_ context.Context, scene string, target string, cooldown time.Duration, dailyLimit int64) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := scene + ":" + target
	now := time.Now()
	meta := l.meta[key]
	day := now.Format("2006-01-02")
	if meta.day != day {
		meta.day = day
		meta.count = 0
	}

	if cooldown > 0 && !meta.lastSent.IsZero() && now.Sub(meta.lastSent) < cooldown {
		return fmt.Errorf("send too frequently")
	}
	if dailyLimit > 0 && meta.count >= dailyLimit {
		return fmt.Errorf("daily send limit exceeded")
	}

	meta.lastSent = now
	meta.count++
	l.meta[key] = meta
	return nil
}
