package logic

import (
	"context"
	"errors"
	"strings"
	"testing"

	"game/server/core/testkit/unit"
	"platform/api/internal/svc"
	"platform/api/internal/types"
	"platform/internal/domain"
)

func TestCreateTenantLogic_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		req     types.CreateTenantReq
		repo    *stubRepo
		wantErr string
	}{
		{
			name: "success",
			req: types.CreateTenantReq{
				Name: "Acme Games",
				Slug: "acme-games",
			},
			repo: &stubRepo{},
		},
		{
			name: "invalid input",
			req: types.CreateTenantReq{
				Name: "",
				Slug: "acme-games",
			},
			repo:    &stubRepo{},
			wantErr: "name and slug are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewCreateTenantLogic(context.Background(), &svc.ServiceContext{Repo: tt.repo})
			resp, err := l.CreateTenant(&tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("CreateTenant() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			unit.RequireNoError(t, err)
			if resp.Id == "" {
				t.Fatalf("resp.Id is empty")
			}
			unit.RequireEqual(t, resp.Name, "Acme Games")
			unit.RequireEqual(t, resp.Slug, "acme-games")
		})
	}
}

func TestGetTenantLogic_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		req     types.GetTenantReq
		repo    *stubRepo
		wantErr string
	}{
		{
			name: "success",
			req:  types.GetTenantReq{TenantId: "tenant-1"},
			repo: &stubRepo{
				getTenantByIDResp: &domain.Tenant{
					ID:   "tenant-1",
					Name: "Acme",
					Slug: "acme",
				},
			},
		},
		{
			name:    "not found",
			req:     types.GetTenantReq{TenantId: "tenant-404"},
			repo:    &stubRepo{getTenantByIDErr: domain.ErrTenantNotFound},
			wantErr: domain.ErrTenantNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewGetTenantLogic(context.Background(), &svc.ServiceContext{Repo: tt.repo})
			resp, err := l.GetTenant(&tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("GetTenant() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			unit.RequireNoError(t, err)
			unit.RequireEqual(t, resp.Id, "tenant-1")
			unit.RequireEqual(t, resp.Name, "Acme")
		})
	}
}

func TestUpdateTenantLogic_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		req     types.UpdateTenantReq
		repo    *stubRepo
		wantErr string
	}{
		{
			name: "success",
			req: types.UpdateTenantReq{
				TenantId: "tenant-1",
				Name:     "Acme Updated",
				Slug:     "acme-updated",
			},
			repo: &stubRepo{},
		},
		{
			name: "invalid",
			req: types.UpdateTenantReq{
				TenantId: "",
				Name:     "Acme Updated",
				Slug:     "acme-updated",
			},
			repo:    &stubRepo{},
			wantErr: "tenantId, name and slug are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewUpdateTenantLogic(context.Background(), &svc.ServiceContext{Repo: tt.repo})
			resp, err := l.UpdateTenant(&tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("UpdateTenant() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			unit.RequireNoError(t, err)
			unit.RequireEqual(t, resp.Id, "tenant-1")
			unit.RequireEqual(t, resp.Name, "Acme Updated")
		})
	}
}

func TestDeleteTenantLogic_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		req     types.DeleteTenantReq
		repo    *stubRepo
		wantErr string
	}{
		{
			name: "success",
			req:  types.DeleteTenantReq{TenantId: "tenant-1"},
			repo: &stubRepo{},
		},
		{
			name:    "invalid",
			req:     types.DeleteTenantReq{TenantId: ""},
			repo:    &stubRepo{},
			wantErr: "tenantId is required",
		},
		{
			name:    "repo error",
			req:     types.DeleteTenantReq{TenantId: "tenant-1"},
			repo:    &stubRepo{deleteTenantErr: domain.ErrTenantNotFound},
			wantErr: domain.ErrTenantNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewDeleteTenantLogic(context.Background(), &svc.ServiceContext{Repo: tt.repo})
			resp, err := l.DeleteTenant(&tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("DeleteTenant() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			unit.RequireNoError(t, err)
			unit.RequireEqual(t, resp.Success, true)
		})
	}
}

func TestMyTenantsLogic_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		req       types.MyTenantsReq
		repo      *stubRepo
		wantErr   string
		wantItems int
	}{
		{
			name:    "missing uid",
			ctx:     context.Background(),
			req:     types.MyTenantsReq{},
			repo:    &stubRepo{},
			wantErr: "x-uid is required",
		},
		{
			name: "from header",
			ctx:  context.Background(),
			req:  types.MyTenantsReq{MemberId: "u-1"},
			repo: &stubRepo{
				listTenantsByMemberIDResp: []*domain.Tenant{
					{ID: "tenant-1", Name: "Acme", Slug: "acme"},
				},
			},
			wantItems: 1,
		},
		{
			name: "from context uid",
			ctx:  unit.ContextWithValue(types.UserIDKey, "u-ctx"),
			req:  types.MyTenantsReq{MemberId: "u-header"},
			repo: &stubRepo{
				listTenantsByMemberIDResp: []*domain.Tenant{
					{ID: "tenant-1", Name: "Acme", Slug: "acme"},
				},
			},
			wantItems: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewMyTenantsLogic(tt.ctx, &svc.ServiceContext{Repo: tt.repo})
			resp, err := l.MyTenants(&tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("MyTenants() error = %v, want contains %q", err, tt.wantErr)
				}
				return
			}

			unit.RequireNoError(t, err)
			unit.RequireEqual(t, len(resp.Items), tt.wantItems)
		})
	}
}

type stubRepo struct {
	getTenantByIDResp         *domain.Tenant
	getTenantByIDErr          error
	deleteTenantErr           error
	listTenantsByMemberIDResp []*domain.Tenant
	listTenantsByMemberIDErr  error
}

func (s *stubRepo) SaveTenant(context.Context, *domain.Tenant) error { return nil }
func (s *stubRepo) UpdateTenant(context.Context, *domain.Tenant) error {
	return nil
}
func (s *stubRepo) DeleteTenant(context.Context, string) error {
	return s.deleteTenantErr
}
func (s *stubRepo) GetTenantByID(context.Context, string) (*domain.Tenant, error) {
	if s.getTenantByIDErr != nil {
		return nil, s.getTenantByIDErr
	}
	if s.getTenantByIDResp == nil {
		return nil, domain.ErrTenantNotFound
	}
	return s.getTenantByIDResp, nil
}
func (s *stubRepo) SaveProject(context.Context, *domain.Project) error { return nil }
func (s *stubRepo) ListProjectsByTenantID(context.Context, string) ([]*domain.Project, error) {
	return nil, errors.New("not implemented")
}
func (s *stubRepo) SaveEnvironment(context.Context, *domain.Environment) error { return nil }
func (s *stubRepo) ListEnvironmentsByProjectID(context.Context, string) ([]*domain.Environment, error) {
	return nil, errors.New("not implemented")
}
func (s *stubRepo) ListTenantsByMemberID(context.Context, string) ([]*domain.Tenant, error) {
	if s.listTenantsByMemberIDErr != nil {
		return nil, s.listTenantsByMemberIDErr
	}
	return s.listTenantsByMemberIDResp, nil
}
