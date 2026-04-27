// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"errors"
	"strings"
	"time"

	"platform/api/internal/svc"
	"platform/api/internal/types"
	"platform/internal/domain"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantLogic) CreateTenant(req *types.CreateTenantReq) (resp *types.CreateTenantResp, err error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Slug = strings.TrimSpace(req.Slug)
	if req.Name == "" || req.Slug == "" {
		return nil, errors.New("name and slug are required")
	}

	tenant, err := domain.NewTenant(newID(), req.Name, req.Slug)
	if err != nil {
		return nil, err
	}
	if err = l.svcCtx.Repo.SaveTenant(l.ctx, tenant); err != nil {
		return nil, err
	}

	return &types.CreateTenantResp{
		Id:   tenant.ID,
		Name: tenant.Name,
		Slug: tenant.Slug,
	}, nil
}

func newID() string {
	return time.Now().UTC().Format("20060102150405.000000000")
}
