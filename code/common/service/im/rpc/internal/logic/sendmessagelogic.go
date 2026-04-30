package logic

import (
	"context"
	"time"

	"im/internal/domain"
	"im/rpc/im"
	"im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendMessageLogic) SendMessage(in *im.SendMessageRequest) (*im.SendMessageResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetSender())
	if err != nil {
		return nil, err
	}
	payload, err := payloadFromJSON(in.GetPayload().GetJson())
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.Router.Deliver(l.ctx, domain.Envelope{
		Domain:   principal.Domain,
		Scope:    principal.Scope,
		Sender:   principal.UserID,
		Receiver: in.GetReceiver(),
		MsgType:  in.GetMsgType(),
		Seq:      in.GetSeq(),
		Payload:  payload,
		SentAt:   ensureSentAt(time.Now()),
	})
	if err != nil {
		return nil, err
	}

	return &im.SendMessageResponse{
		Success:          true,
		OnlineRecipients: int64(result.OnlineRecipients),
		StoredOffline:    result.StoredOffline,
	}, nil
}
