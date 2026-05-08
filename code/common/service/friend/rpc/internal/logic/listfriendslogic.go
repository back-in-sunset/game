package logic

import (
	"context"

	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFriendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFriendsLogic {
	return &ListFriendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListFriendsLogic) ListFriends(in *friend.ListFriendsReq) (*friend.ListFriendsResp, error) {
	domain := in.Domain
	if domain == "" {
		domain = "platform"
	}

	offset := int(in.Offset)
	limit := int(in.Limit)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	ids, ok := l.svcCtx.Cache.ListFriends(l.ctx, in.UserId, domain, in.TenantId)
	if ok && len(ids) >= offset {
		end := offset + limit
		if end > len(ids) {
			end = len(ids)
		}
		if offset < len(ids) {
			return &friend.ListFriendsResp{FriendIds: ids[offset:end], Total: int64(len(ids))}, nil
		}
	}

	friends, total, err := l.svcCtx.FriendStore.ListFriends(l.ctx, in.UserId, domain, in.TenantId, 0, 200)
	if err != nil {
		return nil, err
	}

	if len(friends) > 0 {
		_ = l.svcCtx.Cache.SetFriendList(l.ctx, in.UserId, domain, in.TenantId, friends)
	}

	if offset >= len(friends) {
		return &friend.ListFriendsResp{FriendIds: []int64{}, Total: total}, nil
	}
	end := offset + limit
	if end > len(friends) {
		end = len(friends)
	}

	return &friend.ListFriendsResp{FriendIds: friends[offset:end], Total: total}, nil
}
