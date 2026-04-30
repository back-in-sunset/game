package router

import (
	"context"
	"encoding/json"

	"im/internal/auth"
	"im/internal/contracts"
	"im/internal/domain"
)

type LocalRouter struct {
	localNodeID string
	sessions    contracts.SessionManager
	store       contracts.MessageStore
	registry    contracts.Registry
	presence    interface {
		Lookup(ctx context.Context, principal auth.Principal) (string, bool, error)
	}
	forwarder Forwarder
}

func NewLocalRouter(
	localNodeID string,
	sessions contracts.SessionManager,
	store contracts.MessageStore,
	registry contracts.Registry,
	presence interface {
		Lookup(ctx context.Context, principal auth.Principal) (string, bool, error)
	},
	forwarder Forwarder,
) *LocalRouter {
	return &LocalRouter{
		localNodeID: localNodeID,
		sessions:    sessions,
		store:       store,
		registry:    registry,
		presence:    presence,
		forwarder:   forwarder,
	}
}

func (r *LocalRouter) Deliver(ctx context.Context, envelope domain.Envelope) (contracts.DeliveryResult, error) {
	if err := envelope.Validate(); err != nil {
		return contracts.DeliveryResult{}, err
	}
	if err := r.store.SaveMessage(ctx, envelope); err != nil {
		return contracts.DeliveryResult{}, err
	}

	receiver := auth.Principal{
		UserID: envelope.Receiver,
		Domain: envelope.Domain,
		Scope:  envelope.Scope,
	}

	sent, err := r.DeliverLocal(ctx, envelope)
	if err != nil {
		return contracts.DeliveryResult{}, err
	}
	if sent > 0 {
		return contracts.DeliveryResult{OnlineRecipients: sent}, nil
	}
	if r.presence != nil && r.registry != nil && r.forwarder != nil {
		if nodeID, ok, err := r.presence.Lookup(ctx, receiver); err != nil {
			return contracts.DeliveryResult{}, err
		} else if ok && nodeID != "" && nodeID != r.localNodeID {
			node, err := r.registry.GetNode(ctx, nodeID)
			if err == nil && node.RPC != "" {
				delivered, err := r.forwarder.Forward(ctx, node.RPC, envelope)
				if err == nil && delivered > 0 {
					return contracts.DeliveryResult{OnlineRecipients: delivered}, nil
				}
			}
		}
	}
	if err := r.store.SaveOffline(ctx, receiver, envelope); err != nil {
		return contracts.DeliveryResult{}, err
	}
	return contracts.DeliveryResult{StoredOffline: true}, nil
}

func (r *LocalRouter) DeliverLocal(ctx context.Context, envelope domain.Envelope) (int, error) {
	payload, err := json.Marshal(map[string]any{
		"type":     "message",
		"envelope": envelope,
	})
	if err != nil {
		return 0, err
	}
	receiver := auth.Principal{
		UserID: envelope.Receiver,
		Domain: envelope.Domain,
		Scope:  envelope.Scope,
	}
	return r.sessions.SendToPrincipal(ctx, receiver, payload)
}
