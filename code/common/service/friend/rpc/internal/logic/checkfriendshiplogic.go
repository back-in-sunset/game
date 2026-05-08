package logic

import (
	"context"

	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFriendshipLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckFriendshipLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFriendshipLogic {
	return &CheckFriendshipLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckFriendshipLogic) CheckFriendship(in *friend.CheckFriendshipReq) (*friend.CheckFriendshipResp, error) {
	domain := in.Domain
	if domain == "" {
		domain = "platform"
	}

	isFriend, ok := l.svcCtx.Cache.IsFriend(l.ctx, in.UserId, in.TargetId, domain, in.TenantId)
	if ok {
		return &friend.CheckFriendshipResp{IsFriend: isFriend}, nil
	}

	isFriend, err := l.svcCtx.FriendStore.IsFriend(l.ctx, in.UserId, in.TargetId, domain, in.TenantId)
	if err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.SetFriendCheck(l.ctx, in.UserId, in.TargetId, domain, in.TenantId, isFriend)

	return &friend.CheckFriendshipResp{IsFriend: isFriend}, nil
}
