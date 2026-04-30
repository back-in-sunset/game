package pipeline

import (
	"context"
	"encoding/json"
	"time"

	"im/internal/auth"
	"im/internal/contracts"
	"im/internal/domain"
)

type MessageInput struct {
	Receiver int64          `json:"receiver"`
	MsgType  string         `json:"msg_type"`
	Seq      int64          `json:"seq"`
	Payload  map[string]any `json:"payload"`
}

type QueryInput struct {
	PeerUserID int64 `json:"peer_user_id"`
	Limit      int   `json:"limit"`
	Seq        int64 `json:"seq"`
}

type Messaging struct {
	router contracts.Router
	store  contracts.MessageStore
}

func NewMessaging(router contracts.Router, store contracts.MessageStore) *Messaging {
	return &Messaging{router: router, store: store}
}

func (m *Messaging) HandleSend(ctx context.Context, principal auth.Principal, input MessageInput) ([]byte, error) {
	envelope := domain.Envelope{
		Domain:   principal.Domain,
		Scope:    principal.Scope,
		Sender:   principal.UserID,
		Receiver: input.Receiver,
		MsgType:  input.MsgType,
		Seq:      input.Seq,
		Payload:  input.Payload,
		SentAt:   time.Now().UTC(),
	}

	result, err := m.router.Deliver(ctx, envelope)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]any{
		"type":              "send_ack",
		"seq":               envelope.Seq,
		"online_recipients": result.OnlineRecipients,
		"stored_offline":    result.StoredOffline,
	})
}

func (m *Messaging) DrainOffline(ctx context.Context, principal auth.Principal) ([]byte, error) {
	items, err := m.store.DrainOffline(ctx, principal)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":     "offline_batch",
		"messages": items,
	})
}

func (m *Messaging) HandleListConversations(ctx context.Context, principal auth.Principal) ([]byte, error) {
	items, err := m.store.ListConversations(ctx, principal)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":          "conversation_list",
		"conversations": items,
	})
}

func (m *Messaging) HandleListMessages(ctx context.Context, principal auth.Principal, input QueryInput) ([]byte, error) {
	items, err := m.store.ListMessages(ctx, principal, input.PeerUserID, input.Limit)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":     "message_list",
		"messages": items,
	})
}

func (m *Messaging) HandleMarkRead(ctx context.Context, principal auth.Principal, input QueryInput) ([]byte, error) {
	if err := m.store.MarkRead(ctx, principal, input.PeerUserID, input.Seq); err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":         "read_ack",
		"peer_user_id": input.PeerUserID,
		"seq":          input.Seq,
	})
}
