package imclient

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestFacadeSendSystemNoticeShapesRequest(t *testing.T) {
	fake := &fakeIMClient{sendResp: &SendMessageResponse{Success: true}}
	facade := NewFacade(fake)

	_, err := facade.SendSystemNotice(context.Background(), SystemNotice{
		Scope: ScopeRef{
			Domain:      "tenant",
			TenantID:    "tenant-1",
			ProjectID:   "project-1",
			Environment: "prod",
		},
		Receiver: 1002,
		Seq:      7,
		Payload:  map[string]any{"title": "hello"},
	})
	if err != nil {
		t.Fatalf("SendSystemNotice() error = %v", err)
	}
	if fake.lastSend == nil {
		t.Fatal("expected SendMessage to be called")
	}
	if fake.lastSend.MsgType != MsgTypeSystemNotice {
		t.Fatalf("MsgType = %q, want %q", fake.lastSend.MsgType, MsgTypeSystemNotice)
	}
	if fake.lastSend.Sender != 0 {
		t.Fatalf("Sender = %d, want 0", fake.lastSend.Sender)
	}
	if fake.lastSend.Domain != "tenant" {
		t.Fatalf("Domain = %q, want tenant", fake.lastSend.Domain)
	}
}

func TestFacadeSendDirectMessageRequiresSender(t *testing.T) {
	facade := NewFacade(&fakeIMClient{})
	_, err := facade.SendDirectMessage(context.Background(), DirectMessage{
		Scope:    ScopeRef{Domain: "platform"},
		Receiver: 1002,
		Seq:      1,
		Payload:  map[string]any{"text": "hello"},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

type fakeIMClient struct {
	lastSend *SendMessageRequest
	sendResp *SendMessageResponse
	sendErr  error
}

func (f *fakeIMClient) SendMessage(_ context.Context, in *SendMessageRequest, _ ...grpc.CallOption) (*SendMessageResponse, error) {
	f.lastSend = in
	if f.sendResp == nil {
		f.sendResp = &SendMessageResponse{Success: true}
	}
	return f.sendResp, f.sendErr
}

func (f *fakeIMClient) ListConversations(context.Context, *ListConversationsRequest, ...grpc.CallOption) (*ListConversationsResponse, error) {
	return nil, nil
}

func (f *fakeIMClient) ListMessages(context.Context, *ListMessagesRequest, ...grpc.CallOption) (*ListMessagesResponse, error) {
	return nil, nil
}

func (f *fakeIMClient) MarkRead(context.Context, *MarkReadRequest, ...grpc.CallOption) (*ActionResponse, error) {
	return nil, nil
}

func (f *fakeIMClient) DeliverInternal(context.Context, *DeliverInternalRequest, ...grpc.CallOption) (*DeliverInternalResponse, error) {
	return nil, nil
}
