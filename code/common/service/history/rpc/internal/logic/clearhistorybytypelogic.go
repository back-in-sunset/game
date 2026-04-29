package logic

import (
	"context"

	"history/rpc/history"
	"history/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearHistoryByTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearHistoryByTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearHistoryByTypeLogic {
	return &ClearHistoryByTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearHistoryByTypeLogic) ClearHistoryByType(in *history.ClearHistoryByTypeRequest) (*history.ActionResponse, error) {
	return NewDeleteHistoryLogic(l.ctx, l.svcCtx).ClearHistoryByType(in)
}
