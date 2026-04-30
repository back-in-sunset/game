package svc

import (
	baseconfig "im/internal/config"
	"im/internal/discovery"
	"im/internal/presence"
	"im/internal/router"
	"im/internal/session"
	"im/internal/store"
	"im/rpc/internal/config"
)

type ServiceContext struct {
	Config       config.Config
	MessageStore *store.CompositeStore
	Router       *router.LocalRouter
}

func NewServiceContext(c config.Config) *ServiceContext {
	messageStore, err := store.NewMySQLRedisStore(baseconfig.Config{
		Mysql: c.Mysql,
		Redis: c.Redis,
	})
	if err != nil {
		panic(err)
	}

	// RPC 服务当前仅负责业务入口和持久化；在线会话仍由长连接进程维护。
	sessions := session.NewManagerWithBuckets(64)
	registry, err := discovery.NewEtcdRegistry(c.Discovery.Endpoints, c.Discovery.ServicePrefix, c.Discovery.LeaseTTLSeconds)
	if err != nil {
		panic(err)
	}
	presenceTracker := presence.NewTracker(c.Redis)
	return &ServiceContext{
		Config:       c,
		MessageStore: messageStore,
		Router:       router.NewLocalRouter("", sessions, messageStore, registry, presenceTracker, router.NewGRPCForwarder()),
	}
}

func (s *ServiceContext) Stop() {
	if s.MessageStore != nil {
		_ = s.MessageStore.Close()
	}
}
