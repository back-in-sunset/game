package model

import (
	"context"
	"strings"

	"platform/internal/domain"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type MySQLRepository struct {
	tenantModel       PlatformTenantModel
	projectModel      PlatformProjectModel
	environmentModel  PlatformEnvironmentModel
	tenantMemberModel PlatformTenantMemberModel
}

func NewMySQLRepository(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *MySQLRepository {
	return &MySQLRepository{
		tenantModel:       NewPlatformTenantModel(conn, c, opts...),
		projectModel:      NewPlatformProjectModel(conn, c, opts...),
		environmentModel:  NewPlatformEnvironmentModel(conn, c, opts...),
		tenantMemberModel: NewPlatformTenantMemberModel(conn, c, opts...),
	}
}

func NewMySQLRepositoryWithModels(
	tenantModel PlatformTenantModel,
	projectModel PlatformProjectModel,
	environmentModel PlatformEnvironmentModel,
	tenantMemberModel PlatformTenantMemberModel,
) *MySQLRepository {
	return &MySQLRepository{
		tenantModel:       tenantModel,
		projectModel:      projectModel,
		environmentModel:  environmentModel,
		tenantMemberModel: tenantMemberModel,
	}
}

func (r *MySQLRepository) SaveTenant(ctx context.Context, tenant *domain.Tenant) error {
	_, err := r.tenantModel.Insert(ctx, &PlatformTenant{
		TenantId: tenant.ID,
		Name:     tenant.Name,
		Slug:     tenant.Slug,
		Status:   "active",
	})
	return translateSQLError(err)
}

func (r *MySQLRepository) UpdateTenant(ctx context.Context, tenant *domain.Tenant) error {
	existing, err := r.tenantModel.FindOne(ctx, tenant.ID)
	if err != nil {
		if err == ErrNotFound {
			return domain.ErrTenantNotFound
		}
		return err
	}

	err = r.tenantModel.Update(ctx, &PlatformTenant{
		TenantId: tenant.ID,
		Name:     tenant.Name,
		Slug:     tenant.Slug,
		Status:   existing.Status,
	})
	return translateSQLError(err)
}

func (r *MySQLRepository) DeleteTenant(ctx context.Context, tenantID string) error {
	err := r.tenantModel.Delete(ctx, tenantID)
	if err == ErrNotFound {
		return domain.ErrTenantNotFound
	}
	return err
}

func (r *MySQLRepository) SaveProject(ctx context.Context, project *domain.Project) error {
	if _, err := r.tenantModel.FindOne(ctx, project.TenantID); err != nil {
		if err == ErrNotFound {
			return domain.ErrTenantNotFound
		}
		return err
	}

	_, err := r.projectModel.Insert(ctx, &PlatformProject{
		ProjectId:  project.ID,
		TenantId:   project.TenantID,
		Name:       project.Name,
		ProjectKey: project.Key,
		Status:     "active",
	})
	return translateSQLError(err)
}

func (r *MySQLRepository) GetTenantByID(ctx context.Context, id string) (*domain.Tenant, error) {
	tenant, err := r.tenantModel.FindOne(ctx, id)
	if err != nil {
		if err == ErrNotFound {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}
	return &domain.Tenant{
		ID:   tenant.TenantId,
		Name: tenant.Name,
		Slug: tenant.Slug,
	}, nil
}

func (r *MySQLRepository) ListProjectsByTenantID(ctx context.Context, tenantID string) ([]*domain.Project, error) {
	if _, err := r.tenantModel.FindOne(ctx, tenantID); err != nil {
		if err == ErrNotFound {
			return nil, domain.ErrTenantNotFound
		}
		return nil, err
	}

	projects, err := r.projectModel.ListByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	resp := make([]*domain.Project, 0, len(projects))
	for _, project := range projects {
		resp = append(resp, &domain.Project{
			ID:       project.ProjectId,
			TenantID: project.TenantId,
			Name:     project.Name,
			Key:      project.ProjectKey,
		})
	}
	return resp, nil
}

func (r *MySQLRepository) SaveEnvironment(ctx context.Context, environment *domain.Environment) error {
	if _, err := r.projectModel.FindOne(ctx, environment.ProjectID); err != nil {
		if err == ErrNotFound {
			return domain.ErrProjectNotFound
		}
		return err
	}

	_, err := r.environmentModel.Insert(ctx, &PlatformEnvironment{
		EnvironmentId: environment.ID,
		ProjectId:     environment.ProjectID,
		Name:          environment.Name,
		DisplayName:   environment.DisplayName,
		Status:        "active",
	})
	return translateSQLError(err)
}

func (r *MySQLRepository) ListEnvironmentsByProjectID(ctx context.Context, projectID string) ([]*domain.Environment, error) {
	if _, err := r.projectModel.FindOne(ctx, projectID); err != nil {
		if err == ErrNotFound {
			return nil, domain.ErrProjectNotFound
		}
		return nil, err
	}

	environments, err := r.environmentModel.ListByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	resp := make([]*domain.Environment, 0, len(environments))
	for _, environment := range environments {
		resp = append(resp, &domain.Environment{
			ID:          environment.EnvironmentId,
			ProjectID:   environment.ProjectId,
			Name:        environment.Name,
			DisplayName: environment.DisplayName,
		})
	}
	return resp, nil
}

func (r *MySQLRepository) ListTenantsByMemberID(ctx context.Context, memberID string) ([]*domain.Tenant, error) {
	members, err := r.tenantMemberModel.ListActiveByMemberID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	resp := make([]*domain.Tenant, 0, len(members))
	for _, member := range members {
		tenant, findErr := r.tenantModel.FindOne(ctx, member.TenantId)
		if findErr != nil {
			if findErr == ErrNotFound {
				continue
			}
			return nil, findErr
		}
		resp = append(resp, &domain.Tenant{
			ID:   tenant.TenantId,
			Name: tenant.Name,
			Slug: tenant.Slug,
		})
	}
	return resp, nil
}

func translateSQLError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "uk_platform_tenant_slug"):
		return domain.ErrTenantSlugExists
	case strings.Contains(msg, "uk_platform_project_tenant_key"):
		return domain.ErrProjectKeyExists
	case strings.Contains(msg, "uk_platform_environment_project_name"):
		return domain.ErrEnvironmentNameExists
	default:
		return err
	}
}
