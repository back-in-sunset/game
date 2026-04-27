package logic

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"game/server/core/testkit"
	"platform/api/internal/svc"
	"platform/api/internal/types"
	"platform/internal/domain"
	"platform/model"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestCreateTenantLogic_TableDriven(t *testing.T) {
	ctx, db, repo := mustNewTenantTestRepo(t)

	tests := []struct {
		name    string
		req     types.CreateTenantReq
		wantErr string
	}{
		{
			name: "success",
			req: types.CreateTenantReq{
				Name: "Acme Games",
				Slug: "acme-games",
			},
		},
		{
			name: "invalid input",
			req: types.CreateTenantReq{
				Name: "",
				Slug: "acme-games",
			},
			wantErr: "name and slug are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustResetTenantTables(t, db)
			l := NewCreateTenantLogic(ctx, &svc.ServiceContext{Repo: repo})
			resp, err := l.CreateTenant(&tt.req)
			if tt.wantErr != "" {
				requireErrContains(t, err, tt.wantErr)
				return
			}

			requireNoError(t, err)
			if resp.Id == "" {
				t.Fatalf("resp.Id is empty")
			}
			requireEqual(t, resp.Name, "Acme Games")
			requireEqual(t, resp.Slug, "acme-games")
			requireTenant(t, db, resp.Id, "Acme Games", "acme-games")
		})
	}
}

func TestGetTenantLogic_TableDriven(t *testing.T) {
	ctx, db, repo := mustNewTenantTestRepo(t)

	tests := []struct {
		name    string
		req     types.GetTenantReq
		seed    func(t *testing.T, db *sql.DB)
		wantErr string
	}{
		{
			name: "success",
			req:  types.GetTenantReq{TenantId: "tenant-1"},
			seed: func(t *testing.T, db *sql.DB) {
				mustInsertTenant(t, db, "tenant-1", "Acme", "acme")
			},
		},
		{
			name:    "not found",
			req:     types.GetTenantReq{TenantId: "tenant-404"},
			wantErr: domain.ErrTenantNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustResetTenantTables(t, db)
			if tt.seed != nil {
				tt.seed(t, db)
			}
			l := NewGetTenantLogic(ctx, &svc.ServiceContext{Repo: repo})
			resp, err := l.GetTenant(&tt.req)
			if tt.wantErr != "" {
				requireErrContains(t, err, tt.wantErr)
				return
			}

			requireNoError(t, err)
			requireEqual(t, resp.Id, "tenant-1")
			requireEqual(t, resp.Name, "Acme")
		})
	}
}

func TestUpdateTenantLogic_TableDriven(t *testing.T) {
	ctx, db, repo := mustNewTenantTestRepo(t)

	tests := []struct {
		name    string
		req     types.UpdateTenantReq
		seed    func(t *testing.T, db *sql.DB)
		wantErr string
	}{
		{
			name: "success",
			req: types.UpdateTenantReq{
				TenantId: "tenant-1",
				Name:     "Acme Updated",
				Slug:     "acme-updated",
			},
			seed: func(t *testing.T, db *sql.DB) {
				mustInsertTenant(t, db, "tenant-1", "Acme", "acme")
			},
		},
		{
			name: "invalid",
			req: types.UpdateTenantReq{
				TenantId: "",
				Name:     "Acme Updated",
				Slug:     "acme-updated",
			},
			wantErr: "tenantId, name and slug are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustResetTenantTables(t, db)
			if tt.seed != nil {
				tt.seed(t, db)
			}
			l := NewUpdateTenantLogic(ctx, &svc.ServiceContext{Repo: repo})
			resp, err := l.UpdateTenant(&tt.req)
			if tt.wantErr != "" {
				requireErrContains(t, err, tt.wantErr)
				return
			}

			requireNoError(t, err)
			requireEqual(t, resp.Id, "tenant-1")
			requireEqual(t, resp.Name, "Acme Updated")
			requireTenant(t, db, "tenant-1", "Acme Updated", "acme-updated")
		})
	}
}

func TestDeleteTenantLogic_TableDriven(t *testing.T) {
	ctx, db, repo := mustNewTenantTestRepo(t)

	tests := []struct {
		name    string
		req     types.DeleteTenantReq
		seed    func(t *testing.T, db *sql.DB)
		wantErr string
	}{
		{
			name: "success",
			req:  types.DeleteTenantReq{TenantId: "tenant-1"},
			seed: func(t *testing.T, db *sql.DB) {
				mustInsertTenant(t, db, "tenant-1", "Acme", "acme")
			},
		},
		{
			name:    "invalid",
			req:     types.DeleteTenantReq{TenantId: ""},
			wantErr: "tenantId is required",
		},
		{
			name:    "repo error",
			req:     types.DeleteTenantReq{TenantId: "tenant-1"},
			wantErr: domain.ErrTenantNotFound.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustResetTenantTables(t, db)
			if tt.seed != nil {
				tt.seed(t, db)
			}
			l := NewDeleteTenantLogic(ctx, &svc.ServiceContext{Repo: repo})
			resp, err := l.DeleteTenant(&tt.req)
			if tt.wantErr != "" {
				requireErrContains(t, err, tt.wantErr)
				return
			}

			requireNoError(t, err)
			requireEqual(t, resp.Success, true)
			requireTenantDeleted(t, db, "tenant-1")
		})
	}
}

