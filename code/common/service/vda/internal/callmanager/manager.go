package callmanager

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

var callTTL = 5 * time.Minute // 通话状态 TTL（过期自动清理）

// Manager 管理 1v1 通话的完整生命周期。
// 状态机：ringing → connected → ended / rejected / cancelled
type Manager struct {
	lk    domain.TokenGenerator // LiveKit token 生成器
	store *storage.RedisStore   // Redis 持久化
	rdb   *redis.Client         // Redis 客户端（user_call 约束检查）
}

// New 创建通话管理器。
func New(lk domain.TokenGenerator, store *storage.RedisStore, rdb *redis.Client) *Manager {
	return &Manager{lk: lk, store: store, rdb: rdb}
}

// Initiate 发起新通话。返回 callID、LiveKit 房间名和主叫 token。
func (m *Manager) Initiate(ctx context.Context, caller domain.CallerInfo, calleeID int64) (callID string, lkToken string, lkRoom string, err error) {
	// 主叫是否已在通话中？
	existing, _ := m.store.GetUserCall(ctx, caller.UserID)
	if existing != "" {
		return "", "", "", fmt.Errorf("user %d is already in call %s", caller.UserID, existing)
	}
	// 被叫是否已在通话中？
	existing, _ = m.store.GetUserCall(ctx, calleeID)
	if existing != "" {
		return "", "", "", fmt.Errorf("user %d is already in call %s", calleeID, existing)
	}

	callID = newCallID()
	lkRoom = livekit.RoomNameForCall(callID)
	token, err := m.lk.GenerateToken(lkRoom, caller.UserID, strconv.FormatInt(caller.UserID, 10))
	if err != nil {
		return "", "", "", fmt.Errorf("generate token: %w", err)
	}

	call := domain.VoiceCall{
		CallID:      callID,
		CallerID:    caller.UserID,
		CalleeID:    calleeID,
		Domain:      caller.Domain,
		TenantID:    caller.TenantID,
		ProjectID:   caller.ProjectID,
		Environment: caller.Environment,
		State:       domain.CallStateRinging,
		LiveKitRoom: lkRoom,
	}

	if err := m.store.SaveCall(ctx, call, callTTL); err != nil {
		return "", "", "", fmt.Errorf("save call: %w", err)
	}
	if err := m.store.SetUserCall(ctx, caller.UserID, callID, callTTL); err != nil {
		return "", "", "", fmt.Errorf("set user call: %w", err)
	}

	return callID, token, lkRoom, nil
}

// Accept 接听通话：状态 ringing→connected，返回被叫 token。
func (m *Manager) Accept(ctx context.Context, callID string, userID int64) (string, error) {
	call, err := m.store.GetCall(ctx, callID)
	if err != nil {
		return "", fmt.Errorf("call %s not found: %w", callID, err)
	}
	if call.CalleeID != userID {
		return "", fmt.Errorf("user %d is not the callee of call %s", userID, callID)
	}
	if call.State != domain.CallStateRinging {
		return "", fmt.Errorf("call %s is in state %s, expected ringing", callID, call.State)
	}

	token, err := m.lk.GenerateToken(call.LiveKitRoom, userID, strconv.FormatInt(userID, 10))
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	if err := m.store.UpdateCallState(ctx, callID, domain.CallStateConnected, callTTL); err != nil {
		return "", err
	}
	if err := m.store.SetUserCall(ctx, userID, callID, callTTL); err != nil {
		return "", err
	}
	m.store.SetVoicePresence(ctx, userID, "call:"+callID, callTTL)

	return token, nil
}

// Reject 拒接通话：状态 ringing→rejected，只能被叫操作。
func (m *Manager) Reject(ctx context.Context, callID string, userID int64) error {
	call, err := m.store.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("call %s not found: %w", callID, err)
	}
	if call.CalleeID != userID {
		return fmt.Errorf("user %d is not the callee of call %s", userID, callID)
	}
	if call.State != domain.CallStateRinging {
		return fmt.Errorf("call %s is in state %s, expected ringing", callID, call.State)
	}
	if err := m.store.UpdateCallState(ctx, callID, domain.CallStateRejected, callTTL); err != nil {
		return err
	}
	m.store.DelUserCall(ctx, call.CallerID)
	return nil
}

// Cancel 主叫取消通话：状态 ringing→cancelled，只能主叫操作。
func (m *Manager) Cancel(ctx context.Context, callID string, userID int64) error {
	call, err := m.store.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("call %s not found: %w", callID, err)
	}
	if call.CallerID != userID {
		return fmt.Errorf("user %d is not the caller of call %s", userID, callID)
	}
	if call.State != domain.CallStateRinging {
		return fmt.Errorf("call %s is in state %s, expected ringing", callID, call.State)
	}
	if err := m.store.UpdateCallState(ctx, callID, domain.CallStateCancelled, callTTL); err != nil {
		return err
	}
	m.store.DelUserCall(ctx, call.CallerID)
	return nil
}

// End 挂断通话：状态 connected/ringing→ended，双方均可操作。
func (m *Manager) End(ctx context.Context, callID string, userID int64) error {
	call, err := m.store.GetCall(ctx, callID)
	if err != nil {
		return fmt.Errorf("call %s not found: %w", callID, err)
	}
	if call.CallerID != userID && call.CalleeID != userID {
		return fmt.Errorf("user %d is not a participant of call %s", userID, callID)
	}
	if call.State != domain.CallStateConnected && call.State != domain.CallStateRinging {
		return fmt.Errorf("call %s is in state %s, cannot end", callID, call.State)
	}

	if err := m.store.UpdateCallState(ctx, callID, domain.CallStateEnded, callTTL); err != nil {
		return err
	}
	m.cleanupCall(ctx, call)
	return nil
}

// GetState 查询当前通话状态。
func (m *Manager) GetState(ctx context.Context, callID string) (domain.VoiceCall, error) {
	return m.store.GetCall(ctx, callID)
}

// OnDisconnect 处理 IM 断线：主叫 ringing 时断线→cancel，connected 时断线→清除 presence。
func (m *Manager) OnDisconnect(ctx context.Context, userID int64) error {
	callID, err := m.store.GetUserCall(ctx, userID)
	if err != nil || callID == "" {
		return nil
	}
	call, err := m.store.GetCall(ctx, callID)
	if err != nil {
		return nil
	}
	if call.State == domain.CallStateRinging && call.CallerID == userID {
			// 主叫在 ringing 时断线 → 自动取消通话。
		_ = m.store.UpdateCallState(ctx, callID, domain.CallStateCancelled, callTTL)
		m.cleanupCall(ctx, call)
		return nil
	}
	// connected 状态下断线：只清除用户 presence，对方仍可挂断。
	m.store.DelVoicePresence(ctx, userID)
	m.store.DelUserCall(ctx, userID)
	return nil
}

func (m *Manager) cleanupCall(ctx context.Context, call domain.VoiceCall) {
	m.store.DelUserCall(ctx, call.CallerID)
	m.store.DelUserCall(ctx, call.CalleeID)
	m.store.DelVoicePresence(ctx, call.CallerID)
	m.store.DelVoicePresence(ctx, call.CalleeID)
}

func newCallID() string {
	return "call_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

