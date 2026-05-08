package service

import (
	"context"
	"net"
	"strings"
	"testing"

	"vda/internal/callmanager"
	"vda/internal/domain"
	"vda/internal/roommanager"
	"vda/internal/storage"
	rpc "vda/rpc/vdaclient"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type stubTokenGen struct{}

func (s *stubTokenGen) GenerateToken(roomName string, userID int64, userName string) (string, error) {
	return "fake-token-" + roomName, nil
}

func setupVDATestServer(t *testing.T) (rpc.VDAClient, func()) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := storage.NewRedisStore(rdb, "vda")
	callMgr := callmanager.New(&stubTokenGen{}, store, rdb)
	roomMgr := roommanager.New(&stubTokenGen{}, store, rdb)

	srv := grpc.NewServer()
	rpc.RegisterVDAServer(srv, NewVDAServer(callMgr, roomMgr, "http://livekit:7880"))

	lis := bufconn.Listen(1024 * 1024)
	go srv.Serve(lis)

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		srv.Stop()
		mr.Close()
		rdb.Close()
		t.Fatalf("grpc.NewClient() error = %v", err)
	}

	cleanup := func() {
		conn.Close()
		srv.Stop()
		mr.Close()
		rdb.Close()
	}
	return rpc.NewVDAClient(conn), cleanup
}

func callerInfo(userID int64) *rpc.CallerInfo {
	return &rpc.CallerInfo{
		UserID:      userID,
		Domain:      "platform",
		TenantID:    "",
		ProjectID:   "",
		Environment: "",
	}
}

// --- Call flow tests ---

func TestVDAServer_InitiateCall(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	resp, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
		Caller: callerInfo(1001),
		Callee: 2002,
	})
	if err != nil {
		t.Fatalf("InitiateCall() error = %v", err)
	}
	if resp.CallID == "" {
		t.Fatal("InitiateCall() CallID is empty")
	}
	if resp.State != string(domain.CallStateRinging) {
		t.Fatalf("InitiateCall() State = %q, want ringing", resp.State)
	}
	if resp.LiveKitToken == "" {
		t.Fatal("InitiateCall() LiveKitToken is empty")
	}
	if resp.LiveKitRoom == "" {
		t.Fatal("InitiateCall() LiveKitRoom is empty")
	}
	if resp.LiveKitUrl == "" {
		t.Fatal("InitiateCall() LiveKitUrl is empty")
	}
}

func TestVDAServer_FullCallLifecycle(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	// Initiate.
	initResp, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
		Caller: callerInfo(1001),
		Callee: 2002,
	})
	if err != nil {
		t.Fatalf("InitiateCall() error = %v", err)
	}

	// Accept.
	acceptResp, err := client.AcceptCall(ctx, &rpc.AcceptCallRequest{
		CallID: initResp.CallID,
		UserID: 2002,
	})
	if err != nil {
		t.Fatalf("AcceptCall() error = %v", err)
	}
	if acceptResp.State != string(domain.CallStateConnected) {
		t.Fatalf("AcceptCall() State = %q, want connected", acceptResp.State)
	}
	if acceptResp.LiveKitUrl == "" {
		t.Fatal("AcceptCall() LiveKitUrl is empty")
	}

	// GetCallState should show connected.
	stateResp, err := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: initResp.CallID})
	if err != nil {
		t.Fatalf("GetCallState() error = %v", err)
	}
	if stateResp.State != string(domain.CallStateConnected) {
		t.Fatalf("GetCallState() State = %q, want connected", stateResp.State)
	}

	// End.
	endResp, err := client.EndCall(ctx, &rpc.EndCallRequest{
		CallID: initResp.CallID,
		UserID: 1001,
	})
	if err != nil {
		t.Fatalf("EndCall() error = %v", err)
	}
	if !endResp.Success {
		t.Fatal("EndCall() Success = false")
	}

	// State should be ended.
	stateResp, err = client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: initResp.CallID})
	if err != nil {
		t.Fatalf("GetCallState() after end error = %v", err)
	}
	if stateResp.State != string(domain.CallStateEnded) {
		t.Fatalf("GetCallState() State after end = %q, want ended", stateResp.State)
	}
}

