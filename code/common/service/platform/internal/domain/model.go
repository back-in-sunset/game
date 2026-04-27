package domain

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrTenantIDRequired        = errors.New("tenant id is required")
	ErrTenantNameRequired      = errors.New("tenant name is required")
	ErrTenantSlugRequired      = errors.New("tenant slug is required")
	ErrTenantSlugInvalid       = errors.New("tenant slug is invalid")
	ErrTenantSlugExists        = errors.New("tenant slug already exists")
	ErrTenantNotFound          = errors.New("tenant not found")
	ErrProjectIDRequired       = errors.New("project id is required")
	ErrProjectNameRequired     = errors.New("project name is required")
	ErrProjectKeyRequired      = errors.New("project key is required")
	ErrProjectKeyInvalid       = errors.New("project key is invalid")
	ErrProjectKeyExists        = errors.New("project key already exists in tenant")
	ErrProjectNotFound         = errors.New("project not found")
	ErrEnvironmentIDRequired   = errors.New("environment id is required")
	ErrEnvironmentNameRequired = errors.New("environment name is required")
	ErrEnvironmentNameInvalid  = errors.New("environment name is invalid")
	ErrEnvironmentNameExists   = errors.New("environment name already exists in project")
)

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

type Tenant struct {
	ID   string
	Name string
	Slug string
}

type Project struct {
	ID       string
	TenantID string
	Name     string
	Key      string
}

type Environment struct {
	ID          string
	ProjectID   string
	Name        string
	DisplayName string
}

type Catalog struct {
	tenantsByID         map[string]*Tenant
	tenantIDsBySlug     map[string]string
	projectsByID        map[string]*Project
	projectIDsByScoped  map[string]string
	environmentsByID    map[string]*Environment
	environmentScopedID map[string]string
}

func NewCatalog() *Catalog {
	return &Catalog{
		tenantsByID:         make(map[string]*Tenant),
		tenantIDsBySlug:     make(map[string]string),
		projectsByID:        make(map[string]*Project),
		projectIDsByScoped:  make(map[string]string),
		environmentsByID:    make(map[string]*Environment),
		environmentScopedID: make(map[string]string),
	}
}

func NewTenant(id, name, slug string) (*Tenant, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)

	if id == "" {
		return nil, ErrTenantIDRequired
	}
	if name == "" {
		return nil, ErrTenantNameRequired
	}
	if slug == "" {
		return nil, ErrTenantSlugRequired
	}
	if !slugPattern.MatchString(slug) {
		return nil, ErrTenantSlugInvalid
	}

	return &Tenant{
		ID:   id,
		Name: name,
		Slug: slug,
	}, nil
}

func NewProject(id, tenantID, name, key string) (*Project, error) {
	id = strings.TrimSpace(id)
	tenantID = strings.TrimSpace(tenantID)
	name = strings.TrimSpace(name)
	key = strings.TrimSpace(key)

	if id == "" {
		return nil, ErrProjectIDRequired
	}
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if name == "" {
		return nil, ErrProjectNameRequired
	}
	if key == "" {
		return nil, ErrProjectKeyRequired
	}
	if !slugPattern.MatchString(key) {
		return nil, ErrProjectKeyInvalid
	}

	return &Project{
		ID:       id,
		TenantID: tenantID,
		Name:     name,
		Key:      key,
	}, nil
}

func NewEnvironment(id, projectID, name, displayName string) (*Environment, error) {
	id = strings.TrimSpace(id)
	projectID = strings.TrimSpace(projectID)
	name = strings.TrimSpace(name)
	displayName = strings.TrimSpace(displayName)

	if id == "" {
		return nil, ErrEnvironmentIDRequired
	}
	if projectID == "" {
		return nil, ErrProjectIDRequired
	}
	if name == "" {
		return nil, ErrEnvironmentNameRequired
	}
	if !slugPattern.MatchString(name) {
		return nil, ErrEnvironmentNameInvalid
	}

	return &Environment{
		ID:          id,
		ProjectID:   projectID,
		Name:        name,
		DisplayName: displayName,
	}, nil
}

func (c *Catalog) AddTenant(tenant *Tenant) error {
	if tenant == nil {
		return ErrTenantIDRequired
	}
	if _, exists := c.tenantIDsBySlug[tenant.Slug]; exists {
		return ErrTenantSlugExists
	}

	c.tenantsByID[tenant.ID] = tenant
	c.tenantIDsBySlug[tenant.Slug] = tenant.ID
	return nil
}

func (c *Catalog) AddProject(project *Project) error {
	if project == nil {
		return ErrProjectIDRequired
	}
	if _, exists := c.tenantsByID[project.TenantID]; !exists {
		return ErrTenantNotFound
	}

	scopedKey := project.TenantID + ":" + project.Key
	if _, exists := c.projectIDsByScoped[scopedKey]; exists {
		return ErrProjectKeyExists
	}

	c.projectsByID[project.ID] = project
	c.projectIDsByScoped[scopedKey] = project.ID
	return nil
}

func (c *Catalog) AddEnvironment(environment *Environment) error {
	if environment == nil {
		return ErrEnvironmentIDRequired
	}
	project, exists := c.projectsByID[environment.ProjectID]
	if !exists || project == nil {
		return ErrProjectNotFound
	}

	scopedKey := environment.ProjectID + ":" + environment.Name
	if _, exists := c.environmentScopedID[scopedKey]; exists {
		return ErrEnvironmentNameExists
	}

	c.environmentsByID[environment.ID] = environment
	c.environmentScopedID[scopedKey] = environment.ID
	return nil
}
