package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PlatformTenantModel = (*customPlatformTenantModel)(nil)

type (
	// PlatformTenantModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPlatformTenantModel.
	PlatformTenantModel interface {
		platformTenantModel
	}

	customPlatformTenantModel struct {
		*defaultPlatformTenantModel
	}
)

// NewPlatformTenantModel returns a model for the database table.
func NewPlatformTenantModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PlatformTenantModel {
	return &customPlatformTenantModel{
		defaultPlatformTenantModel: newPlatformTenantModel(conn, c, opts...),
	}
}
