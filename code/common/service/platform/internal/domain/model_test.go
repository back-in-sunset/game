package domain

import "testing"

func TestNewTenantSuccess(t *testing.T) {
	tenant, err := NewTenant("tenant-1", "Acme Games", "acme-games")
	if err != nil {
		t.Fatalf("NewTenant() error = %v", err)
	}
	if tenant.ID != "tenant-1" {
		t.Fatalf("tenant.ID = %q, want %q", tenant.ID, "tenant-1")
	}
	if tenant.Name != "Acme Games" {
		t.Fatalf("tenant.Name = %q, want %q", tenant.Name, "Acme Games")
	}
	if tenant.Slug != "acme-games" {
		t.Fatalf("tenant.Slug = %q, want %q", tenant.Slug, "acme-games")
	}
}

func TestNewTenantEmptyName(t *testing.T) {
	_, err := NewTenant("tenant-1", "", "acme-games")
	if err != ErrTenantNameRequired {
		t.Fatalf("NewTenant() error = %v, want %v", err, ErrTenantNameRequired)
	}
}

func TestCatalogRejectsDuplicateTenantSlug(t *testing.T) {
	catalog := NewCatalog()

	tenant1, _ := NewTenant("tenant-1", "Acme Games", "acme-games")
	tenant2, _ := NewTenant("tenant-2", "Beta Games", "acme-games")

	if err := catalog.AddTenant(tenant1); err != nil {
		t.Fatalf("AddTenant() error = %v", err)
	}
	if err := catalog.AddTenant(tenant2); err != ErrTenantSlugExists {
		t.Fatalf("AddTenant() error = %v, want %v", err, ErrTenantSlugExists)
	}
}

func TestCatalogRejectsProjectWithoutTenant(t *testing.T) {
	catalog := NewCatalog()
	project, _ := NewProject("project-1", "tenant-404", "Project One", "project-one")

	err := catalog.AddProject(project)
	if err != ErrTenantNotFound {
		t.Fatalf("AddProject() error = %v, want %v", err, ErrTenantNotFound)
	}
}

func TestCatalogRejectsDuplicateProjectKeyInSameTenant(t *testing.T) {
	catalog := NewCatalog()
	tenant, _ := NewTenant("tenant-1", "Acme Games", "acme-games")
	project1, _ := NewProject("project-1", "tenant-1", "Project One", "project-one")
	project2, _ := NewProject("project-2", "tenant-1", "Project Two", "project-one")

	if err := catalog.AddTenant(tenant); err != nil {
		t.Fatalf("AddTenant() error = %v", err)
	}
	if err := catalog.AddProject(project1); err != nil {
		t.Fatalf("AddProject() error = %v", err)
	}
	if err := catalog.AddProject(project2); err != ErrProjectKeyExists {
		t.Fatalf("AddProject() error = %v, want %v", err, ErrProjectKeyExists)
	}
}

func TestCatalogRejectsEnvironmentWithoutProject(t *testing.T) {
	catalog := NewCatalog()
	env, _ := NewEnvironment("env-1", "project-404", "dev", "Development")

	err := catalog.AddEnvironment(env)
	if err != ErrProjectNotFound {
		t.Fatalf("AddEnvironment() error = %v, want %v", err, ErrProjectNotFound)
	}
}

func TestCatalogRejectsDuplicateEnvironmentNameInSameProject(t *testing.T) {
	catalog := NewCatalog()
	tenant, _ := NewTenant("tenant-1", "Acme Games", "acme-games")
	project, _ := NewProject("project-1", "tenant-1", "Project One", "project-one")
	env1, _ := NewEnvironment("env-1", "project-1", "dev", "Development")
	env2, _ := NewEnvironment("env-2", "project-1", "dev", "Duplicate Development")

	if err := catalog.AddTenant(tenant); err != nil {
		t.Fatalf("AddTenant() error = %v", err)
	}
	if err := catalog.AddProject(project); err != nil {
		t.Fatalf("AddProject() error = %v", err)
	}
	if err := catalog.AddEnvironment(env1); err != nil {
		t.Fatalf("AddEnvironment() error = %v", err)
	}
	if err := catalog.AddEnvironment(env2); err != ErrEnvironmentNameExists {
		t.Fatalf("AddEnvironment() error = %v, want %v", err, ErrEnvironmentNameExists)
	}
}

func TestNewEnvironmentRejectsInvalidName(t *testing.T) {
	_, err := NewEnvironment("env-1", "project-1", "Prod Env", "Production")
	if err != ErrEnvironmentNameInvalid {
		t.Fatalf("NewEnvironment() error = %v, want %v", err, ErrEnvironmentNameInvalid)
	}
}
