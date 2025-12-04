package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/apache/rocketmq-clients/golang"
	"github.com/apache/rocketmq-clients/golang/credentials"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/threading"
)

// ConsumeHandler 消费处理
type ConsumeHandler interface {
	Consume(key string, value []byte) error
}

type queueOptions struct {
	awaitDuration           time.Duration
	subscriptionExpressions map[string]*golang.FilterExpression
}

// QueueOption option
type QueueOption func(*queueOptions)

type (
	// RkConf 配置
	RkConf struct {
		Endpoint      string `validate:"required"`
		ConsumerGroup string `validate:"required"`
		NameSpace     string
		AccessKey     string
		AccessSecret  string
		SecurityToken string

		Conns int

		MaxMessageNum     int32 `default:"10"`
		InvisibleDuration time.Duration
	}
)

// MustNewQueue 新建队列
func MustNewQueue(c RkConf, handler ConsumeHandler, opts ...QueueOption) queue.MessageQueue {
	q, err := NewQueue(c, handler, opts...)
	if err != nil {
		log.Fatal(err)
	}

	return q
}

// NewQueue 新建队列
func NewQueue(c RkConf, handler ConsumeHandler, opts ...QueueOption) (queue.MessageQueue, error) {
	var options queueOptions
	for _, opt := range opts {
		opt(&options)
	}

	if c.Conns < 1 {
		c.Conns = 1
	}

	q := rocketQueues{
		group: service.NewServiceGroup(),
	}

	for i := 0; i < c.Conns; i++ {
		q.queues = append(q.queues, newRocketQueue(c, handler, options))
	}

	return q, nil
}

// Queue 队列
type Queue struct {
	golang.SimpleConsumer
}

// Start 启动
func (q *Queue) Start() {
	err := q.SimpleConsumer.Start()
	if err != nil {
		log.Fatal(err)
	}
}

// Stop 停止
func (q *Queue) Stop() {
	err := q.SimpleConsumer.GracefulStop()
	if err != nil {
		log.Fatal(err)
	}
}

func newRocketQueue(c RkConf, handler ConsumeHandler, options queueOptions) queue.MessageQueue {
	simpleConsumer, err := golang.NewSimpleConsumer(&golang.Config{
		Endpoint:      c.Endpoint,
		NameSpace:     c.NameSpace,
		ConsumerGroup: c.ConsumerGroup,
		Credentials: &credentials.SessionCredentials{
			AccessKey:     c.AccessKey,
			AccessSecret:  c.AccessSecret,
			SecurityToken: c.SecurityToken,
		},
	},
		golang.WithAwaitDuration(options.awaitDuration),
		golang.WithSubscriptionExpressions(options.subscriptionExpressions),
	)
	if err != nil {
		log.Fatal(err)
	}

	return &rocketQueue{
		c:                c,
		consumer:         simpleConsumer,
		hanlder:          handler,
		channel:          make(chan *golang.MessageView, 10),
		producerRoutines: threading.NewRoutineGroup(),
		consumerRoutines: threading.NewRoutineGroup(),
	}

}

type (
	rocketQueue struct {
		c                RkConf
		consumer         golang.SimpleConsumer
		hanlder          ConsumeHandler
		channel          chan *golang.MessageView
		producerRoutines *threading.RoutineGroup
		consumerRoutines *threading.RoutineGroup
	}
	rocketQueues struct {
		queues []queue.MessageQueue
		group  *service.ServiceGroup
	}
)

func (q *rocketQueue) Start() {
	q.startConsumers()
	q.startProducers()

	q.producerRoutines.Wait()
	close(q.channel)
	q.consumerRoutines.Wait()

}

func (q *rocketQueue) Stop() {
	q.consumer.GracefulStop()
	logx.Info("rocketQueue Stop")
}

func (q *rocketQueue) consumerOne(key string, val []byte) error {
	return q.hanlder.Consume(key, val)
}

func (q *rocketQueue) startConsumers() {
	q.consumerRoutines.RunSafe(func() {
		for msg := range q.channel {
			if err := q.consumerOne(msg.GetMessageId(), msg.GetBody()); err != nil {
				logx.Errorf("consume: %s, error: %v", msg.GetBody(), err)
			}

			if err := q.consumer.Ack(context.Background(), msg); err != nil {
				logx.Errorf("ack failed: msgID: %s, error: %v", msg.GetMessageId(), err)
			}
		}
	})
}

func (q *rocketQueue) startProducers() {
	q.producerRoutines.RunSafe(func() {
		q.consumer.Start()
		for {
			msgs, err := q.consumer.Receive(context.Background(), q.c.MaxMessageNum, q.c.InvisibleDuration)
			if err == io.EOF || err == io.ErrClosedPipe {
				return
			}
			if err != nil {
				logx.Errorf("receive failed message: error: %q", err.Error())
				continue
			}

			for _, msg := range msgs {
				q.channel <- msg
			}

		}
	})

}

func (q rocketQueues) Start() {
	for _, each := range q.queues {
		q.group.Add(each)
	}
	q.group.Start()
}

func (q rocketQueues) Stop() {
	q.group.Stop()
}
