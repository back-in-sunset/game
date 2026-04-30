//go:build integration
// +build integration

package integration

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"game/server/core/testkit"
	"im/internal/auth"
	"im/internal/config"
	"im/internal/domain"
	"im/internal/store"
)

func TestMySQLRedisStore_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	ctx, dsn := testkit.StartMySQLContainer(t, "im")
	db := testkit.OpenMySQLWithRetry(t, ctx, dsn)
	mustInitIMTables(t, db)

	redisAddr := testkit.StartMiniRedis(t)
	imStore, err := store.NewMySQLRedisStore(config.Config{
		Mysql: config.Mysql{DataSource: dsn},
		Redis: config.Redis{Addr: redisAddr, DB: 0, KeyPrefix: "imtest"},
	})
	if err != nil {
		t.Fatalf("NewMySQLRedisStore() error = %v", err)
	}
	defer func() { _ = imStore.Close() }()

	scope := domain.Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"}
	msg1 := domain.Envelope{
		Domain:   domain.DomainTenant,
		Scope:    scope,
		Sender:   1001,
		Receiver: 1002,
		MsgType:  "direct_message",
		Seq:      1,
		Payload:  map[string]any{"text": "hello"},
		SentAt:   time.Now().UTC().Add(-time.Second),
	}
	msg2 := domain.Envelope{
		Domain:   domain.DomainTenant,
		Scope:    scope,
		Sender:   1002,
		Receiver: 1001,
		MsgType:  "direct_message",
		Seq:      2,
		Payload:  map[string]any{"text": "world"},
		SentAt:   time.Now().UTC(),
	}

	if err := imStore.SaveMessage(context.Background(), msg1); err != nil {
		t.Fatalf("SaveMessage(msg1) error = %v", err)
	}
	if err := imStore.SaveMessage(context.Background(), msg2); err != nil {
		t.Fatalf("SaveMessage(msg2) error = %v", err)
	}

	user1 := auth.Principal{UserID: 1001, Domain: domain.DomainTenant, Scope: scope}
	user2 := auth.Principal{UserID: 1002, Domain: domain.DomainTenant, Scope: scope}

	conversations, err := imStore.ListConversations(context.Background(), user1)
	if err != nil {
		t.Fatalf("ListConversations(user1) error = %v", err)
	}
	if len(conversations) != 1 {
		t.Fatalf("len(ListConversations(user1)) = %d, want 1", len(conversations))
	}
	if conversations[0].UnreadCount != 1 {
		t.Fatalf("UnreadCount = %d, want 1", conversations[0].UnreadCount)
	}

	history, err := imStore.ListMessages(context.Background(), user1, 1002, 20)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("len(ListMessages()) = %d, want 2", len(history))
	}
	if history[0].Seq != 1 || history[1].Seq != 2 {
		t.Fatalf("unexpected message seq order: %+v", history)
	}

	if err := imStore.MarkRead(context.Background(), user1, 1002, 2); err != nil {
		t.Fatalf("MarkRead() error = %v", err)
	}
	conversations, err = imStore.ListConversations(context.Background(), user1)
	if err != nil {
		t.Fatalf("ListConversations(user1) after mark read error = %v", err)
	}
	if conversations[0].UnreadCount != 0 {
		t.Fatalf("UnreadCount after MarkRead = %d, want 0", conversations[0].UnreadCount)
	}

	offline := domain.Envelope{
		Domain:   domain.DomainTenant,
		Scope:    scope,
		Sender:   1001,
		Receiver: 1002,
		MsgType:  "system_notice",
		Seq:      3,
		Payload:  map[string]any{"title": "notice"},
		SentAt:   time.Now().UTC(),
	}
	if err := imStore.SaveOffline(context.Background(), user2, offline); err != nil {
		t.Fatalf("SaveOffline() error = %v", err)
	}
	batch, err := imStore.DrainOffline(context.Background(), user2)
	if err != nil {
		t.Fatalf("DrainOffline() error = %v", err)
	}
	if len(batch) != 1 {
		t.Fatalf("len(DrainOffline()) = %d, want 1", len(batch))
	}
	batch, err = imStore.DrainOffline(context.Background(), user2)
	if err != nil {
		t.Fatalf("DrainOffline() second call error = %v", err)
	}
	if len(batch) != 0 {
		t.Fatalf("len(DrainOffline()) second call = %d, want 0", len(batch))
	}
}

func mustInitIMTables(t *testing.T, db *sql.DB) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "model", "im_init.sql"))
	if err != nil {
		t.Fatalf("ReadFile(im_init.sql) error = %v", err)
	}
	statements := strings.Split(string(data), ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("init IM tables failed on %q: %v", stmt, err)
		}
	}
}
