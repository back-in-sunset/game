package service

import (
	"context"

	"vda/internal/callmanager"
	"vda/internal/domain"
	"vda/internal/roommanager"
	rpc "vda/rpc"
)

type VDAServer struct {
	rpc.UnimplementedVDAServer
	calls *callmanager.Manager
	rooms *roommanager.Manager
}

func NewVDAServer(calls *callmanager.Manager, rooms *roommanager.Manager) *VDAServer {
	return &VDAServer{calls: calls, rooms: rooms}
}

func toDomainCaller(c *rpc.CallerInfo) domain.CallerInfo {
	if c == nil {
		return domain.CallerInfo{}
	}
	return domain.CallerInfo{
		UserID:      c.UserID,
		Domain:      c.Domain,
		TenantID:    c.TenantID,
		ProjectID:   c.ProjectID,
		Environment: c.Environment,
	}
}

func (s *VDAServer) InitiateCall(ctx context.Context, req *rpc.InitiateCallRequest) (*rpc.InitiateCallResponse, error) {
	caller := toDomainCaller(req.Caller)
	callID, token, room, err := s.calls.Initiate(ctx, caller, req.Callee)
	if err != nil {
		return nil, err
	}
	return &rpc.InitiateCallResponse{
		CallID:       callID,
		State:        string(domain.CallStateRinging),
		LiveKitRoom:  room,
		LiveKitToken: token,
	}, nil
}

func (s *VDAServer) AcceptCall(ctx context.Context, req *rpc.AcceptCallRequest) (*rpc.AcceptCallResponse, error) {
	token, err := s.calls.Accept(ctx, req.CallID, req.UserID)
	if err != nil {
		return nil, err
	}
	return &rpc.AcceptCallResponse{
		CallID:       req.CallID,
		State:        string(domain.CallStateConnected),
		LiveKitToken: token,
	}, nil
}

func (s *VDAServer) RejectCall(ctx context.Context, req *rpc.RejectCallRequest) (*rpc.ActionResponse, error) {
	if err := s.calls.Reject(ctx, req.CallID, req.UserID); err != nil {
		return nil, err
	}
	return &rpc.ActionResponse{Success: true}, nil
}

func (s *VDAServer) EndCall(ctx context.Context, req *rpc.EndCallRequest) (*rpc.ActionResponse, error) {
	if err := s.calls.End(ctx, req.CallID, req.UserID); err != nil {
		return nil, err
	}
	return &rpc.ActionResponse{Success: true}, nil
}

func (s *VDAServer) CancelCall(ctx context.Context, req *rpc.CancelCallRequest) (*rpc.ActionResponse, error) {
	if err := s.calls.Cancel(ctx, req.CallID, req.UserID); err != nil {
		return nil, err
	}
	return &rpc.ActionResponse{Success: true}, nil
}

func (s *VDAServer) JoinVoiceRoom(ctx context.Context, req *rpc.JoinVoiceRoomRequest) (*rpc.JoinVoiceRoomResponse, error) {
	user := toDomainCaller(req.User)
	token, participants, err := s.rooms.Join(ctx, req.RoomID, user)
	if err != nil {
		return nil, err
	}
	return &rpc.JoinVoiceRoomResponse{
		RoomID:       req.RoomID,
		LiveKitToken: token,
		Participants: participants,
	}, nil
}

func (s *VDAServer) LeaveVoiceRoom(ctx context.Context, req *rpc.LeaveVoiceRoomRequest) (*rpc.ActionResponse, error) {
	if err := s.rooms.Leave(ctx, req.RoomID, req.UserID); err != nil {
		return nil, err
	}
	return &rpc.ActionResponse{Success: true}, nil
}

func (s *VDAServer) MuteToggle(ctx context.Context, req *rpc.MuteToggleRequest) (*rpc.MuteToggleResponse, error) {
	if err := s.rooms.ToggleMute(ctx, req.RoomID, req.UserID, req.Muted); err != nil {
		return nil, err
	}
	return &rpc.MuteToggleResponse{
		RoomID: req.RoomID,
		UserID: req.UserID,
		Muted:  req.Muted,
	}, nil
}

