package model

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"platform/internal/domain"
)

func TestMySQLRepositorySaveTenant(t *testing.T) {
	tenantModel := &stubTenantModel{
		insertResult: stubResult{rowsAffected: 1},
	}
	repo := NewMySQLRepositoryWithModels(tenantModel, &stubProjectModel{}, &stubEnvironmentModel{})
	tenant, _ := domain.NewTenant("tenant-1", "Acme Games", "acme-games")

	if err := repo.SaveTenant(context.Background(), tenant); err != nil {
		t.Fatalf("SaveTenant() error = %v", err)
	}
	if tenantModel.lastInsert == nil {
		t.Fatalf("expected insert to be called")
	}
	if tenantModel.lastInsert.Status != "active" {
		t.Fatalf("tenant status = %q, want %q", tenantModel.lastInsert.Status, "active")
	}
}

func TestMySQLRepositoryGetTenantByID(t *testing.T) {
	tenantModel := &stubTenantModel{
		findOneResp: &PlatformTenant{
			TenantId: "tenant-1",
			Name:     "Acme Games",
			Slug:     "acme-games",
		},
	}
	repo := NewMySQLRepositoryWithModels(tenantModel, &stubProjectModel{}, &stubEnvironmentModel{})

	tenant, err := repo.GetTenantByID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("GetTenantByID() error = %v", err)
	}
	if tenant.ID != "tenant-1" {
		t.Fatalf("tenant.ID = %q, want %q", tenant.ID, "tenant-1")
	}
}

