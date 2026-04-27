package svc

import (
	"context"
	"testing"
)

func TestScopeCarriesTenantProjectEnvironment(t *testing.T) {
	scope := Scope{
		TenantID:    "tenant-1",
		ProjectID:   "project-1",
		Environment: "dev",
	}

	if scope.TenantID != "tenant-1" {
		t.Fatalf("scope.TenantID = %q, want %q", scope.TenantID, "tenant-1")
	}
	if scope.ProjectID != "project-1" {
		t.Fatalf("scope.ProjectID = %q, want %q", scope.ProjectID, "project-1")
	}
	if scope.Environment != "dev" {
		t.Fatalf("scope.Environment = %q, want %q", scope.Environment, "dev")
	}
}

func TestWithScopeAndScopeFromContext(t *testing.T) {
	ctx := context.Background()
	scope := Scope{
		TenantID:    "tenant-1",
		ProjectID:   "project-1",
		Environment: "test",
	}

	ctx = WithScope(ctx, scope)
	got, ok := ScopeFromContext(ctx)
	if !ok {
		t.Fatalf("ScopeFromContext() ok = false, want true")
	}
	if got != scope {
		t.Fatalf("ScopeFromContext() = %+v, want %+v", got, scope)
	}
}

func TestResolveScopeFallsBackWhenMissing(t *testing.T) {
	fallback := Scope{
		TenantID:    "default-tenant",
		ProjectID:   "default-project",
		Environment: "local",
	}

	got := ResolveScope(context.Background(), fallback)
	if got != fallback {
		t.Fatalf("ResolveScope() = %+v, want %+v", got, fallback)
	}
}
