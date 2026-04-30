package goimx

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"time"

	goimprotocol "github.com/Terry-Mao/goim/api/protocol"
)

const (
	Version          = 1
	HeaderSize       = 16
	MaxBodySize      = 64 << 10
	OpAuth           = goimprotocol.OpAuth
	OpAuthReply      = goimprotocol.OpAuthReply
	OpHeartbeat      = goimprotocol.OpHeartbeat
	OpHeartbeatReply = goimprotocol.OpHeartbeatReply
	OpServerPush     = goimprotocol.OpSendMsg
	OpCommandReply   = goimprotocol.OpSendMsgReply
	OpError          = goimprotocol.OpDisconnectReply
)

type Frame struct {
	Ver  int32
	Op   int32
	Seq  int32
	Body []byte
}

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Decode(reader io.Reader) (Frame, error) {
	var header [HeaderSize]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return Frame{}, err
	}
	packLen := int(binary.BigEndian.Uint32(header[0:4]))
	headerLen := int(binary.BigEndian.Uint16(header[4:6]))
	if packLen < HeaderSize || headerLen != HeaderSize {
		return Frame{}, fmt.Errorf("invalid frame header")
	}
	bodyLen := packLen - headerLen
	if bodyLen < 0 || bodyLen > MaxBodySize {
		return Frame{}, fmt.Errorf("invalid frame body size")
	}
	frame := Frame{
		Ver: int32(binary.BigEndian.Uint16(header[6:8])),
		Op:  int32(binary.BigEndian.Uint32(header[8:12])),
		Seq: int32(binary.BigEndian.Uint32(header[12:16])),
	}
	if bodyLen == 0 {
		return frame, nil
	}
	frame.Body = make([]byte, bodyLen)
	if _, err := io.ReadFull(reader, frame.Body); err != nil {
		return Frame{}, err
	}
	return frame, nil
}

func (a *Adapter) Encode(op int32, seq int32, body []byte, dst *bytes.Buffer) ([]byte, error) {
	if len(body) > MaxBodySize {
		return nil, fmt.Errorf("frame body too large")
	}
	if dst == nil {
		dst = &bytes.Buffer{}
	} else {
		dst.Reset()
	}
	total := HeaderSize + len(body)
	var header [HeaderSize]byte
	binary.BigEndian.PutUint32(header[0:4], uint32(total))
	binary.BigEndian.PutUint16(header[4:6], uint16(HeaderSize))
	binary.BigEndian.PutUint16(header[6:8], uint16(Version))
	binary.BigEndian.PutUint32(header[8:12], uint32(op))
	binary.BigEndian.PutUint32(header[12:16], uint32(seq))
	dst.Write(header[:])
	if len(body) > 0 {
		dst.Write(body)
	}
	out := make([]byte, dst.Len())
	copy(out, dst.Bytes())
	return out, nil
}

func (a *Adapter) EncodeJSON(op int32, seq int32, payload any, dst *bytes.Buffer) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return a.Encode(op, seq, body, dst)
}

func (a *Adapter) NewAuthReply(seq int32, body []byte, dst *bytes.Buffer) ([]byte, error) {
	return a.Encode(OpAuthReply, seq, body, dst)
}

func (a *Adapter) NewHeartbeatReply(seq int32, dst *bytes.Buffer) ([]byte, error) {
	return a.Encode(OpHeartbeatReply, seq, []byte(time.Now().UTC().Format(time.RFC3339)), dst)
}

func (a *Adapter) NewCommandReply(seq int32, body []byte, dst *bytes.Buffer) ([]byte, error) {
	return a.Encode(OpCommandReply, seq, body, dst)
}

func (a *Adapter) NewPush(seq int32, body []byte, dst *bytes.Buffer) ([]byte, error) {
	return a.Encode(OpServerPush, seq, body, dst)
}

func (a *Adapter) NewError(seq int32, message string, dst *bytes.Buffer) ([]byte, error) {
	return a.EncodeJSON(OpError, seq, map[string]any{"error": message}, dst)
}
