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

var callTTL = 5 * time.Minute

type Manager struct {
	lk    *livekit.Client
	store *storage.RedisStore
	rdb   *redis.Client
}

func New(lk *livekit.Client, store *storage.RedisStore, rdb *redis.Client) *Manager {
	return &Manager{lk: lk, store: store, rdb: rdb}
}

// Initiate starts a new call. Returns the call ID, LiveKit room, and caller's token.
func (m *Manager) Initiate(ctx context.Context, caller domain.CallerInfo, calleeID int64) (callID string, lkToken string, lkRoom string, err error) {
	// Check if caller is already in a call.
	existing, _ := m.store.GetUserCall(ctx, caller.UserID)
	if existing != "" {
		return "", "", "", fmt.Errorf("user %d is already in call %s", caller.UserID, existing)
	}
	// Check if callee is already in a call.
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

// Accept transitions a call from ringing to connected and returns the callee's token.
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

// Reject transitions a call from ringing to rejected.
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

// Cancel transitions an outgoing call from ringing to cancelled. Only the caller can cancel.
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

// End terminates an active call. Either party can end.
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

// GetState returns the current state of a call.
func (m *Manager) GetState(ctx context.Context, callID string) (domain.VoiceCall, error) {
	return m.store.GetCall(ctx, callID)
}

// OnDisconnect cleans up calls when a user disconnects from IM.
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
		// Caller disconnected while ringing — cancel.
		_ = m.store.UpdateCallState(ctx, callID, domain.CallStateCancelled, callTTL)
		m.cleanupCall(ctx, call)
		return nil
	}
	// For connected calls, just remove user presence (other party can end).
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

