package session

import (
	"context"
	"sync"

	"im/internal/auth"
	"im/internal/contracts"
)

type Manager struct {
	mu    sync.RWMutex
	conns map[string]contracts.Connection
	users map[string][]string
}

func NewManager() *Manager {
	return &Manager{
		conns: make(map[string]contracts.Connection),
		users: make(map[string][]string),
	}
}

func (m *Manager) Bind(principal auth.Principal, conn contracts.Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	connID := conn.ID()
	key := principalKey(principal)
	m.conns[connID] = conn
	m.users[key] = append(m.users[key], connID)
}

func (m *Manager) Unbind(principal auth.Principal, connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.conns, connID)
	key := principalKey(principal)
	ids := m.users[key][:0]
	for _, id := range m.users[key] {
		if id != connID {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		delete(m.users, key)
		return
	}
	m.users[key] = ids
}

func (m *Manager) SendToPrincipal(ctx context.Context, principal auth.Principal, payload []byte) (int, error) {
	m.mu.RLock()
	ids := append([]string(nil), m.users[principalKey(principal)]...)
	conns := make([]contracts.Connection, 0, len(ids))
	for _, id := range ids {
		if conn, ok := m.conns[id]; ok {
			conns = append(conns, conn)
		}
	}
	m.mu.RUnlock()

	sent := 0
	for _, conn := range conns {
		if err := conn.Send(ctx, payload); err != nil {
			return sent, err
		}
		sent++
	}
	return sent, nil
}

func principalKey(principal auth.Principal) string {
	scope := principal.Scope.Normalize()
	if principal.Domain == "platform" {
		return "platform:" + itoa(principal.UserID)
	}
	return "tenant:" + scope.TenantID + ":" + scope.ProjectID + ":" + scope.Environment + ":" + itoa(principal.UserID)
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
