package store

import (
	"context"
	"errors"

	"im/internal/auth"
	"im/internal/domain"
)

type CompositeStore struct {
	archive MessageArchive
	state   SessionStateStore
}

type closer interface {
	Close() error
}

func NewCompositeStore(archive MessageArchive, state SessionStateStore) *CompositeStore {
	return &CompositeStore{
		archive: archive,
		state:   state,
	}
}

func (s *CompositeStore) SaveMessage(ctx context.Context, envelope domain.Envelope) error {
	stored, err := s.archive.AppendMessage(ctx, envelope)
	if err != nil {
		return err
	}
	return s.state.UpsertConversationPair(ctx, stored)
}

func (s *CompositeStore) SaveOffline(ctx context.Context, principal auth.Principal, envelope domain.Envelope) error {
	return s.state.SaveOffline(ctx, principal, envelope)
}

func (s *CompositeStore) DrainOffline(ctx context.Context, principal auth.Principal) ([]domain.Envelope, error) {
	return s.state.DrainOffline(ctx, principal)
}

func (s *CompositeStore) ListConversations(ctx context.Context, principal auth.Principal) ([]domain.Conversation, error) {
	return s.state.ListConversations(ctx, principal)
}

func (s *CompositeStore) ListMessages(ctx context.Context, principal auth.Principal, peerUserID int64, limit int) ([]domain.Envelope, error) {
	return s.archive.ListMessages(ctx, principal, peerUserID, limit)
}

func (s *CompositeStore) MarkRead(ctx context.Context, principal auth.Principal, peerUserID int64, seq int64) error {
	return s.state.MarkRead(ctx, principal, peerUserID, seq)
}

func (s *CompositeStore) Close() error {
	var errs []error
	if c, ok := s.archive.(closer); ok {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if c, ok := s.state.(closer); ok {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
