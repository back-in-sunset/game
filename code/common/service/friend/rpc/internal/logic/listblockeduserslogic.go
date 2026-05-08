package logic

import (
	"context"

	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBlockedUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListBlockedUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBlockedUsersLogic {
	return &ListBlockedUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListBlockedUsersLogic) ListBlockedUsers(in *friend.ListBlockedUsersReq) (*friend.ListBlockedUsersResp, error) {
	domain := in.Domain
	if domain == "" {
		domain = "platform"
	}

	offset := int(in.Offset)
	limit := int(in.Limit)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	blocked, total, err := l.svcCtx.FriendStore.ListBlockedUsers(l.ctx, in.UserId, domain, in.TenantId, offset, limit)
	if err != nil {
		return nil, err
	}

	return &friend.ListBlockedUsersResp{BlockedUserIds: blocked, Total: total}, nil
}
