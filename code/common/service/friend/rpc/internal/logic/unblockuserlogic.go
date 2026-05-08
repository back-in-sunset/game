package logic

import (
	"context"

	"friend/internal/errx"
	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnblockUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnblockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnblockUserLogic {
	return &UnblockUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnblockUserLogic) UnblockUser(in *friend.UnblockUserReq) (*friend.UnblockUserResp, error) {
	domain := in.Domain
	if domain == "" {
		domain = "platform"
	}

	blocked, err := l.svcCtx.FriendStore.IsBlocked(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId)
	if err != nil {
		return nil, err
	}
	if !blocked {
		return nil, errx.ToGRPC(errx.F1010)
	}

	if err := l.svcCtx.FriendStore.UnblockUser(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId); err != nil {
		return nil, err
	}

	_ = l.svcCtx.Cache.DelBlockedCheck(l.ctx, in.UserId, in.BlockedUserId, domain, in.TenantId)

	return &friend.UnblockUserResp{}, nil
}
