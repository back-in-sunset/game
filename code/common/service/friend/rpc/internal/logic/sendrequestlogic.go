package logic

import (
	"context"

	"friend/internal/errx"
	"friend/internal/repository"
	"friend/rpc/friend"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendRequestLogic {
	return &SendRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendRequestLogic) SendRequest(in *friend.SendRequestReq) (*friend.SendRequestResp, error) {
	if in.FromUserId == in.ToUserId {
		return nil, errx.ToGRPC(errx.F1002)
	}

	domain := in.Domain
	if domain == "" {
		domain = "platform"
	}

	if isFriend, _ := l.svcCtx.FriendStore.IsFriend(l.ctx, in.FromUserId, in.ToUserId, domain, in.TenantId); isFriend {
		return nil, errx.ToGRPC(errx.F1003)
	}

	if blocked, _ := l.svcCtx.FriendStore.IsBlocked(l.ctx, in.ToUserId, in.FromUserId, domain, in.TenantId); blocked {
		return nil, errx.ToGRPC(errx.F1007)
	}

	reqs, _, err := l.svcCtx.FriendStore.ListIncomingRequests(l.ctx, in.ToUserId, 0, 0, 1)
	if err != nil {
		return nil, err
	}
	for _, r := range reqs {
		if r.FromUserID == in.FromUserId {
			return nil, errx.ToGRPC(errx.F1009)
		}
	}

	req := &repository.FriendRequest{
		FromUserID: in.FromUserId,
		ToUserID:   in.ToUserId,
		Domain:     domain,
		TenantID:   in.TenantId,
		Message:    in.Message,
	}
	requestID, err := l.svcCtx.FriendStore.CreateRequest(l.ctx, req)
	if err != nil {
		return nil, err
	}

	return &friend.SendRequestResp{RequestId: requestID}, nil
}
