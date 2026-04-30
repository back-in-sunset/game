package pipeline

import (
	"context"
	"encoding/json"
	"fmt"

	"im/internal/auth"
)

type CommandEnvelope struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

func (m *Messaging) HandleCommand(ctx context.Context, principal auth.Principal, payload []byte) ([]byte, error) {
	var cmd CommandEnvelope
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return nil, fmt.Errorf("invalid command payload")
	}

	switch cmd.Action {
	case "send":
		var input MessageInput
		if err := json.Unmarshal(cmd.Data, &input); err != nil {
			return nil, fmt.Errorf("invalid send payload")
		}
		return m.HandleSend(ctx, principal, input)
	case "list_conversations":
		return m.HandleListConversations(ctx, principal)
	case "list_messages":
		var input QueryInput
		if err := json.Unmarshal(cmd.Data, &input); err != nil {
			return nil, fmt.Errorf("invalid list_messages payload")
		}
		return m.HandleListMessages(ctx, principal, input)
	case "mark_read":
		var input QueryInput
		if err := json.Unmarshal(cmd.Data, &input); err != nil {
			return nil, fmt.Errorf("invalid mark_read payload")
		}
		return m.HandleMarkRead(ctx, principal, input)
	default:
		return nil, fmt.Errorf("unsupported action %q", cmd.Action)
	}
}
