package callmanager

import (
	"context"
	"strings"
	"testing"

	"vda/internal/domain"
	"vda/internal/storage"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type stubTokenGen struct{}

func (s *stubTokenGen) GenerateToken(roomName string, userID int64, userName string) (string, error) {
	return "fake-token-" + roomName, nil
}

func newTestManager(t *testing.T) (*Manager, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	store := storage.NewRedisStore(rdb, "vdatest")
	mgr := New(&stubTokenGen{}, store, rdb)
	return mgr, rdb
}

func callerInfo(userID int64) domain.CallerInfo {
	return domain.CallerInfo{
		UserID:      userID,
		Domain:      "platform",
		TenantID:    "",
		ProjectID:   "",
		Environment: "",
	}
}

func TestManager_Initiate_Success(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, token, room, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}
	if callID == "" {
		t.Fatal("Initiate() callID is empty")
	}
	if !strings.Contains(token, "fake-token-") {
		t.Fatalf("Initiate() token = %q, want containing 'fake-token-'", token)
	}
	if !strings.HasPrefix(room, "call_") {
		t.Fatalf("Initiate() room = %q, want prefix 'call_'", room)
	}

	state, err := mgr.GetState(ctx, callID)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if state.State != domain.CallStateRinging {
		t.Fatalf("GetState().State = %q, want %q", state.State, domain.CallStateRinging)
	}
	if state.CallerID != 1001 {
		t.Fatalf("GetState().CallerID = %d, want 1001", state.CallerID)
	}
	if state.CalleeID != 2002 {
		t.Fatalf("GetState().CalleeID = %d, want 2002", state.CalleeID)
	}
}

func TestManager_Initiate_CallerAlreadyInCall(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	_, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("first Initiate() error = %v", err)
	}
	_, _, _, err = mgr.Initiate(ctx, callerInfo(1001), 3003)
	if err == nil {
		t.Fatal("second Initiate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "already in call") {
		t.Fatalf("Initiate() error = %q, want containing 'already in call'", err.Error())
	}
}

func TestManager_Initiate_CalleeAlreadyInCall(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	// User 2002 is the CALLER in the first call (gets user_call set).
	_, _, _, err := mgr.Initiate(ctx, callerInfo(2002), 3003)
	if err != nil {
		t.Fatalf("first Initiate() error = %v", err)
	}
	// Now someone tries to call 2002 as callee while 2002 has an active call.
	_, _, _, err = mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err == nil {
		t.Fatal("second Initiate() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "already in call") {
		t.Fatalf("Initiate() error = %q, want containing 'already in call'", err.Error())
	}
}

func TestManager_Accept_Success(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	token, err := mgr.Accept(ctx, callID, 2002)
	if err != nil {
		t.Fatalf("Accept() error = %v", err)
	}
	if token == "" {
		t.Fatal("Accept() token is empty")
	}
	if !strings.Contains(token, "fake-token-") {
		t.Fatalf("Accept() token = %q, want containing 'fake-token-'", token)
	}

	state, _ := mgr.GetState(ctx, callID)
	if state.State != domain.CallStateConnected {
		t.Fatalf("GetState().State after accept = %q, want %q", state.State, domain.CallStateConnected)
	}
}

func TestManager_Accept_WrongUser(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	_, err = mgr.Accept(ctx, callID, 3003)
	if err == nil {
		t.Fatal("Accept() by wrong user expected error")
	}
	if !strings.Contains(err.Error(), "not the callee") {
		t.Fatalf("Accept() error = %q, want containing 'not the callee'", err.Error())
	}
}

func TestManager_Accept_NotRinging(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}
	if _, err := mgr.Accept(ctx, callID, 2002); err != nil {
		t.Fatalf("first Accept() error = %v", err)
	}
	_, err = mgr.Accept(ctx, callID, 2002)
	if err == nil {
		t.Fatal("second Accept() expected error")
	}
	if !strings.Contains(err.Error(), "expected ringing") {
		t.Fatalf("Accept() error = %q, want containing 'expected ringing'", err.Error())
	}
}

func TestManager_Reject_Success(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	if err := mgr.Reject(ctx, callID, 2002); err != nil {
		t.Fatalf("Reject() error = %v", err)
	}

	state, _ := mgr.GetState(ctx, callID)
	if state.State != domain.CallStateRejected {
		t.Fatalf("GetState().State after reject = %q, want %q", state.State, domain.CallStateRejected)
	}

	// Caller's user_call should be cleared after reject.
	store := storage.NewRedisStore(mgr.rdb, "vdatest")
	uc, err := store.GetUserCall(ctx, 1001)
	if err != nil {
		t.Fatalf("GetUserCall() error = %v", err)
	}
	if uc != "" {
		t.Fatalf("GetUserCall(1001) after reject = %q, want empty", uc)
	}
}

func TestManager_Reject_WrongUser(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	err = mgr.Reject(ctx, callID, 3003)
	if err == nil {
		t.Fatal("Reject() by wrong user expected error")
	}
	if !strings.Contains(err.Error(), "not the callee") {
		t.Fatalf("Reject() error = %q, want containing 'not the callee'", err.Error())
	}
}

