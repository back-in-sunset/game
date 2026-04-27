package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ PlatformTenantMemberModel = (*customPlatformTenantMemberModel)(nil)

type (
	// PlatformTenantMemberModel is an interface to be customized, add more methods here,
	// and implement the added methods in customPlatformTenantMemberModel.
	PlatformTenantMemberModel interface {
		platformTenantMemberModel
		ListActiveByMemberID(ctx context.Context, memberID string) ([]*PlatformTenantMember, error)
	}

	customPlatformTenantMemberModel struct {
		*defaultPlatformTenantMemberModel
	}
)

// NewPlatformTenantMemberModel returns a model for the database table.
func NewPlatformTenantMemberModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) PlatformTenantMemberModel {
	return &customPlatformTenantMemberModel{
		defaultPlatformTenantMemberModel: newPlatformTenantMemberModel(conn, c, opts...),
	}
}

func (m *customPlatformTenantMemberModel) ListActiveByMemberID(ctx context.Context, memberID string) ([]*PlatformTenantMember, error) {
	var resp []*PlatformTenantMember
	query := fmt.Sprintf("select %s from %s where `member_id` = ? and `status` = 'active' order by `tenant_member_id` asc", platformTenantMemberRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, memberID)
	return resp, err
}
