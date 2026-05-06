package roommanager

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"vda/internal/domain"
	"vda/internal/livekit"
	"vda/internal/storage"

	"github.com/redis/go-redis/v9"
)

var defaultTTL = 24 * time.Hour

type Manager struct {
	lk    *livekit.Client
	store *storage.RedisStore
	rdb   *redis.Client
}

func New(lk *livekit.Client, store *storage.RedisStore, rdb *redis.Client) *Manager {
	return &Manager{lk: lk, store: store, rdb: rdb}
}

// Join adds a user to a voice room and returns a LiveKit token.
func (m *Manager) Join(ctx context.Context, roomID string, user domain.CallerInfo) (token string, participants []int64, err error) {
	lkRoom := livekit.RoomNameForRoom(roomID)

	token, err = m.lk.GenerateToken(lkRoom, user.UserID, strconv.FormatInt(user.UserID, 10))
	if err != nil {
		return "", nil, fmt.Errorf("generate token: %w", err)
	}

	if err := m.store.AddRoomParticipant(ctx, roomID, user.UserID); err != nil {
		return "", nil, fmt.Errorf("add participant: %w", err)
	}

	activity := "room:" + roomID
	if err := m.store.SetVoicePresence(ctx, user.UserID, activity, defaultTTL); err != nil {
		return "", nil, fmt.Errorf("set presence: %w", err)
	}

	participants, err = m.store.GetRoomParticipants(ctx, roomID)
	if err != nil {
		return "", nil, err
	}

	return token, participants, nil
}

// Leave removes a user from a voice room.
func (m *Manager) Leave(ctx context.Context, roomID string, userID int64) error {
	if err := m.store.RemoveRoomParticipant(ctx, roomID, userID); err != nil {
		return fmt.Errorf("remove participant: %w", err)
	}
	m.store.DelVoicePresence(ctx, userID)
	return nil
}

// ToggleMute toggles a user's mute state in a room.
func (m *Manager) ToggleMute(ctx context.Context, roomID string, userID int64, muted bool) error {
	return m.store.SetMuted(ctx, roomID, userID, muted)
}

// GetParticipants returns all participants in a room and their mute states.
func (m *Manager) GetParticipants(ctx context.Context, roomID string) ([]domain.Participant, error) {
	ids, err := m.store.GetRoomParticipants(ctx, roomID)
	if err != nil {
		return nil, err
	}
	participants := make([]domain.Participant, 0, len(ids))
	for _, id := range ids {
		muted, _ := m.store.IsMuted(ctx, roomID, id)
		participants = append(participants, domain.Participant{
			UserID: id,
			Muted:  muted,
		})
	}
	return participants, nil
}

// GetMuted returns the list of muted user IDs in a room.
func (m *Manager) GetMuted(ctx context.Context, roomID string) ([]int64, error) {
	vals, err := m.rdb.SMembers(ctx, m.storePrefix()+":room:"+roomID+":muted").Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(vals))
	for _, v := range vals {
		var n int64
		for _, c := range v {
			n = n*10 + int64(c-'0')
		}
		ids = append(ids, n)
	}
	return ids, nil
}

func (m *Manager) storePrefix() string {
	return "vda"
}
