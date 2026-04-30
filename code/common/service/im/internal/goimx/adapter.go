package goimx

import (
	"time"

	goimprotocol "github.com/Terry-Mao/goim/api/protocol"
)

const (
	OpAuthReply      = goimprotocol.OpAuthReply
	OpSendMsg        = goimprotocol.OpSendMsg
	OpSendMsgReply   = goimprotocol.OpSendMsgReply
	OpHeartbeat      = goimprotocol.OpHeartbeat
	OpHeartbeatReply = goimprotocol.OpHeartbeatReply
)

// Adapter keeps our service code pinned to a small set of goim protocol
// primitives while leaving business routing and storage under local control.
type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) NewAuthReply(seq int32, body []byte) *goimprotocol.Proto {
	return &goimprotocol.Proto{
		Ver:  1,
		Op:   OpAuthReply,
		Seq:  seq,
		Body: body,
	}
}

func (a *Adapter) NewHeartbeatReply(seq int32) *goimprotocol.Proto {
	return &goimprotocol.Proto{
		Ver:  1,
		Op:   OpHeartbeatReply,
		Seq:  seq,
		Body: []byte(time.Now().UTC().Format(time.RFC3339)),
	}
}