func TestMyTenantsLogic_TableDriven(t *testing.T) {
	ctx, db, repo := mustNewTenantTestRepo(t)

	tests := []struct {
		name      string
		ctx       context.Context
		req       types.MyTenantsReq
		seed      func(t *testing.T, db *sql.DB)
		wantErr   string
		wantItems int
	}{
		{
			name:    "missing uid",
			ctx:     context.Background(),
			req:     types.MyTenantsReq{},
			wantErr: "x-uid is required",
		},
		{
			name: "from header",
			ctx:  ctx,
			req:  types.MyTenantsReq{MemberId: "u-1"},
			seed: func(t *testing.T, db *sql.DB) {
				mustInsertTenant(t, db, "tenant-1", "Acme", "acme")
				mustInsertTenantMember(t, db, "tm-1", "tenant-1", "u-1", "owner", "active")
			},
			wantItems: 1,
		},
		{
			name: "from context uid",
			ctx:  context.WithValue(ctx, types.UserIDKey, "u-ctx"),
			req:  types.MyTenantsReq{MemberId: "u-header"},
			seed: func(t *testing.T, db *sql.DB) {
				mustInsertTenant(t, db, "tenant-1", "Acme", "acme")
				mustInsertTenantMember(t, db, "tm-2", "tenant-1", "u-ctx", "developer", "active")
			},
			wantItems: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mustResetTenantTables(t, db)
			if tt.seed != nil {
				tt.seed(t, db)
			}
			l := NewMyTenantsLogic(tt.ctx, &svc.ServiceContext{Repo: repo})
			resp, err := l.MyTenants(&tt.req)
			if tt.wantErr != "" {
				requireErrContains(t, err, tt.wantErr)
				return
			}

			requireNoError(t, err)
			requireEqual(t, len(resp.Items), tt.wantItems)
		})
	}
}

func mustNewTenantTestRepo(t *testing.T) (context.Context, *sql.DB, svc.Repository) {
	t.Helper()

	ctx, dsn := testkit.StartMySQLContainer(t, "platform")
	db := testkit.OpenMySQLWithRetry(t, ctx, dsn)
	mustExecSchema(t, db)

	redisAddr := testkit.StartMiniRedis(t)
	repo := model.NewMySQLRepository(sqlx.NewMysql(dsn), cache.CacheConf{
		{
			RedisConf: redis.RedisConf{
				Host: redisAddr,
				Type: "node",
			},
		},
	})
	return ctx, db, repo
}

func mustExecSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	stmts := []string{
		`CREATE TABLE platform_tenant (
tenant_id varchar(64) NOT NULL,
name varchar(128) NOT NULL,
slug varchar(64) NOT NULL,
status varchar(32) NOT NULL DEFAULT 'active',
created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (tenant_id),
UNIQUE KEY uk_platform_tenant_slug (slug)
)`,
		`CREATE TABLE platform_tenant_member (
tenant_member_id varchar(64) NOT NULL,
tenant_id varchar(64) NOT NULL,
member_id varchar(64) NOT NULL,
role varchar(32) NOT NULL DEFAULT 'developer',
status varchar(32) NOT NULL DEFAULT 'active',
joined_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (tenant_member_id),
UNIQUE KEY uk_platform_tenant_member (tenant_id, member_id),
KEY idx_platform_member_status (member_id, status)
)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("db.Exec(schema) error = %v", err)
		}
	}
}

func mustResetTenantTables(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec("DELETE FROM platform_tenant_member"); err != nil {
		t.Fatalf("reset tenant_member error: %v", err)
	}
	if _, err := db.Exec("DELETE FROM platform_tenant"); err != nil {
		t.Fatalf("reset tenant error: %v", err)
	}
}

func mustInsertTenant(t *testing.T, db *sql.DB, id, name, slug string) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO platform_tenant (tenant_id, name, slug, status) VALUES (?, ?, ?, 'active')", id, name, slug); err != nil {
		t.Fatalf("insert tenant error: %v", err)
	}
}

func mustInsertTenantMember(t *testing.T, db *sql.DB, id, tenantID, memberID, role, status string) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO platform_tenant_member (tenant_member_id, tenant_id, member_id, role, status) VALUES (?, ?, ?, ?, ?)", id, tenantID, memberID, role, status); err != nil {
		t.Fatalf("insert tenant_member error: %v", err)
	}
}

func requireTenant(t *testing.T, db *sql.DB, id, name, slug string) {
	t.Helper()
	var gotName string
	var gotSlug string
	err := db.QueryRow("SELECT name, slug FROM platform_tenant WHERE tenant_id = ?", id).Scan(&gotName, &gotSlug)
	requireNoError(t, err)
	requireEqual(t, gotName, name)
	requireEqual(t, gotSlug, slug)
}

func requireTenantDeleted(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	var cnt int
	err := db.QueryRow("SELECT COUNT(1) FROM platform_tenant WHERE tenant_id = ?", id).Scan(&cnt)
	requireNoError(t, err)
	requireEqual(t, cnt, 0)
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func requireEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got = %v, want = %v", got, want)
	}
}

func requireErrContains(t *testing.T, err error, expected string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("error = %v, want contains %q", err, expected)
	}
}
