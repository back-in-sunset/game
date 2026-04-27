package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PlatformProjectModel = (*customPlatformProjectModel)(nil)

type (
	// PlatformProjectModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPlatformProjectModel.
	PlatformProjectModel interface {
		platformProjectModel
		ListByTenantID(ctx context.Context, tenantID string) ([]*PlatformProject, error)
	}

	customPlatformProjectModel struct {
		*defaultPlatformProjectModel
	}
)

// NewPlatformProjectModel returns a model for the database table.
func NewPlatformProjectModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PlatformProjectModel {
	return &customPlatformProjectModel{
		defaultPlatformProjectModel: newPlatformProjectModel(conn, c, opts...),
	}
}

func (m *customPlatformProjectModel) ListByTenantID(ctx context.Context, tenantID string) ([]*PlatformProject, error) {
	var resp []*PlatformProject
	query := fmt.Sprintf("select %s from %s where `tenant_id` = ? order by `project_id` asc", platformProjectRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, tenantID)
	return resp, err
}
