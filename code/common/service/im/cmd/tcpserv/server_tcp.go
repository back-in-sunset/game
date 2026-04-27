package main

import (
	"im/conf"
	"im/internal"
	xtime "im/pkg/time"
	"time"
)

func main() {
	conf.Conf = &conf.Config{
		Debug: true,
		TCP: &conf.TCP{
			Bind:         []string{":8011"},
			Sndbuf:       4,
			Rcvbuf:       4,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      10,
			ReadBufSize:  20,
			Writer:       32,
			WriteBuf:     1024,
			WriteBufSize: 8192,
		},
		Protocol: &conf.Protocol{
			Timer:            32,
			TimerSize:        2048,
			CliProto:         5,
			SvrProto:         10,
			HandshakeTimeout: xtime.Duration(time.Second * 5),
		},
	}
	internal.InitTCP(
		internal.NewServer(conf.Conf),
		[]string{":8011"},
		1,
	)

	s := make(chan int)
	<-s
}
