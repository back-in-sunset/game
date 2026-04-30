package pipeline

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"im/internal/auth"
	"im/internal/domain"
	"im/internal/router"
	"im/internal/session"
	"im/internal/store"
)

func TestHandleCommandListConversations(t *testing.T) {
	mem := store.NewMemoryStore()
	sessions := session.NewManager()
	msg := NewMessaging(router.NewLocalRouter("", sessions, mem, nil, nil, nil), mem)
	principal := auth.Principal{
		UserID: 1,
		Domain: domain.DomainPlatform,
	}

	if err := mem.SaveMessage(context.Background(), domain.Envelope{
		Domain:   domain.DomainPlatform,
		Sender:   2,
		Receiver: 1,
		MsgType:  "direct_message",
		Seq:      1,
		SentAt:   time.Now(),
	}); err != nil {
		t.Fatalf("SaveMessage() error = %v", err)
	}

	reply, err := msg.HandleCommand(context.Background(), principal, []byte(`{"action":"list_conversations"}`))
	if err != nil {
		t.Fatalf("HandleCommand() error = %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(reply, &parsed); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if parsed["type"] != "conversation_list" {
		t.Fatalf("type = %v, want conversation_list", parsed["type"])
	}
}
