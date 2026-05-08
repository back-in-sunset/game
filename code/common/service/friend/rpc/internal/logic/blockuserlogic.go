package logic

import (
	"context"

	"friend/internal/errx"
	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BlockUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBlockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BlockUserLogic {
	return &BlockUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BlockUserLogic) BlockUser(in *friend.BlockUserReq) (*friend.BlockUserResp, error) {
	if in.UserId == in.BlockedUserId {
		return nil, errx.ToGRPC(errx.F1002)
	}

	domain := in.Domain
	if domain == "" {
		domain = "platform"
	}

	blocked, err := l.svcCtx.FriendStore.IsBlocked(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, errx.ToGRPC(errx.F1008)
	}

	if isFriend, _ := l.svcCtx.FriendStore.IsFriend(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId); isFriend {
		if err := l.svcCtx.FriendStore.RemoveFriend(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId); err != nil {
			return nil, err
		}
		_ = l.svcCtx.Cache.DelFriendCheck(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId)
		_ = l.svcCtx.Cache.DelFriendCheck(l.ctx, in.BlockedUserId, in.UserId, domain, in.TenantId)
	}

	if err := l.svcCtx.FriendStore.BlockUser(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.SetBlockedCheck(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId, true)

	return &friend.BlockUserResp{}, nil
}
