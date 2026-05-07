package storage

import (
	"context"
	"fmt"
	"time"

	"vda/internal/domain"

	"github.com/redis/go-redis/v9"
)

// RedisStore 封装 VDA 所有 Redis 操作：通话状态、用户通话绑定、房间参与者、语音 presence。
type RedisStore struct {
	rdb    *redis.Client
	prefix string // key 前缀，默认 "vda"
}

// NewRedisStore 创建 Redis 存储实例。
func NewRedisStore(rdb *redis.Client, prefix string) *RedisStore {
	return &RedisStore{rdb: rdb, prefix: prefix}
}

// ---- 通话状态 (call) ----

func (s *RedisStore) callKey(callID, field string) string {
	return fmt.Sprintf("%s:call:%s:%s", s.prefix, callID, field)
}

func (s *RedisStore) SaveCall(ctx context.Context, call domain.VoiceCall, ttl time.Duration) error {
	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, s.callKey(call.CallID, "caller"), call.CallerID, ttl)
	pipe.Set(ctx, s.callKey(call.CallID, "callee"), call.CalleeID, ttl)
	pipe.Set(ctx, s.callKey(call.CallID, "state"), string(call.State), ttl)
	pipe.Set(ctx, s.callKey(call.CallID, "domain"), call.Domain, ttl)
	pipe.Set(ctx, s.callKey(call.CallID, "tenant_id"), call.TenantID, ttl)
	pipe.Set(ctx, s.callKey(call.CallID, "livekit_room"), call.LiveKitRoom, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetCall(ctx context.Context, callID string) (domain.VoiceCall, error) {
	keys := []string{"caller", "callee", "state", "domain", "tenant_id", "livekit_room"}
	vals, err := s.rdb.MGet(ctx, s.callKey(callID, keys[0]),
		s.callKey(callID, keys[1]),
		s.callKey(callID, keys[2]),
		s.callKey(callID, keys[3]),
		s.callKey(callID, keys[4]),
		s.callKey(callID, keys[5]),
	).Result()
	if err != nil {
		return domain.VoiceCall{}, err
	}
	call := domain.VoiceCall{CallID: callID}
	for i, v := range vals {
		if v == nil {
			continue
		}
		s := v.(string)
		switch keys[i] {
		case "caller":
			call.CallerID = parseI64(s)
		case "callee":
			call.CalleeID = parseI64(s)
		case "state":
			call.State = domain.CallState(s)
		case "domain":
			call.Domain = s
		case "tenant_id":
			call.TenantID = s
		case "livekit_room":
			call.LiveKitRoom = s
		}
	}
	return call, nil
}

func (s *RedisStore) UpdateCallState(ctx context.Context, callID string, state domain.CallState, ttl time.Duration) error {
	_, err := s.rdb.Set(ctx, s.callKey(callID, "state"), string(state), ttl).Result()
	return err
}

func (s *RedisStore) DeleteCall(ctx context.Context, callID string) error {
	keys := []string{"caller", "callee", "state", "domain", "tenant_id", "livekit_room"}
	pipe := s.rdb.Pipeline()
	for _, k := range keys {
		pipe.Del(ctx, s.callKey(callID, k))
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) userCallKey(userID int64) string {
	return fmt.Sprintf("%s:user_call:%d", s.prefix, userID)
}

func (s *RedisStore) SetUserCall(ctx context.Context, userID int64, callID string, ttl time.Duration) error {
	_, err := s.rdb.Set(ctx, s.userCallKey(userID), callID, ttl).Result()
	return err
}

func (s *RedisStore) GetUserCall(ctx context.Context, userID int64) (string, error) {
	val, err := s.rdb.Get(ctx, s.userCallKey(userID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (s *RedisStore) DelUserCall(ctx context.Context, userID int64) error {
	_, err := s.rdb.Del(ctx, s.userCallKey(userID)).Result()
	return err
}

// ---- 房间参与者 (room) ----

func (s *RedisStore) roomParticipantsKey(roomID string) string {
	return fmt.Sprintf("%s:room:%s:participants", s.prefix, roomID)
}

func (s *RedisStore) roomMutedKey(roomID string) string {
	return fmt.Sprintf("%s:room:%s:muted", s.prefix, roomID)
}

func (s *RedisStore) AddRoomParticipant(ctx context.Context, roomID string, userID int64) error {
	_, err := s.rdb.SAdd(ctx, s.roomParticipantsKey(roomID), userID).Result()
	return err
}

func (s *RedisStore) RemoveRoomParticipant(ctx context.Context, roomID string, userID int64) error {
	pipe := s.rdb.Pipeline()
	pipe.SRem(ctx, s.roomParticipantsKey(roomID), userID)
	pipe.SRem(ctx, s.roomMutedKey(roomID), userID)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) GetRoomParticipants(ctx context.Context, roomID string) ([]int64, error) {
	vals, err := s.rdb.SMembers(ctx, s.roomParticipantsKey(roomID)).Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(vals))
	for _, v := range vals {
		ids = append(ids, parseI64(v))
	}
	return ids, nil
}

func (s *RedisStore) SetMuted(ctx context.Context, roomID string, userID int64, muted bool) error {
	if muted {
		_, err := s.rdb.SAdd(ctx, s.roomMutedKey(roomID), userID).Result()
		return err
	}
	_, err := s.rdb.SRem(ctx, s.roomMutedKey(roomID), userID).Result()
	return err
}

func (s *RedisStore) IsMuted(ctx context.Context, roomID string, userID int64) (bool, error) {
	return s.rdb.SIsMember(ctx, s.roomMutedKey(roomID), userID).Result()
}

// ---- 语音 presence (voice presence) ----

func (s *RedisStore) voicePresenceKey(userID int64) string {
	return fmt.Sprintf("%s:presence:voice:%d", s.prefix, userID)
}

func (s *RedisStore) SetVoicePresence(ctx context.Context, userID int64, activity string, ttl time.Duration) error {
	_, err := s.rdb.Set(ctx, s.voicePresenceKey(userID), activity, ttl).Result()
	return err
}

func (s *RedisStore) DelVoicePresence(ctx context.Context, userID int64) error {
	_, err := s.rdb.Del(ctx, s.voicePresenceKey(userID)).Result()
	return err
}

func (s *RedisStore) GetVoicePresence(ctx context.Context, userID int64) (string, error) {
	val, err := s.rdb.Get(ctx, s.voicePresenceKey(userID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func parseI64(s string) int64 {
	var n int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}
	return n
}
