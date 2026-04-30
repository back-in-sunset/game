package logic

import (
	"context"

	"im/rpc/im"
	"im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkReadLogic {
	return &MarkReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *MarkReadLogic) MarkRead(in *im.MarkReadRequest) (*im.ActionResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetUserID())
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.MessageStore.MarkRead(l.ctx, principal, in.GetPeerUserID(), in.GetSeq()); err != nil {
		return nil, err
	}

	return &im.ActionResponse{
		Success: true,
		Message: "ok",
	}, nil
}
