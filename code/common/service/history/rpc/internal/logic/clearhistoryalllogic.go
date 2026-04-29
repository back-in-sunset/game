package logic

import (
	"context"

	"history/rpc/history"
	"history/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearHistoryAllLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewClearHistoryAllLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearHistoryAllLogic {
	return &ClearHistoryAllLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ClearHistoryAllLogic) ClearHistoryAll(in *history.ClearHistoryAllRequest) (*history.ActionResponse, error) {
	return NewDeleteHistoryLogic(l.ctx, l.svcCtx).ClearHistoryAll(in)
}
