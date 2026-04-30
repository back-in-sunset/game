package logic

import (
	"testing"

	"im/internal/domain"
	"im/rpc/im"
)

func TestPrincipalFromRequest(t *testing.T) {
	principal, err := principalFromRequest("tenant", &im.Scope{
		TenantID:    "tenant-1",
		ProjectID:   "project-1",
		Environment: "prod",
	}, 1001)
	if err != nil {
		t.Fatalf("principalFromRequest() error = %v", err)
	}
	if principal.Domain != domain.DomainTenant {
		t.Fatalf("principal.Domain = %q, want %q", principal.Domain, domain.DomainTenant)
	}
	if principal.Scope.TenantID != "tenant-1" {
		t.Fatalf("principal.Scope.TenantID = %q, want tenant-1", principal.Scope.TenantID)
	}
}

func TestPayloadFromJSON(t *testing.T) {
	payload, err := payloadFromJSON(`{"text":"hello","count":1}`)
	if err != nil {
		t.Fatalf("payloadFromJSON() error = %v", err)
	}
	if payload["text"] != "hello" {
		t.Fatalf("payload[text] = %v, want hello", payload["text"])
	}
	if _, ok := payload["count"]; !ok {
		t.Fatal("payload[count] missing")
	}
}
