package session

import (
	"context"
	"hash/fnv"
	"sync"

	"im/internal/auth"
	"im/internal/contracts"
)

type Manager struct {
	buckets []bucket
}

type bucket struct {
	mu    sync.RWMutex
	conns map[string]contracts.Connection
	users map[string][]string
}

func NewManager() *Manager {
	return NewManagerWithBuckets(64)
}

func NewManagerWithBuckets(bucketCount int) *Manager {
	if bucketCount <= 0 {
		bucketCount = 64
	}
	m := &Manager{buckets: make([]bucket, bucketCount)}
	for i := range m.buckets {
		m.buckets[i] = bucket{
			conns: make(map[string]contracts.Connection),
			users: make(map[string][]string),
		}
	}
	return m
}

func (m *Manager) Bind(principal auth.Principal, conn contracts.Connection) {
	key := principalKey(principal)
	b := m.bucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	connID := conn.ID()
	b.conns[connID] = conn
	b.users[key] = append(b.users[key], connID)
}

func (m *Manager) Unbind(principal auth.Principal, connID string) {
	key := principalKey(principal)
	b := m.bucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.conns, connID)
	existing := b.users[key]
	if len(existing) == 0 {
		return
	}
	ids := existing[:0]
	for _, id := range existing {
		if id != connID {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		delete(b.users, key)
		return
	}
	b.users[key] = ids
}

func (m *Manager) SendToPrincipal(ctx context.Context, principal auth.Principal, payload []byte) (int, error) {
	key := principalKey(principal)
	b := m.bucket(key)

	b.mu.RLock()
	ids := append([]string(nil), b.users[key]...)
	conns := make([]contracts.Connection, 0, len(ids))
	for _, id := range ids {
		if conn, ok := b.conns[id]; ok {
			conns = append(conns, conn)
		}
	}
	b.mu.RUnlock()

	sent := 0
	for _, conn := range conns {
		if err := conn.Send(ctx, payload); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}

func (m *Manager) bucket(key string) *bucket {
	idx := int(hashKey(key) % uint32(len(m.buckets)))
	return &m.buckets[idx]
}

func principalKey(principal auth.Principal) string {
	scope := principal.Scope.Normalize()
	if principal.Domain == "platform" {
		return "platform:" + itoa(principal.UserID)
	}
	return "tenant:" + scope.TenantID + ":" + scope.ProjectID + ":" + scope.Environment + ":" + itoa(principal.UserID)
}

func hashKey(key string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return h.Sum32()
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return sign + string(buf[i:])
}