func TestVDAServer_RejectCall(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	initResp, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
		Caller: callerInfo(1001),
		Callee: 2002,
	})
	if err != nil {
		t.Fatalf("InitiateCall() error = %v", err)
	}

	rejResp, err := client.RejectCall(ctx, &rpc.RejectCallRequest{
		CallID: initResp.CallID,
		UserID: 2002,
	})
	if err != nil {
		t.Fatalf("RejectCall() error = %v", err)
	}
	if !rejResp.Success {
		t.Fatal("RejectCall() Success = false")
	}

	stateResp, _ := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: initResp.CallID})
	if stateResp.State != string(domain.CallStateRejected) {
		t.Fatalf("GetCallState() State = %q, want rejected", stateResp.State)
	}
}

func TestVDAServer_CancelCall(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	initResp, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
		Caller: callerInfo(1001),
		Callee: 2002,
	})
	if err != nil {
		t.Fatalf("InitiateCall() error = %v", err)
	}

	cancelResp, err := client.CancelCall(ctx, &rpc.CancelCallRequest{
		CallID: initResp.CallID,
		UserID: 1001,
	})
	if err != nil {
		t.Fatalf("CancelCall() error = %v", err)
	}
	if !cancelResp.Success {
		t.Fatal("CancelCall() Success = false")
	}

	stateResp, _ := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: initResp.CallID})
	if stateResp.State != string(domain.CallStateCancelled) {
		t.Fatalf("GetCallState() State = %q, want cancelled", stateResp.State)
	}
}

func TestVDAServer_EndCall_NotParticipant(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	initResp, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
		Caller: callerInfo(1001),
		Callee: 2002,
	})
	if err != nil {
		t.Fatalf("InitiateCall() error = %v", err)
	}

	_, err = client.EndCall(ctx, &rpc.EndCallRequest{
		CallID: initResp.CallID,
		UserID: 3003,
	})
	if err == nil {
		t.Fatal("EndCall() by non-participant expected error")
	}
}

// --- Room flow tests ---

func TestVDAServer_JoinVoiceRoom(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	resp, err := client.JoinVoiceRoom(ctx, &rpc.JoinVoiceRoomRequest{
		RoomID: "room-1",
		User:   callerInfo(1001),
	})
	if err != nil {
		t.Fatalf("JoinVoiceRoom() error = %v", err)
	}
	if resp.RoomID != "room-1" {
		t.Fatalf("JoinVoiceRoom() RoomID = %q, want room-1", resp.RoomID)
	}
	if resp.LiveKitToken == "" {
		t.Fatal("JoinVoiceRoom() LiveKitToken is empty")
	}
	if len(resp.Participants) == 0 || resp.Participants[0] != 1001 {
		t.Fatalf("JoinVoiceRoom() Participants = %v, want [1001]", resp.Participants)
	}
	if resp.LiveKitUrl == "" {
		t.Fatal("JoinVoiceRoom() LiveKitUrl is empty")
	}
}

func TestVDAServer_JoinLeaveRoom(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	if _, err := client.JoinVoiceRoom(ctx, &rpc.JoinVoiceRoomRequest{
		RoomID: "room-1",
		User:   callerInfo(1001),
	}); err != nil {
		t.Fatalf("JoinVoiceRoom() error = %v", err)
	}

	leaveResp, err := client.LeaveVoiceRoom(ctx, &rpc.LeaveVoiceRoomRequest{
		RoomID: "room-1",
		UserID: 1001,
	})
	if err != nil {
		t.Fatalf("LeaveVoiceRoom() error = %v", err)
	}
	if !leaveResp.Success {
		t.Fatal("LeaveVoiceRoom() Success = false")
	}

	stateResp, _ := client.GetRoomState(ctx, &rpc.GetRoomStateRequest{RoomID: "room-1"})
	if len(stateResp.Participants) != 0 {
		t.Fatalf("GetRoomState() Participants = %v, want empty", stateResp.Participants)
	}
}

func TestVDAServer_MuteToggle(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	if _, err := client.JoinVoiceRoom(ctx, &rpc.JoinVoiceRoomRequest{
		RoomID: "room-1",
		User:   callerInfo(1001),
	}); err != nil {
		t.Fatalf("JoinVoiceRoom() error = %v", err)
	}

	muteResp, err := client.MuteToggle(ctx, &rpc.MuteToggleRequest{
		RoomID: "room-1",
		UserID: 1001,
		Muted:  true,
	})
	if err != nil {
		t.Fatalf("MuteToggle() error = %v", err)
	}
	if !muteResp.Muted {
		t.Fatal("MuteToggle() Muted = false, want true")
	}

	stateResp, _ := client.GetRoomState(ctx, &rpc.GetRoomStateRequest{RoomID: "room-1"})
	found := false
	for _, id := range stateResp.Muted {
		if id == 1001 {
			found = true
		}
	}
	if !found {
		t.Fatalf("GetRoomState() Muted = %v, want containing 1001", stateResp.Muted)
	}
}

