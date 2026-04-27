package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PlatformEnvironmentModel = (*customPlatformEnvironmentModel)(nil)

type (
	// PlatformEnvironmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPlatformEnvironmentModel.
	PlatformEnvironmentModel interface {
		platformEnvironmentModel
		ListByProjectID(ctx context.Context, projectID string) ([]*PlatformEnvironment, error)
	}

	customPlatformEnvironmentModel struct {
		*defaultPlatformEnvironmentModel
	}
)

// NewPlatformEnvironmentModel returns a model for the database table.
func NewPlatformEnvironmentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PlatformEnvironmentModel {
	return &customPlatformEnvironmentModel{
		defaultPlatformEnvironmentModel: newPlatformEnvironmentModel(conn, c, opts...),
	}
}

func (m *customPlatformEnvironmentModel) ListByProjectID(ctx context.Context, projectID string) ([]*PlatformEnvironment, error) {
	var resp []*PlatformEnvironment
	query := fmt.Sprintf("select %s from %s where `project_id` = ? order by `environment_id` asc", platformEnvironmentRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, projectID)
	return resp, err
}
