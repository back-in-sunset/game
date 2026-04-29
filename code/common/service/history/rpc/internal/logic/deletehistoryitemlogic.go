package logic

import (
	"context"

	"history/rpc/history"
	"history/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteHistoryItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteHistoryItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteHistoryItemLogic {
	return &DeleteHistoryItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteHistoryItemLogic) DeleteHistoryItem(in *history.DeleteHistoryItemRequest) (*history.ActionResponse, error) {
	return NewDeleteHistoryLogic(l.ctx, l.svcCtx).DeleteHistoryItem(in)
}
