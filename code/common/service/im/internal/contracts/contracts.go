package contracts

import (
	"context"
	"time"

	"im/internal/auth"
	"im/internal/domain"
)

type AuthProvider interface {
	Authenticate(token string, imDomain domain.IMDomain, scope domain.Scope) (auth.Principal, error)
}

type Registry interface {
	Register(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (Node, error)
	Close() error
}

type Router interface {
	Deliver(ctx context.Context, envelope domain.Envelope) (DeliveryResult, error)
	DeliverLocal(ctx context.Context, envelope domain.Envelope) (int, error)
}

type SessionManager interface {
	Bind(principal auth.Principal, conn Connection)
	Unbind(principal auth.Principal, connID string)
	SendToPrincipal(ctx context.Context, principal auth.Principal, payload []byte) (int, error)
}

type MessageStore interface {
	SaveMessage(ctx context.Context, envelope domain.Envelope) error
	SaveOffline(ctx context.Context, principal auth.Principal, envelope domain.Envelope) error
	DrainOffline(ctx context.Context, principal auth.Principal) ([]domain.Envelope, error)
	ListConversations(ctx context.Context, principal auth.Principal) ([]domain.Conversation, error)
	ListMessages(ctx context.Context, principal auth.Principal, peerUserID int64, limit int) ([]domain.Envelope, error)
	MarkRead(ctx context.Context, principal auth.Principal, peerUserID int64, seq int64) error
}

type Notifier interface {
	Notify(ctx context.Context, envelope domain.Envelope) error
}

type ScopeResolver interface {
	Resolve(imDomain domain.IMDomain, scope domain.Scope) (domain.Scope, error)
}

type Connection interface {
	ID() string
	Send(ctx context.Context, payload []byte) error
	Close() error
}

type Node struct {
	ID        string
	Service   string
	WebSocket string
	TCP       string
	RPC       string
	Domains   []string
	SeenAt    time.Time
}

type DeliveryResult struct {
	OnlineRecipients int
	StoredOffline    bool
}
