package logic

import (
	"context"

	"im/rpc/im"
	"im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMessagesLogic {
	return &ListMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListMessagesLogic) ListMessages(in *im.ListMessagesRequest) (*im.ListMessagesResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetUserID())
	if err != nil {
		return nil, err
	}
	items, err := l.svcCtx.MessageStore.ListMessages(l.ctx, principal, in.GetPeerUserID(), int(in.GetLimit()))
	if err != nil {
		return nil, err
	}

	resp := &im.ListMessagesResponse{
		Messages: make([]*im.Message, 0, len(items)),
	}
	for _, item := range items {
		resp.Messages = append(resp.Messages, toRPCMessage(item))
	}
	return resp, nil
}
