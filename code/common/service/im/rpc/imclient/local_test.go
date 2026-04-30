package imclient

import (
	"context"
	"testing"
	"time"

	"im/internal/auth"
	"im/internal/domain"
	"im/internal/router"
	"im/internal/session"
	"im/internal/store"
)

func TestLocalIMSendAndQuery(t *testing.T) {
	mem := store.NewMemoryStore()
	sessions := session.NewManager()
	client := NewLocalIM(mem, router.NewLocalRouter("", sessions, mem, nil, nil, nil))

	_, err := client.SendMessage(context.Background(), &SendMessageRequest{
		Domain:   "tenant",
		Scope:    &Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"},
		Sender:   1,
		Receiver: 2,
		MsgType:  "direct_message",
		Seq:      1,
		Payload:  &MessagePayload{Json: `{"text":"hello"}`},
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	resp, err := client.ListConversations(context.Background(), &ListConversationsRequest{
		Domain: "tenant",
		Scope:  &Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"},
		UserID: 2,
	})
	if err != nil {
		t.Fatalf("ListConversations() error = %v", err)
	}
	if len(resp.Conversations) != 1 {
		t.Fatalf("len(Conversations) = %d, want 1", len(resp.Conversations))
	}

	msgs, err := client.ListMessages(context.Background(), &ListMessagesRequest{
		Domain:     "tenant",
		Scope:      &Scope{TenantID: "tenant-1", ProjectID: "project-1", Environment: "prod"},
		UserID:     2,
		PeerUserID: 1,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(msgs.Messages) != 1 {
		t.Fatalf("len(Messages) = %d, want 1", len(msgs.Messages))
	}
}

func TestLocalIMDeliverInternal(t *testing.T) {
	mem := store.NewMemoryStore()
	sessions := session.NewManager()
	receiver := testConn{id: "c1"}
	sessions.Bind(authPrincipalPlatform(2), &receiver)
	client := NewLocalIM(mem, router.NewLocalRouter("", sessions, mem, nil, nil, nil))

	resp, err := client.DeliverInternal(context.Background(), &DeliverInternalRequest{
		Message: &Message{
			Domain:      "platform",
			Scope:       &Scope{},
			Sender:      1,
			Receiver:    2,
			MsgType:     "direct_message",
			Seq:         1,
			PayloadJson: `{"text":"hello"}`,
			SentAtUnix:  time.Now().Unix(),
		},
	})
	if err != nil {
		t.Fatalf("DeliverInternal() error = %v", err)
	}
	if resp.Delivered != 1 {
		t.Fatalf("Delivered = %d, want 1", resp.Delivered)
	}
}

type testConn struct {
	id   string
	sent [][]byte
}

func (c *testConn) ID() string { return c.id }
func (c *testConn) Send(_ context.Context, payload []byte) error {
	cp := make([]byte, len(payload))
	copy(cp, payload)
	c.sent = append(c.sent, cp)
	return nil
}
func (c *testConn) Close() error { return nil }

func authPrincipalPlatform(uid int64) auth.Principal {
	return auth.Principal{UserID: uid, Domain: domain.DomainPlatform}
}
