package scope

import (
	"testing"

	"im/internal/config"
	"im/internal/domain"
)

func TestResolverResolveTenantDefaultEnvironment(t *testing.T) {
	resolver := NewResolver(config.ScopeDefaults{DefaultEnvironment: "prod"})

	scope, err := resolver.Resolve(domain.DomainTenant, domain.Scope{
		TenantID:  "tenant-1",
		ProjectID: "project-1",
	})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if scope.Environment != "prod" {
		t.Fatalf("scope.Environment = %q, want %q", scope.Environment, "prod")
	}
}
