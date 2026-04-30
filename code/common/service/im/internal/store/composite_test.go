package store

import (
	"context"
	"testing"
	"time"

	"im/internal/auth"
	"im/internal/domain"
)

func TestCompositeStoreRoutesArchiveAndState(t *testing.T) {
	archive := &fakeArchive{}
	state := &fakeState{}
	store := NewCompositeStore(archive, state)

	envelope := domain.Envelope{
		Domain:   domain.DomainPlatform,
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      7,
		Payload:  map[string]any{"text": "hi"},
		SentAt:   time.Now(),
	}
	if err := store.SaveMessage(context.Background(), envelope); err != nil {
		t.Fatalf("SaveMessage() error = %v", err)
	}
	if archive.appended == nil {
		t.Fatal("archive.AppendMessage was not called")
	}
	if state.upserted == nil {
		t.Fatal("state.UpsertConversationPair was not called")
	}

	principal := auth.Principal{UserID: 2, Domain: domain.DomainPlatform}
	if err := store.SaveOffline(context.Background(), principal, envelope); err != nil {
		t.Fatalf("SaveOffline() error = %v", err)
	}
	if state.offlineSaved == nil {
		t.Fatal("state.SaveOffline was not called")
	}
}

type fakeArchive struct {
	appended *domain.Envelope
}

func (f *fakeArchive) AppendMessage(_ context.Context, envelope domain.Envelope) (StoredMessage, error) {
	f.appended = &envelope
	return StoredMessage{ID: 1, ConversationKey: "platform:1:2", Envelope: envelope}, nil
}

func (f *fakeArchive) ListMessages(_ context.Context, _ auth.Principal, _ int64, _ int) ([]domain.Envelope, error) {
	return nil, nil
}

type fakeState struct {
	upserted     *StoredMessage
	offlineSaved *domain.Envelope
}

func (f *fakeState) UpsertConversationPair(_ context.Context, stored StoredMessage) error {
	f.upserted = &stored
	return nil
}

func (f *fakeState) SaveOffline(_ context.Context, _ auth.Principal, envelope domain.Envelope) error {
	f.offlineSaved = &envelope
	return nil
}

func (f *fakeState) DrainOffline(_ context.Context, _ auth.Principal) ([]domain.Envelope, error) {
	return nil, nil
}

func (f *fakeState) ListConversations(_ context.Context, _ auth.Principal) ([]domain.Conversation, error) {
	return nil, nil
}

func (f *fakeState) MarkRead(_ context.Context, _ auth.Principal, _ int64, _ int64) error {
	return nil
}
