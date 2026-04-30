package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"im/internal/domain"
	grpcim "im/rpc/im"

	"google.golang.org/grpc"
)

type deliverLocalRouter interface {
	DeliverLocal(ctx context.Context, envelope domain.Envelope) (int, error)
}

type internalRPCServer struct {
	router deliverLocalRouter
	grpcim.UnimplementedIMServer
}

func (s *internalRPCServer) DeliverInternal(ctx context.Context, in *grpcim.DeliverInternalRequest) (*grpcim.DeliverInternalResponse, error) {
	msg := in.GetMessage()
	if msg == nil {
		return nil, fmt.Errorf("message is required")
	}
	payload := map[string]any{}
	if msg.GetPayloadJson() != "" {
		if err := json.Unmarshal([]byte(msg.GetPayloadJson()), &payload); err != nil {
			return nil, err
		}
	}
	delivered, err := s.router.DeliverLocal(ctx, domain.Envelope{
		Domain:   domain.IMDomain(msg.GetDomain()),
		Scope:    domain.Scope{TenantID: msg.GetScope().GetTenantID(), ProjectID: msg.GetScope().GetProjectID(), Environment: msg.GetScope().GetEnvironment()},
		Sender:   msg.GetSender(),
		Receiver: msg.GetReceiver(),
		MsgType:  msg.GetMsgType(),
		Seq:      msg.GetSeq(),
		Payload:  payload,
		SentAt:   time.Unix(msg.GetSentAtUnix(), 0).UTC(),
	})
	if err != nil {
		return nil, err
	}
	return &grpcim.DeliverInternalResponse{Delivered: int64(delivered)}, nil
}

func startInternalRPC(addr string, router deliverLocalRouter) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}
	server := grpc.NewServer()
	grpcim.RegisterIMServer(server, &internalRPCServer{router: router})
	go func() {
		_ = server.Serve(lis)
	}()
	return server, lis, nil
}
