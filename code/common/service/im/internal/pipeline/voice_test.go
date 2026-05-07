package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"im/internal/auth"
	"im/internal/domain"

	rpc "vda/rpc"

	"google.golang.org/grpc"
)

// fakeVDAClient implements rpc.VDAClient for testing.
type fakeVDAClient struct {
	lastVoiceReq *rpc.VoiceEventRequest
	voiceResp    *rpc.VoiceEventResponse
	voiceErr     error
}

func (f *fakeVDAClient) HandleVoiceEvent(ctx context.Context, in *rpc.VoiceEventRequest, opts ...grpc.CallOption) (*rpc.VoiceEventResponse, error) {
	f.lastVoiceReq = in
	if f.voiceErr != nil {
		return nil, f.voiceErr
	}
	if f.voiceResp != nil {
		return f.voiceResp, nil
	}
	return &rpc.VoiceEventResponse{}, nil
}

func (f *fakeVDAClient) InitiateCall(ctx context.Context, in *rpc.InitiateCallRequest, opts ...grpc.CallOption) (*rpc.InitiateCallResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) AcceptCall(ctx context.Context, in *rpc.AcceptCallRequest, opts ...grpc.CallOption) (*rpc.AcceptCallResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) RejectCall(ctx context.Context, in *rpc.RejectCallRequest, opts ...grpc.CallOption) (*rpc.ActionResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) EndCall(ctx context.Context, in *rpc.EndCallRequest, opts ...grpc.CallOption) (*rpc.ActionResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) CancelCall(ctx context.Context, in *rpc.CancelCallRequest, opts ...grpc.CallOption) (*rpc.ActionResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) JoinVoiceRoom(ctx context.Context, in *rpc.JoinVoiceRoomRequest, opts ...grpc.CallOption) (*rpc.JoinVoiceRoomResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) LeaveVoiceRoom(ctx context.Context, in *rpc.LeaveVoiceRoomRequest, opts ...grpc.CallOption) (*rpc.ActionResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) MuteToggle(ctx context.Context, in *rpc.MuteToggleRequest, opts ...grpc.CallOption) (*rpc.MuteToggleResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) GetCallState(ctx context.Context, in *rpc.GetCallStateRequest, opts ...grpc.CallOption) (*rpc.CallStateResponse, error) {
	return nil, nil
}
func (f *fakeVDAClient) GetRoomState(ctx context.Context, in *rpc.GetRoomStateRequest, opts ...grpc.CallOption) (*rpc.RoomStateResponse, error) {
	return nil, nil
}

func testPrincipal() auth.Principal {
	return auth.Principal{
		UserID: 1001,
		Domain: domain.DomainPlatform,
	}
}

func newTestVoiceHandler() (*VoiceHandler, *fakeVDAClient) {
	fake := &fakeVDAClient{
		voiceResp: &rpc.VoiceEventResponse{
			CallID:       "test-call-1",
			State:        "ringing",
			LiveKitToken: "test-token",
			LiveKitRoom:  "call_test-call-1",
			PushJson:     `{"type":"call_invite"}`,
		},
	}
	return NewVoiceHandler(fake), fake
}

// --- Call command parsing tests ---

func TestVoiceHandler_CallInvite(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_invite", json.RawMessage(`{"callee":200}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq == nil {
		t.Fatal("lastVoiceReq is nil")
	}
	if fake.lastVoiceReq.Action != "call_invite" {
		t.Fatalf("lastVoiceReq.Action = %q, want call_invite", fake.lastVoiceReq.Action)
	}
	if fake.lastVoiceReq.ReceiverID != 200 {
		t.Fatalf("lastVoiceReq.ReceiverID = %d, want 200", fake.lastVoiceReq.ReceiverID)
	}
}

func TestVoiceHandler_CallAccept(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_accept", json.RawMessage(`{"call_id":"call-1"}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq.CallID != "call-1" {
		t.Fatalf("lastVoiceReq.CallID = %q, want call-1", fake.lastVoiceReq.CallID)
	}
}

func TestVoiceHandler_CallReject(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_reject", json.RawMessage(`{"call_id":"call-1"}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq.Action != "call_reject" {
		t.Fatalf("lastVoiceReq.Action = %q, want call_reject", fake.lastVoiceReq.Action)
	}
}

