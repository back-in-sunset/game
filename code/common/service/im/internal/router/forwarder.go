package router

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"im/internal/domain"
	"im/rpc/im"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Forwarder interface {
	Forward(ctx context.Context, rpcAddr string, envelope domain.Envelope) (int, error)
}

type GRPCForwarder struct {
	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

func NewGRPCForwarder() *GRPCForwarder {
	return &GRPCForwarder{conns: make(map[string]*grpc.ClientConn)}
}

func (f *GRPCForwarder) Forward(ctx context.Context, rpcAddr string, envelope domain.Envelope) (int, error) {
	conn, err := f.conn(rpcAddr)
	if err != nil {
		return 0, err
	}
	client := im.NewIMClient(conn)
	resp, err := client.DeliverInternal(ctx, &im.DeliverInternalRequest{
		Message: &im.Message{
			Domain:      string(envelope.Domain),
			Scope:       &im.Scope{TenantID: envelope.Scope.TenantID, ProjectID: envelope.Scope.ProjectID, Environment: envelope.Scope.Environment},
			Sender:      envelope.Sender,
			Receiver:    envelope.Receiver,
			MsgType:     envelope.MsgType,
			Seq:         envelope.Seq,
			PayloadJson: payloadToJSON(envelope.Payload),
			SentAtUnix:  envelope.SentAt.Unix(),
		},
	})
	if err != nil {
		return 0, err
	}
	return int(resp.GetDelivered()), nil
}

func (f *GRPCForwarder) conn(addr string) (*grpc.ClientConn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if conn, ok := f.conns[addr]; ok {
		return conn, nil
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	f.conns[addr] = conn
	return conn, nil
}

func payloadToJSON(payload map[string]any) string {
	if len(payload) == 0 {
		return "{}"
	}
	raw, _ := jsonMarshal(payload)
	return string(raw)
}

var jsonMarshal = func(v any) ([]byte, error) {
	return json.Marshal(v)
}
