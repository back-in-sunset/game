package logic

import (
	"context"
	"errors"
	"strings"

	"platform/api/internal/svc"
	"platform/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantLogic {
	return &DeleteTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantLogic) DeleteTenant(req *types.DeleteTenantReq) (resp *types.DeleteTenantResp, err error) {
	req.TenantId = strings.TrimSpace(req.TenantId)
	if req.TenantId == "" {
		return nil, errors.New("tenantId is required")
	}

	if err = l.svcCtx.Repo.DeleteTenant(l.ctx, req.TenantId); err != nil {
		return nil, err
	}

	return &types.DeleteTenantResp{Success: true}, nil
}
