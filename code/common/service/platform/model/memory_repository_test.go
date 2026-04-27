package model

import (
	"testing"

	"platform/internal/domain"
)

func TestMemoryRepositorySaveAndGetTenant(t *testing.T) {
	repo := NewMemoryRepository()
	tenant, _ := domain.NewTenant("tenant-1", "Acme Games", "acme-games")

	if err := repo.SaveTenant(tenant); err != nil {
		t.Fatalf("SaveTenant() error = %v", err)
	}

	got, err := repo.GetTenantByID("tenant-1")
	if err != nil {
		t.Fatalf("GetTenantByID() error = %v", err)
	}
	if got.ID != tenant.ID {
		t.Fatalf("got tenant id = %q, want %q", got.ID, tenant.ID)
	}
}

func TestMemoryRepositoryListsProjectsByTenant(t *testing.T) {
	repo := NewMemoryRepository()
	tenant, _ := domain.NewTenant("tenant-1", "Acme Games", "acme-games")
	project1, _ := domain.NewProject("project-1", "tenant-1", "Project One", "project-one")
	project2, _ := domain.NewProject("project-2", "tenant-1", "Project Two", "project-two")

	mustSaveTenant(t, repo, tenant)
	mustSaveProject(t, repo, project1)
	mustSaveProject(t, repo, project2)

	projects, err := repo.ListProjectsByTenantID("tenant-1")
	if err != nil {
		t.Fatalf("ListProjectsByTenantID() error = %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("len(projects) = %d, want %d", len(projects), 2)
	}
}

func TestMemoryRepositoryListsEnvironmentsByProject(t *testing.T) {
	repo := NewMemoryRepository()
	tenant, _ := domain.NewTenant("tenant-1", "Acme Games", "acme-games")
	project, _ := domain.NewProject("project-1", "tenant-1", "Project One", "project-one")
	env1, _ := domain.NewEnvironment("env-1", "project-1", "dev", "Development")
	env2, _ := domain.NewEnvironment("env-2", "project-1", "test", "Testing")

	mustSaveTenant(t, repo, tenant)
	mustSaveProject(t, repo, project)
	mustSaveEnvironment(t, repo, env1)
	mustSaveEnvironment(t, repo, env2)

	environments, err := repo.ListEnvironmentsByProjectID("project-1")
	if err != nil {
		t.Fatalf("ListEnvironmentsByProjectID() error = %v", err)
	}
	if len(environments) != 2 {
		t.Fatalf("len(environments) = %d, want %d", len(environments), 2)
	}
}

func TestMemoryRepositoryRejectsDuplicateTenantSlug(t *testing.T) {
	repo := NewMemoryRepository()
	tenant1, _ := domain.NewTenant("tenant-1", "Acme Games", "acme-games")
	tenant2, _ := domain.NewTenant("tenant-2", "Beta Games", "acme-games")

	mustSaveTenant(t, repo, tenant1)
	err := repo.SaveTenant(tenant2)
	if err != domain.ErrTenantSlugExists {
		t.Fatalf("SaveTenant() error = %v, want %v", err, domain.ErrTenantSlugExists)
	}
}

func TestMemoryRepositoryRejectsDuplicateProjectKeyInTenant(t *testing.T) {
	repo := NewMemoryRepository()
	tenant, _ := domain.NewTenant("tenant-1", "Acme Games", "acme-games")
	project1, _ := domain.NewProject("project-1", "tenant-1", "Project One", "project-one")
	project2, _ := domain.NewProject("project-2", "tenant-1", "Project Two", "project-one")

	mustSaveTenant(t, repo, tenant)
	mustSaveProject(t, repo, project1)

	err := repo.SaveProject(project2)
	if err != domain.ErrProjectKeyExists {
		t.Fatalf("SaveProject() error = %v, want %v", err, domain.ErrProjectKeyExists)
	}
}

func TestMemoryRepositoryRejectsDuplicateEnvironmentNameInProject(t *testing.T) {
	repo := NewMemoryRepository()
	tenant, _ := domain.NewTenant("tenant-1", "Acme Games", "acme-games")
	project, _ := domain.NewProject("project-1", "tenant-1", "Project One", "project-one")
	env1, _ := domain.NewEnvironment("env-1", "project-1", "dev", "Development")
	env2, _ := domain.NewEnvironment("env-2", "project-1", "dev", "Duplicate Development")

	mustSaveTenant(t, repo, tenant)
	mustSaveProject(t, repo, project)
	mustSaveEnvironment(t, repo, env1)

	err := repo.SaveEnvironment(env2)
	if err != domain.ErrEnvironmentNameExists {
		t.Fatalf("SaveEnvironment() error = %v, want %v", err, domain.ErrEnvironmentNameExists)
	}
}

func mustSaveTenant(t *testing.T, repo *MemoryRepository, tenant *domain.Tenant) {
	t.Helper()
	if err := repo.SaveTenant(tenant); err != nil {
		t.Fatalf("SaveTenant() error = %v", err)
	}
}

func mustSaveProject(t *testing.T, repo *MemoryRepository, project *domain.Project) {
	t.Helper()
	if err := repo.SaveProject(project); err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}
}

func mustSaveEnvironment(t *testing.T, repo *MemoryRepository, environment *domain.Environment) {
	t.Helper()
	if err := repo.SaveEnvironment(environment); err != nil {
		t.Fatalf("SaveEnvironment() error = %v", err)
	}
}
