// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"strings"

	"platform/api/internal/svc"
	"platform/api/internal/types"
	"platform/internal/domain"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantLogic {
	return &GetTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantLogic) GetTenant(req *types.GetTenantReq) (resp *types.GetTenantResp, err error) {
	req.TenantId = strings.TrimSpace(req.TenantId)
	tenant, err := l.svcCtx.Repo.GetTenantByID(l.ctx, req.TenantId)
	if err != nil {
		if err == domain.ErrTenantNotFound {
			return nil, err
		}
		return nil, err
	}

	return &types.GetTenantResp{
		Id:   tenant.ID,
		Name: tenant.Name,
		Slug: tenant.Slug,
	}, nil
}
