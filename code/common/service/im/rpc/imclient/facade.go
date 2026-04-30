package imclient

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	MsgTypeDirectMessage = "direct_message"
	MsgTypeSystemNotice  = "system_notice"
	MsgTypeBizPush       = "biz_push"
)

type Facade struct {
	client IM
}

type ScopeRef struct {
	Domain      string
	TenantID    string
	ProjectID   string
	Environment string
}

type DirectMessage struct {
	Scope    ScopeRef
	Sender   int64
	Receiver int64
	Seq      int64
	Payload  map[string]any
}

type SystemNotice struct {
	Scope    ScopeRef
	Receiver int64
	Seq      int64
	Payload  map[string]any
}

type BizPush struct {
	Scope    ScopeRef
	Receiver int64
	Seq      int64
	Payload  map[string]any
}

func NewFacade(client IM) *Facade {
	return &Facade{client: client}
}

func (f *Facade) SendDirectMessage(ctx context.Context, in DirectMessage) (*SendMessageResponse, error) {
	if in.Sender <= 0 {
		return nil, fmt.Errorf("sender is required")
	}
	return f.send(ctx, sendEnvelope{
		scope:    in.Scope,
		sender:   in.Sender,
		receiver: in.Receiver,
		seq:      in.Seq,
		msgType:  MsgTypeDirectMessage,
		payload:  in.Payload,
	})
}

func (f *Facade) SendSystemNotice(ctx context.Context, in SystemNotice) (*SendMessageResponse, error) {
	return f.send(ctx, sendEnvelope{
		scope:    in.Scope,
		sender:   0,
		receiver: in.Receiver,
		seq:      in.Seq,
		msgType:  MsgTypeSystemNotice,
		payload:  in.Payload,
	})
}

func (f *Facade) SendBizPush(ctx context.Context, in BizPush) (*SendMessageResponse, error) {
	return f.send(ctx, sendEnvelope{
		scope:    in.Scope,
		sender:   0,
		receiver: in.Receiver,
		seq:      in.Seq,
		msgType:  MsgTypeBizPush,
		payload:  in.Payload,
	})
}

func (f *Facade) send(ctx context.Context, in sendEnvelope) (*SendMessageResponse, error) {
	if f.client == nil {
		return nil, fmt.Errorf("client is required")
	}
	payloadJSON, err := json.Marshal(in.payload)
	if err != nil {
		return nil, err
	}
	return f.client.SendMessage(ctx, &SendMessageRequest{
		Domain:   in.scope.Domain,
		Scope:    toScope(in.scope),
		Sender:   in.sender,
		Receiver: in.receiver,
		MsgType:  in.msgType,
		Seq:      in.seq,
		Payload:  &MessagePayload{Json: string(payloadJSON)},
	})
}

func toScope(in ScopeRef) *Scope {
	return &Scope{
		TenantID:    in.TenantID,
		ProjectID:   in.ProjectID,
		Environment: in.Environment,
	}
}

type sendEnvelope struct {
	scope    ScopeRef
	sender   int64
	receiver int64
	seq      int64
	msgType  string
	payload  map[string]any
}
