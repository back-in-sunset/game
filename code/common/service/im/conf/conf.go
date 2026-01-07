package conf

import (
	xtime "im/pkg/time"
)

var (
	Conf *Config
)

type Config struct {
	Debug    bool
	TCP      *TCP
	Protocol *Protocol
}

type TCP struct {
	KeepAlive    bool
	Bind         []string
	Rcvbuf       int
	Sndbuf       int
	Reader       int
	ReadBuf      int
	ReadBufSize  int
	Writer       int
	WriteBuf     int
	WriteBufSize int
}

// Protocol is protocol config.
type Protocol struct {
	Timer            int
	TimerSize        int
	SvrProto         int
	CliProto         int
	HandshakeTimeout xtime.Duration
}
