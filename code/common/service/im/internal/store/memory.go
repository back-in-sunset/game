package store

import (
	"context"
	"sort"
	"sync"

	"im/internal/auth"
	"im/internal/domain"
)

type MemoryStore struct {
	mu            sync.RWMutex
	messages      map[string][]domain.Envelope
	offline       map[string][]domain.Envelope
	conversations map[string]conversationState
}

type conversationState struct {
	Domain       domain.IMDomain
	Scope        domain.Scope
	LastMessage  domain.Envelope
	UpdatedAt    int64
	Participants [2]int64
	ReadSeq      map[int64]int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		messages:      make(map[string][]domain.Envelope),
		offline:       make(map[string][]domain.Envelope),
		conversations: make(map[string]conversationState),
	}
}

func (s *MemoryStore) SaveMessage(_ context.Context, envelope domain.Envelope) error {
	key, err := domain.ConversationKey(envelope.Domain, envelope.Scope, envelope.Sender, envelope.Receiver)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[key] = append(s.messages[key], envelope)
	state := s.conversations[key]
	state.Domain = envelope.Domain
	state.Scope = envelope.Scope.Normalize()
	state.LastMessage = envelope
	state.UpdatedAt = envelope.SentAt.UnixNano()
	state.Participants = participants(envelope.Sender, envelope.Receiver)
	if state.ReadSeq == nil {
		state.ReadSeq = make(map[int64]int64)
	}
	s.conversations[key] = state
	return nil
}

func (s *MemoryStore) SaveOffline(_ context.Context, principal auth.Principal, envelope domain.Envelope) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := principalKey(principal)
	s.offline[key] = append(s.offline[key], envelope)
	return nil
}

func (s *MemoryStore) DrainOffline(_ context.Context, principal auth.Principal) ([]domain.Envelope, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := principalKey(principal)
	items := s.offline[key]
	out := make([]domain.Envelope, len(items))
	copy(out, items)
	delete(s.offline, key)
	return out, nil
}

func (s *MemoryStore) ListConversations(_ context.Context, principal auth.Principal) ([]domain.Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.Conversation, 0)
	for key, state := range s.conversations {
		if !samePrincipalSpace(principal, state.Domain, state.Scope) {
			continue
		}
		peer := otherParticipant(state.Participants, principal.UserID)
		if peer == 0 {
			continue
		}
		readSeq := state.ReadSeq[principal.UserID]
		unread := int64(0)
		for _, msg := range s.messages[key] {
			if msg.Receiver == principal.UserID && msg.Seq > readSeq {
				unread++
			}
		}
		out = append(out, domain.Conversation{
			Key:         key,
			Domain:      state.Domain,
			Scope:       state.Scope,
			PeerUserID:  peer,
			LastMessage: state.LastMessage,
			UnreadCount: unread,
			UpdatedAt:   state.LastMessage.SentAt,
			ReadSeq:     readSeq,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out, nil
}

func (s *MemoryStore) ListMessages(_ context.Context, principal auth.Principal, peerUserID int64, limit int) ([]domain.Envelope, error) {
	key, err := domain.ConversationKey(principal.Domain, principal.Scope, principal.UserID, peerUserID)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	items := s.messages[key]
	if limit <= 0 || limit >= len(items) {
		out := make([]domain.Envelope, len(items))
		copy(out, items)
		return out, nil
	}
	out := make([]domain.Envelope, limit)
	copy(out, items[len(items)-limit:])
	return out, nil
}

func (s *MemoryStore) MarkRead(_ context.Context, principal auth.Principal, peerUserID int64, seq int64) error {
	key, err := domain.ConversationKey(principal.Domain, principal.Scope, principal.UserID, peerUserID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.conversations[key]
	if !ok {
		return nil
	}
	if state.ReadSeq == nil {
		state.ReadSeq = make(map[int64]int64)
	}
	if seq > state.ReadSeq[principal.UserID] {
		state.ReadSeq[principal.UserID] = seq
	}
	s.conversations[key] = state
	return nil
}

func principalKey(principal auth.Principal) string {
	scope := principal.Scope.Normalize()
	if principal.Domain == domain.DomainPlatform {
		return "platform:" + itoa(principal.UserID)
	}
	return "tenant:" + scope.TenantID + ":" + scope.ProjectID + ":" + scope.Environment + ":" + itoa(principal.UserID)
}

func samePrincipalSpace(principal auth.Principal, imDomain domain.IMDomain, scope domain.Scope) bool {
	if principal.Domain != imDomain {
		return false
	}
	if imDomain == domain.DomainPlatform {
		return true
	}
	ps := principal.Scope.Normalize()
	ss := scope.Normalize()
	return ps.TenantID == ss.TenantID && ps.ProjectID == ss.ProjectID && ps.Environment == ss.Environment
}

func participants(a, b int64) [2]int64 {
	if a < b {
		return [2]int64{a, b}
	}
	return [2]int64{b, a}
}

func otherParticipant(pair [2]int64, uid int64) int64 {
	if pair[0] == uid {
		return pair[1]
	}
	if pair[1] == uid {
		return pair[0]
	}
	return 0
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}
