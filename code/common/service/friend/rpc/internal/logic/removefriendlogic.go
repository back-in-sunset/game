package logic

import (
	"context"

	"friend/internal/errx"
	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RemoveFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRemoveFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveFriendLogic {
	return &RemoveFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RemoveFriendLogic) RemoveFriend(in *friend.RemoveFriendReq) (*friend.RemoveFriendResp, error) {
	domain := in.Domain
	if domain == "" {
		domain = "platform"
	}

	isFriend, err := l.svcCtx.FriendStore.IsFriend(l.ctx, in.UserId, in.FriendId, domain, in.TenantId)
	if err != nil {
		return nil, err
	}
	if !isFriend {
		return nil, errx.ToGRPC(errx.F1006)
	}

	if err := l.svcCtx.FriendStore.RemoveFriend(l.ctx, in.UserId, in.FriendId, domain, in.TenantId); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.DelFriendCheck(l.ctx, in.UserId, in.FriendId, domain, in.TenantId)
	_ = l.svcCtx.Cache.DelFriendCheck(l.ctx, in.FriendId, in.UserId, domain, in.TenantId)
	_ = l.svcCtx.Cache.RemoveFromList(l.ctx, in.UserId, in.FriendId, domain, in.TenantId)
	_ = l.svcCtx.Cache.RemoveFromList(l.ctx, in.FriendId, in.UserId, domain, in.TenantId)

	return &friend.RemoveFriendResp{}, nil
}
