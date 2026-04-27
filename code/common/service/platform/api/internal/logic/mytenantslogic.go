package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"platform/api/internal/svc"
	"platform/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MyTenantsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMyTenantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MyTenantsLogic {
	return &MyTenantsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MyTenantsLogic) MyTenants(req *types.MyTenantsReq) (resp *types.MyTenantsResp, err error) {
	memberID := resolveMemberIDFromContext(l.ctx)
	if memberID == "" {
		memberID = strings.TrimSpace(req.MemberId)
	}
	if memberID == "" {
		return nil, errors.New("x-uid is required")
	}

	tenants, err := l.svcCtx.Repo.ListTenantsByMemberID(l.ctx, memberID)
	if err != nil {
		return nil, err
	}

	items := make([]types.MyTenantItem, 0, len(tenants))
	for _, tenant := range tenants {
		items = append(items, types.MyTenantItem{
			Id:   tenant.ID,
			Name: tenant.Name,
			Slug: tenant.Slug,
		})
	}
	return &types.MyTenantsResp{Items: items}, nil
}

func resolveMemberIDFromContext(ctx context.Context) string {
	uid := ctx.Value(types.UserIDKey)
	switch v := uid.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}
