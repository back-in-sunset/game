package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	rpc "vda/rpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func callerInfo(userID int64) *rpc.CallerInfo {
	return &rpc.CallerInfo{
		UserID:      userID,
		Domain:      "platform",
		TenantID:    "",
		ProjectID:   "",
		Environment: "",
	}
}

// VDA e2e 集成测试 — 在 Docker 环境中通过 gRPC 测试完整通话/房间生命周期。
// 运行方式：docker compose up --build --abort-on-container-exit
func main() {
	addr := os.Getenv("VDA_ADDR")
	if addr == "" {
		addr = "127.0.0.1:9101"
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: grpc.NewClient: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	client := rpc.NewVDAClient(conn)

	// Wait for VDA to be ready.
	for i := 0; i < 30; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		_, err := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: "ping"})
		cancel()
		if err == nil || !strings.Contains(err.Error(), "Unavailable") {
			break
		}
		fmt.Printf("waiting for vda at %s... (%d/30)\n", addr, i+1)
		time.Sleep(time.Second)
	}

	passed, failed := 0, 0
	run := func(name string, fn func() error) {
		fmt.Printf("  %-50s ", name)
		if err := fn(); err != nil {
			fmt.Printf("FAIL\n    %v\n", err)
			failed++
		} else {
			fmt.Println("PASS")
			passed++
		}
	}

	ctx := context.Background()

	// 等待 VDA 就绪（轮询 GetCallState）。
	fmt.Println("=== Call Lifecycle ===")

	// 1. InitiateCall
	var callID string
	run("InitiateCall", func() error {
		resp, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
			Caller: callerInfo(1001),
			Callee: 2002,
		})
		if err != nil {
			return fmt.Errorf("InitiateCall: %w", err)
		}
		if resp.CallID == "" {
			return fmt.Errorf("CallID is empty")
		}
		if resp.State != "ringing" {
			return fmt.Errorf("state = %q, want ringing", resp.State)
		}
		if resp.LiveKitToken == "" {
			return fmt.Errorf("LiveKitToken is empty")
		}
		if !strings.HasPrefix(resp.LiveKitRoom, "call_") {
			return fmt.Errorf("LiveKitRoom = %q, want prefix call_", resp.LiveKitRoom)
		}
		callID = resp.CallID
		return nil
	})

	// 2. GetCallState after initiate
	run("GetCallState(ringing)", func() error {
		resp, err := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: callID})
		if err != nil {
			return err
		}
		if resp.State != "ringing" {
			return fmt.Errorf("state = %q, want ringing", resp.State)
		}
		if resp.Caller != 1001 {
			return fmt.Errorf("caller = %d, want 1001", resp.Caller)
		}
		if resp.Callee != 2002 {
			return fmt.Errorf("callee = %d, want 2002", resp.Callee)
		}
		return nil
	})

	// 3. AcceptCall
	run("AcceptCall", func() error {
		resp, err := client.AcceptCall(ctx, &rpc.AcceptCallRequest{
			CallID: callID,
			UserID: 2002,
		})
		if err != nil {
			return err
		}
		if resp.State != "connected" {
			return fmt.Errorf("state = %q, want connected", resp.State)
		}
		if resp.LiveKitToken == "" {
			return fmt.Errorf("LiveKitToken is empty")
		}
		return nil
	})

	// 4. GetCallState after accept
	run("GetCallState(connected)", func() error {
		resp, err := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: callID})
		if err != nil {
			return err
		}
		if resp.State != "connected" {
			return fmt.Errorf("state = %q, want connected", resp.State)
		}
		return nil
	})

	// 5. EndCall
	run("EndCall", func() error {
		resp, err := client.EndCall(ctx, &rpc.EndCallRequest{
			CallID: callID,
			UserID: 1001,
		})
		if err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf("Success = false")
		}
		return nil
	})

	// 6. GetCallState after end
	run("GetCallState(ended)", func() error {
		resp, err := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: callID})
		if err != nil {
			return err
		}
		if resp.State != "ended" {
			return fmt.Errorf("state = %q, want ended", resp.State)
		}
		return nil
	})

	// 拒绝和取消流程。
	fmt.Println("\n=== Call Reject/Cancel ===")

	// 7. Initiate + Reject
	run("RejectCall", func() error {
		init, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
			Caller: callerInfo(1001),
			Callee: 2002,
		})
		if err != nil {
			return err
		}
		rej, err := client.RejectCall(ctx, &rpc.RejectCallRequest{
			CallID: init.CallID,
			UserID: 2002,
		})
		if err != nil {
			return err
		}
		if !rej.Success {
			return fmt.Errorf("Success = false")
		}
		state, _ := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: init.CallID})
		if state.State != "rejected" {
			return fmt.Errorf("state = %q, want rejected", state.State)
		}
		return nil
	})

	// 8. Initiate + Cancel
	run("CancelCall", func() error {
		init, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
			Caller: callerInfo(1001),
			Callee: 2002,
		})
		if err != nil {
			return err
		}
		cancel, err := client.CancelCall(ctx, &rpc.CancelCallRequest{
			CallID: init.CallID,
			UserID: 1001,
		})
		if err != nil {
			return err
		}
		if !cancel.Success {
			return fmt.Errorf("Success = false")
		}
		state, _ := client.GetCallState(ctx, &rpc.GetCallStateRequest{CallID: init.CallID})
		if state.State != "cancelled" {
			return fmt.Errorf("state = %q, want cancelled", state.State)
		}
		return nil
	})

	// 房间加入/静音/离开流程。
	fmt.Println("\n=== Room Lifecycle ===")

	// 9. JoinVoiceRoom
	run("JoinVoiceRoom", func() error {
		resp, err := client.JoinVoiceRoom(ctx, &rpc.JoinVoiceRoomRequest{
			RoomID: "room-1",
			User:   callerInfo(1001),
		})
		if err != nil {
			return err
		}
		if resp.RoomID != "room-1" {
			return fmt.Errorf("RoomID = %q, want room-1", resp.RoomID)
		}
		if resp.LiveKitToken == "" {
			return fmt.Errorf("LiveKitToken is empty")
		}
		if len(resp.Participants) == 0 || resp.Participants[0] != 1001 {
			return fmt.Errorf("Participants = %v, want [1001]", resp.Participants)
		}
		return nil
	})

	// 10. GetRoomState
	run("GetRoomState", func() error {
		resp, err := client.GetRoomState(ctx, &rpc.GetRoomStateRequest{RoomID: "room-1"})
		if err != nil {
			return err
		}
		found := false
		for _, id := range resp.Participants {
			if id == 1001 {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("Participants = %v, want containing 1001", resp.Participants)
		}
		return nil
	})

	// 11. MuteToggle
	run("MuteToggle", func() error {
		resp, err := client.MuteToggle(ctx, &rpc.MuteToggleRequest{
			RoomID: "room-1",
			UserID: 1001,
			Muted:  true,
		})
		if err != nil {
			return err
		}
		if !resp.Muted {
			return fmt.Errorf("Muted = false, want true")
		}
		state, _ := client.GetRoomState(ctx, &rpc.GetRoomStateRequest{RoomID: "room-1"})
		found := false
		for _, id := range state.Muted {
			if id == 1001 {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("Muted list missing 1001: %v", state.Muted)
		}
		return nil
	})

	// 12. LeaveVoiceRoom
	run("LeaveVoiceRoom", func() error {
		resp, err := client.LeaveVoiceRoom(ctx, &rpc.LeaveVoiceRoomRequest{
			RoomID: "room-1",
			UserID: 1001,
		})
		if err != nil {
			return err
		}
		if !resp.Success {
			return fmt.Errorf("Success = false")
		}
		state, _ := client.GetRoomState(ctx, &rpc.GetRoomStateRequest{RoomID: "room-1"})
		for _, id := range state.Participants {
			if id == 1001 {
				return fmt.Errorf("participant 1001 still in room after leave")
			}
		}
		return nil
	})

	// IM 事件转发接口（模拟 IM 调用 VDA）。
	fmt.Println("\n=== HandleVoiceEvent (IM relay) ===")

	// 13. HandleVoiceEvent call_invite
	var eventCallID string
	run("HandleVoiceEvent(call_invite)", func() error {
		resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
			Action:     "call_invite",
			Caller:     callerInfo(1001),
			ReceiverID: 2002,
		})
		if err != nil {
			return err
		}
		if resp.CallID == "" {
			return fmt.Errorf("CallID is empty")
		}
		if resp.LiveKitToken == "" {
			return fmt.Errorf("LiveKitToken is empty")
		}
		if len(resp.PushTargets) < 1 {
			return fmt.Errorf("PushTargets empty, want at least callee")
		}
		eventCallID = resp.CallID
		return nil
	})

	// 14. HandleVoiceEvent call_accept
	run("HandleVoiceEvent(call_accept)", func() error {
		resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
			Action: "call_accept",
			Caller: callerInfo(2002),
			CallID: eventCallID,
		})
		if err != nil {
			return err
		}
		if resp.State != "connected" {
			return fmt.Errorf("state = %q, want connected", resp.State)
		}
		return nil
	})

	// 15. HandleVoiceEvent call_end
	run("HandleVoiceEvent(call_end)", func() error {
		resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
			Action: "call_end",
			Caller: callerInfo(1001),
			CallID: eventCallID,
		})
		if err != nil {
			return err
		}
		if resp.State != "ended" {
			return fmt.Errorf("state = %q, want ended", resp.State)
		}
		return nil
	})

	// 16. HandleVoiceEvent room_join
	run("HandleVoiceEvent(room_join)", func() error {
		resp, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
			Action: "room_join",
			Caller: callerInfo(1001),
			RoomID: "room-e2e",
		})
		if err != nil {
			return err
		}
		if resp.LiveKitToken == "" {
			return fmt.Errorf("LiveKitToken is empty")
		}
		return nil
	})

	// 17. HandleVoiceEvent room_leave
	run("HandleVoiceEvent(room_leave)", func() error {
		_, err := client.HandleVoiceEvent(ctx, &rpc.VoiceEventRequest{
			Action: "room_leave",
			Caller: callerInfo(1001),
			RoomID: "room-e2e",
		})
		if err != nil {
			return err
		}
		return nil
	})

	// 18. Verify LiveKit tokens are valid JWT
	run("LiveKitTokenIsJWT", func() error {
		resp, err := client.InitiateCall(ctx, &rpc.InitiateCallRequest{
			Caller: callerInfo(1001),
			Callee: 2002,
		})
		if err != nil {
			return err
		}
		parts := strings.Split(resp.LiveKitToken, ".")
		if len(parts) != 3 {
			return fmt.Errorf("token has %d parts, want 3 (JWT)", len(parts))
		}
		return nil
	})

	fmt.Printf("\n%d passed, %d failed\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