func TestManager_Cancel_Success(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	if err := mgr.Cancel(ctx, callID, 1001); err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}

	state, _ := mgr.GetState(ctx, callID)
	if state.State != domain.CallStateCancelled {
		t.Fatalf("GetState().State after cancel = %q, want %q", state.State, domain.CallStateCancelled)
	}
}

func TestManager_Cancel_NotCaller(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	err = mgr.Cancel(ctx, callID, 2002)
	if err == nil {
		t.Fatal("Cancel() by non-caller expected error")
	}
	if !strings.Contains(err.Error(), "not the caller") {
		t.Fatalf("Cancel() error = %q, want containing 'not the caller'", err.Error())
	}
}

func TestManager_End_ConnectedCall(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}
	if _, err := mgr.Accept(ctx, callID, 2002); err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	if err := mgr.End(ctx, callID, 1001); err != nil {
		t.Fatalf("End() error = %v", err)
	}

	state, _ := mgr.GetState(ctx, callID)
	if state.State != domain.CallStateEnded {
		t.Fatalf("GetState().State after end = %q, want %q", state.State, domain.CallStateEnded)
	}

	store := storage.NewRedisStore(mgr.rdb, "vdatest")
	uc1, _ := store.GetUserCall(ctx, 1001)
	uc2, _ := store.GetUserCall(ctx, 2002)
	if uc1 != "" {
		t.Fatalf("GetUserCall(1001) after end = %q, want empty", uc1)
	}
	if uc2 != "" {
		t.Fatalf("GetUserCall(2002) after end = %q, want empty", uc2)
	}
}

func TestManager_End_NotParticipant(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	err = mgr.End(ctx, callID, 3003)
	if err == nil {
		t.Fatal("End() by non-participant expected error")
	}
	if !strings.Contains(err.Error(), "not a participant") {
		t.Fatalf("End() error = %q, want containing 'not a participant'", err.Error())
	}
}

func TestManager_End_NotEndableState(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}
	if err := mgr.Reject(ctx, callID, 2002); err != nil {
		t.Fatalf("Reject() error = %v", err)
	}

	err = mgr.End(ctx, callID, 1001)
	if err == nil {
		t.Fatal("End() after reject expected error")
	}
	if !strings.Contains(err.Error(), "cannot end") {
		t.Fatalf("End() error = %q, want containing 'cannot end'", err.Error())
	}
}

func TestManager_GetState_Existing(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	call, err := mgr.GetState(ctx, callID)
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	if call.CallID != callID {
		t.Fatalf("GetState().CallID = %q, want %q", call.CallID, callID)
	}
	if call.State != domain.CallStateRinging {
		t.Fatalf("GetState().State = %q, want ringing", call.State)
	}
}

func TestManager_GetState_NonExistent(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	call, err := mgr.GetState(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetState() error = %v", err)
	}
	// Non-existent call returns zero-state VoiceCall — all fields zero/empty.
	if call.CallerID != 0 || call.CalleeID != 0 || call.State != "" {
		t.Fatalf("GetState() for nonexistent call = %+v, want zero fields", call)
	}
}

func TestManager_OnDisconnect_NoActiveCall(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	if err := mgr.OnDisconnect(ctx, 1001); err != nil {
		t.Fatalf("OnDisconnect() error = %v", err)
	}
}

func TestManager_OnDisconnect_CallerDuringRinging(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}

	if err := mgr.OnDisconnect(ctx, 1001); err != nil {
		t.Fatalf("OnDisconnect() error = %v", err)
	}

	state, _ := mgr.GetState(ctx, callID)
	if state.State != domain.CallStateCancelled {
		t.Fatalf("GetState().State after caller disconnect = %q, want %q", state.State, domain.CallStateCancelled)
	}
}

func TestManager_OnDisconnect_CalleeDuringConnected(t *testing.T) {
	mgr, _ := newTestManager(t)
	ctx := context.Background()

	callID, _, _, err := mgr.Initiate(ctx, callerInfo(1001), 2002)
	if err != nil {
		t.Fatalf("Initiate() error = %v", err)
	}
	if _, err := mgr.Accept(ctx, callID, 2002); err != nil {
		t.Fatalf("Accept() error = %v", err)
	}

	if err := mgr.OnDisconnect(ctx, 2002); err != nil {
		t.Fatalf("OnDisconnect() error = %v", err)
	}

	store := storage.NewRedisStore(mgr.rdb, "vdatest")
	uc, _ := store.GetUserCall(ctx, 2002)
	if uc != "" {
		t.Fatalf("GetUserCall(2002) after disconnect = %q, want empty", uc)
	}

	// Call should still exist and be connected.
	state, _ := mgr.GetState(ctx, callID)
	if state.State != domain.CallStateConnected {
		t.Fatalf("GetState().State after callee disconnect = %q, want %q", state.State, domain.CallStateConnected)
	}
}
