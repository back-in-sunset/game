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

type CreateEnvironmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateEnvironmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateEnvironmentLogic {
	return &CreateEnvironmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateEnvironmentLogic) CreateEnvironment(req *types.CreateEnvironmentReq) (resp *types.CreateEnvironmentResp, err error) {
	req.ProjectId = strings.TrimSpace(req.ProjectId)
	req.Name = strings.TrimSpace(req.Name)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.ProjectId == "" || req.Name == "" {
		return nil, errors.New("projectId and name are required")
	}

	environment, err := domain.NewEnvironment(newID(), req.ProjectId, req.Name, req.DisplayName)
	if err != nil {
		return nil, err
	}
	if err = l.svcCtx.Repo.SaveEnvironment(l.ctx, environment); err != nil {
		return nil, err
	}

	return &types.CreateEnvironmentResp{
		Id:          environment.ID,
		ProjectId:   environment.ProjectID,
		Name:        environment.Name,
		DisplayName: environment.DisplayName,
	}, nil
}