func (s *VDAServer) HandleVoiceEvent(ctx context.Context, req *rpc.VoiceEventRequest) (*rpc.VoiceEventResponse, error) {
	switch req.Action {
	case "call_invite":
		caller := toDomainCaller(req.Caller)
		callID, token, room, err := s.calls.Initiate(ctx, caller, req.ReceiverID)
		if err != nil {
			return nil, err
		}
		return &rpc.VoiceEventResponse{
			CallID:       callID,
			State:        string(domain.CallStateRinging),
			LiveKitToken: token,
			LiveKitRoom:  room,
			PushTargets:  []int64{req.ReceiverID, caller.UserID},
		}, nil

	case "call_accept":
		token, err := s.calls.Accept(ctx, req.CallID, req.Caller.UserID)
		if err != nil {
			return nil, err
		}
		call, _ := s.calls.GetState(ctx, req.CallID)
		return &rpc.VoiceEventResponse{
			CallID:       req.CallID,
			State:        string(domain.CallStateConnected),
			LiveKitToken: token,
			LiveKitRoom:  call.LiveKitRoom,
			PushTargets:  []int64{call.CallerID},
		}, nil

	case "call_reject":
		if err := s.calls.Reject(ctx, req.CallID, req.Caller.UserID); err != nil {
			return nil, err
		}
		call, _ := s.calls.GetState(ctx, req.CallID)
		return &rpc.VoiceEventResponse{
			CallID:      req.CallID,
			State:       string(domain.CallStateRejected),
			PushTargets: []int64{call.CallerID},
		}, nil

	case "call_end":
		if err := s.calls.End(ctx, req.CallID, req.Caller.UserID); err != nil {
			return nil, err
		}
		return &rpc.VoiceEventResponse{
			CallID: req.CallID,
			State:  string(domain.CallStateEnded),
		}, nil

	case "call_cancel":
		if err := s.calls.Cancel(ctx, req.CallID, req.Caller.UserID); err != nil {
			return nil, err
		}
		call, _ := s.calls.GetState(ctx, req.CallID)
		return &rpc.VoiceEventResponse{
			CallID:      req.CallID,
			State:       string(domain.CallStateCancelled),
			PushTargets: []int64{call.CalleeID},
		}, nil

	case "room_join":
		user := toDomainCaller(req.Caller)
		token, participants, err := s.rooms.Join(ctx, req.RoomID, user)
		if err != nil {
			return nil, err
		}
		return &rpc.VoiceEventResponse{
			LiveKitToken: token,
			LiveKitRoom:  req.RoomID,
			PushTargets:  participants,
		}, nil

	case "room_leave":
		if err := s.rooms.Leave(ctx, req.RoomID, req.Caller.UserID); err != nil {
			return nil, err
		}
		participants, _ := s.rooms.GetParticipants(ctx, req.RoomID)
		ids := make([]int64, len(participants))
		for i, p := range participants {
			ids[i] = p.UserID
		}
		return &rpc.VoiceEventResponse{
			PushTargets: ids,
		}, nil

	case "mute_toggle":
		s.rooms.ToggleMute(ctx, req.RoomID, req.Caller.UserID, true) // payload would carry actual value
		return &rpc.VoiceEventResponse{}, nil

	case "disconnect":
		s.calls.OnDisconnect(ctx, req.Caller.UserID)
		return &rpc.VoiceEventResponse{}, nil

	default:
		return &rpc.VoiceEventResponse{}, nil
	}
}

func (s *VDAServer) GetCallState(ctx context.Context, req *rpc.GetCallStateRequest) (*rpc.CallStateResponse, error) {
	call, err := s.calls.GetState(ctx, req.CallID)
	if err != nil {
		return nil, err
	}
	return &rpc.CallStateResponse{
		CallID:      call.CallID,
		State:       string(call.State),
		Caller:      call.CallerID,
		Callee:      call.CalleeID,
		LiveKitRoom: call.LiveKitRoom,
	}, nil
}

func (s *VDAServer) GetRoomState(ctx context.Context, req *rpc.GetRoomStateRequest) (*rpc.RoomStateResponse, error) {
	participants, err := s.rooms.GetParticipants(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}
	muted, err := s.rooms.GetMuted(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(participants))
	for i, p := range participants {
		ids[i] = p.UserID
	}
	return &rpc.RoomStateResponse{
		RoomID:       req.RoomID,
		Participants: ids,
		Muted:        muted,
	}, nil
}
