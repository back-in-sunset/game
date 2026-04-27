// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

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

type CreateProjectLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateProjectLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProjectLogic {
	return &CreateProjectLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateProjectLogic) CreateProject(req *types.CreateProjectReq) (resp *types.CreateProjectResp, err error) {
	req.TenantId = strings.TrimSpace(req.TenantId)
	req.Name = strings.TrimSpace(req.Name)
	req.Key = strings.TrimSpace(req.Key)
	if req.TenantId == "" || req.Name == "" || req.Key == "" {
		return nil, errors.New("tenantId, name and key are required")
	}

	project, err := domain.NewProject(newID(), req.TenantId, req.Name, req.Key)
	if err != nil {
		return nil, err
	}
	if err = l.svcCtx.Repo.SaveProject(l.ctx, project); err != nil {
		return nil, err
	}

	return &types.CreateProjectResp{
		Id:       project.ID,
		TenantId: project.TenantID,
		Name:     project.Name,
		Key:      project.Key,
	}, nil
}
