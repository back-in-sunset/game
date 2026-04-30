package store

import (
	"context"

	"im/internal/auth"
	"im/internal/domain"
)

type MessageArchive interface {
	AppendMessage(ctx context.Context, envelope domain.Envelope) (StoredMessage, error)
	ListMessages(ctx context.Context, principal auth.Principal, peerUserID int64, limit int) ([]domain.Envelope, error)
}

type SessionStateStore interface {
	UpsertConversationPair(ctx context.Context, stored StoredMessage) error
	SaveOffline(ctx context.Context, principal auth.Principal, envelope domain.Envelope) error
	DrainOffline(ctx context.Context, principal auth.Principal) ([]domain.Envelope, error)
	ListConversations(ctx context.Context, principal auth.Principal) ([]domain.Conversation, error)
	MarkRead(ctx context.Context, principal auth.Principal, peerUserID int64, seq int64) error
}

type StoredMessage struct {
	ID              int64
	ConversationKey string
	Envelope        domain.Envelope
}
