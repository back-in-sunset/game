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

// Manager 管理语音房间：加入/离开/静音/参与者查询。
type Manager struct {
	lk    domain.TokenGenerator // LiveKit token 生成器
	store *storage.RedisStore   // Redis 持久化
	rdb   *redis.Client         // Redis 客户端（muted 集合操作）
}

// New 创建房间管理器。
func New(lk domain.TokenGenerator, store *storage.RedisStore, rdb *redis.Client) *Manager {
	return &Manager{lk: lk, store: store, rdb: rdb}
}

// Join 用户加入语音房间，返回 LiveKit token 和当前参与者列表。
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

// Leave 用户离开语音房间，同时清除静音状态和 presence。
func (m *Manager) Leave(ctx context.Context, roomID string, userID int64) error {
	if err := m.store.RemoveRoomParticipant(ctx, roomID, userID); err != nil {
		return fmt.Errorf("remove participant: %w", err)
	}
	m.store.DelVoicePresence(ctx, userID)
	return nil
}

// ToggleMute 切换用户在房间中的静音状态。
func (m *Manager) ToggleMute(ctx context.Context, roomID string, userID int64, muted bool) error {
	return m.store.SetMuted(ctx, roomID, userID, muted)
}

// GetParticipants 返回房间所有参与者及其静音状态。
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

// GetMuted 返回房间中所有已静音的用户 ID 列表。
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
