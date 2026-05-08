package logic

import (
	"context"

	"friend/internal/errx"
	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AcceptRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAcceptRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptRequestLogic {
	return &AcceptRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AcceptRequestLogic) AcceptRequest(in *friend.AcceptRequestReq) (*friend.AcceptRequestResp, error) {
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

	if err := l.svcCtx.FriendStore.AddFriend(l.ctx, req.FromUserID, req.ToUserID, req.Domain, req.TenantID); err != nil {
		return nil, err
	}
	if err := l.svcCtx.FriendStore.AddFriend(l.ctx, req.ToUserID, req.FromUserID, req.Domain, req.TenantID); err != nil {
		return nil, err
	}

	if err := l.svcCtx.FriendStore.UpdateRequestStatus(l.ctx, in.RequestId, 1); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.SetFriendCheck(l.ctx, req.FromUserID, req.ToUserID, req.Domain, req.TenantID, true)
	_ = l.svcCtx.Cache.SetFriendCheck(l.ctx, req.ToUserID, req.FromUserID, req.Domain, req.TenantID, true)

	return &friend.AcceptRequestResp{}, nil
}
