package logic

import (
	"context"
	"encoding/json"

	"im/internal/domain"
	"im/rpc/im"
	"im/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeliverInternalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeliverInternalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeliverInternalLogic {
	return &DeliverInternalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeliverInternalLogic) DeliverInternal(in *im.DeliverInternalRequest) (*im.DeliverInternalResponse, error) {
	msg := in.GetMessage()
	if msg == nil {
		return &im.DeliverInternalResponse{}, nil
	}
	payload := map[string]any{}
	if msg.GetPayloadJson() != "" {
		if err := json.Unmarshal([]byte(msg.GetPayloadJson()), &payload); err != nil {
			return nil, err
		}
	}
	delivered, err := l.svcCtx.Router.DeliverLocal(l.ctx, domain.Envelope{
		Domain:   domain.IMDomain(msg.GetDomain()),
		Scope:    domain.Scope{TenantID: msg.GetScope().GetTenantID(), ProjectID: msg.GetScope().GetProjectID(), Environment: msg.GetScope().GetEnvironment()},
		Sender:   msg.GetSender(),
		Receiver: msg.GetReceiver(),
		MsgType:  msg.GetMsgType(),
		Seq:      msg.GetSeq(),
		Payload:  payload,
	})
	if err != nil {
		return nil, err
	}

	return &im.DeliverInternalResponse{Delivered: int64(delivered)}, nil
}
