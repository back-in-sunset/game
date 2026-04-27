package model

import (
	"sort"

	"platform/internal/domain"
)

type MemoryRepository struct {
	tenantsByID         map[string]*domain.Tenant
	tenantIDsBySlug     map[string]string
	projectsByID        map[string]*domain.Project
	projectIDsByScoped  map[string]string
	environmentsByID    map[string]*domain.Environment
	environmentScopedID map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tenantsByID:         make(map[string]*domain.Tenant),
		tenantIDsBySlug:     make(map[string]string),
		projectsByID:        make(map[string]*domain.Project),
		projectIDsByScoped:  make(map[string]string),
		environmentsByID:    make(map[string]*domain.Environment),
		environmentScopedID: make(map[string]string),
	}
}

func (r *MemoryRepository) SaveTenant(tenant *domain.Tenant) error {
	if tenant == nil {
		return domain.ErrTenantIDRequired
	}
	if existingID, exists := r.tenantIDsBySlug[tenant.Slug]; exists && existingID != tenant.ID {
		return domain.ErrTenantSlugExists
	}

	r.tenantsByID[tenant.ID] = tenant
	r.tenantIDsBySlug[tenant.Slug] = tenant.ID
	return nil
}

func (r *MemoryRepository) GetTenantByID(id string) (*domain.Tenant, error) {
	tenant, exists := r.tenantsByID[id]
	if !exists {
		return nil, domain.ErrTenantNotFound
	}
	return tenant, nil
}

func (r *MemoryRepository) SaveProject(project *domain.Project) error {
	if project == nil {
		return domain.ErrProjectIDRequired
	}
	if _, exists := r.tenantsByID[project.TenantID]; !exists {
		return domain.ErrTenantNotFound
	}

	scopedKey := project.TenantID + ":" + project.Key
	if existingID, exists := r.projectIDsByScoped[scopedKey]; exists && existingID != project.ID {
		return domain.ErrProjectKeyExists
	}

	r.projectsByID[project.ID] = project
	r.projectIDsByScoped[scopedKey] = project.ID
	return nil
}

func (r *MemoryRepository) ListProjectsByTenantID(tenantID string) ([]*domain.Project, error) {
	if _, exists := r.tenantsByID[tenantID]; !exists {
		return nil, domain.ErrTenantNotFound
	}

	projects := make([]*domain.Project, 0)
	for _, project := range r.projectsByID {
		if project.TenantID == tenantID {
			projects = append(projects, project)
		}
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].ID < projects[j].ID
	})
	return projects, nil
}

func (r *MemoryRepository) SaveEnvironment(environment *domain.Environment) error {
	if environment == nil {
		return domain.ErrEnvironmentIDRequired
	}
	if _, exists := r.projectsByID[environment.ProjectID]; !exists {
		return domain.ErrProjectNotFound
	}

	scopedKey := environment.ProjectID + ":" + environment.Name
	if existingID, exists := r.environmentScopedID[scopedKey]; exists && existingID != environment.ID {
		return domain.ErrEnvironmentNameExists
	}

	r.environmentsByID[environment.ID] = environment
	r.environmentScopedID[scopedKey] = environment.ID
	return nil
}

func (r *MemoryRepository) ListEnvironmentsByProjectID(projectID string) ([]*domain.Environment, error) {
	if _, exists := r.projectsByID[projectID]; !exists {
		return nil, domain.ErrProjectNotFound
	}

	environments := make([]*domain.Environment, 0)
	for _, environment := range r.environmentsByID {
		if environment.ProjectID == projectID {
			environments = append(environments, environment)
		}
	}

	sort.Slice(environments, func(i, j int) bool {
		return environments[i].ID < environments[j].ID
	})
	return environments, nil
}
