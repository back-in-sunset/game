// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package logic

import (
	"context"
	"strings"

	"platform/api/internal/svc"
	"platform/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListEnvironmentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListEnvironmentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListEnvironmentsLogic {
	return &ListEnvironmentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListEnvironmentsLogic) ListEnvironments(req *types.ListEnvironmentsReq) (resp *types.ListEnvironmentsResp, err error) {
	req.ProjectId = strings.TrimSpace(req.ProjectId)
	environments, err := l.svcCtx.Repo.ListEnvironmentsByProjectID(l.ctx, req.ProjectId)
	if err != nil {
		return nil, err
	}

	items := make([]types.EnvironmentItem, 0, len(environments))
	for _, environment := range environments {
		items = append(items, types.EnvironmentItem{
			Id:          environment.ID,
			ProjectId:   environment.ProjectID,
			Name:        environment.Name,
			DisplayName: environment.DisplayName,
		})
	}
	return &types.ListEnvironmentsResp{Items: items}, nil
}
