package logic

import (
	"context"

	"im/rpc/im"
	"im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListConversationsLogic {
	return &ListConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListConversationsLogic) ListConversations(in *im.ListConversationsRequest) (*im.ListConversationsResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetUserID())
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.MessageStore.ListConversations(l.ctx, principal)
	if err != nil {
		return nil, err
	}

	resp := &im.ListConversationsResponse{
		Conversations: make([]*im.Conversation, 0, len(items)),
	}
	for _, item := range items {
		resp.Conversations = append(resp.Conversations, toRPCConversation(item))
	}
	return resp, nil
}
