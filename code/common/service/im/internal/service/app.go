package service

import (
	"context"

	"im/internal/auth"
	"im/internal/config"
	"im/internal/discovery"
	"im/internal/goimx"
	"im/internal/notifier"
	"im/internal/pipeline"
	"im/internal/presence"
	"im/internal/router"
	"im/internal/scope"
	"im/internal/session"
	"im/internal/store"
	"im/internal/transport/tcp"
	"im/internal/transport/ws"

	"google.golang.org/grpc"
)

type App struct {
	cfg       config.Config
	registry  discovery.Registry
	ws        *ws.Server
	tcp       *tcp.Server
	adapter   *goimx.Adapter
	notifier  *notifier.Notifier
	messaging *pipeline.Messaging
	store     *store.CompositeStore
	router    *router.LocalRouter
	presence  *presence.Tracker
	internal  *grpc.Server
}

func NewApp(cfg config.Config) (*App, error) {
	authProvider, err := auth.NewJWTProvider(cfg.Auth.PublicKeyFile)
	if err != nil {
		return nil, err
	}

	sessions := session.NewManagerWithBuckets(cfg.Session.BucketCount)
	scopeResolver := scope.NewResolver(cfg.Scope)
	messageStore, err := store.NewMySQLRedisStore(cfg)
	if err != nil {
		return nil, err
	}
	presenceTracker := presence.NewTracker(cfg.Redis)
	registry, err := discovery.NewEtcdRegistry(
		cfg.Discovery.Endpoints,
		cfg.Discovery.ServicePrefix,
		cfg.Discovery.LeaseTTLSeconds,
	)
	if err != nil {
		return nil, err
	}
	rt := router.NewLocalRouter(cfg.NodeID, sessions, messageStore, registry, presenceTracker, router.NewGRPCForwarder())
	messaging := pipeline.NewMessaging(rt, messageStore)

	return &App{
		cfg:       cfg,
		registry:  registry,
		ws:        ws.New(cfg.Listen.WebSocket, cfg.NodeID, authProvider, scopeResolver, sessions, messaging, presenceTracker, cfg.Session),
		tcp:       tcp.New(cfg.Listen.TCP, cfg.NodeID, authProvider, scopeResolver, sessions, messaging, presenceTracker, cfg.Session),
		adapter:   goimx.NewAdapter(),
		notifier:  notifier.New(rt),
		messaging: messaging,
		store:     messageStore,
		router:    rt,
		presence:  presenceTracker,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	if err := a.registry.Register(ctx, discovery.Node{
		ID:        a.cfg.NodeID,
		Service:   a.cfg.ServiceName,
		WebSocket: a.cfg.Listen.WebSocket,
		TCP:       a.cfg.Listen.TCP,
		RPC:       a.cfg.Listen.RPC,
		Domains:   []string{"platform", "tenant"},
	}); err != nil {
		return err
	}
	internal, _, err := startInternalRPC(a.cfg.Listen.RPC, a.router)
	if err != nil {
		return err
	}
	a.internal = internal

	if err := a.ws.Start(); err != nil {
		return err
	}
	if err := a.tcp.Start(); err != nil {
		return err
	}
	_ = a.adapter
	return nil
}

func (a *App) Stop(ctx context.Context) {
	_ = a.ws.Stop(ctx)
	_ = a.tcp.Stop(ctx)
	if a.internal != nil {
		a.internal.GracefulStop()
	}
	_ = a.registry.Close()
	if a.presence != nil {
		_ = a.presence.Close()
	}
	if a.store != nil {
		_ = a.store.Close()
	}
}