func TestVoiceHandler_CallEnd(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_end", json.RawMessage(`{"call_id":"call-1"}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq.Action != "call_end" {
		t.Fatalf("lastVoiceReq.Action = %q, want call_end", fake.lastVoiceReq.Action)
	}
}

func TestVoiceHandler_CallCancel(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_cancel", json.RawMessage(`{"call_id":"call-1"}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq.Action != "call_cancel" {
		t.Fatalf("lastVoiceReq.Action = %q, want call_cancel", fake.lastVoiceReq.Action)
	}
}

// --- Room command parsing tests ---

func TestVoiceHandler_RoomJoin(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "room_join", json.RawMessage(`{"room_id":"room-1"}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq.RoomID != "room-1" {
		t.Fatalf("lastVoiceReq.RoomID = %q, want room-1", fake.lastVoiceReq.RoomID)
	}
}

func TestVoiceHandler_RoomLeave(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "room_leave", json.RawMessage(`{"room_id":"room-1"}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq.Action != "room_leave" {
		t.Fatalf("lastVoiceReq.Action = %q, want room_leave", fake.lastVoiceReq.Action)
	}
}

func TestVoiceHandler_MuteToggle(t *testing.T) {
	handler, fake := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "mute_toggle", json.RawMessage(`{"room_id":"r1","muted":true}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if fake.lastVoiceReq.RoomID != "r1" {
		t.Fatalf("lastVoiceReq.RoomID = %q, want r1", fake.lastVoiceReq.RoomID)
	}
}

// --- Error cases ---

func TestVoiceHandler_InvalidJSON_CallInvite(t *testing.T) {
	handler, _ := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_invite", json.RawMessage(`{invalid}`))
	if err == nil {
		t.Fatal("HandleVoiceCommand() expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid call_invite payload") {
		t.Fatalf("error = %q, want containing 'invalid call_invite payload'", err.Error())
	}
}

func TestVoiceHandler_InvalidJSON_CallAccept(t *testing.T) {
	handler, _ := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_accept", json.RawMessage(`{invalid}`))
	if err == nil {
		t.Fatal("HandleVoiceCommand() expected error for invalid JSON")
	}
}

func TestVoiceHandler_InvalidJSON_RoomJoin(t *testing.T) {
	handler, _ := newTestVoiceHandler()
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "room_join", json.RawMessage(`{invalid}`))
	if err == nil {
		t.Fatal("HandleVoiceCommand() expected error for invalid JSON")
	}
}

func TestVoiceHandler_VDAError(t *testing.T) {
	handler := NewVoiceHandler(&fakeVDAClient{
		voiceErr: errors.New("vda down"),
	})
	principal := testPrincipal()

	_, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_invite", json.RawMessage(`{"callee":200}`))
	if err == nil {
		t.Fatal("HandleVoiceCommand() expected error")
	}
	if !strings.Contains(err.Error(), "vda voice event") {
		t.Fatalf("error = %q, want containing 'vda voice event'", err.Error())
	}
}

// --- ACK content tests ---

func TestVoiceHandler_AckContainsState(t *testing.T) {
	handler, _ := newTestVoiceHandler()
	principal := testPrincipal()

	ack, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_invite", json.RawMessage(`{"callee":200}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(ack, &parsed); err != nil {
		t.Fatalf("Unmarshal ack error = %v", err)
	}
	if parsed["state"] != "ringing" {
		t.Fatalf("ack state = %v, want ringing", parsed["state"])
	}
}

func TestVoiceHandler_AckContainsLiveKitRoom(t *testing.T) {
	handler, _ := newTestVoiceHandler()
	principal := testPrincipal()

	ack, _, err := handler.HandleVoiceCommand(context.Background(), principal, "call_invite", json.RawMessage(`{"callee":200}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(ack, &parsed); err != nil {
		t.Fatalf("Unmarshal ack error = %v", err)
	}
	if parsed["livekit_room"] != "call_test-call-1" {
		t.Fatalf("ack livekit_room = %v, want call_test-call-1", parsed["livekit_room"])
	}
}

func TestVoiceHandler_WithPushData(t *testing.T) {
	handler, _ := newTestVoiceHandler()
	principal := testPrincipal()

	_, pushData, err := handler.HandleVoiceCommand(context.Background(), principal, "call_invite", json.RawMessage(`{"callee":200}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if pushData == nil {
		t.Fatal("pushData is nil, want non-nil")
	}
}

func TestVoiceHandler_NoPushData(t *testing.T) {
	fake := &fakeVDAClient{
		voiceResp: &rpc.VoiceEventResponse{
			CallID:       "call-1",
			LiveKitToken: "token",
		},
	}
	handler := NewVoiceHandler(fake)
	principal := testPrincipal()

	_, pushData, err := handler.HandleVoiceCommand(context.Background(), principal, "call_invite", json.RawMessage(`{"callee":200}`))
	if err != nil {
		t.Fatalf("HandleVoiceCommand() error = %v", err)
	}
	if pushData != nil {
		t.Fatal("pushData is non-nil, want nil for empty PushJson")
	}
}

// --- Messaging.HandleVoice delegation test ---

func TestMessaging_HandleVoice_Delegation(t *testing.T) {
	fake := &fakeVDAClient{
		voiceResp: &rpc.VoiceEventResponse{
			CallID:       "call-1",
			LiveKitToken: "token-1",
		},
	}
	msg := &Messaging{voice: NewVoiceHandler(fake)}
	principal := testPrincipal()

	ack, err := msg.HandleVoice(context.Background(), principal, "call_invite", json.RawMessage(`{"callee":200}`))
	if err != nil {
		t.Fatalf("HandleVoice() error = %v", err)
	}
	if ack == nil {
		t.Fatal("HandleVoice() ack is nil")
	}
}

// --- HandleCommand voice routing tests ---

func TestHandleCommand_VoiceRouting_CallInvite(t *testing.T) {
	fake := &fakeVDAClient{
		voiceResp: &rpc.VoiceEventResponse{
			CallID:       "call-1",
			LiveKitToken: "tk",
		},
	}
	msg := &Messaging{voice: NewVoiceHandler(fake)}
	principal := testPrincipal()

	reply, err := msg.HandleCommand(context.Background(), principal, []byte(`{"action":"call_invite","data":{"callee":200}}`))
	if err != nil {
		t.Fatalf("HandleCommand() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(reply, &parsed); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if parsed["livekit_token"] != "tk" {
		t.Fatalf("reply livekit_token = %v, want tk", parsed["livekit_token"])
	}
}

func TestHandleCommand_VoiceRouting_RoomJoin(t *testing.T) {
	fake := &fakeVDAClient{
		voiceResp: &rpc.VoiceEventResponse{
			LiveKitToken: "room-tk",
			LiveKitRoom:  "room-1",
		},
	}
	msg := &Messaging{voice: NewVoiceHandler(fake)}
	principal := testPrincipal()

	reply, err := msg.HandleCommand(context.Background(), principal, []byte(`{"action":"room_join","data":{"room_id":"room-1"}}`))
	if err != nil {
		t.Fatalf("HandleCommand() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(reply, &parsed); err != nil {
		t.Fatalf("Unmarshal error = %v", err)
	}
	if parsed["livekit_token"] != "room-tk" {
		t.Fatalf("reply livekit_token = %v, want room-tk", parsed["livekit_token"])
	}
}

func TestHandleCommand_VoiceRouting_MuteToggle(t *testing.T) {
	fake := &fakeVDAClient{
		voiceResp: &rpc.VoiceEventResponse{},
	}
	msg := &Messaging{voice: NewVoiceHandler(fake)}
	principal := testPrincipal()

	_, err := msg.HandleCommand(context.Background(), principal, []byte(`{"action":"mute_toggle","data":{"room_id":"r1","muted":true}}`))
	if err != nil {
		t.Fatalf("HandleCommand() error = %v", err)
	}
}

func TestHandleCommand_UnknownAction(t *testing.T) {
	fake := &fakeVDAClient{}
	msg := &Messaging{voice: NewVoiceHandler(fake)}
	principal := testPrincipal()

	_, err := msg.HandleCommand(context.Background(), principal, []byte(`{"action":"nonexistent"}`))
	if err == nil {
		t.Fatal("HandleCommand() expected error for unknown action")
	}
	if !strings.Contains(err.Error(), "unsupported action") {
		t.Fatalf("error = %q, want containing 'unsupported action'", err.Error())
	}
}
