package notifier

import (
	"context"
	"time"

	"im/internal/contracts"
	"im/internal/domain"
)

type Notifier struct {
	router contracts.Router
}

func New(router contracts.Router) *Notifier {
	return &Notifier{router: router}
}

func (n *Notifier) Notify(ctx context.Context, envelope domain.Envelope) error {
	if envelope.SentAt.IsZero() {
		envelope.SentAt = time.Now().UTC()
	}
	_, err := n.router.Deliver(ctx, envelope)
	return err
}