func TestVDAServer_MuteToggle_Unmute(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	if _, err := client.JoinVoiceRoom(ctx, &rpc.JoinVoiceRoomRequest{
		RoomID: "room-1",
		User:   callerInfo(1001),
	}); err != nil {
		t.Fatalf("JoinVoiceRoom() error = %v", err)
	}

	client.MuteToggle(ctx, &rpc.MuteToggleRequest{RoomID: "room-1", UserID: 1001, Muted: true})
	client.MuteToggle(ctx, &rpc.MuteToggleRequest{RoomID: "room-1", UserID: 1001, Muted: false})

	stateResp, _ := client.GetRoomState(ctx, &rpc.GetRoomStateRequest{RoomID: "room-1"})
	for _, id := range stateResp.Muted {
		if id == 1001 {
			t.Fatal("GetRoomState() Muted contains 1001 after unmute")
		}
	}
}

func TestVDAServer_GetRoomState_MultipleUsers(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	client.JoinVoiceRoom(ctx, &rpc.JoinVoiceRoomRequest{RoomID: "room-1", User: callerInfo(1001)})
	client.JoinVoiceRoom(ctx, &rpc.JoinVoiceRoomRequest{RoomID: "room-1", User: callerInfo(2002)})
	client.MuteToggle(ctx, &rpc.MuteToggleRequest{RoomID: "room-1", UserID: 1001, Muted: true})

	stateResp, err := client.GetRoomState(ctx, &rpc.GetRoomStateRequest{RoomID: "room-1"})
	if err != nil {
		t.Fatalf("GetRoomState() error = %v", err)
	}

	has1, has2 := false, false
	for _, id := range stateResp.Participants {
		if id == 1001 {
			has1 = true
		}
		if id == 2002 {
			has2 = true
		}
	}
	if !has1 || !has2 {
		t.Fatalf("GetRoomState() Participants = %v, want [1001, 2002]", stateResp.Participants)
	}

	muted1 := false
	for _, id := range stateResp.Muted {
		if id == 1001 {
			muted1 = true
		}
	}
	if !muted1 {
		t.Fatalf("GetRoomState() Muted = %v, want containing 1001", stateResp.Muted)
	}
}

// --- HandleVoiceEvent tests ---

func TestVDAServer_HandleVoiceEvent_CallInvite(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action:     "call_invite",
		Caller:     callerInfo(1001),
		ReceiverID: 2002,
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(call_invite) error = %v", err)
	}
	if resp.CallID == "" {
		t.Fatal("HandleVoiceEvent() CallID is empty")
	}
	if resp.State != string(domain.CallStateRinging) {
		t.Fatalf("HandleVoiceEvent() State = %q, want ringing", resp.State)
	}
	if resp.LiveKitToken == "" {
		t.Fatal("HandleVoiceEvent() LiveKitToken is empty")
	}
	if resp.LiveKitUrl == "" {
		t.Fatal("HandleVoiceEvent() LiveKitUrl is empty")
	}
	if len(resp.PushTargets) != 2 {
		t.Fatalf("HandleVoiceEvent() PushTargets = %v, want [2002, 1001]", resp.PushTargets)
	}
}

func TestVDAServer_HandleVoiceEvent_CallAccept(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	invResp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action:     "call_invite",
		Caller:     callerInfo(1001),
		ReceiverID: 2002,
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(call_invite) error = %v", err)
	}

	acceptResp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_accept",
		Caller: callerInfo(2002),
		CallID: invResp.CallID,
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(call_accept) error = %v", err)
	}
	if acceptResp.State != string(domain.CallStateConnected) {
		t.Fatalf("HandleVoiceEvent() State = %q, want connected", acceptResp.State)
	}
	if len(acceptResp.PushTargets) == 0 || acceptResp.PushTargets[0] != 1001 {
		t.Fatalf("HandleVoiceEvent() PushTargets = %v, want [1001]", acceptResp.PushTargets)
	}
}

