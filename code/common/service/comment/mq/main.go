package main

import (
	"os"
	"os/signal"
	"time"

	"log"

	"github.com/apache/rocketmq-clients/golang"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	Topic         = "QMGTestTopic"
	ConsumerGroup = "QMGTestGroup"
	Endpoint      = "192.168.10.103:8081"
	AccessKey     = ""
	AccessSecret  = ""
)

var (
	// maximum waiting time for receive func
	awaitDuration = time.Second * 5
	// maximum number of messages received at one time
	maxMessageNum int32 = 16
	// invisibleDuration should > 20s
	invisibleDuration = time.Second * 20
	// receive messages in a loop
)

func main() {
	var do Do
	q := MustNewQueue(RkConf{
		ConsumerGroup:     ConsumerGroup,
		Endpoint:          Endpoint,
		AccessKey:         AccessKey,
		AccessSecret:      AccessSecret,
		Conns:             1,
		MaxMessageNum:     10,
		InvisibleDuration: 20 * time.Second,
	}, &do, func(qo *queueOptions) {
		qo.awaitDuration = awaitDuration
		qo.subscriptionExpressions = map[string]*golang.FilterExpression{
			Topic: golang.SUB_ALL,
		}
	})
	q.Start()
	defer q.Stop()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt)

	<-sigChan
	logx.Infof("signal: %s", sigChan)
}

// Do 处理消息
type Do struct{}

// Consume 消费消息
func (d *Do) Consume(key string, val []byte) error {
	logx.Infof("key: %s, val: %s", key, string(val))
	time.Sleep(10 * time.Second)
	log.Println("------------------------------------")
	logx.Infof("consume done")
	logx.ErrorStack()
	return nil
}
