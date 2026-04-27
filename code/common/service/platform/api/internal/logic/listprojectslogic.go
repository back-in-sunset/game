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

type ListProjectsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListProjectsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProjectsLogic {
	return &ListProjectsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListProjectsLogic) ListProjects(req *types.ListProjectsReq) (resp *types.ListProjectsResp, err error) {
	req.TenantId = strings.TrimSpace(req.TenantId)
	projects, err := l.svcCtx.Repo.ListProjectsByTenantID(l.ctx, req.TenantId)
	if err != nil {
		return nil, err
	}

	items := make([]types.ProjectItem, 0, len(projects))
	for _, project := range projects {
		items = append(items, types.ProjectItem{
			Id:       project.ID,
			TenantId: project.TenantID,
			Name:     project.Name,
			Key:      project.Key,
		})
	}
	return &types.ListProjectsResp{Items: items}, nil
}
