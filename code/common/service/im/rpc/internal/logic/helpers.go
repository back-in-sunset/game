package logic

import (
	"encoding/json"
	"fmt"
	"time"

	"im/internal/auth"
	"im/internal/domain"
	"im/rpc/im"
)

func principalFromRequest(imDomain string, rpcScope *im.Scope, userID int64) (auth.Principal, error) {
	scope := scopeFromRPC(rpcScope)
	principal := auth.Principal{
		UserID: userID,
		Domain: domain.IMDomain(imDomain),
		Scope:  scope,
	}
	if userID <= 0 {
		return auth.Principal{}, fmt.Errorf("user_id is required")
	}
	if err := scope.Validate(principal.Domain); err != nil {
		return auth.Principal{}, err
	}
	return principal, nil
}

func scopeFromRPC(in *im.Scope) domain.Scope {
	if in == nil {
		return domain.Scope{}
	}
	return domain.Scope{
		TenantID:    in.GetTenantID(),
		ProjectID:   in.GetProjectID(),
		Environment: in.GetEnvironment(),
	}
}

func payloadFromJSON(raw string) (map[string]any, error) {
	if raw == "" {
		return map[string]any{}, nil
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("invalid payload json: %w", err)
	}
	return out, nil
}

func payloadToJSON(payload map[string]any) string {
	if len(payload) == 0 {
		return "{}"
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func toRPCConversation(in domain.Conversation) *im.Conversation {
	return &im.Conversation{
		Key:             in.Key,
		Domain:          string(in.Domain),
		Scope:           toRPCScope(in.Scope),
		PeerUserID:      in.PeerUserID,
		UnreadCount:     in.UnreadCount,
		ReadSeq:         in.ReadSeq,
		UpdatedAtUnix:   in.UpdatedAt.Unix(),
		LastSender:      in.LastMessage.Sender,
		LastReceiver:    in.LastMessage.Receiver,
		LastMsgType:     in.LastMessage.MsgType,
		LastSeq:         in.LastMessage.Seq,
		LastPayloadJson: payloadToJSON(in.LastMessage.Payload),
		LastSentAtUnix:  in.LastMessage.SentAt.Unix(),
	}
}

func toRPCMessage(in domain.Envelope) *im.Message {
	return &im.Message{
		Domain:      string(in.Domain),
		Scope:       toRPCScope(in.Scope),
		Sender:      in.Sender,
		Receiver:    in.Receiver,
		MsgType:     in.MsgType,
		Seq:         in.Seq,
		PayloadJson: payloadToJSON(in.Payload),
		SentAtUnix:  in.SentAt.Unix(),
	}
}

func toRPCScope(in domain.Scope) *im.Scope {
	return &im.Scope{
		TenantID:    in.TenantID,
		ProjectID:   in.ProjectID,
		Environment: in.Environment,
	}
}

func ensureSentAt(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}
