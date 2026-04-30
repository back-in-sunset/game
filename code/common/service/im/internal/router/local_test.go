package router

import (
	"context"
	"errors"
	"testing"
	"time"

	"im/internal/auth"
	"im/internal/contracts"
	"im/internal/domain"
	"im/internal/session"
	"im/internal/store"
)

func TestLocalRouterStoresOfflineWhenReceiverOffline(t *testing.T) {
	sessions := session.NewManager()
	mem := store.NewMemoryStore()
	router := NewLocalRouter("", sessions, mem, nil, nil, nil)

	envelope := domain.Envelope{
		Domain:   domain.DomainTenant,
		Scope:    domain.Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"},
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      1,
		Payload:  map[string]any{"text": "hi"},
		SentAt:   time.Now(),
	}

	result, err := router.Deliver(context.Background(), envelope)
	if err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if !result.StoredOffline || result.OnlineRecipients != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}

	items, err := mem.DrainOffline(context.Background(), auth.Principal{
		UserID: 2,
		Domain: domain.DomainTenant,
		Scope:  envelope.Scope,
	})
	if err != nil {
		t.Fatalf("DrainOffline() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
}

func TestLocalRouterDeliversOnline(t *testing.T) {
	sessions := session.NewManager()
	mem := store.NewMemoryStore()
	router := NewLocalRouter("", sessions, mem, nil, nil, nil)

	receiver := auth.Principal{
		UserID: 2,
		Domain: domain.DomainPlatform,
	}
	conn := &testConn{id: "c1"}
	sessions.Bind(receiver, conn)

	envelope := domain.Envelope{
		Domain:   domain.DomainPlatform,
		Scope:    domain.Scope{},
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      1,
		Payload:  map[string]any{"text": "hi"},
		SentAt:   time.Now(),
	}

	result, err := router.Deliver(context.Background(), envelope)
	if err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if result.OnlineRecipients != 1 || result.StoredOffline {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(conn.sent) != 1 {
		t.Fatalf("len(conn.sent) = %d, want 1", len(conn.sent))
	}
}

func TestLocalRouterForwardsToRemoteNode(t *testing.T) {
	sessions := session.NewManager()
	mem := store.NewMemoryStore()
	presence := fakePresence{nodeID: "node-b", ok: true}
	registry := fakeRegistry{
		node: contracts.Node{ID: "node-b", RPC: "127.0.0.1:9092"},
	}
	forwarder := &fakeForwarder{delivered: 1}
	router := NewLocalRouter("node-a", sessions, mem, registry, presence, forwarder)

	envelope := domain.Envelope{
		Domain:   domain.DomainPlatform,
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      1,
		Payload:  map[string]any{"text": "hi"},
		SentAt:   time.Now(),
	}

	result, err := router.Deliver(context.Background(), envelope)
	if err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if result.OnlineRecipients != 1 || result.StoredOffline {
		t.Fatalf("unexpected result: %+v", result)
	}
	if forwarder.calls != 1 {
		t.Fatalf("forwarder.calls = %d, want 1", forwarder.calls)
	}
}

func TestLocalRouterFallsBackOfflineWhenRemoteForwardFails(t *testing.T) {
	sessions := session.NewManager()
	mem := store.NewMemoryStore()
	presence := fakePresence{nodeID: "node-b", ok: true}
	registry := fakeRegistry{
		node: contracts.Node{ID: "node-b", RPC: "127.0.0.1:9092"},
	}
	forwarder := &fakeForwarder{err: errors.New("remote down")}
	router := NewLocalRouter("node-a", sessions, mem, registry, presence, forwarder)

	envelope := domain.Envelope{
		Domain:   domain.DomainTenant,
		Scope:    domain.Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"},
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      1,
		Payload:  map[string]any{"text": "hi"},
		SentAt:   time.Now(),
	}

	result, err := router.Deliver(context.Background(), envelope)
	if err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if !result.StoredOffline || result.OnlineRecipients != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}

	items, err := mem.DrainOffline(context.Background(), auth.Principal{
		UserID: 2,
		Domain: domain.DomainTenant,
		Scope:  envelope.Scope,
	})
	if err != nil {
		t.Fatalf("DrainOffline() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
}

func TestLocalRouterFallsBackOfflineWhenPresencePointsLocalButNoSession(t *testing.T) {
	sessions := session.NewManager()
	mem := store.NewMemoryStore()
	presence := fakePresence{nodeID: "node-a", ok: true}
	router := NewLocalRouter("node-a", sessions, mem, nil, presence, nil)

	envelope := domain.Envelope{
		Domain:   domain.DomainPlatform,
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      1,
		Payload:  map[string]any{"text": "hi"},
		SentAt:   time.Now(),
	}

	result, err := router.Deliver(context.Background(), envelope)
	if err != nil {
		t.Fatalf("Deliver() error = %v", err)
	}
	if !result.StoredOffline {
		t.Fatalf("expected offline fallback, got %+v", result)
	}
}

type testConn struct {
	id   string
	sent [][]byte
	err  error
}

func (c *testConn) ID() string { return c.id }

func (c *testConn) Send(_ context.Context, payload []byte) error {
	if c.err != nil {
		return c.err
	}
	cp := make([]byte, len(payload))
	copy(cp, payload)
	c.sent = append(c.sent, cp)
	return nil
}

func (c *testConn) Close() error { return errors.New("not implemented") }

type fakePresence struct {
	nodeID string
	ok     bool
	err    error
}

func (f fakePresence) Lookup(context.Context, auth.Principal) (string, bool, error) {
	return f.nodeID, f.ok, f.err
}

type fakeRegistry struct {
	node contracts.Node
	err  error
}

func (f fakeRegistry) Register(context.Context, contracts.Node) error { return nil }
func (f fakeRegistry) GetNode(context.Context, string) (contracts.Node, error) {
	return f.node, f.err
}
func (f fakeRegistry) Close() error { return nil }

type fakeForwarder struct {
	delivered int
	err       error
	calls     int
}

func (f *fakeForwarder) Forward(context.Context, string, domain.Envelope) (int, error) {
	f.calls++
	return f.delivered, f.err
}
