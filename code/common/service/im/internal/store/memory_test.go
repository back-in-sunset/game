package store

import (
	"context"
	"testing"
	"time"

	"im/internal/auth"
	"im/internal/domain"
)

func TestMemoryStoreSeparatesTenantOfflineInbox(t *testing.T) {
	store := NewMemoryStore()
	envelope := domain.Envelope{
		Domain:   domain.DomainTenant,
		Scope:    domain.Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"},
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      1,
		SentAt:   time.Now(),
	}

	principalA := auth.Principal{UserID: 2, Domain: domain.DomainTenant, Scope: envelope.Scope}
	principalB := auth.Principal{UserID: 2, Domain: domain.DomainTenant, Scope: domain.Scope{TenantID: "tenant-2", ProjectID: "project-1", Environment: "prod"}}

	if err := store.SaveOffline(context.Background(), principalA, envelope); err != nil {
		t.Fatalf("SaveOffline() error = %v", err)
	}

	gotA, err := store.DrainOffline(context.Background(), principalA)
	if err != nil {
		t.Fatalf("DrainOffline(A) error = %v", err)
	}
	if len(gotA) != 1 {
		t.Fatalf("len(DrainOffline(A)) = %d, want 1", len(gotA))
	}

	gotB, err := store.DrainOffline(context.Background(), principalB)
	if err != nil {
		t.Fatalf("DrainOffline(B) error = %v", err)
	}
	if len(gotB) != 0 {
		t.Fatalf("len(DrainOffline(B)) = %d, want 0", len(gotB))
	}
}

func TestMemoryStoreListConversationsAndMarkRead(t *testing.T) {
	store := NewMemoryStore()
	scope := domain.Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"}
	ctx := context.Background()

	messages := []domain.Envelope{
		{Domain: domain.DomainTenant, Scope: scope, Sender: 1, Receiver: 2, MsgType: "direct_message", Seq: 1, SentAt: time.Unix(1, 0)},
		{Domain: domain.DomainTenant, Scope: scope, Sender: 2, Receiver: 1, MsgType: "direct_message", Seq: 2, SentAt: time.Unix(2, 0)},
	}
	for _, msg := range messages {
		if err := store.SaveMessage(ctx, msg); err != nil {
			t.Fatalf("SaveMessage() error = %v", err)
		}
	}

	items, err := store.ListConversations(ctx, auth.Principal{UserID: 1, Domain: domain.DomainTenant, Scope: scope})
	if err != nil {
		t.Fatalf("ListConversations() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].UnreadCount != 1 {
		t.Fatalf("UnreadCount = %d, want 1", items[0].UnreadCount)
	}

	if err := store.MarkRead(ctx, auth.Principal{UserID: 1, Domain: domain.DomainTenant, Scope: scope}, 2, 2); err != nil {
		t.Fatalf("MarkRead() error = %v", err)
	}
	items, err = store.ListConversations(ctx, auth.Principal{UserID: 1, Domain: domain.DomainTenant, Scope: scope})
	if err != nil {
		t.Fatalf("ListConversations() error = %v", err)
	}
	if items[0].UnreadCount != 0 {
		t.Fatalf("UnreadCount after MarkRead = %d, want 0", items[0].UnreadCount)
	}
}
