package presence

import (
	"context"
	"testing"

	"im/internal/auth"
	"im/internal/config"
	"im/internal/domain"

	"github.com/alicebob/miniredis/v2"
)

func TestTrackerBindLookupUnbind(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	defer mr.Close()

	tracker := NewTracker(config.Redis{
		Addr:      mr.Addr(),
		DB:        0,
		KeyPrefix: "imtest",
	})
	defer func() { _ = tracker.Close() }()

	principal := auth.Principal{
		UserID: 1001,
		Domain: domain.DomainTenant,
		Scope: domain.Scope{
			TenantID:    "tenant-1",
			ProjectID:   "project-1",
			Environment: "prod",
		},
	}

	if err := tracker.Bind(context.Background(), principal, "node-a"); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	nodeID, ok, err := tracker.Lookup(context.Background(), principal)
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if !ok || nodeID != "node-a" {
		t.Fatalf("Lookup() = (%q, %v), want (%q, true)", nodeID, ok, "node-a")
	}

	if err := tracker.Unbind(context.Background(), principal, "node-a"); err != nil {
		t.Fatalf("Unbind() error = %v", err)
	}
	_, ok, err = tracker.Lookup(context.Background(), principal)
	if err != nil {
		t.Fatalf("Lookup() after Unbind error = %v", err)
	}
	if ok {
		t.Fatal("expected no presence after unbind")
	}
}