func TestVDAServer_HandleVoiceEvent_CallReject(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	invResp, _ := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_invite", Caller: callerInfo(1001), ReceiverID: 2002,
	})

	resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_reject",
		Caller: callerInfo(2002),
		CallID: invResp.CallID,
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(call_reject) error = %v", err)
	}
	if resp.State != string(domain.CallStateRejected) {
		t.Fatalf("HandleVoiceEvent() State = %q, want rejected", resp.State)
	}
}

func TestVDAServer_HandleVoiceEvent_CallEnd(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	invResp, _ := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_invite", Caller: callerInfo(1001), ReceiverID: 2002,
	})
	client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_accept", Caller: callerInfo(2002), CallID: invResp.CallID,
	})

	resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_end",
		Caller: callerInfo(1001),
		CallID: invResp.CallID,
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(call_end) error = %v", err)
	}
	if resp.State != string(domain.CallStateEnded) {
		t.Fatalf("HandleVoiceEvent() State = %q, want ended", resp.State)
	}
}

func TestVDAServer_HandleVoiceEvent_CallCancel(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	invResp, _ := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_invite", Caller: callerInfo(1001), ReceiverID: 2002,
	})

	resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_cancel",
		Caller: callerInfo(1001),
		CallID: invResp.CallID,
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(call_cancel) error = %v", err)
	}
	if resp.State != string(domain.CallStateCancelled) {
		t.Fatalf("HandleVoiceEvent() State = %q, want cancelled", resp.State)
	}
}

func TestVDAServer_HandleVoiceEvent_RoomJoin(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "room_join",
		Caller: callerInfo(1001),
		RoomID: "room-1",
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(room_join) error = %v", err)
	}
	if resp.LiveKitToken == "" {
		t.Fatal("HandleVoiceEvent() LiveKitToken is empty")
	}
	if resp.LiveKitUrl == "" {
		t.Fatal("HandleVoiceEvent() LiveKitUrl is empty")
	}
	if len(resp.PushTargets) == 0 || resp.PushTargets[0] != 1001 {
		t.Fatalf("HandleVoiceEvent() PushTargets = %v, want [1001]", resp.PushTargets)
	}
}

func TestVDAServer_HandleVoiceEvent_RoomLeave(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "room_join", Caller: callerInfo(1001), RoomID: "room-1",
	})

	resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "room_leave",
		Caller: callerInfo(1001),
		RoomID: "room-1",
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(room_leave) error = %v", err)
	}
	// After leave, PushTargets should be empty (no remaining participants).
	if len(resp.PushTargets) != 0 {
		t.Fatalf("HandleVoiceEvent() PushTargets = %v, want empty", resp.PushTargets)
	}
}

func TestVDAServer_HandleVoiceEvent_Disconnect(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	invResp, _ := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "call_invite", Caller: callerInfo(1001), ReceiverID: 2002,
	})

	_, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "disconnect",
		Caller: callerInfo(1001),
		CallID: invResp.CallID,
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(disconnect) error = %v", err)
	}

	stateResp, _ := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: invResp.CallID})
	if stateResp.State != string(domain.CallStateCancelled) {
		t.Fatalf("GetCallState() after disconnect = %q, want cancelled", stateResp.State)
	}
}

func TestVDAServer_HandleVoiceEvent_UnknownAction(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
		Action: "unknown_action",
		Caller: callerInfo(1001),
	})
	if err != nil {
		t.Fatalf("HandleVoiceEvent(unknown) error = %v", err)
	}
	if resp.CallID != "" || resp.State != "" {
		t.Fatalf("HandleVoiceEvent(unknown) response = %+v, want empty", resp)
	}
}

func TestVDAServer_GetCallState(t *testing.T) {
	client, cleanup := setupVDATestServer(t)
	defer cleanup()
	ctx := context.Background()

	initResp, _ := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
		Caller: callerInfo(1001),
		Callee: 2002,
	})

	stateResp, err := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: initResp.CallID})
	if err != nil {
		t.Fatalf("GetCallState() error = %v", err)
	}
	if stateResp.Caller != 1001 {
		t.Fatalf("GetCallState() Caller = %d, want 1001", stateResp.Caller)
	}
	if stateResp.Callee != 2002 {
		t.Fatalf("GetCallState() Callee = %d, want 2002", stateResp.Callee)
	}
	if !strings.Contains(stateResp.LiveKitRoom, "call_") {
		t.Fatalf("GetCallState() LiveKitRoom = %q, want containing 'call_'", stateResp.LiveKitRoom)
	}
}
