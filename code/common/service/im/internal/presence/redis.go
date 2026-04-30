package presence

import (
	"context"
	"fmt"

	"im/internal/auth"
	"im/internal/config"
	"im/internal/domain"

	"github.com/redis/go-redis/v9"
)

type Tracker struct {
	client    *redis.Client
	keyPrefix string
}

func NewTracker(cfg config.Redis) *Tracker {
	return &Tracker{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		keyPrefix: cfg.KeyPrefix,
	}
}

func (t *Tracker) Bind(ctx context.Context, principal auth.Principal, nodeID string) error {
	return t.client.Set(ctx, t.key(principal), nodeID, 0).Err()
}

func (t *Tracker) Unbind(ctx context.Context, principal auth.Principal, nodeID string) error {
	key := t.key(principal)
	current, err := t.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	if current != nodeID {
		return nil
	}
	return t.client.Del(ctx, key).Err()
}

func (t *Tracker) Lookup(ctx context.Context, principal auth.Principal) (string, bool, error) {
	value, err := t.client.Get(ctx, t.key(principal)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (t *Tracker) Close() error {
	return t.client.Close()
}

func (t *Tracker) key(principal auth.Principal) string {
	scope := principal.Scope.Normalize()
	if principal.Domain == domain.DomainPlatform {
		return fmt.Sprintf("%s:presence:platform:%d", t.keyPrefix, principal.UserID)
	}
	return fmt.Sprintf("%s:presence:tenant:%s:%s:%s:%d", t.keyPrefix, scope.TenantID, scope.ProjectID, scope.Environment, principal.UserID)
}
