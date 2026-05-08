package logic

import (
	"context"

	"friend/internal/errx"
	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRejectRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectRequestLogic {
	return &RejectRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RejectRequestLogic) RejectRequest(in *friend.RejectRequestReq) (*friend.RejectRequestResp, error) {
	req, err := l.svcCtx.FriendStore.GetRequest(l.ctx, in.RequestId)
	if err != nil {
		return nil, err
	}
	if req.Status != 0 {
		return nil, errx.ToGRPC(errx.F1005)
	}
	if req.ToUserID != in.UserId {
		return nil, errx.ToGRPC(errx.F1004)
	}

	if err := l.svcCtx.FriendStore.UpdateRequestStatus(l.ctx, in.RequestId, 2); err != nil {
		return nil, err
	}

	return &friend.RejectRequestResp{}, nil
}