func TestMySQLRepositoryListProjectsByTenantID(t *testing.T) {
	tenantModel := &stubTenantModel{
		findOneResp: &PlatformTenant{TenantId: "tenant-1"},
	}
	projectModel := &stubProjectModel{
		listByTenantIDResp: []*PlatformProject{
			{
				ProjectId:  "project-1",
				TenantId:   "tenant-1",
				Name:       "Project One",
				ProjectKey: "project-one",
			},
			{
				ProjectId:  "project-2",
				TenantId:   "tenant-1",
				Name:       "Project Two",
				ProjectKey: "project-two",
			},
		},
	}
	repo := NewMySQLRepositoryWithModels(tenantModel, projectModel, &stubEnvironmentModel{})

	projects, err := repo.ListProjectsByTenantID(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("ListProjectsByTenantID() error = %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("len(projects) = %d, want %d", len(projects), 2)
	}
}

func TestMySQLRepositoryListEnvironmentsByProjectID(t *testing.T) {
	projectModel := &stubProjectModel{
		findOneResp: &PlatformProject{ProjectId: "project-1"},
	}
	environmentModel := &stubEnvironmentModel{
		listByProjectIDResp: []*PlatformEnvironment{
			{
				EnvironmentId: "env-1",
				ProjectId:     "project-1",
				Name:          "dev",
				DisplayName:   "Development",
			},
			{
				EnvironmentId: "env-2",
				ProjectId:     "project-1",
				Name:          "test",
				DisplayName:   "Testing",
			},
		},
	}
	repo := NewMySQLRepositoryWithModels(&stubTenantModel{}, projectModel, environmentModel)

	environments, err := repo.ListEnvironmentsByProjectID(context.Background(), "project-1")
	if err != nil {
		t.Fatalf("ListEnvironmentsByProjectID() error = %v", err)
	}
	if len(environments) != 2 {
		t.Fatalf("len(environments) = %d, want %d", len(environments), 2)
	}
}

func TestMySQLRepositorySaveProjectValidatesTenantExists(t *testing.T) {
	tenantModel := &stubTenantModel{findOneErr: ErrNotFound}
	repo := NewMySQLRepositoryWithModels(tenantModel, &stubProjectModel{}, &stubEnvironmentModel{})

	err := repo.SaveProject(context.Background(), &domain.Project{
		ID:       "project-1",
		TenantID: "tenant-404",
		Name:     "Project One",
		Key:      "project-one",
	})
	if err != domain.ErrTenantNotFound {
		t.Fatalf("SaveProject() error = %v, want %v", err, domain.ErrTenantNotFound)
	}
}

func TestMySQLRepositorySaveEnvironmentValidatesProjectExists(t *testing.T) {
	projectModel := &stubProjectModel{findOneErr: ErrNotFound}
	repo := NewMySQLRepositoryWithModels(&stubTenantModel{}, projectModel, &stubEnvironmentModel{})

	err := repo.SaveEnvironment(context.Background(), &domain.Environment{
		ID:          "env-1",
		ProjectID:   "project-404",
		Name:        "dev",
		DisplayName: "Development",
	})
	if err != domain.ErrProjectNotFound {
		t.Fatalf("SaveEnvironment() error = %v, want %v", err, domain.ErrProjectNotFound)
	}
}

func TestMySQLRepositoryTranslatesDuplicateErrors(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		target  func(*MySQLRepository) error
		wantErr error
	}{
		{
			name: "tenant duplicate slug",
			err:  errors.New("Error 1062: Duplicate entry 'acme-games' for key 'uk_platform_tenant_slug'"),
			target: func(repo *MySQLRepository) error {
				return repo.SaveTenant(context.Background(), &domain.Tenant{ID: "tenant-1", Name: "Acme", Slug: "acme-games"})
			},
			wantErr: domain.ErrTenantSlugExists,
		},
		{
			name: "project duplicate key",
			err:  errors.New("Error 1062: Duplicate entry 'tenant-1-project-one' for key 'uk_platform_project_tenant_key'"),
			target: func(repo *MySQLRepository) error {
				return repo.SaveProject(context.Background(), &domain.Project{ID: "project-1", TenantID: "tenant-1", Name: "Project One", Key: "project-one"})
			},
			wantErr: domain.ErrProjectKeyExists,
		},
		{
			name: "environment duplicate name",
			err:  errors.New("Error 1062: Duplicate entry 'project-1-dev' for key 'uk_platform_environment_project_name'"),
			target: func(repo *MySQLRepository) error {
				return repo.SaveEnvironment(context.Background(), &domain.Environment{ID: "env-1", ProjectID: "project-1", Name: "dev", DisplayName: "Development"})
			},
			wantErr: domain.ErrEnvironmentNameExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantModel := &stubTenantModel{
				findOneResp: &PlatformTenant{TenantId: "tenant-1"},
				insertErr:   tt.err,
			}
			projectModel := &stubProjectModel{
				findOneResp: &PlatformProject{ProjectId: "project-1"},
				insertErr:   tt.err,
			}
			environmentModel := &stubEnvironmentModel{
				insertErr: tt.err,
			}
			repo := NewMySQLRepositoryWithModels(tenantModel, projectModel, environmentModel)
			err := tt.target(repo)
			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

type stubTenantModel struct {
	insertResult sql.Result
	insertErr    error
	findOneResp  *PlatformTenant
	findOneErr   error
	lastInsert   *PlatformTenant
}

func (s *stubTenantModel) Insert(_ context.Context, data *PlatformTenant) (sql.Result, error) {
	s.lastInsert = data
	if s.insertErr != nil {
		return nil, s.insertErr
	}
	if s.insertResult != nil {
		return s.insertResult, nil
	}
	return stubResult{rowsAffected: 1}, nil
}

func (s *stubTenantModel) FindOne(_ context.Context, _ string) (*PlatformTenant, error) {
	if s.findOneErr != nil {
		return nil, s.findOneErr
	}
	if s.findOneResp == nil {
		return nil, ErrNotFound
	}
	return s.findOneResp, nil
}

func (s *stubTenantModel) FindOneBySlug(_ context.Context, _ string) (*PlatformTenant, error) {
	return nil, ErrNotFound
}
func (s *stubTenantModel) Update(_ context.Context, _ *PlatformTenant) error { return nil }
func (s *stubTenantModel) Delete(_ context.Context, _ string) error          { return nil }

type stubProjectModel struct {
	insertResult       sql.Result
	insertErr          error
	findOneResp        *PlatformProject
	findOneErr         error
	listByTenantIDResp []*PlatformProject
}

func (s *stubProjectModel) Insert(_ context.Context, _ *PlatformProject) (sql.Result, error) {
	if s.insertErr != nil {
		return nil, s.insertErr
	}
	if s.insertResult != nil {
		return s.insertResult, nil
	}
	return stubResult{rowsAffected: 1}, nil
}

func (s *stubProjectModel) FindOne(_ context.Context, _ string) (*PlatformProject, error) {
	if s.findOneErr != nil {
		return nil, s.findOneErr
	}
	if s.findOneResp == nil {
		return nil, ErrNotFound
	}
	return s.findOneResp, nil
}

func (s *stubProjectModel) FindOneByTenantIdProjectKey(_ context.Context, _, _ string) (*PlatformProject, error) {
	return nil, ErrNotFound
}
func (s *stubProjectModel) Update(_ context.Context, _ *PlatformProject) error { return nil }
func (s *stubProjectModel) Delete(_ context.Context, _ string) error           { return nil }
func (s *stubProjectModel) ListByTenantID(_ context.Context, _ string) ([]*PlatformProject, error) {
	return s.listByTenantIDResp, nil
}

type stubEnvironmentModel struct {
	insertResult        sql.Result
	insertErr           error
	findOneResp         *PlatformEnvironment
	findOneErr          error
	listByProjectIDResp []*PlatformEnvironment
}

func (s *stubEnvironmentModel) Insert(_ context.Context, _ *PlatformEnvironment) (sql.Result, error) {
	if s.insertErr != nil {
		return nil, s.insertErr
	}
	if s.insertResult != nil {
		return s.insertResult, nil
	}
	return stubResult{rowsAffected: 1}, nil
}

func (s *stubEnvironmentModel) FindOne(_ context.Context, _ string) (*PlatformEnvironment, error) {
	if s.findOneErr != nil {
		return nil, s.findOneErr
	}
	if s.findOneResp == nil {
		return nil, ErrNotFound
	}
	return s.findOneResp, nil
}

func (s *stubEnvironmentModel) FindOneByProjectIdName(_ context.Context, _, _ string) (*PlatformEnvironment, error) {
	return nil, ErrNotFound
}
func (s *stubEnvironmentModel) Update(_ context.Context, _ *PlatformEnvironment) error { return nil }
func (s *stubEnvironmentModel) Delete(_ context.Context, _ string) error               { return nil }
func (s *stubEnvironmentModel) ListByProjectID(_ context.Context, _ string) ([]*PlatformEnvironment, error) {
	return s.listByProjectIDResp, nil
}

type stubResult struct {
	rowsAffected int64
}

func (s stubResult) LastInsertId() (int64, error) { return 0, nil }

func (s stubResult) RowsAffected() (int64, error) { return s.rowsAffected, nil }
