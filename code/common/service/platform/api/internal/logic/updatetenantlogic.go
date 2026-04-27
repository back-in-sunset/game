package logic

import (
	"context"
	"errors"
	"strings"

	"platform/api/internal/svc"
	"platform/api/internal/types"
	"platform/internal/domain"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantLogic) UpdateTenant(req *types.UpdateTenantReq) (resp *types.UpdateTenantResp, err error) {
	req.TenantId = strings.TrimSpace(req.TenantId)
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	if req.TenantId == "" || req.Name == "" || req.Slug == "" {
		return nil, errors.New("tenantId, name and slug are required")
	}

	tenant, err := domain.NewTenant(req.TenantId, req.Name, req.Slug)
	if err != nil {
		return nil, err
	}

	if err = l.svcCtx.Repo.UpdateTenant(l.ctx, tenant); err != nil {
		return nil, err
	}

	return &types.UpdateTenantResp{
		Id:   tenant.ID,
		Name: tenant.Name,
		Slug: tenant.Slug,
	}, nil
}
