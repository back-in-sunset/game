package logic

import (
	"context"

	"friend/rpc/friend"
	"friend/rpc/friendclient"
	"friend/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRequestsLogic {
	return &ListRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListRequestsLogic) ListRequests(in *friend.ListRequestsReq) (*friend.ListRequestsResp, error) {
	offset := int(in.Offset)
	limit := int(in.Limit)
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	reqs, total, err := l.svcCtx.FriendStore.ListIncomingRequests(l.ctx, in.UserId, in.Status, offset, limit)
	if err != nil {
		return nil, err
	}

	out := make([]*friendclient.FriendRequestInfo, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, &friendclient.FriendRequestInfo{
			RequestId:  r.ID,
			FromUserId: r.FromUserID,
			ToUserId:   r.ToUserID,
			Message:    r.Message,
			Status:     r.Status,
			CreatedAt:  r.CreatedAt,
		})
	}

	return &friend.ListRequestsResp{Requests: out, Total: total}, nil
}
