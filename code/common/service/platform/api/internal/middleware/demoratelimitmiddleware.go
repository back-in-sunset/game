package middleware

import (
	"net/http"
	"sync"
	"time"
)

const (
	demoMaxRequests  = 20              // max requests per window per IP
	demoWindow       = 1 * time.Minute // rate limit window
	demoCleanupIntvl = 5 * time.Minute // stale entry cleanup interval
	demoMaxEntries   = 10_000          // max tracked IPs before forced cleanup
)

type visitorRecord struct {
	timestamps []time.Time
	lastSeen   time.Time
}

type DemoRateLimitMiddleware struct {
	mu       sync.Mutex
	visitors map[string]*visitorRecord
}

func NewDemoRateLimitMiddleware() *DemoRateLimitMiddleware {
	m := &DemoRateLimitMiddleware{
		visitors: make(map[string]*visitorRecord),
	}
	go m.cleanupLoop()
	return m
}

func (m *DemoRateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !m.allow(ip) {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded, please retry later", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

func (m *DemoRateLimitMiddleware) allow(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-demoWindow)

	v, ok := m.visitors[ip]
	if !ok {
		m.visitors[ip] = &visitorRecord{
			timestamps: []time.Time{now},
			lastSeen:   now,
		}
		// Force cleanup if too many entries
		if len(m.visitors) > demoMaxEntries {
			m.pruneUnsafe(now)
		}
		return true
	}
	v.lastSeen = now

	// Evict expired timestamps
	i := 0
	for _, ts := range v.timestamps {
		if ts.After(cutoff) {
			v.timestamps[i] = ts
			i++
		}
	}
	v.timestamps = v.timestamps[:i]

	if len(v.timestamps) >= demoMaxRequests {
		return false
	}

	v.timestamps = append(v.timestamps, now)
	return true
}

func (m *DemoRateLimitMiddleware) pruneUnsafe(now time.Time) {
	cutoff := now.Add(-2 * demoWindow)
	for ip, v := range m.visitors {
		if v.lastSeen.Before(cutoff) {
			delete(m.visitors, ip)
		}
	}
}

func (m *DemoRateLimitMiddleware) cleanupLoop() {
	ticker := time.NewTicker(demoCleanupIntvl)
	defer ticker.Stop()
	for range ticker.C {
		m.mu.Lock()
		cutoff := time.Now().Add(-2 * demoWindow)
		for ip, v := range m.visitors {
			if v.lastSeen.Before(cutoff) {
				delete(m.visitors, ip)
			}
		}
		m.mu.Unlock()
	}
}
