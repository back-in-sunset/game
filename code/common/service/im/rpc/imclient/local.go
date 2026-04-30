package imclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"im/internal/auth"
	"im/internal/contracts"
	"im/internal/domain"

	"google.golang.org/grpc"
)

type localIM struct {
	store  contracts.MessageStore
	router contracts.Router
}

func NewLocalIM(store contracts.MessageStore, router contracts.Router) IM {
	return &localIM{
		store:  store,
		router: router,
	}
}

func (l *localIM) SendMessage(ctx context.Context, in *SendMessageRequest, _ ...grpc.CallOption) (*SendMessageResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetSender())
	if err != nil {
		return nil, err
	}
	payload, err := payloadFromJSON(in.GetPayload().GetJson())
	if err != nil {
		return nil, err
	}
	result, err := l.router.Deliver(ctx, domain.Envelope{
		Domain:   principal.Domain,
		Scope:    principal.Scope,
		Sender:   principal.UserID,
		Receiver: in.GetReceiver(),
		MsgType:  in.GetMsgType(),
		Seq:      in.GetSeq(),
		Payload:  payload,
		SentAt:   time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}
	return &SendMessageResponse{
		Success:          true,
		OnlineRecipients: int64(result.OnlineRecipients),
		StoredOffline:    result.StoredOffline,
	}, nil
}

func (l *localIM) ListConversations(ctx context.Context, in *ListConversationsRequest, _ ...grpc.CallOption) (*ListConversationsResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetUserID())
	if err != nil {
		return nil, err
	}
	items, err := l.store.ListConversations(ctx, principal)
	if err != nil {
		return nil, err
	}
	resp := &ListConversationsResponse{Conversations: make([]*Conversation, 0, len(items))}
	for _, item := range items {
		resp.Conversations = append(resp.Conversations, toRPCConversation(item))
	}
	return resp, nil
}

func (l *localIM) ListMessages(ctx context.Context, in *ListMessagesRequest, _ ...grpc.CallOption) (*ListMessagesResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetUserID())
	if err != nil {
		return nil, err
	}
	items, err := l.store.ListMessages(ctx, principal, in.GetPeerUserID(), int(in.GetLimit()))
	if err != nil {
		return nil, err
	}
	resp := &ListMessagesResponse{Messages: make([]*Message, 0, len(items))}
	for _, item := range items {
		resp.Messages = append(resp.Messages, toRPCMessage(item))
	}
	return resp, nil
}

func (l *localIM) MarkRead(ctx context.Context, in *MarkReadRequest, _ ...grpc.CallOption) (*ActionResponse, error) {
	principal, err := principalFromRequest(in.GetDomain(), in.GetScope(), in.GetUserID())
	if err != nil {
		return nil, err
	}
	if err := l.store.MarkRead(ctx, principal, in.GetPeerUserID(), in.GetSeq()); err != nil {
		return nil, err
	}
	return &ActionResponse{Success: true, Message: "ok"}, nil
}

func (l *localIM) DeliverInternal(ctx context.Context, in *DeliverInternalRequest, _ ...grpc.CallOption) (*DeliverInternalResponse, error) {
	msg := in.GetMessage()
	if msg == nil {
		return &DeliverInternalResponse{}, nil
	}
	payload, err := payloadFromJSON(msg.GetPayloadJson())
	if err != nil {
		return nil, err
	}
	delivered, err := l.router.DeliverLocal(ctx, domain.Envelope{
		Domain:   domain.IMDomain(msg.GetDomain()),
		Scope:    scopeFromRPC(msg.GetScope()),
		Sender:   msg.GetSender(),
		Receiver: msg.GetReceiver(),
		MsgType:  msg.GetMsgType(),
		Seq:      msg.GetSeq(),
		Payload:  payload,
		SentAt:   time.Unix(msg.GetSentAtUnix(), 0).UTC(),
	})
	if err != nil {
		return nil, err
	}
	return &DeliverInternalResponse{Delivered: int64(delivered)}, nil
}

func principalFromRequest(imDomain string, rpcScope *Scope, userID int64) (auth.Principal, error) {
	scope := scopeFromRPC(rpcScope)
	principal := auth.Principal{
		UserID: userID,
		Domain: domain.IMDomain(imDomain),
		Scope:  scope,
	}
	if userID <= 0 {
		return auth.Principal{}, fmt.Errorf("user_id is required")
	}
	if err := scope.Validate(principal.Domain); err != nil {
		return auth.Principal{}, err
	}
	return principal, nil
}

func scopeFromRPC(in *Scope) domain.Scope {
	if in == nil {
		return domain.Scope{}
	}
	return domain.Scope{
		TenantID:    in.GetTenantID(),
		ProjectID:   in.GetProjectID(),
		Environment: in.GetEnvironment(),
	}
}

func payloadFromJSON(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("invalid payload json: %w", err)
	}
	return out, nil
}

func payloadToJSON(payload map[string]any) string {
	if len(payload) == 0 {
		return "{}"
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func toRPCConversation(in domain.Conversation) *Conversation {
	return &Conversation{
		Key:             in.Key,
		Domain:          string(in.Domain),
		Scope:           &Scope{TenantID: in.Scope.TenantID, ProjectID: in.Scope.ProjectID, Environment: in.Scope.Environment},
		PeerUserID:      in.PeerUserID,
		UnreadCount:     in.UnreadCount,
		ReadSeq:         in.ReadSeq,
		UpdatedAtUnix:   in.UpdatedAt.Unix(),
		LastSender:      in.LastMessage.Sender,
		LastReceiver:    in.LastMessage.Receiver,
		LastMsgType:     in.LastMessage.MsgType,
		LastSeq:         in.LastMessage.Seq,
		LastPayloadJson: payloadToJSON(in.LastMessage.Payload),
		LastSentAtUnix:  in.LastMessage.SentAt.Unix(),
	}
}

func toRPCMessage(in domain.Envelope) *Message {
	return &Message{
		Domain:      string(in.Domain),
		Scope:       &Scope{TenantID: in.Scope.TenantID, ProjectID: in.Scope.ProjectID, Environment: in.Scope.Environment},
		Sender:      in.Sender,
		Receiver:    in.Receiver,
		MsgType:     in.MsgType,
		Seq:         in.Seq,
		PayloadJson: payloadToJSON(in.Payload),
		SentAtUnix:  in.SentAt.Unix(),
	}
}
