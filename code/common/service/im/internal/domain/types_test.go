package domain

import "testing"

func TestScopeValidate(t *testing.T) {
	t.Run("platform allows empty scope", func(t *testing.T) {
		if err := (Scope{}).Validate(DomainPlatform); err != nil {
			t.Fatalf("Validate() error = %v", err)
		}
	})

	t.Run("tenant requires full coordinates", func(t *testing.T) {
		err := (Scope{TenantID: "t1"}).Validate(DomainTenant)
		if err == nil {
			t.Fatal("expected validation error")
		}
	})
}

func TestConversationKey(t *testing.T) {
	got, err := ConversationKey(DomainTenant, Scope{
		TenantID:    "tenant-1",
		ProjectID:   "project-1",
		Environment: "prod",
	}, 200, 100)
	if err != nil {
		t.Fatalf("ConversationKey() error = %v", err)
	}

	want := "tenant:tenant-1:project-1:prod:100:200"
	if got != want {
		t.Fatalf("ConversationKey() = %q, want %q", got, want)
	}
}
