package integration

import (
	"database/sql"
	"strings"
	"testing"

	"game/server/core/testkit"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"platform/model"
)

func TestMySQLRepositoryListTenantsByMemberID(t *testing.T) {
	if testing.Short() {
		t.Skip("skip integration test in short mode")
	}

	ctx, dsn := testkit.StartMySQLContainer(t, "platform")
	db := testkit.OpenMySQLWithRetry(t, ctx, dsn)
	mustExecSchema(t, db)

	mustExec(t, db, "INSERT INTO platform_tenant (tenant_id, name, slug, status) VALUES (?, ?, ?, ?)",
		"tenant-1", "Acme Games", "acme-games", "active")
	mustExec(t, db, "INSERT INTO platform_tenant (tenant_id, name, slug, status) VALUES (?, ?, ?, ?)",
		"tenant-2", "Beta Games", "beta-games", "active")
	mustExec(t, db, "INSERT INTO platform_tenant_member (tenant_member_id, tenant_id, member_id, role, status) VALUES (?, ?, ?, ?, ?)",
		"tm-1", "tenant-1", "u-1", "owner", "active")
	mustExec(t, db, "INSERT INTO platform_tenant_member (tenant_member_id, tenant_id, member_id, role, status) VALUES (?, ?, ?, ?, ?)",
		"tm-2", "tenant-2", "u-1", "developer", "inactive")

	redisAddr := testkit.StartMiniRedis(t)
	repo := model.NewMySQLRepository(sqlx.NewMysql(dsn), cache.CacheConf{
		{
			RedisConf: redis.RedisConf{
				Host: redisAddr,
				Type: "node",
			},
		},
	})

	tenants, err := repo.ListTenantsByMemberID(ctx, "u-1")
	if err != nil {
		t.Fatalf("ListTenantsByMemberID() error = %v", err)
	}
	if len(tenants) != 1 {
		t.Fatalf("len(tenants) = %d, want %d", len(tenants), 1)
	}
	if tenants[0].ID != "tenant-1" {
		t.Fatalf("tenants[0].ID = %q, want %q", tenants[0].ID, "tenant-1")
	}
}

func mustExecSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	const ddl = `
CREATE TABLE platform_tenant (
  tenant_id varchar(64) NOT NULL,
  name varchar(128) NOT NULL,
  slug varchar(64) NOT NULL,
  status varchar(32) NOT NULL DEFAULT 'active',
  created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (tenant_id),
  UNIQUE KEY uk_platform_tenant_slug (slug)
);
CREATE TABLE platform_tenant_member (
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
);
`

	for _, stmt := range strings.Split(ddl, ";") {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("exec schema statement error: %v, stmt=%s", err, stmt)
		}
	}
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("db.Exec() error = %v, query=%s", err, query)
	}
}
