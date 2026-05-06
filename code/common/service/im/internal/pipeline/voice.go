package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"im/internal/auth"

	rpc "vda/rpc"
)

type VoiceHandler struct {
	vda rpc.VDAClient
}

func NewVoiceHandler(vda rpc.VDAClient) *VoiceHandler {
	return &VoiceHandler{vda: vda}
}

func (h *VoiceHandler) HandleVoiceCommand(ctx context.Context, principal auth.Principal, action string, data json.RawMessage) ([]byte, json.RawMessage, error) {
	req := &rpc.VoiceEventRequest{
		Action: action,
		Caller: &rpc.CallerInfo{
			UserID:      principal.UserID,
			Domain:      string(principal.Domain),
			TenantID:    principal.Scope.TenantID,
			ProjectID:   principal.Scope.ProjectID,
			Environment: principal.Scope.Environment,
		},
	}

	switch action {
	case "call_invite":
		var input struct {
			Callee int64 `json:"callee"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, nil, fmt.Errorf("invalid call_invite payload: %w", err)
		}
		req.ReceiverID = input.Callee

	case "call_accept", "call_reject", "call_end", "call_cancel":
		var input struct {
			CallID string `json:"call_id"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, nil, fmt.Errorf("invalid %s payload: %w", action, err)
		}
		req.CallID = input.CallID

	case "room_join", "room_leave":
		var input struct {
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, nil, fmt.Errorf("invalid %s payload: %w", action, err)
		}
		req.RoomID = input.RoomID

	case "mute_toggle":
		var input struct {
			RoomID string `json:"room_id"`
			Muted  bool   `json:"muted"`
		}
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, nil, fmt.Errorf("invalid mute_toggle payload: %w", err)
		}
		req.RoomID = input.RoomID
	}

	resp, err := h.vda.HandleVoiceEvent(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("vda voice event: %w", err)
	}

	ack := map[string]any{
		"action": action + "_ack",
	}
	if resp.CallID != "" {
		ack["call_id"] = resp.CallID
	}
	if resp.State != "" {
		ack["state"] = resp.State
	}
	if resp.LiveKitToken != "" {
		ack["livekit_token"] = resp.LiveKitToken
	}
	if resp.LiveKitRoom != "" {
		ack["livekit_room"] = resp.LiveKitRoom
	}
	ackJSON, _ := json.Marshal(ack)

	var pushData json.RawMessage
	if resp.PushJson != "" {
		pushData = json.RawMessage(resp.PushJson)
	} else {
		pushData = nil
	}

	return ackJSON, pushData, nil
}
